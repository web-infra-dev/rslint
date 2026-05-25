import { RuleTester } from '../rule-tester';

// Mirrors aria-query's `ariaPropsMap` — keyed by lowercase canonical name.
// Used to drive procedural error-message generation, matching upstream's
// `__tests__/src/rules/aria-proptypes-test.js` local `errorMessage(name)`
// helper which looks up `aria.get(name.toLowerCase())`.
const ariaPropertyDefinitions: Record<
  string,
  { type: string; values?: (string | boolean)[] }
> = {
  'aria-activedescendant': { type: 'id' },
  'aria-atomic': { type: 'boolean' },
  'aria-autocomplete': {
    type: 'token',
    values: ['inline', 'list', 'both', 'none'],
  },
  'aria-busy': { type: 'boolean' },
  'aria-checked': { type: 'tristate' },
  'aria-colcount': { type: 'integer' },
  'aria-controls': { type: 'idlist' },
  'aria-current': {
    type: 'token',
    values: ['page', 'step', 'location', 'date', 'time', true, false],
  },
  'aria-describedby': { type: 'idlist' },
  'aria-details': { type: 'id' },
  'aria-disabled': { type: 'boolean' },
  'aria-dropeffect': {
    type: 'tokenlist',
    values: ['copy', 'execute', 'link', 'move', 'none', 'popup'],
  },
  'aria-errormessage': { type: 'id' },
  'aria-expanded': { type: 'boolean' },
  'aria-grabbed': { type: 'boolean' },
  'aria-haspopup': {
    type: 'token',
    values: [false, true, 'menu', 'listbox', 'tree', 'grid', 'dialog'],
  },
  'aria-hidden': { type: 'boolean' },
  'aria-invalid': {
    type: 'token',
    values: ['grammar', false, 'spelling', true],
  },
  'aria-label': { type: 'string' },
  'aria-labelledby': { type: 'idlist' },
  'aria-level': { type: 'integer' },
  'aria-live': { type: 'token', values: ['assertive', 'off', 'polite'] },
  'aria-orientation': {
    type: 'token',
    values: ['vertical', 'undefined', 'horizontal'],
  },
  'aria-posinset': { type: 'integer' },
  'aria-pressed': { type: 'tristate' },
  'aria-relevant': {
    type: 'tokenlist',
    values: ['additions', 'all', 'removals', 'text'],
  },
  'aria-rowcount': { type: 'integer' },
  'aria-selected': { type: 'boolean' },
  'aria-setsize': { type: 'integer' },
  'aria-sort': {
    type: 'token',
    values: ['ascending', 'descending', 'none', 'other'],
  },
  'aria-valuemax': { type: 'number' },
  'aria-valuemin': { type: 'number' },
  'aria-valuenow': { type: 'number' },
};

// Mirrors the rule's `errorMessage(name, type, permittedValues)`.
// `${permittedValues}` interpolation hits `Array.prototype.toString` and
// joins with a bare comma (no space). Booleans stringify as "true"/"false".
const errorMessage = (name: string): string => {
  const def = ariaPropertyDefinitions[name.toLowerCase()];
  const { type, values: permittedValues = [] } = def;
  switch (type) {
    case 'tristate':
      return `The value for ${name} must be a boolean or the string "mixed".`;
    case 'token':
      return `The value for ${name} must be a single token from the following: ${permittedValues.join(',')}.`;
    case 'tokenlist':
      return `The value for ${name} must be a list of one or more tokens from the following: ${permittedValues.join(',')}.`;
    case 'idlist':
      return `The value for ${name} must be a list of strings that represent DOM element IDs (idlist)`;
    case 'id':
      return `The value for ${name} must be a string that represents a DOM element ID`;
    default:
      return `The value for ${name} must be a ${type}.`;
  }
};

new RuleTester().run('aria-proptypes', null as never, {
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
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    {
      code: '<div aria-hidden="no" />',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    {
      code: '<div aria-hidden={1234} />',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    {
      code: '<div aria-hidden={`${abc}`} />',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    {
      code: '<div aria-label />',
      errors: [{ message: errorMessage('aria-label') }],
    },
    {
      code: '<div aria-label={true} />',
      errors: [{ message: errorMessage('aria-label') }],
    },
    {
      code: '<div aria-label={false} />',
      errors: [{ message: errorMessage('aria-label') }],
    },
    {
      code: '<div aria-label={1234} />',
      errors: [{ message: errorMessage('aria-label') }],
    },
    {
      code: '<div aria-label={!true} />',
      errors: [{ message: errorMessage('aria-label') }],
    },
    {
      code: '<div aria-checked="yes" />',
      errors: [{ message: errorMessage('aria-checked') }],
    },
    {
      code: '<div aria-checked="no" />',
      errors: [{ message: errorMessage('aria-checked') }],
    },
    {
      code: '<div aria-checked={1234} />',
      errors: [{ message: errorMessage('aria-checked') }],
    },
    {
      code: '<div aria-checked={`${abc}`} />',
      errors: [{ message: errorMessage('aria-checked') }],
    },
    {
      code: '<div aria-level="yes" />',
      errors: [{ message: errorMessage('aria-level') }],
    },
    {
      code: '<div aria-level="no" />',
      errors: [{ message: errorMessage('aria-level') }],
    },
    {
      code: '<div aria-level={`abc`} />',
      errors: [{ message: errorMessage('aria-level') }],
    },
    {
      code: '<div aria-level={true} />',
      errors: [{ message: errorMessage('aria-level') }],
    },
    {
      code: '<div aria-level />',
      errors: [{ message: errorMessage('aria-level') }],
    },
    {
      code: '<div aria-level={"false"} />',
      errors: [{ message: errorMessage('aria-level') }],
    },
    {
      code: '<div aria-level={!"false"} />',
      errors: [{ message: errorMessage('aria-level') }],
    },
    {
      code: '<div aria-valuemax="yes" />',
      errors: [{ message: errorMessage('aria-valuemax') }],
    },
    {
      code: '<div aria-valuemax="no" />',
      errors: [{ message: errorMessage('aria-valuemax') }],
    },
    {
      code: '<div aria-valuemax={`abc`} />',
      errors: [{ message: errorMessage('aria-valuemax') }],
    },
    {
      code: '<div aria-valuemax={true} />',
      errors: [{ message: errorMessage('aria-valuemax') }],
    },
    {
      code: '<div aria-valuemax />',
      errors: [{ message: errorMessage('aria-valuemax') }],
    },
    {
      code: '<div aria-valuemax={"false"} />',
      errors: [{ message: errorMessage('aria-valuemax') }],
    },
    {
      code: '<div aria-valuemax={!"false"} />',
      errors: [{ message: errorMessage('aria-valuemax') }],
    },
    {
      code: '<div aria-sort="" />',
      errors: [{ message: errorMessage('aria-sort') }],
    },
    {
      code: '<div aria-sort="descnding" />',
      errors: [{ message: errorMessage('aria-sort') }],
    },
    {
      code: '<div aria-sort />',
      errors: [{ message: errorMessage('aria-sort') }],
    },
    {
      code: '<div aria-sort={true} />',
      errors: [{ message: errorMessage('aria-sort') }],
    },
    {
      code: '<div aria-sort={"false"} />',
      errors: [{ message: errorMessage('aria-sort') }],
    },
    {
      code: '<div aria-sort="ascending descending" />',
      errors: [{ message: errorMessage('aria-sort') }],
    },
    {
      code: '<div aria-relevant="" />',
      errors: [{ message: errorMessage('aria-relevant') }],
    },
    {
      code: '<div aria-relevant="foobar" />',
      errors: [{ message: errorMessage('aria-relevant') }],
    },
    {
      code: '<div aria-relevant />',
      errors: [{ message: errorMessage('aria-relevant') }],
    },
    {
      code: '<div aria-relevant={true} />',
      errors: [{ message: errorMessage('aria-relevant') }],
    },
    {
      code: '<div aria-relevant={"false"} />',
      errors: [{ message: errorMessage('aria-relevant') }],
    },
    {
      code: '<div aria-relevant="additions removalss" />',
      errors: [{ message: errorMessage('aria-relevant') }],
    },
    {
      code: '<div aria-relevant="additions removalss " />',
      errors: [{ message: errorMessage('aria-relevant') }],
    },

    // Extra rslint lockdowns.
    // Case-insensitive: message reflects ORIGINAL (cased) name.
    {
      code: '<div ARIA-HIDDEN="yes" />',
      errors: [{ message: 'The value for ARIA-HIDDEN must be a boolean.' }],
    },
    {
      code: '<div Aria-Hidden={1234} />',
      errors: [{ message: 'The value for Aria-Hidden must be a boolean.' }],
    },
    // Parenthesized invalid still reports.
    {
      code: '<div aria-hidden={(1234)} />',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    // aria-haspopup unknown token.
    {
      code: '<div aria-haspopup="dropdown" />',
      errors: [{ message: errorMessage('aria-haspopup') }],
    },
    // aria-orientation unknown token.
    {
      code: '<div aria-orientation="diagonal" />',
      errors: [{ message: errorMessage('aria-orientation') }],
    },
    // aria-pressed invalid.
    {
      code: '<div aria-pressed="partial" />',
      errors: [{ message: errorMessage('aria-pressed') }],
    },
    // Nested JSX with invalid attrs at multiple depths.
    {
      code: '<main><section><h2 aria-hidden="yes"><span aria-label={1} /></h2></section></main>',
      errors: [
        { message: errorMessage('aria-hidden') },
        { message: errorMessage('aria-label') },
      ],
    },
    // Mixed valid + invalid attrs on a single element.
    {
      code: '<div aria-label="ok" aria-hidden="yes" aria-level={3} />',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
  ],
});
