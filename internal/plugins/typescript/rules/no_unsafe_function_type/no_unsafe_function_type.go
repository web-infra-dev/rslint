package no_unsafe_function_type

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildBannedFunctionTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "bannedFunctionType",
		Description: "The `Function` type accepts any function-like value.\nPrefer explicitly defining any function parameters and return type.",
	}
}

var NoUnsafeFunctionTypeRule = rule.CreateRule(rule.Rule{
	Name:             "no-unsafe-function-type",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkBannedType := func(ident *ast.Node) {
			if ident == nil || ident.Kind != ast.KindIdentifier {
				return
			}
			if ident.AsIdentifier().Text != "Function" {
				return
			}
			if !isReferenceToGlobalFunction(ctx, ident) {
				return
			}
			ctx.ReportNode(ident, buildBannedFunctionTypeMessage())
		}

		return rule.RuleListeners{
			ast.KindTypeReference: func(node *ast.Node) {
				ref := node.AsTypeReferenceNode()
				if ref == nil {
					return
				}
				checkBannedType(ref.TypeName)
			},
			ast.KindExpressionWithTypeArguments: func(node *ast.Node) {
				if !isInRelevantHeritageClause(node) {
					return
				}
				expr := node.AsExpressionWithTypeArguments()
				if expr == nil {
					return
				}
				checkBannedType(expr.Expression)
			},
		}
	},
})

// isInRelevantHeritageClause reports whether the ExpressionWithTypeArguments
// sits in a heritage position that the upstream rule listens to:
//   - `class X implements ...` (matches upstream `TSClassImplements`)
//   - `interface X extends ...` (matches upstream `TSInterfaceHeritage`)
//
// `class X extends ...` is intentionally excluded: upstream does not register
// a listener for it, so a `Function` reference in that position is left alone.
// Non-heritage uses of ExpressionWithTypeArguments (e.g. JSDoc type contexts)
// are also rejected because upstream's listener set never matches them.
func isInRelevantHeritageClause(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || !ast.IsHeritageClause(parent) {
		return false
	}
	grand := parent.Parent
	switch {
	case grand == nil:
		return false
	case ast.IsInterfaceDeclaration(grand):
		return parent.AsHeritageClause().Token == ast.KindExtendsKeyword
	case ast.IsClassLike(grand):
		return parent.AsHeritageClause().Token == ast.KindImplementsKeyword
	}
	return false
}

// isReferenceToGlobalFunction mirrors upstream's `isReferenceToGlobalFunction`:
// the identifier counts as referring to the global `Function` only when no
// user-source declaration in the *current file* provides it. We approximate
// ESLint's `!ref?.resolved?.defs.length` via the tsgo checker: if the resolved
// symbol has any declaration whose source file is the same as the file being
// linted, the reference is treated as locally bound; otherwise — including the
// unresolved case — it counts as the lib.d.ts-provided global.
//
// Restricting the match to the current source file matches ESLint scope
// manager semantics: a `declare global { interface Function {...} }` written
// in another file is ambient — it merges into the global symbol at the type
// checker level but does NOT create a `def` in the current file's scope
// manager. Accepting "any user-source declaration" (regardless of file) would
// be a false-negative because the upstream rule still fires in that case.
//
// Same-file declaration merging (e.g. an `interface Function` in this file
// that augments the lib type) and same-file imports (`import { Function }`)
// are both honored because their `Declarations` entries live in the file
// being linted.
func isReferenceToGlobalFunction(ctx rule.RuleContext, ident *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return true
	}
	sym := ctx.TypeChecker.GetSymbolAtLocation(ident)
	if sym == nil || len(sym.Declarations) == 0 {
		return true
	}
	for _, decl := range sym.Declarations {
		if decl == nil {
			continue
		}
		if ast.GetSourceFileOfNode(decl) == ctx.SourceFile {
			return false
		}
	}
	return true
}
