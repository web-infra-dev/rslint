import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('forbid-component-props', {} as never, {
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
      errors: [
        {
          messageId: 'propIsForbidden',
        },
      ],
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
      errors: [
        {
          messageId: 'propIsForbidden',
        },
      ],
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
      errors: [
        {
          messageId: 'propIsForbidden',
        },
      ],
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
      errors: [
        {
          messageId: 'propIsForbidden',
        },
      ],
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
      errors: [
        {
          messageId: 'propIsForbidden',
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
              allowedFor: ['ReactModal'],
            },
          ],
        },
      ],
      errors: [
        {
          messageId: 'propIsForbidden',
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
              allowedFor: ['ReactModal'],
            },
          ],
        },
      ],
      errors: [
        {
          messageId: 'propIsForbidden',
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
              disallowedFor: ['this.ReactModal'],
            },
          ],
        },
      ],
      errors: [
        {
          messageId: 'propIsForbidden',
        },
      ],
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
      errors: [
        {
          messageId: 'propIsForbidden',
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
              disallowedFor: ['MyLayout.Content'],
            },
          ],
        },
      ],
      errors: [
        {
          messageId: 'propIsForbidden',
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
              message: 'Please use ourCoolClassName instead of ClassName',
            },
          ],
        },
      ],
      errors: [
        {
          message: 'Please use ourCoolClassName instead of ClassName',
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
        { message: 'Please use ourCoolClassName instead of ClassName' },
        { message: 'Avoid using option' },
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
        { messageId: 'propIsForbidden' },
        { message: 'Avoid using option' },
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
      errors: [
        {
          messageId: 'propIsForbidden',
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
              message: 'Avoid using kebab-case',
            },
          ],
        },
      ],
      errors: [
        {
          message: 'Avoid using kebab-case',
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
      errors: [
        {
          messageId: 'propIsForbidden',
        },
      ],
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
      errors: [
        {
          messageId: 'propIsForbidden',
        },
      ],
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
          message: 'className available only for icons',
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
        { message: 'style available only for SVGs' },
        { message: 'className available only for icons' },
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
        {
          message:
            'Avoid using className for SomeSvg and components that match the `UI*` and `*Icon` patterns',
        },
        {
          message:
            'Avoid using className for SomeSvg and components that match the `UI*` and `*Icon` patterns',
        },
        {
          message:
            'Avoid using className for SomeSvg and components that match the `UI*` and `*Icon` patterns',
        },
        {
          message:
            'Avoid using className for SomeSvg and components that match the `UI*` and `*Icon` patterns',
        },
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
      errors: [{ messageId: 'propIsForbidden' }],
    },
    // JSX namespaced name with default forbid: namespaced tags bypass the
    // DOM lowercase-skip path and are treated as Components, so className
    // is reported.
    {
      code: `<fbt:param className="x" />;`,
      errors: [{ messageId: 'propIsForbidden' }],
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
        { messageId: 'propIsForbidden' },
        { messageId: 'propIsForbidden' },
        { messageId: 'propIsForbidden' },
      ],
    },
    // Spread + multiple named attributes: each forbidden named attr fires
    // independently; the spread is ignored.
    {
      code: `<Foo {...rest} className="x" style={{}} />;`,
      errors: [
        { messageId: 'propIsForbidden' },
        { messageId: 'propIsForbidden' },
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
      errors: [{ messageId: 'propIsForbidden' }],
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
      errors: [{ message: 'First pattern' }],
    },
    // 2-segment member expression `<Foo.Bar>`: tag is canonical "Foo.Bar".
    {
      code: `<Foo.Bar style={{}} />;`,
      options: [
        {
          forbid: [{ propName: 'style', disallowedFor: ['Foo.Bar'] }],
        },
      ],
      errors: [{ messageId: 'propIsForbidden' }],
    },
  ],
});
