// Package memory 提供三层记忆系统。
//
// SQLiteManager 是 Manager 的生产级实现，使用 modernc.org/sqlite（纯 Go，无 CGo）。
// 数据文件默认存储在 ~/.desktop-pet/memory.db。
package memory

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // SQLite driver registration (blank import required by database/sql)
)

// SQLiteManager 是 Manager 的 SQLite 持久化实现。
type SQLiteManager struct {
	db *sql.DB
}

// Ensure SQLiteManager implements Manager.
var _ Manager = (*SQLiteManager)(nil)

// NewSQLiteManager 创建或打开 SQLite 记忆数据库。
// 如果 dbPath 为空，默认使用 ~/.desktop-pet/memory.db。
func NewSQLiteManager(dbPath string) (*SQLiteManager, error) {
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("memory: cannot get home dir: %w", err)
		}
		dbPath = filepath.Join(home, ".desktop-pet", "memory.db")
	}

	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("memory: cannot create data dir %s: %w", dir, err)
	}

	dsn := fmt.Sprintf("file:%s?cache=shared&_journal_mode=WAL", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("memory: cannot open sqlite %s: %w", dbPath, err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	sm := &SQLiteManager{db: db}
	if err := sm.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("memory: migration failed: %w", err)
	}

	return sm, nil
}

// Close 关闭数据库连接。
func (sm *SQLiteManager) Close() error {
	return sm.db.Close()
}

// migrate 创建或更新数据库表结构。
func (sm *SQLiteManager) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS short_term (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at INTEGER NOT NULL DEFAULT (unixepoch())
		)`,
		`CREATE TABLE IF NOT EXISTS core_memory (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			category TEXT NOT NULL DEFAULT 'core',
			importance INTEGER NOT NULL DEFAULT 1,
			updated_at INTEGER NOT NULL DEFAULT (unixepoch())
		)`,
		`CREATE TABLE IF NOT EXISTS long_term (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			category TEXT NOT NULL,
			importance INTEGER NOT NULL,
			timestamp INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL DEFAULT (unixepoch())
		)`,
	}

	for _, m := range migrations {
		if _, err := sm.db.Exec(m); err != nil {
			return fmt.Errorf("memory: migration error: %w", err)
		}
	}

	// 尝试创建 FTS5 索引（可选的，失败不阻塞）
	ftsMigrations := []string{
		`CREATE VIRTUAL TABLE IF NOT EXISTS long_term_fts USING fts5(
			key, value, content='long_term', content_rowid='id'
		)`,
		`CREATE TRIGGER IF NOT EXISTS long_term_ai AFTER INSERT ON long_term BEGIN
			INSERT INTO long_term_fts(rowid, key, value) VALUES (new.id, new.key, new.value);
		END`,
		`CREATE TRIGGER IF NOT EXISTS long_term_ad AFTER DELETE ON long_term BEGIN
			INSERT INTO long_term_fts(long_term_fts, rowid, key, value) VALUES('delete', old.id, old.key, old.value);
		END`,
		`CREATE TRIGGER IF NOT EXISTS long_term_au AFTER UPDATE ON long_term BEGIN
			INSERT INTO long_term_fts(long_term_fts, rowid, key, value) VALUES('delete', old.id, old.key, old.value);
			INSERT INTO long_term_fts(rowid, key, value) VALUES (new.id, new.key, new.value);
		END`,
	}
	for _, m := range ftsMigrations {
		if _, err := sm.db.Exec(m); err != nil {
			// FTS5 可能不可用，静默忽略
			break
		}
	}

	return nil
}

// hasFTS reports whether FTS5 full-text search is available.
func (sm *SQLiteManager) hasFTS() bool {
	row := sm.db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='long_term_fts'")
	var count int
	return row.Scan(&count) == nil && count > 0
}

// ---- L1 短期记忆 ----

// AddShortTerm 添加一条消息到短期记忆。
func (sm *SQLiteManager) AddShortTerm(msg Message) {
	_, err := sm.db.Exec("INSERT INTO short_term (role, content) VALUES (?, ?)", msg.Role, msg.Content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "memory: AddShortTerm error: %v\n", err)
	}
}

// GetShortTerm 返回短期记忆中的所有消息。
func (sm *SQLiteManager) GetShortTerm() []Message {
	rows, err := sm.db.Query("SELECT role, content FROM short_term ORDER BY id ASC")
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()

	var msgs []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.Role, &msg.Content); err != nil {
			continue
		}
		msgs = append(msgs, msg)
	}
	return msgs
}

// ClearShortTerm 清空短期记忆。
func (sm *SQLiteManager) ClearShortTerm() {
	_, _ = sm.db.Exec("DELETE FROM short_term")
}

// ---- L2 核心记忆 ----

// Remember 保存一条核心记忆（upsert）。
func (sm *SQLiteManager) Remember(key, value string, importance int) error {
	_, err := sm.db.Exec(
		`INSERT INTO core_memory (key, value, category, importance, updated_at)
		 VALUES (?, ?, 'core', ?, unixepoch())
		 ON CONFLICT(key) DO UPDATE SET value=excluded.value, importance=excluded.importance, updated_at=unixepoch()`,
		key, value, importance,
	)
	if err != nil {
		return fmt.Errorf("memory: Remember error: %w", err)
	}
	return nil
}

// Recall 根据键名检索核心记忆。
func (sm *SQLiteManager) Recall(key string) (string, error) {
	row := sm.db.QueryRow("SELECT value FROM core_memory WHERE key = ?", key)
	var value string
	if err := row.Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("memory: key %q not found", key)
		}
		return "", fmt.Errorf("memory: Recall error: %w", err)
	}
	return value, nil
}

// GetAllCore 返回所有核心记忆的键值对。
func (sm *SQLiteManager) GetAllCore() (map[string]string, error) {
	rows, err := sm.db.Query("SELECT key, value FROM core_memory")
	if err != nil {
		return nil, fmt.Errorf("memory: GetAllCore error: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			continue
		}
		out[k] = v
	}
	return out, nil
}

// ---- L3 长期记忆 ----

// Store 保存一条长期记忆事实。
func (sm *SQLiteManager) Store(_ context.Context, fact Fact) error {
	ts := fact.Timestamp
	if ts == 0 {
		ts = time.Now().Unix()
	}
	_, err := sm.db.Exec(
		`INSERT INTO long_term (key, value, category, importance, timestamp)
		 VALUES (?, ?, ?, ?, ?)`,
		fact.Key, fact.Value, fact.Category, fact.Importance, ts,
	)
	if err != nil {
		return fmt.Errorf("memory: Store error: %w", err)
	}
	return nil
}

// Search 根据查询字符串搜索长期记忆。
// 如果 FTS5 可用则使用全文检索，否则回退到 LIKE 子串匹配。
func (sm *SQLiteManager) Search(_ context.Context, query string, limit int) ([]Fact, error) {
	if sm.hasFTS() {
		return sm.searchFTS(query, limit)
	}
	return sm.searchLike(query, limit)
}

func (sm *SQLiteManager) searchFTS(query string, limit int) ([]Fact, error) {
	q := `SELECT l.key, l.value, l.category, l.importance, l.timestamp
		  FROM long_term l
		  JOIN long_term_fts f ON l.id = f.rowid
		  WHERE long_term_fts MATCH ?
		  ORDER BY rank
		  LIMIT ?`
	return sm.queryFacts(q, query, limit)
}

func (sm *SQLiteManager) searchLike(query string, limit int) ([]Fact, error) {
	q := `SELECT key, value, category, importance, timestamp
		  FROM long_term
		  WHERE key LIKE ? OR value LIKE ?
		  ORDER BY id DESC
		  LIMIT ?`
	pattern := "%" + strings.ReplaceAll(query, "%", "\\%") + "%"
	return sm.queryFacts(q, pattern, pattern, limit)
}

func (sm *SQLiteManager) queryFacts(q string, args ...any) ([]Fact, error) {
	rows, err := sm.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("memory: Search error: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var facts []Fact
	for rows.Next() {
		var f Fact
		if err := rows.Scan(&f.Key, &f.Value, &f.Category, &f.Importance, &f.Timestamp); err != nil {
			continue
		}
		facts = append(facts, f)
	}
	return facts, nil
}
