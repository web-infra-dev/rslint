package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
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
}

// ParseRstestFnCall parses a final Rstest test/describe registration call.
// Factory calls such as test.each(cases), test.runIf(condition), and
// test.extend(fixtures) are intentionally not returned.
func ParseRstestFnCall(node *ast.Node, ctx rule.RuleContext) *ParsedRstestFnCall {
	if node == nil || node.Kind != ast.KindCallExpression {
		return nil
	}

	call := node.AsCallExpression()
	root, parts, rootInvoked, ok := parseRstestChain(call.Expression)
	if !ok || rootInvoked || root == nil {
		return nil
	}

	resolved, consumed, ok := resolveRstestRoot(root, parts, ctx, nil, 0)
	if !ok {
		return nil
	}

	state := resolved.state
	for _, part := range parts[consumed:] {
		state = applyRstestChainPart(state, part)
		if state == rstestAPIInvalid {
			return nil
		}
	}

	parameterized := false
	switch state {
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

	memberEntries := make([]ParsedRstestFnMemberEntry, 0, len(parts)-consumed)
	members := make([]string, 0, len(parts)-consumed)
	for _, part := range parts[consumed:] {
		members = append(members, part.name)
		memberEntries = append(memberEntries, ParsedRstestFnMemberEntry{
			Name: part.name,
			Node: part.node,
		})
	}

	localName := root.AsIdentifier().Text
	return &ParsedRstestFnCall{
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
		Parameterized: parameterized,
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
	var symbol *ast.Symbol
	if ctx.TypeChecker != nil {
		symbol = ctx.TypeChecker.GetSymbolAtLocation(root)
	}
	name, originalNode, mode := testFramework.ResolveFunctionIdentifierReferenceFromSymbol(
		localName,
		root,
		symbol,
		ctx.SourceFile,
		RstestImportModule,
	)
	if state := directRstestAPIState(name); state != rstestAPIInvalid {
		return rstestResolvedAPI{
			name:         name,
			state:        state,
			originalNode: originalNode,
			mode:         mode,
		}, 0, true
	}

	if testFramework.IsModuleNamespaceSymbol(symbol, RstestImportModule) {
		if len(parts) == 0 || parts[0].invocation != rstestNotInvoked {
			return rstestResolvedAPI{}, 0, false
		}
		state := directRstestAPIState(parts[0].name)
		if state == rstestAPIInvalid {
			return rstestResolvedAPI{}, 0, false
		}
		return rstestResolvedAPI{
			name:         parts[0].name,
			state:        state,
			originalNode: parts[0].node,
			mode:         RSTEST_IMPORT_MODE,
		}, 1, true
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
		resolved, consumed, ok := resolveRstestRoot(aliasRoot, aliasParts, ctx, visited, depth+1)
		if !ok {
			continue
		}
		state := resolved.state
		for _, part := range aliasParts[consumed:] {
			state = applyRstestChainPart(state, part)
			if state == rstestAPIInvalid {
				break
			}
		}
		if state == rstestAPIInvalid {
			continue
		}
		resolved.state = state
		return resolved, 0, true
	}

	return rstestResolvedAPI{}, 0, false
}

func isConstRstestVariableDeclaration(declaration *ast.Node) bool {
	return declaration != nil &&
		declaration.Kind == ast.KindVariableDeclaration &&
		declaration.Parent != nil &&
		declaration.Parent.Kind == ast.KindVariableDeclarationList &&
		declaration.Parent.Flags&ast.NodeFlagsConst != 0 &&
		declaration.AsVariableDeclaration().Initializer != nil
}

func directRstestAPIState(name string) rstestAPIState {
	switch name {
	case "test", "it":
		return rstestAPITestWithExtend
	case "describe":
		return rstestAPIDescribe
	default:
		return rstestAPIInvalid
	}
}

func applyRstestChainPart(state rstestAPIState, part rstestChainPart) rstestAPIState {
	switch state {
	case rstestAPITest, rstestAPITestWithExtend:
		switch part.name {
		case "only", "skip", "todo", "fails", "concurrent", "sequential":
			if part.invocation == rstestNotInvoked {
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
