import { RuleTester } from '../rule-tester';

const errorMessage =
  'Visible, non-interactive elements with click handlers must have at least one keyboard listener.';
const expectedError = { message: errorMessage };

const footerComponentSettings = {
  'jsx-a11y': {
    components: {
      Footer: 'footer',
    },
  },
};

const polymorphicSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
  },
};

const polymorphicAllowListSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
    polymorphicAllowList: ['Foo'],
  },
};

new RuleTester().run('click-events-have-key-events', null as never, {
  valid: [
    // ============================================================
    // Upstream valid suite (mirrors __tests__/src/rules/click-events-have-key-events-test.js)
    // ============================================================
    { code: '<div onClick={() => void 0} onKeyDown={foo}/>;' },
    { code: '<div onClick={() => void 0} onKeyUp={foo} />;' },
    { code: '<div onClick={() => void 0} onKeyPress={foo}/>;' },
    { code: '<div onClick={() => void 0} onKeyDown={foo} onKeyUp={bar} />;' },
    { code: '<div onClick={() => void 0} onKeyDown={foo} {...props} />;' },
    { code: '<div className="foo" />;' },
    { code: '<div onClick={() => void 0} aria-hidden />;' },
    { code: '<div onClick={() => void 0} aria-hidden={true} />;' },
    {
      code: '<div onClick={() => void 0} aria-hidden={false} onKeyDown={foo} />;',
    },
    {
      code: '<div onClick={() => void 0} onKeyDown={foo} aria-hidden={undefined} />;',
    },
    { code: '<input type="text" onClick={() => void 0} />' },
    { code: '<input onClick={() => void 0} />' },
    { code: '<button onClick={() => void 0} className="foo" />' },
    { code: '<option onClick={() => void 0} className="foo" />' },
    { code: '<select onClick={() => void 0} className="foo" />' },
    { code: '<textarea onClick={() => void 0} className="foo" />' },
    { code: '<a onClick={() => void 0} href="http://x.y.z" />' },
    { code: '<a onClick={() => void 0} href="http://x.y.z" tabIndex="0" />' },
    { code: '<input onClick={() => void 0} type="hidden" />;' },
    { code: '<div onClick={() => void 0} role="presentation" />;' },
    { code: '<div onClick={() => void 0} role="none" />;' },
    { code: '<TestComponent onClick={doFoo} />' },
    { code: '<Button onClick={doFoo} />' },
    { code: '<Footer onClick={doFoo} />' },

    // ============================================================
    // Inherent-interactive element survey (no onClick exemption needed
    // because each is interactive by its role schema).
    // ============================================================
    { code: '<a href="x" onClick={() => void 0} />' },
    { code: '<input type="button" onClick={() => void 0} />' },
    { code: '<input type="submit" onClick={() => void 0} />' },
    { code: '<input type="reset" onClick={() => void 0} />' },
    { code: '<input type="image" onClick={() => void 0} />' },
    { code: '<input type="checkbox" onClick={() => void 0} />' },
    { code: '<input type="radio" onClick={() => void 0} />' },
    { code: '<input type="range" onClick={() => void 0} />' },
    { code: '<input type="number" onClick={() => void 0} />' },
    { code: '<th onClick={() => void 0} />' },
    { code: '<td onClick={() => void 0} />' },
    { code: '<tr onClick={() => void 0} />' },
    { code: '<datalist onClick={() => void 0} />' },
    { code: '<menuitem onClick={() => void 0} />' },
    { code: '<summary onClick={() => void 0} />' },

    // ============================================================
    // aria-hidden additional shapes
    // ============================================================
    { code: '<div onClick={() => void 0} aria-hidden="true" />;' },
    { code: '<input onClick={() => void 0} type="HIDDEN" />;' },

    // ============================================================
    // Settings: polymorphicPropName / polymorphicAllowList
    // ============================================================
    {
      code: '<Foo as="button" onClick={() => void 0} />;',
      settings: polymorphicSettings,
    },
    {
      code: '<Foo as={someType} onClick={() => void 0} />;',
      settings: polymorphicSettings,
    },
    {
      code: '<Bar as="div" onClick={() => void 0} />;',
      settings: polymorphicAllowListSettings,
    },

    // ============================================================
    // Spread + direct keyboard listener combo
    // ============================================================
    {
      code: '<div onClick={() => void 0} onKeyDown={foo} {...{onKeyUp: bar}} />;',
    },

    // ============================================================
    // Nested JSX: interactive button child inside valid outer container
    // ============================================================
    {
      code: '<div onClick={() => void 0} onKeyDown={foo}><button onClick={() => void 0}>x</button></div>',
    },

    // ============================================================
    // Real-world component patterns
    // ============================================================
    {
      code: 'function Button({ children }) { return <button onClick={() => void 0}>{children}</button>; }',
    },
    {
      code: 'const items = arr.map(item => <button key={item.id} onClick={() => void 0}>{item.name}</button>);',
    },
    {
      code: 'const Fancy = React.forwardRef((props, ref) => <button ref={ref} onClick={() => void 0} {...props} />);',
    },
  ],
  invalid: [
    // ============================================================
    // Upstream invalid suite
    // ============================================================
    {
      code: '<div onClick={() => void 0} />;',
      errors: [expectedError],
    },
    {
      code: '<div onClick={() => void 0} role={undefined} />;',
      errors: [expectedError],
    },
    {
      code: '<div onClick={() => void 0} {...props} />;',
      errors: [expectedError],
    },
    { code: '<section onClick={() => void 0} />;', errors: [expectedError] },
    { code: '<main onClick={() => void 0} />;', errors: [expectedError] },
    { code: '<article onClick={() => void 0} />;', errors: [expectedError] },
    { code: '<header onClick={() => void 0} />;', errors: [expectedError] },
    { code: '<footer onClick={() => void 0} />;', errors: [expectedError] },
    {
      code: '<div onClick={() => void 0} aria-hidden={false} />;',
      errors: [expectedError],
    },
    { code: '<a onClick={() => void 0} />', errors: [expectedError] },
    {
      code: '<a tabIndex="0" onClick={() => void 0} />',
      errors: [expectedError],
    },
    {
      code: '<Footer onClick={doFoo} />',
      errors: [expectedError],
      settings: footerComponentSettings,
    },

    // ============================================================
    // Non-interactive element survey
    // ============================================================
    { code: '<span onClick={() => void 0} />', errors: [expectedError] },
    { code: '<aside onClick={() => void 0} />', errors: [expectedError] },
    { code: '<nav onClick={() => void 0} />', errors: [expectedError] },
    { code: '<p onClick={() => void 0} />', errors: [expectedError] },
    { code: '<h1 onClick={() => void 0} />', errors: [expectedError] },
    { code: '<h6 onClick={() => void 0} />', errors: [expectedError] },
    { code: '<ul onClick={() => void 0} />', errors: [expectedError] },
    { code: '<ol onClick={() => void 0} />', errors: [expectedError] },
    { code: '<li onClick={() => void 0} />', errors: [expectedError] },

    // ============================================================
    // Listener boundary — nested element reports independently
    // ============================================================
    {
      code: '<button onClick={() => void 0}><div onClick={() => void 0} /></button>',
      errors: [expectedError],
    },

    // ============================================================
    // Spread literal — hasAnyProp's spreadStrict: true default keeps
    // it opaque.
    // ============================================================
    {
      code: '<div onClick={() => void 0} {...{onKeyDown: foo}} />;',
      errors: [expectedError],
    },
    {
      code: '<div {...props} onClick={() => void 0} />;',
      errors: [expectedError],
    },

    // ============================================================
    // polymorphicPropName → non-interactive
    // ============================================================
    {
      code: '<Foo as="div" onClick={() => void 0} />;',
      settings: polymorphicSettings,
      errors: [expectedError],
    },

    // ============================================================
    // Real-world a11y misuse
    // ============================================================
    {
      code: 'function ClickCard() { return <div onClick={handler}>Title</div>; }',
      errors: [expectedError],
    },
    {
      code: 'function App() { return (<><div onClick={() => void 0}>A</div><button onClick={() => void 0}>B</button><section onClick={() => void 0}>C</section></>); }',
      errors: [expectedError, expectedError],
    },
  ],
});
