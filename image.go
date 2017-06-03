package main

import (
	"bytes"
	"errors"
	"fmt"
	"math"
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
	Black = Color{0, 0, 0}
	White = Color{255, 255, 255}
)

type Color struct {
	r byte
	g byte
	b byte
}

// Image represents an image
type Image struct {
	frame   [][]Color
	zBuffer [][]float64
	height  int
	width   int
}

// NewImage returns a new Image with the given height and width
func NewImage(height, width int) *Image {
	frame := make([][]Color, height)
	zBuffer := make([][]float64, height)
	for i := 0; i < height; i++ {
		frame[i] = make([]Color, width)
		zBuffer[i] = make([]float64, width)
		for j := 0; j < width; j++ {
			zBuffer[i][j] = math.Inf(-1)
		}
	}
	image := &Image{
		frame:   frame,
		zBuffer: zBuffer,
		height:  height,
		width:   width,
	}
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
		image.DrawLine(int(p0[0]), int(p0[1]), p0[2], int(p1[0]), int(p1[1]), p1[2], c)
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
			image.DrawLine(int(p0[0]), int(p0[1]), p0[2], int(p1[0]), int(p1[1]), p1[2], c)
			image.DrawLine(int(p1[0]), int(p1[1]), p1[2], int(p2[0]), int(p2[1]), p2[2], c)
			image.DrawLine(int(p2[0]), int(p2[1]), p2[2], int(p0[0]), int(p0[1]), p0[2], c)
			image.Scanline(p0, p1, p2, Color{255, 0, 0})
		}
	}
	return nil
}

// DrawLine draws a single line onto the Image
func (image *Image) DrawLine(x0 int, y0 int, z0 float64, x1 int, y1 int, z1 float64, c Color) {
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
		z0, z1 = z1, z0
	}

	A := 2 * float64(y1-y0)
	B := 2 * -float64(x1-x0)
	m := A / -B
	if m >= 0 {
		if m <= 1 {
			// Draw octants 1 and 5
			d := A + B/2
			dz := (z1 - z0) / float64(x1-x0)
			for x0 <= x1 {
				image.set(x0, y0, z0, c)
				if d > 0 {
					y0++
					d += B
				}
				x0++
				d += A
				z0 += dz
			}
		} else {
			// Draw octants 2 and 6
			d := A/2 + B
			dz := (z1 - z0) / float64(y1-y0)
			for y0 <= y1 {
				image.set(x0, y0, z0, c)
				if d < 0 {
					x0++
					d += A
				}
				y0++
				d += B
				z0 += dz
			}
		}
	} else {
		if m < -1 {
			// Draw octants 3 and 7
			d := A/2 - B
			dz := (z1 - z0) / float64(y1-y0)
			for y0 >= y1 {
				image.set(x0, y0, z0, c)
				if d > 0 {
					x0++
					d += A
				}
				y0--
				d -= B
				z0 += dz
			}
		} else {
			d := A - B/2
			dz := (z1 - z0) / float64(x1-x0)
			for x0 <= x1 {
				image.set(x0, y0, z0, c)
				if d < 0 {
					y0--
					d -= B
				}
				x0++
				d += A
				z0 += dz
			}
		}
	}
}

// Fill completely fills the Image with a single color
func (image *Image) Fill(c Color) {
	for y := 0; y < image.height; y++ {
		for x := 0; x < image.width; x++ {
			image.frame[y][x] = c
		}
	}
}

func (image *Image) set(x, y int, z float64, c Color) {
	if (x < 0 || x >= image.width) || (y < 0 || y >= image.height) {
		return
	}
	if z > image.zBuffer[y][x] {
		// Plot so that the y coodinate is the row, and the x coordinate is the column
		image.frame[y][x] = c

		// Update Z buffer
		image.zBuffer[y][x] = z
	}
}

// SavePpm will save the Image as a ppm
func (image *Image) SavePpm(name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintln("P6", image.width, image.height, 255))
	for y := 0; y < image.height; y++ {
		// Adjust y coordinate that the origin is the bottom left
		adjustedY := image.height - y - 1
		for x := 0; x < image.width; x++ {
			color := image.frame[adjustedY][x]
			buffer.Write([]byte{color.r, color.g, color.b})
		}
	}

	_, err = buffer.WriteTo(f)
	return err
}

// Save will save an Image into a given format
func (image *Image) Save(name string) error {
	index := strings.Index(name, ".")
	extension := ".png"
	if index != -1 {
		extension = name[index:]
		name = name[:index]
	}

	if extension == ".ppm" {
		// save as ppm without converting
		err := image.SavePpm(fmt.Sprint(name, ".ppm"))
		return err
	}

	ppm := fmt.Sprint(name, "-tmp.ppm")
	err := image.SavePpm(ppm)
	if err != nil {
		return err
	}
	defer os.Remove(ppm)
	err = exec.Command("convert", ppm, fmt.Sprint(name, extension)).Run()
	return err
}

// Display displays the Image
func (image *Image) Display() error {
	filename := "tmp.ppm"
	err := image.SavePpm(filename)
	if err != nil {
		return err
	}
	defer os.Remove(filename)

	err = exec.Command("display", filename).Run()
	return err
}

// MakeAnimation converts individual frames to a gif
func MakeAnimation(basename string) error {
	path := fmt.Sprintf("%s/%s*", FramesDirectory, basename)
	gif := fmt.Sprintf("%s.gif", basename)
	err := exec.Command("convert", "-delay", "3", path, gif).Run()
	return err
}

func isVisible(p0, p1, p2 []float64) bool {
	a := []float64{p1[0] - p0[0], p1[1] - p0[1], p1[2] - p0[2]}
	b := []float64{p2[0] - p0[0], p2[1] - p0[1], p2[2] - p0[2]}
	normal := CrossProduct(a, b)
	return normal[2] > 0
}

func (image *Image) Scanline(p0, p1, p2 []float64, c Color) {
	// Re-order points so that p0 is the lowest and p2 is the highest
	if p0[1] > p1[1] {
		p0, p1 = p1, p0
	}
	if p0[1] > p2[1] {
		p0, p2 = p2, p0
	}
	if p1[1] > p2[1] {
		p1, p2 = p2, p1
	}
	if p0[1] == p1[1] {
		if p1[0] < p0[1] {
			p0, p1 = p1, p0
		}
	}
	if p1[1] == p2[1] {
		if p2[0] < p1[0] {
			p1, p2 = p2, p1
		}
	}

	x0 := p0[0]
	x1 := x0
	dx0 := (p2[0] - p0[0]) / float64(int(p2[1])-int(p0[1]))
	dx1 := (p1[0] - p0[0]) / float64(int(p1[1])-int(p0[1]))

	y := int(p0[1])

	z0 := p0[2]
	z1 := p0[2]
	dz0 := (p2[2] - p0[2]) / (p2[1] - p0[1])
	var dz1 float64
	if p0[1] != p1[1] {
		dz1 = (p1[2] - p0[2]) / (p1[1] - p0[1])
	} else {
		dz1 = (p2[2] - p1[2]) / (p2[1] - p1[1])
	}
	// Fill bottom half of polygon
	for y < int(p1[1]) {
		x0 += dx0
		x1 += dx1
		y++
		z0 += dz0
		z1 += dz1
		image.DrawLine(int(x0), y, z0, int(x1), y, z1, c)
	}

	x1 = p1[0]
	z1 = p1[2]
	dx1 = (p2[0] - p1[0]) / float64(int(p2[1])-int(p1[1]))
	// Fill top half of polygon
	for y < int(p2[1]) {
		x0 += dx0
		x1 += dx1
		y++
		z0 += dz0
		z1 += dz1
		image.DrawLine(int(x0), y, z0, int(x1), y, z1, c)
	}
}
