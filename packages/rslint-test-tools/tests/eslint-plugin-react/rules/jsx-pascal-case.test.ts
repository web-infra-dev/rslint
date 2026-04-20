import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-pascal-case', {} as never, {
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
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `<Test_component />;`,
      errors: [
        {
          message:
            'Imported JSX component Test_component must be in PascalCase',
        },
      ],
    },
    {
      code: `<TEST_COMPONENT />;`,
      errors: [
        {
          message:
            'Imported JSX component TEST_COMPONENT must be in PascalCase',
        },
      ],
    },
    {
      code: `<YMCA />;`,
      errors: [
        { message: 'Imported JSX component YMCA must be in PascalCase' },
      ],
    },
    {
      code: `<_TEST_COMPONENT />;`,
      options: [{ allowAllCaps: true }],
      errors: [
        {
          message:
            'Imported JSX component _TEST_COMPONENT must be in PascalCase or SCREAMING_SNAKE_CASE',
        },
      ],
    },
    {
      code: `<TEST_COMPONENT_ />;`,
      options: [{ allowAllCaps: true }],
      errors: [
        {
          message:
            'Imported JSX component TEST_COMPONENT_ must be in PascalCase or SCREAMING_SNAKE_CASE',
        },
      ],
    },
    {
      code: `<TEST-COMPONENT />;`,
      options: [{ allowAllCaps: true }],
      errors: [
        {
          message:
            'Imported JSX component TEST-COMPONENT must be in PascalCase or SCREAMING_SNAKE_CASE',
        },
      ],
    },
    {
      code: `<__ />;`,
      options: [{ allowAllCaps: true }],
      errors: [
        {
          message:
            'Imported JSX component __ must be in PascalCase or SCREAMING_SNAKE_CASE',
        },
      ],
    },
    {
      code: `<_div />;`,
      options: [{ allowLeadingUnderscore: true }],
      errors: [
        { message: 'Imported JSX component _div must be in PascalCase' },
      ],
    },
    {
      code: `<__ />;`,
      options: [{ allowAllCaps: true, allowLeadingUnderscore: true }],
      errors: [
        {
          message:
            'Imported JSX component __ must be in PascalCase or SCREAMING_SNAKE_CASE',
        },
      ],
    },
    {
      code: `<$a />;`,
      errors: [{ message: 'Imported JSX component $a must be in PascalCase' }],
    },
    {
      code: `<Foo_DEPRECATED />;`,
      options: [{ ignore: ['*_FOO'] }],
      errors: [
        {
          message:
            'Imported JSX component Foo_DEPRECATED must be in PascalCase',
        },
      ],
    },
    {
      code: `<Styled.h1 />;`,
      errors: [{ message: 'Imported JSX component h1 must be in PascalCase' }],
    },
    {
      code: `<$Typography.P />;`,
      errors: [
        {
          message: 'Imported JSX component $Typography must be in PascalCase',
        },
      ],
    },
    {
      code: `<STYLED.h1 />;`,
      options: [{ allowNamespace: true }],
      errors: [
        { message: 'Imported JSX component STYLED must be in PascalCase' },
      ],
    },

    // ---- Additional edge cases ----
    {
      code: '<TEST_COMPONENT\n  foo="bar"\n/>;',
      errors: [
        {
          message:
            'Imported JSX component TEST_COMPONENT must be in PascalCase',
        },
      ],
    },
    {
      code: `<TEST_COMPONENT>hi</TEST_COMPONENT>;`,
      errors: [
        {
          message:
            'Imported JSX component TEST_COMPONENT must be in PascalCase',
        },
      ],
    },
    {
      code: `<_test />;`,
      options: [{ allowLeadingUnderscore: true }],
      errors: [
        { message: 'Imported JSX component _test must be in PascalCase' },
      ],
    },
    {
      code: `<BAD_NAME />;`,
      options: [{ ignore: ['GOOD_NAME'] }],
      errors: [
        { message: 'Imported JSX component BAD_NAME must be in PascalCase' },
      ],
    },
    {
      code: `<Foo.bar />;`,
      errors: [{ message: 'Imported JSX component bar must be in PascalCase' }],
    },
    {
      code: `<Foo.bad_name.Baz />;`,
      errors: [
        { message: 'Imported JSX component bad_name must be in PascalCase' },
      ],
    },
    {
      code: `<Foo.Bar.baz />;`,
      errors: [{ message: 'Imported JSX component baz must be in PascalCase' }],
    },
    {
      code: `<div><FOO_A /><FOO_B /></div>;`,
      errors: [
        { message: 'Imported JSX component FOO_A must be in PascalCase' },
        { message: 'Imported JSX component FOO_B must be in PascalCase' },
      ],
    },
    {
      code: `<Wrapper>{true && <Bad_Inner />}</Wrapper>;`,
      errors: [
        { message: 'Imported JSX component Bad_Inner must be in PascalCase' },
      ],
    },
    {
      code: `<Wrapper prop={<Bad_Attr />} />;`,
      errors: [
        { message: 'Imported JSX component Bad_Attr must be in PascalCase' },
      ],
    },
    {
      code: `<$Weird />;`,
      options: [{ allowAllCaps: true }],
      errors: [
        {
          message:
            'Imported JSX component $Weird must be in PascalCase or SCREAMING_SNAKE_CASE',
        },
      ],
    },
    {
      code: `<Modal:bad_header />;`,
      errors: [
        {
          message: 'Imported JSX component bad_header must be in PascalCase',
        },
      ],
    },
    {
      code: `<BAD_NS:Header />;`,
      errors: [
        { message: 'Imported JSX component BAD_NS must be in PascalCase' },
      ],
    },
    {
      code: `<Modal:bad_Name />;`,
      options: [{ allowAllCaps: true }],
      errors: [
        {
          message:
            'Imported JSX component bad_Name must be in PascalCase or SCREAMING_SNAKE_CASE',
        },
      ],
    },
    {
      code: `<BAD_NS:Header />;`,
      options: [{ allowNamespace: true }],
      errors: [
        { message: 'Imported JSX component BAD_NS must be in PascalCase' },
      ],
    },
    {
      code: `<Foo.$Bar />;`,
      errors: [
        { message: 'Imported JSX component $Bar must be in PascalCase' },
      ],
    },
    {
      code: `<Bad_Name<string> />;`,
      errors: [
        { message: 'Imported JSX component Bad_Name must be in PascalCase' },
      ],
    },
    {
      code: `<Año_BAD />;`,
      options: [{ ignore: ['Año_DIFFERENT'] }],
      errors: [
        { message: 'Imported JSX component Año_BAD must be in PascalCase' },
      ],
    },
  ],
});
