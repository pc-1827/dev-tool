package main

// Main application user interface takes local server port and route as an input.
// *Note: Need to develop a UI similar to ngrok in future.

import (
	"CLI"
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please choose an option:")
	fmt.Println("1. Webhook Testing")
	fmt.Println("2. Website Demo")
	fmt.Print("Enter your choice (1 or 2): ")

	choiceStr, _ := reader.ReadString('\n')
	choiceStr = strings.TrimSpace(choiceStr)
	choice, err := strconv.Atoi(choiceStr)
	if err != nil {
		fmt.Println("Invalid choice. Please enter 1 or 2.")
		return
	}

	var number string
	if choice == 1 {
		number = "1"
		fmt.Print("Please enter the address you would like to receive the webhook data:\n")
		fmt.Print("Example: http://localhost:5000/requests\n\n")
		urlStr, _ := reader.ReadString('\n')
		urlStr = strings.TrimSpace(urlStr)

		u, err := url.Parse(urlStr)
		if err != nil {
			panic(err)
		}

		hostParts := strings.Split(u.Host, ":")
		if len(hostParts) != 2 {
			fmt.Println("Invalid host format. Expected host:port")
			return
		}
		port := hostParts[1]

		path := u.Path

		portInt, err := strconv.Atoi(port)
		if err != nil {
			panic(err)
		}

		CLI.SetupRouter(portInt, path, number)
	} else if choice == 2 {
		number = "2"
		CLI.SetupRouter(0, "", number)
	} else {
		fmt.Println("Invalid choice. Please enter 1 or 2.")
	}
}
