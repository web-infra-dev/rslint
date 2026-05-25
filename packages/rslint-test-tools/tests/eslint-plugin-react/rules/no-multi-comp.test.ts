import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// Mirrors upstream's `code.split('\n').join('\r')` line-terminator transform.
const cr = (s: string) => s.replace(/\n/g, '\r');

ruleTester.run('no-multi-comp', {} as never, {
  valid: [
    // ---- Single component declarations ----
    {
      code: `
        var Hello = require('./components/Hello');
        var HelloJohn = createReactClass({
          render: function() {
            return <Hello name="John" />;
          }
        });
      `,
    },
    {
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
    },
    {
      code: `
        var Heading = createReactClass({
          render: function() {
            return (
              <div>
                {this.props.buttons.map(function(button, index) {
                  return <Button {...button} key={index}/>;
                })}
              </div>
            );
          }
        });
      `,
    },

    // ---- ignoreStateless: true ----
    {
      code: `
        function Hello(props) {
          return <div>Hello {props.name}</div>;
        }
        function HelloAgain(props) {
          return <div>Hello again {props.name}</div>;
        }
      `,
      options: [{ ignoreStateless: true }],
    },
    {
      code: `
        function Hello(props) {
          return <div>Hello {props.name}</div>;
        }
        class HelloJohn extends React.Component {
          render() {
            return <Hello name="John" />;
          }
        }
      `,
      options: [{ ignoreStateless: true }],
    },

    // ---- Helper functions are not components ----
    {
      code: `
        import React, { createElement } from "react"
        const helperFoo = () => {
          return true;
        };
        function helperBar() {
          return false;
        };
        function RealComponent() {
          return createElement("img");
        };
      `,
    },

    // ---- React.memo / React.forwardRef wrapping ----
    {
      code: `
        const Hello = React.memo(function(props) {
          return <div>Hello {props.name}</div>;
        });
        class HelloJohn extends React.Component {
          render() {
            return <Hello name="John" />;
          }
        }
      `,
      options: [{ ignoreStateless: true }],
    },

    // ---- forwardRef wrapping a sibling component (nodeWrapsComponent gate) ----
    {
      code: `
        class StoreListItem extends React.PureComponent {
          // A bunch of stuff here
        }
        export default React.forwardRef((props, ref) => <StoreListItem {...props} forwardRef={ref} />);
      `,
      options: [{ ignoreStateless: false }],
    },
    {
      code: `
        class StoreListItem extends React.PureComponent {
          // A bunch of stuff here
        }
        export default React.forwardRef((props, ref) => {
          return <StoreListItem {...props} forwardRef={ref} />
        });
      `,
      options: [{ ignoreStateless: false }],
    },
    {
      code: `
        const HelloComponent = (props) => {
          return <div></div>;
        }
        export default React.forwardRef((props, ref) => <HelloComponent {...props} forwardRef={ref} />);
      `,
      options: [{ ignoreStateless: false }],
    },
    {
      code: `
        class StoreListItem extends React.PureComponent {
          // A bunch of stuff here
        }
        export default React.forwardRef(
          function myFunction(props, ref) {
            return <StoreListItem {...props} forwardedRef={ref} />;
          }
        );
      `,
      options: [{ ignoreStateless: true }],
    },
    {
      code: `
        const HelloComponent = (props) => {
          return <div></div>;
        }
        class StoreListItem extends React.PureComponent {
          // A bunch of stuff here
        }
        export default React.forwardRef(
          function myFunction(props, ref) {
            return <StoreListItem {...props} forwardedRef={ref} />;
          }
        );
      `,
      options: [{ ignoreStateless: true }],
    },
    {
      code: `
        const HelloComponent = (props) => {
          return <div></div>;
        }
        export default React.memo((props, ref) => <HelloComponent {...props} />);
      `,
      options: [{ ignoreStateless: true }],
    },

    // ---- Class field arrow that consumes an unrelated `memo`-named helper ----
    {
      code: `
        import React from 'react';
        function memo() {
          var outOfScope = "hello"
          return null;
        }
        class ComponentY extends React.Component {
          memoCities = memo((cities) => cities.map((v) => ({ label: v })));
          render() {
            return (
              <div>
                <div>Counter</div>
              </div>
            );
          }
        }
      `,
    },

    // ---- forwardRef + .displayName / .propTypes / .defaultProps assignments ----
    {
      code: `
        const MenuList = forwardRef(({onClose, ...props}, ref) => {
          const {t} = useTranslation();
          const handleLogout = useLogoutHandler();

          const onLogout = useCallback(() => {
            onClose();
            handleLogout();
          }, [onClose, handleLogout]);

          return (
            <MuiMenuList ref={ref} {...props}>
              <MuiMenuItem key="logout" onClick={onLogout}>
                {t('global-logout')}
              </MuiMenuItem>
            </MuiMenuList>
          );
        });

        MenuList.displayName = 'MenuList';

        MenuList.propTypes = {
          onClose: PropTypes.func,
        };

        MenuList.defaultProps = {
          onClose: () => null,
        };

        export default MenuList;
      `,
    },
    {
      code: `
        const MenuList = forwardRef(({ onClose, ...props }, ref) => {
          const onLogout = useCallback(() => {
            onClose()
          }, [onClose])

          return (
            <BlnMenuList ref={ref} {...props}>
              <BlnMenuItem key="logout" onClick={onLogout}>
                Logout
              </BlnMenuItem>
            </BlnMenuList>
          )
        })

        MenuList.displayName = 'MenuList'

        MenuList.propTypes = {
          onClose: PropTypes.func
        }

        MenuList.defaultProps = {
          onClose: () => null
        }

        export default MenuList
      `,
    },

    // ---- Real-world extras (parity with Go suite) ----
    {
      code: `
        type Props = { name: string };
        const Hello: React.FC<Props> = (props) => <div>{props.name}</div>;
      `,
    },
    {
      code: `
        interface Props { label: string }
        const Btn = React.forwardRef<HTMLButtonElement, Props>((props, ref) => (
          <button ref={ref}>{props.label}</button>
        ));
      `,
    },
    {
      code: `
        const enhanced = withFoo((props) => <div>{props.x}</div>);
        class App extends React.Component { render() { return <div /> } }
      `,
      options: [{ ignoreStateless: true }],
    },
    {
      code: `
        const App = myObserver((props) => <div>{props.x}</div>);
      `,
      settings: { componentWrapperFunctions: ['myObserver'] },
    },
    {
      code: `
        const App = MyLib.observer((props) => <div>{props.x}</div>);
      `,
      settings: {
        componentWrapperFunctions: [{ property: 'observer', object: 'MyLib' }],
      },
    },
    {
      code: `
        const Hello = React?.memo((props) => <div>{props.x}</div>);
      `,
    },
    {
      code: `
        import { useState } from 'react';
        export function useToggle(init) {
          const [on, setOn] = useState(init);
          return [on, () => setOn(v => !v)];
        }
      `,
    },
    {
      code: `
        someLib.register('foo', (props) => <div>{props.x}</div>);
        someLib.register('bar', (props) => <span>{props.y}</span>);
      `,
    },

    // ---- `componentWrapperFunctions` "<pragma>" placeholder ----
    {
      code: `
        const App = Foo.observer((props) => <div>{props.x}</div>);
      `,
      settings: {
        react: { pragma: 'Foo' },
        componentWrapperFunctions: [
          { property: 'observer', object: '<pragma>' },
        ],
      },
    },

    // ---- ignoreStateless filters object-literal shorthand methods ----
    {
      code: `
        export default {
          A() { return <div /> },
          B() { return <div /> },
          C() { return <div /> },
        };
      `,
      options: [{ ignoreStateless: true }],
    },
  ],
  invalid: [
    {
      code: cr(`
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
        var HelloJohn = createReactClass({
          render: function() {
            return <Hello name="John" />;
          }
        });
      `),
      errors: [{ messageId: 'onlyOneComponent', line: 7 }],
    },
    {
      code: cr(`
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
        class HelloJohn extends React.Component {
          render() {
            return <Hello name="John" />;
          }
        }
        class HelloJohnny extends React.Component {
          render() {
            return <Hello name="Johnny" />;
          }
        }
      `),
      errors: [
        { messageId: 'onlyOneComponent', line: 7 },
        { messageId: 'onlyOneComponent', line: 12 },
      ],
    },
    {
      code: `
        function Hello(props) {
          return <div>Hello {props.name}</div>;
        }
        function HelloAgain(props) {
          return <div>Hello again {props.name}</div>;
        }
      `,
      errors: [{ messageId: 'onlyOneComponent', line: 5 }],
    },
    {
      code: `
        function Hello(props) {
          return <div>Hello {props.name}</div>;
        }
        class HelloJohn extends React.Component {
          render() {
            return <Hello name="John" />;
          }
        }
      `,
      errors: [{ messageId: 'onlyOneComponent', line: 5 }],
    },
    {
      code: `
        export default {
          RenderHello(props) {
            let {name} = props;
            return <div>{name}</div>;
          },
          RenderHello2(props) {
            let {name} = props;
            return <div>{name}</div>;
          }
        };
      `,
      errors: [{ messageId: 'onlyOneComponent', line: 7 }],
    },
    {
      code: `
        exports.Foo = function Foo() {
          return <></>
        }

        exports.createSomeComponent = function createSomeComponent(opts) {
          return function Foo() {
            return <>{opts.a}</>
          }
        }
      `,
      errors: [{ messageId: 'onlyOneComponent', line: 7 }],
    },
    {
      code: `
        class StoreListItem extends React.PureComponent {
          // A bunch of stuff here
        }
        export default React.forwardRef((props, ref) => <div><StoreListItem {...props} forwardRef={ref} /></div>);
      `,
      options: [{ ignoreStateless: false }],
      errors: [{ messageId: 'onlyOneComponent', line: 5 }],
    },
    {
      code: `
        const HelloComponent = (props) => {
          return <div></div>;
        }
        const HelloComponent2 = React.forwardRef((props, ref) => <div></div>);
      `,
      options: [{ ignoreStateless: false }],
      errors: [{ messageId: 'onlyOneComponent', line: 5 }],
    },
    {
      code: `
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = React.forwardRef((props, ref) => <><HelloComponent></HelloComponent></>);
      `,
      options: [{ ignoreStateless: false }],
      errors: [{ messageId: 'onlyOneComponent', line: 5 }],
    },
    {
      code: `
        const forwardRef = React.forwardRef;
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);
      `,
      options: [{ ignoreStateless: false }],
      errors: [{ messageId: 'onlyOneComponent', line: 6 }],
    },
    {
      code: `
        const memo = React.memo;
        const HelloComponent = (props) => {
          return <div></div>;
        };
        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);
      `,
      options: [{ ignoreStateless: false }],
      errors: [{ messageId: 'onlyOneComponent', line: 6 }],
    },
    {
      code: `
        const {forwardRef} = React;
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);
      `,
      options: [{ ignoreStateless: false }],
      errors: [{ messageId: 'onlyOneComponent', line: 6 }],
    },
    {
      code: `
        const {memo} = React;
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);
      `,
      options: [{ ignoreStateless: false }],
      errors: [{ messageId: 'onlyOneComponent', line: 6 }],
    },
    {
      code: `
        import React, { memo } from 'react';
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);
      `,
      options: [{ ignoreStateless: false }],
      errors: [{ messageId: 'onlyOneComponent', line: 6 }],
    },
    {
      code: `
        import {forwardRef} from 'react';
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);
      `,
      options: [{ ignoreStateless: false }],
      errors: [{ messageId: 'onlyOneComponent', line: 6 }],
    },
    {
      code: `
        const { memo } = require('react');
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);
      `,
      options: [{ ignoreStateless: false }],
      errors: [{ messageId: 'onlyOneComponent', line: 6 }],
    },
    {
      code: `
        const {forwardRef} = require('react');
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);
      `,
      options: [{ ignoreStateless: false }],
      errors: [{ messageId: 'onlyOneComponent', line: 6 }],
    },
    {
      code: `
        const forwardRef = require('react').forwardRef;
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);
      `,
      options: [{ ignoreStateless: false }],
      errors: [{ messageId: 'onlyOneComponent', line: 6 }],
    },
    {
      code: `
        const memo = require('react').memo;
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);
      `,
      options: [{ ignoreStateless: false }],
      errors: [{ messageId: 'onlyOneComponent', line: 6 }],
    },
    {
      code: `
        import Foo, { memo, forwardRef } from 'foo';
        const Text = forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        })
        const Label = memo(() => <Text />);
      `,
      settings: { react: { pragma: 'Foo' } },
      errors: [{ messageId: 'onlyOneComponent' }],
    },

    // ---- Real-world extras (parity with Go suite) ----
    {
      code: `
        type Props = { name: string };
        class Hello extends React.Component<Props> { render() { return <div>{this.props.name}</div>; } }
        const Bye: React.FC<Props> = (props) => <div>Bye {props.name}</div>;
      `,
      errors: [{ messageId: 'onlyOneComponent', line: 4 }],
    },
    {
      code: `
        const A = (props) => props.cond ? <div /> : <span />;
        const B = (props) => <div>{props.x}</div>;
      `,
      errors: [{ messageId: 'onlyOneComponent', line: 3 }],
    },
    {
      code: `
        class Outer extends React.Component {
          inner = React.memo((props) => <div>{props.x}</div>);
          render() { return <div /> }
        }
        function Sibling(props) { return <div>{props.y}</div>; }
      `,
      errors: [
        { messageId: 'onlyOneComponent', line: 3 },
        { messageId: 'onlyOneComponent', line: 6 },
      ],
    },
    {
      code: `
        const A = (0, 0, (props) => <div>{props.x}</div>);
        const B = (0, (props) => <span>{props.y}</span>);
      `,
      errors: [{ messageId: 'onlyOneComponent', line: 3 }],
    },
    {
      code: `
        const A = myObserver((props) => <div>{props.x}</div>);
        const B = myObserver((props) => <div>{props.y}</div>);
      `,
      settings: { componentWrapperFunctions: ['myObserver'] },
      errors: [{ messageId: 'onlyOneComponent', line: 3 }],
    },
    {
      code: `
        function Hello(props) { return <div>{props.name}</div>; }
        class HelloAgain extends React.Component { render() { return <div /> } }
      `,
      errors: [{ messageId: 'onlyOneComponent', line: 3 }],
    },
    {
      code: `
        function A() { return <div /> }
        function B() { return <div /> }
        function C() { return <div /> }
      `,
      errors: [
        { messageId: 'onlyOneComponent', line: 3 },
        { messageId: 'onlyOneComponent', line: 4 },
      ],
    },

    // ---- Contract: exact message text + position fields ----
    {
      code: `
        function Hello(props) { return <div /> }
        function HelloAgain(props) { return <div /> }
      `,
      errors: [
        {
          messageId: 'onlyOneComponent',
          message: 'Declare only one React component per file',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 54,
        },
      ],
    },
    {
      code: `
        class First extends React.Component {
          render() { return <div /> }
        }
        class Second extends React.Component {
          render() { return <div /> }
        }
      `,
      errors: [
        {
          messageId: 'onlyOneComponent',
          line: 5,
          column: 9,
          endLine: 7,
          endColumn: 10,
        },
      ],
    },
  ],
});
