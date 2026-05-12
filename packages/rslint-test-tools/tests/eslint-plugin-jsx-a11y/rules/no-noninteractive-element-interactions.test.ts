import { RuleTester } from '../rule-tester';

const errorMessage =
  'Non-interactive elements should not be assigned mouse or keyboard event listeners.';
const expectedError = { message: errorMessage };

// Mirrors `configs.recommended.rules['jsx-a11y/no-noninteractive-element-interactions'][1]`
// from eslint-plugin-jsx-a11y/src/index.js.
const recommendedOptions = [
  {
    handlers: [
      'onClick',
      'onError',
      'onLoad',
      'onMouseDown',
      'onMouseUp',
      'onKeyPress',
      'onKeyDown',
      'onKeyUp',
    ],
    alert: ['onKeyUp', 'onKeyDown', 'onKeyPress'],
    body: ['onError', 'onLoad'],
    dialog: ['onKeyUp', 'onKeyDown', 'onKeyPress'],
    iframe: ['onError', 'onLoad'],
    img: ['onError', 'onLoad'],
  },
];

// Mirrors `configs.strict.rules['jsx-a11y/no-noninteractive-element-interactions'][1]`.
// Strict has no `handlers` override → defaultInteractiveProps (focus + image
// + keyboard + mouse) apply.
const strictOptions = [
  {
    body: ['onError', 'onLoad'],
    iframe: ['onError', 'onLoad'],
    img: ['onError', 'onLoad'],
  },
];

const imageComponentsSettings = {
  'jsx-a11y': {
    components: {
      Image: 'img',
    },
  },
};

const buttonComponentsSettings = {
  'jsx-a11y': {
    components: {
      Button: 'button',
    },
  },
};

const articleComponentsSettings = {
  'jsx-a11y': {
    components: {
      MyArticle: 'article',
    },
  },
};

const polymorphicSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
  },
};

new RuleTester().run('no-noninteractive-element-interactions', null as never, {
  valid: [
    // ============================================================
    // Recommended config — alwaysValid sample (mirrors upstream).
    // ============================================================
    { code: '<TestComponent onClick={doFoo} />', options: recommendedOptions },
    { code: '<Button onClick={doFoo} />', options: recommendedOptions },
    { code: '<Image onClick={() => void 0} />;', options: recommendedOptions },
    {
      code: '<Button onClick={() => void 0} />;',
      settings: buttonComponentsSettings,
      options: recommendedOptions,
    },

    // Inherent-interactive input (every type).
    {
      code: '<input onClick={() => void 0} />',
      options: recommendedOptions,
    },
    {
      code: '<input type="button" onClick={() => void 0} />',
      options: recommendedOptions,
    },
    {
      code: '<input type="checkbox" onClick={() => void 0} />',
      options: recommendedOptions,
    },
    {
      code: '<input type="hidden" onClick={() => void 0} />',
      options: recommendedOptions,
    },
    {
      code: '<input type="text" onClick={() => void 0} />',
      options: recommendedOptions,
    },

    // Interactive elements (anchor / button / option / etc.).
    {
      code: '<a onClick={() => void 0} href="http://x.y.z" />',
      options: recommendedOptions,
    },
    {
      code: '<button onClick={() => void 0} className="foo" />',
      options: recommendedOptions,
    },
    { code: '<menuitem onClick={() => {}} />;', options: recommendedOptions },
    { code: '<area onClick={() => {}} />;', options: recommendedOptions },
    { code: '<tr onClick={() => {}} />;', options: recommendedOptions },

    // Plain <div> — no opinion (no role, not classified as non-interactive).
    { code: '<div onClick={() => void 0} />;', options: recommendedOptions },
    {
      code: '<div onClick={() => void 0} role={undefined} />;',
      options: recommendedOptions,
    },
    { code: '<div onClick={null} />;', options: recommendedOptions },
    {
      code: '<div onClick={() => void 0} aria-hidden />;',
      options: recommendedOptions,
    },
    {
      code: '<div onClick={() => void 0} {...props} />;',
      options: recommendedOptions,
    },

    // Static elements (no opinion).
    { code: '<canvas onClick={() => {}} />;', options: recommendedOptions },
    { code: '<embed onClick={() => {}} />;', options: recommendedOptions },
    { code: '<span onClick={() => {}} />;', options: recommendedOptions },
    { code: '<header onClick={() => {}} />;', options: recommendedOptions },

    // body + per-element allow-list: onLoad filtered → bail.
    { code: '<body onLoad={() => {}} />;', options: recommendedOptions },
    // iframe + per-element allow-list: onLoad filtered.
    { code: '<iframe onLoad={() => {}} />;', options: recommendedOptions },
    // img + onLoad / onError filtered → bail.
    { code: '<img onLoad={() => {}} />;', options: recommendedOptions },
    {
      code: '<img src={currentPhoto.imageUrl} onLoad={this.handleImageLoad} alt="for review" />',
      options: recommendedOptions,
    },

    // Interactive roles on a `<div>` — bail.
    {
      code: '<div role="button" onClick={() => {}} />;',
      options: recommendedOptions,
    },
    {
      code: '<div role="link" onClick={() => {}} />;',
      options: recommendedOptions,
    },
    {
      code: '<div role="menuitem" onClick={() => {}} />;',
      options: recommendedOptions,
    },
    {
      code: '<div role="presentation" onClick={() => {}} />;',
      options: recommendedOptions,
    },

    // Abstract roles — bail.
    {
      code: '<div role="command" onClick={() => {}} />;',
      options: recommendedOptions,
    },
    {
      code: '<div role="widget" onClick={() => {}} />;',
      options: recommendedOptions,
    },
    {
      code: '<div role="window" onClick={() => {}} />;',
      options: recommendedOptions,
    },

    // Non-triggering handler — onCopy isn't in recommended's `handlers`.
    {
      code: '<div role="article" onCopy={() => {}} />;',
      options: recommendedOptions,
    },
    {
      code: '<div role="article" onFocus={() => {}} />;',
      options: recommendedOptions,
    },
    {
      code: '<div role="article" onSubmit={() => {}} />;',
      options: recommendedOptions,
    },

    // ============================================================
    // Default options (no `handlers` override) — focus + image +
    // keyboard + mouse trigger.
    // ============================================================
    { code: '<article onCopy={fn} />' },
    { code: '<article onSubmit={fn} />' },

    // ============================================================
    // contentEditable bail-out — raw match against `"true"`.
    // ============================================================
    { code: '<article contentEditable="true" onClick={fn} />' },
    { code: '<div role="article" contenteditable="true" onClick={fn} />' },

    // ============================================================
    // IsHiddenFromScreenReader / IsPresentationRole.
    // ============================================================
    { code: '<article aria-hidden onClick={fn} />' },
    { code: '<article aria-hidden={true} onClick={fn} />' },
    { code: '<article aria-hidden="true" onClick={fn} />' },
    { code: '<article role="presentation" onClick={fn} />' },
    { code: '<article role="none" onClick={fn} />' },

    // ============================================================
    // Options: explicit `handlers` override / empty array.
    // ============================================================
    {
      code: '<article onMouseDown={fn} />',
      options: [{ handlers: ['onClick'] }],
    },
    {
      code: '<article onClick={fn} />',
      options: [{ handlers: [] }],
    },
    // Per-element allow-list.
    {
      code: '<article onClick={fn} />',
      options: [{ article: ['onClick'] }],
    },

    // ============================================================
    // Settings: polymorphic / components.
    // ============================================================
    {
      code: '<Foo as="button" onClick={() => void 0} />',
      settings: polymorphicSettings,
    },
    {
      code: '<Foo as="input" type="text" onClick={() => void 0} />',
      settings: polymorphicSettings,
    },
    {
      code: '<MyButton onClick={() => void 0} />',
      settings: buttonComponentsSettings,
    },
  ],
  invalid: [
    // ============================================================
    // Recommended config — neverValid sample (mirrors upstream).
    // ============================================================
    {
      code: '<main onClick={() => void 0} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<address onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<article onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<aside onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<br onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<li onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<ul onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<iframe onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<img onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<form onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<h1 onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<table onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },

    // contentEditable nuances.
    {
      code: '<ul contentEditable="false" onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<article contentEditable onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<div contentEditable role="article" onKeyDown={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },

    // Non-interactive roles.
    {
      code: '<div role="alert" onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<div role="article" onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<div role="article" onLoad={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<div role="article" onError={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },
    {
      code: '<div role="article" onKeyDown={() => {}} />;',
      options: recommendedOptions,
      errors: [expectedError],
    },

    // Custom component → DOM (settings).
    {
      code: '<Image onClick={() => void 0} />;',
      settings: imageComponentsSettings,
      options: recommendedOptions,
      errors: [expectedError],
    },

    // ============================================================
    // Strict config — focus / contextMenu / etc. fire under default.
    // ============================================================
    {
      code: '<div role="article" onFocus={() => {}} />;',
      options: strictOptions,
      errors: [expectedError],
    },
    {
      code: '<div role="article" onContextMenu={() => {}} />;',
      options: strictOptions,
      errors: [expectedError],
    },
    {
      code: '<div role="article" onMouseEnter={() => {}} />;',
      options: strictOptions,
      errors: [expectedError],
    },

    // ============================================================
    // Default options (no override) — focus + image + keyboard +
    // mouse handlers trigger; non-default handlers (clipboard, form,
    // touch …) do NOT.
    // ============================================================
    { code: '<article onClick={fn} />', errors: [expectedError] },
    { code: '<article onLoad={fn} />', errors: [expectedError] },
    { code: '<article onError={fn} />', errors: [expectedError] },
    { code: '<article onFocus={fn} />', errors: [expectedError] },
    { code: '<article onKeyDown={fn} />', errors: [expectedError] },

    // contentEditable raw-match nuances.
    {
      code: '<article contentEditable={true} onClick={fn} />',
      errors: [expectedError],
    },
    {
      code: '<article contentEditable="True" onClick={fn} />',
      errors: [expectedError],
    },
    {
      code: '<article contentEditable={"true"} onClick={fn} />',
      errors: [expectedError],
    },

    // role="NONE" / role="PRESENTATION" — case-sensitive.
    {
      code: '<article role="NONE" onClick={fn} />',
      errors: [expectedError],
    },
    {
      code: '<article role="PRESENTATION" onClick={fn} />',
      errors: [expectedError],
    },

    // ============================================================
    // Spread shapes — literal-spread of onClick walks the literal.
    // ============================================================
    {
      code: '<article {...{onClick: fn}} />',
      errors: [expectedError],
    },
    {
      code: '<article {...{onClick}} />',
      errors: [expectedError],
    },

    // ============================================================
    // Settings: polymorphic / components.
    // ============================================================
    {
      code: '<Foo as="article" onClick={() => void 0} />',
      settings: polymorphicSettings,
      errors: [expectedError],
    },
    {
      code: '<MyArticle onClick={() => void 0} />',
      settings: articleComponentsSettings,
      errors: [expectedError],
    },

    // ============================================================
    // Options: explicit `handlers` override pulls in non-default.
    // ============================================================
    {
      code: '<article onSubmit={fn} />',
      options: [{ handlers: ['onSubmit'] }],
      errors: [expectedError],
    },
    // handlers + per-element allow-list combined.
    {
      code: '<iframe onClick={fn} />',
      options: [
        { handlers: ['onClick', 'onLoad'], iframe: ['onLoad', 'onError'] },
      ],
      errors: [expectedError],
    },
  ],
});
