package require_top_level_describe

import (
	_ "embed"
	"fmt"
	"math"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed require_top_level_describe.schema.json
var schemaJSON []byte

type ParsedCall = testFramework.ParsedCall

// Messages carries the wording a framework plugin reports unwrapped
// registrations with. The two plugins this engine serves word them
// differently, and each keeps its own established text.
type Messages struct {
	UnexpectedTestCase string
	UnexpectedHook     string
}

type Runtime struct {
	// Parse resolves a call expression to a framework registration, or nil
	// when the call is not one.
	Parse func(*ast.Node) *ParsedCall
	// DescribeDepth, when set, reports the deepest suite nesting level a
	// describe registration can execute at, including nesting a framework
	// expresses by passing a suite callback by reference. The engine takes
	// the deeper of this and the lexical stack depth, so a semantic answer
	// can only move a describe out of the top level, never into it.
	DescribeDepth func(*ast.Node) int
	// InsideDescribe, when set, reports whether a call executes inside a
	// describe callback that is not one of its lexical ancestors. It can only
	// suppress a report, so an adapter that cannot answer leaves it nil and
	// the engine falls back to lexical nesting alone.
	InsideDescribe func(*ast.Node) bool
}

type Config struct {
	Name     string
	Messages Messages
	Prepare  func(rule.RuleContext) Runtime
}

type options struct {
	maxNumberOfTopLevelDescribes int
}

type describeFrame struct {
	node  *ast.Node
	depth int
}

func parseOptions(rawOptions []any) options {
	opts := options{maxNumberOfTopLevelDescribes: math.MaxInt}
	if len(rawOptions) == 0 {
		return opts
	}

	optsMap, _ := rawOptions[0].(map[string]any)
	if maxValue, ok := internalUtils.CoerceIntegral(optsMap["maxNumberOfTopLevelDescribes"]); ok && maxValue >= 1 {
		opts.maxNumberOfTopLevelDescribes = maxValue
	}
	return opts
}

func tooManyDescribesMessage(maxAllowed int) rule.RuleMessage {
	plural := "s"
	if maxAllowed == 1 {
		plural = ""
	}
	return rule.RuleMessage{
		Id:          "tooManyDescribes",
		Description: fmt.Sprintf("There should not be more than %d describe%s at the top level", maxAllowed, plural),
		Data: map[string]string{
			"max": strconv.Itoa(maxAllowed),
			"s":   plural,
		},
	}
}

func unexpectedMessage(id string, description string) rule.RuleMessage {
	return rule.RuleMessage{Id: id, Description: description}
}

// NewRule creates a top-level-describe rule for a test framework. The framework
// adapter owns call provenance, suite containment and diagnostic wording; this
// engine owns the traversal stack and the top-level describe budget.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.NewSchema(schemaJSON),
		Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			if runtime.Parse == nil {
				return rule.RuleListeners{}
			}
			opts := parseOptions(rawOptions)
			describes := make([]describeFrame, 0, 4)
			topLevelDescribes := 0

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					parsed := runtime.Parse(node)
					if parsed == nil {
						return
					}

					if testFramework.IsCallOfKind(parsed, testFramework.FnKindDescribe) {
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
						if depth != 1 {
							return
						}
						topLevelDescribes++
						if topLevelDescribes > opts.maxNumberOfTopLevelDescribes {
							ctx.ReportNode(node, tooManyDescribesMessage(opts.maxNumberOfTopLevelDescribes))
						}
						return
					}

					if len(describes) > 0 ||
						!testFramework.IsCallOfKind(parsed, testFramework.FnKindTest, testFramework.FnKindHook) {
						return
					}
					if runtime.InsideDescribe != nil && runtime.InsideDescribe(node) {
						return
					}

					if parsed.Kind == testFramework.FnKindTest {
						ctx.ReportNode(node, unexpectedMessage("unexpectedTestCase", config.Messages.UnexpectedTestCase))
						return
					}
					ctx.ReportNode(node, unexpectedMessage("unexpectedHook", config.Messages.UnexpectedHook))
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
