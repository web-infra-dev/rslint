package utils

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func BenchmarkForEachComment(b *testing.B) {
	for _, fixture := range []struct {
		name, source string
		comments     int
	}{
		{"statements", strings.Repeat("let value = 1 /* comment */;\n", 512), 512},
		{"wide", "const values = [" + strings.Repeat("0 /* comment */,", 4096) + "];", 4096},
		{"nested", "const value = " + strings.Repeat("(", 512) + "1 /* comment */" + strings.Repeat(")", 512) + ";", 1},
	} {
		b.Run(fixture.name, func(b *testing.B) {
			parse := func() *ast.SourceFile {
				return parser.ParseSourceFile(ast.SourceFileParseOptions{
					FileName: "/comments.ts", Path: "/comments.ts",
				}, fixture.source, core.ScriptKindTS)
			}
			for _, warm := range []bool{false, true} {
				name := "cold"
				if warm {
					name = "warm"
				}
				b.Run(name, func(b *testing.B) {
					sourceFile := parse()
					if warm {
						ForEachToken(sourceFile.AsNode(), func(*ast.Node) {}, sourceFile)
					}
					b.ReportAllocs()
					for b.Loop() {
						if !warm {
							// Measure first collection without including parsing, and do
							// not let a previous iteration populate this file's cache.
							b.StopTimer()
							sourceFile = parse()
							b.StartTimer()
						}
						comments := 0
						ForEachComment(sourceFile.AsNode(), func(*ast.CommentRange) { comments++ }, sourceFile)
						if comments != fixture.comments {
							b.Fatalf("comments = %d, want %d", comments, fixture.comments)
						}
					}
				})
			}
		})
	}
}
