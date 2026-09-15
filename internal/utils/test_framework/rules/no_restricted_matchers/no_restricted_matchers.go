package no_restricted_matchers

import (
	_ "embed"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed no_restricted_matchers.schema.json
var schemaJSON []byte

type ExpectChain struct {
	Names   []string
	Entries []testFramework.MemberEntry
}

type Runtime struct {
	Parse func(*ast.Node) *ExpectChain
}

type Config struct {
	Name      string
	Prepare   func(rule.RuleContext) Runtime
	Modifiers map[string]bool
}

type restrictedMatcher struct {
	chain   string
	parts   []string
	message string
}

func parseOptions(options []any) []restrictedMatcher {
	if len(options) == 0 {
		return nil
	}

	raw, ok := options[0].(map[string]any)
	if !ok {
		return nil
	}

	matchers := make([]restrictedMatcher, 0, len(raw))
	for chain, rawMessage := range raw {
		parts := strings.Split(chain, ".")
		if chain == "" || slicesContainEmpty(parts) {
			continue
		}

		message, _ := rawMessage.(string)
		matchers = append(matchers, restrictedMatcher{
			chain:   chain,
			parts:   parts,
			message: message,
		})
	}

	sort.Slice(matchers, func(i, j int) bool {
		if len(matchers[i].parts) != len(matchers[j].parts) {
			return len(matchers[i].parts) > len(matchers[j].parts)
		}
		return matchers[i].chain < matchers[j].chain
	})
	return matchers
}

func slicesContainEmpty(parts []string) bool {
	for _, part := range parts {
		if part == "" {
			return true
		}
	}
	return false
}

func isChainRestricted(chain string, restriction restrictedMatcher, modifiers map[string]bool) bool {
	if len(restriction.parts) == 1 && modifiers[restriction.chain] {
		return strings.HasPrefix(chain, restriction.chain)
	}
	if strings.HasSuffix(restriction.chain, ".not") {
		return strings.HasPrefix(chain, restriction.chain)
	}
	return chain == restriction.chain
}

func restrictedChainMessage(restriction string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "restrictedChain",
		Description: "Use of `" + restriction + "` is disallowed",
		Data:        map[string]string{"restriction": restriction},
	}
}

func restrictedChainWithMessage(restriction string, message string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "restrictedChainWithMessage",
		Description: message,
		Data: map[string]string{
			"message":     message,
			"restriction": restriction,
		},
	}
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.NewSchema(schemaJSON),
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			restrictions := parseOptions(options)
			if len(restrictions) == 0 {
				return rule.RuleListeners{}
			}

			runtime := config.Prepare(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					parsed := runtime.Parse(node)
					if parsed == nil || len(parsed.Names) == 0 {
						return
					}

					chain := strings.Join(parsed.Names, ".")
					for _, restriction := range restrictions {
						if !isChainRestricted(chain, restriction, config.Modifiers) {
							continue
						}

						reportRange, ok := testFramework.MemberEntriesRange(ctx.SourceFile, parsed.Entries)
						if !ok {
							return
						}
						if restriction.message == "" {
							ctx.ReportRange(reportRange, restrictedChainMessage(restriction.chain))
						} else {
							ctx.ReportRange(reportRange, restrictedChainWithMessage(restriction.chain, restriction.message))
						}
						return
					}
				},
			}
		},
	}
}
