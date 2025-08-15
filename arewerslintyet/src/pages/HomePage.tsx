import React, { useState, useEffect } from 'react';
import { HeapMapExamples } from '../components/HeatMapExamples';
import Footer from '../components/Footer';
import GraphDataProduction from '../components/GraphDataProduction';
import HeatMapProduction from '../components/HeatMapProduction';
import IsitReadyProduction from '../components/IsitReadyProduction';
import ProgressBarProduction from '../components/ProgressBarProduction';
import { TooltipProvider } from '../components/TooltipContext';

export default function HomePage() {
  return (
    <TooltipProvider>
      {/* Production */}
      <ProgressBarProduction />
      <div className="mx-4">
        <IsitReadyProduction />
        <GraphDataProduction />
        <h2 className="text-4xl my-2">RSLint Tests</h2>
        <HeatMapProduction />
        <HeapMapExamples />
      </div>
      <Footer />
    </TooltipProvider>
  );
}