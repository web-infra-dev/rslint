package group_exports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/group-exports.js
// All 25 valid and 14 invalid cases, including the Flow cases whose syntax
// is also valid TypeScript. Exact ranges come from the pinned rule.
const namedMessage = "Multiple named export declarations; consolidate all named exports into a single export declaration"
const commonJSMessage = "Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`"

func TestGroupExportsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &GroupExportsRule,
		[]rule_tester.ValidTestCase{
			{Code: "export const test = true"},
			{Code: `
      export default {}
      export const test = true
    `},
			{Code: `
      const first = true
      const second = true
      export {
        first,
        second
      }
    `},
			{Code: `
      export default {}
      /* test */
      export const test = true
    `},
			{Code: `
      export default {}
      // test
      export const test = true
    `},
			{Code: `
      export const test = true
      /* test */
      export default {}
    `},
			{Code: `
      export const test = true
      // test
      export default {}
    `},
			{Code: `
      export { default as module1 } from './module-1'
      export { default as module2 } from './module-2'
    `},
			{Code: "module.exports = {} "},
			{Code: `
      module.exports = { test: true,
        another: false }
    `},
			{Code: "exports.test = true"},
			{Code: `
      module.exports = {}
      const test = module.exports
    `},
			{Code: `
      exports.test = true
      const test = exports.test
    `},
			{Code: `
      module.exports = {}
      module.exports.too.deep = true
    `},
			{Code: `
      module.exports.deep.first = true
      module.exports.deep.second = true
    `},
			{Code: `
      module.exports = {}
      exports.too.deep = true
    `},
			{Code: `
      export default {}
      const test = true
      export { test }
    `},
			{Code: `
      const test = true
      export { test }
      const another = true
      export default {}
    `},
			{Code: `
      module.something.else = true
      module.something.different = true
    `},
			{Code: `
      module.exports.test = true
      module.something.different = true
    `},
			{Code: `
      exports.test = true
      module.something.different = true
    `},
			{Code: `
      unrelated = 'assignment'
      module.exports.test = true
    `},
			{Code: `
      type firstType = {
        propType: string
      };
      const first = {};
      export type { firstType };
      export { first };
    `},
			{Code: `
      type firstType = {
        propType: string
      };
      type secondType = {
        propType: string
      };
      export type { firstType, secondType };
    `},
			{Code: `
      export type { type1A, type1B } from './module-1'
      export { method1 } from './module-1'
    `},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `
        export const test = true
        export const another = true
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 2, Column: 9, EndLine: 2, EndColumn: 33},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 9, EndLine: 3, EndColumn: 36},
				}},
			{Code: `
        export { method1 } from './module-1'
        export { method2 } from './module-1'
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 2, Column: 9, EndLine: 2, EndColumn: 45},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 9, EndLine: 3, EndColumn: 45},
				}},
			{Code: `
        module.exports = {}
        module.exports.test = true
        module.exports.another = true
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 9, EndLine: 2, EndColumn: 28},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 9, EndLine: 3, EndColumn: 35},
					{MessageId: "", Message: commonJSMessage, Line: 4, Column: 9, EndLine: 4, EndColumn: 38},
				}},
			{Code: `
        module.exports = {}
        module.exports.test = true
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 9, EndLine: 2, EndColumn: 28},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 9, EndLine: 3, EndColumn: 35},
				}},
			{Code: `
        module.exports = { test: true }
        module.exports.another = true
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 9, EndLine: 2, EndColumn: 40},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 9, EndLine: 3, EndColumn: 38},
				}},
			{Code: `
        module.exports.test = true
        module.exports.another = true
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 9, EndLine: 2, EndColumn: 35},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 9, EndLine: 3, EndColumn: 38},
				}},
			{Code: `
        exports.test = true
        module.exports.another = true
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 9, EndLine: 2, EndColumn: 28},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 9, EndLine: 3, EndColumn: 38},
				}},
			{Code: `
        module.exports = () => {}
        module.exports.attached = true
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 9, EndLine: 2, EndColumn: 34},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 9, EndLine: 3, EndColumn: 39},
				}},
			{Code: `
        module.exports = function test() {}
        module.exports.attached = true
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 9, EndLine: 2, EndColumn: 44},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 9, EndLine: 3, EndColumn: 39},
				}},
			{Code: `
        module.exports = () => {}
        exports.test = true
        exports.another = true
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 9, EndLine: 2, EndColumn: 34},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 9, EndLine: 3, EndColumn: 28},
					{MessageId: "", Message: commonJSMessage, Line: 4, Column: 9, EndLine: 4, EndColumn: 31},
				}},
			{Code: `
        module.exports = "non-object"
        module.exports.attached = true
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 9, EndLine: 2, EndColumn: 38},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 9, EndLine: 3, EndColumn: 39},
				}},
			{Code: `
        module.exports = "non-object"
        module.exports.attached = true
        module.exports.another = true
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 9, EndLine: 2, EndColumn: 38},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 9, EndLine: 3, EndColumn: 39},
					{MessageId: "", Message: commonJSMessage, Line: 4, Column: 9, EndLine: 4, EndColumn: 38},
				}},
			{Code: `
        type firstType = {
          propType: string
        };
        type secondType = {
          propType: string
        };
        const first = {};
        export type { firstType };
        export type { secondType };
        export { first };
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 9, Column: 9, EndLine: 9, EndColumn: 35},
					{MessageId: "", Message: namedMessage, Line: 10, Column: 9, EndLine: 10, EndColumn: 36},
				}},
			{Code: `
        export type { type1 } from './module-1'
        export type { type2 } from './module-1'
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 2, Column: 9, EndLine: 2, EndColumn: 48},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 9, EndLine: 3, EndColumn: 48},
				}},
		},
	)
}

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/group-exports.md
func TestGroupExportsUpstreamDocs(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &GroupExportsRule,
		[]rule_tester.ValidTestCase{
			{Code: `// A single named export declaration -> ok
export const valid = true
`},
			{Code: `const first = true
const second = true

// A single named export declaration -> ok
export {
  first,
  second,
}
`},
			{Code: `// Aggregating exports -> ok
export { default as module1 } from 'module-1'
export { default as module2 } from 'module-2'
`},
			{Code: `// A single exports assignment -> ok
module.exports = {
  first: true,
  second: true
}
`},
			{Code: `const first = true
const second = true

// A single exports assignment -> ok
module.exports = {
  first,
  second,
}
`},
			{Code: `function test() {}
test.property = true
test.another = true

// A single exports assignment -> ok
module.exports = test
`},
			{Code: `const first = true;
type firstType = boolean

// A single named export declaration (type exports handled separately) -> ok
export {first}
export type {firstType}
`},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `// Multiple named export statements -> not ok!
export const first = true
export const second = true
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 26},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 27},
				}},
			{Code: `// Aggregating exports from the same module -> not ok!
export { module1 } from 'module-1'
export { module2 } from 'module-1'
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 35},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 35},
				}},
			{Code: `// Multiple exports assignments -> not ok!
exports.first = true
exports.second = true
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 21},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 22},
				}},
			{Code: `// Multiple exports assignments -> not ok!
module.exports = {}
module.exports.first = true
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 20},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 28},
				}},
			{Code: `// Multiple exports assignments -> not ok!
module.exports = () => {}
module.exports.first = true
module.exports.second = true
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 26},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 28},
					{MessageId: "", Message: commonJSMessage, Line: 4, Column: 1, EndLine: 4, EndColumn: 29},
				}},
			{Code: `type firstType = boolean
type secondType = any

// Multiple named type export statements -> not ok!
export type {firstType}
export type {secondType}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 5, Column: 1, EndLine: 5, EndColumn: 24},
					{MessageId: "", Message: namedMessage, Line: 6, Column: 1, EndLine: 6, EndColumn: 25},
				}},
		},
	)
}
