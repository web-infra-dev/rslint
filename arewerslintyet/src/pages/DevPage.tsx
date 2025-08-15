import React from 'react';
import { HeapMapExamples } from '../components/HeatMapExamples';
import Footer from '../components/Footer';
import GraphDataDevelopment from '../components/GraphDataDevelopment';
import HeatMapDevelopment from '../components/HeatMapDevelopment';
import IsItReadyDevelopment from '../components/IsItReadyDevelopment';
import ProgressBarDevelopment from '../components/ProgressBarDevelopment';
import { TooltipProvider } from '../components/TooltipContext';

export default function DevPage() {
  return (
    <TooltipProvider>
      {/* Development */}
      <ProgressBarDevelopment />
      <div className="mx-4">
        <IsItReadyDevelopment />
        <GraphDataDevelopment />
        <h2 className="text-4xl my-2">RSLint Tests</h2>
        <HeatMapDevelopment />
        <HeapMapExamples />
      </div>
      <Footer />
    </TooltipProvider>
  );
}
