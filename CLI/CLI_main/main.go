package main

import (
	"CLI"
	"fmt"
)

func main() {
	fmt.Println("Welcome to webhook-tester CLI")

	var port int
	fmt.Print("Please enter the port at which your local app is hosted: ")
	fmt.Scanf("%d", &port)

	var endpoint string
	fmt.Print("Please enter the endpoint at which you would like to receive data: ")
	fmt.Scanf("%s", &endpoint)

	CLI.SetupRouter(port, endpoint)
}
