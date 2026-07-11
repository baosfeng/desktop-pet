package event

import (
	"sync"
	"testing"
)

func TestEventTypeConstants(t *testing.T) {
	tests := []struct {
		got  Type
		want string
	}{
		{EventStateChanged, "state.changed"},
		{EventPetSpeak, "pet.speak"},
		{EventPetAction, "pet.action"},
		{EventPetEmotion, "pet.emotion"},
		{EventAgentThinking, "agent.thinking"},
		{EventAgentReply, "agent.reply"},
		{EventMemoryUpdated, "memory.updated"},
		{EventError, "error"},
	}
	for _, tt := range tests {
		t.Run(string(tt.got), func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Errorf("EventType = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestNewMeta_IncludesTimestamp(t *testing.T) {
	m := NewMeta("core", "session-1")
	if m.Timestamp == 0 {
		t.Error("expected non-zero timestamp")
	}
	if m.Source != "core" {
		t.Errorf("Source = %q, want %q", m.Source, "core")
	}
	if m.SessionID != "session-1" {
		t.Errorf("SessionID = %q, want %q", m.SessionID, "session-1")
	}
}

func TestNoopSink_Send_NoError(t *testing.T) {
	sink := NoopSink{}
	err := sink.Send(Event{Kind: EventPetSpeak, Data: "hello"})
	if err != nil {
		t.Errorf("NoopSink.Send returned error: %v", err)
	}
}

func TestNoopSink_Close_NoError(t *testing.T) {
	sink := NoopSink{}
	err := sink.Close()
	if err != nil {
		t.Errorf("NoopSink.Close returned error: %v", err)
	}
}

func TestNoopSink_ConcurrentSafe(t *testing.T) {
	// 验证 NoopSink 在并发下不会 panic
	sink := NoopSink{}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = sink.Send(Event{Kind: EventPetSpeak})
			_ = sink.Close()
		}()
	}
	wg.Wait()
}

func TestSinkInterface_Implemented(t *testing.T) {
	// 编译时验证：确保 NoopSink 实现了 Sink 接口
	var _ Sink = NoopSink{}
	_ = NoopSink{} // suppress unused
}
