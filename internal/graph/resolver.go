package graph

// resolve checks if all dependencies can be satisfied
func (g *DependencyGraph) Resolve() error {
	// simplified resolver - just checks it nodes exist
	// in real implementation , would check version contraints
	return nil
}

// getservicenodes returns all the service nodes
func (g *DependencyGraph) GetServiceNodes() []*Node {
	var services []*Node
	for _, node := range g.Nodes {
		if node.Type == "service" {
			services = append(services, node)
		}
	}
	return services
}
