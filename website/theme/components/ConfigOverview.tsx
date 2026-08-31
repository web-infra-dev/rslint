// cspell:ignore languageoptionsecmaversion languageoptionssourcetype languageoptionsglobals languageoptionsparseroptionsproject languageoptionsparseroptionsprojectservice
import { Link } from '@rspress/core/theme';
import type { RslintConfigEntry } from '@rslint/core';
import {
  ArrowRightIcon,
  BracesIcon,
  FilesIcon,
  SlidersHorizontalIcon,
  type LucideIcon,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@components/ui/card';

type LanguageOptions = NonNullable<RslintConfigEntry['languageOptions']>;
type ParserOptions = NonNullable<LanguageOptions['parserOptions']>;
type TopLevelOption = Exclude<
  Extract<keyof RslintConfigEntry, string>,
  'languageOptions' | 'name'
>;
type LanguageOption = Exclude<
  Extract<keyof LanguageOptions, string>,
  'parserOptions'
>;
type ParserOption = Extract<keyof ParserOptions, string>;
type ConfigOption =
  | TopLevelOption
  | `languageOptions.${LanguageOption}`
  | `languageOptions.parserOptions.${ParserOption}`;

interface GroupItem {
  text: ConfigOption;
  link: string;
}

interface Group {
  name: string;
  link?: string;
  icon: LucideIcon;
  wide?: boolean;
  items: GroupItem[];
}

// Keep the documented, user-facing options tied to @rslint/core's config
// types while intentionally omitting metadata-only fields.
const CONFIG_OPTION_LINKS = {
  basePath: '/config/base-path',
  files: '/config/files',
  ignores: '/config/ignoring-files',
  rules: '/config/rules',
  plugins: '/config/plugins',
  settings: '/config/settings',
  'languageOptions.ecmaVersion':
    '/config/language-options#languageoptionsecmaversion',
  'languageOptions.sourceType':
    '/config/language-options#languageoptionssourcetype',
  'languageOptions.parserOptions.projectService':
    '/config/language-options#languageoptionsparseroptionsprojectservice',
  'languageOptions.parserOptions.project':
    '/config/language-options#languageoptionsparseroptionsproject',
  'languageOptions.globals': '/config/language-options#languageoptionsglobals',
} satisfies Record<ConfigOption, string>;

function option(text: ConfigOption): GroupItem {
  return { text, link: CONFIG_OPTION_LINKS[text] };
}

const OVERVIEW_GROUPS: Group[] = [
  {
    name: 'matching',
    icon: FilesIcon,
    items: [option('basePath'), option('files'), option('ignores')],
  },
  {
    name: 'linting',
    icon: SlidersHorizontalIcon,
    items: [option('rules'), option('plugins'), option('settings')],
  },
  {
    name: 'languageOptions',
    link: '/config/language-options',
    icon: BracesIcon,
    wide: true,
    items: [
      option('languageOptions.ecmaVersion'),
      option('languageOptions.sourceType'),
      option('languageOptions.parserOptions.projectService'),
      option('languageOptions.parserOptions.project'),
      option('languageOptions.globals'),
    ],
  },
];

export default function ConfigOverview() {
  return (
    <div className="mt-8 grid gap-4 md:grid-cols-2">
      {OVERVIEW_GROUPS.map((group) => {
        const Icon = group.icon;

        return (
          <Card
            key={group.name}
            className={`gap-0 overflow-hidden rounded-lg py-0 shadow-none ${group.wide ? 'md:col-span-2' : ''}`}
          >
            <CardHeader className="border-b bg-muted/30 px-5 py-4">
              <CardTitle className="flex items-center gap-2 text-base">
                <Icon
                  aria-hidden="true"
                  className="size-4 text-muted-foreground"
                />
                {group.link ? (
                  <Link
                    className="text-foreground no-underline transition-colors hover:text-primary"
                    href={group.link}
                    style={{ borderBottom: 'none' }}
                  >
                    {group.name}
                  </Link>
                ) : (
                  group.name
                )}
              </CardTitle>
            </CardHeader>
            <CardContent className="grid gap-1 p-2">
              {group.items.map((item) => (
                <Link
                  key={item.text}
                  className="group flex min-w-0 items-center justify-between gap-3 rounded-md px-3 py-2.5 text-sm font-medium text-foreground no-underline transition-colors hover:bg-muted"
                  href={item.link}
                  style={{
                    borderBottom: 'none',
                    color: 'inherit',
                    display: 'flex',
                  }}
                >
                  <code
                    className="min-w-0 break-all text-[13px]"
                    style={{
                      background: 'transparent',
                      color: 'inherit',
                      padding: 0,
                    }}
                  >
                    {item.text}
                  </code>
                  <ArrowRightIcon
                    aria-hidden="true"
                    className="size-4 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5"
                  />
                </Link>
              ))}
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
