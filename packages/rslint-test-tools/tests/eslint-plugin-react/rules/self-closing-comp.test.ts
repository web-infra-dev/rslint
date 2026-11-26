import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectNotSelfClosingError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'notSelfClosing',
    message: 'Empty components are self-closing',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('self-closing-comp', {
  valid: [
    {
      code: `<Foo>bar</Foo>`,
    },
    {
      code: `<div>children</div>`,
    },
    {
      code: `<Foo />`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `var HelloJohn = <Hello name="John" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `var HelloJohn = <Hello.Compound name="John" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `var Profile = <Hello name="John"><img src="picture.png" /></Hello>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `var Profile = <Hello.Compound name="John"><img src="picture.png" /></Hello.Compound>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        <Hello>
          <Hello name="John" />
        </Hello>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        <Hello.Compound>
          <Hello.Compound name="John" />
        </Hello.Compound>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `var HelloJohn = <Hello name="John"> </Hello>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `var HelloJohn = <Hello.Compound name="John"> </Hello.Compound>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `var HelloJohn = <Hello name="John">        </Hello>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `var HelloJohn = <Hello.Compound name="John">        </Hello.Compound>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `var HelloJohn = <div>&nbsp;</div>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `var HelloJohn = <div>{' '}</div>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `var HelloJohn = <Hello name="John">&nbsp;</Hello>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `var HelloJohn = <Hello.Compound name="John">&nbsp;</Hello.Compound>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `var HelloJohn = <div> </div>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `var HelloJohn = <div>        </div>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `var HelloJohn = <Hello name="John"></Hello>;`,
      options: [{ component: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `var HelloJohn = <Hello.Compound name="John"></Hello.Compound>;`,
      options: [{ component: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `var HelloJohn = <Hello name="John">
</Hello>;`,
      options: [{ component: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `var HelloJohn = <Hello.Compound name="John">
</Hello.Compound>;`,
      options: [{ component: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `var HelloJohn = <Hello name="John"> </Hello>;`,
      options: [{ component: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `var HelloJohn = <Hello.Compound name="John"> </Hello.Compound>;`,
      options: [{ component: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `var contentContainer = <div className="content" />;`,
      options: [{ html: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `var contentContainer = <div className="content"><img src="picture.png" /></div>;`,
      options: [{ html: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
        <div>
          <div className="content" />
        </div>
      `,
      options: [{ html: true }],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `<Foo></Foo>`,
      errors: [expectNotSelfClosingError([1, 1, 1, 12])],
      output: `<Foo />`,
    },
    {
      code: `<div></div>`,
      errors: [expectNotSelfClosingError([1, 1, 1, 12])],
      output: `<div />`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `var contentContainer = <div className="content"></div>;`,
      errors: [expectNotSelfClosingError([1, 24, 1, 55])],
      output: `var contentContainer = <div className="content" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `var HelloJohn = <Hello name="John"></Hello>;`,
      errors: [expectNotSelfClosingError([1, 17, 1, 44])],
      output: `var HelloJohn = <Hello name="John" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `var CompoundHelloJohn = <Hello.Compound name="John"></Hello.Compound>;`,
      errors: [expectNotSelfClosingError([1, 25, 1, 70])],
      output: `var CompoundHelloJohn = <Hello.Compound name="John" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `var HelloJohn = <Hello name="John">
</Hello>;`,
      errors: [expectNotSelfClosingError([1, 17, 2, 9])],
      output: `var HelloJohn = <Hello name="John" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `var HelloJohn = <Hello.Compound name="John">
</Hello.Compound>;`,
      errors: [expectNotSelfClosingError([1, 17, 2, 18])],
      output: `var HelloJohn = <Hello.Compound name="John" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `var HelloJohn = <Hello.Compound name="John"></Hello.Compound>;`,
      errors: [expectNotSelfClosingError([1, 17, 1, 62])],
      output: `var HelloJohn = <Hello.Compound name="John" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `var contentContainer = <div className="content"></div>;`,
      options: [{ html: true }],
      errors: [expectNotSelfClosingError([1, 24, 1, 55])],
      output: `var contentContainer = <div className="content" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `var contentContainer = <div className="content">
</div>;`,
      options: [{ html: true }],
      errors: [expectNotSelfClosingError([1, 24, 2, 7])],
      output: `var contentContainer = <div className="content" />;`,
      languageOptions: upstreamModuleOptions,
    },
  ],
});
