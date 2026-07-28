package no_unassigned_vars

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-unassigned-vars

func messageUnassigned(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unassigned",
		Description: "'" + name + "' is always 'undefined' because it's never assigned.",
		Data: map[string]string{
			"name": name,
		},
	}
}

type runState struct {
	ctx                      rule.RuleContext
	declarationWriteBySymbol map[*ast.Symbol]bool
	localExportsScanned      bool
	localExportReadBySym     map[*ast.Symbol]bool
}

type variableSymbols struct {
	values [3]*ast.Symbol
	count  int
}

func (s *variableSymbols) add(sym *ast.Symbol) {
	if sym == nil {
		return
	}
	for _, existing := range s.values[:s.count] {
		if existing == sym {
			return
		}
	}
	if s.count == len(s.values) {
		return
	}
	s.values[s.count] = sym
	s.count++
}

func (s *runState) checkVariableDeclarator(node *ast.Node) {
	if s.shouldSkipDeclarator(node) {
		return
	}

	nameNode := node.AsVariableDeclaration().Name()
	symbols := binderVariableDeclarationSymbols(node)
	if symbols.count == 0 {
		return
	}

	name := nameNode.AsIdentifier().Text
	hasRead := false
	hasWrite := s.symbolsHaveDeclarationWrite(symbols)
	for _, sym := range symbols.values[:symbols.count] {
		for _, refNode := range s.ctx.Refs.References(sym) {
			if utils.IsVariableWriteReference(refNode) {
				hasWrite = true
				break
			}
			if isReadReference(refNode) {
				hasRead = true
			}
		}
		if hasWrite {
			break
		}
	}
	if !hasRead && !hasWrite {
		hasRead = s.isLocalExportRead(symbols)
	}
	if hasWrite || !hasRead {
		return
	}

	s.ctx.ReportNode(node, messageUnassigned(name))
}

// binderVariableDeclarationSymbols returns every binder representation that
// can key ctx.Refs for this variable. A directly exported declaration exposes
// an export symbol on the declaration and a linked local symbol in the nearest
// Locals table. A lone exported declaration resolves uses to the export symbol,
// while merged/repeated declarations resolve them to the local symbol, so the
// rule must query both identities.
func binderVariableDeclarationSymbols(node *ast.Node) variableSymbols {
	var result variableSymbols
	if node == nil || node.Kind != ast.KindVariableDeclaration {
		return result
	}
	nameNode := node.Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return result
	}
	result.add(node.Symbol())
	name := nameNode.Text()
	for current := node.Parent; current != nil; current = current.Parent {
		if !ast.IsLocalsContainer(current) {
			continue
		}
		sym := ast.GetLocals(current)[name]
		if symbolOwnsVariableDeclaration(sym, node) {
			result.add(sym)
			result.add(sym.ExportSymbol)
			break
		}
	}
	return result
}

func symbolOwnsVariableDeclaration(sym *ast.Symbol, node *ast.Node) bool {
	if sym == nil || node == nil {
		return false
	}
	nodeSymbol := node.Symbol()
	if nodeSymbol != nil &&
		(sym == nodeSymbol || sym.ExportSymbol == nodeSymbol || nodeSymbol.ExportSymbol == sym) {
		return true
	}
	root := ast.GetRootDeclaration(node)
	for _, declaration := range sym.Declarations {
		if declaration == node || declaration == root ||
			ast.GetRootDeclaration(declaration) == root {
			return true
		}
	}
	return false
}

// symbolsHaveDeclarationWrite covers declaration-position writes that RefStore
// intentionally omits. Inspect every declaration because repeated `var`
// declarations share one symbol: an initializer or for-in/of binding on any
// sibling declaration assigns the variable for all of them.
func (s *runState) symbolsHaveDeclarationWrite(symbols variableSymbols) bool {
	for _, sym := range symbols.values[:symbols.count] {
		if s.symbolHasDeclarationWrite(sym) {
			return true
		}
	}
	return false
}

func (s *runState) symbolHasDeclarationWrite(sym *ast.Symbol) bool {
	if sym == nil {
		return false
	}
	declarations := sym.Declarations
	if len(declarations) > 1 && s.declarationWriteBySymbol != nil {
		if hasWrite, ok := s.declarationWriteBySymbol[sym]; ok {
			return hasWrite
		}
	}

	hasWrite := false
	for _, declaration := range declarations {
		if declaration != nil && declaration.Kind == ast.KindVariableDeclaration &&
			utils.VariableDeclarationIntroducesWrite(declaration) {
			hasWrite = true
			break
		}
	}
	if len(declarations) > 1 {
		if s.declarationWriteBySymbol == nil {
			s.declarationWriteBySymbol = make(map[*ast.Symbol]bool)
		}
		s.declarationWriteBySymbol[sym] = hasWrite
	}
	return hasWrite
}

// isLocalExportRead supplements ctx.Refs for local named export specifiers.
// RefStore deliberately excludes import/export bindings because its other
// consumers do not treat them as references, while ESLint scope-manager marks
// `export { value }` as a read of value for this rule. The exact binder symbol
// map preserves shadowing and ignores type-only/source-bearing re-exports.
func (s *runState) isLocalExportRead(symbols variableSymbols) bool {
	if !s.localExportsScanned {
		s.localExportsScanned = true
		if strings.Contains(s.ctx.SourceFile.Text(), "export") {
			s.localExportReadBySym = collectLocalExportReads(s.ctx.SourceFile)
		}
	}
	for _, sym := range symbols.values[:symbols.count] {
		if s.localExportReadBySym[sym] {
			return true
		}
	}
	return false
}

func collectLocalExportReads(sourceFile *ast.SourceFile) map[*ast.Symbol]bool {
	if sourceFile == nil {
		return nil
	}
	var result map[*ast.Symbol]bool
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindExportSpecifier &&
			!ast.IsTypeOnlyImportOrExportDeclaration(node) &&
			!utils.IsReExportSpecifier(node) {
			specifier := node.AsExportSpecifier()
			if specifier != nil {
				localName := specifier.Name()
				if specifier.PropertyName != nil {
					localName = specifier.PropertyName
				}
				if target := binderLocalExportTarget(localName); target != nil {
					if result == nil {
						result = make(map[*ast.Symbol]bool)
					}
					result[target] = true
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(sourceFile.AsNode())
	return result
}

func binderLocalExportTarget(localName *ast.Node) *ast.Symbol {
	if localName == nil || localName.Kind != ast.KindIdentifier {
		return nil
	}
	name := localName.Text()
	for current := localName.Parent; current != nil; current = current.Parent {
		if ast.IsLocalsContainer(current) {
			if sym := ast.GetLocals(current)[name]; sym != nil {
				return sym
			}
		}
	}
	return nil
}

func (s *runState) shouldSkipDeclarator(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindVariableDeclaration {
		return true
	}
	varDecl := node.AsVariableDeclaration()
	if varDecl == nil || varDecl.Initializer != nil {
		return true
	}

	nameNode := varDecl.Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return true
	}

	declList := node.Parent
	kind := utils.GetVarDeclListKind(declList)
	if kind != "var" && kind != "let" {
		return true
	}

	if utils.IsInAmbientContext(node) {
		return true
	}

	return false
}

func isReadReference(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if ast.IsPartOfTypeNode(node) || ast.IsPartOfTypeQuery(node) {
		return false
	}
	return !utils.IsNonReferenceIdentifier(node)
}

var NoUnassignedVarsRule = rule.Rule{
	Name:             "no-unassigned-vars",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if ctx.TypeChecker == nil || ctx.Refs == nil {
			return rule.RuleListeners{}
		}

		s := &runState{ctx: ctx}

		return rule.RuleListeners{ast.KindVariableDeclaration: s.checkVariableDeclarator}
	},
}
