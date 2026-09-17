package no_document_cookie

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/referencetracker"
)

const messageID = "no-document-cookie"

var globalObjectNames = [...]string{"globalThis", "window", "self", "global"}

// NoDocumentCookieRule disallows assigning to document.cookie directly.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-document-cookie.js
var NoDocumentCookieRule = rule.Rule{
	Name:   "unicorn/no-document-cookie",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if !sourceMayReferenceDocument(ctx.SourceFile) {
			return nil
		}

		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
				tracker := referencetracker.New(ctx)
				document := &referencetracker.Trace{Properties: map[string]*referencetracker.Trace{
					"cookie": {Read: func(node *ast.Node) {
						if ast.IsAccessExpression(node) && isDirectAssignmentTarget(node) {
							ctx.ReportNode(node, rule.RuleMessage{
								Id:          messageID,
								Description: "Do not use `document.cookie` directly.",
							})
						}
					}},
				}}
				tracker.TrackGlobal("document", document)
				globalObject := &referencetracker.Trace{Properties: map[string]*referencetracker.Trace{"document": document}}
				for _, name := range globalObjectNames {
					tracker.TrackGlobal(name, globalObject)
				}
			},
		}
	},
}

func sourceMayReferenceDocument(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.AsNode().Kind != ast.KindSourceFile {
		return true
	}
	if sourceFile.HasIdentifier("document") {
		return true
	}
	for _, name := range globalObjectNames {
		if sourceFile.HasIdentifier(name) {
			return true
		}
	}
	return false
}

func isDirectAssignmentTarget(node *ast.Node) bool {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	parent := current.Parent
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := parent.AsBinaryExpression()
	return binary != nil && binary.Left == current && binary.OperatorToken != nil &&
		ast.IsAssignmentOperator(binary.OperatorToken.Kind)
}
