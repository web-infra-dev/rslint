import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-misleading-character-class', {
  valid: [
    'var r = /[👍]/u',
    String.raw`var r = /[\uD83D\uDC4D]/u`,
    String.raw`var r = /[\u{1F44D}]/u`,
    'var r = /❇️/',
    'var r = /Á/',
    'var r = /[❇]/',
    'var r = /👶🏻/',
    'var r = /[👶]/u',
    'var r = /🇯🇵/',
    'var r = /[JP]/',
    'var r = /👨‍👩‍👦/',
    'new RegExp()',
    'var r = RegExp(/[👍]/u)',
    'const regex = /[👍]/u; new RegExp(regex);',

    // Solo code points are OK
    String.raw`var r = /[\uD83D]/`,
    String.raw`var r = /[\uDC4D]/`,
    String.raw`var r = /[\u0301]/`,
    String.raw`var r = /[\uFE0F]/`,
    String.raw`var r = /[\u{1F3FB}]/u`,
    'var r = /[🇯]/u',
    'var r = /[🇵]/u',
    String.raw`var r = /[\u200D]/`,

    // Invalid regex patterns — skipped
    `new RegExp('[Á] [ ');`,
    `var r = new RegExp('[Á] [ ');`,
    `var r = RegExp('{ [Á]', 'u');`,

    // Template with substitution
    'var r = RegExp(`${x}[👍]`)',

    // Non-literal flags
    `var r = new RegExp('[🇯🇵]', \`\${foo}\`)`,
    `var r = new RegExp("[👍]", flags)`,

    // Spread
    `const args = ['[👍]', 'i']; new RegExp(...args);`,

    // v flag
    'var r = /[👍]/v',
    String.raw`var r = /^[\q{👶🏻}]$/v`,
    String.raw`var r = /[🇯\q{abc}🇵]/v`,
    'var r = /[🇯[A]🇵]/v',
    'var r = /[🇯[A--B]🇵]/v',

    // allowEscape
    {
      code: String.raw`/[\ud83d\udc4d]/`,
      options: { allowEscape: true },
    },
    {
      code: String.raw`/[A\u0301]/`,
      options: { allowEscape: true },
    },
    {
      code: String.raw`/[👶\u{1f3fb}]/u`,
      options: { allowEscape: true },
    },
    {
      code: String.raw`/[\u{1F1EF}\u{1F1F5}]/u`,
      options: { allowEscape: true },
    },
    {
      code: String.raw`/[\u00B7\u0300-\u036F]/u`,
      options: { allowEscape: true },
    },
    {
      code: String.raw`/[\n\u0305]/`,
      options: { allowEscape: true },
    },
    {
      code: String.raw`RegExp("[\uD83D\uDC4D]")`,
      options: { allowEscape: true },
    },
    {
      code: String.raw`RegExp("[A\u0301]")`,
      options: { allowEscape: true },
    },
  ],
  invalid: [
    // Regex literals
    {
      code: 'var r = /[👍]/',
      errors: [{ messageId: 'surrogatePairWithoutUFlag', line: 1, column: 11 }],
    },
    {
      code: String.raw`var r = /[\uD83D\uDC4D]/`,
      errors: [{ messageId: 'surrogatePairWithoutUFlag', line: 1, column: 11 }],
    },
    {
      code: String.raw`var r = /before[\uD83D\uDC4D]after/`,
      errors: [{ messageId: 'surrogatePairWithoutUFlag', line: 1, column: 17 }],
    },
    {
      code: String.raw`var r = /[before\uD83D\uDC4Dafter]/`,
      errors: [{ messageId: 'surrogatePairWithoutUFlag', line: 1, column: 17 }],
    },

    // combiningClass
    {
      code: 'var r = /[A\u0301]/',
      errors: [{ messageId: 'combiningClass', line: 1, column: 11 }],
    },
    {
      code: 'var r = /[A\u0301]/u',
      errors: [{ messageId: 'combiningClass', line: 1, column: 11 }],
    },
    {
      code: String.raw`var r = /[\u0041\u0301]/`,
      errors: [{ messageId: 'combiningClass', line: 1, column: 11 }],
    },
    {
      code: String.raw`var r = /[\u0041\u0301]/u`,
      errors: [{ messageId: 'combiningClass', line: 1, column: 11 }],
    },
    {
      code: String.raw`var r = /[\u{41}\u{301}]/u`,
      errors: [{ messageId: 'combiningClass', line: 1, column: 11 }],
    },
    {
      code: 'var r = /[\u2747\uFE0F]/',
      errors: [{ messageId: 'combiningClass', line: 1, column: 11 }],
    },
    {
      code: 'var r = /[\u2747\uFE0F]/u',
      errors: [{ messageId: 'combiningClass', line: 1, column: 11 }],
    },

    // emojiModifier
    {
      code: 'var r = /[👶🏻]/u',
      errors: [{ messageId: 'emojiModifier', line: 1, column: 11 }],
    },
    {
      code: String.raw`var r = /[a\uD83C\uDFFB]/u`,
      errors: [{ messageId: 'emojiModifier', line: 1, column: 11 }],
    },
    {
      code: String.raw`var r = /[\u{1F476}\u{1F3FB}]/u`,
      errors: [{ messageId: 'emojiModifier', line: 1, column: 11 }],
    },

    // regionalIndicatorSymbol
    {
      code: 'var r = /[🇯🇵]/u',
      errors: [{ messageId: 'regionalIndicatorSymbol', line: 1, column: 11 }],
    },
    {
      code: String.raw`var r = /[\uD83C\uDDEF\uD83C\uDDF5]/u`,
      errors: [{ messageId: 'regionalIndicatorSymbol', line: 1, column: 11 }],
    },
    {
      code: String.raw`var r = /[\u{1F1EF}\u{1F1F5}]/u`,
      errors: [{ messageId: 'regionalIndicatorSymbol', line: 1, column: 11 }],
    },

    // zwj
    {
      code: 'var r = /[👨‍👩‍👦]/u',
      errors: [{ messageId: 'zwj', line: 1, column: 11 }],
    },
    {
      code: 'var r = /[👩‍👦]/u',
      errors: [{ messageId: 'zwj', line: 1, column: 11 }],
    },
    {
      code: String.raw`var r = /[\uD83D\uDC68\u200D\uD83D\uDC69\u200D\uD83D\uDC66]/u`,
      errors: [{ messageId: 'zwj', line: 1, column: 11 }],
    },
    {
      code: String.raw`var r = /[\u{1F468}\u{200D}\u{1F469}\u{200D}\u{1F466}]/u`,
      errors: [{ messageId: 'zwj', line: 1, column: 11 }],
    },

    // RegExp constructor
    {
      code: 'var r = RegExp("[👍]", "")',
      errors: [{ messageId: 'surrogatePairWithoutUFlag', line: 1, column: 18 }],
    },
    {
      code: 'var r = new RegExp("[👍]", "")',
      errors: [{ messageId: 'surrogatePairWithoutUFlag', line: 1, column: 22 }],
    },
    {
      code: 'var r = new RegExp("[A\u0301]", "")',
      errors: [{ messageId: 'combiningClass', line: 1, column: 22 }],
    },
    {
      code: 'var r = new RegExp("[👶🏻]", "u")',
      errors: [{ messageId: 'emojiModifier', line: 1, column: 22 }],
    },
    {
      code: 'var r = new RegExp("[🇯🇵]", "u")',
      errors: [{ messageId: 'regionalIndicatorSymbol', line: 1, column: 22 }],
    },
    {
      code: 'var r = new globalThis.RegExp("[\u2747\uFE0F]", "")',
      errors: [{ messageId: 'combiningClass', line: 1, column: 33 }],
    },

    // allowEscape: still flags non-escaped
    {
      code: 'var r = /[A\u0301]/',
      options: { allowEscape: true },
      errors: [{ messageId: 'combiningClass' }],
    },
  ],
});
