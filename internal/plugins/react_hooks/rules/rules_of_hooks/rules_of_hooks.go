package rules_of_hooks

import (
	"math"
	"math/big"
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	analysis "github.com/web-infra-dev/rslint/internal/plugins/react_hooks/code_path_analysis"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var RulesOfHooksRule = rule.Rule{
	Name: "react-hooks/rules-of-hooks",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		codePathReactHooksMapStack := make([]map[*analysis.CodePathSegment][]*ast.Node, 0)
		codePathSegmentStack := make([]*analysis.CodePathSegment, 0)

		// Track useEffectEvent functions and current effect
		// This implements the enhanced hook detection for useEffectEvent functions
		// which can only be called from the same component and within useEffect
		useEffectEventFunctions := make(map[*ast.Node]bool)
		var lastEffect *ast.Node

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
		onCodePathEnd := func(codePath *analysis.CodePath, codePathNode *ast.Node) {
			if len(codePathReactHooksMapStack) == 0 {
				return
			}

			// Pop the current hooks map
			reactHooksMap := codePathReactHooksMapStack[len(codePathReactHooksMapStack)-1]
			codePathReactHooksMapStack = codePathReactHooksMapStack[:len(codePathReactHooksMapStack)-1]

			if len(reactHooksMap) == 0 {
				return
			}

			// Set to track cyclic segments
			cyclic := make(map[string]bool)

			// Cache for path counting functions
			countPathsFromStartCache := make(map[string]*big.Int)
			countPathsToEndCache := make(map[string]*big.Int)
			shortestPathLengthToStartCache := make(map[string]*int) // nil indicates cycle

			// Count paths from start to a segment
			var countPathsFromStart func(*analysis.CodePathSegment, []string) *big.Int
			countPathsFromStart = func(segment *analysis.CodePathSegment, pathHistory []string) *big.Int {
				if pathHistory == nil {
					pathHistory = make([]string, 0)
				}

				segmentID := segment.ID()
				if paths, exists := countPathsFromStartCache[segmentID]; exists && paths != nil {
					return paths
				}

				pathList := pathHistory

				// If `pathList` includes the current segment then we've found a cycle!
				// We need to fill `cyclic` with all segments inside cycle
				hasCyclic := false
				for _, path := range pathList {
					if path == segmentID || hasCyclic {
						hasCyclic = true
						cyclic[path] = true
					}
				}
				if hasCyclic {
					return big.NewInt(0)
				}

				pathList = append(pathList, segmentID)

				var paths *big.Int
				if codePath.HasThrownSegment(segment) {
					paths = big.NewInt(0)
				} else if len(segment.PrevSegments()) == 0 {
					paths = big.NewInt(1)
				} else {
					paths = big.NewInt(0)
					for _, prevSegment := range segment.PrevSegments() {
						prevPaths := countPathsFromStart(prevSegment, pathList)
						paths.Add(paths, prevPaths)
					}
				}

				if segment.Reachable() && paths.Cmp(big.NewInt(0)) == 0 {
					countPathsFromStartCache[segmentID] = nil
				} else {
					countPathsFromStartCache[segmentID] = paths
				}

				return paths
			}

			// countPathsToEnd counts the number of code paths from a given segment to the end of the
			// function. For example:
			//
			//     func MyComponent() {
			//         // Segment 1
			//         if condition {
			//             // Segment 2
			//         } else {
			//             // Segment 3
			//         }
			//     }
			//
			// Segments 2 and 3 have one path to the end of MyComponent and
			// segment 1 has two paths to the end of MyComponent since we could
			// either take the path of segment 2 or segment 3.
			//
			// This function also populates the cyclic map with cyclic segments.
			var countPathsToEnd func(*analysis.CodePathSegment, []string) *big.Int
			countPathsToEnd = func(segment *analysis.CodePathSegment, pathHistory []string) *big.Int {
				if pathHistory == nil {
					pathHistory = make([]string, 0)
				}

				segmentID := segment.ID()
				if paths, exists := countPathsToEndCache[segmentID]; exists {
					return paths
				}

				pathList := pathHistory

				// If `pathList` includes the current segment then we've found a cycle!
				// We need to fill `cyclic` with all segments inside cycle
				hasCyclic := false
				for _, path := range pathList {
					if path == segmentID || hasCyclic {
						hasCyclic = true
						cyclic[path] = true
					}
				}
				if hasCyclic {
					return big.NewInt(0)
				}

				// add the current segment to pathList
				pathList = append(pathList, segmentID)

				var paths *big.Int
				if codePath.HasThrownSegment(segment) {
					paths = big.NewInt(0)
				} else if len(segment.NextSegments()) == 0 {
					paths = big.NewInt(1)
				} else {
					paths = big.NewInt(0)
					for _, nextSegment := range segment.NextSegments() {
						nextPaths := countPathsToEnd(nextSegment, pathList)
						paths.Add(paths, nextPaths)
					}
				}

				countPathsToEndCache[segmentID] = paths
				return paths
			}

			// Get shortest path length to start
			var shortestPathLengthToStart func(*analysis.CodePathSegment) int
			shortestPathLengthToStart = func(segment *analysis.CodePathSegment) int {
				segmentID := segment.ID()

				if lengthPtr, exists := shortestPathLengthToStartCache[segmentID]; exists {
					if lengthPtr == nil {
						return math.MaxInt32
					}
					return *lengthPtr
				}

				shortestPathLengthToStartCache[segmentID] = nil

				var length int
				if len(segment.PrevSegments()) == 0 {
					length = 1
				} else {
					length = math.MaxInt32
					for _, prevSegment := range segment.PrevSegments() {
						prevLength := shortestPathLengthToStart(prevSegment)
						if prevLength < length {
							length = prevLength
						}
					}
					if length < math.MaxInt32 {
						length++
					}
				}

				shortestPathLengthToStartCache[segmentID] = &length
				return length
			}

			// Count all paths from start to end
			allPathsFromStartToEnd := countPathsToEnd(codePath.InitialSegment(), nil)

			// Get function name for this code path
			codePathFunctionName := getFunctionName(codePathNode)

			// Check if we're inside a component or hook
			isSomewhereInsideComponentOrHook := isInsideComponentOrHook(codePathNode)
			isDirectlyInsideComponentOrHook := false

			if codePathFunctionName != "" {
				isDirectlyInsideComponentOrHook = isComponentName(codePathFunctionName) || isHookName(codePathFunctionName)
			} else {
				isDirectlyInsideComponentOrHook = isForwardRefCallback(codePathNode) || isMemoCallback(codePathNode)
			}

			// Compute shortest final path length
			shortestFinalPathLength := math.MaxInt32
			for _, finalSegment := range codePath.FinalSegments() {
				if !finalSegment.Reachable() {
					continue
				}
				length := shortestPathLengthToStart(finalSegment)
				if length < shortestFinalPathLength {
					shortestFinalPathLength = length
				}
			}

			// Process each segment with React hooks
			for segment, reactHooks := range reactHooksMap {
				// NOTE: We could report here that the hook is not reachable, but
				// that would be redundant with more general "no unreachable"
				// lint rules.
				if !segment.Reachable() {
					continue
				}

				// If there are any final segments with a shorter path to start then
				// we possibly have an early return.
				//
				// If our segment is a final segment itself then siblings could
				// possibly be early returns.
				possiblyHasEarlyReturn := false
				if len(segment.NextSegments()) == 0 {
					possiblyHasEarlyReturn = shortestFinalPathLength <= shortestPathLengthToStart(segment)
				} else {
					possiblyHasEarlyReturn = shortestFinalPathLength < shortestPathLengthToStart(segment)
				}

				// Count all the paths from the start of our code path to the end of
				// our code path that go _through_ this segment. The critical piece
				// of this is _through_. If we just call `countPathsToEnd(segment)`
				// then we neglect that we may have gone through multiple paths to get
				// to this point! Consider:
				//
				// ```js
				// function MyComponent() {
				//   if (a) {
				//     // Segment 1
				//   } else {
				//     // Segment 2
				//   }
				//   // Segment 3
				//   if (b) {
				//     // Segment 4
				//   } else {
				//     // Segment 5
				//   }
				// }
				// ```
				//
				// In this component we have four code paths:
				//
				// 1. `a = true; b = true`
				// 2. `a = true; b = false`
				// 3. `a = false; b = true`
				// 4. `a = false; b = false`
				//
				// From segment 3 there are two code paths to the end through segment
				// 4 and segment 5. However, we took two paths to get here through
				// segment 1 and segment 2.
				//
				// If we multiply the paths from start (two) by the paths to end (two)
				// for segment 3 we get four. Which is our desired count.
				pathsFromStart := countPathsFromStart(segment, nil)
				pathsToEnd := countPathsToEnd(segment, nil)
				pathsFromStartToEnd := new(big.Int).Mul(pathsFromStart, pathsToEnd)

				// Is this hook a part of a cyclic segment?
				isCyclic := cyclic[segment.ID()]

				// Process each hook in this segment
				for _, hook := range reactHooks {
					// Skip if flow suppression exists
					if hasFlowSuppression(hook) {
						continue
					}

					hookText := getNodeText(hook)
					isUseHook := isUseIdentifier(hook)

					// Report error for use() in try/catch
					if isUseHook && isInsideTryCatch(hook) {
						ctx.ReportNode(hook, buildTryCatchUseMessage(hookText))
						continue
					}

					// Report error for hooks in loops (except use())
					if (isCyclic || isInsideDoWhileLoop(hook)) && !isUseHook {
						ctx.ReportNode(hook, buildLoopHookMessage(hookText))
						continue
					}

					// Check if we're in a valid context for hooks
					if isDirectlyInsideComponentOrHook {
						// Check for async function
						if isInsideAsyncFunction(codePathNode) {
							ctx.ReportNode(hook, buildAsyncComponentHookMessage(hookText))
							continue
						}

						pathsCmp := pathsFromStartToEnd.Cmp(allPathsFromStartToEnd)

						// Check for conditional calls (except use() and do-while loops)
						if !isCyclic &&
							pathsCmp != 0 &&
							!isUseHook &&
							!isInsideDoWhileLoop(hook) {
							var message rule.RuleMessage
							if possiblyHasEarlyReturn {
								message = buildConditionalHookWithEarlyReturnMessage(hookText)
							} else {
								message = buildConditionalHookMessage(hookText)
							}
							ctx.ReportNode(hook, message)
						}
					} else {
						// Handle various invalid contexts
						if isInsideClass(codePathNode) {
							ctx.ReportNode(hook, buildClassHookMessage(hookText))
						} else if codePathFunctionName != "" {
							// Custom message if we found an invalid function name.
							ctx.ReportNode(hook, buildFunctionHookMessage(hookText, codePathFunctionName))
						} else if isTopLevel(codePathNode) {
							// These are dangerous if you have inline requires enabled.
							ctx.ReportNode(hook, buildTopLevelHookMessage(hookText))
						} else if isSomewhereInsideComponentOrHook && !isUseHook {
							// Assume in all other cases the user called a hook in some
							// random function callback. This should usually be true for
							// anonymous function expressions. Hopefully this is clarifying
							// enough in the common case that the incorrect message in
							// uncommon cases doesn't matter.
							// `use(...)` can be called in callbacks.
							ctx.ReportNode(hook, buildGenericHookMessage(hookText))
						}
					}
				}
			}
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
				callExpr := node.AsCallExpression()

				// Check if this is a hook call
				if isHook(callExpr.Expression) {
					// Add the hook node to a map keyed by the code path segment
					if len(codePathReactHooksMapStack) > 0 && len(codePathSegmentStack) > 0 {
						reactHooksMap := codePathReactHooksMapStack[len(codePathReactHooksMapStack)-1]
						codePathSegment := codePathSegmentStack[len(codePathSegmentStack)-1]

						reactHooks := reactHooksMap[codePathSegment]
						if reactHooks == nil {
							reactHooks = []*ast.Node{}
							reactHooksMap[codePathSegment] = reactHooks
						}
						reactHooksMap[codePathSegment] = append(reactHooksMap[codePathSegment], callExpr.Expression)
					}
				}

				// Check for useEffect and useEffectEvent calls
				nodeWithoutNamespace := getNodeWithoutReactNamespace(callExpr.Expression)
				if (isUseEffectIdentifier(nodeWithoutNamespace) || isUseEffectEventIdentifier(nodeWithoutNamespace)) &&
					len(callExpr.Arguments.Nodes) > 0 {
					lastEffect = node
				}
			},
			// CallExpression exit handler
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				if node == lastEffect {
					lastEffect = nil
				}
			},
			ast.KindIdentifier: func(node *ast.Node) {
				// Check for useEffectEvent function references outside effects
				if lastEffect == nil && useEffectEventFunctions[node] {
					nodeText := scanner.GetTextOfNode(node)
					message := "`" + nodeText + "` is a function created with React Hook \"useEffectEvent\", and can only be called from " +
						"the same component."

					// Check if it's being called
					parent := node.Parent
					if parent == nil || parent.Kind != ast.KindCallExpression {
						message += " They cannot be assigned to variables or passed down."
					}

					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "useEffectEventReference",
						Description: message,
					})
				}
			},
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				// function MyComponent() { const onClick = useEffectEvent(...) }
				if isInsideComponentOrHookFromScope(node) {
					recordAllUseEffectEventFunctions(getScope(node))
				}
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				// const MyComponent = () => { const onClick = useEffectEvent(...) }
				if isInsideComponentOrHookFromScope(node) {
					recordAllUseEffectEventFunctions(getScope(node))
				}
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
	switch node.Kind {
	case ast.KindIdentifier:
		return isHookName(node.Text())
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		if name == nil || !isHook(name) {
			return false
		}

		expr := node.AsPropertyAccessExpression().Expression
		if expr == nil || !ast.IsIdentifier(expr) {
			return false
		}

		return isPascalCaseNameSpace(expr.AsIdentifier().Text)
	}
	return false
}

// Helper function to get function name from AST node
func getFunctionName(node *ast.Node) string {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		// function useHook() {}
		// const whatever = function useHook() {};
		//
		// Function declaration or function expression names win over any
		// assignment statements or other renames.
		return node.AsFunctionDeclaration().Name().Text()
	case ast.KindFunctionExpression:
		name := node.AsFunctionExpression().Name()
		if name != nil {
			return node.AsFunctionExpression().Name().Text()
		}
	case ast.KindArrowFunction:
		if node.Parent != nil {
			switch node.Parent.Kind {
			case ast.KindVariableDeclaration, // const useHook = () => {};
				ast.KindShorthandPropertyAssignment, // ({k = () => { useState(); }} = {});
				ast.KindBindingElement,              // const {j = () => { useState(); }} = {};
				ast.KindPropertyAssignment:          // ({f: () => { useState(); }});
				if ast.IsInExpressionContext(node) {
					return node.Parent.Name().Text()
				}
			case ast.KindBinaryExpression:
				if node.Parent.AsBinaryExpression().Right == node {
					left := node.Parent.AsBinaryExpression().Left
					switch left.Kind {
					case ast.KindIdentifier:
						// e = () => { useState(); };
						return left.AsIdentifier().Text
					case ast.KindPropertyAccessExpression:
						// Namespace.useHook = () => { useState(); };
						return left.AsPropertyAccessExpression().Name().Text()
					}
				}
			}
		}
		return ""
	case ast.KindMethodDeclaration:
		// NOTE: We could also support `ClassProperty` and `MethodDefinition`
		// here to be pedantic. However, hooks in a class are an anti-pattern. So
		// we don't allow it to error early.
		//
		// class {useHook = () => {}}
		// class {useHook() {}}
		if ast.GetContainingClass(node) != nil {
			return ""
		}

		// {useHook: () => {}}
		// {useHook() {}}
		return node.AsMethodDeclaration().Name().Text()
	}
	return ""
}

// Helper function to check if node is inside a component or hook
func isInsideComponentOrHook(node *ast.Node) bool {
	// Walk up the AST to find function declarations
	// and check if any of them are components or hooks
	current := node
	for current != nil {
		functionName := getFunctionName(current)
		if functionName != "" && (isComponentName(functionName) || isHookName(functionName)) {
			return true
		}
		if isForwardRefCallback(current) || isMemoCallback(current) {
			return true
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
	current := node
	for current != nil {
		if isAsyncFunction(current) {
			return true
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
	return isReactFunction(node, "use")
}

// Helper function to check if node is at top level
func isTopLevel(node *ast.Node) bool {
	return node.Kind == ast.KindSourceFile
}

// Helper function to check if a call expression is a React function
func isReactFunction(node *ast.Node, functionName string) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindIdentifier:
		// Direct call: forwardRef()
		identifier := node.AsIdentifier()
		if identifier != nil {
			name := scanner.GetTextOfNode(&identifier.Node)
			return name == functionName
		}
	case ast.KindPropertyAccessExpression:
		// Property access: React.forwardRef()
		propAccess := node.AsPropertyAccessExpression()
		if propAccess != nil {
			nameNode := propAccess.Name()
			if nameNode != nil {
				name := scanner.GetTextOfNode(nameNode)
				if name == functionName {
					// Check if the object is React
					expr := propAccess.Expression
					if expr != nil && expr.Kind == ast.KindIdentifier {
						objName := scanner.GetTextOfNode(expr)
						return objName == "React"
					}
				}
			}
		}
	}
	return false
}

// Helper function to check if the node is a callback argument of forwardRef
// This render function should follow the rules of hooks
func isForwardRefCallback(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent
	if parent.Kind == ast.KindCallExpression {
		callExpr := parent.AsCallExpression()
		if callExpr != nil && callExpr.Expression != nil {
			return isReactFunction(callExpr.Expression, "forwardRef")
		}
	}
	return false
}

// Helper function to check if the node is a callback argument of memo
func isMemoCallback(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent
	if parent.Kind == ast.KindCallExpression {
		callExpr := parent.AsCallExpression()
		if callExpr != nil && callExpr.Expression != nil {
			return isReactFunction(callExpr.Expression, "memo")
		}
	}
	return false
}

// Helper function to check for flow suppression comments
func hasFlowSuppression(node *ast.Node) bool {
	// No need implementation
	return false
}

// Helper function to get node text
func getNodeText(node *ast.Node) string {
	// This is a simplified implementation
	// You would extract the text from the source code
	if node != nil && node.Kind == ast.KindIdentifier {
		return scanner.GetTextOfNode(node)
	}
	return ""
}

// Helper function to check if node is inside do-while loop
func isInsideDoWhileLoop(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if current.Kind == ast.KindDoStatement {
			return true
		}
		current = current.Parent
	}
	return false
}

// Helper function to check if function is async
func isAsyncFunction(node *ast.Node) bool {
	if isFunctionLike(node) {
		return ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
	}
	return false
}

// Helper function to check if node is useEffect identifier
func isUseEffectIdentifier(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	text := scanner.GetTextOfNode(node)
	return text == "useEffect"
}

// Helper function to check if node is useEffectEvent identifier
func isUseEffectEventIdentifier(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	text := scanner.GetTextOfNode(node)
	return text == "useEffectEvent"
}

// Helper function to get node without React namespace
func getNodeWithoutReactNamespace(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	// If it's React.someHook, return someHook
	if node.Kind == ast.KindPropertyAccessExpression {
		propAccess := node.AsPropertyAccessExpression()
		if propAccess != nil {
			expr := propAccess.Expression
			if expr != nil && expr.Kind == ast.KindIdentifier {
				identifier := expr.AsIdentifier()
				if identifier != nil && scanner.GetTextOfNode(&identifier.Node) == "React" {
					return propAccess.Name()
				}
			}
		}
	}

	return node
}

// Helper function to get scope (simplified implementation)
func getScope(node *ast.Node) *ast.Node {
	// This is a simplified implementation
	// In a real implementation, you would traverse the scope chain
	return node
}

// Helper function to record all useEffectEvent functions (simplified)
func recordAllUseEffectEventFunctions(scope *ast.Node) {
	// !!! useEffectEvent
}

// Helper function to check if we're inside a component or hook (from scope context)
func isInsideComponentOrHookFromScope(node *ast.Node) bool {
	// This is a simplified implementation based on the existing isInsideComponentOrHook
	return isInsideComponentOrHook(node)
}
