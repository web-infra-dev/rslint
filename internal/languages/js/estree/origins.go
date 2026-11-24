package estree

import (
	"context"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

type OriginTables struct {
	Origins         []uint32            `json:"origins"`
	FirstTokens     map[uint32]uint32   `json:"firstTokens,omitempty"`
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
	chainRecord  uint32
	chainReverse []uint32
	originsOnly  bool
}

// BuildWithOrigins produces the ESTree frame and its node maps in one traversal.
func (b *Builder) BuildWithOrigins(ctx context.Context, file *ast.SourceFile, options Options, table *encoder.NodeIndexTable) ([]byte, OriginTables, error) {
	mapping := originProjection{table: table}
	frame, err := b.build(ctx, file, options, nil, &mapping)
	return frame, mapping.OriginTables, err
}

// ProjectOrigins reproduces the existing frame's node IDs without scanning
// tokens or encoding another frame.
func (b *Builder) ProjectOrigins(ctx context.Context, file *ast.SourceFile, options Options, table *encoder.NodeIndexTable) (OriginTables, error) {
	mapping := originProjection{table: table, originsOnly: true}
	if _, err := b.build(ctx, file, options, nil, &mapping); err != nil {
		return OriginTables{}, err
	}
	return mapping.OriginTables, nil
}
