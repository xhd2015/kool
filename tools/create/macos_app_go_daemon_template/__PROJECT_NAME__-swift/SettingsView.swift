import SwiftUI

struct SettingsView: View {
    @AppStorage("defaultBrowser") private var defaultBrowser = BrowserPreference.default.rawValue

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Settings")
                .font(.title2)
                .fontWeight(.semibold)

            Divider()

            VStack(alignment: .leading, spacing: 8) {
                Text("Default Browser")
                    .font(.headline)

                Text("Choose which browser opens when you click Open in Browser:")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .fixedSize(horizontal: false, vertical: true)

                Picker("Open with", selection: $defaultBrowser) {
                    ForEach(BrowserPreference.allCases) { preference in
                        Text(preference.displayName).tag(preference.rawValue)
                    }
                }
                .pickerStyle(.radioGroup)
                .accessibilityIdentifier("browser-picker")
            }
            .accessibilityIdentifier("default-browser-section")
        }
        .padding(16)
        .frame(minWidth: 400, minHeight: 200)
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("settings-window")
    }
}