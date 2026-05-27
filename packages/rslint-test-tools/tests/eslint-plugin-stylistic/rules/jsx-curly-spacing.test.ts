import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const NO_SPACE_AFTER = `There should be no space after '{'`;
const NO_SPACE_BEFORE = `There should be no space before '}'`;
const SPACE_NEEDED_AFTER = `A space is required after '{'`;
const SPACE_NEEDED_BEFORE = `A space is required before '}'`;
const NO_NEWLINE_AFTER = `There should be no newline after '{'`;
const NO_NEWLINE_BEFORE = `There should be no newline before '}'`;

const MULTILINE_BAR = `<App foo={
bar
} />;`;

ruleTester.run('jsx-curly-spacing', null as never, {
  valid: [
    // default (never; attributes:true, children:false)
    { code: `<App foo={bar} />;` },
    { code: `<App foo={bar}>{bar}</App>;` },
    { code: `<App foo={bar}>{ bar }</App>;` },
    {
      code: `<App foo={{ bar: true, baz: true }}>{{ bar: true, baz: true }}</App>;`,
    },
    { code: `<App>{ foo /* comment */ }</App>` },
    { code: MULTILINE_BAR },
    // attributes: true / false
    { code: `<App foo={bar} />;`, options: [{ attributes: true }] },
    { code: `<App foo={ bar } />;`, options: [{ attributes: false }] },
    // children: true / false
    { code: `<App>{bar}</App>;`, options: [{ children: true }] },
    { code: `<App>{ bar }</App>;`, options: [{ children: false }] },
    // when never / always (object form)
    { code: `<App foo={bar} />;`, options: [{ when: 'never' }] },
    { code: `<App foo={ bar } />;`, options: [{ when: 'always' }] },
    // spacing.objectLiterals
    {
      code: `<App foo={{ bar: true }} />;`,
      options: [{ when: 'never', spacing: { objectLiterals: 'never' } }],
    },
    {
      code: `<App foo={ { bar: true } } />;`,
      options: [{ when: 'never', spacing: { objectLiterals: 'always' } }],
    },
    // attributes / children object: when overrides
    {
      code: `<App foo={ bar } />;`,
      options: [{ attributes: { when: 'always' } }],
    },
    {
      code: `<App>{ bar }</App>;`,
      options: [{ children: { when: 'always' } }],
    },
    // spread attribute
    { code: `<App {...bar} />;` },
    {
      code: `<App { ...bar } />;`,
      options: [{ attributes: { when: 'always' } }],
    },
    // string shorthand
    { code: `<App foo={bar} />;`, options: ['never'] },
    { code: `<App foo={ bar } />;`, options: ['always'] },
    // fragment children
    { code: `<>{bar} {baz}</>;` },
  ],

  invalid: [
    // default: attribute brace spacing reported, child untouched
    {
      code: `<App foo={ bar }>{bar}</App>;`,
      output: `<App foo={bar}>{bar}</App>;`,
      errors: [
        { messageId: 'noSpaceAfter', message: NO_SPACE_AFTER },
        { messageId: 'noSpaceBefore', message: NO_SPACE_BEFORE },
      ],
    },
    // attributes: true
    {
      code: `<App foo={ bar } />;`,
      options: [{ attributes: true }],
      output: `<App foo={bar} />;`,
      errors: [
        { messageId: 'noSpaceAfter', message: NO_SPACE_AFTER },
        { messageId: 'noSpaceBefore', message: NO_SPACE_BEFORE },
      ],
    },
    // children: true
    {
      code: `<App>{ bar }</App>;`,
      options: [{ children: true }],
      output: `<App>{bar}</App>;`,
      errors: [
        { messageId: 'noSpaceAfter', message: NO_SPACE_AFTER },
        { messageId: 'noSpaceBefore', message: NO_SPACE_BEFORE },
      ],
    },
    // when never: space after only
    {
      code: `<App foo={ bar} />;`,
      options: [{ attributes: { when: 'never' } }],
      output: `<App foo={bar} />;`,
      errors: [{ messageId: 'noSpaceAfter', message: NO_SPACE_AFTER }],
    },
    // when never: space before only
    {
      code: `<App foo={bar } />;`,
      options: [{ attributes: { when: 'never' } }],
      output: `<App foo={bar} />;`,
      errors: [{ messageId: 'noSpaceBefore', message: NO_SPACE_BEFORE }],
    },
    // when always: both sides missing
    {
      code: `<App foo={bar} />;`,
      options: [{ when: 'always' }],
      output: `<App foo={ bar } />;`,
      errors: [
        { messageId: 'spaceNeededAfter', message: SPACE_NEEDED_AFTER },
        { messageId: 'spaceNeededBefore', message: SPACE_NEEDED_BEFORE },
      ],
    },
    // when always: space before only (missing after)
    {
      code: `<App foo={bar } />;`,
      options: [{ attributes: { when: 'always' } }],
      output: `<App foo={ bar } />;`,
      errors: [{ messageId: 'spaceNeededAfter', message: SPACE_NEEDED_AFTER }],
    },
    // when always: space after only (missing before)
    {
      code: `<App foo={ bar} />;`,
      options: [{ attributes: { when: 'always' } }],
      output: `<App foo={ bar } />;`,
      errors: [
        { messageId: 'spaceNeededBefore', message: SPACE_NEEDED_BEFORE },
      ],
    },
    // objectLiterals always promotes a flush object literal
    {
      code: `<App foo={{ bar: true }} />;`,
      options: [{ when: 'never', spacing: { objectLiterals: 'always' } }],
      output: `<App foo={ { bar: true } } />;`,
      errors: [
        { messageId: 'spaceNeededAfter', message: SPACE_NEEDED_AFTER },
        { messageId: 'spaceNeededBefore', message: SPACE_NEEDED_BEFORE },
      ],
    },
    // objectLiterals never strips spaces from a spaced object literal
    {
      code: `<App foo={ { bar: true } } />;`,
      options: [{ when: 'always', spacing: { objectLiterals: 'never' } }],
      output: `<App foo={{ bar: true }} />;`,
      errors: [
        { messageId: 'noSpaceAfter', message: NO_SPACE_AFTER },
        { messageId: 'noSpaceBefore', message: NO_SPACE_BEFORE },
      ],
    },
    // children object: always
    {
      code: `<App>{bar}</App>;`,
      options: [{ children: { when: 'always' } }],
      output: `<App>{ bar }</App>;`,
      errors: [
        { messageId: 'spaceNeededAfter', message: SPACE_NEEDED_AFTER },
        { messageId: 'spaceNeededBefore', message: SPACE_NEEDED_BEFORE },
      ],
    },
    // spread attribute: never
    {
      code: `<App { ...bar } />;`,
      options: [{ attributes: { when: 'never' } }],
      output: `<App {...bar} />;`,
      errors: [
        { messageId: 'noSpaceAfter', message: NO_SPACE_AFTER },
        { messageId: 'noSpaceBefore', message: NO_SPACE_BEFORE },
      ],
    },
    // spread attribute: always
    {
      code: `<App {...bar} />;`,
      options: [{ attributes: { when: 'always' } }],
      output: `<App { ...bar } />;`,
      errors: [
        { messageId: 'spaceNeededAfter', message: SPACE_NEEDED_AFTER },
        { messageId: 'spaceNeededBefore', message: SPACE_NEEDED_BEFORE },
      ],
    },
    // multi-line: never + allowMultiline:false → noNewline both sides
    {
      code: MULTILINE_BAR,
      options: [{ when: 'never', allowMultiline: false }],
      output: `<App foo={bar} />;`,
      errors: [
        { messageId: 'noNewlineAfter', message: NO_NEWLINE_AFTER },
        { messageId: 'noNewlineBefore', message: NO_NEWLINE_BEFORE },
      ],
    },
    // multi-line: always + allowMultiline:false → noNewline both sides
    {
      code: MULTILINE_BAR,
      options: [{ when: 'always', allowMultiline: false }],
      output: `<App foo={ bar } />;`,
      errors: [
        { messageId: 'noNewlineAfter', message: NO_NEWLINE_AFTER },
        { messageId: 'noNewlineBefore', message: NO_NEWLINE_BEFORE },
      ],
    },
    // string shorthand: always
    {
      code: `<App foo={bar} />;`,
      options: ['always'],
      output: `<App foo={ bar } />;`,
      errors: [
        { messageId: 'spaceNeededAfter', message: SPACE_NEEDED_AFTER },
        { messageId: 'spaceNeededBefore', message: SPACE_NEEDED_BEFORE },
      ],
    },
    // string shorthand + objectLiterals always
    {
      code: `<App foo={3} bar={{a: 2}} />`,
      options: ['never', { spacing: { objectLiterals: 'always' } }],
      output: `<App foo={3} bar={ {a: 2} } />`,
      errors: [
        { messageId: 'spaceNeededAfter', message: SPACE_NEEDED_AFTER },
        { messageId: 'spaceNeededBefore', message: SPACE_NEEDED_BEFORE },
      ],
    },
  ],
});
