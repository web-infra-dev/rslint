package estree

import (
	"context"
	"errors"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

type projectedToken struct {
	typeCode   uint8
	start, end uint32
}
type projectedComment struct{ kind, start, end uint32 }

// Frame source offsets use uint32. Keep this temporary index compact and
// pointer-free, within the encoding's range checked before visiting syntax.
type tokenCover struct {
	typeCode   uint8
	kind       ast.Kind
	start, end uint32
}

func (p *estreeProjector) recordTokenCover(node *ast.Node) {
	if cover, start, ok := coverForNode(p.file, node); ok {
		cover.start = uint32(start)
		p.tokenCovers = append(p.tokenCovers, cover)
	}
}

func scanTokensAndCommentsInto(
	ctx context.Context,
	scan *scanner.Scanner,
	file *ast.SourceFile,
	offsets utf16Offsets,
	covers []tokenCover,
	tokens []projectedToken,
	comments []projectedComment,
	budget int,
) ([]projectedToken, []projectedComment) {
	// Projection order can differ from source order. Stable sorting preserves
	// the last recorded cover at a repeated position, like index assignment.
	slices.SortStableFunc(covers, func(a, b tokenCover) int {
		return int(a.start) - int(b.start)
	})
	coverIndex := 0
	isTypeScript := file.ScriptKind == core.ScriptKindTS || file.ScriptKind == core.ScriptKindTSX
	scan.SetText(file.Text())
	scan.SetLanguageVariant(file.LanguageVariant)
	scan.SetSkipTrivia(false)
	if comments == nil {
		comments = make([]projectedComment, 0, 8)
	}
	for kind := scan.Scan(); kind != ast.KindEndOfFile; kind = scan.Scan() {
		if uint64(len(tokens))*9+uint64(len(comments))*12+18 > uint64(budget) {
			panic(errors.New("ESTree token columns do not fit frame encoding"))
		}
		if err := ctx.Err(); err != nil {
			panic(err)
		}
		start := scan.TokenStart()
		end := scan.TokenEnd()
		for coverIndex < len(covers) && covers[coverIndex].start <= uint32(start) {
			coverIndex++
		}
		var cover tokenCover
		covered := coverIndex > 0 && covers[coverIndex-1].start == uint32(start)
		if covered {
			cover = covers[coverIndex-1]
		}
		coverEnd := int(cover.end)
		// The syntax tree resolves contexts a plain scanner cannot: compound
		// '>' operators versus generic closers, and JSX strings without JS escapes.
		if covered && (isReplacementKind(cover.kind) || cover.typeCode == tokenPunctuator ||
			cover.typeCode == tokenJSXTextRaw ||
			(cover.kind == ast.KindIdentifier && coverEnd > end)) {
			kind = cover.kind
			end = coverEnd
			scan.ResetTokenState(end)
		}
		utf16Start := offsets.At(start)
		utf16End := offsets.At(end)
		if kind == ast.KindSingleLineCommentTrivia || kind == ast.KindMultiLineCommentTrivia {
			commentKind := uint32(0)
			if kind == ast.KindMultiLineCommentTrivia {
				commentKind = 1
			}
			comments = append(comments, projectedComment{kind: commentKind, start: utf16Start, end: utf16End})
			continue
		}
		if kind >= ast.KindFirstTriviaToken && kind <= ast.KindLastTriviaToken {
			continue
		}
		if kind == ast.KindLessThanSlashToken && end == start+2 {
			tokens = append(tokens,
				projectedToken{typeCode: tokenPunctuator, start: utf16Start, end: offsets.At(start + 1)},
				projectedToken{typeCode: tokenPunctuator, start: offsets.At(start + 1), end: utf16End},
			)
			continue
		}
		typeCode := classifyTokenKind(kind, isTypeScript)
		if covered && end <= coverEnd {
			jsKeywordException := !isTypeScript && cover.typeCode == tokenIdentifier &&
				(kind == ast.KindStaticKeyword || kind == ast.KindLetKeyword || kind == ast.KindYieldKeyword)
			if !jsKeywordException {
				typeCode = cover.typeCode
			}
		}
		tokens = append(tokens, projectedToken{typeCode: typeCode, start: utf16Start, end: utf16End})
	}
	return tokens, comments
}

func coverForNode(file *ast.SourceFile, node *ast.Node) (tokenCover, int, bool) {
	start := node.Pos()
	cover := tokenCover{kind: node.Kind, end: uint32(node.End())}
	switch node.Kind {
	case ast.KindThisKeyword:
		if (file.ScriptKind != core.ScriptKindJS && file.ScriptKind != core.ScriptKindJSX) || !isJSXIdentifier(node) {
			return tokenCover{}, 0, false
		}
		start = scanner.GetTokenPosOfNode(node, file, false)
		cover.typeCode = tokenJSXIdentifier
	case ast.KindGreaterThanEqualsToken, ast.KindGreaterThanGreaterThanToken,
		ast.KindGreaterThanGreaterThanGreaterThanToken, ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken:
		start = scanner.GetTokenPosOfNode(node, file, false)
		cover.typeCode = tokenPunctuator
	case ast.KindIdentifier:
		start = scanner.GetTokenPosOfNode(node, file, false)
		cover.typeCode = tokenIdentifier
		if isJSXIdentifier(node) {
			cover.typeCode = tokenJSXIdentifier
			if file.ScriptKind == core.ScriptKindTSX && node.Parent != nil && node.Parent.Kind == ast.KindJsxNamespacedName {
				cover.typeCode = tokenIdentifier
			}
		}
	case ast.KindPrivateIdentifier:
		start = scanner.GetTokenPosOfNode(node, file, false)
		cover.typeCode = tokenPrivateIdentifier
	case ast.KindStringLiteral:
		start = scanner.GetTokenPosOfNode(node, file, false)
		cover.typeCode = tokenString
		if node.Parent != nil && node.Parent.Kind == ast.KindJsxAttribute {
			cover.typeCode = tokenJSXTextRaw
		}
	case ast.KindNumericLiteral, ast.KindBigIntLiteral:
		start = scanner.GetTokenPosOfNode(node, file, false)
		cover.typeCode = tokenNumeric
	case ast.KindRegularExpressionLiteral:
		start = scanner.GetTokenPosOfNode(node, file, false)
		cover.typeCode = tokenRegexp
	case ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateHead:
		start = scanner.GetTokenPosOfNode(node, file, false)
		cover.typeCode = tokenTemplate
	case ast.KindTemplateMiddle, ast.KindTemplateTail:
		start = scanner.GetTokenPosOfNode(node, file, false)
		cover.typeCode = tokenTemplate
	case ast.KindJsxText, ast.KindJsxTextAllWhiteSpaces:
		cover.typeCode = tokenJSXText
		if file.ScriptKind == core.ScriptKindTSX {
			cover.typeCode = tokenJSXTextRaw
		}
	default:
		return tokenCover{}, 0, false
	}
	return cover, start, true
}

func isJSXIdentifier(node *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindPropertyAccessExpression:
			continue
		case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxOpeningElement,
			ast.KindJsxClosingElement, ast.KindJsxAttribute, ast.KindJsxNamespacedName:
			return true
		default:
			return false
		}
	}
	return false
}

func isReplacementKind(kind ast.Kind) bool {
	return kind == ast.KindRegularExpressionLiteral ||
		kind == ast.KindNoSubstitutionTemplateLiteral ||
		kind == ast.KindTemplateHead ||
		kind == ast.KindTemplateMiddle ||
		kind == ast.KindTemplateTail ||
		kind == ast.KindJsxText ||
		kind == ast.KindJsxTextAllWhiteSpaces
}

const (
	tokenIdentifier uint8 = iota
	tokenKeyword
	tokenPunctuator
	tokenString
	tokenNumeric
	tokenRegexp
	tokenTemplate
	tokenBoolean
	tokenNull
	tokenPrivateIdentifier
	tokenJSXIdentifier
	tokenJSXText
	tokenJSXTextRaw
)

func classifyTokenKind(kind ast.Kind, isTypeScript bool) uint8 {
	switch kind {
	case ast.KindIdentifier, ast.KindAwaitKeyword:
		return tokenIdentifier
	case ast.KindPrivateIdentifier:
		return tokenPrivateIdentifier
	case ast.KindStringLiteral:
		return tokenString
	case ast.KindNumericLiteral, ast.KindBigIntLiteral:
		return tokenNumeric
	case ast.KindRegularExpressionLiteral:
		return tokenRegexp
	case ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail:
		return tokenTemplate
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return tokenBoolean
	case ast.KindNullKeyword:
		return tokenNull
	case ast.KindJsxText, ast.KindJsxTextAllWhiteSpaces:
		return tokenJSXText
	case ast.KindEnumKeyword:
		if isTypeScript {
			return tokenKeyword
		}
		return tokenIdentifier
	}
	if isContextualKeyword(kind) {
		return tokenIdentifier
	}
	if isFutureReservedKeyword(kind) {
		if isTypeScript {
			return tokenKeyword
		}
		return tokenIdentifier
	}
	if kind >= ast.KindFirstKeyword && kind <= ast.KindLastKeyword {
		return tokenKeyword
	}
	return tokenPunctuator
}

func isContextualKeyword(kind ast.Kind) bool {
	values := [...]ast.Kind{
		ast.KindAbstractKeyword, ast.KindAccessorKeyword, ast.KindAsKeyword,
		ast.KindAssertsKeyword, ast.KindAssertKeyword, ast.KindAnyKeyword,
		ast.KindAsyncKeyword, ast.KindAwaitKeyword, ast.KindBooleanKeyword,
		ast.KindConstructorKeyword, ast.KindDeclareKeyword, ast.KindGetKeyword,
		ast.KindImmediateKeyword, ast.KindInferKeyword, ast.KindIntrinsicKeyword,
		ast.KindIsKeyword, ast.KindKeyOfKeyword, ast.KindModuleKeyword,
		ast.KindNamespaceKeyword, ast.KindNeverKeyword, ast.KindOutKeyword,
		ast.KindReadonlyKeyword, ast.KindRequireKeyword, ast.KindNumberKeyword,
		ast.KindObjectKeyword, ast.KindSatisfiesKeyword, ast.KindSetKeyword,
		ast.KindStringKeyword, ast.KindSymbolKeyword, ast.KindTypeKeyword,
		ast.KindUndefinedKeyword, ast.KindUniqueKeyword, ast.KindUnknownKeyword,
		ast.KindUsingKeyword, ast.KindFromKeyword, ast.KindGlobalKeyword,
		ast.KindBigIntKeyword, ast.KindOverrideKeyword, ast.KindOfKeyword,
		ast.KindDeferKeyword,
	}
	for _, value := range values {
		if kind == value {
			return true
		}
	}
	return false
}

func isFutureReservedKeyword(kind ast.Kind) bool {
	return kind == ast.KindImplementsKeyword || kind == ast.KindInterfaceKeyword ||
		kind == ast.KindPackageKeyword || kind == ast.KindPrivateKeyword ||
		kind == ast.KindProtectedKeyword || kind == ast.KindPublicKeyword
}
