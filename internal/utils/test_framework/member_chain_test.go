package test_framework

import (
	"slices"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func parseFirstCall(t *testing.T, code string) (*ast.SourceFile, *ast.Node) {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, code, core.ScriptKindTS)

	var call *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			call = node
			return true
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	if call == nil {
		t.Fatalf("no call expression in %q", code)
	}
	return sourceFile, call
}

func TestGetMemberEntries(t *testing.T) {
	tests := []struct {
		code      string
		wantNames []string
	}{
		{
			code:      "describe.only()",
			wantNames: []string{"describe", "only"},
		},
		{
			code:      `describe[ "only" ]("math", () => {})`,
			wantNames: []string{"describe", "only"},
		},
		{
			code:      "describe[\n  \"only\"\n](\"math\", () => {})",
			wantNames: []string{"describe", "only"},
		},
		{
			code:      "test.concurrent.only()",
			wantNames: []string{"test", "concurrent", "only"},
		},
	}

	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			_, call := parseFirstCall(t, test.code)
			entries := GetMemberEntries(call)
			names := make([]string, len(entries))
			for i, entry := range entries {
				names[i] = entry.Name
			}
			if !slices.Equal(names, test.wantNames) {
				t.Fatalf("member names = %v, want %v", names, test.wantNames)
			}
		})
	}
}

func TestGetMemberEntriesRejectsDynamicElementAccess(t *testing.T) {
	_, call := parseFirstCall(t, "describe[getMode()]()")
	if entries := GetMemberEntries(call); entries != nil {
		t.Fatalf("dynamic element access entries = %#v, want nil", entries)
	}
}

func TestResolveFirstIdentifier(t *testing.T) {
	_, call := parseFirstCall(t, "((test).each)`table`('name', () => {})")
	identifier := ResolveFirstIdentifier(call.AsCallExpression().Expression)
	if identifier == nil || identifier.Kind != ast.KindIdentifier || identifier.AsIdentifier().Text != "test" {
		t.Fatalf("first identifier = %#v, want test", identifier)
	}
}
