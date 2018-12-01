package main

import (
	"math"
	"math/rand"
)

// ReflectDir gets a new direction reflected around normal.
func ReflectionDir(incident, surfaceNormal V3) V3 {
	IdotN := incident.Dot(surfaceNormal)
	return incident.Sub(surfaceNormal.Mul(2 * IdotN))
}

// RefractionDir gets a new direction based on refraction.
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

// Schlick gets the reflectance coefficient (R) using the schlick approximation
// of the fresnel terms.
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

// randHemi produces a uniform random direction on the hemisphere around normal.
// NOTE: createBasis() may not be robust.
func randHemi(normal V3, rng *rand.Rand) V3 {
	b := createBasis(normal)
	z := rng.Float64()
	r := math.Sqrt(1.0 - z*z)
	phi := rng.Float64() * 2 * math.Pi
	y, x := math.Sincos(phi)
	x *= r
	y *= r
	return b[0].Mul(x).Add(b[1].Mul(y).Add(b[2].Mul(z)))
}

// randSphere produces a uniform random direction on a sphere.
func randSphere(rng *rand.Rand) V3 {
	z := rng.Float64()*2 - 1 // [-1,1)
	r := math.Sqrt(1.0 - z*z)
	phi := rng.Float64() * 2 * math.Pi
	y, x := math.Sincos(phi)
	x *= r
	y *= r
	return V3{Float(x), Float(y), Float(z)}
}

// RandomBounceHemisphere gets a
// DEPRECATED
func RandomBounceHemisphere(normal V3, rng *rand.Rand) V3 {
	rval := V3{
		Float(rng.Float64())*2 - 1,
		Float(rng.Float64())*2 - 1,
		Float(rng.Float64())*2 - 1}.Normalize()

	if normal.Dot(rval) < 0 {
		rval = rval.Mul(-1)
	}

	return rval
}

// RandomBounceSphere gets a
// DEPRECATED
func RandomBounceSphere(rng *rand.Rand) V3 {
	return V3{
		Float(rng.Float64())*2 - 1,
		Float(rng.Float64())*2 - 1,
		Float(rng.Float64())*2 - 1}.Normalize()
}
