package jsx_fragments

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestJsxFragmentsUpstream migrates the full valid/invalid suite from upstream
// tests/lib/rules/jsx-fragments.js 1:1. Position assertions cover line/column
// for every invalid case. rslint-specific lock-in cases live in the
// jsx_fragments_extras_test.go file.
func TestJsxFragmentsUpstream(t *testing.T) {
	settings := reactSettings("16.2", "Act", "Frag")
	settingsOld := reactSettings("16.1", "Act", "Frag")

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxFragmentsRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		{Code: `<><Foo /></>`, Tsx: true, Settings: settings},
		{Code: `<Act.Frag><Foo /></Act.Frag>`, Tsx: true, Options: []interface{}{modeElement}, Settings: settings},
		{Code: `<Act.Frag />`, Tsx: true, Options: []interface{}{modeElement}, Settings: settings},
		{
			Code:     "import Act, { Frag as F } from 'react';\n<F><Foo /></F>;",
			Tsx:      true,
			Options:  []interface{}{modeElement},
			Settings: settings,
		},
		{
			Code:     "const F = Act.Frag;\n<F><Foo /></F>;",
			Tsx:      true,
			Options:  []interface{}{modeElement},
			Settings: settings,
		},
		{
			Code:     "const { Frag } = Act;\n<Frag><Foo /></Frag>;",
			Tsx:      true,
			Options:  []interface{}{modeElement},
			Settings: settings,
		},
		{
			Code:     "const { Frag } = require('react');\n<Frag><Foo /></Frag>;",
			Tsx:      true,
			Options:  []interface{}{modeElement},
			Settings: settings,
		},
		{Code: `<Act.Frag key="key"><Foo /></Act.Frag>`, Tsx: true, Options: []interface{}{modeSyntax}, Settings: settings},
		{Code: `<Act.Frag key="key" />`, Tsx: true, Options: []interface{}{modeSyntax}, Settings: settings},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		{
			Code:     `<><Foo /></>`,
			Tsx:      true,
			Settings: settingsOld,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "fragmentsNotSupported", Message: fragmentsNotSupportedDescription, Line: 1, Column: 1},
			},
		},
		{
			Code:     `<Act.Frag><Foo /></Act.Frag>`,
			Tsx:      true,
			Settings: settingsOld,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "fragmentsNotSupported", Line: 1, Column: 1},
			},
		},
		{
			Code:     `<Act.Frag />`,
			Tsx:      true,
			Settings: settingsOld,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "fragmentsNotSupported", Line: 1, Column: 1},
			},
		},
		{
			Code:     `<><Foo /></>`,
			Output:   []string{`<Act.Frag><Foo /></Act.Frag>`},
			Tsx:      true,
			Options:  []interface{}{modeElement},
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferPragma", Message: preferPragmaDescription("Act", "Frag"), Line: 1, Column: 1},
			},
		},
		{
			// SKIP: rslint uses one parser and has opening/closing fragment
			// nodes, so upstream's old TypeScript-parser no-fix variant does
			// not apply.
			Code:     `<><Foo /></>`,
			Skip:     true,
			Tsx:      true,
			Options:  []interface{}{modeElement},
			Settings: settings,
		},
		{
			Code:     `<Act.Frag><Foo /></Act.Frag>`,
			Output:   []string{`<><Foo /></>`},
			Tsx:      true,
			Options:  []interface{}{modeSyntax},
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Message: preferFragmentDescription("Act", "Frag"), Line: 1, Column: 1},
			},
		},
		{
			Code:     `<Act.Frag />`,
			Output:   []string{`<></>`},
			Tsx:      true,
			Options:  []interface{}{modeSyntax},
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 1, Column: 1},
			},
		},
		{
			Code:     "import Act, { Frag as F } from 'react';\n<F />;",
			Output:   []string{"import Act, { Frag as F } from 'react';\n<></>;"},
			Tsx:      true,
			Options:  []interface{}{modeSyntax},
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},
		{
			Code:     "import Act, { Frag as F } from 'react';\n<F><Foo /></F>;",
			Output:   []string{"import Act, { Frag as F } from 'react';\n<><Foo /></>;"},
			Tsx:      true,
			Options:  []interface{}{modeSyntax},
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},
		{
			Code:     "import Act, { Frag } from 'react';\n<Frag><Foo /></Frag>;",
			Output:   []string{"import Act, { Frag } from 'react';\n<><Foo /></>;"},
			Tsx:      true,
			Options:  []interface{}{modeSyntax},
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},
		{
			Code:     "const F = Act.Frag;\n<F><Foo /></F>;",
			Output:   []string{"const F = Act.Frag;\n<><Foo /></>;"},
			Tsx:      true,
			Options:  []interface{}{modeSyntax},
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},
		{
			Code:     "const { Frag } = Act;\n<Frag><Foo /></Frag>;",
			Output:   []string{"const { Frag } = Act;\n<><Foo /></>;"},
			Tsx:      true,
			Options:  []interface{}{modeSyntax},
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},
		{
			Code:     "const { Frag } = require('react');\n<Frag><Foo /></Frag>;",
			Output:   []string{"const { Frag } = require('react');\n<><Foo /></>;"},
			Tsx:      true,
			Options:  []interface{}{modeSyntax},
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},
	})
}

const fragmentsNotSupportedDescription = "Fragments are only supported starting from React v16.2. Please disable the `react/jsx-fragments` rule in `eslint` settings or upgrade your version of React."

func reactSettings(version, pragma, fragment string) map[string]interface{} {
	return map[string]interface{}{
		"react": map[string]interface{}{
			"version":  version,
			"pragma":   pragma,
			"fragment": fragment,
		},
	}
}

func preferPragmaDescription(react, fragment string) string {
	return "Prefer " + react + "." + fragment + " over fragment shorthand"
}

func preferFragmentDescription(react, fragment string) string {
	return "Prefer fragment shorthand over " + react + "." + fragment
}
