package consistent_type_imports

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/typescriptutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed consistent_type_imports.schema.json
var schemaJSON []byte

type consistentTypeImportsOptions struct {
	prefer                  string
	disallowTypeAnnotations bool
	fixStyle                string
}

type importSpecifierKind uint8

const (
	importSpecifierDefault importSpecifierKind = iota
	importSpecifierNamed
	importSpecifierNamespace
)

type importSpecifier struct {
	node       *ast.Node
	name       *ast.Node
	kind       importSpecifierKind
	inlineType bool
}

type reportValueImport struct {
	node             *ast.Node
	allSpecifiers    []importSpecifier
	typeSpecifiers   []importSpecifier
	unusedSpecifiers []importSpecifier
	valueSpecifiers  []importSpecifier
}

type sourceImports struct {
	reports             []*reportValueImport
	typeOnlyNamedImport *ast.Node
}

func messageAvoidImportType() rule.RuleMessage {
	return rule.RuleMessage{Id: "avoidImportType", Description: "Use an `import` instead of an `import type`."}
}

func messageNoImportTypeAnnotations() rule.RuleMessage {
	return rule.RuleMessage{Id: "noImportTypeAnnotations", Description: "`import()` type annotations are forbidden."}
}

func messageTypeOverValue() rule.RuleMessage {
	return rule.RuleMessage{Id: "typeOverValue", Description: "All imports in the declaration are only used as types. Use `import type`."}
}

func messageSomeImportsAreOnlyTypes(specifiers []importSpecifier) rule.RuleMessage {
	names := make([]string, 0, len(specifiers))
	for _, specifier := range specifiers {
		names = append(names, `"`+specifier.name.Text()+`"`)
	}
	typeImports := formatWordList(names)
	return rule.RuleMessage{
		Id:          "someImportsAreOnlyTypes",
		Description: "Imports " + typeImports + " are only used as type.",
		Data:        map[string]string{"typeImports": typeImports},
	}
}

func formatWordList(words []string) string {
	switch len(words) {
	case 0:
		return ""
	case 1:
		return words[0]
	case 2:
		return words[0] + " and " + words[1]
	default:
		return strings.Join(words[:len(words)-1], ", ") + " and " + words[len(words)-1]
	}
}

// ConsistentTypeImportsRule enforces consistent type-only import syntax.
var ConsistentTypeImportsRule = rule.CreateRule(rule.Rule{
	Name:   "consistent-type-imports",
	Schema: rule.NewSchema(schemaJSON),
	Run:    run,
})

func parseOptions(options []any) consistentTypeImportsOptions {
	opts := consistentTypeImportsOptions{
		prefer:                  "type-imports",
		disallowTypeAnnotations: true,
		fixStyle:                "separate-type-imports",
	}
	if len(options) == 0 {
		return opts
	}
	option, _ := options[0].(map[string]interface{})
	if prefer, ok := option["prefer"].(string); ok {
		opts.prefer = prefer
	}
	if disallow, ok := option["disallowTypeAnnotations"].(bool); ok {
		opts.disallowTypeAnnotations = disallow
	}
	if fixStyle, ok := option["fixStyle"].(string); ok {
		opts.fixStyle = fixStyle
	}
	return opts
}

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := parseOptions(options)
	listeners := rule.RuleListeners{}

	if opts.disallowTypeAnnotations {
		listeners[ast.KindImportType] = func(node *ast.Node) {
			reportRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			if importType := node.AsImportTypeNode(); importType != nil && importType.IsTypeOf {
				for _, token := range utils.TokensOfNode(ctx.SourceFile, node) {
					if token.Kind == ast.KindImportKeyword {
						reportRange = core.NewTextRange(token.Start, reportRange.End())
						break
					}
				}
			}
			ctx.ReportRange(reportRange, messageNoImportTypeAnnotations())
		}
	}

	if opts.prefer == "no-type-imports" {
		listeners[ast.KindImportDeclaration] = func(node *ast.Node) {
			declaration := node.AsImportDeclaration()
			if declaration == nil || declaration.ImportClause == nil || !declaration.ImportClause.IsTypeOnly() {
				return
			}
			ctx.ReportNodeWithDeferredFixes(node, messageAvoidImportType(), func() []rule.RuleFix {
				return removeTypeModifierFixes(ctx, node)
			})
		}
		listeners[ast.KindImportSpecifier] = func(node *ast.Node) {
			specifier := node.AsImportSpecifier()
			if specifier == nil || !specifier.IsTypeOnly {
				return
			}
			ctx.ReportNodeWithDeferredFixes(node, messageAvoidImportType(), func() []rule.RuleFix {
				return removeTypeModifierFixes(ctx, node)
			})
		}
		return listeners
	}

	bySource := map[string]*sourceImports{}
	orderedSources := []*sourceImports{}
	hasDecorator := false
	hasJSX := false
	hasJSXFragment := false
	guardDecoratorMetadata := false
	if program := ctx.Program(); program != nil && program.Options() != nil {
		compilerOptions := program.Options()
		guardDecoratorMetadata = compilerOptions.ExperimentalDecorators.IsTrue() && compilerOptions.EmitDecoratorMetadata.IsTrue()
	}
	if guardDecoratorMetadata {
		listeners[ast.KindDecorator] = func(*ast.Node) { hasDecorator = true }
	}
	listeners[ast.KindJsxElement] = func(*ast.Node) { hasJSX = true }
	listeners[ast.KindJsxSelfClosingElement] = func(*ast.Node) { hasJSX = true }
	listeners[ast.KindJsxFragment] = func(*ast.Node) {
		hasJSX = true
		hasJSXFragment = true
	}

	listeners[ast.KindImportDeclaration] = func(node *ast.Node) {
		declaration := node.AsImportDeclaration()
		if declaration == nil || declaration.ImportClause == nil {
			return
		}
		clause := declaration.ImportClause.AsImportClause()
		if clause == nil {
			return
		}
		source := importSource(declaration.ModuleSpecifier)
		imports := bySource[source]
		if imports == nil {
			imports = &sourceImports{}
			bySource[source] = imports
			orderedSources = append(orderedSources, imports)
		}

		specifiers := collectImportSpecifiers(clause)
		if declaration.ImportClause.IsTypeOnly() {
			if imports.typeOnlyNamedImport == nil && allNamedSpecifiers(specifiers) {
				imports.typeOnlyNamedImport = node
			}
		}

		report := &reportValueImport{node: node, allSpecifiers: specifiers}
		for _, specifier := range specifiers {
			if specifier.inlineType {
				continue
			}
			references := referencesOfImportSpecifier(ctx, specifier)
			if len(references) == 0 {
				report.unusedSpecifiers = append(report.unusedSpecifiers, specifier)
				continue
			}
			onlyTypes := true
			for _, reference := range references {
				if !isTypeOnlyReference(reference, declaration.ImportClause.IsTypeOnly()) {
					onlyTypes = false
					break
				}
			}
			if onlyTypes {
				report.typeSpecifiers = append(report.typeSpecifiers, specifier)
			} else {
				report.valueSpecifiers = append(report.valueSpecifiers, specifier)
			}
		}
		if !declaration.ImportClause.IsTypeOnly() && len(report.typeSpecifiers) > 0 {
			imports.reports = append(imports.reports, report)
		}
	}

	listeners[ast.KindEndOfFile] = func(*ast.Node) {
		if hasDecorator {
			return
		}
		for _, imports := range orderedSources {
			for _, report := range imports.reports {
				applyImplicitJSXReferences(ctx, report, hasJSX, hasJSXFragment)
				if len(report.typeSpecifiers) == 0 {
					continue
				}
				allUsedImportsAreTypes := len(report.valueSpecifiers) == 0 && len(report.unusedSpecifiers) == 0
				if allUsedImportsAreTypes {
					declaration := report.node.AsImportDeclaration()
					if declaration.Attributes != nil {
						continue
					}
					ctx.ReportNodeWithDeferredFixes(report.node, messageTypeOverValue(), func() []rule.RuleFix {
						return fixesToTypeImport(ctx, opts.fixStyle, report, imports)
					})
					continue
				}
				ctx.ReportNodeWithDeferredFixes(report.node, messageSomeImportsAreOnlyTypes(report.typeSpecifiers), func() []rule.RuleFix {
					return fixesToTypeImport(ctx, opts.fixStyle, report, imports)
				})
			}
		}
	}
	return listeners
}

func applyImplicitJSXReferences(ctx rule.RuleContext, report *reportValueImport, hasJSX, hasFragment bool) {
	if ctx.Program() == nil || ctx.Program().Options() == nil {
		return
	}
	options := ctx.Program().Options()
	factory := "React"
	if options.JsxFactory != "" {
		factory = typescriptutil.JSXFactoryRoot(options.JsxFactory)
	}
	fragmentFactory := ""
	if options.JsxFragmentFactory != "" {
		fragmentFactory = typescriptutil.JSXFactoryRoot(options.JsxFragmentFactory)
	}
	remaining := report.typeSpecifiers[:0]
	for _, specifier := range report.typeSpecifiers {
		used := hasJSX && specifier.name.Text() == factory
		used = used || hasFragment && fragmentFactory != "" && specifier.name.Text() == fragmentFactory
		if used {
			report.valueSpecifiers = append(report.valueSpecifiers, specifier)
		} else {
			remaining = append(remaining, specifier)
		}
	}
	report.typeSpecifiers = remaining
}

func importSource(node *ast.Node) string {
	if node != nil && ast.IsStringLiteralLike(node) {
		return node.Text()
	}
	return ""
}

func collectImportSpecifiers(clause *ast.ImportClause) []importSpecifier {
	specifiers := []importSpecifier{}
	if clause.Name() != nil {
		specifiers = append(specifiers, importSpecifier{node: clause.AsNode(), name: clause.Name(), kind: importSpecifierDefault})
	}
	if clause.NamedBindings == nil {
		return specifiers
	}
	switch clause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		namespace := clause.NamedBindings.AsNamespaceImport()
		specifiers = append(specifiers, importSpecifier{node: clause.NamedBindings, name: namespace.Name(), kind: importSpecifierNamespace})
	case ast.KindNamedImports:
		named := clause.NamedBindings.AsNamedImports()
		if named != nil && named.Elements != nil {
			for _, node := range named.Elements.Nodes {
				specifier := node.AsImportSpecifier()
				if specifier != nil {
					specifiers = append(specifiers, importSpecifier{node: node, name: specifier.Name(), kind: importSpecifierNamed, inlineType: specifier.IsTypeOnly})
				}
			}
		}
	}
	return specifiers
}

func allNamedSpecifiers(specifiers []importSpecifier) bool {
	for _, specifier := range specifiers {
		if specifier.kind != importSpecifierNamed {
			return false
		}
	}
	return true
}

func referencesOfImportSpecifier(ctx rule.RuleContext, specifier importSpecifier) []*ast.Node {
	if ctx.Refs == nil {
		return nil
	}
	symbol := specifier.node.Symbol()
	if symbol == nil && specifier.name != nil {
		symbol = specifier.name.Symbol()
	}
	return ctx.Refs.References(symbol)
}

func isTypeOnlyReference(reference *ast.Node, importIsTypeOnly bool) bool {
	if reference == nil {
		return false
	}
	current := reference
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	if parent := current.Parent; parent != nil {
		switch parent.Kind {
		case ast.KindExportSpecifier:
			specifier := parent.AsExportSpecifier()
			if specifier != nil && specifier.IsTypeOnly {
				return true
			}
			if parent.Parent != nil && parent.Parent.Parent != nil {
				declaration := parent.Parent.Parent.AsExportDeclaration()
				if declaration != nil && declaration.IsTypeOnly {
					return true
				}
			}
			return importIsTypeOnly
		case ast.KindExportAssignment:
			return importIsTypeOnly
		}
	}
	entity := reference
	for entity.Parent != nil {
		parent := entity.Parent
		if parent.Kind == ast.KindQualifiedName && parent.AsQualifiedName().Left == entity {
			entity = parent
			continue
		}
		if parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == entity {
			if ast.IsOptionalChain(parent) {
				return false
			}
			entity = parent
			continue
		}
		if parent.Kind == ast.KindElementAccessExpression && parent.AsElementAccessExpression().Expression == entity {
			if ast.IsOptionalChain(parent) {
				return false
			}
			break
		}
		break
	}
	if ast.IsPartOfTypeQuery(entity) || ast.IsPartOfTypeNode(entity) {
		return true
	}

	// A computed key in a type property is represented as an expression even
	// though TypeScript can erase an import used only by that key.
	current = reference
	for current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			current = parent
		case ast.KindPropertyAccessExpression:
			if parent.AsPropertyAccessExpression().Expression != current {
				return false
			}
			if ast.IsOptionalChain(parent) {
				return false
			}
			current = parent
		case ast.KindElementAccessExpression:
			if parent.AsElementAccessExpression().Expression != current {
				return false
			}
			if ast.IsOptionalChain(parent) {
				return false
			}
			current = parent
		default:
			if parent.Kind == ast.KindComputedPropertyName && parent.Parent != nil && parent.Parent.Kind == ast.KindPropertySignature {
				return true
			}
			return false
		}
	}
	return false
}

func removeTypeModifierFixes(ctx rule.RuleContext, node *ast.Node) []rule.RuleFix {
	tokens := utils.TokensOfNode(ctx.SourceFile, node)
	for index, token := range tokens {
		if token.Kind != ast.KindTypeKeyword || index+1 >= len(tokens) {
			continue
		}
		end := firstCommentOrPosition(ctx.SourceFile.Text(), token.End, tokens[index+1].Start)
		return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(token.Start, end))}
	}
	return nil
}

func firstCommentOrPosition(text string, start int, end int) int {
	if start < 0 || end > len(text) || start >= end {
		return end
	}
	if offset := strings.Index(text[start:end], "//"); offset >= 0 {
		end = start + offset
	}
	if offset := strings.Index(text[start:end], "/*"); offset >= 0 && start+offset < end {
		end = start + offset
	}
	return end
}

func fixesToTypeImport(ctx rule.RuleContext, fixStyle string, report *reportValueImport, imports *sourceImports) []rule.RuleFix {
	defaultSpecifier, namedSpecifiers, namespaceSpecifier := classifySpecifiers(report.allSpecifiers)
	if namespaceSpecifier != nil && defaultSpecifier == nil {
		if report.node.AsImportDeclaration().Attributes != nil {
			return nil
		}
		return insertTopLevelTypeFixes(ctx, report.node, false)
	}

	if defaultSpecifier != nil {
		if containsSpecifier(report.typeSpecifiers, defaultSpecifier) && len(namedSpecifiers) == 0 && namespaceSpecifier == nil {
			return insertTopLevelTypeFixes(ctx, report.node, true)
		}
		if fixStyle == "inline-type-imports" && !containsSpecifier(report.typeSpecifiers, defaultSpecifier) && len(namedSpecifiers) > 0 && namespaceSpecifier == nil {
			return insertInlineTypeFixes(ctx, namedTypeSpecifiers(namedSpecifiers, report.typeSpecifiers))
		}
	} else if namespaceSpecifier == nil {
		if fixStyle == "inline-type-imports" && len(namedTypeSpecifiers(namedSpecifiers, report.typeSpecifiers)) > 0 {
			return insertInlineTypeFixes(ctx, namedTypeSpecifiers(namedSpecifiers, report.typeSpecifiers))
		}
		if len(namedSpecifiers) > 0 && allContained(namedSpecifiers, report.typeSpecifiers) {
			return insertTopLevelTypeFixes(ctx, report.node, false)
		}
	}

	fixes := []rule.RuleFix{}
	typeNamed := namedTypeSpecifiers(namedSpecifiers, report.typeSpecifiers)
	removeNamed, namedText := namedSpecifierFixes(ctx, typeNamed, namedSpecifiers)
	if len(typeNamed) > 0 {
		if imports.typeOnlyNamedImport != nil {
			fixes = append(fixes, insertNamedSpecifiersFix(ctx, imports.typeOnlyNamedImport, namedText))
		} else {
			sourceText := utils.TrimmedNodeText(ctx.SourceFile, report.node.AsImportDeclaration().ModuleSpecifier)
			prefix := "import type {" + namedText + "} from " + sourceText + ";\n"
			if fixStyle == "inline-type-imports" {
				parts := make([]string, 0, len(typeNamed))
				for _, specifier := range typeNamed {
					parts = append(parts, "type "+utils.TrimmedNodeText(ctx.SourceFile, specifier.node))
				}
				prefix = "import {" + strings.Join(parts, ", ") + "} from " + sourceText + ";\n"
			}
			position := utils.TrimNodeTextRange(ctx.SourceFile, report.node).Pos()
			fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(position, position), prefix))
		}
	}

	if namespaceSpecifier != nil && containsSpecifier(report.typeSpecifiers, namespaceSpecifier) {
		before, ok := utils.TokenBeforePosition(ctx.SourceFile, utils.TrimNodeTextRange(ctx.SourceFile, namespaceSpecifier.node).Pos())
		if ok && before.Kind == ast.KindCommaToken {
			fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(before.Start, utils.TrimNodeTextRange(ctx.SourceFile, namespaceSpecifier.node).End())))
		}
		sourceText := utils.TrimmedNodeText(ctx.SourceFile, report.node.AsImportDeclaration().ModuleSpecifier)
		position := utils.TrimNodeTextRange(ctx.SourceFile, report.node).Pos()
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(position, position), "import type "+utils.TrimmedNodeText(ctx.SourceFile, namespaceSpecifier.node)+" from "+sourceText+";\n"))
	}

	if defaultSpecifier != nil && containsSpecifier(report.typeSpecifiers, defaultSpecifier) {
		if len(report.typeSpecifiers) == len(report.allSpecifiers) {
			fixes = append(fixes, insertTypeAfterImportFix(ctx, report.node))
		} else {
			sourceText := utils.TrimmedNodeText(ctx.SourceFile, report.node.AsImportDeclaration().ModuleSpecifier)
			position := utils.TrimNodeTextRange(ctx.SourceFile, report.node).Pos()
			defaultRange := utils.TrimNodeTextRange(ctx.SourceFile, defaultSpecifier.name)
			comma, ok := utils.TokenAtOrAfter(ctx.SourceFile, defaultRange.End())
			if ok && comma.Kind == ast.KindCommaToken {
				defaultText := strings.TrimSpace(ctx.SourceFile.Text()[defaultRange.Pos():comma.Start])
				fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(position, position), "import type "+defaultText+" from "+sourceText+";\n"))
				after, ok := utils.TokenAtOrAfter(ctx.SourceFile, comma.End)
				if ok {
					removeEnd := firstCommentOrPosition(ctx.SourceFile.Text(), comma.End, after.Start)
					fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(defaultRange.Pos(), removeEnd)))
				}
			}
		}
	}
	fixes = append(fixes, removeNamed...)
	return fixes
}

func classifySpecifiers(specifiers []importSpecifier) (defaultSpecifier *importSpecifier, named []*importSpecifier, namespace *importSpecifier) {
	for index := range specifiers {
		specifier := &specifiers[index]
		switch specifier.kind {
		case importSpecifierDefault:
			defaultSpecifier = specifier
		case importSpecifierNamed:
			named = append(named, specifier)
		case importSpecifierNamespace:
			namespace = specifier
		}
	}
	return defaultSpecifier, named, namespace
}

func containsSpecifier(specifiers []importSpecifier, wanted *importSpecifier) bool {
	if wanted == nil {
		return false
	}
	for _, specifier := range specifiers {
		if specifier.node == wanted.node {
			return true
		}
	}
	return false
}

func allContained(specifiers []*importSpecifier, subset []importSpecifier) bool {
	for _, specifier := range specifiers {
		if !containsSpecifier(subset, specifier) {
			return false
		}
	}
	return true
}

func namedTypeSpecifiers(named []*importSpecifier, types []importSpecifier) []*importSpecifier {
	result := []*importSpecifier{}
	for _, specifier := range named {
		if containsSpecifier(types, specifier) {
			result = append(result, specifier)
		}
	}
	return result
}

func insertTypeAfterImportFix(ctx rule.RuleContext, node *ast.Node) rule.RuleFix {
	first := utils.TokensOfNode(ctx.SourceFile, node)[0]
	return rule.RuleFixReplaceRange(core.NewTextRange(first.End, first.End), " type")
}

func insertTopLevelTypeFixes(ctx rule.RuleContext, node *ast.Node, isDefault bool) []rule.RuleFix {
	fixes := []rule.RuleFix{insertTypeAfterImportFix(ctx, node)}
	declaration := node.AsImportDeclaration()
	clause := declaration.ImportClause.AsImportClause()
	if isDefault && clause.NamedBindings != nil && clause.NamedBindings.Kind == ast.KindNamedImports {
		named := clause.NamedBindings.AsNamedImports()
		if named != nil && (named.Elements == nil || len(named.Elements.Nodes) == 0) {
			tokens := utils.TokensOfNode(ctx.SourceFile, clause.NamedBindings)
			if len(tokens) >= 2 {
				comma, ok := utils.TokenBeforePosition(ctx.SourceFile, tokens[0].Start)
				if ok && comma.Kind == ast.KindCommaToken {
					fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(comma.Start, tokens[len(tokens)-1].End)))
				}
			}
		}
	}
	for _, specifier := range collectImportSpecifiers(clause) {
		if specifier.inlineType {
			fixes = append(fixes, removeTypeModifierFixes(ctx, specifier.node)...)
		}
	}
	return fixes
}

func insertInlineTypeFixes(ctx rule.RuleContext, specifiers []*importSpecifier) []rule.RuleFix {
	fixes := make([]rule.RuleFix, 0, len(specifiers))
	for _, specifier := range specifiers {
		range_ := utils.TrimNodeTextRange(ctx.SourceFile, specifier.node)
		fixes = append(fixes, rule.RuleFixReplaceRange(range_, "type "+ctx.SourceFile.Text()[range_.Pos():range_.End()]))
	}
	return fixes
}

func namedSpecifierFixes(ctx rule.RuleContext, subset, all []*importSpecifier) ([]rule.RuleFix, string) {
	if len(all) == 0 || len(subset) == 0 {
		return nil, ""
	}
	text := ctx.SourceFile.Text()
	if len(subset) == len(all) {
		firstRange := utils.TrimNodeTextRange(ctx.SourceFile, subset[0].node)
		lastRange := utils.TrimNodeTextRange(ctx.SourceFile, subset[len(subset)-1].node)
		opening, _ := utils.TokenBeforePosition(ctx.SourceFile, firstRange.Pos())
		closing, _ := utils.TokenAtOrAfter(ctx.SourceFile, lastRange.End())
		comma, _ := utils.TokenBeforePosition(ctx.SourceFile, opening.Start)
		if opening.Kind == ast.KindOpenBraceToken && closing.Kind == ast.KindCloseBraceToken && comma.Kind == ast.KindCommaToken {
			return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(comma.Start, closing.End))}, text[opening.End:closing.Start]
		}
	}

	groups := [][]*importSpecifier{}
	group := []*importSpecifier{}
	for _, specifier := range all {
		if containsSpecifierPointers(subset, specifier) {
			group = append(group, specifier)
		} else if len(group) > 0 {
			groups = append(groups, group)
			group = nil
		}
	}
	if len(group) > 0 {
		groups = append(groups, group)
	}
	fixes := []rule.RuleFix{}
	texts := []string{}
	for _, group := range groups {
		first := group[0]
		last := group[len(group)-1]
		firstRange := utils.TrimNodeTextRange(ctx.SourceFile, first.node)
		lastRange := utils.TrimNodeTextRange(ctx.SourceFile, last.node)
		before, _ := utils.TokenBeforePosition(ctx.SourceFile, firstRange.Pos())
		after, _ := utils.TokenAtOrAfter(ctx.SourceFile, lastRange.End())
		removeStart := before.End
		if before.Kind == ast.KindCommaToken {
			removeStart = before.Start
		}
		removeEnd := lastRange.End()
		isFirst := all[0].node == first.node
		isLast := all[len(all)-1].node == last.node
		if (isFirst || isLast) && after.Kind == ast.KindCommaToken {
			removeEnd = after.End
		}
		fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(removeStart, removeEnd)))
		texts = append(texts, text[before.End:after.Start])
	}
	return fixes, strings.Join(texts, ",")
}

func containsSpecifierPointers(specifiers []*importSpecifier, wanted *importSpecifier) bool {
	for _, specifier := range specifiers {
		if specifier.node == wanted.node {
			return true
		}
	}
	return false
}

func insertNamedSpecifiersFix(ctx rule.RuleContext, target *ast.Node, insertText string) rule.RuleFix {
	declaration := target.AsImportDeclaration()
	clause := declaration.ImportClause.AsImportClause()
	named := clause.NamedBindings.AsNamedImports()
	tokens := utils.TokensOfNode(ctx.SourceFile, named.AsNode())
	closing := tokens[len(tokens)-1]
	before, _ := utils.TokenBeforePosition(ctx.SourceFile, closing.Start)
	if before.Kind != ast.KindCommaToken && before.Kind != ast.KindOpenBraceToken {
		insertText = "," + insertText
	}
	return rule.RuleFixReplaceRange(core.NewTextRange(closing.Start, closing.Start), insertText)
}
