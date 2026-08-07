package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

const defaultExportName = "default"

type ExportMeta struct {
	Namespace *ExportMap
}

// ExportMap records the statically visible exports of an ES module. It does
// not synthesize default exports from compiler interop settings; HasExport
// keeps that looser existence check for rules that need it.
type ExportMap struct {
	exports    map[string]*ExportMeta
	hasUnknown bool
}

func NewExportMap() *ExportMap {
	return &ExportMap{
		exports: make(map[string]*ExportMeta),
	}
}

func (m *ExportMap) Size() int {
	if m == nil {
		return 0
	}
	size := len(m.exports)
	if m.hasUnknown {
		size++
	}
	return size
}

func (m *ExportMap) Has(name string) bool {
	if m == nil {
		return false
	}
	if _, ok := m.exports[name]; ok {
		return true
	}
	return m.hasUnknown
}

func (m *ExportMap) Get(name string) *ExportMeta {
	if m == nil {
		return nil
	}
	return m.exports[name]
}

func (m *ExportMap) Set(name string, meta *ExportMeta) {
	if m == nil || name == "" {
		return
	}
	if meta == nil {
		meta = &ExportMeta{}
	}
	m.exports[name] = meta
}

func (m *ExportMap) AddUnknown() {
	if m != nil {
		m.hasUnknown = true
	}
}

func (m *ExportMap) MergeFrom(other *ExportMap, includeDefault bool) {
	if m == nil || other == nil {
		return
	}
	for name, meta := range other.exports {
		if !includeDefault && name == defaultExportName {
			continue
		}
		m.Set(name, meta)
	}
	if other.hasUnknown {
		m.AddUnknown()
	}
}

// GetExportMap resolves moduleSpecifier from ctx.SourceFile and returns a
// recursive export map. The second result is false when no export map is
// available, matching eslint-plugin-import's "imports == null" branch.
func GetExportMap(ctx rule.RuleContext, moduleSpecifier *ast.Node) (*ExportMap, bool) {
	if ctx.SourceFile == nil {
		return nil, false
	}
	return getExportMap(ctx, ctx.SourceFile, moduleSpecifier, newExportBuilder(ctx))
}

// HasDefaultExport resolves moduleSpecifier from ctx.SourceFile and reports
// whether the resolved module has a statically visible default export. The
// second result is false when no export map is available, matching
// eslint-plugin-import's "imports == null" branch.
func HasDefaultExport(ctx rule.RuleContext, moduleSpecifier *ast.Node) (bool, bool) {
	return HasExport(ctx, moduleSpecifier, defaultExportName)
}

// HasExport resolves moduleSpecifier from ctx.SourceFile and reports whether
// the resolved module statically exports exportName. The second result is false
// when the target is unresolved or is not an ES module.
func HasExport(ctx rule.RuleContext, moduleSpecifier *ast.Node, exportName string) (bool, bool) {
	if ctx.Program == nil || ctx.SourceFile == nil || moduleSpecifier == nil || !ast.IsStringLiteralLike(moduleSpecifier) {
		return false, false
	}
	return hasExport(ctx, ctx.SourceFile, moduleSpecifier, exportName, newExportBuilder(ctx))
}

// exportBuilder carries one query's traversal state over the per-file export
// structures the index holds. building maps a file to the map being filled
// for it, so a module that re-exports its way back to itself sees the partial
// map rather than recursing forever; seen does the same for the name lookup.
//
// It also decides which of the maps it builds outlive the query. A map is
// only a property of its file when nothing in the file's re-export closure
// forms a cycle: inside a cycle, what a module ends up exporting depends on
// which module the query entered from, so those maps stay private to the
// query that built them.
type exportBuilder struct {
	ctx      rule.RuleContext
	index    *ModuleIndex
	building map[*ast.SourceFile]*ExportMap
	seen     map[exportKey]bool
	// onStack holds the files whose maps are still being filled, so that
	// reaching one again is recognized as a cycle rather than as reuse.
	onStack map[*ast.SourceFile]bool
	// unstable holds the files this query finished but found to be in or
	// below a cycle, so that reusing one keeps its dependents unstable too.
	unstable map[*ast.SourceFile]bool
	// sawCycle reports whether the file currently being built has reached a
	// cycle so far. It is saved and restored around each nested build.
	sawCycle bool
}

func newExportBuilder(ctx rule.RuleContext) *exportBuilder {
	return &exportBuilder{
		ctx:      ctx,
		index:    IndexFor(ctx),
		building: make(map[*ast.SourceFile]*ExportMap),
		seen:     make(map[exportKey]bool),
		onStack:  make(map[*ast.SourceFile]bool),
		unstable: make(map[*ast.SourceFile]bool),
	}
}

type exportKey struct {
	file *ast.SourceFile
	name string
}

func hasExport(ctx rule.RuleContext, origin *ast.SourceFile, moduleSpecifier *ast.Node, exportName string, builder *exportBuilder) (bool, bool) {
	link := resolveExportLinkForLookup(ctx, origin, builder.index.Settings(), moduleSpecifier)
	if link.Target == nil {
		return false, false
	}
	return sourceFileHasExport(ctx, link.Target, exportName, builder)
}

// resolveExportLinkForLookup is the name-lookup counterpart of
// resolveExportLink: it stops before the is-an-ES-module test, which
// sourceFileHasExport applies itself.
func resolveExportLinkForLookup(ctx rule.RuleContext, origin *ast.SourceFile, settings *ModuleSettings, moduleSpecifier *ast.Node) exportLink {
	_, sourceFile, ok := rslint_utils.ResolveModuleFile(ctx.Program, origin, moduleSpecifier)
	if !ok || settings.IsIgnoredPath(sourceFile.FileName()) {
		return exportLink{}
	}
	return exportLink{Target: sourceFile, Resolved: true}
}

func getExportMap(ctx rule.RuleContext, origin *ast.SourceFile, moduleSpecifier *ast.Node, builder *exportBuilder) (*ExportMap, bool) {
	if ctx.Program == nil || origin == nil || moduleSpecifier == nil || !ast.IsStringLiteralLike(moduleSpecifier) {
		return nil, false
	}

	link := resolveExportLink(ctx, origin, builder.index.Settings(), moduleSpecifier)
	if !link.Resolved {
		return nil, false
	}
	return builder.exportMapOf(link.Target), true
}

// exportMapOf applies a file's export steps in source order. The map is
// registered before the steps run, so a dependency that re-exports its way
// back here merges whatever is complete at that point — which is what
// re-walking the statements would also produce.
//
// A file whose closure turned out to be free of cycles is handed to the index
// afterwards, because its map cannot depend on where the query started: a
// dependency can only be half-built when it is also waiting on this file,
// which is exactly the case a cycle-free closure rules out.
func (builder *exportBuilder) exportMapOf(sourceFile *ast.SourceFile) *ExportMap {
	if sourceFile == nil {
		return NewExportMap()
	}
	if existing := builder.building[sourceFile]; existing != nil {
		if builder.onStack[sourceFile] || builder.unstable[sourceFile] {
			builder.sawCycle = true
		}
		return existing
	}
	if cached := builder.index.cachedExportMap(sourceFile); cached != nil {
		return cached
	}

	exports := NewExportMap()
	builder.building[sourceFile] = exports
	builder.onStack[sourceFile] = true
	enclosing := builder.sawCycle
	builder.sawCycle = false

	local := builder.index.localExportsOf(builder.ctx, sourceFile)
	for _, step := range local.Steps {
		builder.applyStep(exports, local, step)
	}

	delete(builder.onStack, sourceFile)
	if builder.sawCycle {
		builder.unstable[sourceFile] = true
	} else {
		builder.index.storeExportMap(sourceFile, exports)
	}
	builder.sawCycle = enclosing || builder.sawCycle
	return exports
}

func (builder *exportBuilder) applyStep(exports *ExportMap, local *localExports, step exportStep) {
	switch step.Kind {
	case exportStepNames:
		for _, name := range step.Names {
			exports.Set(name, nil)
		}

	case exportStepLocalDefault:
		if meta, ok := builder.namespaceImportMeta(local, step.Local); ok {
			exports.Set(defaultExportName, meta)
		} else {
			exports.Set(defaultExportName, nil)
		}

	case exportStepStar:
		if !step.Link.Resolved {
			exports.AddUnknown()
			return
		}
		exports.MergeFrom(builder.exportMapOf(step.Link.Target), false)

	case exportStepNamed:
		var dependency *ExportMap
		if step.FromModule && step.Link.Resolved {
			dependency = builder.exportMapOf(step.Link.Target)
		}
		for _, spec := range step.Specs {
			if !step.FromModule {
				if meta, ok := builder.namespaceImportMeta(local, spec.LocalIdent); ok {
					exports.Set(spec.Exported, meta)
				} else {
					exports.Set(spec.Exported, nil)
				}
				continue
			}
			if !spec.LocalOK {
				continue
			}
			if dependency == nil {
				exports.Set(spec.Exported, nil)
				continue
			}
			if !dependency.Has(spec.Local) {
				continue
			}
			exports.Set(spec.Exported, dependency.Get(spec.Local))
		}
	}
}

// namespaceImportMeta finds the `import * as localName` the file re-exports
// under some other name. Every import declaration before the matching one has
// its export map built along the way, because that is what the search used to
// do while walking the statements, and the maps it leaves behind are visible
// to the rest of this query.
func (builder *exportBuilder) namespaceImportMeta(local *localExports, localName string) (*ExportMeta, bool) {
	if localName == "" {
		return nil, false
	}
	for _, binding := range local.Imports {
		if !binding.Link.Resolved {
			continue
		}
		imports := builder.exportMapOf(binding.Link.Target)
		if binding.NamespaceName == localName {
			return &ExportMeta{Namespace: imports}, true
		}
	}
	return nil, false
}

// IsImportPathIgnored matches eslint-plugin-import's shared `import/ignore`
// setting for resolved import target paths. Callers that ask more than once
// for the same settings should compile them with [compileModuleSettings] and
// ask that instead.
func IsImportPathIgnored(settings map[string]interface{}, fileName string) bool {
	return compileModuleSettings(settings).IsIgnoredPath(fileName)
}

func sourceFileHasExport(ctx rule.RuleContext, sourceFile *ast.SourceFile, exportName string, builder *exportBuilder) (bool, bool) {
	if sourceFile == nil || !ast.IsExternalModule(sourceFile) {
		return false, false
	}

	key := exportKey{file: sourceFile, name: exportName}
	if builder.seen[key] {
		return false, true
	}
	builder.seen[key] = true
	defer delete(builder.seen, key)

	statements := sourceFile.Statements
	if statements == nil {
		return false, true
	}

	for _, stmt := range statements.Nodes {
		if stmt == nil {
			continue
		}

		if exportedDeclarationHasName(stmt, exportName) {
			return true, true
		}

		switch stmt.Kind {
		case ast.KindExportAssignment:
			if exportName == defaultExportName && exportAssignmentHasDefault(ctx, sourceFile, stmt.AsExportAssignment()) {
				return true, true
			}
		case ast.KindNamespaceExportDeclaration:
			if exportName == defaultExportName && compilerOptionsESModuleInterop(ctx) {
				return true, true
			}
		case ast.KindExportDeclaration:
			found, done := exportDeclarationHasName(ctx, sourceFile, stmt.AsExportDeclaration(), exportName, builder)
			if done {
				return found, true
			}
		}
	}

	if exportName == defaultExportName && compilerOptionsESModuleInterop(ctx) && sourceFileHasDirectNamespaceExport(sourceFile) {
		return true, true
	}

	return false, true
}

func exportedDeclarationHasName(stmt *ast.Node, exportName string) bool {
	if !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
		return false
	}

	if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsDefault) {
		return exportName == defaultExportName
	}

	switch stmt.Kind {
	case ast.KindVariableStatement:
		return variableStatementDeclaresName(stmt, exportName)
	case ast.KindFunctionDeclaration,
		ast.KindClassDeclaration,
		ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindEnumDeclaration,
		ast.KindModuleDeclaration:
		name := stmt.Name()
		return name != nil && moduleExportNameMatches(name, exportName)
	}

	return false
}

func exportAssignmentHasDefault(ctx rule.RuleContext, sourceFile *ast.SourceFile, exportAssignment *ast.ExportAssignment) bool {
	if exportAssignment == nil {
		return false
	}
	if !exportAssignment.IsExportEquals {
		return true
	}

	// Match eslint-plugin-import's TypeScript export-assignment visitor:
	// `export = namespace` gets a synthetic default only under esModuleInterop,
	// while non-namespace local declarations and re-export-like expressions do.
	name, ok := exportAssignmentReferencedIdentifier(exportAssignment.Expression)
	if !ok {
		return true
	}
	kind, ok := sourceFileExportAssignmentLocalDeclarationKind(sourceFile, name)
	if !ok {
		return true
	}
	if kind != exportAssignmentLocalDeclarationModule {
		return true
	}
	return compilerOptionsESModuleInterop(ctx)
}

// The tsgo shim exposes CompilerOptions fields but not GetESModuleInterop.
//
//nolint:staticcheck // esModuleInterop still needs to be inspected for import/export compatibility.
func compilerOptionsESModuleInterop(ctx rule.RuleContext) bool {
	if ctx.Program == nil || ctx.Program.Options() == nil {
		return false
	}
	options := ctx.Program.Options()
	if options.ESModuleInterop != core.TSUnknown {
		return options.ESModuleInterop == core.TSTrue
	}
	switch options.Module {
	case core.ModuleKindNode16, core.ModuleKindNodeNext, core.ModuleKindPreserve:
		return true
	default:
		return false
	}
}

type exportAssignmentLocalDeclarationKind int

const (
	exportAssignmentLocalDeclarationOther exportAssignmentLocalDeclarationKind = iota
	exportAssignmentLocalDeclarationModule
)

func sourceFileExportAssignmentLocalDeclarationKind(sourceFile *ast.SourceFile, name string) (exportAssignmentLocalDeclarationKind, bool) {
	if sourceFile == nil || sourceFile.Statements == nil {
		return exportAssignmentLocalDeclarationOther, false
	}
	for _, stmt := range sourceFile.Statements.Nodes {
		if stmt == nil {
			continue
		}

		switch stmt.Kind {
		case ast.KindVariableStatement:
			if variableStatementDeclaresName(stmt, name) {
				return exportAssignmentLocalDeclarationOther, true
			}
		case ast.KindFunctionDeclaration:
			if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsAmbient) && declarationHasName(stmt, name) {
				return exportAssignmentLocalDeclarationOther, true
			}
		case ast.KindClassDeclaration,
			ast.KindInterfaceDeclaration,
			ast.KindTypeAliasDeclaration,
			ast.KindEnumDeclaration:
			if declarationHasName(stmt, name) {
				return exportAssignmentLocalDeclarationOther, true
			}
		case ast.KindModuleDeclaration:
			if declarationHasName(stmt, name) {
				return exportAssignmentLocalDeclarationModule, true
			}
		}
	}
	return exportAssignmentLocalDeclarationOther, false
}

func declarationHasName(stmt *ast.Node, name string) bool {
	declName := stmt.Name()
	return declName != nil && moduleExportNameMatches(declName, name)
}

func variableStatementDeclaresName(stmt *ast.Node, name string) bool {
	declList := stmt.AsVariableStatement().DeclarationList
	if declList == nil || !ast.IsVariableDeclarationList(declList) {
		return false
	}
	for _, decl := range declList.AsVariableDeclarationList().Declarations.Nodes {
		if decl == nil || !ast.IsVariableDeclaration(decl) {
			continue
		}
		matched := false
		rslint_utils.CollectBindingNames(decl.AsVariableDeclaration().Name(), func(_ *ast.Node, bindingName string) {
			if bindingName == name {
				matched = true
			}
		})
		if matched {
			return true
		}
	}
	return false
}

func sourceFileHasDirectNamespaceExport(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.Statements == nil {
		return false
	}
	for _, stmt := range sourceFile.Statements.Nodes {
		if stmt == nil {
			continue
		}
		if exportedDeclarationAddsNamespaceExport(stmt) {
			return true
		}
		if stmt.Kind == ast.KindExportDeclaration && exportDeclarationAddsNamespaceExport(stmt.AsExportDeclaration()) {
			return true
		}
	}
	return false
}

func exportedDeclarationAddsNamespaceExport(stmt *ast.Node) bool {
	if !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
		return false
	}
	switch stmt.Kind {
	case ast.KindVariableStatement,
		ast.KindFunctionDeclaration,
		ast.KindClassDeclaration,
		ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindEnumDeclaration,
		ast.KindModuleDeclaration:
		return true
	}
	return false
}

func exportDeclarationAddsNamespaceExport(exportDecl *ast.ExportDeclaration) bool {
	if exportDecl == nil || exportDecl.ModuleSpecifier != nil || exportDecl.ExportClause == nil {
		return false
	}
	switch exportDecl.ExportClause.Kind {
	case ast.KindNamedExports:
		namedExports := exportDecl.ExportClause.AsNamedExports()
		return namedExports.Elements != nil && len(namedExports.Elements.Nodes) > 0
	case ast.KindNamespaceExport:
		return true
	}
	return false
}

func exportDeclarationHasName(ctx rule.RuleContext, sourceFile *ast.SourceFile, exportDecl *ast.ExportDeclaration, exportName string, builder *exportBuilder) (bool, bool) {
	if exportDecl == nil {
		return false, false
	}

	if exportDecl.ExportClause == nil {
		if exportDecl.ModuleSpecifier == nil || exportName == defaultExportName {
			return false, false
		}
		found, ok := hasExport(ctx, sourceFile, exportDecl.ModuleSpecifier, exportName, builder)
		if !ok {
			return true, true
		}
		return found, found
	}

	switch exportDecl.ExportClause.Kind {
	case ast.KindNamedExports:
		namedExports := exportDecl.ExportClause.AsNamedExports()
		if namedExports.Elements == nil {
			return false, false
		}
		for _, spec := range namedExports.Elements.Nodes {
			if spec == nil || spec.Kind != ast.KindExportSpecifier {
				continue
			}

			exportSpec := spec.AsExportSpecifier()
			if !moduleExportNameMatches(exportSpec.Name(), exportName) {
				continue
			}

			if exportDecl.ModuleSpecifier == nil {
				return true, true
			}

			sourceName := exportSpec.PropertyName
			if sourceName == nil {
				sourceName = exportSpec.Name()
			}

			localName, ok := moduleExportName(sourceName)
			if !ok {
				return false, true
			}

			hasName, ok := hasExport(ctx, sourceFile, exportDecl.ModuleSpecifier, localName, builder)
			if !ok {
				return true, true
			}
			return hasName, hasName
		}
	case ast.KindNamespaceExport:
		namespaceExport := exportDecl.ExportClause.AsNamespaceExport()
		matched := moduleExportNameMatches(namespaceExport.Name(), exportName)
		return matched, matched
	}

	return false, false
}

func moduleExportName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	return rslint_utils.GetStaticPropertyName(node)
}

func moduleExportNameMatches(node *ast.Node, exportName string) bool {
	if node == nil {
		return false
	}
	if exportName == defaultExportName {
		return ast.ModuleExportNameIsDefault(node)
	}
	name, ok := moduleExportName(node)
	return ok && name == exportName
}
