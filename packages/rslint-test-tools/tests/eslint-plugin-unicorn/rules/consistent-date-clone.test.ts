import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string, filename = 'file.js') => ({ code, filename });
const invalid = (
  code: string,
  output: string,
  errorCount = 1,
  filename = 'file.js',
) => ({
  code,
  filename,
  output,
  errors: Array.from({ length: errorCount }, () => ({
    messageId: 'consistent-date-clone/error',
    message: 'Unnecessary `.getTime()` call.',
  })),
});

ruleTester.run('consistent-date-clone', null as never, {
  valid: [
    valid('new Date(date)'),
    valid('date.getTime()'),
    valid('new Date(...date.getTime())'),
    valid('new Date(getTime())'),
    valid('new Date(date.getTime(), extraArgument)'),
    valid('new Date(date.not_getTime())'),
    valid('new Date(date?.getTime())'),
    valid('new NotDate(date.getTime())'),
    valid('new Date(date[getTime]())'),
    valid('new Date(date.getTime(extraArgument))'),
    valid('Date(date.getTime())'),
    valid(`new Date(
  date.getFullYear(),
  date.getMonth(),
  date.getDate(),
  date.getHours(),
  date.getMinutes(),
  date.getSeconds(),
  date.getMilliseconds(),
);`),
    valid(`new Date(
  date.getFullYear(),
  date.getMonth(),
  date.getDate(),
  date.getHours(),
  date.getMinutes(),
  date.getSeconds(),
);`),
  ],
  invalid: [
    invalid('new Date(date.getTime())', 'new Date(date)'),
    invalid('new Date(date.getTime(),)', 'new Date(date,)'),
    invalid(
      'new Date(new Date(date.getTime()).getTime())',
      'new Date(new Date(date))',
      2,
    ),
    invalid('new Date((0, date).getTime())', 'new Date((0, date))'),
    invalid('new Date(date.getTime(/* comment */))', 'new Date(date)'),
    invalid('new Date(date./* comment */getTime())', 'new Date(date)'),
    invalid(
      'new Date((date as Date).getTime())',
      'new Date((date as Date))',
      1,
      'file.ts',
    ),
  ],
});
