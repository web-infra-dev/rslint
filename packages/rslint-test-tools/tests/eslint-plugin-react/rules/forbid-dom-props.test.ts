import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('forbid-dom-props', {} as never, {
  valid: [
    // ---- Upstream valid cases ----
    {
      code: `
        var First = createReactClass({
          render: function() {
            return <Foo id="foo" />;
          }
        });
      `,
      options: [{ forbid: ['id'] }],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo id="bar" style={{color: "red"}} />;
          }
        });
      `,
      options: [{ forbid: ['style', 'id'] }],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <this.Foo bar="baz" />;
          }
        });
      `,
      options: [{ forbid: ['id'] }],
    },
    {
      code: `
        class First extends createReactClass {
          render() {
            return <this.foo id="bar" />;
          }
        }
      `,
      options: [{ forbid: ['id'] }],
    },
    {
      code: `
        const First = (props) => (
          <this.Foo {...props} />
        );
      `,
      options: [{ forbid: ['id'] }],
    },
    {
      code: `
        const First = (props) => (
          <fbt:param name="name">{props.name}</fbt:param>
        );
      `,
      options: [{ forbid: ['id'] }],
    },
    {
      code: `
        const First = (props) => (
          <div name="foo" />
        );
      `,
      options: [{ forbid: ['id'] }],
    },
    {
      code: `
        const First = (props) => (
          <div otherProp="bar" />
        );
      `,
      options: [
        {
          forbid: [{ propName: 'otherProp', disallowedFor: ['span'] }],
        },
      ],
    },
    {
      code: `
        const First = (props) => (
          <div someProp="someValue" />
        );
      `,
      options: [
        {
          forbid: [{ propName: 'someProp', disallowedValues: [] }],
        },
      ],
    },
    {
      code: `
        const First = (props) => (
          <Foo someProp="someValue" />
        );
      `,
      options: [
        {
          forbid: [{ propName: 'someProp', disallowedValues: ['someValue'] }],
        },
      ],
    },
    {
      code: `
        const First = (props) => (
          <div someProp="value" />
        );
      `,
      options: [
        {
          forbid: [{ propName: 'someProp', disallowedValues: ['someValue'] }],
        },
      ],
    },
    {
      code: `
        const First = (props) => (
          <div someProp="someValue" />
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'someProp',
              disallowedValues: ['someValue'],
              disallowedFor: ['span'],
            },
          ],
        },
      ],
    },

    // ---- Additional edge cases ----
    // No options → empty default → no diagnostic.
    { code: `<div id="x" className="y" />;` },
    // Explicit empty forbid → no diagnostic on a DOM intrinsic.
    { code: `<div id="x" />;`, options: [{ forbid: [] }] },
    // Spread-only DOM tag — no JsxAttribute.
    { code: `<div {...props} />;`, options: [{ forbid: ['id'] }] },
    // `<Foo.bar>` — member expression, skipped before the lowercase check.
    { code: `<Foo.bar id="x" />;`, options: [{ forbid: ['id'] }] },
    // Empty-string entries skipped; remaining real entry doesn't include `bar`.
    {
      code: `<div bar="x" />;`,
      options: [{ forbid: ['', 'id'] }],
    },
    // JSX namespaced ATTRIBUTE name (`xlink:href`) — upstream's
    // `node.name.name` returns a JSXIdentifier object, not a string, so
    // `forbid.get(<obj>)` never matches. The user-supplied `xlink:href`
    // string MUST NOT pair up with the namespaced attribute name. (Caught
    // a real bug during port — locks it in across the JS binary too.)
    {
      code: `<svg><use xlink:href="#icon" /></svg>;`,
      options: [{ forbid: ['xlink:href'] }],
    },
    // JSX namespaced TAG name (`<svg:path>`) — typeof tag !== 'string'
    // upstream → skipped, even with `svg:path` in forbid.
    {
      code: `<svg:path id="x" />;`,
      options: [{ forbid: ['id', 'svg:path'] }],
    },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <div id="bar" />;
          }
        });
      `,
      options: [{ forbid: ['id'] }],
      errors: [{ messageId: 'propIsForbidden' }],
    },
    {
      code: `
        class First extends createReactClass {
          render() {
            return <div id="bar" />;
          }
        }
      `,
      options: [{ forbid: ['id'] }],
      errors: [{ messageId: 'propIsForbidden' }],
    },
    {
      code: `
        const First = (props) => (
          <div id="foo" />
        );
      `,
      options: [{ forbid: ['id'] }],
      errors: [{ messageId: 'propIsForbidden' }],
    },
    {
      code: `
        const First = (props) => (
          <div className="foo" />
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              message: 'Please use class instead of ClassName',
            },
          ],
        },
      ],
      errors: [{ message: 'Please use class instead of ClassName' }],
    },
    {
      code: `
        const First = (props) => (
          <span otherProp="bar" />
        );
      `,
      options: [
        {
          forbid: [{ propName: 'otherProp', disallowedFor: ['span'] }],
        },
      ],
      errors: [{ messageId: 'propIsForbidden' }],
    },
    {
      code: `
        const First = (props) => (
          <div someProp="someValue" />
        );
      `,
      options: [
        {
          forbid: [{ propName: 'someProp', disallowedValues: ['someValue'] }],
        },
      ],
      errors: [{ messageId: 'propIsForbiddenWithValue' }],
    },
    {
      code: `
        const First = (props) => (
          <div className="foo">
            <div otherProp="bar" />
          </div>
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              message: 'Please use class instead of ClassName',
            },
            { propName: 'otherProp', message: 'Avoid using otherProp' },
          ],
        },
      ],
      errors: [
        { message: 'Please use class instead of ClassName' },
        { message: 'Avoid using otherProp' },
      ],
    },
    {
      code: `
        const First = (props) => (
          <div className="foo">
            <div otherProp="bar" />
          </div>
        );
      `,
      options: [
        {
          forbid: [
            { propName: 'className' },
            { propName: 'otherProp', message: 'Avoid using otherProp' },
          ],
        },
      ],
      errors: [
        { messageId: 'propIsForbidden' },
        { message: 'Avoid using otherProp' },
      ],
    },
    {
      code: `
        const First = (props) => (
          <form accept='file'>
            <input type="file" id="videoFile" accept="video/*" />
            <input type="hidden" name="fullname" />
          </form>
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'accept',
              disallowedFor: ['form'],
              message: 'Avoid using the accept attribute on <form>',
            },
          ],
        },
      ],
      errors: [{ message: 'Avoid using the accept attribute on <form>' }],
    },
    {
      code: `
        const First = (props) => (
          <div className="foo">
            <input className="boo" />
            <span className="foobar">Foobar</span>
            <div otherProp="bar" />
          </div>
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              disallowedFor: ['div', 'span'],
              message: 'Please use class instead of ClassName',
            },
            { propName: 'otherProp', message: 'Avoid using otherProp' },
          ],
        },
      ],
      errors: [
        { message: 'Please use class instead of ClassName' },
        { message: 'Please use class instead of ClassName' },
        { message: 'Avoid using otherProp' },
      ],
    },
    {
      code: `
        const First = (props) => (
          <div className="foo">
            <input className="boo" />
            <span className="foobar">Foobar</span>
            <div otherProp="bar" />
            <p thirdProp="foo" />
            <div thirdProp="baz" />
            <p thirdProp="bar" />
            <p thirdProp="baz" />
          </div>
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              disallowedFor: ['div', 'span'],
              message: 'Please use class instead of ClassName',
            },
            { propName: 'otherProp', message: 'Avoid using otherProp' },
            {
              propName: 'thirdProp',
              disallowedFor: ['p'],
              disallowedValues: ['bar', 'baz'],
              message: 'Do not use thirdProp with values bar and baz on p',
            },
          ],
        },
      ],
      errors: [
        { message: 'Please use class instead of ClassName' },
        { message: 'Please use class instead of ClassName' },
        { message: 'Avoid using otherProp' },
        { message: 'Do not use thirdProp with values bar and baz on p' },
        { message: 'Do not use thirdProp with values bar and baz on p' },
      ],
    },

    // ---- Additional lock-in cases ----
    // Spread + multiple forbidden named attrs on a DOM tag — each reports.
    {
      code: `<div {...rest} id="x" className="y" />;`,
      options: [{ forbid: ['id', 'className'] }],
      errors: [
        { messageId: 'propIsForbidden' },
        { messageId: 'propIsForbidden' },
      ],
    },
    // Real-world: `target="_blank"` security ban via disallowedValues.
    {
      code: `<a href="/x" target="_blank">click</a>;`,
      options: [
        {
          forbid: [
            {
              propName: 'target',
              disallowedFor: ['a'],
              disallowedValues: ['_blank'],
            },
          ],
        },
      ],
      errors: [{ messageId: 'propIsForbiddenWithValue' }],
    },
    // Same forbid name appearing as both string and object entry —
    // last-write-wins (the second/object entry's message overrides).
    {
      code: `<div id="x" />;`,
      options: [
        {
          forbid: ['id', { propName: 'id', message: 'Object wins' }],
        },
      ],
      errors: [{ message: 'Object wins' }],
    },
    // Same prop attribute appearing twice on the same element — both fire.
    {
      code: `<div id="a" id="b" />;`,
      options: [{ forbid: ['id'] }],
      errors: [
        { messageId: 'propIsForbidden' },
        { messageId: 'propIsForbidden' },
      ],
    },
  ],
});
