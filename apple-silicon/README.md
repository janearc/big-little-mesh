# apple-silicon

The host-local on-device capability provider for the fleet: one Swift process per
host that fronts Apple's on-device frameworks behind a single arbiter (one bouncer
per chip), registered as a model backend and reached over local HTTP. A pipeline
never links Swift — it reaches this through the good-citizen model client and the
`model.v1` / `sidecar.v1` descriptors.

## State (incremental)

- [x] **Core ported + builds.** The bounded single-arbiter plus the transcription
  (audio→text) and synthesis (text→audio, +Personal Voice) capabilities, behind one
  uniform request/result. Exercised by a CLI (`provider transcribe …`, `provider
  say …`).
- [ ] **Serving layer.** Replace the CLI entrypoint with an HTTP server speaking
  `model.v1` operations as protojson, plus `/health`.
- [ ] **Discovery.** Announce a `sidecar.v1.SidecarDescriptor` and register with
  delightd; emit the observability heartbeat.
- [ ] **Text capability.** Add the on-device foundation-model (text→text) capability
  to the same arbiter.

## Build

```
swift build          # builds the provider executable
```

Requires macOS 26 (the on-device frameworks) on Apple Silicon.
