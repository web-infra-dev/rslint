package scope

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// builder walks the AST twice per scope: a hoisting pass that records the
// declarations belonging to the scope, then a visiting pass that descends into
// the children which open nested scopes.
type builder struct {
	manager           *Manager
	collectReferences bool
	referenceNames    map[string]struct{}
}

func (b *builder) push(kind Kind, block *ast.Node, parent *Scope) *Scope {
	s := newScope(kind, block, parent)
	if parent != nil && parent.GlobalAugmentation {
		s.GlobalAugmentation = true
	}
	b.manager.Scopes = append(b.manager.Scopes, s)
	return s
}

// reference records `id` as a read/write of whatever binding it resolves to.
// Declaration names, member names, labels, and other non-reference identifier
// positions are filtered out by isReferenceIdentifier.
func (b *builder) reference(id *ast.Node, s *Scope) {
	if !b.collectReferences || id == nil || s == nil {
		return
	}
	if b.referenceNames != nil {
		if _, relevant := b.referenceNames[id.Text()]; !relevant {
			return
		}
	}
	if !IsReferenceIdentifier(id) {
		return
	}
	space := ESLintReferenceSpace(id)
	ref := &Reference{
		Identifier:       id,
		From:             s,
		isValueReference: space.IncludesValue(),
		isTypeReference:  space.IncludesType(),
	}
	s.References = append(s.References, ref)
	b.manager.References = append(b.manager.References, ref)
}

func (b *builder) buildProgram(sf *ast.SourceFile) *Scope {
	global := b.push(KindGlobal, sf.AsNode(), nil)
	if sf.Statements == nil {
		return global
	}
	// First pass: hoist var / function / class / import names into the global scope.
	b.hoistStatements(sf.Statements.Nodes, global)
	// Second pass: walk children that introduce nested scopes.
	for _, stmt := range sf.Statements.Nodes {
		b.visitStatement(stmt, global)
	}
	return global
}

// hoistStatements collects declarations that belong to the enclosing
// variable scope (var / function / class / enum / interface / type /
// namespace / import). Block-scoped declarations (let/const/class inside
// a nested block) are collected in visitBlock.
func (b *builder) hoistStatements(statements []*ast.Node, s *Scope) {
	for _, stmt := range statements {
		b.hoistStatement(stmt, s, false)
	}
}

// addNamedDecl records a declaration whose binding name is the Name() child
// of `stmt` (covers function, class, interface, type alias, enum, namespace,
// import-equals). Returns without adding when the statement has no identifier
// name (for example anonymous default exports and ambient-module strings).
func (b *builder) addNamedDecl(stmt *ast.Node, s *Scope, kind DefKind, isValue bool) {
	n := stmt.Name()
	if n == nil || n.Kind != ast.KindIdentifier {
		return
	}
	// NodeFlagsAmbient also covers declarations nested in `declare namespace`
	// and declarations parsed from a .d.ts file, where no local `declare`
	// modifier is present.
	isAmbient := ast.HasSyntacticModifier(stmt, ast.ModifierFlagsAmbient) || utils.IsInAmbientContext(stmt)
	s.Add(&Variable{
		Name:            n.Text(),
		ID:              n,
		DefNode:         stmt,
		Parent:          stmt.Parent,
		Kind:            kind,
		IsValueBinding:  isValue,
		DeclareModifier: isAmbient,
	})
}

// hoistStatement records the declarations introduced by `stmt` into scope
// `s`. When `blockScope` is true, only block-scoped bindings (let/const/
// class/function/type/enum/namespace) are added — var is assumed to have
// been hoisted to an outer variable scope. Otherwise, every binding kind
// (including var) is added and the function's child statements are scanned
// for nested `var` declarations that also hoist to this scope.
func (b *builder) hoistStatement(stmt *ast.Node, s *Scope, blockScope bool) {
	if stmt == nil {
		return
	}
	switch stmt.Kind {
	case ast.KindVariableStatement:
		vs := stmt.AsVariableStatement()
		if vs == nil || vs.DeclarationList == nil {
			return
		}
		isVar := utils.IsVarKeyword(vs.DeclarationList)
		if blockScope && isVar {
			return
		}
		declareMod := ast.HasSyntacticModifier(stmt, ast.ModifierFlagsAmbient)
		b.collectVariableDeclarations(vs.DeclarationList, s, declareMod)
	case ast.KindFunctionDeclaration:
		b.addNamedDecl(stmt, s, DefFunctionName, true)
	case ast.KindClassDeclaration:
		b.addNamedDecl(stmt, s, DefClassName, true)
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		b.addNamedDecl(stmt, s, DefType, false)
	case ast.KindEnumDeclaration:
		b.addNamedDecl(stmt, s, DefEnumName, true)
	case ast.KindModuleDeclaration:
		// `declare global { ... }` reopens the global scope rather than binding
		// a namespace called `global`. Ambient `declare module 'str' { ... }`
		// uses a string-literal name and doesn't bind a variable either —
		// addNamedDecl skips that one automatically.
		if !ast.IsGlobalScopeAugmentation(stmt) && !utils.IsQualifiedNamespaceSegment(stmt) {
			b.addNamedDecl(stmt, s, DefNamespaceName, true)
		}
	case ast.KindImportDeclaration:
		if !blockScope {
			b.collectImport(stmt, s)
		}
	case ast.KindImportEqualsDeclaration:
		if !blockScope {
			b.addNamedDecl(stmt, s, DefImport, true)
		}
	case ast.KindForStatement:
		fs := stmt.AsForStatement()
		if !blockScope && fs != nil && fs.Initializer != nil && fs.Initializer.Kind == ast.KindVariableDeclarationList && utils.IsVarKeyword(fs.Initializer) {
			b.collectVariableDeclarations(fs.Initializer, s, false)
		}
	case ast.KindForInStatement, ast.KindForOfStatement:
		fs := stmt.AsForInOrOfStatement()
		if !blockScope && fs != nil && fs.Initializer != nil && fs.Initializer.Kind == ast.KindVariableDeclarationList && utils.IsVarKeyword(fs.Initializer) {
			b.collectVariableDeclarations(fs.Initializer, s, false)
		}
	}
	b.hoistFunctionsThroughScopeLessStatement(stmt, s)

	if blockScope {
		return
	}
	// For variable-scope hoisting, recurse into wrapper statements to collect
	// nested `var` declarations. Skip statement kinds that form scope
	// boundaries of their own (function / class / module bodies are processed
	// by visit*).
	switch stmt.Kind {
	case ast.KindBlock, ast.KindIfStatement, ast.KindWhileStatement,
		ast.KindDoStatement, ast.KindTryStatement, ast.KindCatchClause,
		ast.KindSwitchStatement, ast.KindCaseClause, ast.KindDefaultClause,
		ast.KindCaseBlock, ast.KindForStatement, ast.KindForInStatement,
		ast.KindForOfStatement, ast.KindLabeledStatement, ast.KindWithStatement,
		ast.KindExpressionStatement, ast.KindReturnStatement, ast.KindThrowStatement:
		stmt.ForEachChild(func(child *ast.Node) bool {
			b.hoistVarOnly(child, s)
			return false
		})
	}
}

// hoistFunctionsThroughScopeLessStatement adds function declarations nested
// under statements that create no lexical scope of their own. TSESTree puts
// those declarations in the nearest scope that already exists. A braced body
// stops the walk because visitBlock gives it its own scope; a for loop with a
// let/const initializer similarly owns a scope and is handled by visitFor*.
func (b *builder) hoistFunctionsThroughScopeLessStatement(stmt *ast.Node, s *Scope) {
	if stmt == nil {
		return
	}
	switch stmt.Kind {
	case ast.KindIfStatement:
		ifStmt := stmt.AsIfStatement()
		if ifStmt != nil {
			b.hoistScopeLessFunction(ifStmt.ThenStatement, s)
			b.hoistScopeLessFunction(ifStmt.ElseStatement, s)
		}
	case ast.KindLabeledStatement:
		if labeled := stmt.AsLabeledStatement(); labeled != nil {
			b.hoistScopeLessFunction(labeled.Statement, s)
		}
	case ast.KindWhileStatement:
		if whileStmt := stmt.AsWhileStatement(); whileStmt != nil {
			b.hoistScopeLessFunction(whileStmt.Statement, s)
		}
	case ast.KindDoStatement:
		if doStmt := stmt.AsDoStatement(); doStmt != nil {
			b.hoistScopeLessFunction(doStmt.Statement, s)
		}
	case ast.KindForStatement:
		forStmt := stmt.AsForStatement()
		if forStmt != nil && !hasBlockScopedForInitializer(forStmt.Initializer) {
			b.hoistScopeLessFunction(forStmt.Statement, s)
		}
	case ast.KindForInStatement, ast.KindForOfStatement:
		forInOrOf := stmt.AsForInOrOfStatement()
		if forInOrOf != nil && !hasBlockScopedForInitializer(forInOrOf.Initializer) {
			b.hoistScopeLessFunction(forInOrOf.Statement, s)
		}
	}
}

func hasBlockScopedForInitializer(initializer *ast.Node) bool {
	return initializer != nil && initializer.Kind == ast.KindVariableDeclarationList &&
		!utils.IsVarKeyword(initializer)
}

func (b *builder) hoistScopeLessFunction(stmt *ast.Node, s *Scope) {
	if stmt == nil {
		return
	}
	if stmt.Kind == ast.KindFunctionDeclaration {
		b.addNamedDecl(stmt, s, DefFunctionName, true)
		return
	}
	b.hoistFunctionsThroughScopeLessStatement(stmt, s)
}

// hoistVarOnly walks a subtree and only hoists `var` declarations (into the
// enclosing function/module/global scope). Function-like nodes and ambient
// module declarations terminate the walk.
func (b *builder) hoistVarOnly(node *ast.Node, s *Scope) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindVariableStatement:
		vs := node.AsVariableStatement()
		if vs != nil && vs.DeclarationList != nil && utils.IsVarKeyword(vs.DeclarationList) {
			declareMod := ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient)
			b.collectVariableDeclarations(vs.DeclarationList, s, declareMod)
		}
		return
	case ast.KindForStatement:
		fs := node.AsForStatement()
		if fs != nil && fs.Initializer != nil && fs.Initializer.Kind == ast.KindVariableDeclarationList && utils.IsVarKeyword(fs.Initializer) {
			b.collectVariableDeclarations(fs.Initializer, s, false)
		}
	case ast.KindForInStatement, ast.KindForOfStatement:
		fs := node.AsForInOrOfStatement()
		if fs != nil && fs.Initializer != nil && fs.Initializer.Kind == ast.KindVariableDeclarationList && utils.IsVarKeyword(fs.Initializer) {
			b.collectVariableDeclarations(fs.Initializer, s, false)
		}
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindArrowFunction, ast.KindMethodDeclaration,
		ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindClassStaticBlockDeclaration, ast.KindClassDeclaration,
		ast.KindClassExpression, ast.KindModuleDeclaration:
		return
	}
	node.ForEachChild(func(child *ast.Node) bool {
		b.hoistVarOnly(child, s)
		return false
	})
}

// collectVariableDeclarations walks a VariableDeclarationList and adds each
// binding identifier to the given scope. For destructuring patterns, the
// innermost BindingElement is stored as `DefNode` so that a rule can reach
// the initializer/default of that specific binding site later.
func (b *builder) collectVariableDeclarations(declList *ast.Node, s *Scope, declareMod bool) {
	utils.ForEachVariableDeclarationBinding(declList, func(declaration *ast.Node, identifier *ast.Node, name string) {
		defNode := declaration
		for parent := identifier.Parent; parent != nil && parent != declaration; parent = parent.Parent {
			if parent.Kind == ast.KindBindingElement {
				defNode = parent
				break
			}
		}
		s.Add(&Variable{
			Name:            name,
			ID:              identifier,
			DefNode:         defNode,
			Parent:          defNode.Parent,
			Kind:            DefVariable,
			IsValueBinding:  true,
			DeclareModifier: declareMod,
		})
	})
}

func (b *builder) collectImport(node *ast.Node, s *Scope) {
	importDecl := node.AsImportDeclaration()
	if importDecl == nil || importDecl.ImportClause == nil {
		return
	}
	clause := importDecl.ImportClause.AsImportClause()
	if clause == nil {
		return
	}
	isTypeImport := importDecl.ImportClause.IsTypeOnly()
	// Default import.
	if clause.Name() != nil && clause.Name().Kind == ast.KindIdentifier {
		s.Add(&Variable{
			Name:             clause.Name().Text(),
			ID:               clause.Name(),
			DefNode:          node,
			Parent:           node.Parent,
			Kind:             DefImport,
			IsValueBinding:   !isTypeImport,
			IsTypeOnlyImport: isTypeImport,
		})
	}
	if clause.NamedBindings == nil {
		return
	}
	switch clause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		ns := clause.NamedBindings.AsNamespaceImport()
		if ns != nil && ns.Name() != nil && ns.Name().Kind == ast.KindIdentifier {
			s.Add(&Variable{
				Name:             ns.Name().Text(),
				ID:               ns.Name(),
				DefNode:          node,
				Parent:           node.Parent,
				Kind:             DefImport,
				IsValueBinding:   !isTypeImport,
				IsTypeOnlyImport: isTypeImport,
			})
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
			specTypeOnly := spec.IsTypeOnly || isTypeImport
			s.Add(&Variable{
				Name:             spec.Name().Text(),
				ID:               spec.Name(),
				DefNode:          node,
				Parent:           node.Parent,
				Kind:             DefImport,
				IsValueBinding:   !specTypeOnly,
				IsTypeOnlyImport: specTypeOnly,
			})
		}
	}
}

// visitStatement recurses into nodes that may create nested scopes.
func (b *builder) visitStatement(stmt *ast.Node, parent *Scope) {
	if stmt == nil {
		return
	}
	switch stmt.Kind {
	case ast.KindFunctionDeclaration:
		b.visitFunctionLike(stmt, parent)
	case ast.KindClassDeclaration:
		b.visitClass(stmt, parent, false)
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		b.visitTypeDecl(stmt, parent)
	case ast.KindEnumDeclaration:
		b.visitEnum(stmt, parent)
	case ast.KindModuleDeclaration:
		b.visitModuleDecl(stmt, parent)
	case ast.KindBlock:
		b.visitBlock(stmt, parent)
	case ast.KindForStatement:
		b.visitForStatement(stmt, parent)
	case ast.KindForInStatement, ast.KindForOfStatement:
		b.visitForInOrOf(stmt, parent)
	case ast.KindIfStatement:
		ifStmt := stmt.AsIfStatement()
		if ifStmt != nil {
			b.visitExpression(ifStmt.Expression, parent)
			b.visitStatement(ifStmt.ThenStatement, parent)
			b.visitStatement(ifStmt.ElseStatement, parent)
		}
	case ast.KindWhileStatement:
		ws := stmt.AsWhileStatement()
		if ws != nil {
			b.visitExpression(ws.Expression, parent)
			b.visitStatement(ws.Statement, parent)
		}
	case ast.KindDoStatement:
		ds := stmt.AsDoStatement()
		if ds != nil {
			b.visitExpression(ds.Expression, parent)
			b.visitStatement(ds.Statement, parent)
		}
	case ast.KindTryStatement:
		ts := stmt.AsTryStatement()
		if ts != nil {
			b.visitStatement(ts.TryBlock, parent)
			if ts.CatchClause != nil {
				b.visitCatch(ts.CatchClause, parent)
			}
			b.visitStatement(ts.FinallyBlock, parent)
		}
	case ast.KindSwitchStatement:
		sw := stmt.AsSwitchStatement()
		if sw != nil {
			b.visitExpression(sw.Expression, parent)
			b.visitSwitchCases(sw, parent)
		}
	case ast.KindLabeledStatement:
		ls := stmt.AsLabeledStatement()
		if ls != nil {
			b.visitStatement(ls.Statement, parent)
		}
	case ast.KindReturnStatement:
		if rs := stmt.AsReturnStatement(); rs != nil {
			b.visitExpression(rs.Expression, parent)
		}
	case ast.KindThrowStatement:
		if ts := stmt.AsThrowStatement(); ts != nil {
			b.visitExpression(ts.Expression, parent)
		}
	case ast.KindExpressionStatement:
		if es := stmt.AsExpressionStatement(); es != nil {
			b.visitExpression(es.Expression, parent)
		}
	case ast.KindVariableStatement:
		vs := stmt.AsVariableStatement()
		if vs != nil && vs.DeclarationList != nil {
			b.visitVarDeclList(vs.DeclarationList, parent)
		}
	case ast.KindExportAssignment:
		ea := stmt.AsExportAssignment()
		if ea != nil {
			b.visitExpression(ea.Expression, parent)
		}
	case ast.KindExportDeclaration:
		b.visitExportDeclaration(stmt, parent)
	case ast.KindNamespaceExportDeclaration:
		// `export as namespace Name` reads the exported value. Keep the
		// scope-manager view consistent with IsReferenceIdentifier instead of
		// letting the generic statement walk discard its identifier child.
		b.reference(stmt.Name(), parent)
	case ast.KindImportDeclaration:
		// Bindings only; nothing to reference and no new scopes.
	case ast.KindImportEqualsDeclaration:
		// `import Z = A.B.C` references the left-most name of the qualified
		// module reference; `import Z = require('m')` references nothing.
		if ied := stmt.AsImportEqualsDeclaration(); ied != nil && ied.ModuleReference != nil &&
			ied.ModuleReference.Kind != ast.KindExternalModuleReference {
			b.visitExpression(ied.ModuleReference, parent)
		}
	case ast.KindWithStatement:
		ws := stmt.AsWithStatement()
		if ws != nil {
			b.visitExpression(ws.Expression, parent)
			b.visitStatement(ws.Statement, parent)
		}
	default:
		// Catch-all — traverse children.
		stmt.ForEachChild(func(child *ast.Node) bool {
			b.visitStatement(child, parent)
			return false
		})
	}
}

// visitExportDeclaration records the local name of each `export { local as
// exported }` specifier as a reference. Re-exports (`export { x } from 'm'`)
// name a binding in the other module, so they reference nothing here.
func (b *builder) visitExportDeclaration(stmt *ast.Node, parent *Scope) {
	decl := stmt.AsExportDeclaration()
	if decl == nil || decl.ModuleSpecifier != nil || decl.ExportClause == nil {
		return
	}
	if decl.ExportClause.Kind != ast.KindNamedExports {
		return
	}
	named := decl.ExportClause.AsNamedExports()
	if named == nil || named.Elements == nil {
		return
	}
	for _, elem := range named.Elements.Nodes {
		spec := elem.AsExportSpecifier()
		if spec == nil {
			continue
		}
		local := spec.PropertyName
		if local == nil {
			local = spec.Name()
		}
		if local != nil && local.Kind == ast.KindIdentifier {
			b.reference(local, parent)
		}
	}
}

// visitVarDeclList walks a VariableDeclarationList so that any nested
// function/class/conditional-type scopes in type annotations, initializers,
// or destructuring defaults get created.
func (b *builder) visitVarDeclList(declList *ast.Node, parent *Scope) {
	list := declList.AsVariableDeclarationList()
	if list == nil || list.Declarations == nil {
		return
	}
	for _, decl := range list.Declarations.Nodes {
		vd := decl.AsVariableDeclaration()
		if vd == nil {
			continue
		}
		if vd.Type != nil {
			b.visitExpression(vd.Type, parent)
		}
		if vd.Initializer != nil {
			b.visitExpression(vd.Initializer, parent)
		}
		if vd.Name() != nil {
			b.visitBindingPattern(vd.Name(), parent)
		}
	}
}

// visitBindingPattern recurses into destructuring defaults that may contain
// function / class expressions creating nested scopes, and into computed
// property keys of object patterns.
func (b *builder) visitBindingPattern(pattern *ast.Node, parent *Scope) {
	if pattern == nil {
		return
	}
	pattern.ForEachChild(func(child *ast.Node) bool {
		if child.Kind == ast.KindBindingElement {
			be := child.AsBindingElement()
			if be != nil {
				if be.PropertyName != nil && be.PropertyName.Kind == ast.KindComputedPropertyName {
					if cpn := be.PropertyName.AsComputedPropertyName(); cpn != nil {
						b.visitExpression(cpn.Expression, parent)
					}
				}
				if be.Initializer != nil {
					b.visitExpression(be.Initializer, parent)
				}
				if be.Name() != nil {
					b.visitBindingPattern(be.Name(), parent)
				}
			}
		}
		return false
	})
}

// visitExpression walks an expression to discover nested function/class
// expressions and other scope-creating constructs, recording every identifier
// that reads a binding along the way.
func (b *builder) visitExpression(expr *ast.Node, parent *Scope) {
	if expr == nil {
		return
	}
	switch expr.Kind {
	case ast.KindIdentifier:
		b.reference(expr, parent)
		return
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		b.visitFunctionLike(expr, parent)
		return
	case ast.KindMethodDeclaration, ast.KindConstructor,
		ast.KindGetAccessor, ast.KindSetAccessor:
		// Object-literal shorthand methods (and accessors) also need their
		// own function scope. Class-body members are already processed by
		// visitClass, so we only get here for the object-literal case.
		b.visitFunctionLike(expr, parent)
		return
	case ast.KindClassExpression:
		b.visitClass(expr, parent, true)
		return
	case ast.KindMappedType:
		b.visitMappedType(expr, parent)
		return
	}
	if ast.IsFunctionTypeNode(expr) || ast.IsConstructorTypeNode(expr) ||
		ast.IsCallSignatureDeclaration(expr) || ast.IsConstructSignatureDeclaration(expr) ||
		ast.IsMethodSignatureDeclaration(expr) {
		b.visitFunctionType(expr, parent)
		return
	}
	if expr.Kind == ast.KindConditionalType {
		b.visitConditionalType(expr, parent)
		return
	}
	expr.ForEachChild(func(child *ast.Node) bool {
		b.visitExpression(child, parent)
		return false
	})
}

// visitMappedType creates a scope holding the `[K in T]` type parameter, so
// that references to `K` in the template/value position resolve to it instead
// of leaking outward.
func (b *builder) visitMappedType(node *ast.Node, outer *Scope) {
	mapped := node.AsMappedTypeNode()
	if mapped == nil {
		return
	}
	mappedScope := b.push(KindType, node, outer)
	if mapped.TypeParameter != nil {
		if tp := mapped.TypeParameter.AsTypeParameterDeclaration(); tp != nil &&
			tp.Name() != nil && tp.Name().Kind == ast.KindIdentifier {
			mappedScope.Add(&Variable{
				Name:           tp.Name().Text(),
				ID:             tp.Name(),
				DefNode:        mapped.TypeParameter,
				Parent:         mapped.TypeParameter.Parent,
				Kind:           DefTypeParameter,
				IsValueBinding: false,
			})
		}
	}
	node.ForEachChild(func(child *ast.Node) bool {
		b.visitExpression(child, mappedScope)
		return false
	})
}

// visitConditionalType creates a scope for the conditional type so that
// `infer X` clauses in the extends-position introduce a binding visible to
// nested types and to the true branch. ESLint/typescript-eslint reports an
// inner `infer X` shadowing an outer one within the same conditional chain.
func (b *builder) visitConditionalType(node *ast.Node, outer *Scope) {
	cond := node.AsConditionalTypeNode()
	if cond == nil {
		return
	}
	condScope := b.push(KindType, node, outer)
	collectInferTypes(cond.ExtendsType, condScope)
	if cond.CheckType != nil {
		b.visitExpression(cond.CheckType, outer)
	}
	if cond.ExtendsType != nil {
		b.visitExpression(cond.ExtendsType, condScope)
	}
	if cond.TrueType != nil {
		b.visitExpression(cond.TrueType, condScope)
	}
	if cond.FalseType != nil {
		b.visitExpression(cond.FalseType, outer)
	}
}

// collectInferTypes walks `extendsType` and adds each `infer X` binding to
// the conditional-type scope. Nested function/conditional types stop the
// walk because they introduce their own scopes.
func collectInferTypes(node *ast.Node, s *Scope) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindFunctionType, ast.KindConstructorType, ast.KindConditionalType:
		return
	case ast.KindInferType:
		it := node.AsInferTypeNode()
		if it != nil && it.TypeParameter != nil {
			tp := it.TypeParameter.AsTypeParameterDeclaration()
			if tp != nil && tp.Name() != nil && tp.Name().Kind == ast.KindIdentifier {
				s.Add(&Variable{
					Name:           tp.Name().Text(),
					ID:             tp.Name(),
					DefNode:        it.TypeParameter,
					Parent:         it.TypeParameter.Parent,
					Kind:           DefTypeParameter,
					IsValueBinding: false,
				})
			}
		}
	}
	node.ForEachChild(func(child *ast.Node) bool {
		collectInferTypes(child, s)
		return false
	})
}

// visitFunctionType handles TS function/construct types and call/construct
// signatures. Parameters introduce bindings that may trigger (or be filtered
// out by) rule-specific type-parameter exceptions.
func (b *builder) visitFunctionType(node *ast.Node, outer *Scope) {
	// A method signature's computed key is evaluated in the enclosing scope.
	// Only its type parameters, parameters, and return type belong to the
	// function-type scope.
	var computedName *ast.Node
	if ast.IsMethodSignatureDeclaration(node) {
		if name := node.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
			computedName = name
			if cpn := name.AsComputedPropertyName(); cpn != nil {
				b.visitExpression(cpn.Expression, outer)
			}
		}
	}

	s := b.push(KindFunctionType, node, outer)
	b.addTypeParameters(node, s)
	b.addParameters(node, s, false)
	node.ForEachChild(func(child *ast.Node) bool {
		if child.Kind != ast.KindTypeParameter && child != computedName {
			b.visitExpression(child, s)
		}
		return false
	})
}

// addTypeParameters records `function f<T>` / `class C<T>` / `type X<T>`
// generic names into `s`. Each type parameter is a type-only binding. The
// constraint and default of every parameter are evaluated in `s` too, so that
// `<T extends U, U = V>` sees its siblings.
func (b *builder) addTypeParameters(node *ast.Node, s *Scope) {
	for _, tp := range node.TypeParameters() {
		if tp == nil {
			continue
		}
		tpDecl := tp.AsTypeParameterDeclaration()
		if tpDecl == nil || tpDecl.Name() == nil || tpDecl.Name().Kind != ast.KindIdentifier {
			continue
		}
		s.Add(&Variable{
			Name:           tpDecl.Name().Text(),
			ID:             tpDecl.Name(),
			DefNode:        tp,
			Parent:         tp.Parent,
			Kind:           DefTypeParameter,
			IsValueBinding: false,
		})
		b.visitExpression(tpDecl.Constraint, s)
		b.visitExpression(tpDecl.DefaultType, s)
	}
}

// addParameters records the binding identifiers from each parameter of
// `node` into `s`, optionally recursing into each parameter's type
// annotation and initializer so that nested function types create scopes.
func (b *builder) addParameters(node *ast.Node, s *Scope, recurseAnnotations bool) {
	for _, param := range node.Parameters() {
		if param == nil {
			continue
		}
		// The `this` pseudo-parameter declares no binding, but its type
		// annotation is a real type reference and still has to be walked.
		pn := param.Name()
		isThisParameter := pn != nil && pn.Kind == ast.KindIdentifier && pn.Text() == "this"
		if recurseAnnotations {
			// Parameter decorators (`@dec x: T`) run at class-definition time,
			// in the enclosing class scope (s.Parent for methods). Fall back
			// to `s` for standalone functions (decorators there are illegal
			// in TS but shouldn't crash).
			decScope := s
			if s.Parent != nil {
				decScope = s.Parent
			}
			for _, dec := range param.Decorators() {
				b.visitExpression(dec, decScope)
			}
			if paramDecl := param.AsParameterDeclaration(); paramDecl != nil && paramDecl.Type != nil {
				b.visitExpression(paramDecl.Type, s)
			}
			if init := param.Initializer(); init != nil {
				b.visitExpression(init, s)
			}
		}
		if isThisParameter || pn == nil {
			continue
		}
		if recurseAnnotations && pn.Kind != ast.KindIdentifier {
			b.visitBindingPattern(pn, s)
		}
		utils.CollectBindingNames(pn, func(id *ast.Node, name string) {
			s.Add(&Variable{
				Name:           name,
				ID:             id,
				DefNode:        param,
				Parent:         param.Parent,
				Kind:           DefParameter,
				IsValueBinding: true,
			})
		})
	}
}

// visitFunctionLike handles FunctionDeclaration, FunctionExpression,
// ArrowFunction, MethodDeclaration, Constructor, Get/SetAccessor.
func (b *builder) visitFunctionLike(node *ast.Node, outer *Scope) {
	// Computed method/accessor name (`[expr]`) is evaluated in the enclosing
	// scope (class body or object literal context). Walk it before pushing
	// the function's own scope so that references inside the key expression
	// resolve against the right scope.
	if name := node.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
		if cpn := name.AsComputedPropertyName(); cpn != nil && cpn.Expression != nil {
			b.visitExpression(cpn.Expression, outer)
		}
	}

	// Function expression with a name → intermediate scope holding just
	// the function's own name.
	fnExprScope := outer
	if node.Kind == ast.KindFunctionExpression {
		if n := node.Name(); n != nil && n.Kind == ast.KindIdentifier {
			innerFnName := b.push(KindFunctionExprName, node, outer)
			innerFnName.Add(&Variable{
				Name:           n.Text(),
				ID:             n,
				DefNode:        node,
				Parent:         node.Parent,
				Kind:           DefFnExprName,
				IsValueBinding: true,
			})
			fnExprScope = innerFnName
		}
	}

	fnScope := b.push(KindFunction, node, fnExprScope)
	b.addTypeParameters(node, fnScope)
	b.addParameters(node, fnScope, true)
	if retType := node.Type(); retType != nil {
		b.visitExpression(retType, fnScope)
	}

	// Body.
	body := node.Body()
	if body == nil {
		return
	}
	if body.Kind == ast.KindBlock {
		block := body.AsBlock()
		if block != nil && block.Statements != nil {
			b.hoistStatements(block.Statements.Nodes, fnScope)
			for _, stmt := range block.Statements.Nodes {
				b.visitStatement(stmt, fnScope)
			}
		}
	} else {
		// Arrow expression body.
		b.visitExpression(body, fnScope)
	}
}

func (b *builder) visitClass(node *ast.Node, outer *Scope, isExpression bool) {
	// Class-level decorators run BEFORE the class is defined — in the outer
	// scope. Any references inside them (e.g. `@((t) => { const x = 1; })`)
	// must resolve against outer bindings.
	for _, dec := range node.Decorators() {
		b.visitExpression(dec, outer)
	}

	classScope := b.push(KindClass, node, outer)
	// Inner scope always holds the class name (for both declarations and
	// expressions — ESLint's scope model duplicates the name here).
	if n := node.Name(); n != nil && n.Kind == ast.KindIdentifier {
		_ = isExpression
		classScope.Add(&Variable{
			Name:           n.Text(),
			ID:             n,
			DefNode:        node,
			Parent:         node.Parent,
			Kind:           DefClassInnerName,
			IsValueBinding: true,
		})
	}
	b.addTypeParameters(node, classScope)

	// Heritage clauses (`extends`, `implements`): expressions here are
	// evaluated when the class is defined and can contain IIFEs/arrows whose
	// body may shadow outer bindings. Walk them inside classScope so that
	// class type parameters remain visible to type arguments.
	node.ForEachChild(func(c *ast.Node) bool {
		if c.Kind == ast.KindHeritageClause {
			c.ForEachChild(func(t *ast.Node) bool {
				b.visitExpression(t, classScope)
				return false
			})
		}
		return false
	})

	for _, member := range node.Members() {
		if member == nil {
			continue
		}
		// Member decorators evaluate when the class is defined; use classScope
		// so that class type parameters remain visible.
		for _, dec := range member.Decorators() {
			b.visitExpression(dec, classScope)
		}
		switch member.Kind {
		case ast.KindMethodDeclaration, ast.KindConstructor,
			ast.KindGetAccessor, ast.KindSetAccessor:
			// visitFunctionLike handles the computed-name (if any) itself.
			b.visitFunctionLike(member, classScope)
		case ast.KindPropertyDeclaration:
			// Properties don't go through visitFunctionLike — walk the
			// computed key and the type annotation here.
			if memberName := member.Name(); memberName != nil && memberName.Kind == ast.KindComputedPropertyName {
				if cpn := memberName.AsComputedPropertyName(); cpn != nil && cpn.Expression != nil {
					b.visitExpression(cpn.Expression, classScope)
				}
			}
			if memberType := member.Type(); memberType != nil {
				b.visitExpression(memberType, classScope)
			}
			// A field initializer is its own execution context: eslint-scope
			// materializes a `class-field-initializer` scope for it.
			if init := member.Initializer(); init != nil {
				b.visitExpression(init, b.push(KindClassFieldInitializer, init, classScope))
			}
		case ast.KindIndexSignature:
			// `[key: string]: T` — the key type and the result type are both
			// evaluated in the class scope. The key itself binds nothing.
			for _, param := range member.Parameters() {
				if paramDecl := param.AsParameterDeclaration(); paramDecl != nil {
					b.visitExpression(paramDecl.Type, classScope)
				}
			}
			b.visitExpression(member.Type(), classScope)
		case ast.KindClassStaticBlockDeclaration:
			sb := member.AsClassStaticBlockDeclaration()
			if sb != nil && sb.Body != nil && sb.Body.Kind == ast.KindBlock {
				// Static block is its own variable scope for `var` purposes.
				staticScope := b.push(KindClassStaticBlock, member, classScope)
				block := sb.Body.AsBlock()
				if block != nil && block.Statements != nil {
					b.hoistStatements(block.Statements.Nodes, staticScope)
					for _, stmt := range block.Statements.Nodes {
						b.visitStatement(stmt, staticScope)
					}
				}
			}
		}
	}
}

func (b *builder) visitTypeDecl(node *ast.Node, outer *Scope) {
	typeScope := b.push(KindType, node, outer)
	b.addTypeParameters(node, typeScope)
	// Recurse into type body / heritage / members to discover FunctionType
	// and similar type-level scopes.
	node.ForEachChild(func(child *ast.Node) bool {
		if child.Kind != ast.KindTypeParameter {
			b.visitExpression(child, typeScope)
		}
		return false
	})
}

func (b *builder) visitEnum(node *ast.Node, outer *Scope) {
	enumScope := b.push(KindBlock, node, outer)
	ed := node.AsEnumDeclaration()
	if ed == nil || ed.Members == nil {
		return
	}
	// Enum members are bindings in the enum's inner scope. An outer
	// declaration with the same name is reported as shadowed by the member.
	for _, m := range ed.Members.Nodes {
		if m == nil {
			continue
		}
		em := m.AsEnumMember()
		if em == nil || em.Name() == nil {
			continue
		}
		// TS resolves a string-literal member name to a real name, so `enum E {
		// "a" = 1, b = a }` reads the member. The literal declares no
		// identifier though, which keeps rules from reporting at it.
		if name := em.Name(); name.Kind == ast.KindIdentifier || name.Kind == ast.KindStringLiteral {
			enumScope.Add(&Variable{
				Name:           name.Text(),
				ID:             name,
				DefNode:        m,
				Parent:         m.Parent,
				Kind:           DefEnumMember,
				IsValueBinding: true,
				Anonymous:      name.Kind != ast.KindIdentifier,
			})
		}
		if em.Initializer != nil {
			b.visitExpression(em.Initializer, enumScope)
		}
	}
}

func (b *builder) visitModuleDecl(node *ast.Node, outer *Scope) {
	md := node.AsModuleDeclaration()
	if md == nil {
		return
	}
	// `declare global { ... }` — treat as continuation of the enclosing global
	// scope. Any bindings inside are conceptually global and the scopes under
	// it are marked so shadowing checks are skipped.
	if ast.IsGlobalScopeAugmentation(node) {
		augScope := b.push(KindModule, node, outer)
		augScope.GlobalAugmentation = true
		if md.Body != nil {
			b.walkModuleBlock(md.Body, augScope)
		}
		return
	}
	moduleScope := b.push(KindModule, node, outer)
	// Inherit the global-augmentation flag from the parent chain.
	if outer != nil && outer.GlobalAugmentation {
		moduleScope.GlobalAugmentation = true
	}
	if md.Body == nil {
		return
	}
	if md.Body.Kind == ast.KindModuleDeclaration {
		// Nested namespace chain: `namespace A.B.C { }` — unwrap.
		b.visitModuleDecl(md.Body, moduleScope)
		return
	}
	b.walkModuleBlock(md.Body, moduleScope)
}

func (b *builder) walkModuleBlock(body *ast.Node, s *Scope) {
	if body == nil {
		return
	}
	if body.Kind == ast.KindModuleBlock {
		mb := body.AsModuleBlock()
		if mb != nil && mb.Statements != nil {
			b.hoistStatements(mb.Statements.Nodes, s)
			for _, stmt := range mb.Statements.Nodes {
				b.visitStatement(stmt, s)
			}
		}
	}
}

func (b *builder) visitBlock(block *ast.Node, outer *Scope) {
	parent := block.Parent
	// If this block is the body of a function-like, it has already been processed by visitFunctionLike.
	if parent != nil {
		switch parent.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindArrowFunction, ast.KindMethodDeclaration,
			ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindClassStaticBlockDeclaration:
			return
		case ast.KindCatchClause:
			// Handled in visitCatch.
			return
		}
	}
	s := b.push(KindBlock, block, outer)
	blk := block.AsBlock()
	if blk == nil || blk.Statements == nil {
		return
	}
	// Block-scoped declarations (let/const/class/function-in-strict/type/...) live at this block.
	for _, stmt := range blk.Statements.Nodes {
		b.hoistStatement(stmt, s, true)
	}
	// Vars were already collected by the enclosing variable scope's first
	// pass. Recurse without adding them a second time.
	for _, stmt := range blk.Statements.Nodes {
		b.visitStatement(stmt, s)
	}
}

// forInitScope handles a for/for-in/for-of initializer. A let/const
// initializer introduces a new block scope wrapping the loop body; otherwise
// the initializer stays in `outer`. Returns the scope the loop body lives in.
func (b *builder) forInitScope(forNode *ast.Node, initializer *ast.Node, outer *Scope) *Scope {
	if initializer != nil && initializer.Kind == ast.KindVariableDeclarationList && !utils.IsVarKeyword(initializer) {
		forScope := b.push(KindBlock, forNode, outer)
		b.collectVariableDeclarations(initializer, forScope, false)
		b.visitVarDeclList(initializer, forScope)
		return forScope
	}
	if initializer != nil {
		if initializer.Kind == ast.KindVariableDeclarationList {
			b.visitVarDeclList(initializer, outer)
		} else {
			b.visitExpression(initializer, outer)
		}
	}
	return outer
}

func (b *builder) visitForStatement(stmt *ast.Node, outer *Scope) {
	fs := stmt.AsForStatement()
	if fs == nil {
		return
	}
	body := b.forInitScope(stmt, fs.Initializer, outer)
	if body != outer {
		b.hoistScopeLessFunction(fs.Statement, body)
	}
	if fs.Condition != nil {
		b.visitExpression(fs.Condition, body)
	}
	if fs.Incrementor != nil {
		b.visitExpression(fs.Incrementor, body)
	}
	if fs.Statement != nil {
		b.visitStatement(fs.Statement, body)
	}
}

func (b *builder) visitForInOrOf(stmt *ast.Node, outer *Scope) {
	fs := stmt.AsForInOrOfStatement()
	if fs == nil {
		return
	}
	body := b.forInitScope(stmt, fs.Initializer, outer)
	if body != outer {
		b.hoistScopeLessFunction(fs.Statement, body)
	}
	// The iterated expression is evaluated while a `let`/`const` loop binding is
	// still in its temporal dead zone, so it resolves against the loop scope —
	// eslint-scope models this with a throwaway TDZ scope around the right-hand
	// side. `var` loop bindings hoist out, leaving `body == outer`.
	if fs.Expression != nil {
		b.visitExpression(fs.Expression, body)
	}
	if fs.Statement != nil {
		b.visitStatement(fs.Statement, body)
	}
}

func (b *builder) visitCatch(node *ast.Node, outer *Scope) {
	cc := node.AsCatchClause()
	if cc == nil {
		return
	}
	catchScope := b.push(KindCatch, node, outer)
	if cc.VariableDeclaration != nil {
		vd := cc.VariableDeclaration.AsVariableDeclaration()
		if vd != nil && vd.Name() != nil {
			utils.CollectBindingNames(vd.Name(), func(id *ast.Node, name string) {
				catchScope.Add(&Variable{
					Name:           name,
					ID:             id,
					DefNode:        cc.VariableDeclaration,
					Parent:         node,
					Kind:           DefCatch,
					IsValueBinding: true,
				})
			})
			// Destructuring defaults (`catch ({ message = x })`) are evaluated
			// in the catch scope.
			b.visitBindingPattern(vd.Name(), catchScope)
		}
	}
	if cc.Block != nil && cc.Block.Kind == ast.KindBlock {
		// The catch block is a BlockStatement whose direct parent is this CatchClause.
		// We treat it as nested: create an extra block scope for its body.
		bodyScope := b.push(KindBlock, cc.Block, catchScope)
		blk := cc.Block.AsBlock()
		if blk != nil && blk.Statements != nil {
			for _, stmt := range blk.Statements.Nodes {
				b.hoistStatement(stmt, bodyScope, true)
			}
			for _, stmt := range blk.Statements.Nodes {
				b.visitStatement(stmt, bodyScope)
			}
		}
	}
}

func (b *builder) visitSwitchCases(sw *ast.SwitchStatement, outer *Scope) {
	if sw.CaseBlock == nil {
		return
	}
	cb := sw.CaseBlock.AsCaseBlock()
	if cb == nil || cb.Clauses == nil {
		return
	}
	// Switch introduces a block scope for its clauses.
	switchScope := b.push(KindBlock, sw.AsNode(), outer)
	// First: hoist block-level declarations across all clauses.
	for _, clause := range cb.Clauses.Nodes {
		if clause == nil {
			continue
		}
		c := clause.AsCaseOrDefaultClause()
		if c == nil || c.Statements == nil {
			continue
		}
		for _, s := range c.Statements.Nodes {
			b.hoistStatement(s, switchScope, true)
		}
	}
	// Vars were already collected by the enclosing variable scope's first
	// pass. Recurse into statements without duplicating their definitions.
	for _, clause := range cb.Clauses.Nodes {
		if clause == nil {
			continue
		}
		c := clause.AsCaseOrDefaultClause()
		if c == nil {
			continue
		}
		if c.Expression != nil {
			b.visitExpression(c.Expression, switchScope)
		}
		if c.Statements != nil {
			for _, s := range c.Statements.Nodes {
				b.visitStatement(s, switchScope)
			}
		}
	}
}
