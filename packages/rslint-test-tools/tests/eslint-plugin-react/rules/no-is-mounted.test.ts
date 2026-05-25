import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-is-mounted', {} as never, {
  valid: [
    // ---- Upstream valid cases ----
    {
      code: `
        var Hello = function() {
        };
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello</div>;
          }
        });
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            someNonMemberFunction(arg);
            this.someFunc = this.isMounted;
          },
          render: function() {
            return <div>Hello</div>;
          }
        });
      `,
    },
    {
      code: `
        class Hello extends React.Component {
          notIsMounted() {}
          render() {
            this.notIsMounted();
            return <div>Hello</div>;
          }
        };
      `,
    },
    // ---- Edge: bracket notation is NOT matched ----
    {
      code: `
        class Hello extends React.Component {
          someMethod() {
            if (!this['isMounted']()) { return; }
          }
        };
      `,
    },
    // ---- Edge: top-level, outside any property / method ----
    {
      code: `this.isMounted();`,
    },
  ],
  invalid: [
    {
      code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            if (!this.isMounted()) {
              return;
            }
          },
          render: function() {
            return <div>Hello</div>;
          }
        });
      `,
      errors: [{ messageId: 'noIsMounted' }],
    },
    {
      code: `
        var Hello = createReactClass({
          someMethod: function() {
            if (!this.isMounted()) {
              return;
            }
          },
          render: function() {
            return <div onClick={this.someMethod.bind(this)}>Hello</div>;
          }
        });
      `,
      errors: [{ messageId: 'noIsMounted' }],
    },
    {
      code: `
        class Hello extends React.Component {
          someMethod() {
            if (!this.isMounted()) {
              return;
            }
          }
          render() {
            return <div onClick={this.someMethod.bind(this)}>Hello</div>;
          }
        };
      `,
      errors: [{ messageId: 'noIsMounted' }],
    },
    // ---- Edge: class getter / setter / constructor / object shorthand method ----
    {
      code: `
        class Hello extends React.Component {
          get mounted() {
            return this.isMounted();
          }
        };
      `,
      errors: [{ messageId: 'noIsMounted' }],
    },
    {
      code: `
        class Hello extends React.Component {
          set mounted(_v) {
            this.isMounted();
          }
        };
      `,
      errors: [{ messageId: 'noIsMounted' }],
    },
    {
      code: `
        class Hello extends React.Component {
          constructor() {
            super();
            this.isMounted();
          }
        };
      `,
      errors: [{ messageId: 'noIsMounted' }],
    },
    {
      code: `
        var Hello = {
          someMethod() {
            this.isMounted();
          }
        };
      `,
      errors: [{ messageId: 'noIsMounted' }],
    },
  ],
});
