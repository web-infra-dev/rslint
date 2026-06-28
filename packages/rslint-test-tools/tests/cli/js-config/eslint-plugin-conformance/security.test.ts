/**
 * Conformance: eslint-plugin-security (v4.0.1) mounted in rslint via `plugins`
 * must report identically to ESLint v10. All 14 rules reproduce ESLint v10
 * byte-for-byte: they are plain AST pattern matches (detect-* dangerous usage)
 * with no type info, scope-heavy data flow, or directive handling, so rslint and
 * ESLint agree exactly — no rules or cases excluded. Representative triggers from
 * the upstream test suite.
 */
import { runConformanceSuite } from './conformance.js';
import type { DiffCase } from './harness.js';

const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-unsafe-regex',
    code: '/(x+x+)+y/',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-unsafe-regex',
    code: "new RegExp('x+x+)+y')",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-non-literal-regexp',
    code: "var a = new RegExp(c, 'i')",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-non-literal-require',
    code: 'var a = require(c)',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-non-literal-require',
    code: 'var a = require(`${c}`)',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-non-literal-fs-filename',
    code: "var something = require('fs');\n             var a = something.open(c);",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-non-literal-fs-filename',
    code: "var one = require('fs').readFile;\n             one(filename);",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-non-literal-fs-filename',
    code: "import { readFile as something } from 'fs';\n             something(filename);",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-non-literal-fs-filename',
    code: "var fs = require('fs');\nfs.readFile(`template with ${filename}`);",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-eval-with-expression',
    code: 'eval(a);',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-pseudoRandomBytes',
    code: 'crypto.pseudoRandomBytes',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-possible-timing-attacks',
    code: "if (password === 'mypass') {}",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-possible-timing-attacks',
    code: "if ('mypass' === password) {}",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-no-csrf-before-method-override',
    code: 'express.csrf();express.methodOverride()',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-buffer-noassert',
    code: 'a.readUInt8(0, true)',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-buffer-noassert',
    code: 'a.writeUInt8(0, 0, true)',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-buffer-noassert',
    code: 'a.readDoubleLE(0, true);',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-child-process',
    code: "require('child_process')",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-child-process',
    code: "var child = require('child_process'); child.exec(com)",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-child-process',
    code: "import child from 'child_process'; child.exec(com)",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-child-process',
    code: "\n      const {exec} = require('child_process');\n      exec(str)",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-disable-mustache-escape',
    code: 'a.escapeMarkup = false',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-object-injection',
    code: 'var a = {}; a[b] = 4',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-new-buffer',
    code: 'var a = new Buffer(c)',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-bidi-characters',
    code: '\n      var accessLevel = "user";\n      if (accessLevel != "user\u202e \u2066// Check if admin\u2069 \u2066") {\n          console.log("You are an admin.");\n      }\n      ',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-bidi-characters',
    code: '\n      var isAdmin = false;\n      /*\u202e } \u2066if (isAdmin)\u2069 \u2066 begin admins only */\n          console.log("You are an admin.");\n      /* end admins only \u202e\n\u2066*/\n      /* end admins only \u202e\n { \u2066*/\n        ',
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-unsafe-regex',
    code: '/^d+1337d+$/',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-unsafe-regex',
    code: "new RegExp('^d+1337d+$')",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-non-literal-regexp',
    code: "var a = new RegExp('ab+c', 'i')",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-non-literal-regexp',
    code: "\n            var source = 'ab+c'\n            var a = new RegExp(source, 'i')",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-non-literal-require',
    code: "var a = require('b')",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-non-literal-require',
    code: 'var a = require(`b`)',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-non-literal-fs-filename',
    code: "var fs = require('fs');\n             var a = fs.open('test')",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-non-literal-fs-filename',
    code: "var something = require('some');\n             var a = something.readFile(c);",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-eval-with-expression',
    code: "eval('alert()')",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-eval-with-expression',
    code: 'eval("some nefarious code");',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-pseudoRandomBytes',
    code: 'crypto.randomBytes',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-possible-timing-attacks',
    code: 'if (age === 5) {}',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-no-csrf-before-method-override',
    code: 'express.methodOverride();express.csrf()',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-buffer-noassert',
    code: 'a.readUInt8(0)',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-buffer-noassert',
    code: 'a.readUInt8(0, false)',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-child-process',
    code: "child_process.exec('ls')",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-child-process',
    code: "var { spawn } = require('child_process'); spawn(str);",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-disable-mustache-escape',
    code: 'escapeMarkup = false',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-object-injection',
    code: 'var a = {};',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-new-buffer',
    code: "var a = new Buffer('test')",
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-bidi-characters',
    code: '\n  var accessLevel = "user";\n  if (accessLevel != "user") { // Check if admin\n    console.log("You are an admin.");\n  }\n  ',
  },
  {
    pkg: 'eslint-plugin-security',
    rule: 'detect-bidi-characters',
    code: '\n  var isAdmin = false;\n  /* begin admins only */ if (isAdmin) {\n    console.log("You are an admin.");\n  /* end admins only */ }\n  ',
  },
];

runConformanceSuite('eslint-plugin-security', CASES, CLEAN_CASES);
