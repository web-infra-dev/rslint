import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-useless-catch', {
  valid: [
    'try { foo(); } catch (err) { console.error(err); }',
    'try { foo(); } catch (err) { console.error(err); } finally { bar(); }',
    'try { foo(); } catch (err) { doSomethingBeforeRethrow(); throw err; }',
    'try { foo(); } catch (err) { throw err.msg; }',
    'try { foo(); } catch (err) { throw new Error("whoops!"); }',
    'try { foo(); } catch (err) { throw bar; }',
    'try { foo(); } catch (err) { }',
    'try { foo(); } catch ({ err }) { throw err; }',
    'try { foo(); } catch ([ err ]) { throw err; }',
    "try { throw new Error('foo'); } catch { throw new Error('foo'); }",
    'try { foo(); } catch (err) { throw err!; }',
    'try { foo(); } catch (err) { throw err as Error; }',
  ],
  invalid: [
    {
      code: 'try { foo(); } catch (err) { throw err; }',
      errors: [{ messageId: 'unnecessaryCatch' }],
    },
    {
      code: 'try { foo(); } catch (err) { throw err; } finally { foo(); }',
      errors: [{ messageId: 'unnecessaryCatchClause' }],
    },
    {
      code: 'try { foo(); } catch (err) { /* comment */ throw err; }',
      errors: [{ messageId: 'unnecessaryCatch' }],
    },
    {
      code: 'try { foo(); } catch (err: unknown) { throw err; }',
      errors: [{ messageId: 'unnecessaryCatch' }],
    },
    {
      code: 'try { foo(); } catch (err) { throw (err); }',
      errors: [{ messageId: 'unnecessaryCatch' }],
    },
    {
      code: 'try { foo(); } catch (err) { throw err; unreachable(); }',
      errors: [{ messageId: 'unnecessaryCatch' }],
    },
  ],
});
