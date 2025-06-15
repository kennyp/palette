package palette

import (
	paletteio "github.com/kennyp/palette/io"
	"github.com/kennyp/palette/io/colorbook"
	"github.com/kennyp/palette/io/colorswatch"
	"github.com/kennyp/palette/io/csv"
	"github.com/kennyp/palette/io/json"
)

func init() {
	// Register all format importers and exporters with the default registry

	// Adobe Color Book (.acb)
	paletteio.DefaultRegistry.RegisterImporter(colorbook.NewImporter())
	paletteio.DefaultRegistry.RegisterExporter(colorbook.NewExporter())

	// Adobe Color Swatch (.aco)
	paletteio.DefaultRegistry.RegisterImporter(colorswatch.NewImporter())
	paletteio.DefaultRegistry.RegisterExporter(colorswatch.NewExporter())

	// CSV
	paletteio.DefaultRegistry.RegisterImporter(csv.NewImporter())
	paletteio.DefaultRegistry.RegisterExporter(csv.NewExporter())

	// JSON
	paletteio.DefaultRegistry.RegisterImporter(json.NewImporter())
	paletteio.DefaultRegistry.RegisterExporter(json.NewExporter())
}
