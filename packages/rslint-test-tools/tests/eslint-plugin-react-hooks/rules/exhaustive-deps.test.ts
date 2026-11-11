import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('exhaustive-deps', {
  valid: [
    // Upstream: rule options take precedence over shared settings.
    {
      code: `
        function MyComponent(props) {
          useCustomEffect(() => {
            console.log(props.foo);
          }, []);
        }
      `,
      options: [{ additionalHooks: 'useAnotherEffect' }],
      settings: {
        'react-hooks': { additionalEffectHooks: 'useCustomEffect' },
      },
    },
    // Upstream: automatically inferred dependencies allow an omitted array.
    {
      code: `
        function MyComponent() {
          const [state, setState] = React.useState<number>(0);
          useSpecialEffect(() => {
            const someNumber: typeof state = 2;
            setState(prevState => prevState + someNumber);
          });
        }
      `,
      options: [
        {
          additionalHooks: 'useSpecialEffect',
          experimental_autoDependenciesHooks: ['useSpecialEffect'],
        },
      ],
    },
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
    // (alignment) a local function whose only captures are stable hook values
    // is itself stable when used directly in an effect
    {
      code: `
        function useNoCapture() {
          const [s, setS] = useState(0);
          const transform = (x) => x;
          useEffect(() => { setS(transform(1)); }, []);
          return s;
        }
      `,
    },
  ],
  invalid: [
    // Upstream: shared settings enable dependency checks for custom effects.
    {
      code: `
        function MyComponent(props) {
          useCustomEffect(() => {
            console.log(props.foo);
          }, []);
        }
      `,
      settings: {
        'react-hooks': { additionalEffectHooks: 'useCustomEffect' },
      },
      errors: [
        {
          messageId: '',
          message:
            "React Hook useCustomEffect has a missing dependency: 'props.foo'. Either include it or remove the dependency array.",
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 16,
          suggestions: [
            {
              messageId: '',
              desc: 'Update the dependencies array to be: [props.foo]',
              output: `
        function MyComponent(props) {
          useCustomEffect(() => {
            console.log(props.foo);
          }, [props.foo]);
        }
      `,
            },
          ],
        },
      ],
    },
    // Upstream: requireExplicitEffectDeps rejects an omitted argument.
    {
      code: `
        function MyComponent(props) {
          useEffect(() => {
            console.log(props.foo);
          });
        }
      `,
      options: [{ requireExplicitEffectDeps: true }],
      errors: [
        {
          messageId: '',
          message:
            'React Hook useEffect always requires dependencies. Please add a dependency array or an explicit `undefined`',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
    // Missing dep
    {
      code: `
        function MyComponent({ id }) {
          useEffect(() => { console.log(id); }, []);
        }
      `,
      errors: [
        {
          messageId: '',
          message:
            "React Hook useEffect has a missing dependency: 'id'. Either include it or remove the dependency array.",
          line: 3,
          column: 49,
          endLine: 3,
          endColumn: 51,
          suggestions: [
            {
              messageId: '',
              desc: 'Update the dependencies array to be: [id]',
              output: `
        function MyComponent({ id }) {
          useEffect(() => { console.log(id); }, [id]);
        }
      `,
            },
          ],
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
          messageId: '',
          message:
            "React Hook useEffect has a missing dependency: 'props.id'. Either include it or remove the dependency array.",
          line: 3,
          column: 55,
          endLine: 3,
          endColumn: 57,
          suggestions: [
            {
              messageId: '',
              desc: 'Update the dependencies array to be: [props.id]',
              output: `
        function MyComponent(props) {
          useEffect(() => { console.log(props.id); }, [props.id]);
        }
      `,
            },
          ],
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
          messageId: '',
          message:
            "React Hook useCallback has an unnecessary dependency: 'b'. Either exclude it or remove the dependency array.",
          line: 3,
          column: 32,
          endLine: 3,
          endColumn: 38,
          suggestions: [
            {
              messageId: '',
              desc: 'Update the dependencies array to be: [a]',
              output: `
        function MyComponent({ a, b }) {
          useCallback(() => a, [a]);
        }
      `,
            },
          ],
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
          messageId: '',
          message:
            "React Hook useCallback has a duplicate dependency: 'a'. Either omit it or remove the dependency array.",
          line: 3,
          column: 32,
          endLine: 3,
          endColumn: 38,
          suggestions: [
            {
              messageId: '',
              desc: 'Update the dependencies array to be: [a]',
              output: `
        function MyComponent({ a }) {
          useCallback(() => a, [a]);
        }
      `,
            },
          ],
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
          messageId: '',
          message:
            'React Hook useMemo does nothing when called with only one argument. Did you forget to pass an array of dependencies?',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 18,
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
          messageId: '',
          message:
            "React Hook useEffect has a spread element in its dependency array. This means we can't statically verify whether you've passed the correct dependencies.",
          line: 3,
          column: 32,
          endLine: 3,
          endColumn: 39,
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
          messageId: '',
          message:
            "React Hook useEffect was passed a dependency list that is not an array literal. This means we can't statically verify whether you've passed the correct dependencies.",
          line: 3,
          column: 31,
          endLine: 3,
          endColumn: 32,
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
          messageId: '',
          message:
            "The 'foo' literal is not a valid dependency because it never changes. You can safely remove it.",
          line: 3,
          column: 32,
          endLine: 3,
          endColumn: 37,
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
          messageId: '',
          message:
            "The ref value 'ref.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'ref.current' to a variable inside the effect, and use that variable in the cleanup function.",
          line: 5,
          column: 44,
          endLine: 5,
          endColumn: 51,
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
          messageId: '',
          message:
            'Functions returned from `useEffectEvent` must not be included in the dependency array. Remove `onClick` from the list.',
          line: 4,
          column: 44,
          endLine: 4,
          endColumn: 51,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove the dependency `onClick`',
              output: `
        function MyComponent({ theme }) {
          const onClick = useEffectEvent(() => { console.log(theme); });
          useEffect(() => { onClick(); }, []);
        }
      `,
            },
          ],
        },
      ],
    },
    // (alignment) HOC wrapper (observer/mobx) does not gate analysis
    {
      code: `
        const C = observer((props) => {
          const [s, setS] = useState(0);
          useEffect(() => { console.log(props.x); setS(1); }, []);
          return null;
        });
      `,
      errors: [
        {
          messageId: '',
          message:
            "React Hook useEffect has a missing dependency: 'props.x'. Either include it or remove the dependency array.",
          line: 4,
          column: 63,
          endLine: 4,
          endColumn: 65,
          suggestions: [
            {
              messageId: '',
              desc: 'Update the dependencies array to be: [props.x]',
              output: `
        const C = observer((props) => {
          const [s, setS] = useState(0);
          useEffect(() => { console.log(props.x); setS(1); }, [props.x]);
          return null;
        });
      `,
            },
          ],
        },
      ],
    },
    // (alignment) reactive value captured only via object shorthand inside a local function
    {
      code: `
        function useShorthand({ apId }) {
          const [t, setT] = useState([]);
          const fetchTree = () => { setT([{ apId }]); };
          useEffect(() => { fetchTree(); }, []);
          return t;
        }
      `,
      errors: [
        {
          messageId: '',
          message:
            "React Hook useEffect has a missing dependency: 'fetchTree'. Either include it or remove the dependency array.",
          line: 5,
          column: 45,
          endLine: 5,
          endColumn: 47,
          suggestions: [
            {
              messageId: '',
              desc: 'Update the dependencies array to be: [fetchTree]',
              output: `
        function useShorthand({ apId }) {
          const [t, setT] = useState([]);
          const fetchTree = () => { setT([{ apId }]); };
          useEffect(() => { fetchTree(); }, [fetchTree]);
          return t;
        }
      `,
            },
          ],
        },
      ],
    },
    // (alignment) local function is unstable when it calls another local function
    {
      code: `
        function useCallsLocal({ apId }) {
          const [t, setT] = useState([]);
          const transform = (x) => [x];
          const query = () => { setT(transform(apId)); };
          useEffect(() => { query(); }, []);
          return t;
        }
      `,
      errors: [
        {
          messageId: '',
          message:
            "React Hook useEffect has a missing dependency: 'query'. Either include it or remove the dependency array.",
          line: 6,
          column: 41,
          endLine: 6,
          endColumn: 43,
          suggestions: [
            {
              messageId: '',
              desc: 'Update the dependencies array to be: [query]',
              output: `
        function useCallsLocal({ apId }) {
          const [t, setT] = useState([]);
          const transform = (x) => [x];
          const query = () => { setT(transform(apId)); };
          useEffect(() => { query(); }, [query]);
          return t;
        }
      `,
            },
          ],
        },
      ],
    },
    // (alignment) recursive local function is unstable (self-reference is a capture)
    {
      code: `
        function useRecursive({ tree, id }) {
          const findLabel = (data, target) => {
            for (const node of data) {
              if (node.id === target) return node.label;
              if (node.children) { const r = findLabel(node.children, target); if (r) return r; }
            }
            return '';
          };
          return useMemo(() => findLabel(tree, id), [tree, id]);
        }
      `,
      errors: [
        {
          messageId: '',
          message:
            "React Hook useMemo has a missing dependency: 'findLabel'. Either include it or remove the dependency array.",
          line: 10,
          column: 53,
          endLine: 10,
          endColumn: 63,
          suggestions: [
            {
              messageId: '',
              desc: 'Update the dependencies array to be: [findLabel, tree, id]',
              output: `
        function useRecursive({ tree, id }) {
          const findLabel = (data, target) => {
            for (const node of data) {
              if (node.id === target) return node.label;
              if (node.children) { const r = findLabel(node.children, target); if (r) return r; }
            }
            return '';
          };
          return useMemo(() => findLabel(tree, id), [findLabel, tree, id]);
        }
      `,
            },
          ],
        },
      ],
    },
    // (alignment) ref.current read multiple times in cleanup -> warning at the LAST occurrence
    {
      code: `
        function useRefCleanup() {
          const handleRef = useRef(null);
          useEffect(() => {
            handleRef.current?.addEventListener('a', () => {});
            return () => {
              handleRef.current?.removeEventListener('a', () => {});
              handleRef.current?.removeEventListener('b', () => {});
            };
          }, []);
        }
      `,
      errors: [
        {
          messageId: '',
          message:
            "The ref value 'handleRef.current' will likely have changed by the time this effect cleanup function runs. If this ref points to a node rendered by React, copy 'handleRef.current' to a variable inside the effect, and use that variable in the cleanup function.",
          line: 8,
          column: 25,
          endLine: 8,
          endColumn: 32,
        },
      ],
    },
  ],
});
