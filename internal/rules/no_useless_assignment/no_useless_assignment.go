package no_useless_assignment

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessaryAssignment",
		Description: "The value assigned to '" + name + "' is not used in subsequent statements.",
	}
}

// https://eslint.org/docs/latest/rules/no-useless-assignment
var NoUselessAssignmentRule = rule.Rule{
	Name:   "no-useless-assignment",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if ctx.Refs == nil {
			return rule.RuleListeners{}
		}

		analyzed := false
		analyze := func(*ast.Node) {
			if analyzed {
				return
			}
			analyzed = true
			run(&ctx)
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration:    analyze,
			ast.KindBinaryExpression:       analyze,
			ast.KindPrefixUnaryExpression:  analyze,
			ast.KindPostfixUnaryExpression: analyze,
		}
	},
}

// rawAssignment is one syntactic write found in the file before it is known
// whether the variable it targets can be tracked.
type rawAssignment struct {
	identifier *ast.Node
	root       *ast.Node
	sym        *ast.Symbol
}

func run(ctx *rule.RuleContext) {
	sourceFile := ctx.SourceFile
	isModule := ast.IsExternalModule(sourceFile)
	var exportedNames map[string]bool
	if isModule {
		exportedNames = collectLocallyExportedNames(sourceFile)
	}

	byRoot := make(map[*ast.Node][]rawAssignment)
	var rootOrder []*ast.Node
	collectAssignments(ctx, sourceFile, func(raw rawAssignment) {
		if _, seen := byRoot[raw.root]; !seen {
			rootOrder = append(rootOrder, raw.root)
		}
		byRoot[raw.root] = append(byRoot[raw.root], raw)
	})

	var reports []*ast.Node
	for _, root := range rootOrder {
		reports = append(reports, analyzeRoot(ctx, root, byRoot[root], isModule, exportedNames)...)
	}

	// ESLint verifies one code path at a time and its reports are ordered by
	// position afterwards; match that so nested functions don't reorder output.
	slices.SortFunc(reports, func(a, b *ast.Node) int { return a.Pos() - b.Pos() })
	for _, identifier := range reports {
		ctx.ReportNode(identifier, buildMessage(identifier.Text()))
	}
}

// collectAssignments walks the file once and reports every variable write the
// rule considers: a declarator with an initializer, an assignment expression,
// and an update expression.
func collectAssignments(ctx *rule.RuleContext, sourceFile *ast.SourceFile, emit func(rawAssignment)) {
	add := func(target *ast.Node) {
		forEachTargetIdentifier(target, func(identifier *ast.Node) {
			root := rootOf(identifier)
			if root == nil {
				return
			}
			sym := symbolOfTarget(ctx, identifier)
			if sym == nil {
				return
			}
			emit(rawAssignment{identifier: identifier, root: root, sym: sym})
		})
	}

	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindVariableDeclaration:
			if decl := node.AsVariableDeclaration(); decl.Initializer != nil {
				add(decl.Name())
			}
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				add(binary.Left)
			}
		case ast.KindPrefixUnaryExpression:
			unary := node.AsPrefixUnaryExpression()
			if unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken {
				add(unary.Operand)
			}
		case ast.KindPostfixUnaryExpression:
			add(node.AsPostfixUnaryExpression().Operand)
		}
		node.ForEachChild(visit)
		return false
	}
	sourceFile.AsNode().ForEachChild(visit)
}

// forEachTargetIdentifier yields every identifier an assignment target writes
// to, mirroring ESLint's `extractIdentifiersFromPattern`. A target wrapped in
// a TS assertion — `x as T`, `x!`, `x satisfies T`, `<T>x` — yields nothing:
// upstream does not treat those wrappers as patterns, so such a write is
// invisible to the rule.
func forEachTargetIdentifier(node *ast.Node, cb func(*ast.Node)) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		cb(node)

	case ast.KindParenthesizedExpression:
		forEachTargetIdentifier(node.AsParenthesizedExpression().Expression, cb)

	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		pattern := node.AsBindingPattern()
		if pattern.Elements == nil {
			return
		}
		for _, element := range pattern.Elements.Nodes {
			if binding := element.AsBindingElement(); binding != nil {
				forEachTargetIdentifier(binding.Name(), cb)
			}
		}

	case ast.KindObjectLiteralExpression:
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			switch property.Kind {
			case ast.KindPropertyAssignment:
				forEachTargetIdentifier(property.AsPropertyAssignment().Initializer, cb)
			case ast.KindShorthandPropertyAssignment:
				forEachTargetIdentifier(property.AsShorthandPropertyAssignment().Name(), cb)
			case ast.KindSpreadAssignment:
				forEachTargetIdentifier(property.AsSpreadAssignment().Expression, cb)
			}
		}

	case ast.KindArrayLiteralExpression:
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			forEachTargetIdentifier(element, cb)
		}

	case ast.KindSpreadElement:
		forEachTargetIdentifier(node.AsSpreadElement().Expression, cb)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsToken {
			forEachTargetIdentifier(binary.Left, cb)
		}
	}
}

// symbolOfTarget resolves the variable an assignment target names. Declaration
// names carry their symbol directly; every other target is a reference the
// binder-backed index can resolve.
func symbolOfTarget(ctx *rule.RuleContext, identifier *ast.Node) *ast.Symbol {
	parent := identifier.Parent
	if parent != nil && parent.Name() == identifier {
		switch parent.Kind {
		case ast.KindVariableDeclaration, ast.KindBindingElement:
			return parent.Symbol()
		}
	}
	return ctx.Refs.Resolve(identifier)
}

func analyzeRoot(
	ctx *rule.RuleContext,
	root *ast.Node,
	raws []rawAssignment,
	isModule bool,
	exportedNames map[string]bool,
) []*ast.Node {
	assignByIdent := make(map[*ast.Node]*assignment, len(raws))
	readNodes := make(map[*ast.Node]*ast.Symbol)
	trackedState := make(map[*ast.Symbol]bool)
	var assignments []*assignment

	for _, raw := range raws {
		tracked, known := trackedState[raw.sym]
		if !known {
			tracked = isTrackable(ctx, raw.sym, root, isModule, exportedNames)
			if tracked {
				// The variable is only usable when every read of it happens in
				// this same code path: a read from a nested function may run at
				// any time, so no assignment to the variable can be proven
				// unused.
				reads := readReferences(ctx, raw.sym)
				if len(reads) == 0 {
					// An entirely unread variable is `no-unused-vars`' business.
					tracked = false
				} else {
					for _, read := range reads {
						if rootOf(read) != root {
							tracked = false
							break
						}
					}
				}
				if tracked {
					for _, read := range reads {
						readNodes[read] = raw.sym
					}
				}
			}
			trackedState[raw.sym] = tracked
		}
		if !tracked {
			continue
		}
		// An assignment inside a `try` block is never reported — the block may
		// be abandoned partway through, so the value can still matter — but it
		// still overwrites the variable for every other assignment's sake.
		a := &assignment{
			sym:        raw.sym,
			identifier: raw.identifier,
			silent:     inTryBlockOfRoot(raw.identifier, root),
		}
		assignByIdent[raw.identifier] = a
		assignments = append(assignments, a)
	}

	if len(assignments) == 0 {
		return nil
	}

	b := &builder{assignByIdent: assignByIdent, readNodes: readNodes}
	b.buildRoot(root)

	dead := deadWrites(b.blocks, assignments)
	var reports []*ast.Node
	for _, a := range assignments {
		if !a.silent && dead[a] {
			reports = append(reports, a.identifier)
		}
	}
	return reports
}

// buildRoot lays out the control-flow graph of one code path root.
func (b *builder) buildRoot(root *ast.Node) {
	b.cur = b.newBlock()
	b.cur.hasIncoming = true

	switch root.Kind {
	case ast.KindSourceFile:
		b.statements(root.AsSourceFile().Statements)

	case ast.KindClassStaticBlockDeclaration:
		b.statement(root.AsClassStaticBlockDeclaration().Body)

	case ast.KindPropertyDeclaration:
		b.expr(root.AsPropertyDeclaration().Initializer)

	default:
		for _, parameter := range root.Parameters() {
			b.parameter(parameter)
		}
		// A `typeof x` inside the signature's type annotations references the
		// variable even though nothing in them runs.
		b.expr(root.Type())
		body := root.Body()
		if body == nil {
			return
		}
		if body.Kind == ast.KindBlock {
			b.statement(body)
		} else {
			b.expr(body)
		}
	}
}

// parameter models a parameter binding: its default value is skipped when an
// argument was passed.
func (b *builder) parameter(node *ast.Node) {
	if b.cur == nil {
		return
	}
	parameter := node.AsParameterDeclaration()
	if parameter == nil {
		return
	}
	b.expr(parameter.Type)
	b.bindWithDefault(parameter.Name(), parameter.Initializer)
}

// isTrackable reports whether assignments to sym can be proven unused: the
// variable must be declared in this very code path and must not be reachable
// from outside the file through an export.
func isTrackable(
	ctx *rule.RuleContext,
	sym *ast.Symbol,
	root *ast.Node,
	isModule bool,
	exportedNames map[string]bool,
) bool {
	if len(sym.Declarations) == 0 {
		return false
	}
	declaredHere := false
	for _, decl := range sym.Declarations {
		name := decl.Name()
		if name == nil {
			continue
		}
		if rootOf(name) == root {
			declaredHere = true
			break
		}
	}
	if !declaredHere {
		return false
	}
	if isModule && root.Kind == ast.KindSourceFile && isExported(sym, exportedNames, root) {
		return false
	}
	return true
}

// isExported reports whether a module-level binding leaves the file, in which
// case a later assignment to it can still be observed elsewhere. Both an
// `export` modifier and an `export { ... }` re-export only apply to the
// declaration actually named at the top level of the file — a block-scoped
// variable that merely shares that name is a different binding.
func isExported(sym *ast.Symbol, exportedNames map[string]bool, root *ast.Node) bool {
	for _, decl := range sym.Declarations {
		// `export let x = 1` and `export let { x } = obj` carry the modifier
		// on the variable statement; climb to it from the declarator or the
		// binding element.
		target := decl
	climb:
		for target != nil {
			switch target.Kind {
			case ast.KindBindingElement, ast.KindObjectBindingPattern, ast.KindArrayBindingPattern,
				ast.KindVariableDeclaration, ast.KindVariableDeclarationList:
				target = target.Parent
			default:
				break climb
			}
		}
		if target == nil || target.Parent != root {
			continue
		}
		if ast.HasSyntacticModifier(target, ast.ModifierFlagsExport) {
			return true
		}
		if exportedNames[sym.Name] {
			return true
		}
	}
	return false
}

// collectLocallyExportedNames gathers the local names of `export { ... }`
// declarations, which re-export a binding without a modifier on it.
func collectLocallyExportedNames(sourceFile *ast.SourceFile) map[string]bool {
	var names map[string]bool
	for _, statement := range sourceFile.Statements.Nodes {
		if statement.Kind != ast.KindExportDeclaration {
			continue
		}
		export := statement.AsExportDeclaration()
		if export.ModuleSpecifier != nil || export.ExportClause == nil ||
			export.ExportClause.Kind != ast.KindNamedExports {
			continue
		}
		for _, element := range export.ExportClause.AsNamedExports().Elements.Nodes {
			specifier := element.AsExportSpecifier()
			local := specifier.PropertyName
			if local == nil {
				local = specifier.Name()
			}
			if local == nil || local.Kind != ast.KindIdentifier {
				continue
			}
			if names == nil {
				names = make(map[string]bool)
			}
			names[local.Text()] = true
		}
	}
	return names
}

// readReferences returns the identifiers that read sym. A compound assignment
// or an update expression reads its target before writing it, so those count
// as reads too.
func readReferences(ctx *rule.RuleContext, sym *ast.Symbol) []*ast.Node {
	var reads []*ast.Node
	for _, ref := range ctx.Refs.References(sym) {
		if !utils.IsWriteReference(ref) || isReadWriteReference(ref) {
			reads = append(reads, ref)
		}
	}
	return reads
}

func isReadWriteReference(node *ast.Node) bool {
	current := node
	parent := node.Parent
	for parent != nil {
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression,
			ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression:
			current = parent
			parent = parent.Parent
			continue
		}
		break
	}
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary.Left == current &&
			binary.OperatorToken.Kind != ast.KindEqualsToken &&
			ast.IsAssignmentOperator(binary.OperatorToken.Kind)
	case ast.KindPrefixUnaryExpression:
		unary := parent.AsPrefixUnaryExpression()
		return unary.Operand == current &&
			(unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken)
	case ast.KindPostfixUnaryExpression:
		return parent.AsPostfixUnaryExpression().Operand == current
	}
	return false
}

// inTryBlockOfRoot reports whether node sits in the `try` block of a statement
// belonging to the same code path.
func inTryBlockOfRoot(node *ast.Node, root *ast.Node) bool {
	previous := node
	for current := node.Parent; current != nil && current != root; current = current.Parent {
		if current.Kind == ast.KindTryStatement && current.AsTryStatement().TryBlock == previous {
			return true
		}
		previous = current
	}
	return false
}

// isCodePathRoot reports whether node owns a control-flow graph of its own.
func isCodePathRoot(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindSourceFile,
		ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindConstructor,
		ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindClassStaticBlockDeclaration:
		return true
	case ast.KindPropertyDeclaration:
		return node.AsPropertyDeclaration().Initializer != nil
	}
	return false
}

// rootOf returns the code path root node executes in.
func rootOf(node *ast.Node) *ast.Node {
	previous := node
	var decorator *ast.Node
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindDecorator {
			decorator = current
		}
		if current.Kind == ast.KindSourceFile {
			return current
		}
		if isCodePathRoot(current) && runsInsideRoot(current, previous, decorator) {
			return current
		}
		previous = current
	}
	return nil
}

// runsInsideRoot reports whether a direct child of a code path root runs in
// that root rather than in the surrounding one. A member's name runs where the
// member is declared, and so do the decorators on the member and on each of its
// parameters.
func runsInsideRoot(root *ast.Node, child *ast.Node, decorator *ast.Node) bool {
	switch root.Kind {
	case ast.KindPropertyDeclaration:
		return root.AsPropertyDeclaration().Initializer == child
	case ast.KindClassStaticBlockDeclaration:
		return root.AsClassStaticBlockDeclaration().Body == child
	}
	if root.Name() == child {
		return false
	}
	if decorator != nil && (decorator.Parent == root || decorator.Parent == child) {
		return false
	}
	return true
}
