package serve

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kennyp/palette/cmd/palette/shared"
)

//go:embed templates/index.html
var indexHTML string

// handleIndex serves the main HTML page with HTMX and Alpine.js.
func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(indexHTML))
}

// handleConvert handles file upload and conversion.
func handleConvert(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (32 MB max)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get conversion parameters
	toFormat := r.FormValue("to")
	colorSpace := r.FormValue("colorspace")

	// Validate color space
	if colorSpace != "" {
		if err := shared.ValidateColorSpace(colorSpace); err != nil {
			http.Error(w, fmt.Sprintf("Invalid color space: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Validate target format
	if toFormat == "" {
		http.Error(w, "Target format is required", http.StatusBadRequest)
		return
	}

	// Detect source format from filename
	fromFormat := filepath.Ext(header.Filename)
	if fromFormat == "" {
		http.Error(w, "Cannot detect source format from filename", http.StatusBadRequest)
		return
	}

	// Create temporary input file
	tempInput, err := os.CreateTemp("", "palette-input-*"+fromFormat)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create temp file: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempInput.Name())
	defer tempInput.Close()

	// Copy uploaded file to temp file
	if _, err := io.Copy(tempInput, file); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save uploaded file: %v", err), http.StatusInternalServerError)
		return
	}
	tempInput.Close()

	// Create temporary output file
	tempOutput, err := os.CreateTemp("", "palette-output-*"+toFormat)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create output file: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempOutput.Name())
	tempOutput.Close()

	// Perform conversion
	if err := shared.ConvertFile(tempInput.Name(), tempOutput.Name(), fromFormat, toFormat, colorSpace); err != nil {
		http.Error(w, fmt.Sprintf("Conversion failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Read converted file
	outputData, err := os.ReadFile(tempOutput.Name())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read converted file: %v", err), http.StatusInternalServerError)
		return
	}

	// Determine output filename
	baseName := filepath.Base(header.Filename)
	outputName := baseName[:len(baseName)-len(filepath.Ext(baseName))] + toFormat

	// Send file as download
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", outputName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(outputData)))
	w.WriteHeader(http.StatusOK)
	w.Write(outputData)
}

// FormatInfo represents information about a supported format.
type FormatInfo struct {
	Extension   string `json:"extension"`
	Description string `json:"description"`
}

// handleFormats returns JSON array of supported formats.
func handleFormats(w http.ResponseWriter, r *http.Request) {
	formats := []FormatInfo{
		{Extension: ".acb", Description: "Adobe Color Book"},
		{Extension: ".aco", Description: "Adobe Color Swatch"},
		{Extension: ".csv", Description: "Comma-Separated Values"},
		{Extension: ".json", Description: "JSON"},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(formats)
}

// handleHealth returns a health check response.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
