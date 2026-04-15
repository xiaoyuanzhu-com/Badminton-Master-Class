import Foundation

struct Category: Identifiable, Hashable {
    let id: Int
    let name: String
    let icon: String
    let sortOrder: Int
    let parentId: Int?
}

struct ContentItem: Identifiable {
    let id: Int
    let title: String
    let summary: String
    let thumbnailUrl: String
    let sourceUrl: String
    let sourcePlatform: String
    let authorName: String
    let categoryId: Int
    let sortOrder: Int
}

struct LearningPath: Identifiable, Hashable {
    let id: Int
    let title: String
    let summary: String
    let difficulty: String
    let sortOrder: Int
    var stepCount: Int = 0
}

struct PathStep: Identifiable {
    let id: Int
    let pathId: Int
    let stepOrder: Int
    let day: String
    let title: String
    let note: String
}
