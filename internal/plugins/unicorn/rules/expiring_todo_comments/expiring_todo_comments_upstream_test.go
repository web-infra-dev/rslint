// cspell:ignore Welp somethingrandom HUGETODO
// Ported from eslint-plugin-unicorn v76.0.0 test/expiring-todo-comments.js
// and docs/rules/expiring-todo-comments.md. RegExp ignore options are represented
// by equivalent string patterns because native options cross a JSON boundary.
package expiring_todo_comments

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func upstreamRoot(t *testing.T) rule_tester.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	archive := txtarfs.MustParseFile(t, "testdata/packages.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(root.Dir, name)] = string(data)
	}
	return rule_tester.Root{Dir: root.Dir, FS: utils.NewOverlayVFS(root.FS, files)}
}

// The Go tester registers the rule as "test"; JS keeps the public rule ID.
func TestExpiringTodoCommentsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(upstreamRoot(t), "tsconfig.json", t, &ExpiringTodoCommentsRule,
		[]rule_tester.ValidTestCase{
			{
				Code: "// TODO [2200-12-12]: Too long... Can you feel it?"},
			{
				Code: "// FIXME [2200-12-12]: Too long... Can you feel it?"},
			{
				Code: "// XXX [2200-12-12]: Too long... Can you feel it?"},
			{
				Code: "// TODO (lubien) [2200-12-12]: Too long... Can you feel it?"},
			{
				Code: "// FIXME [2200-12-12] (lubien): Too long... Can you feel it?"},
			{
				Code:    "// Expire Condition [2200-12-12]: new term name",
				Options: []any{map[string]any{"terms": []any{"Expire Condition"}}}},
			{
				Code: "// Expire Condition [2000-01-01]: new term name"},
			{
				Code: "// TODO [>2000]: We sure didn't past this version"},
			{
				Code: "// TODO [>1000]: partial version with > should use semver range semantics"},
			{
				Code: "// TODO [find-up-simple@>1]: find-up-simple is 1.0.1 so >1 should not trigger"},
			{
				Code: "// TODO [engine:node@>22]: node engine is 22.x so >22 should not trigger"},
			{
				Code: "// TODO [peer:eslint@>=11]: peer eslint floor is 10.x so >=11 should not trigger"},
			{
				Code: "// TODO [peer:eslint@>10]: `>10` means `>=11.0.0`, so the 10.x floor should not trigger"},
			{
				Code: "// TODO [peer:does-not-exist@>=1]: a peer dependency we don't declare never triggers"},
			{
				Code: "// TODO [peer:find-up-simple@>=1]: `find-up-simple` is a dependency but not a peer dependency, so `peer:` never triggers"},
			{
				Code: "// TODO [-find-up-simple]: We actually use this."},
			{
				Code: "// TODO [+popura]: I think we won't need a broken package."},
			{
				Code:    "// TODO [+popura-2000-01-01]: A broken package with a date-like name.",
				Options: []any{map[string]any{"checkDates": true}}},
			{
				Code: "// TODO [semver@>1000]: Welp hopefully we won't get at that."},
			{
				Code: "// TODO [semver@>=1000]: Welp hopefully we won't get at that."},
			{
				Code: "// TODO [@lubien/fixture-beta-package@>=1.0.0]: we are using a pre-release"},
			{
				Code: "// TODO [@lubien/fixture-beta-package@>=1.0.0-gamma.1]: beta comes first from gamma"},
			{
				Code: "// TODO [@lubien/fixture-beta-package@>=1.0.0-beta.2]: we are in beta.1"},
			{
				Code: "// TODO [2200-12-12, -find-up-simple]: Combo"},
			{
				Code: "// TODO [2200-12-12, -find-up-simple, +popura]: Combo"},
			{
				Code: "// TODO [2200-12-12, -find-up-simple, +popura, semver@>=1000]: Combo"},
			{
				Code: "// TODO [engine:node@>=100]: When we start supporting only >= 10"},
			{
				Code: "// TODO [2000-01-01]: Expired dates are ignored by default"},
			{
				Code: "// TODO [2200-12-12]: Multiple\n\t\t// TODO [2200-12-12]: Lines"},
			{
				Code: "/*\n\t\t  * TODO [2200-12-12]: Yet\n\t\t  * TODO [engine:node@>=100]: Another\n\t\t  * TODO [+popura]: Way\n\t\t  */"},
			{
				Code: "// TODO"},
			{
				Code: "// TODO [invalid]"},
			{
				Code: "// TODO [] might have [some] that [try [to trick] me]"},
			{
				Code: "// TODO [but [it will]] [fallback] [[[ to the default ]]] rule [["},
			{
				Code:    "// TODO ISSUE-123 fix later",
				Options: []any{map[string]any{"allowWarningComments": false, "ignore": []any{"ISSUE-\\d+"}}}},
			{
				Code:    "// TODO [ISSUE-123] fix later",
				Options: []any{map[string]any{"allowWarningComments": false, "ignore": []any{"ISSUE-\\d+"}}}},
			{
				Code:    "// TODO [1999-01-01, ISSUE-123] fix later",
				Options: []any{map[string]any{"allowWarningComments": false, "ignore": []any{"ISSUE-\\d+"}}}},
			{
				Code:    "// TODO [Issue-123] fix later",
				Options: []any{map[string]any{"allowWarningComments": false, "ignore": []any{"[iI][sS][sS][uU][eE]-\\d+"}}}},
			{
				Code:    "// TODO [2001-01-01]: quite old",
				Options: []any{map[string]any{"date": "2000-01-01"}}},
			{
				Code:    "// TODO [2000-01-01]: too old but ignored in all environments",
				Options: []any{map[string]any{"checkDates": false, "checkDatesOnPullRequests": true}}},
			{
				Code:    "// eslint-disable-next-line test\n\t\t\t\t   // TODO without a date",
				Options: []any{map[string]any{"allowWarningComments": false}}},
			{
				Code:    "/* eslint-disable test */\n\t\t\t\t   // TODO without a date\n\t\t\t\t   // fixme [2000-01-01]: too old'\n\t\t\t\t   /* eslint-enable test */",
				Options: []any{map[string]any{"allowWarningComments": false}}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    "// TODO [2000-01-01]: too old",
				Options: []any{map[string]any{"checkDates": true, "checkDatesOnPullRequests": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. too old", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}}},
			{
				Code:    "/*\n\t\t\t* TODO [2000-01-01]: Yet\n\t\t\t* TODO [2000-01-01]: Another\n\t\t\t* TODO [2000-01-01] Way\n\t\t\t*/",
				Options: []any{map[string]any{"checkDates": true, "checkDatesOnPullRequests": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. Yet", Line: 1, Column: 1, EndLine: 5, EndColumn: 6},
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. Another", Line: 1, Column: 1, EndLine: 5, EndColumn: 6},
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. Way", Line: 1, Column: 1, EndLine: 5, EndColumn: 6}}},
			{
				Code:    "/*\n\t\t\t* TODO [2000-01-01]: Invalid\n\t\t\t* TODO [2200-01-01]: Valid\n\t\t\t* TODO [2000-01-01]: Invalid\n\t\t\t*/",
				Options: []any{map[string]any{"checkDates": true, "checkDatesOnPullRequests": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. Invalid", Line: 1, Column: 1, EndLine: 5, EndColumn: 6},
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. Invalid", Line: 1, Column: 1, EndLine: 5, EndColumn: 6}}},
			{
				Code: "/*\n\t\t\t* Something here\n\t\t\t* TODO [engine:node@>=8]: Invalid\n\t\t\t* Also something here\n\t\t\t*/",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/engineMatches", Message: "Due since Node.js version matched: node>=8. Invalid", Line: 1, Column: 1, EndLine: 5, EndColumn: 6}}},
			{
				Code:    "// fixme [2000-01-01]: too old",
				Options: []any{map[string]any{"checkDates": true, "checkDatesOnPullRequests": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. too old", Line: 1, Column: 1, EndLine: 1, EndColumn: 31}}},
			{
				Code:    "// xxx [2000-01-01]: too old",
				Options: []any{map[string]any{"checkDates": true, "checkDatesOnPullRequests": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. too old", Line: 1, Column: 1, EndLine: 1, EndColumn: 29}}},
			{
				Code:    "// ToDo [2000-01-01]: too old",
				Options: []any{map[string]any{"checkDates": true, "checkDatesOnPullRequests": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. too old", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}}},
			{
				Code:    "// fIxME [2000-01-01]: too old",
				Options: []any{map[string]any{"checkDates": true, "checkDatesOnPullRequests": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. too old", Line: 1, Column: 1, EndLine: 1, EndColumn: 31}}},
			{
				Code:    "// Todoist [2000-01-01]: too old",
				Options: []any{map[string]any{"checkDates": true, "checkDatesOnPullRequests": true, "terms": []any{"Todoist"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. too old", Line: 1, Column: 1, EndLine: 1, EndColumn: 33}}},
			{
				Code:    "// Expire Condition [2000-01-01]: too old",
				Options: []any{map[string]any{"checkDates": true, "checkDatesOnPullRequests": true, "terms": []any{"Expire Condition"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. too old", Line: 1, Column: 1, EndLine: 1, EndColumn: 42}}},
			{
				Code:    "// XxX [2000-01-01]: too old",
				Options: []any{map[string]any{"checkDates": true, "checkDatesOnPullRequests": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. too old", Line: 1, Column: 1, EndLine: 1, EndColumn: 29}}},
			{
				Code: "// TODO [2200-12-12, 2200-12-12]: Multiple dates",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/avoidMultipleDates", Message: "Avoid using multiple expiration dates: 2200-12-12, 2200-12-12. Multiple dates", Line: 1, Column: 1, EndLine: 1, EndColumn: 49}}},
			{
				Code:    "// TODO [2200-12-12, 2200-12-12]: Multiple dates are still invalid",
				Options: []any{map[string]any{"checkDates": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/avoidMultipleDates", Message: "Avoid using multiple expiration dates: 2200-12-12, 2200-12-12. Multiple dates are still invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 67}}},
			{
				Code: "// TODO [>1]: if your package.json version is >1",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/reachedPackageVersion", Message: "Past due package version: >1. if your package.json version is >1", Line: 1, Column: 1, EndLine: 1, EndColumn: 49}}},
			{
				Code: "// TODO [>1, >2]: multiple package versions",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/avoidMultiplePackageVersions", Message: "Avoid using multiple package versions: >1, >2. multiple package versions", Line: 1, Column: 1, EndLine: 1, EndColumn: 44}}},
			{
				Code: "// TODO [>=1]: if your package.json version is >=1",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/reachedPackageVersion", Message: "Past due package version: >=1. if your package.json version is >=1", Line: 1, Column: 1, EndLine: 1, EndColumn: 51}}},
			{
				Code: "// TODO [+find-up-simple]: when you install `find-up-simple`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/havePackage", Message: "Due since find-up-simple was installed. when you install `find-up-simple`", Line: 1, Column: 1, EndLine: 1, EndColumn: 61}}},
			{
				Code: "// TODO [-popura]: when you uninstall `popura`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/dontHavePackage", Message: "Due since popura was removed. when you uninstall `popura`", Line: 1, Column: 1, EndLine: 1, EndColumn: 47}}},
			{
				Code: "// TODO [find-up-simple@>=1]: when `find-up-simple` version is >= 1",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: find-up-simple >= 1. when `find-up-simple` version is >= 1", Line: 1, Column: 1, EndLine: 1, EndColumn: 68}}},
			{
				Code: "// TODO [engine:node@>=8]: when support is for node >= 8",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/engineMatches", Message: "Due since Node.js version matched: node>=8. when support is for node >= 8", Line: 1, Column: 1, EndLine: 1, EndColumn: 57}}},
			{
				Code: "// TODO [peer:eslint@>=9]: when the peer eslint floor reaches >= 9",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/peerVersionMatches", Message: "Due since peer dependency version matched: eslint >= 9. when the peer eslint floor reaches >= 9", Line: 1, Column: 1, EndLine: 1, EndColumn: 67}}},
			{
				Code: "// TODO [peer:eslint@>8]: when the peer eslint floor reaches > 8",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/peerVersionMatches", Message: "Due since peer dependency version matched: eslint > 8. when the peer eslint floor reaches > 8", Line: 1, Column: 1, EndLine: 1, EndColumn: 65}}},
			{
				Code: "// TODO [find-up-simple@>0.2.0]: when `find-up-simple` version is > 0.2.0",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: find-up-simple > 0.2.0. when `find-up-simple` version is > 0.2.0", Line: 1, Column: 1, EndLine: 1, EndColumn: 74}}},
			{
				Code: "// TODO [@lubien/fixture-beta-package@>=1.0.0-alfa.1]: when `@lubien/fixture-beta-package` version is >= 1.0.0-alfa.1",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: @lubien/fixture-beta-package >= 1.0.0-alfa.1. when `@lubien/fixture-beta-package` version is >= 1.0.0-alfa.1", Line: 1, Column: 1, EndLine: 1, EndColumn: 118}}},
			{
				Code: "// TODO [@lubien/fixture-beta-package@>=1.0.0-beta.1]: when `@lubien/fixture-beta-package` version is >= 1.0.0-beta.1",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: @lubien/fixture-beta-package >= 1.0.0-beta.1. when `@lubien/fixture-beta-package` version is >= 1.0.0-beta.1", Line: 1, Column: 1, EndLine: 1, EndColumn: 118}}},
			{
				Code: "// TODO [@lubien/fixture-beta-package@>=1.0.0-beta.0]: when `@lubien/fixture-beta-package` version is >= 1.0.0-beta.0",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: @lubien/fixture-beta-package >= 1.0.0-beta.0. when `@lubien/fixture-beta-package` version is >= 1.0.0-beta.0", Line: 1, Column: 1, EndLine: 1, EndColumn: 118}}},
			{
				Code: "// TODO [@lubien/fixture-beta-package@>0.9]: when `@lubien/fixture-beta-package` prerelease version is > 0.9",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: @lubien/fixture-beta-package > 0.9. when `@lubien/fixture-beta-package` prerelease version is > 0.9", Line: 1, Column: 1, EndLine: 1, EndColumn: 109}}},
			{
				Code: "// TODO [semver>1]: Missing @.",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/missingAtSymbol", Message: "Missing '@' on TODO argument. On 'semver>1' use 'semver@>1'. Missing @.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31}}},
			{
				Code: "// TODO [> 1]: Remove whitespace when it can fix.",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/removeWhitespaces", Message: "Avoid using whitespace on TODO argument. On '> 1' use '>1'. Remove whitespace when it can fix.", Line: 1, Column: 1, EndLine: 1, EndColumn: 50}}},
			{
				Code: "// TODO [semver@> 1]: Remove whitespace when it can fix.",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/removeWhitespaces", Message: "Avoid using whitespace on TODO argument. On 'semver@> 1' use 'semver@>1'. Remove whitespace when it can fix.", Line: 1, Column: 1, EndLine: 1, EndColumn: 57}}},
			{
				Code: "// TODO [semver @>1]: Remove whitespace when it can fix.",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/removeWhitespaces", Message: "Avoid using whitespace on TODO argument. On 'semver @>1' use 'semver@>1'. Remove whitespace when it can fix.", Line: 1, Column: 1, EndLine: 1, EndColumn: 57}}},
			{
				Code: "// TODO [semver@>= 1]: Remove whitespace when it can fix.",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/removeWhitespaces", Message: "Avoid using whitespace on TODO argument. On 'semver@>= 1' use 'semver@>=1'. Remove whitespace when it can fix.", Line: 1, Column: 1, EndLine: 1, EndColumn: 58}}},
			{
				Code: "// TODO [semver @>=1]: Remove whitespace when it can fix.",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/removeWhitespaces", Message: "Avoid using whitespace on TODO argument. On 'semver @>=1' use 'semver@>=1'. Remove whitespace when it can fix.", Line: 1, Column: 1, EndLine: 1, EndColumn: 58}}},
			{
				Code: "// TODO [engine:node @>=1]: Remove whitespace when it can fix.",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/removeWhitespaces", Message: "Avoid using whitespace on TODO argument. On 'engine:node @>=1' use 'engine:node@>=1'. Remove whitespace when it can fix.", Line: 1, Column: 1, EndLine: 1, EndColumn: 63}}},
			{
				Code: "// TODO [engine:node@>= 1]: Remove whitespace when it can fix.",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/removeWhitespaces", Message: "Avoid using whitespace on TODO argument. On 'engine:node@>= 1' use 'engine:node@>=1'. Remove whitespace when it can fix.", Line: 1, Column: 1, EndLine: 1, EndColumn: 63}}},
			{
				Code:    "// TODO",
				Options: []any{map[string]any{"allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}}},
			{
				Code:    "// TODO []",
				Options: []any{map[string]any{"allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO []'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11}}},
			{
				Code:    "// TODO [no meaning at all]",
				Options: []any{map[string]any{"allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [no meaning at all]'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28}}},
			{
				Code:    "// TODO [] might have [some] that [try [to trick] me]",
				Options: []any{map[string]any{"allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [] might have [some] that [try [to...'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 54}}},
			{
				Code:    "// TODO [but [it will]] [fallback] [[[ to the default ]]] rule [[[",
				Options: []any{map[string]any{"allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [but [it will]] [fallback] [[[ to...'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 67}}},
			{
				Code:    "// TODO [engine:npm@>=10000]: Unsupported engine",
				Options: []any{map[string]any{"allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [engine:npm@>=10000]: Unsupported...'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 49}}},
			{
				Code:    "// TODO [engine:somethingrandom@>=10000]: Unsupported engine",
				Options: []any{map[string]any{"allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [engine:somethingrandom@>=10000]:...'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 61}}},
			{
				Code:    "// TODO [2000-01-01, >1]: Combine date with package version",
				Options: []any{map[string]any{"checkDates": true, "checkDatesOnPullRequests": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. Combine date with package version", Line: 1, Column: 1, EndLine: 1, EndColumn: 60},
					{MessageId: "unicorn/reachedPackageVersion", Message: "Past due package version: >1. Combine date with package version", Line: 1, Column: 1, EndLine: 1, EndColumn: 60}}},
			{
				Code: "// TODO [2200-12-12, >1, 2200-12-12, >2]: Multiple dates and package versions",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/avoidMultipleDates", Message: "Avoid using multiple expiration dates: 2200-12-12, 2200-12-12. Multiple dates and package versions", Line: 1, Column: 1, EndLine: 1, EndColumn: 78},
					{MessageId: "unicorn/avoidMultiplePackageVersions", Message: "Avoid using multiple package versions: >1, >2. Multiple dates and package versions", Line: 1, Column: 1, EndLine: 1, EndColumn: 78}}},
			{
				Code: "// TODO [-popura, find-up-simple@>=1]: Combine not having a package with version match",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/dontHavePackage", Message: "Due since popura was removed. Combine not having a package with version match", Line: 1, Column: 1, EndLine: 1, EndColumn: 87},
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: find-up-simple >= 1. Combine not having a package with version match", Line: 1, Column: 1, EndLine: 1, EndColumn: 87}}},
			{
				Code: "// TODO [+find-up-simple, -popura]: Combine presence/absence of packages",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/havePackage", Message: "Due since find-up-simple was installed. Combine presence/absence of packages", Line: 1, Column: 1, EndLine: 1, EndColumn: 73},
					{MessageId: "unicorn/dontHavePackage", Message: "Due since popura was removed. Combine presence/absence of packages", Line: 1, Column: 1, EndLine: 1, EndColumn: 73}}},
			{
				Code:    "// Expire Condition [2000-01-01, semver>1]: Expired TODO and missing symbol",
				Options: []any{map[string]any{"checkDates": true, "checkDatesOnPullRequests": true, "terms": []any{"Expire Condition"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. Expired TODO and missing symbol", Line: 1, Column: 1, EndLine: 1, EndColumn: 76},
					{MessageId: "unicorn/missingAtSymbol", Message: "Missing '@' on TODO argument. On 'semver>1' use 'semver@>1'. Expired TODO and missing symbol", Line: 1, Column: 1, EndLine: 1, EndColumn: 76}}},
			{
				Code: "// TODO [semver @>=1, -popura]: Package uninstalled and whitespace error",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/dontHavePackage", Message: "Due since popura was removed. Package uninstalled and whitespace error", Line: 1, Column: 1, EndLine: 1, EndColumn: 73},
					{MessageId: "unicorn/removeWhitespaces", Message: "Avoid using whitespace on TODO argument. On 'semver @>=1' use 'semver@>=1'. Package uninstalled and whitespace error", Line: 1, Column: 1, EndLine: 1, EndColumn: 73}}},
			{
				Code:    "// HUGETODO [semver @>=1, engine:node@>=8, 2000-01-01, -popura, >1, +find-up-simple, find-up-simple@>=1, peer:eslint@>=9]: Big mix",
				Options: []any{map[string]any{"checkDates": true, "checkDatesOnPullRequests": true, "terms": []any{"HUGETODO"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. Big mix", Line: 1, Column: 1, EndLine: 1, EndColumn: 131},
					{MessageId: "unicorn/reachedPackageVersion", Message: "Past due package version: >1. Big mix", Line: 1, Column: 1, EndLine: 1, EndColumn: 131},
					{MessageId: "unicorn/dontHavePackage", Message: "Due since popura was removed. Big mix", Line: 1, Column: 1, EndLine: 1, EndColumn: 131},
					{MessageId: "unicorn/havePackage", Message: "Due since find-up-simple was installed. Big mix", Line: 1, Column: 1, EndLine: 1, EndColumn: 131},
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: find-up-simple >= 1. Big mix", Line: 1, Column: 1, EndLine: 1, EndColumn: 131},
					{MessageId: "unicorn/peerVersionMatches", Message: "Due since peer dependency version matched: eslint >= 9. Big mix", Line: 1, Column: 1, EndLine: 1, EndColumn: 131},
					{MessageId: "unicorn/engineMatches", Message: "Due since Node.js version matched: node>=8. Big mix", Line: 1, Column: 1, EndLine: 1, EndColumn: 131},
					{MessageId: "unicorn/removeWhitespaces", Message: "Avoid using whitespace on TODO argument. On 'semver @>=1' use 'semver@>=1'. Big mix", Line: 1, Column: 1, EndLine: 1, EndColumn: 131}}},
			{
				Code:    "// TODO [ISSUE-123] fix later",
				Options: []any{map[string]any{"allowWarningComments": false, "ignore": []any{}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [ISSUE-123] fix later'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}}},
			{
				Code:    "\n\t\t\t// TODO fix later\n\t\t\t// TODO ISSUE-123 fix later\n\t\t\t",
				Options: []any{map[string]any{"allowWarningComments": false, "ignore": []any{"[iI][sS][sS][uU][eE]-\\d+"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO fix later'.", Line: 2, Column: 4, EndLine: 2, EndColumn: 21}}},
			{
				Code:    "/*\n\t\t\tTODO Invalid\n\t\t\tTODO ISSUE-123 Valid\n\t\t\t*/",
				Options: []any{map[string]any{"allowWarningComments": false, "ignore": []any{"[iI][sS][sS][uU][eE]-\\d+"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO Invalid'.", Line: 1, Column: 1, EndLine: 4, EndColumn: 6}}},
			{
				Code:    "// TODO [2999-12-01]: Y3K bug",
				Options: []any{map[string]any{"date": "3000-01-01", "checkDates": true, "checkDatesOnPullRequests": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2999-12-01. Y3K bug", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}}},
		})
}

func TestExpiringTodoCommentsDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(upstreamRoot(t), "tsconfig.json", t, &ExpiringTodoCommentsRule,
		[]rule_tester.ValidTestCase{
			{
				Code:     "// ✅\n// TODO [2200-12-25]: Too long... Can you feel it?\n// FIXME [2200-12-25]: Too long... Can you feel it?\n\n// TODO (lubien) [2200-12-12]: You can add something before the arguments.\n// TODO @lubien [2200-12-12]: You can add something before the arguments.\n// FIXME [2200-12-25] (lubien): You can add something after the arguments, before the colon.\n// TODO [2200-12-12] No colon after argument.\n\n// TODO [+react]: Refactor this when we use React.\n// TODO [-lodash]: If we remove lodash we need to change this.\n\n// TODO [lodash@>10]: Lodash has a new way to do this; when we bump to its version let's use it.\n// TODO [lodash@>=10]: Lodash has a new way to do this; when we bump to its version let's use it.\n\n// TODO [2200-12-25, +popura, lodash@>10]: Combo.\n\n// TODO [peer:eslint@>=99]: When our minimum supported `eslint` reaches v99.\n\n// TODO [engine:node@>12]: When we bump to this Node version we can use import/export.\n\n/*\n * TODO [2200-12-25]: Yet\n * TODO [2200-12-25]: Another\n * TODO [2200-12-25]: Way\n */\n",
				FileName: "docs/input.ts",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     "// TODO [2019-11-15]: Refactor this code before the sprint ends.\n// TODO (@lubien) [2019-07-18]: When John delivers his code. I can reuse it.\n// TODO [2019-08-10]: I must refactor this for sure before I deliver.\n",
				FileName: "docs/input.ts",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2019-11-15. Refactor this code before the sprint ends.", Line: 1, Column: 1, EndLine: 1, EndColumn: 65},
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2019-07-18. When John delivers his code. I can reuse it.", Line: 2, Column: 1, EndLine: 2, EndColumn: 77},
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2019-08-10. I must refactor this for sure before I deliver.", Line: 3, Column: 1, EndLine: 3, EndColumn: 70}}},
			{
				Code:     "// TODO [>=1.0.0]: I should work around this when we reach v1.\n// TODO (@lubien) [>0]: For now this is fine but for a stable version we must refactor.\n// FIXME [>10]: This feature is deprecated and should be removed from the next major version.\n",
				FileName: "docs/input.ts",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/reachedPackageVersion", Message: "Past due package version: >=1.0.0. I should work around this when we reach v1.", Line: 1, Column: 1, EndLine: 1, EndColumn: 63},
					{MessageId: "unicorn/reachedPackageVersion", Message: "Past due package version: >0. For now this is fine but for a stable version we must refactor.", Line: 2, Column: 1, EndLine: 2, EndColumn: 88}}},
			{
				Code:     "// TODO [engine:node@>=8]: We can use async/await now.\n// FIXME [engine:node@>=20.0.0]: Hey, node can use import/export now, we should refactor.\n",
				FileName: "docs/input.ts",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/engineMatches", Message: "Due since Node.js version matched: node>=8. We can use async/await now.", Line: 1, Column: 1, EndLine: 1, EndColumn: 55}}},
			{
				Code:     "// TODO [-vue-function-api]: When we remove `vue-function-api` we should refactor this.\n// FIXME [+read-pkg]: If we use this package we don't need to use this function below.\n// XXX @lubien [+react, -jquery]: We can use React for this widget instead of jQuery.\n",
				FileName: "docs/input.ts",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/dontHavePackage", Message: "Due since vue-function-api was removed. When we remove `vue-function-api` we should refactor this.", Line: 1, Column: 1, EndLine: 1, EndColumn: 88},
					{MessageId: "unicorn/havePackage", Message: "Due since read-pkg was installed. If we use this package we don't need to use this function below.", Line: 2, Column: 1, EndLine: 2, EndColumn: 87},
					{MessageId: "unicorn/dontHavePackage", Message: "Due since jquery was removed. We can use React for this widget instead of jQuery.", Line: 3, Column: 1, EndLine: 3, EndColumn: 86}}},
			{
				Code:     "// TODO [vue@>=3]: Refactor to function API when it's stable.\n// FIXME [cerebro@>0.10.0]: This is a quickfix until cerebro fixes this.\n// XXX [popura@>=2.0.0]: This API is deprecated so we should not use it by then.\n",
				FileName: "docs/input.ts",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: vue >= 3. Refactor to function API when it's stable.", Line: 1, Column: 1, EndLine: 1, EndColumn: 62},
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: cerebro > 0.10.0. This is a quickfix until cerebro fixes this.", Line: 2, Column: 1, EndLine: 2, EndColumn: 73}}},
			{
				Code:     "// TODO [peer:eslint@>=9]: Drop the `CLIEngine` fallback once we require ESLint 9.\n",
				FileName: "docs/input.ts",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/peerVersionMatches", Message: "Due since peer dependency version matched: eslint >= 9. Drop the `CLIEngine` fallback once we require ESLint 9.", Line: 1, Column: 1, EndLine: 1, EndColumn: 83}}},
			{
				Code:     "// TODO [+react, -jquery]: We can use React for this widget instead of jQuery.\n// TODO [2019-07-15, +react]: Refactor this if we install React or if we reach that date.\n// TODO [-vue-function-api, vue@>=3]: Now we should use Vue native function API.\n",
				FileName: "docs/input.ts",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/dontHavePackage", Message: "Due since jquery was removed. We can use React for this widget instead of jQuery.", Line: 1, Column: 1, EndLine: 1, EndColumn: 79},
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2019-07-15. Refactor this if we install React or if we reach that date.", Line: 2, Column: 1, EndLine: 2, EndColumn: 90},
					{MessageId: "unicorn/dontHavePackage", Message: "Due since vue-function-api was removed. Now we should use Vue native function API.", Line: 3, Column: 1, EndLine: 3, EndColumn: 81},
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: vue >= 3. Now we should use Vue native function API.", Line: 3, Column: 1, EndLine: 3, EndColumn: 81}}},
			{
				Code:     "/*\n * We should really make this code better.\n * When we support Node.js 12 we can refactor imports.\n * And we also can do [x], [y], [z].\n * TODO [engine:node@>=12]: Use import/export.\n */\n\n/*\n * This code would be so easy if we used `popura` package helpers.\n * When you can, install `popura`, use it and remove dead code.\n * TODO [+popura]: Refactor to use `popura`.\n *\n * You can also use `popura-cli` since we want help on [feature].\n * TODO [+popura-cli]: Document how to use `popura-cli`.\n */\n",
				FileName: "docs/input.ts",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/engineMatches", Message: "Due since Node.js version matched: node>=12. Use import/export.", Line: 1, Column: 1, EndLine: 6, EndColumn: 4}}},
			{
				Code:     "// ❌\n// With `checkDates: true`\n// TODO [2000-01-01]: I'll fix this next week.\n// TODO [2000-01-01, 2001-01-01]: Multiple dates won't work.\n\n// TODO [>1]: If your package.json version is > 1.\n// TODO [>=1]: If your package.json version is >= 1.\n// TODO [>1, >2]: Multiple package versions won't work.\n\n// TODO [+already-have-pkg]: Since we already have it, this reports.\n// TODO [-we-dont-have-this-package]: Since we don't have, trigger a report.\n\n// TODO [read-pkg@>1]: When `read-pkg` version is > 1 don't forget to do this.\n// TODO [read-pkg@>=5.1.1]: When `read-pkg` version is >= 5.1.1 don't forget to do that.\n\n// TODO [peer:eslint@>=8]: Whoops, our minimum supported `eslint` is already >= 8.\n\n// TODO [engine:node@>=8]: Whoops, we are already supporting it!\n\n// TODO: Add unicorns.\n",
				FileName: "docs/input.ts",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. I'll fix this next week.", Line: 3, Column: 1, EndLine: 3, EndColumn: 47},
					{MessageId: "unicorn/avoidMultipleDates", Message: "Avoid using multiple expiration dates: 2000-01-01, 2001-01-01. Multiple dates won't work.", Line: 4, Column: 1, EndLine: 4, EndColumn: 61},
					{MessageId: "unicorn/reachedPackageVersion", Message: "Past due package version: >1. If your package.json version is > 1.", Line: 6, Column: 1, EndLine: 6, EndColumn: 51},
					{MessageId: "unicorn/reachedPackageVersion", Message: "Past due package version: >=1. If your package.json version is >= 1.", Line: 7, Column: 1, EndLine: 7, EndColumn: 53},
					{MessageId: "unicorn/avoidMultiplePackageVersions", Message: "Avoid using multiple package versions: >1, >2. Multiple package versions won't work.", Line: 8, Column: 1, EndLine: 8, EndColumn: 56},
					{MessageId: "unicorn/havePackage", Message: "Due since already-have-pkg was installed. Since we already have it, this reports.", Line: 10, Column: 1, EndLine: 10, EndColumn: 69},
					{MessageId: "unicorn/dontHavePackage", Message: "Due since we-dont-have-this-package was removed. Since we don't have, trigger a report.", Line: 11, Column: 1, EndLine: 11, EndColumn: 77},
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: read-pkg > 1. When `read-pkg` version is > 1 don't forget to do this.", Line: 13, Column: 1, EndLine: 13, EndColumn: 79},
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: read-pkg >= 5.1.1. When `read-pkg` version is >= 5.1.1 don't forget to do that.", Line: 14, Column: 1, EndLine: 14, EndColumn: 89},
					{MessageId: "unicorn/peerVersionMatches", Message: "Due since peer dependency version matched: eslint >= 8. Whoops, our minimum supported `eslint` is already >= 8.", Line: 16, Column: 1, EndLine: 16, EndColumn: 83},
					{MessageId: "unicorn/engineMatches", Message: "Due since Node.js version matched: node>=8. Whoops, we are already supporting it!", Line: 18, Column: 1, EndLine: 18, EndColumn: 65},
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO: Add unicorns.'.", Line: 20, Column: 1, EndLine: 20, EndColumn: 23}}},
		})
}

func TestExpiringTodoCommentsCatalog(t *testing.T) {
	rule_tester.RunRuleTester(upstreamRoot(t), "tsconfig.json", t, &ExpiringTodoCommentsRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code:     "\n\t\t// TODO [eslint@>9]: Drop fallback.\n\t\t// TODO [prettier@>3]: Drop dev fallback.\n\t\t// TODO [peer:eslint@>9]: Drop peer fallback.\n\t\t// TODO [+eslint]: Presence checks still work.\n\t",
				FileName: "catalog/input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/unsupportedCatalogProtocol", Message: "Cannot check dependency version because eslint uses the unsupported `catalog:` protocol. Drop fallback.", Line: 2, Column: 3, EndLine: 2, EndColumn: 38},
					{MessageId: "unicorn/unsupportedCatalogProtocol", Message: "Cannot check dependency version because prettier uses the unsupported `catalog:` protocol. Drop dev fallback.", Line: 3, Column: 3, EndLine: 3, EndColumn: 44},
					{MessageId: "unicorn/unsupportedCatalogProtocol", Message: "Cannot check peer dependency version because eslint uses the unsupported `catalog:` protocol. Drop peer fallback.", Line: 4, Column: 3, EndLine: 4, EndColumn: 48},
					{MessageId: "unicorn/havePackage", Message: "Due since eslint was installed. Presence checks still work.", Line: 5, Column: 3, EndLine: 5, EndColumn: 49}}},
		})
}

func TestExpiringTodoCommentsOtherLanguages(t *testing.T) {
	for _, tc := range []struct{ name, code string }{
		{"supports JSONC comments with @eslint/json", "\n\t\t// TODO [2000-01-01]: Drop\n\t\t// TODO: Update config\n\t\t{\"key\": \"value\"}\n\t"},
		{"ignores JSONC eslint directive comments with @eslint/json", "\n\t\t// eslint-disable-next-line test -- TODO reason\n\t\t// TODO: Update config\n\t\t{\"key\": \"value\"}\n\t"},
		{"supports JSONC block comments with @eslint/json", "\n\t\t/* TODO [2000-01-01]: Drop */\n\t\t{\"key\": \"value\"}\n\t"},
		{"supports HTML comments with @html-eslint", "<!-- TODO [2000-01-01]: Drop -->\n<!-- TODO: Update markup -->\n<div></div>"},
		{"supports Markdown HTML comments with @eslint/markdown", "<!-- TODO [2000-01-01]: Drop -->\n\n<!-- TODO: Update docs -->\n\n# Hello"},
		{"ignores HTML comments inside Markdown fenced code blocks", "```html\n<!-- TODO [2000-01-01]: Inside fence -->\n```\n\n<!-- TODO [2000-01-01]: Outside fence -->"},
		{"ignores HTML comments inside Markdown tilde fenced code blocks", "~~~\n<!-- TODO [2000-01-01]: Inside fence -->\n~~~\n\n<!-- TODO [2000-01-01]: Outside fence -->"},
		{"ignores HTML comments inside Markdown code", "`<!-- TODO [2000-01-01]: Inline code -->`\n\n    <!-- TODO [2000-01-01]: Indented code -->\n\n<!-- TODO [2000-01-01]: Outside code -->"},
		{"reports every TODO line inside a multi-line Markdown HTML comment", "<!--\nTODO [2000-01-01]: First\nTODO [2000-01-01]: Second\n-->"},
		{"handles an unterminated Markdown HTML comment without truncating its text", "<!-- TODO [2000-01-01]: Unterminated"},
		{"supports CSS comments with @eslint/css", "\n\t\t/* TODO [2000-01-01]: Drop */\n\t\t/* TODO: Add styles */\n\t\t.outdated { color: hotpink; }\n\t"},
		{"supports YAML comments with eslint-plugin-yml", "key: value # TODO [2000-01-01]: Drop\n# TODO: Update config"},
		{"supports ESLint disable directives in YAML", "# eslint-disable-next-line test -- TODO reason\n# TODO: Update config\nkey: value"},
		{"supports TOML comments with eslint-plugin-toml", "key = \"value\" # TODO [2000-01-01]: Drop\n# TODO: Update config\n# TODO [2999-01-01]: Later\ntext = \"# TODO [2000-01-01]: String\""},
		{"supports ESLint disable directives in TOML", "# eslint-disable-next-line test -- TODO reason\n# TODO: Update config\nkey = \"value\""},
		{"documentation css", "/* TODO [2019-11-15]: Remove this fallback. */\n.outdated {\n\tcolor: hotpink;\n}\n"},
		{"documentation html", "<!-- TODO [2019-11-15]: Remove this fallback. -->\n<div class=\"outdated\"></div>\n"},
		{"documentation jsonc", "// TODO [2019-11-15]: Update this configuration.\n{\n\t\"key\": \"value\"\n}\n"},
		{"documentation md", "<!-- TODO [2019-11-15]: Update this section. -->\n\n# Heading\n"},
		{"documentation yaml", "# TODO [2019-11-15]: Update this configuration.\nkey: value\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Skip("rslint native rules currently lint JS/TS only; requires the upstream language plugin. Input: " + tc.code)
		})
	}
}
