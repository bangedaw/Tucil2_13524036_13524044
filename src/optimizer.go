package main

import (
	"math"
)

type GridKey struct {
	X, Y, Z int
}

type OptimizationMap struct {
	VoxelGrid map[GridKey]bool
}

func NewOptimizationMap(voxels []BoundingBox, rootMin Vector3, voxelSize float64) *OptimizationMap {
	grid := make(map[GridKey]bool)

	for _, v := range voxels {
		ix := int(math.Round((v.Min.X - rootMin.X) / voxelSize))
		iy := int(math.Round((v.Min.Y - rootMin.Y) / voxelSize))
		iz := int(math.Round((v.Min.Z - rootMin.Z) / voxelSize))

		grid[GridKey{X: ix, Y: iy, Z: iz}] = true
	}

	return &OptimizationMap{VoxelGrid: grid}
}

func (om *OptimizationMap) IsInternalFace(voxelGridX, voxelGridY, voxelGridZ int, direction string) bool {
	targetX, targetY, targetZ := voxelGridX, voxelGridY, voxelGridZ
	switch direction {
	case "Depan":
		targetZ++
	case "Belakang":
		targetZ--
	case "Kanan":
		targetX++
	case "Kiri":
		targetX--
	case "Atas":
		targetY++
	case "Bawah":
		targetY--
	}

	return om.VoxelGrid[GridKey{X: targetX, Y: targetY, Z: targetZ}]
}
