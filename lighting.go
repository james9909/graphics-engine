package main

import "math"

var (
	DefaultViewVector = []float64{0, 0, 1}
)

type LightSource struct {
	location []float64
	color    Color
}

func FlatShading(p0, p1, p2, I_a, K_a, I_i, K_d, K_s, view []float64, lights map[string]LightSource) []float64 {
	I := []float64{0, 0, 0}
	ambient := flatAmbientLight(I_a, K_a)
	for a := range ambient {
		I[a] += ambient[a]
	}
	for _, light := range lights {
		diffuse := flatDiffuseLight(p0, p1, p2, I_i, K_d, light)
		specular := flatSpecularLight(p0, p1, p2, I_i, K_s, light, view)
		for d := range diffuse {
			I[d] += diffuse[d]
		}
		for s := range specular {
			I[s] += specular[s]
		}
	}
	return I
}

func flatAmbientLight(I_a, K_a []float64) []float64 {
	ambient := []float64{
		I_a[0] * K_a[0],
		I_a[1] * K_a[1],
		I_a[2] * K_a[2],
	}
	return ambient
}

func flatDiffuseLight(p0, p1, p2, I_i, K_d []float64, light LightSource) []float64 {
	normal := Normal(p0, p1, p2)

	lightVector := Normalize(light.location)
	normal = Normalize(normal)
	diffuseVector := DotProduct(lightVector, normal)

	diffuse := make([]float64, 3)
	if I_i[0] > 0 || I_i[1] > 0 || I_i[2] > 0 {
		copy(diffuse, I_i)
	} else {
		diffuse = []float64{float64(light.color.r), float64(light.color.g), float64(light.color.b)}
	}

	for i := range diffuse {
		diffuse[i] = math.Max(diffuse[i]*K_d[i]*diffuseVector, 0)
	}

	return diffuse
}

func flatSpecularLight(p0, p1, p2, I_i, K_s []float64, light LightSource, view []float64) []float64 {
	normal := Normal(p0, p1, p2)

	lightVector := Normalize(light.location)
	normal = Normalize(normal)
	dot := DotProduct(lightVector, normal)

	reflect := Normalize(Subtract(Scale(normal, dot*2), light.location))
	specularVector := DotProduct(reflect, view)

	specular := make([]float64, 3)
	if I_i[0] > 0 || I_i[1] > 0 || I_i[2] > 0 {
		copy(specular, I_i)
	} else {
		specular = []float64{float64(light.color.r), float64(light.color.g), float64(light.color.b)}
	}

	for i := range specular {
		specular[i] = math.Max(specular[i]*K_s[i]*specularVector, 0)
	}

	return specular
}
