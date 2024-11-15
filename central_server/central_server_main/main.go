package main

import (
	"central"
	"fmt"
	"net/http"
)

func main() {
	central.SetupRouter()

	fmt.Println("Server listening on :2000")
	http.ListenAndServe(":2000", nil)
}
