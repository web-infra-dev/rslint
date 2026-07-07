package globals

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestGlobalsExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row it covers, so future refactors can't
// silently regress them without breaking a named lock-in.

func globalsError(name string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "globalReassignment",
		Message:   buildGlobalReassignmentMessage(name).Description,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func TestGlobalsExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Dimension 4: parenthesized property mutation remains outside globals. ----
		{Code: `
function Component() {
  (window).location.href = "/";
  return <div />;
}
		`, Tsx: true},
		// ---- Dimension 4: element access mutation remains outside globals. ----
		{Code: `
function Component() {
  window["location"] = {};
  return <div />;
}
		`, Tsx: true},
		// ---- Dimension 4: empty destructuring target degrades gracefully. ----
		{Code: `
function Component(props) {
  [] = props.items;
  ({}) = props;
  return <div />;
}
		`, Tsx: true},
		// ---- Scope: same-block lexical bindings are local, even before declaration. ----
		{Code: `
function Component() {
  someGlobal = true;
  let someGlobal;
  return <div />;
}
		`, Tsx: true},
		// ---- Scope: nested block lexical bindings are local to assignments in that block. ----
		{Code: `
function Component() {
  {
    let someGlobal;
    someGlobal = true;
  }
  return <div />;
}
		`, Tsx: true},
		// ---- Scope: var declarations hoist from nested blocks to the component body. ----
		{Code: `
function Component() {
  someGlobal = true;
  if (cond) {
    var someGlobal;
  }
  return <div />;
}
		`, Tsx: true},
		// ---- Scope: named function expression's own name is local to its body. ----
		{Code: `
const wrapper = function Foo() {
  Foo = true;
  return <div />;
};
		`, Tsx: true},
		// ---- Scope: destructured parameters are local to the component body. ----
		{Code: `
function Component({someGlobal}) {
  someGlobal = true;
  return <div />;
}
		`, Tsx: true},
		// ---- Dimension 4: update expressions are intentionally unmatched. ----
		{Code: `
function Component() {
  ((someGlobal))++;
  return <div />;
}
		`, Tsx: true},
		// ---- Official ESLint: logical assignment operators are not reported by globals. ----
		{Code: `
function Component() {
  a ||= true;
  b &&= true;
  c ??= true;
  return <div />;
}
		`, Tsx: true},
		// ---- Official ESLint with @typescript-eslint/parser: TS wrappers are not globals targets. ----
		{Code: `
type Flag = boolean;
function Component() {
  (someGlobal as any) = true;
  someOtherGlobal! = true;
  (thirdGlobal satisfies Flag) = true;
  return <div />;
}
		`, Tsx: true},
		// ---- Dimension 4: map callbacks are not render helpers for this rule. ----
		{Code: `
function Component({items}) {
  return items.map(item => {
    someGlobal = true;
    return <div key={item} />;
  });
}
		`, Tsx: true},
		// ---- Real-user: facebook/react#34776 event handler global property write is not globals. ----
		{Code: `
function Link({isHashLink, to}) {
  const go = () => {
    if (isHashLink) {
      window.location.hash = to;
    }
  };
  return <button onClick={go} />;
}
		`, Tsx: true},
		// ---- Real-user: facebook/react#31630 effect helper may write outside render. ----
		{Code: `
function Component() {
  useEffect(() => {
    writeAfterRender();
  });
  function writeAfterRender() {
    someGlobal = true;
  }
  return <div />;
}
		`, Tsx: true},
		// ---- Real-user: facebook/react#31544 process property writes in callbacks are not globals. ----
		{Code: `
function Component() {
  const cb = useCallback(() => {
    process.exitCode = 1;
  }, []);
  return <button onClick={cb} />;
}
		`, Tsx: true},
		// Locks in upstream active helper arm: helpers do not activate second-level helpers.
		{Code: `
function Component() {
  function foo() {
    function bar() {
      someGlobal = true;
    }
    bar();
  }
  foo();
  return <div />;
}
		`, Tsx: true},
		// Locks in upstream non-render callback arm: helpers inside useCallback stay inactive.
		{Code: `
function Component() {
  const cb = useCallback(() => {
    const helper = () => {
      someGlobal = true;
    };
    helper();
  });
  return <div onClick={cb} />;
}
		`, Tsx: true},
		// Locks in upstream object-literal helper arm: unrelated member calls do not activate local functions.
		{Code: `
function Component() {
  const foo = () => {
    someGlobal = true;
  };
  other.foo();
  return <div />;
}
		`, Tsx: true},
		// Locks in upstream non-render callback arm for React.useCallback namespace calls.
		{Code: `
function Component() {
  React.useCallback(() => {
    someGlobal = true;
  }, []);
  return <div />;
}
		`, Tsx: true},
		// N/A: private keys, computed property names, class members, and overload
		// signatures are not assignment targets for react-hooks/globals.
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- Dimension 4: parenthesized identifier assignment target. ----
		{
			Code: `
function Component() {
  (someGlobal) = true;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("someGlobal", 3, 4, 3, 14),
			},
		},
		// ---- Dimension 4: array destructuring assignment target. ----
		{
			Code: `
function Component(props) {
  [x] = props.items;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("x", 3, 4, 3, 5),
			},
		},
		// ---- Dimension 4: object destructuring alias assignment target. ----
		{
			Code: `
function Component(props) {
  ({value: x} = props);
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("x", 3, 12, 3, 13),
			},
		},
		// ---- Dimension 4: nested destructuring targets include defaults and rest elements. ----
		{
			Code: `
function Component(props) {
  ({items: [first = fallback, ...rest], meta: {value}} = props);
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("first", 3, 13, 3, 29),
				globalsError("rest", 3, 34, 3, 38),
				globalsError("value", 3, 48, 3, 53),
			},
		},
		// ---- Dimension 4: object destructuring defaults report the default assignment range. ----
		{
			Code: `
function Component(props) {
  ({x = 1, y: z = 2} = props);
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("x", 3, 5, 3, 10),
				globalsError("z", 3, 15, 3, 20),
			},
		},
		// ---- Dimension 4: compound assignment reports the full assignment expression. ----
		{
			Code: `
function Component() {
  x += 1;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("x", 3, 3, 3, 9),
			},
		},
		// ---- Official ESLint: chained plain assignments report each written binding. ----
		{
			Code: `
function Component() {
  a = b = true;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("a", 3, 3, 3, 4),
				globalsError("b", 3, 7, 3, 8),
			},
		},
		// ---- Scope: sibling block lexical declarations do not make a name local here. ----
		{
			Code: `
function Component() {
  someGlobal = true;
  { let someGlobal; }
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("someGlobal", 3, 3, 3, 13),
			},
		},
		// ---- Scope: function declaration names belong to the outer scope. ----
		{
			Code: `
function Component() {
  Component = true;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("Component", 3, 3, 3, 12),
			},
		},
		// ---- Scope: component variables are declared outside the render body. ----
		{
			Code: `
const Component = () => {
  Component = true;
  return <div />;
};
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("Component", 3, 3, 3, 12),
			},
		},
		// ---- Dimension 4: IIFE runs during render. ----
		{
			Code: `
function Component() {
  (() => {
    someGlobal = true;
  })();
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("someGlobal", 4, 5, 4, 15),
			},
		},
		// Locks in upstream JSX prop render-helper arm: returning JSX makes a prop callback render-time.
		{
			Code: `
function Component() {
  const renderItem = item => {
    someGlobal = true;
    return <Item item={item} />;
  };
  return <ItemList renderItem={renderItem} />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("someGlobal", 4, 5, 4, 15),
			},
		},
		// Locks in upstream compiler root arm: hook call makes a hook-like function render-time.
		{
			Code: `
function useFoo(props) {
  useState();
  x = props;
  return {x};
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("x", 4, 3, 4, 4),
			},
		},
		// Locks in upstream useMemo callback arm: direct helper inside useMemo is checked.
		{
			Code: `
function Component() {
  useMemo(() => {
    const helper = () => {
      someGlobal = true;
    };
    helper();
    return 1;
  });
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("someGlobal", 5, 7, 5, 17),
			},
		},
		// Locks in upstream object-literal helper arm: called object helpers run during render.
		{
			Code: `
function Component() {
  const helpers = {
    foo: () => {
      someGlobal = true;
    },
  };
  helpers.foo();
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("someGlobal", 5, 7, 5, 17),
			},
		},
		// Locks in upstream object-literal helper arm: any member call of the object is conservative.
		{
			Code: `
function Component() {
  const helpers = {
    foo: () => {
      someGlobal = true;
    },
  };
  helpers.bar();
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("someGlobal", 5, 7, 5, 17),
			},
		},
		// Locks in upstream object-literal helper arm: nested helper objects still attach to the root binding.
		{
			Code: `
function Component() {
  const helpers = {
    nested: {
      foo: () => {
        someGlobal = true;
      },
    },
  };
  helpers.nested.foo();
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("someGlobal", 6, 9, 6, 19),
			},
		},
		// Locks in upstream useMemo callback arm for React.useMemo namespace calls.
		{
			Code: `
function Component() {
  React.useMemo(() => {
    someGlobal = true;
    return 1;
  }, []);
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("someGlobal", 4, 5, 4, 15),
			},
		},
		// Locks in upstream compiler root arm: React.memo callbacks are components.
		{
			Code: `
export default React.memo(function Component() {
  someGlobal = true;
  return <div />;
});
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("someGlobal", 3, 3, 3, 13),
			},
		},
		// Locks in upstream JSX prop render-helper arm with conditional JSX returns.
		{
			Code: `
function Component() {
  const renderItem = item => {
    someGlobal = true;
    return item ? <Item item={item} /> : null;
  };
  return <ItemList renderItem={renderItem} />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				globalsError("someGlobal", 4, 5, 4, 15),
			},
		},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &GlobalsRule,
		valid,
		invalid,
	)
}
