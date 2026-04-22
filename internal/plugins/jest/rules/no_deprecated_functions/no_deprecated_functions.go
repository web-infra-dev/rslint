package no_deprecated_functions

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message Builder

func buildErrorDeprecatedFunctionMessage(deprecation string, replacement string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "deprecatedFunction",
		Description: fmt.Sprintf("`%s` has been deprecated in favor of `%s`", deprecation, replacement),
	}
}

type deprecation struct {
	deprecated   string
	replacement  string
	minJestMajor int
}

var allDeprecations = []deprecation{
	{"jest.resetModuleRegistry", "jest.resetModules", 15},
	{"jest.addMatchers", "expect.extend", 17},
	{"require.requireMock", "jest.requireMock", 21},
	{"require.requireActual", "jest.requireActual", 21},
	{"jest.runTimersToTime", "jest.advanceTimersByTime", 22},
	{"jest.genMockFromModule", "jest.createMockFromModule", 26},
}

func deprecatedFunctions(jestVersion int) map[string]string {
	m := make(map[string]string)
	for _, d := range allDeprecations {
		if jestVersion >= d.minJestMajor {
			m[d.deprecated] = d.replacement
		}
	}
	return m
}

func memberChainString(entries []utils.ParsedJestFnMemberEntry) string {
	if len(entries) == 0 {
		return ""
	}
	parts := make([]string, len(entries))
	for i, e := range entries {
		parts[i] = e.Name
	}
	return strings.Join(parts, ".")
}

var NoDeprecatedFunctionsRule = rule.Rule{
	Name: "jest/no-deprecated-functions",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		depMap := deprecatedFunctions(utils.JestVersionMajor(utils.GetJestVersion(ctx)))
		if len(depMap) == 0 {
			return rule.RuleListeners{}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := node.AsCallExpression().Expression
				if callee == nil {
					return
				}

				entries := utils.GetJestFnMemberEntries(callee)
				if len(entries) < 2 {
					return
				}

				chain := memberChainString(entries)
				replacement, ok := depMap[chain]
				if !ok {
					return
				}

				replacementCallee := replacement
				last := entries[len(entries)-1]

				if last.Node != nil && (last.Node.Kind == ast.KindStringLiteral || last.Node.Kind == ast.KindNoSubstitutionTemplateLiteral) {
					repParts := strings.SplitN(replacement, ".", 2)
					if len(repParts) == 2 {
						replacementCallee = fmt.Sprintf("%s['%s']", repParts[0], repParts[1])
					}
				}

				ctx.ReportNodeWithFixes(
					callee,
					buildErrorDeprecatedFunctionMessage(chain, replacement),
					rule.RuleFixReplace(ctx.SourceFile, callee, replacementCallee),
				)
			},
		}
	},
}
