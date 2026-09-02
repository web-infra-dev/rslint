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
	rstestAPIHook
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
	// Semantic conclusions, applied on both the call-site chain and the chain
	// consumed while following same-file const aliases. Members deliberately
	// carries no alias-internal information, so anything that has to hold
	// across an alias belongs here instead.
	parameterizedKind RstestParameterizedKind
	executionMode     RstestExecutionMode
	skipped           bool
	todo              bool
	focusEntries      []ParsedRstestFnMemberEntry
}

type rstestAPIProfile uint8

const (
	rstestProfileCore rstestAPIProfile = iota
	rstestProfilePlaywright
)

func parseRstestFnCall(
	node *ast.Node,
	ctx rule.RuleContext,
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
		)
	} else {
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

	switch resolved.state {
	case rstestAPITest, rstestAPITestWithExtend, rstestAPIParameterizedTest:
		resolved.kind = RstestFnTypeTest
	case rstestAPIDescribe, rstestAPIParameterizedDescribe:
		resolved.kind = RstestFnTypeDescribe
	case rstestAPIHook:
		resolved.kind = RstestFnTypeHook
	default:
		return nil
	}

	// Members and MemberEntries are the syntactic view: both cover exactly the
	// members written at the call site, in order, so Members[i] equals
	// MemberEntries[i].Name. Members resolved through an alias are deliberately
	// absent from both — they have no node in this expression, and a rule
	// cannot tell them apart from call-site members. Conclusions that must hold
	// across an alias are exposed as semantic fields instead.
	members := make([]string, 0, len(parts)-consumed)
	memberEntries := make([]ParsedRstestFnMemberEntry, 0, len(parts)-consumed)
	for _, part := range parts[consumed:] {
		members = append(members, part.name)
		memberEntries = append(memberEntries, ParsedRstestFnMemberEntry{
			Name: part.name,
			Node: part.node,
		})
	}

	localName := "import.meta.rstest"
	if root.Kind == ast.KindIdentifier {
		localName = root.AsIdentifier().Text
	}
	parsed := &ParsedRstestFnCall{
		ParsedCall: testFramework.ParsedCall{
			Name:          resolved.name,
			LocalName:     localName,
			Kind:          resolved.kind,
			Members:       members,
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
		ParameterizedKind: resolved.parameterizedKind,
		ExecutionMode:     resolved.executionMode,
		Skipped:           resolved.skipped,
		Todo:              resolved.todo,
	}
	if len(resolved.focusEntries) > 0 {
		parsed.focus = &rstestFocus{entries: resolved.focusEntries}
	}
	return parsed
}

func parseRstestChain(node *ast.Node) (*ast.Node, []rstestChainPart, bool, bool) {
	if node == nil {
		return nil, nil, false, false
	}
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
	case ast.KindBinaryExpression:
		if !internalUtils.IsCommaOperator(node) {
			return nil, nil, false, false
		}
		return parseRstestChain(node.AsBinaryExpression().Right)
	default:
		return nil, nil, false, false
	}
}

func parseImportMetaRstestChain(node *ast.Node) (*ast.Node, []rstestChainPart, bool, bool) {
	if node == nil {
		return nil, nil, false, false
	}
	node = ast.SkipParentheses(node)
	if node == nil {
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
	case ast.KindBinaryExpression:
		if !internalUtils.IsCommaOperator(node) {
			return nil, nil, false, false
		}
		return parseImportMetaRstestChain(node.AsBinaryExpression().Right)
	default:
		return nil, nil, false, false
	}
}

// resolveRstestRootSymbol returns the binding a registration chain's root
// identifier refers to.
//
// Every discrimination this parser promises — an import from `@rstest/core`
// against a same-named import from Vitest or Jest, a global against a local
// declaration that shadows it, an alias against an unrelated variable — is
// decided from this symbol's declarations. Without one,
// [testFramework.ResolveFunctionIdentifierReferenceFromSymbol] can only fall
// back to the identifier's spelling, which is wrong in both directions: a
// `test` imported from Vitest is taken for an Rstest global, while a renamed
// or namespace import from `@rstest/core` stops resolving at all.
//
// The TypeChecker is therefore not the only source consulted. A source-only
// Program — the generation rslint builds for files no tsconfig project owns —
// carries no checker, so the file's reference store answers instead. Its
// binder scope walk resolves imports and declarations authored in this file,
// which is everything the discriminations above need; only bindings declared
// outside the file are beyond it, and those were never Rstest registrations.
func resolveRstestRootSymbol(ctx rule.RuleContext, root *ast.Node) *ast.Symbol {
	if ctx.TypeChecker != nil {
		if symbol := ctx.TypeChecker.GetSymbolAtLocation(root); symbol != nil {
			return symbol
		}
	}
	return ctx.Refs.ResolveInFile(root)
}

func resolveRstestRoot(
	root *ast.Node,
	parts []rstestChainPart,
	ctx rule.RuleContext,
	visited map[*ast.Symbol]bool,
	depth int,
) (rstestResolvedAPI, int, bool) {
	if root == nil || root.Kind != ast.KindIdentifier || depth > 16 {
		return rstestResolvedAPI{}, 0, false
	}

	localName := root.AsIdentifier().Text
	symbol := resolveRstestRootSymbol(ctx, root)
	for _, profile := range rstestAllProfiles {
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

	for _, profile := range rstestAllProfiles {
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
		if !ok && isImportMetaRstest(initializer) {
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
	states, ok := rstestDirectAPIStates[name]
	if !ok {
		return rstestAPIInvalid
	}
	return states[profile]
}

var rstestDirectAPIStates = func() map[string][2]rstestAPIState {
	states := map[string][2]rstestAPIState{
		"test": {
			rstestProfileCore:       rstestAPITestWithExtend,
			rstestProfilePlaywright: rstestAPITestWithExtend,
		},
		"it": {
			rstestProfileCore:       rstestAPITestWithExtend,
			rstestProfilePlaywright: rstestAPIInvalid,
		},
		"describe": {
			rstestProfileCore:       rstestAPIDescribe,
			rstestProfilePlaywright: rstestAPIDescribe,
		},
	}
	// Both profiles expose the same four hooks: @rstest/playwright re-exports
	// them (packages/playwright/src/index.ts:2-9). Hooks accept no chained
	// members, which holds by construction: applyRstestChainPart has no
	// rstestAPIHook branch, so any member invalidates the chain.
	for _, name := range testFramework.HooksOrder {
		states[name] = [2]rstestAPIState{
			rstestProfileCore:       rstestAPIHook,
			rstestProfilePlaywright: rstestAPIHook,
		}
	}
	return states
}()

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
		// PlaywrightTest declares the four hooks as members of the test object,
		// and extend() returns another PlaywrightTest, so `test.beforeEach()`
		// and `test.extend(fixtures).beforeEach()` register a hook. This is a
		// member lookup on the test object, not a modifier applied to an
		// already resolved hook, which is why it is reached from the test state
		// rather than from rstestAPIHook.
		//
		// Only rstestAPITestWithExtend qualifies: that state means the receiver
		// is still a PlaywrightTest. Modifiers such as `.skip` or `.each(...)`
		// leave rstestAPITest, and those results do not expose hooks.
		if profile == rstestProfilePlaywright &&
			state == rstestAPITestWithExtend &&
			part.invocation == rstestNotInvoked &&
			testFramework.IsHookName(part.name) {
			return rstestAPIHook
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
	// `test.beforeEach()` resolves to the hook the member selects, not to the
	// test object it was read from, so the reported name and original node move
	// to the member. Rules keyed on the hook name — prefer-hooks-in-order, for
	// one — would otherwise see every Playwright hook as `test`.
	//
	// No guard against a hook→hook transition is needed: applyRstestChainPart
	// has no rstestAPIHook branch, so a member applied to a hook always
	// invalidates the chain and returns above. Reaching here with a hook state
	// therefore means the member just introduced it.
	if state == rstestAPIHook {
		resolved.name = part.name
		resolved.originalNode = part.node
	}
	resolved.state = state
	// Only semantic conclusions are recorded here. This runs on the alias chain
	// as well as the call-site chain, so anything accumulated becomes visible
	// across aliases — which is why the member names themselves are collected
	// by the caller from parts[consumed:] instead.
	switch part.name {
	case "concurrent":
		// Runtime option merging gives concurrent priority when both modifiers
		// are present, regardless of their source order.
		resolved.executionMode = RstestExecutionConcurrent
	case "each":
		resolved.parameterizedKind = RstestParameterizedEach
	case "for":
		resolved.parameterizedKind = RstestParameterizedFor
	case "only":
		resolved.focusEntries = append(resolved.focusEntries, ParsedRstestFnMemberEntry{
			Name: part.name,
			Node: part.node,
		})
	case "sequential":
		if resolved.executionMode != RstestExecutionConcurrent {
			resolved.executionMode = RstestExecutionSequential
		}
	case "skip":
		resolved.skipped = true
	case "todo":
		resolved.todo = true
	}
	return true
}

func profileImportModule(profile rstestAPIProfile) string {
	if profile == rstestProfilePlaywright {
		return RstestPlaywrightImportModule
	}
	return RstestImportModule
}

var rstestAllProfiles = [...]rstestAPIProfile{rstestProfileCore, rstestProfilePlaywright}

func isImportMetaRstest(node *ast.Node) bool {
	// SkipParentheses dereferences its argument, so the nil check has to come
	// first: callers pass initializers, which are nil for a bare `let x;`.
	if node == nil {
		return false
	}
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
	if node == nil {
		return false
	}
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
