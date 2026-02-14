# Release Process

**Status**: Fully automated via GitHub Actions  
**Last Updated**: 2026-02-09

---

## 🚀 Creating a Release

Releases are **automatically created** when you push a version tag to GitHub.

### 1. Update CHANGELOG.md

```bash
# Edit CHANGELOG.md with new version changes
vim CHANGELOG.md
git add CHANGELOG.md
git commit -m "docs: update CHANGELOG for v1.0.0"
git push
```

### 2. Create and Push Tag

```bash
# Semantic Versioning: major.minor.patch
# - major: Breaking Changes (v2.0.0)
# - minor: New Features
# - patch: Bug Fixes (v1.8.2)

# Create tag
git tag -a v1.0.0 -m "Release v1.0.0"

# Push tag (triggers GitHub Actions)
git push origin v1.0.0
```

### 3. GitHub Actions (Automatic)

Once the tag is pushed, GitHub Actions automatically:

1. ✅ Builds SvelteKit frontend
2. ✅ Builds Go binaries (Linux amd64, arm64)
3. ✅ Creates Docker images (multi-platform)
4. ✅ Pushes to GitHub Container Registry (ghcr.io)
5. ✅ Creates GitHub Release with:
   - Binaries
   - Docker image links
   - Auto-generated changelog
   - Checksums

**Monitor progress**: https://github.com/sbaerlocher/savvy/actions

---

## 📋 Release Checklist

### Before Release

- [ ] All tests pass (`go test ./...`, `npm run test:e2e`)
- [ ] CHANGELOG.md updated
- [ ] Breaking changes documented (if any)
- [ ] Migration guide written (if needed)

### After Release

- [ ] Verify GitHub Release created
- [ ] Verify Docker images available at ghcr.io
- [ ] Test production deployment
- [ ] Monitor for errors

---

## 🔄 Version Schema

**Semantic Versioning 2.0.0**

```
v1.0.0
│ │ │
│ │ └─ Patch: Bug Fixes, Security Patches
│ └─── Minor: New Features (backwards-compatible)
└───── Major: Breaking Changes
```

### Examples

- `v1.8.1` → `v1.8.2`: Bug Fix (Patch)
- `v1.8.2` → `v1.0.0`: New Feature (Minor)
- `v1.0.0` → `v2.0.0`: Breaking Change (Major)

### Pre-Releases

```bash
# Alpha
git tag v2.0.0-alpha.1
git push origin v2.0.0-alpha.1

# Beta
git tag v2.0.0-beta.1
git push origin v2.0.0-beta.1

# Release Candidate
git tag v2.0.0-rc.1
git push origin v2.0.0-rc.1
```

Pre-releases are automatically marked by GoReleaser.

---

## 🐳 Docker Tags

GitHub Actions automatically creates these Docker tags:

```
ghcr.io/sbaerlocher/container/savvy:v1.0.0    # Exact version
ghcr.io/sbaerlocher/container/savvy:1.9       # Minor version
ghcr.io/sbaerlocher/container/savvy:1         # Major version
ghcr.io/sbaerlocher/container/savvy:latest    # Latest stable
```

---

## 🚫 What NOT to Do

- ❌ **Don't build releases locally** - Always use GitHub Actions
- ❌ **Don't run GoReleaser locally** - It's configured for CI only
- ❌ **Don't manually upload binaries** - GitHub Actions handles it
- ❌ **Don't forget to update CHANGELOG.md** - Required before release

---

## 📚 Resources

- **[.github/workflows/release.yml](../.github/workflows/release.yml)** - Release workflow
- **[GoReleaser Docs](https://goreleaser.com/)** - Release automation
- **[Semantic Versioning](https://semver.org/)** - Versioning spec
- **[Conventional Commits](https://www.conventionalcommits.org/)** - Commit format

---

**Need help?** Check [SUPPORT.md](../SUPPORT.md) or [CONTRIBUTING.md](../CONTRIBUTING.md)
