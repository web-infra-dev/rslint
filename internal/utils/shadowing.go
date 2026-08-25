package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
)

// IsNameShadowedBetween walks from `node` up to (but not including) `boundary`,
// returning true if any intermediate scope introduces a binding for `name`.
//
// Scopes examined:
//   - Function-like parameter lists (covers all 7 function-like kinds).
//   - Function expression names and hoisted `var` declarations in function bodies.
//   - Block-scoped declarations (var/let/const/function/class inside a Block).
//   - Catch-clause variable bindings (including destructured patterns).
//   - For-statement `let`/`const` init (scoped to the loop).
//   - For-in / for-of `let`/`const` init (scoped to the loop).
//   - Class declaration/expression names (scoped to the class body).
//   - Enum declaration names.
//
// Use this when a rule tracks a specific declaration (e.g. a parameter, class,
// or function name) and needs to ignore references that were shadowed before
// they reached the declaration site. For scope walks that should also examine
// the SourceFile or module boundary, use `IsShadowed` instead.
func IsNameShadowedBetween(node *ast.Node, boundary *ast.Node, name string) bool {
	prevChild := node
	for current := node.Parent; current != nil && current != boundary; current = current.Parent {
		if ast.IsFunctionLikeDeclaration(current) {
			if HasShadowingParameter(current, name) {
				return true
			}
			if current.Kind == ast.KindFunctionDeclaration || current.Kind == ast.KindFunctionExpression {
				if n := current.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
					return true
				}
			}
			// See IsShadowed: a reference within the parameter list itself
			// (e.g. a default value) is not shadowed by a body-level `var`.
			if !isDirectParameterOf(current, prevChild) {
				if body := current.Body(); body != nil && HasHoistedVarDeclaration(body, name) {
					return true
				}
			}
		}
		switch current.Kind {
		case ast.KindBlock:
			if HasShadowingDeclaration(current, name) {
				return true
			}
		case ast.KindCaseBlock:
			if HasShadowingDeclarationInCaseBlock(current, name) {
				return true
			}
		case ast.KindCatchClause:
			cc := current.AsCatchClause()
			if cc != nil && cc.VariableDeclaration != nil {
				vd := cc.VariableDeclaration.AsVariableDeclaration()
				if vd != nil && vd.Name() != nil && HasNameInBindingPattern(vd.Name(), name) {
					return true
				}
			}
		case ast.KindForStatement:
			forStmt := current.AsForStatement()
			if forStmt != nil && forStmt.Initializer != nil &&
				forStmt.Initializer.Kind == ast.KindVariableDeclarationList &&
				HasVarDeclListWithName(forStmt.Initializer, name) {
				return true
			}
		case ast.KindForInStatement, ast.KindForOfStatement:
			stmt := current.AsForInOrOfStatement()
			if stmt != nil && stmt.Initializer != nil &&
				stmt.Initializer.Kind == ast.KindVariableDeclarationList &&
				HasVarDeclListWithName(stmt.Initializer, name) {
				return true
			}
		case ast.KindClassDeclaration, ast.KindClassExpression, ast.KindEnumDeclaration:
			if n := current.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		}
		prevChild = current
	}
	return false
}

// GetReferenceSymbol returns the variable symbol for the given identifier.
// When the identifier is a shorthand property name or a local export specifier
// name, it returns the local value-binding symbol rather than the property or
// export symbol that `GetSymbolAtLocation` would otherwise produce.
func GetReferenceSymbol(node *ast.Node, typeChecker *checker.Checker) *ast.Symbol {
	if node == nil || typeChecker == nil {
		return nil
	}
	parent := node.Parent
	if parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment {
		shorthand := parent.AsShorthandPropertyAssignment()
		if shorthand != nil && shorthand.Name() == node {
			// GetSymbolAtLocation on a shorthand name always returns the
			// property symbol, even when the value binding ({x} referencing
			// an undeclared x) doesn't exist, so it cannot substitute here.
			return typeChecker.GetShorthandAssignmentValueSymbol(parent)
		}
	}
	if parent != nil && parent.Kind == ast.KindExportSpecifier &&
		!ast.IsTypeOnlyImportOrExportDeclaration(parent) &&
		!IsReExportSpecifier(parent) {
		spec := parent.AsExportSpecifier()
		isLocalName := spec != nil &&
			((spec.PropertyName != nil && spec.PropertyName == node) ||
				(spec.PropertyName == nil && spec.Name() == node))
		if isLocalName {
			if symbol := typeChecker.GetExportSpecifierLocalTargetSymbol(parent); symbol != nil {
				return symbol
			}
		}
	}
	return typeChecker.GetSymbolAtLocation(node)
}

// GetConstVariableInitializer returns the initializer of the single const
// variable declaration referenced by node. Destructured bindings resolve to
// their enclosing VariableDeclaration, matching ESLint scope definitions.
// Parentheses around the reference are transparent; TS assertion wrappers are
// intentionally not. When typeChecker is nil, binder-owned Locals tables are
// used as a best-effort lexical fallback for already-bound source files.
// References with no symbol, multiple declarations, non-variable
// declarations, or let/var declarations return nil.
func GetConstVariableInitializer(node *ast.Node, typeChecker *checker.Checker) *ast.Node {
	node = ast.SkipParentheses(node)
	if node == nil || !ast.IsIdentifier(node) {
		return nil
	}

	var symbol *ast.Symbol
	if typeChecker != nil {
		symbol = GetReferenceSymbol(node, typeChecker)
	} else {
		name := node.AsIdentifier().Text
		for current := node.Parent; current != nil; current = current.Parent {
			if current.Kind == ast.KindFunctionExpression ||
				current.Kind == ast.KindClassExpression {
				if expressionName := current.Name(); expressionName != nil &&
					ast.IsIdentifier(expressionName) &&
					expressionName.AsIdentifier().Text == name {
					// Named function/class expressions bind their name only
					// inside the expression. The binder stores that binding
					// separately from the ordinary Locals table, so stop
					// before an outer declaration with the same name wins.
					return nil
				}
			}
			if !ast.IsLocalsContainer(current) {
				continue
			}
			if local := ast.GetLocals(current)[name]; local != nil {
				symbol = local
				break
			}
		}
	}
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}

	declaration := symbol.Declarations[0]
	if ast.IsBindingElement(declaration) {
		declaration = EnclosingVariableDeclarationOfBindingElement(declaration)
	}
	if declaration == nil || !ast.IsVariableDeclaration(declaration) ||
		!ast.IsVarConst(declaration) {
		return nil
	}
	return declaration.AsVariableDeclaration().Initializer
}

// isDirectParameterOf reports whether child is one of fn's own Parameter
// nodes (not merely nested somewhere inside the parameter list).
func isDirectParameterOf(fn *ast.Node, child *ast.Node) bool {
	for _, param := range fn.Parameters() {
		if param == child {
			return true
		}
	}
	return false
}

// HasEnclosingTypeParameter reports whether any enclosing function-like or
// class declaration declares a type parameter called name. scope-manager keeps
// class type parameters in the lexical scope chain of static members, while
// TypeScript's own resolver deliberately hides them there, so rules that model
// scope-manager variables need this on top of a resolver-based lookup.
func HasEnclosingTypeParameter(node *ast.Node, name string) bool {
	prevChild := node
	inParameterDecorator := false
	crossedScope := false
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindParameter {
			inParameterDecorator = prevChild.Kind == ast.KindDecorator
		}
		isFunctionLike := ast.IsFunctionLikeDeclaration(current)
		isClassLike := ast.IsClassLike(current)
		if (isFunctionLike && !escapesFunctionScope(current, prevChild, inParameterDecorator, crossedScope)) ||
			(isClassLike && !escapesClassScope(prevChild, crossedScope)) {
			for _, typeParameter := range current.TypeParameters() {
				// TypeScript creates synthetic type parameters from JSDoc @template
				// tags and attaches them to the host declaration. They remain comments
				// rather than scope variables in ESTree and scope-manager.
				if typeParameter != nil && typeParameter.Flags&ast.NodeFlagsReparsed == 0 &&
					typeParameter.Name() != nil && typeParameter.Name().Text() == name {
					return true
				}
			}
		}
		if isFunctionLike || isClassLike {
			crossedScope = true
		}
		prevChild = current
	}
	return false
}

// HasEnclosingClassExpressionName reports whether scope-manager exposes an
// enclosing class expression's own name at node. A reference sitting directly
// in the class's decorator acquires the class scope, while a scope created
// inside that decorator is parented outside the class and cannot see the name.
func HasEnclosingClassExpressionName(node *ast.Node, name string) bool {
	prevChild := node
	crossedScope := false
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindClassExpression && !escapesClassScope(prevChild, crossedScope) {
			if className := current.Name(); className != nil && className.Kind == ast.KindIdentifier && className.Text() == name {
				return true
			}
		}
		if ast.IsFunctionLikeDeclaration(current) || ast.IsClassLike(current) {
			crossedScope = true
		}
		prevChild = current
	}
	return false
}

// escapesFunctionScope reports whether a reference that reached fn through
// prevChild is out of fn's own scope, so none of fn's bindings — parameters,
// type parameters, function name, body declarations — reach it.
//
// Two member positions do that. A decorator or a computed property name of fn
// belongs to the enclosing class or object literal: ESTree keeps both on the
// member rather than on the function expression that carries the type
// parameters, the parameters, and the body, so scope-manager evaluates them in
// the enclosing scope. And a function or class nested inside one of fn's
// parameter decorators gets its scope attached to the enclosing class rather
// than to the decorated function; a reference sitting directly in the
// decorator, with no scope in between, still acquires fn's scope.
func escapesFunctionScope(fn *ast.Node, prevChild *ast.Node, inParameterDecorator bool, crossedScope bool) bool {
	if prevChild == nil || !ast.IsFunctionLikeDeclaration(fn) {
		return false
	}
	if prevChild.Kind == ast.KindDecorator || prevChild == fn.Name() {
		return true
	}
	return inParameterDecorator && crossedScope && isDirectParameterOf(fn, prevChild)
}

// escapesClassScope reports whether a reference that reached a class through
// prevChild is out of the class's own scope, so neither its type parameters
// nor a class expression's own name reach it.
//
// A class's decorators are evaluated in the scope holding the class. ESTree
// keeps them on the class node, so a reference sitting directly in one, with
// no scope in between, still acquires the class scope, but scope-manager
// parents a scope created inside a decorator to the scope holding the class,
// leaving the class itself out of the chain.
func escapesClassScope(prevChild *ast.Node, crossedScope bool) bool {
	return prevChild != nil && prevChild.Kind == ast.KindDecorator && crossedScope
}

// IsShadowedFromParameterInitializer reports whether name is declared by the
// body of a function whose parameter list lexically contains node.
//
// It exists because IsShadowed answers with runtime lexical semantics, where
// the parameter environment is a *parent* of the body's variable environment,
// so a body declaration doesn't shadow a default value. scope-manager instead
// keeps parameter initializers and body declarations in one function scope, so
// `getVariableByName` finds the body's binding from a default value. Rules that
// model scope-manager variables need this on top of IsShadowed.
//
// Only the declarations scope-manager puts in that function scope count: vars
// hoisted from any depth, lexical declarations at the top level of the body,
// and function declarations with no block to hold them. A `let` in a nested
// block stays in that block's scope.
//
// A parameter decorator counts the same way while the reference sits directly
// in it, but scope-manager attaches any scope created *inside* a parameter
// decorator to the enclosing class rather than to the decorated function, so
// once an arrow, function, or class body intervenes the decorated function's
// scope is out of the chain entirely.
func IsShadowedFromParameterInitializer(node *ast.Node, name string) bool {
	prevChild := node
	inParameterDecorator := false
	crossedScope := false
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindParameter {
			inParameterDecorator = prevChild.Kind == ast.KindDecorator
		}
		if ast.IsFunctionLikeDeclaration(current) && isDirectParameterOf(current, prevChild) &&
			!escapesFunctionScope(current, prevChild, inParameterDecorator, crossedScope) {
			if body := current.Body(); body != nil &&
				(hasFunctionScopeDeclaration(body, name) || HasHoistedVarDeclaration(body, name)) {
				return true
			}
		}
		if ast.IsFunctionLikeDeclaration(current) || ast.IsClassLike(current) {
			crossedScope = true
		}
		prevChild = current
	}
	return false
}

// hasHoistedFunctionDeclaration reports whether container — a source file,
// function body, block, namespace body, or switch case block — declares a
// function named name in its own scope, including through a construct that
// creates no scope to hold it: the branch of an `if`, the body of a label, or
// the unbraced body of a loop. The declaration is defined in the innermost
// scope that already exists at its position, which is container's. A `for`
// with a `let`/`const` initializer does create a scope of its own, and a
// brace-delimited body keeps the declaration to itself.
func hasHoistedFunctionDeclaration(container *ast.Node, name string) bool {
	if container == nil {
		return false
	}
	switch container.Kind {
	case ast.KindSourceFile:
		sourceFile := container.AsSourceFile()
		return sourceFile != nil && sourceFile.Statements != nil &&
			hasHoistedFunctionDeclarationInStatements(sourceFile.Statements.Nodes, name)

	case ast.KindBlock:
		block := container.AsBlock()
		return block != nil && block.Statements != nil &&
			hasHoistedFunctionDeclarationInStatements(block.Statements.Nodes, name)

	case ast.KindModuleBlock:
		moduleBlock := container.AsModuleBlock()
		return moduleBlock != nil && moduleBlock.Statements != nil &&
			hasHoistedFunctionDeclarationInStatements(moduleBlock.Statements.Nodes, name)

	case ast.KindCaseBlock:
		caseBlock := container.AsCaseBlock()
		if caseBlock == nil || caseBlock.Clauses == nil {
			return false
		}
		for _, clause := range caseBlock.Clauses.Nodes {
			if clause == nil {
				continue
			}
			clauseData := clause.AsCaseOrDefaultClause()
			if clauseData == nil || clauseData.Statements == nil {
				continue
			}
			if hasHoistedFunctionDeclarationInStatements(clauseData.Statements.Nodes, name) {
				return true
			}
		}
	}
	return false
}

func hasHoistedFunctionDeclarationInStatements(statements []*ast.Node, name string) bool {
	for _, stmt := range statements {
		if hasHoistedFunctionDeclarationInStatement(stmt, name) {
			return true
		}
	}
	return false
}

func hasHoistedFunctionDeclarationInStatement(stmt *ast.Node, name string) bool {
	if stmt == nil {
		return false
	}
	switch stmt.Kind {
	case ast.KindFunctionDeclaration:
		if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
			return true
		}

	case ast.KindIfStatement:
		ifStmt := stmt.AsIfStatement()
		if ifStmt == nil {
			return false
		}
		return hasHoistedFunctionDeclarationInStatement(ifStmt.ThenStatement, name) ||
			hasHoistedFunctionDeclarationInStatement(ifStmt.ElseStatement, name)

	case ast.KindLabeledStatement:
		labeled := stmt.AsLabeledStatement()
		return labeled != nil && hasHoistedFunctionDeclarationInStatement(labeled.Statement, name)

	case ast.KindWhileStatement:
		while := stmt.AsWhileStatement()
		return while != nil && hasHoistedFunctionDeclarationInStatement(while.Statement, name)

	case ast.KindDoStatement:
		do := stmt.AsDoStatement()
		return do != nil && hasHoistedFunctionDeclarationInStatement(do.Statement, name)

	case ast.KindForStatement:
		forStmt := stmt.AsForStatement()
		if forStmt == nil || isBlockScopedForInitializer(forStmt.Initializer) {
			return false
		}
		return hasHoistedFunctionDeclarationInStatement(forStmt.Statement, name)

	case ast.KindForInStatement, ast.KindForOfStatement:
		forInOrOf := stmt.AsForInOrOfStatement()
		if forInOrOf == nil || isBlockScopedForInitializer(forInOrOf.Initializer) {
			return false
		}
		return hasHoistedFunctionDeclarationInStatement(forInOrOf.Statement, name)
	}
	return false
}

// isBlockScopedForInitializer reports whether a loop initializer declares
// `let`/`const`, which gives the loop a scope of its own.
func isBlockScopedForInitializer(initializer *ast.Node) bool {
	return initializer != nil && initializer.Kind == ast.KindVariableDeclarationList &&
		!IsVarKeyword(initializer)
}

// hasFunctionScopeDeclaration reports whether a function body declares name in
// any space scope-manager gives a function-scope variable: every value
// declaration HasLocalDeclarationInStatements recognizes at the top level, the
// function declarations hasHoistedFunctionDeclaration reaches through
// scope-less constructs, plus the TypeScript type-space declarations that
// exist as variables only in scope-manager's model.
func hasFunctionScopeDeclaration(body *ast.Node, name string) bool {
	if body == nil || body.Kind != ast.KindBlock {
		return false
	}
	block := body.AsBlock()
	if block == nil || block.Statements == nil {
		return false
	}
	if HasLocalDeclarationInStatements(block.Statements.Nodes, name) ||
		hasHoistedFunctionDeclarationInStatements(block.Statements.Nodes, name) {
		return true
	}
	for _, stmt := range block.Statements.Nodes {
		if stmt == nil {
			continue
		}
		switch stmt.Kind {
		case ast.KindTypeAliasDeclaration, ast.KindInterfaceDeclaration:
			if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		}
	}
	return false
}

// IsShadowed checks whether the given identifier name is shadowed by a local
// declaration at the usage site. It walks from node up to the SourceFile,
// checking every scope boundary for variable/function/class/enum/import
// declarations, namespace bodies, function parameters, catch variables, and
// hoisted var and function declarations.
func IsShadowed(node *ast.Node, name string) bool {
	prevChild := node
	inParameterDecorator := false
	crossedScope := false
	current := node.Parent
	for current != nil {
		if current.Kind == ast.KindParameter {
			inParameterDecorator = prevChild.Kind == ast.KindDecorator
		}
		switch current.Kind {
		case ast.KindSourceFile:
			sf := current.AsSourceFile()
			if sf != nil && sf.Statements != nil {
				if HasLocalDeclarationInStatements(sf.Statements.Nodes, name) {
					return true
				}
			}
			if HasHoistedVarDeclaration(current, name) || hasHoistedFunctionDeclaration(current, name) {
				return true
			}
			return false

		case ast.KindBlock:
			if HasShadowingDeclaration(current, name) || hasHoistedFunctionDeclaration(current, name) {
				return true
			}

		// A `namespace X {}` / `module X {}` body is a scope of its own: its
		// declarations — including `import X = ...` and `var`, which hoist to
		// the module block rather than out of it — shadow the outer binding
		// for the whole block, while the block itself still nests lexically
		// inside its parent scope.
		case ast.KindModuleBlock:
			moduleBlock := current.AsModuleBlock()
			if moduleBlock != nil && moduleBlock.Statements != nil {
				if HasLocalDeclarationInStatements(moduleBlock.Statements.Nodes, name) {
					return true
				}
			}
			if HasHoistedVarDeclaration(current, name) || hasHoistedFunctionDeclaration(current, name) {
				return true
			}

		// A class static block is a variable environment of its own: a `var`
		// inside it hoists to the block rather than out of it, so it shadows
		// the outer binding for the whole block. The static block is not a
		// function-like declaration, so the arm below never reaches its body.
		case ast.KindClassStaticBlockDeclaration:
			if body := current.AsClassStaticBlockDeclaration().Body; body != nil &&
				HasHoistedVarDeclaration(body, name) {
				return true
			}

		case ast.KindCaseBlock:
			if HasShadowingDeclarationInCaseBlock(current, name) || hasHoistedFunctionDeclaration(current, name) {
				return true
			}

		case ast.KindCatchClause:
			catchClause := current.AsCatchClause()
			if catchClause != nil && catchClause.VariableDeclaration != nil {
				varDecl := catchClause.VariableDeclaration.AsVariableDeclaration()
				if varDecl != nil && varDecl.Name() != nil {
					if HasNameInBindingPattern(varDecl.Name(), name) {
						return true
					}
				}
			}

		// A class declaration's name is declared in the scope holding the
		// class, so it shadows the call from anywhere the class does; a class
		// expression's name is declared in the class scope alone, which a
		// scope created inside the class's own decorator sits outside of.
		case ast.KindClassDeclaration, ast.KindClassExpression:
			if current.Kind == ast.KindClassExpression && escapesClassScope(prevChild, crossedScope) {
				break
			}
			if n := current.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}

		case ast.KindEnumDeclaration:
			if n := current.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}

		case ast.KindForStatement:
			forStmt := current.AsForStatement()
			if forStmt != nil && forStmt.Initializer != nil &&
				forStmt.Initializer.Kind == ast.KindVariableDeclarationList &&
				HasVarDeclListWithName(forStmt.Initializer, name) {
				return true
			}

		case ast.KindForInStatement, ast.KindForOfStatement:
			stmt := current.AsForInOrOfStatement()
			if stmt != nil && stmt.Initializer != nil &&
				stmt.Initializer.Kind == ast.KindVariableDeclarationList &&
				HasVarDeclListWithName(stmt.Initializer, name) {
				return true
			}

		default:
			if ast.IsFunctionLikeDeclaration(current) {
				if escapesFunctionScope(current, prevChild, inParameterDecorator, crossedScope) {
					break
				}
				if HasShadowingParameter(current, name) {
					return true
				}
				// Function declarations and expressions can shadow via their own name.
				if current.Kind == ast.KindFunctionDeclaration || current.Kind == ast.KindFunctionExpression {
					if n := current.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
						return true
					}
				}
				// A reference inside the parameter list itself (e.g. a default
				// value `a = foo`) is evaluated in the parameter environment,
				// which is a *parent* of the body's variable environment — so a
				// `var` declared in the body does not shadow it.
				if !isDirectParameterOf(current, prevChild) {
					if body := current.Body(); body != nil && HasHoistedVarDeclaration(body, name) {
						return true
					}
				}
			}
		}
		if ast.IsFunctionLikeDeclaration(current) || ast.IsClassLike(current) {
			crossedScope = true
		}
		prevChild = current
		current = current.Parent
	}
	return false
}

// IsQualifiedNamespaceSegment reports whether a module declaration is one
// segment of a dotted namespace name (`namespace A.B {}`). The parser nests
// one declaration per segment, and ESLint's scope manager creates a variable
// only for a namespace named by a plain identifier, so neither segment of a
// dotted name declares one.
func IsQualifiedNamespaceSegment(declaration *ast.Node) bool {
	if declaration == nil || declaration.Kind != ast.KindModuleDeclaration {
		return false
	}
	if declaration.Parent != nil && declaration.Parent.Kind == ast.KindModuleDeclaration {
		return true
	}
	body := declaration.AsModuleDeclaration().Body
	return body != nil && body.Kind == ast.KindModuleDeclaration
}

// HasShadowingParameter checks if a function-like node has a parameter
// whose binding name matches the given name (supports destructuring patterns).
func HasShadowingParameter(node *ast.Node, name string) bool {
	for _, param := range node.Parameters() {
		if param == nil {
			continue
		}
		nameNode := param.Name()
		if nameNode == nil {
			continue
		}
		found := false
		CollectBindingNames(nameNode, func(_ *ast.Node, n string) {
			if n == name {
				found = true
			}
		})
		if found {
			return true
		}
	}
	return false
}

// HasShadowingDeclaration checks if a block contains a variable, function,
// class, enum, or namespace declaration whose name matches the given name.
func HasShadowingDeclaration(node *ast.Node, name string) bool {
	if node.Kind != ast.KindBlock {
		return false
	}

	block := node.AsBlock()
	if block == nil || block.Statements == nil {
		return false
	}

	return hasShadowingDeclarationInStatements(block.Statements.Nodes, name)
}

// HasShadowingDeclarationInCaseBlock checks whether any case/default clause
// of a switch statement's CaseBlock declares name. All clauses of a switch
// share a single lexical scope (the CaseBlock), so a `let`/`const`/class/
// function declaration in one clause shadows references to that name in
// every other clause of the same switch, regardless of clause order.
func HasShadowingDeclarationInCaseBlock(node *ast.Node, name string) bool {
	if node.Kind != ast.KindCaseBlock {
		return false
	}

	caseBlock := node.AsCaseBlock()
	if caseBlock == nil || caseBlock.Clauses == nil {
		return false
	}

	for _, clause := range caseBlock.Clauses.Nodes {
		if clause == nil {
			continue
		}
		clauseData := clause.AsCaseOrDefaultClause()
		if clauseData == nil || clauseData.Statements == nil {
			continue
		}
		if hasShadowingDeclarationInStatements(clauseData.Statements.Nodes, name) {
			return true
		}
	}
	return false
}

// hasShadowingDeclarationInStatements checks if a list of statements
// (typically a Block's or switch clause's body) contains a variable,
// function, class, enum, or namespace declaration whose name matches name.
func hasShadowingDeclarationInStatements(statements []*ast.Node, name string) bool {
	for _, stmt := range statements {
		if stmt == nil {
			continue
		}
		switch stmt.Kind {
		case ast.KindVariableStatement:
			varStmt := stmt.AsVariableStatement()
			if varStmt == nil || varStmt.DeclarationList == nil {
				continue
			}
			declList := varStmt.DeclarationList.AsVariableDeclarationList()
			if declList == nil || declList.Declarations == nil {
				continue
			}
			for _, decl := range declList.Declarations.Nodes {
				if decl == nil || decl.Kind != ast.KindVariableDeclaration {
					continue
				}
				varDecl := decl.AsVariableDeclaration()
				if varDecl == nil || varDecl.Name() == nil {
					continue
				}
				found := false
				CollectBindingNames(varDecl.Name(), func(_ *ast.Node, n string) {
					if n == name {
						found = true
					}
				})
				if found {
					return true
				}
			}
		case ast.KindFunctionDeclaration:
			if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		case ast.KindClassDeclaration:
			if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		case ast.KindEnumDeclaration:
			if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		case ast.KindModuleDeclaration:
			// `namespace X {}` / `module X {}` introduce a value binding when
			// named by an identifier. Skip ambient modules (`declare module "x"`),
			// which use a string literal name and don't bind a variable, and
			// dotted names, which bind none of their segments.
			modDecl := stmt.AsModuleDeclaration()
			if modDecl != nil && modDecl.Name() != nil && !IsQualifiedNamespaceSegment(stmt) &&
				modDecl.Name().Kind == ast.KindIdentifier && modDecl.Name().Text() == name {
				return true
			}
		}
	}

	return false
}

// HasLocalDeclarationInStatements checks if a list of top-level statements
// (typically from a SourceFile) contains a variable, function, class, enum,
// module, or import declaration with the given name.
func HasLocalDeclarationInStatements(statements []*ast.Node, name string) bool {
	for _, stmt := range statements {
		if stmt == nil {
			continue
		}
		switch stmt.Kind {
		case ast.KindVariableStatement:
			varStmt := stmt.AsVariableStatement()
			if varStmt != nil && varStmt.DeclarationList != nil {
				declList := varStmt.DeclarationList.AsVariableDeclarationList()
				if declList != nil && declList.Declarations != nil {
					for _, decl := range declList.Declarations.Nodes {
						if decl != nil && decl.Kind == ast.KindVariableDeclaration {
							varDecl := decl.AsVariableDeclaration()
							if varDecl != nil && varDecl.Name() != nil {
								if HasNameInBindingPattern(varDecl.Name(), name) {
									return true
								}
							}
						}
					}
				}
			}

		case ast.KindFunctionDeclaration:
			if n := stmt.Name(); n != nil && n.Text() == name {
				return true
			}

		case ast.KindClassDeclaration:
			if n := stmt.Name(); n != nil && n.Text() == name {
				return true
			}

		case ast.KindEnumDeclaration:
			if n := stmt.Name(); n != nil && n.Text() == name {
				return true
			}

		case ast.KindModuleDeclaration:
			// Ambient module (`declare module "x"`) uses a string-literal name
			// and doesn't bind a variable — only identifier-named namespaces
			// do, and a dotted name binds none of its segments.
			modDecl := stmt.AsModuleDeclaration()
			if modDecl != nil && modDecl.Name() != nil && !IsQualifiedNamespaceSegment(stmt) &&
				modDecl.Name().Kind == ast.KindIdentifier && modDecl.Name().Text() == name {
				return true
			}

		case ast.KindImportEqualsDeclaration:
			importEquals := stmt.AsImportEqualsDeclaration()
			if importEquals != nil && importEquals.Name() != nil && importEquals.Name().Text() == name {
				return true
			}

		case ast.KindImportDeclaration:
			importDecl := stmt.AsImportDeclaration()
			if importDecl != nil && importDecl.ImportClause != nil {
				importClause := importDecl.ImportClause.AsImportClause()
				if importClause != nil {
					// Default import: import X from 'mod'
					if importClause.Name() != nil && importClause.Name().Text() == name {
						return true
					}
					// Named/namespace imports
					if importClause.NamedBindings != nil {
						switch importClause.NamedBindings.Kind {
						case ast.KindNamespaceImport:
							nsImport := importClause.NamedBindings.AsNamespaceImport()
							if nsImport != nil && nsImport.Name() != nil && nsImport.Name().Text() == name {
								return true
							}
						case ast.KindNamedImports:
							namedImports := importClause.NamedBindings.AsNamedImports()
							if namedImports != nil && namedImports.Elements != nil {
								for _, elem := range namedImports.Elements.Nodes {
									if elem != nil {
										importSpec := elem.AsImportSpecifier()
										if importSpec != nil && importSpec.Name() != nil && importSpec.Name().Text() == name {
											return true
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return false
}

// HasNameInBindingPattern checks if a binding pattern (including nested
// destructuring) contains a binding with the given name.
func HasNameInBindingPattern(pattern *ast.Node, name string) bool {
	found := false
	CollectBindingNames(pattern, func(_ *ast.Node, n string) {
		if n == name {
			found = true
		}
	})
	return found
}

// HasVarDeclListWithName checks if a VariableDeclarationList contains a
// declaration with the given name (supports destructuring).
func HasVarDeclListWithName(node *ast.Node, name string) bool {
	declList := node.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil {
		return false
	}
	for _, decl := range declList.Declarations.Nodes {
		if decl != nil && decl.Kind == ast.KindVariableDeclaration {
			varDecl := decl.AsVariableDeclaration()
			if varDecl != nil && varDecl.Name() != nil && HasNameInBindingPattern(varDecl.Name(), name) {
				return true
			}
		}
	}
	return false
}

// IsVarKeyword returns true if a VariableDeclarationList uses `var`
// (not `let`/`const`/`using`/`await using`).
func IsVarKeyword(node *ast.Node) bool {
	return node.Flags&ast.NodeFlagsBlockScoped == 0
}

// HasHoistedVarDeclaration recursively scans a subtree for `var` declarations
// with the given name. It stops at nested function boundaries and module
// declarations since `var` does not hoist past them.
func HasHoistedVarDeclaration(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindVariableStatement:
		varStmt := node.AsVariableStatement()
		if varStmt != nil && varStmt.DeclarationList != nil && IsVarKeyword(varStmt.DeclarationList) {
			if HasVarDeclListWithName(varStmt.DeclarationList, name) {
				return true
			}
		}
		return false

	case ast.KindForStatement:
		forStmt := node.AsForStatement()
		if forStmt != nil && forStmt.Initializer != nil &&
			forStmt.Initializer.Kind == ast.KindVariableDeclarationList &&
			IsVarKeyword(forStmt.Initializer) &&
			HasVarDeclListWithName(forStmt.Initializer, name) {
			return true
		}

	case ast.KindForInStatement, ast.KindForOfStatement:
		stmt := node.AsForInOrOfStatement()
		if stmt != nil && stmt.Initializer != nil &&
			stmt.Initializer.Kind == ast.KindVariableDeclarationList &&
			IsVarKeyword(stmt.Initializer) &&
			HasVarDeclListWithName(stmt.Initializer, name) {
			return true
		}

	// var does not hoist past function, static block, or module boundaries.
	case ast.KindModuleDeclaration:
		return false
	default:
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(node) {
			return false
		}
	}

	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if HasHoistedVarDeclaration(child, name) {
			found = true
			return true
		}
		return false
	})
	return found
}
