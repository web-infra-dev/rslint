package prefer_export_from_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_export_from"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Inputs and expected results from eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-export-from.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-export-from.md
func TestPreferExportFromUpstream(t *testing.T) {
	upstreamRule := &prefer_export_from.PreferExportFromRule
	t.Run("CapturedBindings", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, upstreamRule,
			nil,
			[]rule_tester.InvalidTestCase{
				{FileName: "test.js", Code: `import value from "source";
export default (value);`,
					Output: []string{`

export {default} from "source";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 1, 2, 24)},
				},
				{FileName: "test.js", Code: `import value from "source";
export const snapshot = (value);`,
					Output: []string{`

export {default as snapshot} from "source";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("snapshot", 2, 1, 2, 33)},
				},
				{FileName: "test.js", Code: `import {value} from "source";
export default /* keep */ value;`,
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 1, 2, 33)},
				},
				{FileName: "test.js", Code: `import {value} from "source";
export const snapshot = /* keep */ value;`,
					Errors: []rule_tester.InvalidTestCaseError{reexportError("snapshot", 2, 1, 2, 42)},
				},
			},
		)
	})
	t.Run("JavaScript", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, upstreamRule,
			[]rule_tester.ValidTestCase{
				{FileName: "test.js", Code: "import \"foo\";"},
				{FileName: "test.js", Code: "import {} from \"foo\";"},
				{FileName: "test.js", Code: "import * as namespace from \"foo\";"},
				{FileName: "test.js", Code: "import defaultExport from \"foo\";"},
				{FileName: "test.js", Code: "import {named} from \"foo\";"},
				{FileName: "test.js", Code: `const named = import(foo);
export {named};`},
				{FileName: "test.js", Code: "export * from \"foo\";"},
				{FileName: "test.js", Code: "export {default} from \"foo\";"},
				{FileName: "test.js", Code: "export {named} from \"foo\";"},
				{FileName: "test.js", Code: `const defaultExport = require('foo');
export default defaultExport;`},
				{FileName: "test.js", Code: `import {foo} from 'foo';

/**
 * Add documentation to the re-export.
 */
export const baz = foo;`},
				{FileName: "test.js", Code: `import {foo} from 'foo';
/** Add documentation to the re-export. */ export const baz = foo;`},
				{FileName: "test.js", Code: `import defaultExport from 'foo';
export var variable = defaultExport;`},
				{FileName: "test.js", Code: `import defaultExport from 'foo';
export let variable = defaultExport;`},
				{FileName: "test.js", Code: `import defaultExport from 'foo';
export const variable = defaultExport;
use(variable);`},
				{FileName: "test.js", Code: `import defaultExport from 'foo';
export let variable = defaultExport;
variable = 1;`},
				{FileName: "test.js", Code: `import * as namespace from 'foo';
export default namespace;`},
				{FileName: "test.js", Code: `import * as namespace from 'foo';
export {namespace as default};`},
				{FileName: "test.js", Code: `import defaultExport from 'foo';
const variable = defaultExport;
export {variable}`},
				{FileName: "test.js", Code: `import defaultExport from 'foo';
export const {variable} = {variable: defaultExport};`},
			},
			[]rule_tester.InvalidTestCase{
				{FileName: "test.js", Code: `import defaultExport from 'foo';
export default defaultExport;`,
					Output: []string{`

export {default} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 1, 2, 30)},
				},
				{FileName: "test.js", Code: `import defaultExport from 'foo';
export {defaultExport as default};`,
					Output: []string{`

export {default} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 9, 2, 33)},
				},
				{FileName: "test.js", Code: `import defaultExport from 'foo';
export {defaultExport as named};`,
					Output: []string{`

export {default as named} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("named", 2, 9, 2, 31)},
				},
				{FileName: "test.js", Code: `import defaultExport from 'foo';
export const variable = defaultExport;`,
					Output: []string{`

export {default as variable} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("variable", 2, 1, 2, 39)},
				},
				{FileName: "test.js", Code: `import {default as defaultExport} from 'foo';
export default defaultExport;`,
					Output: []string{`

export {default} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 1, 2, 30)},
				},
				{FileName: "test.js", Code: `import {default as defaultExport} from 'foo';
export {defaultExport as default};`,
					Output: []string{`

export {default} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 9, 2, 33)},
				},
				{FileName: "test.js", Code: `import {default as defaultExport} from 'foo';
export {defaultExport as named};`,
					Output: []string{`

export {default as named} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("named", 2, 9, 2, 31)},
				},
				{FileName: "test.js", Code: `import defaultExport from 'foo';
defaultExport.bar = 1;
export {defaultExport as named};
export {defaultExport as default};
export const variable = defaultExport;`,
					Output: []string{`import defaultExport from 'foo';
defaultExport.bar = 1;

export {defaultExport as default};
export const variable = defaultExport;
export {default as named} from 'foo';`, `import defaultExport from 'foo';
defaultExport.bar = 1;


export const variable = defaultExport;
export {default as named, default} from 'foo';`, `import defaultExport from 'foo';
defaultExport.bar = 1;



export {default as named, default, default as variable} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("named", 3, 9, 3, 31), reexportError("default", 4, 9, 4, 33), reexportError("variable", 5, 1, 5, 39)},
				},
				{FileName: "test.js", Code: `import {named} from 'foo';
export default named;`,
					Output: []string{`

export {named as default} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 1, 2, 22)},
				},
				{FileName: "test.js", Code: `import {named} from 'foo';
export {named as default};`,
					Output: []string{`

export {named as default} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 9, 2, 25)},
				},
				{FileName: "test.js", Code: `import {named} from 'foo';
export {named as named};`,
					Output: []string{`

export {named} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("named", 2, 9, 2, 23)},
				},
				{FileName: "test.js", Code: `import {named} from 'foo';
export {named as renamed};`,
					Output: []string{`

export {named as renamed} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("renamed", 2, 9, 2, 25)},
				},
				{FileName: "test.js", Code: `import {named} from 'foo';
export const variable = named;`,
					Output: []string{`

export {named as variable} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("variable", 2, 1, 2, 31)},
				},
				{FileName: "test.js", Code: `import {named} from 'foo';
named.bar = 1;
export {named as named};
export {named as default};
export const variable = named;`,
					Output: []string{`import {named} from 'foo';
named.bar = 1;

export {named as default};
export const variable = named;
export {named} from 'foo';`, `import {named} from 'foo';
named.bar = 1;


export const variable = named;
export {named, named as default} from 'foo';`, `import {named} from 'foo';
named.bar = 1;



export {named, named as default, named as variable} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("named", 3, 9, 3, 23), reexportError("default", 4, 9, 4, 25), reexportError("variable", 5, 1, 5, 31)},
				},
				{FileName: "test.js", Code: `import * as namespace from 'foo';
export {namespace as namespace};`,
					Output: []string{`

export * as namespace from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("namespace", 2, 9, 2, 31)},
				},
				{FileName: "test.js", Code: `import * as namespace from 'foo';
export {namespace as renamed};`,
					Output: []string{`

export * as renamed from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("renamed", 2, 9, 2, 29)},
				},
				{FileName: "test.js", Code: `import * as namespace from 'foo';
export const variable = namespace;`,
					Output: []string{`

export * as variable from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("variable", 2, 1, 2, 35)},
				},
				{FileName: "test.js", Code: `import * as namespace from 'foo';
namespace.bar = 1;
export {namespace as named};
export {namespace as default};
export const variable = namespace;`,
					Output: []string{`import * as namespace from 'foo';
namespace.bar = 1;

export {namespace as default};
export const variable = namespace;
export * as named from 'foo';`, `import * as namespace from 'foo';
namespace.bar = 1;

export {namespace as default};

export * as named from 'foo';
export * as variable from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("named", 3, 9, 3, 27), reexportError("variable", 5, 1, 5, 35)},
				},
				{FileName: "test.js", Code: `import {named1, named2} from 'foo';
export {named1};`,
					Output: []string{`import { named2} from 'foo';

export {named1} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("named1", 2, 9, 2, 15)},
				},
				{FileName: "test.js", Code: `import defaultExport, {named} from 'foo';
export {defaultExport};`,
					Output: []string{`import  {named} from 'foo';

export {default as defaultExport} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("defaultExport", 2, 9, 2, 22)},
				},
				{FileName: "test.js", Code: `import defaultExport, {named} from 'foo';
export {named};`,
					Output: []string{`import defaultExport from 'foo';

export {named} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("named", 2, 9, 2, 14)},
				},
				{FileName: "test.js", Code: `import defaultExport, * as namespace from 'foo';
export {defaultExport};`,
					Output: []string{`import  * as namespace from 'foo';

export {default as defaultExport} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("defaultExport", 2, 9, 2, 22)},
				},
				{FileName: "test.js", Code: `import * as foo from 'foo';
export {foo};
export * as bar from 'foo';`,
					Output: []string{`

export * as bar from 'foo';
export * as foo from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{FileName: "test.js", Code: `import * as foo from 'foo';
export {foo};
export {bar} from 'foo';`,
					Output: []string{`

export {bar} from 'foo';
export * as foo from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{FileName: "test.js", Code: `import * as foo from 'foo';
export {foo};
export {} from 'foo';`,
					Output: []string{`

export {} from 'foo';
export * as foo from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{FileName: "test.js", Code: `import * as foo from 'foo';
export {foo};
export * from 'foo';`,
					Output: []string{`

export * from 'foo';
export * as foo from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{FileName: "test.js", Code: `import foo from 'foo';
export {foo};
export * as bar from 'foo';`,
					Output: []string{`

export * as bar from 'foo';
export {default as foo} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{FileName: "test.js", Code: `import foo from 'foo';
export {foo};
export {bar} from 'foo';`,
					Output: []string{`

export {bar, default as foo} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{FileName: "test.js", Code: `import foo from 'foo';
export {foo};
export {bar,} from 'foo';`,
					Output: []string{`

export {bar, default as foo,} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{FileName: "test.js", Code: `import foo from 'foo';
export {foo};
export {} from 'foo';`,
					Output: []string{`

export {default as foo} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{FileName: "test.js", Code: `import foo from 'foo';
export {foo};
export * from 'foo';`,
					Output: []string{`

export * from 'foo';
export {default as foo} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{FileName: "test.js", Code: `import {named1, named2} from 'foo';
export {named1, named2};`,
					Output: []string{`import { named2} from 'foo';
export { named2};
export {named1} from 'foo';`, `

export {named1, named2} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("named1", 2, 9, 2, 15), reexportError("named2", 2, 17, 2, 23)},
				},
				{FileName: "test.js", Code: `import {named} from 'foo';
export {named as default, named};`,
					Output: []string{`import {named} from 'foo';
export { named};
export {named as default} from 'foo';`, `

export {named as default, named} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 9, 2, 25), reexportError("named", 2, 27, 2, 32)},
				},
				{FileName: "test.js", Code: `import {named, named as renamed} from 'foo';
export {named, renamed};`,
					Output: []string{`import { named as renamed} from 'foo';
export { renamed};
export {named} from 'foo';`, `

export {named, named as renamed} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("named", 2, 9, 2, 14), reexportError("renamed", 2, 16, 2, 23)},
				},
				{FileName: "test.js", Code: `import defaultExport, {named1, named2} from 'foo';
export {named1 as default};
export {named2};
export {defaultExport};`,
					Output: []string{`import  {named1, named2} from 'foo';
export {named1 as default};
export {named2};

export {default as defaultExport} from 'foo';`, `import  { named2} from 'foo';

export {named2};

export {default as defaultExport, named1 as default} from 'foo';`, `



export {default as defaultExport, named1 as default, named2} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 9, 2, 26), reexportError("named2", 3, 9, 3, 15), reexportError("defaultExport", 4, 9, 4, 22)},
				},
				{FileName: "test.js", Code: `import * as foo from 'foo';
import * as bar from 'foo';
export {foo, bar};`,
					Output: []string{`
import * as bar from 'foo';
export { bar};
export * as foo from 'foo';`, `


export * as foo from 'foo';
export * as bar from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 3, 9, 3, 12), reexportError("bar", 3, 14, 3, 17)},
				},
				{FileName: "test.js", Code: `import * as foo from 'foo';
export {foo, foo as bar};`,
					Output: []string{`import * as foo from 'foo';
export { foo as bar};
export * as foo from 'foo';`, `

export * as foo from 'foo';
export * as bar from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12), reexportError("bar", 2, 14, 2, 24)},
				},
				{FileName: "test.js", Code: `import defaultExport from 'foo';
export * from 'foo';
export default defaultExport;`,
					Output: []string{`
export * from 'foo';

export {default} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 3, 1, 3, 30)},
				},
				{FileName: "test.js", Code: `import defaultExport from 'foo';
export {named} from 'foo';
export * from 'foo';
export default defaultExport;`,
					Output: []string{`
export {named, default} from 'foo';
export * from 'foo';
`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 4, 1, 4, 30)},
				},
				{FileName: "test.js", Code: `import defaultExport from './foo.js';
export {named} from './foo.js';
export default defaultExport;`,
					Output: []string{`
export {named, default} from './foo.js';
`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 3, 1, 3, 30)},
				},
				{FileName: "test.js", Code: `import defaultExport from './foo.js';
export {named} from './foo.js?query';
export default defaultExport;`,
					// Keep the dependency requests in their original order.
					Output: []string{"export {default} from './foo.js';\nexport {named} from './foo.js?query';\n"},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 3, 1, 3, 30)},
				},
				{FileName: "test.js", Code: `import * as namespace from 'foo';
export default namespace;
export {namespace};`,
					Output: []string{`import * as namespace from 'foo';
export default namespace;

export * as namespace from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("namespace", 3, 9, 3, 18)},
				},
				{FileName: "test.js", Code: `import * as namespace from 'foo';
export {namespace};
export default namespace;`,
					Output: []string{`import * as namespace from 'foo';

export default namespace;
export * as namespace from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("namespace", 2, 9, 2, 18)},
				},
				{FileName: "test.js", Code: `import {'foo' as foo} from 'foo';
export default foo;`,
					Output: []string{`

export {'foo' as default} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 1, 2, 20)},
				},
				{FileName: "test.js", Code: `import {'foo' as foo} from 'foo';
export {foo};`,
					Output: []string{`

export {'foo' as foo} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{FileName: "test.js", Code: `import {'foo' as foo} from 'foo';
export const bar = foo;`,
					Output: []string{`

export {'foo' as bar} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("bar", 2, 1, 2, 24)},
				},
				{FileName: "test.js", Code: `import {'foo' as foo} from 'foo';
export {foo as 'foo'};`,
					Output: []string{`

export {'foo'} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("'foo'", 2, 9, 2, 21)},
				},
				{FileName: "test.js", Code: `import {foo} from 'foo';
export {foo as 'bar'};`,
					Output: []string{`

export {foo as 'bar'} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("'bar'", 2, 9, 2, 21)},
				},
				{FileName: "test.js", Code: `import {'foo' as foo} from 'foo';
export {foo as "foo"};`,
					Output: []string{`

export {"foo"} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("\"foo\"", 2, 9, 2, 21)},
				},
				{FileName: "test.js", Code: `import {'fo\u{20}o' as foo} from 'foo';
export {foo as "fo o"};`,
					Output: []string{`

export {"fo o"} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("\"fo o\"", 2, 9, 2, 22)},
				},
				{FileName: "test.js", Code: `import {'fo\no' as foo} from 'foo';
export {foo as "fo\u000ao"};`,
					Output: []string{`

export {"fo\u000ao"} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("\"fo\\u000ao\"", 2, 9, 2, 27)},
				},
				{FileName: "test.js", Code: `import {'default' as foo} from 'foo';
export {foo};`,
					Output: []string{`

export {'default' as foo} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{FileName: "test.js", Code: `import {'default' as foo} from 'foo';
export default foo;`,
					Output: []string{`

export {'default' as default} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 1, 2, 20)},
				},
				{FileName: "test.js", Code: `import {'*' as foo} from 'foo';
export {foo};`,
					Output: []string{`

export {'*' as foo} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{FileName: "test.js", Code: `import {named} from 'foo';
use(named);
export {named};`,
					Output: []string{`import {named} from 'foo';
use(named);

export {named} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("named", 3, 9, 3, 14)},
				},
				{FileName: "test.js", Code: `import {foo} from 'foo';
// This is not JSDoc.
export const baz = foo;`,
					Output: []string{`
// This is not JSDoc.

export {foo as baz} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("baz", 3, 1, 3, 24)},
				},
				{FileName: "test.js", Code: `import {foo} from 'foo';

/**
 * Documentation for the direct export.
 */
export {foo};`,
					Output: []string{`

/**
 * Documentation for the direct export.
 */

export {foo} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 6, 9, 6, 12)},
				},
				{FileName: "test.js", Code: `import {foo} from 'foo';
/*
 * This is not JSDoc.
 */
export const baz = foo;`,
					Output: []string{`
/*
 * This is not JSDoc.
 */

export {foo as baz} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("baz", 5, 1, 5, 24)},
				},
				{FileName: "test.js", Code: `import {foo} from 'foo';
/*** This is not JSDoc. */
export const baz = foo;`,
					Output: []string{`
/*** This is not JSDoc. */

export {foo as baz} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("baz", 3, 1, 3, 24)},
				},
				{FileName: "test.js", Code: `import {foo} from 'foo'; /** This comment belongs to the import. */
export const baz = foo;`,
					Output: []string{` /** This comment belongs to the import. */

export {foo as baz} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("baz", 2, 1, 2, 24)},
				},
				{FileName: "test.js", Code: `import {foo} from 'foo';

/**
 * This comment is detached from the declaration.
 */

export const baz = foo;`,
					Output: []string{`

/**
 * This comment is detached from the declaration.
 */


export {foo as baz} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("baz", 7, 1, 7, 24)},
				},
			},
		)
	})
	t.Run("TypeScript", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, upstreamRule,
			[]rule_tester.ValidTestCase{
				{Code: `import {useDispatch as reduxUseDispatch} from 'react-redux'
type MyDispatchType = Dispatch<MyActions>

export const useDispatch: () => DispatchAllActions = reduxUseDispatch`},
				{Code: `import {foo} from 'foo';
type Baz = string;

/**
 * Add documentation to the re-export.
 */
export const baz = foo;`},
				{Code: `import React from "react";
import React from "react";
export {React}`},
				{Code: `type AceEditor = import("ace-builds").Ace.Editor;
import AceEditor from "./advanced-editor";`},
				{Code: `type AceEditor = import("ace-builds").Ace.Editor;
import AceEditor from "./advanced-editor";
export {AceEditor};`},
				{Code: `import AceEditor from "./advanced-editor";
type AceEditor = import("ace-builds").Ace.Editor;
export {AceEditor};`},
				{Code: "export type { bar, foo } from \"foo\";"},
			},
			[]rule_tester.InvalidTestCase{
				{Code: `import { foo } from "foo";
export { foo };
export type { bar } from "foo";`,
					Output: []string{`

export type { bar } from "foo";
export {foo} from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 10, 2, 13)},
				},
				{Code: `import { foo } from "foo";
export { foo };
export { type bar } from "foo";`,
					Output: []string{`

export { type bar, foo } from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 10, 2, 13)},
				},
				{Code: `import { foo } from 'foo';
export { foo };
export type { bar } from "foo";
export { baz } from "foo";`,
					Output: []string{`

export type { bar } from "foo";
export { baz, foo } from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 10, 2, 13)},
				},
				{Code: `import { foo } from 'foo';
export { foo };
export { type bar } from "foo";
export { baz } from "foo";`,
					Output: []string{`

export { type bar, foo } from "foo";
export { baz } from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 10, 2, 13)},
				},
				{Code: `import type { foo } from "foo";
export type { foo };
export type { bar } from "foo";`,
					Output: []string{`

export type { bar, foo } from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 15, 2, 18)},
				},
				{Code: `import { foo } from 'foo';
export { foo };
export { baz } from "foo";
export { type bar } from "foo";`,
					Output: []string{`

export { baz, foo } from "foo";
export { type bar } from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 10, 2, 13)},
				},
				{Code: `import type { foo } from "foo";
export type { foo };
export { type bar } from "foo";`,
					Output: []string{`

export { type bar, type foo } from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 15, 2, 18)},
				},
				{Code: `import type { foo } from 'foo';
export type { foo };
export type { bar } from "foo";
export { baz } from "foo";`,
					Output: []string{`

export type { bar, foo } from "foo";
export { baz } from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 15, 2, 18)},
				},
				{Code: `import type { foo } from 'foo';
export type { foo };
export { baz } from "foo";
export type { bar } from "foo";`,
					Output: []string{`

export { baz } from "foo";
export type { bar, foo } from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 15, 2, 18)},
				},
				{Code: `import type { foo } from 'foo';
export type { foo };
export { type bar } from "foo";
export { baz } from "foo";`,
					Output: []string{`

export { type bar, type foo } from "foo";
export { baz } from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 15, 2, 18)},
				},
				{Code: `import type { foo } from 'foo';
export type { foo };
export { baz } from "foo";
export { type bar } from "foo";`,
					Output: []string{`

export { baz, type foo } from "foo";
export { type bar } from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 15, 2, 18)},
				},
				{Code: `import { type foo } from 'foo';
export type { foo };`,
					Output: []string{`

export {type foo} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 15, 2, 18)},
				},
				{Code: `import { foo } from 'foo';
export type { foo };`,
					Output: []string{`import { foo } from 'foo';

export {type foo} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 15, 2, 18)},
				},
				{Code: `import type { foo } from 'foo';
export { type foo };`,
					Output: []string{`

export {type foo} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 10, 2, 18)},
				},
				{Code: `import type foo from "foo";
export default foo`,
					Output: []string{`

export {type default} from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 1, 2, 19)},
				},
				{Code: `import {type foo} from 'foo';
export {type foo as bar};`,
					Output: []string{`

export {type foo as bar} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("bar", 2, 9, 2, 24)},
				},
				{Code: `import {type foo} from 'foo';
export {type foo as bar};
export {type bar} from 'foo';`,
					Output: []string{`

export {type bar, type foo as bar} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("bar", 2, 9, 2, 24)},
				},
				{Code: `import {type foo as bar} from 'foo';
export {type bar as baz};`,
					Output: []string{`

export {type foo as baz} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("baz", 2, 9, 2, 24)},
				},
				{Code: `import {type foo as foo} from 'foo';
export {type foo as bar};`,
					Output: []string{`

export {type foo as bar} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("bar", 2, 9, 2, 24)},
				},
				{Code: `import {type foo as bar} from 'foo';
export {type bar as bar};`,
					Output: []string{`

export {type foo as bar} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("bar", 2, 9, 2, 24)},
				},
				{Code: `import {type foo as bar} from 'foo';
export {type bar as foo};`,
					Output: []string{`

export {type foo} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 24)},
				},
				// tsgo rejects the legacy assertion syntax with TS2880. The
				// ImportAttributes group covers its supported `with` equivalent.
				{Skip: true, Code: `import json from './foo.json' assert { type: 'json' };
export default json;`,
					Output: []string{`

export {default} from './foo.json' assert { type: 'json' };`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 1, 2, 21)},
				},
				{Skip: true, Code: `import * as json from './foo.json' assert { type: 'json' };
export {json};`,
					Output: []string{`

export * as json from './foo.json' assert { type: 'json' };`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("json", 2, 9, 2, 13)},
				},
				{Skip: true, Code: `import {foo} from './foo.json' assert { type: 'unknown' };
export {foo};
export {bar} from './foo.json';`,
					Output: []string{`

export {bar, foo} from './foo.json';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{Skip: true, Code: `import {foo} from './foo.json';
export {foo};
export {bar} from './foo.json' assert { type: 'unknown' };`,
					Output: []string{`

export {bar, foo} from './foo.json' assert { type: 'unknown' };`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{Code: `import type * as X from 'foo';
export { X };`,
					Output: []string{`

export type * as X from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("X", 2, 10, 2, 11)},
				},
				{Code: `import * as X from 'foo';
export type { X };`,
					Output: []string{`import * as X from 'foo';

export type * as X from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("X", 2, 15, 2, 16)},
				},
				{Code: `import type * as X from 'foo';
export type { X };`,
					Output: []string{`

export type * as X from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("X", 2, 15, 2, 16)},
				},
				{Code: `import * as X from 'foo';
export { X };`,
					Output: []string{`

export * as X from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("X", 2, 10, 2, 11)},
				},
			},
		)
	})
	t.Run("ImportAttributes", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, upstreamRule,
			nil,
			[]rule_tester.InvalidTestCase{
				{FileName: "test.js", Code: `import json from './foo.json' with { type: 'json' };
export default json;`,
					Output: []string{`

export {default} from './foo.json' with { type: 'json' };`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 1, 2, 21)},
				},
				{FileName: "test.js", Code: `import * as json from './foo.json' with { type: 'json' };
export {json};`,
					Output: []string{`

export * as json from './foo.json' with { type: 'json' };`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("json", 2, 9, 2, 13)},
				},
				{FileName: "test.js", Code: `import {foo} from './foo.json' with { type: 'unknown' };
export {foo};
export {bar} from './foo.json';`,
					// Preserve each request's attributes and position.
					Output: []string{"export {foo} from './foo.json' with { type: 'unknown' };\n\nexport {bar} from './foo.json';"},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{FileName: "test.js", Code: `import {foo} from './foo.json';
export {foo};
export {bar} from './foo.json' with { type: 'unknown' };`,
					Output: []string{"export {foo} from './foo.json';\n\nexport {bar} from './foo.json' with { type: 'unknown' };"},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
			},
		)
	})
	t.Run("CheckUsedVariables", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, upstreamRule,
			[]rule_tester.ValidTestCase{
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import defaultExport from 'foo';
use(defaultExport);
export default defaultExport;`},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import defaultExport from 'foo';
use(defaultExport);
export {defaultExport};`},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import {named} from 'foo';
use(named);
export {named};`},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import {named} from 'foo';
use(named);
export default named;`},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import * as namespace from 'foo';
use(namespace);
export {namespace};`},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import * as namespace from 'foo';
use(namespace);
export default namespace;`},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import * as namespace from 'foo';
export {namespace as default};
export {namespace as named};`},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import * as namespace from 'foo';
export default namespace;
export {namespace as named};`},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import defaultExport, {named} from 'foo';
use(defaultExport);
export {named};`},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import defaultExport, {named} from 'foo';
use(named);
export {defaultExport};`},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import {named1, named2} from 'foo';
use(named1);
export {named2};`},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import defaultExport, {named1, named2} from 'foo';
use(defaultExport);
export {named1, named2};`},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import defaultExport, {named1, named2} from 'foo';
use(named1);
export {defaultExport, named2};`},
			},
			[]rule_tester.InvalidTestCase{
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import defaultExport from 'foo';
export {defaultExport as default};
export {defaultExport as named};`,
					Output: []string{`import defaultExport from 'foo';

export {defaultExport as named};
export {default} from 'foo';`, `


export {default, default as named} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 9, 2, 33), reexportError("named", 3, 9, 3, 31)},
				},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import {named} from 'foo';
export {named as default};
export {named as named};`,
					Output: []string{`import {named} from 'foo';

export {named as named};
export {named as default} from 'foo';`, `


export {named as default, named} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 9, 2, 25), reexportError("named", 3, 9, 3, 23)},
				},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import {named} from 'foo';
export default named;
export {named as named};`,
					Output: []string{`import {named} from 'foo';

export {named as named};
export {named as default} from 'foo';`, `


export {named as default, named} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 1, 2, 22), reexportError("named", 3, 9, 3, 23)},
				},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import defaultExport, {named} from 'foo';
export default defaultExport;
export {named};`,
					Output: []string{`import  {named} from 'foo';

export {named};
export {default} from 'foo';`, `


export {default, named} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 1, 2, 30), reexportError("named", 3, 9, 3, 14)},
				},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import defaultExport, {named} from 'foo';
export {defaultExport as default, named};`,
					Output: []string{`import  {named} from 'foo';
export { named};
export {default} from 'foo';`, `

export {default, named} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 9, 2, 33), reexportError("named", 2, 35, 2, 40)},
				},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import defaultExport from 'foo';
export const variable = defaultExport;`,
					Output: []string{`

export {default as variable} from 'foo';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("variable", 2, 1, 2, 39)},
				},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import {notUsedNotExported, exported} from 'foo';
export {exported};`,
					Errors: []rule_tester.InvalidTestCaseError{reexportError("exported", 2, 9, 2, 17, `import {notUsedNotExported, } from 'foo';

export {exported} from 'foo';`)},
				},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import {documented, other, unused} from 'foo';

/**
 * Preserve this documentation.
 */
export const documentedAlias = documented;
export {other};`,
					Errors: []rule_tester.InvalidTestCaseError{reexportError("other", 7, 9, 7, 14, `import {documented,  unused} from 'foo';

/**
 * Preserve this documentation.
 */
export const documentedAlias = documented;

export {other} from 'foo';`)},
				},
			},
		)
	})
	t.Run("LineEndings", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, upstreamRule,
			nil,
			[]rule_tester.InvalidTestCase{
				{FileName: "test.js", Code: "import {foo} from \"foo\";\r\nexport {foo};\r\n",
					Output: []string{"\r\n\r\n\r\nexport {foo} from \"foo\";"},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
				},
				{FileName: "test.js", Code: "import * as namespace from \"foo\";\r\nexport {namespace};\r\n",
					Output: []string{"\r\n\r\n\r\nexport * as namespace from \"foo\";"},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("namespace", 2, 9, 2, 18)},
				},
			},
		)
	})
	t.Run("Comments", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, upstreamRule,
			nil,
			[]rule_tester.InvalidTestCase{
				{FileName: "test.js", Code: `import {a} from "foo";
export {/* c1 */ a /* c2 */};`,
					Errors: []rule_tester.InvalidTestCaseError{reexportError("a", 2, 18, 2, 19)},
				},
				{FileName: "test.js", Code: `import {a /* c */} from "foo";
export {a};`,
					Errors: []rule_tester.InvalidTestCaseError{reexportError("a", 2, 9, 2, 10)},
				},
				{FileName: "test.js", Code: `import {a} from "foo";
// c
export {a};`,
					Output: []string{`
// c

export {a} from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("a", 3, 9, 3, 10)},
				},
			},
		)
	})
	t.Run("CommentedEdits", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, upstreamRule,
			nil,
			[]rule_tester.InvalidTestCase{
				{FileName: "test.js", Code: `import {a, b} from "foo";
export {/* c */ a /* c */};
use(b);`,
					Errors: []rule_tester.InvalidTestCaseError{reexportError("a", 2, 17, 2, 18)},
				},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import {a, b} from "foo";
export {/* c */ a};`,
					Errors: []rule_tester.InvalidTestCaseError{reexportError("a", 2, 17, 2, 18)},
				},
			},
		)
	})
	t.Run("Semicolons", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, upstreamRule,
			nil,
			[]rule_tester.InvalidTestCase{
				{FileName: "test.js", Code: `import {foo} from "foo";
log()
export {foo};
[1].forEach(log);`,
					Output: []string{`
log()
;
[1].forEach(log);
export {foo} from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 3, 9, 3, 12)},
				},
				{FileName: "test.js", Code: `import {foo} from "foo";
log()
export default foo;
[1].forEach(log);`,
					Output: []string{`
log()
;
[1].forEach(log);
export {foo as default} from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 3, 1, 3, 20)},
				},
				{FileName: "test.js", Code: `import {foo} from "foo";
log()
export const bar = foo;
[1].forEach(log);`,
					Output: []string{`
log()
;
[1].forEach(log);
export {foo as bar} from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("bar", 3, 1, 3, 24)},
				},
				{FileName: "test.js", Code: `log()
import {foo} from "foo";
[1].forEach(log);
export {foo};`,
					Output: []string{`log()
;
[1].forEach(log);

export {foo} from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 4, 9, 4, 12)},
				},
			},
		)
	})
	t.Run("AdjacentRemovals", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, upstreamRule,
			nil,
			[]rule_tester.InvalidTestCase{
				{FileName: "test.js", Code: `log()
import {foo} from "foo";
export {foo};
[1].forEach(log);`,
					Output: []string{`log()
;

[1].forEach(log);
export {foo} from "foo";`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 3, 9, 3, 12)},
				},
			},
		)
	})
	t.Run("Documentation", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, upstreamRule,
			[]rule_tester.ValidTestCase{
				{FileName: "test.js", Code: "export {default} from './foo.js';"},
				{FileName: "test.js", Code: "export {named} from './foo.js';"},
				{FileName: "test.js", Code: "export * as namespace from './foo.js';"},
				{FileName: "test.js", Code: `export {
	default,
	default as renamedDefault,
	named,
	named as renamedNamed,
} from './foo.js';`},
				{FileName: "test.js", Code: `// There is no substitution
import * as namespace from './foo.js';
export default namespace;`},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": false}, Code: `import {named1, named2} from './foo.js';

use(named1);

export {named1, named2};`},
			},
			[]rule_tester.InvalidTestCase{
				{FileName: "test.js", Code: `import defaultExport from './foo.js';
export default defaultExport;`,
					Output: []string{`

export {default} from './foo.js';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 1, 2, 30)},
				},
				{FileName: "test.js", Code: `import {named} from './foo.js';
export {named};`,
					Output: []string{`

export {named} from './foo.js';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("named", 2, 9, 2, 14)},
				},
				{FileName: "test.js", Code: `import * as namespace from './foo.js';
export {namespace};`,
					Output: []string{`

export * as namespace from './foo.js';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("namespace", 2, 9, 2, 18)},
				},
				{FileName: "test.js", Code: `import defaultExport, {named} from './foo.js';
export default defaultExport;
export {
	defaultExport as renamedDefault,
	named,
	named as renamedNamed,
};`,
					Output: []string{`import defaultExport, {named} from './foo.js';

export {
	defaultExport as renamedDefault,
	named,
	named as renamedNamed,
};
export {default} from './foo.js';`, "import  {named} from './foo.js';\n\nexport {\n\t\n\tnamed,\n\tnamed as renamedNamed,\n};\nexport {default, default as renamedDefault} from './foo.js';", "import  {named} from './foo.js';\n\nexport {\n\t\n\t\n\tnamed as renamedNamed,\n};\nexport {default, default as renamedDefault, named} from './foo.js';", `


export {default, default as renamedDefault, named, named as renamedNamed} from './foo.js';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("default", 2, 1, 2, 30), reexportError("renamedDefault", 4, 2, 4, 33), reexportError("named", 5, 2, 5, 7), reexportError("renamedNamed", 6, 2, 6, 23)},
				},
				{FileName: "test.js", Options: map[string]any{"checkUsedVariables": true}, Code: `import {named1, named2} from './foo.js';

use(named1);

export {named1, named2};`,
					Output: []string{`import {named1, } from './foo.js';

use(named1);

export {named1, };
export {named2} from './foo.js';`, `import {named1, } from './foo.js';

use(named1);


export {named2, named1} from './foo.js';`},
					Errors: []rule_tester.InvalidTestCaseError{reexportError("named1", 5, 9, 5, 15), reexportError("named2", 5, 17, 5, 23)},
				},
			},
		)
	})
}

func reexportError(name string, line, column, endLine, endColumn int, suggestionOutput ...string) rule_tester.InvalidTestCaseError {
	expected := rule_tester.InvalidTestCaseError{
		MessageId: "error",
		Message:   "Use `export…from` to re-export `" + name + "`.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
	for _, output := range suggestionOutput {
		expected.Suggestions = append(expected.Suggestions, rule_tester.InvalidTestCaseSuggestion{MessageId: "suggestion", Output: output})
	}
	return expected
}
