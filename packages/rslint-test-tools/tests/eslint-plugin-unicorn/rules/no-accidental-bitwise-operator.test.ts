// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-accidental-bitwise-operator.js
import {
  RuleTester,
  type InvalidTestCase,
  type ValidTestCase,
} from '../rule-tester';

const valid = (code: string, filename = 'src/virtual.js'): ValidTestCase => ({
  code,
  filename,
});

const messages = {
  '&': ['&&', 'Unexpected bitwise operator `&`. Did you mean the logical operator `&&`?'],
  '|': ['||', 'Unexpected bitwise operator `|`. Did you mean the logical operator `||`?'],
  '|=': ['||=', 'Unexpected bitwise operator `|=`. Did you mean the logical operator `||=`?'],
} as const;

const invalid = (
  code: string,
  operator: keyof typeof messages,
  filename = 'src/virtual.js',
): InvalidTestCase => ({
  code,
  filename,
  errors: [
    {
      messageId: 'no-accidental-bitwise-operator/error',
      message: messages[operator][1],
    },
  ],
});

const validCases: ValidTestCase[] = [
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
];

const invalidCases: InvalidTestCase[] = [
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
];

new RuleTester().run('no-accidental-bitwise-operator', {} as never, {
  valid: validCases,
  invalid: invalidCases,
});
