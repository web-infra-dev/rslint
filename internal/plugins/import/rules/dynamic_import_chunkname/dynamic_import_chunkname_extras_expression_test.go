package dynamic_import_chunkname_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/dynamic_import_chunkname"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const destructuringInvalidSyntaxMessage = `dynamic imports require a "webpack" comment with valid syntax`

type expressionRow struct {
	expression string
	valid      bool
}

// expressionRows are expressions placed after `x:` in a magic comment that
// also names the chunk. Each row's validity was checked against the pinned
// upstream rule and against Node.js's vm.runInNewContext, which the upstream
// rule uses: they agree on every row. They cover parenthesized operands,
// typeof and delete, and the reads that assignment patterns perform (member
// objects, computed keys and default values that run) and do not perform
// (names that are only written, defaults that do not run).
var expressionRows = []expressionRow{
	{"typeof (missing)", true},
	{"typeof ((missing))", true},
	{"typeof missing", true},
	{"typeof missing.p", false},
	{"typeof (missing).p", false},
	{"typeof delete missing", true},
	{"delete missing", true},
	{"delete (missing)", true},
	{"delete ((missing))", true},
	{"delete missing.p", false},
	{"delete (missing).p", false},
	{"delete missing[0]", false},
	{"(delete missing)", true},
	{"(missing)", false},
	{"(missing).p", false},
	{"(missing) + 1", false},
	{"!(missing)", false},
	{"missing + 1", false},
	{"([flag = missing] = [])[0]", false},
	{"([flag = missing] = [1])[0]", true},
	{"([flag = missing] = [undefined])[0]", false},
	{"([flag = missing] = [void 0])[0]", false},
	{"([flag = missing] = [null])[0]", true},
	{"([flag = missing] = [0])[0]", true},
	{"([flag = missing] = [])", false},
	{"([flag = missing] = [, 1])", false},
	{"([, flag = missing] = [])", false},
	{"([, flag = missing] = [1, 2])", true},
	{"([...rest] = [missing])", false},
	{"([flag = missing, other = missing] = [1])", false},
	{"([flag = missing] = obj)", false},
	{"([flag = missing] = f())", false},
	{"([[inner = missing] = []] = [undefined])", false},
	{"([[inner = missing] = []] = [[1]])", true},
	{"([[inner = missing]] = [[]])", false},
	{"([flag = (missing)] = [])", false},
	{"([flag = missing] = [...xs])", false},
	{"([flag = missing] = [...xs, 1])", false},
	{"({a = missing} = {})", false},
	{"({a = missing} = {a: 1})", true},
	{"({a = missing} = {a: undefined})", false},
	{"({a = missing} = {a: null})", true},
	{"({a = missing} = {b: 1})", false},
	{"({a: b = missing} = {})", false},
	{"({a: b = missing} = {b: 1})", false},
	{"({[missing]: v} = {})", false},
	{"({[missing]: v} = undefined)", false},
	{"({[k]: v = missing} = {})", false},
	{"({...rest} = {a: missing})", false},
	{"({a: missing.p} = {})", false},
	{"({a: missing.p} = obj)", false},
	{"([missing.p] = [])", false},
	{"([missing[0]] = [])", false},
	{"(missing.p = 1)", false},
	{"(missing[0] = 1)", false},
	{"(missing = 1)", true},
	{"(fallback = 1)", true},
	{"([fallback] = [1])", true},
	{"({ fallback } = { fallback: 1 })", true},
	{"({ fallback = 1 } = {})", true},
	{"([fallback = 1] = [])", true},
	{"([fallback = other] = [1])", true},
	{"([a = (missing)] = [])", false},
	{"([(a) = missing] = [])", false},
	{"({a: (b) = missing} = {})", false},
	{"(missing ||= 1)", false},
	{"(fallback ||= 1)", false},
	{"({a: x} = missing)", false},
	{"([x] = missing)", false},
	{"({a = 1} = {a: missing})", false},
	{"([a] = [missing])", false},
	{"({a: b = 1} = {a: missing})", false},
	{"([a = 1] = [missing])", false},
	{"({a} = {a: missing})", false},
	{"([a = missing] = [undefined, 1])", false},
	{"([a = missing] = [1, undefined])", true},
	{"([a = missing, ...r] = [])", false},
	{"({a = missing, ...r} = {})", false},
	{"([{a = missing}] = [{}])", false},
	{"([{a = missing}] = [undefined])", false},
	{"({x: [y = missing]} = {x: []})", false},
	{"({x: [y = missing]} = {x: [1]})", true},
	{"({x = missing} = obj)", false},
	{"(a, missing) = 1", false},
	{"([a] = [missing], 1)", false},
	{"((missing)) = 1", true},
	{"([(missing)] = [])", true},
	{"({a: (missing)} = {})", true},
	{"({a = missing})", false},
	{"([a]) = []", false},
	{"({a: 1} = {})", false},
	{"(missing) = 1", true},
	{"({ get a() { return missing } } = {})", false},
}

func TestDynamicImportChunknameExpressionTable(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, row := range expressionRows {
		code := "import(\n  /* webpackChunkName: \"a\", x: " + row.expression + " */\n  'm',\n)"
		if row.valid {
			valid = append(valid, rule_tester.ValidTestCase{Code: code})
			continue
		}
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: code,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: destructuringInvalidSyntaxMessage, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
			},
		})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &dynamic_import_chunkname.DynamicImportChunknameRule, valid, invalid)
}
