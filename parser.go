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

type Parser struct {
	frame *Image
	em    *Matrix
	tm    *Matrix
	mode  DrawingMode
}

func NewParser() *Parser {
	return &Parser{
		frame: NewImage(DefaultHeight, DefaultWidth),
		em:    NewMatrix(4, 0),
		tm:    IdentityMatrix(4),
		mode:  DrawPolygonMode,
	}
}

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
		if command[0:1] == "#" {
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
		case "apply":
			err = p.apply()
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
	argv := strings.Split(scanner.Text(), " ")
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

	return nil
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
	product, err := dilation.Multiply(p.tm)
	if err != nil {
		return err
	}
	p.tm = product

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
	product, err := translation.Multiply(p.tm)
	if err != nil {
		return err
	}
	p.tm = product

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

	product, err := rotation.Multiply(p.tm)
	if err != nil {
		return err
	}
	p.tm = product
	return nil
}

func (p *Parser) apply() error {
	product, err := p.tm.Multiply(p.em)
	if err != nil {
		return err
	}
	p.em = product
	return nil
}

func (p *Parser) save(argv []string) error {
	switch p.mode {
	case DrawLineMode:
		p.frame.DrawLines(p.em, color.White)
		break
	case DrawPolygonMode:
		p.frame.DrawPolygons(p.em, color.White)
		break
	default:
		return errors.New("invalid draw mode")
	}
	if len(argv) != 1 {
		return fmt.Errorf("\"save\" expects %d argument (%d provided)", 1, len(argv))
	}
	err := p.frame.Save(argv[0])
	return err
}

func (p *Parser) display() error {
	p.frame.Fill(color.Black)
	p.frame.DrawPolygons(p.em, color.White)
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
	return nil
}

func (p *Parser) hermite(argv []string) error {
	if len(argv) != 8 {
		return fmt.Errorf("\"hermite\" expects %d arguments (%d provided)", 8, len(argv))
	}
	values, err := getNumerical(argv)
	if err != nil {
		return err
	}
	err = p.em.AddHermite(values[0], values[1], values[2], values[3], values[4], values[5], values[6], values[7])
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
	err = p.em.AddBezier(values[0], values[1], values[2], values[3], values[4], values[5], values[6], values[7])
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
	return nil
}

func (p *Parser) clear() error {
	p.em = NewMatrix(4, 0)
	return nil
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
	return nil
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
	return nil
}
