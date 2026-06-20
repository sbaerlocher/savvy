# Contributing to Savvy

First off, thank you for considering contributing to Savvy! It's people like you that make Savvy such a great tool for managing digital cards, vouchers, and gift cards.

## 🌟 How Can I Contribute?

There are many ways to contribute to Savvy:

- **Report bugs** - Found a bug? Let us know!
- **Suggest features** - Have an idea? We'd love to hear it!
- **Improve documentation** - Help others understand Savvy better
- **Write code** - Fix bugs or implement new features
- **Write tests** - Help us improve code quality and coverage

## 📋 Prerequisites

Before you begin, ensure you have the following installed:

- **Docker & Docker Compose** (required for development)
- **Go 1.26+** (for local Go development)
- **Node.js 18+ LTS** (Node.js 20+ recommended for frontend development)
- **Make** (optional, for convenience commands)
- **Git** (for version control)

## 🚀 Getting Started

### 1. Fork & Clone

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR_USERNAME/savvy.git
cd savvy

# Add upstream remote
git remote add upstream https://github.com/sbaerlocher/savvy.git
```

### 2. Set Up Development Environment

```bash
# Copy environment configuration
cp .env.example .env
# Edit .env with your local settings

# Start all services with Docker Compose
make dev

# Seed test data (optional, in another terminal)
make seed
```

**Access the application**:

- App (Frontend + API, routed by dde traefik): <https://savvy.test>

### 3. Verify Setup

```bash
# Check services are running
docker compose ps

# View logs
make logs

# Run all tests
go test ./...
```

## 💻 Development Workflow

### Branch Strategy

We use **GitHub Flow** with the following conventions:

- `main` - Production-ready code, protected branch
- `feature/<description>` - New features
- `fix/<description>` - Bug fixes
- `docs/<description>` - Documentation updates
- `refactor/<description>` - Code refactoring
- `test/<description>` - Test improvements

**Examples**:

```bash
git checkout -b feature/add-qr-export
git checkout -b fix/barcode-scanner-mobile
git checkout -b docs/update-api-guide
```

### Code Quality Standards

#### Comment Guidelines

We follow **clean code principles** for comments:

**✅ DO**:

- Keep important security notes (e.g., "SECURITY (SVL-003): Only store isAuthenticated flag")
- Document complex algorithms or non-obvious logic
- Explain "why" when the reason isn't clear from the code
- Use English for all comments

**❌ DON'T**:

- Add obvious comments (e.g., "Create child logger", "Initialize event listeners")
- Use JSDoc for self-explanatory functions
- Add section headers in small files
- Mix German and English comments

**Examples**:

```typescript
// ❌ BAD - Obvious comment
const pwaLogger = logger.child("PWA"); // Create child logger for PWA

// ✅ GOOD - No comment needed, code is self-explanatory
const pwaLogger = logger.child("PWA");

// ✅ GOOD - Important security context
/**
 * SECURITY (SVL-003): Only store isAuthenticated flag, not user data
 * - Prevents XSS attacks from stealing user data via localStorage
 * - User data must be fetched from /api/v1/auth/me (validates session)
 */
function loadAuthFromStorage(): AuthState { ... }
```

### Making Changes

#### Frontend (SvelteKit)

**Architecture**: SvelteKit SPA with TypeScript

- **Stores**: Svelte stores for global state (auth, offline, notifications, i18n)
- **API Clients**: Modular API clients in `client/src/lib/api/`
- **Components**: Reusable Svelte components in `client/src/lib/components/`
- **Routes**: SvelteKit file-based routing in `client/src/routes/`

**Code Style**:

- Use TypeScript for type safety
- Follow clean code principles (minimal comments)
- Use Svelte reactive statements (`$:`) for computed values
- Prefer composition over inheritance

#### Backend (Go)

```bash
# ✅ RECOMMENDED: Start with Docker Compose (includes hot reload)
docker compose up

# Run all tests
go test ./...  # Runs ALL tests (services, handlers, models)

# Run tests with coverage
make test-coverage

# Lint code
make lint

# Format code
make fmt
```

**Key Directories**:

- `internal/handlers/` - HTTP request handlers (Controllers)
  - `api/` - JSON API handlers (cards, vouchers, gift_cards, sessions, batch, etc.)
  - `shares/` - Share handler abstraction (adapter pattern)
- `internal/services/` - Business logic layer
- `internal/repository/` - Data access layer
- `internal/models/` - GORM database models
- `internal/middleware/` - Echo middleware (auth, CORS, CSRF, pgstore, session_keys, etc.)

**Layered Architecture Rules**:

- ✅ Handlers call Services (never direct database access)
- ✅ Services call Repositories (business logic here)
- ✅ Repositories use GORM models (data access only)
- ✅ All services have interfaces for testability

#### Frontend (SvelteKit + TypeScript)

```bash
# ⚠️ ALWAYS use Docker Compose for frontend development
# Local `npm run dev` does NOT work (Vite proxy requires Docker network)
docker compose up

# Run type checking
npm run check

# Run unit tests (Vitest)
npm test

# Run E2E tests (Playwright)
npm run test:e2e

# Build for production
npm run build
```

**Key Directories**:

- `client/src/routes/` - SvelteKit pages and routes
  - `cards/`, `vouchers/`, `gift-cards/` - Resource CRUD pages
  - `merchants/` - Merchant overview and detail pages
  - `profile/`, `security/`, `notifications/` - User settings pages
  - `admin/` - Admin panel (users, merchants, audit-log, email-templates, system-health)
- `client/src/lib/components/` - Reusable Svelte components
  - `settings/` - Settings sub-components (ProfileSection, SecuritySection, etc.)
- `client/src/lib/api/` - API client modules
- `client/src/lib/stores/` - Svelte stores (state management)
- `client/src/lib/i18n/` - Internationalization (de, en, fr)

#### Database Migrations

**IMPORTANT**: We use embedded Go-based migrations (Gormigrate), NOT separate SQL files.

```bash
# Migrations are automatically applied on server startup (AUTO_MIGRATE=true)

# To manually run migrations
make migrate-up

# Check migration status
make migrate-status

# Rollback last migration (use with caution!)
make migrate-down
```

**Creating a new migration**:

1. Add your migration code to `internal/migrations/migrations.go`
2. Use Gormigrate's `Migrate()` function
3. Test locally before committing

### Testing Your Changes

#### Unit Tests (Go)

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/services -run TestCardService_CreateCard

# Run tests with race detection
go test -race ./...
```

**Current Coverage**:

- Services: 80.0%
- Handlers: 80.4%
- Models: 100.0%
- Repositories: 97.2%

#### Unit Tests (Frontend - Vitest)

```bash
cd client

# Run unit tests
npm test

# Run with UI
npm run test:ui

# Run with coverage
npm run test:coverage
```

#### E2E Tests (Playwright)

```bash
cd client

# Install browsers (first time only)
npm run playwright:install

# Run all E2E tests (all browsers in parallel)
npm run test:e2e

# Run specific test file across all browsers (recommended)
npm run test:e2e:run -- tests/e2e/auth.spec.ts

# List tests without running (instant, no Docker startup)
npm run test:e2e:run -- tests/e2e/auth.spec.ts --list

# Run with browser-specific options (faster for quick testing)
npm run test:e2e:run -- tests/e2e/cards.spec.ts --project=chromium --headed

# Debug mode (step-through debugging)
npm run test:e2e:run -- tests/e2e/sharing.spec.ts --debug

# Run in UI mode (interactive debugging)
npm run test:e2e:ui
```

**E2E Test Locations**:

- `client/tests/e2e/` - All Playwright E2E test files (23 test files)
- Test files: `auth.spec.ts`, `cards.spec.ts`, `vouchers.spec.ts`, `gift-cards.spec.ts`,
  `profile.spec.ts`, `security.spec.ts`, etc.

#### Integration Tests

```bash
# Start test environment
docker compose -f docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./...
```

### Code Style & Conventions

#### Go Code Style

We follow standard Go conventions with additional rules:

- **Formatting**: Use `gofmt` (or `goimports`)
- **Linting**: Pass `golangci-lint run` without errors
- **Naming**:
  - Packages: lowercase, single word (e.g., `handlers`, `services`)
  - Interfaces: End with `Interface` (e.g., `CardServiceInterface`)
  - Constructors: Start with `New` (e.g., `NewCardService()`)
- **Error Handling**:
  - Always check errors, never ignore
  - Use `fmt.Errorf` with `%w` for wrapping
  - Return errors up to handler layer
- **Logging**:
  - Use `slog` for structured logging
  - Levels: Debug, Info, Warn, Error
  - Include context (userID, resourceID, etc.)

**Example**:

```go
// ✅ Good
func (s *CardService) CreateCard(ctx context.Context, card *models.Card) error {
    if err := s.cardRepo.Create(ctx, card); err != nil {
        return fmt.Errorf("failed to create card: %w", err)
    }
    return nil
}

// ❌ Bad
func (s *CardService) CreateCard(ctx context.Context, card *models.Card) error {
    s.cardRepo.Create(ctx, card) // Ignoring error!
    return nil
}
```

#### TypeScript/Svelte Code Style

- **Formatting**: Prettier (configured in `.prettierrc`)
- **Linting**: ESLint (configured in `.eslintrc`)
- **Naming**:
  - Components: PascalCase (e.g., `BarcodeScanner.svelte`)
  - Functions: camelCase (e.g., `fetchCards()`)
  - Constants: UPPER_SNAKE_CASE (e.g., `API_BASE_URL`)
- **Types**:
  - Use TypeScript interfaces for all API responses
  - Define types in `client/src/lib/types/`
  - Avoid `any` type
- **Stores**:
  - Use Svelte stores for shared state
  - Store files: lowercase with hyphens (e.g., `auth-store.ts`)

**Example**:

```typescript
// ✅ Good
export async function fetchCards(): Promise<Card[]> {
  const response = await fetch("/api/v1/cards");
  if (!response.ok) {
    throw new Error(`Failed to fetch cards: ${response.statusText}`);
  }
  return await response.json();
}

// ❌ Bad
export async function fetchCards() {
  // Missing return type
  const response = await fetch("/api/v1/cards");
  return await response.json(); // No error handling
}
```

### Commit Message Conventions

We use **Conventional Commits** for clear and semantic commit messages:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style changes (formatting, no logic change)
- `refactor`: Code refactoring
- `perf`: Performance improvement
- `test`: Adding or updating tests
- `chore`: Maintenance tasks (dependencies, build, etc.)
- `ci`: CI/CD changes

**Scopes** (optional):

- `api`, `frontend`, `backend`, `db`, `docker`, `ci`, `docs`, `test`

**Examples**:

```bash
feat(api): add QR code export endpoint
fix(frontend): resolve barcode scanner mobile issue
docs(readme): update installation instructions
refactor(services): extract common validation logic
test(handlers): add missing card handler tests
chore(deps): update Go dependencies to v1.23
```

**Breaking Changes**:

```bash
feat(api)!: change cards API response format

BREAKING CHANGE: The cards API now returns `merchant_name` instead of `merchant`
```

### Pull Request Process

1. **Update your fork**:

   ```bash
   git fetch upstream
   git checkout main
   git merge upstream/main
   git push origin main
   ```

2. **Create a feature branch**:

   ```bash
   git checkout -b feature/my-awesome-feature
   ```

3. **Make your changes**:
   - Write code following our style guidelines
   - Add or update tests
   - Update documentation if needed
   - Ensure all tests pass

4. **Commit your changes**:

   ```bash
   git add .
   git commit -m "feat: add my awesome feature"
   ```

5. **Push to your fork**:

   ```bash
   git push origin feature/my-awesome-feature
   ```

6. **Open a Pull Request**:
   - Go to https://github.com/sbaerlocher/savvy
   - Click "New Pull Request"
   - Select your fork and branch
   - Fill out the PR template (see below)
   - Wait for review

### Pull Request Template

When creating a PR, please fill out the following:

```markdown
## Description

Brief description of what this PR does.

## Type of Change

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Refactoring (no functional changes)
- [ ] Performance improvement
- [ ] Test improvement

## Changes Made

- Change 1
- Change 2
- Change 3

## Testing

- [ ] Unit tests pass (`go test ./...` and `npm test`)
- [ ] E2E tests pass (`npm run test:e2e`)
- [ ] Manual testing completed
- [ ] Code coverage maintained or improved

## Documentation

- [ ] Code is self-documenting (clear naming, comments where needed)
- [ ] README updated (if applicable)
- [ ] ARCHITECTURE.md updated (if applicable)
- [ ] CHANGELOG.md updated (if applicable)

## Checklist

- [ ] My code follows the project's code style
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published

## Screenshots (if applicable)

Add screenshots or GIFs showing the change.

## Related Issues

Closes #123
Fixes #456
```

### Code Review Process

After submitting a PR:

1. **Automated Checks** run automatically:
   - GitHub Actions CI (linting, tests, build)
   - Code coverage check
   - Security scans

2. **Code Review** by maintainers:
   - Usually within 2-3 business days
   - Reviewers may request changes or ask questions
   - Address feedback and push updates

3. **Approval & Merge**:
   - Once approved, maintainer will merge
   - PR will be automatically closed
   - Changes will appear in next release

**What We Look For**:

- ✅ Clean, readable code
- ✅ Proper error handling
- ✅ Adequate test coverage
- ✅ Documentation updates
- ✅ Follows project conventions
- ✅ No breaking changes (or clearly documented)

## 🐛 Reporting Bugs

Found a bug? Help us fix it by reporting it!

### Before Submitting a Bug Report

1. **Check existing issues**: Search [GitHub Issues](https://github.com/sbaerlocher/savvy/issues) first
2. **Verify it's reproducible**: Try to reproduce in a clean environment
3. **Check the documentation**: Make sure it's not expected behavior

### How to Submit a Bug Report

Use our **Bug Report Template**:

```markdown
**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:

1. Go to '...'
2. Click on '...'
3. Scroll down to '...'
4. See error

**Expected behavior**
What you expected to happen.

**Actual behavior**
What actually happened.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Environment:**

- OS: [e.g., macOS 14.0, Ubuntu 22.04, Windows 11]
- Browser: [e.g., Chrome 120, Firefox 115, Safari 17]
- Savvy Version: [e.g., v1.0.0]
- Docker Version: [e.g., 24.0.0]
- Go Version: [e.g., 1.23]

**Logs**
```

Paste relevant logs here (from `make logs` or `docker compose logs`)

```

**Additional context**
Add any other context about the problem here.
```

## 💡 Suggesting Features

Have an idea for a new feature? We'd love to hear it!

### Before Submitting a Feature Request

1. **Check existing issues**: Someone might have already suggested it
2. **Check the roadmap**: See [TODO.md](TODO.md) for planned features
3. **Consider scope**: Make sure it fits Savvy's purpose

### How to Submit a Feature Request

Use our **Feature Request Template**:

```markdown
**Is your feature request related to a problem?**
A clear description of what the problem is. Ex. I'm always frustrated when [...]

**Describe the solution you'd like**
A clear and concise description of what you want to happen.

**Describe alternatives you've considered**
Any alternative solutions or features you've considered.

**Use Cases**
How would this feature be used? Who would benefit from it?

**Additional context**
Add any other context, mockups, or screenshots about the feature request here.

**Willing to Contribute?**

- [ ] I am willing to submit a PR to implement this feature
```

## 📝 Improving Documentation

Documentation improvements are always welcome!

**Types of Documentation**:

- Code comments (inline documentation)
- README.md (user-facing guide)
- ARCHITECTURE.md (technical architecture)
- OPERATIONS.md (deployment & operations)
- AGENTS.md (AI agent instructions)
- API documentation (coming soon)

**How to Contribute**:

1. Follow the same PR process as code changes
2. Use clear, concise language
3. Include examples where helpful
4. Keep formatting consistent

## 🎓 Good First Issues

New to Savvy? Look for issues labeled `good first issue`:

https://github.com/sbaerlocher/savvy/labels/good%20first%20issue

These are beginner-friendly issues that are a great way to get started!

## 💬 Questions?

Need help? Have questions?

- **GitHub Discussions**: https://github.com/sbaerlocher/savvy/discussions
- **GitHub Issues**: For bug reports and feature requests only
- **Documentation**: Check [AGENTS.md](AGENTS.md) and [ARCHITECTURE.md](ARCHITECTURE.md)

## 🙏 Recognition

Contributors are recognized in:

- GitHub Contributors page
- Release notes (for significant contributions)
- CHANGELOG.md

Thank you for contributing to Savvy! 🎉

---

**License**: By contributing to Savvy, you agree that your contributions will be licensed under the [MIT License](LICENSE).
