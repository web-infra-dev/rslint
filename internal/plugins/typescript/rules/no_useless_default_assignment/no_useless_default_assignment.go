package no_useless_default_assignment

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoUselessDefaultAssignmentRule mirrors
// @typescript-eslint/no-useless-default-assignment.
//
// It flags default values (`= expr`) attached to destructuring binding
// elements or anonymous-function parameters whose underlying type cannot
// possibly be `undefined` — the default is unreachable. It also collapses
// `= undefined` defaults to the optional-syntax form on parameters that
// admit undefined.
//
// https://typescript-eslint.io/rules/no-useless-default-assignment
// Upstream source: packages/eslint-plugin/src/rules/no-useless-default-assignment.ts
var NoUselessDefaultAssignmentRule = rule.CreateRule(rule.Rule{
	Name:             "no-useless-default-assignment",
	RequiresTypeInfo: true,
	Run:              run,
})

type Options struct {
	allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing bool
}

func parseOptions(options any) Options {
	opts := Options{}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"].(bool); ok {
		opts.allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = v
	}
	return opts
}

func buildNoStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

func buildPreferOptionalSyntaxMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferOptionalSyntax",
		Description: "Using `= undefined` to make a parameter optional adds unnecessary runtime logic. Use the `?` optional syntax instead.",
	}
}

func buildUselessDefaultAssignmentMessage(kind string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "uselessDefaultAssignment",
		Description: "Default value is useless because the " + kind + " is not optional.",
		Data:        map[string]string{"type": kind},
	}
}

func buildUselessUndefinedMessage(kind string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "uselessUndefined",
		Description: "Default value is useless because it is undefined. Optional " + kind + "s are already undefined by default.",
		Data:        map[string]string{"type": kind},
	}
}

func run(ctx rule.RuleContext, _options []any) rule.RuleListeners {
	options := rule.LegacyUnwrapOptions(_options)
	if ctx.TypeChecker == nil {
		return rule.RuleListeners{}
	}

	opts := parseOptions(options)
	compilerOptions := ctx.Program.Options()
	isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(compilerOptions, compilerOptions.StrictNullChecks)
	if !isStrictNullChecks && !opts.allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing {
		ctx.ReportRange(core.NewTextRange(0, 0), buildNoStrictNullCheckMessage())
	}

	return rule.RuleListeners{
		ast.KindParameter: func(node *ast.Node) {
			initializer := node.AsParameterDeclaration().Initializer
			if initializer == nil {
				return
			}
			checkAssignmentPattern(ctx, node, initializer)
		},
		ast.KindBindingElement: func(node *ast.Node) {
			initializer := node.AsBindingElement().Initializer
			if initializer == nil {
				return
			}
			checkAssignmentPattern(ctx, node, initializer)
		},
	}
}

// checkAssignmentPattern is the tsgo equivalent of upstream's
// `checkAssignmentPattern(node)`. `node` is either a Parameter or a
// BindingElement that owns a non-nil Initializer (== ESLint's
// `AssignmentPattern.right`). `initializer` is the same field, already
// unwrapped so callers don't repeat the field access.
func checkAssignmentPattern(ctx rule.RuleContext, node *ast.Node, initializer *ast.Node) {
	// `= undefined` branch — mirrors upstream's
	// `node.right.type === Identifier && node.right.name === 'undefined'`.
	// Upstream's check is lexical (does not skip parens / TS wrappers); we
	// match that with IsUndefinedIdentifier's paren-only unwrap.
	if utils.IsUndefinedIdentifier(initializer) {
		if node.Kind == ast.KindParameter {
			param := node.AsParameterDeclaration()
			if param.Type != nil {
				typeFromAnno := checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, param.Type)
				if canBeUndefined(typeFromAnno) {
					reportPreferOptionalSyntax(ctx, node, initializer)
					return
				}
			}
		}
		reportUselessUndefined(ctx, node, initializer, typeLabel(node))
		return
	}

	switch node.Kind {
	case ast.KindParameter:
		checkParameterContextualType(ctx, node, initializer)
	case ast.KindBindingElement:
		checkBindingElement(ctx, node, initializer)
	}
}

// checkParameterContextualType handles the
// `parent === ArrowFunctionExpression || FunctionExpression` arm. In tsgo the
// parent function kinds that map to it are ArrowFunction / FunctionExpression
// / MethodDeclaration / Constructor / GetAccessor / SetAccessor — the
// non-anonymous FunctionDeclaration kind is intentionally excluded, matching
// upstream's early-return for top-level function declarations. For each, we
// ask the checker for a contextual signature and use its parameter type; if
// no contextual type is available the rule says nothing.
func checkParameterContextualType(ctx rule.RuleContext, paramNode *ast.Node, initializer *ast.Node) {
	parent := paramNode.Parent
	if parent == nil {
		return
	}
	if !isAnonymousFunctionLike(parent) {
		return
	}

	paramIndex := indexOfParameter(parent, paramNode)
	if paramIndex < 0 {
		return
	}

	contextualType := checker.Checker_getContextualType(ctx.TypeChecker, parent, checker.ContextFlagsNone)
	if contextualType == nil {
		return
	}

	signatures := utils.GetCallSignatures(ctx.TypeChecker, contextualType)
	if len(signatures) == 0 {
		return
	}
	contextualSig := signatures[0]
	// Upstream short-circuits when the contextual signature IS this very
	// function's own declaration — there is no "outside" contract to enforce.
	if checker.Signature_declaration(contextualSig) == parent {
		return
	}

	params := checker.Signature_parameters(contextualSig)
	if paramIndex >= len(params) {
		return
	}
	paramSymbol := params[paramIndex]
	if paramSymbol == nil {
		return
	}

	// Rest parameter on the contextual side — the user's parameter resolves
	// to the element type of an iterable, so a default value remains
	// meaningful per element. Skip.
	if paramSymbol.ValueDeclaration != nil && ast.IsParameterDeclaration(paramSymbol.ValueDeclaration) {
		decl := paramSymbol.ValueDeclaration.AsParameterDeclaration()
		if decl.DotDotDotToken != nil {
			return
		}
	}

	if utils.IsSymbolFlagSet(paramSymbol, ast.SymbolFlagsOptional) {
		return
	}

	paramType := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, paramSymbol)
	if utils.IsTypeParameter(paramType) {
		return
	}
	if canBeUndefined(paramType) {
		return
	}

	reportUselessDefaultAssignment(ctx, paramNode, initializer, "parameter")
}

// checkBindingElement handles the `parent === Property || ArrayPattern` arm:
// for an object-pattern element we look up the property in the source type;
// for an array-pattern element we resolve the tuple slot.
func checkBindingElement(ctx rule.RuleContext, beNode *ast.Node, initializer *ast.Node) {
	parent := beNode.Parent
	if parent == nil {
		return
	}
	switch parent.Kind {
	case ast.KindObjectBindingPattern:
		propType := getTypeOfBindingElementProperty(ctx, beNode)
		if propType == nil {
			return
		}
		if canBeUndefined(propType) {
			return
		}
		reportUselessDefaultAssignment(ctx, beNode, initializer, "property")
	case ast.KindArrayBindingPattern:
		sourceType := getSourceTypeForPattern(ctx, parent)
		if sourceType == nil {
			return
		}
		if !checker.IsTupleType(sourceType) {
			return
		}
		tupleArgs := checker.Checker_getTypeArguments(ctx.TypeChecker, sourceType)
		elementIndex := indexOfBindingElement(parent, beNode)
		if elementIndex < 0 || elementIndex >= len(tupleArgs) {
			return
		}
		elementType := tupleArgs[elementIndex]
		if canBeUndefined(elementType) {
			return
		}
		reportUselessDefaultAssignment(ctx, beNode, initializer, "property")
	}
}

// getTypeOfBindingElementProperty is the tsgo equivalent of upstream's
// `getTypeOfProperty(property)`. The binding element sits inside an
// ObjectBindingPattern; the source of that pattern is what carries the
// property whose type we want.
//
// The optional-symbol branch mirrors the upstream subtlety where an
// optional property destructured directly from a conditional initializer
// (`const { a = ... } = cond ? {a: ...} : {a: ...}`) is treated as
// non-optional iff the property appears in every branch of the conditional.
func getTypeOfBindingElementProperty(ctx rule.RuleContext, beNode *ast.Node) *checker.Type {
	objectPattern := beNode.Parent
	if objectPattern == nil {
		return nil
	}
	sourceType := getSourceTypeForPattern(ctx, objectPattern)
	if sourceType == nil {
		return nil
	}

	be := beNode.AsBindingElement()
	propertyName, ok := getPropertyName(be)
	if !ok {
		return nil
	}

	symbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, sourceType, propertyName)
	if symbol == nil {
		return nil
	}

	if utils.IsSymbolFlagSet(symbol, ast.SymbolFlagsOptional) {
		parent := objectPattern.Parent
		if parent != nil && parent.Kind == ast.KindVariableDeclaration {
			initializer := parent.AsVariableDeclaration().Initializer
			if initializer != nil && hasConditionalInitializer(objectPattern) {
				if !hasPropertyInAllBranches(initializer, propertyName) {
					return nil
				}
			}
		}
	}

	return checker.Checker_getTypeOfSymbol(ctx.TypeChecker, symbol)
}

// getSourceTypeForPattern walks upstream's recursive `getSourceTypeForPattern`,
// translated to tsgo's pattern containers. `pattern` is always a BindingPattern
// (Object/Array) — the binding name of some Parameter, VariableDeclaration,
// or BindingElement.
func getSourceTypeForPattern(ctx rule.RuleContext, pattern *ast.Node) *checker.Type {
	parent := pattern.Parent
	if parent == nil {
		return nil
	}

	switch parent.Kind {
	case ast.KindVariableDeclaration:
		vd := parent.AsVariableDeclaration()
		if vd.Initializer == nil {
			return nil
		}
		return ctx.TypeChecker.GetTypeAtLocation(vd.Initializer)
	case ast.KindParameter:
		return getParameterType(ctx, parent)
	case ast.KindBindingElement:
		beParent := parent.Parent
		if beParent == nil {
			return nil
		}
		switch beParent.Kind {
		case ast.KindObjectBindingPattern:
			return getTypeOfBindingElementProperty(ctx, parent)
		case ast.KindArrayBindingPattern:
			arrayType := getSourceTypeForPattern(ctx, beParent)
			if arrayType == nil {
				return nil
			}
			elementIndex := indexOfBindingElement(beParent, parent)
			return getArrayElementType(ctx, arrayType, elementIndex)
		}
	}
	return nil
}

// getParameterType resolves a destructuring pattern's source via the
// containing function's signature. Mirrors upstream's `isFunction(parent)`
// arm — uses `GetSignatureFromDeclaration` and indexes into the resulting
// parameter symbol list, accounting for `thisParameter` shifting the index
// (the signature parameters slice excludes the `this:` parameter, but the
// function declaration's parameter list includes it at index 0).
func getParameterType(ctx rule.RuleContext, paramNode *ast.Node) *checker.Type {
	funcNode := paramNode.Parent
	if funcNode == nil {
		return nil
	}
	paramIndex := indexOfParameter(funcNode, paramNode)
	if paramIndex < 0 {
		return nil
	}
	signature := ctx.TypeChecker.GetSignatureFromDeclaration(funcNode)
	if signature == nil {
		return nil
	}
	if ast.GetThisParameter(funcNode) != nil {
		paramIndex--
	}
	params := checker.Signature_parameters(signature)
	if paramIndex < 0 || paramIndex >= len(params) {
		return nil
	}
	return checker.Checker_getTypeOfSymbol(ctx.TypeChecker, params[paramIndex])
}

// getArrayElementType mirrors upstream's helper of the same name: tuple slots
// resolve to a fixed element type; non-tuple arrays fall back to the number
// index signature (so `Array<string>` reads as `string`).
func getArrayElementType(ctx rule.RuleContext, arrayType *checker.Type, elementIndex int) *checker.Type {
	if elementIndex < 0 {
		return nil
	}
	if checker.IsTupleType(arrayType) {
		tupleArgs := checker.Checker_getTypeArguments(ctx.TypeChecker, arrayType)
		if elementIndex < len(tupleArgs) {
			return tupleArgs[elementIndex]
		}
		return nil
	}
	numberType := checker.Checker_numberType(ctx.TypeChecker)
	return checker.Checker_getIndexTypeOfType(ctx.TypeChecker, arrayType, numberType)
}

// canBeUndefined is upstream's local helper. Lifts a type to its union
// constituents (a single non-union type is wrapped into a 1-element list by
// UnionTypeParts) and checks for the Undefined flag, plus the any/unknown
// short-circuits that suppress almost every type test in TS.
func canBeUndefined(t *checker.Type) bool {
	if utils.IsTypeAnyType(t) || utils.IsTypeUnknownType(t) {
		return true
	}
	for _, part := range utils.UnionTypeParts(t) {
		if utils.IsTypeFlagSet(part, checker.TypeFlagsUndefined) {
			return true
		}
	}
	return false
}

// hasConditionalInitializer walks up the BindingPattern's ancestors looking
// for the nearest enclosing VariableDeclaration WITH an Initializer that
// happens to be a ConditionalExpression or a LogicalExpression-shaped
// BinaryExpression (`a || b`, `a && b`, `a ?? b` — tsgo flattens these into
// BinaryExpression, where upstream tags them as LogicalExpression).
//
// Mirrors upstream's recursive structure exactly: a VariableDeclaration
// without an Initializer (e.g., the head of `for (const {x = 5} of arr)`)
// does NOT short-circuit — we keep walking up in case an outer
// VariableDeclaration carries the conditional init. In practice the outer
// walk almost always terminates without finding one, but matching upstream
// keeps any edge-case parity intact.
func hasConditionalInitializer(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindVariableDeclaration {
		if init := parent.AsVariableDeclaration().Initializer; init != nil {
			return isConditionalLike(init)
		}
	}
	return hasConditionalInitializer(parent)
}

func isConditionalLike(expr *ast.Node) bool {
	expr = ast.SkipParentheses(expr)
	if expr == nil {
		return false
	}
	if expr.Kind == ast.KindConditionalExpression {
		return true
	}
	if expr.Kind == ast.KindBinaryExpression {
		op := expr.AsBinaryExpression().OperatorToken.Kind
		switch op {
		case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
			return true
		}
	}
	return false
}

// hasPropertyInAllBranches recursively descends into a ConditionalExpression's
// `consequent` and `alternate` arms (and the tsgo equivalent for `&&`/`||`/`??`,
// which flatten to BinaryExpression). At each leaf, the value MUST be an
// ObjectExpression that defines `propertyName` as a non-spread, non-computed,
// non-method-shorthand property — anything looser would let upstream's symbol
// lookup miss the property, so we mirror its strictness.
func hasPropertyInAllBranches(expression *ast.Node, propertyName string) bool {
	expression = ast.SkipParentheses(expression)
	if expression == nil {
		return false
	}

	switch expression.Kind {
	case ast.KindObjectLiteralExpression:
		for _, prop := range expression.AsObjectLiteralExpression().Properties.Nodes {
			// Upstream filters by `prop.type === Property`, which in ESTree
			// covers all keyed forms: `a: 1` (PropertyAssignment), `a`
			// (ShorthandPropertyAssignment), `a() {}` (MethodDeclaration in
			// object-literal position), `get a() / set a()` (accessors).
			// SpreadAssignment is NOT Property in ESTree, so it's excluded.
			if prop.Kind == ast.KindSpreadAssignment {
				continue
			}
			name := prop.Name()
			if name == nil {
				continue
			}
			parsed, ok := utils.GetStaticPropertyName(name)
			if ok && parsed == propertyName {
				return true
			}
		}
		return false
	case ast.KindConditionalExpression:
		ce := expression.AsConditionalExpression()
		return hasPropertyInAllBranches(ce.WhenTrue, propertyName) &&
			hasPropertyInAllBranches(ce.WhenFalse, propertyName)
	}
	return false
}

// getPropertyName is the BindingElement counterpart of upstream's
// `getPropertyName(node.key)`. The "key" we examine is `PropertyName` when
// the binding was renamed (`{ foo: bar }`) and the binding `Name` (an
// Identifier) when it was shorthand (`{ foo }`).
//
// We delegate the actual literal-form recognition to
// `utils.GetStaticPropertyName`, which already covers every form upstream
// recognizes (Identifier / StringLiteral / NumericLiteral / BigIntLiteral /
// NoSubstitutionTemplateLiteral / null / true / false / RegExp wrapped in
// a ComputedPropertyName) plus normalizes numeric-literal cooked values.
func getPropertyName(be *ast.BindingElement) (string, bool) {
	var key *ast.Node
	if be.PropertyName != nil {
		key = be.PropertyName
	} else {
		key = be.Name()
	}
	if key == nil {
		return "", false
	}
	return utils.GetStaticPropertyName(key)
}

// typeLabel reports `'parameter'` for Parameter nodes and `'property'` for
// BindingElement nodes. Mirrors upstream's
// `node.parent.type === Property || ArrayPattern ? 'property' : 'parameter'`
// without re-checking the parent — the listener routing has already
// established the kind.
func typeLabel(node *ast.Node) string {
	if node.Kind == ast.KindBindingElement {
		return "property"
	}
	return "parameter"
}

// isAnonymousFunctionLike mirrors upstream's
// `parent === ArrowFunctionExpression || FunctionExpression` check, expanded
// to cover tsgo's separate kinds for methods, constructors, and accessors
// (in ESTree these wrap a FunctionExpression; in tsgo they have dedicated
// kinds). FunctionDeclaration is intentionally excluded — upstream returns
// from `checkAssignmentPattern` without entering the contextual-type arm
// for top-level function declarations.
func isAnonymousFunctionLike(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindMethodDeclaration,
		ast.KindConstructor,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
		return true
	}
	return false
}

// indexOfParameter returns the 0-based position of `paramNode` within the
// enclosing function-like declaration's parameter list. Returns -1 if not
// found.
func indexOfParameter(funcNode *ast.Node, paramNode *ast.Node) int {
	params := funcNode.Parameters()
	for i, p := range params {
		if p == paramNode {
			return i
		}
	}
	return -1
}

// indexOfBindingElement returns the 0-based index of `child` inside its
// containing ObjectBindingPattern or ArrayBindingPattern. Returns -1 if not
// found. tsgo collapses both pattern kinds onto a single "Elements" accessor
// of `BindingPattern`, so we walk that list.
func indexOfBindingElement(pattern *ast.Node, child *ast.Node) int {
	bp := pattern.AsBindingPattern()
	if bp == nil {
		return -1
	}
	for i, el := range bp.Elements.Nodes {
		if el == child {
			return i
		}
	}
	return -1
}

// reportUselessDefaultAssignment emits the `uselessDefaultAssignment`
// diagnostic with the autofix that removes the ` = <initializer>` text.
func reportUselessDefaultAssignment(ctx rule.RuleContext, node *ast.Node, initializer *ast.Node, kind string) {
	fix := removeDefaultFix(node, initializer)
	ctx.ReportNodeWithFixes(initializer, buildUselessDefaultAssignmentMessage(kind), fix)
}

func reportUselessUndefined(ctx rule.RuleContext, node *ast.Node, initializer *ast.Node, kind string) {
	fix := removeDefaultFix(node, initializer)
	ctx.ReportNodeWithFixes(initializer, buildUselessUndefinedMessage(kind), fix)
}

// reportPreferOptionalSyntax combines removeDefaultFix with the `?` insertion
// after the parameter's binding identifier. Upstream only inserts when the
// binding is an Identifier — array/object patterns can't carry the `?`
// syntactically, so we mirror the gate.
func reportPreferOptionalSyntax(ctx rule.RuleContext, node *ast.Node, initializer *ast.Node) {
	fixes := []rule.RuleFix{removeDefaultFix(node, initializer)}
	param := node.AsParameterDeclaration()
	name := param.Name()
	if name != nil && name.Kind == ast.KindIdentifier {
		insertPos := name.End()
		fixes = append(fixes, rule.RuleFix{
			Text:  "?",
			Range: core.NewTextRange(insertPos, insertPos),
		})
	}
	ctx.ReportNodeWithFixes(initializer, buildPreferOptionalSyntaxMessage(), fixes...)
}

// removeDefaultFix produces the `[leftEnd, initializer.End()]` removal range
// — the same span upstream's `removeDefault(fixer, node)` computes via
// `[node.left.range[1], node.range[1]]`. In ESTree the "left" range includes
// the type annotation (TS attaches `typeAnnotation` to the Identifier); in
// tsgo the type sits on the Parameter itself, so we compute leftEnd as the
// last of Name.End(), QuestionToken.End(), and Type.End() for parameters.
func removeDefaultFix(node *ast.Node, initializer *ast.Node) rule.RuleFix {
	leftEnd := initializer.Pos()
	switch node.Kind {
	case ast.KindParameter:
		param := node.AsParameterDeclaration()
		if name := param.Name(); name != nil {
			leftEnd = name.End()
		}
		if param.QuestionToken != nil && param.QuestionToken.End() > leftEnd {
			leftEnd = param.QuestionToken.End()
		}
		if param.Type != nil && param.Type.End() > leftEnd {
			leftEnd = param.Type.End()
		}
	case ast.KindBindingElement:
		if name := node.AsBindingElement().Name(); name != nil {
			leftEnd = name.End()
		}
	}
	return rule.RuleFix{
		Range: core.NewTextRange(leftEnd, initializer.End()),
		Text:  "",
	}
}
