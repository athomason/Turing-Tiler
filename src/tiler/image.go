package tiler

import (
	"io/ioutil"
	"log"

	"code.google.com/p/freetype-go/freetype"
)

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

	/*
		   # TODO

			GD::Image->trueColor(1);

			# figure out how big a 'normal' character is for approximate layout purposes
			my @font_bounds = GD::Image->stringFT( 0, $ttf_font, $font_size, 0, 0, 0, '5' );
			die "Error: couldn't use font $ttf_font:$font_size: $@" unless @font_bounds;
			my $font_w = $font_bounds[ 2 ] - $font_bounds[ 6 ];
			my $font_h = $font_bounds[ 3 ] - $font_bounds[ 7 ];
			my $font_x = $font_bounds[ 6 ];
			my $font_y = $font_bounds[ 7 ];

			print STDERR "Drawing tile images...\n";

			my %color_cache;

			# draw an image for each tile
			generateTileImage( $_ ) for @tiles;

			for my $input_string ( @input_strings ) {
				print STDERR "Processing input $input_string...\n";

				# check that the input string has only legal symbols
				my $symbol_re = sprintf "[^%s]", join '', keys %symbols;
				if ( $input_string =~ /($symbol_re)/ ) {
					warn "  Warning: invalid symbol ($1) encountered in input string\n";
					next;
				}

				# annotate initial input with head semantics before generating starter tiles
				my @cells;
				push @cells, { head => 0, symbol => $_ } for split //, $input_string;
				$cells[$initial_location]{head} = 1;

				# wrap in boundary symbols at beginning and end
				@cells = (
					{ head => 0, symbol => $boundary_symbol },
					@cells,
					{ head => 0, symbol => $boundary_symbol },
				);

				# seed assembly with starter tiles from cells
				my @assembly = [ ];

				print STDERR "  Generating starter tiles...\n";

				my $leftcell = shift @cells;
				my $rightcell = pop @cells;
				push @{ $assembly[ 0 ] }, cellToTile( $leftcell,  1, 0 );
				push @{ $assembly[ 0 ] }, cellToTile( $_,         0, 0 ) for @cells;
				push @{ $assembly[ 0 ] }, cellToTile( $rightcell, 0, 1 );

				print STDERR "  Assembling transition tiles...\n";

				# assemble until we hit a halting state or the limit is reached
				1 while addTile( \@assembly, \@tiles ) && ( !$max_depth || @assembly <= $max_depth + 1 );

				if ( $max_depth && @assembly > $max_depth + 1 ) {
					pop @assembly; # remove the offending line
					warn "  Warning: assembly hit maximum depth ($max_depth), increase with --max-depth=\n";
					next unless $ignore_depth_failure;
				}

				# remove trailing blank line
				pop @assembly;

				my $size_x = @{ $assembly[ 0 ] };
				my $size_y = @assembly;

				print STDERR "  Transforming matrix...\n";

				# rotation occurs after assembly so that the assembler routine may concern
				# itself only with a single logical layout (bottom-up). here we rotate the
				# assembly matrix to the desired final orientation; tile orientation is
				# also rotated, but in the drawing routines
				( $size_x, $size_y, @assembly ) = computeRotated( \@assembly, $size_x, $size_y );

				# create the master canvas containing the record of the entire computation
				my $target = GD::Image->new( $tile_width * $size_x - $size_x + 1, $tile_height * $size_y - $size_y + 1 );
				$target->saveAlpha( 1 );

				#use Data::Dumper;
				#print Data::Dumper->Dump( [ \@tiles, \@assembly ], [ qw/ tiles assembly / ] );

				print STDERR "  Generating canvas...\n";

				# copy component tiles to the master canvas
				for my $i ( 0 .. @assembly - 1 ) {
					for my $j ( 0 .. @{ $assembly[ $i ] } - 1 ) {
						my $src = $assembly[ $i ][ $j ]{ image };
						next unless defined $src; # silently ignored missing tiles
						$target->copy( $src,
							$tile_width * $j - $j,
							$tile_height * ( @assembly - $i - 1 ) - ( @assembly - $i - 1 ),
							0, 0, $tile_width, $tile_height
						);
					}
				}

				# save the output

				my $output_file = "$name-$input_string.png";
				print STDERR "  Saving image $output_file...\n";
				open OUTPUT, ">", $output_file;
				binmode OUTPUT;
				print OUTPUT $target->png;
				close OUTPUT;
			}

			print STDERR "Done!\n";

			# END MAIN
	*/
}

/*

# starting with the current assembly, try to add any tile drawn from the pool
# that fits in an empty spot adjacent to an existing tile. tiles may only be
# added if at least two bonds are made in so doing (i.e. two single bonds or
# one double bond).
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

# build an initial tile from a cell definition
sub cellToTile {
    my $cell = shift;
    my $left = shift;
    my $right = shift;

    my $tile = {
        name => "seed",
        sides => {
            TOP, {
                bond_strength => $cell->{ head } ? 2 : 1,
                label => $cell->{ head } ?
                    "$initial_state $cell->{ symbol }" :
                    $cell->{ symbol },
            },
            BOTTOM, {
                bond_strength => 1,
                label => '',
            },
            LEFT, {
                bond_strength => $left ? 1 : 2,
                label => '',
            },
            RIGHT, {
                bond_strength => $right ? 1 : 2,
                label => '',
            },
        },
    };

    generateTileImage( $tile );

    return $tile;
}

sub generateTileImage {
    my $tile = shift;

    my $img = new GD::Image( $tile_width, $tile_height );

    my $bg_color = $img->colorAllocate( getLabelColor(
        sprintf( "background%d", exists $tile->{ final } ), 1
    ) );
    $img->filledRectangle( 0, 0, $tile_width, $tile_height, $bg_color );

    for my $side ( TOP, BOTTOM, LEFT, RIGHT ) {
        my $strength = $tile->{ sides }{ $side }{ bond_strength };
        my $label = $tile->{ sides }{ $side }{ label };

        my $oriented_label = sprintf "%d$label$strength", $side % 2;
        my $color = $img->colorAllocate( getLabelColor( $oriented_label ) );

        drawBond( $img, $side, $strength, $color );
        drawString( $img, $side, $strength, $label, $color );
    }

    $tile->{ image } = $img;
}

# for the given label, return a visually well-distributed color that is always
# the same but uncorrelated to the label's contents
sub getLabelColor {
    my $label = shift;
    my $bright = shift || 0;

    return @{ $color_cache{ $label } } if exists $color_cache{ $label };

    # use a has function to get some random-ish data based on the label.
    # here we get 32 bytes from MD5 and unpack the first 24 as longs.
    use Digest::MD5;
    $label .= $prng_tweak;
    my $hash = Digest::MD5->md5( $label ) . Digest::MD5->md5( reverse $label );
    my ( $r1, $r2, $r3 ) = unpack "L*", substr $hash, 0, 24;
    # get a fractional value from the longs
    $_ = exp( log( $_ ) - 32 * log 2 ) for ( $r1, $r2, $r3 );

    my $h = 0.0 + 1.0 * $r1;
    my $s;
    my $v;
    if ( $bright ) {
        $v = 0.9 + 0.1 * $r3;
        $s = 0.0 + 0.02 * $r2;
    }
    else {
        $v = 0.1 + 0.4 * $r3;
        $s = 0.5 + 0.5 * $r2;
    }

    my $cp = int( $h * 6 );
    my $cs = $h - $cp;
    my $ca = ( 1 - $s ) * $v;
    my $cb = ( 1 - ( $s * $cs ) ) * $v;
    my $cc = ( 1 - ( $s * ( 1 - $cs ) ) ) * $v;

    my ( $r, $g, $b );
    if    ( $cp == 0 ) { $r = $v; $g = $cc; $b = $ca; }
    elsif ( $cp == 1 ) { $r = $cb; $g = $v; $b = $ca; }
    elsif ( $cp == 2 ) { $r = $ca; $g = $v; $b = $cc; }
    elsif ( $cp == 3 ) { $r = $ca; $g = $cb; $b = $v; }
    elsif ( $cp == 4 ) { $r = $cc; $g = $ca; $b = $v; }
    elsif ( $cp == 5 ) { $r = $v; $g = $ca; $b = $cb; }

    $r = 0 if $r < 0; $r = 1 if $r > 1;
    $g = 0 if $g < 0; $g = 1 if $g > 1;
    $b = 0 if $b < 0; $b = 1 if $b > 1;
    $r *= 255;
    $g *= 255;
    $b *= 255;

    my $values = [ int $r, int $g, int $b ];
    $color_cache{ $label } = $values;

    return @$values;
}

sub drawBond {
    my $img = shift;
    my $side = shift;
    my $bond_strength = shift;
    my $color = shift;

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

    for my $i ( 0 .. $bond_strength - 1 ) {
        if ( $rot_side == TOP ) {
            #$img->line(
            #    0, $i * $tile_vert_shift, $tile_width - 1,
            #    $i * $tile_vert_shift,
            #    $color
            #);
            $img->filledRectangle(
                0, $i * $tile_vert_shift - $bond_fudge_y,
                $tile_width - 1, $i * $tile_vert_shift + $bond_fudge_y,
                $color
            );
        }
        elsif ( $rot_side == BOTTOM ) {
            $img->filledRectangle(
                0, $tile_height - 1 - $i * $tile_vert_shift - $bond_fudge_y,
                $tile_width - 1, $tile_height - 1 - $i * $tile_vert_shift + $bond_fudge_y,
                $color
            );
        }
        elsif ( $rot_side == LEFT ) {
            $img->filledRectangle(
                $i * $tile_horiz_shift - $bond_fudge_x, 0,
                $i * $tile_horiz_shift + $bond_fudge_x, $tile_height - 1,
                $color
            );
        }
        elsif ( $rot_side == RIGHT ) {
            $img->filledRectangle(
                $tile_width - 1 - $i * $tile_horiz_shift - $bond_fudge_x, 0,
                $tile_width - 1 - $i * $tile_horiz_shift + $bond_fudge_x, $tile_height - 1,
                $color
            );
        }
    }
}

sub drawString {
    my $img = shift;
    my $side = shift;
    my $bond_strength = shift;
    my $bond_shift = $bond_strength - 1;
    my $string = shift;
    my $color = shift;

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
        $y = $tile_vert_margin + $bond_shift * $tile_vert_shift;
        $x = int( ( $tile_width - length( $string ) * $font_w ) / 2 );
    }
    elsif ( $rot_side == BOTTOM ) {
        $y = $tile_height - $font_h - $tile_vert_margin - $bond_shift * $tile_vert_shift;
        $x = int( ( $tile_width - length( $string ) * $font_w ) / 2 );
    }
    elsif ( $rot_side == LEFT ) {
        $y = int( ( $tile_height - $font_h ) / 2 );
        $x = $tile_horiz_margin + $bond_shift * $tile_horiz_shift;
    }
    elsif ( $rot_side == RIGHT ) {
        $y = int( ( $tile_height - $font_h ) / 2 );
        $x = $tile_width - $tile_horiz_margin - length( $string ) * $font_w - $bond_shift * $tile_horiz_shift;
    }
    $x -= $font_x;
    $y -= $font_y;

    #$img->string( $font, $x, $y, $string, $color );
    $img->stringFT( $color, $ttf_font, $font_size, 0, $x, $y, $string );
}

# rotate the @assembly matrix
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
