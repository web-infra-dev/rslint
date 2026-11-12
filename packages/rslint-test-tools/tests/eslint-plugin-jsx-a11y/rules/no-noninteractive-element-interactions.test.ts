import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';
import { singleLineError } from '../matrix';

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

new RuleTester(config).run('no-noninteractive-element-interactions', {
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
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<address onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<article onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<aside onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<br onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<li onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<ul onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<iframe onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<img onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<form onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<h1 onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<table onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },

    // contentEditable nuances.
    {
      code: '<ul contentEditable="false" onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: '<article contentEditable onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: '<div contentEditable role="article" onKeyDown={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },

    // Non-interactive roles.
    {
      code: '<div role="alert" onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: '<div role="article" onClick={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '<div role="article" onLoad={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: '<div role="article" onError={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '<div role="article" onKeyDown={() => {}} />;',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },

    // Custom component → DOM (settings).
    {
      code: '<Image onClick={() => void 0} />;',
      settings: imageComponentsSettings,
      options: recommendedOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },

    // ============================================================
    // Strict config — focus / contextMenu / etc. fire under default.
    // ============================================================
    {
      code: '<div role="article" onFocus={() => {}} />;',
      options: strictOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '<div role="article" onContextMenu={() => {}} />;',
      options: strictOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '<div role="article" onMouseEnter={() => {}} />;',
      options: strictOptions,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },

    // ============================================================
    // Default options (no override) — focus + image + keyboard +
    // mouse handlers trigger; non-default handlers (clipboard, form,
    // touch …) do NOT.
    // ============================================================
    {
      code: '<article onClick={fn} />',
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<article onLoad={fn} />',
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<article onError={fn} />',
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<article onFocus={fn} />',
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<article onKeyDown={fn} />',
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },

    // contentEditable raw-match nuances.
    {
      code: '<article contentEditable={true} onClick={fn} />',
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '<article contentEditable="True" onClick={fn} />',
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '<article contentEditable={"true"} onClick={fn} />',
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },

    // role="NONE" / role="PRESENTATION" — case-sensitive.
    {
      code: '<article role="NONE" onClick={fn} />',
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<article role="PRESENTATION" onClick={fn} />',
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },

    // ============================================================
    // Spread shapes — literal-spread of onClick walks the literal.
    // ============================================================
    {
      code: '<article {...{onClick: fn}} />',
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<article {...{onClick}} />',
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },

    // ============================================================
    // Settings: polymorphic / components.
    // ============================================================
    {
      code: '<Foo as="article" onClick={() => void 0} />',
      settings: polymorphicSettings,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      code: '<MyArticle onClick={() => void 0} />',
      settings: articleComponentsSettings,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },

    // ============================================================
    // Options: explicit `handlers` override pulls in non-default.
    // ============================================================
    {
      code: '<article onSubmit={fn} />',
      options: [{ handlers: ['onSubmit'] }],
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    // handlers + per-element allow-list combined.
    {
      code: '<iframe onClick={fn} />',
      options: [
        { handlers: ['onClick', 'onLoad'], iframe: ['onLoad', 'onError'] },
      ],
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
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
const upstreamOptions1 = [
  {
    body: ['onError', 'onLoad'],
    iframe: ['onError', 'onLoad'],
    img: ['onError', 'onLoad'],
  },
];
const upstreamSettings0 = { 'jsx-a11y': { components: { Button: 'button' } } };
const upstreamSettings1 = { 'jsx-a11y': { components: { Image: 'img' } } };

new RuleTester(config).run('no-noninteractive-element-interactions', {
  valid: [
    ...(
      [
        [
          'onClick',
          '() => void 0',
          '',
          7,
          'color date datetime datetime-local email file image month number password radio range reset search submit tel',
        ],
        ['onClick', '() => void 0', '', 24, 'time url week'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    ...(
      [
        ['onClick', '() => void 0', '', 28, 'a'],
        ['onClick', '() => {}', ';', 29, 'a'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 30',
      code: '<a tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 32',
      code: '<a onClick={() => void 0} href="http://x.y.z" tabIndex="0" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 34',
      code: '<body onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 37',
      code: '<option onClick={() => void 0} className="foo" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 38',
      code: '<select onClick={() => void 0} className="foo" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 39',
      code: '<textarea onClick={() => void 0} className="foo" />',
      options: upstreamOptions0,
    },
    ...(
      [
        [
          'onClick',
          '() => {}',
          ';',
          41,
          'acronym applet audio b base bdi bdo big blink',
        ],
        [
          'onClick',
          '() => {}',
          ';',
          52,
          'center cite col colgroup content data datalist',
        ],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 59',
      code: '<div />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 60',
      code: '<div className="foo" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 61',
      code: '<div className="foo" {...props} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 63',
      code: '<div onClick={() => void 0} aria-hidden={true} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 68',
      code: '<div onKeyUp={() => void 0} aria-hidden={false} />;',
      options: upstreamOptions0,
    },
    ...(
      [
        ['onClick', '() => {}', ';', 70, 'font'],
        ['onSubmit', '() => {}', ';', 71, 'font form'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 73',
      code: '\n      <form onSubmit={this.handleSubmit.bind(this)} method="POST">\n        <button type="submit">\n            Save\n        </button>\n      </form>\n    ',
      options: upstreamOptions0,
    },
    ...(
      [
        ['onClick', '() => {}', ';', 74, 'frame frameset head'],
        ['onClick', '() => {}', ';', 78, 'hgroup i'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 81',
      code: '\n      <iframe\n        name="embeddedExternalPayment"\n        ref="embeddedExternalPayment"\n        style={iframeStyle}\n        onLoad={this.handleLoadIframe}\n      />\n    ',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 82',
      code: '<img {...props} onError={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 85',
      code: '\n        <img\n        ref={this.ref}\n        className="c-responsive-image-placeholder__image"\n        src={src}\n        alt={alt}\n        data-test-id="test-id"\n        onLoad={this.fetchCompleteImage}\n      />\n    ',
      options: upstreamOptions0,
    },
    ...([['onClick', '() => {}', ';', 86, 'kbd keygen link']] as const).flatMap(
      ([handler, body, suffix, firstCase, values]) =>
        values.split(' ').map((tag, index) => {
          const openingElement = `<${tag} ${handler}={${body}} />`;
          const code = openingElement + suffix;
          return {
            name: `upstream v6.10.2 no-noninteractive-element-interactions:recommended valid ${firstCase + index}`,
            code,
            options: upstreamOptions0,
          };
        }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 89',
      code: '<main onClick={null} />;',
      options: upstreamOptions0,
    },
    ...(
      [
        [
          'onClick',
          '() => {}',
          ';',
          90,
          'map meta noembed noscript object param picture q rp rt rtc s samp script section small source spacer',
        ],
        [
          'onClick',
          '() => {}',
          ';',
          109,
          'strike style summary th title track td tt u var video wbr xmp',
        ],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 123',
      code: '<div role="checkbox" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 124',
      code: '<div role="columnheader" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 125',
      code: '<div role="combobox" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 126',
      code: '<div role="grid" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 127',
      code: '<div role="gridcell" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 129',
      code: '<div role="listbox" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 130',
      code: '<div role="menu" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 131',
      code: '<div role="menubar" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 133',
      code: '<div role="menuitemcheckbox" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 134',
      code: '<div role="menuitemradio" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 135',
      code: '<div role="option" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 136',
      code: '<div role="progressbar" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 137',
      code: '<div role="radio" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 138',
      code: '<div role="radiogroup" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 139',
      code: '<div role="row" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 140',
      code: '<div role="rowheader" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 141',
      code: '<div role="scrollbar" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 142',
      code: '<div role="searchbox" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 143',
      code: '<div role="slider" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 144',
      code: '<div role="spinbutton" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 145',
      code: '<div role="switch" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 146',
      code: '<div role="tab" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 147',
      code: '<div role="textbox" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 148',
      code: '<div role="treeitem" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 151',
      code: '<div role="composite" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 152',
      code: '<div role="input" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 153',
      code: '<div role="landmark" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 154',
      code: '<div role="range" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 155',
      code: '<div role="roletype" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 156',
      code: '<div role="sectionhead" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 157',
      code: '<div role="select" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 158',
      code: '<div role="structure" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 159',
      code: '<div role="tablist" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 160',
      code: '<div role="toolbar" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 161',
      code: '<div role="tree" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 162',
      code: '<div role="treegrid" onClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 166',
      code: '<div role="article" onCut={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 167',
      code: '<div role="article" onPaste={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 168',
      code: '<div role="article" onCompositionEnd={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 169',
      code: '<div role="article" onCompositionStart={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 170',
      code: '<div role="article" onCompositionUpdate={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 171',
      code: '<div role="article" onChange={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 172',
      code: '<div role="article" onInput={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 174',
      code: '<div role="article" onSelect={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 175',
      code: '<div role="article" onTouchCancel={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 176',
      code: '<div role="article" onTouchEnd={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 177',
      code: '<div role="article" onTouchMove={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 178',
      code: '<div role="article" onTouchStart={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 179',
      code: '<div role="article" onScroll={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 180',
      code: '<div role="article" onWheel={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 181',
      code: '<div role="article" onAbort={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 182',
      code: '<div role="article" onCanPlay={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 183',
      code: '<div role="article" onCanPlayThrough={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 184',
      code: '<div role="article" onDurationChange={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 185',
      code: '<div role="article" onEmptied={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 186',
      code: '<div role="article" onEncrypted={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 187',
      code: '<div role="article" onEnded={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 188',
      code: '<div role="article" onLoadedData={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 189',
      code: '<div role="article" onLoadedMetadata={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 190',
      code: '<div role="article" onLoadStart={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 191',
      code: '<div role="article" onPause={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 192',
      code: '<div role="article" onPlay={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 193',
      code: '<div role="article" onPlaying={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 194',
      code: '<div role="article" onProgress={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 195',
      code: '<div role="article" onRateChange={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 196',
      code: '<div role="article" onSeeked={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 197',
      code: '<div role="article" onSeeking={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 198',
      code: '<div role="article" onStalled={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 199',
      code: '<div role="article" onSuspend={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 200',
      code: '<div role="article" onTimeUpdate={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 201',
      code: '<div role="article" onVolumeChange={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 202',
      code: '<div role="article" onWaiting={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 203',
      code: '<div role="article" onAnimationStart={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 204',
      code: '<div role="article" onAnimationEnd={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 205',
      code: '<div role="article" onAnimationIteration={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 206',
      code: '<div role="article" onTransitionEnd={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 214',
      code: '<div role="article" onBlur={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 219',
      code: '<div role="article" onDblClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 220',
      code: '<div role="article" onDoubleClick={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 221',
      code: '<div role="article" onDrag={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 222',
      code: '<div role="article" onDragEnd={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 223',
      code: '<div role="article" onDragEnter={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 224',
      code: '<div role="article" onDragExit={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 225',
      code: '<div role="article" onDragLeave={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 226',
      code: '<div role="article" onDragOver={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 227',
      code: '<div role="article" onDragStart={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 228',
      code: '<div role="article" onDrop={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 230',
      code: '<div role="article" onMouseLeave={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 231',
      code: '<div role="article" onMouseMove={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 232',
      code: '<div role="article" onMouseOut={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 233',
      code: '<div role="article" onMouseOver={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 218',
      code: '<div role="article" onContextMenu={() => {}} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended valid 229',
      code: '<div role="article" onMouseEnter={() => {}} />;',
      options: upstreamOptions0,
    },
    ...(
      [
        [
          'onClick',
          '() => void 0',
          '',
          7,
          'color date datetime datetime-local email file image month number password radio range reset search submit tel',
        ],
        ['onClick', '() => void 0', '', 24, 'time url week'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:strict valid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
        };
      }),
    ),
    ...(
      [
        ['onClick', '() => void 0', '', 28, 'a'],
        ['onClick', '() => {}', ';', 29, 'a'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:strict valid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 30',
      code: '<a tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 32',
      code: '<a onClick={() => void 0} href="http://x.y.z" tabIndex="0" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 34',
      code: '<body onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 37',
      code: '<option onClick={() => void 0} className="foo" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 38',
      code: '<select onClick={() => void 0} className="foo" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 39',
      code: '<textarea onClick={() => void 0} className="foo" />',
      options: upstreamOptions1,
    },
    ...(
      [
        [
          'onClick',
          '() => {}',
          ';',
          41,
          'acronym applet audio b base bdi bdo big blink',
        ],
        [
          'onClick',
          '() => {}',
          ';',
          52,
          'center cite col colgroup content data datalist',
        ],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:strict valid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 59',
      code: '<div />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 60',
      code: '<div className="foo" />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 61',
      code: '<div className="foo" {...props} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 63',
      code: '<div onClick={() => void 0} aria-hidden={true} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 68',
      code: '<div onKeyUp={() => void 0} aria-hidden={false} />;',
      options: upstreamOptions1,
    },
    ...(
      [
        ['onClick', '() => {}', ';', 70, 'font'],
        ['onSubmit', '() => {}', ';', 71, 'font form'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:strict valid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 73',
      code: '\n      <form onSubmit={this.handleSubmit.bind(this)} method="POST">\n        <button type="submit">\n            Save\n        </button>\n      </form>\n    ',
      options: upstreamOptions1,
    },
    ...(
      [
        ['onClick', '() => {}', ';', 74, 'frame frameset head'],
        ['onClick', '() => {}', ';', 78, 'hgroup i'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:strict valid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 81',
      code: '\n      <iframe\n        name="embeddedExternalPayment"\n        ref="embeddedExternalPayment"\n        style={iframeStyle}\n        onLoad={this.handleLoadIframe}\n      />\n    ',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 82',
      code: '<img {...props} onError={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 85',
      code: '\n        <img\n        ref={this.ref}\n        className="c-responsive-image-placeholder__image"\n        src={src}\n        alt={alt}\n        data-test-id="test-id"\n        onLoad={this.fetchCompleteImage}\n      />\n    ',
      options: upstreamOptions1,
    },
    ...([['onClick', '() => {}', ';', 86, 'kbd keygen link']] as const).flatMap(
      ([handler, body, suffix, firstCase, values]) =>
        values.split(' ').map((tag, index) => {
          const openingElement = `<${tag} ${handler}={${body}} />`;
          const code = openingElement + suffix;
          return {
            name: `upstream v6.10.2 no-noninteractive-element-interactions:strict valid ${firstCase + index}`,
            code,
            options: upstreamOptions1,
          };
        }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 89',
      code: '<main onClick={null} />;',
      options: upstreamOptions1,
    },
    ...(
      [
        [
          'onClick',
          '() => {}',
          ';',
          90,
          'map meta noembed noscript object param picture q rp rt rtc s samp script section small source spacer',
        ],
        [
          'onClick',
          '() => {}',
          ';',
          109,
          'strike style summary th title track td tt u var video wbr xmp',
        ],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:strict valid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 123',
      code: '<div role="checkbox" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 124',
      code: '<div role="columnheader" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 125',
      code: '<div role="combobox" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 126',
      code: '<div role="grid" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 127',
      code: '<div role="gridcell" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 129',
      code: '<div role="listbox" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 130',
      code: '<div role="menu" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 131',
      code: '<div role="menubar" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 133',
      code: '<div role="menuitemcheckbox" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 134',
      code: '<div role="menuitemradio" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 135',
      code: '<div role="option" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 136',
      code: '<div role="progressbar" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 137',
      code: '<div role="radio" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 138',
      code: '<div role="radiogroup" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 139',
      code: '<div role="row" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 140',
      code: '<div role="rowheader" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 141',
      code: '<div role="scrollbar" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 142',
      code: '<div role="searchbox" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 143',
      code: '<div role="slider" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 144',
      code: '<div role="spinbutton" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 145',
      code: '<div role="switch" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 146',
      code: '<div role="tab" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 147',
      code: '<div role="textbox" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 148',
      code: '<div role="treeitem" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 151',
      code: '<div role="composite" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 152',
      code: '<div role="input" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 153',
      code: '<div role="landmark" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 154',
      code: '<div role="range" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 155',
      code: '<div role="roletype" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 156',
      code: '<div role="sectionhead" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 157',
      code: '<div role="select" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 158',
      code: '<div role="structure" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 159',
      code: '<div role="tablist" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 160',
      code: '<div role="toolbar" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 161',
      code: '<div role="tree" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 162',
      code: '<div role="treegrid" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 166',
      code: '<div role="article" onCut={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 167',
      code: '<div role="article" onPaste={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 168',
      code: '<div role="article" onCompositionEnd={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 169',
      code: '<div role="article" onCompositionStart={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 170',
      code: '<div role="article" onCompositionUpdate={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 171',
      code: '<div role="article" onChange={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 172',
      code: '<div role="article" onInput={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 174',
      code: '<div role="article" onSelect={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 175',
      code: '<div role="article" onTouchCancel={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 176',
      code: '<div role="article" onTouchEnd={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 177',
      code: '<div role="article" onTouchMove={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 178',
      code: '<div role="article" onTouchStart={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 179',
      code: '<div role="article" onScroll={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 180',
      code: '<div role="article" onWheel={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 181',
      code: '<div role="article" onAbort={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 182',
      code: '<div role="article" onCanPlay={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 183',
      code: '<div role="article" onCanPlayThrough={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 184',
      code: '<div role="article" onDurationChange={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 185',
      code: '<div role="article" onEmptied={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 186',
      code: '<div role="article" onEncrypted={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 187',
      code: '<div role="article" onEnded={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 188',
      code: '<div role="article" onLoadedData={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 189',
      code: '<div role="article" onLoadedMetadata={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 190',
      code: '<div role="article" onLoadStart={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 191',
      code: '<div role="article" onPause={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 192',
      code: '<div role="article" onPlay={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 193',
      code: '<div role="article" onPlaying={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 194',
      code: '<div role="article" onProgress={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 195',
      code: '<div role="article" onRateChange={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 196',
      code: '<div role="article" onSeeked={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 197',
      code: '<div role="article" onSeeking={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 198',
      code: '<div role="article" onStalled={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 199',
      code: '<div role="article" onSuspend={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 200',
      code: '<div role="article" onTimeUpdate={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 201',
      code: '<div role="article" onVolumeChange={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 202',
      code: '<div role="article" onWaiting={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 203',
      code: '<div role="article" onAnimationStart={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 204',
      code: '<div role="article" onAnimationEnd={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 205',
      code: '<div role="article" onAnimationIteration={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 206',
      code: '<div role="article" onTransitionEnd={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 0',
      code: '<TestComponent onClick={doFoo} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 1',
      code: '<Button onClick={doFoo} />',
      options: upstreamOptions1,
    },
    ...(
      [
        ['onClick', '() => void 0', ';', 2, 'Image'],
        ['onClick', '() => void 0', '', 4, 'input'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:strict valid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
        };
      }),
    ),
    ...(
      [
        ['onClick', '() => void 0', '', 5, 'button checkbox'],
        ['onClick', '() => void 0', '', 23, 'text'],
        ['onClick', '() => void 0', '', 27, 'hidden'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:strict valid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 31',
      code: '<a onClick={() => void 0} href="http://x.y.z" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 33',
      code: '<area onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 35',
      code: '<button onClick={() => void 0} className="foo" />',
      options: upstreamOptions1,
    },
    ...(
      [
        ['onClick', '() => {}', ';', 36, 'menuitem'],
        ['onClick', '() => {}', ';', 40, 'tr'],
        ['onLoad', '() => {}', ';', 50, 'body'],
        ['onClick', '() => {}', ';', 51, 'canvas'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:strict valid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 62',
      code: '<div onClick={() => void 0} aria-hidden />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 64',
      code: '<div onClick={() => void 0} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 65',
      code: '<div onClick={() => void 0} role={undefined} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 66',
      code: '<div onClick={() => void 0} {...props} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 67',
      code: '<div onClick={null} />;',
      options: upstreamOptions1,
    },
    ...(
      [
        ['onClick', '() => {}', ';', 69, 'embed'],
        ['onClick', '() => {}', ';', 77, 'header'],
        ['onLoad', '() => {}', ';', 80, 'iframe'],
        ['onLoad', '() => {}', ';', 83, 'img'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:strict valid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 84',
      code: '<img src={currentPhoto.imageUrl} onLoad={this.handleImageLoad} alt="for review" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 108',
      code: '<span onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 122',
      code: '<div role="button" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 128',
      code: '<div role="link" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 132',
      code: '<div role="menuitem" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 149',
      code: '<div role="presentation" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 150',
      code: '<div role="command" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 163',
      code: '<div role="widget" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 164',
      code: '<div role="window" onClick={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 165',
      code: '<div role="article" onCopy={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 173',
      code: '<div role="article" onSubmit={() => {}} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict valid 3',
      code: '<Button onClick={() => void 0} />;',
      options: upstreamOptions1,
      settings: upstreamSettings0,
    },
  ],
  invalid: [
    ...(
      [
        ['onClick', '() => {}', ';', 4, 'blockquote'],
        [
          'onClick',
          '() => {}',
          ';',
          6,
          'caption code dd del details dfn dl dir dt em fieldset figcaption figure footer',
        ],
        ['onClick', '() => {}', ';', 22, 'h2 h3 h4 h5 h6 hr html'],
        ['onClick', '() => {}', ';', 31, 'ins label legend'],
        [
          'onClick',
          '() => {}',
          ';',
          35,
          'mark marquee menu meter nav ol optgroup output p pre progress ruby',
        ],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
          errors: [
            singleLineError(
              'noNoninteractiveElementInteractions',
              'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
              openingElement,
              1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 47',
      code: '<section onClick={() => {}} aria-label="Aardvark" />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 48',
      code: '<section onClick={() => {}} aria-labelledby="js_1" />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    ...(
      [
        ['onClick', '() => {}', ';', 49, 'strong sub sup'],
        ['onClick', '() => {}', ';', 53, 'tbody tfoot thead time'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
          errors: [
            singleLineError(
              'noNoninteractiveElementInteractions',
              'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
              openingElement,
              1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 62',
      code: '<div role="alertdialog" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 63',
      code: '<div role="application" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 64',
      code: '<div role="banner" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 65',
      code: '<div role="cell" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 66',
      code: '<div role="complementary" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 67',
      code: '<div role="contentinfo" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 68',
      code: '<div role="definition" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 69',
      code: '<div role="dialog" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 70',
      code: '<div role="directory" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 71',
      code: '<div role="document" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 72',
      code: '<div role="feed" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 73',
      code: '<div role="figure" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 74',
      code: '<div role="form" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 75',
      code: '<div role="group" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 76',
      code: '<div role="heading" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 77',
      code: '<div role="img" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 78',
      code: '<div role="list" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 79',
      code: '<div role="listitem" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 80',
      code: '<div role="log" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 81',
      code: '<div role="main" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 82',
      code: '<div role="marquee" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 83',
      code: '<div role="math" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 84',
      code: '<div role="navigation" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 85',
      code: '<div role="note" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 86',
      code: '<div role="region" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 87',
      code: '<div role="rowgroup" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 88',
      code: '<div role="search" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 89',
      code: '<div role="separator" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 90',
      code: '<div role="status" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 91',
      code: '<div role="table" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 92',
      code: '<div role="tabpanel" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 93',
      code: '<div role="term" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 94',
      code: '<div role="timer" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 95',
      code: '<div role="tooltip" onClick={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 97',
      code: '<div role="article" onKeyPress={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 98',
      code: '<div role="article" onKeyUp={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 102',
      code: '<div role="article" onMouseDown={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:recommended invalid 103',
      code: '<div role="article" onMouseUp={() => {}} />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    ...(
      [
        ['onClick', '() => {}', ';', 4, 'blockquote'],
        [
          'onClick',
          '() => {}',
          ';',
          6,
          'caption code dd del details dfn dl dir dt em fieldset figcaption figure footer',
        ],
        ['onClick', '() => {}', ';', 22, 'h2 h3 h4 h5 h6 hr html'],
        ['onClick', '() => {}', ';', 31, 'ins label legend'],
        [
          'onClick',
          '() => {}',
          ';',
          35,
          'mark marquee menu meter nav ol optgroup output p pre progress ruby',
        ],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:strict invalid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
          errors: [
            singleLineError(
              'noNoninteractiveElementInteractions',
              'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
              openingElement,
              1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 47',
      code: '<section onClick={() => {}} aria-label="Aardvark" />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 48',
      code: '<section onClick={() => {}} aria-labelledby="js_1" />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    ...(
      [
        ['onClick', '() => {}', ';', 49, 'strong sub sup'],
        ['onClick', '() => {}', ';', 53, 'tbody tfoot thead time'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:strict invalid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
          errors: [
            singleLineError(
              'noNoninteractiveElementInteractions',
              'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
              openingElement,
              1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 62',
      code: '<div role="alertdialog" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 63',
      code: '<div role="application" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 64',
      code: '<div role="banner" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 65',
      code: '<div role="cell" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 66',
      code: '<div role="complementary" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 67',
      code: '<div role="contentinfo" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 68',
      code: '<div role="definition" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 69',
      code: '<div role="dialog" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 70',
      code: '<div role="directory" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 71',
      code: '<div role="document" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 72',
      code: '<div role="feed" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 73',
      code: '<div role="figure" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 74',
      code: '<div role="form" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 75',
      code: '<div role="group" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 76',
      code: '<div role="heading" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 77',
      code: '<div role="img" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 78',
      code: '<div role="list" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 79',
      code: '<div role="listitem" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 80',
      code: '<div role="log" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 81',
      code: '<div role="main" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 82',
      code: '<div role="marquee" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 83',
      code: '<div role="math" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 84',
      code: '<div role="navigation" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 85',
      code: '<div role="note" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 86',
      code: '<div role="region" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 87',
      code: '<div role="rowgroup" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 88',
      code: '<div role="search" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 89',
      code: '<div role="separator" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 90',
      code: '<div role="status" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 91',
      code: '<div role="table" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 92',
      code: '<div role="tabpanel" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 93',
      code: '<div role="term" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 94',
      code: '<div role="timer" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 95',
      code: '<div role="tooltip" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 97',
      code: '<div role="article" onKeyPress={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 98',
      code: '<div role="article" onKeyUp={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 102',
      code: '<div role="article" onMouseDown={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 103',
      code: '<div role="article" onMouseUp={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 106',
      code: '<div role="article" onBlur={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 108',
      code: '<div role="article" onDblClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 109',
      code: '<div role="article" onDoubleClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 110',
      code: '<div role="article" onDrag={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 111',
      code: '<div role="article" onDragEnd={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 112',
      code: '<div role="article" onDragEnter={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 113',
      code: '<div role="article" onDragExit={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 114',
      code: '<div role="article" onDragLeave={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 115',
      code: '<div role="article" onDragOver={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 116',
      code: '<div role="article" onDragStart={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 117',
      code: '<div role="article" onDrop={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 119',
      code: '<div role="article" onMouseLeave={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 120',
      code: '<div role="article" onMouseMove={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 121',
      code: '<div role="article" onMouseOut={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 122',
      code: '<div role="article" onMouseOver={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    ...(
      [
        ['onClick', '() => void 0', ';', 0, 'main'],
        ['onClick', '() => {}', ';', 1, 'address article aside'],
        ['onClick', '() => {}', ';', 5, 'br'],
        ['onClick', '() => {}', ';', 20, 'form h1'],
        ['onClick', '() => {}', ';', 29, 'iframe img'],
        ['onClick', '() => {}', ';', 34, 'li'],
        ['onClick', '() => {}', ';', 52, 'table'],
        ['onClick', '() => {}', ';', 57, 'ul'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 no-noninteractive-element-interactions:strict invalid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
          errors: [
            singleLineError(
              'noNoninteractiveElementInteractions',
              'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
              openingElement,
              1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 58',
      code: '<ul contentEditable="false" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 59',
      code: '<article contentEditable onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 60',
      code: '<div contentEditable role="article" onKeyDown={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 61',
      code: '<div role="alert" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 96',
      code: '<div role="article" onKeyDown={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 99',
      code: '<div role="article" onClick={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 100',
      code: '<div role="article" onLoad={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 101',
      code: '<div role="article" onError={() => {}} />;',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-element-interactions:strict invalid 104',
      code: '<Image onClick={() => void 0} />;',
      options: upstreamOptions1,
      settings: upstreamSettings1,
      errors: [
        {
          messageId: 'noNoninteractiveElementInteractions',
          message:
            'Non-interactive elements should not be assigned mouse or keyboard event listeners.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
  ],
});
