import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const message = (name: string) =>
  `React component ${name} must not be in a namespace, as React does not support them`;

ruleTester.run('no-namespace', {} as never, {
  valid: [
    // ---- Upstream valid: JSX elements ----
    { code: '<testcomponent />' },
    { code: '<testComponent />' },
    { code: '<test_component />' },
    { code: '<TestComponent />' },
    { code: '<object.testcomponent />' },
    { code: '<object.testComponent />' },
    { code: '<object.test_component />' },
    { code: '<object.TestComponent />' },
    { code: '<Object.testcomponent />' },
    { code: '<Object.testComponent />' },
    { code: '<Object.test_component />' },
    { code: '<Object.TestComponent />' },

    // ---- Upstream valid: React.createElement ----
    { code: 'React.createElement("testcomponent")' },
    { code: 'React.createElement("testComponent")' },
    { code: 'React.createElement("test_component")' },
    { code: 'React.createElement("TestComponent")' },
    { code: 'React.createElement("object.testcomponent")' },
    { code: 'React.createElement("object.testComponent")' },
    { code: 'React.createElement("object.test_component")' },
    { code: 'React.createElement("object.TestComponent")' },
    { code: 'React.createElement("Object.testcomponent")' },
    { code: 'React.createElement("Object.testComponent")' },
    { code: 'React.createElement("Object.test_component")' },
    { code: 'React.createElement("Object.TestComponent")' },
    { code: 'React.createElement(null)' },
    { code: 'React.createElement(true)' },
    { code: 'React.createElement({})' },
  ],
  invalid: [
    // ---- Upstream invalid: lower-case namespace ----
    ...[
      'ns:testcomponent',
      'ns:testComponent',
      'ns:test_component',
      'ns:TestComponent',
    ].flatMap((name) => [
      { code: `<${name} />`, errors: [{ message: message(name) }] },
      {
        code: `React.createElement("${name}")`,
        errors: [{ message: message(name) }],
      },
    ]),
    // ---- Upstream invalid: upper-case namespace ----
    ...[
      'Ns:testcomponent',
      'Ns:testComponent',
      'Ns:test_component',
      'Ns:TestComponent',
    ].flatMap((name) => [
      { code: `<${name} />`, errors: [{ message: message(name) }] },
      {
        code: `React.createElement("${name}")`,
        errors: [{ message: message(name) }],
      },
    ]),
  ],
});
