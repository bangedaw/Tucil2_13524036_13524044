package main

import (
	"math"
)

type OctreeNode struct {
	Boundary BoundingBox
	Children [8]*OctreeNode
	IsLeaf   bool
	IsSolid  bool
}

func NewOctreeNode(boundary BoundingBox) *OctreeNode {
	return &OctreeNode{
		Boundary: boundary,
		IsLeaf:   true,
		IsSolid:  false,
	}
}

func Divide(bbox BoundingBox) [8]BoundingBox {
	var octants [8]BoundingBox

	cx := bbox.Min.X + (bbox.Max.X-bbox.Min.X)/2.0
	cy := bbox.Min.Y + (bbox.Max.Y-bbox.Min.Y)/2.0
	cz := bbox.Min.Z + (bbox.Max.Z-bbox.Min.Z)/2.0

	octants[0] = BoundingBox{Min: Vector3{bbox.Min.X, bbox.Min.Y, bbox.Min.Z},
		Max: Vector3{cx, cy, cz}}
	octants[1] = BoundingBox{Min: Vector3{cx, bbox.Min.Y, bbox.Min.Z},
		Max: Vector3{bbox.Max.X, cy, cz}}
	octants[2] = BoundingBox{Min: Vector3{bbox.Min.X, bbox.Min.Y, cz},
		Max: Vector3{cx, cy, bbox.Max.Z}}
	octants[3] = BoundingBox{Min: Vector3{cx, bbox.Min.Y, cz},
		Max: Vector3{bbox.Max.X, cy, bbox.Max.Z}}

	octants[4] = BoundingBox{Min: Vector3{bbox.Min.X, cy, bbox.Min.Z},
		Max: Vector3{cx, bbox.Max.Y, cz}}
	octants[5] = BoundingBox{Min: Vector3{cx, cy, bbox.Min.Z},
		Max: Vector3{bbox.Max.X, bbox.Max.Y, cz}}
	octants[6] = BoundingBox{Min: Vector3{bbox.Min.X, cy, cz},
		Max: Vector3{cx, bbox.Max.Y, bbox.Max.Z}}
	octants[7] = BoundingBox{Min: Vector3{cx, cy, cz},
		Max: Vector3{bbox.Max.X, bbox.Max.Y, bbox.Max.Z}}

	return octants
}

func IsIntersecting(box BoundingBox, v0, v1, v2 Vector3) bool {
	triMinX := math.Min(v0.X, math.Min(v1.X, v2.X))
	triMinY := math.Min(v0.Y, math.Min(v1.Y, v2.Y))
	triMinZ := math.Min(v0.Z, math.Min(v1.Z, v2.Z))

	triMaxX := math.Max(v0.X, math.Max(v1.X, v2.X))
	triMaxY := math.Max(v0.Y, math.Max(v1.Y, v2.Y))
	triMaxZ := math.Max(v0.Z, math.Max(v1.Z, v2.Z))

	if triMaxX < box.Min.X || triMinX > box.Max.X {
		return false
	}
	if triMaxY < box.Min.Y || triMinY > box.Max.Y {
		return false
	}
	if triMaxZ < box.Min.Z || triMinZ > box.Max.Z {
		return false
	}

	return true
}

// separating axis theorem for avoiding false positive

func Sub(a, b Vector3) Vector3 {
	return Vector3{a.X - b.X, a.Y - b.Y, a.Z - b.Z}
}

func CrossProduct(a, b Vector3) Vector3 {
	return Vector3{
		a.Y*b.Z - a.Z*b.Y,
		a.Z*b.X - a.X*b.Z,
		a.X*b.Y - a.Y*b.X,
	}
}

func DotProduct(a, b Vector3) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func IsIntersectingSAT(box BoundingBox, v0, v1, v2 Vector3) bool {
	c := Vector3{
		X: (box.Min.X + box.Max.X) / 2.0,
		Y: (box.Min.Y + box.Max.Y) / 2.0,
		Z: (box.Min.Z + box.Max.Z) / 2.0,
	}

	e := Vector3{
		X: (box.Max.X - box.Min.X) / 2.0,
		Y: (box.Max.Y - box.Min.Y) / 2.0,
		Z: (box.Max.Z - box.Min.Z) / 2.0,
	}

	v0 = Sub(v0, c)
	v1 = Sub(v1, c)
	v2 = Sub(v2, c)

	f0 := Sub(v1, v0)
	f1 := Sub(v2, v1)
	f2 := Sub(v0, v2)

	if !axisTest(0, -f0.Z, f0.Y, v0, v1, v2, e) {
		return false
	}
	if !axisTest(0, -f1.Z, f1.Y, v0, v1, v2, e) {
		return false
	}
	if !axisTest(0, -f2.Z, f2.Y, v0, v1, v2, e) {
		return false
	}

	if !axisTest(0, f0.Z, -f0.X, v0, v1, v2, e) {
		return false
	}
	if !axisTest(0, f1.Z, -f1.X, v0, v1, v2, e) {
		return false
	}
	if !axisTest(0, f2.Z, -f2.X, v0, v1, v2, e) {
		return false
	}

	if !axisTest(0, -f0.Y, f0.X, v0, v1, v2, e) {
		return false
	}
	if !axisTest(0, -f1.Y, f1.X, v0, v1, v2, e) {
		return false
	}
	if !axisTest(0, -f2.Y, f2.X, v0, v1, v2, e) {
		return false
	}

	if math.Min(v0.X, math.Min(v1.X, v2.X)) > e.X || math.Max(v0.X, math.Max(v1.X, v2.X)) < -e.X {
		return false
	}
	if math.Min(v0.Y, math.Min(v1.Y, v2.Y)) > e.Y || math.Max(v0.Y, math.Max(v1.Y, v2.Y)) < -e.Y {
		return false
	}
	if math.Min(v0.Z, math.Min(v1.Z, v2.Z)) > e.Z || math.Max(v0.Z, math.Max(v1.Z, v2.Z)) < -e.Z {
		return false
	}

	normal := CrossProduct(f0, f1)
	planeDistance := DotProduct(normal, v0)
	return planeBoxOverlap(normal, planeDistance, e)
}

func axisTest(aX, aY, aZ float64, v0, v1, v2, e Vector3) bool {
	axis := Vector3{aX, aY, aZ}

	p0 := DotProduct(v0, axis)
	p1 := DotProduct(v1, axis)
	p2 := DotProduct(v2, axis)

	min := math.Min(p0, math.Min(p1, p2))
	max := math.Max(p0, math.Max(p1, p2))

	r := e.X*math.Abs(aX) + e.Y*math.Abs(aY) + e.Z*math.Abs(aZ)

	if min > r || max < -r {
		return false
	}
	return true
}

func planeBoxOverlap(normal Vector3, d float64, e Vector3) bool {
	var vmin, vmax Vector3

	if normal.X > 0 {
		vmin.X = -e.X
		vmax.X = e.X
	} else {
		vmin.X = e.X
		vmax.X = -e.X
	}

	if normal.Y > 0 {
		vmin.Y = -e.Y
		vmax.Y = e.Y
	} else {
		vmin.Y = e.Y
		vmax.Y = -e.Y
	}

	if normal.Z > 0 {
		vmin.Z = -e.Z
		vmax.Z = e.Z
	} else {
		vmin.Z = e.Z
		vmax.Z = -e.Z
	}

	if DotProduct(normal, vmin) > d {
		return false
	}
	if DotProduct(normal, vmax) >= d {
		return true
	}

	return false
}
