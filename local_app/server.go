package localapp

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
)

type Request struct {
	ID          int       `json:"id"`
	RequestData string    `json:"request_data"`
	RequestTime time.Time `json:"request_time"`
}

func SetupRouter(db *sql.DB) {

	http.HandleFunc("/requests", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetRequests(w, db)
		case http.MethodPost:
			RecordRequest(w, r, db)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	staticDir := "/home/pc1827/projects/website-hoster/local_app/static"
	url := "http://localhost:3000/"

	err := SendStaticFiles(staticDir, url)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Server listening on :5000")
	http.ListenAndServe(":5000", nil)
}

// SendStaticFiles sends all files in the specified directory to the given URL using POST requests
func SendStaticFiles(dirPath, url string) error {
	fmt.Println("SendStaticFiles function is being called")

	// Walk through all files in the directory
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process files, not directories
		if !info.IsDir() {
			err := sendFile(path, url)
			if err != nil {
				fmt.Printf("Error sending file %s: %v\n", path, err)
			} else {
				fmt.Printf("Successfully sent file %s\n", path)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking the directory: %v", err)
	}

	return nil
}

// sendFile sends a single file to the specified URL
func sendFile(filePath, url string) error {
	fmt.Printf("Sending file: %s\n", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("error creating form file: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("error copying file content: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("error closing writer: %v", err)
	}

	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}
	fmt.Println("Response:", string(responseBody))

	return nil
}

// Fetches the data from the database and displays it in the browser for ease of development.
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

// Inserts data received into the database.
func RecordRequest(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO requests (request_data) VALUES ($1)", string(body))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting data into database: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Request recorded successfully")
}
