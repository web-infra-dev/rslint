package no_chained_comparison

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
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
		for _, part := range node.AsIntersectionTypeNode().Types.Nodes {
			if booleanAnnotation(ctx, part, visiting) {
				return true
			}
		}
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

func callableHasWrites(ctx rule.RuleContext, symbol *ast.Symbol) bool {
	for _, reference := range ctx.Refs.References(symbol) {
		if utils.IsWriteReference(reference) {
			return true
		}
	}
	return false
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

func booleanCall(ctx rule.RuleContext, node *ast.Node, evaluator *utils.StaticStringEvaluator, visiting map[*ast.Node]bool, completed map[*ast.Node]bool) bool {
	if ast.IsOptionalChain(node) {
		return false
	}
	callee := utils.ESTreeCallCallee(node.Expression())
	if callee == nil {
		return false
	}
	if ast.IsAccessExpression(callee) {
		receiver := utils.ESTreeRuntimeExpression(callee.Expression())
		name, known := memberName(callee, evaluator)
		if !known {
			return false
		}
		if ast.IsIdentifier(receiver) && unicornutil.IsGlobalReference(ctx, receiver) && slices.Contains(booleanStaticMethods[receiver.Text()], name) {
			return true
		}
		switch name {
		case "endsWith", "includes", "isWellFormed", "startsWith":
			if unicornutil.IsString(ctx, receiver, evaluator) {
				return true
			}
		}
		switch name {
		case "every", "includes", "some":
			if unicornutil.IsIndexedCollection(ctx, receiver) {
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
	if !ast.IsIdentifier(callee) {
		return false
	}
	if unicornutil.IsGlobalReference(ctx, callee) && (callee.Text() == "Boolean" || callee.Text() == "isNaN" || callee.Text() == "isFinite") {
		return true
	}
	if ctx.Refs == nil {
		return false
	}
	symbol := ctx.Refs.ResolveInFile(callee)
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
				annotation, _ := valueTypeAnnotation(ctx, declaration)
				if !booleanAnnotation(ctx, annotation, visiting) {
					return false
				}
			}
		}
		return found
	}
	declaration := symbol.Declarations[0]
	annotation, optional := valueTypeAnnotation(ctx, declaration)
	if annotation != nil && !utils.IsJSDocSyntaxNode(annotation) && !ast.IsFunctionLike(declaration) {
		return booleanCallableAnnotation(ctx, annotation, visiting)
	}
	if callableHasWrites(ctx, symbol) || overridesJSDocCallable(declaration) {
		return false
	}
	function := sourceFunction(ctx, declaration)
	if function != nil && ast.IsFunctionLike(function) {
		if ast.GetFunctionFlags(function)&(ast.FunctionFlagsAsync|ast.FunctionFlagsGenerator) != 0 {
			return false
		}
		if result, _ := valueTypeAnnotation(ctx, function); result != nil {
			return booleanAnnotation(ctx, result, visiting)
		}
		if annotation == nil {
			body := function.Body()
			if body != nil && body.Kind == ast.KindBlock {
				statements := body.Statements()
				if len(statements) != 1 || statements[0].Kind != ast.KindReturnStatement {
					return false
				}
				body = statements[0].Expression()
			}
			return isBoolean(ctx, body, evaluator, visiting, completed)
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
	if declaration.Kind == ast.KindVariableDeclaration && ast.IsVarConst(declaration.Parent) {
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
		return callee != nil && ast.IsIdentifier(callee) && slices.Contains(names, callee.Text()) && unicornutil.IsGlobalReference(ctx, callee)
	}
	var annotation *ast.Node
	if ast.IsIdentifier(node) && ctx.Refs != nil {
		if symbol := ctx.Refs.ResolveInFile(node); symbol != nil && len(symbol.Declarations) == 1 {
			declaration := symbol.Declarations[0]
			var optional bool
			annotation, optional = valueTypeAnnotation(ctx, declaration)
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
					annotation, _ = valueTypeAnnotation(ctx, function)
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

func memberName(node *ast.Node, evaluator *utils.StaticStringEvaluator) (string, bool) {
	if node.Kind == ast.KindPropertyAccessExpression {
		return utils.AccessExpressionStaticName(node)
	}
	return evaluator.EvalToString(node.AsElementAccessExpression().ArgumentExpression)
}
