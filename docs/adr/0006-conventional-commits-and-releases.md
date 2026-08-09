# 6. Automate releases from Conventional Commits

- Status: accepted
- Date: 2026-08-06 _(recorded retrospectively 2026-08-08)_

Commit messages and PR titles follow
[Conventional Commits](https://www.conventionalcommits.org/). release-please
derives the version and `CHANGELOG.md` from commit types on `master`; GoReleaser
builds and publishes artifacts. `fix:`/`perf:` → patch, `feat:` → minor,
`feat!:`/`BREAKING CHANGE:` → major; `ci|build|docs|refactor|test|chore` → none.
Renovate follows the same convention, typed by whether an update can reach
users: `fix(deps):` for Go modules compiled into the binary, `ci(deps):` for
GitHub Actions, and `chore(deps):` for everything that cannot (test-only
modules, the local `Dockerfile`'s builder image, the pinned toolchain).

No manual release steps, but a mistyped type mis-classifies a release, and
squash-merges make the PR title load-bearing.
