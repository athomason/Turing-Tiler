#!/usr/bin/perl

use 5.6.0;
use strict;
use warnings;

use GD;

use constant TOP => 0;
use constant BOTTOM => 1;
use constant LEFT => 2;
use constant RIGHT => 3;

use constant HEIGHT => 60;
use constant WIDTH => 60;

use constant HSHIFT => WIDTH / 10;
use constant VSHIFT => HEIGHT / 10;

use constant HMARGIN => 2;
use constant VMARGIN => 1;

use constant FONT => gdSmallFont( );

my $font_w = FONT->width( );
my $font_h = FONT->height( );

my $filename = shift @ARGV;

my $me = $0;
$me =~ s,.*/,,;
die "Usage: $me <tape_spec>\n" unless $filename;

my @cells;
my $base;

open FILE, "< $filename" or die "Could not open $filename";
while ( <FILE> ) {
    chomp;
    if ( /^CELL\s+(\w+)/ ) {
        push @cells, { head => 0, symbol => $1 };
    }
    elsif ( /^HEAD\s+(\d+)\s+(\w+)/ ) {
        push @cells, { head => $1, symbol => $2 };
    }
    elsif ( /^BASE\s+(\S+)/ ) {
        $base = $1;
    }
}
close FILE;

die "No cells specified" unless @cells;
die "No filebase specified" unless $base;

my @tiles;

if ( @cells == 1 ) {
    my $tile = GD::Image->new( WIDTH, HEIGHT );
    my $white = $tile->colorAllocate( 255, 255, 255 );
    my $black = $tile->colorAllocate( 0, 0, 0 );
    drawCell( $tile, $cells[ 0 ], 1, 1, $black );
    push @tiles, $tile;
}
else {
    my $leftcell = shift @cells;
    my $rightcell = pop @cells;

    my $lefttile = GD::Image->new( WIDTH, HEIGHT );
    my $leftwhite = $lefttile->colorAllocate( 255, 255, 255 );
    my $leftblack = $lefttile->colorAllocate( 0, 0, 0 );
    drawCell( $lefttile, $leftcell, 1, 0, $leftblack );
    push @tiles, $lefttile;

    for my $cell ( @cells ) {
        my $midtile = GD::Image->new( WIDTH, HEIGHT );
        my $midwhite = $midtile->colorAllocate( 255, 255, 255 );
        my $midblack = $midtile->colorAllocate( 0, 0, 0 );

        drawCell( $midtile, $cell, 0, 0, $midblack );
        push @tiles, $midtile;
    }

    my $righttile = GD::Image->new( WIDTH, HEIGHT );
    my $rightwhite = $righttile->colorAllocate( 255, 255, 255 );
    my $rightblack = $righttile->colorAllocate( 0, 0, 0 );
    drawCell( $righttile, $rightcell, 0, 1, $rightblack );

    push @tiles, $righttile;
}

my $target = GD::Image->new( WIDTH * @tiles, HEIGHT );

for my $i ( 0 .. @tiles - 1 ) {
    my $tile = $tiles[ $i ];

    $target->copy( $tile, WIDTH * $i, 0, 0, 0, WIDTH, HEIGHT );

    open OUTPUT, sprintf "> %s-%d.png", $base, $i + 1;
    binmode OUTPUT;
    print OUTPUT $tile->png;
    close OUTPUT;
}

open OUTPUT, "> seed.png";
binmode OUTPUT;
print OUTPUT $target->png;
close OUTPUT;

sub drawCell {
    my $tile = shift;
    my $cell = shift;
    my $left = shift;
    my $right = shift;
    my $color = shift;

    drawBond( $tile, TOP, $cell->{ head } ? 2 : 1, $color );
    drawString( $tile, TOP, $cell->{ head } ? 2 : 1,
        $cell->{ head } ? "$cell->{ head } $cell->{ symbol }" : $cell->{ symbol },
        $color
    );
    drawBond( $tile, BOTTOM, 1, $color );
    drawBond( $tile, LEFT, $left ? 1 : 2, $color );
    drawBond( $tile, RIGHT, $right ? 1 : 2, $color );
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
