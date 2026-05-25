package no_useless_rename

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessRenameRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessRenameRule,
		[]rule_tester.ValidTestCase{
			// ---- Destructuring (declarations) ----
			{Code: `let {foo} = obj;`},
			{Code: `let {foo: bar} = obj;`},
			{Code: `let {foo: bar, baz: qux} = obj;`},
			{Code: `let {foo: {bar: baz}} = obj;`},
			{Code: `let {foo, bar: {baz: qux}} = obj;`},
			{Code: `let {'foo': bar} = obj;`},
			{Code: `let {'foo': bar, 'baz': qux} = obj;`},
			{Code: `let {'foo': {'bar': baz}} = obj;`},
			{Code: `let {foo, 'bar': {'baz': qux}} = obj;`},
			{Code: `let {['foo']: bar} = obj;`},
			{Code: `let {['foo']: bar, ['baz']: qux} = obj;`},
			{Code: `let {['foo']: {['bar']: baz}} = obj;`},
			{Code: `let {foo, ['bar']: {['baz']: qux}} = obj;`},
			{Code: `let {[foo]: foo} = obj;`},
			{Code: `let {['foo']: foo} = obj;`},
			{Code: `let {[foo]: bar} = obj;`},
			{Code: `function func({foo}) {}`},
			{Code: `function func({foo: bar}) {}`},
			{Code: `function func({foo: bar, baz: qux}) {}`},
			{Code: `({foo}) => {}`},
			{Code: `({foo: bar}) => {}`},
			{Code: `({foo: bar, baz: qui}) => {}`},

			// ---- Imports ----
			{Code: `import * as foo from 'foo';`},
			{Code: `import foo from 'foo';`},
			{Code: `import {foo} from 'foo';`},
			{Code: `import {foo as bar} from 'foo';`},
			{Code: `import {foo as bar, baz as qux} from 'foo';`},
			{Code: `import {'foo' as bar} from 'baz';`},

			// ---- Exports ----
			{Code: `export {foo} from 'foo';`},
			{Code: `var foo = 0;export {foo as bar};`},
			{Code: `var foo = 0; var baz = 0; export {foo as bar, baz as qux};`},
			{Code: `export {foo as bar} from 'foo';`},
			{Code: `export {foo as bar, baz as qux} from 'foo';`},
			{Code: `var foo = 0; export {foo as 'bar'};`},
			{Code: `export {foo as 'bar'} from 'baz';`},
			{Code: `export {'foo' as bar} from 'baz';`},
			{Code: `export {'foo' as 'bar'} from 'baz';`},
			{Code: `export {'' as ' '} from 'baz';`},
			{Code: `export {' ' as ''} from 'baz';`},
			{Code: `export {'foo'} from 'bar';`},

			// ---- Rest elements ----
			{Code: `const {...stuff} = myObject;`},
			{Code: `const {foo, ...stuff} = myObject;`},
			{Code: `const {foo: bar, ...stuff} = myObject;`},

			// ---- { ignoreDestructuring: true } ----
			{Code: `let {foo: foo} = obj;`, Options: map[string]interface{}{"ignoreDestructuring": true}},
			{Code: `let {foo: foo, bar: baz} = obj;`, Options: map[string]interface{}{"ignoreDestructuring": true}},
			{Code: `let {foo: foo, bar: bar} = obj;`, Options: map[string]interface{}{"ignoreDestructuring": true}},

			// ---- { ignoreImport: true } ----
			{Code: `import {foo as foo} from 'foo';`, Options: map[string]interface{}{"ignoreImport": true}},
			{Code: `import {foo as foo, bar as baz} from 'foo';`, Options: map[string]interface{}{"ignoreImport": true}},
			{Code: `import {foo as foo, bar as bar} from 'foo';`, Options: map[string]interface{}{"ignoreImport": true}},

			// ---- { ignoreExport: true } ----
			{Code: `var foo = 0;export {foo as foo};`, Options: map[string]interface{}{"ignoreExport": true}},
			{Code: `var foo = 0;var bar = 0;export {foo as foo, bar as baz};`, Options: map[string]interface{}{"ignoreExport": true}},
			{Code: `var foo = 0;var bar = 0;export {foo as foo, bar as bar};`, Options: map[string]interface{}{"ignoreExport": true}},
			{Code: `export {foo as foo} from 'foo';`, Options: map[string]interface{}{"ignoreExport": true}},
			{Code: `export {foo as foo, bar as baz} from 'foo';`, Options: map[string]interface{}{"ignoreExport": true}},
			{Code: `export {foo as foo, bar as bar} from 'foo';`, Options: map[string]interface{}{"ignoreExport": true}},

			// ---- Extra coverage beyond ESLint's suite ----
			// Numeric keys can't collide with identifier bindings — no rename.
			{Code: `let {0: foo} = obj;`},
			// Empty options object behaves the same as no options at all —
			// all three flags default to false.
			{Code: `let {foo} = obj;`, Options: map[string]interface{}{}},
			// Options explicitly set to false — same as defaults.
			{
				Code: `let {foo: bar} = obj;`,
				Options: map[string]interface{}{
					"ignoreDestructuring": false,
					"ignoreImport":        false,
					"ignoreExport":        false,
				},
			},
			// Array destructuring has no propertyName, so never triggers.
			{Code: `let [foo, bar] = arr;`},
			{Code: `let [{foo}] = arr;`},
			// Plain object literal — not an assignment target, don't touch.
			{Code: `const x = {foo: foo};`},
			{Code: `const y = {foo: foo, bar: bar};`},
			{Code: `const z = {...obj, foo: foo};`},
			// RHS of assignment — not an assignment pattern.
			{Code: `x = {foo: foo};`},
			// TS: destructuring with a type annotation still supports shorthand.
			{Code: `const {foo}: {foo: string} = obj;`},
			{Code: `function fn({foo: bar}: {foo: string}) {}`},
			// TS: type-only imports/exports without rename.
			{Code: `import type {foo as bar} from 'foo';`},
			{Code: `import {type foo as bar} from 'foo';`},
			{Code: `export type {foo as bar};`, Options: map[string]interface{}{}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Destructuring (declarations) — basic ----
			// Full Line/Column/EndLine/EndColumn position coverage for the
			// destructuring-declaration container.
			{
				Code:   `let {foo: foo} = obj;`,
				Output: []string{`let {foo} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 6, EndLine: 1, EndColumn: 14},
				},
			},
			// Assignment-pattern container — full position coverage.
			{
				Code:   `({foo: (foo)} = obj);`,
				Output: []string{`({foo} = obj);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:   `let {\u0061: a} = obj;`,
				Output: []string{`let {a} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment a unnecessarily renamed.", Line: 1, Column: 6},
				},
			},
			{
				Code:   `let {a: \u0061} = obj;`,
				Output: []string{`let {\u0061} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment a unnecessarily renamed.", Line: 1, Column: 6},
				},
			},
			{
				Code:   `let {\u0061: \u0061} = obj;`,
				Output: []string{`let {\u0061} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment a unnecessarily renamed.", Line: 1, Column: 6},
				},
			},
			{
				// StringLiteral key with a unicode escape — decoded value is
				// "a", matches the identifier binding `a`.
				Code:   `let {'\u0061': a} = obj;`,
				Output: []string{`let {a} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment a unnecessarily renamed.", Line: 1, Column: 6},
				},
			},
			{
				Code:   `let {a, foo: foo} = obj;`,
				Output: []string{`let {a, foo} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `let {foo: foo, bar: baz} = obj;`,
				Output: []string{`let {foo, bar: baz} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 6},
				},
			},
			{
				Code:   `let {foo: bar, baz: baz} = obj;`,
				Output: []string{`let {foo: bar, baz} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment baz unnecessarily renamed.", Line: 1, Column: 16},
				},
			},
			{
				Code:   `let {foo: foo, bar: bar} = obj;`,
				Output: []string{`let {foo, bar} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 6},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 16},
				},
			},
			{
				Code:   `let {foo: {bar: bar}} = obj;`,
				Output: []string{`let {foo: {bar}} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 12},
				},
			},
			{
				Code:   `let {foo: {bar: bar}, baz: baz} = obj;`,
				Output: []string{`let {foo: {bar}, baz} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 12},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment baz unnecessarily renamed.", Line: 1, Column: 23},
				},
			},

			// ---- String-literal keys ----
			{
				Code:   `let {'foo': foo} = obj;`,
				Output: []string{`let {foo} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 6},
				},
			},
			{
				Code:   `let {'foo': foo, 'bar': baz} = obj;`,
				Output: []string{`let {foo, 'bar': baz} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 6},
				},
			},
			{
				Code:   `let {'foo': bar, 'baz': baz} = obj;`,
				Output: []string{`let {'foo': bar, baz} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment baz unnecessarily renamed.", Line: 1, Column: 18},
				},
			},
			{
				Code:   `let {'foo': foo, 'bar': bar} = obj;`,
				Output: []string{`let {foo, bar} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 6},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 18},
				},
			},
			{
				Code:   `let {'foo': {'bar': bar}} = obj;`,
				Output: []string{`let {'foo': {bar}} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 14},
				},
			},
			{
				Code:   `let {'foo': {'bar': bar}, 'baz': baz} = obj;`,
				Output: []string{`let {'foo': {bar}, baz} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 14},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment baz unnecessarily renamed.", Line: 1, Column: 27},
				},
			},

			// ---- Destructuring with defaults ----
			{
				Code:   `let {foo: foo = 1, 'bar': bar = 1, baz: baz} = obj;`,
				Output: []string{`let {foo = 1, bar = 1, baz} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 6},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 20},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment baz unnecessarily renamed.", Line: 1, Column: 36},
				},
			},
			{
				Code:   `let {foo: {bar: bar = 1, 'baz': baz = 1}} = obj;`,
				Output: []string{`let {foo: {bar = 1, baz = 1}} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 12},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment baz unnecessarily renamed.", Line: 1, Column: 26},
				},
			},
			{
				Code:   `let {foo: {bar: bar = {}} = {}} = obj;`,
				Output: []string{`let {foo: {bar = {}} = {}} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 12},
				},
			},

			// ---- Assignment-pattern edge cases ----
			{
				// Parenthesised target of an assignment-pattern — shorthand
				// properties can't carry parens around an identifier, so no fix.
				Code: `({foo: (foo) = a} = obj);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code:   `let {foo: foo = (a)} = obj;`,
				Output: []string{`let {foo = (a)} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 6},
				},
			},
			{
				Code:   `let {foo: foo = (a, b)} = obj;`,
				Output: []string{`let {foo = (a, b)} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 6},
				},
			},

			// ---- Destructuring in function parameters ----
			{
				Code:   `function func({foo: foo}) {}`,
				Output: []string{`function func({foo}) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 16},
				},
			},
			{
				Code:   `function func({foo: foo, bar: baz}) {}`,
				Output: []string{`function func({foo, bar: baz}) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 16},
				},
			},
			{
				Code:   `function func({foo: bar, baz: baz}) {}`,
				Output: []string{`function func({foo: bar, baz}) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment baz unnecessarily renamed.", Line: 1, Column: 26},
				},
			},
			{
				Code:   `function func({foo: foo, bar: bar}) {}`,
				Output: []string{`function func({foo, bar}) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 16},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 26},
				},
			},
			{
				Code:   `function func({foo: foo = 1, 'bar': bar = 1, baz: baz}) {}`,
				Output: []string{`function func({foo = 1, bar = 1, baz}) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 16},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 30},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment baz unnecessarily renamed.", Line: 1, Column: 46},
				},
			},
			{
				Code:   `function func({foo: {bar: bar = 1, 'baz': baz = 1}}) {}`,
				Output: []string{`function func({foo: {bar = 1, baz = 1}}) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 22},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment baz unnecessarily renamed.", Line: 1, Column: 36},
				},
			},
			{
				Code:   `function func({foo: {bar: bar = {}} = {}}) {}`,
				Output: []string{`function func({foo: {bar = {}} = {}}) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 22},
				},
			},

			// ---- Destructuring in arrow parameters ----
			{
				Code:   `({foo: foo}) => {}`,
				Output: []string{`({foo}) => {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code:   `({foo: foo, bar: baz}) => {}`,
				Output: []string{`({foo, bar: baz}) => {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code:   `({foo: bar, baz: baz}) => {}`,
				Output: []string{`({foo: bar, baz}) => {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment baz unnecessarily renamed.", Line: 1, Column: 13},
				},
			},
			{
				Code:   `({foo: foo, bar: bar}) => {}`,
				Output: []string{`({foo, bar}) => {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 13},
				},
			},
			{
				Code:   `({foo: foo = 1, 'bar': bar = 1, baz: baz}) => {}`,
				Output: []string{`({foo = 1, bar = 1, baz}) => {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 17},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment baz unnecessarily renamed.", Line: 1, Column: 33},
				},
			},
			{
				Code:   `({foo: {bar: bar = 1, 'baz': baz = 1}}) => {}`,
				Output: []string{`({foo: {bar = 1, baz = 1}}) => {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 9},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment baz unnecessarily renamed.", Line: 1, Column: 23},
				},
			},
			{
				Code:   `({foo: {bar: bar = {}} = {}}) => {}`,
				Output: []string{`({foo: {bar = {}} = {}}) => {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 9},
				},
			},

			// ---- Rest element mixed with renames ----
			{
				Code:   `const {foo: foo, ...stuff} = myObject;`,
				Output: []string{`const {foo, ...stuff} = myObject;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 8},
				},
			},
			{
				Code:   `const {foo: foo, bar: baz, ...stuff} = myObject;`,
				Output: []string{`const {foo, bar: baz, ...stuff} = myObject;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 8},
				},
			},
			{
				Code:   `const {foo: foo, bar: bar, ...stuff} = myObject;`,
				Output: []string{`const {foo, bar, ...stuff} = myObject;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 8},
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 18},
				},
			},

			// ---- Imports ----
			// Import container — full position coverage.
			{
				Code:   `import {foo as foo} from 'foo';`,
				Output: []string{`import {foo} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import foo unnecessarily renamed.", Line: 1, Column: 9, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:   `import {'foo' as foo} from 'foo';`,
				Output: []string{`import {foo} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `import {\u0061 as a} from 'foo';`,
				Output: []string{`import {a} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import a unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `import {a as \u0061} from 'foo';`,
				Output: []string{`import {\u0061} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import a unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `import {\u0061 as \u0061} from 'foo';`,
				Output: []string{`import {\u0061} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import a unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `import {foo as foo, bar as baz} from 'foo';`,
				Output: []string{`import {foo, bar as baz} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `import {foo as bar, baz as baz} from 'foo';`,
				Output: []string{`import {foo as bar, baz} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import baz unnecessarily renamed.", Line: 1, Column: 21},
				},
			},
			{
				Code:   `import {foo as foo, bar as bar} from 'foo';`,
				Output: []string{`import {foo, bar} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import foo unnecessarily renamed.", Line: 1, Column: 9},
					{MessageId: "unnecessarilyRenamed", Message: "Import bar unnecessarily renamed.", Line: 1, Column: 21},
				},
			},

			// ---- Exports ----
			// Export container — full position coverage.
			{
				Code:   `var foo = 0; export {foo as foo};`,
				Output: []string{`var foo = 0; export {foo};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 22, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:   `var foo = 0; export {foo as 'foo'};`,
				Output: []string{`var foo = 0; export {foo};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 22},
				},
			},
			{
				Code:   `export {foo as 'foo'} from 'bar';`,
				Output: []string{`export {foo} from 'bar';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `export {'foo' as foo} from 'bar';`,
				Output: []string{`export {'foo'} from 'bar';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `export {'foo' as 'foo'} from 'bar';`,
				Output: []string{`export {'foo'} from 'bar';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `export {' 👍 ' as ' 👍 '} from 'bar';`,
				Output: []string{`export {' 👍 '} from 'bar';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export  👍  unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `export {'' as ''} from 'bar';`,
				Output: []string{`export {''} from 'bar';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export  unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `var a = 0; export {a as \u0061};`,
				Output: []string{`var a = 0; export {a};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export a unnecessarily renamed.", Line: 1, Column: 20},
				},
			},
			{
				Code:   `var \u0061 = 0; export {\u0061 as a};`,
				Output: []string{`var \u0061 = 0; export {\u0061};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export a unnecessarily renamed.", Line: 1, Column: 25},
				},
			},
			{
				Code:   `var \u0061 = 0; export {\u0061 as \u0061};`,
				Output: []string{`var \u0061 = 0; export {\u0061};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export a unnecessarily renamed.", Line: 1, Column: 25},
				},
			},
			{
				Code:   `var foo = 0; var bar = 0; export {foo as foo, bar as baz};`,
				Output: []string{`var foo = 0; var bar = 0; export {foo, bar as baz};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 35},
				},
			},
			{
				Code:   `var foo = 0; var baz = 0; export {foo as bar, baz as baz};`,
				Output: []string{`var foo = 0; var baz = 0; export {foo as bar, baz};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export baz unnecessarily renamed.", Line: 1, Column: 47},
				},
			},
			{
				Code:   `var foo = 0; var bar = 0;export {foo as foo, bar as bar};`,
				Output: []string{`var foo = 0; var bar = 0;export {foo, bar};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 34},
					{MessageId: "unnecessarilyRenamed", Message: "Export bar unnecessarily renamed.", Line: 1, Column: 46},
				},
			},
			{
				Code:   `export {foo as foo} from 'foo';`,
				Output: []string{`export {foo} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `export {a as \u0061} from 'foo';`,
				Output: []string{`export {a} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export a unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `export {\u0061 as a} from 'foo';`,
				Output: []string{`export {\u0061} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export a unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `export {\u0061 as \u0061} from 'foo';`,
				Output: []string{`export {\u0061} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export a unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `export {foo as foo, bar as baz} from 'foo';`,
				Output: []string{`export {foo, bar as baz} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `var foo = 0; var bar = 0; export {foo as bar, baz as baz} from 'foo';`,
				Output: []string{`var foo = 0; var bar = 0; export {foo as bar, baz} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export baz unnecessarily renamed.", Line: 1, Column: 47},
				},
			},
			{
				Code:   `export {foo as foo, bar as bar} from 'foo';`,
				Output: []string{`export {foo, bar} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 9},
					{MessageId: "unnecessarilyRenamed", Message: "Export bar unnecessarily renamed.", Line: 1, Column: 21},
				},
			},

			// ---- Comment preservation: destructuring ----
			{
				// Comment before the key — outside the property assignment, so
				// the fix still runs.
				Code:   `({/* comment */foo: foo} = {});`,
				Output: []string{`({/* comment */foo} = {});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 16},
				},
			},
			{
				Code:   `({/* comment */foo: foo = 1} = {});`,
				Output: []string{`({/* comment */foo = 1} = {});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 16},
				},
			},
			{
				Code:   `({foo, /* comment */bar: bar} = {});`,
				Output: []string{`({foo, /* comment */bar} = {});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment bar unnecessarily renamed.", Line: 1, Column: 21},
				},
			},
			{
				// Comment between key and colon — would be lost, reject fix.
				Code: `({foo/**/ : foo} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code: `({foo/**/ : foo = 1} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code: `({foo /**/: foo} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code: `({foo /**/: foo = 1} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code: "({foo://\nfoo} = {});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				// Comment between colon and value — would be lost, reject fix.
				Code: `({foo: /**/foo} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				// Parenthesised value with leading comment inside parens — still lost.
				Code: `({foo: (/**/foo)} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code: `({foo: (foo/**/)} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code: "({foo: (foo //\n)} = {});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code: `({foo: /**/foo = 1} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code: `({foo: (/**/foo) = 1} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code: `({foo: (foo/**/) = 1} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				// Comment after the value is outside the property range — fix applies.
				Code:   `({foo: foo/* comment */} = {});`,
				Output: []string{`({foo/* comment */} = {});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code:   "({foo: foo//comment\n,bar} = {});",
				Output: []string{"({foo//comment\n,bar} = {});"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				// Comment inside the AssignmentPattern — preserved by the fix.
				Code:   `({foo: foo/* comment */ = 1} = {});`,
				Output: []string{`({foo/* comment */ = 1} = {});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code:   "({foo: foo // comment\n = 1} = {});",
				Output: []string{"({foo // comment\n = 1} = {});"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code:   `({foo: foo = /* comment */ 1} = {});`,
				Output: []string{`({foo = /* comment */ 1} = {});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code:   "({foo: foo = // comment\n 1} = {});",
				Output: []string{"({foo = // comment\n 1} = {});"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			{
				Code:   `({foo: foo = (1/* comment */)} = {});`,
				Output: []string{`({foo = (1/* comment */)} = {});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},

			// ---- Comment preservation: import specifiers ----
			{
				Code:   `import {/* comment */foo as foo} from 'foo';`,
				Output: []string{`import {/* comment */foo} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import foo unnecessarily renamed.", Line: 1, Column: 22},
				},
			},
			{
				Code:   `import {foo,/* comment */bar as bar} from 'foo';`,
				Output: []string{`import {foo,/* comment */bar} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import bar unnecessarily renamed.", Line: 1, Column: 26},
				},
			},
			{
				Code: `import {foo/**/ as foo} from 'foo';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code: `import {foo /**/as foo} from 'foo';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code: "import {foo //\nas foo} from 'foo';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code: `import {foo as/**/foo} from 'foo';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `import {foo as foo/* comment */} from 'foo';`,
				Output: []string{`import {foo/* comment */} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			{
				Code:   `import {foo as foo/* comment */,bar} from 'foo';`,
				Output: []string{`import {foo/* comment */,bar} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},

			// ---- Comment preservation: export specifiers ----
			{
				Code:   `let foo; export {/* comment */foo as foo};`,
				Output: []string{`let foo; export {/* comment */foo};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 31},
				},
			},
			{
				Code:   `let foo, bar; export {foo,/* comment */bar as bar};`,
				Output: []string{`let foo, bar; export {foo,/* comment */bar};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export bar unnecessarily renamed.", Line: 1, Column: 40},
				},
			},
			{
				Code: `let foo; export {foo/**/as foo};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 18},
				},
			},
			{
				Code: `let foo; export {foo as/**/ foo};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 18},
				},
			},
			{
				Code: `let foo; export {foo as /**/foo};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 18},
				},
			},
			{
				Code: "let foo; export {foo as//comment\n foo};",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 18},
				},
			},
			{
				Code:   `let foo; export {foo as foo/* comment*/};`,
				Output: []string{`let foo; export {foo/* comment*/};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 18},
				},
			},
			{
				Code:   `let foo, bar; export {foo as foo/* comment*/,bar};`,
				Output: []string{`let foo, bar; export {foo/* comment*/,bar};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 23},
				},
			},
			{
				Code:   "let foo, bar; export {foo as foo//comment\n,bar};",
				Output: []string{"let foo, bar; export {foo//comment\n,bar};"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export foo unnecessarily renamed.", Line: 1, Column: 23},
				},
			},

			// ---- Extra coverage: real-world scenarios beyond ESLint's suite ----

			// Multi-line destructuring — the report's position must track the
			// actual line the BindingElement sits on, not the enclosing
			// statement. Full position coverage (required per port-rule guide).
			{
				Code: `let {
  foo: foo,
  bar: baz
} = obj;`,
				Output: []string{`let {
  foo,
  bar: baz
} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 2, Column: 3, EndLine: 2, EndColumn: 11},
				},
			},
			// Multi-line assignment pattern.
			{
				Code: `({
  foo: foo,
  bar: baz
} = obj);`,
				Output: []string{`({
  foo,
  bar: baz
} = obj);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 2, Column: 3},
				},
			},
			// Nested assignment pattern — the inner ObjectLiteralExpression is
			// not a direct LHS, but `ast.IsAssignmentTarget` walks up through
			// PropertyAssignment / ObjectLiteralExpression to confirm the
			// outer `=`. Tests that the listener fires on nested patterns.
			{
				Code:   `({a: {foo: foo}} = obj);`,
				Output: []string{`({a: {foo}} = obj);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 7},
				},
			},
			// Array-containing-object assignment pattern.
			{
				Code:   `[{foo: foo}] = arr;`,
				Output: []string{`[{foo}] = arr;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 3},
				},
			},
			// `for-of` with object destructuring as the iteration target.
			{
				Code:   `for ({foo: foo} of arr) {}`,
				Output: []string{`for ({foo} of arr) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 7},
				},
			},
			// TS: destructuring with a type annotation — the annotation is
			// attached to the VariableDeclaration, not the BindingElement, so
			// the autofix range is unaffected.
			{
				Code:   `const {foo: foo}: {foo: string} = obj;`,
				Output: []string{`const {foo}: {foo: string} = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Destructuring assignment foo unnecessarily renamed.", Line: 1, Column: 8},
				},
			},
			// TS: type-only import with useless rename.
			{
				Code:   `import type {foo as foo} from 'foo';`,
				Output: []string{`import type {foo} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import foo unnecessarily renamed.", Line: 1, Column: 14},
				},
			},
			// TS: inline `type` modifier on an individual import specifier.
			// Matches ESLint's behavior: the fix replaces the entire
			// ImportSpecifier range with the local name, so the `type`
			// modifier is dropped from the output.
			{
				Code:   `import {type foo as foo} from 'foo';`,
				Output: []string{`import {foo} from 'foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Import foo unnecessarily renamed.", Line: 1, Column: 9},
				},
			},
			// TS: type-only export with useless rename.
			{
				Code:   `type Foo = string; export type {Foo as Foo};`,
				Output: []string{`type Foo = string; export type {Foo};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyRenamed", Message: "Export Foo unnecessarily renamed.", Line: 1, Column: 33},
				},
			},
		},
	)
}
