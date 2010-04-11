#!/usr/bin/perl
use strict;
use warnings;

# given a machine definition on STDIN, prints a dot(1) source representing the
# machine's state diagram

# a graph may be created by e.g.
#   cat machine.def | perl machine2dot.pl | dot -Tpng - -o machine.png

use Getopt::Long;

GetOptions(
    'concise!'  => \(my $concise = 1),
    'html!'     => \(my $html),
);

my $left_symbol = '<';
my $right_symbol = '>';

if ($html) {
    $left_symbol = " &#8592;"; # &larr;
    $right_symbol = " &#8594;"; # &rarr;
}

print "digraph machine {\n";

my $initial_state = '1';

my %states;
my %transitions;
while (<STDIN>) {
    chomp;
    s/#.*//;

    if (/^START\s+(\S+)/) {
        $initial_state = $1;
        next;
    }

    next unless /TRANSITION/i;

    my (undef, $source, $read, $write, $move, $target, $halt) = split;
    $states{$source}++;
    $states{$target}++;
    $halt ||= '';
    push @{ $transitions{$source}{$target}{$move}{$halt} }, [$read, $write];
}

my @states = grep {$_ ne $initial_state} sort keys %states;
unshift @states, $initial_state;
for my $state (@states) {
    if ($state eq $initial_state) {
        print "    S$state [label=\"$state\", style=filled, fillcolor=black, fontcolor=white]\n";
    }
    else {
        print "    S$state [label=\"$state\"]\n";
    }
}

for my $source (keys %transitions) {
    my $s_h = $transitions{$source};
    for my $target (keys %$s_h) {
        my $t_h = $s_h->{$target};
        for my $move (keys %$t_h) {
            my $m_h = $t_h->{$move};
            for my $halt (keys %$m_h) {
                my $h_h = $m_h->{$halt};

                my $color = $halt ? 'red' : 'black';

                if ($concise && @$h_h > 1) {
                    my @ts = sort { $a->[0] cmp $b->[0] } @$h_h;
                    my $reads =  join '', map { $_->[0] } @ts;
                    my $writes = join '', map { $_->[1] } @ts;
                    printf qq(    S%s->S%s [label="{[$reads],[$writes]:%s}",color=$color]\n),
                        $source, $target,
                        $move eq 'r' ? $right_symbol : $move eq 'l' ? $left_symbol :
                        $halt
                    ;
                }
                else {
                    for my $t (@$h_h) {
                        my $read =  $t->[0];
                        my $write = $t->[1];
                        printf qq(    S%s->S%s [label="{$read,$write:%s}",color=$color]\n),
                            $source, $target,
                            $move eq 'r' ? $right_symbol : $move eq 'l' ? $left_symbol :
                            $halt
                        ;
                    }
                }
            }
        }
    }
}

print "}\n";
