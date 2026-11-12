import { RuleTester } from '../rule-tester';

// eslint-plugin-unicorn v79.0.0 tests and documentation examples.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-late-event-control.js
const messages = {
  'after-suspension':
    'has no effect after the handler suspends. Call it before `await` or `yield` while the event is still being dispatched.',
  'in-generator':
    'is not called during event dispatch inside a generator function. Use a normal function handler instead.',
  'in-nested-function':
    'may run after synchronous event dispatch has finished. Call it in the outer handler before deferring work.',
};

function invalid(
  filename: string,
  code: string,
  messageId: keyof typeof messages,
  call: string,
) {
  return {
    filename,
    code,
    errors: [{ messageId, message: '`' + call + '` ' + messages[messageId] }],
  };
}

new RuleTester().run('no-late-event-control', null as never, {
  valid: [
    // JavaScript and project-aware TypeScript
    {
      filename: 'file.js',
      code: 'element.addEventListener("click", event => { event.preventDefault(); });',
    },
    {
      filename: 'file.js',
      code: 'element.addEventListener("click", event => { event.stopPropagation(); });',
    },
    {
      filename: 'file.js',
      code: 'element.addEventListener("click", event => { event.stopImmediatePropagation(); });',
    },
    {
      filename: 'file.js',
      code: `element.addEventListener("click", async event => {
	event.preventDefault();
	await somePromise;
});`,
    },
    {
      filename: 'file.js',
      code: 'element.onclick = async event => { await foo(event.preventDefault()); };',
    },
    {
      filename: 'file.js',
      code: 'element.onclick = async event => { log(event.stopPropagation(), await ready()); };',
    },
    {
      filename: 'file.js',
      code: 'element.onclick = async event => { for await (const chunk of event.preventDefault()) { log(chunk); } };',
    },
    {
      filename: 'file.js',
      code: `element.onclick = async event => {
	for (const child of event.preventDefault()) {
		await render(child);
	}
};`,
    },
    {
      filename: 'file.js',
      code: 'async function handler() { const event = getEvent(); await somePromise; event.preventDefault(); }',
    },
    {
      filename: 'file.js',
      code: 'async event => { await somePromise; event.composedPath(); }',
    },
    {
      filename: 'file.js',
      code: 'async event => { await somePromise; event[method](); }',
    },
    {
      filename: 'file.js',
      code: 'async event => { const method = event.preventDefault; await somePromise; method(); }',
    },
    {
      filename: 'file.js',
      code: 'async event => { const {preventDefault} = event; await somePromise; preventDefault(); }',
    },
    {
      filename: 'file.js',
      code: 'async () => { await somePromise; event.preventDefault(); }',
    },
    {
      filename: 'file.js',
      code: `async function outer(event) {
	await somePromise;
	element.addEventListener("click", event => {
		event.preventDefault();
	});
}`,
    },
    {
      filename: 'file.js',
      code: `element.onclick = event => {
	promise.then(async () => {
		await somePromise;
	});
	event.preventDefault();
};`,
    },
    {
      filename: 'file.ts',
      code: `type Control = {
	preventDefault(): void;
};

async function updateControl(event: Control) {
	await event.load();
	event.preventDefault();
}`,
    },
    {
      filename: 'file.js',
      code: 'async function onClick(event) { await load(); event.detail.preventDefault(); }',
    },
    // TypeScript annotations
    {
      filename: 'file.ts',
      code: 'async function updateControl(event: {preventDefault(): void}) { await load(); event.preventDefault(); }',
    },
    {
      filename: 'file.ts',
      code: 'async function updateControl(event: string) { await load(); event.preventDefault(); }',
    },
    {
      filename: 'file.ts',
      code: 'type Event = string; async function updateControl(event: Event) { await load(); event.preventDefault(); }',
    },
    {
      filename: 'file.ts',
      code: 'type SyntheticEvent = string; async function updateControl(value: SyntheticEvent) { await load(); value.preventDefault(); }',
    },
    {
      filename: 'file.ts',
      code: 'namespace React { export type SyntheticEvent = string } async function updateControl(value: React.SyntheticEvent) { await load(); value.preventDefault(); }',
    },
    {
      filename: 'file.ts',
      code: 'import type {Event} from "not-dom"; async function updateControl(event: Event) { await load(); event.preventDefault(); }',
    },
    {
      filename: 'file.ts',
      code: 'import type {NotAnEvent as Event} from "not-dom"; async function updateControl(event: Event) { await load(); event.preventDefault(); }',
    },
    {
      filename: 'file.ts',
      code: 'import Event from "not-dom"; async function updateControl(event: Event) { await load(); event.preventDefault(); }',
    },
    {
      filename: 'file.ts',
      code: 'import type * as DOM from "not-dom"; async function updateControl(event: DOM.Event) { await load(); event.preventDefault(); }',
    },
    {
      filename: 'file.ts',
      code: 'import type {SyntheticEvent} from "not-react"; async function updateControl(value: SyntheticEvent) { await load(); value.preventDefault(); }',
    },
    {
      filename: 'file.ts',
      code: 'import React from "not-react"; async function updateControl(event: React.SyntheticEvent) { await load(); event.preventDefault(); }',
    },
    {
      filename: 'file.ts',
      code: 'import SyntheticEvent from "react"; async function updateControl(event: SyntheticEvent) { await load(); event.preventDefault(); }',
    },
    {
      filename: 'file.ts',
      code: 'import type * as React from "not-react"; async function updateControl(value: React.SyntheticEvent) { await load(); value.preventDefault(); }',
    },
    {
      filename: 'file.ts',
      code: 'import type {NotEvent as SyntheticEvent} from "react"; async function updateControl(event: SyntheticEvent) { await load(); event.preventDefault(); }',
    },
    {
      filename: 'file.ts',
      code: 'import type * as React from "react"; async function updateControl(event: React.NotEvent) { await load(); event.preventDefault(); }',
    },
    // Documentation examples
    {
      filename: 'file.js',
      code: `element.addEventListener('click', async event => {
	event.preventDefault();
	await save();
});`,
    },
    {
      filename: 'file.js',
      code: `element.addEventListener('click', event => {
	event.stopPropagation();
	setTimeout(() => {
		update();
	});
});`,
    },
    {
      filename: 'file.js',
      code: `function prevent(event) {
	event.preventDefault();
}

element.addEventListener('click', event => {
	setTimeout(prevent, 1, event);
});`,
    },
  ],
  invalid: [
    // JavaScript and project-aware TypeScript
    invalid(
      'file.js',
      `element.addEventListener("click", async event => {
	await somePromise;
	event.preventDefault();
});`,
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      `element.addEventListener("click", async event => {
	await somePromise;
	event.stopPropagation();
});`,
      'after-suspension',
      'event.stopPropagation()',
    ),
    invalid(
      'file.js',
      `element.addEventListener("click", async event => {
	await somePromise;
	event.stopImmediatePropagation();
});`,
      'after-suspension',
      'event.stopImmediatePropagation()',
    ),
    invalid(
      'file.js',
      `element.onclick = function * (event) {
	yield somePromise;
	event.preventDefault();
};`,
      'in-generator',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      'element.onclick = function * (event) { event.stopPropagation(yield ready); };',
      'in-generator',
      'event.stopPropagation()',
    ),
    invalid(
      'file.js',
      'element.onclick = async e => { await somePromise; e.preventDefault(); };',
      'after-suspension',
      'e.preventDefault()',
    ),
    invalid(
      'file.js',
      'element.onclick = async mouseDownEvent => { await somePromise; mouseDownEvent.stopPropagation(); };',
      'after-suspension',
      'mouseDownEvent.stopPropagation()',
    ),
    invalid(
      'file.js',
      'element.onclick = async event => { log(await ready(), event.preventDefault()); };',
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      'element.onclick = async event => { event.preventDefault(await ready()); };',
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      `element.onclick = async event => {
	event.preventDefault();
	await somePromise;
	event.stopPropagation();
};`,
      'after-suspension',
      'event.stopPropagation()',
    ),
    invalid(
      'file.js',
      'element.onclick = function * (event) { event.preventDefault(); yield somePromise; };',
      'in-generator',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      'element.onclick = event => setTimeout(() => event.preventDefault());',
      'in-nested-function',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      `element.onclick = event => {
	setTimeout(() => {
		event.stopPropagation();
	}, 1);
};`,
      'in-nested-function',
      'event.stopPropagation()',
    ),
    invalid(
      'file.js',
      'element.onclick = event => { fetch(url).then(() => event.stopImmediatePropagation()); };',
      'in-nested-function',
      'event.stopImmediatePropagation()',
    ),
    invalid(
      'file.js',
      'async event => { await somePromise; event?.preventDefault(); }',
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      'async event => { await somePromise; event.preventDefault?.(); }',
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      `async function handleClick(event) {
	await somePromise;
	event.preventDefault();
}`,
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      `element.onclick = async event => {
	for await (const chunk of stream) {
		log(chunk);
	}
	event.preventDefault();
};`,
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      `element.onclick = async event => {
	for await (const chunk of stream) {
		event.preventDefault();
	}
};`,
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      `element.onclick = async event => {
	for (const item of await load()) {
		event.preventDefault();
	}
};`,
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      `element.onclick = async event => {
	while (condition) {
		event.preventDefault();
		await promise;
	}
};`,
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      `element.onclick = async event => {
	while (event.preventDefault()) {
		await promise;
	}
};`,
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      `element.onclick = async event => {
	try {
		await somePromise;
	} catch {}
	event.preventDefault();
};`,
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      `element.onclick = event => {
	setTimeout(() => {
		requestAnimationFrame(() => {
			event.preventDefault();
		});
	});
};`,
      'in-nested-function',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      `element.onclick = event => {
	button.addEventListener("click", async () => {
		await somePromise;
		event.preventDefault();
	});
};`,
      'in-nested-function',
      'event.preventDefault()',
    ),
    // TypeScript annotations
    invalid(
      'file.ts',
      'async function handleClick(event: Event) { await load(); event.preventDefault(); }',
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.ts',
      'async function handleClick(event: MouseEvent) { await load(); event.stopPropagation(); }',
      'after-suspension',
      'event.stopPropagation()',
    ),
    invalid(
      'file.ts',
      'import React from "react"; async function handleClick(value: React.SyntheticEvent) { await load(); value.preventDefault(); }',
      'after-suspension',
      'value.preventDefault()',
    ),
    invalid(
      'file.ts',
      'import type React from "react"; async function handleClick(value: React.SyntheticEvent) { await load(); value.preventDefault(); }',
      'after-suspension',
      'value.preventDefault()',
    ),
    invalid(
      'file.ts',
      'import type * as React from "react"; async function handleClick(event: React.SyntheticEvent) { await load(); event.stopImmediatePropagation(); }',
      'after-suspension',
      'event.stopImmediatePropagation()',
    ),
    invalid(
      'file.ts',
      'import type * as React from "react"; async function handleClick(value: React.AnimationEvent) { await load(); value.preventDefault(); }',
      'after-suspension',
      'value.preventDefault()',
    ),
    invalid(
      'file.ts',
      'import type * as React from "react"; async function handleClick(value: React.InvalidEvent) { await load(); value.preventDefault(); }',
      'after-suspension',
      'value.preventDefault()',
    ),
    invalid(
      'file.ts',
      'import type * as React from "react"; async function handleClick(value: React.SubmitEvent) { await load(); value.preventDefault(); }',
      'after-suspension',
      'value.preventDefault()',
    ),
    invalid(
      'file.ts',
      'import type * as React from "react"; async function handleClick(value: React.TransitionEvent) { await load(); value.preventDefault(); }',
      'after-suspension',
      'value.preventDefault()',
    ),
    invalid(
      'file.ts',
      'async function handleClick(mouseDownEvent: unknown) { await load(); mouseDownEvent.preventDefault(); }',
      'after-suspension',
      'mouseDownEvent.preventDefault()',
    ),
    invalid(
      'file.ts',
      'async function handleClick(value: Event | MouseEvent) { await load(); value.preventDefault(); }',
      'after-suspension',
      'value.preventDefault()',
    ),
    invalid(
      'file.ts',
      'async function handleClick(change: PictureInPictureEvent) { await load(); change.stopPropagation(); }',
      'after-suspension',
      'change.stopPropagation()',
    ),
    invalid(
      'file.ts',
      'async function handleClick(request: FetchEvent) { await load(); request.preventDefault(); }',
      'after-suspension',
      'request.preventDefault()',
    ),
    invalid(
      'file.ts',
      'async function handleClick(value: Event) { await load(); value!.preventDefault(); }',
      'after-suspension',
      'value.preventDefault()',
    ),
    invalid(
      'file.ts',
      'async function handleClick(value: Event) { await load(); (value as Event).stopPropagation(); }',
      'after-suspension',
      'value.stopPropagation()',
    ),
    invalid(
      'file.ts',
      'async function handleClick(value: Event) { await load(); (<Event>value).stopImmediatePropagation(); }',
      'after-suspension',
      'value.stopImmediatePropagation()',
    ),
    invalid(
      'file.ts',
      'import type {SyntheticEvent} from "react"; async function handleClick(value: SyntheticEvent) { await load(); value.preventDefault(); }',
      'after-suspension',
      'value.preventDefault()',
    ),
    invalid(
      'file.ts',
      'import type {BaseSyntheticEvent} from "react"; async function handleClick(value: BaseSyntheticEvent) { await load(); value.preventDefault(); }',
      'after-suspension',
      'value.preventDefault()',
    ),
    invalid(
      'file.ts',
      'import type {SyntheticEvent as SE} from "react"; async function handleClick(value: SE) { await load(); value.preventDefault(); }',
      'after-suspension',
      'value.preventDefault()',
    ),
    invalid(
      'file.ts',
      'import type {InputEvent as Input} from "react"; async function handleClick(value: Input) { await load(); value.preventDefault(); }',
      'after-suspension',
      'value.preventDefault()',
    ),
    invalid(
      'file.ts',
      'import type {ChangeEvent} from "react"; async function handleClick(value: ChangeEvent<HTMLInputElement>) { await load(); value.preventDefault(); }',
      'after-suspension',
      'value.preventDefault()',
    ),
    invalid(
      'file.ts',
      'import type {ToggleEvent} from "react"; async function handleClick(value: ToggleEvent) { await load(); value.preventDefault(); }',
      'after-suspension',
      'value.preventDefault()',
    ),
    // Computed property names
    invalid(
      'file.js',
      'async event => { await somePromise; event["preventDefault"](); }',
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      'async event => { await somePromise; event[`preventDefault`](); }',
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      "async event => { await somePromise; event['stopPropagation'](); }",
      'after-suspension',
      'event.stopPropagation()',
    ),
    invalid(
      'file.js',
      'async event => { await somePromise; event?.["preventDefault"](); }',
      'after-suspension',
      'event.preventDefault()',
    ),
    // Documentation examples
    invalid(
      'file.js',
      `element.addEventListener('click', async event => {
	await save();
	event.preventDefault();
});`,
      'after-suspension',
      'event.preventDefault()',
    ),
    invalid(
      'file.js',
      `element.addEventListener('click', event => {
	setTimeout(() => {
		event.stopPropagation();
	});
});`,
      'in-nested-function',
      'event.stopPropagation()',
    ),
  ],
});
