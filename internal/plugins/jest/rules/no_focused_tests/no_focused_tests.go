package no_focused_tests

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
)

// Config carries the framework-specific inputs for the no-focused-tests rule.
// Jest supports a focus name prefix (fit/fdescribe); frameworks without such
// aliases (e.g. Rstest) leave FocusPrefix empty and only detect the `.only`
// modifier.
type Config struct {
	// Name is the reported rule name, e.g. "jest/no-focused-tests".
	Name string
	// ParseConfig drives the shared function-call parser (import module and
	// valid call chains).
	ParseConfig jestUtils.FnCallParseConfig
	// FocusPrefix is the single-character prefix that marks a focused alias
	// (Jest: "f" for fit/fdescribe). Empty means the framework has no such
	// alias, so only the `.only` modifier is reported.
	FocusPrefix string
}

// Message Builders

func buildErrorFocusedTestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "focusedTest",
		Description: "Unexpected focused test",
	}
}

func buildErrorSuggestRemoveFocusMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestRemoveFocus",
		Description: "Suggest removing focus from test",
	}
}

// NewRule creates a no-focused-tests rule for a test framework.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name: config.Name,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					jestFnCall := jestUtils.ParseFnCall(node, ctx, config.ParseConfig)
					if jestFnCall == nil ||
						(jestFnCall.Kind != jestUtils.JestFnTypeDescribe &&
							jestFnCall.Kind != jestUtils.JestFnTypeTest) {
						return
					}

					if config.FocusPrefix != "" && strings.HasPrefix(jestFnCall.Name, config.FocusPrefix) {
						callExpr := node.AsCallExpression()
						if callExpr == nil {
							return
						}

						callee := ast.SkipParentheses(callExpr.Expression)
						if callee == nil {
							return
						}

						calleeRange := rslintUtils.TrimNodeTextRange(ctx.SourceFile, callee)
						if jestFnCall.Head.Type == jestUtils.JEST_IMPORT_MODE && jestFnCall.Name != jestFnCall.Head.Local.Value {
							reportNode := jestFnCall.Head.Local.Node
							if reportNode == nil {
								reportNode = callee
							}

							ctx.ReportNode(reportNode, buildErrorFocusedTestMessage())
						} else {
							reportNode := jestFnCall.Head.Local.Node
							if reportNode == nil {
								reportNode = callee
							}

							ctx.ReportNodeWithSuggestions(
								reportNode,
								buildErrorFocusedTestMessage(),
								rule.RuleSuggestion{
									Message: buildErrorSuggestRemoveFocusMessage(),
									FixesArr: []rule.RuleFix{
										rule.RuleFixRemoveRange(core.NewTextRange(calleeRange.Pos(), calleeRange.Pos()+1)),
									},
								},
							)
						}
					} else {
						idx := slices.IndexFunc(jestFnCall.MemberEntries, func(entry jestUtils.ParsedJestFnMemberEntry) bool {
							return entry.Name == "only"
						})
						if idx >= 0 {
							entry := jestFnCall.MemberEntries[idx]
							startRange := entry.Node.Loc.Pos() - 1
							endRange := entry.Node.Loc.End()
							if entry.Node.Kind != ast.KindIdentifier {
								endRange = entry.Node.End() + 1
							}

							ctx.ReportNodeWithSuggestions(
								entry.Node,
								buildErrorFocusedTestMessage(),
								rule.RuleSuggestion{
									Message: buildErrorSuggestRemoveFocusMessage(),
									FixesArr: []rule.RuleFix{
										rule.RuleFixRemoveRange(core.NewTextRange(startRange, endRange)),
									},
								},
							)
						}
					}
				},
			}
		},
	}
}

var NoFocusedTestsRule = NewRule(Config{
	Name:        "jest/no-focused-tests",
	ParseConfig: jestUtils.JestFnCallParseConfig(),
	FocusPrefix: "f",
})
