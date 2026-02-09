package explicit_member_accessibility

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type accessibilityLevel string

const (
	accessibilityExplicit accessibilityLevel = "explicit"
	accessibilityNoPublic accessibilityLevel = "no-public"
	accessibilityOff      accessibilityLevel = "off"
	defaultAccessibility                     = accessibilityExplicit
)

type overrides struct {
	Accessors           accessibilityLevel
	Constructors        accessibilityLevel
	Methods             accessibilityLevel
	ParameterProperties accessibilityLevel
	Properties          accessibilityLevel
}

type options struct {
	Accessibility      accessibilityLevel
	Overrides          overrides
	IgnoredMethodNames map[string]struct{}
}

func parseAccessibility(value any) accessibilityLevel {
	if s, ok := value.(string); ok {
		switch accessibilityLevel(s) {
		case accessibilityExplicit, accessibilityNoPublic, accessibilityOff:
			return accessibilityLevel(s)
		}
	}
	return ""
}

func parseOptions(rawOpts any) options {
	opts := options{
		Accessibility: defaultAccessibility,
		Overrides:     overrides{},
	}
	if rawOpts == nil {
		return opts
	}

	var optsMap map[string]interface{}
	if arr, ok := rawOpts.([]interface{}); ok && len(arr) > 0 {
		optsMap, _ = arr[0].(map[string]interface{})
	} else {
		optsMap, _ = rawOpts.(map[string]interface{})
	}

	if optsMap == nil {
		return opts
	}

	if v := parseAccessibility(optsMap["accessibility"]); v != "" {
		opts.Accessibility = v
	}

	if ignored, ok := optsMap["ignoredMethodNames"]; ok {
		opts.IgnoredMethodNames = parseIgnoredMethodNames(ignored)
	}

	if overridesMap, ok := optsMap["overrides"].(map[string]interface{}); ok {
		if v := parseAccessibility(overridesMap["accessors"]); v != "" {
			opts.Overrides.Accessors = v
		}
		if v := parseAccessibility(overridesMap["constructors"]); v != "" {
			opts.Overrides.Constructors = v
		}
		if v := parseAccessibility(overridesMap["methods"]); v != "" {
			opts.Overrides.Methods = v
		}
		if v := parseAccessibility(overridesMap["parameterProperties"]); v != "" {
			opts.Overrides.ParameterProperties = v
		}
		if v := parseAccessibility(overridesMap["properties"]); v != "" {
			opts.Overrides.Properties = v
		}
	}

	return opts
}

func parseIgnoredMethodNames(value any) map[string]struct{} {
	ignored := map[string]struct{}{}
	switch v := value.(type) {
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				ignored[s] = struct{}{}
			}
		}
	case []string:
		for _, item := range v {
			ignored[item] = struct{}{}
		}
	}
	if len(ignored) == 0 {
		return nil
	}
	return ignored
}

func resolveOverride(value accessibilityLevel, base accessibilityLevel) accessibilityLevel {
	if value == "" {
		return base
	}
	return value
}

func messageMissingAccessibility(memberType string, name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingAccessibility",
		Description: fmt.Sprintf("Missing accessibility modifier on %s.", formatMemberDescription(memberType, name)),
	}
}

func messageUnwantedPublic(memberType string, name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unwantedPublicAccessibility",
		Description: fmt.Sprintf("Public accessibility modifier on %s.", formatMemberDescription(memberType, name)),
	}
}

func messageAddExplicitAccessibility(accessibility string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "addExplicitAccessibility",
		Description: fmt.Sprintf("Add '%s' accessibility modifier", accessibility),
	}
}

func formatMemberDescription(memberType string, name string) string {
	if name == "" {
		return memberType
	}
	return fmt.Sprintf("%s %s", memberType, name)
}

func getLastDecorator(node *ast.Node) *ast.Node {
	decorators := node.Decorators()
	if len(decorators) == 0 {
		return nil
	}
	return decorators[len(decorators)-1]
}

func getMemberStartPos(ctx rule.RuleContext, node *ast.Node) int {
	if decorator := getLastDecorator(node); decorator != nil {
		return scanner.SkipTrivia(ctx.SourceFile.Text(), decorator.End())
	}
	return utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
}

func getMemberHeadRange(ctx rule.RuleContext, node *ast.Node, nameNode *ast.Node) core.TextRange {
	start := getMemberStartPos(ctx, node)
	end := start
	if nameNode != nil {
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
		end = nameRange.End()
	} else {
		end = scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, start).End()
	}
	return core.NewTextRange(start, end)
}

func getParameterPropertyHeadRange(ctx rule.RuleContext, node *ast.Node) core.TextRange {
	start := getMemberStartPos(ctx, node)
	end := start
	param := node.AsParameterDeclaration()
	if param != nil && param.Name() != nil {
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, param.Name())
		end = nameRange.End()
	}
	return core.NewTextRange(start, end)
}

func ruleFixInsertAt(pos int, text string) rule.RuleFix {
	return rule.RuleFixReplaceRange(core.NewTextRange(pos, pos), text)
}

func buildAccessibilitySuggestions(ctx rule.RuleContext, node *ast.Node) []rule.RuleSuggestion {
	insertPos := getMemberStartPos(ctx, node)
	suggestions := make([]rule.RuleSuggestion, 0, 3)
	for _, accessibility := range []string{"public", "private", "protected"} {
		suggestions = append(suggestions, rule.RuleSuggestion{
			Message:  messageAddExplicitAccessibility(accessibility),
			FixesArr: []rule.RuleFix{ruleFixInsertAt(insertPos, accessibility+" ")},
		})
	}
	return suggestions
}

func getAccessibility(node *ast.Node) string {
	flags := ast.GetCombinedModifierFlags(node)
	if flags&ast.ModifierFlagsPublic != 0 {
		return "public"
	}
	if flags&ast.ModifierFlagsPrivate != 0 {
		return "private"
	}
	if flags&ast.ModifierFlagsProtected != 0 {
		return "protected"
	}
	return ""
}

func hasReadonly(node *ast.Node) bool {
	return ast.GetCombinedModifierFlags(node)&ast.ModifierFlagsReadonly != 0
}

func findPublicKeywordRange(ctx rule.RuleContext, node *ast.Node) (core.TextRange, core.TextRange, bool) {
	start := getMemberStartPos(ctx, node)
	s := scanner.GetScannerForSourceFile(ctx.SourceFile, start)
	text := ctx.SourceFile.Text()

	for s.TokenStart() < node.End() {
		if s.Token() == ast.KindPublicKeyword {
			keywordRange := core.NewTextRange(s.TokenStart(), s.TokenEnd())
			removeEnd := s.TokenEnd()

			i := s.TokenEnd()
			for i < len(text) && utils.IsStrWhiteSpace(rune(text[i])) {
				i++
			}
			if i+1 < len(text) && text[i] == '/' && (text[i+1] == '/' || text[i+1] == '*') {
				removeEnd = i
			} else {
				removeEnd = scanner.SkipTrivia(text, s.TokenEnd())
			}

			removeRange := core.NewTextRange(s.TokenStart(), removeEnd)
			return keywordRange, removeRange, true
		}
		s.Scan()
	}
	return core.TextRange{}, core.TextRange{}, false
}

func getMemberName(ctx rule.RuleContext, node *ast.Node, nameNode *ast.Node) string {
	if node.Kind == ast.KindConstructor {
		return "constructor"
	}
	if nameNode == nil {
		return ""
	}
	name, _ := utils.GetNameFromMember(ctx.SourceFile, nameNode)
	return name
}

func getParameterPropertyName(ctx rule.RuleContext, node *ast.Node) string {
	param := node.AsParameterDeclaration()
	if param == nil || param.Name() == nil {
		return ""
	}
	nameNode := param.Name()
	if nameNode.Kind == ast.KindIdentifier {
		return nameNode.AsIdentifier().Text
	}
	nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
	return strings.TrimSpace(ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()])
}

func isPrivateIdentifierName(nameNode *ast.Node) bool {
	return nameNode != nil && nameNode.Kind == ast.KindPrivateIdentifier
}

func isConstructorParameter(node *ast.Node) bool {
	if node.Parent != nil && node.Parent.Kind == ast.KindConstructor {
		return true
	}
	parentFunction := utils.GetParentFunctionNode(node)
	return parentFunction != nil && parentFunction.Kind == ast.KindConstructor
}

func isParameterProperty(node *ast.Node) bool {
	if !ast.IsParameter(node) {
		return false
	}
	if !isConstructorParameter(node) {
		return false
	}
	flags := ast.GetCombinedModifierFlags(node)
	return flags&(ast.ModifierFlagsPublic|ast.ModifierFlagsPrivate|ast.ModifierFlagsProtected|ast.ModifierFlagsReadonly) != 0
}

func checkMemberAccessibility(ctx rule.RuleContext, node *ast.Node, nameNode *ast.Node, check accessibilityLevel, memberType string, ignored map[string]struct{}) {
	if isPrivateIdentifierName(nameNode) {
		return
	}

	memberName := getMemberName(ctx, node, nameNode)
	if len(ignored) > 0 {
		if _, ok := ignored[memberName]; ok {
			return
		}
	}

	if check == accessibilityOff {
		return
	}

	accessibility := getAccessibility(node)
	if check == accessibilityNoPublic && accessibility == "public" {
		keywordRange, removeRange, ok := findPublicKeywordRange(ctx, node)
		if !ok {
			return
		}
		ctx.ReportRangeWithFixes(keywordRange, messageUnwantedPublic(memberType, memberName), rule.RuleFixRemoveRange(removeRange))
		return
	}

	if check == accessibilityExplicit && accessibility == "" {
		headRange := getMemberHeadRange(ctx, node, nameNode)
		ctx.ReportRangeWithSuggestions(headRange, messageMissingAccessibility(memberType, memberName), buildAccessibilitySuggestions(ctx, node)...)
	}
}

func checkPropertyAccessibility(ctx rule.RuleContext, node *ast.Node, check accessibilityLevel) {
	if check == accessibilityOff {
		return
	}
	property := node.AsPropertyDeclaration()
	if property == nil {
		return
	}
	nameNode := property.Name()
	if isPrivateIdentifierName(nameNode) {
		return
	}
	propertyName := getMemberName(ctx, node, nameNode)
	accessibility := getAccessibility(node)
	if check == accessibilityNoPublic && accessibility == "public" {
		keywordRange, removeRange, ok := findPublicKeywordRange(ctx, node)
		if !ok {
			return
		}
		ctx.ReportRangeWithFixes(keywordRange, messageUnwantedPublic("class property", propertyName), rule.RuleFixRemoveRange(removeRange))
		return
	}
	if check == accessibilityExplicit && accessibility == "" {
		headRange := getMemberHeadRange(ctx, node, nameNode)
		ctx.ReportRangeWithSuggestions(headRange, messageMissingAccessibility("class property", propertyName), buildAccessibilitySuggestions(ctx, node)...)
	}
}

func checkParameterPropertyAccessibility(ctx rule.RuleContext, node *ast.Node, check accessibilityLevel) {
	if check == accessibilityOff || !isParameterProperty(node) {
		return
	}

	name := getParameterPropertyName(ctx, node)
	accessibility := getAccessibility(node)
	if check == accessibilityNoPublic {
		if accessibility == "public" && hasReadonly(node) {
			keywordRange, removeRange, ok := findPublicKeywordRange(ctx, node)
			if !ok {
				return
			}
			ctx.ReportRangeWithFixes(keywordRange, messageUnwantedPublic("parameter property", name), rule.RuleFixRemoveRange(removeRange))
		}
		return
	}

	if check == accessibilityExplicit && accessibility == "" {
		headRange := getParameterPropertyHeadRange(ctx, node)
		ctx.ReportRangeWithSuggestions(headRange, messageMissingAccessibility("parameter property", name), buildAccessibilitySuggestions(ctx, node)...)
	}
}

var ExplicitMemberAccessibilityRule = rule.CreateRule(rule.Rule{
	Name: "explicit-member-accessibility",
	Run: func(ctx rule.RuleContext, rawOpts any) rule.RuleListeners {
		opts := parseOptions(rawOpts)
		if opts.IgnoredMethodNames == nil {
			opts.IgnoredMethodNames = map[string]struct{}{}
		}

		baseCheck := opts.Accessibility
		ctorCheck := resolveOverride(opts.Overrides.Constructors, baseCheck)
		accessorCheck := resolveOverride(opts.Overrides.Accessors, baseCheck)
		methodCheck := resolveOverride(opts.Overrides.Methods, baseCheck)
		propCheck := resolveOverride(opts.Overrides.Properties, baseCheck)
		paramPropCheck := resolveOverride(opts.Overrides.ParameterProperties, baseCheck)

		return rule.RuleListeners{
			ast.KindMethodDeclaration: func(node *ast.Node) {
				checkMemberAccessibility(ctx, node, node.AsMethodDeclaration().Name(), methodCheck, "method definition", opts.IgnoredMethodNames)
			},
			ast.KindConstructor: func(node *ast.Node) {
				checkMemberAccessibility(ctx, node, nil, ctorCheck, "method definition", opts.IgnoredMethodNames)
			},
			ast.KindGetAccessor: func(node *ast.Node) {
				checkMemberAccessibility(ctx, node, node.AsGetAccessorDeclaration().Name(), accessorCheck, "get property accessor", opts.IgnoredMethodNames)
			},
			ast.KindSetAccessor: func(node *ast.Node) {
				checkMemberAccessibility(ctx, node, node.AsSetAccessorDeclaration().Name(), accessorCheck, "set property accessor", opts.IgnoredMethodNames)
			},
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				checkPropertyAccessibility(ctx, node, propCheck)
			},
			ast.KindParameter: func(node *ast.Node) {
				checkParameterPropertyAccessibility(ctx, node, paramPropCheck)
			},
		}
	},
})
