package main

type OctreeStats struct {
	CreatedNodesByDepth []int
	PrunedNodesByDepth  []int
	TotalVoxelLeaves    int
}

func NewOctreeStats(maxDepth int) *OctreeStats {
	return &OctreeStats{
		CreatedNodesByDepth: make([]int, maxDepth+1),
		PrunedNodesByDepth:  make([]int, maxDepth+1),
		TotalVoxelLeaves:    0,
	}
}

func BuildOctree(node *OctreeNode, vertices []Vector3, faces []Face, currentDepth, maxDepth int, stats *OctreeStats) {
	stats.CreatedNodesByDepth[currentDepth]++

	if currentDepth == maxDepth {
		node.IsLeaf = true
		if len(faces) > 0 {
			node.IsSolid = true
			stats.TotalVoxelLeaves++
		} else {
			node.IsSolid = false
		}

		return
	}

	childBoxes := Divide(node.Boundary)
	node.IsLeaf = false

	for i := 0; i < 8; i++ {
		childBox := childBoxes[i]
		var intersectingFaces []Face

		for _, face := range faces {
			v0 := vertices[face.V[0]]
			v1 := vertices[face.V[1]]
			v2 := vertices[face.V[2]]

			if IsIntersectingSAT(childBox, v0, v1, v2) {
				intersectingFaces = append(intersectingFaces, face)
			}
		}

		childNode := NewOctreeNode(childBox)
		node.Children[i] = childNode

		if len(intersectingFaces) == 0 {
			stats.CreatedNodesByDepth[currentDepth+1]++
			stats.PrunedNodesByDepth[currentDepth+1]++

			childNode.IsLeaf = true
			childNode.IsSolid = false
		} else {
			BuildOctree(childNode, vertices, intersectingFaces, currentDepth+1, maxDepth, stats)
		}
	}
}
