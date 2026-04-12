package first

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// See: https://github.com/import-js/eslint-plugin-import/blob/main/src/rules/first.js
var FirstRule = rule.Rule{
	Name: "import/first",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// The linter visits SourceFile's children but never fires a KindSourceFile
		// listener, so run the check eagerly before returning listeners.
		checkFirst(ctx, options)
		return rule.RuleListeners{}
	},
}

// getImportValue returns the module specifier string for an import node.
// For ImportDeclaration: the string literal in `from '...'`.
// For ImportEqualsDeclaration with external module: the string in `require('...')`.
func getImportValue(node *ast.Node) string {
	switch node.Kind {
	case ast.KindImportDeclaration:
		spec := node.AsImportDeclaration().ModuleSpecifier
		if spec != nil && spec.Kind == ast.KindStringLiteral {
			return spec.AsStringLiteral().Text
		}
	case ast.KindImportEqualsDeclaration:
		if ast.IsExternalModuleImportEqualsDeclaration(node) {
			expr := ast.GetExternalModuleImportEqualsDeclarationExpression(node)
			if expr.Kind == ast.KindStringLiteral {
				return expr.AsStringLiteral().Text
			}
		}
	}
	return ""
}

// hasReferenceBeforeImport checks whether any binding declared by the import
// node is referenced in a body statement that precedes the import.  This
// mirrors ESLint's shouldSort / getDeclaredVariables logic.
//
// When TypeChecker is available it uses symbol-based matching (accurate: skips
// type annotations, declaration names, and shadowed identifiers).  Otherwise
// it falls back to name-based matching (conservative).
func hasReferenceBeforeImport(ctx rule.RuleContext, body []*ast.Node, importBodyIndex int, importNode *ast.Node) bool {
	bindingNodes := utils.GetImportBindingNodes(importNode)
	if len(bindingNodes) == 0 {
		return false
	}

	importEnd := importNode.End()

	if ctx.TypeChecker != nil {
		return hasReferenceBeforeImportSymbol(ctx, body, importBodyIndex, importEnd, bindingNodes)
	}
	return hasReferenceBeforeImportName(body, importBodyIndex, bindingNodes)
}

// hasReferenceBeforeImportSymbol uses TypeChecker.GetSymbolAtLocation for
// precise symbol matching — only true value/expression references count.
func hasReferenceBeforeImportSymbol(ctx rule.RuleContext, body []*ast.Node, importBodyIndex int, importEnd int, bindingNodes []*ast.Node) bool {
	// Resolve symbols for all import bindings.
	symbolSet := make(map[*ast.Symbol]bool, len(bindingNodes))
	for _, bn := range bindingNodes {
		sym := ctx.TypeChecker.GetSymbolAtLocation(bn)
		if sym != nil {
			symbolSet[sym] = true
			if resolved := ctx.TypeChecker.SkipAlias(sym); resolved != nil && resolved != sym {
				symbolSet[resolved] = true
			}
		}
	}
	if len(symbolSet) == 0 {
		return false
	}

	found := false
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || found {
			return
		}
		// Only check non-declaration identifier references.
		if ast.IsIdentifier(node) && !utils.IsDeclarationIdentifier(node) {
			sym := ctx.TypeChecker.GetSymbolAtLocation(node)
			if sym != nil && symbolSet[sym] {
				// Position check: reference must appear before the import's end,
				// matching ESLint's `reference.identifier.range[0] < node.range[1]`.
				if node.Pos() < importEnd {
					found = true
					return
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found // stop early if found
		})
	}

	for j := range importBodyIndex {
		walk(body[j])
		if found {
			return true
		}
	}
	return false
}

// hasReferenceBeforeImportName is the fallback when TypeChecker is unavailable.
// It matches by identifier text — conservative (may suppress fix when ESLint
// would not, e.g. for type annotations or shadowed names).
func hasReferenceBeforeImportName(body []*ast.Node, importBodyIndex int, bindingNodes []*ast.Node) bool {
	nameSet := make(map[string]bool, len(bindingNodes))
	for _, bn := range bindingNodes {
		nameSet[bn.Text()] = true
	}

	var contains func(*ast.Node) bool
	contains = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if ast.IsIdentifier(node) && !utils.IsDeclarationIdentifier(node) && nameSet[node.AsIdentifier().Text] {
			return true
		}
		result := false
		node.ForEachChild(func(child *ast.Node) bool {
			if contains(child) {
				result = true
				return true
			}
			return false
		})
		return result
	}

	for j := range importBodyIndex {
		if contains(body[j]) {
			return true
		}
	}
	return false
}

func messageFirst() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "first",
		Description: "Import in body of module; reorder to top.",
	}
}

func messageAbsolute() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "absolute",
		Description: "Absolute imports should come before relative imports.",
	}
}

// errorInfo tracks a misplaced import and the source range to extract when
// building the autofix (from the end of the previous statement to the end of
// this import).
type errorInfo struct {
	node      *ast.Node
	rangeFrom int
	rangeTo   int
}

func checkFirst(ctx rule.RuleContext, options any) {
	statements := ctx.SourceFile.Statements
	if statements == nil || len(statements.Nodes) == 0 {
		return
	}

	absoluteFirst := utils.GetOptionsString(options) == "absolute-first"
	body := statements.Nodes
	sourceText := ctx.SourceFile.Text()

	nonImportCount := 0
	anyExpressions := false
	anyRelative := false
	var lastLegalImp *ast.Node
	var errorInfos []errorInfo
	shouldSort := true
	lastSortNodesIndex := 0

	for i, node := range body {
		// Skip directives ('use strict', etc.) that precede any real expression.
		if !anyExpressions && ast.IsPrologueDirective(node) {
			continue
		}
		anyExpressions = true

		if ast.IsImportOrImportEqualsDeclaration(node) {
			// absolute-first: report absolute imports that follow a relative import.
			if absoluteFirst {
				value := getImportValue(node)
				if strings.HasPrefix(value, ".") {
					anyRelative = true
				} else if anyRelative {
					// Report on the specifier/reference, not the whole statement.
					var reportNode *ast.Node
					if node.Kind == ast.KindImportDeclaration {
						reportNode = node.AsImportDeclaration().ModuleSpecifier
					} else {
						reportNode = node.AsImportEqualsDeclaration().ModuleReference
					}
					ctx.ReportNode(reportNode, messageAbsolute())
				}
			}

			if nonImportCount > 0 {
				// This import appears after non-import code.
				// Check if any of its declared names are referenced before it;
				// if so, moving the import could change evaluation order, so
				// disable autofix from this point on.
				if shouldSort {
					if hasReferenceBeforeImport(ctx, body, i, node) {
						shouldSort = false
					}
				}
				if shouldSort {
					lastSortNodesIndex = len(errorInfos)
				}

				rangeFrom := 0
				if i > 0 {
					rangeFrom = body[i-1].End()
				}

				errorInfos = append(errorInfos, errorInfo{
					node:      node,
					rangeFrom: rangeFrom,
					rangeTo:   node.End(),
				})
			} else {
				lastLegalImp = node
			}
		} else {
			nonImportCount++
		}
	}

	if len(errorInfos) == 0 {
		return
	}

	for i, ei := range errorInfos {
		if i == lastSortNodesIndex {
			// The last sortable error carries the combined fix that moves all
			// sortable imports to the top.
			sortNodes := errorInfos[:lastSortNodesIndex+1]
			fixes := buildFix(sourceText, body, lastLegalImp, sortNodes)
			ctx.ReportNodeWithFixes(ei.node, messageFirst(), fixes...)
		} else if i < lastSortNodesIndex {
			// Earlier sortable errors get a no-op fix so the fixer treats them
			// as already handled (avoids overlapping fix conflicts).
			ctx.ReportNodeWithFixes(ei.node, messageFirst(), rule.RuleFix{
				Range: core.NewTextRange(ei.node.End(), ei.node.End()),
				Text:  "",
			})
		} else {
			ctx.ReportNode(ei.node, messageFirst())
		}
	}
}

// buildFix creates a single RuleFix that moves all sortable imports to just
// after lastLegalImp (or to the very beginning of the file when there is no
// legal import).  It mirrors ESLint's approach of combining insert + remove
// operations into one contiguous range replacement.
func buildFix(sourceText string, body []*ast.Node, lastLegalImp *ast.Node, sortNodes []errorInfo) []rule.RuleFix {
	// Collect the source text for each misplaced import (including the
	// whitespace between it and the previous statement).
	var insertParts []string
	for _, sn := range sortNodes {
		nodeText := sourceText[sn.rangeFrom:sn.rangeTo]
		// If the extracted text starts with a non-whitespace character (e.g.
		// the import immediately follows a `}` with no line break), prepend a
		// newline so the moved import appears on its own line.
		if r, _ := utf8.DecodeRuneInString(nodeText); r != utf8.RuneError && !utils.IsStrWhiteSpace(r) {
			nodeText = "\n" + nodeText
		}
		insertParts = append(insertParts, nodeText)
	}
	insertSourceCode := strings.Join(insertParts, "")

	if lastLegalImp == nil {
		// No preceding legal import: place the imports at the very top and
		// preserve the original leading whitespace pattern after them.
		trimmed := strings.TrimSpace(insertSourceCode)
		leadingWSEnd := strings.IndexFunc(insertSourceCode, func(r rune) bool {
			return !utils.IsStrWhiteSpace(r)
		})
		if leadingWSEnd < 0 {
			leadingWSEnd = len(insertSourceCode)
		}
		insertSourceCode = trimmed + insertSourceCode[:leadingWSEnd]
	}

	// Combine all operations (one insert + N removes) into a single text
	// replacement covering [0, lastRemoveEnd).
	type fixer struct {
		rangeStart int
		rangeEnd   int
		text       string
	}

	var fixers []fixer

	if lastLegalImp != nil {
		fixers = append(fixers, fixer{
			rangeStart: lastLegalImp.End(),
			rangeEnd:   lastLegalImp.End(),
			text:       insertSourceCode,
		})
	} else {
		pos := body[0].Pos()
		fixers = append(fixers, fixer{
			rangeStart: pos,
			rangeEnd:   pos,
			text:       insertSourceCode,
		})
	}

	for _, sn := range sortNodes {
		fixers = append(fixers, fixer{
			rangeStart: sn.rangeFrom,
			rangeEnd:   sn.rangeTo,
			text:       "",
		})
	}

	slices.SortFunc(fixers, func(a, b fixer) int {
		if a.rangeStart != b.rangeStart {
			return a.rangeStart - b.rangeStart
		}
		return a.rangeEnd - b.rangeEnd
	})

	var builder strings.Builder
	lastEnd := 0
	overallEnd := 0
	for _, f := range fixers {
		builder.WriteString(sourceText[lastEnd:f.rangeStart])
		builder.WriteString(f.text)
		lastEnd = f.rangeEnd
		if f.rangeEnd > overallEnd {
			overallEnd = f.rangeEnd
		}
	}

	return []rule.RuleFix{
		{
			Range: core.NewTextRange(0, overallEnd),
			Text:  builder.String(),
		},
	}
}
