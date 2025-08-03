package no_shadow

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoShadowOptions struct {
	Allow                                      []string `json:"allow"`
	BuiltinGlobals                             bool     `json:"builtinGlobals"`
	Hoist                                      string   `json:"hoist"`
	IgnoreFunctionTypeParameterNameValueShadow bool     `json:"ignoreFunctionTypeParameterNameValueShadow"`
	IgnoreOnInitialization                     bool     `json:"ignoreOnInitialization"`
	IgnoreTypeValueShadow                      bool     `json:"ignoreTypeValueShadow"`
}

type Variable struct {
	Name       string
	Node       *ast.Node
	IsType     bool
	IsValue    bool
	IsBuiltin  bool
	DeclaredAt *ast.Node
	Scope      *Scope
}

type Scope struct {
	Node      *ast.Node
	Parent    *Scope
	Variables map[string]*Variable
	Children  []*Scope
	Type      ScopeType
}

type ScopeType int

const (
	ScopeTypeGlobal ScopeType = iota
	ScopeTypeModule
	ScopeTypeFunction
	ScopeTypeBlock
	ScopeTypeClass
	ScopeTypeWith
	ScopeTypeTSModule
	ScopeTypeTSEnum
	ScopeTypeFunctionExpressionName
)

var allowedFunctionVariableDefTypes = map[ast.Kind]bool{
	ast.KindCallSignature:       true,
	ast.KindFunctionType:        true,
	ast.KindMethodSignature:     true,
	ast.KindEmptyStatement:      true, // TSEmptyBodyFunctionExpression
	ast.KindFunctionDeclaration: true, // TSDeclareFunction
	ast.KindConstructSignature:  true,
	ast.KindConstructorType:     true,
}

var functionsHoistedNodes = map[ast.Kind]bool{
	ast.KindFunctionDeclaration: true,
}

var typesHoistedNodes = map[ast.Kind]bool{
	ast.KindInterfaceDeclaration: true,
	ast.KindTypeAliasDeclaration: true,
}

var NoShadowRule = rule.Rule{
	Name: "no-shadow",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoShadowOptions{
			Allow:          []string{},
			BuiltinGlobals: false,
			Hoist:          "functions-and-types",
			IgnoreFunctionTypeParameterNameValueShadow: true,
			IgnoreOnInitialization:                     false,
			IgnoreTypeValueShadow:                      true,
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
				if val, ok := optsMap["allow"].([]interface{}); ok {
					opts.Allow = make([]string, len(val))
					for i, v := range val {
						if str, ok := v.(string); ok {
							opts.Allow[i] = str
						}
					}
				}
				if val, ok := optsMap["builtinGlobals"].(bool); ok {
					opts.BuiltinGlobals = val
				}
				if val, ok := optsMap["hoist"].(string); ok {
					opts.Hoist = val
				}
				if val, ok := optsMap["ignoreFunctionTypeParameterNameValueShadow"].(bool); ok {
					opts.IgnoreFunctionTypeParameterNameValueShadow = val
				}
				if val, ok := optsMap["ignoreOnInitialization"].(bool); ok {
					opts.IgnoreOnInitialization = val
				}
				if val, ok := optsMap["ignoreTypeValueShadow"].(bool); ok {
					opts.IgnoreTypeValueShadow = val
				}
			}
		}

		// Built-in globals
		// Nodes that are hoisted
		functionsHoistedNodes := map[ast.Kind]bool{
			ast.KindFunctionDeclaration: true,
		}

		typesHoistedNodes := map[ast.Kind]bool{
			ast.KindInterfaceDeclaration: true,
			ast.KindTypeAliasDeclaration: true,
			ast.KindEnumDeclaration:      true,
			ast.KindClassDeclaration:     true,
		}

		builtinGlobals := map[string]bool{
			"Array": true, "ArrayBuffer": true, "Atomics": true, "BigInt": true, "BigInt64Array": true,
			"BigUint64Array": true, "Boolean": true, "DataView": true, "Date": true, "Error": true,
			"EvalError": true, "Float32Array": true, "Float64Array": true, "Function": true,
			"Infinity": true, "Int16Array": true, "Int32Array": true, "Int8Array": true, "Intl": true,
			"JSON": true, "Map": true, "Math": true, "NaN": true, "Number": true, "Object": true,
			"Promise": true, "Proxy": true, "RangeError": true, "ReferenceError": true, "Reflect": true,
			"RegExp": true, "Set": true, "SharedArrayBuffer": true, "String": true, "Symbol": true,
			"SyntaxError": true, "TypeError": true, "URIError": true, "Uint16Array": true,
			"Uint32Array": true, "Uint8Array": true, "Uint8ClampedArray": true, "WeakMap": true,
			"WeakSet": true, "console": true, "decodeURI": true, "decodeURIComponent": true,
			"encodeURI": true, "encodeURIComponent": true, "escape": true, "eval": true,
			"globalThis": true, "isFinite": true, "isNaN": true, "parseFloat": true, "parseInt": true,
			"unescape": true, "undefined": true, "global": true, "window": true, "document": true,
		}

		// Track all scopes
		globalScope := &Scope{
			Node:      ctx.SourceFile.AsNode(),
			Variables: make(map[string]*Variable),
			Type:      ScopeTypeGlobal,
		}

		// Add built-in globals to global scope if enabled
		if opts.BuiltinGlobals {
			for name := range builtinGlobals {
				globalScope.Variables[name] = &Variable{
					Name:       name,
					IsBuiltin:  true,
					IsValue:    true, // Builtin globals are values
					Scope:      globalScope,
					Node:       ctx.SourceFile.AsNode(), // Use source file as a placeholder
					DeclaredAt: ctx.SourceFile.AsNode(), // Use source file as a placeholder
				}
			}
		}

		scopeStack := []*Scope{globalScope}
		getCurrentScope := func() *Scope {
			return scopeStack[len(scopeStack)-1]
		}

		pushScope := func(node *ast.Node, scopeType ScopeType) {
			newScope := &Scope{
				Node:      node,
				Parent:    getCurrentScope(),
				Variables: make(map[string]*Variable),
				Type:      scopeType,
			}
			getCurrentScope().Children = append(getCurrentScope().Children, newScope)
			scopeStack = append(scopeStack, newScope)
		}

		popScope := func() {
			if len(scopeStack) > 1 {
				scopeStack = scopeStack[:len(scopeStack)-1]
			}
		}

		// Forward declare checkVariable
		var checkVariable func(variable *Variable)

		// Check for function type parameter name value shadow
		isFunctionTypeParameterNameValueShadow := func(variable *Variable, shadowed *Variable) bool {
			if !opts.IgnoreFunctionTypeParameterNameValueShadow {
				return false
			}

			if !variable.IsValue || !shadowed.IsValue {
				return false
			}

			// Only apply to parameters
			if !ast.IsParameter(variable.Node) {
				return false
			}

			// Simple check: if the parameter's parent is an arrow function or function type,
			// and that's ultimately part of a type alias, interface, or other type context,
			// then ignore the shadow.
			node := variable.Node
			parent := node.Parent

			// Skip if no parent
			if parent == nil {
				return false
			}

			// Check if the immediate parent or grandparent indicates a type context
			for depth := 0; depth < 10 && parent != nil; depth++ {
				switch parent.Kind {
				case ast.KindFunctionType,
					ast.KindCallSignature,
					ast.KindMethodSignature,
					ast.KindConstructSignature,
					ast.KindConstructorType:
					// Direct function type contexts - definitely ignore
					return true

				case ast.KindTypeAliasDeclaration,
					ast.KindInterfaceDeclaration:
					// Type declaration contexts - ignore parameters in these
					return true

				case ast.KindFunctionDeclaration,
					ast.KindMethodDeclaration,
					ast.KindFunctionExpression,
					ast.KindConstructor,
					ast.KindGetAccessor,
					ast.KindSetAccessor:
					// Actual function implementations - don't ignore
					return false
				}

				parent = parent.Parent
			}

			return false
		}

		// Helper to add variable to current scope
		addVariable := func(name string, node *ast.Node, isType bool, isValue bool) {
			scope := getCurrentScope()
			if v, exists := scope.Variables[name]; exists {
				// Check for same-scope redeclaration
				if v.IsValue && isValue {
					// Before reporting, check if this should be ignored due to function type parameter shadowing
					if ast.IsParameter(node) && opts.IgnoreFunctionTypeParameterNameValueShadow {
						// Create a temporary variable to use the existing ignore function
						tempVariable := &Variable{
							Name:       name,
							Node:       node,
							IsType:     isType,
							IsValue:    isValue,
							DeclaredAt: node,
							Scope:      scope,
						}

						if isFunctionTypeParameterNameValueShadow(tempVariable, v) {
							// Don't report the error, just update the existing variable
							if isType {
								v.IsType = true
							}
							if isValue {
								v.IsValue = true
							}
							return
						}
					}

					// Same scope redeclaration - report as shadowing
					line, character := scanner.GetLineAndCharacterOfPosition(ctx.SourceFile, v.Node.Pos())
					ctx.ReportNode(node, rule.RuleMessage{
						Id: "noShadow",
						Description: fmt.Sprintf("'%s' is already declared in the upper scope on line %d column %d.",
							name, int(line+1), int(character+1)),
					})
				}
				// Update existing variable
				if isType {
					v.IsType = true
				}
				if isValue {
					v.IsValue = true
				}
			} else {
				variable := &Variable{
					Name:       name,
					Node:       node,
					IsType:     isType,
					IsValue:    isValue,
					DeclaredAt: node,
					Scope:      scope,
				}
				scope.Variables[name] = variable
				// Check immediately for same-scope and cross-scope shadowing
				checkVariable(variable)
			}
		}

		// Helper to find variable in outer scopes
		findVariableInOuterScopes := func(name string, currentScope *Scope) *Variable {
			scope := currentScope.Parent
			for scope != nil {
				if v, exists := scope.Variables[name]; exists {
					return v
				}
				scope = scope.Parent
			}
			return nil
		}

		// Check if scope is a TypeScript module augmenting the global namespace
		var isGlobalAugmentation func(scope *Scope) bool
		isGlobalAugmentation = func(scope *Scope) bool {
			if scope.Type == ScopeTypeTSModule && ast.IsModuleDeclaration(scope.Node) {
				moduleDecl := scope.Node.AsModuleDeclaration()
				nameNode := moduleDecl.Name()
				if ast.IsStringLiteral(nameNode) && nameNode.AsStringLiteral().Text == "global" {
					return true
				}
			}
			return scope.Parent != nil && isGlobalAugmentation(scope.Parent)
		}

		// Check if variable is a this parameter
		isThisParam := func(variable *Variable) bool {
			if !ast.IsParameter(variable.Node) {
				return false
			}
			param := variable.Node.AsParameterDeclaration()
			nameNode := param.Name()
			return ast.IsIdentifier(nameNode) && nameNode.AsIdentifier().Text == "this"
		}

		// Check if it's a type shadowing a value or vice versa
		isTypeValueShadow := func(variable *Variable, shadowed *Variable) bool {
			if !opts.IgnoreTypeValueShadow {
				return false
			}
			return variable.IsValue != shadowed.IsValue
		}

		// Check if allowed by configuration
		isAllowed := func(name string) bool {
			for _, allowed := range opts.Allow {
				if allowed == name {
					return true
				}
			}
			return false
		}

		// Check if variable is a duplicate class name in class scope
		isDuplicatedClassNameVariable := func(variable *Variable) bool {
			if !ast.IsClassDeclaration(variable.Scope.Node) {
				return false
			}
			classDecl := variable.Scope.Node.AsClassDeclaration()
			nameNode := classDecl.Name()
			return nameNode != nil && ast.IsIdentifier(nameNode) &&
				nameNode.AsIdentifier().Text == variable.Name
		}

		// Check if variable is a duplicate enum name
		isDuplicatedEnumNameVariable := func(variable *Variable) bool {
			if variable.Scope.Type != ScopeTypeTSEnum {
				return false
			}
			enumDecl := variable.Scope.Node.AsEnumDeclaration()
			nameNode := enumDecl.Name()
			return ast.IsIdentifier(nameNode) && nameNode.AsIdentifier().Text == variable.Name
		}

		// Helper to determine if a node is hoisted based on options
		isHoisted := func(node *ast.Node) bool {
			switch opts.Hoist {
			case "never":
				return false
			case "all":
				return true
			case "functions":
				return functionsHoistedNodes[node.Kind]
			case "types":
				return typesHoistedNodes[node.Kind]
			case "functions-and-types":
				return functionsHoistedNodes[node.Kind] || typesHoistedNodes[node.Kind]
			default:
				return false
			}
		}

		// Get location info for error reporting
		getDeclaredLocation := func(variable *Variable) (line int, column int, isGlobal bool) {
			if variable.IsBuiltin {
				return 0, 0, true
			}

			line, character := scanner.GetLineAndCharacterOfPosition(ctx.SourceFile, variable.Node.Pos())
			return int(line + 1), int(character + 1), false
		}

		// Process variable declarations
		processVariableDeclaration := func(node *ast.Node) {
			varDecl := node.AsVariableDeclaration()
			nameNode := varDecl.Name()
			if ast.IsIdentifier(nameNode) {
				name := nameNode.AsIdentifier().Text
				addVariable(name, node, false, true)
			}
		}

		// Process function declarations
		processFunctionDeclaration := func(node *ast.Node) {
			funcDecl := node.AsFunctionDeclaration()
			nameNode := funcDecl.Name()
			if nameNode != nil && ast.IsIdentifier(nameNode) {
				// Function declarations are added to the current scope where they are declared
				// The hoisting behavior is handled in the checkVariable function
				addVariable(nameNode.AsIdentifier().Text, node, false, true)
			}
		}

		// Process class declarations
		processClassDeclaration := func(node *ast.Node) {
			classDecl := node.AsClassDeclaration()
			nameNode := classDecl.Name()
			if nameNode != nil && ast.IsIdentifier(nameNode) {
				addVariable(nameNode.AsIdentifier().Text, node, false, true)
			}
		}

		// Process interface declarations
		processInterfaceDeclaration := func(node *ast.Node) {
			interfaceDecl := node.AsInterfaceDeclaration()
			nameNode := interfaceDecl.Name()
			if ast.IsIdentifier(nameNode) {
				addVariable(nameNode.AsIdentifier().Text, node, true, false)
			}
		}

		// Process type alias declarations
		processTypeAliasDeclaration := func(node *ast.Node) {
			typeAlias := node.AsTypeAliasDeclaration()
			nameNode := typeAlias.Name()
			if ast.IsIdentifier(nameNode) {
				addVariable(nameNode.AsIdentifier().Text, node, true, false)
			}
		}

		// Process enum declarations
		processEnumDeclaration := func(node *ast.Node) {
			enumDecl := node.AsEnumDeclaration()
			nameNode := enumDecl.Name()
			if ast.IsIdentifier(nameNode) {
				// Enums create both type and value
				addVariable(nameNode.AsIdentifier().Text, node, true, true)
			}
		}

		// Process parameters
		processParameter := func(node *ast.Node) {
			param := node.AsParameterDeclaration()
			nameNode := param.Name()
			if ast.IsIdentifier(nameNode) {
				addVariable(nameNode.AsIdentifier().Text, node, false, true)
			}
		}

		// Check variable for shadowing
		checkVariable = func(variable *Variable) {
			// Skip certain variables
			if isThisParam(variable) || isDuplicatedClassNameVariable(variable) ||
				isDuplicatedEnumNameVariable(variable) || isAllowed(variable.Name) {
				return
			}

			// Skip if in global augmentation
			if isGlobalAugmentation(variable.Scope) {
				return
			}

			// Check for same-scope redeclaration first
			for _, v := range variable.Scope.Variables {
				if v != variable && v.Name == variable.Name && v.IsValue && variable.IsValue && v.Node.Pos() < variable.Node.Pos() {
					// Same scope redeclaration - always report as shadowing
					line, character := scanner.GetLineAndCharacterOfPosition(ctx.SourceFile, v.Node.Pos())
					ctx.ReportNode(variable.Node, rule.RuleMessage{
						Id: "noShadow",
						Description: fmt.Sprintf("'%s' is already declared in the upper scope on line %d column %d.",
							variable.Name, int(line+1), int(character+1)),
					})
					return
				}
			}

			// Find shadowed variable in outer scopes
			shadowed := findVariableInOuterScopes(variable.Name, variable.Scope)
			if shadowed == nil {
				return
			}

			// Check various ignore conditions
			if isTypeValueShadow(variable, shadowed) ||
				isFunctionTypeParameterNameValueShadow(variable, shadowed) {
				return
			}

			// Handle hoisting behavior
			shadowedIsHoisted := isHoisted(shadowed.DeclaredAt)

			// Handle temporal dead zone and hoisting behavior
			if !shadowed.IsBuiltin {
				if opts.Hoist == "never" {
					// When hoist is "never", don't report shadowing of function declarations
					// This matches the behavior where function declarations are treated as non-hoisted
					if ast.IsFunctionDeclaration(shadowed.DeclaredAt) {
						return
					}
					// For non-function declarations, check TDZ only within same scope
					if !ast.IsParameter(shadowed.DeclaredAt) && variable.Scope == shadowed.Scope {
						if variable.Node.Pos() < shadowed.Node.Pos() {
							return
						}
					}
				} else {
					// With hoisting enabled, check if the shadowed variable is hoisted
					if !shadowedIsHoisted {
						// TDZ check: skip if declared after current variable
						if variable.Node.Pos() < shadowed.Node.Pos() {
							return
						}
					}
				}
			}

			// Report the error
			line, column, isGlobal := getDeclaredLocation(shadowed)

			if isGlobal {
				ctx.ReportNode(variable.Node, rule.RuleMessage{
					Id:          "noShadowGlobal",
					Description: fmt.Sprintf("'%s' is already a global variable.", variable.Name),
				})
			} else {
				ctx.ReportNode(variable.Node, rule.RuleMessage{
					Id: "noShadow",
					Description: fmt.Sprintf("'%s' is already declared in the upper scope on line %d column %d.",
						variable.Name, line, column),
				})
			}
		}

		return rule.RuleListeners{
			// Scope creators
			ast.KindSourceFile: func(node *ast.Node) {
				// Already have global scope
			},
			ast.KindBlock: func(node *ast.Node) {
				pushScope(node, ScopeTypeBlock)
			},
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				processFunctionDeclaration(node)
				pushScope(node, ScopeTypeFunction)
			},
			ast.KindFunctionExpression: func(node *ast.Node) {
				funcExpr := node.AsFunctionExpression()
				pushScope(node, ScopeTypeFunction)
				// Handle function expression name
				if funcExpr.Name() != nil && ast.IsIdentifier(funcExpr.Name()) {
					addVariable(funcExpr.Name().AsIdentifier().Text, node, false, true)
				}
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				pushScope(node, ScopeTypeFunction)
			},
			ast.KindMethodDeclaration: func(node *ast.Node) {
				pushScope(node, ScopeTypeFunction)
			},
			ast.KindConstructor: func(node *ast.Node) {
				pushScope(node, ScopeTypeFunction)
			},
			ast.KindGetAccessor: func(node *ast.Node) {
				pushScope(node, ScopeTypeFunction)
			},
			ast.KindSetAccessor: func(node *ast.Node) {
				pushScope(node, ScopeTypeFunction)
			},
			ast.KindClassDeclaration: func(node *ast.Node) {
				processClassDeclaration(node)
				pushScope(node, ScopeTypeClass)
			},
			ast.KindClassExpression: func(node *ast.Node) {
				pushScope(node, ScopeTypeClass)
				classExpr := node.AsClassExpression()
				if classExpr.Name() != nil && ast.IsIdentifier(classExpr.Name()) {
					addVariable(classExpr.Name().AsIdentifier().Text, node, false, true)
				}
			},
			ast.KindForStatement: func(node *ast.Node) {
				pushScope(node, ScopeTypeBlock)
			},
			ast.KindForInStatement: func(node *ast.Node) {
				pushScope(node, ScopeTypeBlock)
			},
			ast.KindForOfStatement: func(node *ast.Node) {
				pushScope(node, ScopeTypeBlock)
			},
			ast.KindWithStatement: func(node *ast.Node) {
				pushScope(node, ScopeTypeWith)
			},
			ast.KindCatchClause: func(node *ast.Node) {
				pushScope(node, ScopeTypeBlock)
				catch := node.AsCatchClause()
				if catch.VariableDeclaration != nil && ast.IsIdentifier(catch.VariableDeclaration) {
					addVariable(catch.VariableDeclaration.AsIdentifier().Text, catch.VariableDeclaration, false, true)
				}
			},
			ast.KindModuleDeclaration: func(node *ast.Node) {
				moduleDecl := node.AsModuleDeclaration()
				if ast.IsIdentifier(moduleDecl.Name()) {
					addVariable(moduleDecl.Name().AsIdentifier().Text, node, true, true)
				}
				pushScope(node, ScopeTypeTSModule)
			},
			ast.KindEnumDeclaration: func(node *ast.Node) {
				processEnumDeclaration(node)
				pushScope(node, ScopeTypeTSEnum)
			},

			// Variable declarations
			ast.KindVariableDeclaration:  processVariableDeclaration,
			ast.KindParameter:            processParameter,
			ast.KindInterfaceDeclaration: processInterfaceDeclaration,
			ast.KindTypeAliasDeclaration: processTypeAliasDeclaration,

			// Exit listeners for scopes (only pop scope, don't double-check)
			rule.ListenerOnExit(ast.KindBlock): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindFunctionExpression): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindMethodDeclaration): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindConstructor): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindGetAccessor): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindSetAccessor): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindClassDeclaration): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindClassExpression): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindForStatement): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindForInStatement): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindForOfStatement): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindWithStatement): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindCatchClause): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindModuleDeclaration): func(node *ast.Node) {
				popScope()
			},
			rule.ListenerOnExit(ast.KindEnumDeclaration): func(node *ast.Node) {
				popScope()
			},
		}
	},
}
