package no_invalid_void_type

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type NoInvalidVoidTypeOptions struct {
	AllowInGenericTypeArguments interface{} `json:"allowInGenericTypeArguments"`
	AllowAsThisParameter        bool        `json:"allowAsThisParameter"`
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
		return "`void` is only valid as a return type, generic type argument, or `this` parameter type."
	case "invalidVoidNotReturnOrThisParam":
		return "`void` is only valid as a return type or `this` parameter type."
	case "invalidVoidNotReturnOrGeneric":
		return "`void` is only valid as a return type or generic type argument."
	default:
		return "`void` is only valid as a return type."
	}
}

// getEntityNameText reconstructs the dotted name from a TypeName node
// (Identifier or QualifiedName).
func getEntityNameText(name *ast.Node) string {
	if name == nil {
		return ""
	}
	if ast.IsIdentifier(name) {
		return name.AsIdentifier().Text
	}
	if name.Kind == ast.KindQualifiedName {
		qn := name.AsQualifiedName()
		left := getEntityNameText(qn.Left)
		if qn.Right == nil {
			return left
		}
		right := qn.Right.AsIdentifier().Text
		return left + "." + right
	}
	return ""
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

// getParentFunctionDeclarationNode walks up from a union type node to find
// the enclosing FunctionDeclaration or MethodDeclaration that has a body
// (i.e., is an implementation, not a declaration). It only returns a node
// if the union type is in the return type position (not parameter position).
func getParentFunctionDeclarationNode(unionNode *ast.Node) *ast.Node {
	child := unionNode
	current := unionNode.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindFunctionDeclaration:
			// Only return implementations (nodes with a body)
			if current.Body() != nil && current.Type() != nil {
				// Check that we arrived through the return type path
				if child == current.Type() {
					return current
				}
			}
			return nil
		case ast.KindMethodDeclaration:
			if current.Body() != nil && current.Type() != nil {
				if child == current.Type() {
					return current
				}
			}
			return nil
		case ast.KindParameter:
			// In parameter position, not return type
			return nil
		}
		child = current
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

			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if v, exists := optsMap["allowInGenericTypeArguments"]; exists {
					opts.AllowInGenericTypeArguments = v
				}
				if v, exists := optsMap["allowAsThisParameter"]; exists {
					if b, isBool := v.(bool); isBool {
						opts.AllowAsThisParameter = b
					}
				}
			}
		}

		reportNotReturn := func(node *ast.Node) {
			msgId := getNotReturnMessageId(opts)
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          msgId,
				Description: getNotReturnDescription(msgId),
			})
		}

		reportUnionConstituent := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "invalidVoidUnionConstituent",
				Description: "`void` is not valid as a constituent in a union type.",
			})
		}

		reportForGeneric := func(node *ast.Node, genericName string) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "invalidVoidForGeneric",
				Description: "`void` is not valid as a type argument for `" + genericName + "`.",
			})
		}

		// isVoidNeverUnion checks if a union contains only void and never.
		isVoidNeverUnion := func(unionNode *ast.Node) bool {
			union := unionNode.AsUnionTypeNode()
			if union == nil || union.Types == nil {
				return false
			}
			for _, member := range union.Types.Nodes {
				if member.Kind != ast.KindVoidKeyword && member.Kind != ast.KindNeverKeyword {
					return false
				}
			}
			return true
		}

		// isValidUnionWithGenerics checks whether void in a union is valid when
		// allowInGenericTypeArguments is truthy. All non-void/never members must be
		// allowed type references containing void (e.g., void | Promise<void>).
		isValidUnionWithGenerics := func(unionNode *ast.Node) bool {
			union := unionNode.AsUnionTypeNode()
			if union == nil || union.Types == nil {
				return false
			}
			for _, member := range union.Types.Nodes {
				switch member.Kind {
				case ast.KindVoidKeyword, ast.KindNeverKeyword:
					continue
				case ast.KindTypeReference:
					if hasVoidTypeArgument(member) {
						genericName := getEntityNameText(member.AsTypeReferenceNode().TypeName)
						if isGenericAllowedByWhitelist(opts, genericName) {
							continue
						}
					}
					return false
				default:
					return false
				}
			}
			return true
		}

		return rule.RuleListeners{
			ast.KindVoidKeyword: func(node *ast.Node) {
				parent := node.Parent
				if parent == nil {
					reportNotReturn(node)
					return
				}

				switch parent.Kind {
				// --- Union type ---
				case ast.KindUnionType:
					// void | never is always valid
					if isVoidNeverUnion(parent) {
						return
					}
					// Check if the union is in a function/method return type that has overloads
					if declaringFunc := getParentFunctionDeclarationNode(parent); declaringFunc != nil {
						if hasOverloadSignatures(ctx, declaringFunc) {
							return
						}
					}
					// When generics are allowed, check for valid patterns like void | Promise<void>
					if isAllowInGenericTruthy(opts) {
						if isValidUnionWithGenerics(parent) {
							return
						}
						reportUnionConstituent(node)
						return
					}
					// When generics are disabled, any non-void|never union is just invalid void
					reportNotReturn(node)

				// --- Generic type arguments (type-level) ---
				case ast.KindTypeReference:
					if !isAllowInGenericTruthy(opts) {
						reportNotReturn(node)
						return
					}
					// Check whitelist
					if _, isArr := opts.AllowInGenericTypeArguments.([]interface{}); isArr {
						genericName := getEntityNameText(parent.AsTypeReferenceNode().TypeName)
						if !isGenericAllowedByWhitelist(opts, genericName) {
							reportForGeneric(node, normalizeGenericName(genericName))
							return
						}
					}

				// --- Generic type arguments in heritage clauses (extends/implements) ---
				case ast.KindExpressionWithTypeArguments:
					if !isAllowInGenericTruthy(opts) {
						reportNotReturn(node)
						return
					}
					// Check whitelist
					if _, isArr := opts.AllowInGenericTypeArguments.([]interface{}); isArr {
						expr := parent.AsExpressionWithTypeArguments()
						genericName := getEntityNameText(expr.Expression)
						if !isGenericAllowedByWhitelist(opts, genericName) {
							reportForGeneric(node, normalizeGenericName(genericName))
							return
						}
					}

				// --- Generic type arguments on new expressions ---
				case ast.KindNewExpression:
					if !isAllowInGenericTruthy(opts) {
						reportNotReturn(node)
						return
					}

				// --- Default type parameter: <T = void> ---
				case ast.KindTypeParameter:
					typeParam := parent.AsTypeParameter()
					if typeParam.DefaultType == node && isAllowInGenericTruthy(opts) {
						return // void as default is valid when generics are allowed
					}
					reportNotReturn(node)

				// --- Valid return type positions ---
				case ast.KindFunctionType, ast.KindConstructorType,
					ast.KindFunctionDeclaration, ast.KindMethodDeclaration,
					ast.KindArrowFunction, ast.KindFunctionExpression,
					ast.KindMethodSignature, ast.KindCallSignature,
					ast.KindConstructSignature, ast.KindGetAccessor:
					return // always valid

				// --- Parameter: only valid for 'this' parameter ---
				case ast.KindParameter:
					if opts.AllowAsThisParameter {
						param := parent.AsParameterDeclaration()
						if param != nil && param.Name() != nil {
							if id := param.Name().AsIdentifier(); id != nil && id.Text == "this" {
								return
							}
						}
					}
					reportNotReturn(node)

				// --- Everything else is invalid ---
				default:
					reportNotReturn(node)
				}
			},
		}
	},
})
