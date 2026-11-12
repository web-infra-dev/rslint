package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// BooleanAnalyzer proves primitive boolean values using source syntax and, if
// available, type information. It is used serially for one linted file.
type BooleanAnalyzer struct {
	ctx       *rule.RuleContext
	evaluator *utils.StaticStringEvaluator
	completed map[*ast.Node]bool
}

func NewBooleanAnalyzer(ctx *rule.RuleContext) *BooleanAnalyzer {
	return &BooleanAnalyzer{ctx: ctx}
}

func (a *BooleanAnalyzer) StaticEvaluator() *utils.StaticStringEvaluator {
	if a.evaluator == nil {
		a.evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(a.ctx.TypeChecker, a.ctx.SourceFile, a.ctx.Refs)
	}
	return a.evaluator
}

func (a *BooleanAnalyzer) IsExpression(node *ast.Node) bool {
	return a.isBoolean(node, nil)
}

// The native checker supplies type information; syntax still proves booleans
// through any/unknown assertions and other expressions the checker cannot type.
func (a *BooleanAnalyzer) isBoolean(node *ast.Node, visiting map[*ast.Node]bool) (result bool) {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || visiting[node] {
		return false
	}
	// These operations always produce a primitive boolean. Prove that before
	// static evaluation, which otherwise revisits every operand and binding.
	if ast.IsBooleanLiteral(node) || IsLogicalNot(node) || node.Kind == ast.KindDeleteExpression {
		return true
	}
	if ast.IsBinaryExpression(node) {
		operator := node.AsBinaryExpression().OperatorToken.Kind
		switch operator {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
			ast.KindEqualsToken, ast.KindCommaToken:
			// Value-preserving operators can require a proof about their operands.
		default:
			return ast.IsRelationalOperator(operator) || ast.IsEqualityOperator(operator)
		}
	}
	cacheable := ast.IsIdentifier(node) || ast.IsCallExpression(node)
	if cacheable {
		if result, known := a.completed[node]; known {
			return result
		}
	}
	if visiting == nil {
		visiting = make(map[*ast.Node]bool)
	}
	visiting[node] = true
	defer func() {
		delete(visiting, node)
		// Bindings and calls revisit shared initializers or function bodies.
		// Ordinary expression trees do not need an entry for every node.
		if cacheable {
			if a.completed == nil {
				a.completed = make(map[*ast.Node]bool)
			}
			a.completed[node] = result
		}
	}()
	ctx := *a.ctx
	typedBoolean := func() bool { return isBooleanType(ctx, a.expressionType(node, make(map[*ast.Node]bool))) }
	recur := func(child *ast.Node) bool { return a.isBoolean(child, visiting) }
	if ast.IsIdentifier(node) && ctx.Refs != nil {
		if symbol := ctx.Refs.ResolveInFile(node); symbol != nil && len(symbol.Declarations) == 1 {
			annotation, optional := booleanValueAnnotation(ctx, symbol.Declarations[0])
			if utils.IsJSDocSyntaxNode(annotation) {
				if !optional && booleanAnnotation(ctx, annotation, visiting) {
					return true
				}
				// Simplifying an annotation cannot establish the value at this use.
				return ctx.TypeChecker != nil && isBooleanType(ctx, ctx.TypeChecker.GetTypeAtLocation(node)) && typedBoolean()
			}
			if !optional && booleanAnnotation(ctx, annotation, visiting) {
				return true
			}
			if declaration := symbol.Declarations[0]; declaration.Kind == ast.KindVariableDeclaration && ast.IsVarConst(declaration.Parent) && recur(declaration.Initializer()) {
				return true
			}
		}
	}
	if ast.IsCallExpression(node) && a.booleanCall(node, visiting) {
		return true
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
	value, known := a.StaticEvaluator().EvalValueIfNoSideEffects(node)
	if _, boolean := value.(bool); known && boolean && !HasJSDocTypeReference(ctx, node, make(map[*ast.Symbol]bool)) {
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
		return typedBoolean()
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if !ast.IsOptionalChain(node) {
			name, known := a.memberName(node)
			switch name {
			case "dotAll", "global", "hasIndices", "ignoreCase", "multiline", "sticky", "unicode", "unicodeSets":
				if known && regexpReceiver(ctx, node.Expression(), visiting) {
					return true
				}
			}
		}
		return typedBoolean()
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
func (a *BooleanAnalyzer) expressionType(node *ast.Node, visiting map[*ast.Node]bool) *checker.Type {
	ctx := *a.ctx
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
				if annotation, optional := booleanValueAnnotation(ctx, declaration); annotation != nil && !ast.IsFunctionLike(declaration) {
					if optional {
						return actual
					}
					return ctx.TypeChecker.GetTypeFromTypeNode(annotation)
				}
				if declaration.Kind == ast.KindVariableDeclaration && ast.IsVarConst(declaration.Parent) && HasJSDocTypeReference(ctx, node, make(map[*ast.Symbol]bool)) {
					return a.expressionType(declaration.Initializer(), visiting)
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
					if annotation, optional := booleanValueAnnotation(ctx, declaration); utils.IsJSDocSyntaxNode(annotation) && !optional {
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
			if callable := a.expressionType(callee, visiting); callable != nil {
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
			name, known := a.memberName(node)
			if receiver := a.expressionType(node.Expression(), visiting); known && receiver != nil {
				if property := ctx.TypeChecker.GetPropertyOfType(receiver, name); property != nil {
					return checker.Checker_getTypeOfSymbol(ctx.TypeChecker, property)
				}
			}
		}
	}
	return actual
}
