package whtest

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func SetupRouter() {
	http.HandleFunc("/whtest", func(w http.ResponseWriter, r *http.Request) {
		// if r.Method != http.MethodPost {
		// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		// }
		websocket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		var webhook string = WebhookAccepterHandler(websocket)
		log.Println(webhook)
	})
}

func WebhookAccepterHandler(conn *websocket.Conn) string {

	go func() {
		_, webhook, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
		}

		fmt.Print(webhook)
	}()

	fmt.Print("Received the webhook.\n")

	testURL, _ := TestURLGenerator()
	fmt.Printf("%s this is the TestURL\n", testURL)
	TestURLTransfer(conn, testURL)

	return ""
	//function which registers webhook at the third party site.
}

func TestURLTransfer(conn *websocket.Conn, testURL string) {
	fmt.Print("TestURL is being transfered.\n")
	if err := conn.WriteMessage(websocket.TextMessage, []byte(testURL)); err != nil {
		log.Println("Error sending testURL to the CLI", err)
		return
	}
}

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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
