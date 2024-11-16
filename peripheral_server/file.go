package peripheral

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func StoreFileHandler(w http.ResponseWriter, r *http.Request) {
	// Read the file path from query parameters
	pathParam := r.URL.RawQuery // Get raw query to preserve encoding
	log.Println("Received Raw Query:", pathParam)
	values, err := url.ParseQuery(pathParam)
	if err != nil {
		log.Println("Error parsing query parameters:", err)
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}
	filePath := values.Get("path")
	if filePath == "" {
		log.Println("Missing file path in query parameters")
		http.Error(w, "Missing file path", http.StatusBadRequest)
		return
	}
	log.Println("File path to store:", filePath)

	// Read the file data from the request body
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading file data:", err)
		http.Error(w, "Error reading file data", http.StatusInternalServerError)
		return
	}
	log.Println("Read", len(data), "bytes from request body")

	// Construct the full file path
	fullFilePath := filepath.Join(staticDir, filePath)
	log.Println("Full file path:", fullFilePath)

	// Create the directory structure if it doesn't exist
	err = os.MkdirAll(filepath.Dir(fullFilePath), os.ModePerm)
	if err != nil {
		log.Println("Error creating directories:", err)
		http.Error(w, "Error creating directories", http.StatusInternalServerError)
		return
	}

	// Write the file
	err = os.WriteFile(fullFilePath, data, os.ModePerm)
	if err != nil {
		log.Println("Error writing file:", err)
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}

	log.Printf("File %s stored successfully", filePath)
	fmt.Fprintf(w, "File %s stored successfully", filePath)
}

func ServeFileHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path
	if filePath == "/" {
		filePath = "/index.html"
	}
	fullFilePath := filepath.Join(staticDir, filePath)
	// Check if file exists
	_, err := os.Stat(fullFilePath)
	if os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, fullFilePath)
}
