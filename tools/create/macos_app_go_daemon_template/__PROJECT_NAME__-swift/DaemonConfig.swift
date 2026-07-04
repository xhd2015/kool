import Darwin
import Foundation

enum DaemonConfig {
    static let productionPort = __DEFAULT_PORT__
    static let productionStateSubpath = "__STATE_SUBPATH__"
    static let debugStateSubpath = "__STATE_SUBPATH__-debug"

    struct Snapshot: Equatable {
        let port: Int
        let stateDir: String
    }

    private static let lock = NSLock()
    private static var cached: Snapshot?

    static var resolved: Snapshot {
        lock.lock()
        defer { lock.unlock() }
        if let cached {
            return cached
        }
        let snapshot = computeSnapshot()
        cached = snapshot
        return snapshot
    }

    static var resolvedPort: Int { resolved.port }
    static var resolvedStateDir: String { resolved.stateDir }

    static func resetForTesting() {
        lock.lock()
        cached = nil
        lock.unlock()
    }

    private static func computeSnapshot() -> Snapshot {
        let env = ProcessInfo.processInfo.environment
        let home = env["HOME"] ?? NSHomeDirectory()

        let stateDir: String
        if let override = env["DAEMON_STATE_DIR"], !override.isEmpty {
            stateDir = override
        } else if DaemonDebug.isEnabled {
            stateDir = (home as NSString).appendingPathComponent(debugStateSubpath)
        } else {
            stateDir = (home as NSString).appendingPathComponent(productionStateSubpath)
        }

        let port: Int
        if let envPort = env["DAEMON_PORT"], let parsed = Int(envPort) {
            port = parsed
        } else if DaemonDebug.isEnabled {
            port = pickEphemeralPort()
        } else {
            port = productionPort
        }

        return Snapshot(port: port, stateDir: stateDir)
    }

    static func pickEphemeralPort() -> Int {
        let fd = socket(AF_INET, SOCK_STREAM, 0)
        guard fd >= 0 else {
            return productionPort + 1
        }
        defer { close(fd) }

        var addr = sockaddr_in()
        addr.sin_len = UInt8(MemoryLayout<sockaddr_in>.size)
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_port = 0
        addr.sin_addr.s_addr = inet_addr("127.0.0.1")

        let bindResult = withUnsafePointer(to: &addr) {
            $0.withMemoryRebound(to: sockaddr.self, capacity: 1) {
                bind(fd, $0, socklen_t(MemoryLayout<sockaddr_in>.size))
            }
        }
        guard bindResult == 0 else {
            return productionPort + 1
        }

        var bound = sockaddr_in()
        var boundLen = socklen_t(MemoryLayout<sockaddr_in>.size)
        let nameResult = withUnsafeMutablePointer(to: &bound) {
            $0.withMemoryRebound(to: sockaddr.self, capacity: 1) {
                getsockname(fd, $0, &boundLen)
            }
        }
        guard nameResult == 0 else {
            return productionPort + 1
        }
        return Int(UInt16(bigEndian: bound.sin_port))
    }

    static func terminateStaleDaemonIfNeeded() {
        guard DaemonDebug.isEnabled else { return }
        let env = ProcessInfo.processInfo.environment
        if env["DAEMON_PORT"] != nil {
            return
        }

        let stateDir = resolvedStateDir
        let pidPath = DaemonShutdown.pidFilePath(stateDir: stateDir)
        let contents = try? String(contentsOfFile: pidPath, encoding: .utf8)
        let target = DaemonShutdown.quitTarget(
            spawnedPID: nil,
            spawnedRunning: false,
            pidFileContents: contents
        )
        DaemonShutdown.terminate(target: target)
    }
}