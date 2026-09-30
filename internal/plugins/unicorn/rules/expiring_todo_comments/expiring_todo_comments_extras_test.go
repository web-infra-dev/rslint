// cspell:ignore nottodo
package expiring_todo_comments

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Expected diagnostics were compared with eslint-plugin-unicorn v76.0.0.
func TestExpiringTodoCommentsExtras(t *testing.T) {
	rule_tester.RunRuleTester(upstreamRoot(t), "tsconfig.json", t, &ExpiringTodoCommentsRule,
		[]rule_tester.ValidTestCase{
			{Code: "// TODO [spaces@>=1.2.3]: padded prerelease", FileName: "edge/input.tsx", Tsx: true, Options: map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
			{
				Code:     "// TODO [2026-09-30]: same day",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [2026-09-31]: normalized into October",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [2026-13-01]: invalid month",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [2026-00-01]: invalid month",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [2026-09-00]: invalid day",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [2026-09-32]: invalid day",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [+ aa]: name may contain spaces after sign",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [>9007199254740992]: invalid large range",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [>01]: invalid leading zero",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [>1.2.3-beta.01]: invalid numeric prerelease",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "const text = \"// TODO [2000-01-01]: string\"; const re = /TODO/; const template = `// TODO [2000-01-01]`;",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// eslint-disable-next-line no-unused-vars -- TODO [2000-01-01]: directive\nconst value = 1;",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "/* eslint no-warning-comments: off -- TODO */",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO bare",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [2000-01-01]: no terms",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "terms": []any{}}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [2000-01-01]: default explicitly off",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": false, "checkDatesOnPullRequests": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [2000-01-01]: ISSUE-123",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "ignore": []any{"(?<=ISSUE-)\\d+"}}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [2000-01-01]: 中文",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "ignore": []any{"\\p{Lo}+"}}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [2000-01-01]: objects",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "ignore": []any{map[string]any{}}}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// FIXME bare",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "terms": []any{"todo"}, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [>=1.2.3]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [aa@>2]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [bb@>=0]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [cc@>=0]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [dd@>=0]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [ee@>=0]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [+ii]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [peer:union@>=9]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [peer:pre@>=1.0.0]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [peer:empty@>=0]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [peer:upper@>=0.0.0-0]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [peer:workspace@>=0]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true},
			{
				Code:     "// TODO [engine:node@>=1]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "// eslint-disable no-unused-vars -- TODO [2000-01-01]: line", Options: map[string]any{"checkDates": true, "checkDatesOnPullRequests": true, "date": "2026-09-30"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. line", Line: 1, Column: 1, EndLine: 1, EndColumn: 60}}},
			{Code: "// eslint-enable no-unused-vars -- TODO [2000-01-01]: line", Options: map[string]any{"checkDates": true, "checkDatesOnPullRequests": true, "date": "2026-09-30"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. line", Line: 1, Column: 1, EndLine: 1, EndColumn: 59}}},
			{Code: "// TODO [+😀, 😀@>=1, -🦄]: unicode", FileName: "edge/input.tsx", Tsx: true, Options: map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unicorn/havePackage", Message: "Due since 😀 was installed. unicode", Line: 1, Column: 1, EndLine: 1, EndColumn: 36}, {MessageId: "unicorn/versionMatches", Message: "Due since package version matched: 😀 >= 1. unicode", Line: 1, Column: 1, EndLine: 1, EndColumn: 36}, {MessageId: "unicorn/dontHavePackage", Message: "Due since 🦄 was removed. unicode", Line: 1, Column: 1, EndLine: 1, EndColumn: 36}}},
			{Code: "// TODO [+a, +中, a@>=1, 中@>=1]: short", FileName: "edge/input.tsx", Tsx: true, Options: map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [+a, +中, a@>=1, 中@>=1]: short'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 38}}},
			{Code: "// TODO [spaces@>=1.2.3-beta.1]: padded prerelease", FileName: "edge/input.tsx", Tsx: true, Options: map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: spaces >= 1.2.3-beta.1. padded prerelease", Line: 1, Column: 1, EndLine: 1, EndColumn: 51}}},
			{Code: "// TODO [array@>=1]: array", FileName: "edge/input.tsx", Tsx: true, Options: map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: array >= 1. array", Line: 1, Column: 1, EndLine: 1, EndColumn: 27}}},
			{
				Code:     "// TODO [2026-09-29]: yesterday",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2026-09-29. yesterday", Line: 1, Column: 1, EndLine: 1, EndColumn: 32}}},
			{
				Code:     "// TODO [2024-02-30]: normalized March 1",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2024-02-30. normalized March 1", Line: 1, Column: 1, EndLine: 1, EndColumn: 41}}},
			{
				Code:     "// TODO [0000-01-01]: first year",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 0000-01-01. first year", Line: 1, Column: 1, EndLine: 1, EndColumn: 33}}},
			{
				Code:     "// TODO [2000-1-01]: non ISO",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [2000-1-01]: non ISO'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29}}},
			{
				Code:     "// TODO [>1.0.0-ALPHA]: uppercase package prerelease is unknown",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [>1.0.0-ALPHA]: uppercase package...'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 64}}},
			{
				Code:     "// TODO [semver@>=1.0.0-ALPHA]: dependency comparison is insensitive",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: semver >= 1.0.0-ALPHA. dependency comparison is insensitive", Line: 1, Column: 1, EndLine: 1, EndColumn: 69}}},
			{
				Code:     "// TODO [semver@>=1.0.0] [extra]: greedy arguments",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: semver >= 1.0.0. [extra]: greedy arguments", Line: 1, Column: 1, EndLine: 1, EndColumn: 51}}},
			{
				Code:     "// TODO [2000-01-01} nope]: stops at brace",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [2000-01-01} nope]: stops at brace'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 43}}},
			{
				Code:     "// TODO [engine:npm@>=1]: unsupported engine",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [engine:npm@>=1]: unsupported...'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 45}}},
			{
				Code:     "// TODO [@x/pkg>1]: existing at sign prevents repair",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [@x/pkg>1]: existing at sign...'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 53}}},
			{
				Code:     "// TODO [semver\t@>=1]: only spaces are removed",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [semver @>=1]: only spaces are...'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 47}}},
			{
				Code:     "// TODO [2 0 0 0-01-01]: repair date",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/removeWhitespaces", Message: "Avoid using whitespace on TODO argument. On '2 0 0 0-01-01' use '2000-01-01'. repair date", Line: 1, Column: 1, EndLine: 1, EndColumn: 37}}},
			{
				Code:     "// TODO [peer:eslint @>=9]: repair peer",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/removeWhitespaces", Message: "Avoid using whitespace on TODO argument. On 'peer:eslint @>=9' use 'peer:eslint@>=9'. repair peer", Line: 1, Column: 1, EndLine: 1, EndColumn: 40}}},
			{
				Code:     "// TODO [+a]: one-character names are unrecognized",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [+a]: one-character names are...'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 51}}},
			{
				Code:     "// TODO [engine:node>1]: missing engine at",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/missingAtSymbol", Message: "Missing '@' on TODO argument. On 'engine:node>1' use 'engine:node@>1'. missing engine at", Line: 1, Column: 1, EndLine: 1, EndColumn: 43}}},
			{
				Code:     "// TODO [>1.2.3+build]: build metadata",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/reachedPackageVersion", Message: "Past due package version: >1.2.3+build. build metadata", Line: 1, Column: 1, EndLine: 1, EndColumn: 39}}},
			{
				Code:     "// TODO [2000-01-01, unknown]: known condition suppresses fallback",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. known condition suppresses fallback", Line: 1, Column: 1, EndLine: 1, EndColumn: 67}}},
			{
				Code:     "/* TODO bare\nTODO [2000-01-01]: expired\nTODO bare again */",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. expired", Line: 1, Column: 1, EndLine: 3, EndColumn: 19},
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO bare'.", Line: 1, Column: 1, EndLine: 3, EndColumn: 19},
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO bare again'.", Line: 1, Column: 1, EndLine: 3, EndColumn: 19}}},
			{
				Code:     "/*\r\nTODO [2000-01-01]: CRLF\r\nTODO another\r\n*/",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. CRLF", Line: 1, Column: 1, EndLine: 4, EndColumn: 3},
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO another'.", Line: 1, Column: 1, EndLine: 4, EndColumn: 3}}},
			{
				Code:     "/* * TODO without argument */",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': '* TODO without argument'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}}},
			{
				Code:     "// nottodo [2000-01-01]: expiration terms are substrings",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. expiration terms are substrings", Line: 1, Column: 1, EndLine: 1, EndColumn: 57}}},
			{
				Code:     "const value = \"🦄\"; // TODO [2000-01-01]: 中文 🦄",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. 中文 🦄", Line: 1, Column: 21, EndLine: 1, EndColumn: 48}}},
			{
				Code:     "const view = <div title=\"TODO [2000-01-01]\">TODO [2000-01-01]{/* TODO [2000-01-01]: JSX */}</div>;",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. JSX", Line: 1, Column: 63, EndLine: 1, EndColumn: 91}}},
			{
				Code:     "#!/usr/bin/env node TODO [2000-01-01]\n// TODO [2000-01-01]: real comment",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. real comment", Line: 2, Column: 1, EndLine: 2, EndColumn: 35}}},
			{
				Code:     "// TODO bare",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO bare'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13}}},
			{
				Code:     "// TODO [2000-01-01]: duplicate terms",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "terms": []any{"todo", "TODO"}}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. duplicate terms", Line: 1, Column: 1, EndLine: 1, EndColumn: 38}}},
			{
				Code:     "// TODO bare",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false, "terms": []any{"todo", "TODO"}}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO bare'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedComment", Message: "Unexpected 'TODO': 'TODO bare'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13}}},
			{
				Code:     "// note: TODO bare",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "allowWarningComments": false}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'note: TODO bare'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19}}},
			{
				Code:     "// İ [2000-01-01]: Unicode lower expansion",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "terms": []any{"i̇"}}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. Unicode lower expansion", Line: 1, Column: 1, EndLine: 1, EndColumn: 43}}},
			{
				Code:     "// TODO [2000-01-01]: BOM trim\uFEFF",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. BOM trim", Line: 1, Column: 1, EndLine: 1, EndColumn: 32}}},
			{
				Code:     "// TODO [>=1.2.3-beta.1]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/reachedPackageVersion", Message: "Past due package version: >=1.2.3-beta.1. edge", Line: 1, Column: 1, EndLine: 1, EndColumn: 31}}},
			{
				Code:     "// TODO [aa@>=2]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: aa >= 2. edge", Line: 1, Column: 1, EndLine: 1, EndColumn: 23}}},
			{
				Code:     "// TODO [ff@>=2]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: ff >= 2. edge", Line: 1, Column: 1, EndLine: 1, EndColumn: 23}}},
			{
				Code:     "// TODO [gg@>=2]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: gg >= 2. edge", Line: 1, Column: 1, EndLine: 1, EndColumn: 23}}},
			{
				Code:     "// TODO [hh@>=1]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: hh >= 1. edge", Line: 1, Column: 1, EndLine: 1, EndColumn: 23}}},
			{
				Code:     "// TODO [-ii]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/dontHavePackage", Message: "Due since ii was removed. edge", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			{
				Code:     "// TODO [jj@>=1]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/unsupportedCatalogProtocol", Message: "Cannot check dependency version because jj uses the unsupported `catalog:` protocol. edge", Line: 1, Column: 1, EndLine: 1, EndColumn: 23}}},
			{
				Code:     "// TODO [kk@>=0.3]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/versionMatches", Message: "Due since package version matched: kk >= 0.3. edge", Line: 1, Column: 1, EndLine: 1, EndColumn: 25}}},
			{
				Code:     "// TODO [-ll]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/dontHavePackage", Message: "Due since ll was removed. edge", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			{
				Code:     "// TODO [peer:union@>=8]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/peerVersionMatches", Message: "Due since peer dependency version matched: union >= 8. edge", Line: 1, Column: 1, EndLine: 1, EndColumn: 31}}},
			{
				Code:     "// TODO [peer:strict@>=1.2.4]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/peerVersionMatches", Message: "Due since peer dependency version matched: strict >= 1.2.4. edge", Line: 1, Column: 1, EndLine: 1, EndColumn: 36}}},
			{
				Code:     "// TODO [peer:pre@>=1.0.0-beta]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/peerVersionMatches", Message: "Due since peer dependency version matched: pre >= 1.0.0-beta. edge", Line: 1, Column: 1, EndLine: 1, EndColumn: 38}}},
			{
				Code:     "// TODO [peer:wild@>=0]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/peerVersionMatches", Message: "Due since peer dependency version matched: wild >= 0. edge", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}}},
			{
				Code:     "// TODO [peer:catalog@>=0]: edge",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "edge/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/unsupportedCatalogProtocol", Message: "Cannot check peer dependency version because catalog uses the unsupported `catalog:` protocol. edge", Line: 1, Column: 1, EndLine: 1, EndColumn: 33}}},
			{
				Code:     "// TODO [-missing, >1, engine:node@>=1, peer:eslint@>=1]: empty package",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "no-fields/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/dontHavePackage", Message: "Due since missing was removed. empty package", Line: 1, Column: 1, EndLine: 1, EndColumn: 72}}},
			{
				Code:     "// TODO [-missing, >1]: malformed source package",
				Options:  []any{map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true}},
				FileName: "missing/input.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unicorn/dontHavePackage", Message: "Due since missing was removed. malformed source package", Line: 1, Column: 1, EndLine: 1, EndColumn: 49}}},
		})
}

func TestExpiringTodoCommentsWithoutCwdPackage(t *testing.T) {
	base := fixtures.GetRootDir()
	root := rule_tester.Root{Dir: base.Dir, FS: utils.NewOverlayVFS(base.FS, map[string]string{
		tspath.ResolvePath(base.Dir, "nested/package.json"): `{"version":"10.0.0"}`,
	})}
	// Match upstream's cwd-based availability check even when the file has a package.
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &ExpiringTodoCommentsRule,
		[]rule_tester.ValidTestCase{{
			Code:     "// TODO [>1, -missing]: cwd has no package",
			FileName: "nested/file.ts"}},
		[]rule_tester.InvalidTestCase{{
			Code:     "// TODO [>1]: no package",
			FileName: "nested/file.ts",
			Options:  map[string]any{"allowWarningComments": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedComment", Message: "Unexpected 'todo': 'TODO [>1]: no package'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25}}}},
	)
}

func TestExpiringTodoCommentsProcessDirectory(t *testing.T) {
	base := fixtures.GetRootDir()
	root := rule_tester.Root{Dir: base.Dir, FS: utils.NewOverlayVFS(base.FS, map[string]string{
		tspath.ResolvePath(base.Dir, "nested/package.json"): `{"version":"10.0.0"}`,
		tspath.ResolvePath(base.Dir, "other/package.json"):  `{"version":"0.0.0"}`,
	})}
	const code = "// TODO [>1]: source package"
	helper := rule_tester.NewProgramHelper(root)
	raw, source, err := helper.CreateTestProgram(code, "nested/input.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		cwd  string
		want int
	}{
		{base.Dir, 0},
		{tspath.ResolvePath(base.Dir, "other"), 1},
	} {
		t.Run(tc.cwd, func(t *testing.T) {
			ctx := (rule.RuleContext{SourceFile: source, Comments: rule.NewCommentStore(source)}).
				WithProgram(program.NewFromCompiler(raw)).
				WithFileCache(rule.NewFileCacheWithProcessCurrentDirectory(tc.cwd))
			var diagnostics []rule.RuleDiagnostic
			ctx = ctx.WithReporter(ExpiringTodoCommentsRule.Name, rule.SeverityError, func(d rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, d)
			})
			ExpiringTodoCommentsRule.Run(ctx, nil)
			if len(diagnostics) != tc.want {
				t.Fatalf("got %d diagnostics, want %d", len(diagnostics), tc.want)
			}
			if tc.want > 0 {
				d := diagnostics[0]
				if d.Message.Id != "unicorn/reachedPackageVersion" || d.Message.Description != "Past due package version: >1. source package" || d.Range.Pos() != 0 || d.Range.End() != len(code) {
					t.Fatalf("unexpected diagnostic: %#v", d)
				}
			}
		})
	}
}

func TestExpiringTodoCommentsIgnoreTimeout(t *testing.T) {
	// A pathological ignore expression must not become a false-positive report
	// when matching times out. Upstream has no matching deadline.
	rule_tester.RunRuleTester(upstreamRoot(t), "tsconfig.json", t, &ExpiringTodoCommentsRule,
		[]rule_tester.ValidTestCase{{
			Code:    "// TODO [2000-01-01]: " + strings.Repeat("a", 64) + "!",
			Options: map[string]any{"checkDates": true, "checkDatesOnPullRequests": true, "date": "2026-09-30", "ignore": []any{`(?:a|aa)+$`}},
		}}, nil)
}

func TestExpiringTodoCommentsOptions(t *testing.T) {
	for _, options := range [][]any{
		nil, {map[string]any{}}, {map[string]any{"terms": []any{}, "ignore": []any{}, "checkDates": false, "checkDatesOnPullRequests": false, "allowWarningComments": true}},
	} {
		if err := ExpiringTodoCommentsRule.Schema.Validate(options); err != nil {
			t.Fatal(err)
		}
	}
	for _, options := range [][]any{
		{map[string]any{"unknown": true}}, {map[string]any{"terms": "todo"}}, {map[string]any{"ignore": []any{1}}}, {map[string]any{"checkDates": "true"}}, {map[string]any{}, map[string]any{}},
	} {
		if err := ExpiringTodoCommentsRule.Schema.Validate(options); err == nil {
			t.Fatalf("accepted invalid options: %#v", options)
		}
	}
}

func TestExpiringTodoCommentsUnsupportedIgnorePattern(t *testing.T) {
	// esregexp currently lacks Script property escapes. Unsupported patterns are
	// ignored; unlike upstream, this comment still reports. Use a literal range.
	rule_tester.RunRuleTester(upstreamRoot(t), "tsconfig.json", t, &ExpiringTodoCommentsRule, nil,
		[]rule_tester.InvalidTestCase{{
			Code:    "// TODO [2000-01-01]: 中",
			Options: map[string]any{"date": "2026-09-30", "checkDates": true, "checkDatesOnPullRequests": true, "ignore": []any{`\p{Script=Han}+`}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unicorn/expiredTodo", Message: "Past due date: 2000-01-01. 中", Line: 1, Column: 1, EndLine: 1, EndColumn: 24}}}},
	)
}

func TestExpiringTodoCommentsDateOptions(t *testing.T) {
	// Date checks depend on the rule options even in a pull-request job.
	t.Setenv("CI", "true")
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	root := fixtures.GetRootDir()
	helper := rule_tester.NewProgramHelper(root)
	const code = "// TODO [2000-01-01]: remove fallback"
	raw, source, err := helper.CreateTestProgram(code, "input.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	ctx := (rule.RuleContext{SourceFile: source, Comments: rule.NewCommentStore(source)}).
		WithProgram(program.NewFromCompiler(raw)).
		WithFileCache(rule.NewFileCacheWithProcessCurrentDirectory(root.Dir))
	for _, tc := range []struct {
		name                string
		checkDates          bool
		checkOnPullRequests any
		want                int
	}{
		{"disabled", false, nil, 0},
		{"enabled", true, nil, 1},
		{"compatibility false", true, false, 1},
		{"compatibility true", true, true, 1},
		{"compatibility cannot enable dates", false, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options := map[string]any{"date": "2026-09-30", "checkDates": tc.checkDates}
			if tc.checkOnPullRequests != nil {
				options["checkDatesOnPullRequests"] = tc.checkOnPullRequests
			}
			var diagnostics []rule.RuleDiagnostic
			ruleCtx := ctx.WithReporter(ExpiringTodoCommentsRule.Name, rule.SeverityError, func(d rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, d)
			})
			ExpiringTodoCommentsRule.Run(ruleCtx, rule_tester.ResolveTestCaseOptions(t, &ExpiringTodoCommentsRule, options))
			if len(diagnostics) != tc.want {
				t.Fatalf("got %d diagnostics, want %d", len(diagnostics), tc.want)
			}
			if tc.want > 0 {
				d := diagnostics[0]
				if d.Message.Id != "unicorn/expiredTodo" || d.Message.Description != "Past due date: 2000-01-01. remove fallback" || d.Range.Pos() != 0 || d.Range.End() != len(code) {
					t.Fatalf("unexpected diagnostic: %#v", d)
				}
			}
		})
	}
}
