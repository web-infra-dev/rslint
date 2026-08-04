package no_dupe_keys

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type propertyKind uint8

const (
	propertyKindInit propertyKind = iota
	propertyKindGet
	propertyKindSet
)

type propertyState uint8

const (
	propertyStateInit propertyState = 1 << iota
	propertyStateGet
	propertyStateSet
)

// Keep common object literals allocation-free while bounding lookup cost for
// generated or configuration-heavy objects.
const inlineObjectStateCapacity = 16

type objectState struct {
	names    [inlineObjectStateCapacity]string
	states   [inlineObjectStateCapacity]propertyState
	count    int
	overflow map[string]propertyState
}

func registerProperty(state propertyState, kind propertyKind) (propertyState, bool) {
	var flag propertyState
	var conflicts propertyState
	switch kind {
	case propertyKindGet:
		flag = propertyStateGet
		conflicts = propertyStateInit | propertyStateGet
	case propertyKindSet:
		flag = propertyStateSet
		conflicts = propertyStateInit | propertyStateSet
	default:
		flag = propertyStateInit
		conflicts = propertyStateInit | propertyStateGet | propertyStateSet
	}
	return state | flag, state&conflicts != 0
}

func (state *objectState) register(name string, kind propertyKind) bool {
	if state.overflow != nil {
		current := state.overflow[name]
		var duplicate bool
		state.overflow[name], duplicate = registerProperty(current, kind)
		return duplicate
	}

	for index := range state.count {
		if state.names[index] == name {
			var duplicate bool
			state.states[index], duplicate = registerProperty(state.states[index], kind)
			return duplicate
		}
	}
	if state.count < len(state.names) {
		state.names[state.count] = name
		state.states[state.count], _ = registerProperty(0, kind)
		state.count++
		return false
	}

	state.overflow = make(map[string]propertyState, inlineObjectStateCapacity*2)
	for index := range state.count {
		state.overflow[state.names[index]] = state.states[index]
	}
	state.overflow[name], _ = registerProperty(0, kind)
	return false
}

// https://eslint.org/docs/latest/rules/no-dupe-keys
var NoDupeKeysRule = rule.Rule{
	Name: "no-dupe-keys",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				objLit := node.AsObjectLiteralExpression()
				if objLit == nil || objLit.Properties == nil || len(objLit.Properties.Nodes) < 2 {
					return
				}
				state := objectState{}

				for _, prop := range objLit.Properties.Nodes {
					var kind propertyKind
					switch prop.Kind {
					case ast.KindSpreadAssignment:
						continue
					case ast.KindGetAccessor:
						kind = propertyKindGet
					case ast.KindSetAccessor:
						kind = propertyKindSet
					default:
						kind = propertyKindInit
					}

					nameNode := prop.Name()
					if nameNode == nil {
						continue
					}

					var name string
					var isStatic = true
					switch nameNode.Kind {
					case ast.KindIdentifier:
						name = nameNode.AsIdentifier().Text
					case ast.KindStringLiteral:
						name = nameNode.AsStringLiteral().Text
					case ast.KindBigIntLiteral:
						name = utils.NormalizeBigIntLiteral(nameNode.AsBigIntLiteral().Text)
					default:
						name, isStatic = utils.GetStaticPropertyName(nameNode)
					}
					if !isStatic {
						continue
					}

					// Skip __proto__ setters: non-computed __proto__ in PropertyAssignment is
					// a prototype setter, not a regular property. Computed ["__proto__"] is regular.
					if name == "__proto__" && prop.Kind == ast.KindPropertyAssignment && nameNode.Kind != ast.KindComputedPropertyName {
						continue
					}

					if state.register(name, kind) {
						ctx.ReportNode(prop, rule.RuleMessage{
							Id:          "unexpected",
							Description: "Duplicate key '" + name + "'.",
						})
					}
				}
			},
		}
	},
}
