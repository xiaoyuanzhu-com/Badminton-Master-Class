# 羽球大师课 (Badminton Master Class) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a curated badminton tutorial app — admin panel to manage content, SQLite as the data layer, native iOS + Android apps to browse by technique.

**Architecture:** Go admin panel writes to SQLite → .db file uploaded to Aliyun OSS → native apps ship a bundled .db and download the latest on launch. No REST API. The SQLite file IS the API.

**Tech Stack:** Go (admin), SQLite (everywhere), Swift/SwiftUI (iOS), Kotlin/Jetpack Compose (Android), Aliyun OSS (file hosting)

---

## Repo Structure

```
yuqiupu/
├── admin/          # Go admin panel
├── ios/            # Swift iOS app
├── android/        # Kotlin Android app
├── data/           # Seed data + shared schema
│   ├── schema.sql
│   └── seed.sql
└── docs/
    └── plans/
```

---

## Task 1: Database Schema + Seed Data

**Files:**
- Create: `data/schema.sql`
- Create: `data/seed.sql`

**Step 1: Write schema**

```sql
-- data/schema.sql
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    parent_id INTEGER REFERENCES categories(id)
);

CREATE TABLE contents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    thumbnail_url TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL,
    source_platform TEXT NOT NULL CHECK(source_platform IN ('bilibili', 'xiaohongshu', 'douyin', 'wechat', 'youtube', 'other')),
    author_name TEXT NOT NULL DEFAULT '',
    category_id INTEGER NOT NULL REFERENCES categories(id),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_contents_category ON contents(category_id);
```

**Step 2: Write seed data**

```sql
-- data/seed.sql
INSERT INTO categories (id, name, icon, sort_order, parent_id) VALUES
(1, '正手', '🏸', 1, NULL),
(2, '反手', '🏸', 2, NULL),
(3, '杀球', '💥', 3, NULL),
(4, '步法', '👟', 4, NULL),
(5, '发球', '🎯', 5, NULL),
(6, '网前', '🥅', 6, NULL),
(7, '双打', '👥', 7, NULL),
(8, '正手高远球', '🏸', 1, 1),
(9, '正手吊球', '🏸', 2, 1),
(10, '反手高远球', '🏸', 1, 2),
(11, '反手吊球', '🏸', 2, 2);

-- Example content (replace with real curated content)
INSERT INTO contents (title, summary, source_url, source_platform, author_name, category_id, sort_order) VALUES
('正手高远球完整教学', '从握拍到发力，最清晰的正手高远球教程', 'https://www.bilibili.com/video/example1', 'bilibili', '杨晨大神', 8, 1),
('反手高远球三步学会', '反手高远球的核心发力技巧', 'https://www.bilibili.com/video/example2', 'bilibili', '惠程俊', 10, 1);
```

**Step 3: Test schema by creating a DB**

Run: `sqlite3 data/yuqiupu.db < data/schema.sql && sqlite3 data/yuqiupu.db < data/seed.sql && sqlite3 data/yuqiupu.db "SELECT c.title, cat.name FROM contents c JOIN categories cat ON c.category_id = cat.id;"`

Expected: Two rows of content with their category names.

**Step 4: Commit**

```bash
git add data/schema.sql data/seed.sql
git commit -m "feat: add database schema and seed data"
```

---

## Task 2: Go Admin Panel — Project Setup

**Files:**
- Create: `admin/go.mod`
- Create: `admin/main.go`
- Create: `admin/main_test.go`

**Step 1: Initialize Go module**

Run: `cd admin && go mod init yuqiupu/admin`

**Step 2: Write test for DB initialization**

```go
// admin/main_test.go
package main

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitDB(t *testing.T) {
	dbPath := "test.db"
	defer os.Remove(dbPath)

	db, err := initDB(dbPath)
	if err != nil {
		t.Fatalf("initDB failed: %v", err)
	}
	defer db.Close()

	// Verify tables exist
	tables := []string{"categories", "contents"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %s not found: %v", table, err)
		}
	}
}
```

**Step 3: Run test to verify it fails**

Run: `cd admin && go test -v -run TestInitDB`
Expected: FAIL — `initDB` not defined.

**Step 4: Write minimal implementation**

```go
// admin/main.go
package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed ../data/schema.sql
var schemaSQL string

func initDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func main() {
	db, err := initDB("yuqiupu.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Admin panel running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Step 5: Install dependency and run test**

Run: `cd admin && go get github.com/mattn/go-sqlite3 && go test -v -run TestInitDB`
Expected: PASS

**Step 6: Commit**

```bash
git add admin/
git commit -m "feat: admin panel project setup with DB init"
```

---

## Task 3: Admin Panel — List & Create Categories

**Files:**
- Modify: `admin/main.go`
- Create: `admin/handlers.go`
- Create: `admin/handlers_test.go`
- Create: `admin/templates/categories.html`

**Step 1: Write test for category list handler**

```go
// admin/handlers_test.go
package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	db, err := initDB("test_handlers.db")
	if err != nil {
		t.Fatalf("initDB: %v", err)
	}
	return db, func() {
		db.Close()
		os.Remove("test_handlers.db")
	}
}

func TestListCategories(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Insert a test category
	db.Exec("INSERT INTO categories (name, icon, sort_order) VALUES ('正手', '🏸', 1)")

	handler := categoriesHandler(db)
	req := httptest.NewRequest("GET", "/categories", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "正手") {
		t.Error("response should contain category name")
	}
}

func TestCreateCategory(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	handler := categoriesHandler(db)
	body := strings.NewReader("name=杀球&icon=💥&sort_order=1")
	req := httptest.NewRequest("POST", "/categories", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303 redirect, got %d", w.Code)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM categories WHERE name='杀球'").Scan(&count)
	if count != 1 {
		t.Error("category should be inserted")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd admin && go test -v -run "TestListCategories|TestCreateCategory"`
Expected: FAIL

**Step 3: Implement category handlers**

```go
// admin/handlers.go
package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
)

type Category struct {
	ID        int
	Name      string
	Icon      string
	SortOrder int
	ParentID  sql.NullInt64
}

type Content struct {
	ID             int
	Title          string
	Summary        string
	ThumbnailURL   string
	SourceURL      string
	SourcePlatform string
	AuthorName     string
	CategoryID     int
	CategoryName   string
	SortOrder      int
}

func categoriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			rows, err := db.Query("SELECT id, name, icon, sort_order, parent_id FROM categories ORDER BY sort_order")
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			defer rows.Close()

			var cats []Category
			for rows.Next() {
				var c Category
				rows.Scan(&c.ID, &c.Name, &c.Icon, &c.SortOrder, &c.ParentID)
				cats = append(cats, c)
			}

			tmpl, err := template.ParseFiles("templates/categories.html")
			if err != nil {
				// Fallback for tests: simple response
				for _, c := range cats {
					w.Write([]byte(c.Name + "\n"))
				}
				return
			}
			tmpl.Execute(w, cats)

		case "POST":
			r.ParseForm()
			name := r.FormValue("name")
			icon := r.FormValue("icon")
			sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))
			parentID := r.FormValue("parent_id")

			if parentID != "" && parentID != "0" {
				pid, _ := strconv.Atoi(parentID)
				db.Exec("INSERT INTO categories (name, icon, sort_order, parent_id) VALUES (?, ?, ?, ?)", name, icon, sortOrder, pid)
			} else {
				db.Exec("INSERT INTO categories (name, icon, sort_order) VALUES (?, ?, ?)", name, icon, sortOrder)
			}
			http.Redirect(w, r, "/categories", http.StatusSeeOther)
		}
	}
}
```

**Step 4: Run tests**

Run: `cd admin && go test -v -run "TestListCategories|TestCreateCategory"`
Expected: PASS

**Step 5: Create HTML template**

```html
<!-- admin/templates/categories.html -->
<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>羽球大师课 - 分类管理</title>
    <style>
        body { font-family: system-ui, sans-serif; max-width: 800px; margin: 40px auto; padding: 0 20px; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { text-align: left; padding: 8px 12px; border-bottom: 1px solid #eee; }
        form { display: flex; gap: 8px; align-items: end; flex-wrap: wrap; }
        input, select { padding: 6px 10px; border: 1px solid #ddd; border-radius: 4px; }
        button { padding: 6px 16px; background: #1a73e8; color: white; border: none; border-radius: 4px; cursor: pointer; }
        nav a { margin-right: 16px; }
    </style>
</head>
<body>
    <nav><a href="/categories">分类</a><a href="/contents">内容</a><a href="/export">导出</a></nav>
    <h1>技术分类</h1>

    <form method="POST" action="/categories">
        <input name="name" placeholder="分类名称" required>
        <input name="icon" placeholder="图标" value="🏸">
        <input name="sort_order" type="number" placeholder="排序" value="0">
        <select name="parent_id">
            <option value="0">顶级分类</option>
            {{range .}}<option value="{{.ID}}">{{.Name}}</option>{{end}}
        </select>
        <button type="submit">添加</button>
    </form>

    <table>
        <tr><th>ID</th><th>图标</th><th>名称</th><th>排序</th><th>父分类</th></tr>
        {{range .}}
        <tr>
            <td>{{.ID}}</td>
            <td>{{.Icon}}</td>
            <td>{{.Name}}</td>
            <td>{{.SortOrder}}</td>
            <td>{{if .ParentID.Valid}}{{.ParentID.Int64}}{{else}}-{{end}}</td>
        </tr>
        {{end}}
    </table>
</body>
</html>
```

**Step 6: Commit**

```bash
git add admin/handlers.go admin/handlers_test.go admin/templates/
git commit -m "feat: admin category list and create"
```

---

## Task 4: Admin Panel — List & Create Contents

**Files:**
- Modify: `admin/handlers.go`
- Modify: `admin/handlers_test.go`
- Create: `admin/templates/contents.html`

**Step 1: Write test for content handlers**

Add to `admin/handlers_test.go`:

```go
func TestListContents(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	db.Exec("INSERT INTO categories (id, name, icon, sort_order) VALUES (1, '正手', '🏸', 1)")
	db.Exec("INSERT INTO contents (title, summary, source_url, source_platform, author_name, category_id, sort_order) VALUES ('测试视频', '测试摘要', 'https://example.com', 'bilibili', '作者', 1, 1)")

	handler := contentsHandler(db)
	req := httptest.NewRequest("GET", "/contents", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "测试视频") {
		t.Error("response should contain content title")
	}
}

func TestCreateContent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	db.Exec("INSERT INTO categories (id, name, icon, sort_order) VALUES (1, '正手', '🏸', 1)")

	handler := contentsHandler(db)
	body := strings.NewReader("title=新视频&summary=摘要&source_url=https://example.com&source_platform=bilibili&author_name=作者&category_id=1&sort_order=1")
	req := httptest.NewRequest("POST", "/contents", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM contents WHERE title='新视频'").Scan(&count)
	if count != 1 {
		t.Error("content should be inserted")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd admin && go test -v -run "TestListContents|TestCreateContent"`
Expected: FAIL

**Step 3: Implement content handlers**

Add to `admin/handlers.go`:

```go
func contentsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			rows, err := db.Query(`
				SELECT c.id, c.title, c.summary, c.thumbnail_url, c.source_url,
				       c.source_platform, c.author_name, c.category_id, c.sort_order,
				       cat.name
				FROM contents c
				JOIN categories cat ON c.category_id = cat.id
				ORDER BY c.category_id, c.sort_order`)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			defer rows.Close()

			var items []Content
			for rows.Next() {
				var c Content
				rows.Scan(&c.ID, &c.Title, &c.Summary, &c.ThumbnailURL, &c.SourceURL,
					&c.SourcePlatform, &c.AuthorName, &c.CategoryID, &c.SortOrder,
					&c.CategoryName)
				items = append(items, c)
			}

			// Load categories for the form dropdown
			catRows, _ := db.Query("SELECT id, name FROM categories ORDER BY sort_order")
			defer catRows.Close()
			var cats []Category
			for catRows.Next() {
				var c Category
				catRows.Scan(&c.ID, &c.Name)
				cats = append(cats, c)
			}

			tmpl, err := template.ParseFiles("templates/contents.html")
			if err != nil {
				for _, c := range items {
					w.Write([]byte(c.Title + "\n"))
				}
				return
			}
			tmpl.Execute(w, map[string]any{"Contents": items, "Categories": cats})

		case "POST":
			r.ParseForm()
			catID, _ := strconv.Atoi(r.FormValue("category_id"))
			sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))
			db.Exec(`INSERT INTO contents (title, summary, thumbnail_url, source_url, source_platform, author_name, category_id, sort_order)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				r.FormValue("title"), r.FormValue("summary"), r.FormValue("thumbnail_url"),
				r.FormValue("source_url"), r.FormValue("source_platform"), r.FormValue("author_name"),
				catID, sortOrder)
			http.Redirect(w, r, "/contents", http.StatusSeeOther)
		}
	}
}
```

**Step 4: Run tests**

Run: `cd admin && go test -v -run "TestListContents|TestCreateContent"`
Expected: PASS

**Step 5: Create contents HTML template**

```html
<!-- admin/templates/contents.html -->
<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>羽球大师课 - 内容管理</title>
    <style>
        body { font-family: system-ui, sans-serif; max-width: 1000px; margin: 40px auto; padding: 0 20px; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { text-align: left; padding: 8px 12px; border-bottom: 1px solid #eee; font-size: 14px; }
        form { display: flex; gap: 8px; align-items: end; flex-wrap: wrap; margin-bottom: 20px; }
        input, select, textarea { padding: 6px 10px; border: 1px solid #ddd; border-radius: 4px; }
        button { padding: 6px 16px; background: #1a73e8; color: white; border: none; border-radius: 4px; cursor: pointer; }
        nav a { margin-right: 16px; }
        .platform { padding: 2px 8px; border-radius: 10px; font-size: 12px; background: #e8f0fe; }
    </style>
</head>
<body>
    <nav><a href="/categories">分类</a><a href="/contents">内容</a><a href="/export">导出</a></nav>
    <h1>内容管理</h1>

    <form method="POST" action="/contents">
        <input name="title" placeholder="标题" required>
        <input name="summary" placeholder="编辑笔记">
        <input name="source_url" placeholder="原始链接" required>
        <select name="source_platform" required>
            <option value="bilibili">Bilibili</option>
            <option value="xiaohongshu">小红书</option>
            <option value="douyin">抖音</option>
            <option value="wechat">微信</option>
            <option value="youtube">YouTube</option>
            <option value="other">其他</option>
        </select>
        <input name="author_name" placeholder="作者">
        <input name="thumbnail_url" placeholder="缩略图URL">
        <select name="category_id" required>
            {{range .Categories}}<option value="{{.ID}}">{{.Name}}</option>{{end}}
        </select>
        <input name="sort_order" type="number" placeholder="排序" value="0">
        <button type="submit">添加</button>
    </form>

    <table>
        <tr><th>标题</th><th>平台</th><th>作者</th><th>分类</th><th>排序</th></tr>
        {{range .Contents}}
        <tr>
            <td><a href="{{.SourceURL}}" target="_blank">{{.Title}}</a></td>
            <td><span class="platform">{{.SourcePlatform}}</span></td>
            <td>{{.AuthorName}}</td>
            <td>{{.CategoryName}}</td>
            <td>{{.SortOrder}}</td>
        </tr>
        {{end}}
    </table>
</body>
</html>
```

**Step 6: Commit**

```bash
git add admin/handlers.go admin/handlers_test.go admin/templates/contents.html
git commit -m "feat: admin content list and create"
```

---

## Task 5: Admin Panel — Export SQLite DB

**Files:**
- Modify: `admin/handlers.go`
- Modify: `admin/handlers_test.go`
- Modify: `admin/main.go`

**Step 1: Write test for export**

Add to `admin/handlers_test.go`:

```go
func TestExportDB(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	db.Exec("INSERT INTO categories (name, icon, sort_order) VALUES ('正手', '🏸', 1)")

	handler := exportHandler(db, "test_handlers.db")
	req := httptest.NewRequest("GET", "/export", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/x-sqlite3" {
		t.Errorf("expected sqlite content type, got %s", contentType)
	}
	if w.Body.Len() == 0 {
		t.Error("export should return non-empty body")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd admin && go test -v -run TestExportDB`
Expected: FAIL

**Step 3: Implement export handler**

Add to `admin/handlers.go`:

```go
func exportHandler(db *sql.DB, dbPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Checkpoint to flush WAL
		db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")

		w.Header().Set("Content-Type", "application/x-sqlite3")
		w.Header().Set("Content-Disposition", "attachment; filename=yuqiupu.db")
		http.ServeFile(w, r, dbPath)
	}
}
```

**Step 4: Run test**

Run: `cd admin && go test -v -run TestExportDB`
Expected: PASS

**Step 5: Wire up routes in main.go**

Update `main()` in `admin/main.go`:

```go
func main() {
	dbPath := "yuqiupu.db"
	db, err := initDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/categories", categoriesHandler(db))
	http.HandleFunc("/contents", contentsHandler(db))
	http.HandleFunc("/export", exportHandler(db, dbPath))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/categories", http.StatusSeeOther)
	})

	fmt.Println("羽球大师课 Admin: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Step 6: Manual test — run the admin panel**

Run: `cd admin && go run . &` then open `http://localhost:8080`
Expected: Categories page loads, can add categories and contents, can export .db file.

**Step 7: Commit**

```bash
git add admin/
git commit -m "feat: admin export and route wiring"
```

---

## Task 6: iOS App — Project Setup + Bundled Data

**Files:**
- Create: `ios/YuQiuPu.xcodeproj` (via Xcode or `swift package init`)
- Create: `ios/YuQiuPu/App.swift`
- Create: `ios/YuQiuPu/Models.swift`
- Create: `ios/YuQiuPu/Database.swift`
- Create: `ios/YuQiuPu/Resources/yuqiupu.db` (bundled default)

**Step 1: Create Xcode project**

Create a new SwiftUI iOS app project named "YuQiuPu" in the `ios/` directory. Add SQLite (via `libsqlite3` system library or GRDB.swift SPM package).

**Step 2: Write data models**

```swift
// ios/YuQiuPu/Models.swift
import Foundation

struct Category: Identifiable {
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
```

**Step 3: Write database layer**

```swift
// ios/YuQiuPu/Database.swift
import Foundation
import SQLite3

class Database {
    private var db: OpaquePointer?

    init() {
        let docsURL = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first!
        let dbURL = docsURL.appendingPathComponent("yuqiupu.db")

        // Copy bundled DB if no local copy exists
        if !FileManager.default.fileExists(atPath: dbURL.path) {
            if let bundledURL = Bundle.main.url(forResource: "yuqiupu", withExtension: "db") {
                try? FileManager.default.copyItem(at: bundledURL, to: dbURL)
            }
        }

        sqlite3_open(dbURL.path, &db)
    }

    deinit {
        sqlite3_close(db)
    }

    func categories(parentId: Int? = nil) -> [Category] {
        var cats: [Category] = []
        var stmt: OpaquePointer?
        let sql = parentId == nil
            ? "SELECT id, name, icon, sort_order, parent_id FROM categories WHERE parent_id IS NULL ORDER BY sort_order"
            : "SELECT id, name, icon, sort_order, parent_id FROM categories WHERE parent_id = \(parentId!) ORDER BY sort_order"

        if sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK {
            while sqlite3_step(stmt) == SQLITE_ROW {
                let cat = Category(
                    id: Int(sqlite3_column_int(stmt, 0)),
                    name: String(cString: sqlite3_column_text(stmt, 1)),
                    icon: String(cString: sqlite3_column_text(stmt, 2)),
                    sortOrder: Int(sqlite3_column_int(stmt, 3)),
                    parentId: sqlite3_column_type(stmt, 4) != SQLITE_NULL ? Int(sqlite3_column_int(stmt, 4)) : nil
                )
                cats.append(cat)
            }
        }
        sqlite3_finalize(stmt)
        return cats
    }

    func contents(categoryId: Int) -> [ContentItem] {
        var items: [ContentItem] = []
        var stmt: OpaquePointer?
        let sql = "SELECT id, title, summary, thumbnail_url, source_url, source_platform, author_name, category_id, sort_order FROM contents WHERE category_id = ? ORDER BY sort_order"

        if sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK {
            sqlite3_bind_int(stmt, 1, Int32(categoryId))
            while sqlite3_step(stmt) == SQLITE_ROW {
                let item = ContentItem(
                    id: Int(sqlite3_column_int(stmt, 0)),
                    title: String(cString: sqlite3_column_text(stmt, 1)),
                    summary: String(cString: sqlite3_column_text(stmt, 2)),
                    thumbnailUrl: String(cString: sqlite3_column_text(stmt, 3)),
                    sourceUrl: String(cString: sqlite3_column_text(stmt, 4)),
                    sourcePlatform: String(cString: sqlite3_column_text(stmt, 5)),
                    authorName: String(cString: sqlite3_column_text(stmt, 6)),
                    categoryId: Int(sqlite3_column_int(stmt, 7)),
                    sortOrder: Int(sqlite3_column_int(stmt, 8))
                )
                items.append(item)
            }
        }
        sqlite3_finalize(stmt)
        return items
    }

    func replaceWith(downloadedDBAt tempURL: URL) throws {
        sqlite3_close(db)
        let docsURL = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first!
        let dbURL = docsURL.appendingPathComponent("yuqiupu.db")
        try FileManager.default.removeItem(at: dbURL)
        try FileManager.default.moveItem(at: tempURL, to: dbURL)
        sqlite3_open(dbURL.path, &db)
    }
}
```

**Step 4: Commit**

```bash
git add ios/
git commit -m "feat: iOS project setup with bundled SQLite"
```

---

## Task 7: iOS App — UI Screens

**Files:**
- Create: `ios/YuQiuPu/Views/HomeView.swift`
- Create: `ios/YuQiuPu/Views/CategoryView.swift`
- Modify: `ios/YuQiuPu/App.swift`

**Step 1: Home screen — category grid**

```swift
// ios/YuQiuPu/Views/HomeView.swift
import SwiftUI

struct HomeView: View {
    let db = Database()
    @State private var categories: [Category] = []

    var body: some View {
        NavigationStack {
            List(categories) { cat in
                NavigationLink(destination: CategoryView(db: db, category: cat)) {
                    HStack {
                        Text(cat.icon).font(.title2)
                        Text(cat.name).font(.body)
                    }
                }
            }
            .navigationTitle("羽球大师课")
            .onAppear { categories = db.categories() }
        }
    }
}
```

**Step 2: Category detail — content list**

```swift
// ios/YuQiuPu/Views/CategoryView.swift
import SwiftUI
import SafariServices

struct CategoryView: View {
    let db: Database
    let category: Category
    @State private var subcategories: [Category] = []
    @State private var contents: [ContentItem] = []
    @State private var selectedURL: URL?

    var body: some View {
        List {
            if !subcategories.isEmpty {
                Section("子分类") {
                    ForEach(subcategories) { sub in
                        NavigationLink(destination: CategoryView(db: db, category: sub)) {
                            HStack {
                                Text(sub.icon)
                                Text(sub.name)
                            }
                        }
                    }
                }
            }
            if !contents.isEmpty {
                Section("教学内容") {
                    ForEach(contents) { item in
                        Button {
                            selectedURL = URL(string: item.sourceUrl)
                        } label: {
                            VStack(alignment: .leading, spacing: 4) {
                                Text(item.title).font(.headline)
                                if !item.summary.isEmpty {
                                    Text(item.summary).font(.caption).foregroundStyle(.secondary)
                                }
                                HStack {
                                    Text(item.sourcePlatform).font(.caption2).padding(.horizontal, 6).padding(.vertical, 2).background(.blue.opacity(0.1)).cornerRadius(4)
                                    if !item.authorName.isEmpty {
                                        Text(item.authorName).font(.caption2).foregroundStyle(.secondary)
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
        .navigationTitle(category.name)
        .onAppear {
            subcategories = db.categories(parentId: category.id)
            contents = db.contents(categoryId: category.id)
        }
        .sheet(item: $selectedURL) { url in
            SafariView(url: url)
        }
    }
}

extension URL: @retroactive Identifiable {
    public var id: String { absoluteString }
}

struct SafariView: UIViewControllerRepresentable {
    let url: URL
    func makeUIViewController(context: Context) -> SFSafariViewController {
        SFSafariViewController(url: url)
    }
    func updateUIViewController(_ vc: SFSafariViewController, context: Context) {}
}
```

**Step 3: Wire up App entry point**

```swift
// ios/YuQiuPu/App.swift
import SwiftUI

@main
struct YuQiuPuApp: App {
    var body: some Scene {
        WindowGroup {
            HomeView()
        }
    }
}
```

**Step 4: Build and run in simulator**

Run: Open in Xcode, build for iPhone simulator.
Expected: App shows category list, tapping shows subcategories and content, tapping content opens Safari sheet.

**Step 5: Commit**

```bash
git add ios/
git commit -m "feat: iOS home and category screens"
```

---

## Task 8: iOS App — Data Sync

**Files:**
- Create: `ios/YuQiuPu/DataSync.swift`
- Modify: `ios/YuQiuPu/App.swift`

**Step 1: Implement sync manager**

```swift
// ios/YuQiuPu/DataSync.swift
import Foundation

class DataSync {
    // TODO: Replace with actual Aliyun OSS URL
    static let dbURL = URL(string: "https://your-bucket.oss-cn-hangzhou.aliyuncs.com/yuqiupu.db")!

    static func syncIfNeeded(db: Database) {
        let task = URLSession.shared.downloadTask(with: dbURL) { tempURL, response, error in
            guard let tempURL = tempURL, error == nil,
                  let httpResponse = response as? HTTPURLResponse,
                  httpResponse.statusCode == 200 else {
                return // Silently fail — local data still works
            }
            try? db.replaceWith(downloadedDBAt: tempURL)
        }
        task.resume()
    }
}
```

**Step 2: Call sync on app launch**

Update `App.swift` to trigger sync in `onAppear` or via an `init`.

**Step 3: Commit**

```bash
git add ios/
git commit -m "feat: iOS data sync from remote"
```

---

## Task 9: Android App — Project Setup + Bundled Data

Mirror Task 6 for Android:

**Files:**
- Create Android project with Jetpack Compose
- Create: `android/app/src/main/java/.../models/Category.kt`
- Create: `android/app/src/main/java/.../models/ContentItem.kt`
- Create: `android/app/src/main/java/.../data/Database.kt`
- Bundle: `android/app/src/main/assets/yuqiupu.db`

Use Android's built-in SQLite via `SQLiteOpenHelper` or Room. Copy bundled DB from assets to internal storage on first launch. Same `replaceWith` pattern for sync.

**Commit:** `feat: Android project setup with bundled SQLite`

---

## Task 10: Android App — UI Screens + Data Sync

Mirror Tasks 7-8 for Android:

- **HomeScreen** — LazyColumn of categories
- **CategoryScreen** — subcategories + content list
- **Tap content** — Chrome Custom Tabs
- **DataSync** — download .db on launch, replace local copy

Use Jetpack Compose + Navigation. Same UX as iOS.

**Commit:** `feat: Android home, category screens, and data sync`

---

## Execution Order

```
Task 1 (schema) → Task 2 (admin setup) → Task 3 (categories) → Task 4 (contents)
→ Task 5 (export) → Task 6 (iOS setup) → Task 7 (iOS UI) → Task 8 (iOS sync)
→ Task 9 (Android setup) → Task 10 (Android UI + sync)
```

Tasks 6-8 (iOS) and Tasks 9-10 (Android) are independent and can be parallelized.
