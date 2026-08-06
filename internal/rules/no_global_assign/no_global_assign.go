package no_global_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type options struct {
	exceptions map[string]bool
}

func parseOptions(opts any) options {
	var result options
	optsMap := utils.GetOptionsMap(opts)
	if optsMap != nil {
		if exceptions, ok := optsMap["exceptions"].([]interface{}); ok {
			for _, e := range exceptions {
				if s, ok := e.(string); ok {
					if result.exceptions == nil {
						result.exceptions = make(map[string]bool, len(exceptions))
					}
					result.exceptions[s] = true
				}
			}
		}
	}
	return result
}

// isWriteThroughTypeAssertion checks if the identifier reaches its assignment target
// through an AsExpression or TypeAssertionExpression. ESLint's scope analysis does not
// track writes through these TS-specific wrappers, so we skip them to match ESLint.
func isWriteThroughTypeAssertion(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression:
			return true
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
			current = current.Parent
			continue
		default:
			return false
		}
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
	switch ctx.Globals[name] {
	case utils.GlobalAccessReadonly:
		return true
	case utils.GlobalAccessUnset:
		return utils.IsECMAScriptGlobal(name)
	default:
		return false
	}
}

// NoGlobalAssignRule disallows assignments to native objects or read-only global variables
var NoGlobalAssignRule = rule.Rule{
	Name: "no-global-assign",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		opts := parseOptions(options)
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

				if isWriteThroughTypeAssertion(node) {
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
