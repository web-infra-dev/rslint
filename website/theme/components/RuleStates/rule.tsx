import React, { useState, ReactNode } from 'react';
import useSWR, { SWRConfig } from 'swr';
import {
  Card,
  CardHeader,
  CardContent as CardBody,
  CardFooter,
} from '@components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@components/ui/table';
import { CancelSymbol, TableSelector } from './table-selector';
import { usePageData } from '@rspress/core/runtime';
import { ErrorCard } from './error';
import { LoadingCard } from './loading';
import { Badge, Heading, Text } from './ui-utils';
import { Button } from '@components/ui/button';

// Type definitions
type FailingCase = {
  name: string;
  url: string;
};

type Rule = {
  name: string;
  group: string;
  status: string;
  failing_case: FailingCase[];
};

type RuleManifest = {
  rules: Rule[];
};

type RuleStateDescribe = {
  name: string;
  count: number;
  style: 'full' | 'partial-impl' | 'partial-test' | 'total';
};

// Constants
const RULE_MANIFEST_URL =
  'https://raw.githubusercontent.com/web-infra-dev/rslint/main/packages/rslint-test-tools/rule-manifest.json';

// Fetcher function
const fetcher = async (url: string): Promise<RuleManifest> => {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error('Failed to fetch rule manifest');
  }
  return response.json();
};

// Statistics card component
const StatCard: React.FC<{ item: RuleStateDescribe }> = ({ item }) => {
  const getTextColor = (style: RuleStateDescribe['style']): string => {
    const colorMap: Record<RuleStateDescribe['style'], string> = {
      total: 'text-gray-900 dark:text-gray-100',
      'partial-impl': 'text-orange-600 dark:text-orange-400',
      'partial-test': 'text-yellow-600 dark:text-yellow-400',
      full: 'text-green-600 dark:text-green-400',
    };
    return colorMap[style] || colorMap.total;
  };

  return (
    <div className="flex flex-col gap-2 p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
      <dd
        className={`text-4xl font-bold tracking-tight ${getTextColor(item.style)}`}
      >
        {item.count}
      </dd>
      <dt className="text-sm text-gray-600 dark:text-gray-400 font-medium">
        {item.name}
      </dt>
    </div>
  );
};

function hooksFilter(setValue: (value: string) => void) {
  return (value: string) => {
    if (value === CancelSymbol) {
      return setValue('');
    }
    return setValue(value);
  };
}

// Main component
const RuleImplementationStatus: React.FC = () => {
  const { data, error, isLoading, mutate } = useSWR<RuleManifest>(
    RULE_MANIFEST_URL,
    fetcher,
  );

  const [searchValue, setSearchValue] = useState('');
  const [groupValue, setGroupValue] = useState('');
  const [statusValue, setStatusValue] = useState('');

  const rulesData = data?.rules || [];

  // Filter rule data
  const filteredRules = rulesData.filter(rule => {
    const matchesSearch = rule.name
      .toLowerCase()
      .includes(searchValue.toLowerCase());
    const matchesGroup = !groupValue || rule.group === groupValue;
    const matchesStatus = !statusValue || rule.status === statusValue;

    return matchesSearch && matchesGroup && matchesStatus;
  });

  // Get available groups and statuses
  const availableGroups = [...new Set(rulesData.map(rule => rule.group))];
  const availableStatuses = [
    { value: 'full', label: 'Full' },
    { value: 'partial-impl', label: 'Partial Implemented' },
    { value: 'partial-test', label: 'Partial Tested' },
  ];

  // Calculate status statistics
  const statusCount: RuleStateDescribe[] = [
    {
      name: 'Fully Implemented',
      count: filteredRules.filter(item => item.status === 'full').length,
      style: 'full',
    },
    {
      name: 'Partial Implemented',
      count: filteredRules.filter(item => item.status === 'partial-impl')
        .length,
      style: 'partial-impl',
    },
    {
      name: 'Partial Tested',
      count: filteredRules.filter(item => item.status === 'partial-test')
        .length,
      style: 'partial-test',
    },
    {
      name: 'Total Rules',
      count: filteredRules.length,
      style: 'total',
    },
  ];

  // Error state
  if (error) {
    return (
      <ErrorCard
        onRetry={() => mutate()}
        title="Rule Implementation Status"
        message="We encountered an issue while loading the rule information."
        retryButtonText="Try Again"
      />
    );
  }

  // Loading state
  if (isLoading && !data) {
    return (
      <LoadingCard title="Rule Status" loadingText="Loading rules data..." />
    );
  }

  return (
    <div className="flex flex-col gap-8">
      {/* Statistics cards */}
      <Card>
        <CardHeader>
          <Heading>Implementation Overview</Heading>
        </CardHeader>
        <CardBody>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {statusCount.map((item, index) => (
              <StatCard key={index} item={item} />
            ))}
          </div>
        </CardBody>
      </Card>

      {/* Rules table */}
      <Card>
        <CardHeader>
          <Heading>All Rules</Heading>
          <TableSelector
            searchValue={searchValue}
            groupValue={groupValue}
            statusValue={statusValue}
            onSearchChange={setSearchValue}
            onGroupChange={hooksFilter(setGroupValue)}
            onStatusChange={hooksFilter(setStatusValue)}
            groups={availableGroups}
            statuses={availableStatuses}
          />
        </CardHeader>
        <CardBody>
          <div className="overflow-x-auto">
            <Table className="w-full">
              <TableHeader>
                <TableRow>
                  <TableHead className="w-1/2">Rule Name</TableHead>
                  <TableHead className="w-1/8">Group</TableHead>
                  <TableHead className="w-1/8">Status</TableHead>
                  <TableHead className="w-1/4">Failing Cases</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredRules.map(rule => (
                  <TableRow key={rule.name} className="transition-colors">
                    <TableCell className="w-1/2 truncate">
                      <Button
                        variant="link"
                        onClick={() => {
                          window.open(
                            `https://typescript-eslint.io/rules/${rule.name}`,
                            '_blank',
                          );
                        }}
                        className="p-0"
                      >
                        <span title={rule.name} className="font-medium">
                          {rule.name}
                        </span>
                      </Button>
                    </TableCell>
                    <TableCell className="w-1/8">
                      <Badge>{rule.group}</Badge>
                    </TableCell>
                    <TableCell className="w-1/8">
                      <Badge>{rule.status}</Badge>
                    </TableCell>
                    <TableCell className="w-1/4">
                      {rule.failing_case.length > 0 ? (
                        <div className="flex flex-col gap-1">
                          {rule.failing_case.map((caseItem, index) => (
                            <Button
                              variant="link"
                              className="cursor-pointer p-0 items-start justify-start h-4"
                              key={caseItem.url}
                              onClick={() => {
                                window.open(
                                  `https://github.com/web-infra-dev/rslint/blob/main/${caseItem.url}`,
                                  '_blank',
                                );
                              }}
                              asChild
                            >
                              <div>
                                {index + 1}.{caseItem.name}
                              </div>
                            </Button>
                          ))}
                        </div>
                      ) : (
                        <Text className="text-gray-500 dark:text-gray-400">
                          No failing cases
                        </Text>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardBody>
      </Card>
    </div>
  );
};

// Export component
export const RuleApp: React.FC = () => {
  const { page } = usePageData();
  const manifest = page.ruleManifest;

  return (
    <SWRConfig
      value={{
        fallback: {
          [RULE_MANIFEST_URL]: manifest,
        },
      }}
    >
      <RuleImplementationStatus />
    </SWRConfig>
  );
};

export default RuleApp;
