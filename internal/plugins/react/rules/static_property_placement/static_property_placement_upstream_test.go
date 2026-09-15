// TestStaticPropertyPlacementUpstream contains every valid and invalid test case from
// eslint-plugin-react v7.37.5. The upstream parsers.all helper expands each
// case across parser implementations; rslint RuleTester has one TypeScript/TSX
// parser path, so each source case is represented once here.
package static_property_placement

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	upstreamStaticField = "static public field"
	upstreamGetter      = "static getter"
	upstreamAssignment  = "property assignment"
)

var upstreamSettings = map[string]interface{}{
	"react": map[string]interface{}{
		"version": "15",
	},
}

const upstreamCode01 = `
        var MyComponent = createReactClass({
          childContextTypes: {
            something: PropTypes.bool
          },

          contextTypes: {
            something: PropTypes.bool
          },

          getDefaultProps: function() {
            name: 'Bob'
          },

          displayName: 'Hello',

          propTypes: {
            something: PropTypes.bool
          },

          render: function() {
            return null;
          },
        });
      `

const upstreamCode02 = `
        var MyComponent = React.createClass({
          childContextTypes: {
            something: PropTypes.bool
          },

          contextTypes: {
            something: PropTypes.bool
          },

          getDefaultProps: function() {
            name: 'Bob'
          },

          displayName: 'Hello',

          propTypes: {
            something: PropTypes.bool
          },

          render: function() {
            return null;
          },
        });
      `

const upstreamCode03 = `
        const MyComponent = () => {
            return <div>Hello</div>;
        };

        MyComponent.childContextTypes = {
          something: PropTypes.bool
        };

        MyComponent.contextTypes = {
          something: PropTypes.bool
        };

        MyComponent.defaultProps = {
          something: 'Bob'
        };

        MyComponent.displayName = 'Hello';

        MyComponent.propTypes = {
          something: PropTypes.bool
        };
      `

const upstreamCode04 = `
        const MyComponent = () => (<div>Hello</div>);

        MyComponent.childContextTypes = {
          something: PropTypes.bool
        };

        MyComponent.contextTypes = {
          something: PropTypes.bool
        };

        MyComponent.defaultProps = {
          something: 'Bob'
        };

        MyComponent.displayName = 'Hello';

        MyComponent.propTypes = {
          something: PropTypes.bool
        };
      `

const upstreamCode05 = `
        export function MyComponent () {
            return <div>Hello</div>;
        };

        MyComponent.childContextTypes = {
          something: PropTypes.bool
        };

        MyComponent.contextTypes = {
          something: PropTypes.bool
        };

        MyComponent.defaultProps = {
          something: 'Bob'
        };

        MyComponent.displayName = 'Hello';

        MyComponent.propTypes = {
          something: PropTypes.bool
        };
      `

const upstreamCode06 = `
        class Foo {
          static get propTypes() {}
        }
      `

const upstreamCode07 = `
        class Foo {
          static propTypes = {}
        }
      `

const upstreamCode08 = `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }
      `

const upstreamCode09 = `
        class MyComponent extends React.Component {
          static randomlyNamed = {
            name: 'random'
          }
        }
      `

const upstreamCode10 = `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.randomlyNamed = {
          name: 'random'
        }
      `

const upstreamCode11 = `
        class MyComponent extends React.Component {
          static childContextTypes = {
            something: PropTypes.bool
          };
        }
      `

const upstreamCode12 = `
        class MyComponent extends React.Component {
          static get childContextTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `

const upstreamCode13 = `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.childContextTypes = {
          name: PropTypes.string.isRequired
        }
      `

const upstreamCode14 = `
        class MyComponent extends React.Component {
          static contextTypes = {
            something: PropTypes.bool
          };
        }
      `

const upstreamCode15 = `
        class MyComponent extends React.Component {
          static get contextTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `

const upstreamCode16 = `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.contextTypes = {
          name: PropTypes.string.isRequired
        }
      `

const upstreamCode17 = `
        class MyComponent extends React.Component {
          static contextType = MyContext;
        }
      `

const upstreamCode18 = `
        class MyComponent extends React.Component {
          static get contextType() {
             return MyContext;
          }
        }
      `

const upstreamCode19 = `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.contextType = MyContext;
      `

const upstreamCode20 = `
        class MyComponent extends React.Component {
          static displayName = "Hello";
        }
      `

const upstreamCode21 = `
        class MyComponent extends React.Component {
          static get displayName() {
            return "Hello";
          }
        }
      `

const upstreamCode22 = `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.displayName = "Hello";
      `

const upstreamCode23 = `
        class MyComponent extends React.Component {
          static defaultProps = {
            something: 'Bob'
          };
        }
      `

const upstreamCode24 = `
        class MyComponent extends React.Component {
          static get defaultProps() {
            return {
              something: 'Bob'
            };
          }
        }
      `

const upstreamCode25 = `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.defaultProps = {
          name: 'Bob'
        }
      `

const upstreamCode26 = `
        class MyComponent extends React.Component {
          static propTypes = {
            something: PropTypes.bool
          };
        }
      `

const upstreamCode27 = `
        class MyComponent extends React.Component {
          static get propTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `

const upstreamCode28 = `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `

const upstreamCode29 = `
        class MyComponent extends React.Component {
          static childContextTypes = {
            something: PropTypes.bool
          };

          static contextTypes = {
            something: PropTypes.bool
          };

          static contextType = MyContext;

          static displayName = "Hello";

          static defaultProps = {
            something: 'Bob'
          };

          static propTypes = {
            something: PropTypes.bool
          };
        }
      `

const upstreamCode30 = `
        class MyComponent extends React.Component {
          static get childContextTypes() {
            return {
              something: PropTypes.bool
            };
          }

          static get contextTypes() {
            return {
              something: PropTypes.bool
            };
          }

          static get contextType() {
            return MyContext;
          }

          static get displayName() {
            return "Hello";
          }

          static get defaultProps() {
            return {
              something: PropTypes.bool
            };
          }

          static get propTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `

const upstreamCode31 = `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.childContextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.contextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.displayName = "Hello";

        MyComponent.defaultProps = {
          name: 'Bob'
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `

const upstreamCode32 = `
        class MyComponent extends React.Component {
          static childContextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static get displayName() {
            return "Hello"
          }
        }

        MyComponent.defaultProps = {
          name: 'Bob'
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `

const upstreamCode33 = `
        class MyComponent extends React.Component {
          static childContextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static displayName = "Hello";
        }

        const OtherComponent = () => (<div>Hello</div>);

        OtherComponent.defaultProps = {
          name: 'Bob'
        }

        OtherComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `

const upstreamCode34 = `
        class MyComponent extends React.Component {
          static childContextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static displayName = "Hello";
        }

        class OtherComponent extends React.Component {
          static defaultProps = {
            name: 'Bob'
          }

          static propTypes = {
            name: PropTypes.string.isRequired
          }
        }
      `

const upstreamCode35 = `
        class MyComponent extends React.Component {
          static displayName = "Hello";

          myMethod() {
            console.log(MyComponent.displayName);
          }
        }
      `

const upstreamCode36 = `
        class MyComponent extends React.Component {
          static displayName = "Hello";

          myMethod() {
            MyComponent.displayName = "Bonjour";
          }
        }
      `

const upstreamCode37 = `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.childContextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.contextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.contextType = MyContext;

        MyComponent.displayName = "Hello";

        MyComponent.defaultProps = {
          name: 'Bob'
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `

const upstreamCode38 = `
        class MyComponent extends React.Component {
          static childContextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextType = MyContext;

          static get displayName() {
            return "Hello";
          }
        }

        MyComponent.defaultProps = {
          name: 'Bob'
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `

const upstreamCode39 = `
        class MyComponent extends React.Component {
          static childContextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextType = MyContext;

          static get displayName() {
            return "Hello";
          }
        }

        const OtherComponent = () => (<div>Hello</div>);

        OtherComponent.defaultProps = {
          name: 'Bob'
        }

        OtherComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `

const upstreamCode40 = `
        class MyComponent extends React.Component {
          static childContextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextType = MyContext;

          static displayName = "Hello";
        }

        class OtherComponent extends React.Component {
          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static defaultProps = {
            name: 'Bob'
          }

          static propTypes = {
            name: PropTypes.string.isRequired
          }

          static get displayName() {
            return "Hello";
          }
        }
      `

const upstreamCode41 = `
        class MyComponent extends React.Component {
          displayName = 'Foo';
        }
      `

func upstreamValid(code string, options any) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, Settings: upstreamSettings, Options: options, Tsx: true}
}

func upstreamError(messageID, name string) rule_tester.InvalidTestCaseError {
	message := ""
	switch messageID {
	case "notStaticClassProp":
		message = "'" + name + "' should be declared as a static class property."
	case "notGetterClassFunc":
		message = "'" + name + "' should be declared as a static getter class function."
	case "declareOutsideClass":
		message = "'" + name + "' should be declared outside the class body."
	}
	return rule_tester.InvalidTestCaseError{MessageId: messageID, Message: message}
}

func upstreamInvalid(code string, options any, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{Code: code, Settings: upstreamSettings, Options: options, Tsx: true, Errors: errors}
}

func TestStaticPropertyPlacementUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		upstreamValid(upstreamCode01, []interface{}{upstreamAssignment}),
		upstreamValid(upstreamCode02, []interface{}{upstreamAssignment}),
		upstreamValid(upstreamCode03, nil),
		upstreamValid(upstreamCode04, nil),
		upstreamValid(upstreamCode05, nil),
		upstreamValid(upstreamCode06, nil),
		upstreamValid(upstreamCode07, []interface{}{upstreamAssignment}),
		upstreamValid(upstreamCode08, nil),
		upstreamValid(upstreamCode09, nil),
		upstreamValid(upstreamCode09, []interface{}{upstreamAssignment}),
		upstreamValid(upstreamCode10, []interface{}{upstreamAssignment}),
		upstreamValid(upstreamCode10, nil),
		upstreamValid(upstreamCode11, nil),
		upstreamValid(upstreamCode11, []interface{}{upstreamAssignment, map[string]interface{}{"childContextTypes": upstreamStaticField}}),
		upstreamValid(upstreamCode12, []interface{}{upstreamGetter}),
		upstreamValid(upstreamCode12, []interface{}{upstreamAssignment, map[string]interface{}{"childContextTypes": upstreamGetter}}),
		upstreamValid(upstreamCode13, []interface{}{upstreamAssignment}),
		upstreamValid(upstreamCode13, []interface{}{upstreamStaticField, map[string]interface{}{"childContextTypes": upstreamAssignment}}),
		upstreamValid(upstreamCode14, nil),
		upstreamValid(upstreamCode14, []interface{}{upstreamAssignment, map[string]interface{}{"contextTypes": upstreamStaticField}}),
		upstreamValid(upstreamCode15, []interface{}{upstreamGetter}),
		upstreamValid(upstreamCode15, []interface{}{upstreamAssignment, map[string]interface{}{"contextTypes": upstreamGetter}}),
		upstreamValid(upstreamCode16, []interface{}{upstreamAssignment}),
		upstreamValid(upstreamCode16, []interface{}{upstreamStaticField, map[string]interface{}{"contextTypes": upstreamAssignment}}),
		upstreamValid(upstreamCode17, nil),
		upstreamValid(upstreamCode17, []interface{}{upstreamAssignment, map[string]interface{}{"contextType": upstreamStaticField}}),
		upstreamValid(upstreamCode18, []interface{}{upstreamGetter}),
		upstreamValid(upstreamCode18, []interface{}{upstreamAssignment, map[string]interface{}{"contextType": upstreamGetter}}),
		upstreamValid(upstreamCode19, []interface{}{upstreamAssignment}),
		upstreamValid(upstreamCode19, []interface{}{upstreamStaticField, map[string]interface{}{"contextType": upstreamAssignment}}),
		upstreamValid(upstreamCode20, nil),
		upstreamValid(upstreamCode20, []interface{}{upstreamAssignment, map[string]interface{}{"displayName": upstreamStaticField}}),
		upstreamValid(upstreamCode21, []interface{}{upstreamGetter}),
		upstreamValid(upstreamCode21, []interface{}{upstreamAssignment, map[string]interface{}{"displayName": upstreamGetter}}),
		upstreamValid(upstreamCode22, []interface{}{upstreamAssignment}),
		upstreamValid(upstreamCode22, []interface{}{upstreamStaticField, map[string]interface{}{"displayName": upstreamAssignment}}),
		upstreamValid(upstreamCode23, nil),
		upstreamValid(upstreamCode23, []interface{}{upstreamAssignment, map[string]interface{}{"defaultProps": upstreamStaticField}}),
		upstreamValid(upstreamCode24, []interface{}{upstreamGetter}),
		upstreamValid(upstreamCode24, []interface{}{upstreamAssignment, map[string]interface{}{"defaultProps": upstreamGetter}}),
		upstreamValid(upstreamCode25, []interface{}{upstreamAssignment}),
		upstreamValid(upstreamCode25, []interface{}{upstreamStaticField, map[string]interface{}{"defaultProps": upstreamAssignment}}),
		upstreamValid(upstreamCode26, nil),
		upstreamValid(upstreamCode26, []interface{}{upstreamAssignment, map[string]interface{}{"propTypes": upstreamStaticField}}),
		upstreamValid(upstreamCode27, []interface{}{upstreamGetter}),
		upstreamValid(upstreamCode27, []interface{}{upstreamAssignment, map[string]interface{}{"propTypes": upstreamGetter}}),
		upstreamValid(upstreamCode28, []interface{}{upstreamAssignment}),
		upstreamValid(upstreamCode28, []interface{}{upstreamStaticField, map[string]interface{}{"propTypes": upstreamAssignment}}),
		upstreamValid(upstreamCode29, nil),
		upstreamValid(upstreamCode29, []interface{}{upstreamAssignment, map[string]interface{}{"childContextTypes": upstreamStaticField, "contextTypes": upstreamStaticField, "contextType": upstreamStaticField, "displayName": upstreamStaticField, "defaultProps": upstreamStaticField, "propTypes": upstreamStaticField}}),
		upstreamValid(upstreamCode30, []interface{}{upstreamGetter}),
		upstreamValid(upstreamCode30, []interface{}{upstreamAssignment, map[string]interface{}{"childContextTypes": upstreamGetter, "contextTypes": upstreamGetter, "contextType": upstreamGetter, "displayName": upstreamGetter, "defaultProps": upstreamGetter, "propTypes": upstreamGetter}}),
		upstreamValid(upstreamCode31, []interface{}{upstreamAssignment}),
		upstreamValid(upstreamCode31, []interface{}{upstreamStaticField, map[string]interface{}{"childContextTypes": upstreamAssignment, "contextTypes": upstreamAssignment, "displayName": upstreamAssignment, "defaultProps": upstreamAssignment, "propTypes": upstreamAssignment}}),
		upstreamValid(upstreamCode32, []interface{}{upstreamStaticField, map[string]interface{}{"displayName": upstreamGetter, "defaultProps": upstreamAssignment, "propTypes": upstreamAssignment}}),
		upstreamValid(upstreamCode32, []interface{}{upstreamAssignment, map[string]interface{}{"childContextTypes": upstreamStaticField, "contextTypes": upstreamStaticField, "displayName": upstreamGetter}}),
		upstreamValid(upstreamCode33, nil),
		upstreamValid(upstreamCode34, nil),
		upstreamValid(upstreamCode35, []interface{}{upstreamStaticField}),
		upstreamValid(upstreamCode36, []interface{}{upstreamStaticField}),
	}
	invalid := []rule_tester.InvalidTestCase{
		upstreamInvalid(upstreamCode37, nil, upstreamError("notStaticClassProp", "childContextTypes"), upstreamError("notStaticClassProp", "contextTypes"), upstreamError("notStaticClassProp", "contextType"), upstreamError("notStaticClassProp", "displayName"), upstreamError("notStaticClassProp", "defaultProps"), upstreamError("notStaticClassProp", "propTypes")),
		upstreamInvalid(upstreamCode37, []interface{}{upstreamAssignment, map[string]interface{}{"childContextTypes": upstreamStaticField, "contextTypes": upstreamStaticField, "contextType": upstreamStaticField, "displayName": upstreamStaticField, "defaultProps": upstreamStaticField, "propTypes": upstreamStaticField}}, upstreamError("notStaticClassProp", "childContextTypes"), upstreamError("notStaticClassProp", "contextTypes"), upstreamError("notStaticClassProp", "contextType"), upstreamError("notStaticClassProp", "displayName"), upstreamError("notStaticClassProp", "defaultProps"), upstreamError("notStaticClassProp", "propTypes")),
		upstreamInvalid(upstreamCode30, nil, upstreamError("notStaticClassProp", "childContextTypes"), upstreamError("notStaticClassProp", "contextTypes"), upstreamError("notStaticClassProp", "contextType"), upstreamError("notStaticClassProp", "displayName"), upstreamError("notStaticClassProp", "defaultProps"), upstreamError("notStaticClassProp", "propTypes")),
		upstreamInvalid(upstreamCode30, []interface{}{upstreamAssignment, map[string]interface{}{"childContextTypes": upstreamStaticField, "contextTypes": upstreamStaticField, "contextType": upstreamStaticField, "displayName": upstreamStaticField, "defaultProps": upstreamStaticField, "propTypes": upstreamStaticField}}, upstreamError("notStaticClassProp", "childContextTypes"), upstreamError("notStaticClassProp", "contextTypes"), upstreamError("notStaticClassProp", "contextType"), upstreamError("notStaticClassProp", "displayName"), upstreamError("notStaticClassProp", "defaultProps"), upstreamError("notStaticClassProp", "propTypes")),
		upstreamInvalid(upstreamCode29, []interface{}{upstreamAssignment}, upstreamError("declareOutsideClass", "childContextTypes"), upstreamError("declareOutsideClass", "contextTypes"), upstreamError("declareOutsideClass", "contextType"), upstreamError("declareOutsideClass", "displayName"), upstreamError("declareOutsideClass", "defaultProps"), upstreamError("declareOutsideClass", "propTypes")),
		upstreamInvalid(upstreamCode29, []interface{}{upstreamStaticField, map[string]interface{}{"childContextTypes": upstreamAssignment, "contextTypes": upstreamAssignment, "contextType": upstreamAssignment, "displayName": upstreamAssignment, "defaultProps": upstreamAssignment, "propTypes": upstreamAssignment}}, upstreamError("declareOutsideClass", "childContextTypes"), upstreamError("declareOutsideClass", "contextTypes"), upstreamError("declareOutsideClass", "contextType"), upstreamError("declareOutsideClass", "displayName"), upstreamError("declareOutsideClass", "defaultProps"), upstreamError("declareOutsideClass", "propTypes")),
		upstreamInvalid(upstreamCode30, []interface{}{upstreamAssignment}, upstreamError("declareOutsideClass", "childContextTypes"), upstreamError("declareOutsideClass", "contextTypes"), upstreamError("declareOutsideClass", "contextType"), upstreamError("declareOutsideClass", "displayName"), upstreamError("declareOutsideClass", "defaultProps"), upstreamError("declareOutsideClass", "propTypes")),
		upstreamInvalid(upstreamCode30, []interface{}{upstreamGetter, map[string]interface{}{"childContextTypes": upstreamAssignment, "contextTypes": upstreamAssignment, "contextType": upstreamAssignment, "displayName": upstreamAssignment, "defaultProps": upstreamAssignment, "propTypes": upstreamAssignment}}, upstreamError("declareOutsideClass", "childContextTypes"), upstreamError("declareOutsideClass", "contextTypes"), upstreamError("declareOutsideClass", "contextType"), upstreamError("declareOutsideClass", "displayName"), upstreamError("declareOutsideClass", "defaultProps"), upstreamError("declareOutsideClass", "propTypes")),
		upstreamInvalid(upstreamCode29, []interface{}{upstreamGetter}, upstreamError("notGetterClassFunc", "childContextTypes"), upstreamError("notGetterClassFunc", "contextTypes"), upstreamError("notGetterClassFunc", "contextType"), upstreamError("notGetterClassFunc", "displayName"), upstreamError("notGetterClassFunc", "defaultProps"), upstreamError("notGetterClassFunc", "propTypes")),
		upstreamInvalid(upstreamCode29, []interface{}{upstreamStaticField, map[string]interface{}{"childContextTypes": upstreamGetter, "contextTypes": upstreamGetter, "contextType": upstreamGetter, "displayName": upstreamGetter, "defaultProps": upstreamGetter, "propTypes": upstreamGetter}}, upstreamError("notGetterClassFunc", "childContextTypes"), upstreamError("notGetterClassFunc", "contextTypes"), upstreamError("notGetterClassFunc", "contextType"), upstreamError("notGetterClassFunc", "displayName"), upstreamError("notGetterClassFunc", "defaultProps"), upstreamError("notGetterClassFunc", "propTypes")),
		upstreamInvalid(upstreamCode37, []interface{}{upstreamGetter}, upstreamError("notGetterClassFunc", "childContextTypes"), upstreamError("notGetterClassFunc", "contextTypes"), upstreamError("notGetterClassFunc", "contextType"), upstreamError("notGetterClassFunc", "displayName"), upstreamError("notGetterClassFunc", "defaultProps"), upstreamError("notGetterClassFunc", "propTypes")),
		upstreamInvalid(upstreamCode37, []interface{}{upstreamAssignment, map[string]interface{}{"childContextTypes": upstreamGetter, "contextTypes": upstreamGetter, "contextType": upstreamGetter, "displayName": upstreamGetter, "defaultProps": upstreamGetter, "propTypes": upstreamGetter}}, upstreamError("notGetterClassFunc", "childContextTypes"), upstreamError("notGetterClassFunc", "contextTypes"), upstreamError("notGetterClassFunc", "contextType"), upstreamError("notGetterClassFunc", "displayName"), upstreamError("notGetterClassFunc", "defaultProps"), upstreamError("notGetterClassFunc", "propTypes")),
		upstreamInvalid(upstreamCode38, []interface{}{upstreamAssignment, map[string]interface{}{"defaultProps": upstreamGetter, "propTypes": upstreamStaticField, "displayName": upstreamStaticField}}, upstreamError("declareOutsideClass", "childContextTypes"), upstreamError("declareOutsideClass", "contextTypes"), upstreamError("declareOutsideClass", "contextType"), upstreamError("notStaticClassProp", "displayName"), upstreamError("notGetterClassFunc", "defaultProps"), upstreamError("notStaticClassProp", "propTypes")),
		upstreamInvalid(upstreamCode38, []interface{}{upstreamGetter, map[string]interface{}{"childContextTypes": upstreamAssignment, "contextTypes": upstreamAssignment, "contextType": upstreamAssignment, "displayName": upstreamAssignment}}, upstreamError("declareOutsideClass", "childContextTypes"), upstreamError("declareOutsideClass", "contextTypes"), upstreamError("declareOutsideClass", "contextType"), upstreamError("declareOutsideClass", "displayName"), upstreamError("notGetterClassFunc", "defaultProps"), upstreamError("notGetterClassFunc", "propTypes")),
		upstreamInvalid(upstreamCode39, []interface{}{upstreamAssignment, map[string]interface{}{"defaultProps": upstreamStaticField, "propTypes": upstreamGetter}}, upstreamError("declareOutsideClass", "childContextTypes"), upstreamError("declareOutsideClass", "contextTypes"), upstreamError("declareOutsideClass", "contextType"), upstreamError("declareOutsideClass", "displayName")),
		upstreamInvalid(upstreamCode40, []interface{}{upstreamAssignment}, upstreamError("declareOutsideClass", "childContextTypes"), upstreamError("declareOutsideClass", "contextTypes"), upstreamError("declareOutsideClass", "contextType"), upstreamError("declareOutsideClass", "displayName"), upstreamError("declareOutsideClass", "contextTypes"), upstreamError("declareOutsideClass", "defaultProps"), upstreamError("declareOutsideClass", "propTypes"), upstreamError("declareOutsideClass", "displayName")),
		upstreamInvalid(upstreamCode41, []interface{}{upstreamStaticField}, upstreamError("notStaticClassProp", "displayName")),
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StaticPropertyPlacementRule, valid, invalid)
}
