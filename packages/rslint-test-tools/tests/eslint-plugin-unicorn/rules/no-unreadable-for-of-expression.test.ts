// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-unreadable-for-of-expression.js
import { RuleTester } from '../rule-tester';

const valid = (code: string, filename = 'src/virtual.js') => ({
  code,
  filename,
  languageOptions: { sourceType: 'module' as const },
});

function invalid(code: string, loopKind = 'for…of', filename?: string) {
  return {
    ...valid(code, filename),
    errors: [
      {
        messageId: 'no-unreadable-for-of-expression',
        message:
          'Move the complex iterable expression out of the `' +
          loopKind +
          '` loop header.',
      },
    ],
  };
}

new RuleTester().run('no-unreadable-for-of-expression', {} as never, {
  valid: [
    // Template literals
    valid('for (const character of `unicorn`) {}'),
    valid('for await (const character of ``) {}'),
    valid('for (const character of `uni\\u0063orn`) {}'),
    // Upstream snapshot cases
    valid('for (const [name, score] of Iterator.zip([names, scores])) {}'),
    valid(
      'for (const item of combine([first, object.second, getThird(), "fourth"])) {}',
    ),
    valid('for (const item of combine([])) {}'),
    valid('for (const item of items) {}'),
    valid('for await (const item of items) {}'),
    valid('for (const item of object.items) {}'),
    valid('for (const item of this.items) {}'),
    valid(
      'class Foo extends Bar { method() { for (const item of super.items) {} } }',
    ),
    valid('for (const item of [first, second]) {}'),
    valid('for (const [index, item] of [first, second].entries()) {}'),
    valid('for (const item of getItems()) {}'),
    valid('for (const item of getItems(argument)) {}'),
    valid('for (const item of getItems(first, object.second, "third")) {}'),
    valid('for (const item of getItems(`argument`)) {}'),
    valid('for (const item of getItems(createArgument())) {}'),
    valid('for (const character of "abc") {}'),
    valid('for (const part of "a,b".split(",")) {}'),
    valid('for (const item of getItems(`prefix-${argument}`)) {}'),
    valid(
      'for (const element of root.querySelectorAll(`[data-x=${value}]`)) {}',
    ),
    valid('for (const item of object.getItems(argument)) {}'),
    valid('for (const item of object.getItems()) {}'),
    valid('for (const item of Object.keys(object)) {}'),
    valid('for (const item of Object.values(object.items)) {}'),
    valid('for (const [key, value] of Object.entries(getObject())) {}'),
    valid(
      'const Object = {keys: value => value}; for (const item of Object.keys(object)) {}',
    ),
    valid('for (const item of (items)) {}'),
    valid('for (const item of ((object.items))) {}'),
    valid('for (const item of items.map(item => item.value)) {}'),
    valid('for (const item of items.map(callback)) {}'),
    valid('for (const item of (items.map(callback))) {}'),
    valid('for (const item of items.filter(item => item.isEnabled)) {}'),
    valid('for await (const item of items.map(callback)) {}'),
    valid('for await (const item of items.filter(callback)) {}'),
    valid(
      'const Iterator = {from: value => value}; for (const item of Iterator.from(items).map(callback)) {}',
    ),
    valid(
      'const Iterator = {concat: value => value}; for (const item of Iterator.concat(items).map(callback)) {}',
    ),
    valid(
      'const Iterator = {zip: value => value}; for (const item of Iterator.zip(items).filter(callback)) {}',
    ),
    valid(
      'const Iterator = {zipKeyed: value => value}; for (const item of Iterator.zipKeyed(items).take(1)) {}',
    ),
    valid(
      'const globalThis = {Iterator}; for (const item of globalThis.Iterator.from(items).map(callback)) {}',
    ),
    valid('for (const item of items as string[]) {}', 'src/virtual.ts'),
    valid(
      'for (const item of (items satisfies Iterable<string>)) {}',
      'src/virtual.ts',
    ),
    valid('for (const item of items!) {}', 'src/virtual.ts'),
    valid('for (const item of getItems() as string[]) {}', 'src/virtual.ts'),
    valid(
      'for (const item of items.map(callback) as string[]) {}',
      'src/virtual.ts',
    ),
    valid(
      'for (const item of items.filter(callback) as string[]) {}',
      'src/virtual.ts',
    ),
    valid(
      'for (const pair of Iterator.zip([names, scores] as const)) {}',
      'src/virtual.ts',
    ),
    valid(
      'for (const pair of Iterator.zip([names, scores] satisfies Iterable<unknown>[])) {}',
      'src/virtual.ts',
    ),
    valid(
      'for (const pair of Iterator.zip(([names, scores])!)) {}',
      'src/virtual.ts',
    ),
    valid(
      'for (const pair of Iterator.zip([names as string[], scores!])) {}',
      'src/virtual.ts',
    ),
    valid(
      'class Foo { #items; method() { for (const item of this.#items) {} } }',
    ),
    // Documentation examples (https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-unreadable-for-of-expression.md)
    valid(
      'const argument = createArgument(seed);\nfor (const item of getItems(argument)) {\n\tconsole.log(item);\n}',
    ),
    valid(
      'const object = Object.fromEntries(entries);\nfor (const key of Object.keys(object)) {\n\tconsole.log(key);\n}',
    ),
    valid('for (const item of getItems(argument)) {\n\tconsole.log(item);\n}'),
    valid('for (const item of items) {\n\tconsole.log(item);\n}'),
    valid(
      'for (const value of Object.values(object)) {\n\tconsole.log(value);\n}',
    ),
  ],
  invalid: [
    // Upstream snapshot cases
    invalid('for (const pair of Iterator.zip([names, getScores(seed)])) {}'),
    invalid('for (const pair of Iterator.zip([...inputs])) {}'),
    invalid('for (const pair of Iterator.zip([[first], [second]])) {}'),
    invalid('for (const pair of Iterator.zip([first, , second])) {}'),
    invalid('for (const item of getItems(createArgument(seed))) {}'),
    invalid('for (const item of getItems(argument || fallback)) {}'),
    invalid('for (const item of Object.keys({a: 1})) {}'),
    invalid('for (const item of (header || "").split(",")) {}'),
    invalid('for (const item of (condition ? first : second).values()) {}'),
    invalid('for (const item of items[method]()) {}'),
    invalid('for (const item of items[index]) {}'),
    invalid('for (const item of items?.values()) {}'),
    invalid('for (const item of condition ? first : second) {}'),
    invalid('for (const item of first || second) {}'),
    invalid('for (const item of new Set(items)) {}'),
    invalid('for (const item of tag`items`) {}'),
    invalid('for (const item of Object.keys(Object.fromEntries(entries))) {}'),
    invalid(
      'for (const item of Object.values(items.filter(item => item.isEnabled))) {}',
    ),
    invalid('for (const item of items?.map(item => item.value)) {}'),
    invalid('for (const item of items?.map(callback)) {}'),
    invalid('for (const item of items.filter?.(item => item.isEnabled)) {}'),
    invalid('for (const item of items.filter?.(callback)) {}'),
    invalid(
      'for (const item of globalThis.Iterator.from(items).map(callback)) {}',
    ),
    invalid(
      'for (const item of (globalThis).Iterator.from(items).map(callback)) {}',
    ),
    invalid('for (const item of (Iterator).from(items).map(callback)) {}'),
    invalid(
      'for (const item of Iterator.from(items).map(item => item.value)) {}',
    ),
    invalid('for (const item of Iterator.from(items).map(callback)) {}'),
    invalid(
      'for (const item of Iterator.from(items).filter(item => item.isEnabled)) {}',
    ),
    invalid('for (const item of Iterator.from(items).filter(callback)) {}'),
    invalid('for (const item of Iterator.from(items).take(1)) {}'),
    invalid('for (const item of Iterator.from(items).drop(1)) {}'),
    invalid('for (const item of Iterator.from(items).flatMap(callback)) {}'),
    invalid(
      'for (const item of Iterator.concat(first, second).map(callback)) {}',
    ),
    invalid('for (const item of Iterator.zip(iterables).filter(callback)) {}'),
    invalid('for (const item of Iterator.zipKeyed(iterables).take(1)) {}'),
    invalid(
      'for (const item of Iterator.from(items).take(1).map(callback)) {}',
    ),
    invalid('for (const item of items.values().map(item => item.value)) {}'),
    invalid('for (const item of items.values().map(callback)) {}'),
    invalid('for (const item of items.values().filter(callback)) {}'),
    invalid('for (const item of items.values().take(1)) {}'),
    invalid('for (const item of items.values().take(1).map(callback)) {}'),
    invalid('for (const item of items.keys().map(callback)) {}'),
    invalid('for (const item of items.entries().filter(callback)) {}'),
    invalid('for (const item of items.matchAll(pattern).map(callback)) {}'),
    invalid(
      'for await (const item of getItems(createArgument(seed))) {}',
      'for await…of',
    ),
    invalid(
      'for (const item of getItems(createArgument(seed)) as string[]) {}',
      'for…of',
      'src/virtual.ts',
    ),
    invalid(
      'for (const item of (Iterator as typeof Iterator).from(items).map(callback)) {}',
      'for…of',
      'src/virtual.ts',
    ),
    invalid(
      'for (const item of (Iterator.from(items) as Iterator<string>).map(callback)) {}',
      'for…of',
      'src/virtual.ts',
    ),
    invalid(
      'for (const item of Iterator.from(items).take(1) as Iterator<string>) {}',
      'for…of',
      'src/virtual.ts',
    ),
    invalid(
      'for (const item of (items[index] satisfies Iterable<string>)) {}',
      'for…of',
      'src/virtual.ts',
    ),
    invalid(
      'for (const pair of Iterator.zip([names, getScores(seed)] as const)) {}',
      'for…of',
      'src/virtual.ts',
    ),
    invalid(
      'for (const pair of Iterator.zip([[names], [scores]] as const)) {}',
      'for…of',
      'src/virtual.ts',
    ),
    invalid(
      'for (const pair of Iterator.zip([getNames(seed) as string[], scores])) {}',
      'for…of',
      'src/virtual.ts',
    ),
    invalid('for (\n\tconst item of getItems(createArgument(seed))\n) {}'),
    invalid(
      'const iterator = items.values(); for (const item of iterator.map(callback)) {}',
    ),
    invalid(
      'function * generate() { yield 1; } for (const item of generate().filter(callback)) {}',
    ),
    // Documentation examples (https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-unreadable-for-of-expression.md)
    invalid(
      'for (const item of getItems(createArgument(seed))) {\n\tconsole.log(item);\n}',
    ),
    invalid(
      'for (const key of Object.keys(Object.fromEntries(entries))) {\n\tconsole.log(key);\n}',
    ),
  ],
});
