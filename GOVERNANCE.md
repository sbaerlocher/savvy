# Project Governance

This document describes the governance model for the Savvy project.

## Project Overview

Savvy is an open-source digital card, voucher, and gift card management system. It is currently maintained by Simon Bärlocher (@sbaerlocher) and welcomes contributions from the community.

## Governance Model

Savvy follows a **Benevolent Dictator For Life (BDFL)** governance model with a path to meritocratic governance as the project grows.

### Current Phase: BDFL

**Project Lead**: Simon Bärlocher (@sbaerlocher)

As the project founder and current sole maintainer, Simon has final decision-making authority on:

- Project direction and roadmap
- Major architectural changes
- Release timing and versioning
- Acceptance or rejection of contributions
- Addition of new maintainers

### Future Phase: Meritocratic Governance

As the project grows and more contributors join, we plan to transition to a meritocratic governance model with:

- **Maintainers**: Contributors with commit access based on merit
- **Core Team**: Group of maintainers who make consensus-based decisions
- **Technical Steering Committee (TSC)**: For major technical decisions

## Roles & Responsibilities

### Project Lead (BDFL)

**Current**: Simon Bärlocher (@sbaerlocher)

**Responsibilities**:

- Sets project vision and strategic direction
- Makes final decisions on disputes or deadlocks
- Approves addition of new maintainers
- Releases new versions
- Manages project infrastructure (CI/CD, hosting, etc.)
- Ensures project health and sustainability

### Maintainers

**Current**: None (seeking!)

**Responsibilities**:

- Review and merge pull requests
- Triage issues and provide support
- Participate in technical discussions
- Help shape project roadmap
- Mentor new contributors

**Path to Becoming a Maintainer**:

1. Consistent high-quality contributions over 3+ months
2. Deep understanding of codebase and architecture
3. Active participation in code reviews
4. Demonstrated commitment to project values
5. Nomination by existing maintainer(s)
6. Approval by project lead

### Contributors

**Anyone can be a contributor!**

**Responsibilities**:

- Follow [Code of Conduct](CODE_OF_CONDUCT.md) (when created)
- Follow [Contributing Guidelines](CONTRIBUTING.md)
- Respect project maintainers' time and decisions
- Be respectful in discussions and code reviews

**How to Contribute**:

- Report bugs and suggest features
- Submit pull requests
- Improve documentation
- Help others in discussions
- Review pull requests

### Users

**Everyone who uses Savvy!**

**Responsibilities**:

- Provide feedback on features and usability
- Report bugs and issues
- Share experiences with the community
- Respect maintainers' and contributors' time

## Decision-Making Process

### Day-to-Day Decisions

- **Bug fixes**: Maintainers can merge after review
- **Documentation**: Maintainers can merge after review
- **Small features**: Maintainers can merge after discussion

### Major Decisions

Require discussion and approval from project lead:

- **Breaking changes**: API changes, architectural shifts
- **New major features**: Significant additions to the codebase
- **Deprecations**: Removing existing functionality
- **Security policies**: Changes to security practices
- **Governance changes**: Updates to this document

### Decision-Making Timeline

1. **Proposal**: Open an issue or discussion with detailed proposal
2. **Community Input**: Allow 7-14 days for feedback (depending on scope)
3. **Discussion**: Address concerns and iterate on proposal
4. **Decision**: Project lead makes final decision
5. **Implementation**: Approved proposals can be implemented

### Consensus Building

We strive for consensus when possible:

- Open discussion on GitHub Issues or Discussions
- Consider all viewpoints and concerns
- Seek compromise when disagreements arise
- Document decision rationale

### Disagreements & Appeals

If you disagree with a decision:

1. Engage respectfully in discussion
2. Provide clear reasoning and alternatives
3. Accept that final decision rests with project lead
4. If unresolved, you may fork the project (MIT License)

## Communication Channels

### Official Channels

- **GitHub Issues**: Bug reports, feature requests
- **GitHub Discussions**: Questions, ideas, community chat
- **GitHub Pull Requests**: Code contributions
- **Security Email**: security@sbaerlo.ch (private security reports)

### Response Expectations

- **Critical Security Issues**: Within 48 hours
- **Bug Reports**: Within 3-7 days
- **Feature Requests**: Within 7-14 days
- **Pull Requests**: Within 7-14 days (often faster)
- **Discussions**: Best effort, no guarantees

## Roadmap & Planning

### Public Roadmap

See [TODO.md](TODO.md) for:

- Planned features
- Known issues
- Future improvements
- Priority levels

### Feature Scope Decision (2026-06)

A maintainability audit flagged the breadth of the feature set (20+
features for a solo-maintained project — Admin System Health, Email
Template Preview, Impersonation, Batch Export, per-channel notification
preferences, …) as a candidate for trimming.

**Decision: keep the full feature set.** No features are removed. They
are built, tested, and runtime-gated via `ENABLE_*` toggles
(see [README.md](README.md#core-vs-optional-features)), so a small
deployment stays small without deleting code. Scope is managed by
configuration, not amputation. Revisit only if maintenance burden
becomes concrete (recurring security/upkeep cost on a specific
feature), not preemptively.

### Release Cycle

- **Patch releases** (1.8.x): Bug fixes, security updates (as needed)
- **Minor releases** (1.x.0): New features, improvements (monthly or as ready)
- **Major releases** (x.0.0): Breaking changes (infrequent, when necessary)

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **Major** (X.0.0): Breaking changes
- **Minor** (1.X.0): New features, backward-compatible
- **Patch** (1.0.X): Bug fixes, security updates

## Project Values

Savvy is guided by these core values:

1. **User-Centric**: Prioritize user needs and experience
2. **Quality**: Maintain high code quality and test coverage
3. **Security**: Take security seriously, act quickly on vulnerabilities
4. **Transparency**: Open development, public roadmap, documented decisions
5. **Inclusivity**: Welcome contributors of all skill levels
6. **Sustainability**: Build for long-term maintenance and growth

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. While we haven't yet formalized a Code of Conduct, we expect all participants to:

- Be respectful and considerate
- Assume good intentions
- Accept constructive criticism gracefully
- Focus on what's best for the project
- Show empathy towards others

A formal Code of Conduct based on the Contributor Covenant will be adopted as the project grows.

## Contributor Recognition

We recognize contributions in multiple ways:

- **Contributors Page**: All contributors listed on GitHub
- **CHANGELOG.md**: Significant contributions mentioned in release notes
- **CODEOWNERS**: Core contributors listed as code owners
- **README.md**: Major contributors acknowledged

## Forking & Redistribution

Savvy is licensed under the [MIT License](LICENSE), which permits:

- Use for any purpose (commercial or non-commercial)
- Modification and distribution
- Sublicensing
- Private use

If you fork Savvy:

- Please respect the MIT License terms
- Consider contributing improvements back upstream
- Feel free to rename your fork to avoid confusion

## Amendments to Governance

This governance document may be updated as the project evolves.

**Process for Changes**:

1. Propose changes via GitHub Issue or Discussion
2. Allow 14-21 days for community feedback
3. Project lead makes final decision
4. Update document with change history

**Change History**:

- 2026-02-09: Initial governance document created

## Contact

**Project Lead**: Simon Bärlocher (@sbaerlocher)
**Email**: security@sbaerlo.ch (for sensitive matters)
**GitHub**: https://github.com/sbaerlocher
**Project**: https://github.com/sbaerlocher/savvy

## Acknowledgments

This governance model is inspired by:

- Kubernetes Governance
- Node.js Project Governance
- Django Governance Model
- Apache Software Foundation

As Savvy grows, we will continue to refine our governance to best serve the project and community.

---

**Last Updated**: 2026-02-09
