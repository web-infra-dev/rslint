import { RuleTester } from '../rule-tester';

const errorMessage =
  'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.';
const expectedError = { message: errorMessage };

const componentsSettings = {
  'jsx-a11y': {
    components: {
      Button: 'button',
      TestComponent: 'div',
    },
  },
};

const recommendedOptions = [
  {
    handlers: [
      'onClick',
      'onMouseDown',
      'onMouseUp',
      'onKeyPress',
      'onKeyDown',
      'onKeyUp',
    ],
    allowExpressionValues: true,
  },
];

const allowExpressionValuesTrueOptions = [{ allowExpressionValues: true }];
const allowExpressionValuesFalseOptions = [{ allowExpressionValues: false }];

new RuleTester().run('no-static-element-interactions', null as never, {
  valid: [
    // ============================================================
    // Upstream alwaysValid (run under both :strict and :recommended)
    // ============================================================
    // Custom JSX components — not in dom set → exempt.
    { code: '<TestComponent onClick={doFoo} />' },
    { code: '<Button onClick={doFoo} />' },
    { code: '<Button onClick={doFoo} />', settings: componentsSettings },
    // No interactive handlers.
    { code: '<div />;' },
    { code: '<div className="foo" />;' },
    { code: '<div className="foo" {...props} />;' },
    // aria-hidden short-circuit.
    { code: '<div onClick={() => void 0} aria-hidden />;' },
    { code: '<div onClick={() => void 0} aria-hidden={true} />;' },
    // null handler — getPropValue returns null → `!= null` is false.
    { code: '<div onClick={null} />;' },

    // Inherently interactive HTML elements.
    { code: '<input onClick={() => void 0} />' },
    { code: '<input type="button" onClick={() => void 0} />' },
    { code: '<input type="checkbox" onClick={() => void 0} />' },
    { code: '<input type="text" onClick={() => void 0} />' },
    { code: '<input type="hidden" onClick={() => void 0} />' },
    { code: '<button onClick={() => void 0} className="foo" />' },
    { code: '<select onClick={() => void 0} className="foo" />' },
    { code: '<textarea onClick={() => void 0} className="foo" />' },
    { code: '<a onClick={() => void 0} href="http://x.y.z" />' },
    { code: '<a onClick={() => void 0} href="http://x.y.z" tabIndex="0" />' },

    // Interactive role attribute on <div>.
    { code: '<div role="button" onClick={() => {}} />;' },
    { code: '<div role="checkbox" onClick={() => {}} />;' },
    { code: '<div role="link" onClick={() => {}} />;' },
    { code: '<div role="menuitem" onClick={() => {}} />;' },
    { code: '<div role="tab" onClick={() => {}} />;' },
    { code: '<div role="textbox" onClick={() => {}} />;' },
    { code: '<div role="treeitem" onClick={() => {}} />;' },

    // Presentation roles.
    { code: '<div role="presentation" onClick={() => {}} />;' },
    { code: '<div role="presentation" onKeyDown={() => {}} />;' },

    // HTML elements with inherent non-interactive role.
    { code: '<article onClick={() => {}} />;' },
    { code: '<aside onClick={() => {}} />;' },
    { code: '<blockquote onClick={() => {}} />;' },
    { code: '<h1 onClick={() => {}} />;' },
    { code: '<li onClick={() => {}} />;' },
    { code: '<main onClick={() => void 0} />;' },
    { code: '<nav onClick={() => {}} />;' },
    { code: '<p onClick={() => {}} />;' },
    { code: '<table onClick={() => {}} />;' },
    { code: '<section onClick={() => {}} aria-label="Aa" />;' },

    // Abstract roles.
    { code: '<div role="command" onClick={() => {}} />;' },
    { code: '<div role="widget" onClick={() => {}} />;' },
    { code: '<div role="window" onClick={() => {}} />;' },
    { code: '<div role="composite" onClick={() => {}} />;' },

    // Non-interactive role attribute on <div>.
    { code: '<div role="alert" onClick={() => {}} />;' },
    { code: '<div role="article" onClick={() => {}} />;' },
    { code: '<div role="dialog" onClick={() => {}} />;' },
    { code: '<div role="document" onClick={() => {}} />;' },
    { code: '<div role="banner" onClick={() => {}} />;' },
    { code: '<div role="form" onClick={() => {}} />;' },

    // Non-triggering handlers (clipboard / composition / change / touch /
    // scroll / wheel / media / animation) — none in any handler list.
    { code: '<div onCopy={() => {}} />;' },
    { code: '<div onCut={() => {}} />;' },
    { code: '<div onPaste={() => {}} />;' },
    { code: '<div onChange={() => {}} />;' },
    { code: '<div onSubmit={() => {}} />;' },
    { code: '<div onScroll={() => {}} />;' },
    { code: '<div onTouchStart={() => {}} />;' },
    { code: '<div onAnimationStart={() => {}} />;' },

    // ============================================================
    // recommended-only valid (narrowed handlers + allowExpressionValues=true)
    // ============================================================
    // Handlers in the default list but NOT in recommended.handlers.
    { code: '<div onFocus={() => {}} />;', options: recommendedOptions },
    { code: '<div onBlur={() => {}} />;', options: recommendedOptions },
    { code: '<div onContextMenu={() => {}} />;', options: recommendedOptions },
    { code: '<div onDblClick={() => {}} />;', options: recommendedOptions },
    { code: '<div onDrag={() => {}} />;', options: recommendedOptions },
    { code: '<div onMouseEnter={() => {}} />;', options: recommendedOptions },
    { code: '<div onMouseLeave={() => {}} />;', options: recommendedOptions },

    // allowExpressionValues=true — non-literal role exempt.
    {
      code: '<div role={ROLE_BUTTON} onClick={() => {}} />;',
      options: recommendedOptions,
    },
    {
      code: '<div  {...this.props} role={this.props.role} onKeyPress={e => this.handleKeyPress(e)}>{this.props.children}</div>',
      options: recommendedOptions,
    },
    {
      code: '<div role={BUTTON} onClick={() => {}} />;',
      options: allowExpressionValuesTrueOptions,
    },
    {
      code: '<div role={isButton ? "button" : "link"} onClick={() => {}} />;',
      options: allowExpressionValuesTrueOptions,
    },
    {
      code: '<div role={isButton ? "button" : LINK} onClick={() => {}} />;',
      options: allowExpressionValuesTrueOptions,
    },
    {
      code: '<div role={isButton ? BUTTON : LINK} onClick={() => {}} />;',
      options: allowExpressionValuesTrueOptions,
    },
  ],

  invalid: [
    // ============================================================
    // Upstream neverValid (under both :strict and :recommended)
    // ============================================================
    {
      code: '<div onClick={() => void 0} />;',
      errors: [expectedError],
    },
    {
      code: '<div onClick={() => void 0} />;',
      options: recommendedOptions,
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
    {
      code: '<div onKeyUp={() => void 0} aria-hidden={false} />;',
      errors: [expectedError],
    },

    // Static elements without inherent role.
    { code: '<a onClick={() => {}} />;', errors: [expectedError] },
    {
      code: '<a tabIndex="0" onClick={() => void 0} />',
      errors: [expectedError],
    },
    { code: '<acronym onClick={() => {}} />;', errors: [expectedError] },
    { code: '<b onClick={() => {}} />;', errors: [expectedError] },
    { code: '<bdi onClick={() => {}} />;', errors: [expectedError] },
    { code: '<blink onClick={() => {}} />;', errors: [expectedError] },
    { code: '<body onClick={() => {}} />;', errors: [expectedError] },
    { code: '<center onClick={() => {}} />;', errors: [expectedError] },
    { code: '<cite onClick={() => {}} />;', errors: [expectedError] },
    { code: '<colgroup onClick={() => {}} />;', errors: [expectedError] },
    { code: '<font onClick={() => {}} />;', errors: [expectedError] },
    { code: '<head onClick={() => {}} />;', errors: [expectedError] },
    { code: '<header onClick={() => {}} />;', errors: [expectedError] },
    { code: '<hgroup onClick={() => {}} />;', errors: [expectedError] },
    { code: '<i onClick={() => {}} />;', errors: [expectedError] },
    { code: '<kbd onClick={() => {}} />;', errors: [expectedError] },
    { code: '<map onClick={() => {}} />;', errors: [expectedError] },
    { code: '<meta onClick={() => {}} />;', errors: [expectedError] },
    { code: '<noscript onClick={() => {}} />;', errors: [expectedError] },
    { code: '<object onClick={() => {}} />;', errors: [expectedError] },
    { code: '<picture onClick={() => {}} />;', errors: [expectedError] },
    { code: '<q onClick={() => {}} />;', errors: [expectedError] },
    { code: '<rp onClick={() => {}} />;', errors: [expectedError] },
    { code: '<s onClick={() => {}} />;', errors: [expectedError] },
    { code: '<samp onClick={() => {}} />;', errors: [expectedError] },
    { code: '<script onClick={() => {}} />;', errors: [expectedError] },
    { code: '<section onClick={() => {}} />;', errors: [expectedError] },
    { code: '<small onClick={() => {}} />;', errors: [expectedError] },
    { code: '<span onClick={() => {}} />;', errors: [expectedError] },
    { code: '<spacer onClick={() => {}} />;', errors: [expectedError] },
    { code: '<style onClick={() => {}} />;', errors: [expectedError] },
    { code: '<title onClick={() => {}} />;', errors: [expectedError] },
    { code: '<u onClick={() => {}} />;', errors: [expectedError] },
    { code: '<var onClick={() => {}} />;', errors: [expectedError] },
    { code: '<wbr onClick={() => {}} />;', errors: [expectedError] },

    // Keyboard / mouse handler variants present in BOTH lists.
    { code: '<div onKeyDown={() => {}} />;', errors: [expectedError] },
    { code: '<div onKeyPress={() => {}} />;', errors: [expectedError] },
    { code: '<div onKeyUp={() => {}} />;', errors: [expectedError] },
    { code: '<div onClick={() => {}} />;', errors: [expectedError] },
    { code: '<div onMouseDown={() => {}} />;', errors: [expectedError] },
    { code: '<div onMouseUp={() => {}} />;', errors: [expectedError] },
    // Custom component remapped to <div>.
    {
      code: '<TestComponent onClick={doFoo} />',
      settings: componentsSettings,
      errors: [expectedError],
    },

    // ============================================================
    // strict-only invalid (defaults — full focus+keyboard+mouse handlers)
    // ============================================================
    { code: '<div onContextMenu={() => {}} />;', errors: [expectedError] },
    { code: '<div onDblClick={() => {}} />;', errors: [expectedError] },
    { code: '<div onDoubleClick={() => {}} />;', errors: [expectedError] },
    { code: '<div onDrag={() => {}} />;', errors: [expectedError] },
    { code: '<div onDragEnd={() => {}} />;', errors: [expectedError] },
    { code: '<div onDragEnter={() => {}} />;', errors: [expectedError] },
    { code: '<div onDragExit={() => {}} />;', errors: [expectedError] },
    { code: '<div onDragLeave={() => {}} />;', errors: [expectedError] },
    { code: '<div onDragOver={() => {}} />;', errors: [expectedError] },
    { code: '<div onDragStart={() => {}} />;', errors: [expectedError] },
    { code: '<div onDrop={() => {}} />;', errors: [expectedError] },
    { code: '<div onMouseEnter={() => {}} />;', errors: [expectedError] },
    { code: '<div onMouseLeave={() => {}} />;', errors: [expectedError] },
    { code: '<div onMouseMove={() => {}} />;', errors: [expectedError] },
    { code: '<div onMouseOut={() => {}} />;', errors: [expectedError] },
    { code: '<div onMouseOver={() => {}} />;', errors: [expectedError] },

    // Non-literal `role` without allowExpressionValues=true.
    {
      code: '<div role={ROLE_BUTTON} onClick={() => {}} />;',
      errors: [expectedError],
    },
    {
      code: '<div role={BUTTON} onClick={() => {}} />;',
      options: allowExpressionValuesFalseOptions,
      errors: [expectedError],
    },
    {
      code: '<div role={isButton ? "button" : "link"} onClick={() => {}} />;',
      options: allowExpressionValuesFalseOptions,
      errors: [expectedError],
    },
  ],
});
