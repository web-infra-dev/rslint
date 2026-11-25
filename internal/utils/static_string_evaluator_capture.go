package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

type staticReferenceTransfer struct {
	value, site *ast.Node
	copied      bool
	binding     bool
}

type staticCopySource struct {
	literal, length *ast.Node
	hasLength       bool
	hasProperties   bool
	unknown         bool
}

type staticEscapeFacts struct {
	call                  *ast.Node
	collecting            bool
	explicit              map[*ast.Symbol]bool
	sites                 map[*ast.Symbol]map[*ast.Node]staticReferenceFlags
	lengthWrites          map[*ast.Symbol]map[*ast.Node]staticReferenceFlags
	lengthWriteBlocks     map[*ast.Symbol]*ast.Node
	untrackedLengthWrites map[*ast.Node]bool
	copySources           map[*ast.Node]staticCopySource
	captures              map[*ast.Symbol]staticEvalResult
	capturing             map[*ast.Symbol]bool
	ownCaptures           map[*ast.Symbol]*staticObjectValue
	// Completed per-file facts for property-read queries; never cache a
	// read-position-dependent result here.
	constructorMutations map[*ast.Symbol]bool
}

func newStaticEscapeFacts() *staticEscapeFacts {
	return &staticEscapeFacts{explicit: make(map[*ast.Symbol]bool), sites: make(map[*ast.Symbol]map[*ast.Node]staticReferenceFlags), captures: make(map[*ast.Symbol]staticEvalResult), capturing: make(map[*ast.Symbol]bool)}
}

func executionRoot(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindSourceFile, ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			return node
		}
		if node.Parent == nil {
			return nil
		}
		// Computed method names and decorator expressions execute outside
		// the method. Reuse the AST helper's container rules for those paths.
		node = ast.GetThisContainer(node, true, false)
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
	view.useReferenceFacts(e)
	view.capturePoint = initializer
	view.resolving[symbol] = true
	result := view.evalValue(initializer)
	literal := SkipAssertionsAndParens(initializer)
	freshContainer := literal.Kind == ast.KindObjectLiteralExpression || literal.Kind == ast.KindArrayLiteralExpression
	if !result.ok || !IsPrimitiveStaticValue(result.value) && (!freshContainer || !view.hasOnlyPrimitiveChildren(initializer)) {
		result = staticEvalResult{}
	} else if object, ok := result.value.(*staticObjectValue); ok && object.prototypeSet && object.prototype != nil {
		// The own primitive properties were copied at initialization. Retain
		// them separately, without making a snapshot of the live prototype
		// available to whole-object or inherited-property evaluation.
		own := *object
		own.prototype, own.prototypeSet = nil, false
		if e.escapeFacts.ownCaptures == nil {
			e.escapeFacts.ownCaptures = make(map[*ast.Symbol]*staticObjectValue)
		}
		e.escapeFacts.ownCaptures[symbol] = &own
		result = staticEvalResult{}
	}
	e.escapeFacts.captures[symbol] = result
	return result
}

func (e *StaticStringEvaluator) capturedOwnMember(node *ast.Node, key string) staticEvalResult {
	initializer, symbol, ok := e.resolveIdentifierInitializer(node)
	if !ok || e.resolving[symbol] || e.hasPropertyMutation(symbol) ||
		e.capturePoint != nil && initializer.End() > e.capturePoint.Pos() {
		return staticEvalResult{}
	}
	e.resolving[symbol] = true
	defer delete(e.resolving, symbol)
	e.capturedInitializer(initializer, symbol)
	if e.escapeFacts != nil {
		if own := e.escapeFacts.ownCaptures[symbol]; own != nil {
			if value, found := staticObjectOwnProperty(own, key); found {
				return staticEvalResult{value: value, ok: true}
			}
		}
	}
	return e.capturedOwnMember(initializer, key)
}

// Keep every completed reference fact together when constructing an evaluation
// view. Expression caches and recursion stacks remain local to the view.
func (e *StaticStringEvaluator) useReferenceFacts(source *StaticStringEvaluator) {
	e.referenceFlags = source.referenceFlags
	e.bindingTransfers = source.bindingTransfers
	e.storedReferences = source.storedReferences
	e.escapeFacts = source.escapeFacts
	e.arrayPrototypeMutated = source.arrayPrototypeMutated
	e.stringPrototypeMutated = source.stringPrototypeMutated
	e.referenceFlagsComputed = true
}

func (e *StaticStringEvaluator) recordBindingTransfer(target, value, site *ast.Node, copied bool) {
	symbol := e.referenceSymbol(SkipAssertionsAndParens(target))
	if symbol == nil {
		return
	}
	if e.bindingTransfers == nil {
		e.bindingTransfers = make(map[*ast.Symbol][]staticReferenceTransfer)
	}
	e.bindingTransfers[symbol] = append(e.bindingTransfers[symbol], staticReferenceTransfer{value: value, site: site, binding: true, copied: copied})
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

// A straight-line reassignment replaces the old binding. Resolve it at the
// identifier read so an alias retains the object captured before a reassignment.
// Conditional writes remain possible origins.
func (e *StaticStringEvaluator) latestBindingAt(symbol *ast.Symbol, point *ast.Node) *ast.Node {
	var latest *ast.Node
	for _, transfer := range e.bindingTransfers[symbol] {
		node := transfer.site
		if transfer.binding && node.End() <= point.Pos() && e.transferAvailableAt(transfer, point) && node.Kind == ast.KindBinaryExpression &&
			node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && node.Parent.Kind == ast.KindExpressionStatement &&
			node.Parent.Parent == statementContainer(point) && (latest == nil || latest.End() < node.End()) {
			latest = node
		}
	}
	return latest
}

func (e *StaticStringEvaluator) bindingTransferReplaced(transfer staticReferenceTransfer, latest *ast.Node) bool {
	if latest != nil && ast.IsBinaryExpression(latest) {
		left := SkipAssertionsAndParens(latest.AsBinaryExpression().Left)
		if ast.IsAssignmentPattern(left) && ast.IsNodeDescendantOf(transfer.site, left) {
			// A pattern's default executes after its right-hand side and is
			// part of this assignment, not an overwritten earlier binding.
			return false
		}
	}
	// A closure or deferred initializer can reestablish an earlier-written
	// binding after a later straight-line assignment has executed.
	return transfer.binding && latest != nil && e.propertyMutationFollowsRead(latest, latest, transfer.site)
}

func (e *StaticStringEvaluator) transferAvailableAt(transfer staticReferenceTransfer, point *ast.Node) bool {
	if transfer.site == point {
		return false
	}
	if point == nil || executionRoot(point) != executionRoot(transfer.site) {
		return true
	}
	end := point.Pos()
	switch point.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression:
		// Callee and argument evaluation finish before the invocation itself.
		// Do not include the invocation's own stores as earlier transfers.
		end = point.End()
	}
	if transfer.site.End() <= end {
		return true
	}
	// A non-call receiver is evaluated before its right-hand side. For
	// separate expressions, reuse the execution-order proof instead of
	// assuming that later source positions always execute later.
	return transfer.site.Pos() >= point.End() && !e.propertyMutationFollowsRead(transfer.site, point, point)
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
	if (builtin == "Object" || builtin == "Reflect") && name == "setPrototypeOf" {
		e.recordStoredReferences(args[0], args[1], call, false)
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
