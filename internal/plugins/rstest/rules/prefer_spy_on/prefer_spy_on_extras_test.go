// TestPreferSpyOnExtras covers the shapes the migrated suite does not: every
// binding the utilities object is reached through, the assignment and target
// forms Rstest code uses, TypeScript syntax around the factory call, and the
// cases where the rewrite could not keep the source intact and no fix is
// offered.
package prefer_spy_on

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferSpyOnExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferSpyOnRule,
		[]rule_tester.ValidTestCase{
			// ---- Provenance ----
			// A local declaration of the name is a different object.
			{Code: `const rs = { fn: () => () => 1 }; obj.a = rs.fn()`},
			{Code: `function setup(rs) { obj.a = rs.fn() }`},
			{Code: `import { rs } from 'other-mocks'; obj.a = rs.fn()`},
			// Rstest has no standalone `fn` export, and `import.meta.rstest`
			// is the module namespace rather than the utilities object.
			{Code: `import.meta.rstest.fn()`},
			{Code: `obj.a = import.meta.rstest.fn()`},
			{Code: `obj.a = rs.spyOn(other, 'b')`},

			// ---- Compound assignments ----
			// These install the mock only conditionally, or combine it with
			// the current value; a spy is not a replacement for either.
			{Code: `obj.a ??= rs.fn()`},
			{Code: `obj.a ||= rs.fn()`},
			{Code: `obj.a &&= rs.fn()`},
			{Code: `obj.a += rs.fn()`},

			// ---- Right-hand side boundaries ----
			// Only the chain the value is read from is followed: a mock
			// passed to another call, a called mock and a mock inside
			// another expression are not what the property is assigned.
			{Code: `obj.a = wrap(rs.fn())`},
			{Code: `obj.a = helpers.wrap(rs.fn())`},
			{Code: `obj.a = rs.fn()()`},
			{Code: `obj.a = cond ? rs.fn() : other`},
			{Code: `obj.a = [rs.fn()]`},
			{Code: `obj.a = rs.fn`},
			{Code: `obj.a = rs.fn.bind(null)`},

			// ---- Targets ----
			{Code: `class Foo { #bar; test() { this.#bar = rs.fn() } }`},
			{Code: `[obj.a] = [rs.fn()]`},
			{Code: `({ a: obj.a } = { a: rs.fn() })`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Provenance: every binding of the utilities object ----
			{
				Code:   `obj.a = rstest.fn()`,
				Output: []string{`rstest.spyOn(obj, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 1, 20),
			},
			{
				Code:   `import { rs } from '@rstest/core'; obj.a = rs.fn()`,
				Output: []string{`import { rs } from '@rstest/core'; rs.spyOn(obj, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 36, 51),
			},
			{
				Code:   `import { rs as mocker } from '@rstest/core'; obj.a = mocker.fn()`,
				Output: []string{`import { rs as mocker } from '@rstest/core'; mocker.spyOn(obj, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 46, 65),
			},
			{
				Code:   `import { rstest } from 'rstack/test'; obj.a = rstest.fn()`,
				Output: []string{`import { rstest } from 'rstack/test'; rstest.spyOn(obj, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 39, 58),
			},
			{
				Code:   `import * as core from '@rstest/core'; obj.a = core.rs.fn()`,
				Output: []string{`import * as core from '@rstest/core'; core.rs.spyOn(obj, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 39, 59),
			},
			{
				Code:   `const { rs } = require('@rstest/core'); obj.a = rs.fn()`,
				Output: []string{`const { rs } = require('@rstest/core'); rs.spyOn(obj, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 41, 56),
			},
			{
				Code:   `const core = require('@rstest/core'); obj.a = core['rs'].fn()`,
				Output: []string{`const core = require('@rstest/core'); core['rs'].spyOn(obj, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 39, 62),
			},
			{
				Code:   `obj.a = import.meta.rstest.rs.fn()`,
				Output: []string{`import.meta.rstest.rs.spyOn(obj, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 1, 35),
			},
			{
				Code:   `obj.a = (rs as any).fn()`,
				Output: []string{`(rs as any).spyOn(obj, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 1, 25),
			},
			{
				Code:   `obj.a = rs?.fn()`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 1, 17),
			},
			{
				Code:   `obj.a = rs.fn?.()`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 1, 18),
			},
			// ---- Target forms ----
			{
				Code:   `this.a = rs.fn()`,
				Output: []string{`rs.spyOn(this, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 1, 17),
			},
			{
				Code:   `(obj.a) = rs.fn()`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 1, 18),
			},
			{
				Code:   `(obj as any).a = rs.fn()`,
				Output: []string{`rs.spyOn((obj as any), 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 1, 25),
			},
			{
				Code:   `getTarget().a = rs.fn()`,
				Output: []string{`rs.spyOn(getTarget(), 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 1, 24),
			},
			{
				Code:   `obj[0] = rs.fn()`,
				Output: []string{`rs.spyOn(obj, 0).mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 1, 17),
			},
			{
				Code:   `const m = (obj.a = rs.fn())`,
				Output: []string{`const m = (rs.spyOn(obj, 'a').mockImplementation(() => undefined))`},
				Errors: useSpyOn(1, 12, 27),
			},
			{
				Code:   `a.b = c.d = rs.fn()`,
				Output: []string{`a.b = rs.spyOn(c, 'd').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 7, 20),
			},
			// ---- Implementation argument ----
			{
				Code:   `obj.a = rs.fn(impl)`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(impl)`},
				Errors: useSpyOn(1, 1, 20),
			},
			{
				Code:   `obj.a = rs.fn(...implementations)`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(...implementations)`},
				Errors: useSpyOn(1, 1, 34),
			},
			{
				Code:   `obj.a = rs.fn<() => number>(() => 1)`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => 1)`},
				Errors: useSpyOn(1, 1, 37),
			},
			{
				Code:   `obj.a = rs.fn(impl).mockImplementation(other)`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(other)`},
				Errors: useSpyOn(1, 1, 46),
			},
			{
				Code:   `obj.a = rs.fn()['mockImplementation'](other)`,
				Output: []string{`rs.spyOn(obj, 'a')['mockImplementation'](other)`},
				Errors: useSpyOn(1, 1, 45),
			},
			{
				Code:   `obj.a = rs.fn().mockImplementationOnce(first)`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => undefined).mockImplementationOnce(first)`},
				Errors: useSpyOn(1, 1, 46),
			},
			{
				Code:   `obj.a = rs.fn(impl).mockImplementation`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(impl).mockImplementation`},
				Errors: useSpyOn(1, 1, 39),
			},
			// ---- Parentheses and type assertions around the factory ----
			{
				Code:   `obj.a = (rs.fn())`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 1, 18),
			},
			{
				Code:   `obj.a = ((rs.fn()))`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => undefined)`},
				Errors: useSpyOn(1, 1, 20),
			},
			{
				Code:   `foo.bar = (rs.fn()).mockImplementation(baz => baz)`,
				Output: []string{`rs.spyOn(foo, 'bar').mockImplementation(baz => baz)`},
				Errors: useSpyOn(1, 1, 51),
			},
			{
				Code:   `foo.bar = (rs.fn().mockImplementation(baz => baz))`,
				Output: []string{`rs.spyOn(foo, 'bar').mockImplementation(baz => baz)`},
				Errors: useSpyOn(1, 1, 51),
			},
			{
				Code:   `obj.a = (rs.fn().mockReturnValue(1))`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => undefined).mockReturnValue(1)`},
				Errors: useSpyOn(1, 1, 37),
			},
			{
				Code:   `obj.a = (rs.fn()).one.two()`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => undefined).one.two()`},
				Errors: useSpyOn(1, 1, 28),
			},
			{
				Code:   `obj.a = ((rs.fn()).one).two`,
				Output: []string{`(rs.spyOn(obj, 'a').mockImplementation(() => undefined).one).two`},
				Errors: useSpyOn(1, 1, 28),
			},
			{
				Code:   `obj.a = rs.fn() as any`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => undefined) as any`},
				Errors: useSpyOn(1, 1, 23),
			},
			{
				Code:   `obj.a = rs.fn()!`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => undefined)!`},
				Errors: useSpyOn(1, 1, 17),
			},
			{
				Code:   `obj.a = (rs.fn() as Mock).mockReturnValue(1)`,
				Output: []string{`(rs.spyOn(obj, 'a').mockImplementation(() => undefined) as Mock).mockReturnValue(1)`},
				Errors: useSpyOn(1, 1, 45),
			},
			{
				Code:   `obj.a = (rs.fn() satisfies Mock)`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => undefined) satisfies Mock`},
				Errors: useSpyOn(1, 1, 33),
			},
			// ---- Comments: kept when copied, otherwise the fix is withheld ----
			{
				Code:   `obj.a = rs.fn(() => /* stub */ 1)`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => /* stub */ 1)`},
				Errors: useSpyOn(1, 1, 34),
			},
			{
				Code:   `obj[/* key */ 'a'] = rs.fn()`,
				Output: nil,
				Errors: useSpyOn(1, 1, 29),
			},
			{
				Code:   `obj.a = /* mock */ rs.fn()`,
				Output: nil,
				Errors: useSpyOn(1, 1, 27),
			},
			{
				Code:   `obj.a /* mock */ = rs.fn()`,
				Output: nil,
				Errors: useSpyOn(1, 1, 27),
			},
			{
				Code:   `obj.a = rs.fn(/* stub */ impl)`,
				Output: nil,
				Errors: useSpyOn(1, 1, 31),
			},
			{
				Code:   `obj.a = (rs.fn() /* mock */).mockReturnValue(1)`,
				Output: nil,
				Errors: useSpyOn(1, 1, 48),
			},
			{
				Code:   `obj.a = (rs.fn().mockReturnValue(1) /* mock */)`,
				Output: nil,
				Errors: useSpyOn(1, 1, 48),
			},
			// ---- Other withheld fixes ----
			{
				Code:   `class A extends B { m() { super.a = rs.fn() } }`,
				Output: nil,
				Errors: useSpyOn(1, 27, 44),
			},
			{
				Code:   `obj.a = rs.fn(impl, extra)`,
				Output: nil,
				Errors: useSpyOn(1, 1, 27),
			},
		},
	)
}
