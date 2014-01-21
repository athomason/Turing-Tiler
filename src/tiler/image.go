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
	"math"
    "os"
	"regexp"

	"code.google.com/p/freetype-go/freetype"
)

type Cell struct {
	Symbol rune
	Head   bool
}

func (t *Tiler) DrawImages() {
	bytes, err := ioutil.ReadFile(t.FontPath)
	if err != nil {
		log.Panicf("Couldn't read font: %s", err)
	}

	font, err := freetype.ParseFont(bytes)
	if err != nil {
		log.Panicf("Couldn't parse font: %s", err)
	}

	_ = font

	log.Println("Generating tile caches...")
	type String2Tuple struct {
		first, second string
	}
	tileCacheBottom := make(map[string]*Tile)
	tileCacheLeft := make(map[String2Tuple]*Tile)
	tileCacheRight := make(map[String2Tuple]*Tile)
	for _, tile := range t.tiles {
		tileCacheBottom[tile.Sides[Down].Label] = &tile
		tileCacheLeft[String2Tuple{tile.Sides[Left].Label, tile.Sides[Down].Label}] = &tile
		tileCacheRight[String2Tuple{tile.Sides[Right].Label, tile.Sides[Down].Label}] = &tile
	}

	log.Println("Drawing tile images...")
	for _, tile := range t.tiles {
		tile.Image = t.generateImage(&tile)
	}

	/*
	   # TODO

	   # figure out how big a 'normal' character is for approximate layout purposes
	   my @font_bounds = GD::Image->stringFT( 0, $ttf_font, $font_size, 0, 0, 0, '5' );
	   die "Error: couldn't use font $ttf_font:$font_size: $@" unless @font_bounds;
	   my $font_w = $font_bounds[ 2 ] - $font_bounds[ 6 ];
	   my $font_h = $font_bounds[ 3 ] - $font_bounds[ 7 ];
	   my $font_x = $font_bounds[ 6 ];
	   my $font_y = $font_bounds[ 7 ];
	*/
	for _, input := range t.Inputs {
		log.Printf("Processing input %q...", input)

		// check that the input string has only legal symbols
		symbolRx := regexp.MustCompile(fmt.Sprintf("[^%s]", string(t.Symbols)))
		if symbolRx.MatchString(input) {
			log.Printf("  Warning: invalid symbol encountered in input string %q", input)
			continue
		}

		// annotate initial input with head semantics before generating starter tiles
		cells := make([]Cell, len(input))
		for i, r := range input {
			cells[i] = Cell{r, i == t.InitialLocation}
		}

		log.Printf("Generating starter tiles...")

		// seed first row of assembly with starter tiles from input-generated cells
		assembly := [][]*Tile{make([]*Tile, len(cells)+2)}

		// wrap with boundary symbols at beginning and end
		assembly[0][0] = t.cellToTile(&Cell{t.BoundarySymbol, false}, true, false)
		for i, cell := range cells {
			assembly[0][i+1] = t.cellToTile(&cell, false, false)
		}
		assembly[0][len(assembly)-1] = t.cellToTile(&Cell{t.BoundarySymbol, false}, false, true)

		log.Printf("Assembling transition tiles...")
		/*

		       // assemble until we hit a halting state or the limit is reached
		       1 while addTile( \@assembly, \@tiles ) && ( !$max_depth || @assembly <= $max_depth + 1 );

		       if ( $max_depth && @assembly > $max_depth + 1 ) {
		           pop @assembly; // remove the offending line
		           warn "  Warning: assembly hit maximum depth ($max_depth), increase with --max-depth=\n";
		           next unless $ignore_depth_failure;
		       }

		       // remove trailing blank line
		       pop @assembly;

		       my $size_x = @{ $assembly[ 0 ] };
		       my $size_y = @assembly;

		       print STDERR "  Transforming matrix...\n";

		       // rotation occurs after assembly so that the assembler routine may concern
		       // itself only with a single logical layout (bottom-up). here we rotate the
		       // assembly matrix to the desired final orientation; tile orientation is
		       // also rotated, but in the drawing routines
		       ( $size_x, $size_y, @assembly ) = computeRotated( \@assembly, $size_x, $size_y );

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
		   }
		*/
		log.Printf("Done!")
	}
}

/*

// starting with the current assembly, try to add any tile drawn from the pool
// that fits in an empty spot adjacent to an existing tile. tiles may only be
// added if at least two bonds are made in so doing (i.e. two single bonds or
// one double bond).
sub addTile {
    my $assembly = shift;
    my $tiles = shift;

    # search over the entire current assembly
    for my $y ( max( 0, @$assembly - 2 ) .. @$assembly - 1 ) {
        for my $x ( 0 .. @{ $assembly->[ $y ] } - 1 ) {
            my $tile = $assembly->[ $y ][ $x ]; # this tile
            next unless defined $tile;

            my $ltile = $assembly->[ $y ][ $x - 1 ]; # left neighbor
            my $rtile = $assembly->[ $y ][ $x + 1 ]; # right neighbor
            my $ttile = $assembly->[ $y + 1 ][ $x ]; # top neighbor
            my $lltile = $assembly->[ $y - 1 ][ $x - 1 ]; # left and down
            my $lrtile = $assembly->[ $y - 1 ][ $x + 1 ]; # right and down

            # try to add tile left
            if (
                !defined $ltile && # can't step on an existing tile
                defined $lltile && # can only add left if left-down is there
                (
                    $tile->{ sides }{ +LEFT }{ bond_strength } +
                    $lltile->{ sides }{ +TOP }{ bond_strength }
                ) >= 2 # would-be bond sum is good enough
            ) {
                # space looks good, search for a matching tile in the pool
                my $stile = $tile_cache_right
                    { $tile->{ sides }{ +LEFT }{ label } }
                    { $lltile->{ sides }{ +TOP }{ label } };
                if ( defined $stile ) {
                    # found a match, fill it in
                    $assembly->[ $y ][ $x - 1 ] = $stile;
                    return 1;
                }
            }

            # try to add tile right
            if (
                !defined $rtile && # can't step on an existing tile
                defined $lrtile && # can only add right if right-down is there
                (
                    $tile->{ sides }{ +RIGHT }{ bond_strength } +
                    $lrtile->{ sides }{ +TOP }{ bond_strength } >= 2
                ) # would-be bond sum is good enough
            ) {
                # space looks good, search for a matching tile in the pool
                my $stile = $tile_cache_left
                    { $tile->{ sides }{ +RIGHT }{ label } }
                    { $lrtile->{ sides }{ +TOP }{ label } };
                if ( defined $stile ) {
                    # found a match, fill it in
                    $assembly->[ $y ][ $x + 1 ] = $stile;
                    return 1;
                }
            }

            # try to add tile up; adding vertically can ONLY happen by double-bond
            if (
                !defined $ttile &&
                $tile->{ sides }{ +TOP }{ bond_strength } >= 2
            ) {
                my $stile = $tile_cache_bottom{ $tile->{ sides }{ +TOP }{ label } };
                if ( defined $stile ) {
                    # found a match, fill it in
                    $assembly->[ $y + 1 ][ $x ] = $stile;
                    return 1;
                }
            }
        }
    }

    # couldn't find any tile to add anywhere, so halt
    return 0;
}
*/

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
	draw.Draw(im, r, &image.Uniform{bgColor}, image.ZP, draw.Src)

	for _, side := range []Direction{Up, Down, Left, Right} {
		strength := tile.Sides[side].Strength
		label := tile.Sides[side].Label

		orientedLabel := fmt.Sprintf("%d%s%d", side%2, label, strength)
		color := t.getLabelColor(orientedLabel, false)

		t.drawBond(im, side, strength, color)
		t.drawString(im, side, strength, color, label)
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

	r1 := float64(i1) / float64(math.MaxUint64)
	r2 := float64(i2) / float64(math.MaxUint64)
	r3 := float64(i3) / float64(math.MaxUint64)

	h, s, v := r1, 0.0, 0.0
	if bright {
		v = 0.9 + 0.1*r3
		s = 0.0 + 0.02*r2
	} else {
		v = 0.1 + 0.4*r3
		s = 0.5 + 0.5*r2
	}

	cp := math.Floor(h * 6)
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

	r = 255 * math.Max(0, math.Min(1, r))
	g = 255 * math.Max(0, math.Min(1, g))
	b = 255 * math.Max(0, math.Min(1, b))

	c := color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b)}
	if t.colors == nil {
		t.colors = make(map[string]color.RGBA)
	}
	t.colors[label] = c
	return c
}

func (t *Tiler) drawBond(im draw.Image, side Direction, strength int, color color.RGBA) {
	rotSide := Direction((int(side) + t.Rotation) % 4)
	if t.FlipHorizontal {
		if rotSide == Left {
			rotSide = Right
		} else if rotSide == Right {
			rotSide = Left
		}
	}
	if t.FlipVertical {
		if rotSide == Up {
			rotSide = Down
		} else if rotSide == Down {
			rotSide = Up
		}
	}

	for i := 0; i < strength; i++ {
		var r image.Rectangle
		switch rotSide {
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

func (t *Tiler) drawString(im image.Image, side Direction, strength int, color color.RGBA, label string) {
	/*

	   bond_shift := bond_strength - 1

	   my ( $x, $y );

	   my $rot_side = ( $side + $rotation ) % 4;
	   if ( $flip_horiz ) {
	       if ( $rot_side == LEFT ) {
	           $rot_side = RIGHT;
	       }
	       elsif ( $rot_side == RIGHT ) {
	           $rot_side = LEFT;
	       }
	   }
	   if ( $flip_vert ) {
	       if ( $rot_side == TOP ) {
	           $rot_side = BOTTOM;
	       }
	       elsif ( $rot_side == BOTTOM ) {
	           $rot_side = TOP;
	       }
	   }

	   if ( $rot_side == TOP ) {
	       $y = t.tileVertMargin + t.bondShift * t.tileVertShift;
	       $x = int( ( t.TileWidth - length( $string ) * $font_w ) / 2 );
	   }
	   elsif ( $rot_side == BOTTOM ) {
	       $y = t.TileHeight - $font_h - t.tileVertMargin - t.bondShift * t.tileVertShift;
	       $x = int( ( t.TileWidth - length( $string ) * $font_w ) / 2 );
	   }
	   elsif ( $rot_side == LEFT ) {
	       $y = int( ( t.TileHeight - $font_h ) / 2 );
	       $x = t.tileHorizMargin + t.bondShift * t.tileHorizShift;
	   }
	   elsif ( $rot_side == RIGHT ) {
	       $y = int( ( t.TileHeight - $font_h ) / 2 );
	       $x = t.TileWidth - t.tileHorizMargin - length( $string ) * $font_w - t.bondShift * t.tileHorizShift;
	   }
	   $x -= $font_x;
	   $y -= $font_y;

	   #$img->string( $font, $x, $y, $string, $color );
	   $img->stringFT( $color, $ttf_font, $font_size, 0, $x, $y, $string );
	*/
}

// rotate the @assembly matrix
/*
sub computeRotated {
    my $old_assembly = shift;
    my $size_x = shift;
    my $size_y = shift;
    my $new_assembly;

    for my $j ( 0 .. $size_y - 1 ) {
        for my $i ( 0 .. $size_x - 1 ) {
            if ( $rotation == 0 ) {
                # no rotation, but possibly flips
                my $ti = $i;
                my $tj = $j;
                $ti = $size_x - 1 - $ti if     $flip_horiz;
                $tj = $size_y - 1 - $tj if     $flip_vert;
                $new_assembly->[ $tj ][ $ti ] = $old_assembly->[ $j ][ $i ];
            }
            elsif ( $rotation == 1 ) {
                # CCW 90
                my $ti = $i;
                my $tj = $j;
                $ti = $size_x - 1 - $ti if     $flip_horiz;
                $tj = $size_y - 1 - $tj unless $flip_vert;
                $new_assembly->[ $ti ][ $tj ] = $old_assembly->[ $j ][ $i ];
            }
            elsif ( $rotation == 2 ) {
                # 180
                my $ti = $i;
                my $tj = $j;
                $ti = $size_x - 1 - $ti unless $flip_horiz;
                $tj = $size_y - 1 - $tj unless $flip_vert;
                $new_assembly->[ $tj ][ $ti ] = $old_assembly->[ $j ][ $i ];
            }
            elsif ( $rotation == 3 ) {
                # CW 90
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
}
*/
