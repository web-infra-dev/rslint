import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectUsePascalCaseError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'usePascalCase',
    'Imported JSX component TEST_COMPONENT must be in PascalCase',
    range,
  );
}

function expectUsePascalOrSnakeCaseError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'usePascalOrSnakeCase',
    'Imported JSX component __ must be in PascalCase or SCREAMING_SNAKE_CASE',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-pascal-case', {
  valid: [
    // ---- Upstream valid cases ----
    { code: `<testcomponent />;` },
    { code: `<testComponent />;` },
    { code: `<test_component />;` },
    { code: `<TestComponent />;` },
    { code: `<CSSTransitionGroup />;` },
    { code: `<BetterThanCSS />;` },
    { code: `<TestComponent><div /></TestComponent>;` },
    { code: `<Test1Component />;` },
    { code: `<TestComponent1 />;` },
    { code: `<T3StComp0Nent />;` },
    { code: `<Éurströmming />;` },
    { code: `<Año />;` },
    { code: `<Søknad />;` },
    { code: `<T />;` },
    { code: `<YMCA />;`, options: [{ allowAllCaps: true }] },
    { code: `<TEST_COMPONENT />;`, options: [{ allowAllCaps: true }] },
    { code: `<Modal.Header />;` },
    { code: `<qualification.T3StComp0Nent />;` },
    { code: `<IGNORED />;`, options: [{ ignore: ['IGNORED'] }] },
    { code: `<Foo_DEPRECATED />;`, options: [{ ignore: ['*_D*D'] }] },
    {
      code: `<Foo_DEPRECATED />;`,
      options: [{ ignore: ['*_+(DEPRECATED|IGNORED)'] }],
    },
    { code: `<$ />;` },
    { code: `<_ />;` },
    { code: `<H1>Hello!</H1>;` },
    { code: `<Typography.P />;` },
    { code: `<Styled.h1 />;`, options: [{ allowNamespace: true }] },
    {
      code: `<_TEST_COMPONENT />;`,
      options: [{ allowAllCaps: true, allowLeadingUnderscore: true }],
    },
    {
      code: `<_TestComponent />;`,
      options: [{ allowLeadingUnderscore: true }],
    },

    // ---- Additional edge cases ----
    { code: `<Wrapper><span /></Wrapper>;` },
    { code: `<Foo.T />;` },
    { code: `<this.Foo />;` },
    { code: `<Foo.Bar.Baz />;` },
    { code: `<Foo.Bar.Baz.Qux />;` },
    { code: `<Foo.T.Bar />;` },
    { code: `<Foo.Bar.T />;` },
    { code: `<Foo.BAR />;`, options: [{ allowAllCaps: true }] },
    {
      code: `<FOO.h1 />;`,
      options: [{ allowAllCaps: true, allowNamespace: true }],
    },
    {
      code: `<_SAFE />;`,
      options: [
        {
          allowAllCaps: true,
          allowLeadingUnderscore: true,
          allowNamespace: true,
          ignore: ['Never'],
        },
      ],
    },
    { code: `<div><Wrapper /><span /><OtherOK /></div>;` },
    { code: `<Outer>{true && <Inner />}</Outer>;` },
    { code: `<Outer prop={<Inner />} />;` },
    { code: `<><Foo /><Bar /></>;` },
    { code: `const cond = true; <Outer>{cond ? <Yes /> : <No />}</Outer>;` },
    { code: `<AB />;`, options: [{ ignore: ['[A-Z][A-Z]'] }] },
    { code: `<Foo_BAR />;`, options: [{ ignore: ['Foo_@(BAR|BAZ)'] }] },
    { code: `<A_B />;`, options: [{ ignore: ['A_?'] }] },
    { code: `<Modal:Header />;` },
    { code: `<svg:path />;` },
    { code: `<Modal:NAMED />;`, options: [{ allowAllCaps: true }] },
    { code: `<Modal:_Header />;`, options: [{ allowLeadingUnderscore: true }] },
    { code: `<Styled:h1 />;`, options: [{ allowNamespace: true }] },
    {
      code: `<Modal:Foo_DEPRECATED />;`,
      options: [{ ignore: ['*_DEPRECATED'] }],
    },
    { code: `<Año_Old />;`, options: [{ ignore: ['Año_*'] }] },
    { code: `<Año_KEEP />;`, options: [{ ignore: ['Año_KE?P'] }] },
    { code: `<Good />;`, options: [{ ignore: ['[invalid'] }] },
    { code: `<Good />;`, options: [{ ignore: ['[^]'] }] },
    { code: `<Foo<string> />;` },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `<testcomponent />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `<testComponent />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `<test_component />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `<TestComponent />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `<CSSTransitionGroup />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `<BetterThanCSS />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `<TestComponent><div /></TestComponent>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `<Test1Component />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `<TestComponent1 />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `<T3StComp0Nent />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `<Éurströmming />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `<Año />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `<Søknad />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `<T />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `<YMCA />`,
      options: [{ allowAllCaps: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `<TEST_COMPONENT />`,
      options: [{ allowAllCaps: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `<Modal.Header />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `<qualification.T3StComp0Nent />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `<Modal:Header />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `<IGNORED />`,
      options: [{ ignore: ['IGNORED'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `<Foo_DEPRECATED />`,
      options: [{ ignore: ['*_D*D'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `<Foo_DEPRECATED />`,
      options: [{ ignore: ['*_+(DEPRECATED|IGNORED)'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `<$ />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `<_ />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `<H1>Hello!</H1>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `<Typography.P />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `<Styled.h1 />`,
      options: [{ allowNamespace: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `<_TEST_COMPONENT />`,
      options: [{ allowAllCaps: true, allowLeadingUnderscore: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `<_TestComponent />`,
      options: [{ allowLeadingUnderscore: true }],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `<Test_component />;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component Test_component must be in PascalCase',
          [1, 1, 1, 19],
        ),
      ],
    },
    {
      code: `<TEST_COMPONENT />;`,
      errors: [expectUsePascalCaseError([1, 1, 1, 19])],
    },
    {
      code: `<YMCA />;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component YMCA must be in PascalCase',
          [1, 1, 1, 9],
        ),
      ],
    },
    {
      code: `<_TEST_COMPONENT />;`,
      options: [{ allowAllCaps: true }],
      errors: [
        diagnostic(
          'usePascalOrSnakeCase',
          'Imported JSX component _TEST_COMPONENT must be in PascalCase or SCREAMING_SNAKE_CASE',
          [1, 1, 1, 20],
        ),
      ],
    },
    {
      code: `<TEST_COMPONENT_ />;`,
      options: [{ allowAllCaps: true }],
      errors: [
        diagnostic(
          'usePascalOrSnakeCase',
          'Imported JSX component TEST_COMPONENT_ must be in PascalCase or SCREAMING_SNAKE_CASE',
          [1, 1, 1, 20],
        ),
      ],
    },
    {
      code: `<TEST-COMPONENT />;`,
      options: [{ allowAllCaps: true }],
      errors: [
        diagnostic(
          'usePascalOrSnakeCase',
          'Imported JSX component TEST-COMPONENT must be in PascalCase or SCREAMING_SNAKE_CASE',
          [1, 1, 1, 19],
        ),
      ],
    },
    {
      code: `<__ />;`,
      options: [{ allowAllCaps: true }],
      errors: [expectUsePascalOrSnakeCaseError([1, 1, 1, 7])],
    },
    {
      code: `<_div />;`,
      options: [{ allowLeadingUnderscore: true }],
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component _div must be in PascalCase',
          [1, 1, 1, 9],
        ),
      ],
    },
    {
      code: `<__ />;`,
      options: [{ allowAllCaps: true, allowLeadingUnderscore: true }],
      errors: [expectUsePascalOrSnakeCaseError([1, 1, 1, 7])],
    },
    {
      code: `<$a />;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component $a must be in PascalCase',
          [1, 1, 1, 7],
        ),
      ],
    },
    {
      code: `<Foo_DEPRECATED />;`,
      options: [{ ignore: ['*_FOO'] }],
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component Foo_DEPRECATED must be in PascalCase',
          [1, 1, 1, 19],
        ),
      ],
    },
    {
      code: `<Styled.h1 />;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component h1 must be in PascalCase',
          [1, 1, 1, 14],
        ),
      ],
    },
    {
      code: `<$Typography.P />;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component $Typography must be in PascalCase',
          [1, 1, 1, 18],
        ),
      ],
    },
    {
      code: `<STYLED.h1 />;`,
      options: [{ allowNamespace: true }],
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component STYLED must be in PascalCase',
          [1, 1, 1, 14],
        ),
      ],
    },

    // ---- Additional edge cases ----
    {
      code: '<TEST_COMPONENT\n  foo="bar"\n/>;',
      errors: [expectUsePascalCaseError([1, 1, 3, 3])],
    },
    {
      code: `<TEST_COMPONENT>hi</TEST_COMPONENT>;`,
      errors: [expectUsePascalCaseError([1, 1, 1, 17])],
    },
    {
      code: `<_test />;`,
      options: [{ allowLeadingUnderscore: true }],
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component _test must be in PascalCase',
          [1, 1, 1, 10],
        ),
      ],
    },
    {
      code: `<BAD_NAME />;`,
      options: [{ ignore: ['GOOD_NAME'] }],
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component BAD_NAME must be in PascalCase',
          [1, 1, 1, 13],
        ),
      ],
    },
    {
      code: `<Foo.bar />;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component bar must be in PascalCase',
          [1, 1, 1, 12],
        ),
      ],
    },
    {
      code: `<Foo.bad_name.Baz />;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component bad_name must be in PascalCase',
          [1, 1, 1, 21],
        ),
      ],
    },
    {
      code: `<Foo.Bar.baz />;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component baz must be in PascalCase',
          [1, 1, 1, 16],
        ),
      ],
    },
    {
      code: `<div><FOO_A /><FOO_B /></div>;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component FOO_A must be in PascalCase',
          [1, 6, 1, 15],
        ),
        diagnostic(
          'usePascalCase',
          'Imported JSX component FOO_B must be in PascalCase',
          [1, 15, 1, 24],
        ),
      ],
    },
    {
      code: `<Wrapper>{true && <Bad_Inner />}</Wrapper>;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component Bad_Inner must be in PascalCase',
          [1, 19, 1, 32],
        ),
      ],
    },
    {
      code: `<Wrapper prop={<Bad_Attr />} />;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component Bad_Attr must be in PascalCase',
          [1, 16, 1, 28],
        ),
      ],
    },
    {
      code: `<$Weird />;`,
      options: [{ allowAllCaps: true }],
      errors: [
        diagnostic(
          'usePascalOrSnakeCase',
          'Imported JSX component $Weird must be in PascalCase or SCREAMING_SNAKE_CASE',
          [1, 1, 1, 11],
        ),
      ],
    },
    {
      code: `<Modal:bad_header />;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component bad_header must be in PascalCase',
          [1, 1, 1, 21],
        ),
      ],
    },
    {
      code: `<BAD_NS:Header />;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component BAD_NS must be in PascalCase',
          [1, 1, 1, 18],
        ),
      ],
    },
    {
      code: `<Modal:bad_Name />;`,
      options: [{ allowAllCaps: true }],
      errors: [
        diagnostic(
          'usePascalOrSnakeCase',
          'Imported JSX component bad_Name must be in PascalCase or SCREAMING_SNAKE_CASE',
          [1, 1, 1, 19],
        ),
      ],
    },
    {
      code: `<BAD_NS:Header />;`,
      options: [{ allowNamespace: true }],
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component BAD_NS must be in PascalCase',
          [1, 1, 1, 18],
        ),
      ],
    },
    {
      code: `<Foo.$Bar />;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component $Bar must be in PascalCase',
          [1, 1, 1, 13],
        ),
      ],
    },
    {
      code: `<Bad_Name<string> />;`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component Bad_Name must be in PascalCase',
          [1, 1, 1, 21],
        ),
      ],
    },
    {
      code: `<Año_BAD />;`,
      options: [{ ignore: ['Año_DIFFERENT'] }],
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component Año_BAD must be in PascalCase',
          [1, 1, 1, 12],
        ),
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `<Test_component />`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component Test_component must be in PascalCase',
          [1, 1, 1, 19],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `<TEST_COMPONENT />`,
      errors: [expectUsePascalCaseError([1, 1, 1, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `<YMCA />`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component YMCA must be in PascalCase',
          [1, 1, 1, 9],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `<_TEST_COMPONENT />`,
      options: [{ allowAllCaps: true }],
      errors: [
        diagnostic(
          'usePascalOrSnakeCase',
          'Imported JSX component _TEST_COMPONENT must be in PascalCase or SCREAMING_SNAKE_CASE',
          [1, 1, 1, 20],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `<TEST_COMPONENT_ />`,
      options: [{ allowAllCaps: true }],
      errors: [
        diagnostic(
          'usePascalOrSnakeCase',
          'Imported JSX component TEST_COMPONENT_ must be in PascalCase or SCREAMING_SNAKE_CASE',
          [1, 1, 1, 20],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `<TEST-COMPONENT />`,
      options: [{ allowAllCaps: true }],
      errors: [
        diagnostic(
          'usePascalOrSnakeCase',
          'Imported JSX component TEST-COMPONENT must be in PascalCase or SCREAMING_SNAKE_CASE',
          [1, 1, 1, 19],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `<__ />`,
      options: [{ allowAllCaps: true }],
      errors: [expectUsePascalOrSnakeCaseError([1, 1, 1, 7])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `<_div />`,
      options: [{ allowLeadingUnderscore: true }],
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component _div must be in PascalCase',
          [1, 1, 1, 9],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `<__ />`,
      options: [{ allowAllCaps: true, allowLeadingUnderscore: true }],
      errors: [expectUsePascalOrSnakeCaseError([1, 1, 1, 7])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `<$a />`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component $a must be in PascalCase',
          [1, 1, 1, 7],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `<Foo_DEPRECATED />`,
      options: [{ ignore: ['*_FOO'] }],
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component Foo_DEPRECATED must be in PascalCase',
          [1, 1, 1, 19],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `<Styled.h1 />`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component h1 must be in PascalCase',
          [1, 1, 1, 14],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `<$Typography.P />`,
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component $Typography must be in PascalCase',
          [1, 1, 1, 18],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `<STYLED.h1 />`,
      options: [{ allowNamespace: true }],
      errors: [
        diagnostic(
          'usePascalCase',
          'Imported JSX component STYLED must be in PascalCase',
          [1, 1, 1, 14],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
