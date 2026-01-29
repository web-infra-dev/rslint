package inspector

import (
	"unsafe"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// BuildFlowInfoWithDepth builds FlowInfo with a depth limit to avoid infinite recursion
// Flow graphs can have cycles (loops), so we need to limit how deep we go
func (b *Builder) BuildFlowInfoWithDepth(flow *ast.FlowNode, depth int) *FlowInfo {
	if flow == nil || depth <= 0 {
		return nil
	}
	info := &FlowInfo{
		Flags:     uint32(flow.Flags),
		FlagNames: GetFlowFlagNames(flow.Flags),
	}
	// Build shallow node info if present
	if flow.Node != nil {
		info.Node = b.BuildShallowNodeInfo(flow.Node)
	}
	// Include antecedent/antecedents with reduced depth
	if depth > 1 {
		if flow.Antecedent != nil {
			info.Antecedent = b.BuildFlowInfoWithDepth(flow.Antecedent, depth-1)
		}
		if flow.Antecedents != nil {
			info.Antecedents = make([]*FlowInfo, 0)
			for list := flow.Antecedents; list != nil; list = list.Next {
				if list.Flow != nil {
					info.Antecedents = append(info.Antecedents, b.BuildFlowInfoWithDepth(list.Flow, depth-1))
				}
			}
		}
	}
	return info
}

// BuildShallowFlowInfo builds minimal FlowInfo for nested flow nodes (depth=1, no antecedents)
func (b *Builder) BuildShallowFlowInfo(flow *ast.FlowNode) *FlowInfo {
	return b.BuildFlowInfoWithDepth(flow, 1)
}

// BuildFlowInfo builds FlowInfo from a FlowNode
// Uses depth=4 for the top-level flow to show a reasonable chain while avoiding cycles
// Also builds the complete flow graph for visualization
func (b *Builder) BuildFlowInfo(flow *ast.FlowNode) *FlowInfo {
	// Build with depth limit of 4 (top level + 3 nested levels)
	info := b.BuildFlowInfoWithDepth(flow, 4)
	if info != nil {
		// Build the complete graph and attach to top-level
		info.Graph = b.BuildFlowGraph(flow)
	}
	return info
}

// flowNodeId returns a unique ID for a FlowNode using its pointer address
func flowNodeId(flow *ast.FlowNode) uint64 {
	return uint64(uintptr(unsafe.Pointer(flow)))
}

// BuildFlowGraph builds the complete flow graph by traversing all nodes
func (b *Builder) BuildFlowGraph(flow *ast.FlowNode) *FlowGraph {
	if flow == nil {
		return nil
	}

	visited := make(map[*ast.FlowNode]bool)
	nodes := make([]*FlowGraphNode, 0)
	edges := make([]*FlowEdge, 0)

	// BFS to traverse all nodes
	queue := []*ast.FlowNode{flow}
	visited[flow] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Build node info
		node := &FlowGraphNode{
			Id:        flowNodeId(current),
			Flags:     uint32(current.Flags),
			FlagNames: GetFlowFlagNames(current.Flags),
		}
		if current.Node != nil {
			node.NodePos = current.Node.Pos()
			node.NodeEnd = current.Node.End()
			node.NodeKindName = current.Node.Kind.String()
			// Extract source text directly from SourceFile (works for all node types)
			node.NodeText = scanner.GetSourceTextOfNodeFromSourceFile(b.sourceFile, current.Node, false)
		}
		nodes = append(nodes, node)

		// Process antecedent
		if current.Antecedent != nil {
			edges = append(edges, &FlowEdge{
				From: flowNodeId(current.Antecedent),
				To:   flowNodeId(current),
			})
			if !visited[current.Antecedent] {
				visited[current.Antecedent] = true
				queue = append(queue, current.Antecedent)
			}
		}

		// Process antecedents (for labels)
		if current.Antecedents != nil {
			for list := current.Antecedents; list != nil; list = list.Next {
				if list.Flow != nil {
					edges = append(edges, &FlowEdge{
						From: flowNodeId(list.Flow),
						To:   flowNodeId(current),
					})
					if !visited[list.Flow] {
						visited[list.Flow] = true
						queue = append(queue, list.Flow)
					}
				}
			}
		}
	}

	return &FlowGraph{
		Nodes: nodes,
		Edges: edges,
	}
}
