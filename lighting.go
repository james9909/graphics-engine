package main

var (
	DefaultViewVector = []float64{0, 0, -1}
)

func FlatShading(p0, p1, p2, I_a, K_a, I_i, K_d, K_s, view []float64, lights [][]float64) []float64 {
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

func flatDiffuseLight(p0, p1, p2, I_i, K_d, light []float64) []float64 {
	normal := Normal(p0, p1, p2)

	light = Normalize(light)
	normal = Normalize(normal)
	diffuseVector := DotProduct(light, normal)

	diffuse := []float64{
		I_i[0] * K_d[0] * diffuseVector,
		I_i[1] * K_d[1] * diffuseVector,
		I_i[2] * K_d[2] * diffuseVector,
	}

	for i := range diffuse {
		if diffuse[i] < 0 {
			diffuse[i] = 0
		}
	}
	return diffuse
}

func flatSpecularLight(p0, p1, p2, I_i, K_s, light, view []float64) []float64 {
	normal := Normal(p0, p1, p2)

	light = Normalize(light)
	normal = Normalize(normal)
	dot := DotProduct(light, normal)

	reflect := Normalize(Subtract(Scale(normal, dot*2), light))
	specularVector := DotProduct(reflect, view)

	specular := []float64{
		I_i[0] * K_s[0] * specularVector,
		I_i[1] * K_s[1] * specularVector,
		I_i[2] * K_s[2] * specularVector,
	}

	for i := range specular {
		if specular[i] < 0 {
			specular[i] = 0
		}
	}
	return specular
}
