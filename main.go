package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"
)

func main() {

	fmt.Println("Testing matrices:")
	data := [][]float64{
		[]float64{1, 2, 3},
		[]float64{4, 5, 6},
		[]float64{7, 8, 9},
	}
	m1 := NewMatrix(3, 3)
	ident := IdentityMatrix(3)
	fmt.Println("Identity matrix:")
	fmt.Println(ident)
	fmt.Println("3x3 matrix:")
	fmt.Println(m1)
	m1.Multiply(ident)
	fmt.Println("Identity * 3x3:")
	fmt.Println(m1)
	fmt.Println("Setting matrix with pre-defined values:")
	m1.SetMatrix(data)
	fmt.Println(m1)
	m1.Scale(4)
	fmt.Println("Scale by 4:")
	fmt.Println(m1)

	data2 := [][]float64{
		[]float64{9, 8, 7},
		[]float64{6, 5, 4},
		[]float64{3, 2, 1},
	}
	m2 := NewMatrix(3, 3)
	m2.SetMatrix(data2)
	m1.Multiply(m2)
	fmt.Println("Multiply m1*m2:")
	fmt.Println(m1)

	fmt.Println("Generating image...")

	rand.Seed(time.Now().UTC().UnixNano())

	image := NewImage(500, 500)
	for x := 0.0; x < 500; x += 5 {
		em := NewEdgeMatrix()
		em.AddEdge(0, x, 0, x, 499, 0)
		em.AddEdge(499, x, 0, x, 0, 0)
		em.AddEdge(x, x, 0, x, x, 0)
		image.DrawLines(em, color.RGBA{uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256)), 0})
	}
	image.Save("out.png")
	image.Display()
}
