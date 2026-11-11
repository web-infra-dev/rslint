import { RuleTester, type TestCase } from '@rslint/rule-tester';

const ruleName = 'prefer-iterator-zip';
const loopHeader =
  'for (let index = 0; index < Math.min(names.length, scores.length); index++)';
const loopBody = ' { save(names[index], scores[index]); }';
const zipHeader = 'for (const [name_, score] of Iterator.zip([names, scores]))';
const zipBody = ' { save(name_, score); }';

const tester = new RuleTester({
  plugin: 'unicorn',
  fixtureFiles: { 'tsconfig.json': JSON.stringify({ include: ['*.ts'] }) },
  // The upstream snapshot runner configures the browser global name.
  languageOptions: { globals: { name: 'readonly' } },
});

function invalid(
  code: string,
  suggestion: string,
  [line, column, endLine, endColumn]: [number, number, number, number],
  source?: Pick<TestCase, 'filename' | 'languageOptions'>,
): TestCase {
  return {
    ...source,
    code,
    errors: [
      {
        messageId: ruleName,
        message: 'Use `Iterator.zip(…)` to iterate over these arrays together.',
        line,
        column,
        endLine,
        endColumn,
        suggestions: suggestion
          ? [
              {
                messageId: ruleName + '/suggestion',
                desc: 'Use `Iterator.zip(…)`.',
                output: suggestion,
              },
            ]
          : [],
      },
    ],
  };
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-iterator-zip.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-iterator-zip.md
// The preset and related-rule assertions need Unicorn v79 presets and rules missing from rslint.
tester.run(ruleName, {
  valid: [
    // Unsupported loop headers and cached bounds.
    'names.map((name, index) => scores[index])',
    'for (let index = 0; index < names.length; index++)' + loopBody,
    'for (let index = 0; index < length; index++)' + loopBody,
    'for (let index = 0; index < Math.max(names.length, scores.length); index++)' +
      loopBody,
    'for (let index = 0; index < Math.min(names.length, names.length); index++)' +
      loopBody,
    'for (let index = 0; index < Math.min(names.length, scores.length, 5); index++)' +
      loopBody,
    'for (let index = 0; index < Math.min(...lengths); index++)' + loopBody,
    'for (var index = 0; index < Math.min(names.length, scores.length); index++)' +
      loopBody,
    'for (const index = 0; index < Math.min(names.length, scores.length); index++)' +
      loopBody,
    'for (let index = 1; index < Math.min(names.length, scores.length); index++)' +
      loopBody,
    'for (let index = 0, length = 5; index < Math.min(names.length, scores.length); index++)' +
      loopBody,
    'for (let index = 0; index < Math.min(names.length, scores.length); index += 2)' +
      loopBody,
    'for (let index = 0; index < Math.min(names.length, scores.length); index--)' +
      loopBody,
    'for (let index = 0; index < Math.min(names.length, scores.length); index = index + 2)' +
      loopBody,
    'for (let index = 0; index < Math.min(names.length, scores.length); index = otherIndex + 1)' +
      loopBody,
    'for (let index = 0; index <= Math.min(names.length, scores.length); index++)' +
      loopBody,
    'for (let index = 0; index < names.length || index < scores.length; index++)' +
      loopBody,
    'for (let index = 0; index < names.length && enabled && index < scores.length; index++)' +
      loopBody,
    'for (let index = 0; index < names.length && index < names.length; index++)' +
      loopBody,
    'for (let index = 0; index < object.names.length && index < scores.length; index++)' +
      loopBody,
    'for (let index = 0; ; index++)' + loopBody,
    'const length = Math.min(names.length, scores.length);\nfor (let index = 0; index < length; index++)' +
      loopBody +
      ' use(length);',
    'const length = Math.min(names.length, scores.length); prepare(); for (let index = 0; index < length; index++)' +
      loopBody,
    'const length = Math.min(names.length, scores.length), other = 1;\nfor (let index = 0; index < length; index++)' +
      loopBody,
    'for (let index = 0, length = Math.min(names.length, scores.length); index < length; index++) { save(names[index], scores[index]); use(length); }',
    loopHeader + ' save(names[index], scores[index]);',
    // Uses other than element reads, including mixed shadowing.
    loopHeader + ' { save(names[index]); }',
    loopHeader + ' { save(names[index], scores[index], index); }',
    loopHeader + ' { save(names[index + 1], scores[index]); }',
    loopHeader + ' { save(names?.[index], scores[index]); }',
    loopHeader + ' { save(names[index], scores[index]); names.push("new"); }',
    loopHeader +
      ' { save(names[index], scores[index]); scores = replacement; }',
    loopHeader + ' { save(names[index], scores[index]); consume(scores); }',
    loopHeader + ' { save(names[index], scores[index]); index++; }',
    loopHeader + ' { names[index] = scores[index]; }',
    loopHeader + ' { names[index]++; save(scores[index]); }',
    loopHeader + ' { delete names[index]; save(scores[index]); }',
    loopHeader + ' { [names[index]] = scores[index]; }',
    loopHeader + ' { ({value: names[index]} = scores[index]); }',
    loopHeader + ' { for (names[index] of scores[index]) {} }',
    loopHeader + ' { for (names[index] in scores[index]) {} }',
    loopHeader + ' { names[index](scores[index]); }',
    loopHeader + ' { names[index]?.(scores[index]); }',
    loopHeader + ' { names[index]`tag`; save(scores[index]); }',
    loopHeader +
      ' { callbacks.push(() => save(names[index], scores[index])); }',
    loopHeader +
      ' { class Pair { value = save(names[index], scores[index]); } }',
    loopHeader +
      ' { save(names[index], scores[index]); { const names = other; save(names[index]); } }',
    loopHeader +
      ' { save(names[index], scores[index]); { const index = 0; save(names[index]); } }',
    // Receivers and TypeScript wrapper cases.
    'for (let index = 0; index < Math.min(object.names.length, scores.length); index++) { save(object.names[index], scores[index]); }',
    "const names = 'abc'; " + loopHeader + loopBody,
    "const names = {0: 'a', length: 1}; " + loopHeader + loopBody,
    'const scores = new Set(); ' + loopHeader + loopBody,
    {
      code:
        'declare const names: string; declare const scores: number[]; ' +
        loopHeader +
        loopBody,
      filename: 'file.ts',
    },
    {
      code: loopHeader + ' { names[index]!(); save(scores[index]); }',
      filename: 'file.ts',
    },
    {
      code: loopHeader + ' { (names[index]<string>)(scores[index]); }',
      filename: 'file.ts',
    },
    {
      code:
        loopHeader + ' { (names[index]<string>)`tag`; save(scores[index]); }',
      filename: 'file.ts',
    },
    {
      code: loopHeader + ' { (names[index] as number) = scores[index]; }',
      filename: 'file.ts',
    },
    {
      code:
        'declare const names: {length: number; [index: number]: string}; declare const scores: number[]; ' +
        loopHeader +
        loopBody,
      filename: 'file.ts',
      languageOptions: { parserOptions: { project: './tsconfig.json' } },
    },
    // Documentation examples and the upstream empty-file check.
    'for (const [name, score] of Iterator.zip([names, scores])) {\n\tsave(name, score);\n}',
    'index < Math.min(names.length, scores.length);\nindex < names.length && index < scores.length;',
    '',
  ],
  invalid: [
    // Math.min bounds, increments, and input types.
    invalid(
      'for (let index = 0; index < Math.min(Iterators.length, scores.length); index++) { save(Iterators[index], scores[index]); }',
      'for (const [Iterator_, score] of Iterator.zip([Iterators, scores])) { save(Iterator_, score); }',
      [1, 1, 1, 80],
      { languageOptions: { ecmaVersion: 2022 } },
    ),
    invalid(loopHeader + loopBody, zipHeader + zipBody, [1, 1, 1, 76]),
    invalid(
      'function pair(names, scores) { ' + loopHeader + loopBody + ' }',
      'function pair(names, scores) { ' + zipHeader + zipBody + ' }',
      [1, 32, 1, 107],
    ),
    invalid(
      "const names = ['a', 'b']; const scores = [1]; " + loopHeader + loopBody,
      "const names = ['a', 'b']; const scores = [1]; " + zipHeader + zipBody,
      [1, 47, 1, 122],
    ),
    invalid(
      'const names = []; const scores = [1]; ' + loopHeader + loopBody,
      'const names = []; const scores = [1]; ' + zipHeader + zipBody,
      [1, 39, 1, 114],
    ),
    invalid(
      'const names = new Uint8Array(2); const scores = new Float64Array(3); ' +
        loopHeader +
        loopBody,
      'const names = new Uint8Array(2); const scores = new Float64Array(3); ' +
        zipHeader +
        zipBody,
      [1, 70, 1, 145],
    ),
    invalid(
      'const names = new BigInt64Array(2); const scores = new BigUint64Array(3); ' +
        loopHeader +
        loopBody,
      'const names = new BigInt64Array(2); const scores = new BigUint64Array(3); ' +
        zipHeader +
        zipBody,
      [1, 75, 1, 150],
    ),
    invalid(
      'for (let index = 0; index < Math.min(names.length, scores.length); ++index)' +
        loopBody,
      zipHeader + zipBody,
      [1, 1, 1, 76],
    ),
    invalid(
      'for (let index = 0; index < Math.min(names.length, scores.length); index += 1)' +
        loopBody,
      zipHeader + zipBody,
      [1, 1, 1, 79],
    ),
    invalid(
      'for (let index = 0; Math.min(names.length, scores.length) > index; index++)' +
        loopBody,
      zipHeader + zipBody,
      [1, 1, 1, 76],
    ),
    invalid(
      'for (let index = 0; index < Math.min(names.length, scores.length, ranks.length); index++) { save(names[index], scores[index], ranks[index]); }',
      'for (const [name_, score, rank] of Iterator.zip([names, scores, ranks])) { save(name_, score, rank); }',
      [1, 1, 1, 90],
    ),
    // Element reads, shadowing, comments, and generated bindings.
    invalid(
      loopHeader + ' { save(names[index], scores[index], names[index]); }',
      zipHeader + ' { save(name_, score, name_); }',
      [1, 1, 1, 76],
    ),
    invalid(
      loopHeader +
        ' { const name = names[index]; const score = scores[index]; save(name, score); }',
      'for (const [name_, score_] of Iterator.zip([names, scores])) { const name = name_; const score = score_; save(name, score); }',
      [1, 1, 1, 76],
    ),
    invalid(
      loopHeader + ' { if (names[index])' + loopBody + ' }',
      zipHeader + ' { if (name_)' + zipBody + ' }',
      [1, 1, 1, 76],
    ),
    invalid(
      loopHeader +
        ' { if (!names[index]) { continue; } save(names[index], scores[index]); break; }',
      zipHeader + ' { if (!name_) { continue; } save(name_, score); break; }',
      [1, 1, 1, 76],
    ),
    invalid(
      loopHeader + ' { save((names)[(index)], ((scores[index]))); }',
      zipHeader + ' { save(name_, ((score))); }',
      [1, 1, 1, 76],
    ),
    invalid(
      loopHeader + ' { save(names[index].value, scores[index].value); }',
      zipHeader + ' { save(name_.value, score.value); }',
      [1, 1, 1, 76],
    ),
    invalid(
      loopHeader + ' { names[index].save(scores[index]); }',
      zipHeader + ' { name_.save(score); }',
      [1, 1, 1, 76],
    ),
    invalid(
      loopHeader + ' { names[index].value = scores[index]; }',
      zipHeader + ' { name_.value = score; }',
      [1, 1, 1, 76],
    ),
    invalid(
      loopHeader + ' { save(name, names[index], scores[index]); }',
      zipHeader + ' { save(name, name_, score); }',
      [1, 1, 1, 76],
    ),
    invalid(
      loopHeader +
        ' { save(names[index], scores[index]); { const name = "inner"; use(name); } }',
      zipHeader +
        ' { save(name_, score); { const name = "inner"; use(name); } }',
      [1, 1, 1, 76],
    ),
    invalid(
      loopHeader +
        ' { save(names[index], scores[index]); { const index = 1; use(index); } }',
      zipHeader + ' { save(name_, score); { const index = 1; use(index); } }',
      [1, 1, 1, 76],
    ),
    invalid(
      loopHeader +
        ' { /* keep body comment */ save(names[index], scores[index]); }',
      zipHeader + ' { /* keep body comment */ save(name_, score); }',
      [1, 1, 1, 76],
    ),
    invalid(
      loopHeader +
        ' { save(names[/* keep access comment */ index], scores[index]); }',
      '',
      [1, 1, 1, 76],
    ),
    invalid(
      "const name = 'existing'; " +
        loopHeader +
        ' { save(name, names[index], scores[index]); }',
      "const name = 'existing'; " +
        zipHeader +
        ' { save(name, name_, score); }',
      [1, 26, 1, 101],
    ),
    invalid(
      'for (let index = 0; index < Math.min(items.length, item.length); index++) { save(items[index], item[index]); }',
      'for (const [item_, element] of Iterator.zip([items, item])) { save(item_, element); }',
      [1, 1, 1, 74],
    ),
    invalid(
      'for (let index = 0; index < Math.min(data.length, metadata.length); index++) { save(data[index], metadata[index]); }',
      'for (const [datum, metadatum] of Iterator.zip([data, metadata])) { save(datum, metadatum); }',
      [1, 1, 1, 77],
    ),
    invalid(
      'for (let index = 0; /* keep header comment */ index < Math.min(names.length, scores.length); index++)' +
        loopBody,
      '',
      [1, 1, 1, 102],
    ),
    invalid(
      'for (/* before declaration */ let index = 0; index < Math.min(names.length, scores.length); index++)' +
        loopBody,
      '',
      [1, 1, 1, 101],
    ),
    invalid(
      'for (let index = 0; index < Math.min(names.length, scores.length); index++ /* after update */)' +
        loopBody,
      '',
      [1, 1, 1, 95],
    ),
    // TypeScript wrappers and JSX.
    invalid(
      'function pair(names: readonly string[], scores: number[]) { ' +
        loopHeader +
        loopBody +
        ' }',
      'function pair(names: readonly string[], scores: number[]) { ' +
        zipHeader +
        zipBody +
        ' }',
      [1, 61, 1, 136],
      { filename: 'file.ts' },
    ),
    invalid(
      loopHeader + ' { save(names[index]!, scores[index] as number); }',
      zipHeader + ' { save(name_!, score as number); }',
      [1, 1, 1, 76],
      { filename: 'file.ts' },
    ),
    invalid(
      loopHeader +
        ' { save(<string>names[index], scores[index] satisfies number); }',
      zipHeader + ' { save(<string>name_, score satisfies number); }',
      [1, 1, 1, 76],
      { filename: 'file.ts' },
    ),
    invalid(
      loopHeader + ' { save(names[index]<string>, scores[index]); }',
      zipHeader + ' { save(name_<string>, score); }',
      [1, 1, 1, 76],
      { filename: 'file.ts' },
    ),
    invalid(
      loopHeader +
        ' { save(names[index]in object, scores[index]instanceof Number); }',
      zipHeader + ' { save(name_ in object, score instanceof Number); }',
      [1, 1, 1, 76],
    ),
    invalid(
      loopHeader +
        ' { save(names[index]as string, scores[index]satisfies number); }',
      zipHeader + ' { save(name_ as string, score satisfies number); }',
      [1, 1, 1, 76],
      { filename: 'file.ts' },
    ),
    invalid(
      'for (let index = 0; index < Math.min(Cards.length, scores.length); index++) { render(<Card name={Cards[index]} score={scores[index]} />); }',
      'for (const [Card_, score] of Iterator.zip([Cards, scores])) { render(<Card name={Card_} score={score} />); }',
      [1, 1, 1, 76],
      { filename: 'input.jsx' },
    ),
    invalid(
      'declare const names: string[]; declare const scores: readonly number[]; ' +
        loopHeader +
        loopBody,
      'declare const names: string[]; declare const scores: readonly number[]; ' +
        zipHeader +
        zipBody,
      [1, 73, 1, 148],
      {
        filename: 'file.ts',
        languageOptions: { parserOptions: { project: './tsconfig.json' } },
      },
    ),
    invalid(
      'declare const names: Uint8Array; declare const scores: Float64Array; ' +
        loopHeader +
        loopBody,
      'declare const names: Uint8Array; declare const scores: Float64Array; ' +
        zipHeader +
        zipBody,
      [1, 70, 1, 145],
      {
        filename: 'file.ts',
        languageOptions: { parserOptions: { project: './tsconfig.json' } },
      },
    ),
    invalid(
      loopHeader +
        ' { save(names[index], scores[index]); { const names = other; const index = 0; save(names[index]); } }',
      zipHeader +
        ' { save(name_, score); { const names = other; const index = 0; save(names[index]); } }',
      [1, 1, 1, 76],
    ),
    invalid(
      'for (let index = 0; index < Math.min(widgets.length, scores.length); index++) { save(widgets[index], scores[index]); function getWidget() { return widget; } }',
      'for (const [widget_, score] of Iterator.zip([widgets, scores])) { save(widget_, score); function getWidget() { return widget; } }',
      [1, 1, 1, 78],
    ),
    // Conjunctive conditions and cached bounds.
    invalid(
      'for (let index = 0; index < names.length && index < scores.length; index++)' +
        loopBody,
      zipHeader + zipBody,
      [1, 1, 1, 76],
    ),
    invalid(
      'for (let index = 0; names.length > index && index < scores.length; index++)' +
        loopBody,
      zipHeader + zipBody,
      [1, 1, 1, 76],
    ),
    invalid(
      'for (let index = 0; index < names.length && (scores.length > index && index < ranks.length); index++) { save(names[index], scores[index], ranks[index]); }',
      'for (const [name_, score, rank] of Iterator.zip([names, scores, ranks])) { save(name_, score, rank); }',
      [1, 1, 1, 102],
    ),
    invalid(
      'for (let index = 0; index < Math.min(names.length, scores.length); index = index + 1)' +
        loopBody,
      zipHeader + zipBody,
      [1, 1, 1, 86],
    ),
    invalid(
      'for (let index = 0; index < scores.length && index < names.length; index++)' +
        loopBody,
      'for (const [score, name_] of Iterator.zip([scores, names]))' + zipBody,
      [1, 1, 1, 76],
    ),
    invalid(
      'const length = Math.min(names.length, scores.length);\nfor (let index = 0; index < length; index++)' +
        loopBody,
      zipHeader + zipBody,
      [2, 1, 2, 45],
    ),
    invalid(
      'const length = Math.min(names.length, scores.length);\nfor (let index = 0; length > index; index++)' +
        loopBody,
      zipHeader + zipBody,
      [2, 1, 2, 45],
    ),
    invalid(
      'for (let index = 0, length = Math.min(names.length, scores.length); index < length; index++)' +
        loopBody,
      zipHeader + zipBody,
      [1, 1, 1, 93],
    ),
    invalid(
      'const length = Math.min(names.length, scores.length); /* keep cached-bound comment */\nfor (let index = 0; index < length; index++)' +
        loopBody,
      '',
      [2, 1, 2, 45],
    ),
    invalid(
      'let length = Math.min(names.length, scores.length);\nfor (let index = 0; index < length; index++)' +
        loopBody,
      zipHeader + zipBody,
      [2, 1, 2, 45],
    ),
    invalid(
      'const length: number = Math.min(names.length, scores.length);\nfor (let index = 0; index < length; index++)' +
        loopBody,
      zipHeader + zipBody,
      [2, 1, 2, 45],
      { filename: 'file.ts' },
    ),
    invalid(
      loopHeader +
        ' { for (const letter of names[index]) { save(letter, scores[index]); } }',
      zipHeader + ' { for (const letter of name_) { save(letter, score); } }',
      [1, 1, 1, 76],
    ),
    // Documentation example.
    invalid(
      loopHeader + ' {\n\tsave(names[index], scores[index]);\n}',
      zipHeader + ' {\n\tsave(name_, score);\n}',
      [1, 1, 1, 76],
    ),
    // Standalone token-boundary and autofix cases.
    invalid(
      'function pair(names, scores) { ' +
        loopHeader +
        ' { save(typeof(names)[index], scores[index]); return(names)[index]; } }',
      'function pair(names, scores) { for (const [name, score] of Iterator.zip([names, scores])) { save(typeof name, score); return name; } }',
      [1, 32, 1, 107],
      { languageOptions: { globals: { name: 'off' } } },
    ),
    invalid(
      loopHeader + loopBody,
      'for (const [name, score] of Iterator.zip([names, scores])) { save(name, score); }',
      [1, 1, 1, 76],
      { languageOptions: { globals: { name: 'off' } } },
    ),
  ],
});
