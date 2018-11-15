package main

import "image/color"

type Material struct {
	Emittance   V3
	Reflectance V3
	Eta         Float
	Diffuse     Float
	Glossy      Float
}

func ColorToV3(col color.Color) V3 {
	r, g, b, _ := col.RGBA()
	rval := V3{Float(r), Float(g), Float(b)}.Mul(1 / Float(255))
	return rval
}

func V3ToColor(vec V3) color.Color {
	return color.RGBA{
		R: uint8(clamp(255*vec.X(), 0, 255)),
		G: uint8(clamp(255*vec.Y(), 0, 255)),
		B: uint8(clamp(255*vec.Z(), 0, 255)),
		A: 255}
}

func clamp(n, min, max Float) Float {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
