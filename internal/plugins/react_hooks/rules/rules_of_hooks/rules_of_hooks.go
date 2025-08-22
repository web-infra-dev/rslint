package rules_of_hooks

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	analysis "github.com/web-infra-dev/rslint/internal/plugins/react_hooks/code_path_analysis"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Helper function to check if a name is a Hook name
func isHookName(name string) bool {
	if name == "use" {
		return true
	}
	// Match "use" followed by uppercase letter
	matched, _ := regexp.MatchString(`^use[A-Z0-9]`, name)
	return matched
}

// Helper function to check if a name is a component name (PascalCase)
func isComponentName(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Component names start with uppercase letter
	return name[0] >= 'A' && name[0] <= 'Z'
}

// Helper function to check if a function is a hook
func isHook(node *ast.Node) bool {
	if node.Kind == ast.KindIdentifier {
		return isHookName(node.Name().Text())
	} else if node.Kind == ast.KindPropertyAccessExpression {
		name := node.AsPropertyAccessExpression().Name()
		if name == nil || !isHook(name) {
			return false
		}

		expr := node.AsPropertyAccessExpression().Expression
		if expr == nil || !ast.IsIdentifier(expr) {
			return false
		}

		return isPascalCaseNameSpace(expr.AsIdentifier().Text)
	} else {
		return false
	}
}

// Helper function to get function name from AST node
func getFunctionName(node *ast.Node) string {
	// TODO: Extract function name from different node types
	// - FunctionDeclaration: node.Name
	// - ArrowFunction: check parent for variable name
	// - MethodDefinition: check method name
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		// TODO: Get name from function declaration
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl != nil && funcDecl.Name() != nil {
			// TODO: Extract text from identifier
			return ""
		}
		return ""
	case ast.KindArrowFunction:
		// TODO: Get name from parent variable declarator
		return ""
	case ast.KindMethodDeclaration:
		// TODO: Get method name
		return ""
	default:
		return ""
	}
}

// Helper function to check if node is inside a component or hook
func isInsideComponentOrHook(node *ast.Node) bool {
	// TODO: Walk up the AST to find function declarations
	// and check if any of them are components or hooks
	current := node.Parent
	for current != nil {
		if isFunctionLike(current) {
			name := getFunctionName(current)
			if name != "" && (isComponentName(name) || isHook(current)) {
				return true
			}
			// TODO: Check for React.forwardRef and React.memo callbacks
		}
		current = current.Parent
	}
	return false
}

// Helper function to check if node is a function-like construct
func isFunctionLike(node *ast.Node) bool {
	kind := node.Kind
	return kind == ast.KindFunctionDeclaration ||
		kind == ast.KindFunctionExpression ||
		kind == ast.KindArrowFunction ||
		kind == ast.KindMethodDeclaration
}

// Helper function to check if node is inside a loop
func isInsideLoop(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		kind := current.Kind
		if kind == ast.KindForStatement ||
			kind == ast.KindForInStatement ||
			kind == ast.KindForOfStatement ||
			kind == ast.KindWhileStatement ||
			kind == ast.KindDoStatement {
			return true
		}
		current = current.Parent
	}
	return false
}

// Helper function to check if node is inside a conditional
func isInsideConditional(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		kind := current.Kind
		if kind == ast.KindIfStatement ||
			kind == ast.KindConditionalExpression {
			return true
		}
		// TODO: Check for logical operators (&& || ??)
		if kind == ast.KindBinaryExpression {
			binExpr := current.AsBinaryExpression()
			if binExpr != nil {
				op := binExpr.OperatorToken.Kind
				if op == ast.KindAmpersandAmpersandToken ||
					op == ast.KindBarBarToken ||
					op == ast.KindQuestionQuestionToken {
					return true
				}
			}
		}
		current = current.Parent
	}
	return false
}

// Helper function to check if node is inside a class
func isInsideClass(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if current.Kind == ast.KindClassDeclaration ||
			current.Kind == ast.KindClassExpression {
			return true
		}
		current = current.Parent
	}
	return false
}

// Helper function to check if node is inside an async function
func isInsideAsyncFunction(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if isFunctionLike(current) {
			// TODO: Check if function has async modifier
			// This requires checking the modifiers array
			// For now, check specific function types
			if current.Kind == ast.KindFunctionDeclaration {
				funcDecl := current.AsFunctionDeclaration()
				if funcDecl != nil {
					// TODO: Check for async modifier in modifiers
					return false // placeholder
				}
			} else if current.Kind == ast.KindArrowFunction {
				// TODO: Check for async modifier
				return false // placeholder
			}
		}
		current = current.Parent
	}
	return false
}

// Helper function to check if node is inside try/catch
func isInsideTryCatch(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if current.Kind == ast.KindTryStatement ||
			current.Kind == ast.KindCatchClause {
			return true
		}
		current = current.Parent
	}
	return false
}

// Helper function to check if identifier is "use"
func isUseIdentifier(node *ast.Node) bool {
	if node.Kind != ast.KindIdentifier {
		return false
	}
	// TODO: Get text of identifier and check if it's "use"
	return false // placeholder
}

// Helper function to check if call expression is a hook call
func isHookCall(node *ast.Node) (bool, string) {
	if node.Kind != ast.KindCallExpression {
		return false, ""
	}

	callExpr := node.AsCallExpression()
	if callExpr == nil {
		return false, ""
	}

	// Get the callee and extract the hook name
	// Handle different call patterns:
	// - useHook()
	// - React.useHook()
	// - obj.useHook()
	callee := callExpr.Expression
	if callee == nil {
		return false, ""
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		// Direct call: useHook()
		identifier := callee.AsIdentifier()
		if identifier != nil {
			name := scanner.GetTextOfNode(&identifier.Node)
			if isHookName(name) {
				return true, name
			}
		}
	case ast.KindPropertyAccessExpression:
		// Property access: React.useHook(), obj.useHook()
		propAccess := callee.AsPropertyAccessExpression()
		if propAccess != nil {
			nameNode := propAccess.Name()
			if nameNode != nil {
				name := scanner.GetTextOfNode(nameNode)
				if isHookName(name) {
					return true, name
				}
			}
		}
	}

	return false, ""
}

// Helper function to check if node is at top level
func isTopLevel(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if isFunctionLike(current) {
			return false
		}
		current = current.Parent
	}
	return true
}

var RulesOfHooksRule = rule.Rule{
	Name: "react-hooks/rules-of-hooks",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		codePathReactHooksMapStack := make([]map[*analysis.CodePathSegment][]*ast.Node, 0)
		codePathSegmentStack := make([]*analysis.CodePathSegment, 0)
		onCodePathSegmentStart := func(segment *analysis.CodePathSegment, node *ast.Node) {
			codePathSegmentStack = append(codePathSegmentStack, segment)
		}
		onCodePathSegmentEnd := func(segment *analysis.CodePathSegment, node *ast.Node) {
			codePathSegmentStack = codePathSegmentStack[:len(codePathSegmentStack)-1]
		}
		onCodePathStart := func(codePath *analysis.CodePath, node *ast.Node) {
			codePathReactHooksMapStack = append(
				codePathReactHooksMapStack,
				make(map[*analysis.CodePathSegment][]*ast.Node),
			)
		}
		onCodePathEnd := func(codePath *analysis.CodePath, node *ast.Node) {

		}
		analyzer := analysis.NewCodePathAnalyzer(
			onCodePathSegmentStart,
			onCodePathSegmentEnd,
			onCodePathStart,
			onCodePathEnd,
			nil, /*onCodePathSegmentLoop*/
		)
		return rule.RuleListeners{
			rule.WildcardTokenKind: func(node *ast.Node) {
				analyzer.EnterNode(node)
			},
			rule.WildcardExitTokenKind: func(node *ast.Node) {
				analyzer.LeaveNode(node)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				// !!!
			},
		}
	},
}

// Message functions for different error types
func buildConditionalHookMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "conditionalHook",
		Description: `React Hook "` + hookName + `" is called conditionally. React Hooks must be ` +
			"called in the exact same order in every component render.",
	}
}

func buildConditionalHookWithEarlyReturnMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "conditionalHook",
		Description: `React Hook "` + hookName + `" is called conditionally. React Hooks must be ` +
			"called in the exact same order in every component render." +
			" Did you accidentally call a React Hook after an early return?",
	}
}

func buildLoopHookMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "loopHook",
		Description: `React Hook "` + hookName + `" may be executed more than once. Possibly ` +
			"because it is called in a loop. React Hooks must be called in the " +
			"exact same order in every component render.",
	}
}

func buildFunctionHookMessage(hookName, functionName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "functionHook",
		Description: `React Hook "` + hookName + `" is called in function "` + functionName + `" that is neither ` +
			"a React function component nor a custom React Hook function." +
			" React component names must start with an uppercase letter." +
			" React Hook names must start with the word \"use\".",
	}
}

func buildGenericHookMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "genericHook",
		Description: `React Hook "` + hookName + `" cannot be called inside a callback. React Hooks ` +
			"must be called in a React function component or a custom React " +
			"Hook function.",
	}
}

func buildTopLevelHookMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "topLevelHook",
		Description: `React Hook "` + hookName + `" cannot be called at the top level. React Hooks ` +
			"must be called in a React function component or a custom React " +
			"Hook function.",
	}
}

func buildClassHookMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "classHook",
		Description: `React Hook "` + hookName + `" cannot be called in a class component. React Hooks ` +
			"must be called in a React function component or a custom React " +
			"Hook function.",
	}
}

func buildAsyncComponentHookMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "asyncComponentHook",
		Description: `React Hook "` + hookName + `" cannot be called in an async function.`,
	}
}

func buildTryCatchUseMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tryCatchUse",
		Description: `React Hook "` + hookName + `" cannot be called inside a try/catch block.`,
	}
}

// Helper function to check if a name follows PascalCase convention
func isPascalCaseNameSpace(name string) bool {
	if len(name) == 0 {
		return false
	}
	// PascalCase names start with uppercase letter
	return name[0] >= 'A' && name[0] <= 'Z'
}
