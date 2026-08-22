package no_alias_methods

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func buildReplaceAliasMessage(alias string, canonical string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceAlias",
		Description: fmt.Sprintf("Replace %s() with its canonical name of %s()", alias, canonical),
	}
}

type aliasHit struct {
	alias     string
	canonical string
	node      *ast.Node
}

var NoAliasMethodsRule = rule.Rule{
	Name:   "rstest/no-alias-methods",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil {
					return
				}

				hit, ok := firstAliasHit(parsed)
				if !ok {
					return
				}

				message := buildReplaceAliasMessage(hit.alias, hit.canonical)
				ctx.ReportNodeWithDeferredFixes(hit.node, message, func() []rule.RuleFix {
					fixRange, fixText, ok := testFramework.AccessorReplacement(
						ctx.SourceFile,
						hit.node,
						hit.canonical,
					)
					if !ok {
						return nil
					}
					return []rule.RuleFix{{
						Text:  fixText,
						Range: fixRange,
					}}
				})
			},
		}
	},
}

func firstAliasHit(parsed *rstestUtils.ParsedRstestExpectCall) (aliasHit, bool) {
	if parsed == nil ||
		parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
		parsed.MatcherEntry == nil {
		return aliasHit{}, false
	}

	canonical, ok := rstestUtils.RSTEST_MATCHER_ALIASES[parsed.Matcher]
	if !ok {
		return aliasHit{}, false
	}
	return aliasHit{
		alias:     parsed.Matcher,
		canonical: canonical,
		node:      parsed.MatcherEntry.Node,
	}, true
}
