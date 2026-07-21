package no_unassigned_vars

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-unassigned-vars

func messageUnassigned(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unassigned",
		Description: fmt.Sprintf("'%s' is always 'undefined' because it's never assigned.", name),
		Data: map[string]string{
			"name": name,
		},
	}
}

type runState struct {
	ctx  rule.RuleContext
	refs *utils.ReferenceIndex
}

func (s *runState) checkVariableDeclarator(node *ast.Node) {
	if s.shouldSkipDeclarator(node) {
		return
	}

	nameNode := node.AsVariableDeclaration().Name()
	sym := utils.GetVariableDeclarationSymbol(node, s.ctx.TypeChecker)
	if sym == nil {
		return
	}

	name := nameNode.AsIdentifier().Text
	hasRead := false
	hasWrite := false
	s.refs.ForEachReference(sym, func(refNode *ast.Node) bool {
		if utils.IsVariableWriteReference(refNode) {
			hasWrite = true
			return true
		}
		if isReadReference(refNode) {
			hasRead = true
		}
		return false
	})
	if !hasWrite {
		s.refs.ForEachReferenceByName(name, node, func(refNode *ast.Node) bool {
			if utils.IsVariableWriteReference(refNode) {
				hasWrite = true
				return true
			}
			if isReadReference(refNode) {
				hasRead = true
			}
			return false
		})
	}
	if hasWrite || !hasRead {
		return
	}

	s.ctx.ReportNode(node, messageUnassigned(name))
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
	if utils.IsIdentifierInTypeReference(node) {
		return false
	}
	return !utils.IsNonReferenceIdentifier(node)
}

var NoUnassignedVarsRule = rule.Rule{
	Name:             "no-unassigned-vars",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if ctx.TypeChecker == nil {
			return rule.RuleListeners{}
		}

		s := &runState{
			ctx:  ctx,
			refs: utils.NewReferenceIndex(ctx.SourceFile, ctx.TypeChecker),
		}

		return rule.RuleListeners{ast.KindVariableDeclaration: s.checkVariableDeclarator}
	},
}
