package no_global_assign

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_global_assign.schema.json
var schemaJSON []byte

type options struct {
	exceptions map[string]bool
}

func parseOptions(rawOptions []any) options {
	var opts options
	if len(rawOptions) == 0 {
		return opts
	}
	m, _ := rawOptions[0].(map[string]any)
	exceptions, _ := m["exceptions"].([]any)
	for _, e := range exceptions {
		if s, ok := e.(string); ok {
			// Allocate only once a real exception shows up, so files that
			// configure none never pay for the map.
			if opts.exceptions == nil {
				opts.exceptions = make(map[string]bool, len(exceptions))
			}
			opts.exceptions[s] = true
		}
	}
	return opts
}

// stackedTypeWrapperWriteExpression returns the direct assignment or update
// expression reached through more than one TypeScript target wrapper. ESLint's
// scope analysis unwraps one AsExpression, TypeAssertionExpression, or
// NonNullExpression from direct targets; parentheses do not count because
// ESTree does not preserve them.
func stackedTypeWrapperWriteExpression(node *ast.Node) *ast.Node {
	wrappers := 0
	current := node
	for parent := current.Parent; parent != nil; parent = current.Parent {
		switch parent.Kind {
		case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression:
			wrappers++
		case ast.KindParenthesizedExpression:
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary == nil || binary.OperatorToken == nil || binary.Left != current {
				return nil
			}
			if wrappers > 1 && ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				return parent
			}
			return nil
		case ast.KindPrefixUnaryExpression:
			prefix := parent.AsPrefixUnaryExpression()
			if wrappers > 1 && prefix != nil && prefix.Operand == current &&
				(prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken) {
				return parent
			}
			return nil
		case ast.KindPostfixUnaryExpression:
			postfix := parent.AsPostfixUnaryExpression()
			if wrappers > 1 && postfix != nil && postfix.Operand == current &&
				(postfix.Operator == ast.KindPlusPlusToken || postfix.Operator == ast.KindMinusMinusToken) {
				return parent
			}
			return nil
		default:
			return nil
		}
		current = parent
	}
	return nil
}

// isVisitedByEnclosingWritePattern mirrors the child edges followed by
// typescript-eslint's PatternVisitor. A stacked direct write is still a write
// when an enclosing destructuring pattern visits it. Right-hand values,
// computed keys, member expressions, and call arguments are evaluation-only
// boundaries; a call callee and nodes handled by PatternVisitor's default
// visitor remain reachable.
func isVisitedByEnclosingWritePattern(node *ast.Node) bool {
	current := node
	for parent := current.Parent; parent != nil; parent = current.Parent {
		switch parent.Kind {
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary == nil || binary.OperatorToken == nil {
				return false
			}
			if binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Left == current {
				// Do not skip parentheses: typescript-eslint converts a parenthesized
				// array or object on the left to an expression, not a pattern.
				if current.Kind == ast.KindArrayLiteralExpression || current.Kind == ast.KindObjectLiteralExpression {
					return true
				}
			}
			if ast.IsAssignmentOperator(binary.OperatorToken.Kind) && binary.Left != current {
				return false
			}

		case ast.KindForInStatement, ast.KindForOfStatement:
			statement := parent.AsForInOrOfStatement()
			if statement == nil || statement.Initializer != current {
				return false
			}
			return current.Kind == ast.KindArrayLiteralExpression || current.Kind == ast.KindObjectLiteralExpression

		case ast.KindParenthesizedExpression:
			paren := parent.AsParenthesizedExpression()
			if paren == nil || paren.Expression != current {
				return false
			}

		case ast.KindAsExpression:
			asExpression := parent.AsAsExpression()
			if asExpression == nil || asExpression.Expression != current {
				return false
			}

		case ast.KindTypeAssertionExpression:
			assertion := parent.AsTypeAssertion()
			if assertion == nil || assertion.Expression != current {
				return false
			}

		case ast.KindNonNullExpression:
			nonNull := parent.AsNonNullExpression()
			if nonNull == nil || nonNull.Expression != current {
				return false
			}

		case ast.KindSatisfiesExpression:
			satisfies := parent.AsSatisfiesExpression()
			if satisfies == nil || satisfies.Expression != current {
				return false
			}

		case ast.KindCallExpression:
			call := parent.AsCallExpression()
			if call == nil || call.Expression != current {
				return false
			}

		case ast.KindPropertyAssignment:
			property := parent.AsPropertyAssignment()
			if property == nil || property.Initializer != current {
				return false
			}

		case ast.KindShorthandPropertyAssignment:
			shorthand := parent.AsShorthandPropertyAssignment()
			if shorthand == nil || shorthand.Name() != current {
				return false
			}

		case ast.KindPropertyAccessExpression,
			ast.KindElementAccessExpression,
			ast.KindComputedPropertyName,
			ast.KindDecorator:
			return false
		}
		current = parent
	}
	return false
}

func buildGlobalShouldNotBeModifiedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "globalShouldNotBeModified",
		Description: "Read-only global '" + name + "' should not be modified.",
	}
}

// isReadonlyGlobal reports whether name resolves to a global variable that may
// not be assigned to: one the config or a `/* global */` comment declares
// readonly, or an ECMAScript built-in, which is readonly unless one of those
// two sources says otherwise. `writable` lifts the restriction and `off`
// removes the global entirely, so neither reports.
func isReadonlyGlobal(ctx rule.RuleContext, name string) bool {
	return ctx.Globals.Access(name) == utils.GlobalAccessReadonly
}

// NoGlobalAssignRule disallows assignments to native objects or read-only global variables
var NoGlobalAssignRule = rule.Rule{
	Name:   "no-global-assign",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		// A single-entry cache covers the common case of repeated writes to one
		// global without allocating maps for files that report once.
		cache := struct {
			name         string
			globalSymbol *ast.Symbol
			message      rule.RuleMessage
			messageBuilt bool
		}{}

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				name := node.Text()
				if opts.exceptions[name] || !isReadonlyGlobal(ctx, name) {
					return
				}

				if !utils.IsWriteReference(node) {
					return
				}

				if write := stackedTypeWrapperWriteExpression(node); write != nil &&
					!isVisitedByEnclosingWritePattern(write) {
					return
				}

				// A real lexical declaration or a resolved implicit binding (notably
				// CommonJS's wrapper-local `arguments`) shadows an effective global
				// of the same name and is not governed by its readonly setting.
				if ctx.Refs != nil && ctx.Refs.IsDefinedInFile(node) {
					return
				}

				if cache.name != name {
					cache.name = name
					cache.messageBuilt = false
					if ctx.Refs != nil && ctx.TypeChecker != nil {
						cache.globalSymbol = ctx.TypeChecker.GetGlobalSymbol(name, ast.SymbolFlagsValue, nil)
					}
				}

				if ctx.Refs == nil || ctx.TypeChecker == nil {
					if utils.IsShadowed(node, name) {
						return
					}
				} else {
					if cache.globalSymbol == nil {
						if utils.IsShadowed(node, name) {
							return
						}
					} else {
						// Resolve uses the binder's scope walk first, so a local shadow
						// produces its own symbol; only an unshadowed reference falls back
						// to the checker global cached above.
						if ctx.Refs.Resolve(node) != cache.globalSymbol {
							return
						}
					}
				}

				if !cache.messageBuilt {
					cache.message = buildGlobalShouldNotBeModifiedMessage(name)
					cache.messageBuilt = true
				}
				ctx.ReportNode(node, cache.message)
			},
		}
	},
}
