import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-useless-return', {
  valid: [
    'function foo() { return 5; }',
    'function foo() { return null; }',
    'function foo() { return doSomething(); }',
    `
      function foo() {
        if (bar) {
          doSomething();
          return;
        } else {
          doSomethingElse();
        }
        qux();
      }
    `,
    `
      function foo() {
        switch (bar) {
          case 1:
            doSomething();
            return;
          default:
            doSomethingElse();
        }
      }
    `,
    `
      function foo() {
        switch (bar) {
          default:
            doSomething();
            return;
          case 1:
            doSomethingElse();
        }
      }
    `,
    `
      function foo() {
        switch (bar) {
          case 1:
            if (a) {
              doSomething();
              return;
            } else {
              doSomething();
              return;
            }
          default:
            doSomethingElse();
        }
      }
    `,
    `
      function foo() {
        for (var foo = 0; foo < 10; foo++) {
          return;
        }
      }
    `,
    `
      function foo() {
        for (var foo in bar) {
          return;
        }
      }
    `,
    `
      function foo() {
        try {
          return 5;
        } finally {
          return;
        }
      }
    `,
    `
      function foo() {
        try {
          bar();
          return;
        } catch (err) {}
        baz();
      }
    `,
    `
      function foo() {
        if (something) {
          try {
            bar();
            return;
          } catch (err) {}
        }
        baz();
      }
    `,
    `
      function foo() {
        return;
        doSomething();
      }
    `,
    `
      function foo() {
        for (var foo of bar) return;
      }
    `,
    '() => { if (foo) return; bar(); }',
    '() => 5',
    '() => { return; doSomething(); }',
    'if (foo) { return; } doSomething();',
    `
      function foo() {
        if (bar) return;
        return baz;
      }
    `,
    `
      function foo() {
        if (bar) {
          return;
        }
        return baz;
      }
    `,
    `
      function foo() {
        if (bar) baz();
        else return;
        return 5;
      }
    `,
    `
      function foo() {
        return;
        while (foo) return;
        foo;
      }
    `,
    `
      try {
        throw new Error('foo');
        while (false);
      } catch (err) {}
    `,
    `
      function foo(arg) {
        throw new Error('Debugging...');
        if (!arg) {
          return;
        }
        console.log(arg);
      }
    `,
    `
      function foo() {
        try {
          bar();
          return;
        } finally {
          baz();
        }
        qux();
      }
    `,
  ],
  invalid: [
    {
      code: 'function foo() { return; }',
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: 'function foo() { doSomething(); return; }',
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: 'function foo() { if (condition) { bar(); return; } else { baz(); } }',
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: 'function foo() { if (foo) return; }',
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: 'function foo() { bar(); return/**/; }',
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: 'function foo() { bar(); return//\n; }',
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: 'foo(); return;',
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: 'if (foo) { bar(); return; } else { baz(); }',
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: `
        function foo() {
          if (foo) {
            return;
          }
          return;
        }
      `,
      errors: [
        { messageId: 'unnecessaryReturn' },
        { messageId: 'unnecessaryReturn' },
      ],
    },
    {
      code: `
        function foo() {
          switch (bar) {
            case 1:
              doSomething();
            default:
              doSomethingElse();
              return;
          }
        }
      `,
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: `
        function foo() {
          switch (bar) {
            default:
              doSomething();
            case 1:
              doSomething();
              return;
          }
        }
      `,
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: `
        function foo() {
          switch (bar) {
            case 1:
              if (a) {
                doSomething();
                return;
              }
              break;
            default:
              doSomethingElse();
          }
        }
      `,
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: `
        function foo() {
          switch (bar) {
            case 1:
              if (a) {
                doSomething();
                return;
              } else {
                doSomething();
              }
              break;
            default:
              doSomethingElse();
          }
        }
      `,
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: `
        function foo() {
          switch (bar) {
            case 1:
              if (a) {
                doSomething();
                return;
              }
            default:
          }
        }
      `,
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: `
        function foo() {
          try {} catch (err) { return; }
        }
      `,
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: `
        function foo() {
          try {
            foo();
            return;
          } catch (err) {
            return 5;
          }
        }
      `,
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: `
        function foo() {
          if (something) {
            try {
              bar();
              return;
            } catch (err) {}
          }
        }
      `,
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: `
        function foo() {
          try {
            return;
          } catch (err) {
            foo();
          }
        }
      `,
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: `
        function foo() {
          try {
            return;
          } finally {
            bar();
          }
        }
      `,
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: `
        function foo() {
          try {
            bar();
          } catch (e) {
            try {
              baz();
              return;
            } catch (e) {
              qux();
            }
          }
        }
      `,
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: `
        function foo() {
          try {} finally {}
          return;
        }
      `,
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: `
        function foo() {
          try {
            return 5;
          } finally {
            function bar() {
              return;
            }
          }
        }
      `,
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: '() => { return; }',
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
    {
      code: 'function foo() { return; return; }',
      errors: [{ messageId: 'unnecessaryReturn' }],
    },
  ],
});
