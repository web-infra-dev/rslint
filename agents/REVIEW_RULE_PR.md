# Rslint Rule PR Review Agent Guide

## 1. Objective

The primary goal of this agent workflow is to **review Rslint rule implementations** within a Pull Request (PR). You act as an expert Software Engineer and QA Specialist to:

- Identify **security vulnerabilities**, **logic bugs**, and **deviations** from the intended behavior (e.g., ESLint parity).
- Detect **missing test cases** and coverage gaps.
- Verify the implementation locally.

## 2. Prerequisites (MANDATORY)

**You CANNOT proceed without a Pull Request (PR) URL.**

- **Condition**: The user **MUST** provide the PR URL (e.g., `https://github.com/web-infra-dev/rslint/pull/999999`).
- **Action**: If the PR URL is missing, **TERMINATE THE TASK IMMEDIATELY** and request the URL from the user. Do not attempt to guess or review the current branch without explicit instruction.

## 3. Workflow

Once the PR URL is provided, follow this strictly defined process:

### Phase 1: Fetch & Analyze

1.  **Read PR Context**:
    - Extract the branch name and PR number.
    - Use GitHub CLI to fetch the changes:
      ```bash
      gh pr diff <PR_URL>
      ```
2.  **Code Review (Static Analysis)**:
    Analyze the `diff` for specific issues:
    - **Vulnerabilities**: Look for unsafe pointer dereferences (nil panics), infinite loops, or unchecked type assertions.
    - **Logic Parity**: Compare the Go implementation against the original ESLint rule (if applicable). Does it handle all node types?
    - **Edge Cases**:
      - Are comments handled correctly?
      - Does it handle empty bodies or optional chains?
      - Are TypeScript-specific nodes (like `AsExpression`) considered?

### Phase 2: Direct Feedback

If you identify bugs, vulnerabilities, or missing logic:

1.  **Draft Feedback**:
    - Compile all identified issues into a clear list.
    - Present this list to the user for review.

2.  **User Confirmation (MANDATORY)**:
    - **Ask**: "Do you want to post these comments to the PR?"
    - **Wait**: You MUST wait for the user's explicit confirmation.
    - **If Rejected**: Terminate the task or proceed without commenting if instructed.
    - **If Confirmed**: Proceed to step 3.

3.  **Post Comments**: Use the GitHub CLI to send feedback directly to the PR.
    ```bash
    gh pr comment <PR_URL> --body "## Review Feedback\n\nFound a potential issue:\n- **Location**: [File/Line]\n- **Issue**: [Description]\n- **Suggestion**: [Fix]"
    ```
    _Note: Be constructive and specific._

### Phase 3: Local Verification

#### 1. Checkout & Initialize

```bash
gh pr checkout <PR_URL>
go mod download
cd packages/rslint-test-tools && pnpm install
```

#### 2. Verify Go Implementation

Run the Go unit tests for the specific rule package:

```bash
go test -v -count=1 ./internal/rules/... # Target specific rule package
```

If you suspected a bug in Phase 1, create a **new test case** in the Go `_test.go` file to reproduce it.

#### 3. Verify JS/TS Test Cases

Check for the existence and validity of integration tests in `packages/rslint-test-tools`.

**A. Check Existence**

- **Action**: Look for the test file `packages/rslint-test-tools/tests/.../<rule-name>.test.ts`.
- **If Missing**:
  - **Draft Comment**: Prepare a comment about the missing test case.
  - **User Confirmation**: Show the comment content to the user and ask for confirmation.
  - **If Confirmed**:
    - **STOP** and comment on the PR:
      ```bash
      gh pr comment <PR_URL> --body "## Missing Test Case\n\nPlease add the corresponding JS/TS test case in \`packages/rslint-test-tools\`."
      ```

**B. Run Tests**

- **Action**: Build the binary and run the JS tests.
  ```bash
  cd packages/rslint && pnpm run build:bin
  cd ../rslint-test-tools
  pnpm test <rule-name>
  ```

**C. Handle Failures**

- **If Tests Fail**:
  1.  **Analyze**: Investigate the failure. Is it a logic bug in the Go code or an incorrect test case?
  2.  **Attempt Fix**: Try to fix the code or the test locally.
  3.  **If Fixed**:
      - **Draft Comment**: Prepare the solution explanation.
      - **User Confirmation**: Ask the user before posting.
      - **If Confirmed**: Post the solution (diff or explanation) to the PR.
        ```bash
        gh pr comment <PR_URL> --body "## Test Failure Fixed\n\nThe JS tests were failing. I found the issue and verified this fix:\n\n\`\`\`go\n[Insert Fix Code]\n\`\`\`"
        ```
  4.  **If Unfixable**:
      - **Draft Comment**: Prepare the failure report.
      - **User Confirmation**: Ask the user before posting.
      - **If Confirmed**: Report the failure details.
        ```bash
        gh pr comment <PR_URL> --body "## Test Failure Report\n\nThe JS tests are failing and I could not resolve it automatically. Please investigate.\n\nError Log:\n\`\`\`\n[Insert Log]\n\`\`\`"
        ```

## 4. Final Report

After completing the workflow, summarize your actions to the user:

- **Review Status**: Did you find issues?
- **Feedback Link**: Link to the comment you posted (if any).
- **Verification Result**: Did local tests pass? Did you add new coverage?
