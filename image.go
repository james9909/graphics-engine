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

type Image struct {
	frame  [][]color.Color
	height int
	width  int
}

func NewImage(height, width int) *Image {
	frame := make([][]color.Color, height)
	for i := 0; i < height; i++ {
		frame[i] = make([]color.Color, width)
	}
	image := &Image{
		frame:  frame,
		height: height,
		width:  width,
	}
	image.Fill(color.Black)
	return image
}

func (image *Image) DrawLines(em *Matrix, c color.Color) {
	if em.cols < 2 {
		fmt.Println("2 or more points are required for drawing")
		return
	}
	for i := 0; i < em.cols-1; i += 2 {
		p0 := em.GetColumn(i)
		p1 := em.GetColumn(i + 1)
		image.DrawLine(int(p0[0]), int(p0[1]), int(p1[0]), int(p1[1]), c)
	}
}

func (image *Image) DrawPolygons(em *Matrix, c color.Color) {
	if em.cols < 3 {
		fmt.Println("3 or more points are required for drawing")
		return
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
}

func (image *Image) DrawLine(x1, y1, x2, y2 int, c color.Color) {
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

func (image Image) drawOctant1(x1, y1, x2, y2, A, B int, c color.Color) {
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

func (image Image) drawOctant2(x1, y1, x2, y2, A, B int, c color.Color) {
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

func (image Image) drawOctant7(x1, y1, x2, y2, A, B int, c color.Color) {
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

func (image Image) drawOctant8(x1, y1, x2, y2, A, B int, c color.Color) {
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

func (image Image) Fill(c color.Color) {
	for y := 0; y < image.height; y++ {
		for x := 0; x < image.width; x++ {
			image.set(x, y, c)
		}
	}
}

func (image Image) set(x, y int, c color.Color) error {
	if x < 0 || x >= image.width {
		return errors.New("invalid x coordinate")
	}
	if y < 0 || y >= image.height {
		return errors.New("invalid y coordinate")
	}
	image.frame[y][x] = c
	return nil
}

func (image Image) SavePpm(name string) error {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("P3 %d %d %d\n", image.width, image.height, 255))
	for y := 0; y < image.height; y++ {
		for x := 0; x < image.width; x++ {
			color := image.frame[image.height-y-1][x]
			r, g, b, _ := color.RGBA()
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

func (image Image) Save(name string) error {
	index := strings.Index(name, ".")
	if index == -1 {
		return errors.New("no extension provided")
	}
	base := name[:index]
	ppm := base + ".ppm"
	err := image.SavePpm(ppm)
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

func (image Image) Display() error {
	filename := "tmp.ppm"
	err := image.SavePpm(filename)
	if err != nil {
		return err
	}
	args := []string{filename}
	_, err = exec.Command("display", args...).Output()
	if err != nil {
		return err
	}

	err = os.Remove(filename)
	if err != nil {
		return err
	}
	return nil
}

func isVisible(p0, p1, p2 []float64) bool {
	a := []float64{p1[0] - p0[0], p1[1] - p0[1], p1[2] - p0[2]}
	b := []float64{p2[0] - p0[0], p2[1] - p0[1], p2[2] - p0[2]}
	normal := CrossProduct(a, b)
	return normal[2] > 0
}
