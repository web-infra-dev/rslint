package prefer_arrow_callback

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// prefer-arrow-callback requires arrow functions for callbacks when doing so
// preserves the callback's binding semantics.
// https://eslint.org/docs/latest/rules/prefer-arrow-callback

type options struct {
	allowNamedFunctions bool
	allowUnboundThis    bool
}

type scopeInfo struct {
	this  bool
	super bool
	meta  bool
}

type callbackInfo struct {
	isCallback    bool
	isLexicalThis bool
	bindMember    *ast.Node
	bindCall      *ast.Node
}

func parseOptions(raw any) options {
	opts := options{allowUnboundThis: true}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if v, ok := m["allowNamedFunctions"].(bool); ok {
		opts.allowNamedFunctions = v
	}
	if v, ok := m["allowUnboundThis"].(bool); ok {
		opts.allowUnboundThis = v
	}
	return opts
}

func preferArrowCallbackMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferArrowCallback",
		Description: "Unexpected function expression.",
	}
}

func isLogicalExpression(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := node.AsBinaryExpression()
	if bin == nil || bin.OperatorToken == nil {
		return false
	}
	switch bin.OperatorToken.Kind {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
		return true
	default:
		return false
	}
}

func functionScopeRoots(node *ast.Node) []*ast.Node {
	roots := make([]*ast.Node, 0, len(node.Parameters())+1)
	roots = append(roots, node.Parameters()...)
	if body := node.Body(); body != nil {
		roots = append(roots, body)
	}
	return roots
}

func walkCurrentFunctionScope(node *ast.Node, visit func(*ast.Node)) {
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		switch n.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
			return
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			// Method bodies have their own `this`/`arguments`, but computed
			// method names run in the containing scope and can still make an
			// outer callback unsafe to convert.
			walk(n.Name())
			return
		case ast.KindConstructor:
			return
		}

		visit(n)
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	for _, root := range functionScopeRoots(node) {
		walk(root)
	}
}

func getFunctionScopeInfo(node *ast.Node) scopeInfo {
	info := scopeInfo{}
	walkCurrentFunctionScope(node, func(n *ast.Node) {
		switch n.Kind {
		case ast.KindThisKeyword:
			info.this = true
		case ast.KindSuperKeyword:
			info.super = true
		case ast.KindMetaProperty:
			meta := n.AsMetaProperty()
			name := meta.Name()
			if meta.KeywordToken == ast.KindNewKeyword && name != nil && name.Text() == "target" {
				info.meta = true
			}
		}
	})
	return info
}

func hasDuplicateParams(params []*ast.Node) bool {
	seen := map[string]bool{}
	for _, param := range params {
		if param == nil || param.Name() == nil || param.Name().Kind != ast.KindIdentifier {
			return false
		}
		name := param.Name().Text()
		if seen[name] {
			return true
		}
		seen[name] = true
	}
	return false
}

func firstParamName(node *ast.Node) string {
	params := node.Parameters()
	if len(params) == 0 || params[0] == nil || params[0].Name() == nil {
		return ""
	}
	name := params[0].Name()
	if name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.Text()
}

func hasArgumentsBinding(node *ast.Node) bool {
	found := false
	for _, param := range node.Parameters() {
		if param == nil {
			continue
		}
		utils.CollectBindingNames(param.Name(), func(_ *ast.Node, name string) {
			if name == "arguments" {
				found = true
			}
		})
	}
	if found {
		return true
	}
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || found {
			return
		}
		switch n.Kind {
		case ast.KindFunctionDeclaration:
			if name := n.Name(); name != nil && name.Kind == ast.KindIdentifier && name.Text() == "arguments" {
				found = true
			}
			return
		case ast.KindFunctionExpression, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
			return
		case ast.KindVariableDeclaration:
			utils.CollectBindingNames(n.AsVariableDeclaration().Name(), func(_ *ast.Node, name string) {
				if name == "arguments" {
					found = true
				}
			})
		case ast.KindClassDeclaration:
			if name := n.Name(); name != nil && name.Kind == ast.KindIdentifier && name.Text() == "arguments" {
				found = true
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	for _, root := range functionScopeRoots(node) {
		walk(root)
	}
	return found
}

func isValueIdentifier(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	if ast.IsPartOfTypeNode(node) {
		return false
	}
	return !utils.IsNonReferenceIdentifier(node)
}

func hasArgumentsReference(ctx rule.RuleContext, node *ast.Node) bool {
	found := false
	walkCurrentFunctionScope(node, func(n *ast.Node) {
		if found {
			return
		}
		if isValueIdentifier(n) && n.Text() == "arguments" {
			if ctx.TypeChecker != nil {
				symbol := utils.GetReferenceSymbol(n, ctx.TypeChecker)
				if symbol != nil && len(symbol.Declarations) > 0 {
					return
				}
				found = true
				return
			}
			if hasArgumentsBinding(node) {
				return
			}
			found = true
		}
	})
	return found
}

func functionNameReferenced(ctx rule.RuleContext, node *ast.Node) bool {
	name := node.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}
	targetName := name.Text()
	var targetSymbol *ast.Symbol
	if ctx.TypeChecker != nil {
		targetSymbol = utils.GetReferenceSymbol(name, ctx.TypeChecker)
	}

	found := false
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || found {
			return
		}
		if n == name {
			return
		}
		if isValueIdentifier(n) && n.Text() == targetName {
			if targetSymbol == nil || utils.GetReferenceSymbol(n, ctx.TypeChecker) == targetSymbol {
				found = true
				return
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	for _, root := range functionScopeRoots(node) {
		walk(root)
	}
	return found
}

func getCallForCallee(node *ast.Node) *ast.Node {
	// Deliberately skip only parentheses here. utils.IsCallee also treats TS
	// assertion wrappers as transparent, but upstream getCallbackInfo does not:
	// `(function() {}) as any` is not considered a callback by this rule.
	current := node
	parent := current.Parent
	for parent != nil && ast.IsParenthesizedExpression(parent) {
		current = parent
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return nil
	}
	call := parent.AsCallExpression()
	if call != nil && call.Expression == current {
		return parent
	}
	return nil
}

func callHasSingleThisArg(call *ast.Node) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	args := call.AsCallExpression().Arguments
	if args == nil || len(args.Nodes) != 1 {
		return false
	}
	return ast.SkipParentheses(args.Nodes[0]).Kind == ast.KindThisKeyword
}

func getCallbackInfo(node *ast.Node) callbackInfo {
	ret := callbackInfo{}
	current := node
	parent := current.Parent
	bound := false

	for current != nil && parent != nil {
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			current = parent
			parent = parent.Parent
			continue

		case ast.KindConditionalExpression:
			// ESLint's ESTree walk treats every branch of a ConditionalExpression
			// as transparent while looking for the outer call/new callback site.

		case ast.KindBinaryExpression:
			if !isLogicalExpression(parent) {
				return ret
			}

		case ast.KindPropertyAccessExpression:
			member := parent.AsPropertyAccessExpression()
			name := member.Name()
			if member.Expression != current ||
				name == nil ||
				name.Kind != ast.KindIdentifier ||
				name.Text() != "bind" {
				return ret
			}
			call := getCallForCallee(parent)
			if call == nil {
				return ret
			}
			if !bound {
				bound = true
				ret.isLexicalThis = callHasSingleThisArg(call)
				ret.bindMember = parent
				ret.bindCall = call
			}
			current = call
			parent = call.Parent
			continue

		case ast.KindCallExpression:
			if parent.AsCallExpression().Expression != current {
				ret.isCallback = true
			}
			return ret

		case ast.KindNewExpression:
			if parent.AsNewExpression().Expression != current {
				ret.isCallback = true
			}
			return ret

		default:
			return ret
		}
		current = parent
		parent = parent.Parent
	}

	return ret
}

func findTokenRange(sf *ast.SourceFile, start, end int, kind ast.Kind) (core.TextRange, bool) {
	s := scanner.GetScannerForSourceFile(sf, start)
	for s.Token() != ast.KindEndOfFile && s.TokenStart() < end {
		if s.Token() == kind {
			return core.NewTextRange(s.TokenStart(), s.TokenEnd()), true
		}
		s.Scan()
	}
	return core.TextRange{}, false
}

func previousTokenEndBefore(text string, pos int) int {
	i := pos
	for i > 0 {
		i = utils.SkipTrailingWhitespace(text, 0, i)
		if i >= 2 && text[i-2:i] == "*/" {
			if start := strings.LastIndex(text[:i-2], "/*"); start >= 0 {
				i = start
				continue
			}
		}
		lineStart := strings.LastIndexByte(text[:i], '\n') + 1
		if comment := strings.Index(text[lineStart:i], "//"); comment >= 0 {
			i = lineStart + comment
			continue
		}
		return i
	}
	return pos
}

func sameLine(sf *ast.SourceFile, a, b int) bool {
	lineA, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, a)
	lineB, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, b)
	return lineA == lineB
}

func isParenthesized(node *ast.Node) bool {
	return node != nil && node.Parent != nil && ast.IsParenthesizedExpression(node.Parent)
}

func bindMemberDirectlyWrapsNode(member, node *ast.Node) bool {
	if member == nil || member.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	return ast.SkipParentheses(member.AsPropertyAccessExpression().Expression) == node
}

func buildFixes(ctx rule.RuleContext, node *ast.Node, scope scopeInfo, info callbackInfo) []rule.RuleFix {
	if (!info.isLexicalThis && scope.this) || hasDuplicateParams(node.Parameters()) || firstParamName(node) == "this" {
		return nil
	}

	sf := ctx.SourceFile
	text := sf.Text()
	body := node.Body()
	if body == nil {
		return nil
	}
	bodyStart := scanner.SkipTrivia(text, body.Pos())
	if bodyStart < 0 {
		return nil
	}
	nodeRange := utils.TrimNodeTextRange(sf, node)
	functionToken, ok := findTokenRange(sf, nodeRange.Pos(), bodyStart, ast.KindFunctionKeyword)
	if !ok {
		return nil
	}
	leftParenToken, ok := findTokenRange(sf, functionToken.End(), bodyStart, ast.KindOpenParenToken)
	if !ok {
		return nil
	}
	flags := ast.GetFunctionFlags(node)
	if flags&ast.FunctionFlagsAsync != 0 && !sameLine(sf, functionToken.End(), leftParenToken.Pos()) {
		return nil
	}
	insertArrowAt := previousTokenEndBefore(text, bodyStart)

	fixes := []rule.RuleFix{}
	if info.isLexicalThis {
		if !bindMemberDirectlyWrapsNode(info.bindMember, node) || isParenthesized(info.bindMember) {
			return nil
		}
		object := info.bindMember.AsPropertyAccessExpression().Expression
		removeStart := scanner.SkipTrivia(text, object.End())
		removeEnd := utils.TrimNodeTextRange(sf, info.bindCall).End()
		if removeStart < 0 || removeStart >= removeEnd || utils.HasCommentInSpan(sf, removeStart, removeEnd) {
			return nil
		}
		fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(removeStart, removeEnd)))
	}

	if utils.HasCommentInSpan(sf, functionToken.End(), leftParenToken.Pos()) {
		fixes = append(fixes, rule.RuleFixRemoveRange(functionToken))
		if name := node.Name(); name != nil {
			fixes = append(fixes, rule.RuleFixRemove(sf, name))
		}
	} else {
		fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(functionToken.Pos(), leftParenToken.Pos())))
	}
	fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(insertArrowAt, insertArrowAt), " =>"))

	replacedNode := node
	if info.isLexicalThis {
		replacedNode = info.bindCall
	}
	if replacedNode == nil || replacedNode.Parent == nil {
		return fixes
	}
	if replacedNode.Parent.Kind != ast.KindCallExpression &&
		replacedNode.Parent.Kind != ast.KindConditionalExpression &&
		!isParenthesized(replacedNode) &&
		!isParenthesized(node) {
		fixes = append(fixes,
			rule.RuleFixReplaceRange(core.NewTextRange(utils.TrimNodeTextRange(sf, replacedNode).Pos(), utils.TrimNodeTextRange(sf, replacedNode).Pos()), "("),
			rule.RuleFixReplaceRange(core.NewTextRange(utils.TrimNodeTextRange(sf, replacedNode).End(), utils.TrimNodeTextRange(sf, replacedNode).End()), ")"),
		)
	}

	return fixes
}

var PreferArrowCallbackRule = rule.Rule{
	Name: "prefer-arrow-callback",
	Run: func(ctx rule.RuleContext, _raw []any) rule.RuleListeners {
		raw := rule.UnwrapOptions(_raw)
		opts := parseOptions(raw)
		return rule.RuleListeners{
			ast.KindFunctionExpression: func(node *ast.Node) {
				if opts.allowNamedFunctions {
					if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier && name.Text() != "" {
						return
					}
				}
				if ast.GetFunctionFlags(node)&ast.FunctionFlagsGenerator != 0 {
					return
				}
				if functionNameReferenced(ctx, node) || hasArgumentsReference(ctx, node) {
					return
				}

				scope := getFunctionScopeInfo(node)
				info := getCallbackInfo(node)
				if !info.isCallback {
					return
				}
				if (opts.allowUnboundThis && scope.this && !info.isLexicalThis) || scope.super || scope.meta {
					return
				}

				msg := preferArrowCallbackMessage()
				if fixes := buildFixes(ctx, node, scope, info); len(fixes) > 0 {
					ctx.ReportNodeWithFixes(node, msg, fixes...)
				} else {
					ctx.ReportNode(node, msg)
				}
			},
		}
	},
}
