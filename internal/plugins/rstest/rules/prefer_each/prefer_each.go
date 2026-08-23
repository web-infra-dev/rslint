package prefer_each

import (
	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func buildPreferEachMessage(fn string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferEach",
		Description: "prefer using `" + fn + ".each` rather than a manual loop",
		Data: map[string]string{
			"fn": fn,
		},
	}
}

type pendingRegistration struct {
	kind rstestUtils.RstestFnType
	name string
}

type loopFrame struct {
	node          *ast.Node
	registrations []pendingRegistration
}

func isInsideForInOrOfExpression(node *ast.Node, loop *ast.Node) bool {
	if loop.Kind != ast.KindForInStatement && loop.Kind != ast.KindForOfStatement {
		return false
	}
	statement := loop.AsForInOrOfStatement()
	if statement == nil || statement.Expression == nil {
		return false
	}
	for current := node; current != nil && current != loop; current = current.Parent {
		if current == statement.Expression {
			return true
		}
	}
	return false
}

func recommendFn(pending []pendingRegistration) string {
	if len(pending) == 1 && pending[0].kind == rstestUtils.RstestFnTypeTest {
		if pending[0].name == "it" {
			return "it"
		}
		return "test"
	}
	return "describe"
}

var PreferEachRule = rule.Rule{
	Name:   "rstest/prefer-each",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		// NOTE: eslint-plugin-jest keeps one flat list of registrations for the
		// whole file and decides what to do with it from an `inTestCaseCall`
		// boolean that is set by every `test(...)` call and cleared by every
		// `test(...)` exit. Both halves of that leak across scopes: registrations
		// made before a loop can survive into the loop's report, a nested test
		// clears the flag while the outer test is still open, and a loop that
		// runs while the flag happens to be set is skipped entirely.
		//
		// This rule gives each loop its own frame instead. A registration is
		// recorded against the innermost loop that lexically contains it, and a
		// loop is reported from its own frame alone, so what the report says is
		// exactly what the loop registers. A loop that only runs business logic
		// has an empty frame and is never reported, whether or not it sits inside
		// a test callback.
		frames := make([]loopFrame, 0, 4)

		enterLoop := func(node *ast.Node) {
			frames = append(frames, loopFrame{node: node})
		}

		exitLoop := func(node *ast.Node) {
			if len(frames) == 0 {
				return
			}
			frame := frames[len(frames)-1]
			frames = frames[:len(frames)-1]
			if len(frame.registrations) == 0 {
				return
			}
			ctx.ReportNode(node, buildPreferEachMessage(recommendFn(frame.registrations)))
		}

		return rule.RuleListeners{
			ast.KindForStatement:                        enterLoop,
			ast.KindForInStatement:                      enterLoop,
			ast.KindForOfStatement:                      enterLoop,
			rule.ListenerOnExit(ast.KindForStatement):   exitLoop,
			rule.ListenerOnExit(ast.KindForInStatement): exitLoop,
			rule.ListenerOnExit(ast.KindForOfStatement): exitLoop,
			ast.KindCallExpression: func(node *ast.Node) {
				if len(frames) == 0 {
					return
				}
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return
				}
				switch parsed.Kind {
				case testFramework.FnKindTest,
					testFramework.FnKindDescribe,
					testFramework.FnKindHook:
					for i := len(frames) - 1; i >= 0; i-- {
						// The right-hand expression of for-in/of runs once before that
						// loop starts. It belongs to an enclosing loop, if any, rather
						// than to the loop whose rows it produces.
						if isInsideForInOrOfExpression(node, frames[i].node) {
							continue
						}
						frames[i].registrations = append(frames[i].registrations, pendingRegistration{
							kind: parsed.Kind,
							name: parsed.Name,
						})
						break
					}
				}
			},
		}
	},
}
