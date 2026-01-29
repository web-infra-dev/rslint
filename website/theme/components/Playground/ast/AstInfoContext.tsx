import React, { createContext, useContext, useCallback } from 'react';
import type { GetAstInfoResponse } from './types';

interface AstInfoContextValue {
  fetchAstInfo: (
    pos: number,
    end?: number,
    kind?: number,
    fileName?: string,
  ) => Promise<GetAstInfoResponse | null>;
  highlightRange: (pos: number, end: number) => void;
  clearHighlight: () => void;
}

const AstInfoContext = createContext<AstInfoContextValue | null>(null);

interface AstInfoProviderProps {
  children: React.ReactNode;
  onRequestAstInfo?: (
    position: number,
    end?: number,
    kind?: number,
    fileName?: string,
  ) => Promise<GetAstInfoResponse | null>;
  onHighlightRange?: (pos: number, end: number) => void;
  onClearHighlight?: () => void;
}

export const AstInfoProvider: React.FC<AstInfoProviderProps> = ({
  children,
  onRequestAstInfo,
  onHighlightRange,
  onClearHighlight,
}) => {
  const fetchAstInfo = useCallback(
    async (
      pos: number,
      end?: number,
      kind?: number,
      fileName?: string,
    ): Promise<GetAstInfoResponse | null> => {
      if (!onRequestAstInfo) {
        return null;
      }
      return onRequestAstInfo(pos, end, kind, fileName);
    },
    [onRequestAstInfo],
  );

  const highlightRange = useCallback(
    (pos: number, end: number) => {
      onHighlightRange?.(pos, end);
    },
    [onHighlightRange],
  );

  const clearHighlight = useCallback(() => {
    onClearHighlight?.();
  }, [onClearHighlight]);

  return (
    <AstInfoContext.Provider
      value={{ fetchAstInfo, highlightRange, clearHighlight }}
    >
      {children}
    </AstInfoContext.Provider>
  );
};

export const useAstInfoContext = () => {
  const context = useContext(AstInfoContext);
  if (!context) {
    throw new Error('useAstInfoContext must be used within AstInfoProvider');
  }
  return context;
};

export const useAstInfoContextOptional = () => {
  return useContext(AstInfoContext);
};
