// package main implements a simple/naive ray/path tracer
package main

import (
	"fmt"
	"image"
	"image/png"
	"math/rand"
	"os"
	"time"
)

func main() {

	// rand.Seed(time.Now().UnixNano())

	const width = 600
	const height = 400
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	var filmWidth, filmHeight, scale Float
	if width >= height {
		filmWidth = 300
		filmHeight = filmWidth * (Float(height) / Float(width))
		scale = filmWidth / width
	} else {
		filmHeight = 300
		filmWidth = filmHeight * (Float(width) / Float(height))
		scale = filmHeight / height
	}

	focal := V3{-100, 0, 0}

	const raysPerPx = 16
	const bounces = 4

	geoms := Scene()

	fmt.Println("Beginning render")
	start := time.Now()

	for r := 0; r < height; r++ {
		for c := 0; c < width; c++ {
			color := V3{0, 0, 0}
			for count := 0; count < raysPerPx; count++ {
				yJit := Float(rand.Float64())*2 - 1
				xJit := Float(rand.Float64())*2 - 1
				y := filmHeight/2 - (scale * (Float(r) + yJit))
				x := filmWidth/2 - (scale * (Float(c) + xJit))
				rayD := V3{0, x, y}.Sub(focal).Normalize()
				ray := Ray{Orig: focal, Dir: rayD}
				sample := ShootRay(ray, geoms, bounces)
				color = color.Add(sample)
			}
			img.Set(c, r, V3ToColor(color.Mul(1/Float(raysPerPx))))
		}
	}

	fmt.Println("Render complete. Writing to output.png")
	fmt.Println("Took:", time.Since(start).Seconds(), "s")

	file, err := os.Create("output.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	png.Encode(file, img)
}

func Scene() []Geometry {

	return []Geometry{

		// front-of-camera wall (rear of scene)
		Plane{
			Point:  V3{800, 0, 0},
			Normal: V3{-1, 0, 0},
			Mat: Material{
				Col:         V3{0.2, 0.2, 1},
				Specularity: 0.5},
		},

		// behind camera wall
		Plane{
			Point:  V3{-800, 0, 0},
			Normal: V3{1, 0, 0},
			Mat: Material{
				Col:         V3{1, 0.2, 1},
				Specularity: 0.5},
		},

		// ceiling
		Plane{
			Point:  V3{0, 0, 400},
			Normal: V3{0, 0, -1},
			Mat: Material{
				Col:         V3{1, 1, 1},
				Emit:        V3{2, 2, 2},
				Specularity: 0.1},
		},

		// floor
		Plane{
			Point:  V3{0, 0, -200},
			Normal: V3{0, 0, 1},
			Mat: Material{
				Col:         V3{1, 1, 1},
				Specularity: 0.1},
		},

		// left wall
		Plane{
			Point:  V3{0, 400, 0},
			Normal: V3{0, -1, 0},
			Mat: Material{
				Col:         V3{0.2, 1, 0.2},
				Specularity: 0.5},
		},

		// right wall
		Plane{
			Point:  V3{0, -400, 0},
			Normal: V3{0, 1, 0},
			Mat: Material{
				Col:         V3{1, 0.2, 0.2},
				Specularity: 0.5},
		},

		// high mirror ball
		Sphere{
			Center: V3{150, 100, 100},
			Radius: Float(100),
			Mat: Material{
				Col:         V3{1, 1, 0.75},
				Specularity: 0.95},
		},
	}
}
