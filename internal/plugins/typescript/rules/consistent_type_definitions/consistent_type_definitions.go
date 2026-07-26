package consistent_type_definitions

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type DefinitionStyle string

const (
	DefinitionStyleInterface DefinitionStyle = "interface"
	DefinitionStyleType      DefinitionStyle = "type"
)

type ConsistentTypeDefinitionsOptions struct {
	Style DefinitionStyle `json:"style"`
}

type consistentTypeDefinitionsFixer struct {
	sourceFile *ast.SourceFile
}

// modifiersInfo builds a modifier prefix such as "export declare ". The
// default modifier is intentionally excluded because export default
// interfaces need to become a type declaration plus a separate default
// export.
func modifiersInfo(modifiers *ast.ModifierList) (prefix string, hasExport bool, hasDefault bool) {
	if modifiers == nil {
		return "", false, false
	}
	hasDeclare := false
	for _, mod := range modifiers.Nodes {
		switch mod.Kind {
		case ast.KindExportKeyword:
			hasExport = true
		case ast.KindDefaultKeyword:
			hasDefault = true
		case ast.KindDeclareKeyword:
			hasDeclare = true
		}
	}
	var parts []string
	if hasExport {
		parts = append(parts, "export")
	}
	if hasDeclare {
		parts = append(parts, "declare")
	}
	if len(parts) > 0 {
		prefix = strings.Join(parts, " ") + " "
	}
	return prefix, hasExport, hasDefault
}

func (f consistentTypeDefinitionsFixer) typeParametersText(sourceText string, typeParameters *ast.NodeList) string {
	if typeParameters == nil || len(typeParameters.Nodes) == 0 {
		return ""
	}
	firstParameter := typeParameters.Nodes[0]
	lastParameter := typeParameters.Nodes[len(typeParameters.Nodes)-1]
	firstRange := utils.TrimNodeTextRange(f.sourceFile, firstParameter)
	lastRange := utils.TrimNodeTextRange(f.sourceFile, lastParameter)
	return sourceText[firstRange.Pos()-1 : lastRange.End()+1]
}

func (f consistentTypeDefinitionsFixer) typeAliasFix(node *ast.Node, typeAlias *ast.TypeAliasDeclaration) []rule.RuleFix {
	sourceText := f.sourceFile.Text()
	nameRange := utils.TrimNodeTextRange(f.sourceFile, typeAlias.Name())
	nameText := sourceText[nameRange.Pos():nameRange.End()]
	typeParametersText := f.typeParametersText(sourceText, typeAlias.TypeParameters)
	prefix, _, _ := modifiersInfo(typeAlias.Modifiers())

	unwrapped := ast.SkipTypeParentheses(typeAlias.Type)
	bodyRange := utils.TrimNodeTextRange(f.sourceFile, unwrapped)
	bodyText := sourceText[bodyRange.Pos():bodyRange.End()]

	var commentText string
	nameOrTypeParametersEnd := nameRange.End()
	if typeAlias.TypeParameters != nil && len(typeAlias.TypeParameters.Nodes) > 0 {
		lastParameter := typeAlias.TypeParameters.Nodes[len(typeAlias.TypeParameters.Nodes)-1]
		lastRange := utils.TrimNodeTextRange(f.sourceFile, lastParameter)
		nameOrTypeParametersEnd = lastRange.End() + 1
	}

	typeRange := utils.TrimNodeTextRange(f.sourceFile, typeAlias.Type)
	betweenText := sourceText[nameOrTypeParametersEnd:typeRange.Pos()]
	if index := strings.Index(betweenText, "/*"); index >= 0 {
		endIndex := strings.Index(betweenText, "*/")
		if endIndex >= 0 {
			commentText = " " + strings.TrimSpace(betweenText[index:endIndex+2]) + " "
		}
	}

	replacement := prefix + "interface " + nameText + typeParametersText + " " + bodyText
	if commentText != "" {
		replacement = prefix + "interface " + nameText + typeParametersText + commentText + bodyText
	}

	declarationRange := utils.TrimNodeTextRange(f.sourceFile, node)
	return []rule.RuleFix{rule.RuleFixReplaceRange(declarationRange, replacement)}
}

func (f consistentTypeDefinitionsFixer) interfaceFix(node *ast.Node, interfaceDeclaration *ast.InterfaceDeclaration) []rule.RuleFix {
	sourceText := f.sourceFile.Text()
	nameRange := utils.TrimNodeTextRange(f.sourceFile, interfaceDeclaration.Name())
	nameText := sourceText[nameRange.Pos():nameRange.End()]
	typeParametersText := f.typeParametersText(sourceText, interfaceDeclaration.TypeParameters)
	prefix, hasExport, hasDefault := modifiersInfo(interfaceDeclaration.Modifiers())
	declarationRange := utils.TrimNodeTextRange(f.sourceFile, node)

	var bodyStartScanPosition int
	if interfaceDeclaration.HeritageClauses != nil && len(interfaceDeclaration.HeritageClauses.Nodes) > 0 {
		lastClause := interfaceDeclaration.HeritageClauses.Nodes[len(interfaceDeclaration.HeritageClauses.Nodes)-1]
		bodyStartScanPosition = lastClause.End()
	} else if interfaceDeclaration.TypeParameters != nil && len(interfaceDeclaration.TypeParameters.Nodes) > 0 {
		lastParameter := interfaceDeclaration.TypeParameters.Nodes[len(interfaceDeclaration.TypeParameters.Nodes)-1]
		lastRange := utils.TrimNodeTextRange(f.sourceFile, lastParameter)
		bodyStartScanPosition = lastRange.End() + 1
	} else {
		bodyStartScanPosition = nameRange.End()
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

	bodyText := sourceText[openBracePosition:declarationRange.End()]
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

	if hasExport && hasDefault {
		replacement := "type " + nameText + typeParametersText + " = " + bodyText
		if len(extendsTypes) > 0 {
			replacement += " & " + strings.Join(extendsTypes, " & ")
		}
		replacement += "\nexport default " + nameText
		return []rule.RuleFix{rule.RuleFixReplaceRange(declarationRange, replacement)}
	}

	replacement := prefix + "type " + nameText + typeParametersText + " = " + bodyText
	if len(extendsTypes) > 0 {
		replacement += " & " + strings.Join(extendsTypes, " & ")
	}

	declarationEnd := declarationRange.End()
	if declarationEnd < len(sourceText) && sourceText[declarationEnd] == ';' {
		fixRange := declarationRange.WithEnd(declarationEnd + 1)
		replacement += ";"
		return []rule.RuleFix{rule.RuleFixReplaceRange(fixRange, replacement)}
	}

	return []rule.RuleFix{rule.RuleFixReplaceRange(declarationRange, replacement)}
}

// ConsistentTypeDefinitionsRule enforces consistent type definitions
var ConsistentTypeDefinitionsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-definitions",
	Run:  run,
})

func run(ctx rule.RuleContext, _options []any) rule.RuleListeners {
	options := rule.LegacyUnwrapOptions(_options)
	opts := ConsistentTypeDefinitionsOptions{
		Style: DefinitionStyleInterface,
	}

	// Parse options
	if options != nil {
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if str, ok := optArray[0].(string); ok {
				opts.Style = DefinitionStyle(str)
			}
		} else if str, ok := options.(string); ok {
			opts.Style = DefinitionStyle(str)
		}
	}

	fixer := consistentTypeDefinitionsFixer{
		sourceFile: ctx.SourceFile,
	}

	// Helper to check if a type is an object type literal
	isObjectTypeLiteral := func(typeNode *ast.Node) bool {
		if typeNode == nil {
			return false
		}
		return typeNode.Kind == ast.KindTypeLiteral
	}

	// Helper to check if a type alias is a simple object type (not a union, intersection, etc.)
	// Unwraps any number of parenthesized type wrappers before checking.
	isSimpleObjectType := func(typeNode *ast.Node) bool {
		if typeNode == nil {
			return false
		}

		// Unwrap all layers of parenthesized types
		unwrapped := ast.SkipTypeParentheses(typeNode)
		return isObjectTypeLiteral(unwrapped)
	}

	// Helper to check if a modifier list includes a specific modifier kind
	hasModifier := func(modifiers *ast.ModifierList, kind ast.Kind) bool {
		if modifiers == nil {
			return false
		}
		for _, mod := range modifiers.Nodes {
			if mod.Kind == kind {
				return true
			}
		}
		return false
	}

	// Helper to check if interface is in a declare global module
	isInDeclareGlobal := func(node *ast.Node) bool {
		current := node.Parent
		for current != nil {
			if current.Kind == ast.KindModuleDeclaration {
				moduleDecl := current.AsModuleDeclaration()
				if moduleDecl != nil && moduleDecl.Name() != nil {
					if ast.IsIdentifier(moduleDecl.Name()) {
						ident := moduleDecl.Name().AsIdentifier()
						if ident != nil && ident.Text == "global" {
							// Only return true if module has 'declare' keyword
							if hasModifier(moduleDecl.Modifiers(), ast.KindDeclareKeyword) {
								return true
							}
						}
					}
				}
			}
			current = current.Parent
		}
		return false
	}

	interfaceOverTypeMessage := rule.RuleMessage{
		Id:          "interfaceOverType",
		Description: "Use an interface instead of a type literal.",
	}
	typeOverInterfaceMessage := rule.RuleMessage{
		Id:          "typeOverInterface",
		Description: "Use a type literal instead of an interface.",
	}

	checkTypeAlias := func(node *ast.Node) {
		if opts.Style != DefinitionStyleInterface {
			return
		}

		typeAlias := node.AsTypeAliasDeclaration()
		if typeAlias == nil {
			return
		}

		// Only report if it's a simple object type literal
		if !isSimpleObjectType(typeAlias.Type) {
			return
		}

		ctx.ReportNodeWithDeferredFixes(typeAlias.Name(), interfaceOverTypeMessage, func() []rule.RuleFix {
			return fixer.typeAliasFix(node, typeAlias)
		})
	}

	checkInterface := func(node *ast.Node) {
		if opts.Style != DefinitionStyleType {
			return
		}

		interfaceDecl := node.AsInterfaceDeclaration()
		if interfaceDecl == nil {
			return
		}

		// Don't fix interfaces in declare global modules (see typescript-eslint #2707)
		if isInDeclareGlobal(node) {
			ctx.ReportNode(interfaceDecl.Name(), typeOverInterfaceMessage)
			return
		}

		ctx.ReportNodeWithDeferredFixes(interfaceDecl.Name(), typeOverInterfaceMessage, func() []rule.RuleFix {
			return fixer.interfaceFix(node, interfaceDecl)
		})
	}

	return rule.RuleListeners{
		ast.KindTypeAliasDeclaration: checkTypeAlias,
		ast.KindInterfaceDeclaration: checkInterface,
	}
}
