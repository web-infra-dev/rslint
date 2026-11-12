import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('mouse-events-have-key-events', {
  valid: [
    // ============================================================
    // Upstream valid suite (mirrors __tests__/src/rules/mouse-events-have-key-events-test.js)
    // ============================================================
    { code: '<div onMouseOver={() => void 0} onFocus={() => void 0} />;' },
    {
      code: '<div onMouseOver={() => void 0} onFocus={() => void 0} {...props} />;',
    },
    { code: '<div onMouseOver={handleMouseOver} onFocus={handleFocus} />;' },
    {
      code: '<div onMouseOver={handleMouseOver} onFocus={handleFocus} {...props} />;',
    },
    { code: '<div />;' },
    { code: '<div onBlur={() => {}} />' },
    { code: '<div onFocus={() => {}} />' },
    { code: '<div onMouseOut={() => void 0} onBlur={() => void 0} />' },
    {
      code: '<div onMouseOut={() => void 0} onBlur={() => void 0} {...props} />',
    },
    { code: '<div onMouseOut={handleMouseOut} onBlur={handleOnBlur} />' },
    {
      code: '<div onMouseOut={handleMouseOut} onBlur={handleOnBlur} {...props} />',
    },
    { code: '<MyElement />' },
    { code: '<MyElement onMouseOver={() => {}} />' },
    { code: '<MyElement onMouseOut={() => {}} />' },
    { code: '<MyElement onBlur={() => {}} />' },
    { code: '<MyElement onFocus={() => {}} />' },
    { code: '<MyElement onMouseOver={() => {}} {...props} />' },
    { code: '<MyElement onMouseOut={() => {}} {...props} />' },
    { code: '<MyElement onBlur={() => {}} {...props} />' },
    { code: '<MyElement onFocus={() => {}} {...props} />' },

    // ---- Empty options disable both pairings. ----
    {
      code: '<div onMouseOver={() => {}} onMouseOut={() => {}} />',
      options: [{ hoverInHandlers: [], hoverOutHandlers: [] }],
    },

    // ---- Custom handlers — pair check still applies. ----
    {
      code: '<div onMouseOver={() => {}} onFocus={() => {}} />',
      options: [{ hoverInHandlers: ['onMouseOver'] }],
    },
    {
      code: '<div onMouseEnter={() => {}} onFocus={() => {}} />',
      options: [{ hoverInHandlers: ['onMouseEnter'] }],
    },
    {
      code: '<div onMouseOut={() => {}} onBlur={() => {}} />',
      options: [{ hoverOutHandlers: ['onMouseOut'] }],
    },
    {
      code: '<div onMouseLeave={() => {}} onBlur={() => {}} />',
      options: [{ hoverOutHandlers: ['onMouseLeave'] }],
    },
    {
      code: '<div onMouseOver={() => {}} onMouseOut={() => {}} />',
      options: [
        {
          hoverInHandlers: ['onPointerEnter'],
          hoverOutHandlers: ['onPointerLeave'],
        },
      ],
    },
    {
      code: '<div onMouseLeave={() => {}} />',
      options: [{ hoverOutHandlers: ['onPointerLeave'] }],
    },

    // ============================================================
    // Extras: hover-handler value is statically nullish → not counted.
    // ============================================================
    { code: '<div onMouseOver={null} />' },
    { code: '<div onMouseOver={undefined} />' },
    { code: '<div onMouseOut={null} />' },
    { code: '<div onMouseOut={undefined} />' },

    // ============================================================
    // Extras: case-insensitive matching on both sides
    // ============================================================
    { code: '<div onmouseover={() => {}} onfocus={() => {}} />' },
    { code: '<div onmouseout={() => {}} onblur={() => {}} />' },

    // ============================================================
    // Extras: member-expression / namespaced tags don't classify as DOM.
    // ============================================================
    { code: '<Foo.Bar onMouseOver={() => {}} />' },
    { code: '<svg:circle onMouseOver={() => {}} />' },
  ],
  invalid: [
    // ============================================================
    // Upstream invalid suite
    // ============================================================
    {
      code: '<div onMouseOver={() => void 0} />;',
      errors: [
        {
          messageId: 'mouseOver',
          message:
            'onMouseOver must be accompanied by onFocus for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<div onMouseOut={() => void 0} />',
      errors: [
        {
          messageId: 'mouseOut',
          message:
            'onMouseOut must be accompanied by onBlur for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div onMouseOver={() => void 0} onFocus={undefined} />;',
      errors: [
        {
          messageId: 'mouseOver',
          message:
            'onMouseOver must be accompanied by onFocus for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<div onMouseOut={() => void 0} onBlur={undefined} />',
      errors: [
        {
          messageId: 'mouseOut',
          message:
            'onMouseOut must be accompanied by onBlur for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div onMouseOver={() => void 0} {...props} />',
      errors: [
        {
          messageId: 'mouseOver',
          message:
            'onMouseOver must be accompanied by onFocus for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<div onMouseOut={() => void 0} {...props} />',
      errors: [
        {
          messageId: 'mouseOut',
          message:
            'onMouseOut must be accompanied by onBlur for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div onMouseOver={() => {}} onMouseOut={() => {}} />',
      options: [
        {
          hoverInHandlers: ['onMouseOver'],
          hoverOutHandlers: ['onMouseOut'],
        },
      ],
      errors: [
        {
          messageId: 'mouseOver',
          message:
            'onMouseOver must be accompanied by onFocus for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 28,
        },
        {
          messageId: 'mouseOut',
          message:
            'onMouseOut must be accompanied by onBlur for accessibility.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: '<div onPointerEnter={() => {}} onPointerLeave={() => {}} />',
      options: [
        {
          hoverInHandlers: ['onPointerEnter'],
          hoverOutHandlers: ['onPointerLeave'],
        },
      ],
      errors: [
        {
          messageId: 'mouseOver',
          message:
            'onPointerEnter must be accompanied by onFocus for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 31,
        },
        {
          messageId: 'mouseOut',
          message:
            'onPointerLeave must be accompanied by onBlur for accessibility.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: '<div onMouseOver={() => {}} />',
      options: [{ hoverInHandlers: ['onMouseOver'] }],
      errors: [
        {
          messageId: 'mouseOver',
          message:
            'onMouseOver must be accompanied by onFocus for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<div onPointerEnter={() => {}} />',
      options: [{ hoverInHandlers: ['onPointerEnter'] }],
      errors: [
        {
          messageId: 'mouseOver',
          message:
            'onPointerEnter must be accompanied by onFocus for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div onMouseOut={() => {}} />',
      options: [{ hoverOutHandlers: ['onMouseOut'] }],
      errors: [
        {
          messageId: 'mouseOut',
          message:
            'onMouseOut must be accompanied by onBlur for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div onPointerLeave={() => {}} />',
      options: [{ hoverOutHandlers: ['onPointerLeave'] }],
      errors: [
        {
          messageId: 'mouseOut',
          message:
            'onPointerLeave must be accompanied by onBlur for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // ============================================================
    // Extras: boolean attribute form
    // ============================================================
    {
      code: '<div onMouseOver />',
      errors: [
        {
          messageId: 'mouseOver',
          message:
            'onMouseOver must be accompanied by onFocus for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<div onMouseOut />',
      errors: [
        {
          messageId: 'mouseOut',
          message:
            'onMouseOut must be accompanied by onBlur for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },

    // ============================================================
    // Extras: case-insensitive on lowercase
    // ============================================================
    {
      code: '<div onmouseover={() => {}} />',
      errors: [
        {
          messageId: 'mouseOver',
          message:
            'onMouseOver must be accompanied by onFocus for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<div onmouseout={() => {}} />',
      errors: [
        {
          messageId: 'mouseOut',
          message:
            'onMouseOut must be accompanied by onBlur for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },

    // ============================================================
    // Extras: onFocus / onBlur explicitly null → counted as missing.
    // ============================================================
    {
      code: '<div onMouseOver={fn} onFocus={null} />',
      errors: [
        {
          messageId: 'mouseOver',
          message:
            'onMouseOver must be accompanied by onFocus for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div onMouseOut={fn} onBlur={null} />',
      errors: [
        {
          messageId: 'mouseOut',
          message:
            'onMouseOut must be accompanied by onBlur for accessibility.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
  ],
});
