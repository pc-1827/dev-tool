package main

import (
	"fmt"
	"net/http"
	"subdomains"
)

func main() {
	subdomains.SetupRouter()

	fmt.Println("Server listening on :2001")
	http.ListenAndServe(":2001", nil)
}
