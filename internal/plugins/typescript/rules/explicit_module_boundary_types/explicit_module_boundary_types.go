package explicit_module_boundary_types

import (
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/typescriptutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// options mirrors upstream's schema.
type options struct {
	allowArgumentsExplicitlyTypedAsAny        bool
	allowDirectConstAssertionInArrowFunctions bool
	allowedNames                              []string
	allowHigherOrderFunctions                 bool
	allowOverloadFunctions                    bool
	allowTypedFunctionExpressions             bool
}

func parseOptions(rawOpts any) options {
	opts := options{
		allowArgumentsExplicitlyTypedAsAny:        false,
		allowDirectConstAssertionInArrowFunctions: true,
		allowedNames:                  nil,
		allowHigherOrderFunctions:     true,
		allowOverloadFunctions:        false,
		allowTypedFunctionExpressions: true,
	}
	optsMap := utils.GetOptionsMap(rawOpts)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["allowArgumentsExplicitlyTypedAsAny"].(bool); ok {
		opts.allowArgumentsExplicitlyTypedAsAny = v
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
	if v, ok := optsMap["allowHigherOrderFunctions"].(bool); ok {
		opts.allowHigherOrderFunctions = v
	}
	if v, ok := optsMap["allowOverloadFunctions"].(bool); ok {
		opts.allowOverloadFunctions = v
	}
	if v, ok := optsMap["allowTypedFunctionExpressions"].(bool); ok {
		opts.allowTypedFunctionExpressions = v
	}
	return opts
}

var ExplicitModuleBoundaryTypesRule = rule.CreateRule(rule.Rule{
	Name: "explicit-module-boundary-types",
	// Upstream rule isn't type-aware (uses ESLint's scope manager) but a
	// TypeScript rule running on a TS project always has a TypeChecker
	// available, and the symbol-based scope walk we use here is strictly
	// more accurate than re-implementing scope analysis from scratch. We
	// require type info — gap files (no tsconfig project) are skipped by
	// the linter, matching how every other type-aware rule behaves.
	RequiresTypeInfo: true,
	Run:              run,
})

func run(ctx rule.RuleContext, _rawOptions []any) rule.RuleListeners {
	rawOptions := rule.LegacyUnwrapOptions(_rawOptions)
	opts := parseOptions(rawOptions)

	// Per-file state. rslint constructs a new RuleListeners for each file, so
	// these maps are isolated by source.
	checkedFunctions := make(map[*ast.Node]bool)
	alreadyVisited := make(map[*ast.Node]bool)
	functionReturnsMap := make(map[*ast.Node][]*ast.Node)
	// Reassignments-by-symbol, populated by the single pre-pass below and
	// consumed by followReference (instead of re-walking the AST per export
	// specifier). Mirrors ESLint scope manager's pre-computed
	// `variable.references`.
	assignmentsBySymbol := make(map[*ast.Symbol][]*ast.Node)

	// pendingReports collects all diagnostics emitted during the pass so we
	// can sort them by location before forwarding to ctx.Report*. ESLint's
	// RuleTester compares diagnostics in source-order (loc.start), not
	// emission order — multiple reports against the same function
	// (missingReturnType + missingArgType) would otherwise fail because our
	// checkReturnType-then-checkParameters flow emits return first.
	type pendingReport struct {
		rng core.TextRange
		msg rule.RuleMessage
	}
	pending := make([]pendingReport, 0, 8)

	reportRange := func(rng core.TextRange, msg rule.RuleMessage) {
		pending = append(pending, pendingReport{rng: rng, msg: msg})
	}
	reportNode := func(node *ast.Node, msg rule.RuleMessage) {
		reportRange(utils.TrimNodeTextRange(ctx.SourceFile, node), msg)
	}

	reportMissingReturn := func(node *ast.Node) {
		reportRange(functionHeadReportRange(ctx.SourceFile, node), rule.RuleMessage{
			Id:          "missingReturnType",
			Description: "Missing return type on function.",
		})
	}

	// reportNamed / reportUnnamed emit arg-type diagnostics for a parameter.
	// The reported range is always the parameter itself; the message variant
	// depends on whether the parameter has a usable name.
	reportNamed := func(target *ast.Node, messageId, name string) {
		var desc string
		switch messageId {
		case "missingArgType":
			desc = "Argument '" + name + "' should be typed."
		case "anyTypedArg":
			desc = "Argument '" + name + "' should be typed with a non-any type."
		}
		reportNode(target, rule.RuleMessage{
			Id:          messageId,
			Description: desc,
			Data:        map[string]string{"name": name},
		})
	}
	reportUnnamed := func(target *ast.Node, messageId, kindLabel string) {
		var desc string
		switch messageId {
		case "missingArgTypeUnnamed":
			desc = kindLabel + " argument should be typed."
		case "anyTypedArgUnnamed":
			desc = kindLabel + " argument should be typed with a non-any type."
		}
		reportNode(target, rule.RuleMessage{
			Id:          messageId,
			Description: desc,
			Data:        map[string]string{"type": kindLabel},
		})
	}

	// reportParam decides between the named and unnamed message variants.
	// `namedId`/`unnamedId` come from `(missingArgType, missingArgTypeUnnamed)`
	// or `(anyTypedArg, anyTypedArgUnnamed)`. The binding form (Identifier vs
	// array pattern vs object pattern) determines the variant, with a special
	// case for rest parameters whose inner binding is a pattern — upstream
	// labels those as "Rest", not "Array pattern" / "Object pattern".
	reportParam := func(reportNode *ast.Node, namedId, unnamedId string, bindingName *ast.Node, isRest bool) {
		switch bindingName.Kind {
		case ast.KindIdentifier:
			reportNamed(reportNode, namedId, bindingName.AsIdentifier().Text)
		case ast.KindArrayBindingPattern:
			if isRest {
				reportUnnamed(reportNode, unnamedId, "Rest")
			} else {
				reportUnnamed(reportNode, unnamedId, "Array pattern")
			}
		case ast.KindObjectBindingPattern:
			if isRest {
				reportUnnamed(reportNode, unnamedId, "Rest")
			} else {
				reportUnnamed(reportNode, unnamedId, "Object pattern")
			}
		}
	}

	// checkParameter mirrors upstream's `checkParameter`.
	checkParameter := func(param *ast.Node) {
		if param == nil || param.Kind != ast.KindParameter {
			return
		}
		pd := param.AsParameterDeclaration()

		// AssignmentPattern (default-value parameter): upstream silently
		// ignores these because the assignment provides a type via inference.
		// In tsgo, default-value parameters are Parameter nodes whose
		// Initializer is non-nil.
		if pd.Initializer != nil {
			return
		}

		nameNode := pd.Name()
		if nameNode == nil {
			return
		}

		isRest := pd.DotDotDotToken != nil
		// Parameter property (`constructor(public foo)`): upstream recurses
		// with the inner Identifier and reports on it, so the highlight covers
		// only `foo`, not the access modifier. Mirror that by switching the
		// report range to the binding name.
		reportTarget := param
		if !isRest && ast.HasSyntacticModifier(param, ast.ModifierFlagsParameterPropertyModifier) {
			reportTarget = nameNode
		}
		if pd.Type == nil {
			reportParam(reportTarget, "missingArgType", "missingArgTypeUnnamed", nameNode, isRest)
			return
		}
		if !opts.allowArgumentsExplicitlyTypedAsAny && pd.Type.Kind == ast.KindAnyKeyword {
			reportParam(reportTarget, "anyTypedArg", "anyTypedArgUnnamed", nameNode, isRest)
			return
		}
	}

	checkParameters := func(node *ast.Node) {
		for _, p := range node.Parameters() {
			checkParameter(p)
		}
	}

	// isAllowedName mirrors upstream's `isAllowedName`. It receives the
	// "function's owning declaration" node — for FunctionExpression /
	// ArrowFunction that's the VariableDeclaration / PropertyDeclaration /
	// PropertyAssignment / MethodDefinition equivalent; for FunctionDeclaration
	// it's the declaration itself. We use the same lookup as the parent.
	isAllowedName := func(node *ast.Node) bool {
		if node == nil || len(opts.allowedNames) == 0 {
			return false
		}

		switch node.Kind {
		case ast.KindVariableDeclaration, ast.KindFunctionDeclaration:
			id := node.Name()
			if id == nil || id.Kind != ast.KindIdentifier {
				return false
			}
			name := id.AsIdentifier().Text
			for _, n := range opts.allowedNames {
				if n == name {
					return true
				}
			}
			return false
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor,
			ast.KindPropertyDeclaration:
			// Upstream's `isAllowedName` for class members goes through
			// `isStaticMemberAccessOfValue`, which canonicalises the key
			// (Identifier name / quoted string / computed-literal / `null`
			// keyword) and compares against `allowedNames`. AccessorProperty
			// is a PropertyDeclaration with `accessor` modifier in tsgo;
			// abstract methods are MethodDeclaration with `abstract`.
			return matchesAllowedStaticMember(ctx, node, opts.allowedNames)
		case ast.KindPropertyAssignment:
			// `export const foo = { func2() { ... } }` — upstream treats the
			// inner method as `Property` with `method: true`; in tsgo
			// `func2() { ... }` is a MethodDeclaration directly inside the
			// object, not wrapped in PropertyAssignment. But when the value
			// IS an ArrowFunction / FunctionExpression, the wrapper is a
			// PropertyAssignment whose Name() is the key.
			return matchesAllowedStaticMember(ctx, node, opts.allowedNames)
		}
		return false
	}

	// checkFunction handles a FunctionDeclaration body.
	var checkFunction func(node *ast.Node)
	// checkFunctionExpression handles ArrowFunction / FunctionExpression.
	var checkFunctionExpression func(node *ast.Node)
	// checkClassMember handles a class element (MethodDeclaration, PropertyDeclaration, ...).
	var checkClassMember func(node *ast.Node)
	// checkBodyless handles a body-less method (overload signature / abstract method).
	var checkBodyless func(node *ast.Node)
	// checkNode dispatches by node kind. Mirrors upstream's `checkNode`.
	var checkNode func(node *ast.Node)

	// followReference mirrors upstream's `followReference`. Upstream uses
	// ESLint's scope manager (`scope.set.get(name)` + `variable.references`);
	// we use tsgo's TypeChecker, which resolves the same way through the
	// program's symbol table — alias-aware, shadowing-aware, and correctly
	// scoped without a hand-rolled walk.
	//
	// Definition filter: upstream skips CatchClause, ImplicitGlobalVariable,
	// ImportBinding, and Parameter — categories where the value type is
	// supplied by something other than the binding (catch error type, global
	// declaration, import contract, function signature). We mirror those
	// skips by inspecting the declaration's parent chain.
	followReference := func(node *ast.Node) {
		if node == nil || node.Kind != ast.KindIdentifier {
			return
		}
		// ShorthandPropertyAssignment.Name (in `{ foo }`) IS the property
		// declaration name from tsgo's POV, so GetSymbolAtLocation would
		// return the property symbol — not the outer variable that `foo`
		// references. Use GetShorthandAssignmentValueSymbol to get the
		// variable symbol the upstream rule actually wants to follow.
		var sym *ast.Symbol
		if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
			sym = ctx.TypeChecker.GetShorthandAssignmentValueSymbol(node.Parent)
		} else {
			sym = ctx.TypeChecker.GetSymbolAtLocation(node)
		}
		if sym == nil {
			return
		}
		// Resolve through alias chains. `export { name }`'s identifier resolves
		// to an export-alias symbol whose target is the underlying variable;
		// we need the variable's declarations to check, not the specifier's.
		// Import aliases resolve the same way, but their declarations are
		// filtered out by `shouldCheckDefinition` (matching upstream's
		// ImportBinding skip).
		if sym.Flags&ast.SymbolFlagsAlias != 0 {
			if resolved := ctx.TypeChecker.SkipAlias(sym); resolved != nil {
				sym = resolved
			}
		}
		for _, decl := range sym.Declarations {
			if !shouldCheckDefinition(decl) {
				continue
			}
			// Cross-file declarations: a symbol can resolve to declarations in
			// another source file (re-exports, declaration-merging, ambient
			// modules). ctx.ReportRange uses ctx.SourceFile to compute
			// line/column, so reporting on a node whose Pos belongs to a
			// different file would panic with "slice out of range". Restrict
			// to the current file — upstream's per-file scope walk has the
			// same effect.
			if ast.GetSourceFileOfNode(decl) != ctx.SourceFile {
				continue
			}
			checkNode(decl)
		}
		// Walk reassignments — `name = expr` anywhere the same symbol is
		// referenced as the assignment target. Upstream iterates
		// `variable.references` and reports on each `writeExpr`. We mirror
		// by scanning BinaryExpressions whose LHS resolves to the same
		// symbol (TypeChecker handles shadowing — a same-named inner-block
		// `let` produces a different symbol).
		for _, write := range assignmentsBySymbol[sym] {
			checkNode(write)
		}
	}

	checkFunction = func(node *ast.Node) {
		if checkedFunctions[node] {
			return
		}
		checkedFunctions[node] = true

		// Body-less function declarations (overload signatures, `declare
		// function`) map to TSDeclareFunction in ESTree. Upstream's
		// `checkNode` has no case for TSDeclareFunction, so these are
		// SKIPPED entirely — neither return type nor parameters are
		// inspected. tsgo collapses both forms into KindFunctionDeclaration
		// distinguished by Body() == nil; mirror upstream's skip here.
		if node.Body() == nil {
			return
		}

		if isAllowedName(node) || typescriptutil.AncestorHasReturnType(node) {
			return
		}
		if opts.allowOverloadFunctions && hasOverloadSignatures(ctx, node) {
			return
		}
		if !isValidFunctionReturnType(node, functionReturnsMap[node], opts) {
			reportMissingReturn(node)
		}
		checkParameters(node)
	}

	checkFunctionExpression = func(node *ast.Node) {
		if checkedFunctions[node] {
			return
		}
		checkedFunctions[node] = true

		if isAllowedName(node.Parent) ||
			typescriptutil.IsTypedFunctionExpression(node, opts.allowTypedFunctionExpressions) ||
			typescriptutil.AncestorHasReturnType(node) {
			return
		}
		if opts.allowOverloadFunctions && node.Parent != nil &&
			node.Parent.Kind == ast.KindMethodDeclaration &&
			hasOverloadSignatures(ctx, node.Parent) {
			return
		}
		if !isValidFunctionExpressionReturnTypeForRule(node, functionReturnsMap[node], opts) {
			reportMissingReturn(node)
		}
		checkParameters(node)
	}

	checkBodyless = func(node *ast.Node) {
		// Body-less methods: overload signatures, abstract method declarations.
		// Upstream's TSEmptyBodyFunctionExpression handler. Skip the return-type
		// check for constructors / set-accessors; always check parameters.
		isConstructor := node.Kind == ast.KindConstructor
		isSetAccessor := node.Kind == ast.KindSetAccessor
		if !isConstructor && !isSetAccessor && node.Type() == nil {
			// Upstream reports on the TSEmptyBodyFunctionExpression node,
			// which spans from `(` to (typically) the trailing `;`. tsgo
			// doesn't model the empty body separately — emulate the range by
			// scanning from the method name's end to the node end.
			reportRange(bodylessReportRange(ctx.SourceFile, node), rule.RuleMessage{
				Id:          "missingReturnType",
				Description: "Missing return type on function.",
			})
		}
		checkParameters(node)
	}

	checkClassMember = func(node *ast.Node) {
		// Filter out private and #private members. AccessorProperty is a
		// PropertyDeclaration with the `accessor` modifier in tsgo, so the
		// same handling applies.
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsPrivate) {
			return
		}
		name := node.Name()
		if name != nil && name.Kind == ast.KindPrivateIdentifier {
			return
		}

		switch node.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
			if node.Body() == nil {
				// Body-less method (overload signature or abstract method).
				checkBodyless(node)
				return
			}
			// In tsgo the method IS the function-like; upstream walks
			// `MethodDefinition.value` which is the inner FunctionExpression.
			// We dispatch through the same expression-style checks.
			checkMethodLikeBody(node, functionReturnsMap, checkedFunctions, opts, ctx, reportMissingReturn, checkParameters, isAllowedName)
		case ast.KindPropertyDeclaration:
			pd := node.AsPropertyDeclaration()
			if pd.Initializer != nil {
				checkNode(pd.Initializer)
			}
		}
	}

	checkNode = func(node *ast.Node) {
		if node == nil || alreadyVisited[node] {
			return
		}
		alreadyVisited[node] = true

		switch node.Kind {
		case ast.KindArrowFunction, ast.KindFunctionExpression:
			checkFunctionExpression(node)
		case ast.KindArrayLiteralExpression:
			if elems := node.AsArrayLiteralExpression().Elements; elems != nil {
				for _, el := range elems.Nodes {
					checkNode(el)
				}
			}
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor,
			ast.KindPropertyDeclaration:
			checkClassMember(node)
		case ast.KindClassDeclaration, ast.KindClassExpression:
			for _, member := range classMembers(node) {
				checkNode(member)
			}
		case ast.KindFunctionDeclaration:
			checkFunction(node)
		case ast.KindIdentifier:
			followReference(node)
		case ast.KindObjectLiteralExpression:
			if props := node.AsObjectLiteralExpression().Properties; props != nil {
				for _, prop := range props.Nodes {
					checkNode(prop)
				}
			}
		case ast.KindPropertyAssignment:
			pa := node.AsPropertyAssignment()
			if pa.Initializer != nil {
				checkNode(pa.Initializer)
			}
		case ast.KindShorthandPropertyAssignment:
			// `export default { foo }` — recurse into the binding.
			spa := node.AsShorthandPropertyAssignment()
			if spa.Name() != nil {
				checkNode(spa.Name())
			}
		case ast.KindVariableStatement:
			// Wrapped in a list. Upstream's VariableDeclaration → iterate
			// declarators.
			for _, decl := range variableStatementDeclarators(node) {
				checkNode(decl)
			}
		case ast.KindVariableDeclaration:
			vd := node.AsVariableDeclaration()
			if vd.Initializer != nil {
				checkNode(vd.Initializer)
			}
		}
	}

	// isExportedHigherOrderFunction mirrors upstream's
	// `isExportedHigherOrderFunction`. It walks up the parent chain — through
	// ReturnStatements / wrapping functions — to see if a checked exported
	// function ultimately encloses `node`.
	//
	// AST-shape note: upstream's `isFunction` accepts FunctionDeclaration /
	// FunctionExpression / ArrowFunction — and method bodies are reached
	// through the FunctionExpression child of MethodDefinition, so the walk
	// naturally lands on a Function*Expression. tsgo collapses method-likes
	// (MethodDeclaration / GetAccessor / SetAccessor / Constructor) into the
	// function-like itself with no wrapper, so the parent walk lands directly
	// on the method node — we must accept those kinds here too, otherwise
	// `class X { method() { return function () {} } }` exported via the class
	// would never detect the inner function as a higher-order child.
	isFunctionLikeForExport := func(n *ast.Node) bool {
		if typescriptutil.IsFunction(n) {
			return true
		}
		switch n.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
			return true
		}
		return false
	}
	isExportedHigherOrderFunction := func(node *ast.Node) bool {
		current := node.Parent
		for current != nil {
			if current.Kind == ast.KindReturnStatement {
				// upstream skips the wrapping Block — `current.parent.parent`.
				if current.Parent == nil || current.Parent.Parent == nil {
					return false
				}
				current = current.Parent.Parent
				continue
			}
			if !isFunctionLikeForExport(current) {
				return false
			}
			returns := functionReturnsMap[current]
			if !typescriptutil.DoesImmediatelyReturnFunctionExpression(current, returns) {
				return false
			}
			if checkedFunctions[current] {
				return true
			}
			current = current.Parent
		}
		return false
	}

	// Single pre-pass over the AST: collect every function's returns AND every
	// `name = expr` reassignment's symbol-keyed RHS. Both are needed by
	// followReference / isExportedHigherOrderFunction, and doing them in one
	// traversal avoids N-fold re-walks of the file when there are many
	// `export { ... }` specifiers (each followReference would otherwise scan
	// the entire AST). Mirrors how ESLint's scope manager pre-computes
	// `variable.references` once and looks them up by symbol.
	collectMetadata(ctx, ctx.SourceFile.AsNode(), functionReturnsMap, assignmentsBySymbol, nil)

	// Drive checks via export-triggered entry points.
	walkExports(ctx.SourceFile, checkNode, followReference)

	// After all directly-exported functions are checked, scan every function
	// in the map; if it's a higher-order child of a checked export, check it
	// too. Mirrors upstream's `Program:exit` loop. We iterate in source-order
	// (by node position) so the higher-order discovery is deterministic across
	// runs — Go map iteration is randomised.
	higherOrderCandidates := make([]*ast.Node, 0, len(functionReturnsMap))
	for fn := range functionReturnsMap {
		higherOrderCandidates = append(higherOrderCandidates, fn)
	}
	sort.Slice(higherOrderCandidates, func(i, j int) bool {
		return higherOrderCandidates[i].Pos() < higherOrderCandidates[j].Pos()
	})
	for _, fn := range higherOrderCandidates {
		if isExportedHigherOrderFunction(fn) {
			checkNode(fn)
		}
	}

	// Sort pending diagnostics by display-range start so the order matches
	// ESLint's RuleTester (which compares by `loc.start.line`,
	// `loc.start.column`). The display range comes from `getFunctionHeadLoc`
	// for missingReturnType and from the param node for missingArgType —
	// `functionHeadReportRange` already handles the AccessorProperty carve-
	// out, so simple range-based sorting is enough.
	sort.SliceStable(pending, func(i, j int) bool {
		if pending[i].rng.Pos() != pending[j].rng.Pos() {
			return pending[i].rng.Pos() < pending[j].rng.Pos()
		}
		return pending[i].rng.End() < pending[j].rng.End()
	})
	for _, p := range pending {
		ctx.ReportRange(p.rng, p.msg)
	}

	// No live listeners — all work happens eagerly above. Returning nil is
	// not currently supported, so we return an empty map.
	return rule.RuleListeners{}
}

// functionHeadReportRange returns the range used as the diagnostic's display
// range for a missingReturnType report. For most function-likes it forwards
// to utils.GetFunctionHeadLoc — the existing helper that mirrors ESLint's
// `getFunctionHeadLoc`. The one carve-out is AccessorProperty (tsgo:
// PropertyDeclaration with the `accessor` modifier): upstream's
// `getFunctionHeadLoc` has a dedicated PropertyDefinition case but NOT an
// AccessorProperty case, so AccessorProperty arrows fall back to the arrow's
// own loc (`=>` token range), not the property header. Mirror that so the
// sorted order matches upstream when an accessor field's arrow lacks parens
// (`accessor bool = arg => body` reports arg before the head, but
// `bool = arg => body` reports the head first).
func functionHeadReportRange(sf *ast.SourceFile, node *ast.Node) core.TextRange {
	if node.Kind == ast.KindArrowFunction && node.Parent != nil &&
		node.Parent.Kind == ast.KindPropertyDeclaration &&
		ast.HasSyntacticModifier(node.Parent, ast.ModifierFlagsAccessor) {
		af := node.AsArrowFunction()
		if af.EqualsGreaterThanToken != nil {
			arrowRange := scanner.GetRangeOfTokenAtPosition(sf, af.EqualsGreaterThanToken.Pos())
			return core.NewTextRange(arrowRange.Pos(), arrowRange.End())
		}
	}
	return utils.GetFunctionHeadLoc(sf, node)
}

// bodylessReportRange constructs the report range that upstream's
// TSEmptyBodyFunctionExpression occupies: from the opening `(` of the
// parameter list through the end of the declaration (typically `;`).
func bodylessReportRange(sf *ast.SourceFile, node *ast.Node) core.TextRange {
	searchFrom := node.Pos()
	if name := node.Name(); name != nil {
		searchFrom = name.End()
	}
	end := node.End()
	s := scanner.GetScannerForSourceFile(sf, searchFrom)
	for s.TokenStart() < end {
		if s.Token() == ast.KindOpenParenToken {
			return core.NewTextRange(s.TokenStart(), end)
		}
		s.Scan()
	}
	return utils.TrimNodeTextRange(sf, node)
}

// classMembers returns the members of a class declaration/expression.
func classMembers(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindClassDeclaration:
		members := node.AsClassDeclaration().Members
		if members == nil {
			return nil
		}
		return members.Nodes
	case ast.KindClassExpression:
		members := node.AsClassExpression().Members
		if members == nil {
			return nil
		}
		return members.Nodes
	}
	return nil
}

// variableStatementDeclarators returns the VariableDeclaration nodes inside a
// VariableStatement's declaration list.
func variableStatementDeclarators(node *ast.Node) []*ast.Node {
	vs := node.AsVariableStatement()
	if vs.DeclarationList == nil {
		return nil
	}
	list := vs.DeclarationList.AsVariableDeclarationList()
	if list == nil || list.Declarations == nil {
		return nil
	}
	return list.Declarations.Nodes
}

// isValidFunctionReturnType mirrors upstream's `isValidFunctionReturnType`.
func isValidFunctionReturnType(node *ast.Node, returns []*ast.Node, opts options) bool {
	if opts.allowHigherOrderFunctions && typescriptutil.DoesImmediatelyReturnFunctionExpression(node, returns) {
		return true
	}
	if node.Type() != nil {
		return true
	}
	// Constructor / set-accessor never need a return type.
	if node.Kind == ast.KindConstructor || node.Kind == ast.KindSetAccessor {
		return true
	}
	// Constructors in ESLint go through MethodDefinition; in tsgo we check
	// directly via node.Kind above.
	return false
}

// isValidFunctionExpressionReturnTypeForRule mirrors upstream's
// `checkFunctionExpressionReturnType` short-circuit, then falls back to
// `checkFunctionReturnType`. Returns true if no diagnostic is needed.
func isValidFunctionExpressionReturnTypeForRule(node *ast.Node, returns []*ast.Node, opts options) bool {
	if typescriptutil.IsValidFunctionExpressionReturnType(
		node,
		opts.allowTypedFunctionExpressions,
		false, // allowExpressions: explicit-module-boundary-types has no allowExpressions option
		opts.allowDirectConstAssertionInArrowFunctions,
	) {
		return true
	}
	return isValidFunctionReturnType(node, returns, opts)
}

// checkMethodLikeBody is the method-equivalent of checkFunctionExpression.
// Methods in tsgo are MethodDeclaration / GetAccessor / SetAccessor /
// Constructor — distinct kinds from FunctionExpression. They follow the same
// validity gates but operate on a different parent (the class, not a
// PropertyDefinition).
func checkMethodLikeBody(
	node *ast.Node,
	functionReturnsMap map[*ast.Node][]*ast.Node,
	checkedFunctions map[*ast.Node]bool,
	opts options,
	ctx rule.RuleContext,
	reportMissingReturn func(*ast.Node),
	checkParameters func(*ast.Node),
	isAllowedName func(*ast.Node) bool,
) {
	if checkedFunctions[node] {
		return
	}
	checkedFunctions[node] = true

	// `isAllowedName(node)` against the method itself — matches upstream's
	// AllowedName check for MethodDefinition.
	if isAllowedName(node) {
		return
	}
	if typescriptutil.AncestorHasReturnType(node) {
		return
	}
	// Object-literal method shorthand (`{ foo() {} }`) inside a typed parent
	// (variable annotation, type assertion, …) is "typed function expression"
	// territory upstream: ESLint sees the method as `Property > FunctionExpression`
	// and isTypedFunctionExpression climbs through the ObjectExpression's
	// parent. In tsgo the MethodDeclaration is a direct child of the
	// ObjectLiteralExpression with no Property wrapper, so the standard
	// isTypedFunctionExpression check (which expects a function-expression
	// parent) doesn't fire. Bridge that here.
	if opts.allowTypedFunctionExpressions &&
		node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
		parent := typescriptutil.GetEffectiveParent(node.Parent)
		if parent != nil &&
			(typescriptutil.IsTypedParent(parent, node) ||
				typescriptutil.IsPropertyOfObjectWithType(parent, node)) {
			return
		}
	}
	if opts.allowOverloadFunctions && hasOverloadSignatures(ctx, node) {
		return
	}
	// Constructor/set-accessor: skip return-type check; always check params.
	if node.Kind == ast.KindConstructor || node.Kind == ast.KindSetAccessor {
		checkParameters(node)
		return
	}
	if !isValidFunctionReturnType(node, functionReturnsMap[node], opts) {
		reportMissingReturn(node)
	}
	checkParameters(node)
}

// collectMetadata walks `node` recursively, populating two indexes used by
// the rest of the rule:
//
//   - returnsMap[fn] = []ReturnStatement — every return statement nested
//     inside each function-like (drives `doesImmediatelyReturnFunctionExpression`).
//   - assignmentsMap[sym] = []*ast.Node — the RHS expression of every
//     `name = expr` BinaryExpression where the LHS resolves to that symbol
//     (drives `followReference`'s walk of reassignments).
//
// Doing both in one pass keeps the traversal at O(file size) regardless of
// how many `export { ... }` specifiers the file has.
func collectMetadata(
	ctx rule.RuleContext,
	node *ast.Node,
	returnsMap map[*ast.Node][]*ast.Node,
	assignmentsMap map[*ast.Symbol][]*ast.Node,
	currentFn *ast.Node,
) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		if _, ok := returnsMap[node]; !ok {
			returnsMap[node] = nil
		}
		currentFn = node
	case ast.KindReturnStatement:
		if currentFn != nil {
			returnsMap[currentFn] = append(returnsMap[currentFn], node)
		}
	case ast.KindBinaryExpression:
		be := node.AsBinaryExpression()
		if be.OperatorToken != nil && be.OperatorToken.Kind == ast.KindEqualsToken {
			lhs := ast.SkipParentheses(be.Left)
			if lhs != nil && lhs.Kind == ast.KindIdentifier {
				if sym := ctx.TypeChecker.GetSymbolAtLocation(lhs); sym != nil {
					assignmentsMap[sym] = append(assignmentsMap[sym], be.Right)
				}
			}
		}
	}
	node.ForEachChild(func(child *ast.Node) bool {
		collectMetadata(ctx, child, returnsMap, assignmentsMap, currentFn)
		return false
	})
}

// walkExports drives `checkNode` from every export-shaped construct. We start
// at SourceFile.Statements and recurse into ModuleDeclaration bodies so that
// `export function foo()` inside an `export namespace NS { ... }` is reached
// the same way upstream's listeners traverse the AST.
func walkExports(sf *ast.SourceFile, checkNode func(*ast.Node), followReference func(*ast.Node)) {
	if sf == nil || sf.Statements == nil {
		return
	}
	walkStatements(sf.Statements.Nodes, checkNode, followReference)
}

func walkStatements(stmts []*ast.Node, checkNode func(*ast.Node), followReference func(*ast.Node)) {
	for _, stmt := range stmts {
		switch stmt.Kind {
		case ast.KindFunctionDeclaration:
			if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
				checkNode(stmt)
			}
		case ast.KindClassDeclaration:
			if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
				checkNode(stmt)
			}
		case ast.KindVariableStatement:
			if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
				for _, decl := range variableStatementDeclarators(stmt) {
					checkNode(decl)
				}
			}
		case ast.KindModuleDeclaration:
			// `(export) namespace NS { ... }` — recurse into its body to
			// reach nested `export function/class/...` declarations. The
			// namespace itself doesn't need an export modifier — even an
			// internal namespace's `export`-marked members are visible
			// through namespace access, mirroring upstream's traversal of
			// every ExportNamedDeclaration regardless of nesting.
			md := stmt.AsModuleDeclaration()
			if md.Body == nil {
				continue
			}
			body := md.Body
			switch body.Kind {
			case ast.KindModuleBlock:
				if stmts := body.AsModuleBlock().Statements; stmts != nil {
					walkStatements(stmts.Nodes, checkNode, followReference)
				}
			case ast.KindModuleDeclaration:
				// Nested namespace: `namespace A.B { ... }` desugars to
				// `namespace A { namespace B { ... } }`. Recurse.
				walkStatements([]*ast.Node{body}, checkNode, followReference)
			}
		case ast.KindExportAssignment:
			// Both `export default expr` and `export = expr` parse to
			// ExportAssignment in tsgo. Upstream has separate listeners
			// (ExportDefaultDeclaration vs TSExportAssignment) that both
			// call checkNode on the expression. The semantics are the same
			// for this rule — we forward unconditionally.
			ea := stmt.AsExportAssignment()
			if ea.Expression != nil {
				checkNode(ea.Expression)
			}
		case ast.KindExportDeclaration:
			ed := stmt.AsExportDeclaration()
			// `export { ... } from 'mod'` — has a module specifier; skip.
			if ed.ModuleSpecifier != nil {
				continue
			}
			if ed.ExportClause == nil || ed.ExportClause.Kind != ast.KindNamedExports {
				continue
			}
			ne := ed.ExportClause.AsNamedExports()
			if ne.Elements == nil {
				continue
			}
			for _, spec := range ne.Elements.Nodes {
				if spec.Kind != ast.KindExportSpecifier {
					continue
				}
				es := spec.AsExportSpecifier()
				// `export { foo as bar }` — local is `foo`, exported is `bar`.
				// Upstream follows `specifier.local`.
				local := es.PropertyName
				if local == nil {
					local = es.Name()
				}
				followReference(local)
			}
		}
	}
}

// shouldCheckDefinition mirrors upstream's `followReference` definition-type
// filter. Upstream drops CatchClause, ImplicitGlobalVariable, ImportBinding,
// and Parameter — categories where the value's type comes from somewhere
// other than the binding itself. tsgo doesn't tag declarations with the same
// DefinitionType enum, so we recognise the same categories by inspecting
// the declaration's syntactic position.
func shouldCheckDefinition(decl *ast.Node) bool {
	if decl == nil {
		return false
	}
	switch decl.Kind {
	case ast.KindParameter:
		// Definition.Parameter — skipped by upstream.
		return false
	case ast.KindImportClause, ast.KindImportSpecifier, ast.KindNamespaceImport,
		ast.KindImportEqualsDeclaration:
		// Definition.ImportBinding — skipped by upstream.
		return false
	case ast.KindBindingElement:
		// `catch ({ e })` / `function f({ x }) {}` — the binding lives
		// inside a CatchClause or Parameter; walk up to classify.
		for p := decl.Parent; p != nil; p = p.Parent {
			if p.Kind == ast.KindCatchClause || p.Kind == ast.KindParameter {
				return false
			}
			if p.Kind == ast.KindVariableDeclaration {
				return true
			}
		}
		return false
	}
	// Definition.CatchClause — `catch (e)` where `e` is the VariableDeclaration
	// inside the CatchClause node.
	if decl.Kind == ast.KindVariableDeclaration {
		for p := decl.Parent; p != nil; p = p.Parent {
			if p.Kind == ast.KindCatchClause {
				return false
			}
			if p.Kind == ast.KindVariableStatement || p.Kind == ast.KindSourceFile {
				break
			}
		}
	}
	return true
}

// hasOverloadSignatures reports whether `node` is the implementation of an
// overloaded function/method. The TypeChecker bundles overload signatures and
// the implementation under a single symbol — if any sibling declaration on
// the same symbol is body-less, this is the implementation of an overload.
//
// Anonymous `export default function () {}` overloads can't be unified by
// symbol (they have no name), so a syntactic adjacency walk handles that
// shape as a separate branch.
func hasOverloadSignatures(ctx rule.RuleContext, node *ast.Node) bool {
	if nameNode := node.Name(); nameNode != nil {
		if sym := ctx.TypeChecker.GetSymbolAtLocation(nameNode); sym != nil {
			for _, decl := range sym.Declarations {
				if decl == node {
					continue
				}
				switch decl.Kind {
				case ast.KindFunctionDeclaration, ast.KindMethodDeclaration:
					if decl.Body() == nil {
						return true
					}
				}
			}
			return false
		}
	}
	// Anonymous `export default function () { ... }` overloads — neither the
	// implementation nor its preceding overloads have a name, so the
	// TypeChecker can't link them by symbol. Fall back to a syntactic
	// adjacency walk over top-level statements.
	if node.Kind == ast.KindFunctionDeclaration && node.Name() == nil &&
		node.Parent != nil && node.Parent.Kind == ast.KindSourceFile {
		stmts := node.Parent.AsSourceFile().Statements
		if stmts == nil {
			return false
		}
		for _, sib := range stmts.Nodes {
			if sib == node || sib.Kind != ast.KindFunctionDeclaration {
				continue
			}
			if sib.Body() != nil || sib.Name() != nil {
				continue
			}
			if !ast.HasSyntacticModifier(sib, ast.ModifierFlagsExport) ||
				!ast.HasSyntacticModifier(sib, ast.ModifierFlagsDefault) {
				continue
			}
			return true
		}
	}
	return false
}

// matchesAllowedStaticMember mirrors typescript-eslint's
// `isStaticMemberAccessOfValue` over a class member: identifier name match,
// or computed key with literal value matching, against the allowedNames list.
func matchesAllowedStaticMember(ctx rule.RuleContext, node *ast.Node, allowedNames []string) bool {
	name := node.Name()
	if name == nil {
		return false
	}
	resolved := resolvePropertyKey(name)
	if resolved == "" {
		return false
	}
	for _, n := range allowedNames {
		if n == resolved {
			return true
		}
	}
	return false
}

// resolvePropertyKey returns the canonical string form of a property name
// (Identifier, StringLiteral, NumericLiteral, ComputedPropertyName with a
// literal expression, or `null`).
func resolvePropertyKey(name *ast.Node) string {
	switch name.Kind {
	case ast.KindIdentifier:
		return name.AsIdentifier().Text
	case ast.KindStringLiteral:
		return name.AsStringLiteral().Text
	case ast.KindNoSubstitutionTemplateLiteral:
		return name.AsNoSubstitutionTemplateLiteral().Text
	case ast.KindNumericLiteral:
		return name.AsNumericLiteral().Text
	case ast.KindComputedPropertyName:
		expr := ast.SkipParentheses(name.AsComputedPropertyName().Expression)
		if expr == nil {
			return ""
		}
		switch expr.Kind {
		case ast.KindStringLiteral:
			return expr.AsStringLiteral().Text
		case ast.KindNoSubstitutionTemplateLiteral:
			return expr.AsNoSubstitutionTemplateLiteral().Text
		case ast.KindNumericLiteral:
			return expr.AsNumericLiteral().Text
		case ast.KindNullKeyword:
			return "null"
		}
	}
	return ""
}
