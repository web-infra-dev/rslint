package utils

import (
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/evaluator"
)

// StaticStringEvaluator folds expressions to string constants. It wraps tsgo's
// evaluator and adds stable local variable resolution through the TypeChecker.
// Create one evaluator per linted file; it keeps write-reference and recursion
// state for that file.
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

// staticObjectValue and staticArrayValue hold folded object and array literals.
// They exist so member access can be resolved on them; neither converts to a
// string, so a rule that asks for a string value still gets "not statically
// known".
type staticObjectValue struct {
	properties map[string]any
}

type staticArrayValue struct {
	elements []any
}

// objectPassThroughMethods are the `Object` methods that return their argument
// unchanged, so folding can see through them.
var objectPassThroughMethods = map[string]bool{
	"freeze":            true,
	"preventExtensions": true,
	"seal":              true,
}

type staticEvalResult struct {
	value any
	ok    bool
}

// Eval returns the static string value of node, if it can be determined. It
// covers the string-producing subset of ESLint's getStaticValue that rules use
// for computed property names, key arguments, constructor arguments, and other
// string-only rule inputs. It includes nested conditionals, logical
// short-circuiting, String(), String.raw, and local variables with stable
// initializers.
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

// EvalValue returns the static value of node if it can be determined, regardless
// of its type. It allows rules to check if a value is statically known to be a
// non-string (like a boolean or number).
func (staticEvaluator *StaticStringEvaluator) EvalValue(node *ast.Node) (any, bool) {
	if staticEvaluator == nil || node == nil {
		return nil, false
	}
	result := staticEvaluator.evalValue(node)
	if !result.ok {
		return nil, false
	}
	return result.value, true
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
	case ast.KindObjectLiteralExpression:
		if result := staticEvaluator.evalObjectLiteral(node); result.ok {
			return result
		}
	case ast.KindArrayLiteralExpression:
		if result := staticEvaluator.evalArrayLiteral(node); result.ok {
			return result
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if result := staticEvaluator.evalMemberAccess(node); result.ok {
			return result
		}
	case ast.KindCallExpression:
		if result := staticEvaluator.evalStringCall(node); result.ok {
			return result
		}
		if result := staticEvaluator.evalObjectPassThroughCall(node); result.ok {
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
	initializer, symbol, ok := staticEvaluator.resolveIdentifierInitializer(node)
	if !ok || staticEvaluator.resolving[symbol] {
		return staticEvalResult{}
	}

	staticEvaluator.resolving[symbol] = true
	defer delete(staticEvaluator.resolving, symbol)
	return staticEvaluator.evalValue(initializer)
}

// ResolveIdentifierInitializer resolves an identifier to a stable variable
// initializer: const bindings, or let/var bindings with no write references in
// the current source file. It returns false for destructuring, using bindings,
// ambiguous symbols, or files where let/var writes cannot be checked.
func (staticEvaluator *StaticStringEvaluator) ResolveIdentifierInitializer(node *ast.Node) (*ast.Node, bool) {
	if staticEvaluator == nil {
		return nil, false
	}
	initializer, _, ok := staticEvaluator.resolveIdentifierInitializer(node)
	return initializer, ok
}

func (staticEvaluator *StaticStringEvaluator) resolveIdentifierInitializer(node *ast.Node) (*ast.Node, *ast.Symbol, bool) {
	if staticEvaluator == nil || staticEvaluator.typeChecker == nil {
		return nil, nil, false
	}

	expr := SkipAssertionsAndParens(node)
	if expr == nil || expr.Kind != ast.KindIdentifier {
		return nil, nil, false
	}

	symbol := GetReferenceSymbol(expr, staticEvaluator.typeChecker)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil, nil, false
	}

	declarationNode := symbol.Declarations[0]
	if declarationNode == nil || declarationNode.Kind != ast.KindVariableDeclaration {
		return nil, nil, false
	}

	declaration := declarationNode.AsVariableDeclaration()
	if declaration == nil || declaration.Initializer == nil || !isIdentifierWithText(declaration.Name(), expr.AsIdentifier().Text) {
		return nil, nil, false
	}

	declarationList := declarationNode.Parent
	if declarationList == nil || declarationList.Kind != ast.KindVariableDeclarationList {
		return nil, nil, false
	}
	if ast.IsVarUsing(declarationList) || ast.IsVarAwaitUsing(declarationList) {
		return nil, nil, false
	}
	if !ast.IsVarConst(declarationList) && staticEvaluator.hasWrites(symbol) {
		return nil, nil, false
	}
	return declaration.Initializer, symbol, true
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
	case ast.KindEqualsToken:
		return staticEvaluator.evalValue(binary.Right)
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

func (staticEvaluator *StaticStringEvaluator) evalObjectLiteral(node *ast.Node) staticEvalResult {
	object := node.AsObjectLiteralExpression()
	if object == nil || object.Properties == nil {
		return staticEvalResult{}
	}

	properties := map[string]any{}
	for _, property := range object.Properties.Nodes {
		var valueNode *ast.Node
		switch property.Kind {
		case ast.KindPropertyAssignment:
			valueNode = property.AsPropertyAssignment().Initializer
		case ast.KindShorthandPropertyAssignment:
			valueNode = property.Name()
		default:
			return staticEvalResult{}
		}

		key, ok := staticEvaluator.evalPropertyKey(property.Name())
		if !ok {
			return staticEvalResult{}
		}
		value := staticEvaluator.evalValue(valueNode)
		if !value.ok {
			return staticEvalResult{}
		}
		properties[key] = value.value
	}
	return staticEvalResult{value: staticObjectValue{properties: properties}, ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evalArrayLiteral(node *ast.Node) staticEvalResult {
	array := node.AsArrayLiteralExpression()
	if array == nil || array.Elements == nil {
		return staticEvalResult{}
	}

	elements := make([]any, 0, len(array.Elements.Nodes))
	for _, element := range array.Elements.Nodes {
		if element.Kind == ast.KindOmittedExpression {
			elements = append(elements, staticUndefinedValue{})
			continue
		}
		if element.Kind == ast.KindSpreadElement {
			return staticEvalResult{}
		}
		value := staticEvaluator.evalValue(element)
		if !value.ok {
			return staticEvalResult{}
		}
		elements = append(elements, value.value)
	}
	return staticEvalResult{value: staticArrayValue{elements: elements}, ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evalPropertyKey(name *ast.Node) (string, bool) {
	if name == nil {
		return "", false
	}

	switch name.Kind {
	case ast.KindIdentifier:
		return name.AsIdentifier().Text, true
	case ast.KindStringLiteral:
		return name.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return name.AsNoSubstitutionTemplateLiteral().Text, true
	case ast.KindNumericLiteral:
		// Not the source text: `{1e3: x}` has the key "1000", which is what an
		// index expression folds to.
		result := staticEvaluator.evalValue(name)
		if !result.ok {
			return "", false
		}
		return staticValueToString(result.value)
	case ast.KindComputedPropertyName:
		computed := name.AsComputedPropertyName()
		if computed == nil {
			return "", false
		}
		result := staticEvaluator.evalValue(computed.Expression)
		if !result.ok {
			return "", false
		}
		return staticValueToString(result.value)
	}
	return "", false
}

func (staticEvaluator *StaticStringEvaluator) evalMemberAccess(node *ast.Node) staticEvalResult {
	var objectNode *ast.Node
	var key string

	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		if access == nil || access.QuestionDotToken != nil ||
			access.Name() == nil || access.Name().Kind != ast.KindIdentifier {
			return staticEvalResult{}
		}
		objectNode = access.Expression
		key = access.Name().AsIdentifier().Text
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		if access == nil || access.QuestionDotToken != nil {
			return staticEvalResult{}
		}
		argument := staticEvaluator.evalValue(access.ArgumentExpression)
		if !argument.ok {
			return staticEvalResult{}
		}
		value, ok := staticValueToString(argument.value)
		if !ok {
			return staticEvalResult{}
		}
		objectNode = access.Expression
		key = value
	default:
		return staticEvalResult{}
	}

	object := staticEvaluator.evalValue(objectNode)
	if !object.ok {
		return staticEvalResult{}
	}
	return staticMemberValue(object.value, key)
}

// staticMemberValue reads key off a folded object or array literal. Reading a
// key that isn't there yields `undefined`, as it would at runtime. Anything
// else — a non-index string, an array `length` — stays unresolved, because
// folding it would need a numeric value representation this evaluator doesn't
// carry.
func staticMemberValue(object any, key string) staticEvalResult {
	switch object := object.(type) {
	case staticObjectValue:
		if value, ok := object.properties[key]; ok {
			return staticEvalResult{value: value, ok: true}
		}
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	case staticArrayValue:
		index, ok := staticArrayIndex(key)
		if !ok {
			return staticEvalResult{}
		}
		if index < 0 || index >= len(object.elements) {
			return staticEvalResult{value: staticUndefinedValue{}, ok: true}
		}
		return staticEvalResult{value: object.elements[index], ok: true}
	}
	return staticEvalResult{}
}

// staticArrayIndex parses key as a canonical array index string: decimal
// digits only, with no leading zero unless key is exactly "0". JavaScript
// treats any other shape (`"00"`, `"1e3"`, `"-1"`) as an ordinary property
// name, which array literals don't have.
func staticArrayIndex(key string) (int, bool) {
	if key == "0" {
		return 0, true
	}
	if key == "" || key[0] < '1' || key[0] > '9' {
		return 0, false
	}
	for i := 1; i < len(key); i++ {
		if key[i] < '0' || key[i] > '9' {
			return 0, false
		}
	}
	index, err := strconv.Atoi(key)
	if err != nil {
		return 0, false
	}
	return index, true
}

// evalObjectPassThroughCall folds `Object.freeze(x)` and its siblings to the
// value of x.
func (staticEvaluator *StaticStringEvaluator) evalObjectPassThroughCall(node *ast.Node) staticEvalResult {
	call := node.AsCallExpression()
	if call == nil || call.QuestionDotToken != nil {
		return staticEvalResult{}
	}

	callee := ast.SkipOuterExpressions(call.Expression, ast.OEKParentheses|ast.OEKAssertions)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return staticEvalResult{}
	}

	propertyAccess := callee.AsPropertyAccessExpression()
	if propertyAccess == nil || propertyAccess.QuestionDotToken != nil ||
		propertyAccess.Name() == nil || propertyAccess.Name().Kind != ast.KindIdentifier ||
		!objectPassThroughMethods[propertyAccess.Name().AsIdentifier().Text] {
		return staticEvalResult{}
	}

	object := ast.SkipOuterExpressions(propertyAccess.Expression, ast.OEKParentheses|ast.OEKAssertions)
	if !isIdentifierWithText(object, "Object") || IsShadowed(object, "Object") {
		return staticEvalResult{}
	}

	args := node.Arguments()
	if len(args) == 0 || ast.IsSpreadElement(args[0]) {
		return staticEvalResult{}
	}
	return staticEvaluator.evalValue(args[0])
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
	if tagged == nil || tagged.Template == nil || !staticEvaluator.isStringRawTag(tagged.Tag) {
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

func (staticEvaluator *StaticStringEvaluator) isStringRawTag(tag *ast.Node) bool {
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
	return staticEvaluator.isBuiltinStringValue(object, map[*ast.Symbol]bool{})
}

func (staticEvaluator *StaticStringEvaluator) isBuiltinStringValue(node *ast.Node, resolvingAliases map[*ast.Symbol]bool) bool {
	node = SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if isIdentifierWithText(node, "String") && !IsShadowed(node, "String") {
		return true
	}
	initializer, symbol, ok := staticEvaluator.resolveIdentifierInitializer(node)
	if !ok || resolvingAliases[symbol] {
		return false
	}
	resolvingAliases[symbol] = true
	defer delete(resolvingAliases, symbol)
	return staticEvaluator.isBuiltinStringValue(initializer, resolvingAliases)
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
			if symbol := GetReferenceSymbol(node, staticEvaluator.typeChecker); symbol != nil {
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
	case bool, staticNullValue, staticUndefinedValue, staticObjectValue, staticArrayValue:
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
	case staticObjectValue, staticArrayValue:
		return true, true
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
	case staticArrayValue:
		return staticArrayToString(value)
	case staticObjectValue:
		// A plain object's default ToString is "[object Object]", but a
		// folded object literal can't tell whether a property overrides
		// toString/Symbol.toPrimitive; callers only need to know that the
		// value is not a string.
		return "", false
	default:
		return evaluatorString(func() string { return evaluator.AnyToString(value) })
	}
}

// staticArrayToString implements `Array.prototype.join(',')`, which is what
// an array coerces to via the default ToString. Nullish elements fold to "",
// not "null"/"undefined" (unlike stringifying them directly).
func staticArrayToString(value staticArrayValue) (string, bool) {
	parts := make([]string, len(value.elements))
	for i, element := range value.elements {
		if staticValueNullish(element) {
			continue
		}
		part, ok := staticValueToString(element)
		if !ok {
			return "", false
		}
		parts[i] = part
	}
	return strings.Join(parts, ","), true
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
