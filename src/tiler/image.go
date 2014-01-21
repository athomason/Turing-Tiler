package tiler

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	. "math"
	"os"
	"regexp"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
)

type Cell struct {
	Symbol rune
	Head   bool
}

type drawer struct {
	// drawing constants computed from Options
	fontSize,
	bondFudgeX, bondFudgeY,
	tileHorizShift, tileVertShift,
	tileHorizMargin, tileVertMargin,
	fontWidth, fontHeight,
	fontX, fontY int

	font   *truetype.Font
	colors map[string]color.RGBA
}

func (t *Tiler) setupDrawer() {
	t.bondFudgeX = int(Floor(float64(t.TileWidth) / 80))
	t.bondFudgeY = int(Floor(float64(t.TileHeight) / 80))
	t.fontSize = int(Sqrt(float64(t.TileHeight*t.TileWidth)) / 4)
	t.tileHorizShift = int(float64(t.TileWidth) / 10)
	t.tileVertShift = int(float64(t.TileHeight) / 10)
	t.tileHorizMargin = int(float64(t.TileWidth) / 20)
	t.tileVertMargin = int(float64(t.TileHeight) / 20)

	bytes, err := ioutil.ReadFile(t.FontPath)
	if err != nil {
		log.Panicf("Couldn't read font: %s", err)
	}
	t.font, err = freetype.ParseFont(bytes)
	if err != nil {
		log.Panicf("Couldn't parse font: %s", err)
	}

	// figure out how big a representative character is for approximate layout purposes
	ex := '5'
	fupe := t.font.FUnitsPerEm()
	horiz := t.font.HMetric(fupe, t.font.Index(ex))
	vert := t.font.VMetric(fupe, t.font.Index(ex))
	t.fontX = int(horiz.LeftSideBearing)
	t.fontWidth = int(horiz.AdvanceWidth)
	t.fontY = int(vert.TopSideBearing)
	t.fontHeight = int(vert.AdvanceHeight)
	log.Printf("%#v", t)
}

func (t *Tiler) newTypeContext(im *image.RGBA, color color.RGBA) *freetype.Context {
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(t.font)
	c.SetFontSize(t.FontSize)
	c.SetClip(im.Bounds())
	c.SetDst(im)
	c.SetSrc(image.NewUniform(color))
	return c
}

func (t *Tiler) Assemble() {
	for _, input := range t.Inputs {
		t.AssembleOne(input)
	}
}

type Assembly [][]*Tile

func (t *Tiler) AssembleOne(input string) {
	log.Printf("Processing input %q...", input)

	// check that the input string has only legal symbols
	symbolRx := regexp.MustCompile(fmt.Sprintf("[^%s]", string(t.Symbols)))
	if symbolRx.MatchString(input) {
		log.Printf("  Warning: invalid symbol encountered in input string %q", input)
		return
	}

	// annotate initial input with head semantics before generating starter tiles
	cells := make([]Cell, len(input))
	for i, r := range input {
		cells[i] = Cell{r, i == t.InitialLocation}
	}

	log.Printf("Generating starter tiles...")

	// seed first row of assembly with starter tiles from input-generated cells
	assembly := Assembly{make([]*Tile, len(cells)+2)}

	// wrap with boundary symbols at beginning and end
	assembly[0][0] = t.cellToTile(&Cell{t.BoundarySymbol, false}, true, false)
	for i, cell := range cells {
		assembly[0][i+1] = t.cellToTile(&cell, false, false)
	}
	assembly[0][len(assembly)-1] = t.cellToTile(&Cell{t.BoundarySymbol, false}, false, true)

	log.Printf("Assembling transition tiles...")

	// assemble until we hit a halting state or the depth limit is reached
	for t.MaxDepth == 0 || len(assembly) < t.MaxDepth {
		t.addTile(assembly)
	}

	if len(assembly) > t.MaxDepth {
		assembly = assembly[:t.MaxDepth]
		log.Printf("  Warning: assembly hit maximum depth (%d), increase with -max-depth", t.MaxDepth)
	}

	// remove trailing blank line
	assembly = assembly[:len(assembly)-2]

	sizeX, sizeY := len(assembly[0]), len(assembly)

	log.Printf("Transforming matrix...")

	// rotation occurs after assembly so that the assembler routine may concern
	// itself only with a single logical layout (bottom-up). here we rotate the
	// assembly matrix to the desired final orientation; tile orientation is
	// also rotated, but in the drawing routines
	sizeX, sizeY, assembly = t.computeRotated(sizeX, sizeY, assembly)

	/*
		// create the master canvas containing the record of the entire computation
		my $target = GD::Image->new( t.TileWidth * $size_x - $size_x + 1, t.TileHeight * $size_y - $size_y + 1 );
		$target->saveAlpha( 1 );

		print STDERR "  Generating canvas...\n";

		// copy component tiles to the master canvas
		for my $i ( 0 .. @assembly - 1 ) {
			for my $j ( 0 .. @{ $assembly[ $i ] } - 1 ) {
				my $src = $assembly[ $i ][ $j ]{ image };
				next unless defined $src; // silently ignored missing tiles
				$target->copy( $src,
					t.TileWidth * $j - $j,
					t.TileHeight * ( @assembly - $i - 1 ) - ( @assembly - $i - 1 ),
					0, 0, t.TileWidth, t.TileHeight
				);
			}
		}

		// save the output

		my $output_file = "$name-$input_string.png";
		print STDERR "  Saving image $output_file...\n";
		open OUTPUT, ">", $output_file;
		binmode OUTPUT;
		print OUTPUT $target->png;
		close OUTPUT;

	*/
	log.Printf("Done!")
}

// starting with the current assembly, try to add any tile drawn from the pool
// that fits in an empty spot adjacent to an existing tile. tiles may only be
// added if at least two bonds are made in so doing (i.e. two single bonds or
// one double bond).
func (t *Tiler) addTile(assembly Assembly) bool {
	// search over the entire current assembly
	startY := len(assembly) - 2
	if startY < 0 {
		startY = 0
	}
	for y := startY; y < len(assembly)-1; y++ {
		for x := 0; x < len(assembly[y]); x++ {
			// TODO: continue porting
			/*
			               my $tile = $assembly->[ $y ][ $x ]; // this tile
			               next unless defined $tile;

			   			ltile := assembly[y][x-1] // left neighbor
			   			rtile := assembly[y][x+1] // right neighbor
			   			ttile := assembly[y+1][ $x ] // top neighbor
			   			lltile := assembly[y-1][x-1] // left and down
			   			lrtile := assembly[y-1][x+1] // right and down

			               // try to add tile left
			               if (
			                   !defined $ltile && // can't step on an existing tile
			                   defined $lltile && // can only add left if left-down is there
			                   (
			                       $tile->{ sides }{ +LEFT }{ bond_strength } +
			                       $lltile->{ sides }{ +TOP }{ bond_strength }
			                   ) >= 2 // would-be bond sum is good enough
			               ) {
			                   // space looks good, search for a matching tile in the pool
			                   my $stile = $tile_cache_right
			                       { $tile->{ sides }{ +LEFT }{ label } }
			                       { $lltile->{ sides }{ +TOP }{ label } };
			                   if ( defined $stile ) {
			                       // found a match, fill it in
			                       $assembly->[ $y ][ $x - 1 ] = $stile;
			                       return 1;
			                   }
			               }

			               // try to add tile right
			               if (
			                   !defined $rtile && // can't step on an existing tile
			                   defined $lrtile && // can only add right if right-down is there
			                   (
			                       $tile->{ sides }{ +RIGHT }{ bond_strength } +
			                       $lrtile->{ sides }{ +TOP }{ bond_strength } >= 2
			                   ) // would-be bond sum is good enough
			               ) {
			                   // space looks good, search for a matching tile in the pool
			                   my $stile = $tile_cache_left
			                       { $tile->{ sides }{ +RIGHT }{ label } }
			                       { $lrtile->{ sides }{ +TOP }{ label } };
			                   if ( defined $stile ) {
			                       // found a match, fill it in
			                       $assembly->[ $y ][ $x + 1 ] = $stile;
			                       return 1;
			                   }
			               }

			               // try to add tile up; adding vertically can ONLY happen by double-bond
			               if (
			                   !defined $ttile &&
			                   $tile->{ sides }{ +TOP }{ bond_strength } >= 2
			               ) {
			                   my $stile = $tile_cache_bottom{ $tile->{ sides }{ +TOP }{ label } };
			                   if ( defined $stile ) {
			                       // found a match, fill it in
			                       $assembly->[ $y + 1 ][ $x ] = $stile;
			                       return 1;
			                   }
			               }
			*/
		}
	}

	// couldn't find any tile to add anywhere, so halt
	return false
}

func bondStrength(strong bool) int {
	if strong {
		return 2
	}
	return 1
}

// build an initial tile from a cell definition
func (t *Tiler) cellToTile(cell *Cell, left, right bool) *Tile {
	var upLabel string
	if cell.Head {
		upLabel = fmt.Sprintf("%s %s", t.InitialState)
	} else {
		upLabel = string(cell.Symbol)
	}
	tile := Tile{
		Name: "seed",
		Sides: Bonds{
			Up:    Bond{bondStrength(cell.Head), upLabel},
			Down:  Bond{1, ""},
			Left:  Bond{bondStrength(left), ""},
			Right: Bond{bondStrength(right), ""},
		},
	}
	tile.Image = t.generateImage(&tile)
	return &tile
}

func (t *Tiler) generateImage(tile *Tile) image.Image {
	r := image.Rect(0, 0, t.TileWidth, t.TileHeight)
	im := image.NewRGBA(r)
	bgColor := t.getLabelColor(fmt.Sprintf("background%v", tile.Final), true)
	draw.Draw(im, im.Bounds(), &image.Uniform{bgColor}, image.ZP, draw.Src)

	for _, side := range []Direction{Up, Down, Left, Right} {
		strength := tile.Sides[side].Strength
		label := tile.Sides[side].Label

		orientedLabel := fmt.Sprintf("%d%s%d", side%2, label, strength)
		color := t.getLabelColor(orientedLabel, false)

		rotSide := t.rotatedDirection(side)
		t.drawBond(im, rotSide, strength, color)
		t.drawString(im, rotSide, strength, color, label)
	}
	w, err := os.Create(fmt.Sprintf("%s.png", tile.Name))
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()
	png.Encode(w, im)
	return im
}

// for the given label, return a visually well-distributed color that is always
// the same but uncorrelated to the label's contents
func (t *Tiler) getLabelColor(label string, bright bool) color.RGBA {
	if color, exists := t.colors[label]; exists {
		return color
	}

	// use a hash function to get some deterministic but random-looking data
	// based on the label. here we get 32 bytes from MD5 and unpack the first
	// 24 as uint64's.
	var i1, i2, i3 uint64
	hash := md5.Sum([]byte(label + t.ColorTweak))
	buf := bytes.NewReader(hash[:])
	binary.Read(buf, binary.LittleEndian, &i1)
	binary.Read(buf, binary.LittleEndian, &i2)
	binary.Read(buf, binary.LittleEndian, &i3)

	r1 := float64(i1) / float64(MaxUint64)
	r2 := float64(i2) / float64(MaxUint64)
	r3 := float64(i3) / float64(MaxUint64)

	h, s, v := r1, 0.0, 0.0
	if bright {
		v = 0.9 + 0.1*r3
		s = 0.0 + 0.02*r2
	} else {
		v = 0.1 + 0.4*r3
		s = 0.5 + 0.5*r2
	}

	cp := Floor(h * 6)
	cs := h - cp
	ca := (1 - s) * v
	cb := (1 - (s * cs)) * v
	cc := (1 - (s * (1 - cs))) * v

	var r, g, b float64
	switch int(cp) {
	case 0:
		r, g, b = v, cc, ca
	case 1:
		r, g, b = cb, v, ca
	case 2:
		r, g, b = ca, v, cc
	case 3:
		r, g, b = ca, cb, v
	case 4:
		r, g, b = cc, ca, v
	case 5:
		r, g, b = v, ca, cb
	}

	r = 255 * Max(0, Min(1, r))
	g = 255 * Max(0, Min(1, g))
	b = 255 * Max(0, Min(1, b))

	c := color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b)}
	if t.colors == nil {
		t.colors = make(map[string]color.RGBA)
	}
	t.colors[label] = c
	return c
}

func (t *Tiler) drawBond(im draw.Image, side Direction, strength int, color color.RGBA) {
	for i := 0; i < strength; i++ {
		var r image.Rectangle
		switch side {
		case Up:
			r = image.Rect(0, i*t.tileVertShift-t.bondFudgeY,
				t.TileWidth-1, i*t.tileVertShift+t.bondFudgeY)
		case Down:
			r = image.Rect(0, t.TileHeight-1-i*t.tileVertShift-t.bondFudgeY,
				t.TileWidth-1, t.TileHeight-1-i*t.tileVertShift+t.bondFudgeY)
		case Left:
			r = image.Rect(i*t.tileHorizShift-t.bondFudgeX, 0,
				i*t.tileHorizShift+t.bondFudgeX, t.TileHeight-1)
		case Right:
			r = image.Rect(t.TileWidth-1-i*t.tileHorizShift-t.bondFudgeX, 0,
				t.TileWidth-1-i*t.tileHorizShift+t.bondFudgeX, t.TileHeight-1)
		}
		draw.Draw(im, r, &image.Uniform{color}, image.ZP, draw.Src)
	}
}

func (t *Tiler) rotatedDirection(side Direction) Direction {
	side = Direction((int(side) + t.Rotation) % 4)
	if t.FlipHorizontal {
		if side == Left {
			side = Right
		} else if side == Right {
			side = Left
		}
	}
	if t.FlipVertical {
		if side == Up {
			side = Down
		} else if side == Down {
			side = Up
		}
	}
	return side
}

func (t *Tiler) drawString(im *image.RGBA, side Direction, strength int, color color.RGBA, str string) {
	bondShift := strength - 1
	var x, y int
	switch side {
	case Up:
		y = t.tileVertMargin + bondShift*t.tileVertShift
		x = int((t.TileWidth - len(str)*t.fontWidth) / 2)
	case Down:
		y = t.TileHeight - t.fontHeight - t.tileVertMargin - bondShift*t.tileVertShift
		x = int((t.TileWidth - len(str)*t.fontWidth) / 2)
	case Left:
		y = int((t.TileHeight - t.fontHeight) / 2)
		x = t.tileHorizMargin + bondShift*t.tileHorizShift
	case Right:
		y = int((t.TileHeight - t.fontHeight) / 2)
		x = t.TileWidth - t.tileHorizMargin - len(str)*t.fontWidth - bondShift*t.tileHorizShift
	}

	ctx := t.newTypeContext(im, color)
	pt := freetype.Pt(x, y)
	ctx.DrawString(str, pt)
}

// rotate the @assembly matrix
func (t *Tiler) computeRotated(sizeX, sizeY int, assembly Assembly) (int, int, Assembly) {
	return sizeX, sizeY, assembly
	/*
	   my $old_assembly = shift;
	   my $size_x = shift;
	   my $size_y = shift;
	   my $new_assembly;

	   for my $j ( 0 .. $size_y - 1 ) {
	       for my $i ( 0 .. $size_x - 1 ) {
	           if ( $rotation == 0 ) {
	               // no rotation, but possibly flips
	               my $ti = $i;
	               my $tj = $j;
	               $ti = $size_x - 1 - $ti if     $flip_horiz;
	               $tj = $size_y - 1 - $tj if     $flip_vert;
	               $new_assembly->[ $tj ][ $ti ] = $old_assembly->[ $j ][ $i ];
	           }
	           elsif ( $rotation == 1 ) {
	               // CCW 90
	               my $ti = $i;
	               my $tj = $j;
	               $ti = $size_x - 1 - $ti if     $flip_horiz;
	               $tj = $size_y - 1 - $tj unless $flip_vert;
	               $new_assembly->[ $ti ][ $tj ] = $old_assembly->[ $j ][ $i ];
	           }
	           elsif ( $rotation == 2 ) {
	               // 180
	               my $ti = $i;
	               my $tj = $j;
	               $ti = $size_x - 1 - $ti unless $flip_horiz;
	               $tj = $size_y - 1 - $tj unless $flip_vert;
	               $new_assembly->[ $tj ][ $ti ] = $old_assembly->[ $j ][ $i ];
	           }
	           elsif ( $rotation == 3 ) {
	               // CW 90
	               my $ti = $i;
	               my $tj = $j;
	               $ti = $size_x - 1 - $ti unless $flip_horiz;
	               $tj = $size_y - 1 - $tj if     $flip_vert;
	               $new_assembly->[ $ti ][ $tj ] = $old_assembly->[ $j ][ $i ];
	           }
	       }
	   }

	   if ( $rotation % 2 == 1 ) {
	       ( $size_x, $size_y ) = ( $size_y, $size_x );
	   }
	   return $size_x, $size_y, @$new_assembly;
	*/
}
