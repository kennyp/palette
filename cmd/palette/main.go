package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kennyp/palette/cmd/palette/convert"
	"github.com/kennyp/palette/cmd/palette/serve"
	"github.com/urfave/cli/v3"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	app := &cli.Command{
		Name:    "palette",
		Usage:   "Color palette conversion tool",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		Description: `A command-line tool and web server for converting color palette files
between different formats.

Supported formats:
  .acb - Adobe Color Book
  .aco - Adobe Color Swatch
  .csv - Comma-Separated Values
  .json - JSON

For more information about a specific command, use:
  palette <command> --help`,
		Commands: []*cli.Command{
			convert.Command(),
			serve.Command(),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Enable verbose output",
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
