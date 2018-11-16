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
	Params *RenderParams
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
				Params: &params,
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

func RenderWorker(jobs JobQueue, prog Report) {
	for job := range jobs {
		RenderChunk(job)
		prog <- doneSignal
	}
}

func RenderChunk(job RenderJob) {
	imgBounds := job.Img.Bounds()
	for r := job.Bounds.Min.Y; r < job.Bounds.Max.Y; r++ {
		for c := job.Bounds.Min.X; c < job.Bounds.Max.X; c++ {
			if isIn(c, r, imgBounds) {
				color := V3{0, 0, 0}
				for count := 0; count < job.Params.SamplesPerPix; count++ {
					yJit := Float(rand.Float64())*2 - 1 // [-0.5, 0.5)
					xJit := Float(rand.Float64())*2 - 1
					ray := job.Cam.GetRay(Float(c)+xJit, Float(r)+yJit)
					ray.Medium = &ambient
					sample := ShootRay(ray, job.Geoms, job.Params.MaxBounces)
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

func ReflectionDir(incident, surfaceNormal V3) V3 {
	IdotN := incident.Dot(surfaceNormal)
	return incident.Sub(surfaceNormal.Mul(2 * IdotN))
}

func RefractionDir(incident, surfaceNormal V3, eta1, eta2 Float) V3 {
	// TODO: get this crap about incident directions figured out once
	// and for all
	// This works.
	// wikipedia assumes incident ray is from cam to point,
	// PBR assumes incident ray is from point to cam.
	r := eta1 / eta2
	c := incident.Dot(surfaceNormal)
	if c > 0 {
		// in-out
		surfaceNormal = surfaceNormal.Mul(-1)
	}

	rightside := r*c - Float(math.Sqrt(float64(1-(r*r)*(1-c*c))))
	return incident.Mul(-r).Add(surfaceNormal.Mul(rightside))
}

func Schlick(incident, surfaceNormal V3, eta1, eta2 Float) Float {

	cosTheta := float64(incident.Dot(surfaceNormal))
	if eta1 > eta2 && math.Acos(cosTheta) >= math.Asin(eta2/eta1) {
		// at or beyond critical angle for total internal reflection
		return 1
	}

	r0 := Float(math.Pow(float64((eta1-eta2)/(eta1+eta2)), 2))
	costerm := 1 - Float(math.Abs(cosTheta)) // (1 - (-I dot N))
	R := r0 + (1-r0)*(costerm*costerm*costerm*costerm*costerm)
	if R < 0 || R > 1 {
		r0 = 10
	}
	return R
}

// RandomBounceHemisphere gets a
func RandomBounceHemisphere(normal V3) V3 {
	rval := V3{
		Float(rand.Float64())*2 - 1,
		Float(rand.Float64())*2 - 1,
		Float(rand.Float64())*2 - 1}.Normalize()

	if normal.Dot(rval) < 0 {
		rval = rval.Mul(-1)
	}

	return rval
}

// RandomBounceSphere gets a
func RandomBounceSphere() V3 {
	return V3{
		Float(rand.Float64())*2 - 1,
		Float(rand.Float64())*2 - 1,
		Float(rand.Float64())*2 - 1}.Normalize()
}

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

func grid(a, b V3) bool {
	c := b.Sub(a)
	// x, y, z := int(c.X()), int(c.Y()), int(c.Z())
	return int(c.Len())%100 == 0
}

var ambient = Material{
	Emittance:   V3{1, 1, 1},       // general environmental lighting ("sky")
	Reflectance: V3{0.9, 0.9, 0.1}, // ref properties of "dust particles"
	Eta:         1,                 // refraction coefficient of air (approx)
	Diffuse:     0.0 / 100.0,       // % chance of scatter in 100 units distance
	Glossy:      0}                 // meaningless in this context

func ShootRay(r Ray, geoms []Geometry, depth int) (finalColor V3) {
	if depth == 0 {
		return ambient.Emittance
	}

	hit, foundHit := FindNearestHit(r, geoms)
	if !foundHit {
		return ambient.Emittance
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
	if mat.Emittance.Len() > 0 {
		return mat.Emittance // stop early for emitted
	}

	// HACK grids onto planes
	if plane, ok := hit.Geom.(Plane); ok {
		if grid(plane.Point, hit.Point) {
			return V3{0.1, 0.1, 0.1}
		}
	}

	newRay := Ray{Orig: hit.Point, Medium: r.Medium}

	// perform a pure specular reflection sometimes per Diffuse
	// perform a weighted diffuse reflection per Glossy
	newRay.Dir = ReflectionDir(r.Dir, hit.Normal)
	if Float(rand.Float64()) < mat.Diffuse {
		newRay.Dir = RandomBounceHemisphere(hit.Normal)
	}
	// else {
	// 	newRay.Dir = lerp(
	// 		RandomBounceHemisphere(hit.Normal),
	// 		newRay.Dir,
	// 		mat.Glossy).Normalize()
	// }

	if s, ok := hit.Geom.(Sphere); ok && s == geoms[0] { // do this only for the "glass" sphere
		refracCoeff := 1 - Schlick(r.Dir, hit.Normal, r.Medium.Eta, mat.Eta)
		if Float(rand.Float64()) < refracCoeff {
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
		}
	}

	incCol := ShootRay(newRay, geoms, depth-1)

	// rendering equation
	cosTerm := Float(math.Abs(float64(newRay.Dir.Dot(hit.Normal))))
	if cosTerm < 0.5 { // questionable?
		cosTerm = 0.5
	}

	finalColor = hadamard(mat.Reflectance, incCol).Mul(cosTerm).Add(mat.Emittance)

	return
}

func lerp(v0, v1 V3, t Float) V3 {
	return v0.Mul(1 - t).Add(v1.Mul(t))
}

func hadamard(a, b V3) V3 {
	return V3{a.X() * b.X(), a.Y() * b.Y(), a.Z() * b.Z()}
}
