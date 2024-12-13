package central

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func SetupRouter() {
	http.HandleFunc("/whtest", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		MessageAccepterHandler(conn)
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

func SubdomainTransfer(conn *websocket.Conn) {
	fmt.Print("Starting dynamic provisioning of peripheral server.\n")

	subdomain := GenerateRandomString(10)

	ingressIP, err := GetIngressControllerIP()
	if err != nil {
		log.Println("Error getting Ingress Controller IP:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("None"))
		return
	}

	err = DeployPeripheralServer(subdomain)
	if err != nil {
		log.Println("Error deploying peripheral server:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("None"))
		return
	}

	err = CreateDNSRecord(subdomain, ingressIP)
	if err != nil {
		log.Println("Error creating DNS record:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("None"))
		return
	}

	hostName := subdomain + ".<your-domain>"

	if err := conn.WriteMessage(websocket.TextMessage, []byte(hostName)); err != nil {
		log.Println("Error sending subdomain to the CLI:", err)
		return
	}

	go StartCleanupTimer(subdomain)
}

func DeployPeripheralServer(subdomain string) error {
	clientset, err := getKubernetesClient()
	if err != nil {
		return err
	}

	namespace := "default"
	podName := "peripheral-server-pod-" + subdomain
	serviceName := "peripheral-server-service-" + subdomain

	labels := map[string]string{
		"app":       "peripheral-server",
		"subdomain": subdomain,
	}

	// Define the Pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   podName,
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "peripheral-server",
					Image: "<your-acr-name>.azurecr.io/peripheral-server:latest",
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 2001,
						},
					},
				},
			},
		},
	}

	// Create the Pod
	fmt.Println("Creating pod for subdomain:", subdomain)
	_, err = clientset.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create pod: %v", err)
	}

	// Define the Service
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
					Port:       80,
					TargetPort: intstr.FromInt(2001),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// Create the Service
	fmt.Println("Creating service for subdomain:", subdomain)
	_, err = clientset.CoreV1().Services(namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create service: %v", err)
	}

	// Create the Ingress
	err = CreateIngress(subdomain, serviceName, labels)
	if err != nil {
		return fmt.Errorf("failed to create ingress: %v", err)
	}

	return nil
}

func CreateIngress(subdomain, serviceName string, labels map[string]string) error {
	clientset, err := getKubernetesClient()
	if err != nil {
		return err
	}

	namespace := "default"
	ingressName := "peripheral-server-ingress-" + subdomain
	host := subdomain + ".<your-domain>"

	pathType := networkingv1.PathTypePrefix

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ingressName,
			Labels: labels,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "nginx",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: host,
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
	_, err = clientset.NetworkingV1().Ingresses(namespace).Create(context.TODO(), ingress, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create ingress: %v", err)
	}

	return nil
}

func StartCleanupTimer(subdomain string) {
	// Wait for 1 hour
	time.Sleep(1 * time.Hour)
	// Clean up resources
	ingressIP, err := GetIngressControllerIP()
	if err != nil {
		log.Println("Error getting Ingress Controller IP during cleanup:", err)
		return
	}
	err = CleanupUserResources(subdomain, ingressIP)
	if err != nil {
		log.Println("Error cleaning up resources:", err)
	}
}

func CleanupUserResources(subdomain, ingressIP string) error {
	clientset, err := getKubernetesClient()
	if err != nil {
		return err
	}

	namespace := "default"
	podName := "peripheral-server-pod-" + subdomain
	serviceName := "peripheral-server-service-" + subdomain
	ingressName := "peripheral-server-ingress-" + subdomain

	// Delete the Pod
	fmt.Println("Deleting pod for subdomain:", subdomain)
	err = clientset.CoreV1().Pods(namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
	if err != nil {
		log.Println("Failed to delete pod:", err)
	}

	// Delete the Service
	fmt.Println("Deleting service for subdomain:", subdomain)
	err = clientset.CoreV1().Services(namespace).Delete(context.TODO(), serviceName, metav1.DeleteOptions{})
	if err != nil {
		log.Println("Failed to delete service:", err)
	}

	// Delete the Ingress
	fmt.Println("Deleting ingress for subdomain:", subdomain)
	err = clientset.NetworkingV1().Ingresses(namespace).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		log.Println("Failed to delete ingress:", err)
	}

	// Delete the DNS record
	err = DeleteDNSRecord(subdomain)
	if err != nil {
		log.Println("Error deleting DNS record for", subdomain+":", err)
	}

	return nil
}
