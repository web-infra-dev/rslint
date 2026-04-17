import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-implicit-coercion', {
  valid: [
    'Boolean(foo)',
    'foo.indexOf(1) !== -1',
    'Number(foo)',
    'parseInt(foo)',
    'parseFloat(foo)',
    'String(foo)',
    '!foo',
    '~foo',
    '-foo',
    '+1234',
    '-1234',
    '- -1234',
    '+Number(lol)',
    '-parseFloat(lol)',
    '2 * foo',
    '1 * 1234',
    '123 - 0',
    '1 * Number(foo)',
    '1 * parseInt(foo)',
    '1 * parseFloat(foo)',
    'Number(foo) * 1',
    'Number(foo) - 0',
    'parseInt(foo) * 1',
    'parseFloat(foo) * 1',
    '- -Number(foo)',
    '1 * 1234 * 678 * Number(foo)',
    '1 * 1234 * 678 * parseInt(foo)',
    '(1 - 0) * parseInt(foo)',
    '1234 * 1 * 678 * Number(foo)',
    '1234 * 1 * Number(foo) * Number(bar)',
    '1234 * 1 * Number(foo) * parseInt(bar)',
    '1234 * 1 * Number(foo) * parseFloat(bar)',
    '1234 * 1 * parseInt(foo) * parseFloat(bar)',
    '1234 * 1 * parseInt(foo) * Number(bar)',
    '1234 * 1 * parseFloat(foo) * Number(bar)',
    '1234 * Number(foo) * 1 * Number(bar)',
    '1234 * parseInt(foo) * 1 * Number(bar)',
    '1234 * parseFloat(foo) * 1 * parseInt(bar)',
    '1234 * parseFloat(foo) * 1 * Number(bar)',
    '(- -1234) * (parseFloat(foo) - 0) * (Number(bar) - 0)',
    '1234*foo*1',
    '1234*1*foo',
    '1234*bar*1*foo',
    '1234*1*foo*bar',
    '1234*1*foo*Number(bar)',
    '1234*1*Number(foo)*bar',
    '1234*1*parseInt(foo)*bar',
    '0 + foo',
    '~foo.bar()',
    "foo + 'bar'",
    'foo + `${bar}`',

    { code: '!!foo', options: { boolean: false } },
    { code: '~foo.indexOf(1)', options: { boolean: false } },
    { code: '+foo', options: { number: false } },
    { code: '-(-foo)', options: { number: false } },
    { code: 'foo - 0', options: { number: false } },
    { code: '1*foo', options: { number: false } },
    { code: '""+foo', options: { string: false } },
    { code: 'foo += ""', options: { string: false } },
    { code: 'var a = !!foo', options: { boolean: true, allow: ['!!'] } },
    {
      code: 'var a = ~foo.indexOf(1)',
      options: { boolean: true, allow: ['~'] },
    },
    { code: 'var a = ~foo', options: { boolean: true } },
    { code: 'var a = 1 * foo', options: { boolean: true, allow: ['*'] } },
    { code: '- -foo', options: { number: true, allow: ['- -'] } },
    { code: 'foo - 0', options: { number: true, allow: ['-'] } },
    { code: 'var a = +foo', options: { boolean: true, allow: ['+'] } },
    {
      code: 'var a = "" + foo',
      options: { boolean: true, string: true, allow: ['+'] },
    },

    // https://github.com/eslint/eslint/issues/7057
    "'' + 'foo'",
    "`` + 'foo'",
    "'' + `${foo}`",
    "'foo' + ''",
    "'foo' + ``",
    '`${foo}` + ""',
    "foo += 'bar'",
    'foo += `${bar}`',
    {
      code: '`a${foo}`',
      options: { disallowTemplateShorthand: true },
    },
    {
      code: '`${foo}b`',
      options: { disallowTemplateShorthand: true },
    },
    {
      code: '`${foo}${bar}`',
      options: { disallowTemplateShorthand: true },
    },
    {
      code: 'tag`${foo}`',
      options: { disallowTemplateShorthand: true },
    },
    '`${foo}`',
    {
      code: '`${foo}`',
      options: { disallowTemplateShorthand: false },
    },
    '+42',

    // https://github.com/eslint/eslint/issues/14623
    "'' + String(foo)",
    "String(foo) + ''",
    '`` + String(foo)',
    'String(foo) + ``',
    {
      code: "`${'foo'}`",
      options: { disallowTemplateShorthand: true },
    },
    {
      code: '`${`foo`}`',
      options: { disallowTemplateShorthand: true },
    },
    {
      code: '`${String(foo)}`',
      options: { disallowTemplateShorthand: true },
    },

    // https://github.com/eslint/eslint/issues/16373
    'console.log(Math.PI * 1/4)',
    'a * 1 / 2',
    'a * 1 / b',

    // Parenthesised callee on Number/String — parens are transparent in ESLint,
    // so these are treated as already-coerced and NOT flagged.
    '+(Number)(foo)',
    '- -(Number)(foo)',
    '(Number)(foo) * 1',
    '(Number)(foo) - 0',
    "'' + (String)(foo)",
    "(String)(foo) + ''",
    '`` + (String)(foo)',
    '(String)(foo) + ``',
    {
      code: '`${(String)(foo)}`',
      options: { disallowTemplateShorthand: true },
    },
    '+((Number))(foo)',
  ],
  invalid: [
    {
      code: '!!foo',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '!!(foo + bar)',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '!!(foo + bar); var Boolean = null;',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '~foo.indexOf(1)',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '~foo.bar.indexOf(2)',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '+foo',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '-(-foo)',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '+foo.bar',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '1*foo',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'foo*1',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '1*foo.bar',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'foo.bar-0',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '""+foo',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '``+foo',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'foo+""',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'foo+``',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '""+foo.bar',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '``+foo.bar',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'foo.bar+""',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'foo.bar+``',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '`${foo}`',
      options: { disallowTemplateShorthand: true },
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'foo += ""',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'foo += ``',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'var a = !!foo',
      options: { boolean: true, allow: ['~'] },
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'var a = ~foo.indexOf(1)',
      options: { boolean: true, allow: ['!!'] },
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'var a = 1 * foo',
      options: { boolean: true, allow: ['+'] },
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'var a = +foo',
      options: { boolean: true, allow: ['*'] },
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'var a = "" + foo',
      options: { boolean: true, allow: ['*'] },
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'var a = `` + foo',
      options: { boolean: true, allow: ['*'] },
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'typeof+foo',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'typeof +foo',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: "let x ='' + 1n;",
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '~foo?.indexOf(1)',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '~(foo?.indexOf)(1)',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: "~foo[('indexOf')](1)",
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '~foo[(`lastIndexOf`)](1)',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '1 * a / 2',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: '(a * 1) / 2',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'a * 1 / (b * 1)',
      errors: [{ messageId: 'implicitCoercion' }],
    },
    {
      code: 'a * 1 + 2',
      errors: [{ messageId: 'implicitCoercion' }],
    },
  ],
});
