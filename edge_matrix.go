package main

func NewEdgeMatrix() *Matrix {
	return NewMatrix(4, 0)
}

func (m *Matrix) AddPoint(x, y, z int) {
	column := []float64{
		float64(x),
		float64(y),
		float64(z),
		1.0}
	m.AddColumn(column)
}

func (m *Matrix) AddEdge(x0, y0, z0, x1, y1, z1 int) {
	m.AddPoint(x0, y0, z0)
	m.AddPoint(x1, y1, z1)
}
