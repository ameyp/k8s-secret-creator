package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func requireEnv(variable string) []byte {
	value := os.Getenv(variable)
	if value == "" {
		log.Fatalf("ENV variable %s is not set", variable)
	}

	return []byte(value)
}

func getNamespace() string {
	namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Fatalf("Could not read namespace file: %v", err.Error())
	}

	return string(namespace)
}

func getSecretContent() map[string][]byte {
	secretFile := requireEnv("SECRET_FILE")

	content, err := ioutil.ReadFile(string(secretFile))
	if err != nil {
		log.Fatalf("Could not read %s: %v", secretFile, err)
	}

	m := make(map[string]string)

	err = yaml.Unmarshal(content, &m)
	if err != nil {
		log.Fatalf("Could not parse yaml into a map of string -> string, error: %v", err)
	}

	secretContent := make(map[string][]byte)

	for key, val := range m {
		secretContent[key] = []byte(val)
	}

	return secretContent
}

func GetSecretsManager(namespace string) clientv1.SecretInterface {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Could not create k8s client config: %s", err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Could not create k8s clientset: %s", err.Error())
	}

	return clientset.CoreV1().Secrets(string(namespace))
}

func CreateSecret(secretName string, secretContent map[string][]byte, namespace string, secretsManager clientv1.SecretInterface) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			Namespace: namespace,
			Labels: map[string] string {
				"owner": "k8s-secret-creator",
			},
		},
		Type: "Opaque",
		Data: secretContent,
	}

	_, err := secretsManager.Create(context.TODO(), secret, metav1.CreateOptions{})

	if err != nil {
		log.Fatalf("Could not create secret: %s", err.Error())
	}

	log.Print("Created the secret")
}

func main() {
	namespace := getNamespace()
	secretName := string(requireEnv("SECRET_NAME"))

	log.Printf("Managing secret [%s] in namespace [%s]", secretName, namespace)

	secretsManager := GetSecretsManager(namespace)

	// Check if the secret already exists. If it does, delete it.
	secrets, err := secretsManager.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Could not list secrets: %s", err.Error())
	}

	for _, s := range secrets.Items {
		if s.Name == secretName {
			err := secretsManager.Delete(context.TODO(), secretName, metav1.DeleteOptions{})
			if err != nil {
				log.Fatalf("Could not delete existing secret: %s", err.Error())
			}
			log.Print("Deleted existing secret")
		}
	}

	secretContent := getSecretContent()
	CreateSecret(secretName, secretContent, namespace, secretsManager)
}
