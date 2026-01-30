import React, { useState, useCallback } from 'react';
import { useAstInfoContextOptional } from './AstInfoContext';
import type { TypeInfo } from './types';
import { TypeInfoView } from './TypeInfoView';
import { ControlledCollapsibleItem } from './CollapsibleItem';

interface LazyTypeViewProps {
  pos?: number;
  label: string;
  preview: string;
  shallow: TypeInfo;
}

export const LazyTypeView: React.FC<LazyTypeViewProps> = ({
  pos,
  label,
  preview,
  shallow,
}) => {
  const context = useAstInfoContextOptional();
  const [expanded, setExpanded] = useState(false);
  const [loading, setLoading] = useState(false);
  const [fullData, setFullData] = useState<TypeInfo | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Check if shallow data is already full data (has id field)
  const isAlreadyFull = shallow.id !== undefined;
  // Check if this type has a valid position to fetch from
  // pos is always valid now (external files have real positions + fileName)
  const hasValidPos = pos !== undefined;

  const handleToggle = useCallback(async () => {
    if (expanded) {
      setExpanded(false);
      return;
    }

    // If already have full data, or no valid position, just expand
    if (isAlreadyFull || !hasValidPos || !context) {
      setExpanded(true);
      return;
    }

    setLoading(true);
    setError(null);
    try {
      // Pass fileName (for external files) to fetch from the correct source file
      const result = await context.fetchAstInfo(
        pos!,
        undefined,
        undefined,
        shallow.fileName,
      );
      if (result?.type) {
        setFullData(result.type);
      }
      setExpanded(true);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to fetch');
      setExpanded(true);
    } finally {
      setLoading(false);
    }
  }, [expanded, context, pos, shallow.fileName, isAlreadyFull, hasValidPos]);

  // Title: Type flags (e.g., "Type[Object]", "Type[String]")
  const title = `Type[${shallow.flagNames?.map(n => n.replace('checker.', '')).join(' | ') || 'Unknown'}]`;

  return (
    <ControlledCollapsibleItem
      label={label}
      title={title}
      preview={preview}
      expanded={expanded}
      loading={loading}
      onToggle={handleToggle}
    >
      {loading ? (
        <div className="text-gray-400 text-xs py-1">Loading...</div>
      ) : error ? (
        <div className="text-red-500 text-xs py-1">{error}</div>
      ) : (
        <TypeInfoView info={fullData ?? shallow} />
      )}
    </ControlledCollapsibleItem>
  );
};
