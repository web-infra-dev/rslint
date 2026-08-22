package rule_tester

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
)

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
