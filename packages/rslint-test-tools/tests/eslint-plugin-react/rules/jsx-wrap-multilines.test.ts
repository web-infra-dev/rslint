import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-wrap-multilines', {} as never, {
  valid: [
    {
      code: `function App() { return <div>single</div>; }`,
    },
    {
      code: `function App() {\n  return (\n    <div>\n      <span />\n    </div>\n  );\n}`,
    },
  ],
  invalid: [
    {
      code: `function App() {\n  return <div>\n    <span />\n  </div>;\n}`,
      errors: [{ message: 'Missing parentheses around multilines JSX' }],
    },
  ],
});
