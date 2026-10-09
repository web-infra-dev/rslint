package dynamic_import_chunkname_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/dynamic_import_chunkname"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Destructuring defaults that read inherited properties, and array patterns that destructure objects with custom iterators.
// Each row is an expression placed after `x:` in a magic comment that also
// names the chunk. Validity was recorded from the pinned upstream rule and
// checked against Node.js's vm.runInNewContext, which the rule uses.

var valueInheritedRows = []classSyntaxRow{
	{"({toString: value = missing} = {})", true},
	{"({constructor: v = missing} = {})", true},
	{"({hasOwnProperty: v = missing} = {})", true},
	{"({valueOf: v = missing} = {})", true},
	{"({isPrototypeOf: v = missing} = {})", true},
	{"({propertyIsEnumerable: v = missing} = {})", true},
	{"({toLocaleString: v = missing} = {})", true},
	{"({__proto__: v = missing} = {})", true},
	{"({__defineGetter__: v = missing} = {})", true},
	{"({a: value = missing} = {__proto__: {a: 1}})", true},
	{"({a: value = missing} = {'__proto__': {a: 1}})", true},
	{"({a: value = missing} = {a: undefined})", false},
	{"({a: value = missing} = {})", false},
	{"({a = missing} = {b: 1})", false},
	{"({toString = missing} = {})", true},
	{"({a: value = missing} = {a: 1, __proto__: null})", true},
	{"({['toString']: v = missing} = {})", true},
	{"({'toString': v = missing} = {})", true},
	{"({a: {toString: v = missing}} = {a: {}})", true},
	{"({toString: v = missing} = {toString: undefined})", false},
	{"({toString: v = missing} = {toString: 1})", true},
}

var valueIterableRows = []classSyntaxRow{
	{"([] = {[Symbol.iterator]: function* () {}})", true},
	{"([a] = {[Symbol.iterator]: function* () { yield 1 }})", true},
	{"([] = class { static *[Symbol.iterator]() {} })", true},
	{"([] = {})", false},
	{"([] = {a: 1})", false},
	{"([] = {a() {}})", false},
	{"([] = {get a() { return 1 }})", false},
	{"([] = {__proto__: []})", true},
	{"([] = {__proto__: {[Symbol.iterator]: function* () {}}})", true},
	{"([] = 1)", false},
	{"([] = true)", false},
	{"([] = null)", false},
	{"([] = undefined)", false},
	{"([] = function () {})", false},
	{"([] = () => {})", false},
	{"([] = class {})", false},
	{"([] = /x/)", false},
	{"([] = 'ab')", true},
	{"([] = `ab`)", true},
	{"([] = [])", true},
	{"([] = new Set())", true},
	{"({} = null)", false},
	{"({} = undefined)", false},
	{"({} = 1)", true},
	{"({} = {})", true},
	{"({} = [])", true},
	{"([[]] = [{}])", false},
	{"([[]] = [{[Symbol.iterator]: function* () {}}])", true},
	{"([a = missing] = [{}])", true},
	{"([[a = missing]] = [[]])", false},
}

func TestDynamicImportChunknameDestructuringValueTables(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, rows := range [][]classSyntaxRow{valueInheritedRows, valueIterableRows} {
		for _, row := range rows {
			code := "import(\n  /* webpackChunkName: \"a\", x: " + row.expression + " */\n  'm',\n)"
			if row.valid {
				valid = append(valid, rule_tester.ValidTestCase{Code: code})
				continue
			}
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code: code,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: classInvalidSyntaxMessage, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &dynamic_import_chunkname.DynamicImportChunknameRule, valid, invalid)
}
