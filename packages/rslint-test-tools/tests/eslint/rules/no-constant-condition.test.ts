// Every case from ESLint v10.8.1 (156 valid, 151 invalid).
// Messages and complete ranges are additionally pinned from that version;
// upstream inputs, options, and message IDs are unchanged.
// This rule provides neither autofixes nor suggestions; the shared tester
// asserts their absence for every diagnostic.
// https://github.com/eslint/eslint/blob/v10.8.1/tests/lib/rules/no-constant-condition.js
import assert from 'node:assert/strict';
import { mkdtemp, rm } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { test } from 'rstack/test';
import { lint } from '@rslint/core/internal';
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '',
  languageOptions: { ecmaVersion: 2021 },
});

const cases: { valid: (string | ValidTestCase)[]; invalid: TestCase[] } = {
  valid: [
    'if(a);',
    'if(a == 0);',
    'if(a = f());',
    'if(a += 1);',
    'if(a |= 1);',
    'if(a |= true);',
    'if(a |= false);',
    'if(a &= 1);',
    'if(a &= true);',
    'if(a &= false);',
    'if(a >>= 1);',
    'if(a >>= true);',
    'if(a >>= false);',
    'if(a >>>= 1);',
    'if(a ??= 1);',
    'if(a ??= true);',
    'if(a ??= false);',
    'if(a ||= b);',
    'if(a ||= false);',
    'if(a ||= 0);',
    'if(a ||= void 0);',
    'if(+(a ||= 1));',
    'if(f(a ||= true));',
    'if((a ||= 1) + 2);',
    'if(1 + (a ||= true));',
    "if(a ||= '' || false);",
    'if(a ||= void 0 || null);',
    'if((a ||= false) || b);',
    'if(a || (b ||= false));',
    'if((a ||= true) && b);',
    'if(a && (b ||= true));',
    'if(a &&= b);',
    'if(a &&= true);',
    'if(a &&= 1);',
    "if(a &&= 'foo');",
    "if((a &&= '') + false);",
    "if('' + (a &&= null));",
    'if(a &&= 1 && 2);',
    'if((a &&= true) && b);',
    'if(a && (b &&= true));',
    'if((a &&= false) || b);',
    'if(a || (b &&= false));',
    'if(a ||= b ||= false);',
    'if(a &&= b &&= true);',
    'if(a ||= b &&= false);',
    'if(a ||= b &&= true);',
    'if(a &&= b ||= false);',
    'if(a &&= b ||= true);',
    'if(1, a);',
    "if ('every' in []);",
    'if (`\\\n${a}`) {}',
    'if (`${a}`);',
    'if (`${foo()}`);',
    "if (`${a === 'b' && b==='a'}`);",
    "if (`foo${a}` === 'fooa');",
    'if (tag`a`);',
    'if (tag`${a}`);',
    'if (+(a || true));',
    'if (-(a || true));',
    'if (~(a || 1));',
    'if (+(a && 0) === +(b && 0));',
    'while(~!a);',
    'while(a = b);',
    'while(`${a}`);',
    'for(;x < 10;);',
    'for(;;);',
    'for(;`${a}`;);',
    'do{ }while(x)',
    'q > 0 ? 1 : 2;',
    '`${a}` === a ? 1 : 2',
    '`foo${a}` === a ? 1 : 2',
    'tag`a` === a ? 1 : 2',
    'tag`${a}` === a ? 1 : 2',
    'while(x += 3) {}',
    'while(tag`a`) {}',
    'while(tag`${a}`) {}',
    'while(`\\\n${a}`) {}',
    "if(typeof x === 'undefined'){}",
    "if(`${typeof x}` === 'undefined'){}",
    "if(a === 'str' && typeof b){}",
    'typeof a == typeof b',
    "typeof 'a' === 'string'|| typeof b === 'string'",
    "`${typeof 'a'}` === 'string'|| `${typeof b}` === 'string'",
    'if (void a || a);',
    'if (a || void a);',
    "if(xyz === 'str1' && abc==='str2'){}",
    "if(xyz === 'str1' || abc==='str2'){}",
    "if(xyz === 'str1' || abc==='str2' && pqr === 5){}",
    "if(typeof abc === 'string' && abc==='str2'){}",
    "if(false || abc==='str'){}",
    "if(true && abc==='str'){}",
    "if(typeof 'str' && abc==='str'){}",
    "if(abc==='str' || false || def ==='str'){}",
    "if(true && abc==='str' || def ==='str'){}",
    "if(true && typeof abc==='string'){}",
    "if('str1' && a){}",
    "if(a && 'str'){}",
    "if ((foo || true) === 'baz') {}",
    "if ((foo || 'bar') === 'baz') {}",
    "if ((foo || 'bar') !== 'baz') {}",
    "if ((foo || 'bar') == 'baz') {}",
    "if ((foo || 'bar') != 'baz') {}",
    'if ((foo || 233) > 666) {}',
    'if ((foo || 233) < 666) {}',
    'if ((foo || 233) >= 666) {}',
    'if ((foo || 233) <= 666) {}',
    "if ((key || 'k') in obj) {}",
    'if ((foo || {}) instanceof obj) {}',
    "if ((foo || 'bar' || 'bar') === 'bar');",
    {
      code: "if ((foo || 1n) === 'baz') {}",
      languageOptions: {
        ecmaVersion: 11,
      },
    },
    {
      code: 'if (a && 0n || b);',
      languageOptions: {
        ecmaVersion: 11,
      },
    },
    {
      code: 'if(1n && a){};',
      languageOptions: {
        ecmaVersion: 11,
      },
    },
    "if ('' + [y] === '' + [ty]) {}",
    "if ('a' === '' + [ty]) {}",
    "if ('' + [y, m, d] === 'a') {}",
    "if ('' + [y, 'm'] === '' + [ty, 'tm']) {}",
    "if ('' + [y, 'm'] === '' + ['ty']) {}",
    'if ([,] in\n\n($2))\n ;\nelse\n ;',
    "if ([...x]+'' === 'y'){}",
    {
      code: 'while(true);',
      options: [
        {
          checkLoops: false,
        },
      ],
    },
    {
      code: 'for(;true;);',
      options: [
        {
          checkLoops: false,
        },
      ],
    },
    {
      code: 'do{}while(true)',
      options: [
        {
          checkLoops: false,
        },
      ],
    },
    {
      code: 'while(true);',
      options: [
        {
          checkLoops: 'none',
        },
      ],
    },
    {
      code: 'for(;true;);',
      options: [
        {
          checkLoops: 'none',
        },
      ],
    },
    {
      code: 'do{}while(true)',
      options: [
        {
          checkLoops: 'none',
        },
      ],
    },
    {
      code: 'while(true);',
      options: [
        {
          checkLoops: 'allExceptWhileTrue',
        },
      ],
    },
    'while(true);',
    {
      code: 'while(a == b);',
      options: [
        {
          checkLoops: 'all',
        },
      ],
    },
    {
      code: 'do{ }while(x);',
      options: [
        {
          checkLoops: 'all',
        },
      ],
    },
    {
      code: 'for (let x = 0; x <= 10; x++) {};',
      options: [
        {
          checkLoops: 'all',
        },
      ],
    },
    "function* foo(){while(true){yield 'foo';}}",
    "function* foo(){for(;true;){yield 'foo';}}",
    "function* foo(){do{yield 'foo';}while(true)}",
    'function* foo(){while (true) { while(true) {yield;}}}',
    'function* foo() {for (; yield; ) {}}',
    'function* foo() {for (; ; yield) {}}',
    'function* foo() {while (true) {function* foo() {yield;}yield;}}',
    'function* foo() { for (let x = yield; x < 10; x++) {yield;}yield;}',
    'function* foo() { for (let x = yield; ; x++) { yield; }}',
    'if (new Number(x) + 1 === 2) {}',
    'if([a]==[b]) {}',
    'if (+[...a]) {}',
    'if (+[...[...a]]) {}',
    'if (`${[...a]}`) {}',
    'if (`${[a]}`) {}',
    'if (+[a]) {}',
    'if (0 - [a]) {}',
    'if (1 * [a]) {}',
    'if (Boolean(a)) {}',
    'if (Boolean(...args)) {}',
    'if (foo.Boolean(1)) {}',
    'function foo(Boolean) { if (Boolean(1)) {} }',
    'const Boolean = () => {}; if (Boolean(1)) {}',
    {
      code: 'if (Boolean()) {}',
      languageOptions: {
        globals: {
          Boolean: 'off',
        },
      },
    },
    "const undefined = 'lol'; if (undefined) {}",
    {
      code: 'if (undefined) {}',
      languageOptions: {
        globals: {
          undefined: 'off',
        },
      },
    },
  ],
  invalid: [
    {
      code: 'for(;true;);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'for(;``;);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'for(;`foo`;);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'for(;`foo${bar}`;);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'do{}while(true)',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: "do{}while('1')",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'do{}while(0)',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'do{}while(t = -2)',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'do{}while(``)',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'do{}while(`foo`)',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'do{}while(`foo${bar}`)',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'true ? 1 : 2;',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: '1 ? 1 : 2;',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 2,
        },
      ],
    },
    {
      code: 'q = 0 ? 1 : 2;',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: '(q = 0) ? 1 : 2;',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: '`` ? 1 : 2;',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: '`foo` ? 1 : 2;',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: '`foo${bar}` ? 1 : 2;',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'if(-2);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'if(true);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'if(1);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: 'if({});',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'if(0 < 1);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'if(0 || 1);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'if(a, 1);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'if(`foo`);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'if(``);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'if(`\\\n`);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 2,
          endColumn: 2,
        },
      ],
    },
    {
      code: "if(`${'bar'}`);",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: "if(`${'bar' + `foo`}`);",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'if(`foo${false || true}`);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'if(`foo${0 || 1}`);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'if(`foo${bar}`);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'if(`${bar}foo`);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'if(!(true || a));',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'if(!(a && void b && c));',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'if(0 || !(a && null));',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'if(1 + !(a || true));',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'if(!(null && a) > 1);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'if(+(!(a && 0)));',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: "if(!typeof a === 'string');",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: "if(-('foo' || a));",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'if(+(void a && b) === ~(1 || c));',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'if(a ||= true);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'if(a ||= 5);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: "if(a ||= 'foo' || b);",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'if(a ||= b || /regex/);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'if(a ||= b ||= true);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'if(a ||= b ||= c || 1);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'if(!(a ||= true));',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "if(!(a ||= 'foo') === true);",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: "if(!(a ||= 'foo') === false);",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'if(a || (b ||= true));',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'if((a ||= 1) || b);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'if((a ||= true) && true);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'if(true && (a ||= true));',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'if(a &&= false);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'if(a &&= null);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'if(a &&= void b);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'if(a &&= 0 && b);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: "if(a &&= b && '');",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'if(a &&= b &&= false);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'if(a &&= b &&= c && false);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'if(!(a &&= false));',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'if(!(a &&= 0) + 1);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'if(a && (b &&= false));',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'if((a &&= null) && b);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'if(false || (a &&= false));',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'if((a &&= false) || false);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'while([]);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'while(~!0);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'while(x = 1);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'while(function(){});',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'while(true);',
      options: [
        {
          checkLoops: 'all',
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'while(1);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'while(() => {});',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'while(`foo`);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'while(``);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: "while(`${'foo'}`);",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "while(`${'foo' + 'bar'}`);",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'if(typeof x){}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: "if(typeof 'abc' === 'string'){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'if(a = typeof b){}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'if(a, typeof b){}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: "if(typeof 'a' == 'string' || typeof 'b' == 'string'){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: 'while(typeof x){}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'if(1 || void x);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'if(void x);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'if(y = void x);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'if(x, void x);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'if(void x === void y);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'if(void x && a);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'if(a && void x);',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: "if(false && abc==='str'){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: "if(true || abc==='str'){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "if(1 || abc==='str'){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: "if(abc==='str' || true){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "if(abc==='str' || true || def ==='str'){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'if(false || true){}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "if(typeof abc==='str' || true){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: "if('str' || a){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: "if('str' || abc==='str'){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: "if('str1' || 'str2'){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: "if('str1' && 'str2'){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: "if(abc==='str' || 'str'){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: "if(a || 'str'){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'while(x = 1);',
      options: [
        {
          checkLoops: 'all',
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'do{ }while(x = 1)',
      options: [
        {
          checkLoops: 'all',
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'for (;true;) {};',
      options: [
        {
          checkLoops: 'all',
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: "function* foo(){while(true){} yield 'foo';}",
      options: [
        {
          checkLoops: 'all',
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: "function* foo(){while(true){} yield 'foo';}",
      options: [
        {
          checkLoops: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: "function* foo(){while(true){if (true) {yield 'foo';}}}",
      options: [
        {
          checkLoops: 'all',
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: "function* foo(){while(true){if (true) {yield 'foo';}}}",
      options: [
        {
          checkLoops: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: "function* foo(){while(true){yield 'foo';} while(true) {}}",
      options: [
        {
          checkLoops: 'all',
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: "function* foo(){while(true){yield 'foo';} while(true) {}}",
      options: [
        {
          checkLoops: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: "var a = function* foo(){while(true){} yield 'foo';}",
      options: [
        {
          checkLoops: 'all',
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: "var a = function* foo(){while(true){} yield 'foo';}",
      options: [
        {
          checkLoops: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'while (true) { function* foo() {yield;}}',
      options: [
        {
          checkLoops: 'all',
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'while (true) { function* foo() {yield;}}',
      options: [
        {
          checkLoops: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: "function* foo(){if (true) {yield 'foo';}}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'function* foo() {for (let foo = yield; true;) {}}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      code: 'function* foo() {for (foo = yield; true;) {}}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: 'function foo() {while (true) {function* bar() {while (true) {yield;}}}}',
      options: [
        {
          checkLoops: 'all',
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'function foo() {while (true) {const bar = function*() {while (true) {yield;}}}}',
      options: [
        {
          checkLoops: 'all',
        },
      ],
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'function* foo() { for (let foo = 1 + 2 + 3 + (yield); true; baz) {}}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 55,
          endLine: 1,
          endColumn: 59,
        },
      ],
    },
    {
      code: 'if([a]) {}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'if([]) {}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: "if(''+['a']) {}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: "if(''+[]) {}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'if(+1) {}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: "if ([,] + ''){}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'if(/foo/ui);',
      languageOptions: {
        ecmaVersion: 11,
      },
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'if(0n);',
      languageOptions: {
        ecmaVersion: 11,
      },
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'if(0b0n);',
      languageOptions: {
        ecmaVersion: 11,
      },
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'if(0o0n);',
      languageOptions: {
        ecmaVersion: 11,
      },
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'if(0x0n);',
      languageOptions: {
        ecmaVersion: 11,
      },
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'if(0b1n);',
      languageOptions: {
        ecmaVersion: 11,
      },
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'if(0o1n);',
      languageOptions: {
        ecmaVersion: 11,
      },
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'if(0x1n);',
      languageOptions: {
        ecmaVersion: 11,
      },
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'if(0x1n || foo);',
      languageOptions: {
        ecmaVersion: 11,
      },
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'if(class {}) {}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'if(new Foo()) {}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'if(new Boolean(foo)) {}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'if(new String(foo)) {}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'if(new Number(foo)) {}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: "if(`${[...['a']]}`) {}",
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'if (undefined) {}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'if (Boolean(1)) {}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'if (Boolean()) {}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'if (Boolean([a])) {}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'if (Boolean(1)) { function Boolean() {}}',
      errors: [
        {
          messageId: 'unexpected',
          message: 'Unexpected constant condition.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
  ],
};

ruleTester.run('no-constant-condition', cases);

// Exercise the actual fix pipeline too: this diagnostic-only rule must keep
// every upstream invalid input unchanged, with the same complete diagnostics.
test('no-constant-condition preserves every upstream input when autofixing', async () => {
  const root = await mkdtemp(path.join(os.tmpdir(), 'rslint-constant-fix-'));
  try {
    for (const item of cases.invalid) {
      const result = await lint({
        configDirectory: root,
        workingDirectory: root,
        config: [
          {
            languageOptions: {
              ecmaVersion: 2021,
              ...item.languageOptions,
              parserOptions: { projectService: false },
            },
            rules: {
              'no-constant-condition': [
                'error',
                ...(typeof item.options === 'function'
                  ? item.options(root)
                  : (item.options ?? [])),
              ],
            },
          },
        ],
        fileContents: { [path.join(root, 'input.js')]: item.code },
        fix: true,
      });
      assert.equal(result.fileCount, 1);
      assert.equal(result.ruleCount, 1);
      assert.deepEqual(Object.keys(result.output ?? {}), [], item.code);
      assert.deepEqual(
        result.diagnostics.map((diagnostic) => ({
          messageId: diagnostic.messageId,
          message: diagnostic.message,
          line: diagnostic.range.start.line,
          column: diagnostic.range.start.column,
          endLine: diagnostic.range.end.line,
          endColumn: diagnostic.range.end.column,
        })),
        item.errors,
        item.code,
      );
      for (const diagnostic of result.diagnostics) {
        assert.equal(diagnostic.fixes, undefined);
        assert.deepEqual(diagnostic.suggestions ?? [], []);
      }
    }
  } finally {
    await rm(root, { recursive: true, force: true });
  }
});
