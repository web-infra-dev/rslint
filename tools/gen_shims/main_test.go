package main

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"
)

type mirrorImporter struct{ source *types.Package }

func (i mirrorImporter) Import(path string) (*types.Package, error) {
	if path == "source" {
		return i.source, nil
	}
	return importer.Default().Import(path)
}

func TestMirrorPrivateTypes(t *testing.T) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "source.go", `package source
type kind uint8
type recursive struct { next *recursive; value *Exported }
type Store[T any] struct { count int; values []T }
type Exported struct {
 kind kind
 pointer *recursive
 store Store[kind]
 values [3]kind
 lookup map[kind]*recursive
}
`, 0)
	if err != nil {
		t.Fatal(err)
	}
	source, err := new(types.Config).Check("source", fileSet, []*ast.File{file}, nil)
	if err != nil {
		t.Fatal(err)
	}
	original := source.Scope().Lookup("Exported").Type().Underlying().(*types.Struct)
	text := mirrorTypeString(original, func(p *types.Package) string { return p.Name() })
	if strings.Contains(text, "source.kind") || strings.Contains(text, "source.recursive") ||
		!strings.Contains(text, "unsafe.Pointer") || !strings.Contains(text, "source.Store[uint8]") {
		t.Fatalf("private types were not mirrored: %s", text)
	}
	file, err = parser.ParseFile(fileSet, "mirror.go", "package mirror\nimport (\"source\"; \"unsafe\")\ntype Mirror "+text, 0)
	if err != nil {
		t.Fatal(err)
	}
	config := types.Config{Importer: mirrorImporter{source}}
	generated, err := config.Check("mirror", fileSet, []*ast.File{file}, nil)
	if err != nil {
		t.Fatal(err)
	}
	mirror := generated.Scope().Lookup("Mirror").Type().Underlying().(*types.Struct)
	for _, arch := range []string{"386", "amd64", "arm64"} {
		t.Run(arch, func(t *testing.T) {
			sizes := types.SizesFor("gc", arch)
			if sizes.Sizeof(original) != sizes.Sizeof(mirror) || sizes.Alignof(original) != sizes.Alignof(mirror) {
				t.Fatal("mirror size or alignment differs")
			}
			fields := func(s *types.Struct) []*types.Var {
				result := make([]*types.Var, s.NumFields())
				for i := range result {
					result[i] = s.Field(i)
				}
				return result
			}
			want, got := sizes.Offsetsof(fields(original)), sizes.Offsetsof(fields(mirror))
			for i := range want {
				if want[i] != got[i] || sizes.Sizeof(original.Field(i).Type()) != sizes.Sizeof(mirror.Field(i).Type()) {
					t.Errorf("field %s layout differs", original.Field(i).Name())
				}
			}
		})
	}
}
