package main

import (
	"fmt"
	"net/http"
	"peripheral"
)

func main() {
	peripheral.SetupRouter()

	fmt.Println("Server listening on :2001")
	http.ListenAndServe(":2001", nil)
}
