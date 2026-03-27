package utils

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ParsedJestFnCall struct {
	Name      string
	LocalName string
	Kind      JestFnType
	Members   []string
	Modifiers []string
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

	chain := GetMembersChain(node)
	if len(chain) == 0 {
		return nil
	}

	members := append([]string(nil), chain[1:]...)
	callExpr := node.AsCallExpression()
	if isEachFactoryCall(callExpr, members) {
		return nil
	}
	if isInvalidTaggedTemplateCall(callExpr, members) {
		return nil
	}

	localName := chain[0]
	name := resolveOriginalName(node, localName, ctx)
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
	}

	if kind == JestFnTypeExpect {
		parsed.Modifiers = pickExpectModifiers(members)
	}

	return parsed
}

func resolveOriginalName(node *ast.Node, localName string, ctx rule.RuleContext) string {
	if ctx.TypeChecker == nil {
		return localName
	}

	typeChecker := ctx.TypeChecker
	callExpr := node.AsCallExpression()
	if callExpr == nil {
		return localName
	}

	ident := resolveFirstIdentifier(callExpr.Expression)
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return localName
	}

	symbol := typeChecker.GetSymbolAtLocation(ident)
	if symbol == nil {
		return localName
	}

	hasNonJestImportSpecifier := false
	for _, decl := range symbol.Declarations {
		if decl == nil || decl.Kind != ast.KindImportSpecifier {
			continue
		}

		importDecl := findImportDeclaration(decl)
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
			return spec.PropertyName.Text()
		}

		name := spec.Name()
		if name != nil {
			return name.Text()
		}
	}

	if hasNonJestImportSpecifier {
		return ""
	}

	return localName
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

func findImportDeclaration(node *ast.Node) *ast.ImportDeclaration {
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

func pickExpectModifiers(members []string) []string {
	if len(members) == 0 {
		return nil
	}

	modifiers := make([]string, 0, len(members))
	for _, member := range members {
		if !EXPECT_MODIFIER_NAMES[member] {
			break
		}
		modifiers = append(modifiers, member)
	}

	if len(modifiers) == 0 {
		return nil
	}

	return modifiers
}
