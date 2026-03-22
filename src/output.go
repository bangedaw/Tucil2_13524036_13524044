package main

import (
	"fmt"
	"os"
	"time"
)

func CollectVoxels(node *OctreeNode, voxels *[]BoundingBox) {
	if node == nil {
		return
	}

	if node.IsLeaf {
		if node.IsSolid {
			*voxels = append(*voxels, node.Boundary)
		}
		return
	}

	for i := 0; i < 8; i++ {
		CollectVoxels(node.Children[i], voxels)
	}
}

func ExportToObj(voxels []BoundingBox, outputPath string) (int, int, error) {
	file, err := os.Create(outputPath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	vertexCount := 0
	faceCount := 0

	for _, v := range voxels {
		fmt.Fprint(file, "v %f %f %f\n", v.Min.X, v.Min.Y, v.Min.Z)
		fmt.Fprint(file, "v %f %f %f\n", v.Max.X, v.Min.Y, v.Min.Z)
		fmt.Fprint(file, "v %f %f %f\n", v.Max.X, v.Min.Y, v.Max.Z)
		fmt.Fprint(file, "v %f %f %f\n", v.Min.X, v.Min.Y, v.Max.Z)

		fmt.Fprint(file, "v %f %f %f\n", v.Min.X, v.Max.Y, v.Min.Z)
		fmt.Fprint(file, "v %f %f %f\n", v.Max.X, v.Max.Y, v.Min.Z)
		fmt.Fprint(file, "v %f %f %f\n", v.Max.X, v.Max.Y, v.Max.Z)
		fmt.Fprint(file, "v %f %f %f\n", v.Min.X, v.Max.Y, v.Max.Z)

		offset := vertexCount + 1
		fmt.Fprint(file, "f %d %d %d %d\n", offset, offset+1, offset+2, offset+3)
		fmt.Fprint(file, "f %d %d %d %d\n", offset+4, offset+5, offset+6, offset+7)
		fmt.Fprint(file, "f %d %d %d %d\n", offset, offset+1, offset+5, offset+4)
		fmt.Fprint(file, "f %d %d %d %d\n", offset+1, offset+2, offset+6, offset+5)
		fmt.Fprint(file, "f %d %d %d %d\n", offset+2, offset+3, offset+7, offset+6)
		fmt.Fprint(file, "f %d %d %d %d\n", offset+3, offset, offset+4, offset+7)

		vertexCount += 8
		faceCount += 6
	}

	return vertexCount, faceCount, nil
}

func PrintReport(stats *OctreeStats, voxelsCreated, verticesCreated, facesCreated, maxDepth int, execTime time.Duration, outputPath string) {
	fmt.Println("==================================================")
	fmt.Println("==================================================")

	fmt.Printf("Banyaknya voxel yang terbentuk		: %d\n", voxelsCreated)
	fmt.Printf("Banyaknya vertex yang terbentuk		: %d\n", verticesCreated)
	fmt.Printf("Banyaknya faces yang terbentuk		: %d\n", facesCreated)
	fmt.Printf("Kedalaman octree					: %d\n", maxDepth)

	fmt.Println("\nStatistik node octree yang terbentuk:")
	for d := 1; d <= maxDepth; d++ {
		fmt.Printf("Depth %d: %d\n", d, stats.CreatedNodesByDepth[d])
	}

	fmt.Println("\nStatistik node yang tidak perlu ditelusuri:")
	for d := 1; d <= maxDepth; d++ {
		fmt.Printf("Depth %d: %d\n", d, stats.PrunedNodesByDepth[d])
	}

	fmt.Printf("\nLama waktu program berjalan	: %v\n", execTime)
	fmt.Printf("\nPath file .obj				: %v\n", outputPath)
	fmt.Println("==================================================")
	fmt.Println("==================================================")
}
