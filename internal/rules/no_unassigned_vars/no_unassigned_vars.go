package no_unassigned_vars

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
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
	ctx            rule.RuleContext
	pending        *ast.Node
	pendingSymbols variableSymbols
	writesBySymbol map[*ast.Symbol]bool
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
	symbols := binderVariableDeclarationSymbols(node)
	if symbols.count == 0 || s.symbolsHaveKnownWrite(symbols) {
		return
	}
	s.checkReferences(node, symbols)
}

func (s *runState) checkReferences(node *ast.Node, symbols variableSymbols) {
	hasRead := false
	for _, sym := range symbols.values[:symbols.count] {
		for _, refNode := range s.ctx.Refs.References(sym) {
			if utils.IsVariableWriteReference(refNode) {
				return
			}
			if utils.IsReadReference(refNode) {
				hasRead = true
			}
		}
	}
	if hasRead {
		s.ctx.ReportNode(node, messageUnassigned(node.Name().Text()))
	}
}

func (s *runState) visitVariableDeclarator(node *ast.Node) {
	if s.shouldSkipDeclarator(node) {
		return
	}
	symbols := binderVariableDeclarationSymbols(node)
	if symbols.count == 0 || s.symbolsHaveKnownWrite(symbols) {
		return
	}
	// Keep only one candidate pending. A later candidate falls back to the
	// shared reference index instead of building another per-file index here.
	s.finish(nil)
	s.pending = node
	s.pendingSymbols = symbols
}

func (s *runState) visitAssignment(node *ast.Node) {
	if s.pending == nil {
		return
	}
	binary := node.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindEqualsToken || binary.Left.Kind != ast.KindIdentifier {
		return
	}
	if binary.Left.Text() != s.pending.Name().Text() {
		return
	}
	// Resolve only obvious writes encountered after a candidate declaration.
	// Earlier writes and other assignment forms still fall back to References.
	// The shared resolver keeps same-named bindings in different scopes apart.
	sym := s.ctx.Refs.ResolveInFile(binary.Left)
	for _, candidate := range s.pendingSymbols.values[:s.pendingSymbols.count] {
		if sym == candidate {
			s.pending = nil
			// Repeated declarations already have a cache entry. Remember the
			// write for them without allocating a map for a lone declaration.
			if s.writesBySymbol != nil {
				s.writesBySymbol[sym] = true
			}
			return
		}
	}
}

func (s *runState) finish(*ast.Node) {
	if s.pending != nil {
		s.checkReferences(s.pending, s.pendingSymbols)
		s.pending = nil
	}
}

// binderVariableDeclarationSymbols returns every binder representation that
// can key ctx.Refs for this variable. A directly exported declaration exposes
// an export symbol on the declaration and a linked LocalSymbol. A lone exported
// declaration resolves uses to the export symbol, while merged/repeated
// declarations resolve them to the local symbol, so the rule must query both
// identities. The binder provides these links without another scope walk.
func binderVariableDeclarationSymbols(node *ast.Node) variableSymbols {
	var result variableSymbols
	if node == nil || node.Kind != ast.KindVariableDeclaration {
		return result
	}
	nameNode := node.Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return result
	}
	for _, sym := range [...]*ast.Symbol{node.Symbol(), node.LocalSymbol()} {
		if sym != nil {
			result.add(sym)
			result.add(sym.ExportSymbol)
		}
	}
	return result
}

// symbolsHaveKnownWrite covers cached writes and declaration-position writes
// that RefStore intentionally omits. Inspect every declaration because repeated
// `var` declarations share one symbol: an initializer or for-in/of binding on
// any sibling declaration assigns the variable for all of them.
func (s *runState) symbolsHaveKnownWrite(symbols variableSymbols) bool {
	for _, sym := range symbols.values[:symbols.count] {
		if s.symbolHasKnownWrite(sym) {
			return true
		}
	}
	return false
}

func (s *runState) symbolHasKnownWrite(sym *ast.Symbol) bool {
	if sym == nil {
		return false
	}
	if s.writesBySymbol != nil {
		if hasWrite, ok := s.writesBySymbol[sym]; ok {
			return hasWrite
		}
	}
	declarations := sym.Declarations

	hasWrite := false
	for _, declaration := range declarations {
		if declaration != nil && declaration.Kind == ast.KindVariableDeclaration &&
			utils.VariableDeclarationIntroducesWrite(declaration) {
			hasWrite = true
			break
		}
	}
	if len(declarations) > 1 {
		if s.writesBySymbol == nil {
			s.writesBySymbol = make(map[*ast.Symbol]bool)
		}
		s.writesBySymbol[sym] = hasWrite
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

	// Loop bindings are assigned by iteration. Reject them before querying
	// references so a file containing only assigned variables never builds
	// the shared reference index on this rule's behalf.
	return utils.IsVarDeclInForInOrOf(node)
}

var NoUnassignedVarsRule = rule.Rule{
	Name:   "no-unassigned-vars",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		s := &runState{ctx: ctx}
		// A simple assignment always contains a literal '=' token. Files
		// without one can check declarations directly without buffering them.
		if !strings.Contains(ctx.SourceFile.Text(), "=") {
			return rule.RuleListeners{ast.KindVariableDeclaration: s.checkVariableDeclarator}
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: s.visitVariableDeclarator,
			ast.KindBinaryExpression:    s.visitAssignment,
			ast.KindEndOfFile:           s.finish,
		}
	},
}
