import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
import { diagnostic } from '../expectations';

function expectFragmentsNotSupportedError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'fragmentsNotSupported',
    'Fragments are only supported starting from React v16.2. Please disable the `react/jsx-fragments` rule in `eslint` settings or upgrade your version of React.',
    range,
  );
}

function expectPreferFragmentError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'preferFragment',
    'Prefer fragment shorthand over Act.Frag',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

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

ruleTester.run('jsx-fragments', {
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
    {
      code: 'function f(){const F=Act.Frag;} const F=Widget; <F/>;',
      settings,
    },
    {
      code: 'const F=Act.Frag; for (const F of widgets) { <F/>; }',
      settings,
    },
    {
      code: 'const F=Act.Frag; try {} catch (F) { <F/>; }',
      settings,
    },
    {
      code: 'function render(){ function F(){} var F=Act.Frag; return <F/>; }',
      settings,
    },
    {
      code: 'function render(){ var F=Widget; var F=Act.Frag; return <F/>; }',
      settings,
    },
  ],
  invalid: [
    {
      code: '<><Foo /></>',
      settings: settingsOld,
      errors: [expectFragmentsNotSupportedError([1, 1, 1, 13])],
    },
    {
      code: '<Act.Frag><Foo /></Act.Frag>',
      settings: settingsOld,
      errors: [expectFragmentsNotSupportedError([1, 1, 1, 29])],
    },
    {
      code: '<Act.Frag />',
      settings: settingsOld,
      errors: [expectFragmentsNotSupportedError([1, 1, 1, 13])],
    },
    {
      code: '<><Foo /></>',
      options: ['element'],
      settings,
      errors: [
        diagnostic(
          'preferPragma',
          'Prefer Act.Frag over fragment shorthand',
          [1, 1, 1, 13],
        ),
      ],
      output: `<Act.Frag><Foo /></Act.Frag>`,
    },
    {
      code: '<Act.Frag><Foo /></Act.Frag>',
      options: ['syntax'],
      settings,
      errors: [expectPreferFragmentError([1, 1, 1, 29])],
      output: `<><Foo /></>`,
    },
    {
      code: '<Act.Frag />',
      options: ['syntax'],
      settings,
      errors: [expectPreferFragmentError([1, 1, 1, 13])],
      output: `<></>`,
    },
    {
      code: '/** @jsx Preact.h */ <Preact.Frag/>',
      output: '/** @jsx Preact.h */ <></>',
      settings,
      errors: [
        diagnostic(
          'preferFragment',
          'Prefer fragment shorthand over Preact.Frag',
          [1, 22, 1, 36],
        ),
      ],
    },
    {
      code: `
        import Act, { Frag as F } from 'react';
        <F />;
      `,
      options: ['syntax'],
      settings,
      errors: [expectPreferFragmentError([3, 9, 3, 14])],
      output: `
        import Act, { Frag as F } from 'react';
        <></>;
      `,
    },
    {
      code: `
        import Act, { Frag as F } from 'react';
        <F><Foo /></F>;
      `,
      options: ['syntax'],
      settings,
      errors: [expectPreferFragmentError([3, 9, 3, 23])],
      output: `
        import Act, { Frag as F } from 'react';
        <><Foo /></>;
      `,
    },
    {
      code: `
        import Act, { Frag } from 'react';
        <Frag><Foo /></Frag>;
      `,
      options: ['syntax'],
      settings,
      errors: [expectPreferFragmentError([3, 9, 3, 29])],
      output: `
        import Act, { Frag } from 'react';
        <><Foo /></>;
      `,
    },
    {
      code: `
        const F = Act.Frag;
        <F><Foo /></F>;
      `,
      options: ['syntax'],
      settings,
      errors: [expectPreferFragmentError([3, 9, 3, 23])],
      output: `
        const F = Act.Frag;
        <><Foo /></>;
      `,
    },
    {
      code: `
        const { Frag } = Act;
        <Frag><Foo /></Frag>;
      `,
      options: ['syntax'],
      settings,
      errors: [expectPreferFragmentError([3, 9, 3, 29])],
      output: `
        const { Frag } = Act;
        <><Foo /></>;
      `,
    },
    {
      code: `
        const { Frag } = require('react');
        <Frag><Foo /></Frag>;
      `,
      options: ['syntax'],
      settings,
      errors: [expectPreferFragmentError([3, 9, 3, 29])],
      output: `
        const { Frag } = require('react');
        <><Foo /></>;
      `,
    },
    {
      code: 'function render(){ var F=Act.Frag; function F(){} return <F/>; }',
      output:
        'function render(){ var F=Act.Frag; function F(){} return <></>; }',
      settings,
      errors: [expectPreferFragmentError([1, 58, 1, 62])],
    },
    {
      code: 'function render(){ var F=Act.Frag; var F=Widget; return <F/>; }',
      output:
        'function render(){ var F=Act.Frag; var F=Widget; return <></>; }',
      settings,
      errors: [expectPreferFragmentError([1, 57, 1, 61])],
    },
  ],
});
