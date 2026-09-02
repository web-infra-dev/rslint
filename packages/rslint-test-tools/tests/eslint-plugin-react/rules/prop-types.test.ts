import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prop-types', {} as never, {
  valid: [
    {
      code: `function Hello({ name }) { return <div>{name}</div>; } Hello.propTypes = { name: PropTypes.string };`,
    },
    { code: `function Hello() { return <div>Hello</div>; }` },
  ],
  invalid: [
    {
      code: `function Hello({ name }) { return <div>{name}</div>; }`,
      errors: [{ messageId: 'missingPropType' }],
    },
  ],
});
