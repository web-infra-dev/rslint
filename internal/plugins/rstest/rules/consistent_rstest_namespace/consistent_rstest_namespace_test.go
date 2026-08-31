package consistent_rstest_namespace

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

var (
	rsOption     = []any{map[string]any{"fn": "rs"}}
	rstestOption = []any{map[string]any{"fn": "rstest"}}
)

func namespaceError(preferred, disallowed string, line, column, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "consistentNamespace",
		Message:   "Prefer using `" + preferred + "` instead of `" + disallowed + "`",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}
}

func TestConsistentRstestNamespace(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentRstestNamespaceRule,
		[]rule_tester.ValidTestCase{
			// The preferred spelling, imported and global.
			{Code: `import { rs } from '@rstest/core';
rs.mock('./service');`},
			{Code: `rs.clearAllMocks();`},
			{Code: `import { rstest } from '@rstest/core';
rstest.mock('./service');`, Options: rstestOption},
			{Code: `rstest.clearAllMocks();`, Options: rstestOption},

			// An alias names the namespace something else entirely.
			{Code: `import { rstest as testUtils } from '@rstest/core';
testUtils.mock('./service');`},

			// The same name bound by something that is not the Rstest namespace.
			{Code: `import { rstest } from './helpers';
rstest.mock('./service');`},
			{Code: `import * as rstest from '@rstest/core';
rstest.mock('./service');`},
			{Code: `const rstest = require('@rstest/core');
rstest.mock('./service');`},
			{Code: `function run(rstest: { mock(path: string): void }) {
  rstest.mock('./service');
}`},
			{Code: `const rstest = { mock(path: string) {} };
rstest.mock('./service');`},

			// A property that happens to share the name.
			{Code: `const config = { rstest: { mock(path: string) {} } };
config.rstest.mock('./service');`},

			// Reads that are not calls through the namespace.
			{Code: `const spy = rstest.fn;`},

			// import.meta.rstest is not one of the two spellings.
			{Code: `import.meta.rstest.mock('./service');`},

			// Type-only bindings introduce no value to rewrite.
			{Code: `import type { rstest } from '@rstest/core';`},
			{Code: `import { type rstest } from '@rstest/core';`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `import { rstest } from '@rstest/core';
rstest.mock('./service');
rstest.clearAllMocks();`,
				Output: []string{`import { rs } from '@rstest/core';
rs.mock('./service');
rs.clearAllMocks();`},
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 10, 16),
					namespaceError("rs", "rstest", 2, 1, 7),
					namespaceError("rs", "rstest", 3, 1, 7),
				},
			},
			{
				Code:   `import { expect, rs, rstest } from '@rstest/core';`,
				Output: []string{`import { expect, rs } from '@rstest/core';`},
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 22, 28),
				},
			},
			{
				Code:   `import { rstest, rs, expect } from '@rstest/core';`,
				Output: []string{`import { rs, expect } from '@rstest/core';`},
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 10, 16),
				},
			},
			{
				// A comment written for the specifier that stays survives the fix.
				Code:   `import { rstest, /* the namespace */ rs } from '@rstest/core';`,
				Output: []string{`import { /* the namespace */ rs } from '@rstest/core';`},
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 10, 16),
				},
			},
			{
				// A trailing comment belongs to the specifier that is removed.
				Code:   `import { rstest /* remove with namespace */, rs } from '@rstest/core';`,
				Output: []string{`import { rs } from '@rstest/core';`},
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 10, 16),
				},
			},
			{
				// The reverse preference.
				Code: `import { rs } from '@rstest/core';
rs.mock('./service');`,
				Output: []string{`import { rstest } from '@rstest/core';
rstest.mock('./service');`},
				Options: rstestOption,
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rstest", "rs", 1, 10, 12),
					namespaceError("rstest", "rs", 2, 1, 3),
				},
			},
			{
				// A global namespace needs no import to rewrite.
				Code:    `rstest.useFakeTimers();`,
				Output:  []string{`rs.useFakeTimers();`},
				Options: rsOption,
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 1, 7),
				},
			},
			{
				// Optional chaining, arguments and the method name are preserved.
				Code:   `rstest?.mock('./service', () => ({ pay: rstest.fn() }));`,
				Output: []string{`rs?.mock('./service', () => ({ pay: rs.fn() }));`},
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 1, 7),
					namespaceError("rs", "rstest", 1, 41, 47),
				},
			},
			{
				// One namespace object heading a chain of calls reports once.
				Code:   `rstest.mocked(pay).mockReturnValue(true);`,
				Output: []string{`rs.mocked(pay).mockReturnValue(true);`},
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 1, 7),
				},
			},
			{
				// TypeScript expression wrappers do not hide the namespace call.
				Code: `import { rstest } from '@rstest/core';
rstest!.mock('./service');`,
				Output: []string{`import { rs } from '@rstest/core';
rs!.mock('./service');`},
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 10, 16),
					namespaceError("rs", "rstest", 2, 1, 7),
				},
			},
			{
				// The runtime root is reported through an assertion, but its type
				// query prevents a partial binding rewrite.
				Code: `import { rstest } from '@rstest/core';
(rstest as typeof rstest).mock('./service');`,
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 10, 16),
					namespaceError("rs", "rstest", 2, 2, 8),
				},
			},
			{
				// A member invoked as a tagged template is still a namespace use.
				Code:   "import { rstest } from '@rstest/core';\nrstest.fn`service`;",
				Output: []string{"import { rs } from '@rstest/core';\nrs.fn`service`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 10, 16),
					namespaceError("rs", "rstest", 2, 1, 7),
				},
			},
			{
				// A use the rewrite does not reach leaves the import unfixed.
				Code: `import { rstest } from '@rstest/core';
const { fn } = rstest;
rstest.mock('./service');`,
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 10, 16),
					namespaceError("rs", "rstest", 3, 1, 7),
				},
			},
			{
				// Re-exporting the binding is such a use as well.
				Code: `import { rstest } from '@rstest/core';
export { rstest };
rstest.mock('./service');`,
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 10, 16),
					namespaceError("rs", "rstest", 3, 1, 7),
				},
			},
			{
				// An aliased re-export names the binding just as a plain one does.
				Code: `import { rstest } from '@rstest/core';
export { rstest as helper };
rstest.mock('./service');`,
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 10, 16),
					namespaceError("rs", "rstest", 3, 1, 7),
				},
			},
			{
				// A type-only re-export keeps the binding live as well.
				Code: `import { rstest } from '@rstest/core';
export type { rstest as RstestNamespace };
rstest.mock('./service');`,
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 10, 16),
					namespaceError("rs", "rstest", 3, 1, 7),
				},
			},
			{
				// A destructured require binds the namespace under a name the
				// import fix never reaches, so the calls are reported unfixed.
				Code: `const { rstest } = require('@rstest/core');
rstest.mock('./service');`,
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 2, 1, 7),
				},
			},
			{
				// Dropping the only specifier would leave a side-effect import.
				Code: `import { rs } from '@rstest/core';
import { rstest } from '@rstest/core';`,
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 2, 10, 16),
				},
			},
			{
				// An aliased specifier next to the reported one keeps its alias.
				Code:   `import { expect as check, rs, rstest } from '@rstest/core';`,
				Output: []string{`import { expect as check, rs } from '@rstest/core';`},
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 31, 37),
				},
			},
			{
				// A trailing comment on a specifier in the middle of the list goes
				// with it, and the comma that separates the two that stay remains.
				Code:   `import { rs, rstest /* remove with namespace */, expect } from '@rstest/core';`,
				Output: []string{`import { rs, expect } from '@rstest/core';`},
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 14, 20),
				},
			},
			{
				// The same for the last specifier of the list.
				Code:   `import { rs, rstest /* remove with namespace */ } from '@rstest/core';`,
				Output: []string{`import { rs } from '@rstest/core';`},
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 14, 20),
				},
			},
			{
				// A comment past the separating comma belongs to the specifier
				// that follows it, so it stays.
				Code:   `import { rs, rstest, /* the assertion */ expect } from '@rstest/core';`,
				Output: []string{`import { rs, /* the assertion */ expect } from '@rstest/core';`},
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 14, 20),
				},
			},
			{
				// A top-level binding of the preferred spelling would be declared
				// twice by the rename, so the report carries no fix.
				Code: `import { rstest } from '@rstest/core';
const rs = { mock(path: string) {} };
rstest.mock('./service');`,
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 10, 16),
					namespaceError("rs", "rstest", 3, 1, 7),
				},
			},
			{
				// A binding in a nested scope would capture the rewritten call.
				Code: `import { rstest } from '@rstest/core';
function run() {
  const rs = { mock(path: string) {} };
  rstest.mock('./service');
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 1, 10, 16),
					namespaceError("rs", "rstest", 4, 3, 9),
				},
			},
			{
				// A global namespace is captured by such a binding just the same.
				Code: `function run() {
  const rs = { mock(path: string) {} };
  rstest.mock('./service');
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 3, 3, 9),
				},
			},
			{
				// The name taken by an import of something else is taken all the
				// same.
				Code: `import { rs } from './helpers';
import { rstest } from '@rstest/core';
rstest.mock('./service');`,
				Errors: []rule_tester.InvalidTestCaseError{
					namespaceError("rs", "rstest", 2, 10, 16),
					namespaceError("rs", "rstest", 3, 1, 7),
				},
			},
		},
	)
}
