import React, { useState, useEffect } from 'react';
import { Linter, getLinter } from '../utils/bundler';
import { getProductionLintRuns } from '../utils/data';
import IsItReady from './IsItReady';

export default function IsItReadyProduction() {
  const [mostRecent, setMostRecent] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const { mostRecent } = await getProductionLintRuns();
        setMostRecent(mostRecent);
      } catch (error) {
        console.error('Failed to fetch production lint runs:', error);
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
      title="Production"
      description="lint tests"
      percent={mostRecent.percent}
      decision={mostRecent.percent >= 90 ? 'yes' : 'no'}
    />
  );
}
