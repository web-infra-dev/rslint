package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDeps_Upstream_Construction holds the upstream cases whose primary
// diagnostic falls into the "construction" category. Cases were routed by
// matching the first expected error's message against a small regex
// table in the test generator (see /tmp/gen_exhaustive_deps_go_tests.js).
// Splitting upstream's monolithic test file by diagnostic kind makes
// it easier to locate a regression: when one diagnostic path drifts,
// the impact is contained to a single file.

var upstreamConstructionValid = []rule_tester.ValidTestCase{

}

var upstreamConstructionInvalid = []rule_tester.InvalidTestCase{
{
	Code: `
function MyComponent(props) {
  let [, setState] = useState();

  function handleNext(value) {
    setState(value);
  }

  useEffect(() => {
    return Store.subscribe(handleNext);
  }, [handleNext]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'handleNext' function makes the dependencies of useEffect Hook (at line 11) change on every render. Move it inside the useEffect callback. Alternatively, wrap the definition of 'handleNext' in its own useCallback() Hook."},
	},
},

{
	Code: `
function MyComponent(props) {
  let [, setState] = useState();

  const handleNext = (value) => {
    setState(value);
  };

  useEffect(() => {
    return Store.subscribe(handleNext);
  }, [handleNext]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'handleNext' function makes the dependencies of useEffect Hook (at line 11) change on every render. Move it inside the useEffect callback. Alternatively, wrap the definition of 'handleNext' in its own useCallback() Hook."},
	},
},

{
	Code: `
function MyComponent(props) {
  let [, setState] = useState();

  const handleNext = (value) => {
    setState(value);
  };

  useEffect(() => {
    return Store.subscribe(handleNext);
  }, [handleNext]);

  return <div onClick={handleNext} />;
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'handleNext' function makes the dependencies of useEffect Hook (at line 11) change on every render. To fix this, wrap the definition of 'handleNext' in its own useCallback() Hook.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let [, setState] = useState();

  const handleNext = useCallback((value) => {
    setState(value);
  });

  useEffect(() => {
    return Store.subscribe(handleNext);
  }, [handleNext]);

  return <div onClick={handleNext} />;
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  function handleNext1() {
    console.log('hello');
  }
  const handleNext2 = () => {
    console.log('hello');
  };
  let handleNext3 = function() {
    console.log('hello');
  };
  useEffect(() => {
    return Store.subscribe(handleNext1);
  }, [handleNext1]);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, [handleNext2]);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, [handleNext3]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'handleNext1' function makes the dependencies of useEffect Hook (at line 14) change on every render. Move it inside the useEffect callback. Alternatively, wrap the definition of 'handleNext1' in its own useCallback() Hook."},
		{Message: "The 'handleNext2' function makes the dependencies of useLayoutEffect Hook (at line 17) change on every render. Move it inside the useLayoutEffect callback. Alternatively, wrap the definition of 'handleNext2' in its own useCallback() Hook."},
		{Message: "The 'handleNext3' function makes the dependencies of useMemo Hook (at line 20) change on every render. Move it inside the useMemo callback. Alternatively, wrap the definition of 'handleNext3' in its own useCallback() Hook."},
	},
},

{
	Code: `
function MyComponent(props) {
  function handleNext1() {
    console.log('hello');
  }
  const handleNext2 = () => {
    console.log('hello');
  };
  let handleNext3 = function() {
    console.log('hello');
  };
  useEffect(() => {
    handleNext1();
    return Store.subscribe(() => handleNext1());
  }, [handleNext1]);
  useLayoutEffect(() => {
    handleNext2();
    return Store.subscribe(() => handleNext2());
  }, [handleNext2]);
  useMemo(() => {
    handleNext3();
    return Store.subscribe(() => handleNext3());
  }, [handleNext3]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'handleNext1' function makes the dependencies of useEffect Hook (at line 15) change on every render. Move it inside the useEffect callback. Alternatively, wrap the definition of 'handleNext1' in its own useCallback() Hook."},
		{Message: "The 'handleNext2' function makes the dependencies of useLayoutEffect Hook (at line 19) change on every render. Move it inside the useLayoutEffect callback. Alternatively, wrap the definition of 'handleNext2' in its own useCallback() Hook."},
		{Message: "The 'handleNext3' function makes the dependencies of useMemo Hook (at line 23) change on every render. Move it inside the useMemo callback. Alternatively, wrap the definition of 'handleNext3' in its own useCallback() Hook."},
	},
},

{
	Code: `
function MyComponent(props) {
  function handleNext1() {
    console.log('hello');
  }
  const handleNext2 = () => {
    console.log('hello');
  };
  let handleNext3 = function() {
    console.log('hello');
  };
  useEffect(() => {
    handleNext1();
    return Store.subscribe(() => handleNext1());
  }, [handleNext1]);
  useLayoutEffect(() => {
    handleNext2();
    return Store.subscribe(() => handleNext2());
  }, [handleNext2]);
  useMemo(() => {
    handleNext3();
    return Store.subscribe(() => handleNext3());
  }, [handleNext3]);
  return (
    <div
      onClick={() => {
        handleNext1();
        setTimeout(handleNext2);
        setTimeout(() => {
          handleNext3();
        });
      }}
    />
  );
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'handleNext1' function makes the dependencies of useEffect Hook (at line 15) change on every render. To fix this, wrap the definition of 'handleNext1' in its own useCallback() Hook."},
		{Message: "The 'handleNext2' function makes the dependencies of useLayoutEffect Hook (at line 19) change on every render. To fix this, wrap the definition of 'handleNext2' in its own useCallback() Hook.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  function handleNext1() {
    console.log('hello');
  }
  const handleNext2 = useCallback(() => {
    console.log('hello');
  });
  let handleNext3 = function() {
    console.log('hello');
  };
  useEffect(() => {
    handleNext1();
    return Store.subscribe(() => handleNext1());
  }, [handleNext1]);
  useLayoutEffect(() => {
    handleNext2();
    return Store.subscribe(() => handleNext2());
  }, [handleNext2]);
  useMemo(() => {
    handleNext3();
    return Store.subscribe(() => handleNext3());
  }, [handleNext3]);
  return (
    <div
      onClick={() => {
        handleNext1();
        setTimeout(handleNext2);
        setTimeout(() => {
          handleNext3();
        });
      }}
    />
  );
}
`}}},
		{Message: "The 'handleNext3' function makes the dependencies of useMemo Hook (at line 23) change on every render. To fix this, wrap the definition of 'handleNext3' in its own useCallback() Hook.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  function handleNext1() {
    console.log('hello');
  }
  const handleNext2 = () => {
    console.log('hello');
  };
  let handleNext3 = useCallback(function() {
    console.log('hello');
  });
  useEffect(() => {
    handleNext1();
    return Store.subscribe(() => handleNext1());
  }, [handleNext1]);
  useLayoutEffect(() => {
    handleNext2();
    return Store.subscribe(() => handleNext2());
  }, [handleNext2]);
  useMemo(() => {
    handleNext3();
    return Store.subscribe(() => handleNext3());
  }, [handleNext3]);
  return (
    <div
      onClick={() => {
        handleNext1();
        setTimeout(handleNext2);
        setTimeout(() => {
          handleNext3();
        });
      }}
    />
  );
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  const handleNext1 = () => {
    console.log('hello');
  };
  function handleNext2() {
    console.log('hello');
  }
  useEffect(() => {
    return Store.subscribe(handleNext1);
    return Store.subscribe(handleNext2);
  }, [handleNext1, handleNext2]);
  useEffect(() => {
    return Store.subscribe(handleNext1);
    return Store.subscribe(handleNext2);
  }, [handleNext1, handleNext2]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'handleNext1' function makes the dependencies of useEffect Hook (at line 12) change on every render. To fix this, wrap the definition of 'handleNext1' in its own useCallback() Hook.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const handleNext1 = useCallback(() => {
    console.log('hello');
  });
  function handleNext2() {
    console.log('hello');
  }
  useEffect(() => {
    return Store.subscribe(handleNext1);
    return Store.subscribe(handleNext2);
  }, [handleNext1, handleNext2]);
  useEffect(() => {
    return Store.subscribe(handleNext1);
    return Store.subscribe(handleNext2);
  }, [handleNext1, handleNext2]);
}
`}}},
		{Message: "The 'handleNext2' function makes the dependencies of useEffect Hook (at line 12) change on every render. To fix this, wrap the definition of 'handleNext2' in its own useCallback() Hook."},
		{Message: "The 'handleNext1' function makes the dependencies of useEffect Hook (at line 16) change on every render. To fix this, wrap the definition of 'handleNext1' in its own useCallback() Hook.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  const handleNext1 = useCallback(() => {
    console.log('hello');
  });
  function handleNext2() {
    console.log('hello');
  }
  useEffect(() => {
    return Store.subscribe(handleNext1);
    return Store.subscribe(handleNext2);
  }, [handleNext1, handleNext2]);
  useEffect(() => {
    return Store.subscribe(handleNext1);
    return Store.subscribe(handleNext2);
  }, [handleNext1, handleNext2]);
}
`}}},
		{Message: "The 'handleNext2' function makes the dependencies of useEffect Hook (at line 16) change on every render. To fix this, wrap the definition of 'handleNext2' in its own useCallback() Hook."},
	},
},

{
	Code: `
function MyComponent(props) {
  let handleNext = () => {
    console.log('hello');
  };
  if (props.foo) {
    handleNext = () => {
      console.log('hello');
    };
  }
  useEffect(() => {
    return Store.subscribe(handleNext);
  }, [handleNext]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'handleNext' function makes the dependencies of useEffect Hook (at line 13) change on every render. To fix this, wrap the definition of 'handleNext' in its own useCallback() Hook.", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `
function MyComponent(props) {
  let handleNext = useCallback(() => {
    console.log('hello');
  });
  if (props.foo) {
    handleNext = () => {
      console.log('hello');
    };
  }
  useEffect(() => {
    return Store.subscribe(handleNext);
  }, [handleNext]);
}
`}}},
	},
},

{
	Code: `
function MyComponent(props) {
  let [, setState] = useState();
  let taint = props.foo;

  function handleNext(value) {
    let value2 = value * taint;
    setState(value2);
    console.log('hello');
  }

  useEffect(() => {
    return Store.subscribe(handleNext);
  }, [handleNext]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'handleNext' function makes the dependencies of useEffect Hook (at line 14) change on every render. Move it inside the useEffect callback. Alternatively, wrap the definition of 'handleNext' in its own useCallback() Hook."},
	},
},

{
	Code: `
function Counter({ step }) {
  let [count, setCount] = useState(0);

  function increment(x) {
    return x + step;
  }

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count => increment(count));
    }, 1000);
    return () => clearInterval(id);
  }, [increment]);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'increment' function makes the dependencies of useEffect Hook (at line 14) change on every render. Move it inside the useEffect callback. Alternatively, wrap the definition of 'increment' in its own useCallback() Hook."},
	},
},

{
	Code: `
function Component() {
  const foo = {};
  useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' object makes the dependencies of useMemo Hook (at line 4) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Component() {
  const foo = [];
  useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' array makes the dependencies of useMemo Hook (at line 4) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Component() {
  const foo = () => {};
  useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' function makes the dependencies of useMemo Hook (at line 4) change on every render. Move it inside the useMemo callback. Alternatively, wrap the definition of 'foo' in its own useCallback() Hook."},
	},
},

{
	Code: `
function Component() {
  const foo = function bar(){};
  useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' function makes the dependencies of useMemo Hook (at line 4) change on every render. Move it inside the useMemo callback. Alternatively, wrap the definition of 'foo' in its own useCallback() Hook."},
	},
},

{
	Code: `
function Component() {
  const foo = class {};
  useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' class makes the dependencies of useMemo Hook (at line 4) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Component() {
  const foo = true ? {} : "fine";
  useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' conditional could make the dependencies of useMemo Hook (at line 4) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Component() {
  const foo = bar || {};
  useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' logical expression could make the dependencies of useMemo Hook (at line 4) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Component() {
  const foo = bar ?? {};
  useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' logical expression could make the dependencies of useMemo Hook (at line 4) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Component() {
  const foo = bar && {};
  useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' logical expression could make the dependencies of useMemo Hook (at line 4) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Component() {
  const foo = bar ? baz ? {} : null : null;
  useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' conditional could make the dependencies of useMemo Hook (at line 4) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Component() {
  let foo = {};
  useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' object makes the dependencies of useMemo Hook (at line 4) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Component() {
  var foo = {};
  useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' object makes the dependencies of useMemo Hook (at line 4) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Component() {
  const foo = {};
  useCallback(() => {
    console.log(foo);
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' object makes the dependencies of useCallback Hook (at line 6) change on every render. Move it inside the useCallback callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Component() {
  const foo = {};
  useEffect(() => {
    console.log(foo);
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' object makes the dependencies of useEffect Hook (at line 6) change on every render. Move it inside the useEffect callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Component() {
  const foo = {};
  useLayoutEffect(() => {
    console.log(foo);
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' object makes the dependencies of useLayoutEffect Hook (at line 6) change on every render. Move it inside the useLayoutEffect callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Component() {
  const foo = {};
  useImperativeHandle(
    ref,
    () => {
       console.log(foo);
    },
    [foo]
  );
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' object makes the dependencies of useImperativeHandle Hook (at line 9) change on every render. Move it inside the useImperativeHandle callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Foo(section) {
  const foo = section.section_components?.edges ?? [];
  useEffect(() => {
    console.log(foo);
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' logical expression could make the dependencies of useEffect Hook (at line 6) change on every render. Move it inside the useEffect callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Foo(section) {
  const foo = {};
  console.log(foo);
  useMemo(() => {
    console.log(foo);
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' object makes the dependencies of useMemo Hook (at line 7) change on every render. To fix this, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Foo() {
  const foo = <>Hi!</>;
  useMemo(() => {
    console.log(foo);
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' JSX fragment makes the dependencies of useMemo Hook (at line 6) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Foo() {
  const foo = <div>Hi!</div>;
  useMemo(() => {
    console.log(foo);
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' JSX element makes the dependencies of useMemo Hook (at line 6) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Foo() {
  const foo = bar = {};
  useMemo(() => {
    console.log(foo);
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' assignment expression makes the dependencies of useMemo Hook (at line 6) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Foo() {
  const foo = new String('foo'); // Note 'foo' will be boxed, and thus an object and thus compared by reference.
  useMemo(() => {
    console.log(foo);
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' object construction makes the dependencies of useMemo Hook (at line 6) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Foo() {
  const foo = new Map([]);
  useMemo(() => {
    console.log(foo);
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' object construction makes the dependencies of useMemo Hook (at line 6) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Foo() {
  const foo = /reg/;
  useMemo(() => {
    console.log(foo);
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' regular expression makes the dependencies of useMemo Hook (at line 6) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Foo() {
  class Bar {};
  useMemo(() => {
    console.log(new Bar());
  }, [Bar]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'Bar' class makes the dependencies of useMemo Hook (at line 6) change on every render. Move it inside the useMemo callback. Alternatively, wrap the initialization of 'Bar' in its own useMemo() Hook."},
	},
},

{
	Code: `
function Foo() {
  const foo = {};
  useLayoutEffect(() => {
    console.log(foo);
  }, [foo]);
  useEffect(() => {
    console.log(foo);
  }, [foo]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The 'foo' object makes the dependencies of useLayoutEffect Hook (at line 6) change on every render. To fix this, wrap the initialization of 'foo' in its own useMemo() Hook."},
		{Message: "The 'foo' object makes the dependencies of useEffect Hook (at line 9) change on every render. To fix this, wrap the initialization of 'foo' in its own useMemo() Hook."},
	},
},
}

func TestExhaustiveDeps_Upstream_Construction(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		upstreamConstructionValid,
		upstreamConstructionInvalid,
	)
}
