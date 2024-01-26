package whtest

import (
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

	TestURLGenerator(w, r)

	//function which registers webhook at the third party site.
}

func TestURLGenerator(w http.ResponseWriter, r *http.Request) {

}
