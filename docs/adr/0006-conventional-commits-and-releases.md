# 6. Automate releases from Conventional Commits

- Status: accepted
- Date: 2026-08-06 _(recorded retrospectively 2026-08-08)_

Commit messages and PR titles follow
[Conventional Commits](https://www.conventionalcommits.org/). release-please
derives the version and `CHANGELOG.md` from commit types on `master`; GoReleaser
builds and publishes artifacts. `fix:`/`perf:` → patch, `feat:` → minor,
`feat!:`/`BREAKING CHANGE:` → major; `ci|build|docs|refactor|test|chore` → none.
Renovate emits `fix(deps):`/`ci(deps):`.

No manual release steps, but a mistyped type mis-classifies a release, and
squash-merges make the PR title load-bearing.
