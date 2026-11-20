import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectNoChildrenInVoidElError1(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'noChildrenInVoidEl',
    'Void DOM element <br /> cannot receive children.',
    range,
  );
}

function expectNoChildrenInVoidElError2(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'noChildrenInVoidEl',
    'Void DOM element <img /> cannot receive children.',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('void-dom-elements-no-children', {
  valid: [
    {
      code: `<div>Children</div>`,
    },
    {
      code: `<br />`,
    },
    {
      code: `<img src="image.png" />`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `<div>Foo</div>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `<div children="Foo" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `<div dangerouslySetInnerHTML={{ __html: "Foo" }} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `React.createElement("div", {}, "Foo");`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `React.createElement("div", { children: "Foo" });`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `React.createElement("div", { dangerouslySetInnerHTML: { __html: "Foo" } });`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `document.createElement("img");`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `React.createElement("img");`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `React.createElement();`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `document.createElement();`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        const props = {};
        React.createElement("img", props);
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        import React, {createElement} from "react";
        createElement("div");
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        import React, {createElement} from "react";
        createElement("img");
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        import React, {createElement, PureComponent} from "react";
        class Button extends PureComponent {
          handleClick(ev) {
            ev.preventDefault();
          }
          render() {
            return <div onClick={this.handleClick}>Hello</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `<br>Children</br>`,
      errors: [expectNoChildrenInVoidElError1([1, 1, 1, 18])],
    },
    {
      code: `<img children="foo" />`,
      errors: [expectNoChildrenInVoidElError2([1, 1, 1, 23])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `<br>Foo</br>;`,
      errors: [expectNoChildrenInVoidElError1([1, 1, 1, 13])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `<br children="Foo" />;`,
      errors: [expectNoChildrenInVoidElError1([1, 1, 1, 22])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `<img {...props} children="Foo" />;`,
      errors: [expectNoChildrenInVoidElError2([1, 1, 1, 34])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `<br dangerouslySetInnerHTML={{ __html: "Foo" }} />;`,
      errors: [expectNoChildrenInVoidElError1([1, 1, 1, 51])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `React.createElement("br", {}, "Foo");`,
      errors: [expectNoChildrenInVoidElError1([1, 1, 1, 37])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `React.createElement("br", { children: "Foo" });`,
      errors: [expectNoChildrenInVoidElError1([1, 1, 1, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `React.createElement("br", { dangerouslySetInnerHTML: { __html: "Foo" } });`,
      errors: [expectNoChildrenInVoidElError1([1, 1, 1, 74])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        import React, {createElement} from "react";
        createElement("img", {}, "Foo");
      `,
      errors: [expectNoChildrenInVoidElError2([3, 9, 3, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        import React, {createElement} from "react";
        createElement("img", { children: "Foo" });
      `,
      errors: [expectNoChildrenInVoidElError2([3, 9, 3, 50])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        import React, {createElement} from "react";
        createElement("img", { dangerouslySetInnerHTML: { __html: "Foo" } });
      `,
      errors: [expectNoChildrenInVoidElError2([3, 9, 3, 77])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
