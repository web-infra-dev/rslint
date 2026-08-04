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

// Keep common classes allocation-free while bounding the linear lookup cost.
// Overflow starts at twice this size instead of using the raw member count,
// which may be inflated by ignored overloads or dynamic computed members.
const inlineClassStateCapacity = 32

type namedStateEntry struct {
	name  string
	state stateEntry
}

type classState struct {
	inline   [inlineClassStateCapacity]namedStateEntry
	count    int
	overflow map[string]stateEntry
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

func registerState(entry *stateEntry, static bool, kind memberKind) bool {
	state := &entry.nonStatic
	if static {
		state = &entry.static
	}
	var isDuplicate bool
	*state, isDuplicate = registerMember(*state, kind)
	return isDuplicate
}

func (state *classState) register(name string, static bool, kind memberKind) bool {
	if state.overflow != nil {
		entry := state.overflow[name]
		isDuplicate := registerState(&entry, static, kind)
		state.overflow[name] = entry
		return isDuplicate
	}

	for index := range state.count {
		entry := &state.inline[index]
		if entry.name == name {
			return registerState(&entry.state, static, kind)
		}
	}
	if state.count < len(state.inline) {
		entry := &state.inline[state.count]
		entry.name = name
		state.count++
		return registerState(&entry.state, static, kind)
	}

	state.overflow = make(map[string]stateEntry, inlineClassStateCapacity*2)
	for index := range state.count {
		entry := &state.inline[index]
		state.overflow[entry.name] = entry.state
	}
	entry := stateEntry{}
	isDuplicate := registerState(&entry, static, kind)
	state.overflow[name] = entry
	return isDuplicate
}

var NoDupeClassMembersRule = rule.CreateRule(rule.Rule{
	Name: "no-dupe-class-members",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkClass := func(node *ast.Node) {
			members := node.Members()
			if len(members) < 2 {
				return
			}
			state := classState{}

			for _, member := range members {
				// Determine the duplicate-detection category.
				// get + set for the same name is allowed; any other combination is a duplicate.
				var kind memberKind
				switch member.Kind {
				case ast.KindGetAccessor:
					kind = memberKindGet
				case ast.KindSetAccessor:
					kind = memberKindSet
				case ast.KindMethodDeclaration, ast.KindPropertyDeclaration:
					kind = memberKindInit
				case ast.KindConstructor:
					// Static constructor — treated as a static method named "constructor".
					kind = memberKindInit
				default:
					continue
				}
				var isStatic bool
				// In TypeScript-Go's AST, both keyword constructor() and string-literal
				// 'constructor'() parse as KindConstructor. ESLint skips both
				// (kind="constructor"), so we do the same — but only for non-static
				// members. Static constructor() is a regular static method in ESLint.
				if member.Kind == ast.KindConstructor {
					isStatic = ast.IsStatic(member)
					if !isStatic {
						continue
					}
				}
				// Overload signatures and abstract declarations (methods, getters,
				// setters) have no body. Skip them so only concrete implementations
				// participate in duplicate checking. PropertyDeclaration never has a
				// body and must not be skipped.
				if member.Kind != ast.KindPropertyDeclaration && member.Body() == nil {
					continue
				}
				if member.Kind != ast.KindConstructor {
					isStatic = ast.IsStatic(member)
				}

				// Get member name. Static constructors have name=nil in
				// TypeScript-Go's AST; use "constructor" directly.
				var name string
				nameNode := ast.GetNameOfDeclaration(member)
				if nameNode != nil {
					var ok = true
					switch nameNode.Kind {
					case ast.KindIdentifier:
						name = nameNode.AsIdentifier().Text
					case ast.KindStringLiteral:
						name = nameNode.AsStringLiteral().Text
					default:
						name, ok = utils.GetStaticPropertyName(nameNode)
					}
					if !ok {
						continue // computed property with non-static expression
					}
				} else if member.Kind == ast.KindConstructor {
					name = "constructor"
				} else {
					continue
				}

				if state.register(name, isStatic, kind) {
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
