#!/usr/bin/perl

# creates a contact sheet of the rule tiles

use 5.6.0;
use strict;
use warnings;

use GD;

use constant HEIGHT => 60;
use constant WIDTH => 60;

my @mfiles = <move*.png>;
my @rfiles = <replicate*.png>;
my @tfiles = <transition*.png>;

my $x = @mfiles * WIDTH;
my $y = 3 * HEIGHT;

my $target = GD::Image->new( $x, $y );

for my $i ( 0 .. @mfiles - 1 ) {
    my $src = GD::Image->newFromPng( $mfiles[ $i ] );
    $target->copy( $src, WIDTH * $i, HEIGHT * 0, 0, 0, WIDTH, HEIGHT );
}

for my $i ( 0 .. @rfiles - 1 ) {
    my $src = GD::Image->newFromPng( $rfiles[ $i ] );
    $target->copy( $src, WIDTH * $i, HEIGHT * 1, 0, 0, WIDTH, HEIGHT );
}

for my $i ( 0 .. @tfiles - 1 ) {
    my $src = GD::Image->newFromPng( $tfiles[ $i ] );
    $target->copy( $src, WIDTH * $i, HEIGHT * 2, 0, 0, WIDTH, HEIGHT );
}

open OUTPUT, "> rules.png";
binmode OUTPUT;
print OUTPUT $target->png;
close OUTPUT;
