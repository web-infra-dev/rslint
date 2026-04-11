package parameter_properties

import (
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Modifier values that match the ESLint option schema:
// "readonly", "private", "protected", "public",
// "private readonly", "protected readonly", "public readonly"

type options struct {
	allow  map[string]bool
	prefer string
}

func parseOptions(rawOpts any) options {
	o := options{
		allow:  make(map[string]bool),
		prefer: "class-property",
	}
	optsMap := utils.GetOptionsMap(rawOpts)
	if optsMap == nil {
		return o
	}
	if allow, ok := optsMap["allow"].([]interface{}); ok {
		for _, a := range allow {
			if s, ok := a.(string); ok {
				o.allow[s] = true
			}
		}
	}
	if prefer, ok := optsMap["prefer"].(string); ok {
		o.prefer = prefer
	}
	return o
}

// getModifiers builds the ESLint-style modifier string (e.g. "public readonly") from a node's
// accessibility and readonly flags. The result is used to match against the user's "allow" list.
// NOTE: class_literal_property_style has a similar printNodeModifiers, but it also handles
// "static" and serves a different purpose (generating replacement code text).
func getModifiers(node *ast.Node) string {
	var parts []string
	flags := ast.GetCombinedModifierFlags(node)
	if flags&ast.ModifierFlagsPublic != 0 {
		parts = append(parts, "public")
	} else if flags&ast.ModifierFlagsProtected != 0 {
		parts = append(parts, "protected")
	} else if flags&ast.ModifierFlagsPrivate != 0 {
		parts = append(parts, "private")
	}
	if flags&ast.ModifierFlagsReadonly != 0 {
		parts = append(parts, "readonly")
	}
	return strings.Join(parts, " ")
}

func buildPreferClassPropertyMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferClassProperty",
		Description: fmt.Sprintf("Property %s should be declared as a class property.", name),
	}
}

func buildPreferParameterPropertyMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferParameterProperty",
		Description: fmt.Sprintf("Property %s should be declared as a parameter property.", name),
	}
}

// propertyNodes tracks the three pieces that must all exist for a
// "prefer parameter-property" violation: the class property declaration,
// a matching constructor parameter, and a leading `this.X = X` assignment.
type propertyNodes struct {
	classProperty         *ast.Node
	constructorParameter  *ast.Node
	constructorAssignment bool
}

var ParameterPropertiesRule = rule.CreateRule(rule.Rule{
	Name: "parameter-properties",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		if opts.prefer == "class-property" {
			return rule.RuleListeners{
				ast.KindParameter: func(node *ast.Node) {
					if node.Parent == nil {
						return
					}
					// Check if this parameter is a parameter property
					if !ast.IsParameterPropertyDeclaration(node, node.Parent) {
						return
					}

					param := node.AsParameterDeclaration()
					if param == nil {
						return
					}

					// Skip rest parameters (e.g., private ...name: string[])
					if param.DotDotDotToken != nil {
						return
					}

					// Skip destructured parameters (e.g., private [test]: [string])
					nameNode := node.Name()
					if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
						return
					}

					modifiers := getModifiers(node)
					if opts.allow[modifiers] {
						return
					}

					paramName := nameNode.AsIdentifier().Text
					ctx.ReportNode(node, buildPreferClassPropertyMessage(paramName))
				},
			}
		}

		// "parameter-property" mode: use a stack to handle nested classes.
		// Each map tracks {name → propertyNodes} for the current class scope.
		var propertyNodesByNameStack []map[string]*propertyNodes

		getNodesByName := func(name string) *propertyNodes {
			m := propertyNodesByNameStack[len(propertyNodesByNameStack)-1]
			if existing, ok := m[name]; ok {
				return existing
			}
			created := &propertyNodes{}
			m[name] = created
			return created
		}

		// typeAnnotationsMatch checks that both nodes either lack type annotations or
		// have identical annotation text. ESLint compares getText(TSTypeAnnotation)
		// which includes leading trivia (whitespace after the colon). We use
		// GetSourceTextOfNodeFromSourceFile with includeTrivia=true to preserve
		// the same trivia-sensitive comparison — TrimmedNodeText would strip the
		// trivia and falsely match cases like ":  string" vs ": string".
		typeAnnotationsMatch := func(classProp *ast.Node, ctorParam *ast.Node) bool {
			propType := classProp.AsPropertyDeclaration().Type
			paramType := ctorParam.AsParameterDeclaration().Type

			if propType == nil || paramType == nil {
				return propType == paramType
			}

			return scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, propType, true) ==
				scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, paramType, true)
		}

		enterClass := func(node *ast.Node) {
			propertyNodesByNameStack = append(propertyNodesByNameStack, make(map[string]*propertyNodes))

			// Process class body members (equivalent to ClassBody visitor in ESLint)
			members := node.Members()
			for _, member := range members {
				if member.Kind != ast.KindPropertyDeclaration {
					continue
				}
				propDecl := member.AsPropertyDeclaration()
				if propDecl == nil {
					continue
				}
				nameNode := propDecl.Name()
				if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
					continue
				}
				// Skip properties with initializers
				if propDecl.Initializer != nil {
					continue
				}
				// Skip if modifier is in allow list
				if opts.allow[getModifiers(member)] {
					continue
				}
				getNodesByName(nameNode.AsIdentifier().Text).classProperty = member
			}
		}

		exitClass := func(node *ast.Node) {
			if len(propertyNodesByNameStack) == 0 {
				return
			}
			m := propertyNodesByNameStack[len(propertyNodesByNameStack)-1]
			propertyNodesByNameStack = propertyNodesByNameStack[:len(propertyNodesByNameStack)-1]

			// Collect violations and sort by source position so that
			// diagnostics are emitted in deterministic document order
			// (Go map iteration order is not guaranteed).
			type violation struct {
				name string
				node *ast.Node
			}
			var violations []violation
			for name, nodes := range m {
				if nodes.classProperty != nil &&
					nodes.constructorAssignment &&
					nodes.constructorParameter != nil &&
					typeAnnotationsMatch(nodes.classProperty, nodes.constructorParameter) {
					violations = append(violations, violation{name, nodes.classProperty})
				}
			}
			sort.Slice(violations, func(i, j int) bool {
				return violations[i].node.Pos() < violations[j].node.Pos()
			})
			for _, v := range violations {
				ctx.ReportNode(v.node, buildPreferParameterPropertyMessage(v.name))
			}
		}

		listeners := rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				enterClass(node)
			},
			rule.ListenerOnExit(ast.KindClassDeclaration): func(node *ast.Node) {
				exitClass(node)
			},
			ast.KindClassExpression: func(node *ast.Node) {
				enterClass(node)
			},
			rule.ListenerOnExit(ast.KindClassExpression): func(node *ast.Node) {
				exitClass(node)
			},
			ast.KindConstructor: func(node *ast.Node) {
				if len(propertyNodesByNameStack) == 0 {
					return
				}

				constructor := node.AsConstructorDeclaration()
				if constructor == nil {
					return
				}

				// Record plain constructor parameters (not parameter properties).
				// In ESLint's AST, TSParameterProperty is a separate node type that
				// doesn't match Identifier, so parameter properties are naturally skipped.
				// In Go's AST there is no separate kind — we must explicitly skip them.
				var params []*ast.Node
				if constructor.Parameters != nil {
					params = constructor.Parameters.Nodes
				}
				for _, param := range params {
					if param.Kind != ast.KindParameter {
						continue
					}
					// Skip parameter properties (already have a modifier)
					if ast.IsParameterPropertyDeclaration(param, node) {
						continue
					}
					paramDecl := param.AsParameterDeclaration()
					if paramDecl == nil {
						continue
					}
					// Skip rest parameters — ESLint's RestElement !== Identifier.
					if paramDecl.DotDotDotToken != nil {
						continue
					}
					// Skip parameters with default values — ESLint's AssignmentPattern !== Identifier.
					if paramDecl.Initializer != nil {
						continue
					}
					nameNode := param.Name()
					if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
						continue
					}
					getNodesByName(nameNode.AsIdentifier().Text).constructorParameter = param
				}

				// Scan leading statements of the form `this.X = Y` (where Y is an identifier).
				// Stop at the first non-matching statement, matching ESLint's break behavior.
				if constructor.Body == nil {
					return
				}
				statements := constructor.Body.Statements()
				for _, stmt := range statements {
					if stmt.Kind != ast.KindExpressionStatement {
						break
					}
					expr := stmt.Expression()
					if expr == nil || expr.Kind != ast.KindBinaryExpression {
						break
					}
					binExpr := expr.AsBinaryExpression()
					// ESLint checks for AssignmentExpression which covers all assignment
					// operators (=, +=, -=, etc.), not just plain =.
					if !ast.IsAssignmentOperator(binExpr.OperatorToken.Kind) {
						break
					}
					// Left side must be this.X — ESLint checks MemberExpression which
					// covers both PropertyAccessExpression (this.x) and
					// ElementAccessExpression (this[x]) with an Identifier argument.
					left := binExpr.Left
					if left.Kind == ast.KindPropertyAccessExpression {
						propAccess := left.AsPropertyAccessExpression()
						if propAccess.Expression.Kind != ast.KindThisKeyword {
							break
						}
						if propAccess.Name().Kind != ast.KindIdentifier {
							break
						}
					} else if left.Kind == ast.KindElementAccessExpression {
						elemAccess := left.AsElementAccessExpression()
						if elemAccess.Expression.Kind != ast.KindThisKeyword {
							break
						}
						if elemAccess.ArgumentExpression == nil || elemAccess.ArgumentExpression.Kind != ast.KindIdentifier {
							break
						}
					} else {
						break
					}
					// Right side must be an identifier
					right := binExpr.Right
					if right.Kind != ast.KindIdentifier {
						break
					}
					rightName := right.AsIdentifier().Text
					getNodesByName(rightName).constructorAssignment = true
				}
			},
		}

		return listeners
	},
})
