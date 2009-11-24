#!/usr/bin/perl

use 5.6.0;
use strict;
use warnings;

my $filename = shift;
my $string = shift;

my $me = $0;
$me =~ s,.*/,,;
die "Usage: $me <machine_spec> <input_string>\n" unless defined $filename && defined $string; 

$string = lc $string;

my @symbols;
open FILE, "< $filename" or die "Could not open $filename";
while ( <FILE> ) {
    chomp;
    if ( /^SYMBOL\s+(\w+)/ ) {
        push @symbols, $1;
    }
}
close FILE;

my $symbol_re = sprintf "[^%s]", join '', @symbols;

die "Invalid symbol encountered: $1\n" if $string =~ /($symbol_re)/;

open OUTPUT, "> $string.seed" or die "Could not open $string.seed";
print OUTPUT "BASE $string\n";
print OUTPUT "CELL *\n";
my @chars = split //, $string;
printf OUTPUT "HEAD 1 %s\n", shift @chars;
printf OUTPUT "CELL %s\n", shift @chars while @chars;
print OUTPUT "CELL *\n";
close OUTPUT;
