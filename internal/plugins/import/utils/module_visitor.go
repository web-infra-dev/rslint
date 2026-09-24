package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

type VisitModulesOptions struct {
	Commonjs bool
	AMD      bool
	ESModule bool
	Ignore   []string
}

// See https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/utils/moduleVisitor.js
func VisitModules(visitor func(source *ast.StringLiteralLike, node *ast.Node), options VisitModulesOptions) rule.RuleListeners {
	visitors := rule.RuleListeners{}
	ignored := make([]*esregexp.RegExp, 0, len(options.Ignore))
	for _, pattern := range options.Ignore {
		ignored = append(ignored, esregexp.MustCompile(pattern, ""))
	}

	checkSourceValue := func(source *ast.StringLiteralLike, node *ast.Node) {
		if source == nil {
			return
		}

		for _, pattern := range ignored {
			if pattern.TestOrTimeout(source.Text()) {
				return
			}
		}

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

		// A recovered parse of incomplete source can leave `import()` with no
		// arguments at all.
		if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
			return
		}

		modulePath := utils.ESTreeRuntimeExpression(call.Arguments.Nodes[0])
		// Upstream accepts string Literals, not static TemplateLiterals.
		if modulePath == nil || modulePath.Kind != ast.KindStringLiteral {
			return
		}

		checkSourceValue(modulePath, call.AsNode())
	}

	// for CommonJS `require` calls
	checkCommon := func(call *ast.CallExpression) {
		// ESTree has no parenthesized-expression node, so upstream sees a bare
		// `require` identifier through any number of parentheses.
		callee := utils.ESTreeCallCallee(call.Expression)

		if callee == nil || !ast.IsIdentifier(callee) {
			return
		}

		if callee.AsIdentifier().Text != "require" {
			return
		}

		if call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
			return
		}

		modulePath := utils.ESTreeRuntimeExpression(call.Arguments.Nodes[0])
		if modulePath == nil || modulePath.Kind != ast.KindStringLiteral {
			return
		}

		checkSourceValue(modulePath, call.AsNode())
	}

	checkAMD := func(call *ast.CallExpression) {
		callee := utils.ESTreeCallCallee(call.Expression)
		if callee == nil || !ast.IsIdentifier(callee) ||
			(callee.Text() != "require" && callee.Text() != "define") ||
			call.Arguments == nil || len(call.Arguments.Nodes) != 2 {
			return
		}
		modules := utils.ESTreeRuntimeExpression(call.Arguments.Nodes[0])
		if modules == nil || modules.Kind != ast.KindArrayLiteralExpression {
			return
		}
		for _, element := range modules.AsArrayLiteralExpression().Elements.Nodes {
			source := utils.ESTreeRuntimeExpression(element)
			if source == nil || source.Kind != ast.KindStringLiteral ||
				source.Text() == "require" || source.Text() == "exports" {
				continue
			}
			checkSourceValue(source, source)
		}
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
