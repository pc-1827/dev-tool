package CLI_test

import (
	"CLI"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestForwardRequests(t *testing.T) {

	stopServer := make(chan struct{})

	go func() {
		CLI.SetupRouter(5000, "requests")
	}()

	defer func() {
		close(stopServer)
	}()

	time.Sleep(500 * time.Millisecond)

	payload := []byte(`{"test": "testing /cli POST request endpoint"}`)
	request, err := http.NewRequest(http.MethodPost, "/cli", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatalf(err.Error())
	}

	response := httptest.NewRecorder()

	http.DefaultServeMux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, response.Code)
	}
}
