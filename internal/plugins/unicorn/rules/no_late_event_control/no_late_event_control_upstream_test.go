package no_late_event_control_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_late_event_control"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Cases and diagnostic ranges from eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-late-event-control.js
func TestNoLateEventControlUpstream(t *testing.T) {
	t.Run("JavaScript and project-aware TypeScript", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_late_event_control.NoLateEventControlRule,
			[]rule_tester.ValidTestCase{
				{FileName: "file.js", Code: `element.addEventListener("click", event => { event.preventDefault(); });`},
				{FileName: "file.js", Code: `element.addEventListener("click", event => { event.stopPropagation(); });`},
				{FileName: "file.js", Code: `element.addEventListener("click", event => { event.stopImmediatePropagation(); });`},
				{FileName: "file.js", Code: `element.addEventListener("click", async event => {
	event.preventDefault();
	await somePromise;
});`},
				{FileName: "file.js", Code: `element.onclick = async event => { await foo(event.preventDefault()); };`},
				{FileName: "file.js", Code: `element.onclick = async event => { log(event.stopPropagation(), await ready()); };`},
				{FileName: "file.js", Code: `element.onclick = async event => { for await (const chunk of event.preventDefault()) { log(chunk); } };`},
				{FileName: "file.js", Code: `element.onclick = async event => {
	for (const child of event.preventDefault()) {
		await render(child);
	}
};`},
				{FileName: "file.js", Code: `async function handler() { const event = getEvent(); await somePromise; event.preventDefault(); }`},
				{FileName: "file.js", Code: `async event => { await somePromise; event.composedPath(); }`},
				{FileName: "file.js", Code: `async event => { await somePromise; event[method](); }`},
				{FileName: "file.js", Code: `async event => { const method = event.preventDefault; await somePromise; method(); }`},
				{FileName: "file.js", Code: `async event => { const {preventDefault} = event; await somePromise; preventDefault(); }`},
				{FileName: "file.js", Code: `async () => { await somePromise; event.preventDefault(); }`},
				{FileName: "file.js", Code: `async function outer(event) {
	await somePromise;
	element.addEventListener("click", event => {
		event.preventDefault();
	});
}`},
				{FileName: "file.js", Code: `element.onclick = event => {
	promise.then(async () => {
		await somePromise;
	});
	event.preventDefault();
};`},
				{FileName: "file.ts", Code: `type Control = {
	preventDefault(): void;
};

async function updateControl(event: Control) {
	await event.load();
	event.preventDefault();
}`},
				{FileName: "file.js", Code: `async function onClick(event) { await load(); event.detail.preventDefault(); }`},
			},
			[]rule_tester.InvalidTestCase{
				invalidCall("file.js", `element.addEventListener("click", async event => {
	await somePromise;
	event.preventDefault();
});`,
					callError("after-suspension", "event.preventDefault()", 3, 2, 3, 24)),
				invalidCall("file.js", `element.addEventListener("click", async event => {
	await somePromise;
	event.stopPropagation();
});`,
					callError("after-suspension", "event.stopPropagation()", 3, 2, 3, 25)),
				invalidCall("file.js", `element.addEventListener("click", async event => {
	await somePromise;
	event.stopImmediatePropagation();
});`,
					callError("after-suspension", "event.stopImmediatePropagation()", 3, 2, 3, 34)),
				invalidCall("file.js", `element.onclick = function * (event) {
	yield somePromise;
	event.preventDefault();
};`,
					callError("in-generator", "event.preventDefault()", 3, 2, 3, 24)),
				invalidCall("file.js", `element.onclick = function * (event) { event.stopPropagation(yield ready); };`,
					callError("in-generator", "event.stopPropagation()", 1, 40, 1, 74)),
				invalidCall("file.js", `element.onclick = async e => { await somePromise; e.preventDefault(); };`,
					callError("after-suspension", "e.preventDefault()", 1, 51, 1, 69)),
				invalidCall("file.js", `element.onclick = async mouseDownEvent => { await somePromise; mouseDownEvent.stopPropagation(); };`,
					callError("after-suspension", "mouseDownEvent.stopPropagation()", 1, 64, 1, 96)),
				invalidCall("file.js", `element.onclick = async event => { log(await ready(), event.preventDefault()); };`,
					callError("after-suspension", "event.preventDefault()", 1, 55, 1, 77)),
				invalidCall("file.js", `element.onclick = async event => { event.preventDefault(await ready()); };`,
					callError("after-suspension", "event.preventDefault()", 1, 36, 1, 71)),
				invalidCall("file.js", `element.onclick = async event => {
	event.preventDefault();
	await somePromise;
	event.stopPropagation();
};`,
					callError("after-suspension", "event.stopPropagation()", 4, 2, 4, 25)),
				invalidCall("file.js", `element.onclick = function * (event) { event.preventDefault(); yield somePromise; };`,
					callError("in-generator", "event.preventDefault()", 1, 40, 1, 62)),
				invalidCall("file.js", `element.onclick = event => setTimeout(() => event.preventDefault());`,
					callError("in-nested-function", "event.preventDefault()", 1, 45, 1, 67)),
				invalidCall("file.js", `element.onclick = event => {
	setTimeout(() => {
		event.stopPropagation();
	}, 1);
};`,
					callError("in-nested-function", "event.stopPropagation()", 3, 3, 3, 26)),
				invalidCall("file.js", `element.onclick = event => { fetch(url).then(() => event.stopImmediatePropagation()); };`,
					callError("in-nested-function", "event.stopImmediatePropagation()", 1, 52, 1, 84)),
				invalidCall("file.js", `async event => { await somePromise; event?.preventDefault(); }`,
					callError("after-suspension", "event.preventDefault()", 1, 37, 1, 60)),
				invalidCall("file.js", `async event => { await somePromise; event.preventDefault?.(); }`,
					callError("after-suspension", "event.preventDefault()", 1, 37, 1, 61)),
				invalidCall("file.js", `async function handleClick(event) {
	await somePromise;
	event.preventDefault();
}`,
					callError("after-suspension", "event.preventDefault()", 3, 2, 3, 24)),
				invalidCall("file.js", `element.onclick = async event => {
	for await (const chunk of stream) {
		log(chunk);
	}
	event.preventDefault();
};`,
					callError("after-suspension", "event.preventDefault()", 5, 2, 5, 24)),
				invalidCall("file.js", `element.onclick = async event => {
	for await (const chunk of stream) {
		event.preventDefault();
	}
};`,
					callError("after-suspension", "event.preventDefault()", 3, 3, 3, 25)),
				invalidCall("file.js", `element.onclick = async event => {
	for (const item of await load()) {
		event.preventDefault();
	}
};`,
					callError("after-suspension", "event.preventDefault()", 3, 3, 3, 25)),
				invalidCall("file.js", `element.onclick = async event => {
	while (condition) {
		event.preventDefault();
		await promise;
	}
};`,
					callError("after-suspension", "event.preventDefault()", 3, 3, 3, 25)),
				invalidCall("file.js", `element.onclick = async event => {
	while (event.preventDefault()) {
		await promise;
	}
};`,
					callError("after-suspension", "event.preventDefault()", 2, 9, 2, 31)),
				invalidCall("file.js", `element.onclick = async event => {
	try {
		await somePromise;
	} catch {}
	event.preventDefault();
};`,
					callError("after-suspension", "event.preventDefault()", 5, 2, 5, 24)),
				invalidCall("file.js", `element.onclick = event => {
	setTimeout(() => {
		requestAnimationFrame(() => {
			event.preventDefault();
		});
	});
};`,
					callError("in-nested-function", "event.preventDefault()", 4, 4, 4, 26)),
				invalidCall("file.js", `element.onclick = event => {
	button.addEventListener("click", async () => {
		await somePromise;
		event.preventDefault();
	});
};`,
					callError("in-nested-function", "event.preventDefault()", 4, 3, 4, 25)),
			},
		)
	})
	t.Run("TypeScript annotations", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_late_event_control.NoLateEventControlRule,
			[]rule_tester.ValidTestCase{
				{FileName: "file.ts", Code: `async function updateControl(event: {preventDefault(): void}) { await load(); event.preventDefault(); }`},
				{FileName: "file.ts", Code: `async function updateControl(event: string) { await load(); event.preventDefault(); }`},
				{FileName: "file.ts", Code: `type Event = string; async function updateControl(event: Event) { await load(); event.preventDefault(); }`},
				{FileName: "file.ts", Code: `type SyntheticEvent = string; async function updateControl(value: SyntheticEvent) { await load(); value.preventDefault(); }`},
				{FileName: "file.ts", Code: `namespace React { export type SyntheticEvent = string } async function updateControl(value: React.SyntheticEvent) { await load(); value.preventDefault(); }`},
				{FileName: "file.ts", Code: `import type {Event} from "not-dom"; async function updateControl(event: Event) { await load(); event.preventDefault(); }`},
				{FileName: "file.ts", Code: `import type {NotAnEvent as Event} from "not-dom"; async function updateControl(event: Event) { await load(); event.preventDefault(); }`},
				{FileName: "file.ts", Code: `import Event from "not-dom"; async function updateControl(event: Event) { await load(); event.preventDefault(); }`},
				{FileName: "file.ts", Code: `import type * as DOM from "not-dom"; async function updateControl(event: DOM.Event) { await load(); event.preventDefault(); }`},
				{FileName: "file.ts", Code: `import type {SyntheticEvent} from "not-react"; async function updateControl(value: SyntheticEvent) { await load(); value.preventDefault(); }`},
				{FileName: "file.ts", Code: `import React from "not-react"; async function updateControl(event: React.SyntheticEvent) { await load(); event.preventDefault(); }`},
				{FileName: "file.ts", Code: `import SyntheticEvent from "react"; async function updateControl(event: SyntheticEvent) { await load(); event.preventDefault(); }`},
				{FileName: "file.ts", Code: `import type * as React from "not-react"; async function updateControl(value: React.SyntheticEvent) { await load(); value.preventDefault(); }`},
				{FileName: "file.ts", Code: `import type {NotEvent as SyntheticEvent} from "react"; async function updateControl(event: SyntheticEvent) { await load(); event.preventDefault(); }`},
				{FileName: "file.ts", Code: `import type * as React from "react"; async function updateControl(event: React.NotEvent) { await load(); event.preventDefault(); }`},
			},
			[]rule_tester.InvalidTestCase{
				invalidCall("file.ts", `async function handleClick(event: Event) { await load(); event.preventDefault(); }`,
					callError("after-suspension", "event.preventDefault()", 1, 58, 1, 80)),
				invalidCall("file.ts", `async function handleClick(event: MouseEvent) { await load(); event.stopPropagation(); }`,
					callError("after-suspension", "event.stopPropagation()", 1, 63, 1, 86)),
				invalidCall("file.ts", `import React from "react"; async function handleClick(value: React.SyntheticEvent) { await load(); value.preventDefault(); }`,
					callError("after-suspension", "value.preventDefault()", 1, 100, 1, 122)),
				invalidCall("file.ts", `import type React from "react"; async function handleClick(value: React.SyntheticEvent) { await load(); value.preventDefault(); }`,
					callError("after-suspension", "value.preventDefault()", 1, 105, 1, 127)),
				invalidCall("file.ts", `import type * as React from "react"; async function handleClick(event: React.SyntheticEvent) { await load(); event.stopImmediatePropagation(); }`,
					callError("after-suspension", "event.stopImmediatePropagation()", 1, 110, 1, 142)),
				invalidCall("file.ts", `import type * as React from "react"; async function handleClick(value: React.AnimationEvent) { await load(); value.preventDefault(); }`,
					callError("after-suspension", "value.preventDefault()", 1, 110, 1, 132)),
				invalidCall("file.ts", `import type * as React from "react"; async function handleClick(value: React.InvalidEvent) { await load(); value.preventDefault(); }`,
					callError("after-suspension", "value.preventDefault()", 1, 108, 1, 130)),
				invalidCall("file.ts", `import type * as React from "react"; async function handleClick(value: React.SubmitEvent) { await load(); value.preventDefault(); }`,
					callError("after-suspension", "value.preventDefault()", 1, 107, 1, 129)),
				invalidCall("file.ts", `import type * as React from "react"; async function handleClick(value: React.TransitionEvent) { await load(); value.preventDefault(); }`,
					callError("after-suspension", "value.preventDefault()", 1, 111, 1, 133)),
				invalidCall("file.ts", `async function handleClick(mouseDownEvent: unknown) { await load(); mouseDownEvent.preventDefault(); }`,
					callError("after-suspension", "mouseDownEvent.preventDefault()", 1, 69, 1, 100)),
				invalidCall("file.ts", `async function handleClick(value: Event | MouseEvent) { await load(); value.preventDefault(); }`,
					callError("after-suspension", "value.preventDefault()", 1, 71, 1, 93)),
				invalidCall("file.ts", `async function handleClick(change: PictureInPictureEvent) { await load(); change.stopPropagation(); }`,
					callError("after-suspension", "change.stopPropagation()", 1, 75, 1, 99)),
				invalidCall("file.ts", `async function handleClick(request: FetchEvent) { await load(); request.preventDefault(); }`,
					callError("after-suspension", "request.preventDefault()", 1, 65, 1, 89)),
				invalidCall("file.ts", `async function handleClick(value: Event) { await load(); value!.preventDefault(); }`,
					callError("after-suspension", "value.preventDefault()", 1, 58, 1, 81)),
				invalidCall("file.ts", `async function handleClick(value: Event) { await load(); (value as Event).stopPropagation(); }`,
					callError("after-suspension", "value.stopPropagation()", 1, 58, 1, 92)),
				invalidCall("file.ts", `async function handleClick(value: Event) { await load(); (<Event>value).stopImmediatePropagation(); }`,
					callError("after-suspension", "value.stopImmediatePropagation()", 1, 58, 1, 99)),
				invalidCall("file.ts", `import type {SyntheticEvent} from "react"; async function handleClick(value: SyntheticEvent) { await load(); value.preventDefault(); }`,
					callError("after-suspension", "value.preventDefault()", 1, 110, 1, 132)),
				invalidCall("file.ts", `import type {BaseSyntheticEvent} from "react"; async function handleClick(value: BaseSyntheticEvent) { await load(); value.preventDefault(); }`,
					callError("after-suspension", "value.preventDefault()", 1, 118, 1, 140)),
				invalidCall("file.ts", `import type {SyntheticEvent as SE} from "react"; async function handleClick(value: SE) { await load(); value.preventDefault(); }`,
					callError("after-suspension", "value.preventDefault()", 1, 104, 1, 126)),
				invalidCall("file.ts", `import type {InputEvent as Input} from "react"; async function handleClick(value: Input) { await load(); value.preventDefault(); }`,
					callError("after-suspension", "value.preventDefault()", 1, 106, 1, 128)),
				invalidCall("file.ts", `import type {ChangeEvent} from "react"; async function handleClick(value: ChangeEvent<HTMLInputElement>) { await load(); value.preventDefault(); }`,
					callError("after-suspension", "value.preventDefault()", 1, 122, 1, 144)),
				invalidCall("file.ts", `import type {ToggleEvent} from "react"; async function handleClick(value: ToggleEvent) { await load(); value.preventDefault(); }`,
					callError("after-suspension", "value.preventDefault()", 1, 104, 1, 126)),
			},
		)
	})
	t.Run("Computed property names", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_late_event_control.NoLateEventControlRule,
			[]rule_tester.ValidTestCase{},
			[]rule_tester.InvalidTestCase{
				invalidCall("file.js", `async event => { await somePromise; event["preventDefault"](); }`,
					callError("after-suspension", "event.preventDefault()", 1, 37, 1, 62)),
				invalidCall("file.js", "async event => { await somePromise; event[`preventDefault`](); }",
					callError("after-suspension", "event.preventDefault()", 1, 37, 1, 62)),
				invalidCall("file.js", `async event => { await somePromise; event['stopPropagation'](); }`,
					callError("after-suspension", "event.stopPropagation()", 1, 37, 1, 63)),
				invalidCall("file.js", `async event => { await somePromise; event?.["preventDefault"](); }`,
					callError("after-suspension", "event.preventDefault()", 1, 37, 1, 64)),
			},
		)
	})
	t.Run("Documentation examples", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_late_event_control.NoLateEventControlRule,
			[]rule_tester.ValidTestCase{
				{FileName: "file.js", Code: `element.addEventListener('click', async event => {
	event.preventDefault();
	await save();
});`},
				{FileName: "file.js", Code: `element.addEventListener('click', event => {
	event.stopPropagation();
	setTimeout(() => {
		update();
	});
});`},
				{FileName: "file.js", Code: `function prevent(event) {
	event.preventDefault();
}

element.addEventListener('click', event => {
	setTimeout(prevent, 1, event);
});`},
			},
			[]rule_tester.InvalidTestCase{
				invalidCall("file.js", `element.addEventListener('click', async event => {
	await save();
	event.preventDefault();
});`,
					callError("after-suspension", "event.preventDefault()", 3, 2, 3, 24)),
				invalidCall("file.js", `element.addEventListener('click', event => {
	setTimeout(() => {
		event.stopPropagation();
	});
});`,
					callError("in-nested-function", "event.stopPropagation()", 3, 3, 3, 26)),
			},
		)
	})
}

func invalidCall(filename, code string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{FileName: filename, Code: code, Errors: errors}
}

func callError(id, call string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	var description string
	switch id {
	case "after-suspension":
		description = "has no effect after the handler suspends. Call it before `await` or `yield` while the event is still being dispatched."
	case "in-generator":
		description = "is not called during event dispatch inside a generator function. Use a normal function handler instead."
	case "in-nested-function":
		description = "may run after synchronous event dispatch has finished. Call it in the outer handler before deferring work."
	default:
		panic("unknown no-late-event-control message ID: " + id)
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: id,
		Message:   "`" + call + "` " + description,
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}
