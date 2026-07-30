import React from 'react';
import { Badge } from '@components/ui/badge';

export interface RuleVersionBadgeProps {
  introducedIn?: string | null;
}

export const RuleVersionBadge: React.FC<RuleVersionBadgeProps> = ({
  introducedIn,
}) => {
  if (!introducedIn) {
    return (
      <Badge
        variant="unreleased"
        title="This rule has not been assigned to an rslint release yet"
      >
        Unreleased
      </Badge>
    );
  }

  return (
    <a
      href={`https://github.com/web-infra-dev/rslint/releases/tag/v${introducedIn}`}
      target="_blank"
      rel="noreferrer"
      className="group"
      title={`View the rslint v${introducedIn} release`}
    >
      <Badge
        variant="version"
        className="underline-offset-2 group-hover:underline"
      >
        Added in v{introducedIn}
      </Badge>
    </a>
  );
};

export default RuleVersionBadge;
