// Ported from ESLint v9.39.1. All upstream inputs and options are preserved.
// https://github.com/eslint/eslint/blob/v9.39.1/tests/lib/rules/array-callback-return.js
// Full messages, ranges and suggestion descriptions were verified with that version.
// The rule has suggestions only; every diagnostic must have no automatic fix.
import assert from 'node:assert/strict';
import { mkdtemp, realpath, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { lint } from '@rslint/core/internal';
import { test } from 'rstack/test';
import { parse } from 'espree';
import { applyEdits } from '../../../../rslint-rule-tester/src/edits.ts';
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';

const ruleTester = new RuleTester({ plugin: '' });

const cases: { valid: (string | ValidTestCase)[]; invalid: TestCase[] } = {
  valid: [
    'foo.every(function(){}())',
    'foo.every(function(){ return function() { return true; }; }())',
    'foo.every(function(){ return function() { return; }; })',
    'foo.forEach(bar || function(x) { var a=0; })',
    'foo.forEach(bar || function(x) { return a; })',
    'foo.forEach(function() {return function() { var a = 0;}}())',
    'foo.forEach(function(x) { var a=0; })',
    'foo.forEach(function(x) { return a;})',
    'foo.forEach(function(x) { return; })',
    'foo.forEach(function(x) { if (a === b) { return;} var a=0; })',
    'foo.forEach(function(x) { if (a === b) { return x;} var a=0; })',
    'foo.bar().forEach(function(x) { return; })',
    '["foo","bar","baz"].forEach(function(x) { return x; })',
    {
      code: 'foo.forEach(x => { var a=0; })',
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'foo.forEach(x => { if (a === b) { return;} var a=0; })',
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'foo.forEach(x => x)',
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'foo.forEach(val => y += val)',
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'foo.map(async function(){})',
      languageOptions: {
        ecmaVersion: 8,
      },
    },
    {
      code: 'foo.map(async () => {})',
      languageOptions: {
        ecmaVersion: 8,
      },
    },
    {
      code: 'foo.map(function* () {})',
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'Array.from(x, function() { return true; })',
      options: [
        {
          allowImplicit: false,
        },
      ],
    },
    {
      code: 'Int32Array.from(x, function() { return true; })',
      options: [
        {
          allowImplicit: false,
        },
      ],
    },
    'foo.every(function() { return true; })',
    'foo.filter(function() { return true; })',
    'foo.find(function() { return true; })',
    'foo.findIndex(function() { return true; })',
    'foo.findLast(function() { return true; })',
    'foo.findLastIndex(function() { return true; })',
    'foo.flatMap(function() { return true; })',
    'foo.forEach(function() { return; })',
    'foo.map(function() { return true; })',
    'foo.reduce(function() { return true; })',
    'foo.reduceRight(function() { return true; })',
    'foo.some(function() { return true; })',
    'foo.sort(function() { return 0; })',
    'foo.toSorted(function() { return 0; })',
    {
      code: 'foo.every(() => { return true; })',
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    'foo.every(function() { if (a) return true; else return false; })',
    'foo.every(function() { switch (a) { case 0: bar(); default: return true; } })',
    'foo.every(function() { try { bar(); return true; } catch (err) { return false; } })',
    'foo.every(function() { try { bar(); } finally { return true; } })',
    {
      code: 'Array.from(x, function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'Int32Array.from(x, function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.every(function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.filter(function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.find(function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.findIndex(function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.findLast(function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.findLastIndex(function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.flatMap(function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.forEach(function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.map(function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.reduce(function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.reduceRight(function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.some(function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.sort(function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.toSorted(function() { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.every(() => { return; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'foo.every(function() { if (a) return; else return a; })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.every(function() { switch (a) { case 0: bar(); default: return; } })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.every(function() { try { bar(); return; } catch (err) { return; } })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.every(function() { try { bar(); } finally { return; } })',
      options: [
        {
          allowImplicit: true,
        },
      ],
    },
    {
      code: 'foo.forEach(function(x) { return; })',
      options: [
        {
          checkForEach: true,
        },
      ],
    },
    {
      code: 'foo.forEach(function(x) { var a=0; })',
      options: [
        {
          checkForEach: true,
        },
      ],
    },
    {
      code: 'foo.forEach(function(x) { if (a === b) { return;} var a=0; })',
      options: [
        {
          checkForEach: true,
        },
      ],
    },
    {
      code: 'foo.forEach(function() {return function() { if (a == b) { return; }}}())',
      options: [
        {
          checkForEach: true,
        },
      ],
    },
    {
      code: 'foo.forEach(x => { var a=0; })',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'foo.forEach(x => { if (a === b) { return;} var a=0; })',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'foo.forEach(x => { x })',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'foo.forEach(bar || function(x) { return; })',
      options: [
        {
          checkForEach: true,
        },
      ],
    },
    {
      code: 'Array.from(x, function() { return true; })',
      options: [
        {
          checkForEach: true,
        },
      ],
    },
    {
      code: 'Int32Array.from(x, function() { return true; })',
      options: [
        {
          checkForEach: true,
        },
      ],
    },
    {
      code: 'foo.every(() => { return true; })',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'foo.every(function() { if (a) return 1; else return a; })',
      options: [
        {
          checkForEach: true,
        },
      ],
    },
    {
      code: 'foo.every(function() { switch (a) { case 0: return bar(); default: return a; } })',
      options: [
        {
          checkForEach: true,
        },
      ],
    },
    {
      code: 'foo.every(function() { try { bar(); return 1; } catch (err) { return err; } })',
      options: [
        {
          checkForEach: true,
        },
      ],
    },
    {
      code: 'foo.every(function() { try { bar(); } finally { return 1; } })',
      options: [
        {
          checkForEach: true,
        },
      ],
    },
    {
      code: 'foo.every(function() { return; })',
      options: [
        {
          allowImplicit: true,
          checkForEach: true,
        },
      ],
    },
    {
      code: 'foo.forEach((x) => void x)',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'foo.forEach((x) => void bar(x))',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'foo.forEach(function (x) { return void bar(x); })',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'foo.forEach((x) => { return void bar(x); })',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'foo.forEach((x) => { if (a === b) { return void a; } bar(x) })',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    'Arrow.from(x, function() {})',
    'foo.abc(function() {})',
    'every(function() {})',
    'foo[every](function() {})',
    'var every = function() {}',
    {
      code: 'foo[`${every}`](function() {})',
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'foo.every(() => true)',
      languageOptions: {
        ecmaVersion: 6,
      },
    },
  ],
  invalid: [
    {
      code: 'Array.from(x, function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message: 'Array.from() expects a return value from function.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'Array.from(x, function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message: "Array.from() expects a return value from function 'foo'.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'Int32Array.from(x, function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message: 'Array.from() expects a return value from function.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'Int32Array.from(x, function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message: "Array.from() expects a return value from function 'foo'.",
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'foo.every(function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'foo.every(function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.every() expects a return value from function 'foo'.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'foo.filter(function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.filter() expects a return value from function.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'foo.filter(function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.filter() expects a return value from function 'foo'.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'foo.find(function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.find() expects a return value from function.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'foo.find(function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.find() expects a return value from function 'foo'.",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'foo.findLast(function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.findLast() expects a return value from function.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'foo.findLast(function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.findLast() expects a return value from function 'foo'.",
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'foo.findIndex(function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.findIndex() expects a return value from function.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'foo.findIndex(function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.findIndex() expects a return value from function 'foo'.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'foo.findLastIndex(function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.findLastIndex() expects a return value from function.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'foo.findLastIndex(function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.findLastIndex() expects a return value from function 'foo'.",
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'foo.flatMap(function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.flatMap() expects a return value from function.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'foo.flatMap(function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.flatMap() expects a return value from function 'foo'.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'foo.map(function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.map() expects a return value from function.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'foo.map(function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.map() expects a return value from function 'foo'.",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'foo.reduce(function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.reduce() expects a return value from function.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'foo.reduce(function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.reduce() expects a return value from function 'foo'.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'foo.reduceRight(function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.reduceRight() expects a return value from function.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'foo.reduceRight(function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.reduceRight() expects a return value from function 'foo'.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'foo.some(function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.some() expects a return value from function.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'foo.some(function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.some() expects a return value from function 'foo'.",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'foo.sort(function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.sort() expects a return value from function.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'foo.sort(function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.sort() expects a return value from function 'foo'.",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'foo.toSorted(function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.toSorted() expects a return value from function.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'foo.toSorted(function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.toSorted() expects a return value from function 'foo'.",
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'foo.bar.baz.every(function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'foo.bar.baz.every(function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.every() expects a return value from function 'foo'.",
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'foo["every"](function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'foo["every"](function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.every() expects a return value from function 'foo'.",
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'foo[`every`](function() {})',
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'foo[`every`](function foo() {})',
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.every() expects a return value from function 'foo'.",
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'foo.every(() => {})',
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from arrow function.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'foo.every(function() { if (a) return true; })',
      errors: [
        {
          messageId: 'expectedAtEnd',
          message:
            'Array.prototype.every() expects a value to be returned at the end of function.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'foo.every(function cb() { if (a) return true; })',
      errors: [
        {
          messageId: 'expectedAtEnd',
          message:
            "Array.prototype.every() expects a value to be returned at the end of function 'cb'.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'foo.every(function() { switch (a) { case 0: break; default: return true; } })',
      errors: [
        {
          messageId: 'expectedAtEnd',
          message:
            'Array.prototype.every() expects a value to be returned at the end of function.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'foo.every(function foo() { switch (a) { case 0: break; default: return true; } })',
      errors: [
        {
          messageId: 'expectedAtEnd',
          message:
            "Array.prototype.every() expects a value to be returned at the end of function 'foo'.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'foo.every(function() { try { bar(); } catch (err) { return true; } })',
      errors: [
        {
          messageId: 'expectedAtEnd',
          message:
            'Array.prototype.every() expects a value to be returned at the end of function.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'foo.every(function foo() { try { bar(); } catch (err) { return true; } })',
      errors: [
        {
          messageId: 'expectedAtEnd',
          message:
            "Array.prototype.every() expects a value to be returned at the end of function 'foo'.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'foo.every(function() { return; })',
      errors: [
        {
          messageId: 'expectedReturnValue',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'foo.every(function foo() { return; })',
      errors: [
        {
          messageId: 'expectedReturnValue',
          message:
            "Array.prototype.every() expects a return value from function 'foo'.",
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'foo.every(function() { if (a) return; })',
      errors: [
        {
          messageId: 'expectedAtEnd',
          message:
            'Array.prototype.every() expects a value to be returned at the end of function.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
        },
        {
          messageId: 'expectedReturnValue',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: 'foo.every(function foo() { if (a) return; })',
      errors: [
        {
          messageId: 'expectedAtEnd',
          message:
            "Array.prototype.every() expects a value to be returned at the end of function 'foo'.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
        },
        {
          messageId: 'expectedReturnValue',
          message:
            "Array.prototype.every() expects a return value from function 'foo'.",
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: 'foo.every(function() { if (a) return; else return; })',
      errors: [
        {
          messageId: 'expectedReturnValue',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 38,
        },
        {
          messageId: 'expectedReturnValue',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: 'foo.every(function foo() { if (a) return; else return; })',
      errors: [
        {
          messageId: 'expectedReturnValue',
          message:
            "Array.prototype.every() expects a return value from function 'foo'.",
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 42,
        },
        {
          messageId: 'expectedReturnValue',
          message:
            "Array.prototype.every() expects a return value from function 'foo'.",
          line: 1,
          column: 48,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: 'foo.every(cb || function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'foo.every(cb || function foo() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.every() expects a return value from function 'foo'.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'foo.every(a ? function() {} : function() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 23,
        },
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'foo.every(a ? function foo() {} : function bar() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.every() expects a return value from function 'foo'.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 27,
        },
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.every() expects a return value from function 'bar'.",
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: 'foo.every(function(){ return function() {}; }())',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: 'foo.every(function(){ return function foo() {}; }())',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.every() expects a return value from function 'foo'.",
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: 'foo.every(() => {})',
      options: [
        {
          allowImplicit: false,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from arrow function.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'foo.every(() => {})',
      options: [
        {
          allowImplicit: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from arrow function.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'Array.from(x, function() {})',
      options: [
        {
          allowImplicit: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message: 'Array.from() expects a return value from function.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'foo.every(function() {})',
      options: [
        {
          allowImplicit: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'foo.filter(function foo() {})',
      options: [
        {
          allowImplicit: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.filter() expects a return value from function 'foo'.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'foo.find(function foo() {})',
      options: [
        {
          allowImplicit: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.find() expects a return value from function 'foo'.",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'foo.map(function() {})',
      options: [
        {
          allowImplicit: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.map() expects a return value from function.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'foo.reduce(function() {})',
      options: [
        {
          allowImplicit: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.reduce() expects a return value from function.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'foo.reduceRight(function() {})',
      options: [
        {
          allowImplicit: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.reduceRight() expects a return value from function.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'foo.bar.baz.every(function foo() {})',
      options: [
        {
          allowImplicit: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.every() expects a return value from function 'foo'.",
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'foo.every(cb || function() {})',
      options: [
        {
          allowImplicit: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '["foo","bar"].sort(function foo() {})',
      options: [
        {
          allowImplicit: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.sort() expects a return value from function 'foo'.",
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '["foo","bar"].toSorted(function foo() {})',
      options: [
        {
          allowImplicit: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.toSorted() expects a return value from function 'foo'.",
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: 'foo.forEach(x => x)',
      options: [
        {
          allowImplicit: true,
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'wrapBraces',
              desc: 'Wrap the expression in `{}`.',
              output: 'foo.forEach(x => {x})',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach(function(x) { if (a == b) {return x;}})',
      options: [
        {
          allowImplicit: true,
          checkForEach: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from function.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: 'foo.forEach(function bar(x) { return x;})',
      options: [
        {
          allowImplicit: true,
          checkForEach: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            "Array.prototype.forEach() expects no useless return value from function 'bar'.",
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: 'foo.forEach(x => x)',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'wrapBraces',
              desc: 'Wrap the expression in `{}`.',
              output: 'foo.forEach(x => {x})',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach(x => (x))',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'wrapBraces',
              desc: 'Wrap the expression in `{}`.',
              output: 'foo.forEach(x => {(x)})',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach(val => y += val)',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'wrapBraces',
              desc: 'Wrap the expression in `{}`.',
              output: 'foo.forEach(val => {y += val})',
            },
          ],
        },
      ],
    },
    {
      code: '["foo","bar"].forEach(x => ++x)',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'wrapBraces',
              desc: 'Wrap the expression in `{}`.',
              output: '["foo","bar"].forEach(x => {++x})',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.bar().forEach(x => x === y)',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'wrapBraces',
              desc: 'Wrap the expression in `{}`.',
              output: 'foo.bar().forEach(x => {x === y})',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach(function() {return function() { if (a == b) { return a; }}}())',
      options: [
        {
          checkForEach: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from function.',
          line: 1,
          column: 59,
          endLine: 1,
          endColumn: 68,
        },
      ],
    },
    {
      code: 'foo.forEach(function(x) { if (a == b) {return x;}})',
      options: [
        {
          checkForEach: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from function.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: 'foo.forEach(function(x) { if (a == b) {return undefined;}})',
      options: [
        {
          checkForEach: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from function.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: 'foo.forEach(function bar(x) { return x;})',
      options: [
        {
          checkForEach: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            "Array.prototype.forEach() expects no useless return value from function 'bar'.",
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: 'foo.bar().forEach(function bar(x) { return x;})',
      options: [
        {
          checkForEach: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            "Array.prototype.forEach() expects no useless return value from function 'bar'.",
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: '["foo","bar"].forEach(function bar(x) { return x;})',
      options: [
        {
          checkForEach: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            "Array.prototype.forEach() expects no useless return value from function 'bar'.",
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: 'foo.forEach((x) => { return x;})',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'Array.from(x, function() {})',
      options: [
        {
          checkForEach: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message: 'Array.from() expects a return value from function.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'foo.every(function() {})',
      options: [
        {
          checkForEach: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'foo.filter(function foo() {})',
      options: [
        {
          checkForEach: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.filter() expects a return value from function 'foo'.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'foo.filter(function foo() { return; })',
      options: [
        {
          checkForEach: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedReturnValue',
          message:
            "Array.prototype.filter() expects a return value from function 'foo'.",
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: 'foo.every(cb || function() {})',
      options: [
        {
          checkForEach: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.every() expects a return value from function.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'foo.forEach((x) => void x)',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'wrapBraces',
              desc: 'Wrap the expression in `{}`.',
              output: 'foo.forEach((x) => {void x})',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach((x) => void bar(x))',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'wrapBraces',
              desc: 'Wrap the expression in `{}`.',
              output: 'foo.forEach((x) => {void bar(x)})',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach((x) => { return void bar(x); })',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: 'foo.forEach((x) => { if (a === b) { return void a; } bar(x) })',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: 'foo.forEach(x => x)',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'wrapBraces',
              desc: 'Wrap the expression in `{}`.',
              output: 'foo.forEach(x => {x})',
            },
            {
              messageId: 'prependVoid',
              desc: 'Prepend `void` to the expression.',
              output: 'foo.forEach(x => void x)',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach(x => !x)',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'wrapBraces',
              desc: 'Wrap the expression in `{}`.',
              output: 'foo.forEach(x => {!x})',
            },
            {
              messageId: 'prependVoid',
              desc: 'Prepend `void` to the expression.',
              output: 'foo.forEach(x => void !x)',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach(x => (x))',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'wrapBraces',
              desc: 'Wrap the expression in `{}`.',
              output: 'foo.forEach(x => {(x)})',
            },
            {
              messageId: 'prependVoid',
              desc: 'Prepend `void` to the expression.',
              output: 'foo.forEach(x => void (x))',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach((x) => { return x; })',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 31,
          suggestions: [
            {
              messageId: 'prependVoid',
              desc: 'Prepend `void` to the expression.',
              output: 'foo.forEach((x) => { return void x; })',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach((x) => { return !x; })',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'prependVoid',
              desc: 'Prepend `void` to the expression.',
              output: 'foo.forEach((x) => { return void !x; })',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach((x) => { return(x); })',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'prependVoid',
              desc: 'Prepend `void` to the expression.',
              output: 'foo.forEach((x) => { return void (x); })',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach((x) => { return (x + 1); })',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'prependVoid',
              desc: 'Prepend `void` to the expression.',
              output: 'foo.forEach((x) => { return void (x + 1); })',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach((x) => { if (a === b) { return x; } })',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 46,
          suggestions: [
            {
              messageId: 'prependVoid',
              desc: 'Prepend `void` to the expression.',
              output: 'foo.forEach((x) => { if (a === b) { return void x; } })',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach((x) => { if (a === b) { return !x; } })',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 47,
          suggestions: [
            {
              messageId: 'prependVoid',
              desc: 'Prepend `void` to the expression.',
              output:
                'foo.forEach((x) => { if (a === b) { return void !x; } })',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach((x) => { if (a === b) { return (x + a); } })',
      options: [
        {
          checkForEach: true,
          allowVoid: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 52,
          suggestions: [
            {
              messageId: 'prependVoid',
              desc: 'Prepend `void` to the expression.',
              output:
                'foo.forEach((x) => { if (a === b) { return void (x + a); } })',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.filter(bar => { baz(); } )',
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.filter() expects a return value from arrow function.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'foo.filter(\n() => {} )',
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.filter() expects a return value from arrow function.',
          line: 2,
          column: 4,
          endLine: 2,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'foo.filter(bar || ((baz) => {}) )',
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.filter() expects a return value from arrow function.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'foo.filter(bar => { return; })',
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedReturnValue',
          message:
            'Array.prototype.filter() expects a return value from arrow function.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'Array.from(foo, bar => { bar })',
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedInside',
          message: 'Array.from() expects a return value from arrow function.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'foo.forEach(bar => bar)',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'wrapBraces',
              desc: 'Wrap the expression in `{}`.',
              output: 'foo.forEach(bar => {bar})',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach((function () { return (bar) => bar; })())',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'wrapBraces',
              desc: 'Wrap the expression in `{}`.',
              output: 'foo.forEach((function () { return (bar) => {bar}; })())',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach((() => {\n return bar => bar; })())',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'wrapBraces',
              desc: 'Wrap the expression in `{}`.',
              output: 'foo.forEach((() => {\n return bar => {bar}; })())',
            },
          ],
        },
      ],
    },
    {
      code: 'foo.forEach((bar) => { if (bar) { return; } else { return bar ; } })',
      options: [
        {
          checkForEach: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from arrow function.',
          line: 1,
          column: 52,
          endLine: 1,
          endColumn: 64,
        },
      ],
    },
    {
      code: 'foo.filter(function(){})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.filter() expects a return value from function.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'foo.filter(function (){})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.filter() expects a return value from function.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'foo.filter(function\n(){})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.filter() expects a return value from function.',
          line: 1,
          column: 12,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'foo.filter(function bar(){})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.filter() expects a return value from function 'bar'.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'foo.filter(function bar  (){})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.filter() expects a return value from function 'bar'.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'foo.filter(function\n bar() {})',
      errors: [
        {
          messageId: 'expectedInside',
          message:
            "Array.prototype.filter() expects a return value from function 'bar'.",
          line: 1,
          column: 12,
          endLine: 2,
          endColumn: 5,
        },
      ],
    },
    {
      code: 'Array.from(foo, function bar(){})',
      errors: [
        {
          messageId: 'expectedInside',
          message: "Array.from() expects a return value from function 'bar'.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'Array.from(foo, bar ? function (){} : baz)',
      errors: [
        {
          messageId: 'expectedInside',
          message: 'Array.from() expects a return value from function.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'foo.filter(function bar() { return \n })',
      errors: [
        {
          messageId: 'expectedReturnValue',
          message:
            "Array.prototype.filter() expects a return value from function 'bar'.",
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'foo.forEach(function () { \nif (baz) return bar\nelse return\n })',
      options: [
        {
          checkForEach: true,
        },
      ],
      errors: [
        {
          messageId: 'expectedNoReturnValue',
          message:
            'Array.prototype.forEach() expects no useless return value from function.',
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 20,
        },
      ],
    },
    {
      code: "foo?.filter(() => { console.log('hello') })",
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.filter() expects a return value from arrow function.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: "(foo?.filter)(() => { console.log('hello') })",
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.filter() expects a return value from arrow function.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: "Array?.from([], () => { console.log('hello') })",
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'expectedInside',
          message: 'Array.from() expects a return value from arrow function.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: "(Array?.from)([], () => { console.log('hello') })",
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'expectedInside',
          message: 'Array.from() expects a return value from arrow function.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: "foo?.filter((function() { return () => { console.log('hello') } })?.())",
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'expectedInside',
          message:
            'Array.prototype.filter() expects a return value from arrow function.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
  ],
};

ruleTester.run('array-callback-return', cases);

// ESLint v9.39.1 shares this suggestion bug. Keep the upstream corpus intact
// and separately verify safe outputs for bodies that need expression context.
test('wrapBraces preserves callback expression context', async () => {
  const regressionCases = [
    ['() => function () {}', '() => {(function () {})}'],
    ['() => class {}', '() => {(class {})}'],
    ['(x = 0) => "use strict"', '(x = 0) => {("use strict")}'],
    ['x => "use strict"', 'x => {("use strict")}'],
    ["(...x) => 'use strict'", "(...x) => {('use strict')}"],
    ['({x}) => "use strict"', '({x}) => {("use strict")}'],
    ['() => function named() {}', '() => {(function named() {})}'],
    ['() => class Named {}', '() => {(class Named {})}'],
    ['() => async function () {}', '() => {(async function () {})}'],
    ['() => function* () {}', '() => {(function* () {})}'],
    ['() => /* keep */ function () {}', '() => /* keep */ {(function () {})}'],
    ['() => /* keep */ class {}', '() => /* keep */ {(class {})}'],
    [
      '(x = 0) => /* keep */ "use strict"',
      '(x = 0) => /* keep */ {("use strict")}',
    ],
    ['() => (function () {})', '() => {(function () {})}'],
    ['() => (class {})', '() => {(class {})}'],
    ['(x = 0) => ("use strict")', '(x = 0) => {("use strict")}'],
    ['() => ((class {}))', '() => {((class {}))}'],
    ['(x = 0) => "ordinary"', '(x = 0) => {"ordinary"}'],
  ];
  const root = await realpath(
    await mkdtemp(path.join(tmpdir(), 'rslint-array-return-suggestions-')),
  );
  try {
    for (const allowVoid of [false, true]) {
      for (const [callback, wrappedCallback] of regressionCases) {
        const code = `items.forEach(${callback});`;
        const request = {
          workingDirectory: root,
          configDirectory: root,
          config: [
            {
              plugins: [],
              languageOptions: { parserOptions: { projectService: false } },
              rules: {
                'array-callback-return': [
                  'error',
                  { checkForEach: true, allowVoid },
                ],
              },
            },
          ],
          fileContents: { [path.join(root, 'input.js')]: code },
        };
        const result = await lint(request);
        assert.equal(result.fileCount, 1);
        assert.equal(result.ruleCount, 1);
        assert.equal(result.diagnostics.length, 1, code);
        const diagnostic = result.diagnostics[0];
        assert.equal(diagnostic.messageId, 'expectedNoReturnValue');
        assert.equal(
          diagnostic.message,
          'Array.prototype.forEach() expects no useless return value from arrow function.',
        );
        assert.deepEqual(diagnostic.range, {
          start: { line: 1, column: code.indexOf('=>') + 1 },
          end: { line: 1, column: code.indexOf('=>') + 3 },
        });
        assert.equal(diagnostic.fixes, undefined);
        const suggestions = diagnostic.suggestions ?? [];
        assert.deepEqual(
          suggestions.map((suggestion) => suggestion.messageId),
          allowVoid ? ['wrapBraces', 'prependVoid'] : ['wrapBraces'],
        );
        assert.equal(suggestions[0].message, 'Wrap the expression in `{}`.');
        for (const suggestion of suggestions) {
          assert.ok(suggestion.fixes?.length);
          const output = applyEdits(code, suggestion.fixes);
          if (suggestion.messageId === 'wrapBraces') {
            assert.equal(output, `items.forEach(${wrappedCallback});`);
          }
          const program = parse(output, {
            ecmaVersion: 'latest',
            sourceType: 'module',
          });
          if (suggestion.messageId === 'wrapBraces') {
            const statement = program.body[0];
            assert.ok(statement.type === 'ExpressionStatement');
            const call = statement.expression;
            assert.ok(call.type === 'CallExpression');
            const parsedCallback = call.arguments[0];
            assert.ok(parsedCallback.type === 'ArrowFunctionExpression');
            const body = parsedCallback.body;
            assert.ok(body.type === 'BlockStatement');
            assert.equal(body.body.length, 1);
            assert.ok(body.body[0].type === 'ExpressionStatement');
            // The original string value must not become a directive.
            if (callback.includes('use strict')) {
              assert.equal(Reflect.get(body.body[0], 'directive'), undefined);
            }
          }
          const relinted = await lint({
            ...request,
            fileContents: { [path.join(root, 'input.js')]: output },
          });
          assert.equal(relinted.fileCount, 1);
          assert.equal(relinted.ruleCount, 1);
          assert.deepEqual(relinted.diagnostics, [], output);
        }
      }
    }
  } finally {
    await rm(root, { recursive: true, force: true });
  }
});

// Suggestions must stay opt-in even when the real autofix pipeline runs.
test('array-callback-return does not autofix any upstream invalid case', async () => {
  const root = await realpath(
    await mkdtemp(path.join(tmpdir(), 'rslint-array-return-')),
  );
  try {
    for (const [index, item] of cases.invalid.entries()) {
      const result = await lint({
        workingDirectory: root,
        configDirectory: root,
        config: [
          {
            plugins: [],
            languageOptions: {
              ...item.languageOptions,
              parserOptions: { projectService: false },
            },
            rules: {
              'array-callback-return': [
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
      assert.deepEqual(
        result.output ?? {},
        {},
        `invalid ${index}: no automatic edits`,
      );
      assert.equal(
        result.diagnostics.length,
        item.errors.length,
        `invalid ${index}: diagnostics remain`,
      );
    }
  } finally {
    await rm(root, { recursive: true, force: true });
  }
});
