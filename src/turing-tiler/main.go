package main

import (
	"flag"
	"log"
	"unicode/utf8"

	"tiler"
)

func main() {
	var options tiler.Options
	var boundarySymbol string
	// 4/3 is a reasonable aspect ratio for single-character states and symbols
	flag.IntVar(&options.TileWidth, "tile-width", 32, "tile width in pixels")
	flag.IntVar(&options.TileHeight, "tile-height", 24, "tile height in pixels")
	flag.IntVar(&options.MaxDepth, "max-depth", 100, "maximum number of transitions")
	flag.BoolVar(&options.IgnoreDepthFailure, "ignore-depth-failure", false, "proceed when MaxDepth is exceeded")
	flag.StringVar(&options.FontPath, "font-path", "/usr/share/fonts/truetype/ttf-bitstream-vera/Vera.ttf", "path to a truetype font")
	flag.Float64Var(&options.FontSize, "font-size", 12, "font size in points")
	flag.IntVar(&options.Rotation, "rotation", 0, "rotation from 0-3")
	flag.BoolVar(&options.FlipHorizontal, "flip-horizontal", false, "flip the output horizontally")
	flag.BoolVar(&options.FlipVertical, "flip-vertical", false, "flip the output vertically")
	flag.StringVar(&options.ColorTweak, "color-tweak", "", "string which consistently but unpredictably changes color selection")
	flag.StringVar(&boundarySymbol, "boundary-symbol", "*", "boundary symbol")
	flag.Parse()

	if flag.NArg() < 2 {
		log.Fatalf("usage: %s [options] <machine_spec> <input_string> [<input_string>] [...]", "tiler")
	}

	options.MachineFile = flag.Arg(0)
	options.Inputs = flag.Args()[1:]
	options.BoundarySymbol, _ = utf8.DecodeRune([]byte(boundarySymbol))

	log.Printf("Processing %s, %v", options.MachineFile, options.Inputs)
	tiler := options.NewTiler()
	tiler.Assemble()
}
