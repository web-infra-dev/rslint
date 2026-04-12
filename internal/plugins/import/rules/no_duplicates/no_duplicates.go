package no_duplicates

import (
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
)

type ruleOptions struct {
	considerQueryString bool
	preferInline        bool
}

func parseOptions(options any) ruleOptions {
	opts := ruleOptions{}
	optsMap := rslintUtils.GetOptionsMap(options)
	if optsMap != nil {
		if v, ok := optsMap["considerQueryString"]; ok {
			if b, ok := v.(bool); ok {
				opts.considerQueryString = b
			}
		}
		if v, ok := optsMap["prefer-inline"]; ok {
			if b, ok := v.(bool); ok {
				opts.preferInline = b
			}
		}
	}
	return opts
}

// See: https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/no-duplicates.md
var NoDuplicatesRule = rule.Rule{
	Name: "import/no-duplicates",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		sourceFile := ctx.SourceFile
		if sourceFile == nil || sourceFile.Statements == nil {
			return rule.RuleListeners{}
		}

		sourceText := sourceFile.Text()

		// ESLint groups imports by parent scope (n.parent), so imports inside
		// `declare module` blocks are checked independently from top-level imports.
		// We replicate this by processing each scope's statements separately.
		processScope(ctx, sourceFile.Statements.Nodes, opts, sourceText)

		// Recursively process imports inside `declare module` / `declare namespace` blocks.
		walkModuleDeclarations(ctx, sourceFile.Statements.Nodes, opts, sourceText)

		return rule.RuleListeners{}
	},
}

// walkModuleDeclarations recursively finds ModuleDeclaration nodes and processes
// their body statements as independent scopes.
func walkModuleDeclarations(ctx rule.RuleContext, statements []*ast.Node, opts ruleOptions, sourceText string) {
	for _, stmt := range statements {
		if stmt.Kind != ast.KindModuleDeclaration {
			continue
		}
		body := stmt.AsModuleDeclaration().Body
		if body == nil {
			continue
		}
		// ModuleDeclaration body can be a ModuleBlock or another ModuleDeclaration (nested namespaces).
		switch body.Kind {
		case ast.KindModuleBlock:
			blockStatements := body.AsModuleBlock().Statements
			if blockStatements != nil && len(blockStatements.Nodes) > 0 {
				processScope(ctx, blockStatements.Nodes, opts, sourceText)
				walkModuleDeclarations(ctx, blockStatements.Nodes, opts, sourceText)
			}
		case ast.KindModuleDeclaration:
			// `declare module A.B { ... }` nests as ModuleDeclaration → ModuleDeclaration → ModuleBlock
			walkModuleDeclarations(ctx, []*ast.Node{body}, opts, sourceText)
		}
	}
}

// processScope collects and checks duplicate imports within a single scope (statement list).
func processScope(ctx rule.RuleContext, statements []*ast.Node, opts ruleOptions, sourceText string) {
	// Four maps mirror the ESLint rule's import categorization.
	imported := make(map[string][]*ast.Node)
	nsImported := make(map[string][]*ast.Node)
	defaultTypesImported := make(map[string][]*ast.Node)
	namedTypesImported := make(map[string][]*ast.Node)

	for _, stmt := range statements {
		if stmt.Kind != ast.KindImportDeclaration {
			continue
		}

		importDecl := stmt.AsImportDeclaration()
		if importDecl.ModuleSpecifier == nil {
			continue
		}

		resolvedPath := resolveImportPath(importDecl.ModuleSpecifier, ctx, opts, sourceText)
		if resolvedPath == "" {
			continue
		}

		importMap := getImportMap(importDecl, opts, imported, nsImported, defaultTypesImported, namedTypesImported)
		importMap[resolvedPath] = append(importMap[resolvedPath], stmt)
	}

	checkImports(ctx, imported, sourceText, opts)
	checkImports(ctx, nsImported, sourceText, opts)
	checkImports(ctx, defaultTypesImported, sourceText, opts)
	checkImports(ctx, namedTypesImported, sourceText, opts)
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

// resolveImportPath resolves the module specifier to a canonical path for grouping.
// When considerQueryString is true, query strings are preserved so that
// `./mod?a` and `./mod?b` are treated as different modules.
// When false (default), query strings are stripped for comparison.
func resolveImportPath(moduleSpecifier *ast.Node, ctx rule.RuleContext, opts ruleOptions, sourceText string) string {
	if moduleSpecifier == nil || !ast.IsStringLiteralLike(moduleSpecifier) {
		return ""
	}

	sourcePath := getModuleSpecifierText(moduleSpecifier, sourceText)

	if opts.considerQueryString {
		idx := strings.Index(sourcePath, "?")
		if idx >= 0 {
			query := sourcePath[idx:]
			if resolved, ok := utils.Resolve(moduleSpecifier, ctx); ok && resolved != "" {
				return resolved + query
			}
			return sourcePath[:idx] + query
		}
	}

	if resolved, ok := utils.Resolve(moduleSpecifier, ctx); ok && resolved != "" {
		return resolved
	}

	// Strip query strings when not considering them, so `./bar?a` and `./bar?b`
	// map to the same key.
	if !opts.considerQueryString {
		if idx := strings.Index(sourcePath, "?"); idx >= 0 {
			return sourcePath[:idx]
		}
	}
	return sourcePath
}

// getImportMap determines which import map an ImportDeclaration should be routed to,
// mirroring the ESLint rule's `getImportMap` function.
func getImportMap(
	importDecl *ast.ImportDeclaration,
	opts ruleOptions,
	imported, nsImported, defaultTypesImported, namedTypesImported map[string][]*ast.Node,
) map[string][]*ast.Node {
	clause := importDecl.ImportClause
	if clause == nil {
		// Side-effect-only import: `import './foo'`
		return imported
	}

	importClause := clause.AsImportClause()

	if !opts.preferInline && importClause.IsTypeOnly() {
		// `import type X from ...` → defaultTypesImported
		// `import type {X} from ...` → namedTypesImported
		if importClause.Name() != nil && importClause.NamedBindings == nil {
			return defaultTypesImported
		}
		return namedTypesImported
	}

	// `import { type x } from './foo'` → namedTypesImported (when not prefer-inline)
	if !opts.preferInline && hasInlineTypeSpecifiers(importClause) {
		return namedTypesImported
	}

	if importClause.NamedBindings != nil && ast.IsNamespaceImport(importClause.NamedBindings) {
		return nsImported
	}

	return imported
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
		Description: fmt.Sprintf("'%s' imported multiple times.", module),
	}
}

// checkImports reports errors for every module that has more than one import.
// The autofix (if applicable) is attached to the first import; all others get plain reports.
// Groups are reported in document order (by position of first import in each group).
func checkImports(ctx rule.RuleContext, importMap map[string][]*ast.Node, text string, opts ruleOptions) {
	// Collect groups that have duplicates.
	type group struct {
		module string
		nodes  []*ast.Node
	}
	var groups []group
	for module, nodes := range importMap {
		if len(nodes) > 1 {
			groups = append(groups, group{module, nodes})
		}
	}

	// Sort by position of the first import to ensure deterministic, document-order output.
	slices.SortFunc(groups, func(a, b group) int {
		return a.nodes[0].Pos() - b.nodes[0].Pos()
	})

	for _, g := range groups {
		msg := makeMessage(g.module)
		first := g.nodes[0]
		rest := g.nodes[1:]

		fixes := getFix(ctx, first, rest, text, opts)

		firstSource := first.AsImportDeclaration().ModuleSpecifier
		if fixes != nil {
			ctx.ReportNodeWithFixes(firstSource, msg, fixes...)
		} else {
			ctx.ReportNode(firstSource, msg)
		}

		for _, node := range rest {
			ctx.ReportNode(node.AsImportDeclaration().ModuleSpecifier, msg)
		}
	}
}

type specifierInfo struct {
	importNode  *ast.Node
	identifiers []string // raw identifier text segments split by ","
	isEmpty     bool     // true when braces contain no actual specifiers (e.g., `import {} from ...`)
}

// getFix builds autofix operations to merge duplicate imports into the first one.
// Returns nil when autofix is not possible (comments, namespace imports, conflicting defaults).
func getFix(ctx rule.RuleContext, first *ast.Node, rest []*ast.Node, text string, opts ruleOptions) []rule.RuleFix {
	sourceFile := ctx.SourceFile

	// Bail: first import has comments or is a namespace import.
	if hasProblematicComments(first, text, sourceFile) || hasNamespaceImport(first) {
		return nil
	}

	// Bail: multiple different default import names (user must choose which to keep).
	defaultNames := make(map[string]bool)
	if name := getDefaultImportName(first); name != "" {
		defaultNames[name] = true
	}
	for _, node := range rest {
		if name := getDefaultImportName(node); name != "" {
			defaultNames[name] = true
		}
	}
	if len(defaultNames) > 1 {
		return nil
	}

	// Skip rest nodes with comments or namespace imports — they can't be auto-merged.
	var restWithoutComments []*ast.Node
	for _, node := range rest {
		if !hasProblematicComments(node, text, sourceFile) && !hasNamespaceImport(node) {
			restWithoutComments = append(restWithoutComments, node)
		}
	}

	// Collect specifier text from each mergeable rest import that has named bindings.
	var specifiers []specifierInfo
	for _, node := range restWithoutComments {
		importDecl := node.AsImportDeclaration()
		if importDecl.ImportClause == nil {
			continue
		}
		clause := importDecl.ImportClause.AsImportClause()
		if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
			continue
		}

		openBrace, closeBrace := findBraces(node, text)
		if openBrace < 0 || closeBrace < 0 {
			continue
		}

		specifiers = append(specifiers, specifierInfo{
			importNode:  node,
			identifiers: strings.Split(text[openBrace+1:closeBrace], ","),
			isEmpty:     !hasNamedSpecifiers(node),
		})
	}

	// Unnecessary imports: no named specifiers, no namespace — pure side-effect or redundant default.
	var unnecessaryImports []*ast.Node
	for _, node := range restWithoutComments {
		if hasNamedSpecifiers(node) || hasNamespaceImport(node) {
			continue
		}
		isSpecifier := false
		for _, s := range specifiers {
			if s.importNode == node {
				isSpecifier = true
				break
			}
		}
		if !isSpecifier {
			unnecessaryImports = append(unnecessaryImports, node)
		}
	}

	shouldAddDefault := getDefaultImportName(first) == "" && len(defaultNames) == 1
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
		for _, id := range strings.Split(text[firstOpenBrace+1:firstCloseBrace], ",") {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				existingIdentifiers[trimmed] = true
			}
		}
		trimmedInside := strings.TrimRight(text[firstOpenBrace+1:firstCloseBrace], " \t\n\r")
		firstHasTrailingComma = strings.HasSuffix(trimmedInside, ",")
	}

	// Snapshot of first import's specifiers before merge (for prefer-inline conversion).
	firstSpecifierNames := make(map[string]bool, len(existingIdentifiers))
	for k := range existingIdentifiers {
		firstSpecifierNames[k] = true
	}

	// Build specifiersText following ESLint's reduce pattern:
	// `needsComma` tracks whether the next segment needs a leading comma.
	var specBuf strings.Builder
	needsComma := !firstHasTrailingComma && !firstIsEmpty

	for _, spec := range specifiers {
		isTypeSpec := isTypeOnlyImport(spec.importNode)

		// Build text for this specifier's identifiers, deduplicating.
		var specTextBuf strings.Builder
		for _, id := range spec.identifiers {
			trimmed := strings.TrimSpace(id)
			if trimmed == "" || existingIdentifiers[trimmed] {
				continue
			}
			existingIdentifiers[trimmed] = true

			curWithType := id
			if opts.preferInline && isTypeSpec {
				curWithType = "type " + trimmed
			}

			if specTextBuf.Len() > 0 {
				specTextBuf.WriteString(",")
			}
			specTextBuf.WriteString(curWithType)
		}

		specText := specTextBuf.String()
		if specText == "" {
			if !spec.isEmpty {
				needsComma = true
			}
			continue
		}

		if needsComma && !spec.isEmpty {
			specBuf.WriteString(",")
		}
		specBuf.WriteString(specText)

		if !spec.isEmpty {
			needsComma = true
		}
	}

	specifiersText := specBuf.String()

	// --- Build fix operations ---

	var fixes []rule.RuleFix
	firstDecl := first.AsImportDeclaration()
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

	// Determine the default import name to add (if any).
	var defaultImportName string
	if shouldAddDefault {
		for name := range defaultNames {
			defaultImportName = name
			break
		}
	}

	// Insert specifiers / default import into the first import.
	switch {
	case shouldAddDefault && firstOpenBrace < 0 && shouldAddSpecifiers:
		// `import './foo'` → `import def, {...} from './foo'`
		fixes = append(fixes, rule.RuleFix{
			Range: core.NewTextRange(importKeywordEnd, importKeywordEnd),
			Text:  fmt.Sprintf(" %s, {%s} from", defaultImportName, specifiersText),
		})
	case shouldAddDefault && firstOpenBrace < 0 && !shouldAddSpecifiers:
		// `import './foo'` → `import def from './foo'`
		fixes = append(fixes, rule.RuleFix{
			Range: core.NewTextRange(importKeywordEnd, importKeywordEnd),
			Text:  fmt.Sprintf(" %s from", defaultImportName),
		})
	case shouldAddDefault && firstOpenBrace >= 0:
		// `import {...} from './foo'` → `import def, {...} from './foo'`
		fixes = append(fixes, rule.RuleFix{
			Range: core.NewTextRange(importKeywordEnd, importKeywordEnd),
			Text:  fmt.Sprintf(" %s,", defaultImportName),
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
				Text:  fmt.Sprintf(", {%s}", specifiersText),
			})
		} else {
			// `import './foo'` → `import {...} from './foo'`
			fixes = append(fixes, rule.RuleFix{
				Range: core.NewTextRange(importKeywordEnd, importKeywordEnd),
				Text:  fmt.Sprintf(" {%s} from", specifiersText),
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
func hasProblematicComments(node *ast.Node, text string, sourceFile *ast.SourceFile) bool {
	return hasCommentBefore(node, text, sourceFile) ||
		hasCommentAfter(node, text, sourceFile) ||
		hasCommentInsideNonSpecifiers(node, text, sourceFile)
}

// hasCommentBefore returns true if a leading comment ends on the line before or
// the same line as the import starts.
func hasCommentBefore(node *ast.Node, text string, sourceFile *ast.SourceFile) bool {
	lineStarts := sourceFile.ECMALineMap()
	trimmedPos := scanner.SkipTrivia(text, node.Pos())
	nodeLine := scanner.ComputeLineOfPosition(lineStarts, trimmedPos)

	nodeFactory := &ast.NodeFactory{}
	for commentRange := range scanner.GetLeadingCommentRanges(nodeFactory, text, node.Pos()) {
		if scanner.ComputeLineOfPosition(lineStarts, commentRange.End()) >= nodeLine-1 {
			return true
		}
	}
	return false
}

// hasCommentAfter returns true if a trailing comment starts on the same line
// as the import ends.
func hasCommentAfter(node *ast.Node, text string, sourceFile *ast.SourceFile) bool {
	lineStarts := sourceFile.ECMALineMap()
	nodeEndLine := scanner.ComputeLineOfPosition(lineStarts, node.End())

	nodeFactory := &ast.NodeFactory{}
	for commentRange := range scanner.GetTrailingCommentRanges(nodeFactory, text, node.End()) {
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
