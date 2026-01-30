import React, { useState } from 'react';

interface ObjectDisplayProps {
  label: string;
  preview: string;
  children: React.ReactNode;
  kindName?: string; // For Node types, display kind name after label
  defaultExpanded?: boolean;
  onLabelClick?: () => void;
}

export const ObjectDisplay: React.FC<ObjectDisplayProps> = ({
  label,
  preview,
  children,
  kindName,
  defaultExpanded = false,
  onLabelClick,
}) => {
  const [expanded, setExpanded] = useState(defaultExpanded);

  const handleToggle = () => setExpanded(!expanded);

  const handleLabelClick = (e: React.MouseEvent) => {
    if (onLabelClick) {
      e.stopPropagation();
      onLabelClick();
    } else {
      handleToggle();
    }
  };

  return (
    <div className="font-mono text-xs leading-relaxed">
      <div className="flex items-center py-0.5">
        <span
          className="text-[10px] text-gray-500 w-4 flex-shrink-0 text-center cursor-pointer hover:bg-gray-200 rounded"
          onClick={handleToggle}
        >
          {expanded ? '▼' : '▶'}
        </span>
        <span
          className={`text-purple-600 cursor-pointer hover:underline ${onLabelClick ? 'hover:text-purple-800' : ''}`}
          onClick={handleLabelClick}
        >
          {label}
        </span>
        <span className="text-gray-500">:</span>
        {kindName && (
          <span className="text-blue-600 ml-1 underline decoration-dotted">
            {kindName}
          </span>
        )}
        {!expanded && (
          <span className="text-gray-400 ml-1 truncate max-w-[300px]">
            {'{'}
            {preview}
            {'}'}
          </span>
        )}
        {expanded && <span className="text-gray-500 ml-1">{'{'}</span>}
      </div>
      {expanded && (
        <>
          <div className="ml-4 py-1 border-l border-gray-200 pl-2">
            {children}
          </div>
          <div className="text-gray-500 ml-4">{'}'}</div>
        </>
      )}
    </div>
  );
};
