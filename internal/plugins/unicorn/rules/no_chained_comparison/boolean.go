package no_chained_comparison

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// The native checker supplies type information; syntax still proves booleans
// through any/unknown assertions and other expressions the checker cannot type.
func isBoolean(ctx rule.RuleContext, node *ast.Node, evaluator *utils.StaticStringEvaluator, visiting map[*ast.Node]bool) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || visiting[node] {
		return false
	}
	// These operations always produce a primitive boolean. Prove that before
	// static evaluation, which otherwise revisits every operand and binding.
	if ast.IsBooleanLiteral(node) || unicornutil.IsLogicalNot(node) || node.Kind == ast.KindDeleteExpression {
		return true
	}
	if ast.IsBinaryExpression(node) {
		operator := node.AsBinaryExpression().OperatorToken.Kind
		if ast.IsRelationalOperator(operator) || ast.IsEqualityOperator(operator) {
			return true
		}
	}
	if visiting == nil {
		visiting = make(map[*ast.Node]bool)
	}
	visiting[node] = true
	defer delete(visiting, node)
	typedBoolean := func() bool { return isBooleanType(ctx, expressionType(ctx, node, evaluator, make(map[*ast.Node]bool))) }
	recur := func(child *ast.Node) bool { return isBoolean(ctx, child, evaluator, visiting) }
	if ast.IsIdentifier(node) && ctx.Refs != nil {
		if symbol := ctx.Refs.ResolveInFile(node); symbol != nil && len(symbol.Declarations) == 1 {
			annotation, optional := unicornutil.ValueTypeAnnotation(ctx, symbol.Declarations[0])
			if utils.IsJSDocSyntaxNode(annotation) {
				return (!optional && booleanAnnotation(ctx, annotation, visiting)) || typedBoolean()
			}
			if !optional && booleanAnnotation(ctx, annotation, visiting) {
				return true
			}
		}
	}
	// A successful syntax proof avoids repeatedly folding each subtree of a
	// long logical expression. A failed proof must still allow static results
	// such as 0 || false, but not infer a boolean from only one logical operand.
	booleanSyntax := false
	switch node.Kind {
	case ast.KindConditionalExpression:
		booleanSyntax = true
		conditional := node.AsConditionalExpression()
		if recur(conditional.WhenTrue) && recur(conditional.WhenFalse) {
			return true
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			booleanSyntax = true
			if recur(binary.Left) && recur(binary.Right) {
				return true
			}
		}
	}
	// Static boolean results such as 0 || false take precedence over syntax.
	value, known := evaluator.EvalValueIfNoSideEffects(node)
	if _, boolean := value.(bool); known && boolean && !unicornutil.HasJSDocTypeReference(ctx, node, make(map[*ast.Symbol]bool)) {
		return true
	}
	if booleanSyntax {
		return false
	}
	switch node.Kind {
	case ast.KindPrefixUnaryExpression:
		return false
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken, ast.KindCommaToken:
			return recur(binary.Right)
		}
		return false
	case ast.KindAwaitExpression, ast.KindNonNullExpression:
		return recur(node.Expression())
	case ast.KindSatisfiesExpression:
		return booleanAnnotation(ctx, node.Type(), visiting) || recur(node.Expression()) || typedBoolean()
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		return booleanAnnotation(ctx, node.Type(), visiting) || typedBoolean() || ((node.Type().Kind == ast.KindAnyKeyword || node.Type().Kind == ast.KindUnknownKeyword) && recur(node.Expression()))
	case ast.KindCallExpression:
		return booleanCall(ctx, node, evaluator, visiting) || typedBoolean()
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if !ast.IsOptionalChain(node) {
			name, known := memberName(node, evaluator)
			switch name {
			case "dotAll", "global", "hasIndices", "ignoreCase", "multiline", "sticky", "unicode", "unicodeSets":
				if known && regexpReceiver(ctx, node.Expression(), visiting) {
					return true
				}
			}
		}
		return typedBoolean()
	case ast.KindIdentifier:
		if ctx.Refs != nil {
			if symbol := ctx.Refs.ResolveInFile(node); symbol != nil && len(symbol.Declarations) == 1 {
				declaration := symbol.Declarations[0]
				if declaration.Kind == ast.KindVariableDeclaration && ast.IsVarConst(declaration.Parent) {
					return recur(declaration.AsVariableDeclaration().Initializer) || typedBoolean()
				}
			}
		}
	}
	return typedBoolean()
}

func isBooleanType(ctx rule.RuleContext, value *checker.Type) bool {
	if value == nil {
		return false
	}
	if utils.IsUnionType(value) {
		for _, part := range utils.UnionTypeParts(value) {
			if !isBooleanType(ctx, part) {
				return false
			}
		}
		return true
	}
	if utils.IsTypeFlagSet(value, checker.TypeFlagsBooleanLike) {
		return true
	}
	if constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, value); constraint != nil && constraint != value {
		return isBooleanType(ctx, constraint)
	}
	return false
}

// Use expression types before declaration signatures or property symbols, which
// lose call instantiation and flow narrowing. Explicit JSDoc still supplies
// annotations that tsgo does not apply to TypeScript files.
func expressionType(ctx rule.RuleContext, node *ast.Node, evaluator *utils.StaticStringEvaluator, visiting map[*ast.Node]bool) *checker.Type {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || ctx.TypeChecker == nil || visiting[node] {
		return nil
	}
	visiting[node] = true
	defer delete(visiting, node)
	actual := ctx.TypeChecker.GetTypeAtLocation(node)
	known := actual != nil && !utils.IsTypeAnyType(actual) && !utils.IsTypeUnknownType(actual) && !utils.IsIntrinsicErrorType(actual)
	switch node.Kind {
	case ast.KindIdentifier:
		if ctx.Refs != nil {
			if symbol := ctx.Refs.ResolveInFile(node); symbol != nil && len(symbol.Declarations) == 1 {
				declaration := symbol.Declarations[0]
				if declaration.Type() != nil && !utils.IsJSDocSyntaxNode(declaration.Type()) {
					return actual
				}
				if annotation, optional := unicornutil.ValueTypeAnnotation(ctx, declaration); annotation != nil && !ast.IsFunctionLike(declaration) {
					if optional {
						return actual
					}
					return ctx.TypeChecker.GetTypeFromTypeNode(annotation)
				}
				if declaration.Kind == ast.KindVariableDeclaration && ast.IsVarConst(declaration.Parent) && unicornutil.HasJSDocTypeReference(ctx, node, make(map[*ast.Symbol]bool)) {
					return expressionType(ctx, declaration.Initializer(), evaluator, visiting)
				}
			}
		}
	case ast.KindCallExpression:
		if !ast.IsOptionalChain(node) {
			callee := utils.ESTreeCallCallee(node.Expression())
			if callee != nil && ast.IsIdentifier(callee) && ctx.Refs != nil {
				if symbol := ctx.Refs.ResolveInFile(callee); symbol != nil && len(symbol.Declarations) == 1 {
					declaration := symbol.Declarations[0]
					if overridesJSDocCallable(declaration) || callableHasWrites(ctx, symbol) {
						return actual
					}
					if annotation, optional := unicornutil.ValueTypeAnnotation(ctx, declaration); utils.IsJSDocSyntaxNode(annotation) && !optional {
						if ast.IsFunctionLike(declaration) && ast.GetFunctionFlags(declaration)&(ast.FunctionFlagsAsync|ast.FunctionFlagsGenerator) == 0 {
							return ctx.TypeChecker.GetTypeFromTypeNode(annotation)
						}
						// JSDoc function types in .ts files are not bound as
						// callable types by tsgo; their return annotation is usable.
						if annotation.Kind == ast.KindFunctionType {
							return ctx.TypeChecker.GetTypeFromTypeNode(annotation.Type())
						}
					}
				}
			}
			if known {
				return actual
			}
			if callable := expressionType(ctx, callee, evaluator, visiting); callable != nil {
				signatures := ctx.TypeChecker.GetSignaturesOfType(callable, checker.SignatureKindCall)
				if len(signatures) == 1 {
					return ctx.TypeChecker.GetReturnTypeOfSignature(signatures[0])
				}
			}
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if known {
			return actual
		}
		if !ast.IsOptionalChain(node) {
			name, known := memberName(node, evaluator)
			if receiver := expressionType(ctx, node.Expression(), evaluator, visiting); known && receiver != nil {
				if property := ctx.TypeChecker.GetPropertyOfType(receiver, name); property != nil {
					return checker.Checker_getTypeOfSymbol(ctx.TypeChecker, property)
				}
			}
		}
	}
	return actual
}
