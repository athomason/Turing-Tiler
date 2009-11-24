#!/usr/bin/perl
use strict;
use warnings;

# given a machine definition on STDIN, prints a dot(1) source representing the
# machine's state diagram

# a graph may be created by e.g.
#   cat machine.def | perl machine2dot.pl | dot -Tpng - -o machine.png

use Getopt::Long;

GetOptions(
    'html!' => \(my $html),
);

my %states;
my @lines;

my $left_symbol = '<';
my $right_symbol = '>';

if ($html) {
    $left_symbol = " &#8592;"; # &larr;
    $right_symbol = " &#8594;"; # &rarr;
}

print "digraph machine {\n";
while ( <STDIN> ) {
    chomp;
    next unless /TRANSITION/;
    my ( undef, $source, $read, $write, $move, $target, $halt ) = split;
    $states{ $source }++;
    $states{ $target }++;
    push @lines, sprintf "    S%s->S%s [label=\"{$read,$write:%s}\"]\n",
        $source, $target,
        $move eq 'r' ? $right_symbol : $move eq 'l' ? $left_symbol :
        defined $halt ? $halt : ''
    ;
}
print "    S$_ [label=\"$_\"]\n" for sort keys %states;
print for @lines;
print "}\n";
