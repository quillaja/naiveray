package main

import mgl "github.com/go-gl/mathgl/mgl64" // change this to mgl32 for 32 bit floats

// epsilon is a "near zero" value. I'm not really using it mathematically correctly.
const epsilon = 0.0001

// Below aliases are to make it easy to switch between 32 and 64 bit floats

// V3 is a 3d vector.
type V3 = mgl.Vec3

// M4 is a 4x4 matrix.
type M4 = mgl.Mat4

// Float is a floating point number
type Float = float64
