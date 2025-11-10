package utils

import (
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
func IsTrueLiteralType(typeChecker *checker.Checker, t *checker.Type) bool {
	if !IsBooleanLiteralType(t) {
		return false
	}
	// For boolean literals, check if it's specifically the 'true' type
	// by checking if it's an intrinsic type with name "true"
	if IsIntrinsicType(t) && t.AsIntrinsicType().IntrinsicName() == "true" {
		return true
	}
	// Fallback: use TypeToString to distinguish true from false
	// This works because true and false have different string representations
	typeStr := typeChecker.TypeToString(t)
	return typeStr == "true"
}
func IsFalseLiteralType(typeChecker *checker.Checker, t *checker.Type) bool {
	if !IsBooleanLiteralType(t) {
		return false
	}
	// For boolean literals, check if it's specifically the 'false' type
	// by checking if it's an intrinsic type with name "false"
	if IsIntrinsicType(t) && t.AsIntrinsicType().IntrinsicName() == "false" {
		return true
	}
	// Fallback: use TypeToString to distinguish true from false
	typeStr := typeChecker.TypeToString(t)
	return typeStr == "false"
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
			children = append(children, sourceFile.GetOrCreateToken(token, tokenFullStart, tokenEnd, node))
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
		children = append(children, sourceFile.GetOrCreateToken(token, tokenFullStart, tokenEnd, node))
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
