package prefer_namespace_keyword

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var useNamespaceMessage = rule.RuleMessage{
	Id:          "useNamespace",
	Description: "Use 'namespace' instead of 'module' to declare custom TypeScript modules.",
}

const (
	moduleKeywordText    = "module"
	namespaceKeywordText = "namespace"
)

func moduleDeclarationRanges(sourceFile *ast.SourceFile, node *ast.Node) (core.TextRange, core.TextRange, bool) {
	text := sourceFile.Text()
	declarationStart := scanner.SkipTrivia(text, node.Pos())
	keywordStart := declarationStart
	if modifiers := node.Modifiers(); modifiers != nil && len(modifiers.Nodes) != 0 {
		keywordStart = scanner.SkipTrivia(text, modifiers.End())
	}
	if keywordStart < 0 || keywordStart > len(text)-len(moduleKeywordText) {
		return core.TextRange{}, core.TextRange{}, false
	}
	keywordEnd := keywordStart + len(moduleKeywordText)
	// Nested nodes in `module foo.bar {}` inherit ModuleKeyword from the
	// outer declaration but do not have their own keyword in the source.
	if text[keywordStart:keywordEnd] != moduleKeywordText {
		return core.TextRange{}, core.TextRange{}, false
	}
	return node.Loc.WithPos(declarationStart), core.NewTextRange(keywordStart, keywordEnd), true
}

var PreferNamespaceKeywordRule = rule.CreateRule(rule.Rule{
	Name: "prefer-namespace-keyword",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindModuleDeclaration: func(node *ast.Node) {
				moduleDecl := node.AsModuleDeclaration()
				if moduleDecl == nil {
					return
				}

				// Skip if the module name is a string literal (external module declaration)
				name := moduleDecl.Name()
				if name == nil {
					return
				}
				if name.Kind == ast.KindStringLiteral {
					return
				}

				// Check if the keyword before the module name is "module"
				if moduleDecl.Keyword != ast.KindModuleKeyword {
					return
				}

				declarationRange, keywordRange, ok := moduleDeclarationRanges(ctx.SourceFile, node)
				if !ok {
					return
				}

				ctx.ReportRangeWithDeferredFixes(
					declarationRange,
					useNamespaceMessage,
					func() []rule.RuleFix {
						return []rule.RuleFix{rule.RuleFixReplaceRange(keywordRange, namespaceKeywordText)}
					},
				)
			},
		}
	},
})
