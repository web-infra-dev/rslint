import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectNoShouldComponentUpdateError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noShouldComponentUpdate',
    message:
      'Component is not optimized. Please add a shouldComponentUpdate method.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('require-optimization', {
  valid: [
    // ---- Upstream valid: plain class without React extension is not a component ----
    {
      code: `
        class A {}
      `,
    },
    // ---- Upstream valid: React.Component with shouldComponentUpdate ----
    {
      code: `
        import React from "react";
        class YourComponent extends React.Component {
          shouldComponentUpdate () {}
        }
      `,
    },
    // ---- Upstream valid: bare Component with shouldComponentUpdate ----
    {
      code: `
        import React, {Component} from "react";
        class YourComponent extends Component {
          shouldComponentUpdate () {}
        }
      `,
    },
    // ---- Upstream valid: PureRender decorator on bare Component ----
    {
      code: `
        import React, {Component} from "react";
        @reactMixin.decorate(PureRenderMixin)
        class YourComponent extends Component {
          componentDidMount () {}
          render() {}
        }
      `,
    },
    // ---- Upstream valid: createReactClass with shouldComponentUpdate property ----
    {
      code: `
        import React from "react";
        createReactClass({
          shouldComponentUpdate: function () {}
        })
      `,
    },
    // ---- Upstream valid: createReactClass with PureRenderMixin ----
    {
      code: `
        import React from "react";
        createReactClass({
          mixins: [PureRenderMixin]
        })
      `,
    },
    // ---- Upstream valid: PureRender decorator alone ----
    {
      code: `
        @reactMixin.decorate(PureRenderMixin)
        class DecoratedComponent extends Component {}
      `,
    },
    // ---- Upstream valid: stateless functional components are never reported ----
    {
      code: `
        const FunctionalComponent = function (props) {
          return <div />;
        }
      `,
    },
    {
      code: `
        function FunctionalComponent(props) {
          return <div />;
        }
      `,
    },
    {
      code: `
        const FunctionalComponent = (props) => {
          return <div />;
        }
      `,
    },
    // ---- Upstream valid: custom decorator allow-listed ----
    {
      code: `
        @bar
        @pureRender
        @foo
        class DecoratedComponent extends Component {}
      `,
      options: [{ allowDecorators: ['renderPure', 'pureRender'] }],
    },
    // ---- Upstream valid: React.PureComponent (allow option still applies) ----
    {
      code: `
        import React from "react";
        class YourComponent extends React.PureComponent {}
      `,
      options: [{ allowDecorators: ['renderPure', 'pureRender'] }],
    },
    // ---- Upstream valid: bare PureComponent ----
    {
      code: `
        import React, {PureComponent} from "react";
        class YourComponent extends PureComponent {}
      `,
      options: [{ allowDecorators: ['renderPure', 'pureRender'] }],
    },
    // ---- Upstream valid: object with elision-padded array — not a createReactClass call ----
    {
      code: `
        const obj = { prop: [,,,,,] }
      `,
    },
    // ---- Upstream valid: class with class-field handler + shouldComponentUpdate method ----
    {
      code: `
        import React from "react";
        class YourComponent extends React.Component {
          handleClick = () => {}
          shouldComponentUpdate(){
            return true;
          }
          render() {
            return <div onClick={this.handleClick}>123</div>
          }
        }
      `,
    },
    // ---- Edge: ClassExpression assigned to var — same React.Component check applies ----
    {
      code: `
        const YourComponent = class extends React.Component {
          shouldComponentUpdate() { return true; }
        };
      `,
    },
    // ---- Edge: PrivateIdentifier #shouldComponentUpdate — '#' is stripped on .name ----
    {
      code: `
        class YourComponent extends React.Component {
          #shouldComponentUpdate() { return true; }
        }
      `,
    },
    // ---- Edge: getter / setter / static shouldComponentUpdate ----
    {
      code: `
        class YourComponent extends React.Component {
          get shouldComponentUpdate() { return () => true; }
        }
      `,
    },
    {
      code: `
        class YourComponent extends React.Component {
          set shouldComponentUpdate(v) {}
        }
      `,
    },
    {
      code: `
        class YourComponent extends React.Component {
          static shouldComponentUpdate() {}
        }
      `,
    },
    // ---- Edge: paren-wrapped extends ----
    {
      code: `
        class YourComponent extends (React.Component) {
          shouldComponentUpdate() {}
        }
      `,
    },
    // ---- Edge: parens around decorator receiver — `@(reactMixin).decorate(PureRenderMixin)` ----
    {
      code: `
        @(reactMixin).decorate(PureRenderMixin)
        class DecoratedComponent extends Component {}
      `,
    },
    // ---- Edge: createReactClass with shorthand `shouldComponentUpdate()` ----
    {
      code: `
        createReactClass({
          shouldComponentUpdate() {}
        })
      `,
    },
    // ---- Edge: React.createClass (pragma + createClass) qualified call ----
    {
      code: `
        React.createClass({
          shouldComponentUpdate: function () {}
        })
      `,
    },
    // ---- Edge: createReactClass with PureRenderMixin among other mixins ----
    {
      code: `
        createReactClass({
          mixins: [SomeMixin, PureRenderMixin, AnotherMixin]
        })
      `,
    },
    // ---- Edge: paren-wrapped PureRenderMixin in mixins array ----
    {
      code: `
        createReactClass({
          mixins: [(PureRenderMixin)]
        })
      `,
    },
    // ---- Edge: createReactClass with ParenthesizedExpression around obj arg ----
    {
      code: `
        createReactClass(({
          mixins: [PureRenderMixin]
        }))
      `,
    },
    // ---- Edge: settings.react.createClass="myCreateClass" custom factory ----
    {
      code: `
        myCreateClass({
          shouldComponentUpdate: function() {}
        })
      `,
      settings: { react: { createClass: 'myCreateClass' } },
    },
    // ---- Edge: nested class — only the React-extending class is checked ----
    {
      code: `
        class Outer {
          method() {
            class Inner extends React.Component {
              shouldComponentUpdate() {}
            }
          }
        }
      `,
    },
    // ---- Edge: settings.react.pragma="Preact" + Preact.PureComponent ----
    {
      code: `
        class YourComponent extends Preact.PureComponent {}
      `,
      settings: { react: { pragma: 'Preact' } },
    },
    // ---- Real user: export default class extends React.PureComponent ----
    {
      code: `
        export default class extends React.PureComponent {}
      `,
    },
    // ---- Real user: export default class with SCU ----
    {
      code: `
        export default class C extends React.Component {
          shouldComponentUpdate() { return true; }
        }
      `,
    },
    // ---- Real user: export class ... with SCU ----
    {
      code: `
        export class C extends React.Component {
          shouldComponentUpdate() { return true; }
        }
      `,
    },
    // ---- Real user: TS access modifiers on SCU ----
    {
      code: `
        class C extends React.Component {
          public shouldComponentUpdate() { return true; }
        }
      `,
    },
    {
      code: `
        class C extends React.Component {
          protected shouldComponentUpdate() { return true; }
        }
      `,
    },
    // ---- Real user: TS override modifier on SCU ----
    {
      code: `
        class C extends React.Component {
          override shouldComponentUpdate() { return true; }
        }
      `,
    },
    // ---- Real user: SCU with explicit parameters and types ----
    {
      code: `
        class C extends React.Component<P, S> {
          shouldComponentUpdate(nextProps: P, nextState: S): boolean {
            return true;
          }
        }
      `,
    },
    // ---- Real user: namespace-imported React ----
    {
      code: `
        import * as React from "react";
        class C extends React.PureComponent {}
      `,
    },
    // ---- Real user: connect HOC wrapping a class with SCU ----
    {
      code: `
        class Inner extends React.Component {
          shouldComponentUpdate() { return true; }
        }
        const Connected = connect(mapStateToProps)(Inner);
      `,
    },
    // ---- Real user: TS overload signatures + implementation for SCU ----
    {
      code: `
        class C extends React.Component {
          shouldComponentUpdate(): boolean;
          shouldComponentUpdate(nextProps: any): boolean;
          shouldComponentUpdate() { return true; }
        }
      `,
    },
    // ---- Edge: TS abstract class with SCU ----
    {
      code: `
        abstract class C extends React.Component {
          shouldComponentUpdate() {}
        }
      `,
    },
    // ---- Edge: generic type args + SCU ----
    {
      code: `
        class C extends React.PureComponent<Props, State> {}
      `,
    },
    // ---- Walk-up: ObjectExpression with SCU inside class method ----
    {
      code: `
        class C extends React.Component {
          init() {
            const cfg = { shouldComponentUpdate() {} };
          }
        }
      `,
    },
    // ---- Walk-up: nested obj scu inside createReactClass ----
    {
      code: `
        createReactClass({
          config: {
            shouldComponentUpdate: function() {}
          }
        })
      `,
    },
    // ---- Walk-up: PureRender decorator on inner non-React class
    // silences outer React component. ----
    {
      code: `
        class C extends React.Component {
          build() {
            @reactMixin.decorate(PureRenderMixin)
            class Helper {}
          }
        }
      `,
    },
    // ---- Top-level SFC auto-silenced ----
    {
      code: `
        function Comp(props) { return <div />; }
      `,
    },
    {
      code: `
        export default function Comp(props) { return <div />; }
      `,
    },
    {
      code: `
        const Comp = (props) => <div />;
      `,
    },
    // ---- Settings: pragma=Preact, Preact.PureComponent silenced ----
    {
      code: `
        class C extends Preact.PureComponent {}
      `,
      settings: { react: { pragma: 'Preact' } },
    },
    // ---- Settings: pragma=Preact, React.Component no longer recognized ----
    {
      code: `
        class C extends React.Component {}
      `,
      settings: { react: { pragma: 'Preact' } },
    },
    // ---- Settings: createClass=myCreateClass + SCU ----
    {
      code: `
        myCreateClass({ shouldComponentUpdate() {} });
      `,
      settings: { react: { createClass: 'myCreateClass' } },
    },
  ],
  invalid: [
    // ---- Upstream invalid: empty React.Component class ----
    {
      code: `
        import React from "react";
        class YourComponent extends React.Component {}
      `,
      errors: [expectNoShouldComponentUpdateError([3, 9, 3, 55])],
    },
    // ---- Upstream invalid: methods without shouldComponentUpdate ----
    {
      code: `
        import React from "react";
        class YourComponent extends React.Component {
          handleClick() {}
          render() {
            return <div onClick={this.handleClick}>123</div>
          }
        }
      `,
      errors: [expectNoShouldComponentUpdateError([3, 9, 8, 10])],
    },
    // ---- Upstream invalid: class field arrow handler, no SCU ----
    {
      code: `
        import React from "react";
        class YourComponent extends React.Component {
          handleClick = () => {}
          render() {
            return <div onClick={this.handleClick}>123</div>
          }
        }
      `,
      errors: [expectNoShouldComponentUpdateError([3, 9, 8, 10])],
    },
    // ---- Upstream invalid: bare Component, empty body ----
    {
      code: `
        import React, {Component} from "react";
        class YourComponent extends Component {}
      `,
      errors: [expectNoShouldComponentUpdateError([3, 9, 3, 49])],
    },
    // ---- Upstream invalid: empty createReactClass ----
    {
      code: `
        import React from "react";
        createReactClass({})
      `,
      errors: [expectNoShouldComponentUpdateError([3, 26, 3, 28])],
    },
    // ---- Upstream invalid: createReactClass with non-PureRender mixin ----
    {
      code: `
        import React from "react";
        createReactClass({
          mixins: [RandomMixin]
        })
      `,
      errors: [expectNoShouldComponentUpdateError([3, 26, 5, 10])],
    },
    // ---- Upstream invalid: decorator with non-PureRender mixin argument ----
    {
      code: `
        @reactMixin.decorate(SomeOtherMixin)
        class DecoratedComponent extends Component {}
      `,
      errors: [expectNoShouldComponentUpdateError([2, 9, 3, 54])],
    },
    // ---- Upstream invalid: decorators not in allow-list ----
    {
      code: `
        @bar
        @pure
        @foo
        class DecoratedComponent extends Component {}
      `,
      options: [{ allowDecorators: ['renderPure', 'pureRender'] }],
      errors: [expectNoShouldComponentUpdateError([2, 9, 5, 54])],
    },
    // ---- Edge: anonymous ClassExpression — no name, still reported ----
    {
      code: `
        module.exports = class extends React.Component {};
      `,
      errors: [expectNoShouldComponentUpdateError([2, 26, 2, 58])],
    },
    // ---- Edge: shouldComponentUpdate as class field (not a method) — still reported ----
    {
      code: `
        class YourComponent extends React.Component {
          shouldComponentUpdate = () => true;
        }
      `,
      errors: [expectNoShouldComponentUpdateError([2, 9, 4, 10])],
    },
    // ---- Edge: shouldComponentUpdate as STRING-LITERAL key — non-Identifier never matches ----
    {
      code: `
        class YourComponent extends React.Component {
          "shouldComponentUpdate"() {}
        }
      `,
      errors: [expectNoShouldComponentUpdateError([2, 9, 4, 10])],
    },
    // ---- Edge: optional chain in decorator — does NOT match the PureRender shape ----
    {
      code: `
        @reactMixin?.decorate(PureRenderMixin)
        class DecoratedComponent extends Component {}
      `,
      errors: [expectNoShouldComponentUpdateError([2, 9, 3, 54])],
    },
    // ---- Edge: deeper member access in decorator — does NOT match ----
    {
      code: `
        @outer.reactMixin.decorate(PureRenderMixin)
        class DecoratedComponent extends Component {}
      `,
      errors: [expectNoShouldComponentUpdateError([2, 9, 3, 54])],
    },
    // ---- Edge: zero-arg decorate() — does NOT match ----
    {
      code: `
        @reactMixin.decorate()
        class DecoratedComponent extends Component {}
      `,
      errors: [expectNoShouldComponentUpdateError([2, 9, 3, 54])],
    },
    // ---- Edge: non-Identifier first arg — does NOT match ----
    {
      code: `
        @reactMixin.decorate(getMixin())
        class DecoratedComponent extends Component {}
      `,
      errors: [expectNoShouldComponentUpdateError([2, 9, 3, 54])],
    },
    // ---- Edge: wrong wrapper name — does NOT match ----
    {
      code: `
        @otherWrapper.decorate(PureRenderMixin)
        class DecoratedComponent extends Component {}
      `,
      errors: [expectNoShouldComponentUpdateError([2, 9, 3, 54])],
    },
    // ---- Edge: wrong method name — does NOT match ----
    {
      code: `
        @reactMixin.combine(PureRenderMixin)
        class DecoratedComponent extends Component {}
      `,
      errors: [expectNoShouldComponentUpdateError([2, 9, 3, 54])],
    },
    // ---- Edge: CallExpression form decorator with allowDecorators — does NOT match ----
    {
      code: `
        @pureRender()
        class DecoratedComponent extends Component {}
      `,
      options: [{ allowDecorators: ['pureRender'] }],
      errors: [expectNoShouldComponentUpdateError([2, 9, 3, 54])],
    },
    // ---- Edge: createReactClass mixins is NOT an ArrayExpression ----
    {
      code: `
        createReactClass({
          mixins: PureRenderMixin
        })
      `,
      errors: [expectNoShouldComponentUpdateError([2, 26, 4, 10])],
    },
    // ---- Edge: createReactClass mixins element is a CallExpression ----
    {
      code: `
        createReactClass({
          mixins: [resolveMixin()]
        })
      `,
      errors: [expectNoShouldComponentUpdateError([2, 26, 4, 10])],
    },
    // ---- Edge: settings.react.pragma="Preact" + Preact.Component without SCU ----
    {
      code: `
        class YourComponent extends Preact.Component {}
      `,
      settings: { react: { pragma: 'Preact' } },
      errors: [expectNoShouldComponentUpdateError([2, 9, 2, 56])],
    },
    // ---- Edge: nested ClassDeclaration inside non-component class body ----
    {
      code: `
        class Outer {
          method() {
            class Inner extends React.Component {}
            return Inner;
          }
        }
      `,
      errors: [expectNoShouldComponentUpdateError([4, 13, 4, 51])],
    },
    // ---- Edge: createReactClass with paren-wrapped object arg, empty body ----
    {
      code: `
        createReactClass(({}))
      `,
      errors: [expectNoShouldComponentUpdateError([2, 27, 2, 29])],
    },
    // ---- Edge: createReactClass with elision-only mixins array ----
    {
      code: `
        createReactClass({
          mixins: [,,,]
        })
      `,
      errors: [expectNoShouldComponentUpdateError([2, 26, 4, 10])],
    },
    // ---- Edge: anonymous ClassExpression in CallExpression arg position ----
    {
      code: `
        register(class extends React.Component {});
      `,
      errors: [expectNoShouldComponentUpdateError([2, 18, 2, 50])],
    },
    // ---- Real user: export default class without SCU ----
    {
      code: `
        export default class extends React.Component {}
      `,
      errors: [expectNoShouldComponentUpdateError([2, 24, 2, 56])],
    },
    // ---- Real user: export class without SCU ----
    {
      code: `
        export class YourComponent extends React.Component {}
      `,
      errors: [expectNoShouldComponentUpdateError([2, 16, 2, 62])],
    },
    // ---- Real user: export default abstract class without SCU ----
    {
      code: `
        export default abstract class extends React.Component {}
      `,
      errors: [expectNoShouldComponentUpdateError([2, 24, 2, 65])],
    },
    // ---- Real user: connect HOC wrapping inline ClassExpression without SCU ----
    {
      code: `
        const Connected = connect(mapStateToProps)(class extends React.Component {});
      `,
      errors: [expectNoShouldComponentUpdateError([2, 52, 2, 84])],
    },
    // ---- Real user: ClassExpression in expression statement ----
    {
      code: `
        (class extends React.Component {});
      `,
      errors: [expectNoShouldComponentUpdateError([2, 10, 2, 42])],
    },
    // ---- Real user: multiple createReactClass calls — only the missing
    // ones report. ----
    {
      code: `
        const A = createReactClass({});
        const B = createReactClass({ shouldComponentUpdate() {} });
        const C = createReactClass({ mixins: [RandomMixin] });
      `,
      errors: [
        expectNoShouldComponentUpdateError([2, 36, 2, 38]),
        expectNoShouldComponentUpdateError([4, 36, 4, 61]),
      ],
    },
    // ---- Real user: TS access modifier on a non-SCU method — class still
    // missing SCU, reports. ----
    {
      code: `
        class C extends React.Component {
          private otherMethod() {}
        }
      `,
      errors: [expectNoShouldComponentUpdateError([2, 9, 4, 10])],
    },
    // ---- Real user: const C = class extends React.Component {} without SCU ----
    {
      code: `
        const C = class extends React.Component {};
      `,
      errors: [expectNoShouldComponentUpdateError([2, 19, 2, 51])],
    },
    // ---- Real user: nested sibling components, both missing SCU ----
    {
      code: `
        function outer() {
          class A extends React.Component {}
          function inner() {
            class B extends React.Component {}
          }
        }
      `,
      errors: [
        expectNoShouldComponentUpdateError([3, 11, 3, 45]),
        expectNoShouldComponentUpdateError([5, 13, 5, 47]),
      ],
    },
    // ---- Locks in upstream: ClassExpression PureComponent variants are
    // ALWAYS reported. Upstream's PureComponent/decorator shortcut runs only
    // on ClassDeclaration. ----
    {
      code: `
        const X = class extends React.PureComponent {};
      `,
      errors: [expectNoShouldComponentUpdateError([2, 19, 2, 55])],
    },
    {
      code: `
        register(class extends React.PureComponent {});
      `,
      errors: [expectNoShouldComponentUpdateError([2, 18, 2, 54])],
    },
    {
      code: `
        module.exports = class extends React.PureComponent {};
      `,
      errors: [expectNoShouldComponentUpdateError([2, 26, 2, 62])],
    },
    {
      code: `
        const obj = { Comp: class extends React.PureComponent {} };
      `,
      errors: [expectNoShouldComponentUpdateError([2, 29, 2, 65])],
    },
    {
      code: `
        const Connected = connect(mapStateToProps)(class extends React.PureComponent {});
      `,
      errors: [expectNoShouldComponentUpdateError([2, 52, 2, 88])],
    },
    // ---- Walk-up: nested class scu absorbs SCU on inner; outer reports ----
    {
      code: `
        class Outer extends React.Component {
          build() {
            class Inner extends React.Component {
              shouldComponentUpdate() {}
            }
          }
        }
      `,
      errors: [expectNoShouldComponentUpdateError([2, 9, 8, 10])],
    },
    // ---- Walk-up: createReactClass with nested ClassExpression that has
    // SCU. Inner ClassExpression absorbs SCU; outer createReactClass empty,
    // reports. ----
    {
      code: `
        createReactClass({
          inner: class extends React.Component {
            shouldComponentUpdate() {}
          }
        })
      `,
      errors: [expectNoShouldComponentUpdateError([2, 26, 6, 10])],
    },
    // ---- Walk-up: top-level obj-with-SCU does NOT walk up to a class
    // declared in module scope. ----
    {
      code: `
        const cfg = { shouldComponentUpdate() {} };
        class A extends React.Component {}
      `,
      errors: [expectNoShouldComponentUpdateError([3, 9, 3, 43])],
    },
    // ---- SFC inside ClassDeclaration method body is reported (not auto-
    // silenced because `isFunctionInClass` blocks markSCU(self)). ----
    {
      code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          build() {
            function Inner(p) { return <div />; }
            return <Inner />;
          }
        }
      `,
      errors: [expectNoShouldComponentUpdateError([5, 13, 5, 50])],
    },
    {
      code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          build() {
            const Inner = (p) => <div />;
            return <Inner />;
          }
        }
      `,
      errors: [expectNoShouldComponentUpdateError([5, 27, 5, 41])],
    },
    // ---- Settings: pragma=Preact + Preact.Component without SCU ----
    {
      code: `
        class C extends Preact.Component {}
      `,
      settings: { react: { pragma: 'Preact' } },
      errors: [expectNoShouldComponentUpdateError([2, 9, 2, 44])],
    },
    // ---- Settings: createClass=myCreateClass empty obj ----
    {
      code: `
        myCreateClass({});
      `,
      settings: { react: { createClass: 'myCreateClass' } },
      errors: [expectNoShouldComponentUpdateError([2, 23, 2, 25])],
    },
  ],
});
