import Foundation

enum DaemonDebug {
    static var isEnabled: Bool {
        #if DAEMON_DEBUG
        return true
        #else
        return false
        #endif
    }

    static let bundleID = "__BUNDLE_ID__.debug"
    static let appName = "__PROJECT_NAME__-debug"
}