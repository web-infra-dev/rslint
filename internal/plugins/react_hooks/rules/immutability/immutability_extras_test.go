package immutability

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestImmutabilityExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / React Compiler fixture shape it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.

func immutabilityError(info immutableInfo, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "immutableMutation",
		Message:   buildImmutabilityMessage(info).Description,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func TestImmutabilityExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Official docs: create a new array instead of mutating state. ----
		{Code: `
function Component() {
  const [items, setItems] = useState([1, 2, 3]);
  const addItem = () => {
    setItems([...items, 4]);
  };
}
		`, Tsx: true},
		// ---- Official docs: create a new object instead of mutating state. ----
		{Code: `
function Component() {
  const [user, setUser] = useState({name: 'Alice'});
  const updateName = () => {
    setUser({...user, name: 'Bob'});
  };
}
		`, Tsx: true},
		// ---- Branch lock-in: mutation before a hook call freezes the value is allowed. ----
		{Code: `
function Component({a}) {
  const x = {a};
  x.y = true;
  useFreeze(x);
  return <div />;
}
		`, Tsx: true},
		// ---- Dimension 4: lowercase non-React function remains outside the lint target. ----
		{Code: `
function helper(arg) {
  (arg).value = 1;
  arg.items.push(1);
}
		`, Tsx: true},
		// ---- Dimension 4: useRef return values are mutable and intentionally unmatched. ----
		{Code: `
function Component() {
  const ref = useRef(null);
  ref.current = document.body;
  return <div />;
}
		`, Tsx: true},
		// ---- Branch lock-in: state setters from tuple position 2 are not immutable state. ----
		{Code: `
function Component() {
  const [state, setState] = useState({value: 0});
  setState.extra = true;
  return <div>{state.value}</div>;
}
		`, Tsx: true},
		// ---- Dimension 4: parameter shadowing inside nested callbacks does not fall back to name matching. ----
		{Code: `
function Component() {
  const [state] = useState({value: 0});
  const handler = (state) => {
    state.value = 1;
  };
  return <button onClick={handler} />;
}
		`, Tsx: true},
		// ---- Dimension 4: local shadowing in nested helpers does not mutate the outer state binding. ----
		{Code: `
function Component() {
  const [state] = useState({value: 0});
  function update() {
    const state = {value: 0};
    state.value = 1;
  }
  update();
  return <div>{state.value}</div>;
}
		`, Tsx: true},
		// ---- Upstream parity: PascalCase helpers without JSX or hook calls are not components. ----
		{Code: `
const D = <T, P extends keyof T>(obj: T, prop: P, value: T[P]) => {
  if (obj[prop] === undefined) {
    obj[prop] = value;
  }
};
		`, Tsx: true},
		// ---- Upstream parity: component props report property writes, not mutating method calls. ----
		{Code: `
function Component({list}) {
  list.sort();
  return <div />;
}
		`, Tsx: true},
		// ---- Upstream parity: state mutating methods are not emitted by the official rule. ----
		{Code: `
function Component() {
  const [items, setItems] = useState([1, 2, 3]);
  items.push(4);
  setItems(items.sort());
  Object.assign(items, [5]);
  return <div />;
}
		`, Tsx: true},
		// ---- Upstream parity: custom hook return values report property writes, not Object.assign. ----
		{Code: `
function Component() {
  const location = useLocation();
  return Object.assign(location, {query: new URLSearchParams(location.search)});
}
		`, Tsx: true},
		// ---- Upstream parity: non-use helper returns are mutable values. ----
		{Code: `
function Component() {
  const value = signalData({x: 1});
  value.x = 2;
  return <div />;
}
		`, Tsx: true},
		{Code: `
function signalThing(arg) {
  arg.value = 1;
}
		`, Tsx: true},
		// N/A: private property names and class members are not React Compiler
		// render-function inputs for this rule. Flow component/hook syntax is
		// covered as skipped upstream cases because tsgo does not parse it.
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- Official docs: Object property assignment. ----
		{
			Code: `
function Component() {
  const [user, setUser] = useState({name: 'Alice'});
  const updateName = () => {
    user.name = 'Bob';
    setUser(user);
  };
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableUseState, hookName: "useState"}, 5, 5, 5, 9),
			},
		},
		// ---- Official docs: nested object mutation. ----
		{
			Code: `
function UserProfile() {
  const [user, setUser] = useState({
    name: 'Alice',
    settings: {theme: 'light', notifications: true},
  });
  const toggleTheme = () => {
    user.settings.theme = 'dark';
    setUser(user);
  };
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableUseState, hookName: "useState"}, 8, 5, 8, 9),
			},
		},
		// ---- React Compiler fixture: error.modify-state-2 alias of state property. ----
		{
			Code: `
function Foo() {
  const [state, setState] = useState({foo: {bar: 3}});
  const foo = state.foo;
  foo.bar = 1;
  return state;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableUseState, hookName: "useState"}, 5, 3, 5, 6),
			},
		},
		// ---- React Compiler fixture: error.modify-useReducer-state. ----
		{
			Code: `
function Foo() {
  let [state, dispatch] = useReducer(reducer, {foo: 1});
  state.foo = 1;
  return state;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableUseReducer, hookName: "useReducer"}, 4, 3, 4, 8),
			},
		},
		// ---- React Compiler fixture: error.mutate-props. ----
		{
			Code: `
function Foo(props) {
  props.test = 1;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutablePropsOrHookArgs}, 3, 3, 3, 8),
			},
		},
		// ---- React Compiler fixture: error.mutate-hook-argument. ----
		{
			Code: `
function useHook(a, b) {
  b.test = 1;
  a.test = 2;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutablePropsOrHookArgs}, 3, 3, 3, 4),
				immutabilityError(immutableInfo{kind: immutablePropsOrHookArgs}, 4, 3, 4, 4),
			},
		},
		// ---- React Compiler fixture: error.invalid-mutate-context. ----
		{
			Code: `
function Component(props) {
  const context = useContext(FooContext);
  context.value = props.value;
  return context.value;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableHookReturn, hookName: "useContext"}, 4, 3, 4, 10),
			},
		},
		// ---- Real-user: facebook/react#35157 direct state property mutation from docs-shaped repro. ----
		{
			Code: `
function Component() {
  const [user, setUser] = useState({name: 'Alice'});
  user.name = 'Bob';
  return <div>{user.name}</div>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableUseState, hookName: "useState"}, 4, 3, 4, 7),
			},
		},
		// ---- Real-user: reddit react-hooks immutability with useRef-style frozen argument. ----
		{
			Code: `
function Component({isPlaying}) {
  const box = {isPlaying};
  useSomeHook(box);
  box.isPlaying = false;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableFrozenHookArgument}, 5, 3, 5, 6),
			},
		},
		// ---- Dimension 4: parenthesized receiver and TS assertion wrappers. ----
		{
			Code: `
function Component() {
  const [state, setState] = useState({value: 0});
  ((state as any)).value = 1;
  (state!).value = 2;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableUseState, hookName: "useState"}, 4, 5, 4, 10),
				immutabilityError(immutableInfo{kind: immutableUseState, hookName: "useState"}, 5, 4, 5, 9),
			},
		},
		// ---- Dimension 4: JSX-frozen values report the mutating method set covered upstream. ----
		{
			Code: `
function Component() {
  const items = [];
  const map = new Map();
  const element = <Child items={items} map={map} />;
  items["push"](4);
  items.pop();
  map.set("value", 1);
  map.delete("value");
  map.clear();
  return element;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableJSXValue}, 6, 3, 6, 8),
				immutabilityError(immutableInfo{kind: immutableJSXValue}, 7, 3, 7, 8),
				immutabilityError(immutableInfo{kind: immutableJSXValue}, 8, 3, 8, 6),
				immutabilityError(immutableInfo{kind: immutableJSXValue}, 9, 3, 9, 6),
				immutabilityError(immutableInfo{kind: immutableJSXValue}, 10, 3, 10, 6),
			},
		},
		// ---- Dimension 4: nested object destructuring aliases remain immutable. ----
		{
			Code: `
function Component() {
  const [user] = useState({settings: {theme: {dark: false}}});
  const {settings: {theme}} = user;
  theme.dark = true;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableUseState, hookName: "useState"}, 5, 3, 5, 8),
			},
		},
		// ---- Dimension 4: nested array destructuring aliases remain immutable. ----
		{
			Code: `
function Component() {
  const [items] = useState([[{done: false}]]);
  const [[first]] = items;
  first.done = true;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableUseState, hookName: "useState"}, 5, 3, 5, 8),
			},
		},
		// ---- Dimension 4: update and delete expressions. ----
		{
			Code: `
function Component({props}) {
  props.count++;
  delete props.value;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutablePropsOrHookArgs}, 3, 3, 3, 8),
				immutabilityError(immutableInfo{kind: immutablePropsOrHookArgs}, 4, 10, 4, 15),
			},
		},
		// ---- Dimension 4: destructuring assignment target degrades through object patterns. ----
		{
			Code: `
function Component(props) {
  ({value: props.value} = other);
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutablePropsOrHookArgs}, 3, 12, 3, 17),
			},
		},
		// Locks in upstream hook-call freeze arm: values passed to hook calls become immutable afterwards.
		{
			Code: `
function Component({a}) {
  const x = {a};
  useFreeze(x);
  x.y = true;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableFrozenHookArgument}, 5, 3, 5, 4),
			},
		},
		// Locks in upstream transitive callback freeze arm: hook callback captures a later-mutated local.
		{
			Code: `
function Component({count}) {
  const x = {value: 0};
  const cb = useIdentity(() => {
    console.log(x.value, count);
  });
  x.value += count;
  return <div>{cb}</div>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableFrozenHookArgument}, 7, 3, 7, 4),
			},
		},
		// ---- React Compiler fixture: error.invalid-mutate-after-freeze JSX freezes local values. ----
		{
			Code: `
function Component(props) {
  let x = [];
  let _ = <Component x={x} />;
  x.push(props.p2);
  return <div>{_}</div>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableJSXValue}, 5, 3, 5, 4),
			},
		},
		// ---- Dimension 4: JSX spread attributes freeze their referenced object. ----
		{
			Code: `
function Component() {
  const childProps = {};
  const element = <Child {...childProps} />;
  childProps.id = 1;
  return element;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableJSXValue}, 5, 3, 5, 13),
			},
		},
		// ---- React Compiler fixture: createElement-freeze freezes props after React.createElement. ----
		{
			Code: `
function Component() {
  const childProps = {style: {}};
  const element = React.createElement('div', childProps, ['hello']);
  childProps.style.width = 1;
  return <>{element}</>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableJSXValue}, 5, 3, 5, 13),
			},
		},
		// ---- React Compiler fixture: error.invalid-mutate-phi-which-could-be-frozen assignment alias. ----
		{
			Code: `
function Component(props) {
  const frozen = useHook();
  let x;
  if (props.cond) {
    x = frozen;
  } else {
    x = {};
  }
  x.property = true;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableHookReturn, hookName: "useHook"}, 10, 3, 10, 4),
			},
		},
		// ---- Branch lock-in: hook return assignment through destructuring propagates immutability. ----
		{
			Code: `
function Component() {
  let alias;
  ({value: alias} = useThing());
  alias.current = 1;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutableHookReturn, hookName: "useThing"}, 5, 3, 5, 8),
			},
		},
		// ---- Branch lock-in: anonymous default-exported JSX functions are React components. ----
		{
			Code: `
export default function(props) {
  props.value = 1;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutablePropsOrHookArgs}, 3, 3, 3, 8),
			},
		},
		// ---- Dimension 4: nested component scopes still enforce their own props. ----
		{
			Code: `
function Parent() {
  function Child(props) {
    props.value = 1;
    return <span />;
  }
  return <Child value={1} />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				immutabilityError(immutableInfo{kind: immutablePropsOrHookArgs}, 4, 5, 4, 10),
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ImmutabilityRule, valid, invalid)
}
