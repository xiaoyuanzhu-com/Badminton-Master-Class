import Foundation
import SQLite3

// MARK: - Serialization model
//
// Database is a Swift actor. All methods — queries and replaceWith() — execute
// on the actor's serial executor. This means:
//   • Concurrent callers (e.g. PathDetailView's withTaskGroup) are automatically
//     serialized; no two methods touch `db` at the same time.
//   • replaceWith() (called from DataSync after a successful download) cannot
//     run while a query is in progress, eliminating the sqlite3_close() / active
//     statement race that could crash on pull-to-refresh.
//   • File I/O inside replaceWith() runs on the actor executor, not MainActor.
//
// Call sites use the existing `xxxAsync` names; no view changes are needed.

actor Database {
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

    private func categories(parentId: Int?) -> [Category] {
        var results: [Category] = []
        var stmt: OpaquePointer?

        let sql: String
        if let parentId = parentId {
            sql = """
                SELECT c.id, c.name, c.icon, c.sort_order, c.parent_id,
                       (SELECT COUNT(*) FROM contents WHERE category_id = c.id)
                       + (SELECT COALESCE(SUM(sub_count), 0) FROM (
                           SELECT (SELECT COUNT(*) FROM contents WHERE category_id = sc.id) AS sub_count
                           FROM categories sc WHERE sc.parent_id = c.id
                       ))
                FROM categories c WHERE c.parent_id = ? ORDER BY c.sort_order
                """
            guard sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK else { return results }
            sqlite3_bind_int(stmt, 1, Int32(parentId))
        } else {
            sql = """
                SELECT c.id, c.name, c.icon, c.sort_order, c.parent_id,
                       (SELECT COUNT(*) FROM contents WHERE category_id = c.id)
                       + (SELECT COALESCE(SUM(sub_count), 0) FROM (
                           SELECT (SELECT COUNT(*) FROM contents WHERE category_id = sc.id) AS sub_count
                           FROM categories sc WHERE sc.parent_id = c.id
                       ))
                FROM categories c WHERE c.parent_id IS NULL ORDER BY c.sort_order
                """
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
            let contentCount = Int(sqlite3_column_int(stmt, 5))

            results.append(Category(id: id, name: name, icon: icon, sortOrder: sortOrder, parentId: pid, contentCount: contentCount))
        }

        sqlite3_finalize(stmt)
        return results
    }

    private func contents(categoryId: Int) -> [ContentItem] {
        var results: [ContentItem] = []
        var stmt: OpaquePointer?
        let sql = "SELECT id, title, summary, thumbnail_url, source_url, source_platform, author_name, difficulty, duration, editor_notes, category_id, sort_order FROM contents WHERE category_id = ? ORDER BY sort_order"

        guard sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK else { return results }
        sqlite3_bind_int(stmt, 1, Int32(categoryId))

        while sqlite3_step(stmt) == SQLITE_ROW {
            results.append(parseContentRow(stmt))
        }

        sqlite3_finalize(stmt)
        return results
    }

    private func parseContentRow(_ stmt: OpaquePointer?) -> ContentItem {
        let id = Int(sqlite3_column_int(stmt, 0))
        let title = String(cString: sqlite3_column_text(stmt, 1))
        let summary = String(cString: sqlite3_column_text(stmt, 2))
        let thumbnailUrl = String(cString: sqlite3_column_text(stmt, 3))
        let sourceUrl = String(cString: sqlite3_column_text(stmt, 4))
        let sourcePlatform = String(cString: sqlite3_column_text(stmt, 5))
        let authorName = String(cString: sqlite3_column_text(stmt, 6))
        let difficulty = String(cString: sqlite3_column_text(stmt, 7))
        let duration = String(cString: sqlite3_column_text(stmt, 8))
        let editorNotes = String(cString: sqlite3_column_text(stmt, 9))
        let categoryId = Int(sqlite3_column_int(stmt, 10))
        let sortOrder = Int(sqlite3_column_int(stmt, 11))

        return ContentItem(
            id: id, title: title, summary: summary,
            thumbnailUrl: thumbnailUrl, sourceUrl: sourceUrl,
            sourcePlatform: sourcePlatform, authorName: authorName,
            difficulty: difficulty, duration: duration,
            editorNotes: editorNotes,
            categoryId: categoryId, sortOrder: sortOrder
        )
    }

    private func learningPaths() -> [LearningPath] {
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

    private func pathSteps(pathId: Int) -> [PathStep] {
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

    private func pathStepContents(stepId: Int) -> [ContentItem] {
        var results: [ContentItem] = []
        var stmt: OpaquePointer?
        let sql = """
            SELECT c.id, c.title, c.summary, c.thumbnail_url, c.source_url,
                   c.source_platform, c.author_name, c.difficulty, c.duration,
                   c.editor_notes, c.category_id, c.sort_order
            FROM contents c
            JOIN path_step_contents psc ON psc.content_id = c.id
            WHERE psc.step_id = ?
            ORDER BY psc.sort_order
            """

        guard sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK else { return results }
        sqlite3_bind_int(stmt, 1, Int32(stepId))

        while sqlite3_step(stmt) == SQLITE_ROW {
            results.append(parseContentRow(stmt))
        }

        sqlite3_finalize(stmt)
        return results
    }

    private func contentsByIds(_ ids: [Int]) -> [ContentItem] {
        guard !ids.isEmpty else { return [] }
        var results: [ContentItem] = []
        var stmt: OpaquePointer?
        let placeholders = ids.map { _ in "?" }.joined(separator: ",")
        let sql = "SELECT id, title, summary, thumbnail_url, source_url, source_platform, author_name, difficulty, duration, editor_notes, category_id, sort_order FROM contents WHERE id IN (\(placeholders))"

        guard sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK else { return results }
        for (index, id) in ids.enumerated() {
            sqlite3_bind_int(stmt, Int32(index + 1), Int32(id))
        }

        while sqlite3_step(stmt) == SQLITE_ROW {
            results.append(parseContentRow(stmt))
        }

        sqlite3_finalize(stmt)
        // Preserve the order of input IDs
        let lookup = Dictionary(uniqueKeysWithValues: results.map { ($0.id, $0) })
        return ids.compactMap { lookup[$0] }
    }

    private func searchContents(keyword: String) -> [ContentItem] {
        var results: [ContentItem] = []
        guard !keyword.trimmingCharacters(in: .whitespaces).isEmpty else { return results }
        var stmt: OpaquePointer?
        let sql = """
            SELECT c.id, c.title, c.summary, c.thumbnail_url, c.source_url,
                   c.source_platform, c.author_name, c.difficulty, c.duration,
                   c.editor_notes, c.category_id, c.sort_order,
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
            var item = parseContentRow(stmt)
            item.categoryName = String(cString: sqlite3_column_text(stmt, 12))
            results.append(item)
        }

        sqlite3_finalize(stmt)
        return results
    }

    private func searchLearningPaths(keyword: String) -> [LearningPath] {
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

    // MARK: - Public async API (actor methods — serialized by the actor executor)

    func categoriesAsync(parentId: Int?) async -> [Category] {
        categories(parentId: parentId)
    }

    func contentsAsync(categoryId: Int) async -> [ContentItem] {
        contents(categoryId: categoryId)
    }

    func contentsByIdsAsync(_ ids: [Int]) async -> [ContentItem] {
        contentsByIds(ids)
    }

    func searchContentsAsync(keyword: String) async -> [ContentItem] {
        searchContents(keyword: keyword)
    }

    func searchLearningPathsAsync(keyword: String) async -> [LearningPath] {
        searchLearningPaths(keyword: keyword)
    }

    func learningPathsAsync() async -> [LearningPath] {
        learningPaths()
    }

    func pathStepsAsync(pathId: Int) async -> [PathStep] {
        pathSteps(pathId: pathId)
    }

    func pathStepContentsAsync(stepId: Int) async -> [ContentItem] {
        pathStepContents(stepId: stepId)
    }

    // MARK: - Replace DB (for sync)
    //
    // Runs on the actor executor — serialized with all queries. The file I/O
    // (close / remove / move / open) completes fully before any subsequent
    // query can begin, preventing sqlite3_close() from racing with active
    // sqlite3_step() calls in concurrent query tasks.

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
