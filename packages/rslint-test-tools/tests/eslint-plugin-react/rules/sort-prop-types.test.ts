import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('sort-prop-types', {} as never, {
  valid: [
    {
      code: `Component.propTypes = { a: PropTypes.string, b: PropTypes.number };`,
    },
    {
      code: `Component.propTypes = { z: PropTypes.string, a: PropTypes.number };`,
      options: [{ noSortAlphabetically: true }],
    },
    {
      code: `Component.propTypes = { a: PropTypes.string, onChange: PropTypes.func };`,
      options: [{ callbacksLast: true }],
    },
  ],
  invalid: [
    {
      code: `Component.propTypes = { z: PropTypes.string, a: PropTypes.number };`,
      errors: [{ messageId: 'propsNotSorted' }],
    },
    {
      code: `Component.propTypes = { onChange: PropTypes.func, name: PropTypes.string };`,
      options: [{ callbacksLast: true }],
      errors: [{ messageId: 'callbackPropsLast' }],
    },
  ],
});
