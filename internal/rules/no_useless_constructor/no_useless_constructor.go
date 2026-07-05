package no_useless_constructor

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildNoUselessConstructorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noUselessConstructor",
		Description: "Useless constructor.",
	}
}

func buildRemoveConstructorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeConstructor",
		Description: "Remove the constructor.",
	}
}

func checkAccessibility(node *ast.Node, classHasSuperClass bool) bool {
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsPrivate) {
		return false
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsProtected) {
		return false
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsPublic) {
		if classHasSuperClass {
			return false
		}
	}
	return true
}

func checkParams(node *ast.Node, params []*ast.Node) bool {
	for _, param := range params {
		if !ast.IsParameterDeclaration(param) {
			continue
		}
		if ast.IsParameterPropertyDeclaration(param, node) {
			return false
		}
		if ast.HasDecorators(param) {
			return false
		}
	}
	return true
}

func isSimpleParam(param *ast.Node) bool {
	if !ast.IsParameterDeclaration(param) {
		return false
	}
	pd := param.AsParameterDeclaration()
	if pd == nil || pd.Initializer != nil {
		return false
	}
	// Rest params are considered simple regardless of binding pattern
	// (`...[x,y]` counts). `isValidRestSpreadPair` still checks that the
	// rest binding is a plain identifier for passing-through; `...arguments`
	// forwarding is detected separately and does not require a plain name.
	if utils.IsRestParameterDeclaration(param) {
		return true
	}
	name := param.Name()
	return name != nil && name.Kind == ast.KindIdentifier
}

func isSingleSuperCall(statements []*ast.Node) bool {
	if len(statements) != 1 {
		return false
	}
	stmt := statements[0]
	if !ast.IsExpressionStatement(stmt) {
		return false
	}
	expr := ast.SkipParentheses(stmt.Expression())
	return ast.IsSuperCall(expr)
}

// superCallOf returns the CallExpression node of the single super() call
// body, after stripping any parentheses around the expression statement.
func superCallOf(statements []*ast.Node) *ast.Node {
	return ast.SkipParentheses(statements[0].Expression())
}

func isSpreadArguments(args []*ast.Node) bool {
	if len(args) != 1 {
		return false
	}
	arg := args[0]
	if !ast.IsSpreadElement(arg) {
		return false
	}
	inner := ast.SkipParentheses(arg.AsSpreadElement().Expression)
	return inner != nil && inner.Kind == ast.KindIdentifier && inner.Text() == "arguments"
}

func isValidIdentifierPair(paramName *ast.Node, superArg *ast.Node) bool {
	superArg = ast.SkipParentheses(superArg)
	return paramName.Kind == ast.KindIdentifier &&
		superArg != nil &&
		superArg.Kind == ast.KindIdentifier &&
		paramName.Text() == superArg.Text()
}

func isValidRestSpreadPair(param *ast.Node, superArg *ast.Node) bool {
	if !utils.IsRestParameterDeclaration(param) {
		return false
	}
	if !ast.IsSpreadElement(superArg) {
		return false
	}
	inner := ast.SkipParentheses(superArg.AsSpreadElement().Expression)
	if inner == nil {
		return false
	}
	paramName := param.Name()
	return paramName != nil && isValidIdentifierPair(paramName, inner)
}

func isPassingThrough(params []*ast.Node, args []*ast.Node) bool {
	if len(params) != len(args) {
		return false
	}
	for i := range params {
		paramName := params[i].Name()
		if paramName == nil {
			return false
		}
		if utils.IsRestParameterDeclaration(params[i]) {
			if !isValidRestSpreadPair(params[i], args[i]) {
				return false
			}
		} else {
			if !isValidIdentifierPair(paramName, args[i]) {
				return false
			}
		}
	}
	return true
}

func isRedundantSuperCall(statements []*ast.Node, params []*ast.Node) bool {
	if !isSingleSuperCall(statements) {
		return false
	}
	for _, p := range params {
		if !isSimpleParam(p) {
			return false
		}
	}
	call := superCallOf(statements).AsCallExpression()
	var args []*ast.Node
	if call.Arguments != nil {
		args = call.Arguments.Nodes
	}
	return isSpreadArguments(args) || isPassingThrough(params, args)
}

// reportRange returns the range from the start of the constructor node (first
// non-trivia position) to the end of the `constructor` keyword (or the
// `'constructor'` string-literal key) — mirroring ESLint's reported location,
// which stops just before the parameter list's `(`.
func reportRange(ctx rule.RuleContext, node *ast.Node, constructor *ast.ConstructorDeclaration) core.TextRange {
	start := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
	end := node.End()
	if constructor.Parameters != nil {
		text := ctx.SourceFile.Text()
		// Walk backward from the first position inside `(...)` until we find
		// `(`, then strip whitespace between the keyword and `(`.
		p := constructor.Parameters.Pos()
		for p > start && text[p-1] != '(' {
			p--
		}
		if p > start && text[p-1] == '(' {
			p--
		}
		p = utils.SkipTrailingWhitespace(text, start, p)
		end = p
	}
	return core.NewTextRange(start, end)
}

// needsLeadingSemicolon reports whether removing `node` outright would cause
// an ASI hazard with the following class member. Mirrors ESLint's fix for
// cases like `foo = 'bar'\n constructor() {}\n [0]() {}` where simply dropping
// the constructor would leave `foo = 'bar'` followed by `[0]()` and ASI would
// reparse that as `foo = 'bar'[0]()`. A preceding `;` is inserted in its place.
//
// Matches ESLint's token-level check (`nextToken.value === "["`): any class
// element whose first non-trivia token is `[` is an ASI risk — this covers
// computed property names, computed methods/accessors, and TS index
// signatures. A decorator `@`, modifier keyword (`static`/`readonly`/`public`/
// `private`/`protected`/`override`/`abstract`/`declare`/`async`), plain name,
// or a stray `;` (SemicolonClassElement) all shift the first token away from
// `[` and therefore are safe.
func needsLeadingSemicolon(sf *ast.SourceFile, classNode *ast.Node, node *ast.Node) bool {
	nextToken, ok := utils.TokenAtOrAfter(sf, node.End())
	if !ok || nextToken.Kind != ast.KindOpenBracketToken {
		return false
	}
	return utils.NeedsClassMemberLeadingSemicolon(
		sf,
		classNode,
		node,
		nextToken,
		utils.ClassMemberLeadingSemicolonOptions{IncludePropertiesWithoutInitializers: true},
	)
}

var NoUselessConstructorRule = rule.Rule{
	Name: "no-useless-constructor",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindConstructor: func(node *ast.Node) {
				constructor := node.AsConstructorDeclaration()
				if constructor == nil || constructor.Body == nil {
					return
				}
				// tsgo parses `static constructor()` as KindConstructor, but per
				// the class semantics a static method named `constructor` is a
				// regular method — not the class's constructor — so it cannot
				// be "useless" the way this rule means it.
				if ast.IsStatic(node) {
					return
				}

				classNode := ast.GetContainingClass(node)
				if classNode == nil {
					return
				}

				hasSuper := ast.GetExtendsHeritageClauseElement(classNode) != nil

				if !checkAccessibility(node, hasSuper) {
					return
				}

				var params []*ast.Node
				if constructor.Parameters != nil {
					params = constructor.Parameters.Nodes
				}
				if !checkParams(node, params) {
					return
				}

				body := constructor.Body.Statements()

				var useless bool
				if hasSuper {
					useless = isRedundantSuperCall(body, params)
				} else {
					useless = len(body) == 0
				}

				if !useless {
					return
				}

				var fix rule.RuleFix
				if needsLeadingSemicolon(ctx.SourceFile, classNode, node) {
					fix = rule.RuleFixReplace(ctx.SourceFile, node, ";")
				} else {
					fix = rule.RuleFixRemove(ctx.SourceFile, node)
				}

				ctx.ReportRangeWithSuggestions(
					reportRange(ctx, node, constructor),
					buildNoUselessConstructorMessage(),
					rule.RuleSuggestion{
						Message:  buildRemoveConstructorMessage(),
						FixesArr: []rule.RuleFix{fix},
					},
				)
			},
		}
	},
}
