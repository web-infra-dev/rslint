package linter

import (
	"sync"

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

func (f *textSourceFile) Text() string { return f.text }

func (f *textSourceFile) ECMALineMap() []core.TextPos {
	f.lineOnce.Do(func() {
		f.lineMap = core.ComputeECMALineStarts(f.text)
	})
	return f.lineMap
}
