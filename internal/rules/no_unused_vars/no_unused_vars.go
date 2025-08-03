package no_unused_vars

import (
	"fmt"
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type Config struct {
	Vars                      string `json:"vars"`
	VarsIgnorePattern         string `json:"varsIgnorePattern"`
	Args                      string `json:"args"`
	ArgsIgnorePattern         string `json:"argsIgnorePattern"`
	CaughtErrors              string `json:"caughtErrors"`
	CaughtErrorsIgnorePattern string `json:"caughtErrorsIgnorePattern"`
	IgnoreRestSiblings        bool   `json:"ignoreRestSiblings"`
	ReportUsedIgnorePattern   bool   `json:"reportUsedIgnorePattern"`
}

type VariableInfo struct {
	Variable       *ast.Node
	Used           bool
	OnlyUsedAsType bool
	References     []*ast.Node
	Definition     *ast.Node
}

func parseOptions(options interface{}) Config {
	config := Config{
		Vars:                      "all",
		VarsIgnorePattern:         "",
		Args:                      "after-used",
		ArgsIgnorePattern:         "",
		CaughtErrors:              "all",
		CaughtErrorsIgnorePattern: "",
		IgnoreRestSiblings:        false,
		ReportUsedIgnorePattern:   false,
	}

	if options == nil {
		return config
	}

	// Handle object options
	if optsMap, ok := options.(map[string]interface{}); ok {
		if val, ok := optsMap["vars"].(string); ok {
			config.Vars = val
		}
		if val, ok := optsMap["varsIgnorePattern"].(string); ok {
			config.VarsIgnorePattern = val
		}
		if val, ok := optsMap["args"].(string); ok {
			config.Args = val
		}
		if val, ok := optsMap["argsIgnorePattern"].(string); ok {
			config.ArgsIgnorePattern = val
		}
		if val, ok := optsMap["caughtErrors"].(string); ok {
			config.CaughtErrors = val
		}
		if val, ok := optsMap["caughtErrorsIgnorePattern"].(string); ok {
			config.CaughtErrorsIgnorePattern = val
		}
		if val, ok := optsMap["ignoreRestSiblings"].(bool); ok {
			config.IgnoreRestSiblings = val
		}
		if val, ok := optsMap["reportUsedIgnorePattern"].(bool); ok {
			config.ReportUsedIgnorePattern = val
		}
	}

	return config
}

func isInTypeContext(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil {
		switch parent.Kind {
		case ast.KindTypeReference,
			ast.KindTypeAliasDeclaration,
			ast.KindInterfaceDeclaration,
			ast.KindTypeParameter,
			ast.KindTypeQuery,
			ast.KindTypeOperator,
			ast.KindIndexedAccessType,
			ast.KindConditionalType,
			ast.KindInferType,
			ast.KindTypeLiteral,
			ast.KindMappedType:
			return true
		case ast.KindAsExpression,
			ast.KindTypeAssertionExpression,
			ast.KindSatisfiesExpression:
			return true
		}
		parent = parent.Parent
	}
	return false
}

func isPartOfDeclaration(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		varDecl := parent.AsVariableDeclaration()
		return varDecl.Name() == node
	case ast.KindFunctionDeclaration:
		funcDecl := parent.AsFunctionDeclaration()
		return funcDecl.Name() == node
	case ast.KindParameter:
		paramDecl := parent.AsParameterDeclaration()
		return paramDecl.Name() == node
	case ast.KindClassDeclaration:
		classDecl := parent.AsClassDeclaration()
		return classDecl.Name() == node
	case ast.KindInterfaceDeclaration:
		interfaceDecl := parent.AsInterfaceDeclaration()
		return interfaceDecl.Name() == node
	case ast.KindTypeAliasDeclaration:
		typeAlias := parent.AsTypeAliasDeclaration()
		return typeAlias.Name() == node
	case ast.KindEnumDeclaration:
		enumDecl := parent.AsEnumDeclaration()
		return enumDecl.Name() == node
	case ast.KindCatchClause:
		// For catch clauses, the identifier is directly the VariableDeclaration
		// Only the actual catch variable declaration should be considered a declaration
		catchClause := parent.AsCatchClause()
		return catchClause.VariableDeclaration == node
	}

	return false
}

func isPartOfAssignment(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent
	if parent.Kind == ast.KindBinaryExpression {
		binaryExpr := parent.AsBinaryExpression()
		// Check if this is the left side of an assignment - assignments should not count as usage
		if binaryExpr.OperatorToken.Kind == ast.KindEqualsToken && binaryExpr.Left == node {
			return true
		}
	}

	return false
}

func shouldIgnoreVariable(varName string, varInfo *VariableInfo, opts Config) bool {
	// Check if it matches ignore patterns
	if opts.VarsIgnorePattern != "" {
		if matched, _ := regexp.MatchString(opts.VarsIgnorePattern, varName); matched {
			if !varInfo.Used || !opts.ReportUsedIgnorePattern {
				return true
			}
		}
	}

	// Check if it's a function parameter and should be ignored
	if isParameter(varInfo.Definition) {
		return shouldIgnoreParameter(varName, opts)
	}

	// Check if it's a caught error and should be ignored
	if isCaughtError(varInfo.Definition) {
		return shouldIgnoreCaughtError(varName, opts)
	}

	return false
}

func isParameter(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindParameter
}

func isCaughtError(node *ast.Node) bool {
	if node == nil {
		return false
	}
	// Check if the node is within a catch clause or is directly a catch variable
	parent := node.Parent
	for parent != nil {
		if parent.Kind == ast.KindCatchClause {
			return true
		}
		parent = parent.Parent
	}
	return false
}

func shouldIgnoreParameter(varName string, opts Config) bool {
	if opts.Args == "none" {
		return true
	}

	if opts.ArgsIgnorePattern != "" {
		if matched, _ := regexp.MatchString(opts.ArgsIgnorePattern, varName); matched {
			return true
		}
	}

	// For now, implement basic parameter checking
	// TODO: Implement "after-used" logic properly
	return false
}

func shouldIgnoreCaughtError(varName string, opts Config) bool {
	if opts.CaughtErrors == "none" {
		return true
	}

	if opts.CaughtErrorsIgnorePattern != "" {
		if matched, _ := regexp.MatchString(opts.CaughtErrorsIgnorePattern, varName); matched {
			return true
		}
	}

	return false
}

func isExported(varInfo *VariableInfo) bool {
	if varInfo.Variable == nil {
		return false
	}

	// Check for export modifier flags first
	if varInfo.Definition != nil {
		modifierFlags := ast.GetCombinedModifierFlags(varInfo.Definition)
		if modifierFlags&ast.ModifierFlagsExport != 0 {
			return true
		}

		// Also check parent nodes for export modifiers
		parent := varInfo.Definition.Parent
		for parent != nil {
			modifierFlags := ast.GetCombinedModifierFlags(parent)
			if modifierFlags&ast.ModifierFlagsExport != 0 {
				return true
			}
			parent = parent.Parent
		}
	}

	// Check for export declarations by looking up the AST
	parent := varInfo.Variable.Parent
	for parent != nil {
		if parent.Kind == ast.KindExportDeclaration {
			return true
		}
		parent = parent.Parent
	}

	// Also check if it's referenced in an export
	for _, ref := range varInfo.References {
		refParent := ref.Parent
		for refParent != nil {
			if refParent.Kind == ast.KindExportDeclaration {
				return true
			}
			refParent = refParent.Parent
		}
	}

	return false
}

func buildUnusedVarMessage(varName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unusedVar",
		Description: fmt.Sprintf("'%s' is defined but never used.", varName),
	}
}

func buildUsedOnlyAsTypeMessage(varName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "usedOnlyAsType",
		Description: fmt.Sprintf("'%s' is defined but only used as a type.", varName),
	}
}

func buildUsedIgnoredVarMessage(varName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "usedIgnoredVar",
		Description: fmt.Sprintf("'%s' is marked as ignored but is used.", varName),
	}
}

func collectVariableUsages(node *ast.Node, usages map[string][]*ast.Node) {
	if node == nil {
		return
	}

	// Visit current node
	if ast.IsIdentifier(node) && !isPartOfDeclaration(node) && !isPartOfAssignment(node) {
		name := node.AsIdentifier().Text
		usages[name] = append(usages[name], node)
	}

	// Recursively visit all children using ForEachChild
	node.ForEachChild(func(child *ast.Node) bool {
		collectVariableUsages(child, usages)
		return false // Continue traversing
	})
}

func processVariable(ctx rule.RuleContext, nameNode *ast.Node, name string, definition *ast.Node, opts Config, allUsages map[string][]*ast.Node) {
	// Create variable info
	varInfo := &VariableInfo{
		Variable:       nameNode,
		Used:           false,
		OnlyUsedAsType: false,
		References:     []*ast.Node{},
		Definition:     definition,
	}

	// Check if variable has a type annotation (makes it implicitly "used")
	hasTypeAnnotation := false
	if definition != nil && definition.Kind == ast.KindVariableDeclaration {
		varDecl := definition.AsVariableDeclaration()
		if varDecl.Type != nil {
			hasTypeAnnotation = true
		}
	}

	// Check if this variable is used
	if usageNodes, exists := allUsages[name]; exists {
		varInfo.References = usageNodes

		// Remove self-references (the declaration itself)
		filteredUsages := []*ast.Node{}
		for _, usage := range usageNodes {
			if usage.Pos() != varInfo.Variable.Pos() {
				filteredUsages = append(filteredUsages, usage)
			}
		}

		if len(filteredUsages) > 0 {
			// Check if only used in type context
			onlyUsedAsType := true
			for _, usage := range filteredUsages {
				if !isInTypeContext(usage) {
					onlyUsedAsType = false
					break
				}
			}
			varInfo.Used = !onlyUsedAsType
			varInfo.OnlyUsedAsType = onlyUsedAsType
		}
	}

	// If variable has type annotation, consider it used
	if hasTypeAnnotation {
		varInfo.Used = true
		varInfo.OnlyUsedAsType = false
	}

	// Check if we should report this variable
	if shouldIgnoreVariable(name, varInfo, opts) {
		return
	}

	// Skip exported variables
	if isExported(varInfo) {
		return
	}

	// Special handling for function declarations: don't report function name if it has parameters
	// The parameters will be handled separately
	if definition != nil && definition.Kind == ast.KindFunctionDeclaration {
		funcDecl := definition.AsFunctionDeclaration()
		if funcDecl.Parameters != nil && len(funcDecl.Parameters.Nodes) > 0 {
			// Function has parameters, don't report the function name itself
			// The parameter reporting will handle unused parameters
			return
		}
	}

	// Report unused variables
	if varInfo.OnlyUsedAsType && opts.Vars == "all" {
		// Variable is only used in type contexts
		ctx.ReportNode(varInfo.Variable, buildUsedOnlyAsTypeMessage(name))
	} else if !varInfo.Used {
		// Variable is not used at all
		ctx.ReportNode(varInfo.Variable, buildUnusedVarMessage(name))
	} else if varInfo.Used && opts.ReportUsedIgnorePattern {
		// Check if used but matches ignore pattern and should be reported
		if opts.VarsIgnorePattern != "" {
			if matched, _ := regexp.MatchString(opts.VarsIgnorePattern, name); matched {
				ctx.ReportNode(varInfo.Variable, buildUsedIgnoredVarMessage(name))
			}
		}
	}
}

var NoUnusedVarsRule = rule.Rule{
	Name: "no-unused-vars",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		// We need to collect all variable usages once per source file
		allUsages := make(map[string][]*ast.Node)
		collected := false

		// Helper function to get root source file node
		getRootSourceFile := func(node *ast.Node) *ast.Node {
			current := node
			for current.Parent != nil {
				current = current.Parent
			}
			return current
		}

		return rule.RuleListeners{
			// Handle variable declarations
			ast.KindVariableDeclaration: func(node *ast.Node) {
				varDecl := node.AsVariableDeclaration()
				if ast.IsIdentifier(varDecl.Name()) {
					nameNode := varDecl.Name()
					name := nameNode.AsIdentifier().Text

					// Collect usages for the entire source file on first variable
					if !collected {
						sourceFile := getRootSourceFile(node)
						collectVariableUsages(sourceFile, allUsages)
						collected = true
					}

					processVariable(ctx, nameNode, name, node, opts, allUsages)
				}
			},

			// Handle function declarations
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				funcDecl := node.AsFunctionDeclaration()
				if funcDecl.Name() != nil && ast.IsIdentifier(funcDecl.Name()) {
					nameNode := funcDecl.Name()
					name := nameNode.AsIdentifier().Text

					// Collect usages for the entire source file on first variable
					if !collected {
						sourceFile := getRootSourceFile(node)
						collectVariableUsages(sourceFile, allUsages)
						collected = true
					}

					processVariable(ctx, nameNode, name, node, opts, allUsages)
				}
			},

			// Handle function parameters
			ast.KindParameter: func(node *ast.Node) {
				paramDecl := node.AsParameterDeclaration()
				if paramDecl.Name() != nil && ast.IsIdentifier(paramDecl.Name()) {
					nameNode := paramDecl.Name()
					name := nameNode.AsIdentifier().Text

					// Collect usages for the entire source file on first variable
					if !collected {
						sourceFile := getRootSourceFile(node)
						collectVariableUsages(sourceFile, allUsages)
						collected = true
					}

					processVariable(ctx, nameNode, name, node, opts, allUsages)
				}
			},

			// Handle catch clauses
			ast.KindCatchClause: func(node *ast.Node) {
				catchClause := node.AsCatchClause()
				if catchClause.VariableDeclaration != nil && ast.IsIdentifier(catchClause.VariableDeclaration) {
					nameNode := catchClause.VariableDeclaration
					name := nameNode.AsIdentifier().Text

					// Collect usages for the entire source file on first variable
					if !collected {
						sourceFile := getRootSourceFile(node)
						collectVariableUsages(sourceFile, allUsages)
						collected = true
					}

					processVariable(ctx, nameNode, name, nameNode, opts, allUsages)
				}
			},
		}
	},
}
