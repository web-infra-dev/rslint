// cspell:ignore rctx
package no_useless_backreference

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/evaluator"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// staticEvalCtx folds string-typed expressions to constant values for the
// purposes of resolving the pattern / flags arguments of a `RegExp(...)` or
// `new RegExp(...)` call. Mirrors the helper in the
// no-misleading-character-class rule.
type staticEvalCtx struct {
	rctx rule.RuleContext

	evaluator evaluator.Evaluator

	writeRefsComputed bool
	writeRefsMap      map[*ast.Symbol]bool
}

func newStaticEvalCtx(rctx rule.RuleContext) *staticEvalCtx {
	s := &staticEvalCtx{rctx: rctx}
	s.evaluator = evaluator.NewEvaluator(s.evaluateEntity, ast.OEKParentheses)
	return s
}

func (s *staticEvalCtx) writeRefs() map[*ast.Symbol]bool {
	if s.writeRefsComputed {
		return s.writeRefsMap
	}
	s.writeRefsComputed = true
	if s.rctx.TypeChecker == nil || s.rctx.SourceFile == nil {
		return nil
	}
	m := map[*ast.Symbol]bool{}
	var visit func(n *ast.Node)
	visit = func(n *ast.Node) {
		if n == nil {
			return
		}
		if n.Kind == ast.KindIdentifier && utils.IsWriteReference(n) {
			if sym := s.rctx.TypeChecker.GetSymbolAtLocation(n); sym != nil {
				m[sym] = true
			}
		}
		n.ForEachChild(func(c *ast.Node) bool {
			visit(c)
			return false
		})
	}
	visit(&s.rctx.SourceFile.Node)
	s.writeRefsMap = m
	return m
}

func (s *staticEvalCtx) evalStaticString(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	if raw, ok := s.evalStringRawTag(node); ok {
		return raw, true
	}
	res := s.evaluator(node, node)
	if v, ok := res.Value.(string); ok {
		return v, true
	}
	return "", false
}

func (s *staticEvalCtx) evaluateEntity(expr *ast.Node, location *ast.Node) evaluator.Result {
	if s.rctx.TypeChecker == nil {
		return evaluator.Result{}
	}
	expr = ast.SkipParentheses(expr)
	if expr == nil || expr.Kind != ast.KindIdentifier {
		return evaluator.Result{}
	}
	sym := s.rctx.TypeChecker.GetSymbolAtLocation(expr)
	if sym == nil || len(sym.Declarations) != 1 {
		return evaluator.Result{}
	}
	decl := sym.Declarations[0]
	if decl.Kind != ast.KindVariableDeclaration {
		return evaluator.Result{}
	}
	varDecl := decl.AsVariableDeclaration()
	if varDecl == nil || varDecl.Initializer == nil {
		return evaluator.Result{}
	}
	list := decl.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList {
		return evaluator.Result{}
	}
	if list.Flags&ast.NodeFlagsConst == 0 && s.writeRefs()[sym] {
		return evaluator.Result{}
	}
	return s.evaluator(varDecl.Initializer, location)
}

func (s *staticEvalCtx) evalStringRawTag(node *ast.Node) (string, bool) {
	if node.Kind != ast.KindTaggedTemplateExpression {
		return "", false
	}
	tt := node.AsTaggedTemplateExpression()
	if tt == nil || tt.Tag == nil || tt.Template == nil || !s.isStringRawTag(tt.Tag) {
		return "", false
	}
	switch tt.Template.Kind {
	case ast.KindNoSubstitutionTemplateLiteral:
		return tt.Template.Text(), true
	case ast.KindTemplateExpression:
		te := tt.Template.AsTemplateExpression()
		if te == nil {
			return "", false
		}
		var sb strings.Builder
		if te.Head != nil {
			sb.WriteString(te.Head.Text())
		}
		if te.TemplateSpans != nil {
			for _, span := range te.TemplateSpans.Nodes {
				sp := span.AsTemplateSpan()
				if sp == nil {
					return "", false
				}
				sub, ok := s.evalStaticString(sp.Expression)
				if !ok {
					return "", false
				}
				sb.WriteString(sub)
				if sp.Literal != nil {
					sb.WriteString(sp.Literal.Text())
				}
			}
		}
		return sb.String(), true
	}
	return "", false
}

func (s *staticEvalCtx) isStringRawTag(tag *ast.Node) bool {
	tag = ast.SkipParentheses(tag)
	if tag == nil || tag.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	pae := tag.AsPropertyAccessExpression()
	if pae == nil {
		return false
	}
	name := pae.Name()
	if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "raw" {
		return false
	}
	obj := ast.SkipParentheses(pae.Expression)
	if obj == nil {
		return false
	}
	if s.rctx.TypeChecker != nil && s.rctx.Program != nil {
		if t := s.rctx.TypeChecker.GetTypeAtLocation(obj); t != nil {
			if utils.IsBuiltinSymbolLike(s.rctx.Program, s.rctx.TypeChecker, t, "StringConstructor") {
				return true
			}
		}
	}
	if obj.Kind == ast.KindIdentifier {
		return obj.AsIdentifier().Text == "String"
	}
	return false
}
