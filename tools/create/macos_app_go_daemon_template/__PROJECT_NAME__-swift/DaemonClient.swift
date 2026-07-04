import Foundation

struct DaemonInfo: Decodable {
    let storagePath: String
    let port: Int?

    enum CodingKeys: String, CodingKey {
        case storagePath = "storage_path"
        case port
    }
}

enum DaemonClientError: LocalizedError {
    case unreachable(String)
    case badStatus(Int, String)

    var errorDescription: String? {
        switch self {
        case .unreachable(let detail):
            return "daemon unreachable: \(detail)"
        case .badStatus(let code, let body):
            return "daemon returned \(code): \(body)"
        }
    }
}

final class DaemonClient {
    static let shared = DaemonClient()

    var port: Int {
        explicitPort ?? DaemonConfig.resolvedPort
    }

    private let explicitPort: Int?
    private let session: URLSession

    init(port: Int? = nil, session: URLSession = .shared) {
        self.explicitPort = port
        self.session = session
    }

    private var baseURL: String {
        "http://127.0.0.1:\(port)"
    }

    func info() async throws -> DaemonInfo {
        let (data, response) = try await get(path: "/api/info")
        try ensureOK(response, data: data)
        return try JSONDecoder().decode(DaemonInfo.self, from: data)
    }

    func health() async throws -> Bool {
        let (data, response) = try await get(path: "/api/health")
        guard let http = response as? HTTPURLResponse, http.statusCode == 200 else {
            throw DaemonClientError.unreachable("health check failed")
        }
        struct Payload: Decodable { let ok: Bool }
        let payload = try JSONDecoder().decode(Payload.self, from: data)
        return payload.ok
    }

    private func get(path: String) async throws -> (Data, URLResponse) {
        guard let url = URL(string: baseURL + path) else {
            throw DaemonClientError.unreachable("invalid URL")
        }
        var request = URLRequest(url: url)
        request.timeoutInterval = 5
        return try await session.data(for: request)
    }

    private func ensureOK(_ response: URLResponse, data: Data) throws {
        guard let http = response as? HTTPURLResponse else {
            throw DaemonClientError.unreachable("no HTTP response")
        }
        guard http.statusCode == 200 else {
            let body = String(data: data, encoding: .utf8) ?? ""
            throw DaemonClientError.badStatus(http.statusCode, body)
        }
    }
}