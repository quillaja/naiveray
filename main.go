// package main implements a simple/naive ray/path tracer
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"runtime/pprof"
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
	outputF := flag.String("out", "output.png", "output file name (PNG)")
	profF := flag.String("profile", "", "filename for cpu profile output")

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

	geoms := Scene1()

	// profiling
	if *profF != "" {
		prof, err := os.Create(*profF)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(prof)
		defer pprof.StopCPUProfile()
	}

	fmt.Printf("%d x %d image, %d samples per px\n", width, height, raysPerPx)
	fmt.Println("Beginning render")

	start := time.Now()

	Render(geoms, cam, img,
		RenderParams{
			SamplesPerPix: raysPerPx,
			MaxBounces:    bounces,
			ChunkSize:     *chunkSizeF})

	took := time.Since(start).Seconds()

	fmt.Printf("Render complete. Writing to %s\n", *outputF)
	fmt.Printf("Took: %.2f s\n", took)
	fmt.Printf("Time/Sample: %.5f ms\n", 1000*took/Float(width*height*raysPerPx))

	file, err := os.Create(*outputF)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	png.Encode(file, img)
}

func Scene1() []Geometry {

	return []Geometry{

		// high left ball
		Sphere{
			Center: V3{150, 100, -100},
			Radius: 100,
			Mat: Material{
				Reflectance: V3{0.99, 0.99, 0.99},
				Eta:         1.5,
				Diffuse:     0,
				Glossy:      0.95},
		},

		// low right ball
		Sphere{
			Center: V3{200, -100, -100},
			Radius: 100,
			Mat: Material{
				Reflectance: V3{0.99, 0.99, 0.99},
				Diffuse:     1,
				Glossy:      1},
		},

		// mirror ball
		Sphere{
			Center: V3{200, 150, 0},
			Radius: 50,
			Mat: Material{
				Reflectance: V3{0.99, 0.99, 0.99},
				Diffuse:     0},
		},

		// back glowing ball
		Sphere{
			Center: V3{500, 0, 550}, //orig z = -175
			Radius: 200,
			Mat: Material{
				Emittance: V3{15, 15, 15}},
		},

		// front-of-camera wall (rear of scene)
		Plane{
			Point:  V3{800, 0, 0},
			Normal: V3{-1, 0, 0},
			Mat: Material{
				Reflectance: V3{0.2, 0.2, 0.99},
				Diffuse:     1,
				Glossy:      0.05},
		},

		// behind camera wall
		Plane{
			Point:  V3{-800, 0, 0},
			Normal: V3{1, 0, 0},
			Mat: Material{
				Reflectance: V3{0.99, 0.2, 0.99},
				Diffuse:     1,
				Glossy:      0.05},
		},

		// ceiling
		Plane{
			Point:  V3{0, 0, 400},
			Normal: V3{0, 0, -1},
			Mat: Material{
				Reflectance: V3{0.95, 0.95, 0.95},
				// Emittance:   V3{1, 1, 1},
				Diffuse: 1,
				Glossy:  0.05},
		},

		// floor
		Plane{
			Point:  V3{0, 0, -200},
			Normal: V3{0, 0, 1},
			Mat: Material{
				Reflectance: V3{0.95, 0.95, 0.95},
				Diffuse:     1},
		},

		// left wall
		Plane{
			Point:  V3{0, 400, 0},
			Normal: V3{0, -1, 0},
			Mat: Material{
				Reflectance: V3{0.2, 0.99, 0.2},
				Diffuse:     1,
				Glossy:      0.25},
		},

		// right wall
		Plane{
			Point:  V3{0, -400, 0},
			Normal: V3{0, 1, 0},
			Mat: Material{
				Reflectance: V3{1, 0.2, 0.2},
				Diffuse:     1,
				Glossy:      0},
		},
	}
}

func Scene2() []Geometry {
	return []Geometry{
		// Sphere{
		// 	Center: V3{0, 0, 50},
		// 	Radius: 50,
		// 	Mat: Material{
		// 		Reflectance: V3{1, 1, 1},
		// 		Eta:         1.5},
		// },

		// TriangleMesh{
		// 	Verts: []V3{
		// 		V3{0, 0, 100},
		// 		V3{50, 50, 0},
		// 		V3{-50, 50, 0},
		// 		V3{-50, -50, 0},
		// 		V3{50, -50, 0}},
		// 	Index: []int{0, 1, 2, 0, 2, 3, 0, 3, 4, 0, 4, 1, 1, 3, 2, 1, 4, 3},
		// 	Mat: Material{
		// 		Reflectance: V3{1, 0.8, 0.8},
		// 		Diffuse:     0.8},
		// },

		TriangleMesh{
			Verts: []V3{
				V3{0, 0, 70},
				V3{50, 50, 10},
				V3{-50, 50, 10},
				V3{-50, -50, 10},
				V3{50, -50, 10}},
			Index: []int{0, 1, 2, 0, 2, 3, 0, 3, 4, 0, 4, 1},
			Mat: Material{
				Reflectance: V3{1, 0.5, 0.5},
				Diffuse:     0.9},
		},

		Plane{
			Point:  V3{0, 0, 0},
			Normal: V3{0, 0, 1},
			Mat: Material{
				Reflectance: V3{1, 1, 1},
				Diffuse:     1},
		},

		Plane{
			Point:  V3{0, 0, 500},
			Normal: V3{0, 0, -1},
			Mat: Material{
				Emittance: V3{1.5, 1.5, 1.5}},
		},
	}
}
