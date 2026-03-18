import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-useless-catch', {
  valid: [
    'try { foo(); } catch (err) { console.error(err); }',
    'try { foo(); } catch (err) { throw bar; }',
    'try { foo(); } catch (err) { }',
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
  ],
});
