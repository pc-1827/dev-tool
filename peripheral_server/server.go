package peripheral

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Println("Error getting current working directory:", err)
	} else {
		log.Println("Current working directory:", cwd)
	}
}

func SetupRouter() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if websiteDemoMode {
			if r.Method == http.MethodPost {
				StoreFileHandler(w, r)
			} else if r.Method == http.MethodGet {
				ServeFileHandler(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			if dataForwading {
				ForwardDataHandler(w, r)
			} else {
				http.Error(w, "Data forwarding not enabled", http.StatusBadRequest)
			}
		}
	})

	http.HandleFunc("/subdomain", func(w http.ResponseWriter, r *http.Request) {
		websocket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		setWebSocketConnection(websocket)
		MessageAccepterHandler(websocket)
	})
}

var dataForwading = false
var websiteDemoMode = false
var staticDir = "/home/pc1827/projects/webhook-tester/peripheral_server/static"

// Handles receiving the webhook from the CLI
func MessageAccepterHandler(conn *websocket.Conn) {
	var timer *time.Timer

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
				if number == "1" {
					// Webhook Testing Mode
					dataForwading = true
					websiteDemoMode = false

					// If a timer already exists, stop it
					if timer != nil {
						timer.Stop()
					}

					// Set a new timer to disconnect after one hour
					timer = time.AfterFunc(1*time.Hour, func() {
						log.Println("Timer expired, disconnecting...")
						conn.Close()
					})
				} else if number == "2" {
					// Website Demo Mode
					websiteDemoMode = true
					dataForwading = false
					fmt.Println("Switching to Website Demo mode")
					// Close the connection since we only needed to receive the message
					conn.Close()
				} else {
					fmt.Println("Invalid number received")
				}
			}
		}
	}()
}

func ForwardDataHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	go func() {
		conn, err := waitForConnection()
		if err != nil {
			log.Println(err)
			return
		}

		if err := conn.WriteMessage(websocket.TextMessage, []byte(body)); err != nil {
			log.Println("Error sending webhook to whtest server", err)
			return
		}
		fmt.Print("Message received and forwaded to CLI.\n")
	}()
}
