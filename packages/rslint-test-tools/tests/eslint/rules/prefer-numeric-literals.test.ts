import { RuleTester, type InvalidTestCase } from '../rule-tester';

const ruleTester = new RuleTester();

function invalid(code: string, target: string): InvalidTestCase {
  const index = code.indexOf(target);
  if (index < 0) {
    throw new Error(
      `target not found in prefer-numeric-literals test: ${target}`,
    );
  }
  const before = code.slice(0, index);
  const lines = before.split('\n');
  return {
    code,
    errors: [
      {
        messageId: 'useLiteral',
        line: lines.length,
        column: lines[lines.length - 1].length + 1,
      },
    ],
  };
}

const validCases = [
  'parseInt(1);',
  'parseInt(1, 3);',
  'Number.parseInt(1);',
  'Number.parseInt(1, 3);',
  '0b111110111 === 503;',
  '0o767 === 503;',
  '0x1F7 === 503;',
  'a[parseInt](1,2);',
  'parseInt(foo);',
  'parseInt(foo, 2);',
  'Number.parseInt(foo);',
  'Number.parseInt(foo, 2);',
  'parseInt(11, 2);',
  'Number.parseInt(1, 8);',
  'parseInt(1e5, 16);',
  "parseInt('11', '2');",
  "Number.parseInt('11', '8');",
  'parseInt(/foo/, 2);',
  'parseInt(`11${foo}`, 2);',
  "parseInt('11', 2n);",
  "Number.parseInt('11', 8n);",
  "parseInt('11', 16n);",
  'parseInt(`11`, 16n);',
  'parseInt(1n, 2);',
  'class C { #parseInt; foo() { Number.#parseInt("111110111", 2); } }',

  // Shadowed `parseInt` and `Number` should not be reported.
  'function foo(parseInt) { parseInt("111110111", 2); }',
  'function foo() { var parseInt; parseInt("111110111", 2); }',
  'function foo(Number) { Number.parseInt("111110111", 2); }',
  'function foo() { var Number; Number.parseInt("111110111", 2); }',
];

const invalidCases = [
  invalid('parseInt("111110111", 2) === 503;', 'parseInt("111110111", 2)'),
  invalid('parseInt("767", 8) === 503;', 'parseInt("767", 8)'),
  invalid('parseInt("1F7", 16) === 255;', 'parseInt("1F7", 16)'),
  invalid(
    'Number.parseInt("111110111", 2) === 503;',
    'Number.parseInt("111110111", 2)',
  ),
  invalid('Number.parseInt("767", 8) === 503;', 'Number.parseInt("767", 8)'),
  invalid('Number.parseInt("1F7", 16) === 255;', 'Number.parseInt("1F7", 16)'),
  invalid("parseInt('7999', 8);", "parseInt('7999', 8)"),
  invalid("parseInt('1234', 2);", "parseInt('1234', 2)"),
  invalid("parseInt('1234.5', 8);", "parseInt('1234.5', 8)"),
  invalid(
    "parseInt('\\u0031\\ufe0f\\u20e3\\u0033\\ufe0f\\u20e3\\u0033\\ufe0f\\u20e3\\u0037\\ufe0f\\u20e3', 16);",
    "parseInt('\\u0031\\ufe0f\\u20e3\\u0033\\ufe0f\\u20e3\\u0033\\ufe0f\\u20e3\\u0037\\ufe0f\\u20e3', 16)",
  ),
  invalid("Number.parseInt('7999', 8);", "Number.parseInt('7999', 8)"),
  invalid("Number.parseInt('1234', 2);", "Number.parseInt('1234', 2)"),
  invalid("Number.parseInt('1234.5', 8);", "Number.parseInt('1234.5', 8)"),
  invalid(
    "Number.parseInt('\\u0031\\ufe0f\\u20e3\\u0033\\ufe0f\\u20e3\\u0033\\ufe0f\\u20e3\\u0037\\ufe0f\\u20e3', 16);",
    "Number.parseInt('\\u0031\\ufe0f\\u20e3\\u0033\\ufe0f\\u20e3\\u0033\\ufe0f\\u20e3\\u0037\\ufe0f\\u20e3', 16)",
  ),
  invalid('parseInt(`111110111`, 2) === 503;', 'parseInt(`111110111`, 2)'),
  invalid('parseInt(`767`, 8) === 503;', 'parseInt(`767`, 8)'),
  invalid('parseInt(`1F7`, 16) === 255;', 'parseInt(`1F7`, 16)'),
  invalid("parseInt('', 8);", "parseInt('', 8)"),
  invalid('parseInt(``, 8);', 'parseInt(``, 8)'),
  invalid('parseInt(`7999`, 8);', 'parseInt(`7999`, 8)'),
  invalid('parseInt(`1234`, 2);', 'parseInt(`1234`, 2)'),
  invalid('parseInt(`1234.5`, 8);', 'parseInt(`1234.5`, 8)'),

  // Adjacent tokens tests
  invalid("parseInt('11', 2)", "parseInt('11', 2)"),
  invalid("Number.parseInt('67', 8)", "Number.parseInt('67', 8)"),
  invalid("5+parseInt('A', 16)", "parseInt('A', 16)"),
  invalid(
    "function *f(){ yield(Number).parseInt('11', 2) }",
    "(Number).parseInt('11', 2)",
  ),
  invalid(
    "function *f(){ yield(Number.parseInt)('67', 8) }",
    "(Number.parseInt)('67', 8)",
  ),
  invalid("function *f(){ yield(parseInt)('A', 16) }", "(parseInt)('A', 16)"),
  invalid(
    "function *f(){ yield Number.parseInt('11', 2) }",
    "Number.parseInt('11', 2)",
  ),
  invalid(
    "function *f(){ yield/**/Number.parseInt('67', 8) }",
    "Number.parseInt('67', 8)",
  ),
  invalid("function *f(){ yield(parseInt('A', 16)) }", "parseInt('A', 16)"),
  invalid("parseInt('11', 2)+5", "parseInt('11', 2)"),
  invalid("Number.parseInt('17', 8)+5", "Number.parseInt('17', 8)"),
  invalid("parseInt('A', 16)+5", "parseInt('A', 16)"),
  invalid("parseInt('11', 2)in foo", "parseInt('11', 2)"),
  invalid("Number.parseInt('17', 8)in foo", "Number.parseInt('17', 8)"),
  invalid("parseInt('A', 16)in foo", "parseInt('A', 16)"),
  invalid("parseInt('11', 2) in foo", "parseInt('11', 2)"),
  invalid("Number.parseInt('17', 8)/**/in foo", "Number.parseInt('17', 8)"),
  invalid("(parseInt('A', 16))in foo", "parseInt('A', 16)"),

  // Should not autofix if it would remove comments
  invalid("/* comment */Number.parseInt('11', 2);", "Number.parseInt('11', 2)"),
  invalid("Number/**/.parseInt('11', 2);", "Number/**/.parseInt('11', 2)"),
  invalid("Number//\n.parseInt('11', 2);", "Number//\n.parseInt('11', 2)"),
  invalid("Number./**/parseInt('11', 2);", "Number./**/parseInt('11', 2)"),
  invalid("Number.parseInt(/**/'11', 2);", "Number.parseInt(/**/'11', 2)"),
  invalid("Number.parseInt('11', /**/2);", "Number.parseInt('11', /**/2)"),
  invalid("Number.parseInt('11', 2)/* comment */;", "Number.parseInt('11', 2)"),
  invalid("parseInt/**/('11', 2);", "parseInt/**/('11', 2)"),
  invalid("parseInt(//\n'11', 2);", "parseInt(//\n'11', 2)"),
  invalid("parseInt('11'/**/, 2);", "parseInt('11'/**/, 2)"),
  invalid('parseInt(`11`/**/, 2);', 'parseInt(`11`/**/, 2)'),
  invalid("parseInt('11', 2 /**/);", "parseInt('11', 2 /**/)"),
  invalid("parseInt('11', 2)//comment\n;", "parseInt('11', 2)"),

  // Optional chaining
  invalid('parseInt?.("1F7", 16) === 255;', 'parseInt?.("1F7", 16)'),
  invalid(
    'Number?.parseInt("1F7", 16) === 255;',
    'Number?.parseInt("1F7", 16)',
  ),
  invalid(
    'Number?.parseInt?.("1F7", 16) === 255;',
    'Number?.parseInt?.("1F7", 16)',
  ),
  invalid(
    '(Number?.parseInt)("1F7", 16) === 255;',
    '(Number?.parseInt)("1F7", 16)',
  ),
  invalid(
    '(Number?.parseInt)?.("1F7", 16) === 255;',
    '(Number?.parseInt)?.("1F7", 16)',
  ),

  // `parseInt` doesn't support numeric separators. The rule shouldn't autofix in those cases.
  invalid("parseInt('1_0', 2);", "parseInt('1_0', 2)"),
  invalid("Number.parseInt('5_000', 8);", "Number.parseInt('5_000', 8)"),
  invalid("parseInt('0_1', 16);", "parseInt('0_1', 16)"),
  invalid("Number.parseInt('0_0', 16);", "Number.parseInt('0_0', 16)"),
];

for (const [index, cases] of chunks(validCases, 10).entries()) {
  ruleTester.run(
    'prefer-numeric-literals',
    {
      valid: cases,
      invalid: [],
    },
    `prefer-numeric-literals valid ${index + 1}`,
  );
}

for (const [index, cases] of chunks(invalidCases, 10).entries()) {
  ruleTester.run(
    'prefer-numeric-literals',
    {
      valid: [],
      invalid: cases,
    },
    `prefer-numeric-literals invalid ${index + 1}`,
  );
}

function chunks<T>(items: T[], size: number): T[][] {
  const result: T[][] = [];
  for (let index = 0; index < items.length; index += size) {
    result.push(items.slice(index, index + size));
  }
  return result;
}
