## 1. Struct Refactor

- [x] 1.1 Replace `startupCtx`/`startupCancel` fields with `ctx`/`cancel` in `baseWorker` struct
- [x] 1.2 Add `sync.WaitGroup` and `atomic.Bool` (running) fields to `baseWorker`
- [x] 1.3 Remove `Executor` field from `baseWorker`

## 2. Constructor Update

- [x] 2.1 Update `NewWorker()` to no longer create `Executor`; initialize `running` as false

## 3. Start Method

- [x] 3.1 Change `Start()` signature to `Start(ctx context.Context) error`
- [x] 3.2 Create lifecycle context: `b.ctx, b.cancel = context.WithCancel(ctx)`
- [x] 3.3 Set `running = true` before goroutines
- [x] 3.4 Remove defer cancel and nil cleanup (ctx survives beyond Startup)
- [x] 3.5 Pass `b.ctx` to `Runner.Startup(b.ctx)` instead of `startupCtx`

## 4. Go Method

- [x] 4.1 Rewrite `Go()` to use `running` guard + `wg.Go()` with `b.ctx` instead of `Executor.Go()`

## 5. Stop Method

- [x] 5.1 Rewrite `Stop()`: set `running = false` → `cancel()` → `wg.Wait()` → `SetState(Stopping)` → `Runner.Shutdown()` → `SetState(Stopped)`
- [x] 5.2 Remove old `startupCancel`/`startupCtx` cleanup logic and `Executor.Stop()` call

## 6. Cleanup

- [x] 6.1 Update `service.go` if needed (ServiceOptions, BaseService embeddings)
- [x] 6.2 Verify `utility.Executor` is no longer imported from `runner` package
- [x] 6.3 Ensure code compiles without errors
