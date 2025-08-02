package no_empty_object_type

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type NoEmptyObjectTypeOptions struct {
	AllowInterfaces  string `json:"allowInterfaces,omitempty"`
	AllowObjectTypes string `json:"allowObjectTypes,omitempty"`
	AllowWithName    string `json:"allowWithName,omitempty"`
}

func noEmptyMessage(emptyType string) string {
	return strings.Join([]string{
		fmt.Sprintf("%s allows any non-nullish value, including literals like `0` and `\"\"`.", emptyType),
		"- If that's what you want, disable this lint rule with an inline comment or configure the '{{ option }}' rule option.",
		"- If you want a type meaning \"any object\", you probably want `object` instead.",
		"- If you want a type meaning \"any value\", you probably want `unknown` instead.",
	}, "\n")
}

var NoEmptyObjectTypeRule = rule.Rule{
	Name: "no-empty-object-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoEmptyObjectTypeOptions{
			AllowInterfaces:  "never",
			AllowObjectTypes: "never",
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
				if allowInterfaces, ok := optsMap["allowInterfaces"].(string); ok {
					opts.AllowInterfaces = allowInterfaces
				}
				if allowObjectTypes, ok := optsMap["allowObjectTypes"].(string); ok {
					opts.AllowObjectTypes = allowObjectTypes
				}
				if allowWithName, ok := optsMap["allowWithName"].(string); ok {
					opts.AllowWithName = allowWithName
				}
			}
		}

		var allowWithNameRegex *regexp.Regexp
		if opts.AllowWithName != "" {
			allowWithNameRegex = regexp.MustCompile(opts.AllowWithName)
		}

		listeners := rule.RuleListeners{}

		// Handle empty interfaces
		if opts.AllowInterfaces != "always" {
			listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
				interfaceDecl := node.AsInterfaceDeclaration()

				// Check allowWithName
				if allowWithNameRegex != nil {
					nameText := ""
					if interfaceDecl.Name() != nil {
						nameText = getNodeText(ctx, interfaceDecl.Name())
					}
					if allowWithNameRegex.MatchString(nameText) {
						return
					}
				}

				var extendsList []*ast.Node
				if interfaceDecl.HeritageClauses != nil {
					for _, clause := range interfaceDecl.HeritageClauses.Nodes {
						if clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
							extendsList = clause.AsHeritageClause().Types.Nodes
							break
						}
					}
				}

				// Check if interface has members
				if interfaceDecl.Members != nil && len(interfaceDecl.Members.Nodes) > 0 {
					return
				}

				// For with-single-extends, allow if there's exactly one extend
				if len(extendsList) == 1 && opts.AllowInterfaces == "with-single-extends" {
					return
				}

				// Allow if extending multiple interfaces
				if len(extendsList) > 1 {
					return
				}

				// Check if merged with class declaration (not class expression)
				mergedWithClass := false
				if interfaceDecl.Name() != nil {
					symbol := ctx.TypeChecker.GetSymbolAtLocation(interfaceDecl.Name())
					if symbol != nil && symbol.Declarations != nil {
						for _, decl := range symbol.Declarations {
							// Only count class declarations, not class expressions
							if decl.Kind == ast.KindClassDeclaration {
								mergedWithClass = true
								break
							}
						}
					}
				}

				// Report empty interface with no extends
				if len(extendsList) == 0 {
					message := rule.RuleMessage{
						Id:          "noEmptyInterface",
						Description: noEmptyMessage("An empty interface declaration"),
					}

					suggestions := []rule.RuleSuggestion{}
					if !mergedWithClass {
						for _, replacement := range []string{"object", "unknown"} {
							suggestions = append(suggestions, rule.RuleSuggestion{
								Message: rule.RuleMessage{
									Id:          "replaceEmptyInterface",
									Description: fmt.Sprintf("Replace empty interface with `%s`.", replacement),
								},
								FixesArr: []rule.RuleFix{
									{
										Range: utils.TrimNodeTextRange(ctx.SourceFile, node),
										Text:  buildTypeAliasReplacement(ctx, interfaceDecl, replacement),
									},
								},
							})
						}
					}

					if interfaceDecl.Name() != nil {
						ctx.ReportNodeWithSuggestions(interfaceDecl.Name(), message, suggestions...)
					} else {
						ctx.ReportNodeWithSuggestions(node, message, suggestions...)
					}
					return
				}

				// Report interface with single extend
				message := rule.RuleMessage{
					Id:          "noEmptyInterfaceWithSuper",
					Description: "An interface declaring no members is equivalent to its supertype.",
				}

				suggestions := []rule.RuleSuggestion{}
				if !mergedWithClass {
					extendedTypeText := getExtendsText(ctx, extendsList[0])
					suggestions = append(suggestions, rule.RuleSuggestion{
						Message: rule.RuleMessage{
							Id:          "replaceEmptyInterfaceWithSuper",
							Description: "Replace empty interface with a type alias.",
						},
						FixesArr: []rule.RuleFix{
							{
								Range: utils.TrimNodeTextRange(ctx.SourceFile, node),
								Text:  buildTypeAliasReplacement(ctx, interfaceDecl, extendedTypeText),
							},
						},
					})
				}

				ctx.ReportNodeWithSuggestions(interfaceDecl.Name(), message, suggestions...)
			}
		}

		// Handle empty object types
		if opts.AllowObjectTypes != "always" {
			listeners[ast.KindTypeLiteral] = func(node *ast.Node) {
				typeLiteral := node.AsTypeLiteralNode()

				// Check if it has members
				if typeLiteral.Members != nil && len(typeLiteral.Members.Nodes) > 0 {
					return
				}

				// Don't report if part of intersection type
				if node.Parent != nil && ast.IsIntersectionTypeNode(node.Parent) {
					return
				}

				// Check allowWithName for type aliases
				if allowWithNameRegex != nil && node.Parent != nil && ast.IsTypeAliasDeclaration(node.Parent) {
					typeAlias := node.Parent.AsTypeAliasDeclaration()
					nameText := ""
					if typeAlias.Name() != nil {
						nameText = getNodeText(ctx, typeAlias.Name())
					}
					if allowWithNameRegex.MatchString(nameText) {
						return
					}
				}

				message := rule.RuleMessage{
					Id:          "noEmptyObject",
					Description: noEmptyMessage("The `{}` (\"empty object\") type"),
				}

				suggestions := []rule.RuleSuggestion{}
				for _, replacement := range []string{"object", "unknown"} {
					suggestions = append(suggestions, rule.RuleSuggestion{
						Message: rule.RuleMessage{
							Id:          "replaceEmptyObjectType",
							Description: fmt.Sprintf("Replace `{}` with `%s`.", replacement),
						},
						FixesArr: []rule.RuleFix{
							{
								Range: utils.TrimNodeTextRange(ctx.SourceFile, node),
								Text:  replacement,
							},
						},
					})
				}

				ctx.ReportNodeWithSuggestions(node, message, suggestions...)
			}
		}

		return listeners
	},
}

func getNodeText(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil {
		return ""
	}
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	return string(ctx.SourceFile.Text()[textRange.Pos():textRange.End()])
}

func getNodeListTextWithBrackets(ctx rule.RuleContext, nodeList *ast.NodeList) string {
	if nodeList == nil {
		return ""
	}
	// Find the opening and closing angle brackets using scanner
	openBracketPos := nodeList.Pos() - 1

	// Find closing bracket after the nodeList
	s := scanner.GetScannerForSourceFile(ctx.SourceFile, nodeList.End())
	closeBracketPos := nodeList.End()
	for s.TokenStart() < ctx.SourceFile.End() {
		if s.Token() == ast.KindGreaterThanToken {
			closeBracketPos = s.TokenEnd()
			break
		}
		if s.Token() != ast.KindWhitespaceTrivia && s.Token() != ast.KindNewLineTrivia {
			break
		}
		s.Scan()
	}

	textRange := core.NewTextRange(openBracketPos, closeBracketPos)
	return string(ctx.SourceFile.Text()[textRange.Pos():textRange.End()])
}

func buildTypeAliasReplacement(ctx rule.RuleContext, interfaceDecl *ast.InterfaceDeclaration, replacement string) string {
	// Get interface name
	nameText := ""
	if interfaceDecl.Name() != nil {
		nameText = getNodeText(ctx, interfaceDecl.Name())
	}

	// Get type parameters if any
	typeParamsText := ""
	if interfaceDecl.TypeParameters != nil {
		typeParamsText = getNodeListTextWithBrackets(ctx, interfaceDecl.TypeParameters)
	}

	// Check for export modifier
	exportText := ""
	if interfaceDecl.Modifiers() != nil {
		for _, mod := range interfaceDecl.Modifiers().Nodes {
			if mod.Kind == ast.KindExportKeyword {
				exportText = "export "
				break
			}
		}
	}

	return fmt.Sprintf("%stype %s%s = %s", exportText, nameText, typeParamsText, replacement)
}

func getExtendsText(ctx rule.RuleContext, extendsNode *ast.Node) string {
	extendsRange := utils.TrimNodeTextRange(ctx.SourceFile, extendsNode)
	return string(ctx.SourceFile.Text()[extendsRange.Pos():extendsRange.End()])
}
