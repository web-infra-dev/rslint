package utils

import (
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/collections"
	"github.com/microsoft/TypeScript/tsc/shim/evaluator"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// StaticStringEvaluator folds expressions to static JavaScript values and exposes
// string/property-name conversions to callers. It wraps tsgo's evaluator and adds
// stable local variable resolution through the TypeChecker. Internal values such
// as Symbols stay typed until a conversion boundary instead of being stringified.
// Create one evaluator per linted file; it keeps write-reference and recursion
// state for that file.
type StaticStringEvaluator struct {
	// GlobalAccess optionally supplies the rule context's configured globals.
	// When nil, unshadowed standard globals are assumed to be available.
	GlobalAccess           func(string) GlobalAccess
	typeChecker            *checker.Checker
	sourceFile             *ast.SourceFile
	referenceResolver      StaticReferenceResolver
	evaluator              evaluator.Evaluator
	resolveIdentifiers     bool
	resolving              map[*ast.Symbol]bool
	evaluation             *staticEvaluationState
	referenceFlagsComputed bool
	referenceFlags         map[*ast.Symbol]staticReferenceFlags
	bindingTransfers       map[*ast.Symbol][]staticReferenceTransfer
	storedReferences       map[*ast.Symbol][]staticReferenceTransfer
	markedReferences       map[*ast.Node]staticReferenceFlags
	trackCallEscapes       bool
	escapeFacts            *staticEscapeFacts
	capturePoint           *ast.Node
	arrayPrototypeMutated  bool
	stringPrototypeMutated bool
}

func NewStaticStringEvaluator(typeChecker *checker.Checker) *StaticStringEvaluator {
	return NewStaticStringEvaluatorWithSourceFile(typeChecker, nil)
}

func NewStaticStringEvaluatorWithSourceFile(typeChecker *checker.Checker, sourceFile *ast.SourceFile) *StaticStringEvaluator {
	return NewStaticStringEvaluatorWithReferenceResolver(typeChecker, sourceFile, nil)
}

// StaticReferenceResolver is the subset of rule.RefStore used by the evaluator.
// Supplying it reuses the linter's binder-based symbol resolution while the
// evaluator performs its single mutation scan.
type StaticReferenceResolver interface {
	Resolve(node *ast.Node) *ast.Symbol
}

func NewStaticStringEvaluatorWithReferenceResolver(
	typeChecker *checker.Checker,
	sourceFile *ast.SourceFile,
	referenceResolver StaticReferenceResolver,
) *StaticStringEvaluator {
	staticEvaluator := &StaticStringEvaluator{
		typeChecker:        typeChecker,
		sourceFile:         sourceFile,
		referenceResolver:  referenceResolver,
		resolveIdentifiers: true,
		resolving:          map[*ast.Symbol]bool{},
	}
	staticEvaluator.evaluator = evaluator.NewEvaluator(staticEvaluator.evaluateEntity, ast.OEKAssertions)
	return staticEvaluator
}

// NewStaticStringEvaluatorWithCallEscapes adds conservative object-escape
// tracking for consumers that need to trust a computed key after earlier calls.
// Ordinary evaluators retain their existing contract for consumers evaluating
// arguments before the inspected call executes.
func NewStaticStringEvaluatorWithCallEscapes(typeChecker *checker.Checker, sourceFile *ast.SourceFile, referenceResolver StaticReferenceResolver) *StaticStringEvaluator {
	staticEvaluator := NewStaticStringEvaluatorWithReferenceResolver(typeChecker, sourceFile, referenceResolver)
	staticEvaluator.trackCallEscapes = true
	return staticEvaluator
}

// NewStaticStringEvaluatorWithoutScope mirrors eslint-utils getStaticValue
// when no initial scope is supplied: literal expressions can still be folded,
// but identifiers (including built-in globals) cannot be resolved.
func NewStaticStringEvaluatorWithoutScope() *StaticStringEvaluator {
	staticEvaluator := NewStaticStringEvaluator(nil)
	staticEvaluator.resolveIdentifiers = false
	return staticEvaluator
}

type staticNullValue struct{}
type staticUndefinedValue struct{}

// Keep folded numbers in tsgo's representation, including its JS arithmetic.
type staticNumberValue = jsnum.Number

// staticStringNode keeps literal strings backed by their existing AST node so
// nested aggregate evaluation doesn't allocate an interface box per literal.
type staticStringNode ast.Node

// staticRegExpNode keeps a regular-expression literal backed by its AST node.
// It is private because callers only observe its static properties or String
// conversion, never a host-language regexp object.
type staticRegExpNode ast.Node

// staticObjectValue and staticArrayValue hold folded object and array literals
// so member access and JavaScript's default string coercion can be modeled.
type staticObjectValue struct {
	property        staticObjectProperty
	extraProperties []staticObjectProperty
	propertyCount   int
	prototype       any
	prototypeSet    bool
}

type staticObjectProperty struct {
	name  string
	value any
}

type staticArrayValue struct {
	length int
	// Nil slots are holes; reading or iterating a hole produces undefined.
	inline   [2]any
	overflow []any
}

type staticReferenceFlags uint8

const (
	staticReferenceWrite staticReferenceFlags = 1 << iota
	staticReferencePropertyMutation
	// A nested write can reach values held in a fresh object or array. A write
	// to the container itself cannot mutate values copied into that container.
	staticReferenceNestedMutation
	// An unknown call can retain references, invoke callbacks, and mutate
	// objects reachable through the values it receives.
	staticReferenceCallEscape
	staticReferenceCallReceiver
	// Deletions and prototype/computed writes can invalidate data properties.
	staticReferencePropertyShapeMutation
	// Assigning an array's length can delete its indexed data properties.
	staticReferenceArrayLengthMutation
)

// objectPassThroughMethods are the `Object` methods that return their argument
// unchanged, so folding can see through them.
var objectPassThroughMethods = map[string]bool{
	"freeze":            true,
	"preventExtensions": true,
	"seal":              true,
}

// Static evaluation must not turn a compact expression into an unbounded
// allocation. The limit is deliberately much larger than a useful diagnostic
// message while still bounding nested Array#join coercion.
const maxStaticStringLength = 1 << 20

// Nested spreads can expand a short expression into a much larger array.
const maxStaticArrayLength = 1 << 16

const maxStaticObjectProperties = 1 << 16

type staticEvalResult struct {
	value any
	ok    bool
}

// StaticEvalValueKind classifies a static evaluation result for callers that
// only need to distinguish strings from known non-strings and unknown values.
type StaticEvalValueKind uint8

const (
	StaticEvalUnknown StaticEvalValueKind = iota
	StaticEvalString
	StaticEvalNonString
)

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
	return staticValueAsString(result.value)
}

// EvalToString returns JavaScript's String conversion of a statically known
// value. This matches eslint-utils getStringIfConstant after getStaticValue has
// evaluated an expression.
func (staticEvaluator *StaticStringEvaluator) EvalToString(node *ast.Node) (string, bool) {
	if staticEvaluator == nil || node == nil {
		return "", false
	}
	result := staticEvaluator.evalValue(node)
	if !result.ok {
		return "", false
	}
	return staticValueToString(result.value)
}

// EvalAccessExpressionName returns the string property name of a member access.
// Computed keys use this evaluator's scope and JavaScript String conversion;
// private names and unknown keys have no string name. Use a no-scope evaluator
// for eslint-utils ReferenceTracker semantics, which never resolves key identifiers.
func (staticEvaluator *StaticStringEvaluator) EvalAccessExpressionName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	if node.Kind == ast.KindElementAccessExpression {
		return staticEvaluator.EvalToString(node.AsElementAccessExpression().ArgumentExpression)
	}
	return AccessExpressionStaticName(node)
}

// EvalPropertyName returns the string name of an object/class property or
// binding key. It extends GetStaticPropertyName with computed-expression
// evaluation and the template keys unwrapped by tsgo's binding helpers.
// Non-computed identifiers are names, never variable references.
func (staticEvaluator *StaticStringEvaluator) EvalPropertyName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindComputedPropertyName:
		return staticEvaluator.EvalToString(node.AsComputedPropertyName().Expression)
	case ast.KindNoSubstitutionTemplateLiteral:
		return GetStaticExpressionValue(node)
	default:
		return GetStaticPropertyName(node)
	}
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
	if value, ok := staticValueAsString(result.value); ok {
		return value, true
	}
	return result.value, true
}

// StaticPlainObjectOnlyProperty reads the sole own property of a folded object
// with the default prototype. Repeated keys count as one property, with the
// value of the last assignment.
func StaticPlainObjectOnlyProperty(value any, name string) (any, bool) {
	object, ok := value.(*staticObjectValue)
	if !ok || object.propertyCount == 0 || object.prototypeSet ||
		ecmascript.CompareStrings(object.property.name, name) != 0 {
		return nil, false
	}
	propertyValue := object.property.value
	for _, property := range object.extraProperties {
		if ecmascript.CompareStrings(property.name, name) != 0 {
			return nil, false
		}
		propertyValue = property.value
	}
	if text, ok := staticValueAsString(propertyValue); ok {
		return text, true
	}
	return propertyValue, true
}

// EvalArrayValue classifies a statically known value by whether it is an
// array. The second result is false when the expression cannot be folded. It
// keeps the evaluator's private aggregate representation encapsulated while
// allowing callers that mirror eslint-utils' getStaticValue + Array.isArray
// pattern to distinguish arrays from other known values.
func (staticEvaluator *StaticStringEvaluator) EvalArrayValue(node *ast.Node) (isArray bool, known bool) {
	if staticEvaluator == nil || node == nil {
		return false, false
	}
	result := staticEvaluator.evalValue(node)
	if !result.ok {
		return false, false
	}
	_, isArray = result.value.(*staticArrayValue)
	return isArray, true
}

// EvalStringValue returns a string without boxing it into an interface, or
// classifies the result as a known non-string or an unknown value.
func (staticEvaluator *StaticStringEvaluator) EvalStringValue(node *ast.Node) (string, StaticEvalValueKind) {
	if staticEvaluator == nil || node == nil {
		return "", StaticEvalUnknown
	}

	node = SkipAssertionsAndParens(node)
	if node == nil {
		return "", StaticEvalUnknown
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text, StaticEvalString
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text, StaticEvalString
	case ast.KindCallExpression:
		if ast.IsOptionalChain(node) {
			break
		}
		if value, matched, ok := staticEvaluator.evalArrayJoin(node); matched {
			if ok {
				return value, StaticEvalString
			}
			return "", StaticEvalUnknown
		}
	}

	result := staticEvaluator.evalValue(node)
	if !result.ok {
		return "", StaticEvalUnknown
	}
	if value, ok := staticValueAsString(result.value); ok {
		return value, StaticEvalString
	}
	return "", StaticEvalNonString
}

func (staticEvaluator *StaticStringEvaluator) evalUncachedValue(node *ast.Node) staticEvalResult {
	if staticEvaluator.evaluation.optionalChainShortCircuits(node, staticEvaluator) {
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	}

	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return staticEvalResult{value: (*staticStringNode)(node), ok: true}
	case ast.KindRegularExpressionLiteral:
		return staticEvalResult{value: (*staticRegExpNode)(node), ok: true}
	case ast.KindBigIntLiteral:
		text := strings.TrimSuffix(node.Text(), "n")
		if len(text) > maxStaticBigIntBits {
			return staticEvalResult{}
		}
		value, ok := ecmascript.StringToBigInt(text)
		return staticEvalResult{value: value, ok: ok && value.BitLen() <= maxStaticBigIntBits}
	case ast.KindTrueKeyword:
		return staticEvalResult{value: true, ok: true}
	case ast.KindFalseKeyword:
		return staticEvalResult{value: false, ok: true}
	case ast.KindNullKeyword:
		return staticEvalResult{value: staticNullValue{}, ok: true}
	case ast.KindUndefinedKeyword:
		if !staticEvaluator.resolveIdentifiers {
			return staticEvalResult{}
		}
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	case ast.KindIdentifier:
		return staticEvaluator.evalIdentifier(node)
	case ast.KindTemplateExpression:
		return staticEvaluator.evalTemplateExpression(node, false)
	case ast.KindBinaryExpression:
		return staticEvaluator.evalBinaryExpression(node)
	case ast.KindPrefixUnaryExpression:
		return staticEvaluator.evalPrefixUnaryExpression(node)
	case ast.KindConditionalExpression:
		return staticEvaluator.evalConditionalExpression(node)
	case ast.KindVoidExpression:
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	case ast.KindTypeOfExpression:
		operand := staticEvaluator.evalValue(node.Expression())
		if !operand.ok {
			return staticEvalResult{}
		}
		var name string
		switch staticValueKindOf(operand.value) {
		case staticKindString:
			name = "string"
		case staticKindNumber:
			name = "number"
		case staticKindBigInt:
			name = "bigint"
		case staticKindBoolean:
			name = "boolean"
		case staticKindUndefined:
			name = "undefined"
		case staticKindNull:
			name = "object"
		default:
			if staticValueIsAggregate(operand.value) {
				name = "object"
			}
		}
		return staticEvalResult{value: name, ok: name != ""}
	case ast.KindObjectLiteralExpression:
		if result := staticEvaluator.evalObjectLiteral(node); result.ok {
			return result
		}
	case ast.KindArrayLiteralExpression:
		if result := staticEvaluator.evalArrayLiteral(node); result.ok {
			return result
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if result := staticEvaluator.evalWellKnownSymbolMember(node); result.ok {
			return result
		}
		if result := staticEvaluator.evalMemberAccess(node); result.ok {
			return result
		}
	case ast.KindCallExpression:
		if result := staticEvaluator.evalSymbolCall(node); result.ok {
			return result
		}
		if result := staticEvaluator.evalNumericBuiltinCall(node); result.ok {
			return result
		}
		if result := staticEvaluator.evalBuiltinStaticCall(node); result.ok {
			return result
		}
		if result := staticEvaluator.evalStringCall(node); result.ok {
			return result
		}
		if result := staticEvaluator.evalArrayJoinCall(node); result.ok {
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
	if !staticEvaluator.resolveIdentifiers {
		return staticEvalResult{}
	}
	if ast.IsIdentifier(node) {
		name := node.Text()
		if (name == "undefined" || name == "Infinity" || name == "NaN") && staticEvaluator.isBuiltinIdentifier(node, name) {
			switch name {
			case "undefined":
				return staticEvalResult{value: staticUndefinedValue{}, ok: true}
			case "Infinity":
				return staticEvalResult{value: staticNumberValue(math.Inf(1)), ok: true}
			case "NaN":
				return staticEvalResult{value: staticNumberValue(math.NaN()), ok: true}
			}
		}
	}
	initializer, symbol, ok := staticEvaluator.resolveIdentifierInitializer(node)
	if !ok || staticEvaluator.resolving[symbol] {
		return staticEvalResult{}
	}

	staticEvaluator.resolving[symbol] = true
	defer delete(staticEvaluator.resolving, symbol)
	if isAggregateLiteral(initializer) && staticEvaluator.hasPropertyMutation(symbol) {
		return staticEvalResult{}
	}
	if result := staticEvaluator.capturedInitializer(initializer, symbol); result.ok {
		return result
	}
	result := staticEvaluator.evalValue(initializer)
	if result.ok && staticValueIsAggregate(result.value) &&
		!isAggregateLiteral(initializer) && staticEvaluator.hasPropertyMutation(symbol) {
		return staticEvalResult{}
	}
	return result
}

func isAggregateLiteral(node *ast.Node) bool {
	node = SkipAssertionsAndParens(node)
	return node != nil && (node.Kind == ast.KindObjectLiteralExpression ||
		node.Kind == ast.KindArrayLiteralExpression ||
		node.Kind == ast.KindRegularExpressionLiteral)
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
	if staticEvaluator == nil || (staticEvaluator.typeChecker == nil && staticEvaluator.referenceResolver == nil) {
		return nil, nil, false
	}

	expr := SkipAssertionsAndParens(node)
	if expr == nil || expr.Kind != ast.KindIdentifier {
		return nil, nil, false
	}

	symbol := staticEvaluator.referenceSymbol(expr)
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

func (staticEvaluator *StaticStringEvaluator) evalTemplateExpression(node *ast.Node, raw bool) staticEvalResult {
	template := node.AsTemplateExpression()
	if template == nil {
		return staticEvalResult{}
	}

	var builder strings.Builder
	if template.Head != nil {
		builder.WriteString(templateLiteralText(template.Head, raw))
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
				builder.WriteString(templateLiteralText(span.Literal, raw))
			}
		}
	}
	return staticEvalResult{value: ecmascript.CombineSurrogatePairs(builder.String()), ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evalBinaryExpression(node *ast.Node) staticEvalResult {
	binary := node.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return staticEvalResult{}
	}

	switch binary.OperatorToken.Kind {
	case ast.KindCommaToken:
		// A sequence expression evaluates to its final operand. This mirrors
		// eslint-utils getStaticValue for the expression shapes that consume the
		// result rather than its side effects.
		return staticEvaluator.evalValue(binary.Right)
	case ast.KindEqualsToken:
		result := staticEvaluator.evalValue(binary.Right)
		if result.ok && staticValueIsAggregate(result.value) {
			// The assigned aggregate can be mutated by another operand or call
			// argument before the containing expression consumes it.
			return staticEvalResult{}
		}
		return result
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
	case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
		left := staticEvaluator.evalValue(binary.Left)
		right := staticEvaluator.evalValue(binary.Right)
		if !left.ok || !right.ok {
			return staticEvalResult{}
		}
		equal, ok := staticValuesStrictEqual(left.value, right.value)
		if !ok {
			return staticEvalResult{}
		}
		if binary.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
			equal = !equal
		}
		return staticEvalResult{value: equal, ok: true}
	case ast.KindPlusToken:
		left := staticEvaluator.evalValue(binary.Left)
		right := staticEvaluator.evalValue(binary.Right)
		if !left.ok || !right.ok {
			return staticEvalResult{}
		}
		return staticEvaluator.evalBinaryValues(binary.OperatorToken.Kind, left.value, right.value)
	}

	if result := staticEvaluator.evalWithTsgo(node); result.ok {
		return result
	}
	left := staticEvaluator.evalValue(binary.Left)
	right := staticEvaluator.evalValue(binary.Right)
	if left.ok && right.ok {
		return staticEvaluator.evalBinaryValues(binary.OperatorToken.Kind, left.value, right.value)
	}
	return staticEvalResult{}
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
	return staticEvalResult{value: ecmascript.CombineSurrogatePairs(leftString + rightString), ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evalPrefixUnaryExpression(node *ast.Node) staticEvalResult {
	prefix := node.AsPrefixUnaryExpression()
	if prefix == nil {
		return staticEvaluator.evalWithTsgo(node)
	}

	switch prefix.Operator {
	case ast.KindExclamationToken:
		operand := staticEvaluator.evalValue(prefix.Operand)
		if !operand.ok {
			return staticEvalResult{}
		}
		truthy, ok := staticValueTruthy(operand.value)
		if !ok {
			return staticEvalResult{}
		}
		return staticEvalResult{value: !truthy, ok: true}
	case ast.KindPlusToken, ast.KindMinusToken, ast.KindTildeToken:
		if result := staticEvaluator.evalNumericPrefix(prefix); result.ok {
			return result
		}
	}

	return staticEvaluator.evalWithTsgo(node)
}

// evalNumericPrefix folds `+`, `-` and `~` over any static operand. tsgo's
// evaluator folds these operators only when the operand is already a number,
// so string, boolean and nullish operands reach here unfolded.
func (staticEvaluator *StaticStringEvaluator) evalNumericPrefix(prefix *ast.PrefixUnaryExpression) staticEvalResult {
	operand := staticEvaluator.evalValue(prefix.Operand)
	if !operand.ok {
		return staticEvalResult{}
	}
	if value, ok := operand.value.(*big.Int); ok {
		switch prefix.Operator {
		case ast.KindMinusToken:
			return staticEvalResult{value: new(big.Int).Neg(value), ok: true}
		case ast.KindTildeToken:
			return staticEvalResult{value: new(big.Int).Not(value), ok: true}
		default:
			// Unary + applies ToNumber, which throws for a BigInt.
			return staticEvalResult{}
		}
	}
	number, ok := staticValueToNumber(operand.value)
	if !ok {
		return staticEvalResult{}
	}

	switch prefix.Operator {
	case ast.KindPlusToken:
		return staticEvalResult{value: staticNumberValue(number), ok: true}
	case ast.KindMinusToken:
		return staticEvalResult{value: staticNumberValue(-number), ok: true}
	case ast.KindTildeToken:
		return staticEvalResult{value: staticNumberValue(^toInt32(number)), ok: true}
	}
	return staticEvalResult{}
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

	value := &staticObjectValue{}
	for _, property := range object.Properties.Nodes {
		var valueNode *ast.Node
		switch property.Kind {
		case ast.KindPropertyAssignment:
			valueNode = property.AsPropertyAssignment().Initializer
		case ast.KindShorthandPropertyAssignment:
			valueNode = property.Name()
		case ast.KindSpreadAssignment:
			spread := staticEvaluator.evalValue(property.AsSpreadAssignment().Expression)
			if !spread.ok || !value.copyDataProperties(spread.value) {
				return staticEvalResult{}
			}
			continue
		default:
			return staticEvalResult{}
		}

		if isObjectLiteralPrototypeSetter(property) {
			propertyValue := staticEvaluator.evalValue(valueNode)
			if !propertyValue.ok {
				return staticEvalResult{}
			}
			switch propertyValue.value.(type) {
			case *staticRegExpNode, *staticObjectValue, *staticArrayValue:
				value.prototype = propertyValue.value
				value.prototypeSet = true
			case staticNullValue:
				value.prototype = nil
				value.prototypeSet = true
			default:
				// JavaScript ignores primitive prototype-setter values.
			}
			continue
		}

		key, ok := staticEvaluator.evalPropertyKey(property.Name())
		if !ok {
			return staticEvalResult{}
		}
		propertyValue := staticEvaluator.evalValue(valueNode)
		if !propertyValue.ok {
			return staticEvalResult{}
		}
		value.addProperty(key, propertyValue.value)
	}
	return staticEvalResult{value: value, ok: true}
}

func (value *staticObjectValue) addProperty(name string, propertyValue any) {
	property := staticObjectProperty{name: name, value: propertyValue}
	if value.propertyCount == 0 {
		value.property = property
	} else {
		value.extraProperties = append(value.extraProperties, property)
	}
	value.propertyCount++
}

// Object spread copies own enumerable data properties, not the prototype.
func (value *staticObjectValue) copyDataProperties(source any) bool {
	if text, ok := staticValueAsString(source); ok {
		units := ecmascript.StringCodeUnits(text)
		if len(units) > maxStaticObjectProperties-value.propertyCount {
			return false
		}
		for index := range units {
			value.addProperty(strconv.Itoa(index), ecmascript.StringFromCodeUnits(units[index:index+1]))
		}
		return true
	}
	switch source := source.(type) {
	case *staticObjectValue:
		if source.propertyCount > maxStaticObjectProperties-value.propertyCount {
			return false
		}
		if source.propertyCount > 0 {
			value.addProperty(source.property.name, source.property.value)
		}
		for _, property := range source.extraProperties {
			value.addProperty(property.name, property.value)
		}
	case *staticArrayValue:
		for index := range source.length {
			if element, present := source.ownElement(index); present {
				if value.propertyCount >= maxStaticObjectProperties {
					return false
				}
				value.addProperty(strconv.Itoa(index), element)
			}
		}
	case staticNullValue, staticUndefinedValue, bool, staticNumberValue, *big.Int, *staticRegExpNode:
		// These values have no own enumerable properties.
	default:
		return false
	}
	return true
}

func isObjectLiteralPrototypeSetter(property *ast.Node) bool {
	return property != nil && property.Kind == ast.KindPropertyAssignment && ast.IsProtoSetter(property.Name())
}

func (staticEvaluator *StaticStringEvaluator) evalObjectLiteralMember(node *ast.Node, key string) staticEvalResult {
	object := node.AsObjectLiteralExpression()
	if object == nil || object.Properties == nil {
		return staticEvalResult{}
	}
	if slices.ContainsFunc(object.Properties.Nodes, ast.IsSpreadAssignment) {
		value := staticEvaluator.evalObjectLiteral(node)
		if !value.ok {
			return staticEvalResult{}
		}
		return staticMemberValue(value.value, key)
	}

	var ownValue staticEvalResult
	ownFound := false
	prototypeValue := staticEvalResult{value: staticUndefinedValue{}, ok: true}
	prototypeSet := false
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

		if isObjectLiteralPrototypeSetter(property) {
			prototypeNode := SkipAssertionsAndParens(valueNode)
			if prototypeNode != nil && prototypeNode.Kind == ast.KindObjectLiteralExpression {
				prototypeValue = staticEvaluator.evalObjectLiteralMember(prototypeNode, key)
				prototypeSet = true
				continue
			}

			value := staticEvaluator.evalValue(valueNode)
			if !value.ok {
				return staticEvalResult{}
			}
			switch value.value.(type) {
			case *staticRegExpNode, *staticObjectValue, *staticArrayValue:
				prototypeValue = staticPrototypeMemberValue(value.value, key)
				prototypeSet = true
			case staticNullValue:
				prototypeValue = staticEvalResult{value: staticUndefinedValue{}, ok: true}
				prototypeSet = true
			default:
				// JavaScript ignores primitive prototype-setter values.
			}
			continue
		}

		propertyKey, ok := staticEvaluator.evalPropertyKey(property.Name())
		if !ok {
			return staticEvalResult{}
		}
		value := staticEvaluator.evalValue(valueNode)
		if !value.ok {
			return staticEvalResult{}
		}
		if ecmascript.CompareStrings(propertyKey, key) == 0 {
			ownValue = value
			ownFound = true
		}
	}

	if ownFound {
		return ownValue
	}
	if prototypeSet {
		return prototypeValue
	}
	return staticDefaultObjectMember(key)
}

func (staticEvaluator *StaticStringEvaluator) evalArrayLiteral(node *ast.Node) staticEvalResult {
	value, ok := staticEvaluator.evalArrayLiteralValue(node)
	if !ok {
		return staticEvalResult{}
	}
	return staticEvalResult{value: &value, ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evalArrayLiteralValue(node *ast.Node) (staticArrayValue, bool) {
	array := node.AsArrayLiteralExpression()
	if array == nil || array.Elements == nil {
		return staticArrayValue{}, false
	}
	if slices.ContainsFunc(array.Elements.Nodes, ast.IsSpreadElement) {
		values, ok := staticEvaluator.evalElementValues(array.Elements.Nodes)
		if !ok {
			return staticArrayValue{}, false
		}
		return *staticArrayFromValues(values), true
	}

	arrayValue := staticArrayValue{length: len(array.Elements.Nodes)}
	if arrayValue.length > len(arrayValue.inline) {
		arrayValue.overflow = make([]any, arrayValue.length-len(arrayValue.inline))
	}
	for i, element := range array.Elements.Nodes {
		if element.Kind == ast.KindOmittedExpression {
			continue
		}
		elementValue := staticEvaluator.evalValue(element)
		if !elementValue.ok {
			return staticArrayValue{}, false
		}
		arrayValue.set(i, elementValue.value)
	}
	return arrayValue, true
}

func (value *staticArrayValue) set(index int, element any) {
	if index < len(value.inline) {
		value.inline[index] = element
		return
	}
	value.overflow[index-len(value.inline)] = element
}

func (value *staticArrayValue) element(index int) any {
	if element, present := value.ownElement(index); present {
		return element
	}
	return staticUndefinedValue{}
}

func (value *staticArrayValue) ownElement(index int) (any, bool) {
	if index < len(value.inline) {
		element := value.inline[index]
		return element, element != nil
	}
	element := value.overflow[index-len(value.inline)]
	return element, element != nil
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
	objectNode := AccessExpressionObject(node)
	if objectNode == nil {
		return staticEvalResult{}
	}
	key, ok := staticEvaluator.evalAccessExpressionKey(node)
	if !ok {
		return staticEvalResult{}
	}
	if staticEvaluator.resolveIdentifiers {
		if number, ok := staticGlobalNumber(objectNode, key); ok &&
			(staticEvaluator.GlobalAccess == nil || staticEvaluator.GlobalAccess(SkipAssertionsAndParens(objectNode).Text()).IsDeclared()) {
			return staticEvalResult{value: staticNumberValue(number), ok: true}
		}
	}
	if literal := SkipAssertionsAndParens(objectNode); literal != nil && literal.Kind == ast.KindObjectLiteralExpression {
		return staticEvaluator.evalObjectLiteralMember(literal, key)
	}

	object := staticEvaluator.evalValue(objectNode)
	if !object.ok {
		if staticEvaluator.trackCallEscapes {
			return staticEvaluator.capturedOwnMember(objectNode, key)
		}
		return staticEvalResult{}
	}
	if key == "length" {
		if text, ok := staticValueAsString(object.value); ok {
			return staticEvalResult{value: staticNumberValue(ecmascript.StringCodeUnitCount(text)), ok: true}
		}
	}
	return staticMemberValue(object.value, key)
}

func (staticEvaluator *StaticStringEvaluator) evalAccessExpressionKey(node *ast.Node) (string, bool) {
	if key, ok := AccessExpressionStaticName(node); ok {
		return key, true
	}
	if node == nil || node.Kind != ast.KindElementAccessExpression {
		return "", false
	}
	argument := staticEvaluator.evalValue(node.AsElementAccessExpression().ArgumentExpression)
	if !argument.ok {
		return "", false
	}
	return staticValueToString(argument.value)
}

// staticMemberValue reads key off a folded object, array, or string literal.
// Reading a key that isn't there yields `undefined`, as it would at runtime. A
// callable key such as `join` stays unresolved; string indexing is narrower
// (a canonical index only) and
// returns the UTF-16 code unit at that index, matching how JavaScript indexes
// a string.
func staticMemberValue(object any, key string) staticEvalResult {
	switch object := object.(type) {
	case string:
		return staticStringMemberValue(object, key)
	case *staticStringNode:
		str, ok := staticValueAsString(object)
		if !ok {
			return staticEvalResult{}
		}
		return staticStringMemberValue(str, key)
	case *staticRegExpNode:
		return staticRegExpMemberValue((*ast.Node)(object), key)
	case *staticObjectValue:
		if value, ok := staticObjectOwnProperty(object, key); ok {
			return staticEvalResult{value: value, ok: true}
		}
		if object.prototypeSet {
			if object.prototype != nil {
				return staticPrototypeMemberValue(object.prototype, key)
			}
			return staticEvalResult{value: staticUndefinedValue{}, ok: true}
		}
		return staticDefaultObjectMember(key)
	case *staticArrayValue:
		if key == "length" {
			return staticEvalResult{value: staticNumberValue(object.length), ok: true}
		}
		index, ok := staticArrayIndex(key)
		if !ok {
			if staticNumberShapedKey(key) {
				return staticEvalResult{value: staticUndefinedValue{}, ok: true}
			}
			return staticEvalResult{}
		}
		if index >= object.length {
			return staticEvalResult{value: staticUndefinedValue{}, ok: true}
		}
		return staticEvalResult{value: object.element(index), ok: true}
	}
	return staticEvalResult{}
}

func staticDefaultObjectMember(key string) staticEvalResult {
	switch key {
	case "constructor", "toString", "toLocaleString", "valueOf", "hasOwnProperty",
		"isPrototypeOf", "propertyIsEnumerable", "__proto__", "__defineGetter__",
		"__defineSetter__", "__lookupGetter__", "__lookupSetter__":
		return staticEvalResult{}
	default:
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	}
}

// staticPrototypeMemberValue reads a property through an authored __proto__
// setter. RegExp's lastIndex is an own data property and is safe to inherit,
// but its standard getters live on RegExp.prototype and brand-check the
// original receiver. Calling one with the outer object would throw, so those
// reads remain unknown instead of reusing direct-literal getter evaluation.
func staticPrototypeMemberValue(prototype any, key string) staticEvalResult {
	if _, ok := prototype.(*staticRegExpNode); ok {
		if key == "lastIndex" {
			return staticEvalResult{value: staticNumberValue(0), ok: true}
		}
		return staticEvalResult{}
	}
	return staticMemberValue(prototype, key)
}

// staticRegExpMemberValue models the stable own property and getter subset
// eslint-utils exposes for a newly-created RegExp literal. `unicodeSets` is
// deliberately absent because eslint-utils 4.10.1 does not whitelist that
// getter yet.
func staticRegExpMemberValue(node *ast.Node, key string) staticEvalResult {
	pattern, flags := ExtractRegexPatternAndFlags(node.Text())
	switch key {
	case "source":
		return staticEvalResult{value: pattern, ok: true}
	case "flags":
		_, canonicalFlags := ExtractRegexPatternAndFlags(RegExpLiteralStringValue(node.Text()))
		return staticEvalResult{value: canonicalFlags, ok: true}
	case "dotAll":
		return staticEvalResult{value: strings.Contains(flags, "s"), ok: true}
	case "global":
		return staticEvalResult{value: strings.Contains(flags, "g"), ok: true}
	case "hasIndices":
		return staticEvalResult{value: strings.Contains(flags, "d"), ok: true}
	case "ignoreCase":
		return staticEvalResult{value: strings.Contains(flags, "i"), ok: true}
	case "multiline":
		return staticEvalResult{value: strings.Contains(flags, "m"), ok: true}
	case "sticky":
		return staticEvalResult{value: strings.Contains(flags, "y"), ok: true}
	case "unicode":
		return staticEvalResult{value: strings.Contains(flags, "u"), ok: true}
	case "lastIndex":
		return staticEvalResult{value: staticNumberValue(0), ok: true}
	default:
		return staticEvalResult{}
	}
}

// staticStringMemberValue indexes a string by a canonical array-index key, the
// way JavaScript's string indexing does: by UTF-16 code unit, not by byte or
// rune. A non-canonical numeric-shaped key (`"-1"`, `"01"`) yields `undefined`
// like an absent array element; any other key stays unresolved, since this
// evaluator carries no representation for `String.prototype` methods.
func staticStringMemberValue(s string, key string) staticEvalResult {
	index, ok := staticArrayIndex(key)
	if !ok {
		if staticNumberShapedKey(key) {
			return staticEvalResult{value: staticUndefinedValue{}, ok: true}
		}
		return staticEvalResult{}
	}
	units := ecmascript.StringCodeUnits(s)
	if index >= len(units) {
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	}
	return staticEvalResult{value: ecmascript.StringFromCodeUnits(units[index : index+1]), ok: true}
}

func staticObjectOwnProperty(object *staticObjectValue, key string) (any, bool) {
	for i := len(object.extraProperties) - 1; i >= 0; i-- {
		if ecmascript.CompareStrings(object.extraProperties[i].name, key) == 0 {
			return object.extraProperties[i].value, true
		}
	}
	if object.propertyCount > 0 && ecmascript.CompareStrings(object.property.name, key) == 0 {
		return object.property.value, true
	}
	return nil, false
}

// staticNumberShapedKey reports whether key opens the way a number's string
// form can. Such a key is neither an own property of an array literal nor one
// it inherits, so reading it yields `undefined`.
func staticNumberShapedKey(key string) bool {
	if key == "" {
		return false
	}
	switch key[0] {
	case '-', '+', '.':
		return true
	default:
		return key[0] >= '0' && key[0] <= '9'
	}
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
	if !staticEvaluator.resolveIdentifiers {
		return staticEvalResult{}
	}
	argument, ok := staticEvaluator.objectPassThroughArgument(node)
	if !ok {
		return staticEvalResult{}
	}
	return staticEvaluator.evalValue(argument)
}

func (staticEvaluator *StaticStringEvaluator) objectPassThroughArgument(node *ast.Node) (*ast.Node, bool) {
	call := node.AsCallExpression()
	if call == nil {
		return nil, false
	}

	callee := ast.SkipOuterExpressions(call.Expression, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return nil, false
	}
	methodName, ok := staticEvaluator.evalAccessExpressionKey(callee)
	if !ok || !objectPassThroughMethods[methodName] {
		return nil, false
	}

	object := SkipAssertionsAndParens(AccessExpressionObject(callee))
	if !staticEvaluator.isBuiltinIdentifier(object, "Object") {
		return nil, false
	}

	args := node.Arguments()
	if len(args) == 0 || ast.IsSpreadElement(args[0]) {
		return nil, false
	}
	return args[0], true
}

func (staticEvaluator *StaticStringEvaluator) evalArrayJoinCall(node *ast.Node) staticEvalResult {
	value, _, ok := staticEvaluator.evalArrayJoin(node)
	return staticEvalResult{value: value, ok: ok}
}

func (staticEvaluator *StaticStringEvaluator) evalArrayJoin(node *ast.Node) (value string, matched bool, ok bool) {
	call := node.AsCallExpression()
	if call == nil {
		return "", false, false
	}

	callee := SkipAssertionsAndParens(call.Expression)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return "", false, false
	}
	methodName, nameOK := staticEvaluator.evalAccessExpressionKey(callee)
	if !nameOK || methodName != "join" {
		return "", false, false
	}

	receiverNode := AccessExpressionObject(callee)
	if literal := SkipAssertionsAndParens(receiverNode); literal != nil && literal.Kind == ast.KindArrayLiteralExpression {
		arrayLiteral := literal.AsArrayLiteralExpression()
		if arrayLiteral != nil && arrayLiteral.Elements != nil &&
			len(arrayLiteral.Elements.Nodes) <= len((staticArrayValue{}).inline) {
			array, arrayOK := staticEvaluator.evalArrayLiteralValue(literal)
			if !arrayOK {
				return "", true, false
			}
			separator, separatorOK := staticEvaluator.evalArrayJoinSeparator(node)
			if !separatorOK {
				return "", true, false
			}
			value, ok := staticArrayJoin(&array, separator)
			return value, true, ok
		}
	}

	receiver := staticEvaluator.evalValue(receiverNode)
	if !receiver.ok {
		return "", true, false
	}
	array, arrayOK := receiver.value.(*staticArrayValue)
	if !arrayOK {
		return "", true, false
	}

	separator, separatorOK := staticEvaluator.evalArrayJoinSeparator(node)
	if !separatorOK {
		return "", true, false
	}
	value, ok = staticArrayJoin(array, separator)
	return value, true, ok
}

func (staticEvaluator *StaticStringEvaluator) evalArrayJoinSeparator(node *ast.Node) (string, bool) {
	separator := ","
	for i, argumentNode := range node.Arguments() {
		if ast.IsSpreadElement(argumentNode) {
			return "", false
		}
		argument := staticEvaluator.evalValue(argumentNode)
		if !argument.ok {
			return "", false
		}
		if i == 0 && !staticValueUndefined(argument.value) {
			var ok bool
			separator, ok = staticValueToString(argument.value)
			if !ok {
				return "", false
			}
		}
	}
	return separator, true
}

func (staticEvaluator *StaticStringEvaluator) evalStringCall(node *ast.Node) staticEvalResult {
	call := node.AsCallExpression()
	if call == nil {
		return staticEvalResult{}
	}

	callee := ast.SkipOuterExpressions(call.Expression, ast.OEKParentheses|ast.OEKAssertions)
	if !staticEvaluator.isBuiltinStringValue(callee, map[*ast.Symbol]bool{}) {
		return staticEvalResult{}
	}

	arguments, ok := staticEvaluator.evalCallArguments(node)
	if !ok {
		return staticEvalResult{}
	}
	if len(arguments) == 0 {
		return staticEvalResult{value: "", ok: true}
	}
	if symbol, ok := arguments[0].(staticSymbolValue); ok {
		key := symbol.key
		if symbol.kind == staticSymbolWellKnown {
			key = "Symbol." + key
		}
		return staticEvalResult{value: "Symbol(" + key + ")", ok: true}
	}
	value, ok := staticValueToString(arguments[0])
	if !ok {
		return staticEvalResult{}
	}
	return staticEvalResult{value: value, ok: true}
}

// evalBuiltinStaticCall folds the narrow built-in call subset that
// eslint-utils' getStaticValue permits and that string-consuming native rules
// need. It intentionally recognizes only the real String / Array constructors
// (or stable aliases) and only fully static arguments, never arbitrary methods
// named like a built-in.
func (staticEvaluator *StaticStringEvaluator) evalBuiltinStaticCall(node *ast.Node) staticEvalResult {
	call := node.AsCallExpression()
	if call == nil {
		return staticEvalResult{}
	}

	callee := ast.SkipOuterExpressions(call.Expression, ast.OEKParentheses|ast.OEKAssertions)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return staticEvalResult{}
	}
	method, ok := staticEvaluator.evalAccessExpressionKey(callee)
	if !ok {
		return staticEvalResult{}
	}
	arguments, ok := staticEvaluator.evalCallArguments(node)
	if !ok {
		return staticEvalResult{}
	}

	receiverNode := AccessExpressionObject(callee)
	if staticEvaluator.isBuiltinStringValue(receiverNode, map[*ast.Symbol]bool{}) {
		switch method {
		case "fromCharCode":
			return staticStringFromCharCode(arguments)
		}
	}
	if staticEvaluator.isBuiltinArrayValue(receiverNode, map[*ast.Symbol]bool{}) && method == "of" {
		return staticEvalResult{value: staticArrayFromValues(arguments), ok: true}
	}

	receiver := staticEvaluator.evalValue(receiverNode)
	text, ok := staticValueAsString(receiver.value)
	if !receiver.ok || !ok {
		return staticEvalResult{}
	}

	switch method {
	case "toString":
		return staticEvalResult{value: text, ok: true}
	case "indexOf":
		return staticStringIndexOf(text, arguments)
	case "toUpperCase":
		return staticEvalResult{value: ecmascript.StringToUpperCase(text), ok: true}
	case "toLowerCase":
		return staticEvalResult{value: ecmascript.StringToLowerCase(text), ok: true}
	case "trim":
		return staticEvalResult{value: ecmascript.StringTrim(text), ok: true}
	case "slice":
		return staticStringSlice(text, arguments)
	case "substring":
		return staticStringSubstring(text, arguments)
	case "substr":
		return staticStringSubstr(text, arguments)
	case "charAt":
		return staticStringCharAt(text, arguments)
	case "concat":
		return staticStringConcat(text, arguments)
	}

	return staticEvalResult{}
}

func staticStringIndexOf(text string, arguments []any) staticEvalResult {
	needle := "undefined"
	if len(arguments) > 0 {
		var ok bool
		needle, ok = staticValueToString(arguments[0])
		if !ok {
			return staticEvalResult{}
		}
	}

	start, ok := staticArgumentInteger(arguments, 1, 0)
	if !ok {
		return staticEvalResult{}
	}

	textUnits := ecmascript.StringCodeUnits(text)
	needleUnits := ecmascript.StringCodeUnits(needle)
	start = min(max(start, 0), len(textUnits))
	index := indexOfCodeUnits(textUnits, needleUnits, start)
	return staticEvalResult{value: staticNumberValue(index), ok: true}
}

func indexOfCodeUnits(text, needle []uint16, start int) int {
	if len(needle) == 0 {
		return start
	}
	if start >= len(text) || len(needle) > len(text)-start {
		return -1
	}
	if len(needle) == 1 {
		index := slices.Index(text[start:], needle[0])
		if index < 0 {
			return -1
		}
		return start + index
	}

	prefixLength := make([]int, len(needle))
	for i, matched := 1, 0; i < len(needle); i++ {
		for matched > 0 && needle[i] != needle[matched] {
			matched = prefixLength[matched-1]
		}
		if needle[i] == needle[matched] {
			matched++
		}
		prefixLength[i] = matched
	}

	for i, matched := start, 0; i < len(text); i++ {
		for matched > 0 && text[i] != needle[matched] {
			matched = prefixLength[matched-1]
		}
		if text[i] == needle[matched] {
			matched++
			if matched == len(needle) {
				return i - len(needle) + 1
			}
		}
	}
	return -1
}

func (staticEvaluator *StaticStringEvaluator) evalCallArguments(node *ast.Node) ([]any, bool) {
	return staticEvaluator.evalElementValues(node.Arguments())
}

func (staticEvaluator *StaticStringEvaluator) evalElementValues(nodes []*ast.Node) ([]any, bool) {
	values := make([]any, 0, len(nodes))
	for _, argumentNode := range nodes {
		if ast.IsSpreadElement(argumentNode) {
			spread := staticEvaluator.evalValue(argumentNode.AsSpreadElement().Expression)
			if !spread.ok || staticEvaluator.hasModifiedIterator(spread.value) {
				return nil, false
			}
			if array, ok := spread.value.(*staticArrayValue); ok {
				if array.length > maxStaticArrayLength-len(values) {
					return nil, false
				}
				for index := range array.length {
					values = append(values, array.element(index))
				}
			} else if text, ok := staticValueAsString(spread.value); ok {
				// String iteration uses code points, preserving lone surrogates.
				text = ecmascript.CombineSurrogatePairs(text)
				for len(text) > 0 {
					if len(values) == maxStaticArrayLength {
						return nil, false
					}
					_, size := ecmascript.DecodeStringRune(text)
					values = append(values, text[:size])
					text = text[size:]
				}
			} else {
				return nil, false
			}
			continue
		}
		if len(values) == maxStaticArrayLength {
			return nil, false
		}
		if argumentNode.Kind == ast.KindOmittedExpression {
			values = append(values, nil)
			continue
		}
		argument := staticEvaluator.evalValue(argumentNode)
		if !argument.ok {
			return nil, false
		}
		values = append(values, argument.value)
	}
	return values, true
}

func staticArrayFromValues(values []any) *staticArrayValue {
	array := &staticArrayValue{length: len(values)}
	if array.length > len(array.inline) {
		array.overflow = make([]any, array.length-len(array.inline))
	}
	for index, value := range values {
		array.set(index, value)
	}
	return array
}

func staticStringFromCharCode(arguments []any) staticEvalResult {
	units := make([]uint16, 0, len(arguments))
	for _, argument := range arguments {
		number, ok := staticValueToNumber(argument)
		if !ok {
			return staticEvalResult{}
		}
		units = append(units, uint16(toUint32(number)))
	}
	return staticEvalResult{value: ecmascript.StringFromCodeUnits(units), ok: true}
}

func staticStringSlice(text string, arguments []any) staticEvalResult {
	start, ok := staticArgumentInteger(arguments, 0, 0)
	if !ok {
		return staticEvalResult{}
	}
	end, ok := staticArgumentInteger(arguments, 1, math.MaxInt)
	if !ok {
		return staticEvalResult{}
	}
	units := ecmascript.StringCodeUnits(text)
	length := len(units)
	from := normalizeSliceIndex(start, length)
	to := normalizeSliceIndex(end, length)
	if to < from {
		to = from
	}
	return staticEvalResult{value: ecmascript.StringFromCodeUnits(units[from:to]), ok: true}
}

func staticStringSubstring(text string, arguments []any) staticEvalResult {
	start, ok := staticArgumentInteger(arguments, 0, 0)
	if !ok {
		return staticEvalResult{}
	}
	end, ok := staticArgumentInteger(arguments, 1, math.MaxInt)
	if !ok {
		return staticEvalResult{}
	}
	units := ecmascript.StringCodeUnits(text)
	length := len(units)
	from := clampSubstringIndex(start, length)
	to := clampSubstringIndex(end, length)
	if from > to {
		from, to = to, from
	}
	return staticEvalResult{value: ecmascript.StringFromCodeUnits(units[from:to]), ok: true}
}

// staticStringSubstr implements the Annex B String#substr, whose second
// argument is a length rather than an end index.
func staticStringSubstr(text string, arguments []any) staticEvalResult {
	start, ok := staticArgumentInteger(arguments, 0, 0)
	if !ok {
		return staticEvalResult{}
	}
	count, ok := staticArgumentInteger(arguments, 1, math.MaxInt)
	if !ok {
		return staticEvalResult{}
	}
	units := ecmascript.StringCodeUnits(text)
	length := len(units)
	from := normalizeSliceIndex(start, length)
	if count <= 0 {
		return staticEvalResult{value: "", ok: true}
	}
	to := length
	if count < length-from {
		to = from + count
	}
	return staticEvalResult{value: ecmascript.StringFromCodeUnits(units[from:to]), ok: true}
}

func staticStringCharAt(text string, arguments []any) staticEvalResult {
	index, ok := staticArgumentInteger(arguments, 0, 0)
	if !ok {
		return staticEvalResult{}
	}
	units := ecmascript.StringCodeUnits(text)
	if index < 0 || index >= len(units) {
		return staticEvalResult{value: "", ok: true}
	}
	return staticEvalResult{value: ecmascript.StringFromCodeUnits(units[index : index+1]), ok: true}
}

func staticStringConcat(text string, arguments []any) staticEvalResult {
	var builder strings.Builder
	builder.WriteString(text)
	for _, argument := range arguments {
		part, ok := staticValueToString(argument)
		if !ok {
			return staticEvalResult{}
		}
		if len(part) > maxStaticStringLength-builder.Len() {
			return staticEvalResult{}
		}
		builder.WriteString(part)
	}
	return staticEvalResult{value: ecmascript.CombineSurrogatePairs(builder.String()), ok: true}
}

func staticArgumentInteger(arguments []any, index int, defaultValue int) (int, bool) {
	if index >= len(arguments) || staticValueUndefined(arguments[index]) {
		return defaultValue, true
	}
	number, ok := staticValueToNumber(arguments[index])
	if !ok {
		return 0, false
	}
	if math.IsNaN(number) || number == 0 {
		return 0, true
	}
	if math.IsInf(number, 1) {
		return math.MaxInt, true
	}
	if math.IsInf(number, -1) {
		return math.MinInt, true
	}
	if number >= float64(math.MaxInt) {
		return math.MaxInt, true
	}
	if number <= float64(math.MinInt) {
		return math.MinInt, true
	}
	return int(math.Trunc(number)), true
}

func normalizeSliceIndex(index, length int) int {
	if index < 0 {
		return max(length+index, 0)
	}
	return min(index, length)
}

func clampSubstringIndex(index, length int) int {
	return min(max(index, 0), length)
}

func toUint32(number float64) uint32 {
	return ecmascript.NumberToUint32(number)
}

func toInt32(number float64) int32 {
	return int32(toUint32(number))
}

func staticValueToNumber(value any) (float64, bool) {
	switch value := value.(type) {
	case *big.Int:
		return 0, false
	case staticNumberValue:
		return float64(value), true
	case bool:
		if value {
			return 1, true
		}
		return 0, true
	case staticNullValue:
		return 0, true
	case staticUndefinedValue:
		return math.NaN(), true
	case string:
		return parseStaticNumber(value)
	case *staticStringNode:
		text, _ := staticValueAsString(value)
		return parseStaticNumber(text)
	default:
		text, ok := staticValueToString(value)
		if !ok {
			return 0, false
		}
		return parseStaticNumber(text)
	}
}

func parseStaticNumber(text string) (float64, bool) {
	value, ok := ecmascript.StringToNumber(text)
	if !ok {
		// StringToNumber returns NaN for an invalid numeric string. NaN is still
		// a statically known value and downstream coercions must be allowed to
		// observe it (for example, String.fromCharCode(NaN) produces a NUL).
		return math.NaN(), true
	}
	return value, true
}

func (staticEvaluator *StaticStringEvaluator) evalStringRawTag(node *ast.Node) staticEvalResult {
	tagged := node.AsTaggedTemplateExpression()
	if tagged == nil || tagged.Template == nil || !staticEvaluator.isStringRawTag(tagged.Tag) {
		return staticEvalResult{}
	}

	switch tagged.Template.Kind {
	case ast.KindNoSubstitutionTemplateLiteral:
		return staticEvalResult{value: templateLiteralText(tagged.Template, true), ok: true}
	case ast.KindTemplateExpression:
		return staticEvaluator.evalTemplateExpression(tagged.Template, true)
	default:
		return staticEvalResult{}
	}
}

func templateLiteralText(node *ast.Node, raw bool) string {
	if !raw {
		return node.Text()
	}
	text := node.TemplateLiteralLikeData().RawText
	if text == "" {
		// Like tsgo's tagged-template transform, recover raw text from source
		// for no-substitution literals, which do not store a separate RawText.
		text = scanner.GetSourceTextOfNodeFromSourceFile(ast.GetSourceFileOfNode(node), node, false)
		endLength := 2 // Heads and middles end with ${.
		if node.Kind == ast.KindNoSubstitutionTemplateLiteral || node.Kind == ast.KindTemplateTail {
			endLength = 1
		}
		text = text[1 : len(text)-endLength]
	}
	// Both cooked and raw template values normalize CRLF and CR to LF.
	return strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
}

func (staticEvaluator *StaticStringEvaluator) isStringRawTag(tag *ast.Node) bool {
	tag = ast.SkipOuterExpressions(tag, ast.OEKParentheses|ast.OEKAssertions)
	if tag == nil || !ast.IsAccessExpression(tag) {
		return false
	}
	propertyName, ok := staticEvaluator.evalAccessExpressionKey(tag)
	if !ok || propertyName != "raw" {
		return false
	}
	return staticEvaluator.isBuiltinStringValue(AccessExpressionObject(tag), map[*ast.Symbol]bool{})
}

func (staticEvaluator *StaticStringEvaluator) isBuiltinValue(
	node *ast.Node,
	name string,
	resolvingAliases map[*ast.Symbol]bool,
) bool {
	if !staticEvaluator.resolveIdentifiers {
		return false
	}
	node = SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if staticEvaluator.isBuiltinIdentifier(node, name) {
		return true
	}
	initializer, symbol, ok := staticEvaluator.resolveIdentifierInitializer(node)
	if !ok || resolvingAliases[symbol] {
		return false
	}
	resolvingAliases[symbol] = true
	defer delete(resolvingAliases, symbol)
	return staticEvaluator.isBuiltinValue(initializer, name, resolvingAliases)
}

func (staticEvaluator *StaticStringEvaluator) isBuiltinStringValue(node *ast.Node, resolvingAliases map[*ast.Symbol]bool) bool {
	return staticEvaluator.isBuiltinValue(node, "String", resolvingAliases)
}

func (staticEvaluator *StaticStringEvaluator) isBuiltinArrayValue(node *ast.Node, resolvingAliases map[*ast.Symbol]bool) bool {
	return staticEvaluator.isBuiltinValue(node, "Array", resolvingAliases)
}

func (staticEvaluator *StaticStringEvaluator) isBuiltinIdentifier(node *ast.Node, name string) bool {
	if !isIdentifierWithText(node, name) || staticEvaluator.GlobalAccess != nil && !staticEvaluator.GlobalAccess(name).IsDeclared() {
		return false
	}
	if refs, ok := staticEvaluator.referenceResolver.(interface {
		IsGlobalNameReference(location *ast.Node, name string, meaning ast.SymbolFlags) bool
	}); ok {
		// eslint-utils' static lookup considers type-only bindings too. Reuse
		// the existing reference index instead of scanning every enclosing scope.
		return refs.IsGlobalNameReference(node, name, ast.SymbolFlagsAll)
	}
	return !IsShadowed(node, name)
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
	if !result.ok {
		return evaluator.Result{}
	}
	if value, ok := staticValueAsString(result.value); ok {
		return evaluator.Result{Value: value}
	}
	if !staticValueIsTsgoSafe(result.value) {
		return evaluator.Result{}
	}

	return evaluator.Result{Value: result.value}
}

func (staticEvaluator *StaticStringEvaluator) hasWrites(symbol *ast.Symbol) bool {
	if staticEvaluator.sourceFile == nil ||
		(staticEvaluator.typeChecker == nil && staticEvaluator.referenceResolver == nil) {
		return true
	}
	return staticEvaluator.referenceFlagsFor(symbol)&staticReferenceWrite != 0
}

func (staticEvaluator *StaticStringEvaluator) hasPropertyMutation(symbol *ast.Symbol) bool {
	if staticEvaluator.sourceFile == nil ||
		(staticEvaluator.typeChecker == nil && staticEvaluator.referenceResolver == nil) {
		return true
	}
	flags := staticEvaluator.referenceFlagsFor(symbol)
	if staticEvaluator.capturePoint != nil && staticEvaluator.escapeFacts != nil {
		if staticEvaluator.escapeFacts.explicit[symbol] {
			return flags&staticReferencePropertyMutation != 0
		}
		for _, sites := range [...]map[*ast.Node]staticReferenceFlags{staticEvaluator.escapeFacts.sites[symbol], staticEvaluator.escapeFacts.lengthWrites[symbol]} {
			for site := range sites {
				if staticEvaluator.transferAvailableAt(staticReferenceTransfer{site: site}, staticEvaluator.capturePoint) {
					return true
				}
			}
		}
		return false
	}
	return flags&staticReferencePropertyMutation != 0
}

// PropertyReadMutations conservatively detects changes that can invalidate
// a data-property assumption, including unknown calls through aliases. Ordinary
// writes to a known property name are allowed, except array truncation. The caller may exclude specific
// calls whose writes it has already accounted for. With a proven fresh origin
// and an access, later calls in the same invocation and harmless built-ins do
// not invalidate this read. A fresh literal can be its own reference when no
// binding exists. Calls are recorded outermost first, preserving nested copy
// order. Descriptor mutations can replace an iterator; value mutations only
// invalidate the selected read. copiedType preserves known source values erased
// by a native write intersection. Call-escape tracking is required.
func (staticEvaluator *StaticStringEvaluator) PropertyReadMutations(node, origin, access *ast.Node, accountedCalls *collections.OrderedMap[*ast.Node, struct{}]) (descriptorMutation, valueMutation bool, copiedType *checker.Type) {
	if !staticEvaluator.trackCallEscapes || staticEvaluator.sourceFile == nil {
		return true, false, nil
	}
	node = SkipAssertionsAndParens(node)
	symbol := staticEvaluator.referenceSymbol(node)
	if symbol == nil {
		// An inline fresh target has no binding to follow, but its copy
		// sources and built-in prototypes still affect the observed value.
		if node != origin || staticEvaluator.typeChecker == nil ||
			!ast.IsObjectLiteralExpression(node) && !ast.IsArrayLiteralExpression(node) && !ast.IsFunctionExpressionOrArrowFunction(node) {
			return true, false, nil
		}
		staticEvaluator.ensureReferenceFlags()
	} else if staticEvaluator.referenceFlagsFor(symbol)&(staticReferenceWrite|staticReferencePropertyShapeMutation) != 0 {
		return true, false, nil
	}
	if origin != nil && staticEvaluator.typeChecker != nil {
		constructors := []string{"Object"}
		if ast.IsArrayLiteralExpression(origin) {
			constructors = append(constructors, "Array")
		} else if ast.IsFunctionLike(origin) {
			constructors = append(constructors, "Function")
		}
		for _, name := range constructors {
			constructor := staticEvaluator.typeChecker.ResolveName(name, &staticEvaluator.sourceFile.Node, ast.SymbolFlagsValue, false)
			if constructor != nil && staticEvaluator.hasConstructorMutation(constructor) {
				// A fresh target can inherit a setter installed on a built-in
				// prototype. Object.assign does not create a data field over it.
				return true, false, nil
			}
		}
	}
	for site := range staticEvaluator.escapeFacts.sites[symbol] {
		if accountedCalls.Has(site) || origin != nil && access != nil &&
			(staticEvaluator.propertyMutationFollowsRead(site, origin, access) || staticEvaluator.callPreservesPropertyShape(site, origin, access, symbol)) {
			continue
		}
		return true, false, nil
	}
	for call := range accountedCalls.Keys() {
		for _, source := range call.Arguments()[1:] {
			info := staticEvaluator.copySource(source)
			if info.hasEffects {
				return true, false, nil
			}
			if origin != nil && ast.IsArrayLiteralExpression(origin) && access != nil && info.unknown && staticEvaluator.typeChecker != nil {
				key, known := staticEvaluator.evalAccessExpressionKey(access)
				if (!known || isArrayIndexProperty(key)) && GetWellKnownSymbolPropertyOfType(GetConstrainedTypeAtLocation(staticEvaluator.typeChecker, source), "iterator", staticEvaluator.typeChecker) == nil {
					// An opaque source can replace a target iterator absent from its
					// public type. Its yielded values are then unknown as well, so a
					// destructuring default cannot trust only the target's iterator.
					return true, false, nil
				}
			}
		}
	}
	if origin != nil && access != nil && accountedCalls.Size() != 0 {
		key, known := staticEvaluator.evalAccessExpressionKey(access)
		invalidated, unknown, copied := staticEvaluator.copyReadMutation(accountedCalls, access, key, known, ast.IsArrayLiteralExpression(origin))
		copiedType = copied
		if unknown {
			return false, true, nil
		}
		if invalidated || copiedType != nil {
			call, _, _ := accountedCalls.EntryAt(0)
			write := staticEvaluator.propertyWriteAtRead(access)
			if write != nil && staticEvaluator.propertyMutationFollowsRead(write, call, call) {
				copiedType = nil
			} else if invalidated {
				return false, true, nil
			}
		}
	}
	if sites := staticEvaluator.escapeFacts.lengthWrites[symbol]; len(sites) != 0 {
		if origin == nil || access == nil {
			return false, true, nil
		}
		if !ast.IsArrayLiteralExpression(origin) {
			return false, false, copiedType
		}
		// Length writes cannot change named properties. Keep their sites
		// separate so these reads do not scan every truncation in the file.
		key, known := staticEvaluator.evalAccessExpressionKey(access)
		if known && !isArrayIndexProperty(key) {
			return false, false, copiedType
		}
		write := staticEvaluator.propertyWriteAtRead(access)
		if write != nil && staticEvaluator.lengthWritesInReadBlock(symbol, origin, access, write) {
			return false, false, copiedType
		}
		for site := range sites {
			if staticEvaluator.propertyMutationFollowsRead(site, origin, access) ||
				write != nil && staticEvaluator.propertyMutationFollowsRead(write, origin, site) {
				continue
			}
			return false, true, nil
		}
	}
	return false, false, copiedType
}

func (staticEvaluator *StaticStringEvaluator) copyReadMutation(calls *collections.OrderedMap[*ast.Node, struct{}], access *ast.Node, key string, keyKnown, arrayTarget bool) (invalidated, unknown bool, copiedType *checker.Type) {
	if !keyKnown && staticEvaluator.typeChecker != nil && ast.IsElementAccessExpression(access) {
		keyType := GetConstrainedTypeAtLocation(staticEvaluator.typeChecker, access.AsElementAccessExpression().ArgumentExpression)
		var copiedTypes []*checker.Type
		for _, part := range UnionTypeParts(keyType) {
			if !checker.IsTypeUsableAsPropertyName(part) {
				return false, true, nil
			}
			invalidated, unknown, copied := staticEvaluator.copyReadMutation(calls, access, checker.GetPropertyNameFromType(part), true, arrayTarget)
			if invalidated || unknown {
				return invalidated, unknown, nil
			}
			if copied != nil {
				copiedTypes = append(copiedTypes, copied)
			}
		}
		if len(copiedTypes) != 0 {
			copiedTypes = append(copiedTypes, GetConstrainedTypeAtLocation(staticEvaluator.typeChecker, access))
			copiedType = staticEvaluator.typeChecker.GetUnionType(copiedTypes)
		}
		return false, false, copiedType
	}
	arrayIndex := arrayTarget && (!keyKnown || isArrayIndexProperty(key))
	var readType *checker.Type
	addsValue := func(value *checker.Type) bool {
		tc := staticEvaluator.typeChecker
		// Correct an intersection that erased a source's possible read values.
		// When the property's native contract already includes that value,
		// ordinary guards and writes can still narrow the copied data field.
		if readType == nil {
			readType = checker.Checker_getIndexedAccessTypeOrUndefined(tc, tc.GetTypeAtLocation(access.Expression()), checker.Checker_getStringLiteralType(tc, key), checker.AccessFlagsExpressionPosition, nil, nil)
		}
		return readType == nil || !checker.Checker_isTypeAssignableTo(tc, value, readType)
	}
	var index float64
	if arrayIndex && keyKnown {
		index, _ = ecmascript.StringToNumber(key)
	}
	for i := calls.Size() - 1; i >= 0; i-- {
		call, _, _ := calls.EntryAt(i)
		for _, source := range call.Arguments()[1:] {
			info := staticEvaluator.copySource(source)
			if keyKnown && info.literal != nil && ast.IsArrayLiteralExpression(info.literal) {
				index, indexed := staticArrayIndex(key)
				elements := info.literal.AsArrayLiteralExpression().Elements.Nodes
				if !indexed || !info.hasArraySpread && (index >= len(elements) || elements[index].Kind == ast.KindOmittedExpression) {
					// Array holes and intrinsic members are not enumerable entries.
					continue
				}
			}
			if arrayIndex && info.unknown {
				// An opaque source can also copy __proto__ or invoke a getter.
				// A later indexed write alone cannot prove a data property exists.
				return false, true, nil
			}
			if keyKnown && staticEvaluator.typeChecker != nil {
				tc := staticEvaluator.typeChecker
				sourceType := staticEvaluator.copySourceReadType(source, info)
				restores := info.literal != nil && !info.unknown
				addsNewValue := false
				var values []*checker.Type
				for _, part := range UnionTypeParts(sourceType) {
					property := checker.Checker_getPropertyOfType(tc, part, key)
					if property != nil {
						// A type's inherited and intrinsic members are not copied.
						// Tuple indices are own entries; ordinary literal properties
						// retain their object-literal declarations through spreads.
						tupleEntry := info.literal != nil && ast.IsArrayLiteralExpression(info.literal) && isArrayIndexProperty(key) &&
							!info.hasArraySpread
						restores = restores && (tupleEntry || slices.ContainsFunc(property.Declarations, func(declaration *ast.Node) bool {
							return declaration.Parent != nil && ast.IsObjectLiteralExpression(declaration.Parent)
						}))
						valueType := constrainedCopyValueType(tc, checker.Checker_getTypeOfSymbol(tc, property))
						if property.Flags&ast.SymbolFlagsAccessor == ast.SymbolFlagsSetAccessor {
							// A setter without a getter reads undefined, regardless of its write type.
							valueType = tc.GetUndefinedType()
						} else if property.CheckFlags&ast.CheckFlagsSynthetic != 0 || slices.ContainsFunc(property.Declarations, ast.IsSetAccessorDeclaration) {
							// Union and mapped properties can hide their accessor flags.
							if slices.ContainsFunc(tc.GetRootSymbols(property), func(root *ast.Symbol) bool {
								return root.Flags&ast.SymbolFlagsAccessor == ast.SymbolFlagsSetAccessor
							}) {
								valueType = tc.GetUnionType([]*checker.Type{valueType, tc.GetUndefinedType()})
							}
						}
						values = append(values, valueType)
						addsNewValue = addsNewValue || addsValue(valueType)
						if property.Flags&(ast.SymbolFlagsOptional|ast.SymbolFlagsAccessor) != 0 {
							restores = false
						}
					} else {
						restores = false
						if info.hasProperties || info.unknown {
							// An index signature can overwrite a key without exposing a
							// property symbol. Empty literal arrays have no own elements.
							if indexInfo := checker.Checker_getApplicableIndexInfo(tc, part, checker.Checker_getStringLiteralType(tc, key)); indexInfo != nil {
								valueType := constrainedCopyValueType(tc, indexInfo.ValueType())
								values = append(values, valueType)
								addsNewValue = addsNewValue || addsValue(valueType)
							}
						}
					}
				}
				if restores {
					invalidated = false
					copiedType = nil
					if addsNewValue {
						copiedType = tc.GetUnionType(values)
					}
				} else if addsNewValue {
					if copiedType != nil {
						values = append(values, copiedType)
					}
					if readType != nil {
						values = append(values, readType)
					}
					copiedType = tc.GetUnionType(values)
				}
			}
			if arrayIndex && info.hasLength {
				value := staticEvaluator.evalValue(info.length)
				number, numeric := staticValueToNumber(value.value)
				// Numeric keys are copied before length within one source, even
				// when the object literal spells length first. Sources themselves
				// are copied left to right, so a later source can restore an index.
				if !keyKnown || !value.ok || !numeric || number != float64(ecmascript.NumberToUint32(number)) || number <= index {
					invalidated = true
				}
			}
		}
	}
	return invalidated, false, copiedType
}

func constrainedCopyValueType(tc *checker.Checker, value *checker.Type) *checker.Type {
	if constraint := checker.Checker_getBaseConstraintOfType(tc, value); constraint != nil {
		value = constraint
	}
	if IsTypeFlagSetWithUnion(value, checker.TypeFlagsInstantiableNonPrimitive) {
		return tc.GetUnknownType()
	}
	return value
}

func (staticEvaluator *StaticStringEvaluator) copySourceReadType(source *ast.Node, info staticCopySource) *checker.Type {
	if info.sourceType != nil {
		return info.sourceType
	}
	if info.literal != nil {
		source = info.literal
	}
	if ast.IsObjectLiteralExpression(source) {
		// Rechecking a wide literal for every union key is quadratic.
		return checker.Checker_checkExpressionCached(staticEvaluator.typeChecker, source)
	}
	return GetConstrainedTypeAtLocation(staticEvaluator.typeChecker, source)
}

// Copying invokes getters before later properties are enumerated. Reuse the
// effect checker, but also reject implicit object coercions that it deliberately
// does not classify. Only local primitive writes are exempt from explicit effects.
func (staticEvaluator *StaticStringEvaluator) copyGetterIsPure(getter *ast.Node) bool {
	tc := staticEvaluator.typeChecker
	primitive := func(node *ast.Node) bool {
		node = SkipAssertionsAndParens(node)
		return tc != nil && !slices.ContainsFunc(UnionTypeParts(GetConstrainedTypeAtLocation(tc, node)), func(part *checker.Type) bool {
			return part.Flags()&(checker.TypeFlagsPrimitive|checker.TypeFlagsNever) == 0
		})
	}
	var mayCoerceObject func(*ast.Node) bool
	mayCoerceObject = func(node *ast.Node) bool {
		if ast.IsTypeNode(node) || ast.IsFunctionExpressionOrArrowFunction(node) {
			return false
		}
		switch node.Kind {
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			comparesNull := (binary.OperatorToken.Kind == ast.KindEqualsEqualsToken || binary.OperatorToken.Kind == ast.KindExclamationEqualsToken) &&
				(SkipAssertionsAndParens(binary.Left).Kind == ast.KindNullKeyword || SkipAssertionsAndParens(binary.Right).Kind == ast.KindNullKeyword)
			if !ast.IsLogicalOrCoalescingBinaryExpression(node) && binary.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken &&
				binary.OperatorToken.Kind != ast.KindExclamationEqualsEqualsToken && binary.OperatorToken.Kind != ast.KindCommaToken && !comparesNull &&
				(!primitive(binary.Left) || !primitive(binary.Right)) {
				return true
			}
		case ast.KindPrefixUnaryExpression:
			prefix := node.AsPrefixUnaryExpression()
			if prefix.Operator != ast.KindExclamationToken && !primitive(prefix.Operand) {
				return true
			}
		case ast.KindTemplateSpan, ast.KindComputedPropertyName:
			if !primitive(node.Expression()) {
				return true
			}
		}
		return node.ForEachChild(mayCoerceObject)
	}
	safeExpression := func(node *ast.Node) bool {
		return node == nil || !staticEvaluator.HasSideEffect(node, true) && !mayCoerceObject(node)
	}
	localPrimitive := func(node *ast.Node) bool {
		node = SkipAssertionsAndParens(node)
		if tc == nil || !ast.IsIdentifier(node) || !primitive(node) {
			return false
		}
		symbol := tc.GetSymbolAtLocation(node)
		return symbol != nil && symbol.ValueDeclaration != nil && ast.IsVariableDeclaration(symbol.ValueDeclaration) &&
			executionRoot(symbol.ValueDeclaration) == getter && symbol.ValueDeclaration.Initializer() != nil && primitive(symbol.ValueDeclaration.Initializer())
	}
	var safeStatement func(*ast.Node) bool
	safeStatement = func(node *ast.Node) bool {
		if node == nil {
			return true
		}
		switch node.Kind {
		case ast.KindBlock:
			return !slices.ContainsFunc(node.AsBlock().Statements.Nodes, func(statement *ast.Node) bool { return !safeStatement(statement) })
		case ast.KindReturnStatement:
			return safeExpression(node.Expression())
		case ast.KindIfStatement:
			statement := node.AsIfStatement()
			return safeExpression(statement.Expression) && safeStatement(statement.ThenStatement) && safeStatement(statement.ElseStatement)
		case ast.KindVariableStatement:
			list := node.AsVariableStatement().DeclarationList
			if list.Flags&ast.NodeFlagsUsing != 0 {
				// Disposal runs user code when the getter returns.
				return false
			}
			return !slices.ContainsFunc(list.AsVariableDeclarationList().Declarations.Nodes, func(declaration *ast.Node) bool {
				return !ast.IsIdentifier(declaration.Name()) || !safeExpression(declaration.Initializer())
			})
		case ast.KindExpressionStatement:
			expression := SkipAssertionsAndParens(node.Expression())
			if ast.IsPostfixUnaryExpression(expression) {
				return localPrimitive(expression.AsPostfixUnaryExpression().Operand)
			}
			if ast.IsPrefixUnaryExpression(expression) {
				prefix := expression.AsPrefixUnaryExpression()
				if prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken {
					return localPrimitive(prefix.Operand)
				}
			}
			return safeExpression(expression)
		}
		return false
	}
	return getter.Body() != nil && safeStatement(getter.Body())
}

func (staticEvaluator *StaticStringEvaluator) copySource(source *ast.Node) (info staticCopySource) {
	source = SkipAssertionsAndParens(source)
	if source == nil {
		return staticCopySource{unknown: true}
	}
	facts := staticEvaluator.escapeFacts
	if info, cached := facts.copySources[source]; cached {
		return info
	}
	if facts.copySources == nil {
		facts.copySources = make(map[*ast.Node]staticCopySource)
	}
	// Cache immutable source syntax and its original branch contracts, with a conservative cycle sentinel.
	facts.copySources[source] = staticCopySource{unknown: true}
	defer func() { facts.copySources[source] = info }()
	if ast.IsIdentifier(source) {
		initializer, symbol, known := staticEvaluator.resolveIdentifierInitializer(source)
		if known && staticEvaluator.referenceFlagsFor(symbol)&staticReferenceWrite == 0 {
			info = staticEvaluator.copySource(initializer)
			if flags := staticEvaluator.referenceFlagsFor(symbol); flags&staticReferencePropertyMutation != 0 {
				// Losing the source's shape does not remove getter effects. A
				// getter can itself mutate this binding while it is being copied.
				// Descriptor changes can also remove fields or install new getters.
				hasEffects := info.hasEffects || flags&staticReferencePropertyShapeMutation != 0
				if !hasEffects {
					for call := range facts.sites[symbol] {
						// Harmless-call proofs require an established non-Proxy origin.
						if info.literal == nil || !staticEvaluator.callPreservesPropertyShape(call, info.literal, source, symbol) {
							hasEffects = true
							break
						}
					}
				}
				return staticCopySource{unknown: true, hasEffects: hasEffects}
			}
			if info.sourceType != nil {
				// Keep the original read alternatives while respecting narrowing
				// at this particular use of the source binding.
				flowType := GetConstrainedTypeAtLocation(staticEvaluator.typeChecker, source)
				var possible []*checker.Type
				for _, part := range UnionTypeParts(info.sourceType) {
					if checker.Checker_isTypeAssignableTo(staticEvaluator.typeChecker, part, flowType) {
						possible = append(possible, part)
					}
				}
				if len(possible) != 0 {
					info.sourceType = staticEvaluator.typeChecker.GetUnionType(possible)
				}
			}
			return info
		}
	}
	if staticEvaluator.typeChecker != nil {
		var alternatives []*ast.Node
		if ast.IsConditionalExpression(source) {
			conditional := source.AsConditionalExpression()
			alternatives = []*ast.Node{conditional.WhenTrue, conditional.WhenFalse}
		} else if ast.IsLogicalOrCoalescingBinaryExpression(source) {
			expression := source.AsBinaryExpression()
			alternatives = []*ast.Node{expression.Left, expression.Right}
		}
		if alternatives != nil {
			// Native conditional-expression subtype reduction can erase a
			// setter branch that has the same write shape as a data property.
			// Join the original alternatives without that reduction.
			var types []*checker.Type
			for i, alternative := range alternatives {
				branch := staticEvaluator.copySource(alternative)
				t := staticEvaluator.copySourceReadType(alternative, branch)
				if i == 0 && ast.IsBinaryExpression(source) {
					facts := checker.TypeFactsNEUndefinedOrNull
					switch source.AsBinaryExpression().OperatorToken.Kind {
					case ast.KindAmpersandAmpersandToken:
						facts = checker.TypeFactsFalsy
					case ast.KindBarBarToken:
						facts = checker.TypeFactsTruthy
					}
					t = checker.Checker_getTypeWithFacts(staticEvaluator.typeChecker, t, facts)
				}
				if t.Flags()&checker.TypeFlagsNever == 0 {
					info.hasEffects = info.hasEffects || branch.hasEffects
					types = append(types, t)
				}
			}
			info.sourceType = staticEvaluator.typeChecker.GetUnionType(types)
			info.unknown = true
			return info
		}
	}
	if ast.IsArrayLiteralExpression(source) || ast.IsFunctionExpressionOrArrowFunction(source) {
		// Their intrinsic length is non-enumerable.
		if ast.IsArrayLiteralExpression(source) && len(source.AsArrayLiteralExpression().Elements.Nodes) != 0 {
			info.literal = source
			info.hasProperties = true
			info.hasArraySpread = slices.ContainsFunc(source.AsArrayLiteralExpression().Elements.Nodes, ast.IsSpreadElement)
		}
		return info
	}
	if ast.IsObjectLiteralExpression(source) {
		info.literal = source
		for _, property := range source.AsObjectLiteralExpression().Properties.Nodes {
			if ast.IsGetAccessorDeclaration(property) && !staticEvaluator.copyGetterIsPure(property) {
				info.hasEffects = true
			}
			if ast.IsSpreadAssignment(property) {
				spread := staticEvaluator.copySource(property.Expression())
				info.hasEffects = info.hasEffects || spread.hasEffects
				if spread.literal != nil && ast.IsArrayLiteralExpression(spread.literal) && staticEvaluator.typeChecker != nil {
					tc := staticEvaluator.typeChecker
					sourceType := tc.GetTypeAtLocation(spread.literal)
					// A plain array spread can overwrite an explicit property
					// without appearing in that property's native symbol type.
					// Tuples expose their individual indices to the checker.
					if checker.Checker_isArrayType(tc, sourceType) {
						if elementType := checker.Checker_getIndexTypeOfType(tc, sourceType, tc.GetNumberType()); elementType != nil && IsNullableType(constrainedCopyValueType(tc, elementType)) {
							spread.unknown = true
						}
					}
				}
				if spread.unknown {
					info.unknown = true
					continue
				}
				if spread.hasLength {
					info.hasLength, info.length = true, spread.length
				}
				info.hasProperties = info.hasProperties || spread.hasProperties
				continue
			}
			info.hasProperties = true
			key, known := staticEvaluator.evalPropertyKey(property.Name())
			if !known {
				info.unknown = true
				continue
			}
			if key == "length" {
				info.hasLength, info.length = true, nil
				if ast.IsPropertyAssignment(property) {
					info.length = property.Initializer()
				} else if ast.IsShorthandPropertyAssignment(property) {
					info.length = property.Name()
				}
			}
		}
		return info
	}
	if staticEvaluator.typeChecker != nil && !slices.ContainsFunc(UnionTypeParts(GetConstrainedTypeAtLocation(staticEvaluator.typeChecker, source)), func(part *checker.Type) bool {
		return part.Flags()&checker.TypeFlagsPrimitive == 0
	}) {
		return info
	}
	return staticCopySource{unknown: true}
}

func (staticEvaluator *StaticStringEvaluator) lengthWritesInReadBlock(symbol *ast.Symbol, origin, access, write *ast.Node) bool {
	facts := staticEvaluator.escapeFacts
	block, cached := facts.lengthWriteBlocks[symbol]
	if !cached {
		for site := range facts.lengthWrites[symbol] {
			if facts.untrackedLengthWrites[site] || site.Parent == nil || site.Parent.Kind != ast.KindExpressionStatement ||
				site.Parent.Parent == nil || block != nil && block != site.Parent.Parent {
				block = nil
				break
			}
			block = site.Parent.Parent
		}
		if facts.lengthWriteBlocks == nil {
			facts.lengthWriteBlocks = make(map[*ast.Symbol]*ast.Node)
		}
		facts.lengthWriteBlocks[symbol] = block
	}
	// All writes are ordinary statements recorded by the native binder. The
	// immediately preceding indexed assignment supersedes the earlier ones;
	// no direct length assignment can intervene without changing that flow.
	// Exclude class initializers and computed names, whose order can differ.
	root := executionRoot(access)
	return block != nil && block == statementContainer(access) && block == statementContainer(write) &&
		executionRoot(origin.Parent) == root && ast.GetThisContainer(access, true, false) == root && ast.GetThisContainer(write, true, false) == root
}

// Reuse the native flow graph's immediately preceding write. No branch or loop
// traversal is needed to prove that this exact property reference was restored.
func (staticEvaluator *StaticStringEvaluator) propertyWriteAtRead(access *ast.Node) *ast.Node {
	if staticEvaluator.typeChecker == nil || access.FlowNodeData() == nil || access.FlowNodeData().FlowNode == nil {
		return nil
	}
	flow := access.FlowNodeData().FlowNode
	var assignment *ast.Node
	if flow.Flags&ast.FlowFlagsArrayMutation != 0 && ast.IsBinaryExpression(flow.Node) {
		assignment = flow.Node
	} else if flow.Flags&ast.FlowFlagsAssignment != 0 && ast.IsAccessExpression(flow.Node) {
		assignment = ast.GetAssignmentTarget(flow.Node)
	}
	if assignment != nil && ast.IsBinaryExpression(assignment) && assignment.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken &&
		checker.Checker_isMatchingReference(staticEvaluator.typeChecker, access, assignment.AsBinaryExpression().Left) {
		return assignment
	}
	return nil
}

func (staticEvaluator *StaticStringEvaluator) propertyMutationFollowsRead(mutation, origin, access *ast.Node) bool {
	root := executionRoot(access)
	// Invocation and assignment follow argument and right-hand-side evaluation,
	// including an enclosing operation which consumes this read.
	if root == nil || executionRoot(origin.Parent) != root || executionRoot(mutation) != root || mutation.End() <= access.End() {
		return false
	}
	if (ast.IsCallExpression(mutation) || ast.IsNewExpression(mutation) || ast.IsTaggedTemplateExpression(mutation) ||
		ast.IsBinaryExpression(mutation) && ast.IsAccessExpression(SkipAssertionsAndParens(mutation.AsBinaryExpression().Left))) &&
		ast.IsNodeDescendantOf(access, mutation) {
		return true
	}
	// A later iteration can observe a mutation from the previous iteration.
	legacyDecorators := staticEvaluator.typeChecker == nil || checker.Checker_legacyDecorators(staticEvaluator.typeChecker)
	inDecorator := false
	for parent := access.Parent; parent != nil && parent != root; parent = parent.Parent {
		if ast.IsBindingElement(parent) {
			declaration := ast.GetRootDeclaration(parent)
			if ast.IsVariableDeclaration(declaration) && declaration.Initializer() != nil && ast.IsNodeDescendantOf(mutation, declaration.Initializer()) {
				// Both synthetic reads and explicit defaults carry the binding's
				// location, but the initializer executes before either is read.
				return false
			}
		}
		if ast.IsDecorator(parent) {
			inDecorator = true
		}
		if inDecorator && ast.IsClassLike(parent) {
			// Legacy decorator expressions run after static initialization.
			if legacyDecorators && ast.IsNodeDescendantOf(mutation, parent) {
				return false
			}
			inDecorator = false
		}
		if ast.IsIterationStatement(parent, false) {
			return false
		}
		if ast.IsPropertyDeclaration(parent) && !ast.HasStaticModifier(parent) &&
			parent.Initializer() != nil && ast.IsNodeDescendantOf(access, parent.Initializer()) {
			// Instance initializers run during construction, after the class
			// declaration and possibly after later statements in this function.
			return false
		}
	}
	for child := mutation; child.Parent != nil && child.Parent != root; child = child.Parent {
		if (ast.IsComputedPropertyName(child) || !legacyDecorators && ast.IsDecorator(child)) &&
			ast.IsClassElement(child.Parent) && ast.IsClassLike(child.Parent.Parent) &&
			ast.IsNodeDescendantOf(access, child.Parent.Parent) {
			// Computed names and decorator expressions run before static
			// initializers. Static fields and blocks themselves run in order.
			return false
		}
		if ast.IsPropertyDeclaration(child.Parent) && !ast.HasStaticModifier(child.Parent) && child.Parent.Initializer() == child &&
			ast.IsNodeDescendantOf(access, child.Parent.Parent) {
			// An earlier static initializer can construct this class before
			// another static field reads the affected property.
			return false
		}
		if ast.IsCaseClause(child.Parent) && child.Parent.Expression() == child &&
			ast.IsNodeDescendantOf(access, child.Parent.Parent) {
			// A later case selector can execute before an earlier default body.
			return false
		}
	}
	return true
}

func (staticEvaluator *StaticStringEvaluator) callPreservesPropertyShape(call, origin, access *ast.Node, symbol *ast.Symbol) bool {
	if !ast.IsCallExpression(call) {
		return false
	}
	name, known := staticEvaluator.builtinMethodName(call.Expression(), "Object", make(map[*ast.Symbol]bool))
	if !known {
		return !staticEvaluator.escapeFacts.explicit[symbol] && staticEvaluator.arrayPushPreservesProperty(call, origin, access)
	}
	if staticEvaluator.typeChecker == nil {
		return false
	}
	object := staticEvaluator.typeChecker.ResolveName("Object", &staticEvaluator.sourceFile.Node, ast.SymbolFlagsValue, false)
	if object == nil || staticEvaluator.hasConstructorMutation(object) {
		return false
	}
	arguments := call.Arguments()
	for _, argument := range arguments {
		if ast.IsSpreadElement(argument) {
			return false
		}
	}
	if staticEvaluator.isReadOnlyObjectCall(call, name) {
		return true
	}
	switch name {
	case "assign":
		if len(arguments) == 0 {
			return false
		}
		for _, source := range arguments[1:] {
			value := staticEvaluator.evalValue(source)
			if !value.ok {
				return false
			}
			switch object := value.value.(type) {
			case staticNullValue, staticUndefinedValue:
			case *staticObjectValue:
				if object.propertyCount != 0 {
					return false
				}
			default:
				return false
			}
		}
		return true
	case "defineProperty":
		if len(arguments) != 3 {
			return false
		}
		key, known := staticEvaluator.evalAccessExpressionKey(access)
		name := staticEvaluator.evalValue(arguments[1])
		if !known || !name.ok || !IsPrimitiveStaticValue(name.value) {
			return false
		}
		definedKey, known := staticValueToString(name.value)
		if !known || key == definedKey || ast.IsArrayLiteralExpression(origin) && definedKey == "push" {
			return false
		}
		if ast.IsArrayLiteralExpression(origin) && definedKey == "length" && isArrayIndexProperty(key) {
			return false
		}
		// ToPropertyDescriptor can invoke getters, including inherited ones.
		// Only an independently folded plain data descriptor is harmless.
		descriptor := staticEvaluator.evalValue(arguments[2])
		object, ok := descriptor.value.(*staticObjectValue)
		return descriptor.ok && ok && (!object.prototypeSet || object.prototype == nil)
	}
	return false
}

// ArraySetLength deletes indexed elements when it truncates. Its uint32
// boundary is independent of the host's int width.
func isArrayIndexProperty(key string) bool {
	number, known := ecmascript.StringToNumber(key)
	index := ecmascript.NumberToUint32(number)
	return known && index != math.MaxUint32 && ecmascript.NumberToString(float64(index)) == key
}

// The caller establishes a non-Proxy target and an unchanged built-in.
// Key/descriptor inspection does not invoke the target's getters, but converting
// an object-valued property key can execute arbitrary user code.
func (staticEvaluator *StaticStringEvaluator) isReadOnlyObjectCall(call *ast.Node, name string) bool {
	switch name {
	case "keys", "getOwnPropertyNames", "getOwnPropertySymbols", "getOwnPropertyDescriptors":
		return true
	case "hasOwn", "getOwnPropertyDescriptor":
		arguments := call.Arguments()
		if len(arguments) < 2 {
			return false
		}
		key := SkipAssertionsAndParens(arguments[1])
		if value := staticEvaluator.evalValue(key); value.ok {
			return IsPrimitiveStaticValue(value.value)
		}
		if staticEvaluator.typeChecker != nil {
			keyType := GetConstrainedTypeAtLocation(staticEvaluator.typeChecker, key)
			return !slices.ContainsFunc(UnionTypeParts(keyType), func(part *checker.Type) bool {
				return part.Flags()&checker.TypeFlagsPrimitive == 0
			})
		}
	}
	return false
}

func (staticEvaluator *StaticStringEvaluator) hasConstructorMutation(symbol *ast.Symbol) bool {
	facts := staticEvaluator.escapeFacts
	if changed, cached := facts.constructorMutations[symbol]; cached {
		return changed
	}
	changed := staticEvaluator.hasWrites(symbol) || facts.explicit[symbol]
	for site, flags := range facts.sites[symbol] {
		if flags&staticReferenceCallReceiver != 0 {
			continue
		}
		// Aliased read-only methods still refer to the constructor. Reuse the
		// built-in resolver without treating those calls as constructor escapes.
		if ast.IsCallExpression(site) {
			name, known := staticEvaluator.builtinMethodName(site.Expression(), "Object", make(map[*ast.Symbol]bool))
			if known && staticEvaluator.isReadOnlyObjectCall(site, name) {
				continue
			}
		}
		changed = true
		break
	}
	if facts.constructorMutations == nil {
		facts.constructorMutations = make(map[*ast.Symbol]bool)
	}
	facts.constructorMutations[symbol] = changed
	return changed
}

func (staticEvaluator *StaticStringEvaluator) arrayPushPreservesProperty(call, origin, access *ast.Node) bool {
	if !ast.IsArrayLiteralExpression(origin) || staticEvaluator.arrayPrototypeMutated || staticEvaluator.typeChecker == nil {
		return false
	}
	callee := SkipAssertionsAndParens(call.Expression())
	method, known := staticEvaluator.evalAccessExpressionKey(callee)
	key, keyKnown := staticEvaluator.evalAccessExpressionKey(access)
	if !known || method != "push" || !keyKnown || key == "length" {
		return false
	}
	if _, index := staticArrayIndex(key); index {
		return false
	}
	for _, argument := range call.Arguments() {
		if ast.IsSpreadElement(argument) {
			return false
		}
	}
	receiver := staticEvaluator.referenceSymbol(SkipAssertionsAndParens(AccessExpressionObject(callee)))
	if receiver == nil || staticEvaluator.escapeFacts.explicit[receiver] || staticEvaluator.hasWrites(receiver) {
		return false
	}
	tc := staticEvaluator.typeChecker
	methodSymbol := tc.GetSymbolAtLocation(callee)
	arrayMethod := checker.Checker_getPropertyOfType(tc, tc.GetTypeAtLocation(origin), method)
	// An Object.assign source can replace push. Reuse the checker's property
	// declarations to require the original array method, including through an
	// instantiated array type. Explicit method writes are excluded above.
	return methodSymbol != nil && arrayMethod != nil && len(methodSymbol.Declarations) != 0 &&
		slices.Equal(methodSymbol.Declarations, arrayMethod.Declarations)
}

func (staticEvaluator *StaticStringEvaluator) referenceFlagsFor(symbol *ast.Symbol) staticReferenceFlags {
	staticEvaluator.ensureReferenceFlags()
	return staticEvaluator.referenceFlags[symbol]
}

func (staticEvaluator *StaticStringEvaluator) ensureReferenceFlags() {
	if !staticEvaluator.referenceFlagsComputed && staticEvaluator.sourceFile != nil {
		// A computed mutator name may depend on a symbol the caller is already
		// resolving. Analyze it with an independent recursion stack, then publish
		// the completed facts without disturbing the caller's evaluation.
		analysis := NewStaticStringEvaluatorWithReferenceResolver(
			staticEvaluator.typeChecker, staticEvaluator.sourceFile, staticEvaluator.referenceResolver,
		)
		analysis.GlobalAccess = staticEvaluator.GlobalAccess
		analysis.resolveIdentifiers = staticEvaluator.resolveIdentifiers
		analysis.trackCallEscapes = staticEvaluator.trackCallEscapes
		analysis.computeReferenceFlags()
		staticEvaluator.useReferenceFacts(analysis)
	}
}

func (staticEvaluator *StaticStringEvaluator) hasModifiedIterator(value any) bool {
	staticEvaluator.ensureReferenceFlags()
	if _, ok := value.(*staticArrayValue); ok {
		return staticEvaluator.arrayPrototypeMutated
	}
	if staticValueIsString(value) {
		return staticEvaluator.stringPrototypeMutated
	}
	return staticEvaluator.arrayPrototypeMutated || staticEvaluator.stringPrototypeMutated
}

type staticMutationCandidate struct {
	call, access *ast.Node
}

func (staticEvaluator *StaticStringEvaluator) computeReferenceFlags() {
	parentEvaluation := staticEvaluator.evaluation
	staticEvaluator.evaluation = nil
	defer func() { staticEvaluator.evaluation = parentEvaluation }()

	staticEvaluator.referenceFlagsComputed = true
	staticEvaluator.referenceFlags = nil
	staticEvaluator.bindingTransfers = nil
	staticEvaluator.storedReferences = nil
	staticEvaluator.markedReferences = nil
	if staticEvaluator.trackCallEscapes {
		staticEvaluator.escapeFacts = newStaticEscapeFacts()
		staticEvaluator.escapeFacts.collecting = true
		defer func() { staticEvaluator.escapeFacts.collecting = false }()
	}
	var mutationCandidates []staticMutationCandidate
	var shapeWrites []staticMutationCandidate
	var lengthWrites []staticMutationCandidate
	var calls []*ast.Node
	var stores []struct{ target, value, site *ast.Node }

	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindCallExpression {
			callee := SkipAssertionsAndParens(node.Expression())
			if name, known := AccessExpressionStaticName(callee); known {
				if staticEvaluator.escapeFacts != nil {
					staticEvaluator.escapeFacts.call = node
				}
				staticEvaluator.markCallMutation(node, callee, name)
				if staticEvaluator.escapeFacts != nil {
					staticEvaluator.escapeFacts.call = nil
				}
			} else if ast.IsAccessExpression(callee) {
				mutationCandidates = append(mutationCandidates, staticMutationCandidate{call: node, access: callee})
			}
		}
		if staticEvaluator.trackCallEscapes && node.Kind == ast.KindBinaryExpression {
			assignment := node.AsBinaryExpression()
			if ast.IsIdentifier(SkipAssertionsAndParens(assignment.Left)) && ast.IsAssignmentOperator(assignment.OperatorToken.Kind) {
				staticEvaluator.recordBindingTransfer(assignment.Left, assignment.Right, node, false)
			}
		}
		if node.Kind == ast.KindIdentifier && !IsNonReferenceIdentifier(node) {
			if IsWriteReference(node) {
				staticEvaluator.markReference(node, staticReferenceWrite)
				if staticEvaluator.trackCallEscapes {
					if ast.IsShorthandPropertyAssignment(node.Parent) {
						if fallback := node.Parent.AsShorthandPropertyAssignment().ObjectAssignmentInitializer; fallback != nil {
							staticEvaluator.recordBindingTransfer(node, fallback, node.Parent, false)
						}
					}
					// Native assignment-target traversal includes nested patterns
					// and defaults without treating their keys as references.
					for target := ast.GetAssignmentTarget(node); target != nil; target = ast.GetAssignmentTarget(target) {
						if ast.IsBinaryExpression(target) {
							assignment := target.AsBinaryExpression()
							left := SkipAssertionsAndParens(assignment.Left)
							if ast.IsObjectLiteralExpression(left) || ast.IsArrayLiteralExpression(left) {
								staticEvaluator.recordBindingTransfer(node, assignment.Right, target, true)
							}
						} else if ast.IsForOfStatement(target) {
							staticEvaluator.recordBindingTransfer(node, target.Expression(), target, true)
						}
					}
				}
			}
		}
		if ast.IsAccessExpression(node) {
			outer := node
			for outer.Parent != nil && ast.IsOuterExpression(outer.Parent, skipTransparentKinds) {
				outer = outer.Parent
			}
			if outer.Parent != nil && (ast.IsAssignmentTarget(outer) || outer.Parent.Kind == ast.KindDeleteExpression) {
				flags := staticReferencePropertyMutation
				if staticEvaluator.trackCallEscapes {
					name, known := AccessExpressionStaticName(node)
					if !known || name == "__proto__" || outer.Parent.Kind == ast.KindDeleteExpression {
						flags |= staticReferencePropertyShapeMutation
						shapeWrites = append(shapeWrites, staticMutationCandidate{call: outer.Parent, access: node})
					}
					if known && name == "length" {
						lengthWrites = append(lengthWrites, staticMutationCandidate{call: outer.Parent, access: node})
						// Type assertions can hide an actual write from the native
						// flow graph. Ask the binder about the complete target,
						// including wrappers around the receiver or access itself.
						if !binder.IsNarrowableReference(outer) {
							facts := staticEvaluator.escapeFacts
							if facts.untrackedLengthWrites == nil {
								facts.untrackedLengthWrites = make(map[*ast.Node]bool)
							}
							facts.untrackedLengthWrites[outer.Parent] = true
						}
					}
				}
				staticEvaluator.markReference(AccessExpressionObject(node), flags)
				if staticEvaluator.trackCallEscapes && outer.Parent.Kind == ast.KindBinaryExpression {
					assignment := outer.Parent.AsBinaryExpression()
					switch assignment.OperatorToken.Kind {
					case ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
						stores = append(stores, struct{ target, value, site *ast.Node }{AccessExpressionObject(node), assignment.Right, outer.Parent})
					}
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
		// Process calls after their callee and arguments so nested built-in
		// stores are available when propagating the outer invocation's escapes.
		if staticEvaluator.trackCallEscapes && (node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression || node.Kind == ast.KindTaggedTemplateExpression) {
			calls = append(calls, node)
		}
	}
	visit(&staticEvaluator.sourceFile.Node)
	// Property stores can execute after bindings written later in the source,
	// notably class computed names. Resolve them against the completed bindings.
	for _, store := range stores {
		staticEvaluator.recordStoredReferences(store.target, store.value, store.site, false)
	}

	for _, candidate := range mutationCandidates {
		if methodName, ok := staticEvaluator.evalAccessExpressionKey(candidate.access); ok {
			if staticEvaluator.escapeFacts != nil {
				staticEvaluator.escapeFacts.call = candidate.call
			}
			staticEvaluator.markCallMutation(candidate.call, candidate.access, methodName)
			if staticEvaluator.escapeFacts != nil {
				staticEvaluator.escapeFacts.call = nil
			}
		}
	}
	// Complete the store graph before propagating any invocation through it.
	// A computed class name can execute before an earlier static block.
	for _, call := range calls {
		if !ast.IsCallExpression(call) || len(call.Arguments()) < 2 {
			continue
		}
		for _, builtin := range []string{"Object", "Reflect"} {
			if name, known := staticEvaluator.builtinMethodName(call.Expression(), builtin, make(map[*ast.Symbol]bool)); known {
				staticEvaluator.recordBuiltinTransfers(call, builtin, name)
				break
			}
		}
	}
	if staticEvaluator.escapeFacts != nil {
		// Early mutation flags prevented unsound constant folding while
		// collecting syntax. Replay their sites against the complete graph.
		clear(staticEvaluator.escapeFacts.sites)
	}
	// First collect all explicit writes, so a call cannot be classified using
	// a binding whose reassignment appears later in the file.
	staticEvaluator.evaluation = &staticEvaluationState{
		values:                make(map[*ast.Node]staticEvalResult),
		optionalShortCircuits: make(map[*ast.Node]bool),
	}
	for _, call := range calls {
		staticEvaluator.escapeFacts.call = call
		staticEvaluator.markedReferences = nil
		staticEvaluator.markCallEscapes(call)
	}
	// Direct deletes and computed writes must follow later-collected bindings
	// just like call mutations, including aliases assigned in computed names.
	for _, write := range shapeWrites {
		staticEvaluator.escapeFacts.call = write.call
		staticEvaluator.markedReferences = nil
		staticEvaluator.markReference(AccessExpressionObject(write.access), staticReferencePropertyMutation|staticReferencePropertyShapeMutation)
	}
	// Reuse the completed alias/store graph and per-mutation sites. Length
	// writes affect indexed reads, without invalidating unrelated named fields
	// or values captured before the truncation.
	for _, write := range lengthWrites {
		staticEvaluator.escapeFacts.call = write.call
		staticEvaluator.markedReferences = nil
		staticEvaluator.markReference(AccessExpressionObject(write.access), staticReferencePropertyMutation|staticReferenceArrayLengthMutation)
	}
	if staticEvaluator.escapeFacts != nil {
		staticEvaluator.escapeFacts.call = nil
	}
}

func (staticEvaluator *StaticStringEvaluator) markCallMutation(call, callee *ast.Node, name string) {
	object := SkipAssertionsAndParens(AccessExpressionObject(callee))
	if isMutatingArrayMethod(name) {
		staticEvaluator.markReference(object, staticReferencePropertyMutation)
	}
	var objectMutation, reflectMutation bool
	switch name {
	case "defineProperty", "setPrototypeOf":
		objectMutation, reflectMutation = true, true
	case "defineProperties", "assign":
		objectMutation = true
	case "deleteProperty", "set":
		reflectMutation = true
	}
	if len(call.Arguments()) > 0 && ((objectMutation && staticEvaluator.isBuiltinIdentifier(object, "Object")) ||
		(reflectMutation && staticEvaluator.isBuiltinIdentifier(object, "Reflect"))) {
		staticEvaluator.markMutationArgument(call.Arguments(), 0)
		// Reflect.set can define the property on its explicit receiver even
		// when its target is a different object.
		if name == "set" {
			staticEvaluator.markMutationArgument(call.Arguments(), 3)
		}
	}
}

func (staticEvaluator *StaticStringEvaluator) markMutationArgument(arguments []*ast.Node, position int) {
	spreadSeen := false
	for _, argument := range arguments {
		if ast.IsSpreadElement(argument) {
			// A spread may occupy the requested position or be empty.
			staticEvaluator.markReference(argument.Expression(), staticReferencePropertyMutation|staticReferenceNestedMutation)
			spreadSeen = true
			continue
		}
		if position == 0 || spreadSeen {
			staticEvaluator.markReference(argument, staticReferencePropertyMutation)
		}
		if position == 0 {
			return
		}
		position--
	}
}

func (staticEvaluator *StaticStringEvaluator) markReference(node *ast.Node, flag staticReferenceFlags) {
	node = SkipAssertionsAndParens(node)
	followTransfers := flag&(staticReferenceCallEscape|staticReferenceArrayLengthMutation) != 0 ||
		flag&staticReferencePropertyMutation != 0 && staticEvaluator.escapeFacts != nil && staticEvaluator.escapeFacts.call != nil
	if flag&staticReferenceCallEscape != 0 && node != nil {
		value := staticEvaluator.evalValue(node)
		if value.ok && IsPrimitiveStaticValue(value.value) {
			return
		}
	}
	for node != nil {
		if !ast.IsIdentifier(node) || followTransfers {
			// Different reads of one mutable binding can capture different
			// objects, even when they reach the same later mutation.
			if staticEvaluator.markedReferences[node]&flag == flag {
				return
			}
			if staticEvaluator.markedReferences == nil {
				staticEvaluator.markedReferences = make(map[*ast.Node]staticReferenceFlags)
			}
			staticEvaluator.markedReferences[node] |= flag
		}
		if !ast.IsIdentifier(node) {
			break
		}
		if flag&staticReferenceNestedMutation != 0 && flag&staticReferenceCallReceiver == 0 {
			// Destructuring a constructor can reach its prototype without a
			// property-access node, e.g. const {prototype: proto} = Array.
			array := staticEvaluator.isBuiltinIdentifier(node, "Array")
			str := staticEvaluator.isBuiltinIdentifier(node, "String")
			if array && !staticEvaluator.arrayPrototypeMutated || str && !staticEvaluator.stringPrototypeMutated {
				staticEvaluator.invalidateMutationValues()
			}
			staticEvaluator.arrayPrototypeMutated = staticEvaluator.arrayPrototypeMutated || array
			staticEvaluator.stringPrototypeMutated = staticEvaluator.stringPrototypeMutated || str
		}
		symbol := staticEvaluator.referenceSymbol(node)
		if symbol == nil {
			return
		}
		if facts := staticEvaluator.escapeFacts; facts != nil && flag&staticReferencePropertyMutation != 0 {
			if facts.call == nil {
				facts.explicit[symbol] = true
				if staticEvaluator.referenceFlags[symbol]&flag == flag {
					return
				}
			} else {
				sites := facts.sites
				if flag&staticReferenceArrayLengthMutation != 0 {
					if facts.lengthWrites == nil {
						facts.lengthWrites = make(map[*ast.Symbol]map[*ast.Node]staticReferenceFlags)
					}
					sites = facts.lengthWrites
				}
				if sites[symbol] == nil {
					sites[symbol] = make(map[*ast.Node]staticReferenceFlags)
				}
				sites[symbol][facts.call] |= flag
			}
		} else if staticEvaluator.referenceFlags[symbol]&flag == flag {
			return
		}
		if staticEvaluator.referenceFlags == nil {
			staticEvaluator.referenceFlags = make(map[*ast.Symbol]staticReferenceFlags, 8)
		}
		if staticEvaluator.referenceFlags[symbol]&flag != flag {
			staticEvaluator.referenceFlags[symbol] |= flag
			staticEvaluator.invalidateMutationValues()
		}
		var latestBinding *ast.Node
		if followTransfers {
			latestBinding = staticEvaluator.latestBindingAt(symbol, node)
			// Rebinding checks do not depend on property stores. Keep their
			// indexes separate so repeated direct writes do not scan every
			// value stored on the same object.
			groups := [2][]staticReferenceTransfer{staticEvaluator.bindingTransfers[symbol]}
			if flag&(staticReferenceCallEscape|staticReferenceNestedMutation) != 0 {
				groups[1] = staticEvaluator.storedReferences[symbol]
			}
			for _, transfers := range groups {
				for _, transfer := range transfers {
					point := staticEvaluator.escapeFacts.call
					if transfer.binding {
						point = node
					}
					if !staticEvaluator.transferAvailableAt(transfer, point) || staticEvaluator.bindingTransferReplaced(transfer, latestBinding) {
						continue
					}
					if transfer.copied {
						transferredFlags := flag &^ staticReferenceCallReceiver
						if transfer.binding {
							transferredFlags |= staticReferenceNestedMutation
						}
						staticEvaluator.markCopiedReferences(transfer.value, transferredFlags)
					} else {
						staticEvaluator.markReference(transfer.value, flag&^staticReferenceCallReceiver)
					}
				}
			}
		}
		if latestBinding != nil && followTransfers {
			return
		}
		if flag&staticReferencePropertyMutation == 0 || len(symbol.Declarations) != 1 {
			return
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind == ast.KindBindingElement {
			root := ast.GetRootDeclaration(declaration)
			for binding := declaration; binding.Kind == ast.KindBindingElement; binding = binding.Parent.Parent {
				element := binding.AsBindingElement()
				// Rest creates a fresh container. Only a nested mutation can
				// reach the references copied into it.
				if element.DotDotDotToken != nil && flag&staticReferenceNestedMutation == 0 {
					return
				}
				if element.DotDotDotToken != nil && flag&staticReferenceCallEscape != 0 &&
					staticEvaluator.hasOnlyPrimitiveChildren(root.Initializer()) {
					return
				}
				staticEvaluator.markReference(element.Initializer, flag)
				flag |= staticReferenceNestedMutation
			}
			declaration = root
		}
		if declaration.Kind != ast.KindVariableDeclaration && declaration.Kind != ast.KindParameter {
			if followTransfers {
				staticEvaluator.markReturnedReferences(declaration, flag)
			}
			return
		}
		// A local alias may still share its initializer's aggregate even when
		// the binding has other writes. Propagate only property mutations;
		// replacing the alias itself must not taint the original value. The
		// already-marked check above also terminates circular alias chains.
		node = SkipAssertionsAndParens(declaration.Initializer())
		if node == nil && declaration.Parent != nil && declaration.Parent.Parent != nil && ast.IsForOfStatement(declaration.Parent.Parent) {
			node = declaration.Parent.Parent.Expression()
			flag |= staticReferenceNestedMutation
		}
	}
	if node == nil || flag&staticReferencePropertyMutation == 0 {
		return
	}
	// Follow only expressions that can return the same reference. Conditions,
	// computed keys and function bodies do not alias the expression's value.
	switch node.Kind {
	case ast.KindConditionalExpression:
		expression := node.AsConditionalExpression()
		staticEvaluator.markReference(expression.WhenTrue, flag)
		staticEvaluator.markReference(expression.WhenFalse, flag)
	case ast.KindBinaryExpression:
		expression := node.AsBinaryExpression()
		switch expression.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
			ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
			staticEvaluator.markReference(expression.Left, flag)
			staticEvaluator.markReference(expression.Right, flag)
		case ast.KindCommaToken, ast.KindEqualsToken:
			staticEvaluator.markReference(expression.Right, flag)
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		staticEvaluator.markReference(AccessExpressionObject(node), (flag&^staticReferenceCallReceiver)|staticReferenceNestedMutation)
	case ast.KindObjectLiteralExpression:
		if flag&staticReferenceNestedMutation != 0 {
			for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
				switch property.Kind {
				case ast.KindPropertyAssignment:
					staticEvaluator.markReference(property.AsPropertyAssignment().Initializer, flag&^staticReferenceCallReceiver)
				case ast.KindShorthandPropertyAssignment:
					staticEvaluator.markReference(property.Name(), flag&^staticReferenceCallReceiver)
				case ast.KindSpreadAssignment:
					staticEvaluator.markCopiedReferences(property.Expression(), flag)
				case ast.KindMethodDeclaration, ast.KindGetAccessor:
					if followTransfers {
						staticEvaluator.markReturnedReferences(property, flag)
					}
				}
			}
		}
	case ast.KindArrayLiteralExpression:
		if flag&staticReferenceNestedMutation != 0 {
			for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
				if ast.IsSpreadElement(element) {
					staticEvaluator.markCopiedReferences(element.Expression(), flag)
					continue
				}
				staticEvaluator.markReference(element, flag&^staticReferenceCallReceiver)
			}
		}
	case ast.KindCallExpression:
		if flag&staticReferenceCallEscape != 0 &&
			!staticEvaluator.isBuiltinMethodValue(node.Expression(), "Array", "of", map[*ast.Symbol]bool{}) &&
			!staticEvaluator.isKnownObjectOperation(node, node.Expression()) {
			staticEvaluator.markCallReceiver(node.Expression(), flag|staticReferenceCallReceiver)
		}
		if argument, ok := staticEvaluator.objectPassThroughArgument(node); ok {
			staticEvaluator.markReference(argument, flag)
		} else if staticEvaluator.isBuiltinMethodValue(node.Expression(), "Object", "assign", map[*ast.Symbol]bool{}) {
			for _, argument := range node.Arguments() {
				spread := ast.IsSpreadElement(argument)
				argumentFlag := flag
				if spread {
					argument = argument.Expression()
					argumentFlag |= staticReferenceNestedMutation
				}
				if spread {
					staticEvaluator.markCopiedReferences(argument, argumentFlag)
				} else if argumentFlag&staticReferenceNestedMutation != 0 {
					staticEvaluator.markCopiedReferences(argument, argumentFlag)
				} else {
					staticEvaluator.markReference(argument, argumentFlag)
				}
				// A leading spread may be empty, so the next argument can
				// still be the target returned by Object.assign.
				if !spread && flag&staticReferenceNestedMutation == 0 {
					break
				}
			}
		} else if flag&staticReferenceNestedMutation != 0 {
			callee := SkipAssertionsAndParens(node.Expression())
			if name, ok := staticEvaluator.evalAccessExpressionKey(callee); ok && name == "of" &&
				staticEvaluator.isBuiltinArrayValue(AccessExpressionObject(callee), map[*ast.Symbol]bool{}) {
				for _, argument := range node.Arguments() {
					if ast.IsSpreadElement(argument) {
						staticEvaluator.markCopiedReferences(argument.Expression(), flag)
					} else {
						staticEvaluator.markReference(argument, flag)
					}
				}
			}
		}
	case ast.KindNewExpression:
		if flag&staticReferenceCallEscape != 0 {
			staticEvaluator.markReference(node.Expression(), flag)
		}
	case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindClassExpression:
		if followTransfers {
			staticEvaluator.markReturnedReferences(node, flag)
		}
	}
}

func (staticEvaluator *StaticStringEvaluator) referenceSymbol(node *ast.Node) *ast.Symbol {
	if staticEvaluator.referenceResolver != nil {
		if symbol := staticEvaluator.referenceResolver.Resolve(node); symbol != nil {
			return symbol
		}
	}
	return GetReferenceSymbol(node, staticEvaluator.typeChecker)
}

func isMutatingArrayMethod(name string) bool {
	switch name {
	case "copyWithin", "fill", "pop", "push", "reverse", "shift", "sort", "splice", "unshift":
		return true
	default:
		return false
	}
}

func staticValueIsTsgoSafe(value any) bool {
	switch value.(type) {
	case bool, staticNullValue, staticUndefinedValue, staticNumberValue, staticSymbolValue, *big.Int, *staticStringNode, *staticRegExpNode, *staticObjectValue, *staticArrayValue:
		return false
	default:
		return value != nil
	}
}

func staticValueAsString(value any) (string, bool) {
	switch value := value.(type) {
	case string:
		return value, true
	case *staticStringNode:
		node := (*ast.Node)(value)
		switch node.Kind {
		case ast.KindStringLiteral:
			return node.AsStringLiteral().Text, true
		case ast.KindNoSubstitutionTemplateLiteral:
			return node.AsNoSubstitutionTemplateLiteral().Text, true
		}
	}
	return "", false
}

func staticValueIsString(value any) bool {
	_, ok := staticValueAsString(value)
	return ok
}

func staticValueNullish(value any) bool {
	switch value.(type) {
	case staticNullValue, staticUndefinedValue:
		return true
	default:
		return false
	}
}

func staticValueUndefined(value any) bool {
	_, ok := value.(staticUndefinedValue)
	return ok
}

// staticValueKind is the JavaScript typeof-like classification `===` needs.
// Objects and arrays compare by identity, which folding does not model.
type staticValueKind uint8

const (
	staticKindUnknown staticValueKind = iota
	staticKindString
	staticKindNumber
	staticKindBigInt
	staticKindBoolean
	staticKindSymbol
	staticKindNull
	staticKindUndefined
)

func staticValueKindOf(value any) staticValueKind {
	switch value.(type) {
	case *big.Int:
		return staticKindBigInt
	case bool:
		return staticKindBoolean
	case staticSymbolValue:
		return staticKindSymbol
	case staticNullValue:
		return staticKindNull
	case staticUndefinedValue:
		return staticKindUndefined
	}
	if staticValueIsString(value) {
		return staticKindString
	}
	if _, isNumber := value.(interface{ IsNaN() bool }); isNumber {
		return staticKindNumber
	}
	return staticKindUnknown
}

// staticSameValueZeroKey is a comparable map key for a folded primitive.
type staticSameValueZeroKey struct {
	kind  staticValueKind
	value any
}

// StaticValueSameValueZeroKey returns a comparable key such that two folded
// primitives have equal keys exactly when StaticValuesSameValueZero reports
// them equal, so callers can index values instead of comparing pairwise. It
// reports false for values without a modeled primitive identity.
func StaticValueSameValueZeroKey(value any) (any, bool) {
	kind := staticValueKindOf(value)
	switch kind {
	case staticKindBigInt:
		integer, ok := value.(*big.Int)
		if !ok {
			return nil, false
		}
		return staticSameValueZeroKey{kind, integer.String()}, true
	case staticKindString:
		text, _ := staticValueAsString(value)
		return staticSameValueZeroKey{kind, text}, true
	case staticKindNumber:
		number, ok := staticValueToNumber(value)
		if !ok {
			return nil, false
		}
		if math.IsNaN(number) {
			return staticSameValueZeroKey{kind, "NaN"}, true
		}
		if number == 0 {
			number = 0 // SameValueZero: -0 and +0 share a key.
		}
		return staticSameValueZeroKey{kind, number}, true
	case staticKindBoolean:
		flag, ok := value.(bool)
		if !ok {
			return nil, false
		}
		return staticSameValueZeroKey{kind, flag}, true
	case staticKindSymbol:
		symbol, ok := value.(staticSymbolValue)
		if !ok {
			return nil, false
		}
		return staticSameValueZeroKey{kind, symbol}, true
	case staticKindNull, staticKindUndefined:
		return staticSameValueZeroKey{kind: kind}, true
	}
	return nil, false
}

// StaticValuesSameValueZero compares two folded primitives as JavaScript Map
// keys. Aggregate values have no modeled identity.
func StaticValuesSameValueZero(left, right any) (equal, known bool) {
	if !IsPrimitiveStaticValue(left) || !IsPrimitiveStaticValue(right) {
		return false, false
	}
	if leftNumber, ok := left.(jsnum.Number); ok && leftNumber.IsNaN() {
		rightNumber, ok := right.(jsnum.Number)
		return ok && rightNumber.IsNaN(), true
	}
	return staticValuesStrictEqual(left, right)
}

// staticValuesStrictEqual implements the strict equality comparison over folded
// values. Comparing across kinds is always false, including `1n === 1`.
func staticValuesStrictEqual(left any, right any) (equal bool, ok bool) {
	leftKind := staticValueKindOf(left)
	rightKind := staticValueKindOf(right)
	if leftKind == staticKindUnknown || rightKind == staticKindUnknown {
		if staticValueIsAggregate(left) && rightKind != staticKindUnknown ||
			staticValueIsAggregate(right) && leftKind != staticKindUnknown {
			return false, true
		}
		return false, false
	}
	if leftKind != rightKind {
		return false, true
	}

	switch leftKind {
	case staticKindBigInt:
		leftInteger, leftOK := left.(*big.Int)
		rightInteger, rightOK := right.(*big.Int)
		if !leftOK || !rightOK {
			return false, false
		}
		return leftInteger.Cmp(rightInteger) == 0, true
	case staticKindString:
		leftText, _ := staticValueAsString(left)
		rightText, _ := staticValueAsString(right)
		return ecmascript.CompareStrings(leftText, rightText) == 0, true
	case staticKindNumber:
		leftNumber, leftOk := staticValueToNumber(left)
		rightNumber, rightOk := staticValueToNumber(right)
		if !leftOk || !rightOk {
			return false, false
		}
		return leftNumber == rightNumber, true
	case staticKindBoolean:
		leftFlag, leftOk := left.(bool)
		rightFlag, rightOk := right.(bool)
		if !leftOk || !rightOk {
			return false, false
		}
		return leftFlag == rightFlag, true
	case staticKindSymbol:
		leftSymbol, leftOK := left.(staticSymbolValue)
		rightSymbol, rightOK := right.(staticSymbolValue)
		known := leftOK && rightOK
		return known && leftSymbol == rightSymbol, known
	}
	return true, true
}

func staticValueIsAggregate(value any) bool {
	switch value.(type) {
	case *staticRegExpNode, *staticObjectValue, *staticArrayValue:
		return true
	default:
		return false
	}
}

func staticValueTruthy(value any) (truthy bool, ok bool) {
	switch value := value.(type) {
	case *big.Int:
		return value.Sign() != 0, true
	case string:
		return value != "", true
	case staticNumberValue:
		return value != 0 && !value.IsNaN(), true
	case *staticStringNode:
		stringValue, _ := staticValueAsString(value)
		return stringValue != "", true
	case bool:
		return value, true
	case staticSymbolValue:
		return true, true
	case staticNullValue, staticUndefinedValue:
		return false, true
	case *staticRegExpNode, *staticObjectValue, *staticArrayValue:
		return true, true
	default:
		return evaluatorBool(func() bool { return evaluator.IsTruthy(value) })
	}
}

func staticValueToString(value any) (string, bool) {
	switch value := value.(type) {
	case *big.Int:
		return value.String(), true
	case string:
		return value, true
	case staticNumberValue:
		return ecmascript.NumberToString(float64(value)), true
	case *staticStringNode:
		return staticValueAsString(value)
	case *staticRegExpNode:
		return RegExpLiteralStringValue((*ast.Node)(value).Text()), true
	case bool:
		if value {
			return "true", true
		}
		return "false", true
	case staticNullValue:
		return "null", true
	case staticUndefinedValue:
		return "undefined", true
	case staticSymbolValue:
		// JavaScript's implicit ToString(Symbol) throws. Symbol keys remain
		// first-class values and are never coerced into invented string names.
		return "", false
	case *staticArrayValue:
		return staticArrayToString(value)
	case *staticObjectValue:
		if _, overridesToString := staticObjectOwnProperty(value, "toString"); overridesToString || value.prototypeSet {
			return "", false
		}
		return "[object Object]", true
	default:
		return evaluatorString(func() string { return evaluator.AnyToString(value) })
	}
}

// staticArrayToString implements `Array.prototype.join(',')`, which is what
// an array coerces to via the default ToString. Nullish elements fold to "",
// not "null"/"undefined" as a direct string conversion would give.
func staticArrayToString(value *staticArrayValue) (string, bool) {
	return staticArrayJoin(value, ",")
}

func staticArrayJoin(value *staticArrayValue, separator string) (string, bool) {
	if value == nil || value.length < 0 {
		return "", false
	}
	var builder strings.Builder
	capacity := 0
	if value.length > 1 {
		separatorCount := value.length - 1
		if len(separator) > 0 && separatorCount > maxStaticStringLength/len(separator) {
			return "", false
		}
		capacity = separatorCount * len(separator)
	}
	for i := range value.length {
		element := value.element(i)
		elementLength := 0
		switch element := element.(type) {
		case string:
			elementLength = len(element)
		case *staticStringNode:
			stringValue, _ := staticValueAsString(element)
			elementLength = len(stringValue)
		case bool:
			if element {
				elementLength = len("true")
			} else {
				elementLength = len("false")
			}
		}
		if elementLength > maxStaticStringLength-capacity {
			return "", false
		}
		capacity += elementLength
	}
	builder.Grow(capacity)
	for i := range value.length {
		element := value.element(i)
		if i > 0 {
			if len(separator) > maxStaticStringLength-builder.Len() {
				return "", false
			}
			builder.WriteString(separator)
		}
		if staticValueNullish(element) {
			continue
		}
		part, ok := staticValueToString(element)
		if !ok {
			return "", false
		}
		if len(part) > maxStaticStringLength-builder.Len() {
			return "", false
		}
		builder.WriteString(part)
	}
	return ecmascript.CombineSurrogatePairs(builder.String()), true
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
