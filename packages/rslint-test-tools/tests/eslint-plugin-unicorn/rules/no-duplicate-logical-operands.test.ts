// Ported from eslint-plugin-unicorn v80.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/test/no-duplicate-logical-operands.js
import {
  RuleTester,
  type InvalidTestCase,
  type ValidTestCase,
} from '../rule-tester';

function valid(code: string, typescript = false): ValidTestCase {
  return {
    code,
    filename: typescript ? 'src/virtual.ts' : 'src/virtual.js',
    languageOptions: { sourceType: 'module' },
  };
}

// The wrapper checks messages; Go tests also assert ranges, fixes and suggestions.
function invalid(code: string, typescript = false): InvalidTestCase {
  return {
    ...valid(code, typescript),
    errors: [
      {
        messageId: 'no-duplicate-logical-operands',
        message: 'This operand duplicates the left operand.',
      },
    ],
  };
}

new RuleTester().run('no-duplicate-logical-operands', {} as never, {
  valid: [
    // Different operands.
    valid('foo && bar'),
    valid('foo || bar'),

    // Non-adjacent duplicates are ignored.
    valid('foo && bar && foo'),
    valid('foo && (foo && bar)'),

    // Not logical AND/OR.
    valid('foo ?? foo'),

    // Calls and other complex expressions can have side effects.
    valid('getValue() && getValue()'),
    valid('getValue() || getValue()'),
    valid('foo.bar() && foo.bar()'),
    valid('foo + bar && foo + bar'),
    valid('await foo && await foo'),
    valid('foo++ && foo++'),
    valid('(foo = bar) && (foo = bar)'),

    // Computed keys with complex expressions can have side effects.
    valid('foo[bar()] && foo[bar()]'),
    valid('foo[bar + baz] || foo[bar + baz]'),

    // Optional chains are skipped.
    valid('foo?.bar && foo?.bar'),
    valid('foo?.[bar] || foo?.[bar]'),

    // Constants are handled by ESLint core `no-constant-binary-expression`.
    valid('true && true'),
    valid('0 || 0'),

    // Identifier reads inside `with` can trigger proxy traps or getters.
    {
      code: 'with (scope) {foo && foo;}',
      filename: 'src/virtual.js',
      languageOptions: { sourceType: 'script' },
    },

    // Static string keys must be untagged templates without expressions.
    valid('foo[tag`bar`] || foo.bar;'),
    valid('foo[`bar${value}`] || foo.bar;'),
    // Optional members are never simple references.
    valid('foo?.[`bar`] || foo.bar;'),

    // Documentation: correct examples.
    valid('if (isReady) {}'),
    valid('const result = value;'),
    valid('if (object.isReady) {}'),
    valid('object.isReady;'),
  ],
  invalid: [
    // Identifiers are autofixed.
    invalid('foo && foo'),
    invalid('foo || foo'),
    invalid('foo && bar && bar'),
    invalid('foo || bar || bar'),
    invalid('(foo) && (foo)'),
    invalid('this && this'),

    // Member expressions are suggestions because property reads can be observable.
    invalid('foo.bar && foo.bar'),
    invalid('foo.bar || foo.bar'),
    invalid('foo.bar.baz && foo.bar.baz'),
    invalid('foo && bar.baz && bar.baz'),
    invalid('this.foo && this.foo'),
    invalid('class Foo {#foo; method() {return this.#foo && this.#foo;}}'),
    invalid(
      'class Foo extends Bar {method() {return super.foo && super.foo;}}',
    ),
    invalid('foo[bar] && foo[bar]'),
    invalid('foo["bar"] || foo.bar'),

    // TypeScript wrappers around references are ignored for matching.
    invalid('(foo as boolean) && (foo as boolean)', true),
    invalid('(foo as Foo) && (foo as Bar)', true),
    invalid('foo && (foo as Foo)', true),
    invalid('(<boolean>foo) && (<boolean>foo)', true),
    invalid('foo! || foo!', true),
    invalid('foo! && foo', true),
    invalid('foo && foo!', true),
    invalid('(foo satisfies boolean) && (foo satisfies boolean)', true),
    invalid('foo && (foo satisfies Foo)', true),

    // Comments outside the retained left operand disable fixes and suggestions.
    invalid('foo /* keep */ && foo'),
    invalid('foo && /* keep */ foo'),
    invalid('foo && bar /* keep */ && bar'),
    invalid('foo && (foo /* keep */)'),

    // Comments inside the retained left operand are preserved.
    invalid('(foo /* keep */) && foo'),

    // Static string keys, including templates, compare by decoded value.
    invalid('foo[`bar`] || foo.bar;'),
    invalid('foo["bar"] && foo[`b\\u0061r`];'),
    invalid('foo[``] || foo[""];'),

    // Reference contexts report without a fix or suggestion.
    invalid('(object.method || object.method)()'),
    invalid('(object.tag && object.tag)`text`'),
    invalid('delete (object.value || object.value)'),
    invalid('delete (value || value)'),
    invalid(
      'function check() { const local = 1; return (eval || eval)("typeof local"); }',
    ),
    invalid('(object.method || object.method)!()', true),
    invalid('delete ((object.value || object.value) as unknown)', true),
    // A constructor callee is not a reference context.
    invalid('new (Constructor || Constructor)()'),

    // Documentation: incorrect examples.
    invalid('if (isReady && isReady) {}'),
    invalid('const result = value || value;'),
    invalid('if (object.isReady && object.isReady) {}'),
    invalid('object.isReady && object.isReady;'),
  ],
});
