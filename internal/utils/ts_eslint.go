package utils

import (
	"math"
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
				typeParamDecl := decls[0].AsTypeParameter()
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
	return ast.IsParameter(decl) && decl.AsParameterDeclaration().DotDotDotToken != nil
}

/**
 * Gets the declaration for the given variable
 */
func GetDeclaration(
	typeChecker *checker.Checker,
	node *ast.Node,
) *ast.Declaration {
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
	} else if ast.IsVariableDeclaration(parent) || ast.IsPropertyDeclaration(parent) || ast.IsParameter(parent) {
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
