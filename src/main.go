package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	inputFile := "../test/input/cow.obj"
	outputFile := "../test/output/cow-voxelized.obj"
	maxDepth := 5

	startTime := time.Now()

	vertices, faces, rootBoundary, err := ParseOBJ(inputFile)
	if err != nil {
		log.Fatalf("Gagal membaca file .obj: %v", err)
	}

	rootNode := NewOctreeNode(rootBoundary)
	stats := NewOctreeStats(maxDepth)

	var wg sync.WaitGroup
	wg.Add(1)

	BuildOctreeConcurrent(rootNode, vertices, faces, 0, maxDepth, stats, &wg)

	wg.Wait()

	var solidVoxels []BoundingBox
	CollectVoxels(rootNode, &solidVoxels)

	numVertices, numFaces, err := ExportToObj(solidVoxels, outputFile)
	if err != nil {
		log.Fatalf("Gagal menulis file .obj: %v", err)
	}

	executionTime := time.Since(startTime)

	PrintReport(stats, len(solidVoxels), numVertices, numFaces, maxDepth, executionTime, outputFile)

	fmt.Println("\nMembuka Interactive Viewer di http://localhost:8080 ...")
	if err := ObjtoModel(outputFile); err != nil {
		log.Fatalf("Gagal memuat model untuk viewer: %v", err)
	}

	http.HandleFunc("/", handleIndex)

	port := ":8080"
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Viewer Error: %v\n", err)
	}
}
