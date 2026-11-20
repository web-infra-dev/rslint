// Upstream: eslint-plugin-jest@29.16.0; the native range covers the callee.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

type Version = number | string;
const deprecations = [
  [15, 'jest.resetModuleRegistry', 'jest.resetModules'],
  [17, 'jest.addMatchers', 'expect.extend'],
  [21, 'require.requireMock', 'jest.requireMock'],
  [21, 'require.requireActual', 'jest.requireActual'],
  [22, 'jest.runTimersToTime', 'jest.advanceTimersByTime'],
  [26, 'jest.genMockFromModule', 'jest.createMockFromModule'],
] as const;

function accessors(member: string): string[] {
  const [object, property] = member.split('.');
  return [member, `${object}['${property}']`];
}

function deprecatedCases(
  version: Version,
  deprecated: string,
  replacement: string,
): TestCase[] {
  return accessors(deprecated).map((callee, index) => ({
    code: `${callee}()`,
    settings: { jest: { version } },
    errors: [
      {
        messageId: 'deprecatedFunction',
        message: `\`${deprecated}\` has been deprecated in favor of \`${replacement}\``,
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: callee.length + 1,
      },
    ],
    output: `${accessors(replacement)[index]}()`,
  }));
}

function allowedCases(version: Version, member: string): ValidTestCase[] {
  return accessors(member).flatMap((callee) => [
    { code: `${callee}()`, settings: { jest: { version } } },
    { code: callee, settings: { jest: { version } } },
  ]);
}

const originalValid = [
  ['jest', 14],
  ['require("fs")', 14],
  ['jest.resetModuleRegistry', 14],
  ['require.requireActual', 17],
  ['jest.genMockFromModule', 25],
  ['jest.genMockFromModule', '25.1.1'],
  ['require.requireActual', '17.2'],
] as const;
const originalInvalid = [
  [21, 'jest.resetModuleRegistry', 'jest.resetModules'],
  [24, 'jest.addMatchers', 'expect.extend'],
  [21, 'require.requireMock', 'jest.requireMock'],
  [21, 'require.requireActual', 'jest.requireActual'],
  [22, 'jest.runTimersToTime', 'jest.advanceTimersByTime'],
  [26, 'jest.genMockFromModule', 'jest.createMockFromModule'],
  ['26.0.0-next.11', 'jest.genMockFromModule', 'jest.createMockFromModule'],
] as const;
const valid: ValidTestCase[] = originalValid.map(([code, version], index) => ({
  name: `run 0 valid ${index}`,
  code,
  settings: { jest: { version } },
}));
const invalid: TestCase[] = originalInvalid
  .flatMap(([version, deprecated, replacement]) =>
    deprecatedCases(version, deprecated, replacement),
  )
  .map((item, index) => ({ name: `run 0 invalid ${index}`, ...item }));

// Keep the first occurrence of each exact input/configuration, matching the
// migration baseline. Names retain the index before duplicate removal.
function key(item: ValidTestCase | TestCase): string {
  return JSON.stringify([item.code, item.settings ?? {}]);
}
const seen = new Set([...valid, ...invalid].map(key));
function append<T extends ValidTestCase | TestCase>(
  target: T[],
  items: T[],
  prefix: string,
): void {
  items.forEach((item, index) => {
    const identity = key(item);
    if (seen.has(identity)) return;
    seen.add(identity);
    target.push({ name: `${prefix}${index}`, ...item });
  });
}

append(
  valid,
  [
    { code: 'jest', settings: { jest: { version: 14 } } },
    { code: 'require("fs")', settings: { jest: { version: 14 } } },
    ...allowedCases(14, 'jest.resetModuleRegistry'),
    ...allowedCases(17, 'require.requireActual'),
    ...allowedCases(25, 'jest.genMockFromModule'),
    ...allowedCases('25.1.1', 'jest.genMockFromModule'),
    ...allowedCases('17.2', 'require.requireActual'),
  ],
  'upstream no-deprecated-functions / valid ',
);

for (const version of [
  14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27,
]) {
  append(
    valid,
    [
      { code: 'jest' },
      { code: 'require("fs")' },
      ...deprecations
        .filter(([since]) => since > version)
        .flatMap(([, deprecated]) => allowedCases(version, deprecated)),
    ],
    'upstream explict jest version / valid ',
  );
  append(
    invalid,
    deprecations
      .filter(([since]) => since <= version)
      .flatMap(([, deprecated, replacement]) =>
        deprecatedCases(version, deprecated, replacement),
      ),
    'upstream explict jest version / invalid ',
  );
}

new RuleTester(fixtureConfig()).run('no-deprecated-functions', {
  valid,
  invalid,
});
