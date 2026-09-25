package prefer_structured_clone_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_structured_clone"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferStructuredCloneBoundaries(t *testing.T) {
	parenthesized := invalidFunction(
		"((_.cloneDeep))(foo)",
		"_.cloneDeep",
		"((structuredClone))(foo)",
	)
	spacedConfig := invalidFunction(
		"my.cloneDeep(foo)",
		"my.cloneDeep",
		"structuredClone(foo)",
		"  my.cloneDeep  ",
	)
	comments := invalidJSON(
		"JSON.parse((/*a*/JSON.stringify/*b*/)(foo))",
		"structuredClone(/*a*//*b*/foo)",
	)
	jsdocCallee := invalidJSON(
		"JSON.parse(/** @type {any} */ (JSON.stringify)(foo))",
		"structuredClone(/** @type {any} */ foo)",
	)
	tsxTypeArguments := invalidJSON(
		"JSON.parse(JSON.stringify<string>(foo))",
		"structuredClone(foo)",
	)
	tsxTypeArguments.FileName = "case.tsx"
	tsxTypeArguments.Tsx = true

	// Deliberate safety divergence from upstream v75: when JSON.parse is also
	// configured as a custom clone function, emit only the specialized JSON
	// diagnostic. Upstream emits a second suggestion that clones the stringify
	// result string and therefore changes semantics.
	overlap := invalidJSON(
		"JSON.parse(JSON.stringify(foo))",
		"structuredClone(foo)",
	)
	overlap.Options = []any{map[string]any{"functions": []any{"JSON.parse"}}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_structured_clone.PreferStructuredCloneRule,
		[]rule_tester.ValidTestCase{
			{
				Code:            "JSON.parse((JSON.stringify as any)(foo))",
				FileName:        "case.ts",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
			valid("JSON.parse(JSON[\"stringify\"](foo))"),
			valid("JSON[\"parse\"](JSON.stringify(foo))"),
		},
		[]rule_tester.InvalidTestCase{
			parenthesized,
			spacedConfig,
			comments,
			jsdocCallee,
			tsxTypeArguments,
			overlap,
		},
	)
}
