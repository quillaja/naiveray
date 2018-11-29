package main

import (
	"math"
	"math/rand"
)

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

func randHemi(normal V3, rng *rand.Rand) V3 {
	b := createBasis(normal)
	z := rng.Float64()
	r := math.Sqrt(1.0 - z*z)
	theta := rng.Float64()*2*math.Pi - math.Pi
	x, y := math.Sincos(theta)
	x *= r
	y *= r
	return b[0].Mul(x).Add(b[1].Mul(y).Add(b[2].Mul(z)))
}

func randSphere(rng *rand.Rand) V3 {
	z := rng.Float64()*2 - 1 // [-1,1)
	r := math.Sqrt(1.0 - z*z)
	theta := rng.Float64()*2*math.Pi - math.Pi
	x, y := math.Sincos(theta)
	x *= r
	y *= r
	return V3{Float(x), Float(y), Float(z)}
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
