package prefer_es6_class

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// parseMode reads the first option, accepting "always" (default) or "never".
// Any other shape falls back to "always" — matches eslint-plugin-react's
// `context.options[0] || 'always'`.
func parseMode(options any) string {
	mode := "always"
	switch opts := options.(type) {
	case []interface{}:
		if len(opts) > 0 {
			if s, ok := opts[0].(string); ok && s != "" {
				mode = s
			}
		}
	case string:
		if opts != "" {
			mode = opts
		}
	}
	if mode != "always" && mode != "never" {
		mode = "always"
	}
	return mode
}

// skipParenParent walks up ParenthesizedExpression wrappers and returns the
// first non-paren ancestor of `node`, or nil. ESTree flattens parentheses so
// ESLint's `node.parent` is always the first non-paren ancestor; tsgo
// preserves them.
func skipParenParent(node *ast.Node) *ast.Node {
	p := node.Parent
	for p != nil && p.Kind == ast.KindParenthesizedExpression {
		p = p.Parent
	}
	return p
}

// isCreateClassCallee reports whether `callee` names `<createClass>` or
// `<pragma>.<createClass>`. Parentheses are skipped on both the callee
// itself and the pragma identifier.
func isCreateClassCallee(callee *ast.Node, pragma, createClass string) bool {
	if callee == nil {
		return false
	}
	callee = ast.SkipParentheses(callee)
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.AsIdentifier().Text == createClass
	case ast.KindPropertyAccessExpression:
		pa := callee.AsPropertyAccessExpression()
		obj := ast.SkipParentheses(pa.Expression)
		if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != pragma {
			return false
		}
		name := pa.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		return name.AsIdentifier().Text == createClass
	}
	return false
}

// classKeywordStart returns the start position of the ESLint-equivalent
// ClassDeclaration range.
//
// In ESTree the ClassDeclaration itself does NOT include `export` /
// `export default` — those live on an outer `ExportNamedDeclaration` /
// `ExportDefaultDeclaration` wrapper. TS-specific modifiers that stay on the
// ClassDeclaration in TSESTree (`abstract`, `declare`) and decorators are
// part of the node's range. tsgo inlines everything into the
// ClassDeclaration's modifier list, so we selectively skip only
// `export` / `default` keyword modifiers to recover the ESLint position.
func classKeywordStart(text string, node *ast.Node) int {
	mods := node.Modifiers()
	pos := node.Pos()
	if mods != nil {
		for _, mod := range mods.Nodes {
			switch mod.Kind {
			case ast.KindExportKeyword, ast.KindDefaultKeyword:
				pos = mod.End()
			default:
				// First non-export/default modifier (decorator, abstract,
				// declare, …) is where the ESLint range begins.
				return scanner.SkipTrivia(text, mod.Pos())
			}
		}
	}
	return scanner.SkipTrivia(text, pos)
}

func containsArg(args *ast.NodeList, target *ast.Node) bool {
	if args == nil {
		return false
	}
	for _, a := range args.Nodes {
		// The argument may be wrapped in one or more
		// ParenthesizedExpressions before reaching the object literal
		// (ESTree flattens these; tsgo preserves them).
		current := a
		for current != nil && current.Kind == ast.KindParenthesizedExpression {
			current = current.AsParenthesizedExpression().Expression
		}
		if current == target {
			return true
		}
	}
	return false
}

var PreferEs6ClassRule = rule.Rule{
	Name: "react/prefer-es6-class",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		mode := parseMode(options)
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)

		listeners := rule.RuleListeners{}

		if mode == "always" {
			listeners[ast.KindObjectLiteralExpression] = func(node *ast.Node) {
				parent := skipParenParent(node)
				if parent == nil {
					return
				}
				// ESLint's `isES5Component` gate is `node.parent.callee` —
				// any parent node kind that exposes a `.callee` property.
				// In ESTree that is both CallExpression AND NewExpression,
				// so `new createReactClass({...})` also reports.
				var callee *ast.Node
				var arguments *ast.NodeList
				switch parent.Kind {
				case ast.KindCallExpression:
					call := parent.AsCallExpression()
					callee = call.Expression
					arguments = call.Arguments
				case ast.KindNewExpression:
					newExpr := parent.AsNewExpression()
					callee = newExpr.Expression
					arguments = newExpr.Arguments
				default:
					return
				}
				if !isCreateClassCallee(callee, pragma, createClass) {
					return
				}
				// Belt-and-suspenders: make sure the object is an argument
				// (i.e. not somehow the callee itself); in practice the
				// callee check above already rejects object callees.
				if !containsArg(arguments, node) {
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "shouldUseES6Class",
					Description: "Component should use es6 class instead of createClass",
				})
			}
		}

		if mode == "never" {
			// Upstream only subscribes to ClassDeclaration; ClassExpression is
			// not reported even when it extends React.Component. Matching that
			// behavior — lock-in test in the Go test file.
			listeners[ast.KindClassDeclaration] = func(node *ast.Node) {
				if !reactutil.ExtendsReactComponent(node, pragma) {
					return
				}
				// Report from the `class` keyword (not from any preceding
				// modifiers like `export` / `export default`). In ESTree the
				// ClassDeclaration node starts at `class` and modifiers live on
				// the outer Export* wrapper; tsgo inlines modifiers into the
				// ClassDeclaration, so we skip past them to match ESLint's
				// report position.
				classStart := classKeywordStart(ctx.SourceFile.Text(), node)
				ctx.ReportRange(core.NewTextRange(classStart, node.End()), rule.RuleMessage{
					Id:          "shouldUseCreateClass",
					Description: "Component should use createClass instead of es6 class",
				})
			}
		}

		return listeners
	},
}
