package dot_notation

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// This is the TypeScript source from typescript-eslint for reference:
// https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/src/rules/dot-notation.ts
//
// Rule: dot-notation
// Enforce dot notation whenever possible
//
// This rule extends the base JavaScript rule to support TypeScript-specific features:
// - allowPrivateClassPropertyAccess: Whether to allow accessing class members marked as `private` with array notation
// - allowProtectedClassPropertyAccess: Whether to allow accessing class members marked as `protected` with array notation
// - allowIndexSignaturePropertyAccess: Whether to allow accessing properties matching an index signature with array notation
//
// The base rule options are also supported:
// - allowKeywords: Whether to allow keywords such as ["class"]
// - allowPattern: Regular expression of names to allow
//
// Message IDs from base rule:
// - useDot: "[{{key}}]" is better written in dot notation.
// - useBrackets: .{{key}} is a syntax error.

var DotNotationRule = rule.Rule{
	Name: "dot-notation",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// TODO: Implement the rule logic
		// This rule needs to:
		// 1. Check MemberExpression nodes where computed=true (bracket notation)
		// 2. Determine if the property could be accessed with dot notation instead
		// 3. Consider TypeScript-specific options for private/protected members and index signatures
		// 4. Report violations with appropriate message and fix
		
		// Default options:
		// allowIndexSignaturePropertyAccess: false
		// allowKeywords: true
		// allowPattern: ""
		// allowPrivateClassPropertyAccess: false
		// allowProtectedClassPropertyAccess: false
		
		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				// TODO: Implement property access checking
				// Check if bracket notation could be replaced with dot notation
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				// TODO: Implement element access checking
				// This is the TypeScript AST equivalent of computed MemberExpression
			},
		}
	},
}