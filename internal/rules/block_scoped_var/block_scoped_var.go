package block_scoped_var

import (
	"fmt"
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/block-scoped-var
var BlockScopedVarRule = rule.Rule{
	Name:   "block-scoped-var",
	Schema: rule.EmptyArraySchema,
	Run:    run,
}

// varOccurrence is one `var`-declared binding identifier, paired with the
// block-scope range that was active around its declaration.
type varOccurrence struct {
	identifier *ast.Node
	symbol     *ast.Symbol
	scopeRange core.TextRange
}

func run(ctx rule.RuleContext, _ []any) rule.RuleListeners {
	if ctx.SourceFile == nil {
		return nil
	}

	var occurrences []varOccurrence
	declsBySymbol := make(map[*ast.Symbol][]*ast.Node)

	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node.Kind == ast.KindVariableDeclarationList && utils.IsVarKeyword(node) {
			scopeRange := blockScopeRange(ctx.SourceFile, node)
			// A for-in/for-of head has no `=` of its own, but the loop
			// implicitly assigns into the binding on every iteration, which
			// ESLint's scope analysis treats the same as an explicit
			// initializer for this rule's purposes.
			isForInOrForOf := node.Parent != nil &&
				(node.Parent.Kind == ast.KindForInStatement || node.Parent.Kind == ast.KindForOfStatement)
			utils.ForEachVariableDeclarationBinding(node, func(declaration *ast.Node, identifier *ast.Node, _ string) {
				symbol := identifier.Parent.Symbol()
				if symbol == nil {
					return
				}
				occurrences = append(occurrences, varOccurrence{
					identifier: identifier,
					symbol:     symbol,
					scopeRange: scopeRange,
				})
				// ESLint's eslint-scope only creates a self-reference for a
				// declarator that is assigned a value — a bare `var a;` never
				// becomes a "reference," so it can never flag (or be flagged
				// by) a sibling `var` declaration of the same name.
				if isForInOrForOf || hasInitializer(declaration) {
					declsBySymbol[symbol] = append(declsBySymbol[symbol], identifier)
				}
			})
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(ctx.SourceFile.AsNode())

	var diagnostics []diagnostic
	for _, occurrence := range occurrences {
		diagnostics = append(diagnostics, checkOccurrence(ctx, occurrence, declsBySymbol[occurrence.symbol])...)
	}

	// ESLint sorts all of a rule's reports by position before emitting them,
	// so two sibling `var` declarations that flag each other interleave by
	// report location rather than by declaration-processing order.
	sort.SliceStable(diagnostics, func(i, j int) bool {
		return diagnostics[i].textRange.Pos() < diagnostics[j].textRange.Pos()
	})
	for _, d := range diagnostics {
		ctx.ReportRange(d.textRange, d.message)
	}

	return nil
}

type diagnostic struct {
	textRange core.TextRange
	message   rule.RuleMessage
}

// checkOccurrence collects a diagnostic for every use of occurrence's
// declared variable that falls outside the block scope that was active when
// it was declared. "Uses" include both ordinary reads/writes and the binding
// identifiers of any sibling `var` declarations of the same variable that
// have their own initializer (or are a for-in/for-of loop's binding)
// elsewhere in the file — ESLint's scope analysis treats such a declarator's
// own identifier as a reference too, which is what lets two sibling `var`
// declarations of the same name flag each other.
func checkOccurrence(ctx rule.RuleContext, occurrence varOccurrence, siblingDeclarations []*ast.Node) []diagnostic {
	uses := append([]*ast.Node{}, ctx.Refs.References(occurrence.symbol)...)
	uses = append(uses, siblingDeclarations...)
	sort.Slice(uses, func(i, j int) bool {
		return uses[i].Pos() < uses[j].Pos()
	})

	name := occurrence.identifier.Text()
	defRange := utils.TrimNodeTextRange(ctx.SourceFile, occurrence.identifier)
	defLineIndex, defCharIndex := scanner.GetECMALineAndUTF16CharacterOfPosition(ctx.SourceFile, defRange.Pos())
	defLine, defColumn := defLineIndex+1, int(defCharIndex)+1

	var diagnostics []diagnostic
	for _, use := range uses {
		useRange := utils.TrimNodeTextRange(ctx.SourceFile, use)
		if useRange.Pos() < occurrence.scopeRange.Pos() || useRange.End() > occurrence.scopeRange.End() {
			diagnostics = append(diagnostics, diagnostic{
				textRange: useRange,
				message: rule.RuleMessage{
					Id: "outOfScope",
					Description: fmt.Sprintf(
						"'%s' declared on line %d column %d is used outside of binding context.",
						name, defLine, defColumn,
					),
				},
			})
		}
	}
	return diagnostics
}

// hasInitializer reports whether a VariableDeclaration node has an `=`
// initializer.
func hasInitializer(declaration *ast.Node) bool {
	variableDeclaration := declaration.AsVariableDeclaration()
	return variableDeclaration != nil && variableDeclaration.Initializer != nil
}

// blockScopeRange returns the range of the nearest enclosing block scope for
// a `var` declaration list: the innermost Block, for-loop, or switch
// statement containing it, or the whole source file when there is none.
//
// This mirrors ESLint's stack of BlockStatement/ForStatement/ForInStatement/
// ForOfStatement/SwitchStatement/CatchClause/StaticBlock ranges without
// maintaining an explicit traversal stack: the range active at any given
// `var` declaration is always exactly its nearest enclosing node of one of
// those kinds, since scope entry/exit is properly nested. CatchClause and
// StaticBlock are omitted because a `var` inside either is always nested
// inside that construct's own Block first, which this function reaches
// first and which spans the same effective set of positions.
func blockScopeRange(sourceFile *ast.SourceFile, declarationList *ast.Node) core.TextRange {
	for current := declarationList.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindBlock,
			ast.KindForStatement,
			ast.KindForInStatement,
			ast.KindForOfStatement,
			ast.KindSwitchStatement:
			return utils.TrimNodeTextRange(sourceFile, current)
		case ast.KindSourceFile:
			return core.NewTextRange(0, current.End())
		}
	}
	return core.NewTextRange(0, sourceFile.AsNode().End())
}
