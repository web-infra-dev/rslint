package no_redeclare

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// builtinGlobalsFallback enumerates ECMAScript core globals. Used only when
// no tsgo TypeChecker is available (e.g., rules run in environments without
// program/type info). When the checker is present we consult
// `utils.IsSymbolFromDefaultLibrary` instead, which transparently picks up
// DOM / Node / TypeScript `lib.*.d.ts` symbols as well.
var builtinGlobalsFallback = map[string]bool{
	"AggregateError": true, "Array": true, "ArrayBuffer": true,
	"AsyncDisposableStack": true, "AsyncIterator": true, "Atomics": true,
	"BigInt": true, "BigInt64Array": true, "BigUint64Array": true,
	"Boolean": true, "DataView": true, "Date": true,
	"decodeURI": true, "decodeURIComponent": true, "DisposableStack": true,
	"encodeURI": true, "encodeURIComponent": true,
	"Error": true, "escape": true, "EvalError": true,
	"FinalizationRegistry": true, "Float32Array": true, "Float64Array": true,
	"Function": true, "globalThis": true, "Infinity": true,
	"Int8Array": true, "Int16Array": true, "Int32Array": true,
	"Intl": true, "isFinite": true, "isNaN": true,
	"Iterator": true, "JSON": true, "Map": true, "Math": true,
	"NaN": true, "Number": true, "Object": true,
	"parseFloat": true, "parseInt": true, "Promise": true, "Proxy": true,
	"RangeError": true, "ReferenceError": true, "Reflect": true, "RegExp": true,
	"Set": true, "SharedArrayBuffer": true, "String": true,
	"SuppressedError": true, "Symbol": true, "SyntaxError": true,
	"TypeError": true, "Uint8Array": true, "Uint8ClampedArray": true,
	"Uint16Array": true, "Uint32Array": true, "unescape": true,
	"URIError": true, "undefined": true,
	"WeakMap": true, "WeakRef": true, "WeakSet": true,
}

type options struct {
	builtinGlobals         bool
	ignoreDeclarationMerge bool
}

func parseOptions(opts any) options {
	result := options{builtinGlobals: true, ignoreDeclarationMerge: true}
	optsMap := utils.GetOptionsMap(opts)
	if optsMap == nil {
		return result
	}
	if v, ok := optsMap["builtinGlobals"].(bool); ok {
		result.builtinGlobals = v
	}
	if v, ok := optsMap["ignoreDeclarationMerge"].(bool); ok {
		result.ignoreDeclarationMerge = v
	}
	return result
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

var NoRedeclareRule = rule.CreateRule(rule.Rule{
	Name: "no-redeclare",
	Run: func(ctx rule.RuleContext, opts any) rule.RuleListeners {
		o := parseOptions(opts)

		analyzeHoist := func(bodyNode *ast.Node, params []*ast.Node, isProgram bool) {
			s := newScopeDecls()
			for _, p := range params {
				if p == nil || p.Name() == nil {
					continue
				}
				utils.CollectBindingNames(p.Name(), func(id *ast.Node, name string) {
					s.add(name, declInfo{id: id, parentKind: ast.KindParameter})
				})
			}
			bodyNode.ForEachChild(func(child *ast.Node) bool {
				collect(child, s, true)
				return false
			})
			reportScope(ctx, s, o, isProgram)
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
				analyzeBlockScope(ctx, node, o)
			},
			ast.KindForStatement: func(node *ast.Node) {
				analyzeForScope(ctx, node, o)
			},
			ast.KindForInStatement: func(node *ast.Node) {
				analyzeForScope(ctx, node, o)
			},
			ast.KindForOfStatement: func(node *ast.Node) {
				analyzeForScope(ctx, node, o)
			},
			ast.KindSwitchStatement: func(node *ast.Node) {
				analyzeSwitchScope(ctx, node, o)
			},
		}
	},
})

// isBlockBodyOwner reports whether `parent` treats its Block child as the
// body of a scope that we analyze through a dedicated listener (function-like
// or class static block). In those cases the generic Block listener must
// not re-analyze the same body.
func isBlockBodyOwner(parent *ast.Node) bool {
	switch parent.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindArrowFunction, ast.KindMethodDeclaration,
		ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindClassStaticBlockDeclaration:
		return true
	}
	return false
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
			s.add(n.AsIdentifier().Text, declInfo{id: n, parentKind: ast.KindFunctionDeclaration})
		}
		return

	case ast.KindClassDeclaration, ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration, ast.KindEnumDeclaration:
		if !immediate {
			return
		}
		if n := node.Name(); n != nil && n.Kind == ast.KindIdentifier {
			s.add(n.AsIdentifier().Text, declInfo{id: n, parentKind: node.Kind})
		}
		return

	case ast.KindModuleDeclaration:
		if !immediate {
			return
		}
		if n := node.Name(); n != nil && n.Kind == ast.KindIdentifier {
			s.add(n.AsIdentifier().Text, declInfo{id: n, parentKind: ast.KindModuleDeclaration})
		}
		return

	case ast.KindImportDeclaration:
		if !immediate {
			return
		}
		collectImportNames(node, s)
		return

	case ast.KindImportEqualsDeclaration:
		if !immediate {
			return
		}
		ie := node.AsImportEqualsDeclaration()
		if ie != nil && ie.Name() != nil && ie.Name().Kind == ast.KindIdentifier {
			s.add(ie.Name().AsIdentifier().Text, declInfo{id: ie.Name(), parentKind: ast.KindImportEqualsDeclaration})
		}
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
			s.add(name, declInfo{id: id, parentKind: ast.KindVariableDeclaration})
		})
	}
}

func collectImportNames(node *ast.Node, s *scopeDecls) {
	importDecl := node.AsImportDeclaration()
	if importDecl == nil || importDecl.ImportClause == nil {
		return
	}
	clause := importDecl.ImportClause.AsImportClause()
	if clause == nil {
		return
	}
	if clause.Name() != nil && clause.Name().Kind == ast.KindIdentifier {
		s.add(clause.Name().AsIdentifier().Text, declInfo{id: clause.Name(), parentKind: ast.KindImportDeclaration})
	}
	if clause.NamedBindings == nil {
		return
	}
	switch clause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		ns := clause.NamedBindings.AsNamespaceImport()
		if ns != nil && ns.Name() != nil && ns.Name().Kind == ast.KindIdentifier {
			s.add(ns.Name().AsIdentifier().Text, declInfo{id: ns.Name(), parentKind: ast.KindImportDeclaration})
		}
	case ast.KindNamedImports:
		named := clause.NamedBindings.AsNamedImports()
		if named == nil || named.Elements == nil {
			return
		}
		for _, elem := range named.Elements.Nodes {
			if elem == nil {
				continue
			}
			spec := elem.AsImportSpecifier()
			if spec == nil || spec.Name() == nil || spec.Name().Kind != ast.KindIdentifier {
				continue
			}
			s.add(spec.Name().AsIdentifier().Text, declInfo{id: spec.Name(), parentKind: ast.KindImportDeclaration})
		}
	}
}

func analyzeBlockScope(ctx rule.RuleContext, blockNode *ast.Node, o options) {
	block := blockNode.AsBlock()
	if block == nil || block.Statements == nil {
		return
	}
	s := newScopeDecls()
	for _, stmt := range block.Statements.Nodes {
		collectTopLevel(stmt, s)
	}
	reportScope(ctx, s, o, false)
}

func analyzeForScope(ctx rule.RuleContext, node *ast.Node, o options) {
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
	reportScope(ctx, s, o, false)
}

func analyzeSwitchScope(ctx rule.RuleContext, node *ast.Node, o options) {
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
	reportScope(ctx, s, o, false)
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
			s.add(n.AsIdentifier().Text, declInfo{id: n, parentKind: ast.KindFunctionDeclaration})
		}
	case ast.KindClassDeclaration, ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration, ast.KindEnumDeclaration:
		if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier {
			s.add(n.AsIdentifier().Text, declInfo{id: n, parentKind: stmt.Kind})
		}
	case ast.KindModuleDeclaration:
		if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier {
			s.add(n.AsIdentifier().Text, declInfo{id: n, parentKind: ast.KindModuleDeclaration})
		}
	case ast.KindImportDeclaration:
		collectImportNames(stmt, s)
	case ast.KindImportEqualsDeclaration:
		ie := stmt.AsImportEqualsDeclaration()
		if ie != nil && ie.Name() != nil && ie.Name().Kind == ast.KindIdentifier {
			s.add(ie.Name().AsIdentifier().Text, declInfo{id: ie.Name(), parentKind: ast.KindImportEqualsDeclaration})
		}
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

// isBuiltinRedeclaration reports whether `name` in the current scope collides
// with a default-library (lib.*.d.ts) symbol.
//
// With a TypeChecker we resolve the user declaration's identifier and ask
// whether any of the merged symbol's declarations live in a default-lib file.
// Without a TypeChecker we fall back to a hard-coded list of ECMAScript core
// names.
//
// Either way, we first skip user declarations that only contribute to the
// type space (interface / type alias). ESLint's implicit-globals list only
// tracks value-level names, so augmenting a lib `interface Foo` with a user
// `interface Foo` or `type Foo` is not a redeclaration; reporting it would be
// a false positive relative to upstream.
func isBuiltinRedeclaration(ctx rule.RuleContext, name string, decls []declInfo) bool {
	if len(decls) == 0 {
		return false
	}
	if allTypeOnlyDecls(decls) {
		return false
	}
	if ctx.TypeChecker != nil && ctx.Program != nil {
		// Pick any user identifier — all declarations for this name merge
		// into the same symbol, and its declaration list includes the lib
		// ones if and only if this identifier shadows a lib global.
		sym := ctx.TypeChecker.GetSymbolAtLocation(decls[0].id)
		if sym == nil {
			return false
		}
		return utils.IsSymbolFromDefaultLibrary(ctx.Program, sym)
	}
	// Type info unavailable: approximate by matching ECMAScript core names
	// only in scripts (modules don't carry implicit globals).
	if ast.IsExternalModule(ctx.SourceFile) {
		return false
	}
	return builtinGlobalsFallback[name]
}

func allTypeOnlyDecls(decls []declInfo) bool {
	for _, d := range decls {
		if d.parentKind != ast.KindInterfaceDeclaration && d.parentKind != ast.KindTypeAliasDeclaration {
			return false
		}
	}
	return true
}

func reportScope(ctx rule.RuleContext, s *scopeDecls, o options, isProgram bool) {
	if ctx.SourceFile == nil {
		return
	}
	for _, name := range s.order {
		decls := s.decls[name]

		// Apply declaration-merge filtering first (this is an
		// `@typescript-eslint` extension, applied before classic iteration).
		if o.ignoreDeclarationMerge && len(decls) > 1 {
			filtered := applyMergeFilter(decls)
			if len(filtered) == 0 {
				continue
			}
			decls = filtered
		}

		isBuiltin := isProgram && o.builtinGlobals && isBuiltinRedeclaration(ctx, name, decls)

		totalDecls := len(decls)
		if isBuiltin {
			totalDecls++
		}
		if totalDecls < 2 {
			continue
		}

		// Per typescript-eslint's iterateDeclarations order the builtin (if
		// any) comes first and becomes the `declaration`; all user decls are
		// `extraDeclarations`. Each extra's type ('syntax') differs from the
		// builtin declaration's type, so every one is reported with the
		// builtin-specific message. When there is no builtin, decls[0] is
		// the declaration and the extras start at index 1.
		startIdx := 1
		messageId := "redeclared"
		if isBuiltin {
			startIdx = 0
			messageId = "redeclaredAsBuiltin"
		}
		for i := startIdx; i < len(decls); i++ {
			d := decls[i]
			ctx.ReportNode(d.id, rule.RuleMessage{
				Id:          messageId,
				Description: formatMessage(messageId, name),
			})
		}
	}
}

func formatMessage(messageId, name string) string {
	switch messageId {
	case "redeclared":
		return fmt.Sprintf("'%s' is already defined.", name)
	case "redeclaredAsBuiltin":
		return fmt.Sprintf("'%s' is already defined as a built-in global variable.", name)
	}
	return ""
}
