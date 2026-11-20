import { RuleTester, type ExpectedSuggestion } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectMissingRefParameterError(
  range: [line: number, column: number, endLine: number, endColumn: number],
  suggestions?: ExpectedSuggestion[],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingRefParameter',
    message:
      'forwardRef is used with this component but no ref parameter is set',
    line,
    column,
    endLine,
    endColumn,
    ...(suggestions === undefined ? {} : { suggestions }),
  };
}

const ruleTester = new RuleTester(fixtureConfig);

const message =
  'forwardRef is used with this component but no ref parameter is set';

ruleTester.run('forward-ref-uses-ref', {
  valid: [
    {
      code: `
        import { forwardRef } from 'react'
        forwardRef((props, ref) => {
          return null;
        });
      `,
    },
    {
      code: `
        import { forwardRef } from 'react'
        forwardRef((props, ref) => null);
      `,
    },
    {
      code: `
        import { forwardRef } from 'react'
        forwardRef(function (props, ref) {
          return null;
        });
      `,
    },
    {
      code: `
        import { forwardRef } from 'react'
        forwardRef(function Component(props, ref) {
          return null;
        });
      `,
    },
    {
      code: `
        import * as React from 'react'
        React.forwardRef((props, ref) => {
          return null;
        });
      `,
    },
    {
      code: `
        import * as React from 'react'
        React.forwardRef((props, ref) => null);
      `,
    },
    {
      code: `
        import * as React from 'react'
        React.forwardRef(function (props, ref) {
          return null;
        });
      `,
    },
    {
      code: `
        import * as React from 'react'
        React.forwardRef(function Component(props, ref) {
          return null;
        });
      `,
    },
    {
      code: `
        import * as React from 'react'
        function Component(props) {
          return null;
        };
      `,
    },
    {
      code: `
        import * as React from 'react'
        (props) => null;
      `,
    },
  ],
  invalid: [
    {
      code: `
        import { forwardRef } from 'react'
        forwardRef((props) => {
          return null;
        });
      `,
      errors: [
        expectMissingRefParameterError(
          [3, 20, 5, 10],
          [
            {
              messageId: 'addRefParameter',
              desc: 'Add a ref parameter',
              output:
                "\n        import { forwardRef } from 'react'\n        forwardRef((props, ref) => {\n          return null;\n        });\n      ",
            },
            {
              messageId: 'removeForwardRef',
              desc: 'Remove forwardRef wrapper',
              output:
                "\n        import { forwardRef } from 'react'\n        (props) => {\n          return null;\n        };\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        import { forwardRef } from 'react'
        forwardRef(props => {
          return null;
        });
      `,
      errors: [
        expectMissingRefParameterError(
          [3, 20, 5, 10],
          [
            {
              messageId: 'addRefParameter',
              desc: 'Add a ref parameter',
              output:
                "\n        import { forwardRef } from 'react'\n        forwardRef((props, ref) => {\n          return null;\n        });\n      ",
            },
            {
              messageId: 'removeForwardRef',
              desc: 'Remove forwardRef wrapper',
              output:
                "\n        import { forwardRef } from 'react'\n        props => {\n          return null;\n        };\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        import * as React from 'react'
        React.forwardRef((props) => null);
      `,
      errors: [
        expectMissingRefParameterError(
          [3, 26, 3, 41],
          [
            {
              messageId: 'addRefParameter',
              desc: 'Add a ref parameter',
              output:
                "\n        import * as React from 'react'\n        React.forwardRef((props, ref) => null);\n      ",
            },
            {
              messageId: 'removeForwardRef',
              desc: 'Remove forwardRef wrapper',
              output:
                "\n        import * as React from 'react'\n        (props) => null;\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        import { forwardRef } from 'react'
        const Component = forwardRef(function (props) {
          return null;
        });
      `,
      errors: [
        expectMissingRefParameterError(
          [3, 38, 5, 10],
          [
            {
              messageId: 'addRefParameter',
              desc: 'Add a ref parameter',
              output:
                "\n        import { forwardRef } from 'react'\n        const Component = forwardRef(function (props, ref) {\n          return null;\n        });\n      ",
            },
            {
              messageId: 'removeForwardRef',
              desc: 'Remove forwardRef wrapper',
              output:
                "\n        import { forwardRef } from 'react'\n        const Component = function (props) {\n          return null;\n        };\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        import * as React from 'react'
        React.forwardRef(function Component(props) {
          return null;
        });
      `,
      errors: [
        expectMissingRefParameterError(
          [3, 26, 5, 10],
          [
            {
              messageId: 'addRefParameter',
              desc: 'Add a ref parameter',
              output:
                "\n        import * as React from 'react'\n        React.forwardRef(function Component(props, ref) {\n          return null;\n        });\n      ",
            },
            {
              messageId: 'removeForwardRef',
              desc: 'Remove forwardRef wrapper',
              output:
                "\n        import * as React from 'react'\n        function Component(props) {\n          return null;\n        };\n      ",
            },
          ],
        ),
      ],
    },
  ],
});
