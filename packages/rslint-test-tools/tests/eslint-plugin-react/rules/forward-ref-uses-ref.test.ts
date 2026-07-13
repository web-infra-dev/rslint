import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const message =
  'forwardRef is used with this component but no ref parameter is set';

ruleTester.run('forward-ref-uses-ref', {} as never, {
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
      errors: [{ message }],
    },
    {
      code: `
        import { forwardRef } from 'react'
        forwardRef(props => {
          return null;
        });
      `,
      errors: [{ message }],
    },
    {
      code: `
        import * as React from 'react'
        React.forwardRef((props) => null);
      `,
      errors: [{ message }],
    },
    {
      code: `
        import { forwardRef } from 'react'
        const Component = forwardRef(function (props) {
          return null;
        });
      `,
      errors: [{ message }],
    },
    {
      code: `
        import * as React from 'react'
        React.forwardRef(function Component(props) {
          return null;
        });
      `,
      errors: [{ message }],
    },
  ],
});
