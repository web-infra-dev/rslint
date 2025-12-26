package ban_tslint_comment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestBanTslintCommentRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&BanTslintCommentRule,
		// Valid cases - comments that should NOT be flagged
		[]rule_tester.ValidTestCase{
			// Valid TypeScript code
			{Code: `let a: readonly any[] = [];`},
			{Code: `let a = new Array();`},

			// Regular comments mentioning tslint (not directives)
			{Code: `// some other comment`},
			{Code: `// TODO: this is a comment that mentions tslint`},
			{Code: `/* another comment that mentions tslint */`},
			{Code: `// This project used to use tslint`},
			{Code: `/* We migrated from tslint to eslint */`},

			// Comments that don't match the directive pattern
			{Code: `// tslint is deprecated`},
			{Code: `/* tslint was a linter */`},
			{Code: `// about tslint:disable`},
			{Code: `/* discussing tslint:enable */`},
		},
		// Invalid cases - TSLint directives that should be flagged
		[]rule_tester.InvalidTestCase{
			// Basic tslint:disable
			{
				Code: `/* tslint:disable */`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// Basic tslint:enable
			{
				Code: `/* tslint:enable */`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// tslint:disable with specific rules
			{
				Code: `/* tslint:disable:rule1 rule2 rule3... */`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// tslint:enable with specific rules
			{
				Code: `/* tslint:enable:rule1 rule2 rule3... */`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// Single-line comment: tslint:disable-next-line
			{
				Code: `// tslint:disable-next-line`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// Inline tslint:disable-line
			{
				Code: `someCode(); // tslint:disable-line`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 13},
				},
			},

			// tslint:disable-next-line with specific rules
			{
				Code: `// tslint:disable-next-line:rule1 rule2 rule3...`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// Multi-line code with tslint:disable-line
			{
				Code: `if (true) {
  console.log("test");
}
// tslint:disable-line
const x = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 4, Column: 1},
				},
			},

			// tslint:enable-line
			{
				Code: `// tslint:enable-line`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// Multiple spaces before directive
			{
				Code: `//    tslint:disable`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// Block comment with spaces
			{
				Code: `/*   tslint:disable   */`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// tslint:disable with colon separator
			{
				Code: `// tslint:disable:no-console`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// tslint:enable with colon separator
			{
				Code: `// tslint:enable:no-console`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// Tab character before directive
			{
				Code: "//\ttslint:disable",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// Multiple tslint comments in one file
			{
				Code: `// tslint:disable
const x = 1;
// tslint:enable`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
					{MessageId: "commentDetected", Line: 3, Column: 1},
				},
			},

			// tslint:disable-next-line before code
			{
				Code: `// tslint:disable-next-line
const value = "test";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// Block comment tslint:disable-next-line
			{
				Code: `/* tslint:disable-next-line */
const value = "test";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// tslint directive with 's' suffix (alternative format)
			{
				Code: `// tslint:disables`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},

			// tslint directive with 's' suffix for enable
			{
				Code: `// tslint:enables`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentDetected", Line: 1, Column: 1},
				},
			},
		},
	)
}
