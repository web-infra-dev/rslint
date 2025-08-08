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

type AccessibilityLevel string

const (
	AccessibilityExplicit AccessibilityLevel = "explicit"
	AccessibilityNoPublic AccessibilityLevel = "no-public"
	AccessibilityOff      AccessibilityLevel = "off"
)

type Config struct {
	Accessibility      AccessibilityLevel `json:"accessibility,omitempty"`
	IgnoredMethodNames []string           `json:"ignoredMethodNames,omitempty"`
	Overrides          *Overrides         `json:"overrides,omitempty"`
}

type Overrides struct {
	Accessors           AccessibilityLevel `json:"accessors,omitempty"`
	Constructors        AccessibilityLevel `json:"constructors,omitempty"`
	Methods             AccessibilityLevel `json:"methods,omitempty"`
	ParameterProperties AccessibilityLevel `json:"parameterProperties,omitempty"`
	Properties          AccessibilityLevel `json:"properties,omitempty"`
}

func parseOptions(options any) Config {
	config := Config{
		Accessibility: AccessibilityExplicit,
	}

	if options == nil {
		return config
	}

	// Handle both array format and direct object format
	var optsMap map[string]interface{}
	if optsArray, ok := options.([]interface{}); ok && len(optsArray) > 0 {
		if opts, ok := optsArray[0].(map[string]interface{}); ok {
			optsMap = opts
		}
	} else if opts, ok := options.(map[string]interface{}); ok {
		optsMap = opts
	}

	if optsMap != nil {
		if accessibility, ok := optsMap["accessibility"].(string); ok {
			config.Accessibility = AccessibilityLevel(accessibility)
		}

		if ignoredMethodNames, ok := optsMap["ignoredMethodNames"].([]interface{}); ok {
			for _, name := range ignoredMethodNames {
				if strName, ok := name.(string); ok {
					config.IgnoredMethodNames = append(config.IgnoredMethodNames, strName)
				}
			}
		}

		if overrides, ok := optsMap["overrides"].(map[string]interface{}); ok {
			config.Overrides = &Overrides{}
			if accessors, ok := overrides["accessors"].(string); ok {
				config.Overrides.Accessors = AccessibilityLevel(accessors)
			}
			if constructors, ok := overrides["constructors"].(string); ok {
				config.Overrides.Constructors = AccessibilityLevel(constructors)
			}
			if methods, ok := overrides["methods"].(string); ok {
				config.Overrides.Methods = AccessibilityLevel(methods)
			}
			if parameterProperties, ok := overrides["parameterProperties"].(string); ok {
				config.Overrides.ParameterProperties = AccessibilityLevel(parameterProperties)
			}
			if properties, ok := overrides["properties"].(string); ok {
				config.Overrides.Properties = AccessibilityLevel(properties)
			}
		}
	}

	return config
}

func getAccessibilityModifier(node *ast.Node) string {
	switch kind := node.Kind; kind {
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		return getModifierText(method.Modifiers())
	case ast.KindPropertyDeclaration:
		prop := node.AsPropertyDeclaration()
		return getModifierText(prop.Modifiers())
	case ast.KindGetAccessor:
		getter := node.AsGetAccessorDeclaration()
		return getModifierText(getter.Modifiers())
	case ast.KindSetAccessor:
		setter := node.AsSetAccessorDeclaration()
		return getModifierText(setter.Modifiers())
	case ast.KindConstructor:
		ctor := node.AsConstructorDeclaration()
		return getModifierText(ctor.Modifiers())
	case ast.KindParameter:
		// For parameter properties
		param := node.AsParameterDeclaration()
		return getModifierText(param.Modifiers())
	}
	return ""
}

func getModifierText(modifiers *ast.ModifierList) string {
	if modifiers == nil {
		return ""
	}
	for _, mod := range modifiers.Nodes {
		switch mod.Kind {
		case ast.KindPublicKeyword:
			return "public"
		case ast.KindPrivateKeyword:
			return "private"
		case ast.KindProtectedKeyword:
			return "protected"
		}
	}
	return ""
}

func getMemberName(node *ast.Node, ctx rule.RuleContext) string {
	var nameNode *ast.Node
	switch kind := node.Kind; kind {
	case ast.KindMethodDeclaration:
		nameNode = node.AsMethodDeclaration().Name()
	case ast.KindPropertyDeclaration:
		nameNode = node.AsPropertyDeclaration().Name()
	case ast.KindGetAccessor:
		nameNode = node.AsGetAccessorDeclaration().Name()
	case ast.KindSetAccessor:
		nameNode = node.AsSetAccessorDeclaration().Name()
	case ast.KindConstructor:
		return "constructor"
	default:
		nameNode = node
	}

	if nameNode == nil {
		return ""
	}

	name, _ := utils.GetNameFromMember(ctx.SourceFile, nameNode)
	return name
}

func isPrivateIdentifier(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		return method.Name() != nil && method.Name().Kind == ast.KindPrivateIdentifier
	case ast.KindPropertyDeclaration:
		prop := node.AsPropertyDeclaration()
		return prop.Name() != nil && prop.Name().Kind == ast.KindPrivateIdentifier
	case ast.KindGetAccessor:
		getter := node.AsGetAccessorDeclaration()
		return getter.Name() != nil && getter.Name().Kind == ast.KindPrivateIdentifier
	case ast.KindSetAccessor:
		setter := node.AsSetAccessorDeclaration()
		return setter.Name() != nil && setter.Name().Kind == ast.KindPrivateIdentifier
	}
	return false
}

func getMemberKind(node *ast.Node) string {
	switch node.Kind {
	case ast.KindMethodDeclaration:
		return "method"
	case ast.KindConstructor:
		return "constructor"
	case ast.KindGetAccessor:
		return "get"
	case ast.KindSetAccessor:
		return "set"
	}
	return ""
}

func getNodeType(node *ast.Node, memberKind string) string {
	switch memberKind {
	case "constructor":
		return "constructor"
	case "get", "set":
		return memberKind + " property accessor"
	default:
		if node.Kind == ast.KindPropertyDeclaration {
			return "class property"
		}
		return "method definition"
	}
}

// Removed getMemberHeadLoc and getParameterPropertyHeadLoc functions
// Now using ReportNode directly which handles positioning correctly

func getMissingAccessibilityRange(ctx rule.RuleContext, node *ast.Node) core.TextRange {
	// Default to node's name range when available
	findAccessorKeywordStart := func(nameRange core.TextRange) int {
		// Search backwards from name for 'get' or 'set' keyword within the declaration span
		text := ctx.SourceFile.Text()
		startBound := node.Pos()
		endBound := nameRange.Pos()
		if startBound < 0 || endBound > len(text) || startBound >= endBound {
			return nameRange.Pos()
		}
		snippet := text[startBound:endBound]
		// look for last occurrence to get the actual keyword near the name
		idxGet := strings.LastIndex(snippet, "get")
		idxSet := strings.LastIndex(snippet, "set")
		idx := -1
		kw := ""
		if idxGet > idxSet {
			idx = idxGet
			kw = "get"
		} else {
			idx = idxSet
			kw = "set"
		}
		if idx >= 0 {
			// ensure simple word boundary (whitespace or start before; whitespace/paren after)
			abs := startBound + idx
			beforeOk := abs == startBound || (abs > 0 && (text[abs-1] == ' ' || text[abs-1] == '\t' || text[abs-1] == '\n'))
			afterPos := abs + len(kw)
			afterOk := afterPos < len(text) && (text[afterPos] == ' ' || text[afterPos] == '\t' || text[afterPos] == '\n' || text[afterPos] == '(')
			if beforeOk && afterOk {
				return abs
			}
		}
		return nameRange.Pos()
	}
	switch node.Kind {
	case ast.KindConstructor:
		// Highlight the 'constructor' keyword only
		return scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())
	case ast.KindMethodDeclaration:
		m := node.AsMethodDeclaration()
		nameNode := m.Name()
		if nameNode != nil {
			nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
			start := nameRange.Pos()
			// If abstract, start from 'abstract'
			if m.Modifiers() != nil {
				for _, mod := range m.Modifiers().Nodes {
					if mod.Kind == ast.KindAbstractKeyword {
						start = mod.Pos()
						break
					}
				}
			}
			nameText, _ := utils.GetNameFromMember(ctx.SourceFile, nameNode)
			end := nameRange.Pos() + len(nameText)
			return core.NewTextRange(start, end)
		}
		return utils.TrimNodeTextRange(ctx.SourceFile, node)
	case ast.KindGetAccessor:
		g := node.AsGetAccessorDeclaration()
		nameNode := g.Name()
		if nameNode != nil {
			nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
			// Start at the 'get' keyword token by scanning between node start and name
			start := findAccessorKeywordStart(nameRange)
			nameText, _ := utils.GetNameFromMember(ctx.SourceFile, nameNode)
			end := nameRange.Pos() + len(nameText)
			return core.NewTextRange(start, end)
		}
		return utils.TrimNodeTextRange(ctx.SourceFile, node)
	case ast.KindSetAccessor:
		s := node.AsSetAccessorDeclaration()
		nameNode := s.Name()
		if nameNode != nil {
			nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
			// Start at the 'set' keyword token by scanning between node start and name
			start := findAccessorKeywordStart(nameRange)
			nameText, _ := utils.GetNameFromMember(ctx.SourceFile, nameNode)
			end := nameRange.Pos() + len(nameText)
			return core.NewTextRange(start, end)
		}
		return utils.TrimNodeTextRange(ctx.SourceFile, node)
	case ast.KindPropertyDeclaration:
		p := node.AsPropertyDeclaration()
		nameNode := p.Name()
		if nameNode != nil {
			nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
			start := nameRange.Pos()
			if p.Modifiers() != nil {
				// Prefer abstract start if present
				for _, mod := range p.Modifiers().Nodes {
					if mod.Kind == ast.KindAbstractKeyword {
						start = mod.Pos()
						break
					}
				}
				// Otherwise, if accessor keyword present, start there to include `accessor foo`
				if start == nameRange.Pos() {
					for _, mod := range p.Modifiers().Nodes {
						if mod.Kind == ast.KindAccessorKeyword {
							start = mod.Pos()
							break
						}
					}
				}
			}
			nameText, _ := utils.GetNameFromMember(ctx.SourceFile, nameNode)
			end := nameRange.Pos() + len(nameText)
			return core.NewTextRange(start, end)
		}
		return utils.TrimNodeTextRange(ctx.SourceFile, node)
	}
	// Fallback
	return utils.TrimNodeTextRange(ctx.SourceFile, node)
}

var ExplicitMemberAccessibilityRule = rule.Rule{

	Name: "explicit-member-accessibility",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		config := parseOptions(options)

		baseCheck := config.Accessibility
		overrides := config.Overrides

		var ctorCheck, accessorCheck, methodCheck, propCheck, paramPropCheck AccessibilityLevel

		if overrides != nil {
			if overrides.Constructors != "" {
				ctorCheck = overrides.Constructors
			} else {
				ctorCheck = baseCheck
			}
			if overrides.Accessors != "" {
				accessorCheck = overrides.Accessors
			} else {
				accessorCheck = baseCheck
			}
			if overrides.Methods != "" {
				methodCheck = overrides.Methods
			} else {
				methodCheck = baseCheck
			}
			if overrides.Properties != "" {
				propCheck = overrides.Properties
			} else {
				propCheck = baseCheck
			}
			if overrides.ParameterProperties != "" {
				paramPropCheck = overrides.ParameterProperties
			} else {
				// Parameter properties are not checked unless explicitly configured
				paramPropCheck = AccessibilityOff
			}
		} else {
			ctorCheck = baseCheck
			accessorCheck = baseCheck
			methodCheck = baseCheck
			propCheck = baseCheck
			// Parameter properties inherit baseCheck only when it's 'explicit'
			if baseCheck == AccessibilityExplicit {
				paramPropCheck = baseCheck
			} else {
				paramPropCheck = AccessibilityOff
			}
		}

		ignoredMethodNames := make(map[string]bool)
		for _, name := range config.IgnoredMethodNames {
			ignoredMethodNames[name] = true
		}

        checkMethodAccessibilityModifier := func(node *ast.Node) {
			if isPrivateIdentifier(node) {
				return
			}

			memberKind := getMemberKind(node)
			nodeType := getNodeType(node, memberKind)
			check := baseCheck

			switch memberKind {
			case "method":
				check = methodCheck
			case "constructor":
				check = ctorCheck
			case "get", "set":
				check = accessorCheck
			}

			methodName := getMemberName(node, ctx)

			if check == AccessibilityOff || ignoredMethodNames[methodName] {
				return
			}

			accessibility := getAccessibilityModifier(node)

            if check == AccessibilityNoPublic && accessibility == "public" {
				// Find and report on the public keyword specifically, and provide fix
				var modifiers *ast.ModifierList
				switch kind := node.Kind; kind {
				case ast.KindMethodDeclaration:
					modifiers = node.AsMethodDeclaration().Modifiers()
				case ast.KindConstructor:
					modifiers = node.AsConstructorDeclaration().Modifiers()
				case ast.KindGetAccessor:
					modifiers = node.AsGetAccessorDeclaration().Modifiers()
				case ast.KindSetAccessor:
					modifiers = node.AsSetAccessorDeclaration().Modifiers()
				}

				if modifiers != nil {
					for _, mod := range modifiers.Nodes {
						if mod.Kind == ast.KindPublicKeyword {
							message := rule.RuleMessage{
								Id:          "unwantedPublicAccessibility",
								Description: fmt.Sprintf("Public accessibility modifier on %s %s.", nodeType, methodName),
							}
							ctx.ReportNode(mod, message)
							return
						}
					}
				}
            } else if check == AccessibilityExplicit && accessibility == "" {
				// Report precisely on the member name (or keyword for constructors/abstract)
				r := getMissingAccessibilityRange(ctx, node)
				ctx.ReportRange(r, rule.RuleMessage{
					Id:          "missingAccessibility",
					Description: fmt.Sprintf("Missing accessibility modifier on %s %s.", nodeType, methodName),
				})
			}
		}

		checkPropertyAccessibilityModifier := func(node *ast.Node) {
			if isPrivateIdentifier(node) {
				return
			}

			if propCheck == AccessibilityOff {
				return
			}

			nodeType := "class property"
			propertyName := getMemberName(node, ctx)
			accessibility := getAccessibilityModifier(node)

			if propCheck == AccessibilityNoPublic && accessibility == "public" {
				// Find and report on the public keyword specifically, and provide fix
				prop := node.AsPropertyDeclaration()
				if prop.Modifiers() != nil {
					for _, mod := range prop.Modifiers().Nodes {
						if mod.Kind == ast.KindPublicKeyword {
							message := rule.RuleMessage{
								Id:          "unwantedPublicAccessibility",
								Description: fmt.Sprintf("Public accessibility modifier on %s %s.", nodeType, propertyName),
							}
							ctx.ReportNode(mod, message)
							return
						}
					}
				}
			} else if propCheck == AccessibilityExplicit && accessibility == "" {
				// Report precisely on the property name or include accessor/abstract keyword when present
				r := getMissingAccessibilityRange(ctx, node)
				ctx.ReportRange(r, rule.RuleMessage{
					Id:          "missingAccessibility",
					Description: fmt.Sprintf("Missing accessibility modifier on %s %s.", nodeType, propertyName),
				})
			}
		}

        checkParameterPropertyAccessibilityModifier := func(node *ast.Node) {
			if node.Kind != ast.KindParameter {
				return
			}

			param := node.AsParameterDeclaration()

			// Check if it's a parameter property (has modifiers)
			if param.Modifiers() == nil {
				return
			}

			// Check if it has readonly or accessibility modifiers
			hasReadonly := false
			hasAccessibility := false
			var readonlyNode *ast.Node
			for _, mod := range param.Modifiers().Nodes {
				switch kind := mod.Kind; kind {
				case ast.KindReadonlyKeyword:
					hasReadonly = true
					readonlyNode = mod
				case ast.KindPublicKeyword, ast.KindPrivateKeyword, ast.KindProtectedKeyword:
					hasAccessibility = true
				}
			}

            // Consider only parameters that are parameter properties (have readonly or accessibility)
			if !hasReadonly && !hasAccessibility {
				return
			}

			// Must be an identifier or assignment pattern
			name := param.Name()
			if name == nil ||
				(name.Kind != ast.KindIdentifier &&
					name.Kind != ast.KindObjectBindingPattern &&
					name.Kind != ast.KindArrayBindingPattern) {
				return
			}

			nodeType := "parameter property"
			var nodeName string
			if name.Kind == ast.KindIdentifier {
				nodeName = name.AsIdentifier().Text
			} else {
				// For destructured parameters, use a placeholder name
				nodeName = "[destructured]"
			}

			accessibility := getAccessibilityModifier(node)

			if paramPropCheck == AccessibilityOff {
				return
			}

            // Emit at most one diagnostic per parameter property, matching TS-ESLint tests
			if paramPropCheck == AccessibilityExplicit && accessibility == "" {
				var reportRange core.TextRange
				if hasReadonly && readonlyNode != nil {
					reportRange = core.NewTextRange(readonlyNode.Pos(), name.End())
				} else {
					reportRange = core.NewTextRange(node.Pos(), name.End())
				}
				ctx.ReportRange(reportRange, rule.RuleMessage{
					Id:          "missingAccessibility",
					Description: fmt.Sprintf("Missing accessibility modifier on %s %s.", nodeType, nodeName),
				})
				return
			}

			if paramPropCheck == AccessibilityNoPublic && accessibility == "public" {
				if param.Modifiers() != nil {
					for _, mod := range param.Modifiers().Nodes {
						if mod.Kind == ast.KindPublicKeyword {
							message := rule.RuleMessage{
								Id:          "unwantedPublicAccessibility",
								Description: fmt.Sprintf("Public accessibility modifier on %s %s.", nodeType, nodeName),
							}
							ctx.ReportNode(mod, message)
							return
						}
					}
				}
				return
			}
		}

		return rule.RuleListeners{
			ast.KindMethodDeclaration:   checkMethodAccessibilityModifier,
			ast.KindConstructor:         checkMethodAccessibilityModifier,
			ast.KindGetAccessor:         checkMethodAccessibilityModifier,
			ast.KindSetAccessor:         checkMethodAccessibilityModifier,
			ast.KindPropertyDeclaration: checkPropertyAccessibilityModifier,
			ast.KindParameter:           checkParameterPropertyAccessibilityModifier,
		}
	},
}
