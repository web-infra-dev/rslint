import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectNoIsMountedError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noIsMounted',
    message: 'Do not use isMounted',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-is-mounted', {
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
      errors: [expectNoIsMountedError([4, 18, 4, 32])],
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
      errors: [expectNoIsMountedError([4, 18, 4, 32])],
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
      errors: [expectNoIsMountedError([4, 18, 4, 32])],
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
      errors: [expectNoIsMountedError([4, 20, 4, 34])],
    },
    {
      code: `
        class Hello extends React.Component {
          set mounted(_v) {
            this.isMounted();
          }
        };
      `,
      errors: [expectNoIsMountedError([4, 13, 4, 27])],
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
      errors: [expectNoIsMountedError([5, 13, 5, 27])],
    },
    {
      code: `
        var Hello = {
          someMethod() {
            this.isMounted();
          }
        };
      `,
      errors: [expectNoIsMountedError([4, 13, 4, 27])],
    },
  ],
});
