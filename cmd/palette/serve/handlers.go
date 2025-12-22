package serve

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/ajg/form"
	"github.com/go-chi/render"
	"github.com/kennyp/palette/cmd/palette/shared"
)

func init() {
	// Extend render.Decode to support multipart/form-data
	originalDecode := render.Decode
	render.Decode = func(r *http.Request, v any) error {
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "multipart/form-data") {
			// Parse multipart form
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				return err
			}
			// Decode form values into the struct using ajg/form
			decoder := form.NewDecoder(nil)
			return decoder.DecodeValues(v, r.Form)
		}
		// Fall back to original decoder for other content types
		return originalDecode(r, v)
	}
}

//go:embed templates/index.html
var indexHTML string

//go:embed templates/favicon.svg
var faviconSVG []byte

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

// handleFavicon serves the favicon.
func handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.WriteHeader(http.StatusOK)
	w.Write(faviconSVG)
}

// ConvertFormRequest represents a multipart form conversion request.
type ConvertFormRequest struct {
	To         string `form:"to"`
	ColorSpace string `form:"colorspace"`
	BookID     string `form:"book_id"` // Optional: custom BookID for ACB export (4000-65535)
}

// Bind implements render.Binder for multipart form requests.
func (c *ConvertFormRequest) Bind(r *http.Request) error {
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

// handleConvert handles file upload and conversion using chi render.
func handleConvert(w http.ResponseWriter, r *http.Request) {
	// Bind and validate form data using render.Bind (now supports multipart)
	data := &ConvertFormRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			StatusText:     "Invalid request",
			ErrorText:      err.Error(),
		})
		return
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
	if err := shared.ConvertFile(tempInput.Name(), tempOutput.Name(), fromFormat, data.To, data.ColorSpace, data.BookID); err != nil {
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

	// Perform conversion (JSON API doesn't support BookID for now - use empty string)
	if err := shared.ConvertFile(tempInput.Name(), tempOutput.Name(), fromFormat, toFormat, data.ColorSpace, ""); err != nil {
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

//go:embed examples/example-cmyk.json
var exampleCMYKJSON string

//go:embed examples/example-cmyk.csv
var exampleCMYKCSV string

//go:embed examples/example-hsb.csv
var exampleHSBCSV string

//go:embed examples/example-lab.json
var exampleLABJSON string

//go:embed examples/example-lab.csv
var exampleLABCSV string

// handleExamples serves example palette files.
func handleExamples(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	colorspace := r.URL.Query().Get("colorspace")

	if format == "" {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			StatusText:     "Invalid request",
			ErrorText:      "format parameter is required",
		})
		return
	}

	// Default to RGB if no colorspace specified
	if colorspace == "" {
		colorspace = "rgb"
	}

	var content string
	var contentType string
	var filename string

	// Map format + colorspace to example file
	key := format + "-" + colorspace
	switch key {
	case "json-rgb":
		content = exampleJSON
		contentType = "application/json"
		filename = "example-rgb.json"
	case "json-cmyk":
		content = exampleCMYKJSON
		contentType = "application/json"
		filename = "example-cmyk.json"
	case "json-lab":
		content = exampleLABJSON
		contentType = "application/json"
		filename = "example-lab.json"
	case "csv-rgb":
		content = exampleCSV
		contentType = "text/csv"
		filename = "example-rgb.csv"
	case "csv-cmyk":
		content = exampleCMYKCSV
		contentType = "text/csv"
		filename = "example-cmyk.csv"
	case "csv-hsb":
		content = exampleHSBCSV
		contentType = "text/csv"
		filename = "example-hsb.csv"
	case "csv-lab":
		content = exampleLABCSV
		contentType = "text/csv"
		filename = "example-lab.csv"
	default:
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			StatusText:     "Invalid request",
			ErrorText:      fmt.Sprintf("Unsupported combination: format=%s, colorspace=%s", format, colorspace),
		})
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}
