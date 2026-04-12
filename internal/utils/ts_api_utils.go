package utils

import (
	"math"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
)

func UnionTypeParts(t *checker.Type) []*checker.Type {
	if IsUnionType(t) {
		return t.Types()
	}
	return []*checker.Type{t}
}
func IntersectionTypeParts(t *checker.Type) []*checker.Type {
	if IsIntersectionType(t) {
		return t.Types()
	}
	return []*checker.Type{t}
}

func IsTypeFlagSet(t *checker.Type, flags checker.TypeFlags) bool {
	return t != nil && checker.Type_flags(t)&flags != 0
}

func IsIntrinsicType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsIntrinsic)
}

func IsIntrinsicErrorType(t *checker.Type) bool {
	return IsIntrinsicType(t) && t.AsIntrinsicType().IntrinsicName() == "error"
}

func IsIntrinsicVoidType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsVoid)
}

func IsUnionType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsUnion)
}
func IsIntersectionType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsIntersection)
}
func IsTypeAnyType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsAny)
}
func IsTypeUnknownType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsUnknown)
}
func IsObjectType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsObject)
}
func IsTypeParameter(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsTypeParameter)
}
func IsBooleanLiteralType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsBooleanLiteral)
}

// IsTrueLiteralType checks if the type is the boolean literal `true`.
// Handles both TypeFlagsBooleanLiteral (literal `true`) and TypeFlagsBoolean
// (widened boolean that is actually `true` from const narrowing).
func IsTrueLiteralType(t *checker.Type) bool {
	flags := checker.Type_flags(t)
	if flags&checker.TypeFlagsBooleanLiteral != 0 {
		val := t.AsLiteralType().Value()
		if b, ok := val.(bool); ok {
			return b
		}
		return false
	}
	// For TypeFlagsBoolean used by const narrowing (e.g., `const x = true`)
	if flags&checker.TypeFlagsBoolean != 0 && IsIntrinsicType(t) {
		return t.AsIntrinsicType().IntrinsicName() == "true"
	}
	return false
}

// IsFalseLiteralType checks if the type is the boolean literal `false`.
func IsFalseLiteralType(t *checker.Type) bool {
	flags := checker.Type_flags(t)
	if flags&checker.TypeFlagsBooleanLiteral != 0 {
		val := t.AsLiteralType().Value()
		if b, ok := val.(bool); ok {
			return !b
		}
		return false
	}
	if flags&checker.TypeFlagsBoolean != 0 && IsIntrinsicType(t) {
		return t.AsIntrinsicType().IntrinsicName() == "false"
	}
	return false
}

// IsTypeFlagSetWithUnion checks type flags, iterating through union constituents.
// This matches typescript-eslint's isTypeFlagSet which aggregates union constituent flags.
func IsTypeFlagSetWithUnion(t *checker.Type, flags checker.TypeFlags) bool {
	for _, part := range UnionTypeParts(t) {
		if IsTypeFlagSet(part, flags) {
			return true
		}
	}
	return false
}

// IsNullableType checks if the type includes null, undefined, void, any or unknown.
// This matches typescript-eslint's isNullableType utility.
func IsNullableType(t *checker.Type) bool {
	return Some(UnionTypeParts(t), func(part *checker.Type) bool {
		return IsTypeFlagSet(part, checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid)
	})
}

// IsPossiblyFalsy checks if any union constituent of the type could be falsy.
func IsPossiblyFalsy(t *checker.Type) bool {
	return Some(UnionTypeParts(t), func(part *checker.Type) bool {
		return isConstituentPossiblyFalsy(part)
	})
}

// IsPossiblyTruthy checks if any union constituent of the type could be truthy.
func IsPossiblyTruthy(t *checker.Type) bool {
	return Some(UnionTypeParts(t), func(part *checker.Type) bool {
		return isConstituentPossiblyTruthy(part)
	})
}

func GetCallSignatures(typeChecker *checker.Checker, t *checker.Type) []*checker.Signature {
	return checker.Checker_getSignaturesOfType(typeChecker, t, checker.SignatureKindCall)
}
func GetConstructSignatures(typeChecker *checker.Checker, t *checker.Type) []*checker.Signature {
	return checker.Checker_getSignaturesOfType(typeChecker, t, checker.SignatureKindConstruct)
}

// ex. getCallSignaturesOfType
func CollectAllCallSignatures(typeChecker *checker.Checker, t *checker.Type) []*checker.Signature {
	if IsUnionType(t) {
		signatures := []*checker.Signature{}
		for _, subtype := range t.Types() {
			signatures = append(signatures, GetCallSignatures(typeChecker, subtype)...)
		}
		return signatures
	}
	if IsIntersectionType(t) {
		var signatures []*checker.Signature
		for _, subtype := range t.Types() {
			sig := GetCallSignatures(typeChecker, subtype)
			if len(sig) != 0 {
				if signatures != nil {
					return []*checker.Signature{}
				}
				signatures = sig
			}
		}
		if signatures == nil {
			return []*checker.Signature{}
		}
		return signatures
	}
	return checker.Checker_getSignaturesOfType(typeChecker, t, checker.SignatureKindCall)
}

func IsSymbolFlagSet(symbol *ast.Symbol, flag ast.SymbolFlags) bool {
	return symbol != nil && symbol.Flags&flag != 0
}

func IsCallback(
	typeChecker *checker.Checker,
	param *ast.Symbol,
	node *ast.Node,
) bool {
	t := checker.Checker_getApparentType(typeChecker, typeChecker.GetTypeOfSymbolAtLocation(param, node))

	if param.ValueDeclaration != nil && ast.IsParameter(param.ValueDeclaration) && param.ValueDeclaration.AsParameterDeclaration().DotDotDotToken != nil {
		t = checker.Checker_getIndexTypeOfType(typeChecker, t, checker.Checker_numberType(typeChecker))
		if t == nil {
			return false
		}
	}

	for _, subType := range UnionTypeParts(t) {
		if len(GetCallSignatures(typeChecker, subType)) != 0 {
			return true
		}
	}

	return false
}

// TODO(note): why there is no IntersectionTypeParts
func IsThenableType(
	typeChecker *checker.Checker,
	node *ast.Node,
	t *checker.Type,
) bool {
	if t == nil {
		t = typeChecker.GetTypeAtLocation(node)
	}
	for _, typePart := range UnionTypeParts(checker.Checker_getApparentType(typeChecker, t)) {
		then := checker.Checker_getPropertyOfType(typeChecker, typePart, "then")
		if then == nil {
			continue
		}

		thenType := typeChecker.GetTypeOfSymbolAtLocation(then, node)

		for _, subTypePart := range UnionTypeParts(thenType) {
			for _, signature := range checker.Checker_getSignaturesOfType(typeChecker, subTypePart, checker.SignatureKindCall) {
				if len(checker.Signature_parameters(signature)) != 0 && IsCallback(typeChecker, checker.Signature_parameters(signature)[0], node) {
					return true
				}
			}
		}
	}
	return false
}

func GetWellKnownSymbolPropertyOfType(t *checker.Type, name string, typeChecker *checker.Checker) *ast.Symbol {
	return checker.Checker_getPropertyOfType(typeChecker, t, checker.Checker_getPropertyNameForKnownSymbolName(typeChecker, name))
}

// getChildrenFromNonJSDocNode from github.com/microsoft/typescript-go/internal/ls/utilities.go
func GetChildren(node *ast.Node, sourceFile *ast.SourceFile) []*ast.Node {
	var childNodes []*ast.Node
	node.ForEachChild(func(child *ast.Node) bool {
		// Skip reparsed nodes (synthesized from JSDoc) as they have positions
		// in the JSDoc comment range rather than the actual code range, which
		// causes token cache parent mismatches during gap-filling.
		if child.Flags&ast.NodeFlagsReparsed != 0 {
			return false
		}
		childNodes = append(childNodes, child)
		return false
	})
	var children []*ast.Node
	pos := node.Pos()
	for _, child := range childNodes {
		scanner := scanner.GetScannerForSourceFile(sourceFile, pos)
		for pos < child.Pos() {
			token := scanner.Token()
			tokenFullStart := scanner.TokenFullStart()
			tokenEnd := scanner.TokenEnd()
			children = append(children, sourceFile.GetOrCreateToken(token, tokenFullStart, tokenEnd, node, ast.TokenFlagsNone))
			pos = tokenEnd
			scanner.Scan()
		}
		children = append(children, child)
		pos = child.End()
	}
	scanner := scanner.GetScannerForSourceFile(sourceFile, pos)
	for pos < node.End() {
		token := scanner.Token()
		tokenFullStart := scanner.TokenFullStart()
		tokenEnd := scanner.TokenEnd()
		children = append(children, sourceFile.GetOrCreateToken(token, tokenFullStart, tokenEnd, node, ast.TokenFlagsNone))
		pos = tokenEnd
		scanner.Scan()
	}
	return children
}

// Checks if a given compiler option is enabled, accounting for whether all flags
// (except `strictPropertyInitialization`) have been enabled by `strict: true`.
//
// @category Compiler Options
//
// @example
//
//	const optionsLenient = {
//		noImplicitAny: true,
//	};
//
// isStrictCompilerOptionEnabled(optionsLenient, "noImplicitAny"); // true
// isStrictCompilerOptionEnabled(optionsLenient, "noImplicitThis"); // false
//
// @example
//
//	const optionsStrict = {
//		noImplicitThis: false,
//		strict: true,
//	};
//
// isStrictCompilerOptionEnabled(optionsStrict, "noImplicitAny"); // true
// isStrictCompilerOptionEnabled(optionsStrict, "noImplicitThis"); // false
func IsStrictCompilerOptionEnabled(
	options *core.CompilerOptions,
	option core.Tristate,
) bool {
	if options.Strict.IsTrue() {
		return option.IsTrueOrUnknown()
	}
	return option.IsTrue()
	// return (
	// 	(options.strict ? options[option] !== false : options[option] === true) &&
	// 	(option !== "strictPropertyInitialization" ||
	// 		isStrictCompilerOptionEnabled(options, "strictNullChecks"))
	// );
}

// Port https://github.com/JoshuaKGoldberg/ts-api-utils/blob/491c0374725a5dd64632405efea101f20ed5451f/src/tokens.ts#L34
//
// Iterates over all tokens of `node`
//
// @category Nodes - Other Utilities
//
// @example
//
// declare const node: ts.Node;
//
//	forEachToken(node, (token) => {
//		console.log("Found token:", token.getText());
//	});
//
// @param node The node whose tokens should be visited
// @param callback Is called for every token contained in `node`
func ForEachToken(node *ast.Node, callback func(token *ast.Node), sourceFile *ast.SourceFile) {
	queue := make([]*ast.Node, 0)

	for {
		if ast.IsTokenKind(node.Kind) {
			callback(node)
		} else {
			children := GetChildren(node, sourceFile)
			for i := len(children) - 1; i >= 0; i-- {
				queue = append(queue, children[i])
			}
		}

		if len(queue) == 0 {
			break
		}

		node = queue[len(queue)-1]
		queue = queue[:len(queue)-1]
	}
}

// Port https://github.com/JoshuaKGoldberg/ts-api-utils/blob/491c0374725a5dd64632405efea101f20ed5451f/src/comments.ts#L37C17-L37C31
//
// Iterates over all comments owned by `node` or its children.
//
// @category Nodes - Other Utilities
//
// @example
//
// declare const node: ts.Node;
//
//	forEachComment(node, (fullText, comment) => {
//	   console.log(`Found comment at position ${comment.pos}: '${fullText}'.`);
//	});
func ForEachComment(node *ast.Node, callback func(comment *ast.CommentRange), sourceFile *ast.SourceFile) {
	fullText := sourceFile.Text()
	notJsx := sourceFile.LanguageVariant != core.LanguageVariantJSX

	ForEachToken(
		node,
		func(token *ast.Node) {
			if token.Pos() == token.End() {
				return
			}

			if token.Kind != ast.KindJsxText {
				pos := token.Pos()
				if pos == 0 {
					pos = len(scanner.GetShebang(fullText))
				}

				for comment := range scanner.GetLeadingCommentRanges(&ast.NodeFactory{}, fullText, pos) {
					callback(&comment)
				}
			}

			if notJsx || canHaveTrailingTrivia(token) {
				for comment := range scanner.GetTrailingCommentRanges(&ast.NodeFactory{}, fullText, token.End()) {
					callback(&comment)
				}
				return
			}
		},
		sourceFile,
	)
}

// Port https://github.com/JoshuaKGoldberg/ts-api-utils/blob/491c0374725a5dd64632405efea101f20ed5451f/src/comments.ts#L84
//
// Exclude trailing positions that would lead to scanning for trivia inside `JsxText`.
// @internal
func canHaveTrailingTrivia(token *ast.Node) bool {
	switch token.Kind {
	case ast.KindCloseBraceToken:
		// after a JsxExpression inside a JsxElement's body can only be other JsxChild, but no trivia
		return token.Parent.Kind != ast.KindJsxExpression || !isJsxElementOrFragment(token.Parent.Parent)
	case ast.KindGreaterThanToken:
		switch token.Parent.Kind {
		case ast.KindJsxClosingElement:
		case ast.KindJsxClosingFragment:
			// there's only trailing trivia if it's the end of the top element
			return !isJsxElementOrFragment(token.Parent.Parent.Parent)
		case ast.KindJsxOpeningElement:
			// if end is not equal, this is part of the type arguments list. in all other cases it would be inside the element body
			return token.End() != token.Parent.End()
		case ast.KindJsxOpeningFragment:
			return false // would be inside the fragment
		case ast.KindJsxSelfClosingElement:
			// if end is not equal, this is part of the type arguments list
			// there's only trailing trivia if it's the end of the top element
			return token.End() != token.Parent.End() || !isJsxElementOrFragment(token.Parent.Parent)
		}
	}

	return true
}

// Port https://github.com/JoshuaKGoldberg/ts-api-utils/blob/491c0374725a5dd64632405efea101f20ed5451f/src/comments.ts#L118
//
// Test if a node is a `JsxElement` or `JsxFragment`.
// @internal
func isJsxElementOrFragment(node *ast.Node) bool {
	return node.Kind == ast.KindJsxElement || node.Kind == ast.KindJsxFragment
}

// isNumberLiteralZeroOrNaN checks if a number literal type value is 0 or NaN.
// tsgo stores number literal values as a named float64 type,
// so we use ValueToString for reliable string conversion and then parse.
func isNumberLiteralZeroOrNaN(val interface{}) bool {
	s := checker.ValueToString(val)
	if s == "0" || s == "-0" || s == "NaN" {
		return true
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}
	return f == 0 || math.IsNaN(f)
}

func isConstituentPossiblyFalsy(t *checker.Type) bool {
	flags := checker.Type_flags(t)
	if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return true
	}
	if flags&checker.TypeFlagsTypeVariable != 0 {
		return true
	}
	if flags&checker.TypeFlagsNever != 0 {
		return false
	}
	if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
		return true
	}
	if flags&checker.TypeFlagsBooleanLike != 0 {
		if IsTrueLiteralType(t) {
			return false // `true` literal is never falsy
		}
		if IsFalseLiteralType(t) {
			return true // `false` literal is always falsy
		}
		return true // general `boolean` is possibly falsy
	}
	if flags&checker.TypeFlagsStringLiteral != 0 {
		if s, ok := t.AsLiteralType().Value().(string); ok {
			return s == ""
		}
		return false
	}
	if flags&checker.TypeFlagsString != 0 {
		return true
	}
	if flags&checker.TypeFlagsNumberLiteral != 0 {
		return isNumberLiteralZeroOrNaN(t.AsLiteralType().Value())
	}
	if flags&checker.TypeFlagsNumber != 0 {
		return true
	}
	if flags&checker.TypeFlagsBigIntLiteral != 0 {
		return checker.ValueToString(t.AsLiteralType().Value()) == "0n"
	}
	if flags&checker.TypeFlagsBigInt != 0 {
		return true
	}
	if flags&checker.TypeFlagsEnumLiteral != 0 {
		return true
	}
	if flags&checker.TypeFlagsUnion != 0 {
		return IsPossiblyFalsy(t)
	}
	if flags&checker.TypeFlagsIntersection != 0 {
		return Some(t.Types(), isConstituentPossiblyFalsy)
	}
	return false
}

func isConstituentPossiblyTruthy(t *checker.Type) bool {
	flags := checker.Type_flags(t)
	if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return true
	}
	if flags&checker.TypeFlagsTypeVariable != 0 {
		return true
	}
	if flags&checker.TypeFlagsNever != 0 {
		return false
	}
	if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
		return false
	}
	if flags&checker.TypeFlagsBooleanLike != 0 {
		if IsFalseLiteralType(t) {
			return false // `false` literal is never truthy
		}
		if IsTrueLiteralType(t) {
			return true // `true` literal is always truthy
		}
		return true // general `boolean` is possibly truthy
	}
	if flags&checker.TypeFlagsStringLiteral != 0 {
		if s, ok := t.AsLiteralType().Value().(string); ok {
			return s != ""
		}
		return true
	}
	if flags&checker.TypeFlagsString != 0 {
		return true
	}
	if flags&checker.TypeFlagsNumberLiteral != 0 {
		return !isNumberLiteralZeroOrNaN(t.AsLiteralType().Value())
	}
	if flags&checker.TypeFlagsNumber != 0 {
		return true
	}
	if flags&checker.TypeFlagsBigIntLiteral != 0 {
		return checker.ValueToString(t.AsLiteralType().Value()) != "0n"
	}
	if flags&checker.TypeFlagsBigInt != 0 {
		return true
	}
	if flags&checker.TypeFlagsEnumLiteral != 0 {
		return true
	}
	if flags&checker.TypeFlagsUnion != 0 {
		return IsPossiblyTruthy(t)
	}
	if flags&checker.TypeFlagsIntersection != 0 {
		return Every(t.Types(), isConstituentPossiblyTruthy)
	}
	return true
}
