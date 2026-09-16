package max_nested_describe

import (
	_ "embed"
	"fmt"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed max_nested_describe.schema.json
var schemaJSON []byte

const defaultMax = 5

type Runtime struct {
	IsDescribeCall func(*ast.Node) bool
	DescribeDepth  func(*ast.Node) int
	Skip           bool
}

type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
}

type options struct {
	max int
}

type describeFrame struct {
	node  *ast.Node
	depth int
}

func parseOptions(rawOptions []any) options {
	opts := options{max: defaultMax}
	if len(rawOptions) == 0 {
		return opts
	}

	optsMap, _ := rawOptions[0].(map[string]any)
	if maxValue, ok := internalUtils.CoerceIntegral(optsMap["max"]); ok && maxValue >= 0 {
		opts.max = maxValue
	}
	return opts
}

func exceededMaxDepthMessage(depth, maxAllowed int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "exceededMaxDepth",
		Description: fmt.Sprintf("Too many nested describe calls (%d) - maximum allowed is %d", depth, maxAllowed),
		Data: map[string]string{
			"depth": strconv.Itoa(depth),
			"max":   strconv.Itoa(maxAllowed),
		},
	}
}

// NewRule creates a maximum describe-depth rule for a test framework. The
// framework adapter owns call provenance and syntax; this engine owns the
// traversal stack so sibling calls cannot leak depth into one another.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.NewSchema(schemaJSON),
		Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			if runtime.Skip || runtime.IsDescribeCall == nil {
				return rule.RuleListeners{}
			}
			opts := parseOptions(rawOptions)
			describes := make([]describeFrame, 0, defaultMax+1)

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					if !runtime.IsDescribeCall(node) {
						return
					}
					depth := 1
					if len(describes) > 0 {
						depth = describes[len(describes)-1].depth + 1
					}
					if runtime.DescribeDepth != nil {
						if semanticDepth := runtime.DescribeDepth(node); semanticDepth > depth {
							depth = semanticDepth
						}
					}
					describes = append(describes, describeFrame{node: node, depth: depth})
					if depth > opts.max {
						ctx.ReportNode(node, exceededMaxDepthMessage(depth, opts.max))
					}
				},
				rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
					last := len(describes) - 1
					if last >= 0 && describes[last].node == node {
						describes = describes[:last]
					}
				},
			}
		},
	}
}
