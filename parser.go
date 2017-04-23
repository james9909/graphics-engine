package main

import (
	"bufio"
	"errors"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"
)

// DrawingMode defines the type of each drawing mode
type DrawingMode int

const (
	// DrawLineMode is a draw argument that draws 2D lines onto the Image
	DrawLineMode = 0
	// DrawPolygonMode is a draw argument that draws 3D polygons onto the Image
	DrawPolygonMode = 1
)

// Parser is a script parser
type Parser struct {
	frame *Image
	em    *Matrix
	tm    *Matrix
	cs    *Stack
}

// NewParser returns a new parser
func NewParser() *Parser {
	cs := NewStack()
	cs.Push(IdentityMatrix(4))
	return &Parser{
		frame: NewImage(DefaultHeight, DefaultWidth),
		em:    NewMatrix(4, 0),
		tm:    IdentityMatrix(4),
		cs:    cs,
	}
}

// ParseFile parses a file for commands and executes them
func (p *Parser) ParseFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		command := strings.TrimSpace(scanner.Text())
		if len(command) == 0 {
			continue
		}
		if command[0] == '#' {
			continue
		}
		switch command {
		case "line":
			argv := getArguments(scanner)
			err = p.line(argv)
		case "ident":
			p.tm = IdentityMatrix(4)
		case "scale":
			argv := getArguments(scanner)
			err = p.scale(argv)
		case "move":
			argv := getArguments(scanner)
			err = p.move(argv)
		case "rotate":
			argv := getArguments(scanner)
			err = p.rotate(argv)
		case "save":
			argv := getArguments(scanner)
			err = p.save(argv)
		case "display":
			err = p.display()
		case "circle":
			argv := getArguments(scanner)
			err = p.circle(argv)
		case "hermite":
			argv := getArguments(scanner)
			err = p.hermite(argv)
		case "bezier":
			argv := getArguments(scanner)
			err = p.bezier(argv)
		case "box":
			argv := getArguments(scanner)
			err = p.box(argv)
		case "clear":
			p.clear()
		case "sphere":
			argv := getArguments(scanner)
			err = p.sphere(argv)
		case "torus":
			argv := getArguments(scanner)
			err = p.torus(argv)
		case "push":
			top := p.cs.Peek()
			if top != nil {
				p.cs.Push(top.Copy())
			}
		case "pop":
			p.cs.Pop()
		default:
			err = fmt.Errorf("unrecognized command: \"%s\"", command)
		}

		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func getArguments(scanner *bufio.Scanner) []string {
	scanner.Scan()
	line := scanner.Text()
	line = strings.TrimSpace(line)
	argv := strings.Split(line, " ")
	return argv
}

func getNumerical(argv []string) ([]float64, error) {
	values := make([]float64, len(argv))
	for i, v := range argv {
		v, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, errors.New("arguments must be numeric")
		}
		values[i] = v
	}
	return values, nil
}

func (p *Parser) line(argv []string) error {
	if len(argv) != 6 {
		return fmt.Errorf("\"line\" expects %d arguments (%d provided)", 6, len(argv))
	}
	values, err := getNumerical(argv)
	if err != nil {
		return err
	}

	p.em.AddEdge(values[0], values[1], values[2], values[3], values[4], values[5])
	err = p.apply(DrawLineMode)

	return err
}

func (p *Parser) scale(argv []string) error {
	if len(argv) != 3 {
		return fmt.Errorf("\"scale\" expects %d arguments (%d provided)", 3, len(argv))
	}
	values, err := getNumerical(argv)
	if err != nil {
		return err
	}

	dilation := MakeDilation(values[0], values[1], values[2])

	top := p.cs.Pop()
	top, err = dilation.Multiply(top)
	if err != nil {
		return err
	}
	p.cs.Push(top)

	return nil
}

func (p *Parser) move(argv []string) error {
	if len(argv) != 3 {
		return fmt.Errorf("\"move\" expects %d arguments (%d provided)", 3, len(argv))
	}
	values, err := getNumerical(argv)
	if err != nil {
		return err
	}

	translation := MakeTranslation(values[0], values[1], values[2])
	top := p.cs.Pop()
	top, err = translation.Multiply(top)
	if err != nil {
		return err
	}
	p.cs.Push(top)

	return nil
}

func (p *Parser) rotate(argv []string) error {
	if len(argv) != 2 {
		return fmt.Errorf("\"rotate\" expects %d arguments (%d provided)", 2, len(argv))
	}
	axis := strings.ToLower(argv[0])
	theta, err := strconv.ParseFloat(argv[1], 64)
	if err != nil {
		return errors.New("arguments must be numeric")
	}

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
	top, err = rotation.Multiply(top)
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

func (p *Parser) save(argv []string) error {
	if len(argv) != 1 {
		return fmt.Errorf("\"save\" expects %d argument (%d provided)", 1, len(argv))
	}
	err := p.frame.Save(argv[0])
	return err
}

func (p *Parser) display() error {
	err := p.frame.Display()
	return err
}

func (p *Parser) circle(argv []string) error {
	if len(argv) != 4 {
		return fmt.Errorf("\"circle\" expects %d arguments (%d provided)", 4, len(argv))
	}
	values, err := getNumerical(argv)
	if err != nil {
		return err
	}

	p.em.AddCircle(values[0], values[1], values[2], values[3])
	err = p.apply(DrawLineMode)

	return err
}

func (p *Parser) hermite(argv []string) error {
	if len(argv) != 8 {
		return fmt.Errorf("\"hermite\" expects %d arguments (%d provided)", 8, len(argv))
	}
	values, err := getNumerical(argv)
	if err != nil {
		return err
	}

	p.em.AddHermite(values[0], values[1], values[2], values[3], values[4], values[5], values[6], values[7])
	err = p.apply(DrawLineMode)

	return err
}

func (p *Parser) bezier(argv []string) error {
	if len(argv) != 8 {
		return fmt.Errorf("\"bezier\" expects %d arguments (%d provided)", 8, len(argv))
	}
	values, err := getNumerical(argv)
	if err != nil {
		return err
	}

	p.em.AddBezier(values[0], values[1], values[2], values[3], values[4], values[5], values[6], values[7])
	err = p.apply(DrawLineMode)

	return err
}

func (p *Parser) box(argv []string) error {
	if len(argv) != 6 {
		return fmt.Errorf("\"box\" expects %d arguments (%d provided)", 6, len(argv))
	}
	values, err := getNumerical(argv)
	if err != nil {
		return err
	}

	p.em.AddBox(values[0], values[1], values[2], values[3], values[4], values[5])
	err = p.apply(DrawPolygonMode)

	return err
}

func (p *Parser) clear() {
	p.em = NewMatrix(4, 0)
}

func (p *Parser) sphere(argv []string) error {
	if len(argv) != 4 {
		return fmt.Errorf("\"box\" expects %d arguments (%d provided)", 4, len(argv))
	}
	values, err := getNumerical(argv)
	if err != nil {
		return err
	}

	p.em.AddSphere(values[0], values[1], values[2], values[3])
	err = p.apply(DrawPolygonMode)

	return err
}

func (p *Parser) torus(argv []string) error {
	if len(argv) != 5 {
		return fmt.Errorf("\"box\" expects %d arguments (%d provided)", 5, len(argv))
	}
	values, err := getNumerical(argv)
	if err != nil {
		return err
	}

	p.em.AddTorus(values[0], values[1], values[2], values[3], values[4])
	err = p.apply(DrawPolygonMode)

	return err
}
