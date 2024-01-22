package localapp_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	localapp "local_app"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGETRequests(t *testing.T) {
	db, err := localapp.Connect()
	if err != nil {
		db.Close()
		t.Errorf("unable to connect to DB got error:%q", err.Error())
	}
	insertIntoDB(t, db)

	t.Run("returns Traffic in database", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/requests", nil)
		response := httptest.NewRecorder()

		fmt.Printf("Request URL: %s\n", request.URL.Path)

		assertResponseStatus(t, response.Code, http.StatusOK)

		cleanedBody := strings.ReplaceAll(response.Body.String(), "\n", "")

		var requests []localapp.Request
		err := json.Unmarshal([]byte(cleanedBody), &requests)
		if err != nil {
			t.Fatalf("error unmarshalling JSON: %v", err)
		}

		assertResponseBody(t, len(requests), 3)
	})
}

func insertIntoDB(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec("TRUNCATE requests RESTART IDENTITY")
	if err != nil {
		db.Close()
		t.Errorf("unable to truncate requests Table:%q", err.Error())
	}
	_, err = db.Exec("INSERT INTO requests (request_data) VALUES ('{\"name\": \"John\", \"age\": 25, \"city\": \"New York\"}'),('{\"name\": \"Alice\", \"age\": 30, \"city\": \"London\"}'),('{\"name\": \"Bob\", \"age\": 22, \"city\": \"Paris\"}');")
	if err != nil {
		db.Close()
		t.Errorf("unable to insert data into the requests table")
	}
}

func assertResponseBody(t *testing.T, got int, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertResponseStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
}
