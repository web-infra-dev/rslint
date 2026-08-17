package no_import_node

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

const (
	nodeTestModule     = "node:test"
	replacementModule  = "@rstest/core"
	nodeTestSubpathPfx = nodeTestModule + "/"
)

// isNodeTestModule reports whether a module specifier resolves to Node's test
// runner: `node:test` itself, or one of its subpaths such as
// `node:test/reporters`. Only the exact `node:test` specifier has a candidate
// replacement in `@rstest/core`, but every form is reported.
func isNodeTestModule(specifier string) bool {
	return specifier == nodeTestModule || strings.HasPrefix(specifier, nodeTestSubpathPfx)
}

var safeImportedNames = map[string]bool{
	"describe": true,
	"it":       true,
	"test":     true,
}

func noImportNodeTestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noImportNodeTest",
		Description: "Do not import the Node test runner in Rstest test files",
	}
}

func isStringNode(node *ast.Node) bool {
	return node.Kind == ast.KindStringLiteral ||
		node.Kind == ast.KindNoSubstitutionTemplateLiteral
}

func canSafelyReplaceModule(declaration *ast.ImportDeclaration) bool {
	if declaration == nil || declaration.ImportClause == nil {
		return false
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause == nil || clause.Name() != nil || clause.NamedBindings == nil ||
		clause.NamedBindings.Kind != ast.KindNamedImports {
		return false
	}
	named := clause.NamedBindings.AsNamedImports()
	if named == nil || named.Elements == nil || len(named.Elements.Nodes) == 0 {
		return false
	}
	for _, element := range named.Elements.Nodes {
		specifier := element.AsImportSpecifier()
		if specifier == nil || specifier.Name() == nil {
			return false
		}
		importedName := specifier.Name().Text()
		if specifier.PropertyName != nil {
			importedName = specifier.PropertyName.Text()
		}
		if !safeImportedNames[importedName] {
			return false
		}
	}
	return true
}

func replacementForModuleSpecifier(sourceFile *ast.SourceFile, node *ast.Node) (string, bool) {
	if sourceFile == nil || node == nil {
		return "", false
	}
	trimmed := internalUtils.TrimNodeTextRange(sourceFile, node)
	text := sourceFile.Text()
	if trimmed.Pos() < 0 || trimmed.End() > len(text) || trimmed.End()-trimmed.Pos() < 2 {
		return "", false
	}
	raw := text[trimmed.Pos():trimmed.End()]
	quote := raw[0]
	if (quote != '\'' && quote != '"') || raw[len(raw)-1] != quote {
		return "", false
	}
	return string(quote) + replacementModule + string(quote), true
}

var NoImportNodeTestRule = rule.Rule{
	Name:   "rstest/no-import-node-test",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration == nil || declaration.ModuleSpecifier == nil ||
					!isNodeTestModule(declaration.ModuleSpecifier.Text()) {
					return
				}
				message := noImportNodeTestMessage()
				if declaration.ModuleSpecifier.Text() != nodeTestModule || !canSafelyReplaceModule(declaration) {
					ctx.ReportNode(node, message)
					return
				}
				replacement, ok := replacementForModuleSpecifier(ctx.SourceFile, declaration.ModuleSpecifier)
				if !ok {
					ctx.ReportNode(node, message)
					return
				}
				ctx.ReportNodeWithFixes(
					node,
					message,
					rule.RuleFixReplaceRange(
						internalUtils.TrimNodeTextRange(ctx.SourceFile, declaration.ModuleSpecifier),
						replacement,
					),
				)
			},
			// `require('node:test')` is reported as well, but never fixed: the
			// binding forms a require can take do not map onto a specifier
			// rewrite the way a static import does.
			ast.KindCallExpression: func(node *ast.Node) {
				callee := node.AsCallExpression().Expression
				if callee.Kind != ast.KindIdentifier || callee.AsIdentifier().Text != "require" {
					return
				}
				arguments := node.Arguments()
				if len(arguments) == 0 {
					return
				}
				firstArg := arguments[0]
				if firstArg == nil || !isStringNode(firstArg) || !isNodeTestModule(firstArg.Text()) {
					return
				}
				ctx.ReportNode(firstArg, noImportNodeTestMessage())
			},
		}
	},
}
