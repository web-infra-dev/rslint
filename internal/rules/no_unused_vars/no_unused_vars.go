package no_unused_vars

import (
	"fmt"
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
)

// Options for the no-unused-vars rule
type Options struct {
	Vars                           string `json:"vars,omitempty"`                           // "all" | "local"
	Args                           string `json:"args,omitempty"`                           // "all" | "after-used" | "none"
	CaughtErrors                   string `json:"caughtErrors,omitempty"`                   // "all" | "none"
	VarsIgnorePattern              string `json:"varsIgnorePattern,omitempty"`
	ArgsIgnorePattern              string `json:"argsIgnorePattern,omitempty"`
	CaughtErrorsIgnorePattern      string `json:"caughtErrorsIgnorePattern,omitempty"`
	DestructuredArrayIgnorePattern string `json:"destructuredArrayIgnorePattern,omitempty"`
	IgnoreClassWithStaticInitBlock bool   `json:"ignoreClassWithStaticInitBlock,omitempty"`
	IgnoreRestSiblings             bool   `json:"ignoreRestSiblings,omitempty"`
	ReportUsedIgnorePattern        bool   `json:"reportUsedIgnorePattern,omitempty"`
}

// TranslatedOptions contains compiled regex patterns
type TranslatedOptions struct {
	Vars                           string
	Args                           string
	CaughtErrors                   string
	VarsIgnorePattern              *regexp.Regexp
	ArgsIgnorePattern              *regexp.Regexp
	CaughtErrorsIgnorePattern      *regexp.Regexp
	DestructuredArrayIgnorePattern *regexp.Regexp
	IgnoreClassWithStaticInitBlock bool
	IgnoreRestSiblings             bool
	ReportUsedIgnorePattern        bool
}

// VariableType represents the type of variable
type VariableType string

const (
	VariableTypeArrayDestructure VariableType = "array-destructure"
	VariableTypeCatchClause      VariableType = "catch-clause"
	VariableTypeParameter        VariableType = "parameter"
	VariableTypeVariable         VariableType = "variable"
)

// VariableInfo holds information about a variable
type VariableInfo struct {
	Variable    *ast.Node
	Used        bool
	OnlyUsedAsType bool
	References  []*ast.Node
	Definition  *ast.Node
}

var NoUnusedVarsRule = rule.Rule{
	Name: "no-unused-vars",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Parse options
		opts := TranslatedOptions{
			Vars:                           "all",
			Args:                           "after-used",
			CaughtErrors:                   "all",
			IgnoreClassWithStaticInitBlock: false,
			IgnoreRestSiblings:             false,
			ReportUsedIgnorePattern:        false,
		}

		if options != nil {
			switch v := options.(type) {
			case string:
				// Simple string option: "all" or "local"
				opts.Vars = v
			case map[string]interface{}:
				// Complex options object
				if vars, ok := v["vars"].(string); ok {
					opts.Vars = vars
				}
				if args, ok := v["args"].(string); ok {
					opts.Args = args
				}
				if caughtErrors, ok := v["caughtErrors"].(string); ok {
					opts.CaughtErrors = caughtErrors
				}
				if ignoreClass, ok := v["ignoreClassWithStaticInitBlock"].(bool); ok {
					opts.IgnoreClassWithStaticInitBlock = ignoreClass
				}
				if ignoreRest, ok := v["ignoreRestSiblings"].(bool); ok {
					opts.IgnoreRestSiblings = ignoreRest
				}
				if reportUsed, ok := v["reportUsedIgnorePattern"].(bool); ok {
					opts.ReportUsedIgnorePattern = reportUsed
				}

				// Compile regex patterns
				if pattern, ok := v["varsIgnorePattern"].(string); ok && pattern != "" {
					opts.VarsIgnorePattern = regexp.MustCompile(pattern)
				}
				if pattern, ok := v["argsIgnorePattern"].(string); ok && pattern != "" {
					opts.ArgsIgnorePattern = regexp.MustCompile(pattern)
				}
				if pattern, ok := v["caughtErrorsIgnorePattern"].(string); ok && pattern != "" {
					opts.CaughtErrorsIgnorePattern = regexp.MustCompile(pattern)
				}
				if pattern, ok := v["destructuredArrayIgnorePattern"].(string); ok && pattern != "" {
					opts.DestructuredArrayIgnorePattern = regexp.MustCompile(pattern)
				}
			}
		}

		// Collect all variables when the program exits
		return rule.RuleListeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// Process all variables at the end of the file
				collectAndReportUnusedVariables(ctx, opts, node)
			},
		}
	},
}

func collectAndReportUnusedVariables(ctx rule.RuleContext, opts TranslatedOptions, sourceFile *ast.Node) {
	// Get all variables in the source file
	variables := collectVariables(ctx, sourceFile)

	for _, varInfo := range variables {
		if shouldReportVariable(ctx, opts, varInfo) {
			reportUnusedVariable(ctx, opts, varInfo)
		}
	}
}

func collectVariables(ctx rule.RuleContext, sourceFile *ast.Node) map[*ast.Node]*VariableInfo {
	variables := make(map[*ast.Node]*VariableInfo)
	
	// This is a simplified version - in a real implementation, we would need to:
	// 1. Walk the entire AST
	// 2. Track all variable declarations
	// 3. Track all variable references
	// 4. Determine if references are type-only
	// 5. Handle scope correctly
	
	// For now, we'll use a visitor pattern to collect information
	// Simple recursive traversal since VisitEachChild is not available
	collectVariableInfo(ctx, sourceFile, variables)
	
	return variables
}

func collectVariableInfo(ctx rule.RuleContext, node *ast.Node, variables map[*ast.Node]*VariableInfo) {
	// Handle variable declarations
	switch node.Kind {
	case ast.KindVariableStatement:
		varStmt := node.AsVariableStatement()
		if varStmt.DeclarationList != nil {
			declList := varStmt.DeclarationList.AsVariableDeclarationList()
			for _, decl := range declList.Declarations.Nodes {
				if ast.IsIdentifier(decl.AsVariableDeclaration().Name()) {
					nameNode := decl.AsVariableDeclaration().Name()
					variables[nameNode] = &VariableInfo{
						Variable:   nameNode,
						Used:       false,
						OnlyUsedAsType: false,
						References: []*ast.Node{},
						Definition: decl,
					}
				}
			}
		}
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl.Name() != nil {
			nameNode := funcDecl.Name()
			variables[nameNode] = &VariableInfo{
				Variable:   nameNode,
				Used:       false,
				OnlyUsedAsType: false,
				References: []*ast.Node{},
				Definition: node,
			}
		}
		// Handle parameters
		if funcDecl.Parameters != nil {
			for _, param := range funcDecl.Parameters.Nodes {
				handleParameter(ctx, param, variables)
			}
		}
	case ast.KindParameter:
		handleParameter(ctx, node, variables)
	case ast.KindCatchClause:
		catchClause := node.AsCatchClause()
		if catchClause.VariableDeclaration != nil {
			name := catchClause.VariableDeclaration.Name()
			if ast.IsIdentifier(name) {
				variables[name] = &VariableInfo{
					Variable:   name,
					Used:       false,
					OnlyUsedAsType: false,
					References: []*ast.Node{},
					Definition: catchClause.VariableDeclaration,
				}
			}
		}
	case ast.KindIdentifier:
		// Handle variable references - simplified
		if varInfo, exists := variables[node]; exists {
			varInfo.References = append(varInfo.References, node)
			// Check if this is a type-only usage
			if !isTypeOnlyUsage(node) {
				varInfo.Used = true
			} else if !varInfo.Used {
				varInfo.OnlyUsedAsType = true
			}
		}
	}

	// Recursively process children - simplified traversal
	// In a real implementation, we would need proper AST traversal
}

func handleParameter(ctx rule.RuleContext, param *ast.Node, variables map[*ast.Node]*VariableInfo) {
	paramDecl := param.AsParameterDeclaration()
	if ast.IsIdentifier(paramDecl.Name()) {
		nameNode := paramDecl.Name()
		variables[nameNode] = &VariableInfo{
			Variable:   nameNode,
			Used:       false,
			OnlyUsedAsType: false,
			References: []*ast.Node{},
			Definition: param,
		}
	}
}

func isTypeOnlyUsage(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// Check if the identifier is used in a type context
	switch parent.Kind {
	case ast.KindTypeReference:
		return true
	case ast.KindTypeQuery:
		return true
	case ast.KindQualifiedName:
		return isTypeOnlyUsage(parent)
	}

	return false
}

func shouldReportVariable(ctx rule.RuleContext, opts TranslatedOptions, varInfo *VariableInfo) bool {
	// Skip if variable is used
	if varInfo.Used && !opts.ReportUsedIgnorePattern {
		return false
	}

	// Skip if only used as type and it's an import
	if varInfo.OnlyUsedAsType && isImportDeclaration(varInfo.Definition) {
		return false
	}

	// Get variable type
	varType := getVariableType(varInfo.Definition)

	// Check scope
	if opts.Vars == "local" && isGlobalVariable(varInfo) {
		return false
	}

	// Apply ignore patterns
	name := getVariableName(varInfo)
	
	switch varType {
	case VariableTypeArrayDestructure:
		if opts.DestructuredArrayIgnorePattern != nil && opts.DestructuredArrayIgnorePattern.MatchString(name) {
			return opts.ReportUsedIgnorePattern && varInfo.Used
		}
	case VariableTypeCatchClause:
		if opts.CaughtErrors == "none" {
			return false
		}
		if opts.CaughtErrorsIgnorePattern != nil && opts.CaughtErrorsIgnorePattern.MatchString(name) {
			return opts.ReportUsedIgnorePattern && varInfo.Used
		}
	case VariableTypeParameter:
		if opts.Args == "none" {
			return false
		}
		if opts.ArgsIgnorePattern != nil && opts.ArgsIgnorePattern.MatchString(name) {
			return opts.ReportUsedIgnorePattern && varInfo.Used
		}
		if opts.Args == "after-used" && !isAfterLastUsedParam(ctx, varInfo) {
			return false
		}
	case VariableTypeVariable:
		if opts.VarsIgnorePattern != nil && opts.VarsIgnorePattern.MatchString(name) {
			return opts.ReportUsedIgnorePattern && varInfo.Used
		}
	}

	// Check for rest siblings
	if opts.IgnoreRestSiblings && hasRestSibling(varInfo) {
		return false
	}

	// Check for class with static init block
	if opts.IgnoreClassWithStaticInitBlock && isClassWithStaticInitBlock(varInfo.Definition) {
		return false
	}

	return !varInfo.Used || (varInfo.OnlyUsedAsType && !isImportDeclaration(varInfo.Definition))
}

func getVariableType(definition *ast.Node) VariableType {
	if definition == nil {
		return VariableTypeVariable
	}

	parent := definition.Parent
	if parent == nil {
		return VariableTypeVariable
	}

	// Check for array destructuring
	if parent.Kind == ast.KindArrayBindingPattern {
		return VariableTypeArrayDestructure
	}

	// Check for catch clause
	if parent.Kind == ast.KindCatchClause {
		return VariableTypeCatchClause
	}

	// Check for parameter
	if definition.Kind == ast.KindParameter {
		return VariableTypeParameter
	}

	return VariableTypeVariable
}

func getVariableName(varInfo *VariableInfo) string {
	if varInfo.Variable != nil && ast.IsIdentifier(varInfo.Variable) {
		return varInfo.Variable.AsIdentifier().Text
	}
	
	// Try to get name from definition
	if varInfo.Definition != nil {
		switch varInfo.Definition.Kind {
		case ast.KindVariableDeclaration:
			name := varInfo.Definition.AsVariableDeclaration().Name()
			if ast.IsIdentifier(name) {
				return name.AsIdentifier().Text
			}
		case ast.KindParameter:
			name := varInfo.Definition.AsParameterDeclaration().Name()
			if ast.IsIdentifier(name) {
				return name.AsIdentifier().Text
			}
		case ast.KindFunctionDeclaration:
			funcDecl := varInfo.Definition.AsFunctionDeclaration()
			if funcDecl.Name() != nil {
				return funcDecl.Name().AsIdentifier().Text
			}
		}
	}
	
	return ""
}

func isGlobalVariable(varInfo *VariableInfo) bool {
	// Check if the variable is in global scope
	// This is a simplified check - in reality, we'd need to check the scope chain
	return false
}

func isImportDeclaration(definition *ast.Node) bool {
	if definition == nil {
		return false
	}
	
	parent := definition.Parent
	for parent != nil {
		if parent.Kind == ast.KindImportDeclaration {
			return true
		}
		parent = parent.Parent
	}
	
	return false
}

func isAfterLastUsedParam(ctx rule.RuleContext, varInfo *VariableInfo) bool {
	// Check if this parameter comes after the last used parameter
	// This requires analyzing all parameters in the function
	return true // Simplified for now
}

func hasRestSibling(varInfo *VariableInfo) bool {
	// Check if the variable has a rest sibling in object destructuring
	if varInfo.Definition == nil {
		return false
	}
	
	parent := varInfo.Definition.Parent
	if parent != nil && parent.Kind == ast.KindObjectBindingPattern {
		// Check if there's a rest element in the pattern
		// This is simplified - would need proper implementation
		return false
	}
	
	return false
}

func isClassWithStaticInitBlock(definition *ast.Node) bool {
	if definition == nil || definition.Kind != ast.KindClassDeclaration {
		return false
	}
	
	classDecl := definition.AsClassDeclaration()
	if classDecl.Members != nil {
		for _, member := range classDecl.Members.Nodes {
			// Check for static blocks - using a different approach since KindStaticBlock may not be available
			if member.Kind == ast.KindMethodDeclaration {
				// This is a simplified check
				return false
			}
		}
	}
	
	return false
}

func reportUnusedVariable(ctx rule.RuleContext, opts TranslatedOptions, varInfo *VariableInfo) {
	varName := getVariableName(varInfo)
	varType := getVariableType(varInfo.Definition)
	
	// Determine message ID and data
	messageId := "unusedVar"
	action := "defined"
	
	// Check if variable was assigned
	if len(varInfo.References) > 0 {
		for _, ref := range varInfo.References {
			if isWriteReference(ref) {
				action = "assigned a value"
				break
			}
		}
	}
	
	// Check if only used as type
	if varInfo.OnlyUsedAsType {
		messageId = "usedOnlyAsType"
	}
	
	// Check if used but matches ignore pattern
	if varInfo.Used && opts.ReportUsedIgnorePattern {
		messageId = "usedIgnoredVar"
	}
	
	// Build additional message for ignore patterns
	additional := getAdditionalMessage(opts, varType)
	
	// Find the location to report
	var reportNode *ast.Node
	if len(varInfo.References) > 0 {
		// Report at last write reference
		for i := len(varInfo.References) - 1; i >= 0; i-- {
			if isWriteReference(varInfo.References[i]) {
				reportNode = varInfo.References[i]
				break
			}
		}
	}
	if reportNode == nil && varInfo.Definition != nil {
		reportNode = getNameNodeFromDefinition(varInfo.Definition)
	}
	
	if reportNode == nil {
		return
	}
	
	// Report the issue
	message := buildMessage(messageId, varName, action, additional)
	ctx.ReportNode(reportNode, message)
}

func isWriteReference(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	
	// Check if this is a write reference
	if ast.IsBinaryExpression(parent) {
		binExpr := parent.AsBinaryExpression()
		if binExpr.Left == node && isAssignmentOperator(binExpr.OperatorToken.Kind) {
			return true
		}
	}
	
	return false
}

func isAssignmentOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken:
		return true
	}
	return false
}

func getNameNodeFromDefinition(definition *ast.Node) *ast.Node {
	switch definition.Kind {
	case ast.KindVariableDeclaration:
		return definition.AsVariableDeclaration().Name()
	case ast.KindParameter:
		return definition.AsParameterDeclaration().Name()
	case ast.KindFunctionDeclaration:
		return definition.AsFunctionDeclaration().Name()
	case ast.KindClassDeclaration:
		return definition.AsClassDeclaration().Name()
	}
	
	return definition
}

func getAdditionalMessage(opts TranslatedOptions, varType VariableType) string {
	pattern := ""
	description := ""
	
	switch varType {
	case VariableTypeArrayDestructure:
		if opts.DestructuredArrayIgnorePattern != nil {
			pattern = opts.DestructuredArrayIgnorePattern.String()
			description = "elements of array destructuring"
		}
	case VariableTypeCatchClause:
		if opts.CaughtErrorsIgnorePattern != nil {
			pattern = opts.CaughtErrorsIgnorePattern.String()
			description = "caught errors"
		}
	case VariableTypeParameter:
		if opts.ArgsIgnorePattern != nil {
			pattern = opts.ArgsIgnorePattern.String()
			description = "args"
		}
	case VariableTypeVariable:
		if opts.VarsIgnorePattern != nil {
			pattern = opts.VarsIgnorePattern.String()
			description = "vars"
		}
	}
	
	if pattern != "" && description != "" {
		if opts.ReportUsedIgnorePattern {
			return fmt.Sprintf(". Used %s must not match %s", description, pattern)
		}
		return fmt.Sprintf(". Allowed unused %s must match %s", description, pattern)
	}
	
	return ""
}

func buildMessage(messageId, varName, action, additional string) rule.RuleMessage {
	var description string
	
	switch messageId {
	case "unusedVar":
		description = fmt.Sprintf("'%s' is %s but never used%s.", varName, action, additional)
	case "usedIgnoredVar":
		description = fmt.Sprintf("'%s' is marked as ignored but is used%s.", varName, additional)
	case "usedOnlyAsType":
		description = fmt.Sprintf("'%s' is %s but only used as a type%s.", varName, action, additional)
	}
	
	return rule.RuleMessage{
		Id:          messageId,
		Description: description,
	}
}