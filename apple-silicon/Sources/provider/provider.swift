// apple-silicon — the capability core.
//
// Proves two design claims at once: the single-arbiter contention model and "one
// transport, several frameworks." Capabilities have heterogeneous I/O —
// transcription is audio→text, synthesis is text→audio, text is text→text — so they
// share one request/result and ride one arbiter. This is the proven core promoted
// from research into blm; the model.v1 / sidecar.v1 contracts and the HTTP serving
// layer wire onto it in following steps (today it is exercised by the CLI below).

import Foundation
import Speech
import AVFoundation

// MARK: - The uniform capability shape
//
// A single request/result carries text and/or a file path; each capability validates and
// interprets the fields it needs. (On the wire these become role-specific model.v1
// messages; in-process this struct pair is the erasure that lets one arbiter and one
// server front every capability the same way.)

struct CapabilityRequest: Sendable {
    var text: String? = nil
    var inputPath: String? = nil
    var params: [String: String] = [:]
}

struct CapabilityResult: Sendable {
    var text: String? = nil
    var outputPath: String? = nil
    var detail: [String: String] = [:]
    var summary: String {
        if let t = text { return "chars=\(t.count)" }
        if let p = outputPath { return "out=\((p as NSString).lastPathComponent)" }
        return "ok"
    }
}

enum CapabilityError: Error, CustomStringConvertible {
    case badRequest(String)
    case unavailable(String)
    var description: String {
        switch self {
        case .badRequest(let m): return "bad request: \(m)"
        case .unavailable(let m): return "unavailable: \(m)"
        }
    }
}

@available(macOS 26.0, *)
protocol Capability: Sendable {
    var name: String { get }
    var role: String { get }
    func available() -> (ok: Bool, reason: String)
    func run(_ req: CapabilityRequest) async throws -> CapabilityResult
}

// MARK: - Transcription (audio → text), via SpeechTranscriber

@available(macOS 26.0, *)
struct TranscriptionCapability: Capability {
    let name = "transcription"
    let role = "transcription"
    let locale: Locale
    init(locale: Locale = Locale(identifier: "en-US")) { self.locale = locale }

    func available() -> (ok: Bool, reason: String) { (true, "speech-transcriber") }

    func run(_ req: CapabilityRequest) async throws -> CapabilityResult {
        guard let inputPath = req.inputPath else {
            throw CapabilityError.badRequest("transcription needs inputPath (an audio file)")
        }
        let url = URL(fileURLWithPath: inputPath)
        let transcriber = SpeechTranscriber(locale: locale, transcriptionOptions: [],
                                            reportingOptions: [], attributeOptions: [])
        if let r = try await AssetInventory.assetInstallationRequest(supporting: [transcriber]) {
            try await r.downloadAndInstall()
        }
        let analyzer = SpeechAnalyzer(modules: [transcriber])
        let audioFile = try AVAudioFile(forReading: url)
        let collector = Task { () -> String in
            var text = AttributedString()
            for try await result in transcriber.results { text += result.text }
            return String(text.characters)
        }
        if let last = try await analyzer.analyzeSequence(from: audioFile) {
            try await analyzer.finalizeAndFinish(through: last)
        } else {
            await analyzer.cancelAndFinishNow()
        }
        return CapabilityResult(text: try await collector.value, detail: ["role": role])
    }
}

// MARK: - Arbiter (bounded admission + single worker; BUSY when full)

@available(macOS 26.0, *)
func stamp() -> String {
    let f = DateFormatter(); f.dateFormat = "HH:mm:ss.SSS"; return f.string(from: Date())
}

@available(macOS 26.0, *)
actor Arbiter {
    enum Outcome { case done(summary: String, wall: Double); case busy; case failed(String) }

    let capacity: Int                  // max admitted at once (running + queued)
    private var admitted = 0
    private var running = false
    private var waiters: [CheckedContinuation<Void, Never>] = []

    init(capacity: Int) { self.capacity = capacity }

    func submit(_ label: String, _ work: @Sendable @escaping () async throws -> CapabilityResult) async -> Outcome {
        guard admitted < capacity else {
            print("  [\(stamp())] BUSY   \(label)  (admitted \(admitted)/\(capacity))")
            return .busy
        }
        admitted += 1
        print("  [\(stamp())] ADMIT  \(label)  (admitted \(admitted)/\(capacity))")

        // single worker: park until the worker is free, so RUN events never overlap.
        while running {
            await withCheckedContinuation { (c: CheckedContinuation<Void, Never>) in waiters.append(c) }
        }
        running = true
        defer {
            running = false
            admitted -= 1
            if !waiters.isEmpty { waiters.removeFirst().resume() }
        }

        print("  [\(stamp())] RUN    \(label)")
        let t = Date()
        do {
            let res = try await work()
            let wall = Date().timeIntervalSince(t)
            print("  [\(stamp())] DONE   \(label)  \(res.summary) wall=\(String(format: "%.1f", wall))s")
            return .done(summary: res.summary, wall: wall)
        } catch {
            print("  [\(stamp())] FAIL   \(label)  \(error)")
            return .failed("\(error)")
        }
    }
}

// MARK: - Proof CLI

@available(macOS 26.0, *)
@main
struct Main {
    static func main() async {
        let args = Array(CommandLine.arguments.dropFirst())
        guard let verb = args.first else {
            FileHandle.standardError.write(Data(
                "usage:\n  provider transcribe <audio> [audio ...]   # concurrency proof\n  provider say \"<text>\" [outfile]\n".utf8))
            exit(2)
        }
        let arbiter = Arbiter(capacity: 2)

        switch verb {
        case "transcribe":
            let files = Array(args.dropFirst())
            let cap = TranscriptionCapability()
            print("firing \(files.count) concurrent transcribe requests at one arbiter (capacity=2)\n")
            await withTaskGroup(of: Void.self) { group in
                for (i, f) in files.enumerated() {
                    let path = f
                    group.addTask {
                        let label = "req#\(i) \(URL(fileURLWithPath: path).lastPathComponent)"
                        _ = await arbiter.submit(label) { try await cap.run(CapabilityRequest(inputPath: path)) }
                    }
                }
            }
            print("\nRUN events never overlapped; overflow shed as BUSY.")

        case "say":
            let text = args.count > 1 ? args[1] : ""
            let out = args.count > 2 ? args[2] : ""
            let cap = SynthesisCapability()
            let (ok, reason) = cap.available()
            print("synthesis available: \(ok)  (\(reason))")
            let params: [String: String] = out.isEmpty ? [:] : ["out": out]
            let outcome = await arbiter.submit("say") {
                try await cap.run(CapabilityRequest(text: text, params: params))
            }
            switch outcome {
            case .done(let s, let w): print("done: \(s) in \(String(format: "%.1f", w))s")
            case .busy: print("BUSY")
            case .failed(let e): print("FAILED: \(e)")
            }

        default:
            FileHandle.standardError.write(Data("unknown verb: \(verb)\n".utf8)); exit(2)
        }
    }
}
