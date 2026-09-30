package prefer_each

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type Runtime struct {
	Parse func(*ast.Node) *testFramework.ParsedCall
}

type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
	// SingleTestFn names the API recommended when a loop registers exactly one
	// test, given the parsed call's canonical Name. Frameworks differ in which
	// spellings exist (Jest exposes `it` and `test` everywhere; Rstest's
	// Playwright entry has no `it`). Multi-registration loops always recommend
	// `describe`. A nil SingleTestFn recommends `test`.
	SingleTestFn func(name string) string
}

func buildPreferEachMessage(fn string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferEach",
		Description: "prefer using `" + fn + ".each` rather than a manual loop",
		Data: map[string]string{
			"fn": fn,
		},
	}
}

type registration struct {
	kind testFramework.FnKind
	name string
}

type loopFrame struct {
	node          *ast.Node
	registrations []registration
}

func loopBody(loop *ast.Node) *ast.Node {
	switch loop.Kind {
	case ast.KindForStatement:
		if statement := loop.AsForStatement(); statement != nil {
			return statement.Statement
		}
	case ast.KindForInStatement, ast.KindForOfStatement:
		if statement := loop.AsForInOrOfStatement(); statement != nil {
			return statement.Statement
		}
	}
	return nil
}

func isInsideLoopBody(node *ast.Node, loop *ast.Node) bool {
	body := loopBody(loop)
	if body == nil {
		return false
	}
	for current := node; current != nil && current != loop; current = current.Parent {
		if current == body {
			return true
		}
	}
	return false
}

// NOTE: eslint-plugin-jest keeps one flat list of registrations for the whole
// file and decides what to do with it from an `inTestCaseCall` boolean that is
// set by every `test(...)` call and cleared by every `test(...)` exit. Both
// halves of that leak across scopes: registrations made before a loop can
// survive into the loop's report, a nested test clears the flag while the outer
// test is still open, a loop that runs while the flag happens to be set is
// skipped entirely, and an inner business loop discards the registrations its
// enclosing loop already made.
//
// This body gives each loop its own frame instead. A registration is recorded
// against the innermost loop whose body contains it, and a loop is reported
// from its own frame alone, so what the report says is exactly what the loop
// registers. A loop that only runs business logic has an empty frame and is
// never reported, whether or not it sits inside a test callback.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			recommend := func(pending []registration) string {
				if len(pending) == 1 && pending[0].kind == testFramework.FnKindTest {
					if config.SingleTestFn != nil {
						return config.SingleTestFn(pending[0].name)
					}
					return "test"
				}
				return "describe"
			}

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
				ctx.ReportNode(node, buildPreferEachMessage(recommend(frame.registrations)))
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
					parsed := runtime.Parse(node)
					if !testFramework.IsCallOfKind(parsed,
						testFramework.FnKindTest,
						testFramework.FnKindDescribe,
						testFramework.FnKindHook,
					) {
						return
					}
					for i := len(frames) - 1; i >= 0; i-- {
						// Only registrations in a loop's body repeat as part of that
						// loop. Calls in for-in/of iterables and classic for control
						// clauses belong to an enclosing loop, if any.
						if !isInsideLoopBody(node, frames[i].node) {
							continue
						}
						frames[i].registrations = append(frames[i].registrations, registration{
							kind: parsed.Kind,
							name: parsed.Name,
						})
						break
					}
				},
			}
		},
	}
}
