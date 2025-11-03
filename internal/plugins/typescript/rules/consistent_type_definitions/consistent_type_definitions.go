package consistent_type_definitions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type DefinitionStyle string

const (
	DefinitionStyleInterface DefinitionStyle = "interface"
	DefinitionStyleType      DefinitionStyle = "type"
)

type ConsistentTypeDefinitionsOptions struct {
	Style DefinitionStyle `json:"style"`
}

// ConsistentTypeDefinitionsRule enforces consistent type definitions
var ConsistentTypeDefinitionsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-definitions",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := ConsistentTypeDefinitionsOptions{
		Style: DefinitionStyleInterface,
	}

	// Parse options
	if options != nil {
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if str, ok := optArray[0].(string); ok {
				opts.Style = DefinitionStyle(str)
			}
		} else if str, ok := options.(string); ok {
			opts.Style = DefinitionStyle(str)
		}
	}

	// Helper to check if a type is an object type literal (without index signatures or mapped types)
	isObjectTypeLiteral := func(typeNode *ast.Node) bool {
		if typeNode == nil {
			return false
		}
		if typeNode.Kind != ast.KindTypeLiteral {
			return false
		}

		// Check if type literal contains index signatures or mapped types
		typeLiteral := typeNode.AsTypeLiteralNode()
		if typeLiteral == nil || typeLiteral.Members == nil {
			return true
		}

		// If any member is an index signature, this is not a simple object type
		for _, member := range typeLiteral.Members.Nodes {
			if member.Kind == ast.KindIndexSignature {
				return false
			}
		}

		return true
	}

	// Helper to check if a type alias is a simple object type (not a union, intersection, etc.)
	isSimpleObjectType := func(typeNode *ast.Node) bool {
		if typeNode == nil {
			return false
		}

		// Check if it's a parenthesized type wrapping an object type
		if typeNode.Kind == ast.KindParenthesizedType {
			parenthesized := typeNode.AsParenthesizedTypeNode()
			if parenthesized != nil {
				return isObjectTypeLiteral(parenthesized.Type)
			}
		}

		return isObjectTypeLiteral(typeNode)
	}

	// Helper to check if interface is in a globally-scoped module
	isInGlobalModule := func(node *ast.Node) bool {
		current := node.Parent
		for current != nil {
			if current.Kind == ast.KindModuleDeclaration {
				moduleDecl := current.AsModuleDeclaration()
				if moduleDecl != nil && moduleDecl.Name() != nil {
					// Check if module name is 'global'
					if ast.IsIdentifier(moduleDecl.Name()) {
						ident := moduleDecl.Name().AsIdentifier()
						if ident != nil && ident.Text == "global" {
							return true
						}
					}
				}
			}
			current = current.Parent
		}
		return false
	}

	checkTypeAlias := func(node *ast.Node) {
		if opts.Style != DefinitionStyleInterface {
			return
		}

		typeAlias := node.AsTypeAliasDeclaration()
		if typeAlias == nil {
			return
		}

		// Only report if it's a simple object type literal
		if !isSimpleObjectType(typeAlias.Type) {
			return
		}

		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "interfaceOverType",
			Description: "Use an interface instead of a type literal.",
		})
	}

	checkInterface := func(node *ast.Node) {
		if opts.Style != DefinitionStyleType {
			return
		}

		interfaceDecl := node.AsInterfaceDeclaration()
		if interfaceDecl == nil {
			return
		}

		// Don't fix interfaces in global modules (see typescript-eslint #2707)
		if isInGlobalModule(node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "typeOverInterface",
				Description: "Use a type literal instead of an interface.",
			})
			return
		}

		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "typeOverInterface",
			Description: "Use a type literal instead of an interface.",
		})
	}

	return rule.RuleListeners{
		ast.KindTypeAliasDeclaration: checkTypeAlias,
		ast.KindInterfaceDeclaration: checkInterface,
	}
}
