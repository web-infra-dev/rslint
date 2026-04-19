package no_sequences

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type options struct {
	allowInParentheses bool
}

func parseOptions(raw any) options {
	opts := options{allowInParentheses: true}
	optsMap := utils.GetOptionsMap(raw)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["allowInParentheses"].(bool); ok {
		opts.allowInParentheses = v
	}
	return opts
}

// isCommaBinary reports whether node is a BinaryExpression whose operator is
// the comma token — tsgo's collapsed form of ESLint's SequenceExpression.
func isCommaBinary(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := node.AsBinaryExpression()
	return bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindCommaToken
}

// walkUpSkippingParens returns the first ancestor of node that is not a
// ParenthesizedExpression, along with the count of ParenthesizedExpression
// wrappers traversed. The returned child is the direct descendant of that
// ancestor (i.e. the node whose Parent is the returned ancestor).
func walkUpSkippingParens(node *ast.Node) (parent *ast.Node, child *ast.Node, parenDepth int) {
	child = node
	parent = node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parenDepth++
		child = parent
		parent = parent.Parent
	}
	return
}

// isForInitOrUpdate reports whether `child` sits in the init or update slot
// of the `for` statement `parent`. Callers pass the paren-skipped ancestor
// pair from walkUpSkippingParens. ESLint always allows sequences in those
// positions regardless of `allowInParentheses`.
func isForInitOrUpdate(parent, child *ast.Node) bool {
	if parent == nil || parent.Kind != ast.KindForStatement {
		return false
	}
	forStmt := parent.AsForStatement()
	if forStmt == nil {
		return false
	}
	return child == forStmt.Initializer || child == forStmt.Incrementor
}

// isGrammarParenArrowBody reports whether `child` occupies the body slot of
// the arrow function `parent`. Callers pass the paren-skipped ancestor pair
// from walkUpSkippingParens. In tsgo the disambiguating parens in
// `a => (x, y)` materialize as a ParenthesizedExpression wrapper, so arrow
// bodies — unlike do-while / while / if / switch / with grammar parens,
// which are consumed by the statement — need *two* wrappers before a
// sequence expression counts as explicitly parenthesised.
func isGrammarParenArrowBody(parent, child *ast.Node) bool {
	if parent == nil || !ast.IsArrowFunction(parent) {
		return false
	}
	arrow := parent.AsArrowFunction()
	return arrow != nil && arrow.Body == child
}

// firstCommaToken walks down the left spine of a comma BinaryExpression chain
// to find the token at the first comma — matching ESLint's
// `sourceCode.getTokenAfter(node.expressions[0], isCommaToken)`. Parenthesized
// wrappers are skipped so `(a, b), c` still reports at the inner `,`.
func firstCommaToken(node *ast.Node) *ast.Node {
	current := node
	for {
		bin := current.AsBinaryExpression()
		left := ast.SkipParentheses(bin.Left)
		if !isCommaBinary(left) {
			return bin.OperatorToken
		}
		current = left
	}
}

// https://eslint.org/docs/latest/rules/no-sequences
var NoSequencesRule = rule.Rule{
	Name: "no-sequences",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !isCommaBinary(node) {
					return
				}
				// Single walk-up; all downstream checks read from these.
				parent, child, parenDepth := walkUpSkippingParens(node)

				// Only report once per comma chain — skip inner nodes of
				// `(a, b), c` / `a, b, c`.
				if isCommaBinary(parent) {
					return
				}
				// `for (init; cond; update)` — ESLint unconditionally allows
				// sequences in init/update, even when allowInParentheses is
				// false.
				if isForInitOrUpdate(parent, child) {
					return
				}

				if opts.allowInParentheses {
					required := 1
					if isGrammarParenArrowBody(parent, child) {
						// Arrow body parens are materialized as a
						// ParenthesizedExpression wrapper in tsgo, so an
						// "extra" pair of parens bumps the wrapper count to 2.
						required = 2
					}
					if parenDepth >= required {
						return
					}
				}

				commaToken := firstCommaToken(node)
				if commaToken == nil {
					return
				}
				ctx.ReportRange(
					utils.TrimNodeTextRange(ctx.SourceFile, commaToken),
					rule.RuleMessage{
						Id:          "unexpectedCommaExpression",
						Description: "Unexpected use of comma operator.",
					},
				)
			},
		}
	},
}
