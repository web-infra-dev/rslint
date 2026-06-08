import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const message = 'Each then() should return a value or throw';

ruleTester.run('always-return', {} as never, {
  valid: [
    { code: 'hey.then(x => x)' },
    { code: 'hey.then(x => ({}))' },
    { code: 'hey.then(x => { return; })' },
    { code: 'hey.then(x => { return x ? x.id : null })' },
    { code: 'hey.then(x => { return x * 10 })' },
    { code: 'hey.then(x => { process.exit(0); })' },
    { code: 'hey.then(x => { process.abort(); })' },
    { code: 'hey.then(function() { return 42; })' },
    { code: 'hey.then(function() { return new Promise(); })' },
    { code: 'hey.then(function() { return "x"; }).then(doSomethingWicked)' },
    { code: 'hey.then(x => x).then(function() { return "3" })' },
    { code: 'hey.then(function() { throw new Error("msg"); })' },
    {
      code: 'hey.then(function(x) { if (!x) { throw new Error("no x"); } return x; })',
    },
    {
      code: 'hey.then(function(x) { if (x) { return x; } throw new Error("no x"); })',
    },
    {
      code: 'hey.then(function(x) { if (x) { process.exit(0); } throw new Error("no x"); })',
    },
    {
      code: 'hey.then(function(x) { if (x) { process.abort(); } throw new Error("no x"); })',
    },
    { code: 'hey.then(x => { throw new Error("msg"); })' },
    {
      code: 'hey.then(x => { if (!x) { throw new Error("no x"); } return x; })',
    },
    {
      code: 'hey.then(x => { if (x) { return x; } throw new Error("no x"); })',
    },
    { code: 'hey.then(x => { var f = function() { }; return f; })' },
    { code: 'hey.then(x => { if (x) { return x; } else { return x; } })' },
    { code: 'hey.then(x => { return x; var y = "unreachable"; })' },
    { code: 'hey.then(x => { return x; return "unreachable"; })' },
    { code: 'hey.then(x => { return; }, err=>{ log(err); })' },
    { code: 'hey.then(x => { return x && x(); }, err=>{ log(err); })' },
    { code: 'hey.then(x => { return x.y || x(); }, err=>{ log(err); })' },
    {
      code: `hey.then(x => {
        return anotherFunc({
          nested: {
            one: x === 1 ? 1 : 0,
            two: x === 2 ? 1 : 0
          }
        })
      })`,
    },
    {
      code: `hey.then(({x, y}) => {
        if (y) {
          throw new Error(x || y)
        }
        return x
      })`,
    },
    {
      code: 'hey.then(x => { console.log(x) })',
      options: [{ ignoreLastCallback: true }],
    },
    {
      code: 'if(foo) { hey.then(x => { console.log(x) }) }',
      options: [{ ignoreLastCallback: true }],
    },
    {
      code: 'void hey.then(x => { console.log(x) })',
      options: [{ ignoreLastCallback: true }],
    },
    {
      code: `async function foo() {
        await hey.then(x => { console.log(x) })
      }`,
      options: [{ ignoreLastCallback: true }],
    },
    {
      code: 'hey?.then(x => { console.log(x) })',
      options: [{ ignoreLastCallback: true }],
    },
    {
      code: 'foo = (hey.then(x => { console.log(x) }), 42)',
      options: [{ ignoreLastCallback: true }],
    },
    {
      code: '(42, hey.then(x => { console.log(x) }))',
      options: [{ ignoreLastCallback: true }],
    },
    {
      code: `hey
        .then(x => { console.log(x) })
        .catch(e => console.error(e))`,
      options: [{ ignoreLastCallback: true }],
    },
    {
      code: `hey
        .then(x => { console.log(x) })
        .catch(e => console.error(e))
        .finally(() => console.error('end'))`,
      options: [{ ignoreLastCallback: true }],
    },
    {
      code: `hey
        .then(x => { console.log(x) })
        .finally(() => console.error('end'))`,
      options: [{ ignoreLastCallback: true }],
    },
    { code: 'hey.then(x => { globalThis = x })' },
    { code: 'hey.then(x => { globalThis[a] = x })' },
    { code: 'hey.then(x => { globalThis.a = x })' },
    { code: 'hey.then(x => { globalThis.a.n = x })' },
    { code: 'hey.then(x => { globalThis[12] = x })' },
    { code: 'hey.then(x => { globalThis[\'12\']["test"] = x })' },
    {
      code: "hey.then(x => { window['x'] = x })",
      options: [{ ignoreAssignmentVariable: ['globalThis', 'window'] }],
    },

    // Additional port-shape coverage.
    { code: "hey['then'](x => { console.log(x) })" },
    { code: 'hey.then(ok, x => { console.log(x) })' },
    { code: 'hey.then((function() { return 1; }))' },
    { code: 'hey.then(x => void console.log(x))' },

    // Infinite loops without an exiting break should be terminal even when the body falls through.
    { code: 'hey.then(function() { while (true) {} })' },
    { code: 'hey.then(function(x) { while (true) { if (x) return 1; } })' },
    { code: 'hey.then(function() { while (true) { x++; } })' },
    { code: 'hey.then(function() { do {} while (true) })' },
    { code: 'hey.then(function() { while (1) { return 1; } })' },
    { code: "hey.then(function() { while ('truthy') { return 1; } })" },

    // coverage gaps surfaced in review: for / labeled / do-while(false)
    { code: 'hey.then(function() { for (;;) { return 1; } })' },
    { code: 'hey.then(function() { outer: { return 1; } })' },
    { code: 'hey.then(function() { do { return 1; } while (false) })' },

    // switch: a conditional break must not be masked, yet these still terminate
    {
      code: 'hey.then(function(x) { switch (x) { case 1: { return 1; } default: return 2; } })',
    },
    {
      code: 'hey.then(function(x) { switch (x) { case 1: if (y) return 1; else return 2; default: return 3; } })',
    },
    {
      code: 'hey.then(function(x) { switch (x) { case 1: if (y) return 1; default: return 2; } })',
    },
    {
      code: 'hey.then(function(x) { switch (x) { case 1: while (true) { break; } return "a"; default: return "b"; } })',
    },

    // ignoreAssignmentVariable: compound assignment to an ignored var (default globalThis)
    { code: 'hey.then(x => { globalThis.a += x })' },
    { code: 'hey.then(x => { globalThis.a ||= x })' },
    { code: 'hey.then(x => { globalThis.a ??= x })' },
    {
      code: 'hey.then(x => { window.a += x })',
      options: [{ ignoreAssignmentVariable: ['window'] }],
    },

    // try/catch: catch unreachable because the try block cannot throw
    { code: 'hey.then(function() { try { return 1 } catch (e) { log(e) } })' },
    {
      code: 'hey.then(function() { try { return {a: 1} } catch (e) { log(e) } })',
    },
    {
      code: 'hey.then(function() { try { return 1; foo() } catch (e) { log(e) } })',
    },
  ],

  invalid: [
    { code: 'hey.then(x => {})', errors: [{ message }] },
    { code: 'hey.then(function() { })', errors: [{ message }] },
    { code: 'hey.then(function() { }).then(x)', errors: [{ message }] },
    {
      code: 'hey.then(function() { }).then(function() { })',
      errors: [{ message }, { message }],
    },
    {
      code: 'hey.then(function() { return; }).then(function() { })',
      errors: [{ message }],
    },
    {
      code: 'hey.then(function() { doSomethingWicked(); })',
      errors: [{ message }],
    },
    {
      code: 'hey.then(function() { if (x) { return x; } })',
      errors: [{ message }],
    },
    {
      code: 'hey.then(function() { if (x) { return x; } else { }})',
      errors: [{ message }],
    },
    {
      code: 'hey.then(function() { if (x) { } else { return x; }})',
      errors: [{ message }],
    },
    {
      code: 'hey.then(function() { if (x) { process.chdir(); } else { return x; }})',
      errors: [{ message }],
    },
    {
      code: 'hey.then(function() { if (x) { return you.then(function() { return x; }); } })',
      errors: [{ message }],
    },
    { code: 'hey.then( x => { x ? x.id : null })', errors: [{ message }] },
    {
      code: 'hey.then(function(x) { x ? x.id : null })',
      errors: [{ message }],
    },
    {
      code: `(function() {
        return hey.then(x => {
          anotherFunc({
            nested: {
              one: x === 1 ? 1 : 0,
              two: x === 2 ? 1 : 0
            }
          })
        })
      })()`,
      errors: [{ message }],
    },
    {
      code: `hey.then(({x, y}) => {
        if (y) {
          throw new Error(x || y)
        }
      })`,
      errors: [{ message }],
    },
    {
      code: `hey.then(({x, y}) => {
        if (y) {
          return x
        }
      })`,
      errors: [{ message }],
    },
    {
      code: `hey
        .then(function(x) { console.log(x) /* missing return here */ })
        .then(function(y) { console.log(y) /* no error here */ })`,
      options: [{ ignoreLastCallback: true }],
      errors: [{ message }],
    },
    {
      code: 'const foo = hey.then(function(x) {});',
      options: [{ ignoreLastCallback: true }],
      errors: [{ message }],
    },
    {
      code: `function foo() {
        return hey.then(function(x) {});
      }`,
      options: [{ ignoreLastCallback: true }],
      errors: [{ message }],
    },
    {
      code: `async function foo() {
        return await hey.then(x => { console.log(x) })
      }`,
      options: [{ ignoreLastCallback: true }],
      errors: [{ message }],
    },
    {
      code: 'const foo = hey?.then(x => { console.log(x) })',
      options: [{ ignoreLastCallback: true }],
      errors: [{ message }],
    },
    {
      code: 'const foo = (42, hey.then(x => { console.log(x) }))',
      options: [{ ignoreLastCallback: true }],
      errors: [{ message }],
    },
    { code: 'hey.then(x => { invalid = x })', errors: [{ message }] },
    { code: "hey.then(x => { invalid['x'] = x })", errors: [{ message }] },
    { code: 'hey.then(x => { const value = x })', errors: [{ message }] },
    {
      code: 'hey.then(x => { notWindow[x] = x })',
      options: [{ ignoreAssignmentVariable: ['window'] }],
      errors: [{ message }],
    },
    {
      code: "hey.then(x => { notWindow['x'] = x })",
      options: [{ ignoreAssignmentVariable: ['window'] }],
      errors: [{ message }],
    },
    {
      code: "hey.then(x => { windows['x'] = x })",
      options: [{ ignoreAssignmentVariable: ['window'] }],
      errors: [{ message }],
    },
    {
      code: 'hey.then(x => { x() })',
      options: [{ ignoreAssignmentVariable: ['window'] }],
      errors: [{ message }],
    },
    {
      code: 'hey.then(x => { process.exitCode = 1 })',
      errors: [{ message }],
    },

    // switch where a case conditionally breaks out before its return (review fix)
    {
      code: 'hey.then(function(x) { switch (x) { case 1: if (y) break; return "a"; default: return "b"; } })',
      errors: [{ message }],
    },
    {
      code: 'hey.then(function(x) { switch (x) { case 1: { if (y) break; } return "a"; default: return "b"; } })',
      errors: [{ message }],
    },
    // compound assignment to a non-ignored variable is still reported
    { code: 'hey.then(x => { notGlobal.a += x })', errors: [{ message }] },

    // try/catch: catch reachable because the try block may throw
    {
      code: 'hey.then(function() { try { return foo() } catch (e) { log(e) } })',
      errors: [{ message }],
    },
    {
      code: 'hey.then(function() { try { foo(); return 1 } catch (e) { log(e) } })',
      errors: [{ message }],
    },
  ],
});
