package main

import (
	"fmt"
	"net/http"
	"whtest"
)

func main() {
	whtest.SetupRouter()

	fmt.Println("Server listening on :2000")
	http.ListenAndServe(":2000", nil)
}
