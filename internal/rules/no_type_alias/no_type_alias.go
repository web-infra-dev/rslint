package no_type_alias

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoTypeAliasOptions struct {
	AllowAliases          string `json:"allowAliases,omitempty"`
	AllowCallbacks        string `json:"allowCallbacks,omitempty"`
	AllowConditionalTypes string `json:"allowConditionalTypes,omitempty"`
	AllowConstructors     string `json:"allowConstructors,omitempty"`
	AllowGenerics         string `json:"allowGenerics,omitempty"`
	AllowLiterals         string `json:"allowLiterals,omitempty"`
	AllowMappedTypes      string `json:"allowMappedTypes,omitempty"`
	AllowTupleTypes       string `json:"allowTupleTypes,omitempty"`
}

// Default values for options
const (
	ValueAlways                   = "always"
	ValueNever                    = "never"
	ValueInUnions                 = "in-unions"
	ValueInIntersections          = "in-intersections"
	ValueInUnionsAndIntersections = "in-unions-and-intersections"
)

// CompositionType represents the type of composition (union or intersection)
type CompositionType string

const (
	CompositionTypeUnion        CompositionType = "TSUnionType"
	CompositionTypeIntersection CompositionType = "TSIntersectionType"
)

// TypeWithLabel represents a type node with its composition context
type TypeWithLabel struct {
	Node            *ast.Node
	CompositionType CompositionType
}

// isSupportedComposition checks if the composition type is supported by the allowed flags
func isSupportedComposition(isTopLevel bool, compositionType CompositionType, allowed string) bool {
	compositions := []string{ValueInUnions, ValueInIntersections, ValueInUnionsAndIntersections}
	unions := []string{ValueAlways, ValueInUnions, ValueInUnionsAndIntersections}
	intersections := []string{ValueAlways, ValueInIntersections, ValueInUnionsAndIntersections}

	// Check if allowed value is in compositions slice
	isComposition := false
	for _, v := range compositions {
		if v == allowed {
			isComposition = true
			break
		}
	}

	if !isComposition {
		return true
	}

	if isTopLevel {
		return false
	}

	// Check if composition type is union and allowed in unions
	if compositionType == CompositionTypeUnion {
		for _, v := range unions {
			if v == allowed {
				return true
			}
		}
	}

	// Check if composition type is intersection and allowed in intersections
	if compositionType == CompositionTypeIntersection {
		for _, v := range intersections {
			if v == allowed {
				return true
			}
		}
	}

	return false
}

// isValidTupleType checks if the type is a valid tuple type
func isValidTupleType(typeWithLabel TypeWithLabel) bool {
	node := typeWithLabel.Node

	if node.Kind == ast.KindTupleType {
		return true
	}

	if node.Kind == ast.KindTypeOperator {
		typeOp := node.AsTypeOperatorNode()
		if (typeOp.Operator == ast.KindKeyOfKeyword || typeOp.Operator == ast.KindReadonlyKeyword) &&
			typeOp.Type != nil && typeOp.Type.Kind == ast.KindTupleType {
			return true
		}
	}

	return false
}

// isValidGeneric checks if the type is a valid generic type
func isValidGeneric(typeWithLabel TypeWithLabel) bool {
	node := typeWithLabel.Node
	return node.Kind == ast.KindTypeReference &&
		node.AsTypeReference().TypeArguments != nil
}

// getTypes flattens the given type into an array of its dependencies
func getTypes(node *ast.Node, compositionType CompositionType) []TypeWithLabel {
	if node.Kind == ast.KindUnionType || node.Kind == ast.KindIntersectionType {
		var newCompositionType CompositionType
		if node.Kind == ast.KindUnionType {
			newCompositionType = CompositionTypeUnion
		} else {
			newCompositionType = CompositionTypeIntersection
		}

		var types []TypeWithLabel
		var nodes []*ast.Node

		if node.Kind == ast.KindUnionType {
			nodes = node.AsUnionTypeNode().Types.Nodes
		} else {
			nodes = node.AsIntersectionTypeNode().Types.Nodes
		}

		for _, typeNode := range nodes {
			types = append(types, getTypes(typeNode, newCompositionType)...)
		}

		return types
	}

	return []TypeWithLabel{{Node: node, CompositionType: compositionType}}
}

// List of AST kinds that are considered aliases
var aliasTypes = map[ast.Kind]bool{
	ast.KindArrayType:           true,
	ast.KindImportType:          true,
	ast.KindIndexedAccessType:   true,
	ast.KindLiteralType:         true,
	ast.KindTemplateLiteralType: true,
	ast.KindTypeQuery:           true,
	ast.KindTypeReference:       true,
}

// isKeywordType checks if the node is a keyword type
func isKeywordType(kind ast.Kind) bool {
	switch kind {
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

var NoTypeAliasRule = rule.Rule{
	Name: "no-type-alias",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Default options
		opts := NoTypeAliasOptions{
			AllowAliases:          ValueNever,
			AllowCallbacks:        ValueNever,
			AllowConditionalTypes: ValueNever,
			AllowConstructors:     ValueNever,
			AllowGenerics:         ValueNever,
			AllowLiterals:         ValueNever,
			AllowMappedTypes:      ValueNever,
			AllowTupleTypes:       ValueNever,
		}

		// Parse options with dual-format support (handles both array and object formats)
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			// Handle array format: [{ option: value }]
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				// Handle direct object format: { option: value }
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if val, ok := optsMap["allowAliases"].(string); ok {
					opts.AllowAliases = val
				}
				if val, ok := optsMap["allowCallbacks"].(string); ok {
					opts.AllowCallbacks = val
				}
				if val, ok := optsMap["allowConditionalTypes"].(string); ok {
					opts.AllowConditionalTypes = val
				}
				if val, ok := optsMap["allowConstructors"].(string); ok {
					opts.AllowConstructors = val
				}
				if val, ok := optsMap["allowGenerics"].(string); ok {
					opts.AllowGenerics = val
				}
				if val, ok := optsMap["allowLiterals"].(string); ok {
					opts.AllowLiterals = val
				}
				if val, ok := optsMap["allowMappedTypes"].(string); ok {
					opts.AllowMappedTypes = val
				}
				if val, ok := optsMap["allowTupleTypes"].(string); ok {
					opts.AllowTupleTypes = val
				}
			}
		}

		// reportError reports an error for the given node
		reportError := func(node *ast.Node, compositionType CompositionType, isRoot bool, typeLabel string) {
			if isRoot {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noTypeAlias",
					Description: fmt.Sprintf("Type %s are not allowed.", lowercaseFirst(typeLabel)),
				})
			} else {
				compositionTypeStr := "union"
				if compositionType == CompositionTypeIntersection {
					compositionTypeStr = "intersection"
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noCompositionAlias",
					Description: fmt.Sprintf("%s in %s types are not allowed.", typeLabel, compositionTypeStr),
				})
			}
		}

		// checkAndReport checks if the type is allowed and reports an error if not
		checkAndReport := func(optionValue string, isTopLevel bool, typeWithLabel TypeWithLabel, label string) {
			if optionValue == ValueNever ||
				!isSupportedComposition(isTopLevel, typeWithLabel.CompositionType, optionValue) {
				reportError(typeWithLabel.Node, typeWithLabel.CompositionType, isTopLevel, label)
			}
		}

		// validateTypeAliases validates the node looking for aliases, callbacks and literals
		validateTypeAliases := func(typeWithLabel TypeWithLabel, isTopLevel bool) {
			node := typeWithLabel.Node

			switch node.Kind {
			case ast.KindFunctionType:
				// callback
				if opts.AllowCallbacks == ValueNever {
					reportError(node, typeWithLabel.CompositionType, isTopLevel, "Callbacks")
				}
			case ast.KindConditionalType:
				// conditional type
				if opts.AllowConditionalTypes == ValueNever {
					reportError(node, typeWithLabel.CompositionType, isTopLevel, "Conditional types")
				}
			case ast.KindConstructorType:
				// constructor
				if opts.AllowConstructors == ValueNever {
					reportError(node, typeWithLabel.CompositionType, isTopLevel, "Constructors")
				}
			case ast.KindTypeLiteral:
				// literal object type
				checkAndReport(opts.AllowLiterals, isTopLevel, typeWithLabel, "Literals")
			case ast.KindMappedType:
				// mapped type
				checkAndReport(opts.AllowMappedTypes, isTopLevel, typeWithLabel, "Mapped types")
			default:
				if isValidTupleType(typeWithLabel) {
					// tuple types
					checkAndReport(opts.AllowTupleTypes, isTopLevel, typeWithLabel, "Tuple Types")
				} else if isValidGeneric(typeWithLabel) {
					// generics
					if opts.AllowGenerics == ValueNever {
						reportError(node, typeWithLabel.CompositionType, isTopLevel, "Generics")
					}
				} else if isKeywordType(node.Kind) || aliasTypes[node.Kind] {
					// alias / keyword
					checkAndReport(opts.AllowAliases, isTopLevel, typeWithLabel, "Aliases")
				} else if node.Kind == ast.KindTypeOperator {
					typeOp := node.AsTypeOperatorNode()
					if typeOp.Operator == ast.KindKeyOfKeyword ||
						(typeOp.Operator == ast.KindReadonlyKeyword &&
							typeOp.Type != nil &&
							aliasTypes[typeOp.Type.Kind]) {
						// keyof or readonly with alias type
						checkAndReport(opts.AllowAliases, isTopLevel, typeWithLabel, "Aliases")
					} else {
						// unhandled type - shouldn't happen
						reportError(node, typeWithLabel.CompositionType, isTopLevel, "Unhandled")
					}
				} else {
					// unhandled type - shouldn't happen
					reportError(node, typeWithLabel.CompositionType, isTopLevel, "Unhandled")
				}
			}
		}

		return rule.RuleListeners{
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				typeAlias := node.AsTypeAliasDeclaration()
				types := getTypes(typeAlias.Type, "")

				if len(types) == 1 {
					// is a top level type annotation
					validateTypeAliases(types[0], true)
				} else {
					// is a composition type
					for _, typeWithLabel := range types {
						validateTypeAliases(typeWithLabel, false)
					}
				}
			},
		}
	},
}

// lowercaseFirst converts the first character of a string to lowercase
func lowercaseFirst(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]+32) + s[1:]
}
