package forbid_elements

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// identifierStartRegex mirrors upstream's `/^[A-Z_]/` test against a bare
// `Identifier` argument: only PascalCase or `_`-prefixed identifiers are
// considered candidates for the forbid list (component or DOM-shorthand
// names). `$Foo`, `button1`, lowercase-leading names are silently skipped.
var identifierStartRegex = regexp.MustCompile(`^[A-Z_]`)

// literalElementRegex mirrors upstream's `/^[a-z][^.]*$/` test against a
// string-literal argument: must start with a lowercase ASCII letter and
// contain no `.`. Strings starting uppercase, `_`, digits, or containing a
// dot (`"dotted.component"`) fall through unreported — only the
// MemberExpression branch handles dotted identifiers in the createElement
// path.
var literalElementRegex = regexp.MustCompile(`^[a-z][^.]*$`)

const (
	msgForbiddenElement        = "<{{element}}> is forbidden"
	msgForbiddenElementMessage = "<{{element}}> is forbidden, {{message}}"
)

type forbidEntry struct {
	element string
	message string
}

func parseOptions(options any) map[string]forbidEntry {
	indexed := map[string]forbidEntry{}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return indexed
	}
	raw, ok := optsMap["forbid"].([]interface{})
	if !ok {
		return indexed
	}
	for _, item := range raw {
		switch v := item.(type) {
		case string:
			if v == "" {
				continue
			}
			// Mirror upstream `indexedForbidConfigs[item] = { element: item }`:
			// re-setting an existing key replaces the previous entry, so
			// duplicate `forbid: ["button", { element: "button", message: "..." }]`
			// configs resolve to the LAST occurrence's message (covered by an
			// upstream invalid test).
			indexed[v] = forbidEntry{element: v}
		case map[string]interface{}:
			elem, _ := v["element"].(string)
			if elem == "" {
				continue
			}
			entry := forbidEntry{element: elem}
			if msg, ok := v["message"].(string); ok {
				entry.message = msg
			}
			indexed[elem] = entry
		}
	}
	return indexed
}

func reportIfForbidden(ctx rule.RuleContext, indexed map[string]forbidEntry, element string, node *ast.Node) {
	if element == "" {
		return
	}
	entry, ok := indexed[element]
	if !ok {
		return
	}
	if entry.message != "" {
		desc := strings.ReplaceAll(msgForbiddenElementMessage, "{{element}}", element)
		desc = strings.ReplaceAll(desc, "{{message}}", entry.message)
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "forbiddenElement_message",
			Description: desc,
			Data:        map[string]string{"element": element, "message": entry.message},
		})
		return
	}
	desc := strings.ReplaceAll(msgForbiddenElement, "{{element}}", element)
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "forbiddenElement",
		Description: desc,
		Data:        map[string]string{"element": element},
	})
}

// isCreateElementCall mirrors upstream's `isCreateElement(context, node)`
// for argument-shape semantics, including optional-chain pragma access.
//
// Recognized callee shapes:
//
//   - `<pragma>.createElement(arg)`
//   - `<pragma>?.createElement(arg)`         (optional chain on the access)
//   - `createElement(arg)` after `import { createElement } from '<pragma>'`
//     / `const { createElement } = <pragma>` / `const { createElement } =
//     require('<pragma>')` (resolved via reactutil.IsDestructuredFromPragmaImport,
//     with a syntax-only fallback when the TypeChecker is nil)
//
// The Checker is optional — only used to refine the bare-callee branch.
func isCreateElementCall(callee *ast.Node, pragma string, tc *checker.Checker) bool {
	if callee == nil {
		return false
	}
	if pragma == "" {
		// Defensive: GetReactPragma returns the default ("React") when
		// settings.react.pragma is unset, but a caller could pass an
		// empty string explicitly. Mirror the shared helper's fallback
		// so behavior is consistent in that edge.
		pragma = "React"
	}
	callee = reactutil.SkipExpressionWrappers(callee)

	// Bare callee: `createElement(arg)` from a destructured import.
	if callee.Kind == ast.KindIdentifier {
		if callee.AsIdentifier().Text != "createElement" {
			return false
		}
		return reactutil.IsDestructuredFromPragmaImport(callee, pragma, tc)
	}

	// Member-access callee: `<pragma>.createElement(arg)` or
	// `<pragma>?.createElement(arg)`. Optional chain on the access does
	// NOT disqualify — upstream `isCreateElement` only inspects
	// `callee.property.name === 'createElement'` and
	// `callee.object.name === pragma`, both shape checks that pass through
	// `?.` transparently.
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	prop := callee.AsPropertyAccessExpression()
	nameNode := prop.Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier || nameNode.AsIdentifier().Text != "createElement" {
		return false
	}
	pragmaExpr := reactutil.SkipExpressionWrappers(prop.Expression)
	if pragmaExpr.Kind != ast.KindIdentifier || pragmaExpr.AsIdentifier().Text != pragma {
		return false
	}
	return true
}

var ForbidElementsRule = rule.Rule{
	Name: "react/forbid-elements",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		indexed := parseOptions(options)
		if len(indexed) == 0 {
			return rule.RuleListeners{}
		}

		// checkJsxTag mirrors upstream `getText(context, node.name)` —
		// `TrimNodeTextRange` skips leading trivia, so the dotted /
		// namespaced canonical form (e.g. `dotted.Component`, `ns:Name`,
		// `Foo.Bar.Baz`, `this.Foo`) comes through byte-for-byte for any
		// wrapper-free shape. JSX grammar enforces wrapper-freeness at
		// tag-name positions (no parens, type assertions, or non-null
		// allowed), so this captures the full grammar.
		checkJsxTag := func(element *ast.Node) {
			tagName := reactutil.GetJsxTagName(element)
			if tagName == nil {
				return
			}
			reportIfForbidden(ctx, indexed, utils.TrimmedNodeText(ctx.SourceFile, tagName), tagName)
		}

		pragma := reactutil.GetReactPragma(ctx.Settings)

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     checkJsxTag,
			ast.KindJsxSelfClosingElement: checkJsxTag,
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				// `isCreateElement` mirrors upstream's util byte-for-byte,
				// including for optional-chain pragma access. The shared
				// `reactutil.IsCreateElementCallWithChecker` is intentionally
				// not used here because it bails on optional-chain callees
				// (a conservative choice that's tolerated by other rules but
				// not by upstream `forbid-elements`, which inspects
				// `callee.property.name` / `callee.object.name` directly —
				// neither sensitive to optionality).
				if !isCreateElementCall(call.Expression, pragma, ctx.TypeChecker) {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				// Mirror upstream's three argument-shape branches. Wrapper
				// kinds beyond `ParenthesizedExpression` (TS non-null, `as`,
				// `satisfies`, type assertions, `<T>x`) intentionally fall
				// through unreported, matching upstream's exact
				// `argument.type === '...'` dispatch — only an unwrapped
				// Identifier / StringLiteral / MemberExpression is
				// considered. Template literals (`` `button` ``) are also
				// skipped: upstream's `argument.type === 'Literal'` doesn't
				// match TemplateLiteral, and tsgo's KindNoSubstitutionTemplateLiteral
				// is similarly excluded by this switch.
				arg := ast.SkipParentheses(call.Arguments.Nodes[0])
				switch arg.Kind {
				case ast.KindIdentifier:
					name := arg.AsIdentifier().Text
					if identifierStartRegex.MatchString(name) {
						reportIfForbidden(ctx, indexed, name, arg)
					}
				case ast.KindStringLiteral:
					text := arg.AsStringLiteral().Text
					if literalElementRegex.MatchString(text) {
						reportIfForbidden(ctx, indexed, text, arg)
					}
				case ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
					// In ESTree, `true` / `false` / `null` are `Literal` nodes;
					// upstream applies `String(argument.value)` which yields
					// `"true"` / `"false"` / `"null"` — all three match
					// `/^[a-z][^.]*$/`. tsgo splits these into dedicated
					// keyword kinds, so handle them explicitly to preserve
					// upstream's (degenerate but observable) behavior:
					// `React.createElement(true)` + `forbid: ['true']` reports.
					var text string
					switch arg.Kind {
					case ast.KindTrueKeyword:
						text = "true"
					case ast.KindFalseKeyword:
						text = "false"
					case ast.KindNullKeyword:
						text = "null"
					}
					reportIfForbidden(ctx, indexed, text, arg)
				case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
					// Upstream (espree on ESLint ≥8) wraps optional-chain
					// member access in `ChainExpression`, so `argument.type`
					// equals `'ChainExpression'` rather than `'MemberExpression'`
					// and the branch is skipped. tsgo encodes the optional
					// chain as a flag on the same kind, so guard explicitly.
					if ast.IsOptionalChain(arg) {
						return
					}
					reportIfForbidden(ctx, indexed, utils.TrimmedNodeText(ctx.SourceFile, arg), arg)
				}
			},
		}
	},
}
