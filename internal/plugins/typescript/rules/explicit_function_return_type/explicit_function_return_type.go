package explicit_function_return_type

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ExplicitFunctionReturnTypeOptions struct {
	AllowConciseArrowFunctionExpressionsStartingWithVoid bool
	AllowDirectConstAssertionInArrowFunctions            bool
	AllowedNames                                         []string
	AllowExpressions                                     bool
	AllowFunctionsWithoutTypeParameters                  bool
	AllowHigherOrderFunctions                            bool
	AllowIIFEs                                           bool
	AllowTypedFunctionExpressions                        bool
}

type functionInfo struct {
	node    *ast.Node
	returns []*ast.Node
}

func buildMissingReturnTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingReturnType",
		Description: "Missing return type on function.",
	}
}

func parseOptions(options any) ExplicitFunctionReturnTypeOptions {
	opts := ExplicitFunctionReturnTypeOptions{
		AllowConciseArrowFunctionExpressionsStartingWithVoid: false,
		AllowDirectConstAssertionInArrowFunctions:            true,
		AllowedNames:                        []string{},
		AllowExpressions:                    false,
		AllowFunctionsWithoutTypeParameters: false,
		AllowHigherOrderFunctions:           true,
		AllowIIFEs:                          false,
		AllowTypedFunctionExpressions:       true,
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

	if v, ok := optsMap["allowConciseArrowFunctionExpressionsStartingWithVoid"].(bool); ok {
		opts.AllowConciseArrowFunctionExpressionsStartingWithVoid = v
	}
	if v, ok := optsMap["allowDirectConstAssertionInArrowFunctions"].(bool); ok {
		opts.AllowDirectConstAssertionInArrowFunctions = v
	}
	if v, ok := optsMap["allowedNames"].([]interface{}); ok {
		opts.AllowedNames = make([]string, 0, len(v))
		for _, name := range v {
			if str, ok := name.(string); ok {
				opts.AllowedNames = append(opts.AllowedNames, str)
			}
		}
	}
	if v, ok := optsMap["allowExpressions"].(bool); ok {
		opts.AllowExpressions = v
	}
	if v, ok := optsMap["allowFunctionsWithoutTypeParameters"].(bool); ok {
		opts.AllowFunctionsWithoutTypeParameters = v
	}
	if v, ok := optsMap["allowHigherOrderFunctions"].(bool); ok {
		opts.AllowHigherOrderFunctions = v
	}
	if v, ok := optsMap["allowIIFEs"].(bool); ok {
		opts.AllowIIFEs = v
	}
	if v, ok := optsMap["allowTypedFunctionExpressions"].(bool); ok {
		opts.AllowTypedFunctionExpressions = v
	}

	return opts
}

func getParentSkippingParens(node *ast.Node) *ast.Node {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent
}

func hasTypeParameters(node *ast.Node) bool {
	if node == nil {
		return false
	}
	typeParams := node.TypeParameters()
	return len(typeParams) > 0
}

func getSimpleIdentifierName(name *ast.Node) string {
	if name == nil {
		return ""
	}
	if name.Kind == ast.KindComputedPropertyName {
		return ""
	}
	if ast.IsIdentifier(name) {
		identifier := name.AsIdentifier()
		if identifier != nil {
			return identifier.Text
		}
	}
	return ""
}

func getFunctionName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	if name := getSimpleIdentifierName(node.Name()); name != "" {
		return name
	}

	parent := getParentSkippingParens(node)
	if parent == nil {
		return ""
	}

	switch parent.Kind {
	case ast.KindVariableDeclaration:
		return getSimpleIdentifierName(parent.Name())
	case ast.KindPropertyDeclaration:
		return getSimpleIdentifierName(parent.Name())
	case ast.KindPropertyAssignment:
		return getSimpleIdentifierName(parent.Name())
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return getSimpleIdentifierName(parent.Name())
	}

	return ""
}

func isIIFE(node *ast.Node) bool {
	if node == nil {
		return false
	}

	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		callParent := parent.Parent
		if ast.IsCallExpression(callParent) && callParent.Expression() == parent {
			return true
		}
		parent = parent.Parent
	}

	if ast.IsCallExpression(parent) {
		return parent.Expression() == node
	}

	return false
}

func isAllowedFunction(node *ast.Node, opts ExplicitFunctionReturnTypeOptions, allowedNames map[string]struct{}) bool {
	if opts.AllowFunctionsWithoutTypeParameters && !hasTypeParameters(node) {
		return true
	}

	if opts.AllowIIFEs && isIIFE(node) {
		return true
	}

	if len(allowedNames) == 0 {
		return false
	}

	if name := getFunctionName(node); name != "" {
		if _, ok := allowedNames[name]; ok {
			return true
		}
	}

	return false
}

func isDefaultParameterWithTypeAnnotation(parent *ast.Node) bool {
	if parent == nil || parent.Kind != ast.KindParameter {
		return false
	}
	return parent.Type() != nil
}

func isVariableDeclarationWithTypeAnnotation(parent *ast.Node) bool {
	if parent == nil || parent.Kind != ast.KindVariableDeclaration {
		return false
	}
	return parent.Type() != nil
}

func isPropertyDeclarationWithTypeAnnotation(parent *ast.Node) bool {
	if parent == nil || parent.Kind != ast.KindPropertyDeclaration {
		return false
	}
	return parent.Type() != nil
}

func isFunctionArgument(parent *ast.Node, callee *ast.Node) bool {
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	if callee != nil && ast.SkipParentheses(parent.Expression()) == callee {
		return false
	}
	return true
}

func isTypedJSX(parent *ast.Node) bool {
	if parent == nil {
		return false
	}
	return ast.IsJsxExpression(parent) || ast.IsJsxSpreadAttribute(parent)
}

func isTypedParent(parent *ast.Node, callee *ast.Node) bool {
	if parent == nil {
		return false
	}

	if parent.Kind == ast.KindAsExpression || parent.Kind == ast.KindTypeAssertionExpression {
		return true
	}

	if isVariableDeclarationWithTypeAnnotation(parent) {
		return true
	}

	if isDefaultParameterWithTypeAnnotation(parent) {
		return true
	}

	if isPropertyDeclarationWithTypeAnnotation(parent) {
		return true
	}

	if isFunctionArgument(parent, callee) {
		return true
	}

	return isTypedJSX(parent)
}

func isConstructorArgument(parent *ast.Node) bool {
	return parent != nil && parent.Kind == ast.KindNewExpression
}

func isPropertyOfObjectWithType(property *ast.Node) bool {
	if property == nil {
		return false
	}

	if property.Kind != ast.KindPropertyAssignment &&
		property.Kind != ast.KindMethodDeclaration &&
		property.Kind != ast.KindGetAccessor &&
		property.Kind != ast.KindSetAccessor {
		return false
	}

	objectExpr := property.Parent
	if objectExpr == nil || objectExpr.Kind != ast.KindObjectLiteralExpression {
		return false
	}

	parent := getParentSkippingParens(objectExpr)
	return isTypedParent(parent, nil) || isPropertyOfObjectWithType(parent)
}

func isTypedFunctionExpression(node *ast.Node, opts ExplicitFunctionReturnTypeOptions) bool {
	if !opts.AllowTypedFunctionExpressions {
		return false
	}

	parent := getParentSkippingParens(node)
	if parent != nil && isTypedParent(parent, node) {
		return true
	}

	if isPropertyOfObjectWithType(parent) {
		return true
	}

	if (node.Kind == ast.KindMethodDeclaration ||
		node.Kind == ast.KindGetAccessor ||
		node.Kind == ast.KindSetAccessor) &&
		node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
		return isPropertyOfObjectWithType(node)
	}

	return isConstructorArgument(parent)
}

func isConstAssertion(node *ast.Node) bool {
	if node == nil {
		return false
	}

	var typeNode *ast.Node
	switch node.Kind {
	case ast.KindAsExpression:
		typeNode = node.AsAsExpression().Type
	case ast.KindTypeAssertionExpression:
		typeNode = node.AsTypeAssertion().Type
	default:
		return false
	}

	if typeNode == nil || typeNode.Kind != ast.KindTypeReference {
		return false
	}

	typeRef := typeNode.AsTypeReference()
	if typeRef == nil || typeRef.TypeName == nil || !ast.IsIdentifier(typeRef.TypeName) {
		return false
	}

	return typeRef.TypeName.AsIdentifier().Text == "const"
}

func isValidFunctionExpressionReturnType(node *ast.Node, opts ExplicitFunctionReturnTypeOptions) bool {
	if isTypedFunctionExpression(node, opts) {
		return true
	}

	parent := getParentSkippingParens(node)
	if opts.AllowExpressions && parent != nil {
		isClassParent := parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindClassExpression
		isClassMember := isClassParent && (node.Kind == ast.KindMethodDeclaration ||
			node.Kind == ast.KindGetAccessor ||
			node.Kind == ast.KindSetAccessor ||
			node.Kind == ast.KindConstructor)
		isDisallowedParent := parent.Kind == ast.KindVariableDeclaration ||
			parent.Kind == ast.KindPropertyDeclaration ||
			parent.Kind == ast.KindExportAssignment ||
			isClassMember
		if !isDisallowedParent {
			return true
		}
	}

	if !opts.AllowDirectConstAssertionInArrowFunctions || node.Kind != ast.KindArrowFunction {
		return false
	}

	body := ast.SkipParentheses(node.Body())
	for body != nil && body.Kind == ast.KindSatisfiesExpression {
		body = body.AsSatisfiesExpression().Expression
		body = ast.SkipParentheses(body)
	}

	return isConstAssertion(body)
}

func doesImmediatelyReturnFunctionExpression(info *functionInfo) bool {
	if info == nil || info.node == nil {
		return false
	}

	if info.node.Kind == ast.KindArrowFunction {
		body := ast.SkipParentheses(info.node.Body())
		if ast.IsArrowFunction(body) || ast.IsFunctionExpression(body) {
			return true
		}
	}

	if len(info.returns) == 0 {
		return false
	}

	for _, ret := range info.returns {
		argument := ret.Expression()
		if argument == nil {
			return false
		}
		argument = ast.SkipParentheses(argument)
		if !ast.IsArrowFunction(argument) && !ast.IsFunctionExpression(argument) {
			return false
		}
	}

	return true
}

func isSetter(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindSetAccessor
}

func isConstructor(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindConstructor
}

func checkFunctionReturnType(ctx rule.RuleContext, info *functionInfo, opts ExplicitFunctionReturnTypeOptions) {
	if info == nil || info.node == nil {
		return
	}

	if opts.AllowHigherOrderFunctions && doesImmediatelyReturnFunctionExpression(info) {
		return
	}

	if info.node.Type() != nil {
		return
	}

	if isConstructor(info.node) || isSetter(info.node) {
		return
	}

	ctx.ReportRange(getFunctionHeadRange(ctx.SourceFile, info.node), buildMissingReturnTypeMessage())
}

func getStartAfterDecorators(sourceFile *ast.SourceFile, node *ast.Node) int {
	start := node.Pos()
	modifiers := node.Modifiers()
	if modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier.Kind == ast.KindDecorator && modifier.End() > start {
				start = modifier.End()
			}
		}
	}

	return getTokenStart(sourceFile, start)
}

func getTokenStart(sourceFile *ast.SourceFile, pos int) int {
	s := scanner.GetScannerForSourceFile(sourceFile, pos)
	if s.TokenStart() < pos {
		s.Scan()
	}
	return s.TokenStart()
}

func getFunctionDeclarationStart(sourceFile *ast.SourceFile, node *ast.Node) int {
	s := scanner.GetScannerForSourceFile(sourceFile, node.Pos())
	for s.TokenStart() < node.End() {
		if s.Token() == ast.KindAsyncKeyword || s.Token() == ast.KindFunctionKeyword {
			return s.TokenStart()
		}
		s.Scan()
	}
	return getTokenStart(sourceFile, node.Pos())
}

func getOpeningParenPos(sourceFile *ast.SourceFile, node *ast.Node) int {
	if node == nil {
		return 0
	}

	if node.Kind == ast.KindArrowFunction && utils.IsParenlessArrowFunction(node) {
		arrow := node.AsArrowFunction()
		if arrow != nil && arrow.Parameters != nil && len(arrow.Parameters.Nodes) > 0 {
			param := arrow.Parameters.Nodes[0]
			tokenRange := scanner.GetRangeOfTokenAtPosition(sourceFile, param.Pos())
			return tokenRange.Pos()
		}
	}

	scanStart := node.Pos()
	if name := node.Name(); name != nil {
		scanStart = name.End()
	} else {
		scanStart = getTokenStart(sourceFile, scanStart)
	}

	s := scanner.GetScannerForSourceFile(sourceFile, scanStart)
	for s.TokenStart() < node.End() {
		if s.Token() == ast.KindOpenParenToken {
			return s.TokenStart()
		}
		s.Scan()
	}

	return scanStart
}

func getArrowTokenRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	body := node.Body()
	if body == nil {
		return utils.TrimNodeTextRange(sourceFile, node)
	}

	endPos := body.Pos()
	s := scanner.GetScannerForSourceFile(sourceFile, node.Pos())
	arrowRange := core.NewTextRange(node.Pos(), node.Pos())
	for s.TokenStart() < endPos {
		if s.Token() == ast.KindEqualsGreaterThanToken {
			arrowRange = s.TokenRange()
		}
		s.Scan()
	}

	return arrowRange
}

func getFunctionHeadRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	parent := getParentSkippingParens(node)

	if parent != nil && (parent.Kind == ast.KindMethodDeclaration || parent.Kind == ast.KindPropertyDeclaration) {
		start := getStartAfterDecorators(sourceFile, parent)
		end := getOpeningParenPos(sourceFile, node)
		return core.NewTextRange(start, end)
	}

	if parent != nil && parent.Kind == ast.KindPropertyAssignment {
		start := getTokenStart(sourceFile, parent.Pos())
		end := getOpeningParenPos(sourceFile, node)
		return core.NewTextRange(start, end)
	}

	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		start := getStartAfterDecorators(sourceFile, node)
		end := getOpeningParenPos(sourceFile, node)
		return core.NewTextRange(start, end)
	case ast.KindArrowFunction:
		return getArrowTokenRange(sourceFile, node)
	default:
		start := getTokenStart(sourceFile, node.Pos())
		if node.Kind == ast.KindFunctionDeclaration {
			start = getFunctionDeclarationStart(sourceFile, node)
		}
		end := getOpeningParenPos(sourceFile, node)
		return core.NewTextRange(start, end)
	}
}

var ExplicitFunctionReturnTypeRule = rule.CreateRule(rule.Rule{
	Name: "explicit-function-return-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		allowedNames := make(map[string]struct{}, len(opts.AllowedNames))
		for _, name := range opts.AllowedNames {
			allowedNames[name] = struct{}{}
		}

		functionInfoStack := make([]*functionInfo, 0, 4)

		enterFunction := func(node *ast.Node) {
			functionInfoStack = append(functionInfoStack, &functionInfo{
				node:    node,
				returns: []*ast.Node{},
			})
		}

		popFunctionInfo := func() *functionInfo {
			if len(functionInfoStack) == 0 {
				return nil
			}
			info := functionInfoStack[len(functionInfoStack)-1]
			functionInfoStack = functionInfoStack[:len(functionInfoStack)-1]
			return info
		}

		exitFunctionExpression := func(node *ast.Node) {
			info := popFunctionInfo()
			if info == nil {
				return
			}

			if opts.AllowConciseArrowFunctionExpressionsStartingWithVoid &&
				node.Kind == ast.KindArrowFunction &&
				node.Body() != nil &&
				ast.IsVoidExpression(node.Body()) {
				return
			}

			if isAllowedFunction(node, opts, allowedNames) {
				return
			}

			if opts.AllowTypedFunctionExpressions &&
				(isValidFunctionExpressionReturnType(node, opts) || ancestorHasReturnType(node)) {
				return
			}

			checkFunctionReturnType(ctx, info, opts)
		}

		exitFunctionDeclaration := func(node *ast.Node) {
			info := popFunctionInfo()
			if info == nil {
				return
			}

			if isAllowedFunction(node, opts, allowedNames) {
				return
			}

			if opts.AllowTypedFunctionExpressions && node.Type() != nil {
				return
			}

			checkFunctionReturnType(ctx, info, opts)
		}

		return rule.RuleListeners{
			ast.KindArrowFunction:       enterFunction,
			ast.KindFunctionExpression:  enterFunction,
			ast.KindFunctionDeclaration: enterFunction,
			ast.KindMethodDeclaration:   enterFunction,
			ast.KindGetAccessor:         enterFunction,
			ast.KindSetAccessor:         enterFunction,
			ast.KindConstructor:         enterFunction,
			ast.KindReturnStatement: func(node *ast.Node) {
				if len(functionInfoStack) == 0 {
					return
				}
				functionInfoStack[len(functionInfoStack)-1].returns = append(functionInfoStack[len(functionInfoStack)-1].returns, node)
			},
			rule.ListenerOnExit(ast.KindArrowFunction):       exitFunctionExpression,
			rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunctionExpression,
			rule.ListenerOnExit(ast.KindMethodDeclaration):   exitFunctionExpression,
			rule.ListenerOnExit(ast.KindGetAccessor):         exitFunctionExpression,
			rule.ListenerOnExit(ast.KindSetAccessor):         exitFunctionExpression,
			rule.ListenerOnExit(ast.KindConstructor):         exitFunctionExpression,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunctionDeclaration,
		}
	},
})

func ancestorHasReturnType(node *ast.Node) bool {
	ancestor := getParentSkippingParens(node)
	if ancestor == nil {
		return false
	}

	if ancestor.Kind == ast.KindPropertyAssignment {
		ancestor = ancestor.AsPropertyAssignment().Initializer
	}

	isReturnStatement := ancestor.Kind == ast.KindReturnStatement
	isBodylessArrow := ancestor.Kind == ast.KindArrowFunction && ancestor.Body() != nil && ancestor.Body().Kind != ast.KindBlock
	if !isReturnStatement && !isBodylessArrow {
		return false
	}

	for ancestor != nil {
		if ancestor.Kind == ast.KindParenthesizedExpression {
			ancestor = ancestor.Parent
			continue
		}

		switch ancestor.Kind {
		case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindFunctionDeclaration, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
			if ancestor.Type() != nil {
				return true
			}
		case ast.KindVariableDeclaration:
			return ancestor.Type() != nil
		case ast.KindPropertyDeclaration:
			return ancestor.Type() != nil
		case ast.KindExpressionStatement:
			return false
		}

		ancestor = ancestor.Parent
	}

	return false
}
