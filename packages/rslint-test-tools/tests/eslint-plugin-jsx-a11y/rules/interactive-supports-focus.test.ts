import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';
import { singleLineError } from '../matrix';

const componentsSettings = {
  'jsx-a11y': {
    components: {
      Div: 'div',
    },
  },
};

const polymorphicSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
  },
};

const recommendedOptions = [
  {
    tabbable: [
      'button',
      'checkbox',
      'link',
      'searchbox',
      'spinbutton',
      'switch',
      'textbox',
    ],
  },
];

const strictOptions = [
  {
    tabbable: [
      'button',
      'checkbox',
      'link',
      'progressbar',
      'searchbox',
      'slider',
      'spinbutton',
      'switch',
      'textbox',
    ],
  },
];

new RuleTester(config).run('interactive-supports-focus', {
  valid: [
    // ============================================================
    // Upstream alwaysValid sample (mirrors the file's `alwaysValid` array)
    // ============================================================
    { code: '<div />' },
    { code: '<div aria-hidden onClick={() => void 0} />' },
    { code: '<div onClick={() => void 0} />;' },
    { code: '<div onClick={() => void 0} tabIndex={undefined} />;' },
    { code: '<div onClick={() => void 0} tabIndex="bad" />;' },
    { code: '<div onClick={() => void 0} role={undefined} />;' },
    { code: '<div role="section" onClick={() => void 0} />' },
    { code: '<div onClick={() => void 0} {...props} />;' },
    { code: '<input type="text" onClick={() => void 0} />' },
    { code: '<button onClick={() => void 0} className="foo" />' },
    { code: '<select onClick={() => void 0} className="foo" />' },
    { code: '<textarea onClick={() => void 0} className="foo" />' },
    { code: '<a onClick={() => void 0} />' },
    { code: '<a tabIndex="0" onClick={() => void 0} />' },
    { code: '<a tabIndex={0} onClick={() => void 0} />' },
    { code: '<a role="button" href="#" onClick={() => void 0} />' },
    { code: '<a onClick={() => void 0} href="http://x.y.z" />' },
    { code: '<a onClick={() => void 0} href="http://x.y.z" role="button" />' },
    { code: '<TestComponent onClick={doFoo} />' },
    { code: '<input onClick={() => void 0} type="hidden" />;' },
    { code: '<section onClick={() => void 0} />;' },
    { code: '<main onClick={() => void 0} />;' },
    { code: '<header onClick={() => void 0} />;' },
    {
      code: '<div role="textbox" aria-disabled="true" onClick={() => void 0} />',
    },
    { code: '<div role="button" tabIndex="0" onClick={() => void 0} />' },
    { code: '<div role="menuitem" tabIndex="0" onClick={() => void 0} />' },
    { code: '<div role="link" tabIndex="0" onClick={() => void 0} />' },

    // ---- componentsSetting maps <Div> → div; <Div role="button" tabIndex="0">
    //      is already focusable.
    {
      code: '<Div onClick={() => void 0} role="button" tabIndex="0" />',
      settings: componentsSettings,
    },

    // ============================================================
    // Recommended config (default tabbable list)
    // ============================================================
    // ---- gridcell is interactive but NOT in recommended.tabbable; adding
    //      tabIndex="0" satisfies the rule under recommended.
    {
      code: '<div role="gridcell" tabIndex="0" onClick={() => void 0} />',
      options: recommendedOptions,
    },
    // ---- onFocus is not a triggering handler — rule short-circuits.
    {
      code: '<div role="button" onFocus={() => void 0} />',
      options: recommendedOptions,
    },

    // ============================================================
    // Strict config (broader tabbable list)
    // ============================================================
    // ---- progressbar is in strict.tabbable; tabIndex="0" works.
    {
      code: '<div role="progressbar" tabIndex="0" onClick={() => void 0} />',
      options: strictOptions,
    },

    // ============================================================
    // Polymorphic — `as="button"` resolves the JSX component to <button>
    // which is inherently focusable.
    // ============================================================
    {
      code: '<Foo as="button" onClick={() => void 0} />',
      settings: polymorphicSettings,
    },
  ],
  invalid: [
    // ============================================================
    // Default (no options) — tabbable list is empty so everything that
    // qualifies receives the "focusable" message with two suggestions.
    // ============================================================
    {
      code: '<div role="button" onClick={() => void 0} />',
      errors: [
        {
          messageId: 'focusable',
          message:
            "Elements with the 'button' interactive role must be focusable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<div tabIndex={0} role="button" onClick={() => void 0} />',
            },
            {
              messageId: 'tabIndexNegOne',
              desc: 'Add `tabIndex={-1}` to make the element focusable but not reachable via sequential keyboard navigation.',
              output:
                '<div tabIndex={-1} role="button" onClick={() => void 0} />',
            },
          ],
        },
      ],
    },
    {
      code: '<div role="checkbox" onMouseDown={() => void 0} />',
      errors: [
        {
          messageId: 'focusable',
          message:
            "Elements with the 'checkbox' interactive role must be focusable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<div tabIndex={0} role="checkbox" onMouseDown={() => void 0} />',
            },
            {
              messageId: 'tabIndexNegOne',
              desc: 'Add `tabIndex={-1}` to make the element focusable but not reachable via sequential keyboard navigation.',
              output:
                '<div tabIndex={-1} role="checkbox" onMouseDown={() => void 0} />',
            },
          ],
        },
      ],
    },
    {
      code: '<div role="slider" onKeyDown={() => void 0} />',
      errors: [
        {
          messageId: 'focusable',
          message:
            "Elements with the 'slider' interactive role must be focusable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<div tabIndex={0} role="slider" onKeyDown={() => void 0} />',
            },
            {
              messageId: 'tabIndexNegOne',
              desc: 'Add `tabIndex={-1}` to make the element focusable but not reachable via sequential keyboard navigation.',
              output:
                '<div tabIndex={-1} role="slider" onKeyDown={() => void 0} />',
            },
          ],
        },
      ],
    },
    // ---- componentsSetting → role="button" + onClick.
    {
      code: '<Div onClick={() => void 0} role="button" />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'focusable',
          message:
            "Elements with the 'button' interactive role must be focusable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<Div tabIndex={0} onClick={() => void 0} role="button" />',
            },
            {
              messageId: 'tabIndexNegOne',
              desc: 'Add `tabIndex={-1}` to make the element focusable but not reachable via sequential keyboard navigation.',
              output:
                '<Div tabIndex={-1} onClick={() => void 0} role="button" />',
            },
          ],
        },
      ],
    },

    // ============================================================
    // Recommended config — button is in tabbable → "must be tabbable"
    // ============================================================
    {
      code: '<div role="button" onClick={() => void 0} />',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'tabbable',
          message:
            "Elements with the 'button' interactive role must be tabbable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<div tabIndex={0} role="button" onClick={() => void 0} />',
            },
          ],
        },
      ],
    },
    {
      code: '<div role="checkbox" onClick={() => void 0} />',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'tabbable',
          message:
            "Elements with the 'checkbox' interactive role must be tabbable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<div tabIndex={0} role="checkbox" onClick={() => void 0} />',
            },
          ],
        },
      ],
    },
    // ---- gridcell is NOT in recommended.tabbable → "must be focusable"
    {
      code: '<div role="gridcell" onClick={() => void 0} />',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'focusable',
          message:
            "Elements with the 'gridcell' interactive role must be focusable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<div tabIndex={0} role="gridcell" onClick={() => void 0} />',
            },
            {
              messageId: 'tabIndexNegOne',
              desc: 'Add `tabIndex={-1}` to make the element focusable but not reachable via sequential keyboard navigation.',
              output:
                '<div tabIndex={-1} role="gridcell" onClick={() => void 0} />',
            },
          ],
        },
      ],
    },
    {
      code: '<div role="menuitem" onClick={() => void 0} />',
      options: recommendedOptions,
      errors: [
        {
          messageId: 'focusable',
          message:
            "Elements with the 'menuitem' interactive role must be focusable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<div tabIndex={0} role="menuitem" onClick={() => void 0} />',
            },
            {
              messageId: 'tabIndexNegOne',
              desc: 'Add `tabIndex={-1}` to make the element focusable but not reachable via sequential keyboard navigation.',
              output:
                '<div tabIndex={-1} role="menuitem" onClick={() => void 0} />',
            },
          ],
        },
      ],
    },

    // ============================================================
    // Strict config — slider / progressbar are in tabbable
    // ============================================================
    {
      code: '<div role="slider" onClick={() => void 0} />',
      options: strictOptions,
      errors: [
        {
          messageId: 'tabbable',
          message:
            "Elements with the 'slider' interactive role must be tabbable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<div tabIndex={0} role="slider" onClick={() => void 0} />',
            },
          ],
        },
      ],
    },
    {
      code: '<div role="progressbar" onClick={() => void 0} />',
      options: strictOptions,
      errors: [
        {
          messageId: 'tabbable',
          message:
            "Elements with the 'progressbar' interactive role must be tabbable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<div tabIndex={0} role="progressbar" onClick={() => void 0} />',
            },
          ],
        },
      ],
    },
    // ---- gridcell still gets "focusable" under strict.
    {
      code: '<div role="gridcell" onClick={() => void 0} />',
      options: strictOptions,
      errors: [
        {
          messageId: 'focusable',
          message:
            "Elements with the 'gridcell' interactive role must be focusable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<div tabIndex={0} role="gridcell" onClick={() => void 0} />',
            },
            {
              messageId: 'tabIndexNegOne',
              desc: 'Add `tabIndex={-1}` to make the element focusable but not reachable via sequential keyboard navigation.',
              output:
                '<div tabIndex={-1} role="gridcell" onClick={() => void 0} />',
            },
          ],
        },
      ],
    },

    // ============================================================
    // role first valid is interactive — multi-role string variants.
    // ============================================================
    {
      code: '<div role="button heading" onClick={() => void 0} />',
      errors: [
        {
          messageId: 'focusable',
          message:
            "Elements with the 'button heading' interactive role must be focusable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<div tabIndex={0} role="button heading" onClick={() => void 0} />',
            },
            {
              messageId: 'tabIndexNegOne',
              desc: 'Add `tabIndex={-1}` to make the element focusable but not reachable via sequential keyboard navigation.',
              output:
                '<div tabIndex={-1} role="button heading" onClick={() => void 0} />',
            },
          ],
        },
      ],
    },

    // ============================================================
    // Polymorphic — `as="div"` resolves to a plain div, which still trips.
    // ============================================================
    {
      code: '<Foo as="div" role="button" onClick={() => void 0} />',
      settings: polymorphicSettings,
      errors: [
        {
          messageId: 'focusable',
          message:
            "Elements with the 'button' interactive role must be focusable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 54,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<Foo tabIndex={0} as="div" role="button" onClick={() => void 0} />',
            },
            {
              messageId: 'tabIndexNegOne',
              desc: 'Add `tabIndex={-1}` to make the element focusable but not reachable via sequential keyboard navigation.',
              output:
                '<Foo tabIndex={-1} as="div" role="button" onClick={() => void 0} />',
            },
          ],
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
    tabbable: [
      'button',
      'checkbox',
      'link',
      'searchbox',
      'spinbutton',
      'switch',
      'textbox',
    ],
  },
];
const upstreamOptions1 = [
  {
    tabbable: [
      'button',
      'checkbox',
      'link',
      'progressbar',
      'searchbox',
      'slider',
      'spinbutton',
      'switch',
      'textbox',
    ],
  },
];
const upstreamSettings0 = { 'jsx-a11y': { components: { Div: 'div' } } };

new RuleTester(config).run('interactive-supports-focus', {
  valid: [
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 2',
      code: '<div aria-hidden={true == true} onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 3',
      code: '<div aria-hidden={true === true} onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 4',
      code: '<div aria-hidden={hidden !== false} onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 5',
      code: '<div aria-hidden={hidden != false} onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 6',
      code: '<div aria-hidden={1 < 2} onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 7',
      code: '<div aria-hidden={1 <= 2} onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 8',
      code: '<div aria-hidden={2 > 1} onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 9',
      code: '<div aria-hidden={2 >= 1} onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 15',
      code: '<div onClick={() => void 0} aria-hidden={false} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 18',
      code: '<input type="hidden" onClick={() => void 0} tabIndex="-1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 19',
      code: '<input type="hidden" onClick={() => void 0} tabIndex={-1} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 20',
      code: '<input onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 21',
      code: '<input onClick={() => void 0} role="combobox" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 23',
      code: '<option onClick={() => void 0} className="foo" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 25',
      code: '<area href="#" onClick={() => void 0} className="foo" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 26',
      code: '<area onClick={() => void 0} className="foo" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 27',
      code: '<summary onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 29',
      code: '<a onClick="showNextPage();">Next page</a>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 30',
      code: '<a onClick="showNextPage();" tabIndex={undefined}>Next page</a>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 31',
      code: '<a onClick="showNextPage();" tabIndex="bad">Next page</a>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 34',
      code: '<a tabIndex={dynamicTabIndex} onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 38',
      code: '<a onClick={() => void 0} href="http://x.y.z" tabIndex="0" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 39',
      code: '<a onClick={() => void 0} href="http://x.y.z" tabIndex={0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 43',
      code: '<span onClick="submitForm();">Submit</span>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 44',
      code: '<span onClick="submitForm();" tabIndex={undefined}>Submit</span>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 45',
      code: '<span onClick="submitForm();" tabIndex="bad">Submit</span>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 46',
      code: '<span onClick="doSomething();" tabIndex="0">Click me!</span>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 47',
      code: '<span onClick="doSomething();" tabIndex={0}>Click me!</span>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 48',
      code: '<span onClick="doSomething();" tabIndex="-1">Click me too!</span>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 49',
      code: '<a href="javascript:void(0);" onClick="doSomething();">Click ALL the things!</a>',
      options: upstreamOptions0,
    },
    ...(
      [
        ['onClick', '() => void 0', ';', 52, 'article'],
        ['onClick', '() => void 0', ';', 54, 'footer'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 interactive-supports-focus:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 56',
      code: '<div role="checkbox" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 59',
      code: '<div role="menuitemcheckbox" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 60',
      code: '<div role="menuitemradio" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 61',
      code: '<div role="option" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 62',
      code: '<div role="radio" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 63',
      code: '<div role="spinbutton" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 64',
      code: '<div role="switch" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 65',
      code: '<div role="tablist" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 66',
      code: '<div role="tab" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 67',
      code: '<div role="textbox" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 69',
      code: '<Foo.Bar onClick={() => void 0} aria-hidden={false} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 70',
      code: '<Input onClick={() => void 0} type="hidden" />;',
      options: upstreamOptions0,
    },
    ...(
      [
        [
          'button',
          '',
          72,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate',
        ],
        [
          'button',
          '',
          79,
          'onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'button',
          '',
          115,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'checkbox',
          '',
          119,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'checkbox',
          '',
          162,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'link',
          '',
          166,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'link',
          '',
          209,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'gridcell',
          '',
          213,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'gridcell',
          '',
          256,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'menuitem',
          '',
          260,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'menuitem',
          '',
          303,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'menuitemcheckbox',
          '',
          307,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'menuitemcheckbox',
          '',
          350,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'menuitemradio',
          '',
          354,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'menuitemradio',
          '',
          397,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'option',
          '',
          401,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'option',
          '',
          444,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'radio',
          '',
          448,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'radio',
          '',
          491,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'searchbox',
          '',
          495,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'searchbox',
          '',
          538,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'slider',
          '',
          542,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'slider',
          '',
          585,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'spinbutton',
          '',
          589,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'spinbutton',
          '',
          632,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'switch',
          '',
          636,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'switch',
          '',
          679,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'tab',
          '',
          683,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'tab',
          '',
          726,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'textbox',
          '',
          730,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'textbox',
          '',
          773,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'treeitem',
          '',
          777,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'treeitem',
          '',
          820,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'gridcell',
          ' tabIndex="0"',
          824,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'menuitem',
          ' tabIndex="0"',
          846,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'menuitemcheckbox',
          ' tabIndex="0"',
          868,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'menuitemradio',
          ' tabIndex="0"',
          890,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'option',
          ' tabIndex="0"',
          912,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'radio',
          ' tabIndex="0"',
          934,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'slider',
          ' tabIndex="0"',
          956,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'tab',
          ' tabIndex="0"',
          978,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'treeitem',
          ' tabIndex="0"',
          1000,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
      ] as const
    ).flatMap(([role, tabIndex, firstCase, values]) =>
      values.split(' ').map((handler, index) => {
        const openingElement = `<div role="${role}" ${handler}={() => void 0}${tabIndex} />`;
        const code = openingElement;
        return {
          name: `upstream v6.10.2 interactive-supports-focus:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 0',
      code: '<div />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 1',
      code: '<div aria-hidden onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 10',
      code: '<div onClick={() => void 0} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 11',
      code: '<div onClick={() => void 0} tabIndex={undefined} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 12',
      code: '<div onClick={() => void 0} tabIndex="bad" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 13',
      code: '<div onClick={() => void 0} role={undefined} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 14',
      code: '<div role="section" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 16',
      code: '<div onClick={() => void 0} {...props} />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 17',
      code: '<input type="text" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 22',
      code: '<button onClick={() => void 0} className="foo" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 24',
      code: '<select onClick={() => void 0} className="foo" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 28',
      code: '<textarea onClick={() => void 0} className="foo" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 32',
      code: '<a onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 33',
      code: '<a tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 35',
      code: '<a tabIndex={0} onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 36',
      code: '<a role="button" href="#" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 37',
      code: '<a onClick={() => void 0} href="http://x.y.z" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 40',
      code: '<a onClick={() => void 0} href="http://x.y.z" role="button" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 41',
      code: '<TestComponent onClick={doFoo} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 42',
      code: '<input onClick={() => void 0} type="hidden" />;',
      options: upstreamOptions0,
    },
    ...(
      [
        ['onClick', '() => void 0', ';', 50, 'section main'],
        ['onClick', '() => void 0', ';', 53, 'header'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 interactive-supports-focus:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 55',
      code: '<div role="button" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 57',
      code: '<div role="link" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 58',
      code: '<div role="menuitem" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 68',
      code: '<div role="textbox" aria-disabled="true" onClick={() => void 0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 2',
      code: '<div aria-hidden={true == true} onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 3',
      code: '<div aria-hidden={true === true} onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 4',
      code: '<div aria-hidden={hidden !== false} onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 5',
      code: '<div aria-hidden={hidden != false} onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 6',
      code: '<div aria-hidden={1 < 2} onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 7',
      code: '<div aria-hidden={1 <= 2} onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 8',
      code: '<div aria-hidden={2 > 1} onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 9',
      code: '<div aria-hidden={2 >= 1} onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 15',
      code: '<div onClick={() => void 0} aria-hidden={false} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 18',
      code: '<input type="hidden" onClick={() => void 0} tabIndex="-1" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 19',
      code: '<input type="hidden" onClick={() => void 0} tabIndex={-1} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 20',
      code: '<input onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 21',
      code: '<input onClick={() => void 0} role="combobox" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 23',
      code: '<option onClick={() => void 0} className="foo" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 25',
      code: '<area href="#" onClick={() => void 0} className="foo" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 26',
      code: '<area onClick={() => void 0} className="foo" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 27',
      code: '<summary onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 29',
      code: '<a onClick="showNextPage();">Next page</a>',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 30',
      code: '<a onClick="showNextPage();" tabIndex={undefined}>Next page</a>',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 31',
      code: '<a onClick="showNextPage();" tabIndex="bad">Next page</a>',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 34',
      code: '<a tabIndex={dynamicTabIndex} onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 38',
      code: '<a onClick={() => void 0} href="http://x.y.z" tabIndex="0" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 39',
      code: '<a onClick={() => void 0} href="http://x.y.z" tabIndex={0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 43',
      code: '<span onClick="submitForm();">Submit</span>',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 44',
      code: '<span onClick="submitForm();" tabIndex={undefined}>Submit</span>',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 45',
      code: '<span onClick="submitForm();" tabIndex="bad">Submit</span>',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 46',
      code: '<span onClick="doSomething();" tabIndex="0">Click me!</span>',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 47',
      code: '<span onClick="doSomething();" tabIndex={0}>Click me!</span>',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 48',
      code: '<span onClick="doSomething();" tabIndex="-1">Click me too!</span>',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 49',
      code: '<a href="javascript:void(0);" onClick="doSomething();">Click ALL the things!</a>',
      options: upstreamOptions1,
    },
    ...(
      [
        ['onClick', '() => void 0', ';', 52, 'article'],
        ['onClick', '() => void 0', ';', 54, 'footer'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 interactive-supports-focus:strict valid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 56',
      code: '<div role="checkbox" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 59',
      code: '<div role="menuitemcheckbox" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 60',
      code: '<div role="menuitemradio" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 61',
      code: '<div role="option" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 62',
      code: '<div role="radio" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 63',
      code: '<div role="spinbutton" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 64',
      code: '<div role="switch" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 65',
      code: '<div role="tablist" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 66',
      code: '<div role="tab" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 67',
      code: '<div role="textbox" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 69',
      code: '<Foo.Bar onClick={() => void 0} aria-hidden={false} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 70',
      code: '<Input onClick={() => void 0} type="hidden" />;',
      options: upstreamOptions1,
    },
    ...(
      [
        [
          'button',
          '',
          72,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate',
        ],
        [
          'button',
          '',
          79,
          'onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'button',
          '',
          115,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'checkbox',
          '',
          119,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'checkbox',
          '',
          162,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'link',
          '',
          166,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'link',
          '',
          209,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'gridcell',
          '',
          213,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'gridcell',
          '',
          256,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'menuitem',
          '',
          260,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'menuitem',
          '',
          303,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'menuitemcheckbox',
          '',
          307,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'menuitemcheckbox',
          '',
          350,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'menuitemradio',
          '',
          354,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'menuitemradio',
          '',
          397,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'option',
          '',
          401,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'option',
          '',
          444,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'radio',
          '',
          448,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'radio',
          '',
          491,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'searchbox',
          '',
          495,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'searchbox',
          '',
          538,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'slider',
          '',
          542,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'slider',
          '',
          585,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'spinbutton',
          '',
          589,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'spinbutton',
          '',
          632,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'switch',
          '',
          636,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'switch',
          '',
          679,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'tab',
          '',
          683,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'tab',
          '',
          726,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'textbox',
          '',
          730,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'textbox',
          '',
          773,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'treeitem',
          '',
          777,
          'onCopy onCut onPaste onCompositionEnd onCompositionStart onCompositionUpdate onFocus onBlur onChange onInput onSubmit onSelect onTouchCancel onTouchEnd onTouchMove onTouchStart onScroll onWheel onAbort onCanPlay onCanPlayThrough onDurationChange onEmptied onEncrypted onEnded onError onLoadedData onLoadedMetadata onLoadStart onPause onPlay onPlaying onProgress onRateChange onSeeked onSeeking onStalled onSuspend onTimeUpdate onVolumeChange onWaiting onLoad',
        ],
        [
          'treeitem',
          '',
          820,
          'onAnimationStart onAnimationEnd onAnimationIteration onTransitionEnd',
        ],
        [
          'gridcell',
          ' tabIndex="0"',
          824,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'menuitem',
          ' tabIndex="0"',
          846,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'menuitemcheckbox',
          ' tabIndex="0"',
          868,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'menuitemradio',
          ' tabIndex="0"',
          890,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'option',
          ' tabIndex="0"',
          912,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'radio',
          ' tabIndex="0"',
          934,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'tab',
          ' tabIndex="0"',
          956,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        [
          'treeitem',
          ' tabIndex="0"',
          978,
          'onKeyDown onKeyPress onKeyUp onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
      ] as const
    ).flatMap(([role, tabIndex, firstCase, values]) =>
      values.split(' ').map((handler, index) => {
        const openingElement = `<div role="${role}" ${handler}={() => void 0}${tabIndex} />`;
        const code = openingElement;
        return {
          name: `upstream v6.10.2 interactive-supports-focus:strict valid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 0',
      code: '<div />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 1',
      code: '<div aria-hidden onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 10',
      code: '<div onClick={() => void 0} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 11',
      code: '<div onClick={() => void 0} tabIndex={undefined} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 12',
      code: '<div onClick={() => void 0} tabIndex="bad" />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 13',
      code: '<div onClick={() => void 0} role={undefined} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 14',
      code: '<div role="section" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 16',
      code: '<div onClick={() => void 0} {...props} />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 17',
      code: '<input type="text" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 22',
      code: '<button onClick={() => void 0} className="foo" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 24',
      code: '<select onClick={() => void 0} className="foo" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 28',
      code: '<textarea onClick={() => void 0} className="foo" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 32',
      code: '<a onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 33',
      code: '<a tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 35',
      code: '<a tabIndex={0} onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 36',
      code: '<a role="button" href="#" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 37',
      code: '<a onClick={() => void 0} href="http://x.y.z" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 40',
      code: '<a onClick={() => void 0} href="http://x.y.z" role="button" />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 41',
      code: '<TestComponent onClick={doFoo} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 42',
      code: '<input onClick={() => void 0} type="hidden" />;',
      options: upstreamOptions1,
    },
    ...(
      [
        ['onClick', '() => void 0', ';', 50, 'section main'],
        ['onClick', '() => void 0', ';', 53, 'header'],
      ] as const
    ).flatMap(([handler, body, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} ${handler}={${body}} />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 interactive-supports-focus:strict valid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 55',
      code: '<div role="button" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 57',
      code: '<div role="link" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 58',
      code: '<div role="menuitem" tabIndex="0" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 68',
      code: '<div role="textbox" aria-disabled="true" onClick={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 78',
      code: '<div role="button" onFocus={() => void 0} />',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended valid 71',
      code: '<Div onClick={() => void 0} role="button" tabIndex="0" />',
      options: upstreamOptions0,
      settings: upstreamSettings0,
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict valid 71',
      code: '<Div onClick={() => void 0} role="button" tabIndex="0" />',
      options: upstreamOptions1,
      settings: upstreamSettings0,
    },
  ],
  invalid: [
    ...(
      [
        [
          'button',
          '',
          2,
          'onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'checkbox',
          '',
          24,
          'onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop',
        ],
        [
          'checkbox',
          '',
          36,
          'onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'link',
          '',
          45,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'searchbox',
          '',
          67,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'spinbutton',
          '',
          89,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'switch',
          '',
          111,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'textbox',
          '',
          133,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
      ] as const
    ).flatMap(([role, tabIndex, firstCase, values]) =>
      values.split(' ').map((handler, index) => {
        const openingElement = `<div role="${role}" ${handler}={() => void 0}${tabIndex} />`;
        const code = openingElement;
        return {
          name: `upstream v6.10.2 interactive-supports-focus:recommended invalid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
          errors: [
            {
              ...singleLineError(
                'tabbable',
                `Elements with the '${role}' interactive role must be tabbable.`,
                openingElement,
                1,
              ),
              suggestions: [
                {
                  messageId: 'tabIndexZero',
                  desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
                  output: code.replace('<div', '<div tabIndex={0}'),
                },
              ],
            },
          ],
        };
      }),
    ),
    ...(
      [
        [
          'gridcell',
          '',
          156,
          'onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'menuitem',
          '',
          178,
          'onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'menuitemcheckbox',
          '',
          199,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'menuitemradio',
          '',
          221,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'option',
          '',
          243,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'radio',
          '',
          265,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'slider',
          '',
          288,
          'onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        ['slider', '', 307, 'onKeyPress onKeyUp'],
        [
          'tab',
          '',
          309,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'treeitem',
          '',
          331,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
      ] as const
    ).flatMap(([role, tabIndex, firstCase, values]) =>
      values.split(' ').map((handler, index) => {
        const openingElement = `<div role="${role}" ${handler}={() => void 0}${tabIndex} />`;
        const code = openingElement;
        return {
          name: `upstream v6.10.2 interactive-supports-focus:recommended invalid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
          errors: [
            {
              ...singleLineError(
                'focusable',
                `Elements with the '${role}' interactive role must be focusable.`,
                openingElement,
                1,
              ),
              suggestions: [
                {
                  messageId: 'tabIndexZero',
                  desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
                  output: code.replace('<div', '<div tabIndex={0}'),
                },
                {
                  messageId: 'tabIndexNegOne',
                  desc: 'Add `tabIndex={-1}` to make the element focusable but not reachable via sequential keyboard navigation.',
                  output: code.replace('<div', '<div tabIndex={-1}'),
                },
              ],
            },
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended invalid 35',
      code: '<div role="checkbox" onMouseDown={() => void 0} />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'tabbable',
          message:
            "Elements with the 'checkbox' interactive role must be tabbable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<div tabIndex={0} role="checkbox" onMouseDown={() => void 0} />',
            },
          ],
        },
      ],
    },
    ...(
      [
        ['slider', '', 287, 'onClick'],
        ['slider', '', 306, 'onKeyDown'],
      ] as const
    ).flatMap(([role, tabIndex, firstCase, values]) =>
      values.split(' ').map((handler, index) => {
        const openingElement = `<div role="${role}" ${handler}={() => void 0}${tabIndex} />`;
        const code = openingElement;
        return {
          name: `upstream v6.10.2 interactive-supports-focus:recommended invalid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
          errors: [
            {
              ...singleLineError(
                'focusable',
                `Elements with the '${role}' interactive role must be focusable.`,
                openingElement,
                1,
              ),
              suggestions: [
                {
                  messageId: 'tabIndexZero',
                  desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
                  output: code.replace('<div', '<div tabIndex={0}'),
                },
                {
                  messageId: 'tabIndexNegOne',
                  desc: 'Add `tabIndex={-1}` to make the element focusable but not reachable via sequential keyboard navigation.',
                  output: code.replace('<div', '<div tabIndex={-1}'),
                },
              ],
            },
          ],
        };
      }),
    ),
    ...(
      [
        [
          'button',
          '',
          2,
          'onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'checkbox',
          '',
          24,
          'onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop',
        ],
        [
          'checkbox',
          '',
          36,
          'onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'link',
          '',
          45,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'progressbar',
          '',
          68,
          'onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'searchbox',
          '',
          89,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'slider',
          '',
          112,
          'onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp',
        ],
        ['slider', '', 131, 'onKeyPress onKeyUp'],
        [
          'spinbutton',
          '',
          133,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'switch',
          '',
          155,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'textbox',
          '',
          177,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
      ] as const
    ).flatMap(([role, tabIndex, firstCase, values]) =>
      values.split(' ').map((handler, index) => {
        const openingElement = `<div role="${role}" ${handler}={() => void 0}${tabIndex} />`;
        const code = openingElement;
        return {
          name: `upstream v6.10.2 interactive-supports-focus:strict invalid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
          errors: [
            {
              ...singleLineError(
                'tabbable',
                `Elements with the '${role}' interactive role must be tabbable.`,
                openingElement,
                1,
              ),
              suggestions: [
                {
                  messageId: 'tabIndexZero',
                  desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
                  output: code.replace('<div', '<div tabIndex={0}'),
                },
              ],
            },
          ],
        };
      }),
    ),
    ...(
      [
        [
          'gridcell',
          '',
          200,
          'onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'menuitem',
          '',
          222,
          'onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'menuitemcheckbox',
          '',
          243,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'menuitemradio',
          '',
          265,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'option',
          '',
          287,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'radio',
          '',
          309,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'tab',
          '',
          331,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
        [
          'treeitem',
          '',
          353,
          'onClick onContextMenu onDblClick onDoubleClick onDrag onDragEnd onDragEnter onDragExit onDragLeave onDragOver onDragStart onDrop onMouseDown onMouseEnter onMouseLeave onMouseMove onMouseOut onMouseOver onMouseUp onKeyDown onKeyPress onKeyUp',
        ],
      ] as const
    ).flatMap(([role, tabIndex, firstCase, values]) =>
      values.split(' ').map((handler, index) => {
        const openingElement = `<div role="${role}" ${handler}={() => void 0}${tabIndex} />`;
        const code = openingElement;
        return {
          name: `upstream v6.10.2 interactive-supports-focus:strict invalid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
          errors: [
            {
              ...singleLineError(
                'focusable',
                `Elements with the '${role}' interactive role must be focusable.`,
                openingElement,
                1,
              ),
              suggestions: [
                {
                  messageId: 'tabIndexZero',
                  desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
                  output: code.replace('<div', '<div tabIndex={0}'),
                },
                {
                  messageId: 'tabIndexNegOne',
                  desc: 'Add `tabIndex={-1}` to make the element focusable but not reachable via sequential keyboard navigation.',
                  output: code.replace('<div', '<div tabIndex={-1}'),
                },
              ],
            },
          ],
        };
      }),
    ),
    ...(
      [
        ['button', '', 1, 'onClick'],
        ['checkbox', '', 23, 'onClick'],
        ['checkbox', '', 35, 'onMouseDown'],
        ['slider', '', 130, 'onKeyDown'],
      ] as const
    ).flatMap(([role, tabIndex, firstCase, values]) =>
      values.split(' ').map((handler, index) => {
        const openingElement = `<div role="${role}" ${handler}={() => void 0}${tabIndex} />`;
        const code = openingElement;
        return {
          name: `upstream v6.10.2 interactive-supports-focus:strict invalid ${firstCase + index}`,
          code,
          options: upstreamOptions1,
          errors: [
            {
              ...singleLineError(
                'tabbable',
                `Elements with the '${role}' interactive role must be tabbable.`,
                openingElement,
                1,
              ),
              suggestions: [
                {
                  messageId: 'tabIndexZero',
                  desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
                  output: code.replace('<div', '<div tabIndex={0}'),
                },
              ],
            },
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict invalid 221',
      code: '<div role="menuitem" onClick={() => void 0} />',
      options: upstreamOptions1,
      errors: [
        {
          messageId: 'focusable',
          message:
            "Elements with the 'menuitem' interactive role must be focusable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<div tabIndex={0} role="menuitem" onClick={() => void 0} />',
            },
            {
              messageId: 'tabIndexNegOne',
              desc: 'Add `tabIndex={-1}` to make the element focusable but not reachable via sequential keyboard navigation.',
              output:
                '<div tabIndex={-1} role="menuitem" onClick={() => void 0} />',
            },
          ],
        },
      ],
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:recommended invalid 0',
      code: '<Div onClick={() => void 0} role="button" />',
      options: upstreamOptions0,
      settings: upstreamSettings0,
      errors: [
        {
          messageId: 'tabbable',
          message:
            "Elements with the 'button' interactive role must be tabbable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<Div tabIndex={0} onClick={() => void 0} role="button" />',
            },
          ],
        },
      ],
    },
    {
      name: 'upstream v6.10.2 interactive-supports-focus:strict invalid 0',
      code: '<Div onClick={() => void 0} role="button" />',
      options: upstreamOptions1,
      settings: upstreamSettings0,
      errors: [
        {
          messageId: 'tabbable',
          message:
            "Elements with the 'button' interactive role must be tabbable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
          suggestions: [
            {
              messageId: 'tabIndexZero',
              desc: 'Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation.',
              output:
                '<Div tabIndex={0} onClick={() => void 0} role="button" />',
            },
          ],
        },
      ],
    },
  ],
});
