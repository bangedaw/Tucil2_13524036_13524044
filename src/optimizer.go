package main

import (
	"math"
)

// GridKey merepresentasikan koordinat integer (ix, iy, iz) dari voxel dalam grid seragam
type GridKey struct {
	X, Y, Z int
}

// OptimizationMap menyimpan data voxel dalam bentuk grid untuk lookup tetangga yang cepat
type OptimizationMap struct {
	VoxelGrid map[GridKey]bool
}

// NewOptimizationMap menginisialisasi GridMap dari daftar BoundingBox solid voxels
func NewOptimizationMap(voxels []BoundingBox, rootMin Vector3, voxelSize float64) *OptimizationMap {
	grid := make(map[GridKey]bool)

	for _, v := range voxels {
		// Konversi koordinat minimum voxel (float) menjadi koordinat integer grid.
		// Gunakan pembulatan presisi untuk menghindari floating point error.
		ix := int(math.Round((v.Min.X - rootMin.X) / voxelSize))
		iy := int(math.Round((v.Min.Y - rootMin.Y) / voxelSize))
		iz := int(math.Round((v.Min.Z - rootMin.Z) / voxelSize))

		grid[GridKey{X: ix, Y: iy, Z: iz}] = true
	}

	return &OptimizationMap{VoxelGrid: grid}
}

// IsInternalFace mengecek apakah sisi kubus pada arah tertentu adalah sisi internal
func (om *OptimizationMap) IsInternalFace(voxelGridX, voxelGridY, voxelGridZ int, direction string) bool {
	// Tentukan target koordinat tetangga berdasarkan arah yang dicek
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

	// Cek apakah ada voxel solid di koordinat target
	return om.VoxelGrid[GridKey{X: targetX, Y: targetY, Z: targetZ}]
}
