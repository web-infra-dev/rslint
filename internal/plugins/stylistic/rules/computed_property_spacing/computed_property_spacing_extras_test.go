// TestComputedPropertySpacingExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk
// it covers, so future refactors can't silently regress them without breaking
// a named lock-in.
package computed_property_spacing_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/computed_property_spacing"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestComputedPropertySpacingExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&computed_property_spacing.ComputedPropertySpacingRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: receiver wrappers (tsgo ParenthesizedExpression, non-null, type assertion) ----
			// ESLint's ESTree flattens `(obj)` to just `obj`; tsgo preserves the
			// ParenthesizedExpression. The `[`-finding scanner walks from
			// `Expression.End()` (the `)` of `(obj)`), so paren receivers must
			// still locate `[` correctly.
			{Code: `(obj)[foo]`, Options: optsNever()},
			{Code: `(obj)[ foo ]`, Options: optsAlways()},
			{Code: `obj![foo]`, Options: optsNever()},
			{Code: `obj![ foo ]`, Options: optsAlways()},
			{Code: `(obj as any)[foo]`, Options: optsNever()},
			{Code: `(obj as any)[ foo ]`, Options: optsAlways()},
			{Code: `(obj satisfies Record<string, unknown>)[foo]`, Options: optsNever()},

			// ---- Dimension 4: element access argument shapes ----
			// Identifier / NumericLiteral / StringLiteral / TemplateLiteral /
			// PropertyAccess — `]` is always at node.End()-1 regardless of
			// what kind the ArgumentExpression has.
			{Code: `obj[0]`, Options: optsNever()},
			{Code: `obj[ 0 ]`, Options: optsAlways()},
			{Code: "obj[`x`]", Options: optsNever()},
			{Code: "obj[ `x` ]", Options: optsAlways()},
			{Code: `obj[Symbol.iterator]`, Options: optsNever()},
			{Code: `obj[ Symbol.iterator ]`, Options: optsAlways()},

			// ---- Dimension 4: computed key shapes — numeric / template / member ----
			{Code: `var x = {[0]: 1}`, Options: optsNever()},
			{Code: `var x = {[ 0 ]: 1}`, Options: optsAlways()},
			{Code: "var x = {[`k`]: 1}", Options: optsNever()},
			{Code: "var x = {[ `k` ]: 1}", Options: optsAlways()},
			{Code: `var x = {[Symbol.iterator]: 1}`, Options: optsNever()},

			// ---- Dimension 4: comments-but-no-whitespace (preserves comment-as-token semantics) ----
			// `[/* c */ a]` has zero whitespace bytes between `[` and `/*`; our
			// SkipLeadingWhitespace stops at `/` and therefore correctly reports
			// no opening space.
			{Code: `obj[/* c */ a]`, Options: optsNever()},
			{Code: `obj[a /* c */]`, Options: optsNever()},
			{Code: "obj[// line\n a]", Options: optsNever()},

			// ---- Dimension 4: nesting boundaries — ElementAccess inside ComputedPropertyName ----
			// Verifies the three listeners cooperate independently — the inner
			// ElementAccessExpression and the outer ComputedPropertyName share
			// no state.
			{Code: `var x = {[obj[k]]: 1}`, Options: optsNever()},
			{Code: `var x = {[ obj[ k ] ]: 1}`, Options: optsAlways()},

			// ---- Dimension 4: nesting — class in class / class in object ----
			{Code: `class A { foo() { class B { [k]() {} } } }`, Options: optsNever()},
			{Code: `var x = { foo: class { [k]() {} } }`, Options: optsNever()},

			// ---- Dimension 4: IndexedAccessType — never option is NOT default ----
			// `TSIndexedAccessType` skips the `!node.computed` guard upstream
			// (it has no `.computed` field), so the type-side listener always
			// fires regardless of mode. Lock both modes in:
			{Code: `type T = A[K]`, Options: optsNever()},
			{Code: `type T = A[ K ]`, Options: optsAlways()},
			{Code: `type T = A[B | C]`, Options: optsNever()},
			{Code: `type T = A[ B | C ]`, Options: optsAlways()},
			{Code: "type T = A[\nK\n]", Options: optsNever()},
			{Code: "type T = A[\nK\n]", Options: optsAlways()},

			// ---- Dimension 4: class with `accessor` modifier — separate tsgo PropertyDeclaration shape ----
			{Code: `class A { static accessor [k] = 1 }`, Options: optsNever()},
			{Code: `class A { static accessor [ k ] = 1 }`, Options: optsAlways()},

			// ---- Real-user: template-literal computed keys (issue tracker shapes) ----
			{Code: "function f<T extends string>(k: T): unknown { return ({} as any)[`prefix_${k}`]; }", Options: optsNever()},
			// ---- Real-user: ternary argument expression ----
			{Code: `function pick(obj: any, first: boolean) { return obj[first ? "a" : "b"]; }`, Options: optsNever()},

			// ---- N/A: rule scope ----
			// "Identifier vs string / numeric / private / computed key" — the
			// rule by definition handles ONLY ComputedPropertyName. Non-computed
			// names (`{ a: 1 }`, `obj.a`, `class A { #x = 1 }`) are not in scope.
			// Locked in via the upstream non-computed valid cases.

			// ---- N/A: SpreadAssignment / RestElement ----
			// The rule never inspects spread/rest siblings; spreads contain no
			// computed bracket of their own. Locked-in implicitly by upstream
			// "var x = {...y, [k]: 1}"-style mixes — though upstream's suite
			// happens not to include one, the listener structure leaves spread
			// untouched by construction (no listener registered for spreads).

			// ---- N/A: graceful degradation on empty brackets ----
			// `obj[]` / `class A { []() {} }` are parser errors, not just
			// recovered AST shapes. Defensive guards (openPos>=closePos)
			// covered by code path; no positive test required.

			// ---- Robustness: deeper nesting ----
			// Two-layer ComputedPropertyName: outer key contains an
			// expression whose inner shape *also* has a ComputedPropertyName.
			// The outer listener's `closePos` must NOT mistake an inner `]`
			// for the outer one — the scanner-based finder cuts from the
			// inner expression's End(), so the next `]` is the outer.
			{Code: `var x = {[ {[ k ]: 1 }.k ]: 2}`, Options: optsAlways()},
			{Code: `var x = {[{[k]: 1}.k]: 2}`, Options: optsNever()},
			// IndexedAccessType nesting: `T = A[B[C]]` — outer `]` resolved
			// via scanner from inner-most's IndexType.End().
			{Code: `type T = A[B[C]]`, Options: optsNever()},
			{Code: `type T = A[ B[ C ] ]`, Options: optsAlways()},
			// Class method with ElementAccess inside the computed key — the
			// outer ComputedPropertyName and inner ElementAccessExpression
			// listen independently.
			{Code: `class A { [obj[k]](){} }`, Options: optsNever()},
			{Code: `class A { [ obj[ k ] ](){} }`, Options: optsAlways()},
			// declare class member with computed name — PropertyDeclaration
			// with no initializer, gated by enforceForClassMembers.
			{Code: `declare class A { [k]: number }`, Options: optsNever()},

			// =====================================================================
			//   Adversarial verification: each test below pairs with a specific
			//   unsafe assumption in the implementation. If the assumption breaks,
			//   the test fails. Keep the comment that names the assumption.
			// =====================================================================

			// ---- Assumption A: `findOpenBracketFrom(sf, expr.End())` reliably finds the rule's `[` even when receiver carries TS-only suffixes ----
			// Risk: scanner starting from `expr.End()` could mis-tokenize across
			// `<T>` / `!` / `as const` / `satisfies T` chains. Each receiver
			// shape below must still locate the *outer* `[`.
			{Code: `(arr as const)[0]`, Options: optsNever()},
			{Code: `(arr as const)[ 0 ]`, Options: optsAlways()},
			{Code: `(x satisfies number)[k]`, Options: optsNever()},
			{Code: `obj!!![k]`, Options: optsNever()}, // triple non-null assertion (valid TS)

			// ---- Assumption B: `findCloseBracketFrom(sf, arg.End())` skips over `]` bytes embedded INSIDE the argument's tokens ----
			// Risk: a string literal containing `]` ought to be one token, but a
			// hand-rolled byte scan would mistake it for the closing bracket and
			// report at the wrong column. The scanner sees one StringLiteral.
			{Code: `obj["x]y"]`, Options: optsNever()},
			{Code: `obj[ "x]y" ]`, Options: optsAlways()},
			// TemplateExpression with interpolation that itself contains `]`
			// from a nested ElementAccess — outer `]` resolution must traverse
			// past the template's interior.
			{Code: "obj[`pre_${other[k]}_post`]", Options: optsNever()},
			{Code: "obj[ `pre_${other[ k ]}_post` ]", Options: optsAlways()},

			// ---- Assumption C: MappedType `[P in K]` is NOT a ComputedPropertyName ----
			// Risk: my listeners might accidentally match MappedTypeNode's
			// type-parameter brackets. tsgo represents `[P in K]` as
			// TypeParameterDeclaration inside MappedTypeNode — a different
			// kind, NOT KindComputedPropertyName. The `T[P]` on the RHS IS an
			// IndexedAccessType and DOES get checked. So `[P in K]` must be
			// silent in both modes; the inner `T[P]` is the only thing under
			// audit.
			{Code: `type Pick<T, K extends keyof T> = { [P in K]: T[P] }`, Options: optsNever()},
			{Code: `type Pick<T, K extends keyof T> = { [P in K]: T[ P ] }`, Options: optsAlways()},

			// ---- Assumption D: IndexSignature `[k: string]` is NOT a ComputedPropertyName ----
			// Risk: same as Assumption C — tsgo uses IndexSignature, not
			// ComputedPropertyName. Both modes must remain silent.
			{Code: `interface I { [k: string]: any }`, Options: optsNever()},
			{Code: `interface I { [k: string]: any }`, Options: optsAlways()},
			{Code: `type T = { [k: number]: string }`, Options: optsNever()},

			// ---- Assumption E: chained `?.[]` / `[][]` each fire independently in source order ----
			// Risk: incorrect listener registration could miss one or report
			// out of order. Each link must report on its own.
			{Code: `obj?.[x]?.[y]`, Options: optsNever()},
			{Code: `obj?.[ x ]?.[ y ]`, Options: optsAlways()},
			{Code: `obj[a][b][c]`, Options: optsNever()},
			{Code: `obj[ a ][ b ][ c ]`, Options: optsAlways()},

			// ---- Assumption F: PropertyDeclaration / MethodDeclaration parent dispatch covers TS-only modifiers ----
			// Risk: `override`, `accessor`, decorators could change tsgo's
			// parent layout and silently make the listener stop firing.
			{Code: `class B extends A { override [k]() {} }`, Options: optsNever()},
			{Code: `class B extends A { override [ k ]() {} }`, Options: optsAlways()},

			// ---- Upstream parity: JSDoc `@type {...}` interior is silently SKIPPED ----
			// tsgo parses JSDoc type expressions into the regular type AST
			// (IndexedAccessTypeNode etc.) with NodeFlagsJSDoc set. The
			// upstream parser (@typescript-eslint/parser) drops JSDoc bodies
			// entirely, so upstream never sees these nodes and never reports.
			// Without the `inJSDoc` skip we'd over-report by ~30 cases when
			// auditing real codebases like rspack (rspack.config.js files
			// with `/** @type {ConstructorParameters<typeof X>[0]} */` are
			// the dominant case). Verified via differential audit against
			// ESLint+@stylistic on rspack@main.
			{Code: "/** @type {Foo[K]} */\nconst x = 1;", Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: "/** @type {Foo[ K ]} */\nconst x = 1;", Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: "/** @type {Arr[0]} */\nconst x = 1;", Options: optsAlways()},
			{Code: "const x = /** @type {Foo[ K ]} */ (val);", Options: optsNever()},

			// ---- Upstream parity: abstract members are silently SKIPPED ----
			// typescript-eslint represents `abstract [k](): void` as the
			// distinct AST kind TSAbstractMethodDefinition (likewise for
			// TSAbstractPropertyDefinition); upstream stylistic only listens
			// on MethodDefinition / PropertyDefinition, so abstract members
			// are silently skipped. tsgo keeps the same kind and uses a
			// modifier flag, so without the explicit abstract-skip in
			// isAbstractMember (see rule source) we'd over-report. These
			// pairs lock in the upstream-aligned silent behavior — note
			// that intentionally-bad spacing on the abstract member also
			// emits NO diagnostic.
			{Code: `abstract class A { abstract [k](): void }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `abstract class A { abstract [ k ](): void }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `abstract class A { abstract [ k ](): void }`, Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `abstract class A { abstract [k]: number }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `abstract class A { abstract [ k ]: number }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},

			// ---- Assumption G: `Symbol.iterator`-style class member names — by far the most common real-user computed name ----
			// Risk: PropertyAccess receiver inside ComputedPropertyName should
			// pass through without confusing the closePos scanner.
			{Code: `class Iter { [Symbol.iterator]() { return this; } async [Symbol.asyncIterator]() {} }`, Options: optsNever()},
			{Code: `class Iter { [ Symbol.iterator ]() { return this; } }`, Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true})},

			// ---- Assumption H: destructuring with default / nested-pattern / rest sibling ----
			// Risk: BindingElement's PropertyName must still be picked up when
			// there's a default-value initializer, and rest elements (no
			// PropertyName) must be silently skipped.
			{Code: `const { [k]: x = 1 } = obj`, Options: optsNever()},
			{Code: `const { [ k ]: x = 1 } = obj`, Options: optsAlways()},
			{Code: `const { [k]: { a, b } = {} } = obj`, Options: optsNever()},
			{Code: `const { ...rest } = obj`, Options: optsNever()}, // no PropertyName — must not fire
			{Code: `const [a, b, ...rest] = arr`, Options: optsNever()}, // array binding — must not fire

			// ---- Assumption I: spread before a computed key is invisible to the rule ----
			{Code: `const opts = { ...defaults, [key]: value }`, Options: optsNever()},
			{Code: `const opts = { ...defaults, [ key ]: value }`, Options: optsAlways()},

			// ---- Assumption J: `super[k]` / `this[k]` / `new f[k]()` — keyword receivers ----
			{Code: `class B extends A { foo() { return super[k]; } }`, Options: optsNever()},
			{Code: `class B extends A { foo() { return super[ k ]; } }`, Options: optsAlways()},
			{Code: `class B { foo(k) { return this[k]; } }`, Options: optsNever()},
			{Code: `const x = new factory[key]()`, Options: optsNever()},

			// ---- Assumption K: numeric BigInt literal in computed slot ----
			// Risk: tsgo splits Literal into KindNumericLiteral / KindBigIntLiteral.
			// BigInt args must traverse correctly.
			{Code: `obj[1n]`, Options: optsNever()},
			{Code: `obj[ 1n ]`, Options: optsAlways()},
			{Code: `const x = { [1n]: 'one' }`, Options: optsNever()},

			// ---- Assumption L: tuple-type indexed access `T[0]` ----
			{Code: `type First<T extends readonly unknown[]> = T[0]`, Options: optsNever()},
			{Code: `type First<T extends readonly unknown[]> = T[ 0 ]`, Options: optsAlways()},

			// ---- Assumption M: 4-layer same-kind nesting ----
			// Stress test for scanner-based closePos. Each layer's `]` must
			// resolve to the matching `]`, not a sibling.
			{Code: `obj[a[b[c[d]]]]`, Options: optsNever()},
			{Code: `obj[ a[ b[ c[ d ] ] ] ]`, Options: optsAlways()},
			{Code: `type T = A[B[C[D[E]]]]`, Options: optsNever()},

			// ---- Assumption N: BindingElement fires in every binding context ----
			// BindingElement appears under several different parent shapes;
			// each one must dispatch correctly. tsgo uses the same kind
			// regardless of context, so this validates the listener doesn't
			// have an implicit context assumption.
			{Code: `function f({ [a]: x }) {}`, Options: optsNever()},
			{Code: `function f({ [ a ]: x }) {}`, Options: optsAlways()},
			{Code: `try {} catch ({ [a]: x }) {}`, Options: optsNever()},
			{Code: `for (const { [k]: v } of arr) {}`, Options: optsNever()},
			{Code: `for (const { [ k ]: v } of arr) {}`, Options: optsAlways()},

			// ---- Assumption O: decorator on class member with computed name ----
			// Decorators are modifiers in tsgo; they don't change kind or
			// parent layout. Verifies parent dispatch still finds the
			// ComputedPropertyName.
			{Code: `class A { @dec [k]() {} }`, Options: optsNever()},
			{Code: `class A { @dec [ k ]() {} }`, Options: optsAlways()},
			{Code: `class A { @dec [k] = 0 }`, Options: optsNever()},

			// ---- Assumption P: declare class member fires (declare ≠ abstract) ----
			// Unlike `abstract`, the `declare` modifier in typescript-eslint
			// does NOT change the AST kind (member stays MethodDefinition /
			// PropertyDefinition). Upstream DOES report on it, so we must
			// too. This locks in that the abstract-skip in isAbstractMember
			// is correctly scoped to JUST abstract, not all TS modifiers.
			{Code: `declare class A { [k]: number }`, Options: optsNever()},
			{Code: `declare class A { [ k ]: number }`, Options: optsAlways()},

			// ---- Assumption Q: multi-byte char positions (CJK + non-BMP emoji) ----
			// Repository convention requires multi-byte char tests for any
			// position-calculating rule. tsgo TokenStart/TokenEnd are byte
			// offsets (see scanner.TokenText `s.text[tokenStart:pos]` in
			// typescript-go), and `scanner.GetECMALineAndUTF16CharacterOfPosition`
			// converts to UTF-16 column at the diagnostic-output boundary.
			// Verify both layers cooperate: rule reports survive CJK (3-byte
			// UTF-8, 1 UTF-16 unit) and non-BMP emoji (4-byte UTF-8, 2 UTF-16
			// units = surrogate pair). One invalid case below asserts exact
			// column numbers to lock the UTF-16 conversion.
			{Code: "const 表情 = '🎉'; obj[表情]", Options: optsNever()},
			{Code: "const 表情 = '🎉'; obj[ 表情 ]", Options: optsAlways()},
			{Code: "const x = '🎉🚀'; obj[x]", Options: optsNever()},

			// ---- Assumption R: abstract gate covers GetAccessor / SetAccessor (not just MethodDeclaration) ----
			// containerGated.isAbstractMember must fire uniformly across all
			// accessor variants — abstract get/set are silent like abstract
			// methods.
			{Code: `abstract class A { abstract get [ k ](): number }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `abstract class A { abstract set [ k ](v: number): void }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `abstract class A { abstract get [k](): number; abstract set [k](v: number) }`, Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true})},

			// ---- Assumption S: abstract `accessor` property (stage-3 accessor + abstract dual-modifier) ----
			// PropertyDeclaration with both `abstract` and `accessor`.
			// classMemberGated.isAbstractMember must win over the regular
			// accessor-property reporting path.
			{Code: `abstract class A { abstract accessor [ k ]: number }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `abstract class A { abstract accessor [k]: number }`, Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true})},

			// ---- Assumption T: inJSDoc covers ALL JSDoc tags, not just @type ----
			// JSDoc `@param`, `@returns`, `@typedef`, `@property`, etc. all
			// build type AST nodes with NodeFlagsJSDoc set. inJSDoc's flag
			// check is tag-agnostic, so each tag's interior must skip
			// uniformly.
			{Code: "/** @param {Foo[K]} x */\nfunction f(x: any) {}", Options: optsAlways()},
			{Code: "/** @returns {Foo[K]} */\nfunction f(): any {}", Options: optsAlways()},
			{Code: "/** @typedef {{ x: Foo[K] }} T */", Options: optsAlways()},

			// =====================================================================
			//   Context-coverage lock-ins: each verifies the rule fires correctly
			//   in a specific surrounding syntactic context. Listener dispatch is
			//   parent-kind-based; if a context routes nodes through unexpected
			//   parents, listener could silently fail to fire OR fire in wrong
			//   order. These are the spots where "I implemented it and tests
			//   pass" but "a real codebase finds the bug".
			// =====================================================================

			// ---- JSX expression container as ElementAccess host ----
			// (skipped here — shared stylistic fixtures/tsconfig.json doesn't
			// enable jsx. JSX dispatch is verified via the differential audit
			// in the rule README's audit notes; the ElementAccess listener
			// fires regardless of parent kind by construction.)

			// ---- Class static block (ES2022) ----
			// `class A { static { obj[k]; } }` — ElementAccess inside Block
			// inside ClassStaticBlockDeclaration. Our listener fires anyway.
			{Code: `class A { static { const x = obj[k]; } }`, Options: optsNever()},
			{Code: `class A { static { const x = obj[ k ]; } }`, Options: optsAlways()},

			// ---- For-of / for-in with ElementAccess as iterable / object ----
			{Code: `for (const x of arr[k]) {}`, Options: optsNever()},
			{Code: `for (const x of arr[ k ]) {}`, Options: optsAlways()},
			{Code: `for (const x in obj[k]) {}`, Options: optsNever()},

			// ---- Switch case expression ----
			{Code: `function f(x: any) { switch (x) { case obj[k]: break; } }`, Options: optsNever()},
			{Code: `function f(x: any) { switch (x) { case obj[ k ]: break; } }`, Options: optsAlways()},

			// ---- Throw / Return / Yield / Await ----
			{Code: `function f() { throw obj[k]; }`, Options: optsNever()},
			{Code: `function f() { return obj[k]; }`, Options: optsNever()},
			{Code: `async function f() { return await obj[k]; }`, Options: optsNever()},
			{Code: `function* g() { yield obj[k]; }`, Options: optsNever()},

			// ---- Dynamic import argument ----
			{Code: `const x = import(obj[k])`, Options: optsNever()},
			{Code: `const x = import(obj[ k ])`, Options: optsAlways()},

			// ---- NewExpression receiver-chain: `new Cls()[k]` ----
			// outermost is ElementAccess with receiver = NewExpression.
			{Code: `const x = new Cls()[k]`, Options: optsNever()},
			{Code: `const x = new Cls()[ k ]`, Options: optsAlways()},

			// ---- Tagged template tag itself is ElementAccess ----
			{Code: "const x = obj[k]`tpl`", Options: optsNever()},
			{Code: "const x = obj[ k ]`tpl`", Options: optsAlways()},

			// ---- IndexedAccessType inside generic type argument ----
			{Code: `type X = Array<T[K]>`, Options: optsNever()},
			{Code: `type X = Array<T[ K ]>`, Options: optsAlways()},
			// As type assertion target
			{Code: `const x = val as T[K]`, Options: optsNever()},
			{Code: `const x = val as T[ K ]`, Options: optsAlways()},
			// Function return / parameter type
			{Code: `function f(x: A[K]): B[J] { return x as any }`, Options: optsNever()},
			{Code: `function f(x: A[ K ]): B[ J ] { return x as any }`, Options: optsAlways()},
			// Conditional type containing IndexedAccessType
			{Code: `type T<X, K extends keyof X> = X extends Array<infer U> ? U[0] : X[K]`, Options: optsNever()},

			// ---- Namespace / module class member ----
			// MethodDeclaration's parent is ClassDeclaration even inside a
			// namespace; our containerGated dispatch unchanged.
			{Code: `namespace N { class A { [k]() {} } }`, Options: optsNever()},
			{Code: `namespace N { class A { [ k ]() {} } }`, Options: optsAlways()},
			// Module declaration form
			{Code: `module M { class A { [k]() {} } }`, Options: optsNever()},

			// ---- Default export with ElementAccess ----
			{Code: `export default obj[k]`, Options: optsNever()},
			{Code: `export default obj[ k ]`, Options: optsAlways()},

			// ---- `super[k]` / `this[k]` inside class methods (already in J above, re-verifying in get/set/static contexts) ----
			{Code: `class B extends A { static foo() { return super[k] } }`, Options: optsNever()},
			{Code: `class B extends A { get bar() { return super[k] } }`, Options: optsNever()},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Locks in upstream checkSpacing() arm B.1 (always + missing opening): ElementAccess paren receiver ----
			{
				Code:    `(obj)[foo]`,
				Output:  []string{`(obj)[ foo ]`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				},
			},
			// ---- Locks in upstream checkSpacing() arm B.2 (never + unexpected opening): non-null receiver ----
			{
				Code:    `obj![ foo ]`,
				Output:  []string{`obj![foo]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				},
			},
			// ---- Locks in: type-assertion wrapped receiver, never mode ----
			{
				Code:    `(obj as any)[ foo ]`,
				Output:  []string{`(obj as any)[foo]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},

			// ---- Locks in: IndexedAccessType listener arm (no `.computed` guard upstream) ----
			// TSIndexedAccessType is always checked regardless of mode-or-shape
			// of the outer node. Without the dedicated listener, type-level
			// `A[ B ]` would silently pass.
			{
				Code:    `type T = A[B | C]`,
				Output:  []string{`type T = A[ B | C ]`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},

			// ---- Locks in: ComputedPropertyName dispatch — ObjectLiteral method (not class member) ----
			// MethodDeclaration in ObjectLiteral shares its kind with class
			// methods; the grandparent dispatch must classify it as an object
			// property so it's NOT gated by enforceForClassMembers.
			{
				Code:    `var x = {[ k ]() { return 1 }}`,
				Output:  []string{`var x = {[k]() { return 1 }}`},
				Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
				},
			},
			// ---- Locks in: ComputedPropertyName dispatch — ObjectLiteral getter (not class member) ----
			{
				Code:    `var x = {get [ k ]() { return 1 }}`,
				Output:  []string{`var x = {get [k]() { return 1 }}`},
				Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- Locks in: BindingElement (destructuring) dispatch ----
			// Destructuring patterns route through `BindingElement.PropertyName`,
			// not `PropertyAssignment` — a separate parent kind in tsgo.
			{
				Code:    `const { [ a]: x } = obj;`,
				Output:  []string{`const { [a]: x } = obj;`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				},
			},

			// ---- Locks in: optional chain receiver (`?.[`) — the `?.` token sits between Expression.End() and `[` ----
			// The scanner-based `[`-finder skips across `?.` correctly; a
			// byte-walking implementation would mis-classify `?.` as
			// non-trivia and stop early.
			{
				Code:    `(maybe ?? other)?.[ foo ]`,
				Output:  []string{`(maybe ?? other)?.[foo]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Locks in: tabs are whitespace (ECMAScript §12.2 WhiteSpace covers HT) ----
			// SkipLeadingWhitespace / SkipTrailingWhitespace treat `\t` as
			// whitespace; without that, `[\tfoo\t]` (never) would silently pass.
			{
				Code:    "obj[\tfoo\t]",
				Output:  []string{`obj[foo]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter"},
					{MessageId: "unexpectedSpaceBefore"},
				},
			},

			// ---- Locks in: opening-only multi-line short-circuit (closing still reported) ----
			// Upstream's `isTokenOnSameLine(before, first)` short-circuits the
			// opening report when a newline intervenes. The closing report
			// runs independently — if the close-side first/last are same-line
			// with `]`, that side still fires.
			{
				Code:    "obj[\nfoo ]",
				Output:  []string{"obj[\nfoo]"},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore", Line: 2, Column: 4, EndLine: 2, EndColumn: 5},
				},
			},

			// ---- Locks in: nested ElementAccess inside ComputedPropertyName ----
			// Inner brackets belong to ElementAccessExpression, outer to
			// ComputedPropertyName — both listeners report independently.
			{
				Code:    `var x = {[obj[ k ]]: 1}`,
				Output:  []string{`var x = {[obj[k]]: 1}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},

			// ---- Locks in: PropertyDeclaration without initializer (class field) ----
			// PropertyDeclaration → isClassMember=true unconditionally
			// (PropertyDeclaration only appears in classes).
			{
				Code:    `class A { [ k ]; }`,
				Output:  []string{`class A { [k]; }`},
				Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},

			// ---- Locks in: 2-layer IndexedAccessType nesting — closePos must anchor at matching `]` ----
			// In `A[B[ C ]]` (always), the inner `[ C ]` is already spaced
			// — only outer brackets need fixes. The scanner-based closePos
			// for the outer layer must skip past the inner `]` and land on
			// the outer `]`; a naive `node.End()-1` could collide with
			// trailing trivia and misreport.
			{
				Code:    `type T = A[B[ C ]]`,
				Output:  []string{`type T = A[ B[ C ] ]`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},

			// ---- Locks in: ElementAccess inside class member's computed key ----
			// Outer key brackets are class-member-gated; inner ElementAccess
			// brackets always fire. Both should report independently in
			// source order.
			{
				Code:    `class A { [obj[ k ]](){} }`,
				Output:  []string{`class A { [obj[k]](){} }`},
				Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},

			// =====================================================================
			//   Adversarial invalid lock-ins — each pairs with one of the
			//   assumptions in the valid section above; if the implementation's
			//   listener wiring or position resolution drifts, the failure
			//   surfaces here.
			// =====================================================================

			// ---- Lock-in for Assumption E: chained optional access must report in source order ----
			// `obj?.[ x ]?.[ y ]` → 4 errors, columns must increase strictly,
			// alternating opening/closing per `?.[]` group.
			{
				Code:    `obj?.[ x ]?.[ y ]`,
				Output:  []string{`obj?.[x]?.[y]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
				},
			},

			// ---- Lock-in for Assumption E: chained access without optional ----
			{
				Code:    `obj[ a ][ b ]`,
				Output:  []string{`obj[a][b]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},

			// ---- Lock-in for Assumption B: template-literal arg with interpolation ----
			// `obj[ \`pre_${k}\` ]` never → 2 errors. The interpolated `${k}`
			// must not confuse the closePos scanner (no stray `]` inside the
			// template substitution).
			{
				Code:    "obj[ `pre_${k}_post` ]",
				Output:  []string{"obj[`pre_${k}_post`]"},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 21, EndLine: 1, EndColumn: 22},
				},
			},

			// ---- Lock-in for Assumption B: string literal containing `]` byte ----
			// The closing `]` is at the trailing byte; the embedded `]` inside
			// the string must not confuse the scanner.
			{
				Code:    `obj[ "x]y" ]`,
				Output:  []string{`obj["x]y"]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
				},
			},

			// ---- Lock-in for Assumption K: BigInt argument ----
			{
				Code:    `obj[ 1n ]`,
				Output:  []string{`obj[1n]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
				},
			},

			// ---- Lock-in for Assumption H: destructuring with default value ----
			{
				Code:    `const { [ k ]: x = 1 } = obj`,
				Output:  []string{`const { [k]: x = 1 } = obj`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},

			// ---- Lock-in for Assumption C: MappedType `[P in K]` MUST NOT report; only `T[P]` does ----
			// This is the critical false-positive guard. If the listener
			// accidentally matches MappedTypeNode's type-parameter brackets,
			// we'd report 4 errors. The correct behavior is exactly 2.
			{
				Code:    `type Pick<T, K extends keyof T> = { [P in K]: T[ P ] }`,
				Output:  []string{`type Pick<T, K extends keyof T> = { [P in K]: T[P] }`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 49, EndLine: 1, EndColumn: 50},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 51, EndLine: 1, EndColumn: 52},
				},
			},

			// ---- Lock-in for Assumption F: `override` modifier ----
			{
				Code:    `class B extends A { override [ k ]() {} }`,
				Output:  []string{`class B extends A { override [k]() {} }`},
				Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 33, EndLine: 1, EndColumn: 34},
				},
			},

			// ---- Lock-in for Assumption G: Symbol.iterator class member ----
			{
				Code:    `class Iter { [ Symbol.iterator ]() {} }`,
				Output:  []string{`class Iter { [Symbol.iterator]() {} }`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
				},
			},

			// ---- Lock-in for Assumption N: BindingElement in function parameter destructuring ----
			{
				Code:    `function f({ [ k ]: x }) {}`,
				Output:  []string{`function f({ [k]: x }) {}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},

			// ---- Lock-in for Assumption O: decorator on computed class member ----
			{
				Code:    `class A { @dec [ k ]() {} }`,
				Output:  []string{`class A { @dec [k]() {} }`},
				Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
				},
			},

			// ---- Lock-in for Assumption P: declare class member IS reported (not skipped like abstract) ----
			{
				Code:    `declare class A { [ k ]: number }`,
				Output:  []string{`declare class A { [k]: number }`},
				Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},

			// ---- Lock-in for Assumption Q: UTF-16 column output across CJK + non-BMP emoji ----
			// Source: `const 表情 = '🎉'; obj[ 表情 ]`
			//   UTF-16 cols 1-5 `const`, 6 space, 7-8 `表情`, 9 space, 10 `=`,
			//   11 space, 12 `'`, 13-14 `🎉` (surrogate pair, 2 units),
			//   15 `'`, 16 `;`, 17 space, 18-20 `obj`, 21 `[`, 22 space,
			//   23-24 `表情`, 25 space, 26 `]`.
			// unexpectedSpaceAfter loc = [innerLow, firstStart) = cols 22-23.
			// unexpectedSpaceBefore loc = [lastEnd, innerHigh) = cols 25-26.
			// If rslint accidentally byte-indexed or counted code points
			// instead of UTF-16 units, the columns would shift.
			{
				Code:    "const 表情 = '🎉'; obj[ 表情 ]",
				Output:  []string{"const 表情 = '🎉'; obj[表情]"},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
				},
			},

			// ---- Lock-in for Assumption M: 4-layer ElementAccess source order ----
			// `obj[ a[ b[ c[ d ] ] ] ]` never → 8 errors strictly increasing
			// columns (outer-open, ..., inner-open, inner-close, ..., outer-close).
			{
				Code:    `obj[ a[ b[ c[ d ] ] ] ]`,
				Output:  []string{`obj[a[b[c[d]]]]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
		},
	)
}
