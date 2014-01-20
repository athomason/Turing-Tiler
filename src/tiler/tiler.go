package tiler

const (
	top = iota
	left
	bottom
	right
)

type Options struct {
	TileHeight, TileWidth        int
	MaxDepth                     int
	IgnoreDepthFailure           bool
	FontPath                     string
	Rotation                     int // 0 is best for portrait or web (top->down), 3 is best for landscape or monitors (left->right)
	FlipHorizontal, FlipVertical bool
	BoundarySymbol               string
}

type Tiler struct {
	Options

	// computed options
	tileHorizShift, tileVertShift,
	bondFudgeX, bondFudgeY,
	tileHorizMargin, tileVertMargin float64
	fontSize int
}

func (o *Options) NewTiler() *Tiler {
	return &Tiler{
		Options: *o,
	}
}
