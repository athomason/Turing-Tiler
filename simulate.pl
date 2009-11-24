#!/usr/bin/perl

use 5.6.0;
use strict;
use warnings;

use Getopt::Long;
use Time::HiRes 'sleep';

GetOptions(
    # to prevent runaways of non-halting programs
    'max-depth=i'               => \(my $max_depth              = 100),
    'ignore-depth-failure!'     => \(my $ignore_depth_failure),

    'boundary-symbol=s'         => \(my $boundary_symbol        = '*'),

    'sleep=f'                   => \(my $sleep),
    'clear!'                    => \(my $clear),

    'all!'                      => \(my $all                    = 1),
);
my $machine_filename = shift @ARGV;
my $input_string = shift @ARGV;

(my $me = $0) =~ s,.*/,,;
die "Usage: $me [options] <machine_spec> <input_string>\n"
    unless defined $machine_filename && defined $input_string;

my $name = $machine_filename; $name =~ s,.*/,,; $name =~ s/\..*//;
my @symbols;
my @transitions;
my $initial_state = 1;

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
        push @symbols, $1;
    }
    elsif ( /^START\s+(\S+)/ ) {
        $initial_state = $1;
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

die "Error: no symbols specified" unless @symbols;
die "Error: no transitions specified" unless @transitions;

my %transitions;
for my $t (@transitions) {
    $transitions{ $t->{oldstate} }{ $t->{readsymbol} } = $t;
}

print STDERR "Simulating for input $input_string...\n";

# check that the input string has only legal symbols
my $symbol_re = sprintf "[^%s]", join '', @symbols;
if ($input_string =~ /($symbol_re)/) {
    die "invalid symbol ($1) encountered in input string\n";
}

my @tape;
push @tape, $boundary_symbol;
push @tape, split //, $input_string;
push @tape, $boundary_symbol;

my $state = $initial_state;
my $pos = 1;

$|++;

my $n = 0;

my $tape = join '', @tape;
#substr $tape, $pos, 1, uc substr $tape, $pos, 1;
printf "%4d (%2s/%2d) %s\n", $n, $state, $pos, $tape;
printf "             %s^\n", ' 'x$pos;
print "\n";

while (++$n) {

    my $symbol = $tape[$pos];
    my $transition = $transitions{$state}{$symbol};
    die "missing transition for $state/$symbol" unless $transition;

    # write symbol
    my $newsymbol = $transition->{writesymbol};
    $tape[$pos] = $newsymbol;

    # change state
    $state = $transition->{newstate};

    # move head
    if ($transition->{move} eq 'L') {
        $pos--;
    }
    elsif ($transition->{move} eq 'R') {
        $pos++;
    }
    else {
        printf "%4d (%2s/%2d) Halt: '%s'\n", $n, $state, $pos, $transition->{output} || '';
        last;
    }

    my $tape = join '', @tape;
    #substr $tape, $pos, 1, uc substr $tape, $pos, 1;

    system("clear") if $clear;
    printf "%4d (%2s/%2d) %s%s", $n, $state, $pos, $tape, $symbol eq $newsymbol && !$all ? "\r" : "\n";
    printf "             %s^\n", ' 'x$pos;
    print "\n";
    sleep $sleep if $sleep;
    #die "off the tape" if $pos < 0;
}

#use YAML::Syck;
#print Dump({symbols => \@symbols, states => $states, transitions => \%transitions, tape => \@tape});

