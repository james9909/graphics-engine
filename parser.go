package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
)

const (
	DefaultBasename = "frame"  // Default frame basename
	FramesDirectory = "frames" // FramesDirectory is the directory containing all animation frames
	MaxWorkers      = 2        // maximum number of workers
)

var knobs map[string][]float64 // knob table

// Lighting
var ambient []float64                   // ambient lighting
var lightSources map[string]LightSource // light table
var constants map[string][][]float64    // constants table

var formatString string // format string for each frame of the animation

func init() {
	knobs = make(map[string][]float64)

	lightSources = make(map[string]LightSource)
	constants = make(map[string][][]float64)
}

// Parser is a script parser
type Parser struct {
	lexer  *Lexer  // lexer
	backup []Token // token backup

	isAnimated bool   // whether or not to parse as an animation
	frames     int    // number of frames in the animation
	basename   string // animation basename
}

// NewParser returns a new parser
func NewParser() *Parser {
	return &Parser{
		backup:     make([]Token, 0, 10),
		isAnimated: false,
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
		err = p.process(commands)
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
					formatString = fmt.Sprintf("%s/%s-%%0%dd.png", FramesDirectory, p.basename, len(strconv.Itoa(p.frames)))
				}
			}
			return commands, nil
		case tIdent:
			var command Command
			switch LookupIdent(t.value) {
			case MOVE:
				c := MoveCommand{}
				c.args = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				c.knob, _ = p.next(tString)
				command = c
			case SCALE:
				c := ScaleCommand{}
				c.args = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				c.knob, _ = p.next(tString)
				command = c
			case ROTATE:
				c := RotateCommand{}
				c.axis = p.nextIdent()
				c.degrees = p.nextFloat()
				c.knob, _ = p.next(tString)
				command = c
			case LINE:
				c := LineCommand{}
				c.constants, _ = p.next(tString)
				c.p1 = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				c.cs, _ = p.next(tString)
				c.p2 = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				c.cs2, _ = p.next(tString)
				command = c
			case SPHERE:
				c := SphereCommand{}
				c.constants, _ = p.next(tString)
				c.center = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				c.radius = p.nextFloat()
				c.cs, _ = p.next(tString)
				command = c
			case TORUS:
				c := TorusCommand{}
				c.constants, _ = p.next(tString)
				c.center = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				c.r1 = p.nextFloat()
				c.r2 = p.nextFloat()
				c.cs, _ = p.next(tString)
				command = c
			case BOX:
				c := BoxCommand{}
				c.constants, _ = p.next(tString)
				c.p1 = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				c.width = p.nextFloat()
				c.height = p.nextFloat()
				c.depth = p.nextFloat()
				c.cs, _ = p.next(tString)
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
				knob, found := knobs[name]
				if !found {
					knob = make([]float64, p.frames)
				}
				startFrame := p.nextInt()
				if startFrame < 0 || startFrame >= p.frames {
					return nil, fmt.Errorf("invalid start frame %d for knob %s", startFrame, name)
				}
				endFrame := p.nextInt()
				if endFrame < 0 || endFrame >= p.frames || endFrame < startFrame {
					return nil, fmt.Errorf("invalid end frame %d for knob %s", endFrame, name)
				}
				startValue := p.nextFloat()
				endValue := p.nextFloat()
				length := endFrame - startFrame
				delta := (endValue - startValue) / float64(length+1)
				for frame := startFrame; frame <= endFrame; frame++ {
					knob[frame] = startValue
					startValue += delta
				}
				knobs[name] = knob
				p.isAnimated = true
			case BASENAME:
				if p.basename != "" {
					fmt.Fprintln(os.Stderr, "Setting the basename multiple times")
				}
				p.basename = p.nextString()
				formatString = fmt.Sprintf("%s/%s-%%0%dd.png", FramesDirectory, p.basename, len(strconv.Itoa(p.frames)))
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
			case SETKNOBS:
				c := SetKnobsCommand{
					value: p.nextFloat(),
				}
				command = c
			case MESH:
				c := MeshCommand{
					filename: p.nextString(),
				}
				command = c
			case LIGHT:
				name := p.nextString()
				_, found := lightSources[name]
				if found {
					return nil, fmt.Errorf("light %s is already defined", name)
				}
				lightSource := LightSource{
					color:    Color{byte(p.nextInt()), byte(p.nextInt()), byte(p.nextInt())},
					location: []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()},
				}
				lightSources[name] = lightSource
			case AMBIENT:
				ambient = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
			case CONSTANTS:
				constant := make([][]float64, 4)
				name := p.nextString()
				kar, kdr, ksr, kag, kdg, ksg, kab, kdb, ksb := p.nextFloat(), p.nextFloat(), p.nextFloat(), p.nextFloat(), p.nextFloat(), p.nextFloat(), p.nextFloat(), p.nextFloat(), p.nextFloat()
				constant[0] = []float64{kar, kag, kab} // ambient
				constant[1] = []float64{kdr, kdg, kdb} // diffuse
				constant[2] = []float64{ksr, ksg, ksb} // specular
				next := p.peek().tt
				if next == tFloat || next == tInt {
					constant[3] = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				} else {
					constant[3] = []float64{0, 0, 0}
				}
				constants[name] = constant
			}
			if command != nil {
				commands = append(commands, command)
			}
			next := p.nextToken().tt
			if next != tNewline && next != tEOF {
				return nil, fmt.Errorf("unexpected %v at end of statement", next)
			}
		case tString:
			return nil, fmt.Errorf("unrecognized identifier: \"%s\"", t.value)
		}
	}
}

func (p *Parser) process(commands []Command) error {
	if p.isAnimated {
		os.RemoveAll(FramesDirectory)
		os.Mkdir(FramesDirectory, 0755)
	} else {
		p.frames = 1
	}

	var wg sync.WaitGroup
	jobs := make(chan Job, 100)
	for i := 0; i < MaxWorkers; i++ {
		wg.Add(1)
		go worker(NewDrawer(DefaultHeight, DefaultWidth), commands, jobs, &wg)
	}

	var err error
	for frame := 0; frame < p.frames; frame++ {
		jobs <- Job{
			animated: p.isAnimated,
			frame:    frame,
		}
	}

	close(jobs)
	wg.Wait()
	if p.isAnimated {
		fmt.Println("Making animation...")
		err = MakeAnimation(p.basename)
	}
	return err
}

func renderFrame(drawer *Drawer, commands []Command, frame int) error {
	var err error
	for _, command := range commands {
		switch command.(type) {
		case MoveCommand:
			c := command.(MoveCommand)
			x, y, z := c.args[0], c.args[1], c.args[2]
			if c.knob != "" {
				if knob, err := getKnob(c.knob, frame); err == nil {
					x *= knob
					y *= knob
					z *= knob
				} else {
					return err
				}
			}
			err = drawer.Move(x, y, z)
		case ScaleCommand:
			c := command.(ScaleCommand)
			x, y, z := c.args[0], c.args[1], c.args[2]
			if c.knob != "" {
				if knob, err := getKnob(c.knob, frame); err == nil {
					x *= knob
					y *= knob
					z *= knob
				} else {
					return err
				}
			}
			err = drawer.Scale(x, y, z)
		case RotateCommand:
			c := command.(RotateCommand)
			degrees := c.degrees
			if c.knob != "" {
				if knob, err := getKnob(c.knob, frame); err == nil {
					degrees *= knob
				} else {
					return err
				}
			}
			err = drawer.Rotate(c.axis, degrees)
		case LineCommand:
			c := command.(LineCommand)
			err = drawer.Line(c.p1[0], c.p1[1], c.p1[2], c.p2[0], c.p2[1], c.p2[2])
			if err != nil {
				return err
			}
			err = drawer.DrawLines(White)
		case SphereCommand:
			c := command.(SphereCommand)
			err = drawer.Sphere(c.center[0], c.center[1], c.center[2], c.radius)
			if err != nil {
				return err
			}
			if c.constants != "" {
				if constant, err := getConstants(c.constants); err == nil {
					err = drawer.DrawShadedPolygons(constant, lightSources)
				} else {
					return err
				}
			} else {
				drawer.DrawPolygons(White)
			}
		case TorusCommand:
			c := command.(TorusCommand)
			err = drawer.Torus(c.center[0], c.center[1], c.center[2], c.r1, c.r2)
			if err != nil {
				return err
			}
			if c.constants != "" {
				if constant, err := getConstants(c.constants); err == nil {
					err = drawer.DrawShadedPolygons(constant, lightSources)
				} else {
					return err
				}
			} else {
				drawer.DrawPolygons(White)
			}
		case BoxCommand:
			c := command.(BoxCommand)
			err = drawer.Box(c.p1[0], c.p1[1], c.p1[2], c.width, c.height, c.depth)
			if err != nil {
				return err
			}
			if c.constants != "" {
				if constant, err := getConstants(c.constants); err == nil {
					err = drawer.DrawShadedPolygons(constant, lightSources)
				} else {
					return err
				}
			} else {
				drawer.DrawPolygons(White)
			}
		case PopCommand:
			drawer.Pop()
		case PushCommand:
			drawer.Push()
		case SaveCommand:
			c := command.(SaveCommand)
			err = drawer.Save(c.filename)
		case DisplayCommand:
			err = drawer.Display()
		case SetCommand:
			c := command.(SetCommand)
			knobs[c.name][frame] = c.value
		case SetKnobsCommand:
			c := command.(SetKnobsCommand)
			for key := range knobs {
				knobs[key][frame] = c.value
			}
		case MeshCommand:
			c := command.(MeshCommand)
			f, err := os.Open(c.filename)
			if err != nil {
				return err
			}
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				// TODO: Legitimize
				var x, y, z float64
				num, _ := fmt.Sscanf(scanner.Text(), "vertex %f %f %f", &x, &y, &z)
				if num == 3 {
					drawer.AddPoint(x, y, z)
				}
			}
			drawer.apply()
			drawer.DrawPolygons(White)
		}
		if err != nil {
			return err
		}
	}
	return err
}

func getKnob(name string, frame int) (float64, error) {
	if knob, found := knobs[name]; found {
		return knob[frame], nil
	}
	return 0, fmt.Errorf("undefined knob '%s'", name)
}

func getConstants(name string) ([][]float64, error) {
	if constant, found := constants[name]; found {
		return constant, nil
	}
	return nil, fmt.Errorf("undefined constant '%s'", name)
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

// next returns the next token if it matches the given token types
// If the token does not match, error is non-nil
func (p *Parser) next(typs ...TokenType) (string, error) {
	next := p.peek()
	for _, tt := range typs {
		if next.tt == tt {
			p.nextToken()
			return next.value, nil
		}
	}
	return "", fmt.Errorf("expected %v, got %v", typs, next.tt)
}

// nextRequired returns the value of the nextRequired token if its type is valid
// Panics if none of the token types match
func (p *Parser) nextRequired(typs ...TokenType) string {
	next, err := p.next(typs...)
	if err != nil {
		panic(err)
	}
	return next
}

// nextInt returns the next integer token from the lexer
func (p *Parser) nextInt() int {
	v, _ := strconv.Atoi(p.nextRequired(tInt))
	return v
}

// nextFloat returns the next token from the lexer as a float.
func (p *Parser) nextFloat() float64 {
	v, _ := strconv.ParseFloat(p.nextRequired(tInt, tFloat), 64)
	return v
}

// nextString returns the next token from the lexer.
func (p *Parser) nextString() string {
	return p.nextRequired(tString)
}

// nextIdent returns the next identifier from the lexer as a string.
func (p *Parser) nextIdent() string {
	return p.nextRequired(tIdent)
}

// unread adds the token to the list of backup tokens.
// Since channels cannot be "unread", we use a list to backup these tokens
func (p *Parser) unread(token Token) {
	p.backup = append(p.backup, token)
}

// peek returns the next token without consuming it
func (p *Parser) peek() Token {
	token := p.nextToken()
	p.unread(token)
	return token
}

// Job is a struct that tells a worker thread which frames to render
type Job struct {
	frame    int  // frame to render
	animated bool // whether the frame is part of an animation
}

// worker is a worker thread that renders frames
func worker(drawer *Drawer, commands []Command, jobs chan Job, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}
			if job.animated {
				fmt.Println("Rendering frame", job.frame)
			}

			err := renderFrame(drawer, commands, job.frame)
			if job.animated {
				err = drawer.Save(fmt.Sprintf(formatString, job.frame))
				if err != nil {
					return
				}
				drawer.Reset()
			}
		}
	}
}
