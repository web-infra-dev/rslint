package array_callback_return

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/cfg"
)

//go:embed array_callback_return.schema.json
var schemaJSON []byte

type Options struct {
	AllowImplicit bool `json:"allowImplicit"`
	CheckForEach  bool `json:"checkForEach"`
	AllowVoid     bool `json:"allowVoid"`
}

func parseOptions(options []any) Options {
	var opts Options
	if len(options) != 0 {
		values, _ := options[0].(map[string]any)
		opts.AllowImplicit, _ = values["allowImplicit"].(bool)
		opts.CheckForEach, _ = values["checkForEach"].(bool)
		opts.AllowVoid, _ = values["allowVoid"].(bool)
	}
	return opts
}

var targetMethods = map[string]bool{
	"every": true, "filter": true, "find": true, "findIndex": true,
	"findLast": true, "findLastIndex": true, "flatMap": true, "forEach": true,
	"map": true, "reduce": true, "reduceRight": true, "some": true,
	"sort": true, "toSorted": true,
}

// memberName implements the syntax-only property lookup used by ESLint. In
// particular, a computed identifier or interpolated template stays unknown.
func memberName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		if name := node.Name(); name != nil {
			return name.Text()
		}
	case ast.KindElementAccessExpression:
		property := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if property != nil && (property.Kind == ast.KindStringLiteral || property.Kind == ast.KindNoSubstitutionTemplateLiteral) {
			return property.Text()
		}
	}
	return ""
}

func isArrayFrom(callee *ast.Node) bool {
	callee = ast.SkipParentheses(callee)
	if memberName(callee) != "from" {
		return false
	}
	object := ast.SkipParentheses(callee.Expression())
	// Upstream intentionally recognizes any identifier ending in Array, even
	// shadowed/custom constructors, rather than querying a type or global binding.
	return object != nil && object.Kind == ast.KindIdentifier && strings.HasSuffix(object.Text(), "Array")
}

// arrayMethodName follows the callback's destination through the same wrappers
// as upstream. A function returned by an immediately invoked function can be
// the callback; an IIFE itself and unrelated argument positions cannot.
func arrayMethodName(node *ast.Node) string {
	for current := node; current != nil && current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			current = parent
		case ast.KindBinaryExpression:
			switch parent.AsBinaryExpression().OperatorToken.Kind {
			case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
				current = parent
			default:
				return ""
			}
		case ast.KindConditionalExpression:
			current = parent
		case ast.KindReturnStatement:
			var fn *ast.Node
			for ancestor := parent.Parent; ancestor != nil; ancestor = ancestor.Parent {
				if ast.IsFunctionLikeOrClassStaticBlockDeclaration(ancestor) {
					fn = ancestor
					break
				}
			}
			if fn == nil {
				return ""
			}
			wrapper := utils.OutermostParenthesizedExpression(fn)
			call := wrapper.Parent
			if call == nil || call.Kind != ast.KindCallExpression || call.Expression() != wrapper {
				return ""
			}
			current = call
		case ast.KindCallExpression:
			args := parent.Arguments()
			if isArrayFrom(parent.Expression()) && len(args) >= 2 && args[1] == current {
				return "from"
			}
			method := memberName(parent.Expression())
			if targetMethods[method] && len(args) >= 1 && args[0] == current {
				return method
			}
			return ""
		default:
			return ""
		}
	}
	return ""
}

func callbackMessage(id, method string, fn *ast.Node) rule.RuleMessage {
	fullName := "Array.prototype." + method
	if method == "from" {
		fullName = "Array.from"
	}
	name := utils.GetFunctionNameWithKindCore(fn)
	description := fullName + "() expects a return value from " + name + "."
	switch id {
	case "expectedAtEnd":
		description = fullName + "() expects a value to be returned at the end of " + name + "."
	case "expectedNoReturnValue":
		description = fullName + "() expects no useless return value from " + name + "."
	}
	return rule.RuleMessage{Id: id, Description: description, Data: map[string]string{"name": name, "arrayMethodName": fullName}}
}

var wrapBracesMessage = rule.RuleMessage{Id: "wrapBraces", Description: "Wrap the expression in `{}`."}
var prependVoidMessage = rule.RuleMessage{Id: "prependVoid", Description: "Prepend `void` to the expression."}

func prependVoidFixes(ctx rule.RuleContext, expr *ast.Node, anchorEnd int, afterReturn bool) []rule.RuleFix {
	firstToken := scanner.SkipTrivia(ctx.SourceFile.Text(), anchorEnd)
	requiresParens := utils.EslintLikePrecedence(expr) < 16 && !ast.IsParenthesizedExpression(expr.Parent)
	text := "void "
	if afterReturn && firstToken == anchorEnd {
		text = " " + text
	}
	if requiresParens {
		text += "("
	}
	fixes := []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(firstToken, firstToken), text)}
	if requiresParens {
		fixes = append(fixes, rule.RuleFixInsertAfter(expr, ")"))
	}
	return fixes
}

func callbackSuggestions(ctx rule.RuleContext, fn, expr *ast.Node, concise bool, allowVoid bool) []rule.RuleSuggestion {
	var suggestions []rule.RuleSuggestion
	anchorEnd := 0
	if concise {
		anchorEnd = fn.AsArrowFunction().EqualsGreaterThanToken.End()
		bodyStart := scanner.SkipTrivia(ctx.SourceFile.Text(), anchorEnd)
		prefix, suffix := "{", "}"
		// Function/class expressions become declarations at statement start.
		// A bare "use strict" string becomes a directive, which can also make
		// a non-simple parameter list illegal. Keep those bodies in expression
		// context, reusing any parentheses already present in the source.
		if !ast.IsParenthesizedExpression(expr.Parent) &&
			(expr.Kind == ast.KindFunctionExpression || expr.Kind == ast.KindClassExpression ||
				(expr.Kind == ast.KindStringLiteral && expr.Text() == "use strict")) {
			prefix, suffix = "{(", ")}"
		}
		suggestions = append(suggestions, rule.RuleSuggestion{Message: wrapBracesMessage, FixesArr: []rule.RuleFix{
			rule.RuleFixReplaceRange(core.NewTextRange(bodyStart, bodyStart), prefix),
			rule.RuleFixInsertAfter(fn, suffix),
		}})
	} else {
		// The expression may have several parenthesized wrappers. The enclosing
		// return still supplies the token before all of them.
		statement := utils.OutermostParenthesizedExpression(expr).Parent
		anchorEnd = utils.TrimNodeTextRange(ctx.SourceFile, statement).Pos() + len("return")
	}
	if allowVoid {
		suggestions = append(suggestions, rule.RuleSuggestion{Message: prependVoidMessage, FixesArr: prependVoidFixes(ctx, expr, anchorEnd, !concise)})
	}
	return suggestions
}

func checkCallback(ctx rule.RuleContext, fn *ast.Node, opts Options) {
	flags := ast.GetFunctionFlags(fn)
	if flags&(ast.FunctionFlagsAsync|ast.FunctionFlagsGenerator) != 0 {
		return
	}
	method := arrayMethodName(fn)
	if method == "" || (method == "forEach" && !opts.CheckForEach) {
		return
	}
	body := fn.Body()
	if body == nil {
		return
	}
	if body.Kind != ast.KindBlock {
		expr := ast.SkipParentheses(body)
		if method == "forEach" && (!opts.AllowVoid || expr.Kind != ast.KindVoidExpression) {
			ctx.ReportRangeWithDeferredSuggestions(utils.GetFunctionHeadLoc(ctx.SourceFile, fn), callbackMessage("expectedNoReturnValue", method, fn), func() []rule.RuleSuggestion {
				return callbackSuggestions(ctx, fn, expr, true, opts.AllowVoid)
			})
		}
		return
	}
	hasReturn := false
	ast.ForEachReturnStatement(body, func(statement *ast.Node) bool {
		hasReturn = true
		expr := statement.Expression()
		if expr != nil {
			expr = ast.SkipParentheses(expr)
		}
		if method == "forEach" {
			if expr != nil && (!opts.AllowVoid || expr.Kind != ast.KindVoidExpression) {
				ctx.ReportNodeWithDeferredSuggestions(statement, callbackMessage("expectedNoReturnValue", method, fn), func() []rule.RuleSuggestion {
					if !opts.AllowVoid {
						return nil
					}
					return callbackSuggestions(ctx, fn, expr, false, true)
				})
			}
		} else if !opts.AllowImplicit && expr == nil {
			ctx.ReportNode(statement, callbackMessage("expectedReturnValue", method, fn))
		}
		return false
	})
	if method != "forEach" && cfg.Build(fn, cfg.Hooks[struct{}]{}).EndReachable {
		id := "expectedInside"
		if hasReturn {
			id = "expectedAtEnd"
		}
		ctx.ReportRange(utils.GetFunctionHeadLoc(ctx.SourceFile, fn), callbackMessage(id, method, fn))
	}
}

var ArrayCallbackReturnRule = rule.Rule{
	Name:   "array-callback-return",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		check := func(node *ast.Node) { checkCallback(ctx, node, opts) }
		return rule.RuleListeners{ast.KindFunctionExpression: check, ast.KindArrowFunction: check}
	},
}
