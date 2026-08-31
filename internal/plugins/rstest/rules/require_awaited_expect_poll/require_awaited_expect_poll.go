package require_awaited_expect_poll

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func buildNotAwaitedMessage(method string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notAwaited",
		Description: fmt.Sprintf("`%s` calls should be awaited", method),
	}
}

var RequireAwaitedExpectPollRule = rule.Rule{
	Name:   "rstest/require-awaited-expect-poll",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseExpectCall(node)
				if !isAwaitedFactory(parsed) {
					return
				}

				// The parser resolves one chain at its outermost call, so a
				// chain reaches this listener exactly once and needs no
				// reported-node bookkeeping of its own. Upstream keeps a
				// `reported` set because it parses from every call node.
				if isHandled(topMostConsumedNode(parsed.Expression)) {
					return
				}

				_, accessor := testFramework.AccessorReceiverAndParent(&parsed.MemberEntries[0])
				if accessor == nil {
					return
				}

				// NOTE: Unlike ESLint, the message always names the factory as
				// `expect.poll` / `expect.element` even when the call site
				// spells the root differently (`import.meta.rstest.expect`, a
				// renamed import, a test-context expect). Upstream builds the
				// same fixed `expect.` prefix from its own template.
				ctx.ReportNode(accessor, buildNotAwaitedMessage("expect."+string(parsed.Entry)))
			},
		}
	},
}

// isAwaitedFactory reports chains produced by an assertion factory whose
// matchers resolve asynchronously: expect.poll(fn) polls until its matcher
// passes, and expect.element(locator) has none but async matchers.
func isAwaitedFactory(parsed *rstestUtils.ParsedRstestExpectCall) bool {
	if parsed == nil ||
		parsed.Expression == nil ||
		parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
		len(parsed.MemberEntries) == 0 {
		return false
	}
	return parsed.Entry == rstestUtils.RstestExpectEntryPoll ||
		parsed.Entry == rstestUtils.RstestExpectEntryElement
}

// topMostConsumedNode walks out of the assertion chain to the node whose
// parent decides whether the promise is handled.
//
// The chain itself is already collapsed by the parser: ParsedRstestExpectCall.
// Expression is the outermost matcher call or property access, which is what
// upstream's skipMatchersAndModifiers arrives at. What remains is upstream's
// skipSequenceExpressions: a comma expression evaluates to its last operand,
// so an assertion in final position is handled exactly when the comma
// expression is, while any earlier position discards the promise. Conditional
// branches and short-circuit operands similarly flow into the enclosing
// expression, so keep walking until reaching the position that consumes the
// resulting value.
func topMostConsumedNode(node *ast.Node) *ast.Node {
	current := ascendThroughWrappers(node)
	for {
		parent := current.Parent
		if parent == nil {
			return current
		}

		switch parent.Kind {
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken == nil || !binaryValueFlowsFrom(binary, current) {
				return current
			}
		case ast.KindConditionalExpression:
			conditional := parent.AsConditionalExpression()
			if conditional.WhenTrue != current && conditional.WhenFalse != current {
				return current
			}
		default:
			return current
		}

		current = ascendThroughWrappers(parent)
	}
}

// binaryValueFlowsFrom reports whether a short-circuit or comma expression
// can return current's promise as its own value. A right operand always becomes
// the result when evaluated. A promise on the left is also the result of `||`
// and `??` because promises are truthy and non-nullish, but `promise && other`
// evaluates to other and therefore drops the promise.
func binaryValueFlowsFrom(binary *ast.BinaryExpression, current *ast.Node) bool {
	operator := binary.OperatorToken.Kind
	if binary.Right == current {
		return operator == ast.KindCommaToken ||
			ast.IsLogicalOrCoalescingBinaryOperator(operator)
	}
	return binary.Left == current &&
		(operator == ast.KindBarBarToken || operator == ast.KindQuestionQuestionToken)
}

// ascendThroughWrappers returns the outermost node that wraps node without
// consuming its value. ESTree has no node for parentheses and rslint's
// upstream reference never sees a TypeScript assertion in this position, so
// walking past both is what keeps `await (expect.poll(x).toBe(1) as Promise<void>)`
// recognisable as awaited.
func ascendThroughWrappers(node *ast.Node) *ast.Node {
	for node.Parent != nil {
		switch node.Parent.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindAsExpression,
			ast.KindSatisfiesExpression,
			ast.KindNonNullExpression,
			ast.KindTypeAssertionExpression:
			node = node.Parent
		default:
			return node
		}
	}
	return node
}

// isHandled reports whether node's parent does something with the promise the
// assertion produced.
//
// NOTE: Unlike ESLint, which accepts only `await` and `return`, every position
// below hands the promise to something that can settle it — a concise arrow
// body returns it, an initializer, assignment or property binds it, a call
// or constructor argument passes it on (which is how `Promise.all([...])`
// and `Promise.allSettled([...])` are written), a JSX expression passes it
// as a prop or child, a destructuring or parameter default binds it when the
// default runs, and `yield` suspends on it.
// Reporting those would be a false positive; the cost is that a promise stored
// and then dropped goes unreported.
func isHandled(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindAwaitExpression, ast.KindReturnStatement:
		return true
	case ast.KindArrowFunction:
		return parent.AsArrowFunction().Body == node
	case ast.KindVariableDeclaration:
		return parent.AsVariableDeclaration().Initializer == node
	case ast.KindPropertyDeclaration:
		return parent.AsPropertyDeclaration().Initializer == node
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == node
	case ast.KindYieldExpression:
		return parent.AsYieldExpression().Expression == node
	case ast.KindArrayLiteralExpression:
		return true
	case ast.KindCallExpression:
		return parent.AsCallExpression().Expression != node
	case ast.KindNewExpression:
		return parent.AsNewExpression().Expression != node
	case ast.KindJsxExpression:
		jsxExpression := parent.AsJsxExpression()
		return jsxExpression.DotDotDotToken == nil && jsxExpression.Expression == node
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary.OperatorToken != nil &&
			ast.IsAssignmentOperator(binary.OperatorToken.Kind) &&
			binary.Right == node
	case ast.KindBindingElement:
		// A destructuring declaration's own default value, e.g. `const [p =
		// expect.poll(fn).toBe(1)] = values`. BindingElement is shared by
		// array and object patterns, nested patterns, and the bindings a
		// for-of/for-in declaration or a destructured parameter introduces,
		// so this one case covers all of them.
		return parent.AsBindingElement().Initializer == node
	case ast.KindParameter:
		// A plain or destructured parameter's default value, e.g. `function
		// run(p = expect.poll(fn).toBe(1)) {}`.
		return parent.AsParameterDeclaration().Initializer == node
	case ast.KindShorthandPropertyAssignment:
		// An object assignment pattern's shorthand default, e.g. `({p =
		// expect.poll(fn).toBe(1)} = values)`. Object literals can't write a
		// bare `key = value` property, so ts-go parses this default into its
		// own node instead of the BinaryExpression an array pattern's `[p =
		// ...]` produces.
		return parent.AsShorthandPropertyAssignment().ObjectAssignmentInitializer == node
	default:
		return false
	}
}
