package no_restricted_syntax

import (
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
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
	case identifierSelector:
		if v.Name == "*" {
			s.markUniverse()
			return
		}
		ks := kindsForEstreeName(v.Name)
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
	switch v := sel.(type) {
	case identifierSelector:
		return matchesIdentifier(v.Name, node)
	case classSelector:
		if !matches(v.Inner, node, mc) {
			return false
		}
		return matchesClass(node, v.Class)
	case attrSelector:
		if !matches(v.Inner, node, mc) {
			return false
		}
		return matchesAttr(node, v.Path, v.Op, v.Value, mc)
	case combinatorSelector:
		if !matches(v.Right, node, mc) {
			return false
		}
		return matchesCombinator(v, node, mc)
	case pseudoSelector:
		return matchesPseudo(v, node, mc)
	case combinedPseudo:
		if !matches(v.Inner, node, mc) {
			return false
		}
		return matchesPseudo(v.Pseudo, node, mc)
	case unionSelector:
		for _, a := range v.Selectors {
			if matches(a, node, mc) {
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
func matchesIdentifier(name string, node *ast.Node) bool {
	if name == "*" {
		return true
	}
	ks := kindsForEstreeName(name)
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
	return refineEstreeMatch(name, node)
}

// refineEstreeMatch tightens the tsgo→ESTree match for kinds that tsgo
// fuses but ESTree splits. For example, tsgo's BinaryExpression covers the
// ESTree triplet of BinaryExpression / LogicalExpression / AssignmentExpression
// / SequenceExpression — the operator decides which ESTree form it really is.
func refineEstreeMatch(name string, node *ast.Node) bool {
	switch name {
	case "BinaryExpression":
		return node.Kind == ast.KindBinaryExpression && isPlainBinaryOperator(node)
	case "LogicalExpression":
		return node.Kind == ast.KindBinaryExpression && isLogicalOperator(node)
	case "AssignmentExpression":
		return node.Kind == ast.KindBinaryExpression && isAssignmentOperator(node)
	case "SequenceExpression":
		return node.Kind == ast.KindBinaryExpression && isCommaOperator(node)
	case "ChainExpression":
		// ESTree wraps the entire optional chain in a single
		// ChainExpression node. tsgo has no analogue, so we match the
		// outermost link in the chain — the highest node that is itself
		// part of the chain but whose parent is not.
		// `ast.IsOutermostOptionalChain` is too permissive for this
		// purpose (it's true for the link directly under any `?.`-token
		// wrapper, which still emits an inner Match in `a?.b?.()`).
		return ast.IsOptionalChain(node) && (node.Parent == nil || !ast.IsOptionalChain(node.Parent))
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
			return pk == ast.KindClassDeclaration || pk == ast.KindClassExpression
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

// matchesClass evaluates `Foo.bar` — Foo already matched by the inner
// selector, here we check that the node sits at the named field of its
// parent.
func matchesClass(node *ast.Node, class string) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	return nodeIsAtField(node, parent, class)
}

// matchesAttr resolves the attribute path against the node and compares
// the obtained value with the operator/right-hand side of the selector.
func matchesAttr(node *ast.Node, path []string, op attrOp, value attrValue, mc *matchContext) bool {
	val, ok := lookupAttrPath(node, path, mc)
	if op == attrPresent {
		// Presence: any non-zero, non-nil value passes.
		if !ok {
			return false
		}
		return attrTruthy(val)
	}
	if !ok {
		return false
	}
	return compareAttr(val, op, value)
}

// matchesCombinator handles `>` (parent), descendant, `+` (prev sibling),
// `~` (any prior sibling). The right-hand side has already been verified
// against the current node before this function is called.
func matchesCombinator(c combinatorSelector, node *ast.Node, mc *matchContext) bool {
	switch c.Kind {
	case combChild:
		// ESTree wraps `export default <decl>` in an
		// ExportDefaultDeclaration node. tsgo flattens this — a default-
		// exported FunctionDeclaration / ClassDeclaration sits directly
		// under SourceFile with `export default` modifiers. To honour
		// `ExportDefaultDeclaration > FunctionDeclaration`-style
		// selectors, treat such a declaration as if it had a virtual
		// ExportDefaultDeclaration parent.
		if isDefaultExportedDeclaration(node) {
			if selectorMatchesVirtualExportDefault(c.Left) {
				return true
			}
		}
		parent := node.Parent
		if parent == nil {
			return false
		}
		return matches(c.Left, parent, mc)
	case combDescendant:
		if isDefaultExportedDeclaration(node) && selectorMatchesVirtualExportDefault(c.Left) {
			return true
		}
		current := node.Parent
		for current != nil {
			if matches(c.Left, current, mc) {
				return true
			}
			current = current.Parent
		}
		return false
	case combAdjacent, combSibling:
		parent := node.Parent
		if parent == nil {
			return false
		}
		// esquery's sibling combinators only consider list-position
		// siblings — nodes sharing the same NodeList field on the
		// parent. Two scalar children of the same parent (e.g.
		// VariableDeclaration's `id` and `init`) are NOT siblings.
		var siblings []*ast.Node
		for _, list := range listChildrenOf(parent) {
			for _, child := range list {
				if child == node {
					siblings = list
					break
				}
			}
			if siblings != nil {
				break
			}
		}
		if siblings == nil {
			return false
		}
		idx := -1
		for i, sib := range siblings {
			if sib == node {
				idx = i
				break
			}
		}
		if idx <= 0 {
			return false
		}
		if c.Kind == combAdjacent {
			return matches(c.Left, siblings[idx-1], mc)
		}
		for i := idx - 1; i >= 0; i-- {
			if matches(c.Left, siblings[i], mc) {
				return true
			}
		}
		return false
	}
	return false
}

func matchesPseudo(p pseudoSelector, node *ast.Node, mc *matchContext) bool {
	switch p.Name {
	case "is", "matches":
		for _, a := range p.Args {
			if matches(a, node, mc) {
				return true
			}
		}
		return false
	case "not":
		for _, a := range p.Args {
			if matches(a, node, mc) {
				return false
			}
		}
		return true
	case "has":
		for _, a := range p.Args {
			if hasDescendantMatching(node, a, mc) {
				return true
			}
		}
		return false
	case "nth-child":
		idx, _ := nodeIndexInListField(node)
		return idx >= 0 && idx == p.N-1
	case "nth-last-child":
		idx, total := nodeIndexInListField(node)
		if idx < 0 {
			return false
		}
		return total-idx == p.N
	}
	return false
}

func hasDescendantMatching(node *ast.Node, sel selector, mc *matchContext) bool {
	found := false
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		if found {
			return true
		}
		n.ForEachChild(func(child *ast.Node) bool {
			if matches(sel, child, mc) {
				found = true
				return true
			}
			return visit(child)
		})
		return found
	}
	visit(node)
	return found
}

// nodeIsAtField returns whether `node` is positioned at the named ESTree
// field on `parent`. Resolution uses the field-mapping helpers (see
// fields.go).
func nodeIsAtField(node, parent *ast.Node, field string) bool {
	candidates := nodesAtField(parent, field)
	for _, c := range candidates {
		if c == node {
			return true
		}
	}
	return false
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
	if n, ok := current.(*ast.Node); ok {
		return readNodeAttr(n, segment, mc)
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
	// Primitive interpretations: support `.length` on strings / slices.
	if segment == "length" {
		switch v := current.(type) {
		case string:
			return float64(len([]rune(v))), true
		case []*ast.Node:
			return float64(len(v)), true
		}
	}
	return nil, false
}

// attrTruthy evaluates the truthiness of an attribute value the same way
// JavaScript would, so that bare attribute presence (`[label]`) only
// matches when the value is non-empty / non-zero / non-nil.
func attrTruthy(v interface{}) bool {
	if v == nil {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return x != ""
	case float64:
		return x != 0
	case *ast.Node:
		return x != nil
	case []*ast.Node:
		return len(x) > 0
	}
	return true
}

func compareAttr(left interface{}, op attrOp, right attrValue) bool {
	switch op {
	case attrEqual:
		return attrEquals(left, right)
	case attrNotEqual:
		return !attrEquals(left, right)
	case attrLess, attrLessOrEqual, attrGreater, attrGreaterOrEqual:
		l, ok := numericFromAttr(left)
		if !ok {
			return false
		}
		var r float64
		switch right.Kind {
		case attrValueNumber:
			r = right.Num
		default:
			return false
		}
		switch op {
		case attrLess:
			return l < r
		case attrLessOrEqual:
			return l <= r
		case attrGreater:
			return l > r
		case attrGreaterOrEqual:
			return l >= r
		}
	}
	return false
}

func numericFromAttr(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case string:
		n, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	}
	return 0, false
}

func attrEquals(left interface{}, right attrValue) bool {
	switch right.Kind {
	case attrValueString:
		return attrAsString(left) == right.Str
	case attrValueNumber:
		n, ok := numericFromAttr(left)
		return ok && n == right.Num
	case attrValueBool:
		b, ok := left.(bool)
		return ok && b == right.Bool
	case attrValueNull:
		return left == nil
	case attrValueIdent:
		return attrAsString(left) == right.Ident
	case attrValueRegex:
		s := attrAsString(left)
		re, err := regexp2.Compile(right.Regex, regexpFlags(right.Flags))
		if err != nil {
			return false
		}
		ok, _ := re.MatchString(s)
		return ok
	}
	return false
}

func regexpFlags(flags string) regexp2.RegexOptions {
	var opts regexp2.RegexOptions
	for _, c := range flags {
		switch c {
		case 'i':
			opts |= regexp2.IgnoreCase
		case 's':
			opts |= regexp2.Singleline
		case 'm':
			opts |= regexp2.Multiline
		}
	}
	return opts
}

func attrAsString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		// Match JavaScript's "1" not "1.0" formatting.
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'g', -1, 64)
	case nil:
		return ""
	}
	return ""
}

// nodeIndexInListField returns (idx, total) for a node that sits inside
// a NodeList field of its parent. esquery's `:nth-child` only counts
// positions within array-shaped fields (e.g. `Program.body`,
// `CallExpression.arguments`) — a node sitting at a scalar field
// (`ExpressionStatement.expression`, `MemberExpression.object`) is not
// considered a positional child and never matches.
func nodeIndexInListField(node *ast.Node) (int, int) {
	parent := node.Parent
	if parent == nil {
		return -1, 0
	}
	for _, list := range listChildrenOf(parent) {
		for i, child := range list {
			if child == node {
				return i, len(list)
			}
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
	if node == nil {
		return nil
	}
	for {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		default:
			return node
		}
		if node == nil {
			return nil
		}
	}
}

// readNodeAttr extracts the named ESTree-style attribute from a tsgo node.
// Centralising the field map here makes it easy to widen support without
// rewriting the matcher.
func readNodeAttr(node *ast.Node, name string, mc *matchContext) (interface{}, bool) {
	switch name {
	case "type":
		return estreeNameForKind(node), true
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
		// ESTree's `optional` is true only on the link that carries the
		// literal `?.` token (e.g. `foo?.bar`, not the trailing `.baz`
		// in `foo?.bar.baz`). tsgo's IsOptionalChainRoot encodes the
		// same condition.
		return ast.IsOptionalChainRoot(node), true
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
		return readKeyAttr(node)
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
	if node.Kind == ast.KindJsxElement {
		return node.AsJsxElement().OpeningElement, true
	}
	return nil, false
}

func readClosingElementAttr(node *ast.Node) (interface{}, bool) {
	if node.Kind == ast.KindJsxElement {
		return node.AsJsxElement().ClosingElement, true
	}
	return nil, false
}

func readAttributesAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindJsxOpeningElement:
		return node.AsJsxOpeningElement().Attributes, true
	case ast.KindJsxSelfClosingElement:
		return node.AsJsxSelfClosingElement().Attributes, true
	}
	return nil, false
}

func readChildrenAttr(node *ast.Node) (interface{}, bool) {
	if node.Kind == ast.KindJsxElement {
		c := node.AsJsxElement().Children
		if c == nil {
			return []*ast.Node{}, true
		}
		return append([]*ast.Node{}, c.Nodes...), true
	}
	return nil, false
}

// readTagNameAttr returns the JSX tag-name node — esquery exposes this as
// `name` on JSXOpeningElement / JSXSelfClosingElement, but ESLint also
// understands `tagName` for direct compatibility with TSGo-style ASTs.
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
				return "MethodDefinition"
			case ast.KindObjectLiteralExpression:
				return "Property"
			}
		}
		return "Property"
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
	case ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression:
		return estreeNameForKind(unwrapExpression(node))
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
	case ast.KindMetaProperty:
		return "MetaProperty"
	case ast.KindPropertyDeclaration:
		return "PropertyDefinition"
	case ast.KindConstructor:
		return "MethodDefinition"
	case ast.KindClassStaticBlockDeclaration:
		return "StaticBlock"
	case ast.KindArrayBindingPattern:
		return "ArrayPattern"
	case ast.KindObjectBindingPattern:
		return "ObjectPattern"
	}
	return ""
}

func readNameAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text, true
	case ast.KindPrivateIdentifier:
		return node.AsPrivateIdentifier().Text, true
	// JSX: ESTree calls the tag identifier `name`; tsgo stores it as
	// TagName. Also expose it through this attribute path.
	case ast.KindJsxOpeningElement:
		return node.AsJsxOpeningElement().TagName, true
	case ast.KindJsxSelfClosingElement:
		return node.AsJsxSelfClosingElement().TagName, true
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
		return node.AsBigIntLiteral().Text, true
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
	}
	return nil, false
}

func readRawAttr(node *ast.Node, mc *matchContext) (interface{}, bool) {
	if mc == nil || mc.sf == nil {
		return nil, false
	}
	return scanner.GetSourceTextOfNodeFromSourceFile(mc.sf, node, false), true
}

func readOperatorAttr(node *ast.Node) (interface{}, bool) {
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
	return nil, false
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
	text := node.AsBigIntLiteral().Text
	if mc != nil && mc.sf != nil {
		text = scanner.GetSourceTextOfNodeFromSourceFile(mc.sf, node, false)
	}
	if len(text) > 0 && text[len(text)-1] == 'n' {
		text = text[:len(text)-1]
	}
	return text, true
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
	if !isFunctionLikeForParams(node) {
		return nil, false
	}
	params := node.Parameters()
	// Treat nil and empty as the same — `[params.length=0]` should
	// match a function-like with no parameters regardless of whether
	// the underlying NodeList was allocated.
	out := make([]*ast.Node, 0, len(params))
	out = append(out, params...)
	return out, true
}

func readSourceAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindImportDeclaration:
		return unwrapExpression(node.AsImportDeclaration().ModuleSpecifier), true
	case ast.KindExportDeclaration:
		return unwrapExpression(node.AsExportDeclaration().ModuleSpecifier), true
	}
	return nil, false
}

func readCalleeAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindCallExpression:
		return unwrapExpression(node.AsCallExpression().Expression), true
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
		args := node.AsCallExpression().Arguments
		if args == nil {
			return []*ast.Node{}, true
		}
		return append([]*ast.Node{}, args.Nodes...), true
	case ast.KindNewExpression:
		args := node.AsNewExpression().Arguments
		if args == nil {
			return []*ast.Node{}, true
		}
		return append([]*ast.Node{}, args.Nodes...), true
	}
	return nil, false
}

func readExpressionAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindExpressionStatement:
		return unwrapExpression(node.AsExpressionStatement().Expression), true
	case ast.KindParenthesizedExpression:
		return unwrapExpression(node.AsParenthesizedExpression().Expression), true
	case ast.KindAwaitExpression:
		return unwrapExpression(node.AsAwaitExpression().Expression), true
	case ast.KindYieldExpression:
		return unwrapExpression(node.AsYieldExpression().Expression), true
	case ast.KindReturnStatement:
		return unwrapExpression(node.AsReturnStatement().Expression), true
	case ast.KindThrowStatement:
		return unwrapExpression(node.AsThrowStatement().Expression), true
	case ast.KindCallExpression:
		return unwrapExpression(node.AsCallExpression().Expression), true
	case ast.KindNewExpression:
		return unwrapExpression(node.AsNewExpression().Expression), true
	}
	return nil, false
}

func readInitAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindVariableDeclaration:
		return unwrapExpression(node.AsVariableDeclaration().Initializer), true
	case ast.KindForStatement:
		return unwrapExpression(node.AsForStatement().Initializer), true
	case ast.KindBindingElement:
		return unwrapExpression(node.AsBindingElement().Initializer), true
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
	}
	return nil, false
}

func readKeyAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindPropertyDeclaration:
		return node.Name(), true
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
		return unwrapExpression(node.AsPropertyAccessExpression().Expression), true
	case ast.KindElementAccessExpression:
		return unwrapExpression(node.AsElementAccessExpression().Expression), true
	}
	return nil, false
}

func readPropertyAttr(node *ast.Node) (interface{}, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Name(), true
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
		return append([]*ast.Node{}, stmts.Nodes...), true
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
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Body, true
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Body, true
	case ast.KindArrowFunction:
		return node.AsArrowFunction().Body, true
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().Body, true
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
		return append([]*ast.Node{}, stmts.Nodes...), true
	case ast.KindSourceFile:
		stmts := node.AsSourceFile().Statements
		if stmts == nil {
			return []*ast.Node{}, true
		}
		return append([]*ast.Node{}, stmts.Nodes...), true
	case ast.KindClassDeclaration:
		members := node.AsClassDeclaration().Members
		if members == nil {
			return []*ast.Node{}, true
		}
		return append([]*ast.Node{}, members.Nodes...), true
	case ast.KindClassExpression:
		members := node.AsClassExpression().Members
		if members == nil {
			return []*ast.Node{}, true
		}
		return append([]*ast.Node{}, members.Nodes...), true
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
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration:
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
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().AsteriskToken != nil, true
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
