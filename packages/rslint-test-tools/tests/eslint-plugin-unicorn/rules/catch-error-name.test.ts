import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const error = (originalName: string, fixedName: string) => ({
  message: `The catch parameter \`${originalName}\` should be named \`${fixedName}\`.`,
});

const valid = (code: string) => ({ code });

ruleTester.run('catch-error-name', null as never, {
  valid: [
    valid('try {} catch (error) {}'),
    valid('try {} catch (descriptiveError) {}'),
    valid('try {} catch (descriptive_error__) {}'),
    valid('try {} catch (_) {}'),
    valid('try {} catch ({message}) {}'),
    valid('try {} catch {}'),
    valid('obj.catch(error => {})'),
    valid('obj.then(undefined, error => {})'),
    valid('obj.catch?.(bad => {})'),
    valid('obj.then?.(undefined, bad => {})'),
    valid('obj.catch(error => {}, extra)'),
    valid('foo(function (error) {})'),
    { code: 'try {} catch (err) {}', options: [{ name: 'err' }] },
    {
      code: 'try {} catch (skipThisNameCheck) {}',
      options: [{ ignore: ['^skip'] }],
    },
  ],
  invalid: [
    {
      code: 'try {} catch (err) { console.log(err) }',
      output: 'try {} catch (error) { console.log(error) }',
      errors: [error('err', 'error')],
    },
    {
      code: 'obj.catch(err => err)',
      output: 'obj.catch(error => error)',
      errors: [error('err', 'error')],
    },
    {
      code: 'obj.then(null, err => err)',
      output: 'obj.then(null, error => error)',
      errors: [error('err', 'error')],
    },
    {
      code: 'try {} catch (_) { console.log(_) }',
      output: 'try {} catch (error) { console.log(error) }',
      errors: [error('_', 'error')],
    },
    {
      code: 'try {} catch (notMatching) {}',
      output: 'try {} catch (error) {}',
      options: [{ ignore: ['unicorn'] }],
      errors: [error('notMatching', 'error')],
    },
    {
      code: 'try {} catch (e) {}',
      options: [{ name: 'has_space_after ' }],
      errors: [error('e', 'has_space_after ')],
    },
    {
      code: 'try {} catch (e) {}',
      options: [{ name: '1_start_with_a_number' }],
      errors: [error('e', '1_start_with_a_number')],
    },
    {
      code: 'try {} catch (e) {}',
      options: [{ name: '_){} evilCode; if(false' }],
      errors: [error('e', '_){} evilCode; if(false')],
    },
  ],
});
