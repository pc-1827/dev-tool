package CLI

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func ForwardDataHandler(w http.ResponseWriter, r *http.Request, port int, route string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	localServerURL := "http://localhost:" + strconv.Itoa(port) + "/" + route

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
