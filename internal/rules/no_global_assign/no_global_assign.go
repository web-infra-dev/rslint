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

// hasStackedTypeWrappers reports whether the identifier reaches a direct
// assignment or update target through more than one TypeScript expression
// wrapper. ESLint's scope analysis unwraps such a target exactly once — an
// AsExpression, TypeAssertionExpression or NonNullExpression — so
// `(Object as any) = 1` is a write while `((Object as any) as any) = 1` is not.
// Parentheses never count because the ESTree AST has no node for them.
// Destructuring defaults are AssignmentPattern nodes in ESTree rather than
// direct assignment targets, so their inner `=` must not stop this walk. A
// `satisfies` target is not a write at any depth, which IsWriteReference already
// answers.
func hasStackedTypeWrappers(node *ast.Node) bool {
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
				return false
			}
			if utils.IsDefaultValueInDestructuringAssignment(parent) {
				return false
			}
			return wrappers > 1 && ast.IsAssignmentOperator(binary.OperatorToken.Kind)
		case ast.KindPrefixUnaryExpression:
			prefix := parent.AsPrefixUnaryExpression()
			return wrappers > 1 && prefix != nil && prefix.Operand == current &&
				(prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken)
		case ast.KindPostfixUnaryExpression:
			postfix := parent.AsPostfixUnaryExpression()
			return wrappers > 1 && postfix != nil && postfix.Operand == current &&
				(postfix.Operator == ast.KindPlusPlusToken || postfix.Operator == ast.KindMinusMinusToken)
		default:
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

				if hasStackedTypeWrappers(node) {
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
