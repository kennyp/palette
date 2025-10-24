package serve

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/urfave/cli/v3"
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
   palette serve --host 0.0.0.0 --port 8080`,
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
		},
		Action: run,
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	host := cmd.String("host")
	port := cmd.Int("port")
	addr := fmt.Sprintf("%s:%d", host, port)

	// Create chi router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
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
	r.Post("/api/convert", handleConvert)
	r.Get("/api/formats", handleFormats)
	r.Get("/health", handleHealth)

	// Create HTTP server
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Channel to listen for errors coming from the server
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		fmt.Fprintf(cmd.Root().Writer, "Server starting on http://%s\n", addr)
		fmt.Fprintf(cmd.Root().Writer, "Press Ctrl+C to stop\n\n")
		serverErrors <- server.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		return cli.Exit(fmt.Sprintf("Server error: %v", err), 1)

	case sig := <-shutdown:
		fmt.Fprintf(cmd.Root().Writer, "\n%v signal received, shutting down...\n", sig)

		// Create a context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			server.Close()
			return cli.Exit(fmt.Sprintf("Failed to gracefully shutdown: %v", err), 1)
		}

		fmt.Fprintf(cmd.Root().Writer, "Server stopped gracefully\n")
	}

	return nil
}
