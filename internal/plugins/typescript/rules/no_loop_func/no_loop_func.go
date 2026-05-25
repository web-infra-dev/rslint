package no_loop_func

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	core "github.com/web-infra-dev/rslint/internal/rules/no_loop_func"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// isSafe mirrors the upstream @typescript-eslint/no-loop-func `isSafe`.
// Differs from the rslint core variant in two ways:
//   - Only `const` is treated as a constant binding. `using` / `await using`
//     are NOT (matches the ts-eslint plugin, which forks an older ESLint
//     core that didn't yet recognize them).
//   - Type-only references are filtered out earlier (in CollectThroughReferences),
//     mirroring upstream's `reference.isTypeReference` short-circuit.
func isSafe(s *core.RunState, loopNode *ast.Node, ref core.ReferenceEntry) bool {
	sym := ref.Symbol()
	if sym == nil || len(sym.Declarations) == 0 {
		return true
	}
	decl := sym.Declarations[0]

	declList := core.GetDeclListForSymbolDecl(decl)
	kind := core.GetVarDeclListKind(declList)

	if kind == "const" {
		return true
	}

	sf := s.Ctx.SourceFile
	loopRange := utils.TrimNodeTextRange(sf, loopNode)

	if kind == "let" && declList != nil {
		declRange := utils.TrimNodeTextRange(sf, declList)
		if declRange.Pos() > loopRange.Pos() && declRange.End() < loopRange.End() {
			return true
		}
	}

	var excluded *ast.Node
	if kind == "let" {
		excluded = declList
	}
	top := s.GetTopLoopNode(loopNode, excluded)
	border := utils.TrimNodeTextRange(sf, top).Pos()

	varScope := utils.FindEnclosingScope(decl)

	safe := true
	s.ForEachReference(sym, func(refNode *ast.Node) bool {
		if !core.IsWriteRef(refNode) {
			return false
		}
		refScope := utils.FindEnclosingScope(refNode)
		if refScope == varScope {
			refPos := utils.TrimNodeTextRange(sf, refNode).Pos()
			if refPos < border {
				return false
			}
		}
		safe = false
		return true
	})
	return safe
}

// checkForLoops mirrors upstream's `checkForLoops`. Unlike the rslint core
// variant, unsafe variable names are NOT deduplicated — repeated occurrences
// of the same name appear repeatedly in the message, matching ts-eslint.
func checkForLoops(s *core.RunState, node *ast.Node) {
	loopNode := s.GetContainingLoopNode(node)
	if loopNode == nil {
		return
	}

	refs := s.CollectThroughReferences(node)

	if !core.IsAsyncOrGenerator(node) && core.IsIIFE(node) {
		isFunctionExpression := node.Kind == ast.KindFunctionExpression
		name := node.Name()
		isFunctionReferenced := false
		if isFunctionExpression && name != nil {
			refName := name.Text()
			for _, r := range refs {
				if r.Name() == refName {
					isFunctionReferenced = true
					break
				}
			}
		}
		if !isFunctionReferenced {
			s.SkippedIIFEs[node] = true
			return
		}
	}

	var names []string
	for _, r := range refs {
		if r.Symbol() == nil {
			continue
		}
		if isSafe(s, loopNode, r) {
			continue
		}
		names = append(names, r.Name())
	}

	if len(names) == 0 {
		return
	}

	varNames := "'" + strings.Join(names, "', '") + "'"
	s.Ctx.ReportNode(node, core.BuildUnsafeRefsMessage(varNames))
}

// NoLoopFuncRule is @typescript-eslint/no-loop-func. It diverges from the
// rslint core no-loop-func in two intentional ways to stay 1:1 with the
// upstream ts-eslint plugin (which forks an older ESLint core):
//   - `using` / `await using` are not treated as constant bindings.
//   - Repeated unsafe variable names are not deduplicated.
var NoLoopFuncRule = rule.CreateRule(rule.Rule{
	Name:             "no-loop-func",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		if ctx.TypeChecker == nil {
			return rule.RuleListeners{}
		}
		s := core.NewRunState(ctx)
		check := func(node *ast.Node) {
			checkForLoops(s, node)
		}
		return rule.RuleListeners{
			ast.KindArrowFunction:       check,
			ast.KindFunctionExpression:  check,
			ast.KindFunctionDeclaration: check,
			// tsgo splits class/object methods, getters, setters and
			// constructors out of FunctionExpression. ESLint's ESTree
			// flattens them under FunctionExpression, so we register
			// the extra kinds explicitly to match upstream's coverage.
			ast.KindMethodDeclaration: check,
			ast.KindGetAccessor:       check,
			ast.KindSetAccessor:       check,
			ast.KindConstructor:       check,
		}
	},
})
