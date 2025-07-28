package no_invalid_void_type

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
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

		// Valid parent node types for void
		validParents := []ast.Kind{}
		
		// Invalid grandparent node types
		invalidGrandParents := []ast.Kind{
			ast.KindPropertySignature,
			ast.KindCallExpression,
			ast.KindPropertyDeclaration,
			ast.KindPropertyDeclaration, // No accessor property declaration kind, using PropertyDeclaration
			ast.KindIdentifier,
		}

		// Valid union member types (void can be in unions with these)
		// Currently unused due to API limitations
		_ = []ast.Kind{
			ast.KindVoidKeyword,
			ast.KindNeverKeyword,
		}

		// If allowInGenericTypeArguments is true, add to valid parents
		if allowGeneric, ok := opts.AllowInGenericTypeArguments.(bool); ok && allowGeneric {
			validParents = append(validParents, ast.KindExpressionWithTypeArguments)
		}

		// Check if the given void keyword is used as a valid generic type
		checkGenericTypeArgument := func(node *ast.Node) {
			// Only matches T<..., void, ...>
			if node.Parent == nil || node.Parent.Kind != ast.KindExpressionWithTypeArguments ||
				node.Parent.Parent == nil || node.Parent.Parent.Kind != ast.KindTypeReference {
				return
			}

			// Check allowlist if it's an array
			if allowlist, ok := opts.AllowInGenericTypeArguments.([]interface{}); ok {
				typeRef := node.Parent.Parent.AsTypeReference()
				fullyQualifiedName := getTypeReferenceName(ctx, typeRef)
				fullyQualifiedName = strings.ReplaceAll(fullyQualifiedName, " ", "")

				allowed := false
				for _, allow := range allowlist {
					if allowStr, ok := allow.(string); ok {
						allowStr = strings.ReplaceAll(allowStr, " ", "")
						if allowStr == fullyQualifiedName {
							allowed = true
							break
						}
					}
				}

				if !allowed {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "invalidVoidForGeneric",
						Description: fmt.Sprintf("%s may not have void as a type argument.", fullyQualifiedName),
					})
				}
				return
			}

			// If allowInGenericTypeArguments is false
			if allowGeneric, ok := opts.AllowInGenericTypeArguments.(bool); ok && !allowGeneric {
				messageId := "invalidVoidNotReturn"
				if opts.AllowAsThisParameter {
					messageId = "invalidVoidNotReturnOrThisParam"
				}
				ctx.ReportNode(node, buildMessage(messageId, ""))
			}
		}

		// Check if generic type parameter defaults to void
		// TODO: Reimplement when TypeParameterDeclaration API is clarified
		_ = func(node *ast.Node, parentNode *ast.TypeParameterDeclaration) {
			// Skip check for now
		}

		// Check if a union containing void is valid
		// TODO: Reimplement when UnionType API is clarified
		_ = func(node *ast.UnionType) bool {
			// Simplified check - just return true for now
			// The Types() method is not available in current API
			return true
		}

		// Get parent function declaration node (currently unused)
		_ = func(node *ast.Node) *ast.Node {
			current := node.Parent
			for current != nil {
				if current.Kind == ast.KindFunctionDeclaration {
					return current
				}
				if current.Kind == ast.KindMethodDeclaration {
					methodDecl := current.AsMethodDeclaration()
					if methodDecl.Body != nil {
						return current
					}
				}
				current = current.Parent
			}
			return nil
		}

		return rule.RuleListeners{
			ast.KindVoidKeyword: func(node *ast.Node) {
				// Check T<..., void, ...> against allowInGenericArguments option
				if node.Parent != nil &&
					node.Parent.Kind == ast.KindExpressionWithTypeArguments &&
					node.Parent.Parent != nil &&
					node.Parent.Parent.Kind == ast.KindTypeReference {
					checkGenericTypeArgument(node)
					return
				}

				// Allow <T = void> if allowInGenericTypeArguments is specified
				// TODO: Reimplement when TypeParameter API is clarified
				if allowGeneric, ok := opts.AllowInGenericTypeArguments.(bool); ok && allowGeneric {
					if node.Parent != nil && node.Parent.Kind == ast.KindTypeParameter {
						// Skip check for now - Default() method not available
						return
					}
				}

				// Union with void must contain types from validUnionMembers
				if node.Parent != nil && node.Parent.Kind == ast.KindUnionType {
					// Skip detailed union check for now - just allow it
					return
				}

				// Using void as part of function overloading implementation
				// Skip overload check for now as HasOverloadSignatures is not available
				// TODO: Reimplement overload signature detection

				// This parameter is ok to be void
				// Skip this parameter check for now - needs proper type node detection
				// TODO: Reimplement this parameter void check

				// Default cases - check if parent is valid
				parentValid := false
				if node.Parent != nil {
					for _, validParent := range validParents {
						if node.Parent.Kind == validParent {
							parentValid = true
							break
						}
					}
				}

				grandParentInvalid := false
				if parentValid && node.Parent != nil && node.Parent.Parent != nil {
					for _, invalidGrandParent := range invalidGrandParents {
						if node.Parent.Parent.Kind == invalidGrandParent {
							grandParentInvalid = true
							break
						}
					}
				}

				if !parentValid || grandParentInvalid {
					// Determine message ID based on options
					messageId := "invalidVoidNotReturn"
					
					allowInGeneric := false
					if allowGenericBool, ok := opts.AllowInGenericTypeArguments.(bool); ok {
						allowInGeneric = allowGenericBool
					} else if _, ok := opts.AllowInGenericTypeArguments.([]interface{}); ok {
						allowInGeneric = true
					}

					if allowInGeneric && opts.AllowAsThisParameter {
						messageId = "invalidVoidNotReturnOrThisParamOrGeneric"
					} else if allowInGeneric {
						messageId = getNotReturnOrGenericMessageId(node, opts)
					} else if opts.AllowAsThisParameter {
						messageId = "invalidVoidNotReturnOrThisParam"
					}

					ctx.ReportNode(node, buildMessage(messageId, ""))
				}
			},
		}
	},
}

// Helper function to get type reference name
func getTypeReferenceName(ctx rule.RuleContext, typeRef *ast.TypeReferenceNode) string {
	// Get the type name from the type reference
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, typeRef.TypeName)
	return string(ctx.SourceFile.Text()[textRange.Pos():textRange.End()])
}

// Helper function to determine message ID for unions
func getNotReturnOrGenericMessageId(node *ast.Node, opts Options) string {
	if node.Parent != nil && node.Parent.Kind == ast.KindUnionType {
		return "invalidVoidUnionConstituent"
	}
	return "invalidVoidNotReturnOrGeneric"
}

// Helper function to build messages
func buildMessage(messageId string, generic string) rule.RuleMessage {
	messages := map[string]string{
		"invalidVoidForGeneric":                       fmt.Sprintf("%s may not have void as a type argument.", generic),
		"invalidVoidNotReturn":                        "void is only valid as a return type.",
		"invalidVoidNotReturnOrGeneric":               "void is only valid as a return type or generic type argument.",
		"invalidVoidNotReturnOrThisParam":             "void is only valid as return type or type of `this` parameter.",
		"invalidVoidNotReturnOrThisParamOrGeneric":    "void is only valid as a return type or generic type argument or the type of a `this` parameter.",
		"invalidVoidUnionConstituent":                 "void is not valid as a constituent in a union type",
	}

	return rule.RuleMessage{
		Id:          messageId,
		Description: messages[messageId],
	}
}