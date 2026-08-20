package no_inner_declarations

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_inner_declarations.schema.json
var schemaJSON []byte

type ruleOptions struct {
	both                 bool
	blockScopedFunctions string // "allow" or "disallow"
}

// https://eslint.org/docs/latest/rules/no-inner-declarations
var NoInnerDeclarationsRule = rule.Rule{
	Name:   "no-inner-declarations",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		allowBlockScopedFunctions := opts.blockScopedFunctions == "allow" &&
			ctx.LanguageOptions.EffectiveECMAVersion() >= 2015

		listeners := rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				// @typescript-eslint/parser represents overload and ambient
				// signatures as TSDeclareFunction, which the upstream
				// FunctionDeclaration listener never receives. tsgo represents
				// those signatures as body-less FunctionDeclaration nodes.
				if node.Body() == nil {
					return
				}
				if allowBlockScopedFunctions && utils.IsInStrictMode(node, ctx.SourceFile) {
					return
				}
				check(node, "function", &ctx)
			},
		}

		if opts.both {
			listeners[ast.KindVariableDeclarationList] = func(node *ast.Node) {
				if !utils.IsVarKeyword(node) {
					return
				}

				// ESTree uses VariableDeclaration for both declaration
				// statements and for-loop initializers. tsgo wraps statement
				// declarations in VariableStatement but leaves loop initializers
				// as a bare VariableDeclarationList. Preserve ESLint's report
				// range by selecting the corresponding outer node only when it
				// exists.
				reportNode := node
				if node.Parent != nil && node.Parent.Kind == ast.KindVariableStatement {
					reportNode = node.Parent
				}

				check(reportNode, "variable", &ctx)
			}
		}

		return listeners
	},
}

func parseOptions(options []any) ruleOptions {
	opts := ruleOptions{
		both:                 false,
		blockScopedFunctions: "allow", // default: allow block-scoped functions (ES2015+)
	}

	if len(options) == 0 {
		return opts
	}
	if v, _ := options[0].(string); v == "both" {
		opts.both = true
	}
	if len(options) < 2 {
		return opts
	}
	m, _ := options[1].(map[string]any)
	if v, ok := m["blockScopedFunctions"].(string); ok {
		opts.blockScopedFunctions = v
	}

	return opts
}

// isValidParent checks whether the declaration's immediate parent represents
// a valid ("root") position for function/var declarations.
func isValidParent(parent *ast.Node) bool {
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindSourceFile:
		return true
	case ast.KindBlock:
		// A block is valid only if its parent is a function-like node or
		// a class static block.
		gp := parent.Parent
		if gp == nil {
			return false
		}
		switch gp.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			return true
		case ast.KindClassStaticBlockDeclaration:
			return true
		}
		return false
	}
	return false
}

// nearestFunctionName walks up the tree to find the enclosing function (if any)
// and returns a description used in the error message.
func nearestFunctionName(node *ast.Node) string {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindClassStaticBlockDeclaration:
			return "class static block body"
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			return "function body"
		}
		current = current.Parent
	}
	return "program"
}

func check(node *ast.Node, declType string, ctx *rule.RuleContext) {
	parent := node.Parent
	if isValidParent(parent) {
		return
	}

	body := nearestFunctionName(node)

	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "moveDeclToRoot",
		Description: fmt.Sprintf("Move %s declaration to %s root.", declType, body),
	})
}
