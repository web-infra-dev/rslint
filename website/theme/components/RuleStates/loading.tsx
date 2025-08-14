import React from 'react';
import { Card, CardContent, CardHeader } from '@components/ui/card';

export interface LoadingCardProps {
  title?: string;
  loadingText?: string;
}

export const LoadingCard: React.FC<LoadingCardProps> = ({
  title = 'Rule Status',
  loadingText = 'Loading rules data...',
}) => {
  return (
    <Card>
      <CardHeader>
        <p className="scroll-m-20 border-b pb-2 text-xl tracking-tight first:mt-0">
          {title}
        </p>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <div className="leading-7 [&:not(:first-child)]:mt-1 ml-3">
            {loadingText}
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
