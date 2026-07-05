import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unassigned-vars', {
  valid: [
    'let message = "hello"; console.log(message);',
    'let user; user = getUser(); console.log(user.name);',
    'let count; count = 1; count++;',
    'let temp;',
    'let error; if (somethingWentWrong) { error = "Something went wrong"; } console.log(error);',
    'let item; for (item of items) { process(item); }',
    'let config; function setup() { config = { debug: true }; } setup(); console.log(config);',
    'let one = undefined; if (one === two) { }',

    'let x;',
    'var x;',
    'const x = undefined; log(x);',
    'let y = undefined; log(y);',
    'var y = undefined; log(y);',
    'let a = x, b = y; log(a, b);',
    'var a = x, b = y; log(a, b);',
    'const foo = (two) => { let one; if (one !== two) one = two; }',

    'let z: number | undefined = undefined; log(z);',
    'declare let c: string | undefined; log(c);',
    `
const foo = (two: string): void => {
  let one: string | undefined;
  if (one !== two) {
    one = two;
  }
}`,
    `
declare module 'module' {
  import type { T } from 'module';
  let x: T;
  export = x;
}`,
  ],
  invalid: [
    {
      code: 'let status; if (status === "ready") { console.log("Ready!"); }',
      errors: [{ messageId: 'unassigned', line: 1, column: 5 }],
    },
    {
      code: 'let value: number | undefined; console.log(value);',
      errors: [{ messageId: 'unassigned', line: 1, column: 5 }],
    },
    {
      code: 'let x; let a = x, b; log(x, a, b);',
      errors: [
        { messageId: 'unassigned', line: 1, column: 5 },
        { messageId: 'unassigned', line: 1, column: 19 },
      ],
    },
    {
      code: 'const foo = (two) => { let one; if (one === two) {} }',
      errors: [{ messageId: 'unassigned', line: 1, column: 28 }],
    },
    {
      code: 'let user; greet(user);',
      errors: [{ messageId: 'unassigned', line: 1, column: 5 }],
    },
    {
      code: "function test() { let error; return error || 'Unknown error'; }",
      errors: [{ messageId: 'unassigned', line: 1, column: 23 }],
    },
    {
      code: 'let options; const { debug } = options || {};',
      errors: [{ messageId: 'unassigned', line: 1, column: 5 }],
    },
    {
      code: 'let flag; while (!flag) { }',
      errors: [{ messageId: 'unassigned', line: 1, column: 5 }],
    },
    {
      code: 'let config; function init() { return config?.enabled; }',
      errors: [{ messageId: 'unassigned', line: 1, column: 5 }],
    },
    {
      code: 'let x: number; log(x);',
      errors: [{ messageId: 'unassigned', line: 1, column: 5 }],
    },
    {
      code: 'let x: number | undefined; log(x);',
      errors: [{ messageId: 'unassigned', line: 1, column: 5 }],
    },
    {
      code: 'const foo = (two: string): void => { let one: string | undefined; if (one === two) {} }',
      errors: [{ messageId: 'unassigned', line: 1, column: 42 }],
    },
    {
      code: `declare module 'module' {
  let x: string;
}
let y: string;
console.log(y);`,
      errors: [{ messageId: 'unassigned', line: 4, column: 5 }],
    },
  ],
});
