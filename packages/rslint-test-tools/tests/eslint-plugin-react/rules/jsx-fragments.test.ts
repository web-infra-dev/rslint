import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const settings = {
  react: {
    version: '16.2',
    pragma: 'Act',
    fragment: 'Frag',
  },
};

const settingsOld = {
  react: {
    version: '16.1',
    pragma: 'Act',
    fragment: 'Frag',
  },
};

ruleTester.run('jsx-fragments', {} as never, {
  valid: [
    {
      code: '<><Foo /></>',
      settings,
    },
    {
      code: '<Act.Frag><Foo /></Act.Frag>',
      options: ['element'],
      settings,
    },
    {
      code: '<Act.Frag />',
      options: ['element'],
      settings,
    },
    {
      code: `
        import Act, { Frag as F } from 'react';
        <F><Foo /></F>;
      `,
      options: ['element'],
      settings,
    },
    {
      code: `
        const F = Act.Frag;
        <F><Foo /></F>;
      `,
      options: ['element'],
      settings,
    },
    {
      code: `
        const { Frag } = Act;
        <Frag><Foo /></Frag>;
      `,
      options: ['element'],
      settings,
    },
    {
      code: `
        const { Frag } = require('react');
        <Frag><Foo /></Frag>;
      `,
      options: ['element'],
      settings,
    },
    {
      code: '<Act.Frag key="key"><Foo /></Act.Frag>',
      options: ['syntax'],
      settings,
    },
    {
      code: '<Act.Frag key="key" />',
      options: ['syntax'],
      settings,
    },
  ],
  invalid: [
    {
      code: '<><Foo /></>',
      settings: settingsOld,
      errors: [{ messageId: 'fragmentsNotSupported' }],
    },
    {
      code: '<Act.Frag><Foo /></Act.Frag>',
      settings: settingsOld,
      errors: [{ messageId: 'fragmentsNotSupported' }],
    },
    {
      code: '<Act.Frag />',
      settings: settingsOld,
      errors: [{ messageId: 'fragmentsNotSupported' }],
    },
    {
      code: '<><Foo /></>',
      options: ['element'],
      settings,
      errors: [{ messageId: 'preferPragma' }],
    },
    {
      code: '<Act.Frag><Foo /></Act.Frag>',
      options: ['syntax'],
      settings,
      errors: [{ messageId: 'preferFragment' }],
    },
    {
      code: '<Act.Frag />',
      options: ['syntax'],
      settings,
      errors: [{ messageId: 'preferFragment' }],
    },
    {
      code: `
        import Act, { Frag as F } from 'react';
        <F />;
      `,
      options: ['syntax'],
      settings,
      errors: [{ messageId: 'preferFragment' }],
    },
    {
      code: `
        import Act, { Frag as F } from 'react';
        <F><Foo /></F>;
      `,
      options: ['syntax'],
      settings,
      errors: [{ messageId: 'preferFragment' }],
    },
    {
      code: `
        import Act, { Frag } from 'react';
        <Frag><Foo /></Frag>;
      `,
      options: ['syntax'],
      settings,
      errors: [{ messageId: 'preferFragment' }],
    },
    {
      code: `
        const F = Act.Frag;
        <F><Foo /></F>;
      `,
      options: ['syntax'],
      settings,
      errors: [{ messageId: 'preferFragment' }],
    },
    {
      code: `
        const { Frag } = Act;
        <Frag><Foo /></Frag>;
      `,
      options: ['syntax'],
      settings,
      errors: [{ messageId: 'preferFragment' }],
    },
    {
      code: `
        const { Frag } = require('react');
        <Frag><Foo /></Frag>;
      `,
      options: ['syntax'],
      settings,
      errors: [{ messageId: 'preferFragment' }],
    },
  ],
});
