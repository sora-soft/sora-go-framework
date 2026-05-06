## ADDED Requirements

### Requirement: Campaign for leadership
Election SHALL provide a `Campaign(ctx, id)` method that blocks until the caller with the given ID becomes the leader. If the election has no current leader, the first candidate SHALL become leader immediately. If a leader already exists, the caller SHALL block until it can acquire leadership.

#### Scenario: First candidate becomes leader
- **WHEN** `Campaign(ctx, "node-1")` is called on an election with no leader
- **THEN** "node-1" SHALL become the leader and `Campaign` SHALL return without error

#### Scenario: Candidate blocks while leader exists
- **WHEN** `Campaign(ctx, "node-2")` is called while "node-1" is the leader
- **THEN** the call SHALL block until "node-1" resigns or its lease expires

### Requirement: Resign from leadership
Election SHALL provide a `Resign(ctx)` method that releases the current leader's hold, allowing blocked candidates to proceed.

#### Scenario: Resign allows next candidate to become leader
- **WHEN** the current leader calls `Resign(ctx)` and another candidate is blocked on `Campaign`
- **THEN** the resigned leader SHALL no longer be leader and the blocked candidate SHALL become the new leader

### Requirement: Query current leader
Election SHALL provide a `Leader(ctx)` method that returns the ID of the current leader, or an empty string if no leader exists.

#### Scenario: Leader exists
- **WHEN** `Leader(ctx)` is called after a successful campaign
- **THEN** it SHALL return the ID of the current leader

#### Scenario: No leader exists
- **WHEN** `Leader(ctx)` is called on a fresh election with no campaigns
- **THEN** it SHALL return an empty string

### Requirement: Watch leader changes
Election SHALL provide a `Watch(ctx)` method that returns `<-chan string`. The channel SHALL receive the current leader ID (or empty string) immediately, and SHALL receive updates whenever the leader changes. The channel SHALL close on context cancellation or fatal error.

#### Scenario: Watch receives initial leader state
- **WHEN** `Watch(ctx)` is called and "node-1" is the leader
- **THEN** the channel SHALL immediately yield "node-1"

#### Scenario: Watch receives leader change
- **WHEN** a watcher is active and the leader changes from "node-1" to "node-2"
- **THEN** the channel SHALL yield "node-2"

#### Scenario: Watch channel closes on context cancellation
- **WHEN** the context passed to `Watch(ctx)` is cancelled
- **THEN** the channel SHALL be closed
