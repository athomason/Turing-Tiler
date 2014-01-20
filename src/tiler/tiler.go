package tiler

import (
	. "math"
)

const (
	top = iota
	left
	bottom
	right
)

type Options struct {
	TileHeight, TileWidth        int
	MaxDepth                     int
	IgnoreDepthFailure           bool
	FontPath                     string
	Rotation                     int // 0 is best for portrait or web (top->down), 3 is best for landscape or monitors (left->right)
	FlipHorizontal, FlipVertical bool
	BoundarySymbol               string
	MachineFile string
	Inputs []string
}

type Tiler struct {
	Options

	// computed options
	fontSize               int
	bondFudgeX, bondFudgeY int
	tileHorizShift, tileVertShift,
	tileHorizMargin, tileVertMargin float64

	machine Machine
}

func (o *Options) NewTiler() *Tiler {
	t := Tiler{
		Options:         *o,
		bondFudgeX:      int(Floor(float64(o.TileWidth) / 80)),
		bondFudgeY:      int(Floor(float64(o.TileHeight) / 80)),
		fontSize:        int(Sqrt(float64(o.TileHeight*o.TileWidth)) / 4),
		tileHorizShift:  float64(o.TileWidth) / 10,
		tileVertShift:   float64(o.TileHeight) / 10,
		tileHorizMargin: float64(o.TileWidth) / 20,
		tileVertMargin:  float64(o.TileHeight) / 20,
	}
	t.ParseMachine()
	t.GenerateTiles()
	return &t
}

func (t *Tiler) GenerateTiles() {
	/*
	# the tile "pool": things the self-assembler can draw from
	my @tiles;

	print STDERR "Generating tileset 1/3...\n";

	# first set of tiles: transitions from old head states
	for my $transition ( @transitions ) {
		if ( $transition->{ move } eq 'L' ) {
			push @tiles, {
				name => "transition-$transition->{ oldstate }-$transition->{ readsymbol }",
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
				name => "transition-$transition->{ oldstate }-$transition->{ readsymbol }",
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
						label => 'L',
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
				name => "transition-$transition->{ oldstate }-$transition->{ readsymbol }",
				sides => {
					TOP, {
						bond_strength => 1,
						label =>
							defined $transition->{ output } ?
								"$transition->{ writesymbol } [$transition->{ output }]" :
								$transition->{ writesymbol },
					},
					BOTTOM, {
						bond_strength => 2,
						label => "$transition->{ oldstate } $transition->{ readsymbol }",
					},
					LEFT, {
						bond_strength => 1,
						label => 'L',
					},
					RIGHT, {
						bond_strength => 1,
						label => 'R',
					},
				},
				final => 1,
			};
		}
	}

	print STDERR "Generating tileset 2/3...\n";

	my %states = map { ($_->{oldstate} => 1, $_->{newstate} => 1) } @transitions;

	# second set of tiles: exposes a double bond from the new head state
	for my $state (keys %states) {
		for my $symbol (keys %symbols) {
			# moving left
			push @tiles, {
				name => "move-$state-$symbol-left",
				sides => {
					TOP, {
						bond_strength => 2,
						label => "$state $symbol",
					},
					BOTTOM, {
						bond_strength => 1,
						label => $symbol,
					},
					LEFT, {
						bond_strength => 1,
						label => 'L',
					},
					RIGHT, {
						bond_strength => 1,
						label => $state,
					},
				},
			};

			# moving right
			push @tiles, {
				name => "move-$state-$symbol-right",
				sides => {
					TOP, {
						bond_strength => 2,
						label => "$state $symbol",
					},
					BOTTOM, {
						bond_strength => 1,
						label => $symbol,
					},
					LEFT, {
						bond_strength => 1,
						label => $state,
					},
					RIGHT, {
						bond_strength => 1,
						label => 'R',
					},
				},
			};
		}
	}

	print STDERR "Generating tileset 3/3...\n";

	# third set of tiles: replicates non-head state cells
	for my $symbol (keys %symbols) {
		# copying left of head
		push @tiles, {
			name => "replicate-$symbol-left",
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
					label => 'L',
				},
				RIGHT, {
					bond_strength => 1,
					label => 'L',
				},
			},
		};

		# copying right of head
		push @tiles, {
			name => "replicate-$symbol-right",
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
					label => 'R',
				},
				RIGHT, {
					bond_strength => 1,
					label => 'R',
				},
			},
		};
	}

	print STDERR "Generating tile caches...\n";
	my ( %tile_cache_left, %tile_cache_right, %tile_cache_bottom );

	for my $tile ( @tiles ) {
		$tile_cache_bottom
			{ $tile->{ sides }{ +BOTTOM }{ label } } = $tile;
		$tile_cache_left
			{ $tile->{ sides }{ +LEFT   }{ label } }
			{ $tile->{ sides }{ +BOTTOM }{ label } } = $tile;
		$tile_cache_right
			{ $tile->{ sides }{ +RIGHT  }{ label } }
			{ $tile->{ sides }{ +BOTTOM }{ label } } = $tile;
	}
	*/
}
