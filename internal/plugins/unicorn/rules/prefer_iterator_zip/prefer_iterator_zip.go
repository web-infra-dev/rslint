package prefer_iterator_zip

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

const ruleName = "prefer-iterator-zip"

var PreferIteratorZipRule = rule.Rule{
	Name:   "unicorn/" + ruleName,
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{rule.ListenerOnExit(ast.KindForStatement): func(node *ast.Node) {
			checkLoop(ctx, node)
		}}
	},
}

func checkLoop(ctx rule.RuleContext, node *ast.Node) {
	loop := node.AsForStatement()
	if loop.Statement == nil || !ast.IsBlock(loop.Statement) || loop.Initializer == nil ||
		!ast.IsVariableDeclarationList(loop.Initializer) || !ast.IsLet(loop.Initializer) {
		return
	}
	declarations := loop.Initializer.AsVariableDeclarationList().Declarations.Nodes
	if len(declarations) == 0 || len(declarations) > 2 {
		return
	}
	index := declarations[0].Name()
	if !isIdentifier(index) || !isIntegerLiteral(declarations[0].Initializer(), 0) ||
		!incrementsByOne(loop.Incrementor, index.Text()) {
		return
	}

	bounds := matchBounds(ctx, node, declarations, index.Text())
	if len(bounds.inputs) == 0 {
		return
	}
	for _, input := range bounds.inputs {
		// In JavaScript, classify from the source even when a checker is available.
		if hasNonIndexedJSDocType(ctx, input) || unicornutil.IsKnownNonIndexedCollectionWithOptions(ctx, input, unicornutil.ArrayReceiverOptions{}) {
			return
		}
	}
	reads := collectReads(ctx, loop.Statement, declarations[0].Symbol(), bounds.inputs)
	if reads == nil {
		return
	}

	closingParen, ok := utils.TokenAtOrAfter(ctx.SourceFile, loop.Incrementor.End())
	if !ok || closingParen.Kind != ast.KindCloseParenToken {
		return
	}
	header := core.NewTextRange(utils.TrimNodeTextRange(ctx.SourceFile, node).Pos(), closingParen.End)
	ctx.ReportRangeWithDeferredSuggestions(header, rule.RuleMessage{
		Id:          ruleName,
		Description: "Use `Iterator.zip(…)` to iterate over these arrays together.",
	}, func() []rule.RuleSuggestion {
		return buildSuggestion(ctx, header, node, bounds, reads)
	})
}

var jsdocInputTypes = unicornutil.TypeClassifierOptions{
	TargetTypeNames:          utils.NewSetFromItems("Array", "ReadonlyArray"),
	NonTargetTypeNames:       utils.NewSetFromItems("String"),
	AllowNullishInMixedUnion: true,
}

func hasNonIndexedJSDocType(ctx rule.RuleContext, input *ast.Node) bool {
	if !ast.IsInJSFile(input) {
		return false
	}
	seen := make(map[*ast.Symbol]bool)
	for isIdentifier(input) {
		symbol := ctx.Refs.ResolveInFile(input)
		if symbol == nil || len(symbol.Declarations) != 1 || seen[symbol] {
			return false
		}
		declaration := symbol.Declarations[0]
		if !ast.IsParameterDeclaration(declaration) && !ast.IsVariableDeclaration(declaration) {
			return false
		}
		if annotation, _ := unicornutil.ValueTypeAnnotation(ctx, declaration); annotation != nil {
			return isNonIndexedJSDocAnnotation(ctx, annotation)
		}
		if !ast.IsVariableDeclaration(declaration) || !ast.IsVarConst(declaration) {
			return false
		}
		seen[symbol] = true
		input = ast.SkipParentheses(declaration.Initializer())
	}
	return input != nil && ast.IsCallExpression(input) &&
		isNonIndexedJSDocAnnotation(ctx, localCallReturnAnnotation(ctx, input))
}

func isNonIndexedJSDocAnnotation(ctx rule.RuleContext, annotation *ast.Node) bool {
	if annotation == nil {
		return false
	}
	class := unicornutil.ClassifyTypeAnnotation(ctx, annotation, jsdocInputTypes)
	return class == unicornutil.TypeNonTarget || class == unicornutil.TypeNullish
}

func localCallReturnAnnotation(ctx rule.RuleContext, call *ast.Node) *ast.Node {
	callee := utils.ESTreeCallCallee(call.Expression())
	if !isIdentifier(callee) || ast.IsOptionalChain(call) {
		return nil
	}
	symbol := ctx.Refs.ResolveInFile(callee)
	if symbol == nil || len(symbol.Declarations) != 1 || slices.ContainsFunc(ctx.Refs.References(symbol), utils.IsWriteReference) {
		return nil
	}
	declaration := symbol.Declarations[0]
	var function *ast.Node
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		function = declaration
	case ast.KindVariableDeclaration, ast.KindParameter:
		initializer := utils.ESTreeRuntimeExpression(declaration.Initializer())
		if initializer != nil && ast.IsFunctionExpressionOrArrowFunction(initializer) {
			function = initializer
		}
	default:
		return nil
	}
	if function != nil && (ast.GetFunctionFlags(function)&ast.FunctionFlagsAsyncGenerator != 0 || len(function.TypeParameters()) != 0) {
		return nil
	}
	annotation := unicornutil.CallableReturnTypeAnnotation(ctx, declaration)
	if annotation == nil && function != nil && ast.IsVariableDeclaration(declaration) {
		if binding, _ := unicornutil.ValueTypeAnnotation(ctx, declaration); binding == nil {
			return unicornutil.CallableReturnTypeAnnotation(ctx, function)
		}
	}
	return annotation
}

type loopBounds struct {
	inputs          []*ast.Node
	mathObject      *ast.Node
	cachedStatement *ast.Node
}

func matchBounds(ctx rule.RuleContext, node *ast.Node, declarations []*ast.Node, index string) loopBounds {
	test := utils.ESTreeRuntimeExpression(node.AsForStatement().Condition)
	if len(declarations) == 1 && test != nil && ast.IsBinaryExpression(test) &&
		test.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
		return loopBounds{inputs: conjunctiveInputs(test, index)}
	}
	bound := upperBound(test, index)
	if len(declarations) == 1 {
		if bounds := mathMinBounds(bound); bounds.inputs != nil {
			return bounds
		}
	}
	if !isIdentifier(bound) {
		return loopBounds{}
	}
	if len(declarations) == 2 {
		return cachedBounds(ctx, bound, declarations[1])
	}

	if node.Parent == nil || !node.Parent.CanHaveStatements() {
		return loopBounds{}
	}
	siblings := node.Parent.Statements()
	position := slices.Index(siblings, node)
	if position <= 0 {
		return loopBounds{}
	}
	previous := siblings[position-1]
	if !ast.IsVariableStatement(previous) || ast.HasSyntacticModifier(previous, ast.ModifierFlagsExport) {
		return loopBounds{}
	}
	list := previous.AsVariableStatement().DeclarationList
	if !ast.IsLet(list) && !ast.IsVarConst(list) {
		return loopBounds{}
	}
	cached := list.AsVariableDeclarationList().Declarations.Nodes
	if len(cached) != 1 {
		return loopBounds{}
	}
	bounds := cachedBounds(ctx, bound, cached[0])
	bounds.cachedStatement = previous
	return bounds
}

func cachedBounds(ctx rule.RuleContext, bound, declaration *ast.Node) loopBounds {
	if !namedIdentifier(declaration.Name(), bound.Text()) {
		return loopBounds{}
	}
	bounds := mathMinBounds(declaration.Initializer())
	if bounds.inputs == nil {
		return loopBounds{}
	}
	refs := ctx.Refs.References(declaration.Symbol())
	if len(refs) != 1 || refs[0] != bound {
		return loopBounds{}
	}
	return bounds
}

func incrementsByOne(node *ast.Node, index string) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindPrefixUnaryExpression:
		update := node.AsPrefixUnaryExpression()
		return update.Operator == ast.KindPlusPlusToken && namedIdentifier(update.Operand, index)
	case ast.KindPostfixUnaryExpression:
		update := node.AsPostfixUnaryExpression()
		return update.Operator == ast.KindPlusPlusToken && namedIdentifier(update.Operand, index)
	case ast.KindBinaryExpression:
		update := node.AsBinaryExpression()
		if !namedIdentifier(update.Left, index) {
			return false
		}
		switch update.OperatorToken.Kind {
		case ast.KindPlusEqualsToken:
			return isIntegerLiteral(update.Right, 1)
		case ast.KindEqualsToken:
			addition := utils.ESTreeRuntimeExpression(update.Right)
			if addition != nil && ast.IsBinaryExpression(addition) {
				add := addition.AsBinaryExpression()
				return add.OperatorToken.Kind == ast.KindPlusToken &&
					namedIdentifier(add.Left, index) && isIntegerLiteral(add.Right, 1)
			}
		}
	}
	return false
}

func isIntegerLiteral(node *ast.Node, value float64) bool {
	number, ok := unicornutil.SafeIntegerLiteral(node)
	return ok && number == value
}

func isIdentifier(node *ast.Node) bool {
	return node != nil && ast.IsIdentifier(node)
}

func namedIdentifier(node *ast.Node, name string) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return isIdentifier(node) && node.Text() == name
}

func upperBound(node *ast.Node, index string) *ast.Node {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || !ast.IsBinaryExpression(node) {
		return nil
	}
	comparison := node.AsBinaryExpression()
	if comparison.OperatorToken.Kind == ast.KindLessThanToken && namedIdentifier(comparison.Left, index) {
		return utils.ESTreeRuntimeExpression(comparison.Right)
	}
	if comparison.OperatorToken.Kind == ast.KindGreaterThanToken && namedIdentifier(comparison.Right, index) {
		return utils.ESTreeRuntimeExpression(comparison.Left)
	}
	return nil
}

func lengthInput(node *ast.Node) *ast.Node {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || !ast.IsPropertyAccessExpression(node) || ast.IsOptionalChain(node) ||
		!namedIdentifier(node.Name(), "length") {
		return nil
	}
	if object := utils.ESTreeRuntimeExpression(node.Expression()); isIdentifier(object) {
		return object
	}
	return nil
}

func mathMinBounds(node *ast.Node) loopBounds {
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{Method: "min", MinimumArguments: new(2)})
	if !ok {
		return loopBounds{}
	}
	object := utils.ESTreeRuntimeExpression(call.Object)
	if !isIdentifier(object) || object.Text() != "Math" {
		return loopBounds{}
	}
	args := call.Call.Arguments()
	inputs := make([]*ast.Node, len(args))
	for i, argument := range args {
		inputs[i] = lengthInput(argument)
		if inputs[i] == nil {
			return loopBounds{}
		}
	}
	return loopBounds{inputs: distinctInputs(inputs), mathObject: object}
}

func conjunctiveInputs(test *ast.Node, index string) []*ast.Node {
	var inputs []*ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		node = utils.ESTreeRuntimeExpression(node)
		if node != nil && ast.IsBinaryExpression(node) {
			binary := node.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
				return visit(binary.Left) && visit(binary.Right)
			}
		}
		input := lengthInput(upperBound(node, index))
		if input == nil {
			return false
		}
		inputs = append(inputs, input)
		return true
	}
	if !visit(test) {
		return nil
	}
	return distinctInputs(inputs)
}

func distinctInputs(inputs []*ast.Node) []*ast.Node {
	seen := make(map[string]bool, len(inputs))
	for _, input := range inputs {
		if seen[input.Text()] {
			return nil
		}
		seen[input.Text()] = true
	}
	return inputs
}

type inputBinding struct {
	symbol         *ast.Symbol
	argumentsOwner *ast.Node
}

func bindingOf(ctx rule.RuleContext, node *ast.Node) inputBinding {
	binding := inputBinding{symbol: ctx.Refs.ResolveInFile(node)}
	if node.Text() == "arguments" && binding.symbol != nil && len(binding.symbol.Declarations) == 0 {
		// The binder shares an implicit arguments symbol across functions.
		binding.argumentsOwner = ast.GetThisContainer(node, false, false)
	}
	return binding
}

type inputReads struct {
	name    string
	binding inputBinding
	members []*ast.Node
}

func collectReads(ctx rule.RuleContext, body *ast.Node, index *ast.Symbol, inputs []*ast.Node) []inputReads {
	if index == nil {
		return nil
	}
	reads := make([]inputReads, len(inputs))
	byName := make(map[string]*inputReads, len(inputs))
	for i, input := range inputs {
		binding := bindingOf(ctx, input)
		if binding.symbol == index {
			return nil
		}
		reads[i] = inputReads{name: input.Text(), binding: binding}
		byName[input.Text()] = &reads[i]
	}

	// Consume each matching access once. Any remaining reference to an input
	// or the index disqualifies the loop.
	var invalidUse func(*ast.Node) bool
	invalidUse = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindElementAccessExpression:
			access := node.AsElementAccessExpression()
			object := utils.ESTreeRuntimeExpression(access.Expression)
			key := utils.ESTreeRuntimeExpression(access.ArgumentExpression)
			if isIdentifier(object) && isIdentifier(key) && key.Text() == index.Name {
				if input := byName[object.Text()]; input != nil && bindingOf(ctx, object) == input.binding &&
					ctx.Refs.ResolveInFile(key) == index {
					if ast.IsOptionalChain(node) || isCaptured(node, body) || !isElementRead(node) {
						return true
					}
					input.members = append(input.members, node)
					return false
				}
			}
		case ast.KindIdentifier:
			name := node.Text()
			input := byName[name]
			if (name == index.Name || input != nil) && scope.IsReferenceIdentifier(node) {
				binding := bindingOf(ctx, node)
				return binding.symbol == index || (input != nil && binding == input.binding)
			}
		}
		return node.ForEachChild(invalidUse)
	}
	if body.ForEachChild(invalidUse) {
		return nil
	}
	for _, input := range reads {
		if len(input.members) == 0 {
			return nil
		}
	}
	return reads
}

func isCaptured(node, body *ast.Node) bool {
	for parent := node.Parent; parent != body; parent = node.Parent {
		// An object method's computed name is evaluated outside its function.
		if ast.IsClassLike(parent) || (ast.IsFunctionLikeDeclaration(parent) && parent.Name() != node) {
			return true
		}
		node = parent
	}
	return false
}

func isElementRead(node *ast.Node) bool {
	for parent := node.Parent; parent != nil &&
		ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments) &&
		parent.Expression() == node; parent = node.Parent {
		node = parent
	}
	return !unicornutil.IsReferenceTarget(node) &&
		!ast.IsCallExpressionTarget(node, false, false) &&
		!ast.IsTaggedTemplateTag(node, false, false)
}
