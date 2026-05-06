## MODIFIED Requirements

### Requirement: Runtime SHALL be a convention singleton
Runtime SHALL be exposed as a package-level variable `var RT = NewRuntime()` in the `pkg/runner` package (not `pkg/runner/runtime`). No enforcement of uniqueness (no sync.Once).

#### Scenario: Package-level RT is usable directly from runner package
- **WHEN** user imports `pkg/runner`
- **THEN** `runner.RT` SHALL be a pre-initialized Runtime instance

#### Scenario: Same-package access without package qualifier
- **WHEN** code within `pkg/runner` accesses RT
- **THEN** it SHALL use `RT` directly without package qualifier
