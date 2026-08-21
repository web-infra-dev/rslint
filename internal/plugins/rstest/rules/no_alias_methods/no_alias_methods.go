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
					hit, ok = strictComputedIdentifierFallback(parsed, node)
					if !ok {
						return
					}
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

func strictComputedIdentifierFallback(
	parsed *rstestUtils.ParsedRstestExpectCall,
	listenerNode *ast.Node,
) (aliasHit, bool) {
	if parsed == nil ||
		parsed.Head == nil ||
		parsed.Reason != rstestUtils.RstestExpectParseReasonMatcherNotFound ||
		parsed.MatcherEntry != nil ||
		len(parsed.MemberEntries) == 0 {
		return aliasHit{}, false
	}

	last := parsed.MemberEntries[len(parsed.MemberEntries)-1]
	if !testFramework.IsComputedIdentifierAccessor(last.Node) || last.Call != listenerNode {
		return aliasHit{}, false
	}

	canonical, ok := rstestUtils.RSTEST_MATCHER_ALIASES[last.Name]
	if !ok {
		return aliasHit{}, false
	}

	if !strictFallbackPrefixAllowed(parsed.MemberEntries[:len(parsed.MemberEntries)-1], parsed.Head) {
		return aliasHit{}, false
	}

	return aliasHit{
		alias:     last.Name,
		canonical: canonical,
		node:      last.Node,
	}, true
}

func strictFallbackPrefixAllowed(
	prefix []rstestUtils.ParsedRstestFnMemberEntry,
	head *ast.Node,
) bool {
	notCount := 0
	promiseModifierCount := 0
	for _, entry := range prefix {
		if entry.Call == head {
			continue
		}
		if entry.Call != nil {
			return false
		}
		if testFramework.IsComputedIdentifierAccessor(entry.Node) {
			return false
		}
		if !rstestUtils.RSTEST_EXPECT_MODIFIER_NAMES[entry.Name] {
			return false
		}
		if entry.Name == "not" {
			notCount++
			if notCount > 1 {
				return false
			}
			continue
		}
		promiseModifierCount++
		if promiseModifierCount > 1 {
			return false
		}
	}
	return true
}
