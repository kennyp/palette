package serve

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/render"
	"github.com/kennyp/palette/cmd/palette/shared"
)

//go:embed templates/index.html
var indexHTML string

// ErrResponse represents an error response.
type ErrResponse struct {
	HTTPStatusCode int    `json:"-"`
	StatusText     string `json:"status"`
	ErrorText      string `json:"error"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// handleIndex serves the main HTML page with HTMX and Alpine.js.
func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(indexHTML))
}

// ConvertFormRequest represents a multipart form conversion request.
type ConvertFormRequest struct {
	To         string `form:"to"`
	ColorSpace string `form:"colorspace"`
}

// handleConvert handles file upload and conversion using chi render.
func handleConvert(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (32 MB max)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			StatusText:     "Invalid request",
			ErrorText:      fmt.Sprintf("Failed to parse form: %v", err),
		})
		return
	}

	// Get form values
	data := &ConvertFormRequest{
		To:         r.FormValue("to"),
		ColorSpace: r.FormValue("colorspace"),
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			StatusText:     "Invalid request",
			ErrorText:      fmt.Sprintf("Failed to get file: %v", err),
		})
		return
	}
	defer file.Close()

	// Validate color space
	if data.ColorSpace != "" {
		if err := shared.ValidateColorSpace(data.ColorSpace); err != nil {
			render.Render(w, r, &ErrResponse{
				HTTPStatusCode: http.StatusBadRequest,
				StatusText:     "Invalid request",
				ErrorText:      fmt.Sprintf("Invalid color space: %v", err),
			})
			return
		}
	}

	// Validate target format
	if data.To == "" {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			StatusText:     "Invalid request",
			ErrorText:      "Target format is required",
		})
		return
	}

	// Detect source format from filename
	fromFormat := filepath.Ext(header.Filename)
	if fromFormat == "" {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			StatusText:     "Invalid request",
			ErrorText:      "Cannot detect source format from filename",
		})
		return
	}

	// Create temporary input file
	tempInput, err := os.CreateTemp("", "palette-input-*"+fromFormat)
	if err != nil {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusInternalServerError,
			StatusText:     "Server error",
			ErrorText:      fmt.Sprintf("Failed to create temp file: %v", err),
		})
		return
	}
	defer os.Remove(tempInput.Name())
	defer tempInput.Close()

	// Copy uploaded file to temp file
	if _, err := io.Copy(tempInput, file); err != nil {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusInternalServerError,
			StatusText:     "Server error",
			ErrorText:      fmt.Sprintf("Failed to save uploaded file: %v", err),
		})
		return
	}
	tempInput.Close()

	// Create temporary output file
	tempOutput, err := os.CreateTemp("", "palette-output-*"+data.To)
	if err != nil {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusInternalServerError,
			StatusText:     "Server error",
			ErrorText:      fmt.Sprintf("Failed to create output file: %v", err),
		})
		return
	}
	defer os.Remove(tempOutput.Name())
	tempOutput.Close()

	// Perform conversion
	if err := shared.ConvertFile(tempInput.Name(), tempOutput.Name(), fromFormat, data.To, data.ColorSpace); err != nil {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusInternalServerError,
			StatusText:     "Conversion failed",
			ErrorText:      fmt.Sprintf("%v", err),
		})
		return
	}

	// Read converted file
	outputData, err := os.ReadFile(tempOutput.Name())
	if err != nil {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusInternalServerError,
			StatusText:     "Server error",
			ErrorText:      fmt.Sprintf("Failed to read converted file: %v", err),
		})
		return
	}

	// Determine output filename
	baseName := filepath.Base(header.Filename)
	outputName := baseName[:len(baseName)-len(filepath.Ext(baseName))] + data.To

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

// FormatsList is a list of formats for rendering.
type FormatsList struct {
	Formats []FormatInfo `json:"formats"`
}

func (f *FormatsList) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// handleFormats returns JSON array of supported formats.
func handleFormats(w http.ResponseWriter, r *http.Request) {
	formats := []FormatInfo{
		{Extension: ".acb", Description: "Adobe Color Book"},
		{Extension: ".aco", Description: "Adobe Color Swatch"},
		{Extension: ".csv", Description: "Comma-Separated Values"},
		{Extension: ".json", Description: "JSON"},
	}

	render.Render(w, r, &FormatsList{Formats: formats})
}

// HealthResponse represents a health check response.
type HealthResponse struct {
	Status string `json:"status"`
}

func (h *HealthResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// handleHealth returns a health check response.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, &HealthResponse{Status: "ok"})
}

// ConvertRequest represents a JSON API conversion request.
type ConvertRequest struct {
	FileContent []byte `json:"file_content"` // Base64 encoded file content (auto-decoded)
	From        string `json:"from"`         // Source format (.acb, .aco, .csv, .json)
	To          string `json:"to"`           // Target format
	ColorSpace  string `json:"colorspace,omitempty"`
}

// Bind implements render.Binder interface for request validation.
func (c *ConvertRequest) Bind(r *http.Request) error {
	if len(c.FileContent) == 0 {
		return fmt.Errorf("file_content is required")
	}
	if c.To == "" {
		return fmt.Errorf("to format is required")
	}
	if c.ColorSpace != "" {
		if err := shared.ValidateColorSpace(c.ColorSpace); err != nil {
			return err
		}
	}
	return nil
}

// ConvertResponse represents a JSON API conversion response.
type ConvertResponse struct {
	FileContent []byte `json:"file_content"` // Base64 encoded converted file (auto-encoded)
	Filename    string `json:"filename"`
	Format      string `json:"format"`
}

// Render implements render.Renderer interface.
func (c *ConvertResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// handleConvertJSON handles JSON API conversion requests.
func handleConvertJSON(w http.ResponseWriter, r *http.Request) {
	data := &ConvertRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			StatusText:     "Invalid request",
			ErrorText:      err.Error(),
		})
		return
	}

	// Ensure formats have leading dots
	fromFormat := data.From
	if fromFormat == "" {
		fromFormat = ".json" // Default to JSON if not specified
	}
	if !strings.HasPrefix(fromFormat, ".") {
		fromFormat = "." + fromFormat
	}
	toFormat := data.To
	if !strings.HasPrefix(toFormat, ".") {
		toFormat = "." + toFormat
	}

	// Create temporary input file
	tempInput, err := os.CreateTemp("", "palette-json-input-*"+fromFormat)
	if err != nil {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusInternalServerError,
			StatusText:     "Server error",
			ErrorText:      fmt.Sprintf("Failed to create temp file: %v", err),
		})
		return
	}
	defer os.Remove(tempInput.Name())
	defer tempInput.Close()

	// Write file content to temp file
	if _, err := tempInput.Write(data.FileContent); err != nil {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusInternalServerError,
			StatusText:     "Server error",
			ErrorText:      fmt.Sprintf("Failed to write temp file: %v", err),
		})
		return
	}
	tempInput.Close()

	// Create temporary output file
	tempOutput, err := os.CreateTemp("", "palette-json-output-*"+toFormat)
	if err != nil {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusInternalServerError,
			StatusText:     "Server error",
			ErrorText:      fmt.Sprintf("Failed to create output file: %v", err),
		})
		return
	}
	defer os.Remove(tempOutput.Name())
	tempOutput.Close()

	// Perform conversion
	if err := shared.ConvertFile(tempInput.Name(), tempOutput.Name(), fromFormat, toFormat, data.ColorSpace); err != nil {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusInternalServerError,
			StatusText:     "Conversion failed",
			ErrorText:      fmt.Sprintf("%v", err),
		})
		return
	}

	// Read converted file
	outputData, err := os.ReadFile(tempOutput.Name())
	if err != nil {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusInternalServerError,
			StatusText:     "Server error",
			ErrorText:      fmt.Sprintf("Failed to read converted file: %v", err),
		})
		return
	}

	// Send response (JSON will auto-encode []byte to base64)
	render.Render(w, r, &ConvertResponse{
		FileContent: outputData,
		Filename:    "converted" + toFormat,
		Format:      toFormat,
	})
}

//go:embed examples/example.json
var exampleJSON string

//go:embed examples/example.csv
var exampleCSV string

// handleExamples serves example palette files.
func handleExamples(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			StatusText:     "Invalid request",
			ErrorText:      "format parameter is required",
		})
		return
	}

	var content string
	var contentType string
	var filename string

	switch format {
	case "json":
		content = exampleJSON
		contentType = "application/json"
		filename = "example.json"
	case "csv":
		content = exampleCSV
		contentType = "text/csv"
		filename = "example.csv"
	default:
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			StatusText:     "Invalid request",
			ErrorText:      "Unsupported format. Available: json, csv",
		})
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}
