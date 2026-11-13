package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

type staticReferenceTransfer struct {
	value, site *ast.Node
	copied      bool
	binding     bool
}

type staticEscapeFacts struct {
	call       *ast.Node
	collecting bool
	explicit   map[*ast.Symbol]bool
	sites      map[*ast.Symbol]map[*ast.Node]struct{}
	visited    map[*ast.Symbol]staticReferenceFlags
	captures   map[*ast.Symbol]staticEvalResult
	capturing  map[*ast.Symbol]bool
}

func newStaticEscapeFacts() *staticEscapeFacts {
	return &staticEscapeFacts{explicit: make(map[*ast.Symbol]bool), sites: make(map[*ast.Symbol]map[*ast.Node]struct{}), visited: make(map[*ast.Symbol]staticReferenceFlags), captures: make(map[*ast.Symbol]staticEvalResult), capturing: make(map[*ast.Symbol]bool)}
}

func executionRoot(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindSourceFile, ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			return node
		}
		node = node.Parent
	}
	return nil
}

// Capture immutable values at initialization, including primitive elements
// copied into a fresh literal container. Containers retaining object references
// use ordinary evaluation. Capture views share facts without another file scan.
func (e *StaticStringEvaluator) capturedInitializer(initializer *ast.Node, symbol *ast.Symbol) staticEvalResult {
	if !e.trackCallEscapes || e.capturePoint != nil && initializer.End() > e.capturePoint.Pos() {
		return staticEvalResult{}
	}
	e.ensureReferenceFlags()
	if e.escapeFacts == nil || e.escapeFacts.collecting || e.HasSideEffect(initializer, false) {
		return staticEvalResult{}
	}
	for parent := initializer.Parent; parent != nil && parent != executionRoot(initializer); parent = parent.Parent {
		if ast.IsIterationStatement(parent, false) {
			return staticEvalResult{}
		}
	}
	if result, found := e.escapeFacts.captures[symbol]; found {
		return result
	}
	if e.escapeFacts.capturing[symbol] {
		return staticEvalResult{}
	}
	e.escapeFacts.capturing[symbol] = true
	defer delete(e.escapeFacts.capturing, symbol)
	view := NewStaticStringEvaluatorWithReferenceResolver(e.typeChecker, e.sourceFile, e.referenceResolver)
	view.GlobalAccess = e.GlobalAccess
	view.resolveIdentifiers = e.resolveIdentifiers
	view.trackCallEscapes = true
	view.referenceFlagsComputed = true
	view.referenceFlags = e.referenceFlags
	view.storedReferences = e.storedReferences
	view.escapeFacts = e.escapeFacts
	view.capturePoint = initializer
	view.resolving[symbol] = true
	result := view.evalValue(initializer)
	literal := SkipAssertionsAndParens(initializer)
	freshContainer := literal.Kind == ast.KindObjectLiteralExpression || literal.Kind == ast.KindArrayLiteralExpression
	if !result.ok || !IsPrimitiveStaticValue(result.value) && (!freshContainer || !view.hasOnlyPrimitiveChildren(initializer)) {
		result = staticEvalResult{}
	}
	e.escapeFacts.captures[symbol] = result
	return result
}

func (e *StaticStringEvaluator) recordBindingTransfer(target, value, site *ast.Node) {
	symbol := e.referenceSymbol(SkipAssertionsAndParens(target))
	if symbol == nil {
		return
	}
	if e.storedReferences == nil {
		e.storedReferences = make(map[*ast.Symbol][]staticReferenceTransfer)
	}
	e.storedReferences[symbol] = append(e.storedReferences[symbol], staticReferenceTransfer{value: value, site: site, binding: true})
}

func statementContainer(node *ast.Node) *ast.Node {
	for node != nil && node.Parent != nil {
		if node.Parent.Kind == ast.KindBlock || node.Parent.Kind == ast.KindSourceFile {
			return node.Parent
		}
		node = node.Parent
	}
	return nil
}

// A straight-line reassignment replaces the old binding. Conditional writes
// remain possible origins, and transfers after the call cannot affect it.
func (e *StaticStringEvaluator) latestBindingTransfer(symbol *ast.Symbol) *ast.Node {
	if e.escapeFacts == nil || e.escapeFacts.call == nil {
		return nil
	}
	return e.latestBindingAt(symbol, e.escapeFacts.call)
}

func (e *StaticStringEvaluator) latestBindingAt(symbol *ast.Symbol, point *ast.Node) *ast.Node {
	var latest *ast.Node
	for _, transfer := range e.storedReferences[symbol] {
		node := transfer.site
		if transfer.binding && transferAvailableAt(transfer, point) && node.Kind == ast.KindBinaryExpression &&
			node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && node.Parent.Kind == ast.KindExpressionStatement &&
			node.Parent.Parent == statementContainer(point) && (latest == nil || latest.End() < node.End()) {
			latest = node
		}
	}
	return latest
}

func (e *StaticStringEvaluator) transferAvailable(transfer staticReferenceTransfer) bool {
	return transferAvailableAt(transfer, e.escapeFacts.call)
}

func transferAvailableAt(transfer staticReferenceTransfer, point *ast.Node) bool {
	return point == nil || executionRoot(point) != executionRoot(transfer.site) || transfer.site.End() <= point.Pos()
}

func (e *StaticStringEvaluator) recordBuiltinTransfers(call *ast.Node, builtin, name string) {
	if !e.trackCallEscapes {
		return
	}
	args := call.Arguments()
	if len(args) < 2 || ast.IsSpreadElement(args[0]) {
		return
	}
	if builtin == "Object" && name == "assign" {
		for _, source := range args[1:] {
			if ast.IsSpreadElement(source) {
				source = source.Expression()
			}
			e.recordStoredReferences(args[0], source, call, true)
		}
	}
	if (builtin == "Object" || builtin == "Reflect") && name == "defineProperty" && len(args) >= 3 {
		e.recordStoredReferences(args[0], args[2], call, true)
	}
	if builtin == "Object" && name == "defineProperties" {
		e.recordStoredReferences(args[0], args[1], call, true)
	}
	if builtin == "Reflect" && name == "set" && len(args) >= 3 {
		e.recordStoredReferences(args[0], args[2], call, false)
		if len(args) >= 4 {
			e.recordStoredReferences(args[3], args[2], call, false)
		}
	}
}
