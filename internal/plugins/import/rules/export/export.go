package export

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
)

// entryCategory tags each export with the upstream AST type that the
// namespace-merging predicate inspects (`node.parent.type` in eslint-plugin-import).
type entryCategory int

const (
	catOther entryCategory = iota
	catNamespace
	catClass
	catEnum
	catFunction         // FunctionDeclaration with body
	catFunctionOverload // FunctionDeclaration without body (TSDeclareFunction equivalent)
)

type exportEntry struct {
	reportNode *ast.Node
	category   entryCategory
	// isOverload tracks "body-less function" regardless of whether it's a named
	// or default export — the upstream rule strips overloads before checking
	// duplicates, and default `export default function foo();` overloads need
	// the same treatment as named ones.
	isOverload bool
}

const tsTypePrefix = "type:"

// internalSymbolNamePrefix matches tsgo's `ast.InternalSymbolNamePrefix`
// (an invalid-UTF8 byte that "will never occur as IdentifierName"). The
// binder uses it to namespace synthetic symbols like the export-star
// placeholder, the missing-symbol marker, and computed-key markers — none
// of which represent user-visible exports and must be excluded from the
// duplicate-detection map.
const internalSymbolNamePrefix = "\xFE"

// See: https://github.com/import-js/eslint-plugin-import/blob/main/src/rules/export.js
var ExportRule = rule.Rule{
	Name: "import/export",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		sourceFile := ctx.SourceFile
		if sourceFile == nil || sourceFile.Statements == nil {
			return rule.RuleListeners{}
		}

		// Each scope is a flat name → []entry map; namespaces (TS modules)
		// have their own scope, so the same name in different namespaces
		// does not collide.
		var scopes []map[string][]*exportEntry
		queue := [][]*ast.Node{sourceFile.Statements.Nodes}
		for len(queue) > 0 {
			statements := queue[0]
			queue = queue[1:]

			scope := make(map[string][]*exportEntry)
			collectScopeExports(ctx, statements, scope)
			scopes = append(scopes, scope)

			for _, stmt := range statements {
				if stmt.Kind != ast.KindModuleDeclaration {
					continue
				}
				queue = append(queue, moduleBodyStatements(stmt)...)
			}
		}

		for _, scope := range scopes {
			reportScope(ctx, scope)
		}
		return rule.RuleListeners{}
	},
}

// moduleBodyStatements returns the statement lists that should be treated as
// independent scopes for a TS ModuleDeclaration. Handles `module A.B.C {}`
// (which nests as ModuleDeclaration → ModuleDeclaration → ModuleBlock) by
// recursing through the wrapper modules until a real ModuleBlock is reached.
func moduleBodyStatements(modDecl *ast.Node) [][]*ast.Node {
	body := modDecl.AsModuleDeclaration().Body
	if body == nil {
		return nil
	}
	switch body.Kind {
	case ast.KindModuleBlock:
		block := body.AsModuleBlock()
		if block.Statements == nil {
			return nil
		}
		return [][]*ast.Node{block.Statements.Nodes}
	case ast.KindModuleDeclaration:
		return moduleBodyStatements(body)
	}
	return nil
}

func collectScopeExports(ctx rule.RuleContext, statements []*ast.Node, exports map[string][]*exportEntry) {
	addType := func(name string, entry *exportEntry, isType bool) {
		if name == "" {
			return
		}
		key := name
		if isType {
			key = tsTypePrefix + name
		}
		exports[key] = append(exports[key], entry)
	}

	for _, stmt := range statements {
		switch stmt.Kind {
		case ast.KindExportAssignment:
			ea := stmt.AsExportAssignment()
			// `export = X` is a TS-only construct, semantically distinct from
			// `export default X`; upstream's listener tracks default only.
			if ea.IsExportEquals {
				continue
			}
			addType("default", &exportEntry{reportNode: stmt, category: catOther}, false)

		case ast.KindExportDeclaration:
			ed := stmt.AsExportDeclaration()
			if ed.ExportClause == nil {
				// `export * from 'mod'` — expand the upstream module's named
				// exports against the current scope. `export {} from 'mod'`
				// has ExportClause = NamedExports with empty Elements; only
				// truly clause-less form (i.e. `export *`) lands here.
				if ed.ModuleSpecifier != nil {
					handleExportStar(ctx, ctx.SourceFile, stmt, ed.ModuleSpecifier, addType)
				}
				continue
			}
			// NOTE: upstream's ExportSpecifier listener calls
			// `addNamed(name, node, parent)` without an `isType` argument, so
			// type-only re-exports (`export type { Foo } from 'mod'`) live in
			// the *value* bucket alongside `export const Foo`.
			//
			// `export * as X from 'mod'` is parsed as ExportDeclaration +
			// NamespaceExport here, but in upstream's ESTree it is an
			// ExportAllDeclaration whose listener bails out early when
			// `node.exported.name` is set — neither expanding the upstream
			// module nor adding `X` to the duplicate map. We mirror that by
			// skipping NamespaceExport entirely (no entry added, no
			// `export *` expansion).
			switch ed.ExportClause.Kind {
			case ast.KindNamespaceExport:
				// no-op (matches upstream's early-return for `export * as X`)
			case ast.KindNamedExports:
				namedExports := ed.ExportClause.AsNamedExports()
				if namedExports.Elements == nil {
					continue
				}
				for _, spec := range namedExports.Elements.Nodes {
					s := spec.AsExportSpecifier()
					nameNode := s.Name()
					name, ok := exportedNameOf(nameNode)
					if !ok {
						continue
					}
					addType(name, &exportEntry{reportNode: nameNode, category: catOther}, false)
				}
			}

		case ast.KindFunctionDeclaration:
			if !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
				continue
			}
			fn := stmt.AsFunctionDeclaration()
			overload := fn.Body == nil
			if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsDefault) {
				addType("default", &exportEntry{reportNode: stmt, category: catOther, isOverload: overload}, false)
			} else if fn.Name() != nil && fn.Name().Kind == ast.KindIdentifier {
				cat := catFunction
				if overload {
					cat = catFunctionOverload
				}
				addType(fn.Name().AsIdentifier().Text, &exportEntry{reportNode: fn.Name(), category: cat, isOverload: overload}, false)
			}

		case ast.KindClassDeclaration:
			if !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
				continue
			}
			cd := stmt.AsClassDeclaration()
			if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsDefault) {
				addType("default", &exportEntry{reportNode: stmt, category: catOther}, false)
			} else if cd.Name() != nil && cd.Name().Kind == ast.KindIdentifier {
				addType(cd.Name().AsIdentifier().Text, &exportEntry{reportNode: cd.Name(), category: catClass}, false)
			}

		case ast.KindEnumDeclaration:
			if !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
				continue
			}
			ed := stmt.AsEnumDeclaration()
			if ed.Name() != nil && ed.Name().Kind == ast.KindIdentifier {
				addType(ed.Name().AsIdentifier().Text, &exportEntry{reportNode: ed.Name(), category: catEnum}, false)
			}

		case ast.KindInterfaceDeclaration:
			if !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
				continue
			}
			id := stmt.AsInterfaceDeclaration()
			if id.Name() != nil && id.Name().Kind == ast.KindIdentifier {
				addType(id.Name().AsIdentifier().Text, &exportEntry{reportNode: id.Name(), category: catOther}, true)
			}

		case ast.KindTypeAliasDeclaration:
			if !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
				continue
			}
			td := stmt.AsTypeAliasDeclaration()
			if td.Name() != nil && td.Name().Kind == ast.KindIdentifier {
				addType(td.Name().AsIdentifier().Text, &exportEntry{reportNode: td.Name(), category: catOther}, true)
			}

		case ast.KindModuleDeclaration:
			if !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
				continue
			}
			md := stmt.AsModuleDeclaration()
			// Ambient string-named modules can't be exported by name; only
			// identifier-named namespaces participate in the duplicate check.
			if md.Name() != nil && md.Name().Kind == ast.KindIdentifier {
				addType(md.Name().AsIdentifier().Text, &exportEntry{reportNode: md.Name(), category: catNamespace}, false)
			}

		case ast.KindVariableStatement:
			if !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
				continue
			}
			vs := stmt.AsVariableStatement()
			if vs.DeclarationList == nil {
				continue
			}
			declList := vs.DeclarationList.AsVariableDeclarationList()
			if declList.Declarations == nil {
				continue
			}
			for _, decl := range declList.Declarations.Nodes {
				if decl.Kind != ast.KindVariableDeclaration {
					continue
				}
				vd := decl.AsVariableDeclaration()
				if vd.Name() == nil {
					continue
				}
				rslintUtils.CollectBindingNames(vd.Name(), func(ident *ast.Node, name string) {
					addType(name, &exportEntry{reportNode: ident, category: catOther}, false)
				})
			}
		}
	}
}

// handleExportStar resolves an `export * from 'mod'` declaration. The flow:
//
//  1. Resolve the upstream module; bail silently when unresolvable (mirrors
//     upstream's `if (remoteExports == null) { return; }`).
//  2. If the upstream module has parse errors, emit a `Parse errors in
//     imported module 'X': …` diagnostic and return without expanding.
//  3. Add every named export to the current scope as a fresh entry pointing
//     at this `export *` statement — upstream calls `addNamed(name, node)`
//     unconditionally on every ExportAllDeclaration, so two `export * from
//     './mod'` statements legitimately produce two entries per name and a
//     duplicate is reported.
//  4. If zero named exports, emit "No named exports found in module 'X'.".
func handleExportStar(
	ctx rule.RuleContext,
	owner *ast.SourceFile,
	stmt *ast.Node,
	moduleSpecifier *ast.Node,
	addType func(name string, entry *exportEntry, isType bool),
) {
	upstream, ok := lookupModule(ctx, owner, moduleSpecifier)
	if !ok {
		return
	}
	moduleName := rslintUtils.GetStaticStringValue(moduleSpecifier)
	// Parse-error surface mirrors espree-error iteration in upstream's
	// ExportMapBuilder. tsgo splits parse-time checks across syntactic +
	// binder + a portion of the checker (e.g. error 1108 'return outside
	// function' is a checker pass), so we union all three. Triggered through
	// program APIs to ensure they are computed lazily on demand.
	bgCtx := context.Background()
	parseErrs := append(append(append([]*ast.Diagnostic(nil),
		ctx.Program.GetSyntacticDiagnostics(bgCtx, upstream)...),
		ctx.Program.GetBindDiagnostics(bgCtx, upstream)...),
		filterSyntacticChecker(ctx.Program.GetSemanticDiagnostics(bgCtx, upstream))...)
	if len(parseErrs) > 0 {
		ctx.ReportNode(moduleSpecifier, rule.RuleMessage{
			Id:          "parseErrors",
			Description: fmt.Sprintf("Parse errors in imported module '%s': %s", moduleName, formatParseErrors(upstream, parseErrs)),
		})
		return
	}
	names := fileNamedExports(ctx, upstream, map[string]bool{ctx.SourceFile.FileName(): true})
	if len(names) == 0 {
		ctx.ReportNode(moduleSpecifier, rule.RuleMessage{
			Id:          "noNamedExports",
			Description: fmt.Sprintf("No named exports found in module '%s'.", moduleName),
		})
		return
	}
	for _, n := range names {
		addType(n, &exportEntry{reportNode: stmt, category: catOther}, false)
	}
}

// syntacticCheckerErrorCodes lists checker error codes that fire on syntactic
// invariants (so they are conceptually part of "parse errors" from the
// import-plugin perspective) rather than on type-level mistakes. Keep this
// list minimal — type errors should NOT surface as parse errors.
var syntacticCheckerErrorCodes = map[int32]bool{
	1108: true, // A 'return' statement can only be used within a function body.
	1109: true, // Expression expected.
	1124: true, // Digit expected.
	1125: true, // Hexadecimal digit expected.
	1126: true, // Unexpected end of text.
	1127: true, // Invalid character.
	1003: true, // Identifier expected.
	1005: true, // ',' expected. / ';' expected. / '}' expected.
	1014: true, // A rest parameter must be last in a parameter list.
	1015: true, // Parameter cannot have question mark and initializer.
	1016: true, // A required parameter cannot follow an optional parameter.
}

func filterSyntacticChecker(diags []*ast.Diagnostic) []*ast.Diagnostic {
	out := diags[:0:0]
	for _, d := range diags {
		if syntacticCheckerErrorCodes[d.Code()] {
			out = append(out, d)
		}
	}
	return out
}

// formatParseErrors joins each parse-error diagnostic as `<message> (line:col)`
// with a ", " separator, mirroring upstream's
// `errors.map(e => ${e.message} (${e.lineNumber}:${e.column})).join(', ')`.
// Lines and columns are 1-based.
func formatParseErrors(sf *ast.SourceFile, diags []*ast.Diagnostic) string {
	parts := make([]string, 0, len(diags))
	for _, d := range diags {
		line, col := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, d.Pos())
		parts = append(parts, fmt.Sprintf("%s (%d:%d)", d.String(), line+1, col+1))
	}
	return strings.Join(parts, ", ")
}

// lookupModule resolves a module specifier (string literal) to its source
// file. The owner argument identifies the file that *contains* the specifier
// — required for module resolution when the specifier sits inside an
// upstream module reached through `fileNamedExports` recursion (the rule's
// own SourceFile is irrelevant in that case).
func lookupModule(ctx rule.RuleContext, owner *ast.SourceFile, moduleSpecifier *ast.Node) (*ast.SourceFile, bool) {
	if moduleSpecifier == nil || !ast.IsStringLiteralLike(moduleSpecifier) {
		return nil, false
	}
	if owner == nil {
		owner = ctx.SourceFile
	}
	module := ctx.Program.GetResolvedModuleFromModuleSpecifier(owner, moduleSpecifier)
	if module == nil || module.ResolvedFileName == "" {
		return nil, false
	}
	if sf := ctx.Program.GetSourceFile(module.ResolvedFileName); sf != nil {
		return sf, true
	}
	return nil, false
}

// fileNamedExports returns every named export of a source file. The
// syntactic walk handles ES `export X` forms and recurses through
// `export * from` chains (matching upstream's ExportMapBuilder behavior).
// Files written in CommonJS style (`exports.X = ...`, `module.exports.Y =`)
// expose nothing through ES syntax, so we fall back to the binder-resolved
// module symbol's exports table when the syntactic walk yields zero names.
//
// `default` is filtered out per upstream's `if (name !== 'default')` filter.
func fileNamedExports(ctx rule.RuleContext, sf *ast.SourceFile, visited map[string]bool) []string {
	if sf == nil {
		return nil
	}
	fname := sf.FileName()
	if visited[fname] {
		return nil
	}
	visited[fname] = true

	if sf.Statements == nil {
		return nil
	}
	seen := make(map[string]bool)
	var names []string
	add := func(name string) {
		if name == "" || name == "default" || seen[name] {
			return
		}
		seen[name] = true
		names = append(names, name)
	}

	for _, stmt := range sf.Statements.Nodes {
		switch stmt.Kind {
		case ast.KindExportDeclaration:
			ed := stmt.AsExportDeclaration()
			if ed.ExportClause == nil {
				if ed.ModuleSpecifier != nil {
					if up, ok := lookupModule(ctx, sf, ed.ModuleSpecifier); ok {
						for _, n := range fileNamedExports(ctx, up, visited) {
							add(n)
						}
					}
				}
				continue
			}
			switch ed.ExportClause.Kind {
			case ast.KindNamespaceExport:
				if name, ok := exportedNameOf(ed.ExportClause.AsNamespaceExport().Name()); ok {
					add(name)
				}
			case ast.KindNamedExports:
				ne := ed.ExportClause.AsNamedExports()
				if ne.Elements == nil {
					continue
				}
				for _, spec := range ne.Elements.Nodes {
					s := spec.AsExportSpecifier()
					if name, ok := exportedNameOf(s.Name()); ok {
						add(name)
					}
				}
			}
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration, ast.KindEnumDeclaration,
			ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindModuleDeclaration:
			if !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
				continue
			}
			if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsDefault) {
				continue
			}
			if name := declarationName(stmt); name != "" {
				add(name)
			}
		case ast.KindVariableStatement:
			if !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
				continue
			}
			vs := stmt.AsVariableStatement()
			if vs.DeclarationList == nil {
				continue
			}
			declList := vs.DeclarationList.AsVariableDeclarationList()
			if declList.Declarations == nil {
				continue
			}
			for _, decl := range declList.Declarations.Nodes {
				if decl.Kind != ast.KindVariableDeclaration {
					continue
				}
				vd := decl.AsVariableDeclaration()
				if vd.Name() != nil {
					rslintUtils.CollectBindingNames(vd.Name(), func(_ *ast.Node, n string) {
						add(n)
					})
				}
			}
		}
	}
	// Always merge in the binder-resolved module symbol's exports. This
	// covers two cases the syntactic walk misses:
	//   1. Pure CommonJS (`exports.X = ...`) modules — no ES syntax surface.
	//   2. Hybrid ES + CommonJS modules — ES surface only exposes ES names,
	//      so a sibling CJS-style `exports.Y = ...` would slip through.
	// `add` deduplicates against the names already collected via syntax, so
	// re-running over ES names is harmless. The binder folds both ES and
	// CommonJS exports into a single Symbol.Exports table, matching upstream
	// import-plugin's babel-parser view.
	if symbol := sf.AsNode().Symbol(); symbol != nil && symbol.Exports != nil {
		for name := range symbol.Exports {
			if strings.HasPrefix(name, internalSymbolNamePrefix) {
				continue
			}
			add(name)
		}
	}
	return names
}

// declarationName extracts the identifier text of a top-level declaration,
// returning "" when the declaration is anonymous or its name is not a plain
// identifier.
func declarationName(stmt *ast.Node) string {
	switch stmt.Kind {
	case ast.KindFunctionDeclaration:
		fn := stmt.AsFunctionDeclaration()
		if fn.Name() != nil && fn.Name().Kind == ast.KindIdentifier {
			return fn.Name().AsIdentifier().Text
		}
	case ast.KindClassDeclaration:
		cd := stmt.AsClassDeclaration()
		if cd.Name() != nil && cd.Name().Kind == ast.KindIdentifier {
			return cd.Name().AsIdentifier().Text
		}
	case ast.KindEnumDeclaration:
		ed := stmt.AsEnumDeclaration()
		if ed.Name() != nil && ed.Name().Kind == ast.KindIdentifier {
			return ed.Name().AsIdentifier().Text
		}
	case ast.KindInterfaceDeclaration:
		id := stmt.AsInterfaceDeclaration()
		if id.Name() != nil && id.Name().Kind == ast.KindIdentifier {
			return id.Name().AsIdentifier().Text
		}
	case ast.KindTypeAliasDeclaration:
		td := stmt.AsTypeAliasDeclaration()
		if td.Name() != nil && td.Name().Kind == ast.KindIdentifier {
			return td.Name().AsIdentifier().Text
		}
	case ast.KindModuleDeclaration:
		md := stmt.AsModuleDeclaration()
		if md.Name() != nil && md.Name().Kind == ast.KindIdentifier {
			return md.Name().AsIdentifier().Text
		}
	}
	return ""
}

// exportedNameOf returns the textual name of an identifier or string-literal
// export name (the latter is the ES2022 arbitrary-module-namespace form).
func exportedNameOf(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	if node.Kind == ast.KindIdentifier {
		return node.AsIdentifier().Text, true
	}
	if s := rslintUtils.GetStaticStringValue(node); s != "" {
		return s, true
	}
	return "", false
}

func reportScope(ctx rule.RuleContext, scope map[string][]*exportEntry) {
	// Iterate names in deterministic order (by first entry's position) so the
	// diagnostics list matches the source order.
	type group struct {
		name    string
		entries []*exportEntry
	}
	var groups []group
	for name, entries := range scope {
		groups = append(groups, group{name: name, entries: entries})
	}
	sort.SliceStable(groups, func(i, j int) bool {
		return groups[i].entries[0].reportNode.Pos() < groups[j].entries[0].reportNode.Pos()
	})

	for _, g := range groups {
		entries := stripOverloads(g.entries)
		if len(entries) <= 1 {
			continue
		}
		if isNamespaceMerging(entries) {
			continue
		}
		// Sort entries by position so each report fires in source order.
		sort.SliceStable(entries, func(i, j int) bool {
			return entries[i].reportNode.Pos() < entries[j].reportNode.Pos()
		})
		for _, e := range entries {
			if shouldSkipNamespace(e, entries) {
				continue
			}
			ctx.ReportNode(e.reportNode, makeMessage(g.name))
		}
	}
}

func stripOverloads(entries []*exportEntry) []*exportEntry {
	out := make([]*exportEntry, 0, len(entries))
	for _, e := range entries {
		if e.isOverload {
			continue
		}
		out = append(out, e)
	}
	return out
}

// isNamespaceMerging mirrors upstream's predicate: a TS namespace can merge
// with itself, with a single class or enum, or with any number of function
// overloads + at most one impl.
func isNamespaceMerging(entries []*exportEntry) bool {
	types := make(map[entryCategory]bool)
	nonNs := 0
	for _, e := range entries {
		types[e.category] = true
		if e.category != catNamespace {
			nonNs++
		}
	}
	if !types[catNamespace] {
		return false
	}
	switch len(types) {
	case 1:
		return true
	case 2:
		if types[catFunction] || types[catFunctionOverload] {
			return true
		}
		if (types[catClass] || types[catEnum]) && nonNs == 1 {
			return true
		}
	case 3:
		if types[catFunction] && types[catFunctionOverload] {
			return true
		}
	}
	return false
}

// shouldSkipNamespace decides whether to silence the diagnostic on a namespace
// entry when the duplicate is driven by classes/enums/functions instead — the
// user gets a cleaner message focused on the value side.
func shouldSkipNamespace(e *exportEntry, entries []*exportEntry) bool {
	if e.category != catNamespace {
		return false
	}
	if isNamespaceMerging(entries) {
		return false
	}
	for _, other := range entries {
		switch other.category {
		case catEnum, catClass, catFunction, catFunctionOverload:
			return true
		}
	}
	return false
}

func makeMessage(name string) rule.RuleMessage {
	if name == "default" {
		return rule.RuleMessage{
			Id:          "multipleDefault",
			Description: "Multiple default exports.",
		}
	}
	displayName := strings.TrimPrefix(name, tsTypePrefix)
	return rule.RuleMessage{
		Id:          "multipleNamed",
		Description: fmt.Sprintf("Multiple exports of name '%s'.", displayName),
	}
}
