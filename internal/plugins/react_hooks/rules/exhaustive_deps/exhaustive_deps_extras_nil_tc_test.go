package exhaustive_deps

import (
	"reflect"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Lexical analysis must retain full diagnostics when no TypeChecker is available.
func TestExhaustiveDepsRule_NilTypeChecker(t *testing.T) {
	t.Parallel()
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
	program, sourceFile := createExhaustiveDepsProgram(t, "react.tsx", code)
	var messages []string
	ctx := (rule.RuleContext{
		SourceFile: sourceFile,
		Settings: map[string]interface{}{
			"react-hooks": map[string]interface{}{"additionalEffectHooks": "(useTrackedEffect)"},
		},
		TypeChecker: nil,
	}).WithProgram(lintprogram.NewFromCompiler(program)).WithReporter("test/exhaustive-deps", rule.SeverityWarning, func(d rule.RuleDiagnostic) {
		messages = append(messages, d.Message.Description)
	})
	listeners := ExhaustiveDepsRule.Run(ctx, nil)
	var walk func(*ast.Node) bool
	walk = func(n *ast.Node) bool {
		if cb := listeners[n.Kind]; cb != nil {
			cb(n)
		}
		n.ForEachChild(walk)
		return false
	}
	walk(sourceFile.AsNode())
	want := []string{
		"The ref value 'ref.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'ref.current' to a variable inside the effect, and use that variable in the cleanup function.",
		"React Hook useEffect has a missing dependency: 'count'. Either include it or remove the dependency array. You can also do a functional update 'setCount(c => ...)' if you only need 'count' in the 'setCount' call.",
		"React Hook useTrackedEffect has a missing dependency: 'theme'. Either include it or remove the dependency array.",
	}
	slices.Sort(messages)
	slices.Sort(want)
	if !reflect.DeepEqual(messages, want) {
		t.Fatalf("nil checker changed diagnostics:\ngot: %q\nwant: %q", messages, want)
	}
}
