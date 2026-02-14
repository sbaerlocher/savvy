# Pull Request

## Description

<!-- Provide a brief description of the changes in this PR -->

## Type of Change

<!-- Mark the relevant option(s) with an "x" -->

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Refactoring (no functional changes, no API changes)
- [ ] Performance improvement
- [ ] Test improvement
- [ ] Build/CI improvement
- [ ] Other (please describe):

## Changes Made

<!-- List the specific changes you made -->

- Change 1
- Change 2
- Change 3

## Motivation and Context

<!-- Why is this change needed? What problem does it solve? -->
<!-- If it fixes an open issue, please link to the issue here -->

Closes #(issue number)
Fixes #(issue number)
Related to #(issue number)

## Testing

<!-- Describe how you tested your changes -->

### Test Coverage

- [ ] Unit tests pass locally (`go test ./...` or `npm test`)
- [ ] E2E tests pass locally (`npm run test:e2e`)
- [ ] Manual testing completed
- [ ] Code coverage maintained or improved
- [ ] New tests added for new functionality

### Test Details

<!-- Describe your test scenarios -->

**Backend Tests (Go):**

```bash
go test ./...
# or
make test
```

**Frontend Tests (Vitest):**

```bash
cd client
npm test
```

**E2E Tests (Playwright):**

```bash
cd client
npm run test:e2e
```

### Manual Testing Checklist

- [ ] Tested in Chrome/Chromium
- [ ] Tested in Firefox
- [ ] Tested on mobile (iOS/Android)
- [ ] Tested offline functionality (if applicable)
- [ ] Tested with different user roles/permissions
- [ ] Tested error cases

## Documentation

- [ ] Code is self-documenting (clear naming, comments where needed)
- [ ] README.md updated (if applicable)
- [ ] ARCHITECTURE.md updated (if applicable)
- [ ] CHANGELOG.md updated (if applicable)
- [ ] AGENTS.md updated (if applicable)
- [ ] API documentation updated (if applicable)
- [ ] Migration guide added (if breaking change)

## Code Quality Checklist

- [ ] My code follows the project's code style guidelines
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings or errors
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published
- [ ] I have checked for potential security issues

## Go Code Quality (if applicable)

- [ ] `gofmt` applied (code is formatted)
- [ ] `golangci-lint run` passes without errors
- [ ] No new `database.DB` calls in handlers (Clean Architecture)
- [ ] Services use interfaces (for testability)
- [ ] Errors are properly wrapped with context
- [ ] Structured logging used (`slog`)

## Frontend Code Quality (if applicable)

- [ ] TypeScript types added/updated
- [ ] ESLint passes without errors
- [ ] Prettier formatting applied
- [ ] No `any` types used (unless absolutely necessary)
- [ ] Svelte components follow naming conventions
- [ ] API client properly handles errors
- [ ] i18n translations added (DE, EN, FR)

## Database Changes (if applicable)

- [ ] Migration added to `internal/migrations/migrations.go`
- [ ] Migration tested locally
- [ ] Rollback tested (if applicable)
- [ ] Migration documented in PR description
- [ ] Indexes added for query performance (if needed)
- [ ] Foreign keys properly defined

## Breaking Changes

<!-- If this PR introduces breaking changes, describe them here -->

**Breaking Changes:**

- [ ] None
- [ ] API response format changed
- [ ] Database schema changed (requires migration)
- [ ] Configuration changed (environment variables)
- [ ] Frontend component API changed
- [ ] Other:

**Migration Guide:**

<!-- If there are breaking changes, provide a migration guide -->

```
Steps to upgrade:
1. ...
2. ...
3. ...
```

## Screenshots (if applicable)

<!-- Add screenshots or GIFs showing the change -->

**Before:**

<!-- Screenshot of the old behavior -->

**After:**

<!-- Screenshot of the new behavior -->

## Performance Impact

<!-- Describe any performance implications -->

- [ ] No performance impact
- [ ] Performance improved
- [ ] Performance potentially degraded (explain below)
- [ ] Not applicable

**Details:**

<!-- If there's a performance impact, describe it here -->

## Security Considerations

<!-- Describe any security implications -->

- [ ] No security implications
- [ ] Security improved (describe below)
- [ ] Potential security risk (describe below and how it's mitigated)

**Details:**

<!-- Describe security considerations if applicable -->

## Deployment Notes

<!-- Any special deployment considerations? -->

- [ ] No special deployment steps needed
- [ ] Requires environment variable changes (document below)
- [ ] Requires database migration (automatic on startup)
- [ ] Requires Docker image rebuild
- [ ] Requires frontend rebuild (`npm run build:embed`)
- [ ] Other:

**Deployment Steps:**

<!-- List any special deployment steps -->

## Rollback Plan

<!-- How can this change be rolled back if needed? -->

- [ ] Standard rollback (revert to previous version)
- [ ] Database rollback needed (describe below)
- [ ] Manual intervention needed (describe below)

## Additional Notes

<!-- Add any additional notes or context here -->

---

## Checklist for Reviewers

**Code Review Focus:**

- [ ] Clean Architecture principles followed
- [ ] Error handling is appropriate
- [ ] Security considerations addressed
- [ ] Performance implications acceptable
- [ ] Tests are comprehensive
- [ ] Documentation is adequate

**Testing Review:**

- [ ] Manual testing performed
- [ ] All CI checks pass
- [ ] No regressions introduced

---

## Related Links

<!-- Add links to related documentation, issues, or discussions -->

- Documentation:
- Related Issue: #
- Related PR: #
- Discussion: #

---

**Thank you for contributing to Savvy! 🎉**
