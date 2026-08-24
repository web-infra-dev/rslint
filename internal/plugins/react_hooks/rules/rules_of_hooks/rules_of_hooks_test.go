package rules_of_hooks

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/react_hooksutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
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
	filePath := tspath.ResolvePath(rootDir.Dir, "react.tsx")
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
	fs := utils.NewOverlayVFS(rootDir.FS, map[string]string{filePath: code})
	program, err := utils.CreateProgram(
		true, fs, rootDir.Dir, "tsconfig.json", utils.CreateCompilerHost(rootDir.Dir, fs),
	)
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}
	sourceFile := program.GetSourceFile(filePath)
	if sourceFile == nil {
		t.Fatalf("source file not found for %s", filePath)
		return
	}

	diagnosticCount := 0
	ctx := (rule.RuleContext{
		SourceFile:  sourceFile,
		Settings:    nil,
		TypeChecker: nil, // explicitly nil — this is the path under test
	}).WithProgram(lintprogram.NewFromCompiler(program)).WithReporter("test/rules-of-hooks", rule.SeverityWarning, func(rule.RuleDiagnostic) {
		diagnosticCount++
	})

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
	if diagnosticCount != 1 {
		t.Fatalf("reported %d diagnostics with nil TypeChecker, want 1", diagnosticCount)
	}
}

func TestSourceMayUseHooks(t *testing.T) {
	if !sourceMayUseHooks(nil) || !sourceMayUseHooks(&ast.SourceFile{}) {
		t.Fatal("missing parser identifier metadata must conservatively keep listeners")
	}

	for _, testCase := range []struct {
		name string
		code string
		want bool
	}{
		{name: "ordinary call", code: `service.render(value)`, want: false},
		{name: "string only", code: `const marker = "useState()"`, want: false},
		{name: "lowercase suffix", code: `useful()`, want: false},
		{name: "computed property is conservative", code: `React["useState"]()`, want: true},
		{name: "bare use", code: `use(value)`, want: true},
		{name: "identifier hook", code: `useState(value)`, want: true},
		{name: "property hook", code: `React.useState(value)`, want: true},
		{name: "digit suffix", code: `use1(value)`, want: true},
		{name: "escaped identifier", code: `u\u0073eState(value)`, want: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/source.tsx",
				Path:     "/source.tsx",
			}, testCase.code, core.ScriptKindTSX)
			if got := sourceMayUseHooks(sourceFile); got != testCase.want {
				t.Fatalf("sourceMayUseHooks(%q) = %v, want %v", testCase.code, got, testCase.want)
			}
			listeners := RulesOfHooksRule.Run(rule.RuleContext{SourceFile: sourceFile}, nil)
			if got := len(listeners) != 0; got != testCase.want {
				t.Fatalf("listener presence for %q = %v, want %v", testCase.code, got, testCase.want)
			}
		})
	}
}

func TestHookNameShapeMatchesSharedClassifier(t *testing.T) {
	names := []string{"", "use", "UseState", "useful", "useState", "use1"}
	for next := byte(0); next < 128; next++ {
		names = append(names, "use"+string(next)+"Tail")
	}
	for _, name := range names {
		if got, want := hasHookNameShape(name), react_hooksutil.IsHookName(name); got != want {
			t.Fatalf("hasHookNameShape(%q) = %v, shared classifier = %v", name, got, want)
		}
	}
}

func TestReportRangeMatchesNodeTrimming(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/source.tsx",
		Path:     "/source.tsx",
	}, `
function Component() {
  const value = (
    /* leading trivia */ React.useState
  )();
  return value;
}
`, core.ScriptKindTSX)

	checked := 0
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindCallExpression, ast.KindPropertyAccessExpression, ast.KindIdentifier:
			got := reportRange(sourceFile.Text(), node)
			want := utils.TrimNodeTextRange(sourceFile, node)
			if got.Pos() != want.Pos() || got.End() != want.End() {
				t.Fatalf("range for %v = [%d,%d), want [%d,%d)", node.Kind, got.Pos(), got.End(), want.Pos(), want.End())
			}
			checked++
		}
		node.ForEachChild(visit)
		return false
	}
	visit(sourceFile.AsNode())
	if checked == 0 {
		t.Fatal("expected representative report nodes")
	}
}

func TestRulesOfHooksEffectEventShadowing(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RulesOfHooksRule,
		[]rule_tester.ValidTestCase{{
			Code: `
function MyComponent() {
  const onClick = useEffectEvent(() => {});
  function nested(onClick) {
    consume(onClick);
  }
  useEffect(() => { onClick(); });
}
`,
			Tsx: true,
		}},
		nil,
	)
}

func TestRulesOfHooksDocumentRegressions(t *testing.T) {
	const (
		conditional       = `React Hook "useFirst" is called conditionally. React Hooks must be called in the exact same order in every component render.`
		conditionalSecond = `React Hook "useSecond" is called conditionally. React Hooks must be called in the exact same order in every component render.`
		early             = `React Hook "useSecond" is called conditionally. React Hooks must be called in the exact same order in every component render. Did you accidentally call a React Hook after an early return?`
		loop              = `React Hook "useHook" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`
		named             = `React Hook "useHook" is called in function "lower" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`
		callback          = `React Hook "useHook" cannot be called inside a callback. React Hooks must be called in a React function component or a custom React Hook function.`
	)

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RulesOfHooksRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `
const handlers = {
  [key]: () => {
    useHook();
  },
  [otherKey]() {
    useOtherHook();
  },
};
`,
				Tsx: true,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
const _Component: React.FC<Props> = () => {
  useHook();
};
`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					Message: `React Hook "useHook" is called in function "_Component: React.FC<Props>" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`,
				}},
			},
			{
				Code: `
function Component() {
  const handlers = {
    [key]: () => {
      useHook();
    },
    [otherKey]() {
      useOtherHook();
    },
  };
  return handlers;
}
`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: callback},
					{Message: `React Hook "useOtherHook" cannot be called inside a callback. React Hooks must be called in a React function component or a custom React Hook function.`},
				},
			},
			{
				Code: `
function Component(value) {
  if (!value) return null;
  useFirst();
  const resolved = value ?? fallback;
  useSecond();
  return resolved;
}
`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: conditional},
					{Message: early},
				},
			},
			{
				Code: `
function Component(value) {
  switch (value) {
    case 0:
      useFirst();
      if (flag) log();
      return;
    default:
      useSecond();
  }
}
`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: conditional},
					{Message: conditionalSecond},
				},
			},
			{
				Code: `
function useThing(value) {
  switch (value) {
    case 0:
      if (!value || missing) return;
      useFirst();
      break;
    case 1:
      if (!value) return;
      useSecond();
      break;
  }
}
`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: conditional + ` Did you accidentally call a React Hook after an early return?`},
					{Message: early},
				},
			},
			{
				Code: `
function useThing(value) {
  switch (value) {
    case 0:
      if (!value) return;
      useFirst();
      break;
    case 1:
      break;
  }
}
`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{Message: conditional}},
			},
			{
				Code: `
function Component(value) {
  switch (value) {
    case 0:
      useFirst();
      return;
    default:
      useSecond();
  }
}
`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: conditional + ` Did you accidentally call a React Hook after an early return?`},
					{Message: early},
				},
			},
			{
				Code: `
function lower(values) {
  for (const value of values) {
    switch (value) {
      case 0:
        break;
      case 1:
        useHook();
        break;
    }
  }
}
`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{Message: named}},
			},
			{
				Code: `
function lower(values) {
  for (const value of values) {
    switch (value) {
      case 0:
        log(value);
      case 1:
        useHook();
        break;
    }
  }
}
`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: loop},
					{Message: named},
				},
			},
			{
				Code: `
function lower(values) {
  for (const value of values) {
    switch (value) {
      case 0:
        return;
      case 1:
        useHook();
        break;
    }
  }
}
`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: loop},
					{Message: named},
				},
			},
			{
				Code: `
function lower(values) {
  for (const value of values) {
    switch (value) {
      case 0:
        try {
          break;
        } catch {}
      case 1:
        useHook();
        break;
    }
  }
}
`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{Message: named}},
			},
		},
	)
}
