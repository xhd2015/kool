import Darwin
import Foundation

struct DaemonShutdownConfig: Equatable {
    let stateDirEnvKey: String
    let defaultRelativeStateDir: String
}

enum DaemonShutdown {
    static var appDaemon: DaemonShutdownConfig {
        DaemonShutdownConfig(
            stateDirEnvKey: "DAEMON_STATE_DIR",
            defaultRelativeStateDir: DaemonDebug.isEnabled
                ? DaemonConfig.debugStateSubpath
                : DaemonConfig.productionStateSubpath
        )
    }

    enum Target: Equatable {
        case none
        case spawned(pid: Int32)
        case pidFile(pid: Int32)
    }

    static func resolveStateDir(
        config: DaemonShutdownConfig,
        env: [String: String] = ProcessInfo.processInfo.environment,
        home: String = NSHomeDirectory()
    ) -> String {
        if let stateDir = env[config.stateDirEnvKey], !stateDir.isEmpty {
            return stateDir
        }
        return (home as NSString).appendingPathComponent(config.defaultRelativeStateDir)
    }

    static func pidFilePath(stateDir: String) -> String {
        (stateDir as NSString).appendingPathComponent("daemon.pid")
    }

    static func parsePID(_ contents: String) -> Int32? {
        let trimmed = contents.trimmingCharacters(in: .whitespacesAndNewlines)
        guard let pid = Int32(trimmed), pid > 0 else {
            return nil
        }
        return pid
    }

    static func quitTarget(
        spawnedPID: Int32?,
        spawnedRunning: Bool,
        pidFileContents: String?
    ) -> Target {
        if let spawnedPID, spawnedRunning {
            return .spawned(pid: spawnedPID)
        }
        if let pidFileContents, let pid = parsePID(pidFileContents) {
            return .pidFile(pid: pid)
        }
        return .none
    }

    static func terminate(target: Target) {
        switch target {
        case .none:
            return
        case .spawned(let pid), .pidFile(let pid):
            signalPID(pid, sig: SIGTERM)
            usleep(100_000)
            if kill(pid, 0) == 0 {
                signalPID(pid, sig: SIGKILL)
            }
        }
    }

    static func terminateSpawnedProcess(_ process: Process) {
        guard process.isRunning else { return }
        process.terminate()
        let deadline = Date().addingTimeInterval(3)
        while process.isRunning && Date() < deadline {
            usleep(50_000)
        }
        if process.isRunning {
            kill(process.processIdentifier, SIGKILL)
        }
    }

    static func terminateOnQuit(
        config: DaemonShutdownConfig,
        spawnedProcess: Process?,
        env: [String: String] = ProcessInfo.processInfo.environment,
        home: String = NSHomeDirectory()
    ) {
        if let spawnedProcess, spawnedProcess.isRunning {
            terminateSpawnedProcess(spawnedProcess)
            return
        }

        let stateDir = resolveStateDir(config: config, env: env, home: home)
        let pidPath = pidFilePath(stateDir: stateDir)
        let contents = try? String(contentsOfFile: pidPath, encoding: .utf8)
        let target = quitTarget(
            spawnedPID: nil,
            spawnedRunning: false,
            pidFileContents: contents
        )
        terminate(target: target)
    }

    private static func signalPID(_ pid: Int32, sig: Int32) {
        guard pid > 0 else { return }
        kill(pid, sig)
    }
}