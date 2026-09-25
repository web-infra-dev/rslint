package no_empty_named_blocks_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_empty_named_blocks"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const emptyBlockMessage = "Unexpected empty named import block"

func emptyBlockError(line, column, endLine, endColumn int, suggestions ...string) []rule_tester.InvalidTestCaseError {
	expected := rule_tester.InvalidTestCaseError{
		MessageId: "", Message: emptyBlockMessage,
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
	for _, output := range suggestions {
		expected.Suggestions = append(expected.Suggestions, rule_tester.InvalidTestCaseSuggestion{MessageId: "", Output: output})
	}
	return []rule_tester.InvalidTestCaseError{expected}
}

// Every case from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-empty-named-blocks.js
func TestNoEmptyNamedBlocksUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `import 'mod';`},
		{Code: `import Default from 'mod';`},
		{Code: `import { Named } from 'mod';`},
		{Code: `import Default, { Named } from 'mod';`},
		{Code: `import * as Namespace from 'mod';`},
		// TypeScript.
		{Code: `import type Default from 'mod';`},
		{Code: `import type { Named } from 'mod';`},
		{Code: `import type Default, { Named } from 'mod';`},
		{Code: `import type * as Namespace from 'mod';`},
		// Flow's typeof imports are unsupported by the TypeScript parser.
		{Code: `import typeof Default from 'mod'; // babel old`, Skip: true},
		{Code: `import typeof { Named } from 'mod'; // babel old`, Skip: true},
		{Code: `import typeof Default, { Named } from 'mod'; // babel old`, Skip: true},
		{Code: `
        module.exports = {
          rules: {
            'keyword-spacing': ['error', {overrides: {}}],
          }
        };
      `},
		{Code: `
        import { DESCRIPTORS, NODE } from '../helpers/constants';
        // ...
        import { timeLimitedPromise } from '../helpers/helpers';
        // ...
        import { DESCRIPTORS2 } from '../helpers/constants';
      `},
	}
	invalid := []rule_tester.InvalidTestCase{
		{
			Code: `import Default, {} from 'mod';`, Output: []string{`import Default from 'mod';`},
			Errors: emptyBlockError(1, 1, 1, 31),
		},
		{
			Code: `import type Default, {} from 'mod';`, Output: []string{`import type Default from 'mod';`},
			Errors: emptyBlockError(1, 1, 1, 36),
		},
		// Flow's typeof imports are unsupported by the TypeScript parser.
		{
			Code: `import typeof Default, {} from 'mod';`, Output: []string{`import typeof Default from 'mod';`},
			Errors: emptyBlockError(1, 1, 1, 38), Skip: true,
		},
	}
	for _, code := range []string{
		`import {} from 'mod';`,
		`import{}from'mod';`,
		`import {} from'mod';`,
		`import {}from 'mod';`,
		// TypeScript.
		`import type {} from 'mod';`,
		`import type {}from 'mod';`,
		`import type{}from 'mod';`,
		`import type {}from'mod';`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: code, Errors: emptyBlockError(1, 1, 1, len(code)+1, "", `import 'mod';`),
		})
	}
	for _, code := range []string{
		`import typeof {} from 'mod';`,
		`import typeof {}from 'mod';`,
		`import typeof {} from'mod';`,
		`import typeof{}from'mod';`,
	} {
		// Retain all Flow suggestion cases as explained skips.
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: code, Errors: emptyBlockError(1, 1, 1, len(code)+1, "", `import 'mod';`), Skip: true,
		})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_empty_named_blocks.NoEmptyNamedBlocksRule, valid, invalid)
}

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-empty-named-blocks.md
func TestNoEmptyNamedBlocksDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_empty_named_blocks.NoEmptyNamedBlocksRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { mod } from 'mod'`},
			{Code: `import Default, { mod } from 'mod'`},
			{Code: `import type { mod } from 'mod'`},
			// Flow typeof imports are unsupported.
			{Code: `import typeof { mod } from 'mod'`, Skip: true},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `import {} from 'mod'`, Errors: emptyBlockError(1, 1, 1, 21, "", `import 'mod'`)},
			{
				Code: `import Default, {} from 'mod'`, Output: []string{`import Default from 'mod'`},
				Errors: emptyBlockError(1, 1, 1, 30),
			},
			{
				Code: `import type Default, {} from 'mod'`, Output: []string{`import type Default from 'mod'`},
				Errors: emptyBlockError(1, 1, 1, 35),
			},
			{Code: `import type {} from 'mod'`, Errors: emptyBlockError(1, 1, 1, 26, "", `import 'mod'`)},
			// Flow typeof imports are unsupported.
			{Code: `import typeof {} from 'mod'`, Errors: emptyBlockError(1, 1, 1, 28, "", `import 'mod'`), Skip: true},
			{
				Code: `import typeof Default, {} from 'mod'`, Output: []string{`import typeof Default from 'mod'`},
				Errors: emptyBlockError(1, 1, 1, 37), Skip: true,
			},
		})
}
