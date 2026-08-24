package sort_imports

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed sort_imports.schema.json
var schemaJSON []byte

type options struct {
	ignoreCase            bool
	ignoreDeclarationSort bool
	ignoreMemberSort      bool
	allowSeparatedGroups  bool
	memberSyntaxSortOrder []string
}

func parseOptions(raw []any) options {
	opts := options{memberSyntaxSortOrder: []string{"none", "all", "multiple", "single"}}
	if len(raw) == 0 {
		return opts
	}
	m, _ := raw[0].(map[string]any)
	if value, ok := m["ignoreCase"].(bool); ok {
		opts.ignoreCase = value
	}
	if value, ok := m["ignoreDeclarationSort"].(bool); ok {
		opts.ignoreDeclarationSort = value
	}
	if value, ok := m["ignoreMemberSort"].(bool); ok {
		opts.ignoreMemberSort = value
	}
	if value, ok := m["allowSeparatedGroups"].(bool); ok {
		opts.allowSeparatedGroups = value
	}
	if values, ok := m["memberSyntaxSortOrder"].([]any); ok && len(values) == 4 {
		order := make([]string, 0, 4)
		for _, value := range values {
			name, ok := value.(string)
			if !ok {
				return opts
			}
			order = append(order, name)
		}
		opts.memberSyntaxSortOrder = order
	}
	return opts
}

type importMember struct {
	node *ast.Node
	name string
	kind string
}

func importMembers(node *ast.Node) []importMember {
	declaration := node.AsImportDeclaration()
	if declaration == nil || declaration.ImportClause == nil {
		return nil
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause == nil {
		return nil
	}
	members := make([]importMember, 0, 3)
	if name := clause.Name(); name != nil {
		members = append(members, importMember{node: declaration.ImportClause, name: name.AsIdentifier().Text, kind: "default"})
	}
	if clause.NamedBindings == nil {
		return members
	}
	switch clause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		name := clause.NamedBindings.AsNamespaceImport().Name()
		members = append(members, importMember{node: clause.NamedBindings, name: name.AsIdentifier().Text, kind: "namespace"})
	case ast.KindNamedImports:
		named := clause.NamedBindings.AsNamedImports()
		if named != nil && named.Elements != nil {
			for _, element := range named.Elements.Nodes {
				specifier := element.AsImportSpecifier()
				if specifier == nil || specifier.Name() == nil {
					continue
				}
				members = append(members, importMember{node: element, name: specifier.Name().AsIdentifier().Text, kind: "named"})
			}
		}
	}
	return members
}

func usedMemberSyntax(members []importMember) string {
	if len(members) == 0 {
		return "none"
	}
	if members[0].kind == "namespace" {
		return "all"
	}
	if len(members) == 1 {
		return "single"
	}
	return "multiple"
}

func sortableName(name string, ignoreCase bool) string {
	if ignoreCase {
		return ecmascript.StringToLowerCase(name)
	}
	return name
}

func syntaxIndex(order []string, syntax string) int {
	return slices.Index(order, syntax)
}

func namedImportMembers(node *ast.Node) ([]importMember, *ast.Node) {
	declaration := node.AsImportDeclaration()
	if declaration == nil || declaration.ImportClause == nil {
		return nil, nil
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause == nil || clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
		return nil, nil
	}
	namedNode := clause.NamedBindings
	named := namedNode.AsNamedImports()
	if named == nil || named.Elements == nil {
		return nil, namedNode
	}
	members := make([]importMember, 0, len(named.Elements.Nodes))
	for _, element := range named.Elements.Nodes {
		specifier := element.AsImportSpecifier()
		if specifier == nil || specifier.Name() == nil {
			continue
		}
		members = append(members, importMember{node: element, name: specifier.Name().AsIdentifier().Text, kind: "named"})
	}
	return members, namedNode
}

func buildMemberFix(ctx rule.RuleContext, members []importMember, namedNode *ast.Node, ignoreCase bool) []rule.RuleFix {
	if len(members) < 2 || namedNode == nil {
		return nil
	}
	namedRange := utils.TrimNodeTextRange(ctx.SourceFile, namedNode)
	if utils.HasCommentInSpan(ctx.Comments.All(), namedRange.Pos(), namedRange.End()) {
		return nil
	}
	ranges := make([]core.TextRange, len(members))
	for i, member := range members {
		ranges[i] = utils.TrimNodeTextRange(ctx.SourceFile, member.node)
	}
	sorted := slices.Clone(members)
	slices.SortStableFunc(sorted, func(a, b importMember) int {
		comparison := ecmascript.CompareStrings(sortableName(a.name, ignoreCase), sortableName(b.name, ignoreCase))
		if comparison != 0 {
			return comparison
		}
		// Upstream's comparator returns -1 rather than 0 for equal sortable
		// names. V8 consequently places the later case-equivalent specifier
		// first whenever some other inversion causes the fix to run.
		if a.node.Pos() > b.node.Pos() {
			return -1
		}
		return 1
	})
	text := ctx.SourceFile.Text()
	var replacement strings.Builder
	replacement.Grow(ranges[len(ranges)-1].End() - ranges[0].Pos())
	for i, member := range sorted {
		r := utils.TrimNodeTextRange(ctx.SourceFile, member.node)
		replacement.WriteString(text[r.Pos():r.End()])
		if i+1 < len(ranges) {
			replacement.WriteString(text[ranges[i].End():ranges[i+1].Pos()])
		}
	}
	return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(ranges[0].Pos(), ranges[len(ranges)-1].End()), replacement.String())}
}

// SortImportsRule enforces ordering between import declarations and within
// named import member lists.
// https://eslint.org/docs/latest/rules/sort-imports
var SortImportsRule = rule.Rule{
	Name:   "sort-imports",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		var previousDeclaration *ast.Node

		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				members := importMembers(node)
				if !opts.ignoreDeclarationSort {
					if previousDeclaration != nil && opts.allowSeparatedGroups {
						lineMap := ctx.SourceFile.ECMALineMap()
						previousEndLine := scanner.ComputeLineOfPosition(lineMap, previousDeclaration.End())
						currentStartLine := scanner.ComputeLineOfPosition(lineMap, utils.TrimNodeTextRange(ctx.SourceFile, node).Pos())
						if currentStartLine-previousEndLine > 1 {
							previousDeclaration = nil
						}
					}
					if previousDeclaration != nil {
						previousMembers := importMembers(previousDeclaration)
						currentSyntax := usedMemberSyntax(members)
						previousSyntax := usedMemberSyntax(previousMembers)
						currentIndex := syntaxIndex(opts.memberSyntaxSortOrder, currentSyntax)
						previousIndex := syntaxIndex(opts.memberSyntaxSortOrder, previousSyntax)
						if currentIndex != previousIndex {
							if currentIndex < previousIndex {
								ctx.ReportNode(node, rule.RuleMessage{Id: "unexpectedSyntaxOrder", Description: fmt.Sprintf("Expected '%s' syntax before '%s' syntax.", currentSyntax, previousSyntax)})
							}
						} else if len(previousMembers) > 0 && len(members) > 0 && ecmascript.CompareStrings(sortableName(members[0].name, opts.ignoreCase), sortableName(previousMembers[0].name, opts.ignoreCase)) < 0 {
							ctx.ReportNode(node, rule.RuleMessage{Id: "sortImportsAlphabetically", Description: "Imports should be sorted alphabetically."})
						}
					}
					previousDeclaration = node
				}

				if opts.ignoreMemberSort {
					return
				}
				namedMembers, namedNode := namedImportMembers(node)
				firstUnsorted := -1
				for i := 1; i < len(namedMembers); i++ {
					if ecmascript.CompareStrings(sortableName(namedMembers[i-1].name, opts.ignoreCase), sortableName(namedMembers[i].name, opts.ignoreCase)) > 0 {
						firstUnsorted = i
						break
					}
				}
				if firstUnsorted == -1 {
					return
				}
				member := namedMembers[firstUnsorted]
				ctx.ReportNodeWithDeferredFixes(member.node, rule.RuleMessage{Id: "sortMembersAlphabetically", Description: fmt.Sprintf("Member '%s' of the import declaration should be sorted alphabetically.", member.name)}, func() []rule.RuleFix {
					return buildMemberFix(ctx, namedMembers, namedNode, opts.ignoreCase)
				})
			},
		}
	},
}
