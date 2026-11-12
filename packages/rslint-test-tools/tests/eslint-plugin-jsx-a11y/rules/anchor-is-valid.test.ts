import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const components = [{ components: ['Anchor', 'Link'] }];
const specialLink = [{ specialLink: ['hrefLeft', 'hrefRight'] }];
const noHrefAspect = [{ aspects: ['noHref'] }];
const invalidHrefAspect = [{ aspects: ['invalidHref'] }];
const preferButtonAspect = [{ aspects: ['preferButton'] }];
const noHrefInvalidHrefAspect = [{ aspects: ['noHref', 'invalidHref'] }];
const noHrefPreferButtonAspect = [{ aspects: ['noHref', 'preferButton'] }];
const preferButtonInvalidHrefAspect = [
  { aspects: ['preferButton', 'invalidHref'] },
];
const componentsAndSpecialLink = [
  { components: ['Anchor'], specialLink: ['hrefLeft'] },
];
const componentsAndSpecialLinkAndInvalidHrefAspect = [
  {
    components: ['Anchor'],
    specialLink: ['hrefLeft'],
    aspects: ['invalidHref'],
  },
];
const componentsAndSpecialLinkAndNoHrefAspect = [
  { components: ['Anchor'], specialLink: ['hrefLeft'], aspects: ['noHref'] },
];

const componentsSettings = {
  'jsx-a11y': {
    components: {
      Anchor: 'a',
      Link: 'a',
    },
  },
};

new RuleTester(config).run('anchor-is-valid', {
  valid: [
    // ---- DEFAULT ELEMENT 'a' TESTS ----
    { code: '<Anchor />' },
    { code: '<a {...props} />' },
    { code: '<a href="foo" />' },
    { code: '<a href={foo} />' },
    { code: '<a href="/foo" />' },
    { code: '<a href="https://foo.bar.com" />' },
    { code: '<div href="foo" />' },
    { code: '<a href="javascript" />' },
    { code: '<a href="javascriptFoo" />' },
    { code: '<a href={`#foo`}/>' },
    { code: '<a href={"foo"}/>' },
    { code: '<a href={"javascript"}/>' },
    { code: '<a href={`#javascript`}/>' },
    { code: '<a href="#foo" />' },
    { code: '<a href="#javascript" />' },
    { code: '<a href="#javascriptFoo" />' },
    { code: '<UX.Layout>test</UX.Layout>' },
    { code: '<a href={this} />' },

    // ---- CUSTOM ELEMENT TEST FOR ARRAY OPTION ----
    { code: '<Anchor {...props} />', options: components },
    { code: '<Anchor href="foo" />', options: components },
    { code: '<Anchor href={foo} />', options: components },
    { code: '<Anchor href="/foo" />', options: components },
    { code: '<Anchor href="https://foo.bar.com" />', options: components },
    { code: '<div href="foo" />', options: components },
    { code: '<Anchor href={`#foo`}/>', options: components },
    { code: '<Anchor href={"foo"}/>', options: components },
    { code: '<Anchor href="#foo" />', options: components },
    { code: '<Link {...props} />', options: components },
    { code: '<Link href="foo" />', options: components },
    { code: '<Link href={foo} />', options: components },
    { code: '<Link href="/foo" />', options: components },
    { code: '<Link href="https://foo.bar.com" />', options: components },
    { code: '<div href="foo" />', options: components },
    { code: '<Link href={`#foo`}/>', options: components },
    { code: '<Link href={"foo"}/>', options: components },
    { code: '<Link href="#foo" />', options: components },
    { code: '<Link href="#foo" />', settings: componentsSettings },

    // ---- CUSTOM PROP TESTS ----
    { code: '<a {...props} />', options: specialLink },
    { code: '<a hrefLeft="foo" />', options: specialLink },
    { code: '<a hrefLeft={foo} />', options: specialLink },
    { code: '<a hrefLeft="/foo" />', options: specialLink },
    { code: '<a hrefLeft="https://foo.bar.com" />', options: specialLink },
    { code: '<div hrefLeft="foo" />', options: specialLink },
    { code: '<a hrefLeft={`#foo`}/>', options: specialLink },
    { code: '<a hrefLeft={"foo"}/>', options: specialLink },
    { code: '<a hrefLeft="#foo" />', options: specialLink },
    { code: '<UX.Layout>test</UX.Layout>', options: specialLink },
    { code: '<a hrefRight={this} />', options: specialLink },
    { code: '<a hrefRight="foo" />', options: specialLink },
    { code: '<a hrefRight={foo} />', options: specialLink },
    { code: '<a hrefRight="/foo" />', options: specialLink },
    { code: '<a hrefRight="https://foo.bar.com" />', options: specialLink },
    { code: '<div hrefRight="foo" />', options: specialLink },
    { code: '<a hrefRight={`#foo`}/>', options: specialLink },
    { code: '<a hrefRight={"foo"}/>', options: specialLink },
    { code: '<a hrefRight="#foo" />', options: specialLink },

    // ---- CUSTOM BOTH COMPONENTS AND SPECIALLINK TESTS ----
    { code: '<Anchor {...props} />', options: componentsAndSpecialLink },
    { code: '<Anchor hrefLeft="foo" />', options: componentsAndSpecialLink },
    { code: '<Anchor hrefLeft={foo} />', options: componentsAndSpecialLink },
    { code: '<Anchor hrefLeft="/foo" />', options: componentsAndSpecialLink },
    {
      code: '<Anchor hrefLeft="https://foo.bar.com" />',
      options: componentsAndSpecialLink,
    },
    { code: '<div hrefLeft="foo" />', options: componentsAndSpecialLink },
    {
      code: '<Anchor hrefLeft={`#foo`}/>',
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft={"foo"}/>',
      options: componentsAndSpecialLink,
    },
    { code: '<Anchor hrefLeft="#foo" />', options: componentsAndSpecialLink },
    {
      code: '<UX.Layout>test</UX.Layout>',
      options: componentsAndSpecialLink,
    },

    // ---- WITH ONCLICK — DEFAULT ELEMENT 'a' TESTS ----
    { code: '<a {...props} onClick={() => void 0} />' },
    { code: '<a href="foo" onClick={() => void 0} />' },
    { code: '<a href={foo} onClick={() => void 0} />' },
    { code: '<a href="/foo" onClick={() => void 0} />' },
    { code: '<a href="https://foo.bar.com" onClick={() => void 0} />' },
    { code: '<div href="foo" onClick={() => void 0} />' },
    { code: '<a href={`#foo`} onClick={() => void 0} />' },
    { code: '<a href={"foo"} onClick={() => void 0} />' },
    { code: '<a href="#foo" onClick={() => void 0} />' },
    { code: '<a href={this} onClick={() => void 0} />' },

    // ---- WITH ONCLICK — CUSTOM ELEMENT TEST FOR ARRAY OPTION ----
    {
      code: '<Anchor {...props} onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Anchor href="foo" onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Anchor href={foo} onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Anchor href="/foo" onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Anchor href="https://foo.bar.com" onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Anchor href={`#foo`} onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Anchor href={"foo"} onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Anchor href="#foo" onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Link {...props} onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Link href="foo" onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Link href={foo} onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Link href="/foo" onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Link href="https://foo.bar.com" onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<div href="foo" onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Link href={`#foo`} onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Link href={"foo"} onClick={() => void 0} />',
      options: components,
    },
    {
      code: '<Link href="#foo" onClick={() => void 0} />',
      options: components,
    },

    // ---- WITH ONCLICK — CUSTOM PROP TESTS ----
    {
      code: '<a {...props} onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefLeft="foo" onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefLeft={foo} onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefLeft="/foo" onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefLeft href="https://foo.bar.com" onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<div hrefLeft="foo" onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefLeft={`#foo`} onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefLeft={"foo"} onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefLeft="#foo" onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefRight={this} onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefRight="foo" onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefRight={foo} onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefRight="/foo" onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefRight href="https://foo.bar.com" onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<div hrefRight="foo" onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefRight={`#foo`} onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefRight={"foo"} onClick={() => void 0} />',
      options: specialLink,
    },
    {
      code: '<a hrefRight="#foo" onClick={() => void 0} />',
      options: specialLink,
    },

    // ---- WITH ONCLICK — CUSTOM BOTH COMPONENTS AND SPECIALLINK TESTS ----
    {
      code: '<Anchor {...props} onClick={() => void 0} />',
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft="foo" onClick={() => void 0} />',
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft={foo} onClick={() => void 0} />',
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft="/foo" onClick={() => void 0} />',
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft href="https://foo.bar.com" onClick={() => void 0} />',
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft={`#foo`} onClick={() => void 0} />',
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft={"foo"} onClick={() => void 0} />',
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft="#foo" onClick={() => void 0} />',
      options: componentsAndSpecialLink,
    },

    // ---- WITH ASPECTS TESTS — NO HREF ----
    { code: '<a />', options: invalidHrefAspect },
    { code: '<a href={undefined} />', options: invalidHrefAspect },
    { code: '<a href={null} />', options: invalidHrefAspect },
    { code: '<a />', options: preferButtonAspect },
    { code: '<a href={undefined} />', options: preferButtonAspect },
    { code: '<a href={null} />', options: preferButtonAspect },
    { code: '<a />', options: preferButtonInvalidHrefAspect },
    { code: '<a href={undefined} />', options: preferButtonInvalidHrefAspect },
    { code: '<a href={null} />', options: preferButtonInvalidHrefAspect },

    // ---- WITH ASPECTS TESTS — INVALID HREF ----
    { code: '<a href="" />;', options: preferButtonAspect },
    { code: '<a href="#" />', options: preferButtonAspect },
    { code: '<a href={"#"} />', options: preferButtonAspect },
    { code: '<a href="javascript:void(0)" />', options: preferButtonAspect },
    {
      code: '<a href={"javascript:void(0)"} />',
      options: preferButtonAspect,
    },
    { code: '<a href="" />;', options: noHrefAspect },
    { code: '<a href="#" />', options: noHrefAspect },
    { code: '<a href={"#"} />', options: noHrefAspect },
    { code: '<a href="javascript:void(0)" />', options: noHrefAspect },
    { code: '<a href={"javascript:void(0)"} />', options: noHrefAspect },
    { code: '<a href="" />;', options: noHrefPreferButtonAspect },
    { code: '<a href="#" />', options: noHrefPreferButtonAspect },
    { code: '<a href={"#"} />', options: noHrefPreferButtonAspect },
    {
      code: '<a href="javascript:void(0)" />',
      options: noHrefPreferButtonAspect,
    },
    {
      code: '<a href={"javascript:void(0)"} />',
      options: noHrefPreferButtonAspect,
    },

    // ---- WITH ASPECTS TESTS — SHOULD BE BUTTON ----
    {
      code: '<a onClick={() => void 0} />',
      options: invalidHrefAspect,
    },
    {
      code: '<a href="#" onClick={() => void 0} />',
      options: noHrefAspect,
    },
    {
      code: '<a href="javascript:void(0)" onClick={() => void 0} />',
      options: noHrefAspect,
    },
    {
      code: '<a href={"javascript:void(0)"} onClick={() => void 0} />',
      options: noHrefAspect,
    },

    // ---- CUSTOM COMPONENTS AND SPECIAL LINK AND ASPECT ----
    {
      code: '<Anchor hrefLeft={undefined} />',
      options: componentsAndSpecialLinkAndInvalidHrefAspect,
    },
    {
      code: '<Anchor hrefLeft={null} />',
      options: componentsAndSpecialLinkAndInvalidHrefAspect,
    },
  ],
  invalid: [
    // ---- DEFAULT ELEMENT 'a' TESTS — NO HREF ----
    {
      code: '<a />',
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: '<a href={undefined} />',
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<a href={null} />',
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    // ---- DEFAULT ELEMENT 'a' TESTS — INVALID HREF ----
    {
      code: '<a href="" />;',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<a href="#" />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<a href={"#"} />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<a href="javascript:void(0)" />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<a href={"javascript:void(0)"} />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    // ---- DEFAULT ELEMENT 'a' TESTS — SHOULD BE BUTTON ----
    {
      code: '<a onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<a href="#" onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<a href="javascript:void(0)" onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: '<a href={"javascript:void(0)"} onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },

    // ---- CUSTOM ELEMENT TEST FOR ARRAY OPTION — NO HREF ----
    {
      code: '<Link />',
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
      options: components,
    },
    {
      code: '<Link href={undefined} />',
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
      options: components,
    },
    {
      code: '<Link href={null} />',
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
      options: components,
    },
    // ---- CUSTOM ELEMENT TEST FOR ARRAY OPTION — INVALID HREF ----
    {
      code: '<Link href="" />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
      options: components,
    },
    {
      code: '<Link href="#" />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
      options: components,
    },
    {
      code: '<Link href={"#"} />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
      options: components,
    },
    {
      code: '<Link href="javascript:void(0)" />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
      options: components,
    },
    {
      code: '<Link href={"javascript:void(0)"} />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
      options: components,
    },
    {
      code: '<Anchor href="" />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      options: components,
    },
    {
      code: '<Anchor href="#" />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
      options: components,
    },
    {
      code: '<Anchor href={"#"} />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
      options: components,
    },
    {
      code: '<Anchor href="javascript:void(0)" />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
      options: components,
    },
    {
      code: '<Anchor href={"javascript:void(0)"} />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
      options: components,
    },
    // ---- CUSTOM ELEMENT TEST FOR ARRAY OPTION — SHOULD BE BUTTON ----
    {
      code: '<Link onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
      options: components,
    },
    {
      code: '<Link href="#" onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
      options: components,
    },
    {
      code: '<Link href="javascript:void(0)" onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 58,
        },
      ],
      options: components,
    },
    {
      code: '<Link href={"javascript:void(0)"} onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 60,
        },
      ],
      options: components,
    },
    {
      code: '<Anchor onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
      options: components,
    },
    {
      code: '<Anchor href="#" onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
      options: components,
    },
    {
      code: '<Anchor href="javascript:void(0)" onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 60,
        },
      ],
      options: components,
    },
    {
      code: '<Anchor href={"javascript:void(0)"} onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 62,
        },
      ],
      options: components,
    },
    {
      code: '<Link href="#" onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
      settings: componentsSettings,
    },

    // ---- CUSTOM PROP TESTS — NO HREF ----
    {
      code: '<a hrefLeft={undefined} />',
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
      options: specialLink,
    },
    {
      code: '<a hrefLeft={null} />',
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
      options: specialLink,
    },
    // ---- CUSTOM PROP TESTS — INVALID HREF ----
    {
      code: '<a hrefLeft="" />;',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
      options: specialLink,
    },
    {
      code: '<a hrefLeft="#" />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      options: specialLink,
    },
    {
      code: '<a hrefLeft={"#"} />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
      options: specialLink,
    },
    {
      code: '<a hrefLeft="javascript:void(0)" />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
      options: specialLink,
    },
    {
      code: '<a hrefLeft={"javascript:void(0)"} />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
      options: specialLink,
    },
    // ---- CUSTOM PROP TESTS — SHOULD BE BUTTON ----
    {
      code: '<a hrefLeft="#" onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
      options: specialLink,
    },
    {
      code: '<a hrefLeft="javascript:void(0)" onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 59,
        },
      ],
      options: specialLink,
    },
    {
      code: '<a hrefLeft={"javascript:void(0)"} onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 61,
        },
      ],
      options: specialLink,
    },

    // ---- CUSTOM BOTH COMPONENTS AND SPECIAL LINK TESTS — NO HREF ----
    {
      code: '<Anchor Anchor={undefined} />',
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft={null} />',
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
      options: componentsAndSpecialLink,
    },
    // ---- CUSTOM BOTH COMPONENTS AND SPECIAL LINK TESTS — INVALID HREF ----
    {
      code: '<Anchor hrefLeft="" />;',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft="#" />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft={"#"} />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft="javascript:void(0)" />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft={"javascript:void(0)"} />',
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
      options: componentsAndSpecialLink,
    },
    // ---- CUSTOM BOTH COMPONENTS AND SPECIAL LINK TESTS — SHOULD BE BUTTON ----
    {
      code: '<Anchor hrefLeft="#" onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft="javascript:void(0)" onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 64,
        },
      ],
      options: componentsAndSpecialLink,
    },
    {
      code: '<Anchor hrefLeft={"javascript:void(0)"} onClick={() => void 0} />',
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 66,
        },
      ],
      options: componentsAndSpecialLink,
    },

    // ---- WITH ASPECTS TESTS — NO HREF ----
    {
      code: '<a />',
      options: noHrefAspect,
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: '<a />',
      options: noHrefPreferButtonAspect,
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: '<a />',
      options: noHrefInvalidHrefAspect,
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: '<a href={undefined} />',
      options: noHrefAspect,
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<a href={undefined} />',
      options: noHrefPreferButtonAspect,
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<a href={undefined} />',
      options: noHrefInvalidHrefAspect,
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<a href={null} />',
      options: noHrefAspect,
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<a href={null} />',
      options: noHrefPreferButtonAspect,
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<a href={null} />',
      options: noHrefInvalidHrefAspect,
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },

    // ---- WITH ASPECTS TESTS — INVALID HREF ----
    {
      code: '<a href="" />;',
      options: invalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<a href="" />;',
      options: noHrefInvalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<a href="" />;',
      options: preferButtonInvalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<a href="#" />;',
      options: invalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<a href="#" />;',
      options: noHrefInvalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<a href="#" />;',
      options: preferButtonInvalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<a href={"#"} />;',
      options: invalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<a href={"#"} />;',
      options: noHrefInvalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<a href={"#"} />;',
      options: preferButtonInvalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<a href="javascript:void(0)" />;',
      options: invalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<a href="javascript:void(0)" />;',
      options: noHrefInvalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<a href="javascript:void(0)" />;',
      options: preferButtonInvalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<a href={"javascript:void(0)"} />;',
      options: invalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<a href={"javascript:void(0)"} />;',
      options: noHrefInvalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<a href={"javascript:void(0)"} />;',
      options: preferButtonInvalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },

    // ---- WITH ASPECTS TESTS — SHOULD BE BUTTON ----
    {
      code: '<a onClick={() => void 0} />',
      options: preferButtonAspect,
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<a onClick={() => void 0} />',
      options: preferButtonInvalidHrefAspect,
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<a onClick={() => void 0} />',
      options: noHrefPreferButtonAspect,
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<a onClick={() => void 0} />',
      options: noHrefAspect,
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<a onClick={() => void 0} />',
      options: noHrefInvalidHrefAspect,
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<a href="#" onClick={() => void 0} />',
      options: preferButtonAspect,
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<a href="#" onClick={() => void 0} />',
      options: noHrefPreferButtonAspect,
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<a href="#" onClick={() => void 0} />',
      options: preferButtonInvalidHrefAspect,
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<a href="#" onClick={() => void 0} />',
      options: invalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<a href="#" onClick={() => void 0} />',
      options: noHrefInvalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<a href="javascript:void(0)" onClick={() => void 0} />',
      options: preferButtonAspect,
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: '<a href="javascript:void(0)" onClick={() => void 0} />',
      options: noHrefPreferButtonAspect,
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: '<a href="javascript:void(0)" onClick={() => void 0} />',
      options: preferButtonInvalidHrefAspect,
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: '<a href="javascript:void(0)" onClick={() => void 0} />',
      options: invalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: '<a href="javascript:void(0)" onClick={() => void 0} />',
      options: noHrefInvalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: '<a href={"javascript:void(0)"} onClick={() => void 0} />',
      options: preferButtonAspect,
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: '<a href={"javascript:void(0)"} onClick={() => void 0} />',
      options: noHrefPreferButtonAspect,
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: '<a href={"javascript:void(0)"} onClick={() => void 0} />',
      options: preferButtonInvalidHrefAspect,
      errors: [
        {
          messageId: 'preferButton',
          message:
            'Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: '<a href={"javascript:void(0)"} onClick={() => void 0} />',
      options: invalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: '<a href={"javascript:void(0)"} onClick={() => void 0} />',
      options: noHrefInvalidHrefAspect,
      errors: [
        {
          messageId: 'invalidHref',
          message:
            'The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },

    // ---- CUSTOM COMPONENTS AND SPECIAL LINK AND ASPECT ----
    {
      code: '<Anchor hrefLeft={undefined} />',
      options: componentsAndSpecialLinkAndNoHrefAspect,
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<Anchor hrefLeft={null} />',
      options: componentsAndSpecialLinkAndNoHrefAspect,
      errors: [
        {
          messageId: 'noHref',
          message:
            'The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
  ],
});
