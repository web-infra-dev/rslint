import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectLiteralNotInJSXExpressionError1(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'literalNotInJSXExpression',
    'Missing JSX expression container around literal string: "test"',
    range,
  );
}

function expectInvalidPropValueError2(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'invalidPropValue',
    'Invalid prop value: "bar="test""',
    range,
  );
}

function expectNoStringsInJSXError3(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'noStringsInJSX',
    'Strings not allowed in JSX files: "\'Test\'"',
    range,
  );
}

function expectNoStringsInJSXError4(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'noStringsInJSX',
    'Strings not allowed in JSX files: "`Test`"',
    range,
  );
}

function expectNoStringsInAttributesError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'noStringsInAttributes',
    'Strings not allowed in attributes: ""foo bar""',
    range,
  );
}

function expectLiteralNotInJSXExpressionError6(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'literalNotInJSXExpression',
    'Missing JSX expression container around literal string: "baz bob"',
    range,
  );
}

function expectRestrictedAttributeStringError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'restrictedAttributeString',
    'Restricted attribute string: ""test"" in className',
    range,
  );
}

function expectLiteralNotInJSXExpressionError8(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'literalNotInJSXExpression',
    'Missing JSX expression container around literal string: "foo"',
    range,
  );
}

function expectNoStringsInJSXInElementError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'noStringsInJSXInElement',
    'Strings not allowed in JSX files: "\'bar\'" in T',
    range,
  );
}

function expectNoStringsInAttributesInElementError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'noStringsInAttributesInElement',
    'Strings not allowed in attributes: ""bar2"" in T',
    range,
  );
}

function expectNoStringsInJSXError11(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'noStringsInJSX',
    'Strings not allowed in JSX files: "\'foo\'"',
    range,
  );
}

function expectInvalidPropValueError12(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'invalidPropValue',
    'Invalid prop value: "foo1="bar""',
    range,
  );
}

function expectLiteralNotInJSXExpressionInElementError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'literalNotInJSXExpressionInElement',
    'Missing JSX expression container around literal string: "foo" in T',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-no-literals', {
  valid: [
    // ---- noStrings + allowedStrings allow specific attribute strings ----
    {
      code: `
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
      options: [{ noStrings: true, allowedStrings: ['button', 'submit'] }],
    },

    // ---- Default config: literals inside JSXExpressionContainer pass ----
    {
      code: `
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
    },
    {
      code: `
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
    },
    { code: `<div>{'test'}</div>` },
    { code: `var foo = require('foo');` },
    { code: `<Foo bar='test'>{'blarg'}</Foo>` },
    {
      code: `<Foo bar="test">{intl.formatText(message)}</Foo>`,
      options: [{ noStrings: true, ignoreProps: true }],
    },
    { code: `<Foo bar={true} />`, options: [{ noStrings: true }] },
    { code: `<Foo bar={false} />`, options: [{ noStrings: true }] },
    { code: `<Foo bar={100} />`, options: [{ noStrings: true }] },
    { code: `<Foo bar={null} />`, options: [{ noStrings: true }] },
    { code: `<Foo bar={{}} />`, options: [{ noStrings: true }] },

    // ---- allowedStrings ----
    {
      code: `<div>asdf</div>`,
      options: [{ allowedStrings: ['asdf'] }],
    },
    {
      code: `<div>&nbsp;</div>`,
      options: [{ noStrings: true, allowedStrings: ['&nbsp;'] }],
    },
    {
      code: `<div>foo</div>`,
      options: [{ noStrings: true, allowedStrings: ['   foo   '] }],
    },

    // ---- Default: attribute strings allowed ----
    { code: `<img alt='blank image'></img>` },

    // ---- restrictedAttributes ----
    {
      code: `<img src="image.jpg" alt="text" />`,
      options: [{ restrictedAttributes: ['className', 'id'] }],
    },
    {
      code: `<div className="allowed" />`,
      options: [
        { restrictedAttributes: ['className'], allowedStrings: ['allowed'] },
      ],
    },
    {
      code: `<div className="test" id="foo" />`,
      options: [{ restrictedAttributes: [] }],
    },

    // ---- elementOverrides ----
    {
      code: `<T>foo</T>`,
      options: [{ elementOverrides: { T: { allowElement: true } } }],
    },
    {
      code: `<T>foo <div>bar</div></T>`,
      options: [{ elementOverrides: { T: { allowElement: true } } }],
    },
    {
      code: `<T>foo <div>{'bar'}</div></T>`,
      options: [
        {
          elementOverrides: {
            T: { allowElement: true, applyToNestedElements: false },
          },
        },
      ],
    },
    {
      code: `<T>{2}<div>{2}</div></T>`,
      options: [{ elementOverrides: { T: { noStrings: true } } }],
    },
    {
      code: `<T>foo<div>foo</div></T>`,
      options: [{ elementOverrides: { T: { allowedStrings: ['foo'] } } }],
    },
    {
      code: `<T foo="bar"><div foo="bar" /></T>`,
      options: [
        {
          noStrings: true,
          elementOverrides: { T: { noStrings: true, ignoreProps: true } },
        },
      ],
    },
    {
      code: `<T foo={2}><div foo={2} /></T>`,
      options: [{ elementOverrides: { T: { noAttributeStrings: true } } }],
    },
    {
      code: `<T>foo<U>foo</U></T>`,
      options: [
        {
          elementOverrides: {
            T: { allowedStrings: ['foo'] },
            U: { allowedStrings: ['foo'] },
          },
        },
      ],
    },
    {
      code: `
import { T } from 'foo';
<T>{'foo'}</T>
`,
    },
    {
      code: `
import { T as U } from 'foo';
<U>foo</U>
`,
      options: [{ elementOverrides: { T: { allowElement: true } } }],
    },
    {
      code: `
const { T: U } = require('foo');
<U>foo</U>
`,
      options: [{ elementOverrides: { T: { allowElement: true } } }],
    },
    {
      code: `<T.U>foo</T.U>`,
      options: [{ elementOverrides: { 'T.U': { allowElement: true } } }],
    },
    {
      code: `<React.Fragment>foo</React.Fragment>`,
      options: [
        { elementOverrides: { 'React.Fragment': { allowElement: true } } },
      ],
    },
    {
      code: `<div>{'foo'}</div>`,
      options: [{ elementOverrides: { div: { allowElement: true } } }],
    },
    {
      code: `
<div>
  <Input type="text" />
  <Button className="primary" />
  <Image src="photo.jpg" />
</div>
`,
      options: [
        {
          elementOverrides: {
            Input: { restrictedAttributes: ['placeholder'] },
            Button: { restrictedAttributes: ['type'] },
          },
        },
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
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
      options: [{ noStrings: true, allowedStrings: ['button', 'submit'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return (<div>{'test'}</div>);
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            const bar = (<div>{'hello'}</div>);
            return bar;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        var Hello = createReactClass({
          foo: (<div>{'hello'}</div>),
          render() {
            return this.foo;
          },
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return (
              <div>
              </div>
            );
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        var foo = require('foo');
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        <Foo bar='test'>
          {'blarg'}
        </Foo>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        <Foo bar="test">
          {intl.formatText(message)}
        </Foo>
      `,
      options: [{ noStrings: true, ignoreProps: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        <Foo bar="test">
          {translate('my.translate.key')}
        </Foo>
      `,
      options: [{ noStrings: true, ignoreProps: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        class Comp1 extends Component {
          asdf() {}
          render() {
            return <Foo bar={this.asdf} class='xx' />;
          }
        }
      `,
      options: [{ noStrings: true, ignoreProps: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            let foo = \`bar\`;
            return <div />;
          }
        }
      `,
      options: [{ noStrings: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return <div>asdf</div>
          }
        }
      `,
      options: [{ allowedStrings: ['asdf'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return <div>asdf</div>
          }
        }
      `,
      options: [{ noStrings: false, allowedStrings: ['asdf'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return <div>&nbsp;</div>
          }
        }
      `,
      options: [{ noStrings: true, allowedStrings: ['&nbsp;'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
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
      options: [{ noStrings: true, allowedStrings: ['&nbsp;'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return <div>foo: {bar}*</div>
          }
        }
      `,
      options: [{ noStrings: true, allowedStrings: ['foo: ', '*'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return <div>foo</div>
          }
        }
      `,
      options: [{ noStrings: true, allowedStrings: ['   foo   '] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        class Comp1 extends Component {
          asdf() {}
          render() {
            const xx = 'xx';

            return <Foo bar={this.asdf} class={xx} />;
          }
        }
      `,
      options: [{ noStrings: true, ignoreProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        <img alt='blank image'></img>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        <div>&mdash;</div>
      `,
      options: [{ noStrings: true, allowedStrings: ['&mdash;', '—'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        <div>—</div>
      `,
      options: [{ noStrings: true, allowedStrings: ['&mdash;', '—'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        <T>foo</T>
      `,
      options: [{ elementOverrides: { T: { allowElement: true } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        <T>foo <div>bar</div></T>
      `,
      options: [{ elementOverrides: { T: { allowElement: true } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        <T>foo <div>{'bar'}</div></T>
      `,
      options: [
        {
          elementOverrides: {
            T: { allowElement: true, applyToNestedElements: false },
          },
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
        <div>
          <div>{'foo'}</div>
          <T>{2}</T>
        </div>
      `,
      options: [{ elementOverrides: { T: { noStrings: true } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        <T>{2}<div>{2}</div></T>
      `,
      options: [{ elementOverrides: { T: { noStrings: true } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
        <T>{2}<div>{'foo'}</div></T>
      `,
      options: [
        {
          elementOverrides: {
            T: { noStrings: true, applyToNestedElements: false },
          },
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
        <div>
          <div>{'foo'}</div>
          <T>foo</T>
        </div>
      `,
      options: [{ elementOverrides: { T: { allowedStrings: ['foo'] } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `
        <T>foo<div>foo</div></T>
      `,
      options: [{ elementOverrides: { T: { allowedStrings: ['foo'] } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
      code: `
        <T>foo<div>{'foo'}</div></T>
      `,
      options: [
        {
          elementOverrides: {
            T: { allowedStrings: ['foo'], applyToNestedElements: false },
          },
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
      code: `
        <div>
          <div foo={2} />
          <T foo="bar" />
        </div>
      `,
      options: [
        {
          noStrings: true,
          elementOverrides: { T: { noStrings: true, ignoreProps: true } },
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 39 (7.37.5)',
      code: `
        <T foo="bar"><div foo="bar" /></T>
      `,
      options: [
        {
          noStrings: true,
          elementOverrides: { T: { noStrings: true, ignoreProps: true } },
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 40 (7.37.5)',
      code: `
        <T foo="bar"><div foo={2} /></T>
      `,
      options: [
        {
          noStrings: true,
          elementOverrides: {
            T: {
              noStrings: true,
              ignoreProps: true,
              applyToNestedElements: false,
            },
          },
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 41 (7.37.5)',
      code: `
        <div>
          <div foo="foo" />
          <T foo={2} />
        </div>
      `,
      options: [{ elementOverrides: { T: { noAttributeStrings: true } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 42 (7.37.5)',
      code: `
        <T foo={2}><div foo={2} /></T>
      `,
      options: [{ elementOverrides: { T: { noAttributeStrings: true } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 43 (7.37.5)',
      code: `
        <T foo={2}><div foo="foo" /></T>
      `,
      options: [
        {
          elementOverrides: {
            T: { noAttributeStrings: true, applyToNestedElements: false },
          },
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 44 (7.37.5)',
      code: `
        <T>foo<U>foo</U></T>
      `,
      options: [
        {
          elementOverrides: {
            T: { allowedStrings: ['foo'] },
            U: { allowedStrings: ['foo'] },
          },
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 45 (7.37.5)',
      code: `
        import { T } from 'foo';
        <T>{'foo'}</T>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 46 (7.37.5)',
      code: `
        import { T as U } from 'foo';
        <U>foo</U>
      `,
      options: [{ elementOverrides: { T: { allowElement: true } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 47 (7.37.5)',
      code: `
        const { T: U } = require('foo');
        <U>foo</U>
      `,
      options: [{ elementOverrides: { T: { allowElement: true } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 48 (7.37.5)',
      code: `
        const { T: U } = require('foo').Foo;
        <U>foo</U>
      `,
      options: [{ elementOverrides: { T: { allowElement: true } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 49 (7.37.5)',
      code: `
        const { T: U } = require('foo').Foo.Foo;
        <U>foo</U>
      `,
      options: [{ elementOverrides: { T: { allowElement: true } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 50 (7.37.5)',
      code: `
        const foo = 2;
        <T>foo</T>
      `,
      options: [{ elementOverrides: { T: { allowElement: true } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 51 (7.37.5)',
      code: `
        <T.U>foo</T.U>
      `,
      options: [{ elementOverrides: { 'T.U': { allowElement: true } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 52 (7.37.5)',
      code: `
        import { T as U } from 'foo';
        <U.U>foo</U.U>
      `,
      options: [{ elementOverrides: { 'T.U': { allowElement: true } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 53 (7.37.5)',
      code: `
        <React.Fragment>foo</React.Fragment>
      `,
      options: [{ elementOverrides: { Fragment: { allowElement: true } } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 54 (7.37.5)',
      code: `
        <React.Fragment>foo</React.Fragment>
      `,
      options: [
        { elementOverrides: { 'React.Fragment': { allowElement: true } } },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 55 (7.37.5)',
      code: `
        <div>{'foo'}</div>
      `,
      options: [{ elementOverrides: { div: { allowElement: true } } }],
      languageOptions: upstreamModuleOptions,
    },
  ],

  invalid: [
    // ---- Default: bare JSX text reports literalNotInJSXExpression ----
    {
      code: `<div>test</div>`,
      errors: [expectLiteralNotInJSXExpressionError1([1, 6, 1, 10])],
    },
    {
      code: `<>test</>`,
      errors: [expectLiteralNotInJSXExpressionError1([1, 3, 1, 7])],
    },
    {
      code: `
<div>
  asdjfl
</div>
`,
      errors: [
        diagnostic(
          'literalNotInJSXExpression',
          'Missing JSX expression container around literal string: "asdjfl"',
          [3, 3, 3, 9],
        ),
      ],
    },

    // ---- noStrings: attribute literal + JSX child literal ----
    {
      code: `<Foo bar="test">{'Test'}</Foo>`,
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [
        expectInvalidPropValueError2([1, 6, 1, 16]),
        expectNoStringsInJSXError3([1, 18, 1, 24]),
      ],
    },
    {
      code: `<Foo bar="test">Test</Foo>`,
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [
        expectInvalidPropValueError2([1, 6, 1, 16]),
        diagnostic(
          'noStringsInJSX',
          'Strings not allowed in JSX files: "Test"',
          [1, 17, 1, 21],
        ),
      ],
    },
    // ---- TemplateLiteral handler ----
    {
      code: '<Foo>{`Test`}</Foo>',
      options: [{ noStrings: true }],
      errors: [expectNoStringsInJSXError4([1, 7, 1, 13])],
    },
    {
      code: '<Foo bar={`Test`} />',
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [expectNoStringsInJSXError4([1, 11, 1, 17])],
    },
    {
      code: '<Foo bar={`Test ${baz}`} />',
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [
        diagnostic(
          'noStringsInJSX',
          'Strings not allowed in JSX files: "`Test ${baz}`"',
          [1, 11, 1, 24],
        ),
      ],
    },

    // ---- noAttributeStrings ----
    {
      code: `<img alt='blank image'></img>`,
      options: [{ noAttributeStrings: true }],
      errors: [
        diagnostic(
          'noStringsInAttributes',
          'Strings not allowed in attributes: "\'blank image\'"',
          [1, 10, 1, 23],
        ),
      ],
    },
    {
      code: `export const WithAttributes = ({}) => <div title="foo bar" />;`,
      options: [{ noAttributeStrings: true }],
      errors: [expectNoStringsInAttributesError([1, 50, 1, 59])],
    },
    {
      code: `
export const WithAttributesAndChildren = ({}) => (
  <div title="foo bar">baz bob</div>
);
`,
      options: [{ noAttributeStrings: true }],
      errors: [
        expectNoStringsInAttributesError([3, 14, 3, 23]),
        expectLiteralNotInJSXExpressionError6([3, 24, 3, 31]),
      ],
    },

    // ---- restrictedAttributes ----
    {
      code: `<div className="test" />`,
      options: [{ restrictedAttributes: ['className'] }],
      errors: [expectRestrictedAttributeStringError([1, 6, 1, 22])],
    },
    {
      code: `<div className="test" id="foo" title="bar" />`,
      options: [{ restrictedAttributes: ['className', 'id'] }],
      errors: [
        expectRestrictedAttributeStringError([1, 6, 1, 22]),
        diagnostic(
          'restrictedAttributeString',
          'Restricted attribute string: ""foo"" in id',
          [1, 23, 1, 31],
        ),
      ],
    },
    {
      code: `<div title="text">test</div>`,
      options: [{ restrictedAttributes: ['title'], noStrings: true }],
      errors: [
        diagnostic(
          'restrictedAttributeString',
          'Restricted attribute string: ""text"" in title',
          [1, 6, 1, 18],
        ),
        diagnostic(
          'noStringsInJSX',
          'Strings not allowed in JSX files: "test"',
          [1, 19, 1, 23],
        ),
      ],
    },
    {
      code: `<div className="test" title="hello" />`,
      options: [
        {
          noStrings: true,
          ignoreProps: false,
          restrictedAttributes: ['className'],
        },
      ],
      errors: [
        expectRestrictedAttributeStringError([1, 6, 1, 22]),
        diagnostic(
          'invalidPropValue',
          'Invalid prop value: "title="hello""',
          [1, 23, 1, 36],
        ),
      ],
    },

    // ---- elementOverrides ----
    {
      code: `
<div>
  <div>foo</div>
  <T>bar</T>
</div>
`,
      options: [{ elementOverrides: { T: {} } }],
      errors: [
        expectLiteralNotInJSXExpressionError8([3, 8, 3, 11]),
        diagnostic(
          'literalNotInJSXExpressionInElement',
          'Missing JSX expression container around literal string: "bar" in T',
          [4, 6, 4, 9],
        ),
      ],
    },
    {
      code: `
<div>
  <div>foo</div>
  <T>bar</T>
</div>
`,
      options: [{ elementOverrides: { T: { allowElement: true } } }],
      errors: [expectLiteralNotInJSXExpressionError8([3, 8, 3, 11])],
    },
    {
      code: `
<div>
  <div>foo</div>
  <T>{'bar'}</T>
</div>
`,
      options: [{ elementOverrides: { T: { noStrings: true } } }],
      errors: [
        expectLiteralNotInJSXExpressionError8([3, 8, 3, 11]),
        expectNoStringsInJSXInElementError([4, 7, 4, 12]),
      ],
    },
    {
      code: `
<div>
  <div foo1="bar1" />
  <T foo2="bar2" />
</div>
`,
      options: [{ elementOverrides: { T: { noAttributeStrings: true } } }],
      errors: [expectNoStringsInAttributesInElementError([4, 11, 4, 17])],
    },
    {
      code: `
<div>
  <div type="text" />
  <Button type="submit" />
</div>
`,
      options: [
        {
          elementOverrides: {
            Button: { restrictedAttributes: ['type'] },
          },
        },
      ],
      errors: [
        diagnostic(
          'restrictedAttributeStringInElement',
          'Restricted attribute string: ""submit"" in type of Button',
          [4, 11, 4, 24],
        ),
      ],
    },
    {
      code: `
<div>
  <div className="wrapper" id="main" />
  <Button className="btn" id="submit-btn" />
</div>
`,
      options: [
        {
          restrictedAttributes: ['className'],
          elementOverrides: {
            Button: { restrictedAttributes: ['id'] },
          },
        },
      ],
      errors: [
        diagnostic(
          'restrictedAttributeString',
          'Restricted attribute string: ""wrapper"" in className',
          [3, 8, 3, 27],
        ),
        diagnostic(
          'restrictedAttributeStringInElement',
          'Restricted attribute string: ""submit-btn"" in id of Button',
          [4, 27, 4, 42],
        ),
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return (<div>test</div>);
          }
        }
      `,
      errors: [expectLiteralNotInJSXExpressionError1([4, 26, 4, 30])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return (<>test</>);
          }
        }
      `,
      errors: [expectLiteralNotInJSXExpressionError1([4, 23, 4, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            const foo = (<div>test</div>);
            return foo;
          }
        }
      `,
      errors: [expectLiteralNotInJSXExpressionError1([4, 31, 4, 35])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            const varObjectTest = { testKey : (<div>test</div>) };
            return varObjectTest.testKey;
          }
        }
      `,
      errors: [expectLiteralNotInJSXExpressionError1([4, 53, 4, 57])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        var Hello = createReactClass({
          foo: (<div>hello</div>),
          render() {
            return this.foo;
          },
        });
      `,
      errors: [
        diagnostic(
          'literalNotInJSXExpression',
          'Missing JSX expression container around literal string: "hello"',
          [3, 22, 3, 27],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
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
      errors: [
        diagnostic(
          'literalNotInJSXExpression',
          'Missing JSX expression container around literal string: "asdjfl"',
          [6, 17, 6, 23],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
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
      errors: [
        diagnostic(
          'literalNotInJSXExpression',
          'Missing JSX expression container around literal string: "asdjfl\n                test\n                foo"',
          [6, 17, 8, 20],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
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
      errors: [expectLiteralNotInJSXExpressionError1([7, 17, 7, 21])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        <Foo bar="test">
          {'Test'}
        </Foo>
      `,
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [
        expectInvalidPropValueError2([2, 14, 2, 24]),
        expectNoStringsInJSXError3([3, 12, 3, 18]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        <Foo bar="test">
          {'Test' + name}
        </Foo>
      `,
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [
        expectInvalidPropValueError2([2, 14, 2, 24]),
        expectNoStringsInJSXError3([3, 12, 3, 18]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        <Foo bar="test">
          Test
        </Foo>
      `,
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [
        expectInvalidPropValueError2([2, 14, 2, 24]),
        diagnostic(
          'noStringsInJSX',
          'Strings not allowed in JSX files: "Test"',
          [3, 11, 3, 15],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        <Foo>
          {\`Test\`}
        </Foo>
      `,
      options: [{ noStrings: true }],
      errors: [expectNoStringsInJSXError4([3, 12, 3, 18])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `<Foo bar={\`\${baz}\`} />`,
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [
        diagnostic(
          'noStringsInJSX',
          'Strings not allowed in JSX files: "`${baz}`"',
          [1, 11, 1, 19],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `<Foo bar={\`foo\` + 'bar'} />`,
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [
        diagnostic(
          'noStringsInJSX',
          'Strings not allowed in JSX files: "`foo`"',
          [1, 11, 1, 16],
        ),
        diagnostic(
          'noStringsInJSX',
          'Strings not allowed in JSX files: "\'bar\'"',
          [1, 19, 1, 24],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `<Foo bar={\`foo\` + \`bar\`} />`,
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [
        diagnostic(
          'noStringsInJSX',
          'Strings not allowed in JSX files: "`foo`"',
          [1, 11, 1, 16],
        ),
        diagnostic(
          'noStringsInJSX',
          'Strings not allowed in JSX files: "`bar`"',
          [1, 19, 1, 24],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `<Foo bar={'foo' + \`bar\`} />`,
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [
        expectNoStringsInJSXError11([1, 11, 1, 16]),
        diagnostic(
          'noStringsInJSX',
          'Strings not allowed in JSX files: "`bar`"',
          [1, 19, 1, 24],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return <div bar={'foo'}>asdf</div>
          }
        }
      `,
      options: [
        { noStrings: true, allowedStrings: ['asd'], ignoreProps: false },
      ],
      errors: [
        expectNoStringsInJSXError11([4, 30, 4, 35]),
        diagnostic(
          'noStringsInJSX',
          'Strings not allowed in JSX files: "asdf"',
          [4, 37, 4, 41],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `<Foo bar={'bar'} />`,
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [
        diagnostic(
          'noStringsInJSX',
          'Strings not allowed in JSX files: "\'bar\'"',
          [1, 11, 1, 16],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        <img alt='blank image'></img>
      `,
      options: [{ noAttributeStrings: true }],
      errors: [
        diagnostic(
          'noStringsInAttributes',
          'Strings not allowed in attributes: "\'blank image\'"',
          [2, 18, 2, 31],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `export const WithChildren = ({}) => <div>baz bob</div>;`,
      options: [{ noAttributeStrings: true }],
      errors: [expectLiteralNotInJSXExpressionError6([1, 42, 1, 49])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        export const WithAttributesAndChildren = ({}) => (
          <div title="foo bar">baz bob</div>
        );
      `,
      options: [{ noAttributeStrings: true }],
      errors: [
        expectNoStringsInAttributesError([3, 22, 3, 31]),
        expectLiteralNotInJSXExpressionError6([3, 32, 3, 39]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
        <div>
          <div>foo</div>
          <T>bar</T>
        </div>
      `,
      options: [{ elementOverrides: { T: {} } }],
      errors: [
        expectLiteralNotInJSXExpressionError8([3, 16, 3, 19]),
        diagnostic(
          'literalNotInJSXExpressionInElement',
          'Missing JSX expression container around literal string: "bar" in T',
          [4, 14, 4, 17],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
        <div>
          <div>foo</div>
          <T>bar</T>
        </div>
      `,
      options: [{ elementOverrides: { T: { allowElement: true } } }],
      errors: [expectLiteralNotInJSXExpressionError8([3, 16, 3, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
        <T>foo <div>bar</div></T>
      `,
      options: [
        {
          elementOverrides: {
            T: { allowElement: true, applyToNestedElements: false },
          },
        },
      ],
      errors: [
        diagnostic(
          'literalNotInJSXExpression',
          'Missing JSX expression container around literal string: "bar"',
          [2, 21, 2, 24],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
        <div>
          <div>foo</div>
          <T>{'bar'}</T>
        </div>
      `,
      options: [{ elementOverrides: { T: { noStrings: true } } }],
      errors: [
        expectLiteralNotInJSXExpressionError8([3, 16, 3, 19]),
        expectNoStringsInJSXInElementError([4, 15, 4, 20]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `
        <div>
          <div>foo</div>
          <T>{'bar'}<div>{'baz'}</div></T>
        </div>
      `,
      options: [{ elementOverrides: { T: { noStrings: true } } }],
      errors: [
        expectLiteralNotInJSXExpressionError8([3, 16, 3, 19]),
        expectNoStringsInJSXInElementError([4, 15, 4, 20]),
        diagnostic(
          'noStringsInJSXInElement',
          'Strings not allowed in JSX files: "\'baz\'" in T',
          [4, 27, 4, 32],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `
        <div>
          <div>foo</div>
          <T>{'bar'}<div>{'baz'}</div></T>
        </div>
      `,
      options: [
        {
          elementOverrides: {
            T: { noStrings: true, applyToNestedElements: false },
          },
        },
      ],
      errors: [
        expectLiteralNotInJSXExpressionError8([3, 16, 3, 19]),
        expectNoStringsInJSXInElementError([4, 15, 4, 20]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `
        <div>
          <div>{'foo'}</div>
          <T>{'foo'}</T>
        </div>
      `,
      options: [
        {
          noStrings: true,
          elementOverrides: { T: { noStrings: true, allowedStrings: ['foo'] } },
        },
      ],
      errors: [expectNoStringsInJSXError11([3, 17, 3, 22])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `
        <div>
          <div>{'foo'}</div>
          <T>{'foo'}<div>{'foo'}</div></T>
        </div>
      `,
      options: [
        {
          noStrings: true,
          elementOverrides: { T: { noStrings: true, allowedStrings: ['foo'] } },
        },
      ],
      errors: [expectNoStringsInJSXError11([3, 17, 3, 22])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `
        <div>
          <div>{'foo'}</div>
          <T>{'foo'}<div>{'foo'}</div></T>
        </div>
      `,
      options: [
        {
          noStrings: true,
          elementOverrides: {
            T: {
              noStrings: true,
              allowedStrings: ['foo'],
              applyToNestedElements: false,
            },
          },
        },
      ],
      errors: [
        expectNoStringsInJSXError11([3, 17, 3, 22]),
        expectNoStringsInJSXError11([4, 27, 4, 32]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `
        <div>
          <div foo1="bar" />
          <T foo2="bar" />
        </div>
      `,
      options: [
        {
          noStrings: true,
          elementOverrides: { T: { noStrings: true, ignoreProps: true } },
        },
      ],
      errors: [expectInvalidPropValueError12([3, 16, 3, 26])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
        <div>
          <div foo1="bar" />
          <T foo2="bar"><div foo3="bar" /></T>
        </div>
      `,
      options: [
        {
          noStrings: true,
          elementOverrides: { T: { noStrings: true, ignoreProps: true } },
        },
      ],
      errors: [expectInvalidPropValueError12([3, 16, 3, 26])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
        <div>
          <div foo1="bar" />
          <T foo2="bar"><div foo3="bar" /></T>
        </div>
      `,
      options: [
        {
          noStrings: true,
          elementOverrides: {
            T: {
              noStrings: true,
              ignoreProps: true,
              applyToNestedElements: false,
            },
          },
        },
      ],
      errors: [
        expectInvalidPropValueError12([3, 16, 3, 26]),
        diagnostic(
          'invalidPropValue',
          'Invalid prop value: "foo3="bar""',
          [4, 30, 4, 40],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `
        <div>
          <div foo1="bar1" />
          <T foo2="bar2" />
        </div>
      `,
      options: [{ elementOverrides: { T: { noAttributeStrings: true } } }],
      errors: [expectNoStringsInAttributesInElementError([4, 19, 4, 25])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
      code: `
        <div>
          <div foo1="bar1" />
          <T foo2="bar2"><div foo3="bar3" /></T>
        </div>
      `,
      options: [{ elementOverrides: { T: { noAttributeStrings: true } } }],
      errors: [
        expectNoStringsInAttributesInElementError([4, 19, 4, 25]),
        diagnostic(
          'noStringsInAttributesInElement',
          'Strings not allowed in attributes: ""bar3"" in T',
          [4, 36, 4, 42],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `
        <div>
          <div foo1="bar1" />
          <T foo2="bar2"><div foo3="bar3" /></T>
        </div>
      `,
      options: [
        {
          elementOverrides: {
            T: { noAttributeStrings: true, applyToNestedElements: false },
          },
        },
      ],
      errors: [expectNoStringsInAttributesInElementError([4, 19, 4, 25])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `
        <div>
          <div>{'foo'}</div>
          <T>{'bar'}</T>
        </div>
      `,
      options: [{ noStrings: true, elementOverrides: { T: {} } }],
      errors: [expectNoStringsInJSXError11([3, 17, 3, 22])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `
        <div>
          <div>foo</div>
          <T>foo</T>
        </div>
      `,
      options: [{ allowedStrings: ['foo'], elementOverrides: { T: {} } }],
      errors: [expectLiteralNotInJSXExpressionInElementError([4, 14, 4, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 41 (7.37.5)',
      code: `
        <div>
          <div>foo</div>
          <T>foo</T>
          <T>bar</T>
          <T>baz</T>
        </div>
      `,
      options: [
        {
          allowedStrings: ['foo'],
          elementOverrides: { T: { allowedStrings: ['bar'] } },
        },
      ],
      errors: [
        expectLiteralNotInJSXExpressionInElementError([4, 14, 4, 17]),
        diagnostic(
          'literalNotInJSXExpressionInElement',
          'Missing JSX expression container around literal string: "baz" in T',
          [6, 14, 6, 17],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 42 (7.37.5)',
      code: `
        <div>
          <div foo1="bar1" />
          <T foo2="bar2" />
        </div>
      `,
      options: [
        {
          noStrings: true,
          ignoreProps: true,
          elementOverrides: { T: { noStrings: true } },
        },
      ],
      errors: [
        diagnostic(
          'invalidPropValueInElement',
          'Invalid prop value: "foo2="bar2"" in T',
          [4, 14, 4, 25],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 43 (7.37.5)',
      code: `
        <div>
          <div foo1="bar1" />
          <T foo2="bar2" />
        </div>
      `,
      options: [{ noAttributeStrings: true, elementOverrides: { T: {} } }],
      errors: [
        diagnostic(
          'noStringsInAttributes',
          'Strings not allowed in attributes: ""bar1""',
          [3, 21, 3, 27],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 44 (7.37.5)',
      code: `
        <div>
          <T>foo</T>
          <U>bar</U>
        </div>
      `,
      options: [{ elementOverrides: { T: {}, U: {} } }],
      errors: [
        expectLiteralNotInJSXExpressionInElementError([3, 14, 3, 17]),
        diagnostic(
          'literalNotInJSXExpressionInElement',
          'Missing JSX expression container around literal string: "bar" in U',
          [4, 14, 4, 17],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 45 (7.37.5)',
      code: `
        <div>
          <T>foo</T>
          <U>bar</U>
        </div>
      `,
      options: [{ elementOverrides: { T: {}, U: { allowElement: true } } }],
      errors: [expectLiteralNotInJSXExpressionInElementError([3, 14, 3, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 46 (7.37.5)',
      code: `
        <T>foo <U>bar</U></T>
      `,
      options: [{ elementOverrides: { T: {}, U: { allowElement: true } } }],
      errors: [expectLiteralNotInJSXExpressionInElementError([2, 12, 2, 15])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 47 (7.37.5)',
      code: `
        <T>{'foo'}<U>{'bar'}</U></T>
      `,
      options: [{ elementOverrides: { T: { noStrings: true }, U: {} } }],
      errors: [
        diagnostic(
          'noStringsInJSXInElement',
          'Strings not allowed in JSX files: "\'foo\'" in T',
          [2, 13, 2, 18],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 48 (7.37.5)',
      code: `
        <T>foo<U>foo</U></T>
      `,
      options: [
        { elementOverrides: { T: { allowedStrings: ['foo'] }, U: {} } },
      ],
      errors: [
        diagnostic(
          'literalNotInJSXExpressionInElement',
          'Missing JSX expression container around literal string: "foo" in U',
          [2, 18, 2, 21],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 49 (7.37.5)',
      code: `
        <T>foo<U>foo</U></T>
      `,
      options: [
        { elementOverrides: { T: {}, U: { allowedStrings: ['foo'] } } },
      ],
      errors: [expectLiteralNotInJSXExpressionInElementError([2, 12, 2, 15])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 50 (7.37.5)',
      code: `
        <div>
          <Fragment>foo</Fragment>
          <React.Fragment>foo</React.Fragment>
        </div>
      `,
      options: [
        { elementOverrides: { 'React.Fragment': { allowElement: true } } },
      ],
      errors: [expectLiteralNotInJSXExpressionError8([3, 21, 3, 24])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 51 (7.37.5)',
      code: `
        <div>foo</div>
      `,
      options: [{ elementOverrides: { div: { allowElement: true } } }],
      errors: [expectLiteralNotInJSXExpressionError8([2, 14, 2, 17])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
