package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type VisitModulesOptions struct {
	Commonjs bool
	AMD      bool
	ESModule bool
	// Ignore   []string
}

// See https://github.com/import-js/eslint-plugin-import/blob/01c9eb04331d2efa8d63f2d7f4bfec3bc44c94f3/utils/moduleVisitor.js
func VisitModules(visitor func(source *ast.StringLiteralLike, node *ast.Node), options VisitModulesOptions) rule.RuleListeners {
	visitors := rule.RuleListeners{}

	checkSourceValue := func(source *ast.StringLiteralLike, node *ast.Node) {
		if source == nil {
			return
		}

		// TODO: Handle options.Ignore

		visitor(source, node)
	}

	checkSource := func(node *ast.Node) {
		checkSourceValue(node.ModuleSpecifier(), node)
	}

	// for esmodule dynamic `import()` calls
	checkImportCall := func(node *ast.Node) {
		call := node.AsCallExpression()

		if call.Expression.Kind != ast.KindImportKeyword {
			return
		}

		modulePath := call.Arguments.Nodes[0]
		if modulePath == nil || !ast.IsStringLiteralLike(modulePath) {
			return
		}

		checkSourceValue(modulePath, call.AsNode())
	}

	// for CommonJS `require` calls
	checkCommon := func(call *ast.CallExpression) {
		if call.Expression.Kind != ast.KindIdentifier {
			return
		}

		callee := call.Expression.AsIdentifier()

		if callee.Text != "require" {
			return
		}

		if len(call.Arguments.Nodes) < 1 {
			return
		}

		modulePath := call.Arguments.Nodes[0]
		if modulePath == nil || !ast.IsStringLiteralLike(modulePath) {
			return
		}

		checkSourceValue(modulePath, call.AsNode())
	}

	checkAMD := func(node *ast.CallExpression) {
		// TODO: implement this later
	}

	if options.ESModule {
		visitors[ast.KindJSImportDeclaration] = checkSource
		visitors[ast.KindImportDeclaration] = checkSource
		visitors[ast.KindExportDeclaration] = checkSource
		visitors[ast.KindCallExpression] = checkImportCall
		// There is not `ImportExpression` in TypeScript
	}

	if options.Commonjs || options.AMD {
		currentCallExpression, ok := visitors[ast.KindCallExpression]

		visitors[ast.KindCallExpression] = func(node *ast.Node) {
			if ok {
				currentCallExpression(node)
			}
			if options.Commonjs {
				checkCommon(node.AsCallExpression())
			}
			if options.AMD {
				checkAMD(node.AsCallExpression())
			}
		}
	}

	return visitors
}
