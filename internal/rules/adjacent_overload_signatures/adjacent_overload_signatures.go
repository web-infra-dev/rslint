package adjacent_overload_signatures

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type Method struct {
	CallSignature bool
	Name          string
	Static        bool
	NameType      utils.MemberNameType
}

// getMemberMethod gets the name and attribute of the member being processed.
// Returns the name and attribute of the member or nil if it's a member not relevant to the rule.
func getMemberMethod(ctx rule.RuleContext, member *ast.Node) *Method {
	if member == nil {
		return nil
	}

	switch member.Kind {
	case ast.KindExportDeclaration:
		// Export declarations (e.g., export { foo }) are not relevant for this rule
		// They don't declare new functions/methods, just re-export existing ones
		return nil

	case ast.KindFunctionDeclaration:
		funcDecl := member.AsFunctionDeclaration()
		if funcDecl == nil || funcDecl.Name() == nil {
			return nil
		}
		name := funcDecl.Name().Text()
		return &Method{
			Name:          name,
			NameType:      utils.MemberNameTypeNormal,
			CallSignature: false,
			Static:        false,
		}

	case ast.KindMethodDeclaration:
		methodDecl := member.AsMethodDeclaration()
		if methodDecl == nil || methodDecl.Name() == nil {
			return nil
		}
		name, nameType := utils.GetNameFromMember(ctx.SourceFile, methodDecl.Name())
		return &Method{
			Name:          name,
			NameType:      nameType,
			CallSignature: false,
			Static:        ast.IsStatic(member),
		}

	case ast.KindMethodSignature:
		methodSig := member.AsMethodSignatureDeclaration()
		if methodSig == nil || methodSig.Name() == nil {
			return nil
		}
		name, nameType := utils.GetNameFromMember(ctx.SourceFile, methodSig.Name())
		return &Method{
			Name:          name,
			NameType:      nameType,
			CallSignature: false,
			Static:        false,
		}

	case ast.KindCallSignature:
		return &Method{
			Name:          "call",
			NameType:      utils.MemberNameTypeNormal,
			CallSignature: true,
			Static:        false,
		}

	case ast.KindConstructSignature:
		return &Method{
			Name:          "new",
			NameType:      utils.MemberNameTypeNormal,
			CallSignature: false,
			Static:        false,
		}

	case ast.KindConstructor:
		return &Method{
			Name:          "constructor",
			NameType:      utils.MemberNameTypeNormal,
			CallSignature: false,
			Static:        false,
		}
	}

	return nil
}

func getMembers(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindClassDeclaration:
		classDecl := node.AsClassDeclaration()
		if classDecl == nil || classDecl.Members == nil {
			return nil
		}
		return classDecl.Members.Nodes
	case ast.KindSourceFile:
		sourceFile := node.AsSourceFile()
		if sourceFile == nil || sourceFile.Statements == nil {
			return nil
		}
		return sourceFile.Statements.Nodes
	case ast.KindModuleBlock:
		moduleBlock := node.AsModuleBlock()
		if moduleBlock == nil || moduleBlock.Statements == nil {
			return nil
		}
		return moduleBlock.Statements.Nodes
	case ast.KindInterfaceDeclaration:
		interfaceDecl := node.AsInterfaceDeclaration()
		if interfaceDecl == nil || interfaceDecl.Members == nil {
			return nil
		}
		return interfaceDecl.Members.Nodes
	case ast.KindBlock:
		block := node.AsBlock()
		if block == nil || block.Statements == nil {
			return nil
		}
		return block.Statements.Nodes
	case ast.KindTypeLiteral:
		typeLiteral := node.AsTypeLiteralNode()
		if typeLiteral == nil || typeLiteral.Members == nil {
			return nil
		}
		return typeLiteral.Members.Nodes
	}
	return nil
}

func checkBodyForOverloadMethods(ctx rule.RuleContext, node *ast.Node) {
	members := getMembers(node)
	if members == nil {
		return
	}

	// Keep track of the last method we saw for each name
	// When we see a method again, check if it was the immediately previous member
	methodLastSeenIndex := make(map[string]int)
	lastMethodIndex := -1

	for memberIdx, member := range members {
		method := getMemberMethod(ctx, member)
		if method == nil {
			// This member is not a method/function
			continue
		}

		// Create a key for this method (includes name, static, callSignature, nameType)
		key := fmt.Sprintf("%s:%t:%t:%d", method.Name, method.Static, method.CallSignature, method.NameType)

		if prevIndex, seen := methodLastSeenIndex[key]; seen {
			// We've seen this method before
			// Check if it was the immediately previous method
			if lastMethodIndex != memberIdx-1 || prevIndex != lastMethodIndex {
				// There was something between the last occurrence and this one
				staticPrefix := ""
				if method.Static {
					staticPrefix = "static "
				}
				ctx.ReportNode(member, rule.RuleMessage{
					Id:          "adjacentSignature",
					Description: fmt.Sprintf("All %s%s signatures should be adjacent.", staticPrefix, method.Name),
				})
			}
		}

		// Update the last seen index for this method
		methodLastSeenIndex[key] = memberIdx
		lastMethodIndex = memberIdx
	}
}

var AdjacentOverloadSignaturesRule = rule.Rule{
	Name: "adjacent-overload-signatures",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Check the source file at the beginning
		checkBodyForOverloadMethods(ctx, &ctx.SourceFile.Node)

		return rule.RuleListeners{
			ast.KindBlock: func(node *ast.Node) {
				checkBodyForOverloadMethods(ctx, node)
			},
			ast.KindClassDeclaration: func(node *ast.Node) {
				checkBodyForOverloadMethods(ctx, node)
			},
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				checkBodyForOverloadMethods(ctx, node)
			},
			ast.KindModuleBlock: func(node *ast.Node) {
				checkBodyForOverloadMethods(ctx, node)
			},
			ast.KindTypeLiteral: func(node *ast.Node) {
				checkBodyForOverloadMethods(ctx, node)
			},
		}
	},
}
