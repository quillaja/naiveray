package main

import (
	"math"

	mgl "github.com/go-gl/mathgl/mgl64" // change this to mgl32 for 32 bit floats
)

// LookAt returns a camera-to-world 4x4 transformation matrix
func LookAt(camPos, lookPt, camUp V3) M4 {
	// supposedly this does the same thing as the inverse of the mgl.LookAt()
	// dir := lookPt.Sub(camPos).Normalize()
	// left := camUp.Normalize().Cross(dir).Normalize()
	// newUp := dir.Cross(left)
	// return M4{
	// 	left.X(), newUp.X(), dir.X(), camPos.X(),
	// 	left.Y(), newUp.Y(), dir.Y(), camPos.Y(),
	// 	left.Z(), newUp.Z(), dir.Z(), camPos.Z(),
	// 	0, 0, 0, 1}
	return mgl.LookAtV(camPos, lookPt, camUp).Inv()
}

type Camera struct {
	CamToWorld            M4
	Pos                   V3
	Look                  V3
	Up                    V3
	Fov                   Float
	FilmWidth, FilmHeight Float
	imgPlaneW, imgPlaneH  Float
	imgPlaneD             Float
	unitPPW, unitPPH      Float
}

func NewCamera(pos, look, up V3, fov Float, filmw, filmh int) *Camera {
	c := &Camera{
		Pos:        pos,
		Look:       look,
		Up:         up,
		Fov:        fov,
		FilmWidth:  Float(filmw),
		FilmHeight: Float(filmh),
		CamToWorld: LookAt(pos, look, up)}

	if filmw >= filmh {
		c.imgPlaneW = 2.0
		c.imgPlaneH = 2.0 * c.FilmHeight / c.FilmWidth
		c.imgPlaneD = (0.5 * c.imgPlaneH) / math.Tan(float64(fov/2)) // numerator half imgPlaneH
	} else {
		c.imgPlaneH = 2.0
		c.imgPlaneW = 2.0 * c.FilmWidth / c.FilmHeight
		c.imgPlaneD = (0.5 * c.imgPlaneW) / math.Tan(float64(fov/2)) // numerator = half imgPlaneW
	}
	c.unitPPW = c.imgPlaneW / c.FilmWidth
	c.unitPPH = c.imgPlaneH / c.FilmHeight

	return c
}

func (c *Camera) PixelToWorldCoord(col, row Float) V3 {
	// I have no idea why i have to negate camPt.X and camPt.Z in order to
	// make things work right.
	camPt := V3{
		-(c.imgPlaneW/2 - c.unitPPW*col),
		c.imgPlaneH/2 - c.unitPPH*row,
		-c.imgPlaneD}
	worldPt := mgl.TransformCoordinate(camPt, c.CamToWorld)
	return worldPt
}

// GetRay creates a ray from the camera into the scene through the final
// image pixel at (col, row).
func (c *Camera) GetRay(col, row Float) Ray {
	return Ray{
		Orig: c.Pos,
		Dir:  c.PixelToWorldCoord(col, row).Sub(c.Pos).Normalize()}
}
