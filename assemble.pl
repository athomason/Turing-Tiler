#!/usr/bin/perl

use 5.6.0;
use strict;
use warnings;

use GD;
use Getopt::Long;
use List::Util 'max';

use constant TOP => 0;
use constant LEFT => 1;
use constant BOTTOM => 2;
use constant RIGHT => 3;

GetOptions(
    # 4/3 is a reasonable aspect ratio for single-character states and symbols
    'tile-height=i'             => \(my $tile_height            = 24),
    'tile-width=i'              => \(my $tile_width             = 32),

    # the following have sensibly computed defaults based on the tile size
    'tile-horiz-shift=i'        => \(my $tile_horiz_shift),
    'tile-vert-shift=i'         => \(my $tile_vert_shift),
    'bond-fudge-horiz=i'        => \(my $bond_fudge_x),
    'bond-fudge-vert=i'         => \(my $bond_fudge_y),
    'tile-horiz-margin=i'       => \(my $tile_horiz_margin),
    'tile-vert-margin=i'        => \(my $tile_vert_margin),

    # to prevent runaways of non-halting programs
    'max-depth=i'               => \(my $max_depth              = 100),
    'ignore-depth-failure!'     => \(my $ignore_depth_failure),

    'font=s'                    => \(my $ttf_font               = '/usr/share/fonts/truetype/ttf-bitstream-vera/Vera.ttf'),
    'font-size=f'               => \(my $font_size),

    # 0/vert is best for portrait or web (top->down)
    # 3 is best for landscape or monitors (left->right)
    'rotation=i'                => \(my $rotation               = 0),
    'flip-horiz!'               => \(my $flip_horiz             = 0),
    'flip-vert!'                => \(my $flip_vert              = 1),

    'boundary-symbol=s'         => \(my $boundary_symbol        = '*'),

    # colorspace PRNG seed tweaker
    'prng-tweak=s'              => \(my $prng_tweak = ''),
);
my $machine_filename = shift @ARGV;
my @input_strings = @ARGV;

die "Error: requested font file $ttf_font does not exist" unless -e $ttf_font;

my $me = $0;
$me =~ s,.*/,,;
die "Usage: $me [options] <machine_spec> <input_string> [<input_string>] [...]\n"
    unless defined $machine_filename && @input_strings;

GD::Image->trueColor(1);

# calculate sane defaults
$bond_fudge_x = int sprintf "%.0f", $tile_width / 80 unless defined $bond_fudge_x;
$bond_fudge_y = int sprintf "%.0f", $tile_height / 80 unless defined $bond_fudge_y;

$font_size = sqrt( $tile_height * $tile_width ) / 4 unless defined $font_size;

$tile_horiz_shift = $tile_width / 10 unless defined $tile_horiz_shift;
$tile_vert_shift = $tile_height / 10 unless defined $tile_vert_shift;

$tile_horiz_margin = $tile_width / 20 unless defined $tile_horiz_margin;
$tile_vert_margin = $tile_height / 20 unless defined $tile_vert_margin;

# figure out how big a 'normal' character is for approximate layout purposes
my @font_bounds = GD::Image->stringFT( 0, $ttf_font, $font_size, 0, 0, 0, '5' );
die "Error: couldn't use font $ttf_font:$font_size: $@" unless @font_bounds;
my $font_w = $font_bounds[ 2 ] - $font_bounds[ 6 ];
my $font_h = $font_bounds[ 3 ] - $font_bounds[ 7 ];
my $font_x = $font_bounds[ 6 ];
my $font_y = $font_bounds[ 7 ];

my $name = $machine_filename; $name =~ s,.*/,,; $name =~ s/\..*//;
my %symbols;
my @transitions;
my $initial_state = 1;
my $initial_location = 0;

print STDERR "Parsing machine...\n";

# read the machine
open MACHINEFILE, "<", $machine_filename or die "Error: could not open $machine_filename";
while ( <MACHINEFILE> ) {
    chomp;
    s/^\s+//; # strip leading whitespace
    s/#.*//; # strip trailing comments
    if ( /^\s*$/ ) {
        next;
    }
    elsif ( /^NAME\s+(\S+)/ ) {
        $name = $1;
    }
    elsif ( /^SYMBOL\s+(\S+)/ ) {
        $symbols{$1}++;
    }
    elsif ( /^START\s+(\S+)/ ) {
        $initial_state = $1;
    }
    elsif ( /^OFFSET\s+(\d+)/ ) {
        $initial_location = $1;
        die "initial location must be nonnegative" if $initial_location < 0;
    }
    elsif ( /^TRANSITION\s+
        (\S+)\s+ # head state
        (\S+)\s+ # tape symbol
        (\S+)\s+ # write symbol
        ([HhLlRr])\s+ # move action
        (\S+) # new state
        (?:\s+(\S+))? # halting states' output
        /x )
    {
        push @transitions, {
            oldstate => $1,
            readsymbol => $2,
            writesymbol => $3,
            move => uc $4,
            newstate => $5,
            output => $6,
        };
        if ( uc $4 ne 'H' && defined $6 ) {
            warn "Warning: halting string given for non-halting transition: $_\n";
        }
    }
    else {
        warn "Warning: statement could not be parsed: $_\n";
    }
}
close MACHINEFILE;

die "Error: no symbols specified" unless %symbols;
die "Error: no transitions specified" unless @transitions;

$symbols{$boundary_symbol}++;

# the tile "pool": things the self-assembler can draw from
my @tiles;

print STDERR "Generating tileset 1/3...\n";

# first set of tiles: transitions from old head states
for my $transition ( @transitions ) {
    if ( $transition->{ move } eq 'L' ) {
        push @tiles, {
            name => "transition-$transition->{ oldstate }-$transition->{ readsymbol }",
            sides => {
                TOP, {
                    bond_strength => 1,
                    label => $transition->{ writesymbol },
                },
                BOTTOM, {
                    bond_strength => 2,
                    label => "$transition->{ oldstate } $transition->{ readsymbol }",
                },
                LEFT, {
                    bond_strength => 1,
                    label => $transition->{ newstate },
                },
                RIGHT, {
                    bond_strength => 1,
                    label => "R",
                },
            },
        };
    }
    elsif ( $transition->{ move } eq 'R' ) {
        push @tiles, {
            name => "transition-$transition->{ oldstate }-$transition->{ readsymbol }",
            sides => {
                TOP, {
                    bond_strength => 1,
                    label => $transition->{ writesymbol },
                },
                BOTTOM, {
                    bond_strength => 2,
                    label => "$transition->{ oldstate } $transition->{ readsymbol }",
                },
                LEFT, {
                    bond_strength => 1,
                    label => 'L',
                },
                RIGHT, {
                    bond_strength => 1,
                    label => $transition->{ newstate },
                },
            },
        };
    }
    elsif ( $transition->{ move } eq 'H' ) {
        push @tiles, {
            name => "transition-$transition->{ oldstate }-$transition->{ readsymbol }",
            sides => {
                TOP, {
                    bond_strength => 1,
                    label =>
                        defined $transition->{ output } ?
                            "$transition->{ writesymbol } [$transition->{ output }]" :
                            $transition->{ writesymbol },
                },
                BOTTOM, {
                    bond_strength => 2,
                    label => "$transition->{ oldstate } $transition->{ readsymbol }",
                },
                LEFT, {
                    bond_strength => 1,
                    label => 'L',
                },
                RIGHT, {
                    bond_strength => 1,
                    label => 'R',
                },
            },
            final => 1,
        };
    }
}

print STDERR "Generating tileset 2/3...\n";

my %states = map { ($_->{oldstate} => 1, $_->{newstate} => 1) } @transitions;

# second set of tiles: exposes a double bond from the new head state
for my $state (keys %states) {
    for my $symbol (keys %symbols) {
        # moving left
        push @tiles, {
            name => "move-$state-$symbol-left",
            sides => {
                TOP, {
                    bond_strength => 2,
                    label => "$state $symbol",
                },
                BOTTOM, {
                    bond_strength => 1,
                    label => $symbol,
                },
                LEFT, {
                    bond_strength => 1,
                    label => 'L',
                },
                RIGHT, {
                    bond_strength => 1,
                    label => $state,
                },
            },
        };

        # moving right
        push @tiles, {
            name => "move-$state-$symbol-right",
            sides => {
                TOP, {
                    bond_strength => 2,
                    label => "$state $symbol",
                },
                BOTTOM, {
                    bond_strength => 1,
                    label => $symbol,
                },
                LEFT, {
                    bond_strength => 1,
                    label => $state,
                },
                RIGHT, {
                    bond_strength => 1,
                    label => 'R',
                },
            },
        };
    }
}

print STDERR "Generating tileset 3/3...\n";

# third set of tiles: replicates non-head state cells
for my $symbol (keys %symbols) {
    # copying left of head
    push @tiles, {
        name => "replicate-$symbol-left",
        sides => {
            TOP, {
                bond_strength => 1,
                label => $symbol,
            },
            BOTTOM, {
                bond_strength => 1,
                label => $symbol,
            },
            LEFT, {
                bond_strength => 1,
                label => 'L',
            },
            RIGHT, {
                bond_strength => 1,
                label => 'L',
            },
        },
    };

    # copying right of head
    push @tiles, {
        name => "replicate-$symbol-right",
        sides => {
            TOP, {
                bond_strength => 1,
                label => $symbol,
            },
            BOTTOM, {
                bond_strength => 1,
                label => $symbol,
            },
            LEFT, {
                bond_strength => 1,
                label => 'R',
            },
            RIGHT, {
                bond_strength => 1,
                label => 'R',
            },
        },
    };
}

print STDERR "Generating tile caches...\n";
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
