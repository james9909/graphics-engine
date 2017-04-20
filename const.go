package main

type DrawingMode int

const (
	StepSize         float64 = (1.0 / 100.0)
	CircularStepSize float64 = (1.0 / 20.0)

	DefaultHeight = 500
	DefaultWidth  = 500

	DrawLineMode    = 0
	DrawPolygonMode = 1
)
