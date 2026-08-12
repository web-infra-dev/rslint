package linter

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
)

func mustStandaloneTestProgram(t testing.TB, typeScript *compiler.Program, files []*ast.SourceFile) *lintprogram.Program {
	t.Helper()
	standalone, err := lintprogram.NewStandaloneFromTypeScriptSources(typeScript, files)
	if err != nil {
		t.Fatalf("NewStandaloneFromTypeScriptSources: %v", err)
	}
	return standalone
}

func wrapTestPrograms(programs ...*compiler.Program) []*lintprogram.Program {
	return lintprogram.WrapTypeScriptPrograms(programs)
}

func testPrograms(programs ...*lintprogram.Program) []*lintprogram.Program {
	return programs
}
