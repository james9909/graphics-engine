package main

type Command interface {
	Name() string
}

type SaveCommand struct {
	filename string
}

func (c SaveCommand) Name() string {
	return "SAVE"
}

type DisplayCommand struct {
}

func (c DisplayCommand) Name() string {
	return "DISPLAY"
}

type PushCommand struct {
}

func (c PushCommand) Name() string {
	return "PUSH"
}

type PopCommand struct {
}

func (c PopCommand) Name() string {
	return "POP"
}

type TransformCommand struct {
	knob string
}

type MoveCommand struct {
	TransformCommand
	args []float64
}

func (c MoveCommand) Name() string {
	return "MOVE"
}

type ScaleCommand struct {
	TransformCommand
	args []float64
}

func (c ScaleCommand) Name() string {
	return "SCALE"
}

type RotateCommand struct {
	TransformCommand
	axis    string
	degrees float64
}

func (c RotateCommand) Name() string {
	return "ROTATE"
}

type ShapeCommand struct {
	constants string
	cs        string
}

type LineCommand struct {
	ShapeCommand
	p1  []float64
	p2  []float64
	cs2 string
}

func (c LineCommand) Name() string {
	return "LINE"
}

type SphereCommand struct {
	ShapeCommand
	center []float64
	radius float64
}

func (c SphereCommand) Name() string {
	return "SPHERE"
}

type TorusCommand struct {
	ShapeCommand
	center []float64
	r1     float64
	r2     float64
}

func (c TorusCommand) Name() string {
	return "TORUS"
}

type BoxCommand struct {
	ShapeCommand
	p1     []float64
	width  float64
	height float64
	depth  float64
}

func (c BoxCommand) Name() string {
	return "BOX"
}
