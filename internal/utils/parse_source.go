package utils

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/vue/vuesfc"
)

// sourceForParse resolves the text and the script kind one file is parsed
// with. It is the single place where a file's bytes on disk and the bytes the
// TypeScript parser reads are allowed to differ.
//
// Every extension but .vue is its own parse input, read exactly the way the
// default compiler host reads it, which is the condition parse-cache invariant I6
// states.
//
// A .vue file is a Vue Single File Component, which is not a TypeScript source
// text at all. What the parser reads is [vuesfc.Extract]'s projection of its
// <script> blocks: the same byte length as the file, script content at its own
// offsets, everything else blank. The script kind comes from the blocks' `lang`
// attribute rather than from the filename, because the filename cannot carry it.
//
// Both the projection and the kind are deterministic functions of the file's
// text and name, so the parse-cache key, which carries a hash of that text
// plus the resulting script kind, still identifies exactly one parse input.
// Two callers that read the same .vue bytes therefore share one AST, and a
// caller that reads different bytes never sees another's.
func sourceForParse(opts ast.SourceFileParseOptions, text string) (string, core.ScriptKind) {
	if vuesfc.IsFile(opts.FileName) {
		component := vuesfc.Extract(text)
		return component.Text, component.ScriptKind
	}
	return text, core.GetScriptKindFromFileName(opts.FileName)
}

// sourceForParseArgs adapts [sourceForParse] to parser.ParseSourceFile's
// argument order, for the uncached host paths that parse in one expression.
func sourceForParseArgs(
	opts ast.SourceFileParseOptions,
	text string,
) (ast.SourceFileParseOptions, string, core.ScriptKind) {
	parseText, scriptKind := sourceForParse(opts, text)
	return opts, parseText, scriptKind
}

// TargetSourceText returns the text one lint target's ranges index into: the
// text an autofix is spliced into and that is finally written back to the file.
// A byte order mark is left to the caller, which knows whether its consumer
// wants the complete source or the parser's view of it.
//
// For every source the TypeScript parser reads directly, that is the parsed
// text itself.
//
// A Vue Single File Component is the one case where the two differ. What was
// parsed is a projection of its <script> blocks, with template and style bytes
// blanked (see [sourceForParse]); writing that back would erase the rest of
// the component. So the component's own text is read from the generation's
// filesystem instead: the overlay during a fix round, the disk before one.
//
// Substituting one text for the other is exact rather than approximate: the
// projection preserves every byte offset, so a range computed against the
// parsed text indexes the component's text identically.
func TargetSourceText(fileSystem vfs.FS, path string, source ast.SourceFileLike) (string, error) {
	if vuesfc.IsFile(path) {
		text, ok := fileSystem.ReadFile(path)
		if !ok {
			return "", fmt.Errorf("utils: could not read Vue component %q", path)
		}
		return text, nil
	}
	return source.Text(), nil
}
