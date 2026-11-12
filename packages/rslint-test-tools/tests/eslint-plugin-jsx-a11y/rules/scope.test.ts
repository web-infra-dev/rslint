import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const componentsSettings = {
  'jsx-a11y': {
    components: {
      Foo: 'div',
      TableHeader: 'th',
    },
  },
};

new RuleTester(config).run('scope', {
  valid: [
    // ---- Upstream valid ----
    { code: '<div />;' },
    { code: '<div foo />;' },
    { code: '<th scope />' },
    { code: '<th scope="row" />' },
    { code: '<th scope={foo} />' },
    { code: '<th scope={"col"} {...props} />' },
    { code: '<Foo scope="bar" {...props} />' },
    { code: '<TableHeader scope="row" />', settings: componentsSettings },

    // ---- Case-insensitive prop name match: th branch exempts every variant ----
    { code: '<th SCOPE />' },
    { code: '<th Scope />' },
    { code: '<th SCoPE />' },
    { code: '<th scOpe="row" />' },

    // ---- Non-scope attribute names with similar spellings ----
    { code: '<div scoped />' },
    { code: '<div scope-of />' },

    // ---- Namespaced attribute names — composite name !== "scope" ----
    { code: '<div xml:scope />' },
    { code: '<th xml:scope />' },

    // ---- th case-sensitivity: dom-set lookup is lowercase, so <TH> skips ----
    { code: '<TH scope />' },
    { code: '<Th scope="row" />' },

    // ---- Spread attributes — not JsxAttribute, listener never fires ----
    { code: "<div {...{scope: 'row'}} />" },
    { code: '<div {...props} />' },

    // ---- components map: non-DOM target stays skipped ----
    {
      code: '<Foo scope="row" />',
      settings: { 'jsx-a11y': { components: { Foo: 'Bar' } } },
    },
    // Empty `jsx-a11y` settings — defensive: rawType "Foo" not in dom → skip.
    { code: '<Foo scope />', settings: { 'jsx-a11y': {} } },

    // ---- polymorphicPropName resolving to th — exempt ----
    {
      code: '<Box as="th" scope />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },
    {
      code: '<Box as="th" scope="col" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },

    // ---- polymorphicPropName resolving to non-DOM — skipped ----
    {
      code: '<Box as="ComponentName" scope />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },

    // ---- polymorphicAllowList restricts the swap ----
    {
      code: '<Box as="th" scope="row" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: ['Box'],
        },
      },
    },
    {
      code: '<Other as="div" scope="row" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: ['Box'],
        },
      },
    },

    // ---- Member-expression / namespaced tag names — non-DOM ----
    { code: '<UX.Layout scope />' },
    { code: '<this.Foo scope="col" />' },
    { code: '<svg:circle scope="row" />' },

    // ---- th in nested / paired forms (still exempt) ----
    { code: '<table><tr><th scope="col">Header</th></tr></table>' },
    { code: '<th scope="row">Header text</th>' },

    // ---- scope value variants on th — all exempt ----
    { code: '<th scope="col" />' },
    { code: '<th scope="rowgroup" />' },
    { code: '<th scope="colgroup" />' },
    { code: '<th scope={someVar} />' },
    { code: "<th scope={cond ? 'row' : 'col'} />" },
    { code: '<th scope={getScope()} />' },
    { code: '<th scope={`${dynamic}`} />' },

    // ---- TS generic JSX components → th ----
    {
      code: '<List<string> scope="col" />',
      settings: {
        'jsx-a11y': {
          components: {
            List: 'th',
          },
        },
      },
    },

    // ---- TS wrappers around the value — irrelevant to rule ----
    { code: '<th scope={"col" as string} />' },
    { code: '<th scope={"col"!} />' },
    { code: '<th scope={"col" satisfies string} />' },
    { code: '<th scope={("col")} />' },

    // ---- Hyphenated DOM tags (web components) ----
    { code: '<my-element scope />' },
    { code: '<my-element scope="row" />' },

    // ---- Real-world th in semantic table patterns ----
    {
      code: 'function DataTable() { return <table><thead><tr><th scope="col">Name</th><th scope="col">Age</th></tr></thead></table>; }',
    },

    // ---- Comments around / inside the prop don't break extraction ----
    { code: '<th /* before */ scope="col" /* after */ />' },
    { code: '<th scope={/* col */ "col"} />' },

    // ---- Multiple scope attributes on th — each fires, both exempt ----
    { code: '<th scope="row" scope="col" />' },
    { code: '<th scope scope="col" />' },

    // ---- SVG / MathML primitives — not in aria-query's dom map → silent skip ----
    { code: '<rect scope />' },
    { code: '<circle scope="row" />' },
    { code: '<path scope />' },
    { code: '<polygon scope />' },
    { code: '<g scope="col" />' },
    { code: '<text scope />' },
    { code: '<math scope />' },
    { code: '<mn scope />' },
    { code: '<mo scope />' },
    // ---- Modern HTML not in aria-query's dom map ----
    { code: '<template scope />' },
    { code: '<slot scope />' },

    // ---- Spread + scope on th ----
    { code: '<th {...props} scope />' },
    { code: '<th scope {...props} />' },
    { code: '<th id="x" {...props} scope="row" className="y" />' },
    { code: '<th {...a} {...b} scope />' },

    // ---- Component library patterns without polymorphic config ----
    { code: '<TableCell component="th" scope="row" />' },
    { code: '<TableCell as="th" scope="row" />' },

    // ---- JSX as render prop / children-as-function (exempt body) ----
    { code: '<DataTable render={() => <th scope="col" />} />' },
    { code: '<DataTable>{() => <th scope="row" />}</DataTable>' },

    // ---- cloneElement / wrapper patterns with th ----
    { code: 'cloneElement(<th scope />);' },
    { code: '<Provider value={data}>{<th scope="row" />}</Provider>' },

    // ---- th with various body shapes ----
    { code: '<th scope="row">Some text</th>' },
    { code: '<th scope="row">{title}</th>' },
    { code: '<th scope="row">Hello, {name}!</th>' },
    { code: '<th scope="row"></th>' },

    // ---- Whitespace / formatting variations on th ----
    { code: '<th  scope  />' },
    { code: '<th scope = "row" />' },
    { code: '<th\n\tscope\n/>' },

    // ---- Deep table nesting ----
    {
      code: '<table><thead><tr><th scope="col" /></tr></thead><tbody><tr><th scope="row" /></tr></tbody></table>',
    },
    // Map / iteration over headers.
    {
      code: '<table><thead><tr>{headers.map(h => <th scope="col" key={h.id}>{h.label}</th>)}</tr></thead></table>',
    },

    // ---- th with sibling event handlers ----
    { code: '<th scope="col" onClick={fn} onKeyDown={fn} />' },
    { code: '<th onClick={fn} scope="col" />' },

    // ---- Polymorphic + components chain to th ----
    {
      code: '<Box as="TableHeader" scope="row" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          components: { TableHeader: 'th' },
        },
      },
    },

    // ---- Settings with no jsx-a11y key ----
    {
      code: '<th scope="row" />',
      settings: { 'other-plugin': { foo: 'bar' } },
    },
    {
      code: '<Foo scope />',
      settings: { 'some-other': 'value' },
    },

    // ---- Components map with dotted key remaps the member-expression tag ----
    // Both upstream and rslint use plain `componentMap[finalType]` lookup,
    // so dotted keys match. `DataGrid.Header → th` makes it th-exempt.
    {
      code: '<DataGrid.Header scope="col" />',
      settings: {
        'jsx-a11y': { components: { 'DataGrid.Header': 'th' } },
      },
    },
    // ---- polymorphicAllowList with only non-string entries — no swap → not in dom ----
    {
      code: '<Other as="div" scope />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: [123, true, null],
        },
      },
    },
  ],
  invalid: [
    // ---- Upstream invalid ----
    {
      code: '<div scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '<Foo scope="bar" />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },

    // ---- Case-insensitive prop name match → still reports on non-th ----
    {
      code: '<div SCOPE />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '<div Scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '<div SCoPe="row" />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<span scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },

    // ---- Listener boundary — nested elements report independently ----
    {
      code: '<div scope><span scope /></div>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<th scope="col"><div scope /></th>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    // Multiple scope attributes on one non-th element each report.
    {
      code: '<div scope scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },

    // ---- Element kind survey — every DOM tag fires ----
    {
      code: '<a scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: '<button scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<table scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<tr scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<td scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<thead scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<tbody scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<tfoot scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<input scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },

    // ---- scope value variants on a non-th element — all report ----
    {
      code: '<div scope="col" />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<div scope={"row"} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div scope={someVar} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div scope={null} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div scope={undefined} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div scope={0} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<div scope={false} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div scope={""} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '<div scope={fn()} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div scope={obj.x} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div scope={obj?.x} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div scope={`row`} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div scope={`${x}`} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },

    // ---- TS wrappers around the value — irrelevant to rule ----
    {
      code: '<div scope={"col" as string} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div scope={("col")} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div scope={"col"!} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div scope={"col" satisfies string} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },

    // ---- components map flips custom INTO scope-rule coverage ----
    {
      code: '<Cell scope="col" />',
      settings: { 'jsx-a11y': { components: { Cell: 'td' } } },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<Wrapper scope="row" />',
      settings: { 'jsx-a11y': { components: { Wrapper: 'section' } } },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },

    // ---- polymorphicPropName resolving to non-th DOM → reports ----
    {
      code: '<Box as="div" scope />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<Box as="span" scope="col" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<Box as="div" scope />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: ['Box'],
        },
      },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },

    // ---- TypeScript generic JSX components → reports via components map ----
    {
      code: '<List<string> scope />',
      settings: { 'jsx-a11y': { components: { List: 'div' } } },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },

    // ---- Comments around / inside the prop don't suppress ----
    {
      code: '<div /* a */ scope /* b */ />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div scope={/* row */ "row"} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },

    // ---- Real-world misuse patterns ----
    {
      code: 'function Header({ title }) { return <h1 scope="col">{title}</h1>; }',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: '<div role="columnheader" scope="col">Header</div>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'function Cell({ value }) { return <td scope="col">{value}</td>; }',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },

    // ---- Multiple components in one file ----
    {
      code: 'function A() { return <div scope />; }\nfunction B() { return <span scope />; }',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 33,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 2,
          column: 29,
          endLine: 2,
          endColumn: 34,
        },
      ],
    },

    // ---- List-rendered JSX with offending scope ----
    {
      code: 'const items = arr.map((x, i) => <div key={i} scope={x.scope} />);',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 46,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },

    // ---- Class component with state-driven JSX ----
    {
      code: 'class T extends React.Component { render() { return this.state.ready ? <div scope="row" /> : null; } }',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 77,
          endLine: 1,
          endColumn: 88,
        },
      ],
    },

    // ---- Generator / async / IIFE bodies ----
    {
      code: 'function* render() { yield <div scope />; yield <span scope />; }',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 38,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 55,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
    {
      code: 'async function render() { return <div scope />; }',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      code: 'const x = (() => <div scope />)();',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },

    // ---- Fragment + conditional rendering ----
    {
      code: 'const x = <>{cond && <div scope />}</>;',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'function Foo({a, b}) { return a ? <div scope /> : <span scope />; }',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 45,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 57,
          endLine: 1,
          endColumn: 62,
        },
      ],
    },

    // ---- HOC / forwardRef / memo wrappers carrying scope ----
    {
      code: 'const Enhanced = withTracking(({ value }) => <div value={value} scope />);',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 65,
          endLine: 1,
          endColumn: 70,
        },
      ],
    },
    {
      code: 'const FocusInput = React.forwardRef((props, ref) => <div ref={ref} scope {...props} />);',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 68,
          endLine: 1,
          endColumn: 73,
        },
      ],
    },
    {
      code: 'const Item = React.memo(({ id }) => <li id={id} scope>{id}</li>);',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },

    // ---- Extended DOM element survey ----
    {
      code: '<article scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<section scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<aside scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<header scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<footer scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<nav scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '<main scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: '<h1 scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<h6 scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<form scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: '<fieldset scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '<label scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<select scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<option scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<textarea scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '<dialog scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<details scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<summary scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<canvas scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<picture scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<video scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<audio scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<iframe scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    // Table-related neighbors of th — confusable, all illegal.
    {
      code: '<caption scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<colgroup scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '<col scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },

    // ---- Spread + scope mixing on non-th (literal scope reports) ----
    {
      code: '<div {...props} scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div scope {...props} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '<div {...a} scope {...b} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: "<div {...{scope: 'row'}} scope />",
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // ---- Multi-element multi-error scenarios ----
    {
      code: '<><th scope /><div scope /><span scope /></>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 25,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: '<thead scope><tr><th scope="col" /></tr></thead>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<><div scope /><span scope /><th scope /><a scope /></>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 27,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },

    // ---- JSX as render prop / children-as-function (offending body) ----
    {
      code: '<DataTable render={() => <div scope />} />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<DataTable>{() => <div scope />}</DataTable>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },

    // ---- JSX in array literal / cloneElement / Provider (offending) ----
    {
      code: 'const items = [<div scope key="1" />, <span scope key="2" />, <th scope key="3" />];',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 26,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: 'cloneElement(<div scope />);',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<Provider value={data}>{<div scope />}</Provider>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'const obj = { content: <div scope /> };',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },

    // ---- Whitespace / formatting variations on non-th ----
    {
      code: '<div  scope  />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: '<div scope = "row" />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div\n\tscope\n/>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 7,
        },
      ],
    },

    // ---- scope at varied positions ----
    {
      code: '<div scope id="x" className="y" />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '<div id="x" scope className="y" />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div id="x" className="y" scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },

    // ---- th sibling with non-th sibling carrying scope ----
    {
      code: '<><th scope="col" /><td scope="row" /></>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<><th scope /><tr scope /><td scope /></>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 24,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },

    // ---- Component library patterns — components map → non-th ----
    {
      code: '<TableCell scope="col" />',
      settings: { 'jsx-a11y': { components: { TableCell: 'td' } } },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },

    // ---- polymorphic + components chain → non-th ----
    {
      code: '<Box as="DataCell" scope />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          components: { DataCell: 'td' },
        },
      },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },

    // ---- Suspense / ErrorBoundary / Context wrapping offending JSX ----
    {
      code: '<Suspense fallback={<div scope />}><div>x</div></Suspense>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<ErrorBoundary><div scope /></ErrorBoundary>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<Context.Provider value={data}><div scope /></Context.Provider>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },

    // ---- TS-only JSX shapes ----
    {
      code: 'const x = <div scope /> as React.ReactNode;',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'const x = (<div scope />) satisfies any;',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'const x = (<div scope />)!;',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },

    // ---- TypeScript generic JSX with components remap ----
    {
      code: '<List<Header> scope />',
      settings: { 'jsx-a11y': { components: { List: 'div' } } },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },

    // ---- Realistic table cell misuse patterns ----
    {
      code: '<table><tr><td scope="col" /><td scope="col" /></tr><tr><td scope="row" /><td scope="row" /></tr></table>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 27,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 45,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 61,
          endLine: 1,
          endColumn: 72,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 79,
          endLine: 1,
          endColumn: 90,
        },
      ],
    },
    {
      code: '<table><tr><th scope="col" /><td scope="col" /><th scope="col" /><td scope="col" /></tr></table>',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 45,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 70,
          endLine: 1,
          endColumn: 81,
        },
      ],
    },

    // ---- try / catch / finally JSX ----
    {
      code: 'function App() { try { return <div scope />; } catch (e) { return <span scope />; } }',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 41,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 73,
          endLine: 1,
          endColumn: 78,
        },
      ],
    },

    // ---- switch returning JSX ----
    {
      code: "function App({type}) { switch(type) { case 'a': return <div scope />; case 'b': return <span scope />; default: return null; } }",
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 61,
          endLine: 1,
          endColumn: 66,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 94,
          endLine: 1,
          endColumn: 99,
        },
      ],
    },

    // ---- Class with class-field arrow returning JSX ----
    {
      code: 'class C { render = () => <div scope />; }',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },

    // ---- JSX in attribute value triggers nested + outer reports ----
    {
      code: '<div header={<span scope />} scope />',
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 25,
        },
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },

    // ---- Web Components via components map ----
    {
      code: '<my-table scope />',
      settings: { 'jsx-a11y': { components: { 'my-table': 'td' } } },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },

    // ---- Components map with dotted key remapping member-expression tag to non-th DOM ----
    {
      code: '<DataGrid.Header scope="col" />',
      settings: {
        'jsx-a11y': { components: { 'DataGrid.Header': 'div' } },
      },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    // ---- Namespaced tag remapped via composite key ----
    {
      code: '<svg:circle scope />',
      settings: {
        'jsx-a11y': { components: { 'svg:circle': 'div' } },
      },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },

    // ---- Polymorphic prop value forms ----
    {
      code: '<Box as={"div"} scope />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<Box as={`div`} scope />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },

    // ---- Defensive: invalid settings types should still fall back ----
    {
      code: '<div scope />',
      settings: {
        'jsx-a11y': { components: 'invalid' as unknown as object },
      },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '<div as="th" scope />',
      settings: {
        'jsx-a11y': { polymorphicPropName: 123 as unknown as string },
      },
      errors: [
        {
          messageId: 'scopeOnTh',
          message: 'The scope prop can only be used on <th> elements.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
  ],
});
