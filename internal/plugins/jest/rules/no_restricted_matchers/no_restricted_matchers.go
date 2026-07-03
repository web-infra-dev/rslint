package no_restricted_matchers

import (
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type restrictedMatcher struct {
	Chain   string
	Parts   []string
	Message string
}

func buildRestrictedChainMessage(chain string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "restrictedChain",
		Description: "Use of `" + chain + "` is disallowed",
		Data: map[string]string{
			"restriction": chain,
		},
	}
}

func buildRestrictedChainWithMessage(chain string, message string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "restrictedChainWithMessage",
		Description: message,
		Data: map[string]string{
			"message":     message,
			"restriction": chain,
		},
	}
}

func parseOptions(options any) []restrictedMatcher {
	normalized := rule.NormalizeOptions(options)
	if len(normalized) == 0 {
		return nil
	}

	raw, ok := normalized[0].(map[string]interface{})
	if !ok {
		return nil
	}

	matchers := make([]restrictedMatcher, 0, len(raw))
	for chain, rawMessage := range raw {
		parts := strings.Split(chain, ".")
		if chain == "" || len(parts) == 0 {
			continue
		}

		valid := true
		for _, part := range parts {
			if part == "" {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}

		message, _ := rawMessage.(string)
		matchers = append(matchers, restrictedMatcher{
			Chain:   chain,
			Parts:   parts,
			Message: message,
		})
	}

	sort.Slice(matchers, func(i, j int) bool {
		if len(matchers[i].Parts) != len(matchers[j].Parts) {
			return len(matchers[i].Parts) > len(matchers[j].Parts)
		}
		return matchers[i].Chain < matchers[j].Chain
	})

	return matchers
}

func isChainRestricted(chain string, restriction restrictedMatcher) bool {
	if len(restriction.Parts) == 1 && jestUtils.EXPECT_MODIFIER_NAMES[restriction.Chain] {
		return strings.HasPrefix(chain, restriction.Chain)
	}

	if strings.HasSuffix(restriction.Chain, ".not") {
		return strings.HasPrefix(chain, restriction.Chain)
	}

	return chain == restriction.Chain
}

func restrictedRange(entries []jestUtils.ParsedJestFnMemberEntry) (core.TextRange, bool) {
	if len(entries) == 0 || entries[0].Node == nil || entries[len(entries)-1].Node == nil {
		return core.TextRange{}, false
	}

	return core.NewTextRange(entries[0].Node.Pos(), entries[len(entries)-1].Node.End()), true
}

var NoRestrictedMatchersRule = rule.Rule{
	Name: "jest/no-restricted-matchers",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		restrictedMatchers := parseOptions(options)
		if len(restrictedMatchers) == 0 {
			return rule.RuleListeners{}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil || jestFnCall.Kind != jestUtils.JestFnTypeExpect {
					return
				}

				chain := strings.Join(jestFnCall.Members, ".")
				reportRange, ok := restrictedRange(jestFnCall.MemberEntries)
				if !ok {
					return
				}

				for _, restricted := range restrictedMatchers {
					if !isChainRestricted(chain, restricted) {
						continue
					}

					if restricted.Message != "" {
						ctx.ReportRange(reportRange, buildRestrictedChainWithMessage(restricted.Chain, restricted.Message))
					} else {
						ctx.ReportRange(reportRange, buildRestrictedChainMessage(restricted.Chain))
					}
					return
				}
			},
		}
	},
}
