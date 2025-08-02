package no_use_before_define

import (
	"fmt"
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type Config struct {
	AllowNamedExports    bool `json:"allowNamedExports"`
	Classes              bool `json:"classes"`
	Enums                bool `json:"enums"`
	Functions            bool `json:"functions"`
	IgnoreTypeReferences bool `json:"ignoreTypeReferences"`
	Typedefs             bool `json:"typedefs"`
	Variables            bool `json:"variables"`
}

var sentinelTypeRegex = regexp.MustCompile(`^(?:(?:Function|Class)(?:Declaration|Expression)|ArrowFunctionExpression|CatchClause|ImportDeclaration|ExportNamedDeclaration)$`)

// Variable represents a variable declaration in the scope
type Variable struct {
	Name        string
	Node        *ast.Node
	Identifiers []*ast.Node
	References  []*Reference
	DefType     DefinitionType
	Scope       *Scope
}

// Reference represents a reference to a variable
type Reference struct {
	Identifier       *ast.Node
	IsTypeReference  bool
	IsValueReference bool
	From             *Scope
	Init             bool
}

// Scope represents a lexical scope
type Scope struct {
	Node          *ast.Node
	Parent        *Scope
	Children      []*Scope
	Variables     []*Variable
	References    []*Reference
	Type          ScopeType
	VariableScope *Scope // The function or global scope
}

type ScopeType int

const (
	ScopeTypeGlobal ScopeType = iota
	ScopeTypeFunction
	ScopeTypeBlock
	ScopeTypeClass
	ScopeTypeEnum
	ScopeTypeModule
	ScopeTypeFunctionType
)

type DefinitionType int

const (
	DefTypeVariable DefinitionType = iota
	DefTypeFunctionName
	DefTypeClassName
	DefTypeTSEnumName
	DefTypeType
)

func parseOptions(options interface{}) Config {
	config := Config{
		Functions:            true,
		Classes:              true,
		Enums:                true,
		Variables:            true,
		Typedefs:             true,
		IgnoreTypeReferences: true,
		AllowNamedExports:    false,
	}

	if options == nil {
		return config
	}

	// Handle string option
	if str, ok := options.(string); ok {
		if str == "nofunc" {
			config.Functions = false
		}
		return config
	}

	// Handle object options
	if optsMap, ok := options.(map[string]interface{}); ok {
		if val, ok := optsMap["functions"].(bool); ok {
			config.Functions = val
		}
		if val, ok := optsMap["classes"].(bool); ok {
			config.Classes = val
		}
		if val, ok := optsMap["enums"].(bool); ok {
			config.Enums = val
		}
		if val, ok := optsMap["variables"].(bool); ok {
			config.Variables = val
		}
		if val, ok := optsMap["typedefs"].(bool); ok {
			config.Typedefs = val
		}
		if val, ok := optsMap["ignoreTypeReferences"].(bool); ok {
			config.IgnoreTypeReferences = val
		}
		if val, ok := optsMap["allowNamedExports"].(bool); ok {
			config.AllowNamedExports = val
		}
	}

	return config
}

func isFunction(variable *Variable) bool {
	return variable.DefType == DefTypeFunctionName
}

func isTypedef(variable *Variable) bool {
	return variable.DefType == DefTypeType
}

func isOuterEnum(variable *Variable, reference *Reference) bool {
	return variable.DefType == DefTypeTSEnumName &&
		variable.Scope.VariableScope != reference.From.VariableScope
}

func isOuterClass(variable *Variable, reference *Reference) bool {
	return variable.DefType == DefTypeClassName &&
		variable.Scope.VariableScope != reference.From.VariableScope
}

func isOuterVariable(variable *Variable, reference *Reference) bool {
	return variable.DefType == DefTypeVariable &&
		variable.Scope.VariableScope != reference.From.VariableScope
}

func isNamedExports(reference *Reference) bool {
	identifier := reference.Identifier
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindExportSpecifier {
		exportSpec := identifier.Parent.AsExportSpecifier()
		// Check if identifier is either the property name (a in "export { a as b }")
		// or the name (a in "export { a }")
		return exportSpec.PropertyName == identifier || exportSpec.Name() == identifier
	}
	return false
}

func isTypeReference(reference *Reference) bool {
	if reference.IsTypeReference {
		return true
	}
	return referenceContainsTypeQuery(reference.Identifier)
}

func referenceContainsTypeQuery(node *ast.Node) bool {
	// Check if this identifier is part of a typeof expression
	parent := node.Parent
	for parent != nil {
		if parent.Kind == ast.KindTypeQuery {
			return true
		}
		// Stop traversing at certain node types
		if parent.Kind == ast.KindFunctionDeclaration ||
			parent.Kind == ast.KindFunctionExpression ||
			parent.Kind == ast.KindArrowFunction ||
			parent.Kind == ast.KindBlock {
			break
		}
		parent = parent.Parent
	}
	return false
}

func isInRange(node *ast.Node, location int) bool {
	return node != nil && node.Pos() <= location && location <= node.End()
}

func isClassRefInClassDecorator(variable *Variable, reference *Reference) bool {
	if variable.DefType != DefTypeClassName || len(variable.Identifiers) == 0 {
		return false
	}

	// Get the class declaration node
	classNode := variable.Node
	if classNode.Kind != ast.KindClassDeclaration {
		return false
	}

	classDecl := classNode.AsClassDeclaration()
	if classDecl.Modifiers() == nil {
		return false
	}

	// Check if reference is within any decorator
	for _, mod := range classDecl.Modifiers().Nodes {
		if mod.Kind == ast.KindDecorator {
			decorator := mod.AsDecorator()
			if reference.Identifier.Pos() >= decorator.Pos() &&
				reference.Identifier.End() <= decorator.End() {
				return true
			}
		}
	}

	return false
}

func isInInitializer(variable *Variable, reference *Reference) bool {
	if variable.Scope != reference.From {
		return false
	}

	if len(variable.Identifiers) == 0 {
		return false
	}

	node := variable.Identifiers[0].Parent
	location := reference.Identifier.End()

	for node != nil {
		if node.Kind == ast.KindVariableDeclaration {
			varDecl := node.AsVariableDeclaration()
			if varDecl.Initializer != nil && isInRange(varDecl.Initializer, location) {
				return true
			}
			if node.Parent != nil && node.Parent.Parent != nil {
				grandParent := node.Parent.Parent
				if grandParent.Kind == ast.KindForInStatement || grandParent.Kind == ast.KindForOfStatement {
					if grandParent.Kind == ast.KindForInStatement {
						forIn := grandParent.AsForInOrOfStatement()
						if isInRange(forIn.Expression, location) {
							return true
						}
					} else {
						forOf := grandParent.AsForInOrOfStatement()
						if isInRange(forOf.Expression, location) {
							return true
						}
					}
				}
			}
			break
		} else if node.Kind == ast.KindBindingElement {
			bindingElem := node.AsBindingElement()
			if bindingElem.Initializer != nil && isInRange(bindingElem.Initializer, location) {
				return true
			}
		} else if sentinelTypeRegex.MatchString(node.Kind.String()) {
			break
		}
		node = node.Parent
	}

	return false
}

// ScopeManager manages scopes and variables
type ScopeManager struct {
	globalScope  *Scope
	currentScope *Scope
	scopes       []*Scope
}

func newScopeManager() *ScopeManager {
	globalScope := &Scope{
		Type:          ScopeTypeGlobal,
		Variables:     []*Variable{},
		References:    []*Reference{},
		Children:      []*Scope{},
		VariableScope: nil,
	}
	globalScope.VariableScope = globalScope

	return &ScopeManager{
		globalScope:  globalScope,
		currentScope: globalScope,
		scopes:       []*Scope{globalScope},
	}
}

func (sm *ScopeManager) pushScope(node *ast.Node, scopeType ScopeType) {
	newScope := &Scope{
		Node:       node,
		Parent:     sm.currentScope,
		Type:       scopeType,
		Variables:  []*Variable{},
		References: []*Reference{},
		Children:   []*Scope{},
	}

	// Set variable scope
	if scopeType == ScopeTypeFunction || scopeType == ScopeTypeGlobal {
		newScope.VariableScope = newScope
	} else {
		newScope.VariableScope = sm.currentScope.VariableScope
	}

	sm.currentScope.Children = append(sm.currentScope.Children, newScope)
	sm.currentScope = newScope
	sm.scopes = append(sm.scopes, newScope)
}

func (sm *ScopeManager) popScope() {
	if sm.currentScope.Parent != nil {
		sm.currentScope = sm.currentScope.Parent
	}
}

func (sm *ScopeManager) addVariable(name string, node *ast.Node, defType DefinitionType) {
	variable := &Variable{
		Name:        name,
		Node:        node,
		Identifiers: []*ast.Node{node},
		References:  []*Reference{},
		DefType:     defType,
		Scope:       sm.currentScope,
	}
	sm.currentScope.Variables = append(sm.currentScope.Variables, variable)
}

func (sm *ScopeManager) addReference(identifier *ast.Node, isTypeRef bool, isInit bool) {
	ref := &Reference{
		Identifier:       identifier,
		IsTypeReference:  isTypeRef,
		IsValueReference: !isTypeRef,
		From:             sm.currentScope,
		Init:             isInit,
	}
	sm.currentScope.References = append(sm.currentScope.References, ref)
}

func (sm *ScopeManager) resolveReferences() {
	for _, scope := range sm.scopes {
		for _, ref := range scope.References {
			// Find the variable in current scope or parent scopes
			variable := sm.findVariable(ref.Identifier, ref.From)
			if variable != nil {
				variable.References = append(variable.References, ref)
			}
		}
	}
}

func (sm *ScopeManager) findVariable(identifier *ast.Node, fromScope *Scope) *Variable {
	if !ast.IsIdentifier(identifier) {
		return nil
	}

	name := identifier.AsIdentifier().Text
	scope := fromScope

	for scope != nil {
		for _, variable := range scope.Variables {
			if variable.Name == name {
				return variable
			}
		}
		scope = scope.Parent
	}

	return nil
}

func checkForEarlyReferences(scopeManager *ScopeManager, varName string, varPos int, ctx rule.RuleContext, config Config, defType DefinitionType) {
	// Look through all scopes for references to this variable that occur before its definition
	for _, scope := range scopeManager.scopes {
		for _, ref := range scope.References {
			if ref.Identifier.AsIdentifier().Text == varName && ref.Identifier.Pos() < varPos && !ref.Init {
				// Check configuration to see if this type of violation should be reported
				shouldReport := true
				if defType == DefTypeVariable && !config.Variables {
					shouldReport = false
				}
				if defType == DefTypeFunctionName && !config.Functions {
					shouldReport = false
				}
				if defType == DefTypeClassName && !config.Classes {
					shouldReport = false
				}
				if defType == DefTypeTSEnumName && !config.Enums {
					shouldReport = false
				}
				if defType == DefTypeType && !config.Typedefs {
					shouldReport = false
				}
				
				// Check if it's a type reference and should be ignored
				if config.IgnoreTypeReferences && isTypeReference(ref) {
					shouldReport = false
				}
				
				// Check if it's a named export and should be allowed
				if config.AllowNamedExports && isNamedExports(ref) {
					shouldReport = false
				}
				
				if shouldReport {
					ctx.ReportNode(ref.Identifier, rule.RuleMessage{
						Id:          "noUseBeforeDefine",
						Description: fmt.Sprintf("'%s' was used before it was defined.", varName),
					})
				}
			}
		}
	}
}

var NoUseBeforeDefineRule = rule.Rule{
	Name: "no-use-before-define",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		config := parseOptions(options)
		scopeManager := newScopeManager()

		// Helper to check if node is in a type-only context
		isInTypeContext := func(node *ast.Node) bool {
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
					// These are type contexts - assume we're in a type position
					return true
				}
				parent = parent.Parent
			}
			return false
		}

		// Helper to determine if reference is forbidden
		isForbidden := func(variable *Variable, reference *Reference) bool {
			if config.IgnoreTypeReferences && isTypeReference(reference) {
				return false
			}
			if isFunction(variable) {
				return config.Functions
			}
			if isOuterClass(variable, reference) {
				return config.Classes
			}
			if isOuterVariable(variable, reference) {
				return config.Variables
			}
			if isOuterEnum(variable, reference) {
				return config.Enums
			}
			if isTypedef(variable) {
				return config.Typedefs
			}
			return true
		}

		// Helper to check if variable is defined before use
		isDefinedBeforeUse := func(variable *Variable, reference *Reference) bool {
			if len(variable.Identifiers) == 0 {
				return false
			}
			return variable.Identifiers[0].End() <= reference.Identifier.End() &&
				!(reference.IsValueReference && isInInitializer(variable, reference))
		}

		// Check all references in a scope
		checkScope := func(scope *Scope) {
			for _, reference := range scope.References {
				// Skip initialization references
				if reference.Init {
					continue
				}

				// Handle named exports
				if isNamedExports(reference) {
					if config.AllowNamedExports {
						// Skip checking named exports when allowed
						continue
					}
					// When not allowed, check if defined before use
					variable := scopeManager.findVariable(reference.Identifier, reference.From)
					if variable == nil || !isDefinedBeforeUse(variable, reference) {
						ctx.ReportNode(reference.Identifier, rule.RuleMessage{
							Id:          "noUseBeforeDefine",
							Description: fmt.Sprintf("'%s' was used before it was defined.", reference.Identifier.AsIdentifier().Text),
						})
					}
					continue
				}

				// Find the variable
				variable := scopeManager.findVariable(reference.Identifier, reference.From)
				if variable == nil {
					continue
				}

				// Check various conditions
				if len(variable.Identifiers) == 0 ||
					isDefinedBeforeUse(variable, reference) ||
					!isForbidden(variable, reference) ||
					isClassRefInClassDecorator(variable, reference) ||
					reference.From.Type == ScopeTypeFunctionType {
					continue
				}

				// Report the error
				ctx.ReportNode(reference.Identifier, rule.RuleMessage{
					Id:          "noUseBeforeDefine",
					Description: fmt.Sprintf("'%s' was used before it was defined.", variable.Name),
				})
			}
		}

		return rule.RuleListeners{
			// Scope creators
			ast.KindSourceFile: func(node *ast.Node) {
					scopeManager.globalScope.Node = node
			},
			ast.KindBlock: func(node *ast.Node) {
				scopeManager.pushScope(node, ScopeTypeBlock)
			},
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				funcDecl := node.AsFunctionDeclaration()
				if funcDecl.Name() != nil && ast.IsIdentifier(funcDecl.Name()) {
					funcName := funcDecl.Name().AsIdentifier().Text
					scopeManager.addVariable(funcName, funcDecl.Name(), DefTypeFunctionName)
					
					// Check if this function was referenced before being defined
					checkForEarlyReferences(scopeManager, funcName, funcDecl.Name().Pos(), ctx, config, DefTypeFunctionName)
				}
				scopeManager.pushScope(node, ScopeTypeFunction)
			},
			ast.KindFunctionExpression: func(node *ast.Node) {
				scopeManager.pushScope(node, ScopeTypeFunction)
				funcExpr := node.AsFunctionExpression()
				if funcExpr.Name() != nil && ast.IsIdentifier(funcExpr.Name()) {
					scopeManager.addVariable(funcExpr.Name().AsIdentifier().Text, funcExpr.Name(), DefTypeFunctionName)
				}
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				scopeManager.pushScope(node, ScopeTypeFunction)
			},
			ast.KindClassDeclaration: func(node *ast.Node) {
				classDecl := node.AsClassDeclaration()
				if classDecl.Name() != nil && ast.IsIdentifier(classDecl.Name()) {
					className := classDecl.Name().AsIdentifier().Text
					scopeManager.addVariable(className, node, DefTypeClassName)
					
					// Check if this class was referenced before being defined
					checkForEarlyReferences(scopeManager, className, classDecl.Name().Pos(), ctx, config, DefTypeClassName)
				}
				scopeManager.pushScope(node, ScopeTypeClass)
			},
			ast.KindClassExpression: func(node *ast.Node) {
				scopeManager.pushScope(node, ScopeTypeClass)
				classExpr := node.AsClassExpression()
				if classExpr.Name() != nil && ast.IsIdentifier(classExpr.Name()) {
					scopeManager.addVariable(classExpr.Name().AsIdentifier().Text, node, DefTypeClassName)
				}
			},
			ast.KindEnumDeclaration: func(node *ast.Node) {
				enumDecl := node.AsEnumDeclaration()
				if ast.IsIdentifier(enumDecl.Name()) {
					enumName := enumDecl.Name().AsIdentifier().Text
					scopeManager.addVariable(enumName, node, DefTypeTSEnumName)
					
					// Check if this enum was referenced before being defined
					checkForEarlyReferences(scopeManager, enumName, enumDecl.Name().Pos(), ctx, config, DefTypeTSEnumName)
				}
				scopeManager.pushScope(node, ScopeTypeEnum)
			},
			ast.KindModuleDeclaration: func(node *ast.Node) {
				moduleDecl := node.AsModuleDeclaration()
				if ast.IsIdentifier(moduleDecl.Name()) {
					scopeManager.addVariable(moduleDecl.Name().AsIdentifier().Text, node, DefTypeType)
				}
				scopeManager.pushScope(node, ScopeTypeModule)
			},
			ast.KindForStatement: func(node *ast.Node) {
				scopeManager.pushScope(node, ScopeTypeBlock)
			},
			ast.KindForInStatement: func(node *ast.Node) {
				scopeManager.pushScope(node, ScopeTypeBlock)
			},
			ast.KindForOfStatement: func(node *ast.Node) {
				scopeManager.pushScope(node, ScopeTypeBlock)
			},
			ast.KindCatchClause: func(node *ast.Node) {
				scopeManager.pushScope(node, ScopeTypeBlock)
				catch := node.AsCatchClause()
				if catch.VariableDeclaration != nil && ast.IsIdentifier(catch.VariableDeclaration) {
					scopeManager.addVariable(catch.VariableDeclaration.AsIdentifier().Text, catch.VariableDeclaration, DefTypeVariable)
				}
			},
			ast.KindFunctionType: func(node *ast.Node) {
				scopeManager.pushScope(node, ScopeTypeFunctionType)
			},
			ast.KindConstructorType: func(node *ast.Node) {
				scopeManager.pushScope(node, ScopeTypeFunctionType)
			},

			// Variable declarations
			ast.KindVariableStatement: func(node *ast.Node) {
				varStmt := node.AsVariableStatement()
				if varStmt.DeclarationList != nil {
					declList := varStmt.DeclarationList.AsVariableDeclarationList()
					for _, decl := range declList.Declarations.Nodes {
						varDecl := decl.AsVariableDeclaration()
						if ast.IsIdentifier(varDecl.Name()) {
							varName := varDecl.Name().AsIdentifier().Text
							scopeManager.addVariable(varName, varDecl.Name(), DefTypeVariable)
							
							// Check if this variable was referenced before being defined
							checkForEarlyReferences(scopeManager, varName, varDecl.Name().Pos(), ctx, config, DefTypeVariable)
						}
					}
				}
			},
			ast.KindVariableDeclaration: func(node *ast.Node) {
				varDecl := node.AsVariableDeclaration()
				if ast.IsIdentifier(varDecl.Name()) {
					varName := varDecl.Name().AsIdentifier().Text
					scopeManager.addVariable(varName, varDecl.Name(), DefTypeVariable)
					
					// Don't check here since KindVariableStatement already handles it
				}
			},
			ast.KindParameter: func(node *ast.Node) {
				param := node.AsParameterDeclaration()
				if ast.IsIdentifier(param.Name()) {
					scopeManager.addVariable(param.Name().AsIdentifier().Text, param.Name(), DefTypeVariable)
				}
			},
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				interfaceDecl := node.AsInterfaceDeclaration()
				if ast.IsIdentifier(interfaceDecl.Name()) {
					scopeManager.addVariable(interfaceDecl.Name().AsIdentifier().Text, interfaceDecl.Name(), DefTypeType)
				}
			},
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				typeAlias := node.AsTypeAliasDeclaration()
				if ast.IsIdentifier(typeAlias.Name()) {
					scopeManager.addVariable(typeAlias.Name().AsIdentifier().Text, typeAlias.Name(), DefTypeType)
				}
			},

			// Identifier references
			ast.KindIdentifier: func(node *ast.Node) {
				// Skip if this is a declaration
				parent := node.Parent
				if parent != nil {
					switch parent.Kind {
					case ast.KindVariableStatement,
						ast.KindVariableDeclaration,
						ast.KindFunctionDeclaration,
						ast.KindFunctionExpression,
						ast.KindClassDeclaration,
						ast.KindClassExpression,
						ast.KindInterfaceDeclaration,
						ast.KindTypeAliasDeclaration,
						ast.KindEnumDeclaration,
						ast.KindModuleDeclaration,
						ast.KindParameter,
						ast.KindPropertyDeclaration,
						ast.KindPropertySignature,
						ast.KindMethodDeclaration,
						ast.KindMethodSignature,
						ast.KindPropertyAssignment,
						ast.KindShorthandPropertyAssignment,
						ast.KindEnumMember:
						// Check if this identifier is the name being declared
						switch parent.Kind {
						case ast.KindVariableStatement:
							// For variable statements, check if this identifier is in any declaration
							varStmt := parent.AsVariableStatement()
							if varStmt.DeclarationList != nil {
								declList := varStmt.DeclarationList.AsVariableDeclarationList()
								for _, decl := range declList.Declarations.Nodes {
									varDecl := decl.AsVariableDeclaration()
									if varDecl.Name() == node {
										return
									}
								}
							}
						case ast.KindVariableDeclaration:
							if parent.AsVariableDeclaration().Name() == node {
								return
							}
						case ast.KindFunctionDeclaration:
							if parent.AsFunctionDeclaration().Name() == node {
								return
							}
						case ast.KindFunctionExpression:
							if parent.AsFunctionExpression().Name() == node {
								return
							}
						case ast.KindClassDeclaration:
							if parent.AsClassDeclaration().Name() == node {
								return
							}
						case ast.KindClassExpression:
							if parent.AsClassExpression().Name() == node {
								return
							}
						case ast.KindInterfaceDeclaration:
							if parent.AsInterfaceDeclaration().Name() == node {
								return
							}
						case ast.KindTypeAliasDeclaration:
							if parent.AsTypeAliasDeclaration().Name() == node {
								return
							}
						case ast.KindEnumDeclaration:
							if parent.AsEnumDeclaration().Name() == node {
								return
							}
						case ast.KindModuleDeclaration:
							if parent.AsModuleDeclaration().Name() == node {
								return
							}
						case ast.KindParameter:
							if parent.AsParameterDeclaration().Name() == node {
								return
							}
						case ast.KindPropertyDeclaration:
							if parent.AsPropertyDeclaration().Name() == node {
								return
							}
						case ast.KindPropertySignature:
							if parent.AsPropertySignatureDeclaration().Name() == node {
								return
							}
						case ast.KindMethodDeclaration:
							if parent.AsMethodDeclaration().Name() == node {
								return
							}
						case ast.KindMethodSignature:
							if parent.AsMethodSignatureDeclaration().Name() == node {
								return
							}
						case ast.KindPropertyAssignment:
							if parent.AsPropertyAssignment().Name() == node {
								return
							}
						case ast.KindShorthandPropertyAssignment:
							if parent.AsShorthandPropertyAssignment().Name() == node {
								return
							}
						case ast.KindEnumMember:
							if parent.AsEnumMember().Name() == node {
								return
							}
						}
					}
				}

				// Check if it's a property access (e.g., obj.prop)
				if parent != nil && parent.Kind == ast.KindPropertyAccessExpression {
					propAccess := parent.AsPropertyAccessExpression()
					if propAccess.Name() == node {
						return // This is the property name, not a variable reference
					}
				}

				// Check if it's a type reference
				isTypeRef := isInTypeContext(node)

				// Check if it's an initialization
				isInit := false
				if parent != nil {
					if parent.Kind == ast.KindVariableDeclaration {
						isInit = parent.AsVariableDeclaration().Name() == node
					} else if parent.Kind == ast.KindVariableStatement {
						varStmt := parent.AsVariableStatement()
						if varStmt.DeclarationList != nil {
							declList := varStmt.DeclarationList.AsVariableDeclarationList()
							for _, decl := range declList.Declarations.Nodes {
								varDecl := decl.AsVariableDeclaration()
								if varDecl.Name() == node {
									isInit = true
									break
								}
							}
						}
					}
				}

				scopeManager.addReference(node, isTypeRef, isInit)
			},

			// Exit listeners for scopes
			rule.ListenerOnExit(ast.KindBlock): func(node *ast.Node) {
				scopeManager.popScope()
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
				scopeManager.popScope()
			},
			rule.ListenerOnExit(ast.KindFunctionExpression): func(node *ast.Node) {
				scopeManager.popScope()
			},
			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				scopeManager.popScope()
			},
			rule.ListenerOnExit(ast.KindClassDeclaration): func(node *ast.Node) {
				scopeManager.popScope()
			},
			rule.ListenerOnExit(ast.KindClassExpression): func(node *ast.Node) {
				scopeManager.popScope()
			},
			rule.ListenerOnExit(ast.KindEnumDeclaration): func(node *ast.Node) {
				scopeManager.popScope()
			},
			rule.ListenerOnExit(ast.KindModuleDeclaration): func(node *ast.Node) {
				scopeManager.popScope()
			},
			rule.ListenerOnExit(ast.KindForStatement): func(node *ast.Node) {
				scopeManager.popScope()
			},
			rule.ListenerOnExit(ast.KindForInStatement): func(node *ast.Node) {
				scopeManager.popScope()
			},
			rule.ListenerOnExit(ast.KindForOfStatement): func(node *ast.Node) {
				scopeManager.popScope()
			},
			rule.ListenerOnExit(ast.KindCatchClause): func(node *ast.Node) {
				scopeManager.popScope()
			},
			rule.ListenerOnExit(ast.KindFunctionType): func(node *ast.Node) {
				scopeManager.popScope()
			},
			rule.ListenerOnExit(ast.KindConstructorType): func(node *ast.Node) {
				scopeManager.popScope()
			},

			// At the end, resolve references and check all scopes
			rule.ListenerOnExit(ast.KindSourceFile): func(node *ast.Node) {
				// Resolve all references
				scopeManager.resolveReferences()


				// Check all scopes for violations
				var checkAllScopes func(scope *Scope)
				checkAllScopes = func(scope *Scope) {
					checkScope(scope)
					for _, child := range scope.Children {
						checkAllScopes(child)
					}
				}
				checkAllScopes(scopeManager.globalScope)
			},
		}
	},
}
