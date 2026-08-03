import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// Mirrors upstream test/prefer-node-protocol.js (Layer 1). This suite verifies
// registration + wire protocol + ESLint-compatible diagnostic shape end-to-end;
// tsgo-specific edge shapes and branch lock-ins live in the Go extras suite.
const js = (code: string) => ({ code, filename: 'file.js' });
const jsInvalid = (code: string, errors = 1) => ({
  code,
  filename: 'file.js',
  errors,
});
const ts = (code: string) => ({ code, filename: 'file.ts' });
const tsInvalid = (code: string, errors = 1) => ({
  code,
  filename: 'file.ts',
  errors,
});

ruleTester.run('prefer-node-protocol', null as never, {
  valid: [
    // import / export / dynamic import
    js('import unicorn from "unicorn";'),
    js('import fs from "./fs";'),
    js('import fs from "unknown-builtin-module";'),
    js('import fs from "node:fs";'),
    js('async function foo() {\n\tconst fs = await import(fs);\n}'),
    js('async function foo() {\n\tconst fs = await import(0);\n}'),
    js('async function foo() {\n\tconst fs = await import(`fs`);\n}'),
    js('import "punycode";'),
    js('import "node:punycode";'),
    js('import "punycode/";'),
    js('import "fs/";'),
    js('import "test";'),
    js('import "node:test";'),
    js('import "bun";'),
    js('import "bun:jsc";'),
    js('import "bun:sqlite";'),

    // require
    js('const fs = require("node:fs");'),
    js('const fs = require("node:fs/promises");'),
    js('const fs = require(fs);'),
    js('const fs = notRequire("fs");'),
    js('const fs = foo.require("fs");'),
    js('const fs = require.resolve("fs");'),
    js('const fs = require(`fs`);'),
    js('const fs = require?.("fs");'),
    js('const fs = require("fs", extra);'),
    js('const fs = require();'),
    js('const fs = require(...["fs"]);'),
    js('const fs = require("unicorn");'),

    // process.getBuiltinModule
    js('const fs = process.getBuiltinModule("node:fs")'),
    js('const fs = process.getBuiltinModule?.("fs")'),
    js('const fs = process?.getBuiltinModule("fs")'),
    js('const fs = process.notGetBuiltinModule("fs")'),
    js('const fs = notProcess.getBuiltinModule("fs")'),
    js('const fs = process.getBuiltinModule("fs", extra)'),
    js('const fs = process.getBuiltinModule(...["fs"])'),
    js('const fs = process.getBuiltinModule()'),
    js('const fs = process.getBuiltinModule("unicorn")'),
    js(
      'import {getBuiltinModule} from \'node:process\';\nconst fs = getBuiltinModule("fs");',
    ),

    // TypeScript
    ts('const fs = require("node:fs") as "fs";'),
    ts('type fs = typeof import("node:fs");'),
    ts('type fs = typeof SomeType<"fs">;'),
    ts('type fs = typeof fs;'),
  ],
  invalid: [
    // import / export / dynamic import
    jsInvalid('import fs from "fs";'),
    jsInvalid('export {promises} from "fs";'),
    jsInvalid("async function foo() {\n\tconst fs = await import('fs');\n}"),
    jsInvalid('import fs from "fs/promises";'),
    jsInvalid('export {default} from "fs/promises";'),
    jsInvalid(
      "async function foo() {\n\tconst fs = await import('fs/promises');\n}",
    ),
    jsInvalid('import {promises} from "fs";'),
    jsInvalid('export {default as promises} from "fs";'),
    jsInvalid("import {promises} from 'fs';"),
    jsInvalid(
      'async function foo() {\n\tconst fs = await import("fs/promises");\n}',
    ),
    jsInvalid(
      'async function foo() {\n\tconst fs = await import(/* escaped */"\\u{66}s/promises");\n}',
    ),
    jsInvalid('import "buffer";'),
    jsInvalid('import "child_process";'),
    jsInvalid('import "timers/promises";'),

    // require
    jsInvalid('const {promises} = require("fs")'),
    jsInvalid("const fs = require('fs/promises')"),

    // process.getBuiltinModule
    jsInvalid('const fs = process.getBuiltinModule("fs")'),

    // TypeScript
    tsInvalid('const fs = require("fs") as typeof import("fs");', 2),
    tsInvalid('type fs = import("fs");'),
    tsInvalid('type fs = import("fs").fs<"fs">'),
    tsInvalid('const fs = someFunc() as SomeType<typeof import("fs")>;'),
  ],
});
