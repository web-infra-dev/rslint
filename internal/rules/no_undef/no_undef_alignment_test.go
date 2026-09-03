package no_undef

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUndefLatestParserReferenceParity(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUndefRule,
		[]rule_tester.ValidTestCase{
			{
				Code:     `const el = <éclair />; const other = <Missing:Part />;`,
				FileName: "jsx-intrinsic-and-namespaced.jsx",
				TSConfig: "tsconfig.allowJs.json",
				Tsx:      true,
			},
			{
				Code:     `const Éclair = null; const el = <Éclair></Éclair>;`,
				FileName: "jsx-declared-unicode.jsx",
				TSConfig: "tsconfig.allowJs.json",
				Tsx:      true,
			},
			{
				Code:     `type T = typeof this; type U = typeof this.member;`,
				FileName: "type-query-this.ts",
			},
			{
				Code:     `const foo = null, bar = null; const el = <foo:bar />;`,
				FileName: "declared-namespaced-tag.tsx",
				Tsx:      true,
			},
			{
				Code:     `const el = <this.Member />;`,
				FileName: "jsx-this-member.tsx",
				Tsx:      true,
			},
			{
				Code:     `const el = <this />;`,
				FileName: "configured-jsx-this.tsx",
				Tsx:      true,
				Globals:  map[string]any{"this": "readonly"},
			},
			{
				Code:     `const Missing = {}; export as namespace Missing;`,
				FileName: "declared-namespace-export.ts",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `const el = <Missing></Missing>;`,
				FileName: "jsx-closing-tag.jsx",
				TSConfig: "tsconfig.allowJs.json",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 13},
				},
			},
			{
				Code:     `const el = <Éclair />;`,
				FileName: "jsx-unicode-component.jsx",
				TSConfig: "tsconfig.allowJs.json",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 13},
				},
			},
			{
				Code:     `const el = <foo:bar></foo:bar>;`,
				FileName: "tsx-namespaced-tag.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 13},
					{MessageId: "undef", Line: 1, Column: 17},
					{MessageId: "undef", Line: 1, Column: 23},
					{MessageId: "undef", Line: 1, Column: 27},
				},
			},
			{
				Code:     `const el = <this></this>;`,
				FileName: "tsx-this-tag.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 13},
					{MessageId: "undef", Line: 1, Column: 20},
				},
			},
			{
				Code:     `const el = <this:Foo></this:Foo>;`,
				FileName: "tsx-this-namespaced-tag.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 13},
					{MessageId: "undef", Line: 1, Column: 18},
					{MessageId: "undef", Line: 1, Column: 24},
					{MessageId: "undef", Line: 1, Column: 29},
				},
			},
			{
				Code:     `export as namespace Missing;`,
				FileName: "missing-namespace-export.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 21},
				},
			},
		},
	)
}
