package no_dupe_class_members

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoDupeClassMembersRule = rule.Rule{
	Name: "no-dupe-class-members",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkClass := func(node *ast.Node) {
			type memberState struct {
				init bool
				get  bool
				set  bool
			}
			type stateEntry struct {
				nonStatic memberState
				static    memberState
			}
			stateMap := make(map[string]*stateEntry)

			for _, member := range node.Members() {
				// Skip non-static constructors. In TypeScript-Go's AST, both the
				// constructor keyword and string-literal 'constructor'() parse as
				// KindConstructor. ESLint skips both (kind="constructor"). Static
				// constructor() is a regular static method (kind="method") in ESLint.
				if ast.IsConstructorDeclaration(member) && !ast.IsStatic(member) {
					continue
				}
				// Overload signatures and abstract declarations (methods, getters,
				// setters) have no body. Skip them so only concrete implementations
				// participate in duplicate checking. PropertyDeclaration never has a
				// body and must not be skipped.
				if !ast.IsPropertyDeclaration(member) && member.Body() == nil {
					continue
				}

				// Determine the duplicate-detection category.
				// A get + set pair for the same name is allowed; anything else collides.
				var kind string
				switch {
				case ast.IsGetAccessorDeclaration(member):
					kind = "get"
				case ast.IsSetAccessorDeclaration(member):
					kind = "set"
				case ast.IsMethodDeclaration(member), ast.IsPropertyDeclaration(member):
					kind = "init"
				case ast.IsConstructorDeclaration(member):
					// Static constructor — treated as a static method named "constructor".
					kind = "init"
				default:
					continue
				}

				// Static constructors have name=nil in TypeScript-Go's AST; use "constructor".
				var name string
				nameNode := ast.GetNameOfDeclaration(member)
				if nameNode != nil {
					var ok bool
					name, ok = utils.GetStaticPropertyName(nameNode)
					if !ok {
						continue // computed property with non-static expression
					}
				} else if ast.IsConstructorDeclaration(member) {
					name = "constructor"
				} else {
					continue
				}

				// "$" prefix avoids collisions with built-in map keys.
				key := "$" + name
				if stateMap[key] == nil {
					stateMap[key] = &stateEntry{}
				}

				state := &stateMap[key].nonStatic
				if ast.IsStatic(member) {
					state = &stateMap[key].static
				}

				var isDuplicate bool
				switch kind {
				case "get":
					isDuplicate = state.init || state.get
					state.get = true
				case "set":
					isDuplicate = state.init || state.set
					state.set = true
				default: // "init"
					isDuplicate = state.init || state.get || state.set
					state.init = true
				}

				if isDuplicate {
					reportNode := nameNode
					// For computed names (e.g. ['foo']), ESLint reports at the inner
					// expression (the resolved key). The tsgo ComputedPropertyName
					// node starts at `[`, so unwrap it to match ESLint's position.
					if reportNode != nil && ast.IsComputedPropertyName(reportNode) {
						reportNode = reportNode.AsComputedPropertyName().Expression
					}
					if reportNode == nil {
						reportNode = member // static constructor (no name node)
					}
					ctx.ReportNode(reportNode, rule.RuleMessage{
						Id:          "unexpected",
						Description: fmt.Sprintf("Duplicate name '%s'.", name),
					})
				}
			}
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: checkClass,
			ast.KindClassExpression:  checkClass,
		}
	},
}
