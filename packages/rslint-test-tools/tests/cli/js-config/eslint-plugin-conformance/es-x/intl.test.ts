/**
 * Conformance: eslint-plugin-es-x (intl) mounted in rslint via `plugins` must
 * report identically to ESLint v10. es-x rules are AST pattern matches over ES
 * version features / builtin APIs (no type info), so rslint reproduces ESLint
 * byte-for-byte. Representative triggers from the upstream test suite (v9.6.0).
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-displaynames',
    code: 'Intl.DisplayNames',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-durationformat',
    code: 'Intl.DurationFormat',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-getcanonicallocales',
    code: 'Intl.getCanonicalLocales',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-listformat',
    code: 'Intl.ListFormat',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-intl-locale', code: 'Intl.Locale' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-firstdayofweek',
    code: 'const foo = new Intl.Locale(); foo.firstDayOfWeek',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-getcalendars',
    code: 'const foo = new Intl.Locale(); foo.getCalendars()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-getcollations',
    code: 'const foo = new Intl.Locale(); foo.getCollations()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-gethourcycles',
    code: 'const foo = new Intl.Locale(); foo.getHourCycles()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-getnumberingsystems',
    code: 'const foo = new Intl.Locale(); foo.getNumberingSystems()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-gettextinfo',
    code: 'const foo = new Intl.Locale(); foo.getTextInfo()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-gettimezones',
    code: 'const foo = new Intl.Locale(); foo.getTimeZones()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-getweekinfo',
    code: 'const foo = new Intl.Locale(); foo.getWeekInfo()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-pluralrules',
    code: 'Intl.PluralRules',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-relativetimeformat',
    code: 'Intl.RelativeTimeFormat',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-segmenter',
    code: 'Intl.Segmenter',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-supportedvaluesof',
    code: 'Intl.supportedValuesOf',
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-datetimeformat-prototype-formatrange',
    code: 'foo.formatRange(startDate, endDate)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-datetimeformat-prototype-formatrange',
    code: 'formatRange(startDate, endDate)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-datetimeformat-prototype-formatrange',
    code: 'foo.unknown(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-datetimeformat-prototype-formattoparts',
    code: 'foo.formatToParts(now)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-datetimeformat-prototype-formattoparts',
    code: 'formatToParts(now)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-datetimeformat-prototype-formattoparts',
    code: 'foo.unknown(0)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-intl-displaynames', code: 'Intl' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-displaynames',
    code: 'Intl.DateTimeFormat',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-intl-durationformat', code: 'Intl' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-durationformat',
    code: 'Intl.DateTimeFormat',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-getcanonicallocales',
    code: 'Intl',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-getcanonicallocales',
    code: 'Intl.DateTimeFormat',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-intl-listformat', code: 'Intl' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-listformat',
    code: 'Intl.DateTimeFormat',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-intl-locale', code: 'Intl' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale',
    code: 'Intl.DateTimeFormat',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-firstdayofweek',
    code: 'foo.firstDayOfWeek',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-firstdayofweek',
    code: 'firstDayOfWeek',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-getcalendars',
    code: 'foo.getCalendars()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-getcalendars',
    code: 'getCalendars()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-getcollations',
    code: 'foo.getCollations()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-getcollations',
    code: 'getCollations()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-gethourcycles',
    code: 'foo.getHourCycles()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-gethourcycles',
    code: 'getHourCycles()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-getnumberingsystems',
    code: 'foo.getNumberingSystems()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-getnumberingsystems',
    code: 'getNumberingSystems()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-gettextinfo',
    code: 'foo.getTextInfo()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-gettextinfo',
    code: 'getTextInfo()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-gettimezones',
    code: 'foo.getTimeZones()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-gettimezones',
    code: 'getTimeZones()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-getweekinfo',
    code: 'foo.getWeekInfo()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-locale-prototype-getweekinfo',
    code: 'getWeekInfo()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-numberformat-prototype-formatrange',
    code: 'foo.formatRange(2.9, 3.1)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-numberformat-prototype-formatrange',
    code: 'formatRange(2.9, 3.1)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-numberformat-prototype-formatrange',
    code: 'foo.unknown(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-numberformat-prototype-formatrangetoparts',
    code: 'foo.formatRangeToParts(2.9, 3.1)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-numberformat-prototype-formatrangetoparts',
    code: 'formatRangeToParts(2.9, 3.1)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-numberformat-prototype-formatrangetoparts',
    code: 'foo.unknown(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-numberformat-prototype-formattoparts',
    code: 'foo.formatToParts(num)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-numberformat-prototype-formattoparts',
    code: 'formatToParts(num)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-numberformat-prototype-formattoparts',
    code: 'foo.unknown(0)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-intl-pluralrules', code: 'Intl' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-pluralrules',
    code: 'Intl.DateTimeFormat',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-pluralrules-prototype-selectrange',
    code: 'foo.selectRange(2.9, 3.1)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-pluralrules-prototype-selectrange',
    code: 'selectRange(2.9, 3.1)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-pluralrules-prototype-selectrange',
    code: 'foo.unknown(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-relativetimeformat',
    code: 'Intl',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-relativetimeformat',
    code: 'Intl.DateTimeFormat',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-intl-segmenter', code: 'Intl' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-segmenter',
    code: 'Intl.DateTimeFormat',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-supportedvaluesof',
    code: 'Intl',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-intl-supportedvaluesof',
    code: 'Intl.DateTimeFormat',
  },
];

runConformanceSuite('eslint-plugin-es-x', CASES, CLEAN_CASES);
