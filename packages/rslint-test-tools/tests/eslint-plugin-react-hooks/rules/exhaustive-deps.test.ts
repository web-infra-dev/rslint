import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('exhaustive-deps', {} as never, {
  valid: [
    // No captured values, empty deps
    {
      code: `
        function MyComponent(props) {
          useEffect(() => {});
        }
      `,
    },
    {
      code: `
        function MyComponent(props) {
          useEffect(() => { console.log('hi'); }, []);
        }
      `,
    },
    // Stable hook values: setState / dispatch / useRef.current
    {
      code: `
        function MyComponent() {
          const [, setX] = useState(0);
          useEffect(() => { setX(1); }, []);
        }
      `,
    },
    {
      code: `
        function MyComponent() {
          const ref = useRef(null);
          useEffect(() => { ref.current = 1; }, []);
        }
      `,
    },
    {
      code: `
        function MyComponent() {
          const [, dispatch] = useReducer(reducer, 0);
          useEffect(() => { dispatch({ type: 'a' }); }, []);
        }
      `,
    },
    // Effect referring to its declared dep
    {
      code: `
        function MyComponent({ id }) {
          useEffect(() => { console.log(id); }, [id]);
        }
      `,
    },
    // useEffectEvent — return value is stable, no dep needed
    {
      code: `
        function MyComponent({ theme }) {
          const onClick = useEffectEvent(() => { console.log(theme); });
          useEffect(() => { onClick(); }, []);
        }
      `,
    },
    // Property chain: declaring 'props.foo' covers 'props.foo.bar'
    {
      code: `
        function MyComponent(props) {
          useEffect(() => { console.log(props.foo.bar); }, [props.foo]);
        }
      `,
    },
    // Optional chain
    {
      code: `
        function MyComponent({ user }) {
          useEffect(() => { console.log(user?.name); }, [user?.name]);
        }
      `,
    },
    // External (module-scope) value — not a dep
    {
      code: `
        const CONSTANT = 1;
        function MyComponent() {
          useEffect(() => { console.log(CONSTANT); }, []);
        }
      `,
    },
    // additionalHooks rule option
    {
      code: `
        function MyComponent({ id }) {
          useMyEffect(() => { console.log(id); }, [id]);
        }
      `,
      options: [{ additionalHooks: '(useMyEffect)' }],
    },
  ],
  invalid: [
    // Missing dep
    {
      code: `
        function MyComponent({ id }) {
          useEffect(() => { console.log(id); }, []);
        }
      `,
      errors: [
        {
          message:
            "React Hook useEffect has a missing dependency: 'id'. Either include it or remove the dependency array.",
        },
      ],
    },
    // Missing dep with property chain
    {
      code: `
        function MyComponent(props) {
          useEffect(() => { console.log(props.id); }, []);
        }
      `,
      errors: [
        {
          message:
            "React Hook useEffect has a missing dependency: 'props.id'. Either include it or remove the dependency array.",
        },
      ],
    },
    // Unnecessary dep
    {
      code: `
        function MyComponent({ a, b }) {
          useCallback(() => a, [a, b]);
        }
      `,
      errors: [
        {
          message:
            "React Hook useCallback has an unnecessary dependency: 'b'. Either exclude it or remove the dependency array.",
        },
      ],
    },
    // Duplicate dep
    {
      code: `
        function MyComponent({ a }) {
          useCallback(() => a, [a, a]);
        }
      `,
      errors: [
        {
          message:
            "React Hook useCallback has a duplicate dependency: 'a'. Either omit it or remove the dependency array.",
        },
      ],
    },
    // useMemo without deps array
    {
      code: `
        function MyComponent({ a }) {
          useMemo(() => a);
        }
      `,
      errors: [
        {
          message:
            'React Hook useMemo does nothing when called with only one argument. Did you forget to pass an array of dependencies?',
        },
      ],
    },
    // Spread element
    {
      code: `
        function MyComponent({ list }) {
          useEffect(() => {}, [...list]);
        }
      `,
      errors: [
        {
          message:
            "React Hook useEffect has a spread element in its dependency array. This means we can't statically verify whether you've passed the correct dependencies.",
        },
      ],
    },
    // Non-array deps
    {
      code: `
        function MyComponent({ a }) {
          useEffect(() => {}, a);
        }
      `,
      errors: [
        {
          message:
            "React Hook useEffect was passed a dependency list that is not an array literal. This means we can't statically verify whether you've passed the correct dependencies.",
        },
      ],
    },
    // Literal dep
    {
      code: `
        function MyComponent() {
          useEffect(() => {}, ['foo']);
        }
      `,
      errors: [
        {
          message:
            "The 'foo' literal is not a valid dependency because it never changes. You can safely remove it.",
        },
      ],
    },
    // ref.current in cleanup
    {
      code: `
        function MyComponent() {
          const ref = useRef(null);
          useEffect(() => {
            return () => { console.log(ref.current); };
          }, []);
        }
      `,
      errors: [
        {
          message:
            "The ref value 'ref.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'ref.current' to a variable inside the effect, and use that variable in the cleanup function.",
        },
      ],
    },
    // useEffectEvent return value used in deps
    {
      code: `
        function MyComponent({ theme }) {
          const onClick = useEffectEvent(() => { console.log(theme); });
          useEffect(() => { onClick(); }, [onClick]);
        }
      `,
      errors: [
        {
          message:
            'Functions returned from `useEffectEvent` must not be included in the dependency array. Remove `onClick` from the list.',
        },
      ],
    },
  ],
});
