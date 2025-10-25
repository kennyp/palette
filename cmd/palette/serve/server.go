package serve

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "embed"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/google/gops/agent"
	"github.com/urfave/cli/v3"
	"golang.ngrok.com/ngrok/v2"
)

// Command returns the serve subcommand.
func Command() *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "Start the web server for palette conversion",
		Description: `Start a web server that provides a user interface for converting
color palette files between different formats.

The server provides:
  - Web UI with drag-and-drop file upload
  - Format conversion via browser
  - Support for all palette formats (.acb, .aco, .csv, .json)
  - Optional color space conversion

Examples:
   palette serve
   palette serve --port 3000
   palette serve --host 0.0.0.0 --port 8080
   palette serve --ngrok-url https://myapp.ngrok.io
   palette serve --ngrok-url https://myapp.ngrok.io --ngrok-token <token>`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "host",
				Usage:   "Host address to bind to",
				Value:   "localhost",
				Sources: cli.EnvVars("HOST"),
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "Port to listen on",
				Value:   8080,
				Sources: cli.EnvVars("PORT"),
			},
			&cli.StringFlag{
				Name:    "ngrok-url",
				Usage:   "ngrok URL to use (enables ngrok mode)",
				Sources: cli.EnvVars("NGROK_URL"),
			},
			&cli.StringFlag{
				Name:    "ngrok-token",
				Usage:   "ngrok auth token (optional, falls back to local ngrok config)",
				Sources: cli.EnvVars("NGROK_TOKEN"),
			},
		},
		Action: run,
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	// Initialize gops agent for debugging
	if err := agent.Listen(agent.Options{}); err != nil {
		slog.Warn("failed to start gops agent", "error", err)
	}
	defer agent.Close()

	host := cmd.String("host")
	port := cmd.Int("port")
	addr := fmt.Sprintf("%s:%d", host, port)
	ngrokURL := cmd.String("ngrok-url")
	ngrokToken := cmd.String("ngrok-token")

	// Create chi router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	logSchema := httplog.SchemaECS.Concise(host == "localhost" && ngrokURL == "")
	r.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{
		Level:  slog.LevelInfo,
		Schema: logSchema,
		LogRequestHeaders: []string{
			"Ngrok-Auth-User-Email",
			"Ngrok-Auth-User-Name",
		},
	}))

	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Add CORS middleware for development
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Define routes
	r.Get("/", handleIndex)
	r.Get("/health", handleHealth)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Post("/convert", handleConvert)        // Multipart form upload
		r.Post("/v1/convert", handleConvertJSON) // JSON API
		r.Get("/formats", handleFormats)
		r.Get("/examples", handleExamples)
	})

	// Channel to listen for errors coming from the server
	serverErrors := make(chan error, 1)

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Determine if we're using ngrok or standard server
	if ngrokURL != "" {
		// Use ngrok
		return runWithNgrok(ctx, cmd, r, ngrokURL, ngrokToken, serverErrors, shutdown)
	}

	// Standard local server
	return runLocalServer(ctx, cmd, r, addr, serverErrors, shutdown)
}

func runLocalServer(_ context.Context, cmd *cli.Command, handler http.Handler, addr string, serverErrors chan error, shutdown chan os.Signal) error {
	// Create HTTP server
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Start the server
	go func() {
		fmt.Fprintf(cmd.Root().Writer, "Server starting on http://%s\n", addr)
		fmt.Fprintf(cmd.Root().Writer, "Press Ctrl+C to stop\n\n")
		serverErrors <- server.ListenAndServe()
	}()

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		return cli.Exit(fmt.Sprintf("Server error: %v", err), 1)

	case sig := <-shutdown:
		fmt.Fprintf(cmd.Root().Writer, "\n%v signal received, shutting down...\n", sig)

		// Create a context with timeout for shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(shutdownCtx); err != nil {
			server.Close()
			return cli.Exit(fmt.Sprintf("Failed to gracefully shutdown: %v", err), 1)
		}

		fmt.Fprintf(cmd.Root().Writer, "Server stopped gracefully\n")
	}

	return nil
}

//go:embed policy.yaml
var trafficPolicy string

func runWithNgrok(ctx context.Context, cmd *cli.Command, handler http.Handler, ngrokURL string, ngrokToken string, serverErrors chan error, shutdown chan os.Signal) error {
	if !strings.HasPrefix(ngrokURL, "http") {
		ngrokURL = "https://" + ngrokURL
	}

	// Parse the ngrok URL
	_, err := url.Parse(ngrokURL)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Invalid ngrok URL: %v", err), 1)
	}

	// Require auth token when using ngrok
	if ngrokToken == "" {
		return cli.Exit("Error: --ngrok-token is required when using --ngrok-url.\n\nGet your auth token from: https://dashboard.ngrok.com/get-started/your-authtoken\nOr find it in your local config with: ngrok config check", 1)
	}

	// Create ngrok agent with auth token
	agent, err := ngrok.NewAgent(ngrok.WithAuthtoken(ngrokToken))
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create ngrok agent: %v", err), 1)
	}
	if err := agent.Connect(ctx); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to connect ngrok agent: %v", err), 1)
	}

	// Create listener with the specified URL
	listener, err := agent.Listen(ctx,
		ngrok.WithURL(ngrokURL),
		ngrok.WithTrafficPolicy(trafficPolicy),
		ngrok.WithDescription("palette web server"),
	)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create ngrok tunnel: %v", err), 1)
	}
	defer listener.Close()

	// Create HTTP server
	server := &http.Server{
		Handler: handler,
	}

	// Start the server
	go func() {
		fmt.Fprintf(cmd.Root().Writer, "Server starting on %s\n", listener.URL())
		fmt.Fprintf(cmd.Root().Writer, "Press Ctrl+C to stop\n\n")
		serverErrors <- server.Serve(listener)
	}()

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		return cli.Exit(fmt.Sprintf("Server error: %v", err), 1)

	case sig := <-shutdown:
		fmt.Fprintf(cmd.Root().Writer, "\n%v signal received, shutting down...\n", sig)

		// Create a context with timeout for shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(shutdownCtx); err != nil {
			server.Close()
			return cli.Exit(fmt.Sprintf("Failed to gracefully shutdown: %v", err), 1)
		}

		fmt.Fprintf(cmd.Root().Writer, "Server stopped gracefully\n")
	}

	return nil
}
