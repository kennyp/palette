# Palette CLI

A command-line tool and web server for converting color palette files between different formats.

## Installation

```bash
go install github.com/kennyp/palette/cmd/palette@latest
```

Or build from source:

```bash
go build -o palette ./cmd/palette
```

## Usage

The `palette` command provides two subcommands:

### Convert Command

Convert palette files between different formats via the command line.

```bash
# Basic conversion (auto-detect formats from extensions)
palette convert -i colors.aco -o colors.json

# Specify source and target formats explicitly
palette convert -i input.dat -o output.dat --from .acb --to .csv

# Convert with color space transformation
palette convert -i palette.acb -o palette.csv --colorspace RGB
palette convert -i colors.json -o colors.aco --colorspace CMYK
```

**Options:**
- `-i, --input` - Input file path (required)
- `-o, --output` - Output file path (required)
- `--from` - Source format (auto-detected if omitted): `.acb`, `.aco`, `.csv`, `.json`
- `--to` - Target format (inferred from output extension if omitted)
- `--colorspace` - Convert all colors to specified color space: `RGB`, `CMYK`, `LAB`, `HSB`

### Serve Command

Start a web server with a user-friendly interface for palette conversion.

```bash
# Start server on default port (8080)
palette serve

# Specify custom port
palette serve --port 3000

# Bind to specific host
palette serve --host 0.0.0.0 --port 8080
```

**Options:**
- `-p, --port` - Port to listen on (default: 8080)
- `--host` - Host address to bind to (default: localhost)

**Web Interface Features:**
- Drag-and-drop file upload
- Format auto-detection
- Target format selection
- Optional color space conversion
- Instant file download
- Support for all palette formats

**API Endpoints:**
- `GET /` - Web UI
- `POST /api/convert` - File conversion endpoint
- `GET /api/formats` - List supported formats
- `GET /health` - Health check

## Supported Formats

| Format | Extension | Description |
|--------|-----------|-------------|
| Adobe Color Book | `.acb` | Adobe's proprietary color book format |
| Adobe Color Swatch | `.aco` | Adobe color swatch files (v1 & v2) |
| CSV | `.csv` | Comma-separated values with color data |
| JSON | `.json` | JSON format with flexible schema |

## Examples

### CLI Examples

```bash
# Convert Adobe Color Book to JSON
palette convert -i PANTONE.acb -o pantone.json

# Convert Adobe Swatch to CSV with RGB colors
palette convert -i colors.aco -o colors.csv --colorspace RGB

# Convert JSON to Adobe Color Book
palette convert -i my-palette.json -o output.acb
```

### API Examples

```bash
# Convert via API
curl -F "file=@colors.aco" \
     -F "to=.json" \
     http://localhost:8080/api/convert \
     -o converted.json

# Convert with color space transformation
curl -F "file=@palette.acb" \
     -F "to=.csv" \
     -F "colorspace=RGB" \
     http://localhost:8080/api/convert \
     -o output.csv

# Get supported formats
curl http://localhost:8080/api/formats
```

## Environment Variables

The serve command supports environment variables:

- `HOST` - Host address (overridden by `--host` flag)
- `PORT` - Port number (overridden by `--port` flag)

Example:
```bash
PORT=3000 palette serve
```

## Build Information

To include version information in the binary:

```bash
go build -ldflags "\
  -X main.version=1.0.0 \
  -X main.commit=$(git rev-parse --short HEAD) \
  -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o palette ./cmd/palette
```

## Architecture

```
cmd/palette/
├── main.go              # Entry point with urfave/cli v3
├── convert/
│   └── convert.go       # Convert subcommand
├── serve/
│   ├── server.go        # HTTP server with chi v5 router
│   ├── handlers.go      # Request handlers
│   └── templates/
│       └── index.html   # Web UI with HTMX 2.0 & Alpine.js 3.14
└── shared/
    └── converter.go     # Shared conversion logic
```

## Dependencies

- **urfave/cli v3** - Command-line interface framework
- **go-chi/chi v5** - HTTP router
- **HTMX 2.0.7** - Frontend interactions (via CDN)
- **Alpine.js 3.14.9** - Frontend reactivity (via CDN)

## License

Same as the parent Palette library.
