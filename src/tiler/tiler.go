package tiler

import (
	"fmt"
	"log"
	. "math"
)

type Options struct {
	TileHeight, TileWidth        int
	MaxDepth                     int
	IgnoreDepthFailure           bool
	FontPath                     string
	Rotation                     int // 0 is best for portrait or web (top->down), 3 is best for landscape or monitors (left->right)
	FlipHorizontal, FlipVertical bool
	BoundarySymbol               string
	MachineFile                  string
	Inputs                       []string
}

type Tiler struct {
	Options

	// computed options
	fontSize               int
	bondFudgeX, bondFudgeY int
	tileHorizShift, tileVertShift,
	tileHorizMargin, tileVertMargin float64

	machine Machine
	tiles   []Tile // the tile "pool": things the self-assembler can draw from
}

func (o *Options) NewTiler() *Tiler {
	t := Tiler{
		Options:         *o,
		bondFudgeX:      int(Floor(float64(o.TileWidth) / 80)),
		bondFudgeY:      int(Floor(float64(o.TileHeight) / 80)),
		fontSize:        int(Sqrt(float64(o.TileHeight*o.TileWidth)) / 4),
		tileHorizShift:  float64(o.TileWidth) / 10,
		tileVertShift:   float64(o.TileHeight) / 10,
		tileHorizMargin: float64(o.TileWidth) / 20,
		tileVertMargin:  float64(o.TileHeight) / 20,
	}
	t.ParseMachine()
	t.GenerateTiles()
	return &t
}

type Direction int

const (
	Left Direction = iota
	Right
	Up
	Down
	Halt
)

type Tile struct {
	Name  string
	Sides Bonds
	Final bool
}

type Bonds map[Direction]Bond

type Bond struct {
	Strength int
	Label    string
}

func (t *Tiler) GenerateTiles() {
	// first set of tiles: transitions from old head states
	log.Println("Generating tileset 1/3...")
	for _, trans := range t.machine.Transitions {
		tile := Tile{
			Name: fmt.Sprintf("%s-%s", trans.OldState, trans.ReadSymbol),
		}
		switch trans.Move {
		case Left:
			tile.Sides = Bonds{
				Up:    Bond{1, trans.WriteSymbol},
				Left:  Bond{1, trans.NewState},
				Right: Bond{1, "R"},
			}
		case Right:
			tile.Sides = Bonds{
				Up:    Bond{1, trans.WriteSymbol},
				Left:  Bond{1, "L"},
				Right: Bond{1, "R"},
			}
		case Halt:
			var label string
			if trans.Output == "" {
				label = trans.WriteSymbol
			} else {
				label = fmt.Sprintf("%s [%s]", trans.WriteSymbol, trans.Output)
			}
			tile.Sides = Bonds{
				Up:    Bond{1, label},
				Left:  Bond{1, "L"},
				Right: Bond{1, "R"},
			}
			tile.Final = true
		}
		tile.Sides[Down] = Bond{2, fmt.Sprintf("%s %s", trans.OldState, trans.ReadSymbol)}
		t.tiles = append(t.tiles, tile)
	}

	// second set of tiles: exposes a double bond from the new head state
	log.Println("Generating tileset 2/3...")
	states := make(map[string]struct{})
	for _, trans := range t.machine.Transitions {
		states[trans.OldState] = struct{}{}
		states[trans.NewState] = struct{}{}
	}
	for state, _ := range states {
		for _, symbol := range t.machine.Symbols {
			// moving left
			left := Tile{
				Name: fmt.Sprintf("move-%s-%s-left", state, symbol),
				Sides: Bonds{
					Up:    Bond{2, fmt.Sprintf("%s %s", state, symbol)},
					Down:  Bond{1, symbol},
					Left:  Bond{1, "L"},
					Right: Bond{1, state},
				},
			}

			// moving right
			right := Tile{
				Name: fmt.Sprintf("move-%s-%s-right", state, symbol),
				Sides: Bonds{
					Up:    Bond{2, fmt.Sprintf("%s %s", state, symbol)},
					Down:  Bond{1, symbol},
					Left:  Bond{1, state},
					Right: Bond{1, "R"},
				},
			}

			t.tiles = append(t.tiles, left, right)
		}
	}

	// third set of tiles: replicates non-head state cells
	log.Println("Generating tileset 3/3...")
	for _, symbol := range t.machine.Symbols {
		// copying left of head
		left := Tile{
			Name: fmt.Sprintf("replicate-%s-left", symbol),
			Sides: Bonds{
				Up:    Bond{1, symbol},
				Down:  Bond{1, symbol},
				Left:  Bond{1, "L"},
				Right: Bond{1, "L"},
			},
		}

		// copying right of head
		right := Tile{
			Name: fmt.Sprintf("replicate-%s-right", symbol),
			Sides: Bonds{
				Up:    Bond{1, symbol},
				Down:  Bond{1, symbol},
				Left:  Bond{1, "R"},
				Right: Bond{1, "R"},
			},
		}
		t.tiles = append(t.tiles, left, right)
	}

	log.Println("Generating tile caches...")
	for _, tile := range t.tiles {
	}
	/*
		my ( %tile_cache_left, %tile_cache_right, %tile_cache_bottom );

		for my $tile ( @tiles ) {
			$tile_cache_bottom
				{ $tile->{ sides }{ +BOTTOM }{ label } } = $tile;
			$tile_cache_left
				{ $tile->{ sides }{ +LEFT   }{ label } }
				{ $tile->{ sides }{ +BOTTOM }{ label } } = $tile;
			$tile_cache_right
				{ $tile->{ sides }{ +RIGHT  }{ label } }
				{ $tile->{ sides }{ +BOTTOM }{ label } } = $tile;
		}
	*/
}
