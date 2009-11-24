#!/usr/bin/perl

use 5.6.0;
use strict;
use warnings;

use GD;

use constant HEIGHT => 60;
use constant WIDTH => 60;

my @files;

while ( <> ) {
    chomp;
    push @files, [ split ];
}

my $x = @{ $files[ 0 ] } * WIDTH;
my $y = @files * HEIGHT;

my $target = GD::Image->new( $x, $y );

for my $i ( 0 .. @files - 1 ) {
    for my $j ( 0 .. @{ $files[ $i ] } - 1 ) {
        my $src = GD::Image->newFromPng( $files[ $i ][ $j ] . ".png" );
        $target->copy( $src, WIDTH * $j, HEIGHT * $i, 0, 0, WIDTH, HEIGHT );
    }
}

open OUTPUT, "> assembly.png";
binmode OUTPUT;
print OUTPUT $target->png;
close OUTPUT;
