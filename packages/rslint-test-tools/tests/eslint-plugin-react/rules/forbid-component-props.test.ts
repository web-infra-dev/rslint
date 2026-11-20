import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectPropIsForbiddenError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'propIsForbidden',
    message: 'Prop "className" is forbidden on Components',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectPropIsForbiddenError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'propIsForbidden',
    message: 'Prop "style" is forbidden on Components',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectPropIsForbiddenError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'propIsForbidden',
    message: 'Prop "kebab-case-prop" is forbidden on Components',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectError4(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: '',
    message:
      'Avoid using className for SomeSvg and components that match the `UI*` and `*Icon` patterns',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('forbid-component-props', {
  valid: [
    // ---- Upstream valid cases ----
    {
      code: `
        var First = createReactClass({
          render: function() {
            return <div className="foo" />;
          }
        });
      `,
    },
    {
      code: `
        var First = createReactClass({
          render: function() {
            return <div style={{color: "red"}} />;
          }
        });
      `,
      options: [{ forbid: ['style'] }],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo bar="baz" />;
          }
        });
      `,
    },
    {
      code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
      options: [{ forbid: ['style'] }],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
      options: [{ forbid: ['style', 'foo'] }],
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
    },
    {
      code: `
        class First extends createReactClass {
          render() {
            return <this.foo className="bar" />;
          }
        }
      `,
      options: [{ forbid: ['style'] }],
    },
    {
      code: `
        const First = (props) => (
          <this.Foo {...props} />
        );
      `,
    },
    {
      code: `
        const item = (<ReactModal className="foo" />);
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              allowedFor: ['ReactModal'],
            },
          ],
        },
      ],
    },
    {
      code: `
        const item = (<MyLayout.Content className="customFoo" />);
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              allowedFor: ['MyLayout.Content'],
            },
          ],
        },
      ],
    },
    {
      code: `
        const item = (<this.ReactModal className="foo" />);
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              allowedFor: ['this.ReactModal'],
            },
          ],
        },
      ],
    },
    {
      code: `
        const item = (<Foo className="foo" />);
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              disallowedFor: ['ReactModal'],
            },
          ],
        },
      ],
    },
    {
      code: `<fbt:param name="Total number of files" number={true} />`,
    },
    {
      code: `
        const item = (
          <Foo className="bar">
            <ReactModal style={{color: "red"}} />
          </Foo>
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              disallowedFor: ['OtherModal', 'ReactModal'],
            },
            {
              propName: 'style',
              disallowedFor: ['Foo'],
            },
          ],
        },
      ],
    },
    {
      code: `
        const item = (
          <Foo className="bar">
            <ReactModal style={{color: "red"}} />
          </Foo>
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              disallowedFor: ['OtherModal', 'ReactModal'],
            },
            {
              propName: 'style',
              allowedFor: ['ReactModal'],
            },
          ],
        },
      ],
    },
    {
      code: `
        const item = (<this.ReactModal className="foo" />);
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              disallowedFor: ['ReactModal'],
            },
          ],
        },
      ],
    },
    {
      code: `
        const MyComponent = () => (
          <div aria-label="welcome" />
        );
      `,
      options: [
        {
          forbid: [
            {
              propNamePattern: '**-**',
              allowedFor: ['div'],
            },
          ],
        },
      ],
    },
    {
      code: `
        const rootElement = (
          <Root>
            <SomeIcon className="size-lg" />
            <AnotherIcon className="size-lg" />
            <SomeSvg className="size-lg" />
            <UICard className="size-lg" />
            <UIButton className="size-lg" />
          </Root>
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              allowedForPatterns: ['*Icon', '*Svg', 'UI*'],
            },
          ],
        },
      ],
    },
    {
      code: `
        const rootElement = (
          <Root>
            <SomeIcon className="size-lg" />
            <AnotherIcon className="size-lg" />
            <SomeSvg className="size-lg" />
            <UICard className="size-lg" />
            <UIButton className="size-lg" />
            <ButtonLegacy className="size-lg" />
          </Root>
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              allowedFor: ['ButtonLegacy'],
              allowedForPatterns: ['*Icon', '*Svg', 'UI*'],
            },
          ],
        },
      ],
    },
    {
      code: `
        const rootElement = (
          <Root>
            <SomeIcon className="size-lg" />
            <AnotherIcon className="size-lg" />
            <SomeSvg className="size-lg" />
            <UICard className="size-lg" />
            <UIButton className="size-lg" />
          </Root>
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              disallowedFor: ['Modal'],
              disallowedForPatterns: ['*Legacy', 'Shared*'],
            },
          ],
        },
      ],
    },

    // ---- Additional edge cases ----
    // DOM rightmost (`<Foo.bar>`) skipped — tag is "Foo.bar" but `bar` is lowercase.
    { code: `<Foo.bar className="x" />;` },
    // Empty forbid list disables the rule entirely.
    { code: `<Foo className="x" style={{}} />;`, options: [{ forbid: [] }] },
    // Spread-only Component — no JsxAttribute to check.
    { code: `<Foo {...props} />;` },
    // Boolean-shorthand attribute `<Foo bar />` not in default forbid.
    { code: `<Foo bar />;` },
    // JSX Fragment with multiple Components inside — each checked on its own.
    {
      code: `<><Foo bar="x" /><Bar baz="y" /></>;`,
    },
    // `forbid` with empty string entries — empties are skipped.
    {
      code: `<Foo other="x" />;`,
      options: [{ forbid: ['', 'className'] }],
    },
    // DOM ancestors don't shield Component descendants from being checked
    // (default forbid catches Component className, but `<a>` / `<span>`
    // are DOM and skipped).
    {
      code: `
        <a><Foo bar="ok"><span>text</span></Foo></a>
      `,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        const item = (<AntdLayout.Content className="antdFoo" />);
      `,
      options: [
        {
          forbid: [
            { propName: 'className', allowedFor: ['AntdLayout.Content'] },
          ],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        <fbt:param name="Total number of files" number={true} />
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
      errors: [expectPropIsForbiddenError1([5, 25, 5, 40])],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo style={{color: "red"}} />;
          }
        });
      `,
      errors: [expectPropIsForbiddenError2([5, 25, 5, 47])],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
      options: [{ forbid: ['className', 'style'] }],
      errors: [expectPropIsForbiddenError1([5, 25, 5, 40])],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo style={{color: "red"}} />;
          }
        });
      `,
      options: [{ forbid: ['className', 'style'] }],
      errors: [expectPropIsForbiddenError2([5, 25, 5, 47])],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo style={{color: "red"}} />;
          }
        });
      `,
      options: [
        {
          forbid: [
            {
              propName: 'style',
              disallowedFor: ['Foo'],
            },
          ],
        },
      ],
      errors: [expectPropIsForbiddenError2([5, 25, 5, 47])],
    },
    {
      code: `
        const item = (<Foo className="foo" />);
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              allowedFor: ['ReactModal'],
            },
          ],
        },
      ],
      errors: [expectPropIsForbiddenError1([2, 28, 2, 43])],
    },
    {
      code: `
        const item = (<this.ReactModal className="foo" />);
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              allowedFor: ['ReactModal'],
            },
          ],
        },
      ],
      errors: [expectPropIsForbiddenError1([2, 40, 2, 55])],
    },
    {
      code: `
        const item = (<this.ReactModal className="foo" />);
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              disallowedFor: ['this.ReactModal'],
            },
          ],
        },
      ],
      errors: [expectPropIsForbiddenError1([2, 40, 2, 55])],
    },
    {
      code: `
        const item = (<ReactModal className="foo" />);
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              disallowedFor: ['ReactModal'],
            },
          ],
        },
      ],
      errors: [expectPropIsForbiddenError1([2, 35, 2, 50])],
    },
    {
      code: `
        const item = (<MyLayout.Content className="customFoo" />);
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              disallowedFor: ['MyLayout.Content'],
            },
          ],
        },
      ],
      errors: [expectPropIsForbiddenError1([2, 41, 2, 62])],
    },
    {
      code: `
        const item = (<Foo className="foo" />);
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              message: 'Please use ourCoolClassName instead of ClassName',
            },
          ],
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Please use ourCoolClassName instead of ClassName',
          line: 2,
          column: 28,
          endLine: 2,
          endColumn: 43,
        },
      ],
    },
    {
      code: `
        const item = () => (
          <Foo className="foo">
            <Bar option="high" />
          </Foo>
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              message: 'Please use ourCoolClassName instead of ClassName',
            },
            {
              propName: 'option',
              message: 'Avoid using option',
            },
          ],
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Please use ourCoolClassName instead of ClassName',
          line: 3,
          column: 16,
          endLine: 3,
          endColumn: 31,
        },
        {
          messageId: '',
          message: 'Avoid using option',
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 31,
        },
      ],
    },
    {
      code: `
        const item = () => (
          <Foo className="foo">
            <Bar option="high" />
          </Foo>
        );
      `,
      options: [
        {
          forbid: [
            { propName: 'className' },
            {
              propName: 'option',
              message: 'Avoid using option',
            },
          ],
        },
      ],
      errors: [
        expectPropIsForbiddenError1([3, 16, 3, 31]),
        {
          messageId: '',
          message: 'Avoid using option',
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 31,
        },
      ],
    },
    {
      code: `
        const MyComponent = () => (
          <Foo kebab-case-prop={123} />
        );
      `,
      options: [
        {
          forbid: [
            {
              propNamePattern: '**-**',
            },
          ],
        },
      ],
      errors: [expectPropIsForbiddenError3([3, 16, 3, 37])],
    },
    {
      code: `
        const MyComponent = () => (
          <Foo kebab-case-prop={123} />
        );
      `,
      options: [
        {
          forbid: [
            {
              propNamePattern: '**-**',
              message: 'Avoid using kebab-case',
            },
          ],
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Avoid using kebab-case',
          line: 3,
          column: 16,
          endLine: 3,
          endColumn: 37,
        },
      ],
    },
    {
      code: `
        const MyComponent = () => (
          <div>
            <div aria-label="Hello World" />
            <Foo kebab-case-prop={123} />
          </div>
        );
      `,
      options: [
        {
          forbid: [
            {
              propNamePattern: '**-**',
              allowedFor: ['div'],
            },
          ],
        },
      ],
      errors: [expectPropIsForbiddenError3([5, 18, 5, 39])],
    },
    {
      code: `
        const MyComponent = () => (
          <div>
            <div aria-label="Hello World" />
            <h1 data-id="my-heading" />
            <Foo kebab-case-prop={123} />
          </div>
        );
      `,
      options: [
        {
          forbid: [
            {
              propNamePattern: '**-**',
              disallowedFor: ['Foo'],
            },
          ],
        },
      ],
      errors: [expectPropIsForbiddenError3([6, 18, 6, 39])],
    },
    {
      code: `
        const rootElement = () => (
          <Root>
            <SomeIcon className="size-lg" />
            <SomeSvg className="size-lg" />
          </Root>
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              message: 'className available only for icons',
              allowedForPatterns: ['*Icon'],
            },
          ],
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'className available only for icons',
          line: 5,
          column: 22,
          endLine: 5,
          endColumn: 41,
        },
      ],
    },
    {
      code: `
        const rootElement = () => (
          <Root>
            <UICard style={{backgroundColor: black}}/>
            <SomeIcon className="size-lg" />
            <SomeSvg className="size-lg" style={{fill: currentColor}} />
          </Root>
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              message: 'className available only for icons',
              allowedForPatterns: ['*Icon'],
            },
            {
              propName: 'style',
              message: 'style available only for SVGs',
              allowedForPatterns: ['*Svg'],
            },
          ],
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'style available only for SVGs',
          line: 4,
          column: 21,
          endLine: 4,
          endColumn: 53,
        },
        {
          messageId: '',
          message: 'className available only for icons',
          line: 6,
          column: 22,
          endLine: 6,
          endColumn: 41,
        },
      ],
    },
    {
      code: `
        const rootElement = (
          <Root>
            <SomeIcon className="size-lg" />
            <AnotherIcon className="size-lg" />
            <SomeSvg className="size-lg" />
            <UICard className="size-lg" />
            <ButtonLegacy className="size-lg" />
          </Root>
        );
      `,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              disallowedFor: ['SomeSvg'],
              disallowedForPatterns: ['UI*', '*Icon'],
              message:
                'Avoid using className for SomeSvg and components that match the `UI*` and `*Icon` patterns',
            },
          ],
        },
      ],
      errors: [
        expectError4([4, 23, 4, 42]),
        expectError4([5, 26, 5, 45]),
        expectError4([6, 22, 6, 41]),
        expectError4([7, 21, 7, 40]),
      ],
    },

    // ---- Additional lock-in cases ----
    // `<Foo.Bar.Baz>` deep-nested member expression: tag becomes the literal
    // `"undefined.Baz"` (mirrors upstream quirk).
    {
      code: `<Foo.Bar.Baz className="x" />;`,
      options: [
        {
          forbid: [
            {
              propName: 'className',
              disallowedFor: ['undefined.Baz'],
            },
          ],
        },
      ],
      errors: [expectPropIsForbiddenError1([1, 14, 1, 27])],
    },
    // JSX namespaced name with default forbid: namespaced tags bypass the
    // DOM lowercase-skip path and are treated as Components, so className
    // is reported.
    {
      code: `<fbt:param className="x" />;`,
      errors: [expectPropIsForbiddenError1([1, 12, 1, 25])],
    },
    // Multi-level Component nesting — every JsxAttribute on a Component is
    // checked, regardless of depth. DOM `<div>` between Components is skipped.
    {
      code: `
        <Outer className="o1">
          <div className="d1">
            <Inner className="i1">
              <Leaf className="l1" />
            </Inner>
          </div>
        </Outer>
      `,
      errors: [
        expectPropIsForbiddenError1([2, 16, 2, 30]),
        expectPropIsForbiddenError1([4, 20, 4, 34]),
        expectPropIsForbiddenError1([5, 21, 5, 35]),
      ],
    },
    // Spread + multiple named attributes: each forbidden named attr fires
    // independently; the spread is ignored.
    {
      code: `<Foo {...rest} className="x" style={{}} />;`,
      errors: [
        expectPropIsForbiddenError1([1, 16, 1, 29]),
        expectPropIsForbiddenError2([1, 30, 1, 40]),
      ],
    },
    // Direct propName lookup beats a propNamePattern that would also match.
    {
      code: `<Foo className="x" />;`,
      options: [
        {
          forbid: [
            {
              propNamePattern: '*Name',
              message: 'Pattern wins (should NOT see this)',
            },
            { propName: 'className' },
          ],
        },
      ],
      errors: [expectPropIsForbiddenError1([1, 6, 1, 19])],
    },
    // Two pattern entries match the same prop — first-in-insertion-order wins.
    {
      code: `<Foo data-id="x" />;`,
      options: [
        {
          forbid: [
            { propNamePattern: 'data-*', message: 'First pattern' },
            {
              propNamePattern: '*-id',
              message: 'Second pattern (should NOT see this)',
            },
          ],
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'First pattern',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    // 2-segment member expression `<Foo.Bar>`: tag is canonical "Foo.Bar".
    {
      code: `<Foo.Bar style={{}} />;`,
      options: [
        {
          forbid: [{ propName: 'style', disallowedFor: ['Foo.Bar'] }],
        },
      ],
      errors: [expectPropIsForbiddenError2([1, 10, 1, 20])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        const item = (<AntdLayout.Content className="antdFoo" />);
      `,
      options: [
        {
          forbid: [
            { propName: 'className', disallowedFor: ['AntdLayout.Content'] },
          ],
        },
      ],
      errors: [expectPropIsForbiddenError1([2, 43, 2, 62])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        const MyComponent = () => (
          <div>
            <div aria-label="Hello Akul" />
            <Foo kebab-case-prop={123} />
          </div>
        );
      `,
      options: [
        { forbid: [{ propNamePattern: '**-**', allowedFor: ['div'] }] },
      ],
      errors: [expectPropIsForbiddenError3([5, 18, 5, 39])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        const MyComponent = () => (
          <div>
            <div aria-label="Hello Akul" />
            <h1 data-id="my-heading" />
            <Foo kebab-case-prop={123} />
          </div>
        );
      `,
      options: [
        { forbid: [{ propNamePattern: '**-**', disallowedFor: ['Foo'] }] },
      ],
      errors: [expectPropIsForbiddenError3([6, 18, 6, 39])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
