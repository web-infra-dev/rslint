package dynamic_import_chunkname

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// visitClass searches what runs when a class is evaluated: the extends clause
// and computed member names, which run before the class name is initialized,
// and then static initializers and static blocks, which run after. Method
// bodies and instance initializers run later, when called or constructed, so
// they are not searched. Class code is strict.
func (c *referenceChecker) visitClass(class *ast.Node) bool {
	savedStrict, savedScope := c.strict, len(c.scope)
	c.strict = true
	defer func() {
		c.strict = savedStrict
		c.scope = c.scope[:savedScope]
	}()

	index := -1
	if name := class.Name(); name != nil {
		c.scope = append(c.scope, scopeEntry{name: name.Text(), class: true})
		index = len(c.scope) - 1
	}

	var heritageClauses *ast.NodeList
	switch class.Kind {
	case ast.KindClassExpression:
		heritageClauses = class.AsClassExpression().HeritageClauses
	case ast.KindClassDeclaration:
		heritageClauses = class.AsClassDeclaration().HeritageClauses
	}
	if heritageClauses != nil {
		for _, clause := range heritageClauses.Nodes {
			for _, heritage := range clause.AsHeritageClause().Types.Nodes {
				if c.visit(heritage.AsExpressionWithTypeArguments().Expression) {
					return true
				}
			}
		}
	}
	for _, member := range class.Members() {
		if name := member.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
			if c.visit(name.AsComputedPropertyName().Expression) {
				return true
			}
		}
	}

	if index >= 0 {
		c.scope[index].bound = true
	}
	for _, member := range class.Members() {
		switch member.Kind {
		case ast.KindPropertyDeclaration:
			if !ast.HasStaticModifier(member) {
				continue
			}
			if initializer := member.AsPropertyDeclaration().Initializer; initializer != nil && c.visit(initializer) {
				return true
			}
		case ast.KindClassStaticBlockDeclaration:
			if c.visitStaticBlock(member.AsClassStaticBlockDeclaration().Body) {
				return true
			}
		}
	}
	return false
}

// visitStaticBlock searches the statements of a static block that are certain
// to run, with the names the block declares in scope. A name declared with var
// is visible throughout the block, whichever statement declares it.
func (c *referenceChecker) visitStaticBlock(block *ast.Node) bool {
	savedScope := len(c.scope)
	defer func() { c.scope = c.scope[:savedScope] }()
	for _, name := range hoistedVarNames(block) {
		c.scope = append(c.scope, scopeEntry{name: name, bound: true})
	}
	return c.visitStatements(block.AsBlock().Statements.Nodes)
}

// visitStatements runs a list of statements in order, in a block scope that
// holds the let, const, class and function declarations of the list.
func (c *referenceChecker) visitStatements(statements []*ast.Node) bool {
	savedScope := len(c.scope)
	defer func() { c.scope = c.scope[:savedScope] }()
	for _, name := range lexicalNames(statements) {
		c.scope = append(c.scope, scopeEntry{name: name, bound: true})
	}
	for _, statement := range statements {
		if c.visitStatement(statement) {
			return true
		}
	}
	return false
}

// visitStatement searches what a statement is certain to evaluate. Only
// statements that run straight through are entered. A branch or loop body, a
// switch case, a try block and a labeled statement may be skipped, abandoned by
// break or continue, or have their errors caught, so they are not searched;
// the expression that decides whether they run is evaluated, and so is the
// branch of an if whose condition is a constant.
func (c *referenceChecker) visitStatement(statement *ast.Node) bool {
	switch statement.Kind {
	case ast.KindExpressionStatement:
		return c.visit(statement.AsExpressionStatement().Expression)
	case ast.KindVariableStatement:
		return c.visitDeclarations(statement.AsVariableStatement().DeclarationList)
	case ast.KindClassDeclaration:
		return c.visitClass(statement)
	case ast.KindBlock:
		return c.visitStatements(statement.AsBlock().Statements.Nodes)
	case ast.KindThrowStatement:
		if c.visit(statement.AsThrowStatement().Expression) {
			return true
		}
		return c.invalid()
	case ast.KindIfStatement:
		branch := statement.AsIfStatement()
		if c.visit(branch.Expression) {
			return true
		}
		truthy, known := c.evaluator.EvalControlFlowTruthiness(branch.Expression)
		switch {
		case !known:
			return false
		case truthy:
			return c.visitStatement(branch.ThenStatement)
		case branch.ElseStatement != nil:
			return c.visitStatement(branch.ElseStatement)
		}
		return false
	case ast.KindWhileStatement:
		return c.visit(statement.AsWhileStatement().Expression)
	case ast.KindSwitchStatement:
		return c.visit(statement.AsSwitchStatement().Expression)
	case ast.KindForInStatement, ast.KindForOfStatement:
		return c.visit(statement.AsForInOrOfStatement().Expression)
	case ast.KindForStatement:
		return c.visitForHead(statement.AsForStatement())
	case ast.KindTryStatement:
		// The try and catch blocks may throw or skip code, but a finally block
		// runs however the statement completes.
		if finalizer := statement.AsTryStatement().FinallyBlock; finalizer != nil {
			return c.visitStatement(finalizer)
		}
	}
	return false
}

// visitForHead evaluates the initializer, which declares the names the loop
// sees, and the condition, which is evaluated at least once. The body and the
// incrementor run an unknown number of times.
func (c *referenceChecker) visitForHead(loop *ast.ForStatement) bool {
	savedScope := len(c.scope)
	defer func() { c.scope = c.scope[:savedScope] }()
	if initializer := loop.Initializer; initializer != nil {
		if initializer.Kind == ast.KindVariableDeclarationList {
			for _, declaration := range initializer.AsVariableDeclarationList().Declarations.Nodes {
				for _, name := range bindingNames(declaration.Name()) {
					c.scope = append(c.scope, scopeEntry{name: name, bound: true})
				}
			}
			if c.visitDeclarations(initializer) {
				return true
			}
		} else if c.visit(initializer) {
			return true
		}
	}
	return c.visit(loop.Condition)
}

// visitDeclarations evaluates the initializers of a variable declaration list.
// The defaults and computed keys of a binding pattern are not searched.
func (c *referenceChecker) visitDeclarations(list *ast.Node) bool {
	for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
		if initializer := declaration.Initializer(); initializer != nil && c.visit(initializer) {
			return true
		}
	}
	return false
}

// lexicalNames returns the names that a list of statements binds in its own
// block scope: let, const, class and function declarations.
func lexicalNames(statements []*ast.Node) []string {
	var names []string
	for _, statement := range statements {
		switch statement.Kind {
		case ast.KindVariableStatement:
			list := statement.AsVariableStatement().DeclarationList
			if list.Flags&(ast.NodeFlagsLet|ast.NodeFlagsConst) == 0 {
				continue
			}
			for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
				names = append(names, bindingNames(declaration.Name())...)
			}
		case ast.KindClassDeclaration, ast.KindFunctionDeclaration:
			if name := statement.Name(); name != nil {
				names = append(names, name.Text())
			}
		}
	}
	return names
}

// hoistedVarNames returns the names declared with var anywhere in block, not
// counting functions and classes, which have scopes of their own.
func hoistedVarNames(block *ast.Node) []string {
	var names []string
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindFunctionDeclaration,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor,
			ast.KindClassExpression, ast.KindClassDeclaration:
			return false
		case ast.KindVariableDeclarationList:
			if node.Flags&ast.NodeFlagsBlockScoped == 0 {
				for _, declaration := range node.AsVariableDeclarationList().Declarations.Nodes {
					names = append(names, bindingNames(declaration.Name())...)
				}
			}
		}
		return node.ForEachChild(visit)
	}
	block.ForEachChild(visit)
	return names
}

// bindingNames returns the names a binding name or pattern declares.
func bindingNames(name *ast.Node) []string {
	if name == nil {
		return nil
	}
	switch name.Kind {
	case ast.KindIdentifier:
		return []string{name.Text()}
	case ast.KindObjectBindingPattern:
		var names []string
		for _, element := range name.AsBindingPattern().Elements.Nodes {
			names = append(names, bindingNames(element.Name())...)
		}
		return names
	case ast.KindArrayBindingPattern:
		var names []string
		for _, element := range name.AsBindingPattern().Elements.Nodes {
			if element.Kind == ast.KindBindingElement {
				names = append(names, bindingNames(element.Name())...)
			}
		}
		return names
	}
	return nil
}
