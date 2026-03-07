import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-filename-extension', {} as never, {
  valid: [
    {
      code: `const App = () => <div>Hello</div>`,
      filename: 'src/virtual.tsx',
      options: [{ extensions: ['.tsx'] }],
    },
    {
      code: `const App = () => <div />`,
      filename: 'src/virtual.tsx',
      options: [{ extensions: ['.tsx', '.jsx'] }],
    },
  ],
  invalid: [
    {
      // .tsx file has no JSX, but extension is .tsx and allow is "as-needed"
      code: `const x = 1;`,
      filename: 'src/virtual.tsx',
      options: [{ extensions: ['.tsx'], allow: 'as-needed' }],
      errors: [
        { message: "Only files containing JSX may use the extension '.tsx'" },
      ],
    },
  ],
});
