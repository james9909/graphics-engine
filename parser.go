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
	frame *Frame
	em    *Matrix
	tm    *Matrix
}

func NewParser() *Parser {
	return &Parser{
		frame: NewFrame(DefaultHeight, DefaultWidth),
		em:    NewEdgeMatrix(),
		tm:    IdentityMatrix(4),
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
			scanner.Scan()
			argv := strings.Split(scanner.Text(), " ")
			err = p.line(argv)
		case "ident":
			p.tm = IdentityMatrix(4)
		case "scale":
			scanner.Scan()
			argv := strings.Split(scanner.Text(), " ")
			err = p.scale(argv)
		case "move":
			scanner.Scan()
			argv := strings.Split(scanner.Text(), " ")
			err = p.move(argv)
		case "rotate":
			scanner.Scan()
			argv := strings.Split(scanner.Text(), " ")
			err = p.rotate(argv)
		case "apply":
			err = p.apply()
		case "save":
			scanner.Scan()
			argv := strings.Split(scanner.Text(), " ")
			err = p.save(argv)
		case "display":
			err = p.display()
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

func (p *Parser) line(argv []string) error {
	if len(argv) != 6 {
		return fmt.Errorf("\"line\" expects %d arguments (%d provided)", 6, len(argv))
	}
	values := make([]int, 6)
	for i, v := range argv {
		v, err := strconv.Atoi(v)
		if err != nil {
			return errors.New("arguments must be numeric")
		}
		values[i] = v
	}
	p.em.AddEdge(values[0], values[1], values[2], values[3], values[4], values[5])

	return nil
}

func (p *Parser) scale(argv []string) error {
	if len(argv) != 3 {
		return fmt.Errorf("\"scale\" expects %d arguments (%d provided)", 3, len(argv))
	}
	values := make([]float64, 3)
	for i, v := range argv {
		v, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return errors.New("arguments must be numeric")
		}
		values[i] = v
	}
	dilation := Dilation(values[0], values[1], values[2])
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
	values := make([]float64, 3)
	for i, v := range argv {
		v, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return errors.New("arguments must be numeric")
		}
		values[i] = v
	}
	translation := Translation(values[0], values[1], values[2])
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
		rotation = RotationX(theta)
	case "y":
		rotation = RotationY(theta)
	case "z":
		rotation = RotationZ(theta)
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
	p.frame.DrawLines(p.em, color.White)
	if len(argv) != 1 {
		return fmt.Errorf("\"save\" expects %d argument (%d provided)", 1, len(argv))
	}
	err := p.frame.Save(argv[0])
	if err != nil {
		return err
	}
	return nil
}

func (p *Parser) display() error {
	p.frame.DrawLines(p.em, color.White)
	err := p.frame.Display()
	if err != nil {
		return err
	}
	return nil
}
