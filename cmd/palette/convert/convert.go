package convert

import (
	"context"
	"fmt"
	"os"

	"github.com/kennyp/palette/cmd/palette/shared"
	"github.com/urfave/cli/v3"
)

// Command returns the convert subcommand.
func Command() *cli.Command {
	return &cli.Command{
		Name:  "convert",
		Usage: "Convert palette files between different formats",
		Description: `Convert color palette files between supported formats:
   .acb - Adobe Color Book
   .aco - Adobe Color Swatch
   .csv - Comma-Separated Values
   .json - JSON

Examples:
   palette convert -i colors.aco -o colors.json
   palette convert -i palette.acb -o palette.csv --colorspace RGB
   palette convert --input data.json --output output.aco`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Usage:    "Input file path (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Usage:    "Output file path (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "from",
				Usage: "Source format (auto-detect if omitted): .acb, .aco, .csv, .json",
			},
			&cli.StringFlag{
				Name:  "to",
				Usage: "Target format (infer from output extension if omitted): .acb, .aco, .csv, .json",
			},
			&cli.StringFlag{
				Name:  "colorspace",
				Usage: "Convert all colors to specified color space: RGB, CMYK, LAB, HSB",
			},
		},
		Action: run,
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	inputPath := cmd.String("input")
	outputPath := cmd.String("output")
	fromFormat := cmd.String("from")
	toFormat := cmd.String("to")
	colorSpace := cmd.String("colorspace")

	// Validate color space if provided
	if colorSpace != "" {
		if err := shared.ValidateColorSpace(colorSpace); err != nil {
			return cli.Exit(fmt.Sprintf("Error: %v", err), 1)
		}
	}

	// Check if input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return cli.Exit(fmt.Sprintf("Error: input file does not exist: %s", inputPath), 1)
	}

	// Perform conversion
	if err := shared.ConvertFile(inputPath, outputPath, fromFormat, toFormat, colorSpace); err != nil {
		return cli.Exit(fmt.Sprintf("Error: %v", err), 1)
	}

	// Success message
	fromFmt := fromFormat
	if fromFmt == "" {
		fromFmt = shared.DetectFormat(inputPath)
	}
	toFmt := toFormat
	if toFmt == "" {
		toFmt = shared.DetectFormat(outputPath)
	}

	fmt.Fprintf(cmd.Root().Writer, "Successfully converted %s to %s\n", fromFmt, toFmt)
	if colorSpace != "" {
		fmt.Fprintf(cmd.Root().Writer, "Colors converted to %s color space\n", colorSpace)
	}
	fmt.Fprintf(cmd.Root().Writer, "Output written to: %s\n", outputPath)

	return nil
}
