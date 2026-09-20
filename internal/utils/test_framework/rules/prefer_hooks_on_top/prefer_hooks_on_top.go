package prefer_hooks_on_top

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type ParsedCall = testFramework.ParsedCall

type Runtime struct {
	Parse func(*ast.Node) *testFramework.ParsedCall
}

type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
}

func buildNoHookOnTopMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noHookOnTop",
		Description: "Hooks should come before test cases",
	}
}

// NewRule reports setup/teardown hooks declared after the first test case of
// the same scope. A hook is attached to the whole enclosing suite wherever it
// is written, so a hook that follows a test case reads as if it applied only to
// what comes after it while it in fact also wraps every test above it.
//
// A scope here is bounded by both call expressions and function bodies:
// entering either pushes a frame and leaving it pops that frame, so the first
// test case of a describe body does not make hooks in a sibling describe, or in
// a nested callback, report. Frames are pushed unconditionally, which keeps
// enter and exit symmetric even when a call parses as nothing at all.
//
// Function bodies need a frame of their own because a function is not the suite
// that lexically encloses it. A hook written in a function that is never called
// registers nothing at all, and a function passed to `describe` by name is the
// body of that suite rather than a continuation of the file around it, so
// neither may inherit the test cases already registered outside. The cost is
// that a hook registered through a helper the suite does in fact call is judged
// against the helper instead of the suite, and so goes unreported: this rule
// answers how a file reads top to bottom, where a missed report is a smaller
// failure than one that cannot be acted on.
//
// Whether a call registers a test case is the framework parser's decision, not
// this engine's: an API factory such as an Rstest `test.extend({...})` chain
// that only builds another test function parses as no test-framework call, so
// it never opens the scope, while the registration it later produces does.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)

			// hooksContext[len-1] records whether the scope currently being
			// visited has already registered a test case.
			hooksContext := []bool{false}

			pushScope := func(node *ast.Node) {
				hooksContext = append(hooksContext, false)
			}
			popScope := func(node *ast.Node) {
				// Enter pushes for every scope node, so the file frame can only
				// be reached by an exit without its enter. Keep it rather than
				// underflow the stack.
				if len(hooksContext) == 1 {
					return
				}
				hooksContext = hooksContext[:len(hooksContext)-1]
			}

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					parsed := runtime.Parse(node)
					top := len(hooksContext) - 1

					if testFramework.IsCallOfKind(parsed, testFramework.FnKindTest) {
						hooksContext[top] = true
					}

					if hooksContext[top] && testFramework.IsCallOfKind(parsed, testFramework.FnKindHook) {
						ctx.ReportNode(node, buildNoHookOnTopMessage())
					}

					pushScope(node)
				},
				rule.ListenerOnExit(ast.KindCallExpression): popScope,

				ast.KindFunctionDeclaration:                      pushScope,
				rule.ListenerOnExit(ast.KindFunctionDeclaration): popScope,
				ast.KindFunctionExpression:                       pushScope,
				rule.ListenerOnExit(ast.KindFunctionExpression):  popScope,
				ast.KindArrowFunction:                            pushScope,
				rule.ListenerOnExit(ast.KindArrowFunction):       popScope,
				ast.KindMethodDeclaration:                        pushScope,
				rule.ListenerOnExit(ast.KindMethodDeclaration):   popScope,
			}
		},
	}
}
