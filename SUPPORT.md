# Support

Thank you for using Savvy! This document provides information on how to get help and support.

## 📚 Getting Started

Before seeking support, please check these resources:

### Documentation

- **[README.md](README.md)** - Quick start guide, features overview, installation
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Technical architecture and system design
- **[OPERATIONS.md](OPERATIONS.md)** - Deployment, monitoring, and operations
- **[AGENTS.md](AGENTS.md)** - AI agent instructions and technical details
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - How to contribute to the project
- **[DEVELOPMENT.md](DEVELOPMENT.md)** - Development setup and best practices

### Common Questions

**Q: How do I install Savvy?**
A: See [README.md - Installation](README.md#-quick-start) for Docker Compose setup.

**Q: How do I enable OAuth/OIDC authentication?**
A: See [OPERATIONS.md - Environment Variables](OPERATIONS.md) for OAuth configuration.

**Q: How do I add a new language?**
A: See `client/src/lib/i18n/locales/` for existing translations. Copy `en.ts` to create a new language file.

**Q: How do I deploy to production?**
A: See [README.md - Deployment](README.md#-deployment) for Docker and Kubernetes instructions.

**Q: How do I run tests?**
A: See [CONTRIBUTING.md - Testing](CONTRIBUTING.md#testing-your-changes) for test commands.

**Q: What versions are supported?**
A: See [SECURITY.md - Supported Versions](SECURITY.md#supported-versions) for support policy.

## 💬 Community Support (Free)

Community support is provided on a best-effort basis by maintainers and community members.

### GitHub Discussions (Recommended)

**Best for**: Questions, ideas, general discussion

https://github.com/sbaerlocher/savvy/discussions

- **Q&A**: Ask questions and get help from the community
- **Ideas**: Share feature ideas and suggestions
- **Show and Tell**: Share your Savvy setup or use case
- **General**: General discussion about Savvy

**Response Time**: Best effort, typically 1-7 days

### GitHub Issues

**Best for**: Bug reports, feature requests

https://github.com/sbaerlocher/savvy/issues

- **Bug Reports**: Use the [Bug Report template](.github/ISSUE_TEMPLATE/bug_report.md)
- **Feature Requests**: Use the [Feature Request template](.github/ISSUE_TEMPLATE/feature_request.md)

**Response Time**:

- Critical bugs: 1-3 days
- Other bugs: 3-7 days
- Feature requests: 7-14 days

**Please Note**:

- Search existing issues before creating a new one
- Provide as much detail as possible
- Include logs, screenshots, and steps to reproduce
- Be patient - maintainers are volunteers

### Where NOT to Report Issues

❌ **DO NOT use GitHub Issues for**:

- General questions (use Discussions instead)
- Security vulnerabilities (use security@sbaerlo.ch instead)
- Support requests (use Discussions instead)

## 🚨 Security Issues

**NEVER report security vulnerabilities publicly!**

For security issues, please email: **security@sbaerlo.ch**

See [SECURITY.md](SECURITY.md) for our security policy and responsible disclosure process.

## 🔍 Troubleshooting

### Common Issues

#### Issue: "Port 3000 already in use"

**Solution**:

```bash
# Find and kill process using port 3000
make clean-port

# Or manually
lsof -ti:3000 | xargs kill -9
```

#### Issue: "Database connection refused"

**Solution**:

```bash
# Check if PostgreSQL is running
docker compose ps

# View PostgreSQL logs
docker compose logs postgres

# Restart services
docker compose restart
```

#### Issue: "Frontend not loading / 404 errors"

**Solution**:

```bash
# Rebuild frontend
cd client && npm run build:embed

# Restart backend
docker compose restart api
```

#### Issue: "Barcode scanner not working"

**Solution**:

- Ensure you're using HTTPS (required for camera access)
- Check browser permissions for camera access
- Try a different browser (Chrome/Edge recommended)
- Ensure good lighting and steady hand

#### Issue: "Service Worker not updating"

**Solution**:

```bash
# Clear browser cache and service workers
# Chrome: DevTools → Application → Service Workers → Unregister
# Firefox: about:serviceworkers → Unregister

# Or use incognito/private mode for testing
```

### Getting Logs

#### Docker Logs

```bash
# All services
docker compose logs -f

# Backend only
docker compose logs -f api

# Frontend only
docker compose logs -f client

# Database only
docker compose logs -f postgres

# Last 100 lines
docker compose logs --tail=100
```

#### Browser Console Logs

1. Open Developer Tools (F12)
2. Go to Console tab
3. Copy any errors or warnings
4. Include in your bug report

### Debug Mode

```bash
# Enable debug logging
export LOG_LEVEL=debug
docker compose up -d
```

## 📖 Additional Resources

### External Resources

- **Go Documentation**: https://go.dev/doc/
- **Echo Framework**: https://echo.labstack.com/
- **GORM**: https://gorm.io/
- **SvelteKit**: https://kit.svelte.dev/
- **PostgreSQL**: https://www.postgresql.org/docs/
- **Docker**: https://docs.docker.com/
- **Playwright**: https://playwright.dev/

### Related Projects

- **Templ**: https://github.com/a-h/templ (used in earlier versions)
- **html5-qrcode**: https://github.com/mebjas/html5-qrcode
- **bwip-js**: https://github.com/metafloor/bwip-js

## 🎓 Learning Resources

### For Contributors

- [CONTRIBUTING.md](CONTRIBUTING.md) - How to contribute
- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) - Architecture pattern we follow

### For Developers

- [Go by Example](https://gobyexample.com/)
- [SvelteKit Tutorial](https://learn.svelte.dev/)
- [PostgreSQL Tutorial](https://www.postgresqltutorial.com/)

## 🤝 Contributing to Support

You can help support Savvy users by:

- **Answering questions** in GitHub Discussions
- **Reviewing bug reports** and confirming issues
- **Improving documentation** with clarifications and examples
- **Creating tutorials** and blog posts
- **Sharing your experience** with the community

See [CONTRIBUTING.md](CONTRIBUTING.md) for more ways to contribute.

## 📊 Project Status

- **Latest Version**: v1.0.0 (see [CHANGELOG.md](CHANGELOG.md))
- **Test Coverage**: 71.6% (Services), 83.9% (Handlers), 90.9% (Models)
- **Security Audit**: Low Risk (0 Critical, 0 High findings)
- **Maintenance Status**: Actively maintained

## 📧 Contact

### General Support

- **GitHub Discussions**: https://github.com/sbaerlocher/savvy/discussions
- **GitHub Issues**: https://github.com/sbaerlocher/savvy/issues

### Security Issues

- **Email**: security@sbaerlo.ch
- **Response Time**: Within 48 hours

### Project Lead

- **Name**: Simon Bärlocher
- **GitHub**: [@sbaerlocher](https://github.com/sbaerlocher)
- **Email**: security@sbaerlo.ch (for security or sensitive matters)

## ⏰ Response Time Expectations

| Channel            | Topic                    | Expected Response Time |
| ------------------ | ------------------------ | ---------------------- |
| GitHub Issues      | Critical bugs            | 1-3 business days      |
| GitHub Issues      | Regular bugs             | 3-7 business days      |
| GitHub Issues      | Feature requests         | 7-14 business days     |
| GitHub Discussions | Questions                | 1-7 days (best effort) |
| Security Email     | Security vulnerabilities | Within 48 hours        |
| Pull Requests      | Code review              | 7-14 days              |

**Please note**: These are target response times, not guarantees. Maintainers are volunteers with other commitments.

## ⚠️ Important Notes

- **No Phone Support**: We do not provide phone support
- **No Live Chat**: We do not have live chat support
- **Volunteer Project**: Savvy is maintained by volunteers in their free time
- **Best Effort**: Community support is provided on a best-effort basis
- **Be Patient**: Please be patient and respectful of maintainers' time
- **Search First**: Always search existing issues/discussions before posting

## 🙏 Thank You!

Thank you for using Savvy and being part of our community!

Your feedback helps us improve the project for everyone. 🎉

---

**Last Updated**: 2026-02-09
