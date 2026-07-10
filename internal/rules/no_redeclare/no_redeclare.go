package no_redeclare

import (
	"fmt"
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type options struct {
	builtinGlobals         bool
	ignoreDeclarationMerge bool
}

type builtinGlobalsMode int

const (
	builtinGlobalsESLintCore builtinGlobalsMode = iota
	builtinGlobalsTypeScriptLibs
)

func parseOptionsWith(opts any, defaults options, allowIgnoreDeclarationMerge bool) options {
	result := defaults
	optsMap := utils.GetOptionsMap(opts)
	if optsMap == nil {
		return result
	}
	if v, ok := optsMap["builtinGlobals"].(bool); ok {
		result.builtinGlobals = v
	}
	if allowIgnoreDeclarationMerge {
		if v, ok := optsMap["ignoreDeclarationMerge"].(bool); ok {
			result.ignoreDeclarationMerge = v
		}
	}
	return result
}

func coreDefaults() options {
	return options{builtinGlobals: true, ignoreDeclarationMerge: false}
}

func typescriptDefaults() options {
	return options{builtinGlobals: true, ignoreDeclarationMerge: true}
}

var NoRedeclareRule = rule.Rule{
	Name: "no-redeclare",
	Run:  runWithOptions(coreDefaults(), false, builtinGlobalsESLintCore),
}

func RunTSESLint(ctx rule.RuleContext, opts []any) rule.RuleListeners {
	return runWithOptions(typescriptDefaults(), true, builtinGlobalsTypeScriptLibs)(ctx, opts)
}

func runWithOptions(defaults options, allowIgnoreDeclarationMerge bool, builtinMode builtinGlobalsMode) func(rule.RuleContext, []any) rule.RuleListeners {
	return func(ctx rule.RuleContext, opts []any) rule.RuleListeners {
		o := parseOptionsWith(rule.LegacyUnwrapOptions(opts), defaults, allowIgnoreDeclarationMerge)

		analyzeHoist := func(bodyNode *ast.Node, params []*ast.Node, isProgram bool) {
			s := newScopeDecls()
			for _, p := range params {
				if p == nil || p.Name() == nil {
					continue
				}
				utils.CollectBindingNames(p.Name(), func(id *ast.Node, name string) {
					s.addSyntax(name, id, ast.KindParameter)
				})
			}
			bodyNode.ForEachChild(func(child *ast.Node) bool {
				collect(child, s, true)
				return false
			})
			reportScope(ctx, s, o, isProgram, builtinMode)
		}

		// The linter never fires a KindSourceFile listener, so run the
		// program-scope analysis eagerly here.
		if ctx.SourceFile != nil {
			analyzeHoist(ctx.SourceFile.AsNode(), nil, true)
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				analyzeFunctionLike(node, analyzeHoist)
			},
			ast.KindFunctionExpression: func(node *ast.Node) {
				analyzeFunctionLike(node, analyzeHoist)
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				analyzeFunctionLike(node, analyzeHoist)
			},
			ast.KindMethodDeclaration: func(node *ast.Node) {
				analyzeFunctionLike(node, analyzeHoist)
			},
			ast.KindConstructor: func(node *ast.Node) {
				analyzeFunctionLike(node, analyzeHoist)
			},
			ast.KindGetAccessor: func(node *ast.Node) {
				analyzeFunctionLike(node, analyzeHoist)
			},
			ast.KindSetAccessor: func(node *ast.Node) {
				analyzeFunctionLike(node, analyzeHoist)
			},
			ast.KindClassStaticBlockDeclaration: func(node *ast.Node) {
				decl := node.AsClassStaticBlockDeclaration()
				if decl == nil || decl.Body == nil || decl.Body.Kind != ast.KindBlock {
					return
				}
				analyzeHoist(decl.Body, nil, false)
			},
			ast.KindModuleBlock: func(node *ast.Node) {
				analyzeHoist(node, nil, false)
			},
			ast.KindBlock: func(node *ast.Node) {
				parent := node.Parent
				if parent == nil {
					return
				}
				if isBlockBodyOwner(parent) {
					return
				}
				analyzeBlockScope(ctx, node, o, builtinMode)
			},
			ast.KindForStatement: func(node *ast.Node) {
				analyzeForScope(ctx, node, o, builtinMode)
			},
			ast.KindForInStatement: func(node *ast.Node) {
				analyzeForScope(ctx, node, o, builtinMode)
			},
			ast.KindForOfStatement: func(node *ast.Node) {
				analyzeForScope(ctx, node, o, builtinMode)
			},
			ast.KindSwitchStatement: func(node *ast.Node) {
				analyzeSwitchScope(ctx, node, o, builtinMode)
			},
		}
	}
}

// declInfo captures a single declaration of a name inside one scope.
// parentKind is the statement kind that introduced the binding, used to
// apply ignoreDeclarationMerge (class/interface/namespace/function/enum mixing).
type declInfo struct {
	id         *ast.Node
	parentKind ast.Kind
}

type scopeDecls struct {
	order []string
	decls map[string][]declInfo
}

func newScopeDecls() *scopeDecls {
	return &scopeDecls{decls: make(map[string][]declInfo)}
}

func (s *scopeDecls) add(name string, d declInfo) {
	if _, exists := s.decls[name]; !exists {
		s.order = append(s.order, name)
	}
	s.decls[name] = append(s.decls[name], d)
}

func (s *scopeDecls) addSyntax(name string, id *ast.Node, parentKind ast.Kind) {
	s.add(name, declInfo{id: id, parentKind: parentKind})
}

// isBlockBodyOwner reports whether `parent` treats its Block child as the
// body of a scope that we analyze through a dedicated listener (function-like
// or class static block). In those cases the generic Block listener must
// not re-analyze the same body.
func isBlockBodyOwner(parent *ast.Node) bool {
	return ast.IsFunctionLikeOrClassStaticBlockDeclaration(parent)
}

func analyzeFunctionLike(node *ast.Node, analyzeHoist func(*ast.Node, []*ast.Node, bool)) {
	body := node.Body()
	if body == nil || body.Kind != ast.KindBlock {
		// Expression-bodied arrows have no nested declarations beyond params.
		// Duplicate parameter names are already a parse error, so there is
		// nothing useful to report for that case.
		return
	}
	analyzeHoist(body, node.Parameters(), false)
}

// collect walks a subtree accumulating declarations into the enclosing hoist
// scope `s`. When `immediate` is true, every declaration kind is recorded;
// once we descend into a nested block/loop/switch, only `var` declarations
// continue to hoist. Recursion stops at function-like boundaries (separate
// scopes) and at type-only nodes that cannot introduce value bindings.
func collect(node *ast.Node, s *scopeDecls, immediate bool) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindVariableStatement:
		varStmt := node.AsVariableStatement()
		if varStmt == nil || varStmt.DeclarationList == nil {
			return
		}
		isVar := utils.IsVarKeyword(varStmt.DeclarationList)
		if !isVar && !immediate {
			return
		}
		addVariableDeclarations(varStmt.DeclarationList, s)
		return

	case ast.KindVariableDeclarationList:
		// Appears as a ForStatement / ForIn / ForOf initializer.
		isVar := utils.IsVarKeyword(node)
		if !isVar && !immediate {
			return
		}
		addVariableDeclarations(node, s)
		return

	case ast.KindFunctionDeclaration:
		if !immediate {
			return
		}
		// A bodyless FunctionDeclaration is a TypeScript overload signature
		// (upstream `TSDeclareFunction`). ESLint's rule explicitly filters
		// these out before counting declarations.
		if node.Body() == nil {
			return
		}
		if n := node.Name(); n != nil && n.Kind == ast.KindIdentifier {
			s.addSyntax(n.AsIdentifier().Text, n, ast.KindFunctionDeclaration)
		}
		return

	case ast.KindClassDeclaration, ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration, ast.KindEnumDeclaration:
		if !immediate {
			return
		}
		if n := node.Name(); n != nil && n.Kind == ast.KindIdentifier {
			s.addSyntax(n.AsIdentifier().Text, n, node.Kind)
		}
		return

	case ast.KindModuleDeclaration:
		if !immediate {
			return
		}
		if n := node.Name(); n != nil && n.Kind == ast.KindIdentifier {
			s.addSyntax(n.AsIdentifier().Text, n, ast.KindModuleDeclaration)
		}
		return

	case ast.KindImportDeclaration:
		if !immediate {
			return
		}
		addImportDeclarations(node, s)
		return

	case ast.KindImportEqualsDeclaration:
		if !immediate {
			return
		}
		addImportDeclarations(node, s)
		return

	// Function-like and class-like nodes introduce their own scopes — never
	// descend into their interior while collecting for the enclosing scope.
	case ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindConstructor,
		ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindClassExpression, ast.KindClassStaticBlockDeclaration:
		return
	}

	// Everything else is either a wrapper statement (if / try / while / with /
	// labeled / switch case / for / block, …) or an expression. Recurse and
	// mark the inner walk as non-immediate so only `var` continues to hoist.
	node.ForEachChild(func(child *ast.Node) bool {
		collect(child, s, false)
		return false
	})
}

func addVariableDeclarations(declList *ast.Node, s *scopeDecls) {
	list := declList.AsVariableDeclarationList()
	if list == nil || list.Declarations == nil {
		return
	}
	for _, decl := range list.Declarations.Nodes {
		if decl == nil || decl.Kind != ast.KindVariableDeclaration {
			continue
		}
		vd := decl.AsVariableDeclaration()
		if vd == nil || vd.Name() == nil {
			continue
		}
		utils.CollectBindingNames(vd.Name(), func(id *ast.Node, name string) {
			s.addSyntax(name, id, ast.KindVariableDeclaration)
		})
	}
}

func addImportDeclarations(node *ast.Node, s *scopeDecls) {
	parentKind := node.Kind
	for _, id := range utils.GetImportBindingNodes(node) {
		if id != nil && id.Kind == ast.KindIdentifier {
			s.addSyntax(id.AsIdentifier().Text, id, parentKind)
		}
	}
}

func analyzeBlockScope(ctx rule.RuleContext, blockNode *ast.Node, o options, builtinMode builtinGlobalsMode) {
	block := blockNode.AsBlock()
	if block == nil || block.Statements == nil {
		return
	}
	s := newScopeDecls()
	for _, stmt := range block.Statements.Nodes {
		collectTopLevel(stmt, s)
	}
	reportScope(ctx, s, o, false, builtinMode)
}

func analyzeForScope(ctx rule.RuleContext, node *ast.Node, o options, builtinMode builtinGlobalsMode) {
	var initializer *ast.Node
	switch node.Kind {
	case ast.KindForStatement:
		if fs := node.AsForStatement(); fs != nil {
			initializer = fs.Initializer
		}
	case ast.KindForInStatement, ast.KindForOfStatement:
		if fs := node.AsForInOrOfStatement(); fs != nil {
			initializer = fs.Initializer
		}
	}
	if initializer == nil || initializer.Kind != ast.KindVariableDeclarationList {
		return
	}
	if utils.IsVarKeyword(initializer) {
		return
	}
	s := newScopeDecls()
	addVariableDeclarations(initializer, s)
	reportScope(ctx, s, o, false, builtinMode)
}

func analyzeSwitchScope(ctx rule.RuleContext, node *ast.Node, o options, builtinMode builtinGlobalsMode) {
	sw := node.AsSwitchStatement()
	if sw == nil || sw.CaseBlock == nil {
		return
	}
	cb := sw.CaseBlock.AsCaseBlock()
	if cb == nil || cb.Clauses == nil {
		return
	}
	s := newScopeDecls()
	for _, clause := range cb.Clauses.Nodes {
		cc := clause.AsCaseOrDefaultClause()
		if cc == nil || cc.Statements == nil {
			continue
		}
		for _, stmt := range cc.Statements.Nodes {
			collectTopLevel(stmt, s)
		}
	}
	reportScope(ctx, s, o, false, builtinMode)
}

// collectTopLevel records direct block-scoped declarations within the top
// level of a block/switch/for scope. `var` is deliberately skipped because it
// hoists to an enclosing function-like scope, not the block.
func collectTopLevel(stmt *ast.Node, s *scopeDecls) {
	if stmt == nil {
		return
	}
	switch stmt.Kind {
	case ast.KindVariableStatement:
		varStmt := stmt.AsVariableStatement()
		if varStmt == nil || varStmt.DeclarationList == nil {
			return
		}
		if utils.IsVarKeyword(varStmt.DeclarationList) {
			return
		}
		addVariableDeclarations(varStmt.DeclarationList, s)
	case ast.KindFunctionDeclaration:
		// Skip TS overload signatures (bodyless function declarations).
		if stmt.Body() == nil {
			return
		}
		if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier {
			s.addSyntax(n.AsIdentifier().Text, n, ast.KindFunctionDeclaration)
		}
	case ast.KindClassDeclaration, ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration, ast.KindEnumDeclaration:
		if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier {
			s.addSyntax(n.AsIdentifier().Text, n, stmt.Kind)
		}
	case ast.KindModuleDeclaration:
		if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier {
			s.addSyntax(n.AsIdentifier().Text, n, ast.KindModuleDeclaration)
		}
	case ast.KindImportDeclaration:
		addImportDeclarations(stmt, s)
	case ast.KindImportEqualsDeclaration:
		addImportDeclarations(stmt, s)
	}
}

// applyMergeFilter drops declarations that are safe to merge under
// ignoreDeclarationMerge. Returns the list of declarations that still
// constitute a redeclaration (to be reported), possibly empty.
func applyMergeFilter(decls []declInfo) []declInfo {
	if len(decls) <= 1 {
		return decls
	}

	// All interfaces: merging always permitted.
	if allOfKind(decls, ast.KindInterfaceDeclaration) {
		return nil
	}

	// All namespaces: merging always permitted.
	if allOfKind(decls, ast.KindModuleDeclaration) {
		return nil
	}

	// Class + interface + namespace: permitted iff at most one class.
	if allWithinKinds(decls, ast.KindClassDeclaration, ast.KindInterfaceDeclaration, ast.KindModuleDeclaration) {
		classes := filterByKind(decls, ast.KindClassDeclaration)
		if len(classes) <= 1 {
			return nil
		}
		return classes
	}

	// Function + namespace: permitted iff at most one function.
	if allWithinKinds(decls, ast.KindFunctionDeclaration, ast.KindModuleDeclaration) {
		fns := filterByKind(decls, ast.KindFunctionDeclaration)
		if len(fns) <= 1 {
			return nil
		}
		return fns
	}

	// Enum + namespace: permitted iff at most one enum.
	if allWithinKinds(decls, ast.KindEnumDeclaration, ast.KindModuleDeclaration) {
		enums := filterByKind(decls, ast.KindEnumDeclaration)
		if len(enums) <= 1 {
			return nil
		}
		return enums
	}

	return decls
}

func allOfKind(decls []declInfo, kind ast.Kind) bool {
	for _, d := range decls {
		if d.parentKind != kind {
			return false
		}
	}
	return true
}

func allWithinKinds(decls []declInfo, kinds ...ast.Kind) bool {
	for _, d := range decls {
		match := false
		for _, k := range kinds {
			if d.parentKind == k {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	return true
}

func filterByKind(decls []declInfo, kind ast.Kind) []declInfo {
	var result []declInfo
	for _, d := range decls {
		if d.parentKind == kind {
			result = append(result, d)
		}
	}
	return result
}

type programGlobalDeclarations struct {
	ctx                         rule.RuleContext
	builtinMode                 builtinGlobalsMode
	builtinGlobals              bool
	defaultLibraryGlobals       map[string]bool
	defaultLibraryGlobalsLoaded bool
	inlineByName                map[string]rule.InlineGlobal
	inlineOrder                 []string
}

func newProgramGlobalDeclarations(ctx rule.RuleContext, o options, mode builtinGlobalsMode) *programGlobalDeclarations {
	result := &programGlobalDeclarations{
		ctx:            ctx,
		builtinMode:    mode,
		builtinGlobals: o.builtinGlobals,
	}

	for _, declaration := range ctx.InlineGlobals {
		// ESLint removes a name from the global scope when its final inline
		// setting is off, including all earlier comments for that name.
		if !declaration.Declared || len(declaration.NameRanges) == 0 {
			continue
		}
		if result.inlineByName == nil {
			result.inlineByName = make(map[string]rule.InlineGlobal)
		}
		if _, exists := result.inlineByName[declaration.Name]; !exists {
			result.inlineOrder = append(result.inlineOrder, declaration.Name)
		}
		result.inlineByName[declaration.Name] = declaration
	}

	return result
}

func (declarations *programGlobalDeclarations) isImplicitBuiltin(name string, syntax []declInfo) bool {
	if !declarations.builtinGlobals {
		return false
	}

	if finalSetting, exists := declarations.ctx.Globals[name]; exists && !finalSetting {
		// A final inline `:off` suppresses both configured and language globals.
		return false
	}
	if configured, exists := declarations.ctx.ConfigGlobals[name]; exists {
		// Explicit config replaces the language-provided setting.
		return configured
	}

	if declarations.builtinMode != builtinGlobalsTypeScriptLibs || declarations.ctx.Program == nil || declarations.ctx.TypeChecker == nil {
		return utils.IsECMAScriptGlobal(name)
	}

	if identifier := firstSyntaxIdentifier(syntax); identifier != nil {
		symbol := declarations.ctx.TypeChecker.GetSymbolAtLocation(identifier)
		return utils.IsSymbolFromDefaultLibrary(declarations.ctx.Program, symbol)
	}

	// Inline globals have no syntax node to resolve. Build the active default
	// library set only when such a name actually needs a lookup.
	if !declarations.defaultLibraryGlobalsLoaded {
		declarations.defaultLibraryGlobals = make(map[string]bool)
		utils.AddDefaultLibraryGlobals(declarations.defaultLibraryGlobals, declarations.ctx.Program, declarations.ctx.TypeChecker)
		declarations.defaultLibraryGlobalsLoaded = true
	}
	return declarations.defaultLibraryGlobals[name]
}

func firstSyntaxIdentifier(decls []declInfo) *ast.Node {
	for _, declaration := range decls {
		if declaration.id != nil {
			return declaration.id
		}
	}
	return nil
}

func allTypeOnlyDecls(decls []declInfo) bool {
	for _, d := range decls {
		if d.parentKind != ast.KindInterfaceDeclaration && d.parentKind != ast.KindTypeAliasDeclaration {
			return false
		}
	}
	return true
}

func reportScope(ctx rule.RuleContext, s *scopeDecls, o options, isProgram bool, builtinMode builtinGlobalsMode) {
	if ctx.SourceFile == nil {
		return
	}

	if !isProgram {
		for _, name := range s.order {
			decls := filterMergeDeclarations(s.decls[name], o.ignoreDeclarationMerge)
			reportDeclarationSequence(ctx, nil, name, decls, nil, false)
		}
		return
	}

	globals := newProgramGlobalDeclarations(ctx, o, builtinMode)
	isModule := ast.IsExternalModule(ctx.SourceFile)
	handled := make(map[string]bool, len(s.order))
	reports := make([]declarationReport, 0)

	for _, name := range s.order {
		decls := filterMergeDeclarations(s.decls[name], o.ignoreDeclarationMerge)
		inline := globals.inlineByName[name]
		reportProgramDeclarations(ctx, &reports, globals, name, decls, inline.NameRanges, isModule)
		handled[name] = true
	}

	// Inline-only globals never enter the syntax declaration collector.
	for _, name := range globals.inlineOrder {
		if handled[name] {
			continue
		}
		inline := globals.inlineByName[name]
		reportProgramDeclarations(ctx, &reports, globals, name, nil, inline.NameRanges, isModule)
	}

	sort.SliceStable(reports, func(i, j int) bool {
		if reports[i].textRange.Pos() == reports[j].textRange.Pos() {
			return reports[i].textRange.End() < reports[j].textRange.End()
		}
		return reports[i].textRange.Pos() < reports[j].textRange.Pos()
	})
	for _, report := range reports {
		reportRange(ctx, report.textRange, report.messageID, report.name)
	}
}

func filterMergeDeclarations(decls []declInfo, ignoreDeclarationMerge bool) []declInfo {
	if ignoreDeclarationMerge && len(decls) > 1 {
		return applyMergeFilter(decls)
	}
	return decls
}

func reportProgramDeclarations(
	ctx rule.RuleContext,
	reports *[]declarationReport,
	globals *programGlobalDeclarations,
	name string,
	syntax []declInfo,
	comments []core.TextRange,
	isModule bool,
) {
	// A module's syntax declarations live in its module scope, while config and
	// inline globals remain in the outer global scope. Type-only declarations
	// likewise do not collide with value-space globals.
	if isModule || (len(syntax) > 0 && allTypeOnlyDecls(syntax)) {
		reportDeclarationSequence(ctx, reports, name, syntax, nil, false)
		if len(comments) > 0 {
			reportDeclarationSequence(ctx, reports, name, nil, comments, globals.isImplicitBuiltin(name, nil))
		}
		return
	}
	reportDeclarationSequence(ctx, reports, name, syntax, comments, globals.isImplicitBuiltin(name, syntax))
}

// reportDeclarationSequence mirrors ESLint's declaration order: an implicit
// builtin first, then syntax identifiers, then each `/* global */` comment.
func reportDeclarationSequence(ctx rule.RuleContext, reports *[]declarationReport, name string, syntax []declInfo, comments []core.TextRange, implicitBuiltin bool) {
	if implicitBuiltin {
		for _, declaration := range syntax {
			reportNode(ctx, reports, declaration.id, "redeclaredAsBuiltin", name)
		}
		for _, comment := range comments {
			addDeclarationReport(ctx, reports, comment, "redeclaredAsBuiltin", name)
		}
		return
	}

	if len(syntax) > 0 {
		for _, declaration := range syntax[1:] {
			reportNode(ctx, reports, declaration.id, "redeclared", name)
		}
		for _, comment := range comments {
			addDeclarationReport(ctx, reports, comment, "redeclaredBySyntax", name)
		}
		return
	}

	if len(comments) > 1 {
		for _, comment := range comments[1:] {
			addDeclarationReport(ctx, reports, comment, "redeclared", name)
		}
	}
}

type declarationReport struct {
	textRange core.TextRange
	messageID string
	name      string
}

func reportNode(ctx rule.RuleContext, reports *[]declarationReport, node *ast.Node, messageID string, name string) {
	if node == nil {
		return
	}
	if reports == nil {
		ctx.ReportNode(node, rule.RuleMessage{Id: messageID, Description: formatMessage(messageID, name)})
		return
	}
	addDeclarationReport(ctx, reports, utils.TrimNodeTextRange(ctx.SourceFile, node), messageID, name)
}

func addDeclarationReport(ctx rule.RuleContext, reports *[]declarationReport, textRange core.TextRange, messageID string, name string) {
	if reports == nil {
		reportRange(ctx, textRange, messageID, name)
		return
	}
	*reports = append(*reports, declarationReport{textRange: textRange, messageID: messageID, name: name})
}

func reportRange(ctx rule.RuleContext, textRange core.TextRange, messageID string, name string) {
	ctx.ReportRange(textRange, rule.RuleMessage{Id: messageID, Description: formatMessage(messageID, name)})
}

func formatMessage(messageId, name string) string {
	switch messageId {
	case "redeclared":
		return fmt.Sprintf("'%s' is already defined.", name)
	case "redeclaredAsBuiltin":
		return fmt.Sprintf("'%s' is already defined as a built-in global variable.", name)
	case "redeclaredBySyntax":
		return fmt.Sprintf("'%s' is already defined by a variable declaration.", name)
	}
	return ""
}
