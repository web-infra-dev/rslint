package linter

import (
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
)

// textSourceFile is a lightweight ast.SourceFileLike backed only by raw
// source text — no AST, scope, or types. Plugin diagnostics and completed
// native observations use it to retain source frames independently of Program
// generations. Position conversion needs only Text() + ECMALineMap().
type textSourceFile struct {
	text     string
	lineOnce sync.Once
	lineMap  []core.TextPos
}

func newTextSourceFile(text string) *textSourceFile {
	return &textSourceFile{text: text}
}

// A source owns its text projection, with no reference back to the AST. This
// preserves exact-source identity across concurrent native/plugin producers
// without a long-lived map whose strong keys retain completed sources.
var diagnosticTextKey = ast.NewSourceFileDataKey[*textSourceFile]()

func diagnosticTextSource(source ast.SourceFileLike) ast.SourceFileLike {
	file, ok := source.(*ast.SourceFile)
	if !ok || file == nil {
		return source
	}
	return ast.GetOrComputeSourceFileData(file, diagnosticTextKey, func(file *ast.SourceFile) *textSourceFile {
		return newTextSourceFile(file.Text())
	})
}

func (f *textSourceFile) Text() string { return f.text }

func (f *textSourceFile) ECMALineMap() []core.TextPos {
	f.lineOnce.Do(func() {
		f.lineMap = core.ComputeECMALineStarts(f.text)
	})
	return f.lineMap
}
