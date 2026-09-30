package expiring_todo_comments

import (
	"os"
	"strings"
)

// isPullRequest mirrors ci-info 4.4.0's ordered vendor checks. Keep this
// environment policy local to the only rule that consumes it.
// https://github.com/watson/ci-info/blob/v4.4.0/vendors.json
func isPullRequest() bool {
	if os.Getenv("CI") == "false" {
		return false
	}
	has := func(key string) bool { return os.Getenv(key) != "" }
	notFalse := func(key string) bool { value, exists := os.LookupEnv(key); return exists && value != "false" }
	vendors := []struct{ matches, pullRequest bool }{
		{has("AGOLA_GIT_REF"), has("AGOLA_PULL_REQUEST_ID")},   // Agola CI
		{has("ALPIC_HOST"), false},                             // Alpic
		{has("AC_APPCIRCLE"), notFalse("AC_GIT_PR")},           // Appcircle
		{has("APPVEYOR"), has("APPVEYOR_PULL_REQUEST_NUMBER")}, // AppVeyor
		{has("CODEBUILD_BUILD_ARN"), os.Getenv("CODEBUILD_WEBHOOK_EVENT") == "PULL_REQUEST_CREATED" || os.Getenv("CODEBUILD_WEBHOOK_EVENT") == "PULL_REQUEST_UPDATED" || os.Getenv("CODEBUILD_WEBHOOK_EVENT") == "PULL_REQUEST_REOPENED"}, // AWS CodeBuild
		{has("TF_BUILD"), os.Getenv("BUILD_REASON") == "PullRequest"},                    // Azure Pipelines
		{has("bamboo_planKey"), false},                                                   // Bamboo
		{has("BITBUCKET_COMMIT"), has("BITBUCKET_PR_ID")},                                // Bitbucket Pipelines
		{has("BITRISE_IO"), has("BITRISE_PULL_REQUEST")},                                 // Bitrise
		{has("BUDDY_WORKSPACE_ID"), has("BUDDY_EXECUTION_PULL_REQUEST_ID")},              // Buddy
		{has("BUILDKITE"), notFalse("BUILDKITE_PULL_REQUEST")},                           // Buildkite
		{has("CIRCLECI"), has("CIRCLE_PULL_REQUEST")},                                    // CircleCI
		{has("CIRRUS_CI"), has("CIRRUS_PR")},                                             // Cirrus CI
		{has("CF_PAGES"), false},                                                         // Cloudflare Pages
		{has("WORKERS_CI"), false},                                                       // Cloudflare Workers
		{has("CF_BUILD_ID"), has("CF_PULL_REQUEST_NUMBER") || has("CF_PULL_REQUEST_ID")}, // Codefresh
		{has("CM_BUILD_ID"), has("CM_PULL_REQUEST")},                                     // Codemagic
		{os.Getenv("CI_NAME") == "codeship", false},                                      // Codeship
		{has("DRONE"), os.Getenv("DRONE_BUILD_EVENT") == "pull_request"},                 // Drone
		{has("DSARI"), false},                                                            // dsari
		{has("EARTHLY_CI"), false},                                                       // Earthly
		{has("EAS_BUILD"), false},                                                        // Expo Application Services
		{has("GERRIT_PROJECT"), false},                                                   // Gerrit
		{has("GITEA_ACTIONS"), false},                                                    // Gitea Actions
		{has("GITHUB_ACTIONS"), os.Getenv("GITHUB_EVENT_NAME") == "pull_request"},        // GitHub Actions
		{has("GITLAB_CI"), has("CI_MERGE_REQUEST_ID")},                                   // GitLab CI
		{has("GO_PIPELINE_LABEL"), false},                                                // GoCD
		{has("BUILDER_OUTPUT"), false},                                                   // Google Cloud Build
		{has("HARNESS_BUILD_ID"), false},                                                 // Harness CI
		{strings.Contains(os.Getenv("NODE"), "/app/.heroku/node/bin/node"), false},       // Heroku
		{has("HUDSON_URL"), false},                                                       // Hudson
		{has("JENKINS_URL") && has("BUILD_ID"), has("ghprbPullId") || has("CHANGE_ID")},  // Jenkins
		{has("LAYERCI"), has("LAYERCI_PULL_REQUEST")},                                    // LayerCI
		{has("MAGNUM"), false},                                                           // Magnum CI
		{has("NETLIFY"), notFalse("PULL_REQUEST")},                                       // Netlify CI
		{has("NEVERCODE"), notFalse("NEVERCODE_PULL_REQUEST")},                           // Nevercode
		{has("PROW_JOB_ID"), false},                                                      // Prow
		{has("RELEASE_BUILD_ID"), false},                                                 // ReleaseHub
		{has("RENDER"), os.Getenv("IS_PULL_REQUEST") == "true"},                          // Render
		{has("SAILCI"), has("SAIL_PULL_REQUEST_NUMBER")},                                 // Sail CI
		{has("SCREWDRIVER"), notFalse("SD_PULL_REQUEST")},                                // Screwdriver
		{has("SEMAPHORE"), has("PULL_REQUEST_NUMBER")},                                   // Semaphore
		{os.Getenv("CI_NAME") == "sourcehut", false},                                     // Sourcehut
		{has("STRIDER"), false},                                                          // Strider CD
		{has("TASK_ID") && has("RUN_ID"), false},                                         // TaskCluster
		{has("TEAMCITY_VERSION"), false},                                                 // TeamCity
		{has("TRAVIS"), notFalse("TRAVIS_PULL_REQUEST")},                                 // Travis CI
		{has("VELA"), os.Getenv("VELA_PULL_REQUEST") == "1"},                             // Vela
		{has("NOW_BUILDER") || has("VERCEL"), has("VERCEL_GIT_PULL_REQUEST_ID")},         // Vercel
		{has("APPCENTER_BUILD_ID"), false},                                               // Visual Studio App Center
		{os.Getenv("CI") == "woodpecker", os.Getenv("CI_BUILD_EVENT") == "pull_request"}, // Woodpecker
		{has("CI_XCODE_PROJECT"), has("CI_PULL_REQUEST_NUMBER")},                         // Xcode Cloud
		{has("XCS"), false},                                                              // Xcode Server
	}
	result := false
	for _, vendor := range vendors {
		if vendor.matches {
			result = vendor.pullRequest
		}
	}
	return result
}
