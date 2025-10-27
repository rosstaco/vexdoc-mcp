# Plan: GitHub Workflows with Automated Semantic Versioning

Set up automated Go binary builds and GitHub releases using GoReleaser triggered by semantic version tags, including pre-release support. GoReleaser will handle multi-platform builds, checksums, and release creation automatically.

## Steps

1. **Create `.goreleaser.yml` config** in project root with multi-platform builds (linux/darwin/windows for amd64/arm64), version injection via ldflags to replace hardcoded `ServerVersion` in [`internal/mcp/types.go`](internal/mcp/types.go), automatic pre-release detection (`prerelease: auto`), and GitHub release configuration with checksums.

2. **Add `.github/workflows/release.yml`** triggered on `v*.*.*` tags (including pre-release tags like `v0.2.0-beta.1`) that runs tests via existing [`justfile`](justfile) `ci` recipe, then executes GoReleaser to build binaries and create GitHub release with artifacts.

3. **Update version handling** in [`internal/mcp/types.go`](internal/mcp/types.go) to use build-time injection instead of hardcoded `"0.1.0"` constant, allowing version to be set via `-ldflags` during build.

4. **Add CI workflow** `.github/workflows/ci.yml` for PRs and main branch that runs linting, tests, and coverage checks using existing [`justfile`](justfile) recipes without releasing.

5. **Document release process** in [`README.md`](README.md) explaining how to create stable releases (`git tag v0.2.0`) and pre-releases (`git tag v0.2.0-beta.1`), what artifacts are generated, and which GitHub release badge appears.
