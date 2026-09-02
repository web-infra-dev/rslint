// TestConsistentTestFilenameExtras covers what the migrated suite does not:
// the extensions Rstest's default `include` glob accepts, matching against the
// whole path rather than the basename, and every option-parsing branch.
//
// Dimension 4 (universal edge shapes: parenthesized expressions, optional
// chains, literal kinds, type wrappers, computed keys) is N/A for this rule —
// it never reads a node. Its only inputs are the file's path and the two
// options, so the shapes below enumerate paths and option payloads instead.
package consistent_test_filename

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// The diagnostic spans the whole file, so its end column depends on the code
// under test: `endColumn` is the length of the single-line source plus one.
func expectedDefaultPatternError(endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{
		{
			MessageId: "consistentTestFilename",
			Message:   `Use test file name pattern .*\.test\.(c|m)?[tj]sx?$`,
			Line:      1,
			Column:    1,
			EndLine:   1,
			EndColumn: endColumn,
		},
	}
}

func TestConsistentTestFilenameExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentTestFilenameRule,
		[]rule_tester.ValidTestCase{
			// ---- not a test file: allTestPattern gates everything ----
			{Code: `export {}`, FileName: "src/user-service.ts"},
			{Code: `export {}`, FileName: "src/user-service.tsx"},
			// An in-source test is named like the source file it lives in, so
			// the rule never looks at it.
			{Code: `if (import.meta.rstest) { test('adds', () => {}) }`, FileName: "src/add.ts"},
			// `.test` in a directory name, not in the filename.
			{Code: `export {}`, FileName: "test/user-service.ts"},
			// `test` as a bare basename, without the dotted segment.
			{Code: `export {}`, FileName: "src/test.ts"},

			// ---- extensions the default `include` glob accepts ----
			{Code: `export {}`, FileName: "src/user-service.test.ts"},
			{Code: `export {}`, FileName: "src/user-service.test.tsx"},
			{Code: `export {}`, FileName: "src/user-service.test.mts"},
			{Code: `export {}`, FileName: "src/user-service.test.cts"},

			// ---- the required pattern is matched against the whole path ----
			{Code: `export {}`, FileName: "packages/core/tests/user-service.test.ts"},

			// ---- option parsing ----
			// Only `allTestPattern` given: `pattern` keeps its default.
			{
				Code:     `export {}`,
				FileName: "src/user-service.test.ts",
				Options:  map[string]any{"allTestPattern": `.*\.test\.ts$`},
			},
			// An empty option object is the same as no options at all.
			{Code: `export {}`, FileName: "src/user-service.test.ts", Options: map[string]any{}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- the convention the defaults enforce ----
			{Code: `export {}`, FileName: "src/user-service.spec.ts", Errors: expectedDefaultPatternError(10)},
			{Code: `export {}`, FileName: "src/user-service.spec.tsx", Errors: expectedDefaultPatternError(10)},
			// `.mts` and `.cts` are test files under Rstest's default glob, so
			// a spec-named one is held to the same convention.
			{Code: `export {}`, FileName: "src/user-service.spec.mts", Errors: expectedDefaultPatternError(10)},
			{Code: `export {}`, FileName: "src/user-service.spec.cts", Errors: expectedDefaultPatternError(10)},

			// ---- whole-path matching selects a directory ----
			{
				Code:     `export {}`,
				FileName: "src/__tests__/user-service.ts",
				Options: map[string]any{
					"allTestPattern": `__tests__`,
					"pattern":        `.*\.test\.ts$`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentTestFilename",
						Message:   `Use test file name pattern .*\.test\.ts$`,
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 10,
					},
				},
			},

			// ---- the convention can be inverted ----
			{
				Code:     `export {}`,
				FileName: "src/user-service.test.ts",
				Options:  map[string]any{"pattern": `.*\.spec\.ts$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentTestFilename",
						Message:   `Use test file name pattern .*\.spec\.ts$`,
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 10,
					},
				},
			},

			// ---- the file's contents never matter ----
			{
				Code:     `export const users = []`,
				FileName: "src/user-service.spec.ts",
				Errors:   expectedDefaultPatternError(24),
			},
		},
	)
}
