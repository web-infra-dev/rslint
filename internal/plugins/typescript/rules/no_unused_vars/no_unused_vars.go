package no_unused_vars

import (
	"fmt"
	"regexp"
	"strings"

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

	varsIgnoreRe         *regexp.Regexp
	argsIgnoreRe         *regexp.Regexp
	caughtErrorsIgnoreRe *regexp.Regexp
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
		return compilePatterns(config)
	}

	if arr, ok := options.([]interface{}); ok {
		if len(arr) > 0 {
			if m, ok := arr[0].(map[string]interface{}); ok {
				return compilePatterns(parseOptionsFromMap(m, config))
			}
		}
		return compilePatterns(config)
	}

	if optsMap, ok := options.(map[string]interface{}); ok {
		return compilePatterns(parseOptionsFromMap(optsMap, config))
	}

	return compilePatterns(config)
}

func parseOptionsFromMap(optsMap map[string]interface{}, config Config) Config {
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
	return config
}

func compilePatterns(config Config) Config {
	if config.VarsIgnorePattern != "" {
		config.varsIgnoreRe, _ = regexp.Compile(config.VarsIgnorePattern)
	}
	if config.ArgsIgnorePattern != "" {
		config.argsIgnoreRe, _ = regexp.Compile(config.ArgsIgnorePattern)
	}
	if config.CaughtErrorsIgnorePattern != "" {
		config.caughtErrorsIgnoreRe, _ = regexp.Compile(config.CaughtErrorsIgnorePattern)
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
		// Note: KindAsExpression, KindTypeAssertionExpression, KindSatisfiesExpression
		// are NOT included here. Their expression operand is a value context;
		// only the type annotation part is a type context. Since we walk up
		// from the identifier, a value operand will pass through these nodes
		// and continue upward without being misclassified as type-only.
		}
		parent = parent.Parent
	}
	return false
}

func isDeclarationName(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		varDecl := parent.AsVariableDeclaration()
		return varDecl != nil && varDecl.Name() == node
	case ast.KindFunctionDeclaration:
		funcDecl := parent.AsFunctionDeclaration()
		return funcDecl != nil && funcDecl.Name() == node
	case ast.KindParameter:
		paramDecl := parent.AsParameterDeclaration()
		return paramDecl != nil && paramDecl.Name() == node
	case ast.KindClassDeclaration:
		classDecl := parent.AsClassDeclaration()
		return classDecl != nil && classDecl.Name() == node
	case ast.KindInterfaceDeclaration:
		interfaceDecl := parent.AsInterfaceDeclaration()
		return interfaceDecl != nil && interfaceDecl.Name() == node
	case ast.KindTypeAliasDeclaration:
		typeAlias := parent.AsTypeAliasDeclaration()
		return typeAlias != nil && typeAlias.Name() == node
	case ast.KindEnumDeclaration:
		enumDecl := parent.AsEnumDeclaration()
		return enumDecl != nil && enumDecl.Name() == node
	case ast.KindModuleDeclaration:
		moduleDecl := parent.AsModuleDeclaration()
		return moduleDecl != nil && moduleDecl.Name() == node
	case ast.KindCatchClause:
		catchClause := parent.AsCatchClause()
		return catchClause != nil && catchClause.VariableDeclaration == node
	case ast.KindImportSpecifier:
		importSpec := parent.AsImportSpecifier()
		return importSpec != nil && importSpec.Name() == node
	case ast.KindImportClause:
		importClause := parent.AsImportClause()
		return importClause != nil && importClause.Name() == node
	case ast.KindBindingElement:
		bindingElem := parent.AsBindingElement()
		return bindingElem != nil && bindingElem.Name() == node
	case ast.KindNamespaceImport:
		nsImport := parent.AsNamespaceImport()
		return nsImport != nil && nsImport.Name() == node
	case ast.KindImportEqualsDeclaration:
		importEquals := parent.AsImportEqualsDeclaration()
		return importEquals != nil && importEquals.Name() == node
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
		if binaryExpr == nil {
			return false
		}
		if binaryExpr.OperatorToken.Kind == ast.KindEqualsToken && binaryExpr.Left == node {
			return true
		}
	}
	return false
}

func isParameterInWithoutBodyDeclaration(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent
	switch parent.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindMethodDeclaration,
		ast.KindMethodSignature,
		ast.KindConstructor:
		return parent.Body() == nil
	// Type-level function-like constructs never have a body.
	// Parameters in these are part of type signatures.
	case ast.KindFunctionType,
		ast.KindConstructorType,
		ast.KindCallSignature,
		ast.KindConstructSignature,
		ast.KindIndexSignature:
		return true
	}
	return false
}

func isInsideModuleBlock(node *ast.Node) bool {
	if node == nil {
		return false
	}
	parent := node.Parent
	for parent != nil {
		if parent.Kind == ast.KindModuleBlock {
			return true
		}
		if parent.Kind == ast.KindSourceFile {
			return false
		}
		parent = parent.Parent
	}
	return false
}

func isTopLevelDeclaration(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindVariableDeclaration,
		ast.KindFunctionDeclaration,
		ast.KindClassDeclaration,
		ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindEnumDeclaration,
		ast.KindModuleDeclaration:
		return true
	}
	return false
}

func isParameterNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	current := node
	for current != nil {
		if current.Kind == ast.KindParameter {
			return true
		}
		current = current.Parent
	}
	return false
}

func isCaughtErrorNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	parent := node.Parent
	for parent != nil {
		if parent.Kind == ast.KindCatchClause {
			return true
		}
		parent = parent.Parent
	}
	return false
}

// matchesIgnorePattern checks if a variable name matches its category's
// ignore pattern, and whether the match should result in ignoring or
// reporting (when reportUsedIgnorePattern is true and the variable is used).
// Returns: (shouldIgnore bool, matchesPattern bool)
func matchesIgnorePattern(varName string, varInfo *VariableInfo, opts Config) (bool, bool) {
	var re *regexp.Regexp

	if isParameterNode(varInfo.Definition) {
		if opts.Args == "none" {
			return true, false
		}
		re = opts.argsIgnoreRe
	} else if isCaughtErrorNode(varInfo.Definition) {
		if opts.CaughtErrors == "none" {
			return true, false
		}
		re = opts.caughtErrorsIgnoreRe
	} else {
		re = opts.varsIgnoreRe
	}

	if re == nil || !re.MatchString(varName) {
		return false, false
	}

	// Pattern matches. If used + reportUsedIgnorePattern, don't ignore — report instead.
	if varInfo.Used && opts.ReportUsedIgnorePattern {
		return false, true
	}

	return true, true
}

func isBeforeLastUsedParam(ctx rule.RuleContext, paramNode *ast.Node, allUsages map[*ast.Symbol][]*ast.Node) bool {
	if paramNode == nil || paramNode.Parent == nil {
		return false
	}

	funcLike := paramNode.Parent
	params := funcLike.Parameters()
	if len(params) == 0 {
		return false
	}

	paramIndex := -1
	for i, p := range params {
		if p.AsNode() == paramNode {
			paramIndex = i
			break
		}
	}
	if paramIndex < 0 {
		return false
	}

	for i := paramIndex + 1; i < len(params); i++ {
		sibling := params[i]

		// A parameter with a default value (initializer) counts as a
		// meaningful position marker. ESLint's after-used skips params
		// before a later param that has a default value.
		if sibling.AsNode().Initializer() != nil {
			return true
		}

		siblingName := sibling.Name()
		if siblingName == nil || !ast.IsIdentifier(siblingName) {
			continue
		}
		sym := ctx.TypeChecker.GetSymbolAtLocation(siblingName)
		if sym == nil {
			continue
		}
		resolved := ctx.TypeChecker.SkipAlias(sym)
		if usageNodes, exists := allUsages[resolved]; exists {
			for _, usage := range usageNodes {
				if usage.Pos() != siblingName.Pos() {
					return true
				}
			}
		}
	}

	return false
}

func isExported(varInfo *VariableInfo) bool {
	if varInfo.Variable == nil {
		return false
	}

	if varInfo.Definition != nil {
		modifierFlags := ast.GetCombinedModifierFlags(varInfo.Definition)
		if modifierFlags&ast.ModifierFlagsExport != 0 {
			return true
		}

		if isTopLevelDeclaration(varInfo.Definition) {
			parent := varInfo.Definition.Parent
			for parent != nil {
				switch parent.Kind {
				case ast.KindVariableDeclarationList, ast.KindVariableStatement:
					modifierFlags := ast.GetCombinedModifierFlags(parent)
					if modifierFlags&ast.ModifierFlagsExport != 0 {
						return true
					}
				case ast.KindSourceFile, ast.KindModuleBlock:
					goto doneParentWalk
				}
				parent = parent.Parent
			}
		}
	}
doneParentWalk:

	parent := varInfo.Variable.Parent
	for parent != nil {
		if parent.Kind == ast.KindExportDeclaration {
			return true
		}
		parent = parent.Parent
	}

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

func collectSymbolUsages(ctx rule.RuleContext, sourceFile *ast.Node, usages map[*ast.Symbol][]*ast.Node) {
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		if ast.IsIdentifier(node) && !isDeclarationName(node) && !isPartOfAssignment(node) {
			sym := ctx.TypeChecker.GetSymbolAtLocation(node)
			if sym != nil {
				resolved := ctx.TypeChecker.SkipAlias(sym)
				usages[resolved] = append(usages[resolved], node)
			}
			// For shorthand properties like { stats }, the identifier serves
			// as both the property name and the value reference. GetSymbolAtLocation
			// returns the property symbol, but we also need the value symbol to
			// track usage of the referenced variable.
			if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
				valSym := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(node.Parent)
				if valSym != nil {
					resolved := ctx.TypeChecker.SkipAlias(valSym)
					usages[resolved] = append(usages[resolved], node)
				}
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(sourceFile)
}

func processVariable(ctx rule.RuleContext, nameNode *ast.Node, name string, definition *ast.Node, opts Config, allUsages map[*ast.Symbol][]*ast.Node) {
	varInfo := &VariableInfo{
		Variable:       nameNode,
		Used:           false,
		OnlyUsedAsType: false,
		References:     []*ast.Node{},
		Definition:     definition,
	}

	sym := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
	if sym != nil {
		resolved := ctx.TypeChecker.SkipAlias(sym)
		if usageNodes, exists := allUsages[resolved]; exists {
			varInfo.References = usageNodes

			filteredUsages := []*ast.Node{}
			for _, usage := range usageNodes {
				if usage.Pos() != varInfo.Variable.Pos() {
					filteredUsages = append(filteredUsages, usage)
				}
			}

			if len(filteredUsages) > 0 {
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
	}

	// Check ignore patterns (varsIgnorePattern / argsIgnorePattern / caughtErrorsIgnorePattern).
	// If the variable matches its category's pattern and is unused → ignore silently.
	// If it matches but IS used and reportUsedIgnorePattern is true → report as usedIgnoredVar.
	shouldIgnore, matchedPattern := matchesIgnorePattern(name, varInfo, opts)
	if shouldIgnore {
		return
	}

	if isExported(varInfo) {
		return
	}

	// "after-used" for parameters: skip unused params before the last used param.
	// Only applies to direct Parameter nodes, not destructured elements within them.
	if !varInfo.Used && definition != nil && definition.Kind == ast.KindParameter && opts.Args == "after-used" {
		if isBeforeLastUsedParam(ctx, definition, allUsages) {
			return
		}
	}

	// For type-level declarations and imports, being used in a type context
	// IS valid usage — don't report "only used as type".
	// The "usedOnlyAsType" message only applies to value declarations
	// (const, let, function) that are never referenced at runtime.
	isTypeOrImportDeclaration := definition != nil && (definition.Kind == ast.KindInterfaceDeclaration ||
		definition.Kind == ast.KindTypeAliasDeclaration ||
		definition.Kind == ast.KindEnumDeclaration ||
		definition.Kind == ast.KindImportSpecifier ||
		definition.Kind == ast.KindImportClause ||
		definition.Kind == ast.KindNamespaceImport ||
		definition.Kind == ast.KindImportEqualsDeclaration)
	if isTypeOrImportDeclaration && varInfo.OnlyUsedAsType {
		varInfo.Used = true
		varInfo.OnlyUsedAsType = false
	}

	if matchedPattern && varInfo.Used && opts.ReportUsedIgnorePattern {
		ctx.ReportNode(varInfo.Variable, buildUsedIgnoredVarMessage(name))
	} else if varInfo.OnlyUsedAsType && opts.Vars == "all" {
		ctx.ReportNode(varInfo.Variable, buildUsedOnlyAsTypeMessage(name))
	} else if !varInfo.Used {
		ctx.ReportNode(varInfo.Variable, buildUnusedVarMessage(name))
	}
}

var NoUnusedVarsRule = rule.CreateRule(rule.Rule{
	Name: "no-unused-vars",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// .d.ts files: all declarations are ambient type definitions,
		// ESLint does not check them for unused variables.
		if strings.HasSuffix(ctx.SourceFile.FileName(), ".d.ts") {
			return rule.RuleListeners{}
		}

		opts := parseOptions(options)

		allUsages := make(map[*ast.Symbol][]*ast.Node)
		collected := false

		seenWithoutBodyFuncSymbols := make(map[*ast.Symbol]bool)

		getRootSourceFile := func(node *ast.Node) *ast.Node {
			current := node
			for current.Parent != nil {
				current = current.Parent
			}
			return current
		}

		ensureCollected := func(node *ast.Node) {
			if !collected {
				sourceFile := getRootSourceFile(node)
				collectSymbolUsages(ctx, sourceFile, allUsages)
				collected = true
			}
		}

		// processBindingName handles both simple identifiers and destructuring patterns.
		var processBindingName func(nameNode *ast.Node, definition *ast.Node)
		processBindingName = func(nameNode *ast.Node, definition *ast.Node) {
			if nameNode == nil {
				return
			}
			if ast.IsIdentifier(nameNode) {
				identifier := nameNode.AsIdentifier()
				if identifier == nil {
					return
				}
				ensureCollected(definition)
				processVariable(ctx, nameNode, identifier.Text, definition, opts, allUsages)
			} else if nameNode.Kind == ast.KindObjectBindingPattern || nameNode.Kind == ast.KindArrayBindingPattern {
				hasRestSibling := false
				if opts.IgnoreRestSiblings {
					nameNode.ForEachChild(func(child *ast.Node) bool {
						if child.Kind == ast.KindBindingElement {
							elem := child.AsBindingElement()
							if elem != nil && elem.DotDotDotToken != nil {
								hasRestSibling = true
								return true
							}
						}
						return false
					})
				}
				nameNode.ForEachChild(func(child *ast.Node) bool {
					if child.Kind == ast.KindBindingElement {
						elem := child.AsBindingElement()
						if elem != nil && elem.Name() != nil {
							if hasRestSibling && elem.DotDotDotToken == nil {
								return false
							}
							processBindingName(elem.Name(), child)
						}
					}
					return false
				})
			}
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				varDecl := node.AsVariableDeclaration()
				if varDecl == nil {
					return
				}
				if isInsideModuleBlock(node) {
					return
				}
				processBindingName(varDecl.Name(), node)
			},

			ast.KindFunctionDeclaration: func(node *ast.Node) {
				funcDecl := node.AsFunctionDeclaration()
				if funcDecl == nil {
					return
				}
				if funcDecl.Name() == nil || !ast.IsIdentifier(funcDecl.Name()) {
					return
				}

				nameNode := funcDecl.Name()
				identifier := nameNode.AsIdentifier()
				if identifier == nil {
					return
				}

				ensureCollected(node)

				if node.Body() == nil {
					if isInsideModuleBlock(node) {
						return
					}
					sym := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
					if sym != nil {
						resolved := ctx.TypeChecker.SkipAlias(sym)
						if seenWithoutBodyFuncSymbols[resolved] {
							return
						}
						seenWithoutBodyFuncSymbols[resolved] = true
					}
				}

				processVariable(ctx, nameNode, identifier.Text, node, opts, allUsages)
			},

			ast.KindModuleDeclaration: func(node *ast.Node) {
				moduleDecl := node.AsModuleDeclaration()
				if moduleDecl == nil {
					return
				}

				// Skip global scope augmentations: `declare global { ... }`
				if ast.IsGlobalScopeAugmentation(node) {
					return
				}

				nameNode := moduleDecl.Name()
				if nameNode == nil {
					return
				}
				// Skip module augmentations: `declare module 'foo' { ... }`
				if nameNode.Kind == ast.KindStringLiteral {
					return
				}
				if !ast.IsIdentifier(nameNode) {
					return
				}
				identifier := nameNode.AsIdentifier()
				if identifier == nil {
					return
				}

				// Skip namespace augmentations — if the namespace symbol has
				// declarations outside this file, it's augmenting an existing
				// namespace (e.g., `declare namespace NodeJS { ... }`).
				sym := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
				if sym != nil && len(sym.Declarations) > 1 {
					for _, decl := range sym.Declarations {
						if decl != node {
							return
						}
					}
				}

				ensureCollected(node)
				processVariable(ctx, nameNode, identifier.Text, node, opts, allUsages)
			},

			ast.KindClassDeclaration: func(node *ast.Node) {
				classDecl := node.AsClassDeclaration()
				if classDecl == nil || classDecl.Name() == nil || !ast.IsIdentifier(classDecl.Name()) {
					return
				}
				if isInsideModuleBlock(node) {
					return
				}
				nameNode := classDecl.Name()
				identifier := nameNode.AsIdentifier()
				if identifier == nil {
					return
				}
				ensureCollected(node)
				processVariable(ctx, nameNode, identifier.Text, node, opts, allUsages)
			},

			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				interfaceDecl := node.AsInterfaceDeclaration()
				if interfaceDecl == nil || interfaceDecl.Name() == nil || !ast.IsIdentifier(interfaceDecl.Name()) {
					return
				}
				if isInsideModuleBlock(node) {
					return
				}
				nameNode := interfaceDecl.Name()
				identifier := nameNode.AsIdentifier()
				if identifier == nil {
					return
				}
				ensureCollected(node)
				processVariable(ctx, nameNode, identifier.Text, node, opts, allUsages)
			},

			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				typeAlias := node.AsTypeAliasDeclaration()
				if typeAlias == nil || typeAlias.Name() == nil || !ast.IsIdentifier(typeAlias.Name()) {
					return
				}
				if isInsideModuleBlock(node) {
					return
				}
				nameNode := typeAlias.Name()
				identifier := nameNode.AsIdentifier()
				if identifier == nil {
					return
				}
				ensureCollected(node)
				processVariable(ctx, nameNode, identifier.Text, node, opts, allUsages)
			},

			ast.KindEnumDeclaration: func(node *ast.Node) {
				enumDecl := node.AsEnumDeclaration()
				if enumDecl == nil || enumDecl.Name() == nil || !ast.IsIdentifier(enumDecl.Name()) {
					return
				}
				if isInsideModuleBlock(node) {
					return
				}
				nameNode := enumDecl.Name()
				identifier := nameNode.AsIdentifier()
				if identifier == nil {
					return
				}
				ensureCollected(node)
				processVariable(ctx, nameNode, identifier.Text, node, opts, allUsages)
			},

			ast.KindParameter: func(node *ast.Node) {
				paramDecl := node.AsParameterDeclaration()
				if paramDecl == nil {
					return
				}

				if isParameterInWithoutBodyDeclaration(node) {
					return
				}

				if paramDecl.Name() != nil {
					processBindingName(paramDecl.Name(), node)
				}
			},

			ast.KindCatchClause: func(node *ast.Node) {
				catchClause := node.AsCatchClause()
				if catchClause == nil {
					return
				}
				if catchClause.VariableDeclaration != nil && ast.IsIdentifier(catchClause.VariableDeclaration) {
					nameNode := catchClause.VariableDeclaration
					identifier := nameNode.AsIdentifier()
					if identifier == nil {
						return
					}
					ensureCollected(node)
					processVariable(ctx, nameNode, identifier.Text, nameNode, opts, allUsages)
				}
			},
			ast.KindImportSpecifier: func(node *ast.Node) {
				importSpec := node.AsImportSpecifier()
				if importSpec == nil {
					return
				}
				nameNode := importSpec.Name()
				if nameNode == nil || !ast.IsIdentifier(nameNode) {
					return
				}
				identifier := nameNode.AsIdentifier()
				if identifier == nil {
					return
				}
				ensureCollected(node)
				processVariable(ctx, nameNode, identifier.Text, node, opts, allUsages)
			},

			ast.KindImportClause: func(node *ast.Node) {
				// Default import: `import Foo from './foo'`
				importClause := node.AsImportClause()
				if importClause == nil {
					return
				}
				nameNode := importClause.Name()
				if nameNode == nil || !ast.IsIdentifier(nameNode) {
					return
				}
				identifier := nameNode.AsIdentifier()
				if identifier == nil {
					return
				}
				ensureCollected(node)
				processVariable(ctx, nameNode, identifier.Text, node, opts, allUsages)
			},

			ast.KindNamespaceImport: func(node *ast.Node) {
				// Namespace import: `import * as ns from './foo'`
				nsImport := node.AsNamespaceImport()
				if nsImport == nil {
					return
				}
				nameNode := nsImport.Name()
				if nameNode == nil || !ast.IsIdentifier(nameNode) {
					return
				}
				identifier := nameNode.AsIdentifier()
				if identifier == nil {
					return
				}
				ensureCollected(node)
				processVariable(ctx, nameNode, identifier.Text, node, opts, allUsages)
			},

			ast.KindImportEqualsDeclaration: func(node *ast.Node) {
				// `import X = require('foo')` or `import X = Namespace.Y`
				importEquals := node.AsImportEqualsDeclaration()
				if importEquals == nil {
					return
				}
				nameNode := importEquals.Name()
				if nameNode == nil || !ast.IsIdentifier(nameNode) {
					return
				}
				identifier := nameNode.AsIdentifier()
				if identifier == nil {
					return
				}
				ensureCollected(node)
				processVariable(ctx, nameNode, identifier.Text, node, opts, allUsages)
			},
		}
	},
})
