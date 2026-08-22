package prefer_ending_with_an_expect_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_ending_with_an_expect"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferEndingWithAnExpectExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. The upstream-mirrored cases live in
// prefer_ending_with_an_expect_upstream_test.go.
func TestPreferEndingWithAnExpectExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_ending_with_an_expect.PreferEndingWithAnExpectRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized callback and assertion expression ----
			{Code: `test('works', (((() => (((expect(1).toBe(1))))))));`},
			// ---- Dimension 4: optional assertion call chain ----
			{Code: `test('works', () => expect?.(1).toBe(1));`},
			// ---- Dimension 4: element-access assertion chain ----
			{Code: `test('works', () => tester['verify'](1));`, Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"tester.verify"}}}},
			// ---- Dimension 4: function-expression container ----
			{Code: `test('works', function () { expect(1).toBe(1); });`},
			// Locks in upstream create() branch: fewer than two arguments are ignored.
			{Code: `test('missing callback');`},
			// Locks in upstream create() branch: named callback references are ignored.
			{Code: `test('named callback', callback); function callback() {}`},
			// Locks in upstream getLastStatement() expression-body arm.
			{Code: `test('expression body', () => verify());`, Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"verify"}}}},
			// Locks in upstream matcher branch: a configured additional test wrapper is exact-match only.
			{Code: `suite.case('ignored', () => {});`, Options: []interface{}{map[string]interface{}{"additionalTestBlockFunctions": []interface{}{"suite.test"}}}},
			// ---- Real-user: eslint-plugin-jest#1752 awaited expect regression ----
			{Code: `test('awaited assertion', async () => { await expect(load()).resolves.toBe('ok'); });`},
			// ---- Real-user: eslint-plugin-jest#1742 SuperTest assertion chain ----
			{Code: `test('http assertion', () => request(app).get('/').expect(200));`, Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"request.**.expect"}}}},
		},
		[]rule_tester.InvalidTestCase{
			{
				// ---- Dimension 4: empty function body gracefully reports ----
				Code: `test('empty', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mustEndWithExpect",
					Message:   "Tests should end with an assertion",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 5,
				}},
			},
			{
				// ---- Dimension 4: return statement is not a direct assertion call ----
				Code:   `test('return', () => { return expect(1).toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "mustEndWithExpect", Line: 1, Column: 1, EndLine: 1, EndColumn: 5}},
			},
			{
				// Locks in upstream await arm: only an awaited assertion is accepted.
				Code: `
test('multiline', async () => {
  await doWork();
});`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "mustEndWithExpect", Line: 2, Column: 1, EndLine: 2, EndColumn: 5}},
			},
			{
				// Locks in upstream additionalTestBlockFunctions inclusion arm.
				Code:    `suite.test('custom', () => {});`,
				Options: []interface{}{map[string]interface{}{"additionalTestBlockFunctions": []interface{}{"suite.test"}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "mustEndWithExpect", Line: 1, Column: 1, EndLine: 1, EndColumn: 11}},
			},
		},
	)
}
