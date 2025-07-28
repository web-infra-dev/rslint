package no_restricted_types

import (
	"fmt"
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

// Options represents the rule configuration
type Options struct {
	Types map[string]BanConfig `json:"types"`
}

// BanConfig represents the configuration for a banned type
type BanConfig struct {
	Message string   `json:"message,omitempty"`
	FixWith string   `json:"fixWith,omitempty"`
	Suggest []string `json:"suggest,omitempty"`
}

// Remove all whitespace from a string
func removeSpaces(str string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(str, "")
}

// Stringify a node by getting its text and removing spaces
func stringifyNode(node *ast.Node, sourceFile *ast.SourceFile) string {
	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	text := string(sourceFile.Text()[nodeRange.Pos():nodeRange.End()])
	return removeSpaces(text)
}

// Get custom message from ban configuration
func getCustomMessage(bannedConfig interface{}) string {
	if bannedConfig == nil {
		return ""
	}

	switch v := bannedConfig.(type) {
	case bool:
		if v {
			return ""
		}
	case string:
		return " " + v
	case map[string]interface{}:
		if msg, ok := v["message"].(string); ok && msg != "" {
			return " " + msg
		}
		if v["message"] == nil {
			return ""
		}
	case BanConfig:
		if v.Message != "" {
			return " " + v.Message
		}
	}

	return ""
}

// Parse ban configuration from options
func parseBanConfig(value interface{}) (enabled bool, config BanConfig) {
	if value == nil {
		return false, BanConfig{}
	}

	switch v := value.(type) {
	case bool:
		return v, BanConfig{}
	case string:
		return true, BanConfig{Message: v}
	case map[string]interface{}:
		enabled = true
		config = BanConfig{}
		if msg, ok := v["message"].(string); ok {
			config.Message = msg
		}
		if fix, ok := v["fixWith"].(string); ok {
			config.FixWith = fix
		}
		if suggest, ok := v["suggest"].([]interface{}); ok {
			for _, s := range suggest {
				if str, ok := s.(string); ok {
					config.Suggest = append(config.Suggest, str)
				}
			}
		}
		return
	}

	return false, BanConfig{}
}

// Map of type keywords to their AST node kinds
var typeKeywords = map[string]ast.Kind{
	"bigint":    ast.KindBigIntKeyword,
	"boolean":   ast.KindBooleanKeyword,
	"never":     ast.KindNeverKeyword,
	"null":      ast.KindNullKeyword,
	"number":    ast.KindNumberKeyword,
	"object":    ast.KindObjectKeyword,
	"string":    ast.KindStringKeyword,
	"symbol":    ast.KindSymbolKeyword,
	"undefined": ast.KindUndefinedKeyword,
	"unknown":   ast.KindUnknownKeyword,
	"void":      ast.KindVoidKeyword,
}

var NoRestrictedTypesRule = rule.Rule{
	Name: "no-restricted-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := Options{
			Types: make(map[string]BanConfig),
		}

		// Parse options with dual-format support (handles both array and object formats)
		if options != nil {
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
				if types, ok := optsMap["types"].(map[string]interface{}); ok {
					// Build a map of normalized type names to their configurations
					for typeName, config := range types {
						normalizedName := removeSpaces(typeName)
						enabled, banConfig := parseBanConfig(config)
						if enabled {
							opts.Types[normalizedName] = banConfig
						}
					}
				}
			}
		}

		// Helper function to check and report banned types
		checkBannedTypes := func(typeNode *ast.Node, name string) {
			normalizedName := removeSpaces(name)
			banConfig, isBanned := opts.Types[normalizedName]
			if !isBanned {
				return
			}

			// Build the error message
			customMessage := getCustomMessage(banConfig)
			message := rule.RuleMessage{
				Id:          "bannedTypeMessage",
				Description: fmt.Sprintf("Don't use `%s` as a type.%s", name, customMessage),
			}

			// Handle fixes and suggestions
			var fixes []rule.RuleFix
			var suggestions []rule.RuleSuggestion

			if banConfig.FixWith != "" {
				fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, typeNode, banConfig.FixWith))
			}

			if len(banConfig.Suggest) > 0 {
				for _, replacement := range banConfig.Suggest {
					suggestion := rule.RuleSuggestion{
						Message: rule.RuleMessage{
							Id:          "bannedTypeReplacement",
							Description: fmt.Sprintf("Replace `%s` with `%s`.", name, replacement),
						},
						FixesArr: []rule.RuleFix{
							rule.RuleFixReplace(ctx.SourceFile, typeNode, replacement),
						},
					}
					suggestions = append(suggestions, suggestion)
				}
			}

			// Report the diagnostic
			if len(fixes) > 0 {
				ctx.ReportNodeWithFixes(typeNode, message, fixes...)
			} else if len(suggestions) > 0 {
				ctx.ReportNodeWithSuggestions(typeNode, message, suggestions...)
			} else {
				ctx.ReportNode(typeNode, message)
			}
		}

		listeners := rule.RuleListeners{}

		// Add listeners for keyword types
		for keyword, kind := range typeKeywords {
			if _, exists := opts.Types[keyword]; exists {
				listeners[kind] = func(node *ast.Node) {
					// Get the keyword from the node kind
					for k, v := range typeKeywords {
						if v == node.Kind {
							checkBannedTypes(node, k)
							break
						}
					}
				}
			}
		}

		// Check type references
		listeners[ast.KindTypeReference] = func(node *ast.Node) {
			typeRef := node.AsTypeReference()
			
			// First check the type name itself
			typeName := stringifyNode(typeRef.TypeName, ctx.SourceFile)
			
			// Check if just the type name is banned
			if _, exists := opts.Types[removeSpaces(typeName)]; exists {
				checkBannedTypes(typeRef.TypeName, typeName)
			}
			
			// If the type has type arguments, also check the full type reference
			if typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) > 0 {
				fullTypeName := stringifyNode(node, ctx.SourceFile)
				checkBannedTypes(node, fullTypeName)
			}
		}

		// Check empty tuples []
		listeners[ast.KindTupleType] = func(node *ast.Node) {
			tupleType := node.AsTupleTypeNode()
			if len(tupleType.Elements.Nodes) == 0 {
				checkBannedTypes(node, "[]")
			}
		}

		// Check empty type literals {}
		listeners[ast.KindTypeLiteral] = func(node *ast.Node) {
			typeLiteral := node.AsTypeLiteralNode()
			if len(typeLiteral.Members.Nodes) == 0 {
				checkBannedTypes(node, "{}")
			}
		}

		// Check class implements clauses
		listeners[ast.KindClassDeclaration] = func(node *ast.Node) {
			classDecl := node.AsClassDeclaration()
			if classDecl.HeritageClauses != nil {
				for _, heritageClause := range classDecl.HeritageClauses.Nodes {
					clause := heritageClause.AsHeritageClause()
					if clause.Token == ast.KindImplementsKeyword {
						for _, type_ := range clause.Types.Nodes {
							typeName := stringifyNode(type_, ctx.SourceFile)
							checkBannedTypes(type_, typeName)
						}
					}
				}
			}
		}

		// Check interface extends clauses
		listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
			interfaceDecl := node.AsInterfaceDeclaration()
			if interfaceDecl.HeritageClauses != nil {
				for _, heritageClause := range interfaceDecl.HeritageClauses.Nodes {
					clause := heritageClause.AsHeritageClause()
					if clause.Token == ast.KindExtendsKeyword {
						for _, type_ := range clause.Types.Nodes {
							typeName := stringifyNode(type_, ctx.SourceFile)
							checkBannedTypes(type_, typeName)
						}
					}
				}
			}
		}

		return listeners
	},
}