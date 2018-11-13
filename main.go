// package main implements a simple/naive ray/path tracer
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"time"
)

func main() {

	// rand.Seed(time.Now().UnixNano())

	widthF := flag.Int("width", 300, "output width")
	heightF := flag.Int("height", 200, "output height")
	raysPerPxF := flag.Int("rays", 16, "sample rays per pixel")
	bouncesF := flag.Int("bounces", 4, "max bounces per path")
	fovF := flag.Float64("fov", 115.0, "field of view in degrees of widest part")
	xposF := flag.Float64("xpos", -200, "")
	flag.Parse()

	width := *widthF
	height := *heightF
	raysPerPx := *raysPerPxF
	bounces := *bouncesF
	fov := *fovF

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	cam := NewCamera(
		V3{*xposF, 0, 0}, V3{*xposF + 1, 0, 0}, V3{0, 0, 1},
		Float(math.Pi*(fov/180.0)), width, height)

	geoms := Scene()

	fmt.Printf("%d x %d image, %d samples per px\n", width, height, raysPerPx)
	fmt.Println("Beginning render")
	start := time.Now()

	Render(geoms, cam, img, &RenderParams{SamplesPerPix: raysPerPx, MaxBounces: bounces})

	took := time.Since(start).Seconds()
	fmt.Println("Render complete. Writing to output.png")
	fmt.Printf("Took: %.2f s\n", took)
	fmt.Printf("Sec/Sample: %.5f ms\n", 1000*took/Float(width*height*raysPerPx))

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

		// low semi-mirror ball
		Sphere{
			Center: V3{200, -100, -100},
			Radius: Float(100),
			Mat: Material{
				Col:         V3{0.7, 0.7, 1},
				Specularity: 0.65},
		},

		// high mirror ball
		Sphere{
			Center: V3{700, 300, -175},
			Radius: Float(80),
			Mat: Material{
				Emit: V3{5, 1, 5}},
		},
	}
}
