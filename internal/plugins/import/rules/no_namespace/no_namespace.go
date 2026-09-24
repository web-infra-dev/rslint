package no_namespace

import (
	_ "embed"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

//go:embed no_namespace.schema.json
var schemaJSON []byte

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-namespace.js
var NoNamespaceRule = rule.Rule{
	Name:   "import/no-namespace",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var patterns []any
		var ignored []*minimatch3.Matcher
		if len(options) > 0 {
			option, _ := options[0].(map[string]any)
			patterns, _ = option["ignore"].([]any)
			for _, pattern := range patterns {
				glob, _ := pattern.(string)
				ignored = append(ignored, minimatch3.New(glob, minimatch3.Options{MatchBase: true}))
			}
		}
		generated := map[string]bool{}
		return rule.RuleListeners{
			ast.KindNamespaceImport: func(node *ast.Node) {
				source := node.Parent.Parent.AsImportDeclaration().ModuleSpecifier.Text()
				for index, pattern := range ignored {
					if pattern.Match(source) {
						// Array.find stops at an empty pattern too, but its
						// empty-string result is falsy in the upstream condition.
						if patterns[index] != "" {
							return
						}
						break
					}
				}
				// Upstream reports a literal message, without a message ID.
				ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{Description: "Unexpected namespace import."}, func() []rule.RuleFix {
					return namespaceFixes(ctx, node, generated)
				})
			},
		}
	},
}

func namespaceFixes(ctx rule.RuleContext, node *ast.Node, generated map[string]bool) []rule.RuleFix {
	references := ctx.Refs.References(node.Symbol())
	if len(references) == 0 || utils.HasCommentInsideNode(ctx.SourceFile, node) {
		return nil
	}

	type memberUse struct {
		node *ast.Node
		name string
	}
	uses := make([]memberUse, 0, len(references))
	names := []string{}
	scopes := map[string]map[*scope.Scope]bool{}
	for _, reference := range references {
		member := utils.ESTreeParent(reference)
		object, property := utils.MemberExpressionParts(member)
		if utils.ESTreeRuntimeExpression(object) != reference || property == nil || utils.IsInJsxTagName(member) {
			return nil
		}
		// Named imports cannot be assignment or delete targets. Type assertions
		// and parentheses around the target do not make these fixes safe.
		target := member
		for target.Parent != nil && ast.IsOuterExpression(target.Parent, ast.OEKParentheses|ast.OEKAssertions) {
			target = target.Parent
		}
		if utils.IsWriteReference(member) || (target.Parent != nil && ast.IsDeleteExpression(target.Parent)) {
			return nil
		}
		property = utils.ESTreeRuntimeExpression(property)
		// Unlike upstream, never turn a dynamic key into a static import, or
		// emit an invalid binding for a non-identifier export name.
		if member.Kind == ast.KindElementAccessExpression {
			if property.Kind != ast.KindStringLiteral {
				return nil
			}
		} else if property.Kind != ast.KindIdentifier {
			return nil
		}
		name := property.Text()
		keyword := scanner.StringToToken(name)
		if !scanner.IsValidIdentifier(name) ||
			(keyword >= ast.KindFirstReservedWord && keyword <= ast.KindLastReservedWord) ||
			(keyword >= ast.KindFirstFutureReservedWord && keyword <= ast.KindLastFutureReservedWord) ||
			keyword == ast.KindAwaitKeyword || name == "arguments" || name == "eval" ||
			utils.HasCommentInsideNode(ctx.SourceFile, member) {
			return nil
		}
		if scopes[name] == nil {
			names = append(names, name)
			scopes[name] = map[*scope.Scope]bool{}
		}
		uses = append(uses, memberUse{member, name})
	}

	// Only build scopes once every reference can be fixed. Reuse their name
	// indexes instead of copying all declarations for each member access.
	manager := scopeanalysis.Declarations(ctx)
	for _, use := range uses {
		for current := manager.Acquire(use.node); current != nil; current = current.Parent {
			if scopes[use.name][current] {
				break
			}
			scopes[use.name][current] = true
		}
	}
	locals := make(map[string]string, len(names))
	specifiers := make([]string, 0, len(names))
	for _, name := range names {
		isConflict := func(candidate string) bool {
			if generated[candidate] || ctx.Globals.Access(candidate).IsDeclared() {
				return true
			}
			for current := range scopes[name] {
				if len(current.Declarations(candidate)) > 0 {
					return true
				}
			}
			return false
		}
		local := name
		if isConflict(local) {
			base := node.Name().Text() + "_" + name
			local = base
			for suffix := 1; isConflict(local); suffix++ {
				local = base + "_" + strconv.Itoa(suffix)
			}
		}
		locals[name] = local
		// A second member or namespace import must not reuse this binding.
		generated[local] = true
		specifier := name
		if local != name {
			specifier += " as " + local
		}
		specifiers = append(specifiers, specifier)
	}
	fixes := []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, "{ "+strings.Join(specifiers, ", ")+" }")}
	for _, use := range uses {
		fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, use.node,
			utils.SafeReplacementText(ctx.SourceFile, use.node, locals[use.name])))
	}
	return fixes
}
