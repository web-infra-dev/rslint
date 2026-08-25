import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const expectedError = { messageId: 'sortVars' };
const ignoreCase = { ignoreCase: true };

ruleTester.run('sort-vars', {
  valid: [
    "var a=10, b=4, c='abc'",
    'var a, b, c, d',
    'var b; var a; var d;',
    'var _a, a',
    'var A, a',
    'var A, b',
    { code: 'var a, A;', options: ignoreCase },
    { code: 'var A, a;', options: ignoreCase },
    { code: 'var a, B, c;', options: ignoreCase },
    { code: 'var A, b, C;', options: ignoreCase },
    {
      code: 'var {a, b, c} = x;',
      options: ignoreCase,
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: 'var {A, b, C} = x;',
      options: ignoreCase,
      languageOptions: { ecmaVersion: 2015 },
    },
    { code: 'var test = [1,2,3];', languageOptions: { ecmaVersion: 2015 } },
    { code: 'var {a,b} = [1,2];', languageOptions: { ecmaVersion: 2015 } },
    {
      code: 'var [a, B, c] = [1, 2, 3];',
      options: ignoreCase,
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: 'var [A, B, c] = [1, 2, 3];',
      options: ignoreCase,
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: 'var [A, b, C] = [1, 2, 3];',
      options: ignoreCase,
      languageOptions: { ecmaVersion: 2015 },
    },
    { code: 'let {a, b, c} = x;', languageOptions: { ecmaVersion: 2015 } },
    {
      code: 'let [a, b, c] = [1, 2, 3];',
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: 'const {a, b, c} = {a: 1, b: true, c: "Moo"};',
      options: ignoreCase,
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: 'const [a, b, c] = [1, true, "Moo"];',
      options: ignoreCase,
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: 'const [c, a, b] = [1, true, "Moo"];',
      options: ignoreCase,
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: 'var {a, x: {b, c}} = {};',
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: 'var {c, x: {a, c}} = {};',
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: 'var {a, x: [b, c]} = {};',
      languageOptions: { ecmaVersion: 2015 },
    },
    { code: 'var [a, {b, c}] = {};', languageOptions: { ecmaVersion: 2015 } },
    {
      code: 'var [a, {x: {b, c}}] = {};',
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: 'var a = 42, {b, c } = {};',
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: 'var b = 42, {a, c } = {};',
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: 'var [b, {x: {a, c}}] = {};',
      languageOptions: { ecmaVersion: 2015 },
    },
    { code: 'var [b, d, a, c] = {};', languageOptions: { ecmaVersion: 2015 } },
    { code: 'var e, [a, c, d] = {};', languageOptions: { ecmaVersion: 2015 } },
    {
      code: 'var a, [E, c, D] = [];',
      options: ignoreCase,
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: 'var a, f, [e, c, d] = [1,2,3];',
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: [
        'export default class {',
        '    render () {',
        '        let {',
        '            b',
        '        } = this,',
        '            a,',
        '            c;',
        '    }',
        '}',
      ].join('\n'),
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: 'var {} = 1, a',
      options: ignoreCase,
      languageOptions: { ecmaVersion: 2015 },
    },
  ],
  invalid: [
    { code: 'var b, a', errors: [expectedError] },
    { code: 'var b , a', errors: [expectedError] },
    { code: ['var b,', '    a;'].join('\n'), errors: [expectedError] },
    { code: 'var b=10, a=20;', errors: [expectedError] },
    { code: 'var b=10, a=20, c=30;', errors: [expectedError] },
    { code: 'var all=10, a = 1', errors: [expectedError] },
    { code: 'var b, c, a, d', errors: [expectedError] },
    { code: 'var c, d, a, b', errors: [expectedError, expectedError] },
    { code: 'var a, A;', errors: [expectedError] },
    { code: 'var a, B;', errors: [expectedError] },
    { code: 'var a, B, c;', errors: [expectedError] },
    { code: 'var B, a;', options: ignoreCase, errors: [expectedError] },
    { code: 'var B, A, c;', options: ignoreCase, errors: [expectedError] },
    {
      code: 'var d, a, [b, c] = {};',
      options: ignoreCase,
      languageOptions: { ecmaVersion: 2015 },
      errors: [expectedError],
    },
    {
      code: 'var d, a, [b, {x: {c, e}}] = {};',
      options: ignoreCase,
      languageOptions: { ecmaVersion: 2015 },
      errors: [expectedError],
    },
    {
      code: 'var {} = 1, b, a',
      options: ignoreCase,
      languageOptions: { ecmaVersion: 2015 },
      errors: [expectedError],
    },
    { code: 'var b=10, a=f();', errors: [expectedError] },
    { code: 'var b=10, a=b;', errors: [expectedError] },
    {
      code: 'var b = 0, a = `${b}`;',
      languageOptions: { ecmaVersion: 2015 },
      errors: [expectedError],
    },
    {
      code: 'var b = 0, a = `${f()}`',
      languageOptions: { ecmaVersion: 2015 },
      errors: [expectedError],
    },
    { code: 'var b = 0, c = b, a;', errors: [expectedError] },
    { code: 'var b = 0, c = 0, a = b + c;', errors: [expectedError] },
    { code: 'var b = f(), c, d, a;', errors: [expectedError] },
    {
      code: 'var b = `${f()}`, c, d, a;',
      languageOptions: { ecmaVersion: 2015 },
      errors: [expectedError],
    },
    { code: 'var c, a = b = 0', errors: [expectedError] },
  ],
});
