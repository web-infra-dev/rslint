import React from 'react';

interface PropRowProps {
  label: string;
  value: React.ReactNode;
}

export const PropRow: React.FC<PropRowProps> = ({ label, value }) => (
  <div className="flex items-center py-0.5 font-mono text-xs">
    <span className="w-4 flex-shrink-0" />
    <span className="text-purple-600">{label}</span>
    <span className="text-gray-500">:</span>
    <span className="ml-1 break-words text-gray-900">{value}</span>
  </div>
);
