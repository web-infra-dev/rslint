import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';
import { singleLineError } from '../matrix';

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

new RuleTester(config).run('no-static-element-interactions', {
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
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div onClick={() => void 0} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div onClick={() => void 0} role={undefined} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '<div onClick={() => void 0} {...props} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '<div onKeyUp={() => void 0} aria-hidden={false} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },

    // Static elements without inherent role.
    {
      code: '<a onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<a tabIndex="0" onClick={() => void 0} />',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '<acronym onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<b onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<bdi onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<blink onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<body onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<center onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<cite onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<colgroup onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<font onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<head onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<header onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<hgroup onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<i onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<kbd onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<map onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<meta onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<noscript onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<object onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<picture onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<q onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<rp onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<s onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<samp onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<script onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<section onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<small onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<span onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<spacer onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<style onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<title onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<u onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<var onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<wbr onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },

    // Keyboard / mouse handler variants present in BOTH lists.
    {
      code: '<div onKeyDown={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div onKeyPress={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div onKeyUp={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div onMouseDown={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div onMouseUp={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    // Custom component remapped to <div>.
    {
      code: '<TestComponent onClick={doFoo} />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },

    // ============================================================
    // strict-only invalid (defaults — full focus+keyboard+mouse handlers)
    // ============================================================
    {
      code: '<div onContextMenu={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<div onDblClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div onDoubleClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<div onDrag={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div onDragEnd={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div onDragEnter={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div onDragExit={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div onDragLeave={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div onDragOver={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div onDragStart={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div onDrop={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div onMouseEnter={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<div onMouseLeave={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<div onMouseMove={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div onMouseOut={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div onMouseOver={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // Non-literal `role` without allowExpressionValues=true.
    {
      code: '<div role={ROLE_BUTTON} onClick={() => {}} />;',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: '<div role={BUTTON} onClick={() => {}} />;',
      options: allowExpressionValuesFalseOptions,
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: '<div role={isButton ? "button" : "link"} onClick={() => {}} />;',
      options: allowExpressionValuesFalseOptions,
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 63,
        },
      ],
    },
  ],
});

// Additional upstream v6.10.2 inputs. Matrix rows freeze the expected
// classification and configuration; ranges follow the authored source templates.
// No rule implementation or lint output is consulted when generating cases.
const upstreamOptions0 = [
  {
    allowExpressionValues: true,
    handlers: [
      'onClick',
      'onMouseDown',
      'onMouseUp',
      'onKeyPress',
      'onKeyDown',
      'onKeyUp',
    ],
  },
];
const upstreamSettings0 = {
  'jsx-a11y': { components: { Button: 'button', TestComponent: 'div' } },
};

new RuleTester(config).run('no-static-element-interactions', {
  valid: [
    ...(
      [
        [
          'onClick',
          '() => void 0',
          '',
          12,
          'color date datetime datetime-local email file',
        ],
        [
          'onClick',
          '() => void 0',
          '',
          19,
          'image month number password radio range reset search submit tel',
        ],
        ['onClick', '() => void 0', '', 30, 'time url week'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:strict valid ${firstCase + index}`,
          code,
        };
      }),
    ),
    ...(
      [['onClick', '() => {}', ';', 34, 'datalist menuitem']] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:strict valid ${firstCase + index}`,
          code,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 36',
      code: '<option onClick={() => void 0} className="foo" />',
    },
    ...(
      [
        ['onClick', '() => {}', ';', 41, 'audio form'],
        ['onSubmit', '() => {}', ';', 43, 'form'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:strict valid ${firstCase + index}`,
          code,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 46',
      code: '<div role="columnheader" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 47',
      code: '<div role="combobox" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 49',
      code: '<div role="gridcell" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 52',
      code: '<div role="menuitemcheckbox" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 53',
      code: '<div role="menuitemradio" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 54',
      code: '<div role="option" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 55',
      code: '<div role="radio" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 56',
      code: '<div role="rowheader" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 57',
      code: '<div role="searchbox" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 58',
      code: '<div role="slider" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 59',
      code: '<div role="spinbutton" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 60',
      code: '<div role="switch" onClick={() => {}} />;',
    },
    ...(
      [
        ['onClick', '() => {}', ';', 66, 'address'],
        ['onDblClick', '() => void 0', ';', 68, 'article'],
        [
          'onClick',
          '() => {}',
          ';',
          71,
          'br canvas caption code dd del details dfn dir dl dt em embed fieldset figcaption figure footer',
        ],
        [
          'onClick',
          '() => {}',
          ';',
          89,
          'h2 h3 h4 h5 h6 hr html iframe img ins label legend',
        ],
        ['onClick', '() => {}', ';', 103, 'mark marquee menu meter'],
        ['onClick', '() => {}', ';', 108, 'ol optgroup output'],
        ['onClick', '() => {}', ';', 112, 'pre progress ruby'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:strict valid ${firstCase + index}`,
          code,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 116',
      code: '<section onClick={() => {}} aria-labelledby="js_1" />;',
    },
    ...(
      [
        ['onClick', '() => {}', ';', 117, 'strong sub summary sup'],
        [
          'onClick',
          '() => {}',
          ';',
          122,
          'tbody tfoot th thead time tr video ul',
        ],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:strict valid ${firstCase + index}`,
          code,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 132',
      code: '<div role="input" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 133',
      code: '<div role="landmark" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 134',
      code: '<div role="range" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 135',
      code: '<div role="roletype" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 136',
      code: '<div role="sectionhead" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 137',
      code: '<div role="select" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 138',
      code: '<div role="structure" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 142',
      code: '<div role="alertdialog" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 143',
      code: '<div role="application" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 146',
      code: '<div role="cell" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 147',
      code: '<div role="complementary" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 148',
      code: '<div role="contentinfo" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 149',
      code: '<div role="definition" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 151',
      code: '<div role="directory" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 153',
      code: '<div role="feed" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 154',
      code: '<div role="figure" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 155',
      code: '<div role="grid" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 156',
      code: '<div role="group" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 157',
      code: '<div role="heading" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 158',
      code: '<div role="img" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 159',
      code: '<div role="list" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 160',
      code: '<div role="listbox" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 161',
      code: '<div role="listitem" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 162',
      code: '<div role="log" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 163',
      code: '<div role="main" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 164',
      code: '<div role="marquee" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 165',
      code: '<div role="math" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 166',
      code: '<div role="menu" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 167',
      code: '<div role="menubar" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 168',
      code: '<div role="navigation" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 169',
      code: '<div role="note" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 170',
      code: '<div role="progressbar" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 171',
      code: '<div role="radiogroup" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 172',
      code: '<div role="region" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 173',
      code: '<div role="row" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 174',
      code: '<div role="rowgroup" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 175',
      code: '<div role="section" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 176',
      code: '<div role="search" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 177',
      code: '<div role="separator" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 178',
      code: '<div role="scrollbar" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 179',
      code: '<div role="status" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 180',
      code: '<div role="table" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 181',
      code: '<div role="tablist" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 182',
      code: '<div role="tabpanel" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 183',
      code: '<div role="term" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 184',
      code: '<div role="timer" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 185',
      code: '<div role="toolbar" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 186',
      code: '<div role="tooltip" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 187',
      code: '<div role="tree" onClick={() => {}} />;',
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict valid 188',
      code: '<div role="treegrid" onClick={() => {}} />;',
    },
    ...(
      [
        ['onCompositionEnd', '() => {}', ';', 192, 'div'],
        ['onCompositionStart', '() => {}', ';', 193, 'div'],
        ['onCompositionUpdate', '() => {}', ';', 194, 'div'],
        ['onInput', '() => {}', ';', 196, 'div'],
        ['onSelect', '() => {}', ';', 198, 'div'],
        ['onTouchCancel', '() => {}', ';', 199, 'div'],
        ['onTouchEnd', '() => {}', ';', 200, 'div'],
        ['onTouchMove', '() => {}', ';', 201, 'div'],
        ['onWheel', '() => {}', ';', 204, 'div'],
        ['onAbort', '() => {}', ';', 205, 'div'],
        ['onCanPlay', '() => {}', ';', 206, 'div'],
        ['onCanPlayThrough', '() => {}', ';', 207, 'div'],
        ['onDurationChange', '() => {}', ';', 208, 'div'],
        ['onEmptied', '() => {}', ';', 209, 'div'],
        ['onEncrypted', '() => {}', ';', 210, 'div'],
        ['onEnded', '() => {}', ';', 211, 'div'],
        ['onError', '() => {}', ';', 212, 'div'],
        ['onLoadedData', '() => {}', ';', 213, 'div'],
        ['onLoadedMetadata', '() => {}', ';', 214, 'div'],
        ['onLoadStart', '() => {}', ';', 215, 'div'],
        ['onPause', '() => {}', ';', 216, 'div'],
        ['onPlay', '() => {}', ';', 217, 'div'],
        ['onPlaying', '() => {}', ';', 218, 'div'],
        ['onProgress', '() => {}', ';', 219, 'div'],
        ['onRateChange', '() => {}', ';', 220, 'div'],
        ['onSeeked', '() => {}', ';', 221, 'div'],
        ['onSeeking', '() => {}', ';', 222, 'div'],
        ['onStalled', '() => {}', ';', 223, 'div'],
        ['onSuspend', '() => {}', ';', 224, 'div'],
        ['onTimeUpdate', '() => {}', ';', 225, 'div'],
        ['onVolumeChange', '() => {}', ';', 226, 'div'],
        ['onWaiting', '() => {}', ';', 227, 'div'],
        ['onLoad', '() => {}', ';', 228, 'div'],
        ['onAnimationEnd', '() => {}', ';', 231, 'div'],
        ['onAnimationIteration', '() => {}', ';', 232, 'div'],
        ['onTransitionEnd', '() => {}', ';', 233, 'div'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:strict valid ${firstCase + index}`,
          code,
        };
      }),
    ),
    ...(
      [
        [
          'onClick',
          '() => void 0',
          '',
          12,
          'color date datetime datetime-local email file',
        ],
        [
          'onClick',
          '() => void 0',
          '',
          19,
          'image month number password radio range reset search submit tel',
        ],
        ['onClick', '() => void 0', '', 30, 'time url week'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    ...(
      [['onClick', '() => {}', ';', 34, 'datalist menuitem']] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 36',
      code: '<option onClick={() => void 0} className="foo" />',
      options: upstreamOptions0,
    },
    ...(
      [
        ['onClick', '() => {}', ';', 41, 'audio form'],
        ['onSubmit', '() => {}', ';', 43, 'form'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 46',
      code: '<div role="columnheader" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 47',
      code: '<div role="combobox" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 49',
      code: '<div role="gridcell" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 52',
      code: '<div role="menuitemcheckbox" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 53',
      code: '<div role="menuitemradio" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 54',
      code: '<div role="option" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 55',
      code: '<div role="radio" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 56',
      code: '<div role="rowheader" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 57',
      code: '<div role="searchbox" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 58',
      code: '<div role="slider" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 59',
      code: '<div role="spinbutton" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 60',
      code: '<div role="switch" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    ...(
      [
        ['onClick', '() => {}', ';', 66, 'address'],
        ['onDblClick', '() => void 0', ';', 68, 'article'],
        [
          'onClick',
          '() => {}',
          ';',
          71,
          'br canvas caption code dd del details dfn dir dl dt em embed fieldset figcaption figure footer',
        ],
        [
          'onClick',
          '() => {}',
          ';',
          89,
          'h2 h3 h4 h5 h6 hr html iframe img ins label legend',
        ],
        ['onClick', '() => {}', ';', 103, 'mark marquee menu meter'],
        ['onClick', '() => {}', ';', 108, 'ol optgroup output'],
        ['onClick', '() => {}', ';', 112, 'pre progress ruby'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 116',
      code: '<section onClick={() => {}} aria-labelledby="js_1" />;',
      options: upstreamOptions0,
    },
    ...(
      [
        ['onClick', '() => {}', ';', 117, 'strong sub summary sup'],
        [
          'onClick',
          '() => {}',
          ';',
          122,
          'tbody tfoot th thead time tr video ul',
        ],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 132',
      code: '<div role="input" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 133',
      code: '<div role="landmark" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 134',
      code: '<div role="range" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 135',
      code: '<div role="roletype" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 136',
      code: '<div role="sectionhead" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 137',
      code: '<div role="select" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 138',
      code: '<div role="structure" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 142',
      code: '<div role="alertdialog" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 143',
      code: '<div role="application" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 146',
      code: '<div role="cell" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 147',
      code: '<div role="complementary" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 148',
      code: '<div role="contentinfo" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 149',
      code: '<div role="definition" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 151',
      code: '<div role="directory" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 153',
      code: '<div role="feed" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 154',
      code: '<div role="figure" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 155',
      code: '<div role="grid" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 156',
      code: '<div role="group" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 157',
      code: '<div role="heading" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 158',
      code: '<div role="img" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 159',
      code: '<div role="list" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 160',
      code: '<div role="listbox" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 161',
      code: '<div role="listitem" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 162',
      code: '<div role="log" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 163',
      code: '<div role="main" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 164',
      code: '<div role="marquee" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 165',
      code: '<div role="math" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 166',
      code: '<div role="menu" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 167',
      code: '<div role="menubar" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 168',
      code: '<div role="navigation" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 169',
      code: '<div role="note" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 170',
      code: '<div role="progressbar" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 171',
      code: '<div role="radiogroup" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 172',
      code: '<div role="region" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 173',
      code: '<div role="row" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 174',
      code: '<div role="rowgroup" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 175',
      code: '<div role="section" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 176',
      code: '<div role="search" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 177',
      code: '<div role="separator" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 178',
      code: '<div role="scrollbar" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 179',
      code: '<div role="status" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 180',
      code: '<div role="table" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 181',
      code: '<div role="tablist" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 182',
      code: '<div role="tabpanel" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 183',
      code: '<div role="term" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 184',
      code: '<div role="timer" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 185',
      code: '<div role="toolbar" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 186',
      code: '<div role="tooltip" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 187',
      code: '<div role="tree" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 188',
      code: '<div role="treegrid" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    ...(
      [
        ['onCompositionEnd', '() => {}', ';', 192, 'div'],
        ['onCompositionStart', '() => {}', ';', 193, 'div'],
        ['onCompositionUpdate', '() => {}', ';', 194, 'div'],
        ['onInput', '() => {}', ';', 196, 'div'],
        ['onSelect', '() => {}', ';', 198, 'div'],
        ['onTouchCancel', '() => {}', ';', 199, 'div'],
        ['onTouchEnd', '() => {}', ';', 200, 'div'],
        ['onTouchMove', '() => {}', ';', 201, 'div'],
        ['onWheel', '() => {}', ';', 204, 'div'],
        ['onAbort', '() => {}', ';', 205, 'div'],
        ['onCanPlay', '() => {}', ';', 206, 'div'],
        ['onCanPlayThrough', '() => {}', ';', 207, 'div'],
        ['onDurationChange', '() => {}', ';', 208, 'div'],
        ['onEmptied', '() => {}', ';', 209, 'div'],
        ['onEncrypted', '() => {}', ';', 210, 'div'],
        ['onEnded', '() => {}', ';', 211, 'div'],
        ['onError', '() => {}', ';', 212, 'div'],
        ['onLoadedData', '() => {}', ';', 213, 'div'],
        ['onLoadedMetadata', '() => {}', ';', 214, 'div'],
        ['onLoadStart', '() => {}', ';', 215, 'div'],
        ['onPause', '() => {}', ';', 216, 'div'],
        ['onPlay', '() => {}', ';', 217, 'div'],
        ['onPlaying', '() => {}', ';', 218, 'div'],
        ['onProgress', '() => {}', ';', 219, 'div'],
        ['onRateChange', '() => {}', ';', 220, 'div'],
        ['onSeeked', '() => {}', ';', 221, 'div'],
        ['onSeeking', '() => {}', ';', 222, 'div'],
        ['onStalled', '() => {}', ';', 223, 'div'],
        ['onSuspend', '() => {}', ';', 224, 'div'],
        ['onTimeUpdate', '() => {}', ';', 225, 'div'],
        ['onVolumeChange', '() => {}', ';', 226, 'div'],
        ['onWaiting', '() => {}', ';', 227, 'div'],
        ['onLoad', '() => {}', ';', 228, 'div'],
        ['onAnimationEnd', '() => {}', ';', 231, 'div'],
        ['onAnimationIteration', '() => {}', ';', 232, 'div'],
        ['onTransitionEnd', '() => {}', ';', 233, 'div'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 0',
      code: '<TestComponent onClick={doFoo} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 1',
      code: '<Button onClick={doFoo} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 3',
      code: '<div />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 4',
      code: '<div className="foo" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 5',
      code: '<div className="foo" {...props} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 6',
      code: '<div onClick={() => void 0} aria-hidden />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 7',
      code: '<div onClick={() => void 0} aria-hidden={true} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 8',
      code: '<div onClick={null} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 9',
      code: '<input onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    ...(
      [
        ['onClick', '() => void 0', '', 10, 'button checkbox'],
        ['onClick', '() => void 0', '', 18, 'hidden'],
        ['onClick', '() => void 0', '', 29, 'text'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 33',
      code: '<button onClick={() => void 0} className="foo" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 37',
      code: '<select onClick={() => void 0} className="foo" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 38',
      code: '<textarea onClick={() => void 0} className="foo" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 39',
      code: '<a onClick={() => void 0} href="http://x.y.z" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 40',
      code: '<a onClick={() => void 0} href="http://x.y.z" tabIndex="0" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 44',
      code: '<div role="button" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 45',
      code: '<div role="checkbox" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 48',
      code: '<div role="form" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 50',
      code: '<div role="link" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 51',
      code: '<div role="menuitem" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 61',
      code: '<div role="tab" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 62',
      code: '<div role="textbox" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 63',
      code: '<div role="treeitem" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 64',
      code: '<div role="presentation" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 65',
      code: '<div role="presentation" onKeyDown={() => {}} />;',
      options: upstreamOptions0,
    },
    ...(
      [
        ['onClick', '() => {}', ';', 67, 'article'],
        ['onClick', '() => {}', ';', 69, 'aside blockquote'],
        ['onClick', '() => {}', ';', 88, 'h1'],
        ['onClick', '() => {}', ';', 101, 'li'],
        ['onClick', '() => void 0', ';', 102, 'main'],
        ['onClick', '() => {}', ';', 107, 'nav'],
        ['onClick', '() => {}', ';', 111, 'p'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 115',
      code: '<section onClick={() => {}} aria-label="Aa" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 121',
      code: '<table onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 130',
      code: '<div role="command" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 131',
      code: '<div role="composite" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 139',
      code: '<div role="widget" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 140',
      code: '<div role="window" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 141',
      code: '<div role="alert" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 144',
      code: '<div role="article" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 145',
      code: '<div role="banner" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 150',
      code: '<div role="dialog" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 152',
      code: '<div role="document" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    ...(
      [
        ['onCopy', '() => {}', ';', 189, 'div'],
        ['onCut', '() => {}', ';', 190, 'div'],
        ['onPaste', '() => {}', ';', 191, 'div'],
        ['onChange', '() => {}', ';', 195, 'div'],
        ['onSubmit', '() => {}', ';', 197, 'div'],
        ['onTouchStart', '() => {}', ';', 202, 'div'],
        ['onScroll', '() => {}', ';', 203, 'div'],
        ['onAnimationStart', '() => {}', ';', 230, 'div'],
        ['onDoubleClick', '() => {}', ';', 247, 'div'],
        ['onDragEnd', '() => {}', ';', 249, 'div'],
        ['onDragEnter', '() => {}', ';', 250, 'div'],
        ['onDragExit', '() => {}', ';', 251, 'div'],
        ['onDragLeave', '() => {}', ';', 252, 'div'],
        ['onDragOver', '() => {}', ';', 253, 'div'],
        ['onDragStart', '() => {}', ';', 254, 'div'],
        ['onDrop', '() => {}', ';', 255, 'div'],
        ['onMouseMove', '() => {}', ';', 258, 'div'],
        ['onMouseOut', '() => {}', ';', 259, 'div'],
        ['onMouseOver', '() => {}', ';', 260, 'div'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 299',
      code: '<div role={BUTTON} onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 300',
      code: '<div role={isButton ? "button" : "link"} onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 301',
      code: '<div role={isButton ? "button" : LINK} onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 302',
      code: '<div role={isButton ? BUTTON : LINK} onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended valid 2',
      code: '<Button onClick={doFoo} />',
      options: upstreamOptions0,
      settings: upstreamSettings0,
    },
  ],
  invalid: [
    ...(
      [
        ['onClick', '() => void 0', '', 5, 'a'],
        ['onClick', '() => {}', ';', 8, 'applet area'],
        ['onClick', '() => {}', ';', 11, 'base'],
        ['onClick', '() => {}', ';', 13, 'bdo big'],
        ['onClick', '() => {}', ';', 19, 'col'],
        ['onClick', '() => {}', ';', 21, 'content data'],
        ['onClick', '() => {}', ';', 24, 'frame frameset'],
        ['onClick', '() => {}', ';', 31, 'keygen'],
        ['onClick', '() => {}', ';', 34, 'noembed'],
        ['onClick', '() => {}', ';', 37, 'param'],
        ['onClick', '() => {}', ';', 41, 'rt rtc'],
        ['onClick', '() => {}', ';', 48, 'source'],
        ['onClick', '() => {}', ';', 51, 'strike'],
        ['onClick', '() => {}', ';', 54, 'track tt'],
        ['onClick', '() => {}', ';', 59, 'xmp'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:strict invalid ${firstCase + index}`,
          code,
          errors: [
            singleLineError(
              'noStaticElementInteractions',
              'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
              openingElement,
              1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:strict invalid 84',
      code: '<div  {...this.props} role={this.props.role} onKeyPress={e => this.handleKeyPress(e)}>{this.props.children}</div>',
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 87,
        },
      ],
    },
    ...(
      [
        ['onClick', '() => void 0', '', 5, 'a'],
        ['onClick', '() => {}', ';', 8, 'applet area'],
        ['onClick', '() => {}', ';', 11, 'base'],
        ['onClick', '() => {}', ';', 13, 'bdo big'],
        ['onClick', '() => {}', ';', 19, 'col'],
        ['onClick', '() => {}', ';', 21, 'content data'],
        ['onClick', '() => {}', ';', 24, 'frame frameset'],
        ['onClick', '() => {}', ';', 31, 'keygen'],
        ['onClick', '() => {}', ';', 34, 'noembed'],
        ['onClick', '() => {}', ';', 37, 'param'],
        ['onClick', '() => {}', ';', 41, 'rt rtc'],
        ['onClick', '() => {}', ';', 48, 'source'],
        ['onClick', '() => {}', ';', 51, 'strike'],
        ['onClick', '() => {}', ';', 54, 'track tt'],
        ['onClick', '() => {}', ';', 59, 'xmp'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:recommended invalid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
          errors: [
            singleLineError(
              'noStaticElementInteractions',
              'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
              openingElement,
              1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended invalid 1',
      code: '<div onClick={() => void 0} role={undefined} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended invalid 2',
      code: '<div onClick={() => void 0} {...props} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended invalid 3',
      code: '<div onKeyUp={() => void 0} aria-hidden={false} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended invalid 4',
      code: '<a onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended invalid 6',
      code: '<a tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    ...(
      [
        ['onClick', '() => {}', ';', 7, 'acronym'],
        ['onClick', '() => {}', ';', 10, 'b'],
        ['onClick', '() => {}', ';', 12, 'bdi'],
        ['onClick', '() => {}', ';', 15, 'blink body center cite'],
        ['onClick', '() => {}', ';', 20, 'colgroup'],
        ['onClick', '() => {}', ';', 23, 'font'],
        ['onClick', '() => {}', ';', 26, 'head header hgroup i kbd'],
        ['onClick', '() => {}', ';', 32, 'map meta'],
        ['onClick', '() => {}', ';', 35, 'noscript object'],
        ['onClick', '() => {}', ';', 38, 'picture q rp'],
        ['onClick', '() => {}', ';', 43, 's samp script section small'],
        ['onClick', '() => {}', ';', 49, 'spacer span'],
        ['onClick', '() => {}', ';', 52, 'style title'],
        ['onClick', '() => {}', ';', 56, 'u var wbr'],
        ['onKeyDown', '() => {}', ';', 60, 'div'],
        ['onKeyPress', '() => {}', ';', 61, 'div'],
        ['onKeyUp', '() => {}', ';', 62, 'div'],
        ['onClick', '() => {}', ';', 63, 'div'],
        ['onMouseDown', '() => {}', ';', 64, 'div'],
        ['onMouseUp', '() => {}', ';', 65, 'div'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-static-element-interactions:recommended invalid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
          errors: [
            singleLineError(
              'noStaticElementInteractions',
              'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
              openingElement,
              1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-static-element-interactions:recommended invalid 66',
      code: '<TestComponent onClick={doFoo} />',
      options: upstreamOptions0,
      settings: upstreamSettings0,
      errors: [
        {
          messageId: 'noStaticElementInteractions',
          message:
            'Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
  ],
});
