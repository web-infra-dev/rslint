// Package prefer_set_size ports eslint-plugin-unicorn's `prefer-set-size` rule.
package prefer_set_size

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const preferSetSizeMessageID = "prefer-set-size"

var preferSetSizeMessage = rule.RuleMessage{
	Id:          preferSetSizeMessageID,
	Description: "Prefer using `Set#size` instead of `Array#length`.",
}

var setTypeNames = utils.NewSetFromItems("Set", "ReadonlySet")

var PreferSetSizeRule = rule.Rule{
	Name:   "unicorn/prefer-set-size",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				checkMemberExpression(ctx, node)
			},
		}
	},
}

func checkMemberExpression(ctx rule.RuleContext, node *ast.Node) {
	if node == nil || ast.IsOptionalChainRoot(node) {
		return
	}

	member := node.AsPropertyAccessExpression()
	if member == nil || member.Name() == nil || member.Name().Text() != "length" {
		return
	}

	conversion, set := getSetNode(member.Expression)
	if conversion == nil || set == nil || !isSet(ctx, set, nil) {
		return
	}

	property := member.Name()
	ctx.ReportNodeWithDeferredFixes(property, preferSetSizeMessage, func() []rule.RuleFix {
		return createFixes(ctx, node, conversion, set, property)
	})
}

// getSetNode mirrors the two conversion shapes accepted upstream. The runtime
// helper removes tsgo-only parentheses while leaving authored TypeScript
// assertions visible, as TSESTree does for these matches.
func getSetNode(object *ast.Node) (conversion, set *ast.Node) {
	conversion = utils.ESTreeRuntimeExpression(object)
	if conversion == nil {
		return nil, nil
	}

	if ast.IsArrayLiteralExpression(conversion) {
		elements := conversion.AsArrayLiteralExpression().Elements
		if elements == nil || len(elements.Nodes) != 1 || !ast.IsSpreadElement(elements.Nodes[0]) {
			return nil, nil
		}
		return conversion, elements.Nodes[0].AsSpreadElement().Expression
	}

	zeroOptional := false
	oneArgument := 1
	fromCall, ok := unicornutil.MatchDotMethodCall(conversion, unicornutil.DotMethodCallOptions{
		Method:              "from",
		ArgumentsLength:     &oneArgument,
		RejectSpreadElement: true,
		AllowOptionalCall:   zeroOptional,
		AllowOptionalMember: zeroOptional,
	})
	if !ok {
		return nil, nil
	}
	array := utils.ESTreeRuntimeExpression(fromCall.Object)
	if array == nil || !ast.IsIdentifier(array) || array.Text() != "Array" {
		return nil, nil
	}
	return conversion, fromCall.Call.Arguments()[0]
}

func isSet(ctx rule.RuleContext, node *ast.Node, visiting map[*ast.Symbol]bool) bool {
	if node == nil {
		return false
	}
	runtime := utils.ESTreeRuntimeExpression(node)
	if runtime == nil {
		return false
	}
	if runtime != node {
		return isSet(ctx, runtime, visiting)
	}

	// Unicorn resolves the runtime expression before consulting type information.
	// This keeps `unknown satisfies Set` unknown while accepting a Set wrapped by
	// `satisfies` or a non-null assertion, and recognizes compound expressions
	// only when every possible result is a Set.
	switch runtime.Kind {
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		// An informative assertion classifies the runtime expression even when
		// the source-only Program has no TypeChecker.
		if isSetTypeAnnotation(ctx, runtime.Type(), map[*ast.Symbol]bool{}) {
			return true
		}
		return isSet(ctx, runtime.Expression(), visiting)
	case ast.KindSatisfiesExpression:
		satisfies := runtime.AsSatisfiesExpression()
		return satisfies != nil && isSet(ctx, satisfies.Expression, visiting)
	case ast.KindNonNullExpression:
		// `set!` can narrow `Set | null` to Set. Its inner identifier retains
		// the nullable declared type, so inspect the wrapper before unwrapping.
		if isSetWithChecker(ctx, runtime) {
			return true
		}
		nonNull := runtime.AsNonNullExpression()
		return nonNull != nil && isSet(ctx, nonNull.Expression, visiting)
	case ast.KindConditionalExpression:
		conditional := runtime.AsConditionalExpression()
		return conditional != nil &&
			isSet(ctx, conditional.WhenTrue, visiting) &&
			isSet(ctx, conditional.WhenFalse, visiting)
	case ast.KindBinaryExpression:
		binary := runtime.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindCommaToken {
			return isSet(ctx, binary.Right, visiting)
		}
	}

	// Like Unicorn's type helper, direct Set construction and const aliases work
	// in plain JavaScript without a TypeScript program.
	if isSetConstruction(runtime) {
		return true
	}
	if ast.IsIdentifier(runtime) {
		if visiting == nil {
			visiting = make(map[*ast.Symbol]bool)
		}
		if symbol := ctx.Refs.ResolveInFile(runtime); symbol != nil && !visiting[symbol] && len(symbol.Declarations) == 1 {
			declaration := symbol.Declarations[0]
			if isSetTypeAnnotation(ctx, setBindingTypeAnnotation(declaration), map[*ast.Symbol]bool{}) {
				return true
			}
			if declaration != nil && declaration.Kind == ast.KindVariableDeclaration {
				list := declaration.Parent
				variable := declaration.AsVariableDeclaration()
				if list != nil && list.Kind == ast.KindVariableDeclarationList &&
					list.Flags&ast.NodeFlagsConst != 0 && variable != nil &&
					ast.IsIdentifier(variable.Name()) && variable.Initializer != nil {
					visiting[symbol] = true
					result := isSet(ctx, variable.Initializer, visiting)
					delete(visiting, symbol)
					if result {
						return true
					}
				}
			}
		}
	}

	// eslint-plugin-unicorn uses parser type information after its syntax and
	// scope checks. Restrict this fallback to TypeScript inputs: JavaScript's
	// upstream parser has no TypeScript Program, so mutable JS bindings must not
	// become matches merely because tsgo can infer their type.
	return isSetWithChecker(ctx, runtime)
}

func isSetWithChecker(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil || ctx.TypeChecker == nil || ast.IsInJSFile(node) {
		return false
	}
	return unicornutil.ClassifyType(ctx, ctx.TypeChecker.GetTypeAtLocation(node), unicornutil.TypeClassifierOptions{
		TargetTypeNames: setTypeNames,
	}) == unicornutil.TypeTarget
}

func setBindingTypeAnnotation(declaration *ast.Node) *ast.Node {
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

// isSetTypeAnnotation recognizes the source-level annotations that Unicorn's
// type helper can use before a TypeScript Program is available. Type aliases
// and constraints are resolved through the rule's binder, not the checker.
func isSetTypeAnnotation(ctx rule.RuleContext, node *ast.Node, visiting map[*ast.Symbol]bool) bool {
	if node == nil || utils.IsJSDocSyntaxNode(node) {
		return false
	}
	for node.Kind == ast.KindParenthesizedType {
		node = node.AsParenthesizedTypeNode().Type
		if node == nil {
			return false
		}
	}
	if node.Kind != ast.KindTypeReference {
		return false
	}
	return isSetTypeReference(ctx, node.AsTypeReferenceNode().TypeName, visiting)
}

func isSetTypeReference(ctx rule.RuleContext, typeName *ast.Node, visiting map[*ast.Symbol]bool) bool {
	if typeName == nil || !ast.IsIdentifier(typeName) {
		return false
	}
	name := typeName.AsIdentifier().Text
	if setTypeNames.Has(name) {
		return true
	}
	if ctx.Refs == nil {
		return false
	}
	symbol := ctx.Refs.ResolveInFileWithMeaning(
		typeName, ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias,
	)
	if symbol == nil || visiting[symbol] {
		return false
	}
	visiting[symbol] = true
	defer delete(visiting, symbol)

	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}
		switch declaration.Kind {
		case ast.KindTypeAliasDeclaration:
			return isSetTypeAnnotation(ctx, declaration.AsTypeAliasDeclaration().Type, visiting)
		case ast.KindTypeParameter:
			return isSetTypeAnnotation(ctx, declaration.AsTypeParameterDeclaration().Constraint, visiting)
		}
	}
	return false
}

func isSetConstruction(node *ast.Node) bool {
	if node == nil || !ast.IsNewExpression(node) {
		return false
	}
	newExpression := node.AsNewExpression()
	if newExpression == nil {
		return false
	}
	callee := utils.ESTreeRuntimeExpression(newExpression.Expression)
	return callee != nil && ast.IsIdentifier(callee) && callee.Text() == "Set"
}

func createFixes(ctx rule.RuleContext, member, conversion, set, property *ast.Node) []rule.RuleFix {
	conversionRange := utils.TrimNodeTextRange(ctx.SourceFile, conversion)
	// TSESTree does not retain grouping parentheses around a spread argument.
	// Drop tsgo's ParenthesizedExpression wrappers before deriving the replacement
	// range, but keep TypeScript assertion expressions, which need protection when
	// they become the object of `.size`.
	setForFix := utils.ESTreeRuntimeExpression(set)
	if setForFix == nil {
		return nil
	}
	setRange := utils.TrimNodeTextRange(ctx.SourceFile, setForFix)
	if hasCommentOutsideSet(ctx, conversionRange, setRange) {
		return nil
	}

	setText := ctx.SourceFile.Text()[setRange.Pos():setRange.End()]
	if unicornutil.ShouldAddParenthesesToMemberExpressionObject(ctx.SourceFile, setForFix) {
		setText = "(" + setText + ")"
	}
	if conversion.Kind == ast.KindArrayLiteralExpression &&
		utils.NeedsLeadingSpaceForReplacement(ctx.SourceFile.Text(), conversionRange.Pos(), setText) {
		setText = " " + setText
	}

	return []rule.RuleFix{
		rule.RuleFixReplaceRange(utils.TrimNodeTextRange(ctx.SourceFile, property), "size"),
		rule.RuleFixReplaceRange(core.NewTextRange(conversionRange.Pos(), conversionRange.End()), setText),
	}
}

func hasCommentOutsideSet(ctx rule.RuleContext, conversion, set core.TextRange) bool {
	for _, comment := range ctx.Comments.All() {
		if comment.Pos() < conversion.Pos() || comment.End() > conversion.End() {
			continue
		}
		if comment.Pos() < set.Pos() || comment.End() > set.End() {
			return true
		}
	}
	return false
}
