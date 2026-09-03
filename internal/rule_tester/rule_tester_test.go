package rule_tester

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
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
