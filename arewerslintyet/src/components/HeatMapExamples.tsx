import React, { useState, useEffect } from 'react';
import { getRuleExamplesResults } from '../utils/data';
import HeatMapItem from './HeatMapItem';

export function HeapMapExamples() {
  const [ruleExamplesResult, setRuleExamplesResult] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const ruleExamplesResult = await getRuleExamplesResults();
        setRuleExamplesResult(ruleExamplesResult);
      } catch (error) {
        console.error('Failed to fetch rule examples results:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (
    loading ||
    !ruleExamplesResult ||
    Object.keys(ruleExamplesResult).length === 0
  ) {
    return null;
  }

  let items = [];
  for (const ruleName in ruleExamplesResult) {
    const isPassing = ruleExamplesResult[ruleName];
    items.push(
      <HeatMapItem
        key={ruleName}
        tooltipContent={ruleName}
        href={`https://github.com/rslint/rslint/blob/master/crates/rslint_core/src/groups/rules/${ruleName}.rs`}
        isPassing={isPassing}
      />,
    );
  }

  return (
    <>
      <h2 className="text-4xl my-2">Rule Examples</h2>
      <section className="HeatMap">{items}</section>
    </>
  );
}
