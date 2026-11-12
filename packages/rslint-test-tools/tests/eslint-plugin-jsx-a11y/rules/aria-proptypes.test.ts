import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// Mirrors aria-query's `ariaPropsMap` — keyed by lowercase canonical name.
// Used to drive procedural error-message generation, matching upstream's
// `__tests__/src/rules/aria-proptypes-test.js` local `errorMessage(name)`
// helper which looks up `aria.get(name.toLowerCase())`.

// Mirrors the rule's `errorMessage(name, type, permittedValues)`.
// `${permittedValues}` interpolation hits `Array.prototype.toString` and
// joins with a bare comma (no space). Booleans stringify as "true"/"false".

new RuleTester(config).run('aria-proptypes', {
  valid: [
    // Upstream test file — order preserved.
    { code: '<div aria-foo="true" />' },
    { code: '<div abcaria-foo="true" />' },
    { code: '<div aria-hidden={true} />' },
    { code: '<div aria-hidden="true" />' },
    { code: '<div aria-hidden={"false"} />' },
    { code: '<div aria-hidden={!false} />' },
    { code: '<div aria-hidden />' },
    { code: '<div aria-hidden={false} />' },
    { code: '<div aria-hidden={!true} />' },
    { code: '<div aria-hidden={!"yes"} />' },
    { code: '<div aria-hidden={foo} />' },
    { code: '<div aria-hidden={foo.bar} />' },
    { code: '<div aria-hidden={null} />' },
    { code: '<div aria-hidden={undefined} />' },
    { code: '<div aria-hidden={<div />} />' },
    { code: '<div aria-label="Close" />' },
    { code: '<div aria-label={`Close`} />' },
    { code: '<div aria-label={foo} />' },
    { code: '<div aria-label={foo.bar} />' },
    { code: '<div aria-label={null} />' },
    { code: '<div aria-label={undefined} />' },
    { code: '<input aria-invalid={error ? "true" : "false"} />' },
    { code: '<input aria-invalid={undefined ? "true" : "false"} />' },
    { code: '<div aria-checked={true} />' },
    { code: '<div aria-checked="true" />' },
    { code: '<div aria-checked={"false"} />' },
    { code: '<div aria-checked={!false} />' },
    { code: '<div aria-checked />' },
    { code: '<div aria-checked={false} />' },
    { code: '<div aria-checked={!true} />' },
    { code: '<div aria-checked={!"yes"} />' },
    { code: '<div aria-checked={foo} />' },
    { code: '<div aria-checked={foo.bar} />' },
    { code: '<div aria-checked="mixed" />' },
    { code: '<div aria-checked={`mixed`} />' },
    { code: '<div aria-checked={null} />' },
    { code: '<div aria-checked={undefined} />' },
    { code: '<div aria-level={123} />' },
    { code: '<div aria-level={-123} />' },
    { code: '<div aria-level={+123} />' },
    { code: '<div aria-level={~123} />' },
    { code: '<div aria-level={"123"} />' },
    { code: '<div aria-level={`123`} />' },
    { code: '<div aria-level="123" />' },
    { code: '<div aria-level={foo} />' },
    { code: '<div aria-level={foo.bar} />' },
    { code: '<div aria-level={null} />' },
    { code: '<div aria-level={undefined} />' },
    { code: '<div aria-valuemax={123} />' },
    { code: '<div aria-valuemax={-123} />' },
    { code: '<div aria-valuemax={+123} />' },
    { code: '<div aria-valuemax={~123} />' },
    { code: '<div aria-valuemax={"123"} />' },
    { code: '<div aria-valuemax={`123`} />' },
    { code: '<div aria-valuemax="123" />' },
    { code: '<div aria-valuemax={foo} />' },
    { code: '<div aria-valuemax={foo.bar} />' },
    { code: '<div aria-valuemax={null} />' },
    { code: '<div aria-valuemax={undefined} />' },
    { code: '<div aria-sort="ascending" />' },
    { code: '<div aria-sort="ASCENDING" />' },
    { code: '<div aria-sort={"ascending"} />' },
    { code: '<div aria-sort={`ascending`} />' },
    { code: '<div aria-sort="descending" />' },
    { code: '<div aria-sort={"descending"} />' },
    { code: '<div aria-sort={`descending`} />' },
    { code: '<div aria-sort="none" />' },
    { code: '<div aria-sort={"none"} />' },
    { code: '<div aria-sort={`none`} />' },
    { code: '<div aria-sort="other" />' },
    { code: '<div aria-sort={"other"} />' },
    { code: '<div aria-sort={`other`} />' },
    { code: '<div aria-sort={foo} />' },
    { code: '<div aria-sort={foo.bar} />' },
    { code: '<div aria-invalid={true} />' },
    { code: '<div aria-invalid="true" />' },
    { code: '<div aria-invalid={false} />' },
    { code: '<div aria-invalid="false" />' },
    { code: '<div aria-invalid="grammar" />' },
    { code: '<div aria-invalid="spelling" />' },
    { code: '<div aria-invalid={null} />' },
    { code: '<div aria-invalid={undefined} />' },
    { code: '<div aria-relevant="additions" />' },
    { code: '<div aria-relevant={"additions"} />' },
    { code: '<div aria-relevant={`additions`} />' },
    { code: '<div aria-relevant="additions removals" />' },
    { code: '<div aria-relevant="additions additions" />' },
    { code: '<div aria-relevant={"additions removals"} />' },
    { code: '<div aria-relevant={`additions removals`} />' },
    { code: '<div aria-relevant="additions removals text" />' },
    { code: '<div aria-relevant={"additions removals text"} />' },
    { code: '<div aria-relevant={`additions removals text`} />' },
    { code: '<div aria-relevant="additions removals text all" />' },
    { code: '<div aria-relevant={"additions removals text all"} />' },
    { code: '<div aria-relevant={`removals additions text all`} />' },
    { code: '<div aria-relevant={foo} />' },
    { code: '<div aria-relevant={foo.bar} />' },
    { code: '<div aria-relevant={null} />' },
    { code: '<div aria-relevant={undefined} />' },
    { code: '<div aria-activedescendant="ascending" />' },
    { code: '<div aria-activedescendant="ASCENDING" />' },
    { code: '<div aria-activedescendant={"ascending"} />' },
    { code: '<div aria-activedescendant={`ascending`} />' },
    { code: '<div aria-activedescendant="descending" />' },
    { code: '<div aria-activedescendant={"descending"} />' },
    { code: '<div aria-activedescendant={`descending`} />' },
    { code: '<div aria-activedescendant="none" />' },
    { code: '<div aria-activedescendant={"none"} />' },
    { code: '<div aria-activedescendant={`none`} />' },
    { code: '<div aria-activedescendant="other" />' },
    { code: '<div aria-activedescendant={"other"} />' },
    { code: '<div aria-activedescendant={`other`} />' },
    { code: '<div aria-activedescendant={foo} />' },
    { code: '<div aria-activedescendant={foo.bar} />' },
    { code: '<div aria-activedescendant={null} />' },
    { code: '<div aria-activedescendant={undefined} />' },
    { code: '<div aria-labelledby="additions" />' },
    { code: '<div aria-labelledby={"additions"} />' },
    { code: '<div aria-labelledby={`additions`} />' },
    { code: '<div aria-labelledby="additions removals" />' },
    { code: '<div aria-labelledby="additions additions" />' },
    { code: '<div aria-labelledby={"additions removals"} />' },
    { code: '<div aria-labelledby={`additions removals`} />' },
    { code: '<div aria-labelledby="additions removals text" />' },
    { code: '<div aria-labelledby={"additions removals text"} />' },
    { code: '<div aria-labelledby={`additions removals text`} />' },
    { code: '<div aria-labelledby="additions removals text all" />' },
    { code: '<div aria-labelledby={"additions removals text all"} />' },
    { code: '<div aria-labelledby={`removals additions text all`} />' },
    { code: '<div aria-labelledby={foo} />' },
    { code: '<div aria-labelledby={foo.bar} />' },
    { code: '<div aria-labelledby={null} />' },
    { code: '<div aria-labelledby={undefined} />' },

    // Extra rslint lockdowns:
    // Case-insensitive lookup — uppercase / mixed-case prefixes ARE
    // accepted (unlike aria-props, which is case-sensitive).
    { code: '<div ARIA-HIDDEN="true" />' },
    { code: '<div Aria-Hidden="true" />' },
    { code: '<div ARIA-LEVEL={3} />' },
    // Parenthesized values — tsgo preserves Parens; ESTree flattens.
    { code: '<div aria-hidden={(true)} />' },
    { code: '<div aria-hidden={((true))} />' },
    // TS type-assertion wrappers — LITERAL_TYPES has no entry → noop → skip.
    { code: '<div aria-hidden={true as boolean} />' },
    { code: '<div aria-hidden={x!} />' },
    { code: '<div aria-hidden={x satisfies boolean} />' },
    // aria-haspopup heterogeneous values.
    { code: '<div aria-haspopup={true} />' },
    { code: '<div aria-haspopup={false} />' },
    { code: '<div aria-haspopup="menu" />' },
    // aria-current heterogeneous values.
    { code: '<div aria-current="page" />' },
    { code: '<div aria-current={true} />' },
    // aria-orientation includes string "undefined" as a token value.
    { code: '<div aria-orientation="undefined" />' },
    // JSXSpreadAttribute is not visited.
    { code: "<div {...{'aria-hidden': 'yes'}} />" },
    // data-* prefix is not aria-*.
    { code: '<div data-aria-hidden="yes" />' },
    // React patterns.
    { code: '<>{cond && <div aria-hidden>x</div>}</>' },
    { code: 'class C { render() { return <div aria-label="x" />; } }' },
  ],
  invalid: [
    // Upstream test file — order preserved.
    {
      code: '<div aria-hidden="yes" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-hidden must be a boolean.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div aria-hidden="no" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-hidden must be a boolean.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div aria-hidden={1234} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-hidden must be a boolean.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div aria-hidden={`${abc}`} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-hidden must be a boolean.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<div aria-label />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-label must be a string.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '<div aria-label={true} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-label must be a string.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div aria-label={false} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-label must be a string.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div aria-label={1234} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-label must be a string.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div aria-label={!true} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-label must be a string.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div aria-checked="yes" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-checked must be a boolean or the string "mixed".',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div aria-checked="no" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-checked must be a boolean or the string "mixed".',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div aria-checked={1234} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-checked must be a boolean or the string "mixed".',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<div aria-checked={`${abc}`} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-checked must be a boolean or the string "mixed".',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div aria-level="yes" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-level must be a integer.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div aria-level="no" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-level must be a integer.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div aria-level={`abc`} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-level must be a integer.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div aria-level={true} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-level must be a integer.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div aria-level />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-level must be a integer.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '<div aria-level={"false"} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-level must be a integer.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div aria-level={!"false"} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-level must be a integer.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div aria-valuemax="yes" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-valuemax must be a number.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<div aria-valuemax="no" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-valuemax must be a number.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div aria-valuemax={`abc`} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-valuemax must be a number.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div aria-valuemax={true} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-valuemax must be a number.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div aria-valuemax />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-valuemax must be a number.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div aria-valuemax={"false"} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-valuemax must be a number.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div aria-valuemax={!"false"} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-valuemax must be a number.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div aria-sort="" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-sort must be a single token from the following: ascending,descending,none,other.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div aria-sort="descnding" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-sort must be a single token from the following: ascending,descending,none,other.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div aria-sort />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-sort must be a single token from the following: ascending,descending,none,other.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<div aria-sort={true} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-sort must be a single token from the following: ascending,descending,none,other.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div aria-sort={"false"} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-sort must be a single token from the following: ascending,descending,none,other.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<div aria-sort="ascending descending" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-sort must be a single token from the following: ascending,descending,none,other.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<div aria-relevant="" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-relevant must be a list of one or more tokens from the following: additions,all,removals,text.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div aria-relevant="foobar" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-relevant must be a list of one or more tokens from the following: additions,all,removals,text.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<div aria-relevant />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-relevant must be a list of one or more tokens from the following: additions,all,removals,text.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div aria-relevant={true} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-relevant must be a list of one or more tokens from the following: additions,all,removals,text.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div aria-relevant={"false"} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-relevant must be a list of one or more tokens from the following: additions,all,removals,text.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div aria-relevant="additions removalss" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-relevant must be a list of one or more tokens from the following: additions,all,removals,text.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: '<div aria-relevant="additions removalss " />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-relevant must be a list of one or more tokens from the following: additions,all,removals,text.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },

    // Extra rslint lockdowns.
    // Case-insensitive: message reflects ORIGINAL (cased) name.
    {
      code: '<div ARIA-HIDDEN="yes" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for ARIA-HIDDEN must be a boolean.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div Aria-Hidden={1234} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for Aria-Hidden must be a boolean.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    // Parenthesized invalid still reports.
    {
      code: '<div aria-hidden={(1234)} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-hidden must be a boolean.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    // aria-haspopup unknown token.
    {
      code: '<div aria-haspopup="dropdown" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-haspopup must be a single token from the following: false,true,menu,listbox,tree,grid,dialog.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    // aria-orientation unknown token.
    {
      code: '<div aria-orientation="diagonal" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-orientation must be a single token from the following: vertical,undefined,horizontal.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    // aria-pressed invalid.
    {
      code: '<div aria-pressed="partial" />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message:
            'The value for aria-pressed must be a boolean or the string "mixed".',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    // Nested JSX with invalid attrs at multiple depths.
    {
      code: '<main><section><h2 aria-hidden="yes"><span aria-label={1} /></h2></section></main>',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-hidden must be a boolean.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 37,
        },
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-label must be a string.',
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },
    // Mixed valid + invalid attrs on a single element.
    {
      code: '<div aria-label="ok" aria-hidden="yes" aria-level={3} />',
      errors: [
        {
          messageId: 'invalidAriaPropType',
          message: 'The value for aria-hidden must be a boolean.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
  ],
});
