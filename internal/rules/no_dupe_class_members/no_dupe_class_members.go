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
			return ast.IsMethodDeclaration(node)
		}

		isProperty := func(node *ast.Node) bool {
			if ast.IsPropertyDeclaration(node) {
				return true
			}
			return false
		}

		getMemberName := func(node *ast.Node) string {
			// First check if it's a numeric literal that needs evaluation
			var nameNode *ast.Node
			switch {
			case ast.IsMethodDeclaration(node):
				nameNode = node.AsMethodDeclaration().Name()
			case ast.IsPropertyDeclaration(node):
				nameNode = node.AsPropertyDeclaration().Name()
			case ast.IsGetAccessorDeclaration(node):
				nameNode = node.AsGetAccessorDeclaration().Name()
			case ast.IsSetAccessorDeclaration(node):
				nameNode = node.AsSetAccessorDeclaration().Name()
			}
			
			// Check if it's a numeric literal and evaluate it
			if nameNode != nil && nameNode.Kind == ast.KindNumericLiteral {
				numLit := nameNode.AsNumericLiteral()
				// Parse the numeric literal text to get its actual value
				// This will convert both "10" and "1e1" to "10"
				var val float64
				fmt.Sscanf(numLit.Text, "%g", &val)
				return fmt.Sprintf("%g", val)
			}
			
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
				fmt.Printf("DEBUG: Skipping computed property\n")
				return
			}

			memberName := getMemberName(memberNode)
			if memberName == "" {
				fmt.Printf("DEBUG: Member name is empty for kind %v\n", memberNode.Kind)
				return
			}
			
			memberIsStatic := isStatic(memberNode)
			memberKind := getMemberKind(memberNode)
			
			fmt.Printf("DEBUG: Processing member %s (kind: %s, static: %v)\n", memberName, memberKind, memberIsStatic)

			// Initialize maps if needed
			if classMembersMap[classNode] == nil {
				classMembersMap[classNode] = make(map[string]map[bool][]MemberInfo)
			}
			if classMembersMap[classNode][memberName] == nil {
				classMembersMap[classNode][memberName] = make(map[bool][]MemberInfo)
			}

			// Check for duplicates in the same static/instance scope
			existingMembers := classMembersMap[classNode][memberName][memberIsStatic]
			
			// Handle duplicate detection based on member types
			for _, existing := range existingMembers {
				if memberKind == "getter" || memberKind == "setter" {
					// For accessors (getters/setters)
					if existing.kind == memberKind {
						// Same accessor type is a duplicate (e.g., two getters or two setters)
						ctx.ReportNode(memberNode, buildUnexpectedMessage(memberName))
						return
					} else if existing.kind == "method" || existing.kind == "property" {
						// Accessor conflicts with method/property
						ctx.ReportNode(memberNode, buildUnexpectedMessage(memberName))
						return
					}
					// Different accessor types (getter/setter) can coexist, so continue
				} else if memberKind == "method" || memberKind == "property" {
					// For methods and properties, they conflict with any existing member
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

				// Debug: check if we get any class at all
				fmt.Printf("DEBUG: Processing class with %d members\n", len(classDecl.Members.Nodes))

				// Process all members of the class
				for _, member := range classDecl.Members.Nodes {
					fmt.Printf("DEBUG: Processing member of kind %v\n", member.Kind)
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