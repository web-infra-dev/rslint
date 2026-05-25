import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const generateInvalidCases = (
  jestVersion: number | string | undefined,
  deprecation: string,
  replacement: string,
) => {
  const [deprecatedName, deprecatedFunc] = deprecation.split('.');
  const [replacementName, replacementFunc] = replacement.split('.');
  const settings = { jest: { version: jestVersion } };
  const errors = [
    { messageId: 'deprecatedFunction', data: { deprecation, replacement } },
  ];

  return [
    {
      code: `${deprecation}()`,
      output: `${replacement}()`,
      settings,
      errors,
    },
    {
      code: `${deprecatedName}['${deprecatedFunc}']()`,
      output: `${replacementName}['${replacementFunc}']()`,
      settings,
      errors,
    },
  ];
};

ruleTester.run('no-deprecated-functions', {} as never, {
  valid: [
    { code: 'jest', settings: { jest: { version: 14 } } },
    { code: 'require("fs")', settings: { jest: { version: 14 } } },
    { code: 'jest.resetModuleRegistry', settings: { jest: { version: 14 } } },
    { code: 'require.requireActual', settings: { jest: { version: 17 } } },
    { code: 'jest.genMockFromModule', settings: { jest: { version: 25 } } },
    {
      code: 'jest.genMockFromModule',
      settings: { jest: { version: '25.1.1' } },
    },
    { code: 'require.requireActual', settings: { jest: { version: '17.2' } } },
  ],
  invalid: [
    ...generateInvalidCases(
      21,
      'jest.resetModuleRegistry',
      'jest.resetModules',
    ),
    ...generateInvalidCases(24, 'jest.addMatchers', 'expect.extend'),
    ...generateInvalidCases(21, 'require.requireMock', 'jest.requireMock'),
    ...generateInvalidCases(21, 'require.requireActual', 'jest.requireActual'),
    ...generateInvalidCases(
      22,
      'jest.runTimersToTime',
      'jest.advanceTimersByTime',
    ),
    ...generateInvalidCases(
      26,
      'jest.genMockFromModule',
      'jest.createMockFromModule',
    ),
    ...generateInvalidCases(
      '26.0.0-next.11',
      'jest.genMockFromModule',
      'jest.createMockFromModule',
    ),
  ],
});
