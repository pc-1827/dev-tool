package CLI

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

var num int

func SetupRouter(port int, route string, number string) {
	fmt.Println("CLI has successfully connected with your local app")
	fmt.Println("Webhook tester is hosted at port :8000")

	// Calls whtestServerConnection which attempts to connect to the online hosted
	// server through websockets through which data is transferred between servers
	go whtestServerConnection("ws://localhost:2000/whtest", port, route, number)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

var subdomainReceived = false

func SubdomainHandler(conn *websocket.Conn, port int, route string, number string) {
	fmt.Print("Attempting to receive Subdomain.\n")
	_, addressBytes, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Error receiving service address:", err)
		return
	}

	serviceAddress := string(addressBytes)

	if serviceAddress == "None" {
		fmt.Println("No service address received")
	} else {
		conn.Close()
		subdomainReceived = true
		go whtestServerConnection(serviceAddress, port, route, number)

		localServerURL := "http://localhost:" + strconv.Itoa(port) + route

		fmt.Printf("WebSocket traffic will be transferred from %s ---> %s\n", serviceAddress, localServerURL)
	}

}

// After successfully establishing a websocket connection between CLI and server hosted
// MessageTransfer is used to send an encoded message to the hosted server, which helps in
// identifying if the message is received by the CLI or not.
func MessageTransfer(conn *websocket.Conn, number string) {
	fmt.Println("Inside MessageTransfer function")
	// Log the current connection state
	if conn == nil {
		log.Println("WebSocket connection is nil")
		return
	}

	encodedMessage := "EncodedMessage"

	// Create the message by combining encodedMessage and number
	message := encodedMessage + ":" + number

	fmt.Print("Encoded message is being transferred.\n")

	// Log before sending the message
	log.Println("Attempting to send message:", message)

	if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		log.Println("Error sending encoded message to whtest server:", err)
		return
	}

	// Log after successful send
	log.Println("Message sent successfully")
}

func whtestServerConnection(URL string, port int, route string, number string) {
	fmt.Print("Hello, trying to connect to whtest_server.\n")

	// Ensure the URL has the correct scheme
	if !strings.HasPrefix(URL, "ws://") && !strings.HasPrefix(URL, "wss://") {
		URL = "ws://" + URL
	}

	conn, _, err := websocket.DefaultDialer.Dial(URL, nil)
	if err != nil {
		log.Println("WebSocket dial error:", err)
		return
	}

	fmt.Println("Successfully connected with whtest server")

	// Call MessageTransfer function
	fmt.Println("Calling MessageTransfer function")
	MessageTransfer(conn, number)

	if number == "1" {
		// For webhook testing
		if !subdomainReceived {
			SubdomainHandler(conn, port, route, number)
		} else {
			DataTransferHandler(conn, port, route)
		}
	} else if number == "2" {
		// For website demo
		// For webhook testing
		if !subdomainReceived {
			SubdomainHandler(conn, port, route, number)
		}
	}
}
