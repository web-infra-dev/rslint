package prefer_regexp_test

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Mixed unions matter here: a string-or-RegExp pattern is excluded upstream,
// whereas a RegExp-or-number pattern can still receive a suggestion.
type expressionType uint8

const (
	typeUnknown expressionType = iota
	typeString
	typeRegExp
	typeOther
	typeStringOrOther
	typeRegExpOrOther
)

type expressionTypes struct {
	ctx       rule.RuleContext
	evaluator *utils.StaticStringEvaluator
	visiting  map[*ast.Symbol]bool
}

func (types *expressionTypes) expression(node *ast.Node) expressionType {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return typeUnknown
	}
	if node.Kind == ast.KindRegularExpressionLiteral || (node.Kind == ast.KindNewExpression && ast.IsIdentifier(utils.ESTreeCallCallee(node.Expression())) && utils.ESTreeCallCallee(node.Expression()).Text() == "RegExp") {
		return typeRegExp
	}
	if unicornutil.IsString(types.ctx, node, types.evaluator) {
		return typeString
	}
	switch node.Kind {
	case ast.KindIdentifier:
		if types.ctx.Refs != nil {
			symbol := types.ctx.Refs.ResolveInFile(node)
			if symbol != nil && len(symbol.Declarations) == 1 && !types.visiting[symbol] {
				declaration := symbol.Declarations[0]
				if declaration.Kind == ast.KindParameter || declaration.Kind == ast.KindVariableDeclaration || declaration.Kind == ast.KindBindingElement {
					if result := annotationType(declaration.Type()); result != typeUnknown {
						return result
					}
				}
				if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent != nil && ast.IsVariableDeclarationList(declaration.Parent) && ast.IsVarConst(declaration.Parent) && declaration.AsVariableDeclaration().Initializer != nil {
					types.visiting[symbol] = true
					defer delete(types.visiting, symbol)
					return types.expression(declaration.AsVariableDeclaration().Initializer)
				}
			}
		}
	case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression:
		if result := annotationType(node.Type()); result != typeUnknown {
			return result
		}
		return types.expression(node.Expression())
	case ast.KindNonNullExpression:
		return types.expression(node.Expression())
	case ast.KindBinaryExpression:
		if node.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
			return types.expression(node.AsBinaryExpression().Right)
		}
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return combineTypes([]expressionType{types.expression(conditional.WhenTrue), types.expression(conditional.WhenFalse)}, false)
	}
	if value, known := types.evaluator.EvalValueIfNoSideEffects(node); known {
		if _, text := value.(string); text {
			return typeString
		}
		if utils.IsPrimitiveStaticValue(value) {
			text, convertible := types.evaluator.EvalToString(node)
			if !convertible || text != "undefined" || node.Kind == ast.KindVoidExpression || (ast.IsIdentifier(node) && node.Text() == "undefined") {
				return typeOther
			}
		}
		if _, _, regex := types.evaluator.EvalRegExp(node); regex {
			return typeRegExp
		}
	} else if types.evaluator.HasNonUndefinedValueIgnoringPropertyMutations(node) {
		// A folded value behind a mutable member is evidence of an incompatible
		// receiver, never evidence that moving it is safe.
		return typeOther
	}
	switch node.Kind {
	case ast.KindNullKeyword, ast.KindArrayLiteralExpression, ast.KindArrowFunction, ast.KindClassExpression, ast.KindFunctionExpression, ast.KindObjectLiteralExpression:
		return typeOther
	}
	// JavaScript parser services do not supply project types upstream. Authored
	// TS annotations were handled above; only TS files use the optional checker.
	if types.ctx.TypeChecker != nil && !ast.IsSourceFileJS(types.ctx.SourceFile) {
		return checkerType(types.ctx, types.ctx.TypeChecker.GetTypeAtLocation(node))
	}
	return typeUnknown
}

func annotationType(node *ast.Node) expressionType {
	if node == nil || utils.IsJSDocSyntaxNode(node) {
		return typeUnknown
	}
	switch node.Kind {
	case ast.KindParenthesizedType:
		return annotationType(node.Type())
	case ast.KindStringKeyword:
		return typeString
	case ast.KindLiteralType:
		if ast.IsStringLiteralLike(node.AsLiteralTypeNode().Literal) {
			return typeString
		}
		return typeOther
	case ast.KindTypeReference:
		name := node.AsTypeReferenceNode().TypeName
		if !ast.IsIdentifier(name) {
			return typeUnknown
		}
		switch name.Text() {
		case "String":
			return typeString
		case "RegExp":
			return typeRegExp
		case "Array", "ReadonlyArray":
			return typeOther
		}
	case ast.KindUnionType, ast.KindIntersectionType:
		var parts []*ast.Node
		if node.Kind == ast.KindIntersectionType {
			parts = node.AsIntersectionTypeNode().Types.Nodes
		} else {
			parts = node.AsUnionTypeNode().Types.Nodes
		}
		classes := make([]expressionType, 0, len(parts))
		for _, part := range parts {
			classes = append(classes, annotationType(part))
		}
		return combineTypes(classes, node.Kind == ast.KindIntersectionType)
	case ast.KindBigIntKeyword, ast.KindBooleanKeyword, ast.KindNeverKeyword, ast.KindNullKeyword, ast.KindNumberKeyword, ast.KindObjectKeyword, ast.KindSymbolKeyword, ast.KindUndefinedKeyword, ast.KindVoidKeyword, ast.KindArrayType, ast.KindTypeLiteral, ast.KindTupleType, ast.KindFunctionType, ast.KindConstructorType:
		return typeOther
	}
	return typeUnknown
}

func combineTypes(types []expressionType, intersection bool) expressionType {
	hasString, hasRegExp, hasOther := false, false, false
	for _, kind := range types {
		if kind == typeUnknown {
			return typeUnknown
		}
		hasString = hasString || kind == typeString || kind == typeStringOrOther
		hasRegExp = hasRegExp || kind == typeRegExp || kind == typeRegExpOrOther
		hasOther = hasOther || kind == typeOther || kind == typeStringOrOther || kind == typeRegExpOrOther
	}
	if intersection {
		if hasRegExp {
			return typeRegExp
		}
		if hasString {
			return typeString
		}
		return typeOther
	}
	if hasString {
		if hasRegExp || hasOther {
			return typeStringOrOther
		}
		return typeString
	}
	if hasRegExp {
		if hasOther {
			return typeRegExpOrOther
		}
		return typeRegExp
	}
	return typeOther
}

func checkerType(ctx rule.RuleContext, t *checker.Type) expressionType {
	if t == nil || utils.IsTypeAnyType(t) || utils.IsTypeUnknownType(t) || utils.IsIntrinsicErrorType(t) {
		return typeUnknown
	}
	if utils.IsTypeParameter(t) {
		constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, t)
		if constraint == nil || constraint == t {
			return typeUnknown
		}
		return checkerType(ctx, constraint)
	}
	if utils.IsUnionType(t) || utils.IsIntersectionType(t) {
		parts := utils.UnionTypeParts(t)
		intersection := utils.IsIntersectionType(t)
		if intersection {
			parts = utils.IntersectionTypeParts(t)
		}
		classes := make([]expressionType, 0, len(parts))
		for _, part := range parts {
			classes = append(classes, checkerType(ctx, part))
		}
		return combineTypes(classes, intersection)
	}
	if utils.IsTypeFlagSet(t, checker.TypeFlagsString|checker.TypeFlagsStringLiteral) {
		return typeString
	}
	if name, ok := unicornutil.TypeSymbolName(t); ok && name == "RegExp" {
		return typeRegExp
	}
	if symbol := checker.Type_symbol(t); symbol != nil && symbol.Flags&(ast.SymbolFlagsClass|ast.SymbolFlagsInterface) != 0 && checker.Type_objectFlags(t)&(checker.ObjectFlagsClassOrInterface|checker.ObjectFlagsReference) != 0 {
		for _, base := range ctx.TypeChecker.GetBaseTypes(checker.Checker_getDeclaredTypeOfSymbol(ctx.TypeChecker, symbol)) {
			if checkerType(ctx, base) == typeRegExp {
				return typeRegExp
			}
		}
	}
	return typeOther
}
