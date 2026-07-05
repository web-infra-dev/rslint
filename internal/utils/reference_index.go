package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
)

// ReferenceIndex wraps TypeScript-Go's per-symbol reference lookup and keeps a
// small name index only for checker edge cases where symbol identity differs
// from ESLint scope-manager's local variable identity.
type ReferenceIndex struct {
	sourceFile  *ast.SourceFile
	typeChecker *checker.Checker
	refsByName  map[string][]*ast.Node
}

func NewReferenceIndex(sourceFile *ast.SourceFile, typeChecker *checker.Checker) *ReferenceIndex {
	return &ReferenceIndex{
		sourceFile:  sourceFile,
		typeChecker: typeChecker,
	}
}

// GetVariableDeclarationSymbol returns the declared variable symbol for a
// VariableDeclaration. The declaration node's symbol is the closest match to
// ESLint's getDeclaredVariables(); GetSymbolAtLocation is only a fallback for
// older checker states where the declaration symbol was not attached.
func GetVariableDeclarationSymbol(varDecl *ast.Node, typeChecker *checker.Checker) *ast.Symbol {
	if varDecl == nil {
		return nil
	}
	if sym := varDecl.Symbol(); sym != nil {
		return sym
	}
	vd := varDecl.AsVariableDeclaration()
	if vd == nil || vd.Name() == nil || typeChecker == nil {
		return nil
	}
	return typeChecker.GetSymbolAtLocation(vd.Name())
}

// ForEachReference invokes cb for every identifier node resolving to sym, in
// source order. It returns early when cb returns true.
func (i *ReferenceIndex) ForEachReference(sym *ast.Symbol, cb func(*ast.Node) bool) {
	if i == nil || i.sourceFile == nil || i.typeChecker == nil || sym == nil {
		return
	}
	for _, node := range i.typeChecker.GetReferencesToSymbolInFile(i.sourceFile, sym) {
		if cb(node) {
			return
		}
	}
}

// ForEachReferenceByName invokes cb for identifier references with the given
// text that are not shadowed between the reference site and boundary. This is
// a fallback for TypeScript checker cases where a local declaration collides
// with a lib/global symbol and symbol identity no longer matches ESLint's
// scope-manager variable identity.
func (i *ReferenceIndex) ForEachReferenceByName(name string, boundary *ast.Node, cb func(*ast.Node) bool) {
	if i == nil || name == "" {
		return
	}
	i.buildNameIndex()
	for _, node := range i.refsByName[name] {
		if boundary != nil && IsNameShadowedBetween(node, boundary, name) {
			continue
		}
		if cb(node) {
			return
		}
	}
}

func (i *ReferenceIndex) buildNameIndex() {
	if i.refsByName != nil {
		return
	}
	i.refsByName = map[string][]*ast.Node{}
	if i.sourceFile == nil {
		return
	}

	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		if n.Kind == ast.KindIdentifier && !IsNonReferenceIdentifier(n) {
			i.refsByName[n.Text()] = append(i.refsByName[n.Text()], n)
		} else if n.Kind == ast.KindShorthandPropertyAssignment && IsInDestructuringAssignment(n) {
			// Keep destructuring shorthand writes indexed even if a future tsgo
			// traversal skips the child identifier for shorthand nodes.
			if name := n.AsShorthandPropertyAssignment().Name(); name != nil {
				i.refsByName[name.Text()] = append(i.refsByName[name.Text()], name)
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(i.sourceFile.AsNode())
}

// IsVariableWriteReference reports whether an identifier writes to its
// variable. It extends IsWriteReference with writes introduced by variable
// declarations with initializers, for-in/for-of declarations, and catch
// bindings, matching ESLint scope-manager write-reference semantics.
func IsVariableWriteReference(node *ast.Node) bool {
	if IsWriteReference(node) {
		return true
	}
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		varDecl := parent.AsVariableDeclaration()
		return varDecl != nil && varDecl.Name() == node && VariableDeclarationIntroducesWrite(parent)
	case ast.KindBindingElement:
		be := parent.AsBindingElement()
		if be == nil || be.Name() != node {
			return false
		}
		return ast.IsWriteAccessForReference(node) ||
			VariableDeclarationIntroducesWrite(EnclosingVariableDeclarationOfBindingElement(parent))
	}
	return false
}

// VariableDeclarationIntroducesWrite reports whether a VariableDeclaration's
// binding is written at introduction.
func VariableDeclarationIntroducesWrite(varDecl *ast.Node) bool {
	if varDecl == nil || varDecl.Kind != ast.KindVariableDeclaration {
		return false
	}
	return ast.IsWriteAccessForReference(varDecl.Name()) || IsVarDeclInForInOrOf(varDecl)
}

// IsVarDeclInForInOrOf reports whether a VariableDeclaration sits directly
// inside a for-in/for-of initializer.
func IsVarDeclInForInOrOf(varDecl *ast.Node) bool {
	if varDecl == nil || varDecl.Parent == nil {
		return false
	}
	declList := varDecl.Parent
	if declList.Kind != ast.KindVariableDeclarationList || declList.Parent == nil {
		return false
	}
	outer := declList.Parent
	return outer.Kind == ast.KindForInStatement || outer.Kind == ast.KindForOfStatement
}

// EnclosingVariableDeclarationOfBindingElement walks through nested
// BindingElement / BindingPattern layers to the containing VariableDeclaration.
func EnclosingVariableDeclarationOfBindingElement(bindingElement *ast.Node) *ast.Node {
	if bindingElement == nil || bindingElement.Kind != ast.KindBindingElement {
		return nil
	}
	parent := ast.WalkUpBindingElementsAndPatterns(bindingElement)
	if parent == nil || parent.Kind != ast.KindVariableDeclaration {
		return nil
	}
	return parent
}

// GetDeclListForSymbolDecl returns the VariableDeclarationList associated with
// a declaration node, or nil if the declaration is not a variable-like binding.
func GetDeclListForSymbolDecl(decl *ast.Node) *ast.Node {
	for current := decl; current != nil; {
		if current.Kind == ast.KindVariableDeclarationList {
			return current
		}
		if current.Kind == ast.KindVariableDeclaration ||
			current.Kind == ast.KindBindingElement ||
			current.Kind == ast.KindObjectBindingPattern ||
			current.Kind == ast.KindArrayBindingPattern {
			current = current.Parent
			continue
		}
		return nil
	}
	return nil
}
