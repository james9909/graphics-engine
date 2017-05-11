package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	// DefaultHeight is the default height of an Image
	DefaultHeight = 500
	// DefaultWidth is the default width of an Image
	DefaultWidth = 500
)

var (
	Black = NewColor(0, 0, 0)
	White = NewColor(255, 255, 255)
)

type Color struct {
	r byte
	b byte
	g byte
}

func NewColor(r, g, b byte) Color {
	return Color{r, g, b}
}

// Image represents an image
type Image struct {
	frame  [][]Color
	height int
	width  int
}

// NewImage returns a new Image with the given height and width
func NewImage(height, width int) *Image {
	frame := make([][]Color, height)
	for i := 0; i < height; i++ {
		frame[i] = make([]Color, width)
	}
	image := &Image{
		frame:  frame,
		height: height,
		width:  width,
	}
	image.Fill(Black)
	return image
}

// DrawLines draws all lines onto the Image
func (image *Image) DrawLines(em *Matrix, c Color) error {
	if em.cols < 2 {
		return errors.New("2 or more points are required for drawing")
	}
	for i := 0; i < em.cols-1; i += 2 {
		p0 := em.GetColumn(i)
		p1 := em.GetColumn(i + 1)
		image.DrawLine(int(p0[0]), int(p0[1]), int(p1[0]), int(p1[1]), c)
	}
	return nil
}

// DrawPolygons draws all polygons onto the Image
func (image *Image) DrawPolygons(em *Matrix, c Color) error {
	if em.cols < 3 {
		return errors.New("3 or more points are required for drawing")
	}
	for i := 0; i < em.cols-2; i += 3 {
		p0 := em.GetColumn(i)
		p1 := em.GetColumn(i + 1)
		p2 := em.GetColumn(i + 2)
		if isVisible(p0, p1, p2) {
			image.DrawLine(int(p0[0]), int(p0[1]), int(p1[0]), int(p1[1]), c)
			image.DrawLine(int(p1[0]), int(p1[1]), int(p2[0]), int(p2[1]), c)
			image.DrawLine(int(p2[0]), int(p2[1]), int(p0[0]), int(p0[1]), c)
		}
	}
	return nil
}

// DrawLine draws a single line onto the Image
func (image *Image) DrawLine(x1, y1, x2, y2 int, c Color) {
	if x1 > x2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}

	A := 2 * (y2 - y1)
	B := 2 * -(x2 - x1)
	m := float32(A) / float32(-B)
	if m >= 0 {
		if m <= 1 {
			image.drawOctant1(x1, y1, x2, y2, A, B, c)
		} else {
			image.drawOctant2(x1, y1, x2, y2, A, B, c)
		}
	} else {
		if m < -1 {
			image.drawOctant7(x1, y1, x2, y2, A, B, c)
		} else {
			image.drawOctant8(x1, y1, x2, y2, A, B, c)
		}
	}
}

func (image Image) drawOctant1(x1, y1, x2, y2, A, B int, c Color) {
	d := A + B/2
	for x1 <= x2 {
		image.set(x1, y1, c)
		if d > 0 {
			y1++
			d += B
		}
		x1++
		d += A
	}
}

func (image Image) drawOctant2(x1, y1, x2, y2, A, B int, c Color) {
	d := A/2 + B
	for y1 <= y2 {
		image.set(x1, y1, c)
		if d < 0 {
			x1++
			d += A
		}
		y1++
		d += B
	}
}

func (image Image) drawOctant7(x1, y1, x2, y2, A, B int, c Color) {
	d := A/2 + B
	for y1 >= y2 {
		image.set(x1, y1, c)
		if d > 0 {
			x1++
			d += A
		}
		y1--
		d -= B
	}
}

func (image Image) drawOctant8(x1, y1, x2, y2, A, B int, c Color) {
	d := A - B/2
	for x1 <= x2 {
		image.set(x1, y1, c)
		if d < 0 {
			y1--
			d -= B
		}
		x1++
		d += A
	}
}

// Fill completely fills the Image with a single color
func (image Image) Fill(c Color) {
	for y := 0; y < image.height; y++ {
		for x := 0; x < image.width; x++ {
			image.set(x, y, c)
		}
	}
}

func (image Image) set(x, y int, c Color) error {
	if x < 0 || x >= image.width {
		return errors.New("invalid x coordinate")
	}
	if y < 0 || y >= image.height {
		return errors.New("invalid y coordinate")
	}
	image.frame[y][x] = c
	return nil
}

// SavePpm will save the Image as a ppm
func (image Image) SavePpm(name string) error {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("P3 %d %d %d\n", image.width, image.height, 255))
	for y := 0; y < image.height; y++ {
		for x := 0; x < image.width; x++ {
			color := image.frame[image.height-y-1][x]
			buffer.WriteString(fmt.Sprintf("%d %d %d\n", color.r, color.b, color.g))
		}
	}
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(buffer.String())
	return err
}

// Save will save an Image into a given format
func (image Image) Save(name string) error {
	index := strings.Index(name, ".")
	extension := ".ppm"
	if index == -1 {
		extension = ".png"
	} else {
		extension = name[index:]
		name = name[:index]
	}

	if extension == ".ppm" {
		// save as ppm without converting
		err := image.SavePpm(fmt.Sprint(name, ".ppm"))
		return err
	}

	ppm := fmt.Sprintf("%s-tmp.ppm", name)
	err := image.SavePpm(ppm)
	if err != nil {
		return err
	}
	defer os.Remove(ppm)
	args := []string{ppm, fmt.Sprint(name, extension)}
	err = exec.Command("convert", args...).Run()
	return err
}

// Display displays the Image
func (image Image) Display() error {
	filename := "tmp.ppm"
	err := image.SavePpm(filename)
	if err != nil {
		return err
	}
	defer os.Remove(filename)

	args := []string{filename}
	err = exec.Command("display", args...).Run()
	return err
}

// MakeAnimation converts individual frames to a gif
func MakeAnimation(basename string) error {
	path := fmt.Sprintf("%s/%s*", FramesDirectory, basename)
	gif := fmt.Sprintf("%s.gif", basename)
	args := []string{"-delay", "3", path, gif}
	err := exec.Command("convert", args...).Run()
	return err
}

func isVisible(p0, p1, p2 []float64) bool {
	a := []float64{p1[0] - p0[0], p1[1] - p0[1], p1[2] - p0[2]}
	b := []float64{p2[0] - p0[0], p2[1] - p0[1], p2[2] - p0[2]}
	normal := CrossProduct(a, b)
	return normal[2] > 0
}
