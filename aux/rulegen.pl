#!/usr/bin/perl

use 5.6.0;
use strict;
use warnings;

use GD;

use constant TOP => 0;
use constant BOTTOM => 1;
use constant LEFT => 2;
use constant RIGHT => 3;

use constant HEIGHT => 30;
use constant WIDTH => 30;

use constant HSHIFT => WIDTH / 10;
use constant VSHIFT => HEIGHT / 10;

use constant HMARGIN => 2;
use constant VMARGIN => 1;

use constant FONT => gdTinyFont( );

my $font_w = FONT->width( );
my $font_h = FONT->height( );

# read in machine spec
my @symbols;
my $states;
my @transitions;

my $filename = shift @ARGV;

my $me = $0;
$me =~ s,.*/,,;
die "Usage: $me <machine_spec>\n" unless $filename;

open FILE, "< $filename" or die "Could not open $filename";
while ( <FILE> ) {
    chomp;
    if ( /^SYMBOL\s+(\w+)/ ) {
        push @symbols, $1;
    }
    elsif ( /^STATES\s+(\d+)/ ) {
        $states = $1;
    }
    elsif ( /^TRANSITION\s+
        (\d+)\s+ # head state
        (\w+)\s+ # tape symbol
        (\w+)\s+ # write symbol
        ([HhLlRr])\s+ # move action
        (\d+) # new state
        (?:\s+(\S+))? # halting states' output
        /x ) {
        push @transitions, {
            oldstate => $1,
            readsymbol => $2,
            writesymbol => $3,
            move => uc $4,
            newstate => $5,
            output => $6,
        };
    }
    else {
        warn "Unknown or flawed record: $_\n";
    }
}
close FILE;

die "No symbols specified" unless @symbols;
die "Number of states not specified" unless defined $states;
die "No transitions specified" unless @transitions;

my @tiles;

# first set of tiles: transitions from old head states
for my $transition ( @transitions ) {
    if ( $transition->{ move } eq 'L' ) {
        push @tiles, {
            name => "transition-$transition->{ oldstate }-$transition->{ readsymbol }.png",
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
            name => "transition-$transition->{ oldstate }-$transition->{ readsymbol }.png",
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
                    label => "L",
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
            name => "transition-$transition->{ oldstate }-$transition->{ readsymbol }.png",
            sides => {
                TOP, {
                    bond_strength => 1,
                    label =>
                        defined $transition->{ output } ?
                            "$transition->{ writesymbol } $transition->{ output }" :
                            $transition->{ writesymbol },
                },
                BOTTOM, {
                    bond_strength => 2,
                    label => "$transition->{ oldstate } $transition->{ readsymbol }",
                },
                LEFT, {
                    bond_strength => 1,
                    label => "L",
                },
                RIGHT, {
                    bond_strength => 1,
                    label => "R",
                },
            },
        };
    }
}

# second set of tiles: exposes a double bond from the new head state
for my $state ( 1 .. $states ) {
    for my $symbol ( @symbols ) {
        # moving left
        push @tiles, {
            name => "move-$state-$symbol-left.png",
            sides => {
                TOP, {
                    bond_strength => 2,
                    label => "$state $symbol",
                },
                BOTTOM, {
                    bond_strength => 1,
                    label => "$symbol",
                },
                LEFT, {
                    bond_strength => 1,
                    label => "L",
                },
                RIGHT, {
                    bond_strength => 1,
                    label => "$state",
                },
            },
        };

        # moving right
        push @tiles, {
            name => "move-$state-$symbol-right.png",
            sides => {
                TOP, {
                    bond_strength => 2,
                    label => "$state $symbol",
                },
                BOTTOM, {
                    bond_strength => 1,
                    label => "$symbol",
                },
                LEFT, {
                    bond_strength => 1,
                    label => "$state",
                },
                RIGHT, {
                    bond_strength => 1,
                    label => "R",
                },
            },
        };
    }
}

# third set of tiles: replicates non-head state cells
for my $symbol ( @symbols ) {
    # copying left of head
    push @tiles, {
        name => "replicate-$symbol-left.png",
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
                label => "L",
            },
            RIGHT, {
                bond_strength => 1,
                label => "L",
            },
        },
    };

    # copying right of head
    push @tiles, {
        name => "replicate-$symbol-right.png",
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
                label => "R",
            },
            RIGHT, {
                bond_strength => 1,
                label => "R",
            },
        },
    };
}

for my $tile ( @tiles ) {
    my $img = new GD::Image( WIDTH, HEIGHT );
    my $white = $img->colorAllocate( 255, 255, 255 );
    my $black = $img->colorAllocate( 0, 0, 0 );

    for my $side ( TOP, BOTTOM, LEFT, RIGHT ) {
        my $strength = $tile->{ sides }{ $side }{ bond_strength };
        my $label = $tile->{ sides }{ $side }{ label };
        drawBond( $img, $side, $strength, $black );
        drawString( $img, $side, $strength, $label, $black );
    }

    open OUTPUT, "> $tile->{ name }";

    binmode OUTPUT;

    print OUTPUT $img->png;

    close OUTPUT;
}

sub drawBond {
    my $img = shift;
    my $side = shift;
    my $bond_strength = shift;
    my $color = shift;

    for my $i ( 0 .. $bond_strength - 1 ) {
        if ( $side == TOP ) {
            $img->line( 0, $i * VSHIFT, WIDTH - 1, $i * VSHIFT, $color );
        }
        elsif ( $side == BOTTOM ) {
            $img->line( 0, HEIGHT - 1 - $i * VSHIFT, WIDTH - 1, HEIGHT - 1 - $i * VSHIFT, $color );
        }
        elsif ( $side == LEFT ) {
            $img->line( $i * HSHIFT, 0, $i * HSHIFT, HEIGHT - 1, $color );
        }
        elsif ( $side == RIGHT ) {
            $img->line( WIDTH - 1 - $i * HSHIFT, 0, WIDTH - 1 - $i * HSHIFT, HEIGHT - 1, $color );
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

    if ( $side == TOP ) {
        $y = VMARGIN + $bond_shift * VSHIFT;
        $x = int( ( WIDTH - length( $string ) * $font_w ) / 2 );
    }
    elsif ( $side == BOTTOM ) {
        $y = HEIGHT - $font_h - VMARGIN - $bond_shift * VSHIFT;
        $x = int( ( WIDTH - length( $string ) * $font_w ) / 2 );
    }
    elsif ( $side == LEFT ) {
        $y = int( ( HEIGHT - $font_h ) / 2 );
        $x = HMARGIN + $bond_shift * HSHIFT;
    }
    elsif ( $side == RIGHT ) {
        $y = int( ( HEIGHT - $font_h ) / 2 );
        $x = WIDTH - HMARGIN - length( $string ) * $font_w - $bond_shift * HSHIFT;
    }

    $img->string( FONT, $x, $y, $string, $color );
}
