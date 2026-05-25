import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-equals-spacing', null as never, {
  valid: [
    // default (=== never)
    { code: `<App />` },
    { code: `<App foo />` },
    { code: `<App foo="bar" />` },
    { code: `<App foo={e => bar(e)} />` },
    { code: `<App {...props} />` },
    // never
    { code: `<App />`, options: ['never'] },
    { code: `<App foo />`, options: ['never'] },
    { code: `<App foo="bar" />`, options: ['never'] },
    { code: `<App foo={e => bar(e)} />`, options: ['never'] },
    { code: `<App {...props} />`, options: ['never'] },
    // always
    { code: `<App />`, options: ['always'] },
    { code: `<App foo />`, options: ['always'] },
    { code: `<App foo = "bar" />`, options: ['always'] },
    { code: `<App foo = {e => bar(e)} />`, options: ['always'] },
    { code: `<App {...props} />`, options: ['always'] },
  ],

  invalid: [
    // default: space on both sides
    {
      code: `<App foo = {bar} />`,
      output: `<App foo={bar} />`,
      errors: [
        {
          messageId: 'noSpaceBefore',
          message: `There should be no space before '='`,
        },
        {
          messageId: 'noSpaceAfter',
          message: `There should be no space after '='`,
        },
      ],
    },
    // never: space on both sides
    {
      code: `<App foo = {bar} />`,
      options: ['never'],
      output: `<App foo={bar} />`,
      errors: [
        {
          messageId: 'noSpaceBefore',
          message: `There should be no space before '='`,
        },
        {
          messageId: 'noSpaceAfter',
          message: `There should be no space after '='`,
        },
      ],
    },
    // never: space before only
    {
      code: `<App foo ={bar} />`,
      options: ['never'],
      output: `<App foo={bar} />`,
      errors: [
        {
          messageId: 'noSpaceBefore',
          message: `There should be no space before '='`,
        },
      ],
    },
    // never: space after only
    {
      code: `<App foo= {bar} />`,
      options: ['never'],
      output: `<App foo={bar} />`,
      errors: [
        {
          messageId: 'noSpaceAfter',
          message: `There should be no space after '='`,
        },
      ],
    },
    // never: two attributes, mixed
    {
      code: `<App foo= {bar} bar = {baz} />`,
      options: ['never'],
      output: `<App foo={bar} bar={baz} />`,
      errors: [
        {
          messageId: 'noSpaceAfter',
          message: `There should be no space after '='`,
        },
        {
          messageId: 'noSpaceBefore',
          message: `There should be no space before '='`,
        },
        {
          messageId: 'noSpaceAfter',
          message: `There should be no space after '='`,
        },
      ],
    },
    // always: no space on either side
    {
      code: `<App foo={bar} />`,
      options: ['always'],
      output: `<App foo = {bar} />`,
      errors: [
        {
          messageId: 'needSpaceBefore',
          message: `A space is required before '='`,
        },
        {
          messageId: 'needSpaceAfter',
          message: `A space is required after '='`,
        },
      ],
    },
    // always: space before only (missing after)
    {
      code: `<App foo ={bar} />`,
      options: ['always'],
      output: `<App foo = {bar} />`,
      errors: [
        {
          messageId: 'needSpaceAfter',
          message: `A space is required after '='`,
        },
      ],
    },
    // always: space after only (missing before)
    {
      code: `<App foo= {bar} />`,
      options: ['always'],
      output: `<App foo = {bar} />`,
      errors: [
        {
          messageId: 'needSpaceBefore',
          message: `A space is required before '='`,
        },
      ],
    },
    // always: two attributes, mixed
    {
      code: `<App foo={bar} bar ={baz} />`,
      options: ['always'],
      output: `<App foo = {bar} bar = {baz} />`,
      errors: [
        {
          messageId: 'needSpaceBefore',
          message: `A space is required before '='`,
        },
        {
          messageId: 'needSpaceAfter',
          message: `A space is required after '='`,
        },
        {
          messageId: 'needSpaceAfter',
          message: `A space is required after '='`,
        },
      ],
    },
  ],
});
