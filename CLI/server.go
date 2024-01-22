package CLI

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

var num int

func SetupRouter(port int, endpoint string) {
	http.HandleFunc("/cli", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		forwardDataHandler(w, r, port, endpoint)
	})
	localServerURL := "http://localhost:" + strconv.Itoa(port) + "/" + endpoint

	fmt.Println("CLI has successfully connected with your local app")
	fmt.Println("CLI is hosted at port :3000")
	fmt.Printf("Data will forwaded to %s\n", localServerURL)
	http.ListenAndServe(":3000", nil)
}

func forwardDataHandler(w http.ResponseWriter, r *http.Request, port int, endpoint string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	localServerURL := "http://localhost:" + strconv.Itoa(port) + "/" + endpoint

	resp, err := http.Post(localServerURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Error forwarding request data", http.StatusInternalServerError)
		return
	}

	num = num + 1

	fmt.Printf("%d. Data received and forwarded to the local app\n", num)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Unexpected response status: %d", resp.StatusCode), http.StatusInternalServerError)
		return
	}
}
