// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-accidental-bitwise-operator.js
import { RuleTester } from '../rule-tester';

const valid = (code: string, filename = 'src/virtual.js') => ({
  code,
  filename,
  languageOptions: { sourceType: 'module' as const },
});

// The Go suite also asserts ranges and applied suggestions.
function invalid(code: string, bitwise: '&' | '|' | '|=', filename?: string) {
  const logical = { '&': '&&', '|': '||', '|=': '||=' }[bitwise];
  return {
    ...valid(code, filename),
    errors: [
      {
        messageId: 'no-accidental-bitwise-operator/error',
        message: `Unexpected bitwise operator \`${bitwise}\`. Did you mean the logical operator \`${logical}\`?`,
      },
    ],
  };
}

new RuleTester().run('no-accidental-bitwise-operator', {} as never, {
  valid: [
    // Snapshot cases
    valid('a | b;'),
    valid('a & b;'),
    valid('flags & MASK;'),
    valid('x | 0;'),
    valid('x | 1;'),
    valid('options | someVariable;'),
    valid('a ^ b;'),
    valid('a << b;'),
    valid('a >> b;'),
    valid('a >>> b;'),
    valid('~x;'),
    valid('x | (a + b);'),
    valid('foo() | bar();'),
    valid('x | null;'),
    valid('x | 1n;'),
    valid('x | /regex/;'),
    valid('a && b;'),
    valid('a || b;'),
    valid('obj && obj.prop;'),
    valid('obj1 & obj2.a;'),
    valid('obj.a & obj.b;'),
    valid('obj & obj.a.b;'),
    valid('a & b.c;'),
    valid('this & this.a;'),
    valid('obj & obj;'),
    valid('obj & obj.prop();'),
    valid('obj & obj?.a;'),
    valid('x &= {};'),
    valid('x &= 1;'),
    valid('x |= 1;'),
    valid('x |= y;'),
    valid('options | ({} as Foo);', 'src/virtual.ts'),
    // Documentation examples
    valid('if (object && object.property) {\n\t// …\n}'),
    valid('options = options || {};'),
    valid("input ||= '';"),
    valid('const masked = flags & MASK;\nconst truncated = value | 0;'),
  ],
  invalid: [
    // Grouping and statement boundaries
    invalid("ready && value | '';", '|'),
    invalid('obj & obj.prop | fallback;', '&'),
    invalid("value | '' ?? fallback;", '|'),
    invalid("fallback ?? value | '';", '|'),
    invalid("ready && (value | '');", '|'),
    invalid("ready && value /* keep */ | '';", '|'),
    invalid('(obj & obj.prop) === expected;', '&'),
    invalid('foo()\nobj & obj.prop | fallback', '&'),
    invalid("foo()\nvalue | '' ?? fallback", '|'),
    invalid('if (ready)\nobj & obj.prop | fallback', '&'),
    // Snapshot cases
    invalid('obj & obj.a;', '&'),
    invalid('if (obj & obj.prop) {}', '&'),
    invalid('obj & obj[key];', '&'),
    invalid('(obj) & obj.a;', '&'),
    invalid('obj /* comment */ & obj.a;', '&'),
    invalid('options | {};', '|'),
    invalid("options | '';", '|'),
    invalid('options | true;', '|'),
    invalid('options | [];', '|'),
    invalid('options | `template`;', '|'),
    invalid('x | function () {};', '|'),
    invalid('x | (() => {});', '|'),
    invalid('x | class {};', '|'),
    invalid('foo() | {};', '|'),
    invalid('a.b | {};', '|'),
    invalid('a | b | {};', '|'),
    invalid("input |= '';", '|='),
    invalid('input |= {};', '|='),
    invalid('input |= false;', '|='),
    invalid('obj & obj.a;', '&', 'src/virtual.ts'),
    invalid('options | {};', '|', 'src/virtual.ts'),
    // Compatible logical precedence
    invalid('ready || object & object.value;', '&'),
    invalid('ready && object & object.value;', '&'),
    invalid("value | '' || fallback;", '|'),
    // Documentation examples
    invalid('if (object & object.property) {\n\t// …\n}', '&'),
    invalid('options = options | {};', '|'),
    invalid("input |= '';", '|='),
  ],
});
