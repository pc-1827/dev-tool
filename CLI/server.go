package CLI

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var num int

func SetupRouter(port int, route string, webhook string) {
	fmt.Println("CLI has successfully connected with your local app")
	fmt.Println("CLI is hosted at port :3000")

	//  Calls whtestServerConnection which attempts to connect to the online hosted
	//  server through websockets through which data is transferred between servers
	whtestServerConnection(webhook, port, route)
	http.ListenAndServe(":3000", nil)

}

var handlerSwitch chan struct{} = make(chan struct{}, 1)

// DataHandler contains two handlers within itself, the data received initally is
// handled by TestURLHandler(gives user information about the test URL), after that
// all data is handled by DataTransferHandler(which transfers data to local server).
func DataHandler(conn *websocket.Conn, port int, route string) {
	TestURLHandler(conn, port, route)
	handlerSwitch <- struct{}{} // Signal the switch to DataTransferHandler
	DataTransferHandler(conn, port, route)
}

func TestURLHandler(conn *websocket.Conn, port int, route string) {
	fmt.Print("Attempting to receive TestURL.\n")
	_, testURL, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Error receiving Test URL:", err)
		return
	}

	localServerURL := "http://localhost:" + strconv.Itoa(port) + "/" + route
	testServerURL := "http://" + string(testURL) + ".whtest.dev"

	fmt.Printf("WebSocket traffic will be transferred from %s ---> %s\n", testServerURL, localServerURL)
}

// After successfully establishing a websocket connection between CLI and server hosted
// online WebhookTransfer is called which sends webhook to the online server.
func WebhookTransfer(conn *websocket.Conn, webhook string) {
	fmt.Print("Webhook being transferred.\n")
	if err := conn.WriteMessage(websocket.TextMessage, []byte(webhook)); err != nil {
		log.Println("Error sending webhook to whtest server", err)
		return
	}
}

func whtestServerConnection(webhook string, port int, route string) {
	fmt.Print("Hello, trying to connect to whtest_server.\n")

	URL := "ws://localhost:2000/whtest"

	conn, _, err := websocket.DefaultDialer.Dial(URL, nil)
	if err != nil {
		log.Println("WebSocket dial error:", err)
		return
	}

	fmt.Println("Successfully connected with whtest server")
	DataHandler(conn, port, route)
	WebhookTransfer(conn, webhook)
}
