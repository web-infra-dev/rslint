// cspell:ignore Unscopables unscopables

package utils

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/evaluator"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// StaticStringEvaluator folds expressions to string constants. It wraps tsgo's
// evaluator and adds stable local variable resolution through the TypeChecker.
// Create one evaluator per linted file; it keeps write-reference and recursion
// state for that file.
type StaticStringEvaluator struct {
	typeChecker            *checker.Checker
	sourceFile             *ast.SourceFile
	referenceResolver      StaticReferenceResolver
	evaluator              evaluator.Evaluator
	resolveIdentifiers     bool
	resolving              map[*ast.Symbol]bool
	referenceFlagsComputed bool
	referenceFlags         map[*ast.Symbol]staticReferenceFlags
	eslintStaticCalls      bool
	evaluationDepth        int
	evaluationPeak         int
	evaluationGeneration   uint64
	evaluationStepsLeft    int
	stringBudgetLeft       int
	stringWorkBudgetLeft   int
	stringWorkContext      *staticStringWorkContext
	bigIntWorkBudgetLeft   int
	bigIntWorkContext      *staticStringWorkContext
	bigIntCacheBudgetLeft  int
	bigIntExpressionCache  map[*ast.Node]staticBigIntCacheEntry
	aggregateBudgetLeft    int
	aggregateIdentities    map[any]struct{}
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

// StaticGlobalReferenceResolver lets a binder-only evaluator distinguish an
// unresolved global from a local binding without falling back to a whole-file
// syntactic shadowing scan.
type StaticGlobalReferenceResolver interface {
	StaticReferenceResolver
	IsUnshadowedGlobal(node *ast.Node, name string) bool
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

// NewStaticStringEvaluatorForSourceOnlyStaticValue enables the native-call
// subset used by eslint-utils getStaticValue while deliberately omitting a
// TypeChecker. It is opt-in so existing string-evaluator consumers retain
// their exact evaluation surface.
func NewStaticStringEvaluatorForSourceOnlyStaticValue(
	sourceFile *ast.SourceFile,
	referenceResolver StaticReferenceResolver,
) *StaticStringEvaluator {
	staticEvaluator := NewStaticStringEvaluatorWithReferenceResolver(nil, sourceFile, referenceResolver)
	staticEvaluator.eslintStaticCalls = true
	staticEvaluator.stringWorkBudgetLeft = maxStaticStringWorkBudget
	staticEvaluator.stringWorkContext = &staticStringWorkContext{remaining: &staticEvaluator.stringWorkBudgetLeft}
	staticEvaluator.bigIntWorkBudgetLeft = maxStaticBigIntWorkBudget
	staticEvaluator.bigIntWorkContext = &staticStringWorkContext{remaining: &staticEvaluator.bigIntWorkBudgetLeft}
	staticEvaluator.bigIntCacheBudgetLeft = maxStaticBigIntCacheBudget
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
type staticUnknownNumberValue struct{}
type staticUnknownStringValue struct {
	truthy     bool
	numericNaN bool
}
type staticUnknownBooleanValue struct{}
type staticIdentity struct{ _ byte }
type staticOpaqueObjectValue struct{ identity *staticIdentity }
type staticIteratorValue struct {
	values            []any
	kind              string
	identity          *staticIdentity
	stringWorkContext *staticStringWorkContext
}
type staticOptionalChainShortCircuitValue struct{}
type staticRegExpValue struct {
	source            string
	flags             string
	identity          any
	stringWorkContext *staticStringWorkContext
}
type staticDateValue struct {
	milliseconds staticNumberValue
	known        bool
	identity     *staticIdentity
}
type staticBoxedValue struct {
	value             any
	identity          *staticIdentity
	stringWorkContext *staticStringWorkContext
}

type staticCollectionValue struct {
	kind              string
	entries           []staticCollectionEntry
	index             map[staticCollectionKey]int
	identity          *staticIdentity
	stringWorkContext *staticStringWorkContext
}

type staticCollectionEntry struct {
	key   any
	value any
}

type staticCollectionKey struct {
	kind       uint8
	numberBits uint64
	text       string
	identity   any
}

type staticBigIntValue struct {
	value             *big.Int
	truthy            bool
	truthyKnown       bool
	stringWorkContext *staticStringWorkContext
	bigIntWorkContext *staticStringWorkContext
}

func staticAbstractBigInt(truthy bool) staticBigIntValue {
	return staticBigIntValue{truthy: truthy, truthyKnown: true}
}

type staticSymbolValue struct {
	// description is the registry key or well-known property name used for
	// identity. staticSymbolDescription derives the observable description.
	description   string
	global        bool
	wellKnown     bool
	hostDependent bool
}

type staticBuiltinValue string
type staticBuiltinObjectValue string

// staticNumberValue is a number this evaluator computed itself. tsgo hands
// numbers back as jsnum.Number, a type internal/utils cannot import; the IsNaN
// method both types carry is what tells a folded number from a folded bigint.
type staticNumberValue float64

func (value staticNumberValue) IsNaN() bool {
	return math.IsNaN(float64(value))
}

// staticStringNode keeps literal strings backed by their existing AST node so
// nested aggregate evaluation doesn't allocate an interface box per literal.
type staticStringNode ast.Node

// staticObjectValue and staticArrayValue hold folded object and array literals
// so member access and JavaScript's default string coercion can be modeled.
type staticObjectValue struct {
	property          staticObjectProperty
	extraProperties   []staticObjectProperty
	propertyCount     int
	propertyIndices   map[string]int
	prototype         any
	prototypeSet      bool
	enhancedCoercion  bool
	stringWorkContext *staticStringWorkContext
}

type staticObjectProperty struct {
	name      string
	symbol    staticSymbolValue
	symbolKey bool
	value     any
}

type staticArrayValue struct {
	length            int
	inline            [2]any
	overflow          []any
	omitted           map[int]bool
	stringWorkContext *staticStringWorkContext
}

type staticStringWorkContext struct {
	remaining *int
}

func (context *staticStringWorkContext) reserve(cost int) bool {
	if context == nil || context.remaining == nil || cost <= 0 {
		return true
	}
	if cost > *context.remaining {
		*context.remaining = 0
		return false
	}
	*context.remaining -= cost
	return true
}

func (context *staticStringWorkContext) exhaust() {
	if context != nil && context.remaining != nil {
		*context.remaining = 0
	}
}

func (context *staticStringWorkContext) exhausted() bool {
	return context != nil && context.remaining != nil && *context.remaining <= 0
}

type staticReferenceFlags uint8

const (
	staticReferenceWrite staticReferenceFlags = 1 << iota
	staticReferencePropertyMutation
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
const maxStaticStringBudget = maxStaticStringLength * 8
const maxStaticStringWorkBudget = maxStaticStringBudget
const maxJavaScriptStringLength = 1<<29 - 24
const maxStaticAggregateElements = 1 << 14
const maxStaticAggregateBudget = maxStaticAggregateElements * 4

// BigInt results beyond 64 Ki bits are represented by the facts the evaluator
// can still prove (currently truthiness) instead of being materialized. This
// keeps repeated source-only receiver classification linear in source size;
// the separate JavaScript limit below still distinguishes results that V8
// would reject from large values that are merely abstracted locally.
const maxStaticBigIntBits = 1 << 16
const maxJavaScriptBigIntBits = 1 << 30
const maxStaticBigIntWorkBudget = 1 << 23
const maxStaticBigIntCacheBudget = 1 << 20
const maxStaticEvaluationSteps = 1 << 16
const maxStaticEvaluationDepth = 1 << 10

type staticEvalResult struct {
	value                  any
	ok                     bool
	stringCodeUnits        int
	stringCodeUnitsKnown   bool
	stringBudgetGeneration uint64
	stringWorkCharged      bool
	bigIntWorkCharged      bool
	bigIntWorkBytes        int
	cacheableBigInt        bool
}

type staticBigIntCacheEntry struct {
	result staticEvalResult
	depth  int
	steps  int
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
	if prototype, ok := result.value.(staticBuiltinObjectValue); ok && prototype == "Array.prototype" {
		isArray = true
	}
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

func (staticEvaluator *StaticStringEvaluator) evalValue(node *ast.Node) staticEvalResult {
	if !staticEvaluator.eslintStaticCalls {
		return staticEvaluator.evalValueUnbudgeted(node)
	}
	if staticEvaluator.stringWorkBudgetLeft == 0 || staticEvaluator.bigIntWorkBudgetLeft == 0 {
		return staticEvalResult{}
	}
	// eslint-utils evaluates through the JavaScript call stack and turns a
	// RangeError into an unknown static value. Keep enough headroom below the
	// tested default Node stack limits that Go's growable stack cannot prove a
	// value in cases where upstream can fail from recursive evaluation.
	if staticEvaluator.evaluationDepth >= maxStaticEvaluationDepth {
		return staticEvalResult{}
	}
	if staticEvaluator.evaluationDepth == 0 {
		staticEvaluator.evaluationGeneration++
		if staticEvaluator.evaluationGeneration == 0 {
			staticEvaluator.evaluationGeneration++
		}
		staticEvaluator.evaluationStepsLeft = maxStaticEvaluationSteps
		staticEvaluator.evaluationPeak = 0
		staticEvaluator.stringBudgetLeft = maxStaticStringBudget
		staticEvaluator.aggregateBudgetLeft = maxStaticAggregateBudget
		if staticEvaluator.aggregateIdentities == nil {
			staticEvaluator.aggregateIdentities = map[any]struct{}{}
		} else {
			clear(staticEvaluator.aggregateIdentities)
		}
	}
	if staticEvaluator.evaluationStepsLeft == 0 {
		return staticEvalResult{}
	}
	stepsBefore := staticEvaluator.evaluationStepsLeft
	staticEvaluator.evaluationStepsLeft--
	depthBefore := staticEvaluator.evaluationDepth
	staticEvaluator.evaluationDepth++
	cacheNode := SkipAssertionsAndParens(node)
	entry, cached := staticEvaluator.bigIntExpressionCache[cacheNode]
	var result staticEvalResult
	cacheDepth, cacheSteps := 0, 0
	if cached {
		additionalSteps := entry.steps - 1
		if entry.depth > maxStaticEvaluationDepth-depthBefore ||
			additionalSteps > staticEvaluator.evaluationStepsLeft {
			if additionalSteps > staticEvaluator.evaluationStepsLeft {
				staticEvaluator.evaluationStepsLeft = 0
			}
			staticEvaluator.evaluationDepth--
			return staticEvalResult{}
		}
		staticEvaluator.evaluationStepsLeft -= additionalSteps
		staticEvaluator.evaluationPeak = max(
			staticEvaluator.evaluationPeak,
			depthBefore+entry.depth,
		)
		result = entry.result
	} else {
		previousPeak := staticEvaluator.evaluationPeak
		staticEvaluator.evaluationPeak = staticEvaluator.evaluationDepth
		result = staticEvaluator.evalValueUnbudgeted(node)
		cacheDepth = staticEvaluator.evaluationPeak - depthBefore
		cacheSteps = stepsBefore - staticEvaluator.evaluationStepsLeft
		staticEvaluator.evaluationPeak = max(previousPeak, staticEvaluator.evaluationPeak)
	}
	staticEvaluator.evaluationDepth--
	if staticEvaluator.stringWorkBudgetLeft == 0 {
		return staticEvalResult{}
	}
	if !result.ok {
		return result
	}
	if !result.stringCodeUnitsKnown {
		var text string
		var isString bool
		var derivedString bool
		switch value := result.value.(type) {
		case string:
			text, isString = value, true
			derivedString = true
		case *staticStringNode:
			text, isString = staticValueAsString(value)
		}
		if isString {
			if len(text) > maxStaticStringLength*3 {
				if derivedString {
					staticEvaluator.stringWorkBudgetLeft = 0
				}
				return staticEvalResult{}
			}
			result.stringCodeUnits = ecmascript.StringCodeUnitCount(text)
			result.stringCodeUnitsKnown = true
		}
	}
	if result.stringCodeUnitsKnown && result.stringCodeUnits > maxStaticStringLength {
		if _, derivedString := result.value.(string); derivedString {
			staticEvaluator.stringWorkBudgetLeft = 0
		}
		return staticEvalResult{}
	}
	if !staticEvaluator.chargeStaticStringResult(&result) {
		return staticEvalResult{}
	}
	if !staticEvaluator.chargeStaticBigIntResult(&result) {
		return staticEvalResult{}
	}
	result.value = staticEvaluator.attachStaticResourceBudgets(result.value)
	cost := staticAggregateShallowCost(result.value)
	identity := staticAggregateIdentity(result.value)
	if identity != nil {
		if staticEvaluator.aggregateIdentities == nil {
			staticEvaluator.aggregateIdentities = map[any]struct{}{}
		}
		if _, alreadyCharged := staticEvaluator.aggregateIdentities[identity]; alreadyCharged {
			cost = 0
		}
	}
	if cost > staticEvaluator.aggregateBudgetLeft {
		return staticEvalResult{}
	}
	staticEvaluator.aggregateBudgetLeft -= cost
	if identity != nil {
		staticEvaluator.aggregateIdentities[identity] = struct{}{}
	}
	if !cached {
		staticEvaluator.cacheStaticBigIntExpression(cacheNode, result, cacheDepth, cacheSteps)
	}
	return result
}

// chargeStaticBigIntResult bounds repeated materialization of exact BigInt
// values across every receiver in one source file. Cache hits retain the same
// immutable big.Int and therefore carry the charged bit with them.
func (staticEvaluator *StaticStringEvaluator) chargeStaticBigIntResult(result *staticEvalResult) bool {
	if result == nil || result.bigIntWorkCharged {
		return true
	}
	cost := result.bigIntWorkBytes
	value, ok := result.value.(staticBigIntValue)
	if !ok {
		return true
	}
	if value.value != nil {
		cost = max(cost, max((value.value.BitLen()+7)/8, 32))
	}
	if cost <= 0 {
		result.bigIntWorkCharged = true
		return true
	}
	if cost > staticEvaluator.bigIntWorkBudgetLeft {
		staticEvaluator.bigIntWorkBudgetLeft = 0
		return false
	}
	staticEvaluator.bigIntWorkBudgetLeft -= cost
	result.bigIntWorkCharged = true
	return true
}

// cacheStaticBigIntExpression retains only closed BigInt literal/operator
// trees. Identifier evaluation clears cacheableBigInt, so an alias or
// conditional chain must still traverse every binding and cannot use cached
// suffixes to evade the recursion and step limits.
func (staticEvaluator *StaticStringEvaluator) cacheStaticBigIntExpression(
	node *ast.Node,
	result staticEvalResult,
	depth int,
	steps int,
) {
	if node == nil || !result.ok || !result.cacheableBigInt ||
		!isCacheableBigIntExpressionNode(node) || depth <= 0 || steps <= 0 ||
		staticEvaluator.bigIntCacheBudgetLeft <= 0 {
		return
	}
	value, ok := result.value.(staticBigIntValue)
	if !ok {
		return
	}
	cost := 64
	if value.value != nil {
		cost += (value.value.BitLen() + 7) / 8
	}
	if cost > staticEvaluator.bigIntCacheBudgetLeft {
		return
	}
	if staticEvaluator.bigIntExpressionCache == nil {
		staticEvaluator.bigIntExpressionCache = map[*ast.Node]staticBigIntCacheEntry{}
	}
	if _, exists := staticEvaluator.bigIntExpressionCache[node]; exists {
		return
	}
	staticEvaluator.bigIntCacheBudgetLeft -= cost
	staticEvaluator.bigIntExpressionCache[node] = staticBigIntCacheEntry{
		result: result,
		depth:  depth,
		steps:  steps,
	}
}

func isCacheableBigIntExpressionNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindBigIntLiteral:
		return true
	case ast.KindPrefixUnaryExpression:
		operator := node.AsPrefixUnaryExpression().Operator
		return operator == ast.KindMinusToken || operator == ast.KindTildeToken
	case ast.KindBinaryExpression:
		operator := node.AsBinaryExpression().OperatorToken.Kind
		switch operator {
		case ast.KindPlusToken, ast.KindMinusToken, ast.KindAsteriskToken,
			ast.KindSlashToken, ast.KindPercentToken, ast.KindAsteriskAsteriskToken,
			ast.KindBarToken, ast.KindAmpersandToken, ast.KindCaretToken,
			ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken:
			return true
		}
	}
	return false
}

// attachStaticResourceBudgets makes allocations and size-dependent operations
// hidden inside JavaScript primitive coercion share the enhanced evaluator's
// persistent file budgets. The legacy evaluator never calls this helper, so
// its evaluation surface and resource policy stay unchanged.
func (staticEvaluator *StaticStringEvaluator) attachStaticResourceBudgets(value any) any {
	if staticEvaluator == nil || !staticEvaluator.eslintStaticCalls {
		return value
	}
	work := staticEvaluator.stringWorkContext
	switch value := value.(type) {
	case *staticArrayValue:
		if value == nil || value.stringWorkContext == work {
			return value
		}
		value.stringWorkContext = work
		for index := range value.length {
			value.set(index, staticEvaluator.attachStaticResourceBudgets(value.element(index)))
		}
		return value
	case *staticObjectValue:
		if value == nil || value.stringWorkContext == work {
			return value
		}
		value.stringWorkContext = work
		if value.propertyCount > 0 {
			value.property.value = staticEvaluator.attachStaticResourceBudgets(value.property.value)
		}
		for index := range value.extraProperties {
			value.extraProperties[index].value = staticEvaluator.attachStaticResourceBudgets(
				value.extraProperties[index].value,
			)
		}
		value.prototype = staticEvaluator.attachStaticResourceBudgets(value.prototype)
		return value
	case staticCollectionValue:
		if value.stringWorkContext == work {
			return value
		}
		value.stringWorkContext = work
		for index := range value.entries {
			value.entries[index].key = staticEvaluator.attachStaticResourceBudgets(value.entries[index].key)
			value.entries[index].value = staticEvaluator.attachStaticResourceBudgets(value.entries[index].value)
		}
		return value
	case staticIteratorValue:
		if value.stringWorkContext == work {
			return value
		}
		value.stringWorkContext = work
		for index := range value.values {
			value.values[index] = staticEvaluator.attachStaticResourceBudgets(value.values[index])
		}
		return value
	case staticBoxedValue:
		if value.stringWorkContext == work {
			return value
		}
		value.stringWorkContext = work
		value.value = staticEvaluator.attachStaticResourceBudgets(value.value)
		return value
	case staticRegExpValue:
		value.stringWorkContext = work
		return value
	case staticBigIntValue:
		value.stringWorkContext = work
		value.bigIntWorkContext = staticEvaluator.bigIntWorkContext
		return value
	default:
		return value
	}
}

func (staticEvaluator *StaticStringEvaluator) chargeStaticStringResult(result *staticEvalResult) bool {
	if result == nil || !result.stringCodeUnitsKnown ||
		result.stringBudgetGeneration == staticEvaluator.evaluationGeneration {
		return true
	}
	if _, sourceBacked := result.value.(*staticStringNode); sourceBacked {
		// The literal itself is backed by source text, but every static native
		// call that consumes it can allocate a transformed copy. Charge it once
		// per root (rather than permanently) to bound repeated transformations.
		if result.stringCodeUnits > staticEvaluator.stringWorkBudgetLeft {
			staticEvaluator.stringWorkBudgetLeft = 0
			return false
		}
		staticEvaluator.stringWorkBudgetLeft -= result.stringCodeUnits
		result.stringBudgetGeneration = staticEvaluator.evaluationGeneration
		return true
	}
	if !result.stringWorkCharged {
		if result.stringCodeUnits > staticEvaluator.stringWorkBudgetLeft {
			// The value has already been materialized. Exhaust the file budget so
			// later roots stop before repeating the same allocation.
			staticEvaluator.stringWorkBudgetLeft = 0
			return false
		}
		staticEvaluator.stringWorkBudgetLeft -= result.stringCodeUnits
		result.stringWorkCharged = true
	}
	if result.stringCodeUnits > staticEvaluator.stringBudgetLeft {
		return false
	}
	staticEvaluator.stringBudgetLeft -= result.stringCodeUnits
	result.stringBudgetGeneration = staticEvaluator.evaluationGeneration
	return true
}

// evalValueUnbudgeted performs one evaluation step. Recursive evaluation goes
// through evalValue so aliases, calls, and nested literals share one root
// budget in the enhanced source-only evaluator.
func (staticEvaluator *StaticStringEvaluator) evalValueUnbudgeted(node *ast.Node) staticEvalResult {
	node = SkipAssertionsAndParens(node)
	if node == nil {
		return staticEvalResult{}
	}

	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return staticEvalResult{value: (*staticStringNode)(node), ok: true}
	case ast.KindTrueKeyword:
		return staticEvalResult{value: true, ok: true}
	case ast.KindFalseKeyword:
		return staticEvalResult{value: false, ok: true}
	case ast.KindNullKeyword:
		return staticEvalResult{value: staticNullValue{}, ok: true}
	case ast.KindBigIntLiteral:
		if !staticEvaluator.eslintStaticCalls {
			return staticEvaluator.evalWithTsgo(node)
		}
		value := new(big.Int)
		if _, ok := value.SetString(NormalizeBigIntLiteral(node.AsBigIntLiteral().Text), 10); !ok {
			return staticEvalResult{}
		}
		if value.BitLen() > maxStaticBigIntBits {
			return staticEvalResult{
				value: staticAbstractBigInt(value.Sign() != 0), ok: true,
				bigIntWorkBytes: (value.BitLen() + 7) / 8, cacheableBigInt: true,
			}
		}
		return staticEvalResult{
			value: staticBigIntValue{value: value}, ok: true,
			bigIntWorkBytes: (value.BitLen() + 7) / 8, cacheableBigInt: true,
		}
	case ast.KindRegularExpressionLiteral:
		if staticEvaluator.eslintStaticCalls {
			pattern, flags := ExtractRegexPatternAndFlags(node.AsRegularExpressionLiteral().Text)
			source, ok := staticRegExpSource(staticEvaluator.stringWorkContext, pattern)
			if !ok {
				return staticEvalResult{}
			}
			return staticEvalResult{value: staticRegExpValue{
				source: source, flags: canonicalRegExpFlags(flags), identity: node,
			}, ok: true}
		}
	case ast.KindUndefinedKeyword:
		if !staticEvaluator.resolveIdentifiers {
			return staticEvalResult{}
		}
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	case ast.KindIdentifier:
		if !staticEvaluator.resolveIdentifiers {
			return staticEvalResult{}
		}
		if staticEvaluator.eslintStaticCalls {
			if result := staticEvaluator.evalIdentifier(node); result.ok {
				return result
			}
			return staticEvaluator.evalESLintGlobalIdentifier(node)
		}
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
	case ast.KindTypeOfExpression:
		if staticEvaluator.eslintStaticCalls {
			return staticEvaluator.evalESLintTypeOfExpression(node)
		}
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
		if staticEvaluator.eslintStaticCalls {
			return staticEvaluator.evalESLintStaticCall(node)
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
	case ast.KindNewExpression:
		if staticEvaluator.eslintStaticCalls {
			return staticEvaluator.evalESLintStaticNew(node)
		}
	case ast.KindTaggedTemplateExpression:
		if staticEvaluator.eslintStaticCalls {
			return staticEvaluator.evalESLintStringRawTag(node)
		}
		if result := staticEvaluator.evalStringRawTag(node); result.ok {
			return result
		}
	}

	return staticEvaluator.evalWithTsgo(node)
}

func staticAggregateShallowCost(value any) int {
	switch value := value.(type) {
	case *staticArrayValue:
		return value.length
	case *staticObjectValue:
		return value.propertyCount
	case staticCollectionValue:
		if len(value.entries) > maxStaticAggregateElements/2 {
			return maxStaticAggregateElements + 1
		}
		return len(value.entries) * 2
	case staticIteratorValue:
		return len(value.values)
	case staticBoxedValue, staticRegExpValue, staticDateValue, staticOpaqueObjectValue:
		return 1
	default:
		return 0
	}
}

func staticAggregateIdentity(value any) any {
	switch value := value.(type) {
	case *staticArrayValue:
		return value
	case *staticObjectValue:
		return value
	case staticCollectionValue:
		return value.identity
	case staticIteratorValue:
		return value.identity
	case staticBoxedValue:
		return value.identity
	case staticRegExpValue:
		return value.identity
	case staticDateValue:
		return value.identity
	case staticOpaqueObjectValue:
		return value.identity
	default:
		return nil
	}
}

func (staticEvaluator *StaticStringEvaluator) evalIdentifier(node *ast.Node) (result staticEvalResult) {
	if !staticEvaluator.resolveIdentifiers {
		return staticEvalResult{}
	}
	initializer, symbol, ok := staticEvaluator.resolveIdentifierInitializer(node)
	if !ok {
		return staticEvalResult{}
	}
	if staticEvaluator.resolving[symbol] {
		return staticEvalResult{}
	}

	staticEvaluator.resolving[symbol] = true
	defer delete(staticEvaluator.resolving, symbol)
	if isAggregateLiteral(initializer) && staticEvaluator.hasPropertyMutation(symbol) {
		result = staticEvalResult{}
	} else {
		result = staticEvaluator.evalValue(initializer)
	}
	// A primitive result remains semantically reusable, but propagating this
	// marker through a binding would let a cached alias suffix shorten a later
	// deep evaluation. Only the closed initializer expression itself is cached.
	result.cacheableBigInt = false
	if result.ok && staticValueIsAggregate(result.value) &&
		!isAggregateLiteral(initializer) && staticEvaluator.hasPropertyMutation(symbol) {
		result = staticEvalResult{}
	}
	return result
}

func (staticEvaluator *StaticStringEvaluator) evalESLintGlobalIdentifier(node *ast.Node) staticEvalResult {
	if node == nil || !ast.IsIdentifier(node) {
		return staticEvalResult{}
	}
	name := node.AsIdentifier().Text
	if !staticEvaluator.isUnshadowedGlobal(node, name) {
		return staticEvalResult{}
	}
	switch name {
	case "undefined":
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	case "Infinity":
		return staticEvalResult{value: staticNumberValue(math.Inf(1)), ok: true}
	case "NaN":
		return staticEvalResult{value: staticNumberValue(math.NaN()), ok: true}
	case "Array", "ArrayBuffer", "BigInt", "BigInt64Array", "BigUint64Array",
		"Boolean", "DataView", "Date", "decodeURI", "decodeURIComponent",
		"encodeURI", "encodeURIComponent", "escape", "Float32Array", "Float64Array",
		"Function", "Int16Array", "Int32Array", "Int8Array", "isFinite", "isNaN",
		"isPrototypeOf", "JSON", "Map", "Math", "Number", "Object", "parseFloat",
		"parseInt", "Promise", "Proxy", "Reflect", "RegExp", "Set", "String",
		"Symbol", "Uint16Array", "Uint32Array", "Uint8Array", "Uint8ClampedArray",
		"unescape", "WeakMap", "WeakSet":
		return staticEvalResult{value: staticBuiltinValue(name), ok: true}
	}
	return staticEvalResult{}
}

func (staticEvaluator *StaticStringEvaluator) isUnshadowedGlobal(node *ast.Node, name string) bool {
	if resolver, ok := staticEvaluator.referenceResolver.(StaticGlobalReferenceResolver); ok {
		return resolver.IsUnshadowedGlobal(node, name)
	}
	return !IsShadowed(node, name)
}

func isAggregateLiteral(node *ast.Node) bool {
	node = SkipAssertionsAndParens(node)
	return node != nil && (node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindArrayLiteralExpression)
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

func (staticEvaluator *StaticStringEvaluator) evalTemplateExpression(node *ast.Node) staticEvalResult {
	template := node.AsTemplateExpression()
	if template == nil {
		return staticEvalResult{}
	}
	if !staticEvaluator.eslintStaticCalls {
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

	var builder strings.Builder
	codeUnits := 0
	appendText := func(text string, knownUnits int, unitsKnown bool) bool {
		if !unitsKnown {
			if len(text) > (maxStaticStringLength-codeUnits)*3 {
				staticEvaluator.stringWorkContext.exhaust()
				return false
			}
			if !staticEvaluator.stringWorkContext.reserve(max(len(text), 1)) {
				return false
			}
			knownUnits = ecmascript.StringCodeUnitCount(text)
		}
		if knownUnits > maxStaticStringLength-codeUnits {
			staticEvaluator.stringWorkContext.exhaust()
			return false
		}
		if !staticEvaluator.stringWorkContext.reserve(len(text) * 2) {
			return false
		}
		builder.WriteString(text)
		codeUnits += knownUnits
		return true
	}
	if template.Head != nil {
		if !appendText(template.Head.Text(), 0, false) {
			return staticEvalResult{}
		}
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
			if !appendText(value, result.stringCodeUnits, result.stringCodeUnitsKnown) {
				return staticEvalResult{}
			}
			if span.Literal != nil {
				if !appendText(span.Literal.Text(), 0, false) {
					return staticEvalResult{}
				}
			}
		}
	}
	return staticEvalResult{
		value: builder.String(), ok: true, stringCodeUnits: codeUnits, stringCodeUnitsKnown: true,
	}
}

func (staticEvaluator *StaticStringEvaluator) evalBinaryExpression(node *ast.Node) staticEvalResult {
	binary := node.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return staticEvalResult{}
	}

	switch binary.OperatorToken.Kind {
	case ast.KindCommaToken:
		if staticEvaluator.eslintStaticCalls {
			return staticEvaluator.evalValue(binary.Right)
		}
		return staticEvaluator.evalWithTsgo(node)
	case ast.KindEqualsToken:
		result := staticEvaluator.evalValue(binary.Right)
		if !staticEvaluator.eslintStaticCalls && result.ok && staticValueIsAggregate(result.value) {
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
		if !staticEvaluator.eslintStaticCalls &&
			(staticValueIsAggregate(left.value) || staticValueIsAggregate(right.value)) {
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
		if staticValueIsString(left.value) {
			return staticEvaluator.concatStaticValues(left, right)
		}
		if staticValueIsString(right.value) {
			return staticEvaluator.concatStaticValues(left, right)
		}
	}
	if staticEvaluator.eslintStaticCalls {
		return staticEvaluator.evalESLintBinaryExpression(binary)
	}

	return staticEvaluator.evalWithTsgo(node)
}

func (staticEvaluator *StaticStringEvaluator) concatStaticValues(
	left, right staticEvalResult,
) staticEvalResult {
	leftString, ok := staticValueToString(left.value)
	if !ok {
		return staticEvalResult{}
	}
	rightString, ok := staticValueToString(right.value)
	if !ok {
		return staticEvalResult{}
	}
	if !staticEvaluator.eslintStaticCalls {
		return staticEvalResult{value: leftString + rightString, ok: true}
	}
	leftUnits := left.stringCodeUnits
	if !left.stringCodeUnitsKnown {
		leftUnits = ecmascript.StringCodeUnitCount(leftString)
	}
	rightUnits := right.stringCodeUnits
	if !right.stringCodeUnitsKnown {
		rightUnits = ecmascript.StringCodeUnitCount(rightString)
	}
	if leftUnits > maxStaticStringLength || rightUnits > maxStaticStringLength-leftUnits {
		staticEvaluator.stringWorkContext.exhaust()
		return staticEvalResult{}
	}
	if len(leftString) > maxStaticStringLength*3 ||
		len(rightString) > maxStaticStringLength*3-len(leftString) {
		staticEvaluator.stringWorkContext.exhaust()
		return staticEvalResult{}
	}
	if !staticEvaluator.stringWorkContext.reserve((len(leftString) + len(rightString)) * 2) {
		return staticEvalResult{}
	}
	return staticEvalResult{
		value: leftString + rightString, ok: true,
		stringCodeUnits: leftUnits + rightUnits, stringCodeUnitsKnown: true,
	}
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
	if staticEvaluator.eslintStaticCalls {
		if bigint, ok := operand.value.(staticBigIntValue); ok && bigint.value != nil {
			switch prefix.Operator {
			case ast.KindMinusToken:
				value := new(big.Int).Neg(bigint.value)
				return staticEvalResult{
					value: staticBigIntValue{value: value}, ok: true,
					bigIntWorkBytes: (value.BitLen() + 7) / 8,
					cacheableBigInt: operand.cacheableBigInt,
				}
			case ast.KindTildeToken:
				value := new(big.Int).Not(bigint.value)
				return staticEvalResult{
					value: staticBigIntValue{value: value}, ok: true,
					bigIntWorkBytes: (value.BitLen() + 7) / 8,
					cacheableBigInt: operand.cacheableBigInt,
				}
			default:
				// Unary plus throws for BigInt.
				return staticEvalResult{}
			}
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
	if staticEvaluator.eslintStaticCalls && len(object.Properties.Nodes) > maxStaticAggregateElements {
		return staticEvalResult{}
	}
	if staticEvaluator.eslintStaticCalls &&
		!staticEvaluator.stringWorkContext.reserve(max(len(object.Properties.Nodes)*64, 1)) {
		return staticEvalResult{}
	}

	value := &staticObjectValue{enhancedCoercion: staticEvaluator.eslintStaticCalls}
	for _, property := range object.Properties.Nodes {
		var valueNode *ast.Node
		switch property.Kind {
		case ast.KindPropertyAssignment:
			valueNode = property.AsPropertyAssignment().Initializer
		case ast.KindShorthandPropertyAssignment:
			valueNode = property.Name()
		case ast.KindSpreadAssignment:
			if !staticEvaluator.eslintStaticCalls {
				return staticEvalResult{}
			}
			spread := staticEvaluator.evalValue(property.AsSpreadAssignment().Expression)
			if !spread.ok || !staticEvaluator.spreadObjectProperties(value, spread.value) {
				return staticEvalResult{}
			}
			if value.propertyCount > maxStaticAggregateElements {
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
			case *staticObjectValue, *staticArrayValue, staticBuiltinObjectValue:
				value.prototype = propertyValue.value
				value.prototypeSet = true
			case staticNullValue:
				value.prototype = nil
				value.prototypeSet = true
			default:
				if staticValueIsAggregate(propertyValue.value) {
					// eslint-utils does not safely invoke inherited getters on an
					// arbitrary native object used as an object-literal prototype.
					return staticEvalResult{}
				}
				// JavaScript ignores primitive prototype-setter values.
			}
			continue
		}

		propertyValue := staticEvaluator.evalValue(valueNode)
		if !propertyValue.ok {
			return staticEvalResult{}
		}
		if staticEvaluator.eslintStaticCalls {
			key, symbol, ok := staticEvaluator.evalESLintPropertyKey(property.Name())
			if !ok {
				return staticEvalResult{}
			}
			if symbol != nil {
				value.addSymbolProperty(*symbol, propertyValue.value)
			} else if key == "__proto__" {
				// eslint-utils builds object expressions via ordinary assignment,
				// so a computed "__proto__" key reaches the legacy setter too.
				switch propertyValue.value.(type) {
				case *staticObjectValue, *staticArrayValue, staticBuiltinObjectValue:
					value.prototype = propertyValue.value
					value.prototypeSet = true
				case staticNullValue:
					value.prototype = nil
					value.prototypeSet = true
				default:
					if staticValueIsAggregate(propertyValue.value) {
						return staticEvalResult{}
					}
				}
			} else {
				value.addProperty(key, propertyValue.value)
			}
			continue
		}
		key, ok := staticEvaluator.evalPropertyKey(property.Name())
		if !ok {
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
	if value.propertyIndices == nil && value.propertyCount >= 8 {
		value.propertyIndices = make(map[string]int, value.propertyCount)
		if !value.property.symbolKey {
			value.propertyIndices[value.property.name] = 0
		}
		for index, existing := range value.extraProperties {
			if !existing.symbolKey {
				value.propertyIndices[existing.name] = index + 1
			}
		}
	} else if value.propertyIndices != nil {
		value.propertyIndices[name] = value.propertyCount - 1
	}
}

func (value *staticObjectValue) addSymbolProperty(symbol staticSymbolValue, propertyValue any) {
	property := staticObjectProperty{symbol: symbol, symbolKey: true, value: propertyValue}
	if value.propertyCount == 0 {
		value.property = property
	} else {
		value.extraProperties = append(value.extraProperties, property)
	}
	value.propertyCount++
}

func (staticEvaluator *StaticStringEvaluator) spreadObjectProperties(target *staticObjectValue, source any) bool {
	// Object spread ignores null and undefined instead of applying the
	// throwing Object.keys conversion used by Object.entries/keys/values.
	if staticValueNullish(source) {
		return true
	}
	properties, ok := staticEnumerableOwnProperties(staticEvaluator.stringWorkContext, source)
	if !ok {
		return false
	}
	for _, property := range properties {
		if target.propertyCount >= maxStaticAggregateElements {
			return false
		}
		target.addProperty(property.name, property.value)
	}
	if object, ok := source.(*staticObjectValue); ok {
		if object.propertyCount > 0 && object.property.symbolKey {
			if target.propertyCount >= maxStaticAggregateElements {
				return false
			}
			target.addSymbolProperty(object.property.symbol, object.property.value)
		}
		for _, property := range object.extraProperties {
			if property.symbolKey {
				if target.propertyCount >= maxStaticAggregateElements {
					return false
				}
				target.addSymbolProperty(property.symbol, property.value)
			}
		}
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
				if !prototypeValue.ok {
					return staticEvalResult{}
				}
				prototypeSet = true
				continue
			}

			value := staticEvaluator.evalValue(valueNode)
			if !value.ok {
				return staticEvalResult{}
			}
			switch value.value.(type) {
			case *staticObjectValue, *staticArrayValue:
				prototypeValue = staticMemberValue(value.value, key)
				if !prototypeValue.ok {
					return staticEvalResult{}
				}
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
		if propertyKey == key {
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
	return staticEvalResult{value: staticUndefinedValue{}, ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evalArrayLiteral(node *ast.Node) staticEvalResult {
	array := node.AsArrayLiteralExpression()
	if array == nil || array.Elements == nil {
		return staticEvalResult{}
	}
	if staticEvaluator.eslintStaticCalls {
		if len(array.Elements.Nodes) > maxStaticAggregateElements {
			return staticEvalResult{}
		}
		if !staticEvaluator.stringWorkContext.reserve(max(len(array.Elements.Nodes)*32, 1)) {
			return staticEvalResult{}
		}
		elements := make([]staticArrayElement, 0, len(array.Elements.Nodes))
		for _, element := range array.Elements.Nodes {
			switch element.Kind {
			case ast.KindOmittedExpression:
				if len(elements) >= maxStaticAggregateElements {
					return staticEvalResult{}
				}
				elements = append(elements, staticArrayElement{value: staticUndefinedValue{}, omitted: true})
			case ast.KindSpreadElement:
				spread := staticEvaluator.evalValue(element.AsSpreadElement().Expression)
				if !spread.ok {
					return staticEvalResult{}
				}
				values, ok := staticIterableValues(
					staticEvaluator.stringWorkContext,
					spread.value,
					maxStaticAggregateElements-len(elements),
				)
				if !ok {
					return staticEvalResult{}
				}
				for _, value := range values {
					elements = append(elements, staticArrayElement{value: value})
				}
			default:
				if len(elements) >= maxStaticAggregateElements {
					return staticEvalResult{}
				}
				result := staticEvaluator.evalValue(element)
				if !result.ok {
					return staticEvalResult{}
				}
				elements = append(elements, staticArrayElement{value: result.value})
			}
		}
		return staticEvalResult{value: staticArrayFromElements(elements), ok: true}
	}

	arrayValue := &staticArrayValue{length: len(array.Elements.Nodes)}
	if arrayValue.length > len(arrayValue.inline) {
		arrayValue.overflow = make([]any, arrayValue.length-len(arrayValue.inline))
	}
	for i, element := range array.Elements.Nodes {
		if element.Kind == ast.KindOmittedExpression {
			arrayValue.set(i, staticUndefinedValue{})
			if staticEvaluator.eslintStaticCalls {
				if arrayValue.omitted == nil {
					arrayValue.omitted = make(map[int]bool, 1)
				}
				arrayValue.omitted[i] = true
			}
			continue
		}
		if element.Kind == ast.KindSpreadElement {
			return staticEvalResult{}
		}
		elementValue := staticEvaluator.evalValue(element)
		if !elementValue.ok {
			return staticEvalResult{}
		}
		arrayValue.set(i, elementValue.value)
	}
	return staticEvalResult{value: arrayValue, ok: true}
}

func (value *staticArrayValue) set(index int, element any) {
	if index < len(value.inline) {
		value.inline[index] = element
		return
	}
	value.overflow[index-len(value.inline)] = element
}

func (value *staticArrayValue) element(index int) any {
	if index < len(value.inline) {
		return value.inline[index]
	}
	return value.overflow[index-len(value.inline)]
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

func (staticEvaluator *StaticStringEvaluator) evalESLintPropertyKey(name *ast.Node) (string, *staticSymbolValue, bool) {
	if name == nil {
		return "", nil, false
	}
	if name.Kind == ast.KindComputedPropertyName {
		computed := name.AsComputedPropertyName()
		if computed == nil {
			return "", nil, false
		}
		result := staticEvaluator.evalValue(computed.Expression)
		if !result.ok {
			return "", nil, false
		}
		if symbol, ok := result.value.(staticSymbolValue); ok {
			if symbol.hostDependent {
				return "", nil, false
			}
			return "", &symbol, true
		}
		key, ok := staticValueToString(result.value)
		return key, nil, ok
	}
	key, ok := staticEvaluator.evalPropertyKey(name)
	return key, nil, ok
}

func (staticEvaluator *StaticStringEvaluator) evalMemberAccess(node *ast.Node) staticEvalResult {
	objectNode := AccessExpressionObject(node)
	if objectNode == nil {
		return staticEvalResult{}
	}
	if staticEvaluator.eslintStaticCalls {
		// eslint-utils resolves an optional receiver before a computed key.
		// This matters for `null?.[unknown]`, which short-circuits statically.
		object := staticEvaluator.evalValue(objectNode)
		if !object.ok {
			return staticEvalResult{}
		}
		if _, shortCircuited := object.value.(staticOptionalChainShortCircuitValue); shortCircuited {
			if ast.IsOptionalChain(node) && !ast.IsOuterExpression(objectNode, ast.OEKParentheses) {
				return object
			}
			return staticEvalResult{}
		}
		if staticValueNullish(object.value) {
			if node.QuestionDotToken() != nil {
				return staticEvalResult{value: staticOptionalChainShortCircuitValue{}, ok: true}
			}
			return staticEvalResult{}
		}
		key, symbolKey, ok := staticEvaluator.evalESLintAccessExpressionKey(node)
		if !ok {
			return staticEvalResult{}
		}
		if symbolKey != nil {
			return staticESLintSymbolMemberValue(object.value, *symbolKey)
		}
		if result := staticESLintMemberValue(object.value, key); result.ok {
			return result
		}
		return staticESLintBuiltinMember(object.value, key)
	}

	key, ok := staticEvaluator.evalAccessExpressionKey(node)
	if !ok {
		return staticEvalResult{}
	}
	if literal := SkipAssertionsAndParens(objectNode); literal != nil && literal.Kind == ast.KindObjectLiteralExpression {
		return staticEvaluator.evalObjectLiteralMember(literal, key)
	}
	object := staticEvaluator.evalValue(objectNode)
	if !object.ok {
		return staticEvalResult{}
	}
	return staticMemberValue(object.value, key)
}

func staticESLintBuiltinMember(object any, key string) staticEvalResult {
	if key == "__proto__" || key == "caller" || key == "arguments" {
		return staticEvalResult{}
	}
	if prototype, ok := object.(staticBuiltinObjectValue); ok {
		return staticESLintBuiltinPrototypeMember(string(prototype), key)
	}
	builtin, ok := object.(staticBuiltinValue)
	if !ok {
		return staticEvalResult{}
	}
	name := string(builtin)
	if name == "RegExp" && staticRegExpLegacyGetter(key) {
		return staticEvalResult{}
	}
	switch name {
	case "Math":
		switch key {
		case "E":
			return staticEvalResult{value: staticNumberValue(math.E), ok: true}
		case "LN10":
			return staticEvalResult{value: staticNumberValue(math.Ln10), ok: true}
		case "LN2":
			return staticEvalResult{value: staticNumberValue(math.Ln2), ok: true}
		case "LOG10E":
			return staticEvalResult{value: staticNumberValue(math.Log10E), ok: true}
		case "LOG2E":
			return staticEvalResult{value: staticNumberValue(math.Log2E), ok: true}
		case "PI":
			return staticEvalResult{value: staticNumberValue(math.Pi), ok: true}
		case "SQRT1_2":
			return staticEvalResult{value: staticNumberValue(math.Sqrt2 / 2), ok: true}
		case "SQRT2":
			return staticEvalResult{value: staticNumberValue(math.Sqrt2), ok: true}
		}
	case "Number":
		switch key {
		case "EPSILON":
			return staticEvalResult{value: staticNumberValue(math.Nextafter(1, 2) - 1), ok: true}
		case "MAX_SAFE_INTEGER":
			return staticEvalResult{value: staticNumberValue(9007199254740991), ok: true}
		case "MAX_VALUE":
			return staticEvalResult{value: staticNumberValue(math.MaxFloat64), ok: true}
		case "MIN_SAFE_INTEGER":
			return staticEvalResult{value: staticNumberValue(-9007199254740991), ok: true}
		case "MIN_VALUE":
			return staticEvalResult{value: staticNumberValue(math.SmallestNonzeroFloat64), ok: true}
		case "NaN":
			return staticEvalResult{value: staticNumberValue(math.NaN()), ok: true}
		case "NEGATIVE_INFINITY":
			return staticEvalResult{value: staticNumberValue(math.Inf(-1)), ok: true}
		case "POSITIVE_INFINITY":
			return staticEvalResult{value: staticNumberValue(math.Inf(1)), ok: true}
		}
	case "Symbol":
		if staticSymbolProperty(key) {
			if key == "dispose" || key == "asyncDispose" {
				// Node 22 exposes these as Symbol.for("nodejs.*"); Node 23+
				// exposes distinct well-known symbols. Their type and identity
				// relative to themselves are stable, but registry membership and
				// observable descriptions are host-dependent.
				return staticEvalResult{value: staticSymbolValue{description: key, hostDependent: true}, ok: true}
			}
			return staticEvalResult{value: staticSymbolValue{description: key, wellKnown: true}, ok: true}
		}
	}
	if bytes, ok := staticTypedArrayBytesPerElement(name); ok && key == "BYTES_PER_ELEMENT" {
		return staticEvalResult{value: staticNumberValue(bytes), ok: true}
	}
	if staticBuiltinFunctionProperty(name, key) {
		return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
	}
	if key == "prototype" && staticBuiltinHasPrototype(name) {
		if name == "Function" {
			return staticEvalResult{value: staticBuiltinValue("Function.prototype"), ok: true}
		}
		return staticEvalResult{value: staticBuiltinObjectValue(name + ".prototype"), ok: true}
	}
	if staticBuiltinIsFunction(name) {
		switch key {
		case "apply", "bind", "call":
			return staticEvalResult{value: staticBuiltinValue("Function.prototype." + key), ok: true}
		case "constructor":
			return staticEvalResult{value: staticBuiltinValue("Function"), ok: true}
		case "toString":
			return staticEvalResult{value: staticBuiltinValue("Function.prototype.toString"), ok: true}
		}
	}
	if key == "length" || key == "name" {
		if strings.Contains(name, ".") || staticBuiltinIsFunction(name) {
			if key == "name" {
				canonicalName := staticCanonicalBuiltinFunctionName(name)
				if canonicalName == "Function.prototype" {
					return staticEvalResult{value: "", ok: true}
				}
				return staticEvalResult{
					value: canonicalName[strings.LastIndex(canonicalName, ".")+1:], ok: true,
				}
			}
			if length, ok := staticBuiltinFunctionLength(name); ok {
				return staticEvalResult{value: staticNumberValue(length), ok: true}
			}
			return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
		}
	}
	return staticObjectPrototypeMember(key)
}

func staticBuiltinFunctionLength(name string) (int, bool) {
	if strings.HasPrefix(name, "Math.") {
		switch strings.TrimPrefix(name, "Math.") {
		case "random":
			return 0, true
		case "atan2", "hypot", "imul", "max", "min", "pow":
			return 2, true
		default:
			if staticMathMethod(strings.TrimPrefix(name, "Math.")) {
				return 1, true
			}
		}
	}
	switch name {
	case "Function.prototype":
		return 0, true
	case "parseInt", "RegExp":
		return 2, true
	case "Array", "ArrayBuffer", "BigInt", "Boolean", "DataView", "decodeURI",
		"decodeURIComponent", "encodeURI", "encodeURIComponent", "escape", "Function",
		"isFinite", "isNaN", "isPrototypeOf", "Number", "Object", "parseFloat", "Promise",
		"String", "unescape":
		return 1, true
	case "Map", "Set", "Symbol", "WeakMap", "WeakSet":
		return 0, true
	case "Proxy":
		return 2, true
	case "Date":
		return 7, true
	}
	if _, typedArray := staticTypedArrayBytesPerElement(name); typedArray {
		return 3, true
	}
	return 0, false
}

func staticRegExpLegacyGetter(key string) bool {
	switch key {
	case "input", "$_", "lastMatch", "$&", "lastParen", "$+", "leftContext", "$`",
		"rightContext", "$'", "$1", "$2", "$3", "$4", "$5", "$6", "$7", "$8", "$9":
		return true
	}
	return false
}

func staticESLintBuiltinPrototypeMember(name, key string) staticEvalResult {
	if name == "Array.prototype[Symbol.unscopables]" {
		for _, property := range staticArrayUnscopablesPropertyNames {
			if key == property {
				return staticEvalResult{value: true, ok: true}
			}
		}
		// This special object has a null prototype, so even Object.prototype
		// names are absent.
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	}
	if staticBuiltinPrototypeGetterDenied(name, key) {
		return staticEvalResult{}
	}
	if key == "length" && (name == "Array.prototype" || name == "String.prototype") {
		return staticEvalResult{value: staticNumberValue(0), ok: true}
	}
	constructor := strings.TrimSuffix(name, ".prototype")
	if bytes, ok := staticTypedArrayBytesPerElement(constructor); ok && key == "BYTES_PER_ELEMENT" {
		return staticEvalResult{value: staticNumberValue(bytes), ok: true}
	}
	if key == "constructor" {
		return staticEvalResult{value: staticBuiltinValue(constructor), ok: true}
	}
	if key == "__proto__" {
		return staticEvalResult{}
	}
	switch name {
	case "Array.prototype":
		if staticArrayPrototypeMethod(key) {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	case "Map.prototype", "Set.prototype":
		if key == "size" {
			return staticEvalResult{}
		}
		if staticCollectionPrototypeMethod(constructor, key) {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	case "String.prototype":
		if staticStringPrototypeMethod(key) {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	case "Number.prototype":
		if staticNumberPrototypeMethod(key) {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	case "Boolean.prototype":
		if key == "toString" || key == "valueOf" {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	case "BigInt.prototype":
		if key == "toLocaleString" || key == "toString" || key == "valueOf" {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	case "Date.prototype":
		if staticDatePrototypeMethod(key) {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	case "RegExp.prototype":
		if key == "source" || key == "flags" {
			return staticEvalResult{}
		}
		if key == "compile" || key == "exec" || key == "test" || key == "toString" {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	case "Promise.prototype":
		if key == "catch" || key == "finally" || key == "then" {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	case "Symbol.prototype":
		if key == "description" {
			// Accessing the getter on Symbol.prototype itself throws.
			return staticEvalResult{}
		}
		if key == "toString" || key == "valueOf" {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	case "ArrayBuffer.prototype":
		if key == "resize" || key == "slice" || key == "transfer" || key == "transferToFixedLength" {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	case "DataView.prototype":
		if staticDataViewPrototypeMethod(key) {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	case "WeakMap.prototype":
		if key == "delete" || key == "get" || key == "has" || key == "set" {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	case "WeakSet.prototype":
		if key == "add" || key == "delete" || key == "has" {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	default:
		if staticTypedArrayPrototypeMethod(constructor, key) {
			return staticEvalResult{value: staticBuiltinValue(name + "." + key), ok: true}
		}
	}
	return staticObjectPrototypeMember(key)
}

func staticBuiltinPrototypeGetterDenied(name, key string) bool {
	switch name {
	case "ArrayBuffer.prototype":
		switch key {
		case "byteLength", "detached", "maxByteLength", "resizable":
			return true
		}
	case "DataView.prototype":
		switch key {
		case "buffer", "byteLength", "byteOffset":
			return true
		}
	case "Map.prototype", "Set.prototype":
		return key == "size"
	case "RegExp.prototype":
		switch key {
		case "dotAll", "flags", "global", "hasIndices", "ignoreCase", "multiline",
			"source", "sticky", "unicode", "unicodeSets":
			return true
		}
	case "Symbol.prototype":
		return key == "description"
	}
	if _, typedArray := staticTypedArrayBytesPerElement(strings.TrimSuffix(name, ".prototype")); typedArray {
		switch key {
		case "buffer", "byteLength", "byteOffset", "length":
			return true
		}
	}
	return false
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

func (staticEvaluator *StaticStringEvaluator) evalESLintAccessExpressionKey(node *ast.Node) (string, *staticSymbolValue, bool) {
	if key, ok := AccessExpressionStaticName(node); ok {
		return key, nil, true
	}
	if node == nil || node.Kind != ast.KindElementAccessExpression {
		return "", nil, false
	}
	argument := staticEvaluator.evalValue(node.AsElementAccessExpression().ArgumentExpression)
	if !argument.ok {
		return "", nil, false
	}
	if symbol, ok := argument.value.(staticSymbolValue); ok {
		if symbol.hostDependent {
			return "", nil, false
		}
		return "", &symbol, true
	}
	key, ok := staticValueToString(argument.value)
	return key, nil, ok
}

// staticMemberValue reads key off a folded object or array literal. Reading a
// key that isn't there yields `undefined`, as it would at runtime. A key an
// array inherits — `length`, `join` — stays unresolved, because folding it
// would need a numeric or callable value representation this evaluator doesn't
// carry.
func staticMemberValue(object any, key string) staticEvalResult {
	switch object := object.(type) {
	case *staticObjectValue:
		if value, ok := staticObjectOwnProperty(object, key); ok {
			return staticEvalResult{value: value, ok: true}
		}
		if object.prototypeSet && object.prototype != nil {
			return staticMemberValue(object.prototype, key)
		}
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	case *staticArrayValue:
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

func staticObjectOwnProperty(object *staticObjectValue, key string) (any, bool) {
	if object.propertyIndices != nil {
		index, ok := object.propertyIndices[key]
		if !ok {
			return nil, false
		}
		if index == 0 {
			return object.property.value, true
		}
		return object.extraProperties[index-1].value, true
	}
	for i := len(object.extraProperties) - 1; i >= 0; i-- {
		if !object.extraProperties[i].symbolKey && object.extraProperties[i].name == key {
			return object.extraProperties[i].value, true
		}
	}
	if object.propertyCount > 0 && !object.property.symbolKey && object.property.name == key {
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

	callee := ast.SkipOuterExpressions(call.Expression, ast.OEKParentheses|ast.OEKAssertions)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return nil, false
	}
	methodName, ok := staticEvaluator.evalAccessExpressionKey(callee)
	if !ok || !objectPassThroughMethods[methodName] {
		return nil, false
	}

	object := SkipAssertionsAndParens(AccessExpressionObject(callee))
	if !isIdentifierWithText(object, "Object") || IsShadowed(object, "Object") {
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
			array, arrayOK := staticEvaluator.evalInlineArrayLiteral(literal)
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

func (staticEvaluator *StaticStringEvaluator) evalInlineArrayLiteral(node *ast.Node) (staticArrayValue, bool) {
	literal := node.AsArrayLiteralExpression()
	if literal == nil || literal.Elements == nil || len(literal.Elements.Nodes) > len((staticArrayValue{}).inline) {
		return staticArrayValue{}, false
	}

	value := staticArrayValue{length: len(literal.Elements.Nodes)}
	for i, element := range literal.Elements.Nodes {
		if element.Kind == ast.KindOmittedExpression {
			value.set(i, staticUndefinedValue{})
			continue
		}
		if element.Kind == ast.KindSpreadElement {
			return staticArrayValue{}, false
		}
		elementValue := staticEvaluator.evalValue(element)
		if !elementValue.ok {
			return staticArrayValue{}, false
		}
		value.set(i, elementValue.value)
	}
	return value, true
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
	case "toUpperCase":
		return staticEvalResult{value: ecmascript.StringToUpperCase(text), ok: true}
	case "toLowerCase":
		return staticEvalResult{value: ecmascript.StringToLowerCase(text), ok: true}
	case "slice":
		return staticStringSlice(text, arguments)
	case "substring":
		return staticStringSubstring(text, arguments)
	case "substr":
		return staticStringSubstr(text, arguments)
	case "charAt":
		return staticStringCharAt(text, arguments)
	case "concat":
		return staticStringConcat(nil, text, arguments)
	}

	return staticEvalResult{}
}

func (staticEvaluator *StaticStringEvaluator) evalCallArguments(node *ast.Node) ([]any, bool) {
	if staticEvaluator.eslintStaticCalls && len(node.Arguments()) > maxStaticAggregateElements {
		return nil, false
	}
	if staticEvaluator.eslintStaticCalls &&
		!staticEvaluator.stringWorkContext.reserve(max(len(node.Arguments())*16, 1)) {
		return nil, false
	}
	values := make([]any, 0, len(node.Arguments()))
	for _, argumentNode := range node.Arguments() {
		if ast.IsSpreadElement(argumentNode) {
			spread := staticEvaluator.evalValue(argumentNode.AsSpreadElement().Expression)
			if !spread.ok {
				return nil, false
			}
			var spreadValues []any
			var ok bool
			if staticEvaluator.eslintStaticCalls {
				spreadValues, ok = staticIterableValues(
					staticEvaluator.stringWorkContext,
					spread.value,
					maxStaticAggregateElements-len(values),
				)
			} else if array, isArray := spread.value.(*staticArrayValue); isArray {
				spreadValues = make([]any, array.length)
				for index := range array.length {
					spreadValues[index] = array.element(index)
				}
				ok = true
			}
			if !ok {
				return nil, false
			}
			values = append(values, spreadValues...)
			continue
		}
		if staticEvaluator.eslintStaticCalls && len(values) >= maxStaticAggregateElements {
			return nil, false
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
	return staticEvalResult{value: string(utf16.Decode(units)), ok: true}
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
	units := utf16.Encode([]rune(text))
	length := len(units)
	from := normalizeSliceIndex(start, length)
	to := normalizeSliceIndex(end, length)
	if to < from {
		to = from
	}
	return staticEvalResult{value: string(utf16.Decode(units[from:to])), ok: true}
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
	units := utf16.Encode([]rune(text))
	length := len(units)
	from := clampSubstringIndex(start, length)
	to := clampSubstringIndex(end, length)
	if from > to {
		from, to = to, from
	}
	return staticEvalResult{value: string(utf16.Decode(units[from:to])), ok: true}
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
	units := utf16.Encode([]rune(text))
	length := len(units)
	from := normalizeSliceIndex(start, length)
	if count <= 0 {
		return staticEvalResult{value: "", ok: true}
	}
	to := length
	if count < length-from {
		to = from + count
	}
	return staticEvalResult{value: string(utf16.Decode(units[from:to])), ok: true}
}

func staticStringCharAt(text string, arguments []any) staticEvalResult {
	index, ok := staticArgumentInteger(arguments, 0, 0)
	if !ok {
		return staticEvalResult{}
	}
	units := utf16.Encode([]rune(text))
	if index < 0 || index >= len(units) {
		return staticEvalResult{value: "", ok: true}
	}
	return staticEvalResult{value: string(utf16.Decode(units[index : index+1])), ok: true}
}

func staticStringConcat(
	work *staticStringWorkContext,
	text string,
	arguments []any,
) staticEvalResult {
	var builder strings.Builder
	if !work.reserve(len(text) * 2) {
		return staticEvalResult{}
	}
	builder.WriteString(text)
	for _, argument := range arguments {
		part, ok := staticValueToString(argument)
		if !ok {
			return staticEvalResult{}
		}
		if len(part) > maxStaticStringLength-builder.Len() {
			work.exhaust()
			return staticEvalResult{}
		}
		if !work.reserve(len(part) * 2) {
			return staticEvalResult{}
		}
		builder.WriteString(part)
	}
	return staticEvalResult{value: builder.String(), ok: true}
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
	if math.IsNaN(number) || math.IsInf(number, 0) || number == 0 {
		return 0
	}
	remainder := math.Mod(math.Trunc(number), 1<<32)
	if remainder < 0 {
		remainder += 1 << 32
	}
	return uint32(remainder)
}

func toInt32(number float64) int32 {
	return int32(toUint32(number))
}

func staticValueToNumber(value any) (float64, bool) {
	switch value := value.(type) {
	case staticNumberValue:
		return float64(value), true
	case staticUnknownNumberValue, staticBigIntValue, staticSymbolValue:
		return 0, false
	case staticUnknownStringValue:
		if value.numericNaN {
			return math.NaN(), true
		}
		return 0, false
	case staticDateValue:
		if value.known {
			return float64(value.milliseconds), true
		}
		return 0, false
	case staticBoxedValue:
		return staticValueToNumber(value.value)
	case staticBuiltinValue:
		// Native function source text is engine-specific, but neither it nor an
		// ordinary built-in object's default string is a StringNumericLiteral.
		return math.NaN(), true
	case staticBuiltinObjectValue:
		if _, typedArray := staticTypedArrayBytesPerElement(
			strings.TrimSuffix(string(value), ".prototype"),
		); typedArray {
			// TypedArray prototypes have no backing typed-array data. Their
			// inherited primitive conversion throws instead of producing NaN.
			return 0, false
		}
		switch value {
		case "Boolean.prototype", "Number.prototype", "String.prototype", "Array.prototype":
			return 0, true
		case "RegExp.prototype":
			return math.NaN(), true
		case "BigInt.prototype", "Date.prototype", "Symbol.prototype", "Array.prototype[Symbol.unscopables]":
			return 0, false
		default:
			return math.NaN(), true
		}
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
		return staticEvalResult{value: tagged.Template.Text(), ok: true}
	case ast.KindTemplateExpression:
		return staticEvaluator.evalTemplateExpression(tagged.Template)
	default:
		return staticEvalResult{}
	}
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

func (staticEvaluator *StaticStringEvaluator) isBuiltinStringValue(node *ast.Node, resolvingAliases map[*ast.Symbol]bool) bool {
	if !staticEvaluator.resolveIdentifiers {
		return false
	}
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

func (staticEvaluator *StaticStringEvaluator) isBuiltinArrayValue(node *ast.Node, resolvingAliases map[*ast.Symbol]bool) bool {
	node = SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if isIdentifierWithText(node, "Array") && !IsShadowed(node, "Array") {
		return true
	}
	initializer, symbol, ok := staticEvaluator.resolveIdentifierInitializer(node)
	if !ok || resolvingAliases[symbol] {
		return false
	}
	resolvingAliases[symbol] = true
	defer delete(resolvingAliases, symbol)
	return staticEvaluator.isBuiltinArrayValue(initializer, resolvingAliases)
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
	return staticEvaluator.referenceFlagsFor(symbol)&staticReferencePropertyMutation != 0
}

func (staticEvaluator *StaticStringEvaluator) referenceFlagsFor(symbol *ast.Symbol) staticReferenceFlags {
	if !staticEvaluator.referenceFlagsComputed {
		staticEvaluator.computeReferenceFlags()
	}
	return staticEvaluator.referenceFlags[symbol]
}

type staticMutationCandidate struct {
	reference *ast.Node
	access    *ast.Node
}

func (staticEvaluator *StaticStringEvaluator) computeReferenceFlags() {
	staticEvaluator.referenceFlagsComputed = true
	staticEvaluator.referenceFlags = nil
	var mutationCandidates []staticMutationCandidate

	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindIdentifier && !IsNonReferenceIdentifier(node) {
			if IsWriteReference(node) {
				staticEvaluator.markReference(node, staticReferenceWrite)
			}
			parent := node.Parent
			mayStartAccess := parent != nil && (ast.IsAccessExpression(parent) ||
				ast.IsOuterExpression(parent, skipTransparentKinds))
			if mayStartAccess {
				access, outer := outermostAccessFromReference(node)
				if access != nil && outer.Parent != nil {
					if ast.IsAssignmentTarget(outer) {
						staticEvaluator.markReference(node, staticReferencePropertyMutation)
					} else if outer.Parent.Kind == ast.KindCallExpression {
						call := outer.Parent.AsCallExpression()
						if call.Expression == outer {
							if methodName, ok := AccessExpressionStaticName(access); ok {
								if isMutatingArrayMethod(methodName) {
									staticEvaluator.markReference(node, staticReferencePropertyMutation)
								}
							} else {
								mutationCandidates = append(mutationCandidates, staticMutationCandidate{reference: node, access: access})
							}
						}
					}
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(&staticEvaluator.sourceFile.Node)

	for _, candidate := range mutationCandidates {
		if methodName, ok := staticEvaluator.evalAccessExpressionKey(candidate.access); ok && isMutatingArrayMethod(methodName) {
			staticEvaluator.markReference(candidate.reference, staticReferencePropertyMutation)
		}
	}
}

func (staticEvaluator *StaticStringEvaluator) markReference(node *ast.Node, flag staticReferenceFlags) {
	if symbol := staticEvaluator.referenceSymbol(node); symbol != nil {
		if staticEvaluator.referenceFlags == nil {
			staticEvaluator.referenceFlags = make(map[*ast.Symbol]staticReferenceFlags, 8)
		}
		staticEvaluator.referenceFlags[symbol] |= flag
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

func outermostAccessFromReference(node *ast.Node) (access *ast.Node, outer *ast.Node) {
	outer = node
	for outer.Parent != nil {
		parent := outer.Parent
		if ast.IsOuterExpression(parent, skipTransparentKinds) && parent.Expression() == outer {
			outer = parent
			continue
		}
		if ast.IsAccessExpression(parent) && AccessExpressionObject(parent) == outer {
			access = parent
			outer = parent
			continue
		}
		break
	}
	return access, outer
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
	case bool, staticNullValue, staticUndefinedValue, staticOptionalChainShortCircuitValue, staticNumberValue,
		staticUnknownNumberValue, staticUnknownStringValue, staticUnknownBooleanValue,
		staticOpaqueObjectValue, staticRegExpValue, staticDateValue, staticBoxedValue,
		staticCollectionValue, staticBigIntValue, staticSymbolValue, staticBuiltinValue, staticBuiltinObjectValue,
		*staticStringNode, *staticObjectValue, *staticArrayValue:
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
	case staticNullValue, staticUndefinedValue, staticOptionalChainShortCircuitValue:
		return true
	default:
		return false
	}
}

func staticValueUndefined(value any) bool {
	switch value.(type) {
	case staticUndefinedValue, staticOptionalChainShortCircuitValue:
		return true
	default:
		return false
	}
}

// staticValueKind is the JavaScript typeof-like classification `===` needs.
// Objects and arrays compare by identity, which folding does not model, and
// tsgo's bigint representation cannot be inspected from this package; both stay
// unknown.
type staticValueKind uint8

const (
	staticKindUnknown staticValueKind = iota
	staticKindString
	staticKindNumber
	staticKindBoolean
	staticKindBigInt
	staticKindSymbol
	staticKindNull
	staticKindUndefined
)

func staticValueKindOf(value any) staticValueKind {
	switch value.(type) {
	case bool, staticUnknownBooleanValue:
		return staticKindBoolean
	case staticBigIntValue:
		return staticKindBigInt
	case staticSymbolValue:
		return staticKindSymbol
	case staticNullValue:
		return staticKindNull
	case staticUndefinedValue:
		return staticKindUndefined
	case staticOptionalChainShortCircuitValue:
		return staticKindUndefined
	}
	if staticValueIsString(value) {
		return staticKindString
	}
	if _, ok := value.(staticUnknownStringValue); ok {
		return staticKindString
	}
	if _, ok := value.(staticUnknownNumberValue); ok {
		return staticKindNumber
	}
	if _, isNumber := value.(interface{ IsNaN() bool }); isNumber {
		return staticKindNumber
	}
	return staticKindUnknown
}

// staticValuesStrictEqual implements the strict equality comparison over folded
// values. Comparing across kinds is always false, which keeps `1n === 1` right
// even though bigint values themselves stay unknown.
func staticValuesStrictEqual(left any, right any) (equal bool, ok bool) {
	if leftBuiltin, ok := left.(staticBuiltinValue); ok {
		rightBuiltin, rightOK := right.(staticBuiltinValue)
		return rightOK && staticCanonicalBuiltinFunctionName(string(leftBuiltin)) ==
			staticCanonicalBuiltinFunctionName(string(rightBuiltin)), true
	}
	if leftBuiltin, ok := left.(staticBuiltinObjectValue); ok {
		rightBuiltin, rightOK := right.(staticBuiltinObjectValue)
		return rightOK && leftBuiltin == rightBuiltin, true
	}
	leftAggregate := staticValueIsAggregate(left)
	rightAggregate := staticValueIsAggregate(right)
	if leftAggregate != rightAggregate {
		return false, true
	}
	if leftAggregate {
		leftKey, leftOK := staticCanonicalCollectionKey(left, nil)
		rightKey, rightOK := staticCanonicalCollectionKey(right, nil)
		if leftOK && rightOK && leftKey.kind >= 11 && rightKey.kind >= 11 {
			return leftKey == rightKey, true
		}
		// Object/array initializers are reevaluated for each identifier
		// reference, so separately materialized aggregates are never identical.
		return false, true
	}
	leftKind := staticValueKindOf(left)
	rightKind := staticValueKindOf(right)
	if leftKind == staticKindUnknown || rightKind == staticKindUnknown {
		return false, false
	}
	if leftKind != rightKind {
		return false, true
	}

	switch leftKind {
	case staticKindString:
		leftText, leftOK := staticValueAsString(left)
		rightText, rightOK := staticValueAsString(right)
		if !leftOK || !rightOK {
			return false, false
		}
		return leftText == rightText, true
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
	case staticKindBigInt:
		leftBigInt, leftOK := left.(staticBigIntValue)
		rightBigInt, rightOK := right.(staticBigIntValue)
		if !leftOK || !rightOK || leftBigInt.value == nil || rightBigInt.value == nil {
			return false, false
		}
		if !staticReserveBigIntComparisonWork(leftBigInt, rightBigInt) {
			return false, false
		}
		return leftBigInt.value.Cmp(rightBigInt.value) == 0, true
	case staticKindSymbol:
		leftSymbol, leftOK := left.(staticSymbolValue)
		rightSymbol, rightOK := right.(staticSymbolValue)
		if !leftOK || !rightOK {
			return false, false
		}
		return staticSymbolIdentityEqual(leftSymbol, rightSymbol)
	}
	return true, true
}

func staticCanonicalBuiltinFunctionName(name string) string {
	if memberSeparator := strings.IndexByte(name, '.'); memberSeparator > 0 {
		if _, typedArray := staticTypedArrayBytesPerElement(name[:memberSeparator]); typedArray {
			switch name[memberSeparator:] {
			case ".from", ".of":
				return "TypedArray" + name[memberSeparator:]
			}
		}
	}
	if prototypeSeparator := strings.Index(name, ".prototype."); prototypeSeparator > 0 {
		if _, typedArray := staticTypedArrayBytesPerElement(name[:prototypeSeparator]); typedArray {
			canonical := "TypedArray" + name[prototypeSeparator:]
			if canonical == "TypedArray.prototype.toString" {
				return "Array.prototype.toString"
			}
			return canonical
		}
	}
	switch name {
	case "Number.parseInt":
		return "parseInt"
	case "Number.parseFloat":
		return "parseFloat"
	case "Set.prototype.keys":
		return "Set.prototype.values"
	case "Date.prototype.toGMTString":
		return "Date.prototype.toUTCString"
	case "String.prototype.trimLeft":
		return "String.prototype.trimStart"
	case "String.prototype.trimRight":
		return "String.prototype.trimEnd"
	default:
		return name
	}
}

func staticValueIsAggregate(value any) bool {
	switch value := value.(type) {
	case *staticObjectValue, *staticArrayValue, staticOpaqueObjectValue, staticIteratorValue,
		staticRegExpValue, staticDateValue, staticBoxedValue, staticCollectionValue,
		staticBuiltinObjectValue:
		return true
	case staticBuiltinValue:
		return !staticBuiltinIsFunction(string(value))
	default:
		return false
	}
}

func staticValueTruthy(value any) (truthy bool, ok bool) {
	switch value := value.(type) {
	case string:
		return value != "", true
	case staticNumberValue:
		return value != 0 && !value.IsNaN(), true
	case staticUnknownNumberValue, staticUnknownBooleanValue:
		return false, false
	case staticUnknownStringValue:
		if value.truthy {
			return true, true
		}
		return false, false
	case staticBigIntValue:
		if value.value != nil {
			return value.value.Sign() != 0, true
		}
		return value.truthy, value.truthyKnown
	case staticSymbolValue, staticBuiltinValue, staticBuiltinObjectValue, staticOpaqueObjectValue, staticIteratorValue,
		staticRegExpValue, staticDateValue, staticBoxedValue, staticCollectionValue:
		return true, true
	case *staticStringNode:
		stringValue, _ := staticValueAsString(value)
		return stringValue != "", true
	case bool:
		return value, true
	case staticNullValue, staticUndefinedValue, staticOptionalChainShortCircuitValue:
		return false, true
	case *staticObjectValue, *staticArrayValue:
		return true, true
	default:
		return evaluatorBool(func() bool { return evaluator.IsTruthy(value) })
	}
}

func staticValueToString(value any) (string, bool) {
	return staticValueToStringWithWork(value, true)
}

// staticValueToStringWithWork performs JavaScript's primitive string
// conversion. Enhanced-evaluator aggregates carry a shared file work budget;
// charge=false is used only while a surrounding streamed conversion has
// already reserved the complete allocation.
func staticValueToStringWithWork(value any, charge bool) (string, bool) {
	switch value := value.(type) {
	case string:
		return value, true
	case staticNumberValue:
		return ecmascript.NumberToString(float64(value)), true
	case staticBigIntValue:
		if value.value == nil {
			return "", false
		}
		if charge && !value.stringWorkContext.reserve(staticBigIntStringBytesUpperBound(value.value)) {
			return "", false
		}
		return value.value.String(), true
	case staticUnknownStringValue:
		return "", false
	case staticUnknownNumberValue, staticUnknownBooleanValue, staticSymbolValue,
		staticOpaqueObjectValue, staticIteratorValue:
		return "", false
	case staticRegExpValue:
		length := len(value.source) + len(value.flags) + 2
		if charge && value.stringWorkContext != nil &&
			(length > maxStaticStringLength || !value.stringWorkContext.reserve(length)) {
			value.stringWorkContext.exhaust()
			return "", false
		}
		var builder strings.Builder
		builder.Grow(length)
		builder.WriteByte('/')
		builder.WriteString(value.source)
		builder.WriteByte('/')
		builder.WriteString(value.flags)
		return builder.String(), true
	case staticDateValue:
		return "", false
	case staticBoxedValue:
		if _, symbol := value.value.(staticSymbolValue); symbol {
			return "", false
		}
		return staticValueToStringWithWork(value.value, charge)
	case staticCollectionValue:
		return "[object " + value.kind + "]", true
	case staticBuiltinValue:
		if !staticBuiltinIsFunction(string(value)) {
			return "[object " + string(value) + "]", true
		}
		return "", false
	case staticBuiltinObjectValue:
		name := strings.TrimSuffix(string(value), ".prototype")
		if value == "Array.prototype[Symbol.unscopables]" {
			// The object has a null prototype and therefore no ordinary primitive
			// conversion method. String/number coercion throws upstream.
			return "", false
		}
		if name == "Date" || name == "Symbol" {
			return "", false
		}
		if _, typedArray := staticTypedArrayBytesPerElement(name); typedArray {
			return "", false
		}
		switch name {
		case "Array", "String":
			return "", true
		case "Boolean":
			return "false", true
		case "BigInt", "Number":
			return "0", true
		case "RegExp":
			return "/(?:)/", true
		}
		return "[object " + name + "]", true
	case *staticStringNode:
		return staticValueAsString(value)
	case bool:
		if value {
			return "true", true
		}
		return "false", true
	case staticNullValue:
		return "null", true
	case staticUndefinedValue, staticOptionalChainShortCircuitValue:
		return "undefined", true
	case *staticArrayValue:
		return staticArrayToStringWithWork(value, charge)
	case *staticObjectValue:
		_, overridesToString := staticObjectOwnProperty(value, "toString")
		overridesValueOf := false
		overridesToPrimitive := false
		if value.enhancedCoercion {
			_, overridesValueOf = staticObjectOwnProperty(value, "valueOf")
			_, overridesToPrimitive = staticObjectOwnSymbolProperty(value, staticSymbolValue{
				description: "toPrimitive", wellKnown: true,
			})
		}
		if overridesToString || overridesValueOf || overridesToPrimitive || value.prototypeSet {
			return "", false
		}
		return "[object Object]", true
	default:
		return evaluatorString(func() string { return evaluator.AnyToString(value) })
	}
}

func staticBigIntStringBytesUpperBound(value *big.Int) int {
	if value == nil || value.Sign() == 0 {
		return 1
	}
	// ceil(bitLength * log10(2)) is an upper bound on the decimal digit
	// count. 30103/100000 is slightly larger than log10(2).
	digits := (value.BitLen()*30103 + 99999) / 100000
	if value.Sign() < 0 {
		digits++
	}
	return max(digits, 1)
}

// staticArrayToString implements `Array.prototype.join(',')`, which is what
// an array coerces to via the default ToString. Nullish elements fold to "",
// not "null"/"undefined" as a direct string conversion would give.
func staticArrayToString(value *staticArrayValue) (string, bool) {
	return staticArrayToStringWithWork(value, true)
}

func staticArrayToStringWithWork(value *staticArrayValue, charge bool) (string, bool) {
	return staticArrayJoinWithWork(value, ",", charge)
}

type staticStringMeasure struct {
	bytes             int
	codeUnits         int
	temporaryBytes    int
	resourceExhausted bool
}

const maxStaticArrayStringDepth = 1 << 10

func staticArrayJoin(value *staticArrayValue, separator string) (string, bool) {
	return staticArrayJoinWithWork(value, separator, true)
}

func staticArrayJoinWithWork(value *staticArrayValue, separator string, charge bool) (string, bool) {
	if value == nil || value.length < 0 {
		return "", false
	}
	resourceLimited := value.stringWorkContext != nil
	work := value.stringWorkContext
	if !charge {
		work = nil
	}
	measure, ok := measureStaticArrayJoin(
		value,
		separator,
		work,
		map[*staticArrayValue]bool{},
		0,
		resourceLimited,
	)
	if !ok {
		if measure.resourceExhausted {
			value.stringWorkContext.exhaust()
		}
		return "", false
	}
	allocationCost := measure.bytes + measure.temporaryBytes
	if charge && !value.stringWorkContext.reserve(allocationCost) {
		return "", false
	}

	var builder strings.Builder
	builder.Grow(measure.bytes)
	if !appendStaticArrayJoin(
		&builder,
		value,
		separator,
		map[*staticArrayValue]bool{},
		0,
		resourceLimited,
	) {
		return "", false
	}
	return builder.String(), true
}

func measureStaticArrayJoin(
	value *staticArrayValue,
	separator string,
	work *staticStringWorkContext,
	visiting map[*staticArrayValue]bool,
	depth int,
	resourceLimited bool,
) (staticStringMeasure, bool) {
	if value == nil || value.length < 0 || visiting[value] {
		return staticStringMeasure{}, false
	}
	if resourceLimited && depth > maxStaticArrayStringDepth || !work.reserve(1) {
		return staticStringMeasure{resourceExhausted: true}, false
	}
	visiting[value] = true
	defer delete(visiting, value)

	separatorMeasure, ok := measureStaticStringText(separator, work)
	if !ok {
		return staticStringMeasure{resourceExhausted: true}, false
	}
	var measure staticStringMeasure
	for index := range value.length {
		if index > 0 && !addStaticStringMeasure(&measure, separatorMeasure) {
			return staticStringMeasure{resourceExhausted: true}, false
		}
		element := value.element(index)
		if staticValueNullish(element) {
			continue
		}
		part, partOK := measureStaticStringValue(
			element,
			work,
			visiting,
			depth+1,
			resourceLimited,
		)
		if !partOK {
			return part, false
		}
		if !addStaticStringMeasure(&measure, part) {
			return staticStringMeasure{resourceExhausted: true}, false
		}
	}
	return measure, true
}

func measureStaticStringValue(
	value any,
	work *staticStringWorkContext,
	visiting map[*staticArrayValue]bool,
	depth int,
	resourceLimited bool,
) (staticStringMeasure, bool) {
	switch value := value.(type) {
	case string:
		return measureStaticStringText(value, work)
	case *staticStringNode:
		text, ok := staticValueAsString(value)
		if !ok {
			return staticStringMeasure{}, false
		}
		return measureStaticStringText(text, work)
	case staticBigIntValue:
		if value.value == nil || !work.reserve(1) {
			return staticStringMeasure{resourceExhausted: value.value != nil}, false
		}
		length := staticBigIntStringBytesUpperBound(value.value)
		return staticStringMeasure{bytes: length, codeUnits: length, temporaryBytes: length}, true
	case staticRegExpValue:
		if !work.reserve(len(value.source) + len(value.flags) + 2) {
			return staticStringMeasure{resourceExhausted: true}, false
		}
		return staticStringMeasure{
			bytes:     len(value.source) + len(value.flags) + 2,
			codeUnits: ecmascript.StringCodeUnitCount(value.source) + ecmascript.StringCodeUnitCount(value.flags) + 2,
		}, true
	case staticBoxedValue:
		if _, symbol := value.value.(staticSymbolValue); symbol {
			return staticStringMeasure{}, false
		}
		return measureStaticStringValue(value.value, work, visiting, depth, resourceLimited)
	case *staticArrayValue:
		return measureStaticArrayJoin(value, ",", work, visiting, depth, resourceLimited)
	default:
		// All remaining known conversions are bounded primitive formatting or
		// fixed built-in object tags. Reserve their maximum practical work
		// before asking the shared semantic conversion for the exact spelling.
		if !work.reserve(64) {
			return staticStringMeasure{resourceExhausted: true}, false
		}
		text, ok := staticValueToStringWithWork(value, false)
		if !ok {
			return staticStringMeasure{}, false
		}
		measure, ok := measureStaticStringText(text, nil)
		if !ok {
			return staticStringMeasure{resourceExhausted: true}, false
		}
		measure.temporaryBytes = len(text)
		return measure, true
	}
}

func measureStaticStringText(text string, work *staticStringWorkContext) (staticStringMeasure, bool) {
	if !work.reserve(max(len(text), 1)) {
		return staticStringMeasure{resourceExhausted: true}, false
	}
	measure := staticStringMeasure{bytes: len(text), codeUnits: ecmascript.StringCodeUnitCount(text)}
	if measure.bytes > maxStaticStringLength || measure.codeUnits > maxStaticStringLength {
		measure.resourceExhausted = true
		return measure, false
	}
	return measure, true
}

func addStaticStringMeasure(total *staticStringMeasure, part staticStringMeasure) bool {
	if total == nil || part.bytes > maxStaticStringLength-total.bytes ||
		part.codeUnits > maxStaticStringLength-total.codeUnits {
		return false
	}
	total.bytes += part.bytes
	total.codeUnits += part.codeUnits
	if part.temporaryBytes > maxStaticStringWorkBudget-total.temporaryBytes {
		total.temporaryBytes = maxStaticStringWorkBudget + 1
	} else {
		total.temporaryBytes += part.temporaryBytes
	}
	return true
}

func appendStaticArrayJoin(
	builder *strings.Builder,
	value *staticArrayValue,
	separator string,
	visiting map[*staticArrayValue]bool,
	depth int,
	resourceLimited bool,
) bool {
	if builder == nil || value == nil || visiting[value] ||
		resourceLimited && depth > maxStaticArrayStringDepth {
		return false
	}
	visiting[value] = true
	defer delete(visiting, value)
	for index := range value.length {
		if index > 0 {
			builder.WriteString(separator)
		}
		element := value.element(index)
		if staticValueNullish(element) {
			continue
		}
		if nested, ok := element.(*staticArrayValue); ok {
			if !appendStaticArrayJoin(builder, nested, ",", visiting, depth+1, resourceLimited) {
				return false
			}
			continue
		}
		switch element := element.(type) {
		case staticRegExpValue:
			builder.WriteByte('/')
			builder.WriteString(element.source)
			builder.WriteByte('/')
			builder.WriteString(element.flags)
		default:
			part, ok := staticValueToStringWithWork(element, false)
			if !ok {
				return false
			}
			builder.WriteString(part)
		}
	}
	return true
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
