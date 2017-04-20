package main

type Edge struct {
	*Matrix
}

func NewEdgeMatrix() *Edge {
	return &Edge{
		NewMatrix(0, 4),
	}
}

func (e *Edge) AddPoint(x, y, z float64) {
	row := []float64{x, y, z, 1.0}
	e.AddRow(row)
}

func (e *Edge) AddEdge(x0, y0, z0, x1, y1, z1 float64) {
	e.AddPoint(x0, y0, z0)
	e.AddPoint(x1, y1, z1)
}
