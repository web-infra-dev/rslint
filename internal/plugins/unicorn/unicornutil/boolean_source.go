package unicornutil

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// These are source-level proofs used even when no TypeScript project is enabled.
// Reference resolution and JSDoc parsing remain with RefStore and unicornutil.
func booleanAnnotation(ctx rule.RuleContext, node *ast.Node, visiting map[*ast.Node]bool) bool {
	if node == nil || visiting[node] {
		return false
	}
	visiting[node] = true
	defer delete(visiting, node)
	switch node.Kind {
	case ast.KindBooleanKeyword:
		return true
	case ast.KindLiteralType:
		return ast.IsBooleanLiteral(node.AsLiteralTypeNode().Literal)
	case ast.KindTypePredicate:
		return node.AsTypePredicateNode().AssertsModifier == nil
	case ast.KindParenthesizedType, ast.KindJSDocNonNullableType:
		return booleanAnnotation(ctx, node.Type(), visiting)
	case ast.KindUnionType:
		for _, part := range node.AsUnionTypeNode().Types.Nodes {
			if !booleanAnnotation(ctx, part, visiting) {
				return false
			}
		}
		return true
	case ast.KindIntersectionType:
		if !utils.IsJSDocSyntaxNode(node) {
			return false
		}
		found := false
		members := slices.Clone(node.AsIntersectionTypeNode().Types.Nodes)
		for index := 0; index < len(members); index++ {
			part := ast.SkipTypeParentheses(members[index])
			for part.Kind == ast.KindJSDocNonNullableType {
				part = ast.SkipTypeParentheses(part.Type())
			}
			if part.Kind == ast.KindIntersectionType {
				members = append(members, part.AsIntersectionTypeNode().Types.Nodes...)
				continue
			}
			// Object-shaped brands do not change the primitive value. Callable
			// members and unknown types still have to prove a boolean.
			if part.Kind == ast.KindTypeLiteral && !slices.ContainsFunc(part.Members(), func(member *ast.Node) bool { return member.Kind == ast.KindCallSignature }) {
				continue
			}
			if part.Kind == ast.KindTypeReference {
				reference := part.AsTypeReferenceNode()
				if ast.IsIdentifier(reference.TypeName) && reference.TypeName.Text() == "Object" && reference.TypeArguments != nil && len(reference.TypeArguments.Nodes) > 0 {
					continue
				}
			}
			if !booleanAnnotation(ctx, part, visiting) {
				return false
			}
			found = true
		}
		return found
	case ast.KindTypeReference:
		if declaration := typeReferenceDeclaration(ctx, node); declaration != nil && declaration.Kind == ast.KindTypeAliasDeclaration {
			return booleanAnnotation(ctx, declaration.Type(), visiting)
		}
	}
	return false
}

func booleanCallableAnnotation(ctx rule.RuleContext, node *ast.Node, visiting map[*ast.Node]bool) bool {
	if node == nil || visiting[node] {
		return false
	}
	visiting[node] = true
	defer delete(visiting, node)
	switch node.Kind {
	case ast.KindParenthesizedType, ast.KindJSDocNonNullableType:
		return booleanCallableAnnotation(ctx, node.Type(), visiting)
	case ast.KindFunctionType:
		return booleanAnnotation(ctx, node.Type(), visiting)
	case ast.KindUnionType:
		for _, part := range node.AsUnionTypeNode().Types.Nodes {
			if !booleanCallableAnnotation(ctx, part, visiting) {
				return false
			}
		}
		return true
	case ast.KindTypeReference:
		if declaration := typeReferenceDeclaration(ctx, node); declaration != nil {
			if declaration.Kind == ast.KindTypeAliasDeclaration {
				return booleanCallableAnnotation(ctx, declaration.Type(), visiting)
			}
			if declaration.Kind == ast.KindInterfaceDeclaration {
				return booleanCallableAnnotation(ctx, declaration, visiting)
			}
		}
	case ast.KindTypeLiteral, ast.KindInterfaceDeclaration:
		found := false
		for _, member := range node.Members() {
			if member.Kind == ast.KindCallSignature {
				found = true
				if !booleanAnnotation(ctx, member.Type(), visiting) {
					return false
				}
			}
		}
		return found
	}
	return false
}

func typeReferenceDeclaration(ctx rule.RuleContext, node *ast.Node) *ast.Node {
	name := node.AsTypeReferenceNode().TypeName
	if !ast.IsIdentifier(name) || ctx.Refs == nil {
		return nil
	}
	symbol := ctx.Refs.ResolveInFileWithMeaning(name, ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias)
	if symbol != nil && len(symbol.Declarations) == 1 {
		return symbol.Declarations[0]
	}
	return nil
}

type callableWritesCacheKey struct{}

func callableHasWrites(ctx rule.RuleContext, symbol *ast.Symbol) bool {
	writes := rule.CachedByFile(ctx, callableWritesCacheKey{}, func() map[*ast.Symbol]bool {
		return make(map[*ast.Symbol]bool)
	})
	if hasWrites, cached := writes[symbol]; cached {
		return hasWrites
	}
	hasWrites := slices.ContainsFunc(ctx.Refs.References(symbol), utils.IsWriteReference)
	writes[symbol] = hasWrites
	return hasWrites
}

// A JSDoc callback annotation cannot make an async/generator initializer return
// a primitive boolean. Explicit assertions/return types on parameter defaults
// likewise take precedence over that parameter's JSDoc.
func overridesJSDocCallable(declaration *ast.Node) bool {
	if !ast.HasInitializer(declaration) {
		return false
	}
	raw := utils.ESTreeRuntimeExpression(declaration.Initializer())
	value := utils.SkipAssertionsAndParens(raw)
	if value != nil && ast.IsFunctionLike(value) {
		if ast.GetFunctionFlags(value)&(ast.FunctionFlagsAsync|ast.FunctionFlagsGenerator) != 0 {
			return true
		}
		if declaration.Kind == ast.KindParameter && value.Type() != nil {
			return true
		}
	}
	return declaration.Kind == ast.KindParameter && raw != nil && (raw.Kind == ast.KindAsExpression || raw.Kind == ast.KindTypeAssertionExpression)
}

// IsFunctionType reports whether an authored callable type returns booleans.
func (a *BooleanAnalyzer) IsFunctionType(annotation *ast.Node) bool {
	return booleanCallableAnnotation(*a.ctx, annotation, make(map[*ast.Node]bool))
}

// IsFunction reports whether a synchronous function returns a primitive boolean.
// result is its concise body or the expression of its only return statement.
func (a *BooleanAnalyzer) IsFunction(function, result *ast.Node) bool {
	if isAsyncOrGenerator(function) {
		return false
	}
	if annotation := CallableReturnTypeAnnotation(*a.ctx, function); annotation != nil {
		if booleanAnnotation(*a.ctx, annotation, make(map[*ast.Node]bool)) {
			return true
		}
		if utils.IsJSDocSyntaxNode(annotation) {
			return false
		}
	}
	return a.IsExpression(result)
}

// SingleReturnExpression returns the concise body or the value of the only
// statement when it is a return. Multi-statement bodies need flow analysis.
func SingleReturnExpression(function *ast.Node) *ast.Node {
	body := function.Body()
	if body == nil || !ast.IsBlock(body) {
		return body
	}
	if statements := body.Statements(); len(statements) == 1 && ast.IsReturnStatement(statements[0]) {
		return statements[0].Expression()
	}
	return nil
}

func (a *BooleanAnalyzer) booleanCall(node *ast.Node, visiting map[*ast.Node]bool) bool {
	if ast.IsOptionalChain(node) {
		return false
	}
	callee := utils.ESTreeCallCallee(node.Expression())
	if callee == nil {
		return false
	}
	if a.isBuiltinBooleanFunction(callee) {
		return true
	}
	ctx := *a.ctx
	if ast.IsAccessExpression(callee) {
		receiver := utils.ESTreeRuntimeExpression(callee.Expression())
		name, known := a.memberName(callee)
		if !known {
			return false
		}
		switch name {
		case "endsWith", "includes", "isWellFormed", "startsWith":
			if IsString(ctx, receiver, a.StaticEvaluator()) {
				return true
			}
		}
		switch name {
		case "every", "includes", "some":
			if IsIndexedCollection(ctx, receiver) {
				return true
			}
		case "test":
			return regexpReceiver(ctx, receiver, visiting)
		case "has":
			return builtinReceiver(ctx, receiver, visiting, "Map", "ReadonlyMap", "WeakMap", "URLSearchParams", "Headers", "FormData") ||
				builtinReceiver(ctx, receiver, visiting, "Set", "ReadonlySet") ||
				builtinReceiver(ctx, receiver, visiting, "WeakSet")
		case "delete":
			return builtinReceiver(ctx, receiver, visiting, "Map", "Set", "WeakMap", "WeakSet")
		case "isDisjointFrom", "isSubsetOf", "isSupersetOf":
			return builtinReceiver(ctx, receiver, visiting, "Set", "ReadonlySet")
		}
		return false
	}
	return a.isLocalBooleanFunction(callee, visiting)
}

// IsFunctionReference recognizes a local boolean-returning function or a known
// global predicate. It does not infer a comparator from its contextual type.
func (a *BooleanAnalyzer) IsFunctionReference(node *ast.Node) bool {
	return node != nil && (a.isBuiltinBooleanFunction(node) || a.isLocalBooleanFunction(node, make(map[*ast.Node]bool)))
}

func (a *BooleanAnalyzer) isLocalBooleanFunction(node *ast.Node, visiting map[*ast.Node]bool) bool {
	ctx := *a.ctx
	if !ast.IsIdentifier(node) || ctx.Refs == nil {
		return false
	}
	symbol := ctx.Refs.ResolveInFile(node)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	if len(symbol.Declarations) > 1 {
		if callableHasWrites(ctx, symbol) {
			return false
		}
		found := false
		for _, declaration := range symbol.Declarations {
			if !ast.IsFunctionDeclaration(declaration) || !stableFunctionDeclaration(ctx, declaration) {
				return false
			}
			if declaration.Body() == nil {
				found = true
				annotation, _ := booleanValueAnnotation(ctx, declaration)
				if !booleanAnnotation(ctx, annotation, visiting) {
					return false
				}
			}
		}
		return found
	}
	declaration := symbol.Declarations[0]
	annotation, optional := booleanValueAnnotation(ctx, declaration)
	if annotation != nil && !utils.IsJSDocSyntaxNode(annotation) && !ast.IsFunctionLike(declaration) {
		return booleanCallableAnnotation(ctx, annotation, visiting)
	}
	if callableHasWrites(ctx, symbol) || overridesJSDocCallable(declaration) {
		return false
	}
	function := sourceFunction(ctx, declaration)
	if function != nil && ast.IsFunctionLike(function) {
		if isAsyncOrGenerator(function) {
			return false
		}
		if result, _ := booleanValueAnnotation(ctx, function); result != nil {
			return booleanAnnotation(ctx, result, visiting)
		}
		if annotation == nil {
			return a.isBoolean(SingleReturnExpression(function), visiting)
		}
	}
	if optional {
		return false
	}
	return booleanCallableAnnotation(ctx, annotation, visiting)
}

var booleanStaticMethods = map[string][]string{
	"Array": {"isArray"}, "ArrayBuffer": {"isView"}, "Atomics": {"isLockFree"}, "Error": {"isError"},
	"Number":  {"isFinite", "isInteger", "isNaN", "isSafeInteger"},
	"Object":  {"hasOwn", "is", "isExtensible", "isFrozen", "isSealed"},
	"Reflect": {"defineProperty", "deleteProperty", "has", "isExtensible", "preventExtensions", "set", "setPrototypeOf"},
	"URL":     {"canParse"},
}

func (a *BooleanAnalyzer) isBuiltinBooleanFunction(node *ast.Node) bool {
	if ast.IsIdentifier(node) {
		switch node.Text() {
		case "Boolean", "isNaN", "isFinite":
			return a.isGlobal(node)
		}
		return false
	}
	if !ast.IsAccessExpression(node) || ast.IsOptionalChain(node) {
		return false
	}
	receiver := utils.ESTreeRuntimeExpression(node.Expression())
	if !ast.IsIdentifier(receiver) {
		return false
	}
	methods := booleanStaticMethods[receiver.Text()]
	if len(methods) == 0 {
		return false
	}
	name, known := a.memberName(node)
	return known && slices.Contains(methods, name) && a.isGlobal(receiver)
}

func (a *BooleanAnalyzer) isGlobal(node *ast.Node) bool {
	return a.ctx.Globals.Access(node.Text()).IsDeclared() && IsGlobalReference(*a.ctx, node)
}

func regexpReceiver(ctx rule.RuleContext, node *ast.Node, visiting map[*ast.Node]bool) bool {
	return builtinReceiver(ctx, node, visiting, "RegExp")
}

func stableFunctionDeclaration(ctx rule.RuleContext, declaration *ast.Node) bool {
	sourceType := ctx.LanguageOptions.EffectiveSourceType()
	return (sourceType != "script" || declaration.Parent.Kind != ast.KindSourceFile) &&
		utils.IsInStrictModeWithSourceType(declaration, ctx.SourceFile, sourceType)
}

// Parameter defaults do not describe callbacks supplied by the caller.
func sourceFunction(ctx rule.RuleContext, declaration *ast.Node) *ast.Node {
	if ast.IsFunctionDeclaration(declaration) && stableFunctionDeclaration(ctx, declaration) || declaration.Kind == ast.KindFunctionExpression {
		return declaration
	}
	if declaration.Kind == ast.KindVariableDeclaration {
		value := utils.ESTreeRuntimeExpression(declaration.Initializer())
		for value != nil && (value.Kind == ast.KindSatisfiesExpression || value.Kind == ast.KindNonNullExpression) {
			value = utils.ESTreeRuntimeExpression(value.Expression())
		}
		return value
	}
	return nil
}

func builtinReceiver(ctx rule.RuleContext, node *ast.Node, visiting map[*ast.Node]bool, names ...string) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || visiting[node] {
		return false
	}
	visiting[node] = true
	defer delete(visiting, node)
	if node.Kind == ast.KindRegularExpressionLiteral {
		return slices.Contains(names, "RegExp")
	}
	if node.Kind == ast.KindNewExpression {
		callee := utils.ESTreeRuntimeExpression(node.Expression())
		return callee != nil && ast.IsIdentifier(callee) && slices.Contains(names, callee.Text()) && IsGlobalReference(ctx, callee)
	}
	var annotation *ast.Node
	if ast.IsIdentifier(node) && ctx.Refs != nil {
		if symbol := ctx.Refs.ResolveInFile(node); symbol != nil && len(symbol.Declarations) == 1 {
			declaration := symbol.Declarations[0]
			var optional bool
			annotation, optional = booleanValueAnnotation(ctx, declaration)
			if optional {
				return false
			}
			if annotation == nil && declaration.Kind == ast.KindVariableDeclaration && ast.IsVarConst(declaration.Parent) {
				return builtinReceiver(ctx, declaration.Initializer(), visiting, names...)
			}
		}
	} else if node.Kind == ast.KindCallExpression && !ast.IsOptionalChain(node) && ctx.Refs != nil {
		callee := utils.ESTreeCallCallee(node.Expression())
		if callee != nil && ast.IsIdentifier(callee) {
			if symbol := ctx.Refs.ResolveInFile(callee); symbol != nil && len(symbol.Declarations) == 1 && !callableHasWrites(ctx, symbol) {
				declaration := symbol.Declarations[0]
				if function := sourceFunction(ctx, declaration); function != nil && ast.IsFunctionLike(function) && ast.GetFunctionFlags(function)&(ast.FunctionFlagsAsync|ast.FunctionFlagsGenerator) == 0 {
					annotation, _ = booleanValueAnnotation(ctx, function)
				}
			}
		}
	}
	return builtinAnnotation(ctx, annotation, visiting, names)
}

func builtinAnnotation(ctx rule.RuleContext, node *ast.Node, visiting map[*ast.Node]bool, names []string) bool {
	if node == nil || visiting[node] {
		return false
	}
	visiting[node] = true
	defer delete(visiting, node)
	switch node.Kind {
	case ast.KindParenthesizedType, ast.KindJSDocNonNullableType:
		return builtinAnnotation(ctx, node.Type(), visiting, names)
	case ast.KindUnionType:
		for _, part := range node.AsUnionTypeNode().Types.Nodes {
			if !builtinAnnotation(ctx, part, visiting, names) {
				return false
			}
		}
		return true
	case ast.KindTypeReference:
		if declaration := typeReferenceDeclaration(ctx, node); declaration != nil {
			return declaration.Kind == ast.KindTypeAliasDeclaration && builtinAnnotation(ctx, declaration.Type(), visiting, names)
		}
		name := node.AsTypeReferenceNode().TypeName
		return ast.IsIdentifier(name) && slices.Contains(names, name.Text())
	}
	return false
}

func (a *BooleanAnalyzer) memberName(node *ast.Node) (string, bool) {
	if node.Kind == ast.KindPropertyAccessExpression {
		return utils.AccessExpressionStaticName(node)
	}
	return a.StaticEvaluator().EvalAccessExpressionName(node)
}
