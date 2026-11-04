package unicornutil

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IsString proves that an expression produces a string primitive. Unlike
// !IsKnownNonString, unknown values are rejected. Static folding shares the
// caller's evaluator so mutation and side-effect checks retain their policy.
func IsString(ctx rule.RuleContext, node *ast.Node, evaluator *utils.StaticStringEvaluator) bool {
	// This predicate proves source-level behavior, including authored TS types;
	// ambient inference must not turn mutable values into string proofs.
	ctx.TypeChecker = nil
	return isString(ctx, node, evaluator, &stringReceiverWalkState{
		visiting: make(map[*ast.Symbol]bool), memo: make(map[*ast.Symbol]TypeClass),
	})
}

// IsStringWithTypeInformation extends source proofs with project types,
// preserving local runtime bindings and the checker's flow narrowing.
func IsStringWithTypeInformation(ctx rule.RuleContext, node *ast.Node, evaluator *utils.StaticStringEvaluator) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if IsString(ctx, node, evaluator) {
		return true
	}
	if ctx.TypeChecker == nil || node == nil || classifyRequiredStringType(ctx, node) != TypeTarget {
		return false
	}
	return !IsKnownNonStringSyntax(ctx, node) &&
		!hasConflictingStringBinding(ctx, node, evaluator, &stringReceiverWalkState{
			visiting: make(map[*ast.Symbol]bool), memo: make(map[*ast.Symbol]TypeClass),
		})
}

func hasConflictingStringBinding(ctx rule.RuleContext, node *ast.Node, evaluator *utils.StaticStringEvaluator, state *stringReceiverWalkState) (conflicting bool) {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	if state.depth >= maxStringReceiverDepth || state.steps >= maxStringReceiverSteps {
		return true
	}
	state.depth++
	state.steps++
	defer func() { state.depth-- }()
	recur := func(child *ast.Node) bool { return hasConflictingStringBinding(ctx, child, evaluator, state) }
	if ast.IsIdentifier(node) && ctx.Refs != nil {
		symbol := ctx.Refs.ResolveInFile(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		// Script declarations can collide with lib globals. A proof about the
		// ambient constructor must not override the local runtime binding.
		resolved := ctx.TypeChecker.GetSymbolAtLocation(node)
		if resolved == nil || (resolved != symbol && resolved != symbol.ExportSymbol) {
			return true
		}
		if classifyRequiredStringType(ctx, node) == TypeTarget {
			// The checker has a stronger proof at this reference than at the
			// declaration. Initializer writes and JSDoc do not invalidate it.
			if classifyRequiredStringCheckerType(ctx, ctx.TypeChecker.GetTypeOfSymbol(symbol)) != TypeTarget {
				return false
			}
			if annotation, _ := stringValueAnnotation(ctx, symbol.Declarations[0]); annotation != nil && utils.IsJSDocSyntaxNode(annotation) && !isStringValueType(ctx, annotation) {
				return true
			}
		}
		if cached, known := state.memo[symbol]; known {
			return cached == TypeNonTarget
		}
		if state.visiting[symbol] {
			return true
		}
		state.visiting[symbol] = true
		defer func() {
			delete(state.visiting, symbol)
			state.memo[symbol] = TypeTarget
			if conflicting {
				state.memo[symbol] = TypeNonTarget
			}
		}()
		declaration := symbol.Declarations[0]
		if ast.IsVariableDeclaration(declaration) && declaration.AsVariableDeclaration().Initializer != nil {
			if !ast.IsVarConst(declaration.Parent) {
				if _, stable := evaluator.ResolveIdentifierInitializer(node); !stable {
					return true
				}
			}
			return recur(declaration.AsVariableDeclaration().Initializer)
		}
	}
	switch node.Kind {
	case ast.KindCallExpression:
		callee := utils.ESTreeCallCallee(node.Expression())
		if callee != nil && ast.IsIdentifier(callee) && ctx.Refs != nil {
			if symbol := ctx.Refs.ResolveInFile(callee); symbol != nil {
				for _, reference := range ctx.Refs.References(symbol) {
					if !ast.IsDeclarationName(reference) && utils.IsWriteReference(reference) {
						return true
					}
				}
			}
		}
		return recur(callee)
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return recur(node.Expression())
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return recur(conditional.WhenTrue) || recur(conditional.WhenFalse)
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindCommaToken || binary.OperatorToken.Kind == ast.KindEqualsToken {
			return recur(binary.Right)
		}
		return recur(binary.Left) || recur(binary.Right)
	case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression, ast.KindSatisfiesExpression:
		return recur(node.Expression())
	}
	return false
}

// Unicorn allows an unmodified let/var initializer as direct string evidence,
// while its recursive type proof uses the stricter control-flow evaluator.
func isString(ctx rule.RuleContext, node *ast.Node, evaluator *utils.StaticStringEvaluator, state *stringReceiverWalkState) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	if isStringValue(ctx, node, evaluator, state) {
		return true
	}
	if node.Kind == ast.KindSatisfiesExpression {
		annotation := node.Type()
		if annotation.Kind == ast.KindStringKeyword || (annotation.Kind == ast.KindLiteralType && ast.IsStringLiteralLike(annotation.AsLiteralTypeNode().Literal)) {
			return true
		}
	}
	if !ast.IsIdentifier(node) || ctx.Refs == nil {
		return false
	}
	symbol := ctx.Refs.ResolveInFile(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Kind != ast.KindVariableDeclarationList ||
		ast.IsVarConst(declaration.Parent) || ast.IsVarUsing(declaration.Parent) || ast.IsVarAwaitUsing(declaration.Parent) {
		return false
	}
	value, known := evaluator.EvalValueIfNoSideEffects(node)
	_, stringValue := value.(string)
	return known && stringValue && !hasStringJSDocReference(ctx, node, make(map[*ast.Symbol]bool))
}

func isStringValue(ctx rule.RuleContext, node *ast.Node, evaluator *utils.StaticStringEvaluator, state *stringReceiverWalkState) (stringResult bool) {
	if node == nil || state.depth >= maxStringReceiverDepth || state.steps >= maxStringReceiverSteps {
		return false
	}
	state.depth++
	state.steps++
	defer func() { state.depth-- }()
	node = utils.ESTreeRuntimeExpression(node)
	recur := func(child *ast.Node) bool { return isStringValue(ctx, child, evaluator, state) }
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression, ast.KindTypeOfExpression:
		return true
	case ast.KindSatisfiesExpression:
		return recur(node.Expression())
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		if isStringValueType(ctx, node.Type()) {
			return true
		}
		return recur(node.Expression())
	case ast.KindNonNullExpression:
		return recur(node.Expression())
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		if recur(conditional.WhenTrue) && recur(conditional.WhenFalse) {
			return true
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken:
			return isString(ctx, binary.Right, evaluator, state)
		case ast.KindCommaToken:
			return recur(binary.Right)
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			if recur(binary.Left) && recur(binary.Right) {
				return true
			}
		case ast.KindPlusToken, ast.KindPlusEqualsToken:
			return isString(ctx, binary.Left, evaluator, state) || isString(ctx, binary.Right, evaluator, state)
		}
	case ast.KindIdentifier:
		if ctx.Refs != nil {
			symbol := ctx.Refs.ResolveInFile(node)
			if symbol != nil && len(symbol.Declarations) == 1 && !state.visiting[symbol] {
				if cached, known := state.memo[symbol]; known {
					return cached == TypeTarget
				}
				state.visiting[symbol] = true
				defer func() {
					delete(state.visiting, symbol)
					state.memo[symbol] = TypeUnknown
					if stringResult {
						state.memo[symbol] = TypeTarget
					}
				}()
				declaration := symbol.Declarations[0]

				annotation, optional := stringValueAnnotation(ctx, declaration)
				if optional {
					return false
				}
				if annotation != nil {
					if isStringValueType(ctx, annotation) {
						return true
					}
					if declaration.Type() == nil || utils.IsJSDocSyntaxNode(annotation) {
						return false
					}
				}
				if declaration.Kind == ast.KindVariableDeclaration && ast.IsVarConst(declaration.Parent) {
					return recur(declaration.AsVariableDeclaration().Initializer)
				}
			}
		}
	case ast.KindCallExpression:
		callee := utils.ESTreeCallCallee(node.Expression())
		if callee == nil || ast.IsOptionalChain(node) {
			return false
		}
		if ast.IsIdentifier(callee) {
			switch callee.Text() {
			case "String", "Date", "encodeURI", "encodeURIComponent", "decodeURI", "decodeURIComponent", "atob", "btoa":
				if IsGlobalReference(ctx, callee) {
					return true
				}
			}
			return classifyStringFunctionReturn(ctx, callee) == TypeTarget
		}
		if call, ok := MatchDotMethodCall(node, DotMethodCallOptions{}); ok {
			object := utils.ESTreeRuntimeExpression(call.Object)
			method := call.Property.Text()
			if ast.IsIdentifier(object) && object.Text() == "String" && IsGlobalReference(ctx, object) &&
				(method == "fromCharCode" || method == "fromCodePoint" || method == "raw") {
				return true
			}
			switch method {
			case "charAt", "concat", "normalize", "padEnd", "padStart", "repeat", "slice", "substring", "toLocaleLowerCase", "toLocaleUpperCase", "toLowerCase", "toString", "toUpperCase", "toWellFormed", "trim", "trimEnd", "trimStart", "valueOf":
				if isString(ctx, object, evaluator, state) {
					return true
				}
			}
			if method == "join" && classifyArrayReceiver(ctx, object, indexedCollectionTargets, keyedCollectionNames) == arrayClassTarget {
				return true
			}
			if isStringBuiltinValue(ctx, object, "Date", state.visiting) {
				switch method {
				case "toDateString", "toISOString", "toLocaleDateString", "toLocaleString", "toLocaleTimeString", "toString", "toTimeString", "toUTCString":
					return true
				}
			}
		}
	case ast.KindPropertyAccessExpression:
		if !ast.IsOptionalChain(node) && node.Name().Kind == ast.KindIdentifier && (node.Name().Text() == "source" || node.Name().Text() == "flags") {
			return isStringBuiltinValue(ctx, node.Expression(), "RegExp", state.visiting)
		}
	case ast.KindTaggedTemplateExpression:
		tag := utils.ESTreeRuntimeExpression(node.AsTaggedTemplateExpression().Tag)
		if tag.Kind == ast.KindPropertyAccessExpression && !ast.IsOptionalChain(tag) && tag.Name().Kind == ast.KindIdentifier && tag.Name().Text() == "raw" {
			object := utils.ESTreeRuntimeExpression(tag.Expression())
			if ast.IsIdentifier(object) && object.Text() == "String" && IsGlobalReference(ctx, object) {
				return true
			}
		}
	}
	value, known := evaluator.EvalControlFlowValue(node)
	_, stringValue := value.(string)
	return known && stringValue && !hasStringJSDocReference(ctx, node, make(map[*ast.Symbol]bool))
}

// These built-ins have string-valued members even without type services.
// Reuse the caller's symbol guard while following const aliases and authored
// annotations; arbitrary properties and subclass overrides remain unknown.
func isStringBuiltinValue(ctx rule.RuleContext, node *ast.Node, name string, visiting map[*ast.Symbol]bool) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	if node.Kind == ast.KindRegularExpressionLiteral {
		return name == "RegExp"
	}
	if node.Kind == ast.KindNewExpression || (name == "RegExp" && node.Kind == ast.KindCallExpression && !ast.IsOptionalChain(node)) {
		callee := utils.ESTreeRuntimeExpression(node.Expression())
		return ast.IsIdentifier(callee) && callee.Text() == name && IsGlobalReference(ctx, callee)
	}
	if !ast.IsIdentifier(node) || ctx.Refs == nil {
		return false
	}
	symbol := ctx.Refs.ResolveInFile(node)
	if symbol == nil || len(symbol.Declarations) != 1 || visiting[symbol] {
		return false
	}
	visiting[symbol] = true
	defer delete(visiting, symbol)
	declaration := symbol.Declarations[0]
	annotation, optional := stringValueAnnotation(ctx, declaration)
	if optional {
		return false
	}
	if annotation != nil {
		if annotation.Kind != ast.KindTypeReference {
			return false
		}
		typeName := annotation.AsTypeReferenceNode().TypeName
		return ast.IsIdentifier(typeName) && typeName.Text() == name &&
			ctx.Refs.IsGlobalNameReference(typeName, name, ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias)
	}
	return declaration.Kind == ast.KindVariableDeclaration && ast.IsVarConst(declaration.Parent) &&
		isStringBuiltinValue(ctx, declaration.AsVariableDeclaration().Initializer, name, visiting)
}

// Read authored annotations through tsgo's AST, retaining JSDoc unknown types
// and optional parameter metadata so static folding cannot bypass them.
func stringValueAnnotation(ctx rule.RuleContext, declaration *ast.Node) (annotation *ast.Node, optional bool) {
	annotation = ast.GetTypeAnnotationNode(declaration)
	optional = ast.HasQuestionToken(declaration)
	if annotation != nil && (declaration.Kind != ast.KindParameter || !utils.IsJSDocSyntaxNode(annotation)) {
		return annotation, optional
	}
	host := declaration
	if declaration.Kind == ast.KindParameter {
		host = declaration.Parent
	}
	for host != nil {
		for _, doc := range host.JSDoc(ctx.SourceFile) {
			if tags := doc.AsJSDoc().Tags; tags != nil {
				for _, tag := range tags.Nodes {
					var expression *ast.Node
					switch tag.Kind {
					case ast.KindJSDocParameterTag:
						parameter := tag.AsJSDocParameterOrPropertyTag()
						if declaration.Kind == ast.KindParameter && ast.IsIdentifier(declaration.Name()) && parameter.Name().Text() == declaration.Name().Text() {
							expression = parameter.TypeExpression
							optional = optional || parameter.IsBracketed
						}
					case ast.KindJSDocTypeTag:
						if declaration.Kind == ast.KindVariableDeclaration {
							expression = tag.AsJSDocTypeTag().TypeExpression
						}
					case ast.KindJSDocReturnTag:
						if ast.IsFunctionLike(declaration) {
							expression = tag.AsJSDocReturnTag().TypeExpression
						}
					}
					if expression != nil {
						if annotation == nil {
							annotation = expression.Type()
						}
						return annotation, optional
					}
				}
			}
		}
		host = ast.GetNextJSDocCommentLocation(host)
	}
	return annotation, optional
}

func classifyStringCallReturnType(ctx rule.RuleContext, annotation *ast.Node) TypeClass {
	if annotation == nil {
		return TypeUnknown
	}
	annotation = ast.SkipTypeParentheses(annotation)
	switch annotation.Kind {
	case ast.KindFunctionType:
		return classifyStringValueType(ctx, annotation.Type())
	case ast.KindTypeLiteral:
		var classes []TypeClass
		for _, member := range annotation.AsTypeLiteralNode().Members.Nodes {
			if member.Kind == ast.KindCallSignature {
				classes = append(classes, classifyStringValueType(ctx, member.Type()))
			}
		}
		if len(classes) > 0 {
			return combineStringUnion(classes)
		}
	case ast.KindUnionType:
		var classes []TypeClass
		for _, member := range annotation.AsUnionTypeNode().Types.Nodes {
			classes = append(classes, classifyStringCallReturnType(ctx, member))
		}
		if len(classes) > 0 {
			return combineStringUnion(classes)
		}
	case ast.KindIntersectionType:
		var classes []TypeClass
		var visit func(*ast.Node)
		visit = func(member *ast.Node) {
			member = ast.SkipTypeParentheses(member)
			if member.Kind == ast.KindIntersectionType {
				for _, part := range member.AsIntersectionTypeNode().Types.Nodes {
					visit(part)
				}
				return
			}
			if member.Kind == ast.KindTypeLiteral && !slices.ContainsFunc(member.AsTypeLiteralNode().Members.Nodes, func(part *ast.Node) bool { return part.Kind == ast.KindCallSignature }) {
				return
			}
			classes = append(classes, classifyStringCallReturnType(ctx, member))
		}
		visit(annotation)
		if len(classes) > 0 {
			return combineStringUnion(classes)
		}
	}
	return TypeUnknown
}

func isStringValueType(ctx rule.RuleContext, node *ast.Node) bool {
	return classifyStringValueType(ctx, node) == TypeTarget
}

func classifyStringValueType(ctx rule.RuleContext, node *ast.Node) TypeClass {
	if node != nil && node.Kind == ast.KindStringKeyword {
		return TypeTarget
	}
	state := newStringTypeWalkState()
	class := classifyStringTypeSyntax(ctx, node, state, true)
	// Unlike unknown, any absorbs a string intersection. Unresolved or bounded
	// walks also cannot prove that every intersected type excludes any.
	if class == TypeTarget && state.mayBeAny {
		return TypeUnknown
	}
	return class
}

type stringFunctionReturnCacheKey struct{}

// Calls through a local function use its authored return type. Native binding
// annotations, overloads and mutated bindings stay unknown, matching Unicorn's
// source-only function-call proof rather than inferring a body or signature.
func classifyStringFunctionReturn(ctx rule.RuleContext, callee *ast.Node) (class TypeClass) {
	if ctx.Refs == nil {
		return TypeUnknown
	}
	symbol := ctx.Refs.ResolveInFile(callee)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return TypeUnknown
	}
	cache := rule.CachedByFile(ctx, stringFunctionReturnCacheKey{}, func() map[*ast.Symbol]TypeClass {
		return make(map[*ast.Symbol]TypeClass)
	})
	if cached, known := cache[symbol]; known {
		return cached
	}
	defer func() { cache[symbol] = class }()
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindFunctionDeclaration && declaration.Type() != nil && !utils.IsJSDocSyntaxNode(declaration.Type()) {
		return TypeUnknown
	}
	for _, reference := range ctx.Refs.References(symbol) {
		if reference != declaration.Name() && utils.IsWriteReference(reference) {
			return TypeUnknown
		}
	}
	var function *ast.Node
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		function = declaration
	case ast.KindVariableDeclaration:
		initializer := utils.ESTreeRuntimeExpression(declaration.AsVariableDeclaration().Initializer)
		for initializer != nil && (initializer.Kind == ast.KindSatisfiesExpression || initializer.Kind == ast.KindNonNullExpression) {
			initializer = utils.ESTreeRuntimeExpression(initializer.Expression())
		}
		if initializer != nil && (initializer.Kind == ast.KindFunctionExpression || initializer.Kind == ast.KindArrowFunction) {
			function = initializer
		}
	}
	if function != nil {
		if ast.HasSyntacticModifier(function, ast.ModifierFlagsAsync) || len(function.TypeParameters()) > 0 {
			return TypeUnknown
		}
		if (function.Kind == ast.KindFunctionDeclaration && function.AsFunctionDeclaration().AsteriskToken != nil) ||
			(function.Kind == ast.KindFunctionExpression && function.AsFunctionExpression().AsteriskToken != nil) {
			return TypeUnknown
		}
		if annotation, _ := stringValueAnnotation(ctx, function); annotation != nil {
			return classifyStringValueType(ctx, annotation)
		}
	}
	annotation, optional := stringValueAnnotation(ctx, declaration)
	if optional {
		return TypeUnknown
	}
	if declaration.Kind == ast.KindParameter {
		raw := utils.ESTreeRuntimeExpression(declaration.AsParameterDeclaration().Initializer)
		value := utils.SkipAssertionsAndParens(raw)
		if raw != nil && (raw.Kind == ast.KindAsExpression || raw.Kind == ast.KindTypeAssertionExpression ||
			(value != nil && ast.IsFunctionLike(value) && (value.Type() != nil || ast.HasSyntacticModifier(value, ast.ModifierFlagsAsync)))) {
			return TypeUnknown
		}
	}
	return classifyStringCallReturnType(ctx, annotation)
}

// Static folding must not bypass an authored JSDoc type through an alias or a
// larger expression. Traverse runtime children using tsgo and the reference
// store, excluding property names, type syntax, and unevaluated function bodies.
func hasStringJSDocReference(ctx rule.RuleContext, node *ast.Node, visited map[*ast.Symbol]bool) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || ast.IsTypeNode(node) || ast.IsFunctionLike(node) || ast.IsClassLike(node) {
		return false
	}
	if ast.IsIdentifier(node) {
		if ctx.Refs == nil || utils.IsNonReferenceIdentifier(node) {
			return false
		}
		symbol := ctx.Refs.ResolveInFile(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if annotation, _ := stringValueAnnotation(ctx, declaration); annotation != nil && utils.IsJSDocSyntaxNode(annotation) {
			return true
		}
		if visited[symbol] {
			return false
		}
		visited[symbol] = true
		return declaration.Kind == ast.KindVariableDeclaration && hasStringJSDocReference(ctx, declaration.AsVariableDeclaration().Initializer, visited)
	}
	return node.ForEachChild(func(child *ast.Node) bool { return hasStringJSDocReference(ctx, child, visited) })
}
