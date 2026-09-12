import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const multilineRequire = [{ singleline: 'consistent', multiline: 'require' }];

ruleTester.run('jsx-curly-newline', {} as never, {
  valid: [
    { code: '<div>{foo}</div>', options: ['consistent'] },
    { code: '<div>{\nfoo\n}</div>', options: ['consistent'] },
    { code: '<div>{ foo &&\nfoo.bar }</div>', options: ['never'] },
    { code: '<div>{\nfoo\n}</div>', options: multilineRequire },
    { code: '<div foo={bar} />', options: multilineRequire },
  ],
  invalid: [
    {
      code: '<div>{ foo\n}</div>',
      options: ['consistent'],
      output: '<div>{ foo}</div>',
      errors: [{ messageId: 'unexpectedBefore' }],
    },
    {
      code: '<div>{\nfoo}</div>',
      options: multilineRequire,
      output: '<div>{\nfoo\n}</div>',
      errors: [{ messageId: 'expectedBefore' }],
    },
    {
      code: '<div>{\nfoo\n}</div>',
      options: ['never'],
      output: '<div>{foo}</div>',
      errors: [
        { messageId: 'unexpectedAfter' },
        { messageId: 'unexpectedBefore' },
      ],
    },
    {
      code: '<div>{ foo &&\nbar }</div>',
      options: multilineRequire,
      output: '<div>{\n foo &&\nbar \n}</div>',
      errors: [{ messageId: 'expectedAfter' }, { messageId: 'expectedBefore' }],
    },
  ],
});
