package component_hook_factories

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestComponentHookFactoriesExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Dimension 4: declaration/container forms — lowercase factory output is not a component/hook ----
		{Code: `
function makeRenderer(value: string) {
  return function renderValue() {
    return <span>{value}</span>;
  };
}
		`, Tsx: true},
		// ---- Dimension 4: receiver/expression wrappers — bare use() is not a compiler hook name ----
		{Code: `
function makeUse() {
  return function use() {
    return use(resource);
  };
}
		`, Tsx: true},
		// ---- Dimension 4: access/key forms — computed object method is not one of upstream's visited function nodes ----
		{Code: `
function FactoryBag() {
  const bag = {
    ['Component']() {
      return <div />;
    },
  };
  return bag;
}
		`, Tsx: true},
		// ---- Dimension 4: graceful degradation — body-absent declarations do not crash ----
		{Code: `
declare function createDeclaredComponent(): () => JSX.Element;

abstract class BaseFactory {
  abstract createComponent(): () => JSX.Element;
}
		`, Tsx: true},
		// ---- Dimension 4: component param validation — primitive props annotation is not component-like upstream ----
		{Code: `
function createStringComponent() {
  return function Component(props: string) {
    return <div>{props}</div>;
  };
}
		`, Tsx: true},
		// ---- Dimension 4: component return validation — returning a function is not ReactNode-like upstream ----
		{Code: `
function createFunctionReturningComponent() {
  return function Component() {
    return function helper() {};
  };
}
		`, Tsx: true},
		// ---- Real-user: helper factory that returns a lowercase render callback ----
		{Code: `
export function makeCellRenderer(format: (value: string) => string) {
  return function renderCell(value: string) {
    return <td>{format(value)}</td>;
  };
}
		`, Tsx: true},
		// ---- Real-user: module-level memo/forwardRef callbacks are accepted ----
		{Code: `
const Button = memo(function ButtonImpl(props) {
  return <button {...props} />;
});

const Input = forwardRef((props, ref) => {
  return <input ref={ref} {...props} />;
});
		`, Tsx: true},
		// ---- Branch lock-in: nested function bodies are skipped when classifying the returned function ----
		{Code: `
function makeWrapper() {
  return function Component() {
    function helper() {
      return useState();
    }
    return helper;
  };
}
		`, Tsx: true},
		// ---- Branch lock-in: return with no argument is definitely non-ReactNode-like upstream ----
		{Code: `
function makeEmptyComponent() {
  return function Component() {
    return;
  };
}
		`, Tsx: true},
		// ---- Branch lock-in: object, class, new, and function returns are not ReactNode-like upstream ----
		{Code: `
function makeNonNodeComponents() {
  const A = function Component() {
    return {};
  };
  const B = function Component() {
    return class Inner {};
  };
  const C = function Component() {
    return new Thing();
  };
  const D = function Component() {
    return () => null;
  };
  return [A, B, C, D];
}
		`, Tsx: true},
		// ---- Branch lock-in: computed hook member calls are not compiler hook callees ----
		{Code: `
function makeComputedHook() {
  return function useComputed() {
    return Hooks['useThing']();
  };
}
		`, Tsx: true},
		// ---- Branch lock-in: malformed hookPattern is ignored like upstream option parsing ----
		{
			Code: `
function createSignalHook(source) {
  return function signalValue() {
    return signalRead(source);
  };
}
			`,
			Tsx:     true,
			Options: map[string]interface{}{"environment": map[string]interface{}{"hookPattern": "["}},
		},
		// ---- Dimension 4: class declarations/expressions are skipped by the React Compiler traversal ----
		{Code: `
class Holder {
  make = () => {
    return function Component() {
      return <div />;
    };
  };

  method() {
    function Child() {
      return <span />;
    }
    return Child;
  }
}
		`, Tsx: true},
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- Dimension 4: function declaration factory ----
		{
			Code: `
function createFeatureFlaggedComponent(flag: boolean) {
  return function Feature() {
    return flag ? <Enabled /> : <Disabled />;
  };
}
			`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "componentHookFactory"}},
		},
		// ---- Dimension 4: arrow factory assigned to a variable ----
		{
			Code: `
const createComponent = (defaultValue: string) => {
  return function Component() {
    return <span>{defaultValue}</span>;
  };
};
			`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "componentHookFactory"}},
		},
		// ---- Dimension 4: async and generator factory containers ----
		{
			Code: `
async function createAsyncComponent(load: () => Promise<string>) {
  const value = await load();
  return function Component() {
    return <span>{value}</span>;
  };
}

function* createHookGenerator() {
  yield function useGenerated() {
    return useState();
  };
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentHookFactory"},
				{MessageId: "componentHookFactory"},
			},
		},
		// ---- Dimension 4: TS wrappers around produced components/hooks ----
		{
			Code: `
type ComponentFactory = () => () => JSX.Element;
const createWrappedComponent = (() => {
  return ((function Component() {
    return <div />;
  }) as () => JSX.Element)!;
}) satisfies ComponentFactory;
			`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "componentHookFactory"}},
		},
		// ---- Dimension 4: optional chain and element access near factories ----
		{
			Code: `
function makeFactory(registry: any) {
  const Base = registry?.['Component'];
  return function Component() {
    return Base ? <Base /> : null;
  };
}
			`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "componentHookFactory"}},
		},
		// ---- Branch lock-in: component can be inferred from a direct hook call without JSX ----
		{
			Code: `
function makeHookOnlyComponent() {
  return function Component() {
    useState();
    return null;
  };
}
			`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "componentHookFactory"}},
		},
		// ---- Branch lock-in: namespaced compiler hook callees are recognized ----
		{
			Code: `
function makeNamespacedHook() {
  return function useNamespaced() {
    return React.useState(0)[0];
  };
}
			`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "componentHookFactory"}},
		},
		// ---- Branch lock-in: JSX fragments classify nested components ----
		{
			Code: `
function makeFragmentComponent() {
  return function FragmentComponent() {
    return <></>;
  };
}
			`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "componentHookFactory"}},
		},
		// Locks in upstream validateNoDynamicallyCreatedComponentsOrHooks(): non-Identifier parent names print as <anonymous>.
		{
			Code: `
registry.create = function () {
  return function Component() {
    return <div />;
  };
};
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "componentHookFactory",
				Message:   "Components and hooks cannot be created dynamically. The function `Component` appears to be a React component, but it's defined inside `<anonymous>`. Components and Hooks should always be declared at module scope.",
			}},
		},
		// ---- Dimension 4: JSX callback nesting reports at the callback boundary, not the outer component ----
		{
			Code: `
function Parent({items}: {items: string[]}) {
  return (
    <>
      {items.map(item => {
        function Child() {
          return <span>{item}</span>;
        }
        return <Child />;
      })}
    </>
  );
}
			`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "componentHookFactory"}},
		},
		// ---- Dimension 4: object literal property function factory ----
		{
			Code: `
const factories = {
  createComponent: function () {
    return function Component() {
      return <div />;
    };
  },
};
			`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "componentHookFactory"}},
		},
		// ---- Dimension 4: assignments and destructuring defaults ----
		{
			Code: `
let makeComponent;
makeComponent = () => function Component() {
  return <div />;
};

const {
  makeHook = () => function useGenerated() {
    return useReducer(reducer, 0);
  },
} = registry;
void makeHook;
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentHookFactory"},
				{MessageId: "componentHookFactory"},
			},
		},
		// ---- Dimension 4: rest binding and empty patterns are accepted around the reported factory ----
		{
			Code: `
function createFromRest(...rest: Array<() => string>) {
  const [] = rest;
  const [, value = 'fallback'] = rest;
  return function Component() {
    return <span>{value}</span>;
  };
}
			`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "componentHookFactory"}},
		},
		// ---- Dimension 4: memo and forwardRef callbacks created dynamically ----
		{
			Code: `
function createWrapped(kind) {
  const Memoed = memo(function MemoedComponent() {
    return <div>{kind}</div>;
  });
  const Forwarded = forwardRef((props, ref) => {
    return <input ref={ref} {...props} />;
  });
  return [Memoed, Forwarded];
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentHookFactory"},
				{MessageId: "componentHookFactory"},
			},
		},
		// ---- Dimension 4: custom hookPattern option ----
		{
			Code: `
function createSignalHook(source) {
  return function signalValue() {
    return signalRead(source);
  };
}
			`,
			Tsx:     true,
			Options: map[string]interface{}{"environment": map[string]interface{}{"hookPattern": "(?=signal)signal[A-Z]"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "componentHookFactory"}},
		},
		// ---- Branch lock-in: array-wrapped hookPattern option matches context.options shape ----
		{
			Code: `
function createTrackedHook(source) {
  return function trackValue() {
    return trackRead(source);
  };
}
			`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"environment": map[string]interface{}{"hookPattern": "^track[A-Z]"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "componentHookFactory"}},
		},
		// ---- Real-user: library helper that returns an endpoint hook ----
		{
			Code: `
export const makeUseEndpoint = (endpoint: string) => {
  return function useEndpoint() {
    return useQuery(endpoint);
  };
};
			`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "componentHookFactory"}},
		},
		// ---- Branch lock-in: same outer component can dynamically create both component and hook ----
		{
			Code: `
function Component() {
  function Nested() {
    return <div />;
  }
  const useNested = () => useState();
  return <Nested data={useNested()} />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentHookFactory"},
				{MessageId: "componentHookFactory"},
			},
		},
		// Locks in upstream returnsNonNode() arm: ObjectMethod bodies are skipped.
		{
			Code: `
function makeComponentWithUnreachableObjectMethod() {
  return function Component() {
    return <div />;
    const helpers = {
      render() {
        return {};
      },
    };
    return helpers;
  };
}
			`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "componentHookFactory"}},
		},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ComponentHookFactoriesRule,
		valid,
		invalid,
	)
}
