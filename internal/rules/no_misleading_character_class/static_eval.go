// cspell:ignore rctx
package no_misleading_character_class

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/evaluator"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// staticEvalCtx carries per-rule-invocation state for static-value evaluation.
// A single instance is reused by all listener callbacks within one file so
// that the write-reference map and evaluator are computed at most once per
// file.
//
// The evaluator itself is borrowed from tsgo's `internal/evaluator` package
// (exposed via `shim/evaluator`): NewEvaluator takes an "evaluateEntity"
// callback and returns a function that folds StringLiteral / BinaryExpression
// / PrefixUnaryExpression / TemplateExpression / NumericLiteral into their
// constant values. Our callback handles Identifier / PropertyAccessExpression
// by resolving them via TypeChecker (const or let/var with no write refs).
//
// Things not covered by tsgo's evaluator that we handle here:
//   - TaggedTemplateExpression with `String.raw` tag
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

// writeRefs returns a map from symbol → "has non-initial write reference".
// Computed lazily on first call.
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

// evalStaticString attempts to evaluate `node` to a string constant. Returns
// the value and true on success.
//
// Supported shapes (via tsgo's evaluator):
//   - String / NumericLiteral / no-substitution template / template with
//     constant spans
//   - BinaryExpression with `+` (string concat or numeric addition)
//   - PrefixUnaryExpression with `+` / `-` / `~` on numerics
//   - Identifier / PropertyAccessExpression (via our evaluateEntity callback)
//
// Supported shapes handled directly (not covered by tsgo's evaluator):
//   - `String.raw\`...\`` tagged templates
func (s *staticEvalCtx) evalStaticString(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	// Handle String.raw tagged templates before calling the evaluator,
	// because the evaluator doesn't special-case tag functions.
	if raw, ok := s.evalStringRawTag(node); ok {
		return raw, true
	}
	res := s.evaluator(node, node)
	if s, ok := res.Value.(string); ok {
		return s, true
	}
	return "", false
}

// evaluateEntity resolves Identifier and PropertyAccessExpression nodes for
// the evaluator. It's responsible for producing a constant value for symbol
// references (our const + let/var-with-no-writes policy).
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
	// Recursively evaluate the initializer.
	return s.evaluator(varDecl.Initializer, location)
}

// evalStringRawTag returns the raw template string when `node` is a
// `String.raw\`...\`` TaggedTemplateExpression. Otherwise returns "", false.
//
// tsgo's evaluator doesn't special-case any tag functions, so we bridge
// `String.raw` manually.
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

// isStringRawTag reports whether `tag` refers to the built-in String.raw.
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
