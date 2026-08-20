package no_duplicates

import (
	_ "embed"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed no_duplicates.schema.json
var schemaJSON []byte

type ruleOptions struct {
	considerQueryString bool
	preferInline        bool
}

func parseOptions(options []any) ruleOptions {
	opts := ruleOptions{}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})
	opts.considerQueryString, _ = optsMap["considerQueryString"].(bool)
	opts.preferInline, _ = optsMap["prefer-inline"].(bool)
	return opts
}

// See: https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/no-duplicates.md
var NoDuplicatesRule = rule.Rule{
	Name:   "import/no-duplicates",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		sourceFile := ctx.SourceFile
		if sourceFile == nil || sourceFile.Statements == nil {
			return rule.RuleListeners{}
		}

		sourceText := sourceFile.Text()
		resolver := importResolver{
			ctx:        ctx,
			opts:       opts,
			sourceText: sourceText,
		}

		// ESLint groups imports by parent scope (n.parent), so imports inside
		// `declare module` blocks are checked independently from top-level imports.
		// processScope recursively handles those independent statement lists.
		processScope(&resolver, sourceFile.Statements.Nodes)

		return rule.RuleListeners{}
	},
}

type importCategory uint8

const (
	valueImport importCategory = iota
	namespaceImport
	defaultTypeImport
	namedTypeImport
	importCategoryCount
)

type importMap struct {
	entries map[string]importEntry
	groups  []duplicateGroup
}

type importEntry struct {
	first      *ast.Node
	groupIndex int // one-based index into importMap.groups; zero means unique
}

type duplicateGroup struct {
	module string
	first  *ast.Node
	rest   []*ast.Node
}

// processScope collects and checks duplicate imports within a single scope (statement list).
func processScope(resolver *importResolver, statements []*ast.Node) {
	// The four categories mirror the ESLint rule's import maps. Maps are
	// allocated lazily because most scopes use only one category (or none).
	var imports [importCategoryCount]importMap
	var moduleDeclarations []*ast.Node

	for _, stmt := range statements {
		if stmt.Kind == ast.KindModuleDeclaration {
			moduleDeclarations = append(moduleDeclarations, stmt)
			continue
		}
		if stmt.Kind != ast.KindImportDeclaration {
			continue
		}

		importDecl := stmt.AsImportDeclaration()
		if importDecl.ModuleSpecifier == nil {
			continue
		}

		resolvedPath := resolver.resolve(importDecl)
		if resolvedPath == "" {
			continue
		}

		category := getImportCategory(importDecl, resolver.opts)
		importMap := &imports[category]
		if importMap.entries == nil {
			importMap.entries = make(map[string]importEntry)
		}

		entry, exists := importMap.entries[resolvedPath]
		if !exists {
			importMap.entries[resolvedPath] = importEntry{first: stmt}
			continue
		}
		if entry.groupIndex == 0 {
			importMap.groups = append(importMap.groups, duplicateGroup{
				module: resolvedPath,
				first:  entry.first,
				rest:   []*ast.Node{stmt},
			})
			entry.groupIndex = len(importMap.groups)
			importMap.entries[resolvedPath] = entry
			continue
		}
		group := &importMap.groups[entry.groupIndex-1]
		group.rest = append(group.rest, stmt)
	}

	for i := range imports {
		checkImports(resolver, imports[i])
	}

	for _, declaration := range moduleDeclarations {
		processModuleDeclaration(resolver, declaration)
	}
}

// processModuleDeclaration follows dotted namespace declarations until their
// module block, then processes that block as its own import scope.
func processModuleDeclaration(resolver *importResolver, declaration *ast.Node) {
	body := declaration.AsModuleDeclaration().Body
	if body == nil {
		return
	}
	switch body.Kind {
	case ast.KindModuleBlock:
		blockStatements := body.AsModuleBlock().Statements
		if blockStatements != nil && len(blockStatements.Nodes) > 0 {
			processScope(resolver, blockStatements.Nodes)
		}
	case ast.KindModuleDeclaration:
		// `declare module A.B { ... }` nests as ModuleDeclaration → ModuleDeclaration → ModuleBlock.
		processModuleDeclaration(resolver, body)
	}
}

// getModuleSpecifierText returns the string content of a module specifier node.
// Falls back to extracting text from source when the AST field is empty
// (can happen when the module specifier resolves to an empty filename).
func getModuleSpecifierText(moduleSpecifier *ast.Node, sourceText string) string {
	if moduleSpecifier == nil {
		return ""
	}
	// Use the standard utility to get the string literal value.
	if text := rslintUtils.GetStaticStringValue(moduleSpecifier); text != "" {
		return text
	}
	// Fallback: extract text directly from source, stripping quotes.
	pos := scanner.SkipTrivia(sourceText, moduleSpecifier.Pos())
	end := moduleSpecifier.End()
	if pos >= end || pos >= len(sourceText) || end > len(sourceText) {
		return ""
	}
	raw := sourceText[pos:end]
	if len(raw) >= 2 {
		first := raw[0]
		last := raw[len(raw)-1]
		if (first == '\'' && last == '\'') || (first == '"' && last == '"') || (first == '`' && last == '`') {
			return raw[1 : len(raw)-1]
		}
	}
	return raw
}

type importResolver struct {
	ctx            rule.RuleContext
	opts           ruleOptions
	sourceText     string
	normalMode     core.ResolutionMode
	hasNormalMode  bool
	commentsKnown  bool
	hasComments    bool
	commentFactory *ast.NodeFactory
}

// resolve returns the canonical grouping path for an import. Resolution mode
// is identical for ordinary import declarations in one source file, so it is
// computed once. Type-only declarations with a resolution-mode override keep
// their per-declaration mode.
func (r *importResolver) resolve(importDecl *ast.ImportDeclaration) string {
	moduleSpecifier := importDecl.ModuleSpecifier
	if moduleSpecifier == nil || !ast.IsStringLiteralLike(moduleSpecifier) {
		return ""
	}

	sourcePath := getModuleSpecifierText(moduleSpecifier, r.sourceText)

	if r.opts.considerQueryString {
		idx := strings.Index(sourcePath, "?")
		if idx >= 0 {
			query := sourcePath[idx:]
			if resolved, ok := r.resolveModule(importDecl, moduleSpecifier); ok {
				return resolved + query
			}
			return sourcePath[:idx] + query
		}
	}

	if resolved, ok := r.resolveModule(importDecl, moduleSpecifier); ok {
		return resolved
	}

	// Strip query strings when not considering them, so `./bar?a` and `./bar?b`
	// map to the same key.
	if !r.opts.considerQueryString {
		if idx := strings.Index(sourcePath, "?"); idx >= 0 {
			return sourcePath[:idx]
		}
	}
	return sourcePath
}

func (r *importResolver) resolveModule(importDecl *ast.ImportDeclaration, moduleSpecifier *ast.Node) (string, bool) {
	if !r.ctx.Program().IsValid() {
		return "", false
	}

	mode := r.normalMode
	hasOverride := false
	if importDecl.ImportClause != nil && importDecl.ImportClause.AsImportClause().IsTypeOnly() && importDecl.Attributes != nil {
		mode, hasOverride = importDecl.Attributes.GetResolutionModeOverride()
	}
	if !hasOverride && !r.hasNormalMode {
		mode = r.ctx.Program().GetModeForUsageLocation(r.ctx.SourceFile, moduleSpecifier)
		r.normalMode = mode
		r.hasNormalMode = true
	}

	resolved := r.ctx.Program().GetResolvedModule(r.ctx.SourceFile, moduleSpecifier.Text(), mode)
	if resolved == nil || resolved.ResolvedFileName == "" {
		return "", false
	}
	return resolved.ResolvedFileName, true
}

func (r *importResolver) hasProblematicComments(node *ast.Node) bool {
	if !r.commentsKnown {
		r.commentsKnown = true
		r.hasComments = strings.Contains(r.sourceText, "//") || strings.Contains(r.sourceText, "/*")
	}
	if !r.hasComments {
		return false
	}
	if r.commentFactory == nil {
		r.commentFactory = &ast.NodeFactory{}
	}
	return hasProblematicComments(node, r.sourceText, r.ctx.SourceFile, r.commentFactory)
}

// getImportCategory mirrors the ESLint rule's `getImportMap` routing.
func getImportCategory(importDecl *ast.ImportDeclaration, opts ruleOptions) importCategory {
	clause := importDecl.ImportClause
	if clause == nil {
		// Side-effect-only import: `import './foo'`
		return valueImport
	}

	importClause := clause.AsImportClause()

	if !opts.preferInline && importClause.IsTypeOnly() {
		// `import type X from ...` → defaultTypesImported
		// `import type {X} from ...` → namedTypesImported
		if importClause.Name() != nil && importClause.NamedBindings == nil {
			return defaultTypeImport
		}
		return namedTypeImport
	}

	// `import { type x } from './foo'` → namedTypesImported (when not prefer-inline)
	if !opts.preferInline && hasInlineTypeSpecifiers(importClause) {
		return namedTypeImport
	}

	if importClause.NamedBindings != nil && ast.IsNamespaceImport(importClause.NamedBindings) {
		return namespaceImport
	}

	return valueImport
}

// hasInlineTypeSpecifiers checks if any import specifier has the inline `type` modifier
// (e.g., `import { type x } from './foo'`).
func hasInlineTypeSpecifiers(clause *ast.ImportClause) bool {
	if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
		return false
	}
	namedImports := clause.NamedBindings.AsNamedImports()
	if namedImports.Elements == nil {
		return false
	}
	for _, elem := range namedImports.Elements.Nodes {
		spec := elem.AsImportSpecifier()
		if spec != nil && spec.IsTypeOnly {
			return true
		}
	}
	return false
}

// hasNamedSpecifiers returns true if the import has non-default, non-namespace specifiers
// (i.e., `import { x, y }` or `import { type z }`).
func hasNamedSpecifiers(node *ast.Node) bool {
	importDecl := node.AsImportDeclaration()
	if importDecl.ImportClause == nil {
		return false
	}
	clause := importDecl.ImportClause.AsImportClause()
	if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
		return false
	}
	namedImports := clause.NamedBindings.AsNamedImports()
	return namedImports.Elements != nil && len(namedImports.Elements.Nodes) > 0
}

// getDefaultImportName returns the default import identifier name, or empty string.
func getDefaultImportName(node *ast.Node) string {
	importDecl := node.AsImportDeclaration()
	if importDecl.ImportClause == nil {
		return ""
	}
	nameNode := importDecl.ImportClause.AsImportClause().Name()
	if nameNode != nil {
		return nameNode.AsIdentifier().Text
	}
	return ""
}

// hasNamespaceImport returns true if the import uses `* as ns` binding.
func hasNamespaceImport(node *ast.Node) bool {
	importDecl := node.AsImportDeclaration()
	if importDecl.ImportClause == nil {
		return false
	}
	nb := importDecl.ImportClause.AsImportClause().NamedBindings
	return nb != nil && ast.IsNamespaceImport(nb)
}

// isTypeOnlyImport returns true for `import type ...` declarations.
func isTypeOnlyImport(node *ast.Node) bool {
	importDecl := node.AsImportDeclaration()
	if importDecl.ImportClause == nil {
		return false
	}
	return importDecl.ImportClause.AsImportClause().IsTypeOnly()
}

func makeMessage(module string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noDuplicates",
		Description: "'" + module + "' imported multiple times.",
	}
}

// checkImports reports errors for every module that has more than one import.
// The autofix (if applicable) is attached to the first import; all others get plain reports.
// Groups are reported in document order (by position of first import in each group).
func checkImports(resolver *importResolver, imports importMap) {
	// Sort by position of the first import to ensure deterministic, document-order output.
	slices.SortFunc(imports.groups, func(a, b duplicateGroup) int {
		return a.first.Pos() - b.first.Pos()
	})

	for _, g := range imports.groups {
		msg := makeMessage(g.module)

		firstSource := g.first.AsImportDeclaration().ModuleSpecifier
		resolver.ctx.ReportNodeWithDeferredFixes(firstSource, msg, func() []rule.RuleFix {
			return getFix(resolver, g.first, g.rest)
		})

		for _, node := range g.rest {
			resolver.ctx.ReportNode(node.AsImportDeclaration().ModuleSpecifier, msg)
		}
	}
}

type specifierInfo struct {
	importNode     *ast.Node
	identifiersRaw string // raw identifier text between `{` and `}`
	isEmpty        bool   // true when braces contain no actual specifiers (e.g., `import {} from ...`)
}

// getFix builds autofix operations to merge duplicate imports into the first one.
// Returns nil when autofix is not possible (comments, namespace imports, conflicting defaults).
func getFix(resolver *importResolver, first *ast.Node, rest []*ast.Node) []rule.RuleFix {
	text := resolver.sourceText
	opts := resolver.opts

	// Bail: first import has comments or is a namespace import.
	if hasNamespaceImport(first) || resolver.hasProblematicComments(first) {
		return nil
	}

	// Bail: multiple different default import names (user must choose which to keep).
	firstDefaultName := getDefaultImportName(first)
	defaultImportName := firstDefaultName
	for _, node := range rest {
		if name := getDefaultImportName(node); name != "" {
			if defaultImportName != "" && name != defaultImportName {
				return nil
			}
			defaultImportName = name
		}
	}

	// Collect named specifiers and removable side-effect/default imports in one
	// pass. Imports with comments or namespace bindings stay untouched.
	var specifiers []specifierInfo
	var unnecessaryImports []*ast.Node
	for _, node := range rest {
		if hasNamespaceImport(node) || resolver.hasProblematicComments(node) {
			continue
		}

		importDecl := node.AsImportDeclaration()
		if importDecl.ImportClause == nil {
			unnecessaryImports = append(unnecessaryImports, node)
			continue
		}
		clause := importDecl.ImportClause.AsImportClause()
		if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
			unnecessaryImports = append(unnecessaryImports, node)
			continue
		}

		openBrace, closeBrace := findBraces(node, text)
		if openBrace < 0 || closeBrace < 0 {
			namedImports := clause.NamedBindings.AsNamedImports()
			if namedImports.Elements == nil || len(namedImports.Elements.Nodes) == 0 {
				unnecessaryImports = append(unnecessaryImports, node)
			}
			continue
		}

		namedImports := clause.NamedBindings.AsNamedImports()
		specifiers = append(specifiers, specifierInfo{
			importNode:     node,
			identifiersRaw: text[openBrace+1 : closeBrace],
			isEmpty:        namedImports.Elements == nil || len(namedImports.Elements.Nodes) == 0,
		})
	}

	shouldAddDefault := firstDefaultName == "" && defaultImportName != ""
	shouldAddSpecifiers := len(specifiers) > 0
	shouldRemoveUnnecessary := len(unnecessaryImports) > 0

	if !shouldAddDefault && !shouldAddSpecifiers && !shouldRemoveUnnecessary {
		return nil
	}

	// --- Build merged specifier text, deduplicating identifiers ---

	firstOpenBrace, firstCloseBrace := findBraces(first, text)
	firstHasTrailingComma := false
	firstIsEmpty := !hasNamedSpecifiers(first)

	existingIdentifiers := make(map[string]bool)
	if firstOpenBrace >= 0 && firstCloseBrace >= 0 && !firstIsEmpty {
		for id := range strings.SplitSeq(text[firstOpenBrace+1:firstCloseBrace], ",") {
			if trimmed := ecmascript.StringTrim(id); trimmed != "" {
				existingIdentifiers[trimmed] = true
			}
		}
		trimmedInside := strings.TrimRight(text[firstOpenBrace+1:firstCloseBrace], " \t\n\r")
		firstHasTrailingComma = strings.HasSuffix(trimmedInside, ",")
	}

	// Snapshot of first import's specifiers before merge (for prefer-inline conversion).
	var firstSpecifierNames map[string]bool
	if opts.preferInline && isTypeOnlyImport(first) {
		firstSpecifierNames = make(map[string]bool, len(existingIdentifiers))
		for k := range existingIdentifiers {
			firstSpecifierNames[k] = true
		}
	}

	// Build specifiersText following ESLint's reduce pattern:
	// `needsComma` tracks whether the next segment needs a leading comma.
	var specBuf strings.Builder
	needsComma := !firstHasTrailingComma && !firstIsEmpty

	for _, spec := range specifiers {
		isTypeSpec := isTypeOnlyImport(spec.importNode)
		wroteSpecifier := false

		// Append this import's identifiers directly, deduplicating as we go.
		for id := range strings.SplitSeq(spec.identifiersRaw, ",") {
			trimmed := ecmascript.StringTrim(id)
			if trimmed == "" || existingIdentifiers[trimmed] {
				continue
			}
			existingIdentifiers[trimmed] = true

			if wroteSpecifier {
				specBuf.WriteByte(',')
			} else if needsComma && !spec.isEmpty {
				specBuf.WriteByte(',')
			}

			if opts.preferInline && isTypeSpec {
				specBuf.WriteString("type ")
				specBuf.WriteString(trimmed)
			} else {
				specBuf.WriteString(id)
			}
			wroteSpecifier = true
		}

		if !spec.isEmpty {
			needsComma = true
		}
	}

	specifiersText := specBuf.String()

	// --- Build fix operations ---

	firstDecl := first.AsImportDeclaration()
	fixCapacity := len(specifiers) + len(unnecessaryImports) + 2
	if shouldAddSpecifiers && opts.preferInline && isTypeOnlyImport(first) && firstDecl.ImportClause != nil {
		bindings := firstDecl.ImportClause.AsImportClause().NamedBindings
		if bindings != nil && bindings.Kind == ast.KindNamedImports && bindings.AsNamedImports().Elements != nil {
			fixCapacity += len(bindings.AsNamedImports().Elements.Nodes)
		}
	}
	fixes := make([]rule.RuleFix, 0, fixCapacity)
	firstTrimmedPos := scanner.SkipTrivia(text, first.Pos())
	importKeywordEnd := firstTrimmedPos + len("import")

	// prefer-inline: convert `import type {a}` → `import {type a}`.
	if shouldAddSpecifiers && opts.preferInline && isTypeOnlyImport(first) {
		// Remove the `type` keyword after `import`.
		if typeRange := findTypeKeyword(first, text); typeRange.Pos() >= 0 {
			fixes = append(fixes, rule.RuleFix{Range: typeRange, Text: ""})
		}
		// Prefix each existing specifier in the first import with `type`.
		if firstOpenBrace >= 0 && firstCloseBrace >= 0 {
			clause := firstDecl.ImportClause.AsImportClause()
			if clause.NamedBindings != nil && clause.NamedBindings.Kind == ast.KindNamedImports {
				for _, elem := range clause.NamedBindings.AsNamedImports().Elements.Nodes {
					spec := elem.AsImportSpecifier()
					if spec != nil && !spec.IsTypeOnly {
						nameNode := spec.Name()
						if nameNode != nil {
							nameText := nameNode.AsIdentifier().Text
							if firstSpecifierNames[nameText] {
								trimmedNamePos := scanner.SkipTrivia(text, nameNode.Pos())
								fixes = append(fixes, rule.RuleFix{
									Range: core.NewTextRange(trimmedNamePos, nameNode.End()),
									Text:  "type " + nameText,
								})
							}
						}
					}
				}
			}
		}
	}

	// Insert specifiers / default import into the first import.
	switch {
	case shouldAddDefault && firstOpenBrace < 0 && shouldAddSpecifiers:
		// `import './foo'` → `import def, {...} from './foo'`
		fixes = append(fixes, rule.RuleFix{
			Range: core.NewTextRange(importKeywordEnd, importKeywordEnd),
			Text:  " " + defaultImportName + ", {" + specifiersText + "} from",
		})
	case shouldAddDefault && firstOpenBrace < 0 && !shouldAddSpecifiers:
		// `import './foo'` → `import def from './foo'`
		fixes = append(fixes, rule.RuleFix{
			Range: core.NewTextRange(importKeywordEnd, importKeywordEnd),
			Text:  " " + defaultImportName + " from",
		})
	case shouldAddDefault && firstOpenBrace >= 0:
		// `import {...} from './foo'` → `import def, {...} from './foo'`
		fixes = append(fixes, rule.RuleFix{
			Range: core.NewTextRange(importKeywordEnd, importKeywordEnd),
			Text:  " " + defaultImportName + ",",
		})
		if shouldAddSpecifiers {
			fixes = append(fixes, rule.RuleFix{
				Range: core.NewTextRange(firstCloseBrace, firstCloseBrace),
				Text:  specifiersText,
			})
		}
	case !shouldAddDefault && firstOpenBrace < 0 && shouldAddSpecifiers:
		if firstDecl.ImportClause != nil && firstDecl.ImportClause.AsImportClause().Name() != nil {
			// `import def from './foo'` → `import def, {...} from './foo'`
			defName := firstDecl.ImportClause.AsImportClause().Name()
			fixes = append(fixes, rule.RuleFix{
				Range: core.NewTextRange(defName.End(), defName.End()),
				Text:  ", {" + specifiersText + "}",
			})
		} else {
			// `import './foo'` → `import {...} from './foo'`
			fixes = append(fixes, rule.RuleFix{
				Range: core.NewTextRange(importKeywordEnd, importKeywordEnd),
				Text:  " {" + specifiersText + "} from",
			})
		}
	case !shouldAddDefault && firstOpenBrace >= 0 && shouldAddSpecifiers:
		// `import {...} from './foo'` → `import {..., ...} from './foo'`
		fixes = append(fixes, rule.RuleFix{
			Range: core.NewTextRange(firstCloseBrace, firstCloseBrace),
			Text:  specifiersText,
		})
	}

	// Remove merged and unnecessary imports.
	for _, spec := range specifiers {
		fixes = append(fixes, rule.RuleFix{
			Range: getRemoveRange(spec.importNode, text),
			Text:  "",
		})
	}
	for _, node := range unnecessaryImports {
		fixes = append(fixes, rule.RuleFix{
			Range: getRemoveRange(node, text),
			Text:  "",
		})
	}

	if len(fixes) == 0 {
		return nil
	}
	return fixes
}

// getRemoveRange returns the text range to delete an import node,
// including the trailing newline if present.
func getRemoveRange(node *ast.Node, text string) core.TextRange {
	trimmedPos := scanner.SkipTrivia(text, node.Pos())
	end := node.End()
	if end < len(text) && text[end] == '\n' {
		end++
	}
	return core.NewTextRange(trimmedPos, end)
}

// findBraces returns the source positions of `{` and `}` in an import's named bindings.
// Returns (-1, -1) when the import has no named bindings.
func findBraces(node *ast.Node, text string) (openBrace int, closeBrace int) {
	importDecl := node.AsImportDeclaration()
	if importDecl.ImportClause == nil {
		return -1, -1
	}
	clause := importDecl.ImportClause.AsImportClause()
	if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
		return -1, -1
	}

	namedImports := clause.NamedBindings.AsNamedImports()
	pos := namedImports.Pos()
	end := namedImports.End()

	openBrace = -1
	for i := pos; i < end; i++ {
		if text[i] == '{' {
			openBrace = i
			break
		}
	}

	closeBrace = -1
	for i := end - 1; i > pos; i-- {
		if text[i] == '}' {
			closeBrace = i
			break
		}
	}
	return
}

// findTypeKeyword locates the `type` keyword in `import type {...}` and returns
// its range including the trailing space, so removing it converts to `import {...}`.
func findTypeKeyword(node *ast.Node, text string) core.TextRange {
	trimmedPos := scanner.SkipTrivia(text, node.Pos())
	searchStart := trimmedPos + len("import")
	importDecl := node.AsImportDeclaration()
	if importDecl.ImportClause == nil {
		return core.NewTextRange(-1, -1)
	}

	searchEnd := importDecl.ImportClause.End()
	if searchEnd > len(text) {
		searchEnd = len(text)
	}

	idx := strings.Index(text[searchStart:searchEnd], "type")
	if idx < 0 {
		return core.NewTextRange(-1, -1)
	}

	typeStart := searchStart + idx
	typeEnd := typeStart + len("type")
	// Include trailing space so `import type {` becomes `import {`.
	if typeEnd < len(text) && text[typeEnd] == ' ' {
		typeEnd++
	}
	return core.NewTextRange(typeStart, typeEnd)
}

// ---------------------------------------------------------------------------
// Comment detection — autofix bails when comments make merging ambiguous.
// ---------------------------------------------------------------------------

// hasProblematicComments returns true when comments near the import make autofix risky.
// This mirrors ESLint's hasProblematicComments: it checks before, after, and inside
// the import (but outside the `{ ... }` specifier list).
func hasProblematicComments(node *ast.Node, text string, sourceFile *ast.SourceFile, factory *ast.NodeFactory) bool {
	return hasCommentBefore(node, text, sourceFile, factory) ||
		hasCommentAfter(node, text, sourceFile, factory) ||
		hasCommentInsideNonSpecifiers(node, text, sourceFile)
}

// hasCommentBefore returns true if a leading comment ends on the line before or
// the same line as the import starts.
func hasCommentBefore(node *ast.Node, text string, sourceFile *ast.SourceFile, factory *ast.NodeFactory) bool {
	lineStarts := sourceFile.ECMALineMap()
	trimmedPos := scanner.SkipTrivia(text, node.Pos())
	nodeLine := scanner.ComputeLineOfPosition(lineStarts, trimmedPos)

	for commentRange := range scanner.GetLeadingCommentRanges(factory, text, node.Pos()) {
		if scanner.ComputeLineOfPosition(lineStarts, commentRange.End()) >= nodeLine-1 {
			return true
		}
	}
	return false
}

// hasCommentAfter returns true if a trailing comment starts on the same line
// as the import ends.
func hasCommentAfter(node *ast.Node, text string, sourceFile *ast.SourceFile, factory *ast.NodeFactory) bool {
	lineStarts := sourceFile.ECMALineMap()
	nodeEndLine := scanner.ComputeLineOfPosition(lineStarts, node.End())

	for commentRange := range scanner.GetTrailingCommentRanges(factory, text, node.End()) {
		if scanner.ComputeLineOfPosition(lineStarts, commentRange.Pos()) == nodeEndLine {
			return true
		}
	}
	return false
}

// hasCommentInsideNonSpecifiers returns true if there's a comment inside the import
// statement but outside the `{ ... }` specifier list — e.g., `import/* c */{x} from './foo'`
// or `import{y}from/* c */'./foo'`.
// Uses direct text scanning because scanner-based APIs (GetLeadingCommentRanges,
// GetTrailingCommentRanges) only detect comments adjacent to line boundaries,
// missing inline comments between tokens on the same line.
func hasCommentInsideNonSpecifiers(node *ast.Node, text string, sourceFile *ast.SourceFile) bool {
	importDecl := node.AsImportDeclaration()
	if importDecl == nil || importDecl.ModuleSpecifier == nil {
		return false
	}

	trimmedPos := scanner.SkipTrivia(text, node.Pos())
	importEnd := trimmedPos + len("import")
	specStart := scanner.SkipTrivia(text, importDecl.ModuleSpecifier.Pos())

	openBrace, closeBrace := findBraces(node, text)

	// Region 1: between `import` keyword and `{` (or module specifier if no braces).
	region1End := specStart
	if openBrace >= 0 {
		region1End = openBrace + 1
	}
	if hasCommentInRegion(text, importEnd, region1End) {
		return true
	}

	// Region 2: between `}` and module specifier (only when braces exist).
	if closeBrace >= 0 && hasCommentInRegion(text, closeBrace, specStart) {
		return true
	}
	return false
}

// hasCommentInRegion checks for `//` or `/*` comment tokens in a text range.
func hasCommentInRegion(text string, start, end int) bool {
	if start < 0 || end < 0 || start >= end || start >= len(text) {
		return false
	}
	if end > len(text) {
		end = len(text)
	}
	region := text[start:end]
	return strings.Contains(region, "/*") || strings.Contains(region, "//")
}
