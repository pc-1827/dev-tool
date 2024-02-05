package whtest

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	wsConn *websocket.Conn
	connMu sync.Mutex
)

func SetupRouter() {
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		ForwardDataHandler(w, r)
	})

	http.HandleFunc("/whtest", func(w http.ResponseWriter, r *http.Request) {
		// if r.Method != http.MethodPost {
		// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		// }
		websocket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		setWebSocketConnection(websocket)
		var webhook string = WebhookAccepterHandler(websocket)
		fmt.Print(string(webhook))
	})
}

func WebhookAccepterHandler(conn *websocket.Conn) string {

	go func() {
		_, webhook, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
		}

		fmt.Print(string(webhook))
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

func waitForConnection() (*websocket.Conn, error) {
	for {
		conn, err := getWebSocketConnection()
		if err != nil {
			return nil, err
		}
		if conn != nil {
			return conn, nil
		}
		time.Sleep(time.Millisecond * 100) // Adjust the sleep duration based on your needs
	}
}

func setWebSocketConnection(conn *websocket.Conn) {
	connMu.Lock()
	defer connMu.Unlock()
	wsConn = conn
}

func getWebSocketConnection() (*websocket.Conn, error) {
	connMu.Lock()
	defer connMu.Unlock()
	return wsConn, nil
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
