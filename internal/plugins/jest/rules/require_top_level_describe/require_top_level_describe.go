package require_top_level_describe

import (
	_ "embed"
	"fmt"
	"math"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed require_top_level_describe.schema.json
var schemaJSON []byte

type options struct {
	MaxNumberOfTopLevelDescribes int
}

func parseOptions(raw any) options {
	opts := options{MaxNumberOfTopLevelDescribes: math.MaxInt}
	m := internalUtils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}

	if n, ok := internalUtils.CoerceInt(m["maxNumberOfTopLevelDescribes"]); ok && n >= 1 {
		opts.MaxNumberOfTopLevelDescribes = n
	}

	return opts
}

func buildTooManyDescribesMessage(maxAllowed int) rule.RuleMessage {
	s := "s"
	if maxAllowed == 1 {
		s = ""
	}
	return rule.RuleMessage{
		Id:          "tooManyDescribes",
		Description: fmt.Sprintf("There should not be more than %d describe%s at the top level", maxAllowed, s),
		Data: map[string]string{
			"max": strconv.Itoa(maxAllowed),
			"s":   s,
		},
	}
}

var (
	unexpectedTestCaseMessage = rule.RuleMessage{
		Id:          "unexpectedTestCase",
		Description: "All test cases must be wrapped in a describe block.",
	}
	unexpectedHookMessage = rule.RuleMessage{
		Id:          "unexpectedHook",
		Description: "All hooks must be wrapped in a describe block.",
	}
)

var RequireTopLevelDescribeRule = rule.Rule{
	Name:   "jest/require-top-level-describe",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		opts := parseOptions(rule.LegacyUnwrapOptions(_options))
		numberOfDescribeBlocks := 0
		numberOfTopLevelDescribeBlocks := 0

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil {
					return
				}

				if jestFnCall.Kind == utils.JestFnTypeDescribe {
					numberOfDescribeBlocks++

					if numberOfDescribeBlocks == 1 {
						numberOfTopLevelDescribeBlocks++
						if numberOfTopLevelDescribeBlocks > opts.MaxNumberOfTopLevelDescribes {
							ctx.ReportNode(node, buildTooManyDescribesMessage(opts.MaxNumberOfTopLevelDescribes))
						}
					}
					return
				}

				if numberOfDescribeBlocks == 0 {
					switch jestFnCall.Kind {
					case utils.JestFnTypeTest:
						ctx.ReportNode(node, unexpectedTestCaseMessage)
					case utils.JestFnTypeHook:
						ctx.ReportNode(node, unexpectedHookMessage)
					}
				}
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				if !utils.IsTypeOfJestFnCall(node, ctx, utils.JestFnTypeDescribe) {
					return
				}
				if numberOfDescribeBlocks > 0 {
					numberOfDescribeBlocks--
				}
			},
		}
	},
}
