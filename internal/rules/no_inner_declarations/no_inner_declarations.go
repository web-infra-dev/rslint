package no_inner_declarations

import (
	_ "embed"

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

type declarationType uint8

const (
	functionDeclaration declarationType = iota
	variableDeclaration
)

type declarationRoot uint8

const (
	programRoot declarationRoot = iota
	functionBodyRoot
	classStaticBlockBodyRoot
)

var moveDeclarationMessages = [2][3]rule.RuleMessage{
	functionDeclaration: {
		programRoot:              {Id: "moveDeclToRoot", Description: "Move function declaration to program root."},
		functionBodyRoot:         {Id: "moveDeclToRoot", Description: "Move function declaration to function body root."},
		classStaticBlockBodyRoot: {Id: "moveDeclToRoot", Description: "Move function declaration to class static block body root."},
	},
	variableDeclaration: {
		programRoot:              {Id: "moveDeclToRoot", Description: "Move variable declaration to program root."},
		functionBodyRoot:         {Id: "moveDeclToRoot", Description: "Move variable declaration to function body root."},
		classStaticBlockBodyRoot: {Id: "moveDeclToRoot", Description: "Move variable declaration to class static block body root."},
	},
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
				check(node, functionDeclaration, &ctx)
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

				check(reportNode, variableDeclaration, &ctx)
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

// nearestDeclarationRoot walks up the tree to find the declaration's expected
// root container for the diagnostic message.
func nearestDeclarationRoot(node *ast.Node) declarationRoot {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindClassStaticBlockDeclaration:
			return classStaticBlockBodyRoot
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			return functionBodyRoot
		}
		current = current.Parent
	}
	return programRoot
}

func check(node *ast.Node, declaration declarationType, ctx *rule.RuleContext) {
	parent := node.Parent
	if isValidParent(parent) {
		return
	}

	ctx.ReportNode(node, moveDeclarationMessages[declaration][nearestDeclarationRoot(node)])
}
