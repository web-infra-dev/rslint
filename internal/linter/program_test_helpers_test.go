package linter

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
)

func mustSourceOnlyTestProgram(t testing.TB, typeScript *compiler.Program, files []*ast.SourceFile) *lintprogram.Program {
	t.Helper()
	sourceProgram, err := lintprogram.NewFromBoundSources(typeScript, files)
	if err != nil {
		t.Fatalf("NewFromBoundSources: %v", err)
	}
	return sourceProgram
}

func wrapTestPrograms(programs ...*compiler.Program) []*lintprogram.Program {
	return lintprogram.NewFromCompilers(programs)
}

func testPrograms(programs ...*lintprogram.Program) []*lintprogram.Program {
	return programs
}
