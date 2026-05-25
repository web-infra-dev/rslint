package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDepsRule_Positions exercises the diagnostic position
// (Line / Column / EndLine / EndColumn) for every distinct diagnostic
// shape the rule emits. Position errors don't surface in the
// upstream-ported tests (upstream rarely asserts line/column), so we
// dedicate a separate suite that checks each report site to detect
// silent regressions in the trimmed-range output.
//
// Each case carries a single error whose location is asserted to the
// 4-tuple (Line, Column, EndLine, EndColumn). The case body is shaped
// so the start/end line numbers are obvious from the raw template.
func TestExhaustiveDepsRule_Positions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		nil,
		positionInvalidCases,
	)
}

var positionInvalidCases = []rule_tester.InvalidTestCase{
	// Missing dep — the diagnostic is anchored at the deps array node.
	{
		Code: "function C({a}) {\n  useEffect(() => { a; }, []);\n}\n",
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a missing dependency: 'a'. Either include it or remove the dependency array.",
				Line:    2, Column: 27, EndLine: 2, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "function C({a}) {\n  useEffect(() => { a; }, [a]);\n}\n"}},
			},
		},
	},

	// Spread element — anchored at the spread node.
	{
		Code: "function C({list}) {\n  useEffect(() => {}, [...list]);\n}\n",
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect has a spread element in its dependency array. This means we can't statically verify whether you've passed the correct dependencies.",
				Line:    2, Column: 24, EndLine: 2, EndColumn: 31,
			},
		},
	},

	// Non-array deps — anchored at the deps argument expression.
	{
		Code: "function C({a}) {\n  useEffect(() => {}, a);\n}\n",
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect was passed a dependency list that is not an array literal. This means we can't statically verify whether you've passed the correct dependencies.",
				Line:    2, Column: 23, EndLine: 2, EndColumn: 24,
			},
		},
	},

	// Async effect callback — anchored at the callback function node.
	{
		Code: "function C() {\n  useEffect(async () => { await foo(); }, []);\n}\n",
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Line: 2, Column: 13, EndLine: 2, EndColumn: 41,
			},
		},
	},

	// Missing callback — anchored at the hook callee.
	{
		Code: "function C() {\n  useEffect();\n}\n",
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useEffect requires an effect callback. Did you forget to pass a callback to the hook?",
				Line:    2, Column: 3, EndLine: 2, EndColumn: 12,
			},
		},
	},

	// Literal dep — anchored at the literal element.
	{
		Code: "function C() {\n  useEffect(() => {}, ['foo']);\n}\n",
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "The 'foo' literal is not a valid dependency because it never changes. You can safely remove it.",
				Line:    2, Column: 24, EndLine: 2, EndColumn: 29,
			},
		},
	},

	// useMemo without deps — anchored at the hook callee.
	{
		Code: "function C({a}) {\n  useMemo(() => a);\n}\n",
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Message: "React Hook useMemo does nothing when called with only one argument. Did you forget to pass an array of dependencies?",
				Line:    2, Column: 3, EndLine: 2, EndColumn: 10,
			},
		},
	},

	// ref.current in cleanup — anchored at the property name node.
	{
		Code: "function C() {\n  const ref = useRef(null);\n  useEffect(() => {\n    return () => { ref.current; };\n  }, []);\n}\n",
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				Line: 4, Column: 24, EndLine: 4, EndColumn: 31,
			},
		},
	},
}
