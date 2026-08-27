// TestRequireAwaitedExpectPollUpstream migrates the complete
// @vitest/eslint-plugin@v1.6.27 require-awaited-expect-poll suite 1:1.
// Position assertions cover line/column for every invalid case. Rstest-only
// lock-in cases — the expect source matrix, the handled-position extension,
// Dimension 4 shapes and the branch walk — live in
// require_awaited_expect_poll_extras_test.go.
package require_awaited_expect_poll

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireAwaitedExpectPollUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAwaitedExpectPollRule,
		[]rule_tester.ValidTestCase{
			// ---- awaited expect.poll ----
			{Code: `
        test('should pass', async () => {
          await expect.poll(() => element).toBeInTheDocument();
        });
      `},
			// ---- awaited expect.element ----
			{Code: `
        test('should pass', async () => {
          await expect.element(element).toBeInTheDocument();
        });
      `},
			// ---- non-poll method ----
			{Code: `
        test('should pass', () => {
          expect.syncElement(element).toBeInTheDocument();
        });
      `},
			// ---- returned expect.poll ----
			{Code: `
        test('should pass', () => {
          return expect.poll(() => element).toBeInTheDocument();
        });
      `},
			// ---- returned expect.element ----
			{Code: `
        test('should pass', () => {
          return expect.element(element).toBeInTheDocument();
        });
      `},
			// ---- expect without method ----
			{Code: `
        test('should pass', () => {
          return expect(true).toBe(true);
        });
      `},
			// ---- awaited inside SequenceExpression ----
			{Code: `
        test('should pass', async () => {
          (sideEffect(), await expect.poll(() => element).toBeInTheDocument());
        });
      `},
			// ---- awaited outside SequenceExpression ----
			{Code: `
        test('should pass', async () => {
          await (sideEffect(), expect.poll(() => element).toBeInTheDocument());
        });
      `},
			// ---- awaited outside multiple SequenceExpressions ----
			{Code: `
        test('should pass', async () => {
          await (sideEffect(), (sideEffect(), (sideEffect(), expect.poll(() => element).toBeInTheDocument())));
        });
      `},
			// ---- returned from SequenceExpression ----
			{Code: `
        test('should pass', () => {
          return (sideEffect(), expect.poll(() => element).toBeInTheDocument());
        });
      `},
		},
		[]rule_tester.InvalidTestCase{
			// ---- expect.poll not awaited ----
			{
				Code: `
        test('should fail', () => {
          expect.poll(() => element).toBeInTheDocument();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notAwaited",
					Message:   "`expect.poll` calls should be awaited",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 22,
				}},
			},
			// ---- expect.element not awaited ----
			{
				Code: `
        test('should fail', () => {
          expect.element(element).toBeInTheDocument();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notAwaited",
					Message:   "`expect.element` calls should be awaited",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 25,
				}},
			},
			// ---- expect.poll not awaited - accessed with bracket notation ----
			{
				Code: `
        test('should fail', () => {
          expect['poll'](() => element).toBeInTheDocument();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notAwaited",
					Message:   "`expect.poll` calls should be awaited",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 25,
				}},
			},
			// ---- expect.element not awaited - accessed with bracket notation ----
			{
				Code: `
        test('should fail', () => {
          expect['element'](element).toBeInTheDocument();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notAwaited",
					Message:   "`expect.element` calls should be awaited",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 28,
				}},
			},
			// ---- expect.poll not awaited - inside SequenceExpression ----
			{
				Code: `
        test('should fail', () => {
          (expect.poll(() => element).toBeInTheDocument(), expect(true).toBe(true));
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notAwaited",
					Message:   "`expect.poll` calls should be awaited",
					Line:      3,
					Column:    12,
					EndLine:   3,
					EndColumn: 23,
				}},
			},
			// ---- expect.element not awaited - inside SequenceExpression ----
			{
				Code: `
        test('should fail', () => {
          (expect.element(() => element).toBeInTheDocument(), expect(true).toBe(true));
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notAwaited",
					Message:   "`expect.element` calls should be awaited",
					Line:      3,
					Column:    12,
					EndLine:   3,
					EndColumn: 26,
				}},
			},
			// ---- expect.poll returned as part of SequenceExpression ----
			{
				Code: `
        test('should fail', () => {
          return (expect.poll(() => element).toBeInTheDocument(), expect(true).toBe(true));
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notAwaited",
					Message:   "`expect.poll` calls should be awaited",
					Line:      3,
					Column:    19,
					EndLine:   3,
					EndColumn: 30,
				}},
			},
		},
	)
}
