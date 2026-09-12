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

// Memoize declaration-based facts, not use-site checker results: mutable
// bindings can narrow differently at each reference.
type setWalkState struct {
	visiting map[*ast.Symbol]bool
	memo     map[*ast.Symbol]bool
}

func newSetWalkState() *setWalkState {
	return &setWalkState{
		visiting: make(map[*ast.Symbol]bool),
		memo:     make(map[*ast.Symbol]bool),
	}
}

func resolveSetSymbol(state *setWalkState, symbol *ast.Symbol, classify func() bool) bool {
	if symbol == nil || state.visiting[symbol] {
		return false
	}
	if result, ok := state.memo[symbol]; ok {
		return result
	}
	state.visiting[symbol] = true
	result := classify()
	delete(state.visiting, symbol)
	state.memo[symbol] = result
	return result
}

func isSet(ctx rule.RuleContext, node *ast.Node, state *setWalkState) bool {
	if node == nil {
		return false
	}
	runtime := utils.ESTreeRuntimeExpression(node)
	if runtime == nil {
		return false
	}
	if state == nil {
		state = newSetWalkState()
	}

	// Unicorn resolves the runtime expression before consulting type information.
	// This keeps `unknown satisfies Set` unknown while accepting a Set wrapped by
	// `satisfies` or a non-null assertion, and recognizes compound expressions
	// only when every possible result is a Set.
	switch runtime.Kind {
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		// An informative assertion classifies the runtime expression even when
		// the source-only Program has no TypeChecker.
		if isSetTypeAnnotation(ctx, runtime.Type(), nil) {
			return true
		}
		return isSet(ctx, runtime.Expression(), state)
	case ast.KindSatisfiesExpression:
		satisfies := runtime.AsSatisfiesExpression()
		return satisfies != nil && isSet(ctx, satisfies.Expression, state)
	case ast.KindNonNullExpression:
		// `set!` can narrow `Set | null` to Set. Its inner identifier retains
		// the nullable declared type, so inspect the wrapper before unwrapping.
		if isSetWithChecker(ctx, runtime) {
			return true
		}
		nonNull := runtime.AsNonNullExpression()
		return nonNull != nil && isSet(ctx, nonNull.Expression, state)
	case ast.KindConditionalExpression:
		conditional := runtime.AsConditionalExpression()
		return conditional != nil &&
			isSet(ctx, conditional.WhenTrue, state) &&
			isSet(ctx, conditional.WhenFalse, state)
	case ast.KindBinaryExpression:
		binary := runtime.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindCommaToken {
			return isSet(ctx, binary.Right, state)
		}
	}

	// Like Unicorn's type helper, direct Set construction and const aliases work
	// in plain JavaScript without a TypeScript program.
	if isSetConstruction(runtime) || ast.IsIdentifier(runtime) && isSetBinding(ctx, runtime, state) {
		return true
	}

	// eslint-plugin-unicorn uses parser type information after its syntax and
	// scope checks. Restrict this fallback to TypeScript inputs: JavaScript's
	// upstream parser has no TypeScript Program, so mutable JS bindings must not
	// become matches merely because tsgo can infer their type.
	return isSetWithChecker(ctx, runtime)
}

func isSetBinding(ctx rule.RuleContext, node *ast.Node, state *setWalkState) bool {
	if ctx.Refs == nil {
		return false
	}
	symbol := ctx.Refs.ResolveInFile(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	return resolveSetSymbol(state, symbol, func() bool {
		declaration := symbol.Declarations[0]
		if isSetTypeAnnotation(ctx, setBindingTypeAnnotation(declaration), nil) {
			return true
		}
		if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
			return false
		}
		list := declaration.Parent
		variable := declaration.AsVariableDeclaration()
		if list == nil || list.Kind != ast.KindVariableDeclarationList ||
			list.Flags&ast.NodeFlagsConst == 0 || variable == nil ||
			!ast.IsIdentifier(variable.Name()) || variable.Initializer == nil {
			return false
		}
		return isSet(ctx, variable.Initializer, state)
	})
}

func isSetWithChecker(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil || ctx.TypeChecker == nil || ast.IsInJSFile(node) {
		return false
	}
	return unicornutil.ClassifyType(ctx, ctx.TypeChecker.GetTypeAtLocation(node), unicornutil.TypeClassifierOptions{
		TargetTypeNames:     setTypeNames,
		HeritageSymbolFlags: ast.SymbolFlagsClass | ast.SymbolFlagsInterface,
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
// type helper can use before a TypeScript Program is available. Compound types,
// aliases, constraints, and heritage are resolved without requiring the checker.
func isSetTypeAnnotation(ctx rule.RuleContext, node *ast.Node, state *setWalkState) bool {
	if node == nil || utils.IsJSDocSyntaxNode(node) {
		return false
	}
	if state == nil {
		state = newSetWalkState()
	}
	for node.Kind == ast.KindParenthesizedType {
		node = node.AsParenthesizedTypeNode().Type
		if node == nil {
			return false
		}
	}
	switch node.Kind {
	case ast.KindUnionType:
		types := node.AsUnionTypeNode().Types
		if types == nil || len(types.Nodes) == 0 {
			return false
		}
		for _, part := range types.Nodes {
			if !isSetTypeAnnotation(ctx, part, state) {
				return false
			}
		}
		return true
	case ast.KindIntersectionType:
		types := node.AsIntersectionTypeNode().Types
		if types != nil {
			for _, part := range types.Nodes {
				if isSetTypeAnnotation(ctx, part, state) {
					return true
				}
			}
		}
		return false
	case ast.KindTypeReference:
		return isSetTypeReference(ctx, node.AsTypeReferenceNode().TypeName, state)
	default:
		return false
	}
}

func isSetTypeReference(ctx rule.RuleContext, typeName *ast.Node, state *setWalkState) bool {
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
	return resolveSetSymbol(state, symbol, func() bool {
		for _, declaration := range symbol.Declarations {
			if declaration == nil {
				continue
			}
			switch declaration.Kind {
			case ast.KindTypeAliasDeclaration:
				return isSetTypeAnnotation(ctx, declaration.AsTypeAliasDeclaration().Type, state)
			case ast.KindTypeParameter:
				return isSetTypeAnnotation(ctx, declaration.AsTypeParameterDeclaration().Constraint, state)
			case ast.KindInterfaceDeclaration, ast.KindClassDeclaration, ast.KindClassExpression:
				return isSetHeritage(ctx, declaration, state)
			}
		}
		return false
	})
}

func isSetHeritage(ctx rule.RuleContext, declaration *ast.Node, state *setWalkState) bool {
	clauses := utils.GetHeritageClauses(declaration)
	if clauses == nil {
		return false
	}
	for _, node := range clauses.Nodes {
		clause := node.AsHeritageClause()
		if clause == nil || clause.Token != ast.KindExtendsKeyword || clause.Types == nil {
			continue
		}
		for _, node := range clause.Types.Nodes {
			var typeName *ast.Node
			switch node.Kind {
			case ast.KindTypeReference:
				typeName = node.AsTypeReferenceNode().TypeName
			case ast.KindExpressionWithTypeArguments:
				typeName = node.AsExpressionWithTypeArguments().Expression
			}
			if isSetTypeReference(ctx, typeName, state) {
				return true
			}
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
