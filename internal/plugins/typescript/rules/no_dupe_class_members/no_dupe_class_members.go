package no_dupe_class_members

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoDupeClassMembersRule = rule.CreateRule(rule.Rule{
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
				// Skip the class constructor. In TypeScript-Go's AST, both keyword
				// constructor() and string-literal 'constructor'() parse as
				// KindConstructor. ESLint skips both (kind="constructor"), so we
				// do the same — but only for non-static members. Static constructor()
				// is a regular static method in ESLint (kind="method").
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
				// get + set for the same name is allowed; any other combination is a duplicate.
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

				// Get member name. Static constructors have name=nil in
				// TypeScript-Go's AST; use "constructor" directly.
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

				// "$" prefix avoids collisions with built-in object keys like "__proto__".
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
})
