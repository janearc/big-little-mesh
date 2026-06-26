// swift-tools-version: 6.0
//
// apple-silicon -- the host-local on-device capability provider for the fleet.
// A single Swift process per host that fronts Apple's on-device frameworks
// (transcription / synthesis / text) behind one arbiter, registered as a model
// backend and reached over local HTTP. This package is the promotion of the
// research proving sketch into blm; the serving + descriptor layers wire onto the
// model.v1 / sidecar.v1 contracts in following steps.
import PackageDescription

let package = Package(
    name: "apple-silicon",
    platforms: [
        .macOS("26.0"),
    ],
    targets: [
        // The proven core today: the bounded single-arbiter (the "one bouncer per
        // chip") plus the transcription and synthesis capabilities, behind one
        // uniform request/result shape. A CLI entrypoint exercises it; the HTTP
        // serving layer replaces that entrypoint in the next step.
        .executableTarget(
            name: "provider",
            path: "Sources/provider"
        ),
    ]
)
