package graph

import "strings"

type Node struct {
	ID         string
	Label      string
	Properties map[string]any
}

type Edge struct {
	Source       Node
	Target       Node
	Relationship string
	Properties   map[string]any
}

// NodeID returns a canonical, deterministic identifier for a node.
// All analyzers must use this function to guarantee cross-analyzer node identity:
// if Oracle creates PROCEDURE:calculate_tax and Java references the same procedure,
// both produce the same ID and Neo4j MERGE collapses them into one node.
func NodeID(label, name string) string {
	return strings.ToUpper(label) + ":" + strings.ToLower(name)
}

func NewNode(label, name string, props map[string]any) Node {
	if props == nil {
		props = make(map[string]any)
	}
	label = strings.ToUpper(label)
	props["name"] = name // always stored so the web UI can display clean names
	return Node{
		ID:         NodeID(label, name),
		Label:      label,
		Properties: props,
	}
}

func NewEdge(source, target Node, relationship string, props map[string]any) Edge {
	if props == nil {
		props = make(map[string]any)
	}
	return Edge{
		Source:       source,
		Target:       target,
		Relationship: relationship,
		Properties:   props,
	}
}
