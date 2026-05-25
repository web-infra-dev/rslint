import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-no-leaked-render', {} as never, {
  valid: [
    // ---- Default options (validStrategies = [ternary, coerce]) ----
    { code: `const C = () => <div>{customTitle || defaultTitle}</div>` },
    { code: `const C = ({ elements }) => <div>{elements}</div>` },
    {
      code: `const C = ({ elements }) => <div>There are {elements.length} elements</div>`,
    },
    {
      code: `const C = ({ count }) => <div>{!count && 'No results found'}</div>`,
    },
    {
      code: `const C = ({ elements }) => <div>{!!elements.length && <List/>}</div>`,
    },
    {
      code: `const C = ({ elements }) => <div>{Boolean(elements.length) && <List/>}</div>`,
    },
    {
      code: `const C = ({ elements }) => <div>{elements.length > 0 && <List/>}</div>`,
    },
    {
      code: `const C = ({ elements }) => <div>{elements.length ? <List/> : null}</div>`,
    },

    // ---- coerce-only ----
    {
      code: `const C = ({ count }) => <div>{!!count && <List/>}</div>`,
      options: [{ validStrategies: ['coerce'] }],
    },
    {
      code: `const C = ({ direction }) => <div>{!!direction && direction === "down" && "▼"}</div>`,
      options: [{ validStrategies: ['coerce'] }],
    },

    // ---- ternary-only ----
    {
      code: `const C = ({ count }) => <div>{count ? <List/> : null}</div>`,
      options: [{ validStrategies: ['ternary'] }],
    },

    // ---- ignoreAttributes mutes attribute-only reports ----
    {
      code: `const C = ({ enabled, checked }) => <CheckBox checked={enabled && checked} />`,
      options: [{ ignoreAttributes: true }],
    },

    // ---- Default react version (latest) → React 18+ behavior:
    //      '' empty string is safe on the left of `&&` ----
    { code: `const C = () => <>{'' && <Foo/>}</>` },
  ],
  invalid: [
    // ---- Default: ternary fix ----
    {
      code: `const C = ({ count, title }) => <div>{count && title}</div>`,
      errors: [{ message: /Potential leaked value/ }],
    },
    {
      code: `const C = ({ count }) => <div>{count && <span>{count}</span>}</div>`,
      errors: [{ message: /Potential leaked value/ }],
    },
    {
      code: `const C = ({ elements }) => <div>{elements.length && <List/>}</div>`,
      errors: [{ message: /Potential leaked value/ }],
    },
    {
      code: `const C = ({ a, b }) => <div>{(a || b) && <Results/>}</div>`,
      errors: [{ message: /Potential leaked value/ }],
    },

    // ---- coerce-only ----
    {
      code: `const C = ({ count, title }) => <div>{count && title}</div>`,
      options: [{ validStrategies: ['coerce'] }],
      errors: [{ message: /Potential leaked value/ }],
    },
    {
      code: `const C = ({ a, b }) => <div>{(a || b) && <Results/>}</div>`,
      options: [{ validStrategies: ['coerce'] }],
      errors: [{ message: /Potential leaked value/ }],
    },
    // ternary not allowed under coerce-only.
    {
      code: `const C = ({ count, title }) => <div>{count ? title : null}</div>`,
      options: [{ validStrategies: ['coerce'] }],
      errors: [{ message: /Potential leaked value/ }],
    },
    // Inverse ternary `cond ? false : alt`.
    {
      code: `const C = () => <Something checked={isIndeterminate ? false : isChecked} />`,
      options: [{ validStrategies: ['coerce'] }],
      errors: [{ message: /Potential leaked value/ }],
    },

    // ---- ternary-only — boolean coercion on the left is not a free pass ----
    {
      code: `const C = ({ count, title }) => <div>{!!count && title}</div>`,
      options: [{ validStrategies: ['ternary'] }],
      errors: [{ message: /Potential leaked value/ }],
    },
    {
      code: `const C = ({ count, title }) => <div>{count > 0 && title}</div>`,
      options: [{ validStrategies: ['ternary'] }],
      errors: [{ message: /Potential leaked value/ }],
    },

    // ---- ignoreAttributes does NOT silence children-level reports ----
    {
      code: `const C = ({ enabled }) => <Foo bar={<Something>{enabled && <MuchWow/>}</Something>} />`,
      options: [{ ignoreAttributes: true }],
      errors: [{ message: /Potential leaked value/ }],
    },

    // ---- Const-bound non-boolean Identifier still leaks ----
    {
      code: `const isOpen = 0; const C = () => <Popover open={isOpen && items.length > 0} />`,
      options: [{ validStrategies: ['coerce'] }],
      errors: [{ message: /Potential leaked value/ }],
    },
  ],
});
