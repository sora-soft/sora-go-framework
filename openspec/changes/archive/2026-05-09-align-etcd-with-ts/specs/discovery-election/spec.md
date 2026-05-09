## MODIFIED Requirements

### Requirement: Campaign for leadership
Election SHALL provide a `Campaign(ctx, id)` method that blocks until the caller with the given ID becomes the leader. Campaign SHALL use `go.etcd.io/etcd/client/v3/concurrency.Election` with a session backed by an independent lease (not shared with the EtcdComponent's main lease). If a leader already exists, the caller SHALL block until it can acquire leadership. The election key SHALL be `<prefix>/singleton/<name>`.

#### Scenario: First candidate becomes leader
- **WHEN** `Campaign(ctx, "node-1")` is called on an election with no leader
- **THEN** "node-1" SHALL become the leader and `Campaign` SHALL return without error

#### Scenario: Candidate blocks while leader exists
- **WHEN** `Campaign(ctx, "node-2")` is called while "node-1" is the leader
- **THEN** the call SHALL block until "node-1" resigns or its session expires

### Requirement: Resign from leadership
Election SHALL provide a `Resign(ctx)` method that releases the current leader's hold via `concurrency.Election.Resign()`, allowing blocked candidates to proceed.

#### Scenario: Resign allows next candidate
- **WHEN** the current leader calls `Resign(ctx)` and another candidate is blocked on `Campaign`
- **THEN** the resigned leader SHALL no longer be leader and the blocked candidate SHALL become the new leader

### Requirement: Query current leader
Election SHALL provide a `Leader(ctx)` method that returns the ID of the current leader by querying `concurrency.Election.Leader()`. Returns an empty string if no leader exists.

#### Scenario: Leader exists
- **WHEN** `Leader(ctx)` is called after a successful campaign with id "node-1"
- **THEN** it SHALL return "node-1"

#### Scenario: No leader exists
- **WHEN** `Leader(ctx)` is called on a fresh election with no campaigns
- **THEN** it SHALL return an empty string

### Requirement: Watch leader changes
Election SHALL provide a `Watch(ctx)` method that returns `<-chan string`. The channel SHALL receive the current leader ID (or empty string) immediately, then periodically poll for leader changes. The channel SHALL close on context cancellation.

#### Scenario: Watch receives initial leader state
- **WHEN** `Watch(ctx)` is called and "node-1" is the leader
- **THEN** the channel SHALL immediately yield "node-1"

#### Scenario: Watch channel closes on context cancellation
- **WHEN** the context passed to `Watch(ctx)` is cancelled
- **THEN** the channel SHALL be closed
