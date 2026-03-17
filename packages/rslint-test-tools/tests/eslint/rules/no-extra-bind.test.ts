import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-extra-bind', {
  valid: [
    'var a = function(b: any) { return b }.bind(c, d)',
    'var a = function() { this.b }.bind(c)',
    'var a = f.bind(a)',
    'var a = function() { return this.b }.bind(c)',
  ],
  invalid: [
    {
      code: 'var a = function() { return 1; }.bind(b)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() { return 1; }.bind(this)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() { (function(){ this.c }) }.bind(b)',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
