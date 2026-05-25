package prefer_destructuring

import (
	"math"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// destructuringConfig holds the array/object enforcement flags for a given context.
type destructuringConfig struct {
	Array  bool
	Object bool
}

// Options for the prefer-destructuring rule.
// First element: enforcement types (flat or per-node-type).
// Second element: additional enforcement flags.
type options struct {
	VariableDeclarator   destructuringConfig
	AssignmentExpression destructuringConfig

	EnforceForRenamedProperties             bool
	EnforceForDeclarationWithTypeAnnotation bool
}

func parseOptions(raw any) options {
	opts := options{
		VariableDeclarator:   destructuringConfig{Array: true, Object: true},
		AssignmentExpression: destructuringConfig{Array: true, Object: true},
	}

	if raw == nil {
		return opts
	}

	// Options come as an array [enabledTypes, enforcementOptions]
	optArray, isArray := raw.([]interface{})
	if !isArray || len(optArray) == 0 {
		// Try bare object format (from CLI single-option unwrap)
		if m, ok := raw.(map[string]interface{}); ok {
			parseFirstOption(m, &opts)
		}
		return opts
	}

	// Parse first element: enabled types
	if first, ok := optArray[0].(map[string]interface{}); ok {
		parseFirstOption(first, &opts)
	}

	// Parse second element: enforcement options
	if len(optArray) > 1 {
		if second, ok := optArray[1].(map[string]interface{}); ok {
			if v, ok := second["enforceForRenamedProperties"].(bool); ok {
				opts.EnforceForRenamedProperties = v
			}
			if v, ok := second["enforceForDeclarationWithTypeAnnotation"].(bool); ok {
				opts.EnforceForDeclarationWithTypeAnnotation = v
			}
		}
	}

	return opts
}

// parseFirstOption handles the first option element which can be either:
// - flat: { array: bool, object: bool }
// - per-context: { VariableDeclarator: { array, object }, AssignmentExpression: { array, object } }
func parseFirstOption(m map[string]interface{}, opts *options) {
	// Check if flat format (has "array" or "object" key)
	_, hasArray := m["array"]
	_, hasObject := m["object"]
	if hasArray || hasObject {
		// Flat format: only explicitly set keys are enabled, rest defaults to false.
		// This matches ESLint where `{ array: true }` means object is undefined (falsy).
		cfg := destructuringConfig{}
		if v, ok := m["array"].(bool); ok {
			cfg.Array = v
		}
		if v, ok := m["object"].(bool); ok {
			cfg.Object = v
		}
		opts.VariableDeclarator = cfg
		opts.AssignmentExpression = cfg
		return
	}

	// Per-context format
	if vd, ok := m["VariableDeclarator"].(map[string]interface{}); ok {
		cfg := destructuringConfig{} // defaults to false
		if v, ok := vd["array"].(bool); ok {
			cfg.Array = v
		}
		if v, ok := vd["object"].(bool); ok {
			cfg.Object = v
		}
		opts.VariableDeclarator = cfg
	}
	if ae, ok := m["AssignmentExpression"].(map[string]interface{}); ok {
		cfg := destructuringConfig{} // defaults to false
		if v, ok := ae["array"].(bool); ok {
			cfg.Array = v
		}
		if v, ok := ae["object"].(bool); ok {
			cfg.Object = v
		}
		opts.AssignmentExpression = cfg
	}
}

func buildPreferDestructuringMessage(typ string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferDestructuring",
		Description: "Use " + typ + " destructuring.",
	}
}

const precedenceOfAssignmentExpr = 1

// PreferDestructuringRule implements @typescript-eslint/prefer-destructuring.
var PreferDestructuringRule = rule.CreateRule(rule.Rule{
	Name:             "prefer-destructuring",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		return rule.RuleListeners{
			// VariableDeclarator (ESTree) → VariableDeclaration (tsgo)
			ast.KindVariableDeclaration: func(node *ast.Node) {
				varDecl := node.AsVariableDeclaration()
				if varDecl == nil {
					return
				}
				init := varDecl.Initializer
				if init == nil {
					return
				}

				// Skip using/await using declarations
				parent := node.Parent
				if parent != nil && parent.Kind == ast.KindVariableDeclarationList {
					if ast.IsVarUsing(parent) || ast.IsVarAwaitUsing(parent) {
						return
					}
				}

				performCheck(ctx, opts, varDecl.Name(), init, node, true)
			},

			// AssignmentExpression (ESTree) → BinaryExpression with = operator (tsgo)
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				if bin == nil || bin.OperatorToken == nil {
					return
				}
				if bin.OperatorToken.Kind != ast.KindEqualsToken {
					return
				}
				performCheck(ctx, opts, bin.Left, bin.Right, node, false)
			},
		}
	},
})

func performCheck(ctx rule.RuleContext, opts options, leftNode *ast.Node, rightNode *ast.Node, reportNode *ast.Node, isVariableDeclarator bool) {
	if rightNode == nil {
		return
	}

	// Unwrap parentheses on RHS
	right := ast.SkipParentheses(rightNode)

	// RHS must be a member expression (property access or element access)
	if right.Kind != ast.KindPropertyAccessExpression && right.Kind != ast.KindElementAccessExpression {
		return
	}

	// Skip optional chaining
	if ast.IsOptionalChain(right) {
		return
	}

	// Skip super access and private identifiers
	var objectNode *ast.Node
	if right.Kind == ast.KindPropertyAccessExpression {
		pae := right.AsPropertyAccessExpression()
		objectNode = pae.Expression
		if pae.Name() != nil && ast.IsPrivateIdentifier(pae.Name()) {
			return
		}
	} else {
		eae := right.AsElementAccessExpression()
		objectNode = eae.Expression
	}

	if objectNode != nil && ast.SkipParentheses(objectNode).Kind == ast.KindSuperKeyword {
		return
	}

	// Determine whether the LHS has a type annotation
	leftHasTypeAnnotation := hasTypeAnnotation(leftNode)

	// Determine whether to apply fix
	canFix := !leftHasTypeAnnotation && isVariableDeclarator

	// If LHS has type annotation and enforcement is off, skip
	if leftHasTypeAnnotation && !opts.EnforceForDeclarationWithTypeAnnotation {
		return
	}

	// Get the enabled config for this context
	var cfg destructuringConfig
	if isVariableDeclarator {
		cfg = opts.VariableDeclarator
	} else {
		cfg = opts.AssignmentExpression
	}

	// Check for integer-literal index access (array-like)
	if isArrayLiteralIntegerIndexAccess(right) {
		// typescript-eslint uses type info to determine if this is truly iterable
		if ctx.TypeChecker != nil && objectNode != nil {
			objType := ctx.TypeChecker.GetTypeAtLocation(objectNode)
			if !isTypeAnyOrIterableType(objType, ctx.TypeChecker) {
				// Non-iterable: report as object if enforceForRenamedProperties + object enabled
				if !opts.EnforceForRenamedProperties || !cfg.Object {
					return
				}
				ctx.ReportNode(reportNode, buildPreferDestructuringMessage("object"))
				return
			}
		}
		// Iterable or no type checker: report as array
		if cfg.Array {
			ctx.ReportNode(reportNode, buildPreferDestructuringMessage("array"))
		}
		return
	}

	// Object destructuring path
	if !cfg.Object {
		return
	}

	if opts.EnforceForRenamedProperties {
		if canFix && shouldFix(leftNode, right) {
			reportWithFix(ctx, leftNode, right, reportNode)
		} else {
			ctx.ReportNode(reportNode, buildPreferDestructuringMessage("object"))
		}
		return
	}

	// Same-name check: property name must match variable name
	leftName := getIdentifierName(leftNode)
	if leftName == "" {
		return
	}

	propName, ok := utils.AccessExpressionStaticName(right)
	if !ok || propName == "" {
		return
	}

	if leftName != propName {
		return
	}

	if canFix && shouldFix(leftNode, right) {
		reportWithFix(ctx, leftNode, right, reportNode)
	} else {
		ctx.ReportNode(reportNode, buildPreferDestructuringMessage("object"))
	}
}

// hasTypeAnnotation checks if the left-hand side node has a type annotation.
func hasTypeAnnotation(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		parent := node.Parent
		if parent != nil && parent.Kind == ast.KindVariableDeclaration {
			return parent.AsVariableDeclaration().Type != nil
		}
		return false
	case ast.KindArrayBindingPattern, ast.KindObjectBindingPattern:
		parent := node.Parent
		if parent != nil && parent.Kind == ast.KindVariableDeclaration {
			return parent.AsVariableDeclaration().Type != nil
		}
		return false
	}
	return false
}

// isArrayLiteralIntegerIndexAccess checks if the node is a member expression
// accessing an integer index (e.g., x[0], x[1]).
func isArrayLiteralIntegerIndexAccess(node *ast.Node) bool {
	if node.Kind != ast.KindElementAccessExpression {
		return false
	}
	eae := node.AsElementAccessExpression()
	if eae.ArgumentExpression == nil {
		return false
	}
	arg := ast.SkipParentheses(eae.ArgumentExpression)
	if arg.Kind != ast.KindNumericLiteral {
		return false
	}
	val, err := strconv.ParseFloat(arg.Text(), 64)
	if err != nil {
		return false
	}
	return val == math.Trunc(val) && !math.IsInf(val, 0) && !math.IsNaN(val)
}

// isTypeAnyOrIterableType checks if the type is any or has [Symbol.iterator].
// For union types, all members must be any-or-iterable.
func isTypeAnyOrIterableType(t *checker.Type, tc *checker.Checker) bool {
	if t == nil {
		return false
	}
	if utils.IsTypeAnyType(t) {
		return true
	}
	if !utils.IsUnionType(t) {
		return utils.GetWellKnownSymbolPropertyOfType(t, "iterator", tc) != nil
	}
	return utils.Every(utils.UnionTypeParts(t), func(member *checker.Type) bool {
		return isTypeAnyOrIterableType(member, tc)
	})
}

// getIdentifierName returns the name of an Identifier node (skipping parens).
func getIdentifierName(node *ast.Node) string {
	n := ast.SkipParentheses(node)
	if n.Kind == ast.KindIdentifier {
		return n.Text()
	}
	return ""
}

// shouldFix checks if the autofix should be applied.
// Fix only applies for VariableDeclarator with Identifier LHS, non-computed
// PropertyAccessExpression RHS, and matching names.
func shouldFix(leftNode *ast.Node, rightNode *ast.Node) bool {
	left := ast.SkipParentheses(leftNode)
	if left.Kind != ast.KindIdentifier {
		return false
	}
	if rightNode.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	pae := rightNode.AsPropertyAccessExpression()
	if pae.Name() == nil || !ast.IsIdentifier(pae.Name()) {
		return false
	}
	return left.Text() == pae.Name().Text()
}

// reportWithFix reports the diagnostic with an autofix.
func reportWithFix(ctx rule.RuleContext, leftNode *ast.Node, rightNode *ast.Node, reportNode *ast.Node) {
	pae := rightNode.AsPropertyAccessExpression()
	propName := pae.Name().Text()
	objectExpr := pae.Expression

	// Suppress fix if there are comments outside the object expression that would
	// be lost by the rewrite. ESLint checks `getCommentsInside(node).length >
	// getCommentsInside(rightNode.object).length`. We check the two gap regions:
	// 1. Between the identifier and the object expression (covers `id /* c */ = ...`)
	// 2. Between the object expression end and the member-expr end (covers `obj /* c */ .prop`)
	text := ctx.SourceFile.Text()
	idRange := utils.TrimNodeTextRange(ctx.SourceFile, leftNode)
	objRange := utils.TrimNodeTextRange(ctx.SourceFile, objectExpr)
	nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, reportNode)

	if hasCommentInText(text, idRange.End(), objRange.Pos()) ||
		hasCommentInText(text, objRange.End(), nodeRange.End()) {
		ctx.ReportNode(reportNode, buildPreferDestructuringMessage("object"))
		return
	}

	// Get the inner expression text, stripping outer ParenthesizedExpression.
	// In ESTree there is no ParenthesizedExpression node, so ESLint's
	// `sourceCode.getText(rightNode.object)` naturally returns the inner text.
	// We must replicate this by skipping parens and using the inner range.
	innerObj := ast.SkipParentheses(objectExpr)
	innerRange := utils.TrimNodeTextRange(ctx.SourceFile, innerObj)

	// If stripping parens would lose a comment (e.g., `(/* c */ obj)`), suppress fix.
	if hasCommentInText(text, objRange.Pos(), innerRange.Pos()) ||
		hasCommentInText(text, innerRange.End(), objRange.End()) {
		ctx.ReportNode(reportNode, buildPreferDestructuringMessage("object"))
		return
	}

	objectText := text[innerRange.Pos():innerRange.End()]

	// Add parens if the inner expression has lower precedence than assignment.
	// EslintLikePrecedence returns -1 for TS-specific nodes (AsExpression, etc.)
	// which ESLint never sees; these don't need wrapping since they bind tightly.
	prec := utils.EslintLikePrecedence(innerObj)
	if prec >= 0 && prec < precedenceOfAssignmentExpr {
		objectText = "(" + objectText + ")"
	}

	replacement := "{" + propName + "} = " + objectText
	fixRange := core.NewTextRange(idRange.Pos(), nodeRange.End())

	ctx.ReportNodeWithFixes(reportNode, buildPreferDestructuringMessage("object"),
		rule.RuleFixReplaceRange(fixRange, replacement),
	)
}

// hasCommentInText checks whether a `/*` or `//` comment literal exists in the
// source text between positions start and end. This is a simple text scan that
// doesn't require scanner integration, and correctly finds comments anywhere in
// the range (unlike GetCommentsInRange which only scans from the range start).
func hasCommentInText(text string, start, end int) bool {
	if start < 0 || end > len(text) || start >= end {
		return false
	}
	s := text[start:end]
	for i := range len(s) - 1 {
		if s[i] == '/' && (s[i+1] == '/' || s[i+1] == '*') {
			return true
		}
	}
	return false
}
