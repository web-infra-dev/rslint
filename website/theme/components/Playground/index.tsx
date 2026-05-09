import React, { useEffect, useState } from 'react';
import './index.css';

type PlaygroundComponent = React.ComponentType;

const Playground: React.FC = () => {
  const [Component, setComponent] = useState<PlaygroundComponent | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    import('./Client')
      .then((mod) => {
        if (!cancelled) {
          setComponent(() => mod.default);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : String(err));
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  if (error) {
    return (
      <div className="playground-container">
        <div className="playground-loading">
          Failed to load playground: {error}
        </div>
      </div>
    );
  }

  if (!Component) {
    return null;
  }

  return <Component />;
};

export default Playground;
