package whtest

import (
	"fmt"
	"log"
	"net/http"
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
			_, encodedMessage, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			message := string(encodedMessage)
			fmt.Print("Received the encoded message.\n")

			if message == "EncodedMessage" {
				SubdomainTransfer(conn)
			}
		}
	}()
}

var subdomainTimers = make(map[string]*time.Timer)

func SubdomainAvailabilityChecker(conn *websocket.Conn) string {
	subdomains := []string{"subdomain1.whtest.com", "subdomain2.whtest.com", "subdomain3.whtest.com"}

	for _, subdomain := range subdomains {
		timer, exists := subdomainTimers[subdomain]

		// If the timer for this subdomain doesn't exist or has expired, start a new one
		if !exists || !timer.Reset(0) {
			subdomainTimers[subdomain] = time.AfterFunc(1*time.Hour, func() {
				delete(subdomainTimers, subdomain)
			})

			return subdomain
		}
	}

	// If all subdomains are in use, return an empty string
	return "None"
}

func SubdomainTransfer(conn *websocket.Conn) {
	fmt.Print("Subdomain is being transferred.\n")
	subdomain := SubdomainAvailabilityChecker(conn)

	if err := conn.WriteMessage(websocket.TextMessage, []byte(subdomain)); err != nil {
		log.Println("Error sending subdomain to the CLI", err)
		return
	}
}
