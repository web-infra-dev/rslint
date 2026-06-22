/**
 * Conformance: eslint-plugin-es-x (nonstandard) mounted in rslint via `plugins` must
 * report identically to ESLint v10. es-x rules are AST pattern matches over ES
 * version features / builtin APIs (no type info), so rslint reproduces ESLint
 * byte-for-byte. Representative triggers from the upstream test suite (v9.6.0).
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-array-properties',
    code: 'Array.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-array-properties',
    code: 'Array.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-array-properties',
    code: 'Array.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-array-properties',
    code: 'const { foo } = Array;',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-array-prototype-properties',
    code: "['A'].unknown()",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-array-prototype-properties',
    code: "['A'].foo",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-array-prototype-properties',
    code: "['A'].bar",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-array-prototype-properties',
    code: "['A']['01']",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-arraybuffer-properties',
    code: 'ArrayBuffer.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-arraybuffer-properties',
    code: 'ArrayBuffer.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-arraybuffer-properties',
    code: 'ArrayBuffer.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-arraybuffer-prototype-properties',
    code: 'new ArrayBuffer().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-arraybuffer-prototype-properties',
    code: 'new ArrayBuffer().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-arraybuffer-prototype-properties',
    code: 'new ArrayBuffer().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-arraybuffer-prototype-properties',
    code: 'new ArrayBuffer()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-asyncdisposablestack-properties',
    code: 'AsyncDisposableStack.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-asyncdisposablestack-properties',
    code: 'AsyncDisposableStack.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-asyncdisposablestack-prototype-properties',
    code: '(new AsyncDisposableStack()).unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-asyncdisposablestack-prototype-properties',
    code: '(new AsyncDisposableStack()).foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-asyncdisposablestack-prototype-properties',
    code: '(new AsyncDisposableStack())[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-asyncdisposablestack-prototype-properties',
    code: "(new AsyncDisposableStack())['01']",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-atomics-properties',
    code: 'Atomics.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-atomics-properties',
    code: 'Atomics.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-atomics-properties',
    code: 'Atomics.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-bigint-properties',
    code: 'BigInt.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-bigint-properties',
    code: 'BigInt.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-bigint-properties',
    code: 'BigInt.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-bigint-prototype-properties',
    code: '(123n).unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-bigint-prototype-properties',
    code: '(123n).foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-bigint-prototype-properties',
    code: '(123n).bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-bigint-prototype-properties',
    code: '(123n)[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-boolean-properties',
    code: 'Boolean.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-boolean-properties',
    code: 'Boolean.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-boolean-properties',
    code: 'Boolean.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-boolean-prototype-properties',
    code: 'true.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-boolean-prototype-properties',
    code: 'true.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-boolean-prototype-properties',
    code: 'true.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-boolean-prototype-properties',
    code: 'true[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-dataview-properties',
    code: 'DataView.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-dataview-properties',
    code: 'DataView.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-dataview-properties',
    code: 'DataView.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-dataview-prototype-properties',
    code: 'new DataView().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-dataview-prototype-properties',
    code: 'new DataView().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-dataview-prototype-properties',
    code: 'new DataView().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-dataview-prototype-properties',
    code: 'new DataView()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-date-properties',
    code: 'Date.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-date-properties',
    code: 'Date.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-date-properties',
    code: 'Date.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-date-prototype-properties',
    code: 'new Date().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-date-prototype-properties',
    code: 'new Date().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-date-prototype-properties',
    code: 'new Date().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-date-prototype-properties',
    code: 'new Date()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-disposablestack-properties',
    code: 'DisposableStack.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-disposablestack-properties',
    code: 'DisposableStack.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-disposablestack-prototype-properties',
    code: '(new DisposableStack()).unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-disposablestack-prototype-properties',
    code: '(new DisposableStack()).foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-disposablestack-prototype-properties',
    code: '(new DisposableStack())[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-disposablestack-prototype-properties',
    code: "(new DisposableStack())['01']",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-error-properties',
    code: 'Error.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-error-properties',
    code: 'Error.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-error-properties',
    code: 'Error.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-finalizationregistry-properties',
    code: 'FinalizationRegistry.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-finalizationregistry-properties',
    code: 'FinalizationRegistry.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-finalizationregistry-properties',
    code: 'FinalizationRegistry.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-finalizationregistry-prototype-properties',
    code: 'new FinalizationRegistry().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-finalizationregistry-prototype-properties',
    code: 'new FinalizationRegistry().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-finalizationregistry-prototype-properties',
    code: 'new FinalizationRegistry().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-finalizationregistry-prototype-properties',
    code: 'new FinalizationRegistry()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-function-properties',
    code: 'Function.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-function-properties',
    code: 'Function.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-function-properties',
    code: 'Function.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-collator-properties',
    code: 'Intl.Collator.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-collator-properties',
    code: 'Intl.Collator.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-collator-properties',
    code: 'Intl.Collator.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-collator-prototype-properties',
    code: 'new Intl.Collator().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-collator-prototype-properties',
    code: 'new Intl.Collator().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-collator-prototype-properties',
    code: 'new Intl.Collator().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-collator-prototype-properties',
    code: 'new Intl.Collator()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-datetimeformat-properties',
    code: 'Intl.DateTimeFormat.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-datetimeformat-properties',
    code: 'Intl.DateTimeFormat.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-datetimeformat-properties',
    code: 'Intl.DateTimeFormat.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-datetimeformat-properties',
    code: '\n            if (Intl.DateTimeFormat.unknown) {\n                console.log(Intl.DateTimeFormat.unknown())\n            }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-datetimeformat-prototype-properties',
    code: 'new Intl.DateTimeFormat().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-datetimeformat-prototype-properties',
    code: 'new Intl.DateTimeFormat().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-datetimeformat-prototype-properties',
    code: 'new Intl.DateTimeFormat().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-datetimeformat-prototype-properties',
    code: 'new Intl.DateTimeFormat()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-displaynames-properties',
    code: 'Intl.DisplayNames.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-displaynames-properties',
    code: 'Intl.DisplayNames.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-displaynames-properties',
    code: 'Intl.DisplayNames.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-displaynames-prototype-properties',
    code: 'new Intl.DisplayNames().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-displaynames-prototype-properties',
    code: 'new Intl.DisplayNames().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-displaynames-prototype-properties',
    code: 'new Intl.DisplayNames().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-displaynames-prototype-properties',
    code: 'new Intl.DisplayNames()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-durationformat-properties',
    code: 'Intl.DurationFormat.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-durationformat-properties',
    code: 'Intl.DurationFormat.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-durationformat-properties',
    code: 'Intl.DurationFormat.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-durationformat-prototype-properties',
    code: 'new Intl.DurationFormat().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-durationformat-prototype-properties',
    code: 'new Intl.DurationFormat().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-durationformat-prototype-properties',
    code: 'new Intl.DurationFormat().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-durationformat-prototype-properties',
    code: 'new Intl.DurationFormat()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-listformat-properties',
    code: 'Intl.ListFormat.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-listformat-properties',
    code: 'Intl.ListFormat.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-listformat-properties',
    code: 'Intl.ListFormat.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-listformat-prototype-properties',
    code: 'new Intl.ListFormat().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-listformat-prototype-properties',
    code: 'new Intl.ListFormat().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-listformat-prototype-properties',
    code: 'new Intl.ListFormat().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-listformat-prototype-properties',
    code: 'new Intl.ListFormat()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-locale-properties',
    code: 'Intl.Locale.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-locale-properties',
    code: 'Intl.Locale.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-locale-properties',
    code: 'Intl.Locale.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-locale-prototype-properties',
    code: 'new Intl.Locale().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-locale-prototype-properties',
    code: 'new Intl.Locale().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-locale-prototype-properties',
    code: 'new Intl.Locale().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-locale-prototype-properties',
    code: 'new Intl.Locale()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-numberformat-properties',
    code: 'Intl.NumberFormat.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-numberformat-properties',
    code: 'Intl.NumberFormat.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-numberformat-properties',
    code: 'Intl.NumberFormat.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-numberformat-prototype-properties',
    code: 'new Intl.NumberFormat().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-numberformat-prototype-properties',
    code: 'new Intl.NumberFormat().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-numberformat-prototype-properties',
    code: 'new Intl.NumberFormat().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-numberformat-prototype-properties',
    code: 'new Intl.NumberFormat()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-pluralrules-properties',
    code: 'Intl.PluralRules.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-pluralrules-properties',
    code: 'Intl.PluralRules.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-pluralrules-properties',
    code: 'Intl.PluralRules.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-pluralrules-prototype-properties',
    code: 'new Intl.PluralRules().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-pluralrules-prototype-properties',
    code: 'new Intl.PluralRules().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-pluralrules-prototype-properties',
    code: 'new Intl.PluralRules().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-pluralrules-prototype-properties',
    code: 'new Intl.PluralRules()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-properties',
    code: 'Intl.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-properties',
    code: 'Intl.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-properties',
    code: 'Intl.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-properties',
    code: '\n            if (Intl.Unknown) {\n                console.log(new Intl.Unknown())\n            }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-relativetimeformat-properties',
    code: 'Intl.RelativeTimeFormat.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-relativetimeformat-properties',
    code: 'Intl.RelativeTimeFormat.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-relativetimeformat-properties',
    code: 'Intl.RelativeTimeFormat.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-relativetimeformat-prototype-properties',
    code: 'new Intl.RelativeTimeFormat().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-relativetimeformat-prototype-properties',
    code: 'new Intl.RelativeTimeFormat().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-relativetimeformat-prototype-properties',
    code: 'new Intl.RelativeTimeFormat().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-relativetimeformat-prototype-properties',
    code: 'new Intl.RelativeTimeFormat()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-segmenter-properties',
    code: 'Intl.Segmenter.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-segmenter-properties',
    code: 'Intl.Segmenter.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-segmenter-properties',
    code: 'Intl.Segmenter.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-segmenter-prototype-properties',
    code: 'new Intl.Segmenter().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-segmenter-prototype-properties',
    code: 'new Intl.Segmenter().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-segmenter-prototype-properties',
    code: 'new Intl.Segmenter().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-segmenter-prototype-properties',
    code: 'new Intl.Segmenter()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-iterator-properties',
    code: 'Iterator.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-iterator-properties',
    code: 'Iterator.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-iterator-properties',
    code: 'Iterator.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-iterator-prototype-properties',
    code: 'Iterator.from({}).unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-iterator-prototype-properties',
    code: 'Iterator.from({}).foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-iterator-prototype-properties',
    code: 'Iterator.from({}).bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-iterator-prototype-properties',
    code: 'Iterator.from({})[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-json-properties',
    code: 'JSON.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-json-properties',
    code: 'JSON.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-json-properties',
    code: 'JSON.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-map-properties',
    code: 'Map.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-map-properties',
    code: 'Map.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-map-properties',
    code: 'Map.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-map-prototype-properties',
    code: 'new Map().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-map-prototype-properties',
    code: 'new Map().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-map-prototype-properties',
    code: 'new Map().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-map-prototype-properties',
    code: 'new Map()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-math-properties',
    code: 'Math.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-math-properties',
    code: 'Math.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-math-properties',
    code: 'Math.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-number-properties',
    code: 'Number.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-number-properties',
    code: 'Number.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-number-properties',
    code: 'Number.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-number-prototype-properties',
    code: '(123).unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-number-prototype-properties',
    code: '(123).foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-number-prototype-properties',
    code: '(123).bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-number-prototype-properties',
    code: '(123)[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-object-properties',
    code: 'Object.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-object-properties',
    code: 'Object.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-object-properties',
    code: 'Object.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-promise-properties',
    code: 'Promise.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-promise-properties',
    code: 'Promise.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-promise-properties',
    code: 'Promise.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-promise-prototype-properties',
    code: 'Promise.resolve().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-promise-prototype-properties',
    code: 'Promise.resolve().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-promise-prototype-properties',
    code: 'Promise.resolve().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-promise-prototype-properties',
    code: 'Promise.resolve()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-proxy-properties',
    code: 'Proxy.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-proxy-properties',
    code: 'Proxy.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-proxy-properties',
    code: 'Proxy.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-reflect-properties',
    code: 'Reflect.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-reflect-properties',
    code: 'Reflect.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-reflect-properties',
    code: 'Reflect.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-regexp-properties',
    code: 'RegExp.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-regexp-properties',
    code: 'RegExp.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-regexp-properties',
    code: 'RegExp.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-regexp-prototype-properties',
    code: '/foo/u.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-regexp-prototype-properties',
    code: '/foo/u.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-regexp-prototype-properties',
    code: '/foo/u.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-regexp-prototype-properties',
    code: '/foo/u[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-set-properties',
    code: 'Set.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-set-properties',
    code: 'Set.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-set-properties',
    code: 'Set.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-set-prototype-properties',
    code: 'new Set().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-set-prototype-properties',
    code: 'new Set().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-set-prototype-properties',
    code: 'new Set().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-set-prototype-properties',
    code: 'new Set()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-sharedarraybuffer-properties',
    code: 'SharedArrayBuffer.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-sharedarraybuffer-properties',
    code: 'SharedArrayBuffer.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-sharedarraybuffer-properties',
    code: 'SharedArrayBuffer.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-sharedarraybuffer-prototype-properties',
    code: 'new SharedArrayBuffer().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-sharedarraybuffer-prototype-properties',
    code: 'new SharedArrayBuffer().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-sharedarraybuffer-prototype-properties',
    code: 'new SharedArrayBuffer().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-sharedarraybuffer-prototype-properties',
    code: 'new SharedArrayBuffer()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-string-properties',
    code: 'String.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-string-properties',
    code: 'String.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-string-properties',
    code: 'String.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-string-properties',
    code: '\n            if (String.unknown) {\n                console.log(String.unknown())\n            }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-string-properties',
    code: '\n            if (String.unknown) {\n                console.log(String.unknown``)\n            }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-string-prototype-properties',
    code: "'A'.unknown()",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-string-prototype-properties',
    code: "'123'.foo",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-string-prototype-properties',
    code: "'123'.bar",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-string-prototype-properties',
    code: "'123'['01']",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-symbol-properties',
    code: 'Symbol.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-symbol-properties',
    code: 'Symbol.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-symbol-properties',
    code: 'Symbol.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-symbol-prototype-properties',
    code: 'Symbol().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-symbol-prototype-properties',
    code: 'Symbol().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-symbol-prototype-properties',
    code: 'Symbol().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-symbol-prototype-properties',
    code: 'Symbol()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-duration-properties',
    code: 'Temporal.Duration.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-duration-properties',
    code: 'Temporal.Duration.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-duration-prototype-properties',
    code: '(new Temporal.Duration()).unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-duration-prototype-properties',
    code: '(new Temporal.Duration()).foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-duration-prototype-properties',
    code: '(new Temporal.Duration())[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-duration-prototype-properties',
    code: "(new Temporal.Duration())['01']",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-instant-properties',
    code: 'Temporal.Instant.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-instant-properties',
    code: 'Temporal.Instant.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-instant-prototype-properties',
    code: '(new Temporal.Instant()).unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-instant-prototype-properties',
    code: '(new Temporal.Instant()).foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-instant-prototype-properties',
    code: '(new Temporal.Instant())[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-instant-prototype-properties',
    code: "(new Temporal.Instant())['01']",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-now-properties',
    code: 'Temporal.Now.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-now-properties',
    code: 'Temporal.Now.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindate-properties',
    code: 'Temporal.PlainDate.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindate-properties',
    code: 'Temporal.PlainDate.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindate-prototype-properties',
    code: '(new Temporal.PlainDate()).unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindate-prototype-properties',
    code: '(new Temporal.PlainDate()).foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindate-prototype-properties',
    code: '(new Temporal.PlainDate())[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindate-prototype-properties',
    code: "(new Temporal.PlainDate())['01']",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindatetime-properties',
    code: 'Temporal.PlainDateTime.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindatetime-properties',
    code: 'Temporal.PlainDateTime.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindatetime-prototype-properties',
    code: '(new Temporal.PlainDateTime()).unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindatetime-prototype-properties',
    code: '(new Temporal.PlainDateTime()).foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindatetime-prototype-properties',
    code: '(new Temporal.PlainDateTime())[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindatetime-prototype-properties',
    code: "(new Temporal.PlainDateTime())['01']",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainmonthday-properties',
    code: 'Temporal.PlainMonthDay.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainmonthday-properties',
    code: 'Temporal.PlainMonthDay.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainmonthday-prototype-properties',
    code: '(new Temporal.PlainMonthDay()).unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainmonthday-prototype-properties',
    code: '(new Temporal.PlainMonthDay()).foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainmonthday-prototype-properties',
    code: '(new Temporal.PlainMonthDay())[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainmonthday-prototype-properties',
    code: "(new Temporal.PlainMonthDay())['01']",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaintime-properties',
    code: 'Temporal.PlainTime.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaintime-properties',
    code: 'Temporal.PlainTime.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaintime-prototype-properties',
    code: '(new Temporal.PlainTime()).unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaintime-prototype-properties',
    code: '(new Temporal.PlainTime()).foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaintime-prototype-properties',
    code: '(new Temporal.PlainTime())[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaintime-prototype-properties',
    code: "(new Temporal.PlainTime())['01']",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainyearmonth-properties',
    code: 'Temporal.PlainYearMonth.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainyearmonth-properties',
    code: 'Temporal.PlainYearMonth.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainyearmonth-prototype-properties',
    code: '(new Temporal.PlainYearMonth()).unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainyearmonth-prototype-properties',
    code: '(new Temporal.PlainYearMonth()).foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainyearmonth-prototype-properties',
    code: '(new Temporal.PlainYearMonth())[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainyearmonth-prototype-properties',
    code: "(new Temporal.PlainYearMonth())['01']",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-properties',
    code: 'Temporal.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-properties',
    code: 'Temporal.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-zoneddatetime-properties',
    code: 'Temporal.ZonedDateTime.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-zoneddatetime-properties',
    code: 'Temporal.ZonedDateTime.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-zoneddatetime-prototype-properties',
    code: '(new Temporal.ZonedDateTime()).unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-zoneddatetime-prototype-properties',
    code: '(new Temporal.ZonedDateTime()).foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-zoneddatetime-prototype-properties',
    code: '(new Temporal.ZonedDateTime())[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-zoneddatetime-prototype-properties',
    code: "(new Temporal.ZonedDateTime())['01']",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakmap-properties',
    code: 'WeakMap.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakmap-properties',
    code: 'WeakMap.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakmap-properties',
    code: 'WeakMap.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakmap-prototype-properties',
    code: 'new WeakMap().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakmap-prototype-properties',
    code: 'new WeakMap().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakmap-prototype-properties',
    code: 'new WeakMap().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakmap-prototype-properties',
    code: 'new WeakMap()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakref-properties',
    code: 'WeakRef.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakref-properties',
    code: 'WeakRef.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakref-properties',
    code: 'WeakRef.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakref-prototype-properties',
    code: 'new WeakRef().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakref-prototype-properties',
    code: 'new WeakRef().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakref-prototype-properties',
    code: 'new WeakRef().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakref-prototype-properties',
    code: 'new WeakRef()[0]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakset-properties',
    code: 'WeakSet.unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakset-properties',
    code: 'WeakSet.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakset-properties',
    code: 'WeakSet.bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakset-prototype-properties',
    code: 'new WeakSet().unknown()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakset-prototype-properties',
    code: 'new WeakSet().foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakset-prototype-properties',
    code: 'new WeakSet().bar',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakset-prototype-properties',
    code: 'new WeakSet()[0]',
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-array-properties',
    code: 'Array[Symbol.species]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-array-properties',
    code: 'const {[Symbol.species]:foo} = Array',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-array-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-array-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-arraybuffer-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-arraybuffer-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-asyncdisposablestack-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-asyncdisposablestack-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-bigint-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-bigint-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-boolean-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-boolean-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-dataview-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-dataview-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-date-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-date-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-disposablestack-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-disposablestack-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-finalizationregistry-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-finalizationregistry-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-collator-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-collator-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-datetimeformat-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-datetimeformat-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-displaynames-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-displaynames-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-durationformat-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-durationformat-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-listformat-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-listformat-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-locale-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-locale-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-numberformat-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-numberformat-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-pluralrules-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-pluralrules-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-relativetimeformat-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-relativetimeformat-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-segmenter-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-intl-segmenter-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-iterator-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-iterator-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-map-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-map-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-number-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-number-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-promise-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-promise-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-regexp-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-regexp-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-set-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-set-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-sharedarraybuffer-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-sharedarraybuffer-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-string-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-string-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-symbol-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-symbol-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-duration-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-duration-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-instant-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-instant-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindate-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindate-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindatetime-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaindatetime-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainmonthday-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainmonthday-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaintime-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plaintime-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainyearmonth-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-plainyearmonth-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-zoneddatetime-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-temporal-zoneddatetime-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-typed-array-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-typed-array-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakmap-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakmap-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakref-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakref-prototype-properties',
    code: 'foo.toString',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakset-prototype-properties',
    code: 'foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nonstandard-weakset-prototype-properties',
    code: 'foo.toString',
  },
];

runConformanceSuite('eslint-plugin-es-x', CASES, CLEAN_CASES);
