import React, { useState, useCallback } from 'react';
import { useAstInfoContextOptional } from './AstInfoContext';
import type { SignatureInfo } from './types';
import { SignatureInfoView, getSignatureTitle } from './SignatureInfoView';
import { ControlledCollapsibleItem } from './CollapsibleItem';

interface LazySignatureViewProps {
  pos?: number;
  label: string;
  preview: string;
  shallow: SignatureInfo;
}

export const LazySignatureView: React.FC<LazySignatureViewProps> = ({
  pos,
  label,
  preview,
  shallow,
}) => {
  const context = useAstInfoContextOptional();
  const [expanded, setExpanded] = useState(false);
  const [loading, setLoading] = useState(false);
  const [fullData, setFullData] = useState<SignatureInfo | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Check if this signature has a valid position to fetch from
  // pos is always valid now (external files have real positions + fileName)
  const hasValidPos = pos !== undefined;

  const handleToggle = useCallback(async () => {
    if (expanded) {
      setExpanded(false);
      return;
    }

    if (!hasValidPos || !context) {
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
      if (result?.signature) {
        setFullData(result.signature);
      }
      setExpanded(true);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to fetch');
      setExpanded(true);
    } finally {
      setLoading(false);
    }
  }, [expanded, context, pos, shallow.fileName, hasValidPos]);

  // Title: Signature[flags]
  const title = getSignatureTitle(shallow);

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
        <SignatureInfoView info={fullData ?? shallow} />
      )}
    </ControlledCollapsibleItem>
  );
};
