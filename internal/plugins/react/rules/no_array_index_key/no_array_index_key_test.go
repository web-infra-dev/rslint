package no_array_index_key

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoArrayIndexKeyRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoArrayIndexKeyRule, []rule_tester.ValidTestCase{
		// ---- Outside of any tracked iterator, every shape is fine ----
		{Code: `<Foo key="foo" />;`, Tsx: true},
		{Code: `<Foo key={i} />;`, Tsx: true},
		{Code: `<Foo key />;`, Tsx: true},
		{Code: `<Foo key={` + "`" + `foo-${i}` + "`" + `} />;`, Tsx: true},
		{Code: `<Foo key={'foo-' + i} />;`, Tsx: true},

		// ---- Untracked iterators (not in indexParamPositions) ----
		{Code: `foo.bar((baz, i) => <Foo key={i} />)`, Tsx: true},
		{Code: `foo.bar((bar, i) => <Foo key={` + "`" + `foo-${i}` + "`" + `} />)`, Tsx: true},
		{Code: `foo.bar((bar, i) => <Foo key={'foo-' + i} />)`, Tsx: true},

		// ---- Tracked iterators, but key does NOT reference the index ----
		{Code: `foo.map((baz) => <Foo key={baz.id} />)`, Tsx: true},
		{Code: `foo.map((baz, i) => <Foo key={baz.id} />)`, Tsx: true},
		{Code: `foo.map((baz, i) => <Foo key={'foo' + baz.id} />)`, Tsx: true},
		{Code: `foo.map((baz, i) => React.cloneElement(someChild, { ...someChild.props }))`, Tsx: true},
		{Code: `foo.map((baz, i) => cloneElement(someChild, { ...someChild.props }))`, Tsx: true},
		{
			Code: `
foo.map((item, i) => {
  return React.cloneElement(someChild, {
    key: item.id
  })
})
`,
			Tsx: true,
		},
		{
			Code: `
foo.map((item, i) => {
  return cloneElement(someChild, {
    key: item.id
  })
})
`,
			Tsx: true,
		},
		{Code: `foo.map((baz, i) => <Foo key />)`, Tsx: true},
		{Code: `foo.reduce((a, b) => a.concat(<Foo key={b.id} />), [])`, Tsx: true},

		// ---- Index identifier appears, but not as a tracked-by-rule shape ----
		// `i.baz.toString()` — only the direct `index.toString()` shape is
		// tracked; deeper member accesses fall through.
		{Code: `foo.map((bar, i) => <Foo key={i.baz.toString()} />)`, Tsx: true},
		// `i.toString` — accessing the method without invoking it.
		{Code: `foo.map((bar, i) => <Foo key={i.toString} />)`, Tsx: true},
		// `String()` — empty argument list.
		{Code: `foo.map((bar, i) => <Foo key={String()} />)`, Tsx: true},
		// `String(baz)` — non-index argument.
		{Code: `foo.map((bar, i) => <Foo key={String(baz)} />)`, Tsx: true},

		// ---- Tracked iterators with insufficient params for the index slot ----
		{Code: `foo.flatMap((a) => <Foo key={a} />)`, Tsx: true},
		{Code: `foo.reduce((a, b, i) => a.concat(<Foo key={b.id} />), [])`, Tsx: true},
		{Code: `foo.reduceRight((a, b) => a.concat(<Foo key={b.id} />), [])`, Tsx: true},
		{Code: `foo.reduceRight((a, b, i) => a.concat(<Foo key={b.id} />), [])`, Tsx: true},

		// ---- React.Children.* / Children.* — callback at args[1], non-index key ----
		{
			Code: `
React.Children.map(this.props.children, (child, index, arr) => {
  return React.cloneElement(child, { key: child.id });
})
`,
			Tsx: true,
		},
		{
			Code: `
React.Children.map(this.props.children, (child, index, arr) => {
  return cloneElement(child, { key: child.id });
})
`,
			Tsx: true,
		},
		{
			Code: `
Children.forEach(this.props.children, (child, index, arr) => {
  return React.cloneElement(child, { key: child.id });
})
`,
			Tsx: true,
		},
		{
			Code: `
Children.forEach(this.props.children, (child, index, arr) => {
  return cloneElement(child, { key: child.id });
})
`,
			Tsx: true,
		},

		// ---- Optional chaining iterator with non-index key ----
		{Code: `foo?.map(child => <Foo key={child.i} />)`, Tsx: true},

		// ---- tsgo edge shapes that upstream's tests don't exercise ----
		// Parenthesized index reference.
		{Code: `foo.map((bar, i) => <Foo key={(bar.i)} />)`, Tsx: true},
		// Optional-chain receiver in `index?.toString()` — upstream's
		// `astUtil.isCallExpression` accepts both regular and optional
		// call shapes; we lock in that the rule still treats this as
		// the toString() coercion form (an array index escape route).
		{Code: `foo.map((bar, idx) => <Foo key={String('a' + bar.id)} />)`, Tsx: true},
		// Bare `cloneElement` callee with no `import { cloneElement }
		// from 'react'` — upstream returns false for the bare-identifier
		// branch when the binding doesn't resolve to the pragma module,
		// so no report.
		{
			Code: `
const cloneElement = (...args) => args;
foo.map((baz, i) => cloneElement(someChild, { key: i }))
`,
			Tsx: true,
		},
		// ---- Identifier branch only accepts ImportSpecifier ----
		// Upstream `isCreateCloneElement` Identifier branch is gated on
		// `variable.type === 'ImportSpecifier'`. Each shape below
		// produces a non-ImportSpecifier binding and therefore is NOT
		// recognized as a React factory — green-path locks.
		{
			Code: `
const { cloneElement } = React;
foo.map((baz, i) => cloneElement(someChild, { key: i }))
`,
			Tsx: true,
		},
		{
			Code: `
const cloneElement = React.cloneElement;
foo.map((baz, i) => cloneElement(someChild, { key: i }))
`,
			Tsx: true,
		},
		{
			Code: `
const { cloneElement } = require('react');
foo.map((baz, i) => cloneElement(someChild, { key: i }))
`,
			Tsx: true,
		},
		// Module specifier is hardcoded `'react'` upstream — pragma
		// configuration does NOT change it. With pragma "Act" but
		// import from 'act', the bare callee is NOT recognized.
		{
			Code: `
import { cloneElement } from 'act';
foo.map((baz, i) => cloneElement(someChild, { key: i }))
`,
			Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"pragma": "Act"},
			},
		},

		// ---- Universal edge shapes that upstream tests don't cover ----
		// Tagged template — tsgo's KindTaggedTemplateExpression is NOT
		// KindTemplateExpression, so the substitution-walk branch never
		// fires. Lock the negative.
		{Code: `foo.map((bar, i) => <Foo key={tag` + "`" + `foo-${i}` + "`" + `} />)`, Tsx: true},
		// No-substitution template — KindNoSubstitutionTemplateLiteral,
		// not TemplateExpression; even though the source contains an
		// `i`-shaped sequence as text, there is no expression substitution.
		{Code: `foo.map((bar, i) => <Foo key={` + "`" + `foo-i` + "`" + `} />)`, Tsx: true},
		// Conditional inside a binary chain — upstream's flatten only
		// recurses through BinaryExpression, so the index reference
		// inside the ternary is opaque and does NOT report.
		{Code: `foo.map((bar, i) => <Foo key={'a' + (i ? 'x' : 'y')} />)`, Tsx: true},
		// Identity-wrapped index — `identity(i)` is a CallExpression
		// whose callee is neither `String` nor an `index.toString()`
		// shape; not reported.
		{Code: `foo.map((bar, i) => <Foo key={identity(i)} />)`, Tsx: true},
		// Object-literal as the key value (`key: { nested: i }`) —
		// upstream's checkPropValue has no ObjectLiteral branch.
		{Code: `foo.map((bar, i) => React.createElement('Foo', { key: { nested: i } }))`, Tsx: true},
		// `String['toString'](i)` — bracket access doesn't match the
		// dotted `index.toString()` shape and the callee identifier
		// is `String`, so we'd fall into the String branch — but the
		// receiver isn't an Identifier referencing the index, so we
		// stay in valid territory.
		{Code: `foo.map((bar, i) => <Foo key={i['toString']()} />)`, Tsx: true},
		// Destructured index parameter `(bar, [i])` — upstream pushes
		// `undefined`; we return "" and skip. Either way, `i` inside
		// the JSX is NOT recognized as the tracked index because the
		// stack does not contain a matching name.
		{Code: `foo.map((bar, [i]) => <Foo key={i} />)`, Tsx: true},
		// Rest-only index slot `(bar, ...rest)` — pos overshoots; no push.
		{Code: `foo.map((bar, ...rest) => <Foo key={rest[0]} />)`, Tsx: true},
		// String() on a deeper member access (`String(i.foo)`) —
		// upstream's `node.arguments[0]` must itself be the index
		// identifier; a member-access argument is NOT a match.
		{Code: `foo.map((bar, i) => <Foo key={String(i.foo)} />)`, Tsx: true},

		// ---- TS expression wrappers on the pragma identifier / key value ----
		// ESLint's JavaScript parser cannot produce `as` / `!` / `<T>x` /
		// `satisfies` shapes, so its `no-array-index-key` never sees
		// them. We mirror that exactly:
		//   - the pragma identifier behind a TS wrapper is NOT
		//     recognized as the React factory (the dotted access fails
		//     the Identifier check);
		//   - the JSX key value behind a TS wrapper is opaque to
		//     `checkPropValue`'s direct-Identifier branch.
		// Iterator receiver shape is intentionally NOT included here —
		// upstream's `getMapIndexParamName` doesn't inspect the
		// receiver at all (only `callee.property.name`), so
		// `(foo as any[]).map(...)` still fires; that case is in the
		// invalid suite below.
		{Code: `foo.map((bar, i) => (React as any).cloneElement(c, { key: i }))`, Tsx: true},
		{Code: `foo.map((bar, i) => React!.createElement('Foo', { key: i }))`, Tsx: true},
		{Code: `foo.map((bar, i) => <Foo key={(i as any)} />)`, Tsx: true},

		// ---- Optional-chain CallExpression as the key value ----
		// Upstream `astUtil.isCallExpression(node)` rejects
		// `OptionalCallExpression`, and `node.callee.type === 'MemberExpression'`
		// rejects `OptionalMemberExpression`. Three opaque shapes:
		//   - `i?.toString()` (optional member callee)
		//   - `String?.(i)` (optional call)
		//   - `i?.toString?.()` (both)
		// Each is a green-path lock.
		{Code: `foo.map((bar, i) => <Foo key={i?.toString()} />)`, Tsx: true},
		{Code: `foo.map((bar, i) => <Foo key={String?.(i)} />)`, Tsx: true},
		{Code: `foo.map((bar, i) => <Foo key={i?.toString?.()} />)`, Tsx: true},

		// ---- Real user scenarios that should NOT report ----
		// React.createElement / cloneElement OUTSIDE any iterator —
		// stack is empty; the props-key check short-circuits.
		{Code: `React.createElement('Foo', { key: i })`, Tsx: true},
		{Code: `React.cloneElement(child, { key: i })`, Tsx: true},
		// Iterator with a `thisArg` third argument — `args[0]` is the
		// callback, the third arg is irrelevant. Index named `i` lives
		// on the stack inside the callback.
		{Code: `foo.map((bar) => <Foo key={bar.id} />, this)`, Tsx: true},
		// Iterator on an empty array — still recognized as map; index
		// not referenced, so no report.
		{Code: `[].map((bar, i) => <Foo key={bar.id} />)`, Tsx: true},
		// Computed-key with StringLiteral inside (`["key"]: i`) —
		// upstream's `prop.key.name` reads the inner Literal, which
		// has no `.name` field, so `=== 'key'` fails. NOT reported.
		{Code: `foo.map((bar, i) => React.createElement('Foo', { ["key"]: i }))`, Tsx: true},
		// String-literal key `"key"` — same opaque shape, no report.
		{Code: `foo.map((bar, i) => React.createElement('Foo', { "key": i }))`, Tsx: true},
		// Computed-key with NON-`key` Identifier inside (`[other]: i`) —
		// `prop.key.name === 'other' !== 'key'`. NOT reported.
		{Code: `foo.map((bar, i) => React.createElement('Foo', { [other]: i }))`, Tsx: true},
		// Shorthand `{ key }` in cloneElement props when the index
		// parameter is NOT named `key` — `checkPropValue` runs on the
		// Identifier `key`, but `key` is not on the indexParamNames
		// stack, so no report.
		{Code: `foo.map((bar, i) => { const key = bar.id; return React.cloneElement(c, { key }); })`, Tsx: true},
		// Iterator returning a non-JSX value — rule never inspects.
		{Code: `foo.map((bar, i) => bar + i)`, Tsx: true},
		// Empty arrow body — no JSX, no createElement, nothing to check.
		{Code: `foo.map((bar, i) => {})`, Tsx: true},
		// Nested iterator where INNER shadows the outer index name —
		// the stack contains `i` twice; pop order is correct so on exit
		// we still have one `i` left for the outer scope. Locks the
		// stack-pop balance.
		{Code: `foo.map((a, i) => bar.map((c, i) => <X key={c.id} />))`, Tsx: true},
		// `BigInt` / numeric / boolean literal keys — none match the
		// Identifier / Template / Binary / call-form gates.
		{Code: `foo.map((bar, i) => <Foo key={42} />)`, Tsx: true},
		{Code: `foo.map((bar, i) => <Foo key={true} />)`, Tsx: true},
		// Reduce accumulator `a` (param 0) is NOT the index even though
		// the rule "watches" `reduce` — index is at position 2.
		{Code: `foo.reduce((a, b) => a.concat(<Foo key={a} />), [])`, Tsx: true},
		// Default-valued index parameter `(bar, i = 0)` — upstream's
		// AssignmentPattern has no plain `.name`, so the index is
		// NOT pushed. Locked in for input/output parity.
		{Code: `foo.map((bar, i = 0) => <Foo key={i} />)`, Tsx: true},
		// Rule receives extra options — schema is `[]` upstream, but
		// pass-through must not break. `GetOptionsMap` is not used by
		// this rule; we still verify that arbitrary input doesn't
		// disturb behavior.
		{
			Code:    `foo.map((bar, i) => <Foo key={bar.id} />)`,
			Tsx:     true,
			Options: map[string]interface{}{"unknown": true},
		},
		// `settings.react.pragma = "Act"` — `Act.createElement` becomes
		// the recognized factory; `React.createElement` is no longer
		// matched, so referencing the index in `React.createElement`
		// props would NOT trigger the createElement branch. Verified
		// here as a green-path control alongside the invalid mirror
		// below.
		{
			Code: `foo.map((bar, i) => Act.cloneElement(c, { key: bar.id }))`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"pragma": "Act"},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Direct index reference in JSX key attribute ----
		{
			Code: `foo.map((bar, i) => <Foo key={i} />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 31},
			},
		},
		{
			Code: `[{}, {}].map((bar, i) => <Foo key={i} />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 36},
			},
		},
		{
			Code: `foo.map((bar, anything) => <Foo key={anything} />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 38},
			},
		},

		// ---- Template literal containing the index ----
		{
			Code: `foo.map((bar, i) => <Foo key={` + "`" + `foo-${i}` + "`" + `} />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 31},
			},
		},

		// ---- Concatenation chain ----
		{
			Code: `foo.map((bar, i) => <Foo key={'foo-' + i} />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 31},
			},
		},
		{
			Code: `foo.map((bar, i) => <Foo key={'foo-' + i + '-bar'} />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 31},
			},
		},

		// ---- React.cloneElement (member-access callee) ----
		{
			Code: `foo.map((baz, i) => React.cloneElement(someChild, { ...someChild.props, key: i }))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 78},
			},
		},

		// ---- Bare `cloneElement` resolved from `import { cloneElement } from 'react'` ----
		{
			Code: `
import { cloneElement } from 'react';

foo.map((baz, i) => cloneElement(someChild, { ...someChild.props, key: i }))
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 4, Column: 72},
			},
		},

		// ---- React.cloneElement inside a block-bodied callback ----
		{
			Code: `
foo.map((item, i) => {
  return React.cloneElement(someChild, {
    key: i
  })
})
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 4, Column: 10},
			},
		},
		{
			Code: `
import { cloneElement } from 'react';

foo.map((item, i) => {
  return cloneElement(someChild, {
    key: i
  })
})
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 6, Column: 10},
			},
		},

		// ---- Every iterator method in indexParamPositions, JSX path ----
		{
			Code: `foo.forEach((bar, i) => { baz.push(<Foo key={i} />); })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 46},
			},
		},
		{
			Code: `foo.filter((bar, i) => { baz.push(<Foo key={i} />); })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 45},
			},
		},
		{
			Code: `foo.some((bar, i) => { baz.push(<Foo key={i} />); })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 43},
			},
		},
		{
			Code: `foo.every((bar, i) => { baz.push(<Foo key={i} />); })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 44},
			},
		},
		{
			Code: `foo.find((bar, i) => { baz.push(<Foo key={i} />); })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 43},
			},
		},
		{
			Code: `foo.findIndex((bar, i) => { baz.push(<Foo key={i} />); })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 48},
			},
		},
		{
			Code: `foo.reduce((a, b, i) => a.concat(<Foo key={i} />), [])`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 44},
			},
		},
		{
			Code: `foo.flatMap((a, i) => <Foo key={i} />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 33},
			},
		},
		{
			Code: `foo.reduceRight((a, b, i) => a.concat(<Foo key={i} />), [])`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 49},
			},
		},

		// ---- React.createElement (member-access callee) with various index shapes ----
		{
			Code: `foo.map((bar, i) => React.createElement('Foo', { key: i }))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 55},
			},
		},
		{
			Code: `foo.map((bar, i) => React.createElement('Foo', { key: ` + "`" + `foo-${i}` + "`" + ` }))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 55},
			},
		},
		{
			Code: `foo.map((bar, i) => React.createElement('Foo', { key: 'foo-' + i }))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 55},
			},
		},
		{
			Code: `foo.map((bar, i) => React.createElement('Foo', { key: 'foo-' + i + '-bar' }))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 55},
			},
		},

		// ---- Every iterator method, createElement props path ----
		{
			Code: `foo.forEach((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 70},
			},
		},
		{
			Code: `foo.filter((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 69},
			},
		},
		{
			Code: `foo.some((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 67},
			},
		},
		{
			Code: `foo.every((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 68},
			},
		},
		{
			Code: `foo.find((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 67},
			},
		},
		{
			Code: `foo.findIndex((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 72},
			},
		},

		// ---- React.Children.* (callback shifted to args[1]) ----
		{
			Code: `
Children.map(this.props.children, (child, index) => {
  return React.cloneElement(child, { key: index });
})
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 3, Column: 43},
			},
		},
		{
			Code: `
import { cloneElement } from 'react';

Children.map(this.props.children, (child, index) => {
  return cloneElement(child, { key: index });
})
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 5, Column: 37},
			},
		},
		{
			Code: `
React.Children.map(this.props.children, (child, index) => {
  return React.cloneElement(child, { key: index });
})
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 3, Column: 43},
			},
		},
		{
			Code: `
import { cloneElement } from 'react';

React.Children.map(this.props.children, (child, index) => {
  return cloneElement(child, { key: index });
})
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 5, Column: 37},
			},
		},
		{
			Code: `
Children.forEach(this.props.children, (child, index) => {
  return React.cloneElement(child, { key: index });
})
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 3, Column: 43},
			},
		},
		{
			Code: `
import { cloneElement } from 'react';

Children.forEach(this.props.children, (child, index) => {
  return cloneElement(child, { key: index });
})
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 5, Column: 37},
			},
		},
		{
			Code: `
React.Children.forEach(this.props.children, (child, index) => {
  return React.cloneElement(child, { key: index });
})
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 3, Column: 43},
			},
		},
		{
			Code: `
import { cloneElement } from 'react';

React.Children.forEach(this.props.children, (child, index) => {
  return cloneElement(child, { key: index });
})
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 5, Column: 37},
			},
		},

		// ---- Optional-chain iterator ----
		{
			Code: `foo?.map((child, i) => <Foo key={i} />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 34},
			},
		},

		// ---- index.toString() coercion ----
		{
			Code: `
foo.map((bar, index) => (
  <Element key={index.toString()} bar={bar} />
))
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 3, Column: 17},
			},
		},

		// ---- String(index) coercion — report on the index argument ----
		{
			Code: `
foo.map((bar, index) => (
  <Element key={String(index)} bar={bar} />
))
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 3, Column: 24},
			},
		},

		// ---- Trivial multi-line index reference (parens around the JSX) ----
		{
			Code: `
foo.map((bar, index) => (
  <Element key={index} bar={bar} />
))
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 3, Column: 17},
			},
		},

		// ---- Nested iterators: outer index used inside inner callback ----
		// Stack contains both `i` (outer) and `j` (inner); outer name
		// referenced from the inner JSX is reported.
		{
			Code: `foo.map((a, i) => bar.map((c, j) => <X key={i} />))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 45},
			},
		},
		// Inner index used inside inner JSX — both names live on the
		// stack; the same-level match still reports exactly once.
		{
			Code: `foo.map((a, i) => bar.map((c, j) => <X key={j} />))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 45},
			},
		},
		// Same index name reused across nested levels — each push goes
		// onto the stack independently, but the JSX references collapse
		// to the same name; both inner and outer JSX reference reports
		// correctly. Locks in upstream's name-based matching.
		{
			Code: `foo.map((a, i) => bar.map((c, i) => <X key={i} />))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 45},
			},
		},
		// Untracked iterator nested inside a tracked one — the outer
		// index `i` remains on the stack while the inner `bar.qux` does
		// NOT push, so `i` is still reportable from inside the inner
		// callback's JSX.
		{
			Code: `foo.map((a, i) => bar.qux((c, j) => <X key={i} />))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 45},
			},
		},

		// ---- Nested function inside an iterator body keeps tracking ----
		// Upstream pushes/pops only on CallExpression boundaries, NOT on
		// nested function-definition boundaries — so an inner regular
		// function still sees `i` as a tracked index. Locks in parity.
		{
			Code: `
foo.map((bar, i) => function inner() {
  return <Foo key={i} />;
})
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 3, Column: 20},
			},
		},

		// ---- Identifier branch: ANY react-imported name is accepted ----
		// Upstream's `isCreateCloneElement` Identifier branch does NOT
		// verify that the imported name is `createElement` /
		// `cloneElement` — it returns true for any `import { ... } from
		// 'react'` ImportSpecifier. We mirror this quirk byte-for-byte.
		// Here `import { foo } from 'react'` makes the bare callee
		// `foo(...)` look like a React factory call; `key: i` reports.
		{
			Code: `
import { foo } from 'react';

foo.map((bar, i) => foo(c, { key: i }))
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 4, Column: 35},
			},
		},

		// ---- Optional-chain pragma `React?.cloneElement` ----
		// Upstream listens on `'CallExpression, OptionalCallExpression'`
		// and gates on
		// `node.type === 'MemberExpression' || 'OptionalMemberExpression'`,
		// so optional-chain pragma calls ARE recognized. Locked here.
		{
			Code: `foo.map((bar, i) => React?.cloneElement(c, { key: i }))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 51},
			},
		},
		{
			Code: `foo.map((bar, i) => React?.createElement('Foo', { key: i }))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 56},
			},
		},

		// ---- Multi-substitution template — one report per index span ----
		// Upstream's loop reports once per substitution that resolves to
		// the index, even when the whole TemplateLiteral is the report
		// target. `\`${i}-${i}\`` therefore yields TWO reports.
		{
			Code: `foo.map((bar, i) => <Foo key={` + "`" + `${i}-${i}` + "`" + `} />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 31},
				{MessageId: "noArrayIndex", Line: 1, Column: 31},
			},
		},
		// Multiple index-typed leaves in a Binary chain — one report each.
		{
			Code: `foo.map((bar, i) => <Foo key={i + '-' + i} />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 31},
				{MessageId: "noArrayIndex", Line: 1, Column: 31},
			},
		},

		// ---- Multiple `key` keys in the same props object ----
		// Upstream iterates every property whose Identifier name is
		// `key` and runs `checkPropValue` on each. Two index-typed
		// `key`s yield TWO reports.
		{
			Code: `foo.map((bar, i) => React.createElement('Foo', { key: i, key: i }))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 55},
				{MessageId: "noArrayIndex", Line: 1, Column: 63},
			},
		},

		// ---- Nested cloneElement: each call inspects its own key ----
		// Outer CallExpression's listener fires first (column 82), then
		// the inner one (column 69) — both `key: i` pairs report
		// independently.
		{
			Code: `foo.map((bar, i) => React.cloneElement(React.cloneElement(c, { key: i }), { key: i }))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 82},
				{MessageId: "noArrayIndex", Line: 1, Column: 69},
			},
		},

		// ---- Real-user scenarios: settings.react.pragma swap ----
		// With `pragma: "Act"`, `Act.cloneElement(...)` IS the React
		// factory; `key: i` inside it is reported, while the same code
		// using `React.cloneElement` would NOT match (the inverse case
		// is the green-path test in the valid block).
		{
			Code: `foo.map((bar, i) => Act.cloneElement(c, { key: i }))`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"pragma": "Act"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 48},
			},
		},

		// ---- Chained property access on the iterator receiver ----
		// `foo.bar.map(...)` — the callee's property is `map`, the
		// receiver's full shape doesn't matter to the iterator gate.
		{
			Code: `foo.bar.map((c, i) => <X key={i} />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 31},
			},
		},

		// ---- Iterator callback inside an IIFE wrapper ----
		// `foo.map((function(){ return (a,i)=>... })())` — the actual
		// callback at args[0] is a CallExpression, NOT a function,
		// so `getMapIndexParamName` returns "" and the index is not
		// pushed. Renders as a green-path scenario (no report) — locked
		// here as an INVERSE of the canonical "callback must be a
		// function literal" gate. Goes in valid since no diagnostic.
		// (See valid suite — kept here as a comment marker only.)

		// ---- `String(index)` reports on the argument node ----
		// Lock in the upstream report-target distinction: `index.toString()`
		// reports on the CALL, but `String(index)` reports on the
		// ARGUMENT. We already have the simple case; this reinforces
		// it across the index-passed-via-parens form.
		{
			Code: `foo.map((bar, i) => <Foo key={String((i))} />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				// Reports on the unwrapped Identifier `i` — matches
				// upstream's report range (ESTree flattens parens, so
				// `node.arguments[0]` is the bare Identifier).
				{MessageId: "noArrayIndex", Line: 1, Column: 39},
			},
		},

		// ---- JSX element key inside a JSX-children expression ----
		// `<div>{foo.map((bar, i) => <X key={i} />)}</div>` — the
		// iterator lives inside a JsxExpression child; that's a
		// regular CallExpression to our listener.
		{
			Code: `<div>{foo.map((bar, i) => <X key={i} />)}</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 35},
			},
		},

		// ---- Shorthand `{ key }` when the index parameter is named `key` ----
		// Upstream's `prop.key.name === 'key'` matches the shorthand;
		// `checkPropValue(prop.value)` then sees the Identifier `key`,
		// which is the index parameter on the stack. Locks the
		// previously-missed shorthand branch in `checkObjectKeyProp`.
		{
			Code: `foo.map((bar, key) => React.cloneElement(c, { key }))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 47},
			},
		},

		// ---- TS-wrapped iterator receiver still triggers the iterator gate ----
		// Upstream's `getMapIndexParamName` reads `callee.property.name`
		// only; the receiver shape is opaque. So `(foo as any[]).map(...)`
		// IS recognized as a tracked iterator.
		{
			Code: `(foo as any[]).map((bar, i) => <Foo key={i} />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 42},
			},
		},

		// ---- Mixed key kinds: only Identifier-`key` and `[key]` match ----
		// `["key"]: i` (StringLiteral inside computed) is opaque
		// upstream — `prop.key.name` is undefined on a Literal.
		// Plain `key: i` matches via Identifier. Locked here.
		{
			Code: `foo.map((bar, i) => React.createElement('Foo', { key: i, ["key"]: i }))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 55},
			},
		},

		// ---- Computed property `[<Identifier "key">]: ...` ----
		// Upstream ESTree puts the inner expression directly as
		// `prop.key` when `computed: true`. With `[key]: key` and the
		// iterator's index parameter named `key`, `prop.key.name ===
		// 'key'` matches; `checkPropValue(prop.value)` then reports on
		// the value Identifier `key` (which is the index). Locks the
		// upstream computed-property quirk byte-for-byte.
		{
			Code: `foo.map((bar, key) => React.cloneElement(c, { [key]: key }))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 54},
			},
		},

		// ---- React.Children.toArray result piped into .map(...) ----
		// `React.Children.toArray(children).map((c, i) => <X key={i} />)`
		// — `.map` on the toArray result is a normal map; `i` is the
		// index, reported.
		{
			Code: `React.Children.toArray(children).map((c, i) => <X key={i} />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 1, Column: 56},
			},
		},

		// ---- Block-bodied callback with multiple JSX returns ----
		// One report per JSX `key` reference, regardless of position.
		{
			Code: `
foo.map((bar, i) => {
  if (cond) return <A key={i} />;
  return <B key={i} />;
})
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noArrayIndex", Line: 3, Column: 28},
				{MessageId: "noArrayIndex", Line: 4, Column: 18},
			},
		},
	})
}
