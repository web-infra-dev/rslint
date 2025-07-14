# GitHub Actions Workflows

This directory contains GitHub Actions workflows for providing AI feedback and continuous integration.

## Workflows

### 1. CI Workflow (`ci.yml`)

**Triggers:**
- Push to `main` branch
- Pull requests to `main` branch

**Jobs:**
- **build-and-test**: Builds the project, runs tests, and performs code quality checks
- **benchmark**: Runs performance benchmarks (only on pushes to main)

**Features:**
- Automatic submodule initialization and patch application
- Go build and test execution
- Code formatting validation
- Go vet analysis
- Performance testing with sample TypeScript files
- Artifact uploads for reports

### 2. AI Feedback Workflow (`ai-feedback.yml`)

**Triggers:**
- Push to `main` branch
- Pull requests to `main` branch
- Issue comments
- Manual dispatch

**Features:**
- **Structured JSON Output**: Provides machine-readable feedback in JSON format
- **Human-readable Reports**: Generates markdown reports for easy reading
- **Comprehensive Analysis**: Includes build status, test results, code quality, and performance metrics
- **Linter Analysis**: Runs the linter on sample code to validate functionality
- **PR Comments**: Automatically comments on pull requests with feedback
- **Artifact Storage**: Saves all reports and logs for future reference

## AI Feedback Structure

The AI feedback workflow generates structured data in JSON format that includes:

```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "event": "pull_request",
  "repository": "web-infra-dev/rslint",
  "commit": "abc123...",
  "branch": "feature-branch",
  "build_status": "success",
  "tests": {
    "total": 42,
    "passed": 42,
    "failed": 0,
    "status": "passed"
  },
  "code_quality": {
    "formatting": "passed",
    "vet": "passed",
    "dependencies": "up_to_date"
  },
  "performance": {
    "binary_size": "15M",
    "rules_count": 40
  },
  "linter_results": {
    "files_processed": 1,
    "issues_found": 4,
    "rules_triggered": ["no-floating-promises", "no-array-delete"]
  },
  "metadata": {
    "go_version": "go version go1.24.1 linux/amd64",
    "os": "Linux",
    "arch": "x86_64"
  }
}
```

## Usage

The workflows run automatically on relevant events. You can also manually trigger the AI feedback workflow:

1. Go to the Actions tab in GitHub
2. Select "AI Feedback" workflow
3. Click "Run workflow"
4. Choose the branch and click "Run workflow"

## Artifacts

Both workflows upload artifacts that can be downloaded:

- **ai-feedback-report**: Contains JSON and markdown reports
- **benchmark-results**: Contains performance benchmark results
- **Test outputs**: Raw test results and linter outputs

## Benefits for AI Development

1. **Structured Data**: Machine-readable JSON format for easy parsing
2. **Consistent Format**: Standardized structure across all runs
3. **Comprehensive Coverage**: Includes all aspects of code quality and performance
4. **Historical Data**: Artifacts provide historical tracking
5. **Automated Feedback**: Reduces manual review overhead
6. **Integration Ready**: Easy to integrate with AI tools and dashboards