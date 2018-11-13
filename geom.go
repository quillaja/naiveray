package main

import (
	"math"
)

// Hit represents an intersection between a Ray and a Geometry.
type Hit struct {
	Point  V3       // point of intersection
	Normal V3       // surface normal of hit geometry
	Geom   Geometry // reference to the geometry
	T      Float    // parametric "t" on ray of the intersection
}

// Ray represents a ray defined by an origin point and
// a normalized direction vector.
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

// Sphere represents a sphere. It's defined by a center point and a radius.
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

// Plane represents a plane. It's defined by a point and a normal to the plane.
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
