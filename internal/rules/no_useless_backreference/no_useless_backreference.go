package no_useless_backreference

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-useless-backreference
var NoUselessBackreferenceRule = rule.Rule{
	Name: "no-useless-backreference",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		eval := newStaticEvalCtx(ctx)
		return rule.RuleListeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				if isRegexLiteralHandledByConstructor(ctx, node) {
					return
				}
				pattern, flags := utils.ExtractRegexPatternAndFlags(node.Text())
				if pattern == "" && flags == "" {
					return
				}
				rxFlags := utils.ParseRegexFlags(flags)
				checkRegex(ctx, node, pattern, rxFlags)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				handleRegExpConstructor(ctx, node, call.Expression, call.Arguments, eval)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				handleRegExpConstructor(ctx, node, newExpr.Expression, newExpr.Arguments, eval)
			},
		}
	},
}

func handleRegExpConstructor(ctx rule.RuleContext, callNode *ast.Node, callee *ast.Node, args *ast.NodeList, eval *staticEvalCtx) {
	callee = ast.SkipParentheses(callee)
	if !isBuiltinRegExpCallee(ctx, callee) {
		return
	}
	if args == nil || len(args.Nodes) == 0 {
		return
	}

	patternNode := ast.SkipParentheses(args.Nodes[0])
	if patternNode == nil {
		return
	}

	pattern, patternOk := eval.evalStaticString(patternNode)
	if !patternOk {
		return
	}

	flags := ""
	if len(args.Nodes) >= 2 {
		flagsNode := ast.SkipParentheses(args.Nodes[1])
		if flagsNode != nil {
			if v, ok := eval.evalStaticString(flagsNode); ok {
				flags = v
			}
		}
	}

	rxFlags := utils.ParseRegexFlags(flags)
	checkRegex(ctx, callNode, pattern, rxFlags)
}

// isRegexLiteralHandledByConstructor returns true when this regex literal is
// the first arg of a `RegExp(literal, flags)` call — in that case the
// constructor listener owns it (using the override flags).
func isRegexLiteralHandledByConstructor(ctx rule.RuleContext, node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil {
		return false
	}
	var callee *ast.Node
	var args *ast.NodeList
	switch parent.Kind {
	case ast.KindCallExpression:
		c := parent.AsCallExpression()
		callee = c.Expression
		args = c.Arguments
	case ast.KindNewExpression:
		n := parent.AsNewExpression()
		callee = n.Expression
		args = n.Arguments
	default:
		return false
	}
	if !isBuiltinRegExpCallee(ctx, ast.SkipParentheses(callee)) {
		return false
	}
	if args == nil || len(args.Nodes) == 0 {
		return false
	}
	if first := ast.SkipParentheses(args.Nodes[0]); first != node {
		return false
	}
	return true
}

func isBuiltinRegExpCallee(ctx rule.RuleContext, callee *ast.Node) bool {
	if callee == nil {
		return false
	}
	if callee.Kind == ast.KindIdentifier {
		name := callee.AsIdentifier().Text
		if name == "RegExp" {
			// Direct `RegExp` reference — must not be shadowed.
			if utils.IsShadowed(callee, "RegExp") {
				return false
			}
			if ctx.TypeChecker != nil {
				sym := ctx.TypeChecker.GetSymbolAtLocation(callee)
				if sym == nil {
					return false
				}
				return !utils.IsSymbolDeclaredInFile(sym, ctx.SourceFile)
			}
			return true
		}
		// Identifier alias such as `const r = RegExp; new r(...)` — only the
		// type check can recognize this. No syntactic fallback to avoid
		// over-matching arbitrary identifiers when type info is unavailable.
		if ctx.TypeChecker != nil && ctx.Program != nil {
			t := ctx.TypeChecker.GetTypeAtLocation(callee)
			if t != nil && utils.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, t, "RegExpConstructor") {
				return true
			}
		}
		return false
	}
	if callee.Kind == ast.KindPropertyAccessExpression {
		pae := callee.AsPropertyAccessExpression()
		if pae.Name() != nil && pae.Name().Kind == ast.KindIdentifier && pae.Name().AsIdentifier().Text == "RegExp" {
			if pae.Expression != nil && pae.Expression.Kind == ast.KindIdentifier {
				name := pae.Expression.AsIdentifier().Text
				return name == "globalThis" || name == "window" || name == "self" || name == "global"
			}
		}
	}
	return false
}

// checkRegex parses the pattern and reports useless backreferences. `node` is
// the AST node receiving the diagnostic (the regex literal or RegExp call).
func checkRegex(ctx rule.RuleContext, node *ast.Node, pattern string, flags utils.RegexFlags) {
	_, brefs, ok := parsePattern(pattern, flags)
	if !ok {
		return
	}

	for _, bref := range brefs {
		analyzeBackref(ctx, node, bref)
	}
}

type problem struct {
	messageId string
	group     *rxNode
}

func analyzeBackref(ctx rule.RuleContext, node *ast.Node, bref *rxNode) {
	groups := bref.resolved
	if len(groups) == 0 {
		return
	}
	brefPath := pathToRoot(bref)

	problems := make([]*problem, 0, len(groups))
	for _, group := range groups {
		problems = append(problems, classifyPair(brefPath, bref, group))
	}

	// If any pair has no problem, the backreference can match — bail.
	for _, prob := range problems {
		if prob == nil {
			return
		}
	}

	// Prefer same-disjunction problems over disjunctive ones.
	sameDisjunction := make([]*problem, 0, len(problems))
	for _, prob := range problems {
		if prob.messageId != "disjunctive" {
			sameDisjunction = append(sameDisjunction, prob)
		}
	}
	toReport := problems
	if len(sameDisjunction) > 0 {
		toReport = sameDisjunction
	}

	first := toReport[0]
	otherGroups := ""
	otherCount := len(toReport) - 1
	switch {
	case otherCount == 1:
		otherGroups = " and another group"
	case otherCount > 1:
		otherGroups = fmt.Sprintf(" and other %d groups", otherCount)
	}

	desc := descriptionFor(first.messageId, bref.raw, first.group.raw, otherGroups)
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          first.messageId,
		Description: desc,
	})
}

func classifyPair(brefPath []*rxNode, bref *rxNode, group *rxNode) *problem {
	if nodeContains(brefPath, group) {
		// Group is bref's ancestor → bref is nested within the group, which
		// hasn't matched yet when bref starts to match.
		return &problem{messageId: "nested", group: group}
	}

	groupPath := pathToRoot(group)

	// Walk both paths from the root downward to find the lowest common ancestor.
	i := len(brefPath) - 1
	j := len(groupPath) - 1
	for i >= 0 && j >= 0 && brefPath[i] == groupPath[j] {
		i--
		j--
	}
	indexOfLCA := j + 1
	groupCut := groupPath[:indexOfLCA]
	commonPath := groupPath[indexOfLCA:]

	var lowestCommonLookaround *rxNode
	for _, n := range commonPath {
		if isLookaround(n) {
			lowestCommonLookaround = n
			break
		}
	}
	matchingBackward := lowestCommonLookaround != nil && isLookbehind(lowestCommonLookaround)

	if len(groupCut) > 0 && groupCut[len(groupCut)-1].kind == nkAlternative {
		return &problem{messageId: "disjunctive", group: group}
	}
	if !matchingBackward && bref.end <= group.start {
		return &problem{messageId: "forward", group: group}
	}
	if matchingBackward && group.end <= bref.start {
		return &problem{messageId: "backward", group: group}
	}
	for _, n := range groupCut {
		if isNegativeLookaround(n) {
			return &problem{messageId: "intoNegativeLookaround", group: group}
		}
	}
	return nil
}

func descriptionFor(messageId, bref, group, otherGroups string) string {
	switch messageId {
	case "nested":
		return fmt.Sprintf("Backreference '%s' will be ignored. It references group '%s'%s from within that group.", bref, group, otherGroups)
	case "forward":
		return fmt.Sprintf("Backreference '%s' will be ignored. It references group '%s'%s which appears later in the pattern.", bref, group, otherGroups)
	case "backward":
		return fmt.Sprintf("Backreference '%s' will be ignored. It references group '%s'%s which appears before in the same lookbehind.", bref, group, otherGroups)
	case "disjunctive":
		return fmt.Sprintf("Backreference '%s' will be ignored. It references group '%s'%s which is in another alternative.", bref, group, otherGroups)
	case "intoNegativeLookaround":
		return fmt.Sprintf("Backreference '%s' will be ignored. It references group '%s'%s which is in a negative lookaround.", bref, group, otherGroups)
	}
	return ""
}
