package unicornutil

import (
	"math"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// IsKnownNonArray reports whether node is definitely not an Array or
// ReadonlyArray. Typed arrays and keyed collections are non-arrays. Unknown
// receivers return false so syntactic array rules can still report them.
func IsKnownNonArray(ctx rule.RuleContext, node *ast.Node) bool {
	return classifyArrayReceiver(ctx, node, arrayTargets, knownNonArrayNames) == arrayClassNonTarget
}

// IsKnownNonIndexedCollection reports whether node is definitely a keyed
// collection. Typed arrays remain indexed-collection targets.
func IsKnownNonIndexedCollection(ctx rule.RuleContext, node *ast.Node) bool {
	return classifyArrayReceiver(ctx, node, indexedCollectionTargets, keyedCollectionNames) == arrayClassNonTarget
}

// IsArray reports whether node is definitely an Array, ReadonlyArray, or
// tuple, using syntactic shape first and the type checker as a fallback.
func IsArray(ctx rule.RuleContext, node *ast.Node) bool {
	return classifyArrayReceiver(ctx, node, arrayTargets, knownNonArrayNames) == arrayClassTarget
}

// arrayClass is the array-receiver spelling of the shared TypeClass. It is an
// alias rather than its own type so the syntactic path below and the shared
// type-checker classifier can hand values to each other directly.
type arrayClass = TypeClass

const (
	arrayClassUnknown   = TypeUnknown
	arrayClassTarget    = TypeTarget
	arrayClassNonTarget = TypeNonTarget
)

var arrayTargets = utils.NewSetFromItems("Array", "ReadonlyArray")

var indexedCollectionTargets = func() *utils.Set[string] {
	s := utils.NewSetFromItems("Array", "ReadonlyArray")
	for _, name := range typedArrayNames {
		s.Add(name)
	}
	return s
}()

var keyedCollectionNames = utils.NewSetFromItems(
	"Map", "ReadonlyMap", "WeakMap", "Set", "ReadonlySet", "WeakSet",
)

var knownNonArrayNames = func() *utils.Set[string] {
	s := utils.NewSetFromItems(
		"Map", "ReadonlyMap", "WeakMap", "Set", "ReadonlySet", "WeakSet",
	)
	for _, name := range typedArrayNames {
		s.Add(name)
	}
	return s
}()

var typedArrayNames = []string{
	"Int8Array", "Uint8Array", "Uint8ClampedArray", "Int16Array", "Uint16Array",
	"Int32Array", "Uint32Array", "Float16Array", "Float32Array", "Float64Array",
	"BigInt64Array", "BigUint64Array",
}

type arrayReceiverStaticEvaluatorFileCacheKey struct{}
type sourceOnlyArrayReceiverCacheKey struct{}

type arrayReceiverWalkState struct {
	useTypeInformation bool
	visiting           map[*ast.Symbol]bool
	memo               map[*ast.Symbol]arrayClass
	depth              int
	steps              int
}

const (
	maxSourceOnlyArrayReceiverDepth = 1024
	maxSourceOnlyArrayReceiverSteps = 4096
)

// classifyArrayReceiver mirrors the syntactic short-circuits of unicorn's
// getType (array literals, Array()/new Array()/Array.from/Array.of), then falls
// back to a type-checker classification analogous to getTypeScriptType.
func classifyArrayReceiver(ctx rule.RuleContext, node *ast.Node, targetNames, nonTargetNames *utils.Set[string]) arrayClass {
	return classifyArrayReceiverInner(ctx, node, targetNames, nonTargetNames, &arrayReceiverWalkState{
		useTypeInformation: true,
		visiting:           map[*ast.Symbol]bool{},
	})
}

// classifySourceOnlyArrayReceiver follows Unicorn's syntax and static-value
// paths without letting tsgo's ambient JavaScript types turn an otherwise
// unknown expression into a known non-array.
func classifySourceOnlyArrayReceiver(
	ctx rule.RuleContext,
	node *ast.Node,
	targetNames, nonTargetNames *utils.Set[string],
) arrayClass {
	memo := rule.CachedByFile(ctx, sourceOnlyArrayReceiverCacheKey{}, func() map[*ast.Symbol]arrayClass {
		return map[*ast.Symbol]arrayClass{}
	})
	return classifyArrayReceiverInner(ctx, node, targetNames, nonTargetNames, &arrayReceiverWalkState{
		visiting: map[*ast.Symbol]bool{},
		memo:     memo,
	})
}

func classifyArrayReceiverInner(
	ctx rule.RuleContext,
	node *ast.Node,
	targetNames, nonTargetNames *utils.Set[string],
	state *arrayReceiverWalkState,
) arrayClass {
	if node == nil {
		return arrayClassUnknown
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return arrayClassUnknown
	}
	if !state.useTypeInformation {
		if state.depth >= maxSourceOnlyArrayReceiverDepth ||
			state.steps >= maxSourceOnlyArrayReceiverSteps {
			return arrayClassUnknown
		}
		state.depth++
		state.steps++
		defer func() { state.depth-- }()
	}

	// Unicorn deliberately ignores a satisfies annotation and classifies the
	// expression itself. Assertions are different: an informative annotation
	// wins, while any / unknown falls through to the wrapped expression.
	if node.Kind == ast.KindSatisfiesExpression {
		return classifyArrayReceiverInner(
			ctx, node.AsSatisfiesExpression().Expression, targetNames, nonTargetNames, state,
		)
	}
	if node.Kind == ast.KindAsExpression || node.Kind == ast.KindTypeAssertionExpression {
		if state.useTypeInformation {
			if class := classifyArrayTypeNode(ctx, node.Type(), targetNames, nonTargetNames, map[*ast.Symbol]bool{}); class != arrayClassUnknown {
				return class
			}
		}
		return classifyArrayReceiverInner(ctx, node.Expression(), targetNames, nonTargetNames, state)
	}
	if node.Kind == ast.KindNonNullExpression {
		return classifyArrayReceiverInner(ctx, node.Expression(), targetNames, nonTargetNames, state)
	}

	if ast.IsIdentifier(node) {
		if class := classifyArrayIdentifier(ctx, node, targetNames, nonTargetNames, state); class != arrayClassUnknown {
			return class
		}
	}
	if node.Kind == ast.KindConditionalExpression {
		conditional := node.AsConditionalExpression()
		class := combineArrayClassesUnion([]arrayClass{
			classifyArrayReceiverInner(ctx, conditional.WhenTrue, targetNames, nonTargetNames, state),
			classifyArrayReceiverInner(ctx, conditional.WhenFalse, targetNames, nonTargetNames, state),
		})
		if class != arrayClassUnknown || state.useTypeInformation {
			return class
		}
	}
	if ast.IsBinaryExpression(node) &&
		node.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
		return classifyArrayReceiverInner(
			ctx, node.AsBinaryExpression().Right, targetNames, nonTargetNames, state,
		)
	}

	// getStaticValue runs before Unicorn's syntactic target/non-target checks.
	// Identifier and member expressions are intentionally excluded even when
	// eslint-utils can fold them; upstream keeps those shapes unknown. In
	// source-only files, calls take the guarded path below so tsgo's ambient
	// return types cannot classify arbitrary user calls.
	if !state.useTypeInformation && ast.IsCallExpression(node) {
		if class := classifySourceOnlyStaticCall(ctx, node); class != arrayClassUnknown {
			return class
		}
	} else if !ast.IsIdentifier(node) && !ast.IsAccessExpression(node) {
		// Keep source-only evaluation bounded. The shared evaluator is recursive;
		// repeatedly folding the same adversarial const chain here would otherwise
		// be both more permissive than eslint-utils and quadratic in practice.
		if state.useTypeInformation ||
			(!containsInvocationExpression(node) && isBoundedSourceStaticExpression(ctx, node)) {
			staticEvaluator := arrayReceiverStaticEvaluator(ctx)
			if isArray, known := staticEvaluator.EvalArrayValue(node); known {
				if isArray {
					return arrayClassTarget
				}
				return arrayClassNonTarget
			}
		}
	}
	if isSyntacticArrayNode(node) {
		return arrayClassTarget
	}
	if isSyntacticTargetConstructor(node, targetNames) {
		return arrayClassTarget
	}

	if state.useTypeInformation && ctx.TypeChecker != nil {
		t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)
		if class := classifyArrayType(ctx, t, targetNames); class != arrayClassUnknown {
			return class
		}
	}

	if isSyntacticNonArrayNode(node) {
		return arrayClassNonTarget
	}
	return arrayClassUnknown
}

func arrayReceiverStaticEvaluator(ctx rule.RuleContext) *utils.StaticStringEvaluator {
	return rule.CachedByFile(ctx, arrayReceiverStaticEvaluatorFileCacheKey{}, func() *utils.StaticStringEvaluator {
		return utils.NewStaticStringEvaluatorWithReferenceResolver(
			ctx.TypeChecker, ctx.SourceFile, ctx.Refs,
		)
	})
}

const staticCallGlobalMeaning = ast.SymbolFlagsValue | ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
const maxSourceStaticNodeCount = 256

// classifySourceOnlyStaticCall delegates value proofs to rslint's evaluator.
// This function only carries eslint-utils' missing policy: which built-ins may
// be called, which argument conversions can throw, and the result category.
func classifySourceOnlyStaticCall(
	ctx rule.RuleContext,
	node *ast.Node,
) arrayClass {
	call := node.AsCallExpression()
	if call == nil {
		return arrayClassUnknown
	}
	evaluator := arrayReceiverStaticEvaluator(ctx)
	bounded := isBoundedSourceStaticExpression(ctx, node)
	root, method, builtin := staticBuiltinCall(ctx, call.Expression)
	if !bounded && !builtin {
		return arrayClassUnknown
	}
	arguments := node.Arguments()
	if !staticCallArgumentsKnown(ctx, evaluator, arguments) {
		return arrayClassUnknown
	}
	if bounded {
		if isArray, known := evaluator.EvalArrayValue(node); known {
			if isArray {
				return arrayClassTarget
			}
			return arrayClassNonTarget
		}
	}
	if builtin {
		switch {
		case method == "" && root == "Boolean":
			return arrayClassNonTarget
		case method == "" && (root == "Number" || root == "String" || root == "parseFloat"):
			if len(arguments) == 0 || canStaticallyConvertToString(evaluator, arguments[0]) {
				return arrayClassNonTarget
			}
		case method == "" && root == "parseInt":
			if (len(arguments) == 0 || canStaticallyConvertToString(evaluator, arguments[0])) &&
				(len(arguments) < 2 || isStaticPrimitiveArgument(evaluator, arguments[1])) {
				return arrayClassNonTarget
			}
		case method == "" && (root == "isFinite" || root == "isNaN"):
			if len(arguments) == 0 || isStaticPrimitiveArgument(evaluator, arguments[0]) {
				return arrayClassNonTarget
			}
		case method == "" && root == "BigInt":
			if len(arguments) > 0 && isStaticallyValidBigIntArgument(arguments[0]) {
				return arrayClassNonTarget
			}
		case method == "" && root == "Object":
			if len(arguments) == 0 {
				return arrayClassNonTarget
			}
			if isArray, known := evaluator.EvalArrayValue(arguments[0]); known {
				if isArray {
					return arrayClassTarget
				}
				return arrayClassNonTarget
			}
		case root == "Array" && method == "isArray":
			return arrayClassNonTarget
		case root == "Math" && isStaticMathMethod(method) && staticPrimitiveArguments(evaluator, arguments):
			return arrayClassNonTarget
		}
	}
	return arrayClassUnknown
}

// isBoundedSourceStaticExpression is a resource guard around the shared
// evaluator, not a second evaluator. Built-in globals are allowed; local value
// references stay on the recursive classifier's memoized path.
func isBoundedSourceStaticExpression(ctx rule.RuleContext, root *ast.Node) bool {
	stack := []*ast.Node{root}
	for visited := 0; len(stack) > 0; visited++ {
		if visited >= maxSourceStaticNodeCount {
			return false
		}
		last := len(stack) - 1
		node := stack[last]
		stack = stack[:last]
		if ast.IsIdentifier(node) && !utils.IsNonReferenceIdentifier(node) &&
			!isUnshadowedGlobal(ctx, node) {
			return false
		}
		node.ForEachChild(func(child *ast.Node) bool {
			stack = append(stack, child)
			return false
		})
	}
	return true
}

// Calls and constructions have their own source-only policies. Do not let a
// compound expression feed a nested invocation back into the generic
// evaluator and bypass its argument/global checks.
func containsInvocationExpression(root *ast.Node) bool {
	stack := []*ast.Node{root}
	for len(stack) > 0 {
		last := len(stack) - 1
		node := stack[last]
		stack = stack[:last]
		if ast.IsCallExpression(node) || ast.IsNewExpression(node) {
			return true
		}
		node.ForEachChild(func(child *ast.Node) bool {
			stack = append(stack, child)
			return false
		})
	}
	return false
}

func staticBuiltinCall(ctx rule.RuleContext, rawCallee *ast.Node) (root, method string, ok bool) {
	callee := ast.SkipOuterExpressions(rawCallee, ast.OEKParentheses|ast.OEKAssertions)
	if ast.IsIdentifier(callee) && isUnshadowedGlobal(ctx, callee) {
		return callee.AsIdentifier().Text, "", true
	}
	if !ast.IsAccessExpression(callee) {
		return "", "", false
	}
	object := ast.SkipParentheses(utils.AccessExpressionObject(callee))
	method, ok = utils.AccessExpressionStaticName(callee)
	if !ok || !ast.IsIdentifier(object) || !isUnshadowedGlobal(ctx, object) {
		return "", "", false
	}
	return object.AsIdentifier().Text, method, true
}

func staticCallArgumentsKnown(
	ctx rule.RuleContext,
	evaluator *utils.StaticStringEvaluator,
	arguments []*ast.Node,
) bool {
	for _, argument := range arguments {
		if ast.IsSpreadElement(argument) {
			return false
		}
		if isBigIntLiteral(argument) {
			continue
		}
		if !isBoundedSourceStaticExpression(ctx, argument) {
			return false
		}
		if _, known := evaluator.EvalValue(argument); !known {
			return false
		}
	}
	return true
}

func isBigIntLiteral(node *ast.Node) bool {
	node = ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKAssertions)
	return node != nil && node.Kind == ast.KindBigIntLiteral
}

func isStaticallyValidBigIntArgument(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKAssertions)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindBigIntLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindNumericLiteral:
		value, ok := ecmascript.StringToNumber(utils.NormalizeNumericLiteral(node.Text()))
		return ok && !math.IsInf(value, 0) && !math.IsNaN(value) && value == math.Trunc(value)
	}
	return false
}

func canStaticallyConvertToString(evaluator *utils.StaticStringEvaluator, node *ast.Node) bool {
	if isBigIntLiteral(node) {
		return true
	}
	_, ok := evaluator.EvalToString(node)
	return ok
}

func isStaticPrimitiveArgument(evaluator *utils.StaticStringEvaluator, node *ast.Node) bool {
	value, known := evaluator.EvalValue(node)
	if known {
		switch value.(type) {
		case string, bool, interface{ IsNaN() bool }:
			return true
		}
	}
	node = ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKAssertions)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindNullKeyword, ast.KindUndefinedKeyword, ast.KindVoidExpression:
		return true
	}
	return false
}

func staticPrimitiveArguments(evaluator *utils.StaticStringEvaluator, arguments []*ast.Node) bool {
	for _, argument := range arguments {
		if !isStaticPrimitiveArgument(evaluator, argument) {
			return false
		}
	}
	return true
}

func isUnshadowedGlobal(ctx rule.RuleContext, identifier *ast.Node) bool {
	name := identifier.AsIdentifier().Text
	return ctx.Globals.Access(name).IsDeclared() && ctx.Refs != nil &&
		ctx.Refs.IsGlobalNameReference(identifier, name, staticCallGlobalMeaning)
}

func isStaticMathMethod(name string) bool {
	switch name {
	case "abs", "acos", "acosh", "asin", "asinh", "atan", "atan2", "atanh",
		"cbrt", "ceil", "clz32", "cos", "cosh", "exp", "expm1", "floor", "fround",
		"hypot", "imul", "log", "log10", "log1p", "log2", "max", "min",
		"pow", "round", "sign", "sin", "sinh", "sqrt", "tan", "tanh", "trunc":
		return true
	}
	return false
}

func classifyArrayIdentifier(
	ctx rule.RuleContext,
	idNode *ast.Node,
	targetNames, nonTargetNames *utils.Set[string],
	state *arrayReceiverWalkState,
) arrayClass {
	if ctx.Refs == nil || idNode == nil {
		return arrayClassUnknown
	}
	symbol := ctx.Refs.ResolveInFile(idNode)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return arrayClassUnknown
	}
	if state.visiting[symbol] {
		return arrayClassUnknown
	}
	if !state.useTypeInformation {
		if class, ok := state.memo[symbol]; ok {
			return class
		}
	}
	state.visiting[symbol] = true
	defer func() {
		delete(state.visiting, symbol)
	}()

	declaration := symbol.Declarations[0]
	if declaration == nil {
		return arrayClassUnknown
	}
	if state.useTypeInformation {
		annotation := arrayBindingTypeAnnotation(declaration)
		if class := classifyArrayTypeNode(ctx, annotation, targetNames, nonTargetNames, map[*ast.Symbol]bool{}); class != arrayClassUnknown {
			return class
		}
	}
	if !ast.IsVariableDeclaration(declaration) {
		return arrayClassUnknown
	}
	declarationList := declaration.Parent
	variable := declaration.AsVariableDeclaration()
	if declarationList == nil || !ast.IsVariableDeclarationList(declarationList) ||
		(!state.useTypeInformation && (ast.IsVarUsing(declarationList) || ast.IsVarAwaitUsing(declarationList))) ||
		declarationList.Flags&ast.NodeFlagsConst == 0 || variable.Name() == nil ||
		!ast.IsIdentifier(variable.Name()) || variable.Initializer == nil {
		return arrayClassUnknown
	}
	class := classifyArrayReceiverInner(ctx, variable.Initializer, targetNames, nonTargetNames, state)
	if !state.useTypeInformation {
		state.memo[symbol] = class
	}
	return class
}

// arrayBindingTypeAnnotation mirrors definition.name?.typeAnnotation in
// @typescript-eslint/scope-manager, whose Definition kinds carry a binding
// annotation only for variables and parameters. Function-like declarations are
// excluded because Node.Type() exposes their return type, which is not an
// annotation on the identifier binding. Class properties and interface members
// are not reachable here: the caller resolves a bare identifier, which never
// binds to one.
func arrayBindingTypeAnnotation(declaration *ast.Node) *ast.Node {
	if declaration == nil {
		return nil
	}
	switch declaration.Kind {
	case ast.KindVariableDeclaration, ast.KindParameter:
		return declaration.Type()
	default:
		return nil
	}
}

func classifyArrayTypeNode(
	ctx rule.RuleContext,
	node *ast.Node,
	targetNames, nonTargetNames *utils.Set[string],
	visitedSymbols map[*ast.Symbol]bool,
) arrayClass {
	if node == nil {
		return arrayClassUnknown
	}
	for node.Kind == ast.KindParenthesizedType {
		node = node.AsParenthesizedTypeNode().Type
		if node == nil {
			return arrayClassUnknown
		}
	}
	if node.Kind == ast.KindTypeOperator {
		typeOperator := node.AsTypeOperatorNode()
		if typeOperator.Operator != ast.KindReadonlyKeyword {
			return arrayClassUnknown
		}
		return classifyArrayTypeNode(ctx, typeOperator.Type, targetNames, nonTargetNames, visitedSymbols)
	}

	switch node.Kind {
	case ast.KindArrayType, ast.KindTupleType:
		return arrayClassTarget
	case ast.KindNullKeyword, ast.KindUndefinedKeyword:
		return arrayClassNonTarget
	case ast.KindUnionType:
		types := node.AsUnionTypeNode().Types.Nodes
		classes := make([]arrayClass, 0, len(types))
		for _, part := range types {
			classes = append(classes, classifyArrayTypeNode(ctx, part, targetNames, nonTargetNames, visitedSymbols))
		}
		return combineArrayClassesUnion(classes)
	case ast.KindIntersectionType:
		types := node.AsIntersectionTypeNode().Types.Nodes
		classes := make([]arrayClass, 0, len(types))
		for _, part := range types {
			classes = append(classes, classifyArrayTypeNode(ctx, part, targetNames, nonTargetNames, visitedSymbols))
		}
		return combineArrayClassesIntersection(classes)
	case ast.KindTypeReference:
		typeName := node.AsTypeReferenceNode().TypeName
		return classifyArrayTypeReference(ctx, typeName, targetNames, nonTargetNames, visitedSymbols)
	case ast.KindBigIntKeyword, ast.KindBooleanKeyword, ast.KindNeverKeyword,
		ast.KindNumberKeyword, ast.KindStringKeyword, ast.KindSymbolKeyword,
		ast.KindVoidKeyword, ast.KindLiteralType, ast.KindTypeLiteral,
		ast.KindFunctionType, ast.KindConstructorType:
		return arrayClassNonTarget
	default:
		return arrayClassUnknown
	}
}

func classifyArrayTypeReference(
	ctx rule.RuleContext,
	typeName *ast.Node,
	targetNames, nonTargetNames *utils.Set[string],
	visitedSymbols map[*ast.Symbol]bool,
) arrayClass {
	if typeName == nil || !ast.IsIdentifier(typeName) {
		return arrayClassUnknown
	}
	name := typeName.AsIdentifier().Text
	var symbol *ast.Symbol
	if ctx.Refs != nil {
		symbol = ctx.Refs.ResolveInFileWithMeaning(
			typeName, ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias,
		)
	}
	if symbol == nil {
		return classifyKnownArrayTypeName(name, targetNames, nonTargetNames)
	}
	if visitedSymbols[symbol] {
		return arrayClassUnknown
	}
	visitedSymbols[symbol] = true
	defer delete(visitedSymbols, symbol)

	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}
		switch declaration.Kind {
		case ast.KindTypeAliasDeclaration:
			return classifyArrayTypeNode(
				ctx, declaration.AsTypeAliasDeclaration().Type, targetNames, nonTargetNames, visitedSymbols,
			)
		case ast.KindTypeParameter:
			return classifyArrayTypeNode(
				ctx, declaration.AsTypeParameterDeclaration().Constraint, targetNames, nonTargetNames, visitedSymbols,
			)
		case ast.KindInterfaceDeclaration:
			return classifyArrayInterfaceNode(
				ctx, declaration, targetNames, nonTargetNames, visitedSymbols,
			)
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return arrayClassNonTarget
		case ast.KindImportClause, ast.KindImportSpecifier, ast.KindNamespaceImport:
			importedName := name
			if declaration.Kind == ast.KindImportSpecifier {
				if propertyName := declaration.PropertyName(); propertyName != nil && ast.IsIdentifier(propertyName) {
					importedName = propertyName.AsIdentifier().Text
				}
			}
			if targetNames.Has(name) || nonTargetNames.Has(name) ||
				targetNames.Has(importedName) || nonTargetNames.Has(importedName) {
				return arrayClassNonTarget
			}
			return arrayClassUnknown
		}
	}
	return arrayClassUnknown
}

func classifyArrayInterfaceNode(
	ctx rule.RuleContext,
	node *ast.Node,
	targetNames, nonTargetNames *utils.Set[string],
	visitedSymbols map[*ast.Symbol]bool,
) arrayClass {
	declaration := node.AsInterfaceDeclaration()
	if declaration.HeritageClauses == nil || len(declaration.HeritageClauses.Nodes) == 0 {
		return arrayClassNonTarget
	}
	classes := make([]arrayClass, 0, len(declaration.HeritageClauses.Nodes))
	for _, clauseNode := range declaration.HeritageClauses.Nodes {
		clause := clauseNode.AsHeritageClause()
		if clause == nil || clause.Token != ast.KindExtendsKeyword || clause.Types == nil {
			continue
		}
		for _, heritageNode := range clause.Types.Nodes {
			heritage := heritageNode.AsExpressionWithTypeArguments()
			if heritage == nil {
				classes = append(classes, arrayClassUnknown)
				continue
			}
			classes = append(classes, classifyArrayTypeReference(
				ctx, heritage.Expression, targetNames, nonTargetNames, visitedSymbols,
			))
		}
	}
	if len(classes) == 0 {
		return arrayClassNonTarget
	}
	return combineArrayClassesIntersection(classes)
}

func classifyKnownArrayTypeName(name string, targetNames, nonTargetNames *utils.Set[string]) arrayClass {
	if targetNames.Has(name) {
		return arrayClassTarget
	}
	if nonTargetNames.Has(name) {
		return arrayClassNonTarget
	}
	return arrayClassUnknown
}

func combineArrayClassesUnion(classes []arrayClass) arrayClass {
	if len(classes) == 0 {
		return arrayClassNonTarget
	}
	allTarget := true
	allNonTarget := true
	for _, class := range classes {
		allTarget = allTarget && class == arrayClassTarget
		allNonTarget = allNonTarget && class == arrayClassNonTarget
	}
	if allTarget {
		return arrayClassTarget
	}
	if allNonTarget {
		return arrayClassNonTarget
	}
	return arrayClassUnknown
}

func combineArrayClassesIntersection(classes []arrayClass) arrayClass {
	allNonTarget := true
	for _, class := range classes {
		if class == arrayClassTarget {
			return arrayClassTarget
		}
		allNonTarget = allNonTarget && class == arrayClassNonTarget
	}
	if allNonTarget {
		return arrayClassNonTarget
	}
	return arrayClassUnknown
}

func isSyntacticArrayNode(node *ast.Node) bool {
	if node.Kind == ast.KindArrayLiteralExpression {
		return true
	}
	if ast.IsCallExpression(node) {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee == nil {
			return false
		}
		if ast.IsPropertyAccessExpression(callee) && !ast.IsOptionalChain(callee) {
			propertyAccess := callee.AsPropertyAccessExpression()
			object := ast.SkipParentheses(propertyAccess.Expression)
			name := propertyAccess.Name()
			if object != nil && ast.IsIdentifier(object) && object.AsIdentifier().Text == "Array" &&
				name != nil && ast.IsIdentifier(name) {
				method := name.AsIdentifier().Text
				return method == "from" || method == "of"
			}
			return false
		}
		return ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "Array"
	}
	if ast.IsNewExpression(node) {
		callee := ast.SkipParentheses(node.AsNewExpression().Expression)
		return callee != nil && ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "Array"
	}
	return false
}

func isSyntacticTargetConstructor(node *ast.Node, targetNames *utils.Set[string]) bool {
	if node == nil || !ast.IsNewExpression(node) {
		return false
	}
	callee := ast.SkipParentheses(node.AsNewExpression().Expression)
	if callee == nil || !ast.IsIdentifier(callee) {
		return false
	}
	name := callee.AsIdentifier().Text
	return name == "Array" || targetNames.Has(name) && knownNonArrayNames.Has(name)
}

func isSyntacticNonArrayNode(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindNewExpression,
		ast.KindObjectLiteralExpression,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindClassExpression,
		ast.KindTemplateExpression,
		ast.KindNoSubstitutionTemplateLiteral:
		return true
	}
	return false
}

// classifyArrayType mirrors unicorn's getTypeScriptType by delegating to the
// shared type classifier. It takes no non-target name set: upstream ends with
// targetTypeNames.has(typeName) ? target : nonTarget, so a name that is not a
// target is already non-target. Name-based non-target matching belongs to the
// syntactic path, in classifyKnownArrayTypeName.
func classifyArrayType(ctx rule.RuleContext, t *checker.Type, targetNames *utils.Set[string]) arrayClass {
	return ClassifyType(ctx, t, TypeClassifierOptions{
		TargetTypeNames: targetNames,
		IsTargetType: func(t *checker.Type) bool {
			return targetNames.Has("Array") &&
				checker.Checker_isArrayOrTupleType(ctx.TypeChecker, t)
		},
	})
}
