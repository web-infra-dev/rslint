import { RuleTester } from '../rule-tester';

const expectedError = {
  message: 'The scope prop can only be used on <th> elements.',
};

const componentsSettings = {
  'jsx-a11y': {
    components: {
      Foo: 'div',
      TableHeader: 'th',
    },
  },
};

new RuleTester().run('scope', null as never, {
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
    { code: '<div scope />', errors: [expectedError] },
    {
      code: '<Foo scope="bar" />',
      settings: componentsSettings,
      errors: [expectedError],
    },

    // ---- Case-insensitive prop name match → still reports on non-th ----
    { code: '<div SCOPE />', errors: [expectedError] },
    { code: '<div Scope />', errors: [expectedError] },
    { code: '<div SCoPe="row" />', errors: [expectedError] },
    { code: '<span scope />', errors: [expectedError] },

    // ---- Listener boundary — nested elements report independently ----
    {
      code: '<div scope><span scope /></div>',
      errors: [expectedError, expectedError],
    },
    {
      code: '<th scope="col"><div scope /></th>',
      errors: [expectedError],
    },
    // Multiple scope attributes on one non-th element each report.
    {
      code: '<div scope scope />',
      errors: [expectedError, expectedError],
    },

    // ---- Element kind survey — every DOM tag fires ----
    { code: '<a scope />', errors: [expectedError] },
    { code: '<button scope />', errors: [expectedError] },
    { code: '<table scope />', errors: [expectedError] },
    { code: '<tr scope />', errors: [expectedError] },
    { code: '<td scope />', errors: [expectedError] },
    { code: '<thead scope />', errors: [expectedError] },
    { code: '<tbody scope />', errors: [expectedError] },
    { code: '<tfoot scope />', errors: [expectedError] },
    { code: '<input scope />', errors: [expectedError] },

    // ---- scope value variants on a non-th element — all report ----
    { code: '<div scope="col" />', errors: [expectedError] },
    { code: '<div scope={"row"} />', errors: [expectedError] },
    { code: '<div scope={someVar} />', errors: [expectedError] },
    { code: '<div scope={null} />', errors: [expectedError] },
    { code: '<div scope={undefined} />', errors: [expectedError] },
    { code: '<div scope={0} />', errors: [expectedError] },
    { code: '<div scope={false} />', errors: [expectedError] },
    { code: '<div scope={""} />', errors: [expectedError] },
    { code: '<div scope={fn()} />', errors: [expectedError] },
    { code: '<div scope={obj.x} />', errors: [expectedError] },
    { code: '<div scope={obj?.x} />', errors: [expectedError] },
    { code: '<div scope={`row`} />', errors: [expectedError] },
    { code: '<div scope={`${x}`} />', errors: [expectedError] },

    // ---- TS wrappers around the value — irrelevant to rule ----
    { code: '<div scope={"col" as string} />', errors: [expectedError] },
    { code: '<div scope={("col")} />', errors: [expectedError] },
    { code: '<div scope={"col"!} />', errors: [expectedError] },
    {
      code: '<div scope={"col" satisfies string} />',
      errors: [expectedError],
    },

    // ---- components map flips custom INTO scope-rule coverage ----
    {
      code: '<Cell scope="col" />',
      settings: { 'jsx-a11y': { components: { Cell: 'td' } } },
      errors: [expectedError],
    },
    {
      code: '<Wrapper scope="row" />',
      settings: { 'jsx-a11y': { components: { Wrapper: 'section' } } },
      errors: [expectedError],
    },

    // ---- polymorphicPropName resolving to non-th DOM → reports ----
    {
      code: '<Box as="div" scope />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [expectedError],
    },
    {
      code: '<Box as="span" scope="col" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [expectedError],
    },
    {
      code: '<Box as="div" scope />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: ['Box'],
        },
      },
      errors: [expectedError],
    },

    // ---- TypeScript generic JSX components → reports via components map ----
    {
      code: '<List<string> scope />',
      settings: { 'jsx-a11y': { components: { List: 'div' } } },
      errors: [expectedError],
    },

    // ---- Comments around / inside the prop don't suppress ----
    { code: '<div /* a */ scope /* b */ />', errors: [expectedError] },
    { code: '<div scope={/* row */ "row"} />', errors: [expectedError] },

    // ---- Real-world misuse patterns ----
    {
      code: 'function Header({ title }) { return <h1 scope="col">{title}</h1>; }',
      errors: [expectedError],
    },
    {
      code: '<div role="columnheader" scope="col">Header</div>',
      errors: [expectedError],
    },
    {
      code: 'function Cell({ value }) { return <td scope="col">{value}</td>; }',
      errors: [expectedError],
    },

    // ---- Multiple components in one file ----
    {
      code: 'function A() { return <div scope />; }\nfunction B() { return <span scope />; }',
      errors: [expectedError, expectedError],
    },

    // ---- List-rendered JSX with offending scope ----
    {
      code: 'const items = arr.map((x, i) => <div key={i} scope={x.scope} />);',
      errors: [expectedError],
    },

    // ---- Class component with state-driven JSX ----
    {
      code: 'class T extends React.Component { render() { return this.state.ready ? <div scope="row" /> : null; } }',
      errors: [expectedError],
    },

    // ---- Generator / async / IIFE bodies ----
    {
      code: 'function* render() { yield <div scope />; yield <span scope />; }',
      errors: [expectedError, expectedError],
    },
    {
      code: 'async function render() { return <div scope />; }',
      errors: [expectedError],
    },
    {
      code: 'const x = (() => <div scope />)();',
      errors: [expectedError],
    },

    // ---- Fragment + conditional rendering ----
    {
      code: 'const x = <>{cond && <div scope />}</>;',
      errors: [expectedError],
    },
    {
      code: 'function Foo({a, b}) { return a ? <div scope /> : <span scope />; }',
      errors: [expectedError, expectedError],
    },

    // ---- HOC / forwardRef / memo wrappers carrying scope ----
    {
      code: 'const Enhanced = withTracking(({ value }) => <div value={value} scope />);',
      errors: [expectedError],
    },
    {
      code: 'const FocusInput = React.forwardRef((props, ref) => <div ref={ref} scope {...props} />);',
      errors: [expectedError],
    },
    {
      code: 'const Item = React.memo(({ id }) => <li id={id} scope>{id}</li>);',
      errors: [expectedError],
    },

    // ---- Extended DOM element survey ----
    { code: '<article scope />', errors: [expectedError] },
    { code: '<section scope />', errors: [expectedError] },
    { code: '<aside scope />', errors: [expectedError] },
    { code: '<header scope />', errors: [expectedError] },
    { code: '<footer scope />', errors: [expectedError] },
    { code: '<nav scope />', errors: [expectedError] },
    { code: '<main scope />', errors: [expectedError] },
    { code: '<h1 scope />', errors: [expectedError] },
    { code: '<h6 scope />', errors: [expectedError] },
    { code: '<form scope />', errors: [expectedError] },
    { code: '<fieldset scope />', errors: [expectedError] },
    { code: '<label scope />', errors: [expectedError] },
    { code: '<select scope />', errors: [expectedError] },
    { code: '<option scope />', errors: [expectedError] },
    { code: '<textarea scope />', errors: [expectedError] },
    { code: '<dialog scope />', errors: [expectedError] },
    { code: '<details scope />', errors: [expectedError] },
    { code: '<summary scope />', errors: [expectedError] },
    { code: '<canvas scope />', errors: [expectedError] },
    { code: '<picture scope />', errors: [expectedError] },
    { code: '<video scope />', errors: [expectedError] },
    { code: '<audio scope />', errors: [expectedError] },
    { code: '<iframe scope />', errors: [expectedError] },
    // Table-related neighbors of th — confusable, all illegal.
    { code: '<caption scope />', errors: [expectedError] },
    { code: '<colgroup scope />', errors: [expectedError] },
    { code: '<col scope />', errors: [expectedError] },

    // ---- Spread + scope mixing on non-th (literal scope reports) ----
    { code: '<div {...props} scope />', errors: [expectedError] },
    { code: '<div scope {...props} />', errors: [expectedError] },
    { code: '<div {...a} scope {...b} />', errors: [expectedError] },
    {
      code: "<div {...{scope: 'row'}} scope />",
      errors: [expectedError],
    },

    // ---- Multi-element multi-error scenarios ----
    {
      code: '<><th scope /><div scope /><span scope /></>',
      errors: [expectedError, expectedError],
    },
    {
      code: '<thead scope><tr><th scope="col" /></tr></thead>',
      errors: [expectedError],
    },
    {
      code: '<><div scope /><span scope /><th scope /><a scope /></>',
      errors: [expectedError, expectedError, expectedError],
    },

    // ---- JSX as render prop / children-as-function (offending body) ----
    {
      code: '<DataTable render={() => <div scope />} />',
      errors: [expectedError],
    },
    {
      code: '<DataTable>{() => <div scope />}</DataTable>',
      errors: [expectedError],
    },

    // ---- JSX in array literal / cloneElement / Provider (offending) ----
    {
      code: 'const items = [<div scope key="1" />, <span scope key="2" />, <th scope key="3" />];',
      errors: [expectedError, expectedError],
    },
    { code: 'cloneElement(<div scope />);', errors: [expectedError] },
    {
      code: '<Provider value={data}>{<div scope />}</Provider>',
      errors: [expectedError],
    },
    {
      code: 'const obj = { content: <div scope /> };',
      errors: [expectedError],
    },

    // ---- Whitespace / formatting variations on non-th ----
    { code: '<div  scope  />', errors: [expectedError] },
    { code: '<div scope = "row" />', errors: [expectedError] },
    { code: '<div\n\tscope\n/>', errors: [expectedError] },

    // ---- scope at varied positions ----
    { code: '<div scope id="x" className="y" />', errors: [expectedError] },
    { code: '<div id="x" scope className="y" />', errors: [expectedError] },
    { code: '<div id="x" className="y" scope />', errors: [expectedError] },

    // ---- th sibling with non-th sibling carrying scope ----
    {
      code: '<><th scope="col" /><td scope="row" /></>',
      errors: [expectedError],
    },
    {
      code: '<><th scope /><tr scope /><td scope /></>',
      errors: [expectedError, expectedError],
    },

    // ---- Component library patterns — components map → non-th ----
    {
      code: '<TableCell scope="col" />',
      settings: { 'jsx-a11y': { components: { TableCell: 'td' } } },
      errors: [expectedError],
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
      errors: [expectedError],
    },

    // ---- Suspense / ErrorBoundary / Context wrapping offending JSX ----
    {
      code: '<Suspense fallback={<div scope />}><div>x</div></Suspense>',
      errors: [expectedError],
    },
    {
      code: '<ErrorBoundary><div scope /></ErrorBoundary>',
      errors: [expectedError],
    },
    {
      code: '<Context.Provider value={data}><div scope /></Context.Provider>',
      errors: [expectedError],
    },

    // ---- TS-only JSX shapes ----
    {
      code: 'const x = <div scope /> as React.ReactNode;',
      errors: [expectedError],
    },
    {
      code: 'const x = (<div scope />) satisfies any;',
      errors: [expectedError],
    },
    { code: 'const x = (<div scope />)!;', errors: [expectedError] },

    // ---- TypeScript generic JSX with components remap ----
    {
      code: '<List<Header> scope />',
      settings: { 'jsx-a11y': { components: { List: 'div' } } },
      errors: [expectedError],
    },

    // ---- Realistic table cell misuse patterns ----
    {
      code: '<table><tr><td scope="col" /><td scope="col" /></tr><tr><td scope="row" /><td scope="row" /></tr></table>',
      errors: [expectedError, expectedError, expectedError, expectedError],
    },
    {
      code: '<table><tr><th scope="col" /><td scope="col" /><th scope="col" /><td scope="col" /></tr></table>',
      errors: [expectedError, expectedError],
    },

    // ---- try / catch / finally JSX ----
    {
      code: 'function App() { try { return <div scope />; } catch (e) { return <span scope />; } }',
      errors: [expectedError, expectedError],
    },

    // ---- switch returning JSX ----
    {
      code: "function App({type}) { switch(type) { case 'a': return <div scope />; case 'b': return <span scope />; default: return null; } }",
      errors: [expectedError, expectedError],
    },

    // ---- Class with class-field arrow returning JSX ----
    {
      code: 'class C { render = () => <div scope />; }',
      errors: [expectedError],
    },

    // ---- JSX in attribute value triggers nested + outer reports ----
    {
      code: '<div header={<span scope />} scope />',
      errors: [expectedError, expectedError],
    },

    // ---- Web Components via components map ----
    {
      code: '<my-table scope />',
      settings: { 'jsx-a11y': { components: { 'my-table': 'td' } } },
      errors: [expectedError],
    },

    // ---- Components map with dotted key remapping member-expression tag to non-th DOM ----
    {
      code: '<DataGrid.Header scope="col" />',
      settings: {
        'jsx-a11y': { components: { 'DataGrid.Header': 'div' } },
      },
      errors: [expectedError],
    },
    // ---- Namespaced tag remapped via composite key ----
    {
      code: '<svg:circle scope />',
      settings: {
        'jsx-a11y': { components: { 'svg:circle': 'div' } },
      },
      errors: [expectedError],
    },

    // ---- Polymorphic prop value forms ----
    {
      code: '<Box as={"div"} scope />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [expectedError],
    },
    {
      code: '<Box as={`div`} scope />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [expectedError],
    },

    // ---- Defensive: invalid settings types should still fall back ----
    {
      code: '<div scope />',
      settings: {
        'jsx-a11y': { components: 'invalid' as unknown as object },
      },
      errors: [expectedError],
    },
    {
      code: '<div as="th" scope />',
      settings: {
        'jsx-a11y': { polymorphicPropName: 123 as unknown as string },
      },
      errors: [expectedError],
    },
  ],
});
