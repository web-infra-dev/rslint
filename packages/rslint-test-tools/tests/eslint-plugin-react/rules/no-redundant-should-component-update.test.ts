import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
import { diagnostic } from '../expectations';

function expectNoShouldCompUpdateError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'noShouldCompUpdate',
    'Foo does not need shouldComponentUpdate when extending React.PureComponent.',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-redundant-should-component-update', {
  valid: [
    // ---- Upstream valid: shouldComponentUpdate on plain React.Component ----
    {
      code: `
        class Foo extends React.Component {
          shouldComponentUpdate() {
            return true;
          }
        }
      `,
    },
    // ---- Upstream valid: class field arrow shouldComponentUpdate on Component ----
    {
      code: `
        class Foo extends React.Component {
          shouldComponentUpdate = () => {
            return true;
          }
        }
      `,
    },
    // ---- Upstream valid: nested class expression on plain Component ----
    {
      code: `
        function Foo() {
          return class Bar extends React.Component {
            shouldComponentUpdate() {
              return true;
            }
          };
        }
      `,
    },
    // ---- Edge: PureComponent without shouldComponentUpdate — clean ----
    {
      code: `
        class Foo extends React.PureComponent {
          render() { return null; }
        }
      `,
    },
    // ---- Edge: shouldComponentUpdate as STRING-LITERAL key — non-Identifier never matches ----
    {
      code: `
        class Foo extends React.PureComponent {
          "shouldComponentUpdate"() { return true; }
        }
      `,
    },
    // ---- Edge: shouldComponentUpdate as COMPUTED key — non-Identifier never matches ----
    {
      code: '\n        class Foo extends React.PureComponent {\n          [`shouldComponentUpdate`]() { return true; }\n        }\n      ',
    },
    // ---- Edge: extends some other namespace's PureComponent — strict regex ----
    {
      code: `
        class Foo extends Other.PureComponent {
          shouldComponentUpdate() { return true; }
        }
      `,
    },
  ],
  invalid: [
    // ---- Upstream invalid: ClassDeclaration extends React.PureComponent ----
    {
      code: `
        class Foo extends React.PureComponent {
          shouldComponentUpdate() {
            return true;
          }
        }
      `,
      errors: [expectNoShouldCompUpdateError([2, 9, 6, 10])],
    },
    // ---- Upstream invalid: bare PureComponent ----
    {
      code: `
        class Foo extends PureComponent {
          shouldComponentUpdate() {
            return true;
          }
        }
      `,
      errors: [expectNoShouldCompUpdateError([2, 9, 6, 10])],
    },
    // ---- Upstream invalid: class field arrow shouldComponentUpdate ----
    {
      code: `
        class Foo extends React.PureComponent {
          shouldComponentUpdate = () => {
            return true;
          }
        }
      `,
      errors: [expectNoShouldCompUpdateError([2, 9, 6, 10])],
    },
    // ---- Upstream invalid: nested ClassExpression with own name (Bar) ----
    {
      code: `
        function Foo() {
          return class Bar extends React.PureComponent {
            shouldComponentUpdate() {
              return true;
            }
          };
        }
      `,
      errors: [
        diagnostic(
          'noShouldCompUpdate',
          'Bar does not need shouldComponentUpdate when extending React.PureComponent.',
          [3, 18, 7, 12],
        ),
      ],
    },
    // ---- Upstream invalid: nested ClassExpression bare PureComponent ----
    {
      code: `
        function Foo() {
          return class Bar extends PureComponent {
            shouldComponentUpdate() {
              return true;
            }
          };
        }
      `,
      errors: [
        diagnostic(
          'noShouldCompUpdate',
          'Bar does not need shouldComponentUpdate when extending React.PureComponent.',
          [3, 18, 7, 12],
        ),
      ],
    },
    // ---- Upstream invalid: anonymous ClassExpression assigned to var (name from binding) ----
    {
      code: `
        var Foo = class extends PureComponent {
          shouldComponentUpdate() {
            return true;
          }
        }
      `,
      errors: [expectNoShouldCompUpdateError([2, 19, 6, 10])],
    },
    // ---- Edge: parens around extends expression — paren-transparent matching ESTree ----
    {
      code: `
        class Foo extends (React.PureComponent) {
          shouldComponentUpdate() { return true; }
        }
      `,
      errors: [expectNoShouldCompUpdateError([2, 9, 4, 10])],
    },
  ],
});
