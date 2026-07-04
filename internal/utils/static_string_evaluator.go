package utils

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/evaluator"
)

// StaticStringEvaluator folds expressions to string constants. It wraps tsgo's
// evaluator and adds local const identifier resolution through the TypeChecker.
// Create one evaluator per linted file; it keeps a small recursion guard.
type StaticStringEvaluator struct {
	typeChecker       *checker.Checker
	sourceFile        *ast.SourceFile
	evaluator         evaluator.Evaluator
	resolving         map[*ast.Symbol]bool
	writeRefsComputed bool
	writeRefs         map[*ast.Symbol]bool
}

func NewStaticStringEvaluator(typeChecker *checker.Checker) *StaticStringEvaluator {
	return NewStaticStringEvaluatorWithSourceFile(typeChecker, nil)
}

func NewStaticStringEvaluatorWithSourceFile(typeChecker *checker.Checker, sourceFile *ast.SourceFile) *StaticStringEvaluator {
	staticEvaluator := &StaticStringEvaluator{
		typeChecker: typeChecker,
		sourceFile:  sourceFile,
		resolving:   map[*ast.Symbol]bool{},
	}
	staticEvaluator.evaluator = evaluator.NewEvaluator(staticEvaluator.evaluateEntity, ast.OEKAssertions)
	return staticEvaluator
}

type staticNullValue struct{}
type staticUndefinedValue struct{}

type staticEvalResult struct {
	value any
	ok    bool
}

// Eval returns the static string value of node, if it can be determined. It
// covers the string-producing subset of ESLint's getStaticValue that rules use
// for computed property names and key arguments, including nested conditionals,
// logical short-circuiting, String(), String.raw, and local variables with
// stable initializers.
func (staticEvaluator *StaticStringEvaluator) Eval(node *ast.Node) (string, bool) {
	if staticEvaluator == nil || node == nil {
		return "", false
	}
	result := staticEvaluator.evalValue(node)
	if !result.ok {
		return "", false
	}
	value, ok := result.value.(string)
	return value, ok
}

func (staticEvaluator *StaticStringEvaluator) evalValue(node *ast.Node) staticEvalResult {
	node = SkipAssertionsAndParens(node)
	if node == nil {
		return staticEvalResult{}
	}

	switch node.Kind {
	case ast.KindStringLiteral:
		return staticEvalResult{value: node.AsStringLiteral().Text, ok: true}
	case ast.KindNoSubstitutionTemplateLiteral:
		return staticEvalResult{value: node.AsNoSubstitutionTemplateLiteral().Text, ok: true}
	case ast.KindTrueKeyword:
		return staticEvalResult{value: true, ok: true}
	case ast.KindFalseKeyword:
		return staticEvalResult{value: false, ok: true}
	case ast.KindNullKeyword:
		return staticEvalResult{value: staticNullValue{}, ok: true}
	case ast.KindUndefinedKeyword:
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	case ast.KindIdentifier:
		identifier := node.AsIdentifier()
		if identifier != nil && identifier.Text == "undefined" && !IsShadowed(node, "undefined") {
			return staticEvalResult{value: staticUndefinedValue{}, ok: true}
		}
		return staticEvaluator.evalIdentifier(node)
	case ast.KindTemplateExpression:
		return staticEvaluator.evalTemplateExpression(node)
	case ast.KindBinaryExpression:
		return staticEvaluator.evalBinaryExpression(node)
	case ast.KindPrefixUnaryExpression:
		return staticEvaluator.evalPrefixUnaryExpression(node)
	case ast.KindConditionalExpression:
		return staticEvaluator.evalConditionalExpression(node)
	case ast.KindVoidExpression:
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	case ast.KindCallExpression:
		if result := staticEvaluator.evalStringCall(node); result.ok {
			return result
		}
	case ast.KindTaggedTemplateExpression:
		if result := staticEvaluator.evalStringRawTag(node); result.ok {
			return result
		}
	}

	return staticEvaluator.evalWithTsgo(node)
}

func (staticEvaluator *StaticStringEvaluator) evalIdentifier(node *ast.Node) staticEvalResult {
	if staticEvaluator.typeChecker == nil {
		return staticEvalResult{}
	}

	expr := SkipAssertionsAndParens(node)
	if expr == nil || expr.Kind != ast.KindIdentifier {
		return staticEvalResult{}
	}

	symbol := GetReferenceSymbol(expr, staticEvaluator.typeChecker)
	if symbol == nil || len(symbol.Declarations) != 1 || staticEvaluator.resolving[symbol] {
		return staticEvalResult{}
	}

	declarationNode := symbol.Declarations[0]
	if declarationNode == nil || declarationNode.Kind != ast.KindVariableDeclaration {
		return staticEvalResult{}
	}

	declaration := declarationNode.AsVariableDeclaration()
	if declaration == nil || declaration.Initializer == nil || !isIdentifierWithText(declaration.Name(), expr.AsIdentifier().Text) {
		return staticEvalResult{}
	}

	declarationList := declarationNode.Parent
	if declarationList == nil || declarationList.Kind != ast.KindVariableDeclarationList {
		return staticEvalResult{}
	}
	if ast.IsVarUsing(declarationList) || ast.IsVarAwaitUsing(declarationList) {
		return staticEvalResult{}
	}
	if !ast.IsVarConst(declarationList) && staticEvaluator.hasWrites(symbol) {
		return staticEvalResult{}
	}

	staticEvaluator.resolving[symbol] = true
	defer delete(staticEvaluator.resolving, symbol)
	return staticEvaluator.evalValue(declaration.Initializer)
}

func (staticEvaluator *StaticStringEvaluator) evalTemplateExpression(node *ast.Node) staticEvalResult {
	template := node.AsTemplateExpression()
	if template == nil {
		return staticEvalResult{}
	}

	var builder strings.Builder
	if template.Head != nil {
		builder.WriteString(template.Head.Text())
	}
	if template.TemplateSpans != nil {
		for _, spanNode := range template.TemplateSpans.Nodes {
			span := spanNode.AsTemplateSpan()
			if span == nil {
				return staticEvalResult{}
			}
			result := staticEvaluator.evalValue(span.Expression)
			if !result.ok {
				return staticEvalResult{}
			}
			value, ok := staticValueToString(result.value)
			if !ok {
				return staticEvalResult{}
			}
			builder.WriteString(value)
			if span.Literal != nil {
				builder.WriteString(span.Literal.Text())
			}
		}
	}
	return staticEvalResult{value: builder.String(), ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evalBinaryExpression(node *ast.Node) staticEvalResult {
	binary := node.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return staticEvalResult{}
	}

	switch binary.OperatorToken.Kind {
	case ast.KindBarBarToken:
		left := staticEvaluator.evalValue(binary.Left)
		if !left.ok {
			return staticEvalResult{}
		}
		truthy, ok := staticValueTruthy(left.value)
		if !ok {
			return staticEvalResult{}
		}
		if truthy {
			return left
		}
		return staticEvaluator.evalValue(binary.Right)
	case ast.KindAmpersandAmpersandToken:
		left := staticEvaluator.evalValue(binary.Left)
		if !left.ok {
			return staticEvalResult{}
		}
		truthy, ok := staticValueTruthy(left.value)
		if !ok {
			return staticEvalResult{}
		}
		if !truthy {
			return left
		}
		return staticEvaluator.evalValue(binary.Right)
	case ast.KindQuestionQuestionToken:
		left := staticEvaluator.evalValue(binary.Left)
		if !left.ok {
			return staticEvalResult{}
		}
		if staticValueNullish(left.value) {
			return staticEvaluator.evalValue(binary.Right)
		}
		return left
	case ast.KindPlusToken:
		left := staticEvaluator.evalValue(binary.Left)
		right := staticEvaluator.evalValue(binary.Right)
		if !left.ok || !right.ok {
			return staticEvalResult{}
		}
		if _, ok := left.value.(string); ok {
			return staticEvaluator.concatStaticValues(left.value, right.value)
		}
		if _, ok := right.value.(string); ok {
			return staticEvaluator.concatStaticValues(left.value, right.value)
		}
	}

	return staticEvaluator.evalWithTsgo(node)
}

func (staticEvaluator *StaticStringEvaluator) concatStaticValues(left any, right any) staticEvalResult {
	leftString, ok := staticValueToString(left)
	if !ok {
		return staticEvalResult{}
	}
	rightString, ok := staticValueToString(right)
	if !ok {
		return staticEvalResult{}
	}
	return staticEvalResult{value: leftString + rightString, ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evalPrefixUnaryExpression(node *ast.Node) staticEvalResult {
	prefix := node.AsPrefixUnaryExpression()
	if prefix == nil || prefix.Operator != ast.KindExclamationToken {
		return staticEvaluator.evalWithTsgo(node)
	}

	operand := staticEvaluator.evalValue(prefix.Operand)
	if !operand.ok {
		return staticEvalResult{}
	}
	truthy, ok := staticValueTruthy(operand.value)
	if !ok {
		return staticEvalResult{}
	}
	return staticEvalResult{value: !truthy, ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evalConditionalExpression(node *ast.Node) staticEvalResult {
	conditional := node.AsConditionalExpression()
	if conditional == nil {
		return staticEvalResult{}
	}

	condition := staticEvaluator.evalValue(conditional.Condition)
	if !condition.ok {
		return staticEvalResult{}
	}
	truthy, ok := staticValueTruthy(condition.value)
	if !ok {
		return staticEvalResult{}
	}
	if truthy {
		return staticEvaluator.evalValue(conditional.WhenTrue)
	}
	return staticEvaluator.evalValue(conditional.WhenFalse)
}

func (staticEvaluator *StaticStringEvaluator) evalStringCall(node *ast.Node) staticEvalResult {
	call := node.AsCallExpression()
	if call == nil || call.QuestionDotToken != nil {
		return staticEvalResult{}
	}

	callee := ast.SkipOuterExpressions(call.Expression, ast.OEKParentheses|ast.OEKAssertions)
	if callee == nil || callee.Kind != ast.KindIdentifier ||
		callee.AsIdentifier().Text != "String" || IsShadowed(callee, "String") {
		return staticEvalResult{}
	}

	args := node.Arguments()
	if len(args) == 0 {
		return staticEvalResult{value: "", ok: true}
	}
	if ast.IsSpreadElement(args[0]) {
		return staticEvalResult{}
	}
	arg := staticEvaluator.evalValue(args[0])
	if !arg.ok {
		return staticEvalResult{}
	}
	value, ok := staticValueToString(arg.value)
	if !ok {
		return staticEvalResult{}
	}
	return staticEvalResult{value: value, ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evalStringRawTag(node *ast.Node) staticEvalResult {
	tagged := node.AsTaggedTemplateExpression()
	if tagged == nil || tagged.Template == nil || !isStringRawTag(tagged.Tag) {
		return staticEvalResult{}
	}

	switch tagged.Template.Kind {
	case ast.KindNoSubstitutionTemplateLiteral:
		return staticEvalResult{value: tagged.Template.Text(), ok: true}
	case ast.KindTemplateExpression:
		return staticEvaluator.evalTemplateExpression(tagged.Template)
	default:
		return staticEvalResult{}
	}
}

func isStringRawTag(tag *ast.Node) bool {
	tag = ast.SkipOuterExpressions(tag, ast.OEKParentheses|ast.OEKAssertions)
	if tag == nil || tag.Kind != ast.KindPropertyAccessExpression {
		return false
	}

	propertyAccess := tag.AsPropertyAccessExpression()
	if propertyAccess == nil || propertyAccess.QuestionDotToken != nil ||
		!isIdentifierWithText(propertyAccess.Name(), "raw") {
		return false
	}

	object := ast.SkipOuterExpressions(propertyAccess.Expression, ast.OEKParentheses|ast.OEKAssertions)
	return isIdentifierWithText(object, "String") && !IsShadowed(object, "String")
}

func (staticEvaluator *StaticStringEvaluator) evalWithTsgo(node *ast.Node) (result staticEvalResult) {
	defer func() {
		if recover() != nil {
			result = staticEvalResult{}
		}
	}()

	value := staticEvaluator.evaluator(node, node).Value
	if value == nil {
		return staticEvalResult{}
	}
	return staticEvalResult{value: value, ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evaluateEntity(expr *ast.Node, location *ast.Node) evaluator.Result {
	result := staticEvaluator.evalIdentifier(expr)
	if !result.ok || !staticValueIsTsgoSafe(result.value) {
		return evaluator.Result{}
	}

	return evaluator.Result{Value: result.value}
}

func (staticEvaluator *StaticStringEvaluator) hasWrites(symbol *ast.Symbol) bool {
	if staticEvaluator.sourceFile == nil || staticEvaluator.typeChecker == nil {
		return true
	}
	if !staticEvaluator.writeRefsComputed {
		staticEvaluator.computeWriteRefs()
	}
	return staticEvaluator.writeRefs[symbol]
}

func (staticEvaluator *StaticStringEvaluator) computeWriteRefs() {
	staticEvaluator.writeRefsComputed = true
	staticEvaluator.writeRefs = map[*ast.Symbol]bool{}

	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindIdentifier && IsWriteReference(node) {
			if symbol := staticEvaluator.typeChecker.GetSymbolAtLocation(node); symbol != nil {
				staticEvaluator.writeRefs[symbol] = true
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(&staticEvaluator.sourceFile.Node)
}

func staticValueIsTsgoSafe(value any) bool {
	switch value.(type) {
	case bool, staticNullValue, staticUndefinedValue:
		return false
	default:
		return value != nil
	}
}

func staticValueNullish(value any) bool {
	switch value.(type) {
	case staticNullValue, staticUndefinedValue:
		return true
	default:
		return false
	}
}

func staticValueTruthy(value any) (truthy bool, ok bool) {
	switch value := value.(type) {
	case string:
		return value != "", true
	case bool:
		return value, true
	case staticNullValue, staticUndefinedValue:
		return false, true
	default:
		return evaluatorBool(func() bool { return evaluator.IsTruthy(value) })
	}
}

func staticValueToString(value any) (string, bool) {
	switch value := value.(type) {
	case string:
		return value, true
	case bool:
		if value {
			return "true", true
		}
		return "false", true
	case staticNullValue:
		return "null", true
	case staticUndefinedValue:
		return "undefined", true
	default:
		return evaluatorString(func() string { return evaluator.AnyToString(value) })
	}
}

func evaluatorBool(fn func() bool) (value bool, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return fn(), true
}

func evaluatorString(fn func() string) (value string, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return fn(), true
}

func isIdentifierWithText(node *ast.Node, text string) bool {
	return node != nil && node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == text
}
