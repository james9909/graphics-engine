package main

import (
	"errors"
	"fmt"
)

// DrawMode defines the type of each drawing mode
type DrawMode int

const (
	DrawLineMode    DrawMode = iota // DrawLineMode is a draw argument that draws 2D lines onto the Image
	DrawPolygonMode                 // DrawPolygonMode is a draw argument that draws 3D polygons onto the Image
)

// Drawer is a struct that draws on an image
type Drawer struct {
	frame *Image  // underlying image
	em    *Matrix // edge/polygon matrix
	cs    *Stack  // coordinate system stack
}

func NewDrawer(height, width int) *Drawer {
	return &Drawer{
		frame: NewImage(height, width),
		em:    NewMatrix(4, 0),
		cs:    NewStack(),
	}
}

func (d *Drawer) apply(mode DrawMode) error {
	product, err := d.cs.Peek().Multiply(d.em)
	if err != nil {
		return err
	}
	d.em = product
	if err := d.draw(mode); err != nil {
		return err
	}
	d.clear()
	return nil
}

func (d *Drawer) draw(mode DrawMode) error {
	var err error
	switch mode {
	case DrawLineMode:
		err = d.frame.DrawLines(d.em, White)
	case DrawPolygonMode:
		err = d.frame.DrawPolygons(d.em, White)
	default:
		err = fmt.Errorf("undefined draw mode %s", mode)
	}
	return err
}

func (d *Drawer) clear() {
	d.em = NewMatrix(4, 0)
}

// Reset clears the image and edge matrix
func (d *Drawer) Reset() {
	d.clear()
	d.cs = NewStack()
	d.frame = NewImage(d.frame.height, d.frame.width)
}

func (d *Drawer) Line(x0, y0, z0, x1, y1, z1 float64) error {
	d.em.AddEdge(x0, y0, z0, x1, y1, z1)
	err := d.apply(DrawLineMode)
	return err
}

func (d *Drawer) Scale(sx, sy, sz float64) error {
	dilation := MakeDilation(sx, sy, sz)

	top := d.cs.Pop()
	top, err := top.Multiply(dilation)
	if err != nil {
		return err
	}
	d.cs.Push(top)
	return nil
}

func (d *Drawer) Move(x, y, z float64) error {
	translation := MakeTranslation(x, y, z)
	top := d.cs.Pop()
	top, err := top.Multiply(translation)
	if err != nil {
		return err
	}
	d.cs.Push(top)
	return nil
}

func (d *Drawer) Rotate(axis string, theta float64) error {
	theta = degreesToRadians(theta)
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

	top := d.cs.Pop()
	top, err := top.Multiply(rotation)
	if err != nil {
		return err
	}
	d.cs.Push(top)
	return nil
}

func (d *Drawer) Save(filename string) error {
	err := d.frame.Save(filename)
	return err
}

func (d *Drawer) Display() error {
	err := d.frame.Display()
	return err
}

func (d *Drawer) Circle(cx, cy, cz, radius float64) error {
	d.em.AddCircle(cx, cy, cz, radius)
	err := d.apply(DrawLineMode)
	return err
}

func (d *Drawer) Hermite(x0, y0, x1, y1, dx0, dy0, dx1, dy1 float64) error {
	d.em.AddHermite(x0, y0, x1, y1, dx0, dy0, dx1, dy1)
	err := d.apply(DrawLineMode)
	return err
}

func (d *Drawer) Bezier(x0, y0, x1, y1, x2, y2, x3, y3 float64) error {
	d.em.AddBezier(x0, y0, x1, y1, x2, y2, x3, y3)
	err := d.apply(DrawLineMode)
	return err
}

func (d *Drawer) Box(x, y, z, width, height, depth float64) error {
	d.em.AddBox(x, y, z, width, height, depth)
	err := d.apply(DrawPolygonMode)
	return err
}

func (d *Drawer) Sphere(cx, cy, cz, radius float64) error {
	d.em.AddSphere(cx, cy, cz, radius)
	err := d.apply(DrawPolygonMode)
	return err
}

func (d *Drawer) Torus(cx, cy, cz, r1, r2 float64) error {
	d.em.AddTorus(cx, cy, cz, r1, r2)
	err := d.apply(DrawPolygonMode)
	return err
}

func (d *Drawer) Pop() {
	d.cs.Pop()
}

func (d *Drawer) Push() {
	var new *Matrix
	if d.cs.IsEmpty() {
		new = IdentityMatrix()
	} else {
		new = d.cs.Peek().Copy()
	}
	d.cs.Push(new)
}

func (d *Drawer) AddPoint(x, y, z float64) {
	d.em.AddPoint(x, y, z)
}
