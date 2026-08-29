package prefer_ternary

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// TestPreferTernarySourceOnlyComputedKey verifies the no-TypeChecker path
// directly. RuleTester builds a compiler program even for .js inputs, so it
// cannot exercise this source-only context.
func TestPreferTernarySourceOnlyComputedKey(t *testing.T) {
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.js",
		Path:     "/test.js",
	}, `if (t) { (x)['b' + 'ar'] = a; } else { x.bar = b; }`, core.ScriptKindJS)

	comments := rule.NewCommentStore(source)
	var diagnostics []rule.RuleDiagnostic
	ctx := (rule.RuleContext{
		SourceFile:     source,
		Comments:       comments,
		DisableManager: rule.NewDisableManager(source, comments),
	}).WithReporter(PreferTernaryRule.Name, rule.SeverityError, func(d rule.RuleDiagnostic) {
		diagnostics = append(diagnostics, d)
	})

	listeners := PreferTernaryRule.Run(ctx, nil)
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		if listener := listeners[node.Kind]; listener != nil {
			listener(node)
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(source.AsNode())

	if len(diagnostics) != 1 {
		t.Fatalf("source-only computed-key diagnostics = %d, want 1", len(diagnostics))
	}
}
