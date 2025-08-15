import React, { useState, useEffect } from 'react';
import { getLinter } from './bundler';
import { getProductionLintRuns } from './data';
import { ProgressBar } from './ProgressBar';

export default function ProgressBarProduction() {
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
    <ProgressBar linter={getLinter()} mostRecent={mostRecent} dev={false} />
  );
}
