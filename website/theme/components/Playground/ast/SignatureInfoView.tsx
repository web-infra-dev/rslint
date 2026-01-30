import React from 'react';
import type { SignatureInfo } from './types';
import { PropRow } from './PropRow';
import { FlagsDisplay } from './FlagsDisplay';
import { KindDisplay } from './KindDisplay';
import {
  LazyTypeView,
  LazyNodeView,
  LazySymbolView,
  LazyTypeArray,
  LazySymbolArray,
} from './LazyViews';
import { ObjectDisplay } from './ObjectDisplay';

interface SignatureInfoViewProps {
  info?: SignatureInfo;
}

export const SignatureInfoView: React.FC<SignatureInfoViewProps> = ({
  info,
}) => {
  if (!info) {
    return (
      <div className="py-2 text-center text-xs text-gray-400">
        No signature information
      </div>
    );
  }

  return (
    <div className="font-mono text-xs leading-relaxed">
      <PropRow
        label="Flags"
        value={<FlagsDisplay flags={info.flags} names={info.flagNames} />}
      />
      <PropRow label="MinArgumentCount" value={info.minArgumentCount} />

      {info.parameters && info.parameters.length > 0 && (
        <LazySymbolArray label="Parameters" items={info.parameters} />
      )}

      {info.thisParameter && (
        <LazySymbolView
          pos={info.thisParameter.pos}
          label="ThisParameter"
          preview={`name: "${info.thisParameter.name}"`}
          shallow={info.thisParameter}
        />
      )}

      {info.typeParameters && info.typeParameters.length > 0 && (
        <LazyTypeArray label="TypeParameters" items={info.typeParameters} />
      )}

      {info.returnType && (
        <LazyTypeView
          pos={info.returnType.pos}
          label="ReturnType"
          preview={info.returnType.typeString}
          shallow={info.returnType}
        />
      )}

      {info.typePredicate && (
        <ObjectDisplay
          label="TypePredicate"
          preview={`kind: ${info.typePredicate.kind}${info.typePredicate.parameterName ? `, param: "${info.typePredicate.parameterName}"` : ''}`}
        >
          <div className="font-mono text-xs leading-relaxed">
            <PropRow
              label="Kind"
              value={
                <KindDisplay
                  kind={info.typePredicate.kind}
                  kindName={info.typePredicate.kindName}
                />
              }
            />
            {info.typePredicate.parameterName && (
              <PropRow
                label="ParameterName"
                value={info.typePredicate.parameterName}
              />
            )}
            {info.typePredicate.parameterIndex !== undefined && (
              <PropRow
                label="ParameterIndex"
                value={info.typePredicate.parameterIndex}
              />
            )}
            {info.typePredicate.type && (
              <LazyTypeView
                pos={info.typePredicate.type.pos}
                label="Type"
                preview={info.typePredicate.type.typeString}
                shallow={info.typePredicate.type}
              />
            )}
          </div>
        </ObjectDisplay>
      )}

      {info.declaration && (
        <LazyNodeView
          pos={info.declaration.pos}
          label="Declaration"
          kindName={info.declaration.kindName?.replace('ast.', '')}
          preview={
            info.declaration.pos >= 0
              ? `pos: ${info.declaration.pos}, end: ${info.declaration.end}`
              : ''
          }
          shallow={info.declaration}
        />
      )}
    </div>
  );
};

// Helper to get title for top-level display
export const getSignatureTitle = (info: SignatureInfo) => {
  return `Signature[${info.flagNames?.map(n => n.replace('checker.', '')).join(' | ') || ''}]`;
};

// Helper to get preview for top-level display
export const getSignaturePreview = (info: SignatureInfo) => {
  return `params: ${info.parameters?.length ?? 0}, minArgs: ${info.minArgumentCount}`;
};
