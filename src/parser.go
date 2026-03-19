package main

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
)

type Vector3 struct {
	X, Y, Z float64
}

type Face struct {
	V [3]int
}

type BoundingBox struct {
	Min Vector3
	Max Vector3
}

func ParseOBJ(filePath string) ([]Vector3, []Face, BoundingBox, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, BoundingBox{}, err

	}
	defer file.Close()

	var vertices []Vector3
	var faces []Face

	bbox := BoundingBox{
		Min: Vector3{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64},
		Max: Vector3{-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64},
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "v":
			if len(parts) >= 4 {
				x, _ := strconv.ParseFloat(parts[1], 64)
				y, _ := strconv.ParseFloat(parts[2], 64)
				z, _ := strconv.ParseFloat(parts[3], 64)
				v := Vector3{X: x, Y: y, Z: z}
				vertices = append(vertices, v)

				bbox.Min.X = math.Min(bbox.Min.X, x)
				bbox.Min.Y = math.Min(bbox.Min.Y, y)
				bbox.Min.Z = math.Min(bbox.Min.Z, z)
				bbox.Max.X = math.Max(bbox.Min.X, x)
				bbox.Max.Y = math.Max(bbox.Min.Y, y)
				bbox.Max.Z = math.Max(bbox.Min.Z, z)
			}
		case "f":
			if len(parts) >= 4 {
				v1, _ := strconv.Atoi(strings.Split(parts[1], "/")[0])
				v2, _ := strconv.Atoi(strings.Split(parts[2], "/")[0])
				v3, _ := strconv.Atoi(strings.Split(parts[3], "/")[0])

				faces = append(faces, Face{V: [3]int{v1 - 1, v2 - 1, v3 - 1}})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, BoundingBox{}, err
	}

	bbox = MakeCubicBoundingBox(bbox)

	return vertices, faces, bbox, nil
}

func MakeCubicBoundingBox(bbox BoundingBox) BoundingBox {
	dx := bbox.Max.X - bbox.Min.X
	dy := bbox.Max.Y - bbox.Min.Y
	dz := bbox.Max.Z - bbox.Min.Z

	maxDim := math.Max(dx, math.Max(dy, dz))

	centerX := bbox.Min.X + dx/2.0
	centerY := bbox.Min.Y + dy/2.0
	centerZ := bbox.Min.Z + dz/2.0

	halfDim := maxDim / 2.0

	return BoundingBox{
		Min: Vector3{centerX - halfDim, centerY - halfDim, centerZ - halfDim},
		Max: Vector3{centerX + halfDim, centerY + halfDim, centerZ + halfDim},
	}
}
