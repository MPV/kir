# Changelog

## [0.4.5](https://github.com/MPV/kir/compare/v0.4.4...v0.4.5) (2026-08-13)


### Bug Fixes

* reject image references that cannot be reported safely ([a1b7ca9](https://github.com/MPV/kir/commit/a1b7ca907925bb2ca77fa214df90b4957080e335))

## [0.4.4](https://github.com/MPV/kir/compare/v0.4.3...v0.4.4) (2026-08-10)


### Bug Fixes

* keep the images found around a document that fails ([7ceafd3](https://github.com/MPV/kir/commit/7ceafd3b10201d236ece0f1f8dd3882c7c378f5a))

## [0.4.3](https://github.com/MPV/kir/compare/v0.4.2...v0.4.3) (2026-08-10)


### Bug Fixes

* **deps:** update kubernetes monorepo to v0.36.3 ([5b7bda0](https://github.com/MPV/kir/commit/5b7bda0e2ab16810fb4ea88f6873c5cec57c3cb4))

## [0.4.2](https://github.com/MPV/kir/compare/v0.4.1...v0.4.2) (2026-08-09)


### Bug Fixes

* split YAML documents with the Kubernetes YAML reader ([d487e61](https://github.com/MPV/kir/commit/d487e61ef2f9227aa5c9a05c83a2401a6d4994e2))

## [0.4.1](https://github.com/MPV/kir/compare/v0.4.0...v0.4.1) (2026-08-09)


### Bug Fixes

* exit non-zero when a file cannot be processed ([#55](https://github.com/MPV/kir/issues/55)) ([655b764](https://github.com/MPV/kir/commit/655b764d587cccf7e3c2ef179e21141d44c007cb))

## [0.4.0](https://github.com/MPV/kir/compare/v0.3.1...v0.4.0) (2026-08-09)


### Features

* skip non-workload documents instead of aborting the stream ([3ebab6c](https://github.com/MPV/kir/commit/3ebab6c9ff5580cc3fcb878de629e35887d0fd1f))


### Bug Fixes

* process all documents from stdin, not just the first ([4e36f41](https://github.com/MPV/kir/commit/4e36f4144be68981175120b4f4cea97617803521))

## [0.3.1](https://github.com/MPV/kir/compare/v0.3.0...v0.3.1) (2026-08-08)


### Bug Fixes

* publish signed release artifacts (cosign v3 bundle format) ([42f3eb8](https://github.com/MPV/kir/commit/42f3eb8d872e69f4dc71869c159defcd1c5b62a6))

## [0.3.0](https://github.com/MPV/kir/compare/v0.2.0...v0.3.0) (2026-08-06)


### Features

* add `kir --version` and automate releases ([071299f](https://github.com/MPV/kir/commit/071299ffab74311a7a05c6e80cf3a66198d8db38))
