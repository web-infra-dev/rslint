// Upstream: eslint-plugin-unicorn v75.0.0, test/prefer-reflect-apply.js
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const valid = (code: string) => ({ code, filename: 'file.js' });
const invalid = (code: string, output: string) => ({
  code,
  output,
  filename: 'file.js',
  errors: [
    {
      messageId: 'prefer-reflect-apply',
      message: 'Prefer `Reflect.apply()` over `Function#apply()`.',
    },
  ],
});

ruleTester.run('prefer-reflect-apply', null as never, {
  valid: [
    'foo.apply();',
    'foo.apply(null);',
    'foo.apply(this);',
    'foo.apply(null, 42);',
    'foo.apply(this, 42);',
    'foo.apply(bar, arguments);',
    '[].apply(null, [42]);',
    'foo.apply(bar);',
    'foo.apply(bar, []);',
    'foo.apply;',
    'apply;',
    'Reflect.apply(foo, null);',
    'Reflect.apply(foo, null, [bar]);',
    'const apply = "apply"; foo[apply](null, [42]);',
    // Documentation examples.
    ...['null', 'this'].flatMap((receiver) =>
      ['[42]', 'arguments'].map(
        (args) =>
          `function foo() {}\nReflect.apply(foo, ${receiver}, ${args});`,
      ),
    ),
  ].map(valid),
  invalid: [
    invalid('foo.apply(null, [42]);', 'Reflect.apply(foo, null, [42]);'),
    invalid('foo.apply(null, []);', 'Reflect.apply(foo, null, []);'),
    invalid(
      '(foo.bar).apply(null, [42]);',
      'Reflect.apply(foo.bar, null, [42]);',
    ),
    invalid(
      'foo.bar.apply(null, [42]);',
      'Reflect.apply(foo.bar, null, [42]);',
    ),
    invalid(
      'Function.prototype.apply.call(foo, null, [42]);',
      'Reflect.apply(foo, null, [42]);',
    ),
    invalid(
      'Function.prototype.apply.call(foo.bar, null, [42]);',
      'Reflect.apply(foo.bar, null, [42]);',
    ),
    invalid(
      'foo.apply(null, arguments);',
      'Reflect.apply(foo, null, arguments);',
    ),
    invalid(
      'Function.prototype.apply.call(foo, null, arguments);',
      'Reflect.apply(foo, null, arguments);',
    ),
    invalid('foo.apply(this, [42]);', 'Reflect.apply(foo, this, [42]);'),
    invalid(
      'Function.prototype.apply.call(foo, this, [42]);',
      'Reflect.apply(foo, this, [42]);',
    ),
    invalid(
      'foo.apply(this, arguments);',
      'Reflect.apply(foo, this, arguments);',
    ),
    invalid(
      'Function.prototype.apply.call(foo, this, arguments);',
      'Reflect.apply(foo, this, arguments);',
    ),
    invalid('foo["apply"](null, [42]);', 'Reflect.apply(foo, null, [42]);'),
    ...['null', 'this'].flatMap((receiver) =>
      ['[42]', 'arguments'].flatMap((args) =>
        [
          `foo.apply(${receiver}, ${args})`,
          `Function.prototype.apply.call(foo, ${receiver}, ${args})`,
        ].map((call) =>
          invalid(
            `function foo() {}\n${call};`,
            `function foo() {}\nReflect.apply(foo, ${receiver}, ${args});`,
          ),
        ),
      ),
    ),
  ],
});
