package no_invalid_void_type

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoInvalidVoidTypeOptions struct {
	AllowInGenericTypeArguments interface{} `json:"allowInGenericTypeArguments"`
	AllowAsThisParameter        bool        `json:"allowAsThisParameter"`
}

var NoInvalidVoidTypeRule = rule.CreateRule(rule.Rule{
	Name: "no-invalid-void-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoInvalidVoidTypeOptions{
			AllowInGenericTypeArguments: true,
			AllowAsThisParameter:        false,
		}

		// Parse options with dual-format support
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			// Handle array format: [{ option: value }]
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				// Handle direct object format: { option: value }
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if allowInGenericTypeArguments, ok := optsMap["allowInGenericTypeArguments"]; ok {
					opts.AllowInGenericTypeArguments = allowInGenericTypeArguments
				}
				if allowAsThisParameter, ok := optsMap["allowAsThisParameter"].(bool); ok {
					opts.AllowAsThisParameter = allowAsThisParameter
				}
			}
		}

		// Helper to check if void is allowed in generic context
		isAllowedInGeneric := func() bool {
			// If allowInGenericTypeArguments is false, never allow
			if allow, ok := opts.AllowInGenericTypeArguments.(bool); ok && !allow {
				return false
			}

			// If it's true (default), allow all generics
			if allow, ok := opts.AllowInGenericTypeArguments.(bool); ok && allow {
				return true
			}

			// If it's an array/whitelist, this would need more complex checking
			// For now, simplified implementation
			return true
		}

		// Helper to check if union contains only void and never
		isVoidNeverUnion := func(node *ast.Node) bool {
			if node.Kind != ast.KindUnionType {
				return false
			}
			union := node.AsUnionTypeNode()
			if union == nil || union.Types == nil {
				return false
			}

			hasVoid := false
			hasNever := false

			for _, typeNode := range union.Types.Nodes {
				if typeNode.Kind == ast.KindVoidKeyword {
					hasVoid = true
				} else if typeNode.Kind == ast.KindNeverKeyword {
					hasNever = true
				} else {
					// Has other types, not a void | never union
					return false
				}
			}

			return hasVoid && hasNever
		}

		// Helper to check if node is in valid context
		isValidVoidContext := func(node *ast.Node) bool {
			current := node.Parent

			// Walk up the tree to understand context
			for current != nil {
				switch current.Kind {
				// Allow in function return types and function type nodes
				case ast.KindFunctionType, ast.KindConstructorType:
					// For function types, we're valid if we're the return type
					// The void is valid in function type signatures
					return true

				case ast.KindFunctionDeclaration, ast.KindMethodDeclaration,
					ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindMethodSignature:
					// For actual function/method declarations, void is valid as return type
					return true

				// Allow in this parameter if option is enabled
				case ast.KindParameter:
					if opts.AllowAsThisParameter {
						param := current.AsParameterDeclaration()
						if param != nil && param.Name() != nil {
							if id := param.Name().AsIdentifier(); id != nil && id.Text == "this" {
								return true
							}
						}
					}
					// If not a 'this' parameter or option not enabled, continue checking
					return false

				// Allow in generic type arguments (Promise<void>, etc.)
				case ast.KindTypeReference, ast.KindNewExpression, ast.KindCallExpression:
					if isAllowedInGeneric() {
						return true
					}
					// If not allowed in generic, continue checking parent context
					current = current.Parent
					continue

				// Allow in union types if it's void | never
				case ast.KindUnionType:
					if isVoidNeverUnion(current) {
						// void | never is allowed
						return true
					}
					// Continue checking parent to see if this union is in a valid context
					current = current.Parent
					continue

				// For type aliases, continue checking
				case ast.KindTypeAliasDeclaration:
					// Continue to check what kind of type alias this is
					current = current.Parent
					continue
				}

				// Move up the tree
				current = current.Parent
			}

			return false
		}

		return rule.RuleListeners{
			ast.KindVoidKeyword: func(node *ast.Node) {
				// Check if in valid context
				if isValidVoidContext(node) {
					return
				}

				// Report invalid void usage
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "invalidVoidNotReturn",
					Description: "`void` is only valid as a return type or generic type argument.",
				})
			},
		}
	},
})
