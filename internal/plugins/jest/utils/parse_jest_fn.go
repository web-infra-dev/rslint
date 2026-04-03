package utils

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ParsedJestFnCall struct {
	Name            string
	LocalName       string
	Kind            JestFnType
	Members         []string
	MemberEntries   []ParsedJestFnMemberEntry
	Modifiers       []string
	ModifierEntries []ParsedJestFnMemberEntry
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
	if isEachFactoryCall(callExpr, members) {
		return nil
	}
	if isInvalidTaggedTemplateCall(callExpr, members) {
		return nil
	}

	localNode := resolveHeadLocalNode(callExpr)
	name, originalNode, headType := resolveOriginalName(node, localName, localNode, ctx)
	if name == "" {
		return nil
	}
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
		Name:      name,
		LocalName: localName,
		Kind:      kind,
		Members:   members,
		MemberEntries: append(
			[]ParsedJestFnMemberEntry(nil),
			memberEntries[1:]...,
		),
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
		modifiers, modifierEntries := pickExpectModifiersAndEntries(parsed.MemberEntries)
		parsed.Modifiers = modifiers
		parsed.ModifierEntries = modifierEntries
	}

	return parsed
}

func resolveOriginalName(node *ast.Node, localName string, localNode *ast.Node, ctx rule.RuleContext) (string, *ast.Node, JestImportMode) {
	if ctx.TypeChecker == nil {
		return localName, localNode, JEST_GLOBAL_MODE
	}

	typeChecker := ctx.TypeChecker
	callExpr := node.AsCallExpression()
	if callExpr == nil {
		return localName, localNode, JEST_GLOBAL_MODE
	}

	ident := resolveFirstIdentifier(callExpr.Expression)
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return localName, localNode, JEST_GLOBAL_MODE
	}

	symbol := typeChecker.GetSymbolAtLocation(ident)
	if symbol == nil {
		return localName, localNode, JEST_GLOBAL_MODE
	}

	hasNonJestImportSpecifier := false
	for _, decl := range symbol.Declarations {
		if decl == nil || decl.Kind != ast.KindImportSpecifier {
			continue
		}

		importDecl := FindImportDeclaration(decl)
		if importDecl == nil || importDecl.ModuleSpecifier == nil {
			continue
		}
		if importDecl.ModuleSpecifier.Text() != "@jest/globals" {
			hasNonJestImportSpecifier = true
			continue
		}

		spec := decl.AsImportSpecifier()
		if spec == nil || spec.IsTypeOnly {
			continue
		}

		if spec.PropertyName != nil {
			return spec.PropertyName.Text(), spec.PropertyName, JEST_IMPORT_MODE
		}

		name := spec.Name()
		if name != nil {
			return name.Text(), name, JEST_IMPORT_MODE
		}
	}

	if hasNonJestImportSpecifier {
		return "", nil, JEST_GLOBAL_MODE
	}

	return localName, localNode, JEST_GLOBAL_MODE
}

func resolveHeadLocalNode(callExpr *ast.CallExpression) *ast.Node {
	if callExpr == nil {
		return nil
	}
	return resolveFirstIdentifier(callExpr.Expression)
}

func resolveFirstIdentifier(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node
	case ast.KindCallExpression:
		return resolveFirstIdentifier(node.AsCallExpression().Expression)
	case ast.KindPropertyAccessExpression:
		return resolveFirstIdentifier(node.AsPropertyAccessExpression().Expression)
	case ast.KindElementAccessExpression:
		return resolveFirstIdentifier(node.AsElementAccessExpression().Expression)
	case ast.KindTaggedTemplateExpression:
		return resolveFirstIdentifier(node.AsTaggedTemplateExpression().Tag)
	}

	return nil
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
	switch callExpr.Expression.Kind {
	case ast.KindCallExpression, ast.KindTaggedTemplateExpression:
		return false
	default:
		return true
	}
}

func isInvalidTaggedTemplateCall(callExpr *ast.CallExpression, members []string) bool {
	if callExpr == nil || callExpr.Expression == nil || callExpr.Expression.Kind != ast.KindTaggedTemplateExpression {
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

func pickExpectModifiersAndEntries(entries []ParsedJestFnMemberEntry) ([]string, []ParsedJestFnMemberEntry) {
	if len(entries) == 0 {
		return nil, nil
	}

	modifierEntries := make([]ParsedJestFnMemberEntry, 0, len(entries))
	for _, entry := range entries {
		if !EXPECT_MODIFIER_NAMES[entry.Name] {
			break
		}
		modifierEntries = append(modifierEntries, entry)
	}

	if len(modifierEntries) == 0 {
		return nil, nil
	}

	modifiers := make([]string, 0, len(modifierEntries))
	for _, entry := range modifierEntries {
		modifiers = append(modifiers, entry.Name)
	}

	return modifiers, modifierEntries
}
