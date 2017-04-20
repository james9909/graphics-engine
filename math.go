package main

import "math"

func CrossProduct(a, b []float64) []float64 {
	if len(a) < 3 || len(b) < 3 {
		panic("invalid vector length")
	}
	cross := make([]float64, 3)
	cross[0] = a[1]*b[2] - a[2]*b[1]
	cross[1] = a[2]*b[0] - a[0]*b[2]
	cross[2] = a[0]*b[1] - a[1]*b[0]
	return cross
}

func DotProduct(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("vector lengths are unequal")
	}
	sum := 0.0
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func Magnitude(a []float64) float64 {
	sum := 0.0
	for i := range a {
		sum += a[i] * a[i]
	}
	return math.Sqrt(sum)
}
