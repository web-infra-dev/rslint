// Ported from eslint-plugin-unicorn v76.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/test/no-console-spaces.js
// Includes its snapshot cases and docs/rules/no-console-spaces.md examples.
package no_console_spaces_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_console_spaces"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func spaceError(method, position string, line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "no-console-spaces",
		Message:   "Do not use " + position + " space between `console." + method + "` parameters.",
		Line:      line, Column: column, EndLine: line, EndColumn: column + 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
	}
}

func TestNoConsoleSpacesUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "console.log(\"abc\");", FileName: "case.js"},
		{Code: "console.log(\"abc\", \"def\");", FileName: "case.js"},
		{Code: "console.log('abc', \"def\");", FileName: "case.js"},
		{Code: "console.log(`abc`, \"def\");", FileName: "case.js"},
		{Code: "console.log(`\nabc\ndef\n`);", FileName: "case.js"},
		{Code: "console.log(' ', \"def\");", FileName: "case.js"},
		{Code: "console.log(\" \");", FileName: "case.js"},
		{Code: "console.log(\" \", \"b\");", FileName: "case.js"},
		{Code: "console.log(\"a\", \" \");", FileName: "case.js"},
		{Code: "console.log(\" \", \"b\", \"c\");", FileName: "case.js"},
		{Code: "console.log(\"a\", \" \", \"c\");", FileName: "case.js"},
		{Code: "console.log(\"a\", \"b\", \" \");", FileName: "case.js"},
		{Code: "console.log('  ', \"def\");", FileName: "case.js"},
		{Code: "console.log(\"abc  \", \"def\");", FileName: "case.js"},
		{Code: "console.log(\"abc\\t\", \"def\");", FileName: "case.js"},
		{Code: "console.log(\"abc\\n\", \"def\");", FileName: "case.js"},
		{Code: "console.log(\"  abc\", \"def\");", FileName: "case.js"},
		{Code: "console.log(\" abc\", \"def\");", FileName: "case.js"},
		{Code: "console.log(\"abc\", \"def \");", FileName: "case.js"},
		{Code: "console.log();", FileName: "case.js"},
		{Code: "console.log(\"\");", FileName: "case.js"},
		{Code: "console.log(123);", FileName: "case.js"},
		{Code: "console.log(null);", FileName: "case.js"},
		{Code: "console.log(undefined);", FileName: "case.js"},
		{Code: "console.dir(\"abc \");", FileName: "case.js"},
		{Code: "new console.log(\" a \", \" b \");", FileName: "case.js"},
		{Code: "new console.debug(\" a \", \" b \");", FileName: "case.js"},
		{Code: "new console.info(\" a \", \" b \");", FileName: "case.js"},
		{Code: "new console.warn(\" a \", \" b \");", FileName: "case.js"},
		{Code: "new console.error(\" a \", \" b \");", FileName: "case.js"},
		{Code: "log(\" a \", \" b \");", FileName: "case.js"},
		{Code: "debug(\" a \", \" b \");", FileName: "case.js"},
		{Code: "info(\" a \", \" b \");", FileName: "case.js"},
		{Code: "warn(\" a \", \" b \");", FileName: "case.js"},
		{Code: "error(\" a \", \" b \");", FileName: "case.js"},
		{Code: "console[\"log\"](\" a \", \" b \");", FileName: "case.js"},
		{Code: "console[\"debug\"](\" a \", \" b \");", FileName: "case.js"},
		{Code: "console[\"info\"](\" a \", \" b \");", FileName: "case.js"},
		{Code: "console[\"warn\"](\" a \", \" b \");", FileName: "case.js"},
		{Code: "console[\"error\"](\" a \", \" b \");", FileName: "case.js"},
		{Code: "console[log](\" a \", \" b \");", FileName: "case.js"},
		{Code: "console[debug](\" a \", \" b \");", FileName: "case.js"},
		{Code: "console[info](\" a \", \" b \");", FileName: "case.js"},
		{Code: "console[warn](\" a \", \" b \");", FileName: "case.js"},
		{Code: "console[error](\" a \", \" b \");", FileName: "case.js"},
		{Code: "console.foo(\" a \", \" b \");", FileName: "case.js"},
		{Code: "foo.log(\" a \", \" b \");", FileName: "case.js"},
		{Code: "foo.debug(\" a \", \" b \");", FileName: "case.js"},
		{Code: "foo.info(\" a \", \" b \");", FileName: "case.js"},
		{Code: "foo.warn(\" a \", \" b \");", FileName: "case.js"},
		{Code: "foo.error(\" a \", \" b \");", FileName: "case.js"},
		{Code: "lib.console.log(\" a \", \" b \");", FileName: "case.js"},
		{Code: "lib.console.debug(\" a \", \" b \");", FileName: "case.js"},
		{Code: "lib.console.info(\" a \", \" b \");", FileName: "case.js"},
		{Code: "lib.console.warn(\" a \", \" b \");", FileName: "case.js"},
		{Code: "lib.console.error(\" a \", \" b \");", FileName: "case.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		{Code: "console.log(\"abc \", \"def\");", FileName: "case.js", Output: []string{"console.log(\"abc\", \"def\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 17)}},
		{Code: "console.log(\"abc\", \" def\");", FileName: "case.js", Output: []string{"console.log(\"abc\", \"def\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "leading", 1, 21)}},
		{Code: "console.log(\" abc \", \"def\");", FileName: "case.js", Output: []string{"console.log(\" abc\", \"def\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 18)}},
		{Code: "console.debug(\"abc \", \"def\");", FileName: "case.js", Output: []string{"console.debug(\"abc\", \"def\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("debug", "trailing", 1, 19)}},
		{Code: "console.info(\"abc \", \"def\");", FileName: "case.js", Output: []string{"console.info(\"abc\", \"def\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("info", "trailing", 1, 18)}},
		{Code: "console.warn(\"abc \", \"def\");", FileName: "case.js", Output: []string{"console.warn(\"abc\", \"def\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("warn", "trailing", 1, 18)}},
		{Code: "console.error(\"abc \", \"def\");", FileName: "case.js", Output: []string{"console.error(\"abc\", \"def\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("error", "trailing", 1, 19)}},
		{Code: "console.log(\"abc\", \" def \", \"ghi\");", FileName: "case.js", Output: []string{"console.log(\"abc\", \"def\", \"ghi\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "leading", 1, 21), spaceError("log", "trailing", 1, 25)}},
		{Code: "console.log(\"abc \", \"def \", \"ghi\");", FileName: "case.js", Output: []string{"console.log(\"abc\", \"def\", \"ghi\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 17), spaceError("log", "trailing", 1, 25)}},
		{Code: "console.log('abc ', \"def\");", FileName: "case.js", Output: []string{"console.log('abc', \"def\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 17)}},
		{Code: "console.log(`abc `, \"def\");", FileName: "case.js", Output: []string{"console.log(`abc`, \"def\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 17)}},
		{Code: "console.log(`abc ${1 + 2} `, \"def\");", FileName: "case.js", Output: []string{"console.log(`abc ${1 + 2}`, \"def\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 26)}},
		{Code: "console.log(\n\t'abc',\n\t'def ',\n\t'ghi'\n);", FileName: "case.js", Output: []string{"console.log(\n\t'abc',\n\t'def',\n\t'ghi'\n);"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 3, 6)}},
		{Code: "console.error(\n\ttheme.error('✗'),\n\t'Verifying \"packaging\" fixture\\n ',\n\ttheme.error(errorMessage)\n);", FileName: "case.js", Output: []string{"console.error(\n\ttheme.error('✗'),\n\t'Verifying \"packaging\" fixture\\n',\n\ttheme.error(errorMessage)\n);"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("error", "trailing", 3, 34)}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_console_spaces.NoConsoleSpacesRule, valid, invalid)
}

func TestNoConsoleSpacesSnapshots(t *testing.T) {
	valid := []rule_tester.ValidTestCase{}
	invalid := []rule_tester.InvalidTestCase{
		{Code: "console.log(\"abc\", \" def \", \"ghi\");", FileName: "case.js", Output: []string{"console.log(\"abc\", \"def\", \"ghi\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "leading", 1, 21), spaceError("log", "trailing", 1, 25)}},
		{Code: "console.error(\n\ttheme.error('✗'),\n\t'Verifying \"packaging\" fixture\\n ',\n\ttheme.error(errorMessage)\n);", FileName: "case.js", Output: []string{"console.error(\n\ttheme.error('✗'),\n\t'Verifying \"packaging\" fixture\\n',\n\ttheme.error(errorMessage)\n);"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("error", "trailing", 3, 34)}},
		{Code: "console.log(\n\t'abc',\n\t'def ',\n\t'ghi'\n);", FileName: "case.js", Output: []string{"console.log(\n\t'abc',\n\t'def',\n\t'ghi'\n);"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 3, 6)}},
		{Code: "console.log(\"_\", \" leading\", \"_\")", FileName: "case.js", Output: []string{"console.log(\"_\", \"leading\", \"_\")"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "leading", 1, 19)}},
		{Code: "console.log(\"_\", \"trailing \", \"_\")", FileName: "case.js", Output: []string{"console.log(\"_\", \"trailing\", \"_\")"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 27)}},
		{Code: "console.log(\"_\", \" leading and trailing \", \"_\")", FileName: "case.js", Output: []string{"console.log(\"_\", \"leading and trailing\", \"_\")"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "leading", 1, 19), spaceError("log", "trailing", 1, 40)}},
		{Code: "console.log(\"_\", \" log \", \"_\")", FileName: "case.js", Output: []string{"console.log(\"_\", \"log\", \"_\")"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "leading", 1, 19), spaceError("log", "trailing", 1, 23)}},
		{Code: "console.debug(\"_\", \" debug \", \"_\")", FileName: "case.js", Output: []string{"console.debug(\"_\", \"debug\", \"_\")"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("debug", "leading", 1, 21), spaceError("debug", "trailing", 1, 27)}},
		{Code: "console.info(\"_\", \" info \", \"_\")", FileName: "case.js", Output: []string{"console.info(\"_\", \"info\", \"_\")"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("info", "leading", 1, 20), spaceError("info", "trailing", 1, 25)}},
		{Code: "console.warn(\"_\", \" warn \", \"_\")", FileName: "case.js", Output: []string{"console.warn(\"_\", \"warn\", \"_\")"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("warn", "leading", 1, 20), spaceError("warn", "trailing", 1, 25)}},
		{Code: "console.error(\"_\", \" error \", \"_\")", FileName: "case.js", Output: []string{"console.error(\"_\", \"error\", \"_\")"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("error", "leading", 1, 21), spaceError("error", "trailing", 1, 27)}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_console_spaces.NoConsoleSpacesRule, valid, invalid)
}

func TestNoConsoleSpacesDocumentation(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "console.log('abc', 'def');", FileName: "case.js"},
		{Code: "console.debug('abc', 'def');", FileName: "case.js"},
		{Code: "console.info('abc', 'def');", FileName: "case.js"},
		{Code: "console.warn('abc', 'def');", FileName: "case.js"},
		{Code: "console.error('abc', 'def');", FileName: "case.js"},
		{Code: "console.log('abc ');", FileName: "case.js"},
		{Code: "console.log(' abc');", FileName: "case.js"},
		{Code: "console.log('abc  ', 'def');", FileName: "case.js"},
		{Code: "console.log('abc\\t', 'def');", FileName: "case.js"},
		{Code: "console.log('abc\\n', 'def');", FileName: "case.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		{Code: "console.log('abc ', 'def');", FileName: "case.js", Output: []string{"console.log('abc', 'def');"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 17)}},
		{Code: "console.log('abc', ' def');", FileName: "case.js", Output: []string{"console.log('abc', 'def');"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "leading", 1, 21)}},
		{Code: "console.log(\"abc \", \" def\");", FileName: "case.js", Output: []string{"console.log(\"abc\", \"def\");"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 17), spaceError("log", "leading", 1, 22)}},
		{Code: "console.log(`abc `, ` def`);", FileName: "case.js", Output: []string{"console.log(`abc`, `def`);"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("log", "trailing", 1, 17), spaceError("log", "leading", 1, 22)}},
		{Code: "console.debug('abc ', 'def');", FileName: "case.js", Output: []string{"console.debug('abc', 'def');"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("debug", "trailing", 1, 19)}},
		{Code: "console.info('abc ', 'def');", FileName: "case.js", Output: []string{"console.info('abc', 'def');"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("info", "trailing", 1, 18)}},
		{Code: "console.warn('abc ', 'def');", FileName: "case.js", Output: []string{"console.warn('abc', 'def');"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("warn", "trailing", 1, 18)}},
		{Code: "console.error('abc ', 'def');", FileName: "case.js", Output: []string{"console.error('abc', 'def');"}, Errors: []rule_tester.InvalidTestCaseError{spaceError("error", "trailing", 1, 19)}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_console_spaces.NoConsoleSpacesRule, valid, invalid)
}
