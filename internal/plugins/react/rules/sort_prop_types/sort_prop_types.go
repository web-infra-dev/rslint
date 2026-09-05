package sort_prop_types

import (
	_ "embed"
	"math"
	"math/big"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
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

type propName struct {
	text   string
	number float64
	bigInt *big.Int
	kind   propNameKind
}

type propNameKind uint8

const (
	propNameString propNameKind = iota
	propNameNumber
	propNameBigInt
	propNameBoolean
)

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
		var typeAliases map[string]*ast.Node

		report := func(node *ast.Node, id, description string) {
			ctx.ReportNode(node, rule.RuleMessage{Id: id, Description: description})
		}

		checkSorted := func(declarations []*ast.Node, typeMembers bool) {
			type declarationState struct {
				node        *ast.Node
				name        propName
				compareText string
				required    bool
				callback    bool
			}

			callbackPropsLastSeen := map[*ast.Node]bool{}
			requiredPropsFirstSeen := map[*ast.Node]bool{}
			propsNotSortedSeen := map[*ast.Node]bool{}
			var previous declarationState
			hasPrevious := false
			for _, current := range declarations {
				if current == nil {
					continue
				}
				if current.Kind == ast.KindSpreadAssignment {
					// A spread starts an independent ordering group.
					hasPrevious = false
					continue
				}
				name, ok := declarationName(ctx.SourceFile, current, typeMembers)
				if !ok {
					// An unnamed declaration breaks the comparison chain.
					hasPrevious = false
					continue
				}
				state := declarationState{node: current, name: name}
				if opts.requiredFirst {
					state.required = isRequired(current)
				}
				if opts.callbacksLast {
					state.callback = isCallback(name.text)
				}
				if !opts.noSortAlphabetically && opts.ignoreCase {
					state.compareText = ecmascript.StringToLowerCase(propNameToString(name))
				}
				if !hasPrevious {
					previous = state
					hasPrevious = true
					continue
				}
				if opts.requiredFirst {
					if previous.required && !state.required {
						previous = state
						continue
					}
					if !previous.required && state.required {
						if !requiredPropsFirstSeen[current] {
							requiredPropsFirstSeen[current] = true
							report(current, "requiredPropsFirst", requiredPropsFirstText)
						}
						previous = state
						continue
					}
				}
				if opts.callbacksLast {
					if !previous.callback && state.callback {
						previous = state
						continue
					}
					if previous.callback && !state.callback {
						if !callbackPropsLastSeen[previous.node] {
							callbackPropsLastSeen[previous.node] = true
							report(previous.node, "callbackPropsLast", callbackPropsLastText)
						}
						continue
					}
				}
				if !opts.noSortAlphabetically {
					outOfOrder := isPropNameBefore(state.name, previous.name)
					if opts.ignoreCase {
						outOfOrder = ecmascript.CompareStrings(state.compareText, previous.compareText) < 0
					}
					if outOfOrder {
						if !propsNotSortedSeen[current] {
							propsNotSortedSeen[current] = true
							report(current, "propsNotSorted", propsNotSortedText)
						}
						continue
					}
				}
				previous = state
			}
		}

		var checkValue func(*ast.Node)
		checkValue = func(value *ast.Node) {
			if value == nil {
				return
			}
			value = utils.ESTreeRuntimeExpression(value)
			if value == nil {
				return
			}
			switch value.Kind {
			case ast.KindObjectLiteralExpression:
				checkSorted(value.AsObjectLiteralExpression().Properties.Nodes, false)
			case ast.KindIdentifier:
				// The upstream rule resolves an identifier initialized with an
				// object. The binder lookup avoids a whole-file text scan here.
				if decl := declarationObject(ctx.SourceFile, ctx.Refs.Resolve(value)); decl != nil {
					checkSorted(decl.AsObjectLiteralExpression().Properties.Nodes, false)
				}
			case ast.KindCallExpression:
				if !isPropWrapperCallByCalleeName(value, wrappers) {
					return
				}
				args := value.AsCallExpression().Arguments
				if args != nil && len(args.Nodes) != 0 {
					checkValue(args.Nodes[0])
				}
			}
		}

		checkPropTypesParent := func(left, right *ast.Node) {
			if isPropTypesAccess(left) {
				checkValue(right)
			}
		}

		listeners := rule.RuleListeners{
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				for _, prop := range node.AsObjectLiteralExpression().Properties.Nodes {
					if prop.Kind != ast.KindPropertyAssignment {
						continue
					}
					assignment := prop.AsPropertyAssignment()
					if isAuthoredPropertyName(assignment.Name(), "propTypes") {
						// Upstream only examines an inline object in this path; identifier
						// resolution belongs to assignment and class-property declarations.
						value := utils.ESTreeRuntimeExpression(assignment.Initializer)
						if value != nil && value.Kind == ast.KindObjectLiteralExpression {
							checkSorted(value.AsObjectLiteralExpression().Properties.Nodes, false)
						}
					}
				}
			},
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				property := node.AsPropertyDeclaration()
				if hasModifier(property.Modifiers(), ast.KindAccessorKeyword) {
					return
				}
				if property.Type != nil && isAuthoredPropertyName(property.Name(), "props") {
					checkValue(property.Initializer)
					return
				}
				if isAuthoredPropertyName(property.Name(), "propTypes") {
					checkValue(property.Initializer)
				}
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken == nil || binary.OperatorToken.Kind == ast.KindCommaToken {
					return
				}
				checkPropTypesParent(binary.Left, binary.Right)
			},
			ast.KindForInStatement: func(node *ast.Node) {
				statement := node.AsForInOrOfStatement()
				checkPropTypesParent(statement.Initializer, statement.Expression)
			},
			ast.KindForOfStatement: func(node *ast.Node) {
				statement := node.AsForInOrOfStatement()
				checkPropTypesParent(statement.Initializer, statement.Expression)
			},
		}
		if opts.sortShapeProp {
			listeners[ast.KindCallExpression] = func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := utils.ESTreeCallCallee(call.Expression)
				if !isShapeCall(callee) || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				firstArg := utils.ESTreeRuntimeExpression(call.Arguments.Nodes[0])
				if firstArg == nil {
					return
				}
				switch firstArg.Kind {
				case ast.KindObjectLiteralExpression:
					checkSorted(firstArg.AsObjectLiteralExpression().Properties.Nodes, false)
				case ast.KindIdentifier:
					if decl := declarationObject(ctx.SourceFile, ctx.Refs.Resolve(firstArg)); decl != nil {
						checkSorted(decl.AsObjectLiteralExpression().Properties.Nodes, false)
					}
				}
			}
		}
		if opts.checkTypes {
			typeAliases = make(map[string]*ast.Node)
			listeners[ast.KindTypeAliasDeclaration] = func(node *ast.Node) {
				decl := node.AsTypeAliasDeclaration()
				if decl.Name() == nil || decl.Name().Kind != ast.KindIdentifier {
					return
				}
				typeNode := unwrapParenthesizedType(decl.Type)
				if typeNode == nil || typeNode.Kind != ast.KindTypeLiteral {
					return
				}
				name := decl.Name().AsIdentifier().Text
				if _, exists := typeAliases[name]; !exists {
					typeAliases[name] = typeNode
				}
			}
			checkFunction := func(node *ast.Node) {
				if node.Kind == ast.KindFunctionDeclaration && reactutil.FunctionBody(node) == nil {
					return
				}
				typeNode := firstParamType(node)
				checkTypeNode(typeNode, typeAliases, checkSorted)
			}
			listeners[ast.KindFunctionDeclaration] = checkFunction
			listeners[ast.KindArrowFunction] = checkFunction
		}
		return listeners
	},
}

func declarationName(sourceFile *ast.SourceFile, node *ast.Node, typeMembers bool) (propName, bool) {
	if node == nil || node.Name() == nil {
		if typeMembers && sourceFile != nil {
			// typescript-eslint's getKey() falls back to getText(undefined) for
			// nameless type members, which is the complete source text. Keep that
			// observable value so call/index/construct signatures remain part of
			// the same reducer chain.
			return propName{text: sourceFile.Text(), kind: propNameString}, true
		}
		return propName{}, false
	}
	return propNameFromNode(sourceFile, node.Name()), true
}

func propNameFromNode(sourceFile *ast.SourceFile, node *ast.Node) propName {
	if node == nil {
		return propName{}
	}
	if node.Kind == ast.KindComputedPropertyName {
		return propNameFromNode(sourceFile, utils.ESTreeRuntimeExpression(node.AsComputedPropertyName().Expression))
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return propName{text: utils.TrimmedNodeText(sourceFile, node), kind: propNameString}
	case ast.KindStringLiteral:
		value := node.AsStringLiteral().Text
		if value != "" {
			return propName{text: value, kind: propNameString}
		}
		return propName{text: utils.TrimmedNodeText(sourceFile, node), kind: propNameString}
	case ast.KindNumericLiteral:
		valueText, ok := utils.GetStaticPropertyName(node)
		if ok {
			if value, valueOK := ecmascript.StringToNumber(valueText); valueOK && value != 0 {
				return propName{number: value, kind: propNameNumber}
			}
		}
		return propName{text: utils.TrimmedNodeText(sourceFile, node), kind: propNameString}
	case ast.KindBigIntLiteral:
		value := new(big.Int)
		if normalized := utils.NormalizeBigIntLiteral(node.AsBigIntLiteral().Text); normalized != "" {
			if _, ok := value.SetString(normalized, 10); ok && value.Sign() != 0 {
				return propName{bigInt: value, kind: propNameBigInt}
			}
		}
		return propName{text: utils.TrimmedNodeText(sourceFile, node), kind: propNameString}
	case ast.KindRegularExpressionLiteral:
		return propName{text: utils.RegExpLiteralStringValue(node.AsRegularExpressionLiteral().Text), kind: propNameString}
	case ast.KindTrueKeyword:
		return propName{number: 1, kind: propNameBoolean}
	default:
		return propName{text: utils.TrimmedNodeText(sourceFile, node), kind: propNameString}
	}
}

func propNameToString(name propName) string {
	switch name.kind {
	case propNameNumber:
		return ecmascript.NumberToString(name.number)
	case propNameBigInt:
		if name.bigInt != nil {
			return name.bigInt.String()
		}
	case propNameBoolean:
		return "true"
	}
	return name.text
}

func isPropNameBefore(current, previous propName) bool {
	if current.kind == propNameBigInt || previous.kind == propNameBigInt {
		return isBigIntPropNameBefore(current, previous)
	}
	if current.kind == propNameString && previous.kind == propNameString {
		return ecmascript.CompareStrings(current.text, previous.text) < 0
	}
	currentNumber, currentOK := propNameNumberValue(current)
	previousNumber, previousOK := propNameNumberValue(previous)
	return currentOK && previousOK && currentNumber < previousNumber
}

func isBigIntPropNameBefore(current, previous propName) bool {
	if current.kind == propNameBigInt && previous.kind == propNameBigInt {
		return current.bigInt.Cmp(previous.bigInt) < 0
	}
	if current.kind == propNameBigInt && previous.kind == propNameString {
		value, ok := ecmascript.StringToBigInt(previous.text)
		return ok && current.bigInt.Cmp(value) < 0
	}
	if previous.kind == propNameBigInt && current.kind == propNameString {
		value, ok := ecmascript.StringToBigInt(current.text)
		return ok && value.Cmp(previous.bigInt) < 0
	}
	currentNumber, currentOK := propNameNumberValue(current)
	previousNumber, previousOK := propNameNumberValue(previous)
	if current.kind == propNameBigInt && previousOK {
		return compareBigIntAndNumber(current.bigInt, previousNumber) < 0
	}
	if previous.kind == propNameBigInt && currentOK {
		return compareBigIntAndNumber(previous.bigInt, currentNumber) > 0
	}
	return false
}

func propNameNumberValue(name propName) (float64, bool) {
	switch name.kind {
	case propNameNumber, propNameBoolean:
		return name.number, true
	case propNameString:
		return ecmascript.StringToNumber(name.text)
	default:
		return 0, false
	}
}

func compareBigIntAndNumber(value *big.Int, number float64) int {
	if value == nil {
		return 0
	}
	if number != number {
		return 0
	}
	if math.IsInf(number, 1) {
		return -1
	}
	if math.IsInf(number, -1) {
		return 1
	}
	return new(big.Rat).SetInt(value).Cmp(new(big.Rat).SetFloat64(number))
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
	value := utils.ESTreeRuntimeExpression(node.AsPropertyAssignment().Initializer)
	if value == nil || ast.IsOptionalChain(value) {
		return false
	}
	switch value.Kind {
	case ast.KindPropertyAccessExpression:
		return identifierOrPrivateName(value.AsPropertyAccessExpression().Name()) == "isRequired"
	case ast.KindElementAccessExpression:
		argument := utils.ESTreeRuntimeExpression(value.AsElementAccessExpression().ArgumentExpression)
		return argument != nil && argument.Kind == ast.KindIdentifier && argument.AsIdentifier().Text == "isRequired"
	default:
		return false
	}
}

func isCallback(name string) bool {
	return len(name) >= 3 && strings.HasPrefix(name, "on") && name[2] >= 'A' && name[2] <= 'Z'
}

func isShapeCall(callee *ast.Node) bool {
	if callee == nil {
		return false
	}
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		return identifierOrPrivateName(callee.AsPropertyAccessExpression().Name()) == "shape"
	case ast.KindElementAccessExpression:
		argument := utils.ESTreeRuntimeExpression(callee.AsElementAccessExpression().ArgumentExpression)
		return argument != nil && argument.Kind == ast.KindIdentifier && argument.AsIdentifier().Text == "shape"
	default:
		return false
	}
}

func isAuthoredPropertyName(name *ast.Node, expected string) bool {
	if name == nil {
		return false
	}
	if identifierOrPrivateName(name) == expected {
		return true
	}
	if name.Kind == ast.KindComputedPropertyName {
		expression := utils.ESTreeRuntimeExpression(name.AsComputedPropertyName().Expression)
		return identifierOrPrivateName(expression) == expected
	}
	return false
}

func identifierOrPrivateName(node *ast.Node) string {
	return reactutil.IdentifierOrPrivateName(node)
}

func isPropTypesAccess(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || ast.IsOptionalChain(node) {
		return false
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return identifierOrPrivateName(node.AsPropertyAccessExpression().Name()) == "propTypes"
	case ast.KindElementAccessExpression:
		argument := utils.ESTreeRuntimeExpression(node.AsElementAccessExpression().ArgumentExpression)
		return argument != nil && argument.Kind == ast.KindIdentifier && argument.AsIdentifier().Text == "propTypes"
	default:
		return false
	}
}

func hasModifier(modifiers *ast.ModifierList, kind ast.Kind) bool {
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier != nil && modifier.Kind == kind {
			return true
		}
	}
	return false
}

func isPropWrapperCallByCalleeName(call *ast.Node, wrappers []reactutil.PropWrapperEntry) bool {
	if call == nil || call.Kind != ast.KindCallExpression || ast.IsOptionalChain(call) {
		return false
	}
	callee := utils.ESTreeCallCallee(call.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier || ast.IsOptionalChain(callee) {
		return false
	}
	name := callee.AsIdentifier().Text
	for _, wrapper := range wrappers {
		if (wrapper.FromString && wrapper.Raw == name) || (!wrapper.FromString && wrapper.Property == name) {
			return true
		}
	}
	return false
}

func checkTypeNode(typeNode *ast.Node, aliases map[string]*ast.Node, check func([]*ast.Node, bool)) {
	typeNode = unwrapParenthesizedType(typeNode)
	if typeNode == nil {
		return
	}
	switch typeNode.Kind {
	case ast.KindTypeLiteral:
		check(typeNode.AsTypeLiteralNode().Members.Nodes, true)
	case ast.KindTypeReference:
		name := typeNode.AsTypeReferenceNode().TypeName
		if name != nil && name.Kind == ast.KindIdentifier {
			if alias := aliases[name.AsIdentifier().Text]; alias != nil && alias.Kind == ast.KindTypeLiteral {
				check(alias.AsTypeLiteralNode().Members.Nodes, true)
			}
		}
	}
}

func declarationObject(sourceFile *ast.SourceFile, symbol *ast.Symbol) *ast.Node {
	if sourceFile == nil || symbol == nil || symbol.ValueDeclaration == nil {
		return nil
	}
	decl := symbol.ValueDeclaration
	switch decl.Kind {
	case ast.KindBindingElement:
		decl = utils.EnclosingVariableDeclarationOfBindingElement(decl)
	case ast.KindVariableDeclaration:
	default:
		return nil
	}
	if decl == nil || decl.Kind != ast.KindVariableDeclaration || ast.GetSourceFileOfNode(decl) != sourceFile {
		return nil
	}
	variable := decl.AsVariableDeclaration()
	if variable == nil || variable.Initializer == nil {
		return nil
	}
	value := utils.ESTreeRuntimeExpression(variable.Initializer)
	if value != nil && value.Kind == ast.KindObjectLiteralExpression {
		return value
	}
	return nil
}

func unwrapParenthesizedType(typeNode *ast.Node) *ast.Node {
	for typeNode != nil && typeNode.Kind == ast.KindParenthesizedType {
		parenthesized := typeNode.AsParenthesizedTypeNode()
		if parenthesized == nil {
			return nil
		}
		typeNode = parenthesized.Type
	}
	return typeNode
}
