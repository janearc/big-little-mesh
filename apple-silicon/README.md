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

- **Apple Silicon with a Neural Engine (ANE), on macOS 26.** A deliberately high bar on
  two axes: **recency** — it builds against the macOS 26 SDK (deployment target
  `macOS("26.0")` in the repo-root `Package.swift`) via a Swift **6.2.x** toolchain, so an
  older macOS / SDK cannot build it — and **hardware** — it runs on the ANE, which not
  every machine has. This is host-only by nature, not portable.
- **Model assets are not automatically present.** The speech-transcriber assets
  download on first use (via `AssetInventory`), and a **Personal Voice** must be
  enrolled by the user in System Settings to be available at all. Having a Mac is not
  enough — expect a couple of downloads / a setup step.

## Toolchain, formatting & CI

- **Toolchain:** Swift **6.2.x** from the Xcode (beta) that ships the macOS 26 SDK.
- **Formatting / lint** is `swift format`, configured by `.swift-format` at the repo root
  (4-space indent, 100-col) — the Swift equivalent of the Go `gofmt` / Python `ruff`
  gates. Locally:
  ```
  # check
  swift format lint --strict --recursive --configuration .swift-format apple-silicon/Sources apple-silicon/Tests
  # fix in place
  swift format --in-place --recursive --configuration .swift-format apple-silicon/Sources apple-silicon/Tests
  ```
- **CI** lives in `.github/workflows/swift.yml`: format-lint + `swift build` + `swift test`.
  It **cannot run on the Linux runners** the Go/Python lanes use, nor on hosted macOS
  runners (they don't meet the recency bar and carry no ANE) — so it runs on a
  **self-hosted macOS runner**,
  realistically the dev machine. Register one labelled `self-hosted, macOS, ARM64`
  (repo *Settings → Actions → Runners → New self-hosted runner*, macOS/arm64). The lane is
  **path-filtered** to `apple-silicon/**`, so it only fires on Swift changes and never
  blocks the Go/Python lanes; when the runner is offline the Swift checks simply wait.
