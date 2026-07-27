package no_shadow_restricted_names

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ruleOptions struct {
	ReportGlobalThis bool
}

func parseOptions(opts any) ruleOptions {
	o := ruleOptions{ReportGlobalThis: true}
	m := utils.GetOptionsMap(opts)
	if m == nil {
		return o
	}
	if v, ok := m["reportGlobalThis"].(bool); ok {
		o.ReportGlobalThis = v
	}
	return o
}

func buildMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "shadowingRestrictedName",
		Description: "Shadowing of global property '" + name + "'.",
	}
}

// isForInOrOfLoopVariable reports whether the given VariableDeclaration is the
// loop variable of an enclosing `for (... in ...)` or `for (... of ...)`. The
// loop itself assigns a value to the variable on every iteration, so even
// declarations without an explicit initializer count as "written to".
func isForInOrOfLoopVariable(varDecl *ast.Node) bool {
	if varDecl == nil {
		return false
	}
	list := varDecl.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList {
		return false
	}
	loop := list.Parent
	if loop == nil || !ast.IsForInOrOfStatement(loop) {
		return false
	}
	stmt := loop.AsForInOrOfStatement()
	return stmt != nil && stmt.Initializer == list
}

// undefinedAnalysis is the one-time, whole-file precomputation that backs
// isSymbolSafelyShadowingUndefined and hasSameScopeNonVarUndefinedDeclaration,
// so that every "is this bare `undefined` declaration safe to shadow" check
// below is an O(1) map lookup instead of its own fresh scoped walk.
type undefinedAnalysis struct {
	// written holds the symbol of every identifier named "undefined" that is
	// written to (assignment target) anywhere in the file.
	written map[*ast.Symbol]bool
	// scopesWithFnOrClass holds every var-scope (function-like / module-block /
	// source-file / static block) that contains — anywhere inside it, not
	// crossing into a nested scope — a FunctionDeclaration or ClassDeclaration
	// named "undefined".
	scopesWithFnOrClass map[*ast.Node]bool
}

// analyzeUndefinedDeclarations walks the source file once, collecting both
// write-referenced "undefined" symbols and the scopes of any sibling
// function/class named "undefined". A single top-down walk suffices for the
// latter even though it isn't scope-bounded: utils.FindEnclosingScope always
// starts from the declaration's own parent chain, so it attributes each
// FunctionDeclaration/ClassDeclaration to its own immediate enclosing scope
// correctly regardless of how deeply the walk has descended to reach it.
//
// This is a manual, cheaply-filtered walk (only nodes whose text is literally
// "undefined", or whose kind is a function/class declaration) rather than a
// ctx.Refs.References query: the latter would need RefStore to index every
// identifier in the file by name, which costs more than this rule's own
// narrow scan.
func analyzeUndefinedDeclarations(ctx rule.RuleContext) undefinedAnalysis {
	result := undefinedAnalysis{
		written:             map[*ast.Symbol]bool{},
		scopesWithFnOrClass: map[*ast.Node]bool{},
	}
	if ctx.Refs == nil || ctx.SourceFile == nil {
		return result
	}

	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		switch n.Kind {
		case ast.KindIdentifier:
			if n.AsIdentifier().Text == "undefined" && utils.IsWriteReference(n) {
				// Resolve is a single binder scope-walk with no whole-file
				// indexing, and (like GetReferenceSymbol before it) resolves
				// the value-binding symbol for shorthand destructuring
				// assignments rather than the property symbol a TypeChecker
				// lookup would otherwise return.
				if sym := ctx.Refs.Resolve(n); sym != nil {
					result.written[sym] = true
				}
			}
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
			if name := n.Name(); name != nil && ast.IsIdentifier(name) && name.AsIdentifier().Text == "undefined" {
				if scope := utils.FindEnclosingScope(n); scope != nil {
					result.scopesWithFnOrClass[scope] = true
				}
			}
		}
		n.ForEachChild(func(c *ast.Node) bool {
			walk(c)
			return false
		})
	}

	walk(ctx.SourceFile.AsNode())
	return result
}

// isSymbolSafelyShadowingUndefined matches ESLint's safelyShadowsUndefined:
// every def of the symbol must be a plain VariableDeclaration without an
// initializer and not a for-in/of loop variable, AND the symbol must have no
// write references.
func isSymbolSafelyShadowingUndefined(sym *ast.Symbol, writtenSymbols map[*ast.Symbol]bool) bool {
	if sym == nil {
		// No type info — be permissive, same as when TypeChecker is nil.
		return true
	}
	if writtenSymbols[sym] {
		return false
	}
	if len(sym.Declarations) == 0 {
		return true
	}
	for _, decl := range sym.Declarations {
		if decl == nil || decl.Kind != ast.KindVariableDeclaration {
			return false
		}
		vd := decl.AsVariableDeclaration()
		if vd == nil || vd.Initializer != nil || isForInOrOfLoopVariable(decl) {
			return false
		}
	}
	return true
}

// hasSameScopeNonVarUndefinedDeclaration returns true if `identNode`'s
// enclosing var scope (function-like / module-block / source-file / static
// block) contains a sibling FunctionDeclaration or ClassDeclaration named
// "undefined". This is a fallback for ESLint's scope-manager merging that
// tsgo's TypeChecker does not always replicate (for example, TypeScript keeps
// `function undefined() {}` and `var undefined;` in the same scope as distinct
// symbols).
func hasSameScopeNonVarUndefinedDeclaration(identNode *ast.Node, scopesWithFnOrClass map[*ast.Node]bool) bool {
	scope := utils.FindEnclosingScope(identNode)
	return scope != nil && scopesWithFnOrClass[scope]
}

var NoShadowRestrictedNamesRule = rule.Rule{
	Name: "no-shadow-restricted-names",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		opts := parseOptions(options)
		restricted := map[string]bool{
			"undefined": true,
			"NaN":       true,
			"Infinity":  true,
			"arguments": true,
			"eval":      true,
		}
		if opts.ReportGlobalThis {
			restricted["globalThis"] = true
		}

		var analysis undefinedAnalysis
		undefinedComputed := false
		ensureUndefinedAnalysis := func() {
			if undefinedComputed {
				return
			}
			undefinedComputed = true
			analysis = analyzeUndefinedDeclarations(ctx)
		}

		reported := map[*ast.Node]bool{}
		report := func(ident *ast.Node, name string) {
			if ident == nil || reported[ident] {
				return
			}
			reported[ident] = true
			ctx.ReportNode(ident, buildMessage(name))
		}

		// checkIdent reports ident if its name is restricted. When allowSafeUndefined
		// is true (plain `var/let/const undefined;` without initializer), skip the
		// report if the symbol has no write references anywhere in the file.
		checkIdent := func(ident *ast.Node, name string, allowSafeUndefined bool) {
			if !restricted[name] {
				return
			}
			if name == "undefined" && allowSafeUndefined {
				if ctx.Refs == nil {
					// No reference info: be permissive, matching ESLint's
					// safelyShadowsUndefined when the declaration has no initializer.
					return
				}
				ensureUndefinedAnalysis()
				// ident is a declaration name, not a reference position, so it
				// resolves through the binder-attached symbol on its own
				// declaring node rather than ctx.Refs.Resolve.
				var sym *ast.Symbol
				if decl := ident.Parent; decl != nil && decl.Name() == ident {
					sym = decl.Symbol()
				}
				if isSymbolSafelyShadowingUndefined(sym, analysis.written) &&
					!hasSameScopeNonVarUndefinedDeclaration(ident, analysis.scopesWithFnOrClass) {
					return
				}
			}
			report(ident, name)
		}

		checkBinding := func(nameNode *ast.Node, allowSafeUndefined bool) {
			if nameNode == nil {
				return
			}
			utils.CollectBindingNames(nameNode, func(ident *ast.Node, name string) {
				checkIdent(ident, name, allowSafeUndefined)
			})
		}

		checkNamedDeclaration := func(node *ast.Node) {
			n := node.Name()
			if n == nil || !ast.IsIdentifier(n) {
				return
			}
			checkIdent(n, n.AsIdentifier().Text, false)
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				vd := node.AsVariableDeclaration()
				if vd == nil {
					return
				}
				allowSafe := vd.Initializer == nil && !isForInOrOfLoopVariable(node)
				checkBinding(vd.Name(), allowSafe)
			},

			ast.KindParameter: func(node *ast.Node) {
				// Skip parameters in type-level contexts (FunctionType,
				// ConstructorType, CallSignature, ConstructSignature,
				// MethodSignature, IndexSignature). They don't create runtime
				// bindings and are not visited by ESLint's :function selector.
				if node.Parent == nil || !ast.IsFunctionLikeDeclaration(node.Parent) {
					return
				}
				param := node.AsParameterDeclaration()
				if param == nil {
					return
				}
				checkBinding(param.Name(), false)
			},

			ast.KindCatchClause: func(node *ast.Node) {
				cc := node.AsCatchClause()
				if cc == nil || cc.VariableDeclaration == nil {
					return
				}
				vd := cc.VariableDeclaration.AsVariableDeclaration()
				if vd == nil {
					return
				}
				checkBinding(vd.Name(), false)
			},

			ast.KindFunctionDeclaration: checkNamedDeclaration,
			ast.KindFunctionExpression:  checkNamedDeclaration,
			ast.KindClassDeclaration:    checkNamedDeclaration,
			ast.KindClassExpression:     checkNamedDeclaration,

			ast.KindImportDeclaration: func(node *ast.Node) {
				imp := node.AsImportDeclaration()
				if imp == nil || imp.ImportClause == nil {
					return
				}
				clause := imp.ImportClause.AsImportClause()
				if clause == nil {
					return
				}

				checkImportName := func(name *ast.Node) {
					if name != nil && ast.IsIdentifier(name) {
						checkIdent(name, name.AsIdentifier().Text, false)
					}
				}

				// Default import: import X from '...'
				checkImportName(clause.Name())

				if clause.NamedBindings == nil {
					return
				}
				switch clause.NamedBindings.Kind {
				case ast.KindNamespaceImport:
					if ns := clause.NamedBindings.AsNamespaceImport(); ns != nil {
						checkImportName(ns.Name())
					}
				case ast.KindNamedImports:
					named := clause.NamedBindings.AsNamedImports()
					if named == nil || named.Elements == nil {
						return
					}
					for _, elem := range named.Elements.Nodes {
						if elem == nil {
							continue
						}
						if spec := elem.AsImportSpecifier(); spec != nil {
							checkImportName(spec.Name())
						}
					}
				}
			},
		}
	},
}
