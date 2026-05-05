import { test } from '../utils.js';

import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();

ruleTester.run('default', null as never, {
  valid: [
    // ===== Upstream valid (top section) =====
    // SKIP: `import "./malformed.js"` (parse-error fixture). Equivalent
    // bare-import-without-default is covered below.
    test({ code: 'import "./default-files/named-exports";' }),
    test({ code: 'import foo from "./default-files/empty-folder";' }),
    test({ code: 'import { c } from "./default-files/default-export";' }),
    test({ code: 'import foo from "./default-files/default-export";' }),
    test({ code: 'import foo from "./default-files/mixed-exports";' }),
    test({ code: 'import bar from "./default-files/default-export";' }),
    test({ code: 'import CoolClass from "./default-files/default-class";' }),
    test({
      code: 'import bar, { named } from "./default-files/mixed-exports";',
    }),
    // core modules always have a default
    test({ code: 'import crypto from "crypto";' }),
    test({ code: 'import common from "./default-files/common";' }),

    // ES7 / Babel-only export forms not parseable by TS — SKIP: not parseable
    // by TS (`export bar from`, `export bar, { foo } from`, `export bar, * as
    // names from`).
    test({
      code: 'export { default as bar } from "./default-files/default-export";',
    }),
    test({
      code: 'export { default as bar, named } from "./default-files/mixed-exports";',
    }),

    // sanity / regression
    test({ code: 'export { a } from "./default-files/named-exports";' }),
    // #54: import of named-default-export
    test({ code: 'import foo from "./default-files/named-default-export";' }),
    // #94: redux-style `export default connect(App)`
    test({ code: 'import connectedApp from "./default-files/redux";' }),
    // trampoline that resolves
    test({ code: 'import twofer from "./default-files/trampoline";' }),
    // deeper alias chain (rslint extension)
    test({ code: 'import threefer from "./default-files/deep-trampoline";' }),

    // ES2022 arbitrary module-namespace identifier
    test({
      code: 'export { "default" as bar } from "./default-files/default-export";',
    }),

    // JSX (TSX)
    test({
      code: 'import MyCoolComponent from "./default-files/jsx/MyCoolComponent";',
    }),
    test({ code: 'import App from "./default-files/jsx/App";' }),

    // SKIP: #545 Babel-old-only cases (`./default-export-from.js`, etc.).

    // ===== Upstream valid (TypeScript section) =====
    test({ code: 'import foobar from "./default-files/typescript-default";' }),
    test({
      code: 'import foobar from "./default-files/typescript-export-assign-default";',
    }),
    test({
      code: 'import foobar from "./default-files/typescript-export-assign-function";',
    }),
    test({
      code: 'import foobar from "./default-files/typescript-export-assign-mixed";',
    }),
    test({
      code: 'import foobar from "./default-files/typescript-export-assign-default-reexport";',
    }),
    test({
      code: 'import React from "./default-files/typescript-export-assign-default-namespace";',
    }),
    // SKIP: `./typescript-export-react-test-renderer` and
    // `./typescript-extended-config` need parserOptions.tsconfigRootDir override.
    test({
      code: 'import foobar from "./default-files/typescript-export-assign-property";',
    }),

    // ===== rslint-specific: anonymous default shapes =====
    test({
      code: 'import fn from "./default-files/default-export-anonymous-fn";',
    }),
    test({
      code: 'import C from "./default-files/default-export-anonymous-class";',
    }),
    test({ code: 'import a from "./default-files/default-export-arrow";' }),
    test({ code: 'import L from "./default-files/default-export-literal";' }),
    test({
      code: 'import asyncFn from "./default-files/default-export-async";',
    }),
    test({
      code: 'import gen from "./default-files/default-export-generator";',
    }),

    // ===== rslint-specific: alias / rename forms =====
    test({ code: 'import foo from "./default-files/local-rename-default";' }),
    test({ code: 'import x from "./default-files/named-default-via-rename";' }),
    test({
      code: 'import { default as bar } from "./default-files/default-export";',
    }),
    test({
      code: 'import foo, { default as bar } from "./default-files/default-export";',
    }),

    // ===== rslint-specific: import-shape variations =====
    test({ code: 'import * as ns from "./default-files/named-exports";' }),
    test({ code: 'import type Foo from "./default-files/default-export";' }),
    test({
      code: 'import foo, * as ns from "./default-files/default-export";',
    }),
    test({
      code: 'import foo from "./default-files/default-export" with { type: "module" };',
    }),
    test({ code: 'import foo from "./default-files/default-export.ts";' }),
    test({ code: 'import idx from "./default-files/index-folder";' }),
    test({
      code: 'import a from "./default-files/default-export"; import b from "./default-files/default-export";',
    }),

    // ===== rslint-specific: graceful skips =====
    test({ code: 'import foo from "./default-files/non-existent";' }),
    test({ code: 'export { default } from "./default-files/default-export";' }),
    test({ code: 'const m = import("./default-files/named-exports");' }),
    test({ code: 'const m = require("./default-files/named-exports");' }),
    test({ code: 'import m = require("./default-files/named-exports");' }),

    // ===== rslint-specific: parse-error robustness =====
    // Source has a parse error; default still bound — rule passes.
    test({
      code: 'import x from "./default-files/syntax-broken-with-default";',
    }),

    // ===== rslint-specific: declaration-file (`.d.ts`) source =====
    test({ code: 'import v from "./default-files/decl-with-default";' }),

    // (settings-driven cases live in dedicated isolated configs below.)

    // ===== rslint-specific: default value variants =====
    test({ code: 'import x from "./default-files/default-undefined";' }),
    test({ code: 'import x from "./default-files/default-null";' }),

    // ===== rslint-specific: binding identifier variations =====
    test({ code: 'import yield_ from "./default-files/default-export";' }),
    test({ code: 'import async_ from "./default-files/default-export";' }),

    // ===== rslint-specific: schema-empty contract =====
    // upstream `schema: []` — options must be ignored
    test({
      code: 'import foo from "./default-files/default-export";',
      options: [{ someOpt: true }],
    }),
  ],
  invalid: [
    // ===== Upstream invalid (top section) =====
    // SKIP: parse-errors-in-imported-module case (delegated to type-check).

    test({
      code: 'import baz from "./default-files/named-exports";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/named-exports".',
        },
      ],
    }),
    // SKIP: ES7 babel-only `export baz from`, `export baz, { bar } from`,
    // `export baz, * as names from` cases.

    test({
      code: 'import twofer from "./default-files/broken-trampoline";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/broken-trampoline".',
        },
      ],
    }),
    // #328
    test({
      code: 'import barDefault from "./default-files/re-export";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/re-export".',
        },
      ],
    }),

    // ===== Upstream invalid (TypeScript section) =====
    test({
      code: 'import foobar from "./default-files/typescript";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/typescript".',
        },
      ],
    }),
    // SKIP: `typescript-export-as-default-namespace` invalid variant requires
    // a separate no-compiler-options tsconfig; covered indirectly above.

    // ===== rslint-specific: combined import shapes =====
    test({
      code: 'import baz, { a } from "./default-files/named-exports";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/named-exports".',
        },
      ],
    }),
    test({
      code: 'import baz, * as ns from "./default-files/named-exports";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/named-exports".',
        },
      ],
    }),
    test({
      code: 'import type Baz from "./default-files/named-exports";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/named-exports".',
        },
      ],
    }),

    // ===== rslint-specific: folder / index resolution =====
    test({
      code: 'import idx from "./default-files/index-folder-no-default";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/index-folder-no-default".',
        },
      ],
    }),

    // ===== rslint-specific: cycle that bottoms out =====
    test({
      code: 'import x from "./default-files/circular-a";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/circular-a".',
        },
      ],
    }),

    // ===== rslint-specific: type-only / empty modules =====
    test({
      code: 'import T from "./default-files/type-only-exports";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/type-only-exports".',
        },
      ],
    }),
    test({
      code: 'import x from "./default-files/empty-module";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/empty-module".',
        },
      ],
    }),

    // ===== rslint-specific: schema-empty contract =====
    test({
      code: 'import baz from "./default-files/named-exports";',
      options: [{ unknown: true }],
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/named-exports".',
        },
      ],
    }),

    // ===== rslint-specific: parse-error robustness (no default) =====
    test({
      code: 'import x from "./default-files/syntax-broken-no-default";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/syntax-broken-no-default".',
        },
      ],
    }),

    // ===== rslint-specific: declaration file with no default =====
    test({
      code: 'import x from "./default-files/decl-no-default";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/decl-no-default".',
        },
      ],
    }),

    // (settings-driven invalid cases live in dedicated isolated configs below.)

    // ===== rslint-specific: two missing defaults in one file =====
    test({
      code: 'import a from "./default-files/named-exports";\nimport b from "./default-files/re-export";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/named-exports".',
        },
        {
          message:
            'No default export found in imported module "./default-files/re-export".',
        },
      ],
    }),
  ],
});

// =====================================================
// Isolated config: resolveJsonModule + esModuleInterop
// (own rslint.default-isolated.json + own tsconfig — does not pollute the
// shared import-plugin config used by other rules' tests)
// =====================================================
const jsonRuleTester = new RuleTester({
  config: './rslint.default-isolated.json',
});

jsonRuleTester.run('default', null as never, {
  valid: [test({ code: 'import data from "./default-files/data.json";' })],
  invalid: [],
});

// =====================================================
// Isolated config: no esModuleInterop / no allowSyntheticDefaultImports
// `export = X` is no longer accepted as a default import target.
// =====================================================
const noInteropRuleTester = new RuleTester({
  config: './rslint.default-no-interop.json',
});

noInteropRuleTester.run('default', null as never, {
  valid: [
    // ES `export default X` still works without interop.
    test({ code: 'import foo from "./default-files/default-export";' }),
  ],
  invalid: [
    test({
      code: 'import fn from "./default-files/typescript-export-assign-function";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/typescript-export-assign-function".',
        },
      ],
    }),
    test({
      code: 'import Foo from "./default-files/typescript-export-as-default-namespace";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/typescript-export-as-default-namespace".',
        },
      ],
    }),
  ],
});

// =====================================================
// Isolated config: settings["import/ignore"] = ["named-exports"]
// Pattern matches the source path → rule skips even though the resolved
// module has no default.
// =====================================================
const ignoreNamedRuleTester = new RuleTester({
  config: './rslint.default-ignore-named.json',
});

ignoreNamedRuleTester.run('default', null as never, {
  valid: [test({ code: 'import baz from "./default-files/named-exports";' })],
  invalid: [],
});

// =====================================================
// Isolated config: settings["import/ignore"] = []
// User explicitly opts out of every default skip — local files with no
// default must still report.
// =====================================================
const ignoreEmptyRuleTester = new RuleTester({
  config: './rslint.default-ignore-empty.json',
});

ignoreEmptyRuleTester.run('default', null as never, {
  valid: [test({ code: 'import foo from "./default-files/default-export";' })],
  invalid: [
    test({
      code: 'import baz from "./default-files/named-exports";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/named-exports".',
        },
      ],
    }),
  ],
});

// =====================================================
// Isolated config: allowJs + esModuleInterop
// `.mjs` / `.cjs` source files become resolvable.
// =====================================================
const allowJsRuleTester = new RuleTester({
  config: './rslint.default-allow-js.json',
});

allowJsRuleTester.run('default', null as never, {
  valid: [
    test({ code: 'import x from "./default-files/mjs-default.mjs";' }),
    test({ code: 'import x from "./default-files/cjs-module-exports.cjs";' }),
  ],
  invalid: [
    test({
      code: 'import x from "./default-files/mjs-named-only.mjs";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/mjs-named-only.mjs".',
        },
      ],
    }),
    test({
      code: 'import x from "./default-files/cjs-named-exports.cjs";',
      errors: [
        {
          message:
            'No default export found in imported module "./default-files/cjs-named-exports.cjs".',
        },
      ],
    }),
  ],
});
