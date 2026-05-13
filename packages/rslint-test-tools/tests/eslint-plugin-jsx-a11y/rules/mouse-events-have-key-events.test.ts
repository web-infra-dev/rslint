import { RuleTester } from '../rule-tester';

const mouseOverError = {
  message: 'onMouseOver must be accompanied by onFocus for accessibility.',
};
const mouseOutError = {
  message: 'onMouseOut must be accompanied by onBlur for accessibility.',
};
const pointerEnterError = {
  message: 'onPointerEnter must be accompanied by onFocus for accessibility.',
};
const pointerLeaveError = {
  message: 'onPointerLeave must be accompanied by onBlur for accessibility.',
};

new RuleTester().run('mouse-events-have-key-events', null as never, {
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
      errors: [mouseOverError],
    },
    {
      code: '<div onMouseOut={() => void 0} />',
      errors: [mouseOutError],
    },
    {
      code: '<div onMouseOver={() => void 0} onFocus={undefined} />;',
      errors: [mouseOverError],
    },
    {
      code: '<div onMouseOut={() => void 0} onBlur={undefined} />',
      errors: [mouseOutError],
    },
    {
      code: '<div onMouseOver={() => void 0} {...props} />',
      errors: [mouseOverError],
    },
    {
      code: '<div onMouseOut={() => void 0} {...props} />',
      errors: [mouseOutError],
    },
    {
      code: '<div onMouseOver={() => {}} onMouseOut={() => {}} />',
      options: [
        {
          hoverInHandlers: ['onMouseOver'],
          hoverOutHandlers: ['onMouseOut'],
        },
      ],
      errors: [mouseOverError, mouseOutError],
    },
    {
      code: '<div onPointerEnter={() => {}} onPointerLeave={() => {}} />',
      options: [
        {
          hoverInHandlers: ['onPointerEnter'],
          hoverOutHandlers: ['onPointerLeave'],
        },
      ],
      errors: [pointerEnterError, pointerLeaveError],
    },
    {
      code: '<div onMouseOver={() => {}} />',
      options: [{ hoverInHandlers: ['onMouseOver'] }],
      errors: [mouseOverError],
    },
    {
      code: '<div onPointerEnter={() => {}} />',
      options: [{ hoverInHandlers: ['onPointerEnter'] }],
      errors: [pointerEnterError],
    },
    {
      code: '<div onMouseOut={() => {}} />',
      options: [{ hoverOutHandlers: ['onMouseOut'] }],
      errors: [mouseOutError],
    },
    {
      code: '<div onPointerLeave={() => {}} />',
      options: [{ hoverOutHandlers: ['onPointerLeave'] }],
      errors: [pointerLeaveError],
    },

    // ============================================================
    // Extras: boolean attribute form
    // ============================================================
    {
      code: '<div onMouseOver />',
      errors: [mouseOverError],
    },
    {
      code: '<div onMouseOut />',
      errors: [mouseOutError],
    },

    // ============================================================
    // Extras: case-insensitive on lowercase
    // ============================================================
    {
      code: '<div onmouseover={() => {}} />',
      errors: [mouseOverError],
    },
    {
      code: '<div onmouseout={() => {}} />',
      errors: [mouseOutError],
    },

    // ============================================================
    // Extras: onFocus / onBlur explicitly null → counted as missing.
    // ============================================================
    {
      code: '<div onMouseOver={fn} onFocus={null} />',
      errors: [mouseOverError],
    },
    {
      code: '<div onMouseOut={fn} onBlur={null} />',
      errors: [mouseOutError],
    },
  ],
});
