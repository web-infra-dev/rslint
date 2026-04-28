package dot_notation

import (
	"regexp"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Options mirrors @typescript-eslint/dot-notation options
type Options struct {
	AllowIndexSignaturePropertyAccess bool   `json:"allowIndexSignaturePropertyAccess"`
	AllowKeywords                     bool   `json:"allowKeywords"`
	AllowPattern                      string `json:"allowPattern"`
	AllowPrivateClassPropertyAccess   bool   `json:"allowPrivateClassPropertyAccess"`
	AllowProtectedClassPropertyAccess bool   `json:"allowProtectedClassPropertyAccess"`
}

func parseOptions(options any) Options {
	opts := Options{
		AllowKeywords:                     true,
		AllowIndexSignaturePropertyAccess: false,
	}

	if options == nil {
		return opts
	}

	// Parse options with dual-format support (handles both array and object formats)
	var optsMap map[string]interface{}
	var ok bool

	// Handle array format: [{ option: value }]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, ok = optArray[0].(map[string]interface{})
	} else {
		// Handle direct object format: { option: value }
		optsMap, ok = options.(map[string]interface{})
	}

	if ok {
		if v, ok := optsMap["allowIndexSignaturePropertyAccess"].(bool); ok {
			opts.AllowIndexSignaturePropertyAccess = v
		}
		if v, ok := optsMap["allowKeywords"].(bool); ok {
			opts.AllowKeywords = v
		}
		if v, ok := optsMap["allowPattern"].(string); ok {
			opts.AllowPattern = v
		}
		if v, ok := optsMap["allowPrivateClassPropertyAccess"].(bool); ok {
			opts.AllowPrivateClassPropertyAccess = v
		}
		if v, ok := optsMap["allowProtectedClassPropertyAccess"].(bool); ok {
			opts.AllowProtectedClassPropertyAccess = v
		}
	}
	return opts
}

func buildUseDotMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useDot",
		Description: "Use dot notation instead of bracket notation.",
	}
}

func buildUseBracketsMessage(key string) rule.RuleMessage {
	// Keep key for parity with ESLint message data (not used in printing now)
	_ = key
	return rule.RuleMessage{
		Id:          "useBrackets",
		Description: "Property is a keyword - use bracket notation.",
	}
}

// Reserved word set mirrored from ESLint's core dot-notation rule
// (see eslint/lib/rules/utils/keywords.js). This is the ES3 reserved-word
// list, which is what the core rule uses for its keyword check — regardless
// of ECMAScript version, because bracket notation on these names is the only
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

var identRE = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

func isValidIdentifier(name string) bool {
	return identRE.MatchString(name)
}

func isKeyword(name string) bool {
	_, ok := keywordSet[name]
	return ok
}

// hasStringLikeIndexSignature reports whether the type exposes an index
// signature whose key is string-like. Uses the type checker so that mapped
// types (including Record<K, V>) and template literal index keys are
// recognized — not just inline `[key: …]: …` declarations.
//
// Passes the type through unchanged: on union types the checker returns only
// index signatures present on *every* member, and on intersections it merges
// signatures from the parts — which matches typescript-eslint's
// `checker.getIndexInfosOfType(objectType)` behavior exactly.
func hasStringLikeIndexSignature(tc *checker.Checker, t *checker.Type) bool {
	if t == nil || tc == nil {
		return false
	}
	for _, info := range tc.GetIndexInfosOfType(t) {
		if info == nil || info.KeyType() == nil {
			continue
		}
		if info.KeyType().Flags()&checker.TypeFlagsStringLike != 0 {
			return true
		}
	}
	return false
}

// literalKey returns the property key string for the argument of a bracket
// access, matching the set ESLint's core rule handles: string literals,
// no-substitution template literals (`a[`foo`]`), and the bare `null` /
// `true` / `false` keyword literals (`a[null]` / `a[true]` / `a[false]`).
// Computed expressions and number literals are intentionally skipped.
func literalKey(n *ast.Node) (string, bool) {
	if ast.IsStringLiteralLike(n) {
		return n.Text(), true
	}
	switch n.Kind {
	case ast.KindNullKeyword:
		return "null", true
	case ast.KindTrueKeyword:
		return "true", true
	case ast.KindFalseKeyword:
		return "false", true
	}
	return "", false
}

// DotNotationRule enforces dot-notation when safe and allowed by options.
//
// KNOWN LIMITATION: The test infrastructure in /packages/rule-tester doesn't properly pass
// TypeScript compiler options from individual test cases. This means tests that rely on
// specific tsconfig settings (like noPropertyAccessFromIndexSignature) may not work correctly.
// The test runner always uses the same rslint.json config file for all test cases, which
// references a fixed tsconfig.json. Individual test cases can specify different tsconfig files
// via languageOptions.parserOptions.project, but these are ignored by the test runner.
// See: /packages/rule-tester/src/index.ts line 273 - the lint() call doesn't use per-test config.
var DotNotationRule = rule.CreateRule(rule.Rule{
	Name:             "dot-notation",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		// ECMAScript + Unicode flags mirror ESLint's `new RegExp(pattern, 'u')`
		// so user patterns using lookaround, backreferences, or `\p{...}` work
		// identically to the original rule (Go's standard `regexp` / RE2 does not
		// support those). Invalid regex patterns are silently ignored for parity.
		var allowRE *regexp2.Regexp
		if opts.AllowPattern != "" {
			if re, err := regexp2.Compile(opts.AllowPattern, regexp2.ECMAScript|regexp2.Unicode); err == nil {
				allowRE = re
			}
		}

		// Derive allowIndexSignaturePropertyAccess from tsconfig option as well (currently not used directly)
		if ctx.Program != nil {
			_ = ctx.Program.Options()
		}

		listeners := rule.RuleListeners{}

		// Compute allowIndexSignaturePropertyAccess once — respects both the
		// rule option and the `noPropertyAccessFromIndexSignature` tsconfig flag
		// (same derivation as typescript-eslint).
		allowIndexAccess := opts.AllowIndexSignaturePropertyAccess
		if ctx.Program != nil {
			if copts := ctx.Program.Options(); copts != nil && copts.NoPropertyAccessFromIndexSignature.IsTrue() {
				allowIndexAccess = true
			}
		}

		// Handle bracket → dot (ElementAccessExpression)
		listeners[ast.KindElementAccessExpression] = func(node *ast.Node) {
			elem := node.AsElementAccessExpression()
			if elem == nil || elem.ArgumentExpression == nil {
				return
			}

			// Unwrap parentheses on the key so shapes like `a[('foo')]`
			// are still recognized as a literal key.
			propName, ok := literalKey(ast.SkipParentheses(elem.ArgumentExpression))
			if !ok {
				return
			}

			// Option: allow pattern — matches the regex filter from the core
			// rule (JS uses the `u` flag, Go regexp is UTF-8 aware; simple
			// patterns behave identically, ES-only features like lookbehind
			// are not supported).
			if allowRE != nil {
				// Fail open: on regex errors (e.g. catastrophic-backtracking
				// timeouts if MatchTimeout is ever set) skip reporting rather
				// than risk a false positive.
				if matched, err := allowRE.MatchString(propName); err != nil || matched {
					return
				}
			}

			// Base-rule gating: only flag when value is a valid identifier
			// AND, when allowKeywords is off, it's not an ES3 reserved word.
			if !isValidIdentifier(propName) {
				return
			}
			if !opts.AllowKeywords && isKeyword(propName) {
				return
			}

			// typescript-eslint only performs the type-based skip when one of
			// the allow-* options is turned on — skip the checker round-trip
			// otherwise.
			if opts.AllowPrivateClassPropertyAccess ||
				opts.AllowProtectedClassPropertyAccess ||
				allowIndexAccess {

				objType := ctx.TypeChecker.GetTypeAtLocation(elem.Expression)
				nnType := ctx.TypeChecker.GetNonNullableType(objType)
				appType := checker.Checker_getApparentType(ctx.TypeChecker, nnType)

				// Prefer the symbol resolved by the checker at the property
				// node; fall back to scanning the object type's members.
				sym := ctx.TypeChecker.GetSymbolAtLocation(elem.ArgumentExpression)
				if sym == nil {
					sym = checker.Checker_getPropertyOfType(ctx.TypeChecker, appType, propName)
				}
				if sym == nil {
					for _, s := range checker.Checker_getPropertiesOfType(ctx.TypeChecker, appType) {
						if s != nil && s.Name == propName {
							sym = s
							break
						}
					}
				}

				if sym != nil {
					flags := checker.GetDeclarationModifierFlagsFromSymbol(sym)
					if opts.AllowPrivateClassPropertyAccess && (flags&ast.ModifierFlagsPrivate) != 0 {
						return
					}
					if opts.AllowProtectedClassPropertyAccess && (flags&ast.ModifierFlagsProtected) != 0 {
						return
					}
				} else if allowIndexAccess {
					// No named property symbol — allowed via string-like index signature.
					if hasStringLikeIndexSignature(ctx.TypeChecker, appType) {
						return
					}
				}
			}

			// Build the fix: replace `['prop']` with `.prop`, preserving any
			// whitespace between the object and the bracket. Skip the fix if a
			// comment lives inside the brackets (ESLint behavior).
			text := ctx.SourceFile.Text()
			nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			exprRange := utils.TrimNodeTextRange(ctx.SourceFile, elem.Expression)

			bracketStart := exprRange.End()
			for bracketStart < nodeRange.End() && text[bracketStart] != '[' {
				bracketStart++
			}
			bracketEnd := nodeRange.End() - 1
			for bracketEnd > bracketStart && text[bracketEnd] != ']' {
				bracketEnd--
			}

			insideBrackets := core.NewTextRange(bracketStart+1, bracketEnd)
			if utils.HasCommentsInRange(ctx.SourceFile, insideBrackets) {
				ctx.ReportNode(elem.ArgumentExpression, buildUseDotMessage())
				return
			}

			whitespace := ""
			if bracketStart > exprRange.End() {
				whitespace = text[exprRange.End():bracketStart]
			}
			objectText := text[exprRange.Pos():exprRange.End()]
			replacement := objectText + whitespace + "." + propName

			ctx.ReportNodeWithFixes(elem.ArgumentExpression, buildUseDotMessage(), rule.RuleFixReplace(ctx.SourceFile, node, replacement))
		}

		// Handle dot → bracket (PropertyAccessExpression) when keywords are disallowed.
		// Mirrors ESLint core behavior: only when property is an Identifier
		// whose name is in the ES3 reserved-word list.
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

			// `let[...]` parses as a destructuring variable declaration. Skip
			// the autofix in that exact shape (non-optional access on a bare
			// `let` identifier).
			isOptional := pae.QuestionDotToken != nil
			if !isOptional && ast.IsIdentifier(pae.Expression) && pae.Expression.Text() == "let" {
				ctx.ReportNode(pae.Name(), buildUseBracketsMessage(name))
				return
			}

			text := ctx.SourceFile.Text()
			nameRange := utils.TrimNodeTextRange(ctx.SourceFile, pae.Name())
			objRange := utils.TrimNodeTextRange(ctx.SourceFile, pae.Expression)

			// Locate the start of the access operator — either `?.` (optional
			// chain) or `.`. Any whitespace between the object end and the
			// operator belongs to the replacement.
			accessStart := objRange.End()
			for accessStart < nameRange.Pos() && text[accessStart] != '?' && text[accessStart] != '.' {
				accessStart++
			}

			// Suppress autofix if a comment lives between the operator and the
			// property name (ESLint parity). The operator is 1 byte for `.` or
			// 2 bytes for `?.`.
			opLen := 1
			if text[accessStart] == '?' {
				opLen = 2
			}
			gapStart := accessStart + opLen
			if gapStart > nameRange.Pos() {
				gapStart = nameRange.Pos()
			}
			if utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(gapStart, nameRange.Pos())) {
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
			ctx.ReportNodeWithFixes(pae.Name(), buildUseBracketsMessage(name), rule.RuleFixReplace(ctx.SourceFile, node, replacement))
		}

		return listeners
	},
})
