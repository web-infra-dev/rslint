package prefer_set_has

// Static array-size evaluation, known-unique-array detection, and TypeScript
// type-annotation rewriting for prefer-set-has.
//
// Why not reuse utils.StaticStringEvaluator?
//
// That evaluator's public API is `Eval(node) (string, bool)` — it collapses
// every value to a string on the way out, turning `true` into "true", `null`
// into "null" and `undefined` into "undefined". This rule compares array
// elements for SameValueZero uniqueness, so it must keep them typed: `'null'`
// and `null`, `'1'` and `1`, `'true'` and `true` are three pairs of DISTINCT
// elements that a string-collapsed model reports as equal. The same applies to
// `-0` vs `0` and to bigint vs number, both of which this rule distinguishes.
//
// It also falls back to tsgo's evaluator.Evaluator, which exists to compute
// const-enum member values and is wrapped in recover() because it panics on
// unexpected input. This rule emits an AUTOFIX: a wrong uniqueness verdict
// produces a `new Set(...)` that silently drops elements, so its value
// semantics have to be auditable line-by-line against upstream's
// eslint-utils getStaticValue rather than delegated to a black box.
//
// A third evaluator, jsx_a11y/jsxa11yutil/static_eval.go, split off from the
// same shared one for the same reason — a typed value model the string API
// cannot express. The repo's convention is to share helpers, not evaluators.
//
// Helpers we DO reuse:
//
//	utils.NormalizeNumericLiteral / NormalizeBigIntLiteral — text → value
//	utils.IsShadowed                                       — global checks
//	ast.SkipOuterExpressions(OEKParentheses|OEKAssertions) — wrapper unwrap
//	utils.RangeEnclosingDelimiters                         — `<T>` recovery
//
// JS semantics that both evaluators need (ToString, string `+`, template
// folding) are mirrored from StaticStringEvaluator's concatStaticValues /
// evalTemplateExpression / staticValueToString rather than reinvented; the
// split is over the value model, not over what JS does.

import (
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// getKnownArraySize mirrors upstream's `getKnownArraySize`: the statically
// known element count of an array-producing initializer, when determinable.
func getKnownArraySize(ctx rule.RuleContext, node *ast.Node) (int, bool) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return 0, false
	}

	if ast.IsArrayLiteralExpression(node) {
		elements := node.AsArrayLiteralExpression().Elements
		if hasSpread(elements) {
			return 0, false
		}
		return len(elements.Nodes), true
	}

	if isArrayConstructorCall(node) {
		return getArrayConstructorSize(ctx, node)
	}

	if call := matchArrayStaticCall(node, "of"); call != nil {
		args := call.Arguments()
		if hasSpread(argumentList(call)) {
			return 0, false
		}
		return len(args), true
	}

	if call := matchArrayStaticCall(node, "from"); call != nil {
		return getArrayFromSize(ctx, call)
	}

	return 0, false
}

// getArrayConstructorSize mirrors upstream's `getArrayConstructorSize` for
// `Array(...)` / `new Array(...)`.
func getArrayConstructorSize(ctx rule.RuleContext, node *ast.Node) (int, bool) {
	args := callOrNewArguments(node)
	if hasSpreadNodes(args) {
		return 0, false
	}
	if len(args) != 1 {
		return len(args), true
	}

	value, ok := evaluateStaticValue(ctx, args[0])
	if !ok {
		return 0, false
	}
	number, isNumber := value.asNumber()
	if !isNumber {
		// `Array('x')` → length 1.
		return 1, true
	}
	if isArrayLength(number) {
		return int(number), true
	}
	return 0, false
}

// getArrayFromSize mirrors upstream's `getArrayFromSize` for `Array.from(...)`.
func getArrayFromSize(ctx rule.RuleContext, node *ast.Node) (int, bool) {
	args := node.Arguments()
	if len(args) == 0 {
		return 0, false
	}
	source := args[0]
	if ast.IsSpreadElement(source) {
		return 0, false
	}

	source = ast.SkipParentheses(source)
	if ast.IsArrayLiteralExpression(source) {
		elements := source.AsArrayLiteralExpression().Elements
		if hasSpread(elements) {
			return 0, false
		}
		return len(elements.Nodes), true
	}

	// `Array.from('string')` → number of code points.
	if text, isString := staticStringLike(source); isString {
		return len([]rune(text)), true
	}

	return getObjectLength(ctx, source)
}

// getObjectLength mirrors upstream's `getObjectLength`: a single-property
// object literal `{length: <static array length>}`.
func getObjectLength(ctx rule.RuleContext, node *ast.Node) (int, bool) {
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return 0, false
	}
	properties := node.AsObjectLiteralExpression().Properties
	if properties == nil || len(properties.Nodes) != 1 {
		return 0, false
	}
	valueNode, isLength := lengthPropertyValue(properties.Nodes[0])
	if !isLength {
		return 0, false
	}
	value, ok := evaluateStaticValue(ctx, valueNode)
	if !ok {
		return 0, false
	}
	number, isNumber := value.asNumber()
	if !isNumber || !isArrayLength(number) {
		return 0, false
	}
	return int(number), true
}

// lengthPropertyValue mirrors upstream's `isLengthProperty` plus the
// `property.value` read that follows it: for a non-computed `length` property,
// it returns the node holding the value. ESTree models `{length}` as a Property
// whose `value` is the key Identifier, so upstream resolves it through scope;
// tsgo gives shorthand its own kind, so handle it explicitly.
func lengthPropertyValue(property *ast.Node) (*ast.Node, bool) {
	if property == nil {
		return nil, false
	}

	var name *ast.Node
	var value *ast.Node
	switch property.Kind {
	case ast.KindPropertyAssignment:
		name = property.AsPropertyAssignment().Name()
		value = property.AsPropertyAssignment().Initializer
	case ast.KindShorthandPropertyAssignment:
		name = property.AsShorthandPropertyAssignment().Name()
		value = name
	default:
		return nil, false
	}
	if name == nil || value == nil {
		return nil, false
	}

	switch name.Kind {
	case ast.KindIdentifier:
		if name.AsIdentifier().Text == "length" {
			return value, true
		}
	case ast.KindStringLiteral:
		if name.AsStringLiteral().Text == "length" {
			return value, true
		}
	}
	return nil, false
}

// isNonNegativeInteger mirrors upstream's `isNonNegativeInteger`.
func isNonNegativeInteger(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) &&
		value == math.Trunc(value) && value >= 0 &&
		value <= float64(1<<53-1)
}

// isArrayLength mirrors upstream's `isArrayLength`.
func isArrayLength(value float64) bool {
	return isNonNegativeInteger(value) && value <= maxArrayLength
}

// isKnownUniqueArrayExpression mirrors upstream's
// `isKnownUniqueArrayExpression`: a plain array literal with no holes, no
// spreads, and only unique statically-known comparable primitive / null
// elements (excluding `-0`).
func isKnownUniqueArrayExpression(ctx rule.RuleContext, node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil || !ast.IsArrayLiteralExpression(node) {
		return false
	}
	elements := node.AsArrayLiteralExpression().Elements
	if hasSpread(elements) {
		return false
	}

	values := make([]staticValue, 0, len(elements.Nodes))
	for _, element := range elements.Nodes {
		if ast.IsOmittedExpression(element) {
			return false // hole
		}
		value, ok := evaluateStaticValue(ctx, element)
		if !ok || !value.isComparable() || value.isNegativeZero() {
			return false
		}
		values = append(values, value)
	}

	for i := range values {
		for j := range i {
			if sameValueZero(values[j], values[i]) {
				return false
			}
		}
	}
	return true
}

// hasSpread reports whether an element list contains a spread element.
func hasSpread(list *ast.NodeList) bool {
	if list == nil {
		return false
	}
	return hasSpreadNodes(list.Nodes)
}

func hasSpreadNodes(nodes []*ast.Node) bool {
	for _, node := range nodes {
		if node != nil && ast.IsSpreadElement(node) {
			return true
		}
	}
	return false
}

func argumentList(call *ast.Node) *ast.NodeList {
	if ast.IsCallExpression(call) {
		return call.AsCallExpression().Arguments
	}
	return nil
}

// callOrNewArguments returns the argument nodes of a call or new expression.
func callOrNewArguments(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindCallExpression:
		if args := node.AsCallExpression().Arguments; args != nil {
			return args.Nodes
		}
	case ast.KindNewExpression:
		if args := node.AsNewExpression().Arguments; args != nil {
			return args.Nodes
		}
	}
	return nil
}

// setTypeAnnotationText mirrors upstream's `getSetTypeAnnotationText`. It
// returns the replacement `Set<…>` / `ReadonlySet<…>` type text and whether a
// safe rewrite exists. When the declaration has no type annotation, it returns
// ("", false) — but the caller only downgrades to a suggestion when a type
// annotation is present, so that case still autofixes.
func setTypeAnnotationText(ctx rule.RuleContext, declaration *ast.VariableDeclaration) (string, bool) {
	typeAnnotation := declaration.Type
	if typeAnnotation == nil {
		return "", false
	}

	// `T[]` → `Set<T>`
	if typeAnnotation.Kind == ast.KindArrayType {
		elementType := typeAnnotation.AsArrayTypeNode().ElementType
		if hasCommentsOutsideNode(ctx, typeAnnotation, elementType) {
			return "", false
		}
		return "Set<" + elementTypeText(ctx, elementType) + ">", true
	}

	// `readonly T[]` → `ReadonlySet<T>`
	if typeAnnotation.Kind == ast.KindTypeOperator {
		operator := typeAnnotation.AsTypeOperatorNode()
		if operator.Operator == ast.KindReadonlyKeyword &&
			operator.Type != nil && operator.Type.Kind == ast.KindArrayType {
			elementType := operator.Type.AsArrayTypeNode().ElementType
			if hasCommentsOutsideNode(ctx, typeAnnotation, elementType) {
				return "", false
			}
			return "ReadonlySet<" + elementTypeText(ctx, elementType) + ">", true
		}
		return "", false
	}

	// `Array<T>` → `Set<T>`, `ReadonlyArray<T>` → `ReadonlySet<T>`
	if typeAnnotation.Kind != ast.KindTypeReference {
		return "", false
	}
	typeReference := typeAnnotation.AsTypeReferenceNode()
	if typeReference.TypeName == nil || !ast.IsIdentifier(typeReference.TypeName) {
		return "", false
	}
	typeArguments := typeReference.TypeArguments
	if typeArguments == nil || len(typeArguments.Nodes) != 1 {
		return "", false
	}
	// Upstream compares against the whole type-argument list — the `<…>` span —
	// not the type argument inside it, so a comment between the angle brackets
	// and the argument (`Array</* c */ string>`) is still "inside" and the
	// rewrite stays available. The replacement text below carries that comment
	// over verbatim.
	argumentsStart, argumentsEnd, ok := typeArgumentsRange(ctx, typeArguments)
	if !ok {
		return "", false
	}
	annotationRange := utils.TrimNodeTextRange(ctx.SourceFile, typeAnnotation)
	if hasCommentsOutsideRange(ctx, annotationRange.Pos(), annotationRange.End(), argumentsStart, argumentsEnd) {
		return "", false
	}

	typeArgumentsText := ctx.SourceFile.Text()[argumentsStart:argumentsEnd]
	switch typeReference.TypeName.AsIdentifier().Text {
	case "Array":
		return "Set" + typeArgumentsText, true
	case "ReadonlyArray":
		return "ReadonlySet" + typeArgumentsText, true
	}
	return "", false
}

// elementTypeText returns the source text of an array element type. tsgo keeps
// an explicit ParenthesizedType node (`(string | number)[]`) that ESTree
// flattens, so unwrap it to match upstream's `sourceCode.getText(elementType)`,
// which sees the inner type without the wrapping parentheses.
func elementTypeText(ctx rule.RuleContext, elementType *ast.Node) string {
	for elementType != nil && elementType.Kind == ast.KindParenthesizedType {
		elementType = elementType.AsParenthesizedTypeNode().Type
	}
	return utils.TrimmedNodeText(ctx.SourceFile, elementType)
}

// typeArgumentsRange returns the source range of a type-argument list including
// its angle brackets (`<T>`), matching what upstream's ESTree node covers. The
// tsgo node-list range starts after `<` and ends before `>`, so widen it
// outward to the enclosing bracket pair.
func typeArgumentsRange(ctx rule.RuleContext, typeArguments *ast.NodeList) (int, int, bool) {
	return utils.RangeEnclosingDelimiters(
		ctx.SourceFile.Text(), typeArguments.Pos(), typeArguments.End(), '<', '>',
	)
}

// hasCommentsOutsideNode mirrors upstream's `hasCommentsOutsideNode`: whether
// any comment inside `node` lies outside `child`. Used to reject a type rewrite
// that would drop a comment.
func hasCommentsOutsideNode(ctx rule.RuleContext, node *ast.Node, child *ast.Node) bool {
	nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	childRange := utils.TrimNodeTextRange(ctx.SourceFile, child)
	return hasCommentsOutsideRange(ctx, nodeRange.Pos(), nodeRange.End(), childRange.Pos(), childRange.End())
}

// hasCommentsOutsideRange is hasCommentsOutsideNode over raw offsets, for
// callers whose "child" span is not a single node (the `<…>` type-argument
// list, whose brackets are not part of any node's range).
func hasCommentsOutsideRange(ctx rule.RuleContext, nodeStart, nodeEnd, childStart, childEnd int) bool {
	for _, comment := range ctx.Comments.All() {
		if comment.Pos() < nodeStart || comment.End() > nodeEnd {
			continue
		}
		if comment.Pos() < childStart || comment.End() > childEnd {
			return true
		}
	}
	return false
}

// --- static value model -------------------------------------------------

type staticValueKind int

const (
	svUnknown staticValueKind = iota
	svString
	svNumber
	svBigInt
	svBool
	svNull
	svUndefined
)

// staticValue is a statically-evaluated comparable JavaScript value, modeling
// the subset of eslint-utils' getStaticValue that prefer-set-has consumes:
// numbers (for array sizes) and comparable primitives / null (for uniqueness).
type staticValue struct {
	kind    staticValueKind
	str     string
	number  float64
	big     *big.Int
	boolean bool
}

func (v staticValue) asNumber() (float64, bool) {
	if v.kind == svNumber {
		return v.number, true
	}
	return 0, false
}

// isComparable mirrors upstream's `isComparableStaticValue`: null, undefined,
// or any non-object / non-function primitive.
func (v staticValue) isComparable() bool {
	switch v.kind {
	case svString, svNumber, svBigInt, svBool, svNull, svUndefined:
		return true
	}
	return false
}

// isNegativeZero reports whether the value is exactly -0, which upstream
// excludes via `!Object.is(result.value, -0)`.
func (v staticValue) isNegativeZero() bool {
	return v.kind == svNumber && v.number == 0 && math.Signbit(v.number)
}

// sameValueZero mirrors upstream's `isSameValueZero`: SameValueZero equality
// (=== plus NaN equal to NaN).
func sameValueZero(left, right staticValue) bool {
	if left.kind != right.kind {
		return false
	}
	switch left.kind {
	case svString:
		return left.str == right.str
	case svNumber:
		if math.IsNaN(left.number) && math.IsNaN(right.number) {
			return true
		}
		return left.number == right.number
	case svBigInt:
		if left.big == nil || right.big == nil {
			return left.big == right.big
		}
		return left.big.Cmp(right.big) == 0
	case svBool:
		return left.boolean == right.boolean
	case svNull:
		return true
	case svUndefined:
		return true
	}
	return false
}

// evaluateStaticValue mirrors the subset of eslint-utils' getStaticValue that
// this rule needs. It resolves `const` identifier initializers through
// ctx.Refs, and evaluates literals and simple arithmetic. It returns ok=false
// when the value cannot be statically determined.
func evaluateStaticValue(ctx rule.RuleContext, node *ast.Node) (staticValue, bool) {
	return evaluateStaticValueInner(ctx, node, map[*ast.Symbol]bool{})
}

func evaluateStaticValueInner(ctx rule.RuleContext, node *ast.Node, visited map[*ast.Symbol]bool) (staticValue, bool) {
	node = ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKAssertions)
	if node == nil {
		return staticValue{}, false
	}

	switch node.Kind {
	case ast.KindStringLiteral:
		return staticValue{kind: svString, str: node.AsStringLiteral().Text}, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return staticValue{kind: svString, str: node.AsNoSubstitutionTemplateLiteral().Text}, true
	case ast.KindTrueKeyword:
		return staticValue{kind: svBool, boolean: true}, true
	case ast.KindFalseKeyword:
		return staticValue{kind: svBool, boolean: false}, true
	case ast.KindNullKeyword:
		return staticValue{kind: svNull}, true
	case ast.KindNumericLiteral:
		return numberFromText(utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text))
	case ast.KindBigIntLiteral:
		text := utils.NormalizeBigIntLiteral(node.AsBigIntLiteral().Text)
		bi := new(big.Int)
		if _, ok := bi.SetString(text, 10); !ok {
			return staticValue{}, false
		}
		return staticValue{kind: svBigInt, big: bi}, true
	case ast.KindPrefixUnaryExpression:
		return evaluateStaticPrefix(ctx, node.AsPrefixUnaryExpression(), visited)
	case ast.KindBinaryExpression:
		return evaluateStaticBinary(ctx, node.AsBinaryExpression(), visited)
	case ast.KindTemplateExpression:
		return evaluateStaticTemplate(ctx, node, visited)
	case ast.KindPropertyAccessExpression:
		return evaluateStaticPropertyAccess(node)
	case ast.KindIdentifier:
		return evaluateStaticIdentifier(ctx, node, visited)
	}
	return staticValue{}, false
}

// toStringValue mirrors ECMAScript ToString for the primitive kinds this
// evaluator models, matching what utils.StaticStringEvaluator's
// staticValueToString produces for the same inputs.
//
// Numbers are restricted to the plain-decimal band. Inside it — |x| in
// [1e-6, 1e21) plus zero — Go's shortest-decimal 'f' formatting reproduces
// Number::toString exactly (1e20 gives "100000000000000000000", 1e-6 gives
// "0.000001", 9007199254740991 stays unrounded). Outside it JS switches to
// exponential notation ("1e+21", "1e-7") in a format 'f' does not produce, so
// this returns ok=false rather than emit a string JS would never produce.
//
// The strictness matters because the result feeds element-uniqueness
// comparison: a number that stringified differently from JS could make two
// elements JS considers equal look distinct, the array would be judged
// unique, and the emitted `new Set(...)` would silently drop one of them.
// Bailing only costs a missed report.
func (v staticValue) toStringValue() (string, bool) {
	switch v.kind {
	case svString:
		return v.str, true
	case svBool:
		if v.boolean {
			return "true", true
		}
		return "false", true
	case svNull:
		return "null", true
	case svUndefined:
		return "undefined", true
	case svBigInt:
		if v.big == nil {
			return "", false
		}
		return v.big.String(), true
	case svNumber:
		switch {
		case math.IsNaN(v.number):
			return "NaN", true
		case math.IsInf(v.number, 1):
			return "Infinity", true
		case math.IsInf(v.number, -1):
			return "-Infinity", true
		case v.number == 0:
			return "0", true // ToString(-0) is "0"
		}
		if magnitude := math.Abs(v.number); magnitude < 1e-6 || magnitude >= 1e21 {
			return "", false
		}
		return strconv.FormatFloat(v.number, 'f', -1, 64), true
	}
	return "", false
}

func numberFromText(text string) (staticValue, bool) {
	switch text {
	case "Infinity":
		return staticValue{kind: svNumber, number: math.Inf(1)}, true
	case "-Infinity":
		return staticValue{kind: svNumber, number: math.Inf(-1)}, true
	}
	f, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return staticValue{}, false
	}
	return staticValue{kind: svNumber, number: f}, true
}

func evaluateStaticPrefix(ctx rule.RuleContext, prefix *ast.PrefixUnaryExpression, visited map[*ast.Symbol]bool) (staticValue, bool) {
	operand, ok := evaluateStaticValueInner(ctx, prefix.Operand, visited)
	if !ok {
		return staticValue{}, false
	}
	switch prefix.Operator {
	case ast.KindMinusToken:
		if operand.kind == svNumber {
			return staticValue{kind: svNumber, number: -operand.number}, true
		}
		if operand.kind == svBigInt && operand.big != nil {
			return staticValue{kind: svBigInt, big: new(big.Int).Neg(operand.big)}, true
		}
	case ast.KindPlusToken:
		if operand.kind == svNumber {
			return operand, true
		}
	case ast.KindTildeToken:
		if operand.kind == svNumber {
			return staticValue{kind: svNumber, number: float64(^int32(operand.number))}, true
		}
	case ast.KindExclamationToken:
		truthy, tok := operand.truthy()
		if tok {
			return staticValue{kind: svBool, boolean: !truthy}, true
		}
	}
	return staticValue{}, false
}

func evaluateStaticBinary(ctx rule.RuleContext, binary *ast.BinaryExpression, visited map[*ast.Symbol]bool) (staticValue, bool) {
	if binary.OperatorToken == nil {
		return staticValue{}, false
	}
	left, lok := evaluateStaticValueInner(ctx, binary.Left, visited)
	right, rok := evaluateStaticValueInner(ctx, binary.Right, visited)
	if !lok || !rok {
		return staticValue{}, false
	}
	// `+` with a string operand concatenates; every other operator we model is
	// numeric. Both operands are already primitives here, so ToPrimitive is a
	// no-op and ToString decides the result.
	if binary.OperatorToken.Kind == ast.KindPlusToken &&
		(left.kind == svString || right.kind == svString) {
		return concatStaticValues(left, right)
	}

	if left.kind != svNumber || right.kind != svNumber {
		return staticValue{}, false
	}
	switch binary.OperatorToken.Kind {
	case ast.KindPlusToken:
		return staticValue{kind: svNumber, number: left.number + right.number}, true
	case ast.KindMinusToken:
		return staticValue{kind: svNumber, number: left.number - right.number}, true
	case ast.KindAsteriskToken:
		return staticValue{kind: svNumber, number: left.number * right.number}, true
	case ast.KindSlashToken:
		return staticValue{kind: svNumber, number: left.number / right.number}, true
	case ast.KindPercentToken:
		return staticValue{kind: svNumber, number: math.Mod(left.number, right.number)}, true
	case ast.KindAsteriskAsteriskToken:
		return staticValue{kind: svNumber, number: math.Pow(left.number, right.number)}, true
	}
	return staticValue{}, false
}

// concatStaticValues joins two primitives with JS string-concatenation
// semantics, mirroring utils.StaticStringEvaluator's helper of the same name.
func concatStaticValues(left, right staticValue) (staticValue, bool) {
	leftText, lok := left.toStringValue()
	rightText, rok := right.toStringValue()
	if !lok || !rok {
		return staticValue{}, false
	}
	return staticValue{kind: svString, str: leftText + rightText}, true
}

// evaluateStaticTemplate folds a template literal with statically-known
// substitutions, mirroring utils.StaticStringEvaluator's evalTemplateExpression.
func evaluateStaticTemplate(ctx rule.RuleContext, node *ast.Node, visited map[*ast.Symbol]bool) (staticValue, bool) {
	template := node.AsTemplateExpression()
	if template == nil {
		return staticValue{}, false
	}

	var builder strings.Builder
	if template.Head != nil {
		builder.WriteString(template.Head.Text())
	}
	if template.TemplateSpans != nil {
		for _, spanNode := range template.TemplateSpans.Nodes {
			span := spanNode.AsTemplateSpan()
			if span == nil {
				return staticValue{}, false
			}
			value, ok := evaluateStaticValueInner(ctx, span.Expression, visited)
			if !ok {
				return staticValue{}, false
			}
			text, ok := value.toStringValue()
			if !ok {
				return staticValue{}, false
			}
			builder.WriteString(text)
			if span.Literal != nil {
				builder.WriteString(span.Literal.Text())
			}
		}
	}
	return staticValue{kind: svString, str: builder.String()}, true
}

// wellKnownNumberConstants are the `Number.*` / `Math.*` members that
// eslint-utils' getStaticValue resolves and that can appear as an array
// element. Keyed by object name, then property name.
var wellKnownNumberConstants = map[string]map[string]float64{
	"Number": {
		"NaN":               math.NaN(),
		"POSITIVE_INFINITY": math.Inf(1),
		"NEGATIVE_INFINITY": math.Inf(-1),
		"MAX_SAFE_INTEGER":  9007199254740991,
		"MIN_SAFE_INTEGER":  -9007199254740991,
		"MAX_VALUE":         math.MaxFloat64,
		"MIN_VALUE":         5e-324,
		"EPSILON":           2.220446049250313e-16,
	},
	"Math": {
		"PI":      math.Pi,
		"E":       math.E,
		"LN2":     math.Ln2,
		"LN10":    math.Log(10),
		"LOG2E":   math.Log2E,
		"LOG10E":  math.Log10E,
		"SQRT2":   math.Sqrt2,
		"SQRT1_2": math.Sqrt(0.5),
	},
}

// evaluateStaticPropertyAccess resolves the well-known numeric globals. Bare
// `NaN` / `Infinity` already fold via evaluateStaticIdentifier; this covers the
// `Number.NaN` spelling of the same values, which upstream also folds.
func evaluateStaticPropertyAccess(node *ast.Node) (staticValue, bool) {
	access := node.AsPropertyAccessExpression()
	if access == nil || access.QuestionDotToken != nil {
		return staticValue{}, false
	}

	object := ast.SkipOuterExpressions(access.Expression, ast.OEKParentheses|ast.OEKAssertions)
	if object == nil || !ast.IsIdentifier(object) {
		return staticValue{}, false
	}
	objectName := object.AsIdentifier().Text
	members, known := wellKnownNumberConstants[objectName]
	if !known || utils.IsShadowed(object, objectName) {
		return staticValue{}, false
	}

	property := access.Name()
	if property == nil || !ast.IsIdentifier(property) {
		return staticValue{}, false
	}
	value, found := members[property.AsIdentifier().Text]
	if !found {
		return staticValue{}, false
	}
	return staticValue{kind: svNumber, number: value}, true
}

func evaluateStaticIdentifier(ctx rule.RuleContext, node *ast.Node, visited map[*ast.Symbol]bool) (staticValue, bool) {
	// The global numeric/undefined identifiers upstream's getStaticValue models.
	// Each is only a constant when not shadowed by a local binding of the same
	// name (e.g. `const NaN = 0`), matching the Infinity shadow check.
	switch node.AsIdentifier().Text {
	case "Infinity":
		if !utils.IsShadowed(node, "Infinity") {
			return staticValue{kind: svNumber, number: math.Inf(1)}, true
		}
	case "NaN":
		if !utils.IsShadowed(node, "NaN") {
			return staticValue{kind: svNumber, number: math.NaN()}, true
		}
	case "undefined":
		if !utils.IsShadowed(node, "undefined") {
			return staticValue{kind: svUndefined}, true
		}
	}
	if ctx.Refs == nil {
		return staticValue{}, false
	}
	symbol := ctx.Refs.Resolve(node)
	if symbol == nil || visited[symbol] || len(symbol.Declarations) != 1 {
		return staticValue{}, false
	}
	declarationNode := symbol.Declarations[0]
	if declarationNode == nil || declarationNode.Kind != ast.KindVariableDeclaration {
		return staticValue{}, false
	}
	declaration := declarationNode.AsVariableDeclaration()
	if declaration.Initializer == nil {
		return staticValue{}, false
	}
	declarationList := declarationNode.Parent
	if declarationList == nil || declarationList.Kind != ast.KindVariableDeclarationList ||
		!ast.IsVarConst(declarationList) {
		return staticValue{}, false
	}
	visited[symbol] = true
	defer delete(visited, symbol)
	return evaluateStaticValueInner(ctx, declaration.Initializer, visited)
}

// truthy reports JS truthiness for the kinds we model.
func (v staticValue) truthy() (bool, bool) {
	switch v.kind {
	case svString:
		return v.str != "", true
	case svNumber:
		return v.number != 0 && !math.IsNaN(v.number), true
	case svBigInt:
		return v.big != nil && v.big.Sign() != 0, true
	case svBool:
		return v.boolean, true
	case svNull:
		return false, true
	case svUndefined:
		return false, true
	}
	return false, false
}
