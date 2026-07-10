package no_unused_private_class_members

import (
	"fmt"
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// memberKey distinguishes #-private members from public / private-keyword
// members. `#foo` carries the unique prefix so it never collides with a
// public `foo` in the same lookup map.
type memberKey string

func privateKey(name string) memberKey { return memberKey("#private@@" + name) }
func publicKey(name string) memberKey  { return memberKey(name) }

type member struct {
	declNode   *ast.Node // owning declaration (PropertyDeclaration / MethodDeclaration / Parameter / …)
	nameNode   *ast.Node // identifier or private-identifier the diagnostic anchors on
	name       string    // display name; hash-private members include the leading `#`
	key        memberKey
	isPrivate  bool // has the `private` keyword
	isHash     bool // declared with `#name`
	isAccessor bool // get/set accessor or auto-accessor field
	isStatic   bool
	readCount  int
	writeCount int
}

func (m *member) isUsed() bool {
	// Mirrors upstream: any access of an accessor counts (potential side effects);
	// regular members only stay alive when actually read.
	return m.readCount > 0 || (m.writeCount > 0 && m.isAccessor)
}

// classScope is the per-class collection of tracked members.
type classScope struct {
	node      *ast.Node
	className string // "" for anonymous class expressions
	instance  map[memberKey]*member
	static    map[memberKey]*member
}

// thisScope tracks how `this` resolves at each level of the AST walk.
type thisScope struct {
	upper       *thisScope
	owningClass *classScope // non-nil iff THIS scope was created by a Class node
	thisContext *classScope // class the `this` keyword resolves to here, or nil
	isStatic    bool        // when thisContext != nil, true ⇒ `this` is the class object
}

// findClassByName walks outwards looking for the nearest class scope whose
// owning declaration is named `name` (used to resolve `Foo.staticMember`).
func (s *thisScope) findClassByName(name string) *classScope {
	for cur := s; cur != nil; cur = cur.upper {
		if cur.owningClass != nil && cur.owningClass.className == name {
			return cur.owningClass
		}
	}
	return nil
}

type thisAlias struct {
	class    *classScope
	isStatic bool
}

var NoUnusedPrivateClassMembersRule = rule.CreateRule(rule.Rule{
	Name: "no-unused-private-class-members",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// Root scope corresponds to upstream's IntermediateScope-for-Program:
		// no `this` binding.
		root := &thisScope{}
		stack := []*thisScope{root}
		var classes []*classScope

		// aliasBySymbol keys `const X = this` style aliases by the variable's
		// symbol identity. Symbol equality intrinsically handles shadowing
		// and cross-scope visibility — no per-scope frames or name pushes
		// needed. Populated by recordAlias when an init `= this` is seen
		// during the walk.
		aliasBySymbol := map[*ast.Symbol]*thisAlias{}

		// writtenSymbols holds every variable symbol that appears in a
		// non-initializer write position anywhere in the file (collected
		// during the walk via utils.IsWriteReference). At EndOfFile we use
		// this to retroactively invalidate alias-mediated accesses whose
		// variable was reassigned — covers BOTH write-before-read and
		// read-before-write orderings, mirroring upstream's
		// `variable.references.some(ref => ref.isWrite() && !ref.init)`.
		writtenSymbols := map[*ast.Symbol]bool{}

		// pendingAliasAccess records each `member.readCount++` /
		// `writeCount++` that was attributed via an alias, so we can roll
		// it back at EndOfFile if the alias turns out to have been
		// reassigned. Direct (`this.X` / `Foo.X`) and typed-binding
		// (`thing: Foo`) accesses count immediately and are never deferred.
		type pendingAliasAccess struct {
			m       *member
			aliasOf *ast.Symbol
			isWrite bool
		}
		var pendingAliasAccesses []pendingAliasAccess

		current := func() *thisScope { return stack[len(stack)-1] }

		// lookupAliasFor returns the alias info for `node` (a receiver
		// Identifier), plus the alias variable's symbol so the caller can
		// register the access for deferred validation. Symbol-keyed
		// resolution requires TypeChecker; if absent, no aliases resolve.
		lookupAliasFor := func(node *ast.Node) (*thisAlias, *ast.Symbol) {
			if ctx.TypeChecker == nil {
				return nil, nil
			}
			sym := utils.GetReferenceSymbol(node, ctx.TypeChecker)
			if sym == nil {
				return nil, nil
			}
			if a, ok := aliasBySymbol[sym]; ok {
				return a, sym
			}
			return nil, nil
		}

		// ---------- member collection ----------

		memberFromDecl := func(decl *ast.Node) *member {
			name := decl.Name()
			if name == nil {
				return nil
			}
			if name.Kind == ast.KindPrivateIdentifier {
				priv := name.AsPrivateIdentifier()
				// tsgo's PrivateIdentifier.Text already carries the leading
				// `#`; keep it so the display name in reportUnused matches
				// the source identifier verbatim.
				return &member{
					declNode: decl,
					nameNode: name,
					name:     priv.Text,
					key:      privateKey(priv.Text),
					isHash:   true,
				}
			}
			if s, ok := utils.GetStaticPropertyName(name); ok {
				return &member{
					declNode: decl,
					nameNode: name,
					name:     s,
					key:      publicKey(s),
				}
			}
			return nil
		}

		memberFromParameter := func(param *ast.Node) *member {
			n := param.Name()
			if n == nil || n.Kind != ast.KindIdentifier {
				return nil
			}
			id := n.AsIdentifier()
			return &member{
				declNode: param,
				nameNode: n,
				name:     id.Text,
				key:      publicKey(id.Text),
			}
		}

		collectMembers := func(cs *classScope, classNode *ast.Node) {
			for _, m := range classNode.Members() {
				switch m.Kind {
				case ast.KindConstructor:
					ctor := m.AsConstructorDeclaration()
					if ctor == nil || ctor.Parameters == nil {
						continue
					}
					for _, param := range ctor.Parameters.Nodes {
						if param.Kind != ast.KindParameter {
							continue
						}
						if !ast.IsParameterPropertyDeclaration(param, m) {
							continue
						}
						mem := memberFromParameter(param)
						if mem == nil {
							continue
						}
						flags := ast.GetCombinedModifierFlags(param)
						mem.isPrivate = flags&ast.ModifierFlagsPrivate != 0
						cs.instance[mem.key] = mem
					}
				case ast.KindPropertyDeclaration,
					ast.KindMethodDeclaration,
					ast.KindGetAccessor,
					ast.KindSetAccessor:
					mem := memberFromDecl(m)
					if mem == nil {
						continue
					}
					flags := ast.GetCombinedModifierFlags(m)
					mem.isPrivate = flags&ast.ModifierFlagsPrivate != 0
					mem.isStatic = flags&ast.ModifierFlagsStatic != 0
					switch m.Kind {
					case ast.KindGetAccessor, ast.KindSetAccessor:
						mem.isAccessor = true
					case ast.KindPropertyDeclaration:
						if flags&ast.ModifierFlagsAccessor != 0 {
							mem.isAccessor = true
						}
					}
					if mem.isStatic {
						cs.static[mem.key] = mem
					} else {
						cs.instance[mem.key] = mem
					}
				}
				// ClassStaticBlockDeclaration and IndexSignature declare no
				// tracked members.
			}
		}

		// ---------- read/write classification ----------

		// classifyAccess returns true iff `memberExpr` is consumed purely as
		// a write. Mirrors upstream's countReference / isWriteOnlyUsage.
		classifyAccess := func(memberExpr *ast.Node) bool {
			target := ast.GetAssignmentTarget(memberExpr)
			if target == nil {
				return false
			}
			switch target.Kind {
			case ast.KindBinaryExpression:
				op := target.AsBinaryExpression().OperatorToken.Kind
				if op == ast.KindEqualsToken {
					// Simple `=` assignment, destructuring-element default,
					// and direct property-assignment destructuring all land
					// here.
					return true
				}
				return isStatementFormParent(target.Parent)
			case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
				return isStatementFormParent(target.Parent)
			case ast.KindForInStatement, ast.KindForOfStatement:
				return true
			}
			return false
		}

		// count tags the access immediately, and — when it was mediated by a
		// `const X = this` alias — also files a deferred pending entry so
		// EndOfFile can roll the count back if the alias variable turns out
		// to have been reassigned anywhere in the file.
		count := func(memberExpr *ast.Node, m *member, aliasOf *ast.Symbol) {
			isWrite := classifyAccess(memberExpr)
			if isWrite {
				m.writeCount++
			} else {
				m.readCount++
			}
			if aliasOf != nil {
				pendingAliasAccesses = append(pendingAliasAccesses, pendingAliasAccess{
					m: m, aliasOf: aliasOf, isWrite: isWrite,
				})
			}
		}

		// ---------- resolution of the access receiver ----------

		// resolveByTypeAnnotation looks for `obj` declared with an explicit
		// type annotation that points to a known class scope. Mirrors
		// upstream's Identifier → variable-defs → typeAnnotation branch.
		//
		// Class resolution stays bound to the lexical upper chain of the
		// access site (`current().findClassByName`) — same scope-visibility
		// semantics as upstream's `currentScope.findClassScopeWithName`.
		resolveByTypeAnnotation := func(obj *ast.Node) (*classScope, bool, bool) {
			if obj.Kind != ast.KindIdentifier {
				return nil, false, false
			}
			needle := obj.AsIdentifier().Text
			find := current().findClassByName
			// Prefer TypeChecker-backed symbol resolution when available:
			// it correctly walks every binding scope (nested blocks, etc.)
			// in one call. Fall back to a manual walk over enclosing
			// function-like / program bodies when TypeChecker is absent.
			if ctx.TypeChecker != nil {
				if cs, st, ok := resolveBySymbol(ctx.TypeChecker, obj, find); ok {
					return cs, st, true
				}
				return nil, false, false
			}
			scope := utils.FindEnclosingScope(obj)
			for scope != nil {
				if cs, st, ok := matchTypedBinding(scope, needle, find); ok {
					return cs, st, true
				}
				if scope.Kind == ast.KindSourceFile {
					break
				}
				scope = utils.FindEnclosingScope(scope)
			}
			return nil, false, false
		}

		// resolveAccessOwner returns the class scope that `obj` accesses,
		// whether the access is static, and — when the receiver was a
		// `const X = this` alias — the alias variable's symbol (so the
		// caller can register a deferred entry for post-walk validation).
		// Returns (nil, false, nil) if the receiver cannot be resolved.
		resolveAccessOwner := func(obj *ast.Node) (*classScope, bool, *ast.Symbol) {
			// Strip parens and TS type wrappers — `(x as T).foo`,
			// `x satisfies T`, `x!.foo` all preserve the runtime value of
			// `x`, so the access still targets the same class.
			obj = ast.SkipOuterExpressions(obj, ast.OEKParentheses|ast.OEKAssertions)
			scope := current()
			switch obj.Kind {
			case ast.KindThisKeyword:
				if scope.thisContext == nil {
					return nil, false, nil
				}
				return scope.thisContext, scope.isStatic, nil
			case ast.KindIdentifier:
				name := obj.AsIdentifier().Text
				if cs := scope.findClassByName(name); cs != nil {
					return cs, true, nil
				}
				if a, sym := lookupAliasFor(obj); a != nil {
					return a.class, a.isStatic, sym
				}
				if cs, st, ok := resolveByTypeAnnotation(obj); ok {
					return cs, st, nil
				}
			}
			return nil, false, nil
		}

		// handleThisDestructuring counts a READ on each non-computed
		// identifier key inside `pattern` against the current `this`
		// context. Used for `const {a} = this`, `({a} = this)`, and
		// `((({a} = this)) => …)` shapes.
		handleThisDestructuring := func(pattern *ast.Node) {
			scope := current()
			if scope.thisContext == nil {
				return
			}
			members := scope.thisContext.instance
			if scope.isStatic {
				members = scope.thisContext.static
			}
			eachDestructuredKey(pattern, func(name string) {
				if m, ok := members[publicKey(name)]; ok {
					m.readCount++
				}
			})
		}

		// ---------- scope handling ----------

		pushClassScope := func(node *ast.Node) {
			cs := &classScope{
				node:     node,
				instance: map[memberKey]*member{},
				static:   map[memberKey]*member{},
			}
			if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier {
				cs.className = name.AsIdentifier().Text
			}
			collectMembers(cs, node)
			ts := &thisScope{
				upper:       current(),
				owningClass: cs,
				thisContext: cs,
				isStatic:    false,
			}
			stack = append(stack, ts)
			classes = append(classes, cs)
		}

		pushMemberScope := func(_ *ast.Node, isStatic bool) {
			parent := current()
			ts := &thisScope{upper: parent}
			if parent != nil && parent.owningClass != nil {
				ts.thisContext = parent.owningClass
				ts.isStatic = isStatic
			}
			stack = append(stack, ts)
		}

		pushFunctionScope := func(node *ast.Node) {
			parent := current()
			ts := &thisScope{upper: parent}

			// Detect `function (this: Foo)` — the type annotation rebinds
			// `this` inside the function body to a known class scope.
			// Resolution is rooted at the function's *enclosing* scope's
			// upper chain (parent.findClassByName), matching upstream's
			// `upper.findClassScopeWithName` lookup site.
			if params := functionLikeParameters(node); len(params) > 0 {
				first := params[0]
				if first.Kind == ast.KindParameter {
					name := first.Name()
					if name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "this" {
						pd := first.AsParameterDeclaration()
						if pd != nil && pd.Type != nil {
							if cs, st, ok := classScopeFromTypeAnnotation(pd.Type, parent.findClassByName); ok {
								ts.thisContext = cs
								ts.isStatic = st
							}
						}
					}
				}
			}
			stack = append(stack, ts)
		}

		popScope := func(_ *ast.Node) {
			stack = stack[:len(stack)-1]
		}

		// ---------- alias tracking ----------

		recordAlias := func(name *ast.Node, init *ast.Node) {
			if name == nil || init == nil || name.Kind != ast.KindIdentifier {
				return
			}
			if ast.SkipParentheses(init).Kind != ast.KindThisKeyword {
				return
			}
			scope := current()
			if scope.thisContext == nil {
				return
			}
			// Symbol-keyed so shadowing and cross-scope distinctions fall
			// out for free. Requires TypeChecker; without it we can't
			// register the alias (the rule's `private` / parameter-property
			// surfaces are TS-only, so a TypeChecker is virtually always
			// present in real use).
			if ctx.TypeChecker == nil {
				return
			}
			sym := utils.GetReferenceSymbol(name, ctx.TypeChecker)
			if sym == nil {
				return
			}
			aliasBySymbol[sym] = &thisAlias{
				class:    scope.thisContext,
				isStatic: scope.isStatic,
			}
		}

		// handlePropAccess and handleElemAccess process a single
		// PropertyAccessExpression / ElementAccessExpression as a reference
		// to a tracked class member. Extracted so the BinaryExpression /
		// PropertyAssignment workaround below can re-trigger them on
		// computed-key subtrees that the linter's patternVisitor skips.
		var handlePropAccess func(node *ast.Node)
		var handleElemAccess func(node *ast.Node)
		handlePropAccess = func(node *ast.Node) {
			pa := node.AsPropertyAccessExpression()
			if pa == nil {
				return
			}
			keyNode := pa.Name()
			if keyNode == nil {
				return
			}
			if keyNode.Kind == ast.KindPrivateIdentifier {
				handlePrivateRef(node, keyNode, current(), count)
				return
			}
			if keyNode.Kind != ast.KindIdentifier {
				return
			}
			cls, isStatic, aliasSym := resolveAccessOwner(pa.Expression)
			if cls == nil {
				return
			}
			members := cls.instance
			if isStatic {
				members = cls.static
			}
			if m, ok := members[publicKey(keyNode.AsIdentifier().Text)]; ok {
				count(node, m, aliasSym)
			}
		}
		handleElemAccess = func(node *ast.Node) {
			ea := node.AsElementAccessExpression()
			if ea == nil || ea.ArgumentExpression == nil {
				return
			}
			name, ok := utils.GetStaticExpressionValue(ast.SkipParentheses(ea.ArgumentExpression))
			if !ok {
				return
			}
			cls, isStatic, aliasSym := resolveAccessOwner(ea.Expression)
			if cls == nil {
				return
			}
			members := cls.instance
			if isStatic {
				members = cls.static
			}
			if m, ok := members[publicKey(name)]; ok {
				count(node, m, aliasSym)
			}
		}

		// scanComputedKey walks `node` looking for member accesses the
		// linter's patternVisitor would have skipped (it never descends
		// into the *key* of a `[ expr ]: target` element when the parent
		// OLE is an assignment target). Each PropertyAccess / ElementAccess
		// encountered is replayed through the normal handlers.
		var scanComputedKey func(*ast.Node)
		scanComputedKey = func(n *ast.Node) {
			if n == nil {
				return
			}
			switch n.Kind {
			case ast.KindPropertyAccessExpression:
				handlePropAccess(n)
			case ast.KindElementAccessExpression:
				handleElemAccess(n)
			}
			n.ForEachChild(func(child *ast.Node) bool {
				scanComputedKey(child)
				return false
			})
		}

		return rule.RuleListeners{
			// Class scopes.
			ast.KindClassDeclaration:                      pushClassScope,
			rule.ListenerOnExit(ast.KindClassDeclaration): popScope,
			ast.KindClassExpression:                       pushClassScope,
			rule.ListenerOnExit(ast.KindClassExpression):  popScope,

			// Class-member function-likes inherit `this` from the class.
			ast.KindMethodDeclaration: func(node *ast.Node) {
				pushMemberScope(node, ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic))
			},
			rule.ListenerOnExit(ast.KindMethodDeclaration): popScope,
			ast.KindConstructor: func(node *ast.Node) {
				pushMemberScope(node, false)
			},
			rule.ListenerOnExit(ast.KindConstructor): popScope,
			ast.KindGetAccessor: func(node *ast.Node) {
				pushMemberScope(node, ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic))
			},
			rule.ListenerOnExit(ast.KindGetAccessor): popScope,
			ast.KindSetAccessor: func(node *ast.Node) {
				pushMemberScope(node, ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic))
			},
			rule.ListenerOnExit(ast.KindSetAccessor): popScope,
			ast.KindClassStaticBlockDeclaration: func(node *ast.Node) {
				pushMemberScope(node, true)
			},
			rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration): popScope,

			// Regular functions rebind `this` (unless they declare a typed
			// `this` parameter referring to a known class).
			ast.KindFunctionDeclaration:                      pushFunctionScope,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): popScope,
			ast.KindFunctionExpression:                       pushFunctionScope,
			rule.ListenerOnExit(ast.KindFunctionExpression):  popScope,

			// Property / element access: count read or write on the matching
			// tracked member.
			ast.KindPropertyAccessExpression: handlePropAccess,
			ast.KindElementAccessExpression:  handleElemAccess,

			// Workaround for the linter's destructuring walker. patternVisitor
			// on a PropertyAssignment visits only the *Initializer* (target);
			// the *Name* (e.g. `[this.#x]:` ) is silently dropped. Without
			// this listener, member accesses inside a computed destructuring
			// key would never reach the standard PAE / EAE handlers above.
			// We scan the name ONLY when the surrounding OLE is itself an
			// assignment target — value-context property assignments
			// continue to be walked by the linter as usual, so no
			// double-counting occurs.
			ast.KindPropertyAssignment: func(node *ast.Node) {
				pa := node.AsPropertyAssignment()
				if pa == nil {
					return
				}
				name := pa.Name()
				if name == nil || name.Kind != ast.KindComputedPropertyName {
					return
				}
				parent := node.Parent
				if parent == nil || parent.Kind != ast.KindObjectLiteralExpression {
					return
				}
				if !ast.IsAssignmentTarget(parent) {
					return
				}
				scanComputedKey(name)
			},

			ast.KindVariableDeclaration: func(node *ast.Node) {
				vd := node.AsVariableDeclaration()
				if vd == nil || vd.Initializer == nil {
					return
				}
				name := vd.Name()
				if name == nil {
					return
				}
				switch name.Kind {
				case ast.KindIdentifier:
					recordAlias(name, vd.Initializer)
				case ast.KindObjectBindingPattern:
					if ast.SkipParentheses(vd.Initializer).Kind == ast.KindThisKeyword {
						handleThisDestructuring(name)
					}
				}
			},

			ast.KindParameter: func(node *ast.Node) {
				pd := node.AsParameterDeclaration()
				if pd == nil || pd.Initializer == nil {
					return
				}
				name := pd.Name()
				if name == nil {
					return
				}
				if name.Kind == ast.KindObjectBindingPattern &&
					ast.SkipParentheses(pd.Initializer).Kind == ast.KindThisKeyword {
					handleThisDestructuring(name)
				}
			},

			// Universal write tracker. utils.IsWriteReference covers every
			// assignment / compound-op / update / destructuring shape, so a
			// single Identifier listener captures all non-init writes to a
			// variable — exactly what upstream's
			// `variable.references.some(ref => ref.isWrite() && !ref.init)`
			// gives via the scope manager. The symbol is the one shared
			// across all references; symbol identity disambiguates shadows.
			ast.KindIdentifier: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				if !utils.IsWriteReference(node) {
					return
				}
				if sym := utils.GetReferenceSymbol(node, ctx.TypeChecker); sym != nil {
					writtenSymbols[sym] = true
				}
			},

			ast.KindBinaryExpression: func(node *ast.Node) {
				be := node.AsBinaryExpression()
				if be == nil || be.OperatorToken == nil {
					return
				}
				if be.OperatorToken.Kind != ast.KindEqualsToken {
					return
				}
				if ast.SkipParentheses(be.Right).Kind != ast.KindThisKeyword {
					return
				}
				if left := ast.SkipParentheses(be.Left); left.Kind == ast.KindObjectLiteralExpression {
					handleThisDestructuring(left)
				}
			},

			// SourceFile.ForEachChild visits each statement followed by the
			// EndOfFileToken — using its listener as the "Program:exit"
			// equivalent guarantees every top-level reference has been
			// counted before we report.
			ast.KindEndOfFile: func(_ *ast.Node) {
				// Roll back every alias-mediated count whose underlying
				// variable was reassigned anywhere in the file. This is the
				// single-pass equivalent of upstream's eager
				// `variable.references.some(ref => ref.isWrite() && !ref.init)`
				// bailout: by the time we hit EndOfFile both the deferred
				// accesses and the writtenSymbols set are complete, so the
				// reconciliation is order-independent — read-before-write
				// and write-before-read collapse to the same outcome.
				for _, p := range pendingAliasAccesses {
					if !writtenSymbols[p.aliasOf] {
						continue
					}
					if p.isWrite {
						p.m.writeCount--
					} else {
						p.m.readCount--
					}
				}
				for _, cs := range classes {
					reportUnused(ctx, cs.instance)
					reportUnused(ctx, cs.static)
				}
			},
		}
	},
})

// handlePrivateRef matches a `obj.#name` access against the nearest
// enclosing class scope whose member map contains a matching key. Mirrors
// upstream's `PrivateIdentifier` visitor: walk the `upper` chain skipping
// scopes with a null thisContext (regular functions rebind `this`); the
// first scope that has the member wins. Hash-private members live in their
// own per-class namespace, so we don't need to distinguish instance vs
// static at this layer. The access is never alias-mediated (hash-private
// resolution is purely scope-based), so `count` is invoked with a nil
// alias symbol.
func handlePrivateRef(node *ast.Node, key *ast.Node, scope *thisScope, count func(*ast.Node, *member, *ast.Symbol)) {
	k := privateKey(key.AsPrivateIdentifier().Text)
	for cur := scope; cur != nil; cur = cur.upper {
		if cur.thisContext == nil {
			continue
		}
		if m, ok := cur.thisContext.instance[k]; ok {
			count(node, m, nil)
			return
		}
		if m, ok := cur.thisContext.static[k]; ok {
			count(node, m, nil)
			return
		}
	}
}

func reportUnused(ctx rule.RuleContext, members map[memberKey]*member) {
	// Collect and sort by source position so diagnostics emerge in a
	// stable, top-to-bottom order. Go's map iteration is randomized, which
	// would otherwise flip the reported line ordering between runs.
	pending := make([]*member, 0, len(members))
	for _, m := range members {
		if !m.isPrivate && !m.isHash {
			continue
		}
		if m.isUsed() {
			continue
		}
		pending = append(pending, m)
	}
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].nameNode.Pos() < pending[j].nameNode.Pos()
	})
	for _, m := range pending {
		ctx.ReportNode(m.nameNode, rule.RuleMessage{
			Id:          "unusedPrivateClassMember",
			Description: fmt.Sprintf("Private class member '%s' is defined but never used.", m.name),
			Data:        map[string]string{"classMemberName": m.name},
		})
	}
}

// isStatementFormParent returns true when `parent` (skipping parens) is an
// ExpressionStatement. Mirrors upstream's `parent.parent === ExpressionStatement`
// check on compound assignments and UpdateExpressions.
func isStatementFormParent(parent *ast.Node) bool {
	parent = ast.WalkUpParenthesizedExpressions(parent)
	return parent != nil && parent.Kind == ast.KindExpressionStatement
}

// functionLikeParameters returns `node.Parameters()` when `node` is
// function-like, and nil otherwise. Thin wrapper around the public tsgo
// getter to make the IsFunctionLike guard explicit at every call site.
func functionLikeParameters(node *ast.Node) []*ast.Node {
	if node == nil || !ast.IsFunctionLike(node) {
		return nil
	}
	return node.Parameters()
}

// eachDestructuredKey invokes `visit` for every non-computed identifier-keyed
// element of a destructuring pattern. The pattern can be either an
// ObjectBindingPattern (declaration context) or an ObjectLiteralExpression
// being used as a target (assignment context). Computed keys, rest elements,
// and non-identifier keys are intentionally skipped — same as upstream.
func eachDestructuredKey(pattern *ast.Node, visit func(name string)) {
	switch pattern.Kind {
	case ast.KindObjectBindingPattern:
		bp := pattern.AsBindingPattern()
		if bp == nil || bp.Elements == nil {
			return
		}
		for _, el := range bp.Elements.Nodes {
			if el.Kind != ast.KindBindingElement {
				continue
			}
			be := el.AsBindingElement()
			if be == nil {
				continue
			}
			// Rest element (`...rest`) gathers everything else; upstream's
			// `prop.type !== Property` filter drops it. The gathered name
			// is a *binding* target, not a class member key.
			if be.DotDotDotToken != nil {
				continue
			}
			var keyNode *ast.Node
			if be.PropertyName != nil {
				keyNode = be.PropertyName
			} else {
				keyNode = be.Name()
			}
			if keyNode == nil || keyNode.Kind != ast.KindIdentifier {
				continue
			}
			visit(keyNode.AsIdentifier().Text)
		}
	case ast.KindObjectLiteralExpression:
		ole := pattern.AsObjectLiteralExpression()
		if ole == nil || ole.Properties == nil {
			return
		}
		for _, prop := range ole.Properties.Nodes {
			var keyNode *ast.Node
			switch prop.Kind {
			case ast.KindShorthandPropertyAssignment:
				keyNode = prop.AsShorthandPropertyAssignment().Name()
			case ast.KindPropertyAssignment:
				keyNode = prop.AsPropertyAssignment().Name()
			default:
				continue
			}
			if keyNode == nil || keyNode.Kind != ast.KindIdentifier {
				continue
			}
			visit(keyNode.AsIdentifier().Text)
		}
	}
}

// resolveBySymbol uses TypeChecker to find the first declaration of `ident`
// and inspects its explicit type annotation. Mirrors upstream's
// `firstDef.name.typeAnnotation` branch but leverages the type checker's
// proper scope-aware symbol resolution (handles nested blocks, hoisting,
// declaration merging) instead of the manual walk. Returns the resolved
// class scope, isStatic, and ok flag.
//
// Only the FIRST declaration is consulted — same as upstream's
// `variable.defs[0]` check. Const-asserted patterns and complex inferred
// types are intentionally unsupported (mirroring upstream's known limit).
func resolveBySymbol(tc *checker.Checker, ident *ast.Node, find func(string) *classScope) (*classScope, bool, bool) {
	sym := tc.GetSymbolAtLocation(ident)
	if sym == nil || len(sym.Declarations) == 0 {
		return nil, false, false
	}
	decl := sym.Declarations[0]
	var typeNode *ast.Node
	switch decl.Kind {
	case ast.KindParameter:
		typeNode = decl.AsParameterDeclaration().Type
	case ast.KindVariableDeclaration:
		typeNode = decl.AsVariableDeclaration().Type
	default:
		// ClassDeclaration / FunctionDeclaration / etc. don't have a type
		// annotation we can use here. Upstream's "default branch" returns
		// null in the same way.
		return nil, false, false
	}
	if typeNode == nil {
		return nil, false, false
	}
	return classScopeFromTypeAnnotation(typeNode, find)
}

// classScopeFromTypeAnnotation maps a type annotation to a known class
// scope, returning (cls, isStatic, ok). Supports:
//   - `T` referencing a class: instance access.
//   - `typeof T` referencing a class: static access.
func classScopeFromTypeAnnotation(typ *ast.Node, find func(string) *classScope) (*classScope, bool, bool) {
	if typ == nil {
		return nil, false, false
	}
	switch typ.Kind {
	case ast.KindTypeReference:
		tr := typ.AsTypeReferenceNode()
		if tr == nil || tr.TypeName == nil || tr.TypeName.Kind != ast.KindIdentifier {
			return nil, false, false
		}
		if cs := find(tr.TypeName.AsIdentifier().Text); cs != nil {
			return cs, false, true
		}
	case ast.KindTypeQuery:
		tq := typ.AsTypeQueryNode()
		if tq == nil || tq.ExprName == nil || tq.ExprName.Kind != ast.KindIdentifier {
			return nil, false, false
		}
		if cs := find(tq.ExprName.AsIdentifier().Text); cs != nil {
			return cs, true, true
		}
	}
	return nil, false, false
}

// matchTypedBinding finds a binding for `needle` inside `scope` and, if it
// has a type annotation pointing to a known class, returns the resolution.
//
// `scope` is a function-like or program node returned by FindEnclosingScope.
// We inspect both parameters (most common case in the rule's test suite —
// `method(thing: Foo)`) and direct child variable declarations.
func matchTypedBinding(scope *ast.Node, needle string, find func(string) *classScope) (*classScope, bool, bool) {
	// Parameters.
	for _, p := range functionLikeParameters(scope) {
		if p.Kind != ast.KindParameter {
			continue
		}
		pn := p.Name()
		if pn == nil || pn.Kind != ast.KindIdentifier || pn.AsIdentifier().Text != needle {
			continue
		}
		pd := p.AsParameterDeclaration()
		if pd == nil || pd.Type == nil {
			return nil, false, false
		}
		return classScopeFromTypeAnnotation(pd.Type, find)
	}
	// Top-level variable declarations inside the scope's body / source file.
	if cs, st, ok := scanForVariableType(scope, needle, find); ok {
		return cs, st, ok
	}
	return nil, false, false
}

// scanForVariableType walks the immediate body of a function-like or the
// statements of a SourceFile looking for `let|const|var needle: T` and, if
// `T` resolves to a tracked class, returns the resolution.
func scanForVariableType(scope *ast.Node, needle string, find func(string) *classScope) (*classScope, bool, bool) {
	body := getBodyStatements(scope)
	for _, stmt := range body {
		if stmt.Kind != ast.KindVariableStatement {
			continue
		}
		decls := stmt.AsVariableStatement().DeclarationList
		if decls == nil {
			continue
		}
		dl := decls.AsVariableDeclarationList()
		if dl == nil || dl.Declarations == nil {
			continue
		}
		for _, d := range dl.Declarations.Nodes {
			if d.Kind != ast.KindVariableDeclaration {
				continue
			}
			vd := d.AsVariableDeclaration()
			if vd == nil {
				continue
			}
			n := vd.Name()
			if n == nil || n.Kind != ast.KindIdentifier || n.AsIdentifier().Text != needle {
				continue
			}
			if vd.Type == nil {
				return nil, false, false
			}
			return classScopeFromTypeAnnotation(vd.Type, find)
		}
	}
	return nil, false, false
}

// getBodyStatements returns the direct child statements of a function-like
// or SourceFile node.
func getBodyStatements(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindSourceFile:
		return node.AsSourceFile().Statements.Nodes
	case ast.KindFunctionDeclaration:
		if fd := node.AsFunctionDeclaration(); fd != nil && fd.Body != nil {
			return fd.Body.AsBlock().Statements.Nodes
		}
	case ast.KindFunctionExpression:
		if fe := node.AsFunctionExpression(); fe != nil && fe.Body != nil {
			return fe.Body.AsBlock().Statements.Nodes
		}
	case ast.KindArrowFunction:
		if af := node.AsArrowFunction(); af != nil && af.Body != nil && af.Body.Kind == ast.KindBlock {
			return af.Body.AsBlock().Statements.Nodes
		}
	case ast.KindMethodDeclaration:
		if md := node.AsMethodDeclaration(); md != nil && md.Body != nil {
			return md.Body.AsBlock().Statements.Nodes
		}
	case ast.KindConstructor:
		if ctor := node.AsConstructorDeclaration(); ctor != nil && ctor.Body != nil {
			return ctor.Body.AsBlock().Statements.Nodes
		}
	case ast.KindGetAccessor:
		if ga := node.AsGetAccessorDeclaration(); ga != nil && ga.Body != nil {
			return ga.Body.AsBlock().Statements.Nodes
		}
	case ast.KindSetAccessor:
		if sa := node.AsSetAccessorDeclaration(); sa != nil && sa.Body != nil {
			return sa.Body.AsBlock().Statements.Nodes
		}
	case ast.KindClassStaticBlockDeclaration:
		if csb := node.AsClassStaticBlockDeclaration(); csb != nil && csb.Body != nil {
			return csb.Body.AsBlock().Statements.Nodes
		}
	case ast.KindModuleBlock:
		return node.AsModuleBlock().Statements.Nodes
	}
	return nil
}
