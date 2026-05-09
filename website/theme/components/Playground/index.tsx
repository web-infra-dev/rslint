import { useEffect, useState, type ReactNode } from 'react';

const PLAYGROUND_SSG_MARKDOWN = `
# Playground

The rslint Playground is an interactive browser-only tool for trying rules, editing configuration, viewing diagnostics, and inspecting AST output.

Open this page in a browser to use the editor and result panels.
`;

// sideEffects
export default function Playground() {
  if (import.meta.env.SSG_MD) {
    return PLAYGROUND_SSG_MARKDOWN;
  }
  const [element, setElement] = useState<ReactNode>(null);
  useEffect(() => {
    let mounted = true;

    import('./Playground').then(({ default: Playground }) => {
      if (mounted) {
        setElement(<Playground />);
      }
    });

    return () => {
      mounted = false;
    };
  }, []);
  return <>{element}</>;
}
