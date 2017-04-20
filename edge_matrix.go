package main

import "math"

const (
	StepSize         float64 = (1.0 / 1000.0)
	CircularStepSize float64 = (1.0 / 200.0)
)

func NewEdgeMatrix() *Matrix {
	return NewMatrix(4, 0)
}

func (m *Matrix) AddPoint(x, y, z float64) {
	column := []float64{
		x,
		y,
		z,
		1.0}
	m.AddColumn(column)
}

func (m *Matrix) AddEdge(x0, y0, z0, x1, y1, z1 float64) {
	m.AddPoint(x0, y0, z0)
	m.AddPoint(x1, y1, z1)
}

func (m *Matrix) AddCircle(cx, cy, cz, radius float64) {
	x0 := cx + radius
	y0 := cy

	for t := 0.0; t <= 1; t += StepSize {
		theta := 2 * math.Pi * t
		x1 := radius*math.Cos(theta) + cx
		y1 := radius*math.Sin(theta) + cy
		m.AddEdge(x0, y0, cz, x1, y1, cz)
		x0 = x1
		y0 = y1
	}
}

func (m *Matrix) AddHermite(x0, y0, x1, y1, dx0, dy0, dx1, dy1 float64) error {
	coefficientsX, err := generateHermiteCoefficients(x0, dx0, x1, dx1)
	if err != nil {
		return err
	}
	coefficientsY, err := generateHermiteCoefficients(y0, dy0, y1, dy1)
	if err != nil {
		return err
	}
	aX, bX, cX, dX := coefficientsX.Get(0, 0), coefficientsX.Get(1, 0), coefficientsX.Get(2, 0), coefficientsX.Get(3, 0)
	aY, bY, cY, dY := coefficientsY.Get(0, 0), coefficientsY.Get(1, 0), coefficientsY.Get(2, 0), coefficientsY.Get(3, 0)
	for t := 0.0; t <= 1; t += StepSize {
		tSquared := t * t
		tCubed := tSquared * t
		x1 := aX*tCubed + bX*tSquared + cX*t + dX
		y1 := aY*tCubed + bY*tSquared + cY*t + dY
		m.AddEdge(x0, y0, 0, x1, y1, 0)
		x0 = x1
		y0 = y1
	}
	return nil
}

func generateHermiteCoefficients(p0, m0, p1, m1 float64) (*Matrix, error) {
	hermite := NewMatrixFromData([][]float64{
		[]float64{2, -2, 1, 1},
		[]float64{-3, 3, -2, -1},
		[]float64{0, 0, 1, 0},
		[]float64{1, 0, 0, 0},
	})
	coefficients := NewMatrixFromData([][]float64{
		[]float64{p0},
		[]float64{p1},
		[]float64{m0},
		[]float64{m1},
	})
	m, err := hermite.Multiply(coefficients)
	return m, err
}

func (m *Matrix) AddBezier(x0, y0, x1, y1, x2, y2, x3, y3 float64) error {
	coefficientsX, err := generateBezierCoefficients(x0, x1, x2, x3)
	if err != nil {
		return err
	}
	coefficientsY, err := generateBezierCoefficients(y0, y1, y2, y3)
	if err != nil {
		return err
	}
	aX, bX, cX, dX := coefficientsX.Get(0, 0), coefficientsX.Get(1, 0), coefficientsX.Get(2, 0), coefficientsX.Get(3, 0)
	aY, bY, cY, dY := coefficientsY.Get(0, 0), coefficientsY.Get(1, 0), coefficientsY.Get(2, 0), coefficientsY.Get(3, 0)
	for t := 0.0; t <= 1; t += StepSize {
		tSquared := t * t
		tCubed := tSquared * t
		x1 := aX*tCubed + bX*tSquared + cX*t + dX
		y1 := aY*tCubed + bY*tSquared + cY*t + dY
		m.AddEdge(x0, y0, 0, x1, y1, 0)
		x0 = x1
		y0 = y1
	}
	return nil
}

func generateBezierCoefficients(p0, p1, p2, p3 float64) (*Matrix, error) {
	bezier := NewMatrixFromData([][]float64{
		[]float64{-1, 3, -3, 1},
		[]float64{3, -6, 3, 0},
		[]float64{-3, 3, 0, 0},
		[]float64{1, 0, 0, 0},
	})
	coefficients := NewMatrixFromData([][]float64{
		[]float64{p0},
		[]float64{p1},
		[]float64{p2},
		[]float64{p3},
	})
	m, err := bezier.Multiply(coefficients)
	return m, err
}

func (m *Matrix) AddBox(x, y, z, width, height, depth float64) {
	x1 := x + width
	y1 := y + height
	z1 := z - depth

	m.AddEdge(x, y, z, x, y1, z)
	m.AddEdge(x, y, z, x, y, z1)
	m.AddEdge(x, y, z, x1, y, z)
	m.AddEdge(x1, y1, z, x1, y1, z1)
	m.AddEdge(x1, y1, z, x, y1, z)
	m.AddEdge(x1, y1, z, x1, y, z)
	m.AddEdge(x1, y, z1, x1, y, z)
	m.AddEdge(x1, y, z1, x1, y1, z1)
	m.AddEdge(x1, y, z1, x, y, z1)
	m.AddEdge(x1, y1, z1, x, y1, z1)
	m.AddEdge(x, y1, z1, x, y1, z)
	m.AddEdge(x, y1, z1, x, y, z1)
}

func (m *Matrix) AddSphere(cx, cy, cz, radius float64) {
	steps := math.Floor((1.0 / CircularStepSize) + 0.5)
	for r := 0.0; r < steps; r++ {
		phi := 2 * math.Pi * (r / steps)
		rCosPhi := radius * math.Cos(phi)
		rSinPhi := radius * math.Sin(phi)
		for c := 0.0; c < steps; c++ {
			theta := 2 * math.Pi * (c / steps)
			cosTheta := math.Cos(theta)
			sinTheta := math.Sin(theta)

			x := cx + radius*cosTheta
			y := cy + sinTheta*rCosPhi
			z := cz + sinTheta*rSinPhi
			m.AddPoint(x, y, z)
			m.AddPoint(x, y, z)
		}
	}
}

func (m *Matrix) AddTorus(cx, cy, cz, r1, r2 float64) {
	steps := math.Floor((1.0 / CircularStepSize) + 0.5)
	for r := 0.0; r < steps; r++ {
		phi := 2 * math.Pi * (r / steps)
		cosPhi := math.Cos(phi)
		sinPhi := math.Sin(phi)
		for c := 0.0; c < steps; c++ {
			theta := 2 * math.Pi * (c / steps)
			cosTheta := math.Cos(theta)
			sinTheta := math.Sin(theta)

			x := cosPhi*(r1*cosTheta+r2) + cx
			y := r1*sinTheta + cy
			z := sinPhi*(r1*cosTheta+r2) + cz
			m.AddPoint(x, y, z)
			m.AddPoint(x, y, z)
		}
	}
}
