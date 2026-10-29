package no_watch_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/effector/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/effector/rules/no_watch"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoWatchExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_watch.NoWatchRule,
		[]rule_tester.ValidTestCase{
			// A unit-named type outside the effector package is not a unit.
			{Code: `
interface Store<T> { watch(fn: (v: T) => void): void }
declare const $local: Store<number>;
$local.watch(console.log);
`},
			// Untyped receivers carry no unit information.
			{Code: `
declare const anything: any;
anything.watch(console.log);
`},
			// Every union member must be a unit.
			{Code: `
import { createStore } from 'effector';
declare const mixed: ReturnType<typeof createStore<number>> | { watch(fn: () => void): void };
mixed.watch(() => {});
`},
			// Reading the method without calling it.
			{Code: `
import { createStore } from 'effector';
const $store = createStore(0);
const watch = $store.watch;
`},
			// A possibly undefined receiver is a union with undefined, and every union
			// member must be a unit, so upstream does not report it either.
			{Code: `
import { createStore } from 'effector';
declare const $maybe: ReturnType<typeof createStore<number>> | undefined;
$maybe?.watch(console.log);
`},
			// A string key is a Literal property in ESTree, so the selector does not match.
			{Code: `
import { createStore } from 'effector';
const $store = createStore(0);
$store['watch'](console.log);
`},
			// Parentheses ending an optional chain make the ESTree callee a ChainExpression.
			{Code: `
import { createStore } from 'effector';
declare const $maybe: ReturnType<typeof createStore<number>> | undefined;
($maybe?.watch)!(console.log);
`},
			// Only calls match; construction does not.
			{Code: `
import { createStore } from 'effector';
const $store = createStore(0);
new ($store.watch as any)(console.log);
`},
			// Other unit methods are allowed.
			{Code: `
import { createStore } from 'effector';
const $store = createStore(0);
$store.updates.map((v) => v);
`},
		},
		[]rule_tester.InvalidTestCase{
			// Full diagnostic range covers the whole call, including a multiline argument list.
			{
				Code: `
import { createStore } from 'effector';
const $store = createStore(0);
$store.watch((value) => {
  console.log(value);
});
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restricted", Message: restrictedMessage,
					Line: 4, Column: 1, EndLine: 6, EndColumn: 3,
				}},
			},
			// Optional call: ESTree still has a MemberExpression callee inside the ChainExpression.
			{
				Code: `
import { createStore } from 'effector';
const $store = createStore(0);
$store?.watch(console.log);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restricted", Message: restrictedMessage,
					Line: 4, Column: 1, EndLine: 4, EndColumn: 27,
				}},
			},
			// Parenthesized callee and receiver: ESTree drops the parentheses.
			{
				Code: `
import { createEvent } from 'effector';
const event = createEvent();
(event.watch)(console.log);
((event)).watch(console.log);
`,
				Errors: []rule_tester.InvalidTestCaseError{restricted(4, 1), restricted(5, 1)},
			},
			// A computed identifier key has property.name === 'watch' in ESTree.
			{
				Code: `
import { createEvent } from 'effector';
const event = createEvent();
const watch = 'watch';
event[watch](console.log);
`,
				Errors: []rule_tester.InvalidTestCaseError{restricted(5, 1)},
			},
			// Non-null assertion keeps the unit type.
			{
				Code: `
import { createStore } from 'effector';
declare const $maybe: ReturnType<typeof createStore<number>> | undefined;
$maybe!.watch(console.log);
`,
				Errors: []rule_tester.InvalidTestCaseError{restricted(4, 1)},
			},
			// Union of units, domain, and derived units reached through members.
			{
				Code: `
import { createDomain, createEffect, createEvent, createStore } from 'effector';
declare const unit: ReturnType<typeof createStore<number>> | ReturnType<typeof createEvent<number>>;
unit.watch(console.log);
createDomain().watch(console.log);
createEffect().pending.watch(console.log);
createEvent<number>().map((v) => v).watch(console.log);
`,
				Errors: []rule_tester.InvalidTestCaseError{restricted(4, 1), restricted(5, 1), restricted(6, 1), restricted(7, 1)},
			},
			// Nested watch calls are reported independently.
			{
				Code: `
import { createEvent, createStore } from 'effector';
const event = createEvent();
const $store = createStore(0);
event.watch(() => $store.watch(console.log));
`,
				Errors: []rule_tester.InvalidTestCaseError{restricted(5, 1), restricted(5, 19)},
			},
		},
	)
}
