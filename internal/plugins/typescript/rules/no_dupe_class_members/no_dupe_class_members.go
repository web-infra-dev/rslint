package no_dupe_class_members

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type memberKind uint8

const (
	memberKindInit memberKind = iota
	memberKindGet
	memberKindSet
)

type memberState uint8

const (
	memberStateInit memberState = 1 << iota
	memberStateGet
	memberStateSet
)

type stateEntry struct {
	nonStatic memberState
	static    memberState
}

func registerMember(state memberState, kind memberKind) (memberState, bool) {
	var flag memberState
	var conflicts memberState
	switch kind {
	case memberKindGet:
		flag = memberStateGet
		conflicts = memberStateInit | memberStateGet
	case memberKindSet:
		flag = memberStateSet
		conflicts = memberStateInit | memberStateSet
	default:
		flag = memberStateInit
		conflicts = memberStateInit | memberStateGet | memberStateSet
	}
	return state | flag, state&conflicts != 0
}

var NoDupeClassMembersRule = rule.CreateRule(rule.Rule{
	Name: "no-dupe-class-members",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkClass := func(node *ast.Node) {
			stateMap := make(map[string]stateEntry)

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
				var kind memberKind
				switch {
				case ast.IsGetAccessorDeclaration(member):
					kind = memberKindGet
				case ast.IsSetAccessorDeclaration(member):
					kind = memberKindSet
				case ast.IsMethodDeclaration(member), ast.IsPropertyDeclaration(member):
					kind = memberKindInit
				case ast.IsConstructorDeclaration(member):
					// Static constructor — treated as a static method named "constructor".
					kind = memberKindInit
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

				entry := stateMap[name]
				var isDuplicate bool
				if ast.IsStatic(member) {
					entry.static, isDuplicate = registerMember(entry.static, kind)
				} else {
					entry.nonStatic, isDuplicate = registerMember(entry.nonStatic, kind)
				}
				stateMap[name] = entry

				if isDuplicate {
					reportNode := nameNode
					if reportNode == nil {
						reportNode = member // static constructor (no name node)
					}
					ctx.ReportNode(reportNode, rule.RuleMessage{
						Id:          "unexpected",
						Description: "Duplicate name '" + name + "'.",
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
