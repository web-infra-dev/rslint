package no_restricted_syntax

import (
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// matchContext threads the source file through the matcher so attribute
// path resolution can reach the original source text (for attributes like
// `regex.flags` or `source.value` that depend on the lexed form).
type matchContext struct {
	sf *ast.SourceFile
}

// kindSet is the result of computing the set of tsgo kinds a selector might
// match. universe == true means "match every node, regardless of kind" —
// used to avoid materializing the full set of kinds for `*` and pure
// attribute selectors.
type kindSet struct {
	kinds    map[ast.Kind]struct{}
	universe bool
}

func newKindSet() *kindSet {
	return &kindSet{kinds: make(map[ast.Kind]struct{})}
}

func (s *kindSet) addAll(ks []ast.Kind) {
	if s.universe {
		return
	}
	for _, k := range ks {
		s.kinds[k] = struct{}{}
	}
}

func (s *kindSet) markUniverse() {
	s.universe = true
	s.kinds = nil
}

// candidateKinds returns the set of tsgo ast.Kind values that a selector
// might match. For selectors that constrain by attribute / pseudo only
// (no leading kind), the universe flag is set and the listener should be
// registered on every kind in allInterestingKinds.
func candidateKinds(sel selector) *kindSet {
	s := newKindSet()
	collectKinds(sel, s)
	return s
}

func collectKinds(sel selector, s *kindSet) {
	switch v := sel.(type) {
	case subjectSelector:
		collectKinds(v.Inner, s)
	case identifierSelector:
		if v.Name == "*" {
			s.markUniverse()
			return
		}
		ks := v.Kinds
		if ks == nil {
			// Unknown ESTree name — collapse to the empty set so that this
			// selector is never registered. The user's selector simply
			// doesn't apply to any tsgo node.
			return
		}
		s.addAll(ks)
	case classSelector:
		collectKinds(v.Inner, s)
	case attrSelector:
		collectKinds(v.Inner, s)
	case combinatorSelector:
		// The right-hand side is the node being matched; the left side
		// is an ancestor / sibling constraint we evaluate at match time.
		collectKinds(v.Right, s)
	case relativeSelector:
		collectKinds(v.Inner, s)
	case pseudoSelector:
		switch v.Name {
		case "is", "matches":
			for _, a := range v.Args {
				collectKinds(a, s)
			}
		case "not":
			// `:not(...)` says nothing about the kind that matches — every
			// kind could conceivably match. Mark universe.
			s.markUniverse()
		case "has":
			s.markUniverse()
		case "nth-child", "nth-last-child":
			s.markUniverse()
		case "statement", "expression", "declaration", "function", "pattern":
			s.markUniverse()
		}
	case combinedPseudo:
		collectKinds(v.Inner, s)
	case unionSelector:
		for _, a := range v.Selectors {
			collectKinds(a, s)
		}
	}
}

// matches reports whether sel matches the supplied tsgo AST node.
func matches(sel selector, node *ast.Node, mc *matchContext) bool {
	return matchesInScope(sel, node, mc, nil)
}

func matchesInScope(sel selector, node *ast.Node, mc *matchContext, scopeRoot *virtualNodeFacade) bool {
	return matchesInScopeTarget(sel, node, mc, scopeRoot, "physical")
}

// matchesInScopeTarget evaluates the selector against either the physical
// tsgo node (target == "physical") or an ESTree facade on that same node.
// Structural matching carries both the node and its identity through each edge.
func matchesInScopeTarget(sel selector, node *ast.Node, mc *matchContext, scopeRoot *virtualNodeFacade, target string) bool {
	switch v := sel.(type) {
	case subjectSelector:
		return matchesInScopeTarget(v.Inner, node, mc, scopeRoot, target)
	case identifierSelector:
		if target == "physical" {
			if isVirtualOnlyIdentity(v.Name) {
				return false
			}
			return matchesIdentifier(v, node)
		}
		if target != "" {
			return v.Name == "*" || v.Name == target
		}
		return matchesIdentifier(v, node)
	case classSelector:
		if !matchesInScopeTarget(v.Inner, node, mc, scopeRoot, target) {
			return false
		}
		return matchesClassTarget(node, v.Path, mc, scopeRoot, target)
	case attrSelector:
		if !matchesInScopeTarget(v.Inner, node, mc, scopeRoot, target) {
			return false
		}
		if target != "" && target != "physical" {
			return matchesVirtualAttr(node, target, v.Path, v.Op, v.Value, mc)
		}
		return matchesAttr(node, v.Path, v.Op, v.Value, mc)
	case combinatorSelector:
		if v.Kind != combAdjacent && v.Kind != combSibling && !matchesInScopeTarget(v.Right, node, mc, scopeRoot, target) {
			return false
		}
		return matchesCombinatorTarget(v, node, mc, scopeRoot, target)
	case relativeSelector:
		// Relative selectors are meaningful only as arguments to :has(),
		// where the node on the left is the :has() subject.
		return false
	case pseudoSelector:
		return matchesPseudoTarget(v, node, mc, scopeRoot, target)
	case combinedPseudo:
		if !matchesInScopeTarget(v.Inner, node, mc, scopeRoot, target) {
			return false
		}
		return matchesPseudoTarget(v.Pseudo, node, mc, scopeRoot, target)
	case unionSelector:
		for _, a := range v.Selectors {
			if matchesInScopeTarget(a, node, mc, scopeRoot, target) {
				return true
			}
		}
		return false
	}
	return false
}

// matchesIdentifier evaluates the bare type-name portion of a selector.
// "*" matches everything; ESTree-mapped names match their tsgo kinds with
// extra refinement for kinds that fuse multiple ESTree shapes.
func matchesIdentifier(sel identifierSelector, node *ast.Node) bool {
	if sel.Name == "*" {
		return true
	}
	ks := sel.Kinds
	if ks == nil {
		return false
	}
	matchedKind := false
	for _, k := range ks {
		if node.Kind == k {
			matchedKind = true
			break
		}
	}
	if !matchedKind {
		return false
	}
	return refineEstreeMatch(sel.Name, node)
}

// refineEstreeMatch tightens the tsgo→ESTree match for kinds that tsgo
// fuses but ESTree splits. For example, tsgo's BinaryExpression covers the
// ESTree triplet of BinaryExpression / LogicalExpression / AssignmentExpression
// / SequenceExpression — the operator decides which ESTree form it really is.
func refineEstreeMatch(name string, node *ast.Node) bool {
	switch name {
	case "Identifier":
		return node.Kind != ast.KindConstructor && !isJSXIdentifier(node)
	case "JSXIdentifier":
		return isJSXIdentifier(node)
	case "MemberExpression":
		return !isJSXMemberExpression(node)
	case "JSXMemberExpression":
		return isJSXMemberExpression(node)
	case "JSXExpressionContainer":
		return node.Kind == ast.KindJsxExpression && node.AsJsxExpression().DotDotDotToken == nil
	case "JSXEmptyExpression":
		return node.Kind == ast.KindJsxExpression && node.AsJsxExpression().DotDotDotToken == nil && node.AsJsxExpression().Expression == nil
	case "JSXElement":
		return node.Kind == ast.KindJsxElement || node.Kind == ast.KindJsxSelfClosingElement
	case "JSXOpeningElement":
		return node.Kind == ast.KindJsxOpeningElement
	case "Literal":
		return node.Kind != ast.KindConstructor
	case "FunctionExpression":
		// Methods expose a synthetic FunctionExpression through ESTree's
		// MethodDefinition.value / Property.value fields.
		return node.Kind == ast.KindFunctionExpression
	case "TSEnumBody":
		return false
	case "JSXSpreadChild":
		return node.Kind == ast.KindJsxExpression && node.AsJsxExpression().DotDotDotToken != nil
	case "VariableDeclaration":
		return node.Kind == ast.KindVariableStatement ||
			node.Kind == ast.KindVariableDeclarationList && (node.Parent == nil || node.Parent.Kind != ast.KindVariableStatement)
	case "CallExpression":
		return node.Kind == ast.KindCallExpression &&
			node.AsCallExpression().Expression != nil &&
			node.AsCallExpression().Expression.Kind != ast.KindImportKeyword
	case "ImportExpression":
		return node.Kind == ast.KindCallExpression &&
			node.AsCallExpression().Expression != nil &&
			node.AsCallExpression().Expression.Kind == ast.KindImportKeyword
	case "BinaryExpression":
		return node.Kind == ast.KindBinaryExpression && isPlainBinaryOperator(node)
	case "LogicalExpression":
		return node.Kind == ast.KindBinaryExpression && isLogicalOperator(node)
	case "AssignmentExpression":
		return node.Kind == ast.KindBinaryExpression && isAssignmentOperator(node)
	case "SequenceExpression":
		return node.Kind == ast.KindBinaryExpression && isCommaOperator(node)
	case "ChainExpression":
		return isChainRoot(node)
	case "UnaryExpression":
		// ESTree's UnaryExpression wraps `+`, `-`, `!`, `~`, `typeof`, `void`,
		// `delete`. ESTree's UpdateExpression wraps `++`/`--`. tsgo splits
		// `++`/`--` into PrefixUnary/PostfixUnary alongside other prefix
		// operators, so we filter on the operator token.
		switch node.Kind {
		case ast.KindTypeOfExpression, ast.KindVoidExpression, ast.KindDeleteExpression:
			return true
		case ast.KindPrefixUnaryExpression:
			op := node.AsPrefixUnaryExpression().Operator
			return op != ast.KindPlusPlusToken && op != ast.KindMinusMinusToken
		}
		return false
	case "UpdateExpression":
		switch node.Kind {
		case ast.KindPostfixUnaryExpression:
			return true
		case ast.KindPrefixUnaryExpression:
			op := node.AsPrefixUnaryExpression().Operator
			return op == ast.KindPlusPlusToken || op == ast.KindMinusMinusToken
		}
		return false
	case "RestElement":
		// ESTree's RestElement covers two tsgo shapes:
		//   - BindingElement with `...` inside an array/object pattern
		//   - Parameter with `...` in a function parameter list
		switch node.Kind {
		case ast.KindBindingElement:
			return node.AsBindingElement().DotDotDotToken != nil
		case ast.KindParameter:
			return node.AsParameterDeclaration().DotDotDotToken != nil
		}
		return false
	case "Property":
		// ESTree distinguishes object-literal members (Property) from
		// class members (MethodDefinition). tsgo fuses them. A method,
		// getter or setter is only a "Property" when it sits inside an
		// ObjectLiteralExpression.
		switch node.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			return node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression
		}
		return true
	case "TSAbstractMethodDefinition":
		return isClassLikeNode(node.Parent) && ast.HasSyntacticModifier(node, ast.ModifierFlagsAbstract)
	case "MethodDefinition":
		// Mirror of "Property": MethodDefinition only fires for methods
		// / accessors inside a class body. KindConstructor is class-only
		// by construction, so always passes.
		switch node.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			if node.Parent == nil {
				return false
			}
			pk := node.Parent.Kind
			return (pk == ast.KindClassDeclaration || pk == ast.KindClassExpression) && !ast.HasSyntacticModifier(node, ast.ModifierFlagsAbstract)
		case ast.KindConstructor:
			return true
		}
		return false
	case "AssignmentPattern":
		// ESTree's AssignmentPattern covers default-value bindings:
		//   - BindingElement `{ a = 1 }` / `[a = 1]`
		//   - Parameter `function f(a = 1) {}`
		switch node.Kind {
		case ast.KindBindingElement:
			be := node.AsBindingElement()
			return be.Initializer != nil && be.DotDotDotToken == nil
		case ast.KindParameter:
			pd := node.AsParameterDeclaration()
			return pd.Initializer != nil && pd.DotDotDotToken == nil
		}
		return false
	case "SpreadElement":
		// ESTree differentiates spread-in-array from spread-in-object
		// (the latter is SpreadElement in arrays, SpreadElement in calls,
		// and `Property { kind: 'init' }`-like spread in objects). tsgo
		// uses two kinds. Both map to ESTree's SpreadElement here so we
		// accept both without further refinement.
		return node.Kind == ast.KindSpreadElement || node.Kind == ast.KindSpreadAssignment
	}
	return true
}

func isJSXMemberExpression(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindPropertyAccessExpression && current.Parent.AsPropertyAccessExpression().Expression == current {
		current = current.Parent
	}
	return ast.IsJsxTagName(current)
}

func isJSXIdentifier(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier || node.Parent == nil {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindJsxOpeningElement, ast.KindJsxClosingElement, ast.KindJsxSelfClosingElement:
		return node.Parent.TagName() == node
	case ast.KindJsxAttribute:
		return node.Parent.Name() == node
	case ast.KindJsxNamespacedName:
		return true
	case ast.KindPropertyAccessExpression:
		return isJSXMemberExpression(node.Parent)
	}
	return false
}

func isPlainBinaryOperator(node *ast.Node) bool {
	op := node.AsBinaryExpression().OperatorToken.Kind
	if isAssignmentOperatorKind(op) {
		return false
	}
	if isLogicalOperatorKind(op) {
		return false
	}
	if op == ast.KindCommaToken {
		return false
	}
	return true
}

func isLogicalOperator(node *ast.Node) bool {
	return isLogicalOperatorKind(node.AsBinaryExpression().OperatorToken.Kind)
}

func isAssignmentOperator(node *ast.Node) bool {
	return isAssignmentOperatorKind(node.AsBinaryExpression().OperatorToken.Kind)
}

func isCommaOperator(node *ast.Node) bool {
	return node.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken
}

func isLogicalOperatorKind(k ast.Kind) bool {
	switch k {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
		return true
	}
	return false
}

func isAssignmentOperatorKind(k ast.Kind) bool {
	switch k {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken,
		ast.KindBarBarEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// matchesClassTarget evaluates `Foo.bar` — Foo already matched by the inner
// selector, here we check that the node sits at the named field of its parent.
func matchesClassTarget(node *ast.Node, path []string, mc *matchContext, scopeRoot *virtualNodeFacade, target string) bool {
	if len(path) == 0 {
		return false
	}
	ancestor, ancestorTarget := node, target
	for range path {
		var ok bool
		ancestor, ancestorTarget, ok = logicalParentTarget(ancestor, ancestorTarget, scopeRoot)
		if !ok {
			return false
		}
	}
	return nodeIsAtFieldPath(node, target, targetValue(ancestor, ancestorTarget), path, mc)
}

func isVirtualOnlyIdentity(name string) bool {
	switch name {
	case "ClassBody", "TSEnumBody", "JSXEmptyExpression", "TSEmptyBodyFunctionExpression", "ChainExpression":
		return true
	}
	return false
}

// matchesAttr resolves the attribute path against the node and compares
// the obtained value with the operator/right-hand side of the selector.
func matchesAttr(node *ast.Node, path []string, op attrOp, value attrValue, mc *matchContext) bool {
	if (op == attrEqual || op == attrNotEqual) &&
		(value.Kind == attrValueString || value.Kind == attrValueIdent || value.Kind == attrValueRegex) {
		if text, ok, handled := lookupStringAttrPath(node, path, mc); handled {
			if !ok {
				return compareAttr(undefinedAttr{}, op, value)
			}
			return compareStringAttr(text, op, value)
		}
	}

	val, ok := lookupAttrPath(node, path, mc)
	if op == attrPresent {
		// esquery defines presence as a reachable, non-null property. Values
		// such as false, 0, "", and empty arrays still count as present.
		return ok && attrIsPresent(val)
	}
	if !ok {
		val = undefinedAttr{}
	}
	return compareAttr(val, op, value)
}

func matchesVirtualAttr(node *ast.Node, target string, path []string, op attrOp, value attrValue, mc *matchContext) bool {
	val, ok := lookupVirtualAttrPath(node, target, path, mc)
	if op == attrPresent {
		return ok && attrIsPresent(val)
	}
	if !ok {
		val = undefinedAttr{}
	}
	return compareAttr(val, op, value)
}

func lookupVirtualAttrPath(node *ast.Node, target string, path []string, mc *matchContext) (interface{}, bool) {
	if node == nil || len(path) == 0 {
		return nil, false
	}
	var current interface{} = virtualNodeFacade{node: node, typeName: target}
	for _, segment := range path {
		next, ok := stepAttrPath(current, segment, mc)
		if !ok {
			return nil, false
		}
		current = next
	}
	return current, true
}

func attrIsPresent(value interface{}) bool {
	if value == nil {
		return false
	}
	// A typed nil pointer stored in an interface is not equal to nil. AST
	// optional fields commonly take that shape, but ESTree/esquery still sees
	// them as null and therefore absent.
	if node, ok := value.(*ast.Node); ok {
		return node != nil
	}
	return true
}

// lookupStringAttrPath avoids boxing the overwhelmingly common string-valued
// selector paths (for example object.name, callee.property.name, and operator).
// handled is false when the path reaches a non-string ESTree facade, in which
// case the generic interface-based resolver retains the exact behavior.
func lookupStringAttrPath(node *ast.Node, path []string, mc *matchContext) (text string, ok bool, handled bool) {
	if len(path) == 0 {
		return "", false, false
	}

	var current interface{} = node
	for index, segment := range path {
		if index != len(path)-1 {
			next, found := stepAttrPath(current, segment, mc)
			if !found {
				return "", false, true
			}
			current = next
			continue
		}

		switch value := current.(type) {
		case *ast.Node:
			value = unwrapEstreeNode(value)
			if value == nil {
				// A present ESTree property whose value is null remains null when
				// a deeper path is requested; use the generic resolver to retain
				// that distinction from a missing (undefined) property.
				return "", false, false
			}
			switch segment {
			case "name":
				switch value.Kind {
				case ast.KindIdentifier:
					return value.AsIdentifier().Text, true, true
				case ast.KindPrivateIdentifier:
					return strings.TrimPrefix(value.AsPrivateIdentifier().Text, "#"), true, true
				}
				// JSX name attributes resolve to nodes rather than strings, so
				// let the generic resolver preserve that facade's semantics.
				return "", false, false
			case "operator":
				text, found := readOperatorAttrString(value)
				return text, found, true
			case "type":
				if text, ok := readNodeAttr(value, "type", mc); ok {
					if text, ok := text.(string); ok {
						return text, true, true
					}
				}
				return "", false, true
			}
		case metaIdentifier:
			if segment == "name" {
				return value.name, true, true
			}
		}
		return "", false, false
	}
	return "", false, false
}

// matchesCombinatorTarget handles parents, ancestors, and array siblings.
// Sibling subjects also permit esquery's reverse matching direction.
func matchesCombinatorTarget(c combinatorSelector, node *ast.Node, mc *matchContext, scopeRoot *virtualNodeFacade, target string) bool {
	switch c.Kind {
	case combChild, combDescendant:
		parent, parentTarget, ok := logicalParentTarget(node, target, scopeRoot)
		if !ok {
			return false
		}
		if target == "physical" && isDefaultExportedDeclaration(node) && selectorMatchesVirtualExportDefault(c.Left) {
			return true
		}
		for ok {
			if matchesInScopeTarget(c.Left, parent, mc, scopeRoot, parentTarget) {
				return true
			}
			if c.Kind == combChild {
				break
			}
			parent, parentTarget, ok = logicalParentTarget(parent, parentTarget, scopeRoot)
		}
	case combAdjacent, combSibling:
		siblings, idx := siblingsOf(node, target, scopeRoot)
		if idx < 0 {
			return false
		}
		if idx > 0 && matchesInScopeTarget(c.Right, node, mc, scopeRoot, target) {
			if c.Kind == combAdjacent && matchesInScopeTarget(c.Left, unwrapEstreeNode(siblings[idx-1]), mc, scopeRoot, childTarget(siblings[idx-1])) {
				return true
			}
			if c.Kind == combSibling {
				for i := idx - 1; i >= 0; i-- {
					if matchesInScopeTarget(c.Left, unwrapEstreeNode(siblings[i]), mc, scopeRoot, childTarget(siblings[i])) {
						return true
					}
				}
			}
		}
		// These reverse-direction branches mirror esquery 1.7.0's subject
		// handling. Only nodes in the same ESTree array can be siblings.
		if c.Kind == combSibling && isSubjectSelector(c.Left) && matchesInScopeTarget(c.Left, node, mc, scopeRoot, target) {
			for i := idx + 1; i < len(siblings); i++ {
				if matchesInScopeTarget(c.Right, unwrapEstreeNode(siblings[i]), mc, scopeRoot, childTarget(siblings[i])) {
					return true
				}
			}
		}
		if c.Kind == combAdjacent && isSubjectSelector(c.Right) && matchesInScopeTarget(c.Left, node, mc, scopeRoot, target) && idx+1 < len(siblings) {
			return matchesInScopeTarget(c.Right, unwrapEstreeNode(siblings[idx+1]), mc, scopeRoot, childTarget(siblings[idx+1]))
		}
	}
	return false
}

func isSubjectSelector(sel selector) bool {
	_, ok := sel.(subjectSelector)
	return ok
}

func isClassLikeNode(node *ast.Node) bool {
	return node != nil && (node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression)
}

func isFunctionValueNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
		return node.Parent != nil && (isClassLikeNode(node.Parent) || node.Parent.Kind == ast.KindObjectLiteralExpression)
	}
	return false
}

func isEnumDeclarationNode(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindEnumDeclaration
}

func isChainRoot(node *ast.Node) bool {
	// Only the receiver continues a chain. An optional chain in an argument
	// or a computed key belongs to its own ChainExpression.
	return node != nil && ast.IsOptionalChain(node) &&
		(node.Parent == nil || !ast.IsOptionalChain(node.Parent) || node.Parent.Expression() != node)
}

func childTarget(node *ast.Node) string {
	if isChainRoot(unwrapEstreeNode(node)) {
		return "ChainExpression"
	}
	return "physical"
}

func estreeChildValue(node *ast.Node) interface{} {
	if isChainRoot(node) {
		return virtualNodeFacade{node, "ChainExpression"}
	}
	return node
}

func constructorKey(node *ast.Node, mc *matchContext) virtualNodeFacade {
	if mc != nil && mc.sf != nil {
		if token, ok := utils.TokenBeforePosition(mc.sf, node.ParameterList().Pos()-1); ok && token.Kind == ast.KindStringLiteral {
			return virtualNodeFacade{node, "Literal"}
		}
	}
	return virtualNodeFacade{node, "Identifier"}
}

// virtualTarget identifies a folded wrapper. Constructor keys are separate
// scalar facades; self-closing JSX uses the physical node for JSXElement.
func virtualTarget(node *ast.Node) string {
	switch {
	case isChainRoot(node):
		return "ChainExpression"
	case isClassLikeNode(node):
		return "ClassBody"
	case isEnumDeclarationNode(node):
		return "TSEnumBody"
	case node.Kind == ast.KindJsxSelfClosingElement:
		return "JSXOpeningElement"
	case isEmptyJSXExpression(node):
		return "JSXEmptyExpression"
	case isFunctionValueNode(node):
		if node.Body() == nil {
			return "TSEmptyBodyFunctionExpression"
		}
		return "FunctionExpression"
	}
	return ""
}

// virtualParent returns the physical node and wrapper identity that is the
// parent of a virtual ESTree node. The wrapper itself is represented by the
// same tsgo node, so this is enough to evaluate structural selectors without
// allocating synthetic AST nodes.
func virtualParent(node *ast.Node, target string) (*ast.Node, string, bool) {
	switch target {
	case "ChainExpression":
		return structuralParent(node)
	case "Identifier", "Literal":
		if node.Kind == ast.KindConstructor {
			return node, "physical", true
		}
	case "ClassBody":
		if isClassLikeNode(node) {
			return node, "physical", true
		}
	case "TSEnumBody":
		if isEnumDeclarationNode(node) {
			return node, "physical", true
		}
	case "JSXEmptyExpression":
		if isEmptyJSXExpression(node) {
			return node, "physical", true
		}
	case "FunctionExpression", "TSEmptyBodyFunctionExpression":
		if isFunctionValueNode(node) {
			return node, "physical", true
		}
	case "JSXOpeningElement":
		if node.Kind == ast.KindJsxSelfClosingElement {
			return node, "physical", true
		}
	}
	return nil, "", false
}

// logicalParent returns the parent edge in the ESTree graph. Most tsgo nodes
// have a direct physical equivalent, but class members, enum members, and
// method bodies pass through ESTree wrapper nodes that tsgo folds away.
func logicalParentTarget(node *ast.Node, target string, scopeRoot *virtualNodeFacade) (*ast.Node, string, bool) {
	if node == nil || scopeRoot != nil && node == scopeRoot.node && target == scopeRoot.typeName {
		return nil, "", false
	}
	if target != "physical" && target != "" {
		return virtualParent(node, target)
	}
	if isChainRoot(node) {
		return node, "ChainExpression", true
	}
	return structuralParent(node)
}

func structuralParent(node *ast.Node) (*ast.Node, string, bool) {
	parent := estreeParent(node)
	if parent == nil {
		return nil, "", false
	}
	if isClassLikeNode(parent) && nodeInList(node, classMembers(parent)) {
		return parent, "ClassBody", true
	}
	if isEnumDeclarationNode(parent) && nodeInList(node, enumMembers(parent)) {
		return parent, "TSEnumBody", true
	}
	if isFunctionValueNode(parent) && isFunctionValueChild(node, parent) {
		return parent, virtualTarget(parent), true
	}
	if parent.Kind == ast.KindJsxSelfClosingElement {
		return parent, "JSXOpeningElement", true
	}
	return parent, "physical", true
}

func classMembers(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindClassDeclaration:
		if members := node.AsClassDeclaration().Members; members != nil {
			return utils.ESTreeMembers(members.Nodes)
		}
	case ast.KindClassExpression:
		if members := node.AsClassExpression().Members; members != nil {
			return utils.ESTreeMembers(members.Nodes)
		}
	}
	return nil
}

func enumMembers(node *ast.Node) []*ast.Node {
	if node == nil || node.Kind != ast.KindEnumDeclaration {
		return nil
	}
	if members := node.AsEnumDeclaration().Members; members != nil {
		return members.Nodes
	}
	return nil
}

func nodeInList(node *ast.Node, list []*ast.Node) bool {
	for _, item := range list {
		if item == node {
			return true
		}
	}
	return false
}

func isFunctionValueChild(node, parent *ast.Node) bool {
	if node == nil || parent == nil {
		return false
	}
	if node == utils.ESTreeType(parent) || nodeInList(node, statements(parent.TypeParameterList())) {
		return true
	}
	if body := parent.Body(); body != nil && node == body {
		return true
	}
	for _, param := range parent.Parameters() {
		if param == node || unwrapEstreeNode(param) == node || isPlainParameter(param) && node == utils.ESTreeType(param) {
			return true
		}
	}
	return false
}

// selectorTargetsClassBody reports whether the selected (rightmost) identity
// is explicitly constrained to ESTree's virtual ClassBody node.
func selectorTargetsClassBody(sel selector) bool {
	return selectorTargetsEstreeType(sel, "ClassBody")
}

func selectorTargetsJSXEmptyExpression(sel selector) bool {
	return selectorTargetsEstreeType(sel, "JSXEmptyExpression")
}

func isBareWildcardSelector(sel selector) bool {
	switch value := sel.(type) {
	case subjectSelector:
		return isBareWildcardSelector(value.Inner)
	case identifierSelector:
		return value.Name == "*"
	default:
		return false
	}
}

func selectorTargetsEstreeType(sel selector, nodeType string) bool {
	switch value := sel.(type) {
	case subjectSelector:
		return selectorTargetsEstreeType(value.Inner, nodeType)
	case identifierSelector:
		return value.Name == nodeType
	case attrSelector:
		if len(value.Path) == 1 && value.Path[0] == "type" && value.Op == attrEqual {
			if expected, ok := exactStringValue(value.Value); ok && expected == nodeType {
				return true
			}
		}
		return selectorTargetsEstreeType(value.Inner, nodeType)
	case classSelector:
		return selectorTargetsEstreeType(value.Inner, nodeType)
	case combinedPseudo:
		return selectorTargetsEstreeType(value.Inner, nodeType)
	case pseudoSelector:
		switch value.Name {
		case "statement", "expression", "declaration", "function", "pattern":
			return semanticClassMatchesEstreeType(value.Name, nodeType)
		case "is", "matches":
			for _, child := range value.Args {
				if selectorTargetsEstreeType(child, nodeType) {
					return true
				}
			}
		}
	case combinatorSelector:
		return selectorTargetsEstreeType(value.Right, nodeType)
	case unionSelector:
		for _, child := range value.Selectors {
			if selectorTargetsEstreeType(child, nodeType) {
				return true
			}
		}
	}
	return false
}

func semanticClassMatchesEstreeType(class, nodeType string) bool {
	switch class {
	case "statement":
		return strings.HasSuffix(nodeType, "Statement") || strings.HasSuffix(nodeType, "Declaration")
	case "declaration":
		return strings.HasSuffix(nodeType, "Declaration")
	case "function":
		return nodeType == "FunctionDeclaration" || nodeType == "FunctionExpression" || nodeType == "ArrowFunctionExpression"
	case "expression", "pattern":
		return strings.HasSuffix(nodeType, "Expression") ||
			strings.HasSuffix(nodeType, "Literal") ||
			nodeType == "MetaProperty" || nodeType == "Identifier"
	}
	return false
}

func isEmptyJSXExpression(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindJsxExpression && node.AsJsxExpression().DotDotDotToken == nil && node.AsJsxExpression().Expression == nil
}

func matchesPseudoTarget(p pseudoSelector, node *ast.Node, mc *matchContext, scopeRoot *virtualNodeFacade, target string) bool {
	switch p.Name {
	case "is", "matches":
		for _, a := range p.Args {
			if matchesInScopeTarget(a, node, mc, scopeRoot, target) {
				return true
			}
		}
		return false
	case "not":
		for _, a := range p.Args {
			if matchesInScopeTarget(a, node, mc, scopeRoot, target) {
				return false
			}
		}
		return true
	case "has":
		for _, a := range p.Args {
			if hasMatching(node, a, mc, target) {
				return true
			}
		}
		return false
	case "nth-child":
		idx, _ := nodeIndexInListFieldTarget(node, scopeRoot, target)
		return idx >= 0 && idx == p.N-1
	case "nth-last-child":
		idx, total := nodeIndexInListFieldTarget(node, scopeRoot, target)
		if idx < 0 {
			return false
		}
		return total-idx == p.N
	case "statement", "expression", "declaration", "function", "pattern":
		return matchesNodeClassTarget(node, p.Name, scopeRoot, target)
	}
	return false
}

func matchesNodeClassTarget(node *ast.Node, class string, scopeRoot *virtualNodeFacade, target string) bool {
	nodeType := estreeNameForKind(node)
	if target != "" && target != "physical" {
		nodeType = target
	}
	switch class {
	case "statement":
		return strings.HasSuffix(nodeType, "Statement") || strings.HasSuffix(nodeType, "Declaration")
	case "declaration":
		return strings.HasSuffix(nodeType, "Declaration")
	case "function":
		return nodeType == "FunctionDeclaration" || nodeType == "FunctionExpression" || nodeType == "ArrowFunctionExpression"
	case "expression", "pattern":
		parent, _, _ := logicalParentTarget(node, target, scopeRoot)
		isExpression := strings.HasSuffix(nodeType, "Expression") ||
			strings.HasSuffix(nodeType, "Literal") ||
			nodeType == "MetaProperty" ||
			(nodeType == "Identifier" && (parent == nil || parent.Kind != ast.KindMetaProperty))
		if class == "expression" {
			return isExpression
		}
		return strings.HasSuffix(nodeType, "Pattern") || isExpression
	}
	return false
}

func hasMatching(node *ast.Node, sel selector, mc *matchContext, target string) bool {
	if relative, ok := sel.(relativeSelector); ok {
		if relative.Kind != combChild {
			return false
		}
		matched := false
		var visitDirect func(child *ast.Node, childTarget string) bool
		visitDirect = func(child *ast.Node, childTarget string) bool {
			child = unwrapEstreeNode(child)
			if expression := utils.JSDocTypeCastExpression(child); expression != nil {
				return visitDirect(expression, childTarget)
			}
			if utils.IsJSDocSyntaxNode(child) {
				return false
			}
			if isTransparentEstreeContainer(child) {
				forEachLogicalChild(child, childTarget, mc, visitDirect)
				return matched
			}
			if matchesInScopeTarget(relative.Inner, child, mc, &virtualNodeFacade{node, target}, childTarget) {
				matched = true
				return true
			}
			return false
		}
		forEachLogicalChild(node, target, mc, visitDirect)
		return matched
	}
	return hasDescendantMatching(node, sel, mc, target)
}

func hasDescendantMatching(node *ast.Node, sel selector, mc *matchContext, target string) bool {
	found := false
	var visit func(n *ast.Node, currentTarget string) bool
	visit = func(n *ast.Node, currentTarget string) bool {
		n = unwrapEstreeNode(n)
		if found {
			return true
		}
		if expression := utils.JSDocTypeCastExpression(n); expression != nil {
			return visit(expression, currentTarget)
		}
		if utils.IsJSDocSyntaxNode(n) {
			return false
		}
		if !isTransparentEstreeContainer(n) {
			if matchesInScopeTarget(sel, n, mc, &virtualNodeFacade{node, target}, currentTarget) {
				found = true
				return true
			}
		}
		forEachLogicalChild(n, currentTarget, mc, visit)
		return found
	}
	visit(node, target)
	return found
}

func forEachLogicalChild(node *ast.Node, target string, mc *matchContext, visit func(*ast.Node, string) bool) {
	if node == nil {
		return
	}
	visitPhysicalChildren := func(skip func(*ast.Node) bool) bool {
		return node.ForEachChild(func(child *ast.Node) bool {
			// ForEachChild includes operator/modifier tokens and EOF. Only
			// token kinds already exposed by the rule have an ESTree identity.
			if ast.IsToken(child) && !slices.Contains(allInterestingKinds, child.Kind) {
				return false
			}
			if skip != nil && skip(child) {
				return false
			}
			return visit(child, childTarget(child))
		})
	}
	switch target {
	case "ClassBody":
		for _, child := range classMembers(node) {
			if visit(child, "physical") {
				return
			}
		}
		return
	case "TSEnumBody":
		for _, child := range enumMembers(node) {
			if visit(child, "physical") {
				return
			}
		}
		return
	case "FunctionExpression", "TSEmptyBodyFunctionExpression":
		for _, child := range utils.ESTreeTypeParameters(node) {
			if visit(child, "physical") {
				return
			}
		}
		if t := utils.ESTreeType(node); t != nil && visit(t, "physical") {
			return
		}
		for _, child := range utils.ESTreeParameters(node) {
			if visit(child, "physical") {
				return
			}
		}
		if body := node.Body(); body != nil {
			visit(body, "physical")
		}
		return
	case "JSXOpeningElement":
		if node.Kind == ast.KindJsxSelfClosingElement {
			visitPhysicalChildren(nil)
		}
		return
	case "ChainExpression":
		visit(node, "physical")
		return
	case "JSXEmptyExpression", "Identifier", "Literal":
		return
	}

	switch {
	case isClassLikeNode(node):
		if visitPhysicalChildren(func(child *ast.Node) bool {
			return nodeInList(child, classMembers(node))
		}) {
			return
		}
		visit(node, "ClassBody")
	case isEnumDeclarationNode(node):
		if visitPhysicalChildren(func(child *ast.Node) bool {
			return nodeInList(child, enumMembers(node))
		}) {
			return
		}
		visit(node, "TSEnumBody")
	case isFunctionValueNode(node):
		if node.Kind == ast.KindConstructor && visit(node, constructorKey(node, mc).typeName) {
			return
		}
		if visitPhysicalChildren(func(child *ast.Node) bool {
			return isFunctionValueChild(child, node)
		}) {
			return
		}
		visit(node, virtualTarget(node))
	case node.Kind == ast.KindIdentifier && isPlainParameter(node.Parent) && node.Parent.Name() == node:
		if t := utils.ESTreeType(node.Parent); t != nil {
			visit(t, "physical")
		}
	case node.Kind == ast.KindJsxSelfClosingElement:
		visit(node, "JSXOpeningElement")
	case isEmptyJSXExpression(node):
		visit(node, "JSXEmptyExpression")
	default:
		visitPhysicalChildren(nil)
	}
}

// targetValue shares the attribute facade with structural field matching.
func targetValue(node *ast.Node, target string) interface{} {
	if target == "physical" || target == "" {
		return node
	}
	return virtualNodeFacade{node, target}
}

func nodeIsAtFieldPath(node *ast.Node, target string, current interface{}, path []string, mc *matchContext) bool {
	if nodes, ok := current.([]*ast.Node); ok {
		for _, child := range nodes {
			if nodeIsAtFieldPath(node, target, estreeChildValue(unwrapEstreeNode(child)), path, mc) {
				return true
			}
		}
		return false
	}
	if len(path) == 0 {
		switch value := current.(type) {
		case *ast.Node:
			return target == "physical" && node == unwrapEstreeNode(value)
		case virtualNodeFacade:
			return node == value.node && target == value.typeName
		}
		return false
	}
	next, ok := stepAttrPath(current, path[0], mc)
	if !ok {
		if physical, isNode := current.(*ast.Node); isNode && physical != nil {
			next = nodesAtField(physical, path[0])
		} else {
			return false
		}
	}
	return nodeIsAtFieldPath(node, target, next, path[1:], mc)
}

// lookupAttrPath walks `path` against `node` to fetch a comparable value.
// Each segment may resolve to an inner node or a primitive. Intermediate
// failures return ok=false.
func lookupAttrPath(node *ast.Node, path []string, mc *matchContext) (interface{}, bool) {
	var current interface{} = node
	for _, segment := range path {
		next, ok := stepAttrPath(current, segment, mc)
		if !ok {
			return nil, false
		}
		current = next
	}
	return current, true
}

func stepAttrPath(current interface{}, segment string, mc *matchContext) (interface{}, bool) {
	if current == nil {
		// esquery's getPath short-circuits a null intermediate value and
		// returns null, rather than converting it to undefined.
		return nil, true
	}
	if n, ok := current.(*ast.Node); ok {
		n = unwrapEstreeNode(n)
		if n == nil {
			return nil, true
		}
		value, found := readNodeAttr(n, segment, mc)
		if child, ok := value.(*ast.Node); ok && child != n && segment != "parent" {
			return estreeChildValue(child), found
		}
		return value, found
	}
	if rf, ok := current.(regexFacade); ok {
		switch segment {
		case "flags":
			_, flags := splitRegexLiteral(regexLiteralText(rf.node, rf.mc))
			return flags, true
		case "pattern":
			pat, _ := splitRegexLiteral(regexLiteralText(rf.node, rf.mc))
			return pat, true
		}
		return nil, false
	}
	if mi, ok := current.(metaIdentifier); ok {
		switch segment {
		case "name":
			return mi.name, true
		case "type":
			return "Identifier", true
		}
		return nil, false
	}
	if vn, ok := current.(virtualNodeFacade); ok {
		switch segment {
		case "type":
			return vn.typeName, true
		case "parent":
			parent, target, ok := logicalParentTarget(vn.node, vn.typeName, nil)
			if !ok {
				return nil, true
			}
			return targetValue(parent, target), true
		}
		return readVirtualNodeAttr(vn, segment, mc)
	}
	// Primitive interpretations: support `.length` on strings / slices.
	if segment == "length" {
		switch v := current.(type) {
		case string:
			return float64(ecmascript.StringCodeUnitCount(v)), true
		case []*ast.Node:
			return float64(len(v)), true
		}
	}
	if nodes, ok := current.([]*ast.Node); ok {
		index, err := strconv.Atoi(segment)
		if err != nil || index < 0 || index >= len(nodes) || strconv.Itoa(index) != segment {
			return nil, false
		}
		return estreeChildValue(unwrapEstreeNode(nodes[index])), true
	}
	return nil, false
}

type undefinedAttr struct{}

// Keep BigInt distinct from string for typeof and regexp selectors. Its
// canonical decimal text is also the value esquery uses for string equality.
type bigintAttr string

func compareStringAttr(left string, op attrOp, right attrValue) bool {
	equal := attrStringEquals(left, right)
	if op == attrNotEqual {
		return !equal
	}
	return equal
}

func compareAttr(left interface{}, op attrOp, right attrValue) bool {
	switch op {
	case attrEqual:
		if right.Kind == attrValueRegex {
			text, ok := left.(string)
			return ok && attrStringEquals(text, right)
		}
		return attrEquals(left, right)
	case attrNotEqual:
		if right.Kind == attrValueRegex {
			return !attrStringEquals(attrAsString(left), right)
		}
		return !attrEquals(left, right)
	case attrLess, attrLessOrEqual, attrGreater, attrGreaterOrEqual:
		comparison, ok := compareRelationalAttr(left, right)
		if !ok {
			return false
		}
		switch op {
		case attrLess:
			return comparison < 0
		case attrLessOrEqual:
			return comparison <= 0
		case attrGreater:
			return comparison > 0
		case attrGreaterOrEqual:
			return comparison >= 0
		}
	}
	return false
}

func compareRelationalAttr(left interface{}, right attrValue) (int, bool) {
	if value, ok := left.(bigintAttr); ok {
		leftInt, valid := ecmascript.StringToBigInt(string(value))
		if !valid {
			return 0, false
		}
		if right.Kind == attrValueString || right.Kind == attrValueIdent {
			text := right.Str
			if right.Kind == attrValueIdent {
				text = right.Ident
			}
			rightInt, valid := ecmascript.StringToBigInt(text)
			if !valid {
				return 0, false
			}
			return leftInt.Cmp(rightInt), true
		}
		if right.Kind != attrValueNumber || math.IsNaN(right.Num) {
			return 0, false
		}
		// SetInt chooses enough precision to preserve every BigInt bit; the
		// other operand is already an IEEE 754 number in esquery's AST.
		return new(big.Float).SetInt(leftInt).Cmp(new(big.Float).SetFloat64(right.Num)), true
	}
	leftText, leftIsString, leftNumber, leftOK := leftRelationalPrimitive(left)
	rightText, rightIsString, rightNumber, rightOK := rightRelationalPrimitive(right)
	if !leftOK || !rightOK {
		return 0, false
	}
	if leftIsString && rightIsString {
		return ecmascript.CompareStrings(leftText, rightText), true
	}
	if leftIsString {
		leftNumber, leftOK = ecmascript.StringToNumber(leftText)
	}
	if rightIsString {
		rightNumber, rightOK = ecmascript.StringToNumber(rightText)
	}
	if !leftOK || !rightOK || math.IsNaN(leftNumber) || math.IsNaN(rightNumber) {
		return 0, false
	}
	switch {
	case leftNumber < rightNumber:
		return -1, true
	case leftNumber > rightNumber:
		return 1, true
	default:
		return 0, true
	}
}

func leftRelationalPrimitive(value interface{}) (text string, isString bool, number float64, ok bool) {
	switch typed := value.(type) {
	case string:
		return typed, true, 0, true
	case float64:
		return "", false, typed, true
	case int:
		return "", false, float64(typed), true
	case bool:
		if typed {
			return "", false, 1, true
		}
		return "", false, 0, true
	case nil:
		return "", false, 0, true
	case undefinedAttr:
		return "", false, 0, false
	case *ast.Node:
		if typed == nil {
			return "", false, 0, true
		}
		return attrAsString(typed), true, 0, true
	case []*ast.Node, regexFacade, metaIdentifier, virtualNodeFacade:
		return attrAsString(typed), true, 0, true
	default:
		return "", false, 0, false
	}
}

func rightRelationalPrimitive(value attrValue) (text string, isString bool, number float64, ok bool) {
	switch value.Kind {
	case attrValueString:
		return value.Str, true, 0, true
	case attrValueIdent:
		return value.Ident, true, 0, true
	case attrValueNumber:
		return "", false, value.Num, true
	default:
		return "", false, 0, false
	}
}

func attrEquals(left interface{}, right attrValue) bool {
	switch right.Kind {
	case attrValueString:
		return attrAsString(left) == right.Str
	case attrValueNumber:
		return attrAsString(left) == attrAsString(right.Num)
	case attrValueIdent:
		return attrAsString(left) == right.Ident
	case attrValueType:
		return attrTypeOf(left) == right.Ident
	case attrValueRegex:
		return attrStringEquals(attrAsString(left), right)
	}
	return false
}

func attrStringEquals(left string, right attrValue) bool {
	switch right.Kind {
	case attrValueString:
		return left == right.Str
	case attrValueIdent:
		return left == right.Ident
	case attrValueRegex:
		if right.regexPrefix != "" && !strings.HasPrefix(left, right.regexPrefix) {
			return false
		}
		re := right.compiledRegex
		if re == nil {
			var err error
			re, err = esregexp.Compile(right.Regex, right.Flags)
			if err != nil {
				return false
			}
		}
		return re.Test(left)
	}
	return false
}

func attrAsString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case bigintAttr:
		return string(x)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		return ecmascript.NumberToString(x)
	case int:
		return strconv.Itoa(x)
	case nil:
		return "null"
	case undefinedAttr:
		return "undefined"
	case *ast.Node:
		if x == nil {
			return "null"
		}
		return "[object Object]"
	case []*ast.Node:
		if len(x) == 0 {
			return ""
		}
		var text strings.Builder
		text.Grow(len(x) * len("[object Object],"))
		for index, node := range x {
			if index > 0 {
				text.WriteByte(',')
			}
			if node != nil {
				text.WriteString("[object Object]")
			}
		}
		return text.String()
	case regexFacade, metaIdentifier, virtualNodeFacade:
		return "[object Object]"
	}
	return ""
}

func attrTypeOf(v interface{}) string {
	switch v.(type) {
	case undefinedAttr:
		return "undefined"
	case bigintAttr:
		return "bigint"
	case string:
		return "string"
	case bool:
		return "boolean"
	case float64, int:
		return "number"
	default:
		return "object"
	}
}

func nodeIndexInListFieldTarget(node *ast.Node, scopeRoot *virtualNodeFacade, target string) (int, int) {
	siblings, index := siblingsOf(node, target, scopeRoot)
	return index, len(siblings)
}

func siblingsOf(node *ast.Node, target string, scopeRoot *virtualNodeFacade) ([]*ast.Node, int) {
	// All folded children occupy scalar fields, never a sibling array.
	if target != "physical" && target != "ChainExpression" || target == "physical" && isChainRoot(node) {
		return nil, -1
	}
	parent, parentTarget, ok := logicalParentTarget(node, target, scopeRoot)
	if !ok {
		return nil, -1
	}
	var lists [][]*ast.Node
	switch parentTarget {
	case "ClassBody":
		lists = [][]*ast.Node{classMembers(parent)}
	case "TSEnumBody":
		lists = [][]*ast.Node{enumMembers(parent)}
	case "FunctionExpression", "TSEmptyBodyFunctionExpression":
		lists = [][]*ast.Node{utils.ESTreeParameters(parent)}
	case "JSXOpeningElement":
		lists = listChildrenOf(parent)
	case "physical":
		lists = listChildrenOf(parent)
	}
	for _, list := range lists {
		if index, _ := indexInNodeList(node, list); index >= 0 {
			return list, index
		}
	}
	return nil, -1
}

func indexInNodeList(node *ast.Node, list []*ast.Node) (int, int) {
	for index, child := range list {
		if unwrapEstreeNode(child) == node {
			return index, len(list)
		}
	}
	return -1, 0
}

// isDefaultExportedDeclaration reports whether `node` carries the
// `export default` modifier combo that tsgo attaches directly to a
// declaration (instead of wrapping the declaration in an
// ExportDefaultDeclaration as ESTree does).
func isDefaultExportedDeclaration(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
	default:
		return false
	}
	return ast.HasSyntacticModifier(node, ast.ModifierFlagsExportDefault)
}

// selectorMatchesVirtualExportDefault reports whether the given selector
// would match a synthetic ExportDefaultDeclaration node — used to honour
// `ExportDefaultDeclaration > X` selectors when the tsgo node sits
// directly under SourceFile but ESTree would place it under an
// ExportDefaultDeclaration wrapper.
func selectorMatchesVirtualExportDefault(sel selector) bool {
	switch v := sel.(type) {
	case subjectSelector:
		return selectorMatchesVirtualExportDefault(v.Inner)
	case identifierSelector:
		return v.Name == "*" || v.Name == "ExportDefaultDeclaration"
	case unionSelector:
		for _, a := range v.Selectors {
			if selectorMatchesVirtualExportDefault(a) {
				return true
			}
		}
	case combinedPseudo:
		if selectorMatchesVirtualExportDefault(v.Inner) {
			// `:not` cannot be evaluated reliably against a virtual
			// node — ESLint's behaviour is well-defined only on real
			// nodes, so treat the pseudo as conservatively passing.
			return true
		}
	case pseudoSelector:
		if v.Name == "is" || v.Name == "matches" {
			for _, a := range v.Args {
				if selectorMatchesVirtualExportDefault(a) {
					return true
				}
			}
		}
	}
	return false
}

// unwrapExpression strips the tsgo-only "transparent" expression wrappers
// (parentheses, type assertions, non-null, satisfies) so that attribute
// paths like `object.name` see through them just like esquery does on
// ESTree, where these wrappers don't exist (parens) or have different
// shapes that esquery still walks. Without this, a real-world selector
// like `MemberExpression[object.name='console']` would silently miss
// `(console).log`, `(console as any).log`, `console!.log`.
func unwrapExpression(node *ast.Node) *ast.Node {
	return utils.SkipAssertionsAndParens(node)
}

func unwrapEstreeNode(node *ast.Node) *ast.Node {
	for node != nil {
		node = utils.ESTreeRuntimeExpression(node)
		if node == nil {
			return nil
		}
		if isPlainParameter(node) {
			node = node.Name()
			continue
		}
		if node.Kind == ast.KindComputedPropertyName {
			node = node.AsComputedPropertyName().Expression
			continue
		}
		return node
	}
	return nil
}

func isPlainParameter(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindParameter &&
		node.AsParameterDeclaration().Initializer == nil &&
		node.AsParameterDeclaration().DotDotDotToken == nil &&
		!ast.HasSyntacticModifier(node, ast.ModifierFlagsParameterPropertyModifier)
}

func isTransparentEstreeContainer(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindParenthesizedExpression, ast.KindComputedPropertyName, ast.KindSemicolonClassElement,
		ast.KindImportAttributes, ast.KindJsxAttributes, ast.KindHeritageClause, ast.KindExpressionWithTypeArguments:
		return true
	case ast.KindBlock:
		return node.Parent != nil && node.Parent.Kind == ast.KindClassStaticBlockDeclaration
	case ast.KindVariableDeclarationList:
		return node.Parent != nil && node.Parent.Kind == ast.KindVariableStatement
	}
	return false
}

func estreeParent(node *ast.Node) *ast.Node {
	parent := utils.ESTreeParent(node)
	for parent != nil && (isTransparentEstreeContainer(parent) || isPlainParameter(parent)) {
		parent = utils.ESTreeParent(parent)
	}
	return parent
}

func estreeParentValue(node *ast.Node) interface{} {
	parent, target, ok := logicalParentTarget(node, "physical", nil)
	if !ok {
		return nil
	}
	return targetValue(parent, target)
}

// readNodeAttr extracts the named ESTree-style attribute from a tsgo node.
// Centralising the field map here makes it easy to widen support without
// rewriting the matcher.
func readNodeAttr(node *ast.Node, name string, mc *matchContext) (interface{}, bool) {
	if node == nil {
		return nil, false
	}
	switch name {
	case "type":
		if isJSXMemberExpression(node) {
			return "JSXMemberExpression", true
		}
		if isJSXIdentifier(node) {
			return "JSXIdentifier", true
		}
		return estreeNameForKind(node), true
	case "parent":
		return estreeParentValue(node), true
	case "name":
		return readNameAttr(node)
	case "value":
		return readValueAttr(node, mc)
	case "raw":
		return readRawAttr(node, mc)
	case "operator":
		return readOperatorAttr(node)
	case "kind":
		return readKindAttr(node)
	case "optional":
		return readOptionalAttr(node)
	case "computed":
		return readComputedAttr(node)
	case "static":
		return readStaticAttr(node)
	case "shorthand":
		return readShorthandAttr(node)
	case "method":
		return readMethodAttr(node)
	case "superClass":
		return readSuperClassAttr(node)
	case "directive":
		return readDirectiveAttr(node, mc)
	case "update":
		return readUpdateAttr(node)
	case "expressions":
		return readExpressionsAttr(node)
	case "quasis":
		return readQuasisAttr(node)
	case "bigint":
		return readBigintAttr(node, mc)
	case "delegate":
		return readDelegateAttr(node)
	case "label":
		return readLabelAttr(node)
	case "regex":
		return readRegexObject(node, mc)
	case "flags":
		return readRegexFlags(node, mc)
	case "pattern":
		return readRegexPattern(node, mc)
	case "params":
		return readParamsAttr(node)
	case "length":
		return float64(0), false
	case "source":
		return readSourceAttr(node)
	case "callee":
		return readCalleeAttr(node)
	case "arguments":
		return readArgumentsAttr(node)
	case "expression":
		return readExpressionAttr(node)
	case "init":
		return readInitAttr(node)
	case "id":
		return readIdAttr(node)
	case "key":
		return readKeyAttr(node, mc)
	case "left":
		return readLeftAttr(node)
	case "right":
		return readRightAttr(node)
	case "object":
		return readObjectAttr(node)
	case "property":
		return readPropertyAttr(node)
	case "test":
		return readTestAttr(node)
	case "consequent":
		return readConsequentAttr(node)
	case "alternate":
		return readAlternateAttr(node)
	case "body":
		return readBodyAttr(node)
	case "argument":
		return readArgumentAttr(node)
	case "prefix":
		return readPrefixAttr(node)
	case "async":
		return readAsyncAttr(node)
	case "generator":
		return readGeneratorAttr(node)
	case "specifiers":
		return readSpecifiersAttr(node)
	case "selfClosing":
		if node.Kind == ast.KindJsxOpeningElement {
			return false, true
		}
		return nil, false
	case "openingElement":
		return readOpeningElementAttr(node)
	case "closingElement":
		return readClosingElementAttr(node)
	case "attributes":
		return readAttributesAttr(node)
	case "children":
		return readChildrenAttr(node)
	case "tagName":
		return readTagNameAttr(node)
	case "param":
		return readParamAttr(node)
	case "imported":
		return readImportedAttr(node)
	case "local":
		return readLocalAttr(node)
	case "exported":
		return readExportedAttr(node)
	case "meta":
		return readMetaAttr(node)
	case "tag":
		return readTagAttr(node)
	case "quasi":
		return readQuasiAttr(node)
	}
	return nil, false
}

func readOpeningElementAttr(node *ast.Node) (interface{}, bool) {
	if node.Kind == ast.KindJsxSelfClosingElement {
		return virtualNodeFacade{node, "JSXOpeningElement"}, true
	}
	if node.Kind == ast.KindJsxElement {
		return node.AsJsxElement().OpeningElement, true
	}
	return nil, false
}

func readClosingElementAttr(node *ast.Node) (interface{}, bool) {
	if node.Kind == ast.KindJsxSelfClosingElement {
		return nil, true
	}
	if node.Kind == ast.KindJsxElement {
		return node.AsJsxElement().ClosingElement, true
	}
	return nil, false
}

func readAttributesAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindJsxOpeningElement:
		return statements(node.AsJsxOpeningElement().Attributes.AsJsxAttributes().Properties), true
	case ast.KindImportDeclaration:
		return importAttributes(node.AsImportDeclaration().Attributes), true
	case ast.KindExportDeclaration:
		return importAttributes(node.AsExportDeclaration().Attributes), true
	}
	return nil, false
}

func importAttributes(node *ast.Node) []*ast.Node {
	if node == nil || node.AsImportAttributes().Attributes == nil {
		return []*ast.Node{}
	}
	return node.AsImportAttributes().Attributes.Nodes
}

func readChildrenAttr(node *ast.Node) (interface{}, bool) {
	if node.Kind == ast.KindJsxSelfClosingElement {
		return []*ast.Node{}, true
	}
	if node.Kind == ast.KindJsxElement {
		c := node.AsJsxElement().Children
		if c == nil {
			return []*ast.Node{}, true
		}
		return c.Nodes, true
	}
	return nil, false
}

// readTagNameAttr retains rslint's tsgo-style alias for JSX tag names.
func readTagNameAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindJsxOpeningElement:
		return node.AsJsxOpeningElement().TagName, true
	case ast.KindJsxSelfClosingElement:
		return node.AsJsxSelfClosingElement().TagName, true
	case ast.KindJsxClosingElement:
		return node.AsJsxClosingElement().TagName, true
	}
	return nil, false
}

func readParamAttr(node *ast.Node) (interface{}, bool) {
	if node.Kind != ast.KindCatchClause {
		return nil, false
	}
	cc := node.AsCatchClause()
	if cc.VariableDeclaration == nil {
		return nil, true // present but null — falsy for [param]
	}
	return cc.VariableDeclaration.AsVariableDeclaration().Name(), true
}

func readImportedAttr(node *ast.Node) (interface{}, bool) {
	if node.Kind != ast.KindImportSpecifier {
		return nil, false
	}
	is := node.AsImportSpecifier()
	if is.PropertyName != nil {
		return is.PropertyName, true
	}
	return is.Name(), true
}

func readLocalAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindImportSpecifier:
		return node.AsImportSpecifier().Name(), true
	case ast.KindImportClause:
		return node.AsImportClause().Name(), true
	case ast.KindNamespaceImport:
		return node.AsNamespaceImport().Name(), true
	case ast.KindExportSpecifier:
		// ESTree's `local` is the source binding (the LHS of `as`).
		// tsgo flips the storage: for `export { foo as bar }`, foo is
		// PropertyName and bar is Name; for `export { foo }`, only Name
		// is set (PropertyName is nil).
		es := node.AsExportSpecifier()
		if es.PropertyName != nil {
			return es.PropertyName, true
		}
		return es.Name(), true
	}
	return nil, false
}

func readExportedAttr(node *ast.Node) (interface{}, bool) {
	if node.Kind == ast.KindExportSpecifier {
		// ESTree's `exported` is the export name (the RHS of `as`,
		// or the only identifier when no `as` is present). tsgo's
		// Name() always holds that value.
		es := node.AsExportSpecifier()
		return es.Name(), true
	}
	return nil, false
}

func readMetaAttr(node *ast.Node) (interface{}, bool) {
	if node.Kind != ast.KindMetaProperty {
		return nil, false
	}
	// ESTree's `meta` is the keyword identifier (e.g. `new` in
	// `new.target`, `import` in `import.meta`). tsgo stores only the
	// keyword Kind, so synthesize a string the matcher can compare
	// against literal selectors like `[meta.name='new']` or
	// `[meta.name='import']`.
	switch node.AsMetaProperty().KeywordToken {
	case ast.KindNewKeyword:
		return metaIdentifier{name: "new"}, true
	case ast.KindImportKeyword:
		return metaIdentifier{name: "import"}, true
	}
	return nil, false
}

// metaIdentifier stands in for tsgo's missing Identifier wrapper around
// the keyword token of a MetaProperty. It exposes a single `.name`
// attribute so esquery-style paths like `meta.name='new'` resolve.
type metaIdentifier struct {
	name string
}

func readTagAttr(node *ast.Node) (interface{}, bool) {
	if node.Kind == ast.KindTaggedTemplateExpression {
		return node.AsTaggedTemplateExpression().Tag, true
	}
	return nil, false
}

func readQuasiAttr(node *ast.Node) (interface{}, bool) {
	if node.Kind == ast.KindTaggedTemplateExpression {
		return node.AsTaggedTemplateExpression().Template, true
	}
	return nil, false
}

// estreeNameForKind returns the ESTree type name for a tsgo node when one
// can be unambiguously chosen. If multiple ESTree names map to the same
// kind, the canonical one is returned.
func estreeNameForKind(node *ast.Node) string {
	if isJSXIdentifier(node) {
		return "JSXIdentifier"
	}
	if isJSXMemberExpression(node) {
		return "JSXMemberExpression"
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return "Identifier"
	case ast.KindPrivateIdentifier:
		return "PrivateIdentifier"
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindRegularExpressionLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return "Literal"
	case ast.KindArrowFunction:
		return "ArrowFunctionExpression"
	case ast.KindFunctionExpression:
		return "FunctionExpression"
	case ast.KindFunctionDeclaration:
		return "FunctionDeclaration"
	case ast.KindBlock:
		return "BlockStatement"
	case ast.KindVariableStatement:
		return "VariableDeclaration"
	case ast.KindVariableDeclaration:
		return "VariableDeclarator"
	case ast.KindVariableDeclarationList:
		// tsgo wraps the declarators in VariableDeclarationList in three
		// places: VariableStatement (already maps to "VariableDeclaration"),
		// and ForStatement / ForInStatement / ForOfStatement initializers.
		// In ESTree the for-loop position is itself a VariableDeclaration —
		// without this mapping, `[left.type='VariableDeclaration']` on
		// for-in selectors would silently fail.
		return "VariableDeclaration"
	case ast.KindBreakStatement:
		return "BreakStatement"
	case ast.KindContinueStatement:
		return "ContinueStatement"
	case ast.KindCatchClause:
		return "CatchClause"
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call.Expression != nil && call.Expression.Kind == ast.KindImportKeyword {
			return "ImportExpression"
		}
		return "CallExpression"
	case ast.KindNewExpression:
		return "NewExpression"
	case ast.KindEmptyStatement:
		return "EmptyStatement"
	case ast.KindTryStatement:
		return "TryStatement"
	case ast.KindConditionalExpression:
		return "ConditionalExpression"
	case ast.KindBinaryExpression:
		op := node.AsBinaryExpression().OperatorToken.Kind
		if isAssignmentOperatorKind(op) {
			return "AssignmentExpression"
		}
		if isLogicalOperatorKind(op) {
			return "LogicalExpression"
		}
		if op == ast.KindCommaToken {
			return "SequenceExpression"
		}
		return "BinaryExpression"
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return "MemberExpression"
	case ast.KindObjectLiteralExpression:
		return "ObjectExpression"
	case ast.KindArrayLiteralExpression:
		return "ArrayExpression"
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
		return "Property"
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		// Disambiguate Property vs MethodDefinition by the lexical
		// container: object literal → Property, class body → MethodDefinition.
		if node.Parent != nil {
			switch node.Parent.Kind {
			case ast.KindClassDeclaration, ast.KindClassExpression:
				if ast.HasSyntacticModifier(node, ast.ModifierFlagsAbstract) {
					return "TSAbstractMethodDefinition"
				}
				return "MethodDefinition"
			case ast.KindObjectLiteralExpression:
				return "Property"
			}
		}
		return "TSMethodSignature"
	case ast.KindImportDeclaration:
		return "ImportDeclaration"
	case ast.KindClassDeclaration:
		return "ClassDeclaration"
	case ast.KindClassExpression:
		return "ClassExpression"
	// tsgo's split unary-like kinds map back onto ESTree's UnaryExpression
	// (and update for ++/--). The synthetic `type` field on these nodes
	// must therefore report the ESTree name so that path attributes like
	// `[left.type='UnaryExpression']` work.
	case ast.KindTypeOfExpression, ast.KindVoidExpression, ast.KindDeleteExpression:
		return "UnaryExpression"
	case ast.KindAwaitExpression:
		return "AwaitExpression"
	case ast.KindYieldExpression:
		return "YieldExpression"
	case ast.KindPrefixUnaryExpression:
		op := node.AsPrefixUnaryExpression().Operator
		if op == ast.KindPlusPlusToken || op == ast.KindMinusMinusToken {
			return "UpdateExpression"
		}
		return "UnaryExpression"
	case ast.KindPostfixUnaryExpression:
		return "UpdateExpression"
	case ast.KindParenthesizedExpression:
		// ESTree drops parens, so a node walking via `type` ought to see
		// the inner expression's type instead.
		return estreeNameForKind(unwrapExpression(node))
	case ast.KindAsExpression:
		return "TSAsExpression"
	case ast.KindSatisfiesExpression:
		return "TSSatisfiesExpression"
	case ast.KindNonNullExpression:
		return "TSNonNullExpression"
	case ast.KindTypeAssertionExpression:
		return "TSTypeAssertion"
	case ast.KindDecorator:
		return "Decorator"
	case ast.KindThisKeyword:
		return "ThisExpression"
	case ast.KindSuperKeyword:
		return "Super"
	case ast.KindSpreadElement, ast.KindSpreadAssignment:
		return "SpreadElement"
	case ast.KindTaggedTemplateExpression:
		return "TaggedTemplateExpression"
	case ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression:
		return "TemplateLiteral"
	case ast.KindForInStatement:
		return "ForInStatement"
	case ast.KindForOfStatement:
		return "ForOfStatement"
	case ast.KindForStatement:
		return "ForStatement"
	case ast.KindIfStatement:
		return "IfStatement"
	case ast.KindWhileStatement:
		return "WhileStatement"
	case ast.KindDoStatement:
		return "DoWhileStatement"
	case ast.KindReturnStatement:
		return "ReturnStatement"
	case ast.KindThrowStatement:
		return "ThrowStatement"
	case ast.KindSwitchStatement:
		return "SwitchStatement"
	case ast.KindWithStatement:
		return "WithStatement"
	case ast.KindLabeledStatement:
		return "LabeledStatement"
	case ast.KindDebuggerStatement:
		return "DebuggerStatement"
	case ast.KindExpressionStatement:
		return "ExpressionStatement"
	case ast.KindCaseClause, ast.KindDefaultClause:
		return "SwitchCase"
	case ast.KindSourceFile:
		return "Program"
	case ast.KindJsxElement:
		return "JSXElement"
	case ast.KindJsxOpeningElement:
		return "JSXOpeningElement"
	case ast.KindJsxSelfClosingElement:
		return "JSXElement"
	case ast.KindJsxClosingElement:
		return "JSXClosingElement"
	case ast.KindJsxExpression:
		if node.AsJsxExpression().DotDotDotToken != nil {
			return "JSXSpreadChild"
		}
		return "JSXExpressionContainer"
	case ast.KindMetaProperty:
		return "MetaProperty"
	case ast.KindPropertyDeclaration:
		return "PropertyDefinition"
	case ast.KindEnumDeclaration:
		return "TSEnumDeclaration"
	case ast.KindEnumMember:
		return "TSEnumMember"
	case ast.KindConstructor:
		return "MethodDefinition"
	case ast.KindClassStaticBlockDeclaration:
		return "StaticBlock"
	case ast.KindArrayBindingPattern:
		return "ArrayPattern"
	case ast.KindObjectBindingPattern:
		return "ObjectPattern"
	case ast.KindImportAttribute:
		return "ImportAttribute"
	}
	return ""
}

func readNameAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text, true
	case ast.KindPrivateIdentifier:
		return strings.TrimPrefix(node.AsPrivateIdentifier().Text, "#"), true
	// JSX: ESTree calls the tag identifier `name`; tsgo stores it as
	// TagName. Also expose it through this attribute path.
	case ast.KindJsxOpeningElement:
		return node.AsJsxOpeningElement().TagName, true
	case ast.KindJsxClosingElement:
		return node.AsJsxClosingElement().TagName, true
	case ast.KindJsxAttribute:
		// ESTree's JSXAttribute.name is a JSXIdentifier; tsgo stores
		// the attribute name on the unexported `name` field accessed
		// through Name().
		return node.AsJsxAttribute().Name(), true
	}
	return nil, false
}

func readValueAttr(node *ast.Node, mc *matchContext) (interface{}, bool) {
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text, true
	case ast.KindNumericLiteral:
		text := node.AsNumericLiteral().Text
		if n, err := strconv.ParseFloat(text, 64); err == nil {
			return n, true
		}
		return text, true
	case ast.KindBigIntLiteral:
		return bigintAttr(utils.NormalizeBigIntLiteral(node.AsBigIntLiteral().Text)), true
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text, true
	case ast.KindTrueKeyword:
		return true, true
	case ast.KindFalseKeyword:
		return false, true
	case ast.KindNullKeyword:
		return nil, true
	// ESTree's Property.value is the right-hand expression
	// (`a: <value>`). tsgo stores it on PropertyAssignment.Initializer.
	case ast.KindPropertyAssignment:
		return unwrapExpression(node.AsPropertyAssignment().Initializer), true
	// ESTree's PropertyDefinition.value is the field initializer; null
	// when uninitialised. tsgo's PropertyDeclaration uses Initializer.
	case ast.KindPropertyDeclaration:
		return unwrapExpression(node.AsPropertyDeclaration().Initializer), true
	// ESTree stores a class method's function fields under
	// MethodDefinition.value, while tsgo puts them directly on the method.
	case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
		if !isFunctionValueNode(node) {
			return nil, false
		}
		return virtualNodeFacade{node: node, typeName: virtualTarget(node)}, true
	case ast.KindJsxAttribute:
		return node.AsJsxAttribute().Initializer, true
	case ast.KindImportAttribute:
		return unwrapExpression(node.AsImportAttribute().Value), true
	}
	return nil, false
}

func readRawAttr(node *ast.Node, mc *matchContext) (interface{}, bool) {
	if mc == nil || mc.sf == nil || !utils.IsESTreeLiteralKind(node.Kind) {
		return nil, false
	}
	return scanner.GetSourceTextOfNodeFromSourceFile(mc.sf, node, false), true
}

func readOptionalAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return ast.IsOptionalChainRoot(node), true
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call.Expression != nil && call.Expression.Kind == ast.KindImportKeyword {
			return nil, false
		}
		return ast.IsOptionalChainRoot(node), true
	}
	return nil, false
}

func readOperatorAttr(node *ast.Node) (interface{}, bool) {
	text, ok := readOperatorAttrString(node)
	if !ok {
		return nil, false
	}
	return text, true
}

func readOperatorAttrString(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindBinaryExpression:
		return operatorTokenText(node.AsBinaryExpression().OperatorToken.Kind), true
	case ast.KindPrefixUnaryExpression:
		return operatorTokenText(node.AsPrefixUnaryExpression().Operator), true
	case ast.KindPostfixUnaryExpression:
		return operatorTokenText(node.AsPostfixUnaryExpression().Operator), true
	// tsgo splits `typeof` / `void` / `delete` into their own kinds. ESTree
	// keeps them as a UnaryExpression with the corresponding operator
	// string — selectors like `UnaryExpression[operator='typeof']` need
	// the same string back.
	case ast.KindTypeOfExpression:
		return "typeof", true
	case ast.KindVoidExpression:
		return "void", true
	case ast.KindDeleteExpression:
		return "delete", true
	case ast.KindAwaitExpression:
		return "await", true
	case ast.KindYieldExpression:
		return "yield", true
	}
	return "", false
}

func operatorTokenText(k ast.Kind) string {
	switch k {
	case ast.KindPlusToken:
		return "+"
	case ast.KindMinusToken:
		return "-"
	case ast.KindAsteriskToken:
		return "*"
	case ast.KindAsteriskAsteriskToken:
		return "**"
	case ast.KindSlashToken:
		return "/"
	case ast.KindPercentToken:
		return "%"
	case ast.KindEqualsToken:
		return "="
	case ast.KindPlusEqualsToken:
		return "+="
	case ast.KindMinusEqualsToken:
		return "-="
	case ast.KindAsteriskEqualsToken:
		return "*="
	case ast.KindAsteriskAsteriskEqualsToken:
		return "**="
	case ast.KindSlashEqualsToken:
		return "/="
	case ast.KindPercentEqualsToken:
		return "%="
	case ast.KindEqualsEqualsToken:
		return "=="
	case ast.KindEqualsEqualsEqualsToken:
		return "==="
	case ast.KindExclamationEqualsToken:
		return "!="
	case ast.KindExclamationEqualsEqualsToken:
		return "!=="
	case ast.KindLessThanToken:
		return "<"
	case ast.KindLessThanEqualsToken:
		return "<="
	case ast.KindGreaterThanToken:
		return ">"
	case ast.KindGreaterThanEqualsToken:
		return ">="
	case ast.KindLessThanLessThanToken:
		return "<<"
	case ast.KindGreaterThanGreaterThanToken:
		return ">>"
	case ast.KindGreaterThanGreaterThanGreaterThanToken:
		return ">>>"
	case ast.KindAmpersandToken:
		return "&"
	case ast.KindBarToken:
		return "|"
	case ast.KindCaretToken:
		return "^"
	case ast.KindAmpersandAmpersandToken:
		return "&&"
	case ast.KindBarBarToken:
		return "||"
	case ast.KindQuestionQuestionToken:
		return "??"
	case ast.KindInKeyword:
		return "in"
	case ast.KindInstanceOfKeyword:
		return "instanceof"
	case ast.KindCommaToken:
		return ","
	case ast.KindExclamationToken:
		return "!"
	case ast.KindTildeToken:
		return "~"
	case ast.KindPlusPlusToken:
		return "++"
	case ast.KindMinusMinusToken:
		return "--"
	case ast.KindAmpersandEqualsToken:
		return "&="
	case ast.KindBarEqualsToken:
		return "|="
	case ast.KindCaretEqualsToken:
		return "^="
	case ast.KindAmpersandAmpersandEqualsToken:
		return "&&="
	case ast.KindBarBarEqualsToken:
		return "||="
	case ast.KindQuestionQuestionEqualsToken:
		return "??="
	}
	return ""
}

func readKindAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindVariableStatement:
		dl := node.AsVariableStatement().DeclarationList
		if dl == nil {
			return nil, false
		}
		return varListKind(dl), true
	case ast.KindVariableDeclarationList:
		// for-loop initializer position: the same VariableDeclarationList
		// shape exposes `kind` to ESTree-flavoured selectors (e.g.
		// `ForInStatement[left.kind='const']`).
		return varListKind(node), true
	case ast.KindMethodDeclaration:
		// ESTree splits the `kind` field by container:
		// object literal → 'init' (Property.kind)
		// class body    → 'method' (MethodDefinition.kind)
		if node.Parent != nil {
			switch node.Parent.Kind {
			case ast.KindClassDeclaration, ast.KindClassExpression:
				return "method", true
			}
		}
		return "init", true
	case ast.KindConstructor:
		return "constructor", true
	case ast.KindGetAccessor:
		return "get", true
	case ast.KindSetAccessor:
		return "set", true
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
		return "init", true
	case ast.KindExportDeclaration:
		return "value", true
	}
	return nil, false
}

// varListKind returns the ESTree `kind` string for a VariableDeclarationList.
// `await using` shares bits with const + using, so the order matters —
// check the most specific predicates first via the helper functions
// tsgo exposes for exactly this disambiguation.
func varListKind(dl *ast.Node) string {
	switch {
	case ast.IsVarAwaitUsing(dl):
		return "await using"
	case ast.IsVarUsing(dl):
		return "using"
	case ast.IsVarConst(dl):
		return "const"
	case ast.IsVarLet(dl):
		return "let"
	}
	return "var"
}

func readComputedAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindElementAccessExpression:
		return true, true
	case ast.KindPropertyAccessExpression:
		return false, true
	case ast.KindPropertyAssignment, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindPropertyDeclaration:
		name := node.Name()
		return name != nil && name.Kind == ast.KindComputedPropertyName, true
	}
	return nil, false
}

// readShorthandAttr models ESTree's Property.shorthand boolean.
// tsgo's KindShorthandPropertyAssignment maps directly; everything else
// is non-shorthand.
func readShorthandAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindShorthandPropertyAssignment:
		return true, true
	case ast.KindPropertyAssignment, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return false, true
	}
	return nil, false
}

// readMethodAttr models ESTree's Property.method (true for object-method
// shorthand `({ foo() {} })`). Class methods are MethodDefinition, not
// Property, so the attribute is not exposed there.
func readMethodAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindMethodDeclaration:
		// Only object-literal methods are Property.method=true.
		if node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
			return true, true
		}
		return false, true
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment, ast.KindGetAccessor, ast.KindSetAccessor:
		return false, true
	}
	return nil, false
}

// readSuperClassAttr models ESTree's ClassDeclaration.superClass — the
// expression following `extends`. tsgo expresses extends through a
// HeritageClause with KindExtendsKeyword; the first type in that clause
// is the super-class expression.
func readSuperClassAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		var clauses *ast.NodeList
		if node.Kind == ast.KindClassDeclaration {
			clauses = node.AsClassDeclaration().HeritageClauses
		} else {
			clauses = node.AsClassExpression().HeritageClauses
		}
		if clauses == nil {
			return nil, true
		}
		for _, c := range clauses.Nodes {
			hc := c.AsHeritageClause()
			if hc == nil || hc.Token != ast.KindExtendsKeyword {
				continue
			}
			if hc.Types != nil && len(hc.Types.Nodes) > 0 {
				// `extends Foo` — `Foo` lives in ExpressionWithTypeArguments.
				et := hc.Types.Nodes[0]
				if et.Kind == ast.KindExpressionWithTypeArguments {
					return unwrapExpression(et.AsExpressionWithTypeArguments().Expression), true
				}
				return unwrapExpression(et), true
			}
		}
		return nil, true
	}
	return nil, false
}

// readDirectiveAttr models ESTree's ExpressionStatement.directive, which
// is set on a string-literal ExpressionStatement that appears in a
// directive prologue (top-level "use strict", etc.). tsgo doesn't mark
// directives explicitly; recover the directive text by inspecting the
// preceding-sibling pattern.
func readDirectiveAttr(node *ast.Node, mc *matchContext) (interface{}, bool) {
	if node.Kind != ast.KindExpressionStatement {
		return nil, false
	}
	expr := node.AsExpressionStatement().Expression
	if expr == nil || expr.Kind != ast.KindStringLiteral {
		return nil, true
	}
	if !isDirectivePrologue(node) {
		return nil, true
	}
	// ESTree's directive value is the cooked-string content of the
	// literal (without the surrounding quotes).
	return expr.AsStringLiteral().Text, true
}

// isDirectivePrologue reports whether `stmt` is part of a directive
// prologue — a contiguous run of string-literal ExpressionStatements at
// the start of a SourceFile or a function/class body.
func isDirectivePrologue(stmt *ast.Node) bool {
	parent := stmt.Parent
	if parent == nil {
		return false
	}
	var siblings []*ast.Node
	switch parent.Kind {
	case ast.KindSourceFile:
		s := parent.AsSourceFile().Statements
		if s == nil {
			return false
		}
		siblings = s.Nodes
	case ast.KindBlock:
		s := parent.AsBlock().Statements
		if s == nil {
			return false
		}
		siblings = s.Nodes
	case ast.KindModuleBlock:
		s := parent.AsModuleBlock().Statements
		if s == nil {
			return false
		}
		siblings = s.Nodes
	default:
		return false
	}
	for _, s := range siblings {
		if s.Kind != ast.KindExpressionStatement {
			return false
		}
		expr := s.AsExpressionStatement().Expression
		if expr == nil || expr.Kind != ast.KindStringLiteral {
			return false
		}
		if s == stmt {
			return true
		}
	}
	return false
}

// readUpdateAttr models ESTree's ForStatement.update — tsgo names this
// `Incrementor`.
func readUpdateAttr(node *ast.Node) (interface{}, bool) {
	if node.Kind == ast.KindForStatement {
		return unwrapExpression(node.AsForStatement().Incrementor), true
	}
	return nil, false
}

// readExpressionsAttr / readQuasisAttr expose ESTree's
// TemplateLiteral.expressions / TemplateLiteral.quasis. tsgo splits the
// concept across two kinds: NoSubstitutionTemplateLiteral has no
// expressions, TemplateExpression carries Head + TemplateSpans where
// each span owns one expression and one literal piece.
func readExpressionsAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindNoSubstitutionTemplateLiteral:
		return []*ast.Node{}, true
	case ast.KindTemplateExpression:
		spans := node.AsTemplateExpression().TemplateSpans
		out := []*ast.Node{}
		if spans != nil {
			for _, sp := range spans.Nodes {
				out = append(out, sp.AsTemplateSpan().Expression)
			}
		}
		return out, true
	}
	return nil, false
}

func readQuasisAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindNoSubstitutionTemplateLiteral:
		return []*ast.Node{node}, true
	case ast.KindTemplateExpression:
		te := node.AsTemplateExpression()
		out := []*ast.Node{te.Head}
		if te.TemplateSpans != nil {
			for _, sp := range te.TemplateSpans.Nodes {
				out = append(out, sp.AsTemplateSpan().Literal)
			}
		}
		return out, true
	}
	return nil, false
}

// readBigintAttr models ESTree's Literal.bigint — present for BigInt
// literals (`1n`). The string value is the digits without the trailing
// `n`.
// readDelegateAttr models ESTree's YieldExpression.delegate (true for
// `yield*`, false for plain `yield`).
func readDelegateAttr(node *ast.Node) (interface{}, bool) {
	if node.Kind != ast.KindYieldExpression {
		return nil, false
	}
	return node.AsYieldExpression().AsteriskToken != nil, true
}

func readBigintAttr(node *ast.Node, mc *matchContext) (interface{}, bool) {
	if node.Kind != ast.KindBigIntLiteral {
		return nil, false
	}
	return utils.NormalizeBigIntLiteral(node.AsBigIntLiteral().Text), true
}

func readStaticAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindPropertyDeclaration:
		return ast.HasStaticModifier(node), true
	}
	return nil, false
}

func readLabelAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindBreakStatement:
		return node.AsBreakStatement().Label, true
	case ast.KindContinueStatement:
		return node.AsContinueStatement().Label, true
	case ast.KindLabeledStatement:
		return node.AsLabeledStatement().Label, true
	}
	return nil, false
}

// readRegexObject returns a synthetic *ast.Node-equivalent value for the
// `regex` ESTree field. Concretely we return the node itself and let
// readRegexFlags / readRegexPattern interpret it during further lookups.
func readRegexObject(node *ast.Node, mc *matchContext) (interface{}, bool) {
	if node.Kind != ast.KindRegularExpressionLiteral {
		return nil, false
	}
	return regexFacade{node: node, mc: mc}, true
}

// regexFacade represents an ESTree `regex` object — a {pattern, flags}
// pair extracted from the regex literal source. It is opaque to the
// matcher except via subsequent path segments (`.flags`, `.pattern`).
type regexFacade struct {
	node *ast.Node
	mc   *matchContext
}

func readRegexFlags(node *ast.Node, mc *matchContext) (interface{}, bool) {
	if node.Kind == ast.KindRegularExpressionLiteral {
		_, flags := splitRegexLiteral(regexLiteralText(node, mc))
		return flags, true
	}
	return nil, false
}

func readRegexPattern(node *ast.Node, mc *matchContext) (interface{}, bool) {
	if node.Kind == ast.KindRegularExpressionLiteral {
		pat, _ := splitRegexLiteral(regexLiteralText(node, mc))
		return pat, true
	}
	return nil, false
}

func splitRegexLiteral(text string) (string, string) {
	if !strings.HasPrefix(text, "/") {
		return text, ""
	}
	last := strings.LastIndex(text, "/")
	if last <= 0 {
		return text, ""
	}
	return text[1:last], text[last+1:]
}

func regexLiteralText(node *ast.Node, mc *matchContext) string {
	if node == nil {
		return ""
	}
	if mc != nil && mc.sf != nil {
		return scanner.GetSourceTextOfNodeFromSourceFile(mc.sf, node, false)
	}
	return node.Text()
}

func readParamsAttr(node *ast.Node) (interface{}, bool) {
	if !isFunctionLikeForParams(node) || isFunctionValueNode(node) {
		return nil, false
	}
	params := utils.ESTreeParameters(node)
	// Treat nil and empty as the same — `[params.length=0]` should
	// match a function-like with no parameters regardless of whether
	// the underlying NodeList was allocated.
	if params == nil {
		return []*ast.Node{}, true
	}
	return params, true
}

func readSourceAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindImportDeclaration:
		return unwrapExpression(node.AsImportDeclaration().ModuleSpecifier), true
	case ast.KindExportDeclaration:
		return unwrapExpression(node.AsExportDeclaration().ModuleSpecifier), true
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call.Expression != nil && call.Expression.Kind == ast.KindImportKeyword && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			return unwrapExpression(call.Arguments.Nodes[0]), true
		}
	}
	return nil, false
}

func readCalleeAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindCallExpression:
		expression := node.AsCallExpression().Expression
		if expression != nil && expression.Kind == ast.KindImportKeyword {
			return nil, false
		}
		return unwrapExpression(expression), true
	case ast.KindNewExpression:
		return unwrapExpression(node.AsNewExpression().Expression), true
	case ast.KindTaggedTemplateExpression:
		return unwrapExpression(node.AsTaggedTemplateExpression().Tag), true
	}
	return nil, false
}

func readArgumentsAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call.Expression != nil && call.Expression.Kind == ast.KindImportKeyword {
			return nil, false
		}
		args := call.Arguments
		if args == nil {
			return []*ast.Node{}, true
		}
		return args.Nodes, true
	case ast.KindNewExpression:
		args := node.AsNewExpression().Arguments
		if args == nil {
			return []*ast.Node{}, true
		}
		return args.Nodes, true
	}
	return nil, false
}

func readExpressionAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindDecorator:
		return unwrapEstreeNode(node.AsDecorator().Expression), true
	case ast.KindExpressionStatement:
		return unwrapEstreeNode(node.AsExpressionStatement().Expression), true
	case ast.KindAsExpression:
		return unwrapEstreeNode(node.AsAsExpression().Expression), true
	case ast.KindSatisfiesExpression:
		return unwrapEstreeNode(node.AsSatisfiesExpression().Expression), true
	case ast.KindNonNullExpression:
		return unwrapEstreeNode(node.AsNonNullExpression().Expression), true
	case ast.KindTypeAssertionExpression:
		return unwrapEstreeNode(node.AsTypeAssertion().Expression), true
	case ast.KindArrowFunction:
		return node.Body() != nil && node.Body().Kind != ast.KindBlock, true
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		return false, true
	case ast.KindJsxExpression:
		if isEmptyJSXExpression(node) {
			return virtualNodeFacade{node, "JSXEmptyExpression"}, true
		}
		return unwrapEstreeNode(node.AsJsxExpression().Expression), true
	}
	return nil, false
}

// virtualNodeFacade models an ESTree wrapper that tsgo folds into node. The
// underlying node remains available for fields shared with the physical AST,
// while type and parent preserve the wrapper's identity.
type virtualNodeFacade struct {
	node     *ast.Node
	typeName string
}

func readVirtualNodeAttr(facade virtualNodeFacade, name string, mc *matchContext) (interface{}, bool) {
	node := facade.node
	if node == nil {
		return nil, false
	}
	switch facade.typeName {
	case "ChainExpression":
		if name == "expression" {
			return node, true
		}
	case "Literal":
		if node.Kind == ast.KindConstructor && name == "value" {
			return "constructor", true
		}
		if node.Kind == ast.KindConstructor && name == "raw" && mc != nil && mc.sf != nil {
			if token, ok := utils.TokenBeforePosition(mc.sf, node.ParameterList().Pos()-1); ok {
				return token.Text, true
			}
		}
	case "Identifier":
		if node.Kind == ast.KindConstructor && name == "name" {
			return "constructor", true
		}
	case "ClassBody":
		if name == "body" {
			return classMembers(node), true
		}
		if name == "type" || name == "parent" {
			return nil, false
		}
	case "TSEnumBody":
		if name == "members" {
			members := node.AsEnumDeclaration().Members
			if members == nil {
				return []*ast.Node{}, true
			}
			return members.Nodes, true
		}
		if name == "type" || name == "parent" {
			return nil, false
		}
	case "JSXEmptyExpression":
		if name == "type" || name == "parent" {
			return nil, false
		}
	case "JSXOpeningElement":
		if name == "selfClosing" {
			return true, true
		}
		if name == "name" || name == "tagName" {
			return node.AsJsxSelfClosingElement().TagName, true
		}
		if name == "attributes" {
			attributes := node.AsJsxSelfClosingElement().Attributes
			if attributes == nil {
				return []*ast.Node{}, true
			}
			properties := attributes.AsJsxAttributes().Properties
			if properties == nil {
				return []*ast.Node{}, true
			}
			return properties.Nodes, true
		}
		if name == "type" || name == "parent" {
			return nil, false
		}
	case "FunctionExpression", "TSEmptyBodyFunctionExpression":
		switch name {
		case "body":
			return node.Body(), true
		case "params":
			return utils.ESTreeParameters(node), true
		case "async":
			return ast.IsAsyncFunction(node), true
		case "generator":
			return methodValueIsGenerator(node), true
		case "expression":
			return false, true
		case "id":
			return nil, true
		case "type", "parent":
			return nil, false
		}
	}
	return nil, false
}

func methodValueIsGenerator(node *ast.Node) bool {
	return node.Kind == ast.KindMethodDeclaration && node.AsMethodDeclaration().AsteriskToken != nil
}

func readInitAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindVariableDeclaration:
		return unwrapExpression(node.AsVariableDeclaration().Initializer), true
	case ast.KindForStatement:
		return unwrapExpression(node.AsForStatement().Initializer), true
	case ast.KindBindingElement:
		return unwrapExpression(node.AsBindingElement().Initializer), true
	case ast.KindEnumMember:
		return unwrapExpression(node.AsEnumMember().Initializer), true
	}
	return nil, false
}

func readIdAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Name(), true
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Name(), true
	case ast.KindClassDeclaration:
		return node.AsClassDeclaration().Name(), true
	case ast.KindClassExpression:
		return node.AsClassExpression().Name(), true
	case ast.KindVariableDeclaration:
		return node.AsVariableDeclaration().Name(), true
	case ast.KindEnumDeclaration:
		return node.AsEnumDeclaration().Name(), true
	case ast.KindEnumMember:
		return node.AsEnumMember().Name(), true
	}
	return nil, false
}

func readKeyAttr(node *ast.Node, mc *matchContext) (interface{}, bool) {
	switch node.Kind {
	case ast.KindConstructor:
		return constructorKey(node, mc), true
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindPropertyDeclaration:
		return node.Name(), true
	case ast.KindImportAttribute:
		return node.AsImportAttribute().Name(), true
	}
	return nil, false
}

func readLeftAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindBinaryExpression:
		return unwrapExpression(node.AsBinaryExpression().Left), true
	case ast.KindForInStatement, ast.KindForOfStatement:
		return unwrapExpression(node.AsForInOrOfStatement().Initializer), true
	}
	return nil, false
}

func readRightAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindBinaryExpression:
		return unwrapExpression(node.AsBinaryExpression().Right), true
	case ast.KindForInStatement, ast.KindForOfStatement:
		return unwrapExpression(node.AsForInOrOfStatement().Expression), true
	}
	return nil, false
}

func readObjectAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		object := unwrapExpression(node.AsPropertyAccessExpression().Expression)
		if isJSXMemberExpression(node) {
			return object, true
		}
		return object, true
	case ast.KindElementAccessExpression:
		return unwrapExpression(node.AsElementAccessExpression().Expression), true
	}
	return nil, false
}

func readPropertyAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression().Name()
		if isJSXMemberExpression(node) {
			return property, true
		}
		return property, true
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().ArgumentExpression, true
	case ast.KindMetaProperty:
		// ESTree's MetaProperty.property is the trailing identifier
		// (e.g. `target` in `new.target`, `meta` in `import.meta`).
		// tsgo stores it on the unexported `name` field; the public
		// accessor is `Name()`.
		return node.AsMetaProperty().Name(), true
	}
	return nil, false
}

func readTestAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindIfStatement:
		return unwrapExpression(node.AsIfStatement().Expression), true
	case ast.KindConditionalExpression:
		return unwrapExpression(node.AsConditionalExpression().Condition), true
	case ast.KindWhileStatement:
		return unwrapExpression(node.AsWhileStatement().Expression), true
	case ast.KindDoStatement:
		return unwrapExpression(node.AsDoStatement().Expression), true
	case ast.KindForStatement:
		return unwrapExpression(node.AsForStatement().Condition), true
	case ast.KindCaseClause:
		// ESTree's SwitchCase.test is the case expression. tsgo names
		// the same field `Expression` on a CaseOrDefaultClause.
		return unwrapExpression(node.AsCaseOrDefaultClause().Expression), true
	case ast.KindDefaultClause:
		// SwitchCase for `default:` has test === null in ESTree.
		return nil, true
	}
	return nil, false
}

func readConsequentAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindIfStatement:
		return node.AsIfStatement().ThenStatement, true
	case ast.KindConditionalExpression:
		return node.AsConditionalExpression().WhenTrue, true
	case ast.KindCaseClause, ast.KindDefaultClause:
		// ESTree's SwitchCase.consequent is the statement list.
		stmts := node.AsCaseOrDefaultClause().Statements
		if stmts == nil {
			return []*ast.Node{}, true
		}
		return stmts.Nodes, true
	}
	return nil, false
}

func readAlternateAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindIfStatement:
		return node.AsIfStatement().ElseStatement, true
	case ast.KindConditionalExpression:
		return node.AsConditionalExpression().WhenFalse, true
	}
	return nil, false
}

func readBodyAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindClassStaticBlockDeclaration:
		return statements(node.AsClassStaticBlockDeclaration().Body.AsBlock().Statements), true
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Body, true
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Body, true
	case ast.KindArrowFunction:
		return node.AsArrowFunction().Body, true
	case ast.KindIfStatement:
		return node.AsIfStatement().ThenStatement, true
	case ast.KindWhileStatement:
		return node.AsWhileStatement().Statement, true
	case ast.KindDoStatement:
		return node.AsDoStatement().Statement, true
	case ast.KindForStatement:
		return node.AsForStatement().Statement, true
	case ast.KindBlock:
		stmts := node.AsBlock().Statements
		if stmts == nil {
			return []*ast.Node{}, true
		}
		return stmts.Nodes, true
	case ast.KindSourceFile:
		stmts := node.AsSourceFile().Statements
		if stmts == nil {
			return []*ast.Node{}, true
		}
		return stmts.Nodes, true
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return virtualNodeFacade{node, "ClassBody"}, true
	case ast.KindEnumDeclaration:
		return virtualNodeFacade{node: node, typeName: "TSEnumBody"}, true
	}
	return nil, false
}

func readArgumentAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindAwaitExpression:
		return unwrapExpression(node.AsAwaitExpression().Expression), true
	case ast.KindYieldExpression:
		return unwrapExpression(node.AsYieldExpression().Expression), true
	case ast.KindSpreadElement:
		return unwrapExpression(node.AsSpreadElement().Expression), true
	case ast.KindSpreadAssignment:
		return unwrapExpression(node.AsSpreadAssignment().Expression), true
	case ast.KindReturnStatement:
		return unwrapExpression(node.AsReturnStatement().Expression), true
	case ast.KindThrowStatement:
		return unwrapExpression(node.AsThrowStatement().Expression), true
	case ast.KindPrefixUnaryExpression:
		return unwrapExpression(node.AsPrefixUnaryExpression().Operand), true
	case ast.KindPostfixUnaryExpression:
		return unwrapExpression(node.AsPostfixUnaryExpression().Operand), true
	case ast.KindTypeOfExpression:
		return unwrapExpression(node.AsTypeOfExpression().Expression), true
	case ast.KindVoidExpression:
		return unwrapExpression(node.AsVoidExpression().Expression), true
	case ast.KindDeleteExpression:
		return unwrapExpression(node.AsDeleteExpression().Expression), true
	}
	return nil, false
}

func readPrefixAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindPrefixUnaryExpression:
		return true, true
	case ast.KindPostfixUnaryExpression:
		return false, true
	}
	return nil, false
}

func readAsyncAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		return ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync), true
	}
	return nil, false
}

func readGeneratorAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().AsteriskToken != nil, true
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().AsteriskToken != nil, true
	case ast.KindArrowFunction:
		return false, true
	}
	return nil, false
}

func readSpecifiersAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindImportDeclaration:
		return collectImportSpecifiers(node), true
	case ast.KindExportDeclaration:
		return collectExportSpecifiers(node), true
	}
	return nil, false
}

func collectImportSpecifiers(node *ast.Node) []*ast.Node {
	out := []*ast.Node{}
	clause := node.AsImportDeclaration().ImportClause
	if clause == nil {
		return out
	}
	c := clause.AsImportClause()
	if c == nil {
		return out
	}
	if c.Name() != nil {
		out = append(out, clause)
	}
	if c.NamedBindings != nil {
		switch c.NamedBindings.Kind {
		case ast.KindNamespaceImport:
			out = append(out, c.NamedBindings)
		case ast.KindNamedImports:
			ni := c.NamedBindings.AsNamedImports()
			if ni != nil && ni.Elements != nil {
				out = append(out, ni.Elements.Nodes...)
			}
		}
	}
	return out
}

func collectExportSpecifiers(node *ast.Node) []*ast.Node {
	out := []*ast.Node{}
	ed := node.AsExportDeclaration()
	if ed == nil || ed.ExportClause == nil {
		return out
	}
	if ed.ExportClause.Kind == ast.KindNamedExports {
		ne := ed.ExportClause.AsNamedExports()
		if ne != nil && ne.Elements != nil {
			out = append(out, ne.Elements.Nodes...)
		}
	}
	return out
}
