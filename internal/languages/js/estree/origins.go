package estree

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Projection is syntax owned by one file client. Admission and generation are
// supplied separately by its semantic-session binding, never inferred here.
type Projection struct {
	SourceFileName string            `json:"sourceFileName"`
	Frame          []byte            `json:"frame"`
	Tree           []byte            `json:"tree"`
	Origins        []uint32          `json:"origins"`
	Reverse        map[uint32]uint32 `json:"reverse"`
	Aliases        []SequenceAlias   `json:"aliases"`
}
type SequenceAlias struct {
	Index    uint32 `json:"index"`
	Sequence uint32 `json:"sequence"`
	Count    uint32 `json:"count"`
	Start    uint32 `json:"start"`
	End      uint32 `json:"end"`
}

func (b *Builder) Project(ctx context.Context, file *ast.SourceFile, options Options) (Projection, error) {
	tree, table, err := encoder.EncodeSourceFile(file)
	if err != nil {
		return Projection{}, err
	}
	result := Projection{SourceFileName: file.FileName(), Tree: tree, Origins: []uint32{0}, Reverse: make(map[uint32]uint32)}
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
	result.Frame, err = b.Build(ctx, file, options)
	return result, err
}
