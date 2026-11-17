package estree

import (
	"context"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Projection is syntax owned by one file client. Admission and generation are
// supplied separately by its semantic-session binding, never inferred here.
type Projection struct {
	SourceFileName string `json:"sourceFileName"`
	Frame          []byte `json:"frame"`
	Tree           []byte `json:"tree"`
	OriginTables
}
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

func (b *Builder) Project(ctx context.Context, file *ast.SourceFile, options Options) (Projection, error) {
	tree, table, err := encoder.EncodeSourceFile(file)
	if err != nil {
		return Projection{}, err
	}
	origins := newOriginTables(file)
	frame, err := b.Build(ctx, file, origins.record(options, table))
	return Projection{SourceFileName: file.FileName(), Frame: frame, Tree: tree, OriginTables: origins}, err
}

// ProjectOrigins reproduces the immutable frame's node IDs without scanning
// tokens or writing another frame.
func (b *Builder) ProjectOrigins(ctx context.Context, file *ast.SourceFile, options Options, table *encoder.NodeIndexTable) (OriginTables, error) {
	origins := newOriginTables(file)
	_, err := b.build(ctx, file, origins.record(options, table), nil, true)
	return origins, err
}

func newOriginTables(file *ast.SourceFile) OriginTables {
	return OriginTables{Origins: make([]uint32, 1, min(file.NodeCount+1, maxInitialFrameBytes/4)), Reverse: make(map[uint32]uint32)}
}

func (result *OriginTables) record(options Options, table *encoder.NodeIndexTable) Options {
	options.RecordOrigin = func(id uint32, node *ast.Node) {
		if int(id) != len(result.Origins) {
			panic("unordered projection origins")
		}
		var index uint32
		if node != nil {
			index = table.GetIndex(node)
		}
		result.Origins = append(result.Origins, index)
	}
	options.RecordReverse = func(node *ast.Node, id uint32) {
		if index := table.GetIndex(node); index != 0 {
			result.Reverse[index] = id
		}
	}
	options.RecordSequenceAlias = func(node *ast.Node, sequence, count, start, end uint32) {
		result.Aliases = append(result.Aliases, SequenceAlias{table.GetIndex(node), sequence, count, start, end})
	}
	options.sequenceMembers = func(sequenceID uint32, members []uint32) {
		if result.SequenceMembers == nil {
			result.SequenceMembers = make(map[uint32][]uint32)
		}
		result.SequenceMembers[sequenceID] = members
	}
	options.recordChainAlias = func(node, origin *ast.Node, expression, start, end uint32) {
		if index := table.GetIndex(node); index != 0 {
			result.ChainAliases = append(result.ChainAliases, ChainAlias{index, table.GetIndex(origin), expression, start, end})
		}
	}
	return options
}
