package unicornutil

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

type closureTypeToken struct {
	kind       ast.Kind
	start, end int
	lineBreak  bool
}

type closureTypeScope struct {
	closingToken      ast.Kind
	closureParameters bool
	nextParameter     bool
	recordKey         bool
	expectsType       bool
	callableReturns   int
}

// normalizeClosureFunctionTypes adapts function(T): R to { (argument: T): R }.
// tsgo still parses the types, wrappers and parameter metadata. Each callable's
// return type extends to its containing delimiter, including unions and nested
// signatures. A callable without a return annotation ends after its parameters.
func normalizeClosureFunctionTypes(raw string) (string, bool) {
	scan := scanner.NewScanner()
	scan.SetText(raw)
	scan.SetSkipTrivia(true)
	scan.SetSkipJSDocLeadingAsterisks(true)
	var tokens []closureTypeToken
	for kind := scan.Scan(); kind != ast.KindEndOfFile; kind = scan.Scan() {
		tokens = append(tokens, closureTypeToken{
			kind: kind, start: scan.TokenStart(), end: scan.TokenEnd(), lineBreak: scan.HasPrecedingLineBreak(),
		})
	}

	var out strings.Builder
	scopes := []closureTypeScope{{closingToken: ast.KindEndOfFile}}
	position := 0
	changed := false
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		next := ast.KindEndOfFile
		if i+1 < len(tokens) {
			next = tokens[i+1].kind
		}
		scope := &scopes[len(scopes)-1]
		// A newline can end a record member without a comma or semicolon.
		if token.lineBreak && scope.closingToken == ast.KindCloseBraceToken && !scope.expectsType &&
			(token.kind == ast.KindIdentifier || ast.IsKeyword(token.kind) || ast.IsLiteralKind(token.kind) ||
				token.kind == ast.KindOpenParenToken || (token.kind == ast.KindOpenBracketToken && next != ast.KindCloseBracketToken)) {
			out.WriteString(strings.Repeat("}", scope.callableReturns))
			scope.callableReturns = 0
			scope.recordKey = true
		}
		out.WriteString(raw[position:token.start])
		position = token.end
		previous := ast.KindUnknown
		if i > 0 {
			previous = tokens[i-1].kind
		}
		switch token.kind {
		case ast.KindImportKeyword, ast.KindTypeOfKeyword, ast.KindExtendsKeyword, ast.KindNewKeyword,
			ast.KindKeyOfKeyword, ast.KindInferKeyword, ast.KindAssertsKeyword, ast.KindIsKeyword:
			// These TypeScript type operators are unsupported in upstream JSDoc.
			// Property names, qualified names and named arrow parameters still use them.
			if !scope.recordKey && previous != ast.KindDotToken && next != ast.KindColonToken {
				return "", false
			}
		}
		if scope.nextParameter && token.kind != scope.closingToken {
			scope.nextParameter = false
			switch {
			case token.kind == ast.KindDotDotDotToken:
				out.WriteString("...argument:")
				continue
			case token.kind == ast.KindThisKeyword:
				if next != ast.KindColonToken {
					return "", false
				}
			default:
				out.WriteString("argument:")
			}
		}
		if token.kind == ast.KindFunctionKeyword && next == ast.KindOpenParenToken &&
			!scope.recordKey && previous != ast.KindDotToken {
			out.WriteString("{(")
			scope.expectsType = false
			scopes = append(scopes, closureTypeScope{
				closingToken: ast.KindCloseParenToken, closureParameters: true, nextParameter: true,
			})
			i++
			position = tokens[i].end
			changed = true
			continue
		}
		switch token.kind {
		case ast.KindColonToken, ast.KindBarToken, ast.KindAmpersandToken,
			ast.KindEqualsGreaterThanToken, ast.KindDotToken, ast.KindReadonlyKeyword:
			scope.expectsType = true
		case ast.KindQuestionToken, ast.KindExclamationToken:
			// Prefix modifiers keep expecting a type; postfix modifiers complete it.
		default:
			scope.expectsType = false
		}

		closingToken := ast.KindUnknown
		switch token.kind {
		case ast.KindOpenParenToken:
			closingToken = ast.KindCloseParenToken
		case ast.KindOpenBracketToken:
			closingToken = ast.KindCloseBracketToken
		case ast.KindOpenBraceToken:
			closingToken = ast.KindCloseBraceToken
		case ast.KindLessThanToken:
			closingToken = ast.KindGreaterThanToken
		}
		if closingToken != ast.KindUnknown {
			scopes = append(scopes, closureTypeScope{
				closingToken: closingToken, recordKey: token.kind == ast.KindOpenBraceToken,
			})
		} else {
			switch token.kind {
			case ast.KindCloseParenToken, ast.KindCloseBracketToken, ast.KindCloseBraceToken, ast.KindGreaterThanToken:
				if len(scopes) == 1 || token.kind != scope.closingToken {
					return "", false
				}
				out.WriteString(strings.Repeat("}", scope.callableReturns))
				out.WriteString(raw[token.start:token.end])
				parameters := scope.closureParameters
				scopes = scopes[:len(scopes)-1]
				if parameters {
					if next == ast.KindColonToken {
						scopes[len(scopes)-1].callableReturns++
					} else {
						out.WriteString(": unknown}")
					}
				}
				continue
			case ast.KindCommaToken, ast.KindSemicolonToken, ast.KindEqualsToken:
				out.WriteString(strings.Repeat("}", scope.callableReturns))
				scope.callableReturns = 0
				if token.kind == ast.KindEqualsToken && scope.closureParameters {
					// Optional Closure parameters have no effect on the return type.
					continue
				}
				scope.recordKey = scope.closingToken == ast.KindCloseBraceToken
				scope.nextParameter = scope.closureParameters
			case ast.KindColonToken:
				scope.recordKey = false
			}
		}
		out.WriteString(raw[token.start:token.end])
	}
	if len(scopes) != 1 {
		return "", false
	}
	out.WriteString(strings.Repeat("}", scopes[0].callableReturns))
	out.WriteString(raw[position:])
	return out.String(), changed
}
