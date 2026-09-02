import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-new-func', {
  valid: [
    'var a = new _function("b", "c", "return b+c");',
    'var a = _function("b", "c", "return b+c");',
    {
      code: 'class Function {}; new Function()',
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      code: 'const fn = () => { class Function {}; new Function() }',
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    'function Function() {}; Function()',
    'var fn = function () { function Function() {}; Function() }',
    'var x = function Function() { Function(); }',
    'call(Function)',
    'new Class(Function)',
    'foo[Function]()',
    'foo(Function.bind)',
    'Function.toString()',
    'Function[call]()',
  ],
  invalid: [
    {
      code: 'var a = new Function("b", "c", "return b+c");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'var a = Function("b", "c", "return b+c");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'var a = Function.call(null, "b", "c", "return b+c");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'var a = Function.apply(null, ["b", "c", "return b+c"]);',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'var a = Function.bind(null, "b", "c", "return b+c")();',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'var a = Function.bind(null, "b", "c", "return b+c");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'var a = Function["call"](null, "b", "c", "return b+c");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'var a = (Function?.call)(null, "b", "c", "return b+c");',
      languageOptions: { ecmaVersion: 2021 },
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: "const fn = () => { class Function {} }; new Function('', '')",
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: "var fn = function () { function Function() {} }; Function('', '')",
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
  ],
});
