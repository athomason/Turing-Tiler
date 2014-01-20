package tiler

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Transition struct {
	OldState, ReadSymbol, WriteSymbol, Move, NewState, Output string
}

type Machine struct {
	Name            string
	Symbols         []string
	Transitions     []Transition
	InitialState    string
	InitialLocation int
}

var (
	parserWhitespaceRx = regexp.MustCompile("^\\s+")
	parserCommentRx    = regexp.MustCompile("#.*")
	parserNameRx       = regexp.MustCompile("^NAME\\s+(\\S+)")
	parserSymbolRx     = regexp.MustCompile("^SYMBOL\\s+(\\S+)")
	parserStartRx      = regexp.MustCompile("^START\\s+(\\S+)")
	parserOffsetRx     = regexp.MustCompile("^OFFSET\\s+(\\d+)")
	parserTransitionRx = regexp.MustCompile("^TRANSITION\\s+(\\S+)\\s+(\\S+)\\s+(\\S+)\\s+([HhLlRr])\\s+(\\S+)(?:\\s+(\\S+))?")
)

func (t *Tiler) ParseMachine() {
	f, err := os.Open(t.MachineFile)
	if err != nil {
		log.Panicf("Couldn't open %s: %s", t.MachineFile, err)
	}
	defer f.Close()

	// default to filename with dirname and extension stripped
	name := t.MachineFile
	if n := strings.LastIndex(name, "/"); n >= 0 && n < len(name) {
		name = name[n+1:]
	}
	if n := strings.Index(name, "."); n >= 0 {
		name = name[:n]
	}

	m := Machine{
		Name:            name,
		Symbols:         make([]string, 0),
		Transitions:     make([]Transition, 0),
		InitialState:    "1",
		InitialLocation: 0,
	}

	log.Printf("Parsing machine from %q...", t.MachineFile)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		line = parserWhitespaceRx.ReplaceAllString(line, "")
		line = parserCommentRx.ReplaceAllString(line, "")
		if len(line) == 0 {
			continue
		}
		if c := parserNameRx.FindStringSubmatch(line); c != nil {
			m.Name = c[1]
		} else if c := parserSymbolRx.FindStringSubmatch(line); c != nil {
			m.Symbols = append(m.Symbols, c[1])
		} else if c := parserStartRx.FindStringSubmatch(line); c != nil {
			m.InitialState = c[1]
		} else if c := parserOffsetRx.FindStringSubmatch(line); c != nil {
			m.InitialLocation, _ = strconv.Atoi(c[1]) // can't err because \d+
		} else if c := parserTransitionRx.FindStringSubmatch(line); c != nil {
			/*
				more clearly:
				/^TRANSITION\s+
					(\S+)\s+      # head state
					(\S+)\s+      # tape symbol
					(\S+)\s+      # write symbol
					([HhLlRr])\s+ # move action
					(\S+)         # new state
					(?:\s+(\S+))? # halting states' output
					/x
			*/
			t := Transition{c[1], c[2], c[3], c[4], c[5], c[6]}
			if t.Output != "" && strings.ToUpper(t.Move) != "H" {
				log.Panicf("Halting string given for non-halting transition: %q", line)
			}
			m.Transitions = append(m.Transitions, t)
		} else {
			log.Printf("Warning: statement could not be parsed: %q", line)
		}
	}

	if len(m.Symbols) == 0 {
		log.Panicf("No symbols specified")
	}
	if len(m.Transitions) == 0 {
		log.Panicf("No transitions specified")
	}
	m.Symbols = append(m.Symbols, t.BoundarySymbol)
	t.machine = m
}
