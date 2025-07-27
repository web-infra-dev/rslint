package no_dupe_class_members

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

func buildUnexpectedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: fmt.Sprintf("Duplicate name '%s'.", name),
	}
}

var NoDupeClassMembersRule = rule.Rule{
	Name: "no-dupe-class-members",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		type MemberInfo struct {
			node     *ast.Node
			isStatic bool
			kind     string // "method", "property", "getter", "setter"
		}

		// Track class members: map[className][memberName][static/instance] -> []MemberInfo
		classMembersMap := make(map[*ast.Node]map[string]map[bool][]MemberInfo)

		isMethod := func(node *ast.Node) bool {
			if ast.IsMethodDeclaration(node) {
				method := node.AsMethodDeclaration()
				// Skip TypeScript empty body function expressions (method overloads)
				if method.Body == nil {
					return false
				}
				return true
			}
			return false
		}

		isProperty := func(node *ast.Node) bool {
			if ast.IsPropertyDeclaration(node) {
				return true
			}
			return false
		}

		getMemberName := func(node *ast.Node) string {
			// Use the robust utility function for getting member names
			memberName, _ := utils.GetNameFromMember(ctx.SourceFile, node)
			if memberName != "" {
				return memberName
			}

			// Fallback for specific node types
			switch {
			case ast.IsMethodDeclaration(node):
				method := node.AsMethodDeclaration()
				if method.Name() != nil && ast.IsIdentifier(method.Name()) {
					return method.Name().AsIdentifier().Text
				}
			case ast.IsPropertyDeclaration(node):
				prop := node.AsPropertyDeclaration()
				if prop.Name() != nil && ast.IsIdentifier(prop.Name()) {
					return prop.Name().AsIdentifier().Text
				}
			case ast.IsGetAccessorDeclaration(node):
				getter := node.AsGetAccessorDeclaration()
				if getter.Name() != nil && ast.IsIdentifier(getter.Name()) {
					return getter.Name().AsIdentifier().Text
				}
			case ast.IsSetAccessorDeclaration(node):
				setter := node.AsSetAccessorDeclaration()
				if setter.Name() != nil && ast.IsIdentifier(setter.Name()) {
					return setter.Name().AsIdentifier().Text
				}
			}
			return ""
		}

		isStatic := func(node *ast.Node) bool {
			switch node.Kind {
			case ast.KindMethodDeclaration:
				method := node.AsMethodDeclaration()
				return utils.IncludesModifier(method, ast.KindStaticKeyword)
			case ast.KindPropertyDeclaration:
				prop := node.AsPropertyDeclaration()
				return utils.IncludesModifier(prop, ast.KindStaticKeyword)
			case ast.KindGetAccessor:
				getter := node.AsGetAccessorDeclaration()
				return utils.IncludesModifier(getter, ast.KindStaticKeyword)
			case ast.KindSetAccessor:
				setter := node.AsSetAccessorDeclaration()
				return utils.IncludesModifier(setter, ast.KindStaticKeyword)
			}
			return false
		}

		isComputed := func(node *ast.Node) bool {
			switch {
			case ast.IsMethodDeclaration(node):
				method := node.AsMethodDeclaration()
				return method.Name() != nil && method.Name().Kind == ast.KindComputedPropertyName
			case ast.IsPropertyDeclaration(node):
				prop := node.AsPropertyDeclaration()
				return prop.Name() != nil && prop.Name().Kind == ast.KindComputedPropertyName
			case ast.IsGetAccessorDeclaration(node):
				getter := node.AsGetAccessorDeclaration()
				return getter.Name() != nil && getter.Name().Kind == ast.KindComputedPropertyName
			case ast.IsSetAccessorDeclaration(node):
				setter := node.AsSetAccessorDeclaration()
				return setter.Name() != nil && setter.Name().Kind == ast.KindComputedPropertyName
			}
			return false
		}

		getMemberKind := func(node *ast.Node) string {
			if ast.IsGetAccessorDeclaration(node) {
				return "getter"
			}
			if ast.IsSetAccessorDeclaration(node) {
				return "setter"
			}
			if isMethod(node) {
				return "method"
			}
			if isProperty(node) {
				return "property"
			}
			return ""
		}

		processMember := func(classNode *ast.Node, memberNode *ast.Node) {
			// Skip computed properties
			if isComputed(memberNode) {
				return
			}

			// Skip TypeScript empty body function expressions (method overloads)
			if ast.IsMethodDeclaration(memberNode) {
				method := memberNode.AsMethodDeclaration()
				if method.Body == nil {
					return
				}
			}

			memberName := getMemberName(memberNode)
			if memberName == "" {
				return
			}

			memberIsStatic := isStatic(memberNode)
			memberKind := getMemberKind(memberNode)

			// Initialize maps if needed
			if classMembersMap[classNode] == nil {
				classMembersMap[classNode] = make(map[string]map[bool][]MemberInfo)
			}
			if classMembersMap[classNode][memberName] == nil {
				classMembersMap[classNode][memberName] = make(map[bool][]MemberInfo)
			}

			// Check for duplicates in the same static/instance scope
			existingMembers := classMembersMap[classNode][memberName][memberIsStatic]
			
			// Special handling for getter/setter pairs
			if memberKind == "getter" || memberKind == "setter" {
				// Check if there's already a non-accessor member with the same name
				for _, existing := range existingMembers {
					if existing.kind != "getter" && existing.kind != "setter" {
						// Report duplicate for mixing accessor with non-accessor
						ctx.ReportNode(memberNode, buildUnexpectedMessage(memberName))
						return
					}
					// Check if we already have the same accessor type
					if existing.kind == memberKind {
						ctx.ReportNode(memberNode, buildUnexpectedMessage(memberName))
						return
					}
				}
			} else {
				// For non-accessor members, any existing member is a duplicate
				if len(existingMembers) > 0 {
					ctx.ReportNode(memberNode, buildUnexpectedMessage(memberName))
					return
				}
			}

			// Add the member to tracking
			classMembersMap[classNode][memberName][memberIsStatic] = append(
				existingMembers,
				MemberInfo{
					node:     memberNode,
					isStatic: memberIsStatic,
					kind:     memberKind,
				},
			)
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				classDecl := node.AsClassDeclaration()
				if classDecl.Members == nil {
					return
				}

				// Process all members of the class
				for _, member := range classDecl.Members.Nodes {
					processMember(node, member)
				}
			},
			ast.KindClassExpression: func(node *ast.Node) {
				classExpr := node.AsClassExpression()
				if classExpr.Members == nil {
					return
				}

				// Process all members of the class expression
				for _, member := range classExpr.Members.Nodes {
					processMember(node, member)
				}
			},
		}
	},
}