import React, { useState, useEffect } from 'react';
import { Card, CardHeader, CardContent as CardBody } from '@components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@components/ui/table';
import { CancelSymbol, TableSelector } from './table-selector';
import { Badge, Heading, PresetBadge, Text } from './ui-utils';
import { Button } from '@components/ui/button';
import manifest from '@/generated/rule-manifest.json';

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
  docPath: string | null;
  presets: { name: string; value: unknown }[];
};

type RuleStateDescribe = {
  name: string;
  count: number;
  style: 'full' | 'partial-impl' | 'partial-test' | 'total';
};

function groupToRouteSlug(group: string): string {
  return group.replace(/^@/, '');
}

function getRuleUrl(rule: Rule): { url: string; isInternal: boolean } {
  if (rule.docPath) {
    const slug = groupToRouteSlug(rule.group);
    return { url: `/rules/${slug}/${rule.name}`, isInternal: true };
  }
  // Fallback to external docs for rules without local documentation
  if (rule.group === '@typescript-eslint') {
    return {
      url: `https://typescript-eslint.io/rules/${rule.name}`,
      isInternal: false,
    };
  }
  return {
    url: `https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/${rule.name}.md`,
    isInternal: false,
  };
}

// --- URL params helpers ---

function getInitialParam(key: string): string {
  if (typeof window === 'undefined') return '';
  const params = new URLSearchParams(window.location.search);
  return params.get(key) || '';
}

function syncFiltersToUrl(filters: Record<string, string>): void {
  if (typeof window === 'undefined') return;
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(filters)) {
    if (value) params.set(key, value);
  }
  const search = params.toString();
  const newUrl = search
    ? `${window.location.pathname}?${search}`
    : window.location.pathname;
  window.history.replaceState(null, '', newUrl);
}

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
  const [searchValue, setSearchValue] = useState(() =>
    getInitialParam('search'),
  );
  const [groupValue, setGroupValue] = useState(() => getInitialParam('group'));
  const [statusValue, setStatusValue] = useState(() =>
    getInitialParam('status'),
  );
  const [presetValue, setPresetValue] = useState(() =>
    getInitialParam('preset'),
  );

  // Sync filters to URL whenever they change
  useEffect(() => {
    syncFiltersToUrl({
      search: searchValue,
      group: groupValue,
      status: statusValue,
      preset: presetValue,
    });
  }, [searchValue, groupValue, statusValue, presetValue]);

  const rulesData = manifest.rules as Rule[];

  // Filter rule data
  const filteredRules = rulesData.filter((rule) => {
    const matchesSearch = rule.name
      .toLowerCase()
      .includes(searchValue.toLowerCase());
    const matchesGroup = !groupValue || rule.group === groupValue;
    const matchesStatus = !statusValue || rule.status === statusValue;
    const matchesPreset =
      !presetValue ||
      (rule.presets && rule.presets.some((p) => p.name === presetValue));

    return matchesSearch && matchesGroup && matchesStatus && matchesPreset;
  });

  // Get available groups and statuses
  const availableGroups = [...new Set(rulesData.map((rule) => rule.group))];
  const availableStatuses = [
    { value: 'full', label: 'Full' },
    { value: 'partial-impl', label: 'Partial Implemented' },
    { value: 'partial-test', label: 'Partial Tested' },
  ];

  // Build preset filter options from levels that have at least one rule
  const availablePresets = [
    ...new Set(rulesData.flatMap((r) => (r.presets || []).map((p) => p.name))),
  ]
    .sort()
    .map((level) => ({ value: level, label: level }));

  // Calculate status statistics
  const statusCount: RuleStateDescribe[] = [
    {
      name: 'Fully Implemented',
      count: filteredRules.filter((item) => item.status === 'full').length,
      style: 'full',
    },
    {
      name: 'Partial Implemented',
      count: filteredRules.filter((item) => item.status === 'partial-impl')
        .length,
      style: 'partial-impl',
    },
    {
      name: 'Partial Tested',
      count: filteredRules.filter((item) => item.status === 'partial-test')
        .length,
      style: 'partial-test',
    },
    {
      name: 'Total Rules',
      count: filteredRules.length,
      style: 'total',
    },
  ];

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
            presetValue={presetValue}
            onSearchChange={setSearchValue}
            onGroupChange={hooksFilter(setGroupValue)}
            onStatusChange={hooksFilter(setStatusValue)}
            onPresetChange={hooksFilter(setPresetValue)}
            groups={availableGroups}
            statuses={availableStatuses}
            presetOptions={availablePresets}
          />
        </CardHeader>
        <CardBody>
          <div className="overflow-x-auto">
            <Table className="w-full">
              <TableHeader>
                <TableRow>
                  <TableHead className="w-2/6">Rule Name</TableHead>
                  <TableHead className="w-1/6">Group</TableHead>
                  <TableHead className="w-1/6">Preset</TableHead>
                  <TableHead className="w-1/12">Status</TableHead>
                  <TableHead className="w-1/6">Failing Cases</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredRules.map((rule) => (
                  <TableRow key={rule.name} className="transition-colors">
                    <TableCell className="truncate">
                      <Button
                        variant="link"
                        onClick={() => {
                          const { url, isInternal } = getRuleUrl(rule);
                          if (isInternal) {
                            window.location.href = url;
                          } else {
                            window.open(url, '_blank');
                          }
                        }}
                        className="p-0"
                      >
                        <span title={rule.name} className="font-medium">
                          {rule.name}
                        </span>
                      </Button>
                    </TableCell>
                    <TableCell>
                      <Badge>{rule.group}</Badge>
                    </TableCell>
                    <TableCell>
                      {rule.presets?.length > 0 && (
                        <div className="flex flex-wrap gap-1">
                          {rule.presets.map((p) => (
                            <PresetBadge key={p.name} preset={p.name} />
                          ))}
                        </div>
                      )}
                    </TableCell>
                    <TableCell>
                      <Badge>{rule.status}</Badge>
                    </TableCell>
                    <TableCell>
                      {rule.failing_case.length > 0 ? (
                        <div className="flex flex-col gap-1">
                          {rule.failing_case.map((caseItem, index) => (
                            <Button
                              variant="link"
                              className="cursor-pointer p-0 justify-start h-4"
                              key={caseItem.url}
                              onClick={() => {
                                window.open(
                                  `https://github.com/web-infra-dev/rslint/blob/main/${caseItem.url}`,
                                  '_blank',
                                );
                              }}
                            >
                              {index + 1}.{caseItem.name}
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
  return <RuleImplementationStatus />;
};

export default RuleApp;
