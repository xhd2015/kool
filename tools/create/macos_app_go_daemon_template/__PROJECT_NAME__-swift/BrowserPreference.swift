import Foundation

enum BrowserPreference: String, CaseIterable, Identifiable {
    case `default` = "default"
    case chrome = "chrome"
    case firefox = "firefox"
    case opera = "opera"

    var id: String { rawValue }

    var displayName: String {
        switch self {
        case .default: return "Default"
        case .chrome: return "Chrome"
        case .firefox: return "Firefox"
        case .opera: return "Opera"
        }
    }

    static func fromStored(_ value: String) -> BrowserPreference {
        BrowserPreference(rawValue: value) ?? .default
    }
}