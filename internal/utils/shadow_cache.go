package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// ShadowCache answers IsShadowed for one source file without repeating the work
// that walk performs.
//
// IsShadowed walks from a node up to the SourceFile and, at the top, scans every
// top-level statement plus the hoisted `var` and function declarations of the
// whole file. That is a file-sized amount of work per call, so a rule that asks
// about one name at many nodes in the same file is quadratic in the number of
// those nodes. The cache removes that in three steps, from cheapest to last
// resort:
//
//   - A one-off declaration index rules the walk out entirely for a name the
//     file never binds, which is the common case for `Promise`, `undefined`,
//     `test` and the like.
//   - The source-file arm — the only file-sized part of the walk — is answered
//     once per name and shared by every node in the file, whatever scope it
//     sits in. Nodes in scopes of their own, such as one inline callback per
//     call site, get nothing from a per-scope cache and everything from this.
//   - What is left is the walk through the scopes between the node and the
//     file, whose cost is the nesting depth. That is answered once per scope,
//     so nodes that share one answer it once.
//
// A cache belongs to one file and one rule invocation; it holds AST nodes and
// must not outlive them.
type ShadowCache struct {
	sourceFile *ast.SourceFile
	names      map[string]bool
	shadowed   map[shadowCacheKey]bool
	files      map[fileShadowKey]bool
}

// shadowCacheKey identifies one scope walk: the nearest enclosing node the walk
// inspects, the name it looks for, and — only for the few scopes whose answer
// depends on which child the walk entered through — that child.
type shadowCacheKey struct {
	scope *ast.Node
	entry *ast.Node
	name  string
}

// NewShadowCache returns a cache for one source file.
func NewShadowCache(sourceFile *ast.SourceFile) *ShadowCache {
	return &ShadowCache{sourceFile: sourceFile}
}

// IsShadowed answers utils.IsShadowed(node, name), reusing work across calls on
// the same file. The answer is identical to calling IsShadowed directly.
func (cache *ShadowCache) IsShadowed(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}
	// Every binding has a declaration name, so the index is a superset of what
	// the walk can find: a miss rules the walk out, and a hit only costs the
	// walk. A file that never declares the name at all — the common case for
	// `Promise`, `undefined`, `test` — therefore pays one indexing pass and
	// nothing more.
	if !cache.declaresName(name) {
		return false
	}
	return cache.walk(node, name)
}

// DeclaresName reports whether the file binds name anywhere, from a single
// indexing pass. It is the fast-fail IsShadowed applies, exposed for callers
// that want to skip other work as well.
func (cache *ShadowCache) DeclaresName(name string) bool {
	return cache.declaresName(name)
}

func (cache *ShadowCache) declaresName(name string) bool {
	if cache.names == nil {
		cache.names = map[string]bool{}
		if cache.sourceFile != nil {
			var visit func(*ast.Node)
			visit = func(node *ast.Node) {
				if node.Kind == ast.KindIdentifier && ast.IsDeclarationName(node) && bindsDeclaredName(node) {
					cache.names[node.Text()] = true
				}
				node.ForEachChild(func(child *ast.Node) bool {
					visit(child)
					return false
				})
			}
			visit(cache.sourceFile.AsNode())
		}
	}
	return cache.names[name]
}

// bindsDeclaredName reports whether an identifier in declaration-name position
// introduces a binding a scope walk can find. Member names — object literal and
// class members, enum members, type parameters, JSX attributes — name a
// property, so an unrelated `const helper = { Promise() {} }` must not put
// `Promise` in the index and send every `Promise.resolve(…)` down the walk.
func bindsDeclaredName(name *ast.Node) bool {
	if name.Parent == nil {
		return false
	}
	switch name.Parent.Kind {
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
		ast.KindPropertyDeclaration, ast.KindPropertySignature,
		ast.KindMethodDeclaration, ast.KindMethodSignature,
		ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindEnumMember, ast.KindTypeParameter, ast.KindJsxAttribute:
		return false
	}
	return true
}

// walk answers IsShadowed once per (scope, name). A name the index cannot rule
// out — a parameter or local called `Promise` anywhere in the file — otherwise
// puts a full walk, including a scan of every top-level statement, on every
// node asking about it.
func (cache *ShadowCache) walk(node *ast.Node, name string) bool {
	scope, entry := shadowScope(node)
	if scope == nil {
		return isShadowed(node, name, cache.fileArm())
	}
	key := shadowCacheKey{scope: scope, name: name}
	if shadowScopeReadsEntry(scope) {
		key.entry = entry
	}
	if result, cached := cache.shadowed[key]; cached {
		return result
	}
	// Every node between node and scope is a plain expression: the walk neither
	// inspects it nor counts it as a crossed scope, so resuming from entry sees
	// exactly what a walk from node would.
	result := isShadowed(entry, name, cache.fileArm())
	if cache.shadowed == nil {
		cache.shadowed = map[shadowCacheKey]bool{}
	}
	cache.shadowed[key] = result
	return result
}

func (cache *ShadowCache) fileArm() map[fileShadowKey]bool {
	if cache.files == nil {
		cache.files = map[fileShadowKey]bool{}
	}
	return cache.files
}

// shadowScope returns the nearest ancestor of node that a scope walk inspects,
// along with the child of it that the walk arrives through.
func shadowScope(node *ast.Node) (scope *ast.Node, entry *ast.Node) {
	entry = node
	for current := node.Parent; current != nil; current = current.Parent {
		if isShadowScope(current) {
			return current, entry
		}
		entry = current
	}
	return nil, nil
}

// isShadowScope lists every node kind IsShadowed acts on, so that two nodes
// sharing a scope share its answer.
func isShadowScope(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindSourceFile, ast.KindBlock, ast.KindModuleBlock, ast.KindCaseBlock,
		ast.KindCatchClause, ast.KindClassStaticBlockDeclaration, ast.KindEnumDeclaration,
		ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
		ast.KindParameter:
		return true
	}
	return ast.IsFunctionLikeDeclaration(node) || ast.IsClassLike(node)
}

// shadowScopeReadsEntry reports whether a scope's answer depends on the child
// the walk arrives through: a parameter decorator, a class expression's own
// name, and a function's parameter environment are all outside the scope that
// encloses the rest of the construct.
func shadowScopeReadsEntry(scope *ast.Node) bool {
	return scope.Kind == ast.KindParameter || scope.Kind == ast.KindClassExpression ||
		ast.IsFunctionLikeDeclaration(scope)
}
