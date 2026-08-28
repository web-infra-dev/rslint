package throw_new_error

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const messageID = "throw-new-error"

var message = rule.RuleMessage{
	Id:          messageID,
	Description: "Use `new` when creating an error.",
}

// customError mirrors upstream's `/^(?:[A-Z][\da-z]*)*Error$/`: zero or more
// capitalized runs, then a literal `Error`. It matches `Error`, `TypeError`,
// `CustomError` and `ABCError`, but not `error` or `getError`.
//
// The pattern is written here rather than taken from a rule option, a config
// file, or the source under lint, and RE2 and JavaScript read it identically,
// which is what AGENTS.md requires of the stdlib regexp.
var customError = regexp.MustCompile(`^(?:[A-Z][\da-z]*)*Error$`)

// taggedErrorPath is Effect's `Data.TaggedError`, a factory rather than a
// constructor. Upstream exempts it.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/issues/2654
const taggedErrorPath = "Data.TaggedError"

// ThrowNewErrorRule requires `new` when creating an error.
//
// The name is historical: like upstream, the rule listens to every call
// expression, not just `throw` statements, so `const error = Error()` is
// reported too.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/throw-new-error.js
var ThrowNewErrorRule = rule.Rule{
	Name:   "unicorn/throw-new-error",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil {
					return
				}

				// `new` cannot be applied to an optional-chained call, e.g.
				// `Error?.()`.
				if call.QuestionDotToken != nil {
					return
				}

				// A decorator call is not a construction site.
				if node.Parent != nil && node.Parent.Kind == ast.KindDecorator {
					return
				}

				// ESTree unwraps parentheses transparently; tsgo keeps
				// ParenthesizedExpression nodes, so the callee has to be
				// unwrapped before it is classified. Only parentheses are
				// skipped, not TS assertions, so that a callee such as
				// `(Error as any)` stays unreported exactly as upstream
				// leaves it.
				callee := ast.SkipParentheses(call.Expression)
				if callee == nil {
					return
				}

				// Walk the whole callee rather than testing the outermost node.
				// The optional-chain flag stops at a parenthesis — `(a?.b).c`
				// ends the chain — but upstream's hasOptionalChainElement keeps
				// walking through it and skips the call either way.
				if hasOptionalChainElement(callee) {
					return
				}

				if unicornutil.NodeMatchesPath(callee, taggedErrorPath) {
					return
				}

				if !isCustomErrorCallee(callee) {
					return
				}

				ctx.ReportNode(node, message)
			},
		}
	},
}

// hasOptionalChainElement reports whether any link of the callee chain is
// optional, mirroring upstream's helper of the same name.
//
// `new` cannot be applied to an optional-chained call, so `Error?.()`,
// `lib?.Error()`, `lib?.foo.Error()` and `(lib?.mod).Error()` are all left
// alone.
func hasOptionalChainElement(node *ast.Node) bool {
	for current := node; current != nil; {
		if ast.IsOptionalChain(current) {
			return true
		}
		switch current.Kind {
		case ast.KindCallExpression:
			call := current.AsCallExpression()
			if call == nil {
				return false
			}
			current = ast.SkipParentheses(call.Expression)
		case ast.KindPropertyAccessExpression:
			access := current.AsPropertyAccessExpression()
			if access == nil {
				return false
			}
			current = ast.SkipParentheses(access.Expression)
		case ast.KindElementAccessExpression:
			access := current.AsElementAccessExpression()
			if access == nil {
				return false
			}
			current = ast.SkipParentheses(access.Expression)
		case ast.KindNonNullExpression:
			nonNull := current.AsNonNullExpression()
			if nonNull == nil {
				return false
			}
			current = ast.SkipParentheses(nonNull.Expression)
		default:
			return false
		}
	}
	return false
}

// isCustomErrorCallee reports whether callee names an error constructor: a bare
// identifier such as `Error` or `TypeError`, or a static member such as
// `lib.Error` or `lib.mod.Error`.
//
// A computed access is an element access in tsgo and never reaches here, which
// is why `lib[Error]()` and `lib["Error"]()` stay unreported, matching
// upstream's `!callee.computed` check.
func isCustomErrorCallee(callee *ast.Node) bool {
	switch callee.Kind {
	case ast.KindIdentifier:
		return customError.MatchString(callee.AsIdentifier().Text)
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access == nil || access.Name() == nil {
			return false
		}
		return customError.MatchString(access.Name().Text())
	default:
		return false
	}
}
