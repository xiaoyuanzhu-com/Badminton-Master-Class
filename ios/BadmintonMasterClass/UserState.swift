import Foundation
import Combine

// MARK: - Persistence model

private struct UserStateData: Codable {
    var favorites: [Int]
    var pathProgress: [String: [Int]]  // JSON keys must be strings

    init(favorites: [Int] = [], pathProgress: [String: [Int]] = [:]) {
        self.favorites = favorites
        self.pathProgress = pathProgress
    }
}

// MARK: - UserState

@MainActor
final class UserState: ObservableObject {
    @Published var favorites: [Int] = []
    @Published var pathProgress: [Int: Set<Int>] = [:]

    private var cancellable: AnyCancellable?

    private static var fileURL: URL {
        FileManager.default
            .urls(for: .documentDirectory, in: .userDomainMask)[0]
            .appendingPathComponent("user_state.json")
    }

    init() {
        load()
        // Auto-save whenever favorites or pathProgress change
        cancellable = Publishers.Merge(
            $favorites.map { _ in () },
            $pathProgress.map { _ in () }
        )
        .debounce(for: .milliseconds(300), scheduler: RunLoop.main)
        .sink { [weak self] in
            self?.save()
        }
    }

    // MARK: - Favorites

    func toggleFavorite(contentId: Int) {
        if let index = favorites.firstIndex(of: contentId) {
            favorites.remove(at: index)
        } else {
            favorites.append(contentId)
        }
    }

    func isFavorite(contentId: Int) -> Bool {
        favorites.contains(contentId)
    }

    // MARK: - Path Progress

    func toggleStepComplete(pathId: Int, stepId: Int) {
        var steps = pathProgress[pathId] ?? []
        if steps.contains(stepId) {
            steps.remove(stepId)
        } else {
            steps.insert(stepId)
        }
        pathProgress[pathId] = steps
    }

    func isStepComplete(pathId: Int, stepId: Int) -> Bool {
        pathProgress[pathId]?.contains(stepId) ?? false
    }

    // MARK: - Persistence

    private func load() {
        let url = Self.fileURL
        guard FileManager.default.fileExists(atPath: url.path) else { return }
        do {
            let data = try Data(contentsOf: url)
            let decoded = try JSONDecoder().decode(UserStateData.self, from: data)
            self.favorites = decoded.favorites
            self.pathProgress = decoded.pathProgress.reduce(into: [:]) { result, pair in
                if let key = Int(pair.key) {
                    result[key] = Set(pair.value)
                }
            }
        } catch {
            print("[UserState] Failed to load: \(error)")
        }
    }

    private func save() {
        let stringProgress = pathProgress.reduce(into: [String: [Int]]()) { result, pair in
            result[String(pair.key)] = Array(pair.value)
        }
        let stateData = UserStateData(favorites: favorites, pathProgress: stringProgress)
        do {
            let data = try JSONEncoder().encode(stateData)
            try data.write(to: Self.fileURL, options: .atomic)
        } catch {
            print("[UserState] Failed to save: \(error)")
        }
    }
}
