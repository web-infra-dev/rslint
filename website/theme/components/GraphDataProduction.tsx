import React, { useState, useEffect } from 'react';
import { getProductionLintRuns } from './data';
import Graph from './Graph';

export default function GraphDataProduction() {
  const [graphData, setGraphData] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const { graphData } = await getProductionLintRuns();
        setGraphData(graphData);
      } catch (error) {
        console.error('Failed to fetch production graph data:', error);
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
