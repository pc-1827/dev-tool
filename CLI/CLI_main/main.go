package main

// Main application user interface takes local server port and route as an input.
// *Note: Need to develop a UI similar to ngrok in future.

import (
	"CLI"
	"fmt"
)

func main() {
	fmt.Println("Welcome to webhook-tester CLI")

	var port int
	fmt.Print("Please enter the port at which your local app is hosted: ")
	fmt.Scanf("%d", &port)

	var route string
	fmt.Print("Please enter the route at which you would like to receive data: ")
	fmt.Print("Please enter the route without '/'")
	fmt.Scanf("%s", &route)

	// var webhook string
	// fmt.Print("Please enter the webhook from which you would like to recieve data: ")
	// fmt.Scanf("%s", &webhook)

	CLI.SetupRouter(port, route)
}
