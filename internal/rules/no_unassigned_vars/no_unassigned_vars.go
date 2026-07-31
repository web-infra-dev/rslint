package no_unassigned_vars

import (
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
			if utils.IsReadReference(refNode) {
				hasRead = true
			}
		}
		if hasWrite {
			break
		}
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

var NoUnassignedVarsRule = rule.Rule{
	Name: "no-unassigned-vars",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		s := &runState{ctx: ctx}

		return rule.RuleListeners{ast.KindVariableDeclaration: s.checkVariableDeclarator}
	},
}
