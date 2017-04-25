package main

import (
	"errors"
	"fmt"
	"image/color"
	"os"
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
	frame  *Image
	em     *Matrix
	tm     *Matrix
	cs     *Stack
	lexer  *Lexer
	backup []Token
}

// NewParser returns a new parser
func NewParser() *Parser {
	cs := NewStack()
	cs.Push(IdentityMatrix(4))
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
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	p.lexer = NewLexer()
	p.lexer.Lex(f)
	for {
		t := p.next()
		switch t.tt {
		case tError:
			return errors.New(t.value)
		case tEOF:
			return nil
		case tIdent:
			err := p.parseIdent(t)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unexpected token %v", t)
		}
	}
}

func (p *Parser) parseIdent(t Token) error {
	var err error
	switch t.value {
	case "line":
		x0 := p.nextFloat()
		y0 := p.nextFloat()
		z0 := p.nextFloat()
		x1 := p.nextFloat()
		y1 := p.nextFloat()
		z1 := p.nextFloat()
		err = p.line(x0, y0, z0, x1, y1, z1)
	case "ident":
		p.tm = IdentityMatrix(4)
	case "scale":
		sx := p.nextFloat()
		sy := p.nextFloat()
		sz := p.nextFloat()
		err = p.scale(sx, sy, sz)
	case "move":
		x := p.nextFloat()
		y := p.nextFloat()
		z := p.nextFloat()
		err = p.move(x, y, z)
	case "rotate":
		axis := p.nextString()
		theta := p.nextFloat()
		err = p.rotate(axis, theta)
	case "save":
		filename := p.nextString()
		err = p.save(filename)
	case "display":
		err = p.display()
	case "circle":
		cx := p.nextFloat()
		cy := p.nextFloat()
		cz := p.nextFloat()
		radius := p.nextFloat()
		err = p.circle(cx, cy, cz, radius)
	case "hermite":
		x0 := p.nextFloat()
		y0 := p.nextFloat()
		x1 := p.nextFloat()
		y1 := p.nextFloat()
		dx0 := p.nextFloat()
		dy0 := p.nextFloat()
		dx1 := p.nextFloat()
		dy1 := p.nextFloat()
		err = p.hermite(x0, y0, x1, y1, dx0, dy0, dx1, dy1)
	case "bezier":
		x0 := p.nextFloat()
		y0 := p.nextFloat()
		x1 := p.nextFloat()
		y1 := p.nextFloat()
		x2 := p.nextFloat()
		y2 := p.nextFloat()
		x3 := p.nextFloat()
		y3 := p.nextFloat()
		err = p.bezier(x0, y0, x1, y1, x2, y2, x3, y3)
	case "box":
		x := p.nextFloat()
		y := p.nextFloat()
		z := p.nextFloat()
		width := p.nextFloat()
		height := p.nextFloat()
		depth := p.nextFloat()
		err = p.box(x, y, z, width, height, depth)
	case "clear":
		p.clear()
	case "sphere":
		cx := p.nextFloat()
		cy := p.nextFloat()
		cz := p.nextFloat()
		radius := p.nextFloat()
		err = p.sphere(cx, cy, cz, radius)
	case "torus":
		cx := p.nextFloat()
		cy := p.nextFloat()
		cz := p.nextFloat()
		r1 := p.nextFloat()
		r2 := p.nextFloat()
		err = p.torus(cx, cy, cz, r1, r2)
	case "push":
		top := p.cs.Peek()
		if top != nil {
			p.cs.Push(top.Copy())
		}
	case "pop":
		p.cs.Pop()
	default:
		err = fmt.Errorf("unrecognized identifier: \"%s\"", t.value)
	}
	return err
}

func (p *Parser) next() Token {
	lenBackup := len(p.backup)
	if lenBackup > 0 {
		token := p.backup[lenBackup-1]
		p.backup = p.backup[:lenBackup-1]
		return token
	}
	token := <-p.lexer.out
	for token.tt == tComment {
		token = <-p.lexer.out
	}
	return token
}

func (p *Parser) nextFloat() float64 {
	if p.requireNext(tInt) != nil && p.requireNext(tFloat) != nil {
		panic(fmt.Errorf("expected %v, got %v", tFloat, p.peek().tt))
	}
	v, _ := strconv.ParseFloat(p.next().value, 64)
	return v
}

func (p *Parser) nextString() string {
	if p.requireNext(tString) != nil {
		panic(fmt.Errorf("expected %v, got %v", tString, p.peek().tt))
	}
	return p.next().value
}

func (p *Parser) unread(token Token) {
	if p.backup == nil {
		p.backup = make([]Token, 0, 10)
	}
	p.backup = append(p.backup, token)
}

func (p *Parser) peek() Token {
	token := p.next()
	p.unread(token)
	return token
}

func (p *Parser) requireNext(tt TokenType) error {
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
