package rules_of_hooks

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Test suites in this package are split across three files by case origin
// rather than by diagnostic kind:
//
//   - rules_of_hooks_upstream_test.go — TestRulesOfHooksRule_Upstream:
//     valid + invalid cases ported from upstream `RulesOfHooks-test.js`.
//   - rules_of_hooks_extras_test.go   — TestRulesOfHooksRule_Extras:
//     rslint-specific edges (tsgo AST quirks, naming / container boundary,
//     path-counting edges, extra useEffectEvent / settings shapes).
//   - rules_of_hooks_test.go (this file) — TestRulesOfHooksNilTypeChecker:
//     end-to-end nil-safety check for ctx.TypeChecker.

func TestRulesOfHooksNilTypeChecker(t *testing.T) {
	t.Parallel()
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "react.tsx")
	// Code intentionally exercises every listener that touches the
	// useEffectEvent resolver: a binding declaration, a JSX-attribute
	// reference, and a callee-position reference inside a regular
	// callback. If the rule deref'd a nil tc anywhere along that flow,
	// running listeners would panic.
	code := `
function MyComponent({ theme }: { theme: string }) {
  const onClick = useEffectEvent(() => {
    console.log(theme);
  });
  useEffect(() => {
    onClick();
  });
  return <Child onClick={onClick} />;
}
`
	fs := utils.NewOverlayVFSForFile(filePath, code)
	program, err := utils.CreateProgram(
		true, fs, rootDir, "tsconfig.json", utils.CreateCompilerHost(rootDir, fs),
	)
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}
	sourceFile := program.GetSourceFile(filePath)
	if sourceFile == nil {
		t.Fatalf("source file not found for %s", filePath)
		return
	}

	ctx := rule.RuleContext{
		SourceFile:  sourceFile,
		Program:     program,
		Settings:    nil,
		TypeChecker: nil, // explicitly nil — this is the path under test
		ReportNode: func(node *ast.Node, msg rule.RuleMessage) {
			// noop — we only care that nothing panics.
		},
		ReportRange: func(_ core.TextRange, _ rule.RuleMessage) {
		},
		ReportNodeWithFixes: func(_ *ast.Node, _ rule.RuleMessage, _ ...rule.RuleFix) {
		},
		ReportRangeWithFixes: func(_ core.TextRange, _ rule.RuleMessage, _ ...rule.RuleFix) {
		},
		ReportNodeWithSuggestions: func(_ *ast.Node, _ rule.RuleMessage, _ ...rule.RuleSuggestion) {
		},
		ReportRangeWithSuggestions: func(_ core.TextRange, _ rule.RuleMessage, _ ...rule.RuleSuggestion) {
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Run() panicked with nil TypeChecker: %v", r)
		}
	}()

	listeners := RulesOfHooksRule.Run(ctx, nil)

	// Walk the entire tree once, dispatching each node to the matching
	// listener. This exercises the same code paths the linter runtime
	// would hit in production.
	var walk func(n *ast.Node) bool
	walk = func(n *ast.Node) bool {
		if n == nil {
			return false
		}
		if cb, ok := listeners[n.Kind]; ok {
			cb(n)
		}
		n.ForEachChild(walk)
		return false
	}
	walk(sourceFile.AsNode())
}

