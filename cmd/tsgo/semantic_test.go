package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
)

type semanticFixture struct {
	semantic     Semantic
	program      *compiler.Program
	sourceFile   *ast.SourceFile
	sourceFileID int
}

func buildSemanticFixture(t *testing.T, source string) semanticFixture {
	t.Helper()

	tmpDir := t.TempDir()
	tsconfigPath := filepath.Join(tmpDir, "tsconfig.json")
	sourcePath := filepath.Join(tmpDir, "index.ts")

	if err := os.WriteFile(tsconfigPath, []byte(`{"include":["./index.ts"]}`), 0o644); err != nil {
		t.Fatalf("Failed to write tsconfig: %v", err)
	}
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("Failed to write fixture source: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}

	program, err := CreateProgram("tsconfig.json")
	if err != nil {
		t.Fatalf("Failed to create program: %v", err)
	}

	var sourceFile *ast.SourceFile
	var sourceFileID int
	for id, file := range program.GetSourceFiles() {
		if strings.HasSuffix(string(file.FileName()), "index.ts") {
			sourceFile = file
			sourceFileID = id
			break
		}
	}
	if sourceFile == nil {
		t.Fatalf("input file index.ts not found in program")
	}

	return semanticFixture{
		semantic:     CollectSemantic(program),
		program:      program,
		sourceFile:   sourceFile,
		sourceFileID: sourceFileID,
	}
}

func snapshotSemantic(semantic Semantic, file *ast.SourceFile, fileID int) []string {
	sourceText := file.Text()
	lines := []string{}

	var visit func(node *ast.Node, depth int)
	visit = func(node *ast.Node, depth int) {
		if node == nil {
			return
		}

		if node.Kind != ast.KindSourceFile && node.Kind != ast.KindEndOfFile {
			key := NodeReference{
				SourceFileId: fileID,
				Start:        node.Pos(),
				End:          node.End(),
			}

			typeStr := "<none>"
			if typeID, ok := semantic.Node2type[key]; ok && typeID != 0 {
				typeInfo := semantic.Typetab[typeID]
				name := fmt.Sprintf("type#%d", typeID)
				if n, exists := semantic.TypeExtra.Name[int(typeID)]; exists {
					name = string(n)
				}
				typeStr = fmt.Sprintf("%s(flags=%d)", name, typeInfo.Flags)
			}

			symStr := "<none>"
			if symID, ok := semantic.Node2sym[key]; ok {
				if sym, exists := semantic.Symtab[symID]; exists {
					symStr = fmt.Sprintf("%s(flags=%d,check=%d)", string(sym.Name), sym.Flags, sym.CheckFlags)
				} else {
					symStr = "sym#" + strconv.FormatInt(int64(symID), 10)
				}
			}

			if typeStr != "<none>" || symStr != "<none>" {
				snippet := strings.TrimSpace(sourceText[node.Pos():node.End()])
				indent := strings.Repeat("  ", depth)
				lines = append(lines, fmt.Sprintf("%s%s [%d,%d] %q type=%s sym=%s", indent, node.Kind, node.Pos(), node.End(), snippet, typeStr, symStr))
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child, depth+1)
			return false
		})
	}

	visit(file.AsNode(), 0)
	return lines
}

func TestSemanticSnapshot_PrimitiveTypes(t *testing.T) {
	fixture := buildSemanticFixture(t, "let a:number = 1;\nlet b: number = 2;")
	snapshot := strings.Join(snapshotSemantic(fixture.semantic, fixture.sourceFile, fixture.sourceFileID), "\n")

	expected := strings.Join([]string{
		`  KindVariableStatement [0,17] "let a:number = 1;" type=any(flags=1) sym=<none>`,
		`    KindVariableDeclarationList [0,16] "let a:number = 1" type=any(flags=1) sym=<none>`,
		`      KindVariableDeclaration [3,16] "a:number = 1" type=number(flags=64) sym=<none>`,
		`        KindIdentifier [3,5] "a" type=number(flags=64) sym=a(flags=2,check=0)`,
		`        KindNumberKeyword [6,12] "number" type=number(flags=64) sym=<none>`,
		`        KindNumericLiteral [14,16] "1" type=1(flags=2048) sym=<none>`,
		`  KindVariableStatement [17,36] "let b: number = 2;" type=any(flags=1) sym=<none>`,
		`    KindVariableDeclarationList [17,35] "let b: number = 2" type=any(flags=1) sym=<none>`,
		`      KindVariableDeclaration [21,35] "b: number = 2" type=number(flags=64) sym=<none>`,
		`        KindIdentifier [21,23] "b" type=number(flags=64) sym=b(flags=2,check=0)`,
		`        KindNumberKeyword [24,31] "number" type=number(flags=64) sym=<none>`,
		`        KindNumericLiteral [33,35] "2" type=2(flags=2048) sym=<none>`,
	}, "\n")

	if snapshot != expected {
		t.Fatalf("semantic snapshot mismatch.\nGot:\n%s\n\nExpected:\n%s", snapshot, expected)
	}
}

func TestSemanticSnapshot_ElementAccess(t *testing.T) {
	fixture := buildSemanticFixture(t, `const obj = {a:1}; const aa = obj["a"]`)
	snapshot := strings.Join(snapshotSemantic(fixture.semantic, fixture.sourceFile, fixture.sourceFileID), "\n")

	expected := strings.Join([]string{
		`  KindVariableStatement [0,18] "const obj = {a:1};" type=any(flags=1) sym=<none>`,
		`    KindVariableDeclarationList [0,17] "const obj = {a:1}" type=any(flags=1) sym=<none>`,
		`      KindVariableDeclaration [5,17] "obj = {a:1}" type={ a: number; }(flags=1048576) sym=<none>`,
		`        KindIdentifier [5,9] "obj" type={ a: number; }(flags=1048576) sym=obj(flags=2,check=0)`,
		`        KindObjectLiteralExpression [11,17] "{a:1}" type={ a: number; }(flags=1048576) sym=<none>`,
		`          KindPropertyAssignment [13,16] "a:1" type=number(flags=64) sym=<none>`,
		`            KindIdentifier [13,14] "a" type=number(flags=64) sym=a(flags=4,check=0)`,
		`            KindNumericLiteral [15,16] "1" type=1(flags=2048) sym=<none>`,
		`  KindVariableStatement [18,38] "const aa = obj[\"a\"]" type=any(flags=1) sym=<none>`,
		`    KindVariableDeclarationList [18,38] "const aa = obj[\"a\"]" type=any(flags=1) sym=<none>`,
		`      KindVariableDeclaration [24,38] "aa = obj[\"a\"]" type=number(flags=64) sym=<none>`,
		`        KindIdentifier [24,27] "aa" type=number(flags=64) sym=aa(flags=2,check=0)`,
		`        KindElementAccessExpression [29,38] "obj[\"a\"]" type=number(flags=64) sym=<none>`,
		`          KindIdentifier [29,33] "obj" type={ a: number; }(flags=1048576) sym=obj(flags=2,check=0)`,
		`          KindStringLiteral [34,37] "\"a\"" type=number(flags=64) sym=a(flags=33554436,check=0)`,
	}, "\n")

	if snapshot != expected {
		t.Fatalf("semantic snapshot mismatch.\nGot:\n%s\n\nExpected:\n%s", snapshot, expected)
	}
}
