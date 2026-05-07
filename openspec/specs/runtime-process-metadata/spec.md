### Requirement: Runtime SHALL capture process start time
Runtime SHALL record `time.Now()` at construction time and expose it via `StartTime() time.Time`.

#### Scenario: StartTime reflects construction moment
- **WHEN** `NewRuntime()` is called at time T
- **THEN** `rt.StartTime()` SHALL return T

### Requirement: Runtime SHALL capture process working directory
Runtime SHALL record `os.Getwd()` at construction time and expose it via `Root() string`.

#### Scenario: Root reflects cwd at construction
- **WHEN** `NewRuntime()` is called while cwd is `/app`
- **THEN** `rt.Root()` SHALL return `/app`

### Requirement: Runtime SHALL be a convention singleton
Runtime SHALL be exposed as a package-level variable `var RT = NewRuntime()` in the `pkg/runner` package (not `pkg/runner/runtime`). No enforcement of uniqueness (no sync.Once).

#### Scenario: Package-level RT is usable directly from runner package
- **WHEN** user imports `pkg/runner`
- **THEN** `runner.RT` SHALL be a pre-initialized Runtime instance

#### Scenario: Same-package access without package qualifier
- **WHEN** code within `pkg/runner` accesses RT
- **THEN** it SHALL use `RT` directly without package qualifier
