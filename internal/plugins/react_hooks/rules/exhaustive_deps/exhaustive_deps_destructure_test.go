package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDepsRule_Destructure exercises all destructure-binding
// forms found in real-world React code. Each shape exercises a TC
// resolution path that, before the rsbuild-found regression, could
// silently mis-classify the binding's symbol as external (TC sometimes
// returns the SOURCE TYPE's property symbol rather than the local
// BindingElement, especially when the source object is typed via an
// external `.d.ts`).
//
// Lock-in goal: when the destructure-bound name is captured in the
// callback AND listed in deps, NO diagnostic must fire (the receiver
// reference must NOT be silently dropped). When it's NOT in deps, the
// missing-dep diagnostic fires correctly.
func TestExhaustiveDepsRule_Destructure(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		destructureValid,
		destructureInvalid,
	)
}

var destructureValid = []rule_tester.ValidTestCase{
	// Plain destructure rename — the rsbuild regression case.
	{Code: `
		declare function useData(): { siteData: { lang: string } };
		function MyHook() {
			const { siteData: { lang: defaultLang } } = useData();
			return useCallback((url: string) => defaultLang + url, [defaultLang]);
		}
	`, Tsx: true},

	// Same as above but flat (no nested) — TC may resolve differently.
	{Code: `
		declare function useData(): { value: number };
		function MyHook() {
			const { value: localValue } = useData();
			return useCallback(() => localValue, [localValue]);
		}
	`, Tsx: true},

	// Array destructure with rename via tuple typing.
	{Code: `
		declare function useTuple(): readonly [number, string];
		function MyHook() {
			const [num, str] = useTuple();
			return useCallback(() => num + str, [num, str]);
		}
	`, Tsx: true},

	// Destructure with default value.
	{Code: `
		declare function useData(): { value?: number };
		function MyHook() {
			const { value = 0 } = useData();
			return useCallback(() => value, [value]);
		}
	`, Tsx: true},

	// Destructure of a parameter (props is an Object pattern parameter).
	{Code: `
		function MyComponent({ id, name }: { id: number; name: string }) {
			useEffect(() => { console.log(id, name); }, [id, name]);
		}
	`, Tsx: true},

	// Nested destructure with property chain in body — declaring the
	// receiver covers nested member access.
	{Code: `
		declare function useShape(): { a: { b: number } };
		function MyHook() {
			const { a } = useShape();
			return useCallback(() => a.b, [a]);
		}
	`, Tsx: true},

	// Destructure rename + the renamed binding is referenced as receiver
	// of property chain in callback.
	{Code: `
		declare function useShape(): { user: { profile: { name: string } } };
		function MyHook() {
			const { user: u } = useShape();
			return useCallback(() => u.profile.name, [u.profile]);
		}
	`, Tsx: true},

	// Destructure with `as` cast — common pattern when refining types.
	{Code: `
		declare function useData(): unknown;
		function MyHook() {
			const { id } = useData() as { id: number };
			return useCallback(() => id, [id]);
		}
	`, Tsx: true},

	// Spread in destructure — bound name is local.
	{Code: `
		declare function useData(): { id: number; rest: any };
		function MyHook() {
			const { id, ...other } = useData();
			void other;
			return useCallback(() => id, [id]);
		}
	`, Tsx: true},

	// Multiple useCallback referencing the same destructured renamed binding.
	{Code: `
		declare function useData(): { siteData: { lang: string } };
		function MyHook() {
			const { siteData: { lang: defaultLang } } = useData();
			const a = useCallback(() => defaultLang + 'a', [defaultLang]);
			const b = useCallback(() => defaultLang + 'b', [defaultLang]);
			return [a, b];
		}
	`, Tsx: true},
}

var destructureInvalid = []rule_tester.InvalidTestCase{
	// Plain destructure rename — missing.
	{
		Code: `
			declare function useData(): { siteData: { lang: string } };
			function MyHook() {
				const { siteData: { lang: defaultLang } } = useData();
				return useCallback((url: string) => defaultLang + url, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useCallback has a missing dependency: 'defaultLang'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			declare function useData(): { siteData: { lang: string } };
			function MyHook() {
				const { siteData: { lang: defaultLang } } = useData();
				return useCallback((url: string) => defaultLang + url, [defaultLang]);
			}
		`}},
			},
		},
	},

	// Renamed destructure declared as dep but NOT used — unnecessary.
	// Lock-in: the diagnostic must NOT carry the "Outer scope values"
	// suffix (which is the rsbuild regression we fixed). The bound name
	// is local to the component, so the suffix is suppressed.
	{
		Code: `
			declare function useData(): { siteData: { lang: string } };
			function MyHook({ id }: { id: number }) {
				const { siteData: { lang: defaultLang } } = useData();
				void defaultLang;
				return useCallback(() => id, [id, defaultLang]);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				// Plain "unnecessary" — no "Outer scope values" suffix.
				Message: "React Hook useCallback has an unnecessary dependency: 'defaultLang'. Either exclude it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			declare function useData(): { siteData: { lang: string } };
			function MyHook({ id }: { id: number }) {
				const { siteData: { lang: defaultLang } } = useData();
				void defaultLang;
				return useCallback(() => id, [id]);
			}
		`}},
			},
		},
	},

	// Destructured prop, renamed; property chain read inside callback.
	// Dep key is the full chain `u.name` (matches upstream's getDependency
	// behavior — non-call, non-.current reads ascend the receiver chain).
	{
		Code: `
			function MyComponent({ user: u }: { user: { name: string } }) {
				useEffect(() => { console.log(u.name); }, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'u.name'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			function MyComponent({ user: u }: { user: { name: string } }) {
				useEffect(() => { console.log(u.name); }, [u.name]);
			}
		`}},
			},
		},
	},

	// Tuple destructure — TC may resolve `num` to the tuple element type.
	{
		Code: `
			declare function useTuple(): readonly [number, string];
			function MyHook() {
				const [num, str] = useTuple();
				void str;
				return useCallback(() => num, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useCallback has a missing dependency: 'num'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			declare function useTuple(): readonly [number, string];
			function MyHook() {
				const [num, str] = useTuple();
				void str;
				return useCallback(() => num, [num]);
			}
		`}},
			},
		},
	},

	// Renamed destructure with default value, missing in deps.
	{
		Code: `
			declare function useData(): { value?: number };
			function MyHook() {
				const { value: v = 0 } = useData();
				return useCallback(() => v, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useCallback has a missing dependency: 'v'. Either include it or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			declare function useData(): { value?: number };
			function MyHook() {
				const { value: v = 0 } = useData();
				return useCallback(() => v, [v]);
			}
		`}},
			},
		},
	},

	// Two renamed destructures — both bound names captured via property
	// chain, dep key is full `aliasA.name` / `aliasB.name`.
	{
		Code: `
			declare function useShape(): { a: { name: string }; b: { name: string } };
			function MyHook() {
				const { a: aliasA, b: aliasB } = useShape();
				return useCallback(() => aliasA.name + aliasB.name, []);
			}
		`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useCallback has missing dependencies: 'aliasA.name' and 'aliasB.name'. Either include them or remove the dependency array.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
			declare function useShape(): { a: { name: string }; b: { name: string } };
			function MyHook() {
				const { a: aliasA, b: aliasB } = useShape();
				return useCallback(() => aliasA.name + aliasB.name, [aliasA.name, aliasB.name]);
			}
		`}},
			},
		},
	},
}
