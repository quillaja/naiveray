package main

import (
	"bufio"
	"io"
	"math"
	"strconv"
	"strings"
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
	Orig   V3
	Dir    V3
	Medium *Material
}

// Point returns the point along the ray at "t".
func (r Ray) Point(t Float) V3 {
	return r.Orig.Add(r.Dir.Mul(t))
}

// Geometry abstracts the interface for various geometries.
type Geometry interface {
	Hits(hit *Hit, r Ray) bool
	Material() Material
}

// Sphere represents a sphere. It's defined by a center point and a radius.
type Sphere struct {
	Center V3
	Radius Float
	Mat    Material
}

func (s Sphere) Hits(hit *Hit, r Ray) bool {
	OC := r.Orig.Sub(s.Center)
	dirDotOC := r.Dir.Dot(OC)
	OClen := OC.Len()
	rside := dirDotOC*dirDotOC - (OClen * OClen) + (s.Radius * s.Radius)
	t := -dirDotOC

	if rside == 0 {
		p := r.Point(t)
		hit.T = t
		hit.Point = p
		hit.Normal = p.Sub(s.Center).Normalize()
		hit.Geom = s
		return true

	} else if rside > 0 {
		rside = Float(math.Sqrt(float64(rside)))
		p1 := r.Point(t - rside)
		hit.T = t - rside
		hit.Point = p1
		hit.Normal = p1.Sub(s.Center).Normalize()
		hit.Geom = s
		return true
	}

	return false
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

func (p Plane) Hits(hit *Hit, r Ray) bool {
	denom := r.Dir.Dot(p.Normal)

	if denom != 0 {
		t := p.Point.Sub(r.Orig).Dot(p.Normal) / denom
		hit.T = t
		hit.Point = r.Point(t)
		hit.Normal = p.Normal
		hit.Geom = p
		return true
	}

	return false // plane and ray are parallel

}

func (p Plane) Material() Material {
	return p.Mat
}

type TriangleMesh struct {
	Verts []V3
	Index []int
	Mat   Material
}

func (t TriangleMesh) count() int {
	return len(t.Index) / 3
}

func (t TriangleMesh) get(which int) (verts [3]V3) {
	for i, v := range t.Index[which*3 : (which+1)*3] {
		verts[i] = t.Verts[v]
	}
	return
}

func (t TriangleMesh) Hits(hit *Hit, r Ray) bool {
	// for each triangle, test ray and triangle,
	// while also finding the minimun t
	minT := Float(math.Inf(1))
	found := false
	candidate := &Hit{}
	for i := 0; i < t.count(); i++ {
		verts := t.get(i)
		if rayTriangle(candidate, verts, r) {
			if epsilon < candidate.T && candidate.T < minT {
				*hit = *candidate
				hit.Geom = t
				minT = hit.T
				found = true
			}
		}
	}

	return found
}

func (t TriangleMesh) Material() Material {
	return t.Mat
}

func rayTriangle(hit *Hit, verts [3]V3, r Ray) bool {
	// get normal, use 1st vert as 'root'. assume CCW winding
	// (though i guess it doesnt really matter). then find intersection with
	// plane of triangle.
	a := verts[0]
	b := verts[1]
	c := verts[2]
	ba := b.Sub(a)
	ca := c.Sub(a)
	normal := ba.Cross(ca).Normalize()

	// ray-plane
	denom := r.Dir.Dot(normal)
	var t Float
	if denom == 0 { // triangle and ray are perpendicular
		return false
	}
	t = a.Sub(r.Orig).Dot(normal) / denom
	if t < epsilon { // triangle is behind ray
		return false
	}

	// ray hit plane somewhere, and that is point p
	p := r.Point(t)

	// then use cross and dotproducts to determine
	// if the point is within the triangle.
	// already have B-A
	cb := c.Sub(b)
	ac := a.Sub(c)

	// compare cross products with triangle normal because the cross products
	// were calculated in the same CCW direction as the normal
	baCp := ba.Cross(p.Sub(a)).Dot(normal)
	cbCp := cb.Cross(p.Sub(b)).Dot(normal)
	acCp := ac.Cross(p.Sub(c)).Dot(normal)

	// don't need to check if <0 because the normal is derived from b and c.
	// therefore, the normal and cross products are always one the same 'side'
	if baCp >= 0 && cbCp >= 0 && acCp >= 0 {
		// p in triangle
		hit.Normal = normal
		hit.Point = p
		hit.T = t
		return true
	}

	return false
}

// TriangleMeshFromOBJ parses limited info about verticies "v" and
// faces "f" from an OBJ file to produce a TrangleMesh.
func TriangleMeshFromOBJ(r io.Reader) (t TriangleMesh) {

	t.Verts = []V3{}
	t.Index = []int{}

	scan := bufio.NewScanner(r)

	for scan.Scan() {
		line := scan.Text()
		parts := strings.Split(line, "  ")
		switch parts[0] {
		case "v":
			v := V3{}
			for i := 1; i < 4; i++ {
				n, _ := strconv.ParseFloat(parts[i], 64)
				v[i-1] = Float(n)
			}
			t.Verts = append(t.Verts, v)

		case "f":
			f := make([]int, 3)
			for i := 1; i < 4; i++ {
				n, _ := strconv.Atoi(strings.Split(parts[i], "/")[0])
				f[i-1] = n - 1 // OBJ indices start at 1
			}
			t.Index = append(t.Index, f...)
		}
	}
	if err := scan.Err(); err != nil {
		panic(err)
	}

	return
}
