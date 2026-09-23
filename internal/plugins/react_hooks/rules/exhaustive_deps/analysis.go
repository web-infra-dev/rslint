package exhaustive_deps

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/collections"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/react_hooksutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

// The shared scope graph supplies lexical bindings, including parameter
// defaults and block scopes, without requiring a TypeChecker. Build it only
// after encountering a reactive hook that needs scope analysis.
type runCaches struct {
	ctx               *rule.RuleContext
	manager           *scope.Manager
	children          map[*scope.Scope][]*scope.Scope
	references        map[*scope.Variable][]*scope.Reference
	byIdentifier      map[*ast.Node]*scope.Reference
	stable            map[*scope.Variable]bool
	functions         map[*scope.Variable]bool
	setStateCallSites map[*ast.Node]*ast.Node
	stateVariables    map[*ast.Node]bool
	effectEvents      map[*ast.Node]bool
}

func (c *runCaches) scopes() *scope.Manager {
	if c.manager != nil {
		return c.manager
	}
	c.manager = scopeanalysis.References(*c.ctx, nil)
	c.children = make(map[*scope.Scope][]*scope.Scope)
	c.references = make(map[*scope.Variable][]*scope.Reference)
	c.byIdentifier = make(map[*ast.Node]*scope.Reference, len(c.manager.References))
	c.stable = make(map[*scope.Variable]bool)
	c.functions = make(map[*scope.Variable]bool)
	c.setStateCallSites = make(map[*ast.Node]*ast.Node)
	c.stateVariables = make(map[*ast.Node]bool)
	c.effectEvents = make(map[*ast.Node]bool)
	for _, s := range c.manager.Scopes {
		c.children[s.Parent] = append(c.children[s.Parent], s)
	}
	for _, ref := range c.manager.References {
		c.byIdentifier[ref.Identifier] = ref
		if v := ref.Resolved(); v != nil {
			c.references[v] = append(c.references[v], ref)
		}
	}
	return c.manager
}

func firstVariable(s *scope.Scope, name string) *scope.Variable {
	if vars := s.ByName[name]; len(vars) > 0 {
		return vars[0]
	}
	return nil
}

func variableDeclaration(v *scope.Variable) *ast.Node {
	if v == nil || v.Kind != scope.DefVariable {
		return nil
	}
	d := v.DefNode
	if d.Kind == ast.KindBindingElement {
		d = utils.EnclosingVariableDeclarationOfBindingElement(d)
	}
	if d != nil && d.Kind == ast.KindVariableDeclaration {
		return d
	}
	return nil
}

func (c *runCaches) stableValue(v *scope.Variable) bool {
	if result, ok := c.stable[v]; ok {
		return result
	}
	result := c.computeStableValue(v)
	c.stable[v] = result
	return result
}

func (c *runCaches) computeStableValue(v *scope.Variable) bool {
	decl := variableDeclaration(v)
	if decl == nil {
		return false
	}
	vd := decl.AsVariableDeclaration()
	init := stripAsExpression(vd.Initializer)
	if init == nil {
		return false
	}
	if decl.Parent.Kind == ast.KindVariableDeclarationList && decl.Parent.Flags&ast.NodeFlagsConst != 0 {
		switch init.Kind {
		case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNullKeyword:
			return true
		}
	}
	if init.Kind != ast.KindCallExpression || ast.IsOptionalChain(init) {
		return false
	}
	callee := react_hooksutil.StripReactNamespace(utils.ESTreeCallCallee(init.AsCallExpression().Expression))
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return false
	}
	name := callee.Text()
	id := vd.Name()
	if id.Kind == ast.KindIdentifier {
		if name == "useRef" {
			return true
		}
		if name == "useEffectEvent" {
			for _, ref := range c.references[v] {
				c.effectEvents[ref.Identifier] = true
			}
			return true
		}
		return false
	}
	if id.Kind != ast.KindArrayBindingPattern || len(id.AsBindingPattern().Elements.Nodes) != 2 {
		return false
	}
	elements := id.AsBindingPattern().Elements.Nodes
	second := simpleBindingIdentifier(elements[1])
	first := simpleBindingIdentifier(elements[0])
	switch name {
	case "useState", "useReducer", "useActionState", "useTransition":
		if second == v.ID {
			if name == "useState" {
				// ESLint counts the initializer as a write. Subsequent references
				// are visited in source order, including writes in nested scopes.
				writes := 0
				declarations := v.Scope.Declarations(v.Name)
				declarationIndex := 0
				countDeclarations := func(before int) {
					for declarationIndex < len(declarations) && declarations[declarationIndex].ID.Pos() < before {
						d := variableDeclaration(declarations[declarationIndex])
						if d != nil && (d.AsVariableDeclaration().Initializer != nil || utils.IsVarDeclInForInOrOf(d)) {
							writes++
						}
						declarationIndex++
					}
				}
				for _, ref := range c.references[v] {
					countDeclarations(ref.Identifier.Pos())
					if utils.IsWriteReference(ref.Identifier) {
						writes++
					}
					if writes > 1 {
						return false
					}
					state := elements[0]
					if first != nil {
						state = first
					} else if state.Kind != ast.KindBindingElement || state.AsBindingElement().Name() == nil {
						state = nil
					}
					c.setStateCallSites[ref.Identifier] = state
				}
				countDeclarations(c.ctx.SourceFile.AsNode().End() + 1)
				if writes > 1 {
					return false
				}
			}
			return true
		}
		if name == "useState" && first == v.ID {
			for _, ref := range c.references[v] {
				c.stateVariables[ref.Identifier] = true
			}
		}
	}
	return false
}

func simpleBindingIdentifier(element *ast.Node) *ast.Node {
	if element == nil || element.Kind != ast.KindBindingElement {
		return nil
	}
	b := element.AsBindingElement()
	if b.Initializer != nil || b.DotDotDotToken != nil || b.Name() == nil || b.Name().Kind != ast.KindIdentifier {
		return nil
	}
	return b.Name()
}

func withinScope(s, ancestor *scope.Scope) bool {
	for ; s != nil; s = s.Parent {
		if s == ancestor {
			return true
		}
	}
	return false
}

func (c *runCaches) functionWithoutCaptures(v *scope.Variable, component *scope.Scope, pure map[*scope.Scope]bool) bool {
	if result, ok := c.functions[v]; ok {
		return result
	}
	c.functions[v] = false
	var fnScope *scope.Scope
	for _, child := range c.children[component] {
		if v.Kind == scope.DefFunctionName && child.Block == v.DefNode ||
			v.Kind == scope.DefVariable && child.Block.Parent == v.DefNode {
			fnScope = child
			break
		}
	}
	if fnScope == nil {
		return false
	}
	var captures func(*scope.Scope) bool
	captures = func(s *scope.Scope) bool {
		for _, ref := range s.References {
			target := ref.Resolved()
			if target != nil && !withinScope(target.Scope, fnScope) && pure[target.Scope] && !c.stableValue(target) {
				return true
			}
		}
		for _, child := range c.children[s] {
			if captures(child) {
				return true
			}
		}
		return false
	}
	result := !captures(fnScope)
	c.functions[v] = result
	return result
}

func (c *runCaches) gather(callback *ast.Node, callbackScope, component *scope.Scope, pure map[*scope.Scope]bool, isEffect bool) (*dependencyMap, map[string]bool) {
	dependencies := &dependencyMap{}
	optionalChains := make(map[string]bool)
	cleanups := collections.OrderedMap[string, depReference]{}
	var visit func(*scope.Scope)
	visit = func(s *scope.Scope) {
		for _, ref := range s.References {
			v := ref.Resolved()
			if v == nil || !pure[v.Scope] {
				continue
			}
			id := ref.Identifier
			root := getDependencyNode(id)
			key, ok := analyzePropertyChainText(root, optionalChains)
			if !ok {
				continue
			}
			record := depReference{Reference: ref, WriteExpr: writeExpression(id), DepNodeRoot: root}
			parent := utils.ESTreeParent(root)
			if isEffect && utils.ESTreeRuntimeExpression(root).Kind == ast.KindIdentifier && parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.Name().Text() == "current" {
				inReturnedFunction := false
				for current := s; current != nil && current != callbackScope; current = current.Parent {
					if current.Kind == scope.KindFunction {
						p := utils.ESTreeParent(current.Block)
						inReturnedFunction = p != nil && p.Kind == ast.KindReturnStatement
					}
				}
				if inReturnedFunction {
					cleanups.Set(key, record)
				}
			}
			if isInsideTypePosition(id) {
				continue
			}
			if decl := variableDeclaration(v); decl != nil && decl.AsVariableDeclaration().Initializer != nil && utils.ESTreeRuntimeExpression(decl.AsVariableDeclaration().Initializer) == utils.ESTreeParent(callback) {
				continue
			}
			dep := dependencies.GetOrZero(key)
			if dep == nil {
				dep = &dependency{IsStable: c.stableValue(v) || c.functionWithoutCaptures(v, component, pure)}
				dependencies.Set(key, dep)
			}
			dep.Refs = append(dep.Refs, record)
		}
		for _, child := range c.children[s] {
			visit(child)
		}
	}
	visit(callbackScope)
	for key, ref := range cleanups.Entries() {
		assigned := false
		for _, candidate := range c.references[ref.Resolved()] {
			member := utils.ESTreeParent(candidate.Identifier)
			if member != nil && member.Kind == ast.KindPropertyAccessExpression && member.Name().Text() == "current" {
				if assignment, ok := getAssignmentBinaryExpr(utils.ESTreeParent(member)); ok && utils.ESTreeRuntimeExpression(assignment.Left) == member {
					assigned = true
					break
				}
			}
		}
		if assigned {
			continue
		}
		c.ctx.ReportNode(utils.ESTreeParent(ref.DepNodeRoot).Name(), rule.RuleMessage{Description: "The ref value '" + key + ".current' will likely have changed by the time this effect cleanup function runs. " +
			"If this ref points to a node rendered by React, copy '" + key + ".current' to a variable inside the effect, and use that variable in the cleanup function."})
	}
	return dependencies, optionalChains
}

// ESLint's writeExpr is the assigned value, and is absent for ++/--.
// The compiler already identifies assignment targets through destructuring.
func writeExpression(id *ast.Node) *ast.Node {
	if id.Parent == nil || !utils.IsWriteReference(id) {
		return nil
	}

	assignmentID := id
	for assignmentID.Parent != nil && ast.IsOuterExpression(assignmentID.Parent, ast.OEKParentheses|ast.OEKAssertions) {
		assignmentID = assignmentID.Parent
	}
	target := ast.GetAssignmentTarget(assignmentID)
	// A parenthesized destructuring target remains an expression in
	// typescript-eslint, so its identifiers are reads rather than writes.
	for current := id; current != nil && current != target; current = current.Parent {
		if (current.Kind == ast.KindObjectLiteralExpression || current.Kind == ast.KindArrayLiteralExpression) && current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
			return nil
		}
	}

	if target == nil {
		return nil
	}
	if id.Parent.Kind == ast.KindShorthandPropertyAssignment {
		if init := id.Parent.AsShorthandPropertyAssignment().ObjectAssignmentInitializer; init != nil {
			return init
		}
	}
	switch target.Kind {
	case ast.KindBinaryExpression:
		return utils.ESTreeRuntimeExpression(target.AsBinaryExpression().Right)
	case ast.KindForInStatement, ast.KindForOfStatement:
		return utils.ESTreeRuntimeExpression(target.AsForInOrOfStatement().Expression)
	}
	return nil
}

func (c *runCaches) usedOutside(v *scope.Variable, callbackScope *scope.Scope, deps *ast.Node) bool {
	for _, ref := range c.references[v] {
		if writeExpression(ref.Identifier) != nil || !withinScope(ref.From, callbackScope) && !containsNode(deps, ref.Identifier) {
			return true
		}
	}
	return false
}
