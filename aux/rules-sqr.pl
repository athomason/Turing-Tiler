#!/usr/bin/perl

# creates a contact sheet of the rule tiles

use 5.6.0;
use strict;
use warnings;

use GD;

use constant HEIGHT => 30;
use constant WIDTH => 30;

my @files;
push @files, <move*.png>;
push @files, <replicate*.png>;
push @files, <transition*.png>;

my $columns = int sqrt( @files );

my $x = $columns * WIDTH;
my $y = int( @files / $columns ) * HEIGHT;

my $target = GD::Image->new( $x, $y );

for my $i ( 0 .. @files - 1 ) {
    my $src = GD::Image->newFromPng( $files[ $i ] );
    $target->copy( $src, WIDTH * ( $i % $columns ), HEIGHT * int( $i / $columns ), 0, 0, WIDTH, HEIGHT );
}

open OUTPUT, "> sqr_rules.png";
binmode OUTPUT;
print OUTPUT $target->png;
close OUTPUT;
