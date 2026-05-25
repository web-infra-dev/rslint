import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-max-depth', {} as never, {
  valid: [
    // ---- Default max=2 ----
    { code: `<App />` },
    {
      code: `
        <App>
          <foo>
            <bar />
          </foo>
        </App>
      `,
    },

    // ---- Explicit max=1, depth 1 ----
    {
      code: `
        <App>
          <foo />
        </App>
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        <App>
          <foo>
            <bar />
          </foo>
        </App>
      `,
      options: [{ max: 2 }],
    },

    // ---- Identifier resolution to JSX ----
    {
      code: `
        const x = <div><em>x</em></div>;
        <div>{x}</div>;
      `,
      options: [{ max: 2 }],
    },
    {
      code: `const foo = (x) => <div><em>{x}</em></div>;`,
      options: [{ max: 2 }],
    },

    // ---- Fragments ----
    { code: `<></>;` },
    {
      code: `
        <>
          <foo />
        </>
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        const x = <><em>x</em></>;
        <>{x}</>;
      `,
      options: [{ max: 2 }],
    },

    // ---- Identifier-resolved table fragment, depth fits max=2 ----
    {
      code: `
        const x = (
          <tr>
            <td>1</td>
            <td>2</td>
          </tr>
        );
        <tbody>
          {x}
        </tbody>
      `,
      options: [{ max: 2 }],
    },

    // ---- Function-with-loop returns leaf JSX, max=1 ----
    {
      code: `
        const Example = props => {
          for (let i = 0; i < length; i++) {
            return <Text key={i} />;
          }
        };
      `,
      options: [{ max: 1 }],
    },

    // ---- React.Fragment with an interpolation ----
    {
      code: `
        export function MyComponent() {
          const A = <React.Fragment>{<div />}</React.Fragment>;
          return <div>{A}</div>;
        }
      `,
    },

    // ---- Circular references (single level) ----
    {
      code: `
        function Component() {
          let first = "";
          const second = first;
          first = second;
          return <div id={first} />;
        }
      `,
    },

    // ---- Circular references (multi level) ----
    {
      code: `
        function Component() {
          let first = "";
          let second = "";
          let third = "";
          let fourth = "";
          const fifth = first;
          first = second;
          second = third;
          third = fourth;
          fourth = fifth;
          return <div id={first} />;
        }
      `,
    },

    // ---- tsgo preserves parens; SkipParentheses keeps `{(x)}` aligned ----
    {
      code: `
        const x = <div><span /></div>;
        <div>{(x)}</div>;
      `,
      options: [{ max: 2 }],
    },

    // ---- TS-only wrappers do NOT trigger the resolve path ----
    {
      code: `
        const x = <div><span /></div>;
        <div>{x as any}</div>;
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        const x = <div><span /></div>;
        <div>{x!}</div>;
      `,
      options: [{ max: 1 }],
    },

    // ---- JSX inside an attribute is at depth 0 ----
    { code: `<div title={<span />} />`, options: [{ max: 1 }] },

    // ---- Text-only children remain leaves ----
    { code: `<div>hello</div>`, options: [{ max: 0 }] },

    // ---- Identifier resolution to non-JSX ----
    {
      code: `
        const x = "hello";
        <div>{x}</div>;
      `,
      options: [{ max: 0 }],
    },
    {
      code: `
        const x = 42;
        <div>{x}</div>;
      `,
      options: [{ max: 0 }],
    },

    // ---- Reassignment to non-JSX after JSX init bails ----
    {
      code: `
        let x = <div><span /></div>;
        x = "";
        <div>{x}</div>;
      `,
      options: [{ max: 1 }],
    },

    // ---- Parameter binding has no resolvable write ----
    {
      code: `
        function Foo(x) {
          return <div>{x}</div>;
        }
      `,
      options: [{ max: 0 }],
    },

    // ---- Destructured binding doesn't resolve ----
    {
      code: `
        const { x } = obj;
        <div>{x}</div>;
      `,
      options: [{ max: 0 }],
    },

    // ---- Property assignment is not a write to the binding ----
    {
      code: `
        let x = "hello";
        obj.x = <div><span /></div>;
        <div>{x}</div>;
      `,
      options: [{ max: 1 }],
    },

    // ---- Compound assignment is not a write ----
    {
      code: `
        let x = "hello";
        x += "world";
        <div>{x}</div>;
      `,
      options: [{ max: 0 }],
    },

    // ---- Block-level shadowing via inner let ----
    {
      code: `
        let x = "";
        if (cond) {
          let x = <div><span /></div>;
          void x;
        }
        <div>{x}</div>;
      `,
      options: [{ max: 1 }],
    },

    // ---- Function parameter shadowing ----
    {
      code: `
        let x = "";
        function inner(x) {
          x = <div><span /></div>;
          return x;
        }
        <div>{x}</div>;
      `,
      options: [{ max: 1 }],
    },

    // ---- Function-body let shadowing ----
    {
      code: `
        let x = "";
        function inner() {
          let x = <div><span /></div>;
          return x;
        }
        <div>{x}</div>;
      `,
      options: [{ max: 1 }],
    },

    // ---- ConditionalExpression breaks the parent walk ----
    {
      code: `<div>{cond ? <span /> : null}</div>`,
      options: [{ max: 0 }],
    },

    // ---- Class component returning fitting JSX ----
    {
      code: `
        class Foo extends React.Component {
          render() {
            return (
              <div>
                <span />
              </div>
            );
          }
        }
      `,
      options: [{ max: 1 }],
    },

    // ---- Arrow with block body ----
    {
      code: `
        const Foo = () => {
          return (
            <div>
              <span />
            </div>
          );
        };
      `,
      options: [{ max: 1 }],
    },

    // ---- JsxExpression wrapping JSX inside a Fragment is opaque (max=3 fits) ----
    {
      code: `
        const A = <Fragment>{<div><span /></div>}</Fragment>;
        <div>{A}</div>;
      `,
      options: [{ max: 3 }],
    },

    // ---- Option-shape coverage ----
    { code: `<App><foo><bar/></foo></App>`, options: [] },
    { code: `<App><foo><bar/></foo></App>`, options: [{}] },
    {
      code: `<App><foo><bar/></foo></App>`,
      options: [{ max: 'two' }],
    },
    {
      code: `<App><foo><bar/></foo></App>`,
      options: [{ max: false }],
    },
    {
      code: `<App><foo><bar/></foo></App>`,
      options: [{ max: null }],
    },
    {
      code: `<App><foo><bar/></foo></App>`,
      options: [{ max: -1 }],
    },
    {
      code: `<App><foo><bar/></foo></App>`,
      options: [{ max: 2, unknown: true }],
    },
    {
      code: `
        <a><b><c><d><e><f><g><h><i /></h></g></f></e></d></c></b></a>
      `,
      options: [{ max: 100 }],
    },

    // ---- JsxExpression non-Identifier inner shapes ----
    { code: `<div>{this.x}</div>`, options: [{ max: 0 }] },
    { code: `<div>{x?.y}</div>`, options: [{ max: 0 }] },
    { code: `<div>{makeJSX()}</div>`, options: [{ max: 0 }] },
    { code: `<div>{...arr}</div>`, options: [{ max: 0 }] },
    { code: `<div>{}</div>`, options: [{ max: 0 }] },
    { code: `<div>{/* nothing */}</div>`, options: [{ max: 0 }] },
    { code: `<div>{x + y}</div>`, options: [{ max: 0 }] },
    { code: `<div>{cond && <span />}</div>`, options: [{ max: 0 }] },

    // ---- Real React patterns ----
    {
      code: `
        <Suspense fallback={<Loading />}>
          <App />
        </Suspense>
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        <Provider>
          <Consumer>
            {(value) => <span>{value}</span>}
          </Consumer>
        </Provider>
      `,
      options: [{ max: 2 }],
    },
    {
      code: `<ul>{items.map((item) => <li>{item}</li>)}</ul>`,
      options: [{ max: 1 }],
    },
    {
      code: `
        const Foo = () => {
          const memo = useMemo(() => <div><span /></div>, []);
          return <div>{memo}</div>;
        };
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        function Foo({ children = <span /> }) {
          return <div>{children}</div>;
        }
      `,
      options: [{ max: 0 }],
    },

    // ---- Identifier resolution edge cases ----
    {
      code: `
        let x = x;
        <div>{x}</div>;
      `,
      options: [{ max: 0 }],
    },
    {
      code: `
        let x = y;
        let y = x;
        <div>{x}</div>;
      `,
      options: [{ max: 0 }],
    },
    {
      code: `
        let a = "";
        let b = a;
        let c = b;
        let d = c;
        let e = d;
        <div>{e}</div>;
      `,
      options: [{ max: 0 }],
    },
    {
      code: `
        const x = makeJSX();
        <div>{x}</div>;
      `,
      options: [{ max: 0 }],
    },

    // ---- JSX shape edges ----
    { code: `<></>;`, options: [{ max: 0 }] },
    { code: `<div></div>;`, options: [{ max: 0 }] },
    {
      code: `
        <div>
          {/* comment */}
          <span />
        </div>
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        <div {...props}>
          <span {...spanProps} />
        </div>
      `,
      options: [{ max: 1 }],
    },
    { code: `<a.b.c.d />;`, options: [{ max: 0 }] },
    {
      code: `const F = () => <div><span /></div>;`,
      options: [{ max: 1 }],
    },

    // ---- TS satisfies wrapper not peeled ----
    {
      code: `
        const x = <div><span /></div>;
        <div>{x satisfies any}</div>;
      `,
      options: [{ max: 1 }],
    },

    // ---- Reassignment in control flow that fits ----
    {
      code: `
        function Foo() {
          let x = "";
          try { x = <div />; } catch (e) {}
          return <div>{x}</div>;
        }
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        function Foo() {
          let x = "";
          while (cond) { x = <div />; }
          return <div>{x}</div>;
        }
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        function Foo() {
          let x = "";
          switch (kind) { case "a": x = <div />; break; }
          return <div>{x}</div>;
        }
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        function Foo() {
          if (cond) { var x = ""; }
          return <div>{x}</div>;
        }
      `,
      options: [{ max: 0 }],
    },

    // ---- For-loop init scope isolation ----
    {
      code: `
        let x = <a />;
        function Foo() {
          for (let x of items) {
            return <div>{x}</div>;
          }
        }
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        let x = <a />;
        function Foo() {
          for (let x = 0; x < n; x++) {
            return <span key={x} />;
          }
        }
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        let x = <a />;
        function Foo() {
          for (let x in obj) {
            return <span key={x} />;
          }
        }
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        let x = "";
        for (const x of items) { void x; }
        <div>{x}</div>;
      `,
      options: [{ max: 0 }],
    },

    // ---- Inner for-of `let x` reassignment must not leak to outer `let x` ----
    {
      code: `
        let x = "";
        for (let x of items) { x = <div><span /></div>; }
        const y = <wrap>{x}</wrap>;
      `,
      options: [{ max: 1 }],
    },
    // for (let x = …; …; …) variant
    {
      code: `
        let x = "";
        for (let x = 0; x < n; x++) { x = <div><span /></div>; }
        const y = <wrap>{x}</wrap>;
      `,
      options: [{ max: 1 }],
    },
    // for-in variant
    {
      code: `
        let k = "";
        for (let k in obj) { k = <div><span /></div>; }
        const y = <wrap>{k}</wrap>;
      `,
      options: [{ max: 1 }],
    },
    // ---- catch (x) parameter shadows outer let x ----
    {
      code: `
        let x = "";
        try { foo(); } catch (x) { x = <div><span /></div>; }
        const y = <wrap>{x}</wrap>;
      `,
      options: [{ max: 1 }],
    },
    // Destructured catch param: catch ({ name })
    {
      code: `
        let err = "";
        try { foo(); } catch ({ err }) { err = <div><span /></div>; }
        const y = <wrap>{err}</wrap>;
      `,
      options: [{ max: 1 }],
    },

    // ---- Parameter / destructured default ----
    {
      code: `
        function Foo(x = <a />) {
          return <wrap>{x}</wrap>;
        }
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        function Foo({ children = <span /> }) {
          return <div>{children}</div>;
        }
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        function Foo(x = <div><span /></div>) {
          x = "";
          return <wrap>{x}</wrap>;
        }
      `,
      options: [{ max: 1 }],
    },

    // ---- Switch case-let scope ----
    {
      code: `
        function Foo(kind) {
          switch (kind) {
            case 'a': let x = <a />; break;
            case 'b': return <wrap>{x}</wrap>;
          }
        }
      `,
      options: [{ max: 1 }],
    },

    // ---- Destructured VariableDeclaration default is NOT a write ----
    // Differential validation against upstream confirms ESLint's scope
    // manager surfaces the destructure RHS, not the BindingElement default.
    {
      code: `
        const { x = <div><span /></div> } = obj;
        const y = <wrap>{x}</wrap>;
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        let { x = <div><span /></div> } = obj;
        const y = <wrap>{x}</wrap>;
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
        const [x = <div><span /></div>] = arr;
        const y = <wrap>{x}</wrap>;
      `,
      options: [{ max: 1 }],
    },
  ],

  invalid: [
    // ---- max=0 with depth 1 ----
    {
      code: `
        <App>
          <foo />
        </App>
      `,
      options: [{ max: 0 }],
      errors: [{ messageId: 'wrongDepth' }],
    },
    // ---- max=0 with depth 1, JsxExpression child whose inner is non-JSX ----
    {
      code: `
        <App>
          <foo>{bar}</foo>
        </App>
      `,
      options: [{ max: 0 }],
      errors: [{ messageId: 'wrongDepth' }],
    },
    // ---- max=1 with depth 2 ----
    {
      code: `
        <App>
          <foo>
            <bar />
          </foo>
        </App>
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },
    // ---- Identifier resolution puts <span/> at depth 2 ----
    {
      code: `
        const x = <div><span /></div>;
        <div>{x}</div>;
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },
    // ---- Chain through `let y = x` ----
    {
      code: `
        const x = <div><span /></div>;
        let y = x;
        <div>{y}</div>;
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },
    // ---- Two interpolations report independently ----
    {
      code: `
        const x = <div><span /></div>;
        let y = x;
        <div>{x}-{y}</div>;
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }, { messageId: 'wrongDepth' }],
    },
    // ---- Inline `{<div>...</div>}`; <span/> at depth 3 ----
    {
      code: `
        <div>
        {<div><div><span /></div></div>}
        </div>
      `,
      errors: [{ messageId: 'wrongDepth' }],
    },
    // ---- Fragment with depth 1, max=0 ----
    {
      code: `
        <>
          <foo />
        </>
      `,
      options: [{ max: 0 }],
      errors: [{ messageId: 'wrongDepth' }],
    },
    // ---- Nested fragments, max=1, depth 2 ----
    {
      code: `
        <>
          <>
            <bar />
          </>
        </>
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },
    // ---- Fragment-wrapped duplicate interpolations ----
    {
      code: `
        const x = <><span /></>;
        let y = x;
        <>{x}-{y}</>;
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }, { messageId: 'wrongDepth' }],
    },
    // ---- Identifier-resolved table — both <td>s exceed max=1 ----
    {
      code: `
        const x = (
          <tr>
            <td>1</td>
            <td>2</td>
          </tr>
        );
        <tbody>
          {x}
        </tbody>
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }, { messageId: 'wrongDepth' }],
    },
    // ---- Deeply nested modal scaffold, max=4 ----
    {
      code: `
        <div className="custom_modal">
          <Modal className={classes.modal} open={isOpen} closeAfterTransition>
            <Fade in={isOpen}>
              <DialogContent>
                <Icon icon="cancel" onClick={onClose} popoverText="Close Modal" />
                <div className="modal_content">{children}</div>
                <div className={clsx('modal_buttons', classes.buttons)}>
                  <Button className="modal_buttons--cancel" onClick={onCancel}>
                    {cancelMsg ? cancelMsg : 'Cancel'}
                  </Button>
                </div>
              </DialogContent>
            </Fade>
          </Modal>
        </div>
      `,
      options: [{ max: 4 }],
      errors: [{ messageId: 'wrongDepth' }],
    },
    // ---- Parens around resolved identifier still report ----
    {
      code: `
        const x = <div><span /></div>;
        <div>{(x)}</div>;
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Reassignment: latest write wins ----
    {
      code: `
        let x = "";
        x = <div><span /></div>;
        <div>{x}</div>;
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Reassignment: declaration without init ----
    {
      code: `
        let x;
        x = <div><span /></div>;
        <div>{x}</div>;
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Reassignment inside an if branch ----
    {
      code: `
        function Foo() {
          let x = "";
          if (cond) {
            x = <div><span /></div>;
          }
          return <div>{x}</div>;
        }
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Closure write from a non-shadowing inner function ----
    {
      code: `
        function outer() {
          let x = "";
          function inner() {
            x = <div><span /></div>;
          }
          return <div>{x}</div>;
        }
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Reassignment chain y → x ----
    {
      code: `
        let x = "";
        x = <div><span /></div>;
        let y = x;
        <div>{y}</div>;
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Multi-decl statement ----
    {
      code: `
        let a = 1, b = <div><span /></div>, c = 3;
        <div>{b}</div>;
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- TS generic component tag ----
    {
      code: `
        <Foo<string>>
          <Bar>
            <Baz />
          </Bar>
        </Foo>
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Member-access tag name ----
    {
      code: `
        <Module.Foo>
          <Bar>
            <Baz />
          </Bar>
        </Module.Foo>
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Mixed element / fragment nesting ----
    {
      code: `
        <App>
          <>
            <Foo>
              <Bar />
            </Foo>
          </>
        </App>
      `,
      options: [{ max: 2 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Boundary depth with exact message text ----
    {
      code: `<App><foo /></App>`,
      options: [{ max: 0 }],
      errors: [
        {
          messageId: 'wrongDepth',
          message:
            'Expected the depth of nested jsx elements to be <= 0, but found 1.',
        },
      ],
    },
    {
      code: `<App><foo><bar /></foo></App>`,
      options: [{ max: 0 }],
      errors: [
        {
          messageId: 'wrongDepth',
          message:
            'Expected the depth of nested jsx elements to be <= 0, but found 2.',
        },
      ],
    },
    {
      code: `<a><b><c><d /></c></b></a>`,
      options: [{ max: 2 }],
      errors: [
        {
          messageId: 'wrongDepth',
          message:
            'Expected the depth of nested jsx elements to be <= 2, but found 3.',
        },
      ],
    },
    {
      code: `<a><b><c><d><e><f><g /></f></e></d></c></b></a>`,
      options: [{ max: 5 }],
      errors: [
        {
          messageId: 'wrongDepth',
          message:
            'Expected the depth of nested jsx elements to be <= 5, but found 6.',
        },
      ],
    },

    // ---- Reassignment in control flow that exceeds ----
    {
      code: `
        function Foo() {
          let x = "";
          while (cond) { x = <div><span /></div>; }
          return <div>{x}</div>;
        }
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },
    {
      code: `
        function Foo() {
          let x = "";
          try { x = <div><span /></div>; } catch (e) {}
          return <div>{x}</div>;
        }
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },
    {
      code: `
        function Foo() {
          let x = "";
          switch (kind) {
            case "deep": x = <div><span /></div>; break;
          }
          return <div>{x}</div>;
        }
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Closure descent through TWO function levels ----
    {
      code: `
        function outer() {
          let x = "";
          function mid() {
            function inner() {
              x = <div><span /></div>;
            }
          }
          return <div>{x}</div>;
        }
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Multiple use sites of the same binding ----
    {
      code: `
        const x = <div><span /></div>;
        <div>{x}{x}{x}</div>;
      `,
      options: [{ max: 1 }],
      errors: [
        { messageId: 'wrongDepth' },
        { messageId: 'wrongDepth' },
        { messageId: 'wrongDepth' },
      ],
    },

    // ---- Default parameter JSX itself violates + descend reports ----
    {
      code: `
        function Foo({ children = <div><span /></div> }) {
          return <div>{children}</div>;
        }
      `,
      options: [{ max: 0 }],
      errors: [{ messageId: 'wrongDepth' }, { messageId: 'wrongDepth' }],
    },

    // ---- Same-line multi-element ----
    {
      code: `<a><b><c /></b></a>`,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- TS generic with multiple type args ----
    {
      code: `
        <Foo<string, number>>
          <Bar>
            <Baz />
          </Bar>
        </Foo>
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Stress: aliasing chain to deep JSX value ----
    {
      code: `
        let a = "";
        const b = a;
        const c = b;
        const d = c;
        const e = d;
        a = <div><span /></div>;
        <div>{e}</div>;
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- For-init binding reassigned to JSX in loop body ----
    {
      code: `
        function Foo() {
          for (let x of items) {
            x = <div><span /></div>;
            return <div>{x}</div>;
          }
        }
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },
    {
      code: `
        function Foo() {
          for (let x = <a />; cond; ) {
            x = <div><span /></div>;
            return <div>{x}</div>;
          }
        }
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Parameter default JSX exceeds via descend ----
    {
      code: `
        function Foo(x = <div><span /></div>) {
          return <wrap>{x}</wrap>;
        }
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Destructured parameter default exceeds (also leaf in default) ----
    {
      code: `
        function Foo({ children = <div><span /></div> }) {
          return <div>{children}</div>;
        }
      `,
      options: [{ max: 0 }],
      errors: [{ messageId: 'wrongDepth' }, { messageId: 'wrongDepth' }],
    },

    // ---- Switch case-let cross-clause resolution ----
    {
      code: `
        function Foo(kind) {
          switch (kind) {
            case 'a': let x = <div><span /></div>; break;
            case 'b': return <wrap>{x}</wrap>;
          }
        }
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },

    // ---- Parameter reassigned in body ----
    {
      code: `
        function Foo(x = <a />) {
          x = <div><span /></div>;
          return <wrap>{x}</wrap>;
        }
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'wrongDepth' }],
    },
  ],
});
