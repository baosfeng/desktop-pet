package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteManager_NewAndClose(t *testing.T) {
	dir, err := os.MkdirTemp("", "memory-sqlite-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	dbPath := filepath.Join(dir, "test.db")
	sm, err := NewSQLiteManager(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteManager failed: %v", err)
	}
	defer sm.Close()

	// 验证表已创建
	rows, err := sm.db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	tables := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		tables[name] = true
	}

	for _, expected := range []string{"short_term", "core_memory", "long_term"} {
		if !tables[expected] {
			t.Errorf("missing table: %s", expected)
		}
	}
}

func TestSQLiteManager_ShortTerm(t *testing.T) {
	sm := newTestSQLite(t)
	defer sm.Close()

	sm.AddShortTerm(Message{Role: "user", Content: "hello"})
	sm.AddShortTerm(Message{Role: "assistant", Content: "hi there"})

	msgs := sm.GetShortTerm()
	if len(msgs) != 2 {
		t.Fatalf("got %d messages, want 2", len(msgs))
	}
	if msgs[0].Content != "hello" {
		t.Errorf("msg[0].Content = %q, want %q", msgs[0].Content, "hello")
	}
	if msgs[1].Role != "assistant" {
		t.Errorf("msg[1].Role = %q, want %q", msgs[1].Role, "assistant")
	}
}

func TestSQLiteManager_ClearShortTerm(t *testing.T) {
	sm := newTestSQLite(t)
	defer sm.Close()

	sm.AddShortTerm(Message{Role: "user", Content: "hello"})
	sm.ClearShortTerm()

	if len(sm.GetShortTerm()) != 0 {
		t.Error("expected empty after clear")
	}
}

func TestSQLiteManager_RememberRecall(t *testing.T) {
	sm := newTestSQLite(t)
	defer sm.Close()

	err := sm.Remember("user_name", "Alice", 3)
	if err != nil {
		t.Fatal(err)
	}

	val, err := sm.Recall("user_name")
	if err != nil {
		t.Fatal(err)
	}
	if val != "Alice" {
		t.Errorf("Recall = %q, want %q", val, "Alice")
	}
}

func TestSQLiteManager_RecallNotFound(t *testing.T) {
	sm := newTestSQLite(t)
	defer sm.Close()

	_, err := sm.Recall("nonexistent")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestSQLiteManager_RememberUpsert(t *testing.T) {
	sm := newTestSQLite(t)
	defer sm.Close()

	_ = sm.Remember("key1", "value1", 1)
	_ = sm.Remember("key1", "value2", 5)

	val, err := sm.Recall("key1")
	if err != nil {
		t.Fatal(err)
	}
	if val != "value2" {
		t.Errorf("after upsert, Recall = %q, want %q", val, "value2")
	}
}

func TestSQLiteManager_GetAllCore(t *testing.T) {
	sm := newTestSQLite(t)
	defer sm.Close()

	_ = sm.Remember("a", "1", 1)
	_ = sm.Remember("b", "2", 2)

	all, err := sm.GetAllCore()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("got %d core facts, want 2", len(all))
	}
	if all["a"] != "1" {
		t.Errorf("all[a] = %q, want %q", all["a"], "1")
	}
}

func TestSQLiteManager_StoreAndSearch(t *testing.T) {
	ctx := context.Background()
	sm := newTestSQLite(t)
	defer sm.Close()

	_ = sm.Store(ctx, Fact{Key: "user_mood", Value: "today is happy", Category: "event", Importance: 3})
	_ = sm.Store(ctx, Fact{Key: "user_meal", Value: "ate pizza for lunch", Category: "event", Importance: 2})

	// 空查询应返回所有
	results, err := sm.Search(ctx, "happy", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("Search('happy') got %d results, want 1", len(results))
	}

	results, err = sm.Search(ctx, "pizza", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("Search('pizza') got %d results, want 1", len(results))
	}
}

func TestSQLiteManager_SearchNoResults(t *testing.T) {
	ctx := context.Background()
	sm := newTestSQLite(t)
	defer sm.Close()

	results, err := sm.Search(ctx, "nonexistent", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSQLiteManager_SearchLimit(t *testing.T) {
	ctx := context.Background()
	sm := newTestSQLite(t)
	defer sm.Close()

	for i := 0; i < 5; i++ {
		_ = sm.Store(ctx, Fact{Key: "test", Value: "searchable content", Category: "event", Importance: 1})
	}
	results, err := sm.Search(ctx, "searchable", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) > 2 {
		t.Errorf("Search with limit=2 returned %d results", len(results))
	}
}

func TestSQLiteManager_DefaultDBPath(t *testing.T) {
	// 临时替换 HOME 以测试默认路径
	oldHome := os.Getenv("HOME")
	t.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", oldHome)

	sm, err := NewSQLiteManager("")
	if err != nil {
		t.Fatalf("NewSQLiteManager with empty path failed: %v", err)
	}
	defer sm.Close()

	// 验证数据文件已创建
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".desktop-pet", "memory.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("default DB file not created at %s", dbPath)
	}
}

// newTestSQLite 创建一个测试用的 SQLiteManager（临时文件）。
func newTestSQLite(t *testing.T) *SQLiteManager {
	t.Helper()
	dir, err := os.MkdirTemp("", "memory-sqlite-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	dbPath := filepath.Join(dir, "test.db")
	sm, err := NewSQLiteManager(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteManager failed: %v", err)
	}
	t.Cleanup(func() { sm.Close() })
	return sm
}
