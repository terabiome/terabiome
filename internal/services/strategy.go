package services

import yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"

type DepthFirstStrategy struct{}

func (s *DepthFirstStrategy) BuildBatches(uow yamlcontracts.UnitOfWork) [][]yamlcontracts.UnitOfWorkSequence {
	graph := uow.GetSequenceDependencyGraph()
	return groupByDepthUsingGraph(uow.Sequences, graph.FirstDegreeGraph)
}

// groupByDepthUsingGraph uses the pre-validated adjacency list to find max depth.
func groupByDepthUsingGraph(sequences []yamlcontracts.UnitOfWorkSequence, graph map[string]map[string]struct{}) [][]yamlcontracts.UnitOfWorkSequence {
	if len(sequences) == 0 {
		return nil
	}

	nameToIdx := make(map[string]int)
	for i, seq := range sequences {
		nameToIdx[seq.Name] = i
	}

	depths := make([]int, len(sequences))
	for i, seq := range sequences {
		// Use the graph to find dependencies instead of raw DependsOn slice
		depths[i] = calculateDepthFromGraph(seq.Name, graph, nameToIdx, make(map[string]bool))
	}

	maxDepth := 0
	for _, d := range depths {
		if d > maxDepth {
			maxDepth = d
		}
	}

	batches := make([][]yamlcontracts.UnitOfWorkSequence, maxDepth+1)
	for i, seq := range sequences {
		batches[depths[i]] = append(batches[depths[i]], seq)
	}

	return batches
}

func calculateDepthFromGraph(name string, graph map[string]map[string]struct{}, nameToIdx map[string]int, visited map[string]bool) int {
	if visited[name] {
		return 0
	}
	visited[name] = true

	deps, exists := graph[name]
	if !exists || len(deps) == 0 {
		return 0
	}

	maxParentDepth := -1
	for depName := range deps {
		parentDepth := calculateDepthFromGraph(depName, graph, nameToIdx, visited)
		if parentDepth > maxParentDepth {
			maxParentDepth = parentDepth
		}
	}

	if maxParentDepth == -1 {
		return 0
	}
	return maxParentDepth + 1
}
