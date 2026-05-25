package explicit_function_return_type

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/typescriptutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type options struct {
	allowConciseArrowFunctionExpressionsStartingWithVoid bool
	allowDirectConstAssertionInArrowFunctions            bool
	allowedNames                                         []string
	allowExpressions                                     bool
	allowFunctionsWithoutTypeParameters                  bool
	allowHigherOrderFunctions                            bool
	allowIIFEs                                           bool
	allowTypedFunctionExpressions                        bool
}

func parseOptions(rawOpts any) options {
	opts := options{
		allowConciseArrowFunctionExpressionsStartingWithVoid: false,
		allowDirectConstAssertionInArrowFunctions:            true,
		allowedNames:                        nil,
		allowExpressions:                    false,
		allowFunctionsWithoutTypeParameters: false,
		allowHigherOrderFunctions:           true,
		allowIIFEs:                          false,
		allowTypedFunctionExpressions:       true,
	}

	optsMap := utils.GetOptionsMap(rawOpts)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["allowConciseArrowFunctionExpressionsStartingWithVoid"].(bool); ok {
		opts.allowConciseArrowFunctionExpressionsStartingWithVoid = v
	}
	if v, ok := optsMap["allowDirectConstAssertionInArrowFunctions"].(bool); ok {
		opts.allowDirectConstAssertionInArrowFunctions = v
	}
	if v, ok := optsMap["allowedNames"].([]interface{}); ok {
		for _, name := range v {
			if s, ok := name.(string); ok {
				opts.allowedNames = append(opts.allowedNames, s)
			}
		}
	}
	if v, ok := optsMap["allowExpressions"].(bool); ok {
		opts.allowExpressions = v
	}
	if v, ok := optsMap["allowFunctionsWithoutTypeParameters"].(bool); ok {
		opts.allowFunctionsWithoutTypeParameters = v
	}
	if v, ok := optsMap["allowHigherOrderFunctions"].(bool); ok {
		opts.allowHigherOrderFunctions = v
	}
	if v, ok := optsMap["allowIIFEs"].(bool); ok {
		opts.allowIIFEs = v
	}
	if v, ok := optsMap["allowTypedFunctionExpressions"].(bool); ok {
		opts.allowTypedFunctionExpressions = v
	}
	return opts
}

type functionInfo struct {
	node    *ast.Node
	returns []*ast.Node
}

var ExplicitFunctionReturnTypeRule = rule.CreateRule(rule.Rule{
	Name: "explicit-function-return-type",
	Run:  run,
})

func run(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
	opts := parseOptions(rawOptions)
	functionStack := make([]*functionInfo, 0)

	enterFunction := func(node *ast.Node) {
		functionStack = append(functionStack, &functionInfo{node: node})
	}

	popFunctionInfo := func() *functionInfo {
		if len(functionStack) == 0 {
			return nil
		}
		info := functionStack[len(functionStack)-1]
		functionStack = functionStack[:len(functionStack)-1]
		return info
	}

	report := func(node *ast.Node) {
		loc := utils.GetFunctionHeadLoc(ctx.SourceFile, node)
		ctx.ReportRange(loc, rule.RuleMessage{
			Id:          "missingReturnType",
			Description: "Missing return type on function.",
		})
	}

	// checkFunctionReturnType checks the common validity conditions:
	// - allowHigherOrderFunctions + doesImmediatelyReturnFunctionExpression
	// - has explicit return type
	// - is constructor or setter
	// Returns true if the function is valid (no error should be reported).
	checkFunctionReturnType := func(info *functionInfo) bool {
		node := info.node
		// Skip bodyless functions: declare functions, abstract methods, overload signatures.
		// ESLint models these as TSDeclareFunction / TSAbstractMethodDefinition which are
		// separate node types not visited by the rule. In tsgo they share the same Kind.
		if node.Body() == nil {
			return true
		}
		if opts.allowHigherOrderFunctions && typescriptutil.DoesImmediatelyReturnFunctionExpression(node, info.returns) {
			return true
		}
		if node.Type() != nil {
			return true
		}
		if node.Kind == ast.KindConstructor || node.Kind == ast.KindSetAccessor {
			return true
		}
		return false
	}

	isAllowedFunction := func(node *ast.Node) bool {
		if opts.allowFunctionsWithoutTypeParameters && node.TypeParameters() == nil {
			return true
		}
		if opts.allowIIFEs && isIIFE(node) {
			return true
		}
		if len(opts.allowedNames) == 0 {
			return false
		}
		return isNameAllowed(ctx.SourceFile, node, opts.allowedNames)
	}

	// exitFunctionExpression handles ArrowFunction and FunctionExpression exit
	exitFunctionExpression := func(node *ast.Node) {
		info := popFunctionInfo()
		if info == nil {
			return
		}

		// allowConciseArrowFunctionExpressionsStartingWithVoid
		if opts.allowConciseArrowFunctionExpressionsStartingWithVoid &&
			node.Kind == ast.KindArrowFunction {
			af := node.AsArrowFunction()
			if af.Body != nil && af.Body.Kind != ast.KindBlock {
				if ast.SkipParentheses(af.Body).Kind == ast.KindVoidExpression {
					return
				}
			}
		}

		if isAllowedFunction(node) {
			return
		}

		if opts.allowTypedFunctionExpressions &&
			(typescriptutil.IsValidFunctionExpressionReturnType(
				node,
				opts.allowTypedFunctionExpressions,
				opts.allowExpressions,
				opts.allowDirectConstAssertionInArrowFunctions,
			) || typescriptutil.AncestorHasReturnType(node)) {
			return
		}

		if !checkFunctionReturnType(info) {
			report(node)
		}
	}

	// exitFunctionDeclaration handles FunctionDeclaration exit
	exitFunctionDeclaration := func(node *ast.Node) {
		info := popFunctionInfo()
		if info == nil {
			return
		}
		if isAllowedFunction(node) {
			return
		}
		if opts.allowTypedFunctionExpressions && node.Type() != nil {
			return
		}
		if !checkFunctionReturnType(info) {
			report(node)
		}
	}

	// exitMethodOrAccessor handles MethodDeclaration and GetAccessor exit
	exitMethodOrAccessor := func(node *ast.Node) {
		info := popFunctionInfo()
		if info == nil {
			return
		}
		if isAllowedFunction(node) {
			return
		}

		// In ESLint, object methods/getters are Property > FunctionExpression, so they go
		// through exitFunctionExpression. In tsgo they are MethodDeclaration/GetAccessor
		// directly inside ObjectLiteralExpression. Apply the same expression-path checks.
		if node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
			if opts.allowTypedFunctionExpressions &&
				(isObjectMethodTyped(node, opts) || typescriptutil.AncestorHasReturnType(node) || opts.allowExpressions) {
				return
			}
		}

		if !checkFunctionReturnType(info) {
			report(node)
		}
	}

	// Constructors and setters never need return types — just pop the stack.
	exitConstructorOrSetter := func(_ *ast.Node) {
		popFunctionInfo()
	}

	return rule.RuleListeners{
		// Enter listeners - push to stack
		ast.KindFunctionDeclaration: enterFunction,
		ast.KindFunctionExpression:  enterFunction,
		ast.KindArrowFunction:       enterFunction,
		ast.KindMethodDeclaration:   enterFunction,
		ast.KindGetAccessor:         enterFunction,
		ast.KindConstructor:         enterFunction,
		ast.KindSetAccessor:         enterFunction,

		// Exit listeners
		rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunctionDeclaration,
		rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunctionExpression,
		rule.ListenerOnExit(ast.KindArrowFunction):       exitFunctionExpression,
		rule.ListenerOnExit(ast.KindMethodDeclaration):   exitMethodOrAccessor,
		rule.ListenerOnExit(ast.KindGetAccessor):         exitMethodOrAccessor,
		rule.ListenerOnExit(ast.KindConstructor):         exitConstructorOrSetter,
		rule.ListenerOnExit(ast.KindSetAccessor):         exitConstructorOrSetter,

		// Return statement listener
		ast.KindReturnStatement: func(node *ast.Node) {
			if len(functionStack) > 0 {
				info := functionStack[len(functionStack)-1]
				info.returns = append(info.returns, node)
			}
		},
	}
}

// isIIFE checks if a function node is the callee of an immediately invoked call expression.
func isIIFE(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
	return callee == node
}

// isObjectMethodTyped checks if an object method is in a typed context.
// This handles the tsgo-specific case where object method shorthand (e.g., { foo() {} })
// is a MethodDeclaration inside ObjectLiteralExpression.
//
// NOTE: IsConstructorArgument is intentionally NOT checked here. In ESLint,
// isConstructorArgument is only for direct function arguments to new expressions
// (e.g., `new Foo(() => {})`), not for methods inside objects passed to new expressions
// (e.g., `new Proxy(obj, { get() {} })`).
func isObjectMethodTyped(node *ast.Node, opts options) bool {
	if !opts.allowTypedFunctionExpressions {
		return false
	}
	objectExpr := node.Parent
	if objectExpr == nil || objectExpr.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	parent := typescriptutil.GetEffectiveParent(objectExpr)
	if parent == nil {
		return false
	}
	return typescriptutil.IsTypedParent(parent, node) ||
		typescriptutil.IsPropertyOfObjectWithType(parent, node)
}

// isNameAllowed checks if a function's name is in the allowed names list.
func isNameAllowed(sourceFile *ast.SourceFile, node *ast.Node, allowedNames []string) bool {
	var funcName string

	switch node.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		// Check if the function expression has a name
		if node.Kind == ast.KindFunctionExpression {
			fe := node.AsFunctionExpression()
			if fe.Name() != nil && fe.Name().Kind == ast.KindIdentifier {
				funcName = fe.Name().AsIdentifier().Text
			}
		}
		if funcName == "" {
			parent := node.Parent
			if parent == nil {
				return false
			}
			switch parent.Kind {
			case ast.KindVariableDeclaration:
				decl := parent.AsVariableDeclaration()
				if decl.Name() != nil && decl.Name().Kind == ast.KindIdentifier {
					funcName = decl.Name().AsIdentifier().Text
				}
			case ast.KindMethodDeclaration:
				md := parent.AsMethodDeclaration()
				if md.Name() != nil && md.Name().Kind == ast.KindIdentifier {
					funcName = md.Name().AsIdentifier().Text
				}
			case ast.KindPropertyDeclaration:
				pd := parent.AsPropertyDeclaration()
				if pd.Name() != nil && pd.Name().Kind == ast.KindIdentifier {
					funcName = pd.Name().AsIdentifier().Text
				}
			case ast.KindPropertyAssignment:
				pa := parent.AsPropertyAssignment()
				if pa.Name() != nil && pa.Name().Kind == ast.KindIdentifier {
					funcName = pa.Name().AsIdentifier().Text
				}
			}
		}
	case ast.KindFunctionDeclaration:
		fd := node.AsFunctionDeclaration()
		if fd.Name() != nil && fd.Name().Kind == ast.KindIdentifier {
			funcName = fd.Name().AsIdentifier().Text
		}
	case ast.KindMethodDeclaration:
		md := node.AsMethodDeclaration()
		if md.Name() != nil && md.Name().Kind == ast.KindIdentifier {
			funcName = md.Name().AsIdentifier().Text
		}
	case ast.KindGetAccessor:
		ga := node.AsGetAccessorDeclaration()
		if ga.Name() != nil && ga.Name().Kind == ast.KindIdentifier {
			funcName = ga.Name().AsIdentifier().Text
		}
	}

	if funcName == "" {
		return false
	}
	for _, name := range allowedNames {
		if name == funcName {
			return true
		}
	}
	return false
}
