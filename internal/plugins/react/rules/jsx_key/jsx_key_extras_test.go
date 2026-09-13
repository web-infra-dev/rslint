package jsx_key

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxKeyExtras(t *testing.T) {
	checkFragmentShorthand := map[string]interface{}{"checkFragmentShorthand": true}
	checkKeyMustBeforeSpread := map[string]interface{}{"checkKeyMustBeforeSpread": true}
	warnOnDuplicates := map[string]interface{}{"warnOnDuplicates": true}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxKeyRule, []rule_tester.ValidTestCase{
		{
			Code: `
[1, 2, 3].map((item) => {
  return item === 'bar' ? <div key={item}>{item}</div> : <span key={item}>{item}</span>;
})
`,
			Tsx: true,
		},
		{Code: `[1, 2, 3].map(x => { return x && <App key={x} />; });`, Tsx: true},
		{Code: `[1, 2, 3].map(x => { return x && y && <App key={x} />; });`, Tsx: true},
		{Code: `[1, 2, 3].map(x => { return x && foo(); });`, Tsx: true},

		// ---- Additional edge cases ----
		// Nested JSX: keys are not required on direct JSX children of a
		// JSX parent.
		{Code: `<ul><li/><li/></ul>;`, Tsx: true},
		// Array.from with no callback (non-function 2nd arg) — no report.
		{Code: `Array.from([1, 2, 3], "not a function");`, Tsx: true},
		// `xs?.map(...)` — optional chaining on the call. Same behavior as
		// `xs.map(...)`.
		{Code: `xs?.map(x => <App key={x} />);`, Tsx: true},
		// `xs.map?.(...)` — optional call. Same behavior.
		{Code: `xs.map?.(x => <App key={x} />);`, Tsx: true},
		// Parenthesized arrow callback.
		{Code: `[1,2,3].map((x => <App key={x} />));`, Tsx: true},
		// Spread element alongside keyed JSX in an array — spread is not a
		// JsxElement, only the keyed element is inspected.
		{Code: `const rest = []; [<App key="k" />, ...rest];`, Tsx: true},
		// Omitted array element (hole) is not a JsxElement.
		{Code: `[, <App key="k" />];`, Tsx: true},
		// warnOnDuplicates: distinct literal texts do not trigger.
		{
			Code: `
const spans = [
  <span key="a"/>,
  <span key="b"/>,
];
`,
			Tsx:     true,
			Options: warnOnDuplicates,
		},
		// checkKeyMustBeforeSpread false (default): key after spread in an
		// array is NOT flagged.
		{Code: `[<App {...obj} key="k" />, <App key="k2" />];`, Tsx: true},
		// checkFragmentShorthand false (default): fragment in iterator is
		// not flagged.
		{Code: `[1, 2, 3].map(x => <>{x}</>);`, Tsx: true},
		// checkFragmentShorthand false: fragment in array literal is not
		// flagged.
		{Code: `[<></>];`, Tsx: true},
		// Nested React.Children.toArray — both the inner Children.toArray
		// and the deepest map are skipped.
		{
			Code: `
React.Children.toArray(
  React.Children.toArray([1, 2, 3].map(x => <App />))
);
`,
			Tsx: true,
		},
		// Children.toArray balanced with an adjacent unguarded map: only
		// the unguarded one would normally flag, but it's keyed — no report.
		{
			Code: `
React.Children.toArray([1, 2, 3].map(x => <A />));
[1, 2, 3].map(x => <B key={x} />);
`,
			Tsx: true,
		},
		// Fragment appearing as a JSX child (not an array element) — even
		// with checkFragmentShorthand on, upstream only reports array-level
		// fragments. Locks that behavior.
		{
			Code:    `<div><></><></></div>;`,
			Tsx:     true,
			Options: checkFragmentShorthand,
		},
		// Map callback with async arrow — still an ArrowFunction, still
		// subject to the iterator check. Keyed → no report.
		{Code: `[1, 2, 3].map(async x => <App key={x} />);`, Tsx: true},
		// Deeply nested parens around the arrow callback.
		{Code: `[1, 2, 3].map((((x => <App key={x} />))));`, Tsx: true},
		// Bracket access `xs["map"]` is NOT matched by the rule (upstream
		// uses `callee.property.name="map"`, i.e. dot access only).
		{Code: `xs["map"](x => <App />);`, Tsx: true},
		// Chained calls: `xs.filter(...).map(x => <A key={x} />)` — keyed.
		{Code: `xs.filter(x => x > 0).map(x => <App key={x} />);`, Tsx: true},
		// Array.from with a non-function second argument (identifier/any).
		{Code: `Array.from([1, 2, 3], 123);`, Tsx: true},
		// Generic call — `.map<T>(...)` still a PropertyAccessExpression.
		{Code: `xs.map<number>(x => <App key={x} />);`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// Intentional divergence: upstream only peels conditional/logical expression
		// bodies, but these block callbacks also return JSX without keys.
		{
			Code: `
[1, 2, 3].map((item) => {
  return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;
})`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `
[1, 2, 3].map(function(item) {
  return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;
})`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `
Array.from([1, 2, 3], (item) => {
  return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;
})`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `
import { Fragment } from 'react';

const ITEMS = ['bar', 'foo'];

export default function BugIssue() {
  return (
    <Fragment>
      {ITEMS.map((item) => {
        return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;
      })}
    </Fragment>
  );
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `[1, 2, 3].map(x => { return x && <App />; });`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `[1, 2, 3].map(x => { return x || y || <App />; });`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},

		// ---- Additional edge cases ----
		// Line / column coverage on a simple array-missing-key case.
		{
			Code: "\n[\n  <App />,\n  <App />,\n];\n",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArrayKey", Line: 3, Column: 3},
				{MessageId: "missingArrayKey", Line: 4, Column: 3},
			},
		},
		// Full message-text assertion on a plain missingIterKey.
		{
			Code: `[1].map(x => <App />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingIterKey",
					Message:   `Missing "key" prop for element in iterator`,
					Line:      1,
					Column:    14,
				},
			},
		},
		// Full message-text assertion on missingArrayKey.
		{
			Code: `[<App />];`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingArrayKey",
					Message:   `Missing "key" prop for element in array`,
					Line:      1,
					Column:    2,
				},
			},
		},
		// Default pragma — missingIterKeyUsePrag uses "React.Fragment".
		{
			Code:    `[1, 2, 3].map(x => <>{x}</>);`,
			Tsx:     true,
			Options: checkFragmentShorthand,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingIterKeyUsePrag",
					Message:   `Missing "key" prop for element in iterator. Shorthand fragment syntax does not support providing keys. Use React.Fragment instead`,
				},
			},
		},
		// Default pragma — missingArrayKeyUsePrag uses "React.Fragment".
		{
			Code:    `[<></>];`,
			Tsx:     true,
			Options: checkFragmentShorthand,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingArrayKeyUsePrag",
					Message:   `Missing "key" prop for element in array. Shorthand fragment syntax does not support providing keys. Use React.Fragment instead`,
				},
			},
		},
		// Optional chaining on both the member access and the call.
		{
			Code: `[1, 2, 3].map?.(x => <App />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `xs?.map(x => <App />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		// Parenthesized arrow callback inside Array.from — ESTree flattens
		// the parens; tsgo preserves them. Regression case.
		{
			Code: `Array.from([1, 2, 3], ((x => <App />)));`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		// Parenthesized JSX body of arrow map callback.
		{
			Code: `[1,2,3].map(x => (<App />));`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		// warnOnDuplicates with checkKeyMustBeforeSpread both on —
		// independent diagnostics on the same element.
		{
			Code: `
const spans = [
  <span {...o} key="a"/>,
  <span key="a"/>,
];
`,
			Tsx: true,
			Options: map[string]interface{}{
				"warnOnDuplicates":         true,
				"checkKeyMustBeforeSpread": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "keyBeforeSpread"},
				{MessageId: "nonUniqueKeys", Line: 3},
				{MessageId: "nonUniqueKeys", Line: 4},
			},
		},
		// warnOnDuplicates groups keys by raw source text — `{'a'}` and
		// `"a"` are different texts, so they're NOT duplicates (matches
		// upstream's `getText` behavior).
		{
			Code: `
const spans = [
  <span key={'a'}/>,
  <span key="a"/>,
  <span key={'a'}/>,
];
`,
			Tsx:     true,
			Options: warnOnDuplicates,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonUniqueKeys", Line: 3},
				{MessageId: "nonUniqueKeys", Line: 5},
			},
		},
		// ---- Additional container / positioning coverage ----
		// Map callback missing key: assert precise Line/Column/EndLine/
		// EndColumn for the JSX element in a multi-line case.
		{
			Code: "\n[1, 2, 3].map(x =>\n  <App />\n);\n",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingIterKey",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 10,
				},
			},
		},
		// Array-missing-key: precise end line/column on a multi-line element.
		{
			Code: "\n[\n  <App\n    x={1}\n  />,\n];\n",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingArrayKey",
					Line:      3,
					Column:    3,
					EndLine:   5,
					EndColumn: 5,
				},
			},
		},
		// keyBeforeSpread in array context reports on the array, not the
		// offending element.
		{
			Code:    "\n[\n  <App {...x} key=\"k\" />,\n];\n",
			Tsx:     true,
			Options: checkKeyMustBeforeSpread,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "keyBeforeSpread",
					Line:      2,
					Column:    1,
				},
			},
		},
		// Three-way independence: keys across sibling arrays don't collide.
		// The outer array has JSX children (the inner arrays) that are
		// themselves arrays, not JSX — no missing-array-key on the outer.
		// Each inner array evaluates its own keys independently.
		{
			Code: `
const nested = [
  [<A key="x" />, <A />],
  [<A key="x" />],
];
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArrayKey", Line: 3, Column: 19},
			},
		},
		// ---- Upstream nested-`Children.toArray` boolean-flag quirk ----
		//
		// eslint-plugin-react uses a plain boolean `isWithinChildrenToArray`
		// that gets toggled on/off per Children.toArray enter/exit. When TWO
		// Children.toArray calls both lexically enclose some code, the inner
		// call's exit clobbers the flag even though the outer is still
		// enclosing. Subsequent map/from calls inside the outer are then
		// treated as "not inside toArray" and reported.
		//
		// These cases reproduce the quirk and lock our output 1:1 with
		// upstream ESLint — verified by running eslint-plugin-react@7
		// against these exact inputs. A depth-counter implementation would
		// silently diverge (zero reports); do not "fix" the boolean.
		{
			Code: `
React.Children.toArray([
  React.Children.toArray(a),
  xs.map(x => <A/>),
]);
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey", Line: 4, Column: 15},
			},
		},
		{
			Code: `
React.Children.toArray(
  bar(
    React.Children.toArray(a),
    xs.map(y => <A/>)
  )
);
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey", Line: 5, Column: 17},
			},
		},
		// Multiple JSX children of a JSX parent, duplicate keys across
		// spread/non-spread siblings. checkKeyMustBeforeSpread catches the
		// spread violator; warnOnDuplicates clusters the identical keys.
		{
			Code: `
const div = (
  <div>
    <span {...o} key="a" />
    <span key="a" />
  </div>
);
`,
			Tsx: true,
			Options: map[string]interface{}{
				"warnOnDuplicates":         true,
				"checkKeyMustBeforeSpread": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "keyBeforeSpread"},
				{MessageId: "nonUniqueKeys", Line: 4},
				{MessageId: "nonUniqueKeys", Line: 5},
			},
		},
	})
}

// Regression coverage for parenthesized siblings and file-local runtime pragmas.
func TestJsxKeyParityRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&JsxKeyRule,
		[]rule_tester.ValidTestCase{
			// jsx-key/probe/keyed-negative
			{
				Code: "[(<X key=\"a\"/>), ((<Y key=\"b\"/>))];",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
			},
			// jsx-key/probe/duplicates-default
			{
				Code: "[(<X key=\"a\"/>), (<Y key=\"a\"/>)];",
				Tsx:  true,
			},
			// jsx-key/probe/spread-order-negative
			{
				Code: "[(<X key=\"a\" {...p}/>)];",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
			},
			// jsx-key/probe/nested-arrays
			{
				Code: "[[(<X key=\"a\"/>)], [(<Y key=\"a\"/>)]];",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
			},
			// jsx-key/probe/no-conditional-unwrapping
			{
				Code: "[(ok ? <X/> : <Y/>)];",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
			},
			// jsx-key/probe/jsx-expression-boundary
			{
				Code: "<div>{(<X key=\"a\"/>)}<Y key=\"a\"/></div>;",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
			},
			// jsx-key/probe/array-toArray-exemption
			{
				Code: "React.Children.toArray([(<X/>), (<></>)]);",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
			},
			// jsx-key/probe/pragma-toArray
			{
				Code: "/** @jsx Preact.h */ Preact.Children.toArray(xs.map(x=><X/>));",
				Tsx:  true,
			},
			// jsx-key/probe/pragma-overrides-settings
			{
				Code: "/** @jsx Preact.h */ Preact.Children.toArray(xs.map(x=><X/>));",
				Tsx:  true,
				Settings: map[string]interface {
				}{
					"react": map[string]interface {
					}{
						"pragma":   "Act",
						"fragment": "Frag",
					},
				},
			},
			// jsx-key/probe/pragma-fragment-disabled
			{
				Code: "/** @jsx Preact.h */ [(<></>)];",
				Tsx:  true,
			},
			// jsx-key/probe/pragma-first-comment
			{
				Code: "/** @jsx First.h */ /** @jsx Second.h */ First.Children.toArray([(<X/>)]);",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
			},
			// jsx-key/probe/pragma-line-comment
			{
				Code: `// @jsx Preact.h
Preact.Children.toArray(Array.from(xs,x=><X/>));`,
				Tsx: true,
			},
			// jsx-key/probe/pragma-string-not-comment
			{
				Code: "const text=\"@jsx Preact.h\"; React.Children.toArray([(<X/>)]);",
				Tsx:  true,
			},
			// jsx-key/probe/pragma-jsx-text-not-comment
			{
				Code: "<div>@jsx Preact.h</div>; React.Children.toArray([(<X/>)]);",
				Tsx:  true,
			},
			// jsx-key/probe/jsxFrag-alone-not-runtime
			{
				Code: "/** @jsxFrag Other */ Act.Children.toArray([(<X/>)]);",
				Tsx:  true,
				Settings: map[string]interface {
				}{
					"react": map[string]interface {
					}{
						"pragma":   "Act",
						"fragment": "Frag",
					},
				},
			},
			// jsx-key/probe/settings-only
			{
				Code: "Act.Children.toArray([(<X/>)]);",
				Tsx:  true,
				Settings: map[string]interface {
				}{
					"react": map[string]interface {
					}{
						"pragma":   "Act",
						"fragment": "Frag",
					},
				},
			},
			// jsx-key/probe/pragma-invalid-fallback
			{
				Code: "/** @jsx 123 */ React.Children.toArray([(<X/>)]);",
				Tsx:  true,
				Settings: map[string]interface {
				}{
					"react": map[string]interface {
					}{
						"pragma":   "Act",
						"fragment": "Frag",
					},
				},
			},
			// jsx-key/probe/pragma-ecma-whitespace
			{
				Code: "/** @jsx Preact.h */ Preact.Children.toArray([(<X/>)]);",
				Tsx:  true,
			},
			// jsx-key/probe/pragma-after-code
			{
				Code: "foo(); /** @jsx Preact.h */ Preact.Children.toArray([(<X/>)]);",
				Tsx:  true,
			},
			// jsx-key/probe/destructured-children
			{
				Code: "/** @jsx Preact.h */ Children.toArray([(<X/>)]);",
				Tsx:  true,
			},
			// Computed identifier spelling cannot establish a known runtime method.
			{
				Code: "xs[map](x=><X/>);",
				Tsx:  true,
			},
		},
		[]rule_tester.InvalidTestCase{
			// jsx-key/probe/parenthesized-missing
			{
				Code: "[(<X/>)];",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingArrayKey",
						Message:   "Missing \"key\" prop for element in array",
						Line:      1,
						Column:    3,
						EndLine:   1,
						EndColumn: 7,
					},
				},
			},
			// jsx-key/probe/deep-parenthesized-paired
			{
				Code: "[(( /* comment */ <X></X>))];",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingArrayKey",
						Message:   "Missing \"key\" prop for element in array",
						Line:      1,
						Column:    19,
						EndLine:   1,
						EndColumn: 26,
					},
				},
			},
			// jsx-key/probe/multiline-trivia
			{
				Code: `[
 (/* trivia */
 <X/>
 )
];`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingArrayKey",
						Message:   "Missing \"key\" prop for element in array",
						Line:      3,
						Column:    2,
						EndLine:   3,
						EndColumn: 6,
					},
				},
			},
			// jsx-key/probe/unicode-prefix
			{
				Code: "const text=\"𐐀\"; [(<X/>)];",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingArrayKey",
						Message:   "Missing \"key\" prop for element in array",
						Line:      1,
						Column:    20,
						EndLine:   1,
						EndColumn: 24,
					},
				},
			},
			// jsx-key/probe/duplicate-mixed
			{
				Code: "[(<X key=\"a\"/>), <Y key=\"a\"/>];",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "nonUniqueKeys",
						Message:   "`key` prop must be unique",
						Line:      1,
						Column:    6,
						EndLine:   1,
						EndColumn: 13,
					},
					{
						MessageId: "nonUniqueKeys",
						Message:   "`key` prop must be unique",
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 28,
					},
				},
			},
			// jsx-key/probe/duplicate-all-parenthesized
			{
				Code: "[((<X key=\"a\"/>)), (<Y key=\"a\"/>)];",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "nonUniqueKeys",
						Message:   "`key` prop must be unique",
						Line:      1,
						Column:    7,
						EndLine:   1,
						EndColumn: 14,
					},
					{
						MessageId: "nonUniqueKeys",
						Message:   "`key` prop must be unique",
						Line:      1,
						Column:    24,
						EndLine:   1,
						EndColumn: 31,
					},
				},
			},
			// jsx-key/probe/all-checks
			// The Go harness keeps listener emission order; the public API sorts
			// these same diagnostics into source order, as ESLint does.
			{
				Code: "[(<X {...p} key=\"a\"/>), (<Y key=\"a\"/>), (<Z/>), (<></>)];",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "keyBeforeSpread",
						Message:   "`key` prop must be placed before any `{...spread}, to avoid conflicting with React’s new JSX transform: https://reactjs.org/blog/2020/09/22/introducing-the-new-jsx-transform.html`",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 57,
					},
					{
						MessageId: "missingArrayKey",
						Message:   "Missing \"key\" prop for element in array",
						Line:      1,
						Column:    42,
						EndLine:   1,
						EndColumn: 46,
					},
					{
						MessageId: "nonUniqueKeys",
						Message:   "`key` prop must be unique",
						Line:      1,
						Column:    13,
						EndLine:   1,
						EndColumn: 20,
					},
					{
						MessageId: "nonUniqueKeys",
						Message:   "`key` prop must be unique",
						Line:      1,
						Column:    29,
						EndLine:   1,
						EndColumn: 36,
					},
					{
						MessageId: "missingArrayKeyUsePrag",
						Message:   "Missing \"key\" prop for element in array. Shorthand fragment syntax does not support providing keys. Use React.Fragment instead",
						Line:      1,
						Column:    50,
						EndLine:   1,
						EndColumn: 55,
					},
				},
			},
			// jsx-key/probe/holes-spread
			{
				Code: "[, ...xs, (<X/>), null];",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingArrayKey",
						Message:   "Missing \"key\" prop for element in array",
						Line:      1,
						Column:    12,
						EndLine:   1,
						EndColumn: 16,
					},
				},
			},
			// jsx-key/probe/pragma-rejects-old-setting
			{
				Code: "/** @jsx Preact.h */ Act.Children.toArray(xs.map(x=><X/>));",
				Tsx:  true,
				Settings: map[string]interface {
				}{
					"react": map[string]interface {
					}{
						"pragma":   "Act",
						"fragment": "Frag",
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKey",
						Message:   "Missing \"key\" prop for element in iterator",
						Line:      1,
						Column:    53,
						EndLine:   1,
						EndColumn: 57,
					},
				},
			},
			// jsx-key/probe/pragma-rejects-default
			{
				Code: "/** @jsx Preact.h */ React.Children.toArray(xs.map(x=><X/>));",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKey",
						Message:   "Missing \"key\" prop for element in iterator",
						Line:      1,
						Column:    55,
						EndLine:   1,
						EndColumn: 59,
					},
				},
			},
			// jsx-key/probe/pragma-array-fragment
			{
				Code: "/** @jsx Preact.h */ [(<></>)];",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingArrayKeyUsePrag",
						Message:   "Missing \"key\" prop for element in array. Shorthand fragment syntax does not support providing keys. Use Preact.Fragment instead",
						Line:      1,
						Column:    24,
						EndLine:   1,
						EndColumn: 29,
					},
				},
			},
			// jsx-key/probe/pragma-iterator-fragment
			{
				Code: "/** @jsx Preact.h */ xs.map(x=><></>);",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKeyUsePrag",
						Message:   "Missing \"key\" prop for element in iterator. Shorthand fragment syntax does not support providing keys. Use Preact.Fragment instead",
						Line:      1,
						Column:    32,
						EndLine:   1,
						EndColumn: 37,
					},
				},
			},
			// jsx-key/probe/fragment-setting-independent
			{
				Code: "/** @jsx Preact.h */ /** @jsxFrag Other */ [(<></>)];",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
				Settings: map[string]interface {
				}{
					"react": map[string]interface {
					}{
						"pragma":   "Act",
						"fragment": "Frag",
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingArrayKeyUsePrag",
						Message:   "Missing \"key\" prop for element in array. Shorthand fragment syntax does not support providing keys. Use Preact.Frag instead",
						Line:      1,
						Column:    46,
						EndLine:   1,
						EndColumn: 51,
					},
				},
			},
			// jsx-key/probe/iterator-fragment-setting-independent
			{
				Code: "/** @jsx Preact.h */ xs.map(x=><></>);",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates":         true,
					"checkKeyMustBeforeSpread": true,
					"checkFragmentShorthand":   true,
				},
				Settings: map[string]interface {
				}{
					"react": map[string]interface {
					}{
						"pragma":   "Act",
						"fragment": "Frag",
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKeyUsePrag",
						Message:   "Missing \"key\" prop for element in iterator. Shorthand fragment syntax does not support providing keys. Use Preact.Frag instead",
						Line:      1,
						Column:    32,
						EndLine:   1,
						EndColumn: 37,
					},
				},
			},
			// jsx-key/probe/pragma-exit-state
			{
				Code: "/** @jsx Preact.h */ Preact.Children.toArray([(<X/>)]); [(<Y/>)];",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingArrayKey",
						Message:   "Missing \"key\" prop for element in array",
						Line:      1,
						Column:    59,
						EndLine:   1,
						EndColumn: 63,
					},
				},
			},
			// Dynamic toArray is not a proven Children.toArray exemption.
			{
				Code: "React.Children[toArray](xs.map(x=><X/>));",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKey",
					},
				},
			},
		})
}
