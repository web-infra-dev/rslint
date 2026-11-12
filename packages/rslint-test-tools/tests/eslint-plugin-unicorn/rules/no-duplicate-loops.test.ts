// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-duplicate-loops.js
import { RuleTester } from '../rule-tester';

const valid = (code: string, filename = 'src/virtual.js') => ({
  code,
  filename,
  languageOptions: { sourceType: 'module' as const },
});

// The Go suite also verifies the full diagnostic ranges and absence of edits.
function invalid(
  code: string,
  method: 'map' | 'filter',
  loopKind: 'for…of' | 'for await…of',
  filename?: string,
) {
  return {
    ...valid(code, filename),
    errors: [
      {
        messageId: 'no-duplicate-loops',
        message:
          'Do not use `.' +
          method +
          '()` directly in a `' +
          loopKind +
          '` loop header. It creates an intermediate array before the loop.',
      },
    ],
  };
}

new RuleTester().run('no-duplicate-loops', {} as never, {
  valid: [
    valid(
      'const iterator = items.values(); for (const item of iterator.map(callback)) {}',
    ),
    valid(
      'function * generate() { yield 1; } for (const item of generate().filter(callback)) {}',
    ),
    valid('for (const item of items) {}'),
    valid('for await (const item of items) {}'),
    valid('for (const key in items.map(callback)) {}'),
    valid('items.map(callback);'),
    valid('items.filter(callback);'),
    valid('const mapped = items.map(callback); for (const item of mapped) {}'),
    valid('for (const item of items.map()) {}'),
    valid('for (const item of items.filter()) {}'),
    valid('for (const item of items?.map(callback)) {}'),
    valid('for (const item of items.map?.(callback)) {}'),
    valid('for (const item of items["map"](callback)) {}'),
    valid('for (const item of items[method](callback)) {}'),
    valid('for (const [index, item] of items.entries()) {}'),
    valid('for (const [index, item] of items.map(callback).entries()) {}'),
    valid('for (const [index, item] of items.filter(callback).entries()) {}'),
    valid('for (const item of Array.from(items)) {}'),
    valid('for (const item of [...items]) {}'),
    valid('for (const item of Array.prototype.map.call(items, callback)) {}'),
    valid('for (const item of Iterator.from(items).map(callback)) {}'),
    valid('for (const item of Iterator.from(items).filter(callback)) {}'),
    valid(
      'for (const item of Iterator.from(items).map(callback).filter(callback)) {}',
    ),
    valid('for (const item of (Iterator).from(items).map(callback)) {}'),
    valid(
      'for (const item of Iterator.concat(first, second).filter(callback)) {}',
    ),
    valid('for (const item of Iterator.zip(iterables).map(callback)) {}'),
    valid(
      'for (const item of Iterator.zipKeyed(iterables).filter(callback)) {}',
    ),
    valid(
      'for (const item of globalThis.Iterator.from(items).map(callback)) {}',
    ),
    valid(
      'for (const item of (globalThis).Iterator.from(items).map(callback)) {}',
    ),
    valid(
      'for (const item of globalThis.Iterator.zip(iterables).filter(callback)) {}',
    ),
    valid('for (const item of items.values().map(callback)) {}'),
    valid('for (const item of items.keys().map(callback)) {}'),
    valid('for (const item of items.entries().filter(callback)) {}'),
    valid('for (const item of items.matchAll(pattern).map(callback)) {}'),
    valid('for (const item of Iterator.from(items).take(1).map(callback)) {}'),
    valid(
      'for (const item of Iterator.from(items).flatMap(callback).drop(1).filter(callback)) {}',
    ),
    valid(
      'function foo(items: string[]) { for (const item of items) {} }',
      'src/virtual.ts',
    ),
    valid(
      'function foo(iterator: IteratorObject<string> | undefined) { for (const item of iterator!.map(callback)) {} }',
      'src/virtual.ts',
    ),
    valid(
      'for (const item of (Iterator as typeof Iterator).from(items).map(callback)) {}',
      'src/virtual.ts',
    ),
    valid(
      'for (const item of (Iterator.from(items) as Iterator<string>).map(callback)) {}',
      'src/virtual.ts',
    ),
    valid(
      'for (const item of (items as Iterator<string>).map(callback)) {}',
      'src/virtual.ts',
    ),
    // Documentation examples
    valid(`for (const value of values) {
	console.log(value.property);
}`),
    valid(`for (const value of values) {
	if (value.isEnabled) {
		console.log(value);
	}
}`),
  ],
  invalid: [
    invalid('for (const item of items.map(callback)) {}', 'map', 'for…of'),
    invalid(
      'for (const item of items.filter(callback)) {}',
      'filter',
      'for…of',
    ),
    invalid(
      'for (const item of items.map(callback, thisArgument)) {}',
      'map',
      'for…of',
    ),
    invalid(
      'for (const item of items.filter(callback, thisArgument)) {}',
      'filter',
      'for…of',
    ),
    invalid('for (const item of (items).map(callback)) {}', 'map', 'for…of'),
    invalid(
      'for (const item of ((items)).filter(callback)) {}',
      'filter',
      'for…of',
    ),
    invalid('for (const item of (items.map(callback))) {}', 'map', 'for…of'),
    invalid(
      'for await (const item of items.map(callback)) {}',
      'map',
      'for await…of',
    ),
    invalid(
      'for await (const item of items.filter(async item => item.isEnabled)) {}',
      'filter',
      'for await…of',
    ),
    invalid(
      'for (const item of items.map(/* comment */ callback)) {}',
      'map',
      'for…of',
    ),
    invalid(
      'const Iterator = {from: items => items}; for (const item of Iterator.from(items).map(callback)) {}',
      'map',
      'for…of',
    ),
    invalid(
      'const Iterator = {zip: iterables => iterables}; for (const item of Iterator.zip(iterables).filter(callback)) {}',
      'filter',
      'for…of',
    ),
    invalid(
      'const globalThis = {Iterator: {from: items => items}}; for (const item of globalThis.Iterator.from(items).map(callback)) {}',
      'map',
      'for…of',
    ),
    invalid(
      `for (const item of items
	.map(item => item.value)
) {}`,
      'map',
      'for…of',
    ),
    invalid(
      `for (
	const item of items.filter(item => item.isEnabled)
) {}`,
      'filter',
      'for…of',
    ),
    invalid(
      'function foo(items: string[]) { for (const item of (items as string[]).map(item => item.trim())) {} }',
      'map',
      'for…of',
      'src/virtual.ts',
    ),
    invalid(
      'for (const item of items.map(callback) as string[]) {}',
      'map',
      'for…of',
      'src/virtual.ts',
    ),
    // Documentation examples
    invalid(
      `for (const value of values.map(value => value.property)) {
	console.log(value);
}`,
      'map',
      'for…of',
    ),
    invalid(
      `for (const value of values.filter(value => value.isEnabled)) {
	console.log(value);
}`,
      'filter',
      'for…of',
    ),
  ],
});
