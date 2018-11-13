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
type Report chan int

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
		for p := range progress {
			done += p
			fmt.Printf("Rendered chunk %d of %d        \r", done, totalPieces) //100*Float(done)/Float(total))
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
	// dy := img.Bounds().Dy() / div
	// dx := img.Bounds().Dx() / div
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
		prog <- 1
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

// RandomBounce gets a
func RandomBounce(normal V3) V3 {
	rval := V3{
		Float(rand.Float64())*2 - 1,
		Float(rand.Float64())*2 - 1,
		Float(rand.Float64())*2 - 1}.Normalize()

	if normal.Dot(rval) < 0 {
		rval = rval.Mul(-1)
	}

	return rval
}

func FindNearestHit(r Ray, geoms []Geometry) (min Hit, foundHit bool) {
	minT := Float(math.Inf(1))
	for _, g := range geoms {
		hits := g.Hits(r)
		for _, h := range hits {
			if epsilon < h.T && h.T < minT {
				min = h
				minT = h.T
				foundHit = true
			}
		}
	}

	return
}

func ShootRay(r Ray, geoms []Geometry, depth int) (finalColor V3) {
	if depth == 0 {
		return // returns black
	}

	hit, foundHit := FindNearestHit(r, geoms)

	if !foundHit {
		fmt.Println("didnt find hit")
		return // return black
	}

	mat := hit.Geom.Material()
	if mat.Emit.Len() > 0 {
		return mat.Emit // stop early for emitted
	}

	reflectD := ReflectionDir(r.Dir, hit.Normal)
	if mat.Specularity < 1 {
		reflectD = lerp(
			RandomBounce(hit.Normal),
			reflectD,
			mat.Specularity).Normalize()
	}

	incCol := ShootRay(Ray{Orig: hit.Point, Dir: reflectD}, geoms, depth-1)

	// rendering equation
	cosFalloff := r.Dir.Mul(-1).Dot(hit.Normal)
	cosFalloff = Float(math.Min(math.Max(0.1, float64(cosFalloff)), 1)) // clamp [0.1,1]

	finalColor = hadamard(mat.Col, incCol).Mul(cosFalloff).Add(mat.Emit)

	return
}

func lerp(v0, v1 V3, t Float) V3 {
	return v0.Mul(1 - t).Add(v1.Mul(t))
}

func hadamard(a, b V3) V3 {
	return V3{a.X() * b.X(), a.Y() * b.Y(), a.Z() * b.Z()}
}
