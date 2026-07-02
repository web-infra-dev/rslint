package linter

import (
	"sync"

	"github.com/microsoft/typescript-go/shim/core"
)

// textSourceFile is a lightweight ast.SourceFileLike backed only by raw
// source text — no AST, scope, or types. It lets diagnostics produced
// outside ts-go (ESLint-plugin rules run in a Node worker) render
// line/column through the scanner, which only needs Text() + ECMALineMap().
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
