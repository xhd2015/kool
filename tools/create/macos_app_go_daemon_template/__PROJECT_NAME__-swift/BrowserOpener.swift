import AppKit
import Foundation

enum BrowserOpener {
    private static let bundleIDs: [BrowserPreference: String] = [
        .chrome: "com.google.Chrome",
        .firefox: "org.mozilla.firefox",
        .opera: "com.operasoftware.Opera",
    ]

    private static let appNames: [BrowserPreference: String] = [
        .chrome: "Google Chrome",
        .firefox: "Firefox",
        .opera: "Opera",
    ]

    static func open(url: URL, browser: BrowserPreference) {
        switch browser {
        case .default:
            NSWorkspace.shared.open(url)
            return
        case .chrome, .firefox, .opera:
            if let appURL = applicationURL(for: browser) {
                let configuration = NSWorkspace.OpenConfiguration()
                NSWorkspace.shared.open([url], withApplicationAt: appURL, configuration: configuration) { _, error in
                    if error != nil {
                        NSWorkspace.shared.open(url)
                    }
                }
                return
            }
            openWithCLI(url: url, browser: browser)
        }
    }

    private static func applicationURL(for browser: BrowserPreference) -> URL? {
        guard let bundleID = bundleIDs[browser] else {
            return nil
        }
        return NSWorkspace.shared.urlForApplication(withBundleIdentifier: bundleID)
    }

    private static func openWithCLI(url: URL, browser: BrowserPreference) {
        guard let appName = appNames[browser] else {
            NSWorkspace.shared.open(url)
            return
        }
        let process = Process()
        process.executableURL = URL(fileURLWithPath: "/usr/bin/open")
        process.arguments = ["-a", appName, url.absoluteString]
        do {
            try process.run()
            process.waitUntilExit()
            if process.terminationStatus != 0 {
                NSWorkspace.shared.open(url)
            }
        } catch {
            NSWorkspace.shared.open(url)
        }
    }
}