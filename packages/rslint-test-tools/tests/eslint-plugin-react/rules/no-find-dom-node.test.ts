import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectNoFindDOMNodeError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noFindDOMNode',
    message:
      'Do not use findDOMNode. It doesn’t work with function components and is deprecated in StrictMode. See https://reactjs.org/docs/react-dom.html#finddomnode',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-find-dom-node', {
  valid: [
    // ---- Upstream valid cases ----
    {
      code: `
        var Hello = function() {};
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello</div>;
          }
        });
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            someNonMemberFunction(arg);
            this.someFunc = React.findDOMNode;
          },
          render: function() {
            return <div>Hello</div>;
          }
        });
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            React.someFunc(this);
          },
          render: function() {
            return <div>Hello</div>;
          }
        });
      `,
    },
    // ---- Edge: bracket notation is NOT matched ----
    {
      code: `
        class Hello extends Component {
          componentDidMount() {
            React['findDOMNode'](this).scrollIntoView();
          }
        };
      `,
    },
    // ---- Edge: similar-but-not-equal identifier name ----
    {
      code: `
        class Hello extends Component {
          componentDidMount() {
            findDOMNodes(this);
          }
        };
      `,
    },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            React.findDOMNode(this).scrollIntoView();
          },
          render: function() {
            return <div>Hello</div>;
          }
        });
      `,
      errors: [expectNoFindDOMNodeError([4, 13, 4, 30])],
    },
    {
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            ReactDOM.findDOMNode(this).scrollIntoView();
          },
          render: function() {
            return <div>Hello</div>;
          }
        });
      `,
      errors: [expectNoFindDOMNodeError([4, 13, 4, 33])],
    },
    {
      code: `
        class Hello extends Component {
          componentDidMount() {
            findDOMNode(this).scrollIntoView();
          }
          render() {
            return <div>Hello</div>;
          }
        };
      `,
      errors: [expectNoFindDOMNodeError([4, 13, 4, 24])],
    },
    {
      code: `
        class Hello extends Component {
          componentDidMount() {
            this.node = findDOMNode(this);
          }
          render() {
            return <div>Hello</div>;
          }
        };
      `,
      errors: [expectNoFindDOMNodeError([4, 25, 4, 36])],
    },
    // ---- Edge: parenthesized / optional / private / deeply chained ----
    {
      code: `
        class Hello extends Component {
          componentDidMount() {
            (React.findDOMNode)(this);
          }
        };
      `,
      errors: [expectNoFindDOMNodeError([4, 14, 4, 31])],
    },
    {
      code: `
        class Hello extends Component {
          componentDidMount() {
            React?.findDOMNode(this);
          }
        };
      `,
      errors: [expectNoFindDOMNodeError([4, 13, 4, 31])],
    },
    {
      code: `
        class Hello extends Component {
          componentDidMount() {
            React.findDOMNode?.(this);
          }
        };
      `,
      errors: [expectNoFindDOMNodeError([4, 13, 4, 30])],
    },
    {
      code: `
        class Hello extends Component {
          componentDidMount() {
            pkg.react.findDOMNode(this);
          }
        };
      `,
      errors: [expectNoFindDOMNodeError([4, 13, 4, 34])],
    },
  ],
});
