# apple-silicon

The host-local, on-device capability provider for the fleet: one process per host
that fronts Apple's on-device frameworks behind a single arbiter (one front door to
the one Neural Engine per host), reached over the mesh. A pipeline never links Swift —
it reaches this through the good-citizen model client and the `model.v1` /
`sidecar.v1` descriptors.

**Library-first.** The `Capabilities` library is the core; the daemon and any shell
wrapper *call* it. There is no standalone CLI — this is a machine-first component.

## State (incremental)

- [x] **`Capabilities` library** — the single-arbiter (one bouncer per chip) +
  transcription/synthesis, with a `Router` that dispatches by role and sheds BUSY.
- [x] **Contracts wired** — `model.v1` `Invoke` messages + a Swift `Codable` emitter
  (`gen/swift`); the `ProviderService` adapter maps `InvokeRequest` ⟷ the capability core.
- [x] **HTTP transport** — a `Network.framework` daemon (`provider`) serving `POST
  /invoke` (model.v1 protojson) + `GET /health` over loopback; the thin executable wires
  the real capabilities behind the arbiter. Env: `PROVIDER_PORT` (8077),
  `PROVIDER_CAPACITY` (2).
- [ ] **Discovery** — announce a `sidecar.v1.SidecarDescriptor`; delightd's poll picks it up.
- [ ] **Text capability** — add the on-device foundation model (text→text) to the arbiter.

## Build

The Swift package is rooted at the **repo root** (one package, mirroring the single
`go.mod`):

```
swift build      # from the blm repo root
swift test
```

## Requirements

- **Apple Silicon with a Neural Engine, on macOS 26.** The recency bar is high on purpose
  — AFM just landed and it's a moving target — so it builds against the macOS 26 SDK
  (`macOS("26.0")` in `Package.swift`) with a Swift 6.2.x toolchain, and it runs on the
  ANE. Host-only, accordingly.
- **Model assets aren't bundled.** The transcriber assets download on first use (via
  `AssetInventory`), and a **Personal Voice** has to be enrolled in System Settings to
  exist at all. Expect a download or two.

## Toolchain, formatting & CI

- **Toolchain:** Swift 6.2.x (the Xcode that ships the macOS 26 SDK).
- **Formatting** is `swift format` against `.swift-format` (4-space, 100-col) — the
  `gofmt` / `ruff` of this corner:
  ```
  swift format lint --strict --recursive --configuration .swift-format apple-silicon/Sources apple-silicon/Tests
  swift format --in-place --recursive --configuration .swift-format apple-silicon/Sources apple-silicon/Tests   # fix
  ```
- **CI** (`.github/workflows/swift.yml`): format-lint + `swift build` + `swift test`, on a
  **self-hosted macOS runner** — it needs the ANE and the macOS 26 SDK, so the Linux and
  hosted runners can't. Label it `self-hosted, macOS, ARM64`; it's path-filtered to
  `apple-silicon/**`, so it stays out of the Go/Python lanes.
