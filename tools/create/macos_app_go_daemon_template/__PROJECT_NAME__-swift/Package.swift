// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "__PROJECT_NAME__-swift",
    platforms: [.macOS(.v13)],
    products: [
        .executable(name: "__PROJECT_NAME__", targets: ["__PROJECT_NAME__"]),
    ],
    targets: [
        .executableTarget(
            name: "__PROJECT_NAME__",
            path: ".",
            swiftSettings: [
                .define("DAEMON_DEBUG", .when(configuration: .debug)),
            ]
        ),
    ]
)