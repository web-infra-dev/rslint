package no_export

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintutils "github.com/web-infra-dev/rslint/internal/utils"
)

func buildUnexpectedExportMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedExport",
		Description: "Do not export from a test file",
	}
}

func isModuleExportsMemberExpression(node *ast.Node, ctx rule.RuleContext) bool {
	if node == nil || !utils.IsMemberAccessNode(node) {
		return false
	}

	current := node
	innermostProperty := ""
	for utils.IsMemberAccessNode(current) {
		property, ok := rslintutils.AccessExpressionStaticName(current)
		if !ok {
			return false
		}
		innermostProperty = property
		current = ast.SkipParentheses(rslintutils.AccessExpressionObject(current))
	}

	if current == nil || current.Kind != ast.KindIdentifier ||
		current.AsIdentifier().Text != "module" || innermostProperty != "exports" {
		return false
	}

	// A locally declared variable or parameter named `module` shadows the
	// CommonJS global and must not be treated as an export target.
	if ctx.TypeChecker != nil {
		if symbol := ctx.TypeChecker.GetSymbolAtLocation(current); symbol != nil {
			for _, declaration := range symbol.Declarations {
				if ast.GetSourceFileOfNode(declaration) == ctx.SourceFile {
					return false
				}
			}
		}
	}

	return true
}

func hasExportModifier(node *ast.Node) bool {
	return ast.HasSyntacticModifier(node, ast.ModifierFlagsExport)
}

var NoExportRule = rule.Rule{
	Name: "jest/no-export",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		_ = options
		exportNodes := make([]*ast.Node, 0)
		hasJestBlock := false

		collectExport := func(node *ast.Node) {
			exportNodes = append(exportNodes, node)
		}

		collectExportStatement := func(node *ast.Node) {
			if hasExportModifier(node) {
				collectExport(node)
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if utils.IsTypeOfJestFnCall(node, ctx, utils.JestFnTypeDescribe, utils.JestFnTypeTest) {
					hasJestBlock = true
				}
			},
			ast.KindVariableStatement:       collectExportStatement,
			ast.KindFunctionDeclaration:     collectExportStatement,
			ast.KindClassDeclaration:        collectExportStatement,
			ast.KindInterfaceDeclaration:    collectExportStatement,
			ast.KindTypeAliasDeclaration:    collectExportStatement,
			ast.KindEnumDeclaration:         collectExportStatement,
			ast.KindModuleDeclaration:       collectExportStatement,
			ast.KindImportEqualsDeclaration: collectExportStatement,
			ast.KindExportDeclaration:       collectExport,
			ast.KindExportAssignment:        collectExport,
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				if bin == nil || bin.OperatorToken == nil || !ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
					return
				}

				for _, operand := range []*ast.Node{bin.Left, bin.Right} {
					operand = ast.SkipParentheses(operand)
					if operand != nil && isModuleExportsMemberExpression(operand, ctx) {
						collectExport(operand)
					}
				}
			},
			rule.ListenerOnExit(ast.KindEndOfFile): func(node *ast.Node) {
				_ = node
				if !hasJestBlock || len(exportNodes) == 0 {
					return
				}
				for _, exportNode := range exportNodes {
					ctx.ReportNode(exportNode, buildUnexpectedExportMessage())
				}
			},
		}
	},
}
