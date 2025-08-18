import React from 'react';
import { Card, CardContent, CardFooter, CardHeader } from '@components/ui/card';
import { Button } from '@components/ui/button';
import { Heading } from './ui-utils';
import { AlertCircleIcon, RefreshCcw } from 'lucide-react';

export interface ErrorCardProps {
  onRetry: () => void;
  title?: string;
  message?: string;
  retryButtonText?: string;
}

export const ErrorCard: React.FC<ErrorCardProps> = ({
  onRetry,
  title = 'Rule Implementation Status',
  message = 'We encountered an issue while loading the rule information.',
  retryButtonText = 'Try Again',
}) => {
  return (
    <Card className="border-red-200 dark:border-red-800">
      <CardHeader className="text-center">
        <div className="flex items-center justify-center gap-3">
          <div className="w-12 h-12 bg-red-100 dark:bg-red-900/20 rounded-full flex items-center justify-center">
            <AlertCircleIcon className="w-6 h-6 text-red-600 dark:text-red-400" />
          </div>
        </div>
        <Heading>{title}</Heading>
      </CardHeader>
      <CardContent className="text-center space-y-6">
        <div className="space-y-4">
          <div className="text-gray-600 dark:text-gray-400 max-w-md mx-auto">
            <p className="text-base leading-relaxed">{message}</p>
          </div>
        </div>

        <div className="flex justify-center">
          <Button
            onClick={onRetry}
            variant="outline"
            size="lg"
            className="cursor-pointer bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 border-red-200 dark:border-red-700 text-red-700 dark:text-red-300 hover:text-red-800 dark:hover:text-red-200"
          >
            <RefreshCcw className="w-4 h-4 mr-2" />
            {retryButtonText}
          </Button>
        </div>
      </CardContent>
      <CardFooter className="pt-0"></CardFooter>
    </Card>
  );
};
