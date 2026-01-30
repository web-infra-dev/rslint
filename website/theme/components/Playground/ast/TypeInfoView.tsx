import React from 'react';
import type { TypeInfo } from './types';
import { PropRow } from './PropRow';
import { FlagsDisplay } from './FlagsDisplay';
import {
  LazySymbolView,
  LazyTypeView,
  LazyTypeArray,
  LazySymbolArray,
  LazySignatureArray,
  IndexInfoArray,
} from './LazyViews';

interface TypeInfoViewProps {
  info?: TypeInfo;
}

export const TypeInfoView: React.FC<TypeInfoViewProps> = ({ info }) => {
  if (!info) {
    return (
      <div className="py-2 text-center text-xs text-gray-400">
        No type information
      </div>
    );
  }

  return (
    <div className="font-mono text-xs leading-relaxed">
      <PropRow label="TypeToString()" value={info.typeString} />
      {info.id !== undefined && <PropRow label="Id" value={info.id} />}
      <PropRow
        label="Flags"
        value={<FlagsDisplay flags={info.flags} names={info.flagNames} />}
      />
      {info.objectFlags !== undefined && (
        <PropRow
          label="ObjectFlags"
          value={
            <FlagsDisplay
              flags={info.objectFlags}
              names={info.objectFlagNames}
            />
          }
        />
      )}
      {info.intrinsicName && (
        <PropRow label="IntrinsicName" value={info.intrinsicName} />
      )}

      {info.value !== undefined && (
        <PropRow label="Value" value={JSON.stringify(info.value)} />
      )}

      {info.freshType && (
        <LazyTypeView
          pos={info.freshType.pos}
          label="FreshType"
          preview={info.freshType.typeString}
          shallow={info.freshType}
        />
      )}

      {info.regularType && (
        <LazyTypeView
          pos={info.regularType.pos}
          label="RegularType"
          preview={info.regularType.typeString}
          shallow={info.regularType}
        />
      )}

      {info.symbol && (
        <LazySymbolView
          pos={info.symbol.pos}
          label="Symbol"
          preview={`name: "${info.symbol.name}", flags: ${info.symbol.flags}`}
          shallow={info.symbol}
        />
      )}

      {info.aliasSymbol && (
        <LazySymbolView
          pos={info.aliasSymbol.pos}
          label="AliasSymbol"
          preview={`name: "${info.aliasSymbol.name}", flags: ${info.aliasSymbol.flags}`}
          shallow={info.aliasSymbol}
        />
      )}

      {info.typeArguments && info.typeArguments.length > 0 && (
        <LazyTypeArray label="TypeArguments" items={info.typeArguments} />
      )}

      {info.baseTypes && info.baseTypes.length > 0 && (
        <LazyTypeArray label="BaseTypes" items={info.baseTypes} />
      )}

      {info.types && info.types.length > 0 && (
        <LazyTypeArray label="Types" items={info.types} />
      )}

      {info.properties && info.properties.length > 0 && (
        <LazySymbolArray label="Properties" items={info.properties} />
      )}

      {info.callSignatures && info.callSignatures.length > 0 && (
        <LazySignatureArray
          label="CallSignatures"
          items={info.callSignatures}
        />
      )}

      {info.constructSignatures && info.constructSignatures.length > 0 && (
        <LazySignatureArray
          label="ConstructSignatures"
          items={info.constructSignatures}
        />
      )}

      {info.indexInfos && info.indexInfos.length > 0 && (
        <IndexInfoArray label="IndexInfos" items={info.indexInfos} />
      )}

      {info.constraint && (
        <LazyTypeView
          pos={info.constraint.pos}
          label="Constraint"
          preview={info.constraint.typeString}
          shallow={info.constraint}
        />
      )}

      {info.default && (
        <LazyTypeView
          pos={info.default.pos}
          label="Default"
          preview={info.default.typeString}
          shallow={info.default}
        />
      )}

      {info.target && (
        <LazyTypeView
          pos={info.target.pos}
          label="Target"
          preview={info.target.typeString}
          shallow={info.target}
        />
      )}
    </div>
  );
};

// Helper to get title for top-level display
export const getTypeTitle = (info: TypeInfo) => {
  return `Type[${info.flagNames?.map(n => n.replace('checker.', '')).join(' | ') || 'Unknown'}]`;
};

// Helper to get preview for top-level display
export const getTypePreview = (info: TypeInfo) => {
  return info.typeString;
};
