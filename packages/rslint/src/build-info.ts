import { generatedBuildInfo } from './build-info.generated.js';

export interface BuildInfo {
  /** The rslint release version, or "unreleased" for a development build. */
  readonly version: string;
  readonly typescript: {
    /** Full commit SHA of the embedded TypeScript compiler. */
    readonly commit: string;
    /** An upstream release at that exact commit, or null. */
    readonly releaseVersion: string | null;
  };
}

/** Compiler binding shipped with this installation; no Git or network access. */
export const buildInfo: BuildInfo = Object.freeze({
  ...generatedBuildInfo,
  typescript: Object.freeze({ ...generatedBuildInfo.typescript }),
});
