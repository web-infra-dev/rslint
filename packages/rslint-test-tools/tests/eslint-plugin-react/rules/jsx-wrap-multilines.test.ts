import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectMissingParensError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'missingParens',
    'Missing parentheses around multilines JSX',
    range,
  );
}

function expectExtraParensError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'extraParens',
    'Expected no parentheses around multilines JSX',
    range,
  );
}

function expectParensOnNewLinesError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'parensOnNewLines',
    'Parentheses around JSX should be on separate lines',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-wrap-multilines', {
  valid: [
    {
      code: `function App() { return <div>single</div>; }`,
    },
    {
      code: `function App() {\n  return (\n    <div>\n      <span />\n    </div>\n  );\n}`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <p>Hello {this.props.name}</p>;
    }
  });
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <p>Hello {this.props.name}</p>;
    }
  });
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <>Hello {this.props.name}</>;
    }
  });
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <>Hello {this.props.name}</>;
    }
  });
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return (<div>
        <p>Hello {this.props.name}</p>
      </div>);
    }
  });
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <p>Hello {this.props.name}</p>;
    }
  });
`,
      options: [{ return: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <>Hello {this.props.name}</>;
    }
  });
`,
      options: [{ return: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return (<div>
        <p>Hello {this.props.name}</p>
      </div>);
    }
  });
`,
      options: [{ return: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <div>
        <p>Hello {this.props.name}</p>
      </div>;
    }
  });
`,
      options: [{ return: 'ignore' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <div>
        <p>Hello {this.props.name}</p>
      </div>;
    }
  });
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <div>
        <p>Hello {this.props.name}</p>
      </div>;
    }
  });
`,
      options: [{ return: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `var hello = foo ? <p>Hello</p> : <p>Hi</p>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `var hello = foo ? <p>Hello</p> : <p>Hi</p>;`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `var hello = foo ? <>Hello</> : <>Hi</>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
  var hello = foo ? (<div>
    <p>Hello</p>
  </div>) : (<div>
    <p>Hi</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `var hello = foo ? <p>Hello</p> : <p>Hi</p>;`,
      options: [{ declaration: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
  var hello = foo ? (<div>
    <p>Hello</p>
  </div>) : (<div>
    <p>Hi</p>
  </div>);
`,
      options: [{ declaration: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
  var hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`,
      options: [{ declaration: 'ignore' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
  var hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
  var hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`,
      options: [{ declaration: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `var hello; hello = foo ? <p>Hello</p> : <p>Hi</p>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `var hello; hello = foo ? <p>Hello</p> : <p>Hi</p>;`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
  var hello;
  hello = foo ? (<div>
    <p>Hello</p>
  </div>) : (<div>
    <p>Hi</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `var hello; hello = foo ? <p>Hello</p> : <p>Hi</p>;`,
      options: [{ assignment: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
  var hello;
  hello = foo ? (<div>
    <p>Hello</p>
  </div>) : (<div>
    <p>Hi</p>
  </div>);
`,
      options: [{ assignment: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
  var hello;
  hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`,
      options: [{ assignment: 'ignore' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
  var hello;
  hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
  var hello;
  hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`,
      options: [{ assignment: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `var hello = <p>Hello</p>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `var hello = <p>Hello</p>;`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
  var hello = (<div>
    <p>Hello</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
  var hello = (<>
    <p>Hello</p>
  </>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `var hello = <p>Hello</p>;`,
      options: [{ declaration: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
  var hello = (<div>
    <p>Hello</p>
  </div>);
`,
      options: [{ declaration: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
  var hello = <div>
    <p>Hello</p>
  </div>;
`,
      options: [{ declaration: 'ignore' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
  var hello = <div>
    <p>Hello</p>
  </div>;
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `
  var hello = <>
    <p>Hello</p>
  </>;
`,
      options: [{ declaration: 'ignore' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
      code: `
  var hello = <>
    <p>Hello</p>
  </>;
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
      code: `
  var hello = <div>
    <p>Hello</p>
  </div>;
`,
      options: [{ declaration: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 39 (7.37.5)',
      code: `var hello; hello = <p>Hello</p>;`,
      options: [{ declaration: 'ignore' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 40 (7.37.5)',
      code: `var hello; hello = <p>Hello</p>;`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 41 (7.37.5)',
      code: `var hello; hello = <p>Hello</p>;`,
      options: [{ declaration: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 42 (7.37.5)',
      code: `
  var hello;
  hello = (<div>
    <p>Hello</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 43 (7.37.5)',
      code: `
  var hello;
  hello = (<>
    <p>Hello</p>
  </>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 44 (7.37.5)',
      code: `
  var hello;
  hello = (<div>
    <p>Hello</p>
  </div>);
`,
      options: [{ assignment: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 45 (7.37.5)',
      code: `
  var hello;
  hello = <div>
    <p>Hello</p>
  </div>;
`,
      options: [{ assignment: 'ignore' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 46 (7.37.5)',
      code: `
  var hello;
  hello = <div>
    <p>Hello</p>
  </div>;
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 47 (7.37.5)',
      code: `
  var hello;
  hello = <>
    <p>Hello</p>
  </>;
`,
      options: [{ assignment: 'ignore' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 48 (7.37.5)',
      code: `
  var hello;
  hello = <>
    <p>Hello</p>
  </>;
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 49 (7.37.5)',
      code: `
  var hello;
  hello = <div>
    <p>Hello</p>
  </div>;
`,
      options: [{ assignment: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 50 (7.37.5)',
      code: `
  var hello = () => (<div>
    <p>Hello</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 51 (7.37.5)',
      code: `
  var hello = () => (<>
    <p>Hello</p>
  </>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 52 (7.37.5)',
      code: `var hello = () => <p>Hello</p>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 53 (7.37.5)',
      code: `
  var hello = () => (<div>
    <p>Hello</p>
  </div>);
`,
      options: [{ arrow: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 54 (7.37.5)',
      code: `var hello = () => <p>Hello</p>;`,
      options: [{ arrow: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 55 (7.37.5)',
      code: `
  var hello = () => <div>
    <p>Hello</p>
  </div>;
`,
      options: [{ arrow: 'ignore' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 56 (7.37.5)',
      code: `
  var hello = () => <div>
    <p>Hello</p>
  </div>;
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 57 (7.37.5)',
      code: `
  var hello = () => <>
    <p>Hello</p>
  </>;
`,
      options: [{ arrow: 'ignore' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 58 (7.37.5)',
      code: `
  var hello = () => <>
    <p>Hello</p>
  </>;
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 59 (7.37.5)',
      code: `
  var hello = () => <div>
    <p>Hello</p>
  </div>;
`,
      options: [{ arrow: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 60 (7.37.5)',
      code: `foo ? <p>Hello</p> : null;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 61 (7.37.5)',
      code: `foo ? <p>Hello</p> : null;`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 62 (7.37.5)',
      code: `foo ? <p>Hello</p> : null;`,
      options: [{ condition: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 63 (7.37.5)',
      code: `
  <div>
    {foo ? <div>
        <p>Hello</p>
      </div> : null}
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 64 (7.37.5)',
      code: `
  <div>
    {foo ? <div>
        <p>Hello</p>
      </div> : null}
  </div>
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 65 (7.37.5)',
      code: `
  <div>
    {foo ? (<div>
        <p>Hello</p>
      </div>) : null}
  </div>
`,
      options: [{ condition: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 66 (7.37.5)',
      code: `
  <div>
    {foo ? (<>
        <p>Hello</p>
      </>) : null}
  </div>
`,
      options: [{ condition: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 67 (7.37.5)',
      code: `foo && <p>Hello</p>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 68 (7.37.5)',
      code: `foo && <p>Hello</p>;`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 69 (7.37.5)',
      code: `
  <div>
    {foo &&
      <div>
        <p>Hello World</p>
      </div>
    }
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 70 (7.37.5)',
      code: `
  <div>
    {foo &&
      <div>
        <p>Hello World</p>
      </div>
    }
  </div>
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 71 (7.37.5)',
      code: `
  <div>
    {foo &&
      (<div>
        <p>Hello World</p>
      </div>)
    }
  </div>
`,
      options: [{ logical: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 72 (7.37.5)',
      code: `
  <div>
    {foo &&
      (<>
        <p>Hello World</p>
      </>)
    }
  </div>
`,
      options: [{ logical: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 73 (7.37.5)',
      code: `<div prop={<p>Hello</p>}></div>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 74 (7.37.5)',
      code: `<div prop={<p>Hello</p>}></div>`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 75 (7.37.5)',
      code: `
  <div prop={
    <div>
      <p>Hello</p>
    </div>
  }>
    <p>Hello</p>
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 76 (7.37.5)',
      code: `
  <div prop={
    <div>
      <p>Hello</p>
    </div>
  }>
    <p>Hello</p>
  </div>
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 77 (7.37.5)',
      code: `
  <div prop={
    (<div>
      <p>Hello</p>
    </div>)
  }>
    <p>Hello</p>
  </div>
`,
      options: [{ prop: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 78 (7.37.5)',
      code: `
  <div prop={
    (<>
      <p>Hello</p>
    </>)
  }>
    <p>Hello</p>
  </div>
`,
      options: [{ prop: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 79 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return (
        <div>
          <p>Hello {this.props.name}</p>
        </div>
      );
    }
  });
`,
      options: [{ return: 'parens-new-line' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 80 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return (
        <>
          <p>Hello {this.props.name}</p>
        </>
      );
    }
  });
`,
      options: [{ return: 'parens-new-line' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 81 (7.37.5)',
      code: `
  var hello = foo ? (
    <div>
    <p>Hello</p>
    </div>
  ) : (
    <div>
      <p>Hi</p>
    </div>
  );
`,
      options: [{ declaration: 'parens-new-line' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 82 (7.37.5)',
      code: `
  var hello;
  hello = foo ? (
    <div>
      <p>Hello</p>
    </div>
  ) : (
    <div>
      <p>Hi</p>
    </div>
  );
`,
      options: [{ assignment: 'parens-new-line' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 83 (7.37.5)',
      code: `
  var hello = (
    <div>
      <p>Hello</p>
    </div>
  );
`,
      options: [{ declaration: 'parens-new-line' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 84 (7.37.5)',
      code: `
  var hello;
  hello = (
    <div>
      <p>Hello</p>
    </div>
  );
`,
      options: [{ assignment: 'parens-new-line' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 85 (7.37.5)',
      code: `
  var hello = () => (
    <div>
      <p>Hello</p>
    </div>
  );
`,
      options: [{ arrow: 'parens-new-line' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 86 (7.37.5)',
      code: `
  <div>
    {foo ? (
      <div>
        <p>Hello</p>
      </div>
    ) : null}
  </div>
`,
      options: [{ condition: 'parens-new-line' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 87 (7.37.5)',
      code: `
  <div>
    {foo && (
      <div>
        <p>Hello World</p>
      </div>
    )}
  </div>
`,
      options: [{ logical: 'parens-new-line' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 88 (7.37.5)',
      code: `
  <div prop={(
    <div>
      <p>Hello</p>
    </div>
  )}>
    <p>Hello</p>
  </div>
`,
      options: [{ prop: 'parens-new-line' }],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `function App() {\n  return <div>\n    <span />\n  </div>;\n}`,
      errors: [expectMissingParensError([2, 10, 4, 9])],
      output: `function App() {
  return (<div>
    <span />
  </div>);
}`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return (<div>
        <p>Hello {this.props.name}</p>
      </div>);
    }
  });
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [expectExtraParensError([4, 15, 6, 13])],
      output: `
  var Hello = createReactClass({
    render: function() {
      return <div>
        <p>Hello {this.props.name}</p>
      </div>;
    }
  });
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <div>
        <p>Hello {this.props.name}</p>
      </div>;
    }
  });
`,
      errors: [expectMissingParensError([4, 14, 6, 13])],
      output: `
  var Hello = createReactClass({
    render: function() {
      return (<div>
        <p>Hello {this.props.name}</p>
      </div>);
    }
  });
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <>
        <p>Hello {this.props.name}</p>
      </>;
    }
  });
`,
      errors: [expectMissingParensError([4, 14, 6, 10])],
      output: `
  var Hello = createReactClass({
    render: function() {
      return (<>
        <p>Hello {this.props.name}</p>
      </>);
    }
  });
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return (<>
        <p>Hello {this.props.name}</p>
      </>);
    }
  });
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [expectExtraParensError([4, 15, 6, 10])],
      output: `
  var Hello = createReactClass({
    render: function() {
      return <>
        <p>Hello {this.props.name}</p>
      </>;
    }
  });
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <div>
        <p>Hello {this.props.name}</p>
      </div>;
    }
  });
`,
      options: [{ return: true }],
      errors: [expectMissingParensError([4, 14, 6, 13])],
      output: `
  var Hello = createReactClass({
    render: function() {
      return (<div>
        <p>Hello {this.props.name}</p>
      </div>);
    }
  });
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <>
        <p>Hello {this.props.name}</p>
      </>;
    }
  });
`,
      options: [{ return: true }],
      errors: [expectMissingParensError([4, 14, 6, 10])],
      output: `
  var Hello = createReactClass({
    render: function() {
      return (<>
        <p>Hello {this.props.name}</p>
      </>);
    }
  });
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
  var hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`,
      errors: [
        expectMissingParensError([2, 21, 4, 9]),
        expectMissingParensError([4, 12, 6, 9]),
      ],
      output: `
  var hello = foo ? (<div>
    <p>Hello</p>
  </div>) : (<div>
    <p>Hi</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
  var hello = foo ? (<div>
    <p>Hello</p>
  </div>) : (<div>
    <p>Hi</p>
  </div>);
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [
        expectExtraParensError([2, 22, 4, 9]),
        expectExtraParensError([4, 14, 6, 9]),
      ],
      output: `
  var hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
  var hello = foo ? <>
    <p>Hello</p>
  </> : <>
    <p>Hi</p>
  </>;
`,
      errors: [
        expectMissingParensError([2, 21, 4, 6]),
        expectMissingParensError([4, 9, 6, 6]),
      ],
      output: `
  var hello = foo ? (<>
    <p>Hello</p>
  </>) : (<>
    <p>Hi</p>
  </>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
  var hello = foo ? (<>
    <p>Hello</p>
  </>) : (<>
    <p>Hi</p>
  </>);
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [
        expectExtraParensError([2, 22, 4, 6]),
        expectExtraParensError([4, 11, 6, 6]),
      ],
      output: `
  var hello = foo ? <>
    <p>Hello</p>
  </> : <>
    <p>Hi</p>
  </>;
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
  var hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`,
      options: [{ declaration: true }],
      errors: [
        expectMissingParensError([2, 21, 4, 9]),
        expectMissingParensError([4, 12, 6, 9]),
      ],
      output: `
  var hello = foo ? (<div>
    <p>Hello</p>
  </div>) : (<div>
    <p>Hi</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
  var hello = foo ? <>
    <p>Hello</p>
  </> : <>
    <p>Hi</p>
  </>;
`,
      options: [{ declaration: true }],
      errors: [
        expectMissingParensError([2, 21, 4, 6]),
        expectMissingParensError([4, 9, 6, 6]),
      ],
      output: `
  var hello = foo ? (<>
    <p>Hello</p>
  </>) : (<>
    <p>Hi</p>
  </>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
  var hello;
  hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`,
      errors: [
        expectMissingParensError([3, 17, 5, 9]),
        expectMissingParensError([5, 12, 7, 9]),
      ],
      output: `
  var hello;
  hello = foo ? (<div>
    <p>Hello</p>
  </div>) : (<div>
    <p>Hi</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
  var hello;
  hello = foo ? (<div>
    <p>Hello</p>
  </div>) : (<div>
    <p>Hi</p>
  </div>);
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [
        expectExtraParensError([3, 18, 5, 9]),
        expectExtraParensError([5, 14, 7, 9]),
      ],
      output: `
  var hello;
  hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
  var hello;
  hello = foo ? <>
    <p>Hello</p>
  </> : <>
    <p>Hi</p>
  </>;
`,
      errors: [
        expectMissingParensError([3, 17, 5, 6]),
        expectMissingParensError([5, 9, 7, 6]),
      ],
      output: `
  var hello;
  hello = foo ? (<>
    <p>Hello</p>
  </>) : (<>
    <p>Hi</p>
  </>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
  var hello;
  hello = foo ? (<>
    <p>Hello</p>
  </>) : (<>
    <p>Hi</p>
  </>);
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [
        expectExtraParensError([3, 18, 5, 6]),
        expectExtraParensError([5, 11, 7, 6]),
      ],
      output: `
  var hello;
  hello = foo ? <>
    <p>Hello</p>
  </> : <>
    <p>Hi</p>
  </>;
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
  var hello;
  hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`,
      options: [{ assignment: true }],
      errors: [
        expectMissingParensError([3, 17, 5, 9]),
        expectMissingParensError([5, 12, 7, 9]),
      ],
      output: `
  var hello;
  hello = foo ? (<div>
    <p>Hello</p>
  </div>) : (<div>
    <p>Hi</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
  var hello;
  hello = foo ? <>
    <p>Hello</p>
  </> : <>
    <p>Hi</p>
  </>;
`,
      options: [{ assignment: true }],
      errors: [
        expectMissingParensError([3, 17, 5, 6]),
        expectMissingParensError([5, 9, 7, 6]),
      ],
      output: `
  var hello;
  hello = foo ? (<>
    <p>Hello</p>
  </>) : (<>
    <p>Hi</p>
  </>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
  var hello = <div>
    <p>Hello</p>
  </div>;
`,
      errors: [expectMissingParensError([2, 15, 4, 9])],
      output: `
  var hello = (<div>
    <p>Hello</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
  var hello = (<div>
    <p>Hello</p>
  </div>);
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [expectExtraParensError([2, 16, 4, 9])],
      output: `
  var hello = <div>
    <p>Hello</p>
  </div>;
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
  var hello = <>
    <p>Hello</p>
  </>;
`,
      errors: [expectMissingParensError([2, 15, 4, 6])],
      output: `
  var hello = (<>
    <p>Hello</p>
  </>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
  var hello = (<>
    <p>Hello</p>
  </>);
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [expectExtraParensError([2, 16, 4, 6])],
      output: `
  var hello = <>
    <p>Hello</p>
  </>;
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `
  var hello = <div>
    <p>Hello</p>
  </div>;
`,
      options: [{ declaration: true }],
      errors: [expectMissingParensError([2, 15, 4, 9])],
      output: `
  var hello = (<div>
    <p>Hello</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
  var hello;
  hello = <div>
    <p>Hello</p>
  </div>;
`,
      errors: [expectMissingParensError([3, 11, 5, 9])],
      output: `
  var hello;
  hello = (<div>
    <p>Hello</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
  var hello;
  hello = <>
    <p>Hello</p>
  </>;
`,
      errors: [expectMissingParensError([3, 11, 5, 6])],
      output: `
  var hello;
  hello = (<>
    <p>Hello</p>
  </>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
  var hello;
  hello = <div>
    <p>Hello</p>
  </div>;
`,
      options: [{ assignment: true }],
      errors: [expectMissingParensError([3, 11, 5, 9])],
      output: `
  var hello;
  hello = (<div>
    <p>Hello</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
  var hello;
  hello = (<div>
    <p>Hello</p>
  </div>);
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [expectExtraParensError([3, 12, 5, 9])],
      output: `
  var hello;
  hello = <div>
    <p>Hello</p>
  </div>;
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
  var hello = () => <div>
    <p>Hello</p>
  </div>;
`,
      errors: [expectMissingParensError([2, 21, 4, 9])],
      output: `
  var hello = () => (<div>
    <p>Hello</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `
  var hello = () => (<div>
    <p>Hello</p>
  </div>);
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [expectExtraParensError([2, 22, 4, 9])],
      output: `
  var hello = () => <div>
    <p>Hello</p>
  </div>;
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `
  var hello = () => <>
    <p>Hello</p>
  </>;
`,
      errors: [expectMissingParensError([2, 21, 4, 6])],
      output: `
  var hello = () => (<>
    <p>Hello</p>
  </>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `
  var hello = () => (<>
    <p>Hello</p>
  </>);
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [expectExtraParensError([2, 22, 4, 6])],
      output: `
  var hello = () => <>
    <p>Hello</p>
  </>;
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `
  var hello = () => <div>
    <p>Hello</p>
  </div>;
`,
      options: [{ arrow: true }],
      errors: [expectMissingParensError([2, 21, 4, 9])],
      output: `
  var hello = () => (<div>
    <p>Hello</p>
  </div>);
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `
  <div>
    {foo ? <div>
        <p>Hello</p>
      </div> : null}
  </div>
`,
      options: [{ condition: 'parens' }],
      errors: [expectMissingParensError([3, 12, 5, 13])],
      output: `
  <div>
    {foo ? (<div>
        <p>Hello</p>
      </div>) : null}
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `
  <div>
    {foo ? (<div>
        <p>Hello</p>
      </div>) : null}
  </div>
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [expectExtraParensError([3, 13, 5, 13])],
      output: `
  <div>
    {foo ? <div>
        <p>Hello</p>
      </div> : null}
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
  <div>
    {foo ? <>
        <p>Hello</p>
      </> : null}
  </div>
`,
      options: [{ condition: 'parens' }],
      errors: [expectMissingParensError([3, 12, 5, 10])],
      output: `
  <div>
    {foo ? (<>
        <p>Hello</p>
      </>) : null}
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
  <div>
    {foo ? (<>
        <p>Hello</p>
      </>) : null}
  </div>
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [expectExtraParensError([3, 13, 5, 10])],
      output: `
  <div>
    {foo ? <>
        <p>Hello</p>
      </> : null}
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `
  <div>
    {foo ? <div>
        <p>Hello</p>
      </div> : null}
  </div>
`,
      options: [{ condition: true }],
      errors: [expectMissingParensError([3, 12, 5, 13])],
      output: `
  <div>
    {foo ? (<div>
        <p>Hello</p>
      </div>) : null}
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
      code: `
  <div>
    {foo &&
      <div>
        <p>Hello World</p>
      </div>
    }
  </div>
`,
      options: [{ logical: 'parens' }],
      errors: [expectMissingParensError([4, 7, 6, 13])],
      output: `
  <div>
    {foo &&
      (<div>
        <p>Hello World</p>
      </div>)
    }
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `
  <div>
    {foo &&
      (<div>
        <p>Hello World</p>
      </div>)
    }
  </div>
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [expectExtraParensError([4, 8, 6, 13])],
      output: `
  <div>
    {foo &&
      <div>
        <p>Hello World</p>
      </div>
    }
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `
  <div>
    {foo &&
      <>
        <p>Hello World</p>
      </>
    }
  </div>
`,
      options: [{ logical: 'parens' }],
      errors: [expectMissingParensError([4, 7, 6, 10])],
      output: `
  <div>
    {foo &&
      (<>
        <p>Hello World</p>
      </>)
    }
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `
  <div>
    {foo &&
      (<>
        <p>Hello World</p>
      </>)
    }
  </div>
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [expectExtraParensError([4, 8, 6, 10])],
      output: `
  <div>
    {foo &&
      <>
        <p>Hello World</p>
      </>
    }
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 41 (7.37.5)',
      code: `
  <div>
    {foo &&
      <div>
        <p>Hello World</p>
      </div>
    }
  </div>
`,
      options: [{ logical: true }],
      errors: [expectMissingParensError([4, 7, 6, 13])],
      output: `
  <div>
    {foo &&
      (<div>
        <p>Hello World</p>
      </div>)
    }
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 42 (7.37.5)',
      code: `
  <div prop={
    <div>
      <p>Hello</p>
    </div>
  }>
    <p>Hello</p>
  </div>
`,
      options: [{ prop: 'parens' }],
      errors: [expectMissingParensError([3, 5, 5, 11])],
      output: `
  <div prop={
    (<div>
      <p>Hello</p>
    </div>)
  }>
    <p>Hello</p>
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 43 (7.37.5)',
      code: `
  <div prop={
    (<div>
      <p>Hello</p>
    </div>)
  }>
    <p>Hello</p>
  </div>
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [expectExtraParensError([3, 6, 5, 11])],
      output: `
  <div prop={
    <div>
      <p>Hello</p>
    </div>
  }>
    <p>Hello</p>
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 44 (7.37.5)',
      code: `
  <div prop={
    <>
      <p>Hello</p>
    </>
  }>
    <p>Hello</p>
  </div>
`,
      options: [{ prop: 'parens' }],
      errors: [expectMissingParensError([3, 5, 5, 8])],
      output: `
  <div prop={
    (<>
      <p>Hello</p>
    </>)
  }>
    <p>Hello</p>
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 45 (7.37.5)',
      code: `
  <div prop={
    (<>
      <p>Hello</p>
    </>)
  }>
    <p>Hello</p>
  </div>
`,
      options: [
        {
          declaration: 'never',
          assignment: 'never',
          return: 'never',
          arrow: 'never',
          condition: 'never',
          logical: 'never',
          prop: 'never',
        },
      ],
      errors: [expectExtraParensError([3, 6, 5, 8])],
      output: `
  <div prop={
    <>
      <p>Hello</p>
    </>
  }>
    <p>Hello</p>
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 46 (7.37.5)',
      code: `
  <div prop={
    <div>
      <p>Hello</p>
    </div>
  }>
    <p>Hello</p>
  </div>
`,
      options: [{ prop: true }],
      errors: [expectMissingParensError([3, 5, 5, 11])],
      output: `
  <div prop={
    (<div>
      <p>Hello</p>
    </div>)
  }>
    <p>Hello</p>
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 47 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <div>
        <p>Hello {this.props.name}</p>
      </div>;
    }
  });
`,
      options: [{ return: 'parens-new-line' }],
      errors: [expectMissingParensError([4, 14, 6, 13])],
      output: `
  var Hello = createReactClass({
    render: function() {
      return (
      <div>
        <p>Hello {this.props.name}</p>
      </div>
      );
    }
  });
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 48 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return <>
        <p>Hello {this.props.name}</p>
      </>;
    }
  });
`,
      options: [{ return: 'parens-new-line' }],
      errors: [expectMissingParensError([4, 14, 6, 10])],
      output: `
  var Hello = createReactClass({
    render: function() {
      return (
      <>
        <p>Hello {this.props.name}</p>
      </>
      );
    }
  });
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 49 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return (<div>
        <p>Hello {this.props.name}</p>
      </div>);
    }
  });
`,
      options: [{ return: 'parens-new-line' }],
      errors: [expectParensOnNewLinesError([4, 15, 6, 13])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 50 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return (

      <div>
        <p>Hello {this.props.name}</p>
      </div>);
    }
  });
`,
      options: [{ return: 'parens-new-line' }],
      errors: [expectParensOnNewLinesError([6, 7, 8, 13])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 51 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return (<div>
        <p>Hello {this.props.name}</p>
      </div>

      );
    }
  });
`,
      options: [{ return: 'parens-new-line' }],
      errors: [expectParensOnNewLinesError([4, 15, 6, 13])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 52 (7.37.5)',
      code: `
  var Hello = createReactClass({
    render: function() {
      return (<>
        <p>Hello {this.props.name}</p>
      </>);
    }
  });
`,
      options: [{ return: 'parens-new-line' }],
      errors: [expectParensOnNewLinesError([4, 15, 6, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 53 (7.37.5)',
      code: `
  var hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`,
      options: [{ declaration: 'parens-new-line' }],
      errors: [
        expectMissingParensError([2, 21, 4, 9]),
        expectMissingParensError([4, 12, 6, 9]),
      ],
      output: `
  var hello = foo ? (
  <div>
    <p>Hello</p>
  </div>
  ) : (
  <div>
    <p>Hi</p>
  </div>
  );
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 54 (7.37.5)',
      code: `
  var hello = foo ? <>
    <p>Hello</p>
  </> : <>
    <p>Hi</p>
  </>;
`,
      options: [{ declaration: 'parens-new-line' }],
      errors: [
        expectMissingParensError([2, 21, 4, 6]),
        expectMissingParensError([4, 9, 6, 6]),
      ],
      output: `
  var hello = foo ? (
  <>
    <p>Hello</p>
  </>
  ) : (
  <>
    <p>Hi</p>
  </>
  );
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 55 (7.37.5)',
      code: `
  var hello = foo ? (<>
    <p>Hello</p>
  </>) : (<>
    <p>Hi</p>
  </>);
`,
      options: [{ declaration: 'parens-new-line' }],
      errors: [
        expectParensOnNewLinesError([2, 22, 4, 6]),
        expectParensOnNewLinesError([4, 11, 6, 6]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 56 (7.37.5)',
      code: `
  var hello = foo ? (<div>
    <p>Hello</p>
  </div>) : (<div>
    <p>Hi</p>
  </div>);
`,
      options: [{ declaration: 'parens-new-line' }],
      errors: [
        expectParensOnNewLinesError([2, 22, 4, 9]),
        expectParensOnNewLinesError([4, 14, 6, 9]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 57 (7.37.5)',
      code: `
  var hello;
  hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`,
      options: [{ assignment: 'parens-new-line' }],
      errors: [
        expectMissingParensError([3, 17, 5, 9]),
        expectMissingParensError([5, 12, 7, 9]),
      ],
      output: `
  var hello;
  hello = foo ? (
  <div>
    <p>Hello</p>
  </div>
  ) : (
  <div>
    <p>Hi</p>
  </div>
  );
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 58 (7.37.5)',
      code: `
  var hello;
  hello = foo ? <>
    <p>Hello</p>
  </> : <>
    <p>Hi</p>
  </>;
`,
      options: [{ assignment: 'parens-new-line' }],
      errors: [
        expectMissingParensError([3, 17, 5, 6]),
        expectMissingParensError([5, 9, 7, 6]),
      ],
      output: `
  var hello;
  hello = foo ? (
  <>
    <p>Hello</p>
  </>
  ) : (
  <>
    <p>Hi</p>
  </>
  );
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 59 (7.37.5)',
      code: `
  var hello;
  hello = foo ? (<div>
    <p>Hello</p>
  </div>) : (<div>
    <p>Hi</p>
  </div>);
`,
      options: [{ assignment: 'parens-new-line' }],
      errors: [
        expectParensOnNewLinesError([3, 18, 5, 9]),
        expectParensOnNewLinesError([5, 14, 7, 9]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 60 (7.37.5)',
      code: `
  var hello;
  hello = foo ? (<>
    <p>Hello</p>
  </>) : (<>
    <p>Hi</p>
  </>);
`,
      options: [{ assignment: 'parens-new-line' }],
      errors: [
        expectParensOnNewLinesError([3, 18, 5, 6]),
        expectParensOnNewLinesError([5, 11, 7, 6]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 61 (7.37.5)',
      code: `
  var hello = <div>
    <p>Hello</p>
  </div>;
`,
      options: [{ declaration: 'parens-new-line' }],
      errors: [expectMissingParensError([2, 15, 4, 9])],
      output: `
  var hello = (
  <div>
    <p>Hello</p>
  </div>
  );
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 62 (7.37.5)',
      code: `
  var hello = (<div>
    <p>Hello</p>
  </div>);
`,
      options: [{ declaration: 'parens-new-line' }],
      errors: [expectParensOnNewLinesError([2, 16, 4, 9])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 63 (7.37.5)',
      code: `
  var hello;
  hello = <div>
    <p>Hello</p>
  </div>;
`,
      options: [{ assignment: 'parens-new-line' }],
      errors: [expectMissingParensError([3, 11, 5, 9])],
      output: `
  var hello;
  hello = (
  <div>
    <p>Hello</p>
  </div>
  );
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 64 (7.37.5)',
      code: `
  var hello;
  hello = (<div>
    <p>Hello</p>
  </div>);
`,
      options: [{ assignment: 'parens-new-line' }],
      errors: [expectParensOnNewLinesError([3, 12, 5, 9])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 65 (7.37.5)',
      code: `
  var hello = () => (<div>
    <p>Hello</p>
  </div>);
`,
      options: [{ arrow: 'parens-new-line' }],
      errors: [expectParensOnNewLinesError([2, 22, 4, 9])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 66 (7.37.5)',
      code: `
  var hello = () => (<>
    <p>Hello</p>
  </>);
`,
      options: [{ arrow: 'parens-new-line' }],
      errors: [expectParensOnNewLinesError([2, 22, 4, 6])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 67 (7.37.5)',
      code: `
  var hello = () => <div>
    <p>Hello</p>
  </div>;
`,
      options: [{ arrow: 'parens-new-line' }],
      errors: [expectMissingParensError([2, 21, 4, 9])],
      output: `
  var hello = () => (
  <div>
    <p>Hello</p>
  </div>
  );
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 68 (7.37.5)',
      code: `
  var hello = () => <>
    <p>Hello</p>
  </>;
`,
      options: [{ arrow: 'parens-new-line' }],
      errors: [expectMissingParensError([2, 21, 4, 6])],
      output: `
  var hello = () => (
  <>
    <p>Hello</p>
  </>
  );
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 69 (7.37.5)',
      code: `
  <div>
    {foo ? (<div>
        <p>Hello</p>
      </div>) : null}
  </div>
`,
      options: [{ condition: 'parens-new-line' }],
      errors: [expectParensOnNewLinesError([3, 13, 5, 13])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 70 (7.37.5)',
      code: `
  <div>
    {foo ? (<>
        <p>Hello</p>
      </>) : null}
  </div>
`,
      options: [{ condition: 'parens-new-line' }],
      errors: [expectParensOnNewLinesError([3, 13, 5, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 71 (7.37.5)',
      code: `
  <div>
    {foo ? <div>
        <p>Hello</p>
      </div> : null}
  </div>
`,
      options: [{ condition: 'parens-new-line' }],
      errors: [expectMissingParensError([3, 12, 5, 13])],
      output: `
  <div>
    {foo ? (
    <div>
        <p>Hello</p>
      </div>
    ) : null}
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 72 (7.37.5)',
      code: `
  <div>
    {foo ? <>
        <p>Hello</p>
      </> : null}
  </div>
`,
      options: [{ condition: 'parens-new-line' }],
      errors: [expectMissingParensError([3, 12, 5, 10])],
      output: `
  <div>
    {foo ? (
    <>
        <p>Hello</p>
      </>
    ) : null}
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 73 (7.37.5)',
      code: `
  <div>
    {foo &&
      (<div>
        <p>Hello World</p>
      </div>)
    }
  </div>
`,
      options: [{ logical: 'parens-new-line' }],
      errors: [expectParensOnNewLinesError([4, 8, 6, 13])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 74 (7.37.5)',
      code: `
  <div>
    {foo &&
      <div>
        <p>Hello World</p>
      </div>
    }
  </div>
`,
      options: [{ logical: 'parens-new-line' }],
      errors: [expectMissingParensError([4, 7, 6, 13])],
      output: `
  <div>
    {foo &&
      (
      <div>
        <p>Hello World</p>
      </div>
      )
    }
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 75 (7.37.5)',
      code: `
  <div>
    {foo &&
      <>
        <p>Hello World</p>
      </>
    }
  </div>
`,
      options: [{ logical: 'parens-new-line' }],
      errors: [expectMissingParensError([4, 7, 6, 10])],
      output: `
  <div>
    {foo &&
      (
      <>
        <p>Hello World</p>
      </>
      )
    }
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 76 (7.37.5)',
      code: `
  <div prop={
    (<div>
      <p>Hello</p>
    </div>)
  }>
    <p>Hello</p>
  </div>
`,
      options: [{ prop: 'parens-new-line' }],
      errors: [expectParensOnNewLinesError([3, 6, 5, 11])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 77 (7.37.5)',
      code: `
  <div prop={
    (<>
      <p>Hello</p>
    </>)
  }>
    <p>Hello</p>
  </div>
`,
      options: [{ prop: 'parens-new-line' }],
      errors: [expectParensOnNewLinesError([3, 6, 5, 8])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 78 (7.37.5)',
      code: `
  <div prop={
    <div>
      <p>Hello</p>
    </div>
  }>
    <p>Hello</p>
  </div>
`,
      options: [{ prop: 'parens-new-line' }],
      errors: [expectMissingParensError([3, 5, 5, 11])],
      output: `
  <div prop={
    (
    <div>
      <p>Hello</p>
    </div>
    )
  }>
    <p>Hello</p>
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 79 (7.37.5)',
      code: `
  <div prop={
    <>
      <p>Hello</p>
    </>
  }>
    <p>Hello</p>
  </div>
`,
      options: [{ prop: 'parens-new-line' }],
      errors: [expectMissingParensError([3, 5, 5, 8])],
      output: `
  <div prop={
    (
    <>
      <p>Hello</p>
    </>
    )
  }>
    <p>Hello</p>
  </div>
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 80 (7.37.5)',
      code: `
  export default () =>
    <div>
        with newline without parentheses eslint crashes
    </div>`,
      options: [
        {
          declaration: 'parens-new-line',
          assignment: 'parens-new-line',
          return: 'parens-new-line',
          arrow: 'parens-new-line',
          condition: 'parens-new-line',
          logical: 'parens-new-line',
          prop: 'parens-new-line',
        },
      ],
      errors: [expectMissingParensError([3, 5, 5, 11])],
      output: `
  export default () =>
    (
    <div>
        with newline without parentheses eslint crashes
    </div>
    )`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 81 (7.37.5)',
      code: `
  const Component = () =>
    <div>
      <p>Some text</p>
    </div>

  export { Component as default }
`,
      options: [
        {
          declaration: 'parens-new-line',
          assignment: 'parens-new-line',
          return: 'parens-new-line',
          arrow: 'parens-new-line',
          condition: 'parens-new-line',
          logical: 'ignore',
          prop: 'ignore',
        },
      ],
      errors: [expectMissingParensError([3, 5, 5, 11])],
      output: `
  const Component = () =>
    (
    <div>
      <p>Some text</p>
    </div>
    )

  export { Component as default }
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 82 (7.37.5)',
      code: `
const Component = props => (
  <div>
    {true &&
      <div>
        <p>Some text</p>
      </div>
    }
  </div>
)
`,
      options: [{ logical: 'parens-new-line' }],
      errors: [expectMissingParensError([5, 7, 7, 13])],
      output: `
const Component = props => (
  <div>
    {true &&
      (
      <div>
        <p>Some text</p>
      </div>
      )
    }
  </div>
)
`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 83 (7.37.5)',
      code: `
        import React from 'react';

        const A =
        <div>
            B
        </div>;
      `,
      options: [{ declaration: 'parens-new-line' }],
      errors: [expectMissingParensError([5, 9, 7, 15])],
      output: `
        import React from 'react';

        const A =
        (
        <div>
            B
        </div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 84 (7.37.5)',
      code: `
        export default () => <h1>
            <span>
                hello
            </span>
        </h1>;
      `,
      options: [{ arrow: 'parens-new-line' }],
      errors: [expectMissingParensError([2, 30, 6, 14])],
      output: `
        export default () => (
        <h1>
            <span>
                hello
            </span>
        </h1>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 85 (7.37.5)',
      code: `
        export default () => (<h1>
            <span>
                hello
            </span>
        </h1>);
      `,
      options: [{ arrow: 'parens-new-line' }],
      errors: [expectParensOnNewLinesError([2, 31, 6, 14])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
