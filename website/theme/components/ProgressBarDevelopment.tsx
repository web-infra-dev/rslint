import React, { useState, useEffect } from 'react';
import { getLinter } from './bundler';
import { getDevelopmentLintRuns } from './data';
import { ProgressBar } from './ProgressBar';

export default function ProgressBarDevelopment() {
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
    <ProgressBar linter={getLinter()} mostRecent={mostRecent} dev={true} />
  );
}
