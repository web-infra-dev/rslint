package func_names

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed func_names.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/func-names
var FuncNamesRule = rule.Rule{
	Name:   "func-names",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		handle := func(node *ast.Node) {
			handleFunction(ctx, opts, node)
		}

		return rule.RuleListeners{
			ast.KindFunctionExpression: handle,
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				// Upstream only listens on "ExportDefaultDeclaration >
				// FunctionDeclaration": a plain top-level FunctionDeclaration
				// always has a name syntactically, so it's never checked;
				// only the `export default function() {}` form (name
				// optional) needs one. tsgo parses that form directly as a
				// FunctionDeclaration carrying Export+Default modifiers,
				// with no intermediate export-declaration node to match on.
				if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExportDefault) {
					return
				}
				handle(node)
			},
		}
	},
}

type funcNamesOptions struct {
	mode           string
	generatorsMode string
}

func parseOptions(options []any) funcNamesOptions {
	opts := funcNamesOptions{mode: "always"}
	if len(options) > 0 {
		if s, ok := options[0].(string); ok {
			opts.mode = s
		}
	}
	if len(options) > 1 {
		if m, ok := options[1].(map[string]any); ok {
			if g, ok := m["generators"].(string); ok {
				opts.generatorsMode = g
			}
		}
	}
	return opts
}

func (o funcNamesOptions) configFor(node *ast.Node) string {
	if o.generatorsMode != "" && ast.GetFunctionFlags(node)&ast.FunctionFlagsGenerator != 0 {
		return o.generatorsMode
	}
	return o.mode
}

// isRecursiveNamedFunctionExpression reports whether node is a named
// FunctionExpression whose own recursion binding (the name that's visible
// only inside the function's own body, e.g. `foo` in `var a = function foo()
// { foo(); }`) is actually referenced. Upstream skips reporting entirely in
// this case, since removing the name would break the recursive call. tsgo's
// binder attaches that binding's symbol directly to the FunctionExpression
// node itself (node.Symbol()), so ctx.Refs.References can answer this with
// the same scope-correct resolution used everywhere else (a same-named inner
// declaration correctly shadows it).
//
// A FunctionDeclaration (the other node kind this rule handles, via the
// `export default function` form) never reaches this: getDeclaredVariables
// on it names the *outer* binding, whose reference count is irrelevant here,
// and the "never" branch already excludes FunctionDeclaration from reporting
// regardless.
func isRecursiveNamedFunctionExpression(ctx rule.RuleContext, node *ast.Node) bool {
	if node.Kind != ast.KindFunctionExpression || node.Name() == nil || ctx.Refs == nil {
		return false
	}
	sym := node.Symbol()
	if sym == nil {
		return false
	}
	return len(ctx.Refs.References(sym)) > 0
}

// hasInferredName mirrors upstream's hasInferredName: true when the function
// sits in a position where the ECMAScript NamedEvaluation semantics give it
// a `.name` derived from its assignment target. isObjectOrClassMethod is
// deliberately not ported here: it recognizes object-literal shorthand
// methods and get/set accessors (`{ foo() {} }`, `{ get foo() {} }`), which
// tsgo parses directly as MethodDeclaration/GetAccessor/SetAccessor — never
// as a FunctionExpression with such a parent (see AST_PATTERNS.md's
// FunctionExpression/MethodDeclaration split), so the check can never be
// true for a node this rule's FunctionExpression listener sees.
func hasInferredName(node *ast.Node) bool {
	parent := ast.WalkUpParenthesizedExpressions(node.Parent)
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindVariableDeclaration:
		vd := parent.AsVariableDeclaration()
		return parent.Name().Kind == ast.KindIdentifier &&
			vd.Initializer != nil && ast.SkipParentheses(vd.Initializer) == node
	case ast.KindPropertyAssignment:
		pa := parent.AsPropertyAssignment()
		return pa.Initializer != nil && ast.SkipParentheses(pa.Initializer) == node
	case ast.KindPropertyDeclaration:
		// A TS auto-accessor (`accessor foo = ...`) is also a
		// PropertyDeclaration in tsgo, but upstream parses it as
		// AccessorProperty rather than PropertyDefinition, so its initializer
		// is not an inferred-name context.
		pd := parent.AsPropertyDeclaration()
		return !ast.HasSyntacticModifier(parent, ast.ModifierFlagsAccessor) &&
			pd.Initializer != nil && ast.SkipParentheses(pd.Initializer) == node
	case ast.KindShorthandPropertyAssignment:
		spa := parent.AsShorthandPropertyAssignment()
		return spa.ObjectAssignmentInitializer != nil &&
			ast.SkipParentheses(spa.ObjectAssignmentInitializer) == node
	case ast.KindParameter:
		p := parent.AsParameterDeclaration()
		return parent.Name().Kind == ast.KindIdentifier &&
			p.Initializer != nil && ast.SkipParentheses(p.Initializer) == node
	case ast.KindBindingElement:
		be := parent.AsBindingElement()
		return parent.Name().Kind == ast.KindIdentifier &&
			be.Initializer != nil && ast.SkipParentheses(be.Initializer) == node
	case ast.KindBinaryExpression:
		// Covers a plain or compound assignment (`foo = function(){}`,
		// `foo ??= function(){}`) and a destructuring default inside a
		// non-shorthand property or array element (`{key: foo =
		// function(){}}`, `[foo = function(){}]`, always plain `=`): tsgo
		// parses all of these as BinaryExpression. Upstream's
		// AssignmentExpression check doesn't gate on operator either — see
		// the "quux ??= function() {};" as-needed example in the rule doc.
		be := parent.AsBinaryExpression()
		return ast.IsAssignmentOperator(be.OperatorToken.Kind) &&
			ast.SkipParentheses(be.Left).Kind == ast.KindIdentifier &&
			ast.SkipParentheses(be.Right) == node
	}
	return false
}

func handleFunction(ctx rule.RuleContext, opts funcNamesOptions, node *ast.Node) {
	if isRecursiveNamedFunctionExpression(ctx, node) {
		return
	}

	hasName := node.Name() != nil
	config := opts.configFor(node)

	switch config {
	case "never":
		if hasName && node.Kind != ast.KindFunctionDeclaration {
			report(ctx, node, "named", "Unexpected named %s.")
		}
	case "as-needed":
		if !hasName && !hasInferredName(node) {
			report(ctx, node, "unnamed", "Unexpected unnamed %s.")
		}
	default: // "always"
		if !hasName {
			report(ctx, node, "unnamed", "Unexpected unnamed %s.")
		}
	}
}

func report(ctx rule.RuleContext, node *ast.Node, messageId string, format string) {
	name := utils.GetFunctionNameWithKindCore(node)
	loc := utils.GetFunctionHeadLoc(ctx.SourceFile, node)
	ctx.ReportRange(loc, rule.RuleMessage{
		Id:          messageId,
		Description: fmt.Sprintf(format, name),
		Data:        map[string]string{"name": name},
	})
}
