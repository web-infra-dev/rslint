package throw_new_error

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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

				// A decorator call is not a construction site. Skip syntax that
				// tsgo preserves or synthesizes but ESTree does not expose.
				parent := utils.ESTreeParent(node)
				if parent != nil && parent.Kind == ast.KindDecorator {
					return
				}

				// ESTree omits parentheses and JavaScript JSDoc casts. Authored
				// TypeScript assertions remain intact, so a callee such as
				// `(Error as any)` stays unreported exactly as upstream leaves it.
				callee := utils.ESTreeRuntimeExpression(call.Expression)
				if callee == nil {
					return
				}

				// Walk the whole callee rather than testing the outermost node.
				// The optional-chain flag stops at a parenthesis — `(a?.b).c`
				// ends the chain — but upstream's hasOptionalChainElement keeps
				// walking through it and skips the call either way.
				if unicornutil.HasOptionalChainElement(callee) {
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
