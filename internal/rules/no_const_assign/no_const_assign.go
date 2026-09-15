package no_const_assign

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Message builder
func buildConstMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "const",
		Description: "'" + name + "' is constant.",
	}
}

func isConstantBindingSymbol(symbol *ast.Symbol, sourceFile *ast.SourceFile) bool {
	if symbol == nil || sourceFile == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if ast.GetSourceFileOfNode(declaration) != sourceFile {
			continue
		}
		declList := utils.GetDeclListForSymbolDecl(declaration)
		if declList != nil && declList.Flags&ast.NodeFlagsConstant != 0 {
			return true
		}
	}
	return false
}

// NoConstAssignRule disallows reassigning constant variables.
var NoConstAssignRule = rule.Rule{
	Name:   "no-const-assign",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// Every constant variable declaration contains one of these literal
		// keywords. Avoid an identifier listener when the rule cannot report;
		// false positives in comments or strings only retain the normal path.
		if ctx.SourceFile != nil {
			text := ctx.SourceFile.Text()
			if !strings.Contains(text, "const") && !strings.Contains(text, "using") {
				return nil
			}
		}

		return rule.RuleListeners{
			// Check every write reference in the file — including shorthand
			// destructuring writes like `({x} = {x: 1})`, since ctx.Refs.Resolve
			// resolves the shorthand name's own binding symbol rather than the
			// property symbol a TypeChecker lookup would otherwise return.
			ast.KindIdentifier: func(node *ast.Node) {
				if !utils.IsWriteReference(node) {
					return
				}
				symbol := ctx.Refs.Resolve(node)
				if !isConstantBindingSymbol(symbol, ctx.SourceFile) {
					return
				}
				ctx.ReportNode(node, buildConstMessage(node.Text()))
			},
		}
	},
}
