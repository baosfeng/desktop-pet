package fsm

import (
	"testing"

	"github.com/desktop-pet/petcore/internal/event"
)

func TestStateConstants(t *testing.T) {
	tests := []struct {
		got  State
		want string
	}{
		{StateIdle, "idle"},
		{StateAttention, "attention"},
		{StateInteraction, "interaction"},
		{StateSpeaking, "speaking"},
	}
	for _, tt := range tests {
		t.Run(string(tt.got), func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Errorf("State = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		from  State
		evt   event.Type
		valid bool
	}{
		{StateIdle, event.EventStateChanged, true},
		{StateIdle, event.EventPetSpeak, true},
		{StateAttention, event.EventStateChanged, true},
		{StateAttention, event.EventPetSpeak, true},
		{StateAttention, event.EventError, true},
		{StateInteraction, event.EventAgentReply, true},
		{StateInteraction, event.EventError, true},
		{StateSpeaking, event.EventStateChanged, true},
		{StateIdle, event.EventAgentReply, false},
		{StateIdle, event.EventMemoryUpdated, false},
		{StateAttention, event.EventPetAction, false},
		{StateInteraction, event.EventPetEmotion, false},
		{StateSpeaking, event.EventPetSpeak, false},
	}
	for _, tt := range tests {
		name := string(tt.from) + "+" + string(tt.evt)
		t.Run(name, func(t *testing.T) {
			got := IsValidTransition(tt.from, tt.evt)
			if got != tt.valid {
				t.Errorf("IsValidTransition(%q, %q) = %v, want %v", tt.from, tt.evt, got, tt.valid)
			}
		})
	}
}

func TestErrTransitionNotAllowed_Error(t *testing.T) {
	err := &ErrTransitionNotAllowed{From: StateIdle, Evt: event.EventAgentReply}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestTransitionsFrom_NonEmpty(t *testing.T) {
	evts := TransitionsFrom(StateIdle)
	if len(evts) == 0 {
		t.Error("expected transitions from idle")
	}
}

func TestTransitionsFrom_UnknownState(t *testing.T) {
	evts := TransitionsFrom("unknown")
	if len(evts) != 0 {
		t.Errorf("expected empty transitions, got %d", len(evts))
	}
}

func TestMachineInterface(t *testing.T) {
	var _ Machine = (*MockMachine)(nil)
}

func TestMockMachine_DefaultBehavior(t *testing.T) {
	m := NewMockMachine(StateIdle)
	if m.Current() != StateIdle {
		t.Errorf("Current() = %q, want %q", m.Current(), StateIdle)
	}
	if err := m.Transition(event.EventStateChanged); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMockMachine_OnTransition(t *testing.T) {
	m := NewMockMachine(StateIdle)
	called := false
	m.OnTransition(func(from, to State) {
		called = true
	})
	_ = m.Transition(event.EventStateChanged)
	// MockMachine.Transition is a no-op, but OnTransition handler is stored
	if m.Current() != StateIdle {
		t.Errorf("expected state unchanged in mock")
	}
	m.onFn(StateIdle, StateAttention)
	if !called {
		t.Error("expected OnTransition callback to be callable")
	}
}
