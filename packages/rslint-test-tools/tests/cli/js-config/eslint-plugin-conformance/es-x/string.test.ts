/**
 * Conformance: eslint-plugin-es-x (string) mounted in rslint via `plugins` must
 * report identically to ESLint v10. es-x rules are AST pattern matches over ES
 * version features / builtin APIs (no type info), so rslint reproduces ESLint
 * byte-for-byte. Representative triggers from the upstream test suite (v9.6.0).
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-fromcodepoint',
    code: 'String.fromCodePoint',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-at',
    code: "'123'.at(-1)",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-endswith',
    code: "'foo'.endsWith('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-includes',
    code: "'foo'.includes('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-iswellformed',
    code: "'foo'.isWellFormed()",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-matchall',
    code: "'foo'.matchAll('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-normalize',
    code: "'foo'.normalize('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-padstart-padend',
    code: "'foo'.padStart(2)",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-padstart-padend',
    code: "'foo'.padEnd(2)",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-repeat',
    code: "'foo'.repeat(3)",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-replaceall',
    code: "'foo'.replaceAll('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-replaceall',
    code: '\n            const str = "I think Ruth\'s dog is cuter than your dog!"\n            if (String.prototype.replaceAll) {\n                console.log(str.replaceAll("dog", "monkey"))\n            }\n            if (typeof String.prototype.replaceAll === \'undefined\') {\n                console.log(str.replaceAll("dog", "monkey"))\n            } else {\n                console.log(str.replaceAll("dog", "monkey"))\n            }\n            const a = String.prototype.replaceAll\n              ? str.replaceAll("dog", "monkey")\n              : str.replaceAll("dog", "monkey");',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-startswith',
    code: "'foo'.startsWith('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-substr',
    code: "'foo'.substr()",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-trim',
    code: "'foo'.trim()",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-trimstart-trimend',
    code: "'foo'.trimStart(2)",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-trimstart-trimend',
    code: "'foo'.trimEnd(2)",
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-string-raw', code: 'String.raw' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-template-literals', code: '`foo`' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-template-literals', code: 'tag`foo`' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-template-literals',
    code: '`foo${a}bar${b}baz`',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-template-literals',
    code: '\nconst a1 = `foo`\nconst a2 = `foo${bar}baz`\nconst a3 = tag`foo`\n            ',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-unicode-codepoint-escapes',
    code: '\\u{45} = 1',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-unicode-codepoint-escapes',
    code: 'a\\u{45}b = 1',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-unicode-codepoint-escapes',
    code: "'\\u{45}'",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-unicode-codepoint-escapes',
    code: "'a\\u{45}b'",
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-create-html-methods',
    code: 'foo.charAt(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-fromcodepoint',
    code: 'String',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-fromcodepoint',
    code: 'String.raw',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-at',
    code: 'foo.at(-1)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-string-prototype-at', code: 'at(-1)' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-at',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-codepointat',
    code: 'foo.codePointAt(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-codepointat',
    code: 'codePointAt(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-codepointat',
    code: 'foo.charAt(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-endswith',
    code: "foo.endsWith('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-endswith',
    code: "endsWith('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-endswith',
    code: 'foo.charAt(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-includes',
    code: "foo.includes('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-includes',
    code: "includes('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-includes',
    code: 'foo.charAt(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-iswellformed',
    code: 'isWellFormed()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-iswellformed',
    code: 'foo.isWellFormed()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-matchall',
    code: "foo.matchAll('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-matchall',
    code: "matchAll('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-matchall',
    code: 'foo.charAt(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-normalize',
    code: "foo.normalize('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-normalize',
    code: "normalize('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-normalize',
    code: 'foo.charAt(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-padstart-padend',
    code: 'foo.padStart(2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-padstart-padend',
    code: 'foo.padEnd(2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-padstart-padend',
    code: 'padStart(2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-padstart-padend',
    code: 'padEnd(2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-repeat',
    code: 'foo.repeat(3)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-repeat',
    code: 'repeat(3)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-repeat',
    code: 'foo.charAt(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-replaceall',
    code: "foo.replaceAll('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-replaceall',
    code: "replaceAll('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-replaceall',
    code: 'foo.charAt(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-startswith',
    code: "foo.startsWith('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-startswith',
    code: "startsWith('a')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-startswith',
    code: 'foo.charAt(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-substr',
    code: 'foo.substr()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-substr',
    code: 'substr()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-substr',
    code: 'foo.charAt(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-trim',
    code: 'foo.trim()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-trim',
    code: 'trim()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-trim',
    code: 'foo.charAt(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-trimleft-trimright',
    code: 'foo.charAt(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-trimstart-trimend',
    code: 'foo.trimStart(2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-trimstart-trimend',
    code: 'foo.trimEnd(2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-trimstart-trimend',
    code: 'trimStart(2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-prototype-trimstart-trimend',
    code: 'trimEnd(2)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-string-raw', code: 'String' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-string-raw',
    code: 'String.fromCodePoint',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-template-literals', code: "'foo'" },
  { pkg: 'eslint-plugin-es-x', rule: 'no-template-literals', code: '"bar"' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-unicode-codepoint-escapes',
    code: 'foo = 1',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-unicode-codepoint-escapes',
    code: '\\u0045 = 1',
  },
];

runConformanceSuite('eslint-plugin-es-x', CASES, CLEAN_CASES);
