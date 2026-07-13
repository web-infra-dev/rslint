package utils

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

type ParsedJestFnCall struct {
	Name            string
	LocalName       string
	Kind            JestFnType
	Members         []string
	MemberEntries   []ParsedJestFnMemberEntry
	Modifiers       []string
	ModifierEntries []ParsedJestFnMemberEntry
	Matcher         string
	MatcherEntry    *ParsedJestFnMemberEntry
	Head            ParsedJestFnCallHead
}

type ParsedJestFnCallHead struct {
	Type     JestImportMode
	Local    ParsedJestFnCallHeadEntry
	Original ParsedJestFnCallHeadEntry
}

type ParsedJestFnCallHeadEntry struct {
	Value string
	Node  *ast.Node
}

const (
	ExpectParseReasonNone            = ""
	ExpectParseReasonMatcherNotFound = "matcher-not-found"
	ExpectParseReasonModifierUnknown = "modifier-unknown"
	jestGlobalsModule                = "@jest/globals"
)

func IsTypeOfJestFnCall(node *ast.Node, ctx rule.RuleContext, kinds ...JestFnType) bool {
	parsed := ParseJestFnCall(node, ctx)
	if parsed == nil || len(kinds) == 0 {
		return false
	}

	return slices.Contains(kinds, parsed.Kind)
}

func ParseJestFnCall(node *ast.Node, ctx rule.RuleContext) *ParsedJestFnCall {
	if node == nil || node.Kind != ast.KindCallExpression {
		return nil
	}

	memberEntries := GetJestFnMemberEntries(node)
	if len(memberEntries) == 0 {
		return nil
	}

	localName := memberEntries[0].Name
	members := make([]string, 0, len(memberEntries)-1)
	for _, entry := range memberEntries[1:] {
		members = append(members, entry.Name)
	}

	callExpr := node.AsCallExpression()
	if isEachFactoryCall(callExpr, members) || isInvalidTaggedTemplateCall(callExpr, members) || isInnerExpectCall(node, localName, members, ctx.Settings) {
		return nil
	}

	localNode := resolveHeadLocalNode(callExpr)
	name, originalNode, headType := ResolveJestFunctionReference(node, localName, localNode, ctx)
	if name == "" {
		return nil
	}
	name = ApplyGlobalJestAlias(name, ctx.Settings)
	if !JEST_METHOD_NAMES[name] {
		return nil
	}

	kind := GetJestKind(name)
	if kind == JestFnTypeUnknown {
		return nil
	}
	if kind != JestFnTypeExpect && kind != JestFnTypeJest && !isValidJestCall(name, members) {
		return nil
	}

	parsed := &ParsedJestFnCall{
		Name:          name,
		LocalName:     localName,
		Kind:          kind,
		Members:       members,
		MemberEntries: memberEntries[1:],
		Head: ParsedJestFnCallHead{
			Type: headType,
			Local: ParsedJestFnCallHeadEntry{
				Value: localName,
				Node:  localNode,
			},
			Original: ParsedJestFnCallHeadEntry{
				Value: name,
				Node:  originalNode,
			},
		},
	}

	if kind == JestFnTypeExpect {
		if !applyParsedExpectCall(parsed) {
			return nil
		}
	}

	return parsed
}

// FindTopMostCallExpression walks up member/call chains to the outermost CallExpression,
// matching eslint-plugin-jest's findTopMostCallExpression.
func FindTopMostCallExpression(node *ast.Node) *ast.Node {
	if node == nil || node.Kind != ast.KindCallExpression {
		return node
	}

	top := node
	parent := node.Parent
	for parent != nil {
		if parent.Kind == ast.KindCallExpression {
			top = parent
			parent = parent.Parent
			continue
		}
		if parent.Kind != ast.KindPropertyAccessExpression &&
			parent.Kind != ast.KindElementAccessExpression {
			break
		}
		parent = parent.Parent
	}

	return top
}

func FindImportDeclaration(node *ast.Node) *ast.ImportDeclaration {
	current := node
	for current != nil {
		switch current.Kind {
		case ast.KindImportDeclaration, ast.KindJSImportDeclaration:
			return current.AsImportDeclaration()
		}
		current = current.Parent
	}
	return nil
}

func applyParsedExpectCall(parsed *ParsedJestFnCall) bool {
	modifierEntries, matcher, err := FindExpectModifiersAndMatcher(parsed.MemberEntries)
	if err != "" {
		return false
	}

	parsed.ModifierEntries = modifierEntries
	parsed.Matcher = matcher.Name
	parsed.MatcherEntry = matcher
	if len(modifierEntries) > 0 {
		parsed.Modifiers = make([]string, len(modifierEntries))
		for i, entry := range modifierEntries {
			parsed.Modifiers[i] = entry.Name
		}
	}
	return true
}

// isInnerExpectCall detects expect() calls embedded in a matcher chain
// (e.g. expect(1) in expect(1).toBe(2)) before running type-checker resolution.
func isInnerExpectCall(node *ast.Node, localName string, members []string, settings map[string]interface{}) bool {
	if len(members) > 0 || !IsMemberAccessNode(node.Parent) {
		return false
	}
	if FindTopMostCallExpression(node) == node {
		return false
	}
	name := ApplyGlobalJestAlias(localName, settings)
	return GetJestKind(name) == JestFnTypeExpect
}

func FindExpectModifiersAndMatcher(entries []ParsedJestFnMemberEntry) (
	[]ParsedJestFnMemberEntry,
	*ParsedJestFnMemberEntry,
	string,
) {
	if len(entries) == 0 {
		return nil, nil, ExpectParseReasonMatcherNotFound
	}

	modifiers := make([]ParsedJestFnMemberEntry, 0, len(entries))
	for _, member := range entries {
		parent := member.Node.Parent
		if parent == nil {
			return nil, nil, ExpectParseReasonModifierUnknown
		}

		grandparent := parent.Parent
		if grandparent != nil && grandparent.Kind == ast.KindCallExpression {
			return modifiers, &member, ExpectParseReasonNone
		}

		switch len(modifiers) {
		case 0:
			if !EXPECT_MODIFIER_NAMES[member.Name] {
				return nil, nil, ExpectParseReasonModifierUnknown
			}
		case 1:
			if member.Name != "not" {
				return nil, nil, ExpectParseReasonModifierUnknown
			}
			first := modifiers[0].Name
			if first != "rejects" && first != "resolves" {
				return nil, nil, ExpectParseReasonModifierUnknown
			}
		default:
			return nil, nil, ExpectParseReasonModifierUnknown
		}

		modifiers = append(modifiers, member)
	}

	return nil, nil, ExpectParseReasonMatcherNotFound
}

func ResolveJestFunctionReference(node *ast.Node, localName string, localNode *ast.Node, ctx rule.RuleContext) (string, *ast.Node, JestImportMode) {
	if ctx.TypeChecker == nil {
		return localName, localNode, JEST_GLOBAL_MODE
	}

	typeChecker := ctx.TypeChecker
	callExpr := node.AsCallExpression()
	if callExpr == nil {
		return localName, localNode, JEST_GLOBAL_MODE
	}

	ident := ResolveFirstIdentifier(callExpr.Expression)
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return localName, localNode, JEST_GLOBAL_MODE
	}

	symbol := typeChecker.GetSymbolAtLocation(ident)
	if symbol == nil {
		return localName, localNode, JEST_GLOBAL_MODE
	}

	hasLocalNonJestDeclaration := false
	for _, decl := range symbol.Declarations {
		if decl == nil {
			continue
		}

		if name, originalNode, ok := resolveJestGlobalsImportSpecifier(decl); ok {
			return name, originalNode, JEST_IMPORT_MODE
		}

		if name, originalNode, ok := resolveJestGlobalsRequireBinding(decl); ok {
			return name, originalNode, JEST_IMPORT_MODE
		}

		if ctx.SourceFile != nil && ast.GetSourceFileOfNode(decl) == ctx.SourceFile {
			hasLocalNonJestDeclaration = true
		}
	}

	if hasLocalNonJestDeclaration {
		return "", nil, JEST_GLOBAL_MODE
	}

	return localName, localNode, JEST_GLOBAL_MODE
}

func resolveJestGlobalsImportSpecifier(decl *ast.Node) (string, *ast.Node, bool) {
	if decl == nil || decl.Kind != ast.KindImportSpecifier {
		return "", nil, false
	}

	importDecl := FindImportDeclaration(decl)
	if importDecl == nil || importDecl.ModuleSpecifier == nil || importDecl.ModuleSpecifier.Text() != jestGlobalsModule {
		return "", nil, false
	}

	spec := decl.AsImportSpecifier()
	if spec == nil || spec.IsTypeOnly {
		return "", nil, false
	}

	if spec.PropertyName != nil {
		return spec.PropertyName.Text(), spec.PropertyName, true
	}

	name := spec.Name()
	if name == nil {
		return "", nil, false
	}

	return name.Text(), name, true
}

func resolveJestGlobalsRequireBinding(decl *ast.Node) (string, *ast.Node, bool) {
	if decl == nil || decl.Kind != ast.KindBindingElement {
		return "", nil, false
	}

	varDecl := internalUtils.EnclosingVariableDeclarationOfBindingElement(decl)
	if varDecl == nil || !isJestGlobalsRequireCall(varDecl.AsVariableDeclaration().Initializer) {
		return "", nil, false
	}

	binding := decl.AsBindingElement()
	if binding == nil {
		return "", nil, false
	}

	nameNode := binding.Name()
	if binding.PropertyName != nil {
		if name := getPropertyName(binding.PropertyName); name != "" {
			return name, binding.PropertyName, true
		}
	}

	if nameNode != nil {
		if name := getPropertyName(nameNode); name != "" {
			return name, nameNode, true
		}
	}

	return "", nil, false
}

func isJestGlobalsRequireCall(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil || !ast.IsRequireCall(node, true /*requireStringLiteralLikeArgument*/) {
		return false
	}

	args := node.Arguments()
	if len(args) == 0 || args[0] == nil {
		return false
	}

	specifier := ast.SkipParentheses(args[0])
	if specifier == nil {
		return false
	}

	switch specifier.Kind {
	case ast.KindStringLiteral:
		return specifier.AsStringLiteral().Text == jestGlobalsModule
	case ast.KindNoSubstitutionTemplateLiteral:
		return specifier.AsNoSubstitutionTemplateLiteral().Text == jestGlobalsModule
	default:
		return false
	}
}

func resolveHeadLocalNode(callExpr *ast.CallExpression) *ast.Node {
	if callExpr == nil {
		return nil
	}
	return ResolveFirstIdentifier(callExpr.Expression)
}

// ResolveFirstIdentifier walks the left side of a call/member chain and returns
// the first identifier it finds, if any.
func ResolveFirstIdentifier(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	node = ast.SkipParentheses(node)
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node
	case ast.KindCallExpression:
		return ResolveFirstIdentifier(node.AsCallExpression().Expression)
	case ast.KindPropertyAccessExpression:
		return ResolveFirstIdentifier(node.AsPropertyAccessExpression().Expression)
	case ast.KindElementAccessExpression:
		return ResolveFirstIdentifier(node.AsElementAccessExpression().Expression)
	case ast.KindTaggedTemplateExpression:
		return ResolveFirstIdentifier(node.AsTaggedTemplateExpression().Tag)
	}

	return nil
}

func isEachFactoryCall(callExpr *ast.CallExpression, members []string) bool {
	if callExpr == nil || len(members) == 0 || members[len(members)-1] != "each" {
		return false
	}

	// Only skip the factory layer so members (e.g. each/only/skip) are preserved on the actual call.
	// .each has a "factory call + actual call" shape:
	// - factory:  describe.each(...)
	// - actual:
	// 	- CallExpression: describe.each(...)(...)
	// 	- TaggedTemplateExpression: describe[`each`](...)(...)
	//  - PropertyAccessExpression: describe["each"](...)(...)
	expr := ast.SkipParentheses(callExpr.Expression)
	if expr == nil {
		return true
	}
	switch expr.Kind {
	case ast.KindCallExpression, ast.KindTaggedTemplateExpression:
		return false
	default:
		return true
	}
}

func isInvalidTaggedTemplateCall(callExpr *ast.CallExpression, members []string) bool {
	if callExpr == nil || callExpr.Expression == nil {
		return false
	}
	expr := ast.SkipParentheses(callExpr.Expression)
	if expr == nil || expr.Kind != ast.KindTaggedTemplateExpression {
		return false
	}

	return len(members) == 0 || members[len(members)-1] != "each"
}

func isValidJestCall(name string, members []string) bool {
	chain := name
	if len(members) > 0 {
		chain += "." + strings.Join(members, ".")
	}

	_, ok := VALID_JEST_FN_CALL_CHAINS[chain]
	return ok
}

func UnwrapBasicTypeAssertions(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		default:
			return node
		}
	}
	return node
}

func UnwrapTypeAssertions(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		default:
			return node
		}
	}
	return node
}

func GetAccessorReceiverAndParent(entry *ParsedJestFnMemberEntry) (*ast.Node, *ast.Node) {
	if entry == nil || entry.Node == nil || entry.Node.Parent == nil {
		return nil, nil
	}

	parent := entry.Node.Parent
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Expression, parent
	case ast.KindElementAccessExpression:
		return parent.AsElementAccessExpression().Expression, parent
	default:
		return nil, nil
	}
}

func IsNamedMember(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text == name
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text == name
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text == name
	case ast.KindPrivateIdentifier:
		return node.AsPrivateIdentifier().Text == name
	default:
		return false
	}
}

// ReceiverBeforeInvocation returns the expression before .m() or ["m"]() on a
// call, such as `expect(x).not` before `.toBe()`.
func ReceiverBeforeInvocation(matcherCall *ast.Node) *ast.Node {
	if matcherCall == nil || matcherCall.Kind != ast.KindCallExpression {
		return nil
	}

	expr := matcherCall.AsCallExpression().Expression
	switch expr.Kind {
	case ast.KindPropertyAccessExpression:
		return expr.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		return expr.AsElementAccessExpression().Expression
	default:
		return nil
	}
}

// TestCallbackInfo describes a Jest test callback passed by reference (e.g. it('foo', getValue)).
type TestCallbackInfo struct {
	FunctionNode *ast.Node
	Name         string
}

// ResolveTestCallbackFunction resolves the callback function node for a Jest test call.
// Inline callbacks are not returned; they are tracked via the enclosing test call.
func ResolveTestCallbackFunction(ctx rule.RuleContext, callExpr *ast.CallExpression) TestCallbackInfo {
	if callExpr == nil || callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) < 2 {
		return TestCallbackInfo{}
	}

	callback := ast.SkipParentheses(callExpr.Arguments.Nodes[1])
	if callback == nil || ast.IsFunctionExpressionOrArrowFunction(callback) {
		return TestCallbackInfo{}
	}
	if callback.Kind != ast.KindIdentifier {
		return TestCallbackInfo{}
	}

	name := callback.AsIdentifier().Text
	decl := internalUtils.GetDeclaration(ctx.TypeChecker, callback)
	if decl == nil {
		return TestCallbackInfo{Name: name}
	}

	switch decl.Kind {
	case ast.KindFunctionDeclaration:
		fn := decl.AsFunctionDeclaration()
		if fn == nil {
			return TestCallbackInfo{Name: name}
		}
		return TestCallbackInfo{FunctionNode: fn.AsNode(), Name: name}
	case ast.KindVariableDeclaration:
		vd := decl.AsVariableDeclaration()
		if vd == nil {
			return TestCallbackInfo{Name: name}
		}
		if fn := testCallbackInitializerFunction(vd.Initializer); fn != nil {
			return TestCallbackInfo{FunctionNode: fn, Name: name}
		}
		return TestCallbackInfo{Name: name}
	default:
		return TestCallbackInfo{Name: name}
	}
}

func testCallbackInitializerFunction(initializer *ast.Node) *ast.Node {
	if initializer == nil {
		return nil
	}
	init := ast.SkipParentheses(initializer)
	if ast.IsFunctionExpressionOrArrowFunction(init) {
		return init
	}
	return nil
}

// ResolveNamedFunctionCallback returns the function declaration node and name when
// a Jest test call uses a named function reference as its callback (e.g. it('foo', getValue)).
func ResolveNamedFunctionCallback(ctx rule.RuleContext, callExpr *ast.CallExpression) (*ast.Node, string) {
	info := ResolveTestCallbackFunction(ctx, callExpr)
	if info.FunctionNode != nil && info.FunctionNode.Kind == ast.KindFunctionDeclaration {
		return info.FunctionNode, info.Name
	}
	return nil, info.Name
}
