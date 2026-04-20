import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-find-dom-node', {} as never, {
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
      errors: [{ messageId: 'noFindDOMNode' }],
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
      errors: [{ messageId: 'noFindDOMNode' }],
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
      errors: [{ messageId: 'noFindDOMNode' }],
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
      errors: [{ messageId: 'noFindDOMNode' }],
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
      errors: [{ messageId: 'noFindDOMNode' }],
    },
    {
      code: `
        class Hello extends Component {
          componentDidMount() {
            React?.findDOMNode(this);
          }
        };
      `,
      errors: [{ messageId: 'noFindDOMNode' }],
    },
    {
      code: `
        class Hello extends Component {
          componentDidMount() {
            React.findDOMNode?.(this);
          }
        };
      `,
      errors: [{ messageId: 'noFindDOMNode' }],
    },
    {
      code: `
        class Hello extends Component {
          componentDidMount() {
            pkg.react.findDOMNode(this);
          }
        };
      `,
      errors: [{ messageId: 'noFindDOMNode' }],
    },
  ],
});
