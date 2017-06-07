package main

// CrossProduct calculates the cross product of two vectors
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

func Normal(p0, p1, p2 []float64) []float64 {
	return []float64{
		(p1[1]-p0[1])*(p2[2]-p0[2]) - (p1[2]-p0[2])*(p2[1]-p0[1]),
		(p1[2]-p0[2])*(p2[0]-p0[0]) - (p1[0]-p0[0])*(p2[2]-p0[2]),
		(p1[0]-p0[0])*(p2[1]-p0[1]) - (p1[1]-p0[1])*(p2[0]-p0[0]),
	}
}

// DotProduct calculates the dot product of two vectors
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
