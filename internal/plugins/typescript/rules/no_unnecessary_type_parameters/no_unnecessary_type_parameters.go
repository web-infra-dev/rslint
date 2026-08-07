package no_unnecessary_type_parameters

import (
	"math"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var replaceUsagesWithConstraintMessage = rule.RuleMessage{
	Id:          "replaceUsagesWithConstraint",
	Description: "Replace all usages of type parameter with its constraint.",
}

func buildSoleMessage(name, descriptor, uses string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "sole",
		Description: "Type parameter " + name + " is " + uses + " in the " + descriptor + " signature.",
		Data: map[string]string{
			"name":       name,
			"descriptor": descriptor,
			"uses":       uses,
		},
	}
}

var NoUnnecessaryTypeParametersRule = rule.CreateRule(rule.Rule{
	Name:             "no-unnecessary-type-parameters",
	RequiresTypeInfo: true,
	Schema:           rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkFunctionNode := func(node *ast.Node) { checkNode(ctx, node, "function") }
		checkClassNode := func(node *ast.Node) { checkNode(ctx, node, "class") }

		return rule.RuleListeners{
			// Mirrors upstream's selector groups. TSConstructSignatureDeclaration
			// (interface `new <T>(): T`) is deliberately omitted: upstream's own
			// selector list excludes it too, so it is never checked.
			ast.KindArrowFunction:       checkFunctionNode,
			ast.KindFunctionDeclaration: checkFunctionNode,
			ast.KindFunctionExpression:  checkFunctionNode,
			// tsgo represents a method's value directly on the MethodDeclaration
			// node (no separate FunctionExpression/TSEmptyBodyFunctionExpression
			// wrapper the way ESTree does), so this single kind covers both.
			ast.KindMethodDeclaration: checkFunctionNode,
			ast.KindCallSignature:     checkFunctionNode,
			ast.KindConstructorType:   checkFunctionNode,
			ast.KindFunctionType:      checkFunctionNode,
			ast.KindMethodSignature:   checkFunctionNode,
			ast.KindClassDeclaration:  checkClassNode,
			ast.KindClassExpression:   checkClassNode,
		}
	},
})

func checkNode(ctx rule.RuleContext, node *ast.Node, descriptor string) {
	typeParams := node.TypeParameters()
	if len(typeParams) == 0 {
		return
	}

	// Lazily resolved: only pay for the type-checker walk when at least one
	// type parameter isn't already proven repeated by the cheap AST check.
	var counts map[*ast.Node]int

	for _, typeParamNode := range typeParams {
		typeParam := typeParamNode.AsTypeParameterDeclaration()
		if typeParam == nil {
			continue
		}
		nameNode := typeParam.Name()
		if nameNode == nil || !ast.IsIdentifier(nameNode) {
			continue
		}
		sym := typeParamNode.Symbol()
		if sym == nil {
			continue
		}

		if isTypeParameterRepeatedInAST(ctx, typeParamNode, node, sym) {
			continue
		}

		if counts == nil {
			counts = countTypeParameterUsage(ctx.TypeChecker, node)
		}
		useCount := counts[nameNode]
		if useCount == 0 || useCount > 2 {
			continue
		}

		uses := "used only once"
		if useCount == 1 {
			uses = "never used"
		}

		msg := buildSoleMessage(nameNode.AsIdentifier().Text, descriptor, uses)
		container := node
		ctx.ReportNodeWithDeferredSuggestions(typeParamNode, msg, func() []rule.RuleSuggestion {
			return []rule.RuleSuggestion{{
				Message:  replaceUsagesWithConstraintMessage,
				FixesArr: buildReplaceWithConstraintFixes(ctx, container, typeParamNode, typeParam, sym),
			}}
		})
	}
}

// isTypeParameterRepeatedInAST is the cheap syntactic shortcut: if the type
// parameter is textually referenced at least twice within its own declaring
// signature (excluding its own declaration and anything past the function
// body / end of signature), it's provably not "sole" and the type-checker
// walk can be skipped entirely.
func isTypeParameterRepeatedInAST(ctx rule.RuleContext, typeParamNode *ast.Node, container *ast.Node, sym *ast.Symbol) bool {
	cutoff := declaringSignatureEnd(container)
	total := 0
	for _, ref := range ctx.Refs.References(sym) {
		// References inside the type parameter's own declaration (e.g. a
		// self-referential constraint like `T extends Foo<T>`) don't count.
		if ref.Pos() < typeParamNode.End() && ref.End() > typeParamNode.Pos() {
			continue
		}
		// References outside the declaring signature (i.e. only reachable
		// from the function body) don't count either.
		if ref.Pos() > cutoff {
			continue
		}
		if isDirectGenericTypeArgumentUsage(ref) {
			return true
		}
		total++
		if total >= 2 {
			return true
		}
	}
	return false
}

// declaringSignatureEnd returns the position past which a reference no
// longer belongs to node's own declaring signature: the start of its body
// (function-like with an implementation, or a class's member list), the end
// of an explicit return-type annotation when there's no body, or unbounded
// when neither exists.
func declaringSignatureEnd(node *ast.Node) int {
	if body := node.Body(); body != nil {
		return body.Pos()
	}
	if ast.IsClassLike(node) {
		if members := node.MemberList(); members != nil {
			return members.Pos()
		}
	}
	if returnType := node.Type(); returnType != nil {
		return returnType.End()
	}
	return math.MaxInt
}

// isDirectGenericTypeArgumentUsage reports whether identNode — a bare
// reference to the type parameter used as a type (`T`) — sits directly as a
// type argument of an outer generic type reference, e.g. the T in `Map<T,
// V>` or `Foo<T>`. Such a usage is treated as an automatic "used multiple
// times", mirroring upstream, unless the outer reference is `Array<T>` /
// `ReadonlyArray<T>`, which defer to the more precise type-checker phase.
func isDirectGenericTypeArgumentUsage(identNode *ast.Node) bool {
	immediateParent := identNode.Parent
	if immediateParent == nil || immediateParent.Kind != ast.KindTypeReference {
		return false
	}
	// ESLint's AST has no parenthesized-type node, so in `Foo<(T)>` the bare
	// reference is what its type-argument list holds. Peel tsgo's wrappers to
	// reach the node that actually sits in the list.
	argument := immediateParent
	for argument.Parent != nil && argument.Parent.Kind == ast.KindParenthesizedType {
		argument = argument.Parent
	}
	grandparent := skipUnionIntersectionUpward(argument.Parent)
	if grandparent == nil || !canHaveTypeArgumentsList(grandparent.Kind) {
		return false
	}
	if !slices.Contains(grandparent.TypeArguments(), argument) {
		return false
	}
	// The Array<T>/ReadonlyArray<T> exclusion only makes sense for an actual
	// type reference (e.g. `Foo<T>`), not for a type argument attached to a
	// call/new/tagged-template/JSX expression (e.g. `foo<T>()`).
	if grandparent.Kind == ast.KindTypeReference {
		outerName := grandparent.AsTypeReferenceNode().TypeName
		if outerName != nil && outerName.Kind == ast.KindIdentifier {
			switch outerName.AsIdentifier().Text {
			case "Array", "ReadonlyArray":
				return false
			}
		}
	}
	return true
}

// canHaveTypeArgumentsList reports whether kind is a node that can directly
// own a `<...>` type-argument list — matching every construct ESLint
// represents via a TSTypeParameterInstantiation wrapper: generic type
// references, and generic call/new/tagged-template/JSX expressions.
func canHaveTypeArgumentsList(kind ast.Kind) bool {
	switch kind {
	case ast.KindTypeReference, ast.KindExpressionWithTypeArguments,
		ast.KindImportType, ast.KindTypeQuery,
		ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression,
		ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement:
		return true
	}
	return false
}

func skipUnionIntersectionUpward(node *ast.Node) *ast.Node {
	for node != nil && (node.Kind == ast.KindUnionType || node.Kind == ast.KindIntersectionType) {
		node = node.Parent
	}
	return node
}

// countTypeParameterUsage resolves, for every type parameter reachable from
// node, how many times the type checker sees it used across the signature
// (including inferred return types). For a class, every member contributes
// (with generic type-argument usages always treated as "multiple"); for a
// function-like node, only its own signature contributes.
func countTypeParameterUsage(tc *checker.Checker, node *ast.Node) map[*ast.Node]int {
	counts := make(map[*ast.Node]int)

	if ast.IsClassLike(node) {
		for _, typeParamNode := range node.TypeParameters() {
			collectTypeParameterUsageCounts(tc, typeParamNode, counts, true)
		}
		if members := node.MemberList(); members != nil {
			for _, member := range members.Nodes {
				collectTypeParameterUsageCounts(tc, member, counts, true)
			}
		}
	} else {
		collectTypeParameterUsageCounts(tc, node, counts, false)
	}

	return counts
}

// collectTypeParameterUsageCounts walks the resolved type graph rooted at
// node, incrementing foundIdentifierUsages for every type-parameter
// declaration identifier it encounters. Each type parameter always
// contributes a baseline "self" count of 1 (from being enumerated as one of
// its declaring signature's own type parameters, see the
// `signature.TypeParameters()` walk in visitSignature and the class
// self-visit loop in countTypeParameterUsage); every further occurrence adds
// 1 (or 2, when assumeMultipleUses applies) on top of that baseline. A final
// count of 1 or 2 is "sole"; 3 or more means it's genuinely used more than
// once.
//
// fromClass marks members visited on behalf of a class: type-reference type
// arguments always count as multiple use there, since a class's type
// parameters are shared across every member.
func collectTypeParameterUsageCounts(tc *checker.Checker, node *ast.Node, foundIdentifierUsages map[*ast.Node]int, fromClass bool) {
	visitedProperties := make(map[*checker.Type]bool)
	visitedConstraints := make(map[*ast.Node]bool)
	visitedDefault := false
	typeUsages := make(map[*checker.Type]int)
	functionLikeType := false

	incrementIdentifierCount := func(id *ast.Node, assumeMultipleUses bool) {
		value := 1
		if assumeMultipleUses {
			value = 2
		}
		foundIdentifierUsages[id] += value
	}

	// incrementTypeUsages guards against unbounded recursion on recursive
	// generic types (e.g. `type T = { [P in keyof T]: T }`): once a type has
	// been seen more than 9 times, every referenced type parameter has
	// already been counted enough to qualify as used, so visiting stops.
	incrementTypeUsages := func(t *checker.Type) int {
		typeUsages[t]++
		return typeUsages[t]
	}

	var visitType func(t *checker.Type, assumeMultipleUses bool, isReturnType bool)

	visitSymbolsListOnce := func(owner *checker.Type, symbols []*ast.Symbol, assumeMultipleUses bool) {
		if visitedProperties[owner] {
			return
		}
		visitedProperties[owner] = true
		for _, sym := range symbols {
			visitType(tc.GetTypeOfSymbol(sym), assumeMultipleUses, false)
		}
	}

	visitTypesList := func(types []*checker.Type, assumeMultipleUses bool) {
		for _, t := range types {
			visitType(t, assumeMultipleUses, false)
		}
	}

	visitSignature := func(sig *checker.Signature) {
		if sig == nil {
			return
		}
		if this := sig.ThisParameter(); this != nil {
			visitType(tc.GetTypeOfSymbol(this), false, false)
		}
		for _, param := range sig.Parameters() {
			visitType(tc.GetTypeOfSymbol(param), false, false)
		}
		for _, typeParam := range sig.TypeParameters() {
			visitType(typeParam, false, false)
		}
		returnType := tc.GetReturnTypeOfSignature(sig)
		if predicate := tc.GetTypePredicateOfSignature(sig); predicate != nil && predicate.Type() != nil {
			returnType = predicate.Type()
		}
		visitType(returnType, false, true)
	}

	visitType = func(t *checker.Type, assumeMultipleUses bool, isReturnType bool) {
		if t == nil || incrementTypeUsages(t) > 9 {
			return
		}

		switch {
		case utils.IsTypeParameter(t):
			sym := t.Symbol()
			if sym == nil || len(sym.Declarations) == 0 {
				return
			}
			// A polymorphic `this` type also carries TypeFlagsTypeParameter, but
			// its symbol is the declaring class/interface, so its first
			// declaration is not a type parameter. Such a type contributes no
			// count and has no constraint or default to recurse into.
			decl := sym.Declarations[0]
			if !ast.IsTypeParameterDeclaration(decl) {
				return
			}
			declTypeParam := decl.AsTypeParameterDeclaration()
			nameNode := declTypeParam.Name()
			if nameNode != nil {
				incrementIdentifierCount(nameNode, assumeMultipleUses)
			}
			// Resolve the constraint/default through the checker on t itself
			// (not by re-resolving the raw declaration node) so a mapped
			// type's own bound type parameter (`K` in `{ [K in T]: ... }`)
			// picks up whatever substitution applies to this specific
			// instantiation, e.g. when reached through an inferred return
			// type of a call to another generic function.
			if declTypeParam.Constraint != nil && !visitedConstraints[declTypeParam.Constraint] {
				visitedConstraints[declTypeParam.Constraint] = true
				visitType(tc.GetConstraintOfTypeParameter(t), false, false)
			}
			if declTypeParam.DefaultType != nil && !visitedDefault {
				visitedDefault = true
				visitType(tc.GetDefaultFromTypeParameter(t), false, false)
			}

		case t.Alias() != nil && t.Alias().TypeArguments() != nil:
			// A generic type-alias reference like `Exclude<T, null>`. We don't
			// descend into the alias's own definition, so it's safest to assume
			// every argument is used multiple times.
			visitTypesList(t.Alias().TypeArguments(), true)

		case utils.IsUnionType(t) || utils.IsIntersectionType(t):
			visitTypesList(t.Types(), assumeMultipleUses)

		case utils.IsTypeFlagSet(t, checker.TypeFlagsIndexedAccess):
			iat := t.AsIndexedAccessType()
			visitType(iat.ObjectType(), assumeMultipleUses, false)
			visitType(iat.IndexType(), assumeMultipleUses, false)

		case utils.IsTypeReference(t):
			target := t.Target()
			for _, typeArgument := range checker.Checker_getTypeArguments(tc, t) {
				thisAssumeMultipleUses := fromClass || assumeMultipleUses
				if !thisAssumeMultipleUses {
					switch {
					case checker.IsTupleType(target):
						thisAssumeMultipleUses = isReturnType && !target.AsTupleType().IsReadonly()
					case tc.IsArrayType(target):
						sym := t.Symbol()
						thisAssumeMultipleUses = isReturnType && sym != nil && sym.Name == "Array"
					default:
						thisAssumeMultipleUses = true
					}
				}
				visitType(typeArgument, thisAssumeMultipleUses, isReturnType)
			}

		case utils.IsTypeFlagSet(t, checker.TypeFlagsTemplateLiteral):
			for _, subType := range t.AsTemplateLiteralType().Types() {
				visitType(subType, assumeMultipleUses, false)
			}

		case utils.IsTypeFlagSet(t, checker.TypeFlagsConditional):
			ct := t.AsConditionalType()
			visitType(ct.CheckType(), assumeMultipleUses, false)
			visitType(ct.ExtendsType(), assumeMultipleUses, false)

		case utils.IsObjectType(t):
			properties := checker.Checker_getPropertiesOfType(tc, t)
			visitSymbolsListOnce(t, properties, false)

			if checker.Type_objectFlags(t)&checker.ObjectFlagsMapped != 0 {
				visitType(checker.Checker_getTypeParameterFromMappedType(tc, t), false, false)
				if len(properties) == 0 {
					template := checker.Checker_getTemplateTypeFromMappedType(tc, t)
					if template == nil {
						template = checker.Checker_getConstraintTypeFromMappedType(tc, t)
					}
					visitType(template, false, false)
				}
				// A key remapped through `as` is a signature position like any
				// other. A mapped type's name type is internal to the checker,
				// so upstream — which only has the public ts.Type surface —
				// never descends here and reads a type parameter spelled only
				// in an `as` clause as unused.
				visitType(checker.Checker_getNameTypeFromMappedType(tc, t), false, false)
			}

			// Every index signature instantiates its value type once per key,
			// so all of them count as multiple uses. The public ts.Type
			// surface only offers getStringIndexType/getNumberIndexType, so
			// upstream sees no use at all in a symbol-keyed or pattern-keyed
			// signature.
			for _, info := range checker.Checker_getIndexInfosOfType(tc, t) {
				visitType(info.ValueType(), true, false)
			}

			for _, sig := range utils.GetCallSignatures(tc, t) {
				functionLikeType = true
				visitSignature(sig)
			}
			for _, sig := range utils.GetConstructSignatures(tc, t) {
				functionLikeType = true
				visitSignature(sig)
			}

		case utils.IsTypeFlagSet(t, checker.TypeFlagsIndex):
			// `keyof T`.
			visitType(t.AsIndexType().Target(), assumeMultipleUses, false)

		case utils.IsTypeFlagSet(t, checker.TypeFlagsStringMapping):
			// `Uppercase<T>` / `Lowercase<T>` / `Capitalize<T>` / `Uncapitalize<T>`.
			visitType(t.AsStringMappingType().Target(), assumeMultipleUses, false)
		}
	}

	if node.Kind == ast.KindCallSignature || node.Kind == ast.KindConstructor {
		functionLikeType = true
		visitSignature(tc.GetSignatureFromDeclaration(node))
	}
	if !functionLikeType {
		visitType(tc.GetTypeAtLocation(node), false, false)
	}
}

// buildReplaceWithConstraintFixes builds the suggestion that replaces every
// use of the type parameter with its constraint (or "unknown" when it has
// none) and removes the type parameter from the declaration. Unlike
// detection, which only looks at the declaring signature, this replaces
// every reference reachable from the type parameter's declaration,
// including uses inside a function body.
func buildReplaceWithConstraintFixes(ctx rule.RuleContext, container *ast.Node, typeParamNode *ast.Node, typeParam *ast.TypeParameterDeclaration, sym *ast.Symbol) []rule.RuleFix {
	sourceFile := ctx.SourceFile

	constraintText := "unknown"
	var unwrappedConstraint *ast.Node
	if typeParam.Constraint != nil {
		unwrappedConstraint = ast.SkipTypeParentheses(typeParam.Constraint)
		if unwrappedConstraint.Kind != ast.KindAnyKeyword {
			constraintText = utils.TrimmedNodeText(sourceFile, unwrappedConstraint)
		}
	}
	fixes := make([]rule.RuleFix, 0, len(ctx.Refs.References(sym))+1)
	for _, refNode := range ctx.Refs.References(sym) {
		text := constraintText
		if constraintNeedsParentheses(refNode, unwrappedConstraint) {
			text = "(" + constraintText + ")"
		}
		fixes = append(fixes, rule.RuleFixReplace(sourceFile, refNode, text))
	}

	fixes = append(fixes, removeTypeParameterFix(sourceFile, container, typeParamNode))
	return fixes
}

// Binding levels of the TypeScript type grammar, loosest first. A constraint
// written at a level looser than its substitution site accepts has to be
// parenthesized or the surrounding operators re-associate: replacing T in
// `T[]` with `readonly string[]` must yield `(readonly string[])[]`, since
// `readonly string[][]` is an array of `string[]` instead.
const (
	precLoose        = iota // conditional, function and constructor types
	precUnion               // `A | B`
	precIntersection        // `A & B`
	precTypeOperator        // `keyof A`, `readonly A`, `unique symbol`
	precPostfix             // `A[]`, `A[B]`, and every primary type
)

func typePrecedence(node *ast.Node) int {
	switch node.Kind {
	case ast.KindConditionalType, ast.KindFunctionType, ast.KindConstructorType:
		return precLoose
	case ast.KindUnionType:
		return precUnion
	case ast.KindIntersectionType:
		return precIntersection
	case ast.KindTypeOperator:
		return precTypeOperator
	}
	return precPostfix
}

// constraintNeedsParentheses reports whether constraint, written in place of
// a reference to the type parameter, has to be wrapped to keep its meaning.
func constraintNeedsParentheses(refNode *ast.Node, constraint *ast.Node) bool {
	if constraint == nil {
		return false
	}
	reference := refNode.Parent
	if reference == nil {
		return false
	}
	site := reference.Parent
	if site == nil || site.Kind == ast.KindParenthesizedType {
		// The position already carries parentheses of its own.
		return false
	}
	if typePrecedence(constraint) < requiredPrecedence(site, reference) {
		return true
	}
	// Upstream wraps a union, intersection or conditional constraint in these
	// positions whether or not the grammar asks for it. Keep those parentheses
	// so the suggestion reads the same as ESLint's.
	switch constraint.Kind {
	case ast.KindUnionType, ast.KindIntersectionType, ast.KindConditionalType:
		switch site.Kind {
		case ast.KindArrayType, ast.KindIndexedAccessType, ast.KindIntersectionType, ast.KindUnionType:
			return true
		}
	}
	return false
}

// requiredPrecedence returns the tightest binding a type must have to sit at
// child's position within site without parentheses.
func requiredPrecedence(site *ast.Node, child *ast.Node) int {
	switch site.Kind {
	case ast.KindArrayType:
		return precPostfix
	case ast.KindIndexedAccessType:
		// Only the object half is a postfix operand; the index sits inside
		// brackets, which accept any type.
		if site.AsIndexedAccessTypeNode().ObjectType == child {
			return precPostfix
		}
	case ast.KindTypeOperator, ast.KindIntersectionType:
		return precTypeOperator
	case ast.KindUnionType:
		return precIntersection
	case ast.KindConditionalType:
		// The check and extends halves are parsed without conditional and
		// function types; the branches take a full type.
		conditional := site.AsConditionalTypeNode()
		if conditional.CheckType == child || conditional.ExtendsType == child {
			return precUnion
		}
	}
	return precLoose
}

// removeTypeParameterFix removes typeParamNode from container's `<...>`
// clause: the whole clause when it's the only type parameter, otherwise
// typeParamNode plus its separating comma.
func removeTypeParameterFix(sourceFile *ast.SourceFile, container *ast.Node, typeParamNode *ast.Node) rule.RuleFix {
	typeParams := container.TypeParameters()
	paramRange := utils.TrimNodeTextRange(sourceFile, typeParamNode)

	if len(typeParams) == 1 {
		list := container.TypeParameterList()
		start, end, ok := utils.RangeEnclosingDelimiters(sourceFile.Text(), list.Pos(), list.End(), '<', '>')
		if !ok {
			return rule.RuleFixRemoveRange(paramRange)
		}
		return rule.RuleFixRemoveRange(core.NewTextRange(start, end))
	}

	index := slices.Index(typeParams, typeParamNode)
	if index == 0 {
		commaEnd := findTokenEnd(sourceFile, paramRange.End(), ast.KindCommaToken)
		nextParamStart := utils.TrimNodeTextRange(sourceFile, typeParams[1]).Pos()
		nextStart := utils.SkipLeadingWhitespace(sourceFile.Text(), commaEnd, nextParamStart)
		return rule.RuleFixRemoveRange(core.NewTextRange(paramRange.Pos(), nextStart))
	}

	commaStart := findTokenStart(sourceFile, typeParams[index-1].End(), ast.KindCommaToken)
	return rule.RuleFixRemoveRange(core.NewTextRange(commaStart, paramRange.End()))
}

func findTokenStart(sourceFile *ast.SourceFile, from int, kind ast.Kind) int {
	s := scanner.GetScannerForSourceFile(sourceFile, from)
	for s.Token() != kind && s.Token() != ast.KindEndOfFile {
		s.Scan()
	}
	return s.TokenStart()
}

func findTokenEnd(sourceFile *ast.SourceFile, from int, kind ast.Kind) int {
	s := scanner.GetScannerForSourceFile(sourceFile, from)
	for s.Token() != kind && s.Token() != ast.KindEndOfFile {
		s.Scan()
	}
	return s.TokenEnd()
}
