import ServiceManagement
import SwiftUI

@MainActor
final class DaemonReadiness: ObservableObject {
    static let shared = DaemonReadiness()

    @Published private(set) var isReady = false

    private init() {}

    func markReady() {
        isReady = true
    }
}

class AppDelegate: NSObject, NSApplicationDelegate {
    private var daemonProcess: Process?

    func daemonInfo() -> (port: Int, pid: Int32) {
        let port = DaemonConfig.resolvedPort
        let pid = daemonProcess.map { $0.processIdentifier } ?? 0
        let running = daemonProcess?.isRunning ?? false
        if running && pid > 0 {
            return (port, pid)
        }
        return (port, -1)
    }

    func restartDaemon() {
        Task { @MainActor in
            if let proc = daemonProcess, proc.isRunning {
                proc.terminate()
                proc.waitUntilExit()
            }
            daemonProcess = nil
            await ensureDaemonRunning()
        }
    }

    func applicationWillTerminate(_ notification: Notification) {
        DaemonShutdown.terminateOnQuit(
            config: DaemonShutdown.appDaemon,
            spawnedProcess: daemonProcess
        )
        daemonProcess = nil
    }

    func applicationDidFinishLaunching(_ notification: Notification) {
        Task { @MainActor in
            await ensureDaemonRunning()
        }
    }

    @MainActor
    private func ensureDaemonRunning() async {
        defer { DaemonReadiness.shared.markReady() }
        if (try? await DaemonClient.shared.health()) == true {
            return
        }
        DaemonConfig.terminateStaleDaemonIfNeeded()
        spawnDaemon(config: DaemonConfig.resolved)
        for _ in 0..<50 {
            if (try? await DaemonClient.shared.health()) == true {
                return
            }
            try? await Task.sleep(nanoseconds: 100_000_000)
        }
        print("Warning: daemon health check failed after spawn")
    }

    private func spawnDaemon(config: DaemonConfig.Snapshot) {
        let binary = daemonBinaryPath()
        let process = Process()
        process.executableURL = URL(fileURLWithPath: binary)
        process.arguments = [
            "serve",
            "--port", String(config.port),
            "--state-dir", config.stateDir,
        ]
        var env = ProcessInfo.processInfo.environment
        env["DAEMON_PORT"] = String(config.port)
        env["DAEMON_STATE_DIR"] = config.stateDir
        process.environment = env
        do {
            try FileManager.default.createDirectory(
                atPath: config.stateDir,
                withIntermediateDirectories: true
            )
            try process.run()
            daemonProcess = process
        } catch {
            print("Failed to spawn daemon at \(binary): \(error)")
        }
    }

    private func daemonBinaryPath() -> String {
        if let cli = ProcessInfo.processInfo.environment["DAEMON_CLI"], !cli.isEmpty {
            return cli
        }
        let bundled = Bundle.main.bundleURL
            .appendingPathComponent("Contents/MacOS/__DAEMON_NAME__")
            .path
        if FileManager.default.fileExists(atPath: bundled) {
            return bundled
        }
        return "/usr/local/bin/__DAEMON_NAME__"
    }
}

@available(macOS 15.0, *)
@main
struct MenuBarApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) var appDelegate
    @AppStorage("autoStart") private var autoStart = false

    init() {
        _autoStart.wrappedValue = SMAppService.mainApp.status == .enabled
    }

    var body: some Scene {
        Window("Settings", id: "settings") {
            SettingsView()
        }
        .windowResizability(.contentSize)
        .defaultLaunchBehavior(.suppressed)

        MenuBarExtra {
            MenuBarDropdownContent(
                autoStart: $autoStart,
                showSettings: showSettingsWindow,
                restartDaemon: { appDelegate.restartDaemon() },
                daemonInfo: { appDelegate.daemonInfo() }
            )
        } label: {
            Image(systemName: "bell")
                .imageScale(.small)
                .accessibilityIdentifier("menu-bar-extra")
        }
    }

    private func showSettingsWindow(openWindow: OpenWindowAction) {
        NSApp.setActivationPolicy(.regular)
        NSApp.activate(ignoringOtherApps: true)
        openWindow(id: "settings")
        if let window = NSApp.windows.first(where: { $0.title == "Settings" }) {
            window.makeKeyAndOrderFront(nil)
            return
        }
        Task { @MainActor in
            for _ in 0..<15 {
                openWindow(id: "settings")
                if let window = NSApp.windows.first(where: { $0.title == "Settings" }) {
                    window.makeKeyAndOrderFront(nil)
                    return
                }
                try? await Task.sleep(nanoseconds: 100_000_000)
            }
        }
    }
}

@available(macOS 15.0, *)
private struct MenuBarDropdownContent: View {
    @Binding var autoStart: Bool
    @AppStorage("defaultBrowser") private var defaultBrowser = BrowserPreference.default.rawValue
    @ObservedObject private var daemonReadiness = DaemonReadiness.shared
    @Environment(\.openWindow) private var openWindow
    let showSettings: (OpenWindowAction) -> Void
    let restartDaemon: () -> Void
    let daemonInfo: () -> (port: Int, pid: Int32)

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            Toggle("Auto Start", isOn: $autoStart)
                .padding(.horizontal, 8)
                .padding(.vertical, 3)
                .onChange(of: autoStart) { enabled in
                    do {
                        if enabled {
                            try SMAppService.mainApp.register()
                        } else {
                            try SMAppService.mainApp.unregister()
                        }
                    } catch {
                        print("Auto Start toggle failed: \(error)")
                        autoStart = !enabled
                    }
                }

            let info = daemonInfo()
            let portStr = info.port < 0 ? "-" : String(info.port)
            let pidStr = info.pid < 0 ? "-" : String(info.pid)
            Button("Restart Daemon (Port: \(portStr), PID: \(pidStr))") {
                restartDaemon()
            }
            .padding(.horizontal, 8)
            .padding(.vertical, 3)

            Button(OpenInBrowserLabelFormatter.format(browser: defaultBrowser)) {
                let info = daemonInfo()
                guard info.port >= 0, let url = URL(string: "http://127.0.0.1:\(info.port)") else {
                    return
                }
                BrowserOpener.open(
                    url: url,
                    browser: BrowserPreference.fromStored(defaultBrowser)
                )
            }
            .disabled(info.pid < 0)
            .padding(.horizontal, 8)
            .padding(.vertical, 3)

            Divider()

            SettingsMenuButton(showSettings: showSettings)

            Button("Quit") {
                NSApplication.shared.terminate(nil)
            }
            .padding(.horizontal, 8)
            .padding(.vertical, 3)
        }
        .padding(.vertical, 4)
        .frame(minWidth: 220)
        .accessibilityIdentifier("menu-bar-dropdown")
        .task {
            while !daemonReadiness.isReady {
                try? await Task.sleep(nanoseconds: 50_000_000)
            }
        }
    }
}

@available(macOS 15.0, *)
private struct SettingsMenuButton: View {
    @Environment(\.openWindow) private var openWindow
    let showSettings: (OpenWindowAction) -> Void

    var body: some View {
        Button("Settings…") {
            showSettings(openWindow)
        }
        .padding(.horizontal, 8)
        .padding(.vertical, 3)
        .accessibilityIdentifier("settings-menu-button")
    }
}