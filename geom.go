package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/go-gl/mathgl/mgl64"
)

const epsilon = 0.0001

// V3 is a 3d vector.
type V3 = mgl64.Vec3

// Float is a floating point number
type Float = float64

// Hit represents an intersection between a Ray and a Geom.
type Hit struct {
	Point  V3       // point of intersection
	Normal V3       // surface normal of hit geometry
	Geom   Geometry // reference to the geometry
	T      Float    // parametric "t" on ray of the intersection
}

// Ray represents a ray
type Ray struct {
	Orig V3
	Dir  V3
}

// Point returns the point along the ray at "t".
func (r Ray) Point(t Float) V3 {
	return r.Orig.Add(r.Dir.Mul(t))
}

// Geometry abstracts the interface for various geometries.
type Geometry interface {
	Hits(r Ray) []Hit
	Material() Material
}

type Sphere struct {
	Center V3
	Radius Float
	Mat    Material
}

func (s Sphere) Hits(r Ray) []Hit {
	OC := r.Orig.Sub(s.Center)
	dirDotOC := r.Dir.Dot(OC)
	OClen := OC.Len()
	rside := dirDotOC*dirDotOC - (OClen * OClen) + (s.Radius * s.Radius)
	t := -dirDotOC

	if rside == 0 {
		p := r.Point(t)
		return []Hit{
			Hit{
				T:      t,
				Point:  p,
				Normal: p.Sub(s.Center).Normalize(),
				Geom:   s}}

	} else if rside > 0 {
		rside = Float(math.Sqrt(float64(rside)))
		p1 := r.Point(t - rside)
		p2 := r.Point(t + rside)
		return []Hit{
			Hit{
				T:      t - rside,
				Point:  p1,
				Normal: p1.Sub(s.Center).Normalize(),
				Geom:   s},
			Hit{
				T:      t + rside,
				Point:  p2,
				Normal: p2.Sub(s.Center).Normalize(),
				Geom:   s}}
	}

	return []Hit{}
}

func (s Sphere) Material() Material {
	return s.Mat
}

type Plane struct {
	Point  V3
	Normal V3
	Mat    Material
}

func (p Plane) Hits(r Ray) []Hit {
	denom := r.Dir.Dot(p.Normal)

	if denom != 0 {
		t := p.Point.Sub(r.Orig).Dot(p.Normal) / denom
		return []Hit{
			Hit{
				T:      t,
				Point:  r.Point(t),
				Normal: p.Normal,
				Geom:   p}}
	}

	return []Hit{} // plane and ray are parallel

}

func (p Plane) Material() Material {
	return p.Mat
}

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

/////////////////////////////////////////

type Material struct {
	Emit        V3
	Col         V3
	Specularity Float
}

func ColorToV3(col color.Color) V3 {
	r, g, b, _ := col.RGBA()
	rval := V3{Float(r), Float(g), Float(b)}.Mul(1 / 255)
	return rval
}

func V3ToColor(vec V3) color.Color {
	return color.RGBA{
		R: uint8(255 * vec.X()),
		G: uint8(255 * vec.Y()),
		B: uint8(255 * vec.Z()),
		A: 255}
}
