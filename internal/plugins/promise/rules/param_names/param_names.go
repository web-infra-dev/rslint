package param_names

import (
	"fmt"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// skipTransparent only unwraps parens — ESLint's ESTree parser drops
// parentheses, so `new (Promise)(...)` and `new Promise((fn))` already have
// the identifier / function visible at the ESLint level. TS-only wrappers
// (type assertions, non-null, satisfies) are intentionally NOT unwrapped:
// the original rule treats `new (Promise as any)(...)` as not-a-Promise-
// constructor (callee.type !== 'Identifier') and skips silently, and we
// mirror that — a user writing `as` is often deliberately signalling "this
// isn't the standard Promise constructor, don't lint me".
const skipTransparent = ast.OEKParentheses

const (
	defaultResolvePattern = "^_?resolve$"
	defaultRejectPattern  = "^_?reject$"
)

type Options struct {
	ResolvePattern string
	RejectPattern  string
}

func parseOptions(options any) Options {
	opts := Options{
		ResolvePattern: defaultResolvePattern,
		RejectPattern:  defaultRejectPattern,
	}
	optsMap := utils.GetOptionsMap(options)
	if optsMap != nil {
		if v, ok := optsMap["resolvePattern"].(string); ok && v != "" {
			opts.ResolvePattern = v
		}
		if v, ok := optsMap["rejectPattern"].(string); ok && v != "" {
			opts.RejectPattern = v
		}
	}
	return opts
}

func buildResolveParamNamesMessage(pattern string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "resolveParamNames",
		Description: fmt.Sprintf(`Promise constructor parameters must be named to match "%s"`, pattern),
	}
}

func buildRejectParamNamesMessage(pattern string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "rejectParamNames",
		Description: fmt.Sprintf(`Promise constructor parameters must be named to match "%s"`, pattern),
	}
}

// paramName returns the plain identifier name of a parameter, or "" if the
// parameter is a destructuring pattern, has a default value (AssignmentPattern
// in ESTree), or is a rest element — matching ESLint's `params[i].name`, which
// is only defined on Identifier-shaped parameters.
func paramName(param *ast.Node) string {
	if param == nil || !ast.IsParameter(param) {
		return ""
	}
	decl := param.AsParameterDeclaration()
	if decl == nil {
		return ""
	}
	if decl.Initializer != nil || decl.DotDotDotToken != nil {
		return ""
	}
	name := decl.Name()
	if name == nil || !ast.IsIdentifier(name) {
		return ""
	}
	return name.AsIdentifier().Text
}

// regexMatch wraps regexp2.MatchString, discarding the timeout error.
func regexMatch(re *regexp2.Regexp, s string) bool {
	matched, _ := re.MatchString(s)
	return matched
}

var ParamNamesRule = rule.Rule{
	Name: "promise/param-names",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		// ECMAScript + Unicode flags mirror ESLint's `new RegExp(pattern, 'u')`
		// so user patterns using lookaround, backreferences, or `\p{...}` work
		// identically to the original rule (Go's standard `regexp` / RE2 does not).
		const reOpts = regexp2.ECMAScript | regexp2.Unicode
		resolveRe, err := regexp2.Compile(opts.ResolvePattern, reOpts)
		if err != nil {
			return rule.RuleListeners{}
		}
		rejectRe, err := regexp2.Compile(opts.RejectPattern, reOpts)
		if err != nil {
			return rule.RuleListeners{}
		}

		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				callee := ast.SkipOuterExpressions(node.AsNewExpression().Expression, skipTransparent)
				if callee == nil || !ast.IsIdentifier(callee) || callee.AsIdentifier().Text != "Promise" {
					return
				}
				args := node.Arguments()
				if len(args) != 1 {
					return
				}
				executor := ast.SkipOuterExpressions(args[0], skipTransparent)
				if executor == nil || !ast.IsFunctionExpressionOrArrowFunction(executor) {
					return
				}
				// Filter out the TS `this` parameter — it's a tsgo-specific
				// ParameterDeclaration that @typescript-eslint/parser strips
				// from `.params` before the ESLint rule sees it.
				params := make([]*ast.Node, 0, len(executor.Parameters()))
				for _, p := range executor.Parameters() {
					if !ast.IsThisParameter(p) {
						params = append(params, p)
					}
				}
				if len(params) == 0 {
					return
				}

				if resolveName := paramName(params[0]); resolveName != "" && !regexMatch(resolveRe, resolveName) {
					ctx.ReportNode(params[0], buildResolveParamNamesMessage(opts.ResolvePattern))
				}
				if len(params) >= 2 {
					if rejectName := paramName(params[1]); rejectName != "" && !regexMatch(rejectRe, rejectName) {
						ctx.ReportNode(params[1], buildRejectParamNamesMessage(opts.RejectPattern))
					}
				}
			},
		}
	},
}
