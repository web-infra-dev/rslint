package no_inner_declarations

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ruleOptions struct {
	both                 bool
	blockScopedFunctions string // "allow" or "disallow"
}

// https://eslint.org/docs/latest/rules/no-inner-declarations
var NoInnerDeclarationsRule = rule.Rule{
	Name: "no-inner-declarations",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		listeners := rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				if opts.blockScopedFunctions == "allow" {
					return
				}
				check(node, "function", &ctx)
			},
		}

		if opts.both {
			listeners[ast.KindVariableStatement] = func(node *ast.Node) {
				varStmt := node.AsVariableStatement()
				if varStmt == nil || varStmt.DeclarationList == nil {
					return
				}

				// Only check var declarations, not let/const/using
				// BlockScoped = Let | Const | Using
				if varStmt.DeclarationList.Flags&ast.NodeFlagsBlockScoped != 0 {
					return
				}

				check(node, "variable", &ctx)
			}
		}

		return listeners
	},
}

func parseOptions(opts any) ruleOptions {
	result := ruleOptions{
		both:                 false,
		blockScopedFunctions: "allow", // default: allow block-scoped functions (ES2015+)
	}

	if opts == nil {
		return result
	}

	// ESLint format: options can be:
	//   ["both"] or ["functions"]  (first element is a string)
	//   ["both", { blockScopedFunctions: "disallow" }]  (second element is an object)
	//   "both" or "functions"  (unwrapped by rule_tester when single element)
	//   { blockScopedFunctions: "disallow" }  (object only, when first arg is default)

	switch v := opts.(type) {
	case string:
		if v == "both" {
			result.both = true
		}
	case []interface{}:
		if len(v) > 0 {
			if str, ok := v[0].(string); ok && str == "both" {
				result.both = true
			}
		}
		if len(v) > 1 {
			if obj, ok := v[1].(map[string]interface{}); ok {
				if bsf, ok := obj["blockScopedFunctions"].(string); ok {
					result.blockScopedFunctions = bsf
				}
			}
		}
	case map[string]interface{}:
		if bsf, ok := v["blockScopedFunctions"].(string); ok {
			result.blockScopedFunctions = bsf
		}
	}

	return result
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
	case ast.KindModuleBlock:
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
