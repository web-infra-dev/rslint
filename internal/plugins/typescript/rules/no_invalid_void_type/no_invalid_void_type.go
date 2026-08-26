package no_invalid_void_type

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_invalid_void_type.schema.json
var schemaJSON []byte

type NoInvalidVoidTypeOptions struct {
	AllowInGenericTypeArguments interface{} `json:"allowInGenericTypeArguments"`
	AllowAsThisParameter        bool        `json:"allowAsThisParameter"`
}

func parseOptions(options []any) NoInvalidVoidTypeOptions {
	opts := NoInvalidVoidTypeOptions{
		AllowInGenericTypeArguments: true,
		AllowAsThisParameter:        false,
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})
	if v, exists := optsMap["allowInGenericTypeArguments"]; exists {
		opts.AllowInGenericTypeArguments = v
	}
	if v, exists := optsMap["allowAsThisParameter"]; exists {
		if b, isBool := v.(bool); isBool {
			opts.AllowAsThisParameter = b
		}
	}
	return opts
}

// isAllowInGenericTruthy returns true when allowInGenericTypeArguments is not false
// (i.e., true or an array whitelist).
func isAllowInGenericTruthy(opts NoInvalidVoidTypeOptions) bool {
	if allow, ok := opts.AllowInGenericTypeArguments.(bool); ok {
		return allow
	}
	// Array whitelist counts as truthy
	return true
}

// isAllowInGenericTrue returns true only for the boolean true option. Upstream
// treats a whitelist differently from true outside of type references.
func isAllowInGenericTrue(opts NoInvalidVoidTypeOptions) bool {
	allow, ok := opts.AllowInGenericTypeArguments.(bool)
	return ok && allow
}

// getNotReturnMessageId returns the appropriate message ID based on which options are enabled.
func getNotReturnMessageId(opts NoInvalidVoidTypeOptions) string {
	allowGeneric := isAllowInGenericTruthy(opts)
	if opts.AllowAsThisParameter && allowGeneric {
		return "invalidVoidNotReturnOrThisParamOrGeneric"
	}
	if opts.AllowAsThisParameter {
		return "invalidVoidNotReturnOrThisParam"
	}
	if allowGeneric {
		return "invalidVoidNotReturnOrGeneric"
	}
	return "invalidVoidNotReturn"
}

// getNotReturnDescription returns the message description matching the message ID.
func getNotReturnDescription(messageId string) string {
	switch messageId {
	case "invalidVoidNotReturnOrThisParamOrGeneric":
		return "void is only valid as a return type or generic type argument or the type of a `this` parameter."
	case "invalidVoidNotReturnOrThisParam":
		return "void is only valid as return type or type of `this` parameter."
	case "invalidVoidNotReturnOrGeneric":
		return "void is only valid as a return type or generic type argument."
	default:
		return "void is only valid as a return type."
	}
}

// getGenericName mirrors removing ASCII spaces from SourceCode#getText(typeName).
// It intentionally preserves comments and escaped identifier spelling.
func getGenericName(sourceFile *ast.SourceFile, name *ast.Node) string {
	if name == nil {
		return ""
	}
	return normalizeGenericName(utils.TrimmedNodeText(sourceFile, name))
}

// normalizeGenericName removes all spaces for whitelist comparison.
func normalizeGenericName(name string) string {
	return strings.ReplaceAll(name, " ", "")
}

// isGenericAllowedByWhitelist checks if a specific generic type name is permitted
// by the allowInGenericTypeArguments option.
func isGenericAllowedByWhitelist(opts NoInvalidVoidTypeOptions, genericName string) bool {
	if allow, ok := opts.AllowInGenericTypeArguments.(bool); ok {
		return allow
	}
	if whitelist, ok := opts.AllowInGenericTypeArguments.([]interface{}); ok {
		normalized := normalizeGenericName(genericName)
		for _, item := range whitelist {
			if str, ok := item.(string); ok {
				if normalizeGenericName(str) == normalized {
					return true
				}
			}
		}
		return false
	}
	return true
}

// hasVoidTypeArgument checks if a TypeReference node has void as any of its type arguments.
func hasVoidTypeArgument(node *ast.Node) bool {
	typeRef := node.AsTypeReferenceNode()
	if typeRef == nil || typeRef.TypeArguments == nil {
		return false
	}
	for _, arg := range typeRef.TypeArguments.Nodes {
		if arg.Kind == ast.KindVoidKeyword {
			return true
		}
	}
	return false
}

// isValidUnionType mirrors upstream's valid-union check. Generic option
// validation is reported on the nested void keyword itself, so it must not
// cause an additional diagnostic on a sibling void union constituent.
func isValidUnionType(node *ast.Node) bool {
	union := node.AsUnionTypeNode()
	if union == nil || union.Types == nil {
		return false
	}
	for _, member := range union.Types.Nodes {
		switch member.Kind {
		case ast.KindVoidKeyword, ast.KindNeverKeyword:
			continue
		case ast.KindTypeReference:
			if hasVoidTypeArgument(member) {
				continue
			}
		}
		return false
	}
	return true
}

// getParentFunctionDeclarationNode mirrors the ESTree ancestor walk used by
// upstream. A bodyless Go FunctionDeclaration corresponds to ESTree's
// TSDeclareFunction and is therefore skipped.
func getParentFunctionDeclarationNode(unionNode *ast.Node) *ast.Node {
	current := unionNode.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindFunctionDeclaration:
			if current.Body() != nil {
				return current
			}
		case ast.KindMethodDeclaration:
			if current.Body() != nil {
				return current
			}
		}
		current = current.Parent
	}
	return nil
}

// getContainerMembers returns the statements or members of a container node.
func getContainerMembers(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindSourceFile:
		sf := node.AsSourceFile()
		if sf == nil || sf.Statements == nil {
			return nil
		}
		return sf.Statements.Nodes
	case ast.KindBlock:
		block := node.AsBlock()
		if block == nil || block.Statements == nil {
			return nil
		}
		return block.Statements.Nodes
	case ast.KindModuleBlock:
		mb := node.AsModuleBlock()
		if mb == nil || mb.Statements == nil {
			return nil
		}
		return mb.Statements.Nodes
	case ast.KindClassDeclaration:
		cd := node.AsClassDeclaration()
		if cd == nil || cd.Members == nil {
			return nil
		}
		return cd.Members.Nodes
	case ast.KindClassExpression:
		ce := node.AsClassExpression()
		if ce == nil || ce.Members == nil {
			return nil
		}
		return ce.Members.Nodes
	}
	return nil
}

// getFunctionDeclName returns a comparable name string for a function or method
// declaration, used for matching overload signatures with their implementation.
func getFunctionDeclName(ctx rule.RuleContext, node *ast.Node) string {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl == nil {
			return ""
		}
		// export default function - use special sentinel name
		if utils.IncludesModifier(funcDecl, ast.KindDefaultKeyword) {
			return "__default__"
		}
		if funcDecl.Name() == nil {
			return ""
		}
		return funcDecl.Name().Text()
	case ast.KindMethodDeclaration:
		methodDecl := node.AsMethodDeclaration()
		if methodDecl == nil || methodDecl.Name() == nil {
			return ""
		}
		name, _ := utils.GetNameFromMember(ctx.SourceFile, methodDecl.Name())
		return name
	}
	return ""
}

// hasOverloadSignatures checks if the given function/method implementation
// has overload signatures (sibling declarations with the same name but no body)
// in the same container.
func hasOverloadSignatures(ctx rule.RuleContext, node *ast.Node) bool {
	container := node.Parent
	if container == nil {
		return false
	}
	members := getContainerMembers(container)
	if members == nil {
		return false
	}

	nodeName := getFunctionDeclName(ctx, node)
	if nodeName == "" {
		return false
	}

	for _, member := range members {
		if member == node {
			continue
		}
		switch member.Kind {
		case ast.KindFunctionDeclaration, ast.KindMethodDeclaration:
			// An overload declaration is a function/method without a body
			if member.Body() != nil {
				continue
			}
			memberName := getFunctionDeclName(ctx, member)
			if memberName == nodeName {
				return true
			}
		}
	}
	return false
}

var NoInvalidVoidTypeRule = rule.CreateRule(rule.Rule{
	Name:   "no-invalid-void-type",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		allowGeneric := isAllowInGenericTruthy(opts)
		allowAllGenerics := isAllowInGenericTrue(opts)
		_, hasGenericWhitelist := opts.AllowInGenericTypeArguments.([]interface{})
		notReturnMessageId := getNotReturnMessageId(opts)
		notReturnMessage := rule.RuleMessage{
			Id:          notReturnMessageId,
			Description: getNotReturnDescription(notReturnMessageId),
		}

		return rule.RuleListeners{
			ast.KindVoidKeyword: func(node *ast.Node) {
				parent := node.Parent
				if parent == nil {
					ctx.ReportNode(node, notReturnMessage)
					return
				}

				switch parent.Kind {
				// --- Union type ---
				case ast.KindUnionType:
					if isValidUnionType(parent) {
						return
					}
					if declaringFunc := getParentFunctionDeclarationNode(parent); declaringFunc != nil {
						if hasOverloadSignatures(ctx, declaringFunc) {
							return
						}
					}
					if allowGeneric && !opts.AllowAsThisParameter {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "invalidVoidUnionConstituent",
							Description: "void is not valid as a constituent in a union type",
						})
						return
					}
					ctx.ReportNode(node, notReturnMessage)

				// --- Generic type arguments (type-level) ---
				case ast.KindTypeReference:
					if !allowGeneric {
						ctx.ReportNode(node, notReturnMessage)
						return
					}
					if hasGenericWhitelist {
						genericName := getGenericName(ctx.SourceFile, parent.AsTypeReferenceNode().TypeName)
						if !isGenericAllowedByWhitelist(opts, genericName) {
							ctx.ReportNode(node, rule.RuleMessage{
								Id:          "invalidVoidForGeneric",
								Description: genericName + " may not have void as a type argument.",
							})
							return
						}
					}

				// --- Non-reference generic type arguments ---
				case ast.KindExpressionWithTypeArguments, ast.KindNewExpression,
					ast.KindTaggedTemplateExpression, ast.KindImportType, ast.KindTypeQuery,
					ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement:
					if allowAllGenerics {
						return
					}
					ctx.ReportNode(node, notReturnMessage)

				// --- Default type parameter: <T = void> ---
				case ast.KindTypeParameter:
					typeParam := parent.AsTypeParameterDeclaration()
					if allowGeneric && typeParam.DefaultType != nil && typeParam.DefaultType.Kind == ast.KindVoidKeyword {
						if typeParam.DefaultType == node {
							return
						}
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "invalidVoidNotReturnOrGeneric",
							Description: "void is only valid as a return type or generic type argument.",
						})
						return
					}
					ctx.ReportNode(node, notReturnMessage)

				// --- Valid return type positions ---
				case ast.KindFunctionType, ast.KindConstructorType,
					ast.KindFunctionDeclaration, ast.KindMethodDeclaration,
					ast.KindArrowFunction, ast.KindFunctionExpression,
					ast.KindMethodSignature, ast.KindCallSignature,
					ast.KindConstructSignature, ast.KindGetAccessor, ast.KindSetAccessor,
					ast.KindConstructor, ast.KindIndexSignature, ast.KindTypePredicate:
					return // always valid

				// --- Parameters ---
				case ast.KindParameter:
					param := parent.AsParameterDeclaration()
					if param != nil && param.Name() != nil {
						if param.DotDotDotToken != nil || ast.IsBindingPattern(param.Name()) {
							return
						}
						if opts.AllowAsThisParameter {
							if id := param.Name().AsIdentifier(); id != nil && id.Text == "this" {
								return
							}
						}
					}
					ctx.ReportNode(node, notReturnMessage)

				// ESTree permits destructuring annotations because their grandparent is
				// an ObjectPattern or ArrayPattern rather than an Identifier.
				case ast.KindVariableDeclaration:
					declaration := parent.AsVariableDeclaration()
					if declaration != nil && declaration.Name() != nil && ast.IsBindingPattern(declaration.Name()) {
						return
					}
					ctx.ReportNode(node, notReturnMessage)

				// --- Everything else is invalid ---
				default:
					ctx.ReportNode(node, notReturnMessage)
				}
			},
		}
	},
})
