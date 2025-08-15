import React, { useState, useEffect } from 'react';
import { getDevelopmentLintResults } from '../utils/data';
import { HeapMap } from './HeatMap';

export default function HeatMapDevelopment() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const data = await getDevelopmentLintResults();
        setData(data);
      } catch (error) {
        console.error('Failed to fetch development lint results:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading || !data) {
    return null;
  }

  return (
    <section className="HeatMap">
      <HeapMap lintResults={data} />
    </section>
  );
}
