package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type rstestAPIState uint8

const (
	rstestAPIInvalid rstestAPIState = iota
	rstestAPITest
	rstestAPITestWithExtend
	rstestAPIDescribe
	rstestAPIParameterizedTest
	rstestAPIParameterizedDescribe
)

type rstestInvocation uint8

const (
	rstestNotInvoked rstestInvocation = iota
	rstestCallInvoked
	rstestTagInvoked
)

type rstestChainPart struct {
	name       string
	node       *ast.Node
	invocation rstestInvocation
}

type rstestResolvedAPI struct {
	name         string
	kind         RstestFnType
	state        rstestAPIState
	originalNode *ast.Node
	mode         RstestImportMode
	profile      rstestAPIProfile
	// members collects every chain member applied to the resolved API,
	// including members consumed while following same-file const aliases.
	members           []string
	parameterizedKind RstestParameterizedKind
}

type rstestAPIProfile uint8

const (
	rstestProfileCore rstestAPIProfile = iota
	rstestProfilePlaywright
)

// ParseRstestFnCall parses a final Rstest test/describe registration call.
// Factory calls such as test.each(cases), test.runIf(condition), and
// test.extend(fixtures) are intentionally not returned.
func ParseRstestFnCall(node *ast.Node, ctx rule.RuleContext) *ParsedRstestFnCall {
	return parseRstestFnCall(node, ctx, false, false)
}

func ParseRstestFnCallWithOfficialExtensions(
	node *ast.Node,
	ctx rule.RuleContext,
) *ParsedRstestFnCall {
	return parseRstestFnCall(node, ctx, true, true)
}

func parseRstestFnCall(
	node *ast.Node,
	ctx rule.RuleContext,
	includeImportMeta bool,
	includePlaywright bool,
) *ParsedRstestFnCall {
	if node == nil || node.Kind != ast.KindCallExpression {
		return nil
	}

	call := node.AsCallExpression()
	root, parts, rootInvoked, ok := parseRstestChain(call.Expression)
	var resolved rstestResolvedAPI
	consumed := 0
	if ok && !rootInvoked && root != nil {
		resolved, consumed, ok = resolveRstestRoot(
			root,
			parts,
			ctx,
			nil,
			0,
			includeImportMeta,
			includePlaywright,
		)
	} else if includeImportMeta {
		var importMetaRoot *ast.Node
		importMetaRoot, parts, rootInvoked, ok = parseImportMetaRstestChain(call.Expression)
		if ok && !rootInvoked && importMetaRoot != nil && len(parts) > 0 {
			state := directRstestAPIState(rstestProfileCore, parts[0].name)
			if state != rstestAPIInvalid && parts[0].invocation == rstestNotInvoked {
				resolved = rstestResolvedAPI{
					name:         parts[0].name,
					state:        state,
					originalNode: parts[0].node,
					mode:         RSTEST_IMPORT_MODE,
					profile:      rstestProfileCore,
				}
				root = importMetaRoot
				consumed = 1
			} else {
				ok = false
			}
		}
	}
	if !ok || root == nil {
		return nil
	}

	for _, part := range parts[consumed:] {
		if !applyResolvedRstestChainPart(&resolved, part) {
			return nil
		}
	}

	parameterized := false
	switch resolved.state {
	case rstestAPITest, rstestAPITestWithExtend:
		resolved.kind = RstestFnTypeTest
	case rstestAPIDescribe:
		resolved.kind = RstestFnTypeDescribe
	case rstestAPIParameterizedTest:
		resolved.kind = RstestFnTypeTest
		parameterized = true
	case rstestAPIParameterizedDescribe:
		resolved.kind = RstestFnTypeDescribe
		parameterized = true
	default:
		return nil
	}

	// MemberEntries only covers members written at the call site, because
	// members resolved through an alias have no node in this expression.
	memberEntries := make([]ParsedRstestFnMemberEntry, 0, len(parts)-consumed)
	for _, part := range parts[consumed:] {
		memberEntries = append(memberEntries, ParsedRstestFnMemberEntry{
			Name: part.name,
			Node: part.node,
		})
	}

	localName := "import.meta.rstest"
	if root.Kind == ast.KindIdentifier {
		localName = root.AsIdentifier().Text
	}
	return &ParsedRstestFnCall{
		ParsedCall: testFramework.ParsedCall{
			Name:          resolved.name,
			LocalName:     localName,
			Kind:          resolved.kind,
			Members:       resolved.members,
			MemberEntries: memberEntries,
			Head: ParsedRstestFnCallHead{
				Type: resolved.mode,
				Local: ParsedRstestFnCallHeadEntry{
					Value: localName,
					Node:  root,
				},
				Original: ParsedRstestFnCallHeadEntry{
					Value: resolved.name,
					Node:  resolved.originalNode,
				},
			},
		},
		Parameterized:     parameterized,
		ParameterizedKind: resolved.parameterizedKind,
	}
}

func parseRstestChain(node *ast.Node) (*ast.Node, []rstestChainPart, bool, bool) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return nil, nil, false, false
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node, nil, false, true
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		root, parts, rootInvoked, ok := parseRstestChain(property.Expression)
		if !ok {
			return nil, nil, false, false
		}
		nameNode := property.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return nil, nil, false, false
		}
		return root, append(parts, rstestChainPart{
			name: nameNode.AsIdentifier().Text,
			node: nameNode,
		}), rootInvoked, true
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		root, parts, rootInvoked, ok := parseRstestChain(element.Expression)
		if !ok {
			return nil, nil, false, false
		}
		nameNode := ast.SkipParentheses(element.ArgumentExpression)
		name, ok := internalUtils.GetStaticStringLiteralValue(nameNode)
		if !ok || name == "" {
			return nil, nil, false, false
		}
		return root, append(parts, rstestChainPart{
			name: name,
			node: nameNode,
		}), rootInvoked, true
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		root, parts, rootInvoked, ok := parseRstestChain(call.Expression)
		if !ok {
			return nil, nil, false, false
		}
		if len(parts) == 0 {
			if rootInvoked {
				return nil, nil, false, false
			}
			return root, parts, true, true
		}
		if parts[len(parts)-1].invocation != rstestNotInvoked {
			return nil, nil, false, false
		}
		parts[len(parts)-1].invocation = rstestCallInvoked
		return root, parts, rootInvoked, true
	case ast.KindTaggedTemplateExpression:
		tagged := node.AsTaggedTemplateExpression()
		root, parts, rootInvoked, ok := parseRstestChain(tagged.Tag)
		if !ok || len(parts) == 0 || parts[len(parts)-1].invocation != rstestNotInvoked {
			return nil, nil, false, false
		}
		parts[len(parts)-1].invocation = rstestTagInvoked
		return root, parts, rootInvoked, true
	default:
		return nil, nil, false, false
	}
}

func parseImportMetaRstestChain(node *ast.Node) (*ast.Node, []rstestChainPart, bool, bool) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return nil, nil, false, false
	}
	if ast.IsOptionalChain(node) {
		return nil, nil, false, false
	}
	if isImportMetaRstest(node) {
		return node, nil, false, true
	}

	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		root, parts, rootInvoked, ok := parseImportMetaRstestChain(property.Expression)
		if !ok {
			return nil, nil, false, false
		}
		nameNode := property.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return nil, nil, false, false
		}
		return root, append(parts, rstestChainPart{
			name: nameNode.AsIdentifier().Text,
			node: nameNode,
		}), rootInvoked, true
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		root, parts, rootInvoked, ok := parseImportMetaRstestChain(element.Expression)
		if !ok {
			return nil, nil, false, false
		}
		nameNode := ast.SkipParentheses(element.ArgumentExpression)
		name, ok := internalUtils.GetStaticStringLiteralValue(nameNode)
		if !ok || name == "" {
			return nil, nil, false, false
		}
		return root, append(parts, rstestChainPart{
			name: name,
			node: nameNode,
		}), rootInvoked, true
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		root, parts, rootInvoked, ok := parseImportMetaRstestChain(call.Expression)
		if !ok {
			return nil, nil, false, false
		}
		if len(parts) == 0 {
			return nil, nil, false, false
		}
		if parts[len(parts)-1].invocation != rstestNotInvoked {
			return nil, nil, false, false
		}
		parts[len(parts)-1].invocation = rstestCallInvoked
		return root, parts, rootInvoked, true
	case ast.KindTaggedTemplateExpression:
		tagged := node.AsTaggedTemplateExpression()
		root, parts, rootInvoked, ok := parseImportMetaRstestChain(tagged.Tag)
		if !ok || len(parts) == 0 || parts[len(parts)-1].invocation != rstestNotInvoked {
			return nil, nil, false, false
		}
		parts[len(parts)-1].invocation = rstestTagInvoked
		return root, parts, rootInvoked, true
	default:
		return nil, nil, false, false
	}
}

func resolveRstestRoot(
	root *ast.Node,
	parts []rstestChainPart,
	ctx rule.RuleContext,
	visited map[*ast.Symbol]bool,
	depth int,
	includeImportMeta bool,
	includePlaywright bool,
) (rstestResolvedAPI, int, bool) {
	if root == nil || root.Kind != ast.KindIdentifier || depth > 16 {
		return rstestResolvedAPI{}, 0, false
	}

	localName := root.AsIdentifier().Text
	var symbol *ast.Symbol
	if ctx.TypeChecker != nil {
		symbol = ctx.TypeChecker.GetSymbolAtLocation(root)
	}
	for _, profile := range rstestProfiles(includePlaywright) {
		module := profileImportModule(profile)
		name, originalNode, mode := testFramework.ResolveFunctionIdentifierReferenceFromSymbol(
			localName,
			root,
			symbol,
			ctx.SourceFile,
			module,
		)
		if state := directRstestAPIState(profile, name); state != rstestAPIInvalid {
			return rstestResolvedAPI{
				name:         name,
				state:        state,
				originalNode: originalNode,
				mode:         mode,
				profile:      profile,
			}, 0, true
		}
	}

	if includeImportMeta {
		if name, originalNode, ok := resolveImportMetaRstestBinding(symbol); ok {
			if state := directRstestAPIState(rstestProfileCore, name); state != rstestAPIInvalid {
				return rstestResolvedAPI{
					name:         name,
					state:        state,
					originalNode: originalNode,
					mode:         RSTEST_IMPORT_MODE,
					profile:      rstestProfileCore,
				}, 0, true
			}
		}
	}

	for _, profile := range rstestProfiles(includePlaywright) {
		if testFramework.IsModuleNamespaceSymbol(symbol, profileImportModule(profile)) {
			if len(parts) == 0 || parts[0].invocation != rstestNotInvoked {
				return rstestResolvedAPI{}, 0, false
			}
			state := directRstestAPIState(profile, parts[0].name)
			if state == rstestAPIInvalid {
				return rstestResolvedAPI{}, 0, false
			}
			return rstestResolvedAPI{
				name:         parts[0].name,
				state:        state,
				originalNode: parts[0].node,
				mode:         RSTEST_IMPORT_MODE,
				profile:      profile,
			}, 1, true
		}
	}

	if symbol == nil || (visited != nil && visited[symbol]) {
		return rstestResolvedAPI{}, 0, false
	}

	tracking := false
	for _, declaration := range symbol.Declarations {
		if !isConstRstestVariableDeclaration(declaration) {
			continue
		}
		initializer := declaration.AsVariableDeclaration().Initializer
		aliasRoot, aliasParts, aliasRootInvoked, ok := parseRstestChain(initializer)
		if includeImportMeta && !ok && isImportMetaRstest(initializer) {
			if len(parts) == 0 || parts[0].invocation != rstestNotInvoked {
				continue
			}
			state := directRstestAPIState(rstestProfileCore, parts[0].name)
			if state == rstestAPIInvalid {
				continue
			}
			return rstestResolvedAPI{
				name:         parts[0].name,
				state:        state,
				originalNode: parts[0].node,
				mode:         RSTEST_IMPORT_MODE,
				profile:      rstestProfileCore,
			}, 1, true
		}
		if !ok || aliasRootInvoked {
			continue
		}
		if !tracking {
			if visited == nil {
				visited = make(map[*ast.Symbol]bool)
			}
			visited[symbol] = true
			defer delete(visited, symbol)
			tracking = true
		}
		resolved, consumed, ok := resolveRstestRoot(
			aliasRoot,
			aliasParts,
			ctx,
			visited,
			depth+1,
			includeImportMeta,
			includePlaywright,
		)
		if !ok {
			continue
		}
		valid := true
		for _, part := range aliasParts[consumed:] {
			if !applyResolvedRstestChainPart(&resolved, part) {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}
		return resolved, 0, true
	}

	return rstestResolvedAPI{}, 0, false
}

func resolveImportMetaRstestBinding(symbol *ast.Symbol) (string, *ast.Node, bool) {
	if symbol == nil {
		return "", nil, false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindBindingElement {
			continue
		}
		variable := internalUtils.EnclosingVariableDeclarationOfBindingElement(declaration)
		if variable == nil || !isImportMetaRstest(variable.AsVariableDeclaration().Initializer) {
			continue
		}
		binding := declaration.AsBindingElement()
		if binding.PropertyName != nil {
			name, ok := internalUtils.GetStaticStringLiteralValue(binding.PropertyName)
			if ok {
				return name, binding.PropertyName, true
			}
			if binding.PropertyName.Kind == ast.KindIdentifier {
				return binding.PropertyName.AsIdentifier().Text, binding.PropertyName, true
			}
		}
		name := binding.Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.AsIdentifier().Text, name, true
		}
	}
	return "", nil, false
}

func isConstRstestVariableDeclaration(declaration *ast.Node) bool {
	return declaration != nil &&
		declaration.Kind == ast.KindVariableDeclaration &&
		declaration.Parent != nil &&
		declaration.Parent.Kind == ast.KindVariableDeclarationList &&
		declaration.Parent.Flags&ast.NodeFlagsConst != 0 &&
		declaration.AsVariableDeclaration().Initializer != nil
}

func directRstestAPIState(profile rstestAPIProfile, name string) rstestAPIState {
	switch name {
	case "test", "it":
		if profile == rstestProfilePlaywright && name == "it" {
			return rstestAPIInvalid
		}
		return rstestAPITestWithExtend
	case "describe":
		return rstestAPIDescribe
	default:
		return rstestAPIInvalid
	}
}

func applyRstestChainPart(profile rstestAPIProfile, state rstestAPIState, part rstestChainPart) rstestAPIState {
	switch state {
	case rstestAPITest, rstestAPITestWithExtend:
		switch part.name {
		case "only", "skip", "todo", "fails", "concurrent", "sequential":
			if part.invocation == rstestNotInvoked {
				return rstestAPITest
			}
		case "fail":
			if profile == rstestProfilePlaywright && part.invocation == rstestNotInvoked {
				return rstestAPITest
			}
		case "runIf", "skipIf":
			if part.invocation == rstestCallInvoked {
				return rstestAPITest
			}
		case "each", "for":
			if part.invocation == rstestCallInvoked || part.invocation == rstestTagInvoked {
				return rstestAPIParameterizedTest
			}
		case "extend":
			if state == rstestAPITestWithExtend && part.invocation == rstestCallInvoked {
				return rstestAPITestWithExtend
			}
		case "describe":
			if profile == rstestProfilePlaywright && part.invocation == rstestNotInvoked {
				return rstestAPIDescribe
			}
		}
	case rstestAPIDescribe:
		switch part.name {
		case "only", "skip", "todo", "concurrent", "sequential":
			if part.invocation == rstestNotInvoked {
				return rstestAPIDescribe
			}
		case "runIf", "skipIf":
			if part.invocation == rstestCallInvoked {
				return rstestAPIDescribe
			}
		case "each", "for":
			if part.invocation == rstestCallInvoked || part.invocation == rstestTagInvoked {
				return rstestAPIParameterizedDescribe
			}
		}
	}
	return rstestAPIInvalid
}

func applyResolvedRstestChainPart(resolved *rstestResolvedAPI, part rstestChainPart) bool {
	state := applyRstestChainPart(resolved.profile, resolved.state, part)
	if state == rstestAPIInvalid {
		resolved.state = rstestAPIInvalid
		return false
	}
	resolved.state = state
	resolved.members = append(resolved.members, part.name)
	switch part.name {
	case "each":
		resolved.parameterizedKind = RstestParameterizedEach
	case "for":
		resolved.parameterizedKind = RstestParameterizedFor
	}
	return true
}

func profileImportModule(profile rstestAPIProfile) string {
	if profile == rstestProfilePlaywright {
		return RstestPlaywrightImportModule
	}
	return RstestImportModule
}

func rstestProfiles(includePlaywright bool) []rstestAPIProfile {
	if includePlaywright {
		return []rstestAPIProfile{rstestProfileCore, rstestProfilePlaywright}
	}
	return []rstestAPIProfile{rstestProfileCore}
}

func isImportMetaRstest(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		return property.Name() != nil &&
			property.Name().Text() == "rstest" &&
			isImportMeta(property.Expression)
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		name, ok := internalUtils.GetStaticStringLiteralValue(ast.SkipParentheses(element.ArgumentExpression))
		return ok && name == "rstest" && isImportMeta(element.Expression)
	default:
		return false
	}
}

func isImportMeta(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil || node.Kind != ast.KindMetaProperty {
		return false
	}
	meta := node.AsMetaProperty()
	return meta != nil &&
		scanner.TokenToString(meta.KeywordToken) == "import" &&
		meta.Name() != nil &&
		meta.Name().Text() == "meta"
}
