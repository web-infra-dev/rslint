package no_eval

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var noEvalMessage = rule.RuleMessage{
	Id:          "unexpected",
	Description: "`eval` can be harmful.",
}

func sourceMayUseEval(sourceFile *ast.SourceFile) bool {
	// Parsed identifier names are normalized, including Unicode escapes. The
	// text checks retain computed string accesses such as window['eval']; an
	// escaped computed name necessarily contains both '[' and a backslash.
	// Stay conservative for direct callers that do not provide parser metadata.
	if sourceFile == nil || sourceFile.Identifiers == nil {
		return true
	}
	if _, ok := sourceFile.Identifiers["eval"]; ok {
		return true
	}
	text := sourceFile.Text()
	return strings.Contains(text, "eval") ||
		(strings.Contains(text, "[") && strings.Contains(text, "\\"))
}

// https://eslint.org/docs/latest/rules/no-eval
var NoEvalRule = rule.Rule{
	Name: "no-eval",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		if !sourceMayUseEval(ctx.SourceFile) {
			return nil
		}

		options := rule.LegacyUnwrapOptions(_options)
		allowIndirect := false
		optsMap := utils.GetOptionsMap(options)
		if optsMap != nil {
			if v, ok := optsMap["allowIndirect"].(bool); ok {
				allowIndirect = v
			}
		}

		if allowIndirect {
			// Only flag direct eval() calls
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					call := node.AsCallExpression()
					// Optional calls eval?.() are not direct eval
					if call.QuestionDotToken != nil {
						return
					}
					callee := ast.SkipParentheses(call.Expression)
					if callee != nil && ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "eval" {
						ctx.ReportNode(callee, noEvalMessage)
					}
				},
			}
		}

		return noEvalListeners(ctx)
	},
}

type noEvalState struct {
	ctx              rule.RuleContext
	evalBinding      sourceBindingState
	windowStatus     globalObjectStatus
	globalStatus     globalObjectStatus
	globalThisStatus globalObjectStatus
}

type sourceBindingState uint8

const (
	sourceBindingUnchecked sourceBindingState = iota
	sourceBindingAbsent
	sourceBindingPresent
)

type bindingPresence struct {
	checked bool
	present bool
}

type globalObjectStatus struct {
	bindingPresence
	knownChecked bool
	known        bool
}

func noEvalListeners(ctx rule.RuleContext) rule.RuleListeners {
	state := &noEvalState{ctx: ctx}
	return rule.RuleListeners{
		ast.KindIdentifier:              state.checkIdentifier,
		ast.KindElementAccessExpression: state.checkElementAccess,
	}
}

func (state *noEvalState) checkIdentifier(node *ast.Node) {
	if node.AsIdentifier().Text != "eval" {
		return
	}

	parent := node.Parent
	if parent == nil {
		return
	}

	// Property names are already identifier nodes, so handle dot access here
	// instead of dispatching a listener for every PropertyAccessExpression.
	if ast.IsPropertyAccessExpression(parent) && parent.AsPropertyAccessExpression().Name() == node {
		state.checkMemberAccess(parent.AsPropertyAccessExpression().Expression, node)
		return
	}

	// Grouping parentheses are absent from ESTree. Walk through them so direct
	// calls retain ESLint's scope-independent behavior without a call listener.
	callee := utils.OutermostParenthesizedExpression(node)
	if callee.Parent != nil && ast.IsCallExpression(callee.Parent) &&
		callee.Parent.AsCallExpression().Expression == callee {
		state.ctx.ReportNode(node, noEvalMessage)
		return
	}

	if utils.IsNonReferenceIdentifier(node) || !state.isGlobalEvalReference(node) {
		return
	}
	state.ctx.ReportNode(node, noEvalMessage)
}

func (state *noEvalState) checkElementAccess(node *ast.Node) {
	elementAccess := node.AsElementAccessExpression()
	argument := elementAccess.ArgumentExpression
	if argument == nil || utils.GetStaticStringValue(argument) != "eval" {
		return
	}
	state.checkMemberAccess(elementAccess.Expression, argument)
}

func (state *noEvalState) checkMemberAccess(object *ast.Node, reportNode *ast.Node) {
	object = ast.SkipParentheses(object)
	if object == nil {
		return
	}
	if object.Kind == ast.KindThisKeyword {
		if isThisReferringToGlobal(object, state.ctx.SourceFile) {
			state.ctx.ReportNode(reportNode, noEvalMessage)
		}
		return
	}
	if state.isGlobalObjectChain(object) {
		state.ctx.ReportNode(reportNode, noEvalMessage)
	}
}

func (state *noEvalState) isGlobalEvalReference(node *ast.Node) bool {
	if state.ctx.Globals["eval"] == utils.GlobalAccessOff {
		return false
	}
	// A full-file miss is cached because IsShadowed's SourceFile check scans
	// top-level declarations; repeating it for many global eval references can
	// otherwise become quadratic. Files with any binding retain the exact
	// per-reference lexical check.
	if state.evalBinding == sourceBindingAbsent {
		return true
	}
	if state.evalBinding == sourceBindingUnchecked {
		if sourceHasValueBinding(state.ctx.SourceFile, "eval") {
			state.evalBinding = sourceBindingPresent
		} else {
			state.evalBinding = sourceBindingAbsent
			return true
		}
	}
	return !utils.IsShadowed(node, "eval")
}

func isGlobalObjectName(name string) bool {
	return name == "window" || name == "global" || name == "globalThis"
}

// isGlobalObjectChain accepts chained forms such as window.window and
// window['window'], then verifies that the root is a known, unshadowed global.
func (state *noEvalState) isGlobalObjectChain(node *ast.Node) bool {
	chainName := ""
	for {
		node = ast.SkipParentheses(node)
		if node == nil {
			return false
		}
		if ast.IsIdentifier(node) {
			name := node.AsIdentifier().Text
			return isGlobalObjectName(name) &&
				(chainName == "" || chainName == name) &&
				state.isGlobalObjectReference(node, name)
		}
		if ast.IsPropertyAccessExpression(node) {
			propertyAccess := node.AsPropertyAccessExpression()
			name := propertyAccess.Name()
			if name == nil || !isGlobalObjectName(name.Text()) {
				return false
			}
			if chainName == "" {
				chainName = name.Text()
			} else if chainName != name.Text() {
				return false
			}
			node = propertyAccess.Expression
			continue
		}
		if ast.IsElementAccessExpression(node) {
			elementAccess := node.AsElementAccessExpression()
			name := utils.GetStaticStringValue(elementAccess.ArgumentExpression)
			if !isGlobalObjectName(name) {
				return false
			}
			if chainName == "" {
				chainName = name
			} else if chainName != name {
				return false
			}
			node = elementAccess.Expression
			continue
		}
		return false
	}
}

func (state *noEvalState) isGlobalObjectReference(identifier *ast.Node, name string) bool {
	if state.ctx.Globals[name] == utils.GlobalAccessOff {
		return false
	}

	status := state.globalObjectStatus(name)
	if isInsideNamedExpression(identifier, name) {
		return false
	}
	if !state.hasSourceBinding(name, &status.bindingPresence) {
		return state.isKnownGlobalObject(name, status)
	}

	// Resolve only when this file can shadow the global root. The common
	// no-binding case above avoids a NameResolver/checker round trip entirely.
	if state.ctx.Refs != nil {
		if symbol := state.ctx.Refs.Resolve(identifier); symbol != nil {
			return isGlobalObjectSymbol(symbol, state.ctx.SourceFile)
		}
	}
	return state.isKnownGlobalObject(name, status) && !utils.IsShadowed(identifier, name)
}

func isInsideNamedExpression(identifier *ast.Node, name string) bool {
	for current := identifier.Parent; current != nil; current = current.Parent {
		if (current.Kind == ast.KindFunctionExpression || current.Kind == ast.KindClassExpression) &&
			current.Name() != nil && current.Name().Text() == name {
			return true
		}
	}
	return false
}

func isGlobalObjectSymbol(symbol *ast.Symbol, sourceFile *ast.SourceFile) bool {
	if !utils.IsValueSymbolDeclaredInFile(symbol, sourceFile) {
		return true
	}
	if sourceFile == nil || ast.IsExternalModule(sourceFile) {
		return false
	}

	// ESLint's global-scope variable remains a global-object candidate even
	// when a script declares it itself. Only a nested binding (or a module's
	// top-level binding, handled above) shadows that candidate. The binder's
	// SourceFile.Locals identity also handles a `var` nested syntactically in a
	// block but hoisted into the script's global scope.
	topLevel := sourceFile.Locals[symbol.Name]
	if topLevel == symbol {
		return true
	}
	if topLevel != nil {
		for _, declaration := range symbol.Declarations {
			for _, topLevelDeclaration := range topLevel.Declarations {
				if declaration == topLevelDeclaration {
					return true
				}
			}
		}
	}
	return false
}

func (state *noEvalState) globalObjectStatus(name string) *globalObjectStatus {
	switch name {
	case "window":
		return &state.windowStatus
	case "global":
		return &state.globalStatus
	default:
		return &state.globalThisStatus
	}
}

func (state *noEvalState) isKnownGlobalObject(name string, status *globalObjectStatus) bool {
	if status.knownChecked {
		return status.known
	}
	status.knownChecked = true
	if access, ok := state.ctx.Globals[name]; ok {
		status.known = access.IsDeclared()
	} else if state.ctx.TypeChecker != nil {
		status.known = state.ctx.TypeChecker.GetGlobalSymbol(name, ast.SymbolFlagsValue, nil) != nil
	} else {
		status.known = true
	}
	return status.known
}

func (state *noEvalState) hasSourceBinding(name string, presence *bindingPresence) bool {
	if presence.checked {
		return presence.present
	}
	presence.checked = true
	presence.present = sourceHasValueBinding(state.ctx.SourceFile, name)
	return presence.present
}

func sourceHasValueBinding(sourceFile *ast.SourceFile, name string) bool {
	if sourceFile == nil || !sourceFile.IsBound() {
		return true
	}
	// Binder links every locals container in source order, so a miss here
	// disproves shadowing without walking the full AST for every reference.
	for container := sourceFile.AsNode(); container != nil; {
		data := container.LocalsContainerData()
		if data == nil {
			return true
		}
		if len(data.Locals) != 0 &&
			utils.IsValueSymbolDeclaredInFile(data.Locals[name], sourceFile) {
			return true
		}
		if (container.Kind == ast.KindFunctionExpression || container.Kind == ast.KindClassExpression) &&
			container.Name() != nil && container.Name().Text() == name {
			return true
		}
		container = data.NextContainer
	}
	return false
}

// isThisReferringToGlobal checks if 'this' at the given position refers to the global object.
// It uses ast.GetThisContainer to find the enclosing "this scope" (skipping arrow functions
// and class computed property names) and then determines whether 'this' is the global object.
func isThisReferringToGlobal(thisNode *ast.Node, sourceFile *ast.SourceFile) bool {
	// GetThisContainer with includeArrowFunctions=false skips arrow functions.
	// With includeClassComputedPropertyName=false, computed property names in
	// classes are transparent — the walker jumps past them to the outer scope.
	container := ast.GetThisContainer(thisNode, false /*includeArrowFunctions*/, false /*includeClassComputedPropertyName*/)

	switch container.Kind {
	case ast.KindSourceFile:
		// Top level of script — 'this' is always global (even in strict mode).
		// In modules, 'this' is undefined.
		return !ast.IsExternalModule(sourceFile)

	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		// In strict mode, 'this' is undefined — not global.
		if utils.IsInStrictMode(thisNode, sourceFile) {
			return false
		}
		// Check how the function is used to determine if 'this' defaults to global.
		return utils.IsDefaultThisBinding(container)

	default:
		// MethodDeclaration, Constructor, GetAccessor, SetAccessor,
		// PropertyDeclaration (field value), ClassStaticBlockDeclaration, etc.
		// In all these cases 'this' refers to the instance/class, not global.
		return false
	}
}
