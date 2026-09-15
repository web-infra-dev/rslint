package reactutil

import (
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

// Expectations were checked with eslint-plugin-react 7.37.5 and ESLint 9.39.5.
func TestIsExplicitReactComponent(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, code string
		want       []bool
	}{
		{name: "class declaration", code: `/** @extends React.Component */ class C {}`, want: []bool{true}},
		{name: "class expression", code: `/** @extends React.Component */ const C = class {};`, want: []bool{true}},
		{name: "parenthesized class expression", code: `/** @extends React.Component */ const C = (class {});`, want: []bool{true}},
		{name: "exported class", code: `/** @extends React.Component */ export class C {}`, want: []bool{true}},
		{name: "exported class expression", code: `/** @extends React.Component */ export const C = class {};`, want: []bool{false}},
		{name: "comment inside export", code: `export /** @extends React.Component */ const C = class {};`, want: []bool{true}},
		{name: "default export", code: `/** @extends React.Component */ export default (class {});`, want: []bool{true}},
		{name: "nested class", code: `/** @extends React.Component */ class Outer { C = class {}; }`, want: []bool{true, false}},
		{name: "function declaration", code: `/** @extends React.Component */ function C() {}`, want: []bool{true}},
		{name: "function expression", code: `/** @extends React.Component */ const C = function() {};`, want: []bool{true}},
		{name: "arrow function", code: `/** @extends React.Component */ const C = () => null;`, want: []bool{true}},
		{name: "intervening ordinary comment", code: `/** @extends React.Component */ const /* ordinary */ C = function() {};`, want: []bool{false}},
		{name: "intervening export comment", code: `/** @extends React.Component */ export /* ordinary */ const C = () => null;`, want: []bool{false}},
		{name: "object property function", code: `const Box = { /** @extends React.Component */ C: function() {} };`, want: []bool{true}},
		{name: "object method", code: `const Box = { /** @extends React.Component */ C() {} };`, want: []bool{true}},
		{name: "object getter", code: `const Box = { /** @extends React.Component */ get C() {} };`, want: []bool{true}},
		{name: "object setter", code: `const Box = { /** @extends React.Component */ set C(value) {} };`, want: []bool{true}},
		{name: "object literal comment", code: `const Box = /** @extends React.Component */ { C: class {} };`, want: []bool{true}},
		{name: "assignment comment", code: `const Box = {}; /** @extends React.Component */ Box.C = class {};`, want: []bool{true}},
		{name: "same-line local declaration", code: `function outer() { /** @extends React.Component */ const C = () => null; }`, want: []bool{false, true}},
		{name: "braced name", code: `/** @extends {React.Component} */ class C {}`, want: []bool{false}},
		{name: "generic name", code: `/** @extends React.Component<Props> */ class C {}`, want: []bool{false}},
		{name: "description", code: `/** @extends React.Component description */ class C {}`, want: []bool{false}},
		{name: "description before valid tag", code: `/** @extends Other nonsense
 * @augments React.Component */ class C {}`, want: []bool{false}},
		{name: "braces before valid tag", code: `/** @extends {Other}
 * @augments React.Component */ class C {}`, want: []bool{true}},
		{name: "multiline tag", code: `/** @extends
 * React.Component */ class C {}`, want: []bool{true}},
		{name: "augments pure component", code: `/** @augments React.PureComponent */ class C {}`, want: []bool{true}},
		{name: "multiple comments", code: `/** @extends React.Component */
/** ordinary */ class C {}`, want: []bool{false}},
		{name: "line comment", code: `/** @extends React.Component */
// ordinary
class C {}`, want: []bool{false}},
		{name: "blank line", code: `/** @extends React.Component */

class C {}`, want: []bool{false}},
		{name: "adjacent line", code: `/** @extends React.Component */
class C {}`, want: []bool{true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, script := range []struct {
				name string
				kind core.ScriptKind
			}{
				{"typescript", core.ScriptKindTSX}, {"javascript", core.ScriptKindJSX},
			} {
				t.Run(script.name, func(t *testing.T) {
					sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/component.tsx", Path: "/component.tsx"}, tc.code, script.kind)
					var got []bool
					var visit func(*ast.Node) bool
					visit = func(node *ast.Node) bool {
						switch node.Kind {
						case ast.KindClassDeclaration, ast.KindClassExpression, ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
							got = append(got, IsExplicitReactComponent(node))
						}
						node.ForEachChild(visit)
						return false
					}
					visit(sf.AsNode())
					if !slices.Equal(got, tc.want) {
						t.Fatalf("IsExplicitReactComponent() = %v, want %v", got, tc.want)
					}
				})
			}
		})
	}
}
