import Foundation

enum OpenInBrowserLabelFormatter {
    static func format(browser: String) -> String {
        switch browser {
        case "", "default":
            return "Open in Browser"
        case "chrome":
            return "Open in Browser(Chrome)"
        case "firefox":
            return "Open in Browser(Firefox)"
        case "opera":
            return "Open in Browser(Opera)"
        default:
            return "Open in Browser"
        }
    }
}