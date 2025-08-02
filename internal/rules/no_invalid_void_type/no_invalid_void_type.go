package no_invalid_void_type

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type Options struct {
	AllowAsThisParameter        bool        `json:"allowAsThisParameter"`
	AllowInGenericTypeArguments interface{} `json:"allowInGenericTypeArguments"`
}

var NoInvalidVoidTypeRule = rule.Rule{
	Name: "no-invalid-void-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := Options{
			AllowAsThisParameter:        false,
			AllowInGenericTypeArguments: true,
		}

		// Parse options with dual-format support (handles both array and object formats)
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
				if allowAsThisParam, ok := optsMap["allowAsThisParameter"].(bool); ok {
					opts.AllowAsThisParameter = allowAsThisParam
				}
				if allowInGeneric, ok := optsMap["allowInGenericTypeArguments"]; ok {
					opts.AllowInGenericTypeArguments = allowInGeneric
				}
			}
		}

		// Helper function to check if node is a return type
		isReturnType := func(node *ast.Node) bool {
			current := node
			for current != nil && current.Parent != nil {
				parent := current.Parent
				switch parent.Kind {
				case ast.KindFunctionDeclaration:
					funcDecl := parent.AsFunctionDeclaration()
					if funcDecl.Type != nil && isNodeInSubtree(funcDecl.Type, node) {
						return true
					}
				case ast.KindMethodDeclaration:
					methodDecl := parent.AsMethodDeclaration()
					if methodDecl.Type != nil && isNodeInSubtree(methodDecl.Type, node) {
						return true
					}
				case ast.KindArrowFunction:
					arrowFunc := parent.AsArrowFunction()
					if arrowFunc.Type != nil && isNodeInSubtree(arrowFunc.Type, node) {
						return true
					}
				case ast.KindFunctionExpression:
					funcExpr := parent.AsFunctionExpression()
					if funcExpr.Type != nil && isNodeInSubtree(funcExpr.Type, node) {
						return true
					}
				case ast.KindConstructSignature, ast.KindCallSignature, ast.KindMethodSignature:
					// These should also allow void return types
					return true
				}
				current = parent
			}
			return false
		}

		// Helper function to check if node is a this parameter
		isThisParameter := func(node *ast.Node) bool {
			if !opts.AllowAsThisParameter {
				return false
			}

			current := node
			for current != nil && current.Parent != nil {
				if current.Parent.Kind == ast.KindParameter {
					param := current.Parent.AsParameterDeclaration()
					if param.Name() != nil {
						paramNameNode := param.Name().AsNode()
						textRange := utils.TrimNodeTextRange(ctx.SourceFile, paramNameNode)
						text := string(ctx.SourceFile.Text()[textRange.Pos():textRange.End()])
						return text == "this"
					}
				}
				current = current.Parent
			}
			return false
		}

		// Helper function to check if void is in union with never (which is allowed)
		isValidUnionWithNever := func(node *ast.Node) bool {
			if node.Parent == nil || node.Parent.Kind != ast.KindUnionType {
				return false
			}

			unionType := node.Parent.AsUnionTypeNode()
			types := unionType.Types.Nodes

			// Check if this union contains 'never' alongside void
			hasNever := false
			for _, t := range types {
				if t.Kind == ast.KindNeverKeyword {
					hasNever = true
					break
				}
			}

			return hasNever
		}

		// Helper function to check if we're in a valid overload context
		isValidOverloadUnion := func(node *ast.Node) bool {
			if node.Parent == nil || node.Parent.Kind != ast.KindUnionType {
				return false
			}

			// Allow void | never unions
			if isValidUnionWithNever(node) {
				return true
			}

			// Allow void | Promise<T> and similar patterns where the other type is a generic with void
			unionType := node.Parent.AsUnionTypeNode()
			types := unionType.Types.Nodes

			for _, t := range types {
				if t == node {
					continue // Skip the void node itself
				}

				// Check if the other type is a generic that might contain void (like Promise<void>)
				if t.Kind == ast.KindTypeReference {
					typeRef := t.AsTypeReference()
					if typeRef.TypeArguments != nil {
						// This is a generic type, check if it uses void as type argument
						for _, arg := range typeRef.TypeArguments.Nodes {
							if arg.Kind == ast.KindVoidKeyword {
								// This union contains void and a generic with void - likely valid (e.g., void | Promise<void>)
								return true
							}
						}
					}
				}
			}

			return false
		}

		return rule.RuleListeners{
			ast.KindVoidKeyword: func(node *ast.Node) {
				// Check for return types first (always allowed)
				if isReturnType(node) {
					return
				}

				// Check for this parameter (allowed if enabled)
				if isThisParameter(node) {
					return
				}

				// Check if it's a valid union with never
				if isValidUnionWithNever(node) {
					return
				}

				// Handle generic type arguments
				if isInGenericTypeArgument(node) {
					// Check allowlist if it's an array
					if allowlist, ok := opts.AllowInGenericTypeArguments.([]interface{}); ok {
						typeRefName := getGenericTypeName(ctx, node)
						if typeRefName != "" {
							allowed := false
							for _, allow := range allowlist {
								if allowStr, ok := allow.(string); ok {
									allowStr = strings.ReplaceAll(allowStr, " ", "")
									if allowStr == typeRefName {
										allowed = true
										break
									}
								}
							}

							if !allowed {
								ctx.ReportNode(node, rule.RuleMessage{
									Id:          "invalidVoidForGeneric",
									Description: fmt.Sprintf("%s may not have void as a type argument.", typeRefName),
								})
								return
							}
						}
						return
					}

					// If allowInGenericTypeArguments is false, report error
					if allowGeneric, ok := opts.AllowInGenericTypeArguments.(bool); ok && !allowGeneric {
						messageId := "invalidVoidNotReturn"
						if opts.AllowAsThisParameter {
							messageId = "invalidVoidNotReturnOrThisParam"
						}
						ctx.ReportNode(node, buildMessage(messageId, ""))
						return
					}

					// If allowInGenericTypeArguments is true (default), allow it
					if allowGeneric, ok := opts.AllowInGenericTypeArguments.(bool); ok && allowGeneric {
						return
					}

					// Default case - if allowInGenericTypeArguments is not explicitly set, default to true
					return
				}

				// Check for valid overload unions (like void | string in overloads)
				if isValidOverloadUnion(node) {
					return
				}

				// For all other cases, report as invalid
				messageId := getInvalidVoidMessageId(node, opts)
				ctx.ReportNode(node, buildMessage(messageId, ""))
			},
		}
	},
}

// Helper function to check if a node is within a subtree of another node
func isNodeInSubtree(root *ast.Node, target *ast.Node) bool {
	if root == target {
		return true
	}

	current := target
	for current != nil {
		if current == root {
			return true
		}
		current = current.Parent
	}
	return false
}

// Helper function to check if node is in a generic type argument
func isInGenericTypeArgument(node *ast.Node) bool {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		if parent.Kind == ast.KindTypeReference {
			typeRef := parent.AsTypeReference()
			if typeRef.TypeArguments != nil {
				// Check if the current node is within the type arguments
				for _, arg := range typeRef.TypeArguments.Nodes {
					if isNodeInSubtree(arg, node) {
						return true
					}
				}
			}
		}
		current = parent
	}
	return false
}

// Helper function to get the generic type name for allowlist checking
func getGenericTypeName(ctx rule.RuleContext, node *ast.Node) string {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		if parent.Kind == ast.KindTypeReference {
			typeRef := parent.AsTypeReference()
			textRange := utils.TrimNodeTextRange(ctx.SourceFile, typeRef.TypeName)
			name := string(ctx.SourceFile.Text()[textRange.Pos():textRange.End()])
			return strings.ReplaceAll(name, " ", "")
		}
		current = parent
	}
	return ""
}

// Helper function to determine message ID based on context
func getInvalidVoidMessageId(node *ast.Node, opts Options) string {
	// Determine base message based on allowed options first
	allowInGeneric := false
	if allowGenericBool, ok := opts.AllowInGenericTypeArguments.(bool); ok {
		allowInGeneric = allowGenericBool
	} else if _, ok := opts.AllowInGenericTypeArguments.([]interface{}); ok {
		allowInGeneric = true
	}

	// Check if we're in a union type (only show union-specific message when generics are allowed)
	// When allowInGenericTypeArguments is false, we should use the basic invalidVoidNotReturn message instead
	if node.Parent != nil && node.Parent.Kind == ast.KindUnionType && allowInGeneric {
		return "invalidVoidUnionConstituent"
	}

	if allowInGeneric && opts.AllowAsThisParameter {
		return "invalidVoidNotReturnOrThisParamOrGeneric"
	} else if allowInGeneric {
		return "invalidVoidNotReturnOrGeneric"
	} else if opts.AllowAsThisParameter {
		return "invalidVoidNotReturnOrThisParam"
	}

	return "invalidVoidNotReturn"
}

// Helper function to build messages
func buildMessage(messageId string, generic string) rule.RuleMessage {
	messages := map[string]string{
		"invalidVoidForGeneric":                    fmt.Sprintf("%s may not have void as a type argument.", generic),
		"invalidVoidNotReturn":                     "void is only valid as a return type.",
		"invalidVoidNotReturnOrGeneric":            "void is only valid as a return type or generic type argument.",
		"invalidVoidNotReturnOrThisParam":          "void is only valid as return type or type of `this` parameter.",
		"invalidVoidNotReturnOrThisParamOrGeneric": "void is only valid as a return type or generic type argument or the type of a `this` parameter.",
		"invalidVoidUnionConstituent":              "void is not valid as a constituent in a union type",
	}

	return rule.RuleMessage{
		Id:          messageId,
		Description: messages[messageId],
	}
}
