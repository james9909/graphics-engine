package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

const (
	DefaultBasename = "frame"  // Default frame basename
	FramesDirectory = "frames" // FramesDirectory is the directory containing all animation frames
)

// Parser is a script parser
type Parser struct {
	drawer   *Drawer   // image drawer
	lexer    *Lexer    // lexer
	backup   []Token   // token backup
	commands []Command // list of commands

	isAnimated   bool                 // whether or not to parse as an animation
	frames       int                  // number of frames in the animation
	basename     string               // animation basename
	formatString string               // format string for each frame of the animation
	symbols      *SymbolTable         // symbol table
	knobs        map[string][]float64 // knob table
}

// NewParser returns a new parser
func NewParser() *Parser {
	return &Parser{
		drawer:     NewDrawer(DefaultHeight, DefaultWidth),
		backup:     make([]Token, 0, 50),
		isAnimated: false,
		symbols:    NewSymbolTable(),
		knobs:      make(map[string][]float64),
	}
}

// ParseInput parses a file for commands and executes them
func (p *Parser) ParseInput() error {
	scanner := bufio.NewScanner(os.Stdin)
	var input bytes.Buffer
	for scanner.Scan() {
		input.Write(scanner.Bytes())
		input.WriteRune('\n')
	}
	err := p.ParseString(input.String())
	return err
}

// ParseFile parses a file for commands and executes them
func (p *Parser) ParseFile(filename string) error {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = p.ParseString(string(input))
	return err
}

// ParseString parses a string for commands and executes them
func (p *Parser) ParseString(input string) error {
	p.lexer = Lex(input)
	commands, err := p.parse()
	if err == nil {
		p.commands = commands
		err = p.process()
	}
	return err
}

func (p *Parser) parse() ([]Command, error) {
	commands := make([]Command, 0, 50)
	for {
		t := p.nextToken()
		switch t.tt {
		case tError:
			return nil, errors.New(t.value)
		case tEOF:
			if p.isAnimated {
				if p.basename == "" {
					fmt.Fprintf(os.Stderr, "No basename provided: using default basename '%s'\n", DefaultBasename)
					p.basename = DefaultBasename
					p.formatString = fmt.Sprintf("%s/%s-%%0%dd.png", FramesDirectory, p.basename, len(strconv.Itoa(p.frames)))
				}
			}
			return commands, nil
		case tIdent:
			var command Command
			switch LookupIdent(t.value) {
			case MOVE:
				c := MoveCommand{}
				c.args = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				if p.peek().tt == tString {
					c.knob = p.nextToken().value
					p.symbols.Set(c.knob, 0)
				}
				command = c
			case SCALE:
				c := ScaleCommand{}
				c.args = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				if p.peek().tt == tString {
					c.knob = p.nextToken().value
					p.symbols.Set(c.knob, 0)
				}
				command = c
			case ROTATE:
				c := RotateCommand{}
				c.axis = p.nextIdent()
				c.degrees = p.nextFloat()
				if p.peek().tt == tString {
					c.knob = p.nextToken().value
					p.symbols.Set(c.knob, 0)
				}
				command = c
			case LINE:
				c := LineCommand{}
				if p.peek().tt == tString {
					c.constants = p.nextToken().value
				}
				c.p1 = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				if p.peek().tt == tString {
					c.cs = p.nextToken().value
				}
				c.p2 = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				if p.peek().tt == tString {
					c.cs2 = p.nextToken().value
				}
				command = c
			case SPHERE:
				c := SphereCommand{}
				if p.peek().tt == tString {
					c.constants = p.nextToken().value
				}
				c.center = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				c.radius = p.nextFloat()
				if p.peek().tt == tString {
					c.cs = p.nextToken().value
				}
				command = c
			case TORUS:
				c := TorusCommand{}
				if p.peek().tt == tString {
					c.constants = p.nextToken().value
				}
				c.center = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				c.r1 = p.nextFloat()
				c.r2 = p.nextFloat()
				if p.peek().tt == tString {
					c.cs = p.nextToken().value
				}
				command = c
			case BOX:
				c := BoxCommand{}
				if p.peek().tt == tString {
					c.constants = p.nextToken().value
				}
				c.p1 = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				c.width = p.nextFloat()
				c.height = p.nextFloat()
				c.depth = p.nextFloat()
				if p.peek().tt == tString {
					c.cs = p.nextToken().value
				}
				command = c
			case POP:
				command = PopCommand{}
			case PUSH:
				command = PushCommand{}
			case SAVE:
				command = SaveCommand{
					filename: p.nextString(),
				}
			case DISPLAY:
				command = DisplayCommand{}
			case VARY:
				if p.frames == 0 {
					return nil, errors.New("number of frames is not set")
				}
				name := p.nextString()
				knob, found := p.knobs[name]
				if !found {
					knob = make([]float64, p.frames)
				}
				startFrame := p.nextInt()
				if startFrame < 0 || startFrame >= p.frames {
					return nil, fmt.Errorf("invalid start frame %d for knob %s", name, startFrame)
				}
				endFrame := p.nextInt()
				if endFrame < 0 || endFrame >= p.frames || endFrame < startFrame {
					return nil, fmt.Errorf("invalid end frame %d for knob %s", name, endFrame)
				}
				startValue := p.nextFloat()
				endValue := p.nextFloat()
				length := endFrame - startFrame
				delta := (endValue - startValue) / float64(length+1)
				for frame := startFrame; frame <= endFrame; frame++ {
					knob[frame] = startValue
					startValue += delta
				}
				p.knobs[name] = knob
				p.isAnimated = true
			case BASENAME:
				if p.basename != "" {
					fmt.Fprintln(os.Stderr, "Setting the basename multiple times")
				}
				p.basename = p.nextString()
				p.formatString = fmt.Sprintf("%s/%s-%%0%dd.png", FramesDirectory, p.basename, len(strconv.Itoa(p.frames)))
				p.isAnimated = true
			case FRAMES:
				if p.frames != 0 {
					fmt.Fprintln(os.Stderr, "Setting the number of frames multiple times")
				}
				p.frames = p.nextInt()
				if p.frames <= 0 {
					return nil, errors.New("number of frames must be greater than zero")
				}
				p.isAnimated = true
			case SET:
				c := SetCommand{
					name:  p.nextString(),
					value: p.nextFloat(),
				}
				command = c
			case SETALL:
				c := SetAllCommand{
					value: p.nextFloat(),
				}
				command = c
			}
			if command != nil {
				commands = append(commands, command)
			}
			if err := p.expect(tNewline); err != nil {
				return nil, fmt.Errorf("unexpected %v at end of statement", p.peek().tt)
			}
		case tString:
			return nil, fmt.Errorf("unrecognized identifier: \"%s\"", t.value)
		}
	}
	return commands, nil
}

func (p *Parser) process() error {
	if p.isAnimated {
		os.RemoveAll(FramesDirectory)
		os.Mkdir(FramesDirectory, 0755)
	} else {
		p.frames = 1
	}
	var err error
	for frame := 0; frame < p.frames; frame++ {
		for knob := range p.knobs {
			p.symbols.Set(knob, p.knobs[knob][frame])
		}
		if p.isAnimated {
			fmt.Printf("Rendering frame %d/%d", frame+1, p.frames)
		}
		start := time.Now()
		for _, command := range p.commands {
			switch command.(type) {
			case MoveCommand:
				c := command.(MoveCommand)
				x, y, z := c.args[0], c.args[1], c.args[2]
				if c.knob != "" {
					if knob, err := p.getSymbolValue(c.knob); err == nil {
						x *= knob
						y *= knob
						z *= knob
					} else {
						return err
					}
				}
				err = p.drawer.Move(x, y, z)
			case ScaleCommand:
				c := command.(ScaleCommand)
				x, y, z := c.args[0], c.args[1], c.args[2]
				if c.knob != "" {
					if knob, err := p.getSymbolValue(c.knob); err == nil {
						x *= knob
						y *= knob
						z *= knob
					} else {
						return err
					}
				}
				err = p.drawer.Scale(x, y, z)
			case RotateCommand:
				c := command.(RotateCommand)
				degrees := c.degrees
				if c.knob != "" {
					if knob, err := p.getSymbolValue(c.knob); err == nil {
						degrees *= knob
					} else {
						return err
					}
				}
				err = p.drawer.Rotate(c.axis, degrees)
			case LineCommand:
				c := command.(LineCommand)
				err = p.drawer.Line(c.p1[0], c.p1[1], c.p1[2], c.p2[0], c.p2[1], c.p2[2])
			case SphereCommand:
				c := command.(SphereCommand)
				err = p.drawer.Sphere(c.center[0], c.center[1], c.center[2], c.radius)
			case TorusCommand:
				c := command.(TorusCommand)
				err = p.drawer.Torus(c.center[0], c.center[1], c.center[2], c.r1, c.r2)
			case BoxCommand:
				c := command.(BoxCommand)
				err = p.drawer.Box(c.p1[0], c.p1[1], c.p1[2], c.width, c.height, c.depth)
			case PopCommand:
				p.drawer.Pop()
			case PushCommand:
				p.drawer.Push()
			case SaveCommand:
				c := command.(SaveCommand)
				err = p.drawer.Save(c.filename)
			case DisplayCommand:
				err = p.drawer.Display()
			case SetCommand:
				c := command.(SetCommand)
				p.symbols.Set(c.name, c.value)
			case SetAllCommand:
				c := command.(SetAllCommand)
				for key := range p.symbols.table {
					p.symbols.Set(key, c.value)
				}
			}
			if err != nil {
				return err
			}
		}
		elapsed := time.Since(start)
		if p.isAnimated {
			fmt.Printf(" - %s\n", elapsed)
			err = p.drawer.Save(fmt.Sprintf(p.formatString, frame))
			if err != nil {
				return err
			}
			p.drawer.Reset()
		}
	}
	if p.isAnimated {
		fmt.Println("Making animation...")
		err = MakeAnimation(p.basename)
	}
	return err
}

func (p *Parser) getSymbolValue(symbolName string) (float64, error) {
	if value, found := p.symbols.Get(symbolName); found {
		return value, nil
	} else {
		return 0, fmt.Errorf("undefined symbol '%s'", symbolName)
	}
}

// nextToken returns the nextToken token from the lexer
func (p *Parser) nextToken() Token {
	lenBackup := len(p.backup)
	// Use the token from backup if it exists
	if lenBackup > 0 {
		token := p.backup[lenBackup-1]
		p.backup = p.backup[:lenBackup-1]
		return token
	}
	token := p.lexer.NextToken()
	return token
}

// nextInt returns the next integer token from the lexer
// Panics if the next token is not an integer
func (p *Parser) nextInt() int {
	if p.expect(tInt) != nil {
		panic(fmt.Errorf("expected %v, got %v", tInt, p.peek().tt))
	}
	v, _ := strconv.Atoi(p.nextToken().value)
	return v
}

// nextFloat returns the next token from the lexer as a float.
// Panics if the next token is not a float or integer
func (p *Parser) nextFloat() float64 {
	if p.expect(tInt) != nil && p.expect(tFloat) != nil {
		panic(fmt.Errorf("expected %v, got %v", tFloat, p.peek().tt))
	}
	v, _ := strconv.ParseFloat(p.nextToken().value, 64)
	return v
}

// nextString returns the next token from the lexer.
// Panics if the next token is not a string
func (p *Parser) nextString() string {
	if p.expect(tString) != nil {
		panic(fmt.Errorf("expected %v, got %v", tString, p.peek().tt))
	}
	return p.nextToken().value
}

// nextIdent returns the next identifier from the lexer as a string.
// Panics if the next token is not an identifier
func (p *Parser) nextIdent() string {
	if p.expect(tIdent) != nil {
		panic(fmt.Errorf("expected %v, got %v", tIdent, p.peek().tt))
	}
	return p.nextToken().value
}

// unread adds the token to the list of backup tokens.
// Since channels cannot be "unread", we use a list to backup these tokens
func (p *Parser) unread(token Token) {
	p.backup = append(p.backup, token)
}

// peek returns the next token without consuming it
func (p *Parser) peek() Token {
	token := p.next()
	p.unread(token)
	return token
}

// expect returns an error if the next token is not a certain type
func (p *Parser) expect(tt TokenType) error {
	other := p.peek().tt
	if other != tt {
		return fmt.Errorf("expected %v, got %v", tt, other)
	}
	return nil
}
