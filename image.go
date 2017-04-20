package main

import (
	"bytes"
	"fmt"
	"image/color"
	"os"
	"os/exec"
)

type Pixel struct {
	color color.Color
}

type Image struct {
	pixels [][]Pixel
	height int
	width  int
}

func NewImage(height, width int) *Image {
	pixels := make([][]Pixel, height)
	for i := 0; i < width; i++ {
		pixels[i] = make([]Pixel, width)
	}
	return &Image{
		pixels: pixels,
		height: height,
		width:  width,
	}
}

func (im Image) drawLine(x1, y1, x2, y2 int, c color.Color) {
	if x1 > x2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}

	A := 2 * (y2 - y1)
	B := 2 * -(x2 - x1)
	m := float32(A) / float32(-B)
	if m >= 0 {
		if m <= 1 {
			im.drawOctant1(x1, y1, x2, y2, A, B, c)
		} else {
			im.drawOctant2(x1, y1, x2, y2, A, B, c)
		}
	} else {
		if m < -1 {
			im.drawOctant7(x1, y1, x2, y2, A, B, c)
		} else {
			im.drawOctant8(x1, y1, x2, y2, A, B, c)
		}
	}
}

func (im Image) drawOctant1(x1, y1, x2, y2, A, B int, c color.Color) {
	d := A + B/2
	for x1 <= x2 {
		im.set(x1, y1, c)
		if d > 0 {
			y1++
			d += B
		}
		x1++
		d += A
	}
}

func (im Image) drawOctant2(x1, y1, x2, y2, A, B int, c color.Color) {
	d := A/2 + B
	for y1 <= y2 {
		im.set(x1, y1, c)
		if d < 0 {
			x1++
			d += A
		}
		y1++
		d += B
	}
}

func (im Image) drawOctant7(x1, y1, x2, y2, A, B int, c color.Color) {
	d := A/2 + B
	for y1 >= y2 {
		im.set(x1, y1, c)
		if d > 0 {
			x1++
			d += A
		}
		y1--
		d -= B
	}
}

func (im Image) drawOctant8(x1, y1, x2, y2, A, B int, c color.Color) {
	d := A - B/2
	for x1 <= x2 {
		im.set(x1, y1, c)
		if d < 0 {
			y1--
			d -= B
		}
		x1++
		d += A
	}
}

func (im Image) fill(c color.Color) {
	for x := 0; x < im.height; x++ {
		for y := 0; y < im.width; y++ {
			im.set(x, y, c)
		}
	}
}

func (im Image) set(x, y int, c color.Color) {
	im.pixels[y][x].color = c
}

func (im Image) save(name string) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("P3 %d %d %d\n", im.width, im.height, 255))
	for x := 0; x < im.height; x++ {
		for y := 0; y < im.width; y++ {
			pixel := im.pixels[x][y]
			r, g, b, _ := pixel.color.RGBA()
			buffer.WriteString(fmt.Sprintf("%d %d %d\n", r/256, g/256, b/256))
		}
	}
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	f.WriteString(buffer.String())
	f.Close()
}

func (im Image) display() {
	filename := "tmp.ppm"
	im.save(filename)
	args := []string{filename}
	_, err := exec.Command("display", args...).Output()
	if err != nil {
		panic(err)
	}
	os.Remove(filename)
}