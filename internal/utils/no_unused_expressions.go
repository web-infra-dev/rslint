package utils

import "github.com/microsoft/typescript-go/shim/ast"

// NoUnusedExpressionOptions contains the expression-shape switches shared by
// ESLint core and typescript-eslint's no-unused-expressions rules.
type NoUnusedExpressionOptions struct {
	AllowShortCircuit    bool
	AllowTernary         bool
	AllowTaggedTemplates bool
	EnforceForJSX        bool
}

// ParseNoUnusedExpressionOptions reads the shared ESLint-compatible options.
// ignoreDirectives only affects ESLint's legacy ecmaVersion: 3 mode, which
// rslint does not expose; directive prologues are always skipped.
func ParseNoUnusedExpressionOptions(raw any) NoUnusedExpressionOptions {
	opts := NoUnusedExpressionOptions{}
	optsMap := GetOptionsMap(raw)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["allowShortCircuit"].(bool); ok {
		opts.AllowShortCircuit = v
	}
	if v, ok := optsMap["allowTernary"].(bool); ok {
		opts.AllowTernary = v
	}
	if v, ok := optsMap["allowTaggedTemplates"].(bool); ok {
		opts.AllowTaggedTemplates = v
	}
	if v, ok := optsMap["enforceForJSX"].(bool); ok {
		opts.EnforceForJSX = v
	}
	return opts
}

// IsDisallowedUnusedExpression mirrors the no-unused-expressions checker:
// true means the expression has no accepted side effect and should be reported.
// This is rule semantics, not a general-purpose purity predicate.
func IsDisallowedUnusedExpression(node *ast.Node, opts NoUnusedExpressionOptions) bool {
	node = ast.SkipOuterExpressions(
		node,
		ast.OEKParentheses|ast.OEKTypeAssertions|ast.OEKNonNullAssertions,
	)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindArrayLiteralExpression,
		ast.KindArrowFunction,
		ast.KindClassExpression,
		ast.KindFalseKeyword,
		ast.KindFunctionExpression,
		ast.KindIdentifier,
		ast.KindMetaProperty,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNullKeyword,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindObjectLiteralExpression,
		ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindRegularExpressionLiteral,
		ast.KindStringLiteral,
		ast.KindSuperKeyword,
		ast.KindTemplateExpression,
		ast.KindThisKeyword,
		ast.KindTrueKeyword,
		ast.KindTypeOfExpression:
		return true

	case ast.KindJsxElement, ast.KindJsxFragment, ast.KindJsxSelfClosingElement:
		return opts.EnforceForJSX

	case ast.KindTaggedTemplateExpression:
		return !opts.AllowTaggedTemplates

	case ast.KindConditionalExpression:
		if !opts.AllowTernary {
			return true
		}
		cond := node.AsConditionalExpression()
		return cond == nil ||
			IsDisallowedUnusedExpression(cond.WhenTrue, opts) ||
			IsDisallowedUnusedExpression(cond.WhenFalse, opts)

	case ast.KindBinaryExpression:
		bin := node.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil {
			return false
		}
		opKind := bin.OperatorToken.Kind
		if ast.IsLogicalOrCoalescingBinaryOperator(opKind) {
			if opts.AllowShortCircuit {
				return IsDisallowedUnusedExpression(bin.Right, opts)
			}
			return true
		}
		if opKind == ast.KindCommaToken {
			return true
		}
		return !ast.IsAssignmentOperator(opKind)

	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		return prefix != nil && !isUpdateOperator(prefix.Operator)

	case ast.KindExpressionWithTypeArguments:
		return IsDisallowedUnusedExpression(node.Expression(), opts)

	case ast.KindAwaitExpression,
		ast.KindCallExpression,
		ast.KindDeleteExpression,
		ast.KindNewExpression,
		ast.KindPostfixUnaryExpression,
		ast.KindVoidExpression,
		ast.KindYieldExpression:
		return false

	// The current ESLint rule does not list TSSatisfiesExpression, so the
	// tsgo equivalent stays in the unknown/allowed bucket.
	default:
		return false
	}
}

func isUpdateOperator(kind ast.Kind) bool {
	return kind == ast.KindPlusPlusToken || kind == ast.KindMinusMinusToken
}

// IsDirectivePrologueStatement reports whether node is a leading string-literal
// expression statement in a Program, function body, or TS module block.
func IsDirectivePrologueStatement(node *ast.Node) bool {
	return isDirectivePrologueStatement(node, false)
}

// IsDirectivePrologueStatementIncludingClassStaticBlocks matches the
// @typescript-eslint/parser static-block behavior used by its extension rule.
func IsDirectivePrologueStatementIncludingClassStaticBlocks(node *ast.Node) bool {
	return isDirectivePrologueStatement(node, true)
}

func isDirectivePrologueStatement(node *ast.Node, includeClassStaticBlocks bool) bool {
	if node == nil || !ast.IsPrologueDirective(node) {
		return false
	}
	parent := node.Parent
	if parent == nil || !isDirectivePrologueContainer(parent, includeClassStaticBlocks) {
		return false
	}
	for _, stmt := range parent.Statements() {
		if stmt == node {
			return true
		}
		if !ast.IsPrologueDirective(stmt) {
			return false
		}
	}
	return false
}

func isDirectivePrologueContainer(node *ast.Node, includeClassStaticBlocks bool) bool {
	switch node.Kind {
	case ast.KindSourceFile, ast.KindModuleBlock:
		return true
	case ast.KindBlock:
		if node.Parent == nil {
			return false
		}
		if includeClassStaticBlocks {
			return ast.IsFunctionLikeOrClassStaticBlockDeclaration(node.Parent)
		}
		return ast.IsFunctionLike(node.Parent)
	default:
		return false
	}
}
