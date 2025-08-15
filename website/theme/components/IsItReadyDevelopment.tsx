import React, { useState, useEffect } from 'react';
import { Linter, getLinter } from './bundler';
import { getDevelopmentLintRuns } from './data';
import IsItReady from './IsItReady';

export default function IsItReadyDevelopment() {
  const [mostRecent, setMostRecent] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const { mostRecent } = await getDevelopmentLintRuns();
        setMostRecent(mostRecent);
      } catch (error) {
        console.error('Failed to fetch development lint runs:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading || !mostRecent) {
    return null;
  }

  return (
    <IsItReady
      title="Development"
      description="lint tests"
      percent={mostRecent.percent}
      decision={mostRecent.percent >= 90 ? 'yes' : 'no'}
    />
  );
}
