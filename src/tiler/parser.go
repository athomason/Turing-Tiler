package tiler

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Machine struct {
	Name            string
	Symbols         []rune
	Transitions     []Transition
	InitialState    string
	InitialLocation int
}

type Transition struct {
	OldState, ReadSymbol, WriteSymbol string
	Move                              Direction
	NewState, Output                  string
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

func (t *Tiler) ParseMachine() *Machine {
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
		Symbols:         make([]rune, 0),
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
			r, _ := utf8.DecodeRune([]byte(c[1]))
			m.Symbols = append(m.Symbols, r)
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
			t := Transition{c[1], c[2], c[3], letterToDirection(c[4]), c[5], c[6]}
			if t.Output != "" && t.Move != Halt {
				log.Panicf("Halting string given for non-halting transition %v: %q", t.Move, line)
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
	return &m
}

func letterToDirection(letter string) Direction {
	switch strings.ToLower(letter) {
	case "l":
		return Left
	case "r":
		return Right
	case "h":
		return Halt
	}
	log.Panicf("parserTransitionRx should only match /[HhLlRr]/ but got %q", letter)
	return 0
}
