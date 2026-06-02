package no_deprecated_api

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/evaluator"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// constStringEvaluator reuses tsgo's own constant evaluator (shim/evaluator) —
// the same logic the compiler uses for enum members and literal types. The
// evaluateEntity callback is a no-op: like ESLint's getStringIfConstant invoked
// without a scope here, variable/enum references are not resolved — only
// literal, operator, and template folding.
var constStringEvaluator = evaluator.NewEvaluator(
	func(*ast.Node, *ast.Node) evaluator.Result { return evaluator.Result{} },
	ast.OEKParentheses,
)

// referenceTracker is the tsgo/rslint equivalent of `@eslint-community/eslint-utils`'s
// ReferenceTracker, specialized for n/no-deprecated-api.
//
// ESLint's ReferenceTracker is built on the ESLint scope manager
// (globalScope.set, variable.references, findVariable). rslint has no such
// scope graph, so this re-implements the same observable semantics on the tsgo
// AST. Two substitutions replace the scope manager:
//
//   - "is this identifier a reference to the global X" -> utils.IsShadowed
//     (a name is a live global reference when no local declaration shadows it).
//   - "find every read reference of this binding" -> findVariableReferences,
//     which prefers the TypeChecker (GetSymbolAtLocation symbol identity) and
//     falls back to AST scope analysis (IsNameShadowedBetween) when no type
//     information is available.
type referenceTracker struct {
	sourceFile     *ast.SourceFile
	checker        *checker.Checker // may be nil → scope-analysis fallback
	allIdentifiers []*ast.Node
	results        []trackedReference
	variableStack  []*ast.Node
}

// globalObjectNames mirrors ReferenceTracker's default global-object list: a
// deprecated global may also be reached as `global.X`, `globalThis.X`, etc.
var globalObjectNames = []string{"global", "globalThis", "self", "window"}

type refType int

const (
	refRead refType = iota
	refCall
	refConstruct
)

type trackedReference struct {
	node *ast.Node
	path []string
	typ  refType
	info *deprecatedInfo
}

func newReferenceTracker(ctx rule.RuleContext) *referenceTracker {
	t := &referenceTracker{
		sourceFile: ctx.SourceFile,
		checker:    ctx.TypeChecker,
	}
	collectIdentifiers(ctx.SourceFile.AsNode(), &t.allIdentifiers)
	return t
}

func collectIdentifiers(node *ast.Node, out *[]*ast.Node) {
	node.ForEachChild(func(child *ast.Node) bool {
		if child.Kind == ast.KindIdentifier {
			*out = append(*out, child)
		}
		collectIdentifiers(child, out)
		return false
	})
}

func (t *referenceTracker) emit(node *ast.Node, path []string, typ refType, info *deprecatedInfo) {
	p := make([]string, len(path))
	copy(p, path)
	t.results = append(t.results, trackedReference{node: node, path: p, typ: typ, info: info})
}

// capture runs fn with a fresh results buffer and returns what it emitted,
// restoring the previous buffer afterwards. Used by the ESM legacy filter.
func (t *referenceTracker) capture(fn func()) []trackedReference {
	saved := t.results
	t.results = nil
	fn()
	captured := t.results
	t.results = saved
	return captured
}

// --- global reference collection -------------------------------------------

// collectGlobalRefs returns every read reference to the global `name`, or nil if
// `name` is a "modified global" (declared or written somewhere) — mirroring
// ReferenceTracker's isModifiedGlobal early-out.
func (t *referenceTracker) collectGlobalRefs(name string) []*ast.Node {
	var refs []*ast.Node
	modified := false
	for _, id := range t.allIdentifiers {
		if id.AsIdentifier().Text != name {
			continue
		}
		if utils.IsNonReferenceIdentifier(id) {
			continue
		}
		if utils.IsShadowed(id, name) {
			continue
		}
		if utils.IsWriteReference(id) {
			modified = true
			continue
		}
		if !isReadReference(id) {
			continue
		}
		refs = append(refs, id)
	}
	if modified {
		return nil
	}
	return refs
}

func (t *referenceTracker) iterateGlobalReferences(tm map[string]*traceMap) {
	for key, nextMap := range tm {
		for _, ref := range t.collectGlobalRefs(key) {
			if nextMap.read != nil {
				t.emit(ref, []string{key}, refRead, nextMap.read)
			}
			t.iteratePropertyReferences(ref, []string{key}, nextMap)
		}
	}
	// `global.X` / `globalThis.X` / `self.X` / `window.X`: track the whole map
	// starting from the global-object reference (no READ at the bare object).
	wrapper := &traceMap{children: tm}
	for _, objName := range globalObjectNames {
		if _, isTracked := tm[objName]; isTracked {
			continue // avoid double-handling a name that's both a key and an object alias
		}
		for _, ref := range t.collectGlobalRefs(objName) {
			t.iteratePropertyReferences(ref, nil, wrapper)
		}
	}
}

// --- CommonJS / ESM reference collection -----------------------------------

func (t *referenceTracker) iterateCjsReferences(tm map[string]*traceMap) {
	for _, reqId := range t.collectGlobalRefs("require") {
		callNode := callFromCallee(reqId)
		if callNode == nil {
			continue
		}
		key := moduleNameArg(callNode)
		if key == "" {
			continue
		}
		nextMap, ok := tm[key]
		if !ok {
			continue
		}
		path := []string{key}
		if nextMap.read != nil {
			t.emit(callNode, path, refRead, nextMap.read)
		}
		t.iteratePropertyReferences(callNode, path, nextMap)
	}
}

func (t *referenceTracker) iterateEsmReferences(tm map[string]*traceMap) {
	if t.sourceFile.Statements == nil {
		return
	}
	for _, stmt := range t.sourceFile.Statements.Nodes {
		moduleId, decl, isExportAll, specs := importDeclarationInfo(stmt)
		if decl == nil {
			continue
		}
		nextMap, ok := tm[moduleId]
		if !ok {
			continue
		}
		path := []string{moduleId}
		if nextMap.read != nil {
			t.emit(decl, path, refRead, nextMap.read)
		}
		if isExportAll {
			for key, child := range nextMap.children {
				if child.read != nil {
					t.emit(decl, append(append([]string{}, path...), key), refRead, child.read)
				}
			}
			continue
		}
		for _, spec := range specs {
			t.iterateImportReferencesLegacy(spec, path, nextMap)
		}
	}
}

// iterateImportReferencesLegacy reproduces ReferenceTracker's `mode: "legacy"`
// handling for CJS-style trace maps imported via ESM: each module map is wrapped
// as `{ default: map, ...map }`, then the synthetic "default" segment is dropped
// from reported paths and bare-module READs (path length < 2) are filtered out.
func (t *referenceTracker) iterateImportReferencesLegacy(spec importSpec, path []string, moduleMap *traceMap) {
	legacy := &traceMap{children: map[string]*traceMap{"default": moduleMap}}
	for k, v := range moduleMap.children {
		legacy.children[k] = v
	}

	captured := t.capture(func() {
		t.iterateImportReferences(spec, path, legacy)
	})
	for _, r := range captured {
		np := filterExceptDefault(r.path)
		if len(np) >= 2 || r.typ != refRead {
			r.path = np
			t.results = append(t.results, r)
		}
	}
}

func (t *referenceTracker) iterateImportReferences(spec importSpec, path []string, tm *traceMap) {
	switch spec.kind {
	case specDefault, specNamed:
		key := spec.importedName
		nextMap := tm.children[key]
		if nextMap == nil {
			return
		}
		np := append(append([]string{}, path...), key)
		if nextMap.read != nil {
			t.emit(spec.node, np, refRead, nextMap.read)
		}
		if spec.localBinding != nil {
			t.iterateVariableReferences(spec.localBinding, np, nextMap, false)
		}
	case specNamespace:
		if spec.localBinding != nil {
			t.iterateVariableReferences(spec.localBinding, path, tm, false)
		}
	case specExport:
		key := spec.importedName
		nextMap := tm.children[key]
		if nextMap == nil {
			return
		}
		np := append(append([]string{}, path...), key)
		if nextMap.read != nil {
			t.emit(spec.node, np, refRead, nextMap.read)
		}
	}
}

// --- core recursion --------------------------------------------------------

// iteratePropertyReferences walks up from `node` matching member/call/new and
// destructuring shapes against the trace map, mirroring ReferenceTracker's
// _iteratePropertyReferences.
func (t *referenceTracker) iteratePropertyReferences(node *ast.Node, path []string, tm *traceMap) {
	for isPassThrough(node) {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return
	}

	switch parent.Kind {
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if utils.AccessExpressionObject(parent) != node {
			return
		}
		key, ok := accessKey(parent)
		if !ok {
			return
		}
		nextMap := tm.children[key]
		if nextMap == nil {
			return
		}
		np := append(append([]string{}, path...), key)
		if nextMap.read != nil {
			t.emit(parent, np, refRead, nextMap.read)
		}
		t.iteratePropertyReferences(parent, np, nextMap)

	case ast.KindCallExpression:
		if parent.AsCallExpression().Expression == node && tm.call != nil {
			t.emit(parent, path, refCall, tm.call)
		}

	case ast.KindNewExpression:
		if parent.AsNewExpression().Expression == node && tm.construct != nil {
			t.emit(parent, path, refConstruct, tm.construct)
		}

	case ast.KindBinaryExpression:
		bin := parent.AsBinaryExpression()
		if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == node {
			t.iterateLhsReferences(bin.Left, path, tm)
			t.iteratePropertyReferences(parent, path, tm)
		}

	case ast.KindVariableDeclaration:
		vd := parent.AsVariableDeclaration()
		if vd.Initializer == node {
			t.iterateLhsReferences(vd.Name(), path, tm)
		}

	case ast.KindBindingElement:
		be := parent.AsBindingElement()
		if be.Initializer == node { // default value: `{X: y = <node>}`
			t.iterateLhsReferences(be.Name(), path, tm)
		}

	case ast.KindParameter:
		pd := parent.AsParameterDeclaration()
		if pd.Initializer == node { // default value: `function (y = <node>)`
			t.iterateLhsReferences(pd.Name(), path, tm)
		}
	}
}

// iterateLhsReferences mirrors ReferenceTracker._iterateLhsReferences: it walks
// a binding target (identifier or object pattern) propagating the trace map.
func (t *referenceTracker) iterateLhsReferences(pat *ast.Node, path []string, tm *traceMap) {
	if pat == nil {
		return
	}
	switch pat.Kind {
	case ast.KindIdentifier:
		t.iterateVariableReferences(pat, path, tm, false)

	case ast.KindObjectBindingPattern:
		for _, elem := range bindingPatternElements(pat) {
			be := elem.AsBindingElement()
			if be.DotDotDotToken != nil {
				continue // rest element
			}
			key, ok := bindingElementKey(be)
			if !ok {
				continue
			}
			nextMap := tm.children[key]
			if nextMap == nil {
				continue
			}
			np := append(append([]string{}, path...), key)
			if nextMap.read != nil {
				t.emit(elem, np, refRead, nextMap.read)
			}
			t.iterateLhsReferences(be.Name(), np, nextMap)
		}

	case ast.KindObjectLiteralExpression:
		// Assignment-destructuring LHS: `({Key: target} = ...)`. tsgo models this
		// as an ObjectLiteralExpression, whereas ESTree (and upstream) use a single
		// ObjectPattern for both the declaration and assignment forms.
		for _, prop := range pat.AsObjectLiteralExpression().Properties.Nodes {
			var key string
			var ok bool
			var target *ast.Node
			switch prop.Kind {
			case ast.KindPropertyAssignment:
				pa := prop.AsPropertyAssignment()
				key, ok = utils.GetStaticPropertyName(pa.Name())
				target = pa.Initializer
			case ast.KindShorthandPropertyAssignment:
				if n := prop.AsShorthandPropertyAssignment().Name(); n != nil && n.Kind == ast.KindIdentifier {
					key, ok, target = n.AsIdentifier().Text, true, n
				}
			}
			if !ok || target == nil {
				continue
			}
			nextMap := tm.children[key]
			if nextMap == nil {
				continue
			}
			np := append(append([]string{}, path...), key)
			if nextMap.read != nil {
				t.emit(prop, np, refRead, nextMap.read)
			}
			t.iterateLhsReferences(target, np, nextMap)
		}
	}
}

// iterateVariableReferences mirrors ReferenceTracker._iterateVariableReferences:
// for every read reference of `declName`, optionally report a READ then keep
// walking member accesses. variableStack guards against self-referential cycles
// (e.g. `let fs = fs || require('fs')`).
func (t *referenceTracker) iterateVariableReferences(declName *ast.Node, path []string, tm *traceMap, shouldReport bool) {
	if declName == nil || declName.Kind != ast.KindIdentifier {
		return
	}
	for _, v := range t.variableStack {
		if v == declName {
			return
		}
	}
	t.variableStack = append(t.variableStack, declName)
	defer func() { t.variableStack = t.variableStack[:len(t.variableStack)-1] }()

	for _, ref := range t.findVariableReferences(declName) {
		if shouldReport && tm.read != nil {
			t.emit(ref, path, refRead, tm.read)
		}
		t.iteratePropertyReferences(ref, path, tm)
	}
}

// findVariableReferences returns every read reference of the binding introduced
// by `declName`. It prefers TypeChecker symbol identity, falling back to AST
// scope analysis when no type information is available (the hybrid strategy).
func (t *referenceTracker) findVariableReferences(declName *ast.Node) []*ast.Node {
	name := declName.AsIdentifier().Text

	if t.checker != nil {
		if sym := utils.GetReferenceSymbol(declName, t.checker); sym != nil {
			var refs []*ast.Node
			for _, id := range t.allIdentifiers {
				if id == declName || id.AsIdentifier().Text != name {
					continue
				}
				if utils.IsNonReferenceIdentifier(id) || !isReadReference(id) {
					continue
				}
				if utils.GetReferenceSymbol(id, t.checker) == sym {
					refs = append(refs, id)
				}
			}
			return refs
		}
	}

	// Fallback: scope analysis. References live within the declaration's
	// enclosing scope and must not be shadowed before reaching the binding.
	scope := utils.FindEnclosingScope(declName)
	var refs []*ast.Node
	for _, id := range t.allIdentifiers {
		if id == declName || id.AsIdentifier().Text != name {
			continue
		}
		if utils.IsNonReferenceIdentifier(id) || !isReadReference(id) {
			continue
		}
		if scope != nil && ast.FindAncestor(id, func(n *ast.Node) bool { return n == scope }) == nil {
			continue
		}
		if utils.IsNameShadowedBetween(id, scope, name) {
			continue
		}
		refs = append(refs, id)
	}
	return refs
}

// --- import/export specifier modelling -------------------------------------

type specKind int

const (
	specDefault specKind = iota
	specNamed
	specNamespace
	specExport
)

type importSpec struct {
	kind         specKind
	node         *ast.Node // specifier node (READ report anchor)
	importedName string    // source export name ("default" for default imports)
	localBinding *ast.Node // local binding identifier (for variable tracking)
}

// importDeclarationInfo extracts the module id, the declaration node (READ
// anchor), whether it's `export * from`, and the list of specifiers from an
// import/export-from statement. Returns decl == nil for unrelated statements.
func importDeclarationInfo(stmt *ast.Node) (moduleId string, decl *ast.Node, isExportAll bool, specs []importSpec) {
	switch stmt.Kind {
	case ast.KindImportDeclaration:
		imp := stmt.AsImportDeclaration()
		moduleId = utils.GetStaticStringValue(imp.ModuleSpecifier)
		if moduleId == "" {
			return "", nil, false, nil
		}
		decl = stmt
		if imp.ImportClause == nil {
			return moduleId, decl, false, nil // side-effect import: `import 'x'`
		}
		clause := imp.ImportClause.AsImportClause()
		if name := clause.Name(); name != nil {
			specs = append(specs, importSpec{kind: specDefault, node: name, importedName: "default", localBinding: name})
		}
		if clause.NamedBindings != nil {
			switch clause.NamedBindings.Kind {
			case ast.KindNamespaceImport:
				ns := clause.NamedBindings.AsNamespaceImport()
				specs = append(specs, importSpec{kind: specNamespace, node: clause.NamedBindings, localBinding: ns.Name()})
			case ast.KindNamedImports:
				named := clause.NamedBindings.AsNamedImports()
				if named.Elements != nil {
					for _, elem := range named.Elements.Nodes {
						is := elem.AsImportSpecifier()
						imported := is.Name()
						if is.PropertyName != nil {
							imported = is.PropertyName
						}
						specs = append(specs, importSpec{
							kind:         specNamed,
							node:         elem,
							importedName: identifierOrLiteralText(imported),
							localBinding: is.Name(),
						})
					}
				}
			}
		}
		return moduleId, decl, false, specs

	case ast.KindExportDeclaration:
		exp := stmt.AsExportDeclaration()
		if exp.ModuleSpecifier == nil {
			return "", nil, false, nil
		}
		moduleId = utils.GetStaticStringValue(exp.ModuleSpecifier)
		if moduleId == "" {
			return "", nil, false, nil
		}
		decl = stmt
		if exp.ExportClause == nil {
			return moduleId, decl, true, nil // `export * from 'x'`
		}
		if exp.ExportClause.Kind == ast.KindNamedExports {
			named := exp.ExportClause.AsNamedExports()
			if named.Elements != nil {
				for _, elem := range named.Elements.Nodes {
					es := elem.AsExportSpecifier()
					local := es.Name()
					if es.PropertyName != nil {
						local = es.PropertyName
					}
					specs = append(specs, importSpec{
						kind:         specExport,
						node:         elem,
						importedName: identifierOrLiteralText(local),
					})
				}
			}
		}
		return moduleId, decl, false, specs
	}
	return "", nil, false, nil
}

// --- small AST helpers -----------------------------------------------------

// isPassThrough reports whether `node`'s value flows unchanged to its parent —
// parentheses, TS type wrappers, both arms of `?:`, `||`/`&&`/`??`, and the last
// operand of a comma sequence. Mirrors ReferenceTracker.isPassThrough, plus
// ParenthesizedExpression (which ESTree drops but tsgo keeps).
func isPassThrough(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	// Parentheses + TS type-expression wrappers (as / satisfies / ! / <T>):
	// ESTree drops these, tsgo keeps them. OEKParentheses|OEKAssertions covers
	// exactly {Parenthesized, As, Satisfies, NonNull, TypeAssertion}.
	if ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions) {
		return true
	}
	switch parent.Kind {
	case ast.KindConditionalExpression:
		ce := parent.AsConditionalExpression()
		return ce.WhenTrue == node || ce.WhenFalse == node
	case ast.KindBinaryExpression:
		bin := parent.AsBinaryExpression()
		if bin.OperatorToken == nil {
			return false
		}
		switch bin.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
			return true
		case ast.KindCommaToken:
			return bin.Right == node
		}
	}
	return false
}

// callFromCallee returns the CallExpression whose callee is `node` (after
// pass-through unwrapping), or nil.
func callFromCallee(node *ast.Node) *ast.Node {
	for isPassThrough(node) {
		node = node.Parent
	}
	parent := node.Parent
	if parent != nil && parent.Kind == ast.KindCallExpression && parent.AsCallExpression().Expression == node {
		return parent
	}
	return nil
}

// foldConstantString returns the compile-time-constant string value of an
// expression (string / template literals, `+` concatenation, numeric/operator
// folding), or ("", false) if it isn't a constant. Used for require() module
// specifiers and member keys (e.g. require('f' + 's'), x['Slow' + 'Buffer']).
// Backed by tsgo's evaluator, so it matches the compiler's own folding semantics.
func foldConstantString(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	r := constStringEvaluator(node, node)
	if r.Value == nil {
		return "", false
	}
	return evaluator.AnyToString(r.Value), true
}

// accessKey returns the property key of a member access, folding constant
// concatenations in element-access keys (a['x' + 'y']).
func accessKey(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		if name := node.AsPropertyAccessExpression().Name(); name != nil {
			return name.Text(), true
		}
	case ast.KindElementAccessExpression:
		return foldConstantString(node.AsElementAccessExpression().ArgumentExpression)
	}
	return "", false
}

// moduleNameArg returns the constant string value of a call's first argument, or "".
func moduleNameArg(callNode *ast.Node) string {
	call := callNode.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return ""
	}
	if s, ok := foldConstantString(call.Arguments.Nodes[0]); ok {
		return s
	}
	return ""
}

// bindingElementKey returns the property key a binding element destructures
// (`{Key: local}`, `{'Key': local}`, `{['Key']: local}`, shorthand `{Key}`).
func bindingElementKey(be *ast.BindingElement) (string, bool) {
	if be.PropertyName != nil {
		if be.PropertyName.Kind == ast.KindComputedPropertyName {
			return foldConstantString(be.PropertyName.AsComputedPropertyName().Expression)
		}
		return utils.GetStaticPropertyName(be.PropertyName)
	}
	name := be.Name()
	if name != nil && name.Kind == ast.KindIdentifier {
		return name.AsIdentifier().Text, true
	}
	return "", false
}

func bindingPatternElements(pat *ast.Node) []*ast.Node {
	var elems []*ast.Node
	pat.ForEachChild(func(child *ast.Node) bool {
		if child.Kind == ast.KindBindingElement {
			elems = append(elems, child)
		}
		return false
	})
	return elems
}

func identifierOrLiteralText(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text
	}
	return ""
}

// isReadReference reports whether an identifier occurrence reads its binding.
// Only the bare left side of a plain `=` assignment is write-only; everything
// else (compound assignment, ++/--, member objects, arguments) reads.
func isReadReference(id *ast.Node) bool {
	parent := id.Parent
	if parent == nil {
		return true
	}
	if parent.Kind == ast.KindBinaryExpression {
		bin := parent.AsBinaryExpression()
		if bin.Left == id && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
			return false
		}
	}
	return true
}

func filterExceptDefault(path []string) []string {
	var out []string
	for i, p := range path {
		if i == 1 && p == "default" {
			continue
		}
		out = append(out, p)
	}
	return out
}

// toName renders a reported API path, e.g. ("buffer","Buffer") + CONSTRUCT ->
// "new buffer.Buffer()". The `node:` prefix is stripped for display.
func toName(typ refType, path []string) string {
	base := unprefixNodeColon(strings.Join(path, "."))
	switch typ {
	case refCall:
		return base + "()"
	case refConstruct:
		return "new " + base + "()"
	default:
		return base
	}
}
