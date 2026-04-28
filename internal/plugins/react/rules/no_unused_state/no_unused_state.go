package no_unused_state

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// skipAssertionsAndParens strips parentheses and all TS assertion wrappers
// (as, satisfies, !, <T>) from an expression, matching ESLint's
// unwrapTSAsExpression(uncast(node)) pattern.
func skipAssertionsAndParens(node *ast.Node) *ast.Node {
	return ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKAssertions)
}

// getName extracts a static string name from a node. For Identifiers returns
// the text, for StringLiterals returns the text, for NumericLiterals returns
// the string representation, for NoSubstitutionTemplateLiterals returns the
// raw text. Returns "" for everything else.
func getName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	node = skipAssertionsAndParens(node)
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text
	case ast.KindNumericLiteral:
		return utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text)
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text
	case ast.KindTrueKeyword:
		return "true"
	case ast.KindFalseKeyword:
		return "false"
	}
	return ""
}

// isThisExpression checks if a node is a ThisExpression after unwrapping
// type assertions and parentheses.
func isThisExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return skipAssertionsAndParens(node).Kind == ast.KindThisKeyword
}

// isSetStateCall checks if a node is a this.setState(...) call.
func isSetStateCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	callee := skipAssertionsAndParens(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	pa := callee.AsPropertyAccessExpression()
	if !isThisExpression(pa.Expression) {
		return false
	}
	nameNode := pa.Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return false
	}
	return nameNode.AsIdentifier().Text == "setState"
}

// stateField records a state property definition node and its name.
type stateField struct {
	node *ast.Node
	name string
}

// classInfo tracks state field definitions and usages within a component.
type classInfo struct {
	stateFields []*stateField
	usedFields  map[string]bool
	aliases     map[string]bool // per-method/function aliases for this.state
	abandoned   bool
	typeChecker *checker.Checker // may be nil for plain-JS files
}

func newClassInfo(tc *checker.Checker) *classInfo {
	return &classInfo{
		usedFields:  make(map[string]bool),
		typeChecker: tc,
	}
}

// symbolAt returns the symbol at `node` when the TypeChecker is available,
// or nil otherwise.
func (ci *classInfo) symbolAt(node *ast.Node) *ast.Symbol {
	if ci.typeChecker == nil || node == nil {
		return nil
	}
	return ci.typeChecker.GetSymbolAtLocation(node)
}

// addStateFields records all named properties from an ObjectLiteralExpression
// as state field definitions. In ESTree every object member (including
// shorthand methods and getters/setters) is a Property — we match that by
// handling all tsgo member kinds that carry a static key.
func (ci *classInfo) addStateFields(objLit *ast.Node) {
	if objLit == nil || objLit.Kind != ast.KindObjectLiteralExpression {
		return
	}
	ole := objLit.AsObjectLiteralExpression()
	if ole.Properties == nil {
		return
	}
	for _, prop := range ole.Properties.Nodes {
		switch prop.Kind {
		case ast.KindPropertyAssignment:
			nameNode := prop.AsPropertyAssignment().Name()
			if nameNode == nil {
				continue
			}
			if name := getPropertyKeyName(nameNode); name != "" {
				ci.stateFields = append(ci.stateFields, &stateField{node: prop, name: name})
			}
		case ast.KindShorthandPropertyAssignment:
			nameNode := prop.AsShorthandPropertyAssignment().Name()
			if nameNode != nil && nameNode.Kind == ast.KindIdentifier {
				ci.stateFields = append(ci.stateFields, &stateField{node: prop, name: nameNode.AsIdentifier().Text})
			}
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			// ESTree represents shorthand methods and accessors as Property
			// nodes; we must track them for parity.
			nameNode := prop.Name()
			if nameNode == nil {
				continue
			}
			if name := getPropertyKeyName(nameNode); name != "" {
				ci.stateFields = append(ci.stateFields, &stateField{node: prop, name: name})
			}
		}
	}
}

// getPropertyKeyName extracts the static name from a property name node,
// delegating to utils.GetStaticPropertyName. Returns "" for dynamic keys.
func getPropertyKeyName(nameNode *ast.Node) string {
	if nameNode == nil {
		return ""
	}
	name, ok := utils.GetStaticPropertyName(nameNode)
	if !ok {
		return ""
	}
	return name
}

// addUsedStateField marks a state field name as used.
func (ci *classInfo) addUsedStateField(node *ast.Node) {
	if ci == nil || ci.abandoned {
		return
	}
	name := getName(node)
	if name != "" {
		ci.usedFields[name] = true
	}
}

// isStateReference checks if a node refers to this.state, an alias, or a
// lifecycle state parameter.
func (ci *classInfo) isStateReference(node *ast.Node) bool {
	if ci == nil || ci.abandoned {
		return false
	}
	node = skipAssertionsAndParens(node)

	// Direct: this.state
	if node.Kind == ast.KindPropertyAccessExpression {
		pa := node.AsPropertyAccessExpression()
		if isThisExpression(pa.Expression) {
			nameNode := pa.Name()
			if nameNode != nil && nameNode.Kind == ast.KindIdentifier && nameNode.AsIdentifier().Text == "state" {
				return true
			}
		}
	}

	// Alias
	if node.Kind == ast.KindIdentifier && ci.aliases != nil {
		if ci.aliases[node.AsIdentifier().Text] {
			return true
		}
	}

	// Lifecycle state parameter (shouldComponentUpdate, componentDidUpdate, etc.)
	return ci.isStateParameterReference(node)
}

// lifecycleMethodsWithStateParam lists lifecycle methods whose second parameter
// is the previous/next state object.
var lifecycleMethodsWithStateParam = map[string]bool{
	"shouldComponentUpdate":      true,
	"componentWillUpdate":        true,
	"UNSAFE_componentWillUpdate": true,
	"getSnapshotBeforeUpdate":    true,
	"componentDidUpdate":         true,
}

// isStateParameterReference checks if an identifier refers to the state
// parameter of a lifecycle method. When a TypeChecker is available, uses
// symbol identity for precise matching (handles shadowing correctly);
// otherwise falls back to name-based matching.
func (ci *classInfo) isStateParameterReference(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}

	// Try symbol-based matching first (precise, handles shadowing).
	if ci.typeChecker != nil {
		useSym := ci.typeChecker.GetSymbolAtLocation(node)
		if useSym != nil {
			paramSym := findLifecycleStateParamSymbol(node, ci.typeChecker)
			return paramSym != nil && useSym == paramSym
		}
	}

	// Fallback: name-based matching (no TypeChecker or symbol unavailable).
	return isStateParameterReferenceByName(node)
}

// findLifecycleStateParamSymbol walks up from `node` to find the nearest
// enclosing lifecycle method, and returns the symbol of its 2nd parameter.
func findLifecycleStateParamSymbol(node *ast.Node, tc *checker.Checker) *ast.Symbol {
	for p := node.Parent; p != nil; p = p.Parent {
		switch p.Kind {
		case ast.KindMethodDeclaration:
			paramNode := getLifecycleStateParam(p)
			if paramNode != nil {
				return tc.GetSymbolAtLocation(paramNode)
			}
		case ast.KindFunctionExpression:
			if p.Parent != nil && p.Parent.Kind == ast.KindPropertyAssignment {
				paramNode := getES5LifecycleStateParam(p)
				if paramNode != nil {
					return tc.GetSymbolAtLocation(paramNode)
				}
			}
		case ast.KindArrowFunction, ast.KindFunctionDeclaration:
			return nil
		}
	}
	return nil
}

// getLifecycleStateParam returns the 2nd parameter's name node of a class
// method if it is a lifecycle method (or static getDerivedStateFromProps).
func getLifecycleStateParam(method *ast.Node) *ast.Node {
	md := method.AsMethodDeclaration()
	nameNode := md.Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return nil
	}
	methodName := nameNode.AsIdentifier().Text
	isGDSFP := ast.IsStatic(method) && methodName == "getDerivedStateFromProps"
	if !isGDSFP && !lifecycleMethodsWithStateParam[methodName] {
		return nil
	}
	params := method.Parameters()
	if len(params) < 2 {
		return nil
	}
	secondParam := params[1]
	if secondParam.Kind != ast.KindParameter {
		return nil
	}
	paramName := secondParam.AsParameterDeclaration().Name()
	if paramName != nil && paramName.Kind == ast.KindIdentifier {
		return paramName
	}
	return nil
}

// getES5LifecycleStateParam returns the 2nd parameter's name node of an ES5
// component method (FunctionExpression inside PropertyAssignment).
func getES5LifecycleStateParam(fn *ast.Node) *ast.Node {
	pa := fn.Parent.AsPropertyAssignment()
	keyNode := pa.Name()
	if keyNode == nil || keyNode.Kind != ast.KindIdentifier {
		return nil
	}
	if !lifecycleMethodsWithStateParam[keyNode.AsIdentifier().Text] {
		return nil
	}
	params := fn.Parameters()
	if len(params) < 2 {
		return nil
	}
	secondParam := params[1]
	if secondParam.Kind != ast.KindParameter {
		return nil
	}
	paramName := secondParam.AsParameterDeclaration().Name()
	if paramName != nil && paramName.Kind == ast.KindIdentifier {
		return paramName
	}
	return nil
}

// isStateParameterReferenceByName is the name-based fallback for
// isStateParameterReference when no TypeChecker is available.
func isStateParameterReferenceByName(node *ast.Node) bool {
	name := node.AsIdentifier().Text
	for p := node.Parent; p != nil; p = p.Parent {
		switch p.Kind {
		case ast.KindMethodDeclaration:
			paramNode := getLifecycleStateParam(p)
			if paramNode != nil && paramNode.AsIdentifier().Text == name {
				return true
			}
		case ast.KindFunctionExpression:
			if p.Parent != nil && p.Parent.Kind == ast.KindPropertyAssignment {
				paramNode := getES5LifecycleStateParam(p)
				if paramNode != nil && paramNode.AsIdentifier().Text == name {
					return true
				}
			}
		case ast.KindArrowFunction, ast.KindFunctionDeclaration:
			return false
		}
	}
	return false
}

// handleStateDestructuring processes an ObjectBindingPattern that destructures
// this.state, recording used fields and new aliases.
func (ci *classInfo) handleStateDestructuring(pat *ast.Node) {
	if pat == nil || pat.Kind != ast.KindObjectBindingPattern {
		return
	}
	bp := pat.AsBindingPattern()
	if bp.Elements == nil {
		return
	}
	for _, elem := range bp.Elements.Nodes {
		if elem.Kind != ast.KindBindingElement {
			continue
		}
		be := elem.AsBindingElement()
		if be.DotDotDotToken != nil {
			// Rest element — add as alias
			localName := be.Name()
			if localName != nil && localName.Kind == ast.KindIdentifier && ci.aliases != nil {
				ci.aliases[localName.AsIdentifier().Text] = true
			}
			continue
		}
		// Regular property — mark the key as used
		if be.PropertyName != nil {
			ci.addUsedStateField(be.PropertyName)
		} else {
			// Shorthand: { foo } — key is same as local name
			localName := be.Name()
			ci.addUsedStateField(localName)
		}
	}
}

// handleAssignment processes assignment patterns (both VariableDeclaration and
// BinaryExpression assignments) to track state aliases and destructuring.
func (ci *classInfo) handleAssignment(left, right *ast.Node) {
	if ci == nil || ci.abandoned {
		return
	}
	right = skipAssertionsAndParens(right)

	switch left.Kind {
	case ast.KindIdentifier:
		// alias = this.state
		if ci.isStateReference(right) && ci.aliases != nil {
			ci.aliases[left.AsIdentifier().Text] = true
		}
	case ast.KindObjectBindingPattern:
		if ci.isStateReference(right) {
			// const { foo } = this.state
			ci.handleStateDestructuring(left)
		} else if isThisExpression(right) {
			// const { state } = this / const { state: myState } = this
			bp := left.AsBindingPattern()
			if bp.Elements == nil {
				break
			}
			for _, elem := range bp.Elements.Nodes {
				if elem.Kind != ast.KindBindingElement {
					continue
				}
				be := elem.AsBindingElement()
				if be.DotDotDotToken != nil {
					continue
				}
				// Check if the property name is "state"
				propName := ""
				if be.PropertyName != nil {
					propName = getName(be.PropertyName)
				} else {
					propName = getName(be.Name())
				}
				if propName != "state" {
					continue
				}
				localName := be.Name()
				if localName == nil {
					continue
				}
				nameStr := getName(localName)
				if nameStr != "" && ci.aliases != nil {
					// const { state: aliasName } = this
					ci.aliases[nameStr] = true
				} else if localName.Kind == ast.KindObjectBindingPattern {
					// const { state: { foo } } = this
					ci.handleStateDestructuring(localName)
				}
			}
		}
	}
}

// walkES6Component walks an ES6 class component body collecting state
// definitions and usages.
func walkES6Component(ci *classInfo, classNode *ast.Node) {
	members := classNode.Members()
	if members == nil {
		return
	}

	for _, member := range members {
		if ci.abandoned {
			return
		}
		switch member.Kind {
		case ast.KindPropertyDeclaration:
			processPropertyDeclaration(ci, member)
		case ast.KindMethodDeclaration:
			processMethodDeclaration(ci, member)
		case ast.KindConstructor:
			processConstructor(ci, member)
		case ast.KindGetAccessor, ast.KindSetAccessor:
			ci.aliases = make(map[string]bool)
			walkBody(ci, member)
			ci.aliases = nil
		}
	}
}

// processPropertyDeclaration handles class property declarations like
// `state = { ... }`, arrow function class methods, and any other initializer
// that may reference `this.state`.
//
// ESLint's visitor model automatically walks ALL descendants of every class
// member. We must explicitly walk every initializer to match that behavior;
// skipping non-arrow / non-state initializers causes false positives for
// patterns like `myProp = this.state.foo`.
func processPropertyDeclaration(ci *classInfo, member *ast.Node) {
	pd := member.AsPropertyDeclaration()
	nameNode := pd.Name()
	init := skipAssertionsAndParens(pd.Initializer)

	isStatic := ast.IsStatic(member)

	// Check for `state = { ... }` property
	if !isStatic && nameNode != nil && nameNode.Kind == ast.KindIdentifier &&
		nameNode.AsIdentifier().Text == "state" && init != nil &&
		init.Kind == ast.KindObjectLiteralExpression {
		ci.addStateFields(init)
	}

	// Handle static getDerivedStateFromProps as class property (arrow function)
	if isStatic && nameNode != nil && nameNode.Kind == ast.KindIdentifier &&
		nameNode.AsIdentifier().Text == "getDerivedStateFromProps" &&
		init != nil && (init.Kind == ast.KindArrowFunction || init.Kind == ast.KindFunctionExpression) {
		processGDSFPBody(ci, init)
		return
	}

	// Walk the initializer for state usage detection.
	// Non-static arrow functions get their own alias scope (matching ESLint's
	// ClassProperty enter/exit which creates aliases only for ArrowFE).
	// All other initializers are walked with the current (nil) alias scope,
	// which still detects direct `this.state.foo` access.
	if init == nil {
		return
	}
	if !isStatic && init.Kind == ast.KindArrowFunction {
		ci.aliases = make(map[string]bool)
		walkBody(ci, init)
		ci.aliases = nil
	} else {
		walkBody(ci, init)
	}
}

// processMethodDeclaration handles class method declarations.
func processMethodDeclaration(ci *classInfo, member *ast.Node) {
	ci.aliases = make(map[string]bool)
	walkBody(ci, member)
	ci.aliases = nil
}

// processConstructor handles the class constructor, looking for
// `this.state = { ... }` assignments.
func processConstructor(ci *classInfo, member *ast.Node) {
	ci.aliases = make(map[string]bool)
	walkBody(ci, member)
	ci.aliases = nil
}

// processGDSFPBody processes the body of a getDerivedStateFromProps method,
// finding accesses on the second parameter (state). Uses TypeChecker symbol
// identity when available for precise shadowing handling; falls back to
// name-based matching otherwise.
func processGDSFPBody(ci *classInfo, fn *ast.Node) {
	params := fn.Parameters()
	if len(params) < 2 {
		return
	}
	secondParam := params[1]
	if secondParam.Kind != ast.KindParameter {
		return
	}
	paramName := secondParam.AsParameterDeclaration().Name()
	if paramName == nil || paramName.Kind != ast.KindIdentifier {
		return
	}
	stateParamName := paramName.AsIdentifier().Text

	// Resolve the parameter's symbol for precise matching.
	paramSymbol := ci.symbolAt(paramName)

	// isStateParam reports whether `id` refers to the state parameter.
	isStateParam := func(id *ast.Node) bool {
		if id == nil || id.Kind != ast.KindIdentifier {
			return false
		}
		if paramSymbol != nil {
			idSym := ci.symbolAt(id)
			if idSym != nil {
				return idSym == paramSymbol
			}
		}
		// Fallback: name-based matching.
		return id.AsIdentifier().Text == stateParamName
	}

	// Walk the body looking for uses of the state parameter
	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || ci.abandoned {
			return
		}
		if n.Kind == ast.KindPropertyAccessExpression {
			pa := n.AsPropertyAccessExpression()
			obj := skipAssertionsAndParens(pa.Expression)
			if isStateParam(obj) {
				ci.addUsedStateField(pa.Name())
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return ci.abandoned
		})
	}

	var body *ast.Node
	switch fn.Kind {
	case ast.KindArrowFunction:
		body = fn.AsArrowFunction().Body
	case ast.KindFunctionExpression:
		body = fn.AsFunctionExpression().Body
	case ast.KindMethodDeclaration:
		body = fn.AsMethodDeclaration().Body
	}
	if body != nil {
		walk(body)
	}
}

// walkES5Component walks an ES5 component (createReactClass object literal)
// collecting state definitions and usages.
func walkES5Component(ci *classInfo, objLit *ast.Node) {
	ole := objLit.AsObjectLiteralExpression()
	if ole.Properties == nil {
		return
	}

	for _, prop := range ole.Properties.Nodes {
		if ci.abandoned {
			return
		}
		if prop.Kind != ast.KindPropertyAssignment && prop.Kind != ast.KindMethodDeclaration &&
			prop.Kind != ast.KindGetAccessor && prop.Kind != ast.KindSetAccessor {
			continue
		}

		// Extract key name via the generic Name() method (works for all member kinds).
		keyName := ""
		if nameNode := prop.Name(); nameNode != nil && nameNode.Kind == ast.KindIdentifier {
			keyName = nameNode.AsIdentifier().Text
		}

		if keyName == "getInitialState" {
			processGetInitialState(ci, prop)
		} else {
			// ESLint's visitor model walks ALL descendants inside the ES5
			// object. FunctionExpression values get their own alias scope
			// (matching ESLint's FunctionExpression handler). Method/accessor
			// shorthand also gets alias scopes. All other values (arrows,
			// plain expressions) are walked without aliases but still detect
			// direct `this.state.foo` access.
			switch prop.Kind {
			case ast.KindPropertyAssignment:
				init := prop.AsPropertyAssignment().Initializer
				if init != nil && init.Kind == ast.KindFunctionExpression {
					ci.aliases = make(map[string]bool)
					walkBody(ci, init)
					ci.aliases = nil
				} else if init != nil {
					walkBody(ci, init)
				}
			case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
				ci.aliases = make(map[string]bool)
				walkBody(ci, prop)
				ci.aliases = nil
			}
		}
	}
}

// processGetInitialState handles the getInitialState method in ES5 components.
func processGetInitialState(ci *classInfo, prop *ast.Node) {
	var body *ast.Node
	switch prop.Kind {
	case ast.KindPropertyAssignment:
		init := prop.AsPropertyAssignment().Initializer
		if init == nil || init.Kind != ast.KindFunctionExpression {
			return
		}
		body = init.AsFunctionExpression().Body
	case ast.KindMethodDeclaration:
		body = prop.AsMethodDeclaration().Body
	}

	if body == nil || body.Kind != ast.KindBlock {
		return
	}

	block := body.AsBlock()
	if block.Statements == nil || len(block.Statements.Nodes) == 0 {
		return
	}
	stmts := block.Statements.Nodes
	lastStmt := stmts[len(stmts)-1]
	if lastStmt.Kind != ast.KindReturnStatement {
		return
	}
	rs := lastStmt.AsReturnStatement()
	if rs.Expression == nil {
		return
	}
	retVal := skipAssertionsAndParens(rs.Expression)
	if retVal.Kind == ast.KindObjectLiteralExpression {
		ci.addStateFields(retVal)
	}
}

// walkBody recursively walks a function/method body processing state-related
// AST patterns. Stops at nested class boundaries to avoid cross-component
// interference.
func walkBody(ci *classInfo, node *ast.Node) {
	if node == nil || ci.abandoned {
		return
	}

	// Skip nested React components
	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return
	}

	processNode(ci, node)

	if ci.abandoned {
		return
	}

	node.ForEachChild(func(child *ast.Node) bool {
		walkBody(ci, child)
		return ci.abandoned
	})
}

// processNode handles a single node during the body walk.
func processNode(ci *classInfo, node *ast.Node) {
	switch node.Kind {
	case ast.KindCallExpression:
		processCallExpression(ci, node)

	case ast.KindBinaryExpression:
		processAssignmentExpression(ci, node)

	case ast.KindVariableDeclaration:
		processVariableDeclarator(ci, node)

	case ast.KindPropertyAccessExpression:
		processMemberExpression(ci, node)

	case ast.KindElementAccessExpression:
		processElementAccess(ci, node)

	case ast.KindJsxSpreadAttribute:
		// If spreading this.state in JSX, give up
		jsa := node.AsJsxSpreadAttribute()
		if jsa.Expression != nil && ci.isStateReference(jsa.Expression) {
			ci.abandoned = true
		}

	case ast.KindSpreadAssignment:
		// Object spread: { ...this.state }
		sa := node.AsSpreadAssignment()
		if sa.Expression != nil && ci.isStateReference(sa.Expression) {
			ci.abandoned = true
		}

	case ast.KindSpreadElement:
		// Array/call spread: [...this.state]
		se := node.AsSpreadElement()
		if se.Expression != nil && ci.isStateReference(se.Expression) {
			ci.abandoned = true
		}

	}
}

// processCallExpression handles this.setState() calls.
func processCallExpression(ci *classInfo, node *ast.Node) {
	call := node.AsCallExpression()
	unwrappedNode := skipAssertionsAndParens(node)
	if unwrappedNode.Kind != ast.KindCallExpression {
		return
	}
	callExpr := unwrappedNode.AsCallExpression()

	if !isSetStateCall(unwrappedNode) {
		return
	}

	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return
	}
	firstArg := skipAssertionsAndParens(callExpr.Arguments.Nodes[0])

	switch firstArg.Kind {
	case ast.KindObjectLiteralExpression:
		ci.addStateFields(firstArg)
	case ast.KindArrowFunction:
		af := firstArg.AsArrowFunction()
		body := skipAssertionsAndParens(af.Body)
		if body != nil && body.Kind == ast.KindObjectLiteralExpression {
			ci.addStateFields(body)
		}
		// Add first param as alias for state
		params := firstArg.Parameters()
		if len(params) > 0 {
			firstParam := params[0]
			if firstParam.Kind == ast.KindParameter {
				paramName := firstParam.AsParameterDeclaration().Name()
				if paramName != nil {
					if paramName.Kind == ast.KindObjectBindingPattern {
						ci.handleStateDestructuring(paramName)
					} else if paramName.Kind == ast.KindIdentifier && ci.aliases != nil {
						ci.aliases[paramName.AsIdentifier().Text] = true
					}
				}
			}
		}
	}
}

// processAssignmentExpression handles assignment expressions like
// `this.state = {}` and alias assignments.
func processAssignmentExpression(ci *classInfo, node *ast.Node) {
	bin := node.AsBinaryExpression()
	if bin.OperatorToken == nil || !ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
		return
	}
	// Only handle simple assignment (=), not compound (+=, etc.)
	if bin.OperatorToken.Kind != ast.KindEqualsToken {
		return
	}

	left := skipAssertionsAndParens(bin.Left)
	right := skipAssertionsAndParens(bin.Right)

	// Check for `this.state = { ... }`
	if left.Kind == ast.KindPropertyAccessExpression {
		pa := left.AsPropertyAccessExpression()
		if isThisExpression(pa.Expression) {
			nameNode := pa.Name()
			if nameNode != nil && nameNode.Kind == ast.KindIdentifier &&
				nameNode.AsIdentifier().Text == "state" &&
				right.Kind == ast.KindObjectLiteralExpression {
				// Check if we're in a constructor
				if isInConstructor(node) {
					ci.addStateFields(right)
					return
				}
			}
		}
	}

	// Handle alias assignments: alias = this.state, or destructuring
	ci.handleAssignment(left, right)
}

// isInConstructor checks if a node is inside a constructor method.
func isInConstructor(node *ast.Node) bool {
	for p := node.Parent; p != nil; p = p.Parent {
		switch p.Kind {
		case ast.KindConstructor:
			return true
		case ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindFunctionDeclaration:
			// If we hit a function boundary before constructor, not in constructor
			// BUT we need to check if this function IS the constructor value
			if p.Parent != nil && p.Parent.Kind == ast.KindConstructor {
				return true
			}
			return false
		case ast.KindMethodDeclaration:
			return false
		}
	}
	return false
}

// processVariableDeclarator handles variable declarations like
// `const { foo } = this.state` and `const state = this.state`.
func processVariableDeclarator(ci *classInfo, node *ast.Node) {
	vd := node.AsVariableDeclaration()
	if vd.Initializer == nil {
		return
	}
	ci.handleAssignment(vd.Name(), vd.Initializer)
}

// processMemberExpression handles property access expressions like
// `this.state.foo` and `alias.foo`.
func processMemberExpression(ci *classInfo, node *ast.Node) {
	pa := node.AsPropertyAccessExpression()
	obj := skipAssertionsAndParens(pa.Expression)

	if ci.isStateReference(obj) {
		// Record that we saw this property being accessed
		ci.addUsedStateField(pa.Name())
		return
	}

	// If this.state is used as a call argument (not setState), give up
	if ci.isStateReference(node) {
		if node.Parent != nil && node.Parent.Kind == ast.KindCallExpression {
			ci.abandoned = true
		}
	}
}

// processElementAccess handles element access expressions like
// `this.state['foo']` and `this.state[expr]`.
func processElementAccess(ci *classInfo, node *ast.Node) {
	ea := node.AsElementAccessExpression()
	obj := skipAssertionsAndParens(ea.Expression)

	if ci.isStateReference(obj) {
		argExpr := ea.ArgumentExpression
		if argExpr == nil {
			return
		}
		argExpr = skipAssertionsAndParens(argExpr)
		// If the access key is a static literal, record the used field.
		// In ESTree, true/false/null are Literals, so ESLint's
		// `node.property.type !== 'Literal'` check does NOT give up on them.
		switch argExpr.Kind {
		case ast.KindStringLiteral:
			ci.usedFields[argExpr.AsStringLiteral().Text] = true
		case ast.KindNumericLiteral:
			ci.usedFields[utils.NormalizeNumericLiteral(argExpr.AsNumericLiteral().Text)] = true
		case ast.KindNoSubstitutionTemplateLiteral:
			ci.usedFields[argExpr.AsNoSubstitutionTemplateLiteral().Text] = true
		case ast.KindTrueKeyword:
			ci.usedFields["true"] = true
		case ast.KindFalseKeyword:
			ci.usedFields["false"] = true
		case ast.KindNullKeyword:
			ci.usedFields["null"] = true
		default:
			// Dynamic computed access — give up
			ci.abandoned = true
		}
	}
}

var NoUnusedStateRule = rule.Rule{
	Name: "react/no-unused-state",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)

		processComponent := func(ci *classInfo) {
			if ci.abandoned {
				return
			}
			for _, field := range ci.stateFields {
				if !ci.usedFields[field.name] {
					ctx.ReportNode(field.node, rule.RuleMessage{
						Id:          "unusedStateField",
						Description: fmt.Sprintf("Unused state field: '%s'", field.name),
					})
				}
			}
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				if !reactutil.ExtendsReactComponent(node, pragma) {
					return
				}
				ci := newClassInfo(ctx.TypeChecker)
				walkES6Component(ci, node)
				processComponent(ci)
			},
			ast.KindClassExpression: func(node *ast.Node) {
				if !reactutil.ExtendsReactComponent(node, pragma) {
					return
				}
				ci := newClassInfo(ctx.TypeChecker)
				walkES6Component(ci, node)
				processComponent(ci)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if !reactutil.IsCreateClassCall(call, pragma, createClass) {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				arg := ast.SkipParentheses(call.Arguments.Nodes[0])
				if arg.Kind != ast.KindObjectLiteralExpression {
					return
				}
				ci := newClassInfo(ctx.TypeChecker)
				walkES5Component(ci, arg)
				processComponent(ci)
			},
		}
	},
}
