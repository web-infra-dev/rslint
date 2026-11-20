import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectStylePropNotObjectError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'stylePropNotObject',
    message: 'Style prop value must be an object',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('style-prop-object', {
  valid: [
    {
      code: `<div style={{ color: 'red' }} />`,
    },
    {
      code: `const s = { color: 'red' }; <div style={s} />`,
    },
    {
      code: `<div className="foo" />`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `<div style={{ color: "red" }} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `<Hello style={{ color: "red" }} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        function redDiv() {
          const styles = { color: "red" };
          return <div style={styles} />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        function redDiv() {
          const styles = { color: "red" };
          return <Hello style={styles} />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        const styles = { color: "red" };
        function redDiv() {
          return <div style={styles} />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        function redDiv(props) {
          return <div style={props.styles} />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        import styles from './styles';
        function redDiv() {
          return <div style={styles} />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        import mystyles from './styles';
        const styles = Object.assign({ color: 'red' }, mystyles);
        function redDiv() {
          return <div style={styles} />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        const otherProps = { style: { color: "red" } };
        const { a, b, ...props } = otherProps;
        <div {...props} />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        const styles = Object.assign({ color: 'red' }, mystyles);
        React.createElement("div", { style: styles });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `<div style></div>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        React.createElement(MyCustomElem, {
          [style]: true
        }, 'My custom Elem')
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        let style;
        <div style={style}></div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        let style = null;
        <div style={style}></div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        let style = undefined;
        <div style={style}></div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `<div style={undefined}></div>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        const props = { style: undefined };
        <div {...props} />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        const otherProps = { style: undefined };
        const { a, b, ...props } = otherProps;
        <div {...props} />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        React.createElement("div", {
          style: undefined
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        let style;
        React.createElement("div", {
          style
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `<div style={null}></div>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        const props = { style: null };
        <div {...props} />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        const otherProps = { style: null };
        const { a, b, ...props } = otherProps;
        <div {...props} />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        React.createElement("div", {
          style: null
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        const MyComponent = (props) => {
          React.createElement(MyCustomElem, {
            ...props
          });
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `<MyComponent style="myStyle" />`,
      options: [{ allow: ['MyComponent'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `React.createElement(MyComponent, { style: "mySpecialStyle" })`,
      options: [{ allow: ['MyComponent'] }],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `<div style="color: red" />`,
      errors: [expectStylePropNotObjectError([1, 6, 1, 24])],
    },
    {
      code: `<div style={42} />`,
      errors: [expectStylePropNotObjectError([1, 6, 1, 16])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `<div style="color: 'red'" />`,
      errors: [expectStylePropNotObjectError([1, 6, 1, 26])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `<Hello style="color: 'red'" />`,
      errors: [expectStylePropNotObjectError([1, 8, 1, 28])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `<div style={true} />`,
      errors: [expectStylePropNotObjectError([1, 6, 1, 18])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        const styles = 'color: "red"';
        function redDiv2() {
          return <div style={styles} />;
        }
      `,
      errors: [expectStylePropNotObjectError([4, 23, 4, 37])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        const styles = 'color: "red"';
        function redDiv2() {
          return <Hello style={styles} />;
        }
      `,
      errors: [expectStylePropNotObjectError([4, 25, 4, 39])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        const styles = true;
        function redDiv() {
          return <div style={styles} />;
        }
      `,
      errors: [expectStylePropNotObjectError([4, 23, 4, 37])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `<MyComponent style="myStyle" />`,
      options: [{ allow: ['MyOtherComponent'] }],
      errors: [expectStylePropNotObjectError([1, 14, 1, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `React.createElement(MyComponent, { style: "mySpecialStyle" })`,
      options: [{ allow: ['MyOtherComponent'] }],
      errors: [expectStylePropNotObjectError([1, 43, 1, 59])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
