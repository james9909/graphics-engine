package main

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"strings"
)

const (
	DefaultHeight = 500
	DefaultWidth  = 500
)

type Pixel struct {
	color color.Color
}

type Frame struct {
	pixels [][]Pixel
	height int
	width  int
}

func NewFrame(height, width int) *Frame {
	pixels := make([][]Pixel, height)
	for i := 0; i < height; i++ {
		pixels[i] = make([]Pixel, width)
	}
	frame := &Frame{
		pixels: pixels,
		height: height,
		width:  width,
	}
	frame.Fill(color.Black)
	return frame
}

func (frame *Frame) DrawLines(em *Matrix, c color.Color) {
	m := em.GetMatrix()
	for i := 0; i < em.cols-1; i += 2 {
		x0, y0 := m[0][i], m[1][i]
		x1, y1 := m[0][i+1], m[1][i+1]
		frame.DrawLine(int(x0), int(y0), int(x1), int(y1), c)
	}
}

func (frame *Frame) DrawLine(x1, y1, x2, y2 int, c color.Color) {
	if x1 > x2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}

	A := 2 * (y2 - y1)
	B := 2 * -(x2 - x1)
	m := float32(A) / float32(-B)
	if m >= 0 {
		if m <= 1 {
			frame.drawOctant1(x1, y1, x2, y2, A, B, c)
		} else {
			frame.drawOctant2(x1, y1, x2, y2, A, B, c)
		}
	} else {
		if m < -1 {
			frame.drawOctant7(x1, y1, x2, y2, A, B, c)
		} else {
			frame.drawOctant8(x1, y1, x2, y2, A, B, c)
		}
	}
}

func (frame Frame) drawOctant1(x1, y1, x2, y2, A, B int, c color.Color) {
	d := A + B/2
	for x1 <= x2 {
		frame.set(x1, y1, c)
		if d > 0 {
			y1++
			d += B
		}
		x1++
		d += A
	}
}

func (frame Frame) drawOctant2(x1, y1, x2, y2, A, B int, c color.Color) {
	d := A/2 + B
	for y1 <= y2 {
		frame.set(x1, y1, c)
		if d < 0 {
			x1++
			d += A
		}
		y1++
		d += B
	}
}

func (frame Frame) drawOctant7(x1, y1, x2, y2, A, B int, c color.Color) {
	d := A/2 + B
	for y1 >= y2 {
		frame.set(x1, y1, c)
		if d > 0 {
			x1++
			d += A
		}
		y1--
		d -= B
	}
}

func (frame Frame) drawOctant8(x1, y1, x2, y2, A, B int, c color.Color) {
	d := A - B/2
	for x1 <= x2 {
		frame.set(x1, y1, c)
		if d < 0 {
			y1--
			d -= B
		}
		x1++
		d += A
	}
}

func (frame Frame) Fill(c color.Color) {
	for y := 0; y < frame.height; y++ {
		for x := 0; x < frame.width; x++ {
			frame.set(x, y, c)
		}
	}
}

func (frame Frame) set(x, y int, c color.Color) error {
	if x < 0 || x >= frame.width {
		return errors.New("invalid x coordinate")
	}
	if y < 0 || y >= frame.height {
		return errors.New("invalid y coordinate")
	}
	frame.pixels[y][x].color = c
	return nil
}

func (frame Frame) SavePpm(name string) error {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("P3 %d %d %d\n", frame.width, frame.height, 255))
	for y := 0; y < frame.height; y++ {
		for x := 0; x < frame.width; x++ {
			pixel := frame.pixels[y][x]
			r, g, b, _ := pixel.color.RGBA()
			buffer.WriteString(fmt.Sprintf("%d %d %d\n", r/256, g/256, b/256))
		}
	}
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	f.WriteString(buffer.String())
	f.Close()
	return err
}

func (frame Frame) Save(name string) error {
	index := strings.Index(name, ".")
	if index == -1 {
		return errors.New("no extension provided")
	}
	base := name[:index]
	ppm := base + ".ppm"
	err := frame.SavePpm(ppm)
	if err != nil {
		return err
	}
	args := []string{ppm, name}
	_, err = exec.Command("convert", args...).Output()
	if err != nil {
		return err
	}

	err = os.Remove(ppm)
	if err != nil {
		return err
	}
	return nil
}

func (frame Frame) Display() error {
	filename := "tmp.ppm"
	frame.SavePpm(filename)
	args := []string{filename}
	_, err := exec.Command("display", args...).Output()
	if err != nil {
		return err
	}

	err = os.Remove(filename)
	if err != nil {
		return err
	}
	return nil
}
