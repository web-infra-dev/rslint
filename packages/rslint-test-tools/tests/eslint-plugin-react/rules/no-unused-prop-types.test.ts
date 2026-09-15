import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unused-prop-types', {} as never, {
  valid: [
    {
      code: `class Hello extends React.Component { render() { return <div>{this.props.name}</div>; } } Hello.propTypes = { name: PropTypes.string };`,
    },
    {
      code: `function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes = { name: PropTypes.string };`,
    },
  ],
  invalid: [
    {
      code: `class Hello extends React.Component { render() { return <div />; } } Hello.propTypes = { name: PropTypes.string };`,
      errors: [{ messageId: 'unusedPropType' }],
    },
  ],
});
