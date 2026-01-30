import React, { useState } from 'react';

interface KindDisplayProps {
  kind: number;
  kindName?: string;
}

export const KindDisplay: React.FC<KindDisplayProps> = ({ kind, kindName }) => {
  const [showTooltip, setShowTooltip] = useState(false);
  const hasKindName = kindName && kindName.length > 0;

  return (
    <span
      className="relative inline-block"
      onMouseEnter={() => hasKindName && setShowTooltip(true)}
      onMouseLeave={() => setShowTooltip(false)}
    >
      <span
        className={`font-medium text-blue-900 ${
          hasKindName
            ? 'cursor-help border-b border-dashed border-blue-400 hover:border-blue-600'
            : ''
        }`}
      >
        {kind}
      </span>
      {showTooltip && hasKindName && (
        <div className="absolute bottom-full left-0 z-50 mb-1 min-w-max rounded bg-gray-800 px-2 py-1 text-xs text-white shadow-lg">
          <div className="whitespace-nowrap">{kindName}</div>
        </div>
      )}
    </span>
  );
};
