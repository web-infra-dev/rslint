import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectPropIsForbiddenError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'propIsForbidden',
    message: 'Prop "id" is forbidden on DOM Nodes',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: '',
    message: 'Please use class instead of ClassName',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: '',
    message: 'Avoid using otherProp',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('forbid-dom-props', {
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
      errors: [expectPropIsForbiddenError([5, 25, 5, 33])],
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
      errors: [expectPropIsForbiddenError([4, 25, 4, 33])],
    },
    {
      code: `
        const First = (props) => (
          <div id="foo" />
        );
      `,
      options: [{ forbid: ['id'] }],
      errors: [expectPropIsForbiddenError([3, 16, 3, 24])],
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
      errors: [expectError2([3, 16, 3, 31])],
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
      errors: [
        {
          messageId: 'propIsForbidden',
          message: 'Prop "otherProp" is forbidden on DOM Nodes',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 32,
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
          forbid: [{ propName: 'someProp', disallowedValues: ['someValue'] }],
        },
      ],
      errors: [
        {
          messageId: 'propIsForbiddenWithValue',
          message:
            'Prop "someProp" with value "someValue" is forbidden on DOM Nodes',
          line: 3,
          column: 16,
          endLine: 3,
          endColumn: 36,
        },
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
            {
              propName: 'className',
              message: 'Please use class instead of ClassName',
            },
            { propName: 'otherProp', message: 'Avoid using otherProp' },
          ],
        },
      ],
      errors: [expectError2([3, 16, 3, 31]), expectError3([4, 18, 4, 33])],
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
        {
          messageId: 'propIsForbidden',
          message: 'Prop "className" is forbidden on DOM Nodes',
          line: 3,
          column: 16,
          endLine: 3,
          endColumn: 31,
        },
        expectError3([4, 18, 4, 33]),
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
      errors: [
        {
          messageId: '',
          message: 'Avoid using the accept attribute on <form>',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 30,
        },
      ],
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
        expectError2([3, 16, 3, 31]),
        expectError2([5, 19, 5, 37]),
        expectError3([6, 18, 6, 33]),
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
        expectError2([3, 16, 3, 31]),
        expectError2([5, 19, 5, 37]),
        expectError3([6, 18, 6, 33]),
        {
          messageId: '',
          message: 'Do not use thirdProp with values bar and baz on p',
          line: 9,
          column: 16,
          endLine: 9,
          endColumn: 31,
        },
        {
          messageId: '',
          message: 'Do not use thirdProp with values bar and baz on p',
          line: 10,
          column: 16,
          endLine: 10,
          endColumn: 31,
        },
      ],
    },

    // ---- Additional lock-in cases ----
    // Spread + multiple forbidden named attrs on a DOM tag — each reports.
    {
      code: `<div {...rest} id="x" className="y" />;`,
      options: [{ forbid: ['id', 'className'] }],
      errors: [
        expectPropIsForbiddenError([1, 16, 1, 22]),
        {
          messageId: 'propIsForbidden',
          message: 'Prop "className" is forbidden on DOM Nodes',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 36,
        },
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
      errors: [
        {
          messageId: 'propIsForbiddenWithValue',
          message:
            'Prop "target" with value "_blank" is forbidden on DOM Nodes',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 29,
        },
      ],
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
      errors: [
        {
          messageId: '',
          message: 'Object wins',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    // Same prop attribute appearing twice on the same element — both fire.
    {
      code: `<div id="a" id="b" />;`,
      options: [{ forbid: ['id'] }],
      errors: [
        expectPropIsForbiddenError([1, 6, 1, 12]),
        expectPropIsForbiddenError([1, 13, 1, 19]),
      ],
    },
  ],
});
