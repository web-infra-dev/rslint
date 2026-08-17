// Package no_standalone_expect holds the framework-neutral traversal shared by
// jest/no-standalone-expect and rstest/no-standalone-expect.
//
// Upstream eslint-plugin-jest tracks the enclosing block with a stack of kinds
// pushed on enter and popped on exit. rslint's listeners cannot consult a scope
// object, so the stack stores source ranges instead and an assertion looks up
// the innermost range containing it.
//
// That substitution introduces a failure mode upstream does not have: a range
// may be unavailable. Every scope the exit handler pops must therefore be
// pushed on enter under the *same* predicate, falling back to a range derived
// from the node itself when the preferred one cannot be found. Guarding the
// push with an extra condition instead makes a nested registration without an
// inline callback — `test.todo("x")`, `test("x", cb)`, a custom block
// function — pop the enclosing test's scope, and every assertion after it in
// that body is reported as standalone.
//
// Keeping the stack balanced and choosing what the fallback scope covers are
// separate decisions: an entry whose range is empty still balances the stack
// while containing no assertion. Runtime.ArgumentsOutsideTestScope picks
// between the two, because the frameworks want different answers for an
// assertion handed to a registration as an argument.
//
// Known deliberate divergence from upstream: method, constructor and accessor
// bodies count as `function` scopes, so an assertion inside a class method is
// allowed here and reported upstream.
package no_standalone_expect

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type blockType string

const (
	blockTypeTest     blockType = "test"
	blockTypeFunction blockType = "function"
	blockTypeDescribe blockType = "describe"
	blockTypeArrow    blockType = "arrow"
	blockTypeTemplate blockType = "template"
)

// ExpectCallKind classifies a call expression for the traversal.
type ExpectCallKind int

const (
	// NotExpectCall is any call that is not an assertion entry point.
	NotExpectCall ExpectCallKind = iota
	// StaticExpectCall is an `expect.*` form that builds a value rather than
	// asserting, such as `expect.any(String)`. It is never reported, and it
	// never opens a scope.
	StaticExpectCall
	// AssertingExpectCall is an assertion that has to sit inside a test block.
	AssertingExpectCall
)

// Runtime supplies the framework-specific decisions the traversal needs.
type Runtime struct {
	// IsTestCall reports whether a CallExpression registers a test case.
	IsTestCall func(*ast.Node) bool
	// IsDescribeCall reports whether a CallExpression registers a suite.
	IsDescribeCall func(*ast.Node) bool
	// ClassifyExpectCall decides whether a CallExpression is an assertion, and
	// whether that assertion actually asserts.
	ClassifyExpectCall func(*ast.Node) ExpectCallKind
	// ArgumentsOutsideTestScope reports whether the arguments of a registration
	// sit outside the test case it registers. They are evaluated while the file
	// is collected rather than while the case runs, so an assertion among them
	// never runs as part of a test. Upstream eslint-plugin-jest allows it — its
	// kind stack covers the whole call subtree — and the jest binding follows
	// that contract; the rstest binding reports it.
	//
	// The flag only reaches a registration with no inline callback. Whenever one
	// is found the scope is its body either way, so an assertion in the name or
	// the options object is already outside it.
	ArgumentsOutsideTestScope bool
	// Skip turns the rule into a no-op for this file.
	Skip bool
}

// Config describes one framework's binding of the shared rule.
type Config struct {
	Name    string
	Schema  *rule.Schema
	Prepare func(rule.RuleContext) Runtime
}

type scopeEntry struct {
	kind  blockType
	start int
	end   int
}

func buildUnexpectedExpectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedExpect",
		Description: "Expect must be inside of a test block",
	}
}

func parseAdditionalTestBlockFunctions(rawOptions []any) map[string]struct{} {
	names := map[string]struct{}{}
	if len(rawOptions) == 0 {
		return names
	}

	optsMap, _ := rawOptions[0].(map[string]any)
	list, ok := optsMap["additionalTestBlockFunctions"].([]any)
	if !ok {
		return names
	}

	for _, item := range list {
		if name, ok := item.(string); ok {
			names[name] = struct{}{}
		}
	}

	return names
}

func getBlockType(block *ast.Node, isDescribeCall func(*ast.Node) bool) blockType {
	fn := block.Parent
	if !testFramework.IsFunction(fn) {
		return ""
	}

	if ast.IsFunctionDeclaration(fn) {
		return blockTypeFunction
	}

	if fn.Parent != nil && fn.Parent.Kind == ast.KindVariableDeclaration {
		return blockTypeFunction
	}

	switch fn.Kind {
	case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
		return blockTypeFunction
	}

	parent := fn.Parent
	if parent != nil && parent.Kind == ast.KindCallExpression &&
		isDescribeCall != nil && isDescribeCall(parent) {
		return blockTypeDescribe
	}

	return ""
}

func calleeIsTaggedTemplateExpression(node *ast.Node) bool {
	callExpr := node.AsCallExpression()
	return callExpr != nil &&
		callExpr.Expression != nil &&
		callExpr.Expression.Kind == ast.KindTaggedTemplateExpression
}

// isStandaloneArrow reports whether an arrow opens a scope of its own. An arrow
// passed straight to a call is the callback of whatever that call is, and the
// call's own handling decides what scope it opens.
func isStandaloneArrow(node *ast.Node) bool {
	return node.Parent != nil && node.Parent.Kind != ast.KindCallExpression
}

func getFunctionBodyRange(fn *ast.Node) (int, int, bool) {
	if fn == nil {
		return 0, 0, false
	}

	body := fn.Body()
	if body == nil {
		return 0, 0, false
	}

	return body.Pos(), body.End(), true
}

// getTestScopeRange returns the body of the callback a registration runs. The
// search goes back to front because the callback is the last argument in
// `(name, fn)` and `(name, options, fn)` alike, while `(name, fn, timeout)`
// ends in a value with no body.
func getTestScopeRange(node *ast.Node) (int, int, bool) {
	callExpr := node.AsCallExpression()
	if callExpr == nil || callExpr.Arguments == nil {
		return 0, 0, false
	}

	for i := len(callExpr.Arguments.Nodes) - 1; i >= 0; i-- {
		if start, end, ok := getFunctionBodyRange(callExpr.Arguments.Nodes[i]); ok {
			return start, end, true
		}
	}

	return 0, 0, false
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: config.Schema,
		Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			if runtime.Skip {
				return rule.RuleListeners{}
			}

			additionalTestBlockFunctions := parseAdditionalTestBlockFunctions(rawOptions)
			scopeStack := make([]scopeEntry, 0, 8)

			isCustomTestBlockFunction := func(node *ast.Node) bool {
				if len(additionalTestBlockFunctions) == 0 {
					return false
				}
				callExpr := node.AsCallExpression()
				if callExpr == nil {
					return false
				}
				name := testFramework.CalleeChainName(callExpr.Expression)
				if name == "" {
					return false
				}
				_, ok := additionalTestBlockFunctions[name]
				return ok
			}

			isTestBlockCall := func(node *ast.Node) bool {
				return (runtime.IsTestCall != nil && runtime.IsTestCall(node)) ||
					isCustomTestBlockFunction(node)
			}

			currentScopeKind := func() blockType {
				if len(scopeStack) == 0 {
					return ""
				}
				return scopeStack[len(scopeStack)-1].kind
			}

			pushScope := func(kind blockType, start, end int) {
				scopeStack = append(scopeStack, scopeEntry{kind: kind, start: start, end: end})
			}

			popScope := func(kind blockType) {
				if len(scopeStack) == 0 || scopeStack[len(scopeStack)-1].kind != kind {
					return
				}
				scopeStack = scopeStack[:len(scopeStack)-1]
			}

			getContainingBlock := func(node *ast.Node) blockType {
				pos := node.Pos()
				for i := len(scopeStack) - 1; i >= 0; i-- {
					scope := scopeStack[i]
					if scope.start <= pos && pos < scope.end {
						return scope.kind
					}
				}
				return ""
			}

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					if runtime.ClassifyExpectCall != nil {
						switch runtime.ClassifyExpectCall(node) {
						case StaticExpectCall:
							return
						case AssertingExpectCall:
							if parent := getContainingBlock(node); parent == "" || parent == blockTypeDescribe {
								ctx.ReportNode(node, buildUnexpectedExpectMessage())
							}
							return
						}
					}

					if isTestBlockCall(node) {
						start, end, ok := getTestScopeRange(node)
						if !ok {
							// A registration with no inline callback still has to
							// push, or the exit handler pops somebody else's scope.
							// What that scope covers is the policy choice: the whole
							// call takes its arguments in, an empty range leaves an
							// assertion among them with no test block around it.
							if runtime.ArgumentsOutsideTestScope {
								start, end = node.Pos(), node.Pos()
							} else {
								start, end = node.Pos(), node.End()
							}
						}
						pushScope(blockTypeTest, start, end)
					}

					if calleeIsTaggedTemplateExpression(node) {
						// The scope spans the whole call, not just the tagged
						// template that is its callee: a tagged template returns the
						// function that registers the cases, so the callback handed
						// to that call is the part an assertion can sit in.
						pushScope(blockTypeTemplate, node.Pos(), node.End())
					}
				},
				rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
					if calleeIsTaggedTemplateExpression(node) {
						popScope(blockTypeTemplate)
					}

					if isTestBlockCall(node) && currentScopeKind() == blockTypeTest {
						popScope(blockTypeTest)
					}
				},
				ast.KindBlock: func(node *ast.Node) {
					if kind := getBlockType(node, runtime.IsDescribeCall); kind != "" {
						pushScope(kind, node.Pos(), node.End())
					}
				},
				rule.ListenerOnExit(ast.KindBlock): func(node *ast.Node) {
					kind := getBlockType(node, runtime.IsDescribeCall)
					if kind != "" && currentScopeKind() == kind {
						popScope(kind)
					}
				},
				ast.KindArrowFunction: func(node *ast.Node) {
					if !isStandaloneArrow(node) {
						return
					}
					start, end, ok := getFunctionBodyRange(node)
					if !ok {
						start, end = node.Pos(), node.End()
					}
					pushScope(blockTypeArrow, start, end)
				},
				rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
					if isStandaloneArrow(node) && currentScopeKind() == blockTypeArrow {
						popScope(blockTypeArrow)
					}
				},
			}
		},
	}
}
