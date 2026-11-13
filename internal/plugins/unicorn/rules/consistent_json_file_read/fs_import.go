package consistent_json_file_read

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// fsImport returns the module and the imported name. A namespace import has
// no imported name, and a default import has the name "default".
func fsImport(symbol *ast.Symbol) (module, imported string) {
	if symbol == nil || len(symbol.Declarations) != 1 {
		return "", ""
	}
	declaration := symbol.Declarations[0]
	switch declaration.Kind {
	case ast.KindImportClause:
		imported = "default"
	case ast.KindNamespaceImport:
	case ast.KindImportSpecifier:
		name := declaration.PropertyName()
		if name == nil {
			name = declaration.Name()
		}
		if !ast.IsIdentifier(name) {
			return "", ""
		}
		imported = name.Text()
	default:
		return "", ""
	}
	if ast.IsTypeOnlyImportDeclaration(declaration) {
		return "", ""
	}
	importNode := ast.FindAncestorKind(declaration, ast.KindImportDeclaration)
	if importNode == nil || !ast.IsStringLiteral(importNode.ModuleSpecifier()) {
		return "", ""
	}
	switch importNode.ModuleSpecifier().Text() {
	case "fs", "node:fs":
		return "fs", imported
	case "fs/promises", "node:fs/promises":
		return "fs/promises", imported
	}
	return "", ""
}

// Only calls tied to a Node.js import can prove that this use of an options
// object leaves it alone. A method name alone is enough to recognize a read for
// the rule, but a user-defined method of the same name can mutate the options.
func (refs *optionReferences) isNodeFileRead(call, argument *ast.Node) bool {
	args := call.Arguments()
	if len(args) < 2 || len(args) > 3 || args[1] != argument ||
		ast.NodeIsMissing(args[0]) || ast.IsSpreadElement(args[0]) ||
		(len(args) == 3 && ast.IsSpreadElement(args[2])) {
		return false
	}
	callee := utils.SkipAssertionsAndParens(call.Expression())
	if callee == nil {
		return false
	}
	if ast.IsIdentifier(callee) {
		module, method := fsImport(refs.ctx.Refs.ResolveInFile(callee))
		return isNodeReadMethod(module, method, len(args))
	}
	if !ast.IsAccessExpression(callee) {
		return false
	}
	method, _ := utils.AccessExpressionStaticName(callee)
	if method != "readFile" && method != "readFileSync" {
		return false
	}
	receiver := utils.SkipAssertionsAndParens(utils.AccessExpressionObject(callee))
	if receiver == nil || !ast.IsIdentifier(receiver) {
		return false
	}
	symbol := refs.ctx.Refs.ResolveInFile(receiver)
	module, imported := fsImport(symbol)
	if module == "fs" && imported == "promises" {
		module = "fs/promises"
	} else if imported != "" && imported != "default" {
		return false
	}
	return isNodeReadMethod(module, method, len(args)) && refs.hasOnlyReadCalls(symbol, module)
}

func isNodeReadMethod(module, method string, arguments int) bool {
	switch module {
	case "fs":
		return (method == "readFileSync" && arguments <= 2) || (method == "readFile" && arguments <= 3)
	case "fs/promises":
		return method == "readFile" && arguments <= 2
	}
	return false
}

// Imported objects are mutable. Do not trust their methods when the receiver
// itself escapes or a use can replace a method.
func (refs *optionReferences) hasOnlyReadCalls(symbol *ast.Symbol, module string) bool {
	for _, reference := range refs.ctx.Refs.References(symbol) {
		if !utils.IsReadReference(reference) {
			continue
		}
		receiver := utils.OutermostParenthesizedExpression(reference)
		access := receiver.Parent
		if access == nil || !ast.IsAccessExpression(access) || utils.AccessExpressionObject(access) != receiver {
			return false
		}
		method, _ := utils.AccessExpressionStaticName(access)
		callee := utils.OutermostParenthesizedExpression(access)
		call := callee.Parent
		if call == nil || !ast.IsCallExpression(call) || call.Expression() != callee || !isNodeReadMethod(module, method, len(call.Arguments())) {
			return false
		}
	}
	return true
}
