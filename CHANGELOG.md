# Changelog

## [Unreleased]

## [0.2.0] - 2026-11-07

### Breaking

- See repository diff for API changes in this release (secure defaults, renamed packages where applicable, context-aware constructors).

### Security

- Hardened defaults: TLS, timeouts, body limits, PII redaction where logging applies.

### Fixed

- Bug fixes from the v0.2.0 hardening pass.

### Changed

- CI uses go-version-file; added race, staticcheck, govulncheck, golangci-lint (gosec).



- `Watch(ctx)` is cancellable: the listener is cancelled via `CancelListenConfig` when ctx is done.
- New `Client.Close` cancels all active listeners.
- `Download` honors ctx during retry backoff.
- Config files are always written with mode 0600 (existing files are tightened).
- Deployment env vars use the `NACOS_*` prefix only.

## [0.1.0] - 2026-11-07

### Added

- `NewClient` / `NewClientWithAPI` for config download and watch
- `LoadConfigFromEnvAndMeta`, `NamespaceFromEnv`, `ShouldSkip`, and `WriteConfigFile`
- English README, examples, and package documentation
