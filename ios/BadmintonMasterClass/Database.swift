import Foundation
import SQLite3

final class Database {
    static let shared = Database()

    private var db: OpaquePointer?
    private let dbName = "bmc.db"
    private var dbURL: URL {
        FileManager.default
            .urls(for: .documentDirectory, in: .userDomainMask)[0]
            .appendingPathComponent(dbName)
    }

    private init() {
        copyBundledDBIfNeeded()
        openDatabase()
    }

    // MARK: - Setup

    private func copyBundledDBIfNeeded() {
        let fileManager = FileManager.default
        guard !fileManager.fileExists(atPath: dbURL.path) else { return }

        guard let bundledURL = Bundle.main.url(forResource: "bmc", withExtension: "db") else {
            print("[Database] Bundled DB not found in bundle")
            return
        }

        do {
            try fileManager.copyItem(at: bundledURL, to: dbURL)
            print("[Database] Copied bundled DB to Documents")
        } catch {
            print("[Database] Failed to copy bundled DB: \(error)")
        }
    }

    private func openDatabase() {
        if sqlite3_open(dbURL.path, &db) != SQLITE_OK {
            print("[Database] Failed to open DB: \(String(cString: sqlite3_errmsg(db)))")
            db = nil
        }
    }

    private func closeDatabase() {
        if db != nil {
            sqlite3_close(db)
            db = nil
        }
    }

    // MARK: - Queries

    func categories(parentId: Int?) -> [Category] {
        var results: [Category] = []
        var stmt: OpaquePointer?

        let sql: String
        if let parentId = parentId {
            sql = "SELECT id, name, icon, sort_order, parent_id FROM categories WHERE parent_id = ? ORDER BY sort_order"
            guard sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK else { return results }
            sqlite3_bind_int(stmt, 1, Int32(parentId))
        } else {
            sql = "SELECT id, name, icon, sort_order, parent_id FROM categories WHERE parent_id IS NULL ORDER BY sort_order"
            guard sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK else { return results }
        }

        while sqlite3_step(stmt) == SQLITE_ROW {
            let id = Int(sqlite3_column_int(stmt, 0))
            let name = String(cString: sqlite3_column_text(stmt, 1))
            let icon = String(cString: sqlite3_column_text(stmt, 2))
            let sortOrder = Int(sqlite3_column_int(stmt, 3))
            let pid: Int? = sqlite3_column_type(stmt, 4) == SQLITE_NULL
                ? nil
                : Int(sqlite3_column_int(stmt, 4))

            results.append(Category(id: id, name: name, icon: icon, sortOrder: sortOrder, parentId: pid))
        }

        sqlite3_finalize(stmt)
        return results
    }

    func contents(categoryId: Int) -> [ContentItem] {
        var results: [ContentItem] = []
        var stmt: OpaquePointer?
        let sql = "SELECT id, title, summary, thumbnail_url, source_url, source_platform, author_name, category_id, sort_order FROM contents WHERE category_id = ? ORDER BY sort_order"

        guard sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK else { return results }
        sqlite3_bind_int(stmt, 1, Int32(categoryId))

        while sqlite3_step(stmt) == SQLITE_ROW {
            let id = Int(sqlite3_column_int(stmt, 0))
            let title = String(cString: sqlite3_column_text(stmt, 1))
            let summary = String(cString: sqlite3_column_text(stmt, 2))
            let thumbnailUrl = String(cString: sqlite3_column_text(stmt, 3))
            let sourceUrl = String(cString: sqlite3_column_text(stmt, 4))
            let sourcePlatform = String(cString: sqlite3_column_text(stmt, 5))
            let authorName = String(cString: sqlite3_column_text(stmt, 6))
            let categoryId = Int(sqlite3_column_int(stmt, 7))
            let sortOrder = Int(sqlite3_column_int(stmt, 8))

            results.append(ContentItem(
                id: id, title: title, summary: summary,
                thumbnailUrl: thumbnailUrl, sourceUrl: sourceUrl,
                sourcePlatform: sourcePlatform, authorName: authorName,
                categoryId: categoryId, sortOrder: sortOrder
            ))
        }

        sqlite3_finalize(stmt)
        return results
    }

    func learningPaths() -> [LearningPath] {
        var results: [LearningPath] = []
        var stmt: OpaquePointer?
        let sql = """
            SELECT lp.id, lp.title, lp.summary, lp.difficulty, lp.sort_order,
                   (SELECT COUNT(*) FROM path_steps WHERE path_id = lp.id) AS step_count
            FROM learning_paths lp
            ORDER BY lp.sort_order
            """

        guard sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK else { return results }

        while sqlite3_step(stmt) == SQLITE_ROW {
            let id = Int(sqlite3_column_int(stmt, 0))
            let title = String(cString: sqlite3_column_text(stmt, 1))
            let summary = String(cString: sqlite3_column_text(stmt, 2))
            let difficulty = String(cString: sqlite3_column_text(stmt, 3))
            let sortOrder = Int(sqlite3_column_int(stmt, 4))
            let stepCount = Int(sqlite3_column_int(stmt, 5))

            results.append(LearningPath(
                id: id, title: title, summary: summary,
                difficulty: difficulty, sortOrder: sortOrder,
                stepCount: stepCount
            ))
        }

        sqlite3_finalize(stmt)
        return results
    }

    func pathSteps(pathId: Int) -> [PathStep] {
        var results: [PathStep] = []
        var stmt: OpaquePointer?
        let sql = "SELECT id, path_id, step_order, day, title, note FROM path_steps WHERE path_id = ? ORDER BY step_order"

        guard sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK else { return results }
        sqlite3_bind_int(stmt, 1, Int32(pathId))

        while sqlite3_step(stmt) == SQLITE_ROW {
            let id = Int(sqlite3_column_int(stmt, 0))
            let pathId = Int(sqlite3_column_int(stmt, 1))
            let stepOrder = Int(sqlite3_column_int(stmt, 2))
            let day = sqlite3_column_text(stmt, 3).map { String(cString: $0) } ?? ""
            let title = String(cString: sqlite3_column_text(stmt, 4))
            let note = sqlite3_column_text(stmt, 5).map { String(cString: $0) } ?? ""

            results.append(PathStep(
                id: id, pathId: pathId, stepOrder: stepOrder,
                day: day, title: title, note: note
            ))
        }

        sqlite3_finalize(stmt)
        return results
    }

    func pathStepContents(stepId: Int) -> [ContentItem] {
        var results: [ContentItem] = []
        var stmt: OpaquePointer?
        let sql = """
            SELECT c.id, c.title, c.summary, c.thumbnail_url, c.source_url,
                   c.source_platform, c.author_name, c.category_id, c.sort_order
            FROM contents c
            JOIN path_step_contents psc ON psc.content_id = c.id
            WHERE psc.step_id = ?
            ORDER BY psc.sort_order
            """

        guard sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK else { return results }
        sqlite3_bind_int(stmt, 1, Int32(stepId))

        while sqlite3_step(stmt) == SQLITE_ROW {
            let id = Int(sqlite3_column_int(stmt, 0))
            let title = String(cString: sqlite3_column_text(stmt, 1))
            let summary = String(cString: sqlite3_column_text(stmt, 2))
            let thumbnailUrl = String(cString: sqlite3_column_text(stmt, 3))
            let sourceUrl = String(cString: sqlite3_column_text(stmt, 4))
            let sourcePlatform = String(cString: sqlite3_column_text(stmt, 5))
            let authorName = String(cString: sqlite3_column_text(stmt, 6))
            let categoryId = Int(sqlite3_column_int(stmt, 7))
            let sortOrder = Int(sqlite3_column_int(stmt, 8))

            results.append(ContentItem(
                id: id, title: title, summary: summary,
                thumbnailUrl: thumbnailUrl, sourceUrl: sourceUrl,
                sourcePlatform: sourcePlatform, authorName: authorName,
                categoryId: categoryId, sortOrder: sortOrder
            ))
        }

        sqlite3_finalize(stmt)
        return results
    }

    func contentsByIds(_ ids: [Int]) -> [ContentItem] {
        guard !ids.isEmpty else { return [] }
        var results: [ContentItem] = []
        var stmt: OpaquePointer?
        let placeholders = ids.map { _ in "?" }.joined(separator: ",")
        let sql = "SELECT id, title, summary, thumbnail_url, source_url, source_platform, author_name, category_id, sort_order FROM contents WHERE id IN (\(placeholders))"

        guard sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK else { return results }
        for (index, id) in ids.enumerated() {
            sqlite3_bind_int(stmt, Int32(index + 1), Int32(id))
        }

        while sqlite3_step(stmt) == SQLITE_ROW {
            let id = Int(sqlite3_column_int(stmt, 0))
            let title = String(cString: sqlite3_column_text(stmt, 1))
            let summary = String(cString: sqlite3_column_text(stmt, 2))
            let thumbnailUrl = String(cString: sqlite3_column_text(stmt, 3))
            let sourceUrl = String(cString: sqlite3_column_text(stmt, 4))
            let sourcePlatform = String(cString: sqlite3_column_text(stmt, 5))
            let authorName = String(cString: sqlite3_column_text(stmt, 6))
            let categoryId = Int(sqlite3_column_int(stmt, 7))
            let sortOrder = Int(sqlite3_column_int(stmt, 8))

            results.append(ContentItem(
                id: id, title: title, summary: summary,
                thumbnailUrl: thumbnailUrl, sourceUrl: sourceUrl,
                sourcePlatform: sourcePlatform, authorName: authorName,
                categoryId: categoryId, sortOrder: sortOrder
            ))
        }

        sqlite3_finalize(stmt)
        // Preserve the order of input IDs
        let lookup = Dictionary(uniqueKeysWithValues: results.map { ($0.id, $0) })
        return ids.compactMap { lookup[$0] }
    }

    func searchContents(keyword: String) -> [ContentItem] {
        var results: [ContentItem] = []
        guard !keyword.trimmingCharacters(in: .whitespaces).isEmpty else { return results }
        var stmt: OpaquePointer?
        let sql = """
            SELECT c.id, c.title, c.summary, c.thumbnail_url, c.source_url,
                   c.source_platform, c.author_name, c.category_id, c.sort_order,
                   COALESCE(cat.name, '') AS category_name
            FROM contents c
            LEFT JOIN categories cat ON cat.id = c.category_id
            WHERE c.title LIKE ? OR c.summary LIKE ? OR c.author_name LIKE ?
            ORDER BY c.sort_order
            """

        guard sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK else { return results }
        let pattern = "%\(keyword)%"
        sqlite3_bind_text(stmt, 1, (pattern as NSString).utf8String, -1, nil)
        sqlite3_bind_text(stmt, 2, (pattern as NSString).utf8String, -1, nil)
        sqlite3_bind_text(stmt, 3, (pattern as NSString).utf8String, -1, nil)

        while sqlite3_step(stmt) == SQLITE_ROW {
            let id = Int(sqlite3_column_int(stmt, 0))
            let title = String(cString: sqlite3_column_text(stmt, 1))
            let summary = String(cString: sqlite3_column_text(stmt, 2))
            let thumbnailUrl = String(cString: sqlite3_column_text(stmt, 3))
            let sourceUrl = String(cString: sqlite3_column_text(stmt, 4))
            let sourcePlatform = String(cString: sqlite3_column_text(stmt, 5))
            let authorName = String(cString: sqlite3_column_text(stmt, 6))
            let categoryId = Int(sqlite3_column_int(stmt, 7))
            let sortOrder = Int(sqlite3_column_int(stmt, 8))
            let categoryName = String(cString: sqlite3_column_text(stmt, 9))

            results.append(ContentItem(
                id: id, title: title, summary: summary,
                thumbnailUrl: thumbnailUrl, sourceUrl: sourceUrl,
                sourcePlatform: sourcePlatform, authorName: authorName,
                categoryId: categoryId, sortOrder: sortOrder,
                categoryName: categoryName
            ))
        }

        sqlite3_finalize(stmt)
        return results
    }

    func searchLearningPaths(keyword: String) -> [LearningPath] {
        var results: [LearningPath] = []
        guard !keyword.trimmingCharacters(in: .whitespaces).isEmpty else { return results }
        var stmt: OpaquePointer?
        let sql = """
            SELECT lp.id, lp.title, lp.summary, lp.difficulty, lp.sort_order,
                   (SELECT COUNT(*) FROM path_steps WHERE path_id = lp.id) AS step_count
            FROM learning_paths lp
            WHERE lp.title LIKE ? OR lp.summary LIKE ?
            ORDER BY lp.sort_order
            """

        guard sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK else { return results }
        let pattern = "%\(keyword)%"
        sqlite3_bind_text(stmt, 1, (pattern as NSString).utf8String, -1, nil)
        sqlite3_bind_text(stmt, 2, (pattern as NSString).utf8String, -1, nil)

        while sqlite3_step(stmt) == SQLITE_ROW {
            let id = Int(sqlite3_column_int(stmt, 0))
            let title = String(cString: sqlite3_column_text(stmt, 1))
            let summary = String(cString: sqlite3_column_text(stmt, 2))
            let difficulty = String(cString: sqlite3_column_text(stmt, 3))
            let sortOrder = Int(sqlite3_column_int(stmt, 4))
            let stepCount = Int(sqlite3_column_int(stmt, 5))

            results.append(LearningPath(
                id: id, title: title, summary: summary,
                difficulty: difficulty, sortOrder: sortOrder,
                stepCount: stepCount
            ))
        }

        sqlite3_finalize(stmt)
        return results
    }

    // MARK: - Async wrappers (off-main-thread queries)

    private static let queryQueue = DispatchQueue(label: "com.bmc.db.query", qos: .userInitiated)

    func categoriesAsync(parentId: Int?) async -> [Category] {
        await withCheckedContinuation { continuation in
            Self.queryQueue.async {
                let result = self.categories(parentId: parentId)
                continuation.resume(returning: result)
            }
        }
    }

    func contentsAsync(categoryId: Int) async -> [ContentItem] {
        await withCheckedContinuation { continuation in
            Self.queryQueue.async {
                let result = self.contents(categoryId: categoryId)
                continuation.resume(returning: result)
            }
        }
    }

    func contentsByIdsAsync(_ ids: [Int]) async -> [ContentItem] {
        await withCheckedContinuation { continuation in
            Self.queryQueue.async {
                let result = self.contentsByIds(ids)
                continuation.resume(returning: result)
            }
        }
    }

    func searchContentsAsync(keyword: String) async -> [ContentItem] {
        await withCheckedContinuation { continuation in
            Self.queryQueue.async {
                let result = self.searchContents(keyword: keyword)
                continuation.resume(returning: result)
            }
        }
    }

    func searchLearningPathsAsync(keyword: String) async -> [LearningPath] {
        await withCheckedContinuation { continuation in
            Self.queryQueue.async {
                let result = self.searchLearningPaths(keyword: keyword)
                continuation.resume(returning: result)
            }
        }
    }

    func learningPathsAsync() async -> [LearningPath] {
        await withCheckedContinuation { continuation in
            Self.queryQueue.async {
                let result = self.learningPaths()
                continuation.resume(returning: result)
            }
        }
    }

    func pathStepsAsync(pathId: Int) async -> [PathStep] {
        await withCheckedContinuation { continuation in
            Self.queryQueue.async {
                let result = self.pathSteps(pathId: pathId)
                continuation.resume(returning: result)
            }
        }
    }

    func pathStepContentsAsync(stepId: Int) async -> [ContentItem] {
        await withCheckedContinuation { continuation in
            Self.queryQueue.async {
                let result = self.pathStepContents(stepId: stepId)
                continuation.resume(returning: result)
            }
        }
    }

    // MARK: - Replace DB (for sync)

    func replaceWith(downloadedDBAt url: URL) {
        closeDatabase()

        let fileManager = FileManager.default
        do {
            if fileManager.fileExists(atPath: dbURL.path) {
                try fileManager.removeItem(at: dbURL)
            }
            try fileManager.moveItem(at: url, to: dbURL)
            print("[Database] Replaced DB with downloaded version")
        } catch {
            print("[Database] Failed to replace DB: \(error)")
        }

        openDatabase()
    }
}
