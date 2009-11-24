#!/usr/bin/perl

use 5.6.0;
use strict;
use warnings;

$|++;

print STDERR "What is the name of this machine? ";
my $name = <>;
chomp $name;

my $filename = "$name.machine";
if ( -e $filename ) {
    die "$filename already exists! Please remove it first.\n";
}

print STDERR "How many states does this machine have? ";
my $states = <>;
chomp $states;

print STDERR "What non-boundary symbols are allowed on the tape? ";
my @symbols = ( grep( { /\S/ } split //, <> ), '*' );

open FILE, ">", $filename or die "Could not write to $filename: $!\n";

print FILE "NAME $name\n\n";
print FILE "STATES $states\n\n";
print FILE "SYMBOL $_\n" for @symbols;

for my $state ( 1 .. $states ) {
    print FILE "\n";
    for my $symbol ( @symbols ) {
        print FILE "TRANSITION $state $symbol $symbol h $state\n";
    }
}

close FILE;

print STDERR "\n$filename has been generated.\n";
