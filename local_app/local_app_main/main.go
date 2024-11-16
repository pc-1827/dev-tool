package main

import (
	"bytes"
	"fmt"
	localapp "local_app"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func main() {
	db, err := localapp.Connect()

	if err != nil {
		db.Close()
		log.Fatal("Failed to connect to the database:", err)
	}
	defer db.Close()

	// Check command-line arguments
	if len(os.Args) > 1 {
		number := os.Args[1]
		if number == "2" {
			SendStaticFilesToPeripheralServer()
		}
	}

	localapp.SetupRouter(db)
}

func SendStaticFilesToPeripheralServer() {
	staticDir := "/home/pc1827/projects/webhook-tester/local_app/static" // Adjust if necessary

	err := filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// Read the file contents
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Get the relative file path
			relativePath, err := filepath.Rel(staticDir, path)
			if err != nil {
				return err
			}

			// Prepare the URL with file path as query parameter
			urlStr := fmt.Sprintf("http://localhost:2001/?path=%s", url.QueryEscape(relativePath))

			// Send the data via HTTP POST
			resp, err := http.Post(urlStr, "application/octet-stream", bytes.NewBuffer(data))
			if err != nil {
				fmt.Printf("Error sending file %s: %v\n", relativePath, err)
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Failed to send file %s: status code %d\n", relativePath, resp.StatusCode)
				return fmt.Errorf("failed to send file %s", relativePath)
			}

			fmt.Printf("Sent file: %s\n", relativePath)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error sending static files: %v\n", err)
	}
}
