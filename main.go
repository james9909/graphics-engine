package main

import (
	"bytes"
	"fmt"
	"os"
)

func main() {
	flag := "flag{heres_a_ppm}"
	width, height := 500, 500

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("P3 %d %d 255\n", width, height))
	i := 0
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			buffer.WriteString(fmt.Sprintf("%d %d %d\n", x%255, y%255, flag[i%len(flag)]))
			i++
		}
	}

	f, err := os.OpenFile("out.ppm", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	f.WriteString(buffer.String())
	f.Close()
}
