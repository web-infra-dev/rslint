import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectForbiddenPropTypeError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'forbiddenPropType',
    message: 'Prop type "any" is forbidden',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectForbiddenPropTypeError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'forbiddenPropType',
    message: 'Prop type "array" is forbidden',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectForbiddenPropTypeError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'forbiddenPropType',
    message: 'Prop type "object" is forbidden',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectForbiddenPropTypeError4(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'forbiddenPropType',
    message: 'Prop type "instanceOf" is forbidden',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('forbid-prop-types', {
  valid: [
    // ---- Upstream valid cases (propTypes) ----
    {
      code: `
        var First = createReactClass({
          render: function() {
            return <div />;
          }
        });
      `,
    },
    {
      code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <div />;
          }
        });
      `,
    },
    {
      code: `
        var First = createReactClass({
          propTypes: {
            s: PropTypes.string,
            n: PropTypes.number,
            i: PropTypes.instanceOf,
            b: PropTypes.bool
          },
          render: function() {
            return <div />;
          }
        });
      `,
    },
    {
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        })
      `,
      options: [{ forbid: ['any', 'object'] }],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: {
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ forbid: ['any', 'array'] }],
    },
    {
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
          a: PropTypes.string,
          b: PropTypes.string
        };
        First.propTypes.justforcheck = PropTypes.string;
      `,
    },
    {
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
          elem: PropTypes.instanceOf(HTMLElement)
        };
      `,
    },
    {
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "aria-controls": PropTypes.string
        };
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            let { a, ...b } = obj;
            let c = { ...d };
            return <div />;
          }
        });
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          propTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
    },
    {
      code: `
        class Test extends React.component {
          static propTypes = {
            intl: React.propTypes.number,
            ...propTypes
          };
        }
      `,
    },
    {
      code: `
        class Test extends React.component {
          static get propTypes() {
            return {
              intl: React.propTypes.number,
              ...propTypes
            };
          };
        }
      `,
    },

    // ---- contextTypes / childContextTypes — valid (option ON, but values
    // are not in the default forbid list) ----
    {
      code: `
        var First = createReactClass({
          childContextTypes: externalPropTypes,
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ checkContextTypes: true }],
    },
    {
      code: `
        var First = createReactClass({
          childContextTypes: {
            s: PropTypes.string,
            n: PropTypes.number,
            i: PropTypes.instanceOf,
            b: PropTypes.bool
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ checkContextTypes: true }],
    },

    // ---- propWrappers / packages ----
    {
      code: `
        class TestComponent extends React.Component {
          static defaultProps = function () {
            const date = new Date();
            return {
              date
            };
          }();
        }
      `,
    },
    {
      code: `
        class HeroTeaserList extends React.Component {
          render() { return null; }
        }
        HeroTeaserList.propTypes = Object.assign({
          heroIndex: PropTypes.number,
          preview: PropTypes.bool,
        }, componentApi, teaserListProps);
      `,
    },
    {
      code: `
        import PropTypes from "prop-types";
        const Foo = {
          foo: PropTypes.string,
        };
        const Bar = {
          bar: PropTypes.shape(Foo),
        };
      `,
    },
    {
      code: `
        import yup from "yup"
        const formValidation = Yup.object().shape({
          name: Yup.string(),
          customer_ids: Yup.array()
        });
      `,
    },
    {
      code: `
        import yup from "Yup"
        const validation = yup.object().shape({
          address: yup.object({
            city: yup.string(),
            zip: yup.string(),
          })
        })
      `,
      options: [{ forbid: ['string', 'object'] }],
    },
    {
      code: `
        import yup from "yup"
        Yup.array(
          Yup.object().shape({
            value: Yup.number()
          })
        )
      `,
      options: [{ forbid: ['number'] }],
    },
    {
      code: `
        import CustomPropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: CustomPropTypes.shape({
            b: CustomPropTypes.String,
            c: CustomPropTypes.number.isRequired,
          })
        }
      `,
    },
    {
      code: `
        import CustomReact from "react"
        class Component extends React.Component {};
        Component.propTypes = {
          b: CustomReact.PropTypes.string,
        }
      `,
    },
    {
      code: `
        import PropTypes from "yup"
        class Component extends React.Component {};
        Component.propTypes = {
          b: PropTypes.array(),
        }
      `,
    },
    {
      code: `
        import { PropTypes, shape, any } from "yup"
        class Component extends React.Component {};
        Component.propTypes = {
          b: PropTypes.any,
        }
      `,
      options: [{ forbid: ['any'] }],
    },
    {
      code: `
        import { PropTypes } from "not-react"
        class Component extends React.Component {};
        Component.propTypes = {
          b: PropTypes.array(),
        }
      `,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            o: PropTypes.object,
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ forbid: ['any', 'array'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        var First = createReactClass({
          childContextTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ forbid: ['any', 'object'], checkContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        var First = createReactClass({
          childContextTypes: {
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ forbid: ['any', 'array'], checkContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        var First = createReactClass({
          childContextTypes: {
            o: PropTypes.object,
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ forbid: ['any', 'array'], checkContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.childContextTypes = {
          a: PropTypes.string,
          b: PropTypes.string
        };
        First.childContextTypes.justforcheck = PropTypes.string;
      `,
      options: [{ checkContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.childContextTypes = {
          elem: PropTypes.instanceOf(HTMLElement)
        };
      `,
      options: [{ checkContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.childContextTypes = {
          "aria-controls": PropTypes.string
        };
      `,
      options: [{ checkContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            let { a, ...b } = obj;
            let c = { ...d };
            return <div />;
          }
        });
      `,
      options: [{ checkContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        var Hello = createReactClass({
          childContextTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ checkContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        class Test extends React.component {
          static childContextTypes = {
            intl: React.childContextTypes.number,
            ...childContextTypes
          };
        }
      `,
      options: [{ checkContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        class Test extends React.component {
          static get childContextTypes() {
            return {
              intl: React.childContextTypes.number,
              ...childContextTypes
            };
          };
        }
      `,
      options: [{ checkContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        var First = createReactClass({
          childContextTypes: externalPropTypes,
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ checkChildContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        var First = createReactClass({
          childContextTypes: {
            s: PropTypes.string,
            n: PropTypes.number,
            i: PropTypes.instanceOf,
            b: PropTypes.bool
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ checkChildContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        var First = createReactClass({
          childContextTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ forbid: ['any', 'object'], checkChildContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        var First = createReactClass({
          childContextTypes: {
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ forbid: ['any', 'array'], checkChildContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        var First = createReactClass({
          childContextTypes: {
            o: PropTypes.object,
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ forbid: ['any', 'array'], checkChildContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.childContextTypes = {
          a: PropTypes.string,
          b: PropTypes.string
        };
        First.childContextTypes.justforcheck = PropTypes.string;
      `,
      options: [{ checkChildContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.childContextTypes = {
          elem: PropTypes.instanceOf(HTMLElement)
        };
      `,
      options: [{ checkChildContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.childContextTypes = {
          "aria-controls": PropTypes.string
        };
      `,
      options: [{ checkChildContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            let { a, ...b } = obj;
            let c = { ...d };
            return <div />;
          }
        });
      `,
      options: [{ checkChildContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
        var Hello = createReactClass({
          childContextTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ checkChildContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
        class Test extends React.component {
          static childContextTypes = {
            intl: React.childContextTypes.number,
            ...childContextTypes
          };
        }
      `,
      options: [{ checkChildContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `
        class Test extends React.component {
          static get childContextTypes() {
            return {
              intl: React.childContextTypes.number,
              ...childContextTypes
            };
          };
        }
      `,
      options: [{ checkChildContextTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // ---- Upstream invalid cases (propTypes) ----
    {
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [expectForbiddenPropTypeError1([4, 13, 4, 29])],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: {
            n: PropTypes.number
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [
        {
          messageId: 'forbiddenPropType',
          message: 'Prop type "number" is forbidden',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 32,
        },
      ],
      options: [{ forbid: ['number'] }],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [expectForbiddenPropTypeError1([4, 13, 4, 40])],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [expectForbiddenPropTypeError2([4, 13, 4, 31])],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [expectForbiddenPropTypeError3([4, 13, 4, 32])],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array,
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [
        expectForbiddenPropTypeError2([4, 13, 4, 31]),
        expectForbiddenPropTypeError3([5, 13, 5, 32]),
      ],
    },
    {
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
            a: PropTypes.array,
            o: PropTypes.object
        };
        class Second extends React.Component {
          render() {
            return <div />;
          }
        }
        Second.propTypes = {
            a: PropTypes.array,
            o: PropTypes.object
        };
      `,
      errors: [
        expectForbiddenPropTypeError2([8, 13, 8, 31]),
        expectForbiddenPropTypeError3([9, 13, 9, 32]),
        expectForbiddenPropTypeError2([17, 13, 17, 31]),
        expectForbiddenPropTypeError3([18, 13, 18, 32]),
      ],
    },
    {
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = forbidExtraProps({
            a: PropTypes.array
        });
      `,
      errors: [expectForbiddenPropTypeError2([8, 13, 8, 31])],
      settings: {
        propWrapperFunctions: ['forbidExtraProps'],
      },
    },
    {
      code: `
        import { forbidExtraProps } from "airbnb-prop-types";
        export const propTypes = {dpm: PropTypes.any};
        export default function Component() {}
        Component.propTypes = propTypes;
      `,
      errors: [expectForbiddenPropTypeError1([3, 35, 3, 53])],
    },
    {
      code: `
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.array,
            o: PropTypes.object
          };
          render() {
            return <div />;
          }
        }
      `,
      errors: [
        expectForbiddenPropTypeError2([4, 13, 4, 31]),
        expectForbiddenPropTypeError3([5, 13, 5, 32]),
      ],
    },
    {
      code: `
        class Component extends React.Component {
          static get propTypes() {
            return {
              a: PropTypes.array,
              o: PropTypes.object
            };
          };
          render() {
            return <div />;
          }
        }
      `,
      errors: [
        expectForbiddenPropTypeError2([5, 15, 5, 33]),
        expectForbiddenPropTypeError3([6, 15, 6, 34]),
      ],
    },
    {
      code: `
        class Component extends React.Component {
          static propTypes = forbidExtraProps({
            a: PropTypes.array,
            o: PropTypes.object
          });
          render() {
            return <div />;
          }
        }
      `,
      errors: [
        expectForbiddenPropTypeError2([4, 13, 4, 31]),
        expectForbiddenPropTypeError3([5, 13, 5, 32]),
      ],
      settings: {
        propWrapperFunctions: ['forbidExtraProps'],
      },
    },
    {
      code: `
        var Hello = createReactClass({
          propTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ forbid: ['instanceOf'] }],
      errors: [expectForbiddenPropTypeError4([4, 13, 4, 59])],
    },
    {
      code: `
        var object = PropTypes.object;
        var Hello = createReactClass({
          propTypes: {
            retailer: object,
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ forbid: ['object'] }],
      errors: [expectForbiddenPropTypeError3([5, 13, 5, 29])],
    },

    // ---- contextTypes — invalid (option ON) ----
    {
      code: `
        var First = createReactClass({
          contextTypes: {
            a: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ checkContextTypes: true }],
      errors: [expectForbiddenPropTypeError1([4, 13, 4, 29])],
    },
    {
      code: `
        class Foo extends Component {
          static contextTypes = {
            a: PropTypes.any
          }
          render() {
            return <div />;
          }
        }
      `,
      options: [{ checkContextTypes: true }],
      errors: [expectForbiddenPropTypeError1([4, 13, 4, 29])],
    },
    {
      code: `
        class Foo extends Component {
          static get contextTypes() {
            return {
              a: PropTypes.any
            };
          }
          render() {
            return <div />;
          }
        }
      `,
      options: [{ checkContextTypes: true }],
      errors: [expectForbiddenPropTypeError1([5, 15, 5, 31])],
    },
    {
      code: `
        class Foo extends Component {
          render() {
            return <div />;
          }
        }
        Foo.contextTypes = {
          a: PropTypes.any
        };
      `,
      options: [{ checkContextTypes: true }],
      errors: [expectForbiddenPropTypeError1([8, 11, 8, 27])],
    },

    // ---- childContextTypes — invalid (option ON) ----
    {
      code: `
        var First = createReactClass({
          childContextTypes: {
            a: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ checkChildContextTypes: true }],
      errors: [expectForbiddenPropTypeError1([4, 13, 4, 29])],
    },
    {
      code: `
        class Foo extends Component {
          static childContextTypes = {
            a: PropTypes.any
          }
          render() {
            return <div />;
          }
        }
      `,
      options: [{ checkChildContextTypes: true }],
      errors: [expectForbiddenPropTypeError1([4, 13, 4, 29])],
    },
    {
      code: `
        class Foo extends Component {
          static get childContextTypes() {
            return {
              a: PropTypes.any
            };
          }
          render() {
            return <div />;
          }
        }
      `,
      options: [{ checkChildContextTypes: true }],
      errors: [expectForbiddenPropTypeError1([5, 15, 5, 31])],
    },
    {
      code: `
        class Foo extends Component {
          render() {
            return <div />;
          }
        }
        Foo.childContextTypes = {
          a: PropTypes.any
        };
      `,
      options: [{ checkChildContextTypes: true }],
      errors: [expectForbiddenPropTypeError1([8, 11, 8, 27])],
    },

    // ---- Imported PropTypes / packages ----
    {
      code: `
        import { object, string } from "prop-types";
        function C({ a, b }) { return [a, b]; }
        C.propTypes = {
          a: object,
          b: string
        };
      `,
      options: [{ forbid: ['object'] }],
      errors: [expectForbiddenPropTypeError3([5, 11, 5, 20])],
    },
    {
      code: `
        import { objectOf, any } from "prop-types";
        function C({ a }) { return a; }
        C.propTypes = {
          a: objectOf(any)
        };
      `,
      options: [{ forbid: ['any'] }],
      errors: [expectForbiddenPropTypeError1([5, 11, 5, 27])],
    },
    {
      code: `
        import { shape, any } from "prop-types";
        function C({ a }) { return a; }
        C.propTypes = {
          a: shape({
            b: any
          })
        };
      `,
      options: [{ forbid: ['any'] }],
      errors: [expectForbiddenPropTypeError1([6, 13, 6, 19])],
    },
    {
      code: `
        var First = createReactClass({
          propTypes: {
            s: PropTypes.shape({
              o: PropTypes.object
            })
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [expectForbiddenPropTypeError3([5, 15, 5, 34])],
    },
    {
      code: `
        import React from './React';

        import { arrayOf, object } from 'prop-types';

        const App = ({ foo }) => (
          <div>
            Hello world {foo}
          </div>
        );

        App.propTypes = {
          foo: arrayOf(object)
        }

        export default App;
      `,
      errors: [expectForbiddenPropTypeError3([13, 11, 13, 31])],
    },
    {
      code: `
        import CustomPropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: CustomPropTypes.shape({
            b: CustomPropTypes.String,
            c: CustomPropTypes.object.isRequired,
          })
        }
      `,
      errors: [expectForbiddenPropTypeError3([7, 13, 7, 49])],
    },
    {
      code: `
        import CustomReact from "react"
        class Component extends React.Component {};
        Component.propTypes = {
          b: CustomReact.PropTypes.object,
        }
      `,
      errors: [expectForbiddenPropTypeError3([5, 11, 5, 42])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [expectForbiddenPropTypeError2([4, 13, 4, 42])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.object.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [expectForbiddenPropTypeError3([4, 13, 4, 43])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        });
        var Second = createReactClass({
          propTypes: {
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [
        expectForbiddenPropTypeError2([4, 13, 4, 31]),
        expectForbiddenPropTypeError3([12, 13, 12, 32]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        import { forbidExtraProps } from "airbnb-prop-types";
        export const propTypes = {a: PropTypes.any};
        export default function Component() {}
        Component.propTypes = forbidExtraProps(propTypes);
      `,
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [expectForbiddenPropTypeError1([3, 35, 3, 51])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        class Component extends React.Component {
          static get propTypes() {
            return forbidExtraProps({
              a: PropTypes.array,
              o: PropTypes.object
            });
          }
          render() {
            return <div />;
          }
        }
      `,
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [
        expectForbiddenPropTypeError2([5, 15, 5, 33]),
        expectForbiddenPropTypeError3([6, 15, 6, 34]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        function Foo(props) {
          return <div />;
        }
        Foo.contextTypes = {
          a: PropTypes.any
        };
      `,
      options: [{ checkContextTypes: true }],
      errors: [expectForbiddenPropTypeError1([6, 11, 6, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
        const Foo = (props) => {
          return <div />;
        };
        Foo.contextTypes = {
          a: PropTypes.any
        };
      `,
      options: [{ checkContextTypes: true }],
      errors: [expectForbiddenPropTypeError1([6, 11, 6, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
        class Component extends React.Component {
          static contextTypes = forbidExtraProps({
            a: PropTypes.array,
            o: PropTypes.object
          });
          render() {
            return <div />;
          }
        }
      `,
      options: [{ checkContextTypes: true }],
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [
        expectForbiddenPropTypeError2([4, 13, 4, 31]),
        expectForbiddenPropTypeError3([5, 13, 5, 32]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
        class Component extends React.Component {
          static get contextTypes() {
            return forbidExtraProps({
              a: PropTypes.array,
              o: PropTypes.object
            });
          }
          render() {
            return <div />;
          }
        }
      `,
      options: [{ checkContextTypes: true }],
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [
        expectForbiddenPropTypeError2([5, 15, 5, 33]),
        expectForbiddenPropTypeError3([6, 15, 6, 34]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.contextTypes = forbidExtraProps({
          a: PropTypes.array,
          o: PropTypes.object
        });
      `,
      options: [{ checkContextTypes: true }],
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [
        expectForbiddenPropTypeError2([8, 11, 8, 29]),
        expectForbiddenPropTypeError3([9, 11, 9, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `
        function Component(props) {
          return <div />;
        }
        Component.contextTypes = forbidExtraProps({
          a: PropTypes.array,
          o: PropTypes.object
        });
      `,
      options: [{ checkContextTypes: true }],
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [
        expectForbiddenPropTypeError2([6, 11, 6, 29]),
        expectForbiddenPropTypeError3([7, 11, 7, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `
        const Component = (props) => {
          return <div />;
        };
        Component.contextTypes = forbidExtraProps({
          a: PropTypes.array,
          o: PropTypes.object
        });
      `,
      options: [{ checkContextTypes: true }],
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [
        expectForbiddenPropTypeError2([6, 11, 6, 29]),
        expectForbiddenPropTypeError3([7, 11, 7, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `
        var Hello = createReactClass({
          contextTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ forbid: ['instanceOf'], checkContextTypes: true }],
      errors: [expectForbiddenPropTypeError4([4, 13, 4, 59])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `
        class Component extends React.Component {
          static contextTypes = {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          }
          render() {
            return <div />;
          }
        }
      `,
      options: [{ forbid: ['instanceOf'], checkContextTypes: true }],
      errors: [expectForbiddenPropTypeError4([4, 13, 4, 59])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `
        class Component extends React.Component {
          static get contextTypes() {
            return {
              retailer: PropTypes.instanceOf(Map).isRequired,
              requestRetailer: PropTypes.func.isRequired
            };
          }
          render() {
            return <div />;
          }
        }
      `,
      options: [{ forbid: ['instanceOf'], checkContextTypes: true }],
      errors: [expectForbiddenPropTypeError4([5, 15, 5, 61])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.contextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `,
      options: [{ forbid: ['instanceOf'], checkContextTypes: true }],
      errors: [expectForbiddenPropTypeError4([8, 11, 8, 57])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
        function Component(props) {
          return <div />;
        }
        Component.contextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `,
      options: [{ forbid: ['instanceOf'], checkContextTypes: true }],
      errors: [expectForbiddenPropTypeError4([6, 11, 6, 57])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
        const Component = (props) => {
          return <div />;
        };
        Component.contextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        }
      `,
      options: [{ forbid: ['instanceOf'], checkContextTypes: true }],
      errors: [expectForbiddenPropTypeError4([6, 11, 6, 57])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `
        function Foo(props) {
          return <div />;
        }
        Foo.childContextTypes = {
          a: PropTypes.any
        };
      `,
      options: [{ checkChildContextTypes: true }],
      errors: [expectForbiddenPropTypeError1([6, 11, 6, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 41 (7.37.5)',
      code: `
        const Foo = (props) => {
          return <div />;
        };
        Foo.childContextTypes = {
          a: PropTypes.any
        };
      `,
      options: [{ checkChildContextTypes: true }],
      errors: [expectForbiddenPropTypeError1([6, 11, 6, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 42 (7.37.5)',
      code: `
        class Component extends React.Component {
          static childContextTypes = forbidExtraProps({
            a: PropTypes.array,
            o: PropTypes.object
          });
          render() {
            return <div />;
          }
        }
      `,
      options: [{ checkChildContextTypes: true }],
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [
        expectForbiddenPropTypeError2([4, 13, 4, 31]),
        expectForbiddenPropTypeError3([5, 13, 5, 32]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 43 (7.37.5)',
      code: `
        class Component extends React.Component {
          static get childContextTypes() {
            return forbidExtraProps({
              a: PropTypes.array,
              o: PropTypes.object
            });
          }
          render() {
            return <div />;
          }
        }
      `,
      options: [{ checkChildContextTypes: true }],
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [
        expectForbiddenPropTypeError2([5, 15, 5, 33]),
        expectForbiddenPropTypeError3([6, 15, 6, 34]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 44 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.childContextTypes = forbidExtraProps({
          a: PropTypes.array,
          o: PropTypes.object
        });
      `,
      options: [{ checkChildContextTypes: true }],
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [
        expectForbiddenPropTypeError2([8, 11, 8, 29]),
        expectForbiddenPropTypeError3([9, 11, 9, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 45 (7.37.5)',
      code: `
        function Component(props) {
          return <div />;
        }
        Component.childContextTypes = forbidExtraProps({
          a: PropTypes.array,
          o: PropTypes.object
        });
      `,
      options: [{ checkChildContextTypes: true }],
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [
        expectForbiddenPropTypeError2([6, 11, 6, 29]),
        expectForbiddenPropTypeError3([7, 11, 7, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 46 (7.37.5)',
      code: `
        const Component = (props) => {
          return <div />;
        };
        Component.childContextTypes = forbidExtraProps({
          a: PropTypes.array,
          o: PropTypes.object
        });
      `,
      options: [{ checkChildContextTypes: true }],
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [
        expectForbiddenPropTypeError2([6, 11, 6, 29]),
        expectForbiddenPropTypeError3([7, 11, 7, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 47 (7.37.5)',
      code: `
        var Hello = createReactClass({
          childContextTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ forbid: ['instanceOf'], checkChildContextTypes: true }],
      errors: [expectForbiddenPropTypeError4([4, 13, 4, 59])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 48 (7.37.5)',
      code: `
        class Component extends React.Component {
          static childContextTypes = {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          }
          render() {
            return <div />;
          }
        }
      `,
      options: [{ forbid: ['instanceOf'], checkChildContextTypes: true }],
      errors: [expectForbiddenPropTypeError4([4, 13, 4, 59])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 49 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.childContextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `,
      options: [{ forbid: ['instanceOf'], checkChildContextTypes: true }],
      errors: [expectForbiddenPropTypeError4([8, 11, 8, 57])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 50 (7.37.5)',
      code: `
        function Component(props) {
          return <div />;
        }
        Component.childContextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `,
      options: [{ forbid: ['instanceOf'], checkChildContextTypes: true }],
      errors: [expectForbiddenPropTypeError4([6, 11, 6, 57])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 51 (7.37.5)',
      code: `
        const Component = (props) => {
          return <div />;
        };
        Component.childContextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `,
      options: [{ forbid: ['instanceOf'], checkChildContextTypes: true }],
      errors: [expectForbiddenPropTypeError4([6, 11, 6, 57])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 54 (7.37.5)',
      code: `
        import { objectOf, any } from "prop-types";
        function C({ a }) { return a; }
        C.propTypes = {
          a: objectOf(any)
        };
      `,
      options: [{ forbid: ['objectOf'] }],
      errors: [
        {
          messageId: 'forbiddenPropType',
          message: 'Prop type "objectOf" is forbidden',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 27,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 56 (7.37.5)',
      code: `
        import { any } from "prop-types";
        function C({ a }) { return a; }
        C.propTypes = {
          a: PropTypes.shape({
            b: any
          })
        };
      `,
      options: [{ forbid: ['any'] }],
      errors: [expectForbiddenPropTypeError1([6, 13, 6, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 59 (7.37.5)',
      code: `
        import React from './React';

        import PropTypes, { arrayOf } from 'prop-types';

        const App = ({ foo }) => (
          <div>
            Hello world {foo}
          </div>
        );

        App.propTypes = {
          foo: arrayOf(PropTypes.object)
        }

        export default App;
      `,
      errors: [expectForbiddenPropTypeError3([13, 11, 13, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 61 (7.37.5)',
      code: `
        import { PropTypes as CustomPropTypes } from "react";
        class Component extends React.Component {};
        Component.propTypes = {
          a: CustomPropTypes.shape({
            b: CustomPropTypes.String,
            c: CustomPropTypes.object.isRequired,
          })
        }
      `,
      errors: [expectForbiddenPropTypeError3([7, 13, 7, 49])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
