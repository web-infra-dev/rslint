import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('forbid-prop-types', {} as never, {
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 2,
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
      errors: 4,
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
      errors: 1,
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
      errors: 1,
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
      errors: 2,
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
      errors: 2,
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
      errors: 2,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
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
      errors: 1,
    },
    {
      code: `
        import CustomReact from "react"
        class Component extends React.Component {};
        Component.propTypes = {
          b: CustomReact.PropTypes.object,
        }
      `,
      errors: 1,
    },
  ],
});
