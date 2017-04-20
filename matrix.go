package main

import (
	"bytes"
	"fmt"
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
		data[i] = make([]float64, cols)
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

func (m *Matrix) Scale(n float64) {
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			m.data[i][j] *= n
		}
	}
}

func (m *Matrix) Multiply(m2 *Matrix) {
	if m.cols != m2.rows {
		panic("invalid matrix dimensions")
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
	m.SetMatrix(product.data)
}

func (m *Matrix) AddRow(row []float64) {
	m.SetMatrix(append(m.data, row))
}
