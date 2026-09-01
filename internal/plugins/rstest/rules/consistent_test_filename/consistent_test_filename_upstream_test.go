// TestConsistentTestFilenameUpstream migrates the complete
// @vitest/eslint-plugin@v1.6.27 consistent-test-filename suite
// (tests/consistent-test-filename.test.ts) 1:1. Every case is a filename, so
// the code is irrelevant and stays as upstream wrote it. Position assertions
// cover the file-start range that carries the diagnostic. Extension coverage,
// path matching and option-parsing branches live in
// consistent_test_filename_extras_test.go.
package consistent_test_filename

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTestFilenameUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentTestFilenameRule,
		[]rule_tester.ValidTestCase{
			{Code: `export {}`, FileName: "1.test.ts"},
			{
				Code:     `export {}`,
				FileName: "1.spec.ts",
				Options:  map[string]any{"pattern": `.*\.spec\.ts$`},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `export {}`,
				FileName: "1.spec.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentTestFilename",
						Message:   `Use test file name pattern .*\.test\.(c|m)?[tj]sx?$`,
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 1,
					},
				},
			},
			{
				Code:     `export {}`,
				FileName: "__tests__/2.ts",
				Options: map[string]any{
					"allTestPattern": `__tests__`,
					"pattern":        `.*\.spec\.ts$`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentTestFilename",
						Message:   `Use test file name pattern .*\.spec\.ts$`,
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 1,
					},
				},
			},
		},
	)
}
