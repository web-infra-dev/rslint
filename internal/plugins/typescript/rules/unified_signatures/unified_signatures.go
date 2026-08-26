package unified_signatures

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed unified_signatures.schema.json
var schemaJSON []byte

type UnifiedSignaturesOptions struct {
	IgnoreDifferentlyNamedParameters  bool `json:"ignoreDifferentlyNamedParameters"`
	IgnoreOverloadsWithDifferentJSDoc bool `json:"ignoreOverloadsWithDifferentJSDoc"`
}

var UnifiedSignaturesRule = rule.CreateRule(rule.Rule{
	Name:   "unified-signatures",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := UnifiedSignaturesOptions{
			IgnoreDifferentlyNamedParameters:  false,
			IgnoreOverloadsWithDifferentJSDoc: false,
		}

		if len(options) > 0 {
			optsMap, _ := options[0].(map[string]interface{})
			if val, ok := optsMap["ignoreDifferentlyNamedParameters"].(bool); ok {
				opts.IgnoreDifferentlyNamedParameters = val
			}
			if val, ok := optsMap["ignoreOverloadsWithDifferentJSDoc"].(bool); ok {
				opts.IgnoreOverloadsWithDifferentJSDoc = val
			}
		}

		// TODO: Implement full unified-signatures logic
		// This rule checks for function overloads that could be unified into a single signature
		//
		// The implementation should:
		// 1. Track consecutive function/method/constructor overloads
		// 2. Compare signatures to see if they can be unified using:
		//    - Optional parameters (? operator)
		//    - Union types (|)
		//    - Rest parameters (...)
		// 3. Check for differences in:
		//    - Parameter types
		//    - Return types
		//    - Parameter names (if ignoreDifferentlyNamedParameters is false)
		//    - JSDoc comments (if ignoreOverloadsWithDifferentJSDoc is false)
		// 4. Report when overloads can be unified
		//
		// Example:
		//   function x(x: number): void;    <- Can be unified
		//   function x(x: string): void;    <- into: function x(x: number | string): void;
		//
		// This is a complex rule that requires analyzing multiple consecutive overloads
		// and determining whether they can be simplified

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				// TODO: Implement function overload checking
				// Check if this function has overloads and analyze them
			},
			ast.KindMethodDeclaration: func(node *ast.Node) {
				// TODO: Implement method overload checking
				// Check if this method has overloads and analyze them
			},
			ast.KindConstructorType: func(node *ast.Node) {
				// TODO: Implement constructor overload checking
				// Check if this constructor has overloads and analyze them
			},
		}
	},
})
