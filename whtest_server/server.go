package whtest

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func SetupRouter() {
	http.HandleFunc("/whtest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		WebhookAccepterHandler(w, r)
	})
}

func WebhookAccepterHandler(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusInternalServerError)
		return
	}

	var webhook string
	err = json.Unmarshal(body, &webhook)
	if err != nil {
		http.Error(w, "Error unmarshalling JSON", http.StatusInternalServerError)
		return
	}

	TestURLTransfer(w, r)

	//function which registers webhook at the third party site.
}

func TestURLTransfer(w http.ResponseWriter, r *http.Request) {

}

func TestURLGenerator() (string, error) {
	byteSize := (6 + 1) / 2

	URlBytes := make([]byte, byteSize)
	_, err := rand.Read(URlBytes)
	if err != nil {
		return "", err
	}

	testURL := hex.EncodeToString(URlBytes)

	testURL = testURL[:6]

	return testURL, nil
}
