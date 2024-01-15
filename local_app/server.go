package localapp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type Request struct {
	ID          int       `json:"id"`
	RequestData string    `json:"request_data"`
	RequestTime time.Time `json:"request_time"`
}

var Router = http.NewServeMux()

func SetupRouter(db *sql.DB) {
	// Use http.HandleFunc to create handlers for routes
	Router.HandleFunc("/requests", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetRequests(w, db)
		case http.MethodPost:
			RecordRequest(w, r, db)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func StartServer() {
	fmt.Println("Server listening on :5000")
	http.ListenAndServe(":5000", Router)
}

func GetRequests(w http.ResponseWriter, db *sql.DB) {
	rows, err := db.Query("SELECT * FROM requests ORDER BY id DESC;")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying the database: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var requests []Request

	for rows.Next() {
		var req Request
		err := rows.Scan(&req.ID, &req.RequestData, &req.RequestTime)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
			return
		}
		requests = append(requests, req)
	}

	w.Header().Set("Content-Type", "application/json")

	for _, req := range requests {
		rowJSON, err := json.Marshal(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error converting row to JSON: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write(rowJSON)
		w.Write([]byte("\n"))
		w.Write([]byte("\n"))
	}
}

func RecordRequest(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusInternalServerError)
		return
	}

	// Insert the data into the 'requests' table
	_, err = db.Exec("INSERT INTO requests (request_data) VALUES ($1)", string(body))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting data into database: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Request recorded successfully")
}
