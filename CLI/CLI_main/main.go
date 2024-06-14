package main

// Main application user interface takes local server port and route as an input.
// *Note: Need to develop a UI similar to ngrok in future.

import (
	"CLI"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func main() {
	//fmt.Println("Welcome to webhook-tester CLI")

	var URL string
	fmt.Print("Please enter the address you would like to receive the webhook data:\n")
	fmt.Print("Example: http://localhost:5000/requests\n\n")
	fmt.Scanf("%d", &URL)

	u, err := url.Parse("http://localhost:5000/requests")
	if err != nil {
		panic(err)
	}

	hostParts := strings.Split(u.Host, ":")
	port := hostParts[1]

	path := u.Path

	portInt, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}

	CLI.SetupRouter(portInt, path)
}
