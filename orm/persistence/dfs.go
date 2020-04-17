package persistence

type color int

const (
	white color = iota
	gray
	black
)

func hasCycle(graph *PersistGraph) bool {
	rootNodes := graph.filterRoots()

	nodeColors := make(map[CompositeAction]color, len(graph.knownBoxes.flattActions()))

	for _, node := range rootNodes {
		if nodeColors[node] != white {
			continue
		}

		if findCycle(node, nodeColors) {
			return true
		}
	}

	return false
}

func findCycle(node CompositeAction, colors map[CompositeAction]color) bool {
	foundCycle := false
	colors[node] = gray

	for _, nodeI := range node.children() {
		if colors[nodeI] == gray {
			return true
		}

		foundCycle = foundCycle || findCycle(nodeI, colors)
	}

	colors[node] = black

	return foundCycle
}
