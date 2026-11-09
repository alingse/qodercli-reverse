# GitHub Actions

Workflows live under `.github/workflows/`.

- `go-ci.yml` — formatting, module verification, tests, vet, build, race tests.
- `go-security.yml` — govulncheck and module integrity.
- `go-release.yml` — cross-platform release builds on version tags.
- `go-attestation.yml` — build provenance attestation.
- `cask-validation.yml` — Ruby Cask syntax validation.
- `deployment.yml` — deployment readiness info.

All workflows use least-privilege permissions. Secrets are never logged.