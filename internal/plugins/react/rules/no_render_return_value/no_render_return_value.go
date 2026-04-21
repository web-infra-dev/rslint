package no_render_return_value

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// calleeObjectRe selects the acceptable callee-object names by React version.
// Mirrors upstream's if/else-if chain:
//
//	default / >= 15.0.0 → ReactDOM
//	^0.14.0             → React or ReactDOM
//	^0.13.0             → React
//
// Anything else (e.g. 0.0.1) falls back to the default ReactDOM, matching
// upstream's uninitialized-branch behavior.
var (
	reReactDOM   = regexp.MustCompile(`^ReactDOM$`)
	reReactOrDOM = regexp.MustCompile(`^React(DOM)?$`)
	reReact      = regexp.MustCompile(`^React$`)
)

func calleeObjectRe(settings map[string]interface{}) *regexp.Regexp {
	// >= 15.0.0 — the default branch when version is unset (ParseReactVersion
	// returns 999,999,999) and therefore ≥ 15.
	if !reactutil.ReactVersionLessThan(settings, 15, 0, 0) {
		return reReactDOM
	}
	// ^0.14.0 → [0.14.0, 0.15.0)
	if !reactutil.ReactVersionLessThan(settings, 0, 14, 0) &&
		reactutil.ReactVersionLessThan(settings, 0, 15, 0) {
		return reReactOrDOM
	}
	// ^0.13.0 → [0.13.0, 0.14.0)
	if !reactutil.ReactVersionLessThan(settings, 0, 13, 0) &&
		reactutil.ReactVersionLessThan(settings, 0, 14, 0) {
		return reReact
	}
	return reReactDOM
}

// matchedObjectName reports the Identifier text of the call's callee object
// when the callee shape is `<Identifier>.render` (or `<Identifier>[render]`
// where `render` is an Identifier reference literally named "render") AND the
// identifier matches the version-selected pattern. Returns "" otherwise.
//
// Two bracket-access sub-cases mirror upstream's `'name' in callee.property`
// guard exactly:
//
//   - `ReactDOM['render']()` — property is a StringLiteral; ESTree Literal
//     has no `.name`, so `'name' in property` is false → upstream skips. We
//     skip too.
//   - `ReactDOM[render]()` — property is an Identifier reference literally
//     named "render"; ESTree Identifier has `.name === 'render'` → upstream
//     triggers. We trigger too.
//
// Parentheses on the callee, the pragma identifier, and the bracket argument
// are transparently skipped.
func matchedObjectName(call *ast.CallExpression, pattern *regexp.Regexp) string {
	callee := ast.SkipParentheses(call.Expression)
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		prop := callee.AsPropertyAccessExpression()
		obj := ast.SkipParentheses(prop.Expression)
		if obj.Kind != ast.KindIdentifier {
			return ""
		}
		name := obj.AsIdentifier().Text
		if !pattern.MatchString(name) {
			return ""
		}
		nameNode := prop.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return ""
		}
		if nameNode.AsIdentifier().Text != "render" {
			return ""
		}
		return name
	case ast.KindElementAccessExpression:
		ea := callee.AsElementAccessExpression()
		obj := ast.SkipParentheses(ea.Expression)
		if obj.Kind != ast.KindIdentifier {
			return ""
		}
		name := obj.AsIdentifier().Text
		if !pattern.MatchString(name) {
			return ""
		}
		// ESTree's `'name' in callee.property` test: only Identifier nodes
		// have `.name`. Literal (StringLiteral / NumericLiteral) and
		// computed key expressions do not, so upstream silently skips them.
		// Mirror by accepting ONLY Identifier here — even if its referent
		// happens to evaluate to "render" at runtime, upstream doesn't see
		// that, and neither do we.
		arg := ast.SkipParentheses(ea.ArgumentExpression)
		if arg == nil || arg.Kind != ast.KindIdentifier {
			return ""
		}
		if arg.AsIdentifier().Text != "render" {
			return ""
		}
		return name
	}
	return ""
}

// NoRenderReturnValueRule flags uses of the return value of `ReactDOM.render`
// (or `React.render` on legacy 0.13/0.14 React). The rule fires only when the
// call sits in a position that consumes the return value — `var x = ...`,
// `{ k: ... }`, `return ...`, an arrow expression body, or the RHS of an
// assignment — matching upstream exactly.
var NoRenderReturnValueRule = rule.Rule{
	Name: "react/no-render-return-value",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		pattern := calleeObjectRe(ctx.Settings)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				objName := matchedObjectName(call, pattern)
				if objName == "" {
					return
				}
				// Walk up ParenthesizedExpression wrappers transparently —
				// ESTree flattens parens, so upstream's `parent.type` check
				// already sees the non-paren ancestor. Equivalent to
				// upstream's implicit flattening.
				parent := ast.WalkUpParenthesizedExpressions(node.Parent)
				if parent == nil {
					return
				}
				if !consumesReturnValue(parent, node) {
					return
				}
				ctx.ReportNode(ast.SkipParentheses(call.Expression), rule.RuleMessage{
					Id:          "noReturnValue",
					Description: "Do not depend on the return value from " + objName + ".render",
				})
			},
		}
	},
}

// consumesReturnValue mirrors upstream's parent-type allow-list, mapped onto
// tsgo's AST:
//
//	ESTree                     | tsgo
//	---------------------------|-----
//	VariableDeclarator         | VariableDeclaration
//	Property                   | PropertyAssignment
//	ReturnStatement            | ReturnStatement (same)
//	ArrowFunctionExpression    | ArrowFunction, only when `call` is the body
//	AssignmentExpression       | BinaryExpression with an assignment operator
//	                           |   (covers compound / logical assignment too,
//	                           |   matching ESTree's single AssignmentExpression
//	                           |   type)
//
// `call` is the original CallExpression (before the caller's paren-walk); for
// the arrow-body case, ParenthesizedExpression wrappers around the body are
// skipped so `(a) => (call())` reaches the same ArrowFunction.
func consumesReturnValue(parent *ast.Node, call *ast.Node) bool {
	switch parent.Kind {
	case ast.KindVariableDeclaration,
		ast.KindPropertyAssignment,
		ast.KindReturnStatement:
		return true
	case ast.KindArrowFunction:
		af := parent.AsArrowFunction()
		if af.Body == nil {
			return false
		}
		return ast.SkipParentheses(af.Body) == call
	case ast.KindBinaryExpression:
		// `ast.IsAssignmentExpression(parent, false)` covers `=`, `+=`, `-=`,
		// `*=`, `/=`, `%=`, `**=`, `<<=`, `>>=`, `>>>=`, `&=`, `|=`, `^=`,
		// `&&=`, `||=`, `??=` — matching ESTree's single AssignmentExpression
		// type. `false` means "include compound assignments".
		return ast.IsAssignmentExpression(parent, false)
	}
	return false
}
