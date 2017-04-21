package main

import (
	"bytes"
	"errors"
	"fmt"
	"math"
)

type Matrix struct {
	data [][]float64
	rows int
	cols int
}

func (m Matrix) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{\n")
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			buffer.WriteString(fmt.Sprintf("%.2f, ", m.data[i][j]))
		}
		buffer.WriteString("\n")
	}
	buffer.WriteString("}\n")
	return buffer.String()
}

func NewMatrix(rows, cols int) *Matrix {
	data := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		data[i] = make([]float64, cols, cols*2)
	}
	return &Matrix{
		data: data,
		rows: rows,
		cols: cols,
	}
}

func NewMatrixFromData(data [][]float64) *Matrix {
	m := NewMatrix(len(data), len(data[0]))
	m.data = data
	m.rows = len(data)
	m.cols = len(data[0])
	return m
}

func IdentityMatrix(size int) *Matrix {
	m := NewMatrix(size, size)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if i == j {
				m.data[i][j] = 1
			} else {
				m.data[i][j] = 0
			}
		}
	}
	return m
}

func (m *Matrix) Copy() *Matrix {
	return NewMatrixFromData(m.data)
}

func (m Matrix) Get(r, c int) float64 {
	return m.data[r][c]
}

func (m Matrix) GetMatrix() [][]float64 {
	return m.data
}

func (m *Matrix) SetMatrix(data [][]float64) {
	m.data = data
	m.rows = len(data)
	m.cols = len(data[0])
}

func (m *Matrix) Scale(n float64) *Matrix {
	m2 := NewMatrix(m.rows, m.cols)
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			m2.data[i][j] = m.Get(i, j) * n
		}
	}
	return m2
}

func (m *Matrix) Multiply(m2 *Matrix) (*Matrix, error) {
	if m.cols != m2.rows {
		return nil, fmt.Errorf("column/row mismatch: (%d x %d) * (%d x %d)", m.rows, m.cols, m2.rows, m2.cols)
	}

	product := NewMatrix(m.rows, m2.cols)
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m2.cols; j++ {
			var sum float64
			for k := 0; k < m.cols; k++ {
				sum += m.Get(i, k) * m2.Get(k, j)
			}
			product.data[i][j] = sum
		}
	}
	return product, nil
}

func (m *Matrix) AddColumn(column []float64) error {
	if len(column) != m.rows {
		return errors.New("incorrect number of rows")
	}
	for i, v := range column {
		m.data[i] = append(m.data[i], v)
	}
	m.cols++
	return nil
}

func MakeTranslation(x, y, z float64) *Matrix {
	data := [][]float64{
		[]float64{1, 0, 0, float64(x)},
		[]float64{0, 1, 0, float64(y)},
		[]float64{0, 0, 1, float64(z)},
		[]float64{0, 0, 0, 1},
	}
	m := NewMatrixFromData(data)
	return m
}

func MakeDilation(sx, sy, sz float64) *Matrix {
	data := [][]float64{
		[]float64{sx, 0, 0, 0},
		[]float64{0, sy, 0, 0},
		[]float64{0, 0, sz, 0},
		[]float64{0, 0, 0, 1},
	}
	m := NewMatrixFromData(data)
	return m
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

func MakeRotX(theta float64) *Matrix {
	theta = degreesToRadians(theta)
	data := [][]float64{
		[]float64{1, 0, 0, 0},
		[]float64{0, math.Cos(theta), -math.Sin(theta), 0},
		[]float64{0, math.Sin(theta), math.Cos(theta), 0},
		[]float64{0, 0, 0, 1},
	}
	m := NewMatrixFromData(data)
	return m
}

func MakeRotY(theta float64) *Matrix {
	theta = degreesToRadians(theta)
	data := [][]float64{
		[]float64{math.Cos(theta), 0, -math.Sin(theta), 0},
		[]float64{0, 1, 0, 0},
		[]float64{math.Sin(theta), 0, math.Cos(theta), 0},
		[]float64{0, 0, 0, 1},
	}
	m := NewMatrixFromData(data)
	return m
}

func MakeRotZ(theta float64) *Matrix {
	theta = degreesToRadians(theta)
	data := [][]float64{
		[]float64{math.Cos(theta), -math.Sin(theta), 0, 0},
		[]float64{math.Sin(theta), math.Cos(theta), 0, 0},
		[]float64{0, 0, 1, 0},
		[]float64{0, 0, 0, 1},
	}
	m := NewMatrixFromData(data)
	return m
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

func (m *Matrix) AddPolygon(x0, y0, z0, x1, y1, z1, x2, y2, z2 float64) {
	m.AddPoint(x0, y0, z0)
	m.AddPoint(x1, y1, z1)
	m.AddPoint(x2, y2, z2)
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

func (m *Matrix) AddHermite(x0, y0, x1, y1, dx0, dy0, dx1, dy1 float64) {
	coefficientsX := generateHermiteCoefficients(x0, dx0, x1, dx1)
	coefficientsY := generateHermiteCoefficients(y0, dy0, y1, dy1)
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
}

func generateHermiteCoefficients(p0, m0, p1, m1 float64) *Matrix {
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
	m, _ := hermite.Multiply(coefficients)
	return m
}

func (m *Matrix) AddBezier(x0, y0, x1, y1, x2, y2, x3, y3 float64) error {
	coefficientsX := generateBezierCoefficients(x0, x1, x2, x3)
	coefficientsY := generateBezierCoefficients(y0, y1, y2, y3)
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

func generateBezierCoefficients(p0, p1, p2, p3 float64) *Matrix {
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
	m, _ := bezier.Multiply(coefficients)
	return m
}

func (m *Matrix) AddBox(x, y, z, width, height, depth float64) {
	x1 := x + width
	y1 := y + height
	z1 := z - depth

	// Front
	m.AddPolygon(x, y1, z, x, y, z, x1, y, z)
	m.AddPolygon(x, y1, z, x1, y, z, x1, y1, z)
	// Back
	m.AddPolygon(x1, y1, z1, x1, y, z1, x, y, z1)
	m.AddPolygon(x1, y1, z1, x, y, z1, x, y1, z1)
	// Top
	m.AddPolygon(x, y1, z1, x, y1, z, x1, y1, z)
	m.AddPolygon(x, y1, z1, x1, y1, z, x1, y1, z1)
	// Bottom
	m.AddPolygon(x1, y, z1, x1, y, z, x, y, z)
	m.AddPolygon(x1, y, z1, x, y, z, x, y, z1)
	// Left
	m.AddPolygon(x, y1, z1, x, y, z1, x, y, z)
	m.AddPolygon(x, y1, z1, x, y, z, x, y1, z)
	// Right
	m.AddPolygon(x1, y1, z, x1, y, z, x1, y, z1)
	m.AddPolygon(x1, y1, z, x1, y, z1, x1, y1, z1)
}

func (m *Matrix) AddSphere(cx, cy, cz, radius float64) {
	points := NewMatrix(4, 0)
	points.generateSphere(cx, cy, cz, radius)
	steps := int(1.0/CircularStepSize) + 1
	endLatitude := steps - 1
	endLongitude := steps - 1
	for latitude := 0; latitude < endLatitude; latitude++ {
		latitudeStart := latitude * steps
		nextLatitudeStart := latitudeStart + steps
		for longitude := 0; longitude < endLongitude; longitude++ {
			index := latitudeStart + longitude
			indexNextLatitude := nextLatitudeStart + longitude
			if longitude > 0 {
				m.AddPolygon(
					points.Get(0, index), points.Get(1, index), points.Get(2, index),
					points.Get(0, index+1), points.Get(1, index+1), points.Get(2, index+1),
					points.Get(0, indexNextLatitude), points.Get(1, indexNextLatitude), points.Get(2, indexNextLatitude))
			}
			// Don't draw the triangles at the end pole of the sphere
			if longitude < endLongitude-1 {
				m.AddPolygon(
					points.Get(0, index+1), points.Get(1, index+1), points.Get(2, index+1),
					points.Get(0, indexNextLatitude+1), points.Get(1, indexNextLatitude+1), points.Get(2, indexNextLatitude+1),
					points.Get(0, indexNextLatitude), points.Get(1, indexNextLatitude), points.Get(2, indexNextLatitude))
			}
		}
	}
}

func (m *Matrix) generateSphere(cx, cy, cz, radius float64) {
	for r := 0.0; r < 1+CircularStepSize; r += CircularStepSize {
		phi := 2 * math.Pi * r
		rCosPhi := radius * math.Cos(phi)
		rSinPhi := radius * math.Sin(phi)
		for c := 0.0; c < 1+CircularStepSize; c += CircularStepSize {
			theta := math.Pi * c
			cosTheta := math.Cos(theta)
			sinTheta := math.Sin(theta)

			x := cx + radius*cosTheta
			y := cy + sinTheta*rCosPhi
			z := cz + sinTheta*rSinPhi
			m.AddPoint(x, y, z)
		}
	}
}

func (m *Matrix) AddTorus(cx, cy, cz, r1, r2 float64) {
	points := NewMatrix(4, 0)
	points.generateTorus(cx, cy, cz, r1, r2)
	steps := int(1.0/CircularStepSize) + 1
	endLatitude := steps - 1
	endLongitude := steps - 1
	for latitude := 0; latitude < endLatitude; latitude++ {
		for longitude := 0; longitude < endLongitude; longitude++ {
			index := latitude*steps + longitude
			m.AddPolygon(
				points.Get(0, index), points.Get(1, index), points.Get(2, index),
				points.Get(0, index+1), points.Get(1, index+1), points.Get(2, index+1),
				points.Get(0, index+steps), points.Get(1, index+steps), points.Get(2, index+steps))
			m.AddPolygon(
				points.Get(0, index+steps+1), points.Get(1, index+steps+1), points.Get(2, index+steps+1),
				points.Get(0, index+steps), points.Get(1, index+steps), points.Get(2, index+steps),
				points.Get(0, index+1), points.Get(1, index+1), points.Get(2, index+1))
		}
	}
}

func (m *Matrix) generateTorus(cx, cy, cz, r1, r2 float64) {
	for r := 0.0; r < 1+CircularStepSize; r += CircularStepSize {
		phi := 2 * math.Pi * r
		cosPhi := math.Cos(phi)
		sinPhi := math.Sin(phi)
		for c := 0.0; c < 1+CircularStepSize; c += CircularStepSize {
			theta := 2 * math.Pi * c
			cosTheta := math.Cos(theta)
			sinTheta := math.Sin(theta)

			x := cosPhi*(r1*cosTheta+r2) + cx
			y := r1*sinTheta + cy
			z := sinPhi*(r1*cosTheta+r2) + cz
			m.AddPoint(x, y, z)
		}
	}
}
