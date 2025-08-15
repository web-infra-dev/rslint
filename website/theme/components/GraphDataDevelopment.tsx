import React, { useState, useEffect } from 'react';
import { getDevelopmentLintRuns } from './data';
import Graph from './Graph';

export default function GraphDataDevelopment() {
  const [graphData, setGraphData] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const { graphData } = await getDevelopmentLintRuns();
        setGraphData(graphData);
      } catch (error) {
        console.error('Failed to fetch development graph data:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading || graphData.length === 0) {
    return null;
  }

  return <Graph graphData={graphData} />;
}
