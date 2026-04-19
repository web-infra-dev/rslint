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
		if !ast.IsParameter(param) {
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
	if !ast.IsParameter(param) {
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
		for p > start && isAsciiSpace(text[p-1]) {
			p--
		}
		end = p
	}
	return core.NewTextRange(start, end)
}

func isAsciiSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// needsLeadingSemicolon reports whether removing `node` outright would cause
// an ASI hazard with the following class member. Mirrors ESLint's fix for
// cases like `foo = 'bar'\n constructor() {}\n [0]() {}` where simply dropping
// the constructor would leave `foo = 'bar'` followed by `[0]()` and ASI would
// reparse that as `foo = 'bar'[0]()`. A preceding `;` is inserted in its place.
func needsLeadingSemicolon(sf *ast.SourceFile, classNode *ast.Node, node *ast.Node) bool {
	members := classNode.Members()
	idx := -1
	for i, m := range members {
		if m == node {
			idx = i
			break
		}
	}
	if idx < 0 || idx+1 >= len(members) {
		return false
	}
	nextName := members[idx+1].Name()
	if nextName == nil || !ast.IsComputedPropertyName(nextName) {
		return false
	}
	if idx == 0 {
		return false
	}
	prev := members[idx-1]
	// Only a PropertyDeclaration's initializer can greedily consume the next
	// `[...]`. Method / accessor / constructor / static-block / index-signature
	// all terminate at their own boundary, so the class body parser resumes
	// cleanly for the next member regardless of the final character.
	if !ast.IsPropertyDeclaration(prev) {
		return false
	}
	text := sf.Text()
	i := prev.End() - 1
	for i > prev.Pos() && isAsciiSpace(text[i]) {
		i--
	}
	if text[i] == ';' {
		return false
	}
	pd := prev.AsPropertyDeclaration()
	if pd != nil && pd.Initializer != nil {
		init := pd.Initializer
		// Postfix `++`/`--` are restricted productions — ASI always fires
		// after them. Mirrors ESLint's PUNCTUATORS allowlist.
		if init.Kind == ast.KindPostfixUnaryExpression {
			return false
		}
		// Arrow functions with a block body terminate at their own `}`, so
		// the following `[...]` cannot member-access into them. Mirrors
		// ESLint's needsPrecedingSemicolon (the `}` → ObjectExpression /
		// function-expr / class-expr branches explicitly exclude arrows).
		if init.Kind == ast.KindArrowFunction {
			if arrow := init.AsArrowFunction(); arrow != nil && arrow.Body != nil && arrow.Body.Kind == ast.KindBlock {
				return false
			}
		}
	}
	return true
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
