package explicit_member_accessibility

import (
	"strings"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type accessibilityLevel string

const (
	levelExplicit accessibilityLevel = "explicit"
	levelNoPublic accessibilityLevel = "no-public"
	levelOff      accessibilityLevel = "off"
)

type overrides struct {
	accessors           accessibilityLevel
	constructors        accessibilityLevel
	methods             accessibilityLevel
	parameterProperties accessibilityLevel
	properties          accessibilityLevel
}

type options struct {
	accessibility      accessibilityLevel
	ignoredMethodNames map[string]bool
	overrides          overrides
}

func parseOptions(rawOpts any) options {
	opts := options{
		accessibility:      levelExplicit,
		ignoredMethodNames: map[string]bool{},
	}
	optsMap := utils.GetOptionsMap(rawOpts)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["accessibility"].(string); ok {
		opts.accessibility = accessibilityLevel(v)
	}
	if v, ok := optsMap["ignoredMethodNames"].([]interface{}); ok {
		for _, n := range v {
			if s, ok := n.(string); ok {
				opts.ignoredMethodNames[s] = true
			}
		}
	}
	if ov, ok := optsMap["overrides"].(map[string]interface{}); ok {
		if v, ok := ov["accessors"].(string); ok {
			opts.overrides.accessors = accessibilityLevel(v)
		}
		if v, ok := ov["constructors"].(string); ok {
			opts.overrides.constructors = accessibilityLevel(v)
		}
		if v, ok := ov["methods"].(string); ok {
			opts.overrides.methods = accessibilityLevel(v)
		}
		if v, ok := ov["parameterProperties"].(string); ok {
			opts.overrides.parameterProperties = accessibilityLevel(v)
		}
		if v, ok := ov["properties"].(string); ok {
			opts.overrides.properties = accessibilityLevel(v)
		}
	}
	return opts
}

func resolveCheck(base, override accessibilityLevel) accessibilityLevel {
	if override != "" {
		return override
	}
	return base
}

func buildMissingAccessibilityMessage(nodeType, name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingAccessibility",
		Description: "Missing accessibility modifier on " + nodeType + " " + name + ".",
	}
}

func buildUnwantedPublicAccessibilityMessage(nodeType, name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unwantedPublicAccessibility",
		Description: "Public accessibility modifier on " + nodeType + " " + name + ".",
	}
}

func buildAddExplicitAccessibilityMessage(accessibility string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "addExplicitAccessibility",
		Description: "Add '" + accessibility + "' accessibility modifier",
	}
}

// requiresQuoting reports whether the given name must be quoted to be a valid
// JavaScript property identifier (e.g. contains spaces or special chars).
// Mirrors the upstream `requiresQuoting` in @typescript-eslint/type-utils.
func requiresQuoting(s string) bool {
	if s == "" {
		return true
	}
	runes := []rune(s)
	if !isIdentifierStart(runes[0]) {
		return true
	}
	for _, r := range runes[1:] {
		if !isIdentifierPart(r) {
			return true
		}
	}
	return false
}

func isIdentifierStart(r rune) bool {
	return r == '$' || r == '_' || unicode.IsLetter(r)
}

func isIdentifierPart(r rune) bool {
	return isIdentifierStart(r) || unicode.IsDigit(r)
}

// getMemberName returns the diagnostic-friendly member name, matching upstream
// getNameFromMember in @typescript-eslint/eslint-plugin/util/misc.ts.
//
// tsgo's AST differs from ESLint's ESTree in two ways relevant here:
//   - ConstructorDeclaration has no Name() — ESLint synthesizes an Identifier
//     "constructor" as the key.
//   - A computed property name `[x]` is wrapped in a ComputedPropertyName node
//     in tsgo, whereas ESLint exposes the inner expression directly as
//     `member.key`. We unwrap to match.
func getMemberName(sf *ast.SourceFile, member *ast.Node) string {
	if member.Kind == ast.KindConstructor {
		return "constructor"
	}
	nameNode := member.Name()
	if nameNode == nil {
		return ""
	}
	if nameNode.Kind == ast.KindComputedPropertyName {
		expr := nameNode.AsComputedPropertyName().Expression
		if expr != nil {
			nameNode = expr
		}
	}
	switch nameNode.Kind {
	case ast.KindIdentifier:
		return nameNode.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return "#" + nameNode.AsPrivateIdentifier().Text
	case ast.KindStringLiteral:
		text := nameNode.AsStringLiteral().Text
		if requiresQuoting(text) {
			return `"` + text + `"`
		}
		return text
	case ast.KindNumericLiteral:
		text := nameNode.AsNumericLiteral().Text
		if requiresQuoting(text) {
			return `"` + text + `"`
		}
		return text
	case ast.KindBigIntLiteral:
		// Upstream's `${member.key.value}` for a BigInt literal coerces
		// to the decimal string without the `n` suffix. tsgo stores the
		// raw text including `n`, so strip it.
		text := nameNode.AsBigIntLiteral().Text
		text = strings.TrimSuffix(text, "n")
		if requiresQuoting(text) {
			return `"` + text + `"`
		}
		return text
	}
	r := utils.TrimNodeTextRange(sf, nameNode)
	return sf.Text()[r.Pos():r.End()]
}

// memberHeadStart is the position to use as the start of the report range
// when emitting `missingAccessibility`. It matches upstream getMemberHeadLoc:
// the position of the first non-decorator modifier, or — when only decorators
// are present (or no modifiers at all) — the position of the next token.
func memberHeadStart(sf *ast.SourceFile, node *ast.Node) int {
	mods := node.Modifiers()
	if mods != nil {
		var lastDecoratorEnd = -1
		for _, m := range mods.Nodes {
			if m.Kind == ast.KindDecorator {
				lastDecoratorEnd = m.End()
				continue
			}
			return utils.TrimNodeTextRange(sf, m).Pos()
		}
		if lastDecoratorEnd >= 0 {
			return scanner.SkipTrivia(sf.Text(), lastDecoratorEnd)
		}
	}
	return utils.TrimNodeTextRange(sf, node).Pos()
}

// methodOrPropertyHeadEnd is the position of the end of the head range
// (the end of the member key for methods/properties, or the end of the
// `constructor` keyword for constructors).
func methodOrPropertyHeadEnd(sf *ast.SourceFile, node *ast.Node) int {
	if node.Kind == ast.KindConstructor {
		// Compute position of the `constructor` keyword: it is the next token
		// after any modifiers. Then add the length of the keyword.
		text := sf.Text()
		ctorStart := utils.TrimNodeTextRange(sf, node).Pos()
		if mods := node.Modifiers(); mods != nil && len(mods.Nodes) > 0 {
			ctorStart = scanner.SkipTrivia(text, mods.Nodes[len(mods.Nodes)-1].End())
		}
		return ctorStart + len("constructor")
	}
	if name := node.Name(); name != nil {
		return name.End()
	}
	return node.End()
}

// findPublicKeyword locates the `public` modifier on a class member or parameter
// property and returns its keyword range, plus the range that an autofix should
// remove (keyword + trailing whitespace, stopping at the first comment).
func findPublicKeyword(sf *ast.SourceFile, node *ast.Node) (kwRange, removeRange core.TextRange, ok bool) {
	mods := node.Modifiers()
	if mods == nil {
		return core.TextRange{}, core.TextRange{}, false
	}
	for _, m := range mods.Nodes {
		if m.Kind != ast.KindPublicKeyword {
			continue
		}
		kwRange = utils.TrimNodeTextRange(sf, m)
		text := sf.Text()
		end := m.End()
		for end < len(text) {
			c := text[end]
			if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
				end++
				continue
			}
			break
		}
		removeRange = core.NewTextRange(kwRange.Pos(), end)
		return kwRange, removeRange, true
	}
	return core.TextRange{}, core.TextRange{}, false
}

// accessibilityOf returns the explicit accessibility keyword on a node ("public",
// "private", or "protected"), or "" when no accessibility modifier is present.
// Uses HasSyntacticModifier — same convention as the rest of the plugin
// (parameter_properties.go, member_ordering.go, etc.) — instead of walking the
// modifier list manually.
func accessibilityOf(node *ast.Node) string {
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsPublic) {
		return "public"
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsPrivate) {
		return "private"
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsProtected) {
		return "protected"
	}
	return ""
}

func hasReadonlyModifier(node *ast.Node) bool {
	return ast.HasSyntacticModifier(node, ast.ModifierFlagsReadonly)
}

// missingAccessibilitySuggestions returns the three "Add 'public/private/protected' "
// suggestion fixes for a member or parameter-property node.
func missingAccessibilitySuggestions(sf *ast.SourceFile, node *ast.Node) []rule.RuleSuggestion {
	insertPos := memberHeadStart(sf, node)
	insertRange := core.NewTextRange(insertPos, insertPos)
	build := func(accessibility string) rule.RuleSuggestion {
		return rule.RuleSuggestion{
			Message:  buildAddExplicitAccessibilityMessage(accessibility),
			FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(insertRange, accessibility+" ")},
		}
	}
	return []rule.RuleSuggestion{build("public"), build("private"), build("protected")}
}

func memberNodeType(kind ast.Kind) string {
	switch kind {
	case ast.KindGetAccessor:
		return "get property accessor"
	case ast.KindSetAccessor:
		return "set property accessor"
	default:
		return "method definition"
	}
}

// isClassMember reports whether the given node is a direct member of a class
// declaration / class expression. Several listener kinds (KindMethodDeclaration,
// KindGetAccessor, KindSetAccessor, KindPropertyDeclaration) also fire for
// object literal members in tsgo, but the upstream rule only inspects class
// members (its ESTree selectors are MethodDefinition / PropertyDefinition,
// neither of which exists for object literals or interfaces).
func isClassMember(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return true
	}
	return false
}

var ExplicitMemberAccessibilityRule = rule.CreateRule(rule.Rule{
	Name: "explicit-member-accessibility",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		sf := ctx.SourceFile

		baseCheck := opts.accessibility
		ctorCheck := resolveCheck(baseCheck, opts.overrides.constructors)
		accessorCheck := resolveCheck(baseCheck, opts.overrides.accessors)
		methodCheck := resolveCheck(baseCheck, opts.overrides.methods)
		propCheck := resolveCheck(baseCheck, opts.overrides.properties)
		paramPropCheck := resolveCheck(baseCheck, opts.overrides.parameterProperties)

		checkMethod := func(node *ast.Node) {
			if !isClassMember(node) {
				return
			}
			nameNode := node.Name()
			if nameNode != nil && nameNode.Kind == ast.KindPrivateIdentifier {
				return
			}

			var check accessibilityLevel
			nodeType := memberNodeType(node.Kind)
			switch node.Kind {
			case ast.KindMethodDeclaration:
				check = methodCheck
			case ast.KindConstructor:
				check = ctorCheck
			case ast.KindGetAccessor, ast.KindSetAccessor:
				check = accessorCheck
			default:
				check = methodCheck
			}

			methodName := getMemberName(sf, node)
			if check == levelOff || opts.ignoredMethodNames[methodName] {
				return
			}

			accessibility := accessibilityOf(node)

			if check == levelNoPublic && accessibility == "public" {
				kwRange, removeRange, ok := findPublicKeyword(sf, node)
				if !ok {
					return
				}
				ctx.ReportRangeWithFixes(
					kwRange,
					buildUnwantedPublicAccessibilityMessage(nodeType, methodName),
					rule.RuleFixRemoveRange(removeRange),
				)
				return
			}
			if check == levelExplicit && accessibility == "" {
				start := memberHeadStart(sf, node)
				end := methodOrPropertyHeadEnd(sf, node)
				ctx.ReportRangeWithSuggestions(
					core.NewTextRange(start, end),
					buildMissingAccessibilityMessage(nodeType, methodName),
					missingAccessibilitySuggestions(sf, node)...,
				)
			}
		}

		checkProperty := func(node *ast.Node) {
			if !isClassMember(node) {
				return
			}
			nameNode := node.Name()
			if nameNode != nil && nameNode.Kind == ast.KindPrivateIdentifier {
				return
			}

			nodeType := "class property"
			propertyName := getMemberName(sf, node)
			accessibility := accessibilityOf(node)

			if propCheck == levelNoPublic && accessibility == "public" {
				kwRange, removeRange, ok := findPublicKeyword(sf, node)
				if !ok {
					return
				}
				ctx.ReportRangeWithFixes(
					kwRange,
					buildUnwantedPublicAccessibilityMessage(nodeType, propertyName),
					rule.RuleFixRemoveRange(removeRange),
				)
				return
			}
			if propCheck == levelExplicit && accessibility == "" {
				start := memberHeadStart(sf, node)
				end := methodOrPropertyHeadEnd(sf, node)
				ctx.ReportRangeWithSuggestions(
					core.NewTextRange(start, end),
					buildMissingAccessibilityMessage(nodeType, propertyName),
					missingAccessibilitySuggestions(sf, node)...,
				)
			}
		}

		checkParameterProperty := func(node *ast.Node) {
			if node.Parent == nil || !ast.IsParameterPropertyDeclaration(node, node.Parent) {
				return
			}
			paramName := node.Name()
			if paramName == nil || paramName.Kind != ast.KindIdentifier {
				return
			}
			nodeType := "parameter property"
			nodeName := paramName.AsIdentifier().Text
			accessibility := accessibilityOf(node)

			switch paramPropCheck {
			case levelExplicit:
				if accessibility != "" {
					return
				}
				start := memberHeadStart(sf, node)
				end := paramName.End()
				ctx.ReportRangeWithSuggestions(
					core.NewTextRange(start, end),
					buildMissingAccessibilityMessage(nodeType, nodeName),
					missingAccessibilitySuggestions(sf, node)...,
				)
			case levelNoPublic:
				// Upstream only flags `public readonly` parameter properties under
				// no-public. A bare `public x` parameter would lose its parameter-
				// property semantics if `public` were removed, so it is left alone.
				if accessibility != "public" || !hasReadonlyModifier(node) {
					return
				}
				kwRange, removeRange, ok := findPublicKeyword(sf, node)
				if !ok {
					return
				}
				ctx.ReportRangeWithFixes(
					kwRange,
					buildUnwantedPublicAccessibilityMessage(nodeType, nodeName),
					rule.RuleFixRemoveRange(removeRange),
				)
			}
		}

		return rule.RuleListeners{
			ast.KindMethodDeclaration: checkMethod,
			ast.KindConstructor:       checkMethod,
			ast.KindGetAccessor:       checkMethod,
			ast.KindSetAccessor:       checkMethod,
			ast.KindPropertyDeclaration: checkProperty,
			ast.KindParameter:           checkParameterProperty,
		}
	},
})
