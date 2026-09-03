package no_empty_interface

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_empty_interface.schema.json
var schemaJSON []byte

type NoEmptyInterfaceOptions struct {
	AllowSingleExtends bool `json:"allowSingleExtends"`
}

func parseOptions(options []any) NoEmptyInterfaceOptions {
	opts := NoEmptyInterfaceOptions{
		AllowSingleExtends: false,
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})
	if allowSingleExtends, ok := optsMap["allowSingleExtends"].(bool); ok {
		opts.AllowSingleExtends = allowSingleExtends
	}
	return opts
}

var (
	noEmptyMessage = rule.RuleMessage{
		Id:          "noEmpty",
		Description: "An empty interface is equivalent to `{}`.",
	}
	noEmptyWithSuperMessage = rule.RuleMessage{
		Id:          "noEmptyWithSuper",
		Description: "An interface declaring no members is equivalent to its supertype.",
	}
)

func extendedType(interfaceDecl *ast.InterfaceDeclaration) (int, *ast.Node) {
	if interfaceDecl.HeritageClauses == nil {
		return 0, nil
	}

	count := 0
	var single *ast.Node
	for _, clauseNode := range interfaceDecl.HeritageClauses.Nodes {
		clause := clauseNode.AsHeritageClause()
		if clause == nil || clause.Token != ast.KindExtendsKeyword || clause.Types == nil {
			continue
		}
		for _, typeNode := range clause.Types.Nodes {
			if typeNode == nil {
				continue
			}
			count++
			if count == 1 {
				single = typeNode
			} else {
				return count, nil
			}
		}
	}
	return count, single
}

const declarationScanCacheThreshold = 8

// mergeDetector mirrors upstream's same-scope class lookup with binder symbols.
// Most merges are pairs, so cache only unusually large declaration sets to keep
// the common path allocation-free without making adversarial input quadratic.
type mergeDetector struct {
	classScopesBySymbol map[*ast.Symbol]map[*ast.Node]bool
}

func (d *mergeDetector) isMergedWithClass(node *ast.Node) bool {
	symbol := node.Symbol()
	if symbol == nil || len(symbol.Declarations) <= 1 {
		return false
	}

	scope := utils.FindEnclosingScope(node)
	if len(symbol.Declarations) <= declarationScanCacheThreshold {
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindClassDeclaration && utils.FindEnclosingScope(declaration) == scope {
				return true
			}
		}
		return false
	}

	if d.classScopesBySymbol == nil {
		d.classScopesBySymbol = make(map[*ast.Symbol]map[*ast.Node]bool)
	}
	classScopes := d.classScopesBySymbol[symbol]
	if classScopes == nil {
		classScopes = make(map[*ast.Node]bool)
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindClassDeclaration {
				classScopes[utils.FindEnclosingScope(declaration)] = true
			}
		}
		d.classScopesBySymbol[symbol] = classScopes
	}
	return classScopes[scope]
}

// isInDeclaredModule reproduces the upstream scope check. Only the module
// scope directly containing the interface counts; an outer declared namespace
// must not affect a nested, non-declared namespace. Dotted namespace names are
// represented as nested ModuleDeclarations without an intervening ModuleBlock,
// so their segments form one scope for this purpose.
func isInDeclaredModule(node *ast.Node, sourceFile *ast.SourceFile) bool {
	if !sourceFile.IsDeclarationFile {
		return false
	}

	scope := utils.FindEnclosingScope(node)
	if scope == nil || scope.Kind != ast.KindModuleBlock || scope.Parent == nil || scope.Parent.Kind != ast.KindModuleDeclaration {
		return false
	}

	for module := scope.Parent; module != nil && module.Kind == ast.KindModuleDeclaration; module = module.Parent {
		if ast.HasSyntacticModifier(module, ast.ModifierFlagsAmbient) {
			return true
		}
		if !utils.IsQualifiedNamespaceSegment(module) {
			return false
		}
	}
	return false
}

func validSourceRange(textRange core.TextRange, sourceLength int) bool {
	return textRange.Pos() >= 0 && textRange.Pos() <= textRange.End() && textRange.End() <= sourceLength
}

type interfaceEditBuilder struct {
	sourceFile    *ast.SourceFile
	node          *ast.Node
	interfaceDecl *ast.InterfaceDeclaration
	extendedType  *ast.Node
}

// build preserves export/trivia preceding `interface` and the complete source
// spelling of generic parameters. An explicit `declare` belongs to the ESTree
// interface node upstream, so it is intentionally removed by starting there.
func (b interfaceEditBuilder) build() []rule.RuleFix {
	declarationRange := utils.TrimNodeTextRange(b.sourceFile, b.node)
	nameRange := utils.TrimNodeTextRange(b.sourceFile, b.interfaceDecl.Name())
	extendedRange := utils.TrimNodeTextRange(b.sourceFile, b.extendedType)
	sourceText := b.sourceFile.Text()
	if !validSourceRange(declarationRange, len(sourceText)) ||
		!validSourceRange(nameRange, len(sourceText)) ||
		!validSourceRange(extendedRange, len(sourceText)) ||
		nameRange.Pos() < declarationRange.Pos() ||
		extendedRange.Pos() < declarationRange.Pos() ||
		extendedRange.End() > declarationRange.End() {
		return nil
	}

	nameEnd := nameRange.End()
	if typeParameters := b.interfaceDecl.TypeParameters; typeParameters != nil && len(typeParameters.Nodes) > 0 {
		closingToken := scanner.GetRangeOfTokenAtPosition(b.sourceFile, typeParameters.End())
		if !validSourceRange(closingToken, len(sourceText)) || closingToken.End() < nameEnd {
			return nil
		}
		nameEnd = closingToken.End()
	}

	fixStart := -1
	interfaceStart := -1
	sourceScanner := scanner.GetScannerForSourceFile(b.sourceFile, declarationRange.Pos())
	for sourceScanner.TokenStart() < nameRange.Pos() {
		switch sourceScanner.Token() {
		case ast.KindDeclareKeyword:
			fixStart = sourceScanner.TokenStart()
		case ast.KindInterfaceKeyword:
			interfaceStart = sourceScanner.TokenStart()
		}
		if sourceScanner.Token() == ast.KindEndOfFile {
			break
		}
		sourceScanner.Scan()
	}
	if interfaceStart < 0 {
		return nil
	}
	if fixStart < 0 {
		fixStart = interfaceStart
	}
	fixRange := core.NewTextRange(fixStart, declarationRange.End())
	if !validSourceRange(fixRange, len(sourceText)) || nameEnd > declarationRange.End() {
		return nil
	}

	replacement := "type " + sourceText[nameRange.Pos():nameEnd] + " = " + sourceText[extendedRange.Pos():extendedRange.End()]
	return []rule.RuleFix{rule.RuleFixReplaceRange(fixRange, replacement)}
}

type interfaceEditKind uint8

const (
	interfaceEditUnknown interfaceEditKind = iota
	interfaceEditNone
	interfaceEditAutofix
	interfaceEditSuggestion
)

type interfaceEditPlan struct {
	ctx           rule.RuleContext
	node          *ast.Node
	builder       interfaceEditBuilder
	mergeDetector *mergeDetector
	kind          interfaceEditKind
}

func (p *interfaceEditPlan) classify() interfaceEditKind {
	if p.kind != interfaceEditUnknown {
		return p.kind
	}

	// The upstream fixer produces invalid syntax for a default-exported type
	// alias. Keep the diagnostic but withhold an unsafe edit.
	if ast.HasSyntacticModifier(p.node, ast.ModifierFlagsDefault) {
		p.kind = interfaceEditNone
	} else if p.mergeDetector.isMergedWithClass(p.node) {
		p.kind = interfaceEditNone
	} else if isInDeclaredModule(p.node, p.ctx.SourceFile) {
		p.kind = interfaceEditSuggestion
	} else {
		p.kind = interfaceEditAutofix
	}
	return p.kind
}

func (p *interfaceEditPlan) buildFixes() []rule.RuleFix {
	if p.classify() != interfaceEditAutofix {
		return nil
	}
	return p.builder.build()
}

func (p *interfaceEditPlan) buildSuggestions() []rule.RuleSuggestion {
	if p.classify() != interfaceEditSuggestion {
		return nil
	}
	fixes := p.builder.build()
	if len(fixes) == 0 {
		return nil
	}
	return []rule.RuleSuggestion{{
		Message:  noEmptyWithSuperMessage,
		FixesArr: fixes,
	}}
}

var NoEmptyInterfaceRule = rule.CreateRule(rule.Rule{
	Name:   "no-empty-interface",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		merges := mergeDetector{}

		return rule.RuleListeners{
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				interfaceDecl := node.AsInterfaceDeclaration()
				if interfaceDecl == nil || interfaceDecl.Name() == nil {
					return
				}

				if interfaceDecl.Members != nil && len(interfaceDecl.Members.Nodes) > 0 {
					return
				}

				extendsCount, superType := extendedType(interfaceDecl)
				if extendsCount == 0 {
					ctx.ReportNode(interfaceDecl.Name(), noEmptyMessage)
					return
				}

				if extendsCount != 1 || opts.AllowSingleExtends {
					return
				}

				plan := interfaceEditPlan{
					ctx:           ctx,
					node:          node,
					mergeDetector: &merges,
					builder: interfaceEditBuilder{
						sourceFile:    ctx.SourceFile,
						node:          node,
						interfaceDecl: interfaceDecl,
						extendedType:  superType,
					},
				}
				ctx.ReportNodeWithDeferredFixesAndSuggestions(
					interfaceDecl.Name(),
					noEmptyWithSuperMessage,
					plan.buildFixes,
					plan.buildSuggestions,
				)
			},
		}
	},
})
