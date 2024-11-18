package central

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func SetupRouter() {
	http.HandleFunc("/whtest", func(w http.ResponseWriter, r *http.Request) {
		websocket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		setWebSocketConnection(websocket)
		MessageAccepterHandler(websocket)
	})
}

// Handles receiving the webhook from the CLI
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
				// No further change needed in central server
				SubdomainTransfer(conn)
			}
		}
	}()
}

func SubdomainTransfer(conn *websocket.Conn) {
	fmt.Print("Starting dynamic provisioning of peripheral server.\n")

	// Generate a unique identifier for the user
	userID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Deploy peripheral server to Kubernetes and get the service address
	serviceAddress, err := DeployPeripheralServer(userID)
	if err != nil {
		log.Println("Error deploying peripheral server:", err)
		if err := conn.WriteMessage(websocket.TextMessage, []byte("None")); err != nil {
			log.Println("Error sending error message to the CLI:", err)
		}
		return
	}

	// Send the service address to the CLI
	if err := conn.WriteMessage(websocket.TextMessage, []byte(serviceAddress)); err != nil {
		log.Println("Error sending service address to the CLI:", err)
		return
	}

	// Start a timer to delete the user's resources after 1 hour
	go StartCleanupTimer(userID)
}

func DeployPeripheralServer(userID string) (string, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config
	config, err = rest.InClusterConfig()
	if err != nil {
		// If in-cluster config fails, fallback to kubeconfig
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return "", fmt.Errorf("failed to create Kubernetes config: %v", err)
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Define the namespace to deploy to
	namespace := "default"

	// Create unique names for the deployment and service using userID
	deploymentName := "peripheral-server-deployment-" + userID
	serviceName := "peripheral-server-service-" + userID

	// Define labels to identify resources belonging to this user
	labels := map[string]string{
		"app":    "peripheral-server",
		"userID": userID,
	}

	// Define the deployment
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
							Image: "pc1827/peripheral-server:latest", // Ensure this image is available in the cluster
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

	// Create the deployment
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	fmt.Println("Creating deployment for user:", userID)
	_, err = deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create deployment: %v", err)
	}

	// Define the service
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
					Port:       2001,
					TargetPort: intstr.FromInt(2001),
					// Optionally assign a static NodePort
					// NodePort:   30000 + (user-specific offset),
				},
			},
			Type: corev1.ServiceTypeNodePort, // Use NodePort to expose the service
		},
	}

	// Create the service
	servicesClient := clientset.CoreV1().Services(namespace)
	fmt.Println("Creating service for user:", userID)
	svc, err := servicesClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create service: %v", err)
	}

	// Get the NodePort assigned
	nodePort := svc.Spec.Ports[0].NodePort

	// Use localhost since Minikube maps NodePorts to localhost
	serviceAddress := fmt.Sprintf("localhost:%d", nodePort)
	fmt.Println("Peripheral server is accessible at:", serviceAddress)

	return serviceAddress, nil
}

func StartCleanupTimer(userID string) {
	// Wait for 1 hour
	time.Sleep(5 * time.Minute)

	// Cleanup resources
	err := CleanupUserResources(userID)
	if err != nil {
		log.Println("Error cleaning up resources for user", userID+":", err)
	} else {
		log.Println("Successfully cleaned up resources for user", userID)
	}
}

func CleanupUserResources(userID string) error {
	var config *rest.Config
	var err error

	// Try in-cluster config
	config, err = rest.InClusterConfig()
	if err != nil {
		// If in-cluster config fails, fallback to kubeconfig
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return fmt.Errorf("failed to create Kubernetes config: %v", err)
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	namespace := "default"

	// Names of the resources
	deploymentName := "peripheral-server-deployment-" + userID
	serviceName := "peripheral-server-service-" + userID

	// Delete the deployment
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	fmt.Println("Deleting deployment for user:", userID)
	err = deploymentsClient.Delete(context.TODO(), deploymentName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %v", err)
	}

	// Delete the service
	servicesClient := clientset.CoreV1().Services(namespace)
	fmt.Println("Deleting service for user:", userID)
	err = servicesClient.Delete(context.TODO(), serviceName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete service: %v", err)
	}

	return nil
}

// Helper functions
func int32Ptr(i int32) *int32 { return &i }
