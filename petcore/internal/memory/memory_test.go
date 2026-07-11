package memory

import (
	"context"
	"testing"
)

func TestInMemoryManager_ShortTerm(t *testing.T) {
	m := NewInMemoryManager()
	m.AddShortTerm(Message{Role: "user", Content: "hello"})
	m.AddShortTerm(Message{Role: "assistant", Content: "hi"})

	msgs := m.GetShortTerm()
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

func TestInMemoryManager_ClearShortTerm(t *testing.T) {
	m := NewInMemoryManager()
	m.AddShortTerm(Message{Role: "user", Content: "hello"})
	m.ClearShortTerm()
	if len(m.GetShortTerm()) != 0 {
		t.Error("expected empty after clear")
	}
}

func TestInMemoryManager_RememberRecall(t *testing.T) {
	m := NewInMemoryManager()
	err := m.Remember("user_name", "Alice", 3)
	if err != nil {
		t.Fatal(err)
	}
	val, err := m.Recall("user_name")
	if err != nil {
		t.Fatal(err)
	}
	if val != "Alice" {
		t.Errorf("Recall = %q, want %q", val, "Alice")
	}
}

func TestInMemoryManager_RecallNotFound(t *testing.T) {
	m := NewInMemoryManager()
	_, err := m.Recall("nonexistent")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestInMemoryManager_GetAllCore(t *testing.T) {
	m := NewInMemoryManager()
	_ = m.Remember("a", "1", 1)
	_ = m.Remember("b", "2", 2)

	all, err := m.GetAllCore()
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

func TestInMemoryManager_StoreAndSearch(t *testing.T) {
	ctx := context.Background()
	m := NewInMemoryManager()

	_ = m.Store(ctx, Fact{Key: "user_mood", Value: "today is happy", Category: "event", Importance: 3})
	_ = m.Store(ctx, Fact{Key: "user_meal", Value: "ate pizza for lunch", Category: "event", Importance: 2})

	results, err := m.Search(ctx, "happy", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("Search('happy') got %d results, want 1", len(results))
	}

	results, err = m.Search(ctx, "pizza", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("Search('pizza') got %d results, want 1", len(results))
	}
}

func TestInMemoryManager_SearchNoResults(t *testing.T) {
	ctx := context.Background()
	m := NewInMemoryManager()
	results, err := m.Search(ctx, "nonexistent", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestInMemoryManager_SearchLimit(t *testing.T) {
	ctx := context.Background()
	m := NewInMemoryManager()
	for i := 0; i < 5; i++ {
		_ = m.Store(ctx, Fact{Key: "test", Value: "searchable content", Category: "event", Importance: 1})
	}
	results, err := m.Search(ctx, "searchable", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) > 2 {
		t.Errorf("Search with limit=2 returned %d results", len(results))
	}
}

func TestInMemoryManager_ConcurrentSafe(t *testing.T) {
	m := NewInMemoryManager()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			m.AddShortTerm(Message{Role: "user", Content: "hi"})
			_, _ = m.Recall("key")
		}
		done <- struct{}{}
	}()
	go func() {
		for i := 0; i < 100; i++ {
			_ = m.Remember("key", "val", 1)
			m.GetShortTerm()
		}
		done <- struct{}{}
	}()
	<-done
	<-done
}

func TestManagerInterface(t *testing.T) {
	// 编译时验证
	var _ Manager = (*InMemoryManager)(nil)
}
