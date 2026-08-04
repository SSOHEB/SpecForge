# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Public API in `pkg/config` for library consumers to load, validate, and watch configurations.
- Hot reload config watcher with debouncing and robust directory-level monitoring.
- Optional functional/method-style config access API to generated Go code.
- Cobra CLI for `SpecForge` with subcommands: `defaults`, `validate`, `generate`, `schema`, `docs`, `init`, and `completion`.
- Markdown documentation generator for generating human-readable configuration docs.
- Configuration examples demonstrating generalizability using Redis and PostgreSQL patterns.
- Comprehensive testing pyramid including golden tests, black-box integration tests, and benchmarks.
- Release cross-compilation pipeline utilizing GoReleaser, GitHub Actions workflow, and Dockerfile packaging.
- Narrative user guide, `CONTRIBUTING.md`, package documentation (`doc.go`), and inline symbol comments.
- MIT License.

### Changed
- Standardized project name to `SpecForge`, updating module path, CLI binary, and environment prefixes.
- Replaced custom build script with a standard GoReleaser pipeline.
- Improved watcher startup sequence to register filesystem watches synchronously during initialization.

### Fixed
- Resolved race conditions in the configuration watcher preventing concurrent reloads from overwriting state.
- Fixed an issue where the hot reload watcher could miss filesystem events occurring during startup.
- Fixed GoReleaser v2 deprecated properties in the CI release workflow.
