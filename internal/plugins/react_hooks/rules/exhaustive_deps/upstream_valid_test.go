package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDeps_Upstream_Valid holds the upstream cases whose primary
// diagnostic falls into the "valid" category. Cases were routed by
// matching the first expected error's message against a small regex
// table in the test generator (see /tmp/gen_exhaustive_deps_go_tests.js).
// Splitting upstream's monolithic test file by diagnostic kind makes
// it easier to locate a regression: when one diagnostic path drifts,
// the impact is contained to a single file.

var upstreamValidValid = []rule_tester.ValidTestCase{
{
	Code: `
function MyComponent() {
  const local = {};
  useEffect(() => {
    console.log(local);
  });
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  useEffect(() => {
    const local = {};
    console.log(local);
  }, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const local = someFunc();
  useEffect(() => {
    console.log(local);
  }, [local]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  useEffect(() => {
    console.log(props.foo);
  }, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const local1 = {};
  {
    const local2 = {};
    useEffect(() => {
      console.log(local1);
      console.log(local2);
    });
  }
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const local1 = someFunc();
  {
    const local2 = someFunc();
    useCallback(() => {
      console.log(local1);
      console.log(local2);
    }, [local1, local2]);
  }
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const local1 = someFunc();
  function MyNestedComponent() {
    const local2 = someFunc();
    useCallback(() => {
      console.log(local1);
      console.log(local2);
    }, [local2]);
  }
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const local = someFunc();
  useEffect(() => {
    console.log(local);
    console.log(local);
  }, [local]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  useEffect(() => {
    console.log(unresolved);
  }, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const local = someFunc();
  useEffect(() => {
    console.log(local);
  }, [,,,local,,,]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent({ foo }) {
  useEffect(() => {
    console.log(foo.length);
  }, [foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent({ foo }) {
  useEffect(() => {
    console.log(foo.length);
    console.log(foo.slice(0));
  }, [foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent({ history }) {
  useEffect(() => {
    return history.listen();
  }, [history]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {});
  useLayoutEffect(() => {});
  useImperativeHandle(props.innerRef, () => {});
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
  }, [props.foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
    console.log(props.bar);
  }, [props.bar, props.foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
    console.log(props.bar);
  }, [props.foo, props.bar]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  const local = someFunc();
  useEffect(() => {
    console.log(props.foo);
    console.log(props.bar);
    console.log(local);
  }, [props.foo, props.bar, local]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  const local = {};
  useEffect(() => {
    console.log(props.foo);
    console.log(props.bar);
  }, [props, props.foo]);

  let color = someFunc();
  useEffect(() => {
    console.log(props.foo.bar.baz);
    console.log(color);
  }, [props.foo, props.foo.bar.baz, color]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo?.bar?.baz ?? null);
  }, [props.foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo?.bar);
  }, [props.foo?.bar]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo?.bar);
  }, [props.foo.bar]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo.bar);
  }, [props.foo?.bar]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo.bar);
    console.log(props.foo?.bar);
  }, [props.foo?.bar]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo.bar);
    console.log(props.foo?.bar);
  }, [props.foo.bar]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo);
    console.log(props.foo?.bar);
  }, [props.foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo?.toString());
  }, [props.foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useMemo(() => {
    console.log(props.foo?.toString());
  }, [props.foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useCallback(() => {
    console.log(props.foo?.toString());
  }, [props.foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useCallback(() => {
    console.log(props.foo.bar?.toString());
  }, [props.foo.bar]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useCallback(() => {
    console.log(props.foo?.bar?.toString());
  }, [props.foo.bar]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useCallback(() => {
    console.log(props.foo.bar.toString());
  }, [props?.foo?.bar]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useCallback(() => {
    console.log(props.foo?.bar?.baz);
  }, [props?.foo.bar?.baz]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const myEffect = () => {
    // Doesn't use anything
  };
  useEffect(myEffect, []);
}
`,
	Tsx:  true,
},

{
	Code: `
const local = {};
function MyComponent() {
  const myEffect = () => {
    console.log(local);
  };
  useEffect(myEffect, []);
}
`,
	Tsx:  true,
},

{
	Code: `
const local = {};
function MyComponent() {
  function myEffect() {
    console.log(local);
  }
  useEffect(myEffect, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const local = someFunc();
  function myEffect() {
    console.log(local);
  }
  useEffect(myEffect, [local]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  function myEffect() {
    console.log(global);
  }
  useEffect(myEffect, []);
}
`,
	Tsx:  true,
},

{
	Code: `
const local = {};
function MyComponent() {
  const myEffect = () => {
    otherThing()
  }
  const otherThing = () => {
    console.log(local);
  }
  useEffect(myEffect, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent({delay}) {
  const local = {};
  const myEffect = debounce(() => {
    console.log(local);
  }, delay);
  useEffect(myEffect, [myEffect]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent({myEffect}) {
  useEffect(myEffect, [,myEffect]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent({myEffect}) {
  useEffect(myEffect, [,myEffect,,]);
}
`,
	Tsx:  true,
},

{
	Code: `
let local = {};
function myEffect() {
  console.log(local);
}
function MyComponent() {
  useEffect(myEffect, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent({myEffect}) {
  useEffect(myEffect, [myEffect]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent({myEffect}) {
  useEffect(myEffect);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useCustomEffect(() => {
    console.log(props.foo);
  });
}
`,
	Tsx:  true,
	Options: map[string]interface{}{"additionalHooks": "useCustomEffect"},
},

{
	Code: `
function MyComponent(props) {
  useSpecialEffect(() => {
    console.log(props.foo);
  }, null);
}
`,
	Tsx:  true,
	Options: map[string]interface{}{"additionalHooks": "useSpecialEffect", "experimental_autoDependenciesHooks": []interface{}{"useSpecialEffect"}},
},

{
	Code: `
function MyComponent(props) {
  useCustomEffect(() => {
    console.log(props.foo);
  }, [props.foo]);
}
`,
	Tsx:  true,
	Options: map[string]interface{}{"additionalHooks": "useCustomEffect"},
},

{
	Code: `
function MyComponent(props) {
  useCustomEffect(() => {
    console.log(props.foo);
  }, []);
}
`,
	Tsx:  true,
	Options: map[string]interface{}{"additionalHooks": "useAnotherEffect"},
},

{
	Code: `
function MyComponent(props) {
  useWithoutEffectSuffix(() => {
    console.log(props.foo);
  }, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  return renderHelperConfusedWithEffect(() => {
    console.log(props.foo);
  }, []);
}
`,
	Tsx:  true,
},

{
	Code: `
const local = {};
useEffect(() => {
  console.log(local);
}, []);
`,
	Tsx:  true,
},

{
	Code: `
const local1 = {};
{
  const local2 = {};
  useEffect(() => {
    console.log(local1);
    console.log(local2);
  }, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const ref = useRef();
  useEffect(() => {
    console.log(ref.current);
  }, [ref]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const ref = useRef();
  useEffect(() => {
    console.log(ref.current);
  }, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent({ maybeRef2, foo }) {
  const definitelyRef1 = useRef();
  const definitelyRef2 = useRef();
  const maybeRef1 = useSomeOtherRefyThing();
  const [state1, setState1] = useState();
  const [state2, setState2] = React.useState();
  const [state3, dispatch1] = useReducer();
  const [state4, dispatch2] = React.useReducer();
  const [state5, maybeSetState] = useFunnyState();
  const [state6, maybeDispatch] = useFunnyReducer();
  const [state9, dispatch5] = useActionState();
  const [state10, dispatch6] = React.useActionState();
  const [isPending1] = useTransition();
  const [isPending2, startTransition2] = useTransition();
  const [isPending3] = React.useTransition();
  const [isPending4, startTransition4] = React.useTransition();
  const mySetState = useCallback(() => {}, []);
  let myDispatch = useCallback(() => {}, []);

  useEffect(() => {
    // Known to be static
    console.log(definitelyRef1.current);
    console.log(definitelyRef2.current);
    console.log(maybeRef1.current);
    console.log(maybeRef2.current);
    setState1();
    setState2();
    dispatch1();
    dispatch2();
    dispatch5();
    dispatch6();
    startTransition1();
    startTransition2();
    startTransition3();
    startTransition4();

    // Dynamic
    console.log(state1);
    console.log(state2);
    console.log(state3);
    console.log(state4);
    console.log(state5);
    console.log(state6);
    console.log(isPending2);
    console.log(isPending4);
    mySetState();
    myDispatch();

    // Not sure; assume dynamic
    maybeSetState();
    maybeDispatch();
  }, [
    // Dynamic
    state1, state2, state3, state4, state5, state6, state9, state10,
    maybeRef1, maybeRef2,
    isPending2, isPending4,

    // Not sure; assume dynamic
    mySetState, myDispatch,
    maybeSetState, maybeDispatch

    // In this test, we don't specify static deps.
    // That should be okay.
  ]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent({ maybeRef2 }) {
  const definitelyRef1 = useRef();
  const definitelyRef2 = useRef();
  const maybeRef1 = useSomeOtherRefyThing();

  const [state1, setState1] = useState();
  const [state2, setState2] = React.useState();
  const [state3, dispatch1] = useReducer();
  const [state4, dispatch2] = React.useReducer();

  const [state5, maybeSetState] = useFunnyState();
  const [state6, maybeDispatch] = useFunnyReducer();

  const mySetState = useCallback(() => {}, []);
  let myDispatch = useCallback(() => {}, []);

  useEffect(() => {
    // Known to be static
    console.log(definitelyRef1.current);
    console.log(definitelyRef2.current);
    console.log(maybeRef1.current);
    console.log(maybeRef2.current);
    setState1();
    setState2();
    dispatch1();
    dispatch2();

    // Dynamic
    console.log(state1);
    console.log(state2);
    console.log(state3);
    console.log(state4);
    console.log(state5);
    console.log(state6);
    mySetState();
    myDispatch();

    // Not sure; assume dynamic
    maybeSetState();
    maybeDispatch();
  }, [
    // Dynamic
    state1, state2, state3, state4, state5, state6,
    maybeRef1, maybeRef2,

    // Not sure; assume dynamic
    mySetState, myDispatch,
    maybeSetState, maybeDispatch,

    // In this test, we specify static deps.
    // That should be okay too!
    definitelyRef1, definitelyRef2, setState1, setState2, dispatch1, dispatch2
  ]);
}
`,
	Tsx:  true,
},

{
	Code: `
const MyComponent = forwardRef((props, ref) => {
  useImperativeHandle(ref, () => ({
    focus() {
      alert(props.hello);
    }
  }))
});
`,
	Tsx:  true,
},

{
	Code: `
const MyComponent = forwardRef((props, ref) => {
  useImperativeHandle(ref, () => ({
    focus() {
      alert(props.hello);
    }
  }), [props.hello])
});
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  let obj = someFunc();
  useEffect(() => {
    obj.foo = true;
  }, [obj]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  let foo = {}
  useEffect(() => {
    foo.bar.baz = 43;
  }, [foo.bar]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const myRef = useRef();
  useEffect(() => {
    const handleMove = () => {};
    myRef.current = {};
    return () => {
      console.log(myRef.current.toString())
    };
  }, []);
  return <div />;
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const myRef = useRef();
  useEffect(() => {
    const handleMove = () => {};
    myRef.current = {};
    return () => {
      console.log(myRef?.current?.toString())
    };
  }, []);
  return <div />;
}
`,
	Tsx:  true,
},

{
	Code: `
function useMyThing(myRef) {
  useEffect(() => {
    const handleMove = () => {};
    myRef.current = {};
    return () => {
      console.log(myRef.current.toString())
    };
  }, [myRef]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const myRef = useRef();
  useEffect(() => {
    const handleMove = () => {};
    const node = myRef.current;
    node.addEventListener('mousemove', handleMove);
    return () => node.removeEventListener('mousemove', handleMove);
  }, []);
  return <div ref={myRef} />;
}
`,
	Tsx:  true,
},

{
	Code: `
function useMyThing(myRef) {
  useEffect(() => {
    const handleMove = () => {};
    const node = myRef.current;
    node.addEventListener('mousemove', handleMove);
    return () => node.removeEventListener('mousemove', handleMove);
  }, [myRef]);
  return <div ref={myRef} />;
}
`,
	Tsx:  true,
},

{
	Code: `
function useMyThing(myRef) {
  useCallback(() => {
    const handleMouse = () => {};
    myRef.current.addEventListener('mousemove', handleMouse);
    myRef.current.addEventListener('mousein', handleMouse);
    return function() {
      setTimeout(() => {
        myRef.current.removeEventListener('mousemove', handleMouse);
        myRef.current.removeEventListener('mousein', handleMouse);
      });
    }
  }, [myRef]);
}
`,
	Tsx:  true,
},

{
	Code: `
function useMyThing() {
  const myRef = useRef();
  useEffect(() => {
    const handleMove = () => {
      console.log(myRef.current)
    };
    window.addEventListener('mousemove', handleMove);
    return () => window.removeEventListener('mousemove', handleMove);
  }, []);
  return <div ref={myRef} />;
}
`,
	Tsx:  true,
},

{
	Code: `
function useMyThing() {
  const myRef = useRef();
  useEffect(() => {
    const handleMove = () => {
      return () => window.removeEventListener('mousemove', handleMove);
    };
    window.addEventListener('mousemove', handleMove);
    return () => {};
  }, []);
  return <div ref={myRef} />;
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const local1 = 42;
  const local2 = '42';
  const local3 = null;
  useEffect(() => {
    console.log(local1);
    console.log(local2);
    console.log(local3);
  }, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const local1 = 42;
  const local2 = '42';
  const local3 = null;
  useEffect(() => {
    console.log(local1);
    console.log(local2);
    console.log(local3);
  }, [local1, local2, local3]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  const local = props.local;
  useEffect(() => {}, [local]);
}
`,
	Tsx:  true,
},

{
	Code: `
function Foo({ activeTab }) {
  useEffect(() => {
    window.scrollTo(0, 0);
  }, [activeTab]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  useEffect(() => {
    console.log(props.foo.bar.baz);
  }, [props]);
  useEffect(() => {
    console.log(props.foo.bar.baz);
  }, [props.foo]);
  useEffect(() => {
    console.log(props.foo.bar.baz);
  }, [props.foo.bar]);
  useEffect(() => {
    console.log(props.foo.bar.baz);
  }, [props.foo.bar.baz]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  const fn = useCallback(() => {
    console.log(props.foo.bar.baz);
  }, [props]);
  const fn2 = useCallback(() => {
    console.log(props.foo.bar.baz);
  }, [props.foo]);
  const fn3 = useMemo(() => {
    console.log(props.foo.bar.baz);
  }, [props.foo.bar]);
  const fn4 = useMemo(() => {
    console.log(props.foo.bar.baz);
  }, [props.foo.bar.baz]);
}
`,
	Tsx:  true,
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
  }, []);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, []);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  function handleNext() {
    console.log('hello');
  }
  useEffect(() => {
    return Store.subscribe(handleNext);
  }, []);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext);
  }, []);
  useMemo(() => {
    return Store.subscribe(handleNext);
  }, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent(props) {
  let [, setState] = useState();
  let [, dispatch] = React.useReducer();

  function handleNext1(value) {
    let value2 = value * 100;
    setState(value2);
    console.log('hello');
  }
  const handleNext2 = (value) => {
    setState(foo(value));
    console.log('hello');
  };
  let handleNext3 = function(value) {
    console.log(value);
    dispatch({ type: 'x', value });
  };
  useEffect(() => {
    return Store.subscribe(handleNext1);
  }, []);
  useLayoutEffect(() => {
    return Store.subscribe(handleNext2);
  }, []);
  useMemo(() => {
    return Store.subscribe(handleNext3);
  }, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function useInterval(callback, delay) {
  const savedCallback = useRef();
  useEffect(() => {
    savedCallback.current = callback;
  });
  useEffect(() => {
    function tick() {
      savedCallback.current();
    }
    if (delay !== null) {
      let id = setInterval(tick, delay);
      return () => clearInterval(id);
    }
  }, [delay]);
}
`,
	Tsx:  true,
},

{
	Code: `
function Counter() {
  const [count, setCount] = useState(0);

  useEffect(() => {
    let id = setInterval(() => {
      setCount(c => c + 1);
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
},

{
	Code: `
function Counter(unstableProp) {
  let [count, setCount] = useState(0);
  setCount = unstableProp
  useEffect(() => {
    let id = setInterval(() => {
      setCount(c => c + 1);
    }, 1000);
    return () => clearInterval(id);
  }, [setCount]);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
},

{
	Code: `
function Counter() {
  const [count, setCount] = useState(0);

  function tick() {
    setCount(c => c + 1);
  }

  useEffect(() => {
    let id = setInterval(() => {
      tick();
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
},

{
	Code: `
function Counter() {
  const [count, dispatch] = useReducer((state, action) => {
    if (action === 'inc') {
      return state + 1;
    }
  }, 0);

  useEffect(() => {
    let id = setInterval(() => {
      dispatch('inc');
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
},

{
	Code: `
function Counter() {
  const [count, dispatch] = useReducer((state, action) => {
    if (action === 'inc') {
      return state + 1;
    }
  }, 0);

  const tick = () => {
    dispatch('inc');
  };

  useEffect(() => {
    let id = setInterval(tick, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
},

{
	Code: `
function Podcasts() {
  useEffect(() => {
    setPodcasts([]);
  }, []);
  let [podcasts, setPodcasts] = useState(null);
}
`,
	Tsx:  true,
},

{
	Code: `
function withFetch(fetchPodcasts) {
  return function Podcasts({ id }) {
    let [podcasts, setPodcasts] = useState(null);
    useEffect(() => {
      fetchPodcasts(id).then(setPodcasts);
    }, [id]);
  }
}
`,
	Tsx:  true,
},

{
	Code: `
function Podcasts({ id }) {
  let [podcasts, setPodcasts] = useState(null);
  useEffect(() => {
    function doFetch({ fetchPodcasts }) {
      fetchPodcasts(id).then(setPodcasts);
    }
    doFetch({ fetchPodcasts: API.fetchPodcasts });
  }, [id]);
}
`,
	Tsx:  true,
},

{
	Code: `
function Counter() {
  let [count, setCount] = useState(0);

  function increment(x) {
    return x + 1;
  }

  useEffect(() => {
    let id = setInterval(() => {
      setCount(increment);
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
},

{
	Code: `
function Counter() {
  let [count, setCount] = useState(0);

  function increment(x) {
    return x + 1;
  }

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count => increment(count));
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
},

{
	Code: `
import increment from './increment';
function Counter() {
  let [count, setCount] = useState(0);

  useEffect(() => {
    let id = setInterval(() => {
      setCount(count => count + increment);
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return <h1>{count}</h1>;
}
`,
	Tsx:  true,
},

{
	Code: `
function withStuff(increment) {
  return function Counter() {
    let [count, setCount] = useState(0);

    useEffect(() => {
      let id = setInterval(() => {
        setCount(count => count + increment);
      }, 1000);
      return () => clearInterval(id);
    }, []);

    return <h1>{count}</h1>;
  }
}
`,
	Tsx:  true,
},

{
	Code: `
function App() {
  const [query, setQuery] = useState('react');
  const [state, setState] = useState(null);
  useEffect(() => {
    let ignore = false;
    fetchSomething();
    async function fetchSomething() {
      const result = await (await fetch('http://hn.algolia.com/api/v1/search?query=' + query)).json();
      if (!ignore) setState(result);
    }
    return () => { ignore = true; };
  }, [query]);
  return (
    <>
      <input value={query} onChange={e => setQuery(e.target.value)} />
      {JSON.stringify(state)}
    </>
  );
}
`,
	Tsx:  true,
},

{
	Code: `
function Example() {
  const foo = useCallback(() => {
    foo();
  }, []);
}
`,
	Tsx:  true,
},

{
	Code: `
function Example({ prop }) {
  const foo = useCallback(() => {
    if (prop) {
      foo();
    }
  }, [prop]);
}
`,
	Tsx:  true,
},

{
	Code: `
function Hello() {
  const [state, setState] = useState(0);
  useEffect(() => {
    const handleResize = () => setState(window.innerWidth);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  });
}
`,
	Tsx:  true,
},

{
	Code: `
function Example() {
  useEffect(() => {
    arguments
  }, [])
}
`,
	Tsx:  true,
},

{
	Code: `
function Example() {
  useEffect(() => {
    const bar = () => {
      arguments;
    };
    bar();
  }, [])
}
`,
	Tsx:  true,
},

{
	Code: `
function Example(props) {
  useEffect(() => {
    let topHeight = 0;
    topHeight = props.upperViewHeight;
  }, [props.upperViewHeight]);
}
`,
	Tsx:  true,
},

{
	Code: `
function Example(props) {
  useEffect(() => {
    let topHeight = 0;
    topHeight = props?.upperViewHeight;
  }, [props?.upperViewHeight]);
}
`,
	Tsx:  true,
},

{
	Code: `
function Example(props) {
  useEffect(() => {
    let topHeight = 0;
    topHeight = props?.upperViewHeight;
  }, [props]);
}
`,
	Tsx:  true,
},

{
	Code: `
function useFoo(foo){
  return useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function useFoo(){
  const foo = "hi!";
  return useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function useFoo(){
  let {foo} = {foo: 1};
  return useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function useFoo(){
  let [foo] = [1];
  return useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function useFoo() {
  const foo = "fine";
  if (true) {
    // Shadowed variable with constant construction in a nested scope is fine.
    const foo = {};
  }
  return useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent({foo}) {
  return useMemo(() => foo, [foo])
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  const foo = true ? "fine" : "also fine";
  return useMemo(() => foo, [foo]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent() {
  useEffect(() => {
    console.log('banana banana banana');
  }, undefined);
}
`,
	Tsx:  true,
},

// SKIP: unsupported settings shape
{
	Skip: true,
	Code: `
function MyComponent(props) {
  useCustomEffect(() => {
    console.log(props.foo);
  });
}
`,
	Tsx:  true,
},

// SKIP: unsupported settings shape
{
	Skip: true,
	Code: `
function MyComponent(props) {
  useCustomEffect(() => {
    console.log(props.foo);
  }, [props.foo]);
}
`,
	Tsx:  true,
},

// SKIP: unsupported settings shape
{
	Skip: true,
	Code: `
function MyComponent(props) {
  useCustomEffect(() => {
    console.log(props.foo);
  }, []);
}
`,
	Tsx:  true,
	Options: map[string]interface{}{"additionalHooks": "useAnotherEffect"},
},

// SKIP: unsupported settings shape
{
	Skip: true,
	Code: `
function MyComponent(props) {
  useCustomEffect(() => {
    console.log(props.foo);
  }, [props.foo]);
  useAnotherEffect(() => {
    console.log(props.bar);
  }, [props.bar]);
}
`,
	Tsx:  true,
},

{
	Code: `
function MyComponent({ theme }) {
  const onStuff = useEffectEvent(() => {
    showNotification(theme);
  });
  useEffect(() => {
    onStuff();
  }, []);
  React.useEffect(() => {
    onStuff();
  }, []);
}
`,
	Tsx:  true,
},
}

var upstreamValidInvalid = []rule_tester.InvalidTestCase{

}

func TestExhaustiveDeps_Upstream_Valid(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		upstreamValidValid,
		upstreamValidInvalid,
	)
}
