package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Alignment lock-in tests. Each case pins a behavior that previously diverged
// from eslint-plugin-react-hooks and was verified against BOTH v4.6.2 and
// v7.1.1 (identical output) before being encoded here.

var alignmentValid = []rule_tester.ValidTestCase{
	// A local function whose only captures are stable hook values is itself
	// stable when referenced directly in an effect — no missing-dep report.
	{Code: `
function useNoCapture() {
  const [s, setS] = useState(0);
  const transform = (x: number) => x;
  useEffect(() => { setS(transform(1)); }, []);
  return s;
}
`, Tsx: true},
}

var alignmentInvalid = []rule_tester.InvalidTestCase{
	// (1) HOC wrapper (observer/mobx) no longer gates analysis: the effect
	// inside an `observer(() => {...})` component is analyzed like any other.
	{
		Code: `
const C = observer((props: any) => {
  const [s, setS] = useState(0);
  useEffect(() => { console.log(props.x); setS(1); }, []);
  return null;
});
`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useEffect has a missing dependency: 'props.x'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
const C = observer((props: any) => {
  const [s, setS] = useState(0);
  useEffect(() => { console.log(props.x); setS(1); }, [props.x]);
  return null;
});
`}}},
		},
	},
	// (2) A reactive value captured ONLY via an object-literal shorthand
	// (`{ apId }`) inside a local function is still a capture, so the function
	// is unstable and must be a dependency.
	{
		Code: `
function useShorthand({ apId }: { apId: string }) {
  const [t, setT] = useState<any[]>([]);
  const fetchTree = () => { setT([{ apId }]); };
  useEffect(() => { fetchTree(); }, []);
  return t;
}
`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useEffect has a missing dependency: 'fetchTree'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function useShorthand({ apId }: { apId: string }) {
  const [t, setT] = useState<any[]>([]);
  const fetchTree = () => { setT([{ apId }]); };
  useEffect(() => { fetchTree(); }, [fetchTree]);
  return t;
}
`}}},
		},
	},
	// (3) A local function is unstable when it calls ANOTHER local function
	// (upstream "won't check functions deeper": the inner function is not a
	// known-stable hook value, so the outer one captures it).
	{
		Code: `
function useCallsLocal({ apId }: { apId: string }) {
  const [t, setT] = useState<any[]>([]);
  const transform = (x: string) => [x];
  const query = () => { setT(transform(apId)); };
  useEffect(() => { query(); }, []);
  return t;
}
`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useEffect has a missing dependency: 'query'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function useCallsLocal({ apId }: { apId: string }) {
  const [t, setT] = useState<any[]>([]);
  const transform = (x: string) => [x];
  const query = () => { setT(transform(apId)); };
  useEffect(() => { query(); }, [query]);
  return t;
}
`}}},
		},
	},
	// (4) A recursive local function is unstable (its self-reference resolves
	// to the component-scope binding, which is a capture).
	{
		Code: `
function useRecursive({ tree, id }: { tree: any[]; id: string }) {
  const findLabel = (data: any[], target: string): string => {
    for (const node of data) {
      if (node.id === target) return node.label;
      if (node.children) { const r = findLabel(node.children, target); if (r) return r; }
    }
    return '';
  };
  return useMemo(() => findLabel(tree, id), [tree, id]);
}
`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "React Hook useMemo has a missing dependency: 'findLabel'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function useRecursive({ tree, id }: { tree: any[]; id: string }) {
  const findLabel = (data: any[], target: string): string => {
    for (const node of data) {
      if (node.id === target) return node.label;
      if (node.children) { const r = findLabel(node.children, target); if (r) return r; }
    }
    return '';
  };
  return useMemo(() => findLabel(tree, id), [findLabel, tree, id]);
}
`}}},
		},
	},
	// (5) When a ref's `.current` is read multiple times inside the cleanup
	// function, the warning points at the LAST occurrence (upstream overwrites
	// per dependency), not the first.
	{
		Code: `
function useRefCleanup() {
  const handleRef = useRef<HTMLElement>(null);
  useEffect(() => {
    handleRef.current?.addEventListener('a', () => {});
    return () => {
      handleRef.current?.removeEventListener('a', () => {});
      handleRef.current?.removeEventListener('b', () => {});
    };
  }, []);
}
`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: "The ref value 'handleRef.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'handleRef.current' to a variable inside the effect, and use that variable in the cleanup function."},
		},
	},
}

func TestExhaustiveDeps_Alignment(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		alignmentValid,
		alignmentInvalid,
	)
}
