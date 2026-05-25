import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-no-literals', {} as never, {
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
  ],

  invalid: [
    // ---- Default: bare JSX text reports literalNotInJSXExpression ----
    {
      code: `<div>test</div>`,
      errors: [{ messageId: 'literalNotInJSXExpression' }],
    },
    {
      code: `<>test</>`,
      errors: [{ messageId: 'literalNotInJSXExpression' }],
    },
    {
      code: `
<div>
  asdjfl
</div>
`,
      errors: [{ messageId: 'literalNotInJSXExpression' }],
    },

    // ---- noStrings: attribute literal + JSX child literal ----
    {
      code: `<Foo bar="test">{'Test'}</Foo>`,
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [
        { messageId: 'invalidPropValue' },
        { messageId: 'noStringsInJSX' },
      ],
    },
    {
      code: `<Foo bar="test">Test</Foo>`,
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [
        { messageId: 'invalidPropValue' },
        { messageId: 'noStringsInJSX' },
      ],
    },
    // ---- TemplateLiteral handler ----
    {
      code: '<Foo>{`Test`}</Foo>',
      options: [{ noStrings: true }],
      errors: [{ messageId: 'noStringsInJSX' }],
    },
    {
      code: '<Foo bar={`Test`} />',
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [{ messageId: 'noStringsInJSX' }],
    },
    {
      code: '<Foo bar={`Test ${baz}`} />',
      options: [{ noStrings: true, ignoreProps: false }],
      errors: [{ messageId: 'noStringsInJSX' }],
    },

    // ---- noAttributeStrings ----
    {
      code: `<img alt='blank image'></img>`,
      options: [{ noAttributeStrings: true }],
      errors: [{ messageId: 'noStringsInAttributes' }],
    },
    {
      code: `export const WithAttributes = ({}) => <div title="foo bar" />;`,
      options: [{ noAttributeStrings: true }],
      errors: [{ messageId: 'noStringsInAttributes' }],
    },
    {
      code: `
export const WithAttributesAndChildren = ({}) => (
  <div title="foo bar">baz bob</div>
);
`,
      options: [{ noAttributeStrings: true }],
      errors: [
        { messageId: 'noStringsInAttributes' },
        { messageId: 'literalNotInJSXExpression' },
      ],
    },

    // ---- restrictedAttributes ----
    {
      code: `<div className="test" />`,
      options: [{ restrictedAttributes: ['className'] }],
      errors: [{ messageId: 'restrictedAttributeString' }],
    },
    {
      code: `<div className="test" id="foo" title="bar" />`,
      options: [{ restrictedAttributes: ['className', 'id'] }],
      errors: [
        { messageId: 'restrictedAttributeString' },
        { messageId: 'restrictedAttributeString' },
      ],
    },
    {
      code: `<div title="text">test</div>`,
      options: [{ restrictedAttributes: ['title'], noStrings: true }],
      errors: [
        { messageId: 'restrictedAttributeString' },
        { messageId: 'noStringsInJSX' },
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
        { messageId: 'restrictedAttributeString' },
        { messageId: 'invalidPropValue' },
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
        { messageId: 'literalNotInJSXExpression' },
        { messageId: 'literalNotInJSXExpressionInElement' },
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
      errors: [{ messageId: 'literalNotInJSXExpression' }],
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
        { messageId: 'literalNotInJSXExpression' },
        { messageId: 'noStringsInJSXInElement' },
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
      errors: [{ messageId: 'noStringsInAttributesInElement' }],
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
      errors: [{ messageId: 'restrictedAttributeStringInElement' }],
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
        { messageId: 'restrictedAttributeString' },
        { messageId: 'restrictedAttributeStringInElement' },
      ],
    },
  ],
});
