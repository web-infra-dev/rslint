package dot_notation

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"

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

// Reserved keywords that should trigger dot -> bracket when allowKeywords=false.
// Excludes identifiers that TS-ESLint treats as safe to access via dot even when allowKeywords=false
// (per tests): let, yield, eval, arguments, and literals true/false/null.
var keywordSet = map[string]struct{}{
	"break": {}, "case": {}, "catch": {}, "class": {}, "const": {}, "continue": {}, "debugger": {}, "default": {},
	"delete": {}, "do": {}, "else": {}, "export": {}, "extends": {}, "finally": {}, "for": {}, "function": {},
	"if": {}, "import": {}, "in": {}, "instanceof": {}, "new": {}, "return": {}, "super": {}, "switch": {},
	"this": {}, "throw": {}, "try": {}, "typeof": {}, "var": {}, "void": {}, "while": {}, "with": {},
	// intentionally not including: let, yield, eval, arguments, true, false, null
}

var identRE = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

func isValidIdentifier(name string) bool {
	return identRE.MatchString(name)
}

func isKeyword(name string) bool {
	_, ok := keywordSet[name]
	return ok
}

func typeHasIndexSignature(t *checker.Type) bool {
	if t == nil {
		return false
	}
	// Explore alias target declarations first if present
	if alias := checker.Type_alias(t); alias != nil && alias.Symbol() != nil {
		if decls := alias.Symbol().Declarations; len(decls) > 0 {
			for _, decl := range decls {
				if decl == nil {
					continue
				}
				switch decl.Kind {
				case ast.KindTypeAliasDeclaration:
					ta := decl.AsTypeAliasDeclaration()
					if ta != nil && ta.Type != nil && ta.Type.Kind == ast.KindTypeLiteral {
						tl := ta.Type.AsTypeLiteralNode()
						if tl != nil && tl.Members != nil {
							for _, m := range tl.Members.Nodes {
								if m != nil && m.Kind == ast.KindIndexSignature {
									return true
								}
							}
						}
					}
				case ast.KindInterfaceDeclaration:
					iface := decl.AsInterfaceDeclaration()
					if iface != nil && iface.Members != nil {
						for _, m := range iface.Members.Nodes {
							if m != nil && m.Kind == ast.KindIndexSignature {
								return true
							}
						}
					}
				case ast.KindTypeLiteral:
					tl := decl.AsTypeLiteralNode()
					if tl != nil && tl.Members != nil {
						for _, m := range tl.Members.Nodes {
							if m != nil && m.Kind == ast.KindIndexSignature {
								return true
							}
						}
					}
				}
			}
		}
	}

	sym := checker.Type_symbol(t)
	if sym == nil || len(sym.Declarations) == 0 {
		return false
	}
	for _, decl := range sym.Declarations {
		if decl == nil {
			continue
		}
		switch decl.Kind {
		case ast.KindInterfaceDeclaration:
			iface := decl.AsInterfaceDeclaration()
			if iface != nil && iface.Members != nil {
				for _, m := range iface.Members.Nodes {
					if m != nil && m.Kind == ast.KindIndexSignature {
						return true
					}
				}
			}
		case ast.KindTypeAliasDeclaration:
			alias := decl.AsTypeAliasDeclaration()
			if alias != nil && alias.Type != nil && alias.Type.Kind == ast.KindTypeLiteral {
				tl := alias.Type.AsTypeLiteralNode()
				if tl != nil && tl.Members != nil {
					for _, m := range tl.Members.Nodes {
						if m != nil && m.Kind == ast.KindIndexSignature {
							return true
						}
					}
				}
			}
		case ast.KindClassDeclaration:
			classDecl := decl.AsClassDeclaration()
			if classDecl != nil && classDecl.Members != nil {
				for _, m := range classDecl.Members.Nodes {
					if m != nil && m.Kind == ast.KindIndexSignature {
						return true
					}
				}
			}
		case ast.KindTypeLiteral:
			tl := decl.AsTypeLiteralNode()
			if tl != nil && tl.Members != nil {
				for _, m := range tl.Members.Nodes {
					if m != nil && m.Kind == ast.KindIndexSignature {
						return true
					}
				}
			}
		}
	}
	return false
}

// hasAnyIndexSignature walks unions/intersections to detect an index signature on any part
func hasAnyIndexSignature(t *checker.Type) bool {
	if t == nil {
		return false
	}
	if utils.IsUnionType(t) {
		for _, part := range t.Types() {
			if hasAnyIndexSignature(part) {
				return true
			}
		}
		return false
	}
	if utils.IsIntersectionType(t) {
		for _, part := range t.Types() {
			if hasAnyIndexSignature(part) {
				return true
			}
		}
		return false
	}
	return typeHasIndexSignature(t)
}

func getStringLiteralValue(srcFile *ast.SourceFile, n *ast.Node) (string, bool) {
	switch n.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		rng := utils.TrimNodeTextRange(srcFile, n)
		text := srcFile.Text()[rng.Pos():rng.End()]
		if len(text) >= 2 {
			quote := text[0]
			if (quote == '\'' || quote == '"' || quote == '`') && text[len(text)-1] == quote {
				return text[1 : len(text)-1], true
			}
		}
		// Fallback to raw text without outer quotes
		return strings.Trim(text, "'\"`"), true
	case ast.KindNullKeyword:
		return "null", true
	case ast.KindTrueKeyword:
		return "true", true
	case ast.KindFalseKeyword:
		return "false", true
	default:
		return "", false
	}
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
	Name: "dot-notation",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		var allowRE *regexp.Regexp
		if opts.AllowPattern != "" {
			if re, err := regexp.Compile(opts.AllowPattern); err == nil {
				allowRE = re
			}
			// Note: Invalid regex patterns are silently ignored for compatibility
		}

		// Derive allowIndexSignaturePropertyAccess from tsconfig option as well (currently not used directly)
		if ctx.Program != nil {
			_ = ctx.Program.Options()
		}

		listeners := rule.RuleListeners{}

		// Handle bracket → dot (ElementAccessExpression)
		listeners[ast.KindElementAccessExpression] = func(node *ast.Node) {
			elem := node.AsElementAccessExpression()
			if elem == nil || elem.ArgumentExpression == nil {
				return
			}

			// Only for simple string literals (and no-substitution templates)
			propName, ok := getStringLiteralValue(ctx.SourceFile, elem.ArgumentExpression)
			if !ok {
				return
			}

			// Option: allow pattern
			if allowRE != nil && allowRE.MatchString(propName) {
				return
			}

			// Option: allow keywords via bracket notation when allowKeywords is false.
			// When allowKeywords is false, 'null', 'true', and 'false' should be allowed in bracket notation.
			if !opts.AllowKeywords && (propName == "null" || propName == "true" || propName == "false") {
				return
			}

			// TS-specific relaxations
			objType := ctx.TypeChecker.GetTypeAtLocation(elem.Expression)
			nnType := ctx.TypeChecker.GetNonNullableType(objType)
			appType := checker.Checker_getApparentType(ctx.TypeChecker, nnType)

			// Try resolve symbol to check modifiers
			sym := checker.Checker_getPropertyOfType(ctx.TypeChecker, appType, propName)
			if sym == nil {
				for _, s := range checker.Checker_getPropertiesOfType(ctx.TypeChecker, appType) {
					if s != nil && s.Name == propName {
						sym = s
						break
					}
				}
			}

			// Check if we should allow based on modifiers
			if sym != nil {
				flags := checker.GetDeclarationModifierFlagsFromSymbol(sym)
				if (flags & ast.ModifierFlagsPrivate) != 0 {
					if opts.AllowPrivateClassPropertyAccess {
						return
					}
					// Continue to report error - private property with bracket notation
				} else if (flags & ast.ModifierFlagsProtected) != 0 {
					if opts.AllowProtectedClassPropertyAccess {
						return
					}
					// Continue to report error - protected property with bracket notation
				}
			} else {
				// Property not found as explicit declaration - check index signatures
				allowIndexAccess := opts.AllowIndexSignaturePropertyAccess
				if ctx.Program != nil {
					if copts := ctx.Program.Options(); copts != nil && copts.NoPropertyAccessFromIndexSignature.IsTrue() {
						allowIndexAccess = true
					}
				}

				// Check if the type has index signatures AND the property can only be accessed via index signature
				if hasAnyIndexSignature(appType) && allowIndexAccess {
					// When noPropertyAccessFromIndexSignature is true OR allowIndexSignaturePropertyAccess is true,
					// properties accessible only via index signature should use bracket notation
					return
				}
			}

			// If there is a declared property with this exact name, prefer dot; otherwise, fall back to index signature rules
			if isValidIdentifier(propName) && (opts.AllowKeywords || (!isKeyword(propName))) {
				// Build the fix: replace ['prop'] with .prop
				text := ctx.SourceFile.Text()
				nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				exprRange := utils.TrimNodeTextRange(ctx.SourceFile, elem.Expression)

				// Find the bracket position
				bracketStart := exprRange.End()
				for bracketStart < nodeRange.End() && text[bracketStart] != '[' {
					bracketStart++
				}

				// Check if there's whitespace (including newlines) before the bracket
				whitespace := ""
				if bracketStart > exprRange.End() {
					whitespace = text[exprRange.End():bracketStart]
				}

				// Build replacement preserving whitespace
				objectText := text[exprRange.Pos():exprRange.End()]
				replacement := objectText + whitespace + "." + propName

				// Report on the node with the fix
				ctx.ReportNodeWithFixes(node, buildUseDotMessage(), rule.RuleFixReplace(ctx.SourceFile, node, replacement))
			}
		}

		// Handle dot → bracket (PropertyAccessExpression) when keywords are disallowed
		listeners[ast.KindPropertyAccessExpression] = func(node *ast.Node) {
			if opts.AllowKeywords {
				return
			}
			pae := node.AsPropertyAccessExpression()
			if pae == nil || pae.Name() == nil || pae.Expression == nil {
				return
			}
			name := pae.Name().Text()
			if !isKeyword(name) && name != "true" && name != "false" && name != "null" {
				return
			}
			// Avoid autofix if comments present (heuristic)
			textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			if !utils.HasCommentsInRange(ctx.SourceFile, textRange) {
				text := ctx.SourceFile.Text()
				objRange := utils.TrimNodeTextRange(ctx.SourceFile, pae.Expression)

				// Find the dot position
				dotPos := objRange.End()
				for dotPos < textRange.End() && text[dotPos] != '.' {
					dotPos++
				}

				// Preserve whitespace before the dot
				whitespace := ""
				if dotPos > objRange.End() {
					whitespace = text[objRange.End():dotPos]
				}

				objectText := text[objRange.Pos():objRange.End()]
				replacement := objectText + whitespace + "[\"" + name + "\"]"
				ctx.ReportNodeWithFixes(node, buildUseBracketsMessage(name), rule.RuleFixReplace(ctx.SourceFile, node, replacement))
				return
			}
			ctx.ReportNode(node, buildUseBracketsMessage(name))
		}

		return listeners
	},
})
