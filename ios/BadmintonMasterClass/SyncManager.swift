import Foundation

enum SyncState: Equatable {
    case idle
    case syncing
    case success
    case failed
}

@MainActor
final class SyncManager: ObservableObject {
    static let shared = SyncManager()

    @Published private(set) var state: SyncState = .idle

    private init() {}

    func setSyncing() {
        state = .syncing
    }

    func setSuccess() {
        state = .success
        // Auto-dismiss after 2 seconds
        Task {
            try? await Task.sleep(nanoseconds: 2_000_000_000)
            if state == .success {
                state = .idle
            }
        }
    }

    func setFailed() {
        state = .failed
        // Auto-dismiss after 3 seconds
        Task {
            try? await Task.sleep(nanoseconds: 3_000_000_000)
            if state == .failed {
                state = .idle
            }
        }
    }
}
