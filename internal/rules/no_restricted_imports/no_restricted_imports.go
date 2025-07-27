package no_restricted_imports

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gobwas/glob"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
)

// PathRestriction represents a restricted import path configuration
type PathRestriction struct {
	Name             string   `json:"name"`
	Message          string   `json:"message"`
	ImportNames      []string `json:"importNames"`
	AllowTypeImports bool     `json:"allowTypeImports"`
}

// PatternRestriction represents a restricted pattern configuration
type PatternRestriction struct {
	Group            []string `json:"group"`
	Regex            string   `json:"regex"`
	Message          string   `json:"message"`
	CaseSensitive    bool     `json:"caseSensitive"`
	AllowTypeImports bool     `json:"allowTypeImports"`
	ImportNames      []string `json:"importNames"`
}

// NoRestrictedImportsOptions represents the rule options
type NoRestrictedImportsOptions struct {
	Paths    []interface{} `json:"paths"`
	Patterns []interface{} `json:"patterns"`
}

// ParseOptions parses the rule options from various formats
func parseOptions(options any) ([]interface{}, []interface{}) {
	if options == nil {
		return nil, nil
	}

	// Handle array of strings or objects (first format)
	if arr, ok := options.([]interface{}); ok {
		if len(arr) == 0 {
			return nil, nil
		}

		// Check if first element is an object with paths/patterns
		if len(arr) == 1 {
			if obj, ok := arr[0].(map[string]interface{}); ok {
				paths, _ := obj["paths"].([]interface{})
				patterns, _ := obj["patterns"].([]interface{})
				return paths, patterns
			}
		}

		// Otherwise, it's an array of path restrictions
		return arr, nil
	}

	// Handle object with paths/patterns properties
	if obj, ok := options.(map[string]interface{}); ok {
		paths, _ := obj["paths"].([]interface{})
		patterns, _ := obj["patterns"].([]interface{})
		return paths, patterns
	}

	return nil, nil
}

// ParsePathRestriction parses a path restriction from various formats
func parsePathRestriction(item interface{}) (*PathRestriction, error) {
	// String format
	if str, ok := item.(string); ok {
		return &PathRestriction{Name: str}, nil
	}

	// Object format
	if obj, ok := item.(map[string]interface{}); ok {
		pr := &PathRestriction{}
		
		if name, ok := obj["name"].(string); ok {
			pr.Name = name
		}
		
		if message, ok := obj["message"].(string); ok {
			pr.Message = message
		}
		
		if allowTypeImports, ok := obj["allowTypeImports"].(bool); ok {
			pr.AllowTypeImports = allowTypeImports
		}
		
		if importNames, ok := obj["importNames"].([]interface{}); ok {
			for _, name := range importNames {
				if str, ok := name.(string); ok {
					pr.ImportNames = append(pr.ImportNames, str)
				}
			}
		}
		
		return pr, nil
	}

	return nil, fmt.Errorf("invalid path restriction format")
}

// ParsePatternRestriction parses a pattern restriction from various formats
func parsePatternRestriction(item interface{}) (*PatternRestriction, error) {
	// String format (treated as group pattern)
	if str, ok := item.(string); ok {
		return &PatternRestriction{Group: []string{str}}, nil
	}

	// Object format
	if obj, ok := item.(map[string]interface{}); ok {
		pr := &PatternRestriction{CaseSensitive: true} // Default case sensitive
		
		if group, ok := obj["group"].([]interface{}); ok {
			for _, g := range group {
				if str, ok := g.(string); ok {
					pr.Group = append(pr.Group, str)
				}
			}
		}
		
		if regex, ok := obj["regex"].(string); ok {
			pr.Regex = regex
		}
		
		if message, ok := obj["message"].(string); ok {
			pr.Message = message
		}
		
		if caseSensitive, ok := obj["caseSensitive"].(bool); ok {
			pr.CaseSensitive = caseSensitive
		}
		
		if allowTypeImports, ok := obj["allowTypeImports"].(bool); ok {
			pr.AllowTypeImports = allowTypeImports
		}
		
		if importNames, ok := obj["importNames"].([]interface{}); ok {
			for _, name := range importNames {
				if str, ok := name.(string); ok {
					pr.ImportNames = append(pr.ImportNames, str)
				}
			}
		}
		
		return pr, nil
	}

	return nil, fmt.Errorf("invalid pattern restriction format")
}

// GlobMatcher represents a compiled glob pattern matcher
type GlobMatcher struct {
	patterns []glob.Glob
	excludes []glob.Glob
}

// NewGlobMatcher creates a new glob matcher from patterns
func NewGlobMatcher(patterns []string, caseSensitive bool) (*GlobMatcher, error) {
	matcher := &GlobMatcher{}
	
	for _, pattern := range patterns {
		isExclude := strings.HasPrefix(pattern, "!")
		if isExclude {
			pattern = pattern[1:] // Remove the !
		}
		
		// Compile the glob pattern
		g, err := glob.Compile(pattern, '/')
		if err != nil {
			return nil, err
		}
		
		if isExclude {
			matcher.excludes = append(matcher.excludes, g)
		} else {
			matcher.patterns = append(matcher.patterns, g)
		}
	}
	
	return matcher, nil
}

// Matches checks if a string matches the glob patterns
func (m *GlobMatcher) Matches(str string) bool {
	// Check if any exclude pattern matches
	for _, exclude := range m.excludes {
		if exclude.Match(str) {
			return false
		}
	}
	
	// Check if any include pattern matches
	for _, pattern := range m.patterns {
		if pattern.Match(str) {
			return true
		}
	}
	
	return false
}

var NoRestrictedImportsRule = rule.Rule{
	Name: "no-restricted-imports",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		paths, patterns := parseOptions(options)
		
		// Return empty listeners if no restrictions configured
		if len(paths) == 0 && len(patterns) == 0 {
			return rule.RuleListeners{}
		}

		// Parse path restrictions
		var pathRestrictions []*PathRestriction
		allowedTypeImportPaths := make(map[string]bool)
		
		for _, item := range paths {
			pr, err := parsePathRestriction(item)
			if err != nil {
				continue
			}
			pathRestrictions = append(pathRestrictions, pr)
			
			if pr.AllowTypeImports {
				allowedTypeImportPaths[pr.Name] = true
			}
		}

		// Parse pattern restrictions  
		var patternRestrictions []*PatternRestriction
		var typeImportMatchers []*GlobMatcher
		var typeImportRegexes []*regexp.Regexp
		
		for _, item := range patterns {
			pr, err := parsePatternRestriction(item)
			if err != nil {
				continue
			}
			patternRestrictions = append(patternRestrictions, pr)
			
			if pr.AllowTypeImports {
				if len(pr.Group) > 0 {
					matcher, err := NewGlobMatcher(pr.Group, pr.CaseSensitive)
					if err == nil {
						typeImportMatchers = append(typeImportMatchers, matcher)
					}
				}
				
				if pr.Regex != "" {
					flags := "(?s)" // s flag for . to match newlines
					if !pr.CaseSensitive {
						flags += "(?i)" // i flag for case insensitive
					}
					re, err := regexp.Compile(flags + pr.Regex)
					if err == nil {
						typeImportRegexes = append(typeImportRegexes, re)
					}
				}
			}
		}

		// Helper to check if import is allowed for type imports
		isAllowedTypeImport := func(importSource string) bool {
			// Check paths
			if allowedTypeImportPaths[importSource] {
				return true
			}
			
			// Check patterns
			for _, matcher := range typeImportMatchers {
				if matcher.Matches(importSource) {
					return true
				}
			}
			
			for _, re := range typeImportRegexes {
				if re.MatchString(importSource) {
					return true
				}
			}
			
			return false
		}

		// Helper to check import against path restrictions
		checkPathRestrictions := func(node *ast.Node, importSource string, importedNames []string) {
			for _, pr := range pathRestrictions {
				if pr.Name != importSource {
					continue
				}
				
				// Check if specific import names are restricted
				if len(pr.ImportNames) > 0 && len(importedNames) > 0 {
					for _, importedName := range importedNames {
						for _, restrictedName := range pr.ImportNames {
							if importedName == restrictedName {
								msg := pr.Message
								if msg == "" {
									msg = fmt.Sprintf("'%s' import from '%s' is restricted.", restrictedName, pr.Name)
								}
								ctx.ReportNode(node, rule.RuleMessage{
									Id:          "importNameWithCustomMessage",
									Description: msg,
								})
								return
							}
						}
					}
				} else {
					// Entire module is restricted
					msg := pr.Message
					if msg == "" {
						msg = fmt.Sprintf("'%s' import is restricted.", pr.Name)
					}
					messageId := "path"
					if pr.Message != "" {
						messageId = "pathWithCustomMessage"
					}
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          messageId,
						Description: msg,
					})
					return
				}
			}
		}

		// Helper to check import against pattern restrictions
		checkPatternRestrictions := func(node *ast.Node, importSource string) {
			for _, pr := range patternRestrictions {
				matched := false
				
				// Check glob patterns
				if len(pr.Group) > 0 {
					matcher, err := NewGlobMatcher(pr.Group, pr.CaseSensitive)
					if err == nil && matcher.Matches(importSource) {
						matched = true
					}
				}
				
				// Check regex
				if pr.Regex != "" && !matched {
					flags := "(?s)"
					if !pr.CaseSensitive {
						flags += "(?i)"
					}
					re, err := regexp.Compile(flags + pr.Regex)
					if err == nil && re.MatchString(importSource) {
						matched = true
					}
				}
				
				if matched {
					msg := pr.Message
					if msg == "" {
						msg = fmt.Sprintf("'%s' import is restricted by pattern.", importSource)
					}
					messageId := "patterns"
					if pr.Message != "" {
						messageId = "patternWithCustomMessage"
					}
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          messageId,
						Description: msg,
					})
					return
				}
			}
		}

		// Helper to check if import/export is type-only
		isTypeOnlyImportExport := func(node *ast.Node) bool {
			switch node.Kind {
			case ast.KindImportDeclaration:
				importDecl := node.AsImportDeclaration()
				if importDecl.ImportClause == nil {
					return false
				}
				
				importClause := importDecl.ImportClause.AsImportClause()
				if importClause.IsTypeOnly {
					return true
				}
				
				// Check if all specifiers are type-only
				if importClause.NamedBindings != nil && ast.IsNamedImports(importClause.NamedBindings) {
					namedImports := importClause.NamedBindings.AsNamedImports()
					if namedImports.Elements != nil {
						allTypeOnly := true
						for _, elem := range namedImports.Elements.Nodes {
							if !elem.AsImportSpecifier().IsTypeOnly {
								allTypeOnly = false
								break
							}
						}
						return allTypeOnly
					}
				}
				
			case ast.KindExportDeclaration:
				exportDecl := node.AsExportDeclaration()
				if exportDecl.IsTypeOnly {
					return true
				}
				
				// Check if all specifiers are type-only
				if exportDecl.ExportClause != nil && ast.IsNamedExports(exportDecl.ExportClause) {
					namedExports := exportDecl.ExportClause.AsNamedExports()
					if namedExports.Elements != nil {
						allTypeOnly := true
						for _, elem := range namedExports.Elements.Nodes {
							if !elem.AsExportSpecifier().IsTypeOnly {
								allTypeOnly = false
								break
							}
						}
						return allTypeOnly
					}
				}
				
			case ast.KindImportEqualsDeclaration:
				importEqualsDecl := node.AsImportEqualsDeclaration()
				return importEqualsDecl.IsTypeOnly
			}
			
			return false
		}

		// Helper to get imported names from import declaration
		getImportedNames := func(node *ast.Node) []string {
			var names []string
			
			switch node.Kind {
			case ast.KindImportDeclaration:
				importDecl := node.AsImportDeclaration()
				if importDecl.ImportClause == nil {
					return names
				}
				
				importClause := importDecl.ImportClause.AsImportClause()
				
				// Default import
				if importClause.Name() != nil {
					names = append(names, importClause.Name().AsIdentifier().Text)
				}
				
				// Named imports
				if importClause.NamedBindings != nil {
					if ast.IsNamedImports(importClause.NamedBindings) {
						namedImports := importClause.NamedBindings.AsNamedImports()
						if namedImports.Elements != nil {
							for _, elem := range namedImports.Elements.Nodes {
								importSpec := elem.AsImportSpecifier()
								// Use the imported name (what's being imported from the module)
								if importSpec.PropertyName != nil {
									// import { originalName as localName }
									names = append(names, importSpec.PropertyName.AsIdentifier().Text)
								} else if importSpec.Name() != nil {
									// import { name }
									names = append(names, importSpec.Name().AsIdentifier().Text)
								}
							}
						}
					} else if ast.IsNamespaceImport(importClause.NamedBindings) {
						// Namespace imports don't have individual names
					}
				}
				
			case ast.KindExportDeclaration:
				exportDecl := node.AsExportDeclaration()
				if exportDecl.ExportClause != nil && ast.IsNamedExports(exportDecl.ExportClause) {
					namedExports := exportDecl.ExportClause.AsNamedExports()
					if namedExports.Elements != nil {
						for _, elem := range namedExports.Elements.Nodes {
							exportSpec := elem.AsExportSpecifier()
							// Use the imported name (what's being re-exported from the module)
							if exportSpec.PropertyName != nil {
								// export { originalName as exportedName }
								names = append(names, exportSpec.PropertyName.AsIdentifier().Text)
							} else if exportSpec.Name() != nil {
								// export { name }
								names = append(names, exportSpec.Name().AsIdentifier().Text)
							}
						}
					}
				}
				
			case ast.KindImportEqualsDeclaration:
				importEqualsDecl := node.AsImportEqualsDeclaration()
				if importEqualsDecl.Name() != nil {
					names = append(names, importEqualsDecl.Name().AsIdentifier().Text)
				}
			}
			
			return names
		}

		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				importDecl := node.AsImportDeclaration()
				importSource := strings.Trim(importDecl.ModuleSpecifier.AsStringLiteral().Text, "\"'`")
				
				// Skip if type-only import and allowed
				if isTypeOnlyImportExport(node) && isAllowedTypeImport(importSource) {
					return
				}
				
				importedNames := getImportedNames(node)
				
				// Check path restrictions
				checkPathRestrictions(node, importSource, importedNames)
				
				// Check pattern restrictions
				checkPatternRestrictions(node, importSource)
			},
			
			ast.KindExportDeclaration: func(node *ast.Node) {
				exportDecl := node.AsExportDeclaration()
				if exportDecl.ModuleSpecifier == nil {
					return
				}
				
				importSource := strings.Trim(exportDecl.ModuleSpecifier.AsStringLiteral().Text, "\"'`")
				
				// Skip if type-only export and allowed
				if isTypeOnlyImportExport(node) && isAllowedTypeImport(importSource) {
					return
				}
				
				// Get exported names for restriction checking
				exportedNames := getImportedNames(node)
				checkPathRestrictions(node, importSource, exportedNames)
				checkPatternRestrictions(node, importSource)
			},
			
			ast.KindImportEqualsDeclaration: func(node *ast.Node) {
				tsImportEquals := node.AsImportEqualsDeclaration()
				
				// Only handle external module references
				if tsImportEquals.ModuleReference.Kind != ast.KindExternalModuleReference {
					return
				}
				
				externalRef := tsImportEquals.ModuleReference.AsExternalModuleReference()
				importSource := strings.Trim(externalRef.Expression.AsStringLiteral().Text, "\"'`")
				
				// Skip if type-only import and allowed
				if isTypeOnlyImportExport(node) && isAllowedTypeImport(importSource) {
					return
				}
				
				// Get the imported name
				importedNames := getImportedNames(node)
				
				checkPathRestrictions(node, importSource, importedNames)
				checkPatternRestrictions(node, importSource)
			},
		}
	},
}