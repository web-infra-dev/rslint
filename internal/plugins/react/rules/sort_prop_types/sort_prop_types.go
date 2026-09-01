package sort_prop_types

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed sort_prop_types.schema.json
var schemaJSON []byte

const (
	requiredPropsFirstText = "Required prop types must be listed before all other prop types"
	callbackPropsLastText  = "Callback prop types must be listed after all other prop types"
	propsNotSortedText     = "Prop types declarations should be sorted alphabetically"
)

type options struct {
	requiredFirst, callbacksLast, ignoreCase, noSortAlphabetically, sortShapeProp, checkTypes bool
}

func parseOptions(input []any) options {
	if len(input) == 0 {
		return options{}
	}
	m, _ := input[0].(map[string]any)
	return options{
		requiredFirst:        boolOption(m, "requiredFirst"),
		callbacksLast:        boolOption(m, "callbacksLast"),
		ignoreCase:           boolOption(m, "ignoreCase"),
		noSortAlphabetically: boolOption(m, "noSortAlphabetically"),
		sortShapeProp:        boolOption(m, "sortShapeProp"),
		checkTypes:           boolOption(m, "checkTypes"),
	}
}

func boolOption(m map[string]any, key string) bool {
	v, _ := m[key].(bool)
	return v
}

// SortPropTypesRule enforces the ordering constraints of react/sort-prop-types.
var SortPropTypesRule = rule.Rule{
	Name:   "react/sort-prop-types",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, input []any) rule.RuleListeners {
		opts := parseOptions(input)
		wrappers := reactutil.GetPropWrapperFunctions(ctx.Settings)
		typeAliases := map[string]*ast.Node{}

		report := func(node *ast.Node, id, description string) {
			ctx.ReportNode(node, rule.RuleMessage{Id: id, Description: description})
		}

		checkSorted := func(declarations []*ast.Node) {
			callbackPropsLastSeen := map[*ast.Node]bool{}
			requiredPropsFirstSeen := map[*ast.Node]bool{}
			propsNotSortedSeen := map[*ast.Node]bool{}
			var previous *ast.Node
			for _, current := range declarations {
				if current == nil {
					continue
				}
				if current.Kind == ast.KindSpreadAssignment {
					// A spread starts an independent ordering group.
					previous = nil
					continue
				}
				name, ok := declarationName(ctx.SourceFile, current)
				if !ok {
					continue
				}
				if previous == nil {
					previous = current
					continue
				}
				previousName, _ := declarationName(ctx.SourceFile, previous)
				if opts.requiredFirst {
					if isRequired(previous) && !isRequired(current) {
						previous = current
						continue
					}
					if !isRequired(previous) && isRequired(current) {
						if !requiredPropsFirstSeen[current] {
							requiredPropsFirstSeen[current] = true
							report(current, "requiredPropsFirst", requiredPropsFirstText)
						}
						previous = current
						continue
					}
				}
				if opts.callbacksLast {
					previousCallback := isCallback(previousName)
					currentCallback := isCallback(name)
					if !previousCallback && currentCallback {
						previous = current
						continue
					}
					if previousCallback && !currentCallback {
						if !callbackPropsLastSeen[previous] {
							callbackPropsLastSeen[previous] = true
							report(previous, "callbackPropsLast", callbackPropsLastText)
						}
						continue
					}
				}
				left, right := previousName, name
				if opts.ignoreCase {
					left = ecmascript.StringToLowerCase(left)
					right = ecmascript.StringToLowerCase(right)
				}
				if !opts.noSortAlphabetically && right < left {
					if !propsNotSortedSeen[current] {
						propsNotSortedSeen[current] = true
						report(current, "propsNotSorted", propsNotSortedText)
					}
					continue
				}
				previous = current
			}
		}

		var checkValue func(*ast.Node)
		checkValue = func(value *ast.Node) {
			if value == nil {
				return
			}
			// ESTree removes parentheses, but keeps TypeScript expression wrappers.
			value = ast.SkipParentheses(value)
			if value == nil {
				return
			}
			switch value.Kind {
			case ast.KindObjectLiteralExpression:
				checkSorted(value.AsObjectLiteralExpression().Properties.Nodes)
			case ast.KindIdentifier:
				// The upstream rule resolves an identifier initialized with an
				// object. The binder lookup avoids a whole-file text scan here.
				if decl := declarationObject(ctx.Refs.Resolve(value)); decl != nil {
					checkSorted(decl.AsObjectLiteralExpression().Properties.Nodes)
				}
			case ast.KindCallExpression:
				if !reactutil.IsPropWrapperCall(value, wrappers) {
					return
				}
				args := value.AsCallExpression().Arguments
				if args != nil && len(args.Nodes) != 0 {
					checkValue(args.Nodes[0])
				}
			}
		}

		listeners := rule.RuleListeners{
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				for _, prop := range node.AsObjectLiteralExpression().Properties.Nodes {
					if prop.Kind != ast.KindPropertyAssignment {
						continue
					}
					assignment := prop.AsPropertyAssignment()
					name, ok := utils.GetStaticPropertyName(assignment.Name())
					if ok && name == "propTypes" {
						// Upstream only examines an inline object in this path; identifier
						// resolution belongs to assignment and class-property declarations.
						value := ast.SkipParentheses(assignment.Initializer)
						if value != nil && value.Kind == ast.KindObjectLiteralExpression {
							checkSorted(value.AsObjectLiteralExpression().Properties.Nodes)
						}
					}
				}
			},
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				property := node.AsPropertyDeclaration()
				name, ok := utils.GetStaticPropertyName(property.Name())
				if ok && name == "propTypes" {
					checkValue(property.Initializer)
				}
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				left := reactutil.SkipExpressionWrappers(binary.Left)
				if left == nil || left.Kind != ast.KindPropertyAccessExpression {
					return
				}
				name := left.AsPropertyAccessExpression().Name()
				if name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "propTypes" {
					checkValue(binary.Right)
				}
			},
		}
		if opts.sortShapeProp {
			listeners[ast.KindCallExpression] = func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := reactutil.SkipExpressionWrappers(call.Expression)
				if !isShapeCall(callee) || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				checkValue(call.Arguments.Nodes[0])
			}
		}
		if opts.checkTypes {
			listeners[ast.KindTypeAliasDeclaration] = func(node *ast.Node) {
				decl := node.AsTypeAliasDeclaration()
				if decl.Name() == nil || decl.Name().Kind != ast.KindIdentifier || decl.Type == nil || decl.Type.Kind != ast.KindTypeLiteral {
					return
				}
				name := decl.Name().AsIdentifier().Text
				if _, exists := typeAliases[name]; !exists {
					typeAliases[name] = decl.Type
				}
			}
			checkFunction := func(node *ast.Node) {
				typeNode := firstParamType(node)
				checkTypeNode(typeNode, typeAliases, checkSorted)
			}
			listeners[ast.KindFunctionDeclaration] = checkFunction
			listeners[ast.KindArrowFunction] = checkFunction
		}
		return listeners
	},
}

func declarationName(sourceFile *ast.SourceFile, node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	name := node.Name()
	if name == nil {
		return "", false
	}
	if value, ok := utils.GetStaticPropertyName(name); ok {
		return value, true
	}
	if name.Kind == ast.KindComputedPropertyName {
		return utils.TrimmedNodeText(sourceFile, name.AsComputedPropertyName().Expression), true
	}
	return utils.TrimmedNodeText(sourceFile, name), true
}

// firstParamType mirrors upstream's direct ESTree `param.typeAnnotation`
// lookup. A defaulted parameter becomes an AssignmentPattern there, so it
// must not expose the inner identifier's annotation.
func firstParamType(node *ast.Node) *ast.Node {
	params := reactutil.FunctionParameters(node)
	if len(params) == 0 {
		return nil
	}
	param := params[0].AsParameterDeclaration()
	if param == nil || param.Initializer != nil {
		return nil
	}
	return param.Type
}

func isRequired(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPropertyAssignment {
		return false
	}
	value := reactutil.SkipExpressionWrappers(node.AsPropertyAssignment().Initializer)
	return value != nil && value.Kind == ast.KindPropertyAccessExpression &&
		value.AsPropertyAccessExpression().Name() != nil &&
		value.AsPropertyAccessExpression().Name().Kind == ast.KindIdentifier &&
		value.AsPropertyAccessExpression().Name().AsIdentifier().Text == "isRequired"
}

func isCallback(name string) bool {
	return len(name) >= 3 && strings.HasPrefix(name, "on") && name[2] >= 'A' && name[2] <= 'Z'
}

func isShapeCall(callee *ast.Node) bool {
	if callee == nil {
		return false
	}
	if callee.Kind == ast.KindIdentifier {
		return callee.AsIdentifier().Text == "shape"
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	name := callee.AsPropertyAccessExpression().Name()
	return name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "shape"
}

func checkTypeNode(typeNode *ast.Node, aliases map[string]*ast.Node, check func([]*ast.Node)) {
	if typeNode == nil {
		return
	}
	switch typeNode.Kind {
	case ast.KindTypeLiteral:
		check(typeNode.AsTypeLiteralNode().Members.Nodes)
	case ast.KindTypeReference:
		name := typeNode.AsTypeReferenceNode().TypeName
		if name != nil && name.Kind == ast.KindIdentifier {
			if alias := aliases[name.AsIdentifier().Text]; alias != nil && alias.Kind == ast.KindTypeLiteral {
				check(alias.AsTypeLiteralNode().Members.Nodes)
			}
		}
	}
}

func declarationObject(symbol *ast.Symbol) *ast.Node {
	if symbol == nil || symbol.ValueDeclaration == nil {
		return nil
	}
	decl := symbol.ValueDeclaration
	if decl.Kind != ast.KindVariableDeclaration {
		return nil
	}
	value := reactutil.SkipExpressionWrappers(decl.AsVariableDeclaration().Initializer)
	if value != nil && value.Kind == ast.KindObjectLiteralExpression {
		return value
	}
	return nil
}
