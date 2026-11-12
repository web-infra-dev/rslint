package no_late_event_control_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_late_event_control"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoLateEventControlExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_late_event_control.NoLateEventControlRule,
		[]rule_tester.ValidTestCase{
			// Parenthesized optional callees and dynamic keys.
			{FileName: "file.js", Code: `async function submit(event) {
	await save();
	(event?.preventDefault)();
	const method = "preventDefault";
	event[method]();
}`},
			// Private members and local shadowing.
			{FileName: "file.js", Code: `class Control {
	#preventDefault() {}
	async handle(event) {
		await save();
		event.#preventDefault();
		{
			const event = getControl();
			event.preventDefault();
		}
	}
}`},
			// Suspension in nested methods stays within the method.
			{FileName: "file.js", Code: `async function handle(event) {
	while (ready()) {
		event.preventDefault();
		const listeners = { async onClick() { await save(); } };
	}
	event.stopPropagation();
}`},
			// One-time loop initializers and iterables.
			{FileName: "file.js", Code: `async function handle(event) {
	for (event.preventDefault(); ready(); tick()) await save();
}
async function handleKeys(event) {
	for (const key in event.preventDefault()) await save(key);
}`},
			// A default parameter and a computed method key can run synchronously.
			{FileName: "file.js", Code: `async function submit(event, result = event.preventDefault()) {
	const methods = { [event.stopPropagation()]() {} };
	await save(result);
}`},
			// Top-level suspension stays outside handlers.
			{FileName: "file.js", Code: `await prepare();
const submit = event => event.preventDefault();`},
			// JSDoc non-event type overrides the parameter name.
			{FileName: "file.js", Code: `/** @param {string} event */
async function submit(event) {
	await save();
	event.preventDefault();
}`},
		},
		[]rule_tester.InvalidTestCase{
			// Parentheses, folded bracket keys, optional calls, and UTF-16 ranges.
			invalidCall("file.js", `async function submit(event) {
	await save();
	log("😀"); (event.preventDefault)();
	event?.[
		"stop" + "Propagation"
	]?.();
}`,
				callError("after-suspension", "event.preventDefault()", 3, 13, 3, 37),
				callError("after-suspension", "event.stopPropagation()", 4, 2, 6, 7),
			),
			// An inner initializer repeats when the outer loop suspends.
			invalidCall("file.js", `async function submit(event) {
	while (ready()) {
		for (event.preventDefault(); check(); tick()) update();
		await save();
	}
}`,
				callError("after-suspension", "event.preventDefault()", 3, 8, 3, 30),
			),
			// Assignment and destructuring targets run on each iteration.
			invalidCall("file.js", `async function submit(event) {
	for (items[event.stopPropagation()] of stream) await save();
}
async function submitAsync(event) {
	for await (const {item = event.preventDefault()} of stream) use(item);
}`,
				callError("after-suspension", "event.stopPropagation()", 2, 13, 2, 36),
				callError("after-suspension", "event.preventDefault()", 5, 27, 5, 49),
			),
			// Computed method keys are evaluated in the enclosing function.
			invalidCall("file.js", `async function submit(event) {
	while (ready()) {
		event.stopPropagation();
		const methods = { [await save()]() {}, [event.preventDefault()]() {} };
	}
	class Control { static { event.stopImmediatePropagation(); } }
}`,
				callError("after-suspension", "event.stopPropagation()", 3, 3, 3, 26),
				callError("after-suspension", "event.preventDefault()", 4, 43, 4, 65),
				callError("after-suspension", "event.stopImmediatePropagation()", 6, 27, 6, 59),
			),
			// Destructured bindings and generator methods.
			invalidCall("file.js", `async function submit({ detail: { event } }) {
	await save();
	event.preventDefault();
}
const listeners = {
	*handle(event) {
		event.stopPropagation();
		return () => event.preventDefault();
	}
};`,
				callError("after-suspension", "event.preventDefault()", 3, 2, 3, 24),
				callError("in-generator", "event.stopPropagation()", 7, 3, 7, 26),
				callError("in-nested-function", "event.preventDefault()", 8, 16, 8, 38),
			),
			// An async JSX event handler.
			invalidCall("file.jsx", `const form = <button onClick={async e => {
	await save();
	e.preventDefault();
}} />;`,
				callError("after-suspension", "e.preventDefault()", 3, 2, 3, 20),
			),
			// React imports inside aliases and unions and assertions on receivers.
			invalidCall("file.ts", `import type {MouseEvent as ReactMouseEvent} from "react";
type HandlerEvent = ReactMouseEvent | MouseEvent;
async function submit(input: HandlerEvent) {
	await save();
	(input satisfies HandlerEvent)!.stopPropagation();
}`,
				callError("after-suspension", "input.stopPropagation()", 5, 2, 5, 51),
			),
			// JSDoc event type works with a different parameter name.
			invalidCall("file.js", `/** @param {MouseEvent} input */
async function submit(input) {
	await save();
	input.preventDefault();
}`,
				callError("after-suspension", "input.preventDefault()", 4, 2, 4, 24),
			),
			// Repeated calls use the declaring parameter's type, including after shadowing.
			invalidCall("file.ts", `async function submit(event: MouseEvent) {
	await save();
	event.preventDefault();
	event.stopPropagation();
	async function update(event: {preventDefault(): void}) {
		await save();
		event.preventDefault();
	}
	event.stopImmediatePropagation();
}`,
				callError("after-suspension", "event.preventDefault()", 3, 2, 3, 24),
				callError("after-suspension", "event.stopPropagation()", 4, 2, 4, 25),
				callError("after-suspension", "event.stopImmediatePropagation()", 9, 2, 9, 34),
			),
			// An explicit generic argument supplies the event type.
			invalidCall("file.ts", `type Identity<T> = T;
async function submit(input: Identity<MouseEvent>) {
	await save();
	input.preventDefault();
}`,
				callError("after-suspension", "input.preventDefault()", 4, 2, 4, 24),
			),
		},
	)
}
