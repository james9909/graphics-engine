package main

import "image/color"

func main() {
	image := NewImage(500, 500)
	image.fill(color.Black)
	c := color.RGBA{255, 0, 0, 0}

	centerX, centerY := image.height/2, image.width/2

	for x := 0; x < 500; x += 10 {
		image.drawLine(centerX, centerY, x, 0, c)
		image.drawLine(centerX, centerY, x, 499, c)
	}

	for y := 0; y < 500; y += 10 {
		image.drawLine(centerX, centerY, 0, y, c)
		image.drawLine(centerX, centerY, 499, y, c)
	}

	image.drawLine(centerX, centerY, 499, 499, c)

	image.save("out.ppm")
	image.display()
}
