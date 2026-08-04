package prefer_set_has_test

// TestPreferSetHasUpstream migrates the full valid/invalid suite from upstream
// test/prefer-set-has.js 1:1. Position assertions cover line/column for every
// invalid case. rslint-specific lock-in cases live in the
// prefer_set_has_extras_test.go file.

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_set_has"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func errorAt(name string, line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "error",
		Message:   "`" + name + "` should be a `Set`, and use `" + name + ".has()` to check existence or non-existence.",
		Line:      line,
		Column:    column,
	}
}

func minimumItems(n int) any {
	return []any{map[string]any{"minimumItems": n}}
}

func TestPreferSetHasUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_set_has.PreferSetHasRule,
		[]rule_tester.ValidTestCase{
			// ---- Upstream: valid (default options) ----
			{Code: "const foo = new Set([1, 2, 3]);\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
			// Only called once
			{Code: "const foo = [1, 2, 3];\nconst isExists = foo.includes(1);"},
			{Code: "while (a) {\n\tconst foo = [1, 2, 3];\n\tconst isExists = foo.includes(1);\n}"},
			{Code: "const foo = [1, 2, 3];\n(() => {})(foo.includes(1));"},
			// Not `VariableDeclarator`
			{Code: "foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const exists = foo.includes(1);"},
			{Code: "const exists = [1, 2, 3].includes(1);"},
			// Didn't call `includes()`
			{Code: "const foo = [1, 2, 3];"},
			// Not `CallExpression`
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes;\n}"},
			// Not `foo.includes()`
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn includes(foo);\n}"},
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn bar.includes(foo);\n}"},
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo[includes](1);\n}"},
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.indexOf(1) !== -1;\n}"},
			// Unsupported extra references
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\tfoo.includes(1);\n\tfoo.length = 1;\n}"},
			// Duplicates and `-0` can change the values produced by iteration/spread/forEach.
			{Code: "const foo = [1, 1, 2];\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [NaN, NaN];\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [0, -0];\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 1, 2];\nconst length = foo.length;\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			// Computed `length` access is not recognized as a length read, so the rule bails
			{Code: "const foo = [1, 2, 3];\nfoo['length'];\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 1, 2];\ncall(...foo);\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 1, 2];\nfoo.forEach(element => {\n\tconsole.log(element);\n});\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [-0];\nconst values = [...foo];\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const value = -0;\nconst foo = [value];\nconst values = [...foo];\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			// Unknown uniqueness
			{Code: "const value = 1;\nconst foo = [value, value];\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const first = getValue();\nconst second = getValue();\nconst foo = [first, second];\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			// Unsupported initializer shapes
			{Code: "const foo = [1, , 2];\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [...bar];\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = Array.of(1, 2, 3);\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			// Unsupported extra references
			{Code: "const foo = [1, 2, 3];\nconsole.log(foo[0]);\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "function f(a, b, c) {\n\tconst foo = [a];\n\treturn c.map(i => b.includes(foo[i]));\n}"},
			{Code: "const foo = [1, 2, 3];\nfunction unicorn(index) {\n\treturn bar.includes(foo[index]) || bar.includes(foo[index + 1]);\n}"},
			{Code: "const foo = [1, 2, 3];\nfunction unicorn(value) {\n\treturn foo[value].includes(value);\n}"},
			{Code: "const foo = [1, 2, 3];\nconst object = {...foo};\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 2, 3];\nfoo.forEach((element, index) => {\n\tconsole.log(element, index);\n});\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 2, 3];\nfoo.forEach((...elements) => {\n\tconsole.log(elements[1]);\n});\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 2, 3];\nfoo.forEach(function (element) {\n\tconsole.log(arguments[1]);\n});\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 2, 3];\nfoo.forEach(callback);\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			// `.length` writes
			{Code: "const foo = [1, 2, 3];\ndelete foo.length;\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 2, 3];\nfoo.length++;\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 2, 3];\nconst object = {};\n\nfor (foo.length in object) {}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 2, 3];\nconst values = [];\n\nfor (foo.length of values) {}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 2, 3];\nconst object = {};\n\n({length: foo.length} = object);\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 2, 3];\nconst object = {};\n\n({length: foo.length = 0} = object);\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 2, 3];\nconst object = {};\n\n({...foo.length} = object);\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 2, 3];\nconst values = [];\n\n[...foo.length] = values;\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			// One-off lookup with supported extra references
			{Code: "const foo = [1, 2, 3];\nconst values = [...foo];\nconst exists = foo.includes(1);"},
			{Code: "const foo = [1, 2, 3];\nconst length = foo.length;\nconst exists = foo.includes(1);"},
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\tif (foo.includes(1)) {}\n\treturn foo;\n}"},
			// Declared more than once
			{Code: "var foo = [1, 2, 3];\nvar foo = [4, 5, 6];\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = bar;\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			// Extra arguments
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes();\n}"},
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes(1, 1);\n}"},
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes(1, 0);\n}"},
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes(1, undefined);\n}"},
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes(...[1]);\n}"},
			// Optional
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo?.includes(1);\n}"},
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes?.(1);\n}"},
			{Code: "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo?.includes?.(1);\n}"},
			// Different scope
			{Code: "function unicorn() {\n\tconst foo = [1, 2, 3];\n}\nfunction unicorn2() {\n\treturn foo.includes(1);\n}"},
			// `export`
			{Code: "export const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "module.exports = [1, 2, 3];\nfunction unicorn() {\n\treturn module.exports.includes(1);\n}"},
			{Code: "const foo = [1, 2, 3];\nexport {foo};\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = [1, 2, 3];\nexport default foo;\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = [1, 2, 3];\nexport {foo as bar};\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = [1, 2, 3];\nmodule.exports = foo;\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = [1, 2, 3];\nexports = foo;\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = [1, 2, 3];\nmodule.exports.foo = foo;\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			// `Array()`
			{Code: "const foo = NotArray(1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			// `new Array()`
			{Code: "const foo = new NotArray(1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			// `Array.from()` / `Array.of()` — Not `Array`
			{Code: "const foo = NotArray.from({length: 1}, (_, index) => index);\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = NotArray.of(1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			// Not `Listed`
			{Code: "const foo = Array.notListed();\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			// Computed
			{Code: "const foo = Array[from]({length: 1}, (_, index) => index);\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = Array[of](1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			// Not Identifier
			{Code: "const foo = 'Array'.from({length: 1}, (_, index) => index);\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = 'Array'.of(1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = Array['from']({length: 1}, (_, index) => index);\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = Array['of'](1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = of(1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = from({length: 1}, (_, index) => index);\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			// Methods — Not call
			{Code: "const foo = bar.filter;\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = new bar.filter();\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			// Not MemberExpression
			{Code: "const foo = filter();\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			// Computed
			{Code: "const foo = bar[filter]();\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			// Not `Identifier`
			{Code: "const foo = bar[\"filter\"]();\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			// Not listed method
			{Code: "const foo = bar.notListed();\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			// concat/slice on non-array receiver
			{Code: "const foo = bar.slice();\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			{Code: "const foo = bar.concat();\nfunction unicorn() {\n\treturn foo.includes(1);\n}"},
			// `lodash`
			{Code: "const foo = _.map([1, 2, 3], value => value);\nfunction unicorn() {\n\treturn _.includes(foo, 1);\n}"},
			{Code: "const text = 'abc'.slice();\ntext.includes('ab') || text.includes('bc');"},
			{Code: "const text = `abc`.concat('def');\ntext.includes('ab') || text.includes('bc');"},
			{Code: "const text = `1abc`.slice();\ntext.includes('ab') || text.includes('bc');"},
			{Code: "let items = [1, 2, 3];\nitems = 'abc';\nconst foo = items.slice();\nfoo.includes('ab') || foo.includes('bc');"},
			// `Iterator.concat()`
			{Code: "const foo = Iterator.concat(bar);\nfoo.includes(1) || foo.includes(2);"},

			// ---- Upstream: valid (minimumItems: 5) ----
			{Code: "const foo = [1, 2, 3, 4];\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = Array.of(1, 2, 3, 4);\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = Array.from({length: 4}, (_, index) => index);\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = Array.from([1, 2, 3, 4]);\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = Array.from([1, 2, 3, 4, ...bar]);\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = Array.from(...bar);\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = Array.from('test');\nfunction unicorn() {\n\treturn foo.includes('t');\n}", Options: minimumItems(5)},
			{Code: "const foo = Array(4);\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = Array(...bar);\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = Array(1, 2, 3, 4);\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = new Array(4);\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = new Array(...bar);\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = Array(2 ** 32);\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = bar.map(value => value);\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = Array.of(...bar);\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = [1, 2, 3, 4, ...bar];\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = Array.from({\n\tlength: 5,\n\t*[Symbol.iterator]() {\n\t\tyield 1;\n\t},\n});\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},
			{Code: "const foo = Array.from({length: 2 ** 32}, (_, index) => index);\nfunction unicorn() {\n\treturn foo.includes(1);\n}", Options: minimumItems(5)},

			// ---- Upstream: valid (TypeScript, `.length` write targets) ----
			{Code: "@connect(\n\tstate => {\n\t\tconst availableComponents = ['is']\n\t\tif (nsConfig.enabled) availableComponents.push('ns')\n\t\tif (jsConfig.enabled) availableComponents.push('js')\n\n\t\treturn {\n\t\t\tavailableComponents,\n\t\t}\n\t},\n)\nexport default class A {}"},
			{Code: "const foo = [1, 2, 3];\n(foo.length as number) = 1;\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 2, 3];\n(foo.length!) = 1;\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
			{Code: "const foo = [1, 2, 3];\n(<number>foo.length) = 1;\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Upstream: invalid (default options) ----
			{
				Code:   "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// Called multiple times
			{
				Code:   "const foo = [1, 2, 3];\nconst isExists = foo.includes(1);\nconst isExists2 = foo.includes(2);",
				Output: []string{"const foo = new Set([1, 2, 3]);\nconst isExists = foo.has(1);\nconst isExists2 = foo.has(2);"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// `ForOfStatement`
			{
				Code:   "const foo = [1, 2, 3];\nfor (const a of b) {\n\tfoo.includes(1);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfor (const a of b) {\n\tfoo.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "async function unicorn() {\n\tconst foo = [1, 2, 3];\n\tfor await (const a of b) {\n\t\tfoo.includes(1);\n\t}\n}",
				Output: []string{"async function unicorn() {\n\tconst foo = new Set([1, 2, 3]);\n\tfor await (const a of b) {\n\t\tfoo.has(1);\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 2, 8)},
			},
			// `ForStatement`
			{
				Code:   "const foo = [1, 2, 3];\nfor (let i = 0; i < n; i++) {\n\tfoo.includes(1);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfor (let i = 0; i < n; i++) {\n\tfoo.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// `ForInStatement`
			{
				Code:   "const foo = [1, 2, 3];\nfor (let a in b) {\n\tfoo.includes(1);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfor (let a in b) {\n\tfoo.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// `WhileStatement`
			{
				Code:   "const foo = [1, 2, 3];\nwhile (a)  {\n\tfoo.includes(1);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nwhile (a)  {\n\tfoo.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// `DoWhileStatement`
			{
				Code:   "const foo = [1, 2, 3];\ndo {\n\tfoo.includes(1);\n} while (a)",
				Output: []string{"const foo = new Set([1, 2, 3]);\ndo {\n\tfoo.has(1);\n} while (a)"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "const foo = [1, 2, 3];\ndo {\n\t// …\n} while (foo.includes(1))",
				Output: []string{"const foo = new Set([1, 2, 3]);\ndo {\n\t// …\n} while (foo.has(1))"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// `FunctionDeclaration`
			{
				Code:   "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "const foo = [1, 2, 3];\nfunction * unicorn() {\n\treturn foo.includes(1);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfunction * unicorn() {\n\treturn foo.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "const foo = [1, 2, 3];\nasync function unicorn() {\n\treturn foo.includes(1);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nasync function unicorn() {\n\treturn foo.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "const foo = [1, 2, 3];\nasync function * unicorn() {\n\treturn foo.includes(1);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nasync function * unicorn() {\n\treturn foo.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// `FunctionExpression`
			{
				Code:   "const foo = [1, 2, 3];\nconst unicorn = function () {\n\treturn foo.includes(1);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nconst unicorn = function () {\n\treturn foo.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// `ArrowFunctionExpression`
			{
				Code:   "const foo = [1, 2, 3];\nconst unicorn = () => foo.includes(1);",
				Output: []string{"const foo = new Set([1, 2, 3]);\nconst unicorn = () => foo.has(1);"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "const foo = [1, 2, 3];\nconst a = {\n\tb() {\n\t\treturn foo.includes(1);\n\t}\n};",
				Output: []string{"const foo = new Set([1, 2, 3]);\nconst a = {\n\tb() {\n\t\treturn foo.has(1);\n\t}\n};"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "const foo = [1, 2, 3];\nclass A {\n\tb() {\n\t\treturn foo.includes(1);\n\t}\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nclass A {\n\tb() {\n\t\treturn foo.has(1);\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// SpreadElement
			{
				Code:   "const foo = [...bar];\nfunction unicorn() {\n\treturn foo.includes(1);\n}\nbar.pop();",
				Output: []string{"const foo = new Set([...bar]);\nfunction unicorn() {\n\treturn foo.has(1);\n}\nbar.pop();"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// Multiple references
			{
				Code:   "const foo = [1, 2, 3];\nfunction unicorn() {\n\tconst exists = foo.includes(1);\n\tfunction isExists(find) {\n\t\treturn foo.includes(find);\n\t}\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfunction unicorn() {\n\tconst exists = foo.has(1);\n\tfunction isExists(find) {\n\t\treturn foo.has(find);\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "function wrap() {\n\tconst foo = [1, 2, 3];\n\n\tfunction unicorn() {\n\t\treturn foo.includes(1);\n\t}\n}\n\nconst bar = [4, 5, 6];\n\nfunction unicorn() {\n\treturn bar.includes(1);\n}",
				Output: []string{"function wrap() {\n\tconst foo = new Set([1, 2, 3]);\n\n\tfunction unicorn() {\n\t\treturn foo.has(1);\n\t}\n}\n\nconst bar = new Set([4, 5, 6]);\n\nfunction unicorn() {\n\treturn bar.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 2, 8), errorAt("bar", 9, 7)},
			},
			{
				Code:   "const foo = [1, 2, 3];\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.has(value);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "const foo = [1, 2, 3];\nconst bar = [...foo];\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nconst bar = [...foo];\n\nfunction unicorn(value) {\n\treturn foo.has(value);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "const foo = [1, 2, 3];\ncall(...foo);\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\ncall(...foo);\n\nfunction unicorn(value) {\n\treturn foo.has(value);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "const foo = [1, 2, 3];\nnew Call(...foo);\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nnew Call(...foo);\n\nfunction unicorn(value) {\n\treturn foo.has(value);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "const foo = [1, 2, 3];\nconst length = foo.length;\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nconst length = foo.size;\n\nfunction unicorn(value) {\n\treturn foo.has(value);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "const foo = [1, 2, 3];\nfoo.forEach(element => {\n\tconsole.log(element);\n});\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfoo.forEach(element => {\n\tconsole.log(element);\n});\n\nfunction unicorn(value) {\n\treturn foo.has(value);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// Different scope
			{
				Code:   "const foo = [1, 2, 3];\nfunction wrap() {\n\tconst exists = foo.includes(1);\n\tconst bar = [1, 2, 3];\n\n\tfunction outer(find) {\n\t\tconst foo = [1, 2, 3];\n\t\twhile (a) {\n\t\t\tfoo.includes(1);\n\t\t}\n\n\t\tfunction inner(find) {\n\t\t\tconst bar = [1, 2, 3];\n\t\t\twhile (a) {\n\t\t\t\tconst exists = bar.includes(1);\n\t\t\t}\n\t\t}\n\t}\n}",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfunction wrap() {\n\tconst exists = foo.has(1);\n\tconst bar = [1, 2, 3];\n\n\tfunction outer(find) {\n\t\tconst foo = new Set([1, 2, 3]);\n\t\twhile (a) {\n\t\t\tfoo.has(1);\n\t\t}\n\n\t\tfunction inner(find) {\n\t\t\tconst bar = new Set([1, 2, 3]);\n\t\t\twhile (a) {\n\t\t\t\tconst exists = bar.has(1);\n\t\t\t}\n\t\t}\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7), errorAt("foo", 7, 9), errorAt("bar", 13, 10)},
			},
			// `Array()`
			{
				Code:   "const foo = Array(1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output: []string{"const foo = new Set(Array(1, 2));\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// `new Array()`
			{
				Code:   "const foo = new Array(1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output: []string{"const foo = new Set(new Array(1, 2));\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// `Array.from()`
			{
				Code:   "const foo = Array.from({length: 1}, (_, index) => index);\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output: []string{"const foo = new Set(Array.from({length: 1}, (_, index) => index));\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// `Array.of()`
			{
				Code:   "const foo = Array.of(1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output: []string{"const foo = new Set(Array.of(1, 2));\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// slice/concat with array receiver
			{
				Code:   "const foo = [1, 2, 3].slice();\nfoo.includes(1) || foo.includes(2);",
				Output: []string{"const foo = new Set([1, 2, 3].slice());\nfoo.has(1) || foo.has(2);"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "const foo = [1, 2, 3].concat(4);\nfoo.includes(1) || foo.includes(2);",
				Output: []string{"const foo = new Set([1, 2, 3].concat(4));\nfoo.has(1) || foo.has(2);"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "const items = [1, 2, 3];\nconst foo = items.slice();\nfoo.includes(1) || foo.includes(2);",
				Output: []string{"const items = [1, 2, 3];\nconst foo = new Set(items.slice());\nfoo.has(1) || foo.has(2);"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 2, 7)},
			},
			{
				Code:   "const items = [1, 2, 3];\nconst foo = items.concat(4);\nfoo.includes(1) || foo.includes(2);",
				Output: []string{"const items = [1, 2, 3];\nconst foo = new Set(items.concat(4));\nfoo.has(1) || foo.has(2);"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 2, 7)},
			},
			// `lodash` — bar is not array, but code not broken
			{
				Code:   "const foo = _([1,2,3]);\nconst bar = foo.map(value => value);\nfunction unicorn() {\n\treturn bar.includes(1);\n}",
				Output: []string{"const foo = _([1,2,3]);\nconst bar = new Set(foo.map(value => value));\nfunction unicorn() {\n\treturn bar.has(1);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("bar", 2, 7)},
			},

			// ---- Upstream: invalid (minimumItems: 5) ----
			{
				Code:    "const foo = [1, 2, 3, 4, 5];\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output:  []string{"const foo = new Set([1, 2, 3, 4, 5]);\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Options: minimumItems(5),
				Errors:  []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:    "const foo = Array.of(1, 2, 3, 4, 5);\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output:  []string{"const foo = new Set(Array.of(1, 2, 3, 4, 5));\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Options: minimumItems(5),
				Errors:  []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:    "const foo = Array.from({length: 5}, (_, index) => index);\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output:  []string{"const foo = new Set(Array.from({length: 5}, (_, index) => index));\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Options: minimumItems(5),
				Errors:  []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:    "const foo = Array.from({length: 2 + 3}, (_, index) => index);\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output:  []string{"const foo = new Set(Array.from({length: 2 + 3}, (_, index) => index));\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Options: minimumItems(5),
				Errors:  []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:    "const foo = Array.from([1, 2, 3, 4, 5]);\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output:  []string{"const foo = new Set(Array.from([1, 2, 3, 4, 5]));\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Options: minimumItems(5),
				Errors:  []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:    "const foo = Array.from('hello');\nfunction unicorn() {\n\treturn foo.includes('h');\n}",
				Output:  []string{"const foo = new Set(Array.from('hello'));\nfunction unicorn() {\n\treturn foo.has('h');\n}"},
				Options: minimumItems(5),
				Errors:  []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:    "const foo = Array(5);\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output:  []string{"const foo = new Set(Array(5));\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Options: minimumItems(5),
				Errors:  []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:    "const foo = Array(2 + 3);\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output:  []string{"const foo = new Set(Array(2 + 3));\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Options: minimumItems(5),
				Errors:  []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:    "const foo = Array(1, 2, 3, 4, 5);\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output:  []string{"const foo = new Set(Array(1, 2, 3, 4, 5));\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Options: minimumItems(5),
				Errors:  []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:    "const foo = new Array(5);\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output:  []string{"const foo = new Set(new Array(5));\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Options: minimumItems(5),
				Errors:  []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:    "const foo = new Array(1, 2, 3, 4, 5);\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
				Output:  []string{"const foo = new Set(new Array(1, 2, 3, 4, 5));\nfunction unicorn() {\n\treturn foo.has(1);\n}"},
				Options: minimumItems(5),
				Errors:  []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},

			// ---- Upstream: invalid (TypeScript, type-annotation autofix) ----
			{
				Code:   "const a: Array<'foo' | 'bar'> = ['foo', 'bar']\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.includes(someString)) {\n\t\tconsole.log(123)\n\t}\n}",
				Output: []string{"const a: Set<'foo' | 'bar'> = new Set(['foo', 'bar'])\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.has(someString)) {\n\t\tconsole.log(123)\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("a", 1, 7)},
			},
			{
				Code:   "const a: string[] = ['foo', 'bar']\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.includes(someString)) {\n\t\tconsole.log(123)\n\t}\n}",
				Output: []string{"const a: Set<string> = new Set(['foo', 'bar'])\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.has(someString)) {\n\t\tconsole.log(123)\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("a", 1, 7)},
			},
			{
				Code:   "const a: (string | number)[] = ['foo', 'bar']\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.includes(someString)) {\n\t\tconsole.log(123)\n\t}\n}",
				Output: []string{"const a: Set<string | number> = new Set(['foo', 'bar'])\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.has(someString)) {\n\t\tconsole.log(123)\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("a", 1, 7)},
			},
			{
				Code:   "const a: ReadonlyArray<string> = ['foo', 'bar']\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.includes(someString)) {\n\t\tconsole.log(123)\n\t}\n}",
				Output: []string{"const a: ReadonlySet<string> = new Set(['foo', 'bar'])\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.has(someString)) {\n\t\tconsole.log(123)\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("a", 1, 7)},
			},
			{
				Code:   "const a: readonly string[] = ['foo', 'bar']\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.includes(someString)) {\n\t\tconsole.log(123)\n\t}\n}",
				Output: []string{"const a: ReadonlySet<string> = new Set(['foo', 'bar'])\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.has(someString)) {\n\t\tconsole.log(123)\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("a", 1, 7)},
			},
			{
				Code:   "const a: readonly (string | number)[] = ['foo', 'bar']\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.includes(someString)) {\n\t\tconsole.log(123)\n\t}\n}",
				Output: []string{"const a: ReadonlySet<string | number> = new Set(['foo', 'bar'])\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.has(someString)) {\n\t\tconsole.log(123)\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("a", 1, 7)},
			},
			// Type annotation that can't be safely rewritten → suggestion only.
			{
				Code: "type Items = string[]\nconst a: Items = ['foo', 'bar']\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.includes(someString)) {\n\t\tconsole.log(123)\n\t}\n}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "error",
					Message:   "`a` should be a `Set`, and use `a.has()` to check existence or non-existence.",
					Line:      2,
					Column:    7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestion",
						Output:    "type Items = string[]\nconst a: Items = new Set(['foo', 'bar'])\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.has(someString)) {\n\t\tconsole.log(123)\n\t}\n}",
					}},
				}},
			},
			{
				Code: "const a: [string, string] = ['foo', 'bar']\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.includes(someString)) {\n\t\tconsole.log(123)\n\t}\n}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "error",
					Message:   "`a` should be a `Set`, and use `a.has()` to check existence or non-existence.",
					Line:      1,
					Column:    7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestion",
						Output:    "const a: [string, string] = new Set(['foo', 'bar'])\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.has(someString)) {\n\t\tconsole.log(123)\n\t}\n}",
					}},
				}},
			},
			{
				Code: "const a: string /* comment */ [] = ['foo', 'bar']\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.includes(someString)) {\n\t\tconsole.log(123)\n\t}\n}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "error",
					Message:   "`a` should be a `Set`, and use `a.has()` to check existence or non-existence.",
					Line:      1,
					Column:    7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestion",
						Output:    "const a: string /* comment */ [] = new Set(['foo', 'bar'])\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.has(someString)) {\n\t\tconsole.log(123)\n\t}\n}",
					}},
				}},
			},
			{
				Code:   "const foo: string[] = ['a', 'b']\nconst length = foo.length\n\nfunction has(value) {\n\treturn foo.includes(value)\n}",
				Output: []string{"const foo: Set<string> = new Set(['a', 'b'])\nconst length = foo.size\n\nfunction has(value) {\n\treturn foo.has(value)\n}"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
		},
	)
}
