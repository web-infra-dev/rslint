package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnusedVarsJsxFactory covers the JSX factory pragma being marked as used
// whenever a file contains JSX, in every jsx runtime. This mirrors
// @typescript-eslint/parser's `jsxPragma` baseline (default "React"), which
// treats the factory as used on any JSX regardless of runtime. Before the fix
// this only worked for jsx: "preserve"/"react-native".
func TestNoUnusedVarsJsxFactory(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// react-jsx (automatic runtime): React is referenced only by JSX.
		{
			Code:     "import React from 'react';\nexport const A = () => <div />;\n",
			Tsx:      true,
			TSConfig: "tsconfig.react-jsx.json",
		},
		// react-jsx: fragment-only usage still marks the factory.
		{
			Code:     "import React from 'react';\nexport const A = () => <></>;\n",
			Tsx:      true,
			TSConfig: "tsconfig.react-jsx.json",
		},
		// react-jsx: mixed import where the named binding is used in code and
		// the default (React) is used only by JSX — neither is reported.
		{
			Code:     "import React, { useMemo } from 'react';\nexport const A = () => {\n  useMemo(() => 1, []);\n  return <div />;\n};\n",
			Tsx:      true,
			TSConfig: "tsconfig.react-jsx.json",
		},
		// classic react runtime: TS resolves JSX to React.createElement, but the
		// reference is implicit (no identifier node), so it must still be marked.
		{
			Code:     "import React from 'react';\nexport const A = () => <div />;\n",
			Tsx:      true,
			TSConfig: "tsconfig.react.json",
		},
		// preserve (default fixture): unchanged behavior — still not reported.
		{
			Code: "import React from 'react';\nexport const A = () => <div />;\n",
			Tsx:  true,
		},
		// File-level pragmas are resolved by the checker, including qualified
		// factories whose first identifier names the imported binding.
		{
			Code:     "/** @jsx X.createElement */\nimport X from './x';\nexport const A = () => <div />;\n",
			Tsx:      true,
			TSConfig: "tsconfig.react-jsx.json",
		},
		{
			Code:     "/** @jsx h */\n/** @jsxFrag Fragment */\nimport { h, Fragment } from './x';\nexport const A = () => <></>;\n",
			Tsx:      true,
			TSConfig: "tsconfig.react-jsx.json",
		},
	}

	invalid := []rule_tester.InvalidTestCase{
		// react-jsx: React imported but the file has no JSX at all → genuinely
		// unused, still reported.
		{
			Code:     "import React from 'react';\nexport const x = 1;\n",
			Tsx:      true,
			TSConfig: "tsconfig.react-jsx.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unusedVar", Line: 1, Column: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeUnusedImportDeclaration",
						Output:    "export const x = 1;\n",
					}},
				},
			},
		},
		// react-jsx: a non-pragma import that is unused must still be reported
		// even though the file contains JSX — only the JSX factory is exempt.
		{
			Code:     "import { Foo } from './foo';\nexport const A = () => <div />;\n",
			Tsx:      true,
			TSConfig: "tsconfig.react-jsx.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unusedVar", Line: 1, Column: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeUnusedImportDeclaration",
						Output:    "export const A = () => <div />;\n",
					}},
				},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, valid, invalid)
}
