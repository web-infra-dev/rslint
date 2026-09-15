package prefer_called_with

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type ExpectCall struct {
	Matcher      string
	MatcherEntry testFramework.MemberEntry
	Modifiers    []string
}

type Runtime struct {
	Parse func(*ast.Node) *ExpectCall
}

type Config struct {
	Name         string
	Prepare      func(rule.RuleContext) Runtime
	Replacements map[string]string
	Autofix      bool
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					parsed := runtime.Parse(node)
					if parsed == nil || slices.Contains(parsed.Modifiers, "not") {
						return
					}
					preferred, ok := config.Replacements[parsed.Matcher]
					if !ok {
						return
					}

					matcherName := preferred
					if !config.Autofix {
						// Jest's message data names the source matcher, not the replacement.
						matcherName = parsed.Matcher
					}
					message := rule.RuleMessage{
						Id:          "preferCalledWith",
						Description: "Prefer " + preferred + "(/* expected args */)",
						Data:        map[string]string{"matcherName": matcherName},
					}
					reportNode := parsed.MatcherEntry.Node
					if reportNode == nil {
						reportNode = node
					}
					if !config.Autofix {
						ctx.ReportNode(reportNode, message)
						return
					}
					ctx.ReportNodeWithDeferredFixes(reportNode, message, func() []rule.RuleFix {
						fixRange, fixText, ok := testFramework.AccessorReplacement(ctx.SourceFile, parsed.MatcherEntry.Node, preferred)
						if !ok {
							return nil
						}
						return []rule.RuleFix{rule.RuleFixReplaceRange(fixRange, fixText)}
					})
				},
			}
		},
	}
}
