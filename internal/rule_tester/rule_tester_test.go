package rule_tester

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestResolveTestCaseFileName(t *testing.T) {
	t.Parallel()

	first := resolveTestCaseFileName("", false)
	second := resolveTestCaseFileName("", false)
	tsx := resolveTestCaseFileName("", true)

	if first == second {
		t.Fatalf("default test filenames must be unique, got %q twice", first)
	}
	if !strings.HasPrefix(first, "filename-") || !strings.HasSuffix(first, ".ts") {
		t.Fatalf("default TypeScript filename = %q, want filename-<id>.ts", first)
	}
	if !strings.HasPrefix(tsx, "filename-") || !strings.HasSuffix(tsx, ".tsx") {
		t.Fatalf("default TSX filename = %q, want filename-<id>.tsx", tsx)
	}
	if explicit := resolveTestCaseFileName("explicit.jsx", false); explicit != "explicit.jsx" {
		t.Fatalf("explicit filename = %q, want explicit.jsx", explicit)
	}
}

func TestRunRuleTesterPropagatesLanguageOptions(t *testing.T) {
	probe := rule.Rule{
		Name: "language-options-probe",
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			want, _ := options[0].(string)
			return rule.RuleListeners{
				ast.KindExpressionStatement: func(_ *ast.Node) {
					if ctx.LanguageOptions.EffectiveSourceType() != want {
						ctx.ReportRange(core.NewTextRange(0, 1), rule.RuleMessage{
							Id:          "unexpectedSourceType",
							Description: "unexpected source type",
						})
					}
				},
			}
		},
	}

	RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&probe,
		[]ValidTestCase{{
			Code:            `"script";`,
			Options:         []any{"script"},
			LanguageOptions: rule.LanguageOptions{SourceType: "script"},
		}},
		[]InvalidTestCase{{
			Code:            `"script";`,
			Options:         []any{"script"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []InvalidTestCaseError{{
				MessageId: "unexpectedSourceType",
				Line:      1,
				Column:    1,
				EndLine:   1,
				EndColumn: 2,
			}},
		}},
	)
}

func TestConvertESLintTestCasesPreservesFilenameGlobalsAndTSX(t *testing.T) {
	var suite ESLintTestSuite
	err := json.Unmarshal([]byte(`{
		"valid": [{
			"code": "<div />;",
			"filename": "example.tsx",
			"globals": {"console": "readonly"},
			"tsx": true
		}],
		"invalid": [{
			"code": "<div />;",
			"filename": "example.tsx",
			"globals": {"window": "writable"},
			"tsx": true,
			"errors": []
		}]
	}`), &suite)
	if err != nil {
		t.Fatalf("unmarshal ESLint test suite: %v", err)
	}

	valid := ConvertESLintTestCase(suite.Valid[0])
	if valid.FileName != "example.tsx" {
		t.Fatalf("valid filename = %q, want example.tsx", valid.FileName)
	}
	if !valid.Tsx {
		t.Fatal("valid tsx = false, want true")
	}
	if valid.Globals["console"] != "readonly" {
		t.Fatalf("valid globals = %+v, want console readonly", valid.Globals)
	}

	invalid := ConvertESLintInvalidTestCase(suite.Invalid[0])
	if invalid.FileName != "example.tsx" {
		t.Fatalf("invalid filename = %q, want example.tsx", invalid.FileName)
	}
	if !invalid.Tsx {
		t.Fatal("invalid tsx = false, want true")
	}
	if invalid.Globals["window"] != "writable" {
		t.Fatalf("invalid globals = %+v, want window writable", invalid.Globals)
	}
}
