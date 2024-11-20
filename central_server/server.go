package central

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// SetupRouter sets up the HTTP handler for WebSocket connections.
func SetupRouter() {
	http.HandleFunc("/whtest", func(w http.ResponseWriter, r *http.Request) {
		websocket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		MessageAccepterHandler(websocket)
	})
}

// MessageAccepterHandler handles incoming messages from the CLI over WebSocket.
func MessageAccepterHandler(conn *websocket.Conn) {
	go func() {
		for {
			_, encodedMessageBytes, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			message := string(encodedMessageBytes)
			fmt.Print("Received the encoded message.\n")

			parts := strings.Split(message, ":")
			if len(parts) != 2 {
				fmt.Println("Invalid message format")
				return
			}
			encodedMessage := parts[0]
			number := parts[1]

			if encodedMessage == "EncodedMessage" {
				fmt.Println("Received number:", number)
				SubdomainTransfer(conn)
			}
		}
	}()
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateRandomString generates a random alphanumeric string of length n.
func GenerateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// SubdomainTransfer handles the provisioning of resources and communication with the CLI.
func SubdomainTransfer(conn *websocket.Conn) {
	fmt.Print("Starting dynamic provisioning of peripheral server.\n")

	// Generate a random 10-character alphanumeric string
	subdomain := GenerateRandomString(10)

	// Obtain the Ingress Controller's external IP
	ingressIP, err := GetIngressControllerIP()
	if err != nil {
		log.Println("Error getting Ingress Controller IP:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("None"))
		return
	}

	// Deploy peripheral server to Kubernetes and create Ingress
	err = DeployPeripheralServer(subdomain)
	if err != nil {
		log.Println("Error deploying peripheral server:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("None"))
		return
	}

	// Send the subdomain to the CLI
	hostName := subdomain + ".pc-1827.online"
	if err := conn.WriteMessage(websocket.TextMessage, []byte(hostName)); err != nil {
		log.Println("Error sending subdomain to the CLI:", err)
		return
	}

	// Start a timer to delete the user's resources after 1 hour
	go StartCleanupTimer(subdomain, ingressIP)
}

// DeployPeripheralServer creates a Deployment, Service, and Ingress for the subdomain.
func DeployPeripheralServer(subdomain string) error {
	// Get Kubernetes client
	clientset, err := getKubernetesClient()
	if err != nil {
		return err
	}

	namespace := "default"

	deploymentName := "peripheral-server-deployment-" + subdomain
	serviceName := "peripheral-server-service-" + subdomain

	labels := map[string]string{
		"app":       "peripheral-server",
		"subdomain": subdomain,
	}

	// Define the Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   deploymentName,
			Labels: labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "peripheral-server",
							Image: "gcr.io/your-gcp-project-id/peripheral-server:latest", // Update to your image
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 2001,
								},
							},
						},
					},
				},
			},
		},
	}

	// Create the Deployment
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	fmt.Println("Creating deployment for subdomain:", subdomain)
	_, err = deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment: %v", err)
	}

	// Define the Service (Type ClusterIP)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   serviceName,
			Labels: labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,                   // Port exposed by the Service
					TargetPort: intstr.FromInt(2001), // Port your application listens on
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// Create the Service
	servicesClient := clientset.CoreV1().Services(namespace)
	fmt.Println("Creating service for subdomain:", subdomain)
	_, err = servicesClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create service: %v", err)
	}

	// Create the Ingress resource
	err = CreateIngress(subdomain, serviceName, labels)
	if err != nil {
		return fmt.Errorf("failed to create ingress: %v", err)
	}

	return nil
}

// CreateIngress creates an Ingress resource to route traffic from the subdomain to the service.
func CreateIngress(subdomain, serviceName string, labels map[string]string) error {
	clientset, err := getKubernetesClient()
	if err != nil {
		return err
	}

	namespace := "default"

	ingressName := "peripheral-server-ingress-" + subdomain
	hostName := subdomain + ".pc-1827.online"

	ingressClient := clientset.NetworkingV1().Ingresses(namespace)

	pathType := networkingv1.PathTypePrefix

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ingressName,
			Labels: labels,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: hostName,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceName,
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	fmt.Println("Creating ingress for subdomain:", subdomain)
	_, err = ingressClient.Create(context.TODO(), ingress, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
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

// StartCleanupTimer starts a timer to clean up resources after the specified duration.
func StartCleanupTimer(subdomain, ingressIP string) {
	// Wait for 1 hour
	time.Sleep(1 * time.Hour)

	// Cleanup resources
	err := CleanupUserResources(subdomain, ingressIP)
	if err != nil {
		log.Println("Error cleaning up resources for subdomain", subdomain+":", err)
	} else {
		log.Println("Successfully cleaned up resources for subdomain", subdomain)
	}
}

// CleanupUserResources deletes the Deployment, Service and Ingress for the subdomain.
func CleanupUserResources(subdomain, ingressIP string) error {
	clientset, err := getKubernetesClient()
	if err != nil {
		return err
	}

	namespace := "default"

	deploymentName := "peripheral-server-deployment-" + subdomain
	serviceName := "peripheral-server-service-" + subdomain
	ingressName := "peripheral-server-ingress-" + subdomain

	// Delete the Deployment
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	fmt.Println("Deleting deployment for subdomain:", subdomain)
	err = deploymentsClient.Delete(context.TODO(), deploymentName, metav1.DeleteOptions{})
	if err != nil {
		log.Println("Failed to delete deployment:", err)
	}

	// Delete the Service
	servicesClient := clientset.CoreV1().Services(namespace)
	fmt.Println("Deleting service for subdomain:", subdomain)
	err = servicesClient.Delete(context.TODO(), serviceName, metav1.DeleteOptions{})
	if err != nil {
		log.Println("Failed to delete service:", err)
	}

	// Delete the Ingress
	ingressClient := clientset.NetworkingV1().Ingresses(namespace)
	fmt.Println("Deleting ingress for subdomain:", subdomain)
	err = ingressClient.Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		log.Println("Failed to delete ingress:", err)
	}

	return nil
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

// int32Ptr returns a pointer to an int32.
func int32Ptr(i int32) *int32 { return &i }
