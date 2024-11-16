package central

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
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

var subdomainTimers = make(map[string]*time.Timer)

func SubdomainAvailabilityChecker() string {
	subdomains := []string{"subdomain1.whtest.com", "subdomain2.whtest.com", "subdomain3.whtest.com"}

	for _, subdomain := range subdomains {
		timer, exists := subdomainTimers[subdomain]

		// If the timer for this subdomain doesn't exist or has expired, start a new one
		if !exists || timer == nil {
			// Timer either doesn't exist or has expired, so create a new one
			subdomainTimers[subdomain] = time.AfterFunc(1*time.Hour, func() {
				delete(subdomainTimers, subdomain)
			})

			return subdomain
		}
	}

	// If all subdomains are in use, return "None"
	return "None"
}

func SubdomainTransfer(conn *websocket.Conn) {
	fmt.Print("Subdomain is being transferred.\n")
	subdomain := SubdomainAvailabilityChecker()

	if err := conn.WriteMessage(websocket.TextMessage, []byte(subdomain)); err != nil {
		log.Println("Error sending subdomain to the CLI", err)
		return
	}
}
