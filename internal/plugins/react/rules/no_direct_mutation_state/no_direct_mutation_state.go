package no_direct_mutation_state

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// inConstructorExemption mirrors eslint-plugin-react's
// `shouldIgnoreComponent` check: mutation is exempt when we are lexically
// inside the component's constructor AND NOT inside any nested
// CallExpression.
//
// Walks from `node.Parent` toward `component` (exclusive) looking for a
// KindConstructor ancestor and any KindCallExpression ancestor.
func inConstructorExemption(node, component *ast.Node) bool {
	inConstructor := false
	inCallExpression := false
	for p := node.Parent; p != nil && p != component; p = p.Parent {
		switch p.Kind {
		case ast.KindConstructor:
			inConstructor = true
		case ast.KindCallExpression:
			inCallExpression = true
		}
	}
	return inConstructor && !inCallExpression
}

// innermostMemberExpression mirrors eslint-plugin-react's
// `getOuterMemberExpression`, which walks `node.object` inward while it is
// still a member expression, returning the innermost member expression whose
// base is no longer a member expression.
//
// Both PropertyAccessExpression (`.foo`) and ElementAccessExpression
// (`['foo']`) are traversed, matching ESTree's unified `MemberExpression`.
// Parentheses are transparently skipped at each step (ESTree flattens them;
// tsgo preserves them).
func innermostMemberExpression(node *ast.Node) *ast.Node {
	current := node
	for {
		current = ast.SkipParentheses(current)
		var inner *ast.Node
		switch current.Kind {
		case ast.KindPropertyAccessExpression:
			inner = current.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			inner = current.AsElementAccessExpression().Expression
		default:
			return current
		}
		innerSkipped := ast.SkipParentheses(inner)
		switch innerSkipped.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			current = innerSkipped
		default:
			return current
		}
	}
}

// isThisStateMember mirrors eslint-plugin-react's
// `componentUtil.isStateMemberExpression`: matches `this.state` where the
// property is a bare Identifier named `state`. `this['state']` does NOT
// match (ESLint checks `node.property.name === 'state'`, which is undefined
// for non-Identifier properties).
func isThisStateMember(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	prop := node.AsPropertyAccessExpression()
	if ast.SkipParentheses(prop.Expression).Kind != ast.KindThisKeyword {
		return false
	}
	name := prop.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}
	return name.AsIdentifier().Text == "state"
}

var NoDirectMutationStateRule = rule.Rule{
	Name: "react/no-direct-mutation-state",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)

		report := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noDirectMutation",
				Description: "Do not mutate state directly. Use setState().",
			})
		}

		// check handles both assignments and update expressions. `target` is
		// the outer-most member expression being mutated (node.left for an
		// assignment; node.argument for ++/--). `reportNode` is the AST node
		// whose position the diagnostic should point at — matching ESLint:
		//
		//   - For assignments: `node.left.object` (i.e. the object of the
		//     assignment's LHS member expression, which for `this.state.x = ...`
		//     is `this.state` and for `this.state = ...` is `this`).
		//   - For updates: the innermost `this.state` member expression.
		//
		// Both reportNodes start at the same column as the overall chain, so
		// line/column assertions align with ESLint.
		check := func(node, target, reportNode *ast.Node) {
			target = ast.SkipParentheses(target)
			switch target.Kind {
			case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			default:
				return
			}
			innermost := innermostMemberExpression(target)
			if !isThisStateMember(innermost) {
				return
			}
			component := reactutil.GetEnclosingReactComponentOrStateless(node, pragma, createClass)
			if component == nil {
				return
			}
			if inConstructorExemption(node, component) {
				return
			}
			report(reportNode)
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				if bin.OperatorToken == nil || !ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
					return
				}
				left := ast.SkipParentheses(bin.Left)
				// Report position mirrors ESLint's `node.left.object` — the
				// direct base of the assignment's LHS (e.g. `this.state` for
				// `this.state.foo = ...`, or `this` for `this.state = ...`).
				var reportNode *ast.Node
				switch left.Kind {
				case ast.KindPropertyAccessExpression:
					reportNode = left.AsPropertyAccessExpression().Expression
				case ast.KindElementAccessExpression:
					reportNode = left.AsElementAccessExpression().Expression
				default:
					return
				}
				check(node, left, reportNode)
			},

			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				pf := node.AsPrefixUnaryExpression()
				if pf.Operator != ast.KindPlusPlusToken && pf.Operator != ast.KindMinusMinusToken {
					return
				}
				operand := ast.SkipParentheses(pf.Operand)
				// ESLint guards with `node.argument.type !== 'MemberExpression'`.
				switch operand.Kind {
				case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
				default:
					return
				}
				check(node, operand, innermostMemberExpression(operand))
			},

			ast.KindPostfixUnaryExpression: func(node *ast.Node) {
				pf := node.AsPostfixUnaryExpression()
				if pf.Operator != ast.KindPlusPlusToken && pf.Operator != ast.KindMinusMinusToken {
					return
				}
				operand := ast.SkipParentheses(pf.Operand)
				switch operand.Kind {
				case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
				default:
					return
				}
				check(node, operand, innermostMemberExpression(operand))
			},
		}
	},
}
