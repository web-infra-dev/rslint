package unicornutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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
type sourceOnlyArrayReceiverStaticEvaluatorFileCacheKey struct{}
type sourceOnlyArrayReceiverClassificationFileCacheKey struct{}

type sourceOnlyArrayReceiverClassificationCache struct {
	symbols map[*ast.Symbol]sourceOnlyArrayReceiverClassificationEntry
}

type sourceOnlyArrayReceiverClassificationEntry struct {
	class           arrayClass
	additionalDepth int
	terminal        bool
}

type arrayReceiverClassifier struct {
	ctx                  rule.RuleContext
	targetNames          *utils.Set[string]
	nonTargetNames       *utils.Set[string]
	staticEvaluator      *utils.StaticStringEvaluator
	allowTypeInformation bool
	sourceOnlySemantics  bool
	recursionDepth       int
	recursionPeak        int
	recursionExhausted   bool
	sourceOnlyCache      *sourceOnlyArrayReceiverClassificationCache
}

const maxSourceOnlyArrayReceiverDepth = 1 << 10

// classifyArrayReceiver mirrors the syntactic short-circuits of unicorn's
// getType (array literals, Array()/new Array()/Array.from/Array.of), then falls
// back to a type-checker classification analogous to getTypeScriptType.
func classifyArrayReceiver(ctx rule.RuleContext, node *ast.Node, targetNames, nonTargetNames *utils.Set[string]) arrayClass {
	classifier := arrayReceiverClassifier{
		ctx:                  ctx,
		targetNames:          targetNames,
		nonTargetNames:       nonTargetNames,
		staticEvaluator:      arrayReceiverStaticEvaluator(ctx),
		allowTypeInformation: true,
	}
	return classifier.classify(node, map[*ast.Symbol]bool{})
}

func classifySourceOnlyIndexedCollectionReceiver(ctx rule.RuleContext, node *ast.Node) arrayClass {
	classifier := arrayReceiverClassifier{
		ctx:                 ctx,
		targetNames:         indexedCollectionTargets,
		nonTargetNames:      keyedCollectionNames,
		staticEvaluator:     sourceOnlyArrayReceiverStaticEvaluator(ctx),
		sourceOnlySemantics: true,
		sourceOnlyCache:     sourceOnlyArrayReceiverClassificationCacheForFile(ctx),
	}
	return classifier.classify(node, map[*ast.Symbol]bool{})
}

func (classifier *arrayReceiverClassifier) classify(node *ast.Node, visitedSymbols map[*ast.Symbol]bool) arrayClass {
	if classifier.sourceOnlySemantics {
		if classifier.recursionExhausted {
			return arrayClassUnknown
		}
		if classifier.recursionDepth >= maxSourceOnlyArrayReceiverDepth {
			// Treat exhaustion as terminal for this receiver. Otherwise an outer
			// conditional can retry a shorter suffix with the static evaluator
			// while unwinding and effectively segment an unbounded chain.
			classifier.recursionExhausted = true
			return arrayClassUnknown
		}
		classifier.recursionDepth++
		classifier.recursionPeak = max(classifier.recursionPeak, classifier.recursionDepth)
		defer func() { classifier.recursionDepth-- }()
	}
	if node == nil {
		return arrayClassUnknown
	}
	parenthesizedOptionalAccess := classifier.sourceOnlySemantics && node.Kind == ast.KindParenthesizedExpression &&
		func() bool {
			inner := ast.SkipParentheses(node)
			return inner != nil && ast.IsAccessExpression(inner) && ast.IsOptionalChain(inner)
		}()
	node = ast.SkipParentheses(node)
	if node == nil {
		return arrayClassUnknown
	}

	// Unicorn deliberately ignores a satisfies annotation and classifies the
	// expression itself. Assertions are different: an informative annotation
	// wins, while any / unknown falls through to the wrapped expression.
	if node.Kind == ast.KindSatisfiesExpression {
		return classifier.classify(node.AsSatisfiesExpression().Expression, visitedSymbols)
	}
	if node.Kind == ast.KindAsExpression || node.Kind == ast.KindTypeAssertionExpression {
		if class := classifyArrayTypeNode(
			classifier.ctx, node.Type(), classifier.targetNames, classifier.nonTargetNames, map[*ast.Symbol]bool{},
		); class != arrayClassUnknown {
			return class
		}
		return classifier.classify(node.Expression(), visitedSymbols)
	}
	if node.Kind == ast.KindNonNullExpression {
		return classifier.classify(node.Expression(), visitedSymbols)
	}
	// An array literal is a target even when one of its elements cannot be
	// folded. Taking this shape shortcut before static evaluation also avoids
	// rebuilding deeply nested literal aliases just to rediscover the outer
	// array kind.
	if classifier.sourceOnlySemantics && node.Kind == ast.KindArrayLiteralExpression {
		return arrayClassTarget
	}

	if ast.IsIdentifier(node) {
		if class := classifier.classifyIdentifier(node, visitedSymbols); class != arrayClassUnknown {
			return class
		}
		if classifier.recursionExhausted {
			return arrayClassUnknown
		}
	}
	if node.Kind == ast.KindConditionalExpression {
		conditional := node.AsConditionalExpression()
		class := combineArrayClassesUnion([]arrayClass{
			classifier.classify(conditional.WhenTrue, visitedSymbols),
			classifier.classify(conditional.WhenFalse, visitedSymbols),
		})
		if classifier.recursionExhausted {
			return arrayClassUnknown
		}
		if !classifier.sourceOnlySemantics || class != arrayClassUnknown {
			return class
		}
	}
	if ast.IsBinaryExpression(node) &&
		node.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
		class := classifier.classify(node.AsBinaryExpression().Right, visitedSymbols)
		if classifier.recursionExhausted {
			return arrayClassUnknown
		}
		if !classifier.sourceOnlySemantics || class != arrayClassUnknown {
			return class
		}
	}
	// getStaticValue runs before Unicorn's syntactic target/non-target checks.
	// A known array member is still a target; a known non-array member stays
	// unknown, matching upstream's getStaticType special case.
	if !ast.IsIdentifier(node) && (classifier.sourceOnlySemantics || !ast.IsAccessExpression(node)) {
		if isArray, known := classifier.staticEvaluator.EvalArrayValue(node); known {
			if isArray {
				return arrayClassTarget
			}
			if !classifier.sourceOnlySemantics || !ast.IsAccessExpression(node) || parenthesizedOptionalAccess {
				return arrayClassNonTarget
			}
		}
	}

	if isSyntacticArrayNode(node) {
		return arrayClassTarget
	}
	if isSyntacticTargetConstructor(node, classifier.targetNames) {
		return arrayClassTarget
	}

	if classifier.allowTypeInformation && classifier.ctx.TypeChecker != nil {
		t := utils.GetConstrainedTypeAtLocation(classifier.ctx.TypeChecker, node)
		if class := classifyArrayType(classifier.ctx, t, classifier.targetNames); class != arrayClassUnknown {
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

func sourceOnlyArrayReceiverStaticEvaluator(ctx rule.RuleContext) *utils.StaticStringEvaluator {
	return rule.CachedByFile(ctx, sourceOnlyArrayReceiverStaticEvaluatorFileCacheKey{}, func() *utils.StaticStringEvaluator {
		return utils.NewStaticStringEvaluatorForSourceOnlyStaticValue(
			ctx.SourceFile, sourceOnlyStaticReferenceResolver{refs: ctx.Refs, globals: ctx.Globals},
		)
	})
}

func sourceOnlyArrayReceiverClassificationCacheForFile(
	ctx rule.RuleContext,
) *sourceOnlyArrayReceiverClassificationCache {
	return rule.CachedByFile(ctx, sourceOnlyArrayReceiverClassificationFileCacheKey{}, func() *sourceOnlyArrayReceiverClassificationCache {
		return &sourceOnlyArrayReceiverClassificationCache{
			symbols: map[*ast.Symbol]sourceOnlyArrayReceiverClassificationEntry{},
		}
	})
}

type sourceOnlyStaticReferenceResolver struct {
	refs    *rule.RefStore
	globals rule.Globals
}

func (resolver sourceOnlyStaticReferenceResolver) Resolve(node *ast.Node) *ast.Symbol {
	if resolver.refs == nil {
		return nil
	}
	return resolver.refs.ResolveInFile(node)
}

func (resolver sourceOnlyStaticReferenceResolver) IsUnshadowedGlobal(node *ast.Node, name string) bool {
	return resolver.refs != nil && node != nil && ast.IsIdentifier(node) &&
		node.AsIdentifier().Text == name && resolver.globals.Access(name).IsDeclared() &&
		!resolver.refs.IsDefinedInFile(node)
}

func (classifier *arrayReceiverClassifier) classifyIdentifier(
	idNode *ast.Node,
	visitedSymbols map[*ast.Symbol]bool,
) (class arrayClass) {
	if classifier.ctx.Refs == nil || idNode == nil {
		return arrayClassUnknown
	}
	symbol := classifier.ctx.Refs.ResolveInFile(idNode)
	if symbol == nil {
		return arrayClassUnknown
	}
	if visitedSymbols[symbol] {
		return arrayClassUnknown
	}
	memoizeClassification := classifier.sourceOnlySemantics && classifier.sourceOnlyCache != nil
	if memoizeClassification {
		if cached, exists := classifier.sourceOnlyCache.symbols[symbol]; exists {
			if class, replay := classifier.replaySourceOnlyClassification(cached); replay {
				return class
			}
		}
		startDepth := classifier.recursionDepth
		previousPeak := classifier.recursionPeak
		classifier.recursionPeak = startDepth
		defer func() {
			localPeak := classifier.recursionPeak
			classifier.recursionPeak = max(previousPeak, localPeak)
			// Unknown is a stable, conservative result for an immutable symbol
			// too. Publishing only after the initializer has been classified keeps
			// cycles out of the lookup path, while memoizing unknown prevents a
			// shared conditional dependency from expanding exponentially.
			entry := sourceOnlyArrayReceiverClassificationEntry{
				class:           class,
				additionalDepth: localPeak - startDepth,
			}
			if classifier.recursionExhausted {
				// The rejected next node proves that this symbol needs at least one
				// level beyond the observed peak. Deeper contexts can replay the
				// terminal unknown in O(1); shallower contexts recompute and replace
				// it with either an exact result or a stronger lower bound.
				entry.class = arrayClassUnknown
				entry.additionalDepth++
				entry.terminal = true
			}
			classifier.sourceOnlyCache.symbols[symbol] = entry
		}()
	}
	if len(symbol.Declarations) != 1 {
		return arrayClassUnknown
	}
	visitedSymbols[symbol] = true
	defer delete(visitedSymbols, symbol)

	declaration := symbol.Declarations[0]
	if declaration == nil {
		return arrayClassUnknown
	}
	if annotation := arrayBindingTypeAnnotation(declaration); annotation != nil {
		if class := classifyArrayTypeNode(
			classifier.ctx, annotation, classifier.targetNames, classifier.nonTargetNames, map[*ast.Symbol]bool{},
		); class != arrayClassUnknown {
			return class
		}
	}
	if !ast.IsVariableDeclaration(declaration) {
		return arrayClassUnknown
	}
	declarationList := declaration.Parent
	variable := declaration.AsVariableDeclaration()
	if declarationList == nil || !ast.IsVariableDeclarationList(declarationList) ||
		declarationList.Flags&ast.NodeFlagsConst == 0 ||
		classifier.sourceOnlySemantics && (ast.IsVarUsing(declarationList) || ast.IsVarAwaitUsing(declarationList)) ||
		variable.Name() == nil ||
		!ast.IsIdentifier(variable.Name()) || variable.Initializer == nil {
		return arrayClassUnknown
	}
	return classifier.classify(variable.Initializer, visitedSymbols)
}

func (classifier *arrayReceiverClassifier) replaySourceOnlyClassification(
	entry sourceOnlyArrayReceiverClassificationEntry,
) (arrayClass, bool) {
	projectedDepth := classifier.recursionDepth + entry.additionalDepth
	if projectedDepth > maxSourceOnlyArrayReceiverDepth {
		// Recreate the terminal state and peak that produced this conservative
		// result. Otherwise an outer symbol can store an undersized lower bound,
		// or a conditional/comma expression can bypass the guard via static eval.
		classifier.recursionPeak = max(classifier.recursionPeak, maxSourceOnlyArrayReceiverDepth)
		classifier.recursionExhausted = true
		return arrayClassUnknown, true
	}
	if entry.terminal {
		return arrayClassUnknown, false
	}
	classifier.recursionPeak = max(classifier.recursionPeak, projectedDepth)
	return entry.class, true
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
