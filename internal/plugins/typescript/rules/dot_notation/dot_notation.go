package dot_notation

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"

	"github.com/web-infra-dev/rslint/internal/rule"
	coreDotNotation "github.com/web-infra-dev/rslint/internal/rules/dot_notation"
)

//go:embed dot_notation.schema.json
var schemaJSON []byte

// Options mirrors @typescript-eslint/dot-notation options.
type Options struct {
	AllowIndexSignaturePropertyAccess bool   `json:"allowIndexSignaturePropertyAccess"`
	AllowKeywords                     bool   `json:"allowKeywords"`
	AllowPattern                      string `json:"allowPattern"`
	AllowPrivateClassPropertyAccess   bool   `json:"allowPrivateClassPropertyAccess"`
	AllowProtectedClassPropertyAccess bool   `json:"allowProtectedClassPropertyAccess"`
}

func parseOptions(options []any) Options {
	opts := Options{
		AllowKeywords:                     true,
		AllowIndexSignaturePropertyAccess: false,
	}

	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})

	if v, ok := optsMap["allowIndexSignaturePropertyAccess"].(bool); ok {
		opts.AllowIndexSignaturePropertyAccess = v
	}
	if v, ok := optsMap["allowKeywords"].(bool); ok {
		opts.AllowKeywords = v
	}
	if v, ok := optsMap["allowPattern"].(string); ok {
		opts.AllowPattern = v
	}
	if v, ok := optsMap["allowPrivateClassPropertyAccess"].(bool); ok {
		opts.AllowPrivateClassPropertyAccess = v
	}
	if v, ok := optsMap["allowProtectedClassPropertyAccess"].(bool); ok {
		opts.AllowProtectedClassPropertyAccess = v
	}
	return opts
}

// hasStringLikeIndexSignature reports whether the type exposes an index
// signature whose key is string-like. The checker handles mapped types,
// template-literal keys, unions, and intersections with TypeScript semantics.
func hasStringLikeIndexSignature(tc *checker.Checker, t *checker.Type) bool {
	if t == nil || tc == nil {
		return false
	}
	for _, info := range tc.GetIndexInfosOfType(t) {
		if info != nil && info.KeyType() != nil && info.KeyType().Flags()&checker.TypeFlagsStringLike != 0 {
			return true
		}
	}
	return false
}

// DotNotationRule extends ESLint core's dot-notation rule with TypeScript's
// private, protected, and index-signature allowances. Delegating base
// diagnostics and fixes keeps messages, ranges, and edit behavior aligned with
// the exact core rule that typescript-eslint extends upstream.
var DotNotationRule = rule.CreateRule(rule.Rule{
	Name:             "dot-notation",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		baseListeners := coreDotNotation.DotNotationRule.Run(ctx, options)

		allowIndexAccess := opts.AllowIndexSignaturePropertyAccess
		if program := ctx.Program(); program != nil {
			if compilerOptions := program.Options(); compilerOptions != nil && compilerOptions.NoPropertyAccessFromIndexSignature.IsTrue() {
				allowIndexAccess = true
			}
		}

		hasTypeAwareAllowance := opts.AllowPrivateClassPropertyAccess ||
			opts.AllowProtectedClassPropertyAccess ||
			allowIndexAccess
		if !hasTypeAwareAllowance {
			// The core property-access listener cannot report when keywords are
			// allowed. Omitting it avoids a callback for every ordinary dot access.
			if opts.AllowKeywords {
				delete(baseListeners, ast.KindPropertyAccessExpression)
			}
			return baseListeners
		}

		baseElementListener := baseListeners[ast.KindElementAccessExpression]
		listeners := rule.RuleListeners{
			ast.KindElementAccessExpression: func(node *ast.Node) {
				elem := node.AsElementAccessExpression()
				if elem == nil || elem.ArgumentExpression == nil {
					return
				}

				tc := ctx.TypeChecker
				propertySymbol := tc.GetSymbolAtLocation(elem.ArgumentExpression)
				var objectType *checker.Type
				if propertySymbol == nil {
					objectType = tc.GetNonNullableType(tc.GetTypeAtLocation(elem.Expression))
					keyNode := ast.SkipParentheses(elem.ArgumentExpression)
					if ast.IsStringLiteral(keyNode) {
						propertyName := keyNode.Text()
						apparentType := checker.Checker_getApparentType(tc, objectType)
						propertySymbol = checker.Checker_getPropertyOfType(tc, apparentType, propertyName)
						if propertySymbol == nil {
							for _, candidate := range checker.Checker_getPropertiesOfType(tc, apparentType) {
								if candidate != nil && candidate.Name == propertyName {
									propertySymbol = candidate
									break
								}
							}
						}
					}
				}

				if propertySymbol != nil {
					flags := checker.GetDeclarationModifierFlagsFromSymbol(propertySymbol)
					if opts.AllowPrivateClassPropertyAccess && flags&ast.ModifierFlagsPrivate != 0 {
						return
					}
					if opts.AllowProtectedClassPropertyAccess && flags&ast.ModifierFlagsProtected != 0 {
						return
					}
				} else if allowIndexAccess {
					if objectType == nil {
						objectType = tc.GetNonNullableType(tc.GetTypeAtLocation(elem.Expression))
					}
					if hasStringLikeIndexSignature(tc, objectType) {
						return
					}
				}

				baseElementListener(node)
			},
		}

		if !opts.AllowKeywords {
			listeners[ast.KindPropertyAccessExpression] = baseListeners[ast.KindPropertyAccessExpression]
		}
		return listeners
	},
})
