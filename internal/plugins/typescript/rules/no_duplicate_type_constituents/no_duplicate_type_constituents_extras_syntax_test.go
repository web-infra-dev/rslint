package no_duplicate_type_constituents

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func parseSyntaxMembers(code string) (*ast.SourceFile, []*ast.Node) {
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/syntax.ts", Path: "/syntax.ts"}, code, core.ScriptKindTS)
	return source, source.Statements.Nodes[0].AsTypeAliasDeclaration().Type.AsUnionTypeNode().Types.Nodes
}

func TestNoDuplicateTypeConstituentsSyntax(t *testing.T) {
	for _, test := range []struct {
		left, right string
		equal       bool
	}{
		{"{ [K in 'a' as 'b'] }", "{ [K in 'a']: 'b' }", false},
		{"{ [K in 'a' as 'b']: 'c' }", "{ [K in 'a' as 'b']: 'c' }", true},
		{"{ [K in 'a']: 'b' }", "{ [K in 'a' as 'b']: 'a' }", false},
		{"(<T extends string>() => T)", "(<T = string>() => T)", false},
		{"(<T extends string = 'a'>() => T)", "(<T extends string = 'a'>() => T)", true},
		{"((x: string) => void)", "((x?: string) => void)", false},
		{"((x: string) => void)", "((y: string) => void)", false},
		{"(new () => void)", "(abstract new () => void)", false},
		{"{ a?: string }", "{ a: string }", false},
		{"[a?: string]", "[a: string]", false},
		{"readonly string[]", "string[]", false},
		{"{a: typeof import('x').T}", "{a: import('x').T}", false},
		{"{a: import('x').T<A>}", "{a: import('x').T<A, B>}", false},
		{"{a: 1}", "{a: 0x1}", false},
		{"{a: 'x'}", "{a: \"x\"}", false},
		{"{a: `x${string}y`}", "{a: `x${number}y`}", false},
		{"{a: `x${string}y`}", "{a: `x${string}y`}", true},
		{"{a: string}", "{ \\u0061: /*comment*/ (string); }", true},
		{"{a: string; b: number}", "{a: string, b: number,}", true},
		{"{a: string; b: number}", "{b: number; a: string}", false},
		{"(string | number)", "((string | number))", true},
	} {
		t.Run(test.left+" vs "+test.right, func(t *testing.T) {
			source, members := parseSyntaxMembers("type T = " + test.left + " | " + test.right + ";")
			index := newTypeSyntaxIndex(source)
			if got := index.equal(members[0], members[1]); got != test.equal {
				t.Fatalf("equality = %v, want %v", got, test.equal)
			}
			if test.equal && index.hash(members[0]) != index.hash(members[1]) {
				t.Fatal("equal syntax has different hashes")
			}
			// Force a collision in the cache: finding candidates must still use
			// exact field-aware equality, rather than trusting the fingerprint.
			index.hashes[syntaxNode(members[0])] = 1
			index.hashes[syntaxNode(members[1])] = 1
			if got := index.find([]*ast.Node{members[0]}, members[1]); (got != nil) != test.equal {
				t.Fatal("hash collision changed equality")
			}
		})
	}
}

func TestNoDuplicateTypeConstituentsSyntaxCandidateOrder(t *testing.T) {
	source, members := parseSyntaxMembers("type T = {a: string} | {a: (string)} | {a: string;};")
	index := newTypeSyntaxIndex(source)
	if got := index.find(members[:2], members[2]); got != members[0] {
		t.Fatal("previous must refer to the first equivalent constituent")
	}
}

func syntaxBenchmarkSource(count int, shape string) string {
	var code strings.Builder
	code.WriteString("type T = ")
	for i := range count {
		if i != 0 {
			code.WriteString(" | ")
		}
		switch shape {
		case "literal":
			fmt.Fprintf(&code, "'member%d'", i)
		case "object":
			fmt.Fprintf(&code, "{a: {b: 'member%d'}}", i)
		case "duplicate":
			code.WriteString("{a: {b: string}}")
		}
	}
	code.WriteByte(';')
	return code.String()
}

func BenchmarkNoDuplicateTypeConstituentsSyntaxIndex(b *testing.B) {
	for _, shape := range []string{"literal", "object", "duplicate"} {
		for _, count := range []int{100, 500, 2500} {
			b.Run(fmt.Sprintf("%s/%d", shape, count), func(b *testing.B) {
				source, members := parseSyntaxMembers(syntaxBenchmarkSource(count, shape))
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					index := newTypeSyntaxIndex(source)
					buckets := make(map[uint64][]*ast.Node)
					for _, member := range members {
						hash := index.hash(member)
						if index.find(buckets[hash], member) == nil {
							buckets[hash] = append(buckets[hash], member)
						}
					}
				}
			})
		}
	}
}
