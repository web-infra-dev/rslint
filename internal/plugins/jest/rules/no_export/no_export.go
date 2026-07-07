package no_export

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
)

func buildUnexpectedExportMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedExport",
		Description: "Do not export from a test file",
	}
}

func isLocallyDeclaredIdentifier(node *ast.Node, ctx rule.RuleContext) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	if symbol := ctx.TypeChecker.GetSymbolAtLocation(node); symbol != nil {
		for _, declaration := range symbol.Declarations {
			if ast.GetSourceFileOfNode(declaration) == ctx.SourceFile {
				return true
			}
		}
	}
	return false
}

func isCommonJSExportsMemberExpression(node *ast.Node, ctx rule.RuleContext) bool {
	if node == nil || !utils.IsMemberAccessNode(node) {
		return false
	}

	current := node
	innermostProperty := ""
	innermostPropertyKnown := false
	for utils.IsMemberAccessNode(current) {
		innermostProperty, innermostPropertyKnown = rslintUtils.AccessExpressionStaticName(current)
		current = ast.SkipParentheses(rslintUtils.AccessExpressionObject(current))
	}

	if current == nil || current.Kind != ast.KindIdentifier {
		return false
	}

	switch current.AsIdentifier().Text {
	case "exports":
		return !isLocallyDeclaredIdentifier(current, ctx)
	case "module":
		return innermostPropertyKnown && innermostProperty == "exports" &&
			!isLocallyDeclaredIdentifier(current, ctx)
	default:
		return false
	}
}

func hasExportModifier(node *ast.Node) bool {
	return ast.HasSyntacticModifier(node, ast.ModifierFlagsExport)
}

var NoExportRule = rule.Rule{
	Name: "jest/no-export",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.UnwrapOptions(_options)
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
					if operand != nil && isCommonJSExportsMemberExpression(operand, ctx) {
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
