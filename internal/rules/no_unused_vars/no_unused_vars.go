package no_unused_vars

import (
	"fmt"
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Options for the no-unused-vars rule
type Options struct {
	Vars                           string `json:"vars,omitempty"`         // "all" | "local"
	Args                           string `json:"args,omitempty"`         // "all" | "after-used" | "none"
	CaughtErrors                   string `json:"caughtErrors,omitempty"` // "all" | "none"
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
	Variable       *ast.Node
	Used           bool
	OnlyUsedAsType bool
	References     []*ast.Node
	Definition     *ast.Node
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

		// Use global state to collect all variables and usages
		variables := make(map[string]*VariableInfo)
		usages := make(map[string][]*ast.Node)
		processed := false

		return rule.RuleListeners{
			// Collect variable declarations and process immediately
			ast.KindVariableDeclaration: func(node *ast.Node) {
				varDecl := node.AsVariableDeclaration()
				if ast.IsIdentifier(varDecl.Name()) {
					nameNode := varDecl.Name()
					name := nameNode.AsIdentifier().Text
					variables[name] = &VariableInfo{
						Variable:       nameNode,
						Used:           false,
						OnlyUsedAsType: false,
						References:     []*ast.Node{},
						Definition:     node,
					}
					
					// Process immediately after adding the variable
					if !processed {
						processed = true
						processUnusedVariables(ctx, opts, variables, usages)
					}
				}
			},
			
			// Collect function declarations and process
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				funcDecl := node.AsFunctionDeclaration()
				if funcDecl.Name() != nil && ast.IsIdentifier(funcDecl.Name()) {
					nameNode := funcDecl.Name()
					name := nameNode.AsIdentifier().Text
					variables[name] = &VariableInfo{
						Variable:       nameNode,
						Used:           false,
						OnlyUsedAsType: false,
						References:     []*ast.Node{},
						Definition:     node,
					}
				}
				
				// Process after function declarations too (for standalone functions)
				if !processed && len(variables) > 0 {
					processed = true
					processUnusedVariables(ctx, opts, variables, usages)
				}
			},
			
			// Collect parameter declarations
			ast.KindParameter: func(node *ast.Node) {
				paramDecl := node.AsParameterDeclaration()
				if ast.IsIdentifier(paramDecl.Name()) {
					nameNode := paramDecl.Name()
					name := nameNode.AsIdentifier().Text
					variables[name] = &VariableInfo{
						Variable:       nameNode,
						Used:           false,
						OnlyUsedAsType: false,
						References:     []*ast.Node{},
						Definition:     node,
					}
				}
			},
			
			// Collect catch clause variables and process
			ast.KindCatchClause: func(node *ast.Node) {
				catchClause := node.AsCatchClause()
				if catchClause.VariableDeclaration != nil {
					nameDecl := catchClause.VariableDeclaration
					if ast.IsIdentifier(nameDecl.Name()) {
						nameNode := nameDecl.Name()
						name := nameNode.AsIdentifier().Text
						variables[name] = &VariableInfo{
							Variable:       nameNode,
							Used:           false,
							OnlyUsedAsType: false,
							References:     []*ast.Node{},
							Definition:     nameDecl,
						}
					}
				}
			},
			
			// Collect identifier usages
			ast.KindIdentifier: func(node *ast.Node) {
				// Skip identifiers that are part of declarations
				if !isPartOfDeclaration(node) {
					name := node.AsIdentifier().Text
					usages[name] = append(usages[name], node)
				}
			},
			
			// Also trigger processing in blocks for cases like try/catch
			ast.KindBlock: func(node *ast.Node) {
				if !processed && len(variables) > 0 {
					processed = true
					processUnusedVariables(ctx, opts, variables, usages)
				}
			},
			
			// Process on try statements for catch clauses
			ast.KindTryStatement: func(node *ast.Node) {
				if !processed && len(variables) > 0 {
					processed = true
					processUnusedVariables(ctx, opts, variables, usages)
				}
			},
		}
	},
}

func processUnusedVariables(ctx rule.RuleContext, opts TranslatedOptions, variables map[string]*VariableInfo, usages map[string][]*ast.Node) {
	// Mark variables as used based on usages
	for varName, usageNodes := range usages {
		if varInfo, exists := variables[varName]; exists {
			varInfo.References = usageNodes
			for _, usage := range usageNodes {
				if !isTypeOnlyUsage(usage) {
					varInfo.Used = true
					break
				} else {
					varInfo.OnlyUsedAsType = true
				}
			}
		}
	}

	// Report unused variables
	for _, varInfo := range variables {
		if shouldReportVariable(ctx, opts, varInfo, variables) {
			reportUnusedVariable(ctx, opts, varInfo)
		}
	}
}

func collectAllNodesRecursive(node *ast.Node, variables map[string]*VariableInfo, usages map[string][]*ast.Node) {
	if node == nil {
		return
	}

	// Collect declarations
	switch node.Kind {
	case ast.KindVariableDeclaration:
		varDecl := node.AsVariableDeclaration()
		if ast.IsIdentifier(varDecl.Name()) {
			nameNode := varDecl.Name()
			name := nameNode.AsIdentifier().Text
			variables[name] = &VariableInfo{
				Variable:       nameNode,
				Used:           false,
				OnlyUsedAsType: false,
				References:     []*ast.Node{},
				Definition:     node,
			}
		}
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl.Name() != nil && ast.IsIdentifier(funcDecl.Name()) {
			nameNode := funcDecl.Name()
			name := nameNode.AsIdentifier().Text
			variables[name] = &VariableInfo{
				Variable:       nameNode,
				Used:           false,
				OnlyUsedAsType: false,
				References:     []*ast.Node{},
				Definition:     node,
			}
		}
	case ast.KindParameter:
		paramDecl := node.AsParameterDeclaration()
		if ast.IsIdentifier(paramDecl.Name()) {
			nameNode := paramDecl.Name()
			name := nameNode.AsIdentifier().Text
			variables[name] = &VariableInfo{
				Variable:       nameNode,
				Used:           false,
				OnlyUsedAsType: false,
				References:     []*ast.Node{},
				Definition:     node,
			}
		}
	case ast.KindCatchClause:
		catchClause := node.AsCatchClause()
		if catchClause.VariableDeclaration != nil {
			nameDecl := catchClause.VariableDeclaration
			if ast.IsIdentifier(nameDecl.Name()) {
				nameNode := nameDecl.Name()
				name := nameNode.AsIdentifier().Text
				variables[name] = &VariableInfo{
					Variable:       nameNode,
					Used:           false,
					OnlyUsedAsType: false,
					References:     []*ast.Node{},
					Definition:     nameDecl,
				}
			}
		}
	case ast.KindIdentifier:
		// Collect usage
		if !isPartOfDeclaration(node) {
			name := node.AsIdentifier().Text
			usages[name] = append(usages[name], node)
		}
	}

	// Recursively traverse all children 
	traverseAllChildren(node, func(child *ast.Node) {
		collectAllNodesRecursive(child, variables, usages)
	})
}

func traverseAllChildren(node *ast.Node, callback func(*ast.Node)) {
	if node == nil {
		return
	}

	// Traverse all possible child nodes based on the node type
	switch node.Kind {
	case ast.KindSourceFile:
		sourceFile := node.AsSourceFile()
		for _, stmt := range sourceFile.Statements.Nodes {
			callback(stmt)
		}
	case ast.KindVariableStatement:
		varStmt := node.AsVariableStatement()
		if varStmt.DeclarationList != nil {
			callback(varStmt.DeclarationList)
		}
	case ast.KindVariableDeclarationList:
		declList := node.AsVariableDeclarationList()
		for _, decl := range declList.Declarations.Nodes {
			callback(decl)
		}
	case ast.KindVariableDeclaration:
		varDecl := node.AsVariableDeclaration()
		if varDecl.Name() != nil {
			callback(varDecl.Name())
		}
		if varDecl.Initializer != nil {
			callback(varDecl.Initializer)
		}
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl.Name() != nil {
			callback(funcDecl.Name())
		}
		if funcDecl.Parameters != nil {
			for _, param := range funcDecl.Parameters.Nodes {
				callback(param)
			}
		}
		if funcDecl.Body != nil {
			callback(funcDecl.Body)
		}
	case ast.KindParameter:
		paramDecl := node.AsParameterDeclaration()
		if paramDecl.Name() != nil {
			callback(paramDecl.Name())
		}
	case ast.KindBlock:
		block := node.AsBlock()
		for _, stmt := range block.Statements.Nodes {
			callback(stmt)
		}
	case ast.KindExpressionStatement:
		exprStmt := node.AsExpressionStatement()
		if exprStmt.Expression != nil {
			callback(exprStmt.Expression)
		}
	case ast.KindCallExpression:
		callExpr := node.AsCallExpression()
		if callExpr.Expression != nil {
			callback(callExpr.Expression)
		}
		if callExpr.Arguments != nil {
			for _, arg := range callExpr.Arguments.Nodes {
				callback(arg)
			}
		}
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		if propAccess.Expression != nil {
			callback(propAccess.Expression)
		}
		if propAccess.Name() != nil {
			callback(propAccess.Name())
		}
	case ast.KindTryStatement:
		tryStmt := node.AsTryStatement()
		if tryStmt.TryBlock != nil {
			callback(tryStmt.TryBlock)
		}
		if tryStmt.CatchClause != nil {
			callback(tryStmt.CatchClause)
		}
		if tryStmt.FinallyBlock != nil {
			callback(tryStmt.FinallyBlock)
		}
	case ast.KindCatchClause:
		catchClause := node.AsCatchClause()
		if catchClause.VariableDeclaration != nil {
			callback(catchClause.VariableDeclaration)
		}
		if catchClause.Block != nil {
			callback(catchClause.Block)
		}
	case ast.KindBinaryExpression:
		binExpr := node.AsBinaryExpression()
		if binExpr.Left != nil {
			callback(binExpr.Left)
		}
		if binExpr.Right != nil {
			callback(binExpr.Right)
		}
	}
}

func collectAllNodesHelper(node *ast.Node, variables map[string]*VariableInfo, usages map[string][]*ast.Node, visited map[*ast.Node]bool) {
	if node == nil || visited[node] {
		return
	}
	visited[node] = true

	// Collect variable declarations
	switch node.Kind {
	case ast.KindVariableDeclaration:
		varDecl := node.AsVariableDeclaration()
		if ast.IsIdentifier(varDecl.Name()) {
			nameNode := varDecl.Name()
			name := nameNode.AsIdentifier().Text
			variables[name] = &VariableInfo{
				Variable:       nameNode,
				Used:           false,
				OnlyUsedAsType: false,
				References:     []*ast.Node{},
				Definition:     node,
			}
		}
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl.Name() != nil && ast.IsIdentifier(funcDecl.Name()) {
			nameNode := funcDecl.Name()
			name := nameNode.AsIdentifier().Text
			variables[name] = &VariableInfo{
				Variable:       nameNode,
				Used:           false,
				OnlyUsedAsType: false,
				References:     []*ast.Node{},
				Definition:     node,
			}
		}
	case ast.KindParameter:
		paramDecl := node.AsParameterDeclaration()
		if ast.IsIdentifier(paramDecl.Name()) {
			nameNode := paramDecl.Name()
			name := nameNode.AsIdentifier().Text
			variables[name] = &VariableInfo{
				Variable:       nameNode,
				Used:           false,
				OnlyUsedAsType: false,
				References:     []*ast.Node{},
				Definition:     node,
			}
		}
	case ast.KindCatchClause:
		catchClause := node.AsCatchClause()
		if catchClause.VariableDeclaration != nil {
			nameDecl := catchClause.VariableDeclaration
			if ast.IsIdentifier(nameDecl.Name()) {
				nameNode := nameDecl.Name()
				name := nameNode.AsIdentifier().Text
				variables[name] = &VariableInfo{
					Variable:       nameNode,
					Used:           false,
					OnlyUsedAsType: false,
					References:     []*ast.Node{},
					Definition:     nameDecl,
				}
			}
		}
	case ast.KindIdentifier:
		// Collect usage
		if !isPartOfDeclaration(node) {
			name := node.AsIdentifier().Text
			usages[name] = append(usages[name], node)
		}
	}

	// Recursively traverse children using simplified traversal
	simpleTraverseChildren(node, func(child *ast.Node) {
		collectAllNodesHelper(child, variables, usages, visited)
	})
}

func simpleTraverseChildren(node *ast.Node, callback func(*ast.Node)) {
	if node == nil {
		return
	}

	// Basic traversal for common node types
	switch node.Kind {
	case ast.KindSourceFile:
		sourceFile := node.AsSourceFile()
		for _, stmt := range sourceFile.Statements.Nodes {
			callback(stmt)
		}
	case ast.KindVariableStatement:
		varStmt := node.AsVariableStatement()
		if varStmt.DeclarationList != nil {
			callback(varStmt.DeclarationList)
		}
	case ast.KindVariableDeclarationList:
		declList := node.AsVariableDeclarationList()
		for _, decl := range declList.Declarations.Nodes {
			callback(decl)
		}
	case ast.KindVariableDeclaration:
		varDecl := node.AsVariableDeclaration()
		if varDecl.Initializer != nil {
			callback(varDecl.Initializer)
		}
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl.Parameters != nil {
			for _, param := range funcDecl.Parameters.Nodes {
				callback(param)
			}
		}
		if funcDecl.Body != nil {
			callback(funcDecl.Body)
		}
	case ast.KindBlock:
		block := node.AsBlock()
		for _, stmt := range block.Statements.Nodes {
			callback(stmt)
		}
	case ast.KindExpressionStatement:
		exprStmt := node.AsExpressionStatement()
		if exprStmt.Expression != nil {
			callback(exprStmt.Expression)
		}
	case ast.KindCallExpression:
		callExpr := node.AsCallExpression()
		if callExpr.Expression != nil {
			callback(callExpr.Expression)
		}
		if callExpr.Arguments != nil {
			for _, arg := range callExpr.Arguments.Nodes {
				callback(arg)
			}
		}
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		if propAccess.Expression != nil {
			callback(propAccess.Expression)
		}
	case ast.KindExportDeclaration:
		exportDecl := node.AsExportDeclaration()
		if exportDecl.ExportClause != nil {
			callback(exportDecl.ExportClause)
		}
	case ast.KindNamedExports:
		namedExports := node.AsNamedExports()
		for _, element := range namedExports.Elements.Nodes {
			callback(element)
		}
	case ast.KindExportSpecifier:
		exportSpec := node.AsExportSpecifier()
		if exportSpec.Name() != nil {
			callback(exportSpec.Name())
		}
	case ast.KindReturnStatement:
		returnStmt := node.AsReturnStatement()
		if returnStmt.Expression != nil {
			callback(returnStmt.Expression)
		}
	case ast.KindBinaryExpression:
		binExpr := node.AsBinaryExpression()
		if binExpr.Left != nil {
			callback(binExpr.Left)
		}
		if binExpr.Right != nil {
			callback(binExpr.Right)
		}
	}
}

func collectDeclarations(node *ast.Node, variables map[string]*VariableInfo) {
	if node == nil {
		return
	}

	switch node.Kind {
	case ast.KindVariableStatement:
		varStmt := node.AsVariableStatement()
		if varStmt.DeclarationList != nil {
			declList := varStmt.DeclarationList.AsVariableDeclarationList()
			for _, decl := range declList.Declarations.Nodes {
				varDecl := decl.AsVariableDeclaration()
				if ast.IsIdentifier(varDecl.Name()) {
					nameNode := varDecl.Name()
					name := nameNode.AsIdentifier().Text
					variables[name] = &VariableInfo{
						Variable:       nameNode,
						Used:           false,
						OnlyUsedAsType: false,
						References:     []*ast.Node{},
						Definition:     decl,
					}
				}
			}
		}
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl.Name() != nil && ast.IsIdentifier(funcDecl.Name()) {
			nameNode := funcDecl.Name()
			name := nameNode.AsIdentifier().Text
			variables[name] = &VariableInfo{
				Variable:       nameNode,
				Used:           false,
				OnlyUsedAsType: false,
				References:     []*ast.Node{},
				Definition:     node,
			}
		}
		// Handle parameters
		if funcDecl.Parameters != nil {
			for _, param := range funcDecl.Parameters.Nodes {
				paramDecl := param.AsParameterDeclaration()
				if ast.IsIdentifier(paramDecl.Name()) {
					nameNode := paramDecl.Name()
					name := nameNode.AsIdentifier().Text
					variables[name] = &VariableInfo{
						Variable:       nameNode,
						Used:           false,
						OnlyUsedAsType: false,
						References:     []*ast.Node{},
						Definition:     param,
					}
				}
			}
		}
	case ast.KindCatchClause:
		catchClause := node.AsCatchClause()
		if catchClause.VariableDeclaration != nil {
			nameDecl := catchClause.VariableDeclaration
			if ast.IsIdentifier(nameDecl.Name()) {
				nameNode := nameDecl.Name()
				name := nameNode.AsIdentifier().Text
				variables[name] = &VariableInfo{
					Variable:       nameNode,
					Used:           false,
					OnlyUsedAsType: false,
					References:     []*ast.Node{},
					Definition:     nameDecl,
				}
			}
		}
	}

	// Recursively traverse children
	traverseChildren(node, func(child *ast.Node) {
		collectDeclarations(child, variables)
	})
}

func collectUsages(node *ast.Node, usages map[string][]*ast.Node) {
	if node == nil {
		return
	}

	if node.Kind == ast.KindIdentifier {
		// Skip identifiers that are part of declarations
		if !isPartOfDeclaration(node) {
			name := node.AsIdentifier().Text
			usages[name] = append(usages[name], node)
		}
	}

	// Recursively traverse children
	traverseChildren(node, func(child *ast.Node) {
		collectUsages(child, usages)
	})
}

func traverseChildren(node *ast.Node, callback func(*ast.Node)) {
	if node == nil {
		return
	}

	// This is a simplified traversal - we traverse known child nodes
	switch node.Kind {
	case ast.KindSourceFile:
		sourceFile := node.AsSourceFile()
		for _, stmt := range sourceFile.Statements.Nodes {
			callback(stmt)
		}
	case ast.KindVariableStatement:
		varStmt := node.AsVariableStatement()
		if varStmt.DeclarationList != nil {
			callback(varStmt.DeclarationList)
		}
	case ast.KindVariableDeclarationList:
		declList := node.AsVariableDeclarationList()
		for _, decl := range declList.Declarations.Nodes {
			callback(decl)
		}
	case ast.KindVariableDeclaration:
		varDecl := node.AsVariableDeclaration()
		if varDecl.Initializer != nil {
			callback(varDecl.Initializer)
		}
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl.Body != nil {
			callback(funcDecl.Body)
		}
	case ast.KindBlock:
		block := node.AsBlock()
		for _, stmt := range block.Statements.Nodes {
			callback(stmt)
		}
	case ast.KindExpressionStatement:
		exprStmt := node.AsExpressionStatement()
		if exprStmt.Expression != nil {
			callback(exprStmt.Expression)
		}
	case ast.KindCallExpression:
		callExpr := node.AsCallExpression()
		if callExpr.Expression != nil {
			callback(callExpr.Expression)
		}
		if callExpr.Arguments != nil {
			for _, arg := range callExpr.Arguments.Nodes {
				callback(arg)
			}
		}
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		if propAccess.Expression != nil {
			callback(propAccess.Expression)
		}
	case ast.KindTryStatement:
		tryStmt := node.AsTryStatement()
		if tryStmt.TryBlock != nil {
			callback(tryStmt.TryBlock)
		}
		if tryStmt.CatchClause != nil {
			callback(tryStmt.CatchClause)
		}
	case ast.KindCatchClause:
		catchClause := node.AsCatchClause()
		if catchClause.Block != nil {
			callback(catchClause.Block)
		}
	case ast.KindArrowFunction:
		arrowFunc := node.AsArrowFunction()
		if arrowFunc.Body != nil {
			callback(arrowFunc.Body)
		}
	case ast.KindFunctionExpression:
		funcExpr := node.AsFunctionExpression()
		if funcExpr.Body != nil {
			callback(funcExpr.Body)
		}
	case ast.KindReturnStatement:
		returnStmt := node.AsReturnStatement()
		if returnStmt.Expression != nil {
			callback(returnStmt.Expression)
		}
	case ast.KindIfStatement:
		ifStmt := node.AsIfStatement()
		if ifStmt.Expression != nil {
			callback(ifStmt.Expression)
		}
		if ifStmt.ThenStatement != nil {
			callback(ifStmt.ThenStatement)
		}
		if ifStmt.ElseStatement != nil {
			callback(ifStmt.ElseStatement)
		}
	case ast.KindForStatement:
		forStmt := node.AsForStatement()
		if forStmt.Initializer != nil {
			callback(forStmt.Initializer)
		}
		if forStmt.Condition != nil {
			callback(forStmt.Condition)
		}
		if forStmt.Incrementor != nil {
			callback(forStmt.Incrementor)
		}
		if forStmt.Statement != nil {
			callback(forStmt.Statement)
		}
	case ast.KindWhileStatement:
		whileStmt := node.AsWhileStatement()
		if whileStmt.Expression != nil {
			callback(whileStmt.Expression)
		}
		if whileStmt.Statement != nil {
			callback(whileStmt.Statement)
		}
	case ast.KindBinaryExpression:
		binExpr := node.AsBinaryExpression()
		if binExpr.Left != nil {
			callback(binExpr.Left)
		}
		if binExpr.Right != nil {
			callback(binExpr.Right)
		}
	case ast.KindClassDeclaration:
		classDecl := node.AsClassDeclaration()
		if classDecl.Members != nil {
			for _, member := range classDecl.Members.Nodes {
				callback(member)
			}
		}
	case ast.KindMethodDeclaration:
		methodDecl := node.AsMethodDeclaration()
		if methodDecl.Body != nil {
			callback(methodDecl.Body)
		}
	case ast.KindObjectLiteralExpression:
		objLiteral := node.AsObjectLiteralExpression()
		if objLiteral.Properties != nil {
			for _, prop := range objLiteral.Properties.Nodes {
				callback(prop)
			}
		}
	case ast.KindPropertyAssignment:
		propAssign := node.AsPropertyAssignment()
		if propAssign.Initializer != nil {
			callback(propAssign.Initializer)
		}
	case ast.KindArrayLiteralExpression:
		arrayLiteral := node.AsArrayLiteralExpression()
		if arrayLiteral.Elements != nil {
			for _, element := range arrayLiteral.Elements.Nodes {
				if element != nil {
					callback(element)
				}
			}
		}
	}
}

func isPartOfDeclaration(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

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
	}

	return false
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

func shouldReportVariable(ctx rule.RuleContext, opts TranslatedOptions, varInfo *VariableInfo, allVariables map[string]*VariableInfo) bool {
	// Skip if variable is used
	if varInfo.Used && !opts.ReportUsedIgnorePattern {
		return false
	}

	// Skip if only used as type and it's an import
	if varInfo.OnlyUsedAsType && isImportDeclaration(varInfo.Definition) {
		return false
	}

	// Skip if variable is exported
	if isExported(varInfo) {
		return false
	}

	// Skip if function references other local variables (TypeScript-ESLint specific behavior)
	if isFunctionReferencingLocalVariables(varInfo, allVariables) {
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
		if opts.Args == "after-used" && !isAfterLastUsedParam(ctx, varInfo, allVariables) {
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
	// In a module/file context, top-level declarations are considered global
	if varInfo.Definition == nil {
		return false
	}

	// Check if the definition is at the top level of a source file
	parent := varInfo.Definition.Parent
	for parent != nil {
		switch parent.Kind {
		case ast.KindSourceFile:
			return true // This is a top-level declaration
		case ast.KindBlock, ast.KindFunctionDeclaration, ast.KindArrowFunction, ast.KindFunctionExpression:
			return false // This is inside a function or block
		}
		parent = parent.Parent
	}

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

func isExported(varInfo *VariableInfo) bool {
	if varInfo.Definition == nil {
		return false
	}

	// Check if the variable is part of an export declaration
	parent := varInfo.Definition.Parent
	for parent != nil {
		switch parent.Kind {
		case ast.KindExportDeclaration:
			return true
		case ast.KindModuleDeclaration:
			return true
		}
		parent = parent.Parent
	}

	// Check if variable is referenced in an export statement
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

func isAfterLastUsedParam(ctx rule.RuleContext, varInfo *VariableInfo, allVariables map[string]*VariableInfo) bool {
	// Check if this parameter comes after the last used parameter
	if varInfo.Definition == nil || varInfo.Definition.Kind != ast.KindParameter {
		return false
	}

	// Find the parent function
	parent := varInfo.Definition.Parent
	for parent != nil {
		if parent.Kind == ast.KindFunctionDeclaration || parent.Kind == ast.KindFunctionExpression || parent.Kind == ast.KindArrowFunction {
			break
		}
		parent = parent.Parent
	}

	if parent == nil {
		return false
	}

	// Get all parameters
	var parameters []*ast.Node
	switch parent.Kind {
	case ast.KindFunctionDeclaration:
		funcDecl := parent.AsFunctionDeclaration()
		if funcDecl.Parameters != nil {
			parameters = funcDecl.Parameters.Nodes
		}
	case ast.KindFunctionExpression:
		funcExpr := parent.AsFunctionExpression()
		if funcExpr.Parameters != nil {
			parameters = funcExpr.Parameters.Nodes
		}
	case ast.KindArrowFunction:
		arrowFunc := parent.AsArrowFunction()
		if arrowFunc.Parameters != nil {
			parameters = arrowFunc.Parameters.Nodes
		}
	}

	// Find the index of current parameter
	currentParamIndex := -1
	for i, param := range parameters {
		if param == varInfo.Definition {
			currentParamIndex = i
			break
		}
	}

	if currentParamIndex == -1 {
		return false
	}

	// Check if any parameter after this one is used
	// For "after-used" mode: only report unused parameters that come after all used parameters
	for i := currentParamIndex + 1; i < len(parameters); i++ {
		param := parameters[i]
		if param.Kind == ast.KindParameter {
			paramDecl := param.AsParameterDeclaration()
			if ast.IsIdentifier(paramDecl.Name()) {
				paramName := paramDecl.Name().AsIdentifier().Text
				if paramInfo, exists := allVariables[paramName]; exists && paramInfo.Used {
					return false // This parameter comes before a used parameter, so don't report it
				}
			}
		}
	}

	return true // This parameter comes after all used parameters, so it can be reported
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

func isFunctionReferencingLocalVariables(varInfo *VariableInfo, allVariables map[string]*VariableInfo) bool {
	// Only apply this to function declarations
	if varInfo.Definition == nil || varInfo.Definition.Kind != ast.KindFunctionDeclaration {
		return false
	}

	// Get the function body
	funcDecl := varInfo.Definition.AsFunctionDeclaration()
	if funcDecl.Body == nil {
		return false
	}

	// Check if the function body references any local variables
	referencedVars := make(map[string]bool)
	collectReferencesInNode(funcDecl.Body, referencedVars)

	// Check if any of the referenced variables are local variables in the same scope
	for refVar := range referencedVars {
		if localVarInfo, exists := allVariables[refVar]; exists {
			// Skip if it's the function itself
			if localVarInfo == varInfo {
				continue
			}
			// If it references another local variable, consider it "used"
			return true
		}
	}

	return false
}

func collectReferencesInNode(node *ast.Node, refs map[string]bool) {
	if node == nil {
		return
	}

	if node.Kind == ast.KindIdentifier {
		// Skip identifiers that are part of declarations
		if !isPartOfDeclaration(node) {
			name := node.AsIdentifier().Text
			refs[name] = true
		}
	}

	// Recursively traverse children
	simpleTraverseChildren(node, func(child *ast.Node) {
		collectReferencesInNode(child, refs)
	})
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
