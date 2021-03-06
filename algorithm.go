package main

import (
	"fmt"
	"image"
	"math"
	"math/rand"
	"runtime"
	"sync"
)

type JobQueue chan RenderJob
type Report chan struct{}

var doneSignal = struct{}{}

type RenderJob struct {
	Bounds image.Rectangle
	Params RenderParams
	Geoms  []Geometry
	Cam    *Camera
	Img    *image.RGBA
}

type RenderParams struct {
	SamplesPerPix int
	MaxBounces    int
	ChunkSize     int
}

func DefaultRenderParams() RenderParams {
	return RenderParams{
		SamplesPerPix: 16,
		MaxBounces:    4,
		ChunkSize:     32}
}

// Render manages the goroutine job-based rendering system.
func Render(scene []Geometry, cam *Camera, img *image.RGBA, params RenderParams) {

	chunkSize := params.ChunkSize

	hPieces := math.Ceil(float64(img.Bounds().Dx()) / float64(chunkSize))
	vPieces := math.Ceil(float64(img.Bounds().Dy()) / float64(chunkSize))
	totalPieces := int(hPieces * vPieces)

	queue := make(JobQueue, totalPieces)
	progress := make(Report, totalPieces) // semi-arbitrary chan length

	// progress reporter
	go func() {
		done := 0
		fmt.Printf("Rendered chunk %d of %d        \r", done, totalPieces)
		for range progress {
			done++
			fmt.Printf("Rendered chunk %d of %d        \r", done, totalPieces)
		}
	}()

	//spawn workers
	workers := runtime.NumCPU()
	wg := sync.WaitGroup{}
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func() {
			RenderWorker(queue, progress)
			wg.Done()
		}()
	}

	// divide img into chunks (RenderJobs)
	// and put jobs in queue
	for y := 0; y < img.Bounds().Max.Y; y += chunkSize {
		for x := 0; x < img.Bounds().Max.X; x += chunkSize {

			job := RenderJob{
				Bounds: image.Rect(x, y, x+chunkSize, y+chunkSize),
				Params: params,
				Geoms:  scene,
				Cam:    cam,
				Img:    img}

			queue <- job
		}
	}

	// wait for workers to complete
	close(queue)
	wg.Wait()
	close(progress)
}

// RenderWorker gets jobs and processes them.
func RenderWorker(jobs JobQueue, prog Report) {
	for job := range jobs {
		RenderChunk(job)
		prog <- doneSignal
	}
}

// RenderChunk performs the main rendering task, sampling each pixel of the
// chunk a certain number of times and writing the result to the final image.
func RenderChunk(job RenderJob) {
	rng := rand.New(rand.NewSource(int64(job.Img.Bounds().Min.X + job.Img.Bounds().Min.Y)))
	imgBounds := job.Img.Bounds()
	for r := job.Bounds.Min.Y; r < job.Bounds.Max.Y; r++ {
		for c := job.Bounds.Min.X; c < job.Bounds.Max.X; c++ {
			if isIn(c, r, imgBounds) {
				color := V3{0, 0, 0}
				for count := 0; count < job.Params.SamplesPerPix; count++ {
					yJit := Float(rng.Float64()) //*2 - 1 // [-0.5, 0.5)
					xJit := Float(rng.Float64()) //*2 - 1
					ray := job.Cam.GetRay(Float(c)+xJit, Float(r)+yJit)
					ray.Medium = &ambient
					sample := ShootRay(ray, job.Geoms, 0, job.Params.MaxBounces, rng)
					color = color.Add(sample)
				}
				job.Img.Set(c, r, V3ToColor(color.Mul(1/Float(job.Params.SamplesPerPix))))
			}
		}
	}
}

func isIn(c, r int, bound image.Rectangle) bool {
	return bound.Min.X <= c && c < bound.Max.X &&
		bound.Min.Y <= r && r < bound.Max.Y
}

///////////////////////////////////////////////////////////////

// FindNearestHit intersects the ray with all geoms and returns the nearest
// hit, if any.
func FindNearestHit(r Ray, geoms []Geometry) (min Hit, foundHit bool) {
	minT := Float(math.Inf(1))
	h := Hit{}
	for _, g := range geoms {
		if g.Hits(&h, r) {
			if epsilon < h.T && h.T < minT {
				min = h
				minT = h.T
				foundHit = true
			}
		}
	}

	return
}

// bullseye figures out if point is 100 units away from center, radially,
// allowing a sort of bullseye pattern to be made.
func bullseye(center, point V3) bool {
	c := point.Sub(center)
	return int(c.Len())%100 == 0
}

// grid figures out if the point is 100x units away from the plane "center"
// in X or Y, allowing a sort of grid to be made.
func grid(plane Plane, point V3) bool {
	// create new basis vectors in the plane
	b := createBasis(plane.Normal)
	x := b[0]
	y := b[1]

	// project point-plane.Point onto new x y axes
	diff := point.Sub(plane.Point)
	dx := diff.Dot(x) // x and y are normalized, so len=1
	dy := diff.Dot(y)

	return int(dx)%100 == 0 || int(dy)%100 == 0
}

// createBasis creates a basis coordinate system using direction.
func createBasis(direction V3) [3]V3 {
	z := direction.Normalize()
	var diff V3
	if math.Abs(z.X()) < 0.5 {
		diff = V3{1, 0, 0}
	} else {
		diff = V3{0, 1, 0}
	}
	x := z.Cross(diff).Normalize()
	y := x.Cross(z)

	return [3]V3{x, y, z}
}

var ambient = Material{
	Emittance:   V3{1, 1, 1},       // general environmental lighting ("sky")
	Reflectance: V3{0.9, 0.9, 0.1}, // ref properties of "dust particles"
	Eta:         1,                 // refraction coefficient of air (approx)
	Diffuse:     0.0 / 100.0}       // % chance of scatter in 100 units distance

var black = V3{}

// ShootRay recursively samples in the Ray r direction and produces a color value.
func ShootRay(r Ray, geoms []Geometry, depth, maxDepth int, rng *rand.Rand) (finalColor V3) {
	if depth == maxDepth {
		return black
	}

	hit, foundHit := FindNearestHit(r, geoms)
	if !foundHit {
		return black
	}

	// "haze"
	// if rand.Float64() < hit.T*ambient.Diffuse { // chance per some unit length
	// 	// cause ray to "redirect" at some random point along the ray
	// 	// between the ray's origin and the geometry it hit.
	// 	redirectP := r.Point(hit.T * Float(rand.Float64())) // some random point along Ray r
	// 	incCol := ShootRay(Ray{Orig: redirectP, Dir: RandomBounceSphere()}, geoms, depth-1)
	// 	return hadamard(incCol, ambient.Reflectance)
	// }

	mat := hit.Geom.Material()
	if mat.Emittance != black {
		return mat.Emittance // stop early for emitted
	}

	// HACK grids onto planes
	if plane, ok := hit.Geom.(Plane); ok {
		if grid(plane, hit.Point) {
			return black
		}
	}

	// russian roulette
	reflectance := mat.Reflectance
	// if depth > maxDepth/2 {
	// 	// given termination probability Q (should be low), accept by 1-Q
	//	// and weight by 1/(1-Q). here, max = 1-Q.
	// 	max := Float(math.Max(math.Max(float64(reflectance.X()), float64(reflectance.Y())), float64(reflectance.Z())))
	// 	if Float(rand.Float64()) < max {
	// 		reflectance = reflectance.Mul(1 / max)
	// 	} else {
	// 		return mat.Emittance // ? just from smallrt
	// 	}
	// }

	newRay := Ray{Orig: hit.Point, Medium: r.Medium}
	var incCol V3

	// perform a pure specular reflection sometimes per Diffuse
	if Float(rng.Float64()) < mat.Diffuse {
		newRay.Dir = randHemi(hit.Normal, rng)
		incCol = ShootRay(newRay, geoms, depth+1, maxDepth, rng)
	} else {
		if s, ok := hit.Geom.(Sphere); ok && s == geoms[0] { // do this only for the "glass" sphere
			// specular refraction
			refracCoeff := 1 - Schlick(r.Dir, hit.Normal, r.Medium.Eta, mat.Eta)
			if Float(rng.Float64()) < refracCoeff {
				if r.Dir.Dot(hit.Normal) < 0 {
					// out-to-inside
					newRay.Dir = RefractionDir(r.Dir, hit.Normal, r.Medium.Eta, mat.Eta)
					newRay.Medium = &mat
				} else {
					// in-to-outside
					newRay.Dir = RefractionDir(r.Dir, hit.Normal, r.Medium.Eta, ambient.Eta)
					newRay.Medium = &ambient
				}
				// newRay.Orig = r.Point(hit.T + 0.0001)
			} else {
				// specilar reflection if not refraction
				newRay.Dir = ReflectionDir(r.Dir, hit.Normal)
			}
		} else {
			// specular reflection
			newRay.Dir = ReflectionDir(r.Dir, hit.Normal)
		}

		incCol = ShootRay(newRay, geoms, depth+1, maxDepth, rng)
	}

	// rendering equation
	finalColor = hadamard(reflectance, incCol) //.Mul(abscos(hit.Normal, r.Dir)).Add(mat.Emittance)

	return
}

func lerp(v0, v1 V3, t Float) V3 {
	return v0.Mul(1 - t).Add(v1.Mul(t))
}

func hadamard(a, b V3) V3 {
	return V3{a.X() * b.X(), a.Y() * b.Y(), a.Z() * b.Z()}
}

func abscos(a, b V3) Float {
	return Float(math.Abs(float64(a.Dot(b))))
}
