package no_require_imports

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoRequireImportsOptions struct {
	Allow         []string `json:"allow"`
	AllowAsImport bool     `json:"allowAsImport"`
}

// isStringOrTemplateLiteral checks if a node is a string literal or template literal
func isStringOrTemplateLiteral(node *ast.Node) bool {
	return (node.Kind == ast.KindStringLiteral) ||
		(node.Kind == ast.KindTemplateExpression && node.AsTemplateExpression().TemplateSpans == nil) ||
		(node.Kind == ast.KindNoSubstitutionTemplateLiteral)
}

// getStaticStringValue extracts static string value from literal or template
func getStaticStringValue(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text, true
	case ast.KindTemplateExpression:
		// Only handle simple template literals without expressions
		te := node.AsTemplateExpression()
		if te.TemplateSpans == nil || len(te.TemplateSpans.Nodes) == 0 {
			return te.Head.Text(), true
		}
	case ast.KindNoSubstitutionTemplateLiteral:
		// Handle simple template literals `string`
		return node.AsNoSubstitutionTemplateLiteral().Text, true
	}
	return "", false
}

// isGlobalRequire checks if the require is the global require function
func isGlobalRequire(ctx rule.RuleContext, node *ast.Node) bool {
	// Walk up to find the source file and traverse all variable declarations
	sourceFile := ctx.SourceFile
	if sourceFile == nil {
		return true
	}

	// Check if 'require' is defined anywhere in the current scope context
	return !isRequireLocallyDefined(sourceFile, node)
}

// isRequireLocallyDefined checks if require is locally defined in any containing scope
func isRequireLocallyDefined(sourceFile *ast.SourceFile, callNode *ast.Node) bool {
	// Start from the call node and walk up through containing scopes
	currentNode := callNode

	for currentNode != nil {
		// Check the immediate parent context for local require definitions
		if hasLocalRequireInContext(currentNode) {
			return true
		}
		currentNode = currentNode.Parent
	}

	// Also check the top-level statements
	return hasLocalRequireInStatements(sourceFile.Statements)
}

// hasLocalRequireInContext checks if a node's immediate context defines require
func hasLocalRequireInContext(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	// Check for variable statements that declare 'require'
	switch parent.Kind {
	case ast.KindVariableStatement:
		varStmt := parent.AsVariableStatement()
		if varStmt.DeclarationList != nil {
			declList := varStmt.DeclarationList.AsVariableDeclarationList()
			for _, declarator := range declList.Declarations.Nodes {
				varDecl := declarator.AsVariableDeclaration()
				if ast.IsIdentifier(varDecl.Name()) && varDecl.Name().AsIdentifier().Text == "require" {
					return true
				}
			}
		}
	case ast.KindBlock:
		// Check all statements in the block for require declarations
		block := parent.AsBlock()
		return hasLocalRequireInStatements(block.Statements)
	case ast.KindSourceFile:
		// Check all top-level statements
		sourceFile := parent.AsSourceFile()
		return hasLocalRequireInStatements(sourceFile.Statements)
	}

	return false
}

// hasLocalRequireInStatements checks if any statements define a local 'require'
func hasLocalRequireInStatements(statements *ast.NodeList) bool {
	if statements == nil {
		return false
	}

	for _, stmt := range statements.Nodes {
		if hasRequireDeclaration(stmt) {
			return true
		}
	}
	return false
}

// hasRequireDeclaration checks if a statement declares a 'require' variable
func hasRequireDeclaration(stmt *ast.Node) bool {
	switch stmt.Kind {
	case ast.KindVariableStatement:
		varStmt := stmt.AsVariableStatement()
		if varStmt.DeclarationList != nil {
			declList := varStmt.DeclarationList.AsVariableDeclarationList()
			for _, declarator := range declList.Declarations.Nodes {
				varDecl := declarator.AsVariableDeclaration()
				if ast.IsIdentifier(varDecl.Name()) && varDecl.Name().AsIdentifier().Text == "require" {
					return true
				}
			}
		}
	case ast.KindBlock:
		block := stmt.AsBlock()
		return hasLocalRequireInStatements(block.Statements)
	}
	return false
}

var NoRequireImportsRule = rule.CreateRule(rule.Rule{
	Name: "no-require-imports",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoRequireImportsOptions{
			Allow:         []string{},
			AllowAsImport: false,
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
				if allow, ok := optsMap["allow"].([]interface{}); ok {
					for _, pattern := range allow {
						if str, ok := pattern.(string); ok {
							opts.Allow = append(opts.Allow, str)
						}
					}
				}
				if allowAsImport, ok := optsMap["allowAsImport"].(bool); ok {
					opts.AllowAsImport = allowAsImport
				}
			}
		}

		// Compile regex patterns
		var allowPatterns []*regexp.Regexp
		for _, pattern := range opts.Allow {
			if compiled, err := regexp.Compile(pattern); err == nil {
				allowPatterns = append(allowPatterns, compiled)
			}
		}

		isImportPathAllowed := func(importPath string) bool {
			for _, pattern := range allowPatterns {
				if pattern.MatchString(importPath) {
					return true
				}
			}
			return false
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()

				// Check if this is a require call or require?.() call
				var isRequireCall bool

				if ast.IsIdentifier(callExpr.Expression) {
					identifier := callExpr.Expression.AsIdentifier()
					if identifier.Text == "require" {
						isRequireCall = true
					}
				} else if callExpr.QuestionDotToken != nil {
					// Handle optional chaining: require?.()
					// The expression should be require for require?.()
					if ast.IsIdentifier(callExpr.Expression) {
						identifier := callExpr.Expression.AsIdentifier()
						if identifier != nil && identifier.Text == "require" {
							isRequireCall = true
						}
					}
				}

				if !isRequireCall {
					return
				}

				// Check if it's the global require
				if !isGlobalRequire(ctx, node) {
					return
				}

				// Check if first argument matches allowed patterns
				if len(callExpr.Arguments.Nodes) > 0 && isStringOrTemplateLiteral(callExpr.Arguments.Nodes[0]) {
					if argValue, ok := getStaticStringValue(callExpr.Arguments.Nodes[0]); ok {
						if isImportPathAllowed(argValue) {
							return
						}
					}
				}

				// Report the error
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noRequireImports",
					Description: "A `require()` style import is forbidden.",
				})
			},

			ast.KindExternalModuleReference: func(node *ast.Node) {
				extModRef := node.AsExternalModuleReference()

				// Check if expression matches allowed patterns
				if isStringOrTemplateLiteral(extModRef.Expression) {
					if argValue, ok := getStaticStringValue(extModRef.Expression); ok {
						if isImportPathAllowed(argValue) {
							return
						}
					}
				}

				// Check if allowAsImport is true and parent is TSImportEqualsDeclaration
				if opts.AllowAsImport && node.Parent != nil &&
					node.Parent.Kind == ast.KindImportEqualsDeclaration {
					return
				}

				// Report the error
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noRequireImports",
					Description: "A `require()` style import is forbidden.",
				})
			},
		}
	},
})
