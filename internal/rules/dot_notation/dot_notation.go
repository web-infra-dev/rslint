package dot_notation

import (
	_ "embed"
	"encoding/json"
	"regexp"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed dot-notation.schema.json
var schemaJSON []byte

// Options mirrors ESLint core's dot-notation rule options
// (see eslint/lib/rules/dot-notation.js meta.schema).
type Options struct {
	AllowKeywords bool
	AllowPattern  string
}

func parseOptions(options any) Options {
	opts := Options{AllowKeywords: true}
	optsMap := utils.GetOptionsMap(options)
	if optsMap != nil {
		if v, ok := optsMap["allowKeywords"].(bool); ok {
			opts.AllowKeywords = v
		}
		if v, ok := optsMap["allowPattern"].(string); ok {
			opts.AllowPattern = v
		}
	}
	return opts
}

// Reserved word set mirrored from ESLint's core dot-notation rule
// (see eslint/lib/rules/utils/keywords.js). This is the ES3 reserved-word
// list, which the core rule uses for its keyword check regardless of
// ECMAScript version, because bracket notation on these names is the only
// ES3-compatible form.
var keywordSet = map[string]struct{}{
	"abstract": {}, "boolean": {}, "break": {}, "byte": {}, "case": {}, "catch": {},
	"char": {}, "class": {}, "const": {}, "continue": {}, "debugger": {}, "default": {},
	"delete": {}, "do": {}, "double": {}, "else": {}, "enum": {}, "export": {},
	"extends": {}, "false": {}, "final": {}, "finally": {}, "float": {}, "for": {},
	"function": {}, "goto": {}, "if": {}, "implements": {}, "import": {}, "in": {},
	"instanceof": {}, "int": {}, "interface": {}, "long": {}, "native": {}, "new": {},
	"null": {}, "package": {}, "private": {}, "protected": {}, "public": {}, "return": {},
	"short": {}, "static": {}, "super": {}, "switch": {}, "synchronized": {}, "this": {},
	"throw": {}, "throws": {}, "transient": {}, "true": {}, "try": {}, "typeof": {},
	"var": {}, "void": {}, "volatile": {}, "while": {}, "with": {},
}

func isKeyword(name string) bool {
	_, ok := keywordSet[name]
	return ok
}

// validIdentifierRE mirrors upstream's ASCII-only `/^[a-zA-Z_$][\w$]*$/u`.
var validIdentifierRE = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

func isValidIdentifier(name string) bool {
	return validIdentifierRE.MatchString(name)
}

// decimalIntegerRE mirrors upstream's DECIMAL_INTEGER_PATTERN from ast-utils.js:
// /^(?:0|0[0-7]*[89]\d*|[1-9](?:_?\d)*)$/u
var decimalIntegerRE = regexp.MustCompile(`^(?:0|0[0-7]*[89]\d*|[1-9](?:_?\d)*)$`)

// isDecimalIntegerLiteral reports whether node is a numeric literal whose raw
// source text is a plain decimal integer (as opposed to an octal-style legacy
// literal such as `0123`, or a hex/octal/binary literal such as `0x1`/`0b1`).
// Appending `.prop` directly after such a literal is ambiguous with the start
// of a fractional part, so the autofix must insert a separating space.
func isDecimalIntegerLiteral(sourceFile *ast.SourceFile, node *ast.Node) bool {
	if node.Kind != ast.KindNumericLiteral {
		return false
	}
	raw := utils.TrimmedNodeText(sourceFile, node)
	return decimalIntegerRE.MatchString(raw)
}

// identContinueRE matches a single identifier-continue character. Used to
// detect whether the property name we're about to splice in via dot notation
// would fuse with the token that immediately follows the original bracket
// access (e.g. `foo['bar']instanceof baz` -> `foo.bar instanceof baz`, not
// `foo.bar` fused directly into `instanceof` with no separating space).
var identContinueRE = regexp.MustCompile(`^[A-Za-z0-9_$]`)

func buildUseDotMessage(key string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useDot",
		Description: "[" + key + "] is better written in dot notation.",
		Data:        map[string]string{"key": key},
	}
}

func buildUseBracketsMessage(key string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useBrackets",
		Description: "." + key + " is a syntax error.",
		Data:        map[string]string{"key": key},
	}
}

// jsonQuote mirrors JS's JSON.stringify(str) for the message-data formatting
// of a string literal key (e.g. "b" -> `"b"`).
func jsonQuote(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return `"` + s + `"`
	}
	return string(b)
}

// literalKey returns the property key text (for reporting) and its "value"
// used for the dot-notation validity check, for the set of computed-key forms
// ESLint's core rule handles: string literals, no-substitution template
// literals (a[`foo`]), and the bare `null` / `true` / `false` keyword
// literals (`a[null]` / `a[true]` / `a[false]`). Numeric literals and any
// other computed expression are intentionally left unhandled — ok is false.
func literalKey(n *ast.Node) (value string, formatted string, ok bool) {
	if ast.IsStringLiteralLike(n) {
		text := n.Text()
		if n.Kind == ast.KindNoSubstitutionTemplateLiteral {
			return text, "`" + text + "`", true
		}
		return text, jsonQuote(text), true
	}
	switch n.Kind {
	case ast.KindNullKeyword:
		return "null", "null", true
	case ast.KindTrueKeyword:
		return "true", "true", true
	case ast.KindFalseKeyword:
		return "false", "false", true
	}
	return "", "", false
}

// DotNotationRule enforces dot notation whenever possible, matching ESLint
// core's `dot-notation` rule (eslint/lib/rules/dot-notation.js) 1:1. This is
// the plain, non-type-aware port; the TypeScript-enhanced variant with
// allow{Private,Protected}ClassPropertyAccess / allowIndexSignaturePropertyAccess
// is registered separately as "@typescript-eslint/dot-notation"
// (internal/plugins/typescript/rules/dot_notation).
var DotNotationRule = rule.Rule{
	Name:   "dot-notation",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		opts := parseOptions(options)

		// ECMAScript + Unicode flags mirror ESLint's `new RegExp(pattern, 'u')`
		// so user patterns using lookaround, backreferences, or `\p{...}` work
		// identically to the original rule (Go's standard `regexp` / RE2 does
		// not support those). Invalid regex patterns are silently ignored, as
		// upstream would throw at rule-construction time in that case, which
		// is out of scope here.
		var allowRE *regexp2.Regexp
		if opts.AllowPattern != "" {
			if re, err := regexp2.Compile(opts.AllowPattern, regexp2.ECMAScript|regexp2.Unicode); err == nil {
				allowRE = re
			}
		}

		sourceFile := ctx.SourceFile

		// checkComputedProperty reports and (when safe) autofixes a bracket
		// access whose key is a literal recognized by literalKey. keyNode is
		// the parenthesis-unwrapped key (e.g. `a[('foo')]` -> the `'foo'`
		// node), used for the diagnostic position so it matches ESLint's
		// ESTree-based position (which has no wrapper node for parens).
		checkComputedProperty := func(node *ast.Node, elem *ast.ElementAccessExpression, keyNode *ast.Node, value, formattedKey string) {
			if !isValidIdentifier(value) {
				return
			}
			if !opts.AllowKeywords && isKeyword(value) {
				return
			}
			if allowRE != nil {
				if matched, err := allowRE.MatchString(value); err != nil || matched {
					// Fail open on regex errors (e.g. a timeout if MatchTimeout
					// were ever set): skip reporting rather than risk a false
					// positive.
					return
				}
			}

			text := sourceFile.Text()
			nodeRange := utils.TrimNodeTextRange(sourceFile, node)
			exprRange := utils.TrimNodeTextRange(sourceFile, elem.Expression)

			// Locate the opening `[`; "?." (if present) and any accidental
			// whitespace live between the object and it.
			bracketStart := exprRange.End()
			for bracketStart < nodeRange.End() && text[bracketStart] != '[' {
				bracketStart++
			}
			bracketEnd := nodeRange.End() - 1
			for bracketEnd > bracketStart && text[bracketEnd] != ']' {
				bracketEnd--
			}

			// HasCommentsInRange only walks trivia contiguous from its range's
			// start position, stopping at the first real token - it can't see
			// past the key token itself. Anchor it once right after `[`
			// (catches a comment before the key, e.g. `foo[ /* c */ 'bar' ]`)
			// and once right after the key (catches a comment after it, e.g.
			// `foo[ 'bar' /* c */ ]`).
			keyEnd := utils.TrimNodeTextRange(sourceFile, elem.ArgumentExpression).End()
			hasComment := utils.HasCommentsInRange(sourceFile, core.NewTextRange(bracketStart+1, bracketEnd)) ||
				utils.HasCommentsInRange(sourceFile, core.NewTextRange(keyEnd, bracketEnd))
			if hasComment {
				// Report but don't fix: a comment lives inside the brackets.
				ctx.ReportNode(keyNode, buildUseDotMessage(formattedKey))
				return
			}

			// Unlike the dot->bracket direction, there is no operator token in
			// the source between the object and `[` for plain bracket access
			// (operatorLen 0) - only an optional chain has one ("?.", 2 chars).
			operatorLen := 0
			sep := "."
			if elem.QuestionDotToken != nil {
				operatorLen = 2 // "?."
				sep = "?."
			} else if isDecimalIntegerLiteral(sourceFile, elem.Expression) {
				sep = " ."
			}

			gapEnd := bracketStart - operatorLen
			whitespace := ""
			if gapEnd > exprRange.End() {
				whitespace = text[exprRange.End():gapEnd]
			}

			objectText := text[exprRange.Pos():exprRange.End()]
			replacement := objectText + whitespace + sep + value

			// Guard against the replacement identifier fusing with whatever
			// token immediately follows the closing bracket (no existing
			// whitespace between them), e.g. `foo['bar']instanceof baz`.
			nextCharIdx := bracketEnd + 1
			if nextCharIdx < len(text) && identContinueRE.MatchString(string(text[nextCharIdx])) {
				replacement += " "
			}

			ctx.ReportNodeWithFixes(keyNode, buildUseDotMessage(formattedKey), rule.RuleFixReplace(sourceFile, node, replacement))
		}

		listeners := rule.RuleListeners{}

		listeners[ast.KindElementAccessExpression] = func(node *ast.Node) {
			elem := node.AsElementAccessExpression()
			if elem == nil || elem.ArgumentExpression == nil {
				return
			}
			// Unwrap parentheses on the key so shapes like `a[('foo')]` are
			// still recognized as a literal key, and so the reported position
			// matches ESLint's ESTree-based position (no wrapper node for
			// parens there).
			keyNode := ast.SkipParentheses(elem.ArgumentExpression)
			value, formattedKey, ok := literalKey(keyNode)
			if !ok {
				return
			}
			checkComputedProperty(node, elem, keyNode, value, formattedKey)
		}

		// Handle dot -> bracket (PropertyAccessExpression) when keywords are
		// disallowed. Mirrors ESLint core: only when the property is an
		// Identifier whose name is in the ES3 reserved-word list.
		listeners[ast.KindPropertyAccessExpression] = func(node *ast.Node) {
			if opts.AllowKeywords {
				return
			}
			pae := node.AsPropertyAccessExpression()
			if pae == nil || pae.Name() == nil || pae.Expression == nil {
				return
			}
			if !ast.IsIdentifier(pae.Name()) {
				return
			}
			name := pae.Name().Text()
			if !isKeyword(name) {
				return
			}

			// `let[...]` parses as a destructuring variable declaration, not
			// a MemberExpression: suppress the autofix in that exact shape
			// (non-optional access on a bare `let` identifier), matching
			// upstream's fixer guard.
			isOptional := pae.QuestionDotToken != nil
			if !isOptional && ast.IsIdentifier(pae.Expression) && pae.Expression.Text() == "let" {
				ctx.ReportNode(pae.Name(), buildUseBracketsMessage(name))
				return
			}

			text := sourceFile.Text()
			nameRange := utils.TrimNodeTextRange(sourceFile, pae.Name())
			objRange := utils.TrimNodeTextRange(sourceFile, pae.Expression)

			// Locate the start of the access operator (either `?.` or `.`);
			// any whitespace between the object end and the operator belongs
			// to the replacement.
			accessStart := objRange.End()
			for accessStart < nameRange.Pos() && text[accessStart] != '?' && text[accessStart] != '.' {
				accessStart++
			}

			opLen := 1
			if text[accessStart] == '?' {
				opLen = 2
			}
			gapStart := accessStart + opLen
			if gapStart > nameRange.Pos() {
				gapStart = nameRange.Pos()
			}
			if utils.HasCommentsInRange(sourceFile, core.NewTextRange(gapStart, nameRange.Pos())) {
				// Report but don't fix: a comment lives between the operator
				// and the property name.
				ctx.ReportNode(pae.Name(), buildUseBracketsMessage(name))
				return
			}

			objectText := text[objRange.Pos():objRange.End()]
			preOp := text[objRange.End():accessStart] // whitespace before operator
			var replacement string
			if isOptional {
				replacement = objectText + preOp + "?.[\"" + name + "\"]"
			} else {
				replacement = objectText + preOp + "[\"" + name + "\"]"
			}
			ctx.ReportNodeWithFixes(pae.Name(), buildUseBracketsMessage(name), rule.RuleFixReplace(sourceFile, node, replacement))
		}

		return listeners
	},
}
