package prefer_each_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_each"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func preferEachAt(fn string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "preferEach",
		Message:   "prefer using `" + fn + ".each` rather than a manual loop",
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

// TestPreferEachLoopScoping pins down which loop a registration belongs to and
// what the resulting message names. Each loop is judged from its own frame, so
// a loop is reported for what it registers and for nothing else, no matter how
// deeply it sits inside test callbacks.
func TestPreferEachLoopScoping(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_each.PreferEachRule,
		[]rule_tester.ValidTestCase{
			// Registrations in a loop header run once, not once per iteration, so a
			// loop whose header alone registers is not a candidate for `.each`.
			{Code: `for (const r of (beforeEach(() => {}), rows)) {}`},
			{Code: `for (const r of (function () { it("x", () => {}); return rows; })()) {}`},
			// Registrations in classic for control clauses do not belong to the
			// loop body and cannot be replaced with `.each`.
			{Code: `for (let i = register(it('once', () => {})); i < 2; i++) {}`},
			{Code: `for (let i = 0; check(it('condition', () => {}), i < 2); i++) {}`},
			{Code: `for (let i = 0; i < 2; step(it('update', () => {}), i++)) {}`},
			// A loop that registers nothing is business logic wherever it sits,
			// including after a nested registration in the same callback.
			{Code: `test('outer', () => {
  test('inner', () => {});
  for (const row of rows) {
    consume(row);
  }
});`},
			// Same, for a callback passed by name.
			{Code: `const cb = () => {
  for (const row of rows) {
    consume(row);
  }
};
test('a', cb);`},
			// The enclosing loop registers nothing of its own.
			{Code: `for (const row of rows) {
  consume(row);
  for (const item of row.items) {
    consume(item);
  }
}`},
		},
		[]rule_tester.InvalidTestCase{
			// Only the body registers once per iteration, so the header's `it` does
			// not turn the recommendation into `describe.each`.
			{
				Code: `for (const r of (function () { it("x", () => {}); return rows; })()) { it(r, () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEach", Message: "prefer using `it.each` rather than a manual loop", Line: 1, Column: 1, EndLine: 1, EndColumn: 90},
				},
			},
			// The inner loop's body registers nothing; the loop that registers on
			// every iteration is the outer one, through the inner loop's header.
			{
				Code: `for (const a of as) { for (const b of (function () { it("x", () => {}); return bs; })()) { consume(b); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEach", Message: "prefer using `it.each` rather than a manual loop", Line: 1, Column: 1, EndLine: 1, EndColumn: 107},
				},
			},
			// Two directly nested loops that each register are each reported; the
			// inner loop starting does not discard what the outer loop registered.
			{
				Code: `for (const a of as) { it("p", () => {}); for (const b of bs) { it("q", () => {}); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEach", Message: "prefer using `it.each` rather than a manual loop", Line: 1, Column: 42, EndLine: 1, EndColumn: 84},
					{MessageId: "preferEach", Message: "prefer using `it.each` rather than a manual loop", Line: 1, Column: 1, EndLine: 1, EndColumn: 86},
				},
			},
			// ---- a registering loop inside a test callback is still a registering loop ----
			{
				Code: `test('outer', () => {
  for (const row of rows) {
    test(row.name, () => {});
  }
});`,
				Errors: []rule_tester.InvalidTestCaseError{preferEachAt("test", 2, 3, 4, 4)},
			},
			{
				Code: `test('outer', async () => {
  for (const row of rows) {
    test(row.name, () => {});
  }
});`,
				Errors: []rule_tester.InvalidTestCaseError{preferEachAt("test", 2, 3, 4, 4)},
			},
			{
				Code: `test('outer', function () {
  for (const row of rows) {
    it(row.name, () => {});
  }
});`,
				Errors: []rule_tester.InvalidTestCaseError{preferEachAt("it", 2, 3, 4, 4)},
			},
			{
				Code: `test('outer', (done) => {
  for (const row of rows) {
    test(row.name, () => {});
  }
  done();
});`,
				Errors: []rule_tester.InvalidTestCaseError{preferEachAt("test", 2, 3, 4, 4)},
			},
			{
				// The callback is passed by name, so the loop is lexically at top
				// level; either way it registers once per row.
				Code: `const cb = () => {
  for (const row of rows) {
    test(row.name, () => {});
  }
};
test('a', cb);`,
				Errors: []rule_tester.InvalidTestCaseError{preferEachAt("test", 2, 3, 4, 4)},
			},
			{
				Code: `test.each(rows)('%s', () => {
  for (const row of rows) {
    test(row.name, () => {});
  }
});`,
				Errors: []rule_tester.InvalidTestCaseError{preferEachAt("test", 2, 3, 4, 4)},
			},
			{
				Code: `test('outer', () => {
  for (const row of rows) {
    describe(row.name, () => {});
  }
});`,
				Errors: []rule_tester.InvalidTestCaseError{preferEachAt("describe", 2, 3, 4, 4)},
			},

			// ---- registrations nested in a test callback belong to the loop around it ----
			{
				Code: `for (const row of rows) {
  test(row.name, () => {
    beforeEach(() => {});
  });
}`,
				Errors: []rule_tester.InvalidTestCaseError{preferEachAt("describe", 1, 1, 5, 2)},
			},
			{
				Code: `for (const row of rows) {
  test(row.name, () => {
    test('inner', () => {});
  });
}`,
				Errors: []rule_tester.InvalidTestCaseError{preferEachAt("describe", 1, 1, 5, 2)},
			},
			{
				Code: `for (const row of rows) {
  describe(row.name, () => {
    test('inner', () => {});
  });
}`,
				Errors: []rule_tester.InvalidTestCaseError{preferEachAt("describe", 1, 1, 5, 2)},
			},

			// ---- a function in a non-callback argument position is not special ----
			{
				Code: `test('a', () => {}, (function () {
  for (const row of rows) {
    beforeEach(() => {});
  }
  return 5;
})());`,
				Errors: []rule_tester.InvalidTestCaseError{preferEachAt("describe", 2, 3, 4, 4)},
			},
			{
				// The inner iterable runs once per outer iteration, so its
				// registration belongs to the outer loop, not the inner loop.
				Code: `for (const suite of suites) {
  for (const row of getRows(test(suite.name, () => {}))) {
    consume(row);
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{preferEachAt("test", 1, 1, 5, 2)},
			},
			{
				// An inner classic for initializer likewise runs once per outer
				// iteration and therefore belongs to the outer loop.
				Code: `for (const suite of suites) {
  for (let i = register(test(suite.name, () => {})); i < 1; i++) {
    consume(i);
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{preferEachAt("test", 1, 1, 5, 2)},
			},
		},
	)
}
