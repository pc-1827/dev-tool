package whtest

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func SetupRouter() {
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		ForwardDataHandler(w, r)
	})

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

// // Handles receiving the webhook from the CLI
// func MessageAccepterHandler(conn *websocket.Conn) string {

// 	message := ""

// 	go func() {
// 		_, encodedMessage, err := conn.ReadMessage()
// 		if err != nil {
// 			log.Println(err)
// 		}

// 		message = string(encodedMessage)
// 		//fmt.Print(string(encodedMessage))
// 	}()

// 	fmt.Print("Received the encoded message.\n")
// 	fmt.Print(message)

// 	if message == "EncodedMessage" {
// 		testURL, _ := TestURLGenerator()
// 		fmt.Printf("%s this is the TestURL\n", testURL)
// 		TestURLTransfer(conn, testURL)
// 	}

// 	return ""
// 	//function which registers webhook at the third party site.
// }

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
				testURL, err := TestURLGenerator()
				if err != nil {
					log.Println("Error generating TestURL:", err)
					return
				}
				fmt.Printf("%s this is the TestURL\n", testURL)
				TestURLTransfer(conn, testURL)
			}
		}
	}()
}

// Sends TestURL to the CLI
func TestURLTransfer(conn *websocket.Conn, testURL string) {
	fmt.Print("TestURL is being transfered.\n")
	if err := conn.WriteMessage(websocket.TextMessage, []byte(testURL)); err != nil {
		log.Println("Error sending testURL to the CLI", err)
		return
	}
}

// TestURLGenerator used to random string for TestURL
// *Note: This will be probably replaced with pre generated custom subdomains.
func TestURLGenerator() (string, error) {
	fmt.Print("TestURL is being generated\n")
	byteSize := (6 + 1) / 2

	URlBytes := make([]byte, byteSize)
	_, err := rand.Read(URlBytes)
	if err != nil {
		return "", err
	}

	testURL := hex.EncodeToString(URlBytes)

	testURL = testURL[:6]

	return testURL, nil
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
		fmt.Print("Message received on /webhook and forwaded to CLI.\n")
	}()
}
