package exhaustive_deps

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestExhaustiveDepsRule_NilTypeChecker verifies the rule operates
// without panicking when the TypeChecker is unavailable. rslint
// schedules rules without `RequiresTypeInfo: true` against "gap files"
// (files in the program but not in `typeInfoFiles`) with a nil checker;
// every code path that would normally consult TC must degrade
// gracefully via the structural / name-walk fallback.
//
// The fixture below is shaped to exercise every TC-dependent branch:
//   - useState binding (stable-known-hook-value)
//   - useRef binding + ref.current cleanup
//   - useEffectEvent binding + reference inside effect
//   - missing dep with property chain (getDependency walk)
//   - declared dep that resolves outside the component (computeExternalDeps)
//   - construction warning (scanForConstructions)
//   - additionalHooks via settings
//
// Success criterion: the rule's listeners run end-to-end without
// panicking. Reports may differ from the TC-enabled path (the fallback
// is structural and best-effort), so we don't assert specific messages
// here — we only assert non-panic behavior.
func TestExhaustiveDepsRule_NilTypeChecker(t *testing.T) {
	t.Parallel()
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "react.tsx")
	code := `
function MyComponent({theme}) {
  const [count, setCount] = useState(0);
  const ref = useRef(null);
  const onClick = useEffectEvent(() => {
    console.log(theme);
  });
  const handler = () => {
    return Store.subscribe(onClick);
  };
  useEffect(() => {
    onClick();
    setCount(count + 1);
    handler();
    console.log(props.foo.bar);
    return () => {
      console.log(ref.current);
    };
  }, []);
  useTrackedEffect(() => { console.log(theme); }, []);
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

	// Settings include `additionalHooks` so the additional-hooks code
	// path is also exercised under nil TC.
	settings := map[string]interface{}{
		"react-hooks": map[string]interface{}{"additionalHooks": "(useTrackedEffect)"},
	}

	ctx := rule.RuleContext{
		SourceFile:  sourceFile,
		Program:     program,
		Settings:    settings,
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
		ReportNodeWithFixesAndSuggestions: func(_ *ast.Node, _ rule.RuleMessage, _ []rule.RuleFix, _ []rule.RuleSuggestion) {
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ExhaustiveDepsRule.Run() panicked with nil TypeChecker: %v", r)
		}
	}()

	listeners := ExhaustiveDepsRule.Run(ctx, nil)

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
