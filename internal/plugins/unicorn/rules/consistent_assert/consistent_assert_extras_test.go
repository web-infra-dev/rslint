package consistent_assert_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/consistent_assert"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentAssertExtras(t *testing.T) {
	aliasCode := "import {strict as verify} from 'node:assert';\nverify(value);"
	aliasOutput := "import {strict as verify} from 'node:assert';\nverify.ok(value);"

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&consistent_assert.ConsistentAssertRule,
		[]rule_tester.ValidTestCase{
			valid("import assert from 'node:test';\nassert(value);"),
			valid("const assert = value => value;\nassert(value);"),
			valid("import assert from 'node:assert';\nassert.call(null, value);"),
			valid("import assert from 'node:assert';\nconsume(assert);"),
			validTS("import assert from 'node:assert';\n(assert as unknown as ((value: unknown) => void))(value);"),
			validTS("import {type strict as assert} from 'assert';\nassert(value);"),
			valid("import {strict as assert} from 'node:assert/strict';\nassert(value);"),
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:            aliasCode,
				FileName:        "file.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Output:          []string{aliasOutput},
				Errors:          []rule_tester.InvalidTestCaseError{expectedError(aliasCode, "verify", 1)},
			},
			invalid(
				"import assert from 'node:assert';\n((assert))(value);",
				"import assert from 'node:assert';\n((assert.ok))(value);",
				"assert",
				1,
			),
		},
	)
}
