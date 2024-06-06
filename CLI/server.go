package CLI

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var num int

func SetupRouter(port int, route string) {
	fmt.Println("CLI has successfully connected with your local app")
	fmt.Println("CLI is hosted at port :3000")

	// Calls whtestServerConnection which attempts to connect to the online hosted
	// server through websockets through which data is transferred between servers
	go whtestServerConnection(port, route)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

var handlerSwitch chan struct{} = make(chan struct{}, 1)

// DataHandler contains two handlers within itself, the data received initially is
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
// MessageTransfer is used to send an encoded message to the hosted server, which helps in
// identifying if the message is received by the CLI or not.
func MessageTransfer(conn *websocket.Conn) {
	fmt.Println("Inside MessageTransfer function")
	// Log the current connection state
	if conn == nil {
		log.Println("WebSocket connection is nil")
		return
	}

	encodedMessage := "EncodedMessage"
	fmt.Print("Encoded message is being transferred.\n")

	// Log before sending the message
	log.Println("Attempting to send message:", encodedMessage)

	if err := conn.WriteMessage(websocket.TextMessage, []byte(encodedMessage)); err != nil {
		log.Println("Error sending encoded message to whtest server:", err)
		return
	}

	// Log after successful send
	log.Println("Message sent successfully")
}

func whtestServerConnection(port int, route string) {
	fmt.Print("Hello, trying to connect to whtest_server.\n")

	URL := "ws://localhost:2000/whtest"

	conn, _, err := websocket.DefaultDialer.Dial(URL, nil)
	if err != nil {
		log.Println("WebSocket dial error:", err)
		return
	}

	fmt.Println("Successfully connected with whtest server")

	// Start data handling
	go DataHandler(conn, port, route)

	// Call MessageTransfer function
	fmt.Println("Calling MessageTransfer function")
	MessageTransfer(conn)
}
