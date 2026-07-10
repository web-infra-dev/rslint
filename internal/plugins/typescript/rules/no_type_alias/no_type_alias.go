package no_type_alias

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type optionValue string

const (
	optionAlways                   optionValue = "always"
	optionNever                    optionValue = "never"
	optionInUnions                 optionValue = "in-unions"
	optionInIntersections          optionValue = "in-intersections"
	optionInUnionsAndIntersections optionValue = "in-unions-and-intersections"
)

type compositionType int

const (
	compositionNone compositionType = iota
	compositionUnion
	compositionIntersection
)

type NoTypeAliasOptions struct {
	AllowAliases          optionValue `json:"allowAliases"`
	AllowCallbacks        optionValue `json:"allowCallbacks"`
	AllowConditionalTypes optionValue `json:"allowConditionalTypes"`
	AllowConstructors     optionValue `json:"allowConstructors"`
	AllowGenerics         optionValue `json:"allowGenerics"`
	AllowLiterals         optionValue `json:"allowLiterals"`
	AllowMappedTypes      optionValue `json:"allowMappedTypes"`
	AllowTupleTypes       optionValue `json:"allowTupleTypes"`
}

func parseSimpleOption(val string) optionValue {
	if val == string(optionAlways) || val == string(optionNever) {
		return optionValue(val)
	}
	return optionNever
}

func parseCompositionOption(val string) optionValue {
	switch optionValue(val) {
	case optionAlways, optionNever, optionInUnions, optionInIntersections, optionInUnionsAndIntersections:
		return optionValue(val)
	}
	return optionNever
}

func parseOptions(options any) NoTypeAliasOptions {
	opts := NoTypeAliasOptions{
		AllowAliases:          optionNever,
		AllowCallbacks:        optionNever,
		AllowConditionalTypes: optionNever,
		AllowConstructors:     optionNever,
		AllowGenerics:         optionNever,
		AllowLiterals:         optionNever,
		AllowMappedTypes:      optionNever,
		AllowTupleTypes:       optionNever,
	}
	if options == nil {
		return opts
	}
	var optsMap map[string]interface{}
	if arr, ok := options.([]interface{}); ok && len(arr) > 0 {
		optsMap, _ = arr[0].(map[string]interface{})
	} else {
		optsMap, _ = options.(map[string]interface{})
	}
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["allowAliases"].(string); ok {
		opts.AllowAliases = parseCompositionOption(v)
	}
	if v, ok := optsMap["allowCallbacks"].(string); ok {
		opts.AllowCallbacks = parseSimpleOption(v)
	}
	if v, ok := optsMap["allowConditionalTypes"].(string); ok {
		opts.AllowConditionalTypes = parseSimpleOption(v)
	}
	if v, ok := optsMap["allowConstructors"].(string); ok {
		opts.AllowConstructors = parseSimpleOption(v)
	}
	if v, ok := optsMap["allowGenerics"].(string); ok {
		opts.AllowGenerics = parseSimpleOption(v)
	}
	if v, ok := optsMap["allowLiterals"].(string); ok {
		opts.AllowLiterals = parseCompositionOption(v)
	}
	if v, ok := optsMap["allowMappedTypes"].(string); ok {
		opts.AllowMappedTypes = parseCompositionOption(v)
	}
	if v, ok := optsMap["allowTupleTypes"].(string); ok {
		opts.AllowTupleTypes = parseCompositionOption(v)
	}
	return opts
}

func buildNoTypeAliasMessage(alias string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noTypeAlias",
		Description: "Type " + strings.ToLower(alias) + " are not allowed.",
	}
}

func buildNoCompositionAliasMessage(typeName string, composition compositionType) rule.RuleMessage {
	compositionName := "intersection"
	if composition == compositionUnion {
		compositionName = "union"
	}

	return rule.RuleMessage{
		Id:          "noCompositionAlias",
		Description: typeName + " in " + compositionName + " types are not allowed.",
	}
}

func unwrapParenthesized(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedType {
		parenthesized := node.AsParenthesizedTypeNode()
		if parenthesized == nil {
			break
		}
		node = parenthesized.Type
	}
	return node
}

var NoTypeAliasRule = rule.CreateRule(rule.Rule{
	Name: "no-type-alias",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		opts := parseOptions(options)

		isSupportedComposition := func(isTopLevel bool, composition compositionType, allowed optionValue) bool {
			if allowed != optionInUnions && allowed != optionInIntersections && allowed != optionInUnionsAndIntersections {
				return true
			}
			if isTopLevel {
				return false
			}
			if composition == compositionUnion {
				return allowed == optionInUnions || allowed == optionInUnionsAndIntersections
			}
			if composition == compositionIntersection {
				return allowed == optionInIntersections || allowed == optionInUnionsAndIntersections
			}
			return false
		}

		reportError := func(node *ast.Node, composition compositionType, isTopLevel bool, aliasLabel string) {
			if node == nil {
				return
			}
			if isTopLevel {
				ctx.ReportNode(node, buildNoTypeAliasMessage(aliasLabel))
				return
			}
			ctx.ReportNode(node, buildNoCompositionAliasMessage(aliasLabel, composition))
		}

		type typeWithComposition struct {
			node        *ast.Node
			composition compositionType
		}

		var getTypes func(node *ast.Node, composition compositionType) []typeWithComposition
		getTypes = func(node *ast.Node, composition compositionType) []typeWithComposition {
			if node == nil {
				return nil
			}
			if node.Kind == ast.KindParenthesizedType {
				return getTypes(unwrapParenthesized(node), composition)
			}
			switch node.Kind {
			case ast.KindUnionType:
				union := node.AsUnionTypeNode()
				if union == nil || union.Types == nil {
					return nil
				}
				var out []typeWithComposition
				for _, t := range union.Types.Nodes {
					out = append(out, getTypes(t, compositionUnion)...)
				}
				return out
			case ast.KindIntersectionType:
				intersection := node.AsIntersectionTypeNode()
				if intersection == nil || intersection.Types == nil {
					return nil
				}
				var out []typeWithComposition
				for _, t := range intersection.Types.Nodes {
					out = append(out, getTypes(t, compositionIntersection)...)
				}
				return out
			default:
				return []typeWithComposition{{node: node, composition: composition}}
			}
		}

		isAliasKind := func(node *ast.Node) bool {
			node = unwrapParenthesized(node)
			if node == nil {
				return false
			}
			switch node.Kind {
			case ast.KindArrayType,
				ast.KindImportType,
				ast.KindIndexedAccessType,
				ast.KindLiteralType,
				ast.KindTemplateLiteralType,
				ast.KindTypeQuery,
				ast.KindTypeReference:
				return true
			default:
				return false
			}
		}

		isKeywordType := func(node *ast.Node) bool {
			node = unwrapParenthesized(node)
			if node == nil {
				return false
			}
			switch node.Kind {
			case ast.KindAnyKeyword,
				ast.KindBigIntKeyword,
				ast.KindBooleanKeyword,
				ast.KindIntrinsicKeyword,
				ast.KindNeverKeyword,
				ast.KindNullKeyword,
				ast.KindNumberKeyword,
				ast.KindObjectKeyword,
				ast.KindStringKeyword,
				ast.KindSymbolKeyword,
				ast.KindUndefinedKeyword,
				ast.KindUnknownKeyword,
				ast.KindVoidKeyword:
				return true
			default:
				return false
			}
		}

		isValidTupleType := func(t typeWithComposition) bool {
			node := unwrapParenthesized(t.node)
			if node == nil {
				return false
			}
			if node.Kind == ast.KindTupleType {
				return true
			}
			if node.Kind != ast.KindTypeOperator {
				return false
			}
			typeOp := node.AsTypeOperatorNode()
			if typeOp == nil || typeOp.Type == nil {
				return false
			}
			if typeOp.Operator != ast.KindKeyOfKeyword && typeOp.Operator != ast.KindReadonlyKeyword {
				return false
			}
			unwrappedType := unwrapParenthesized(typeOp.Type)
			return unwrappedType != nil && unwrappedType.Kind == ast.KindTupleType
		}

		isValidGeneric := func(t typeWithComposition) bool {
			node := unwrapParenthesized(t.node)
			if node == nil || node.Kind != ast.KindTypeReference {
				return false
			}
			typeRef := node.AsTypeReferenceNode()
			return typeRef != nil && typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) > 0
		}

		checkAndReport := func(option optionValue, isTopLevel bool, t typeWithComposition, label string) {
			if option == optionNever || !isSupportedComposition(isTopLevel, t.composition, option) {
				reportError(t.node, t.composition, isTopLevel, label)
			}
		}

		validateTypeAlias := func(t typeWithComposition, isTopLevel bool) {
			if t.node == nil {
				return
			}
			node := unwrapParenthesized(t.node)
			if node == nil {
				return
			}
			switch node.Kind {
			case ast.KindFunctionType:
				if opts.AllowCallbacks == optionNever {
					reportError(t.node, t.composition, isTopLevel, "Callbacks")
				}
				return
			case ast.KindConditionalType:
				if opts.AllowConditionalTypes == optionNever {
					reportError(t.node, t.composition, isTopLevel, "Conditional types")
				}
				return
			case ast.KindConstructorType:
				if opts.AllowConstructors == optionNever {
					reportError(t.node, t.composition, isTopLevel, "Constructors")
				}
				return
			case ast.KindTypeLiteral:
				checkAndReport(opts.AllowLiterals, isTopLevel, t, "Literals")
				return
			case ast.KindMappedType:
				checkAndReport(opts.AllowMappedTypes, isTopLevel, t, "Mapped types")
				return
			}

			if isValidTupleType(t) {
				checkAndReport(opts.AllowTupleTypes, isTopLevel, t, "Tuple Types")
				return
			}

			if isValidGeneric(t) {
				if opts.AllowGenerics == optionNever {
					reportError(t.node, t.composition, isTopLevel, "Generics")
				}
				return
			}

			if isKeywordType(node) || isAliasKind(node) {
				checkAndReport(opts.AllowAliases, isTopLevel, t, "Aliases")
				return
			}

			if node.Kind == ast.KindTypeOperator {
				typeOp := node.AsTypeOperatorNode()
				if typeOp != nil {
					if typeOp.Operator == ast.KindKeyOfKeyword {
						checkAndReport(opts.AllowAliases, isTopLevel, t, "Aliases")
						return
					}
					if typeOp.Operator == ast.KindReadonlyKeyword && typeOp.Type != nil {
						unwrappedOperand := unwrapParenthesized(typeOp.Type)
						if isAliasKind(unwrappedOperand) || isKeywordType(unwrappedOperand) {
							checkAndReport(opts.AllowAliases, isTopLevel, t, "Aliases")
							return
						}
					}
				}
			}

			reportError(t.node, t.composition, isTopLevel, "Unhandled")
		}

		return rule.RuleListeners{
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				decl := node.AsTypeAliasDeclaration()
				if decl == nil || decl.Type == nil {
					return
				}
				types := getTypes(decl.Type, compositionNone)
				if len(types) == 0 {
					return
				}
				if len(types) == 1 {
					validateTypeAlias(types[0], true)
					return
				}
				for _, t := range types {
					validateTypeAlias(t, false)
				}
			},
		}
	},
})
