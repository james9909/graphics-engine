package main

import (
	"errors"
	"fmt"
	"image/color"
	"io/ioutil"
	"strconv"
)

// DrawingMode defines the type of each drawing mode
type DrawingMode int

const (
	// DrawLineMode is a draw argument that draws 2D lines onto the Image
	DrawLineMode DrawingMode = iota
	// DrawPolygonMode is a draw argument that draws 3D polygons onto the Image
	DrawPolygonMode
)

// Parser is a script parser
type Parser struct {
	frame    *Image    // current image
	em       *Matrix   // underlying edge/polygon matrix
	tm       *Matrix   // transformation matrix
	cs       *Stack    // relative coordinate system stack
	lexer    *Lexer    // lexer
	backup   []Token   // token backup
	commands []Command // list of commands
}

// NewParser returns a new parser
func NewParser() *Parser {
	cs := NewStack()
	cs.Push(IdentityMatrix(4)) // stack should begin with the identity matrix
	return &Parser{
		frame:  NewImage(DefaultHeight, DefaultWidth),
		em:     NewMatrix(4, 0),
		tm:     IdentityMatrix(4),
		cs:     cs,
		backup: make([]Token, 0, 10),
	}
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
	commands, err := p.parseCommands()
	if err == nil {
		p.commands = commands
		err = p.parse()
	}
	return err
}

func (p *Parser) parseCommands() ([]Command, error) {
	commands := make([]Command, 0, 10)
	for {
		t := p.next()
		switch t.tt {
		case tError:
			return nil, errors.New(t.value)
		case tEOF:
			return commands, nil
		case tIdent:
			var command Command
			switch Lookup(t.value) {
			case MOVE:
				c := MoveCommand{}
				c.args = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				if p.peek().tt == tString {
					c.knob = p.next().value
				}
				command = c
			case SCALE:
				c := ScaleCommand{}
				c.args = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				if p.peek().tt == tString {
					c.knob = p.next().value
				}
				command = c
			case ROTATE:
				c := RotateCommand{}
				c.axis = p.nextIdent()
				c.degrees = p.nextFloat()
				if p.peek().tt == tString {
					c.knob = p.next().value
				}
				command = c
			case LINE:
				c := LineCommand{}
				if p.peek().tt == tString {
					c.constants = p.next().value
				}
				c.p1 = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				if p.peek().tt == tString {
					c.cs = p.next().value
				}
				c.p2 = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				if p.peek().tt == tString {
					c.cs2 = p.next().value
				}
				command = c
			case SPHERE:
				c := SphereCommand{}
				if p.peek().tt == tString {
					c.constants = p.next().value
				}
				c.center = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				c.radius = p.nextFloat()
				if p.peek().tt == tString {
					c.cs = p.next().value
				}
				command = c
			case TORUS:
				c := TorusCommand{}
				if p.peek().tt == tString {
					c.constants = p.next().value
				}
				c.center = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				c.r1 = p.nextFloat()
				c.r2 = p.nextFloat()
				if p.peek().tt == tString {
					c.cs = p.next().value
				}
				command = c
			case BOX:
				c := BoxCommand{}
				if p.peek().tt == tString {
					c.constants = p.next().value
				}
				c.p1 = []float64{p.nextFloat(), p.nextFloat(), p.nextFloat()}
				c.width = p.nextFloat()
				c.height = p.nextFloat()
				c.depth = p.nextFloat()
				if p.peek().tt == tString {
					c.cs = p.next().value
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
			case tIllegal:
				return commands, fmt.Errorf("unrecognized identifier : \"%s\"", t)
			}
			if command != nil {
				commands = append(commands, command)
			}
			if err := p.expect(tNewline); err != nil {
				return nil, fmt.Errorf("unexpected %v at end of statement", p.peek().tt)
			}
		}
	}
	return commands, nil
}

func (p *Parser) parse() error {
	var err error
	for _, command := range p.commands {
		switch command.(type) {
		case MoveCommand:
			c := command.(MoveCommand)
			err = p.move(c.args[0], c.args[1], c.args[2])
		case ScaleCommand:
			c := command.(ScaleCommand)
			err = p.scale(c.args[0], c.args[1], c.args[2])
		case RotateCommand:
			c := command.(RotateCommand)
			err = p.rotate(c.axis, c.degrees)
		case LineCommand:
			c := command.(LineCommand)
			err = p.line(c.p1[0], c.p1[1], c.p1[2], c.p2[0], c.p2[1], c.p2[2])
		case SphereCommand:
			c := command.(SphereCommand)
			err = p.sphere(c.center[0], c.center[1], c.center[2], c.radius)
		case TorusCommand:
			c := command.(TorusCommand)
			err = p.torus(c.center[0], c.center[1], c.center[2], c.r1, c.r2)
		case BoxCommand:
			c := command.(BoxCommand)
			err = p.box(c.p1[0], c.p1[1], c.p1[2], c.width, c.height, c.depth)
		case PopCommand:
			p.cs.Pop()
		case PushCommand:
			top := p.cs.Peek()
			if top != nil {
				p.cs.Push(top.Copy())
			}
		case SaveCommand:
			c := command.(SaveCommand)
			err = p.save(c.filename)
		case DisplayCommand:
			err = p.display()
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// next returns the next token from the lexer
func (p *Parser) next() Token {
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

// nextFloat returns the next token from the lexer as a float.
// Panics if the next token is not a float or integer
func (p *Parser) nextFloat() float64 {
	if p.expect(tInt) != nil && p.expect(tFloat) != nil {
		panic(fmt.Errorf("expected %v, got %v", tFloat, p.peek().tt))
	}
	v, _ := strconv.ParseFloat(p.next().value, 64)
	return v
}

// nextString returns the next token from the lexer.
// Panics if the next token is not a string
func (p *Parser) nextString() string {
	if p.expect(tString) != nil {
		panic(fmt.Errorf("expected %v, got %v", tString, p.peek().tt))
	}
	return p.next().value
}

// nextIdent returns the next identifier from the lexer as a string.
// Panics if the next token is not an identifier
func (p *Parser) nextIdent() string {
	if p.expect(tIdent) != nil {
		panic(fmt.Errorf("expected %v, got %v", tIdent, p.peek().tt))
	}
	return p.next().value
}

// unread adds the token to the list of backup tokens.
// Since channels cannot be "unread", we use a list to backup these tokens
func (p *Parser) unread(token Token) {
	if p.backup == nil {
		p.backup = make([]Token, 0, 10)
	}
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

func (p *Parser) line(x0, y0, z0, x1, y1, z1 float64) error {
	p.em.AddEdge(x0, y0, z0, x1, y1, z1)
	err := p.apply(DrawLineMode)
	return err
}

func (p *Parser) scale(sx, sy, sz float64) error {
	dilation := MakeDilation(sx, sy, sz)

	top := p.cs.Pop()
	top, err := top.Multiply(dilation)
	if err != nil {
		return err
	}
	p.cs.Push(top)
	return nil
}

func (p *Parser) move(x, y, z float64) error {
	translation := MakeTranslation(x, y, z)
	top := p.cs.Pop()
	top, err := top.Multiply(translation)
	if err != nil {
		return err
	}
	p.cs.Push(top)
	return nil
}

func (p *Parser) rotate(axis string, theta float64) error {
	var rotation *Matrix
	switch axis {
	case "x":
		rotation = MakeRotX(theta)
	case "y":
		rotation = MakeRotY(theta)
	case "z":
		rotation = MakeRotZ(theta)
	default:
		return errors.New("axis must be \"x\", \"y\", or \"z\"")
	}

	top := p.cs.Pop()
	top, err := top.Multiply(rotation)
	if err != nil {
		return err
	}
	p.cs.Push(top)
	return nil
}

func (p *Parser) apply(mode DrawingMode) error {
	product, err := p.cs.Peek().Multiply(p.em)
	if err != nil {
		return err
	}
	p.em = product
	p.draw(mode)
	p.clear()
	return nil
}

func (p *Parser) draw(mode DrawingMode) {
	switch mode {
	case DrawLineMode:
		p.frame.DrawLines(p.em, color.White)
	case DrawPolygonMode:
		p.frame.DrawPolygons(p.em, color.White)
	default:
	}
}

func (p *Parser) save(filename string) error {
	err := p.frame.Save(filename)
	return err
}

func (p *Parser) display() error {
	err := p.frame.Display()
	return err
}

func (p *Parser) circle(cx, cy, cz, radius float64) error {
	p.em.AddCircle(cx, cy, cz, radius)
	err := p.apply(DrawLineMode)
	return err
}

func (p *Parser) hermite(x0, y0, x1, y1, dx0, dy0, dx1, dy1 float64) error {
	p.em.AddHermite(x0, y0, x1, y1, dx0, dy0, dx1, dy1)
	err := p.apply(DrawLineMode)
	return err
}

func (p *Parser) bezier(x0, y0, x1, y1, x2, y2, x3, y3 float64) error {
	p.em.AddBezier(x0, y0, x1, y1, x2, y2, x3, y3)
	err := p.apply(DrawLineMode)
	return err
}

func (p *Parser) box(x, y, z, width, height, depth float64) error {
	p.em.AddBox(x, y, z, width, height, depth)
	err := p.apply(DrawPolygonMode)
	return err
}

func (p *Parser) clear() {
	p.em = NewMatrix(4, 0)
}

func (p *Parser) sphere(cx, cy, cz, radius float64) error {
	p.em.AddSphere(cx, cy, cz, radius)
	err := p.apply(DrawPolygonMode)
	return err
}

func (p *Parser) torus(cx, cy, cz, r1, r2 float64) error {
	p.em.AddTorus(cx, cy, cz, r1, r2)
	err := p.apply(DrawPolygonMode)
	return err
}
