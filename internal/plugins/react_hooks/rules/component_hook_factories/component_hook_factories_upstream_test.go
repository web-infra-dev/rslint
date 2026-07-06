package component_hook_factories

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestComponentHookFactoriesUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- React docs: module-level component example is accepted ----
		{Code: `
function Component({ defaultValue }) {
  return <span>{defaultValue}</span>;
}
		`, Tsx: true},
		// ---- React docs: module-level hook example is accepted ----
		{Code: `
function useData(endpoint) {
  return useMemo(() => endpoint, [endpoint]);
}
		`, Tsx: true},
		// ---- React docs: troubleshooting replacement example is accepted ----
		{Code: `
function Button({color, children}) {
  return (
    <button style={{backgroundColor: color}}>
      {children}
    </button>
  );
}

function App() {
  return (
    <>
      <Button color="red">Red</Button>
      <Button color="blue">Blue</Button>
    </>
  );
}
		`, Tsx: true},
		// ---- Upstream compiler heuristic: PascalCase alone is not enough ----
		{Code: `
function createComponent(defaultValue) {
  return function Component() {
    return null;
  };
}
		`, Tsx: true},
		// ---- Upstream compiler heuristic: hook name alone is not enough ----
		{Code: `
function createCustomHook(endpoint) {
  return function useData() {
    return endpoint;
  };
}
		`, Tsx: true},
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- React docs: factory function creating components ----
		{
			Code: `
function createComponent(defaultValue) {
  return function Component() {
    return <span>{defaultValue}</span>;
  };
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "componentHookFactory",
				Message:   "Components and hooks cannot be created dynamically. The function `Component` appears to be a React component, but it's defined inside `createComponent`. Components and Hooks should always be declared at module scope.",
				Line:      2,
				Column:    10,
				EndLine:   2,
				EndColumn: 25,
			}},
		},
		// ---- React docs: component defined inside component ----
		{
			Code: `
function Parent() {
  function Child() {
    return <div />;
  }
  return <Child />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "componentHookFactory",
				Line:      2,
				Column:    10,
			}},
		},
		// ---- React docs: hook factory function ----
		{
			Code: `
function createCustomHook(endpoint) {
  return function useData() {
    return useMemo(() => endpoint, [endpoint]);
  };
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "componentHookFactory",
				Line:      2,
				Column:    10,
			}},
		},
		// ---- React docs: dynamic button factory troubleshooting example ----
		{
			Code: `
function makeButton(color) {
  return function Button({children}) {
    return (
      <button style={{backgroundColor: color}}>
        {children}
      </button>
    );
  };
}

const RedButton = makeButton('red');
const BlueButton = makeButton('blue');
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "componentHookFactory",
				Line:      2,
				Column:    10,
			}},
		},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ComponentHookFactoriesRule,
		valid,
		invalid,
	)
}
