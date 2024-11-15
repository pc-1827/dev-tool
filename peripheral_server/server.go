package peripheral

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func SetupRouter() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		if dataForwading {
			ForwardDataHandler(w, r)
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

// Handles receiving the webhook from the CLI
func MessageAccepterHandler(conn *websocket.Conn) {
	var timer *time.Timer

	go func() {
		for {
			_, encodedMessage, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			message := string(encodedMessage)
			fmt.Print("Received the encoded message.\n")

			if message == "EncodedMessage" {
				dataForwading = true

				// If a timer already exists, stop it
				if timer != nil {
					timer.Stop()
				}

				// Set a new timer to disconnect after one hour
				timer = time.AfterFunc(1*time.Hour, func() {
					log.Println("Timer expired, disconnecting...")
					conn.Close()
				})
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
