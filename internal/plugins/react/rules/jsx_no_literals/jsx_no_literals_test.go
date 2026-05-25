// cspell:ignore asdjfl blarg mdash
package jsx_no_literals

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxNoLiterals(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoLiteralsRule, []rule_tester.ValidTestCase{
		// ---- noStrings + allowedStrings allow specific attribute strings ----
		{
			Code: `
class Comp1 extends Component {
  render() {
    return (
      <div>
        <button type="button"></button>
      </div>
    );
  }
}
`,
			Tsx: true,
			Options: map[string]interface{}{
				"noStrings":      true,
				"allowedStrings": []interface{}{"button", "submit"},
			},
		},

		// ---- Default config: literals inside JSXExpressionContainer pass ----
		{
			Code: `
class Comp2 extends Component {
  render() {
    return (
      <div>
        {'asdjfl'}
      </div>
    );
  }
}
`,
			Tsx: true,
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    return (
      <>
        {'asdjfl'}
      </>
    );
  }
}
`,
			Tsx: true,
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    return (<div>{'test'}</div>);
  }
}
`,
			Tsx: true,
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    const bar = (<div>{'hello'}</div>);
    return bar;
  }
}
`,
			Tsx: true,
		},
		{
			Code: `
var Hello = createReactClass({
  foo: (<div>{'hello'}</div>),
  render() {
    return this.foo;
  },
});
`,
			Tsx: true,
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    return (
      <div>
        {'asdjfl'}
        {'test'}
        {'foo'}
      </div>
    );
  }
}
`,
			Tsx: true,
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    return (
      <div>
      </div>
    );
  }
}
`,
			Tsx: true,
		},
		{Code: `var foo = require('foo');`, Tsx: true},

		// ---- Attribute string literals are not Literal-handler reports by default ----
		{
			Code: `
<Foo bar='test'>
  {'blarg'}
</Foo>
`,
			Tsx: true,
		},

		// ---- ignoreProps: true suppresses attribute / non-content reports ----
		{
			Code: `
<Foo bar="test">
  {intl.formatText(message)}
</Foo>
`,
			Tsx: true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": true},
		},
		{
			Code: `
<Foo bar="test">
  {translate('my.translate.key')}
</Foo>
`,
			Tsx: true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": true},
		},

		// ---- noStrings allows non-string JSX expressions ----
		{Code: `<Foo bar={true} />`, Tsx: true, Options: map[string]interface{}{"noStrings": true}},
		{Code: `<Foo bar={false} />`, Tsx: true, Options: map[string]interface{}{"noStrings": true}},
		{Code: `<Foo bar={100} />`, Tsx: true, Options: map[string]interface{}{"noStrings": true}},
		{Code: `<Foo bar={null} />`, Tsx: true, Options: map[string]interface{}{"noStrings": true}},
		{Code: `<Foo bar={{}} />`, Tsx: true, Options: map[string]interface{}{"noStrings": true}},

		// ---- ignoreProps allows method-bound and identifier props with class attr ----
		{
			Code: `
class Comp1 extends Component {
  asdf() {}
  render() {
    return <Foo bar={this.asdf} class='xx' />;
  }
}
`,
			Tsx: true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": true},
		},

		// ---- Template literals in non-JSX contexts are ignored ----
		{
			Code: "\nclass Comp1 extends Component {\n  render() {\n    let foo = `bar`;\n    return <div />;\n  }\n}\n",
			Tsx:  true,
			Options: map[string]interface{}{"noStrings": true},
		},

		// ---- allowedStrings exempts specific JSX text ----
		{
			Code: `
class Comp1 extends Component {
  render() {
    return <div>asdf</div>
  }
}
`,
			Tsx:     true,
			Options: map[string]interface{}{"allowedStrings": []interface{}{"asdf"}},
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    return <div>asdf</div>
  }
}
`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": false, "allowedStrings": []interface{}{"asdf"}},
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    return <div>&nbsp;</div>
  }
}
`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "allowedStrings": []interface{}{"&nbsp;"}},
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    return (
      <div>
        &nbsp;
      </div>
    );
  }
}
`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "allowedStrings": []interface{}{"&nbsp;"}},
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    return <div>foo: {bar}*</div>
  }
}
`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "allowedStrings": []interface{}{"foo: ", "*"}},
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    return <div>foo</div>
  }
}
`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "allowedStrings": []interface{}{"   foo   "}},
		},

		// ---- Identifier expression children are unaffected ----
		{
			Code: `
class Comp1 extends Component {
  asdf() {}
  render() {
    const xx = 'xx';

    return <Foo bar={this.asdf} class={xx} />;
  }
}
`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false},
		},

		// ---- Default: attribute strings are allowed ----
		{Code: `<img alt='blank image'></img>`, Tsx: true},

		// ---- Mixed allowedStrings (entity + glyph) ----
		{
			Code:    `<div>&mdash;</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "allowedStrings": []interface{}{"&mdash;", "—"}},
		},
		{
			Code:    `<div>—</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "allowedStrings": []interface{}{"&mdash;", "—"}},
		},

		// ---- restrictedAttributes only fires on listed attributes ----
		{
			Code:    `<img src="image.jpg" alt="text" />`,
			Tsx:     true,
			Options: map[string]interface{}{"restrictedAttributes": []interface{}{"className", "id"}},
		},
		{
			Code:    `<div className="allowed" />`,
			Tsx:     true,
			Options: map[string]interface{}{"restrictedAttributes": []interface{}{"className"}, "allowedStrings": []interface{}{"allowed"}},
		},
		{
			Code: `<div className="test" title="hello" />`,
			Tsx:  true,
			Options: map[string]interface{}{
				"noStrings":            true,
				"ignoreProps":          true,
				"restrictedAttributes": []interface{}{"className"},
				"allowedStrings":       []interface{}{"test"},
			},
		},
		{
			Code:    `<div className="test" id="foo" />`,
			Tsx:     true,
			Options: map[string]interface{}{"restrictedAttributes": []interface{}{}},
		},

		// ---- elementOverrides: allowElement ----
		{
			Code: `<T>foo</T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowElement": true},
				},
			},
		},
		{
			Code: `<T>foo <div>bar</div></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowElement": true},
				},
			},
		},
		{
			Code: `<T>foo <div>{'bar'}</div></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowElement": true, "applyToNestedElements": false},
				},
			},
		},
		{
			Code: `
<div>
  <div>{'foo'}</div>
  <T>{2}</T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true},
				},
			},
		},
		{
			Code: `<T>{2}<div>{2}</div></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true},
				},
			},
		},
		{
			Code: `<T>{2}<div>{'foo'}</div></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true, "applyToNestedElements": false},
				},
			},
		},
		{
			Code: `
<div>
  <div>{'foo'}</div>
  <T>foo</T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowedStrings": []interface{}{"foo"}},
				},
			},
		},
		{
			Code: `<T>foo<div>foo</div></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowedStrings": []interface{}{"foo"}},
				},
			},
		},
		{
			Code: `<T>foo<div>{'foo'}</div></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowedStrings": []interface{}{"foo"}, "applyToNestedElements": false},
				},
			},
		},
		{
			Code: `
<div>
  <div foo={2} />
  <T foo="bar" />
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"noStrings": true,
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true, "ignoreProps": true},
				},
			},
		},
		{
			Code: `<T foo="bar"><div foo="bar" /></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"noStrings": true,
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true, "ignoreProps": true},
				},
			},
		},
		{
			Code: `<T foo="bar"><div foo={2} /></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"noStrings": true,
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true, "ignoreProps": true, "applyToNestedElements": false},
				},
			},
		},
		{
			Code: `
<div>
  <div foo="foo" />
  <T foo={2} />
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noAttributeStrings": true},
				},
			},
		},
		{
			Code: `<T foo={2}><div foo={2} /></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noAttributeStrings": true},
				},
			},
		},
		{
			Code: `<T foo={2}><div foo="foo" /></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noAttributeStrings": true, "applyToNestedElements": false},
				},
			},
		},
		{
			Code: `<T>foo<U>foo</U></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowedStrings": []interface{}{"foo"}},
					"U": map[string]interface{}{"allowedStrings": []interface{}{"foo"}},
				},
			},
		},

		// ---- Renamed-import resolution ----
		{
			Code: `
import { T } from 'foo';
<T>{'foo'}</T>
`,
			Tsx: true,
		},
		{
			Code: `
import { T as U } from 'foo';
<U>foo</U>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowElement": true},
				},
			},
		},
		{
			Code: `
const { T: U } = require('foo');
<U>foo</U>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowElement": true},
				},
			},
		},
		{
			Code: `
const { T: U } = require('foo').Foo;
<U>foo</U>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowElement": true},
				},
			},
		},
		{
			Code: `
const { T: U } = require('foo').Foo.Foo;
<U>foo</U>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowElement": true},
				},
			},
		},
		{
			Code: `
const foo = 2;
<T>foo</T>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowElement": true},
				},
			},
		},

		// ---- Compound (member-access) tag-name overrides ----
		{
			Code: `<T.U>foo</T.U>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T.U": map[string]interface{}{"allowElement": true},
				},
			},
		},
		{
			Code: `
import { T as U } from 'foo';
<U.U>foo</U.U>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T.U": map[string]interface{}{"allowElement": true},
				},
			},
		},
		{
			Code: `<React.Fragment>foo</React.Fragment>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"Fragment": map[string]interface{}{"allowElement": true},
				},
			},
		},
		{
			Code: `<React.Fragment>foo</React.Fragment>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"React.Fragment": map[string]interface{}{"allowElement": true},
				},
			},
		},

		// ---- HTML element overrides are silently rejected (regex gates them out) ----
		{
			Code: `<div>{'foo'}</div>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"div": map[string]interface{}{"allowElement": true},
				},
			},
		},

		// ---- Per-element restrictedAttributes ----
		{
			Code: `
<div>
  <Input type="text" />
  <Button className="primary" />
  <Image src="photo.jpg" />
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"Input":  map[string]interface{}{"restrictedAttributes": []interface{}{"placeholder"}},
					"Button": map[string]interface{}{"restrictedAttributes": []interface{}{"type"}},
				},
			},
		},
		{
			Code: `
<div title="container">
  <Button className="btn" />
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"restrictedAttributes": []interface{}{"className"},
				"elementOverrides": map[string]interface{}{
					"Button": map[string]interface{}{"restrictedAttributes": []interface{}{"disabled"}},
				},
			},
		},
		{
			Code: `<Button className="btn" />`,
			Tsx:  true,
			Options: map[string]interface{}{
				"noAttributeStrings": true,
				"elementOverrides": map[string]interface{}{
					"Button": map[string]interface{}{"restrictedAttributes": []interface{}{"type"}},
				},
			},
		},

		// ---- tsgo edge: paren-wrapped literal in JSX content (default OK) ----
		// `<div>{('foo')}</div>` parses with explicit ParenthesizedExpression
		// in tsgo (ESTree drops parens). Default config: literal lives under
		// JsxExpression after paren-skip, so the "wrapped" form is accepted.
		{Code: `<div>{('foo')}</div>`, Tsx: true},
		{Code: `<div>{(('foo'))}</div>`, Tsx: true},

		// ---- tsgo edge: TS `as` wrapper around a literal ----
		// Parent walk does NOT peel TSAsExpression (matches upstream's
		// `getParentIgnoringBinaryExpressions`, which only skips Binary).
		// The literal's effective parent is AsExpression — not in the
		// JSX-shape set — so no report fires even with noStrings:true.
		{Code: `<div>{'foo' as string}</div>`, Tsx: true},
		{
			Code:    `<div>{'foo' as string}</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true},
		},

		// ---- tsgo edge: JsxFragment outer / JsxElement override-target inner ----
		// jsxElementAncestors collects only JsxElement / JsxSelfClosingElement;
		// fragments are skipped because they have no name. Confirms an
		// `allowElement` override on T still matches when T is wrapped in a
		// fragment.
		{
			Code: `<><T>foo</T></>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowElement": true},
				},
			},
		},

		// ---- tsgo edge: namespaced attribute name passes through unchanged ----
		// Default config has no restrictedAttributes / noStrings, so the
		// attribute string is allowed.
		{Code: `<svg xlink:href="https://example.com" />`, Tsx: true},

		// ---- Namespaced attribute is INVISIBLE to restrictedAttributes ----
		// Upstream gates `restrictedAttributes` on `name.type === 'JSXIdentifier'`,
		// so a JsxNamespacedName never triggers this branch even when the
		// joined name (`xlink:href`) is in the list. Locks in the asymmetry.
		{
			Code:    `<svg xlink:href="anything" />`,
			Tsx:     true,
			Options: map[string]interface{}{"restrictedAttributes": []interface{}{"xlink:href"}},
		},

		// ---- tsgo edge: paren-wrapped require() in destructure init ----
		// `isRequireCall` peels ParenthesizedExpression at every read site so
		// the wrapped form behaves the same as upstream's ESTree-flat shape.
		{
			Code: `
const { T: U } = (require('foo')).Foo;
<U>foo</U>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowElement": true},
				},
			},
		},

		// ---- tsgo edge: rest binding in destructure is ignored ----
		// Upstream's `property.type === 'Property'` filter excludes
		// `RestElement`. We mirror by skipping BindingElements with a
		// DotDotDotToken so a `Rest → Rest` no-op entry never lands in the
		// renamed-import map.
		{
			Code: `
const { T, ...rest } = require('foo');
<T>foo</T>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowElement": true},
				},
			},
		},

		// ---- HTML-entity decoding parity: source uses entity, allowedStrings
		// lists only the decoded form. ESTree's `node.value` is decoded; we
		// decode tsgo's raw JsxText so the cooked-set lookup matches. ----
		{
			Code:    `<div>&amp;</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "allowedStrings": []interface{}{"&"}},
		},
		{
			Code:    `<div>&mdash;</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "allowedStrings": []interface{}{"—"}},
		},
		// And the inverse: source uses the glyph, allowedStrings lists only
		// the entity form — matches via the raw-source path.
		{
			Code:    `<div>—</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "allowedStrings": []interface{}{"—"}},
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Default: bare JSX text reports literalNotInJSXExpression ----
		{
			Code: `
class Comp1 extends Component {
  render() {
    return (<div>test</div>);
  }
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "test"`, Line: 4, Column: 18},
			},
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    return (<>test</>);
  }
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "test"`},
			},
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    const foo = (<div>test</div>);
    return foo;
  }
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "test"`},
			},
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    const varObjectTest = { testKey : (<div>test</div>) };
    return varObjectTest.testKey;
  }
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "test"`},
			},
		},
		{
			Code: `
var Hello = createReactClass({
  foo: (<div>hello</div>),
  render() {
    return this.foo;
  },
});
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "hello"`},
			},
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    return (
      <div>
        asdjfl
      </div>
    );
  }
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "asdjfl"`},
			},
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    return (
      <div>
        asdjfl
        test
        foo
      </div>
    );
  }
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: "Missing JSX expression container around literal string: \"asdjfl\n        test\n        foo\""},
			},
		},
		{
			Code: `
class Comp1 extends Component {
  render() {
    return (
      <div>
        {'asdjfl'}
        test
        {'foo'}
      </div>
    );
  }
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "test"`},
			},
		},

		// ---- noStrings: attribute literal + JSX child literal ----
		{
			Code: `
<Foo bar="test">
  {'Test'}
</Foo>
`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropValue", Message: `Invalid prop value: "bar="test""`},
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'Test'"`},
			},
		},
		{
			Code: `
<Foo bar="test">
  {'Test' + name}
</Foo>
`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropValue", Message: `Invalid prop value: "bar="test""`},
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'Test'"`},
			},
		},
		{
			Code: `
<Foo bar="test">
  Test
</Foo>
`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropValue", Message: `Invalid prop value: "bar="test""`},
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "Test"`},
			},
		},

		// ---- TemplateLiteral handler ----
		{
			Code:    "\n<Foo>\n  {`Test`}\n</Foo>\n",
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: "Strings not allowed in JSX files: \"`Test`\""},
			},
		},
		{
			Code:    "<Foo bar={`Test`} />",
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: "Strings not allowed in JSX files: \"`Test`\""},
			},
		},
		{
			Code:    "<Foo bar={`${baz}`} />",
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: "Strings not allowed in JSX files: \"`${baz}`\""},
			},
		},
		{
			Code:    "<Foo bar={`Test ${baz}`} />",
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: "Strings not allowed in JSX files: \"`Test ${baz}`\""},
			},
		},
		{
			Code:    "<Foo bar={`foo` + 'bar'} />",
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: "Strings not allowed in JSX files: \"`foo`\""},
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'bar'"`},
			},
		},
		{
			Code:    "<Foo bar={`foo` + `bar`} />",
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: "Strings not allowed in JSX files: \"`foo`\""},
				{MessageId: "noStringsInJSX", Message: "Strings not allowed in JSX files: \"`bar`\""},
			},
		},
		{
			Code:    "<Foo bar={'foo' + `bar`} />",
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'foo'"`},
				{MessageId: "noStringsInJSX", Message: "Strings not allowed in JSX files: \"`bar`\""},
			},
		},

		// ---- noStrings + allowedStrings exempts only listed strings ----
		{
			Code: `
class Comp1 extends Component {
  render() {
    return <div bar={'foo'}>asdf</div>
  }
}
`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "allowedStrings": []interface{}{"asd"}, "ignoreProps": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'foo'"`},
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "asdf"`},
			},
		},
		{
			Code:    `<Foo bar={'bar'} />`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'bar'"`},
			},
		},

		// ---- noAttributeStrings ----
		{
			Code:    `<img alt='blank image'></img>`,
			Tsx:     true,
			Options: map[string]interface{}{"noAttributeStrings": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInAttributes", Message: `Strings not allowed in attributes: "'blank image'"`},
			},
		},
		{
			Code:    `export const WithChildren = ({}) => <div>baz bob</div>;`,
			Tsx:     true,
			Options: map[string]interface{}{"noAttributeStrings": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "baz bob"`},
			},
		},
		{
			Code:    `export const WithAttributes = ({}) => <div title="foo bar" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"noAttributeStrings": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInAttributes", Message: `Strings not allowed in attributes: ""foo bar""`},
			},
		},
		{
			Code: `
export const WithAttributesAndChildren = ({}) => (
  <div title="foo bar">baz bob</div>
);
`,
			Tsx:     true,
			Options: map[string]interface{}{"noAttributeStrings": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInAttributes", Message: `Strings not allowed in attributes: ""foo bar""`},
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "baz bob"`},
			},
		},

		// ---- restrictedAttributes ----
		{
			Code:    `<div className="test" />`,
			Tsx:     true,
			Options: map[string]interface{}{"restrictedAttributes": []interface{}{"className"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "restrictedAttributeString", Message: `Restricted attribute string: ""test"" in className`},
			},
		},
		{
			Code:    `<div className="test" id="foo" title="bar" />`,
			Tsx:     true,
			Options: map[string]interface{}{"restrictedAttributes": []interface{}{"className", "id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "restrictedAttributeString", Message: `Restricted attribute string: ""test"" in className`},
				{MessageId: "restrictedAttributeString", Message: `Restricted attribute string: ""foo"" in id`},
			},
		},
		{
			Code:    `<div src="image.jpg" />`,
			Tsx:     true,
			Options: map[string]interface{}{"noAttributeStrings": true, "restrictedAttributes": []interface{}{"className"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInAttributes", Message: `Strings not allowed in attributes: ""image.jpg""`},
			},
		},
		{
			Code:    `<div title="text">test</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"restrictedAttributes": []interface{}{"title"}, "noStrings": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "restrictedAttributeString", Message: `Restricted attribute string: ""text"" in title`},
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "test"`},
			},
		},
		{
			Code:    `<div className="test" title="hello" />`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false, "restrictedAttributes": []interface{}{"className"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "restrictedAttributeString", Message: `Restricted attribute string: ""test"" in className`},
				{MessageId: "invalidPropValue", Message: `Invalid prop value: "title="hello""`},
			},
		},
		{
			Code:    `<div className="test" title="hello" />`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": true, "restrictedAttributes": []interface{}{"className"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "restrictedAttributeString", Message: `Restricted attribute string: ""test"" in className`},
			},
		},

		// ---- elementOverrides errors ----
		{
			Code: `
<div>
  <div>foo</div>
  <T>bar</T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{"T": map[string]interface{}{}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "foo"`},
				{MessageId: "literalNotInJSXExpressionInElement", Message: `Missing JSX expression container around literal string: "bar" in T`},
			},
		},
		{
			Code: `
<div>
  <div>foo</div>
  <T>bar</T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{"T": map[string]interface{}{"allowElement": true}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "foo"`},
			},
		},
		{
			Code: `<T>foo <div>bar</div></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{"T": map[string]interface{}{"allowElement": true, "applyToNestedElements": false}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "bar"`},
			},
		},
		{
			Code: `
<div>
  <div>foo</div>
  <T>{'bar'}</T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{"T": map[string]interface{}{"noStrings": true}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "foo"`},
				{MessageId: "noStringsInJSXInElement", Message: `Strings not allowed in JSX files: "'bar'" in T`},
			},
		},
		{
			Code: `
<div>
  <div>foo</div>
  <T>{'bar'}<div>{'baz'}</div></T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{"T": map[string]interface{}{"noStrings": true}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "foo"`},
				{MessageId: "noStringsInJSXInElement", Message: `Strings not allowed in JSX files: "'bar'" in T`},
				{MessageId: "noStringsInJSXInElement", Message: `Strings not allowed in JSX files: "'baz'" in T`},
			},
		},
		{
			Code: `
<div>
  <div>foo</div>
  <T>{'bar'}<div>{'baz'}</div></T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{"T": map[string]interface{}{"noStrings": true, "applyToNestedElements": false}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "foo"`},
				{MessageId: "noStringsInJSXInElement", Message: `Strings not allowed in JSX files: "'bar'" in T`},
			},
		},
		{
			Code: `
<div>
  <div>{'foo'}</div>
  <T>{'foo'}</T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"noStrings": true,
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true, "allowedStrings": []interface{}{"foo"}},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'foo'"`},
			},
		},
		{
			Code: `
<div>
  <div>{'foo'}</div>
  <T>{'foo'}<div>{'foo'}</div></T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"noStrings": true,
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true, "allowedStrings": []interface{}{"foo"}},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'foo'"`},
			},
		},
		{
			Code: `
<div>
  <div>{'foo'}</div>
  <T>{'foo'}<div>{'foo'}</div></T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"noStrings": true,
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true, "allowedStrings": []interface{}{"foo"}, "applyToNestedElements": false},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'foo'"`},
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'foo'"`},
			},
		},
		{
			Code: `
<div>
  <div foo1="bar" />
  <T foo2="bar" />
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"noStrings": true,
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true, "ignoreProps": true},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropValue", Message: `Invalid prop value: "foo1="bar""`},
			},
		},
		{
			Code: `
<div>
  <div foo1="bar" />
  <T foo2="bar"><div foo3="bar" /></T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"noStrings": true,
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true, "ignoreProps": true},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropValue", Message: `Invalid prop value: "foo1="bar""`},
			},
		},
		{
			Code: `
<div>
  <div foo1="bar" />
  <T foo2="bar"><div foo3="bar" /></T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"noStrings": true,
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true, "ignoreProps": true, "applyToNestedElements": false},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropValue", Message: `Invalid prop value: "foo1="bar""`},
				{MessageId: "invalidPropValue", Message: `Invalid prop value: "foo3="bar""`},
			},
		},
		{
			Code: `
<div>
  <div foo1="bar1" />
  <T foo2="bar2" />
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noAttributeStrings": true},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInAttributesInElement", Message: `Strings not allowed in attributes: ""bar2"" in T`},
			},
		},
		{
			Code: `
<div>
  <div foo1="bar1" />
  <T foo2="bar2"><div foo3="bar3" /></T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noAttributeStrings": true},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInAttributesInElement", Message: `Strings not allowed in attributes: ""bar2"" in T`},
				{MessageId: "noStringsInAttributesInElement", Message: `Strings not allowed in attributes: ""bar3"" in T`},
			},
		},
		{
			Code: `
<div>
  <div foo1="bar1" />
  <T foo2="bar2"><div foo3="bar3" /></T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noAttributeStrings": true, "applyToNestedElements": false},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInAttributesInElement", Message: `Strings not allowed in attributes: ""bar2"" in T`},
			},
		},
		{
			Code: `
<div>
  <div>{'foo'}</div>
  <T>{'bar'}</T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"noStrings": true,
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'foo'"`},
			},
		},
		{
			Code: `
<div>
  <div>foo</div>
  <T>foo</T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"allowedStrings": []interface{}{"foo"},
				"elementOverrides": map[string]interface{}{"T": map[string]interface{}{}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpressionInElement", Message: `Missing JSX expression container around literal string: "foo" in T`},
			},
		},
		{
			Code: `
<div>
  <div>foo</div>
  <T>foo</T>
  <T>bar</T>
  <T>baz</T>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"allowedStrings": []interface{}{"foo"},
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowedStrings": []interface{}{"bar"}},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpressionInElement", Message: `Missing JSX expression container around literal string: "foo" in T`},
				{MessageId: "literalNotInJSXExpressionInElement", Message: `Missing JSX expression container around literal string: "baz" in T`},
			},
		},
		{
			Code: `
<div>
  <div foo1="bar1" />
  <T foo2="bar2" />
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"noStrings":   true,
				"ignoreProps": true,
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropValueInElement", Message: `Invalid prop value: "foo2="bar2"" in T`},
			},
		},
		{
			Code: `
<div>
  <div foo1="bar1" />
  <T foo2="bar2" />
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"noAttributeStrings": true,
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInAttributes", Message: `Strings not allowed in attributes: ""bar1""`},
			},
		},
		{
			Code: `
<div>
  <T>foo</T>
  <U>bar</U>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{},
					"U": map[string]interface{}{},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpressionInElement", Message: `Missing JSX expression container around literal string: "foo" in T`},
				{MessageId: "literalNotInJSXExpressionInElement", Message: `Missing JSX expression container around literal string: "bar" in U`},
			},
		},
		{
			Code: `
<div>
  <T>foo</T>
  <U>bar</U>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{},
					"U": map[string]interface{}{"allowElement": true},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpressionInElement", Message: `Missing JSX expression container around literal string: "foo" in T`},
			},
		},
		{
			Code: `<T>foo <U>bar</U></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{},
					"U": map[string]interface{}{"allowElement": true},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpressionInElement", Message: `Missing JSX expression container around literal string: "foo" in T`},
			},
		},
		{
			Code: `<T>{'foo'}<U>{'bar'}</U></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true},
					"U": map[string]interface{}{},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSXInElement", Message: `Strings not allowed in JSX files: "'foo'" in T`},
			},
		},
		{
			Code: `<T>foo<U>foo</U></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"allowedStrings": []interface{}{"foo"}},
					"U": map[string]interface{}{},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpressionInElement", Message: `Missing JSX expression container around literal string: "foo" in U`},
			},
		},
		{
			Code: `<T>foo<U>foo</U></T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{},
					"U": map[string]interface{}{"allowedStrings": []interface{}{"foo"}},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpressionInElement", Message: `Missing JSX expression container around literal string: "foo" in T`},
			},
		},
		{
			Code: `
<div>
  <Fragment>foo</Fragment>
  <React.Fragment>foo</React.Fragment>
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"React.Fragment": map[string]interface{}{"allowElement": true},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "foo"`},
			},
		},
		{
			Code: `<div>foo</div>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"div": map[string]interface{}{"allowElement": true},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "foo"`},
			},
		},
		{
			Code: `
<div>
  <div type="text" />
  <Button type="submit" />
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"Button": map[string]interface{}{"restrictedAttributes": []interface{}{"type"}},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "restrictedAttributeStringInElement", Message: `Restricted attribute string: ""submit"" in type of Button`},
			},
		},
		{
			Code: `
<div>
  <Input placeholder="Enter text" type="password" />
  <Button type="submit" disabled="true" />
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"Input":  map[string]interface{}{"restrictedAttributes": []interface{}{"placeholder"}},
					"Button": map[string]interface{}{"restrictedAttributes": []interface{}{"disabled"}},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "restrictedAttributeStringInElement", Message: `Restricted attribute string: ""Enter text"" in placeholder of Input`},
				{MessageId: "restrictedAttributeStringInElement", Message: `Restricted attribute string: ""true"" in disabled of Button`},
			},
		},
		{
			Code: `
<div>
  <div className="wrapper" id="main" />
  <Button className="btn" id="submit-btn" />
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"restrictedAttributes": []interface{}{"className"},
				"elementOverrides": map[string]interface{}{
					"Button": map[string]interface{}{"restrictedAttributes": []interface{}{"id"}},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "restrictedAttributeString", Message: `Restricted attribute string: ""wrapper"" in className`},
				{MessageId: "restrictedAttributeStringInElement", Message: `Restricted attribute string: ""submit-btn"" in id of Button`},
			},
		},
		{
			Code: `
<div>
  <div foo1="bar1" />
  <T foo2="bar2" />
</div>
`,
			Tsx: true,
			Options: map[string]interface{}{
				"noAttributeStrings": true,
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"restrictedAttributes": []interface{}{"foo2"}},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInAttributes", Message: `Strings not allowed in attributes: ""bar1""`},
				{MessageId: "restrictedAttributeStringInElement", Message: `Restricted attribute string: ""bar2"" in foo2 of T`},
			},
		},

		// ---- Position assertions per messageId variant ----
		// One small-source case per emit-site so future refactors of the
		// trimmed-range / report path can't silently shift columns.
		{
			Code: `<div>test</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpression", Message: `Missing JSX expression container around literal string: "test"`, Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
			},
		},
		{
			Code:    `<Foo bar="test"></Foo>`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropValue", Message: `Invalid prop value: "bar="test""`, Line: 1, Column: 6, EndLine: 1, EndColumn: 16},
			},
		},
		{
			Code:    `<div>{'foo'}</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'foo'"`, Line: 1, Column: 7, EndLine: 1, EndColumn: 12},
			},
		},
		{
			Code:    `<img alt='blank image' />`,
			Tsx:     true,
			Options: map[string]interface{}{"noAttributeStrings": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInAttributes", Message: `Strings not allowed in attributes: "'blank image'"`, Line: 1, Column: 10, EndLine: 1, EndColumn: 23},
			},
		},
		{
			Code:    `<div className="test" />`,
			Tsx:     true,
			Options: map[string]interface{}{"restrictedAttributes": []interface{}{"className"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "restrictedAttributeString", Message: `Restricted attribute string: ""test"" in className`, Line: 1, Column: 6, EndLine: 1, EndColumn: 22},
			},
		},
		{
			Code: `<T>foo</T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpressionInElement", Message: `Missing JSX expression container around literal string: "foo" in T`, Line: 1, Column: 4, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: `<T>{'foo'}</T>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSXInElement", Message: `Strings not allowed in JSX files: "'foo'" in T`, Line: 1, Column: 5, EndLine: 1, EndColumn: 10},
			},
		},
		{
			Code: `<T foo="bar" />`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noAttributeStrings": true},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInAttributesInElement", Message: `Strings not allowed in attributes: ""bar"" in T`, Line: 1, Column: 8, EndLine: 1, EndColumn: 13},
			},
		},
		{
			Code: `<T foo="bar" />`,
			Tsx:  true,
			Options: map[string]interface{}{
				"noStrings":   true,
				"ignoreProps": false,
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"noStrings": true},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropValueInElement", Message: `Invalid prop value: "foo="bar"" in T`, Line: 1, Column: 4, EndLine: 1, EndColumn: 13},
			},
		},
		{
			Code: `<T foo="bar" />`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{"restrictedAttributes": []interface{}{"foo"}},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "restrictedAttributeStringInElement", Message: `Restricted attribute string: ""bar"" in foo of T`, Line: 1, Column: 4, EndLine: 1, EndColumn: 13},
			},
		},

		// ---- tsgo edge cases ----
		// Paren-wrapped literal in JSX content fires under noStrings. Locks
		// in `skipBinaryAndParens` peeling the explicit ParenthesizedExpression
		// node that tsgo retains (vs. ESTree's parse-time drop).
		{
			Code:    `<div>{('foo')}</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'foo'"`},
			},
		},
		// Paren around a binary keeps both literals reachable.
		{
			Code:    `<Foo bar={('foo' + 'bar')} />`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'foo'"`},
				{MessageId: "noStringsInJSX", Message: `Strings not allowed in JSX files: "'bar'"`},
			},
		},
		// Namespaced attribute is invisible to restrictedAttributes (upstream
		// gates that branch on JsxIdentifier only) but the noStrings
		// fall-through DOES still cover it — emits invalidPropValue.
		{
			Code:    `<svg xlink:href="bad" />`,
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true, "ignoreProps": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropValue", Message: `Invalid prop value: "xlink:href="bad""`},
			},
		},
		// Nested templates: only the OUTER (parent = JsxExpression) reports;
		// the inner template's parent is TemplateSpan, outside the gate.
		{
			Code:    "<div>{`a ${`b`} c`}</div>",
			Tsx:     true,
			Options: map[string]interface{}{"noStrings": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStringsInJSX", Message: "Strings not allowed in JSX files: \"`a ${`b`} c`\""},
			},
		},
		// JsxFragment outer with override-target inner: the fragment is
		// invisible to jsxElementAncestors, so the T override still applies.
		{
			Code: `<><T>foo</T></>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"elementOverrides": map[string]interface{}{
					"T": map[string]interface{}{},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalNotInJSXExpressionInElement", Message: `Missing JSX expression container around literal string: "foo" in T`},
			},
		},
	})
}
