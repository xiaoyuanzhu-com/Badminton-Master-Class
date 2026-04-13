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
