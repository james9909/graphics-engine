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
		data[i] = make([]float64, cols, 64)
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
