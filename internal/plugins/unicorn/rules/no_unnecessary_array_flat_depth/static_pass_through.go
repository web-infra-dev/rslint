// cspell:ignore unscopables
package no_unnecessary_array_flat_depth

import (
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

const (
	maxStaticPassThroughDepth   = 1024
	maxStaticPassThroughSteps   = 4096
	maxStaticEvaluatorNodes     = 4096
	maxStaticEvaluatorBytes     = 1 << 20
	maxStaticEvaluatorFileNodes = 65536
	maxStaticEvaluatorFileBytes = 16 << 20
)

type staticPassThroughStateCacheKey struct{}

type staticPassThroughState struct {
	ctx                     rule.RuleContext
	evaluator               *utils.StaticStringEvaluator
	safeSymbols             map[staticSafetyCacheKey]bool
	expansionSymbolCosts    map[*ast.Symbol]staticExpansionCost
	expansionInitializers   map[*ast.Symbol]staticExpansionInitializer
	remainingEvaluatorNodes int
	remainingEvaluatorBytes int
	evaluatorPreflightDone  bool
	evaluatorPreflightSafe  bool
	stringCodeUnitLengths   map[*ast.Node]int
}

type staticPassThroughWalk struct {
	visiting  map[*ast.Symbol]bool
	depth     int
	steps     int
	exhausted bool
}

type staticSafetyCacheKey struct {
	symbol           *ast.Symbol
	allowPassThrough bool
}

type staticExpansionCost struct {
	nodes int
	bytes int
}

type staticExpansionInitializer struct {
	node   *ast.Node
	stable bool
}

func staticPassThroughStateForFile(ctx rule.RuleContext) *staticPassThroughState {
	return rule.CachedByFile(ctx, staticPassThroughStateCacheKey{}, func() *staticPassThroughState {
		return &staticPassThroughState{
			ctx:                     ctx,
			evaluator:               staticPassThroughEvaluator(ctx),
			safeSymbols:             map[staticSafetyCacheKey]bool{},
			expansionSymbolCosts:    map[*ast.Symbol]staticExpansionCost{},
			expansionInitializers:   map[*ast.Symbol]staticExpansionInitializer{},
			remainingEvaluatorNodes: maxStaticEvaluatorFileNodes,
			remainingEvaluatorBytes: maxStaticEvaluatorFileBytes,
			stringCodeUnitLengths:   map[*ast.Node]int{},
		}
	})
}

func (state *staticPassThroughState) isKnownNonArray(node *ast.Node) bool {
	argument, ok := staticObjectPassThroughArgument(state.ctx, node)
	if !ok {
		return false
	}
	if !state.isSafe(
		argument,
		&staticPassThroughWalk{visiting: map[*ast.Symbol]bool{}},
		true,
	) {
		return false
	}
	if isArray, known := state.evalArrayValue(node); known {
		return !isArray
	}
	return state.isFallbackKnownNonArray(
		argument,
		&staticPassThroughWalk{visiting: map[*ast.Symbol]bool{}},
		true,
	)
}

func (walk *staticPassThroughWalk) enter() bool {
	if walk.exhausted || walk.depth >= maxStaticPassThroughDepth ||
		walk.steps >= maxStaticPassThroughSteps {
		walk.exhausted = true
		return false
	}
	walk.depth++
	walk.steps++
	return true
}

func (walk *staticPassThroughWalk) leave() {
	walk.depth--
}

func (state *staticPassThroughState) isSafe(
	node *ast.Node,
	walk *staticPassThroughWalk,
	allowPassThrough bool,
) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || !walk.enter() {
		return false
	}
	defer walk.leave()

	if argument, ok := staticObjectPassThroughArgument(state.ctx, node); ok {
		return allowPassThrough && state.isSafe(argument, walk, true)
	}
	if ast.IsCallExpression(node) || ast.IsNewExpression(node) ||
		node.Kind == ast.KindAwaitExpression || node.Kind == ast.KindYieldExpression ||
		node.Kind == ast.KindDeleteExpression {
		return false
	}
	if ast.IsBinaryExpression(node) &&
		ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
		return false
	}
	if ast.IsAccessExpression(node) {
		if state.isSafeStaticGlobalMember(node) {
			return true
		}
		safe := true
		node.ForEachChild(func(child *ast.Node) bool {
			if !state.isSafe(child, walk, isStaticPassThroughAliasReference(child)) {
				safe = false
				return true
			}
			return false
		})
		return safe && state.isSafeMember(node)
	}
	if ast.IsIdentifier(node) && !utils.IsNonReferenceIdentifier(node) {
		if state.ctx.Refs != nil {
			symbol := state.ctx.Refs.Resolve(node)
			if !utils.IsValueSymbolDeclaredInFile(symbol, state.ctx.SourceFile) {
				return true
			}
			cacheKey := staticSafetyCacheKey{symbol: symbol, allowPassThrough: allowPassThrough}
			if safe, exists := state.safeSymbols[cacheKey]; exists {
				return safe
			}
			if walk.visiting[symbol] {
				return false
			}
			initializer, _, ok := state.constIdentifierInitializer(node)
			if !ok {
				state.safeSymbols[cacheKey] = false
				return false
			}
			walk.visiting[symbol] = true
			safe := state.isSafe(initializer, walk, allowPassThrough)
			delete(walk.visiting, symbol)
			if !walk.exhausted {
				state.safeSymbols[cacheKey] = safe
			}
			return safe
		}
		return false
	}

	safe := true
	node.ForEachChild(func(child *ast.Node) bool {
		if !state.isSafe(child, walk, isStaticPassThroughAliasReference(child)) {
			safe = false
			return true
		}
		return false
	})
	return safe
}

func isStaticPassThroughAliasReference(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return node != nil && ast.IsIdentifier(node) && !utils.IsNonReferenceIdentifier(node)
}

func (state *staticPassThroughState) isSafeMember(node *ast.Node) bool {
	name, ok := safeStaticMemberName(node)
	if !ok {
		return false
	}
	object := utils.ESTreeRuntimeExpression(utils.AccessExpressionObject(node))
	if object == nil {
		return false
	}
	if !ast.IsIdentifier(object) {
		switch object.Kind {
		case ast.KindArrayLiteralExpression:
			return isSafeArrayMember(object, name)
		case ast.KindObjectLiteralExpression:
			return isSafeObjectMember(object, name)
		case ast.KindStringLiteral:
			return state.isSafeStringMember(object, name)
		default:
			return false
		}
	}
	if state.isSafeStaticBuiltinGlobal(object) {
		return true
	}
	initializer, _, known := state.constIdentifierInitializer(object)
	initializer = utils.ESTreeRuntimeExpression(initializer)
	return known && initializer != nil && initializer.Kind == ast.KindStringLiteral &&
		state.isSafeStringMember(initializer, name)
}

func safeStaticMemberName(node *ast.Node) (string, bool) {
	if ast.IsElementAccessExpression(node) {
		property := utils.ESTreeRuntimeExpression(node.AsElementAccessExpression().ArgumentExpression)
		if property == nil || property.Kind != ast.KindStringLiteral &&
			property.Kind != ast.KindNumericLiteral {
			return "", false
		}
	}
	return utils.AccessExpressionStaticName(node)
}

func (state *staticPassThroughState) isSafeStringMember(node *ast.Node, name string) bool {
	if name == "length" {
		return true
	}
	index, err := strconv.Atoi(name)
	if err != nil || index < 0 || strconv.Itoa(index) != name {
		return false
	}
	length, ok := state.stringCodeUnitLengths[node]
	if !ok {
		length = ecmascript.StringCodeUnitCount(node.AsStringLiteral().Text)
		state.stringCodeUnitLengths[node] = length
	}
	return index < length
}

func isSafeArrayMember(node *ast.Node, name string) bool {
	elements := node.AsArrayLiteralExpression().Elements.Nodes
	for _, element := range elements {
		if element.Kind == ast.KindSpreadElement {
			return false
		}
	}
	if name == "length" {
		return true
	}
	index, err := strconv.Atoi(name)
	return err == nil && index >= 0 && index < len(elements) &&
		strconv.Itoa(index) == name && elements[index].Kind != ast.KindOmittedExpression
}

func isSafeObjectMember(node *ast.Node, name string) bool {
	found := false
	for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment &&
			property.Kind != ast.KindShorthandPropertyAssignment {
			return false
		}
		propertyName := property.Name()
		if propertyName == nil || ast.IsComputedPropertyName(propertyName) {
			return false
		}
		staticName, ok := utils.GetStaticPropertyName(propertyName)
		if !ok || staticName == "__proto__" {
			return false
		}
		if staticName == name {
			found = true
		}
	}
	return found
}

func (state *staticPassThroughState) isFallbackKnownNonArray(
	node *ast.Node,
	walk *staticPassThroughWalk,
	allowObjectLiteral bool,
) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || !walk.enter() {
		return false
	}
	defer walk.leave()

	if argument, ok := staticObjectPassThroughArgument(state.ctx, node); ok {
		return state.isFallbackKnownNonArray(argument, walk, allowObjectLiteral)
	}
	if state.isSafeStaticBuiltinGlobal(node) {
		return true
	}
	if ast.IsIdentifier(node) && !utils.IsNonReferenceIdentifier(node) {
		symbol := state.ctx.Refs.Resolve(node)
		if walk.visiting[symbol] {
			return false
		}
		initializer, _, ok := state.constIdentifierInitializer(node)
		if !ok {
			return false
		}
		walk.visiting[symbol] = true
		known := state.isFallbackAliasKnownNonArray(initializer, walk)
		delete(walk.visiting, symbol)
		return known
	}
	if state.isSafeStaticGlobalMember(node) {
		return true
	}
	if state.isFallbackArrayLength(node) {
		return true
	}
	switch node.Kind {
	case ast.KindBigIntLiteral:
		return true
	case ast.KindObjectLiteralExpression:
		return allowObjectLiteral && state.isFallbackObjectLiteral(node, walk)
	default:
		return false
	}
}

func (state *staticPassThroughState) isFallbackArrayLength(node *ast.Node) bool {
	if !ast.IsAccessExpression(node) {
		return false
	}
	name, ok := safeStaticMemberName(node)
	if !ok || name != "length" {
		return false
	}
	object := utils.ESTreeRuntimeExpression(utils.AccessExpressionObject(node))
	if object == nil || object.Kind != ast.KindArrayLiteralExpression {
		return false
	}
	isArray, known := state.evalArrayValue(object)
	return known && isArray
}

func (state *staticPassThroughState) isFallbackAliasKnownNonArray(
	node *ast.Node,
	walk *staticPassThroughWalk,
) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	if argument, ok := staticObjectPassThroughArgument(state.ctx, node); ok {
		return state.isFallbackKnownNonArray(argument, walk, false)
	}
	if ast.IsIdentifier(node) && !utils.IsNonReferenceIdentifier(node) {
		return state.isFallbackKnownNonArray(node, walk, false)
	}
	return node.Kind == ast.KindBigIntLiteral
}

func (state *staticPassThroughState) isFallbackObjectLiteral(
	node *ast.Node,
	walk *staticPassThroughWalk,
) bool {
	for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
		var value *ast.Node
		switch property.Kind {
		case ast.KindPropertyAssignment:
			name := property.Name()
			if name == nil || ast.IsComputedPropertyName(name) {
				return false
			}
			value = property.AsPropertyAssignment().Initializer
		case ast.KindShorthandPropertyAssignment:
			value = property.Name()
		case ast.KindSpreadAssignment:
			value = property.Expression()
		default:
			return false
		}
		if _, known := state.evalArrayValue(value); !known &&
			!state.isFallbackKnownNonArray(value, walk, false) {
			return false
		}
	}
	return true
}

func (state *staticPassThroughState) evalArrayValue(node *ast.Node) (bool, bool) {
	if !state.consumeEvaluatorBudget(node) {
		return false, false
	}
	return state.evaluator.EvalArrayValue(node)
}

func (state *staticPassThroughState) consumeEvaluatorBudget(node *ast.Node) bool {
	if state.remainingEvaluatorNodes <= 0 || state.remainingEvaluatorBytes <= 0 {
		return false
	}
	cost, cacheable := state.expansionCost(node, map[*ast.Symbol]bool{}, 0)
	if !cacheable || cost.nodes > maxStaticEvaluatorNodes ||
		cost.bytes > maxStaticEvaluatorBytes || !state.ensureEvaluatorPreflight() ||
		cost.nodes > state.remainingEvaluatorNodes ||
		cost.bytes > state.remainingEvaluatorBytes {
		return false
	}
	state.remainingEvaluatorNodes -= cost.nodes
	state.remainingEvaluatorBytes -= cost.bytes
	return true
}

func (state *staticPassThroughState) expansionCost(
	node *ast.Node,
	visiting map[*ast.Symbol]bool,
	depth int,
) (staticExpansionCost, bool) {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return staticExpansionCost{}, true
	}
	if depth >= maxStaticPassThroughDepth {
		return staticExpansionOverflow(), false
	}
	cost := staticExpansionCost{nodes: 1, bytes: staticLiteralBytes(node)}
	if ast.IsIdentifier(node) && !utils.IsNonReferenceIdentifier(node) {
		initializer, symbol, ok := state.stableExpansionInitializer(node)
		if !ok {
			return cost, true
		}
		if cached, exists := state.expansionSymbolCosts[symbol]; exists {
			return addExpansionCosts(cost, cached), true
		}
		if visiting[symbol] {
			return staticExpansionOverflow(), false
		}
		visiting[symbol] = true
		initializerCost, cacheable := state.expansionCost(initializer, visiting, depth+1)
		delete(visiting, symbol)
		if cacheable {
			state.expansionSymbolCosts[symbol] = initializerCost
		}
		return addExpansionCosts(cost, initializerCost), cacheable
	}
	node.ForEachChild(func(child *ast.Node) bool {
		childCost, cacheable := state.expansionCost(child, visiting, depth+1)
		cost = addExpansionCosts(cost, childCost)
		if !cacheable {
			cost = staticExpansionOverflow()
			return true
		}
		return cost.nodes > maxStaticEvaluatorNodes || cost.bytes > maxStaticEvaluatorBytes
	})
	// The shared evaluator tries its builtin-call path before its Object
	// pass-through path, so a pass-through argument may be evaluated twice.
	if argument, ok := staticObjectPassThroughArgument(state.ctx, node); ok {
		argumentCost, cacheable := state.expansionCost(argument, visiting, depth+1)
		cost = addExpansionCosts(cost, argumentCost)
		if !cacheable {
			return staticExpansionOverflow(), false
		}
	}
	return cost, true
}

func (state *staticPassThroughState) ensureEvaluatorPreflight() bool {
	if state.evaluatorPreflightDone {
		return state.evaluatorPreflightSafe
	}
	state.evaluatorPreflightDone = true
	remainingNodes := state.remainingEvaluatorNodes
	remainingBytes := state.remainingEvaluatorBytes
	safe := true
	var visit func(*ast.Node, int)
	visit = func(node *ast.Node, depth int) {
		if node == nil || !safe {
			return
		}
		if depth >= maxStaticPassThroughDepth {
			safe = false
			return
		}
		if ast.IsCallExpression(node) {
			if key := staticDynamicReferenceCalleeKey(node); key != nil {
				cost, cacheable := state.expansionCost(key, map[*ast.Symbol]bool{}, 0)
				if !cacheable || cost.nodes > maxStaticEvaluatorNodes ||
					cost.bytes > maxStaticEvaluatorBytes ||
					cost.nodes > remainingNodes || cost.bytes > remainingBytes {
					safe = false
					return
				}
				remainingNodes -= cost.nodes
				remainingBytes -= cost.bytes
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child, depth+1)
			return !safe
		})
	}
	visit(state.ctx.SourceFile.AsNode(), 0)
	if !safe {
		return false
	}
	state.remainingEvaluatorNodes = remainingNodes
	state.remainingEvaluatorBytes = remainingBytes
	state.evaluatorPreflightSafe = true
	return true
}

func staticDynamicReferenceCalleeKey(node *ast.Node) *ast.Node {
	callee := ast.SkipOuterExpressions(
		node.AsCallExpression().Expression,
		ast.OEKParentheses|ast.OEKAssertions,
	)
	if !ast.IsElementAccessExpression(callee) {
		return nil
	}
	if _, known := utils.AccessExpressionStaticName(callee); known {
		return nil
	}
	base := callee
	for ast.IsAccessExpression(base) {
		base = ast.SkipOuterExpressions(
			utils.AccessExpressionObject(base),
			ast.OEKParentheses|ast.OEKAssertions,
		)
	}
	if !ast.IsIdentifier(base) || utils.IsNonReferenceIdentifier(base) {
		return nil
	}
	return callee.AsElementAccessExpression().ArgumentExpression
}

func (state *staticPassThroughState) stableExpansionInitializer(
	node *ast.Node,
) (*ast.Node, *ast.Symbol, bool) {
	initializer, symbol, declarationList, ok := state.variableInitializer(node)
	if !ok {
		return nil, nil, false
	}
	if cached, exists := state.expansionInitializers[symbol]; exists {
		return cached.node, symbol, cached.stable
	}
	if !ast.IsVarConst(declarationList) {
		for _, reference := range state.ctx.Refs.References(symbol) {
			if utils.IsWriteReference(reference) {
				state.expansionInitializers[symbol] = staticExpansionInitializer{}
				return nil, nil, false
			}
		}
	}
	state.expansionInitializers[symbol] = staticExpansionInitializer{node: initializer, stable: true}
	return initializer, symbol, true
}

func (state *staticPassThroughState) constIdentifierInitializer(
	node *ast.Node,
) (*ast.Node, *ast.Symbol, bool) {
	initializer, symbol, declarationList, ok := state.variableInitializer(node)
	if !ok || !ast.IsVarConst(declarationList) ||
		!utils.IsValueSymbolDeclaredInFile(symbol, state.ctx.SourceFile) {
		return nil, nil, false
	}
	return initializer, symbol, true
}

func (state *staticPassThroughState) variableInitializer(
	node *ast.Node,
) (*ast.Node, *ast.Symbol, *ast.Node, bool) {
	if state.ctx.Refs == nil {
		return nil, nil, nil, false
	}
	symbol := state.ctx.Refs.Resolve(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil, nil, nil, false
	}
	declarationNode := symbol.Declarations[0]
	if declarationNode == nil || declarationNode.Kind != ast.KindVariableDeclaration {
		return nil, nil, nil, false
	}
	declaration := declarationNode.AsVariableDeclaration()
	name := declaration.Name()
	if declaration.Initializer == nil || name == nil || !ast.IsIdentifier(name) ||
		name.AsIdentifier().Text != node.AsIdentifier().Text {
		return nil, nil, nil, false
	}
	declarationList := utils.GetDeclListForSymbolDecl(declarationNode)
	if declarationList == nil || ast.IsVarUsing(declarationList) ||
		ast.IsVarAwaitUsing(declarationList) {
		return nil, nil, nil, false
	}
	return declaration.Initializer, symbol, declarationList, true
}

func staticLiteralBytes(node *ast.Node) int {
	switch node.Kind {
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateHead,
		ast.KindTemplateMiddle,
		ast.KindTemplateTail:
		return min(len(node.Text()), maxStaticEvaluatorBytes+1)
	default:
		return 0
	}
}

func staticExpansionOverflow() staticExpansionCost {
	return staticExpansionCost{
		nodes: maxStaticEvaluatorNodes + 1,
		bytes: maxStaticEvaluatorBytes + 1,
	}
}

func addExpansionCosts(left, right staticExpansionCost) staticExpansionCost {
	return staticExpansionCost{
		nodes: saturatedAdd(left.nodes, right.nodes, maxStaticEvaluatorNodes),
		bytes: saturatedAdd(left.bytes, right.bytes, maxStaticEvaluatorBytes),
	}
}

func saturatedAdd(left, right, limit int) int {
	if left > limit || right > limit || right > limit-left {
		return limit + 1
	}
	return left + right
}

// @eslint-community/eslint-utils 4.10.1 builtinNames. Unicorn v74 lets its
// static evaluator read these globals directly; the scope checks below keep
// shadowed names out.
var staticBuiltinGlobalNames = map[string]bool{
	"Array": true, "ArrayBuffer": true, "BigInt": true,
	"BigInt64Array": true, "BigUint64Array": true, "Boolean": true,
	"DataView": true, "Date": true, "decodeURI": true,
	"decodeURIComponent": true, "encodeURI": true, "encodeURIComponent": true,
	"escape": true, "Float32Array": true, "Float64Array": true,
	"Function": true, "Infinity": true, "Int16Array": true,
	"Int32Array": true, "Int8Array": true, "isFinite": true,
	"isNaN": true, "isPrototypeOf": true, "JSON": true,
	"Map": true, "Math": true, "NaN": true,
	"Number": true, "Object": true, "parseFloat": true,
	"parseInt": true, "Promise": true, "Proxy": true,
	"Reflect": true, "RegExp": true, "Set": true,
	"String": true, "Symbol": true, "Uint16Array": true,
	"Uint32Array": true, "Uint8Array": true, "Uint8ClampedArray": true,
	"undefined": true, "unescape": true, "WeakMap": true,
	"WeakSet": true,
}

var staticGlobalProperties = map[string]map[string]bool{
	"Math": {
		"E": true, "LN2": true, "LN10": true, "LOG2E": true, "LOG10E": true,
		"PI": true, "SQRT1_2": true, "SQRT2": true,
	},
	"Number": {
		"EPSILON": true, "MAX_SAFE_INTEGER": true, "MAX_VALUE": true,
		"MIN_SAFE_INTEGER": true, "MIN_VALUE": true, "NaN": true,
		"NEGATIVE_INFINITY": true, "POSITIVE_INFINITY": true,
	},
	"String": {"raw": true},
	"Symbol": {
		"asyncIterator": true, "hasInstance": true, "isConcatSpreadable": true,
		"iterator": true, "match": true, "matchAll": true, "replace": true,
		"search": true, "species": true, "split": true, "toPrimitive": true,
		"toStringTag": true, "unscopables": true,
	},
}

func (state *staticPassThroughState) isSafeStaticBuiltinGlobal(node *ast.Node) bool {
	if node == nil || !ast.IsIdentifier(node) || state.ctx.Refs == nil {
		return false
	}
	name := node.AsIdentifier().Text
	return staticBuiltinGlobalNames[name] && state.ctx.Globals.Access(name).IsDeclared() &&
		state.ctx.Refs.IsGlobalReference(node)
}

func (state *staticPassThroughState) isSafeStaticGlobalMember(node *ast.Node) bool {
	if node == nil || !ast.IsAccessExpression(node) || ast.IsOptionalChain(node) {
		return false
	}
	object := utils.ESTreeRuntimeExpression(utils.AccessExpressionObject(node))
	if object == nil || !ast.IsIdentifier(object) || state.ctx.Refs == nil ||
		!state.ctx.Globals.Access(object.AsIdentifier().Text).IsDeclared() ||
		!state.ctx.Refs.IsGlobalReference(object) {
		return false
	}
	property, ok := staticGlobalMemberName(node)
	return ok && staticGlobalProperties[object.AsIdentifier().Text][property]
}

func staticGlobalMemberName(node *ast.Node) (string, bool) {
	if ast.IsPropertyAccessExpression(node) {
		name := node.AsPropertyAccessExpression().Name()
		if name != nil && name.Kind != ast.KindPrivateIdentifier {
			return name.Text(), true
		}
		return "", false
	}
	if !ast.IsElementAccessExpression(node) {
		return "", false
	}
	property := utils.ESTreeRuntimeExpression(node.AsElementAccessExpression().ArgumentExpression)
	if property == nil || property.Kind != ast.KindStringLiteral {
		return "", false
	}
	return property.AsStringLiteral().Text, true
}
