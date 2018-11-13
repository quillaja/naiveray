// package main implements a simple/naive ray/path tracer
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type VecFlag V3

func (v *VecFlag) String() string {
	return fmt.Sprint(V3(*v))
}

func (v *VecFlag) Set(val string) error {
	in := [3]Float{}
	for i, p := range strings.Split(val, ",")[:3] {
		n, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return err
		}
		in[i] = Float(n)
	}

	*v = VecFlag(in)

	return nil
}

func main() {

	// rand.Seed(time.Now().UnixNano())

	// TODO: flag vars need cleaning up and streamlined.
	// many things can be stored right into the actual vars.
	widthF := flag.Int("width", 300, "output width")
	heightF := flag.Int("height", 200, "output height")
	raysPerPxF := flag.Int("rays", 16, "sample rays per pixel")
	bouncesF := flag.Int("bounces", 4, "max bounces per path")
	chunkSizeF := flag.Int("chunk", 32, "square chunk size for render chunk")

	fovF := flag.Float64("fov", 115.0, "field of view in degrees of widest part")
	camPosF := VecFlag(V3{-100, 0, 0})
	lookF := VecFlag(V3{0, 0, 0})
	camUpF := VecFlag(V3{0, 0, 1})
	flag.Var(&camPosF, "cam", "camera position v3 in world coords")
	flag.Var(&lookF, "look", "v3 point in world coords the camera is looking at")
	flag.Var(&camUpF, "up", "camera up v3 in world coords")

	flag.Parse()

	width := *widthF
	height := *heightF
	raysPerPx := *raysPerPxF
	bounces := *bouncesF
	fov := *fovF

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	cam := NewCamera(
		V3(camPosF), V3(lookF), V3(camUpF),
		Float(math.Pi*(fov/180.0)), width, height)

	geoms := Scene()

	fmt.Printf("%d x %d image, %d samples per px\n", width, height, raysPerPx)
	fmt.Println("Beginning render")
	start := time.Now()

	Render(geoms, cam, img,
		RenderParams{
			SamplesPerPix: raysPerPx,
			MaxBounces:    bounces,
			ChunkSize:     *chunkSizeF})

	took := time.Since(start).Seconds()
	fmt.Println("Render complete. Writing to output.png")
	fmt.Printf("Took: %.2f s\n", took)
	fmt.Printf("Time/Sample: %.5f ms\n", 1000*took/Float(width*height*raysPerPx))

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
