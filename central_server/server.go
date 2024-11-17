package central

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

	// Deploy peripheral server to Kubernetes and get the service address
	serviceAddress, err := DeployPeripheralServer()
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
}

func DeployPeripheralServer() (string, error) {
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

	// Create a unique name for the deployment and service
	deploymentName := "peripheral-server-deployment"
	serviceName := "peripheral-server-service"

	// Define the deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "peripheral-server",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "peripheral-server",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "peripheral-server",
							Image: "peripheral_server:latest", // Ensure this image is available in the cluster
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

	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	_, err = deploymentsClient.Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		fmt.Println("Creating deployment...")
		_, err = deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to create deployment: %v", err)
		}
	} else {
		fmt.Println("Deployment already exists.")
	}

	// Define the service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "peripheral-server",
			},
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       2001,
					TargetPort: intstr.FromInt(2001),
				},
			},
			Type: corev1.ServiceTypeNodePort, // Use NodePort to expose the service
		},
	}

	servicesClient := clientset.CoreV1().Services(namespace)
	svc, err := servicesClient.Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		fmt.Println("Creating service...")
		svc, err = servicesClient.Create(context.TODO(), service, metav1.CreateOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to create service: %v", err)
		}
	} else {
		fmt.Println("Service already exists.")
	}

	// Get the NodePort assigned
	nodePort := svc.Spec.Ports[0].NodePort

	// Get the node IP address
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil || len(nodes.Items) == 0 {
		return "", fmt.Errorf("failed to get node information: %v", err)
	}

	// For Minikube, get the IP from minikube
	nodeIP, err := getMinikubeIP()
	if err != nil {
		return "", fmt.Errorf("failed to get Minikube IP: %v", err)
	}

	serviceAddress := fmt.Sprintf("%s:%d", nodeIP, nodePort)
	fmt.Println("Peripheral server is accessible at:", serviceAddress)

	return serviceAddress, nil
}

func getMinikubeIP() (string, error) {
	cmd := exec.Command("minikube", "ip")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Helper functions
func int32Ptr(i int32) *int32 { return &i }

func intstrPtr(i int32) intstr.IntOrString {
	return intstr.IntOrString{Type: intstr.Int, IntVal: i}
}

// var subdomainTimers = make(map[string]*time.Timer)

// func SubdomainAvailabilityChecker() string {
// 	subdomains := []string{"subdomain1.whtest.com", "subdomain2.whtest.com", "subdomain3.whtest.com"}

// 	for _, subdomain := range subdomains {
// 		timer, exists := subdomainTimers[subdomain]

// 		// If the timer for this subdomain doesn't exist or has expired, start a new one
// 		if !exists || timer == nil {
// 			// Timer either doesn't exist or has expired, so create a new one
// 			subdomainTimers[subdomain] = time.AfterFunc(1*time.Hour, func() {
// 				delete(subdomainTimers, subdomain)
// 			})

// 			return subdomain
// 		}
// 	}

// 	// If all subdomains are in use, return "None"
// 	return "None"
// }

// func SubdomainTransfer(conn *websocket.Conn) {
// 	fmt.Print("Subdomain is being transferred.\n")
// 	subdomain := SubdomainAvailabilityChecker()

// 	if err := conn.WriteMessage(websocket.TextMessage, []byte(subdomain)); err != nil {
// 		log.Println("Error sending subdomain to the CLI", err)
// 		return
// 	}
// }
