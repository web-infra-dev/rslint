package handle_callback_err

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed handle_callback_err.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/handle-callback-err.js
var HandleCallbackErrRule = rule.Rule{
	Name:   "node/handle-callback-err",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		name := "err"
		if len(options) > 0 {
			if value, ok := options[0].(string); ok && value != "" {
				name = value
			}
		}
		var pattern *esregexp.RegExp
		if strings.HasPrefix(name, "^") {
			var err error
			pattern, err = esregexp.Compile(name, "u")
			if err != nil {
				return nil
			}
		}

		check := func(node *ast.Node) {
			if node.Body() == nil {
				return
			}
			// Upstream chooses the first bound name, including destructuring,
			// rather than requiring the first parameter to be an identifier.
			var first *ast.Node
			for _, parameter := range node.Parameters() {
				utils.CollectBindingNames(parameter.Name(), func(id *ast.Node, _ string) {
					if first == nil {
						first = id
					}
				})
				if first != nil {
					break
				}
			}
			if first == nil || (pattern == nil && first.Text() != name) ||
				(pattern != nil && !pattern.Test(first.Text())) {
				return
			}
			// Parameter properties also have a field symbol; use the local binding.
			symbol := node.Locals()[first.Text()]
			// RefStore excludes declaration writes. ESLint counts defaults and
			// initialized redeclarations as references, even without a read.
			if symbol != nil {
				for _, declaration := range symbol.Declarations {
					if hasInitialization(declaration, node, first.Text()) {
						return
					}
				}
			}
			if len(ctx.Refs.References(symbol)) != 0 {
				return
			}

			ctx.ReportRange(utils.ESTreeFunctionRange(ctx.SourceFile, node), rule.RuleMessage{
				Id:          "expected",
				Description: "Expected error to be handled.",
			})
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: check,
			ast.KindFunctionExpression:  check,
			ast.KindArrowFunction:       check,
			ast.KindMethodDeclaration:   check,
			ast.KindConstructor:         check,
			ast.KindGetAccessor:         check,
			ast.KindSetAccessor:         check,
		}
	},
}

func hasInitialization(declaration, function *ast.Node, name string) bool {
	initialized := false
	for node := declaration; node != nil && node != function; node = node.Parent {
		switch node.Kind {
		case ast.KindBindingElement, ast.KindParameter, ast.KindVariableDeclaration:
			if node.Initializer() != nil {
				initialized = true
			}
			if node.Kind == ast.KindVariableDeclaration {
				loop := node.Parent.Parent
				initialized = initialized || loop.Kind == ast.KindForInStatement || loop.Kind == ast.KindForOfStatement
			}
		case ast.KindCatchClause:
			// A hoisted var can write a catch binding instead of the parameter.
			binding := node.AsCatchClause().VariableDeclaration
			if binding != nil && utils.HasNameInBindingPattern(binding.Name(), name) {
				return false
			}
		}
	}
	return initialized
}
