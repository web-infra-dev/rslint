package explicit_member_accessibility

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

type AccessibilityLevel string

const (
	AccessibilityExplicit  AccessibilityLevel = "explicit"
	AccessibilityNoPublic  AccessibilityLevel = "no-public"
	AccessibilityOff       AccessibilityLevel = "off"
)

type Config struct {
	Accessibility       AccessibilityLevel `json:"accessibility,omitempty"`
	IgnoredMethodNames  []string           `json:"ignoredMethodNames,omitempty"`
	Overrides           *Overrides         `json:"overrides,omitempty"`
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
	switch node.Kind {
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
	for _, mod := range modifiers.NodeList.Nodes {
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

func hasDecorators(node *ast.Node) bool {
	// Check if node has decorator modifiers
	return ast.GetCombinedModifierFlags(node)&ast.ModifierFlagsDecorator != 0
}

func findPublicKeywordRange(ctx rule.RuleContext, node *ast.Node) (core.TextRange, core.TextRange) {
	var modifiers *ast.ModifierList
	switch node.Kind {
	case ast.KindMethodDeclaration:
		modifiers = node.AsMethodDeclaration().Modifiers()
	case ast.KindPropertyDeclaration:
		modifiers = node.AsPropertyDeclaration().Modifiers()
	case ast.KindGetAccessor:
		modifiers = node.AsGetAccessorDeclaration().Modifiers()
	case ast.KindSetAccessor:
		modifiers = node.AsSetAccessorDeclaration().Modifiers()
	case ast.KindConstructor:
		modifiers = node.AsConstructorDeclaration().Modifiers()
	case ast.KindParameter:
		modifiers = node.AsParameterDeclaration().Modifiers()
	}

	if modifiers == nil {
		return core.NewTextRange(0, 0), core.NewTextRange(0, 0)
	}

	for i, mod := range modifiers.NodeList.Nodes {
		if mod.Kind == ast.KindPublicKeyword {
			keywordRange := core.NewTextRange(mod.Pos(), mod.End())
			
			// Calculate range to remove (including following whitespace)
			removeEnd := mod.End()
			if i+1 < len(modifiers.NodeList.Nodes) {
				removeEnd = modifiers.NodeList.Nodes[i+1].Pos()
			} else {
				// Find next token after public keyword
				text := string(ctx.SourceFile.Text())
				for removeEnd < len(text) && (text[removeEnd] == ' ' || text[removeEnd] == '\t') {
					removeEnd++
				}
			}
			
			removeRange := core.NewTextRange(mod.Pos(), removeEnd)
			return keywordRange, removeRange
		}
	}

	return core.NewTextRange(0, 0), core.NewTextRange(0, 0)
}

func getMemberName(node *ast.Node, ctx rule.RuleContext) string {
	var nameNode *ast.Node
	switch node.Kind {
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
		return fmt.Sprintf("%s property accessor", memberKind)
	default:
		if node.Kind == ast.KindPropertyDeclaration {
			return "class property"
		}
		return "method definition"
	}
}

// Removed getMemberHeadLoc and getParameterPropertyHeadLoc functions
// Now using ReportNode directly which handles positioning correctly

func isAbstract(node *ast.Node) bool {
	var modifiers *ast.ModifierList
	switch node.Kind {
	case ast.KindMethodDeclaration:
		modifiers = node.AsMethodDeclaration().Modifiers()
	case ast.KindPropertyDeclaration:
		modifiers = node.AsPropertyDeclaration().Modifiers()
	case ast.KindGetAccessor:
		modifiers = node.AsGetAccessorDeclaration().Modifiers()
	case ast.KindSetAccessor:
		modifiers = node.AsSetAccessorDeclaration().Modifiers()
	}

	if modifiers != nil {
		for _, mod := range modifiers.NodeList.Nodes {
			if mod.Kind == ast.KindAbstractKeyword {
				return true
			}
		}
	}
	return false
}

func isAccessorProperty(node *ast.Node) bool {
	if node.Kind != ast.KindPropertyDeclaration {
		return false
	}
	
	prop := node.AsPropertyDeclaration()
	if prop.Modifiers() != nil {
		for _, mod := range prop.Modifiers().NodeList.Nodes {
			if mod.Kind == ast.KindAccessorKeyword {
				return true
			}
		}
	}
	return false
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
				paramPropCheck = baseCheck
			}
		} else {
			ctorCheck = baseCheck
			accessorCheck = baseCheck
			methodCheck = baseCheck
			propCheck = baseCheck
			paramPropCheck = baseCheck
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
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unwantedPublicAccessibility",
					Description: fmt.Sprintf("Public accessibility modifier on %s %s.", nodeType, methodName),
				})
			} else if check == AccessibilityExplicit && accessibility == "" {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "missingAccessibility",
					Description: fmt.Sprintf("Missing accessibility modifier on %s %s.", nodeType, methodName),
				})
			}
		}

		checkPropertyAccessibilityModifier := func(node *ast.Node) {
			if isPrivateIdentifier(node) {
				return
			}

			nodeType := "class property"
			propertyName := getMemberName(node, ctx)
			accessibility := getAccessibilityModifier(node)

			if propCheck == AccessibilityNoPublic && accessibility == "public" {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unwantedPublicAccessibility",
					Description: fmt.Sprintf("Public accessibility modifier on %s %s.", nodeType, propertyName),
				})
			} else if propCheck == AccessibilityExplicit && accessibility == "" {
				ctx.ReportNode(node, rule.RuleMessage{
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
			for _, mod := range param.Modifiers().NodeList.Nodes {
				if mod.Kind == ast.KindReadonlyKeyword {
					hasReadonly = true
				} else if mod.Kind == ast.KindPublicKeyword ||
					mod.Kind == ast.KindPrivateKeyword ||
					mod.Kind == ast.KindProtectedKeyword {
					hasAccessibility = true
				}
			}

			// A parameter property must have readonly OR accessibility modifier
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

			switch paramPropCheck {
			case AccessibilityExplicit:
				if accessibility == "" {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "missingAccessibility",
						Description: fmt.Sprintf("Missing accessibility modifier on %s %s.", nodeType, nodeName),
					})
				}
			case AccessibilityNoPublic:
				if accessibility == "public" {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unwantedPublicAccessibility",
						Description: fmt.Sprintf("Public accessibility modifier on %s %s.", nodeType, nodeName),
					})
				}
			}
		}

		return rule.RuleListeners{
			ast.KindMethodDeclaration: checkMethodAccessibilityModifier,
			ast.KindConstructor: checkMethodAccessibilityModifier,
			ast.KindGetAccessor: checkMethodAccessibilityModifier,
			ast.KindSetAccessor: checkMethodAccessibilityModifier,
			ast.KindPropertyDeclaration: checkPropertyAccessibilityModifier,
			ast.KindParameter: checkParameterPropertyAccessibilityModifier,
		}
	},
}

func getMissingAccessibilitySuggestions(node *ast.Node, ctx rule.RuleContext) []rule.RuleSuggestion {
	suggestions := []rule.RuleSuggestion{}
	accessibilities := []string{"public", "private", "protected"}

	for _, accessibility := range accessibilities {
		insertPos := node.Pos()
		insertText := accessibility + " "

		// If node has decorators, insert after the last decorator
		if hasDecorators(node) {
			// For now, skip decorator handling as the API has changed
			// TODO: Update decorator handling when API is stabilized
		}

		// For abstract members, insert after "abstract" keyword
		if isAbstract(node) {
			var modifiers *ast.ModifierList
			switch node.Kind {
			case ast.KindMethodDeclaration:
				modifiers = node.AsMethodDeclaration().Modifiers()
			case ast.KindPropertyDeclaration:
				modifiers = node.AsPropertyDeclaration().Modifiers()
			}
			
			if modifiers != nil {
				for _, mod := range modifiers.NodeList.Nodes {
					if mod.Kind == ast.KindAbstractKeyword {
						insertPos = mod.Pos()
						insertText = accessibility + " abstract "
						break
					}
				}
			}
		}

		// For accessor properties, insert before "accessor" keyword
		if isAccessorProperty(node) {
			prop := node.AsPropertyDeclaration()
			if prop.Modifiers() != nil {
				for _, mod := range prop.Modifiers().NodeList.Nodes {
					if mod.Kind == ast.KindAccessorKeyword {
						insertPos = mod.Pos()
						break
					}
				}
			}
		}

		suggestions = append(suggestions, rule.RuleSuggestion{
			Message: rule.RuleMessage{
				Id:          "addExplicitAccessibility",
				Description: fmt.Sprintf("Add '%s' accessibility modifier", accessibility),
			},
			FixesArr: []rule.RuleFix{
				{
					Range: core.NewTextRange(insertPos, insertPos),
					Text:  insertText,
				},
			},
		})
	}

	return suggestions
}