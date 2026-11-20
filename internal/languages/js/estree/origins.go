package estree

import (
	"context"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

type OriginTables struct {
	Origins         []uint32            `json:"origins"`
	Reverse         map[uint32]uint32   `json:"reverse"`
	Aliases         []SequenceAlias     `json:"aliases"`
	SequenceMembers map[uint32][]uint32 `json:"sequenceMembers,omitempty"`
	ChainAliases    []ChainAlias        `json:"chainAliases,omitempty"`
}
type SequenceAlias struct {
	Index    uint32 `json:"index"`
	Sequence uint32 `json:"sequence"`
	Count    uint32 `json:"count"`
	Start    uint32 `json:"start"`
	End      uint32 `json:"end"`
}

type ChainAlias struct {
	Index      uint32 `json:"index"`
	Origin     uint32 `json:"origin"`
	Expression uint32 `json:"expression"`
	Start      uint32 `json:"start"`
	End        uint32 `json:"end"`
}

type originProjection struct {
	OriginTables
	table        *encoder.NodeIndexTable
	nodes        []*ast.Node
	reverse      map[*ast.Node]uint32
	chainRecord  uint32
	chainReverse []*ast.Node
}

// ProjectOrigins reproduces the existing frame's node IDs without scanning
// tokens or encoding another frame.
func (b *Builder) ProjectOrigins(ctx context.Context, file *ast.SourceFile, options Options, table *encoder.NodeIndexTable) (OriginTables, error) {
	mapping := originProjection{table: table, reverse: make(map[*ast.Node]uint32)}
	if _, err := b.build(ctx, file, options, nil, &mapping); err != nil {
		return OriginTables{}, err
	}
	return mapping.OriginTables, nil
}

func (m *originProjection) finish() {
	// Projected IDs start at one. Synthetic nodes have no compiler index.
	m.Origins = make([]uint32, len(m.nodes)+1)
	for i, node := range m.nodes {
		if node != nil {
			m.Origins[i+1] = m.table.GetIndex(node)
		}
	}
	m.Reverse = make(map[uint32]uint32, len(m.reverse))
	for node, id := range m.reverse {
		if id == detachedChainReverse {
			continue
		}
		if index := m.table.GetIndex(node); index != 0 {
			m.Reverse[index] = id
		}
	}
}
