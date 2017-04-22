package main

// DrawingMode defines the type of each drawing mode
type DrawingMode int

const (
	// StepSize is the number of steps to take when drawing 2D curves
	StepSize float64 = (1.0 / 100.0)
	//CircularStepSize is the number of steps to take when drawing 3D curves
	CircularStepSize float64 = (1.0 / 20.0)

	// DefaultHeight is the default height of an Image
	DefaultHeight = 500
	// DefaultWidth is the default width of an Image
	DefaultWidth = 500

	// DrawLineMode is a draw argument that draws 2D lines onto the Image
	DrawLineMode = 0
	// DrawPolygonMode is a draw argument that draws 3D polygons onto the Image
	DrawPolygonMode = 1
)
