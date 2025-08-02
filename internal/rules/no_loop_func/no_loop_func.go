package no_loop_func

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Set to track IIFE nodes that should be skipped
type iifeTacker struct {
	skippedIIFENodes map[*ast.Node]bool
}

func newIIFETracker() *iifeTacker {
	return &iifeTacker{
		skippedIIFENodes: make(map[*ast.Node]bool),
	}
}

func (it *iifeTacker) add(node *ast.Node) {
	it.skippedIIFENodes[node] = true
}

func (it *iifeTacker) has(node *ast.Node) bool {
	return it.skippedIIFENodes[node]
}

// Check if node is an IIFE (Immediately Invoked Function Expression)
func isIIFE(node *ast.Node) bool {
	parent := node.Parent
	return parent != nil &&
		parent.Kind == ast.KindCallExpression &&
		parent.AsCallExpression().Expression == node
}

// Gets the containing loop node of a specified node
func getContainingLoopNode(node *ast.Node, iifeTracker *iifeTacker) *ast.Node {
	currentNode := node
	for currentNode.Parent != nil {
		parent := currentNode.Parent

		switch parent.Kind {
		case ast.KindWhileStatement, ast.KindDoStatement:
			return parent

		case ast.KindForStatement:
			// `init` is outside of the loop
			forStmt := parent.AsForStatement()
			if forStmt.Initializer != currentNode {
				return parent
			}

		case ast.KindForInStatement, ast.KindForOfStatement:
			// `right` is outside of the loop
			forInOf := parent.AsForInOrOfStatement()
			if forInOf.Expression != currentNode {
				return parent
			}

		case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindFunctionDeclaration:
			// We don't need to check nested functions.
			// We need to check nested functions only in case of IIFE.
			if iifeTracker.has(parent) {
				break
			}
			return nil
		}

		currentNode = currentNode.Parent
	}

	return nil
}

// Gets the most outer loop node
func getTopLoopNode(node *ast.Node, excludedNode *ast.Node) *ast.Node {
	border := 0
	if excludedNode != nil {
		border = excludedNode.End()
	}

	retv := node
	containingLoopNode := node

	for containingLoopNode != nil && containingLoopNode.Pos() >= border {
		retv = containingLoopNode
		containingLoopNode = getContainingLoopNode(containingLoopNode, &iifeTacker{})
	}

	return retv
}

// Check if the reference is safe
func isSafe(loopNode *ast.Node, reference *ast.Symbol, variable *ast.Symbol, ctx rule.RuleContext) bool {
	if variable == nil || len(variable.Declarations) == 0 {
		// Variables without declarations are likely global variables or built-ins
		// These are generally safe since they're not loop-bound
		return true
	}

	declaration := variable.Declarations[0]

	// Check if this is a declaration from a library file (lib.*.d.ts)
	// These are global/built-in declarations and should be considered safe
	sourceFile := ast.GetSourceFileOfNode(declaration)
	if sourceFile != nil {
		fileName := sourceFile.FileName()
		if strings.Contains(fileName, "lib.") && strings.HasSuffix(fileName, ".d.ts") {
			return true
		}
		// Also check for node_modules/@types or similar built-in type definitions
		if strings.Contains(fileName, "node_modules/@types") ||
			strings.Contains(fileName, "typescript/lib/") {
			return true
		}
	}

	// Get the kind of variable declaration
	kind := ""
	if declaration.Parent != nil && declaration.Parent.Kind == ast.KindVariableDeclaration {
		varDecl := declaration.Parent.AsVariableDeclaration()
		flags := varDecl.Flags
		if flags&ast.NodeFlagsConst != 0 {
			kind = "const"
		} else if flags&ast.NodeFlagsLet != 0 {
			kind = "let"
		} else {
			kind = "var"
		}
	}

	// Variables declared by `const` are safe
	if kind == "const" {
		return true
	}

	// Variables declared by `let` in the loop are safe
	// It's a different instance from the next loop step's
	if kind == "let" && declaration.Parent != nil &&
		declaration.Parent.Pos() > loopNode.Pos() &&
		declaration.Parent.End() < loopNode.End() {
		return true
	}

	// Check if this is a function parameter, which is generally safe
	if declaration.Parent != nil && declaration.Parent.Kind == ast.KindParameter {
		return true
	}

	// For 'var' declarations, check if they're actually loop control variables
	if kind == "var" {
		// Check if this variable is declared as part of a loop statement
		parent := declaration.Parent
		for parent != nil {
			switch parent.Kind {
			case ast.KindForStatement:
				forStmt := parent.AsForStatement()
				if forStmt.Initializer != nil &&
					(declaration.Pos() >= forStmt.Initializer.Pos() &&
						declaration.End() <= forStmt.Initializer.End()) {
					// This is a loop control variable - unsafe
					return false
				}
			case ast.KindForInStatement, ast.KindForOfStatement:
				forInOf := parent.AsForInOrOfStatement()
				if forInOf.Initializer != nil &&
					(declaration.Pos() >= forInOf.Initializer.Pos() &&
						declaration.End() <= forInOf.Initializer.End()) {
					// This is a loop control variable - unsafe
					return false
				}
			}
			parent = parent.Parent
		}

		// If it's a 'var' but not a loop control variable, it might still be unsafe
		// if it's modified within the loop. For now, we'll be conservative.
		return false
	}

	// For variables not covered by the above cases, assume they're unsafe
	return false
}

// Gets unsafe variable references in a function
func getUnsafeRefs(node *ast.Node, loopNode *ast.Node, ctx rule.RuleContext) []string {
	unsafeRefs := []string{}
	seenVars := make(map[string]bool)

	// Traverse the function body to find variable references
	var checkNode func(n *ast.Node)
	checkNode = func(n *ast.Node) {
		if n == nil {
			return
		}

		// Check if this is an identifier that references a variable
		if n.Kind == ast.KindIdentifier {
			identifier := n.AsIdentifier()
			varName := identifier.Text

			// Skip if we've already checked this variable
			if seenVars[varName] {
				return
			}
			seenVars[varName] = true

			// Get the symbol for this identifier
			symbol := ctx.TypeChecker.GetSymbolAtLocation(n)
			if symbol == nil {
				return
			}

			// Check if this is a type reference - those are always safe
			if symbol.Flags&ast.SymbolFlagsType != 0 ||
				symbol.Flags&ast.SymbolFlagsInterface != 0 ||
				symbol.Flags&ast.SymbolFlagsTypeAlias != 0 {
				return
			}

			// Check if the variable is declared outside the function
			if len(symbol.Declarations) > 0 {
				varDecl := symbol.Declarations[0]
				// If the variable is declared outside this function, check if it's safe
				if varDecl.Pos() < node.Pos() || varDecl.End() > node.End() {
					if !isSafe(loopNode, symbol, symbol, ctx) {
						unsafeRefs = append(unsafeRefs, varName)
					}
				}
			} else {
				// Variables with no declarations are likely globals or built-ins
				// These should be safe, so we don't add them to unsafe refs
			}
		}

		// Recursively check child nodes
		n.ForEachChild(func(child *ast.Node) bool {
			checkNode(child)
			return false
		})
	}

	// Start checking from the function body
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		if body := node.AsFunctionDeclaration().Body; body != nil {
			checkNode(body)
		}
	case ast.KindFunctionExpression:
		if body := node.AsFunctionExpression().Body; body != nil {
			checkNode(body)
		}
	case ast.KindArrowFunction:
		arrow := node.AsArrowFunction()
		if arrow.Body != nil {
			checkNode(arrow.Body)
		}
	}

	return unsafeRefs
}

// Build error message for unsafe references
func buildUnsafeRefsMessage(varNames []string) rule.RuleMessage {
	quotedNames := make([]string, len(varNames))
	for i, name := range varNames {
		quotedNames[i] = fmt.Sprintf("'%s'", name)
	}
	return rule.RuleMessage{
		Id:          "unsafeRefs",
		Description: fmt.Sprintf("Function declared in a loop contains unsafe references to variable(s) %s.", strings.Join(quotedNames, ", ")),
	}
}

var NoLoopFuncRule = rule.Rule{
	Name: "no-loop-func",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		iifeTracker := newIIFETracker()

		checkForLoops := func(node *ast.Node) {
			loopNode := getContainingLoopNode(node, iifeTracker)
			if loopNode == nil {
				return
			}

			// Check if this is an IIFE
			isGenerator := false

			switch node.Kind {
			case ast.KindFunctionDeclaration:
				fn := node.AsFunctionDeclaration()
				isGenerator = fn.AsteriskToken != nil
			case ast.KindFunctionExpression:
				fn := node.AsFunctionExpression()
				isGenerator = fn.AsteriskToken != nil
			}

			if !isGenerator && isIIFE(node) {
				isFunctionExpression := node.Kind == ast.KindFunctionExpression

				// Check if the function is referenced elsewhere
				isFunctionReferenced := false
				if isFunctionExpression {
					funcExpr := node.AsFunctionExpression()
					if funcExpr.Name() != nil && ast.IsIdentifier(funcExpr.Name()) {
						// For simplicity, we'll assume named function expressions might be referenced
						// A more accurate check would require full scope analysis
						isFunctionReferenced = true
					}
				}

				if !isFunctionReferenced {
					iifeTracker.add(node)
					return
				}
			}

			// Get unsafe references
			unsafeRefs := getUnsafeRefs(node, loopNode, ctx)

			if len(unsafeRefs) > 0 {
				ctx.ReportNode(node, buildUnsafeRefsMessage(unsafeRefs))
			}
		}

		return rule.RuleListeners{
			ast.KindArrowFunction:       checkForLoops,
			ast.KindFunctionDeclaration: checkForLoops,
			ast.KindFunctionExpression:  checkForLoops,
		}
	},
}
