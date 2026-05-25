package exhaustive_deps

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestExhaustiveDeps_Upstream_RefCleanup holds the upstream cases whose primary
// diagnostic falls into the "ref_cleanup" category. Cases were routed by
// matching the first expected error's message against a small regex
// table in the test generator (see /tmp/gen_exhaustive_deps_go_tests.js).
// Splitting upstream's monolithic test file by diagnostic kind makes
// it easier to locate a regression: when one diagnostic path drifts,
// the impact is contained to a single file.

var upstreamRefCleanupValid = []rule_tester.ValidTestCase{

}

var upstreamRefCleanupInvalid = []rule_tester.InvalidTestCase{
{
	Code: `
function MyComponent() {
  const myRef = useRef();
  useEffect(() => {
    const handleMove = () => {};
    myRef.current.addEventListener('mousemove', handleMove);
    return () => myRef.current.removeEventListener('mousemove', handleMove);
  }, []);
  return <div ref={myRef} />;
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The ref value 'myRef.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'myRef.current' to a variable inside the effect, and use that variable in the cleanup function."},
	},
},

{
	Code: `
function MyComponent() {
  const myRef = useRef();
  useEffect(() => {
    const handleMove = () => {};
    myRef?.current?.addEventListener('mousemove', handleMove);
    return () => myRef?.current?.removeEventListener('mousemove', handleMove);
  }, []);
  return <div ref={myRef} />;
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The ref value 'myRef.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'myRef.current' to a variable inside the effect, and use that variable in the cleanup function."},
	},
},

{
	Code: `
function MyComponent() {
  const myRef = useRef();
  useEffect(() => {
    const handleMove = () => {};
    myRef.current.addEventListener('mousemove', handleMove);
    return () => myRef.current.removeEventListener('mousemove', handleMove);
  });
  return <div ref={myRef} />;
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The ref value 'myRef.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'myRef.current' to a variable inside the effect, and use that variable in the cleanup function."},
	},
},

{
	Code: `
function useMyThing(myRef) {
  useEffect(() => {
    const handleMove = () => {};
    myRef.current.addEventListener('mousemove', handleMove);
    return () => myRef.current.removeEventListener('mousemove', handleMove);
  }, [myRef]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The ref value 'myRef.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'myRef.current' to a variable inside the effect, and use that variable in the cleanup function."},
	},
},

{
	Code: `
function useMyThing(myRef) {
  useEffect(() => {
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
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The ref value 'myRef.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'myRef.current' to a variable inside the effect, and use that variable in the cleanup function."},
	},
},

{
	Code: `
function useMyThing(myRef, active) {
  useEffect(() => {
    const handleMove = () => {};
    if (active) {
      myRef.current.addEventListener('mousemove', handleMove);
      return function() {
        setTimeout(() => {
          myRef.current.removeEventListener('mousemove', handleMove);
        });
      }
    }
  }, [myRef, active]);
}
`,
	Tsx:  true,
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The ref value 'myRef.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'myRef.current' to a variable inside the effect, and use that variable in the cleanup function."},
	},
},

{
	Code: `
        function MyComponent() {
          const myRef = useRef();
          useLayoutEffect_SAFE_FOR_SSR(() => {
            const handleMove = () => {};
            myRef.current.addEventListener('mousemove', handleMove);
            return () => myRef.current.removeEventListener('mousemove', handleMove);
          });
          return <div ref={myRef} />;
        }
      `,
	Tsx:  true,
	Options: map[string]interface{}{"additionalHooks": "useLayoutEffect_SAFE_FOR_SSR"},
	Errors: []rule_tester.InvalidTestCaseError{
		{Message: "The ref value 'myRef.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'myRef.current' to a variable inside the effect, and use that variable in the cleanup function."},
	},
},
}

func TestExhaustiveDeps_Upstream_RefCleanup(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ExhaustiveDepsRule,
		upstreamRefCleanupValid,
		upstreamRefCleanupInvalid,
	)
}
