package secrets

import (
	"context"

	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetSecretsManager(namespace string) (clientv1.SecretInterface, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset.CoreV1().Secrets(string(namespace)), nil
}

// The values for secretContent must be base64-encoded.
func CreateSecret(secretName string, secretContent map[string]string, namespace string, secretsManager clientv1.SecretInterface) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			Namespace: namespace,
			Labels: map[string] string {
				"owner": "k8s-secret-creator",
			},
		},
		Type: "Opaque",
		StringData: secretContent,
	}

	_, err := secretsManager.Create(context.TODO(), secret, metav1.CreateOptions{})

	if err != nil {
		return err
	}

	return nil
}

