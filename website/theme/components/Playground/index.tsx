import React, { useEffect, useState } from 'react';

// sideEffects
export default function Playground() {
  if (import.meta.env.SSG_MD) {
    return `
# Playground

This is a rslint playground page
`;
  }
  const [element, setElement] = useState<React.ReactNode>(null);
  useEffect(() => {
    import('./Playground').then(({ default: Playground }) => {
      setElement(<Playground />);
    });
  }, []);
  return <>{element}</>;
}
