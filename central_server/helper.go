package central

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

// GenerateRandomString generates a random alphanumeric string of length n.
func GenerateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return strings.ToLower(string(b))
}

// Helper function to get a Kubernetes clientset.
func getKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig if not running inside a cluster
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kubernetes config: %v", err)
		}
	}
	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	return clientset, nil
}

// GetIngressControllerIP retrieves the external IP address of the Ingress controller.
func GetIngressControllerIP() (string, error) {
	clientset, err := getKubernetesClient()
	if err != nil {
		return "", err
	}

	// Adjust the namespace and service name according to your deployment
	svc, err := clientset.CoreV1().Services("ingress-nginx").Get(context.TODO(), "ingress-nginx-controller", metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get Ingress controller service: %v", err)
	}
	if len(svc.Status.LoadBalancer.Ingress) == 0 {
		return "", fmt.Errorf("ingress controller external IP not available yet")
	}
	ip := svc.Status.LoadBalancer.Ingress[0].IP
	return ip, nil
}

// int32Ptr returns a pointer to an int32.
func int32Ptr(i int32) *int32 { return &i }
