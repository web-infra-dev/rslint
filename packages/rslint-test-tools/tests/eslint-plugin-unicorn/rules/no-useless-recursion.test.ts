// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-useless-recursion.js
import { RuleTester } from '@rslint/rule-tester';

function recursiveReturn(
  code: string,
  line: number,
  column: number,
  filename = 'case.js',
) {
  return {
    code,
    filename,
    errors: [
      {
        messageId: 'no-useless-recursion',
        message: 'Use a loop instead of this recursive function call.',
        line,
        column,
        endLine: line,
        endColumn: column + 3,
      },
    ],
  };
}

new RuleTester({ plugin: 'unicorn' }).run('no-useless-recursion', {
  valid: [
    // Also a documentation example.
    `function foo(bar) {
	while (bar.baz) {
		bar = bar.baz;
	}

	return bar;
}`,
    `function foo(bar) {
	for (; bar.baz; bar = bar.baz) {}

	return bar;
}`,
    // Also a documentation example.
    `function foo(bar) {
	if (Array.isArray(bar)) {
		return bar.map(baz => foo(baz));
	}

	return bar;
}`,
    `function foo(value) {
	return 1 + foo(value);
}`,
    `function foo(value) {
	return value ? foo(value.next) : value;
}`,
    `function foo(value) {
	function bar() {
		return foo(value);
	}

	return bar();
}`,
    `function foo(value) {
	const foo = () => value;
	return foo();
}`,
    `function foo(value) {
	return foo?.(value);
}`,
    `function foo(value) {
	return object.foo(value);
}`,
    `const foo = value => {
	return foo(value);
};`,
    `const foo = function (value) {
	return foo(value);
};`,
    `function * foo(value) {
	return foo(value);
}`,
    `const object = {
	foo(value) {
		return foo(value);
	},
};`,
    `class Foo {
	foo(value) {
		return foo(value);
	}
}`,
    `const object = {
	get foo() {
		return this.foo;
	},
};`,
    `function foo(value) {
	try {
		return foo(value.next);
	} finally {
		cleanup(value);
	}
}`,
    `function foo(value) {
	try {
		return foo(value.next);
	} catch (error) {
		if (value.retry) {
			throw error;
		}

		return value;
	}
}`,
    `function foo(value) {
	using resource = acquire(value);
	return foo(value.next);
}`,
    `async function foo(value) {
	await using resource = acquire(value);
	return foo(value.next);
}`,
    `function foo(value) {
	for (using resource = acquire(value); value.next;) {
		return foo(value.next);
	}
}`,
    `async function foo(value) {
	for (await using resource = acquire(value); value.next;) {
		return foo(value.next);
	}
}`,
    `function foo(value) {
	for (using resource of resources) {
		return foo(value.next);
	}
}`,
    `async function foo(value) {
	for await (await using resource of resources) {
		return foo(value.next);
	}
}`,
  ],
  invalid: [
    // Also a documentation example.
    recursiveReturn(
      `function foo(bar) {
	if (bar.baz) {
		return foo(bar.baz);
	}

	return bar;
}`,
      3,
      10,
    ),
    recursiveReturn(
      `const bar = function foo(value) {
	return foo(value.next);
};`,
      2,
      9,
    ),
    recursiveReturn(
      `function foo(first, second) {
	return foo(first.next, second.next);
}`,
      2,
      9,
    ),
    recursiveReturn(
      `function foo() {
	return foo();
}`,
      2,
      9,
    ),
    recursiveReturn(
      `async function foo(value) {
	if (value) {
		return foo(value.next);
	}

	return value;
}`,
      3,
      10,
    ),
    // TypeScript parser.
    recursiveReturn(
      `function foo(bar: Bar) {
	return foo(bar.baz as Bar);
}`,
      2,
      9,
      'case.ts',
    ),
  ],
});
