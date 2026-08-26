package consistent_type_definitions

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed consistent_type_definitions.schema.json
var schemaJSON []byte

type DefinitionStyle string

const (
	DefinitionStyleInterface DefinitionStyle = "interface"
	DefinitionStyleType      DefinitionStyle = "type"
)

type ConsistentTypeDefinitionsOptions struct {
	Style DefinitionStyle `json:"style"`
}

func parseOptions(options []any) ConsistentTypeDefinitionsOptions {
	opts := ConsistentTypeDefinitionsOptions{
		Style: DefinitionStyleInterface,
	}
	if len(options) == 0 {
		return opts
	}
	if str, ok := options[0].(string); ok {
		opts.Style = DefinitionStyle(str)
	}
	return opts
}

type consistentTypeDefinitionsFixer struct {
	sourceFile *ast.SourceFile
}

func (f consistentTypeDefinitionsFixer) tokenBounds(from int, to int, kind ast.Kind) (int, int, bool) {
	sourceScanner := scanner.GetScannerForSourceFile(f.sourceFile, from)
	for sourceScanner.TokenStart() < to {
		if sourceScanner.Token() == kind {
			return sourceScanner.TokenStart(), sourceScanner.TokenEnd(), true
		}
		if sourceScanner.Token() == ast.KindEndOfFile {
			break
		}
		sourceScanner.Scan()
	}
	return 0, 0, false
}

// beforeEqualsEnd mirrors sourceCode.getTokenBefore(equalsToken,
// { includeComments: true }).range[1] from the upstream fixer. Between a type
// name (or its type parameters) and '=' only trivia can occur, so the last
// comment is the only token that can supersede the name as the preserved end.
func (f consistentTypeDefinitionsFixer) beforeEqualsEnd(from int, to int) (int, bool) {
	sourceScanner := scanner.GetScannerForSourceFile(f.sourceFile, from)
	sourceScanner.ResetTokenState(from)
	sourceScanner.SetSkipTrivia(false)
	sourceScanner.Scan()

	end := from
	for sourceScanner.TokenStart() < to {
		switch sourceScanner.Token() {
		case ast.KindEqualsToken:
			return end, true
		case ast.KindSingleLineCommentTrivia, ast.KindMultiLineCommentTrivia:
			end = sourceScanner.TokenEnd()
		case ast.KindEndOfFile:
			return 0, false
		}
		sourceScanner.Scan()
	}
	return 0, false
}

func (f consistentTypeDefinitionsFixer) typeAliasFix(node *ast.Node, typeAlias *ast.TypeAliasDeclaration) []rule.RuleFix {
	declarationRange := utils.TrimNodeTextRange(f.sourceFile, node)
	typeKeywordStart, typeKeywordEnd, ok := f.tokenBounds(declarationRange.Pos(), declarationRange.End(), ast.KindTypeKeyword)
	if !ok {
		return nil
	}

	nameRange := utils.TrimNodeTextRange(f.sourceFile, typeAlias.Name())
	nameOrTypeParametersEnd := nameRange.End()
	if typeAlias.TypeParameters != nil && len(typeAlias.TypeParameters.Nodes) > 0 {
		nameOrTypeParametersEnd = scanner.GetRangeOfTokenAtPosition(f.sourceFile, typeAlias.TypeParameters.End()).End()
	}

	bodyRange := utils.TrimNodeTextRange(f.sourceFile, ast.SkipTypeParentheses(typeAlias.Type))
	beforeEqualsEnd, ok := f.beforeEqualsEnd(nameOrTypeParametersEnd, bodyRange.Pos())
	if !ok {
		return nil
	}

	return []rule.RuleFix{
		// Match the upstream fixer's three edits so the merged ESLint-compatible
		// fix starts at `type`, leaving modifiers such as `export` untouched.
		rule.RuleFixReplaceRange(core.NewTextRange(typeKeywordStart, typeKeywordEnd), "interface"),
		rule.RuleFixReplaceRange(core.NewTextRange(beforeEqualsEnd, bodyRange.Pos()), " "),
		rule.RuleFixRemoveRange(core.NewTextRange(bodyRange.End(), declarationRange.End())),
	}
}

func (f consistentTypeDefinitionsFixer) interfaceFix(node *ast.Node, interfaceDeclaration *ast.InterfaceDeclaration) []rule.RuleFix {
	sourceText := f.sourceFile.Text()
	declarationRange := utils.TrimNodeTextRange(f.sourceFile, node)
	interfaceKeywordStart, interfaceKeywordEnd, ok := f.tokenBounds(declarationRange.Pos(), declarationRange.End(), ast.KindInterfaceKeyword)
	if !ok {
		return nil
	}

	nameRange := utils.TrimNodeTextRange(f.sourceFile, interfaceDeclaration.Name())
	nameText := sourceText[nameRange.Pos():nameRange.End()]
	typeNameEnd := nameRange.End()
	if interfaceDeclaration.TypeParameters != nil && len(interfaceDeclaration.TypeParameters.Nodes) > 0 {
		// NodeList.End() points at the closing `>` token's full start. Scanning
		// from there preserves comments, trailing commas, and whitespace before
		// `>` that a last-parameter-end + 1 calculation would truncate.
		typeNameEnd = scanner.GetRangeOfTokenAtPosition(f.sourceFile, interfaceDeclaration.TypeParameters.End()).End()
	}

	var bodyStartScanPosition int
	if interfaceDeclaration.HeritageClauses != nil && len(interfaceDeclaration.HeritageClauses.Nodes) > 0 {
		lastClause := interfaceDeclaration.HeritageClauses.Nodes[len(interfaceDeclaration.HeritageClauses.Nodes)-1]
		bodyStartScanPosition = lastClause.End()
	} else {
		bodyStartScanPosition = typeNameEnd
	}

	sourceScanner := scanner.GetScannerForSourceFile(f.sourceFile, bodyStartScanPosition)
	openBracePosition := -1
	for sourceScanner.TokenStart() < declarationRange.End() {
		if sourceScanner.Token() == ast.KindOpenBraceToken {
			openBracePosition = sourceScanner.TokenStart()
			break
		}
		sourceScanner.Scan()
	}
	if openBracePosition == -1 {
		return nil
	}

	var extendsTypes []string
	if interfaceDeclaration.HeritageClauses != nil {
		for _, clause := range interfaceDeclaration.HeritageClauses.Nodes {
			heritageClause := clause.AsHeritageClause()
			if heritageClause == nil || heritageClause.Token != ast.KindExtendsKeyword {
				continue
			}
			if heritageClause.Types != nil {
				for _, typeNode := range heritageClause.Types.Nodes {
					typeRange := utils.TrimNodeTextRange(f.sourceFile, typeNode)
					extendsTypes = append(extendsTypes, sourceText[typeRange.Pos():typeRange.End()])
				}
			}
		}
	}

	isDefaultExport := ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) && ast.HasSyntacticModifier(node, ast.ModifierFlagsDefault)
	fixCount := 2
	if isDefaultExport {
		fixCount += 2 // Remove the export wrapper and append a default export.
	} else if len(extendsTypes) > 0 {
		fixCount++ // Append the heritage types after the body.
	}
	fixes := make([]rule.RuleFix, 0, fixCount)
	if isDefaultExport {
		fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(declarationRange.Pos(), interfaceKeywordStart)))
	}
	fixes = append(fixes,
		rule.RuleFixReplaceRange(core.NewTextRange(interfaceKeywordStart, interfaceKeywordEnd), "type"),
		rule.RuleFixReplaceRange(core.NewTextRange(typeNameEnd, openBracePosition), " = "),
	)

	var suffix string
	if len(extendsTypes) > 0 {
		suffix += " & " + strings.Join(extendsTypes, " & ")
	}
	if isDefaultExport {
		suffix += "\nexport default " + nameText
	}
	if suffix != "" {
		fixes = append(fixes, rule.RuleFixReplaceRange(
			core.NewTextRange(declarationRange.End(), declarationRange.End()),
			suffix,
		))
	}

	return fixes
}

// ConsistentTypeDefinitionsRule enforces consistent type definitions
var ConsistentTypeDefinitionsRule = rule.CreateRule(rule.Rule{
	Name:   "consistent-type-definitions",
	Schema: rule.NewSchema(schemaJSON),
	Run:    run,
})

var (
	interfaceOverTypeMessage = rule.RuleMessage{
		Id:          "interfaceOverType",
		Description: "Use an `interface` instead of a `type`.",
	}
	typeOverInterfaceMessage = rule.RuleMessage{
		Id:          "typeOverInterface",
		Description: "Use a `type` instead of an `interface`.",
	}
)

func isInDeclareGlobal(node *ast.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsGlobalScopeAugmentation(current) && ast.HasSyntacticModifier(current, ast.ModifierFlagsAmbient) {
			return true
		}
	}
	return false
}

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := parseOptions(options)

	fixer := consistentTypeDefinitionsFixer{
		sourceFile: ctx.SourceFile,
	}

	switch opts.Style {
	case DefinitionStyleInterface:
		return rule.RuleListeners{
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				typeAlias := node.AsTypeAliasDeclaration()
				if typeAlias == nil || typeAlias.Type == nil || ast.SkipTypeParentheses(typeAlias.Type).Kind != ast.KindTypeLiteral {
					return
				}

				ctx.ReportNodeWithDeferredFixes(typeAlias.Name(), interfaceOverTypeMessage, func() []rule.RuleFix {
					return fixer.typeAliasFix(node, typeAlias)
				})
			},
		}
	case DefinitionStyleType:
		return rule.RuleListeners{
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				interfaceDeclaration := node.AsInterfaceDeclaration()
				if interfaceDeclaration == nil {
					return
				}

				// Don't fix interfaces in declare global modules (see typescript-eslint #2707).
				if isInDeclareGlobal(node) {
					ctx.ReportNode(interfaceDeclaration.Name(), typeOverInterfaceMessage)
					return
				}

				ctx.ReportNodeWithDeferredFixes(interfaceDeclaration.Name(), typeOverInterfaceMessage, func() []rule.RuleFix {
					return fixer.interfaceFix(node, interfaceDeclaration)
				})
			},
		}
	default:
		return nil
	}
}
