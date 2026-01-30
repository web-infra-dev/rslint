import React, { useState } from 'react';

interface FlagsDisplayProps {
  flags?: number;
  names?: string[];
}

export const FlagsDisplay: React.FC<FlagsDisplayProps> = ({ flags, names }) => {
  const [showTooltip, setShowTooltip] = useState(false);
  const hasNames = names && names.length > 0;
  const displayValue = flags ?? 0;

  return (
    <span
      className="relative inline-block"
      onMouseEnter={() => hasNames && setShowTooltip(true)}
      onMouseLeave={() => setShowTooltip(false)}
    >
      <span
        className={`font-medium text-blue-900 ${
          hasNames
            ? 'cursor-help border-b border-dashed border-blue-400 hover:border-blue-600'
            : ''
        }`}
      >
        {displayValue}
      </span>
      {showTooltip && hasNames && (
        <div className="absolute bottom-full left-0 z-50 mb-1 min-w-max rounded bg-gray-800 px-2 py-1 text-xs text-white shadow-lg">
          {names.map((name, i) => (
            <div key={i} className="whitespace-nowrap">
              {name}
            </div>
          ))}
        </div>
      )}
    </span>
  );
};
