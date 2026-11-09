import { RuleTester } from '../rule-tester';

// Inputs and diagnostic messages from eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-export-from.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-export-from.md
// The Go suite checks ranges, fixes and suggestions.
const tester = new RuleTester();

function reexport(exported: string) {
  return {
    messageId: 'error',
    message: `Use \`export…from\` to re-export \`${exported}\`.`,
  };
}

// Captured Bindings
tester.run('prefer-export-from', null as never, {
  valid: [],
  invalid: [
    {
      code: 'import value from "source";\nexport default (value);',
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: 'import value from "source";\nexport const snapshot = (value);',
      filename: 'src/virtual.js',
      errors: [reexport('snapshot')],
    },
    {
      code: 'import {value} from "source";\nexport default /* keep */ value;',
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: 'import {value} from "source";\nexport const snapshot = /* keep */ value;',
      filename: 'src/virtual.js',
      errors: [reexport('snapshot')],
    },
  ],
});

// Java Script
tester.run('prefer-export-from', null as never, {
  valid: [
    {
      code: 'import "foo";',
      filename: 'src/virtual.js',
    },
    {
      code: 'import {} from "foo";',
      filename: 'src/virtual.js',
    },
    {
      code: 'import * as namespace from "foo";',
      filename: 'src/virtual.js',
    },
    {
      code: 'import defaultExport from "foo";',
      filename: 'src/virtual.js',
    },
    {
      code: 'import {named} from "foo";',
      filename: 'src/virtual.js',
    },
    {
      code: 'const named = import(foo);\nexport {named};',
      filename: 'src/virtual.js',
    },
    {
      code: 'export * from "foo";',
      filename: 'src/virtual.js',
    },
    {
      code: 'export {default} from "foo";',
      filename: 'src/virtual.js',
    },
    {
      code: 'export {named} from "foo";',
      filename: 'src/virtual.js',
    },
    {
      code: "const defaultExport = require('foo');\nexport default defaultExport;",
      filename: 'src/virtual.js',
    },
    {
      code: "import {foo} from 'foo';\n\n/**\n * Add documentation to the re-export.\n */\nexport const baz = foo;",
      filename: 'src/virtual.js',
    },
    {
      code: "import {foo} from 'foo';\n/** Add documentation to the re-export. */ export const baz = foo;",
      filename: 'src/virtual.js',
    },
    {
      code: "import defaultExport from 'foo';\nexport var variable = defaultExport;",
      filename: 'src/virtual.js',
    },
    {
      code: "import defaultExport from 'foo';\nexport let variable = defaultExport;",
      filename: 'src/virtual.js',
    },
    {
      code: "import defaultExport from 'foo';\nexport const variable = defaultExport;\nuse(variable);",
      filename: 'src/virtual.js',
    },
    {
      code: "import defaultExport from 'foo';\nexport let variable = defaultExport;\nvariable = 1;",
      filename: 'src/virtual.js',
    },
    {
      code: "import * as namespace from 'foo';\nexport default namespace;",
      filename: 'src/virtual.js',
    },
    {
      code: "import * as namespace from 'foo';\nexport {namespace as default};",
      filename: 'src/virtual.js',
    },
    {
      code: "import defaultExport from 'foo';\nconst variable = defaultExport;\nexport {variable}",
      filename: 'src/virtual.js',
    },
    {
      code: "import defaultExport from 'foo';\nexport const {variable} = {variable: defaultExport};",
      filename: 'src/virtual.js',
    },
  ],
  invalid: [
    {
      code: "import defaultExport from 'foo';\nexport default defaultExport;",
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: "import defaultExport from 'foo';\nexport {defaultExport as default};",
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: "import defaultExport from 'foo';\nexport {defaultExport as named};",
      filename: 'src/virtual.js',
      errors: [reexport('named')],
    },
    {
      code: "import defaultExport from 'foo';\nexport const variable = defaultExport;",
      filename: 'src/virtual.js',
      errors: [reexport('variable')],
    },
    {
      code: "import {default as defaultExport} from 'foo';\nexport default defaultExport;",
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: "import {default as defaultExport} from 'foo';\nexport {defaultExport as default};",
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: "import {default as defaultExport} from 'foo';\nexport {defaultExport as named};",
      filename: 'src/virtual.js',
      errors: [reexport('named')],
    },
    {
      code: "import defaultExport from 'foo';\ndefaultExport.bar = 1;\nexport {defaultExport as named};\nexport {defaultExport as default};\nexport const variable = defaultExport;",
      filename: 'src/virtual.js',
      errors: [reexport('named'), reexport('default'), reexport('variable')],
    },
    {
      code: "import {named} from 'foo';\nexport default named;",
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: "import {named} from 'foo';\nexport {named as default};",
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: "import {named} from 'foo';\nexport {named as named};",
      filename: 'src/virtual.js',
      errors: [reexport('named')],
    },
    {
      code: "import {named} from 'foo';\nexport {named as renamed};",
      filename: 'src/virtual.js',
      errors: [reexport('renamed')],
    },
    {
      code: "import {named} from 'foo';\nexport const variable = named;",
      filename: 'src/virtual.js',
      errors: [reexport('variable')],
    },
    {
      code: "import {named} from 'foo';\nnamed.bar = 1;\nexport {named as named};\nexport {named as default};\nexport const variable = named;",
      filename: 'src/virtual.js',
      errors: [reexport('named'), reexport('default'), reexport('variable')],
    },
    {
      code: "import * as namespace from 'foo';\nexport {namespace as namespace};",
      filename: 'src/virtual.js',
      errors: [reexport('namespace')],
    },
    {
      code: "import * as namespace from 'foo';\nexport {namespace as renamed};",
      filename: 'src/virtual.js',
      errors: [reexport('renamed')],
    },
    {
      code: "import * as namespace from 'foo';\nexport const variable = namespace;",
      filename: 'src/virtual.js',
      errors: [reexport('variable')],
    },
    {
      code: "import * as namespace from 'foo';\nnamespace.bar = 1;\nexport {namespace as named};\nexport {namespace as default};\nexport const variable = namespace;",
      filename: 'src/virtual.js',
      errors: [reexport('named'), reexport('variable')],
    },
    {
      code: "import {named1, named2} from 'foo';\nexport {named1};",
      filename: 'src/virtual.js',
      errors: [reexport('named1')],
    },
    {
      code: "import defaultExport, {named} from 'foo';\nexport {defaultExport};",
      filename: 'src/virtual.js',
      errors: [reexport('defaultExport')],
    },
    {
      code: "import defaultExport, {named} from 'foo';\nexport {named};",
      filename: 'src/virtual.js',
      errors: [reexport('named')],
    },
    {
      code: "import defaultExport, * as namespace from 'foo';\nexport {defaultExport};",
      filename: 'src/virtual.js',
      errors: [reexport('defaultExport')],
    },
    {
      code: "import * as foo from 'foo';\nexport {foo};\nexport * as bar from 'foo';",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: "import * as foo from 'foo';\nexport {foo};\nexport {bar} from 'foo';",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: "import * as foo from 'foo';\nexport {foo};\nexport {} from 'foo';",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: "import * as foo from 'foo';\nexport {foo};\nexport * from 'foo';",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: "import foo from 'foo';\nexport {foo};\nexport * as bar from 'foo';",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: "import foo from 'foo';\nexport {foo};\nexport {bar} from 'foo';",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: "import foo from 'foo';\nexport {foo};\nexport {bar,} from 'foo';",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: "import foo from 'foo';\nexport {foo};\nexport {} from 'foo';",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: "import foo from 'foo';\nexport {foo};\nexport * from 'foo';",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: "import {named1, named2} from 'foo';\nexport {named1, named2};",
      filename: 'src/virtual.js',
      errors: [reexport('named1'), reexport('named2')],
    },
    {
      code: "import {named} from 'foo';\nexport {named as default, named};",
      filename: 'src/virtual.js',
      errors: [reexport('default'), reexport('named')],
    },
    {
      code: "import {named, named as renamed} from 'foo';\nexport {named, renamed};",
      filename: 'src/virtual.js',
      errors: [reexport('named'), reexport('renamed')],
    },
    {
      code: "import defaultExport, {named1, named2} from 'foo';\nexport {named1 as default};\nexport {named2};\nexport {defaultExport};",
      filename: 'src/virtual.js',
      errors: [
        reexport('default'),
        reexport('named2'),
        reexport('defaultExport'),
      ],
    },
    {
      code: "import * as foo from 'foo';\nimport * as bar from 'foo';\nexport {foo, bar};",
      filename: 'src/virtual.js',
      errors: [reexport('foo'), reexport('bar')],
    },
    {
      code: "import * as foo from 'foo';\nexport {foo, foo as bar};",
      filename: 'src/virtual.js',
      errors: [reexport('foo'), reexport('bar')],
    },
    {
      code: "import defaultExport from 'foo';\nexport * from 'foo';\nexport default defaultExport;",
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: "import defaultExport from 'foo';\nexport {named} from 'foo';\nexport * from 'foo';\nexport default defaultExport;",
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: "import defaultExport from './foo.js';\nexport {named} from './foo.js';\nexport default defaultExport;",
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: "import defaultExport from './foo.js';\nexport {named} from './foo.js?query';\nexport default defaultExport;",
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: "import * as namespace from 'foo';\nexport default namespace;\nexport {namespace};",
      filename: 'src/virtual.js',
      errors: [reexport('namespace')],
    },
    {
      code: "import * as namespace from 'foo';\nexport {namespace};\nexport default namespace;",
      filename: 'src/virtual.js',
      errors: [reexport('namespace')],
    },
    {
      code: "import {'foo' as foo} from 'foo';\nexport default foo;",
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: "import {'foo' as foo} from 'foo';\nexport {foo};",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: "import {'foo' as foo} from 'foo';\nexport const bar = foo;",
      filename: 'src/virtual.js',
      errors: [reexport('bar')],
    },
    {
      code: "import {'foo' as foo} from 'foo';\nexport {foo as 'foo'};",
      filename: 'src/virtual.js',
      errors: [reexport("'foo'")],
    },
    {
      code: "import {foo} from 'foo';\nexport {foo as 'bar'};",
      filename: 'src/virtual.js',
      errors: [reexport("'bar'")],
    },
    {
      code: "import {'foo' as foo} from 'foo';\nexport {foo as \"foo\"};",
      filename: 'src/virtual.js',
      errors: [reexport('"foo"')],
    },
    {
      code: "import {'fo\\u{20}o' as foo} from 'foo';\nexport {foo as \"fo o\"};",
      filename: 'src/virtual.js',
      errors: [reexport('"fo o"')],
    },
    {
      code: "import {'fo\\no' as foo} from 'foo';\nexport {foo as \"fo\\u000ao\"};",
      filename: 'src/virtual.js',
      errors: [reexport('"fo\\u000ao"')],
    },
    {
      code: "import {'default' as foo} from 'foo';\nexport {foo};",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: "import {'default' as foo} from 'foo';\nexport default foo;",
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: "import {'*' as foo} from 'foo';\nexport {foo};",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: "import {named} from 'foo';\nuse(named);\nexport {named};",
      filename: 'src/virtual.js',
      errors: [reexport('named')],
    },
    {
      code: "import {foo} from 'foo';\n// This is not JSDoc.\nexport const baz = foo;",
      filename: 'src/virtual.js',
      errors: [reexport('baz')],
    },
    {
      code: "import {foo} from 'foo';\n\n/**\n * Documentation for the direct export.\n */\nexport {foo};",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: "import {foo} from 'foo';\n/*\n * This is not JSDoc.\n */\nexport const baz = foo;",
      filename: 'src/virtual.js',
      errors: [reexport('baz')],
    },
    {
      code: "import {foo} from 'foo';\n/*** This is not JSDoc. */\nexport const baz = foo;",
      filename: 'src/virtual.js',
      errors: [reexport('baz')],
    },
    {
      code: "import {foo} from 'foo'; /** This comment belongs to the import. */\nexport const baz = foo;",
      filename: 'src/virtual.js',
      errors: [reexport('baz')],
    },
    {
      code: "import {foo} from 'foo';\n\n/**\n * This comment is detached from the declaration.\n */\n\nexport const baz = foo;",
      filename: 'src/virtual.js',
      errors: [reexport('baz')],
    },
  ],
});

// Type Script
tester.run('prefer-export-from', null as never, {
  valid: [
    {
      code: "import {useDispatch as reduxUseDispatch} from 'react-redux'\ntype MyDispatchType = Dispatch<MyActions>\n\nexport const useDispatch: () => DispatchAllActions = reduxUseDispatch",
    },
    {
      code: "import {foo} from 'foo';\ntype Baz = string;\n\n/**\n * Add documentation to the re-export.\n */\nexport const baz = foo;",
    },
    {
      code: 'import React from "react";\nimport React from "react";\nexport {React}',
    },
    {
      code: 'type AceEditor = import("ace-builds").Ace.Editor;\nimport AceEditor from "./advanced-editor";',
    },
    {
      code: 'type AceEditor = import("ace-builds").Ace.Editor;\nimport AceEditor from "./advanced-editor";\nexport {AceEditor};',
    },
    {
      code: 'import AceEditor from "./advanced-editor";\ntype AceEditor = import("ace-builds").Ace.Editor;\nexport {AceEditor};',
    },
    {
      code: 'export type { bar, foo } from "foo";',
    },
  ],
  invalid: [
    {
      code: 'import { foo } from "foo";\nexport { foo };\nexport type { bar } from "foo";',
      errors: [reexport('foo')],
    },
    {
      code: 'import { foo } from "foo";\nexport { foo };\nexport { type bar } from "foo";',
      errors: [reexport('foo')],
    },
    {
      code: 'import { foo } from \'foo\';\nexport { foo };\nexport type { bar } from "foo";\nexport { baz } from "foo";',
      errors: [reexport('foo')],
    },
    {
      code: 'import { foo } from \'foo\';\nexport { foo };\nexport { type bar } from "foo";\nexport { baz } from "foo";',
      errors: [reexport('foo')],
    },
    {
      code: 'import type { foo } from "foo";\nexport type { foo };\nexport type { bar } from "foo";',
      errors: [reexport('foo')],
    },
    {
      code: 'import { foo } from \'foo\';\nexport { foo };\nexport { baz } from "foo";\nexport { type bar } from "foo";',
      errors: [reexport('foo')],
    },
    {
      code: 'import type { foo } from "foo";\nexport type { foo };\nexport { type bar } from "foo";',
      errors: [reexport('foo')],
    },
    {
      code: 'import type { foo } from \'foo\';\nexport type { foo };\nexport type { bar } from "foo";\nexport { baz } from "foo";',
      errors: [reexport('foo')],
    },
    {
      code: 'import type { foo } from \'foo\';\nexport type { foo };\nexport { baz } from "foo";\nexport type { bar } from "foo";',
      errors: [reexport('foo')],
    },
    {
      code: 'import type { foo } from \'foo\';\nexport type { foo };\nexport { type bar } from "foo";\nexport { baz } from "foo";',
      errors: [reexport('foo')],
    },
    {
      code: 'import type { foo } from \'foo\';\nexport type { foo };\nexport { baz } from "foo";\nexport { type bar } from "foo";',
      errors: [reexport('foo')],
    },
    {
      code: "import { type foo } from 'foo';\nexport type { foo };",
      errors: [reexport('foo')],
    },
    {
      code: "import { foo } from 'foo';\nexport type { foo };",
      errors: [reexport('foo')],
    },
    {
      code: "import type { foo } from 'foo';\nexport { type foo };",
      errors: [reexport('foo')],
    },
    {
      code: 'import type foo from "foo";\nexport default foo',
      errors: [reexport('default')],
    },
    {
      code: "import {type foo} from 'foo';\nexport {type foo as bar};",
      errors: [reexport('bar')],
    },
    {
      code: "import {type foo} from 'foo';\nexport {type foo as bar};\nexport {type bar} from 'foo';",
      errors: [reexport('bar')],
    },
    {
      code: "import {type foo as bar} from 'foo';\nexport {type bar as baz};",
      errors: [reexport('baz')],
    },
    {
      code: "import {type foo as foo} from 'foo';\nexport {type foo as bar};",
      errors: [reexport('bar')],
    },
    {
      code: "import {type foo as bar} from 'foo';\nexport {type bar as bar};",
      errors: [reexport('bar')],
    },
    {
      code: "import {type foo as bar} from 'foo';\nexport {type bar as foo};",
      errors: [reexport('foo')],
    },
    // The Go suite retains the four legacy `assert` cases as skips (TS2880).
    {
      code: "import type * as X from 'foo';\nexport { X };",
      errors: [reexport('X')],
    },
    {
      code: "import * as X from 'foo';\nexport type { X };",
      errors: [reexport('X')],
    },
    {
      code: "import type * as X from 'foo';\nexport type { X };",
      errors: [reexport('X')],
    },
    {
      code: "import * as X from 'foo';\nexport { X };",
      errors: [reexport('X')],
    },
  ],
});

// Import Attributes
tester.run('prefer-export-from', null as never, {
  valid: [],
  invalid: [
    {
      code: "import json from './foo.json' with { type: 'json' };\nexport default json;",
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: "import * as json from './foo.json' with { type: 'json' };\nexport {json};",
      filename: 'src/virtual.js',
      errors: [reexport('json')],
    },
    {
      code: "import {foo} from './foo.json' with { type: 'unknown' };\nexport {foo};\nexport {bar} from './foo.json';",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: "import {foo} from './foo.json';\nexport {foo};\nexport {bar} from './foo.json' with { type: 'unknown' };",
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
  ],
});

// Check Used Variables
tester.run('prefer-export-from', null as never, {
  valid: [
    {
      code: "import defaultExport from 'foo';\nuse(defaultExport);\nexport default defaultExport;",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
    },
    {
      code: "import defaultExport from 'foo';\nuse(defaultExport);\nexport {defaultExport};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
    },
    {
      code: "import {named} from 'foo';\nuse(named);\nexport {named};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
    },
    {
      code: "import {named} from 'foo';\nuse(named);\nexport default named;",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
    },
    {
      code: "import * as namespace from 'foo';\nuse(namespace);\nexport {namespace};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
    },
    {
      code: "import * as namespace from 'foo';\nuse(namespace);\nexport default namespace;",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
    },
    {
      code: "import * as namespace from 'foo';\nexport {namespace as default};\nexport {namespace as named};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
    },
    {
      code: "import * as namespace from 'foo';\nexport default namespace;\nexport {namespace as named};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
    },
    {
      code: "import defaultExport, {named} from 'foo';\nuse(defaultExport);\nexport {named};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
    },
    {
      code: "import defaultExport, {named} from 'foo';\nuse(named);\nexport {defaultExport};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
    },
    {
      code: "import {named1, named2} from 'foo';\nuse(named1);\nexport {named2};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
    },
    {
      code: "import defaultExport, {named1, named2} from 'foo';\nuse(defaultExport);\nexport {named1, named2};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
    },
    {
      code: "import defaultExport, {named1, named2} from 'foo';\nuse(named1);\nexport {defaultExport, named2};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
    },
  ],
  invalid: [
    {
      code: "import defaultExport from 'foo';\nexport {defaultExport as default};\nexport {defaultExport as named};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
      errors: [reexport('default'), reexport('named')],
    },
    {
      code: "import {named} from 'foo';\nexport {named as default};\nexport {named as named};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
      errors: [reexport('default'), reexport('named')],
    },
    {
      code: "import {named} from 'foo';\nexport default named;\nexport {named as named};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
      errors: [reexport('default'), reexport('named')],
    },
    {
      code: "import defaultExport, {named} from 'foo';\nexport default defaultExport;\nexport {named};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
      errors: [reexport('default'), reexport('named')],
    },
    {
      code: "import defaultExport, {named} from 'foo';\nexport {defaultExport as default, named};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
      errors: [reexport('default'), reexport('named')],
    },
    {
      code: "import defaultExport from 'foo';\nexport const variable = defaultExport;",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
      errors: [reexport('variable')],
    },
    {
      code: "import {notUsedNotExported, exported} from 'foo';\nexport {exported};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
      errors: [reexport('exported')],
    },
    {
      code: "import {documented, other, unused} from 'foo';\n\n/**\n * Preserve this documentation.\n */\nexport const documentedAlias = documented;\nexport {other};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
      errors: [reexport('other')],
    },
  ],
});

// Line Endings
tester.run('prefer-export-from', null as never, {
  valid: [],
  invalid: [
    {
      code: 'import {foo} from "foo";\r\nexport {foo};\r\n',
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: 'import * as namespace from "foo";\r\nexport {namespace};\r\n',
      filename: 'src/virtual.js',
      errors: [reexport('namespace')],
    },
  ],
});

// Comments
tester.run('prefer-export-from', null as never, {
  valid: [],
  invalid: [
    {
      code: 'import {a} from "foo";\nexport {/* c1 */ a /* c2 */};',
      filename: 'src/virtual.js',
      errors: [reexport('a')],
    },
    {
      code: 'import {a /* c */} from "foo";\nexport {a};',
      filename: 'src/virtual.js',
      errors: [reexport('a')],
    },
    {
      code: 'import {a} from "foo";\n// c\nexport {a};',
      filename: 'src/virtual.js',
      errors: [reexport('a')],
    },
  ],
});

// Commented Edits
tester.run('prefer-export-from', null as never, {
  valid: [],
  invalid: [
    {
      code: 'import {a, b} from "foo";\nexport {/* c */ a /* c */};\nuse(b);',
      filename: 'src/virtual.js',
      errors: [reexport('a')],
    },
    {
      code: 'import {a, b} from "foo";\nexport {/* c */ a};',
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
      errors: [reexport('a')],
    },
  ],
});

// Semicolons
tester.run('prefer-export-from', null as never, {
  valid: [],
  invalid: [
    {
      code: 'import {foo} from "foo";\nlog()\nexport {foo};\n[1].forEach(log);',
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
    {
      code: 'import {foo} from "foo";\nlog()\nexport default foo;\n[1].forEach(log);',
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: 'import {foo} from "foo";\nlog()\nexport const bar = foo;\n[1].forEach(log);',
      filename: 'src/virtual.js',
      errors: [reexport('bar')],
    },
    {
      code: 'log()\nimport {foo} from "foo";\n[1].forEach(log);\nexport {foo};',
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
  ],
});

// Adjacent Removals
tester.run('prefer-export-from', null as never, {
  valid: [],
  invalid: [
    {
      code: 'log()\nimport {foo} from "foo";\nexport {foo};\n[1].forEach(log);',
      filename: 'src/virtual.js',
      errors: [reexport('foo')],
    },
  ],
});

// Documentation
tester.run('prefer-export-from', null as never, {
  valid: [
    {
      code: "export {default} from './foo.js';",
      filename: 'src/virtual.js',
    },
    {
      code: "export {named} from './foo.js';",
      filename: 'src/virtual.js',
    },
    {
      code: "export * as namespace from './foo.js';",
      filename: 'src/virtual.js',
    },
    {
      code: "export {\n\tdefault,\n\tdefault as renamedDefault,\n\tnamed,\n\tnamed as renamedNamed,\n} from './foo.js';",
      filename: 'src/virtual.js',
    },
    {
      code: "// There is no substitution\nimport * as namespace from './foo.js';\nexport default namespace;",
      filename: 'src/virtual.js',
    },
    {
      code: "import {named1, named2} from './foo.js';\n\nuse(named1);\n\nexport {named1, named2};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: false }],
    },
  ],
  invalid: [
    {
      code: "import defaultExport from './foo.js';\nexport default defaultExport;",
      filename: 'src/virtual.js',
      errors: [reexport('default')],
    },
    {
      code: "import {named} from './foo.js';\nexport {named};",
      filename: 'src/virtual.js',
      errors: [reexport('named')],
    },
    {
      code: "import * as namespace from './foo.js';\nexport {namespace};",
      filename: 'src/virtual.js',
      errors: [reexport('namespace')],
    },
    {
      code: "import defaultExport, {named} from './foo.js';\nexport default defaultExport;\nexport {\n\tdefaultExport as renamedDefault,\n\tnamed,\n\tnamed as renamedNamed,\n};",
      filename: 'src/virtual.js',
      errors: [
        reexport('default'),
        reexport('renamedDefault'),
        reexport('named'),
        reexport('renamedNamed'),
      ],
    },
    {
      code: "import {named1, named2} from './foo.js';\n\nuse(named1);\n\nexport {named1, named2};",
      filename: 'src/virtual.js',
      options: [{ checkUsedVariables: true }],
      errors: [reexport('named1'), reexport('named2')],
    },
  ],
});
