package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	secrets "github.com/ameyp/k8s-secret-creator/secrets"
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

func getSecretContent() map[string]string {
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

	secretContent := make(map[string]string)

	for key, val := range m {
		secretContent[key] = val
	}

	return secretContent
}

func main() {
	namespace := getNamespace()
	secretName := string(requireEnv("SECRET_NAME"))

	log.Printf("Managing secret [%s] in namespace [%s]", secretName, namespace)

	secretsManager, err := secrets.GetSecretsManager(namespace)

	if err != nil {
		log.Fatalf("Could not get secrets manager: %s", err.Error())
	}

	// Check if the secret already exists. If it does, delete it.
	secretList, err := secretsManager.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Could not list secrets: %s", err.Error())
	}

	for _, s := range secretList.Items {
		if s.Name == secretName {
			err := secretsManager.Delete(context.TODO(), secretName, metav1.DeleteOptions{})
			if err != nil {
				log.Fatalf("Could not delete existing secret: %s", err.Error())
			}
			log.Print("Deleted existing secret")
		}
	}

	secretContent := getSecretContent()
	err = secrets.CreateSecret(secretName, secretContent, namespace, secretsManager)
	if err != nil {
		log.Fatalf("Could not create the secret: %s", err.Error())
	}

	log.Print("Created the secret")
}
