package main

import (
	"bytes"
	"errors"
	"fmt"
	"math"
)

const (
	// StepSize is the number of steps to take when drawing 2D curves
	StepSize float64 = (1.0 / 100.0)
	//CircularStepSize is the number of steps to take when drawing 3D curves
	CircularStepSize float64 = (1.0 / 20.0)
)

// Matrix represents a matrix
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

// NewMatrix returns a new Matrix with a given number of rows and columns
func NewMatrix(rows, cols int) *Matrix {
	data := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		data[i] = make([]float64, cols, cols)
	}
	return &Matrix{
		data: data,
		rows: rows,
		cols: cols,
	}
}

// NewMatrixFromData returns a new Matrix with preset data
func NewMatrixFromData(data [][]float64) *Matrix {
	m := NewMatrix(len(data), len(data[0]))
	m.data = data
	m.rows = len(data)
	m.cols = len(data[0])
	return m
}

// IdentityMatrix returns an identity matrix
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

// Copy returns a copy of a Matrix
func (m *Matrix) Copy() *Matrix {
	return NewMatrixFromData(m.data)
}

// Get returns the value at a certain row and column in a Matrix
func (m Matrix) Get(r, c int) float64 {
	return m.data[r][c]
}

// GetColumn returns a column of the Matrix
func (m Matrix) GetColumn(c int) []float64 {
	col := make([]float64, m.rows)
	for i := 0; i < m.rows; i++ {
		col[i] = m.Get(i, c)
	}
	return col
}

// GetMatrix returns a 2D array that represents the matrix
func (m Matrix) GetMatrix() [][]float64 {
	return m.data
}

// SetMatrix sets the data for a Matrix
func (m *Matrix) SetMatrix(data [][]float64) {
	m.data = data
	m.rows = len(data)
	m.cols = len(data[0])
}

// Scale scales a matrix by a factor
func (m *Matrix) Scale(n float64) *Matrix {
	m2 := NewMatrix(m.rows, m.cols)
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			m2.data[i][j] = m.Get(i, j) * n
		}
	}
	return m2
}

// Multiply returns the product of two Matrices
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

// AddColumn adds a new column to the matrix
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

// MakeTranslation returns a translation Matrix
func MakeTranslation(x, y, z float64) *Matrix {
	data := [][]float64{
		{1, 0, 0, float64(x)},
		{0, 1, 0, float64(y)},
		{0, 0, 1, float64(z)},
		{0, 0, 0, 1},
	}
	m := NewMatrixFromData(data)
	return m
}

// MakeDilation returns a dilation matrix
func MakeDilation(sx, sy, sz float64) *Matrix {
	data := [][]float64{
		{sx, 0, 0, 0},
		{0, sy, 0, 0},
		{0, 0, sz, 0},
		{0, 0, 0, 1},
	}
	m := NewMatrixFromData(data)
	return m
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

// MakeRotX returns a rotation matrix for the X axis
func MakeRotX(theta float64) *Matrix {
	theta = degreesToRadians(theta)
	data := [][]float64{
		{1, 0, 0, 0},
		{0, math.Cos(theta), math.Sin(theta), 0},
		{0, -math.Sin(theta), math.Cos(theta), 0},
		{0, 0, 0, 1},
	}
	m := NewMatrixFromData(data)
	return m
}

// MakeRotY returns a rotation matrix for the Y axis
func MakeRotY(theta float64) *Matrix {
	theta = degreesToRadians(theta)
	data := [][]float64{
		{math.Cos(theta), 0, -math.Sin(theta), 0},
		{0, 1, 0, 0},
		{math.Sin(theta), 0, math.Cos(theta), 0},
		{0, 0, 0, 1},
	}
	m := NewMatrixFromData(data)
	return m
}

// MakeRotZ returns a rotation matrix for the Z axis
func MakeRotZ(theta float64) *Matrix {
	theta = degreesToRadians(theta)
	data := [][]float64{
		{math.Cos(theta), math.Sin(theta), 0, 0},
		{-math.Sin(theta), math.Cos(theta), 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
	m := NewMatrixFromData(data)
	return m
}

// AddPoint adds a point to the matrix as a column
func (m *Matrix) AddPoint(x, y, z float64) {
	column := []float64{
		x,
		y,
		z,
		1.0}
	m.AddColumn(column)
}

// AddEdge adds two points to the matrix
func (m *Matrix) AddEdge(x0, y0, z0, x1, y1, z1 float64) {
	m.AddPoint(x0, y0, z0)
	m.AddPoint(x1, y1, z1)
}

// AddTriangle adds three points to the matrix
func (m *Matrix) AddTriangle(x0, y0, z0, x1, y1, z1, x2, y2, z2 float64) {
	m.AddPoint(x0, y0, z0)
	m.AddPoint(x1, y1, z1)
	m.AddPoint(x2, y2, z2)
}

// AddCircle adds a series of points defining a circle to the matrix
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

// AddHermite adds a series of points defining a hermite curve to the matrix
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
		{2, -2, 1, 1},
		{-3, 3, -2, -1},
		{0, 0, 1, 0},
		{1, 0, 0, 0},
	})
	coefficients := NewMatrixFromData([][]float64{
		{p0},
		{p1},
		{m0},
		{m1},
	})
	m, _ := hermite.Multiply(coefficients)
	return m
}

// AddBezier adds a series of points defining a bezier curve to the matrix
func (m *Matrix) AddBezier(x0, y0, x1, y1, x2, y2, x3, y3 float64) {
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
}

func generateBezierCoefficients(p0, p1, p2, p3 float64) *Matrix {
	bezier := NewMatrixFromData([][]float64{
		{-1, 3, -3, 1},
		{3, -6, 3, 0},
		{-3, 3, 0, 0},
		{1, 0, 0, 0},
	})
	coefficients := NewMatrixFromData([][]float64{
		{p0},
		{p1},
		{p2},
		{p3},
	})
	m, _ := bezier.Multiply(coefficients)
	return m
}

// AddBox adds a series of points defining a 3D box to the matrix
func (m *Matrix) AddBox(x, y, z, width, height, depth float64) {
	x1 := x + width
	y1 := y - height
	z1 := z - depth

	// Front
	m.AddTriangle(x, y1, z, x, y, z, x1, y, z)
	m.AddTriangle(x, y1, z, x1, y, z, x1, y1, z)
	// Back
	m.AddTriangle(x1, y1, z1, x1, y, z1, x, y, z1)
	m.AddTriangle(x1, y1, z1, x, y, z1, x, y1, z1)
	// Top
	m.AddTriangle(x, y1, z1, x, y1, z, x1, y1, z)
	m.AddTriangle(x, y1, z1, x1, y1, z, x1, y1, z1)
	// Bottom
	m.AddTriangle(x1, y, z1, x1, y, z, x, y, z)
	m.AddTriangle(x1, y, z1, x, y, z, x, y, z1)
	// Left
	m.AddTriangle(x, y1, z1, x, y, z1, x, y, z)
	m.AddTriangle(x, y1, z1, x, y, z, x, y1, z)
	// Right
	m.AddTriangle(x1, y1, z, x1, y, z, x1, y, z1)
	m.AddTriangle(x1, y1, z, x1, y, z1, x1, y1, z1)
}

// AddSphere adds a series of points defining a 3D sphere to the matrix
func (m *Matrix) AddSphere(cx, cy, cz, radius float64) {
	points := NewMatrix(4, 0)
	points.generateSphere(cx, cy, cz, radius)
	steps := int(1.0/CircularStepSize) + 1
	endLatitude := steps - 1
	endLongitude := steps - 1
	modulus := points.cols
	for latitude := 0; latitude < endLatitude; latitude++ {
		start := latitude * steps
		nextStart := (start + steps) % modulus
		for longitude := 0; longitude < endLongitude; longitude++ {
			p0 := start + longitude
			p1 := p0 + 1
			p2 := nextStart + longitude
			p3 := p2 + 1

			if longitude > 0 {
				m.AddTriangle(
					points.Get(0, p0), points.Get(1, p0), points.Get(2, p0),
					points.Get(0, p3), points.Get(1, p3), points.Get(2, p3),
					points.Get(0, p2), points.Get(1, p2), points.Get(2, p2))
			}
			if longitude != endLongitude-1 {
				m.AddTriangle(
					points.Get(0, p3), points.Get(1, p3), points.Get(2, p3),
					points.Get(0, p0), points.Get(1, p0), points.Get(2, p0),
					points.Get(0, p1), points.Get(1, p1), points.Get(2, p1))
			}
		}
	}
}

func (m *Matrix) generateSphere(cx, cy, cz, radius float64) {
	steps := float64(int(1.0 / CircularStepSize))
	for r := 0.0; r < steps; r++ {
		phi := math.Pi * (2 * r / steps)
		rCosPhi := radius * math.Cos(phi)
		rSinPhi := radius * math.Sin(phi)
		for c := 0.0; c <= steps; c++ {
			theta := math.Pi * (c / steps)
			cosTheta := math.Cos(theta)
			sinTheta := math.Sin(theta)

			x := cx + radius*cosTheta
			y := cy + sinTheta*rCosPhi
			z := cz + sinTheta*rSinPhi
			m.AddPoint(x, y, z)
		}
	}
}

// AddTorus adds a series of points defining a 3D torus to the matrix
func (m *Matrix) AddTorus(cx, cy, cz, r1, r2 float64) {
	points := NewMatrix(4, 0)
	points.generateTorus(cx, cy, cz, r1, r2)
	steps := int(1.0/CircularStepSize) + 1
	endLatitude := steps - 1
	endLongitude := steps - 1
	for latitude := 0; latitude < endLatitude; latitude++ {
		start := latitude * steps
		for longitude := 0; longitude < endLongitude; longitude++ {
			p0 := start + longitude
			p1 := p0 + 1
			p2 := p0 + steps
			p3 := p2 + 1
			m.AddTriangle(
				points.Get(0, p0), points.Get(1, p0), points.Get(2, p0),
				points.Get(0, p1), points.Get(1, p1), points.Get(2, p1),
				points.Get(0, p2), points.Get(1, p2), points.Get(2, p2))
			m.AddTriangle(
				points.Get(0, p3), points.Get(1, p3), points.Get(2, p3),
				points.Get(0, p2), points.Get(1, p2), points.Get(2, p2),
				points.Get(0, p1), points.Get(1, p1), points.Get(2, p1))
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
