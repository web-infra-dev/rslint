package func_style

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed func_style.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/func-style
var FuncStyleRule = rule.Rule{
	Name:   "func-style",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		enforceDeclarations := opts.style == "declaration"

		// Tracks, for each function-like node currently being visited, whether
		// a direct `this`/`super` reference was seen inside it. Only an arrow
		// function's own frame is ever read (an arrow bound to a variable that
		// itself needs `this`/`super` can't become a declaration), but every
		// function-like listener pushes/pops a frame so a nested function's
		// usage doesn't leak into an enclosing arrow's frame.
		var stack []bool

		reportExpression := func(node *ast.Node) {
			start := utils.FindFunctionKeywordPos(ctx.SourceFile, node)
			ctx.ReportRange(core.NewTextRange(start, node.End()), rule.RuleMessage{
				Id:          "expression",
				Description: "Expected a function expression.",
			})
		}

		reportDeclaration := func(declarator *ast.Node) {
			ctx.ReportNode(declarator, rule.RuleMessage{
				Id:          "declaration",
				Description: "Expected a function declaration.",
			})
		}

		checkAssignedToVariable := func(node *ast.Node) {
			// ESTree flattens parentheses, so `var foo = (function(){})`
			// still has a VariableDeclarator parent for the FunctionExpression
			// upstream visits. tsgo instead wraps it in an explicit
			// ParenthesizedExpression, so the declarator has to be found by
			// walking back up through any such wrapper(s) first.
			declarator := ast.WalkUpParenthesizedExpressions(node.Parent)
			if declarator == nil || declarator.Kind != ast.KindVariableDeclaration {
				return
			}
			if opts.allowTypeAnnotation && declarator.AsVariableDeclaration().Type != nil {
				return
			}
			isNamed := isNamedExportedDeclarator(declarator)

			if enforceDeclarations && (opts.namedExports == "" || !isNamed) {
				reportDeclaration(declarator)
			}
			if isNamed && opts.namedExports == "declaration" {
				reportDeclaration(declarator)
			}
		}

		listeners := rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				fd := node.AsFunctionDeclaration()
				if fd.Body == nil {
					// Overload signature or ambient `declare function`: ESLint
					// parses these as TSDeclareFunction, a node kind with no
					// listener attached, so upstream never visits them.
					return
				}
				stack = append(stack, false)

				if !enforceDeclarations && !isDefaultExport(node) &&
					(opts.namedExports == "" || !isNamedExport(node)) &&
					!isOverloadedFunction(node) {
					reportExpression(node)
				}
				if isNamedExport(node) && opts.namedExports == "expression" && !isOverloadedFunction(node) {
					reportExpression(node)
				}
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
				if node.AsFunctionDeclaration().Body == nil {
					return
				}
				stack = stack[:len(stack)-1]
			},

			ast.KindThisKeyword: func(node *ast.Node) {
				if len(stack) > 0 {
					stack[len(stack)-1] = true
				}
			},
			ast.KindSuperKeyword: func(node *ast.Node) {
				if len(stack) > 0 {
					stack[len(stack)-1] = true
				}
			},
		}

		// ESTree has no MethodDefinition/Property-method node kind of its
		// own: an object or class method's body lives on a FunctionExpression
		// value, so upstream's FunctionExpression listener fires for it too,
		// pushing a frame that isolates the method's own `this`/`super`
		// binding from an enclosing arrow. tsgo instead parses these as their
		// own dedicated kinds, never as KindFunctionExpression, so each needs
		// the identical enter/exit treatment here or a `super`/`this` used
		// only inside a nested method would wrongly mark an outer arrow's
		// frame as needing `this`/`super`.
		enterFunctionExpressionLike := func(node *ast.Node) {
			stack = append(stack, false)
			checkAssignedToVariable(node)
		}
		exitFunctionExpressionLike := func(node *ast.Node) {
			stack = stack[:len(stack)-1]
		}
		for _, kind := range []ast.Kind{
			ast.KindFunctionExpression,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor,
		} {
			listeners[kind] = enterFunctionExpressionLike
			listeners[rule.ListenerOnExit(kind)] = exitFunctionExpressionLike
		}

		if !opts.allowArrowFunctions {
			listeners[ast.KindArrowFunction] = func(node *ast.Node) {
				stack = append(stack, false)
			}
			listeners[rule.ListenerOnExit(ast.KindArrowFunction)] = func(node *ast.Node) {
				hasThisOrSuper := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if hasThisOrSuper {
					return
				}
				checkAssignedToVariable(node)
			}
		}

		return listeners
	},
}

type funcStyleOptions struct {
	style               string
	allowArrowFunctions bool
	allowTypeAnnotation bool
	namedExports        string
}

func parseOptions(options []any) funcStyleOptions {
	opts := funcStyleOptions{style: "expression"}
	if len(options) > 0 {
		if s, ok := options[0].(string); ok {
			opts.style = s
		}
	}
	if len(options) > 1 {
		if m, ok := options[1].(map[string]any); ok {
			if v, ok := m["allowArrowFunctions"].(bool); ok {
				opts.allowArrowFunctions = v
			}
			if v, ok := m["allowTypeAnnotation"].(bool); ok {
				opts.allowTypeAnnotation = v
			}
			if overrides, ok := m["overrides"].(map[string]any); ok {
				if v, ok := overrides["namedExports"].(string); ok {
					opts.namedExports = v
				}
			}
		}
	}
	return opts
}

// isDefaultExport reports whether node itself carries both the `export` and
// `default` modifiers (`export default function foo() {}`).
func isDefaultExport(node *ast.Node) bool {
	flags := node.ModifierFlags()
	return flags&ast.ModifierFlagsExport != 0 && flags&ast.ModifierFlagsDefault != 0
}

// isNamedExport reports whether node carries `export` without `default`.
func isNamedExport(node *ast.Node) bool {
	flags := node.ModifierFlags()
	return flags&ast.ModifierFlagsExport != 0 && flags&ast.ModifierFlagsDefault == 0
}

// isNamedExportedDeclarator reports whether declarator (a VariableDeclaration
// node, i.e. an ESTree VariableDeclarator) sits inside a named-exported
// `export var/let/const ...` statement. tsgo attaches the `export` modifier to
// the enclosing VariableStatement, not to the declarator itself.
func isNamedExportedDeclarator(declarator *ast.Node) bool {
	list := declarator.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList {
		return false
	}
	stmt := list.Parent
	if stmt == nil || stmt.Kind != ast.KindVariableStatement {
		return false
	}
	return isNamedExport(stmt)
}

// isOverloadedFunction reports whether node (a FunctionDeclaration with a
// body) is the implementation of a set of overload signatures: one or more
// same-named, body-less FunctionDeclarations declared alongside it. ESLint
// parses those signatures as TSDeclareFunction and skips reporting on the
// implementation entirely, since an overloaded function can't be rewritten as
// an expression. tsgo parses every signature as an ordinary
// KindFunctionDeclaration with a nil Body, so the port re-derives the same
// "is this name overloaded" answer by scanning siblings instead.
func isOverloadedFunction(node *ast.Node) bool {
	nameNode := node.Name()
	if nameNode == nil {
		return false
	}
	name := nameNode.Text()

	parent := node.Parent
	if parent == nil {
		return false
	}

	// A `switch` case's overload signatures may be spread across sibling
	// cases (falling through to a shared implementation), so the search
	// covers every clause of the enclosing switch, not just the current one.
	if parent.Kind == ast.KindCaseClause || parent.Kind == ast.KindDefaultClause {
		caseBlock := parent.Parent
		if caseBlock == nil || caseBlock.Kind != ast.KindCaseBlock {
			return false
		}
		for _, clause := range caseBlock.AsCaseBlock().Clauses.Nodes {
			if hasOverloadSignature(clause.Statements(), name) {
				return true
			}
		}
		return false
	}

	if !parent.CanHaveStatements() {
		return false
	}
	return hasOverloadSignature(parent.Statements(), name)
}

func hasOverloadSignature(statements []*ast.Node, name string) bool {
	for _, stmt := range statements {
		if stmt.Kind != ast.KindFunctionDeclaration {
			continue
		}
		fd := stmt.AsFunctionDeclaration()
		if fd.Body != nil {
			continue
		}
		if stmt.Name() != nil && stmt.Name().Text() == name {
			return true
		}
	}
	return false
}
