package prefer_iterator_zip

// cspell:words gertd

import (
	"strings"
	"sync"

	"github.com/gertd/go-pluralize"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

var inflection = sync.OnceValue(func() *pluralize.Client {
	client := pluralize.NewClient()
	// The JavaScript inflector treats species as uncountable.
	client.AddUncountableRule("species")
	return client
})

func buildSuggestion(ctx rule.RuleContext, header core.TextRange, loop *ast.Node, bounds loopBounds, reads []inputReads) []rule.RuleSuggestion {
	if bounds.mathObject != nil && (!ctx.Globals.Access("Math").IsDeclared() || !ctx.Refs.IsGlobalReference(bounds.mathObject)) {
		return nil
	}
	// The replacement discards the old loop bindings and uses the surrounding
	// Iterator. Preserve suggestions for older editions with a runtime polyfill.
	if ctx.Globals.Access("Iterator") == utils.GlobalAccessOff ||
		!ctx.Refs.IsGlobalNameReference(loop.Parent, "Iterator", ast.SymbolFlagsValue|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias) ||
		ast.FindAncestorKind(loop, ast.KindWithStatement) != nil {
		return nil
	}
	comments := ctx.Comments.All()
	if utils.HasCommentInSpan(comments, header.Pos(), header.End()) {
		return nil
	}
	var fixes []rule.RuleFix
	if bounds.cachedStatement != nil {
		start := utils.TrimNodeTextRange(ctx.SourceFile, bounds.cachedStatement).Pos()
		if utils.HasCommentInSpan(comments, start, header.Pos()) {
			return nil
		}
		fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(start, header.Pos())))
	}
	for _, input := range reads {
		for _, member := range input.members {
			span := utils.TrimNodeTextRange(ctx.SourceFile, member)
			if utils.HasCommentInSpan(comments, span.Pos(), span.End()) || ast.FindAncestorKind(member, ast.KindWithStatement) != nil {
				return nil
			}
		}
	}

	used, enclosing := namesInScope(ctx, loop.AsForStatement().Statement)
	inputs := make([]string, 0, len(reads))
	names := make([]string, 0, len(reads))
	for _, input := range reads {
		// The Go port differs from pluralize on some inflections, such as cookies.
		name := inflection().Singular(input.name)
		if name == "" || name == input.name {
			name = "element"
		}
		if !isValidBindingName(name) {
			name += "_"
			if !isValidBindingName(name) {
				return nil
			}
		}
		for used[name] || hasEnclosingDeclaration(enclosing, name) || ctx.Globals.Access(name).IsDeclared() {
			name += "_"
		}
		used[name] = true
		names = append(names, name)
		inputs = append(inputs, input.name)
	}
	newHeader := "for (const [" + strings.Join(names, ", ") + "] of Iterator.zip([" + strings.Join(inputs, ", ") + "]))"
	fixes = append(fixes, rule.RuleFixReplaceRange(header, newHeader))
	for i, input := range reads {
		for _, member := range input.members {
			text := utils.SafeReplacementText(ctx.SourceFile, member, names[i])
			fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, member, text))
		}
	}
	return []rule.RuleSuggestion{{
		Message:  rule.RuleMessage{Id: ruleName + "/suggestion", Description: "Use `Iterator.zip(…)`."},
		FixesArr: fixes,
	}}
}

type scopeChildrenKey struct{}

func namesInScope(ctx rule.RuleContext, body *ast.Node) (map[string]bool, *scope.Scope) {
	manager := scopeanalysis.References(ctx, nil)
	children := rule.CachedByFile(ctx, scopeChildrenKey{}, func() map[*scope.Scope][]*scope.Scope {
		byParent := make(map[*scope.Scope][]*scope.Scope)
		for _, current := range manager.Scopes {
			if current.Parent != nil {
				byParent[current.Parent] = append(byParent[current.Parent], current)
			}
		}
		return byParent
	})
	bodyScope := manager.Acquire(body)
	used := map[string]bool{"Iterator": true}
	var collect func(*scope.Scope)
	collect = func(current *scope.Scope) {
		for name := range current.ByName {
			used[name] = true
		}
		for _, ref := range current.References {
			if ref.Resolved() == nil {
				used[ref.Identifier.Text()] = true
			}
		}
		for _, child := range children[current] {
			collect(child)
		}
	}
	collect(bodyScope)
	return used, bodyScope.Parent
}

func hasEnclosingDeclaration(enclosing *scope.Scope, name string) bool {
	for current := enclosing; current != nil; current = current.Parent {
		if len(current.Declarations(name)) != 0 {
			return true
		}
	}
	return false
}

func isValidBindingName(name string) bool {
	if !scanner.IsIdentifierText(name, core.LanguageVariantStandard) {
		return false
	}
	token := scanner.StringToToken(name)
	if token >= ast.KindFirstKeyword && token <= ast.KindLastFutureReservedWord {
		return false
	}
	// Unicorn reserves strict-mode bindings, global properties, and these
	// TypeScript contextual keywords when generating variable names.
	switch name {
	case "await", "yield", "arguments", "eval", "globalThis", "Infinity", "NaN", "undefined",
		"as", "any", "boolean", "constructor", "declare", "get", "module", "require", "number", "set", "string", "symbol", "type", "from", "of":
		return false
	}
	return true
}
