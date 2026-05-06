## 1. Interface & Types

- [x] 1.1 Create `pkg/component/` package directory
- [x] 1.2 Define internal `componentImpl` interface (Connect/Disconnect/SetOptions/GetOptions) in `interface.go`
- [x] 1.3 Define public `Component` interface (Start/Stop/LoadOptions/GetMetaInfo) in `interface.go`
- [x] 1.4 Define `ComponentMetadata` struct (Name/Ready/Version/Options) in `interface.go`

## 2. Base Implementation

- [x] 2.1 Create `base.go` with `baseComponent` struct (impl/Name/Version/ready/refCount/mu fields)
- [x] 2.2 Implement `NewBaseComponent(name, version string, impl componentImpl) *baseComponent` constructor
- [x] 2.3 Implement `Start(ctx context.Context) error` with mutex-protected refCount: 0→1 calls impl.Connect, >0 only increments
- [x] 2.4 Handle connect failure: keep refCount at 0, keep ready false, return error
- [x] 2.5 Implement `Stop() error` with mutex-protected refCount: 1→0 calls impl.Disconnect, >1 only decrements
- [x] 2.6 Handle Stop when refCount is already 0: return nil without calling Disconnect

## 3. Options & Metadata

- [x] 3.1 Implement `LoadOptions(opts any) error` delegating to impl.SetOptions(opts)
- [x] 3.2 Implement `GetMetaInfo() ComponentMetadata` composing Name/Ready/Version from base + Options from impl.GetOptions()

## 4. Verification

- [x] 4.1 Verify code compiles with `go build ./pkg/component/...`
