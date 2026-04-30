package utils

import (
	"fmt"
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
)

type ConstraintTypeInfo struct {
	ConstraintType  *checker.Type
	IsTypeParameter bool
}

/**
 * Returns whether the type is a generic and what its constraint is.
 *
 * If the type is not a generic, `isTypeParameter` will be `false`, and
 * `constraintType` will be the same as the input type.
 *
 * If the type is a generic, and it is constrained, `isTypeParameter` will be
 * `true`, and `constraintType` will be the constraint type.
 *
 * If the type is a generic, but it is not constrained, `constraintType` will be
 * `undefined` (rather than an `unknown` type), due to https://github.com/microsoft/TypeScript/issues/60475
 *
 * Successor to {@link getConstrainedTypeAtLocation} due to https://github.com/typescript-eslint/typescript-eslint/issues/10438
 *
 * This is considered internal since it is unstable for now and may have breaking changes at any time.
 * Use at your own risk.
 *
 * @internal
 *
 */
func GetConstraintInfo(
	typeChecker *checker.Checker,
	t *checker.Type,
) (constraintType *checker.Type, isTypeParameter bool) {
	if checker.Type_flags(t)&checker.TypeFlagsTypeParameter != 0 {
		return checker.Checker_getBaseConstraintOfType(typeChecker, t), true
	}
	return t, false
}

type TypeAwaitable int32

const (
	TypeAwaitableAlways TypeAwaitable = iota
	TypeAwaitableNever
	TypeAwaitableMay
)

func NeedsToBeAwaited(
	typeChecker *checker.Checker,
	node *ast.Node,
	t *checker.Type,
) TypeAwaitable {
	constraintType, isTypeParameter := GetConstraintInfo(typeChecker, t)

	// unconstrained generic types should be treated as unknown
	if isTypeParameter && constraintType == nil {
		return TypeAwaitableMay
	}

	// `any` and `unknown` types may need to be awaited
	if IsTypeAnyType(constraintType) || IsTypeUnknownType(constraintType) {
		return TypeAwaitableMay
	}

	// 'thenable' values should always be be awaited
	if IsThenableType(typeChecker, node, constraintType) {
		return TypeAwaitableAlways
	}

	// anything else should not be awaited
	return TypeAwaitableNever
}

func GetConstrainedTypeAtLocation(typeChecker *checker.Checker, node *ast.Node) *checker.Type {
	nodeType := typeChecker.GetTypeAtLocation(node)

	constraint := checker.Checker_getBaseConstraintOfType(typeChecker, nodeType)
	if constraint != nil {
		return constraint
	}

	return nodeType
}

/**
 * Get the type name of a given type.
 * @param typeChecker The context sensitive TypeScript TypeChecker.
 * @param type The type to get the name of.
 */
func GetTypeName(
	typeChecker *checker.Checker,
	t *checker.Type,
) string {
	// It handles `string` and string literal types as string.
	if checker.Type_flags(t)&checker.TypeFlagsStringLike != 0 {
		return "string"
	}

	// If the type is a type parameter which extends primitive string types,
	// but it was not recognized as a string like. So check the constraint
	// type of the type parameter.
	if IsTypeParameter(t) {
		// `type.getConstraint()` method doesn't return the constraint type of
		// the type parameter for some reason. So this gets the constraint type
		// via AST.
		symbol := checker.Type_symbol(t)
		decls := symbol.Declarations
		if len(decls) > 0 {
			if ast.IsTypeParameterDeclaration(decls[0]) {
				typeParamDecl := decls[0].AsTypeParameterDeclaration()
				if typeParamDecl.Constraint != nil {
					return GetTypeName(typeChecker, checker.Checker_getTypeFromTypeNode(typeChecker, typeParamDecl.Constraint))
				}
			}
		}
	}

	// If the type is a union and all types in the union are string like,
	// return `string`. For example:
	// - `"a" | "b"` is string.
	// - `string | string[]` is not string.
	if IsUnionType(t) && Every(UnionTypeParts(t), func(t *checker.Type) bool {
		return GetTypeName(typeChecker, t) == "string"
	}) {
		return "string"
	}

	// If the type is an intersection and a type in the intersection is string
	// like, return `string`. For example: `string & {__htmlEscaped: void}`
	if IsIntersectionType(t) && Some(IntersectionTypeParts(t), func(t *checker.Type) bool {
		return GetTypeName(typeChecker, t) == "string"
	}) {
		return "string"
	}

	return typeChecker.TypeToString(t)
}

/**
 * Gets the location of the head of the given for statement variant for reporting.
 *
 * - `for (const foo in bar) expressionOrBlock`
 *    ^^^^^^^^^^^^^^^^^^^^^^
 *
 * - `for (const foo of bar) expressionOrBlock`
 *    ^^^^^^^^^^^^^^^^^^^^^^
 *
 * - `for await (const foo of bar) expressionOrBlock`
 *    ^^^^^^^^^^^^^^^^^^^^^^^^^^^^
 *
 * - `for (let i = 0; i < 10; i++) expressionOrBlock`
 *    ^^^^^^^^^^^^^^^^^^^^^^^^^^^^
 */
func GetForStatementHeadLoc(
	sourceFile *ast.SourceFile,
	node *ast.Node,
) core.TextRange {
	var statement *ast.Node
	if ast.IsForStatement(node) {
		statement = node.AsForStatement().Statement
	} else {
		statement = node.AsForInOrOfStatement().Statement
	}
	return TrimNodeTextRange(sourceFile, node).WithEnd(statement.Pos())
}

/**
 * Gets the location of a function node's "head" for reporting.
 * Matches the behavior of typescript-eslint's getFunctionHeadLoc:
 *
 * - `function foo() {}`         → `function foo`
 * - `(function() {})`           → `function`
 * - `() => {}`                  → `=>`
 * - `class A { method() {} }`   → `method`
 * - `class A { get foo() {} }`  → `get foo`
 * - `class A { static async foo() {} }` → `static async foo`
 * - `class A { foo = () => {} }` → `foo = `
 * - `{ foo: function() {} }`    → `foo: function`
 * - `export default function() {}` → `function`
 */
func GetFunctionHeadLoc(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	parent := node.Parent

	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		start := nodeStartSkippingDecorators(sourceFile, node)
		// Start scanning for the parameters `(` after any decorator factory
		// (e.g. `@dec()`) and after the method name. Nameless constructors
		// fall back to the first token after the decorators.
		searchFrom := start.Pos()
		if name := node.Name(); name != nil {
			searchFrom = name.End()
		}
		if parenPos := findOpenParenPosFrom(sourceFile, searchFrom, node.End()); parenPos >= 0 {
			return start.WithEnd(parenPos)
		}
		if node.Body() != nil {
			return start.WithEnd(node.Body().Pos())
		}
		return start

	case ast.KindArrowFunction:
		if parent != nil && (parent.Kind == ast.KindPropertyDeclaration || parent.Kind == ast.KindPropertyAssignment) {
			start := nodeStartSkippingDecorators(sourceFile, parent)
			if parenPos := findOpenParenPos(sourceFile, node); parenPos >= 0 {
				return start.WithEnd(parenPos)
			}
			af := node.AsArrowFunction()
			if af.Parameters != nil && len(af.Parameters.Nodes) > 0 {
				paramStart := scanner.GetRangeOfTokenAtPosition(sourceFile, af.Parameters.Nodes[0].Pos())
				return start.WithEnd(paramStart.Pos())
			}
			return start.WithEnd(af.EqualsGreaterThanToken.Pos())
		}
		af := node.AsArrowFunction()
		arrowRange := scanner.GetRangeOfTokenAtPosition(sourceFile, af.EqualsGreaterThanToken.Pos())
		return core.NewTextRange(arrowRange.Pos(), arrowRange.End())

	case ast.KindFunctionExpression:
		if parent != nil && (parent.Kind == ast.KindPropertyAssignment || parent.Kind == ast.KindPropertyDeclaration) {
			start := nodeStartSkippingDecorators(sourceFile, parent)
			if parenPos := findOpenParenPos(sourceFile, node); parenPos >= 0 {
				return start.WithEnd(parenPos)
			}
			if node.Body() != nil {
				return start.WithEnd(node.Body().Pos())
			}
			return start
		}
		start := TrimNodeTextRange(sourceFile, node)
		if parenPos := findOpenParenPos(sourceFile, node); parenPos >= 0 {
			return start.WithEnd(parenPos)
		}
		if node.Body() != nil {
			return start.WithEnd(node.Body().Pos())
		}
		return start

	case ast.KindFunctionDeclaration:
		start := findFunctionKeywordPos(sourceFile, node)
		if parenPos := findOpenParenPos(sourceFile, node); parenPos >= 0 {
			return core.NewTextRange(start, parenPos)
		}
		if node.Body() != nil {
			return core.NewTextRange(start, node.Body().Pos())
		}
		return TrimNodeTextRange(sourceFile, node)

	case ast.KindFunctionType:
		// Mirror ESLint's astUtils.getFunctionHeadLoc fallback for
		// TSFunctionType: range from the node's start to the opening '(' of
		// its parameters. With no type parameters this collapses to a
		// zero-width position at '('; with `<T>(...)` the range covers the
		// type-parameter list, exactly as upstream produces.
		trimmed := TrimNodeTextRange(sourceFile, node)
		if parenPos := findOpenParenPos(sourceFile, node); parenPos >= 0 {
			return trimmed.WithEnd(parenPos)
		}
		return trimmed
	}

	return TrimNodeTextRange(sourceFile, node)
}

// GetFunctionNameWithKind mirrors ESLint's astUtils.getFunctionNameWithKind.
// It produces a human-readable description of the function used in diagnostic
// messages (e.g., `"function 'foo'"`, `"static private method '#bar'"`,
// `"arrow function"`, `"constructor"`). Modifier order matches ESLint:
// static, private, async, generator, then the function-kind keyword.
//
// For nameless function-likes, walks the parent (VariableDeclaration,
// PropertyAssignment, PropertyDeclaration, PropertySignature,
// TypeAliasDeclaration) to recover the binding name where ESLint does the
// same via its `getName` / `getOuterName` helpers.
//
// For a FunctionExpression assigned as the value of an object literal
// property (`var obj = { foo: function () {} }`), classifies the node as a
// "method" — ESTree models that case via `Property.value === FunctionExpression`
// and ESLint's classifier emits "method"; tsgo only collapses method-shorthand
// (`{ foo() {} }`) into MethodDeclaration, so we recover the same description
// from the parent.
//
// Callers that need the upper-cased form (e.g., as the leading {{name}}
// placeholder of a sentence) should apply UpperCaseFirstASCII — this helper
// returns the lower-cased form to keep call sites simple.
func GetFunctionNameWithKind(node *ast.Node) string {
	if node.Kind == ast.KindConstructor {
		return "constructor"
	}

	flags := ast.GetFunctionFlags(node)
	isAsync := flags&ast.FunctionFlagsAsync != 0
	isGenerator := flags&ast.FunctionFlagsGenerator != 0

	parent := node.Parent
	isStatic, isPrivate := false, false
	// Direct class member (MethodDeclaration / GetAccessor / SetAccessor):
	// modifiers and private-key live on the function-like node itself.
	if parent != nil && (parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindClassExpression) {
		switch node.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			isStatic = ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic)
			if n := node.Name(); n != nil && n.Kind == ast.KindPrivateIdentifier {
				isPrivate = true
			}
		}
	}
	// Class field with arrow / function-expression initializer: modifiers and
	// private-key live on the surrounding PropertyDeclaration. Mirrors ESLint
	// v9's `parent.type === "PropertyDefinition" && parent.value === node`
	// branch in `astUtils.getFunctionNameWithKind`.
	if parent != nil && parent.Kind == ast.KindPropertyDeclaration {
		if grandparent := parent.Parent; grandparent != nil &&
			(grandparent.Kind == ast.KindClassDeclaration || grandparent.Kind == ast.KindClassExpression) {
			switch node.Kind {
			case ast.KindArrowFunction, ast.KindFunctionExpression:
				if ast.HasSyntacticModifier(parent, ast.ModifierFlagsStatic) {
					isStatic = true
				}
				if n := parent.Name(); n != nil && n.Kind == ast.KindPrivateIdentifier {
					isPrivate = true
				}
			}
		}
	}

	var tokens []string
	if isStatic {
		tokens = append(tokens, "static")
	}
	if isPrivate {
		tokens = append(tokens, "private")
	}
	if isAsync {
		tokens = append(tokens, "async")
	}
	if isGenerator {
		tokens = append(tokens, "generator")
	}

	switch node.Kind {
	case ast.KindGetAccessor:
		tokens = append(tokens, "getter")
	case ast.KindSetAccessor:
		tokens = append(tokens, "setter")
	case ast.KindMethodDeclaration:
		tokens = append(tokens, "method")
	case ast.KindArrowFunction:
		tokens = append(tokens, "arrow", "function")
	case ast.KindFunctionExpression:
		if parent != nil && parent.Kind == ast.KindPropertyAssignment {
			tokens = append(tokens, "method")
		} else {
			tokens = append(tokens, "function")
		}
	default:
		tokens = append(tokens, "function")
	}

	if name := getFunctionDisplayName(node); name != "" {
		tokens = append(tokens, fmt.Sprintf("'%s'", name))
	}

	return strings.Join(tokens, " ")
}

// getFunctionDisplayName resolves the user-visible name of a function-like
// node — first by inspecting the node's own name, then by walking the parent
// for variable / property / type binding sites that ESLint's `getName` covers.
func getFunctionDisplayName(node *ast.Node) string {
	if n := node.Name(); n != nil {
		switch n.Kind {
		case ast.KindPrivateIdentifier:
			return n.AsPrivateIdentifier().Text
		case ast.KindIdentifier:
			return n.AsIdentifier().Text
		}
		if s, ok := GetStaticPropertyName(n); ok {
			return s
		}
	}
	parent := node.Parent
	if parent == nil {
		return ""
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		if n := parent.Name(); n != nil && n.Kind == ast.KindIdentifier {
			return n.AsIdentifier().Text
		}
	case ast.KindPropertyAssignment, ast.KindPropertyDeclaration, ast.KindPropertySignature:
		if n := parent.Name(); n != nil {
			if n.Kind == ast.KindIdentifier {
				return n.AsIdentifier().Text
			}
			if n.Kind == ast.KindPrivateIdentifier {
				// PrivateIdentifier.Text already includes the leading '#',
				// matching ESLint's `getName` for PropertyDefinition with a
				// PrivateIdentifier key.
				return n.AsPrivateIdentifier().Text
			}
			if s, ok := GetStaticPropertyName(n); ok {
				return s
			}
		}
	case ast.KindTypeAliasDeclaration:
		if n := parent.Name(); n != nil && n.Kind == ast.KindIdentifier {
			return n.AsIdentifier().Text
		}
	}
	return ""
}

// UpperCaseFirstASCII returns s with its first byte mapped to upper case if
// the byte is an ASCII lowercase letter; otherwise returns s unchanged.
// Sufficient for ESLint's `astUtils.upperCaseFirst` since all function-kind
// tokens are ASCII English ("function", "method", "constructor", …).
func UpperCaseFirstASCII(s string) string {
	if s == "" {
		return s
	}
	r := s[0]
	if r >= 'a' && r <= 'z' {
		return string(r-('a'-'A')) + s[1:]
	}
	return s
}

// nodeStartSkippingDecorators returns a TextRange whose start is the first
// non-decorator token of the node. This matches ESLint's
// getFunctionHeadLoc, which excludes leading decorators on MethodDefinition
// and PropertyDefinition from the reported function head range.
func nodeStartSkippingDecorators(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	fallback := TrimNodeTextRange(sourceFile, node)
	mods := node.Modifiers()
	if mods == nil || len(mods.Nodes) == 0 {
		return fallback
	}
	var lastDecoratorEnd int
	for _, mod := range mods.Nodes {
		if mod.Kind == ast.KindDecorator && mod.End() > lastDecoratorEnd {
			lastDecoratorEnd = mod.End()
		}
	}
	if lastDecoratorEnd == 0 {
		return fallback
	}
	tokenAfter := scanner.GetRangeOfTokenAtPosition(sourceFile, lastDecoratorEnd)
	return core.NewTextRange(tokenAfter.Pos(), fallback.End())
}

// findOpenParenPos finds the position of the first '(' token in a function node.
func findOpenParenPos(sourceFile *ast.SourceFile, node *ast.Node) int {
	return findOpenParenPosFrom(sourceFile, node.Pos(), node.End())
}

// findOpenParenPosFrom scans for the first '(' token within [start, end).
func findOpenParenPosFrom(sourceFile *ast.SourceFile, start int, end int) int {
	s := scanner.GetScannerForSourceFile(sourceFile, start)
	for s.TokenStart() < end {
		if s.Token() == ast.KindOpenParenToken {
			return s.TokenStart()
		}
		s.Scan()
	}
	return -1
}

// findFunctionKeywordPos returns the start position of the function head,
// skipping only `export` and `default` keywords. Other modifiers like `async`
// and `declare` are kept because they are part of the function signature
// (matching ESLint's behavior where FunctionDeclaration.loc excludes export/default).
func findFunctionKeywordPos(sourceFile *ast.SourceFile, node *ast.Node) int {
	s := scanner.GetScannerForSourceFile(sourceFile, node.Pos())
	end := node.End()
	for s.TokenStart() < end {
		tok := s.Token()
		if tok == ast.KindExportKeyword || tok == ast.KindDefaultKeyword {
			s.Scan()
			continue
		}
		return s.TokenStart()
	}
	return TrimNodeTextRange(sourceFile, node).Pos()
}

var arrayPredicateFunctions = []string{"every", "filter", "find", "findIndex", "findLast", "findLastIndex", "some"}

func IsArrayMethodCallWithPredicate(
	typeChecker *checker.Checker,
	node *ast.CallExpression,
) bool {
	if !ast.IsAccessExpression(node.Expression) {
		return false
	}

	propertyName, ok := checker.Checker_getAccessedPropertyName(typeChecker, node.Expression)
	if !ok || !slices.Contains(arrayPredicateFunctions, propertyName) {
		return false
	}

	t := GetConstrainedTypeAtLocation(typeChecker, node.Expression.Expression())
	return TypeRecurser(t, func(t *checker.Type) bool {
		return checker.Checker_isArrayOrTupleType(typeChecker, t)
	})
}

func IsRestParameterDeclaration(decl *ast.Declaration) bool {
	return ast.IsParameterDeclaration(decl) && decl.AsParameterDeclaration().DotDotDotToken != nil
}

// GetDeclaration returns the first declaration of the symbol at `node`.
//
// Returns nil when `typeChecker` or `node` is nil. Rules with optional
// type info (those that do not set `RequiresTypeInfo: true`) are scheduled
// with a nil TypeChecker on "gap files" — files in the program but not in
// `typeInfoFiles` (see internal/linter/linter.go). Rather than requiring
// every caller to nil-guard manually, this helper degrades gracefully:
// no checker → no declaration → caller falls back to structural checks.
// The `node == nil` guard mirrors the same convention already used by
// `GetReferenceSymbol` in shadowing.go.
func GetDeclaration(
	typeChecker *checker.Checker,
	node *ast.Node,
) *ast.Declaration {
	if typeChecker == nil || node == nil {
		return nil
	}
	symbol := typeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return nil
	}
	if len(symbol.Declarations) > 0 {
		return symbol.Declarations[0]
	}
	return nil
}

/**
 * @returns true if the type is `any[]`
 */
func IsTypeAnyArrayType(
	t *checker.Type,
	typeChecker *checker.Checker,
) bool {
	return checker.Checker_isArrayType(typeChecker, t) &&
		IsTypeAnyType(checker.Checker_getTypeArguments(typeChecker, t)[0])
}

/**
 * @returns true if the type is `unknown[]`
 */
func IsTypeUnknownArrayType(
	t *checker.Type,
	typeChecker *checker.Checker,
) bool {
	return checker.Checker_isArrayType(typeChecker, t) &&
		IsTypeUnknownType(checker.Checker_getTypeArguments(typeChecker, t)[0])
}

/**
 * Does a simple check to see if there is an any being assigned to a non-any type.
 *
 * This also checks generic positions to ensure there's no unsafe sub-assignments.
 * Note: in the case of generic positions, it makes the assumption that the two types are the same.
 *
 * @example See tests for examples
 *
 * @returns false if it's safe, or an object with the two types if it's unsafe
 */
func IsUnsafeAssignment(
	t *checker.Type,
	receiverT *checker.Type,
	typeChecker *checker.Checker,
	senderNode *ast.Node,
) (receiver *checker.Type, sender *checker.Type, unsafe bool) {
	return isUnsafeAssignmentWorker(
		t,
		receiverT,
		typeChecker,
		senderNode,
		map[*checker.Type]*Set[*checker.Type]{},
	)
}

func isUnsafeAssignmentWorker(
	t *checker.Type,
	receiver *checker.Type,
	typeChecker *checker.Checker,
	senderNode *ast.Node,
	visited map[*checker.Type]*Set[*checker.Type],
) (*checker.Type, *checker.Type, bool) {
	if IsTypeAnyType(t) {
		// Allow assignment of any ==> unknown.
		if IsTypeUnknownType(receiver) {
			return nil, nil, false
		}

		if !IsTypeAnyType(receiver) {
			return receiver, t, true
		}
	}

	typeAlreadyVisited, ok := visited[t]

	if ok {
		if typeAlreadyVisited.Has(receiver) {
			return nil, nil, false
		}
		typeAlreadyVisited.Add(receiver)
	} else {
		visited[t] = NewSetFromItems(receiver)
	}

	if checker.IsNonDeferredTypeReference(t) && checker.IsNonDeferredTypeReference(receiver) {
		// TODO - figure out how to handle cases like this,
		// where the types are assignable, but not the same type
		/*
		   function foo(): ReadonlySet<number> { return new Set<any>(); }

		   // and

		   type Test<T> = { prop: T }
		   type Test2 = { prop: string }
		   declare const a: Test<any>;
		   const b: Test2 = a;
		*/

		if t.Target() != receiver.Target() {
			// if the type references are different, assume safe, as we won't know how to compare the two types
			// the generic positions might not be equivalent for both types
			return nil, nil, false
		}

		if senderNode != nil && ast.IsNewExpression(senderNode) && ast.IsIdentifier(senderNode.Expression()) && senderNode.Expression().Text() == "Map" && len(senderNode.Arguments()) == 0 && senderNode.TypeArguments() == nil {
			// special case to handle `new Map()`
			// unfortunately Map's default empty constructor is typed to return `Map<any, any>` :(
			// https://github.com/typescript-eslint/typescript-eslint/issues/2109#issuecomment-634144396
			return nil, nil, false
		}

		typeArguments := checker.Checker_getTypeArguments(typeChecker, t)
		if typeArguments == nil {
			return nil, nil, false
		}
		receiverTypeArguments := checker.Checker_getTypeArguments(typeChecker, receiver)
		if receiverTypeArguments == nil {
			return nil, nil, false
		}

		for i, arg := range typeArguments {
			receiverArg := receiverTypeArguments[i]

			_, _, unsafe := isUnsafeAssignmentWorker(arg, receiverArg, typeChecker, senderNode, visited)
			if unsafe {
				return receiver, t, true
			}
		}

		return nil, nil, false
	}

	return nil, nil, false
}

/**
 * Returns the contextual type of a given node.
 * Contextual type is the type of the target the node is going into.
 * i.e. the type of a called function's parameter, or the defined type of a variable declaration
 */
func GetContextualType(
	typeChecker *checker.Checker,
	node *ast.Node,
) *checker.Type {
	parent := node.Parent

	if ast.IsCallExpression(parent) || ast.IsNewExpression(parent) {
		if node == parent.Expression() {
			// is the callee, so has no contextual type
			return nil
		}
	} else if ast.IsVariableDeclaration(parent) || ast.IsPropertyDeclaration(parent) || ast.IsParameterDeclaration(parent) {
		if t := parent.Type(); t != nil {
			return checker.Checker_getTypeFromTypeNode(typeChecker, t)
		}
		return nil
	} else if parent.Kind == ast.KindJsxExpression {
		// For JSX expressions, get the contextual type of the node itself
		// The contextual type flows from the JSX attribute down to the expression
		return checker.Checker_getContextualType(typeChecker, node, checker.ContextFlagsNone)
	} else if ast.IsIdentifier(node) && (ast.IsPropertyAssignment(parent) || ast.IsShorthandPropertyAssignment(parent)) {
		return checker.Checker_getContextualType(typeChecker, node, checker.ContextFlagsNone)
	} else if ast.IsBinaryExpression(parent) && parent.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && parent.AsBinaryExpression().Right == node {
		// is RHS of assignment
		return typeChecker.GetTypeAtLocation(parent.AsBinaryExpression().Left)
	} else if parent.Kind != ast.KindJsxExpression && !ast.IsTemplateSpan(parent) {
		// parent is not something we know we can get the contextual type of
		return nil
	}
	// TODO - support return statement checking

	return checker.Checker_getContextualType(typeChecker, node, checker.ContextFlagsNone)
}

func GetThisExpression(
	node *ast.Node,
) *ast.Node {
	for {
		node = ast.SkipParentheses(node)

		if ast.IsCallExpression(node) {
			node = node.Expression()
		} else if node.Kind == ast.KindThisKeyword {
			return node
		} else if ast.IsAccessExpression(node) {
			node = node.Expression()
		} else {
			break
		}
	}

	return nil
}

/*
 * If passed an enum member, returns the type of the parent. Otherwise,
 * returns itself.
 *
 * For example:
 * - `Fruit` --> `Fruit`
 * - `Fruit.Apple` --> `Fruit`
 */
func getBaseEnumType(typeChecker *checker.Checker, t *checker.Type) *checker.Type {
	symbol := checker.Type_symbol(t)
	if !IsSymbolFlagSet(symbol, ast.SymbolFlagsEnumMember) {
		return t
	}

	return typeChecker.GetTypeAtLocation(
		symbol.ValueDeclaration.Parent,
	)
}

/**
 * Retrieve only the Enum literals from a type. for example:
 * - 123 --> []
 * - {} --> []
 * - Fruit.Apple --> [Fruit.Apple]
 * - Fruit.Apple | Vegetable.Lettuce --> [Fruit.Apple, Vegetable.Lettuce]
 * - Fruit.Apple | Vegetable.Lettuce | 123 --> [Fruit.Apple, Vegetable.Lettuce]
 * - T extends Fruit --> [Fruit]
 */
func GetEnumLiterals(t *checker.Type) []*checker.Type {
	return Filter(
		UnionTypeParts(t),
		func(subType *checker.Type) bool {
			return IsTypeFlagSet(subType, checker.TypeFlagsEnumLiteral)
		},
	)
}

/**
 * A type can have 0 or more enum types. For example:
 * - 123 --> []
 * - {} --> []
 * - Fruit.Apple --> [Fruit]
 * - Fruit.Apple | Vegetable.Lettuce --> [Fruit, Vegetable]
 * - Fruit.Apple | Vegetable.Lettuce | 123 --> [Fruit, Vegetable]
 * - T extends Fruit --> [Fruit]
 */
func GetEnumTypes(
	typeChecker *checker.Checker,
	t *checker.Type,
) []*checker.Type {
	return Map(GetEnumLiterals(t), func(t *checker.Type) *checker.Type { return getBaseEnumType(typeChecker, t) })
}

type DiscriminatedAnyType uint8

const (
	DiscriminatedAnyTypeAny DiscriminatedAnyType = iota
	DiscriminatedAnyTypePromiseAny
	DiscriminatedAnyTypeAnyArray
	DiscriminatedAnyTypeSafe
)

/**
  * @returns `DiscriminatedAnyTypeAny ` if the type is `any`, `DiscriminatedAnyTypeAnyArray` if the type is `any[]` or `readonly any[]`, `DiscriminatedAnyTypePromiseAny` if the type is `Promise<any>`,
*          otherwise it returns `DiscriminatedAnyTypeSafe`.
*/
func DiscriminateAnyType(
	t *checker.Type,
	typeChecker *checker.Checker,
	program *compiler.Program,
	node *ast.Node,
) DiscriminatedAnyType {
	return discriminateAnyTypeWorker(t, typeChecker, program, node, NewSetFromItems[*checker.Type]())
}

func discriminateAnyTypeWorker(
	t *checker.Type,
	typeChecker *checker.Checker,
	program *compiler.Program,
	node *ast.Node,
	// TODO(port): do we really need visited here?
	visited *Set[*checker.Type],
) DiscriminatedAnyType {
	if visited.Has(t) {
		return DiscriminatedAnyTypeSafe
	}
	visited.Add(t)
	if IsTypeAnyType(t) {
		return DiscriminatedAnyTypeAny
	}
	if IsTypeAnyArrayType(t, typeChecker) {
		return DiscriminatedAnyTypeAnyArray
	}

	foundPromiseAny := TypeRecurser(t, func(t *checker.Type) bool {
		if !IsThenableType(typeChecker, node, t) {
			return false
		}
		awaitedType := checker.Checker_getAwaitedType(typeChecker, t)
		if awaitedType == nil {
			return false
		}
		awaitedAnyType := discriminateAnyTypeWorker(awaitedType, typeChecker, program, node, visited)
		return awaitedAnyType == DiscriminatedAnyTypeAny
	})

	if foundPromiseAny {
		return DiscriminatedAnyTypePromiseAny
	}

	return DiscriminatedAnyTypeSafe
}

func GetParentFunctionNode(
	node *ast.Node,
) *ast.Node {
	current := node.Parent
	for current != nil {
		if ast.IsFunctionLikeDeclaration(current) {
			return current
		}

		current = current.Parent
	}

	return nil
}

func IsHigherPrecedenceThanAwait(node *ast.Node) bool {
	nodePrecedence := ast.GetExpressionPrecedence(node)
	awaitPrecedence := ast.GetOperatorPrecedence(ast.KindAwaitExpression, ast.KindUnknown, ast.OperatorPrecedenceFlagsNone)
	return nodePrecedence > awaitPrecedence
}

// EslintLikePrecedence returns a numeric precedence matching ESLint's
// astUtils.getPrecedence so behavior parity holds for tsgo nodes that ESLint
// classifies (e.g. ArrowFunction = 1, ConditionalExpression = 3). Returns -1
// for TypeScript-only kinds (AsExpression, etc.) so the caller wraps them in
// parentheses defensively, matching ESLint's behavior on unknown node types.
func EslintLikePrecedence(node *ast.Node) int {
	if node == nil {
		return -1
	}
	switch node.Kind {
	case ast.KindArrowFunction:
		return 1
	case ast.KindYieldExpression:
		return 1
	case ast.KindConditionalExpression:
		return 3
	case ast.KindBinaryExpression:
		bin := node.AsBinaryExpression()
		if bin.OperatorToken == nil {
			return -1
		}
		op := bin.OperatorToken.Kind
		if op == ast.KindCommaToken {
			return 0
		}
		if ast.IsAssignmentOperator(op) {
			return 1
		}
		switch op {
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return 4
		case ast.KindAmpersandAmpersandToken:
			return 5
		case ast.KindBarToken:
			return 6
		case ast.KindCaretToken:
			return 7
		case ast.KindAmpersandToken:
			return 8
		case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken,
			ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			return 9
		case ast.KindLessThanToken, ast.KindLessThanEqualsToken,
			ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken,
			ast.KindInKeyword, ast.KindInstanceOfKeyword:
			return 10
		case ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken,
			ast.KindGreaterThanGreaterThanGreaterThanToken:
			return 11
		case ast.KindPlusToken, ast.KindMinusToken:
			return 12
		case ast.KindAsteriskToken, ast.KindSlashToken, ast.KindPercentToken:
			return 13
		case ast.KindAsteriskAsteriskToken:
			return 15
		}
		return 20
	case ast.KindPrefixUnaryExpression:
		op := node.AsPrefixUnaryExpression().Operator
		if op == ast.KindPlusPlusToken || op == ast.KindMinusMinusToken {
			return 17
		}
		return 16
	case ast.KindPostfixUnaryExpression:
		return 17
	case ast.KindAwaitExpression, ast.KindDeleteExpression,
		ast.KindVoidExpression, ast.KindTypeOfExpression:
		return 16
	case ast.KindCallExpression:
		return 18
	case ast.KindNewExpression:
		return 19
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindSuperKeyword,
		ast.KindNullKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword,
		ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression, ast.KindArrayLiteralExpression,
		ast.KindObjectLiteralExpression, ast.KindFunctionExpression,
		ast.KindClassExpression, ast.KindParenthesizedExpression,
		ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
		ast.KindTaggedTemplateExpression, ast.KindSpreadElement,
		ast.KindMetaProperty:
		return 20
	}
	// TypeScript-specific (AsExpression, SatisfiesExpression,
	// TypeAssertionExpression, ...) and any other kind ESLint does not
	// classify: return -1 to force wrapping for safety.
	return -1
}

func IsStrongPrecedenceNode(innerNode *ast.Node) bool {
	return ast.IsLiteralKind(innerNode.Kind) ||
		ast.IsBooleanLiteral(innerNode) ||
		ast.IsParenthesizedExpression(innerNode) ||
		innerNode.Kind == ast.KindIdentifier ||
		innerNode.Kind == ast.KindTypeReference ||
		innerNode.Kind == ast.KindTypeOperator ||
		innerNode.Kind == ast.KindArrayLiteralExpression ||
		innerNode.Kind == ast.KindObjectLiteralExpression ||
		innerNode.Kind == ast.KindPropertyAccessExpression ||
		innerNode.Kind == ast.KindElementAccessExpression ||
		innerNode.Kind == ast.KindCallExpression ||
		innerNode.Kind == ast.KindNewExpression ||
		innerNode.Kind == ast.KindTaggedTemplateExpression ||
		innerNode.Kind == ast.KindExpressionWithTypeArguments
}

func IsParenlessArrowFunction(node *ast.Node) bool {
	if !ast.IsArrowFunction(node) {
		return false
	}

	n := node.AsArrowFunction()

	return n.Parameters.End() == n.EqualsGreaterThanToken.Pos()
}

type MemberNameType uint8

const (
	MemberNameTypePrivate MemberNameType = iota
	MemberNameTypeQuoted
	MemberNameTypeNormal
	MemberNameTypeExpression
)

/**
 * Gets a string name representation of the name of the given MethodDefinition
 * or PropertyDefinition node, with handling for computed property names.
 */
func GetNameFromMember(sourceFile *ast.SourceFile, member *ast.Node) (string, MemberNameType) {
	switch member.Kind {
	case ast.KindIdentifier:
		return member.AsIdentifier().Text, MemberNameTypeNormal
	case ast.KindPrivateIdentifier:
		return member.AsPrivateIdentifier().Text, MemberNameTypePrivate
	case ast.KindComputedPropertyName:
		expr := member.AsComputedPropertyName().Expression
		// TODO(port): support boolean keywords, null keywords, etc
		if ast.IsLiteralExpression(expr) {
			text := expr.Text()
			if !scanner.IsValidIdentifier(text) {
				return "\"" + text + "\"", MemberNameTypeQuoted
			}
			return text, MemberNameTypeNormal
		}
	}

	r := TrimNodeTextRange(sourceFile, member)
	return sourceFile.Text()[r.Pos():r.End()], MemberNameTypeExpression
}

// GetPropertyInfo extracts the property node and formatted property name from a PropertyAccessExpression
// or ElementAccessExpression. Returns the property node and a formatted string like ".propertyName" or "[index]".
// Returns (nil, "") if the node is neither a property access nor an element access expression.
//
// Note: When called from ast.KindPropertyAccessExpression or ast.KindElementAccessExpression listeners,
// the returned property is guaranteed to be non-nil because:
//   - PropertyAccessExpression.Name() always returns a valid Identifier node
//   - ElementAccessExpression.ArgumentExpression always exists (after SkipParentheses)
//
// The nil return case only applies when called with nodes of other types.
func GetPropertyInfo(sourceFile *ast.SourceFile, node *ast.Node) (*ast.Node, string) {
	var property *ast.Node
	var propertyName string

	if ast.IsPropertyAccessExpression(node) {
		property = node.Name()
		loc := TrimNodeTextRange(sourceFile, property)
		propertyName = "." + sourceFile.Text()[loc.Pos():loc.End()]
	} else if ast.IsElementAccessExpression(node) {
		property = ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		loc := TrimNodeTextRange(sourceFile, property)
		propertyName = "[" + sourceFile.Text()[loc.Pos():loc.End()] + "]"
	}

	return property, propertyName
}

// IsInObjectLiteralMethod checks if a function node is defined as a method in an object literal.
// This includes both shorthand method syntax ({ methodA() {} }) and property assignment syntax
// ({ methodA: function() {} }). Returns true if the function is an object literal method.
func IsInObjectLiteralMethod(functionNode *ast.Node) bool {
	if functionNode == nil {
		return false
	}

	parent := functionNode.Parent
	if parent == nil {
		return false
	}

	// Direct object literal (shorthand method syntax): { methodA() {} }
	if ast.IsObjectLiteralExpression(parent) {
		return true
	}

	// Property assignment syntax: { methodA: function() {} } or { methodA: () => {} }
	if ast.IsPropertyAssignment(parent) || ast.IsShorthandPropertyAssignment(parent) || ast.IsMethodDeclaration(parent) {
		grandParent := parent.Parent
		if grandParent != nil && ast.IsObjectLiteralExpression(grandParent) {
			return true
		}
	}

	return false
}

// IsSymbolDeclaredInFile reports whether the given symbol has at least one
// declaration in the specified source file. Use this to distinguish locally
// declared symbols (shadowed) from globals provided by lib.d.ts.
func IsSymbolDeclaredInFile(symbol *ast.Symbol, sf *ast.SourceFile) bool {
	if symbol == nil {
		return false
	}
	for _, decl := range symbol.Declarations {
		if ast.GetSourceFileOfNode(decl) == sf {
			return true
		}
	}
	return false
}

// GetStaticPropertyName extracts the static name from a property name node.
// It handles Identifier, StringLiteral, NumericLiteral, and ComputedPropertyName
// (with static string, numeric, BigInt, or template literal expressions).
// Returns the name and whether it's a static (non-computed or statically-computable) name.
func GetStaticPropertyName(nameNode *ast.Node) (string, bool) {
	switch nameNode.Kind {
	case ast.KindIdentifier:
		return nameNode.AsIdentifier().Text, true
	case ast.KindStringLiteral:
		return nameNode.AsStringLiteral().Text, true
	case ast.KindNumericLiteral:
		return NormalizeNumericLiteral(nameNode.AsNumericLiteral().Text), true
	case ast.KindComputedPropertyName:
		expr := nameNode.AsComputedPropertyName().Expression
		switch expr.Kind {
		case ast.KindStringLiteral:
			return expr.AsStringLiteral().Text, true
		case ast.KindNumericLiteral:
			return NormalizeNumericLiteral(expr.AsNumericLiteral().Text), true
		case ast.KindBigIntLiteral:
			return NormalizeBigIntLiteral(expr.AsBigIntLiteral().Text), true
		case ast.KindNoSubstitutionTemplateLiteral:
			return expr.AsNoSubstitutionTemplateLiteral().Text, true
		case ast.KindNullKeyword:
			return "null", true
		case ast.KindTrueKeyword:
			return "true", true
		case ast.KindFalseKeyword:
			return "false", true
		case ast.KindRegularExpressionLiteral:
			return expr.AsRegularExpressionLiteral().Text, true
		}
		return "", false
	default:
		return "", false
	}
}

// NormalizeNumericLiteral parses a numeric literal text and returns its
// normalized string representation, matching ESLint's String(node.value) behavior.
// e.g., "0x1" -> "1", "1.0" -> "1", "1e2" -> "100"
func NormalizeNumericLiteral(text string) string {
	// ParseFloat doesn't handle JS octal (0o) or binary (0b) prefixes.
	// Use big.Int to handle arbitrary precision, then convert to float64
	// to match JavaScript's String(Number(...)) behavior.
	if len(text) > 2 && text[0] == '0' && (text[1] == 'o' || text[1] == 'O' || text[1] == 'b' || text[1] == 'B') {
		if n, ok := new(big.Int).SetString(text, 0); ok {
			f, _ := new(big.Float).SetInt(n).Float64()
			return strconv.FormatFloat(f, 'f', -1, 64)
		}
		return text
	}
	f, err := strconv.ParseFloat(text, 64)
	if err != nil {
		// ParseFloat returns +/-Inf with ErrRange for overflow (e.g. 1e309).
		// Only return raw text for true parse failures.
		if !math.IsInf(f, 0) {
			return text
		}
	}
	if math.IsInf(f, 1) {
		return "Infinity"
	}
	if math.IsInf(f, -1) {
		return "-Infinity"
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// NormalizeBigIntLiteral normalizes a BigInt literal to its decimal string
// representation, matching ESLint's String(node.value) behavior.
// e.g., "1n" -> "1", "0x1n" -> "1", "0o1n" -> "1", "0b1n" -> "1"
func NormalizeBigIntLiteral(text string) string {
	s := strings.TrimSuffix(text, "n")
	i, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return s
	}
	return strconv.FormatInt(i, 10)
}

// GetStaticExpressionValue returns the static string value of a literal expression,
// or ("", false) if the value cannot be statically determined.
//
// Unlike [GetStaticPropertyName], which is designed for property name nodes
// (Identifier, ComputedPropertyName, etc.) and treats Identifier as a static name,
// this function is for arbitrary value expressions — it only recognizes
// compile-time-constant literals and does NOT treat Identifier as static
// (since a[b] where b is a variable is dynamic).
//
// Supported node kinds:
//   - StringLiteral: returns the string value
//   - NumericLiteral: returns the normalized numeric string (e.g. "0x1" → "1")
//   - NoSubstitutionTemplateLiteral: returns the template text
//   - RegularExpressionLiteral: returns the source text (e.g. /foo/g),
//     matching JavaScript's implicit toString coercion when used as a property key
//
// This is the expression-level complement to [GetStaticPropertyName]:
// use GetStaticPropertyName for property name nodes (object keys, class members),
// and GetStaticExpressionValue for value positions (element access arguments, etc.).
func GetStaticExpressionValue(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text, true
	case ast.KindNumericLiteral:
		return NormalizeNumericLiteral(node.AsNumericLiteral().Text), true
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text, true
	case ast.KindRegularExpressionLiteral:
		return node.AsRegularExpressionLiteral().Text, true
	}
	return "", false
}

// IsSameReference reports whether two AST nodes refer to the same runtime value.
// It recursively compares member expression chains (PropertyAccessExpression,
// ElementAccessExpression), walking through the object/property structure.
//
// Behavior details:
//   - Parenthesized expressions and type assertions (as, <T>) are transparently
//     unwrapped on both sides via [ast.SkipOuterExpressions].
//   - Optional chaining is ignored: a.b and a?.b are considered the same reference,
//     matching ESLint's isSameReference semantics.
//   - Cross-syntax comparison is supported via static property names:
//     a.b and a['b'] are the same reference; a[0] and a['0'] likewise.
//   - For non-static element access (a[x]), falls back to comparing the argument
//     nodes structurally (same Kind + same Identifier/ThisKeyword).
//   - Function calls break the chain: a.b() and a.b() are NOT the same reference,
//     because each call may return a different value.
//
// This implements the same logic as ESLint's astUtils.isSameReference combined
// with astUtils.getStaticPropertyName, adapted for the TypeScript AST.
func IsSameReference(left, right *ast.Node) bool {
	left = ast.SkipOuterExpressions(left, ast.OEKParentheses|ast.OEKTypeAssertions)
	right = ast.SkipOuterExpressions(right, ast.OEKParentheses|ast.OEKTypeAssertions)

	if left == nil || right == nil {
		return false
	}

	// Base cases: Identifier and ThisKeyword.
	if left.Kind == ast.KindIdentifier && right.Kind == ast.KindIdentifier {
		return left.AsIdentifier().Text == right.AsIdentifier().Text
	}
	if left.Kind == ast.KindThisKeyword && right.Kind == ast.KindThisKeyword {
		return true
	}

	// Member expression comparison.
	if ast.IsAccessExpression(left) && ast.IsAccessExpression(right) {
		// Try static property name comparison first (handles cross-type: a.b vs a['b']).
		leftName, leftOK := AccessExpressionStaticName(left)
		if leftOK {
			rightName, rightOK := AccessExpressionStaticName(right)
			if rightOK && leftName == rightName {
				return IsSameReference(AccessExpressionObject(left), AccessExpressionObject(right))
			}
			return false
		}

		// Non-static: fall back to same-kind, same-index comparison (e.g. a[x] = a[x]).
		if left.Kind == right.Kind && left.Kind == ast.KindElementAccessExpression {
			leftArg := left.AsElementAccessExpression().ArgumentExpression
			rightArg := right.AsElementAccessExpression().ArgumentExpression
			if isSameSimpleNode(leftArg, rightArg) {
				return IsSameReference(left.AsElementAccessExpression().Expression, right.AsElementAccessExpression().Expression)
			}
		}
	}

	return false
}

// AccessExpressionStaticName returns the static property name of an access expression
// (PropertyAccessExpression or ElementAccessExpression), or ("", false) if not static.
func AccessExpressionStaticName(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		if name != nil {
			return name.Text(), true
		}
	case ast.KindElementAccessExpression:
		return GetStaticExpressionValue(node.AsElementAccessExpression().ArgumentExpression)
	}
	return "", false
}

// AccessExpressionObject returns the object expression of an access expression.
func AccessExpressionObject(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().Expression
	}
	return nil
}

// isSameSimpleNode checks if two nodes are the same simple reference (Identifier or ThisKeyword).
// Used as a fallback for comparing non-static element access arguments like a[x] vs a[x].
func isSameSimpleNode(left, right *ast.Node) bool {
	if left == nil || right == nil || left.Kind != right.Kind {
		return false
	}
	switch left.Kind {
	case ast.KindIdentifier:
		return left.AsIdentifier().Text == right.AsIdentifier().Text
	case ast.KindThisKeyword:
		return true
	}
	return false
}

// CollectBindingNames recursively extracts all identifier names from a binding
// pattern (ObjectBindingPattern, ArrayBindingPattern) or plain Identifier.
// For each identifier found, it calls the callback with the identifier node and its name.
func CollectBindingNames(nameNode *ast.Node, callback func(ident *ast.Node, name string)) {
	if nameNode == nil {
		return
	}

	switch nameNode.Kind {
	case ast.KindIdentifier:
		name := nameNode.AsIdentifier().Text
		if name != "" {
			callback(nameNode, name)
		}

	case ast.KindObjectBindingPattern:
		nameNode.ForEachChild(func(child *ast.Node) bool {
			if child.Kind == ast.KindBindingElement {
				bindingElem := child.AsBindingElement()
				if bindingElem != nil && bindingElem.Name() != nil {
					CollectBindingNames(bindingElem.Name(), callback)
				}
			}
			return false
		})

	case ast.KindArrayBindingPattern:
		nameNode.ForEachChild(func(child *ast.Node) bool {
			if child.Kind == ast.KindBindingElement {
				bindingElem := child.AsBindingElement()
				if bindingElem != nil && bindingElem.Name() != nil {
					CollectBindingNames(bindingElem.Name(), callback)
				}
			}
			return false
		})
	}
}

// IsNullLiteral checks if a node is the null keyword, unwrapping parentheses.
func IsNullLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return ast.SkipParentheses(node).Kind == ast.KindNullKeyword
}

// FindEnclosingScope finds the nearest function-like, class static block,
// module block, or source file scope for a node. Uses the tsgo public
// function IsFunctionLikeOrClassStaticBlockDeclaration.
// This is commonly needed by rules that walk write references or check
// variable scoping (e.g. prefer-const, no-var, no-class-assign).
func FindEnclosingScope(node *ast.Node) *ast.Node {
	return ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(n) {
			return true
		}
		switch n.Kind {
		case ast.KindSourceFile, ast.KindModuleBlock:
			return true
		}
		return false
	})
}

// IsDeclarationIdentifier checks if the node is the name (identifier) of a declaration.
// Unlike ast.IsDeclarationName(), this returns false for ShorthandPropertyAssignment
// names since they are both declaration names AND value references.
func IsDeclarationIdentifier(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		return parent.AsVariableDeclaration().Name() == node
	case ast.KindFunctionDeclaration:
		return parent.AsFunctionDeclaration().Name() == node
	case ast.KindParameter:
		return parent.AsParameterDeclaration().Name() == node
	case ast.KindClassDeclaration:
		return parent.AsClassDeclaration().Name() == node
	case ast.KindClassExpression:
		return parent.AsClassExpression().Name() == node
	case ast.KindFunctionExpression:
		return parent.AsFunctionExpression().Name() == node
	case ast.KindInterfaceDeclaration:
		return parent.AsInterfaceDeclaration().Name() == node
	case ast.KindTypeAliasDeclaration:
		return parent.AsTypeAliasDeclaration().Name() == node
	case ast.KindEnumDeclaration:
		return parent.AsEnumDeclaration().Name() == node
	case ast.KindModuleDeclaration:
		return parent.AsModuleDeclaration().Name() == node
	case ast.KindCatchClause:
		return parent.AsCatchClause().VariableDeclaration == node
	case ast.KindImportSpecifier:
		return parent.AsImportSpecifier().Name() == node
	case ast.KindImportClause:
		return parent.AsImportClause().Name() == node
	case ast.KindBindingElement:
		return parent.AsBindingElement().Name() == node
	case ast.KindNamespaceImport:
		return parent.AsNamespaceImport().Name() == node
	case ast.KindImportEqualsDeclaration:
		return parent.AsImportEqualsDeclaration().Name() == node
	case ast.KindEnumMember:
		return parent.AsEnumMember().Name() == node
	case ast.KindTypeParameter:
		return parent.AsTypeParameterDeclaration().Name() == node
	}
	return false
}

// GetDeclarationIdentifier returns the name node of a declaration.
func GetDeclarationIdentifier(decl *ast.Node) *ast.Node {
	if decl == nil {
		return nil
	}
	switch decl.Kind {
	case ast.KindVariableDeclaration:
		return decl.AsVariableDeclaration().Name()
	case ast.KindFunctionDeclaration:
		return decl.AsFunctionDeclaration().Name()
	case ast.KindClassDeclaration:
		return decl.AsClassDeclaration().Name()
	case ast.KindClassExpression:
		return decl.AsClassExpression().Name()
	case ast.KindInterfaceDeclaration:
		return decl.AsInterfaceDeclaration().Name()
	case ast.KindTypeAliasDeclaration:
		return decl.AsTypeAliasDeclaration().Name()
	case ast.KindEnumDeclaration:
		return decl.AsEnumDeclaration().Name()
	case ast.KindModuleDeclaration:
		return decl.AsModuleDeclaration().Name()
	case ast.KindImportSpecifier:
		return decl.AsImportSpecifier().Name()
	case ast.KindImportClause:
		return decl.AsImportClause().Name()
	case ast.KindNamespaceImport:
		return decl.AsNamespaceImport().Name()
	case ast.KindImportEqualsDeclaration:
		return decl.AsImportEqualsDeclaration().Name()
	case ast.KindParameter:
		return decl.AsParameterDeclaration().Name()
	case ast.KindBindingElement:
		return decl.AsBindingElement().Name()
	case ast.KindTypeParameter:
		return decl.AsTypeParameterDeclaration().Name()
	}
	return nil
}

// GetImportBindingNodes returns the local binding identifier nodes declared by
// an import statement. Returns nil for side-effect imports (e.g. `import 'foo'`).
// Handles ImportDeclaration (default, named, namespace) and ImportEqualsDeclaration.
func GetImportBindingNodes(node *ast.Node) []*ast.Node {
	var nodes []*ast.Node
	switch node.Kind {
	case ast.KindImportDeclaration:
		importDecl := node.AsImportDeclaration()
		if importDecl.ImportClause == nil {
			return nil
		}
		clause := importDecl.ImportClause.AsImportClause()
		if clause == nil {
			return nil
		}
		if clause.Name() != nil {
			nodes = append(nodes, clause.Name())
		}
		if clause.NamedBindings != nil {
			nb := clause.NamedBindings
			switch nb.Kind {
			case ast.KindNamespaceImport:
				nsImport := nb.AsNamespaceImport()
				if nsImport != nil && nsImport.Name() != nil {
					nodes = append(nodes, nsImport.Name())
				}
			case ast.KindNamedImports:
				namedImports := nb.AsNamedImports()
				if namedImports != nil && namedImports.Elements != nil {
					for _, elem := range namedImports.Elements.Nodes {
						importSpec := elem.AsImportSpecifier()
						if importSpec != nil && importSpec.Name() != nil {
							nodes = append(nodes, importSpec.Name())
						}
					}
				}
			}
		}
	case ast.KindImportEqualsDeclaration:
		importEquals := node.AsImportEqualsDeclaration()
		if importEquals.Name() != nil {
			nodes = append(nodes, importEquals.Name())
		}
	}
	return nodes
}

// VisitDestructuringIdentifiers calls fn for each identifier target in a
// destructuring assignment pattern (object/array literal on the left side
// of an assignment expression). Handles shorthand properties, renamed
// properties, default values, rest/spread, and arbitrary nesting.
// This does NOT handle declaration-level destructuring (BindingPattern) —
// use CollectBindingNames for that.
func VisitDestructuringIdentifiers(node *ast.Node, fn func(*ast.Node)) {
	node.ForEachChild(func(child *ast.Node) bool {
		switch child.Kind {
		case ast.KindIdentifier:
			fn(child)
		case ast.KindShorthandPropertyAssignment:
			shorthand := child.AsShorthandPropertyAssignment()
			if shorthand != nil && shorthand.Name() != nil {
				fn(shorthand.Name())
			}
		case ast.KindPropertyAssignment:
			pa := child.AsPropertyAssignment()
			if pa != nil && pa.Initializer != nil {
				if pa.Initializer.Kind == ast.KindIdentifier {
					fn(pa.Initializer)
				} else {
					VisitDestructuringIdentifiers(pa.Initializer, fn)
				}
			}
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression, ast.KindSpreadElement:
			VisitDestructuringIdentifiers(child, fn)
		case ast.KindSpreadAssignment:
			child.ForEachChild(func(gc *ast.Node) bool {
				if gc.Kind == ast.KindIdentifier {
					fn(gc)
				}
				return false
			})
		case ast.KindBinaryExpression:
			// Default value: [x = 5] → visit left side only
			be := child.AsBinaryExpression()
			if be != nil && be.Left != nil {
				if be.Left.Kind == ast.KindIdentifier {
					fn(be.Left)
				} else {
					VisitDestructuringIdentifiers(be.Left, fn)
				}
			}
		}
		return false
	})
}

// IsSpecificMemberAccess reports whether `node` is a member access of the
// form `<objectName>.<methodName>`. Both dot (`Object.defineProperty`) and
// bracket-with-static-string (`Object['defineProperty']`) forms are matched,
// each transparently unwrapping parentheses (e.g. `(Object).defineProperty`)
// and optional chaining (`Object?.defineProperty`, `Object?.['defineProperty']`).
// Mirrors ESLint's `astUtils.isSpecificMemberAccess`.
//
// If `objectName` is the empty string, the object identity check is skipped
// — any expression on the left of the method is accepted — matching ESLint's
// behavior when the `objectName` argument is `null`.
func IsSpecificMemberAccess(node *ast.Node, objectName, methodName string) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	var obj *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		pae := node.AsPropertyAccessExpression()
		name := pae.Name()
		if name == nil || !ast.IsIdentifier(name) || name.AsIdentifier().Text != methodName {
			return false
		}
		obj = pae.Expression
	case ast.KindElementAccessExpression:
		eae := node.AsElementAccessExpression()
		argText, ok := GetStaticExpressionValue(ast.SkipParentheses(eae.ArgumentExpression))
		if !ok || argText != methodName {
			return false
		}
		obj = eae.Expression
	default:
		return false
	}
	if objectName == "" {
		return true
	}
	obj = ast.SkipParentheses(obj)
	return obj != nil && ast.IsIdentifier(obj) && obj.AsIdentifier().Text == objectName
}

// AreNodesStructurallyEqual reports whether two AST subtrees have identical
// syntactic shape and leaf values, transparently unwrapping
// ParenthesizedExpression on both sides at every level. Useful for rules that
// compare computed keys, duplicate case expressions, or any pattern that must
// be evaluated at the source-syntax level rather than by semantic reference
// identity (for which see [IsSameReference]).
//
// Leaf comparison:
//   - Identifier / PrivateIdentifier: `.Text` equality.
//   - StringLiteral / NoSubstitutionTemplateLiteral / TemplateHead / Middle /
//     Tail / RegularExpressionLiteral: textual equality.
//   - NumericLiteral: normalized numeric value equality (e.g. `0x10` == `16`).
//   - BigIntLiteral: normalized bigint value equality (e.g. `0x1n` == `1n`).
//   - All other kinds (keyword tokens, punctuation tokens, composite nodes):
//     Kind must match, and the non-nil children visited by [ast.Node.ForEachChild]
//     must be pairwise structurally equal in order.
//
// Comments and whitespace are not part of the AST and are therefore ignored
// (so `a+b` and `a + b` compare equal). Optional chaining IS preserved
// (`a.b` != `a?.b`). Type-only syntax (`as T`, `<T>x`, `x!`, `x satisfies T`)
// is compared as-is — callers that want to see through it should strip it
// first via [ast.SkipOuterExpressions] before calling this helper.
func AreNodesStructurallyEqual(a, b *ast.Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	a = ast.SkipParentheses(a)
	b = ast.SkipParentheses(b)
	if a == nil || b == nil {
		return a == b
	}
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case ast.KindIdentifier:
		return a.AsIdentifier().Text == b.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return a.AsPrivateIdentifier().Text == b.AsPrivateIdentifier().Text
	case ast.KindStringLiteral:
		return a.AsStringLiteral().Text == b.AsStringLiteral().Text
	case ast.KindNoSubstitutionTemplateLiteral:
		return a.AsNoSubstitutionTemplateLiteral().Text == b.AsNoSubstitutionTemplateLiteral().Text
	case ast.KindNumericLiteral:
		// Note: tsgo already normalizes numeric literals at parse time
		// (`0x1` / `1e2` / `1.0` are all stored as their decimal form).
		// Normalize again to be explicit about the intent; two literals
		// that differ only in source form (e.g. `0x1` vs `1`) are treated
		// as equal here. This is slightly more forgiving than ESLint's
		// token-level comparison, which would see them as distinct — but
		// the raw source form is not recoverable from the tsgo AST
		// without a *SourceFile, which we deliberately don't take.
		return NormalizeNumericLiteral(a.AsNumericLiteral().Text) ==
			NormalizeNumericLiteral(b.AsNumericLiteral().Text)
	case ast.KindBigIntLiteral:
		return NormalizeBigIntLiteral(a.AsBigIntLiteral().Text) ==
			NormalizeBigIntLiteral(b.AsBigIntLiteral().Text)
	case ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail,
		ast.KindRegularExpressionLiteral:
		return a.Text() == b.Text()
	case ast.KindPrefixUnaryExpression:
		// tsgo stores the operator as a Kind field, not as a child node, so
		// ForEachChild would otherwise collapse `+x` and `-x` (both have one
		// child, the Operand). Compare the Operator field before recursing.
		ap, bp := a.AsPrefixUnaryExpression(), b.AsPrefixUnaryExpression()
		return ap.Operator == bp.Operator && AreNodesStructurallyEqual(ap.Operand, bp.Operand)
	case ast.KindPostfixUnaryExpression:
		// Same gotcha as PrefixUnaryExpression — ForEachChild omits Operator.
		ap, bp := a.AsPostfixUnaryExpression(), b.AsPostfixUnaryExpression()
		return ap.Operator == bp.Operator && AreNodesStructurallyEqual(ap.Operand, bp.Operand)
	case ast.KindMetaProperty:
		// `new.target` and `import.meta` both use MetaProperty; the meta
		// keyword lives in KeywordToken (Kind), which ForEachChild doesn't
		// visit. In practice `name` (target vs meta) already distinguishes
		// them, but compare the keyword explicitly for principled alignment.
		am, bm := a.AsMetaProperty(), b.AsMetaProperty()
		return am.KeywordToken == bm.KeywordToken && AreNodesStructurallyEqual(am.Name(), bm.Name())
	}
	// Composite / pure-token kinds: compare children pairwise. Token kinds
	// without children (operators, keywords) fall through the empty loop and
	// return true, which is correct — Kind already uniquely identifies them.
	var aKids, bKids []*ast.Node
	a.ForEachChild(func(c *ast.Node) bool {
		aKids = append(aKids, c)
		return false
	})
	b.ForEachChild(func(c *ast.Node) bool {
		bKids = append(bKids, c)
		return false
	})
	if len(aKids) != len(bKids) {
		return false
	}
	for i := range aKids {
		if !AreNodesStructurallyEqual(aKids[i], bKids[i]) {
			return false
		}
	}
	return true
}

// HasSameTokens reports whether two nodes produce the same token stream when
// viewed at the raw-source level — matching ESLint's
// `sourceCode.getTokens(a)` vs `sourceCode.getTokens(b)` semantics, which
// preserves the original source form of each literal. Unlike
// [AreNodesStructurallyEqual], this helper distinguishes:
//
//   - `'a'` vs `"a"` (different quote style)
//   - `0x1` vs `1` (different numeric source form)
//   - `1n` vs `0x1n` (different bigint source form)
//   - `1e2` vs `100` / `1.0` vs `1`
//
// Implementation: we recurse on the AST using [ast.SkipParentheses] and
// [ast.Node.ForEachChild]. At leaf nodes (no children — identifiers,
// literals, keyword tokens) we compare the raw source slice via
// [scanner.GetSourceTextOfNodeFromSourceFile]. For composite nodes we
// recurse on children pairwise AND scan the "gaps" between children
// (and before/after the first/last child) with [scanner.Scanner] to pick
// up punctuation, keyword tokens, and operators that tsgo's ForEachChild
// does not visit — `(` `)` `,` `.` between children of a CallExpression,
// the `+`/`-` operator of a PrefixUnaryExpression, the `new`/`import`
// keyword of a MetaProperty, and so on. Whitespace and comments in a
// gap are trivia (scanner skips them), so the comparison is
// whitespace-insensitive exactly like ESLint's `getTokens`.
//
// Parens: stripped once at the top level (matches ESLint / ESTree, where
// outer parens wrapping an operand aren't nodes and their tokens fall
// outside the operand's range). Parens INSIDE a compound expression —
// e.g. `(x).y` — ARE visible tokens in ESLint's view, and preserved here
// by the recursion not calling SkipParentheses again.
//
// Templates: TemplateExpression / TemplateSpan children already cover
// the whole template source range contiguously, so the gap between any
// two children inside a template is empty. This means gap scanning never
// enters template-expression context and thus never needs the scanner's
// `ReScanTemplateToken` (which isn't exposed through the shim).
//
// Use this helper when porting an ESLint rule whose oracle is token-level
// equality (e.g. `no-self-compare`'s `hasSameTokens`); use
// [AreNodesStructurallyEqual] when the rule's oracle is structural AST
// equality and literal-form / trivia differences should NOT matter (e.g.
// duplicate case detection).
func HasSameTokens(sourceFile *ast.SourceFile, a, b *ast.Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	return hasSameTokens(sourceFile, ast.SkipParentheses(a), ast.SkipParentheses(b))
}

// hasSameTokens is the recursive core. It does NOT call SkipParentheses
// on its inputs — parens nested inside a compound expression are visible
// tokens in ESLint's per-node getTokens view (e.g. `(x).y` has tokens
// `[(, x, ), ., y]`), so a recursive paren strip would collapse them.
func hasSameTokens(sf *ast.SourceFile, a, b *ast.Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Kind != b.Kind {
		return false
	}
	aKids, bKids := collectKids(a), collectKids(b)
	// Leaves (no children via ForEachChild). Two sub-classes collide here:
	//   1. True leaves (Identifier, Literal, keyword tokens) — raw source
	//      text is exactly the single token's text, so raw-text equality
	//      is correct.
	//   2. Empty composites (`[]`, `{}`) — raw text includes the brackets
	//      AND any whitespace/comments inside, so raw-text would
	//      incorrectly distinguish `[]` from `[ ]` or `[\n/*c*/\n]`. ESLint's
	//      `getTokens` treats these as equivalent (only `[`, `]` tokens).
	// For class 2 we scan tokens; for class 1 we keep the raw-text shortcut
	// because some leaf kinds (e.g. TemplateHead / TemplateTail inside a
	// TemplateExpression) cannot be re-scanned standalone — the scanner
	// needs `ReScanTemplateToken` context that isn't exposed through the
	// shim.
	if len(aKids) == 0 && len(bKids) == 0 {
		switch a.Kind {
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
			return sameTokensInRange(sf, a.Pos(), a.End(), b.Pos(), b.End())
		}
		return scanner.GetSourceTextOfNodeFromSourceFile(sf, a, false) ==
			scanner.GetSourceTextOfNodeFromSourceFile(sf, b, false)
	}
	if len(aKids) != len(bKids) {
		return false
	}
	// Compare children pairwise AND compare the token sequences living in
	// the gaps between children (and the prefix / suffix gaps). Gap tokens
	// are the operators / punctuation / keywords that ForEachChild does not
	// yield as nodes — `(` `)` `,` `.` between call arguments, `+` / `-` for
	// PrefixUnaryExpression, `new` / `import` for MetaProperty, and so on.
	//
	// With zero children the loop is skipped and the two nodes are compared
	// entirely via the trailing gap scan — this covers both simple leaves
	// (identifiers, literals, keyword tokens: one token each) AND empty
	// composites (`[]`, `{}`: bracket/brace tokens only). The scanner treats
	// whitespace and comments as trivia, so `[]` and `[ ]` compare equal —
	// matching ESLint's `getTokens` semantics.
	prevA, prevB := a.Pos(), b.Pos()
	for i := range aKids {
		if !sameTokensInRange(sf, prevA, aKids[i].Pos(), prevB, bKids[i].Pos()) {
			return false
		}
		if !hasSameTokens(sf, aKids[i], bKids[i]) {
			return false
		}
		prevA, prevB = aKids[i].End(), bKids[i].End()
	}
	return sameTokensInRange(sf, prevA, a.End(), prevB, b.End())
}

func collectKids(n *ast.Node) []*ast.Node {
	var out []*ast.Node
	n.ForEachChild(func(c *ast.Node) bool { out = append(out, c); return false })
	return out
}

// sameTokensInRange reports whether scanning [aStart, aEnd) and
// [bStart, bEnd) produces the same sequence of (kind, raw text) pairs.
// Trivia (whitespace, comments) is skipped by the scanner, matching
// ESLint's `getTokens` which excludes comments by default.
func sameTokensInRange(sf *ast.SourceFile, aStart, aEnd, bStart, bEnd int) bool {
	var sa, sb *scanner.Scanner
	if aStart < aEnd {
		sa = scanner.GetScannerForSourceFile(sf, aStart)
	}
	if bStart < bEnd {
		sb = scanner.GetScannerForSourceFile(sf, bStart)
	}
	liveA := func() bool {
		return sa != nil && sa.Token() != ast.KindEndOfFile && sa.TokenStart() < aEnd && sa.TokenEnd() <= aEnd
	}
	liveB := func() bool {
		return sb != nil && sb.Token() != ast.KindEndOfFile && sb.TokenStart() < bEnd && sb.TokenEnd() <= bEnd
	}
	for {
		la, lb := liveA(), liveB()
		if !la && !lb {
			return true
		}
		if la != lb || sa.Token() != sb.Token() || sa.TokenText() != sb.TokenText() {
			return false
		}
		sa.Scan()
		sb.Scan()
	}
}

// IsArgumentOfSpecificCall reports whether `node` sits at argument position
// `index` of a call to `<objectName>.<methodName>(...)` — covering optional
// chaining and parenthesized callee expressions, e.g. `(Object?.defineProperty)(...)`.
// This is the common shape for detecting property-descriptor arguments in
// `Object.defineProperty` / `Reflect.defineProperty`, mutation targets in
// `Object.assign`, and similar well-known API calls.
func IsArgumentOfSpecificCall(node *ast.Node, index int, objectName, methodName string) bool {
	if node == nil || node.Parent == nil || node.Parent.Kind != ast.KindCallExpression {
		return false
	}
	call := node.Parent.AsCallExpression()
	if call.Arguments == nil {
		return false
	}
	args := call.Arguments.Nodes
	if index < 0 || index >= len(args) || args[index] != node {
		return false
	}
	return IsSpecificMemberAccess(call.Expression, objectName, methodName)
}
