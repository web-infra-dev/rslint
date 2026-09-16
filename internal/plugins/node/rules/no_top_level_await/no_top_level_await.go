package no_top_level_await

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_top_level_await.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-top-level-await.js
var NoTopLevelAwaitRule = rule.Rule{
	Name:   "node/no-top-level-await",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		ignoreBin, _ := opts["ignoreBin"].(bool)
		if ignoreBin && strings.HasPrefix(ctx.SourceFile.Text(), "#!/usr/bin/env") {
			return nil
		}
		program := ctx.Program()
		fileName := ctx.SourceFile.FileName()
		if program == nil || fileName == "<input>" {
			return nil
		}
		pkg := nodeutil.FindPackage(program, fileName)
		if pkg == nil {
			return nil
		}
		relative := tspath.GetRelativePathFromDirectory(pkg.Directory(), fileName, tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true})
		converted, ok := nodeutil.ConvertPath(relative, opts, ctx.Settings)
		if !ok {
			return nil
		}
		absolute := tspath.ResolvePath(pkg.Directory(), converted)
		if ignoreBin && pkg.IsBinFile(program, absolute) || nodeutil.IsUnpublished(program, pkg, absolute) {
			return nil
		}

		report := func(node *ast.Node) {
			for child, parent := node, node.Parent; parent != nil; child, parent = parent, parent.Parent {
				// ESTree keeps computed method names and member decorators
				// outside the method's FunctionExpression.
				if ast.IsFunctionLikeDeclaration(parent) && child != parent.Name() && child.Kind != ast.KindDecorator {
					return
				}
			}
			start := scanner.GetTokenPosOfNode(node, ctx.SourceFile, false)
			ctx.ReportRange(node.Loc.WithPos(start), rule.RuleMessage{
				Id:          "forbidden",
				Description: "Top-level `await` is forbidden in published modules.",
			})
		}
		return rule.RuleListeners{
			ast.KindAwaitExpression: report,
			ast.KindForOfStatement: func(node *ast.Node) {
				if node.AsForInOrOfStatement().AwaitModifier != nil {
					report(node)
				}
			},
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if ast.IsVarAwaitUsing(node) {
					// ESTree's statement declaration includes its semicolon;
					// a declaration in a loop header ends at the list itself.
					if node.Parent.Kind == ast.KindVariableStatement {
						node = node.Parent
					}
					report(node)
				}
			},
		}
	},
}
